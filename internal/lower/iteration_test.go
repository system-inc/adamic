package lower

import (
	"context"
	"errors"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIteratorGapsAreExplicit(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"interface Step {value:number;done?:boolean} const source={ [Symbol.iterator](){return{next():Step{return{value:1};}}}};for(const value of source){console.log(`${value}`);break;}",
		"interface Step {value:number;done:boolean} function closing():{return():Step}|undefined{return undefined;} const source={[Symbol.iterator](){return{...closing(),next():Step{return{value:1,done:false};}}}};for(const value of source){break;}",
		"interface Source {[Symbol.iterator]():{next():{value:number;done:boolean}}} function consume(source:Source):void {for(const value of source){break;}} consume({[Symbol.iterator](){return{next(){return{value:1,done:false};}}}});",
		"const source={[Symbol.iterator](){return{next:()=>({value:1,done:false}),return:()=>({value:0,done:true})};}};const iterator=source[Symbol.iterator]();iterator.next=():{value:number;done:boolean}=>({value:2,done:true});for(const value of source){break;}",
		"class Result {declare value:number;declare done:boolean;} const source={[Symbol.iterator](){return{next(){return new Result();}}}};for(const value of source){break;}",
		"const source={[Symbol.iterator](){return{next(){return{value:1,done:false};}}}};const copy={...source};for(const value of copy){break;}",
		"const source={[Symbol.iterator](){return{next(){return{value:1,done:false};}}}};const kept:(number|undefined)[]=[...source];",
	} {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, source)
			var gap *NotYet
			if !errors.As(err, &gap) {
				t.Fatalf("got %v, want an explicit iterator NotYet", err)
			}
		})
	}
}

func TestIteratorViewsCannotHideReturn(t *testing.T) {
	t.Parallel()
	source := `function plain(){return{[Symbol.iterator](){return{next(){return{value:1,done:false};}}}};}
 function closing(){return{[Symbol.iterator](){return{next(){return{value:1,done:false};},return(){return{value:0,done:true};}}}};}
 function consume(value:ReturnType<typeof plain>):void{for(const item of value){break;}}
 consume(closing());`
	_, err := lowerSource(t, source)
	var gap *NotYet
	if !errors.As(err, &gap) || !strings.Contains(gap.What, "hide a return") {
		t.Fatalf("got %v, want a hidden-return refusal", err)
	}
}

func TestIteratorViewsCannotEraseReceivers(t *testing.T) {
	t.Parallel()
	source := `const source={[Symbol.iterator](){return{next(){return{value:1,done:false};}}}};
 class Other{[Symbol.iterator](){return{next(){return{value:1,done:false};}}}}
 function consume(value:typeof source):void{for(const item of value){break;}}const other=new Other();consume(other);`
	_, err := lowerSource(t, source)
	var gap *NotYet
	if !errors.As(err, &gap) || !strings.Contains(gap.What, "receiver convention") {
		t.Fatalf("got %v, want an erased-receiver refusal", err)
	}
}

func TestIteratorDestructuringDoesNotLieAboutExhaustion(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, "const source={[Symbol.iterator](){return{next(){return{value:1,done:true};}}}};const [value]=source;console.log(`${value}`);")
	var refused *Refused
	if !errors.As(err, &refused) || !strings.Contains(refused.Fix, "default") {
		t.Fatalf("got %v, want a default or undefined fix", err)
	}
}

func TestGeneratorsAreRefusedEvenWithoutYield(t *testing.T) {
	t.Parallel()
	for _, source := range []string{"function* empty(){return 1;} empty();", "const empty=function*(){return 1;};empty();", "const source={*empty(){return 1;}};source.empty();"} {
		_, err := lowerSource(t, source)
		var refused *Refused
		if !errors.As(err, &refused) || !strings.Contains(refused.Fix, "suspended frames") {
			t.Errorf("got %v, want the generator ownership refusal", err)
		}
	}
}

func TestLiteralMethodCapturesCannotMakeCycles(t *testing.T) {
	t.Parallel()
	program, err := lowerSource(t, `function make():void{
 let holder:{read():number}|undefined;
 const value={read():number{return holder===undefined?0:1;}};
 holder=value;
 console.log('made');
 } make();`)
	if err != nil || len(program.GraphTypes) == 0 {
		t.Fatalf("want graph ownership for literal method captures, got %v", err)
	}

}

func TestLiteralMethodViewsDoNotLoseThis(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"const own={value:1,read():number{return this.value;}};const view:{readonly value:number;read:()=>number}=own;console.log(`${view.read()}`);",
		"const own={value:1,read():number{return this.value;}};const view:{readonly value:number;read():number}=own;console.log(`${view.read()}`);",
	} {
		_, err := lowerSource(t, source)
		var gap *NotYet
		if !errors.As(err, &gap) || !strings.Contains(gap.What, "erases its receiver") {
			t.Fatalf("got %v, want an explicit erased literal-method refusal", err)
		}
	}
}

