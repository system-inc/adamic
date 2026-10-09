package lower

import (
	"errors"
	"strings"
	"testing"
)

func TestPlaceholderTypedBoundaryAssignmentResult(t *testing.T) {
	t.Parallel()
	assertPlaceholderTypedBoundary(t, `let value:string=undefined!;let target:string=undefined!;function take(x:string):void{}take(target=value);`, "assignment result", 1)
}

func TestPlaceholderTypedBoundaryReturnAssignment(t *testing.T) {
	t.Parallel()
	assertPlaceholderTypedBoundary(t, `let value:string=undefined!;let target:string=undefined!;function take():string{return target=value;}take();`, "return", 1)
}

func TestPlaceholderTypedBoundaryAliasReset(t *testing.T) {
	t.Parallel()
	assertPlaceholderTypedBoundary(t, `class Slot {value:string=undefined!;} const source=new Slot();source.value='written';const alias=source;alias.value=undefined!;const typed:string=source.value;`, "assignment", 1)
}

func TestPlaceholderTypedBoundaryArgument(t *testing.T) {
	t.Parallel()
	assertPlaceholderTypedBoundary(t, `let value:string=undefined!;function take(x:string):void{}take(value);`, "argument", 1)
}

func TestPlaceholderTypedBoundarySavedCopy(t *testing.T) {
	t.Parallel()
	assertPlaceholderTypedBoundary(t, `let value:string=undefined!;const saved=value;value='written';function take(x:string):void{}take(saved);`, "argument", 1)
}

func TestPlaceholderTypedBoundaryProperty(t *testing.T) {
	t.Parallel()
	assertPlaceholderTypedBoundary(t, `let value:string=undefined!;console.log(value.length.toString());`, "property read", 1)
}

func TestPlaceholderTypedBoundaryCall(t *testing.T) {
	t.Parallel()
	assertPlaceholderTypedBoundary(t, `let value:()=>void=undefined!;value();`, "call", 1)
}

func TestPlaceholderTypedBoundaryReturn(t *testing.T) {
	t.Parallel()
	assertPlaceholderTypedBoundary(t, `let value:string=undefined!;function take():string{return value;}take();`, "return", 1)
}

func TestPlaceholderTypedBoundaryTypedLocal(t *testing.T) {
	t.Parallel()
	assertPlaceholderTypedBoundary(t, `let value:string=undefined!;const target:string=value;`, "assignment", 1)
}

func TestPlaceholderTypedBoundaryOrdinaryField(t *testing.T) {
	t.Parallel()
	assertPlaceholderTypedBoundary(t, `let value:string=undefined!;const target={text:''};target.text=value;`, "assignment", 1)
}

func TestPlaceholderTypedBoundaryArrayElement(t *testing.T) {
	t.Parallel()
	assertPlaceholderTypedBoundary(t, `let value:string=undefined!;const target=[value];`, "container element", 1)
}

func TestPlaceholderTypedBoundaryOptionalArrayElement(t *testing.T) {
	t.Parallel()
	assertPlaceholderTypedBoundary(t, `let value:string=undefined!;const target:(string|undefined)[]=[];target.push(value);`, "argument", 1)
}

func TestPlaceholderTypedBoundaryElementWrite(t *testing.T) {
	t.Parallel()
	assertPlaceholderTypedBoundary(t, `let value:string=undefined!;const target:(string|undefined)[]=[''];target[0]=value;`, "assignment", 1)
}

func TestPlaceholderTypedBoundaryOptionalReceiver(t *testing.T) {
	t.Parallel()
	assertPlaceholderTypedBoundary(t, `let value:string=undefined!;function take(x:string|undefined):void{console.log(x===undefined?'absent':'present');}take(value);`, "", 0)
}

func TestPlaceholderTypedBoundaryOptionalParameterLaterTUse(t *testing.T) {
	t.Parallel()
	assertPlaceholderTypedBoundary(t, `let value:string=undefined!;function take(x:string|undefined):void{if(x){const saved=x;console.log(saved.length.toString());}}take(value);`, "property read", 1)
}

func TestPlaceholderTypedBoundaryPresence(t *testing.T) {
	t.Parallel()
	assertPlaceholderTypedBoundary(t, `let value:number=undefined!;if(value){console.log('present');}console.log((value??3).toString());`, "", 0)
}

func TestPlaceholderTypedBoundaryDominatingWrite(t *testing.T) {
	t.Parallel()
	assertPlaceholderTypedBoundary(t, `let value:number=undefined!;value=0;console.log(value.toString());`, "", 0)
}

func TestPlaceholderTypedBoundarySavedAfterWrite(t *testing.T) {
	t.Parallel()
	assertPlaceholderTypedBoundary(t, `let value:number=undefined!;value=0;const saved=value;console.log(saved.toString());`, "", 0)
}

func TestPlaceholderTypedBoundaryPlaceholderFieldCopy(t *testing.T) {
	t.Parallel()
	assertPlaceholderTypedBoundary(t, `class Box{value:string=undefined!;}const a=new Box();const b=new Box();b.value=a.value;console.log(b.value===undefined?'absent':'present');`, "", 0)
}

func assertPlaceholderTypedBoundary(t *testing.T, source, use string, expectedChecks int) {
	t.Helper()
	program, err := lowerTypeScriptAssertionSource(t, source)
	if err != nil {
		t.Fatal(err)
	}
	checked := 0
	for _, site := range program.PlaceholderChecks {
		if site.Status == "checked" {
			checked++
			if site.Use != use {
				t.Fatalf("use %q, want %q", site.Use, use)
			}
		}
	}
	if checked != expectedChecks {
		t.Fatalf("checked sites %d, want %d: %#v", checked, expectedChecks, program.PlaceholderChecks)
	}
}

func TestPlaceholderWeakSlotStaysNotYet(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, `import type { Weak } from 'adamic'; let value: Weak<{value:number}> = undefined!;`)
	var notYet *NotYet
	if !errors.As(err, &notYet) || !strings.Contains(notYet.What, "a placeholder slot holding Weak") {
		t.Fatalf("unsupported weak placeholder: %v", err)
	}
}
