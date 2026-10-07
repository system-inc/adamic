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

func TestParserRepresentationProbes(t *testing.T) {
	t.Parallel()
	for _, gap := range []struct{ path, stdout, notYet, refused string }{
		{path: "gaps/1_recursive_state.ts", stdout: "35\n35\n", notYet: "a first-class nested function reference from another nested function"},
		{path: "gaps/2_state_arrow_cycle.ts", stdout: "35\n35\n", refused: "a cycle reference counting can't free"},
		{path: "gaps/3_recursive_callable.ts", stdout: "1\n0\n"},
		{path: "gaps/4_structural_ranges.ts", stdout: "1\n", refused: "a cycle reference counting can't free"},
		{path: "gaps/5_conditional_panic.ts", stdout: "T\n", notYet: "reading panic"},
		{path: "gaps/7_postfix_property.ts", stdout: "0\n1\n", notYet: "a PostfixUnaryExpression"},
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
