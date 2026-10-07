package assertions

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

func command(t *testing.T, dir, name string, args ...string) []byte {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	log, err := os.CreateTemp(t.TempDir(), "stdout-")
	if err != nil {
		t.Fatal(err)
	}
	defer log.Close()
	cmd.Stdout = log
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err = cmd.Run(); err != nil {
		t.Fatalf("%s %v: %v\n%s", name, args, err, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("%s stderr: %s", name, stderr.String())
	}
	data, err := os.ReadFile(log.Name())
	if err != nil {
		t.Fatal(err)
	}
	return data
}
func root(t *testing.T) string {
	t.Helper()
	p, err := filepath.Abs("../../../../..")
	if err != nil {
		t.Fatal(err)
	}
	return p
}
func oracle(t *testing.T) string {
	t.Helper()
	cohere := filepath.Join(root(t), "cohere")
	source, err := filepath.Abs("testdata/driver.go")
	if err != nil {
		t.Fatal(err)
	}
	virtual := filepath.Join(cohere, "adamic_wave1_next_oracle.go")
	data, err := json.Marshal(map[string]any{"Replace": map[string]string{virtual: source}})
	if err != nil {
		t.Fatal(err)
	}
	overlay := filepath.Join(t.TempDir(), "overlay.json")
	if err = os.WriteFile(overlay, data, 0644); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "oracle")
	command(t, cohere, "go", "build", "-overlay="+overlay, "-o", binary, virtual)
	return binary
}
func file(t *testing.T, name string, data []byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
	return path
}
func compile(t *testing.T, entry string) (string, string) {
	t.Helper()
	program, err := load.Load([]string{entry})
	if err != nil {
		t.Fatal(err)
	}
	ir, err := lower.Lower(context.Background(), program)
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "port")
	if err = native.Build(native.C(ir), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	emitted := file(t, "main.mjs", []byte(javascript.JavaScript(ir)))
	return binary, emitted
}
func equal(t *testing.T, got, want []byte) {
	t.Helper()
	if bytes.Equal(got, want) {
		return
	}
	a, b := strings.Split(string(got), "\n"), strings.Split(string(want), "\n")
	for i := 0; i < len(a) && i < len(b); i++ {
		if a[i] != b[i] {
			t.Fatalf("line %d got %q Go %q", i+1, a[i], b[i])
		}
	}
	t.Fatalf("output size got %d Go %d", len(got), len(want))
}
func runSides(t *testing.T, entry, binary, emitted, corpus string, args ...string) [][]byte {
	t.Helper()
	runner := filepath.Join(root(t), "oracle/node.mjs")
	sourceArgs := append([]string{"--disable-warning=ExperimentalWarning", runner, entry, corpus}, args...)
	emittedArgs := append([]string{"--disable-warning=ExperimentalWarning", runner, emitted, corpus}, args...)
	nativeArgs := append([]string{corpus}, args...)
	return [][]byte{command(t, "", "node", sourceArgs...), command(t, "", "node", emittedArgs...), command(t, "", binary, nativeArgs...)}
}

var imports = regexp.MustCompile(`from '([^']+)'`)

func absoluteImports(t *testing.T, path string, data []byte) []byte {
	t.Helper()
	return imports.ReplaceAllFunc(data, func(part []byte) []byte {
		match := imports.FindSubmatch(part)
		if !strings.HasPrefix(string(match[1]), ".") {
			return part
		}
		return []byte("from '" + filepath.ToSlash(filepath.Join(filepath.Dir(path), string(match[1]))) + "'")
	})
}

