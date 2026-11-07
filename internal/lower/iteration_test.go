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

func TestIteratorGapsAreExplicit(test *testing.T) {
	test.Parallel()
	for _, source := range []string{
		"interface Step {value:number;done?:boolean} const source={ [Symbol.iterator](){return{next():Step{return{value:1};}}}};for(const value of source){console.log(`${value}`);break;}",
		"interface Step {value:number;done:boolean} function closing():{return():Step}|undefined{return undefined;} const source={[Symbol.iterator](){return{...closing(),next():Step{return{value:1,done:false};}}}};for(const value of source){break;}",
		"interface Source {[Symbol.iterator]():{next():{value:number;done:boolean}}} function consume(source:Source):void {for(const value of source){break;}} consume({[Symbol.iterator](){return{next(){return{value:1,done:false};}}}});",
		"const source={[Symbol.iterator](){return{next:()=>({value:1,done:false}),return:()=>({value:0,done:true})};}};const iterator=source[Symbol.iterator]();iterator.next=():{value:number;done:boolean}=>({value:2,done:true});for(const value of source){break;}",
		"class Result {declare value:number;declare done:boolean;} const source={[Symbol.iterator](){return{next(){return new Result();}}}};for(const value of source){break;}",
		"const source={[Symbol.iterator](){return{next(){return{value:1,done:false};}}}};const copy={...source};for(const value of copy){break;}",
		"const source={[Symbol.iterator](){return{next(){return{value:1,done:false};}}}};const kept:(number|undefined)[]=[...source];",
	} {
		test.Run(source, func(test *testing.T) {
			test.Parallel()
			_, err := lowerSource(test, source)
			var gap *NotYet
			if !errors.As(err, &gap) {
				test.Fatalf("got %v, want an explicit iterator NotYet", err)
			}
		})
	}
}

func TestIteratorViewsCannotHideReturn(test *testing.T) {
	test.Parallel()
	source := `function plain(){return{[Symbol.iterator](){return{next(){return{value:1,done:false};}}}};}
 function closing(){return{[Symbol.iterator](){return{next(){return{value:1,done:false};},return(){return{value:0,done:true};}}}};}
 function consume(value:ReturnType<typeof plain>):void{for(const item of value){break;}}
 consume(closing());`
	_, err := lowerSource(test, source)
	var gap *NotYet
	if !errors.As(err, &gap) || !strings.Contains(gap.What, "hide a return") {
		test.Fatalf("got %v, want a hidden-return refusal", err)
	}
}

func TestIteratorViewsCannotEraseReceivers(test *testing.T) {
	test.Parallel()
	source := `const source={[Symbol.iterator](){return{next(){return{value:1,done:false};}}}};
 class Other{[Symbol.iterator](){return{next(){return{value:1,done:false};}}}}
 function consume(value:typeof source):void{for(const item of value){break;}}const other=new Other();consume(other);`
	_, err := lowerSource(test, source)
	var gap *NotYet
	if !errors.As(err, &gap) || !strings.Contains(gap.What, "receiver convention") {
		test.Fatalf("got %v, want an erased-receiver refusal", err)
	}
}

func TestIteratorDestructuringDoesNotLieAboutExhaustion(test *testing.T) {
	test.Parallel()
	_, err := lowerSource(
		test,
		"const source={[Symbol.iterator](){return{next(){return{value:1,done:true};}}}};const [value]=source;console.log(`${value}`);",
	)
	var refused *Refused
	if !errors.As(err, &refused) || !strings.Contains(refused.Fix, "default") {
		test.Fatalf("got %v, want a default or undefined fix", err)
	}
}

