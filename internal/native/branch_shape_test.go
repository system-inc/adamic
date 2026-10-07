package native

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

// The runtime inputs include values a lossy conversion would confuse with a case.
// Breaks, continues and owned case locals must keep their original scope behavior.
func TestNumericSwitchMatchesNode(t *testing.T) {
	t.Parallel()
	const source = `function select(value: number): string {
    switch (value) {
        case 4: return 'four';
        case 0: return 'zero';
        case 1: return 'one';
        case 2: return 'two';
        case 3: return 'three';
        case 2147483647: return 'maximum';
        case 5: return 'five';
        case 0: return 'duplicate';
        default: return 'other';
    }
}
for (const value of [-2, 0, -0, 1, 2, 3, 2147483647, -2147483648, 1.5, -1.5, 2147483648, -2147483649, NaN, Infinity, -Infinity, 1e100]) {
    console.log(select(value));
}
let result = '';
for (let index = 0; index < 5; index += 1) {
    switch (index) {
        case 0: result += 'a'; break;
        case 1: continue;
        case 2: { const label = 'x'.repeat(index); result += label; break; }
        case 3: result += 'b'; break;
        default: result += 'c'; break;
    }
}
console.log(result);
`
	directory := t.TempDir()
	path := filepath.Join(directory, "switch.a")
	if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	loaded, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	program, err := lower.Lower(context.Background(), loaded)
	if err != nil {
		t.Fatal(err)
	}
	generated := C(program)
	if !strings.Contains(generated, "switch (") {
		t.Fatal("numeric cases still emit comparison chains")
	}
	want := runWithInput(t, source, "node", "--input-type=module-typescript")
	for _, sanitize := range []bool{false, true} {
		binary := filepath.Join(directory, "switch")
		if err := Build(generated, binary, Options{Sanitize: sanitize}); err != nil {
			t.Fatal(err)
		}
		if got := runWithInput(t, "", binary); got != want {
			t.Fatalf("sanitize %v: native %q; Node %q", sanitize, got, want)
		}
	}
}
