package lower

import (
	"github.com/system-inc/adamic/internal/ir"
	"testing"
)

func TestReadinessElisionRequiresDominatingAssignment(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		name, source string
		checked      int
	}{
		{"branch-join", `function f(flag:boolean):void { let value:number=undefined!; if(flag){value=1;}else{value=2;} console.log(value.toString()); } f(true);`, 0},
		{"zero-iteration", `function f(flag:boolean):void { let value:number=undefined!; while(flag){value=1;flag=false;} console.log(value.toString()); } f(false);`, 1},
		{"loop-body", `function f(flag:boolean):void { let value:number=undefined!; while(flag){value=1;console.log(value.toString());flag=false;} } f(true);`, 0},
		{"do-loop", `function f():void { let value:number=undefined!; do{value=1;}while(false);console.log(value.toString()); } f();`, 0},
		{"captured-write", `function f():void { let value:number=undefined!; const read=():void=>{value=1;console.log(value.toString());}; read(); } f();`, 0},
		{"captured-read", `function f():void { let value:number=undefined!; const read=():void=>{console.log(value.toString());};read(); } f();`, 1},
		{"throw-before-write", `function fail():number{throw new Error('no');}function f():void {let value:number=undefined!;try{value=fail();}catch{}console.log(value.toString());}f();`, 1},
		{"catch-write", `function fail():number{throw new Error('no');}function f():void {let value:number=undefined!;try{value=fail();}catch{value=2;}console.log(value.toString());}f();`, 0},
		{"finally-write", `function f():void {let value:number=undefined!;try{throw new Error('no');}catch{}finally{value=2;}console.log(value.toString());}f();`, 0},
		{"field-write", `class Box {value:number=undefined!;}const box=new Box();box.value=1;console.log(box.value.toString());`, 0},
		{"field-join", `class Box {value:number=undefined!;}function f(flag:boolean):void{const box=new Box();if(flag){box.value=1;}else{box.value=2;}console.log(box.value.toString());}f(true);`, 0},
		{"field-missing-join", `class Box {value:number=undefined!;}function f(flag:boolean):void{const box=new Box();if(flag){box.value=1;}console.log(box.value.toString());}f(false);`, 1},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			program, err := lowerTypeScriptAssertionSource(t, probe.source)
			if err != nil {
				t.Fatal(err)
			}
			checked := 0
			count := func(node any) bool {
				switch read := node.(type) {
				case ir.Read:
					if read.Readiness != "" {
						checked++
					}
				case ir.Property:
					if read.Readiness != "" {
						checked++
					}
				}
				return true
			}
			walk(program.Main, count)
			for _, function := range program.Functions {
				walk(function.Body, count)
			}
			for _, site := range program.PlaceholderChecks {
				if site.Status == "checked" {
					checked++
				}
			}
			if checked != probe.checked {
				t.Fatalf("%d readiness checks, want %d", checked, probe.checked)
			}
		})
	}
}
