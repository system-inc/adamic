package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
)

// Runtime/backend proof with explicit typed IR, independent of source lowering.
func TestNodeProcessDirectoryRuntime_000(t *testing.T) {
	t.Parallel()
	nodeProcessDirectoryRuntime(t, 0)
}
func TestNodeProcessDirectoryRuntime_001(t *testing.T) {
	t.Parallel()
	nodeProcessDirectoryRuntime(t, 1)
}
func TestNodeProcessDirectoryRuntime_002(t *testing.T) {
	t.Parallel()
	nodeProcessDirectoryRuntime(t, 2)
}
func TestNodeProcessDirectoryRuntime_003(t *testing.T) {
	t.Parallel()
	nodeProcessDirectoryRuntime(t, 3)
}
func TestNodeProcessDirectoryRuntime_004(t *testing.T) {
	t.Parallel()
	nodeProcessDirectoryRuntime(t, 4)
}

func nodeProcessDirectoryRuntime(t *testing.T, shard int) {
	// Canonical, so cwd() reads back what chdir was given: macOS's /tmp is a link to /private/tmp.
	scratch, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	child, missing := filepath.Join(scratch, "child"), filepath.Join(scratch, "absent")
	if err := os.Mkdir(child, 0700); err != nil {
		t.Fatal(err)
	}
	source, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/node_process_runtime/node_process_directory_mutation.a"))
	if err != nil {
		t.Fatal(err)
	}
	truth := execute(t, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), source, scratch, child, missing)
	if truth.exitCode != 0 || len(truth.stderr) != 0 || !strings.HasPrefix(string(truth.stdout), "true\ntrue\ntrue\nENOENT\nENOENT:") {
		t.Fatalf("unexpected Node result: %+v", truth)
	}
	program := &ir.Program{Source: filepath.Base(source), Strings: []string{scratch, child, missing}, Locals: []ir.Local{{Name: "memoized", Type: ir.String, Function: -1}, {Name: "error", Type: ir.Object, Function: -1}}}
	cwd := ir.ProcessCall{Operation: "cwd", Of: ir.String}
	change := func(index int) ir.Statement {
		return ir.Evaluate{Value: ir.ProcessCall{Operation: "chdir", Of: ir.Object, Arguments: []ir.Expression{ir.StringConstant{Index: index}}}}
	}
	equal := func(value ir.Expression, index int) ir.Statement {
		return ir.WriteLine{Stream: ir.Stdout, Value: ir.BooleanToString{Value: ir.Binary{Operator: ir.Equal, Left: value, Right: ir.StringConstant{Index: index}}}}
	}
	errorValue := ir.Read{Local: 1, Of: ir.Object}
	program.Main = []ir.Statement{
		ir.Evaluate{Value: cwd}, change(0), equal(cwd, 0), ir.Declare{Local: 0, Value: cwd}, change(1), equal(cwd, 1), equal(ir.Read{Local: 0, Of: ir.String}, 0),
		ir.Try{HasCatch: true, CatchLocal: 1, Body: []ir.Statement{change(2)}, Catch: []ir.Statement{
			ir.WriteLine{Stream: ir.Stdout, Value: ir.Property{Object: errorValue, Name: "code", Of: ir.String}},
			ir.WriteLine{Stream: ir.Stdout, Value: ir.Property{Object: errorValue, Name: "message", Of: ir.String}},
		}},
	}
	if shard < 3 {
		nodeProcessRuntimeAgreement(t, program, truth, shard)
		return
	}
	data, err := os.ReadFile(filepath.Join(repository, "internal/native/runtime/node_process.c"))
	if err != nil {
		t.Fatal(err)
	}
	for name, anchor := range map[string]string{"cache invalidation": "adamic_release(current_directory); current_directory = NULL;", "error code": "case ENOENT: code = \"ENOENT\";"} {
		if (name == "cache invalidation") != (shard == 3) {
			continue
		}
		t.Run(name, func(t *testing.T) {
			replacement := "(void)current_directory;"
			if name == "error code" {
				replacement = "case ENOENT: code = \"ENOTDIR\";"
			}
			if strings.Count(string(data), anchor) != 1 {
				t.Fatal("mutant anchor changed")
			}
			runtime := strings.Replace(string(data), anchor, replacement, 1)
			runtime = nodeProcessMutantNames(runtime, "mutant_node_")
			code := nodeProcessMutantNames(native.C(program), "mutant_node_")
			mutant := filepath.Join(t.TempDir(), "mutant")
			if err := native.Build(runtime+"\n"+code, mutant, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			bad := execute(t, mutant)
			if bad.exitCode != 0 || len(bad.stderr) != 0 || disagreement(truth, bad) != "stdout differs" {
				t.Fatal("mutant not caught only by Node stdout comparison")
			}
			t.Log("clean runtime mutant caught by Node stdout comparison")
		})
	}
}

