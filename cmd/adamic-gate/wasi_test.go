package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWASISourceGates(t *testing.T) {
	root := t.TempDir()
	os.WriteFile(filepath.Join(root, "go.mod"), []byte("module gateprobe\n"), 0600)
	check := func(source string, want bool) {
		t.Helper()
		os.WriteFile(filepath.Join(root, "probe_test.go"), []byte("package gateprobe\nimport (\"os\";\"testing\")\n"+source), 0600)
		gates, err := wasiGates(root)
		if want {
			if err != nil || !gates["gateprobe::TestProbe"] {
				t.Fatalf("gate missing: %v %v", gates, err)
			}
		} else if err == nil {
			t.Fatalf("unrecognized gate accepted: %v", gates)
		}
	}
	check(`func TestProbe(t *testing.T) { if os.Getenv("ADAMIC_ORACLE_WASI") != "1" { t.Skip("opt in") } }`, true)
	check(`func TestProbe(t *testing.T) { root := os.Getenv("WASI_SYSROOT"); if root == "" { t.Skip("missing") } }`, true)
	check(`func TestProbe(t *testing.T) { root := os.Getenv("WASI_SYSROOT"); if root == "" { t.Skip("missing") }; if err := probe(); err != nil { t.Skip("tool missing") }; if err := compile(root); err != nil { t.Fatal(err) } }`, true)
	check(`func TestProbe(t *testing.T) { root := os.Getenv("WASI_SYSROOT"); if root == "" { t.Skip("missing") }; flags := append([]string{}, root); if len(flags) == 0 { t.Fatal("empty") } }`, true)
	check(`func TestProbe(t *testing.T) { disabled := strings.Contains(os.Getenv("NEW_WASI"), "no"); if disabled { t.Skip("disabled") } }`, false)
	check(`func TestProbe(t *testing.T) { if strings.Contains(os.Getenv("NEW_WASI"), "yes") { t.Skip("missing") } }`, false)
	check(`func TestProbe(t *testing.T) { if _, ok := os.LookupEnv("NEW_WASI"); !ok { t.Skip("missing") } }`, false)
	check(`func helper(t *testing.T) { if os.Getenv("NEW_WASI") == "" { t.Skip("missing") } }`, false)
	check(`func TestProbe(t *testing.T) { if os.Getenv(name) == "" { t.Skip("missing") } }`, false)
	check(`func TestProbe(t *testing.T) { value, _ := syscall.Getenv("NEW_WASI"); if value == "" { t.Skip("missing") } }`, false)
	check(`func TestProbe(t *testing.T) { value := customEnv("NEW_WASI"); if value == "" { t.Skip("missing") } }`, false)
	check(`const key = "NEW_WASI"; func TestProbe(t *testing.T) { value, _ := syscall.Getenv(key); if value == "" { t.Skip("missing") } }`, false)
}

func TestRequiredWASISkips(t *testing.T) {
	p := plan{Units: []unit{{Package: "gateprobe", Test: "TestProbe/a", WASI: true}, {Package: "gateprobe", Test: "TestProbe/b", WASI: true}, {Package: "gateprobe", Test: "TestOptional"}}}
	if errors := wasiSkips(p, 7, []result{{Package: "gateprobe", Test: "TestOptional", Action: "skip"}, {Package: "gateprobe", Test: "TestProbe/a", Action: "pass"}}); len(errors) != 0 {
		t.Fatal(errors)
	}
	for _, name := range []string{"TestProbe", "TestProbe/a", "TestProbe/a/nested"} {
		errors := wasiSkips(p, 7, []result{{Package: "gateprobe", Test: name, Action: "skip", Reason: "fixture does not lower"}})
		if len(errors) == 0 || !strings.Contains(errors[0], "shard 7 required WASI unit gateprobe::TestProbe/a skipped") || !strings.Contains(errors[0], "fixture does not lower") {
			t.Fatalf("mandatory skip accepted: %s %v", name, errors)
		}
	}
}

