package lower

import (
	"errors"
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
			_, err := lowerSource(t, probe.source)
			var refused *Refused
			if !errors.As(err, &refused) || refused.What != "the non-null assertion !" || refused.Fix != "write ?? panic('why it can't be missing'), or narrow and handle the missing case" {
				t.Fatalf("want historical .a assertion refused, got %v", err)
			}
		})
	}
}
