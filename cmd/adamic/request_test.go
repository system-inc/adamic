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

func requestSource(test *testing.T, source string) string {
	test.Helper()
	path := filepath.Join(test.TempDir(), "program.a")
	if writeError := os.WriteFile(path, []byte(source), 0644); writeError != nil {
		test.Fatal(writeError)
	}
	return path
}

// Not parallel: the command and checker use process-wide diagnostics and environment.
func TestWASIRequestSelection(test *testing.T) {
	for _, testCase := range []struct {
		name, source     string
		handler, refused bool
	}{
		{"exported", "export function handleRequest(request: string): string { return request; }", true, false},
		{
			"alias",
			"function serve(request: string): string { return request; } export { serve as handleRequest };",
			true,
			false,
		},
		{
			"alias wrong parameter",
			"function serve(request: number): string { return `${request}`; } export { serve as handleRequest };",
			false,
			true,
		},
		{
			"alias wrong result",
			"function serve(request: string): number { return request.length; } export { serve as handleRequest };",
			false,
			true,
		},
		{"private", "function handleRequest(request: string): string { return request; }", false, false},
		{
			"wrong parameter",
			"export function handleRequest(request: number): string { return `${request}`; }",
			false,
			true,
		},
		{
			"wrong result",
			"export function handleRequest(request: string): number { return request.length; }",
			false,
			true,
		},
		{
			"literal parameter",
			"export function handleRequest(request: 'only'): string { return request; }",
			false,
			true,
		},
		{
			"optional parameter",
			"export function handleRequest(request?: string): string { return request ?? ''; }",
			false,
			true,
		},
		{
			"default parameter",
			"export function handleRequest(request: string = ''): string { return request; }",
			false,
			true,
		},
		{
			"two parameters",
			"export function handleRequest(request: string, other: string): string { return request + other; }",
			false,
			true,
		},
		{"generic", "export function handleRequest<T>(request: string): string { return request; }", false, true},
		{"function value", "export const handleRequest = (request: string): string => request;", false, true},
	} {
		test.Run(testCase.name, func(test *testing.T) {
			path := requestSource(test, testCase.source)
			checked, checkCode := check([]string{path})
			if checked == nil || checkCode != 0 {
				test.Fatalf("checker refused handler fixture, code %d", checkCode)
			}
			name, requestFunctionError := requestFunction(checked)
			invalidSignature := testCase.refused
			if (requestFunctionError != nil) != invalidSignature {
				test.Fatalf(
					"signature selection returned %q, %v, want refused %t",
					name,
					requestFunctionError,
					invalidSignature,
				)
			}
			if testCase.name == "alias" && name != "serve" {
				test.Fatalf("alias selected %q, want serve", name)
			}
			program, handler, code := compileWASI(path)
			if testCase.refused {
				if code != 1 || program != nil {
					test.Fatalf("accepted invalid handler, code %d", code)
				}
				return
			}
			if program == nil || code != 0 {
				test.Fatalf("compile failed, code %d", code)
			}
			if (handler >= 0) != testCase.handler {
				test.Fatalf("handler = %d, want present %t", handler, testCase.handler)
			}
		})
	}
}

// Not parallel: this calls the command in-process, whose checker and diagnostics are shared.
func TestWASIRequest(test *testing.T) {
	if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		test.Skip("set ADAMIC_ORACLE_WASI=1")
	}
	for _, fixture := range []string{"request.a", "request_alias.a"} {
		test.Run(fixture, func(test *testing.T) {
			output := filepath.Join(test.TempDir(), "request.wasm")
			source, absolutePathError := filepath.Abs(filepath.Join("testdata/wasi", fixture))
			if absolutePathError != nil {
				test.Fatal(absolutePathError)
			}
			if code := run([]string{"build", "--target", "wasm32-wasi", source, "-o", output, "--count"}); code != 0 {
				test.Fatalf("build returned %d", code)
			}
			timeoutContext, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			command := exec.CommandContext(
				timeoutContext,
				"node",
				"--disable-warning=ExperimentalWarning",
				"testdata/wasi/request-host.mjs",
				output,
				source,
				filepath.Join(test.TempDir(), "stdout"),
			)
			if result, combinedOutputError := command.CombinedOutput(); combinedOutputError != nil {
				test.Fatalf("request host: %v\n%s", combinedOutputError, result)
			} else {
				test.Log(string(result))
			}
		})
	}
}

// Not parallel: native.Build writes native.runtimeBuilds and the shared os.UserCacheDir()/adamic/runtime cache.
func TestRequestNativeStaysCommand(test *testing.T) {
	program, code := compile("testdata/wasi/request.a")
	if code != 0 {
		test.Fatalf("compile returned %d", code)
	}
	source := native.C(program)
	if strings.Contains(source, "adamic_request(") {
		test.Fatal("native C acquired request ABI")
	}
	binary := filepath.Join(test.TempDir(), "program")
	if buildError := native.Build(source, binary, native.Options{Sanitize: true}); buildError != nil {
		test.Fatal(buildError)
	}
	result, combinedOutputError := exec.Command(binary).CombinedOutput()
	if combinedOutputError != nil || string(result) != "dependency initialized\nentry initialized\n" {
		test.Fatalf("native command: %v, %q", combinedOutputError, result)
	}
}

// Not parallel: uses the command's shared checker and diagnostics.
func TestWASIRequestModuleSelection(test *testing.T) {
	dir := test.TempDir()
	dependency := filepath.Join(dir, "dependency.a")
	if writeError := os.WriteFile(
		dependency,
		[]byte("export function handleRequest(request: string): string { return request; }"),
		0644,
	); writeError != nil {
		test.Fatal(writeError)
	}
	path := filepath.Join(dir, "program.a")
	for _, testCase := range []struct {
		source  string
		refused bool
	}{
		{"import { handleRequest } from './dependency.a'; console.log(handleRequest('dependency'));", false},
		{
			"import { handleRequest as dependency } from './dependency.a'; export function handleRequest(request: string): string { return dependency(request); }",
			true,
		},
	} {
		if writeError := os.WriteFile(path, []byte(testCase.source), 0644); writeError != nil {
			test.Fatal(writeError)
		}
		program, handler, code := compileWASI(path)
		if testCase.refused {
			if code != 1 || program != nil {
				test.Fatal("accepted ambiguous IR function names")
			}
		} else if code != 0 || program == nil || handler != -1 {
			test.Fatalf("dependency selected as handler: code %d, handler %d", code, handler)
		}
	}
}

// Not parallel: uses the command's shared checker and diagnostics.
func TestWASIRequestThrows(test *testing.T) {
	if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		test.Skip("set ADAMIC_ORACLE_WASI=1")
	}
	output := filepath.Join(test.TempDir(), "request.wasm")
	source := requestSource(
		test,
		"export function handleRequest(request: string): string { throw new Error(request); }",
	)
	if code := run([]string{"build", "--target", "wasm32-wasi", source, "-o", output}); code != 0 {
		test.Fatalf("build returned %d", code)
	}
	timeoutContext, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	command := exec.CommandContext(
		timeoutContext,
		"node",
		"--disable-warning=ExperimentalWarning",
		"testdata/wasi/throw-host.mjs",
		output,
	)
	result, combinedOutputError := command.CombinedOutput()
	exit, ok := combinedOutputError.(*exec.ExitError)
	if !ok || exit.ExitCode() != 70 || string(result) != "adamic: panic: Error: boundary\n" {
		test.Fatalf("uncaught handler exception: %v, %q", combinedOutputError, result)
	}
}
