package lower

import (
	"github.com/system-inc/adamic/internal/ir"
	"testing"
)

func TestPrimitiveDictionaryProjectionKeepsOriginalContract(t *testing.T) {
	l, target, release := mixedUnionLowering(t, `interface Ref {[key:string]:string[]} type Target=string|number|boolean|Ref|number[]|string[]|null|undefined;`)
	defer release()
	node := l.program.Files()[0].Statements.Nodes[0]
	original, err := l.viewContract(node, target)
	if err != nil {
		t.Fatal(err)
	}
	unsupported := l.result.ViewContracts[original-1].Unsupported
	if unsupported == "" {
		t.Fatal("control lost original unsupported reference alternatives")
	}
	selected, ok := l.dictionaryPrimitiveReadContract(node, target)
	if !ok || selected == original {
		t.Fatal("no distinct primitive read certificate")
	}
	if l.result.ViewContractTypes[int(target.Id())] != original || l.result.ViewContracts[original-1].Unsupported != unsupported {
		t.Fatal("original contract was erased or weakened")
	}
	if !ir.PrimitiveDictionaryReadCertificate(l.result, ir.Property{DictionaryPrimitive: true, DictionaryKey: ir.Undefined{}, ViewContract: selected}) {
		t.Fatal("selected primitive descriptor incomplete")
	}
}

func TestPrimitiveDictionaryProjectionCannotTrustUnknown(t *testing.T) {
	for _, source := range []string{
		`type Target=string|number|boolean|unknown;`,
		`type Target=string|number|boolean|any;`,
		`type Target<T>=string|number|boolean|T;`,
	} {
		t.Run(source, func(t *testing.T) {
			l, target, release := mixedUnionLowering(t, source)
			defer release()
			node := l.program.Files()[0].Statements.Nodes[0]
			if _, ok := l.dictionaryPrimitiveReadContract(node, target); ok {
				t.Fatal("unknown/generic arm authorized a primitive projection")
			}
		})
	}
}
