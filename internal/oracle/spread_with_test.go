package oracle

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/native_runtime"
)

// spreadWithAgreeing are the literals #xy74m5x stopped that lower to adamic_object_spread_with (#k7g6f9h): an
// overwrite of another kind than the source's slot, one the source's type makes optional, and one it doesn't
// declare, with the width a { readonly a: number } may hold beside it. Each is held to its pinned output and to
// Node, through Object.keys, JSON.stringify, for-in and a later read where the program prints them. graph marks
// Memory's two graph-program pins, whose counted build must also end with nothing held in a region.
var spreadWithAgreeing = []struct {
	path, want string
	graph      bool
}{
	{"spread_with/width_extra.ts", "a,extra,k\n{\"a\":1,\"extra\":2,\"k\":5}\na,extra,k\n1 5\n{\"a\":1,\"extra\":2,\"k\":5}\n", false},
	{"spread_with/kind_change.ts", "k,label\n{\"k\":\"ten\",\"label\":\"one\"}\nstring ten one\nk,n\n{\"k\":5,\"n\":2}\nnumber 7\n", false},
	{"spread_with/optional_absent.ts", "a,z,k\n{\"a\":1,\"z\":3,\"k\":\"set\"}\n1 set 3\na,k,z\n{\"a\":1,\"k\":\"set\",\"z\":3}\n1 set 3\n", false},
	{"spread_with/resolver_245.ts", "path,extension,packageId,resolvedUsingTsExtension,originalPath\n{\"path\":\"/b.ts\",\"extension\":\".ts\",\"resolvedUsingTsExtension\":false,\"originalPath\":\"/c.ts\"}\n/c.ts\npath,extension,packageId,originalPath,resolvedUsingTsExtension\n{\"path\":\"/b.ts\",\"extension\":\".ts\",\"originalPath\":\"/c.ts\",\"resolvedUsingTsExtension\":false}\n/c.ts\npath,extension,packageId,resolvedUsingTsExtension\n{\"path\":\"/a.ts\",\"extension\":\".ts\",\"resolvedUsingTsExtension\":false}\nundefined\n", false},
	{"spread_with/resolver_279.ts", "resolvedModule,failedLookupLocations,affectingLocations,resolutionDiagnostics\n{\"resolvedModule\":{\"resolvedFileName\":\"/a.ts\"},\"failedLookupLocations\":[\"/x\",\"/y\"],\"resolutionDiagnostics\":[{\"messageText\":\"m\"}]}\nresolvedModule,alternateResult,failedLookupLocations,affectingLocations,resolutionDiagnostics\n{\"alternateResult\":\"/alt\",\"failedLookupLocations\":[\"/y\"],\"resolutionDiagnostics\":[{\"messageText\":\"m\"}]}\nmissing\n", false},
	{"spread_with/uniform.ts", "10 0 7 1\n", false},
	{"spread_with/effect_order.ts", "1 2 5\n", false},
	{"spread_with/backstop_plain.ts", "{\"length\":2,\"k\":1}\nend\n", false},
	{"spread_with/graph_member.ts", "b a b a\n", true},
	{"spread_with/graph_fresh.ts", "b fresh b\n", true},
	// #xy74m5x's own rows, which its stop held until a spread built its own layout.
	{"spread_overwrite/number_present.ts", "undefined\n", false},
	{"spread_overwrite/optional_string.ts", "true\nundefined\ntrue\nundefined\n", false},
	{"spread_overwrite/optional_number.ts", "true\nundefined\ntrue\nundefined\n", false},
	{"spread_overwrite/width.ts", "1 5\n", false},
	{"spread_overwrite/width_method.ts", "1 5\n", false},
	{"typed_presence/copied-empty-object.ts", "undefined\nundefined\n", false},
}

// Each SpreadWith literal lowers and agrees with Node natively, released and sanitized with the leak check, on the
// counted build with its counts balanced, and on the JavaScript backend. Its C calls adamic_object_spread_with, so
// a row can't pass through the copy it replaced, and a graph row is a graph program that adopts the result.
func TestSpreadWithAgrees(t *testing.T) {
	t.Parallel()
	for _, pin := range spreadWithAgreeing {
		t.Run(pin.path, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", pin.path))
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			if truth.exitCode != 0 || string(truth.stdout) != pin.want {
				t.Fatalf("Node: exit %d, stdout %q, want %q, stderr %q", truth.exitCode, truth.stdout, pin.want, truth.stderr)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatalf("Lower: %v", err)
			}
			code := native.C(program)
			if !strings.Contains(code, "adamic_object_spread_with(") {
				t.Fatal("the C never calls adamic_object_spread_with")
			}
			if pin.graph && (len(program.GraphTypes) == 0 || !strings.Contains(code, "adamic_graph_adopt_owned(")) {
				t.Fatalf("not a graph program that adopts: %d graph types", len(program.GraphTypes))
			}
			t.Run("native", func(t *testing.T) {
				got := releasedUncached(t, program)
				if difference := disagreement(truth, got); difference != "" {
					t.Errorf("%s: exit %d, stdout %q, stderr %q", difference, got.exitCode, got.stdout, got.stderr)
				}
			})
			t.Run("sanitized", func(t *testing.T) {
				got, leaked, comparable := nativelyCheckingLeaks(t, program, true)
				if leaked != "" {
					t.Errorf("leaks: %s", leaked)
				}
				if !comparable {
					return
				}
				if difference := disagreement(truth, got); difference != "" {
					t.Errorf("%s: exit %d, stdout %q, stderr %q", difference, got.exitCode, got.stdout, got.stderr)
				}
			})
			t.Run("counted", func(t *testing.T) {
				binary := filepath.Join(t.TempDir(), "counted")
				if err := native.Build(code, binary, native_runtime.Options{Count: true}); err != nil {
					t.Fatal(err)
				}
				got := execute(t, binary)
				if report := leakcheck.Unbalanced(leakRun(got)); report != "" {
					t.Errorf("counts: %s", report)
				}
				if string(got.stdout) != pin.want {
					t.Errorf("got stdout %q", got.stdout)
				}
				match := leakcheck.CountsLine.FindSubmatch(got.stderr)
				if match == nil {
					return
				}
				t.Logf("counts: allocations %s, frees %s, in regions %s", match[1], match[2], match[6])
				if pin.graph && string(match[6]) != "0" {
					t.Errorf("in regions %s at exit, want 0", match[6])
				}
			})
			t.Run("javascript", func(t *testing.T) {
				got := onJavaScriptBackend(t, program)
				if difference := disagreement(truth, got); difference != "" {
					t.Errorf("%s: exit %d, stdout %q, stderr %q", difference, got.exitCode, got.stdout, got.stderr)
				}
			})
		})
	}
}

