package oracle

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckedViewUntaggedSourceDispatch(test *testing.T) {
	test.Parallel()
	path, _ := filepath.Abs("../../stage3/interface-downcasts/untagged/fixtures/binding-name-source-good.ts")
	loaded, loadError := load.Load([]string{path})
	if loadError != nil {
		test.Fatal(loadError)
	}
	_, lowerError := lower.Lower(context.Background(), loaded)
	if lowerError != nil {
		test.Fatal(lowerError)
	}

	data, readError := os.ReadFile("../../stage3/interface-downcasts/untagged/source-refusals.json")
	if readError != nil {
		test.Fatal(readError)
	}
	pins := map[string]string{}
	if decodeError := json.Unmarshal(data, &pins); decodeError != nil {
		test.Fatal(decodeError)
	}
	for _, family := range []string{"binding-name", "option-element", "structural"} {
		for _, variant := range []string{"good", "wrong", "nested", "absent"} {
			test.Run(family+"/"+variant, func(test *testing.T) {
				test.Parallel()
				if variant == "absent" {
					path, _ := filepath.Abs(
						"../../stage3/interface-downcasts/untagged/fixtures/" + family + "-source-" + variant + ".a",
					)
					loaded, loadError := load.Load([]string{path})
					if loadError != nil {
						test.Fatal(loadError)
					}
					_, lowerError := lower.Lower(context.Background(), loaded)
					if lowerError == nil || !strings.Contains(lowerError.Error(), "optional field value") {
						test.Fatalf("base optional widening boundary changed: %v", lowerError)
					}
					if got := onNode(test, path); got.exitCode != 0 {
						test.Fatalf("Node control: %#v", got)
					}
					return
				}
				program, path := interfaceFixture(test, "untagged/fixtures/"+family+"-source-"+variant)
				// untagged source mutation anchor
				node := onNode(test, path)
				test.Logf("Node: exit=%d stdout=%q stderr=%q", node.exitCode, node.stdout, node.stderr)
				for backend, got := range map[string]run{
					"native":           releasedUncached(test, program),
					"native-sanitized": func() run { got, _ := nativelyUncached(test, program); return got }(),
					"javascript":       onJavaScriptBackend(test, program),
				} {
					test.Logf("%s: exit=%d stdout=%q stderr=%q", backend, got.exitCode, got.stdout, got.stderr)
					if variant == "good" || variant == "absent" || variant == "empty" {
						if difference := disagreement(node, got); difference != "" {
							test.Fatal(backend + ": " + difference)
						}
					} else {
						expected, ok := pins[family+"/"+variant]
						if !ok {
							test.Fatal("missing refusal pin")
						}
						if difference := disagreement(
							run{exitCode: 70, stderr: []byte(expected)},
							got,
						); difference != "" {
							test.Errorf("%s: %s; got %#v", backend, difference, got)
						}
					}
				}
			})
		}
	}
}

// Fault injection at lowered source reads, using the same IR consumed by both backends.
func TestCheckedViewUntaggedCallableUnion(test *testing.T) {
	test.Parallel()
	for _, variant := range []string{"good-number", "good-string"} {
		test.Run(variant, func(test *testing.T) {
			test.Parallel()
			program, path := interfaceFixture(test, "untagged/fixtures/callable-union-"+variant)
			node := onNode(test, path)
			if difference := disagreement(run{stdout: []byte("true\n")}, node); difference != "" {
				test.Fatal(difference)
			}
			for backend, got := range map[string]run{
				"native":           releasedUncached(test, program),
				"native-sanitized": func() run { got, _ := nativelyUncached(test, program); return got }(),
				"javascript":       onJavaScriptBackend(test, program),
			} {
				test.Logf("%s: exit=%d stdout=%q stderr=%q", backend, got.exitCode, got.stdout, got.stderr)
				if difference := disagreement(node, got); difference != "" {
					test.Errorf("%s: %s", backend, difference)
				}
			}
		})
	}
}

