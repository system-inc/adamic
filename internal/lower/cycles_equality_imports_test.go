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
		if err := os.MkdirAll(filepath.Join(root, directory), 0700); err != nil {
			test.Fatal(err)
		}
	}
	for _, name := range []string{"parse.a", "context.a"} {
		source, err := os.ReadFile(filepath.Join("../..", filepath.Dir(path), name))
		if err != nil {
			test.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, filepath.Dir(path), name), source, 0600); err != nil {
			test.Fatal(err)
		}
	}
	inputs := buildcache.Inputs{Name: "cycle-equality-import-control", Files: cycleEqualityProgramFiles(path)}
	before, err := buildcache.Key(root, inputs)
	if err != nil {
		test.Fatal(err)
	}
	context := filepath.Join(root, filepath.Dir(path), "context.a")
	file, err := os.OpenFile(context, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		test.Fatal(err)
	}
	_, writeErr := file.WriteString("\n// Harmless sibling-module key control.\n")
	closeErr := file.Close()
	if writeErr != nil || closeErr != nil {
		test.Fatalf("append context comment: %v, %v", writeErr, closeErr)
	}
	after, err := buildcache.Key(root, inputs)
	if err != nil {
		test.Fatal(err)
	}
	if before == after {
		test.Fatal("fixture input key ignored imported context.a; compiler inputs must not mask a missing import")
	}
}
