// Package buildlog writes the compiler's phase lines to $ADAMIC_BUILD_LOG: how long a program took to load, to lower
// and to emit as C, so loom can size lowering across the tree (#9x03wa7). internal/load, internal/lower and
// internal/native each time their own phase, and neither of the first two can import internal/native, whose buildLog
// writes the log's clang lines; this package imports nothing of the tree, so all three can reach it.
package buildlog

import (
	"fmt"
	"os"
	"time"
)

// Phase appends one line about a phase of the compiler to $ADAMIC_BUILD_LOG when it is set, in the format of
// internal/native's buildLog: 'phase <name> - off <seconds>', name load, lower or emit, - since nothing keys a phase,
// off since none reads a cache. started is when the phase began, so a caller's
//
//	defer buildlog.Phase("load", time.Now())
//
// logs its whole call, an error's return included, once a call. Every reader of the log matches its own kind of line,
// so these pass them by.
//
// The line never reaches what is built: nothing it writes is an input to the program, the C or clang, and a log that
// can't be opened writes nothing and fails nothing, as buildLog's doesn't.
func Phase(name string, started time.Time) {
	path := os.Getenv("ADAMIC_BUILD_LOG")
	if path == "" {
		return
	}
	line := fmt.Sprintf("phase %s - off %.3f\n", name, time.Since(started).Seconds())
	if file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644); err == nil {
		file.WriteString(line)
		file.Close()
	}
}
