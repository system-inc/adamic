package lower

import (
	"context"
	"errors"
	"github.com/system-inc/adamic/internal/fresh"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"os"
	"path/filepath"
	"testing"
)

func TestOptionalWriteErasure(t *testing.T) {
	for _, probe := range []struct {
		name, source string
		proven       bool
	}{
		{"fresh", "({x:1} as {x:number;y?:number}).y=3;", true},
		{"stored", "const source={x:1}; const wider:{x:number;y?:number}=source; wider.y=3;", false},
		{"plain", "const actual={x:1,y:2}; const source:{x:number}=actual; const wider:{x:number;y?:number}=source; wider.y=3;", false},
		{"tag", "class Real{x=1;y=2;} const source:{x:number}=new Real(); const wider:{x:number;y?:number}=source; if(wider instanceof Real){wider.y=3;}", true},
		{"tag-condition-call", "class Real{x=1;y=2;} const source:{x:number}=new Real(); let wider:{x:number;y?:number}=source; function change():boolean{wider={x:1};return true;} if(wider instanceof Real){if(change()){wider.y=3;}}", false},
		{"tag-call", "class Real{x=1;y=2;} const source:{x:number}=new Real(); let wider:{x:number;y?:number}=source; function change():void{wider={x:1};} if(wider instanceof Real){change();wider.y=3;}", false},
		{"never-reduced", "function take(value:{kind:0}&{kind:1}):void{(value as {x:number;y?:number}).y=3;} console.log('done');", true},
		{"never", "function take(value:never):void{(value as {x:number;y?:number}).y=3;} console.log('done');", true},
	} {
		t.Run(probe.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), probe.name+".a")
			if err := os.WriteFile(path, []byte(probe.source), 0600); err != nil {
				t.Fatal(err)
			}
			loaded, err := load.Load([]string{path})
			if err != nil {
				t.Fatal(err)
			}
			program, err := Lower(context.Background(), loaded)
			if probe.name == "fresh" {
				var notYet *NotYet
				if !errors.As(err, &notYet) || notYet.What != "writing a possibly absent optional own field" {
					t.Fatalf("got %v, want the candidate optional-own-field write boundary", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			seen := 0
			inspect := func(node any) bool {
				if write, ok := node.(ir.SetProperty); ok && write.WriteContract != 0 && write.Name == "y" {
					seen++
					if write.WriteProven != probe.proven {
						t.Errorf("erasure=%v, want %v", write.WriteProven, probe.proven)
					}
				}
				return true
			}
			walk(program.Main, inspect)
			for _, function := range program.Functions {
				walk(function.Body, inspect)
			}
			if seen != 1 {
				t.Fatalf("got %d checked writes, want 1", seen)
			}
			if probe.name == "stored" {
				for _, proof := range fresh.ProveWrites(program) {
					if proof.Name == "y" && proof.Unaliased {
						t.Error("stored literal was counted unaliased")
					}
				}
			}
		})
	}
}

// A terminal base implementation does not make a virtual call unreachable.
// Its override can return an object whose slot still needs the checked write.
func TestOptionalVirtualReceiverIsNotUnreachable(t *testing.T) {
	program := &ir.Program{
		Functions: []ir.Function{
			{Returns: ir.Object, Body: []ir.Statement{ir.Panic{Message: ir.StringConstant{}}}},
			{Returns: ir.Object, Body: []ir.Statement{ir.Return{Value: ir.ObjectLiteral{}}}},
		},
		MethodTargets: map[int][]int{0: {0, 1}},
	}
	call := ir.Call{Function: 0, Virtual: 1, Returns: ir.Object}
	if trappedViewReceiver(program, call) {
		t.Fatal("a returning override was treated as unreachable")
	}
	program.MethodTargets[0] = []int{0}
	if !trappedViewReceiver(program, call) {
		t.Fatal("terminal targets lost their unreachable proof")
	}
}