// A stored function that takes a boolean fits neither member of Target, read directly or through a nested holder, so
// the read is refused where it stands rather than compiled to stop at run time where Node prints true (Compiler's
// ruling (b) on #aecx10a). In .a the cast before it is refused first: its tag test, kind === 'Receiver', is one every
// source value passes (#q97a1rr); its .ts twin keeps the cast a checked view and is refused at the read.
func TestCheckedViewUntaggedCallableUnionMisfitRefuses(test *testing.T) {
	test.Parallel()
	for variant, at := range map[string]struct{ cast, read string }{
		"wrong":  {":7:12:", ":8:16:"},
		"nested": {":8:12:", ":9:16:"},
	} {
		test.Run(variant, func(test *testing.T) {
			test.Parallel()
			path, absolutePathError := filepath.Abs(
				filepath.Join(repository, interfaceFixturePath("untagged/fixtures/callable-union-"+variant)),
			)
			if absolutePathError != nil {
				test.Fatal(absolutePathError)
			}
			if difference := disagreement(run{stdout: []byte("true\n")}, onNode(test, path)); difference != "" {
				test.Fatal(difference)
			}
			_, loweredError := lowered(test, path)
			var refused *lower.Refused
			if !errors.As(loweredError, &refused) ||
				!strings.Contains(loweredError.Error(), "callable-union-"+variant+".a"+at.cast) ||
				!strings.Contains(refused.What, "a downcast to Receiver whose tag test is vacuous") {
				test.Fatalf("want the located vacuous downcast refusal at %s, got %v", at.cast, loweredError)
			}
			_, loweredError = lowered(test, typeScriptTwin(test, path))
			if !errors.As(loweredError, &refused) ||
				!strings.Contains(loweredError.Error(), "callable-union-"+variant+".ts"+at.read) ||
				refused.What != "checked view read of field value with unsupported representation conversion contract" {
				test.Fatalf(
					"twin: want the located representation conversion refusal at %s, got %v",
					at.read,
					loweredError,
				)
			}
		})
	}
}

// These receipts keep remaining adapter boundaries distinct from completed
// candidates. Their valid Node controls do not certify Adamic support.
func TestCheckedViewUntaggedOptionalCallableControl(test *testing.T) {
	test.Parallel()
	program, path := interfaceFixture(test, "untagged/fixtures/callable-union-optional-boundary")
	node := onNode(test, path)
	if difference := disagreement(run{stdout: []byte("true\n")}, node); difference != "" {
		test.Fatal(difference)
	}
	for backend, got := range map[string]run{
		"native":           releasedUncached(test, program),
		"native-sanitized": func() run { got, _ := nativelyUncached(test, program); return got }(),
		"javascript":       onJavaScriptBackend(test, program),
	} {
		if difference := disagreement(node, got); difference != "" {
			test.Errorf("%s: %s", backend, difference)
		}
	}
}

func TestCheckedViewUntaggedSourceFlows(test *testing.T) {
	test.Parallel()
	for _, flow := range []string{"helper", "generic", "callback", "stored"} {
		for _, variant := range []string{"good", "wrong"} {
			test.Run(flow+"/"+variant, func(test *testing.T) {
				test.Parallel()
				program, path := interfaceFixture(test, "untagged/fixtures/flow-"+flow+"-"+variant)
				want := run{stdout: []byte("true\n")}
				if difference := disagreement(want, onNode(test, path)); difference != "" {
					test.Fatal("Node: " + difference)
				}
				if variant == "wrong" {
					want = run{exitCode: 70, stderr: []byte(
						"adamic: panic: field read failed: holder.value matches no member of Target; expected Target, found object\n",
					)}
				}
				for backend, got := range map[string]run{
					"native":           releasedUncached(test, program),
					"native-sanitized": func() run { got, _ := nativelyUncached(test, program); return got }(),
					"javascript":       onJavaScriptBackend(test, program),
				} {
					test.Logf("%s: exit=%d stdout=%q stderr=%q", backend, got.exitCode, got.stdout, got.stderr)
					if difference := disagreement(want, got); difference != "" {
						test.Errorf("%s: %s", backend, difference)
					}
				}
			})
		}
	}
}

