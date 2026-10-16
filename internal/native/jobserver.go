package native

import (
	"errors"
	"fmt"
	"os"
	"sync"
)

// A Jobserver bounds native's C compiles (#cyew6p3): a FIFO holding one byte a token, as make's fifo jobserver
// protocol fills it. Each clang native runs under one takes a token first and gives it back when clang exits, whatever
// it exited with, so Loom bounds how many compile at once across every build on a box (Workshop's native products ran
// 68 to 79 clang at once, 99% busy, Oct 10). A nil Jobserver bounds nothing, as a process given none always did.
//
// It's handed in, and native reads no environment for it (#yzwm4rj): the archive's build takes it as a parameter
// (RuntimeRequest.Build), so the build reads no state set outside its own initializers, which lever 3's census holds,
// and every other compile takes the process's (UseJobserver), which internal/ccjobserver sets from ADAMIC_CC_JOBSERVER.
// It moves no key: no key, Parts or Portable reads it, since how many compile at once changes no byte
// (TestTheJobserverNeverReachesTheBytes). Under a trace it moves none either, as long as the FIFO sits under /run or
// /dev (Loom puts it in /run/user/<uid> or /dev/shm): those are kernel places buildcache keys nothing in. A launcher
// wrapped around clang would instead be a program the trace keys or refuses (Loom's ruling, Oct 13).
type Jobserver struct {
	path string
	fifo *os.File
}

// OpenJobserver opens the jobserver whose FIFO is at path, for reading and writing, as make opens its own: a FIFO
// opened so never blocks on its other end. One that can't be opened is an error naming it.
func OpenJobserver(path string) (*Jobserver, error) {
	fifo, err := os.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		return nil, fmt.Errorf("native: the compile jobserver %s: %w", path, err)
	}
	return &Jobserver{path: path, fifo: fifo}, nil
}

// Hold runs compile holding one of the jobserver's tokens, given back however compile ends, a panic included; a nil
// jobserver runs it unbounded. A token that can't be taken fails the compile, naming the jobserver, rather than letting
// it run unbounded.
func (jobserver *Jobserver) Hold(compile func() error) error {
	if jobserver == nil {
		return compile()
	}
	token := make([]byte, 1)
	if _, err := jobserver.fifo.Read(token); err != nil {
		return fmt.Errorf("native: taking a compile token from the jobserver %s: %w", jobserver.path, err)
	}
	defer jobserver.fifo.Write(token)
	return compile()
}

// processJobserver is the jobserver every compile but the archive's build runs under (UseJobserver), set once a
// process.
var processJobserver struct {
	mutex     sync.Mutex
	set       bool
	jobserver *Jobserver
	err       error
}

// ErrJobserverSet is UseJobserver given another jobserver than the one the process already has.
var ErrJobserverSet = errors.New("native: the process's compile jobserver is set already, to another")

// UseJobserver sets the jobserver Build, BuildFlags, BuildTSGo, BuildSplitTSGo, the runtime cache and the header scan
// behind RuntimeRequestFor compile under: jobserver, or err, why the one named couldn't be opened, which each of those
// compiles then fails with rather than running unbounded. It's set once a process: the same again changes nothing, and
// anything else is ErrJobserverSet, so two packages can't name two FIFOs and nothing swaps the bound mid-run (Compiler's
// condition on #yzwm4rj). The archive's build, RuntimeRequest.Build, never reads it: it compiles under the jobserver
// it's handed, unbounded when that's nil.
func UseJobserver(jobserver *Jobserver, err error) error {
	processJobserver.mutex.Lock()
	defer processJobserver.mutex.Unlock()
	if processJobserver.set {
		if processJobserver.jobserver == jobserver && processJobserver.err == err {
			return nil
		}
		return ErrJobserverSet
	}
	processJobserver.set, processJobserver.jobserver, processJobserver.err = true, jobserver, err
	return nil
}

// defaultJobserver is the process's jobserver (UseJobserver): nil when none was set, or why the one named couldn't be
// opened. It's read only where a compile outside the archive's build starts, and handed down from there, never in a
// function RuntimeRequest.Build reaches (Node's rail on #yzwm4rj).
func defaultJobserver() (*Jobserver, error) {
	processJobserver.mutex.Lock()
	defer processJobserver.mutex.Unlock()
	return processJobserver.jobserver, processJobserver.err
}
