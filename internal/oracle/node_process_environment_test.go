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
func TestNodeProcessEnvironmentRuntime_000(t *testing.T) {
	t.Parallel()
	nodeProcessEnvironmentRuntime(t, 0)
}
func TestNodeProcessEnvironmentRuntime_001(t *testing.T) {
	t.Parallel()
	nodeProcessEnvironmentRuntime(t, 1)
}
func TestNodeProcessEnvironmentRuntime_002(t *testing.T) {
	t.Parallel()
	nodeProcessEnvironmentRuntime(t, 2)
}
func TestNodeProcessEnvironmentRuntime_003(t *testing.T) {
	t.Parallel()
	nodeProcessEnvironmentRuntime(t, 3)
}
func TestNodeProcessEnvironmentRuntime_004(t *testing.T) {
	t.Parallel()
	nodeProcessEnvironmentRuntime(t, 4)
}

func nodeProcessEnvironmentRuntime(t *testing.T, shard int) {
	source := "internal/oracle/testdata/node_process_runtime/node_process_environment_mutation.a"
	program := &ir.Program{Source: filepath.Base(source), Strings: []string{"ADAMIC_HOST_PROCESS_MUTATION", "", "0", "héllo 🌍", "missing"}}
	key := ir.StringConstant{Index: 0}
	for _, index := range []int{1, 2, 3} {
		program.Main = append(program.Main,
			ir.Evaluate{Value: ir.NodeFSFile{Operation: "host_env_set", Arguments: []ir.Expression{key, ir.StringConstant{Index: index}}, Of: ir.String}},
			ir.WriteLine{Stream: ir.Stdout, Value: ir.Coalesce{Value: ir.NodeFSFile{Operation: "host_env", Arguments: []ir.Expression{key}, Of: ir.String}, Fallback: ir.StringConstant{Index: 4}, Of: ir.String}},
		)
	}
	program.Main = append(program.Main,
		ir.WriteLine{Stream: ir.Stdout, Value: ir.BooleanToString{Value: ir.NodeFSFile{Operation: "host_env_delete", Arguments: []ir.Expression{key}, Of: ir.Boolean}}},
		ir.WriteLine{Stream: ir.Stdout, Value: ir.Coalesce{Value: ir.NodeFSFile{Operation: "host_env", Arguments: []ir.Expression{key}, Of: ir.String}, Fallback: ir.StringConstant{Index: 4}, Of: ir.String}},
	)
	absolute, err := filepath.Abs(filepath.Join(repository, source))
	if err != nil {
		t.Fatal(err)
	}
	truth := execute(t, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), absolute)
	if truth.exitCode != 0 || string(truth.stdout) != "\n0\nhéllo 🌍\ntrue\nmissing\n" {
		t.Fatalf("unexpected Node observation: %+v", truth)
	}
	if shard < 3 {
		nodeProcessRuntimeAgreement(t, program, truth, shard)
		return
	}
	data, err := os.ReadFile(filepath.Join(repository, "internal/native/runtime/host_runtime.c"))
	if err != nil {
		t.Fatal(err)
	}
	for name, anchor := range map[string]string{"set": "(void)setenv(key, text, 1);", "delete": "(void)unsetenv(key);"} {
		if (name == "set") != (shard == 3) {
			continue
		}
		t.Run(name, func(t *testing.T) {
			if strings.Count(string(data), anchor) != 1 {
				t.Fatal("mutant anchor changed")
			}
			t.Parallel()
			directory := t.TempDir()
			entries, err := os.ReadDir(filepath.Join(repository, "internal/native/runtime"))
			if err != nil {
				t.Fatal(err)
			}
			for _, entry := range entries {
				if entry.IsDir() {
					continue
				}
				data, err := os.ReadFile(filepath.Join(repository, "internal/native/runtime", entry.Name()))
				if err != nil {
					t.Fatal(err)
				}
				if entry.Name() == "host_runtime.c" {
					data = []byte(strings.Replace(string(data), anchor, "(void)key;", 1))
				}
				if err := os.WriteFile(filepath.Join(directory, entry.Name()), data, 0600); err != nil {
					t.Fatal(err)
				}
			}
			code := native.C(program)
			options := native.Options{Sanitize: true}
			archive, err := native.RuntimeLibraryForSource(directory, code, options)
			if err != nil {
				t.Fatal(err)
			}
			mutant := filepath.Join(t.TempDir(), "mutant")
			if err := native.BuildWithRuntime(code, mutant, options, archive); err != nil {
				t.Fatal(err)
			}
			bad := execute(t, mutant)
			if bad.exitCode != 0 || len(bad.stderr) != 0 || disagreement(truth, bad) != "stdout differs" {
				t.Fatal("mutant did not fail only Node stdout comparison")
			}
			t.Log("clean runtime mutant caught by Node stdout comparison")
		})
	}
}
