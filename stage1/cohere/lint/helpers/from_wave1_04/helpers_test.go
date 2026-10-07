package wave104

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type Case struct {
	Rule, File, Source, Config, Cwd, Origin string
	Present                                 bool
}

func repository(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs("../../../../..")
	if err != nil {
		t.Fatal(err)
	}
	return root
}
func run(t *testing.T, dir string, args ...string) []byte {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	command := exec.CommandContext(ctx, args[0], args[1:]...)
	command.Dir = dir
	output, err := os.CreateTemp(t.TempDir(), "output-")
	if err != nil {
		t.Fatal(err)
	}
	command.Stdout = output
	var stderr bytes.Buffer
	command.Stderr = &stderr
	err = command.Run()
	if closeErr := output.Close(); closeErr != nil {
		t.Fatal(closeErr)
	}
	if err != nil || stderr.Len() != 0 {
		t.Fatalf("%s: %v stderr=%s", args[0], err, stderr.String())
	}
	data, err := os.ReadFile(output.Name())
	if err != nil {
		t.Fatal(err)
	}
	return data
}
func goOracle(t *testing.T) string {
	t.Helper()
	root := filepath.Join(repository(t), "cohere")
	local, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	virtual := filepath.Join(root, "adamic_wave104_root.go")
	replacements := map[string]string{virtual: filepath.Join(local, "oracle.go.txt"), filepath.Join(root, "internal/lint/rules/tailwind/adamic_wave104_exports.go"): filepath.Join(local, "exports.go.txt")}
	encoded, err := json.Marshal(map[string]any{"Replace": replacements})
	if err != nil {
		t.Fatal(err)
	}
	overlay := filepath.Join(t.TempDir(), "overlay.json")
	if err = os.WriteFile(overlay, encoded, 0644); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "oracle")
	run(t, root, "go", "build", "-overlay="+overlay, "-o", binary, virtual)
	return binary
}
func build(t *testing.T, from, to string) (string, string, string) {
	return buildHelper(t, "project_root.a", "main.a", from, to)
}
func buildHelper(t *testing.T, helper, main, from, to string) (string, string, string) {
	t.Helper()
	folder := t.TempDir()
	changed := 0
	for _, file := range []string{helper, main} {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		text := string(data)
		if file == helper && from != "" {
			if strings.Count(text, from) != 1 {
				t.Fatal("mutant anchor changed")
			}
			text = strings.Replace(text, from, to, 1)
			changed++
		}
		if file == main {
			options := filepath.Join(repository(t), "stage1/cohere/lint/helpers/options_json.ts")
			nodes := filepath.Join(repository(t), "stage1/typescript/parser/nodes.ts")
			text = strings.Replace(text, "../options_json.ts", filepath.ToSlash(options), 1)
			text = strings.Replace(text, "../../../../typescript/parser/nodes.ts", filepath.ToSlash(nodes), 1)
			if main == "suffix_main.a" || main == "vector_main.a" {
				scanner := filepath.Join(repository(t), "stage1/cohere/lint/helpers/from_wave1_04/scan_number.a")
				text = strings.Replace(text, "./scan_number.a", filepath.ToSlash(scanner), 1)
			}
		}
		if err = os.WriteFile(filepath.Join(folder, file), []byte(text), 0644); err != nil {
			t.Fatal(err)
		}
	}
	if from != "" && changed != 1 {
		t.Fatal("mutation was not applied")
	}
	entry := filepath.Join(folder, main)
	program, err := load.Load([]string{entry})
	if err != nil {
		t.Fatal(err)
	}
	ir, err := lower.Lower(context.Background(), program)
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(folder, "helper")
	if err = native.Build(native.C(ir), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	js := filepath.Join(folder, "helper.mjs")
	if err = os.WriteFile(js, []byte(javascript.JavaScript(ir)), 0644); err != nil {
		t.Fatal(err)
	}
	return entry, js, binary
}
func cases(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile("cases.json")
	if err != nil {
		t.Fatal(err)
	}
	var rows []Case
	if err = json.Unmarshal(data, &rows); err != nil {
		t.Fatal(err)
	}
	names := map[string]int{}
	for _, row := range rows {
		names[row.Rule]++
	}
	if len(names) != 6 {
		t.Fatalf("expected six consumers: %v", names)
	}
	for name, count := range names {
		t.Logf("%s: %d captured inputs", name, count)
	}
	paths := []string{"", "tsconfig.json", "/", "/tsconfig.json", "//server//a/../config.json", "a/../b/tsconfig.json", "../a/../../tsconfig.json", "../", "../../", "a//b//", "a/./b/../tsconfig.json", "/../../b/config", "C:\\x\\tsconfig.json", "C:/x/../config", "😀/été/../config.json", "\x00/a/config", "line\n/a/file", "a/..", "a/../", "a/../../", "///...//file", ".", "..", "////"}
	// A bounded compositional path population distinguishes normalization from fallback identity.
	for _, a := range []string{"", "/", "./", "../", "//"} {
		for _, b := range []string{"a", "..", ".", "é", "😀", "\\"} {
			for _, c := range []string{"", "/", "//", "/../", "/./"} {
				paths = append(paths, a+b+c+"config.json")
			}
		}
	}
	controls := 0
	for _, path := range paths {
		for _, cwd := range []string{"", "/different", "rel/../raw-cwd", "./", "\\x\\y", "é😀", "cwd\x00", "cwd\n"} {
			for _, present := range []bool{false, true} {
				rows = append(rows, Case{Config: path, Cwd: cwd, Present: present, Origin: "path and lazy-read control"})
				controls++
			}
		}
	}
	t.Logf("%d fixture inputs plus %d controls = %d cases", len(rows)-controls, controls, len(rows))
	data, err = json.Marshal(rows)
	if err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(t.TempDir(), "cases.json")
	if err = os.WriteFile(file, data, 0644); err != nil {
		t.Fatal(err)
	}
	return file
}
func observations(t *testing.T, entry, js, binary, path string) map[string][]byte {
	t.Helper()
	runner := filepath.Join(repository(t), "oracle/node.mjs")
	return map[string][]byte{"source Node": run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, entry, path), "emitted JavaScript": run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, js, path), "ASan/UBSan native": run(t, "", binary, path)}
}
func firstDifference(got, want []byte) string {
	a, b := strings.Split(string(got), "\n"), strings.Split(string(want), "\n")
	for i := 0; i < len(a) && i < len(b); i++ {
		if a[i] != b[i] {
			return fmt.Sprintf("line %d got=%q Go=%q", i+1, a[i], b[i])
		}
	}
	return fmt.Sprintf("bytes got=%d Go=%d", len(got), len(want))
}

