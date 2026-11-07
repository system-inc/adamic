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

// An arrow or a function expression in .ts claiming a predicate its body doesn't prove is a checked
// closure (#1w5bve2, R-T80 for every declaration shape, ruled Oct 27): it checks its claim's true side
// at each of its own returns, so every call of it, direct, through a value or as an argument, stops
// on a lying true. Each source is held to Node first, then run natively under the sanitizers and
// through the JavaScript backend; the .a twin stays refused.
func TestCheckedPredicateClosures(test *testing.T) {
	test.Parallel()
	for _, probe := range []struct{ name, source, nodeOut, checkedOut, predicate string }{
		// A concise arrow held by a name and called directly.
		{"concise arrow", `const isNumber = (value: number | string): value is number => true;
function use(value: number | string): void { if (isNumber(value)) console.log("claimed"); }
use("text");`, "claimed\n", "", "isNumber"},
		// A function expression with a block body.
		{
			"function expression",
			`const isNumber = function (value: number | string): value is number { return value !== ""; };
function use(value: number | string): void { if (isNumber(value)) console.log("claimed"); }
use("text");`,
			"claimed\n",
			"",
			"isNumber",
		},
		// An arrow written in place as an argument, the shape of tsc's filter and findAncestor callbacks.
		{
			"argument",
			`function count(values: (number | string)[], test: (value: number | string) => value is number): number {
  let found = 0;
  for (const value of values) { if (test(value)) found += 1; }
  return found;
}
console.log(String(count([1, "a"], (value): value is number => true)));`,
			"2\n",
			"",
			"an arrow",
		},
		// An honest claim the body prover can't read checks and agrees with Node.
		{"honest", `const isNumber = (value: number | string): value is number => ["number"].includes(typeof value);
function use(value: number | string): void { if (isNumber(value)) { console.log("number"); } else { console.log("string"); } }
use(1); use("text");`, "number\nstring\n", "number\nstring\n", ""},
	} {
		test.Run(probe.name, func(test *testing.T) {
			test.Parallel()
			directory := test.TempDir()
			path := filepath.Join(directory, "main.ts")
			if writeError := os.WriteFile(path, []byte(probe.source+"\n"), 0o644); writeError != nil {
				test.Fatal(writeError)
			}
			runner, absolutePathError := filepath.Abs("../../oracle/node.mjs")
			if absolutePathError != nil {
				test.Fatal(absolutePathError)
			}
			run := func(command *exec.Cmd, stdout, stderrPrefix, stderrPart string, code int) {
				test.Helper()
				var out, errs bytes.Buffer
				command.Stdout, command.Stderr = &out, &errs
				got := 0
				if runError := command.Run(); runError != nil {
					var exit *exec.ExitError
					if !errors.As(runError, &exit) {
						test.Fatal(runError)
					}
					got = exit.ExitCode()
				}
				if out.String() != stdout ||
					got != code ||
					!strings.HasPrefix(errs.String(), stderrPrefix) ||
					!strings.Contains(errs.String(), stderrPart) ||
					stderrPrefix == "" && errs.Len() != 0 {
					test.Fatalf(
						"%s: exit %d stdout %q stderr %q; want %d %q, stderr %q...%q",
						command.Path,
						got,
						out.String(),
						errs.String(),
						code,
						stdout,
						stderrPrefix,
						stderrPart,
					)
				}
			}
			run(exec.Command("node", "--disable-warning=ExperimentalWarning", runner, path), probe.nodeOut, "", "", 0)
			program, lowerTypeScriptAssertionSourceError := lowerTypeScriptAssertionSource(test, probe.source)
			if lowerTypeScriptAssertionSourceError != nil {
				test.Fatal(lowerTypeScriptAssertionSourceError)
			}
			stdout, prefix, part, code := probe.checkedOut, "", "", 0
			if probe.predicate != "" {
				prefix, part, code = "adamic: panic: predicate "+
					probe.predicate+
					", declared at ", ", true branch: source string | number, target number", 70
				if _, lowerSourceError := lowerSource(test, probe.source); lowerSourceError == nil || !strings.Contains(
					lowerSourceError.Error(),
					"type predicate",
				) && !strings.Contains(lowerSourceError.Error(), "predicate argument") {
					test.Fatalf("lying .a body admitted: %v", lowerSourceError)
				}
			}
			binary := filepath.Join(directory, "native")
			if buildError := native.Build(
				native.C(program),
				binary,
				native.Options{Sanitize: true},
			); buildError != nil {
				test.Fatal(buildError)
			}
			command := exec.Command(binary)
			command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1")
			run(command, stdout, prefix, part, code)
			generated := filepath.Join(directory, "generated.mjs")
			if writeError := os.WriteFile(generated, []byte(javascript.JavaScript(program)), 0o644); writeError != nil {
				test.Fatal(writeError)
			}
			run(
				exec.Command("node", "--disable-warning=ExperimentalWarning", runner, generated),
				stdout,
				prefix,
				part,
				code,
			)
		})
	}
}

