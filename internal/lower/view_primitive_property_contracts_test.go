package lower

import (
	"github.com/system-inc/adamic/internal/ir"
	"testing"
)

func TestMixedPrimitivePropertyCertificateRejectsUnsupported(t *testing.T) {
	base := ir.Program{ViewContracts: []ir.ViewContract{
		{Kind: ir.ViewUnion, Of: ir.Union, Members: []ir.ViewContractID{2, 3}},
		{Kind: ir.ViewScalar, Of: ir.String},
		{Kind: ir.ViewObject, Of: ir.Object},
	}}
	if !mixedPrimitivePropertyChecks(&base, 1) {
		t.Fatal("complete read selector rejected")
	}
	for _, edit := range []func(*ir.Program){
		func(p *ir.Program) { p.ViewContracts[0].Unsupported = "unknown" },
		func(p *ir.Program) { p.ViewContracts[2].Unsupported = "never" },
		func(p *ir.Program) { p.ViewContracts[2].Kind = ir.ViewUnknown },
		func(p *ir.Program) { p.ViewContracts[2].Nominal = "opaque class" },
		func(p *ir.Program) { p.ViewContracts[0].Members = []ir.ViewContractID{2, 9} },
		func(p *ir.Program) { p.ViewContracts[0].Members = []ir.ViewContractID{2} },
	} {
		p := base
		p.ViewContracts = append([]ir.ViewContract(nil), base.ViewContracts...)
		edit(&p)
		if mixedPrimitivePropertyChecks(&p, 1) {
			t.Fatal("unsupported descriptor supplied a read certificate")
		}
	}
}
