package oracle

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckedViewMapCertificates(t *testing.T) {
	names := []string{"key-schema", "readonly-covariant", "mutable-invariant", "clone", "boolean"}
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
			mutant := strings.Contains(name, "schema") || strings.Contains(name, "wrong") || strings.Contains(name, "opposite") || name == "mutable-invariant"
			for _, got := range []run{native, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if !mutant {
					if diff := disagreement(truth, got); diff != "" {
						t.Fatalf("%s: %#v", diff, got)
					}
				} else if got.exitCode != 70 || !strings.Contains(string(got.stderr), "node.value") {
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
