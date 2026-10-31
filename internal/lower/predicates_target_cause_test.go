package lower

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
)

// A checked predicate whose target is a union checks the disjunction of its members' checks (#c7w6j0y,
// rule 1). The honest call agrees with Node whichever member the value is, and a lie whose value is
// neither member stops at the call on sanitized native and the JavaScript backend. Mutant: the first
// member's check alone, which stops the honest StringLiteral.
const unionTargetPrelude = `interface Node { readonly kind: number; }
interface Identifier extends Node { readonly kind: 80; readonly escapedText: string; }
interface StringLiteral extends Node { readonly kind: 11; readonly text: string; }
function tagOf(node: Node): number { return node.kind; }
function show(node: Node): void { if (isNameLike(node)) { console.log(String(node.kind)); } else { console.log("no"); } }
`

func TestCheckedPredicateUnionTargetChecksEveryMember(t *testing.T) {
	t.Parallel()
	t.Run("honest", func(t *testing.T) {
		t.Parallel()
		lowersAndAgreesWithNodeAs(t, unionTargetPrelude+`function isNameLike(node: Node): node is Identifier | StringLiteral { return tagOf(node) === 80 || tagOf(node) === 11; }
show({ kind: 11, text: "s" } as Node);
show({ kind: 80, escapedText: "i" } as Node);
show({ kind: 7 } as Node);`, ".ts")
	})
	t.Run("lie", func(t *testing.T) {
		t.Parallel()
		program, err := lowerTypeScriptAssertionSource(t, unionTargetPrelude+`function isNameLike(node: Node): node is Identifier | StringLiteral { return tagOf(node) !== 0; }
show({ kind: 7 } as Node);`)
		if err != nil {
			t.Fatal(err)
		}
		directory := t.TempDir()
		binary := filepath.Join(directory, "native")
		if err := native.Build(native.C(program), binary, native.Options{Sanitize: true}); err != nil {
			t.Fatal(err)
		}
		runner, err := filepath.Abs("../../oracle/node.mjs")
		if err != nil {
			t.Fatal(err)
		}
		generated := filepath.Join(directory, "generated.mjs")
		if err := os.WriteFile(generated, []byte(javascript.JavaScript(program)), 0o644); err != nil {
			t.Fatal(err)
		}
		for _, command := range []*exec.Cmd{exec.Command(binary), exec.Command("node", "--disable-warning=ExperimentalWarning", runner, generated)} {
			command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1")
			var out, errs bytes.Buffer
			command.Stdout, command.Stderr = &out, &errs
			code := 0
			if err := command.Run(); err != nil {
				var exit *exec.ExitError
				if !errors.As(err, &exit) {
					t.Fatal(err)
				}
				code = exit.ExitCode()
			}
			if code != 70 || out.Len() != 0 || !strings.HasPrefix(errs.String(), "adamic: panic: predicate isNameLike at ") || !strings.HasSuffix(errs.String(), " true branch: source Node, target Identifier | StringLiteral\n") {
				t.Fatalf("%s: exit %d stdout %q stderr %q; want 70 and the predicate's located message", command.Path, code, out.String(), errs.String())
			}
		}
	})
}

// An untagged target no cheap check can establish is NotYet under its own cause, which the meter files
// with the casts the runtime can't check, group 10 (#c7w6j0y, Adamic's ruling): a base with no required
// field, as CompilerOptions has only optional fields and an index signature, and a required field holding
// untagged records, as the build infos do. An untagged target of primitive fields is rule 3's, a
// structural check not built yet, and keeps the old reason, as does a branded supertype, which the tag-set
// rule takes. Mutant: the cause dropped, which puts the first two back under the old reason.
func TestCheckedPredicateTargetCause(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, types, target, want string }{
		{"options", `interface Options { outFile?: string; strict?: boolean; [option: string]: string | boolean | string[] | undefined; }`, `Options & { readonly types: string[] }`, "a checked predicate target whose base has no required field to check (Options & { readonly types: string[]; })"},
		{"build info", `interface Diagnostic { readonly file: string; readonly start: number; } interface Info { readonly version: string; readonly diagnostics: readonly Diagnostic[]; }`, `Info`, "a checked predicate target whose check would read past the depth bound (Info)"},
		{"primitive fields", `interface Referenced { readonly file: string; readonly index: number; }`, `Referenced`, "checked predicate overload target Referenced"},
		{"string array", `interface Plain { readonly version: string; readonly root: readonly string[]; readonly errors: true | undefined; }`, `Plain`, "checked predicate overload target Plain"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			source := probe.types + "\nfunction isTarget(value: object): value is " + probe.target + " { return value !== null; }\nfunction show(value: object): void { if (isTarget(value)) { console.log(\"yes\"); } else { console.log(\"no\"); } }\nshow({ version: \"1\" });"
			_, err := lowerTypeScriptAssertionSource(t, source)
			var pending *NotYet
			if !errors.As(err, &pending) || !strings.Contains(err.Error(), probe.want) {
				t.Fatalf("want NotYet with %q, got %v", probe.want, err)
			}
		})
	}
}