// Generators are ruled NotYet, never Refused (docs/rulings.md, open question 17): the stop says
// support is intended. A yield stops at its function, and so does a generator with none. A top-level
// declaration consumed by a for-of lowers since #qk8rztp's S1c (generators_test.go), and one nested
// in an ordinary function's body since S3a (generators_nested_test.go); every form here is one that
// still doesn't lower, a nested one inside an async function among them.
func TestGeneratorsAreNotYetEvenWithoutYield(test *testing.T) {
	test.Parallel()
	for _, source := range []string{
		"function* empty(){return 1;} empty();",
		"const empty=function*(){return 1;};empty();",
		"const source={*empty(){return 1;}};source.empty();",
		"async function outer(){function* numbers(){yield 1;} for (const value of numbers()) console.log(String(value));} await outer();",
		"class Numbers{*each(){yield 1;}} for (const value of new Numbers().each()) console.log(String(value));",
	} {
		_, err := lowerSource(test, source)
		var gap *NotYet
		if !errors.As(err, &gap) || !strings.Contains(gap.What, "a generator function") {
			test.Errorf("%s: got %v, want the generator NotYet", source, err)
		}
	}
}

func TestLiteralMethodCapturesCannotMakeCycles(test *testing.T) {
	test.Parallel()
	program := graphAccepted(test, `function make():void{
 let holder:{read():number}|undefined;
 const value={read():number{return holder===undefined?0:1;}};
 holder=value;
 console.log('made');
 } make();`)
	if len(program.GraphTypes) == 0 {
		test.Fatal("missing graph ownership")
	}
}

func TestLiteralMethodViewsDoNotLoseThis(test *testing.T) {
	test.Parallel()
	lowersAndAgreesWithNode(
		test,
		"const own={value:1,read():number{return this.value;}};const view:{readonly value:number;read:()=>number}=own;console.log(`${view.read()}`);",
	)
}

func TestLiteralMethodSignatureViewsDoNotLoseThis(test *testing.T) {
	test.Parallel()
	lowersAndAgreesWithNode(
		test,
		"const own={value:1,read():number{return this.value;}};const view:{readonly value:number;read():number}=own;console.log(`${view.read()}`);",
	)
}

func TestIteratorDescriptorReasons(test *testing.T) {
	test.Parallel()
	probes := []struct{ source, reason string }{
		{
			"interface Step{value:number;done?:boolean} const source={[Symbol.iterator](){return{next():Step{return{value:1};}}}};for(const value of source){break;}",
			"required boolean done",
		},
		{
			"interface Step{value?:number;done:boolean} const source={[Symbol.iterator](){return{next():Step{return{done:false};}}}};for(const value of source){break;}",
			"represented value field",
		},
		{
			"interface Step{value:number;done:boolean} function closing():{return():Step}|undefined{return undefined;} const source={[Symbol.iterator](){return{...closing(),next():Step{return{value:1,done:false};}}}};for(const value of source){break;}",
			"optional iterator method",
		},
		{
			"class Iterator{next(unused=1):{value:number;done:boolean}{return{value:1,done:false};}[Symbol.iterator]():Iterator{return this;}}for(const value of new Iterator()){break;}",
			"arguments or overloads",
		},
		{
			"class Result{declare value:number;declare done:boolean;}const source={[Symbol.iterator](){return{next(){return new Result();}}}};for(const value of source){break;}",
			"declare or abstract class field",
		},
		{
			"const source={[Symbol.iterator](){return{next:()=>({value:1,done:false})};}};const iterator=source[Symbol.iterator]();iterator.next=():{value:number;done:boolean}=>({value:2,done:true});for(const value of source){break;}",
			"replacing an iterator protocol",
		},
		{
			"const source={[Symbol.iterator](){return{next(){return{value:1,done:false};}}}};const copy={...source};for(const value of copy){break;}",
			"object spread",
		},
	}
	for _, probe := range probes {
		test.Run(probe.reason, func(test *testing.T) {
			test.Parallel()
			_, err := lowerSource(test, probe.source)
			var gap *NotYet
			if !errors.As(err, &gap) || !strings.Contains(gap.What, probe.reason) {
				test.Fatalf("got %v, want iterator reason %s", err, probe.reason)
			}
		})
	}
}

