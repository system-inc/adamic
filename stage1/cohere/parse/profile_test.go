package parse

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func run(t *testing.T, directory, name string, args ...string) []byte {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	command := exec.CommandContext(ctx, name, args...)
	command.Dir = directory
	file, err := os.Create(filepath.Join(t.TempDir(), "stdout"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	command.Stdout = file
	var stderr bytes.Buffer
	command.Stderr = &stderr
	if err = command.Run(); err != nil || stderr.Len() != 0 {
		t.Fatalf("%s %v: %v\n%s", name, args, err, &stderr)
	}
	data, err := os.ReadFile(file.Name())
	if err != nil {
		t.Fatal(err)
	}
	return data
}

// Not parallel: the exact shipping artifact is compared with a second build before byte checks.
func TestShippedProfileAgreesWithGo(t *testing.T) {
	binary := os.Getenv("ADAMIC_STAGE1_PARSE_BINARY")
	if binary == "" {
		t.Skip("set ADAMIC_STAGE1_PARSE_BINARY to the profile-built shipping binary")
	}
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	rebuilt := filepath.Join(t.TempDir(), "parse")
	run(t, repository, "go", "run", "./cmd/adamic-stage1", "-driver", "parse", "-o", rebuilt)
	first, err := os.ReadFile(binary)
	if err != nil {
		t.Fatal(err)
	}
	second, err := os.ReadFile(rebuilt)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first, second) {
		t.Fatal("shipping artifact differs from same-source/profile rebuild")
	}
	t.Logf("tested shipping binary SHA256 %x; second build byte-identical", sha256.Sum256(first))
	source := os.Getenv("ADAMIC_TYPESCRIPT_SOURCE")
	if source == "" {
		t.Fatal("missing pinned compiler corpus")
	}
	compiler, err := filepath.Abs(filepath.Join(repository, "cohere/TypeScript/tsc"))
	if err != nil {
		t.Fatal(err)
	}
	virtual := filepath.Join(compiler, "adamic_parser_oracle.go")
	overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{virtual: filepath.Join(repository, "stage1/typescript/parser/testdata/oracle.go")}})
	if err != nil {
		t.Fatal(err)
	}
	work := t.TempDir()
	overlayFile := filepath.Join(work, "overlay.json")
	if err = os.WriteFile(overlayFile, overlay, 0o644); err != nil {
		t.Fatal(err)
	}
	oracle := filepath.Join(work, "oracle")
	run(t, compiler, "go", "build", "-pgo=off", "-overlay="+overlayFile, "-o", oracle, virtual)
	files, err := filepath.Glob(filepath.Join(source, "src/compiler/*.ts"))
	if err != nil {
		t.Fatal(err)
	}
	err = filepath.WalkDir(filepath.Join(source, "src/compiler"), func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() && filepath.Ext(path) == ".ts" && filepath.Dir(path) != filepath.Join(source, "src/compiler") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 77 {
		t.Fatalf("expected 77 benchmark files, got %d", len(files))
	}
	fixture := filepath.Join(work, "fixture.a")
	if err = os.WriteFile(fixture, []byte("const x = 0.1 * 10 - 1; const y = {a: [1, 2]};\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	files = append(files, fixture)
	var manifest bytes.Buffer
	for _, p := range files {
		fmt.Fprintln(&manifest, p)
	}
	manifestFile := filepath.Join(work, "manifest.txt")
	if err = os.WriteFile(manifestFile, manifest.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
	want := run(t, "", oracle, "--manifest", manifestFile, "--whole")
	got := run(t, "", binary, "--manifest", manifestFile, "--ast")
	if !bytes.Equal(got, want) {
		t.Fatalf("shipping AST bytes differ from Go: got %d want %d", len(got), len(want))
	}
	node := run(t, "", "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), filepath.Join(repository, "stage1/cohere/parse/parse.a"), "--manifest", manifestFile, "--ast")
	if !bytes.Equal(node, want) {
		t.Fatal("source Node AST bytes differ from Go")
	}
	mutant := append([]byte{}, got...)
	mutant[14] ^= 1
	if bytes.Equal(mutant, want) {
		t.Fatal("same-length output-byte mutant survived")
	}
	t.Logf("same shipping binary and source Node match Go: %d full AST bytes; byte mutant caught", len(want))
}
