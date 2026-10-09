// Command compare reports exact token-stream equality, hashes and the first difference.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type Stream struct {
	Bytes  int64  `json:"bytes"`
	Lines  int64  `json:"lines"`
	SHA256 string `json:"sha256"`
}
type Difference struct {
	Offset int64 `json:"offset"` // Zero-based byte offset.
	Line   int64 `json:"line"`   // One-based, measured in the common prefix.
	Column int64 `json:"column"` // One-based byte column, not UTF-16.
	Left   int   `json:"left"`   // -1 means EOF.
	Right  int   `json:"right"`
}
type Report struct {
	Equal bool        `json:"equal"`
	Left  Stream      `json:"left"`
	Right Stream      `json:"right"`
	First *Difference `json:"first_difference"`
}

// Compare reads fixed-size blocks, so long JSON token rows need no scanner limit.
// Hashes are evidence only: every byte and both EOFs decide equality.
func Compare(left, right io.Reader) (Report, error) {
	result := Report{Equal: true}
	lh, rh := sha256.New(), sha256.New()
	lb, rb := make([]byte, 65536), make([]byte, 65536)
	var offset int64
	line, column := int64(1), int64(1)
	for {
		ln, le := io.ReadFull(left, lb)
		rn, re := io.ReadFull(right, rb)
		if le != nil && le != io.EOF && le != io.ErrUnexpectedEOF {
			return Report{}, fmt.Errorf("left: %w", le)
		}
		if re != nil && re != io.EOF && re != io.ErrUnexpectedEOF {
			return Report{}, fmt.Errorf("right: %w", re)
		}
		lh.Write(lb[:ln])
		rh.Write(rb[:rn])
		result.Left.Bytes += int64(ln)
		result.Right.Bytes += int64(rn)
		for _, b := range lb[:ln] {
			if b == '\n' {
				result.Left.Lines++
			}
		}
		for _, b := range rb[:rn] {
			if b == '\n' {
				result.Right.Lines++
			}
		}
		n := max(ln, rn)
		for i := 0; i < n && result.First == nil; i++ {
			l, r := -1, -1
			if i < ln {
				l = int(lb[i])
			}
			if i < rn {
				r = int(rb[i])
			}
			if l != r {
				result.Equal = false
				result.First = &Difference{offset + int64(i), line, column, l, r}
			} else if l == '\n' {
				line++
				column = 1
			} else {
				column++
			}
		}
		offset += int64(n)
		if ln == 0 && rn == 0 {
			break
		}
	}
	result.Left.SHA256 = hex.EncodeToString(lh.Sum(nil))
	result.Right.SHA256 = hex.EncodeToString(rh.Sum(nil))
	return result, nil
}
func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: compare NODE_STDOUT NATIVE_STDOUT")
		os.Exit(2)
	}
	left, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	defer left.Close()
	right, err := os.Open(os.Args[2])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	defer right.Close()
	report, err := Compare(left, right)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	if err := json.NewEncoder(os.Stdout).Encode(report); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	if !report.Equal {
		os.Exit(1)
	}
}
