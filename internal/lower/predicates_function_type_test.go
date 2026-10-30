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

// A claim on a function type is checked at every call through it in .ts (#1w5bve2, agreed Oct 30):
// a parameter's, an interface method's, and an overload's whose implementation checks its own. The
// lies are asserts claims filled by a function returning void, which tsc relates to an asserts
// signature, so nothing but the call's check stands between the claim and the read after it. Each
// stops at that call with the located message, on sanitized native and the JavaScript backend, where
// Node prints undefined. Honest programs agree with Node, a generic claim checked against the target
// its call resolved to among them. Mutant: the call through a function type left unchecked, which
// reddens every lie.
func TestFunctionTypeClaimIsCheckedAtTheCall(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"param_honest", "iface_honest", "trycast"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			source, err := os.ReadFile("testdata/predicates_function_type/" + name + ".ts")
			if err != nil {
				t.Fatal(err)
			}
			lowersAndAgreesWithNodeAs(t, string(source), ".ts")
		})
	}
	for _, lie := range []struct{ name, callee, where string }{
		{"param_asserts_lie", "assertIdentifier", ":3:108"},
		{"iface_asserts_lie", "a.assertIdentifier", ":5:48"},
		{"overload_checked_impl", "assert", ":4:101"},
	} {
		t.Run(lie.name, func(t *testing.T) {
			t.Parallel()
			source, err := os.ReadFile("testdata/predicates_function_type/" + lie.name + ".ts")
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowerTypeScriptAssertionSource(t, string(source))
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
			prefix := "adamic: panic: predicate " + lie.callee + " at "
			suffix := lie.where + " asserts branch: source Node, target Identifier\n"
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
				if code != 70 || out.Len() != 0 || !strings.HasPrefix(errs.String(), prefix) || !strings.HasSuffix(errs.String(), suffix) {
					t.Fatalf("%s: exit %d stdout %q stderr %q; want 70 and %q...%q", command.Path, code, out.String(), errs.String(), prefix, suffix)
				}
			}
		})
	}
}

// What the check at the call can't reach keeps its refusal. A library's claim is called inside the
// library, so a checked parameter forwarded to filter or find keeps predicateArguments' proof (Lower's
// must). An overload whose implementation takes a plain callback calls it unchecked, so the overload's
// argument keeps its proof too. A target with no runtime test is NotYet at the call, never a vacuous
// check. .a needs proof and refuses the claim. Mutants: the library test dropped, which admits the
// forwards; the overload's implementation rule dropped, which admits overload_union_impl.
func TestFunctionTypeClaimKeepsWhatTheCallCantCheck(t *testing.T) {
	t.Parallel()
	read := func(name string) string {
		source, err := os.ReadFile("testdata/predicates_function_type/" + name + ".ts")
		if err != nil {
			t.Fatal(err)
		}
		return string(source)
	}
	for _, refusal := range []struct{ name, what string }{
		{"forward_filter", `an unproven predicate argument for parameter predicate (argument "test")`},
		{"forward_find", `an unproven predicate argument for parameter predicate (argument "test")`},
		{"overload_union_impl", `an unproven predicate argument for parameter assert (argument "(node: Node): void => {}")`},
	} {
		_, err := lowerTypeScriptAssertionSource(t, read(refusal.name))
		var refused *Refused
		if !errors.As(err, &refused) || refused.What != refusal.what {
			t.Errorf("%s: want refused as %q in .ts, got %v", refusal.name, refusal.what, err)
		}
	}
	_, err := lowerTypeScriptAssertionSource(t, read("untagged"))
	var pending *NotYet
	if !errors.As(err, &pending) || !strings.Contains(err.Error(), "predicate overload target Named") {
		t.Errorf("untagged: want NotYet naming the target Named, got %v", err)
	}
	_, err = lowerSource(t, read("iface_honest"))
	var refused *Refused
	if !errors.As(err, &refused) || refused.What != "a type predicate whose return is not proven (there is no body proving this parameter)" {
		t.Errorf("iface_honest in .a: want the claim refused, got %v", err)
	}
}
