package decisions

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func run(t *testing.T, directory, name string, args ...string) []byte {
	t.Helper()
	command := exec.Command(name, args...)
	command.Dir = directory
	output, err := os.CreateTemp(t.TempDir(), "output-")
	if err != nil {
		t.Fatal(err)
	}
	defer output.Close()
	command.Stdout = output
	var stderr bytes.Buffer
	command.Stderr = &stderr
	if err := command.Run(); err != nil || stderr.Len() != 0 {
		t.Fatalf("%s: %v %s", name, err, &stderr)
	}
	data, err := os.ReadFile(output.Name())
	if err != nil {
		t.Fatal(err)
	}
	return data
}
func outputs(t *testing.T, root, probe string) [][]byte {
	t.Helper()
	program, err := load.Load([]string{probe})
	if err != nil {
		t.Fatal(err)
	}
	lowered, err := lower.Lower(context.Background(), program)
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	binary := filepath.Join(directory, "probe")
	if err := native.Build(native.C(lowered), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	emitted := filepath.Join(directory, "probe.mjs")
	if err := os.WriteFile(emitted, []byte(javascript.JavaScript(lowered)), 0644); err != nil {
		t.Fatal(err)
	}
	runner := filepath.Join(root, "oracle/node.mjs")
	return [][]byte{run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, probe), run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, emitted), run(t, "", binary)}
}

// Only extracted decision contracts are certified; not a complete registered rule.
func TestTailwindDecisions(t *testing.T) {
	root, err := filepath.Abs("../../../../..")
	if err != nil {
		t.Fatal(err)
	}
	owned, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	scratch := t.TempDir()
	cohere := filepath.Join(root, "cohere")
	if got := strings.TrimSpace(string(run(t, cohere, "git", "rev-parse", "HEAD"))); got != "715ba94f3608a6500086b1076ce5cb7e51b836db" {
		t.Fatal("Go oracle pin drift", got)
	}
	original, err := os.ReadFile(filepath.Join(cohere, "internal/lint/rules/tailwind/no_conflicting_classes.go"))
	if err != nil {
		t.Fatal(err)
	}
	begin := strings.Index(string(original), "\tresolved := make([]classFacts")
	if begin < 0 {
		t.Fatal("resolver overlay start anchor changed")
	}
	end := strings.Index(string(original)[begin:], "\n\tfindings := ") + begin
	if begin < 0 || end < begin {
		t.Fatal("resolver overlay anchors changed")
	}
	patched := string(original[:begin]) + "\tresolved := wave12Facts\n" + string(original[end:])
	committed, err := os.ReadFile(filepath.Join(owned, "decision-conflicts-overlay.go.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if patched != string(committed) {
		t.Fatal("conflict oracle changed beyond the resolved-facts injection")
	}
	virtual := filepath.Join(cohere, "adamic_wave12_tailwind.go")
	overlays := map[string]string{virtual: filepath.Join(owned, "decision-oracle.go.txt"), filepath.Join(cohere, "internal/lint/rules/tailwind/adamic_wave12_shim.go"): filepath.Join(owned, "decision-shim.go.txt"), filepath.Join(cohere, "internal/lint/rules/tailwind/no_conflicting_classes.go"): filepath.Join(owned, "decision-conflicts-overlay.go.txt")}
	overlay, _ := json.Marshal(map[string]any{"Replace": overlays})
	overlayPath := filepath.Join(scratch, "overlay.json")
	os.WriteFile(overlayPath, overlay, 0644)
	oracle := filepath.Join(scratch, "oracle")
	run(t, cohere, "go", "build", "-overlay="+overlayPath, "-o", oracle, virtual)
	want := run(t, "", oracle, filepath.Join(owned, "decision-cases.json"))
	for side, output := range outputs(t, root, filepath.Join(owned, "decision-probe.a")) {
		if !bytes.Equal(output, want) {
			os.WriteFile("/tmp/wave12-tailwind-got.txt", output, 0644)
			os.WriteFile("/tmp/wave12-tailwind-want.txt", want, 0644)
			t.Fatalf("side %d differs from Go", side)
		}
	}
	t.Logf("all decision cases agree, %d identical bytes, Go/Node/emitted JS/sanitized native", len(want))
	for _, change := range []struct{ slug, from, to string }{
		{"better-tailwindcss-enforce-shorthand-classes", "if (allPresent) continue;", "if (!allPresent) continue;"},
		{"better-tailwindcss-no-concatenated-classes", "return before ? fields[fields.length - 1] : fields[0];", "return before ? fields[0] : fields[fields.length - 1];"},
		{"better-tailwindcss-no-conflicting-classes", "if (subject.composes) continue;", "if (!subject.composes) continue;"},
	} {
		t.Run(change.slug, func(t *testing.T) {
			directory := t.TempDir()
			for _, slug := range []string{"better-tailwindcss-enforce-shorthand-classes", "better-tailwindcss-no-concatenated-classes", "better-tailwindcss-no-conflicting-classes"} {
				target := filepath.Join(directory, slug)
				os.MkdirAll(target, 0755)
				for _, name := range []string{"decision.a", "decision-probe.a"} {
					data, e := os.ReadFile(filepath.Join(filepath.Dir(owned), slug, name))
					if os.IsNotExist(e) {
						continue
					}
					if e != nil {
						t.Fatal(e)
					}
					if slug == change.slug && name == "decision.a" {
						if strings.Count(string(data), change.from) != 1 {
							t.Fatal("mutation anchor changed")
						}
						data = []byte(strings.Replace(string(data), change.from, change.to, 1))
					}
					if e = os.WriteFile(filepath.Join(target, name), data, 0644); e != nil {
						t.Fatal(e)
					}
				}
			}
			for side, output := range outputs(t, root, filepath.Join(directory, "better-tailwindcss-enforce-shorthand-classes/decision-probe.a")) {
				if bytes.Equal(output, want) {
					t.Fatalf("mutant survived side %d", side)
				}
				t.Logf("compiling clean mutant killed only by comparison, side %d", side)
			}
		})
	}
}
