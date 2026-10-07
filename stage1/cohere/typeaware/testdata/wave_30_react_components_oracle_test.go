// Overlay-only test in the unmodified Go production rule package. The oracle
// calls its private production helpers, rather than copying their decisions.
package react

import (
	"fmt"
	hir "github.com/system-inc/cohere/internal/lint/ecmascript/high_level_intermediate_representation"
	"os"
	"strings"
	"testing"
)

func wave30RefBase(kind refsAccessKind, variant int) *refsAccessType {
	return &refsAccessType{Kind: kind, RefId: variant % 2, HasRefId: variant >= 2, Span: hir.IdentifierId(variant % 2), HasSpan: variant >= 2, RefSpan: hir.IdentifierId(9 + variant), HasRefSpan: variant%2 == 1}
}
func wave30RefFixture(id int) *refsAccessType {
	if id == 0 {
		return nil
	}
	if id <= 24 {
		return wave30RefBase(refsAccessKind((id-1)/4), (id-1)%4)
	}
	r := &refsAccessType{Kind: refsStructure}
	switch {
	case id <= 28:
		r.Value = wave30RefBase(refsRef, id-25)
	case id <= 32:
		r.Value = wave30RefBase(refsRefValue, id-29)
	case id <= 36:
		v := id - 33
		k := refsRef
		if v%2 != 0 {
			k = refsRefValue
		}
		r.Function = &refsFunctionType{ReadRefEffect: v%2 == 1, RefAccessSpan: hir.IdentifierId(20 + v), HasRefAccessSpan: v >= 2, ReturnType: wave30RefBase(k, v)}
	case id == 37:
		r.Value = wave30RefFixture(25)
	case id == 38 || id == 39:
		r.Function = &refsFunctionType{ReadRefEffect: id == 38, RefAccessSpan: 77, HasRefAccessSpan: true}
		if id == 39 {
			r.Function.ReturnType = &refsAccessType{Kind: refsNone}
		}
	default:
		panic("invalid fixture")
	}
	return r
}
func wave30RefWire(v *refsAccessType) string {
	if v == nil {
		return "nil"
	}
	fn := "nil"
	if v.Function != nil {
		f := v.Function
		fn = fmt.Sprintf("%t,%d,%t(%s)", f.ReadRefEffect, f.RefAccessSpan, f.HasRefAccessSpan, wave30RefWire(f.ReturnType))
	}
	return fmt.Sprintf("%d,%d,%t,%d,%t,%d,%t{%s}{%s}", v.Kind, v.RefId, v.HasRefId, v.Span, v.HasSpan, v.RefSpan, v.HasRefSpan, wave30RefWire(v.Value), fn)
}
func TestWave30ReactComponentOracle(t *testing.T) {
	directory := os.Getenv("ADAMIC_WAVE30_COMPONENT_ORACLE")
	if directory == "" {
		t.Fatal("missing output directory")
	}
	var refs, purity strings.Builder
	for a := 0; a < 40; a++ {
		for b := 0; b < 40; b++ {
			left, right := wave30RefFixture(a), wave30RefFixture(b)
			next := 1000
			mint := func() int { next++; return next }
			joined := refsJoin(left, right, mint)
			folded := refsJoinMany([]*refsAccessType{left, right, wave30RefFixture(29)}, mint)
			fmt.Fprintf(&refs, "%d,%d\t%t\t%s\t%s\t%s\n", a, b, refsTypeEqual(left, right), wave30RefWire(joined), wave30RefWire(folded), wave30RefWire(refsDestructure(left)))
		}
	}
	for kind := -1; kind < 4; kind++ {
		for _, name := range []string{"", "Math", "Date", "performance", "window", "globalThis", "other"} {
			for _, property := range []string{"", "Math", "Date", "performance", "random", "now", "floor"} {
				values := map[hir.IdentifierId]purityValue{2: {Kind: purityValueImpure, CanonicalName: "preserved"}}
				if kind >= 0 {
					values[1] = purityValue{Kind: purityValueKind(kind), CanonicalName: name}
				}
				purityLoadProperty(nil, &hir.Instruction{LValue: hir.Place{Identifier: 2}}, values, hir.Place{Identifier: 1}, property)
				result := values[2]
				fmt.Fprintf(&purity, "%d,%s,%s\t%d,%s\n", kind, name, property, result.Kind, result.CanonicalName)
			}
		}
	}
	for name, data := range map[string]string{"refs": refs.String(), "purity": purity.String()} {
		if err := os.WriteFile(directory+"/"+name+"-go.stdout", []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
	}
}
