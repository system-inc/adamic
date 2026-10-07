package oracle

import (
	"context"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckedViewCallables(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, stdout string }{{"function-field", "8\n"}, {"class-method", "10\n"}} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			program, path := interfaceFixture(t, "lane2/"+probe.name)
			want := run{stdout: []byte(probe.stdout)}
			if difference := disagreement(want, onNode(t, path)); difference != "" {
				t.Fatal("Node: " + difference)
			}
			actual, _ := nativelyUncached(t, program)
			for _, got := range []run{actual, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if difference := disagreement(want, got); difference != "" {
					t.Fatalf("%s; got %#v", difference, got)
				}
			}
		})
	}
}

func TestCheckedViewOpaqueSignature(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs("../../stage3/interface-downcasts/lane2/opaque-signature.a")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	_, err = lower.Lower(context.Background(), program)
	if err == nil || !strings.Contains(err.Error(), "Adamic 0.1 refuses a checked view call to member Runner.run; its signature cannot be checked at runtime and no compatible implementation is proven;") {
		t.Fatalf("want pinned signature refusal, got %v", err)
	}
}

func TestCheckedViewGenericArrayCast(t *testing.T) {
	t.Parallel()
	program, path := interfaceFixture(t, "lane2/native-generic-array-cast")
	want := run{stdout: []byte("1\n")}
	if difference := disagreement(want, onNode(t, path)); difference != "" {
		t.Fatal(difference)
	}
	actual, _ := nativelyUncached(t, program)
	for _, got := range []run{actual, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
		if difference := disagreement(want, got); difference != "" {
			t.Fatalf("%s; got %#v", difference, got)
		}
	}
}

func TestCheckedViewSameMap(t *testing.T) {
	program, path := interfaceFixture(t, "lane2/same-map")
	want := run{stdout: []byte("true:1:20:3:true\n")}
	if difference := disagreement(want, onNode(t, path)); difference != "" {
		t.Fatal(difference)
	}
	actual, _ := nativelyUncached(t, program)
	for _, got := range []run{actual, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
		if difference := disagreement(want, got); difference != "" {
			t.Fatalf("%s; got %#v", difference, got)
		}
	}
}

func TestCheckedViewParserReturns(t *testing.T) {
	for _, probe := range []struct{ name, stdout string }{{"native-same-map-return-cast", "2\nundefined\n"}, {"native-sorted-empty-conditional", "0\n1\n"}, {"optional-sort", "1,10,9\n1,9,10\nfalse,true,true\na,𝄞,\uE000\n"}} {
		t.Run(probe.name, func(t *testing.T) {
			program, path := interfaceFixture(t, "lane2/"+probe.name)
			want := run{stdout: []byte(probe.stdout)}
			if difference := disagreement(want, onNode(t, path)); difference != "" {
				t.Fatal(difference)
			}
			actual, _ := nativelyUncached(t, program)
			for _, got := range []run{actual, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if difference := disagreement(want, got); difference != "" {
					t.Fatalf("%s; got %#v", difference, got)
				}
			}
		})
	}
}
