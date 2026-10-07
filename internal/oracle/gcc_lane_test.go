package oracle

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
)

// This lane deliberately has no observation cache. The compiler can be switched to clang to
// measure the same work on the same fixtures, rather than comparing different harnesses.
func TestGCCAgreesWithNode(t *testing.T) {
	if os.Getenv("ADAMIC_GCC_LANE") != "1" {
		t.Skip("opt in with ADAMIC_GCC_LANE=1")
	}
	t.Parallel()
	compiler := os.Getenv("ADAMIC_LANE_CC")
	if compiler == "" {
		compiler = "gcc"
	}
	sanitize := os.Getenv("ADAMIC_LANE_SANITIZE") == "1"
	options := native.Options{Compiler: compiler, Sanitize: sanitize}
	t.Logf("compiler=%s flags=%s", compiler, strings.Join(gccLaneFlags(options), " "))
	// Build once before the fixture fan-out. A failed runtime still allows every emitted
	// translation unit to be checked independently for GCC diagnostics.
	library, runtimeError := gccLaneRuntime(t, options)
	for _, fixture := range fixtures {
		if !fixture.lowers {
			continue
		}
		t.Run(fixture.path, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, fixture.path))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			// Execute Node directly: even a shell without ADAMIC_GATE_UNCACHED cannot reuse results.
			oracle := execute(t, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle", "node.mjs"), path)
			sourceOracle := oracle
			if fixture.checked {
				script := filepath.Join(t.TempDir(), "program.mjs")
				if err := os.WriteFile(script, []byte(javascript.JavaScript(program)), 0o644); err != nil {
					t.Fatal(err)
				}
				oracle = execute(t, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), script)
			}
			binary := filepath.Join(t.TempDir(), "program")
			code := native.C(program)
			if runtimeError != nil {
				source := filepath.Join(filepath.Dir(binary), "main.c")
				if err := os.WriteFile(source, []byte(code), 0o644); err != nil {
					t.Fatal(err)
				}
				arguments := append(gccLaneFlags(options), "-I", filepath.Join(repository, "internal/native/runtime"), "-c", source, "-o", binary+".o")
				output, err := exec.Command(compiler, arguments...).CombinedOutput()
				if err != nil {
					t.Errorf("emitted C: %v\n%s", err, output)
				}
				t.Fatalf("runtime blocks linking: %v", runtimeError)
			}
			if err := gccLaneBuild(t, code, binary, library, options); err != nil {
				t.Fatal(err)
			}
			actual := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=0"}, binary)
			gccLaneRecord(t, "observation.json", map[string]any{"fixture": fixture.path, "node_stdout": string(oracle.stdout), "node_stderr": string(oracle.stderr), "node_exit": oracle.exitCode, "native_stdout": string(actual.stdout), "native_stderr": string(actual.stderr), "native_exit": actual.exitCode})
			if difference := disagreement(oracle, actual); difference != "" {
				t.Errorf("%s: Node exit %d, native exit %d\nstdout %s\nstderr %s", difference, oracle.exitCode, actual.exitCode, firstLaneDifference(oracle.stdout, actual.stdout), firstLaneDifference(oracle.stderr, actual.stderr))
			}
			if fixture.checked && (actual.exitCode != 70 || sourceOracle.exitCode == 70) {
				t.Errorf("inserted check did not fire: exit %d", actual.exitCode)
			}
			if sanitize && oracle.exitCode == 0 {
				report := leakChecked(t, code, binary)
				gccLaneRecord(t, "leaks.json", report)
				if report != "" {
					t.Errorf("leaks: %s", report)
				}
			}
		})
	}
}

func firstLaneDifference(want, got []byte) string {
	if bytes.Equal(want, got) {
		return "identical"
	}
	left, right := bytes.Split(want, []byte("\n")), bytes.Split(got, []byte("\n"))
	for line := 0; line < len(left) || line < len(right); line++ {
		var a, b []byte
		if line < len(left) {
			a = left[line]
		}
		if line < len(right) {
			b = right[line]
		}
		if !bytes.Equal(a, b) || line >= len(left) || line >= len(right) {
			return fmt.Sprintf("line %d: Node %q, native %q", line+1, a, b)
		}
	}
	return "different bytes"
}

