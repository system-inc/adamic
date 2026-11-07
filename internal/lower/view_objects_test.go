package lower

import (
	"github.com/system-inc/adamic/internal/ir"
	"strings"
	"testing"
)

func TestViewObjectContractsAreAvailableToEraser(test *testing.T) {
	test.Parallel()
	// .ts, where the cast stays a checked view: in .a a downcast whose checks no proof erases is refused (#m8dvvmf).
	program, lowerSourceAsError := lowerSourceAs(test, `interface Base {readonly kind:'box'|'other'}
 interface Box extends Base {readonly kind:'box';readonly child:{readonly ready:boolean}}
 function visit(node:Base):boolean {return (node as Box).child.ready}
 const raw={kind:'box' as const,child:{ready:true}}; console.log(String(visit(raw)));`, ".ts")
	if lowerSourceAsError != nil {
		test.Fatal(lowerSourceAsError)
	}
	found := false
	for _, function := range program.Functions {
		walk(function.Body, func(node any) bool {
			if property, ok := node.(ir.Property); ok && property.Name == "child" && property.View != "" {
				if property.ViewContract == 0 {
					test.Fatal("object read has no semantic contract for the eraser")
				}
				contract := program.ViewContracts[int(property.ViewContract)-1]
				if contract.Kind != ir.ViewObject || len(contract.Fields) != 1 || contract.Fields[0].Name != "ready" {
					test.Fatalf("incomplete child contract: %#v", contract)
				}
				field := program.ViewContracts[int(contract.Fields[0].Contract)-1]
				if field.Kind != ir.ViewScalar || field.Of != ir.Boolean {
					test.Fatalf("missing scalar child contract: %#v", field)
				}
				found = true
			}
			return true
		})
	}
	if !found {
		test.Fatal("no transitive object read found")
	}
}

func TestViewObjectWritesNeedSourceCertificate(test *testing.T) {
	test.Parallel()
	for _, assignment := range []string{
		`view.child={ready:true};`,
		`({child:view.child}={child:{ready:true}});`,
	} {
		_, lowerSourceError := lowerSource(test, `interface Base {readonly kind:'box'|'other'}
  interface Box extends Base {readonly kind:'box';child:{ready:boolean}}
  const raw={kind:'box' as const,child:{ready:true,other:42}};
  const held:Base=raw;const view=held as Box;`+assignment+`console.log(String(raw.child.other));`)
		if lowerSourceError == nil || !strings.Contains(lowerSourceError.Error(), "source-slot type certificate") {
			test.Fatalf("unchecked write could invalidate another alias: %v", lowerSourceError)
		}
	}
}
