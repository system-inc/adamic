package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A fix that rewrites a file while the early type check is still running never crashes the run (#sm79kfv).
//
// The early check starts beside the fix phase's walk. When the fixer rewrites a file the graph is rebuilt,
// and that check used to be cancelled. The compiler's whole-program check cannot be cancelled part way: a
// checker that sees its context end marks itself cancelled, the compiler hands it the next file of its
// group anyway, and GetGlobalDiagnostics panics on a compiler goroutine nothing can recover. A bare
// `cohere`, the gate itself, died that way in 5 of 6 runs on a project whose check outlasted the fix walk.
//
// Here the race is not left to luck. COHERE_TEST_HOLD_EARLY_CHECK holds the early check until the fix phase
// and its rebuild are over, so it always runs after the point where the cancel used to happen, and
// `--single-threaded` puts every file in one checker's group, so a cancelled checker always has a next file
// to panic on. Cancelling the early check at the rebuild again fails this every time.
func TestAFixThatRewritesAFileDuringTheEarlyCheckDoesNotCrashTheRun(t *testing.T) {
	t.Parallel()
	binary := buildCohere(t)
	root := t.TempDir()
	write := func(name string, contents string) {
		t.Helper()
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("tsconfig.json", `{"compilerOptions":{"strict":true,"noEmit":true,"target":"es2022","module":"esnext","moduleResolution":"bundler"},"include":["source"]}`)
	write("CohereSettings.json", `{"rules":{"prefer-const":"error"}}`)
	write("package.json", `{"name":"fixture","private":true,"type":"module"}`)
	write(".gitignore", ".cache/\nnode_modules/\n")
	write("source/fixable.ts", "export function fixable(): number {\n    let unchanged = 1;\n    return unchanged;\n}\n")
	for index := range 4 {
		write(fmt.Sprintf("source/other%d.ts", index), fmt.Sprintf("export const other%d: number = %d;\n", index, index))
	}

	output, exitCode := runCohereWithEnvironment(t, binary, root, []string{"COHERE_TEST_HOLD_EARLY_CHECK=1"}, "--single-threaded")

	if strings.Contains(output, "panic") {
		t.Fatalf("the run panicked after its fix phase rebuilt the graph:\n%s", output)
	}
	if exitCode != 0 {
		t.Fatalf("exit %d, want 0:\n%s", exitCode, output)
	}
	// Each of these proves the run reached the shape that used to crash, rather than passing by skipping it.
	if !strings.Contains(output, "graph rebuilt in") {
		t.Fatalf("the fix phase did not rebuild the graph, so the early check was never retired:\n%s", output)
	}
	if !strings.Contains(output, "types: 0 diagnostics over 5 files") {
		t.Fatalf("the types phase did not report over the rebuilt graph's five files:\n%s", output)
	}
	rewritten, err := os.ReadFile(filepath.Join(root, "source/fixable.ts"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(rewritten), "const unchanged = 1;") {
		t.Fatalf("prefer-const's fix did not reach disk, so nothing was rewritten during the check:\n%s", rewritten)
	}
}

// `--explain` explains the file as it stands and writes nothing (#sm79kfv, ruled by @system_cohere).
//
// It used to run the writing fix phase before explaining, so `cohere --explain SecretRow.tsx` in ahra
// rewrote two other files and explained a tree that was no longer the one asked about. Now it is a
// `--no-fix` run with an explanation after it: the fixable file's bytes stay put, the fix phase says what it
// would rewrite, and naming `--fix` beside it is refused.
func TestExplainWritesNothing(t *testing.T) {
	t.Parallel()
	binary := buildCohere(t)
	root := t.TempDir()
	write := func(name string, contents string) {
		t.Helper()
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("tsconfig.json", `{"compilerOptions":{"strict":true,"noEmit":true,"target":"es2022","module":"esnext","moduleResolution":"bundler"},"include":["source"]}`)
	write("CohereSettings.json", `{"rules":{"prefer-const":"error"}}`)
	write("package.json", `{"name":"fixture","private":true,"type":"module"}`)
	write(".gitignore", ".cache/\nnode_modules/\n")
	const fixable = "export function fixable(): number {\n    let unchanged = 1;\n    return unchanged;\n}\n"
	write("source/fixable.ts", fixable)

	output, _ := runCohere(t, binary, root, "--explain", "source/fixable.ts")
	onDisk, err := os.ReadFile(filepath.Join(root, "source/fixable.ts"))
	if err != nil {
		t.Fatal(err)
	}
	if string(onDisk) != fixable {
		t.Fatalf("--explain rewrote the file:\n%s\noutput:\n%s", onDisk, output)
	}
	if !strings.Contains(output, "would be rewritten") || strings.Contains(output, "graph rebuilt in") {
		t.Fatalf("--explain's fix phase should say what it would rewrite and rebuild nothing:\n%s", output)
	}
	if !strings.Contains(output, "explain: ") {
		t.Fatalf("no explanation was printed, so the run never reached what it was asked for:\n%s", output)
	}

	refused, exitCode := runCohere(t, binary, root, "--explain", "source/fixable.ts", "--fix")
	if exitCode == 0 || !strings.Contains(refused, "--explain and --fix contradict each other") {
		t.Fatalf("--explain with --fix was not refused (exit %d):\n%s", exitCode, refused)
	}
}
