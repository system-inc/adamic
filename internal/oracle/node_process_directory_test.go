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
func TestNodeProcessDirectoryRuntime_000(test *testing.T) {
	test.Parallel()
	nodeProcessDirectoryRuntime(test, 0)
}
func TestNodeProcessDirectoryRuntime_001(test *testing.T) {
	test.Parallel()
	nodeProcessDirectoryRuntime(test, 1)
}
func TestNodeProcessDirectoryRuntime_002(test *testing.T) {
	test.Parallel()
	nodeProcessDirectoryRuntime(test, 2)
}
func TestNodeProcessDirectoryRuntime_003(test *testing.T) {
	test.Parallel()
	nodeProcessDirectoryRuntime(test, 3)
}
func TestNodeProcessDirectoryRuntime_004(test *testing.T) {
	test.Parallel()
	nodeProcessDirectoryRuntime(test, 4)
}

func nodeProcessDirectoryRuntime(test *testing.T, shard int) {
	// Canonical, so cwd() reads back what chdir was given: macOS's /tmp is a link to /private/tmp.
	scratch, err := filepath.EvalSymlinks(test.TempDir())
	if err != nil {
		test.Fatal(err)
	}
	child, missing := filepath.Join(scratch, "child"), filepath.Join(scratch, "absent")
	if err := os.Mkdir(child, 0700); err != nil {
		test.Fatal(err)
	}
	source, err := filepath.Abs(
		filepath.Join(repository, "internal/oracle/testdata/node_process_runtime/node_process_directory_mutation.a"),
	)
	if err != nil {
		test.Fatal(err)
	}
	truth := execute(
		test,
		"node",
		"--disable-warning=ExperimentalWarning",
		filepath.Join(repository, "oracle/node.mjs"),
		source,
		scratch,
		child,
		missing,
	)
	if truth.exitCode != 0 ||
		len(truth.stderr) != 0 ||
		!strings.HasPrefix(string(truth.stdout), "true\ntrue\ntrue\nENOENT\nENOENT:") {
		test.Fatalf("unexpected Node result: %+v", truth)
	}
	program := &ir.Program{
		Source:  filepath.Base(source),
		Strings: []string{scratch, child, missing},
		Locals: []ir.Local{
			{Name: "memoized", Type: ir.String, Function: -1},
			{Name: "error", Type: ir.Object, Function: -1},
		},
	}
	cwd := ir.ProcessCall{Operation: "cwd", Of: ir.String}
	change := func(index int) ir.Statement {
		return ir.Evaluate{Value: ir.ProcessCall{
			Operation: "chdir",
			Of:        ir.Object,
			Arguments: []ir.Expression{ir.StringConstant{Index: index}},
		}}
	}
	equal := func(value ir.Expression, index int) ir.Statement {
		return ir.WriteLine{Stream: ir.Stdout, Value: ir.BooleanToString{
			Value: ir.Binary{Operator: ir.Equal, Left: value, Right: ir.StringConstant{Index: index}},
		}}
	}
	errorValue := ir.Read{Local: 1, Of: ir.Object}
	program.Main = []ir.Statement{
		ir.Evaluate{Value: cwd},
		change(0),
		equal(cwd, 0),
		ir.Declare{Local: 0, Value: cwd},
		change(1),
		equal(cwd, 1),
		equal(ir.Read{Local: 0, Of: ir.String}, 0),
		ir.Try{HasCatch: true, CatchLocal: 1, Body: []ir.Statement{change(2)}, Catch: []ir.Statement{
			ir.WriteLine{Stream: ir.Stdout, Value: ir.Property{Object: errorValue, Name: "code", Of: ir.String}},
			ir.WriteLine{Stream: ir.Stdout, Value: ir.Property{Object: errorValue, Name: "message", Of: ir.String}},
		}},
	}
	if shard < 3 {
		nodeProcessRuntimeAgreement(test, program, truth, shard)
		return
	}
	data, err := os.ReadFile(filepath.Join(repository, "internal/native/runtime/node_process.c"))
	if err != nil {
		test.Fatal(err)
	}
	for name, anchor := range map[string]string{
		"cache invalidation": "adamic_release(current_directory); current_directory = NULL;",
		"error code":         "case ENOENT: code = \"ENOENT\";",
	} {
		if (name == "cache invalidation") != (shard == 3) {
			continue
		}
		test.Run(name, func(test *testing.T) {
			replacement := "(void)current_directory;"
			if name == "error code" {
				replacement = "case ENOENT: code = \"ENOTDIR\";"
			}
			if strings.Count(string(data), anchor) != 1 {
				test.Fatal("mutant anchor changed")
			}
			runtime := strings.Replace(string(data), anchor, replacement, 1)
			runtime = nodeProcessMutantNames(runtime, "mutant_node_")
			code := nodeProcessMutantNames(native.C(program), "mutant_node_")
			mutant := filepath.Join(test.TempDir(), "mutant")
			if err := native.Build(runtime+"\n"+code, mutant, native.Options{Sanitize: true}); err != nil {
				test.Fatal(err)
			}
			bad := execute(test, mutant)
			if bad.exitCode != 0 || len(bad.stderr) != 0 || disagreement(truth, bad) != "stdout differs" {
				test.Fatal("mutant not caught only by Node stdout comparison")
			}
			test.Log("clean runtime mutant caught by Node stdout comparison")
		})
	}
}

