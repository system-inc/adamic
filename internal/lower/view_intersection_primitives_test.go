package lower

import (
	"github.com/system-inc/adamic/internal/ir"
	"testing"
)

func TestPrimitiveIntersectionDescendantKeepsUnsupported(t *testing.T) {
	p := &ir.Program{ViewContracts: []ir.ViewContract{{Kind: ir.ViewUnion, Members: []ir.ViewContractID{2, 3}}, {Kind: ir.ViewScalar, Of: ir.String}, {Kind: ir.ViewScalar, Of: ir.Boolean}}}
	l := lowering{result: p}
	if !l.primitiveIntersectionDescendant(1) {
		t.Fatal("primitive descendant rejected")
	}
	for _, family := range []string{"never", "unknown", "untagged object union"} {
		p.ViewContracts[2].Unsupported = family
		if l.primitiveIntersectionDescendant(1) {
			t.Fatal("unsupported descendant was deferred:", family)
		}
	}
	p.ViewContracts[2] = ir.ViewContract{Kind: ir.ViewObject, Of: ir.Object}
	if l.primitiveIntersectionDescendant(1) {
		t.Fatal("object descendant obtained primitive exemption")
	}
}
