package lint

import (
	"bytes"
	"crypto/sha256"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Not parallel: fixture capture and shipping artifact rebuilding use process-wide state.
func TestShippedProfileAgreesWithGo(t *testing.T) {
	binary := os.Getenv("ADAMIC_STAGE1_LINT_BINARY")
	if binary == "" {
		t.Skip("set ADAMIC_STAGE1_LINT_BINARY to the profile-built shipping binary")
	}
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	rebuilt := filepath.Join(t.TempDir(), "lint")
	execute(t, root, "go", "run", "./cmd/adamic-stage1", "-driver", "lint", "-o", rebuilt)
	first, err := os.ReadFile(binary)
	if err != nil {
		t.Fatal(err)
	}
	second, err := os.ReadFile(rebuilt)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first, second) {
		t.Fatal("shipping binary differs from the same-source/profile rebuild")
	}
	t.Logf("tested shipping binary SHA256 %x; second build byte-identical", sha256.Sum256(first))
	oracle := goOracle(t)
	directory, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	rows := append(generated(t), volumeGenerated(t)...)
	for _, row := range upstream(t) {
		if strings.HasSuffix(row, "\tunsupported-recovery") {
			checkRecoveryRefusal(t, oracle, binary, directory, row)
		} else {
			rows = append(rows, row)
		}
	}
	source := os.Getenv("ADAMIC_TYPESCRIPT_SOURCE")
	if source == "" {
		t.Fatal("missing pinned compiler corpus")
	}
	err = filepath.WalkDir(filepath.Join(source, "src/compiler"), func(path string, entry os.DirEntry, err error) error {
		if err == nil && !entry.IsDir() && filepath.Ext(path) == ".ts" {
			rows = append(rows, path)
		}
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	compare(t, oracle, binary, directory, manifest(t, rows))
	t.Log("the tested profile-built binary is the shipping artifact, including all compiler files and supported upstream cases")
}