// The comparison uses the same C and Node runners on a real fixture with one changed byte.
// Opted-in runs use GCC; ordinary gates keep the existing clang prerequisite.
func TestGCCLaneComparisonCatchesMutants(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "dedication/dedication.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	program.Strings[0] += "!"
	binary := filepath.Join(t.TempDir(), "mutant")
	options := native.Options{Compiler: "clang"}
	if os.Getenv("ADAMIC_GCC_LANE") == "1" {
		options.Compiler = "gcc"
	}
	library, err := gccLaneRuntime(t, options)
	if err != nil {
		t.Fatal(err)
	}
	if err := gccLaneBuild(t, native.C(program), binary, library, options); err != nil {
		t.Fatal(err)
	}
	truth := execute(t, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), path)
	mutant := execute(t, binary)
	if difference := disagreement(truth, mutant); difference != "stdout differs" {
		t.Fatalf("mutant: %q", difference)
	}
	if difference := firstLaneDifference(truth.stdout, mutant.stdout); !strings.HasPrefix(difference, "line 1:") {
		t.Fatal(difference)
	}
}

// The diagnostic lane builds fresh snapshots without a runtime or observation cache.
// Keeping its warning policy here leaves every production clang build unchanged.
func gccLaneFlags(options native.Options) []string {
	flags := native.Flags(options)
	if options.Compiler == "gcc" {
		kept := flags[:0]
		for _, flag := range flags {
			if flag != "-Werror" {
				kept = append(kept, flag)
			}
		}
		flags = kept
	}
	return flags
}

func gccLaneRecord(t *testing.T, name string, value any) {
	t.Helper()
	root := os.Getenv("ADAMIC_LANE_REPORT")
	if root == "" {
		return
	}
	directory := filepath.Join(root, strings.ReplaceAll(t.Name(), "/", "__"))
	if err := os.MkdirAll(directory, 0755); err != nil {
		t.Fatal(err)
	}
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, name), data, 0644); err != nil {
		t.Fatal(err)
	}
}

func gccLaneCommand(t *testing.T, name string, arguments ...string) error {
	t.Helper()
	output, err := exec.Command(name, arguments...).CombinedOutput()
	gccLaneRecord(t, name+"-"+filepath.Base(arguments[len(arguments)-1])+".diagnostics.json", map[string]any{"command": append([]string{name}, arguments...), "output": string(output)})
	if len(output) != 0 {
		t.Logf("compiler diagnostics: %s", output)
	}
	if err != nil {
		return fmt.Errorf("%s: %w\n%s", name, err, output)
	}
	return nil
}

func gccLaneRuntime(t *testing.T, options native.Options) (string, error) {
	t.Helper()
	directory := t.TempDir()
	source := filepath.Join(repository, "internal/native/runtime")
	entries, err := os.ReadDir(source)
	if err != nil {
		return "", err
	}
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".c") && !strings.HasSuffix(entry.Name(), ".h") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(source, entry.Name()))
		if err != nil {
			return "", err
		}
		if err := os.WriteFile(filepath.Join(directory, entry.Name()), data, 0644); err != nil {
			return "", err
		}
	}
	var objects []string
	var failures []string
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".c") {
			continue
		}
		object := filepath.Join(directory, entry.Name()+".o")
		args := append(gccLaneFlags(options), "-c", filepath.Join(directory, entry.Name()), "-o", object)
		if err := gccLaneCommand(t, options.Compiler, args...); err != nil {
			failures = append(failures, err.Error())
		}
		objects = append(objects, object)
	}
	if len(failures) != 0 {
		return "", fmt.Errorf("runtime: %s", strings.Join(failures, "\n"))
	}
	library := filepath.Join(directory, "runtime.a")
	return library, gccLaneCommand(t, "ar", append([]string{"rcs", library}, objects...)...)
}

func gccLaneBuild(t *testing.T, code, binary, library string, options native.Options) error {
	t.Helper()
	source := filepath.Join(filepath.Dir(binary), "main.c")
	if err := os.WriteFile(source, []byte(code), 0644); err != nil {
		return err
	}
	gccLaneRecord(t, "emitted.json", code)
	args := append(gccLaneFlags(options), "-I", filepath.Dir(library), "-o", binary, source)
	args = append(args, native.RuntimeLinkFlags(library)...)
	args = append(args, "-lm")
	return gccLaneCommand(t, options.Compiler, args...)
}

func TestGCCLaneWarningPolicy(t *testing.T) {
	for _, compiler := range []string{"gcc", "clang"} {
		flags := gccLaneFlags(native.Options{Compiler: compiler})
		fatal := false
		for _, flag := range flags {
			if flag == "-Werror" {
				fatal = true
			}
		}
		if fatal != (compiler == "clang") {
			t.Fatalf("%s flags: %v", compiler, flags)
		}
	}
}
