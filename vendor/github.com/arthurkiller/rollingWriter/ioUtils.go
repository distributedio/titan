package bunnystub

import (
	"compress/gzip"
	"errors"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type IOWriter interface {
	io.Writer
	Close() error
}

var (
	// BufferSize defined the buffer size
	// about 2MB
	BufferSize = 0x6ffffff

	// ErrInternal defined the internal error
	ErrInternal = errors.New("error internal")
	// ErrInvalidArgument defined the invalid argument
	ErrInvalidArgument = errors.New("error argument invalid")
	// ErrMakingLogDirFailed means this path can not be created
	ErrMakingLogDirFailed = errors.New("error in make dir with given path")
)

type limittedWriter struct {
	wr        io.Writer
	condition func()
}

func warpWriter(w io.Writer, condition func()) io.Writer { return limittedWriter{w, condition} }

func (w limittedWriter) Write(s []byte) (int, error) {
	w.condition()
	return w.wr.Write(s)
}

// NewIOWriter generate a iofilter writer with given ioManager
func NewIOWriter(path string, kind RotatePolicy, ops ...Option) (IOWriter, error) {
	if path == "" {
		return nil, ErrInvalidArgument
	}
	ops = append(ops, WithPath(path))

	var m Manager
	switch kind {
	case TimeRotate:
		m = NewTimeRotateManager(ops...)
	case VolumeRotate:
		m = NewVolumeRotateManager(ops...)
	case WithoutRotate:
		fallthrough
	default:
		ss := strings.Split(path, "/")
		dirpath := strings.Join(ss[:len(ss)-1], "/")
		err := os.MkdirAll(dirpath, 0755)
		if err != nil {
			return nil, err
		}
		file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, m.FileMode())
		if err != nil {
			return nil, err
		}
		return file, nil
	}

	if m.NameParts()[1] == "" {
		return nil, ErrInvalidArgument
	}

	if m.LockFree() {
		var writer = &lockFreeWriter{
			buffer:    make(chan []byte, BufferSize),
			precision: time.Tick(time.Duration(Precision) * time.Second),
			manager:   m,
		}

		parts := m.NameParts()
		path, prefix, suffix := parts[0], parts[1], parts[2]
		err := os.MkdirAll(path, 0755)
		if err != nil {
			return nil, err
		}

		name := path + prefix + suffix
		file, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_APPEND, m.FileMode())
		if err != nil {
			return nil, err
		}
		writer.file = file

		writer.startWg.Add(1)

		go writer.conditionWrite()

		writer.startWg.Wait()

		return writer, nil
	}

	var writer = &fileWriter{precision: time.Tick(time.Duration(Precision) * time.Second), manager: m}

	parts := m.NameParts()
	path, prefix, suffix := parts[0], parts[1], parts[2]

	err := os.MkdirAll(path, 0755)
	if err != nil {
		return nil, err
	}

	name := path + prefix + suffix
	file, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_APPEND, m.FileMode())
	if err != nil {
		return nil, err
	}

	// XXX
	if m.Concurency() > 0 {
		writer.concurrency = int64(m.Concurency())
		if m.Spining() {
			writer.condition = func() {
				writer.spaningInvoke()
			}
		} else {
			var c bool
			writer.condition = func() {
				c = writer.invoke()
				if !c {
					// TODO do somting here
					time.Sleep(time.Microsecond)
				}
			}
		}
		//warpWriter(writer.file, condition)
	}

	writer.file = file
	writer.startWg.Add(1)

	go writer.conditionManager()

	// wait for conditionwrite start
	writer.startWg.Wait()

	return writer, nil
}

type fileWriter struct {
	l                  sync.RWMutex
	file               *os.File
	manager            Manager
	startWg            sync.WaitGroup
	precision          <-chan time.Time
	close              chan byte
	condition          func()
	concurrency        int64
	concurrencyCounter int64
}

// invoke ask for the write permission, if not permitated, invoke will return
func (w *fileWriter) invoke() bool {
	if atomic.LoadInt64(&w.concurrencyCounter) < w.concurrency {
		atomic.AddInt64(&w.concurrencyCounter, 1)
		return true
	} else {
		return false
	}
}

// spining invoke ask for the permission with a spinning-like behavior
func (w *fileWriter) spaningInvoke() {
	for {
		if atomic.LoadInt64(&w.concurrencyCounter) < w.concurrency {
			atomic.AddInt64(&w.concurrencyCounter, 1)
			return
		}
	}
}

// handout give back the resource
func (w *fileWriter) handout() {
	atomic.AddInt64(&w.concurrencyCounter, -1)
}

func (w *fileWriter) Write(s []byte) (int, error) {
	w.l.RLock()
	defer w.l.RUnlock()
	if w.concurrency > 0 {
		w.condition()
	}
	n, err := w.file.Write(s)
	w.handout()
	w.manager.WriteN(n)
	return n, err
}

func (w *fileWriter) Close() error {
	close(w.close)

	return w.file.Close()
}

