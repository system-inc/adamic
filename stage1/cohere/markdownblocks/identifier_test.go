package markdownblocks

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode"
)

func TestMdastIdentifierScalars(t *testing.T) {
	root, e := filepath.Abs(repository)
	if e != nil {
		t.Fatal(e)
	}
	dir := t.TempDir()
	cohere := filepath.Join(root, "cohere")
	mainPath := filepath.Join(cohere, "cmd/adamic_identifier/main.go")
	driver, e := filepath.Abs("testdata/identifier_go.go")
	if e != nil {
		t.Fatal(e)
	}
	overlay, e := json.Marshal(map[string]any{"Replace": map[string]string{mainPath: driver}})
	if e != nil {
		t.Fatal(e)
	}
	overlayPath := filepath.Join(dir, "overlay.json")
	write(t, overlayPath, overlay)
	goBinary := filepath.Join(dir, "go-identifier")
	build := bounded(t, "go", "build", "-overlay="+overlayPath, "-o", goBinary, mainPath)
	build.Dir = cohere
	if output, e := build.CombinedOutput(); e != nil {
		t.Fatalf("Go identifier %v %s", e, output)
	}
	truth := execute(t, nil, goBinary)
	clean(t, "actual Go NormalizeIdentifier", truth)
	main, e := filepath.Abs("testdata/identifier_probe.ts")
	if e != nil {
		t.Fatal(e)
	}
	program := lowered(t, main)
	answer, binary := natively(t, program)
	for _, side := range []run{answer, onNode(t, main), onJavaScriptBackend(t, program)} {
		clean(t, "identifier scalar oracle", side)
		equal(t, "identifier scalar oracle", side.stdout, truth.stdout)
	}
	if report := leaks(t, program, binary); report != "" {
		t.Fatal(report)
	}
	scratch := t.TempDir()
	if e := os.Mkdir(filepath.Join(scratch, "testdata"), 0755); e != nil {
		t.Fatal(e)
	}
	for _, file := range []string{"identifier.ts", "identifierCaseKeys.ts", "identifierCaseValues.ts", "testdata/identifier_probe.ts"} {
		data, e := os.ReadFile(file)
		if e != nil {
			t.Fatal(e)
		}
		if file == "identifierCaseKeys.ts" {
			if !strings.Contains(string(data), "65,") {
				t.Fatal("case table mutation anchor")
			}
			data = []byte(strings.Replace(string(data), "65,", "64,", 1))
		}
		write(t, filepath.Join(scratch, file), data)
	}
	mutant := onNode(t, filepath.Join(scratch, "testdata/identifier_probe.ts"))
	clean(t, "case table mutant", mutant)
	if bytes.Equal(mutant.stdout, truth.stdout) {
		t.Fatal("case table mutant survived")
	}
	t.Logf("all1112064 Unicode scalars match actual Go normalization, native/source/backend and sanitizer/leaks; output-only case-table mutant caught at byte%d; Go Unicode%s", firstDifference(string(mutant.stdout), string(truth.stdout)), unicode.Version)
}
