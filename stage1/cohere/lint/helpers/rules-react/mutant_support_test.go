package reacthelpers

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native"
)

// All semantic mutants run on Node and emitted JavaScript. One edited AST helper
// also runs sanitized native, which must equal the mutated Node bytes, as lint's
// TestMutants requires. Unmutated agreement cases keep all three backends.
const nativeCanaryMutant = "is_this_expression.a"

// Silent clang builds get five minutes; a silent or stalled mutant run gets two
// minutes before first output and one after it. Both bounds expire before the
// package timeout and childguard kills the entire child process group.
var buildGuard = childguard.Options{FirstOutput: 5 * time.Minute, Stall: 5 * time.Minute, Ceiling: 10 * time.Minute}
var runGuard = childguard.Options{FirstOutput: 2 * time.Minute, Stall: time.Minute, Ceiling: 5 * time.Minute}

func buildNative(t *testing.T, program *ir.Program) string {
	t.Helper()
	directory := t.TempDir()
	source := filepath.Join(directory, "native.c")
	binary := filepath.Join(directory, "native")
	write(t, source, []byte(native.C(program)))
	// native.Build invokes clang in process. Put it in a child test so childguard
	// can stop both that call and clang rather than leave a timed-out goroutine.
	command := exec.Command(os.Args[0], "-test.run=^$")
	command.Env = append(os.Environ(), "ADAMIC_REACT_BUILD_SOURCE="+source, "ADAMIC_REACT_BUILD_BINARY="+binary)
	if output, err := childguard.CombinedOutput(command, buildGuard); err != nil {
		t.Fatalf("%s native build: %v\n%s", t.Name(), err, output)
	}
	return binary
}

// The guarded build enters in a child; ordinary package runs keep their test
// selection and do no extra compilation here. The parent owns the source and binary.
func TestMain(m *testing.M) {
	source := os.Getenv("ADAMIC_REACT_BUILD_SOURCE")
	if source == "" {
		os.Exit(m.Run())
	}
	data, err := os.ReadFile(source)
	if err == nil {
		err = native.Build(string(data), os.Getenv("ADAMIC_REACT_BUILD_BINARY"), native.Options{Sanitize: true})
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Exit(0)
}

func checkMutant(t *testing.T, directory, input, entry string, canary bool, want []byte) {
	t.Helper()
	names := []string{"Node", "emitted JavaScript"}
	if canary {
		names = []string{"Node", "sanitized native", "emitted JavaScript"}
	}
	outputs := entryOutputs(t, directory, input, entry, canary)
	for i, output := range outputs {
		if bytes.Equal(output, want) {
			t.Fatalf("%s mutant survived on %s", t.Name(), names[i])
		}
		t.Logf("%s compiling semantic mutant caught by Go output", names[i])
	}
	if canary {
		if !bytes.Equal(outputs[0], outputs[1]) {
			t.Fatalf("%s native canary differs from mutated Node", t.Name())
		}
		t.Logf("sanitized native canary equals mutated Node: %d bytes", len(outputs[1]))
	}
}

// Every parallel AST mutant needs an isolated tree. A shared tree would let one
// mutant change the source another backend is compiling or running.
func copyAstMutantTree(t *testing.T) string {
	t.Helper()
	originalRoot, _ := filepath.Abs("../../../../../")
	root := t.TempDir()
	err := filepath.WalkDir(filepath.Join(originalRoot, "stage1"), func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if entry.Name() == "testdata" || entry.Name() == "gaps" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".a") && !strings.HasSuffix(path, ".ts") {
			return nil
		}
		relative, _ := filepath.Rel(originalRoot, path)
		target := filepath.Join(root, relative)
		if err = os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0644)
	})
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Join(root, "stage1/cohere/lint/helpers/rules-react")
}
