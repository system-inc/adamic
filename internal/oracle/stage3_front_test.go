package oracle

import (
	"os"
	"path/filepath"
	"testing"
)

// An explicit open-enum default handles unmatched numbers and matched fallthrough.
// Only an actual never read asserts unreachability. Counting the scrutinee's
// calls also holds the merged lowering to one evaluation.
func TestStage3EnumFallthrough(t *testing.T) {
	t.Parallel()
	const source = `enum Kind { A, B }
let reads = 0;
function input(value: number): Kind { reads++; return value; }
function describe(value: Kind): string {
    let text = '';
    switch (input(value)) {
        case Kind.A: text += 'A';
        default: text += 'D';
        case Kind.B: text += 'B';
    }
    return text;
}
function known(): string {
    const value: Kind = Kind.A;
    let text = '';
    switch (value) {
        case Kind.A: text += 'A';
        default: text += 'D';
    }
    return text;
}
console.log(describe(Kind.A));
console.log(describe(Kind.B));
console.log(known());
console.log(` + "`${reads}`" + `);
`
	for _, unmatched := range []bool{false, true} {
		name := "matched"
		programSource := source
		truth := run{stdout: []byte("ADB\nB\nAD\n2\n")}
		if unmatched {
			name = "unmatched"
			programSource += "console.log(describe(input(42)));\n"
			truth.stdout = append(append([]byte{}, truth.stdout...), []byte("DB\n")...)
		}
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), "enum_fallthrough.a")
			if err := os.WriteFile(path, []byte(programSource), 0600); err != nil {
				t.Fatal(err)
			}
			if got := onNode(t, path); disagreement(truth, got) != "" {
				t.Fatalf("source Node: %+v", got)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			native, binary := natively(t, program)
			for backend, got := range map[string]run{"native": native, "release": released(t, program), "javascript": onJavaScriptBackend(t, program)} {
				if difference := disagreement(truth, got); difference != "" {
					t.Errorf("%s: %s: %+v", backend, difference, got)
				}
			}
			if report := leaks(t, program, binary); report != "" {
				t.Fatal(report)
			}
		})
	}
}
