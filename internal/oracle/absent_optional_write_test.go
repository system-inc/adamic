package oracle

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// A literal that leaves an optional field out has no slot for it, so a later write to that field is NotYet
// at compile time ("writing a possibly absent optional own field"), never the runtime's "a field the checker
// proved is there is missing". On main f245bfe4 the guard saw only a literal whose context was an interface:
// with type Branded = Mut & { brand: undefined }, v.name = "w" panicked natively where Node and the
// JavaScript backend print w (Language's m3 on #krv1hgh), and so did the write through const mv: Mut = v (m1).
// The guard now matches the field by its declarations, whatever spelled the type (Adamic's ruling), so every
// road below stops at its write: an interface (m2, as before), an intersection, an upcast view of one, an
// alias of either, a generic constrained to the interface, and a compound and a logical assignment. When two
// members both declare the field (Mut & { name?: string; ... }), the intersection's property is a symbol of
// its own, and only its declarations tie it to Mut's: an upcast and a generic write through Mut's. The controls lower
// and agree with Node: a read of the absent field (m4), and a write to a field the program's only literal gave.
// Mutants: noteOmittedOptionalsOf skipping intersections again turns the intersection rows red (m3 panics
// natively); fieldDeclarations giving the property's own symbol alone turns the two shared rows red.
func TestAbsentOptionalWriteIsNotYetWhateverSpellsTheType(test *testing.T) {
	test.Parallel()
	const mut = "interface Mut { name?: string; }\n"
	const branded = "type Branded = Mut & { brand: undefined };\n"
	for _, probe := range []struct{ name, source string }{
		{"interface", mut +
			"interface Branded extends Mut { brand: undefined; }\nconst v: Branded = { brand: undefined };\nv.name = 'w';\nconsole.log(String(v.name));\n"},
		{
			"intersection",
			mut + branded + "const v: Branded = { brand: undefined };\nv.name = 'w';\nconsole.log(String(v.name));\n",
		},
		{"intersection_upcast", mut +
			branded +
			"const v: Branded = { brand: undefined };\nconst mv: Mut = v;\nmv.name = 'w';\nconsole.log(String(mv.name));\n"},
		{
			"alias_of_interface",
			mut + "type Alias = Mut;\nconst v: Alias = {};\nv.name = 'w';\nconsole.log(String(v.name));\n",
		},
		{"alias_of_intersection", mut +
			branded +
			"type Again = Branded;\nconst v: Again = { brand: undefined };\nv.name = 'w';\nconsole.log(String(v.name));\n"},
		{"shared_upcast", mut +
			"type Both = Mut & { name?: string; brand: undefined };\nconst v: Both = { brand: undefined };\nconst mv: Mut = v;\nmv.name = 'w';\nconsole.log(String(mv.name));\n"},
		{"shared_generic", mut +
			"type Both = Mut & { name?: string; brand: undefined };\nfunction put<T extends Mut>(x: T): void {\n\tx.name = 'w';\n}\nconst v: Both = { brand: undefined };\nput(v);\nconsole.log(String(v.name));\n"},
		{"generic", mut +
			branded +
			"function put<T extends Mut>(x: T): void {\n\tx.name = 'w';\n}\nconst v: Branded = { brand: undefined };\nput(v);\nconsole.log(String(v.name));\n"},
		{
			"compound",
			mut + branded + "const v: Branded = { brand: undefined };\nv.name += 'w';\nconsole.log(String(v.name));\n",
		},
		{
			"logical",
			mut + branded + "const v: Branded = { brand: undefined };\nv.name ??= 'w';\nconsole.log(String(v.name));\n",
		},
	} {
		for _, extension := range []string{".a", ".ts"} {
			test.Run(probe.name+extension, func(test *testing.T) {
				test.Parallel()
				path := filepath.Join(test.TempDir(), probe.name+extension)
				if err := os.WriteFile(path, []byte(probe.source), 0o644); err != nil {
					test.Fatal(err)
				}
				truth := onNode(test, path)
				if truth.exitCode != 0 || !strings.HasSuffix(string(truth.stdout), "w\n") {
					test.Fatalf(
						"source Node: exit %d, stdout %q, stderr %q",
						truth.exitCode,
						truth.stdout,
						truth.stderr,
					)
				}
				_, err := lowered(test, path)
				var notYet *lower.NotYet
				if !errors.As(err, &notYet) ||
					!strings.Contains(err.Error(), "writing a possibly absent optional own field") {
					test.Fatalf("want the write NotYet as a possibly absent optional own field, got %v", err)
				}
			})
		}
	}
	for _, control := range []struct{ name, source, node string }{
		{
			"read_absent",
			mut + "const v: Mut & { brand: undefined } = { brand: undefined };\nconsole.log(String(v.name));\n",
			"undefined\n",
		},
		{
			"given",
			mut +
				branded +
				"const v: Branded = { brand: undefined, name: 'a' };\nv.name = 'w';\nconsole.log(String(v.name));\n",
			"w\n",
		},
	} {
		for _, extension := range []string{".a", ".ts"} {
			test.Run(control.name+extension, func(test *testing.T) {
				test.Parallel()
				path := filepath.Join(test.TempDir(), control.name+extension)
				if err := os.WriteFile(path, []byte(control.source), 0o644); err != nil {
					test.Fatal(err)
				}
				truth := onNode(test, path)
				if disagreement(run{stdout: []byte(control.node)}, truth) != "" {
					test.Fatalf(
						"source Node: exit %d, stdout %q, stderr %q",
						truth.exitCode,
						truth.stdout,
						truth.stderr,
					)
				}
				program, err := lowered(test, path)
				if err != nil {
					test.Fatalf("want it lowered, got %v", err)
				}
				sanitized, _ := nativelyUncached(test, program)
				for backend, got := range map[string]run{
					"sanitized":  sanitized,
					"release":    releasedUncached(test, program),
					"javascript": onJavaScriptBackend(test, program),
				} {
					if difference := disagreement(truth, got); difference != "" {
						test.Errorf("%s: %s, stderr %q", backend, difference, got.stderr)
					}
				}
			})
		}
	}
}
