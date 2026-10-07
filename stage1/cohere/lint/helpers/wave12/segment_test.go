package wave12

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestSegmentMatchesGo(t *testing.T) {
	entry, _ := filepath.Abs("segment_main.a")
	corpus, _ := filepath.Abs("testdata/segment_cases.json")
	want := execute(t, "", upstreamFor(t, "segment_oracle.go.txt"), corpus)
	for i, got := range backends(t, entry, corpus) {
		if !bytes.Equal(got, want) {
			t.Fatalf("backend %d differs at row %d", i, difference(got, want))
		}
	}
	t.Logf("%d cases and %d bytes match actual Go, Node, sanitized native and emitted JavaScript", bytes.Count(want, []byte{'\n'}), len(want))
}
func TestSegmentMutants(t *testing.T) {
	corpus, _ := filepath.Abs("testdata/segment_cases.json")
	want := execute(t, "", upstreamFor(t, "segment_oracle.go.txt"), corpus)
	for _, mutant := range []struct{ name, old, replacement string }{
		{"drop-final-empty", "parts.push(input.slice(lastPosition));", "if(lastPosition < input.length) { parts.push(input.slice(lastPosition)); }"},
		{"pop-unmatched-closer", "closing.length > 0 && character === closing[closing.length - 1]", "closing.length > 0"},
		{"escape-disabled", "if(character === '\\\\') { index++; }", "if(character === '\\\\') { index += 0; }"},
	} {
		t.Run(mutant.name, func(t *testing.T) {
			parent := t.TempDir()
			directory := filepath.Join(parent, "wave12")
			if err := os.Mkdir(directory, 0755); err != nil {
				t.Fatal(err)
			}
			for _, file := range []string{"segment_main.a", "collapse_segment.a", "../options_json.ts"} {
				data, err := os.ReadFile(file)
				if err != nil {
					t.Fatal(err)
				}
				if file == "collapse_segment.a" {
					if strings.Count(string(data), mutant.old) != 1 {
						t.Fatal("mutant anchor changed")
					}
					data = []byte(strings.Replace(string(data), mutant.old, mutant.replacement, 1))
				}
				if err := os.WriteFile(filepath.Join(directory, file), data, 0644); err != nil {
					t.Fatal(err)
				}
			}
			for i, got := range backends(t, filepath.Join(directory, "segment_main.a"), corpus) {
				if bytes.Equal(got, want) {
					t.Fatalf("backend %d mutant survived", i)
				}
				t.Logf("backend %d compiled mutant caught only by comparison at row %d", i, difference(got, want))
			}
		})
	}
}

func TestSegmentSeparatorRefusals(t *testing.T) {
	entry, _ := filepath.Abs("segment_main.a")
	directory := t.TempDir()
	corpus := filepath.Join(directory, "cases.json")
	if err := os.WriteFile(corpus, []byte(`[{"input":"a,b","separator":128}]`), 0644); err != nil {
		t.Fatal(err)
	}
	commands := backendCommands(t, entry, corpus)
	expected := "adamic: panic: NotYet: CSS segment separator must be an ASCII byte\n"
	for _, separator := range []float64{-1, 128, 255, 0.5} {
		data, _ := json.Marshal([]map[string]any{{"input": "a,b", "separator": separator}})
		if err := os.WriteFile(corpus, data, 0644); err != nil {
			t.Fatal(err)
		}
		for i, command := range commands {
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
			process := exec.CommandContext(ctx, command[0], command[1:]...)
			var output, errors bytes.Buffer
			process.Stdout = &output
			process.Stderr = &errors
			err := process.Run()
			cancel()
			exit, ok := err.(*exec.ExitError)
			if !ok || exit.ExitCode() != 70 || output.Len() != 0 || errors.String() != expected {
				t.Fatalf("backend %d separator %v: exit %v output %q stderr %q", i, separator, err, &output, &errors)
			}
		}
	}
	t.Log("four unsupported separator inputs refuse identically on all three backends")
	parent := t.TempDir()
	mutated := filepath.Join(parent, "wave12")
	if err := os.Mkdir(mutated, 0755); err != nil {
		t.Fatal(err)
	}
	for _, file := range []string{"segment_main.a", "collapse_segment.a", "../options_json.ts"} {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		if file == "collapse_segment.a" {
			anchor := "separator < 0 || separator > 127 || !Number.isInteger(separator)"
			if strings.Count(string(data), anchor) != 1 {
				t.Fatal("guard mutant anchor changed")
			}
			data = []byte(strings.Replace(string(data), anchor, "false", 1))
		}
		if err := os.WriteFile(filepath.Join(mutated, file), data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	for i, got := range backends(t, filepath.Join(mutated, "segment_main.a"), corpus) {
		if len(got) == 0 {
			t.Fatal("compiled guard mutant produced no observation")
		}
		t.Logf("backend %d compiled separator-guard mutant finishes, caught only by refusal comparison", i)
	}
}
