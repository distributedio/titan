# RollingWriter [![Build Status](https://travis-ci.org/arthurkiller/rollingWriter.svg?branch=master)](https://travis-ci.org/arthurkiller/rollingWriter) [![Go Report Card](https://goreportcard.com/badge/github.com/arthurkiller/rollingwriter)](https://goreportcard.com/report/github.com/arthurkiller/rollingwriter) [![GoDoc](https://godoc.org/github.com/arthurkiller/rollingWriter?status.svg)](https://godoc.org/github.com/arthurkiller/rollingWriter) [![codecov](https://codecov.io/gh/arthurkiller/rollingwriter/branch/master/graph/badge.svg)](https://codecov.io/gh/arthurkiller/rollingwriter)
RollingWriter is an auto rotate io.Writer implementation. It always works with logger.

__New Version v2.0 is comeing out! Much more Powerfull and Efficient. Try it by follow the demo__

it contains 2 separate patrs:
* Manager: decide when to rotate the file with policy
    RlingPolicy give out the rolling policy
    1. WithoutRolling: no rolling will happen
    2. TimeRolling: rolling by time
    3. VolumeRolling: rolling by file size

* IOWriter: impement the io.Writer and do the io write
    * Writer: not parallel safe writer
    * LockedWriter: parallel safe garented by lock
    * AsyncWtiter: parallel safe async writer

## Features
* Auto rotate
* Parallel safe
* Implement go io.Writer
* Time rotate with corn style task schedual
* Volume rotate
* Max remain rolling files with auto clean

## Quick Start
```golang
	writer, err := rollingwriter.NewWriterFromConfig(&config)
	if err != nil {
		panic(err)
	}

	writer.Write([]byte("hello, world"))
```
for more, goto demo to see more details

## Contribute && TODO
* the Buffered writer need to be dcescuessed
* fix the FIXME

Now I am about to release the v1.0.0-prerelease with redesigned interface

Any new feature inneed pls [put up an issue](https://github.com/arthurkiller/rollingWriter/issues/new)
