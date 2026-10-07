package decisions

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// Not parallel: sequential sanitized builds bound compiler memory across 17 controls.
func TestNamedListenerDeclarations(t *testing.T) {
	root, err := filepath.Abs("../../../../..")
	if err != nil {
		t.Fatal(err)
	}
	owned, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	cohere, err := filepath.EvalSymlinks(filepath.Join(root, "cohere"))
	if err != nil {
		t.Fatal(err)
	}
	virtual := filepath.Join(cohere, "adamic_wave12_listeners.go")
	tailwind := filepath.Join(cohere, "internal/lint/rules/tailwind/adamic_wave12_listeners.go")
	overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{
		virtual:  filepath.Join(owned, "listeners-oracle.go.txt"),
		tailwind: filepath.Join(owned, "listeners-tailwind-overlay.go.txt"),
	}})
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	overlayPath := filepath.Join(directory, "overlay.json")
	if err := os.WriteFile(overlayPath, overlay, 0644); err != nil {
		t.Fatal(err)
	}
	oracle := filepath.Join(directory, "oracle")
	run(t, cohere, "go", "build", "-overlay="+overlayPath, "-o", oracle, virtual)
	want := run(t, "", oracle)
	rows := strings.Split(strings.TrimSpace(string(want)), "\n")
	if len(rows) != 17 {
		t.Fatalf("expected 17 observed rules, got %d", len(rows))
	}
	for side, got := range outputs(t, root, filepath.Join(owned, "listeners-main.a")) {
		if !bytes.Equal(got, want) {
			t.Fatalf("named listener domain differs from actual Go rule listeners on backend %d: got %s, want %s", side, got, want)
		}
	}
	t.Logf("17 named listener declarations, %d bytes, match actual Go listener keys on source Node, emitted JavaScript and ASan/UBSan native", len(want))
	entry, err := os.ReadFile(filepath.Join(owned, "listeners-main.a"))
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range rows {
		fields := strings.Split(row, "\t")
		if len(fields) != 2 {
			t.Fatalf("invalid Go row %q", row)
		}
		descriptorPath := filepath.Join(owned, "..", fields[0], "rule.json")
		if data, err := os.ReadFile(descriptorPath); err == nil {
			var descriptor struct{ Kinds []string }
			if err := json.Unmarshal(data, &descriptor); err != nil {
				t.Fatal(err)
			}
			actual := strings.Split(fields[1], ",")
			sort.Strings(actual)
			sort.Strings(descriptor.Kinds)
			if strings.Join(actual, ",") != strings.Join(descriptor.Kinds, ",") {
				t.Fatalf("%s rule.json kinds differ from actual Go: got %v want %v", fields[0], descriptor.Kinds, actual)
			}
		} else if !os.IsNotExist(err) {
			t.Fatal(err)
		}
		slug := fields[0]
		t.Run(slug, func(t *testing.T) {
			mutation := t.TempDir()
			for _, original := range rows {
				name := strings.Split(original, "\t")[0]
				target := filepath.Join(mutation, name)
				if err := os.Mkdir(target, 0755); err != nil {
					t.Fatal(err)
				}
				source, err := os.ReadFile(filepath.Join(owned, "..", name, "listener.a"))
				if err != nil {
					t.Fatal(err)
				}
				if name == slug {
					first := strings.Split(fields[1], ",")[0]
					from := "= [" + strconv.Quote(first)
					if strings.Count(string(source), from) != 1 {
						t.Fatal("listener mutant anchor changed")
					}
					// Unknown is a valid kind name and compiles; only the external domain comparison catches it.
					source = []byte(strings.Replace(string(source), from, "= [\"Unknown\"", 1))
				}
				if err := os.WriteFile(filepath.Join(target, "listener.a"), source, 0644); err != nil {
					t.Fatal(err)
				}
			}
			rewritten := strings.ReplaceAll(string(entry), "'./listener.a'", "'./"+filepath.Base(owned)+"/listener.a'")
			rewritten = strings.ReplaceAll(rewritten, "'../", "'./")
			main := filepath.Join(mutation, "main.a")
			if err := os.WriteFile(main, []byte(rewritten), 0644); err != nil {
				t.Fatal(err)
			}
			for side, got := range outputs(t, root, main) {
				if bytes.Equal(got, want) {
					t.Fatalf("compiling listener mutant survived on backend %d", side)
				}
				t.Logf("named listener mutant compiles and exits cleanly, caught only by actual-Go comparison on backend %d", side)
			}
		})
	}
}
