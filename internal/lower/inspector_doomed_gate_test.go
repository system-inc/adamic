package lower

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/system-inc/adamic/internal/load"
)

func TestInspectorDoomedAbsentGateSkipsMemberLookups(t *testing.T) {
	t.Parallel()
	l, file := inspectorDoomedAnalysis(t, `const object={value:1}; function lookup(name:string):number{return 1;} lookup("inspector"); object.value;`)
	member := inspectorDoomedNode(t, file, "value", false)
	if l.inspectorProgramHasRequire() || !l.inspectorRequireKnown {
		t.Fatal("an unrelated call with an inspector literal must not open the gate")
	}
	// The false fact is memoized. Neither entry may use the program or checker
	// again: a member lookup past the gate would dereference the nil checker.
	l.program, l.checker = nil, nil
	if root, require := l.inspectorRecordBuiltUse(member); root != nil || require != nil {
		t.Fatal("the absent gate reached the recorder")
	}
	if value, handled, err := l.inspectorDoomedValue(member); value != nil || handled || err != nil {
		t.Fatalf("the absent gate reached the trap builder: %v %v %v", value, handled, err)
	}
	if len(l.inspectorSkippedContracts) != 0 {
		t.Fatal("the absent gate recorded a member")
	}
}

func TestInspectorDoomedGateSeesImportedModule(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	entry := filepath.Join(dir, "main.ts")
	for name, source := range map[string]string{
		"main.ts":   `import { run } from './second.js'; run();`,
		"second.ts": `import 'node:process'; export function run():void { const inspector:typeof import('inspector')=require('inspector'); const session=new inspector.Session(); session.post('Profiler.enable'); }`,
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(source), 0600); err != nil {
			t.Fatal(err)
		}
	}
	program, err := load.Load([]string{entry})
	if err != nil {
		t.Fatal(err)
	}
	if len(program.Files()) != 1 {
		t.Fatal("the second module must not be a named root")
	}
	p, err := Lower(t.Context(), program)
	if err != nil {
		t.Fatalf("inspector require in a second module: %v", err)
	}
	const message = `node:inspector.Session.post was reached through session, whose require("inspector") throws on every host Adamic targets`
	if !slices.Contains(p.Strings, message) {
		t.Fatal("the graph gate lost the imported module's doomed-use Panic")
	}
}
