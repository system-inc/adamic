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

// Not parallel: variants are compiled sequentially to bound compiler memory.
func TestExtractedDecisions(t *testing.T) {
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
	virtual := filepath.Join(cohere, "adamic_wave12_decisions.go")
	overlay, _ := json.Marshal(map[string]any{"Replace": map[string]string{virtual: filepath.Join(owned, "decision-oracle.go.txt")}})
	overlayPath := filepath.Join(scratch, "overlay.json")
	os.WriteFile(overlayPath, overlay, 0644)
	oracle := filepath.Join(scratch, "oracle")
	run(t, cohere, "go", "build", "-overlay="+overlayPath, "-o", oracle, virtual)
	want := run(t, "", oracle, filepath.Join(owned, "decision-cases.json"))
	for side, output := range outputs(t, root, filepath.Join(owned, "decision-probe.a")) {
		if !bytes.Equal(output, want) {
			t.Fatalf("side %d predicates or messages differ: %q vs %q", side, output, want)
		}
	}
	t.Logf("ten manually extracted decision cases, %d identical bytes, Go/source Node/emitted JS/sanitized native", len(want))
	for _, change := range []struct{ slug, from, to string }{
		{"next-google-font-preconnect", "tag === 'link'", "tag === 'LINK'"},
		{"next-no-css-tags", "tag === 'link'", "tag === 'LINK'"},
		{"next-inline-script-id", "boundDefaultNames.includes(tag)", "!boundDefaultNames.includes(tag)"},
		{"next-next-script-for-ga", "tag !== 'script'", "tag === 'script'"},
		{"next-no-before-interactive-script-outside-document", "!document &&", "document &&"},
	} {
		t.Run(change.slug, func(t *testing.T) {
			directory := t.TempDir()
			for _, slug := range []string{"next-google-font-preconnect", "next-no-css-tags", "next-inline-script-id", "next-next-script-for-ga", "next-no-before-interactive-script-outside-document"} {
				target := filepath.Join(directory, slug)
				os.MkdirAll(target, 0755)
				for _, name := range []string{"decision.a", "messages.a", "decision-probe.a"} {
					source := filepath.Join(filepath.Dir(owned), slug, name)
					data, err := os.ReadFile(source)
					if os.IsNotExist(err) {
						continue
					}
					if err != nil {
						t.Fatal(err)
					}
					if slug == change.slug && name == "decision.a" {
						if strings.Count(string(data), change.from) != 1 {
							t.Fatal("mutation anchor changed")
						}
						data = []byte(strings.Replace(string(data), change.from, change.to, 1))
					}
					if err := os.WriteFile(filepath.Join(target, name), data, 0644); err != nil {
						t.Fatal(err)
					}
				}
			}
			for side, output := range outputs(t, root, filepath.Join(directory, "next-google-font-preconnect/decision-probe.a")) {
				if bytes.Equal(output, want) {
					t.Fatalf("mutant survived side %d", side)
				}
				t.Logf("compiling decision mutant caught only by output comparison side %d", side)
			}
		})
	}
}
