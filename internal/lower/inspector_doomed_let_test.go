package lower

import (
	"errors"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"os"
	"strings"
	"testing"
)

func inspectorLetFixture(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile("../oracle/testdata/inspector_doomed_let.a")
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestInspectorDoomedLetTSCShape(t *testing.T) {
	t.Parallel()
	source := inspectorLetFixture(t)
	l, file := inspectorDoomedAnalysis(t, source)
	for _, name := range []string{"post", "disconnect"} {
		if root, _ := l.inspectorDoomedUse(inspectorDoomedNode(t, file, name, false)); root == nil {
			t.Fatalf("tsc shape %s must be doomed", name)
		}
	}
	if _, err := lowerSourceAs(t, source, ".ts"); err != nil {
		t.Fatal(err)
	}
}

func inspectorLetRefused(t *testing.T, body string) {
	t.Helper()
	source := inspectorDoomedPrelude + "import type {Session} from 'inspector';\n" + body
	l, file := inspectorDoomedAnalysis(t, source)
	if root, _ := l.inspectorDoomedUse(inspectorDoomedNode(t, file, "post", false)); root != nil {
		t.Fatal("unclassified let must stay refused")
	}
	_, err := lowerSourceAs(t, source, ".ts")
	var stop *NotYet
	if !errors.As(err, &stop) {
		t.Fatalf("control must stay NotYet: %v", err)
	}
}

func TestInspectorLetParameterStaysRefused(t *testing.T) {
	t.Parallel()
	inspectorLetRefused(t, "function run(other:Session):void { let active:Session|undefined=session; active=other; if(active) active.post('Profiler.stop'); }")
}

func TestInspectorLetControlsStayRefused(t *testing.T) {
	t.Parallel()
	for name, body := range map[string]string{
		"imported":              "import {Session as Imported} from 'inspector'; function run():void { let active:Session|undefined=session; active=new Imported(); if(active) active.post('Profiler.stop'); }",
		"module":                "let active:Session|undefined=session; if(active) active.post('Profiler.stop');",
		"exported":              "export let active:Session|undefined=session; if(active) active.post('Profiler.stop');",
		"for of":                "function run(sessions:Session[]):void { let active:Session|undefined=session; for(active of sessions) {} if(active) active.post('Profiler.stop'); }",
		"declared for of":       "function run(sessions:(Session|undefined)[]):void { for(let active of sessions) { if(active) active.post('Profiler.stop'); active=session; } }",
		"for in":                "function run(sessions:Record<string,Session>):void { let active:Session|string|undefined=session; for(active in sessions) {} if(active && typeof active!=='string') active.post('Profiler.stop'); }",
		"coalesce":              "function run():void { let active:Session|undefined; active ??= session; if(active) active.post('Profiler.stop'); }",
		"bare reachable":        "function run():void { let active:Session|undefined; if(active) active.post('Profiler.stop'); }",
		"closure parameter":     "function run():void { let active:Session|undefined=session; function write(other:Session):void { active=other; } if(active) active.post('Profiler.stop'); }",
		"generator parameter":   "function run():void { let active:Session|undefined=session; function* write(other:Session):Generator<void> { active=other; yield; } if(active) active.post('Profiler.stop'); }",
		"logical or":            "function run():void { let active:Session|undefined; active ||= session; if(active) active.post('Profiler.stop'); }",
		"logical and":           "function run():void { let active:Session|undefined=session; active &&= session; if(active) active.post('Profiler.stop'); }",
		"destructure":           "function run(other:Session):void { let active:Session|undefined=session; [active]=[other]; if(active) active.post('Profiler.stop'); }",
		"object destructure":    "function run(other:Session):void { let active:Session|undefined=session; ({slot:active}={slot:other}); if(active) active.post('Profiler.stop'); }",
		"shorthand destructure": "function run(other:Session):void { let active:Session|undefined=session; ({active}={active:other}); if(active) active.post('Profiler.stop'); }",
		"array default target":  "function run(other:Session):void { let active:Session|undefined=session; [active=session]=[other]; if(active) active.post('Profiler.stop'); }",
		"object default target": "function run(other:Session):void { let active:Session|undefined=session; ({slot:active=session}={slot:other}); if(active) active.post('Profiler.stop'); }",
	} {
		t.Run(name, func(t *testing.T) { inspectorLetRefused(t, body) })
	}
}

func TestInspectorLetNarrowedUse(t *testing.T) {
	t.Parallel()
	l, file := inspectorDoomedAnalysis(t, inspectorDoomedPrelude+"function run():void { let active:import('inspector').Session|undefined=session; if(active) active.post('Profiler.stop'); }")
	if root, _ := l.inspectorDoomedUse(inspectorDoomedNode(t, file, "post", false)); root == nil {
		t.Fatal("narrowed let must be doomed")
	}
}

func TestInspectorLetOptionalStaysRefused(t *testing.T) {
	t.Parallel()
	inspectorLetRefused(t, "function run():void { let active:Session|undefined=session; active?.post('Profiler.stop'); }")
}

func TestInspectorBareLetDeclaredTypeRequired(t *testing.T) {
	t.Parallel()
	source := inspectorDoomedPrelude + "function run():void { let active:import('inspector').Session; function enable():void { active=session; } function stop():void { active.post('Profiler.stop'); } }"

	l, file := inspectorDoomedAnalysis(t, source)
	if root, _ := l.inspectorDoomedUse(inspectorDoomedNode(t, file, "post", false)); root != nil {
		t.Fatal("bare Session without declared undefined must stay refused")
	}
}

func TestInspectorLetLanguageClosure(t *testing.T) {
	t.Parallel()
	source := `import 'node:process'; import type {Session} from 'inspector';
function run(): string {
 let active: Session;
 const stop = (): string => { active.post('Profiler.stop'); return 'posted'; };
 let answer: string;
 try { answer = stop(); } catch (error) { answer = error instanceof TypeError ? 'caught' : 'other'; }
 if (answer === 'never') {
 const inspector: typeof import('inspector') = require('inspector');
 const session = new inspector.Session();
 active = session;
 }
 return answer;
}
console.log(run());`
	l, file := inspectorDoomedAnalysis(t, source)
	if root, _ := l.inspectorDoomedUse(inspectorDoomedNode(t, file, "post", false)); root != nil {
		t.Fatal("Language closure must stay refused")
	}
	_, err := lowerSourceAs(t, source, ".ts")
	var stop *NotYet
	if !errors.As(err, &stop) {
		t.Fatalf("Language closure must stay NotYet: %v", err)
	}
}

func TestInspectorLetHatchesCheckBeforeTrap(t *testing.T) {
	t.Parallel()
	for name, hatch := range map[string]string{"nonnull": "activeSession!", "cast": "(activeSession as Session)"} {
		t.Run(name, func(t *testing.T) {
			source := inspectorLetFixture(t)
			source = strings.ReplaceAll(source, ` | "stopping"`, "")
			source = strings.ReplaceAll(source, ` && activeSession !== "stopping"`, "")
			source = strings.ReplaceAll(source, `activeSession = "stopping";`, "")
			source = strings.Replace(source, "return disable();", hatch+".post('Profiler.stop'); return 'unreachable';", 1)
			p, err := lowerSourceAs(t, source, ".ts")
			if err != nil {
				if name == "cast" && strings.Contains(err.Error(), "cast the runtime can't check") {
					t.Log(err)
					return
				}
				t.Fatal(err)
			}
			found := false
			for _, message := range p.Strings {
				if strings.Contains(message, "non-null") || strings.Contains(message, "cast") {
					found = true
					t.Log(message)
				}
			}
			if !found {
				t.Fatal("hatch must retain its own check before trap")
			}
		})
	}
}

// A non-null callee does not narrow the Session receiver inside its optional access.
func TestInspectorLetWrappedOptionalCalleeStaysRefused(t *testing.T) {
	t.Parallel()
	l, file := inspectorDoomedAnalysis(t, inspectorDoomedPrelude+"function run():void { let active:import('inspector').Session|undefined=session; function stop():void { (active?.post)!('Profiler.stop'); } }")
	call := inspectorDoomedNode(t, file, "post", false)
	for call.Kind != ast.KindCallExpression {
		call = call.Parent
	}
	if root, _ := l.inspectorDoomedUse(call); root != nil {
		t.Fatal("wrapped optional callee must not hide the receiver flow type")
	}
}

func TestInspectorLetParenthesizedCalleeKeepsReceiverFlow(t *testing.T) {
	t.Parallel()
	source := inspectorDoomedPrelude + `function run():void {
 let active:import('inspector').Session|undefined=session;
 function clear():void { active=undefined; }
 const stop=():void=>{
  // @ts-ignore: expose the receiver's actual union type to the admission predicate.
  (active.post)('Profiler.stop');
 };
}`
	l, file := inspectorDoomedAnalysis(t, source)
	member := inspectorDoomedNode(t, file, "post", false)
	if !l.includesUndefined(l.checker.GetTypeAtLocation(member.Expression())) {
		t.Fatal("pin requires an unnarrowed receiver")
	}
	call := member
	for call.Kind != ast.KindCallExpression {
		call = call.Parent
	}
	if l.nodeLibraryMember(call) != "node:inspector.Session.post" {
		t.Fatal("pin requires an inspector call")
	}
	if root, _ := l.inspectorDoomedUse(call); root != nil {
		t.Fatal("parenthesized callee must retain the receiver flow gate")
	}
}

func TestInspectorLetDerivedWriteStaysRefused(t *testing.T) {
	t.Parallel()
	inspectorLetRefused(t, `function run():void {
 let other:Session|"stopping"|undefined;
 let active:Session|string[]|undefined;
 function enable():void { other=session; }
 function stop():void { other="stopping"; }
 function write():void { if(other === "stopping") active=other.split(''); }
 function post():void { if(active && !Array.isArray(active)) active.post('Profiler.stop'); }
 }`)
}
