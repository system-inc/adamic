package lower

import (
	"strings"
	"testing"
)

func TestCensusOverloadReturnProof(test *testing.T) {
	test.Parallel()
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
		test.Run(probe.name, func(test *testing.T) {
			test.Parallel()
			program, lowerSourceError := lowerSource(test, probe.source)
			if lowerSourceError != nil {
				test.Fatal(lowerSourceError)
			}
			checked := false
			for _, constant := range program.Strings {
				checked = checked || strings.Contains(constant, "overload 1 of choose result")
			}
			if checked != probe.checked {
				test.Fatalf("result check = %t, want %t", checked, probe.checked)
			}
		})
	}
}
