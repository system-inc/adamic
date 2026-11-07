package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native"
)

// This is an independent runtime/backend proof with explicit typed IR.
// It constructs the typed IR explicitly and does not claim that the fixture passes the checker.
func TestNodeProcessEnvironmentRuntime_000(test *testing.T) {
	test.Parallel()
	nodeProcessEnvironmentRuntime(test, 0)
}
func TestNodeProcessEnvironmentRuntime_001(test *testing.T) {
	test.Parallel()
	nodeProcessEnvironmentRuntime(test, 1)
}
func TestNodeProcessEnvironmentRuntime_002(test *testing.T) {
	test.Parallel()
	nodeProcessEnvironmentRuntime(test, 2)
}
func TestNodeProcessEnvironmentRuntime_003(test *testing.T) {
	test.Parallel()
	nodeProcessEnvironmentRuntime(test, 3)
}
func TestNodeProcessEnvironmentRuntime_004(test *testing.T) {
	test.Parallel()
	nodeProcessEnvironmentRuntime(test, 4)
}

func nodeProcessEnvironmentRuntime(test *testing.T, shard int) {
	source := "internal/oracle/testdata/node_process_runtime/node_process_environment_mutation.a"
	program := &ir.Program{
		Source:  filepath.Base(source),
		Strings: []string{"ADAMIC_HOST_PROCESS_MUTATION", "", "0", "héllo 🌍", "missing"},
	}
	key := ir.StringConstant{Index: 0}
	for _, index := range []int{1, 2, 3} {
		program.Main = append(program.Main,
			ir.Evaluate{Value: ir.NodeFSFile{
				Operation: "host_env_set",
				Arguments: []ir.Expression{key, ir.StringConstant{Index: index}},
				Of:        ir.String,
			}},
			ir.WriteLine{Stream: ir.Stdout, Value: ir.Coalesce{
				Value:    ir.NodeFSFile{Operation: "host_env", Arguments: []ir.Expression{key}, Of: ir.String},
				Fallback: ir.StringConstant{Index: 4},
				Of:       ir.String,
			}},
		)
	}
	program.Main = append(program.Main,
		ir.WriteLine{Stream: ir.Stdout, Value: ir.BooleanToString{
			Value: ir.NodeFSFile{Operation: "host_env_delete", Arguments: []ir.Expression{key}, Of: ir.Boolean},
		}},
		ir.WriteLine{Stream: ir.Stdout, Value: ir.Coalesce{
			Value:    ir.NodeFSFile{Operation: "host_env", Arguments: []ir.Expression{key}, Of: ir.String},
			Fallback: ir.StringConstant{Index: 4},
			Of:       ir.String,
		}},
	)
	absolute, absolutePathError := filepath.Abs(filepath.Join(repository, source))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	truth := execute(
		test,
		"node",
		"--disable-warning=ExperimentalWarning",
		filepath.Join(repository, "oracle/node.mjs"),
		absolute,
	)
	if truth.exitCode != 0 || string(truth.stdout) != "\n0\nhéllo 🌍\ntrue\nmissing\n" {
		test.Fatalf("unexpected Node observation: %+v", truth)
	}
	if shard < 3 {
		nodeProcessRuntimeAgreement(test, program, truth, shard)
		return
	}
	data, readError := os.ReadFile(filepath.Join(repository, "internal/native/runtime/host_runtime.c"))
	if readError != nil {
		test.Fatal(readError)
	}
	for name, anchor := range map[string]string{
		"set":    "(void)setenv(key, text, 1);",
		"delete": "(void)unsetenv(key);",
	} {
		if (name == "set") != (shard == 3) {
			continue
		}
		test.Run(name, func(test *testing.T) {
			if strings.Count(string(data), anchor) != 1 {
				test.Fatal("mutant anchor changed")
			}
			test.Parallel()
			directory := test.TempDir()
			entries, readDirectoryError := os.ReadDir(filepath.Join(repository, "internal/native/runtime"))
			if readDirectoryError != nil {
				test.Fatal(readDirectoryError)
			}
			for _, entry := range entries {
				if entry.IsDir() {
					continue
				}
				data, readError := os.ReadFile(filepath.Join(repository, "internal/native/runtime", entry.Name()))
				if readError != nil {
					test.Fatal(readError)
				}
				if entry.Name() == "host_runtime.c" {
					data = []byte(strings.Replace(string(data), anchor, "(void)key;", 1))
				}
				if writeError := os.WriteFile(filepath.Join(directory, entry.Name()), data, 0600); writeError != nil {
					test.Fatal(writeError)
				}
			}
			code := native.C(program)
			options := native.Options{Sanitize: true}
			archive, runtimeLibraryForSourceError := native.RuntimeLibraryForSource(directory, code, options)
			if runtimeLibraryForSourceError != nil {
				test.Fatal(runtimeLibraryForSourceError)
			}
			mutant := filepath.Join(test.TempDir(), "mutant")
			if buildWithRuntimeError := native.BuildWithRuntime(
				code,
				mutant,
				options,
				archive,
			); buildWithRuntimeError != nil {
				test.Fatal(buildWithRuntimeError)
			}
			bad := execute(test, mutant)
			if bad.exitCode != 0 || len(bad.stderr) != 0 || disagreement(truth, bad) != "stdout differs" {
				test.Fatal("mutant did not fail only Node stdout comparison")
			}
			test.Log("clean runtime mutant caught by Node stdout comparison")
		})
	}
}
