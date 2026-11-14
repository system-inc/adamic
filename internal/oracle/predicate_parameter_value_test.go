package oracle

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/lower"
)

// tsc's getFirstJSDocTag (utilitiesPublic.ts:1280) hands its predicate parameter to core.ts's find, whose
// narrowing overload believes every true answer while its implementation holds the callback as a plain
// boolean one (#jqf1drp). In .ts the parameter is passed as its checked closure, built at the pass site:
// it calls the parameter's value and checks the parameter's own claim on each true answer before handing
// it on (rulings.md:140, #vdt9f1q's road). Each program is the stored .a, which Node runs, and its .ts
// copy, which lowers; the prelude is find's overloads as core.ts writes them (with stage 3's array[i]!),
// a Circle | Square union tagged by kind, getFirst forwarding to find, and isCircle, a direct kind
// compare. Source mutants (predicate_parameter_value.mutants.md): the parameter passed straight through,
// with no closure; the parameter count dropped from the closure's stop; the spread dropped from
// predicateArguments; and the closure's build skipped where an identifier is read, which the admitted
// guard stops.

// predicateParameterSite is where getFirst passes its parameter to find in every fixture: the checked
// closure's words name it.
func predicateParameterSite(t *testing.T, path string) string {
	t.Helper()
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for number, line := range strings.Split(string(source), "\n") {
		if column := strings.Index(line, "return find(shapes, predicate);"); column >= 0 {
			return fmt.Sprintf("%s:%d:%d", path, number+1, column+len("return find(shapes, ")+1)
		}
	}
	t.Fatalf("%s: no pass site", path)
	return ""
}

// predicateParameterClosures counts the checked closures a program holds.
func predicateParameterClosures(program *ir.Program) int {
	count := 0
	for _, function := range program.Functions {
		if function.Source.Role == "checked_predicate_parameter" {
			count++
		}
	}
	return count
}

// The honest programs lower through the checked closure and agree with Node on sanitized and released
// native and the JavaScript backend, with no leak: honest calls getFirst with isCircle, and conditional
// with `pick ? isCircle : isRound`, two proven predicates claiming Circle, which getFirst takes as one
// value. uncalled is honest beside a second generic forwarder no call reaches: the contracts pass admits
// its parameter, but lowering never lowers its body, so the admitted guard, which judges only lowered
// regions, doesn't stop it, and the one closure is getFirst's. As .a, honest is the twin the parameter
// keeps refused in, with its fix named.
func TestPredicateParameterPassedAsCheckedClosure(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, nodeOut string }{
		{"honest", "radius 3\nnone\n"},
		{"conditional", "radius 3\n"},
		{"uncalled", "radius 3\nnone\n"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			witness, path := predicateParameterInput(t, probe.name)
			node := onNode(t, witness)
			if difference := disagreement(run{stdout: []byte(probe.nodeOut)}, node); difference != "" {
				t.Fatal("source Node: " + difference)
			}
			if probe.name == "honest" {
				_, err := lowered(t, witness)
				var refused *lower.Refused
				if !errors.As(err, &refused) {
					t.Fatalf(".a witness: want the parameter refused, got %v", err)
				}
				t.Logf(".a: %v", err)
				if refused.What != `an unproven predicate argument for parameter predicate (argument "predicate"): a parameter carries no proof of its claim` || refused.Fix != "prove it with a named predicate whose body proves both branches, passed directly, or return a boolean and narrow at the caller (adamic/no-type-predicate)" {
					t.Fatalf(".a witness: want the parameter refused with its fix, got %q, fix %q", refused.What, refused.Fix)
				}
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			if closures := predicateParameterClosures(program); closures != 1 {
				t.Fatalf("want one checked closure, got %d", closures)
			}
			sanitized, binary := nativelyUncached(t, program)
			for index, got := range []run{sanitized, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if difference := disagreement(node, got); difference != "" {
					t.Fatalf("%s on %s: exit %d, stdout %q, stderr %q", difference, []string{"sanitized native", "release native", "the JavaScript backend"}[index], got.exitCode, got.stdout, got.stderr)
				}
			}
			if report := leaksUncached(t, program, binary); report != "" {
				t.Fatal(report)
			}
		})
	}
}

