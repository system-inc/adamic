package oracle

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// A host object is lowered only through the members its registration names, and rule 800 finds a member by the symbol
// an expression names. Seen through a program type, performance's eventLoopUtilization or toJSON is the type's member,
// so on main native died with "compiler bug: a method the checker proved is there is missing" where Node printed the
// result (#8k12t4s, Language's h11, h2 and h4). A registered host object seen through a type that reaches a member no
// registration names is now NotYet at the relation (expression.go, nodeHostViewReach): through an initializer, a
// parameter, an assignment, a return and an object literal's slot, and through a union's arm, which the walk judges
// member by member, since a union's own properties are only the ones its members share. A value holding the host
// object, an object, an array, a tuple or a Map of it, is walked with its view to the host object inside (Workers'
// read), and a generic argument is judged against its parameter as declared, whose constraint is what the body names. A
// function or a class seen as a callable view is walked through its call and construct results, and a program's class
// instance through its members. Past its depth bound the walk fails closed on what may hold a host object. The fixtures
// sit in reland_refused, since the flow and fresh corpora lower every program in testdata. Mutants: the guard dropped,
// the walk's union branch dropped, the container walk dropped, the declared argument dropped, the depth bound's quiet
// stop for a container restored, the top-level signature pairing dropped, and a class instance taken for a library
// container again. Each turns its pins red, and with ADAMIC_HOST_VIEW_MUTANT=1 each pin a mutant admits must die
// natively again.
func TestNodeHostViewReachNotYet(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ file, at, subject, view, member, node string }{
		{"node_host_view_method.ts", "2:66", "a host object of type Performance", "{ eventLoopUtilization(): { readonly idle: number; }; }", "eventLoopUtilization", "true\n"},
		{"node_host_view_initializer.ts", "2:58", "a host object of type Performance", "{ toJSON(): { readonly timeOrigin: string; }; }", "toJSON", "undefined\n"},
		{"node_host_view_parameter.ts", "3:27", "a host object of type Performance", "{ toJSON(): { readonly timeOrigin: string; }; }", "toJSON", "undefined\n"},
		{"node_host_view_assignment.ts", "3:5", "a host object of type Performance", "{ eventLoopUtilization(): { readonly idle: number; }; } | undefined", "eventLoopUtilization", "true\n"},
		{"node_host_view_return.ts", "2:81", "a host object of type Performance", "{ eventLoopUtilization(): { readonly idle: number; }; }", "eventLoopUtilization", "true\n"},
		{"node_host_view_slot.ts", "2:89", "a host object of type Performance", "{ eventLoopUtilization(): { readonly idle: number; }; }", "eventLoopUtilization", "true\n"},
		// A union view is judged member by member: typeof narrows to the arm that reaches the member.
		{"node_host_view_union.ts", "2:75", "a host object of type Performance", "number | { eventLoopUtilization(): { readonly idle: number; }; }", "eventLoopUtilization", "true\n"},
		// A container holding the host object is walked with its view to the host object inside.
		{"node_host_view_container_object.ts", "4:36", "a value of type { p: Performance; } holding a host object", "{ readonly p: View; }", "eventLoopUtilization", "true\n"},
		{"node_host_view_container_array.ts", "4:29", "a value of type Performance[] holding a host object", "readonly View[]", "eventLoopUtilization", "true\n"},
		{"node_host_view_container_tuple.ts", "4:39", "a value of type [string, Performance] holding a host object", "readonly [string, View]", "eventLoopUtilization", "true\n"},
		{"node_host_view_container_map.ts", "4:41", "a value of type Map<string, Performance> holding a host object", "ReadonlyMap<string, View>", "eventLoopUtilization", "true\n"},
		// Promise.resolve of the host object, typed Promise<View>, stopped on d9dc670e only at the
		// structural-value refusal; the walk stops it first, at the type argument.
		{"node_host_view_container_promise.ts", "3:66", "a value of type Promise<Performance> holding a host object", "Promise<View>", "eventLoopUtilization", "true\n"},
		// A function seen as a callable view gives what its call gives (Workers' read of 20cf2e2d), and
		// a class instance holds what its members hold, though it's a reference to itself with this.
		{"node_host_view_function.ts", "4:26", "a value of type () => Performance holding a host object", "() => View", "eventLoopUtilization", "true\n"},
		{"node_host_view_class_method.ts", "4:32", "a value of type Maker holding a host object", "{ make(): View; }", "eventLoopUtilization", "true\n"},
		// Ten containers deep is past the walk's bound, which fails closed on what may hold a host
		// object (no member: the message says where it stopped instead).
		{"node_host_view_container_deep.ts", "4:196", "a value of type", "", "", "true\n"},
		// A generic body names its parameter's constraint, so an argument is judged against the
		// parameter as declared, a literal made at the call included.
		{"node_host_view_generic.ts", "4:24", "a host object of type Performance", "T", "eventLoopUtilization", "true\n"},
		{"node_host_view_generic_array.ts", "4:7", "a value of type Performance[] holding a host object", "readonly T[]", "eventLoopUtilization", "true\n"},
	} {
		t.Run(probe.file, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/reland_refused", probe.file))
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			if disagreement(run{stdout: []byte(probe.node)}, truth) != "" {
				t.Fatalf("source Node: exit %d, stdout %q, stderr %q", truth.exitCode, truth.stdout, truth.stderr)
			}
			program, err := lowered(t, path)
			if os.Getenv("ADAMIC_HOST_VIEW_MUTANT") == "1" {
				if err != nil {
					t.Skipf("this mutant still refuses it: %v", err)
				}
				native, _ := nativelyUncached(t, program)
				for name, got := range map[string]run{"native": native, "release": releasedUncached(t, program)} {
					if disagreement(truth, got) == "" || !strings.Contains(string(got.stderr), "compiler bug: a method the checker proved is there is missing") {
						t.Fatalf("the mutant's %s program: exit %d, stdout %q, stderr %q", name, got.exitCode, got.stdout, got.stderr)
					}
					t.Logf("mutant admits it: %s exit %d, stderr %q where Node prints %q", name, got.exitCode, got.stderr, truth.stdout)
				}
				return
			}
			what := probe.file + ":" + probe.at + ": stage 0 can't lower " + probe.subject + " seen as " + probe.view + ", which reaches node:perf_hooks.Performance." + probe.member + ", a member no host registration names"
			if probe.member == "" {
				what = probe.file + ":" + probe.at + ": stage 0 can't lower " + probe.subject
				if err == nil || !strings.Contains(err.Error(), "which goes past the depth the view check follows, at ") {
					t.Fatalf("want it refused past the depth the view check follows, got %v", err)
				}
			}
			var notYet *lower.NotYet
			if !errors.As(err, &notYet) || !strings.Contains(err.Error(), what) {
				t.Fatalf("want %q, got %v", what, err)
			}
		})
	}
}

