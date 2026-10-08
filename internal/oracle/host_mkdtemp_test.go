package oracle

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

const hostMkdtempFixture = "internal/oracle/testdata/host_mkdtemp.a"
const hostMkdtempUnusedFixture = "internal/oracle/testdata/host_mkdtemp_unused.a"

func init() {
	for _, path := range []string{hostMkdtempFixture, hostMkdtempUnusedFixture} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{path, true, false})
	}
}

func TestHostMkdtempWASIRefusal(t *testing.T) {
	for _, fixture := range []string{hostMkdtempFixture, "internal/oracle/testdata/closure_convention_host24.a"} {
		t.Run(filepath.Base(fixture), func(t *testing.T) {
			path, _ := filepath.Abs(filepath.Join(repository, fixture))
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			binary := filepath.Join(t.TempDir(), "program.wasm")
			err = native.Build(native.C(program), binary, native.Options{Target: "wasm32-wasi"})
			var refused *native.WASIMkdtempRefusal
			if !errors.As(err, &refused) || !strings.Contains(err.Error(), "WASI preview1 refuses node:fs.mkdtempSync") || !strings.Contains(err.Error(), "0700") || strings.Contains(err.Error(), "clang") {
				t.Fatalf("expected the private-directory WASI refusal before clang: %v", err)
			}
			if _, statError := os.Stat(binary); !os.IsNotExist(statError) {
				t.Fatalf("refused program published a binary: %v", statError)
			}
			t.Log(err)
		})
	}
}

func TestHostMkdtempUnusedWASI(t *testing.T) {
	path, _ := filepath.Abs(filepath.Join(repository, hostMkdtempUnusedFixture))
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	expected := onNode(t, path)
	binary := filepath.Join(t.TempDir(), "program.wasm")
	if err := native.Build(native.C(program), binary, native.Options{Target: "wasm32-wasi"}); err != nil {
		t.Fatal(err)
	}
	actual := execute(t, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle", "wasi.mjs"), binary)
	if difference := disagreement(expected, actual); difference != "" {
		t.Fatal(difference)
	}
}

// Use a seven-character template while keeping libc's atomic creation, cleanup,
// sanitizer and leak checks intact. Only Node's six-character suffix catches it.
func TestHostMkdtempSuffixNodeMutant(t *testing.T) {
	path, _ := filepath.Abs(filepath.Join(repository, hostMkdtempFixture))
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	expected := onNode(t, path)
	runtime, err := os.ReadFile(filepath.Join(repository, "internal/native/runtime/node_fs_file.c"))
	if err != nil {
		t.Fatal(err)
	}
	source := string(runtime)
	if strings.Count(source, "malloc(length + 7)") != 1 || strings.Count(source, `"XXXXXX", 7`) != 2 {
		t.Fatal("mkdtemp template mutation sites moved")
	}
	changed := strings.Replace(source, "malloc(length + 7)", "malloc(length + 8)", 1)
	changed = strings.ReplaceAll(changed, `"XXXXXX", 7`, `"XXXXXXX", 8`)
	changed = strings.ReplaceAll(changed, "adamic_fs_file_", "adamic_fs_file_mutant_")
	mutant := changed + "\n" + strings.ReplaceAll(native.C(program), "adamic_fs_file_", "adamic_fs_file_mutant_")
	binary := filepath.Join(t.TempDir(), "mutant")
	if err := native.Build(mutant, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	actual := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1:halt_on_error=1"}, binary)
	if actual.exitCode != 0 || len(actual.stderr) != 0 {
		t.Fatalf("mutant failed outside Node comparison: %+v", actual)
	}
	if bytes.Equal(expected.stdout, actual.stdout) {
		t.Fatal("Node did not catch the seven-character suffix")
	}
	t.Log("Node alone caught the seven-character suffix; exit 0, no sanitizer findings or leaks")
}

func TestHostMkdtempCounts(t *testing.T) {
	for _, path := range []string{hostMkdtempFixture, hostMkdtempUnusedFixture} {
		row := counted(t, path, false, nil, false, false)
		t.Log(row)
		data, err := os.ReadFile(countsPath)
		if err != nil {
			t.Fatal(err)
		}
		text := string(data)
		if strings.Contains(text, row+"\n") {
			continue
		}
		if !*updateCounts {
			t.Fatalf("mkdtemp fixture counts missing or changed: %s", row)
		}
		if strings.Contains(text, "| "+path+" |") {
			t.Fatal("existing fixture counts changed; review explicitly")
		}
		at := -1
		for index, fixture := range fixtures {
			if fixture.path != path {
				continue
			}
			for previous := index - 1; previous >= 0; previous-- {
				boundary := "| " + fixtures[previous].path + " |"
				if found := strings.Index(text, boundary); found >= 0 {
					at = found + strings.Index(text[found:], "\n") + 1
					break
				}
			}
			break
		}
		if at < 0 {
			t.Fatal("preceding fixture count row missing")
		}
		if err := os.WriteFile(countsPath, []byte(text[:at]+row+"\n"+text[at:]), 0644); err != nil {
			t.Fatal(err)
		}
	}
}

// This mutation keeps refusal and its type intact but removes the target name
// from the real diagnostic. Only the target refusal check can reject it.
func TestHostMkdtempRefusalDiagnosticMutant(t *testing.T) {
	original, _ := filepath.Abs(filepath.Join(repository, "internal/native/library_node_mkdtemp_wasi.go"))
	source, err := os.ReadFile(original)
	if err != nil {
		t.Fatal(err)
	}
	old := "WASI preview1 refuses node:fs.mkdtempSync"
	if bytes.Count(source, []byte(old)) != 1 {
		t.Fatal("refusal diagnostic mutation site moved")
	}
	directory := t.TempDir()
	replacement := filepath.Join(directory, "library_node_mkdtemp_wasi.go")
	if err := os.WriteFile(replacement, bytes.Replace(source, []byte(old), []byte("unsupported target refuses node:fs.mkdtempSync"), 1), 0600); err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(map[string]map[string]string{"Replace": {original: replacement}})
	if err != nil {
		t.Fatal(err)
	}
	overlay := filepath.Join(directory, "overlay.json")
	if err := os.WriteFile(overlay, data, 0600); err != nil {
		t.Fatal(err)
	}
	command := exec.Command("go", "test", "-overlay", overlay, "./internal/oracle", "-run", "^TestHostMkdtempWASIRefusal$", "-count=1", "-v")
	command.Dir = repository
	output, err := command.CombinedOutput()
	if err == nil || bytes.Contains(output, []byte("build failed")) || bytes.Contains(output, []byte("native: clang failed")) || !bytes.Contains(output, []byte("expected the private-directory WASI refusal before clang")) {
		t.Fatalf("diagnostic mutant did not reach the refusal check: %v\n%s", err, output)
	}
	t.Log("diagnostic mutant caught by the target refusal check; no clang or linker failure")
}
