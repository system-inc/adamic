package oracle

import (
	"github.com/system-inc/adamic/internal/ir"
	"strings"
	"testing"
)

var checkedJSONFixtures = []string{"json_object", "json_object_misfit", "json_list", "json_list_misfit", "json_validated", "structural", "json_literal", "json_literal_misfit", "json_literal_result", "json_literal_result_misfit", "json_boolean_list", "json_optional", "json_recovery", "json_domain_misfit", "json_union", "json_union_misfit", "json_nested_list", "json_nested_list_misfit", "json_boxed_list", "stock_decode_entities", "json_nul_literal"}

func TestCheckedJSON(t *testing.T) {
	for _, name := range checkedJSONFixtures {
		t.Run(name, func(t *testing.T) {
			program, path := checkedAnyProgram(t, name)
			truth := onNode(t, path)
			want := truth
			if name == "json_object_misfit" {
				want = run{exitCode: 70, stderr: []byte("adamic: panic: checked any: raw.options.limit needs number, found string\n")}
			}
			if name == "json_list_misfit" {
				want = run{exitCode: 70, stderr: []byte("adamic: panic: checked any: value[0] needs number, found string\n")}
			}
			if name == "json_literal_misfit" {
				want = run{exitCode: 70, stderr: []byte("adamic: panic: checked any: value.type needs \"list\", found string\n")}
			}
			if name == "json_literal_result_misfit" {
				want = run{exitCode: 70, stderr: []byte("adamic: panic: checked any: value + 1 needs 1, found number\n")}
			}
			if name == "json_domain_misfit" {
				want = run{exitCode: 70, stderr: []byte("adamic: panic: checked any: raw.number needs JSON value | undefined, found non-JSON number\n")}
			}
			if name == "json_union_misfit" {
				want = run{exitCode: 70, stderr: []byte("adamic: panic: checked any: value.type needs \"list\" | \"object\", found string\n")}
			}
			if name == "json_nested_list_misfit" {
				want = run{exitCode: 70, stderr: []byte("adamic: panic: checked any: raw.paths[0].path needs string, found number\n")}
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
			if strings.HasSuffix(name, "_misfit") {
				removed := 0
				for index := range program.Functions {
					body := program.Functions[index].Body
					for at, statement := range body {
						returned, ok := statement.(ir.Return)
						if !ok {
							continue
						}
						check, ok := returned.Value.(ir.CheckedJSON)
						if !ok {
							continue
						}
						returned.Value = ir.Narrow{Value: check.Value, To: check.Of}
						body[at] = returned
						removed++
					}
				}

				if removed == 0 {
					t.Fatal("mutant removed no structural check")
				}
				// Read guards intentionally remain, so the exact path pins the initial recursive check.
				mutant := onJavaScriptBackend(t, program)
				if disagreement(want, mutant) == "" {
					t.Fatal("missing recursive guard escaped pinned path")
				}
				mutatedNative, _ := nativelyUncached(t, program)
				if disagreement(want, mutatedNative) == "" {
					t.Fatal("native recursive mutant escaped pinned path")
				}
				t.Logf("mutant removing recursive check caught: JavaScript %+v; native %+v", mutant, mutatedNative)
			}
			t.Logf("source Node %+v; checked backends %+v", truth, want)
		})
	}
}
