package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

func TestCheckedViewSetIntrinsics(t *testing.T) {
	for _, variant := range []string{"add", "has", "identity", "number", "boolean"} {
		t.Run(variant, func(t *testing.T) {
			program, path := interfaceFixture(t, "lane5/later-ranked-callables/set-intrinsic/"+variant)
			truth := onNode(t, path)
			want := "1\n"
			if variant == "has" {
				want = "true\n"
			}
			if variant == "identity" {
				want = "true/true/1\n1\n"
			}
			if variant == "number" || variant == "boolean" {
				want = "true/true/1\n"
			}
			if truth.exitCode != 0 || string(truth.stdout) != want {
				t.Fatalf("Node %#v", truth)
			}
			checkedViewCallableFormerUnionControl(t, program, truth)
		})
	}
}

// Every probe in this directory must compile under the actual production flags.
func TestCheckedViewSetIntrinsicsC(t *testing.T) {
	paths, err := filepath.Glob(filepath.Join(repository, "stage3/interface-downcasts/lane5/later-ranked-callables/set-intrinsic/*.a"))
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) < 2 {
		t.Fatal("missing Set probes")
	}
	if !strings.Contains(strings.Join(native.Flags(native.Options{}), " "), "-Werror") {
		t.Fatal("production flags lost -Werror")
	}
	for _, path := range paths {
		t.Run(filepath.Base(path), func(t *testing.T) {
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			if err := native.Build(native.C(program), filepath.Join(t.TempDir(), "probe"), native.Options{}); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestCheckedViewSetIntrinsicSignatures(t *testing.T) {
	for _, variant := range []string{"wrong-parameter", "wrong-result"} {
		t.Run(variant, func(t *testing.T) {
			program, path := interfaceFixture(t, "lane5/later-ranked-callables/set-intrinsic/"+variant)
			truth := onNode(t, path)
			if truth.exitCode != 0 {
				t.Fatalf("Node %#v", truth)
			}
			sanitized, _ := nativelyUncached(t, program)
			for _, got := range []run{sanitized, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				found := "incompatible parameter representations"
				if variant == "wrong-result" {
					found = "incompatible result representation"
				}
				if got.exitCode != 70 || len(got.stdout) != 0 || !strings.Contains(string(got.stderr), "viewed.has") || !strings.Contains(string(got.stderr), "expected") || !strings.Contains(string(got.stderr), found) {
					t.Fatalf("signature refusal %#v", got)
				}
			}
		})
	}
}

// This unit can refresh its rows even when unrelated inherited fixtures fail.
func TestCheckedViewSetIntrinsicCounts(t *testing.T) {
	paths, err := filepath.Glob(filepath.Join(repository, "stage3/interface-downcasts/lane5/later-ranked-callables/set-intrinsic/*.a"))
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(repository, "internal/oracle/counts.md")
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSuffix(string(contents), "\n"), "\n")
	for _, fixture := range paths {
		relative, err := filepath.Rel(repository, fixture)
		if err != nil {
			t.Fatal(err)
		}
		row := counted(t, relative, false, nil, false, false)
		if !*updateCounts {
			if !strings.Contains(string(contents), row+"\n") {
				t.Errorf("unrecorded Set intrinsic counts: %s", row)
			}
			continue
		}
		key := strings.Split(row, " | ")[0] + " | "
		replaced := false
		for i, line := range lines {
			if strings.HasPrefix(line, key) {
				lines[i] = row
				replaced = true
				break
			}
		}
		if !replaced {
			lines = append(lines, row)
		}
	}
	if *updateCounts {
		if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0644); err != nil {
			t.Fatal(err)
		}
	}
}