func TestIteratorDescriptorReasons(t *testing.T) {
	t.Parallel()
	probes := []struct{ source, reason string }{
		{"interface Step{value:number;done?:boolean} const source={[Symbol.iterator](){return{next():Step{return{value:1};}}}};for(const value of source){break;}", "required boolean done"},
		{"interface Step{value?:number;done:boolean} const source={[Symbol.iterator](){return{next():Step{return{done:false};}}}};for(const value of source){break;}", "represented value field"},
		{"interface Step{value:number;done:boolean} function closing():{return():Step}|undefined{return undefined;} const source={[Symbol.iterator](){return{...closing(),next():Step{return{value:1,done:false};}}}};for(const value of source){break;}", "optional iterator method"},
		{"class Iterator{next(unused=1):{value:number;done:boolean}{return{value:1,done:false};}[Symbol.iterator]():Iterator{return this;}}for(const value of new Iterator()){break;}", "arguments or overloads"},
		{"class Result{declare value:number;declare done:boolean;}const source={[Symbol.iterator](){return{next(){return new Result();}}}};for(const value of source){break;}", "declare or abstract class field"},
		{"const source={[Symbol.iterator](){return{next:()=>({value:1,done:false})};}};const iterator=source[Symbol.iterator]();iterator.next=():{value:number;done:boolean}=>({value:2,done:true});for(const value of source){break;}", "replacing an iterator protocol"},
		{"const source={[Symbol.iterator](){return{next(){return{value:1,done:false};}}}};const copy={...source};for(const value of copy){break;}", "object spread"},
	}
	for _, probe := range probes {
		t.Run(probe.reason, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, probe.source)
			var gap *NotYet
			if !errors.As(err, &gap) || !strings.Contains(gap.What, probe.reason) {
				t.Fatalf("got %v, want iterator reason %s", err, probe.reason)
			}
		})
	}
}

func TestRepresentedMethodReplacementIsNotYet(t *testing.T) {
	t.Parallel()
	// Exercise the lowering safeguard directly; the up-front unbound-method pass also refuses this write.
	path := filepath.Join(t.TempDir(), "write.a")
	if err := os.WriteFile(path, []byte("class Box{next():number{return 1;}}const value=new Box();value.next=():number=>2;"), 0644); err != nil {
		t.Fatal(err)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	entry := program.Files()[0]
	checker, release := program.Checker(context.Background(), entry)
	defer release()
	lowering := &lowering{program: program, checker: checker, result: &ir.Program{}, this: -1, functionIndex: -1}
	if err := lowering.declareModule(entry.Statements.Nodes); err != nil {
		t.Fatal(err)
	}
	if _, err := lowering.statements(entry.Statements.Nodes[:2]); err != nil {
		t.Fatal(err)
	}
	assignment := entry.Statements.Nodes[2].AsExpressionStatement().Expression.AsBinaryExpression()
	_, err = lowering.setProperty(assignment.Left, assignment.Right)
	var gap *NotYet
	if !errors.As(err, &gap) || !strings.Contains(gap.What, "replacing a represented method") {
		t.Fatalf("got %v, want a method-replacement refusal", err)
	}
}

func TestDestructuredMethodsCannotLoadOwnSlots(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"const own={next():number{return 1;}};const {next}=own;console.log(`${next()}`);",
		"class Box{next():number{return 1;}}const own=new Box();const {next}=own;console.log(`${next()}`);",
		"class Box{next():number{return 1;}}const own=new Box();const view:{next:()=>number}=own;console.log(`${view.next()}`);",
		"const own={next():number{return 1;}};const view:{next:()=>number}=own;const {next}=view;console.log(`${next()}`);",
		"const own={};const {constructor}=own;",
	} {
		_, err := lowerSource(t, source)
		var refused *Refused
		var gap *NotYet
		if !errors.As(err, &refused) && !errors.As(err, &gap) {
			t.Fatalf("got %v, want a refusal before a method becomes an own-slot load", err)
		}
	}
}

func TestGenericIteratorViewsPreserveNativeArguments(t *testing.T) {
	t.Parallel()
	source := `class Sequence<T>{readonly data:readonly T[];constructor(data:readonly T[]){this.data=data;}[Symbol.iterator]():Sequence<T>{return this;}next():{value:number;done:boolean}{const ignored=this.data[0];return{value:1,done:false};}}
 const source:Sequence<number|undefined>=new Sequence([1]);for(const value of source){break;}`
	_, err := lowerSource(t, source)
	var refused *Refused
	if !errors.As(err, &refused) || !strings.Contains(refused.What, "nominal ancestry") {
		t.Fatalf("got %v, want the earlier nominal generic-view refusal", err)
	}
}

func TestIteratorMapperIndexHasNumberRepresentation(t *testing.T) {
	t.Parallel()
	source := `class Sequence{[Symbol.iterator]():Sequence{return this;}next():{value:number;done:boolean}{return{value:1,done:true};}}const source=new Sequence();const values=Array.from(source,(value:number,index:number|undefined):number=>value);`
	_, err := lowerSource(t, source)
	var gap *NotYet
	if !errors.As(err, &gap) || !strings.Contains(gap.What, "index representation") {
		t.Fatalf("got %v, want a mapper index representation refusal", err)
	}
}

func TestIteratorSymbolKeysAreNotStringKeys(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, `const source={value:1,[Symbol.iterator](){return{next(){return{value:1,done:true};}}}};console.log(Object.keys(source).join(','));`)
	var gap *NotYet
	if !errors.As(err, &gap) || !strings.Contains(gap.What, "symbol-key storage") {
		t.Fatalf("got %v, want explicit symbol-key storage refusal", err)
	}
}
