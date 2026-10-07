// Compiled only as an overlay in the pinned cohere nexus package. This calls
// production private helpers unchanged; it is not a shared harness edit.
package nexus

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/bundled"
	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/tsoptions"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
	"github.com/system-inc/cohere/internal/lint/rule"
)

func TestWave10NextNativeAtomsAgainstProduction(t *testing.T) {
	binary := os.Getenv("ADAMIC_WAVE10_NEXT_ATOMS")
	if binary == "" {
		t.Fatal("missing native atom binary")
	}
	directory := t.TempDir()
	entry := filepath.Join(directory, "entry.a")
	stream := filepath.Join(directory, "source/system/StandardStreams.d.ts")
	configPath := filepath.Join(directory, "tsconfig.json")
	if err := os.MkdirAll(filepath.Dir(stream), 0755); err != nil {
		t.Fatal(err)
	}
	source := `import './source/system/StandardStreams';new Promise((_r,reject)=>{setTimeout(reject,10);});`
	for p, text := range map[string]string{entry: source, stream: `export declare function blockStandardStreams():void;`, configPath: `{"compilerOptions":{"noLib":true},"files":["source/system/StandardStreams.d.ts"]}`} {
		if err := os.WriteFile(p, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	host := compiler.NewCachedFSCompilerHost(directory, bundled.WrapFS(osvfs.FS()), bundled.LibPath(), nil, nil, nil)
	config, diagnostics := tsoptions.GetParsedCommandLineOfConfigFile(configPath, nil, nil, host, nil)
	if config == nil || len(diagnostics) > 0 || len(config.Errors) > 0 {
		t.Fatal("invalid fixture config")
	}
	config.CompilerOptions().AllowNonTsExtensions = core.TSTrue
	config = config.WithFileNames([]string{entry, stream})
	program := compiler.NewProgram(compiler.ProgramOptions{Config: config, Host: host, SingleThreaded: core.TSTrue})
	file := program.GetSourceFile(entry)
	if file == nil {
		t.Fatal("missing fixture source")
	}
	c, release := program.GetTypeCheckerForFile(context.Background(), file)
	defer release()
	ctx := rule.Context{SourceFile: file, TypeChecker: c, Program: rule.ViewProgram(program, file, CorrectnessRequireBlockingStandardStreams)}
	index := correctnessRequireBlockingStandardStreamsBuildIndex(ctx.Program)
	var timer, executor *ast.Node
	var walk func(*ast.Node)
	walk = func(n *ast.Node) {
		if n.Kind == ast.KindArrowFunction {
			executor = n
		}
		if n.Kind == ast.KindCallExpression && n.AsCallExpression().Expression.Kind == ast.KindIdentifier && n.AsCallExpression().Expression.Text() == "setTimeout" {
			timer = n
		}
		n.ForEachChild(func(child *ast.Node) bool { walk(child); return false })
	}
	walk(file.AsNode())
	lost := correctnessNoUnclearedRaceTimeoutHandleIsLost(ctx, executor, timer)
	reach := "does not reach streams"
	if index.canBlock[file.Path()] {
		reach = "reaches streams"
	}
	imported := "entry"
	if index.imported[program.GetSourceFile(stream).Path()] {
		imported = "imported"
	}
	handle := "kept"
	if lost {
		handle = "lost"
	}
	expected := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n", correctnessNoProcessExitAfterOutputState{",1,2,"}.with(",1,").key(), correctnessNoProcessExitAfterOutputState{",1,", ","}.entering(1).key(), reach, imported, handle)
	var output bytes.Buffer
	for _, executable := range strings.Split(binary, ":") {
		command := exec.Command(executable, configPath, entry, "--atoms")
		got, err := command.CombinedOutput()
		if err != nil {
			t.Fatalf("native atom execution: %v: %s", err, got)
		}
		output.Write(got)
	}
	got := output.Bytes()
	if string(got) != expected {
		t.Fatalf("native atoms mismatch: got %q Go %q", got, expected)
	}
	t.Logf("five native atom lines equal production Go bytes: %q", expected)
}
