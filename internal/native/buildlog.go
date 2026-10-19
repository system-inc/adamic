package native

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// buildLog appends one line about a clang step to $ADAMIC_BUILD_LOG when it is set, beside buildcache's census lines
// (internal/buildcache's record): '<kind> <name> <key12> hit|miss|off <seconds>'. kind is cCompile for a translation
// unit, cArchive for the runtime's archive, cLink for a link, and cBuild for one clang call that compiles a program's one
// unit and links it (a one-unit build off macOS), so neither half is taken for the other. name is the unit, archive or
// output it made; key12 is the first 12 of the cache key the step was looked up by, or - when nothing keys it. hit is an
// object or archive the cache held, miss one made into the cache, off a step that read no cache. seconds is how long the
// tool ran for it, a wait for a compile token left out: a split unit's hit is its preprocessing, an archive's hit runs
// nothing and is 0, and an archive's miss is ar alone, its units' compiles each on their own cCompile line. Loom reads these lines to put C compile and link on a run's critical path (#s0jdsaa);
// every reader of the census matches its own kind of line, so these pass them by.
//
// The line never reaches what is built: nothing it writes is an input to clang, and a log that can't be opened writes
// nothing and fails nothing, as record's doesn't.
func buildLog(kind, name, key, outcome string, seconds time.Duration) {
	path := os.Getenv("ADAMIC_BUILD_LOG")
	if path == "" {
		return
	}
	if key == "" {
		key = "-"
	}
	// Every whitespace character in name becomes _, so an output path holding a tab or a newline can neither split the
	// line nor write a second one.
	line := fmt.Sprintf("%s %s %s %s %.3f\n", kind, strings.Join(strings.Fields(name), "_"), key[:min(12, len(key))], outcome, seconds.Seconds())
	if file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644); err == nil {
		file.WriteString(line)
		file.Close()
	}
}

// timed runs step and adds how long it took to total.
func timed(total *time.Duration, step func() error) error {
	started := time.Now()
	err := step()
	*total += time.Since(started)
	return err
}
