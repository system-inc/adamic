package oracle

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// A host object is lowered only through the members its registration names, and rule 800 finds a
// member by the symbol an expression names. Seen through a program type, performance's
// eventLoopUtilization or toJSON is the type's member, so on main native died with "compiler bug: a
// method the checker proved is there is missing" where Node printed the result (#8k12t4s, Language's
// h11, h2 and h4). A registered host object seen through a type that reaches a member no
// registration names is now NotYet at the relation (expression.go, nodeHostViewReach): through an
// initializer, a parameter, an assignment, a return and an object literal's slot. The fixtures sit in
// reland_refused, since the flow and fresh corpora lower every program in testdata. Mutant: the guard
// dropped; with ADAMIC_HOST_VIEW_MUTANT=1 the pins must lower, and each dies natively again.
func TestNodeHostViewReachNotYet(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ file, at, view, member, node string }{
		{"node_host_view_method.ts", "2:66", "{ eventLoopUtilization(): { readonly idle: number; }; }", "eventLoopUtilization", "true\n"},
		{"node_host_view_initializer.ts", "2:58", "{ toJSON(): { readonly timeOrigin: string; }; }", "toJSON", "undefined\n"},
		{"node_host_view_parameter.ts", "3:27", "{ toJSON(): { readonly timeOrigin: string; }; }", "toJSON", "undefined\n"},
		{"node_host_view_assignment.ts", "3:5", "{ eventLoopUtilization(): { readonly idle: number; }; } | undefined", "eventLoopUtilization", "true\n"},
		{"node_host_view_return.ts", "2:81", "{ eventLoopUtilization(): { readonly idle: number; }; }", "eventLoopUtilization", "true\n"},
		{"node_host_view_slot.ts", "2:89", "{ eventLoopUtilization(): { readonly idle: number; }; }", "eventLoopUtilization", "true\n"},
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
					t.Fatalf("the mutant must admit the program: %v", err)
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
			what := probe.file + ":" + probe.at + ": stage 0 can't lower a host object of type Performance seen as " + probe.view + ", which reaches node:perf_hooks.Performance." + probe.member + ", a member no host registration names"
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
func TestNodeHostViewStopsHeld(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ file, at, what string }{
		{"node_host_view_held_namespace.ts", "2:102", "a namespace object used as a value"},
		{"node_host_view_held_destructured_namespace.ts", "2:20", "a namespace object used as a value"},
		{"node_host_view_held_buffer.ts", "3:44", "Buffer or Hash viewed as another object type (native host internal slots)"},
		{"node_host_view_held_buffer_method.ts", "3:63", "Buffer or Hash viewed as another object type (native host internal slots)"},
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
