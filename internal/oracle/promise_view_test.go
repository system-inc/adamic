package oracle

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// An async promise is held as an adamic_async_promise with no object layout, but tsc lets a Promise<T> go where
// a structural PromiseLike<T> is held. On main 1c1d6568 that lowered, and clang refused the C that put an
// adamic_async_promise* where an adamic_object* goes, as an initializer and as an argument (#v5b6pzp,
// J-q0gnv3p). Ruled by Compiler: NotYet at the view, keyed on representation (shape_fit.go's
// promiseWhereObjectGoes), naming the fix. Each probe pins the whole stop, .a and .ts. The controls lower and
// agree with Node on sanitized, release and JavaScript: a promise kept typed Promise<T>, one going where a union
// holding PromiseLike<T> is held (boxed, so typeof reads object), an object literal's PromiseLike<T> field (the
// literal is made with its own field type), all three measured agreeing natively on main first, and an await's
// operand, which tsc gives a PromiseLike<T> slot but the await reads as a promise (async_plain.a's shape). The
// neighbours stop where they stopped on main, in their own words and never in this one's: unknown, a type that names no promise (promise_into_void.go), and
// Promise.all and Promise.resolve of a promise, the unproved promise surface and adoption (async.go). Mutant:
// promiseWhereObjectGoes answering "", and both probes lower again, to clang's error natively.
func TestPromiseIsNotYetWhereAnObjectIsHeld(t *testing.T) {
	t.Parallel()
	const stop = "a Promise going where a plain object type is held (keep it typed Promise<T>, or await it)"
	for _, probe := range []struct{ name, source, node, where string }{
		{"initializer", "const viewed: PromiseLike<number> = Promise.resolve(7);\nconsole.log(typeof viewed);\n", "object\n", ":1:37"},
		{"argument", "function show(value: PromiseLike<number>): string {\n\treturn typeof value;\n}\nconsole.log(show(Promise.resolve(7)));\n", "object\n", ":4:18"},
	} {
		for _, extension := range []string{".a", ".ts"} {
			t.Run(probe.name+extension, func(t *testing.T) {
				t.Parallel()
				path := filepath.Join(t.TempDir(), probe.name+extension)
				if err := os.WriteFile(path, []byte(probe.source), 0o644); err != nil {
					t.Fatal(err)
				}
				truth := onNode(t, path)
				if disagreement(run{stdout: []byte(probe.node)}, truth) != "" {
					t.Fatalf("source Node: exit %d, stdout %q, stderr %q", truth.exitCode, truth.stdout, truth.stderr)
				}
				_, err := lowered(t, path)
				var notYet *lower.NotYet
				want := path + probe.where + ": stage 0 can't lower " + stop + " yet"
				if !errors.As(err, &notYet) || err.Error() != want {
					t.Fatalf("want NotYet %q, got %v", want, err)
				}
			})
		}
	}
	for _, control := range []struct{ name, source, node string }{
		{"typed_promise", "const kept: Promise<number> = Promise.resolve(7);\nconsole.log(typeof kept);\n", "object\n"},
		{"union_slot", "const u: number | PromiseLike<number> = Promise.resolve(7);\nconsole.log(typeof u);\n", "object\n"},
		{"literal_field", "const n: { p: PromiseLike<number> } = { p: Promise.resolve(7) };\nconsole.log(typeof n.p);\n", "object\n"},
		{"awaited", "async function seven(): Promise<number> {\n\treturn 7;\n}\nasync function forwarding(): Promise<number> {\n\treturn await seven();\n}\nconsole.log(String(await (forwarding())));\n", "7\n"},
	} {
		for _, extension := range []string{".a", ".ts"} {
			t.Run(control.name+extension, func(t *testing.T) {
				t.Parallel()
				path := filepath.Join(t.TempDir(), control.name+extension)
				if err := os.WriteFile(path, []byte(control.source), 0o644); err != nil {
					t.Fatal(err)
				}
				truth := onNode(t, path)
				if disagreement(run{stdout: []byte(control.node)}, truth) != "" {
					t.Fatalf("source Node: exit %d, stdout %q, stderr %q", truth.exitCode, truth.stdout, truth.stderr)
				}
				program, err := lowered(t, path)
				if err != nil {
					t.Fatalf("want it lowered, got %v", err)
				}
				sanitized, _ := nativelyUncached(t, program)
				for backend, got := range map[string]run{"sanitized": sanitized, "release": releasedUncached(t, program), "javascript": onJavaScriptBackend(t, program)} {
					if difference := disagreement(truth, got); difference != "" {
						t.Errorf("%s: %s, stderr %q", backend, difference, got.stderr)
					}
				}
			})
		}
	}
	for _, neighbour := range []struct{ name, source string }{
		{"unknown_slot", "const v: unknown = Promise.resolve(7);\nconsole.log(typeof v);\n"},
		{"promise_all", "async function main(): Promise<void> {\n\tconst values = await Promise.all([Promise.resolve(1), Promise.resolve(2)]);\n\tconsole.log(String(values[0] + values[1]));\n}\nvoid main();\n"},
		{"resolve_adopts", "async function main(): Promise<void> {\n\tconst inner = Promise.resolve(3);\n\tconsole.log(String(await Promise.resolve(inner)));\n}\nvoid main();\n"},
	} {
		for _, extension := range []string{".a", ".ts"} {
			t.Run(neighbour.name+extension, func(t *testing.T) {
				t.Parallel()
				path := filepath.Join(t.TempDir(), neighbour.name+extension)
				if err := os.WriteFile(path, []byte(neighbour.source), 0o644); err != nil {
					t.Fatal(err)
				}
				_, err := lowered(t, path)
				if err == nil || strings.Contains(err.Error(), stop) {
					t.Fatalf("want it stopped in its own words, as on main, got %v", err)
				}
				t.Logf("%s stops as on main: %v", neighbour.name+extension, err)
			})
		}
	}
}
