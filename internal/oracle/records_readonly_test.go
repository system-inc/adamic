package oracle

import (
	"errors"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// #6c4swvt piece 4: a readonly string-index type the program first puts a record into is held as that record
// (lower's records_readonly.go), so tsc's read-only option workers can take Readonly<OptionsBase>. The honest
// reads agree with Node on both backends with nothing leaked, the cast back to the writable record stays refused,
// and a parameter handed a record and then an enumeration's object stops at the second call by name (Runtime's
// must). B2's road, an enumeration's object put first into such a type, keeps its storage: enum_views' tests hold
// it. A readonly record seen as a writable one is a widening, since it may hold a narrower record
// (readonly_writable.ts), and so is the way back through a type parameter and through a field, Language's q02
// and q06 (readonly_back_generic.ts, readonly_back_field.ts). Mutants: the classification never answering a record, and readonly_honest.ts stops at
// its view again; the readonly covariance dropped from widened, and it's refused as a widening; the way back
// admitted, and readonly_writable.ts lowers.
func TestReadonlyRecordViews(t *testing.T) {
	t.Parallel()
	for name, view := range map[string]struct{ node, notYet, refused, at string }{
		"readonly_honest.ts":       {node: "watchFile=number,excludeFiles=object,synchronous=boolean:true\nwatchFile=number,excludeFiles=object,synchronous=boolean:false\n"},
		"readonly_leak.ts":         {node: "4\n", refused: "(adamic/no-unchecked-cast)"},
		"readonly_mixed.ts":        {node: "3\n4\n", notYet: "is held as a record, since ", at: ":18:"},
		"readonly_writable.ts":     {node: "object\n", refused: "seen as OptionsBase, which can write OptionsBase where Readonly<OptionsBase> is read"},
		"readonly_back_generic.ts": {node: "object\n", refused: "Readonly<T> seen as T"},
		"readonly_back_field.ts":   {node: "object\n", refused: "(adamic/invariant-mutable)"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path := optionsBasePath(t, name)
			node := onNode(t, path)
			if node.exitCode != 0 || string(node.stdout) != view.node {
				t.Fatalf("Node: %s", optionsBaseShown(node))
			}
			program, err := lowered(t, path)
			switch {
			case view.notYet != "":
				var notYet *lower.NotYet
				if !errors.As(err, &notYet) || !strings.Contains(notYet.What, view.notYet) || !strings.HasPrefix(notYet.Where, path+view.at) {
					t.Fatalf("want NotYet naming %q at %s, got %v", view.notYet, view.at, err)
				}
				return
			case view.refused != "":
				var refused *lower.Refused
				if !errors.As(err, &refused) || !strings.Contains(err.Error(), view.refused) {
					t.Fatalf("want refused %q, got %v", view.refused, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("must lower: %v", err)
			}
			native, sanitized := natively(t, program)
			for backend, got := range map[string]run{"native": native, "javascript": onJavaScriptBackend(t, program)} {
				if difference := disagreement(node, got); difference != "" {
					t.Fatalf("%s disagrees with Node (%s): Node %s; %s %s", backend, difference, optionsBaseShown(node), backend, optionsBaseShown(got))
				}
			}
			if report := leaks(t, program, sanitized); report != "" {
				t.Fatal(report)
			}
		})
	}
}

// Runtime's sharing seam on piece 4: a readonly view of a record is an alias of it, never immutable data, so two
// parallel tasks, one reading through the view and one writing the record, are refused at compile time.
func TestReadonlyRecordViewShared(t *testing.T) {
	t.Parallel()
	path := optionsBasePath(t, "readonly_share.a")
	_, err := lowered(t, path)
	var refused *lower.Refused
	if !errors.As(err, &refused) {
		t.Fatalf("want the task's write to the captured record refused, got %v", err)
	}
	t.Logf("refused: %v", err)
}
