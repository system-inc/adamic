package lower

import (
	"context"
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
