package oracle

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/compilerinputs"
	"github.com/system-inc/adamic/internal/fresh"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

func init() {
	for _, name := range []string{"flat", "array"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/ownership-query/" + name + ".a", true, false})
	}
}

var ownershipProvenPaths = []string{
	"internal/oracle/testdata/ownership-query/flat.a",
	"internal/oracle/testdata/ownership-query/array.a",
	"internal/oracle/testdata/moves/accepted/cleanup.a",
	"internal/oracle/testdata/moves/accepted/objects.a",
}

func ownershipProgram(t *testing.T, path string) (string, run) {
	t.Helper()
	absolute, err := filepath.Abs(filepath.Join(repository, path))
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, absolute)
	checked, err := load.Load([]string{absolute})
	if err != nil {
		t.Fatal(err)
	}
	p, err := lower.LowerWithOptions(context.Background(), checked, lower.Options{OwnershipQuery: true})
	if err != nil {
		t.Fatal(err)
	}
	answers := fresh.OwnershipTransfers(p)
	if len(answers) == 0 {
		t.Fatal("no ownership transfer exercised")
	}
	for _, answer := range answers {
		if answer.Answer.Verdict != fresh.Proven {
			t.Fatalf("move not Proven: %+v", answer)
		}
	}
	script := filepath.Join(t.TempDir(), "main.mjs")
	if err := os.WriteFile(script, []byte(javascript.JavaScript(p)), 0600); err != nil {
		t.Fatal(err)
	}
	if difference := disagreement(truth, onNode(t, script)); difference != "" {
		t.Fatal("JavaScript: " + difference)
	}
	return native.C(p), truth
}

// Code and portable runtime fingerprints declare every input. Race binaries are
// built ahead; each leaf runs one Proven move independently of compilation.
func ownershipProduct(t *testing.T, tsan bool) string {
	t.Helper()
	options := native.Options{Sanitize: !tsan, ThreadSanitize: tsan}
	inputs := buildcache.Inputs{Name: fmt.Sprintf("ownership-query-native-%t", tsan), Files: append(compilerinputs.Lowering(), "internal/oracle/ownership_query_test.go")}
	for _, path := range ownershipProvenPaths {
		inputs.Files = append(inputs.Files, path)
	}
	request, err := native.RuntimeRequestFor(options, "ADAMIC_CANONICAL_CLOSURES")
	if err != nil {
		t.Fatal(err)
	}
	inputs.Flags = append(inputs.Flags, request.Portable...)
	inputs.Toolchain = []string{buildcache.Tool(request.Compiler, "--version"), buildcache.Tool(request.Archiver, "--version"), buildcache.Tool("node", "--version")}
	return buildcache.Product(t, inputs, func(directory string) error {
		for i, path := range ownershipProvenPaths {
			code, truth := ownershipProgram(t, path)
			if err := os.WriteFile(filepath.Join(directory, fmt.Sprintf("truth%d", i)), truth.stdout, 0600); err != nil {
				return err
			}

			if err := native.Build(code, filepath.Join(directory, fmt.Sprintf("case%d", i)), options); err != nil {
				return err
			}
		}
		return nil
	})
}
func TestProduct_OwnershipQueryASan(t *testing.T) { t.Parallel(); ownershipProduct(t, false) }
func TestProduct_OwnershipQueryTSan(t *testing.T) {
	t.Parallel()
	if runtime.GOOS != "linux" {
		t.Skip("Linux TSan")
	}
	ownershipProduct(t, true)
}

func ownershipNative(t *testing.T, index int, tsan bool) {
	t.Helper()
	if tsan && runtime.GOOS != "linux" {
		t.Skip("Linux TSan")
	}
	directory := ownershipProduct(t, tsan)
	output, err := os.ReadFile(filepath.Join(directory, fmt.Sprintf("truth%d", index)))
	if err != nil {
		t.Fatal(err)
	}
	truth := run{stdout: output}
	binary := filepath.Join(directory, fmt.Sprintf("case%d", index))
	attempts := 1
	if tsan {
		attempts = 3
	}
	for _, threads := range []string{"1", "4"} {
		for i := 0; i < attempts; i++ {
			got := executeParallel(t, threads, !tsan, binary)
			if difference := disagreement(truth, got); difference != "" {
				t.Fatalf("threads=%s %s: %+v", threads, difference, got)
			}
		}
	}
}
func TestOwnershipQueryFlatASan(t *testing.T)    { t.Parallel(); ownershipNative(t, 0, false) }
func TestOwnershipQueryArrayASan(t *testing.T)   { t.Parallel(); ownershipNative(t, 1, false) }
func TestOwnershipQueryCleanupASan(t *testing.T) { t.Parallel(); ownershipNative(t, 2, false) }
func TestOwnershipQueryObjectsASan(t *testing.T) { t.Parallel(); ownershipNative(t, 3, false) }
func TestOwnershipQueryFlatTSan(t *testing.T)    { t.Parallel(); ownershipNative(t, 0, true) }
func TestOwnershipQueryArrayTSan(t *testing.T)   { t.Parallel(); ownershipNative(t, 1, true) }
func TestOwnershipQueryCleanupTSan(t *testing.T) { t.Parallel(); ownershipNative(t, 2, true) }
func TestOwnershipQueryObjectsTSan(t *testing.T) { t.Parallel(); ownershipNative(t, 3, true) }

func ownershipRefused(t *testing.T, name, want string) {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/ownership-query", name+".a"))
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	if truth.exitCode != 0 || string(truth.stdout) != want {
		t.Fatalf("Node: %+v", truth)
	}
	var previous string
	for _, enabled := range []bool{false, true} {
		checked, err := load.Load([]string{path})
		if err != nil {
			t.Fatal(err)
		}
		_, err = lower.LowerWithOptions(context.Background(), checked, lower.Options{OwnershipQuery: enabled})
		var refused *lower.Refused
		if !errors.As(err, &refused) {
			t.Fatalf("expected named move refusal, got %v", err)
		}
		if !enabled {
			previous = err.Error()
		} else if previous != err.Error() {
			t.Fatalf("refusal changed: %v", err)
		}
	}
	t.Logf("Node %q; unchanged refusal: %s", want, previous)
}
func TestOwnershipQueryNestedNode(t *testing.T) { t.Parallel(); ownershipRefused(t, "nested", "2\n") }
func TestOwnershipQueryRingNode(t *testing.T)   { t.Parallel(); ownershipRefused(t, "ring", "2 2\n") }
func TestOwnershipQueryOutsideRingNode(t *testing.T) {
	t.Parallel()
	ownershipRefused(t, "outside-ring", "2 2\n")
}
