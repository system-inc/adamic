// Package dumpdiff locates the first byte-different record in the parser v2 dump.
package dumpdiff

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"
)

type Difference struct {
	Line          int
	ReferencePath string
	ActualPath    string
	Reference     []byte
	Actual        []byte
}

type cursor struct {
	file                    string
	node, diagnostic, jsdoc int
}

func (c *cursor) path(row []byte) string {
	text := string(row)
	switch {
	case strings.HasPrefix(text, "file\t"):
		c.file = strings.TrimSuffix(text[5:], "\n")
		c.node, c.diagnostic, c.jsdoc = 0, 0, 0
		return c.file + "/header"
	case strings.HasPrefix(text, "diagnostics "):
		return c.file + "/parseDiagnostics/length"
	case strings.HasPrefix(text, "jsDocDiagnostics "):
		return c.file + "/jsDocDiagnostics/length"
	case strings.HasPrefix(text, "diagnostic "):
		path := fmt.Sprintf("%s/parseDiagnostics/%d", c.file, c.diagnostic)
		c.diagnostic++
		return path
	case strings.HasPrefix(text, "jsDocDiagnostic "):
		path := fmt.Sprintf("%s/jsDocDiagnostics/%d", c.file, c.jsdoc)
		c.jsdoc++
		return path
	default:
		path := fmt.Sprintf("%s/preorder/%d", c.file, c.node)
		c.node++
		return path
	}
}

// Compare keeps only the current records. Paths use zero-based preorder indices,
// including attached JSDoc, exactly as the v2 driver visits them. The dump has no
// parent/depth information, so these are record paths, not invented AST edges.
// Every byte matters, including a missing final newline. Nil means equal.
func Compare(reference, actual io.Reader) (*Difference, error) {
	left, right := bufio.NewReader(reference), bufio.NewReader(actual)
	var lc, rc cursor
	for line := 1; ; line++ {
		l, le := left.ReadBytes('\n')
		r, re := right.ReadBytes('\n')
		if le != nil && le != io.EOF {
			return nil, fmt.Errorf("reference: %w", le)
		}
		if re != nil && re != io.EOF {
			return nil, fmt.Errorf("actual: %w", re)
		}
		lp, rp := "EOF", "EOF"
		if len(l) > 0 {
			lp = lc.path(l)
		}
		if len(r) > 0 {
			rp = rc.path(r)
		}
		if !bytes.Equal(l, r) {
			return &Difference{line, lp, rp, l, r}, nil
		}
		if le == io.EOF && re == io.EOF {
			return nil, nil
		}
	}
}
