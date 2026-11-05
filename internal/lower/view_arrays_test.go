package lower

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
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

func TestViewReadonlyArrayLengthIsRaw(t *testing.T) {
	t.Parallel()
	_, err := lowerSourceAs(t, `interface Node {readonly kind:number}
interface Holder {readonly kind:'Holder'; readonly items:readonly Node[]}
const raw = {kind:'Holder' as const, items:[{kind:1}]};
const base:{readonly kind:'Holder'}=raw;
const holder = base as Holder;
console.log(String(holder.items.length));`, ".ts")
	if err != nil {
		t.Fatalf("length reads no element: %v", err)
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

// The source inventory grows with IR, while each admitted handout names a real
// oracle pin and must actually lower to a checked load or callback annotation.
func TestViewArrayAllowlist(t *testing.T) {
	t.Parallel()
	listed := map[string]bool{}
	for kind := range viewArrayRoads {
		listed[kind.Name()] = true
	}
	for name := range irOperandHolders(t) {
		if strings.HasPrefix(name, "Array") && !listed[name] {
			t.Errorf("ir.%s has no array-road classification", name)
		}
	}
	pins := map[string]string{"index": "index-good", "at": "at-lie", "for-of": "for-of-good", "values": "values-lie", "forEach": "forEach-good", "map": "map-good", "filter": "filter-good", "some": "some-good", "every": "every-good", "find": "find-good", "findIndex": "findIndex-good", "findLast": "findLast-good", "findLastIndex": "findLastIndex-good", "reduce": "reduce-good"}
	for road, pin := range pins {
		t.Run(road, func(t *testing.T) {
			t.Parallel()
			source, err := os.ReadFile(filepath.Join("../oracle/testdata/views_v3", pin+".ts"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowerSourceAs(t, string(source), ".ts")
			if err != nil {
				t.Fatal(err)
			}
			found := false
			inspect := func(node any) bool {
				switch value := node.(type) {
				case ir.ViewElement:
					found = found || ((road == "index" || road == "at" || road == "values") && (road != "at" || value.Relative))
				case ir.ForOf:
					found = found || (road == "for-of" && value.ElementCheck != nil)
				case ir.ArrayMap:
					found = found || (road == "map" && value.ElementCheck != nil)
				case ir.ArrayReduce:
					found = found || (road == "reduce" && value.ElementCheck != nil)
				case ir.ArrayVisit:
					found = found || (road == value.Method && value.ElementCheck != nil)
				}
				return true
			}
			walk(program.Main, inspect)
			for _, fn := range program.Functions {
				walk(fn.Body, inspect)
			}
			if !found {
				t.Fatalf("%s does not exercise the checked %s road", pin, road)
			}
		})
	}
	for kind, road := range viewArrayRoads {
		if road.handout || road.noElement {
			continue
		}
		t.Run(kind.Name(), func(t *testing.T) {
			source := ir.Read{Of: ir.Array}
			value := reflect.New(kind).Elem()
			field := value.FieldByName("Array")
			if kind == reflect.TypeFor[ir.ArrayFrom]() {
				if viewReadOf(nil, value.Interface().(ir.Expression)) != viewReadNotYet {
					t.Fatal("Array.from must fail closed")
				}
				return
			}
			if !field.IsValid() {
				t.Fatalf("denied road has no array operand: %v", kind)
			}
			field.Set(reflect.ValueOf(source))
			program := &ir.Program{ViewOrigins: []ir.Expression{source}, ViewContracts: []ir.ViewContract{{Kind: ir.ViewArray, ArrayReadonly: true}}, Main: []ir.Statement{ir.Evaluate{Value: value.Interface().(ir.Expression)}}}
			l := &lowering{result: program}
			var stop *NotYet
			if err := l.checkViewArrays(); !errors.As(err, &stop) || !strings.Contains(stop.What, road.member) {
				t.Fatalf("unchecked %s road did not fail closed: %v", road.member, err)
			}
		})
	}
}
