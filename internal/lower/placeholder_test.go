package lower

import (
	"errors"
	"strings"
	"testing"
)

func TestPlaceholderTypedBoundaryAssignmentResult(test *testing.T) {
	test.Parallel()
	assertPlaceholderTypedBoundary(
		test,
		`let value:string=undefined!;let target:string=undefined!;function take(x:string):void{}take(target=value);`,
		"assignment result",
		1,
	)
}

func TestPlaceholderTypedBoundaryReturnAssignment(test *testing.T) {
	test.Parallel()
	assertPlaceholderTypedBoundary(
		test,
		`let value:string=undefined!;let target:string=undefined!;function take():string{return target=value;}take();`,
		"return",
		1,
	)
}

func TestPlaceholderTypedBoundaryAliasReset(test *testing.T) {
	test.Parallel()
	assertPlaceholderTypedBoundary(
		test,
		`class Slot {value:string=undefined!;} const source=new Slot();source.value='written';const alias=source;alias.value=undefined!;const typed:string=source.value;`,
		"assignment",
		1,
	)
}

func TestPlaceholderTypedBoundaryArgument(test *testing.T) {
	test.Parallel()
	assertPlaceholderTypedBoundary(
		test,
		`let value:string=undefined!;function take(x:string):void{}take(value);`,
		"argument",
		1,
	)
}

func TestPlaceholderTypedBoundarySavedCopy(test *testing.T) {
	test.Parallel()
	assertPlaceholderTypedBoundary(
		test,
		`let value:string=undefined!;const saved=value;value='written';function take(x:string):void{}take(saved);`,
		"argument",
		1,
	)
}

func TestPlaceholderTypedBoundaryProperty(test *testing.T) {
	test.Parallel()
	assertPlaceholderTypedBoundary(
		test,
		`let value:string=undefined!;console.log(value.length.toString());`,
		"property read",
		1,
	)
}

func TestPlaceholderTypedBoundaryCall(test *testing.T) {
	test.Parallel()
	assertPlaceholderTypedBoundary(test, `let value:()=>void=undefined!;value();`, "call", 1)
}

func TestPlaceholderTypedBoundaryReturn(test *testing.T) {
	test.Parallel()
	assertPlaceholderTypedBoundary(
		test,
		`let value:string=undefined!;function take():string{return value;}take();`,
		"return",
		1,
	)
}

func TestPlaceholderTypedBoundaryTypedLocal(test *testing.T) {
	test.Parallel()
	assertPlaceholderTypedBoundary(test, `let value:string=undefined!;const target:string=value;`, "assignment", 1)
}

func TestPlaceholderTypedBoundaryOrdinaryField(test *testing.T) {
	test.Parallel()
	assertPlaceholderTypedBoundary(
		test,
		`let value:string=undefined!;const target={text:''};target.text=value;`,
		"assignment",
		1,
	)
}

func TestPlaceholderTypedBoundaryArrayElement(test *testing.T) {
	test.Parallel()
	assertPlaceholderTypedBoundary(test, `let value:string=undefined!;const target=[value];`, "container element", 1)
}

func TestPlaceholderTypedBoundaryOptionalArrayElement(test *testing.T) {
	test.Parallel()
	assertPlaceholderTypedBoundary(
		test,
		`let value:string=undefined!;const target:(string|undefined)[]=[];target.push(value);`,
		"argument",
		1,
	)
}

func TestPlaceholderTypedBoundaryElementWrite(test *testing.T) {
	test.Parallel()
	assertPlaceholderTypedBoundary(
		test,
		`let value:string=undefined!;const target:(string|undefined)[]=[''];target[0]=value;`,
		"assignment",
		1,
	)
}

func TestPlaceholderTypedBoundaryOptionalReceiver(test *testing.T) {
	test.Parallel()
	assertPlaceholderTypedBoundary(
		test,
		`let value:string=undefined!;function take(x:string|undefined):void{console.log(x===undefined?'absent':'present');}take(value);`,
		"",
		0,
	)
}

func TestPlaceholderTypedBoundaryOptionalParameterLaterTUse(test *testing.T) {
	test.Parallel()
	assertPlaceholderTypedBoundary(
		test,
		`let value:string=undefined!;function take(x:string|undefined):void{if(x){const saved=x;console.log(saved.length.toString());}}take(value);`,
		"property read",
		1,
	)
}

func TestPlaceholderTypedBoundaryPresence(test *testing.T) {
	test.Parallel()
	assertPlaceholderTypedBoundary(
		test,
		`let value:number=undefined!;if(value){console.log('present');}console.log((value??3).toString());`,
		"",
		0,
	)
}

func TestPlaceholderTypedBoundaryDominatingWrite(test *testing.T) {
	test.Parallel()
	assertPlaceholderTypedBoundary(test, `let value:number=undefined!;value=0;console.log(value.toString());`, "", 0)
}

func TestPlaceholderTypedBoundarySavedAfterWrite(test *testing.T) {
	test.Parallel()
	assertPlaceholderTypedBoundary(
		test,
		`let value:number=undefined!;value=0;const saved=value;console.log(saved.toString());`,
		"",
		0,
	)
}

func TestPlaceholderTypedBoundaryPlaceholderFieldCopy(test *testing.T) {
	test.Parallel()
	assertPlaceholderTypedBoundary(
		test,
		`class Box{value:string=undefined!;}const a=new Box();const b=new Box();b.value=a.value;console.log(b.value===undefined?'absent':'present');`,
		"",
		0,
	)
}

func assertPlaceholderTypedBoundary(test *testing.T, source, use string, expectedChecks int) {
	test.Helper()
	program, lowerTypeScriptAssertionSourceError := lowerTypeScriptAssertionSource(test, source)
	if lowerTypeScriptAssertionSourceError != nil {
		test.Fatal(lowerTypeScriptAssertionSourceError)
	}
	checked := 0
	for _, site := range program.PlaceholderChecks {
		if site.Status == "checked" {
			checked++
			if site.Use != use {
				test.Fatalf("use %q, want %q", site.Use, use)
			}
		}
	}
	if checked != expectedChecks {
		test.Fatalf("checked sites %d, want %d: %#v", checked, expectedChecks, program.PlaceholderChecks)
	}
}

func TestPlaceholderWeakSlotStaysNotYet(test *testing.T) {
	test.Parallel()
	_, lowerSourceError := lowerSource(
		test,
		`import type { Weak } from 'adamic'; let value: Weak<{value:number}> = undefined!;`,
	)
	var notYet *NotYet
	if !errors.As(lowerSourceError, &notYet) || !strings.Contains(notYet.What, "a placeholder slot holding Weak") {
		test.Fatalf("unsupported weak placeholder: %v", lowerSourceError)
	}
}
