package lower

import (
	"os"
	"strings"
	"testing"
)

func TestOverloadCallbackServed(t *testing.T) {
	t.Parallel()
	source, err := os.ReadFile("../oracle/testdata/overload_callback_served.a")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := lowerSource(t, string(source)); err != nil {
		t.Fatal(err)
	}
}

func TestOverloadCallbackUnserved(t *testing.T) {
	t.Parallel()
	base := `interface Node {readonly kind:string;readonly text:string;}
interface BindingElement extends Node {readonly kind:'binding';}
interface OmittedExpression extends Node {readonly kind:'omitted';}
type ArrayBindingElement=BindingElement|OmittedExpression;
function visit<T extends Node>(node:T):T;
function visit(node:ArrayBindingElement):ArrayBindingElement{return {kind:'omitted',text:'changed'};}
`
	for index, source := range []string{
		`function run(visitor:(node:Node)=>Node):Node{return visitor({kind:'foreign',text:'x'});}console.log(run(visit).text);`,
		`function run(visitor:(node:BindingElement)=>BindingElement):BindingElement{return visitor({kind:'binding',text:'x'});}console.log(run(visit).kind);`,
		`function run(visitor:<T extends Node>(node:T)=>T):Node{return visitor({kind:'foreign',text:'x'});}console.log(run(visit).text);`,
		`const alias=visit;console.log(alias({kind:'binding',text:'x'}).text);`,
		`function run(visitor:(node:BindingElement)=>{}):{} {return visitor({kind:'binding',text:'x'});}console.log(typeof run(visit));`,
	} {
		name := []string{"input", "result", "generic", "alias", "representation"}[index]
		t.Run(name, func(t *testing.T) {
			_, err := lowerSource(t, base+source)
			if err == nil || !strings.Contains(err.Error(), "indirect value of an overload") {
				t.Fatalf("expected an unserved callback refusal, got %v", err)
			}
		})
	}
}

func TestOverloadCallbackFieldStorage(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, `interface Node {readonly kind:string;readonly metric:number|undefined;}
interface Binding extends Node {readonly kind:'binding';}
interface NumericBinding extends Binding {readonly metric:number;}
interface Omitted extends Node {readonly kind:'omitted';}
function visit<T extends Node>(node:T):T;
function visit(node:Binding|Omitted):Binding|Omitted{return node;}
function run(visitor:(node:NumericBinding)=>Node):Node{return visitor({kind:'binding',metric:1});}
console.log(run(visit).kind);`)
	if err == nil || !strings.Contains(err.Error(), "indirect value of an overload") {
		t.Fatalf("expected a field storage refusal, got %v", err)
	}
}
