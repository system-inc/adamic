// Package ccjobserver hands native's C compiles the jobserver Loom names in ADAMIC_CC_JOBSERVER (#cyew6p3, #yzwm4rj).
// native reads no environment for it, so the archive's build can take it as a value and key on nothing else, and a
// process whose native builds Loom bounds imports this package once, for its effect: it opens the FIFO the variable
// names and sets it as the process's (native_runtime.UseJobserver). compilercensus holds every package that builds through
// native to importing it, so no build goes unbounded for want of the import.
package ccjobserver

import (
	"fmt"
	"os"

	"github.com/system-inc/adamic/internal/native_runtime"
)

// Variable names the jobserver's FIFO, which Loom's tree builder makes under /run or /dev and fills with a token a
// compile it allows at once.
const Variable = "ADAMIC_CC_JOBSERVER"

// process is the jobserver Variable names, nil when it names none, and failed why the one it names couldn't be opened.
var process, failed = open(os.Getenv(Variable))

// open opens the jobserver at path, none for an empty one, the error naming the variable.
func open(path string) (*native_runtime.Jobserver, error) {
	if path == "" {
		return nil, nil
	}
	jobserver, err := native_runtime.OpenJobserver(path)
	if err != nil {
		return nil, fmt.Errorf("%s=%s: %w", Variable, path, err)
	}
	return jobserver, nil
}

// A process the variable names no jobserver for sets none, leaving native_runtime's unset, so a test binary that imports
// this package through another (native's own, by leakcheck) can still set its children's, as
// internal/native_runtime's inChild sets its own.
func init() {
	if process == nil && failed == nil {
		return
	}
	if err := native_runtime.UseJobserver(process, failed); err != nil {
		panic(err)
	}
}

// Process is the jobserver Variable names, nil when it names none, or why the one it names couldn't be opened: what
// runtimeproducts hands an archive's build (native_runtime.RuntimeRequest.Build).
func Process() (*native_runtime.Jobserver, error) {
	return process, failed
}
