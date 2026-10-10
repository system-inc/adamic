package oracle

import (
	"context"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

var checkedJSONNextFixtures = []string{"json_recursive", "json_recursive_misfit", "json_depth_misfit", "json_dictionary", "json_dictionary_misfit", "json_nullable", "json_nullable_misfit", "json_nullable_validated", "json_nullable_array", "json_nullable_array_misfit", "json_dictionary_order_misfit"}

func TestCheckedJSONNext(t *testing.T) {
	messages := map[string]string{
		"json_dictionary_order_misfit": "raw.1 needs number, found string",
		"json_nullable_array_misfit":   "raw[0] needs number, found string",
		"json_recursive_misfit":        "raw.next.value needs number, found string",
		"json_depth_misfit":            "raw" + strings.Repeat(".next", 64) + ".value needs JSON value | undefined, found non-JSON recursion",
		"json_dictionary_misfit":       "raw.beta needs number, found string",
		"json_nullable_misfit":         "raw.option needs string | null | undefined, found number",
	}
	for _, name := range checkedJSONNextFixtures {
		t.Run(name, func(t *testing.T) {
			program, path := checkedAnyProgram(t, name)
			truth := onNode(t, path)
			want := truth
			if message := messages[name]; message != "" {
				want = run{exitCode: 70, stderr: []byte("adamic: panic: checked any: " + message + "\n")}
			}
			js := onJavaScriptBackend(t, program)
			if d := disagreement(want, js); d != "" {
				t.Fatalf("JavaScript %s: got %+v want %+v", d, js, want)
			}
			observed, binary := nativelyUncached(t, program)
			if d := disagreement(want, observed); d != "" {
				t.Fatalf("native %s: got %+v want %+v", d, observed, want)
			}
			if want.exitCode == 0 {
				if report := leaksUncached(t, program, binary); report != "" {
					t.Fatal(report)
				}
			}
			if messages[name] != "" {
				removed := 0
				for index := range program.Functions {
					for at, statement := range program.Functions[index].Body {
						returned, ok := statement.(ir.Return)
						if !ok {
							continue
						}
						check, ok := returned.Value.(ir.CheckedJSON)
						if !ok {
							continue
						}
						returned.Value = ir.Narrow{Value: check.Value, To: check.Of}
						program.Functions[index].Body[at] = returned
						removed++
					}
				}
				if removed != 1 {
					t.Fatalf("mutant must remove precisely the point-of-use check, got %d", removed)
				}
				mutant := onJavaScriptBackend(t, program)
				if d := disagreement(truth, mutant); d != "" {
					t.Fatalf("unchecked JavaScript must match Node, %s: %+v", d, mutant)
				}
				mutatedNative, _ := nativelyUncached(t, program)
				if d := disagreement(truth, mutatedNative); d != "" {
					t.Fatalf("unchecked native must match Node, %s: %+v", d, mutatedNative)
				}
				if disagreement(want, mutant) == "" || disagreement(want, mutatedNative) == "" {
					t.Fatal("missing boundary check escaped pinned observation")
				}
				t.Logf("point-of-use mutant caught in both backends; no later typed read: Node %+v", truth)
			}
			t.Logf("Node %+v; checked backends %+v", truth, want)
		})
	}
}

func TestCheckedJSONNextRefusals(t *testing.T) {
	for _, name := range []string{"json_dictionary_alias_write", "json_dictionary_nul_key", "json_dictionary_proto_key", "json_dictionary_computed_key", "json_symbol_contract", "callable", "json_callable_contract", "json_maplike_pending"} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(repository, "internal/load/testdata/0.1/refuse/checked_any", name+".a")
			path, absoluteError := filepath.Abs(path)
			if absoluteError != nil {
				t.Fatal(absoluteError)
			}
			source, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			temporary := filepath.Join(t.TempDir(), "boundary.ts")
			if err := os.WriteFile(temporary, source, 0600); err != nil {
				t.Fatal(err)
			}
			loaded, err := load.Load([]string{temporary})
			if err != nil {
				t.Fatal(err)
			}
			program, err := lower.Lower(context.Background(), loaded)
			if err == nil || program != nil {
				t.Fatal("unsupported contract or alias mutation became accepted")
			}
			if name == "json_dictionary_alias_write" && !strings.Contains(err.Error(), "checked JSON view alias") {
				t.Fatalf("wrong point of refusal: %v", err)
			}
			if name == "json_dictionary_nul_key" && !strings.Contains(err.Error(), "NUL key in a checked JSON dictionary") {
				t.Fatalf("wrong point of refusal: %v", err)
			}
			if name == "json_callable_contract" && !strings.Contains(err.Error(), "checked JSON callable contract") {
				t.Fatalf("wrong callable refusal: %v", err)
			}
			for witness, reason := range map[string]string{"json_dictionary_proto_key": "prototype setter in a checked JSON dictionary", "json_dictionary_computed_key": "computed key in a checked JSON dictionary", "json_symbol_contract": "symbol field in a checked JSON contract"} {
				if name == witness && !strings.Contains(err.Error(), reason) {
					t.Fatalf("wrong point of refusal: %v", err)
				}
			}
			if name == "json_maplike_pending" {
				truth := onNode(t, path)
				if truth.exitCode != 0 || string(truth.stdout) != "true\n" || len(truth.stderr) != 0 {
					t.Fatalf("stock semantic witness: %+v", truth)
				}
				t.Skipf("awaits compiler/records-maplike: any-valued dictionary, current stop: %v", err)
			}
			t.Logf("%v", err)
		})
	}
}