// What keeps its stop, each where and in the words it stops: a SpreadWith literal whose source may be undefined
// keeps #xy74m5x's (#k7g6f9h's first cut; typed_presence_test.go pins 65-leading-overwrite.ts the same), and one whose source may be a class instance with an accessor, a
// getter, a setter or an object literal's own, stops before adamic_object_spread_with would refuse it at run time,
// by the source's type or by the accessor registry (#vjvcnnj narrows both). If a stop goes missing, the program
// is run natively under the sanitizers and its run is part of the failure.
func TestSpreadWithStops(t *testing.T) {
	t.Parallel()
	const held = "after a spread whose source holds k as"
	const accessor = "which adamic_object_spread_with doesn't copy yet (#vjvcnnj)"
	// An accessor row's words end the stop's, so the type's stop and the registry's are told apart.
	for _, pin := range []struct {
		file, node, at, words string
		ends                  bool
	}{
		{"spread_overwrite/number_source.ts", "undefined\nundefined\n", "k: undefined", held, false},
		{"spread_with/accessor_type.ts", "t 1\n", "...source", accessor, true},
		{"spread_with/accessor_getter.ts", "t 1\n20\n", "...source", accessor + ", Sized among them", true},
		{"spread_with/accessor_setter.ts", "true\ng 1\n", "...source", accessor + ", Gate among them", true},
		{"spread_with/accessor_literal.ts", "l 1\n3\n", "...source", accessor + ", an object literal with accessors among them", true},
	} {
		t.Run(pin.file, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", pin.file))
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			if disagreement(run{stdout: []byte(pin.node)}, truth) != "" {
				t.Fatalf("source Node: exit %d, stdout %q, stderr %q", truth.exitCode, truth.stdout, truth.stderr)
			}
			want, err := spreadOverwriteLocation(path, pin.at)
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err == nil {
				native, _ := natively(t, program)
				t.Fatalf("lowered; want a stop at %s saying %q. Sanitized native: exit %d, stdout %q, stderr %q", want, pin.words, native.exitCode, native.stdout, native.stderr)
			}
			var notYet *lower.NotYet
			if !errors.As(err, &notYet) {
				t.Fatalf("want a NotYet, got %v", err)
			}
			if notYet.Where != want || !strings.Contains(err.Error(), pin.words) || pin.ends && !strings.HasSuffix(notYet.What, pin.words) {
				t.Fatalf("got %q at %s; want %q at %s", err.Error(), notYet.Where, pin.words, want)
			}
		})
	}
}

// A SpreadWith literal's source stored as an array, a tuple here, seen through { readonly length: number }, is the
// ruled run-time NotYet (Adamic on #k7g6f9h, under #7fg2fp3's form): Node copies its indices and runs on, and both
// backends print exactly Node's output up to the spread, then stop, exit 70, with ir.SpreadWithSourceStop whole.
// Native runs sanitized, ASan with UBSan, so a source read as an object's storage fails there loudly instead.
// backstop_plain.ts, a plain object through the same type, agrees (TestSpreadWithAgrees).
func TestSpreadWithSourceBackstop(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/spread_with/backstop_tuple.ts"))
	if err != nil {
		t.Fatal(err)
	}
	const before = "{\"length\":2,\"k\":1}\n"
	truth := onNode(t, path)
	if truth.exitCode != 0 || string(truth.stdout) != before+"{\"0\":1,\"1\":2,\"k\":1}\nend\n" {
		t.Fatalf("Node: exit %d, stdout %q, stderr %q", truth.exitCode, truth.stdout, truth.stderr)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatalf("Lower: %v", err)
	}
	sanitized, _ := natively(t, program)
	stop := "adamic: panic: " + ir.SpreadWithSourceStop + "\n"
	for name, got := range map[string]run{"native": sanitized, "javascript": onJavaScriptBackend(t, program)} {
		if got.exitCode != 70 || string(got.stdout) != before || string(got.stderr) != stop {
			t.Errorf("%s: exit %d, stdout %q, stderr %q; want exit 70, stdout %q, stderr %q", name, got.exitCode, got.stdout, got.stderr, before, stop)
		}
	}
}

// The neighbors whose fields each land in the slot their spread brings keep the copy, as on main: no SpreadWith
// literal, so the census moves only the literals #xy74m5x stopped.
func TestSpreadWithLeavesNeighbors(t *testing.T) {
	t.Parallel()
	for _, file := range []string{"spread_overwrite/neighbors.ts", "spread_overwrite/string_source.ts"} {
		t.Run(file, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", file))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatalf("Lower: %v", err)
			}
			if strings.Contains(native.C(program), "adamic_object_spread_with(") {
				t.Error("a neighbor calls adamic_object_spread_with")
			}
		})
	}
}
