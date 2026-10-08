package lower

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

func TestProgramRegionMembership(t *testing.T) {
	path := filepath.Join(t.TempDir(), "main.a")
	source := `import type { Weak } from 'adamic';
interface Node { id: string; children: Node[]; parent: Weak<Node>; }
interface Leaf { text: string; }
function run():void {
 const root:Node={id:'r',children:[],parent:undefined};
 const child:Node={id:'c',children:[],parent:root};root.children.push(child);
 const leaf:Leaf={text:'leaf'};console.log(leaf.text);
 const cache=new Map<string,number>();cache.set('r',1);
 const scratch:number[]=[1,2];console.log('scratch');
 console.log('tree');
}run();`
	if err := os.WriteFile(path, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	loaded, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	off, err := LowerWithOptions(context.Background(), loaded, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if off.ProgramRegion || len(off.ProgramTypes) != 0 {
		t.Fatal("prototype enabled by default")
	}
	on, err := LowerWithOptions(context.Background(), loaded, Options{ProgramRegion: true})
	if err != nil {
		t.Fatal(err)
	}
	markedObjects, markedArrays := 0, 0
	for _, function := range on.Functions {
		walk(function.Body, func(node any) bool {
			switch node := node.(type) {
			case ir.ObjectLiteral:
				for _, field := range node.Fields {
					if field.Name == "id" {
						if !node.ProgramRegion {
							t.Fatal("recursive Node left counted")
						}
						markedObjects++
					}
					if field.Name == "text" && node.ProgramRegion {
						t.Fatal("acyclic Leaf made a member")
					}
				}
			case ir.MapNew:
				if node.ProgramRegion {
					t.Fatal("scalar relation scratch made a member")
				}
			case ir.ArrayLiteral:
				if node.Element == ir.Number && node.ProgramRegion {
					t.Fatal("scalar temporary array made a member")
				}
				if node.ProgramRegion {
					markedArrays++
				}
			}
			return true
		})
	}
	if markedObjects != 2 || markedArrays != 2 {
		t.Fatalf("membership objects=%d arrays=%d", markedObjects, markedArrays)
	}
}

func TestProgramRegionViewCycleIsNotOwnership(t *testing.T) {
	edges := map[string][]string{"a": {"b"}, "b": {"a"}, "string": nil}
	if got := ProgramRegionSCC(edges, map[string][]string{"a": {"string"}}); len(got) != 0 {
		t.Fatalf("view-only SCC selected: %v", got)
	}
	edges["a"] = append(edges["a"], "a")
	if got := ProgramRegionSCC(edges, map[string][]string{"a": {"a", "string"}}); !got["a"] || !got["b"] || got["string"] {
		t.Fatal(got)
	}
}

func TestProgramRegionTaskBoundary(t *testing.T) {
	path := filepath.Join(t.TempDir(), "main.a")
	source := `import { parallelMap } from 'adamic';
interface Tree { readonly children: readonly Tree[]; }
const tree: Tree = {children:[]};
console.log('start');
const items: readonly number[] = [1,2];
const values = parallelMap(items, (value:number):number => value+1);
console.log('end');`
	if err := os.WriteFile(path, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	loaded, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = LowerWithOptions(context.Background(), loaded, Options{ProgramRegion: true}); err == nil || !strings.Contains(err.Error(), "Program-region members") {
		t.Fatalf("task boundary: %v", err)
	}
}