// isCircleLie says yes to a Square, and Node prints "radius undefined". lie hands it to getFirst
// directly, and escape through getFirst held as a value. isCircleLie is a checked predicate, so the value
// it becomes is its forwarder, which checks it first, inside the closure's call: each program stops there
// in the forwarder's words. With the forwarder's check taken out of the IR, the closure's own check stops
// the lie at getFirst's pass site, exit 70 with nothing printed, on sanitized and released native and the
// JavaScript backend. That second run is the closure's own stop, which the source mutant reddens.
func TestPredicateParameterLieStopsAtTheClosure(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"lie", "escape"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			witness, path := predicateParameterInput(t, name)
			if difference := disagreement(run{stdout: []byte("radius undefined\n")}, onNode(t, witness)); difference != "" {
				t.Fatal("source Node: " + difference)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			declared := ""
			source, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			for number, line := range strings.Split(string(source), "\n") {
				if strings.HasPrefix(line, "function isCircleLie(") {
					declared = fmt.Sprintf("%s:%d:1", path, number+1)
				}
			}
			stops := func(program *ir.Program, message string) {
				t.Helper()
				want := run{stderr: []byte(message), exitCode: 70}
				sanitized, _ := nativelyUncached(t, program)
				for index, got := range []run{sanitized, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
					t.Logf("%s: exit %d, stdout %q, stderr %q", []string{"sanitized native", "release native", "the JavaScript backend"}[index], got.exitCode, got.stdout, got.stderr)
					if difference := disagreement(want, got); difference != "" {
						t.Fatalf("%s on %s: want %q", difference, []string{"sanitized native", "release native", "the JavaScript backend"}[index], message)
					}
				}
			}
			stops(program, fmt.Sprintf("adamic: panic: predicate isCircleLie used as a value, declared at %s, true branch: source Shape, target Circle\n", declared))
			changes := 0
			for index := range program.Functions {
				if !strings.HasSuffix(program.Functions[index].Name, "_value") {
					continue
				}
				program.Functions[index].Body = mutateReadiness(program.Functions[index].Body, func(node any) any {
					if panic, ok := node.(ir.Panic); ok {
						constant, ok := panic.Message.(ir.StringConstant)
						if ok && strings.Contains(program.Strings[constant.Index], " used as a value, ") {
							changes++
							return ir.Evaluate{Value: ir.NumberConstant{Value: 0}}
						}
					}
					return node
				})
			}
			if changes != 1 {
				t.Fatalf("the forwarder's check: removed %d", changes)
			}
			stops(program, fmt.Sprintf("adamic: panic: predicate parameter predicate passed at %s, true branch: source Shape, target Circle\n", predicateParameterSite(t, path)))
		})
	}
}

// What doesn't reach the closure stops at its own wall, measured: `isCircle.bind(undefined)` at bind's
// stop for a narrowing signature, and an overload set at its own stop as a value, which main has.
func TestPredicateParameterOtherRoadsStop(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, what string }{
		{"bind", "bind of a narrowing function signature"},
		{"overload", "an overloaded function as a value"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, path := predicateParameterInput(t, probe.name)
			_, err := lowered(t, path)
			t.Logf("%v", err)
			var notYet *lower.NotYet
			if !errors.As(err, &notYet) || notYet.What != probe.what {
				t.Fatalf("want NotYet %q, got %v", probe.what, err)
			}
		})
	}
}

// predicateParameterRuns is what program does on sanitized and released native and the JavaScript
// backend, each beside node, for a row whose stop didn't hold.
func predicateParameterRuns(t *testing.T, program *ir.Program, node run) string {
	t.Helper()
	sanitized, _ := nativelyUncached(t, program)
	described := []string{}
	for index, got := range []run{sanitized, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
		against := disagreement(node, got)
		if against == "" {
			against = "agrees with Node"
		}
		described = append(described, fmt.Sprintf("%s exit %d, stdout %q, stderr %q, %s", []string{"sanitized native", "release native", "the JavaScript backend"}[index], got.exitCode, got.stdout, got.stderr, against))
	}
	return strings.Join(described, "; ")
}

// predicateParameterSpreadSite is where getFirstSpread spreads its arguments into find.
func predicateParameterSpreadSite(t *testing.T, path string) string {
	t.Helper()
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for number, line := range strings.Split(string(source), "\n") {
		if column := strings.Index(line, "return find(...args);"); column >= 0 {
			return fmt.Sprintf("%s:%d:%d", path, number+1, column+len("return find(")+1)
		}
	}
	t.Fatalf("%s: no spread site", path)
	return ""
}

