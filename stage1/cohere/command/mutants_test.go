package command

import (
	"bytes"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func mutantSource(t *testing.T, file, from, to string) string {
	t.Helper()
	root := t.TempDir()
	for _, directory := range []string{"command", "config", "formatfiles", "gitignore"} {
		target := filepath.Join(root, directory)
		if err := os.Mkdir(target, 0755); err != nil {
			t.Fatal(err)
		}
		entries, err := os.ReadDir(filepath.Join("..", directory))
		if err != nil {
			t.Fatal(err)
		}
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".ts") {
				continue
			}
			content, err := os.ReadFile(filepath.Join("..", directory, entry.Name()))
			if err != nil {
				t.Fatal(err)
			}
			text := string(content)
			if directory == "command" && entry.Name() == file {
				if strings.Count(text, from) != 1 {
					t.Fatalf("mutant must replace one occurrence %q", from)
				}
				text = strings.Replace(text, from, to, 1)
			}
			write(t, filepath.Join(target, entry.Name()), text, 0644)
		}
	}
	return filepath.Join(root, "command", "probe.ts")
}
func TestExecutableMutants(t *testing.T) {
	// A baseline disagreement must stop the mutant run, including when this test is selected alone.
	TestOriginalCommandBoundaries(t)
	for _, mutant := range []struct{ name, file, from, to string }{
		{"root marker precedence", "location.ts", "if(regularFile(absoluteFrom(directory, 'tsconfig.json')))", "if(regularFile(absoluteFrom(directory, 'tsconfig.json')) && !regularFile(absoluteFrom(directory, 'Package.swift')))"},
		{"dot from subdirectory", "scope.ts", "if(path === clean(projectRoot))", "if(path === clean(cwd))"},
		{"scope value copy", "scope.ts", "const result = this.copy();\n        result.description += `, ${count} of them in the program`;", "const result = this.copy();\n        this.description += ' leaked';\n        result.description += `, ${count} of them in the program`;"},
		{"nearest owner tie", "ownership.ts", "common > ownerShared", "common >= ownerShared"},
	} {
		t.Run(mutant.name, func(t *testing.T) {
			source := mutantSource(t, mutant.file, mutant.from, mutant.to)
			program := lowered(t, source)
			binary := filepath.Join(t.TempDir(), "mutant")
			if err := native.Build(native.C(program), binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			got := originalTests(t, source, binary)
			if got.exitCode == 0 {
				t.Fatal("mutant survived Go's original tests")
			}
			text := string(got.stdout) + string(got.stderr)
			for _, side := range []string{"ADAMIC_NATIVE_PROBE", "ADAMIC_NODE_PROBE", "ADAMIC_BACKEND_PROBE"} {
				if !strings.Contains(text, side) {
					t.Fatalf("%s not compared:\n%s", side, text)
				}
			}
			if strings.Count(text, "error <nil> stderr ") < 3 || strings.Contains(text, "AddressSanitizer") || strings.Contains(text, "runtime error:") {
				t.Fatalf("mutant must fail on output after a clean execution:\n%s", text)
			}
			t.Logf("native, Node and backend executed cleanly; Go original tests caught %s", mutant.name)
		})
	}
	t.Run("bare boolean", func(t *testing.T) {
		source := mutantSource(t, "flags.ts", "if(equals < 0) {\n                value = 'true';", "if(equals < 0) {\n                value = 'false';")
		program := lowered(t, source)
		binary := filepath.Join(t.TempDir(), "mutant")
		if err := native.Build(native.C(program), binary, native.Options{Sanitize: true}); err != nil {
			t.Fatal(err)
		}
		oracle := filepath.Join(t.TempDir(), "oracle")
		build := bounded(t, "go", "build", "-o", oracle, "testdata/definitions.go")
		if out, err := build.CombinedOutput(); err != nil {
			t.Fatalf("%v %s", err, out)
		}
		input := `{"kind":"flags","program":"cohere","arguments":["--types"]}`
		want := execute(t, nil, oracle, absolute(t, filepath.Join(repository, "cohere/command/cohere/main.go")), input)
		if want.exitCode != 0 || len(want.stderr) > 0 {
			t.Fatal("Go flag oracle failed")
		}
		baseline := absolute(t, "probe.ts")
		baselineResult := onNode(t, baseline, input)
		if baselineResult.exitCode != 0 || len(baselineResult.stderr) > 0 || !bytes.Equal(baselineResult.stdout, want.stdout) {
			t.Fatal("baseline disagrees with Go")
		}
		for _, side := range []string{binary, wrapper(t, source, false), wrapper(t, source, true)} {
			got := execute(t, nil, side, input)
			if got.exitCode != 0 || len(got.stderr) > 0 {
				t.Fatalf("mutant failed execution: %d %s", got.exitCode, got.stderr)
			}
			if bytes.Equal(got.stdout, want.stdout) {
				t.Fatal("boolean mutant survived")
			}
		}
		t.Log("native, Node and backend executed cleanly; Go flag caught bare boolean mutation")
	})
}