func TestNodeProcessDirectorySymlinksRuntime_000(test *testing.T) {
	test.Parallel()
	nodeProcessDirectorySymlinksRuntime(test, 0)
}
func TestNodeProcessDirectorySymlinksRuntime_001(test *testing.T) {
	test.Parallel()
	nodeProcessDirectorySymlinksRuntime(test, 1)
}
func TestNodeProcessDirectorySymlinksRuntime_002(test *testing.T) {
	test.Parallel()
	nodeProcessDirectorySymlinksRuntime(test, 2)
}
func TestNodeProcessDirectorySymlinksRuntime_003(test *testing.T) {
	test.Parallel()
	nodeProcessDirectorySymlinksRuntime(test, 3)
}

func nodeProcessDirectorySymlinksRuntime(test *testing.T, shard int) {
	scratch, err := filepath.EvalSymlinks(test.TempDir())
	if err != nil {
		test.Fatal(err)
	}
	child, link := filepath.Join(scratch, "child"), filepath.Join(scratch, "link")
	if err := os.Mkdir(child, 0700); err != nil {
		test.Fatal(err)
	}
	if err := os.Symlink("child", link); err != nil {
		test.Fatal(err)
	}
	source := filepath.Join(test.TempDir(), "main.a")
	if err := os.WriteFile(source, []byte(
		"process.chdir(process.argv[2]); console.log(process.cwd()); process.chdir('..'); console.log(process.cwd());",
	), 0600); err != nil {
		test.Fatal(err)
	}
	truth := onNodeWith(test, inputRun{arguments: []string{link}}, source)
	program := &ir.Program{Source: "process-directory-symlinks.a", Strings: []string{link, ".."}}
	for index := range 2 {
		program.Main = append(program.Main, ir.Evaluate{Value: ir.ProcessCall{
			Operation: "chdir",
			Arguments: []ir.Expression{ir.StringConstant{Index: index}},
			Of:        ir.Object,
		}}, ir.WriteLine{Stream: ir.Stdout, Value: ir.ProcessCall{Operation: "cwd", Of: ir.String}})
	}
	if shard < 3 {
		nodeProcessRuntimeAgreement(test, program, truth, shard)
		return
	}
	if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		test.Skip("set ADAMIC_ORACLE_WASI=1")
	}
	{
		data, err := os.ReadFile(filepath.Join(repository, "internal/native/runtime/node_process.c"))
		if err != nil {
			test.Fatal(err)
		}
		anchor := "char *resolved = realpath(path, NULL);"
		if strings.Count(string(data), anchor) != 1 {
			test.Fatal("physical cwd mutant anchor changed")
		}
		runtime := strings.Replace(string(data), anchor, "char *resolved = strdup(path);", 1)
		runtime = nodeProcessMutantNames(runtime, "mutant_node_")
		code := nodeProcessMutantNames(native.C(program), "mutant_node_")
		bad := onWASI(test, runtime+"\n"+code)
		if bad.exitCode != 0 || len(bad.stderr) != 0 || disagreement(truth, bad) != "stdout differs" {
			test.Fatalf("virtual cwd mutant failed outside Node comparison: %+v", bad)
		}
		test.Log("virtual symlink cwd mutant ran cleanly and was caught only by Node stdout")
	}
}

// Separate runtime options are Loom units: a cold archive belongs to one shard,
// rather than accumulating sanitized, release and WASI builds in one test.
// All shards observe Node independently. Native checks reuse their sanitized
// binary for LeakSanitizer; macOS retains its counted leak check.
func nodeProcessRuntimeAgreement(test *testing.T, program *ir.Program, truth run, shard int) {
	test.Helper()
	switch shard {
	case 0:
		got, binary := nativelyUncached(test, program)
		for _, result := range []run{got, nodeProcessJavaScript(test, program)} {
			if difference := disagreement(truth, result); difference != "" {
				test.Fatal(difference)
			}
		}
		if report := leaksUncached(test, program, binary); report != "" {
			test.Fatal(report)
		}
	case 1:
		if difference := disagreement(truth, releasedUncached(test, program)); difference != "" {
			test.Fatal(difference)
		}
	case 2:
		if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
			test.Skip("set ADAMIC_ORACLE_WASI=1")
		}
		if difference := disagreement(truth, onWASI(test, native.C(program))); difference != "" {
			test.Fatal(difference)
		}
	default:
		test.Fatalf("unknown runtime agreement shard %d", shard)
	}
}

// These proofs always execute Node. The cached Node helper's identity eagerly
// builds three native archives, unrelated to a JavaScript or WASI comparison.
func nodeProcessJavaScript(test *testing.T, program *ir.Program) run {
	test.Helper()
	path := filepath.Join(test.TempDir(), "program.mjs")
	if err := os.WriteFile(path, []byte(javascript.JavaScript(program)), 0600); err != nil {
		test.Fatal(err)
	}
	return onNodeWith(test, inputRun{}, path)
}