func TestWASIPreflight(t *testing.T) {
	t.Setenv("ADAMIC_TEST_WASI", "")
	t.Setenv("ADAMIC_ORACLE_WASI", "")
	t.Setenv("WASI_SYSROOT", "")
	if wasiReady() == nil {
		t.Fatal("missing variables accepted")
	}
	t.Setenv("ADAMIC_TEST_WASI", "1")
	t.Setenv("ADAMIC_ORACLE_WASI", "1")
	sdk := t.TempDir()
	root := filepath.Join(sdk, "share", "wasi-sysroot")
	t.Setenv("WASI_SYSROOT", root)
	if wasiReady() == nil {
		t.Fatal("missing SDK accepted")
	}
	for _, name := range []string{"share/wasi-sysroot/include/stdlib.h", "share/wasi-sysroot/lib/wasm32-wasi/libc.a", "bin/clang", "bin/wasm-ld"} {
		path := filepath.Join(sdk, name)
		os.MkdirAll(filepath.Dir(path), 0700)
		os.WriteFile(path, []byte("probe"), 0700)
	}
	if err := wasiReady(); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"ADAMIC_TEST_WASI", "ADAMIC_ORACLE_WASI", "WASI_SYSROOT"} {
		saved := os.Getenv(name)
		os.Setenv(name, "")
		if wasiReady() == nil {
			t.Fatalf("missing %s accepted", name)
		}
		os.Setenv(name, saved)
	}
	for _, name := range []string{"share/wasi-sysroot/include/stdlib.h", "share/wasi-sysroot/lib/wasm32-wasi/libc.a", "bin/clang", "bin/wasm-ld"} {
		path := filepath.Join(sdk, name)
		os.Remove(path)
		if wasiReady() == nil {
			t.Fatalf("missing SDK component %s accepted", name)
		}
		os.WriteFile(path, []byte("probe"), 0700)
	}
	clang := filepath.Join(sdk, "bin", "clang")
	os.Chmod(clang, 0600)
	if wasiReady() == nil {
		t.Fatal("nonexecutable SDK compiler accepted")
	}
	os.Chmod(clang, 0700)
	legacy := filepath.Join(root, "include", "stdlib.h")
	multiarch := filepath.Join(root, "include", "wasm32-wasi", "stdlib.h")
	os.MkdirAll(filepath.Dir(multiarch), 0700)
	os.Rename(legacy, multiarch)
	if err := wasiReady(); err != nil {
		t.Fatalf("SDK 27 multiarch headers refused: %v", err)
	}
}

func TestWASICompilerSelectors(t *testing.T) {
	selectors := patterns([]unit{{Package: "gateprobe/internal/native", Test: "TestWASI", WASI: true}, {Package: "gateprobe/internal/native", Test: "TestNative"}})
	if len(selectors) != 2 || selectors[1] != "^TestWASI$" {
		t.Fatalf("native and SDK compilers mixed: %v", selectors)
	}
}

func TestWASISDKResumeIdentity(t *testing.T) {
	t.Setenv("ADAMIC_TOOLS", "")
	sdk := t.TempDir()
	root := filepath.Join(sdk, "share", "wasi-sysroot")
	os.MkdirAll(root, 0700)
	t.Setenv("WASI_SYSROOT", root)
	identity := func() string {
		t.Helper()
		value, err := executionIdentity()
		if err != nil {
			t.Fatal(err)
		}
		return value
	}
	before := identity()
	os.WriteFile(filepath.Join(root, "header.h"), []byte("changed header"), 0600)
	if identity() == before {
		t.Fatal("WASI sysroot bytes omitted from resume identity")
	}
	before = identity()
	os.MkdirAll(filepath.Join(sdk, "bin"), 0700)
	os.WriteFile(filepath.Join(sdk, "bin", "wasm-ld"), []byte("changed linker"), 0700)
	if identity() == before {
		t.Fatal("WASI SDK linker omitted from resume identity")
	}
}

func TestWASICurrentSourceAudit(t *testing.T) {
	gates, err := wasiGates("../..")
	if err != nil {
		t.Fatal(err)
	}
	_ = gates
}
