package lower

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// Exact TypeScript 6.0.3 sys.ts:1672-1691. Stubs isolate only the Profile
// member reasons; the original helper calls and source columns stay unchanged.
const inspectorCleanupPaths = "        function cleanupPaths(profile: import(\"inspector\").Profiler.Profile) {\n            let externalFileCounter = 0;\n            const remappedPaths = new Map<string, string>();\n            const normalizedDir = normalizeSlashes(_path.dirname(executingFilePath));\n            // Windows rooted dir names need an extra `/` prepended to be valid file:/// urls\n            const fileUrlRoot = `file://${getRootLength(normalizedDir) === 1 ? \"\" : \"/\"}${normalizedDir}`;\n            for (const node of profile.nodes) {\n                if (node.callFrame.url) {\n                    const url = normalizeSlashes(node.callFrame.url);\n                    if (containsPath(fileUrlRoot, url, useCaseSensitiveFileNames)) {\n                        node.callFrame.url = getRelativePathToDirectoryOrUrl(fileUrlRoot, url, fileUrlRoot, createGetCanonicalFileName(useCaseSensitiveFileNames), /*isAbsolutePathAnUrl*/ true);\n                    }\n                    else if (!nativePattern.test(url)) {\n                        node.callFrame.url = (remappedPaths.has(url) ? remappedPaths : remappedPaths.set(url, `external${externalFileCounter}.js`)).get(url)!;\n                        externalFileCounter++;\n                    }\n                }\n            }\n            return profile;\n        }\n"

func TestInspectorProfileMeterPositions(test *testing.T) {
	test.Parallel()
	prefix := `import type {Profiler} from 'node:inspector';
const _path = {dirname(p:string):string{return p;}};
const executingFilePath='file:///project/a.ts', useCaseSensitiveFileNames=true, nativePattern=/native/;
function normalizeSlashes(p:string):string{return p;}
function getRootLength(_p:string):number{return 1;}
function containsPath(_a:string,_b:string,_c:boolean):boolean{return true;}
function createGetCanonicalFileName(_s:boolean):(p:string)=>string{return (p:string)=>p;}
function getRelativePathToDirectoryOrUrl(_a:string,b:string,_c:string,_d:(p:string)=>string,_e:boolean):string{return b;}
`
	source := prefix + strings.Repeat("\n", 1671-strings.Count(prefix, "\n")) + inspectorCleanupPaths
	path := writeAgreementSource(test, "sys.ts", source)
	program, loadError := load.Load([]string{path})
	if loadError != nil {
		test.Fatal(loadError)
	}
	file := program.Files()[0]
	checked, release := program.Checker(context.Background(), file)
	defer release()
	l := lowering{program: program, checker: checked, result: &ir.Program{}}
	var got []string
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		name := l.nodeLibraryMember(node)
		if name == "node:inspector.Profile.nodes" ||
			name == "node:inspector.ProfileNode.callFrame" ||
			name == "node:inspector.CallFrame.url" {
			kind := ""
			if node.Kind == ast.KindIdentifier {
				kind = "name"
			} else if node.Kind == ast.KindPropertyAccessExpression {
				kind = "access"
			}
			if kind != "" {
				got = append(got, fmt.Sprintf("%s %s %s", filepath.Base(program.Where(node)), name, kind))
				if nodeLibraryRefusalError := l.nodeLibraryRefusal(node); nodeLibraryRefusalError != nil {
					test.Errorf("Profile member still refused: %v", nodeLibraryRefusalError)
				}
			}
		}
		node.ForEachChild(visit)
		return false
	}
	file.AsNode().ForEachChild(visit)
	slices.Sort(got)
	want := []string{
		"sys.ts:1678:32 node:inspector.Profile.nodes access",
		"sys.ts:1678:40 node:inspector.Profile.nodes name",
		"sys.ts:1679:21 node:inspector.CallFrame.url access",
		"sys.ts:1679:21 node:inspector.ProfileNode.callFrame access",
		"sys.ts:1679:26 node:inspector.ProfileNode.callFrame name",
		"sys.ts:1679:36 node:inspector.CallFrame.url name",
		"sys.ts:1680:50 node:inspector.CallFrame.url access",
		"sys.ts:1680:50 node:inspector.ProfileNode.callFrame access",
		"sys.ts:1680:55 node:inspector.ProfileNode.callFrame name",
		"sys.ts:1680:65 node:inspector.CallFrame.url name",
		"sys.ts:1682:25 node:inspector.CallFrame.url access",
		"sys.ts:1682:25 node:inspector.ProfileNode.callFrame access",
		"sys.ts:1682:30 node:inspector.ProfileNode.callFrame name",
		"sys.ts:1682:40 node:inspector.CallFrame.url name",
		"sys.ts:1685:25 node:inspector.CallFrame.url access",
		"sys.ts:1685:25 node:inspector.ProfileNode.callFrame access",
		"sys.ts:1685:30 node:inspector.ProfileNode.callFrame name",
		"sys.ts:1685:40 node:inspector.CallFrame.url name",
	}
	if !slices.Equal(got, want) {
		test.Fatalf("Profile meter positions: got %q, want %q", got, want)
	}
	test.Log("18 exact Profile source positions; zero Profile member refusals")
}

