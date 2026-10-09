package oracle

import (
	"context"
	"encoding/json"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckedViewUntaggedSourceDispatch(t *testing.T) {
	t.Parallel()
	path, _ := filepath.Abs("../../stage3/interface-downcasts/untagged/fixtures/binding-name-source-good.a")
	loaded, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	_, err = lower.Lower(context.Background(), loaded)
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile("../../stage3/interface-downcasts/untagged/source-refusals.json")
	if err != nil {
		t.Fatal(err)
	}
	pins := map[string]string{}
	if err = json.Unmarshal(data, &pins); err != nil {
		t.Fatal(err)
	}
	for _, family := range []string{"binding-name", "option-element", "structural"} {
		for _, variant := range []string{"good", "wrong", "nested", "absent"} {
			t.Run(family+"/"+variant, func(t *testing.T) {
				t.Parallel()
				if variant == "absent" {
					path, _ := filepath.Abs("../../stage3/interface-downcasts/untagged/fixtures/" + family + "-source-" + variant + ".a")
					loaded, err := load.Load([]string{path})
					if err != nil {
						t.Fatal(err)
					}
					_, err = lower.Lower(context.Background(), loaded)
					if err == nil || !strings.Contains(err.Error(), "optional field value") {
						t.Fatalf("base optional widening boundary changed: %v", err)
					}
					if got := onNode(t, path); got.exitCode != 0 {
						t.Fatalf("Node control: %#v", got)
					}
					return
				}
				program, path := interfaceFixture(t, "untagged/fixtures/"+family+"-source-"+variant)
				// untagged source mutation anchor
				node := onNode(t, path)
				t.Logf("Node: exit=%d stdout=%q stderr=%q", node.exitCode, node.stdout, node.stderr)
				for backend, got := range map[string]run{"native": releasedUncached(t, program), "native-sanitized": func() run { got, _ := nativelyUncached(t, program); return got }(), "javascript": onJavaScriptBackend(t, program)} {
					t.Logf("%s: exit=%d stdout=%q stderr=%q", backend, got.exitCode, got.stdout, got.stderr)
					if variant == "good" || variant == "absent" || variant == "empty" {
						if difference := disagreement(node, got); difference != "" {
							t.Fatal(backend + ": " + difference)
						}
					} else {
						expected, ok := pins[family+"/"+variant]
						if !ok {
							t.Fatal("missing refusal pin")
						}
						if difference := disagreement(run{exitCode: 70, stderr: []byte(expected)}, got); difference != "" {
							t.Errorf("%s: %s; got %#v", backend, difference, got)
						}
					}
				}
			})
		}
	}
}

// Fault injection at lowered source reads, using the same IR consumed by both backends.
func TestCheckedViewUntaggedCallableUnion(t *testing.T) {
	t.Parallel()
	for _, variant := range []string{"good-number", "good-string", "wrong", "nested"} {
		t.Run(variant, func(t *testing.T) {
			t.Parallel()
			program, path := interfaceFixture(t, "untagged/fixtures/callable-union-"+variant)
			node := onNode(t, path)
			if difference := disagreement(run{stdout: []byte("true\n")}, node); difference != "" {
				t.Fatal(difference)
			}
			for backend, got := range map[string]run{"native": releasedUncached(t, program), "native-sanitized": func() run { got, _ := nativelyUncached(t, program); return got }(), "javascript": onJavaScriptBackend(t, program)} {
				t.Logf("%s: exit=%d stdout=%q stderr=%q", backend, got.exitCode, got.stdout, got.stderr)
				if variant != "wrong" && variant != "nested" {
					if difference := disagreement(node, got); difference != "" {
						t.Errorf("%s: %s", backend, difference)
					}
				} else {
					field := "view.value"
					if variant == "nested" {
						field = "view.holder.value"
					}
					want := run{exitCode: 70, stderr: []byte("adamic: panic: field read failed: " + field + " expected Target, found function with incompatible parameter representations\n")}
					if difference := disagreement(want, got); difference != "" {
						t.Errorf("%s: %s", backend, difference)
					}
				}
			}
		})
	}
}

// These receipts keep remaining adapter boundaries distinct from completed
// candidates. Their valid Node controls do not certify Adamic support.
func TestCheckedViewUntaggedOptionalCallableControl(t *testing.T) {
	t.Parallel()
	program, path := interfaceFixture(t, "untagged/fixtures/callable-union-optional-boundary")
	node := onNode(t, path)
	if difference := disagreement(run{stdout: []byte("true\n")}, node); difference != "" {
		t.Fatal(difference)
	}
	for backend, got := range map[string]run{"native": releasedUncached(t, program), "native-sanitized": func() run { got, _ := nativelyUncached(t, program); return got }(), "javascript": onJavaScriptBackend(t, program)} {
		if difference := disagreement(node, got); difference != "" {
			t.Errorf("%s: %s", backend, difference)
		}
	}
}

func TestCheckedViewUntaggedSourceFlows(t *testing.T) {
	t.Parallel()
	for _, flow := range []string{"helper", "generic", "callback", "stored"} {
		for _, variant := range []string{"good", "wrong"} {
			t.Run(flow+"/"+variant, func(t *testing.T) {
				t.Parallel()
				program, path := interfaceFixture(t, "untagged/fixtures/flow-"+flow+"-"+variant)
				want := run{stdout: []byte("true\n")}
				if difference := disagreement(want, onNode(t, path)); difference != "" {
					t.Fatal("Node: " + difference)
				}
				if variant == "wrong" {
					want = run{exitCode: 70, stderr: []byte("adamic: panic: field read failed: holder.value matches no member of Target; expected Target, found object\n")}
				}
				for backend, got := range map[string]run{"native": releasedUncached(t, program), "native-sanitized": func() run { got, _ := nativelyUncached(t, program); return got }(), "javascript": onJavaScriptBackend(t, program)} {
					t.Logf("%s: exit=%d stdout=%q stderr=%q", backend, got.exitCode, got.stdout, got.stderr)
					if difference := disagreement(want, got); difference != "" {
						t.Errorf("%s: %s", backend, difference)
					}
				}
			})
		}
	}
}

func TestCheckedViewUntaggedOwnClassData(t *testing.T) {
	t.Parallel()
	for _, variant := range []string{"good", "wrong"} {
		t.Run(variant, func(t *testing.T) {
			t.Parallel()
			program, path := interfaceFixture(t, "untagged/fixtures/class-data-"+variant)
			node := onNode(t, path)
			if difference := disagreement(run{stdout: []byte("true\n")}, node); difference != "" {
				t.Fatal(difference)
			}
			want := node
			if variant != "good" {
				want = run{exitCode: 70, stderr: []byte("adamic: panic: field read failed: view.value matches no member of Target; expected Target, found object\n")}
			}
			for backend, got := range map[string]run{"native": releasedUncached(t, program), "native-sanitized": func() run { got, _ := nativelyUncached(t, program); return got }(), "javascript": onJavaScriptBackend(t, program)} {
				t.Logf("%s: exit=%d stdout=%q stderr=%q", backend, got.exitCode, got.stdout, got.stderr)
				if difference := disagreement(want, got); difference != "" {
					t.Errorf("%s: %s", backend, difference)
				}
			}
		})
	}
}

func TestCheckedViewUntaggedRecursive(t *testing.T) {
	t.Parallel()
	for _, variant := range []string{"good", "absent", "wrong", "nested"} {
		t.Run(variant, func(t *testing.T) {
			t.Parallel()
			program, path := interfaceFixture(t, "untagged/fixtures/recursive-"+variant)
			node := onNode(t, path)
			if difference := disagreement(run{stdout: []byte("true\n")}, node); difference != "" {
				t.Fatal(difference)
			}
			want := node
			if variant == "wrong" || variant == "nested" || variant == "cycle-wrong" {
				want = run{exitCode: 70, stderr: []byte("adamic: panic: field read failed: view.value matches no member of Target; expected Target, found object\n")}
			}
			for backend, got := range map[string]run{"native": releasedUncached(t, program), "native-sanitized": func() run { got, _ := nativelyUncached(t, program); return got }(), "javascript": onJavaScriptBackend(t, program)} {
				t.Logf("%s: exit=%d stdout=%q stderr=%q", backend, got.exitCode, got.stdout, got.stderr)
				if difference := disagreement(want, got); difference != "" {
					t.Errorf("%s: %s", backend, difference)
				}
			}
		})
	}
}

func TestCheckedViewUntaggedArrayPending(t *testing.T) {
	t.Parallel()
	for _, variant := range []string{"good", "wrong", "nested", "empty", "mixed"} {
		t.Run(variant, func(t *testing.T) {
			t.Parallel()
			path, _ := filepath.Abs("../../stage3/interface-downcasts/untagged/fixtures/array-union-" + variant + ".a")
			loaded, err := load.Load([]string{path})
			if err != nil {
				t.Fatal(err)
			}
			_, err = lower.Lower(context.Background(), loaded)
			if unsupported, ok := err.(*lower.NotYet); !ok || !(strings.Contains(unsupported.What, "views-v3: array element kind") || variant == "empty" && unsupported.What == "an array of never") {
				t.Fatalf("array admission must stay NotYet: %v", err)
			}
			// Array membership needs V3 element-kind metadata and hole-aware reads.
			t.Skip("awaits compiler/views-v3: array element kind and holes (b065fa576)")
		})
	}
}
