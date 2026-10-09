package dispatch

import (
	"fmt"
	"io"
	"os"
	"sync"
	"unicode/utf8"
)

// Progress is what the launcher says on its own while the caller waits, apart from cohere's report (#ytqqv8v).
// A default run ends in one footer line, and the launcher's lines would print above it, so each kind says
// only as much as the caller needs:
//
//   - A step, such as building HEAD, extracting the compiler or building the Swift engine, takes seconds
//     to minutes, and silence through that would look hung. On a terminal it is one dim line, rewritten by
//     the next step and cleared before cohere prints, and to a pipe or a file it is nothing.
//   - A note, such as what a prune or a trim removed, is housekeeping already written to the prune log.
//     It prints nowhere.
//   - With --verbose (or --dispatch-verbose) both print in full, one line each, as they always did.
//
// A failure is none of these: it prints as it always did, after clearing a step that was showing.
type Progress struct {
	lock     sync.Mutex
	out      io.Writer
	verbose  bool
	terminal bool
	color    bool
	// showing is whether a step's line is on the screen, waiting to be cleared or rewritten.
	showing bool
}

// Report is the launcher's progress, to stderr. NewProgress makes one a test can read.
var Report = NewProgress(os.Stderr, false, isTerminal(os.Stderr))

// NewProgress reports to out, in full when verbose, and as one rewritten line when out is a terminal.
func NewProgress(out io.Writer, verbose bool, terminal bool) *Progress {
	_, noColor := os.LookupEnv("NO_COLOR")
	return &Progress{out: out, verbose: verbose, terminal: terminal, color: terminal && !noColor}
}

// SetVerbose turns full lines on, for --verbose and --dispatch-verbose.
func (progress *Progress) SetVerbose(verbose bool) {
	progress.lock.Lock()
	defer progress.lock.Unlock()
	progress.verbose = verbose
}

// stepWidth is how much of a step a terminal line shows. A line wider than the terminal wraps, and a
// carriage return clears only the row it is on, so a step is kept to one row of an ordinary terminal.
const stepWidth = 78

// Step says what the launcher is doing now.
func (progress *Progress) Step(format string, arguments ...any) {
	progress.lock.Lock()
	defer progress.lock.Unlock()
	text := fmt.Sprintf(format, arguments...)
	switch {
	case progress.verbose:
		progress.clearLocked()
		fmt.Fprintln(progress.out, text)
	case progress.terminal:
		if utf8.RuneCountInString(text) > stepWidth {
			runes := []rune(text)
			text = string(runes[:stepWidth-1]) + "…"
		}
		if progress.color {
			text = "\x1b[2m" + text + "\x1b[22m"
		}
		fmt.Fprint(progress.out, "\r\x1b[2K"+text)
		progress.showing = true
	}
}

// Note says something only a reader of --verbose wants; the prune log already holds it.
func (progress *Progress) Note(format string, arguments ...any) {
	progress.lock.Lock()
	defer progress.lock.Unlock()
	if progress.verbose {
		progress.clearLocked()
		fmt.Fprintf(progress.out, format+"\n", arguments...)
	}
}

// Fail says what went wrong, always, on a line of its own.
func (progress *Progress) Fail(format string, arguments ...any) {
	progress.lock.Lock()
	defer progress.lock.Unlock()
	progress.clearLocked()
	fmt.Fprintf(progress.out, format+"\n", arguments...)
}

// ToolOutput is where a tool the launcher runs, such as SwiftPM, should write its own output: the stream
// under --verbose, after clearing a step, and nil otherwise, which keeps it for a failure to quote.
func (progress *Progress) ToolOutput() io.Writer {
	progress.lock.Lock()
	defer progress.lock.Unlock()
	if !progress.verbose {
		return nil
	}
	progress.clearLocked()
	return progress.out
}

// Clear removes a step's line, so what prints next starts on a clean row. The launcher calls it before
// cohere runs, and cohere before it reports.
func (progress *Progress) Clear() {
	progress.lock.Lock()
	defer progress.lock.Unlock()
	progress.clearLocked()
}

func (progress *Progress) clearLocked() {
	if progress.showing {
		fmt.Fprint(progress.out, "\r\x1b[2K")
		progress.showing = false
	}
}

// isTerminal reports whether a stream is a terminal rather than a pipe or a file.
func isTerminal(stream *os.File) bool {
	information, err := stream.Stat()
	return err == nil && information.Mode()&os.ModeCharDevice != 0
}
