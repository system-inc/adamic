package native

import (
	"fmt"
	"os"
	"sync"
)

// JobserverVariable names a jobserver that bounds native's C compiles (#cyew6p3): a FIFO holding one byte a token, as
// make's fifo jobserver protocol fills it. Each clang native runs takes a token first and gives it back when clang
// exits, whatever it exited with, so Loom bounds how many compile at once across every build on a box (Workshop's
// native products ran 68 to 79 clang at once, 99% busy, Oct 10). The process-local clang pool also applies when unset.
//
// It moves no key: nothing reads it into a unit's or a runtime's key, and the compiler is still keyed by its own path
// and report. Under a trace it moves none either, as long as the FIFO sits under /run or /dev (Loom puts it in
// /run/user/<uid> or /dev/shm): those are kernel places buildcache keys nothing in, and a build may write outside its
// tree. A launcher wrapped around clang would instead be a program the trace keys or refuses (Loom's ruling, Oct 13).
const JobserverVariable = "ADAMIC_CC_JOBSERVER"

// compilerJobserver is the jobserver's FIFO, opened once for the process at its first compile.
type compilerJobserver struct {
	once sync.Once
	fifo *os.File
	err  error
}

var jobserver = &compilerJobserver{}

// withCompilerToken runs compile holding one of the jobserver's tokens, when one is named. A FIFO that can't be opened
// or read fails the compile, naming it, rather than letting it run unbounded.
func withCompilerToken(compile func() error) error {
	release := acquireClang()
	defer release()
	path := os.Getenv(JobserverVariable)
	if path == "" {
		return compile()
	}
	jobserver.once.Do(func() {
		// Opened for reading and writing, as make opens its own: a FIFO opened so never blocks on its other end.
		jobserver.fifo, jobserver.err = os.OpenFile(path, os.O_RDWR, 0)
	})
	if jobserver.err != nil {
		return fmt.Errorf("native: %s=%s: %w", JobserverVariable, path, jobserver.err)
	}
	token := make([]byte, 1)
	if _, err := jobserver.fifo.Read(token); err != nil {
		return fmt.Errorf("native: taking a compile token from %s=%s: %w", JobserverVariable, path, err)
	}
	defer jobserver.fifo.Write(token)
	return compile()
}
