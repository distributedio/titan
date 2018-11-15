package rollingwriter

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"sync/atomic"
	"unsafe"
)

// Writer provide a synchronous file writer
// if Lock is set true, write will be guaranteed by lock
type Writer struct {
	file            *os.File
	absolutePath    string
	fire            chan string
	cf              *Config
	rollingfilelist []string
}

// LockedWriter provide a synchronous writer with lock
// write operate will be guaranteed by lock
type LockedWriter struct {
	Writer
	lock sync.Mutex
}

// AsynchronousWriter provide a asynchronous writer with the writer to confirm the write
type AsynchronousWriter struct {
	Writer
	ctx     chan int
	queue   chan []byte
	errChan chan error
	closed  int32
	wg      sync.WaitGroup
	once    sync.Once
}

// BufferWriter provide a parallel safe bufferd writer
// TODO TBD
type BufferWriter struct {
	Writer
	wr io.Writer
}

// buffer pool for asynchronous writer
var _asyncBufferPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, BufferSize)
	},
}

// NewWriterFromConfig generate the rollingWriter with given config
func NewWriterFromConfig(c *Config) (RollingWriter, error) {
	// makeup log path and create
	if c.LogPath == "" || c.FileName == "" {
		return nil, ErrInvalidArgument
	}

	// make dir for path if not exist
	if err := os.MkdirAll(c.LogPath, 0700); err != nil {
		return nil, err
	}

	filepath := LogFilePath(c)
	// open the file and get the FD
	file, err := os.OpenFile(filepath, DefaultFileFlag, DefaultFileMode) // Open file witl rw-r--r-- by default
	if err != nil {
		return nil, err
	}

	filel := make([]string, 0, 7)
	if c.MaxRemain > 0 {
		filel = make([]string, c.MaxRemain*7)
	}

	// Start the Manager
	mng, err := NewManager(c)
	if err != nil {
		return nil, err
	}
	var writer RollingWriter
	if c.Asynchronous { // async writer
		wr := &AsynchronousWriter{
			ctx:     make(chan int),
			queue:   make(chan []byte, QueueSize),
			errChan: make(chan error),
			wg:      sync.WaitGroup{},
			closed:  0,
			Writer: Writer{
				file:            file,
				absolutePath:    filepath,
				fire:            mng.Fire(),
				cf:              c,
				rollingfilelist: filel,
			},
		}
		// start the asynchronous writer
		wr.wg.Add(1)
		go wr.writer()
		wr.wg.Wait()
		writer = wr
	} else {
		if c.Lock { // locked writer
			writer = &LockedWriter{
				Writer: Writer{
					file:            file,
					absolutePath:    filepath,
					fire:            mng.Fire(),
					cf:              c,
					rollingfilelist: filel,
				},
			}
		} else { // normal writer
			writer = &Writer{
				file:            file,
				absolutePath:    filepath,
				fire:            mng.Fire(),
				cf:              c,
				rollingfilelist: filel,
			}
		}
	}
	return writer, nil
}

// NewWriter generate the rollingWriter with given option
func NewWriter(ops ...Option) (RollingWriter, error) {
	cfg := NewDefaultConfig()
	for _, opt := range ops {
		opt(&cfg)
	}
	return NewWriterFromConfig(&cfg)
}

// NewWriterFromConfigFile generate the rollingWriter with given config file
func NewWriterFromConfigFile(path string) (RollingWriter, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	cfg := NewDefaultConfig()
	buf, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(buf, cfg)
	if err != nil {
		return nil, err
	}

	return NewWriterFromConfig(&cfg)
}

// AutoRemove will delete the oldest file
func (w *Writer) AutoRemove() {
	for len(w.rollingfilelist) > w.cf.MaxRemain {
		// remove the oldest file
		file := w.rollingfilelist[0]
		if err := os.Remove(file); err != nil {
			log.Println("error in auto remove log file", err)
		}
		w.rollingfilelist = w.rollingfilelist[1:]
	}
}

// CompressFile compress log file write into .gz and remove source file
func (w *Writer) CompressFile(oldfile *os.File, cmpname string) error {
	cmpfile, err := os.OpenFile(cmpname, DefaultFileFlag, DefaultFileMode)
	defer cmpfile.Close()
	if err != nil {
		return err
	}
	gw := gzip.NewWriter(cmpfile)
	defer gw.Close()

	if _, err := oldfile.Seek(0, 0); err != nil {
		return err
	}
	if _, err := io.Copy(gw, oldfile); err != nil {
		if errR := os.Remove(cmpname); err != nil {
			return errR
		}
		return err
	}
	return os.Remove(cmpname + ".tmp") //remove *.log.tmp file
}

