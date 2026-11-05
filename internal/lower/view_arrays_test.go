package lower

import (
	"errors"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func TestViewArrayReadonlyDescriptor(t *testing.T) {
	t.Parallel()
	for _, spelling := range []string{"readonly Node[]", "ReadonlyArray<Node>", "Node[]"} {
		t.Run(spelling, func(t *testing.T) {
			t.Parallel()
			l, target, release := mixedUnionLowering(t, "interface Node {readonly kind:number} type Target="+spelling+";")
			defer release()
			node := l.program.Files()[0].Statements.Nodes[0]
			id, err := l.viewContract(node, target)
			if err != nil {
				t.Fatal(err)
			}
			contract := l.result.ViewContracts[id-1]
			wantReadonly := spelling != "Node[]"
			if contract.Kind != ir.ViewArray || contract.Element == 0 || contract.ArrayReadonly != wantReadonly {
				t.Fatalf("lost array contract: %#v", contract)
			}
		})
	}
}

func TestViewMutableArrayFieldNotYet(t *testing.T) {
	t.Parallel()
	for _, field := range []string{"items: Node[]", "readonly items: Node[]", "readonly items: Node[] | undefined"} {
		t.Run(field, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, `interface Node {readonly kind:number}
interface Holder {readonly kind:'Holder'; `+field+`}
const raw = {kind:'Holder' as const, items:[{kind:1}]};
const base:{readonly kind:'Holder'}=raw;
const holder = base as Holder;
console.log(String(holder.items !== undefined));`, ".ts")
			var notYet *NotYet
			if !errors.As(err, &notYet) || notYet.What != "a checked view of mutable array field items" {
				t.Fatalf("want mutable-array NotYet at read, got %v", err)
			}
		})
	}
}

func TestViewReadonlyArrayStillFailsClosed(t *testing.T) {
	t.Parallel()
	_, err := lowerSourceAs(t, `interface Node {readonly kind:number}
interface Holder {readonly kind:'Holder'; readonly items:readonly Node[]}
const raw = {kind:'Holder' as const, items:[{kind:1}]};
const base:{readonly kind:'Holder'}=raw;
const holder = base as Holder;
console.log(String(holder.items.length));`, ".ts")
	if err == nil || !strings.Contains(err.Error(), "awaits views-v3: array element kind") {
		t.Fatalf("readonly elements admitted before checks exist: %v", err)
	}
}

func TestViewMutableArrayFieldUndemanded(t *testing.T) {
	t.Parallel()
	_, err := lowerSourceAs(t, `interface Node {readonly kind:number}
interface Holder {readonly kind:'Holder'; readonly items:Node[]}
const raw = {kind:'Holder' as const, items:[{kind:1}]};
const base:{readonly kind:'Holder'}=raw;
const holder = base as Holder;
console.log(holder.kind);`, ".ts")
	if err != nil {
		t.Fatalf("unused mutable field refused at cast: %v", err)
	}
}
