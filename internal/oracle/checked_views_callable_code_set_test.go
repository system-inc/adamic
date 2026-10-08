package oracle

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCheckedViewCallableCodeSetDomains(t *testing.T) {
	program, path := interfaceFixture(t, "lane5/code/set-domains/good")
	truth := onNode(t, path)
	if truth.exitCode != 0 || string(truth.stdout) != "completed\ncompleted\n" {
		t.Fatalf("source Node: %#v", truth)
	}
	checkedViewCallableFormerUnionControl(t, program, truth)
	program, path = interfaceFixture(t, "lane5/code/set-domains/wrong-element")
	truth = onNode(t, path)
	if truth.exitCode != 0 || string(truth.stdout) != "completed\n" {
		t.Fatalf("negative source Node: %#v", truth)
	}
	sanitized, _ := nativelyUncached(t, program)
	stopping, err := os.ReadFile(filepath.Join(filepath.Dir(path), "expected.stderr"))
	if err != nil {
		t.Fatal(err)
	}
	for _, got := range []run{sanitized, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
		if got.exitCode != 70 || len(got.stdout) != 0 || string(got.stderr) != string(stopping) {
			t.Fatalf("negative exit %d stdout %q stderr %q", got.exitCode, got.stdout, got.stderr)
		}
	}
}

func TestCheckedViewCallableCodeSetPrimitivePairs(t *testing.T) {
	paths, err := filepath.Glob(filepath.Join(repository, "stage3/interface-downcasts/lane5/code/set-primitive/*/good.a"))
	if err != nil || len(paths) != 8 {
		t.Fatalf("fixtures: %v, %v", paths, err)
	}
	for _, path := range paths {
		t.Run(filepath.Base(filepath.Dir(path)), func(t *testing.T) {
			path, err := filepath.Abs(path)
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			if truth.exitCode != 0 {
				t.Fatalf("source Node: %#v", truth)
			}
			checkedViewCallableFormerUnionControl(t, program, truth)
			negative := filepath.Join(filepath.Dir(path), "wrong-element.a")
			truth = onNode(t, negative)
			if truth.exitCode != 0 {
				t.Fatalf("negative Node: %#v", truth)
			}
			program, err = lowered(t, negative)
			if err != nil {
				t.Fatal(err)
			}
			sanitized, _ := nativelyUncached(t, program)
			stopping, err := os.ReadFile(filepath.Join(filepath.Dir(path), "expected.stderr"))
			if err != nil {
				t.Fatal(err)
			}
			for _, got := range []run{sanitized, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if got.exitCode != 70 || len(got.stdout) != 0 || string(got.stderr) != string(stopping) {
					t.Fatalf("negative exit %d stdout %q stderr %q", got.exitCode, got.stdout, got.stderr)
				}
				t.Logf("negative: %s", got.stderr)
			}
		})
	}
}

func TestCheckedViewCallableCodeSetInstantiations(t *testing.T) {
	program, path := interfaceFixture(t, "lane5/code/set-domains/generic-instantiations")
	truth := onNode(t, path)
	if truth.exitCode != 0 || string(truth.stdout) != "completed\ncompleted\n" {
		t.Fatalf("source Node: %#v", truth)
	}
	stopping, err := os.ReadFile(filepath.Join(filepath.Dir(path), "expected.stderr"))
	if err != nil {
		t.Fatal(err)
	}
	sanitized, _ := nativelyUncached(t, program)
	for _, got := range []run{sanitized, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
		if got.exitCode != 70 || string(got.stdout) != "completed\n" || string(got.stderr) != string(stopping) {
			t.Fatalf("generic domain exit %d stdout %q stderr %q", got.exitCode, got.stdout, got.stderr)
		}
	}
}
