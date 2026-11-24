package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A presence-only read of a field through a checked view answers what Node answers (Adamic's ruling on #y5n0x0j,
// lower's viewFieldPresence): view_field_presence.ts holds tsbuildPublic.ts:1371's guard on a Date and on its own
// member, and every presence form on a value of the cast's member, another plain member, a member holding the
// field as another type or as undefined, and a Date. Both backends print Node's lines, and the counted build
// frees all it made (the fixture's registration below). On main f9acabd0 the guard's cast was NotYet ("a cast that
// makes a viewed readonly array writable"), and a guard with no such field stopped a Date at the read with exit 70
// where Node answered false (J-n1ak9k5, t6). Mutants (view_field_presence.mutants.md): native's Date keeping its
// time under _fsFileTime again answers the last line's Date as present; SlotDefined emitted as truthiness answers
// the empty string and 0 as undefined.
func TestViewFieldPresenceAnswersNode(t *testing.T) {
	t.Parallel()
	const want = "a true\nb false\n" +
		"true false yes 1 false true and\n" +
		"false true no 0 false true not\n" +
		"false true no 0 false true not\n" +
		"true false yes 1 false true and\n" +
		"false true no 0 true false not\n" +
		"false true no 0 true false not\n" +
		"false true no 0 true false not\n" +
		"true true | false false\n"
	program, path := castFixture(t, "view_field_presence.ts")
	if node := onNode(t, path); node.exitCode != 0 || string(node.stdout) != want {
		t.Fatalf("source on Node: exit %d, stdout %q; the fixture's lines are %q", node.exitCode, node.stdout, want)
	}
	native, _ := natively(t, program)
	javaScript := onJavaScriptBackend(t, program)
	for backend, result := range map[string]run{"native": native, "JavaScript": javaScript} {
		if result.exitCode != 0 || string(result.stdout) != want {
			t.Errorf("%s: exit %d, stdout %q, stderr %q; want Node's %q", backend, result.exitCode, result.stdout, result.stderr, want)
		}
	}
}

// A read whose value is used keeps the view's read and its stop, at the read, on both backends alike, while Node
// runs on: a value read, typeof, the left of ||, the right of an && whose value is read, and a name Date's own type
// declares (getDay), which Node answers from Date's prototype. Mutant: viewFieldPresenceOperand without its
// truthinessOnly test admits typeof, which then lowers and agrees with Node, and the typeof row goes red.
func TestViewFieldPresenceKeepsTheViewsRead(t *testing.T) {
	t.Parallel()
	const entry = "interface Entry { readonly watcher: string }\n"
	for _, probe := range []struct{ name, source, node string }{
		{"value", entry + "function read(value: Entry | Date): string {\n\tconst watcher = (value as Entry).watcher;\n\treturn watcher;\n}\nconsole.log('reading');\nconsole.log(read(new Date(5)));\n", "reading\nundefined\n"},
		{"typeof", entry + "function kind(value: Entry | Date): string {\n\treturn typeof (value as Entry).watcher;\n}\nconsole.log('reading');\nconsole.log(kind(new Date(5)));\n", "reading\nundefined\n"},
		{"or", entry + "function or(value: Entry | Date): string {\n\treturn (value as Entry).watcher || 'none';\n}\nconsole.log('reading');\nconsole.log(or(new Date(5)));\n", "reading\nnone\n"},
		{"and_value", entry + "function and(value: Entry | Date, flag: boolean): string {\n\tconst both = flag && (value as Entry).watcher;\n\treturn String(both);\n}\nconsole.log('reading');\nconsole.log(and(new Date(5), true));\n", "reading\nundefined\n"},
		{"date_member", "interface Stamped { readonly getDay: number; readonly at: string }\nfunction stamped(value: Stamped | Date): boolean {\n\treturn !!(value as Stamped).getDay;\n}\nconsole.log('reading');\nconsole.log(String(stamped(new Date(5))));\n", "reading\ntrue\n"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), probe.name+".ts")
			if err := os.WriteFile(path, []byte(probe.source), 0o644); err != nil {
				t.Fatal(err)
			}
			if node := onNode(t, path); node.exitCode != 0 || string(node.stdout) != probe.node {
				t.Fatalf("source on Node must run on: exit %d, stdout %q", node.exitCode, node.stdout)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatalf("lowering: %v; the view's read lowers today", err)
			}
			native, _ := natively(t, program)
			javaScript := onJavaScriptBackend(t, program)
			for backend, result := range map[string]run{"native": native, "JavaScript": javaScript} {
				if result.exitCode != 70 || string(result.stdout) != "reading\n" || !strings.HasPrefix(string(result.stderr), "adamic: panic: field read failed:") {
					t.Errorf("%s: exit %d, stdout %q, stderr %q; want 70 at the view's read", backend, result.exitCode, result.stdout, result.stderr)
				}
			}
			if string(native.stderr) != string(javaScript.stderr) {
				t.Errorf("the backends stop differently: native %q, JavaScript %q", native.stderr, javaScript.stderr)
			}
		})
	}
}