// A value assignable to the parameter's type may declare more parameters than the type names, and the
// closure calls it with only those, while Node hands it find's index. Without the stop, measured at -2
// (J-7cm0rvp): optional_index's isCircleSeen and arrow_optional's arrow print "seen at undefined" where
// Node prints "seen at 0", default_index's prints "seen at -1", and optional_decides' isCircleAt and
// proven_optional's proven isCircleSeen answer none where Node finds the circle, a wrong answer with no
// lie. Each stops at getFirst's pass site until the closure forwards every argument.
func TestPredicateParameterClosureStopsOnMoreParameters(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, nodeOut string }{
		{"optional_index", "seen at 0\nseen at 1\nradius 3\n"},
		{"default_index", "seen at 0\nseen at 1\nradius 3\n"},
		{"optional_decides", "radius 3\n"},
		{"arrow_optional", "seen at 0\nseen at 1\nradius 3\n"},
		{"proven_optional", "radius 3\n"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			witness, path := predicateParameterInput(t, probe.name)
			node := onNode(t, witness)
			if difference := disagreement(run{stdout: []byte(probe.nodeOut)}, node); difference != "" {
				t.Fatal("source Node: " + difference)
			}
			_, err := lowered(t, witness)
			t.Logf(".a: %v", err)
			program, err := lowered(t, path)
			if err == nil {
				t.Fatalf(".ts lowered: %s", predicateParameterRuns(t, program, node))
			}
			t.Logf(".ts: %v", err)
			var notYet *lower.NotYet
			want := "a predicate parameter passed as a checked closure, which may call a function that observes its argument count, collects rest arguments or declares more parameters than its type names"
			if !errors.As(err, &notYet) || notYet.What != want || notYet.Where != predicateParameterSite(t, path) {
				t.Fatalf("want NotYet %q at %s, got %v", want, predicateParameterSite(t, path), err)
			}
		})
	}
}

// find(...args) hands find's narrowing overload a predicate no written argument is, so predicateArguments
// stops at the spread rather than skip the parameter: spread_static's lying static method and
// spread_declared's lying declaration would each reach find raw, and Node prints "radius undefined".
func TestPredicateParameterSpreadStops(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"spread_static", "spread_declared"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			witness, path := predicateParameterInput(t, name)
			node := onNode(t, witness)
			if difference := disagreement(run{stdout: []byte("radius undefined\n")}, node); difference != "" {
				t.Fatal("source Node: " + difference)
			}
			_, err := lowered(t, witness)
			t.Logf(".a: %v", err)
			program, err := lowered(t, path)
			if err == nil {
				t.Fatalf(".ts lowered: %s", predicateParameterRuns(t, program, node))
			}
			t.Logf(".ts: %v", err)
			var notYet *lower.NotYet
			want := "a spread argument covering a predicate parameter"
			if !errors.As(err, &notYet) || notYet.What != want || notYet.Where != predicateParameterSpreadSite(t, path) {
				t.Fatalf("want NotYet %q at %s, got %v", want, predicateParameterSpreadSite(t, path), err)
			}
		})
	}
}

// The parameter count's control: direct_control passes isCircleSeen to find itself, with no closure
// between, so find's index reaches it as Node hands it, and it agreed with Node at -2 (J-7cm0rvp). Its
// getFirst is never called, so it's a second uncalled forwarder the admitted guard leaves alone.
func TestPredicateParameterClosureCountControls(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, nodeOut string }{
		{"direct_control", "seen at 0\nseen at 1\nradius 3\n"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			witness, path := predicateParameterInput(t, probe.name)
			node := onNode(t, witness)
			if difference := disagreement(run{stdout: []byte(probe.nodeOut)}, node); difference != "" {
				t.Fatal("source Node: " + difference)
			}
			_, err := lowered(t, witness)
			t.Logf(".a: %v", err)
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			if closures := predicateParameterClosures(program); closures != 0 {
				t.Fatalf("want no checked closure, got %d", closures)
			}
			sanitized, _ := nativelyUncached(t, program)
			for index, got := range []run{sanitized, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if difference := disagreement(node, got); difference != "" {
					t.Fatalf("%s on %s: exit %d, stdout %q, stderr %q", difference, []string{"sanitized native", "release native", "the JavaScript backend"}[index], got.exitCode, got.stdout, got.stderr)
				}
			}
		})
	}
}

// predicateParameterInput is a fixture's stored .a, which Node runs, and its .ts copy.
func predicateParameterInput(t *testing.T, name string) (string, string) {
	t.Helper()
	return predicateInputFrom(t, "internal/lower/testdata/predicate_parameter", name)
}
