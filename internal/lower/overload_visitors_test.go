package lower

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
)

func TestOverloadVisitorPaths(t *testing.T) {
	base := `interface Base {readonly kind:number;}
interface Named extends Base {readonly text:string;}
function visit<T extends Base>(nodes:readonly T[],visitor:(node:T)=>void):void;
function visit(nodes:readonly Base[],visitor:(node:Base)=>void):void {BODY}
const nodes:readonly Named[]=[{kind:1,text:'name'}];
visit(nodes,node=>console.log(node.text));`
	for _, probe := range []struct{ name, body, path string }{
		{"unguarded presence", `for(const node of nodes) visitor(node);`, "visitor argument path is unproven"},
		{"fabricated node", `visitor({kind:0});`, "visitor argument path is unproven"},
		{"stored callback", `const holder={visitor}; holder.visitor({kind:0});`, "visitor argument path"},
		{"returned callback", `function escape(){return visitor;} escape()({kind:0});`, "visitor argument path"},
		{"intervening element effect", `function touch(node:Base):void{} for(const node of nodes){if(node!==undefined){touch(node);visitor(node);}}`, "visitor argument path is unproven"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			_, err := lowerSource(t, strings.Replace(base, "BODY", probe.body, 1))
			if err == nil || !strings.Contains(err.Error(), probe.path) {
				t.Fatalf("wanted %s: %v", probe.path, err)
			}
		})
	}
}

func TestOverloadVisitorResultIndependent(t *testing.T) {
	source := `interface Base {readonly kind:number;}
interface Named extends Base {readonly text:string;}
function visit<T extends Base,U extends Base>(nodes:readonly T[],visitor:(node:T)=>void):readonly U[];
function visit(nodes:readonly Base[],visitor:(node:Base)=>void):readonly Base[]{for(const node of nodes){if(node!==undefined)visitor(node);} return nodes;}
const nodes:readonly Named[]=[{kind:1,text:'name'}];
console.log(visit<Named,Named>(nodes,node=>console.log(node.text))[0]?.text??'missing');`
	for _, extension := range []string{".a", ".ts"} {
		path := filepath.Join(t.TempDir(), "main"+extension)
		if err := os.WriteFile(path, []byte(source), 0600); err != nil {
			t.Fatal(err)
		}
		program, err := load.Load([]string{path})
		if err != nil {
			t.Fatal(err)
		}
		_, err = Lower(context.Background(), program)
		var refused *Refused
		if !errors.As(err, &refused) || !strings.Contains(refused.What, "result[]") || !strings.Contains(refused.What, "result element covariance") {
			t.Fatalf("input proof must not bless U result: %v", err)
		}
	}
}

func TestOverloadVisitorResultStorage(t *testing.T) {
	source := `interface Base {readonly kind:number;}
interface Named extends Base {readonly text:string;}
interface Wide {readonly value:string|number;}
interface Narrow {readonly value:string;}
function visit<T extends Base>(nodes:readonly T[],visitor:(node:T)=>readonly Wide[]):void;
function visit(nodes:readonly Base[],visitor:(node:Base)=>readonly Wide[]):void {for(const node of nodes){if(node!==undefined)visitor(node);}}
const nodes:readonly Named[]=[{kind:1,text:'name'}];
function visitor(node:Named):readonly Narrow[]{return [{value:node.text}];}
visit(nodes,visitor);`
	path := filepath.Join(t.TempDir(), "main.ts")
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	_, err = Lower(context.Background(), program)
	if err == nil || !strings.Contains(err.Error(), "visitor result path with different field or element storage") {
		t.Fatalf("callback covariance cannot rebox result fields: %v", err)
	}
}
