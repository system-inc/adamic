package oracle

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckedViewMapCertificates(t *testing.T) {
	names := []string{"entry-nominal-nested-producer", "entry-nominal-nested-tuple", "entry-nominal-nested-schema", "entry-nominal-nested-field", "entry-nominal-nested-null", "entry-nominal-nested-undefined", "entry-nominal-nested-both", "entry-nominal-nested", "entry-nominal-gap-nested-read", "entry-nominal-gap-nested-unread", "entry-convert-key-packed-boolean", "entry-convert-key-boxed-mixed", "entry-convert-key-boxed-boolean", "entry-convert-key-boxed-string", "entry-convert-key-boxed-optional-number", "entry-convert-key-boxed-optional-boolean", "entry-convert-gap-key-union", "entry-convert-key-boxed", "entry-convert-gap-nested-boolean", "entry-convert-nested-boolean", "entry-convert-nested-boolean-optional", "entry-convert-nested-boolean-boxed", "entry-convert-nested-boolean-finite", "entry-convert-nested-boolean-schema", "entry-convert-gap-nested-union", "entry-convert-nested-union", "entry-convert-nested-union-boolean", "entry-convert-nested-union-string", "entry-convert-nested-union-schema", "entry-convert-nested-union-finite", "entry-convert-nested-union-nullish", "entry-nominal", "entry-nominal-null", "entry-nominal-undefined", "entry-nominal-both", "entry-nominal-schema", "entry-nominal-key", "entry-nominal-derived", "entry-nominal-generic", "entry-nominal-nullish-storage", "key-schema", "readonly-covariant", "mutable-invariant", "clone", "boolean", "structural", "structural-schema", "structural-covariant", "structural-payload-wrong", "structural-mutable-invariant", "object-key", "object-key-schema", "structural-schema-unused-payload", "optional-number", "array", "array-schema", "array-payload-wrong", "array-covariant", "array-mutable-invariant", "array-mutable", "array-schema-unused-payload", "array-recursive", "node-array", "node-array-own-schema", "node-array-readonly-schema", "node-array-covariant", "node-array-own-schema-unused-payload", "optional-array", "optional-array-schema", "optional-object", "entry-mixed-schema", "entry-null-schema", "entry-undefined-schema", "entry-boolean-undefined", "entry-lifecycle", "entry-live-mutation", "entry-callable-plain", "entry-callable-null", "entry-callable-undefined", "entry-callable-both", "entry-callable-result-schema", "entry-callable-parameter-schema", "entry-callable-contravariant", "entry-callable-clone", "entry-tuple", "entry-tuple-schema", "entry-tuple-covariant", "entry-tuple-mutable-invariant", "entry-tuple-length-schema", "entry-tuple-readonly-schema", "entry-tuple-nullish", "entry-brand-string", "entry-brand-number", "entry-brand-boolean", "entry-brand-array", "entry-brand-nullish", "entry-brand-nullish-effects", "entry-convert-owned-get", "entry-convert-owned-callback", "entry-brand-schema", "entry-convert-number-null", "entry-convert-number-undefined", "entry-convert-boolean-null", "entry-convert-boolean-undefined", "entry-convert-string-null", "entry-convert-object-null", "entry-convert-array-null", "entry-convert-tuple-null", "entry-convert-callable-null", "entry-convert-false-null", "entry-convert-false-undefined", "entry-convert-snapshot", "entry-convert-iterator", "entry-convert-maybe-number", "entry-convert-maybe-boolean", "entry-convert-nan-null", "entry-convert-nan-undefined", "entry-convert-negative-zero", "entry-convert-boolean-iterator", "entry-convert-boolean-mixed", "entry-convert-optional-object", "entry-convert-optional-object-schema", "entry-convert-optional-array", "entry-convert-optional-array-schema", "entry-convert-optional-string", "entry-convert-optional-string-schema", "entry-convert-optional-callable", "entry-convert-optional-callable-schema", "entry-convert-gap-nested-array", "entry-convert-nested-array", "entry-convert-nested-array-deep", "entry-convert-nested-array-schema", "entry-convert-nested-array-mutable-invariant", "entry-convert-gap-key", "entry-convert-key", "entry-convert-key-schema", "entry-convert-key-mutable-invariant", "entry-convert-gap-optional-object", "entry-convert-gap-optional-array", "entry-convert-schema", "entry-convert-mutable-invariant"}
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

func TestCheckedViewMapApprovedPhantom(t *testing.T) {
	program, path := interfaceFixture(t, "nullish/maps/phantom-refused")
	truth := onNode(t, path)
	native, binary := nativelyUncached(t, program)
	for _, got := range []run{native, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
		if diff := disagreement(truth, got); diff != "" {
			t.Fatal(diff)
		}
	}
	if report := leaks(t, program, binary); report != "" {
		t.Fatal(report)
	}
	t.Logf("Node stdout=%q; approved void brand has no runtime nominal demand", truth.stdout)
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

func TestCheckedViewMapApprovedBrandBoundaries(t *testing.T) {
	for _, family := range []string{"nested-array-brand", "nested-object-brand"} {
		for _, use := range []string{"read", "unread"} {
			t.Run(family+"-"+use, func(t *testing.T) {
				path, err := filepath.Abs(filepath.Join(repository, "stage3/interface-downcasts/nullish/maps/"+family+"-"+use+".a"))
				if err != nil {
					t.Fatal(err)
				}
				truth := onNode(t, path)
				program, err := lowered(t, path)
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

func TestCheckedViewMapStorageGaps(t *testing.T) {
	for _, family := range []string{"optional-tuple", "rest-tuple"} {
		t.Run(family, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "stage3/interface-downcasts/nullish/maps/entry-convert-gap-"+family+".a"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			truth := onNode(t, path)
			if err != nil {
				if !strings.Contains(err.Error(), "field value") || !strings.Contains(err.Error(), "Map") {
					t.Fatalf("gap refused away from demanded read: %v", err)
				}
				t.Logf("Node=%q; named compile refusal: %v", truth.stdout, err)
				return
			}
			native, _ := nativelyUncached(t, program)
			for _, got := range []run{native, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if got.exitCode != 70 || !strings.Contains(string(got.stderr), "node.value") {
					t.Fatalf("unsupported conversion ran on: %#v", got)
				}
			}
			t.Logf("Node=%q; unsupported conversion stops at node.value", truth.stdout)
		})
	}
}

func TestCheckedViewMapNominalGaps(t *testing.T) {
	for _, family := range []string{"method", "array", "mutable", "recursive"} {
		for _, use := range []string{"read", "unread"} {
			t.Run(family+"-"+use, func(t *testing.T) {
				path, err := filepath.Abs(filepath.Join(repository, "stage3/interface-downcasts/nullish/maps/entry-nominal-gap-"+family+"-"+use+".a"))
				if err != nil {
					t.Fatal(err)
				}
				truth := onNode(t, path)
				program, err := lowered(t, path)
				if use == "read" {
					if err == nil || !strings.Contains(err.Error(), "field value") || !strings.Contains(err.Error(), "Map") {
						t.Fatalf("unsupported nominal family escaped demanded read: %v", err)
					}
					t.Logf("Node=%q; named read refusal=%v", truth.stdout, err)
					return
				}
				if err != nil {
					t.Fatal(err)
				}
				native, binary := nativelyUncached(t, program)
				for _, got := range []run{native, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
					if diff := disagreement(truth, got); diff != "" {
						t.Fatal(diff)
					}
				}
				if report := leaks(t, program, binary); report != "" {
					t.Fatal(report)
				}
				t.Logf("Node=%q; unread nominal gap admitted", truth.stdout)
			})
		}
	}
}

func TestCheckedViewMapRequiredBrandRefusal(t *testing.T) {
	for _, use := range []string{"read", "unread"} {
		t.Run(use, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "stage3/interface-downcasts/nullish/maps/entry-nominal-gap-required-brand-"+use+".a"))
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			_, err = lowered(t, path)
			if err == nil || !strings.Contains(err.Error(), "primitive brand member brand whose type is not void") {
				t.Fatalf("required primitive brand language boundary changed: %v", err)
			}
			t.Logf("Node=%q; existing language refusal at type declaration=%v", truth.stdout, err)
		})
	}
}

func TestCheckedViewArrayJSONStorageRefusal(t *testing.T) {
	for _, family := range []string{"boxed", "packed"} {
		t.Run(family, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "stage3/interface-downcasts/nullish/maps/entry-array-json-"+family+"-read.a"))
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			_, err = lowered(t, path)
			if err == nil || !strings.Contains(err.Error(), "JSON.stringify a boxed or packed boolean array requiring checked element conversion") {
				t.Fatalf("JSON bypassed the checked extraction boundary: %v", err)
			}
			t.Logf("Node=%q; named consumer refusal=%v", truth.stdout, err)
		})
	}
}
