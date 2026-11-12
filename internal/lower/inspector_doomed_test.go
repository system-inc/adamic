package lower

import (
	"errors"
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

const inspectorDoomedPrelude = "import 'node:process';\nconst inspector: typeof import('inspector') = require('inspector');\nconst session = new inspector.Session();\n"

func inspectorDoomedAnalysis(t *testing.T, source string) (*lowering, *ast.SourceFile) {
	t.Helper()
	program, err := load.Load([]string{writeAgreementSource(t, "main.ts", source)})
	if err != nil {
		t.Fatal(err)
	}
	checked, release := program.Checker(t.Context(), program.Files()[0])
	t.Cleanup(release)
	return &lowering{program: program, checker: checked, result: &ir.Program{}, this: -1, functionIndex: -1}, program.Files()[0]
}
func inspectorDoomedNode(t *testing.T, file *ast.SourceFile, name string, declaration bool) *ast.Node {
	t.Helper()
	var found *ast.Node
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if declaration && node.Kind == ast.KindVariableDeclaration && node.Name().Text() == name {
			found = node.Name()
		}
		if !declaration && node.Kind == ast.KindPropertyAccessExpression && node.Name().Text() == name {
			found = node
		}
		return node.ForEachChild(visit)
	}
	file.AsNode().ForEachChild(visit)
	if found == nil {
		t.Fatalf("missing %s", name)
	}
	return found
}
func TestInspectorOtherRequireBindingsAreNotDoomed(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"fs", "source-map-support"} {
		t.Run(name, func(t *testing.T) {
			l, file := inspectorDoomedAnalysis(t, "import 'node:process'; const value: typeof import('inspector') = require('"+name+"');")
			if l.inspectorDoomedBinding(inspectorDoomedNode(t, file, "value", true), map[*ast.Symbol]bool{}) != nil {
				t.Fatalf("%s require must not create a doomed binding", name)
			}
		})
	}
}
func TestInspectorLetReceiverStaysRefused(t *testing.T) {
	t.Parallel()
	source := strings.Replace(inspectorDoomedPrelude, "const inspector:", "let inspector:", 1) + "session.post('Profiler.enable');"
	l, file := inspectorDoomedAnalysis(t, source)
	if root, _ := l.inspectorDoomedUse(inspectorDoomedNode(t, file, "post", false)); root != nil {
		t.Fatal("let receiver must not be doomed")
	}
	_, err := lowerSourceAs(t, source, ".ts")
	var stop *NotYet
	if !errors.As(err, &stop) || !strings.Contains(stop.What, "node:inspector.") {
		t.Fatalf("let receiver must stay refused: %v", err)
	}
}
func TestInspectorNonDoomedMergedMemberStaysNotYet(t *testing.T) {
	t.Parallel()
	l, file := inspectorDoomedAnalysis(t, "import type {Session} from 'node:inspector'; function use(session:Session):void{session.post('Profiler.enable');}")
	err := l.mergedFieldRead(inspectorDoomedNode(t, file, "post", false))
	var stop *NotYet
	if !errors.As(err, &stop) || stop.What != "node:inspector.Session.post" {
		t.Fatalf("non-doomed Session.post must keep rule 800 NotYet, got %v", err)
	}
}
func TestInspectorImportedReceiverStaysRefused(t *testing.T) {
	t.Parallel()
	_, err := lowerSourceAs(t, "import * as inspector from 'node:inspector';const session=new inspector.Session();session.post('Profiler.enable');", ".ts")
	var stop *NotYet
	if !errors.As(err, &stop) || stop.What != "node:inspector.Session.post" {
		t.Fatalf("imported inspector must stay refused: %v", err)
	}
}
func TestInspectorFilesystemRequireStaysRefused(t *testing.T) {
	t.Parallel()
	_, err := lowerSourceAs(t, "import 'node:process'; const fs:typeof import('fs')=require('fs');fs.readFileSync('x');", ".ts")
	var stop *NotYet
	if !errors.As(err, &stop) || !strings.Contains(stop.What, "node:module.require") {
		t.Fatalf("fs require must stay refused: %v", err)
	}
}
func TestInspectorLetSessionStaysRefused(t *testing.T) {
	t.Parallel()
	_, err := lowerSourceAs(t, inspectorDoomedPrelude+"let activeSession: import('inspector').Session = session; const s=activeSession;s.disconnect();", ".ts")
	var stop *NotYet
	if !errors.As(err, &stop) || stop.What != "node:inspector.Session.disconnect" {
		t.Fatalf("let-derived session must stay refused: %v", err)
	}
}
func TestInspectorDoomedEscapedReadRefused(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"d1.ts", "d1-control.ts"} {
		t.Run(name, func(t *testing.T) {
			source, err := os.ReadFile(filepath.Join("testdata/inspector_doomed", name))
			if err != nil {
				t.Fatal(err)
			}
			_, err = lowerSourceAs(t, string(source), ".ts")
			var stop *NotYet
			if !errors.As(err, &stop) || stop.What != "reading session" {
				t.Fatalf("%s must stop at the ordinary read: want reading session, got %v", name, err)
			}
		})
	}
}
func TestInspectorDoomedHoistedReadKeepsDeadZone(t *testing.T) {
	t.Parallel()
	source, err := os.ReadFile("testdata/inspector_doomed/hoisted.ts")
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowerSourceAs(t, string(source), ".ts")
	if err != nil {
		var stop *NotYet
		if errors.As(err, &stop) && stop.What == "reading session" {
			return
		}
		t.Fatal(err)
	}
	checked := false
	scan := func(node any) bool {
		if read, ok := node.(ir.Read); ok && program.Locals[read.Local].Name == "session" && program.DeadZoneMayThrow(read) {
			checked = true
		}
		return true
	}
	for _, function := range program.Functions {
		walk(function.Body, scan)
	}
	if !checked {
		t.Fatal("hoisted use lost its ordinary dead-zone read")
	}
}
func TestInspectorDoomedMemberValueHasPanic(t *testing.T) {
	t.Parallel()
	l, file := inspectorDoomedAnalysis(t, inspectorDoomedPrelude+"const post=session.post;")
	if err := l.declareModule(file.Statements.Nodes); err != nil {
		t.Fatal(err)
	}
	member := inspectorDoomedNode(t, file, "post", false)
	if err := l.mergedFieldRead(member); err != nil {
		t.Fatal(err)
	}
	value, handled, err := l.inspectorDoomedValue(member)
	effects, ok := value.(ir.Effects)
	if err != nil || !handled || !ok || len(effects.Body) != 2 {
		t.Fatalf("doomed member value lost its read-then-Panic: %#v %v %v", value, handled, err)
	}
	if _, ok := effects.Body[0].(ir.Evaluate); !ok {
		t.Fatal("ordinary read must precede Panic")
	}
	panic, ok := effects.Body[1].(ir.Panic)
	if !ok {
		t.Fatal("missing Panic")
	}
	const want = `node:inspector.Session.post was reached through session, whose require("inspector") throws on every host Adamic targets`
	if got := l.result.Strings[panic.Message.(ir.StringConstant).Index]; got != want {
		t.Fatalf("trap text: got %q want %q", got, want)
	}
	l.result.Main = []ir.Statement{ir.Evaluate{Value: value}}
	if err := l.checkInspectorDoomedContracts(); err != nil {
		t.Fatal(err)
	}
	l.result.Main = nil
	if err := l.checkInspectorDoomedContracts(); err == nil || !strings.Contains(err.Error(), "compiler bug: skipped inspector contract") {
		t.Fatalf("lost trap must fail closed: %v", err)
	}
}
func TestInspectorDoomedMemberFormsNeverCall(t *testing.T) {
	t.Parallel()
	for _, form := range []string{"const post=session.post;", "session.post.bind(session);", "function accept(callback:(name:string)=>void):void{} accept(session.post);"} {
		t.Run(form, func(t *testing.T) {
			program, err := lowerSourceAs(t, inspectorDoomedPrelude+form, ".ts")
			if err != nil {
				var stop *NotYet
				var refused *Refused
				if !errors.As(err, &stop) && !errors.As(err, &refused) {
					t.Fatal(err)
				}
				return
			}
			seen := false
			scan := func(node any) bool {
				if panic, ok := node.(ir.Panic); ok {
					if text, ok := panic.Message.(ir.StringConstant); ok && strings.Contains(program.Strings[text.Index], "throws on every host Adamic targets") {
						seen = true
					}
				}
				return true
			}
			walk(program.Main, scan)
			for _, f := range program.Functions {
				walk(f.Body, scan)
			}
			if !seen {
				t.Fatal("doomed member form lowered without its Panic")
			}
		})
	}
}
func TestInspectorDoomedEnableMeterSites(t *testing.T) {
	t.Parallel()
	prefix := "import 'node:process';\nfunction scope():void {\n let activeSession: import('inspector').Session | 'stopping' | undefined;\n let profilePath='';\n"
	source := prefix + strings.Repeat("\n", 1644-strings.Count(prefix, "\n")) + inspectorEnableCPUProfiler + "}\n"
	program, err := load.Load([]string{writeAgreementSource(t, "sys.ts", source)})
	if err != nil {
		t.Fatal(err)
	}
	checked, release := program.Checker(t.Context(), program.Files()[0])
	l := &lowering{program: program, checker: checked, result: &ir.Program{}}
	sites := map[string]bool{}
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if !ast.IsExpressionNode(node) || ast.IsPartOfTypeNode(node) {
			return node.ForEachChild(visit)
		}
		if name := l.nodeLibraryMember(node); strings.HasPrefix(name, "node:inspector.Session") {
			site := filepath.Base(program.Where(node))
			sites[site] = true
			if root, _ := l.inspectorDoomedUse(node); root == nil {
				t.Errorf("not doomed: %s %s", site, name)
			}
			if err := l.nodeLibraryRefusal(node); err != nil {
				t.Errorf("meter site stays refused: %v", err)
			}
		}
		return node.ForEachChild(visit)
	}
	program.Files()[0].AsNode().ForEachChild(visit)
	release()
	got := []string{}
	for s := range sites {
		got = append(got, s)
	}
	slices.Sort(got)
	want := []string{"sys.ts:1651:32", "sys.ts:1651:42", "sys.ts:1655:29", "sys.ts:1655:33", "sys.ts:1655:43", "sys.ts:1656:13", "sys.ts:1656:21", "sys.ts:1658:13", "sys.ts:1658:21", "sys.ts:1659:17", "sys.ts:1659:25"}
	slices.Sort(want)
	if !slices.Equal(got, want) {
		t.Fatalf("11 exact doomed meter sites: got %v want %v", got, want)
	}
	_, err = Lower(t.Context(), program)
	if err != nil {
		var stop *NotYet
		if !errors.As(err, &stop) || !strings.Contains(stop.Where, "sys.ts:1646:") {
			t.Fatalf("full enable shape must stop only at captured-let guard: %v", err)
		}
		t.Log(fmt.Sprintf("full guard: %v", err))
	}
	isolated := strings.Replace(source, " let activeSession: import('inspector').Session | 'stopping' | undefined;", "", 1)
	isolated = strings.Replace(isolated, "            if (activeSession) {\n                cb();\n                return false;\n            }", "\n\n\n", 1)
	isolated = strings.Replace(isolated, "                    activeSession = session;", "                    // isolated from the captured let", 1)
	if _, err := lowerSourceAs(t, isolated, ".ts"); err != nil {
		t.Fatalf("all 11 isolated sites must lower: %v", err)
	}
}