// AsynchronousWriterErrorChan return the error channel for asyn writer
func AsynchronousWriterErrorChan(wr RollingWriter) (chan error, error) {
	if w, ok := wr.(*AsynchronousWriter); ok {
		return w.errChan, nil
	}
	return nil, ErrInvalidArgument
}

// Reopen do the rotate, open new file and swap FD then trate the old FD
func (w *Writer) Reopen(file string) error {
	// do the rename
	if err := os.Rename(w.absolutePath, file); err != nil {
		return err
	}

	// open & swap the file
	newfile, err := os.OpenFile(w.absolutePath, DefaultFileFlag, DefaultFileMode)
	if err != nil {
		return err
	}

	// swap the unsafe pointer
	oldfile := atomic.SwapPointer((*unsafe.Pointer)(unsafe.Pointer(&w.file)), unsafe.Pointer(newfile))

	if w.cf.MaxRemain > 0 {
		// add to the delete file list
		w.rollingfilelist = append(w.rollingfilelist, file)
	}

	// Do aditional jobs
	go func() {
		defer (*os.File)(oldfile).Close()
		if w.cf.Compress {
			if err := os.Rename(file, file+".tmp"); err != nil {
				log.Println("error in compress rename tempfile", err)
				return
			}
			if err := w.CompressFile((*os.File)(oldfile), file); err != nil {
				log.Println("error in compress log file", err)
				return
			}
		}

		if w.cf.MaxRemain > 0 {
			w.AutoRemove()
		}
	}()

	return nil
}

func (w *Writer) Write(b []byte) (int, error) {
	select {
	case filename := <-w.fire:
		// do the reopen
		if err := w.Reopen(filename); err != nil {
			return 0, err
		}
		return w.file.Write(b)
	default:
		return w.file.Write(b)
	}
}

func (w *LockedWriter) Write(b []byte) (int, error) {
	select {
	case filename := <-w.fire:
		// do the reopen
		if err := w.Reopen(filename); err != nil {
			return 0, err
		}

		w.lock.Lock()
		defer w.lock.Unlock()
		return w.file.Write(b)
	default:
		w.lock.Lock()
		defer w.lock.Unlock()
		return w.file.Write(b)
	}
}

// Only when the error channel is empty, otherwise nothing will write and the last error will be return
// return the error channel
func (w *AsynchronousWriter) Write(b []byte) (int, error) {
	if atomic.LoadInt32(&w.closed) == 0 {
		select {
		case err := <-w.errChan:
			// return the error
			// NOTICE this error caused by last write maybe ignored
			return 0, err

		// here we will not handle the ctx close
		//case <-w.ctx:
		//	return 0, ErrClosed

		case filename := <-w.fire:
			// do the reopen
			if err := w.Reopen(filename); err != nil {
				return 0, err
			}

			l := len(b)
			for len(b) > 0 {
				buf := _asyncBufferPool.Get().([]byte)
				n := copy(buf, b)
				w.queue <- buf[:n]
				b = b[n:]
			}
			return l, nil
		default:
			// here we need to block while the channel is full
			w.queue <- append(_asyncBufferPool.Get().([]byte)[0:], b...)[:len(b)]
			return len(b), nil
		}
	}
	return 0, ErrClosed
}

// writer do the asynchronous write independently
// Take care of reopen, I am not sure if there need no lock
func (w *AsynchronousWriter) writer() {
	var err error
	w.wg.Done()
	for {
		select {
		case b := <-w.queue:
			if _, err = w.file.Write(b); err != nil {
				w.errChan <- err
			}
			_asyncBufferPool.Put(b)
		case <-w.ctx:
			// writer write the buffered bytes then exit on context closed
			return
		}
	}
}

// Close the file and return
func (w *Writer) Close() error {
	return w.file.Close()
}

// Close lock and close the file
func (w *LockedWriter) Close() error {
	w.lock.Lock()
	defer w.lock.Unlock()
	return w.file.Close()
}

// Close lock and close the file
func (w *AsynchronousWriter) Close() error {
	w.once.Do(func() {
		atomic.StoreInt32(&w.closed, 1)
		close(w.ctx)
	})

	w.onClose()
	return w.file.Close()
}

// onClose process remaining bufferd data for asynchronous writer
func (w *AsynchronousWriter) onClose() {
	var err error
	for {
		select {
		case b := <-w.queue:
			// flush all remaining field
			if _, err = w.file.Write(b); err != nil {
				select {
				case w.errChan <- err:
				default:
					_asyncBufferPool.Put(b)
					return
				}
			}
			_asyncBufferPool.Put(b)
		default: // after the queue was empty, return
			return
		}
	}
}
