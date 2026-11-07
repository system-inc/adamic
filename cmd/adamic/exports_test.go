package main

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Not parallel: checker diagnostics and in-process command state are shared.
func TestWASIExportSelection(test *testing.T) {
	tests := []struct{ source, path string }{
		{`export function bad(value: readonly boolean[]): void {}`, "bad.value"},
		{
			`interface R { readonly outer: { readonly middle: { readonly x: number } } } export function bad(value: R): void {}`,
			"bad.value.outer.middle",
		},
		{`interface R { readonly x: number | string } export function bad(value: R): void {}`, "bad.value.x"},
		{`interface R { x: number } export function bad(value: R): void {}`, "bad.value.x"},
		{`interface R { readonly x?: number } export function bad(value: R): void {}`, "bad.value.x"},
		{`export function bad(value: number[]): void {}`, "bad.value"},
		{`export function bad(value: 'only'): void {}`, "bad.value"},
		{`export function bad(value?: string): void {}`, "bad"},
		{`export function bad<T>(value: number): void {}`, "bad"},
		{`export function bad(value: ReadonlyMap<string,number>): void {}`, "bad.value"},
		{`export function bad(value: Readonly<{}>): void {}`, "bad.value"},
	}
	for _, testCase := range tests {
		path := requestSource(test, testCase.source)
		program, code := check([]string{path})
		if code != 0 {
			test.Fatalf("source check: %d", code)
		}
		_, exportSignaturesError := exportSignatures(program, []string{"bad"})
		if exportSignaturesError == nil || !strings.Contains(exportSignaturesError.Error(), testCase.path) {
			test.Fatalf("named unsupported signature: %v, want path %s", exportSignaturesError, testCase.path)
		}
		selected, exportSignaturesError := exportSignatures(program, nil)
		if exportSignaturesError != nil || len(selected) != 0 {
			test.Fatalf("automatic selection: %v %v", selected, exportSignaturesError)
		}
	}
}

// Not parallel: calls the process-wide compiler driver.
func TestWASIExports(test *testing.T) {
	if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		test.Skip("set ADAMIC_ORACLE_WASI=1")
	}
	directory := test.TempDir()
	binary := filepath.Join(directory, "exports.wasm")
	table := filepath.Join(directory, "exports.json")
	source, absolutePathError := filepath.Abs("../../internal/native/wasm/exports/types.a")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	if code := run(
		[]string{"build", "--target", "wasm32-wasi", "--reactor", source, "-o", binary, "--count", "--abi-json", table},
	); code != 0 {
		test.Fatalf("build: %d", code)
	}
	bytes, readError := os.ReadFile(table)
	if readError != nil {
		test.Fatal(readError)
	}
	var signatures struct {
		Version int
		Exports []struct{ Name string }
	}
	if decodeError := json.Unmarshal(bytes, &signatures); decodeError != nil {
		test.Fatal(decodeError)
	}
	if signatures.Version != 1 || len(signatures.Exports) != 14 {
		test.Fatalf("table: %s", bytes)
	}
	for _, export := range signatures.Exports {
		if export.Name == "unsupported" {
			test.Fatal("unsupported automatic export")
		}
	}
	timeoutContext, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	command := exec.CommandContext(
		timeoutContext,
		"node",
		"--disable-warning=ExperimentalWarning",
		"../../internal/native/wasm/exports/oracle.mjs",
		binary,
		source,
		table,
	)
	result, combinedOutputError := command.CombinedOutput()
	if combinedOutputError != nil {
		test.Fatalf("oracle: %v\n%s", combinedOutputError, result)
	}
	test.Log(string(result))
	if code := run([]string{
		"build",
		"--target",
		"wasm32-wasi",
		source,
		"-o",
		filepath.Join(directory, "bad.wasm"),
		"--reactor",
		"--export",
		"unsupported",
	}); code != 1 {
		test.Fatalf("unsupported build: %d", code)
	}
	if code := run([]string{
		"build",
		"--target",
		"wasm32-wasi",
		source,
		"-o",
		filepath.Join(directory, "one.wasm"),
		"--reactor",
		"--export",
		"numberValue",
		"--abi-json",
		table,
	}); code != 0 {
		test.Fatalf("narrowed build: %d", code)
	}
	bytes, readError = os.ReadFile(table)
	if readError != nil {
		test.Fatal(readError)
	}
	if decodeError := json.Unmarshal(bytes, &signatures); decodeError != nil {
		test.Fatal(decodeError)
	}
	if len(signatures.Exports) != 1 || signatures.Exports[0].Name != "numberValue" {
		test.Fatalf("narrowed table: %s", bytes)
	}
}
