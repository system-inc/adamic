// Loaded in the production React package through an owned Go overlay.
package react

import (
	"encoding/json"
	"fmt"
	hir "github.com/system-inc/cohere/internal/lint/ecmascript/high_level_intermediate_representation"
	"os"
	"strings"
	"testing"
)

type wave04Value struct {
	Kind, RefId, Span, Value, Function int
	HasRefId, HasSpan                  bool
}
type wave04Function struct {
	Effect bool
	Return int
}

func TestWave04RefsKernel(t *testing.T) {
	// Not parallel: output path is supplied for this isolated oracle invocation.
	rows := []wave04Value{}
	for kind := 0; kind < 6; kind++ {
		rows = append(rows, wave04Value{Kind: kind, Value: -1, Function: -1})
	}
	rows = append(rows,
		wave04Value{Kind: 2, RefId: 1, HasRefId: true, Value: -1, Function: -1},
		wave04Value{Kind: 2, RefId: 2, HasRefId: true, Value: -1, Function: -1},
		wave04Value{Kind: 3, RefId: 1, HasRefId: true, Value: -1, Function: -1},
		wave04Value{Kind: 3, RefId: 2, HasRefId: true, Value: -1, Function: -1},
		wave04Value{Kind: 4, RefId: 1, HasRefId: true, Span: 3, HasSpan: true, Value: -1, Function: -1},
		wave04Value{Kind: 4, RefId: 2, HasRefId: true, Span: 3, HasSpan: true, Value: -1, Function: -1},
		wave04Value{Kind: 4, Span: 4, HasSpan: true, Value: -1, Function: -1},
		wave04Value{Kind: 4, Span: 3, Value: -1, Function: -1},
		wave04Value{Kind: 5, Value: 8, Function: -1},
		wave04Value{Kind: 5, Value: 9, Function: -1},
		wave04Value{Kind: 5, Value: -1, Function: 0},
		wave04Value{Kind: 5, Value: -1, Function: 1},
		wave04Value{Kind: 5, Value: -1, Function: 2},
		wave04Value{Kind: 5, Value: -1, Function: 3},
		wave04Value{Kind: 5, Value: 14, Function: 0},
		wave04Value{Kind: 5, Value: 15, Function: 1},
		wave04Value{Kind: 2, RefId: 0, HasRefId: true, Value: -1, Function: -1},
	)
	frows := []wave04Function{{false, 8}, {false, 9}, {true, 9}, {false, 12}}
	values := make([]*refsAccessType, len(rows))
	funcs := make([]*refsFunctionType, len(frows))
	for i, r := range rows {
		values[i] = &refsAccessType{Kind: refsAccessKind(r.Kind), RefId: r.RefId, HasRefId: r.HasRefId, Span: hir.IdentifierId(r.Span), HasSpan: r.HasSpan}
	}
	for i, r := range frows {
		funcs[i] = &refsFunctionType{ReadRefEffect: r.Effect, ReturnType: values[r.Return]}
	}
	for i, r := range rows {
		if r.Value >= 0 {
			values[i].Value = values[r.Value]
		}
		if r.Function >= 0 {
			values[i].Function = funcs[r.Function]
		}
	}
	get := func(i int) *refsAccessType {
		if i < 0 {
			return nil
		}
		return values[i]
	}
	var out strings.Builder
	for i := -1; i < len(values); i++ {
		for j := -1; j < len(values); j++ {
			fmt.Fprintf(&out, "equal %d %d %t\n", i, j, refsTypeEqual(get(i), get(j)))
		}
	}
	names := []string{"", "use", "useA", "use9", "usea", "useÉ", "useRef", "ref", "Ref", "aRef", "1Ref", "$Ref", "_Ref", "a-Ref", "éRef", "a😀Ref"}
	for i, n := range names {
		fmt.Fprintf(&out, "name %d %t %t\n", i, refsIsHookName(n), refsIsRefLikeName(n))
	}
	data, err := json.Marshal(struct {
		Values    []wave04Value
		Functions []wave04Function
		Names     []string
		Expected  string
	}{rows, frows, names, out.String()})
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(os.Getenv("WAVE04_REFS_KERNEL"), data, 0600); err != nil {
		t.Fatal(err)
	}
}
