package oracle

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestCheckedViewCallableOriginalMixedCallbackABI(t *testing.T) {
	for _, fixture := range []struct{ name, out string }{{"mixed_union_callback", "WORD\n7\nmissing\n9\nRETURNED\n10\nARROW\n0,1\n"}, {"mixed_union_callback_variance", "5\n"}} {
		t.Run(fixture.name, func(t *testing.T) {
			path, err := filepath.Abs(checkedViewFixturePath(filepath.Join(repository+"/internal/lower/testdata/predicates", fixture.name+".a")))
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			if truth.exitCode != 0 || string(truth.stdout) != fixture.out {
				t.Fatalf("Node %#v", truth)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			sanitized, binary := nativelyUncached(t, program)
			for _, got := range []run{releasedUncached(t, program), sanitized, onJavaScriptBackend(t, program)} {
				if difference := disagreement(truth, got); difference != "" {
					t.Fatal(difference)
				}
			}
			if report := leaksUncached(t, program, binary); report != "" {
				t.Fatal(report)
			}
		})
	}
}

// A supported scalar helper read cannot erase its receiver's original never obligation.
func TestCheckedViewCallableRetainsNeverWiderHelper(t *testing.T) {
	directory := os.Getenv("ADAMIC_BRAND_ORIGINAL_DECLS")
	if directory == "" {
		t.Skip("set original declarations")
	}
	source := "import type { ArrowFunction } from " + strconv.Quote(filepath.ToSlash(filepath.Join(directory, "compiler/types.d.ts"))) + `;
interface Base { readonly kind: number; }
function helper(value: { readonly name?: string }): void {
 console.log(String(value.name));
}
helper({name: 'ordinary'});
const raw = {kind: 220, name: 'for' + 'ged'};
const base: Base = raw;
helper(base as ArrowFunction);
`
	path := filepath.Join(t.TempDir(), "never-helper.a")
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	if difference := disagreement(run{stdout: []byte("ordinary\nforged\n")}, onNode(t, path)); difference != "" {
		t.Fatal(difference)
	}
	program, err := lowered(t, path)
	if os.Getenv("ADAMIC_CALLABLE_NEVER_MUTANT") == "1" {
		if err != nil {
			t.Fatal(err)
		}
		truth := run{stdout: []byte("ordinary\nforged\n")}
		for _, got := range []run{releasedUncached(t, program), onJavaScriptBackend(t, program)} {
			if difference := disagreement(truth, got); difference != "" {
				t.Fatal(difference)
			}
		}
		t.Log("broad scalar exemption executed unsupported original never through wider helper")
		return
	}
	if err == nil || !strings.Contains(err.Error(), "field name with unsupported never contract") || !strings.Contains(err.Error(), "never-helper.a:4:") {
		t.Fatalf("wider helper lost original never refusal: %v", err)
	}
}