// A host object seen through a type that reaches only registered members still lowers and agrees
// with Node, related or destructured (Language's h10 and h12), on sanitized native, release native and
// JavaScript.
func TestNodeHostViewRegisteredAgrees(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source string }{
		{"timeOrigin related", "import { performance } from \"node:perf_hooks\";\nconst p: { readonly timeOrigin: number } = performance;\nconsole.log(String(p.timeOrigin > 0));\n"},
		{"timeOrigin destructured", "import { performance } from \"node:perf_hooks\";\nconst { timeOrigin } = performance;\nconsole.log(String(timeOrigin > 0));\n"},
		{"now related", "import { performance } from \"node:perf_hooks\";\nconst p: { now(): number } = performance;\nconsole.log(String(p.now() >= 0));\n"},
		{"realpathSync through a view giving a string", "import * as fs from \"node:fs\";\nconst real: (path: string) => string = fs.realpathSync;\nconst native: (path: string) => string = fs.realpathSync.native;\nconsole.log(String(real(\".\").length > 0 && native(\".\").length > 0));\n"},
		{"timeOrigin through a union's arm", "import { performance } from \"node:perf_hooks\";\nconst p: number | { readonly timeOrigin: number } = performance;\nif (typeof p !== \"number\") console.log(String(p.timeOrigin > 0)); else console.log(\"none\");\n"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), "main.ts")
			if err := os.WriteFile(path, []byte(probe.source), 0o644); err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			if disagreement(run{stdout: []byte("true\n")}, truth) != "" {
				t.Fatalf("source Node: exit %d, stdout %q, stderr %q", truth.exitCode, truth.stdout, truth.stderr)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatalf("a registered member's view must lower: %v", err)
			}
			native, _ := natively(t, program)
			for backend, got := range map[string]run{"native": native, "release": released(t, program), "javascript": onJavaScriptBackend(t, program)} {
				if problem := disagreement(truth, got); problem != "" {
					t.Fatalf("%s: %s: exit %d, stdout %q, stderr %q", backend, problem, got.exitCode, got.stdout, got.stderr)
				}
			}
		})
	}
}