// Not parallel: bound compiler and sanitizer memory while comparing the same source artifact.
func TestProjectRootMatchesGo(t *testing.T) {
	path := cases(t)
	want := run(t, "", goOracle(t), path)
	entry, js, binary := build(t, "", "")
	for side, got := range observations(t, entry, js, binary, path) {
		if !bytes.Equal(got, want) {
			t.Fatalf("%s %s", side, firstDifference(got, want))
		}
	}
	t.Logf("%d bytes, %d lines identical to actual Go on Node, emitted JS and sanitized native", len(want), bytes.Count(want, []byte("\n")))
}

// Not parallel: each compiling semantic mutant owns an independent native build.
func TestProjectRootMutants(t *testing.T) {
	path := cases(t)
	want := run(t, "", goOracle(t), path)
	for _, change := range []struct{ name, from, to string }{{"accept empty configured path", "options.present && options.configFilePath !== ''", "options.present"}, {"omit parent path cleaning", "if(part === '..')", "if(false)"}, {"clean the fallback directory", "return program.currentDirectory();", "return directory(program.currentDirectory());"}} {
		t.Run(change.name, func(t *testing.T) {
			entry, js, binary := build(t, change.from, change.to)
			for side, got := range observations(t, entry, js, binary, path) {
				if bytes.Equal(got, want) {
					t.Fatal("mutant survived", side)
				}
				t.Logf("%s compiled, finished cleanly and was caught only by Go comparison on %s: %s", change.name, side, firstDifference(got, want))
			}
		})
	}
}
