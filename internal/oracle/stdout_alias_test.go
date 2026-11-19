package oracle

import (
	"bytes"
	"path/filepath"
	"testing"
)

// #1trc7d3: a const alias of process.stdout writes and reads isTTY and columns as process.stdout does, on the
// release and sanitized builds and the JavaScript backend, held to Node byte for byte with nothing leaked:
// strings from the module and a function in call order beside console.log (alias_strings.a), a Uint8Array's
// bytes (alias_bytes.a, on #vndytkn), and the terminal reads, piped here (alias_tty.a). Each fixture's Node
// output is pinned, so none can pass by printing nothing.
func TestStdoutAlias(t *testing.T) {
	t.Parallel()
	for name, printed := range map[string][]byte{
		"alias_strings.a": []byte("a\nb\nc\nd\n"),
		"alias_bytes.a":   {0x80, 0xff, 0x0a, 'a', 'f', 't', 'e', 'r', '\n'},
		"alias_tty.a":     []byte("not a terminal\nno columns\n"),
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/stdout_alias", name))
			if err != nil {
				t.Fatal(err)
			}
			node := onNode(t, path)
			if node.exitCode != 0 || !bytes.Equal(node.stdout, printed) || len(node.stderr) != 0 {
				t.Fatalf("Node: exit %d, stdout %q, stderr %q; want stdout %q", node.exitCode, node.stdout, node.stderr, printed)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatalf("a const alias of process.stdout must lower: %v", err)
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
