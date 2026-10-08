package oracle

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCheckedViewCallableCodeHigherOrder(t *testing.T) {
	paths, err := filepath.Glob(filepath.Join(repository, "stage3/interface-downcasts/lane5/code/higher-order/*/good.a"))
	if err != nil || len(paths) == 0 {
		t.Fatalf("fixtures: %v, %v", paths, err)
	}
	for _, path := range paths {
		path, err := filepath.Abs(path)
		if err != nil {
			t.Fatal(err)
		}
		t.Run(filepath.Base(filepath.Dir(path)), func(t *testing.T) {
			truth := onNode(t, path)
			want := "completed\n"
			if truth.exitCode != 0 || string(truth.stdout) != want {
				t.Fatalf("source Node: %#v", truth)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			checkedViewCallableFormerUnionControl(t, program, truth)
			stopping, err := os.ReadFile(filepath.Join(filepath.Dir(path), "expected.stderr"))
			if err != nil {
				t.Fatal(err)
			}
			for _, variant := range []string{"wrong-arity", "wrong-callback"} {
				t.Run(variant, func(t *testing.T) {
					negative := filepath.Join(filepath.Dir(path), variant+".a")
					truth = onNode(t, negative)
					if truth.exitCode != 0 || string(truth.stdout) != want {
						t.Fatalf("negative Node: %#v", truth)
					}
					program, err = lowered(t, negative)
					if err != nil {
						t.Fatal(err)
					}
					sanitized, _ := nativelyUncached(t, program)
					for _, got := range []run{sanitized, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
						if got.exitCode != 70 || len(got.stdout) != 0 || string(got.stderr) != string(stopping) {
							t.Fatalf("negative exit %d stdout %q stderr %q", got.exitCode, got.stdout, got.stderr)
						}
						t.Logf("negative: %s", got.stderr)
					}
				})
			}
		})
	}
}

func TestCheckedViewCallableCodeCallbackExecution(t *testing.T) {
	for _, probe := range []struct{ name, want string }{
		{"c-395", "abc\ncompleted\n"},
		{"c-2078", "result\ncompleted\n"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			program, path := interfaceFixture(t, "lane5/code/higher-order/"+probe.name+"/callback-execution")
			truth := onNode(t, path)
			if truth.exitCode != 0 || string(truth.stdout) != probe.want {
				t.Fatalf("source Node: %#v", truth)
			}
			checkedViewCallableFormerUnionControl(t, program, truth)
		})
	}
}