func (w *fileWriter) conditionManager() {
	chk := w.manager.Enable()
	w.startWg.Done()

	for {
		select {
		case lastname := <-chk:
			oldFile := w.file
			oldFileName := oldFile.Name()
			parts := w.manager.NameParts()
			path, suf := parts[0], parts[2]

			err := os.Rename(oldFileName, lastname)
			if err != nil {
				log.Println("error in rename file", err)
			}

			//mkdir for the path
			err = os.MkdirAll(path, 0755)
			if err != nil {
				oldFileName = "./" + suf
				log.Println("error in mkdir: file cannot be created in this path", path, ", log has moved into './' dir", oldFileName)
			}

			// open & swap the file
			file, err := os.OpenFile(oldFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, w.manager.FileMode())
			if err != nil {
				log.Println("error in reopen file", err)
				continue
			}
			w.l.Lock()
			w.file = file
			w.l.Unlock()

			// Do additional jobs like compresing the log file
			go func() {
				defer oldFile.Close()
				if w.manager.Compress() {
					// Do compress the log file
					// name the compressed file
					// delete the old file
					cmpname := lastname + ".gz"
					cmpfile, err := os.OpenFile(cmpname, os.O_RDWR|os.O_CREATE|os.O_APPEND, w.manager.FileMode())
					defer cmpfile.Close()
					if err != nil {
						log.Println("error in reopen additional goroution", err)
						return
					}
					gw := gzip.NewWriter(cmpfile)
					defer gw.Close()

					oldFile.Seek(0, 0)
					_, err = io.Copy(gw, oldFile)
					if err != nil {
						cmpfile.Close()
						os.Remove(cmpname)
						return
					}
					os.Remove(lastname) //remove *.log
				}
			}()
		case <-w.close:
			break
		}
	}
}

type lockFreeWriter struct {
	file   *os.File
	buffer chan []byte
	close  chan byte

	startWg   sync.WaitGroup
	precision <-chan time.Time
	version   int64
	manager   Manager
	wg        sync.WaitGroup
}

func (w *lockFreeWriter) Write(s []byte) (int, error) {
	// TODO add a pool here
	// Reduce function call
	w.wg.Add(1)
	defer w.wg.Done()
	n := len(s)
	w.manager.WriteN(n)
	w.buffer <- s
	return n, nil
}

func (w *lockFreeWriter) Close() error {
	close(w.close)

	return w.file.Close()
}

func (w *lockFreeWriter) conditionWrite() {
	chk := w.manager.Enable()
	w.startWg.Done()

	for {
		select {
		case path := <-chk:
			w.reopen(path)
		case v := <-w.buffer:
			_, err := w.file.Write(v)
			// if ignore error then do nothing
			if err != nil && !w.manager.IgnoreOK() {
				// FIXME is this a good way?
				i, errn := w.file.Stat()
				if errn == nil {
					log.Println("err in condition write", err, i)
				} else {
					log.Println("err in condition write", err)
				}
				parts := w.manager.NameParts()
				path, prefix, suffix := parts[0], parts[1], parts[2]
				retryOpenFileName := path + prefix + suffix
				retryOpenfile, err := os.OpenFile(retryOpenFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, w.manager.FileMode())
				if err != nil {
					retryOpenFileName := "./" + prefix + suffix
					retryOpenfile, err := os.OpenFile(retryOpenFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, w.manager.FileMode())
					if err != nil {
						w.file = os.Stderr
						continue
					}
					w.file = retryOpenfile
					continue
				}
				w.file = retryOpenfile
			}
		case <-w.close:
			break
		}
	}
}

func (w *lockFreeWriter) reopen(lastname string) {
	oldFile := w.file
	oldFileName := oldFile.Name()
	err := os.Rename(oldFileName, lastname)
	if err != nil {
		log.Println("error in rename file", err)
	}

	//mkdir for the path
	parts := w.manager.NameParts()
	path, suf := parts[0], parts[2]
	err = os.MkdirAll(path, 0755)
	if err != nil {
		oldFileName = "./" + suf
		log.Println("error in mkdir: file cannot be created in this path", path, ", log has moved into './' dir", oldFileName, err)
	}

	// open & swap the file
	file, err := os.OpenFile(oldFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, w.manager.FileMode())
	if err != nil {
		log.Println("error in reopen file", err)
		return
	}
	w.file = file

	// Do additional jobs like compresing the log file
	go func() {
		w.wg.Wait()
		if w.manager.Compress() {
			// 1.Do compress the log file
			// 2.name the compressed file
			// 3.delete the old file
			cmpname := lastname + ".gz"
			cmpfile, err := os.OpenFile(cmpname, os.O_RDWR|os.O_CREATE|os.O_APPEND, w.manager.FileMode())
			defer cmpfile.Close()
			if err != nil {
				log.Println("error in reopen doing compress", err)
				return
			}
			gw := gzip.NewWriter(cmpfile)
			defer gw.Close()

			oldFile.Seek(0, 0)
			_, err = io.Copy(gw, oldFile)
			if err != nil {
				log.Println("error in compress log file", err)
				cmpfile.Close()
				os.Remove(cmpname)
				return
			}
			defer os.Remove(lastname)
		}
		defer oldFile.Close()
	}()
}