// What stays refused in .ts: an asserts arrow, and an arrow whose body assigns the claimed parameter,
// which would check a value other than the argument.
func TestCheckedPredicateClosuresRefused(test *testing.T) {
	test.Parallel()
	for _, probe := range []struct{ name, source string }{
		{
			"asserts",
			`const check: (value: number | string) => asserts value is number = (value: number | string): asserts value is number => {};
function use(value: number | string): void { check(value); console.log(String(value + 1)); }
use(1);`,
		},
		// A false-side lie: Node throws at toUpperCase on 1, and native would read a number as a string.
		// The closure checks only its true side, so the walk (#1761dah) stops the narrowed read.
		{"false side narrowed", `const isNumber = (value: number | string): value is number => false;
function use(value: number | string): void { if (!isNumber(value)) { console.log(value.toUpperCase()); } }
use(1);`},
		// A let can hold any function of the arrow's type when it's called (Lower's mutable.ts shape).
		{"let callee", `let isNumber = (value: number | string): value is number => true;
function use(value: number | string): void { if (isNumber(value)) console.log("claimed"); }
use("text");`},
		// A let alias of a const-bound arrow names the arrow's declaration but can be reassigned.
		{"let alias", `const isNumber = (value: number | string): value is number => true;
let test = isNumber;
function use(value: number | string): void { if (test(value)) console.log("claimed"); }
use("text");`},
		// Compiler's must: a namespace member assigned in the return built its own return, unchecked.
		{"namespace assignment return", `namespace N { export let ok = false; }
const isS = (value: unknown): value is string => { return N.ok = true; };
const x: unknown = 1;
if (isS(x)) console.log(String(x.length));`},
		{"local assignment return", `let ok = false;
const isS = (value: unknown): value is string => { return ok = true; };
const x: unknown = 1;
if (isS(x)) console.log(String(x.length));`},
		// Lower's must: with a default, the parameter holds 1 when the caller passed undefined, so a
		// check of the local would pass while the caller's own undefined is narrowed to number.
		{"defaulted parameter", `const isNumber = (value: unknown = 1): value is number => true;
const v: unknown = undefined;
if (isNumber(v)) console.log(v.toFixed(1));`},
		{
			"assigned parameter",
			`const isNumber = (value: number | string): value is number => { value = 1; return true; };
function use(value: number | string): void { if (isNumber(value)) console.log("claimed"); }
use("text");`,
		},
	} {
		test.Run(probe.name, func(test *testing.T) {
			test.Parallel()
			if _, lowerTypeScriptAssertionSourceError := lowerTypeScriptAssertionSource(
				test,
				probe.source,
			); lowerTypeScriptAssertionSourceError == nil {
				test.Fatal("admitted; want it refused")
			}
		})
	}
}