// A presence read Node answers through a prototype the value's own slots can't answer keeps today's compile-time
// stop: Language's e01 to e05 (#y5n0x0j, 19:15Z: Object.prototype's toString on a Date, a class's method, Map's
// size, a function's name, an array's length), a name a getter in the program declares, and == undefined, which
// null answers too. Each stop's words are main f9acabd0's (J-qwqtrpq), unchanged. Mutant: presenceAnsweredOwn
// answering true admits e01, which lowers, and its row goes red.
func TestViewFieldPresenceKeepsItsCompileStops(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, node, stop string }{
		{"e01", "interface Watcher { readonly watcher: string; readonly toString: () => string }\nfunction hasStringer(value: Date | Watcher): boolean {\n\treturn !!(value as Watcher).toString;\n}\nconsole.log(String(hasStringer(new Date(0))));\n",
			"true\n", "a callable field toString read through a checked view"},
		{"e02", "class Polled {\n\tpoll(): number {\n\t\treturn 1;\n\t}\n}\ninterface Watcher { readonly watcher: string; readonly poll: () => number }\nfunction isPolled(value: Polled | { readonly watcher: string }): boolean {\n\treturn (value as Watcher).poll !== undefined;\n}\nconsole.log(String(isPolled(new Polled())) + \" \" + String(isPolled({ watcher: \"w\" })));\n",
			"true false\n", "a checked view of Watcher over the union Polled"},
		{"e03", "interface Sized { readonly size: number }\nfunction sized(value: Map<string, number> | { readonly label: string }): boolean {\n\treturn (value as Sized).size !== undefined;\n}\nconsole.log(String(sized(new Map())) + \" \" + String(sized({ label: \"l\" })));\n",
			"true false\n", "a Map or Set going where a plain object type is held"},
		{"e04", "interface Named { readonly name: string }\nfunction named(value: (() => number) | { readonly label: string }): boolean {\n\treturn !!(value as Named).name;\n}\nfunction counter(): number {\n\treturn 1;\n}\nconsole.log(String(named(counter)) + \" \" + String(named({ label: \"l\" })));\n",
			"true false\n", "a checked view of Named over a union held as other than an object"},
		{"e05", "interface Long { readonly length: number }\nfunction long(value: readonly string[] | { readonly label: string }): boolean {\n\treturn (value as Long).length === undefined;\n}\nconsole.log(String(long([])) + \" \" + String(long({ label: \"l\" })));\n",
			"false true\n", "a checked view of Long over a union held as other than an object"},
		{"accessor", "class Box {\n\tget watcher(): string {\n\t\treturn 'box';\n\t}\n}\ninterface Entry { readonly watcher: string }\nfunction has(value: Entry | Date): boolean {\n\treturn !!(value as Entry).watcher;\n}\nconsole.log(new Box().watcher);\nconsole.log(String(has(new Date(5))));\n",
			"box\nfalse\n", "a getter in a checked field contract"},
		{"loose", "interface Entry { readonly watcher: string }\nfunction loose(value: Entry | Date): boolean {\n\treturn (value as Entry).watcher == undefined;\n}\nconsole.log(String(loose(new Date(5))));\n",
			"true\n", "refuses ==; use ==="},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), probe.name+".ts")
			if err := os.WriteFile(path, []byte(probe.source), 0o644); err != nil {
				t.Fatal(err)
			}
			if node := onNode(t, path); node.exitCode != 0 || string(node.stdout) != probe.node {
				t.Fatalf("source on Node: exit %d, stdout %q, want %q", node.exitCode, node.stdout, probe.node)
			}
			if _, err := lowered(t, path); err == nil || !strings.Contains(err.Error(), probe.stop) {
				t.Fatalf("want today's stop %q, got %v", probe.stop, err)
			}
		})
	}
}

func init() {
	// The fixture agrees with Node on every backend, the counted build freeing all it made.
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/view_field_presence.ts", true, false})
}
