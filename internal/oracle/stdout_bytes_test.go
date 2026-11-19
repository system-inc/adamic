package oracle

import (
	"bytes"
	"path/filepath"
	"testing"
)

// #vndytkn: process.stdout.write of a Uint8Array writes its bytes as they are, on both backends, as Node does:
// every byte value and a sequence that as a string would spell a lone surrogate (bytes_all.a), the bytes as
// they stood at the call however the array changes after (mutate_after.a), nothing for an empty array, a
// view's own bytes (subarray.a), call order beside string writes and console.log (interleave.a), and a fresh
// array borrowed by the call and released after it, under the sanitizers with nothing leaked, from a function
// (temporary.a) and made in the argument itself (literal.a, Memory's should).
// Each fixture's Node output is pinned too, so a fixture can't pass by printing nothing. Mutants: the bytes
// written through write_text, as a string's are, and bytes_all.a prints EF BF BD natively where Node prints
// ED A0 80; one byte short, and every non-empty fixture differs; the write held back to the next one, and
// mutate_after.a writes the changed bytes.
func TestStdoutWritesBytes(t *testing.T) {
	t.Parallel()
	all := make([]byte, 0, 259)
	for value := 0; value < 256; value++ {
		all = append(all, byte(value))
	}
	all = append(all, 0xed, 0xa0, 0x80)
	for name, printed := range map[string][]byte{
		"bytes_all.a":    all,
		"mutate_after.a": {0x50, 0x80, 0xff, 0x0a, 0x41, 0x41, 0xff, 0x0a},
		"empty.a":        []byte("after\n"),
		"subarray.a":     []byte("BC\n"),
		"interleave.a":   []byte("a\nb\nc\nd\ne\n"),
		"temporary.a":    []byte("pq\nrs\ntu\n"),
		"literal.a":      []byte("a\nb\nc\n"),
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/stdout_bytes", name))
			if err != nil {
				t.Fatal(err)
			}
			node := onNode(t, path)
			if node.exitCode != 0 || !bytes.Equal(node.stdout, printed) || len(node.stderr) != 0 {
				t.Fatalf("Node: exit %d, stdout %q, stderr %q; want stdout %q", node.exitCode, node.stdout, node.stderr, printed)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatalf("a Uint8Array written to stdout must lower: %v", err)
			}
			sanitized, binary := natively(t, program)
			for backend, got := range map[string]run{"release": released(t, program), "sanitized": sanitized, "javascript": onJavaScriptBackend(t, program)} {
				if difference := disagreement(node, got); difference != "" {
					t.Fatalf("%s disagrees with Node (%s): Node %q; %s exit %d, stdout %q, stderr %q", backend, difference, node.stdout, backend, got.exitCode, got.stdout, got.stderr)
				}
			}
			if report := leaks(t, program, binary); report != "" {
				t.Fatal(report)
			}
		})
	}
}
