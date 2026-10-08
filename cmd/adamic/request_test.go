package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/native"
)

func requestSource(t *testing.T, source string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "program.a")
	if err := os.WriteFile(path, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

// Not parallel: the command and checker use process-wide diagnostics and environment.
func TestWASIRequestSelection(t *testing.T) {
	for _, test := range []struct {
		name, source     string
		handler, refused bool
	}{
		{"exported", "export function handleRequest(request: string): string { return request; }", true, false},
		{"alias", "function serve(request: string): string { return request; } export { serve as handleRequest };", true, false},
		{"alias wrong parameter", "function serve(request: number): string { return `${request}`; } export { serve as handleRequest };", false, true},
		{"alias wrong result", "function serve(request: string): number { return request.length; } export { serve as handleRequest };", false, true},
		{"private", "function handleRequest(request: string): string { return request; }", false, false},
		{"wrong parameter", "export function handleRequest(request: number): string { return `${request}`; }", false, true},
		{"wrong result", "export function handleRequest(request: string): number { return request.length; }", false, true},
		{"literal parameter", "export function handleRequest(request: 'only'): string { return request; }", false, true},
		{"optional parameter", "export function handleRequest(request?: string): string { return request ?? ''; }", false, true},
		{"default parameter", "export function handleRequest(request: string = ''): string { return request; }", false, true},
		{"two parameters", "export function handleRequest(request: string, other: string): string { return request + other; }", false, true},
		{"generic", "export function handleRequest<T>(request: string): string { return request; }", false, true},
		{"function value", "export const handleRequest = (request: string): string => request;", false, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			path := requestSource(t, test.source)
			checked, checkCode := check([]string{path})
			if checked == nil || checkCode != 0 {
				t.Fatalf("checker refused handler fixture, code %d", checkCode)
			}
			name, err := requestFunction(checked)
			invalidSignature := test.refused
			if (err != nil) != invalidSignature {
				t.Fatalf("signature selection returned %q, %v, want refused %t", name, err, invalidSignature)
			}
			if test.name == "alias" && name != "serve" {
				t.Fatalf("alias selected %q, want serve", name)
			}
			program, handler, code := compileWASI(path)
			if test.refused {
				if code != 1 || program != nil {
					t.Fatalf("accepted invalid handler, code %d", code)
				}
				return
			}
			if program == nil || code != 0 {
				t.Fatalf("compile failed, code %d", code)
			}
			if (handler >= 0) != test.handler {
				t.Fatalf("handler = %d, want present %t", handler, test.handler)
			}
		})
	}
}

// Not parallel: this calls the command in-process, whose checker and diagnostics are shared.
func TestWASIRequest(t *testing.T) {
	if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		t.Skip("set ADAMIC_ORACLE_WASI=1")
	}
	for _, fixture := range []string{"request.a", "request_alias.a"} {
		t.Run(fixture, func(t *testing.T) {
			output := filepath.Join(t.TempDir(), "request.wasm")
			source, err := filepath.Abs(filepath.Join("testdata/wasi", fixture))
			if err != nil {
				t.Fatal(err)
			}
			if code := run([]string{"build", "--target", "wasm32-wasi", source, "-o", output, "--count"}); code != 0 {
				t.Fatalf("build returned %d", code)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			command := exec.CommandContext(ctx, "node", "--disable-warning=ExperimentalWarning", "testdata/wasi/request-host.mjs", output, source, filepath.Join(t.TempDir(), "stdout"))
			if result, err := command.CombinedOutput(); err != nil {
				t.Fatalf("request host: %v\n%s", err, result)
			} else {
				t.Log(string(result))
			}
		})
	}
}

func TestRequestNativeStaysCommand(t *testing.T) {
	program, code := compile("testdata/wasi/request.a")
	if code != 0 {
		t.Fatalf("compile returned %d", code)
	}
	source := native.C(program)
	if strings.Contains(source, "adamic_request(") {
		t.Fatal("native C acquired request ABI")
	}
	binary := filepath.Join(t.TempDir(), "program")
	if err := native.Build(source, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	result, err := exec.Command(binary).CombinedOutput()
	if err != nil || string(result) != "dependency initialized\nentry initialized\n" {
		t.Fatalf("native command: %v, %q", err, result)
	}
}

// Not parallel: uses the command's shared checker and diagnostics.
func TestWASIRequestModuleSelection(t *testing.T) {
	dir := t.TempDir()
	dependency := filepath.Join(dir, "dependency.a")
	if err := os.WriteFile(dependency, []byte("export function handleRequest(request: string): string { return request; }"), 0644); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "program.a")
	for _, test := range []struct {
		source  string
		refused bool
	}{
		{"import { handleRequest } from './dependency.a'; console.log(handleRequest('dependency'));", false},
		{"import { handleRequest as dependency } from './dependency.a'; export function handleRequest(request: string): string { return dependency(request); }", true},
	} {
		if err := os.WriteFile(path, []byte(test.source), 0644); err != nil {
			t.Fatal(err)
		}
		program, handler, code := compileWASI(path)
		if test.refused {
			if code != 1 || program != nil {
				t.Fatal("accepted ambiguous IR function names")
			}
		} else if code != 0 || program == nil || handler != -1 {
			t.Fatalf("dependency selected as handler: code %d, handler %d", code, handler)
		}
	}
}

// Not parallel: uses the command's shared checker and diagnostics.
func TestWASIRequestThrows(t *testing.T) {
	if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		t.Skip("set ADAMIC_ORACLE_WASI=1")
	}
	output := filepath.Join(t.TempDir(), "request.wasm")
	source := requestSource(t, "export function handleRequest(request: string): string { throw new Error(request); }")
	if code := run([]string{"build", "--target", "wasm32-wasi", source, "-o", output}); code != 0 {
		t.Fatalf("build returned %d", code)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, "node", "--disable-warning=ExperimentalWarning", "testdata/wasi/throw-host.mjs", output)
	result, err := command.CombinedOutput()
	exit, ok := err.(*exec.ExitError)
	if !ok || exit.ExitCode() != 70 || string(result) != "adamic: panic: Error: boundary\n" {
		t.Fatalf("uncaught handler exception: %v, %q", err, result)
	}
}