// What stopped these before stops them still: process imported whole is a namespace object, related
// or destructured (h1, h3), and a Buffer seen as another object type is refused whole (h13, h14).
// Workers' other roads stop where they did on d9dc670e, each at its own rule.
func TestNodeHostViewStopsHeld(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ file, at, what string }{
		{"node_host_view_held_namespace.ts", "2:102", "a namespace object used as a value"},
		{"node_host_view_held_destructured_namespace.ts", "2:20", "a namespace object used as a value"},
		{"node_host_view_held_buffer.ts", "3:44", "Buffer or Hash viewed as another object type (native host internal slots)"},
		{"node_host_view_held_buffer_method.ts", "3:63", "Buffer or Hash viewed as another object type (native host internal slots)"},
		// Workers' roads that other rules already stop: a mutable container's view can write what it
		// reads, a literal's element is judged where it's made, an intersection with object loses its
		// slot tags, satisfies keeps the host type (rule 800 stops the unregistered member's own read,
		// and widening it afterwards is the plain relation), and a generic's result is the
		// instantiated host type.
		{"node_host_view_held_container_mutable.ts", "4:27", "which can write View where Performance is read"},
		{"node_host_view_held_literal_element.ts", "3:30", "a host object of type Performance seen as View, which reaches node:perf_hooks.Performance.eventLoopUtilization"},
		{"node_host_view_held_intersection.ts", "3:40", "an object with a nullable field viewed as unknown or object"},
		{"node_host_view_held_satisfies.ts", "4:20", "node:perf_hooks.EventLoopUtilization.idle"},
		{"node_host_view_held_satisfies_widened.ts", "3:11", "a host object of type Performance seen as View, which reaches node:perf_hooks.Performance.eventLoopUtilization"},
		{"node_host_view_held_generic_return.ts", "4:20", "node:perf_hooks.EventLoopUtilization.idle"},
		// A callable one level down, a property or an argument, is already judged at 20cf2e2d.
		{"node_host_view_held_callable_property.ts", "4:44", "a value of type { get: () => Performance; } holding a host object seen as { readonly get: () => View; }, which reaches node:perf_hooks.Performance.eventLoopUtilization"},
		{"node_host_view_held_function_argument.ts", "4:30", "a host object of type Performance seen as View, which reaches node:perf_hooks.Performance.eventLoopUtilization"},
		// Overloads taking the same arguments with a narrower result first (TestNodeRegisteredOverloads
		// in internal/lower): mkdtempSync isn't a value at all, and realpathSync's value takes one path
		// and nothing Node reads as options, so its string overload is the one a view can reach.
		{"node_host_view_held_overload_value.ts", "3:40", "node:fs.mkdtempSync as a value"},
		{"node_host_view_held_overload_options.ts", "3:20", "is admitted only as a const called with one path"},
	} {
		t.Run(probe.file, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/reland_refused", probe.file))
			if err != nil {
				t.Fatal(err)
			}
			_, err = lowered(t, path)
			if err == nil || !strings.Contains(err.Error(), probe.file+":"+probe.at) || !strings.Contains(err.Error(), probe.what) {
				t.Fatalf("want it stopped at %s by %q, got %v", probe.at, probe.what, err)
			}
		})
	}
}
