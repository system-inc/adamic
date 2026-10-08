package lower

import (
	"errors"
	"strings"
	"testing"
)

func TestPlaceholderTypedBoundaries(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		name, source, use string
		checked           int
	}{
		{"assignment-result", `let value:string=undefined!;let target:string=undefined!;function take(x:string):void{}take(target=value);`, "assignment result", 1},
		{"return-assignment", `let value:string=undefined!;let target:string=undefined!;function take():string{return target=value;}take();`, "return", 1},
		{"alias-reset", `class Slot {value:string=undefined!;} const source=new Slot();source.value='written';const alias=source;alias.value=undefined!;const typed:string=source.value;`, "assignment", 1},
		{"argument", `let value:string=undefined!;function take(x:string):void{}take(value);`, "argument", 1},
		{"saved-copy", `let value:string=undefined!;const saved=value;value='written';function take(x:string):void{}take(saved);`, "argument", 1},
		{"property", `let value:string=undefined!;console.log(value.length.toString());`, "property read", 1},
		{"call", `let value:()=>void=undefined!;value();`, "call", 1},
		{"return", `let value:string=undefined!;function take():string{return value;}take();`, "return", 1},
		{"typed-local", `let value:string=undefined!;const target:string=value;`, "assignment", 1},
		{"ordinary-field", `let value:string=undefined!;const target={text:''};target.text=value;`, "assignment", 1},
		{"array-element", `let value:string=undefined!;const target=[value];`, "container element", 1},
		{"optional-array-element", `let value:string=undefined!;const target:(string|undefined)[]=[];target.push(value);`, "argument", 1},
		{"element-write", `let value:string=undefined!;const target:(string|undefined)[]=[''];target[0]=value;`, "assignment", 1},
		{"optional-receiver", `let value:string=undefined!;function take(x:string|undefined):void{console.log(x===undefined?'absent':'present');}take(value);`, "", 0},
		{"optional-parameter-later-T-use", `let value:string=undefined!;function take(x:string|undefined):void{if(x){const saved=x;console.log(saved.length.toString());}}take(value);`, "property read", 1},
		{"presence", `let value:number=undefined!;if(value){console.log('present');}console.log((value??3).toString());`, "", 0},
		{"dominating-write", `let value:number=undefined!;value=0;console.log(value.toString());`, "", 0},
		{"saved-after-write", `let value:number=undefined!;value=0;const saved=value;console.log(saved.toString());`, "", 0},
		{"placeholder-field-copy", `class Box{value:string=undefined!;}const a=new Box();const b=new Box();b.value=a.value;console.log(b.value===undefined?'absent':'present');`, "", 0},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			program, err := lowerTypeScriptAssertionSource(t, probe.source)
			if err != nil {
				t.Fatal(err)
			}
			checked := 0
			for _, site := range program.PlaceholderChecks {
				if site.Status == "checked" {
					checked++
					if site.Use != probe.use {
						t.Fatalf("use %q, want %q", site.Use, probe.use)
					}
				}
			}
			if checked != probe.checked {
				t.Fatalf("checked sites %d, want %d: %#v", checked, probe.checked, program.PlaceholderChecks)
			}
		})
	}
}

func TestPlaceholderWeakSlotStaysNotYet(t *testing.T) {
	_, err := lowerSource(t, `import type { Weak } from 'adamic'; let value: Weak<{value:number}> = undefined!;`)
	var notYet *NotYet
	if !errors.As(err, &notYet) || !strings.Contains(notYet.What, "a placeholder slot holding Weak") {
		t.Fatalf("unsupported weak placeholder: %v", err)
	}
}
