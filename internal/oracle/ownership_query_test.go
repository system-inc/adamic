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

func ownershipProgram(test *testing.T, path string) (string, run) {
	test.Helper()
	absolute, err := filepath.Abs(filepath.Join(repository, path))
	if err != nil {
		test.Fatal(err)
	}
	truth := onNode(test, absolute)
	checked, err := load.Load([]string{absolute})
	if err != nil {
		test.Fatal(err)
	}
	p, err := lower.LowerWithOptions(context.Background(), checked, lower.Options{OwnershipQuery: true})
	if err != nil {
		test.Fatal(err)
	}
	answers := fresh.OwnershipTransfers(p)
	if len(answers) == 0 {
		test.Fatal("no ownership transfer exercised")
	}
	for _, answer := range answers {
		if answer.Answer.Verdict != fresh.Proven {
			test.Fatalf("move not Proven: %+v", answer)
		}
	}
	script := filepath.Join(test.TempDir(), "main.mjs")
	if err := os.WriteFile(script, []byte(javascript.JavaScript(p)), 0600); err != nil {
		test.Fatal(err)
	}
	if difference := disagreement(truth, onNode(test, script)); difference != "" {
		test.Fatal("JavaScript: " + difference)
	}
	return native.C(p), truth
}

// Code and portable runtime fingerprints declare every input. Race binaries are
// built ahead; each leaf runs one Proven move independently of compilation.
func ownershipProduct(test *testing.T, tsan bool) string {
	test.Helper()
	options := native.Options{Sanitize: !tsan, ThreadSanitize: tsan}
	inputs := buildcache.Inputs{
		Name:  fmt.Sprintf("ownership-query-native-%t", tsan),
		Files: append(compilerinputs.Lowering(), "internal/oracle/ownership_query_test.go"),
	}
	for _, path := range ownershipProvenPaths {
		inputs.Files = append(inputs.Files, path)
	}
	request, err := native.RuntimeRequestFor(options, "ADAMIC_CANONICAL_CLOSURES")
	if err != nil {
		test.Fatal(err)
	}
	inputs.Flags = append(inputs.Flags, request.Portable...)
	inputs.Toolchain = []string{
		buildcache.Tool(request.Compiler, "--version"),
		buildcache.Tool(request.Archiver, "--version"),
		buildcache.Tool("node", "--version"),
	}
	return buildcache.Product(test, inputs, func(directory string) error {
		for i, path := range ownershipProvenPaths {
			code, truth := ownershipProgram(test, path)
			if err := os.WriteFile(
				filepath.Join(directory, fmt.Sprintf("truth%d", i)),
				truth.stdout,
				0600,
			); err != nil {
				return err
			}

			if err := native.Build(code, filepath.Join(directory, fmt.Sprintf("case%d", i)), options); err != nil {
				return err
			}
		}
		return nil
	})
}
func TestProduct_OwnershipQueryASan(test *testing.T) { test.Parallel(); ownershipProduct(test, false) }
func TestProduct_OwnershipQueryTSan(test *testing.T) {
	test.Parallel()
	if runtime.GOOS != "linux" {
		test.Skip("Linux TSan")
	}
	ownershipProduct(test, true)
}

func ownershipNative(test *testing.T, index int, tsan bool) {
	test.Helper()
	if tsan && runtime.GOOS != "linux" {
		test.Skip("Linux TSan")
	}
	directory := ownershipProduct(test, tsan)
	output, err := os.ReadFile(filepath.Join(directory, fmt.Sprintf("truth%d", index)))
	if err != nil {
		test.Fatal(err)
	}
	truth := run{stdout: output}
	binary := filepath.Join(directory, fmt.Sprintf("case%d", index))
	attempts := 1
	if tsan {
		attempts = 3
	}
	for _, threads := range []string{"1", "4"} {
		for i := 0; i < attempts; i++ {
			got := executeParallel(test, threads, !tsan, binary)
			if difference := disagreement(truth, got); difference != "" {
				test.Fatalf("threads=%s %s: %+v", threads, difference, got)
			}
		}
	}
}
func TestOwnershipQueryFlatASan(test *testing.T)    { test.Parallel(); ownershipNative(test, 0, false) }
func TestOwnershipQueryArrayASan(test *testing.T)   { test.Parallel(); ownershipNative(test, 1, false) }
func TestOwnershipQueryCleanupASan(test *testing.T) { test.Parallel(); ownershipNative(test, 2, false) }
func TestOwnershipQueryObjectsASan(test *testing.T) { test.Parallel(); ownershipNative(test, 3, false) }
func TestOwnershipQueryFlatTSan(test *testing.T)    { test.Parallel(); ownershipNative(test, 0, true) }
func TestOwnershipQueryArrayTSan(test *testing.T)   { test.Parallel(); ownershipNative(test, 1, true) }
func TestOwnershipQueryCleanupTSan(test *testing.T) { test.Parallel(); ownershipNative(test, 2, true) }
func TestOwnershipQueryObjectsTSan(test *testing.T) { test.Parallel(); ownershipNative(test, 3, true) }

func ownershipRefused(test *testing.T, name, want string) {
	test.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/ownership-query", name+".a"))
	if err != nil {
		test.Fatal(err)
	}
	truth := onNode(test, path)
	if truth.exitCode != 0 || string(truth.stdout) != want {
		test.Fatalf("Node: %+v", truth)
	}
	var previous string
	for _, enabled := range []bool{false, true} {
		checked, err := load.Load([]string{path})
		if err != nil {
			test.Fatal(err)
		}
		_, err = lower.LowerWithOptions(context.Background(), checked, lower.Options{OwnershipQuery: enabled})
		var refused *lower.Refused
		if !errors.As(err, &refused) {
			test.Fatalf("expected named move refusal, got %v", err)
		}
		if !enabled {
			previous = err.Error()
		} else if previous != err.Error() {
			test.Fatalf("refusal changed: %v", err)
		}
	}
	test.Logf("Node %q; unchanged refusal: %s", want, previous)
}
func TestOwnershipQueryNestedNode(test *testing.T) {
	test.Parallel()
	ownershipRefused(test, "nested", "2\n")
}
func TestOwnershipQueryRingNode(test *testing.T) {
	test.Parallel()
	ownershipRefused(test, "ring", "2 2\n")
}
func TestOwnershipQueryOutsideRingNode(test *testing.T) {
	test.Parallel()
	ownershipRefused(test, "outside-ring", "2 2\n")
}
