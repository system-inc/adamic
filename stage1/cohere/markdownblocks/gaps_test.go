package markdownblocks

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

// Not parallel: shared markdownMemory configuration and artifacts build cache; existing helper controls parallel execution.
func TestParserRepresentationProbes(t *testing.T) {
	parallelMarkdown(t)
	for _, gap := range []struct{ path, stdout, notYet, refused, refusedExact string }{
		{path: "gaps/1_recursive_state.ts", stdout: "35\n35\n"},
		{path: "gaps/2_state_arrow_cycle.ts", stdout: "35\n35\n", refused: "a cycle reference counting can't free"},
		{path: "gaps/3_recursive_callable.ts", stdout: "1\n0\n"},
		{path: "gaps/14_missing_path_key.ts", stdout: "key=\"\" present=false\n"},
		{path: "gaps/4_structural_ranges.ts", stdout: "1\n", refused: "a cycle reference counting can't free"},
		{path: "gaps/5_conditional_panic.ts", stdout: "T\n"},
		{path: "gaps/7_postfix_property.ts", stdout: "0\n1\n", notYet: "a PostfixUnaryExpression"},
		{path: "gaps/8_generic_callback_result.ts", stdout: "value\n", notYet: "a function returning Result"},
		{path: "gaps/9_mixed_path_names.ts", stdout: "children,0\n", notYet: "join on an array of objects, arrays, maps or functions"},
		{path: "gaps/10_multiple_push.ts", stdout: "1,2\n", notYet: "push with other than one value"},
		{path: "gaps/11_function_expression.ts", stdout: "value\n"},
		{path: "gaps/12_conditional_empty_array.ts", stdout: "0\n"},
		{path: "gaps/15_string_or.ts", stdout: "fallback\nx\n"},
		{path: "gaps/16_array_shift.ts", stdout: "1\n2\n", refusedExact: ":2:16: Adamic 0.1 refuses inherited library member shift read as an own field; prototype members are not stored in an object's shape; call the method on its receiver, or wrap that call in an arrow (unbound-method)"},
		{path: "gaps/17_long_optional_chain.ts", stdout: "1\n"},
	} {
		t.Run(gap.path, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(gap.path)
			if err != nil {
				t.Fatal(err)
			}
			source := onNode(t, path)
			clean(t, "Node probe", source)
			equal(t, "Node probe", source.stdout, []byte(gap.stdout))
			program, err := load.Load([]string{path})
			if err != nil {
				t.Fatal(err)
			}
			lowered, err := lower.Lower(context.Background(), program)
			if gap.notYet != "" {
				var notYet *lower.NotYet
				if !errors.As(err, &notYet) || notYet.What != gap.notYet {
					t.Fatalf("NotYet gap changed: %v; update GAPS.md", err)
				}
				return
			}
			if gap.refusedExact != "" {
				var refused *lower.Refused
				if !errors.As(err, &refused) || err.Error() != path+gap.refusedExact {
					t.Fatalf("Exact refusal changed: %v; update GAPS.md", err)
				}
				return
			}
			if gap.refused != "" {
				var refused *lower.Refused
				if !errors.As(err, &refused) || !strings.Contains(refused.What, gap.refused) {
					t.Fatalf("Refused shape changed: %v; update GAPS.md", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			native, binary := natively(t, lowered)
			for _, side := range []run{native, onJavaScriptBackend(t, lowered)} {
				clean(t, "closed callable probe", side)
				equal(t, "closed callable probe", side.stdout, source.stdout)
			}
			if report := leaks(t, lowered, binary); report != "" {
				t.Fatal(report)
			}
		})
	}
}