func TestNodeProcessDirectorySymlinksRuntime_000(t *testing.T) {
	t.Parallel()
	nodeProcessDirectorySymlinksRuntime(t, 0)
}
func TestNodeProcessDirectorySymlinksRuntime_001(t *testing.T) {
	t.Parallel()
	nodeProcessDirectorySymlinksRuntime(t, 1)
}
func TestNodeProcessDirectorySymlinksRuntime_002(t *testing.T) {
	t.Parallel()
	nodeProcessDirectorySymlinksRuntime(t, 2)
}
func TestNodeProcessDirectorySymlinksRuntime_003(t *testing.T) {
	t.Parallel()
	nodeProcessDirectorySymlinksRuntime(t, 3)
}

func nodeProcessDirectorySymlinksRuntime(t *testing.T, shard int) {
	scratch, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	child, link := filepath.Join(scratch, "child"), filepath.Join(scratch, "link")
	if err := os.Mkdir(child, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("child", link); err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(t.TempDir(), "main.a")
	if err := os.WriteFile(source, []byte("process.chdir(process.argv[2]); console.log(process.cwd()); process.chdir('..'); console.log(process.cwd());"), 0600); err != nil {
		t.Fatal(err)
	}
	truth := onNodeWith(t, inputRun{arguments: []string{link}}, source)
	program := &ir.Program{Source: "process-directory-symlinks.a", Strings: []string{link, ".."}}
	for index := range 2 {
		program.Main = append(program.Main, ir.Evaluate{Value: ir.ProcessCall{Operation: "chdir", Arguments: []ir.Expression{ir.StringConstant{Index: index}}, Of: ir.Object}}, ir.WriteLine{Stream: ir.Stdout, Value: ir.ProcessCall{Operation: "cwd", Of: ir.String}})
	}
	if shard < 3 {
		nodeProcessRuntimeAgreement(t, program, truth, shard)
		return
	}
	if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		t.Skip("set ADAMIC_ORACLE_WASI=1")
	}
	{
		data, err := os.ReadFile(filepath.Join(repository, "internal/native/runtime/node_process.c"))
		if err != nil {
			t.Fatal(err)
		}
		anchor := "char *resolved = realpath(path, NULL);"
		if strings.Count(string(data), anchor) != 1 {
			t.Fatal("physical cwd mutant anchor changed")
		}
		runtime := strings.Replace(string(data), anchor, "char *resolved = strdup(path);", 1)
		runtime = nodeProcessMutantNames(runtime, "mutant_node_")
		code := nodeProcessMutantNames(native.C(program), "mutant_node_")
		bad := onWASI(t, runtime+"\n"+code)
		if bad.exitCode != 0 || len(bad.stderr) != 0 || disagreement(truth, bad) != "stdout differs" {
			t.Fatalf("virtual cwd mutant failed outside Node comparison: %+v", bad)
		}
		t.Log("virtual symlink cwd mutant ran cleanly and was caught only by Node stdout")
	}
}

// Separate runtime options are Loom units: a cold archive belongs to one shard,
// rather than accumulating sanitized, release and WASI builds in one test.
// All shards observe Node independently. Native checks reuse their sanitized
// binary for LeakSanitizer; macOS retains its counted leak check.
func nodeProcessRuntimeAgreement(t *testing.T, program *ir.Program, truth run, shard int) {
	t.Helper()
	switch shard {
	case 0:
		got, binary := nativelyUncached(t, program)
		for _, result := range []run{got, nodeProcessJavaScript(t, program)} {
			if difference := disagreement(truth, result); difference != "" {
				t.Fatal(difference)
			}
		}
		if report := leaksUncached(t, program, binary); report != "" {
			t.Fatal(report)
		}
	case 1:
		if difference := disagreement(truth, releasedUncached(t, program)); difference != "" {
			t.Fatal(difference)
		}
	case 2:
		if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
			t.Skip("set ADAMIC_ORACLE_WASI=1")
		}
		if difference := disagreement(truth, onWASI(t, native.C(program))); difference != "" {
			t.Fatal(difference)
		}
	default:
		t.Fatalf("unknown runtime agreement shard %d", shard)
	}
}

// These proofs always execute Node. The cached Node helper's identity eagerly
// builds three native archives, unrelated to a JavaScript or WASI comparison.
func nodeProcessJavaScript(t *testing.T, program *ir.Program) run {
	t.Helper()
	path := filepath.Join(t.TempDir(), "program.mjs")
	if err := os.WriteFile(path, []byte(javascript.JavaScript(program)), 0600); err != nil {
		t.Fatal(err)
	}
	return onNodeWith(t, inputRun{}, path)
}
