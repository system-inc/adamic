//go:build wave28landing

// Compiled into the unchanged lint harness through a Go overlay. Each captured
// case owns a subtest, so a dependency refusal does not hide later rules.
package lint

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func wave28CapturedCounts() map[string]int {
	return map[string]int{
		"base/correctness-require-verify-optional-parity": 22,
		"block-scoped-var": 109,
		"getter-return":    106,
		"nexus/correctness-no-uncleared-race-timeout": 21,
		"no-throw-literal":         48,
		"no-useless-backreference": 408,
		"prefer-arrow-callback":    83,
		"react/jsx-fragments":      46,
		"react/jsx-no-undef":       45,
	}
}

func TestWave28EveryCapturedCase(t *testing.T) {
	rows := upstream(t)
	want := wave28CapturedCounts()
	counts := map[string]int{}
	for _, row := range rows {
		fields := strings.Split(row, "\t")
		if _, owned := want[fields[1]]; owned {
			counts[fields[1]]++
		}
	}
	if os.Getenv("ADAMIC_WAVE28_CAPTURE_COUNT_MUTANT") == "1" {
		want["base/correctness-require-verify-optional-parity"]++
	}
	for name, expected := range want {
		if counts[name] != expected {
			t.Fatalf("%s captured %d cases, want %d", name, counts[name], expected)
		}
	}
	directory, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	oracle := goOracle(t)
	var selected []string
	for _, row := range rows {
		if _, owned := want[strings.Split(row, "\t")[1]]; owned {
			selected = append(selected, row)
		}
	}
	// The Go parser, not this test, decides which unchanged inputs need recovery.
	rows = recoveryRows(t, oracle, selected)
	binary := buildPort(t, directory, true)
	module := emittedJavaScript(t, directory)
	for _, name := range []string{
		"base/correctness-require-verify-optional-parity", "block-scoped-var", "getter-return",
		"nexus/correctness-no-uncleared-race-timeout", "no-throw-literal", "no-useless-backreference",
		"prefer-arrow-callback", "react/jsx-fragments", "react/jsx-no-undef",
	} {
		t.Run(name, func(t *testing.T) {
			index := 0
			for _, row := range rows {
				fields := strings.Split(row, "\t")
				if fields[1] != name {
					continue
				}
				index++
				t.Run(fmt.Sprintf("case-%03d", index), func(t *testing.T) {
					config := filepath.Join(t.TempDir(), "tsconfig.json")
					options := fmt.Sprintf(`{"compilerOptions":{"strict":true},"files":[%q]}`, fields[0])
					if err := os.WriteFile(config, []byte(options), 0644); err != nil {
						t.Fatal(err)
					}
					compareWithJavaScript(t, oracle, binary, directory, manifest(t, []string{"program " + config, row}), module)
				})
			}
		})
	}
}

// The upstream rule deliberately tolerates this recovered input. Reuse the
// oracle's existing diagnostics-to-recovery selector; do not alter its source.
func TestWave28RecoveredThrow(t *testing.T) {
	oracle := goOracle(t)
	rows := upstream(t)
	var selected []string
	for _, row := range rows {
		fields := strings.Split(row, "\t")
		if fields[1] != "no-throw-literal" {
			continue
		}
		source, err := os.ReadFile(fields[0])
		if err != nil {
			t.Fatal(err)
		}
		if string(source) == "function f() { throw; }\n" {
			selected = append(selected, row)
		}
	}
	if len(selected) != 1 {
		t.Fatalf("recovered throw capture: got %d rows, want 1", len(selected))
	}
	recovered := recoveryRows(t, oracle, selected)
	fields := strings.Split(recovered[0], "\t")
	if fields[6] != "recovery" {
		t.Fatalf("Go did not mark this malformed input for recovery: %q", recovered[0])
	}
	directory, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	config := filepath.Join(t.TempDir(), "tsconfig.json")
	options := fmt.Sprintf(`{"compilerOptions":{"strict":true},"files":[%q]}`, fields[0])
	if err := os.WriteFile(config, []byte(options), 0644); err != nil {
		t.Fatal(err)
	}
	compareWithJavaScript(t, oracle, buildPort(t, directory, true), directory,
		manifest(t, []string{"program " + config, recovered[0]}), emittedJavaScript(t, directory))
}

// Recheck every upstream input that the Go parser marks for recovery. The first
// all-case run remains evidence of the shared typed-row recovery gap.
func TestWave28RecoveredCases(t *testing.T) {
	oracle := goOracle(t)
	want := wave28CapturedCounts()
	var selected []string
	for _, row := range upstream(t) {
		if _, owned := want[strings.Split(row, "\t")[1]]; owned {
			selected = append(selected, row)
		}
	}
	rows := recoveryRows(t, oracle, selected)
	directory, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	binary := buildPort(t, directory, true)
	module := emittedJavaScript(t, directory)
	recovered := 0
	for index, row := range rows {
		fields := strings.Split(row, "\t")
		if len(fields) <= 6 || fields[6] != "recovery" {
			continue
		}
		recovered++
		t.Run(fmt.Sprintf("%s/row-%03d", fields[1], index+1), func(t *testing.T) {
			config := filepath.Join(t.TempDir(), "tsconfig.json")
			options := fmt.Sprintf(`{"compilerOptions":{"strict":true},"files":[%q]}`, fields[0])
			if err := os.WriteFile(config, []byte(options), 0644); err != nil {
				t.Fatal(err)
			}
			compareWithJavaScript(t, oracle, binary, directory, manifest(t, []string{"program " + config, row}), module)
		})
	}
	if recovered == 0 {
		t.Fatal("upstream recovery control captured no recovered cases")
	}
	t.Logf("Go-selected recovered cases: %d", recovered)
}