func TestRepresentedMethodReplacementIsNotYet(test *testing.T) {
	test.Parallel()
	// Exercise the lowering safeguard directly; the up-front unbound-method pass also refuses this write.
	path := filepath.Join(test.TempDir(), "write.a")
	if err := os.WriteFile(
		path,
		[]byte("class Box{next():number{return 1;}}const value=new Box();value.next=():number=>2;"),
		0644,
	); err != nil {
		test.Fatal(err)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		test.Fatal(err)
	}
	entry := program.Files()[0]
	checker, release := program.Checker(context.Background(), entry)
	defer release()
	lowering := &lowering{program: program, checker: checker, result: &ir.Program{}, this: -1, functionIndex: -1}
	if err := lowering.declareModule(entry.Statements.Nodes); err != nil {
		test.Fatal(err)
	}
	if _, err := lowering.statements(entry.Statements.Nodes[:2]); err != nil {
		test.Fatal(err)
	}
	assignment := entry.Statements.Nodes[2].AsExpressionStatement().Expression.AsBinaryExpression()
	_, err = lowering.setProperty(assignment.Left, assignment.Right)
	var gap *NotYet
	if !errors.As(err, &gap) || !strings.Contains(gap.What, "replacing a represented method") {
		test.Fatalf("got %v, want a method-replacement refusal", err)
	}
}

func TestDestructuredMethodsCannotLoadOwnSlots(test *testing.T) {
	test.Parallel()
	for _, source := range []string{
		"const own={next():number{return 1;}};const {next}=own;console.log(`${next()}`);",
		"class Box{next():number{return 1;}}const own=new Box();const {next}=own;console.log(`${next()}`);",
		"class Box{next():number{return 1;}}const own=new Box();const view:{next:()=>number}=own;console.log(`${view.next()}`);",
		"const own={next():number{return 1;}};const view:{next:()=>number}=own;const {next}=view;console.log(`${next()}`);",
		"const own={};const {constructor}=own;",
	} {
		_, err := lowerSource(test, source)
		var refused *Refused
		var gap *NotYet
		if !errors.As(err, &refused) && !errors.As(err, &gap) {
			test.Fatalf("got %v, want a refusal before a method becomes an own-slot load", err)
		}
	}
}

func TestGenericIteratorViewsPreserveNativeArguments(test *testing.T) {
	test.Parallel()
	source := `class Sequence<T>{readonly data:readonly T[];constructor(data:readonly T[]){this.data=data;}[Symbol.iterator]():Sequence<T>{return this;}next():{value:number;done:boolean}{const ignored=this.data[0];return{value:1,done:false};}}
 const source:Sequence<number|undefined>=new Sequence([1]);for(const value of source){break;}`
	_, err := lowerSource(test, source)
	var refused *Refused
	if !errors.As(err, &refused) || !strings.Contains(refused.What, "nominal ancestry") {
		test.Fatalf("got %v, want the earlier nominal generic-view refusal", err)
	}
}

func TestIteratorMapperIndexHasNumberRepresentation(test *testing.T) {
	test.Parallel()
	source := `class Sequence{[Symbol.iterator]():Sequence{return this;}next():{value:number;done:boolean}{return{value:1,done:true};}}const source=new Sequence();const values=Array.from(source,(value:number,index:number|undefined):number=>value);`
	_, err := lowerSource(test, source)
	var gap *NotYet
	if !errors.As(err, &gap) || !strings.Contains(gap.What, "index representation") {
		test.Fatalf("got %v, want a mapper index representation refusal", err)
	}
}

func TestIteratorSymbolKeysAreNotStringKeys(test *testing.T) {
	test.Parallel()
	_, err := lowerSource(
		test,
		`const source={value:1,[Symbol.iterator](){return{next(){return{value:1,done:true};}}}};console.log(Object.keys(source).join(','));`,
	)
	var gap *NotYet
	if !errors.As(err, &gap) || !strings.Contains(gap.What, "symbol-key storage") {
		test.Fatalf("got %v, want explicit symbol-key storage refusal", err)
	}
}