func TestCheckedViewUntaggedOwnClassData(test *testing.T) {
	test.Parallel()
	for _, variant := range []string{"good", "wrong"} {
		test.Run(variant, func(test *testing.T) {
			test.Parallel()
			program, path := interfaceFixture(test, "untagged/fixtures/class-data-"+variant)
			node := onNode(test, path)
			if difference := disagreement(run{stdout: []byte("true\n")}, node); difference != "" {
				test.Fatal(difference)
			}
			want := node
			if variant != "good" {
				want = run{exitCode: 70, stderr: []byte(
					"adamic: panic: field read failed: view.value matches no member of Target; expected Target, found object\n",
				)}
			}
			for backend, got := range map[string]run{
				"native":           releasedUncached(test, program),
				"native-sanitized": func() run { got, _ := nativelyUncached(test, program); return got }(),
				"javascript":       onJavaScriptBackend(test, program),
			} {
				test.Logf("%s: exit=%d stdout=%q stderr=%q", backend, got.exitCode, got.stdout, got.stderr)
				if difference := disagreement(want, got); difference != "" {
					test.Errorf("%s: %s", backend, difference)
				}
			}
		})
	}
}

func TestCheckedViewUntaggedRecursive(test *testing.T) {
	test.Parallel()
	for _, variant := range []string{"good", "absent", "wrong", "nested"} {
		test.Run(variant, func(test *testing.T) {
			test.Parallel()
			program, path := interfaceFixture(test, "untagged/fixtures/recursive-"+variant)
			node := onNode(test, path)
			if difference := disagreement(run{stdout: []byte("true\n")}, node); difference != "" {
				test.Fatal(difference)
			}
			want := node
			if variant == "wrong" || variant == "nested" || variant == "cycle-wrong" {
				want = run{exitCode: 70, stderr: []byte(
					"adamic: panic: field read failed: view.value matches no member of Target; expected Target, found object\n",
				)}
			}
			for backend, got := range map[string]run{
				"native":           releasedUncached(test, program),
				"native-sanitized": func() run { got, _ := nativelyUncached(test, program); return got }(),
				"javascript":       onJavaScriptBackend(test, program),
			} {
				test.Logf("%s: exit=%d stdout=%q stderr=%q", backend, got.exitCode, got.stdout, got.stderr)
				if difference := disagreement(want, got); difference != "" {
					test.Errorf("%s: %s", backend, difference)
				}
			}
		})
	}
}

// In .a each cast is refused first, since its tag test, kind === 'Receiver', is one every source value passes
// (#q97a1rr); its .ts twin keeps the cast a checked view and stays NotYet at the array.
func TestCheckedViewUntaggedArrayPending(test *testing.T) {
	test.Parallel()
	for _, variant := range []string{"good", "wrong", "nested", "empty", "mixed"} {
		test.Run(variant, func(test *testing.T) {
			test.Parallel()
			path, _ := filepath.Abs("../../stage3/interface-downcasts/untagged/fixtures/array-union-" + variant + ".a")
			_, loweredError := lowered(test, path)
			if refused, ok := loweredError.(*lower.Refused); !ok ||
				!strings.Contains(refused.What, "a downcast to Receiver whose tag test is vacuous") {
				test.Fatalf("want the vacuous downcast refusal, got %v", loweredError)
			}
			_, loweredError = lowered(test, typeScriptTwin(test, path))
			if unsupported, ok := loweredError.(*lower.NotYet); !ok || !(strings.Contains(
				unsupported.What,
				"views-v3: array element kind",
			) || variant == "empty" && unsupported.What == "an array of never") {
				test.Fatalf("twin: array admission must stay NotYet: %v", loweredError)
			}
			// Array membership needs V3 element-kind metadata and hole-aware reads.
			test.Skip("awaits compiler/views-v3: array element kind and holes (b065fa576)")
		})
	}
}