// Not parallel: bound native compilation and corpus memory.
func TestAssertions(t *testing.T) {
	oracle := oracle(t)
	cases, _ := filepath.Abs("testdata/cases.json")
	want := command(t, "", oracle, cases)
	adapted := file(t, "ast.json", command(t, "", oracle, "--ast", cases))
	entry, _ := filepath.Abs("main.a")
	binary, emitted := compile(t, entry)
	for side, got := range runSides(t, entry, binary, emitted, adapted) {
		equal(t, got, want)
		t.Logf("full Go AST contract side %d: %d identical bytes", side, len(got))
	}
	data, err := os.ReadFile(cases)
	if err != nil {
		t.Fatal(err)
	}
	var rows []map[string]any
	if err = json.Unmarshal(data, &rows); err != nil {
		t.Fatal(err)
	}
	var parsed []map[string]any
	var jsx []map[string]any
	for _, row := range rows {
		if strings.HasSuffix(row["file"].(string), ".tsx") {
			jsx = append(jsx, row)
		} else {
			parsed = append(parsed, row)
		}
	}
	encoded, _ := json.Marshal(parsed)
	raw := file(t, "source.json", encoded)
	sourceWant := command(t, "", oracle, raw)
	sourceAST := file(t, "source-ast.json", command(t, "", oracle, "--ast", raw))
	for side, got := range runSides(t, entry, binary, emitted, sourceAST, "--source") {
		equal(t, got, sourceWant)
		t.Logf("independent source side %d: %d cases identical", side, len(parsed))
	}
	t.Logf("fixtures total %d, independent %d, JSX gap %d", len(rows), len(parsed), len(jsx))
	original, _ := filepath.Abs("precedence.a")
	source, _ := os.ReadFile(original)
	mutated := file(t, "precedence.a", absoluteImports(t, original, []byte(strings.Replace(string(source), "inner > outer", "inner >= outer", 1))))
	module, _ := filepath.Abs("rule.a")
	data, _ = os.ReadFile(module)
	mutantModule := file(t, "rule.a", []byte(strings.Replace(string(absoluteImports(t, module, data)), filepath.ToSlash(original), filepath.ToSlash(mutated), 1)))
	data, _ = os.ReadFile(entry)
	mutantEntry := file(t, "main.a", []byte(strings.Replace(string(absoluteImports(t, entry, data)), filepath.ToSlash(module), filepath.ToSlash(mutantModule), 1)))
	mutantNative, mutantJS := compile(t, mutantEntry)
	for side, got := range runSides(t, mutantEntry, mutantNative, mutantJS, adapted) {
		if bytes.Equal(got, want) {
			t.Fatal("mutant survived")
		}
		t.Logf("equal-precedence mutant compiles and runs; comparison alone caught side %d", side)
	}
	if len(jsx) == 0 {
		t.Fatal("missing JSX control")
	}
	encoded, _ = json.Marshal(jsx[:1])
	gap := file(t, "gap.json", command(t, "", oracle, "--ast", file(t, "raw-gap.json", encoded)))
	runner := filepath.Join(root(t), "oracle/node.mjs")
	for _, cmd := range []*exec.Cmd{exec.Command("node", "--disable-warning=ExperimentalWarning", runner, entry, gap, "--source"), exec.Command("node", "--disable-warning=ExperimentalWarning", runner, emitted, gap, "--source"), exec.Command(binary, gap, "--source")} {
		log, err := os.CreateTemp(t.TempDir(), "refusal-")
		if err != nil {
			t.Fatal(err)
		}
		cmd.Stdout = log
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		err = cmd.Run()
		log.Close()
		if err == nil || !strings.Contains(stderr.String(), "NotYet: assertion JSX parser") {
			t.Fatalf("refusal: %v %s", err, stderr.String())
		}
	}
	for side, cmd := range []*exec.Cmd{exec.Command(oracle, "--count", cases), exec.Command("node", "--disable-warning=ExperimentalWarning", runner, entry, adapted, "--count"), exec.Command(binary, adapted, "--count")} {
		log, err := os.CreateTemp(t.TempDir(), "rate-")
		if err != nil {
			t.Fatal(err)
		}
		cmd.Stdout = log
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		start := time.Now()
		err = cmd.Run()
		seconds := time.Since(start).Seconds()
		log.Close()
		if err != nil || stderr.Len() != 0 {
			t.Fatalf("rate: %v %s", err, stderr.String())
		}
		output, _ := os.ReadFile(log.Name())
		t.Logf("fixture rate side %d findings=%s elapsed=%.6fs", side, strings.TrimSpace(string(output)), seconds)
	}
}

// Not parallel: compiler AST JSON is compared in bounded batches of five files.
func TestAssertionCorpus(t *testing.T) {
	var rows []map[string]any
	for _, directory := range []string{filepath.Join(root(t), "stage1"), "/tmp/lint-wave1-typescript/src/compiler"} {
		err := filepath.WalkDir(directory, func(path string, entry os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() {
				return nil
			}
			if !strings.HasSuffix(path, ".a") && !strings.HasSuffix(path, ".ts") && !strings.HasSuffix(path, ".tsx") {
				return nil
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			rows = append(rows, map[string]any{"file": path, "source": string(data), "rule": "@typescript-eslint/consistent-type-assertions", "options": nil})
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	oracle := oracle(t)
	entry, _ := filepath.Abs("main.a")
	binary, emitted := compile(t, entry)
	total := 0
	for begin := 0; begin < len(rows); begin += 5 {
		end := begin + 5
		if end > len(rows) {
			end = len(rows)
		}
		encoded, _ := json.Marshal(rows[begin:end])
		raw := file(t, "batch.json", encoded)
		want := command(t, "", oracle, raw)
		adapted := file(t, "ast.json", command(t, "", oracle, "--ast", raw))
		for _, got := range runSides(t, entry, binary, emitted, adapted) {
			equal(t, got, want)
		}
		total += len(want)
	}
	t.Logf("compiler/stage1 %d files, %d bytes per backend identical; projected AST certification", len(rows), total)
}
