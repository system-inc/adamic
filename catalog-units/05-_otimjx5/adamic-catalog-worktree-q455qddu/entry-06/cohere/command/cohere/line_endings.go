package main

import (
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/edit"
)

// crlfFiles counts the files the formatter was handed with CRLF line endings.
//
// The house format is LF, and formatting is never customized (Kirk's ruling), so a CRLF file is
// rewritten by `--fix` and reported by `--no-fix`, which is right. What a reader can't see is why every
// file in a fresh Windows checkout is suddenly unformatted: git's core.autocrlf, the default there,
// writes CRLF on checkout and LF on commit, so the files look changed only to cohere. One note says so
// and names the fix, rather than the formatter quietly fighting git on every run.
type crlfFiles struct {
	mutex sync.Mutex
	files map[string]struct{}
}

// observing wraps the format transform so it counts each file whose text arrives with CRLF. The format
// pass runs in parallel, so the count is guarded.
func (c *crlfFiles) observing(transform edit.Transform) edit.Transform {
	return func(fileName string, text string, parsed *ast.SourceFile) (string, error) {
		if strings.Contains(text, "\r\n") {
			c.mutex.Lock()
			if c.files == nil {
				c.files = map[string]struct{}{}
			}
			c.files[fileName] = struct{}{}
			c.mutex.Unlock()
		}
		return transform(fileName, text, parsed)
	}
}

// has reports whether fileName arrived with CRLF. A nil set holds none.
func (c *crlfFiles) has(fileName string) bool {
	if c == nil {
		return false
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	_, held := c.files[fileName]
	return held
}

// crlfReason is what a CRLF file's would-change finding adds, so the cause and its one-line fix reach a
// default run, where the note above does not (#dr78rt8).
const crlfReason = ". It has CRLF line endings, and the house format is LF. If git converts them on checkout " +
	"(core.autocrlf, the default on Windows), add `* text=auto eol=lf` to the repository's .gitattributes"

// report prints the note once, after the fix phase, when any file arrived with CRLF.
func (c *crlfFiles) report(out io.Writer, write bool) {
	if len(c.files) == 0 {
		return
	}
	noun, them := "file has", "it"
	if len(c.files) != 1 {
		noun, them = "files have", "them"
	}
	consequence := "--no-fix reports " + them + " as unformatted"
	if write {
		consequence = "--fix rewrote " + them + " to LF"
	}
	fmt.Fprintf(out, "note: %d %s CRLF line endings, and the house format is LF, so %s. "+
		"If git converts them on checkout (core.autocrlf, the default on Windows), add `* text=auto eol=lf` "+
		"to the repository's .gitattributes so the checkout is LF too\n", len(c.files), noun, consequence)
}
