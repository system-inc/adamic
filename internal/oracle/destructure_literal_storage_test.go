package oracle

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// Two lets that hold either a number or a string, narrowed apart by what they were given.
const literalStorageLets = "let first: number | string = 1;\nlet second: number | string = 'two';\n"

// literalStorageProgram writes source as a program of its own and returns its path.
func literalStorageProgram(t *testing.T, name, source string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name+".a")
	if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// A literal written into an array pattern, [first, second] = [second, first], is stored as the
// pattern's slots, number | string here, while the checker narrows each element by what's written in
// it, [string, number]. Read by the checker's tuple, the first slot's box was read as a number and
// native printed two 5.4268827562133e-310 where Node prints two 1 (#fsb6n82). Each program is held to
// Node natively, sanitized and on the JavaScript backend. A tuple local is stored as its own type and
// takes the checker's path, its number | string elements read as destructureFrom reads them (before,
// every union element stopped). A parameter destructured from a literal argument is destructureFrom's,
// which reads the parameter's declared tuple, the literal's storage, and is the control.
//
// The other arm reads a value from where it's held at the type the checker gives it there, and for a
// local or a field that's the flow-narrowed type, not the declared one it's stored as. A local or a
// field declared as a union of tuples and narrowed by an assignment is the same shape one hop later
// (Compiler's question on #fsb6n82), and so is a union whose member has a union element. Each agreed
// with Node on main before this fix (or stopped), and each is held here: union_local and union_field
// read a narrowed [string, number] out of a declared [number, string] | [string, number], and the
// single_narrowed and union_member_wider programs read a [number | string, number] the fix now lets
// through, assigned, held in a local or held in a field.
func TestDestructureLiteralStorageAgrees(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source string }{
		{"union_swap", literalStorageLets + "[first, second] = [second, first];\nconsole.log(`${first} ${second}`);\n"},
		{"number_swap", "let first = 1;\nlet second = 2;\n[first, second] = [second, first];\nconsole.log(`${first} ${second}`);\n"},
		{"rotation", literalStorageLets + "let third: number | string = 3;\n[first, second, third] = [third, first, second];\nconsole.log(`${first} ${second} ${third}`);\n"},
		{"tuple_local", literalStorageLets + "const tuple: [number | string, number | string] = [second, first];\n[first, second] = tuple;\nconsole.log(`${first} ${second}`);\n"},
		{"parameter", literalStorageLets + "function show([a, b]: [number | string, number | string]): string {\n\treturn `${a} ${b}`;\n}\nconsole.log(show([second, first]));\n"},
		{"union_local", literalStorageLets + "let tuple: [number, string] | [string, number] = ['a', 1];\n[first, second] = tuple;\nconsole.log(`${first} ${second}`);\n"},
		{"union_field", literalStorageLets + "const holder: { tuple: [number, string] | [string, number] } = { tuple: [1, 'x'] };\nholder.tuple = ['a', 1];\n[first, second] = holder.tuple;\nconsole.log(`${first} ${second}`);\n"},
		{"single_narrowed", literalStorageLets + "let tuple: [number | string, number] = [1, 2];\ntuple = ['a', 2];\n[first, second] = tuple;\nconsole.log(`${first} ${second}`);\n"},
		{"union_member_wider_local", literalStorageLets + "let tuple: [number | string, number] | [boolean, boolean] = ['a', 1];\n[first, second] = tuple;\nconsole.log(`${first} ${second}`);\n"},
		{"union_member_wider_assigned", literalStorageLets + "let tuple: [number | string, number] | [boolean, boolean] = [true, false];\ntuple = ['a', 1];\n[first, second] = tuple;\nconsole.log(`${first} ${second}`);\n"},
		{"union_member_wider_field", literalStorageLets + "const holder: { tuple: [number | string, number] | [boolean, boolean] } = { tuple: [true, false] };\nholder.tuple = ['a', 1];\n[first, second] = holder.tuple;\nconsole.log(`${first} ${second}`);\n"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			path := literalStorageProgram(t, probe.name, probe.source)
			want := onNode(t, path)
			if want.exitCode != 0 {
				t.Fatalf("Node: %#v", want)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			sanitized, _ := nativelyUncached(t, program)
			for backend, got := range map[string]run{"native": releasedUncached(t, program), "sanitized": sanitized, "javascript": onJavaScriptBackend(t, program)} {
				if diff := disagreement(want, got); diff != "" {
					t.Errorf("%s: %s, stdout %q, stderr %q", backend, diff, got.stdout, got.stderr)
				}
			}
		})
	}
}

// What the judgment can't read yet stops, in its own words, never prints a box as a number: a union
// slot holding an object, which a destructuring read takes only when every member is one a template
// writes, as destructureFrom does; a nested pattern, whose inner literal is refused as a mutable
// [string] seen as [string | number]; a defaulted slot and a rest element, whose literals stop first;
// the object twin, which assignment doesn't lower; and a declaration, whose names are held as the
// pattern's slots while destructureFrom reads the checker's narrowed literal, so the two disagree and
// it stops.
func TestDestructureLiteralStorageStops(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, what string }{
		{"box_swap", "interface Box {\n\tlabel: string;\n}\nfunction show(value: string | Box): string {\n\treturn typeof value === 'string' ? value : value.label;\n}\nlet first: string | Box = { label: 'boxed' };\nlet second: string | Box = 'plain';\n[first, second] = [second, first];\nconsole.log(`${show(first)} ${show(second)}`);\n", "a tuple element of type string | Box"},
		{"nested", literalStorageLets + "[[first], second] = [[second], first];\nconsole.log(`${first} ${second}`);\n", "a value of type [string] seen as [string | number], which can write string | number where string is read"},
		{"defaulted", literalStorageLets + "[first = 0, second] = [second, first];\nconsole.log(`${first} ${second}`);\n", "a tuple element of another type than its tuple's"},
		{"rest", literalStorageLets + "let third: number | string = 3;\nlet rest: (number | string)[] = [];\n[first, ...rest] = [second, third, first];\nconsole.log(`${first} ${rest.join(',')}`);\n", "a tuple literal with more values than its tuple has elements"},
		{"object_twin", literalStorageLets + "({ x: first, y: second } = { x: second, y: first });\nconsole.log(`${first} ${second}`);\n", "assigning to an ObjectLiteralExpression"},
		{"declaration", literalStorageLets + "const [p, q]: [number | string, number | string] = [second, first];\nconsole.log(`${p} ${q}`);\n", "a destructured name held otherwise than its field"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			path := literalStorageProgram(t, probe.name, probe.source)
			if node := onNode(t, path); node.exitCode != 0 {
				t.Fatalf("Node: %#v", node)
			}
			_, err := lowered(t, path)
			var pending *lower.NotYet
			var refused *lower.Refused
			switch {
			case errors.As(err, &pending) && pending.What == probe.what:
			case errors.As(err, &refused) && strings.HasPrefix(refused.What, probe.what):
			default:
				t.Fatalf("want a stop %q, got %v", probe.what, err)
			}
		})
	}
}