func TestInspectorProfileOtherMembersStayRefused(test *testing.T) {
	test.Parallel()
	for name, source := range map[string]string{
		"Debugger url":         `import type {Debugger} from 'node:inspector'; function read(frame:Debugger.CallFrame):string{return frame.url;}`,
		"Debugger view":        `import type {Debugger} from 'node:inspector'; function read(view:{url:string}):string{return view.url;} function project(frame:Debugger.CallFrame):string{return read(frame);}`,
		"Runtime functionName": `import type {Runtime} from 'node:inspector'; function read(frame:Runtime.CallFrame):string{return frame.functionName;}`,
		"Session connect":      `import type {Session} from 'node:inspector'; function use(s:Session):void{s.connect();}`,
	} {
		test.Run(name, func(test *testing.T) {
			_, lowerSourceAsError := lowerSourceAs(test, source, ".ts")
			var notYet *NotYet
			if !errors.As(lowerSourceAsError, &notYet) || !strings.Contains(notYet.What, "node:inspector.") {
				test.Fatalf("unbound inspector member must stay NotYet: %v", lowerSourceAsError)
			}
		})
	}
}

func TestInspectorSessionMeterSitesStayRefused(test *testing.T) {
	test.Parallel()
	source := `import type * as Inspector from 'node:inspector';
function use(inspector:typeof Inspector,session:Inspector.Session):void {
 if(!inspector || !inspector.Session)return;
 const created = new inspector.Session();
 session.connect();
 session.post('Profiler.enable',()=>{});
 session.post('Profiler.start',()=>{});
 session.post('Profiler.stop',()=>{});
 session.disconnect();
}`
	path := writeAgreementSource(test, "session.ts", source)
	program, loadError := load.Load([]string{path})
	if loadError != nil {
		test.Fatal(loadError)
	}
	file := program.Files()[0]
	checked, release := program.Checker(context.Background(), file)
	defer release()
	l := lowering{program: program, checker: checked, result: &ir.Program{}}
	sites := map[string]bool{}
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		var notYet *NotYet
		if nodeLibraryRefusalError := l.nodeLibraryRefusal(
			node,
		); errors.As(nodeLibraryRefusalError, &notYet) && strings.HasPrefix(notYet.What, "node:inspector.Session") {
			sites[notYet.Where+": "+notYet.What] = true
		}
		node.ForEachChild(visit)
		return false
	}
	file.AsNode().ForEachChild(visit)
	if len(sites) != 15 {
		test.Fatalf("want 15 unchanged Session meter sites, got %d: %v", len(sites), sites)
	}
}
