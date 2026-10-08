package load

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/bundled"
	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/tsoptions"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/cachedvfs"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
)

// Use the same options and source view with the unmodified bundled declarations.
// Capture specialization must preserve errors this checker reports, even when
// CaptureFacts could justify a narrower type than the indexed library result.
func stockCaptureDiagnostics(t *testing.T, paths []string) []string {
	t.Helper()
	directory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	current := tspath.RootedDirectoryPathFromAbsolute(directory)
	fs := &sourceFS{FS: osvfs.FS()}
	roots := make([]tspath.RootedFilePath, 0, len(paths)+1)
	for _, path := range paths {
		root, err := rootFileName(fs, current, path)
		if err != nil {
			t.Fatal(err)
		}
		roots = append(roots, root)
	}
	roots = append(roots, preludePath)
	files := cachedvfs.From(bundled.WrapFS(fs))
	config := tsoptions.NewParsedCommandLine(compilerOptions(), roots, nil, current, files.CaseSensitivity())
	host := compiler.NewCachedFSCompilerHost(files, bundled.LibPath(), nil, nil, nil)
	program := compiler.NewProgram(compiler.ProgramOptions{Config: config, Host: host, SingleThreaded: core.TSTrue})
	if program == nil {
		t.Fatal("stock checker built no program")
	}
	return (&Program{compiler: program, fs: fs}).diagnostics(context.Background())
}

func TestRegExpStockDiagnosticBoundary(t *testing.T) {
	for _, source := range []string{
		`const result=/(a)/.exec('a');if(result){const capture:string=result[1];console.log(capture);}`,
		`const result='a'.match(/(a)/);if(result){const capture:string=result[1];console.log(capture);}`,
		`for(const result of 'a'.matchAll(/(a)/g)){const capture:string=result[1];console.log(capture);}`,
	} {
		paths := writeProgram(t, [2]string{"main.a", source})
		stock := stockCaptureDiagnostics(t, paths)
		if len(stock) != 1 || !strings.Contains(stock[0], "TS2322") {
			t.Fatalf("stock witness changed: %v", stock)
		}
		diagnostics := checkErrors(t, paths)
		if len(diagnostics) != 1 || !strings.Contains(diagnostics[0], "TS2322") {
			t.Fatalf("stock error disappeared: %v", diagnostics)
		}
	}
}

func TestRuntimeRegExpCaptureBoundary(t *testing.T) {
	for _, source := range []string{
		`function read(pattern:string){const result=new RegExp(pattern).exec('');if(result){const capture=result[1];console.log(capture??'absent');}}`,
		`function read(pattern:string){const result=''.match(new RegExp(pattern));if(result){const capture=result[1];console.log(capture??'absent');}}`,
		`function read(pattern:string){for(const result of ''.matchAll(new RegExp(pattern,'g'))){const capture=result[1];console.log(capture??'absent');}}`,
		`function read(pattern:string){const result=new RegExp(pattern).exec('');if(result){const capture=result.groups?.missing;console.log(capture??'absent');}}`,
	} {
		paths := writeProgram(t, [2]string{"main.a", source})
		got := declarationsByName(t, paths)["variable capture"].Type
		if got != "string | undefined" {
			t.Fatalf("runtime pattern capture: %q", got)
		}
	}
}
