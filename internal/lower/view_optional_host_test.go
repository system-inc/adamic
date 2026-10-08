package lower

import (
	"strings"
	"testing"
)

func TestOptionalHostMethodBoundaries(t *testing.T) {
	for _, test := range []struct{ name, source, pin string }{
		{"writable unknown", `interface Base { run: unknown; } interface Target {run?(path:string):string;} function probe(value:Base):void {const host=value as Target; const bound=host.run?.bind(host);} probe({run:(path:string):string=>path});`, "adamic/no-unchecked-cast"},
		{"function property", `interface Base {readonly run: unknown;} interface Target {run?: (path:string)=>string;} function probe(value:Base):void {const host=value as Target;} probe({run:(path:string):string=>path});`, "adamic/no-unchecked-cast"},
		{"ordinary optional method", `interface Target {run?(path:string):string;} const host:Target={run(path:string):string{return path;}}; const bound=host.run?.bind(host);`, "unbound-method"},
		{"detached method", `interface Base {readonly run: unknown;} interface Target {run?(path:string):string;} function probe(value:Base):void {const host=value as Target; const detached=host.run;} probe({run:(path:string):string=>path});`, "unbound-method"},
		{"invented Error data", `interface HostError extends Error {code?:string;} const host=new Error("plain") as HostError; console.log(host.code??"missing");`, "adamic/no-unchecked-cast"},
		{"invented Error method", `interface HostError extends Error {run?(path:string):string;} const host=new Error("plain") as HostError; const bound=host.run?.bind(host);`, "adamic/no-unchecked-cast"},
		{"receiver method realpath-binding", `// Original tsc declaration/read; adjacent carriers and helpers reduced.

interface Base {readonly getCurrentDirectory:()=>string;readonly realpath:unknown;}
interface Target {getCurrentDirectory(): string;realpath?(path: string): string;}
function probe(value:Base):void {const host=value as Target;const wrapped={realpath:host.realpath?.bind(host)};console.log(wrapped.realpath===undefined?"none":wrapped.realpath("abc"));}
probe({getCurrentDirectory:():string=>"root",realpath(path:string):string {return path+"/3";}});
`, "erases its receiver"},
		{"receiver method directory-condition", "// Original tsc declaration/read; adjacent carriers and helpers reduced.\n\ninterface Base {readonly fileExists:(path:string)=>boolean;readonly directoryExists:unknown;}\ninterface Target {fileExists(path: string): boolean;directoryExists?(path: string): boolean;}\nfunction probe(value:Base):void {const host=value as Target;function directoryExists(path:string):boolean{return path.length===3;}const wrapped={directoryExists:host.directoryExists && directoryExists};console.log(`${wrapped.directoryExists===undefined?false:wrapped.directoryExists(\"abc\")}`);}\nprobe({fileExists:(path:string):boolean=>true,directoryExists(path:string):boolean {return true;}});\n", "erases its receiver"},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := lowerSource(t, test.source)
			if err == nil || !strings.Contains(err.Error(), test.pin) {
				t.Fatalf("want %s, got %v", test.pin, err)
			}
		})
	}
}
