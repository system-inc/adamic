package lower

import (
	"strings"
	"testing"
)

func TestCensusOverloadReturnProof(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		name, source string
		checked      bool
	}{
		{"parameter return", `function choose(value: string): string;
function choose(value: string | undefined): string | undefined { return value; }
console.log(choose('yes'));`, false},
		{"lying return", `function choose(value: string): string;
function choose(value: string | undefined): string | undefined { return undefined; }
console.log(choose('yes'));`, true},
		{"rebound parameter", `function choose(value: string): string;
function choose(value: string | undefined): string | undefined { value = undefined; return value; }
console.log(choose('yes'));`, true},
		{"captured write", `function choose(value: string): string;
function choose(value: string | undefined): string | undefined { const clear = (): void => { value = undefined; }; clear(); return value; }
console.log(choose('yes'));`, true},
		{"empty return", `function choose(value: string): string;
function choose(value: string | undefined): string | undefined { if (value === 'yes') return value; return; }
console.log(choose('no'));`, true},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			program, err := lowerSource(t, probe.source)
			if err != nil {
				t.Fatal(err)
			}
			checked := false
			for _, constant := range program.Strings {
				checked = checked || strings.Contains(constant, "overload 1 of choose result")
			}
			if checked != probe.checked {
				t.Fatalf("result check = %t, want %t", checked, probe.checked)
			}
		})
	}
}
