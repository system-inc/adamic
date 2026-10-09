package lower

import (
	"github.com/system-inc/adamic/internal/ir"
	"strings"
	"testing"
)

func TestOptionalObjectAliasContract(t *testing.T) {
	t.Parallel()
	program, err := lowerSource(t, `interface Base {readonly kind:number;readonly original?:Base|undefined}
 interface Node extends Base {readonly text:string;readonly original?:Node|undefined}
 function read(node:Base):string {const view=node as Node;const alias=view.original;return alias===undefined?"missing":alias.text}
 const child:Node={kind:2,text:"ok"};console.log(read({kind:1,original:child}));`)
	if err != nil {
		t.Fatal(err)
	}
	parent, child := false, false
	for _, function := range program.Functions {
		walk(function.Body, func(node any) bool {
			if p, ok := node.(ir.Property); ok {
				if p.Name == "original" {
					parent = p.Absent && p.View != "" && ir.OptionalObjectViewContract(program.ViewContracts, p.ViewContract)
				}
				if p.Name == "text" {
					child = p.View != "" && p.ViewContract > 0 && program.ViewContracts[p.ViewContract-1].Kind == ir.ViewScalar
				}
			}
			return true
		})
	}
	if !parent || !child {
		t.Fatalf("lost parent or child certificate: parent=%t child=%t", parent, child)
	}
}
func TestOptionalObjectUnprovenSourceRefused(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, `interface Base {readonly kind:number}
 interface Node extends Base {readonly original?:Node|undefined}
 function read(node:Base):Node|undefined {const view=node as Node;return view.original}
 console.log("unused");`)
	if err == nil || !strings.Contains(err.Error(), "optional field original has no proven compatible presence/type") || !strings.Contains(err.Error(), "keep compatible optional fields") {
		t.Fatalf("unproven source needs a path and fix: %v", err)
	}
}
func TestOptionalObjectGetterRemainsRefused(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, `interface Base {readonly kind:number;readonly original?:Base|undefined}
 interface Node extends Base {readonly text:string;readonly original?:Node|undefined}
 function read(node:Base):string {const view=node as Node;const alias=view.original;return alias===undefined?"missing":alias.text}
 const child:Node={kind:2,text:"ok"};const raw={kind:1,get original(){return child}};console.log(read(raw));`)
	if err == nil || !(strings.Contains(err.Error(), "getter in a checked field contract") || strings.Contains(err.Error(), "optional, accessor")) {
		t.Fatalf("getter must not use a data-slot certificate: %v", err)
	}
}
