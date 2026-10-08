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
func TestWASIExportSelection(t *testing.T) {
	tests := []struct{ source, path string }{
		{`export function bad(value: readonly boolean[]): void {}`, "bad.value"},
		{`interface R { readonly outer: { readonly middle: { readonly x: number } } } export function bad(value: R): void {}`, "bad.value.outer.middle"},
		{`interface R { readonly x: number | string } export function bad(value: R): void {}`, "bad.value.x"},
		{`interface R { x: number } export function bad(value: R): void {}`, "bad.value.x"},
		{`interface R { readonly x?: number } export function bad(value: R): void {}`, "bad.value.x"},
		{`export function bad(value: number[]): void {}`, "bad.value"},
		{`export function bad(value: 'only'): void {}`, "bad.value"},
		{`export function bad(value?: string): void {}`, "bad"},
		{`export function bad<T>(value: number): void {}`, "bad"},
		{`export function bad(value: ReadonlyMap<string,number>): void {}`, "bad.value"},
	}
	for _, test := range tests {
		path := requestSource(t, test.source)
		program, code := check([]string{path})
		if code != 0 {
			t.Fatalf("source check: %d", code)
		}
		_, err := exportSignatures(program, []string{"bad"})
		if err == nil || !strings.Contains(err.Error(), test.path) {
			t.Fatalf("named unsupported signature: %v, want path %s", err, test.path)
		}
		selected, err := exportSignatures(program, nil)
		if err != nil || len(selected) != 0 {
			t.Fatalf("automatic selection: %v %v", selected, err)
		}
	}
}

// Not parallel: calls the process-wide compiler driver.
func TestWASIExports(t *testing.T) {
	if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		t.Skip("set ADAMIC_ORACLE_WASI=1")
	}
	directory := t.TempDir()
	binary := filepath.Join(directory, "exports.wasm")
	table := filepath.Join(directory, "exports.json")
	source, err := filepath.Abs("../../internal/native/wasm/exports/types.a")
	if err != nil {
		t.Fatal(err)
	}
	if code := run([]string{"build", "--target", "wasm32-wasi", "--reactor", source, "-o", binary, "--count", "--abi-json", table}); code != 0 {
		t.Fatalf("build: %d", code)
	}
	bytes, err := os.ReadFile(table)
	if err != nil {
		t.Fatal(err)
	}
	var signatures struct {
		Version int
		Exports []struct{ Name string }
	}
	if err = json.Unmarshal(bytes, &signatures); err != nil {
		t.Fatal(err)
	}
	if signatures.Version != 1 || len(signatures.Exports) != 14 {
		t.Fatalf("table: %s", bytes)
	}
	for _, export := range signatures.Exports {
		if export.Name == "unsupported" {
			t.Fatal("unsupported automatic export")
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	command := exec.CommandContext(ctx, "node", "--disable-warning=ExperimentalWarning", "../../internal/native/wasm/exports/oracle.mjs", binary, source, table)
	result, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("oracle: %v\n%s", err, result)
	}
	t.Log(string(result))
	if code := run([]string{"build", "--target", "wasm32-wasi", source, "-o", filepath.Join(directory, "bad.wasm"), "--reactor", "--export", "unsupported"}); code != 1 {
		t.Fatalf("unsupported build: %d", code)
	}
	if code := run([]string{"build", "--target", "wasm32-wasi", source, "-o", filepath.Join(directory, "one.wasm"), "--reactor", "--export", "numberValue", "--abi-json", table}); code != 0 {
		t.Fatalf("narrowed build: %d", code)
	}
	bytes, err = os.ReadFile(table)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(bytes, &signatures); err != nil {
		t.Fatal(err)
	}
	if len(signatures.Exports) != 1 || signatures.Exports[0].Name != "numberValue" {
		t.Fatalf("narrowed table: %s", bytes)
	}
}
