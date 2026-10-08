package oracle

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckedViewMapCertificates(t *testing.T) {
	names := []string{"key-schema", "readonly-covariant", "mutable-invariant", "clone", "boolean", "structural", "structural-schema", "structural-covariant", "structural-payload-wrong", "structural-mutable-invariant", "object-key", "object-key-schema", "structural-schema-unused-payload", "optional-number", "array", "array-schema", "array-payload-wrong", "array-covariant", "array-mutable-invariant", "array-mutable", "array-schema-unused-payload", "array-recursive", "node-array", "node-array-own-schema", "node-array-readonly-schema", "node-array-covariant", "node-array-own-schema-unused-payload", "optional-array", "optional-array-schema", "optional-object", "entry-mixed-schema", "entry-null-schema", "entry-undefined-schema", "entry-boolean-undefined", "entry-lifecycle", "entry-live-mutation", "entry-callable-plain", "entry-callable-null", "entry-callable-undefined", "entry-callable-both", "entry-callable-result-schema", "entry-callable-parameter-schema", "entry-callable-contravariant", "entry-callable-clone", "entry-tuple", "entry-tuple-schema", "entry-tuple-covariant", "entry-tuple-mutable-invariant", "entry-tuple-length-schema", "entry-tuple-readonly-schema", "entry-tuple-nullish"}
	for _, variant := range []string{"null", "undefined", "both"} {
		for _, mutation := range []string{"", "-wrong", "-value-schema", "-opposite"} {
			if variant == "both" && mutation == "-opposite" {
				continue
			}
			names = append(names, "scalar-"+variant+mutation)
		}
	}
	for _, variant := range []string{"null", "undefined", "both"} {
		names = append(names, "mixed-"+variant, "mixed-"+variant+"-schema")
	}
	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			program, path := interfaceFixture(t, "nullish/maps/"+name)
			truth := onNode(t, path)
			native, binary := nativelyUncached(t, program)
			mutant := strings.Contains(name, "schema") || strings.Contains(name, "wrong") || strings.Contains(name, "opposite") || strings.Contains(name, "mutable-invariant")
			for _, got := range []run{native, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if !mutant {
					if diff := disagreement(truth, got); diff != "" {
						t.Fatalf("%s: %#v", diff, got)
					}
				} else if got.exitCode != 70 || !strings.Contains(string(got.stderr), mapReadField(name)) {
					t.Fatalf("map certificate mutant ran on: %#v", got)
				}
			}
			if !mutant {
				if report := leaks(t, program, binary); report != "" {
					t.Fatal(report)
				}
			}
			t.Logf("Node exit=%d stdout=%q; mutant=%t", truth.exitCode, truth.stdout, mutant)
		})
	}
}

func TestCheckedViewMapCertificateWrites(t *testing.T) {
	for _, name := range []string{"write-good", "write-wrong", "unknown"} {
		t.Run(name, func(t *testing.T) {
			if name == "write-wrong" {
				path, err := filepath.Abs(filepath.Join(repository, "stage3/interface-downcasts/nullish/maps/write-wrong.a"))
				if err != nil {
					t.Fatal(err)
				}
				_, err = lowered(t, path)
				truth := onNode(t, path)
				if err == nil || !strings.Contains(err.Error(), "invariant-mutable") {
					t.Fatalf("widened writable alias escaped: %v", err)
				}
				t.Logf("Node stdout=%q; caught existing invariant check: %v", truth.stdout, err)
				return
			}
			program, path := interfaceFixture(t, "nullish/maps/"+name)
			truth := onNode(t, path)
			native, binary := nativelyUncached(t, program)
			for _, got := range []run{native, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if name == "write-good" {
					if diff := disagreement(truth, got); diff != "" {
						t.Fatalf("%s %#v", diff, got)
					}
				} else if got.exitCode != 70 || !strings.Contains(string(got.stderr), "Map contract failed") {
					t.Fatalf("certificate mutant ran on: %#v", got)
				}
			}
			if name == "write-good" {
				if report := leaks(t, program, binary); report != "" {
					t.Fatal(report)
				}
			}
			t.Logf("Node exit=%d stdout=%q", truth.exitCode, truth.stdout)
		})
	}
}

func TestCheckedViewMapPhantomRefusal(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "stage3/interface-downcasts/nullish/maps/phantom-refused.a"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = lowered(t, path)
	truth := onNode(t, path)
	if err == nil || !strings.Contains(err.Error(), "field value") || !strings.Contains(err.Error(), "Map") {
		t.Fatalf("phantom certificate fabricated: %v", err)
	}
	t.Logf("Node stdout=%q; caught unsupported brand at read: %v", truth.stdout, err)
}

func mapReadField(name string) string {
	if name == "array-payload-wrong" {
		return "values[0]"
	}
	if name == "structural-payload-wrong" {
		return "entry.count"
	}
	return "node.value"
}

func TestCheckedViewMapEntryFamilyBoundaries(t *testing.T) {
	for _, family := range []string{"nested-array-brand", "nested-object-brand"} {
		for _, use := range []string{"read", "unread"} {
			t.Run(family+"-"+use, func(t *testing.T) {
				path, err := filepath.Abs(filepath.Join(repository, "stage3/interface-downcasts/nullish/maps/"+family+"-"+use+".a"))
				if err != nil {
					t.Fatal(err)
				}
				truth := onNode(t, path)
				program, err := lowered(t, path)
				if use == "read" {
					if err == nil || !strings.Contains(err.Error(), "field value") || !strings.Contains(err.Error(), "Map key/value certificate") {
						t.Fatalf("unsupported entry escaped or refused away from read: %v", err)
					}
					t.Logf("Node stdout=%q; named read refusal: %v", truth.stdout, err)
					return
				}
				if err != nil {
					t.Fatal(err)
				}
				native, binary := nativelyUncached(t, program)
				for _, got := range []run{native, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
					if diff := disagreement(truth, got); diff != "" {
						t.Fatalf("%s: %#v", diff, got)
					}
				}
				if report := leaks(t, program, binary); report != "" {
					t.Fatal(report)
				}
			})
		}
	}
}

func TestCheckedViewMapUnionEntryStorage(t *testing.T) {
	for _, family := range []string{"number", "boolean", "string", "object", "array", "mixed"} {
		for _, form := range []string{"null", "both"} {
			t.Run(family+"-"+form, func(t *testing.T) {
				program, path := interfaceFixture(t, "nullish/maps/entry-"+family+"-"+form)
				truth := onNode(t, path)
				native, binary := nativelyUncached(t, program)
				for _, got := range []run{native, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
					if diff := disagreement(truth, got); diff != "" {
						t.Fatalf("%s: %#v", diff, got)
					}
				}
				if report := leaks(t, program, binary); report != "" {
					t.Fatal(report)
				}
				t.Logf("Node stdout=%q", truth.stdout)
			})
		}
	}
}
