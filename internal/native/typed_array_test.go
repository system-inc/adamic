package native

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestTypedArrayRuntime(t *testing.T) {
	t.Parallel()
	source, err := os.ReadFile("testdata/typed-arrays/runtime.c")
	if err != nil {
		t.Fatal(err)
	}
	expected, err := exec.Command("node", "testdata/typed-arrays/node.a").CombinedOutput()
	if err != nil {
		t.Fatalf("Node: %v\n%s", err, expected)
	}
	for _, sanitized := range []bool{false, true} {
		name := "release"
		if sanitized {
			name = "sanitized"
		}
		t.Run(name, func(t *testing.T) {
			binary := filepath.Join(t.TempDir(), "typed-arrays")
			if err := Build(string(source), binary, Options{Sanitize: sanitized, Count: true}); err != nil {
				t.Fatal(err)
			}
			command := exec.Command(binary)
			command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1")
			var diagnostics bytes.Buffer
			command.Stderr = &diagnostics
			got, err := command.Output()
			if err != nil {
				t.Fatalf("runtime: %v\n%s", err, diagnostics.String())
			}
			if string(got) != string(expected) {
				left, right := strings.Split(string(got), "\n"), strings.Split(string(expected), "\n")
				for index := 0; index < len(left) && index < len(right); index++ {
					if left[index] != right[index] {
						t.Fatalf("runtime differs from Node at line %d: got %q want %q", index+1, left[index], right[index])
					}
				}
				t.Fatalf("runtime differs from Node: %d lines instead of %d", len(left), len(right))
			}
			for _, mode := range []string{"write", "check"} {
				for _, index := range []string{"3", "-1", "0.5", "NaN", "Infinity", "-Infinity"} {
					got, err := exec.Command(binary, mode, index).CombinedOutput()
					failure, ok := err.(*exec.ExitError)
					if !ok || failure.ExitCode() != 70 {
						t.Fatalf("%s %s: wanted panic 70, got %v\n%s", mode, index, err, got)
					}
					want := "adamic: panic: index " + index + " is outside an array of length 3\n"
					if !strings.HasPrefix(string(got), want) {
						t.Fatalf("%s %s: got %q want %q", mode, index, got, want)
					}
				}
			}
			// Node decides which numeric constructor and set arguments are RangeErrors.
			const rangeOracle = `const [mode, text] = process.argv.slice(1); const value = Number(text);
try { if (mode === 'length') { new Uint8Array(value); } else { const a = new Uint8Array(3); a.set(a, value); } }
catch (error) { console.log(error.name); }`
			for _, probe := range [][]string{{"length", "-1"}, {"length", "Infinity"}, {"length", "9007199254740992"}, {"offset", "-1"}, {"offset", "1"}, {"offset", "Infinity"}, {"kind"}} {
				if probe[0] != "kind" {
					oracle, err := exec.Command("node", "-e", rangeOracle, "--", probe[0], probe[1]).CombinedOutput()
					if err != nil || string(oracle) != "RangeError\n" {
						t.Fatalf("Node range probe %v: %v %q", probe, err, oracle)
					}
				}
				got, err := exec.Command(binary, probe...).CombinedOutput()
				failure, ok := err.(*exec.ExitError)
				if !ok || failure.ExitCode() != 70 || !strings.HasPrefix(string(got), "adamic: panic: ") {
					t.Fatalf("%v: wanted panic 70, got %v\n%s", probe, err, got)
				}
			}
		})
	}
}
