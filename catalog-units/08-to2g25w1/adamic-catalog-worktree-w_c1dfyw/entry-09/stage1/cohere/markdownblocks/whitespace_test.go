package markdownblocks

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// These are typed path facts supplied by the external parser, not a native parser claim.
// Six independent policy observations cover preserve/always/never, ordinary/link modes.
func testWhitespacePolicy(t *testing.T, cases, fork string) {
	t.Helper()
	raw, err := os.ReadFile(cases)
	if err != nil {
		t.Fatal(err)
	}
	var expected strings.Builder
	count := 0
	for _, line := range strings.Split(string(raw), "\n") {
		if !strings.HasPrefix(line, "S\t") {
			continue
		}
		fields := strings.Split(line, "\t")
		expected.WriteString(fields[22] + "\n")
		count++
	}
	want := []byte(expected.String())
	main, err := filepath.Abs("testdata/whitespace_probe.ts")
	if err != nil {
		t.Fatal(err)
	}
	program := lowered(t, main)
	answer, binary := natively(t, program, cases)
	for _, side := range []struct {
		name   string
		result run
	}{
		{"whitespace actual original fork", execute(t, nil, "node", "testdata/whitespace_library.mjs", fork, cases)},
		{"whitespace native", answer}, {"whitespace source Node", onNode(t, main, cases)}, {"whitespace backend", onJavaScriptBackend(t, program, cases)},
	} {
		clean(t, side.name, side.result)
		equal(t, side.name, side.result.stdout, want)
	}
	if report := leaks(t, program, binary, cases); report != "" {
		t.Fatal(report)
	}
	for _, m := range []struct{ name, from, to string }{
		{"policy CJ spacing tie", "return spaces > empties", "return spaces >= empties"},
		{"policy Korean pair", "return (previous === 'k-letter'", "return (previous === 'non-cjk'"},
		{"policy single line", "kind === 'tableCell' ||", "kind === 'missing-tableCell' ||"},
	} {
		t.Run(m.name, func(t *testing.T) {
			scratch := t.TempDir()
			if err := os.Mkdir(filepath.Join(scratch, "testdata"), 0755); err != nil {
				t.Fatal(err)
			}
			for _, file := range []string{"whitespace.ts", "whitespaceCodec.ts", "codec.ts", "document.ts", "commandStack.ts", "width.ts", "widthTables.ts", "widthRuneRanges.ts", "emojiMatcher.ts", "testdata/whitespace_probe.ts"} {
				b, err := os.ReadFile(file)
				if err != nil {
					t.Fatal(err)
				}
				if file == "whitespace.ts" {
					if strings.Count(string(b), m.from) != 1 {
						t.Fatal("mutation anchor")
					}
					b = []byte(strings.Replace(string(b), m.from, m.to, 1))
				}
				write(t, filepath.Join(scratch, file), b)
			}
			result := onNode(t, filepath.Join(scratch, "testdata/whitespace_probe.ts"), cases)
			clean(t, m.name, result)
			if bytes.Equal(result.stdout, want) {
				t.Fatal("survived")
			}
			t.Log("caught by Go policy document classification")
		})
	}
	t.Logf("%d whitespace nodes, six modes each, match Go and unchanged original private fork functions", count)
}
