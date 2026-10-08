package lower

import (
	"errors"
	"github.com/system-inc/adamic/internal/ir"
	"testing"
)

func TestAssignmentProofs(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"enum E { A, B } let slot:E=E.B; const member:E.A=slot=E.A;",
		"enum E { A, B } let slot:E=E.B; function value():E.A { return slot=E.A; }",
		"let value:string|undefined=undefined; function word():string|undefined { return 'present'; } const result=word(); if(result!==undefined) { const length=(value=result).length; }",
		"type A={readonly kind:'a'; readonly value:string}; type B={readonly kind:'b'; readonly value:number}; let slot:A|B={kind:'b',value:1}; function input():A|B{return {kind:'a',value:'text'};} const value=(slot=input() as A).value;",
	} {
		if _, err := lowerSource(t, source); err != nil {
			t.Fatal(err)
		}
	}
}

func TestAssignmentProofsDoNotInventMembersOrUniqueAliases(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"enum E { A, B } function value(e:E):E.A { if(e===E.B) return E.A; let slot:E=E.B; return slot=e; }",
		"interface Animal {readonly name:string} interface Dog extends Animal {readonly bark:string} const dog:Dog={name:'dog',bark:'woof'}; let dogs:Dog[]=[]; const animals:Animal[]=(dogs=[dog]); animals.push({name:'cat'});",
	} {
		_, err := lowerSource(t, source)
		var refused *Refused
		if !errors.As(err, &refused) {
			t.Fatalf("want refusal, got %v", err)
		}
	}
}

// Effects is transparent to closed-input analysis, including its internal stores.
func TestAssignmentProofsClosedInputStillInspectsEffects(t *testing.T) {
	program, err := lowerSource(t, `function parser(scanner: { scan: () => number }): () => number {
 let token=0;
 function read(): number { return token; }
 function next(): number { return token=scanner.scan(); }
 next(); return read;
 } console.log(String(parser({scan: () => 1})()));`)
	if err != nil {
		t.Fatal(err)
	}
	l := lowering{result: program}
	input := -1
	for i, local := range program.Locals {
		if local.Name == "scanner" {
			input = i
		}
	}
	if input < 0 || !l.closedFrameInput(input) {
		t.Fatal("closed input was not proven")
	}
	owner := program.Locals[input].Function
	program.Functions[owner].Body = append(program.Functions[owner].Body, ir.Evaluate{Value: ir.Effects{Body: []ir.Statement{ir.SetProperty{}}, Result: ir.NumberConstant{}}})
	if l.closedFrameInput(input) {
		t.Fatal("an internal field store was hidden by Effects")
	}
}
