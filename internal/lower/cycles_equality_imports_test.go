package lower

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
)

// Hold fixture dependencies independently of compilerinputs.Lowering(): today's
// broad "internal" compiler input must not mask a missing sibling-module input.
func TestCycleEqualityProgramInputsTrackSiblingChanges(test *testing.T) {
	test.Parallel()
	const path = "internal/oracle/testdata/program_region/parse.a"
	root := test.TempDir()
	for _, directory := range []string{"review", "stage1", "stage3/fixtures/cycles", filepath.Dir(path)} {
		if createDirectoryError := os.MkdirAll(filepath.Join(root, directory), 0700); createDirectoryError != nil {
			test.Fatal(createDirectoryError)
		}
	}
	for _, name := range []string{"parse.a", "context.a"} {
		source, readError := os.ReadFile(filepath.Join("../..", filepath.Dir(path), name))
		if readError != nil {
			test.Fatal(readError)
		}
		if writeError := os.WriteFile(filepath.Join(root, filepath.Dir(path), name), source, 0600); writeError != nil {
			test.Fatal(writeError)
		}
	}
	inputs := buildcache.Inputs{Name: "cycle-equality-import-control", Files: cycleEqualityProgramFiles(path)}
	before, keyError := buildcache.Key(root, inputs)
	if keyError != nil {
		test.Fatal(keyError)
	}
	context := filepath.Join(root, filepath.Dir(path), "context.a")
	file, openError := os.OpenFile(context, os.O_APPEND|os.O_WRONLY, 0600)
	if openError != nil {
		test.Fatal(openError)
	}
	_, writeError := file.WriteString("\n// Harmless sibling-module key control.\n")
	closeError := file.Close()
	if writeError != nil || closeError != nil {
		test.Fatalf("append context comment: %v, %v", writeError, closeError)
	}
	after, keyError := buildcache.Key(root, inputs)
	if keyError != nil {
		test.Fatal(keyError)
	}
	if before == after {
		test.Fatal("fixture input key ignored imported context.a; compiler inputs must not mask a missing import")
	}
}
