package physicaldirection

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
	"strconv"
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

// Not parallel: each backend and semantic mutant runs sequentially to bound native compiler memory.
func TestRuleContracts(t *testing.T) {
	goOracle := oracle(t)
	cases, err := filepath.Abs("testdata/cases.json")
	if err != nil {
		t.Fatal(err)
	}
	want := command(t, "", goOracle, cases)
	adapted := file(t, "adapted.json", command(t, "", goOracle, "--ast", cases))
	entry, err := filepath.Abs("main.a")
	if err != nil {
		t.Fatal(err)
	}
	binary, emitted := compile(t, entry)
	for i, got := range runSides(t, entry, binary, emitted, adapted) {
		equal(t, got, want)
		t.Logf("Go AST contract side %d identical: %d bytes", i, len(got))
	}
	var rows []map[string]any
	data, err := os.ReadFile(cases)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(data, &rows); err != nil {
		t.Fatal(err)
	}
	var sourceRows []map[string]any
	var excluded []map[string]any
	for _, row := range rows {
		if strings.HasSuffix(row["file"].(string), ".tsx") && strings.Contains(row["source"].(string), "<") {
			excluded = append(excluded, row)
		} else {
			sourceRows = append(sourceRows, row)
		}
	}
	sourceData, err := json.Marshal(sourceRows)
	if err != nil {
		t.Fatal(err)
	}
	sourceCases := file(t, "source.json", sourceData)
	sourceAdapted := file(t, "source-adapted.json", command(t, "", goOracle, "--ast", sourceCases))
	sourceWant := command(t, "", goOracle, sourceCases)
	for i, got := range runSides(t, entry, binary, emitted, sourceAdapted, "--source") {
		equal(t, got, sourceWant)
		t.Logf("independently parsed source side %d identical: %d bytes", i, len(got))
	}
	t.Logf("fixture contracts: total=%d independently parsed=%d JSX unsupported=%d", len(rows), len(sourceRows), len(excluded))
	for _, slug := range []string{"structure-tailwind-no-physical-direction", "eslint-comments-require-description", "next-google-font-display"} {
		t.Run(slug, func(t *testing.T) {
			directory := filepath.Join(root(t), "stage1/cohere/lint/rules", slug)
			changeData, err := os.ReadFile(filepath.Join(directory, "mutant.json"))
			if err != nil {
				t.Fatal(err)
			}
			var mutation struct{ From, To string }
			if err = json.Unmarshal(changeData, &mutation); err != nil {
				t.Fatal(err)
			}
			module := filepath.Join(directory, "rule.a")
			source, err := os.ReadFile(module)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Count(string(source), mutation.From) != 1 {
				t.Fatal("mutant anchor count")
			}
			mutated := file(t, "rule.a", absoluteImports(t, module, []byte(strings.Replace(string(source), mutation.From, mutation.To, 1))))
			mainData, err := os.ReadFile(entry)
			if err != nil {
				t.Fatal(err)
			}
			mainData = absoluteImports(t, entry, mainData)
			mainData = []byte(strings.Replace(string(mainData), filepath.ToSlash(module), filepath.ToSlash(mutated), 1))
			mutantEntry := file(t, "main.a", mainData)
			mutantBinary, mutantJS := compile(t, mutantEntry)
			for i, got := range runSides(t, mutantEntry, mutantBinary, mutantJS, adapted) {
				if bytes.Equal(got, want) {
					t.Fatalf("semantic mutant survived side %d", i)
				}
				t.Logf("compiling/run-successful mutant caught only by output comparison on side %d", i)
			}
		})
	}
	if len(excluded) == 0 {
		t.Fatal("JSX gap control missing")
	}
	first, err := json.Marshal(excluded[:1])
	if err != nil {
		t.Fatal(err)
	}
	gap := file(t, "gap.json", command(t, "", goOracle, "--ast", file(t, "raw-gap.json", first)))
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
		if err == nil || !strings.Contains(stderr.String(), "NotYet: stage1 JSX parser adapter") {
			t.Fatalf("gap did not refuse: %v %s", err, stderr.String())
		}
	}
}

// Not parallel: bounded batches keep sanitizer and JSON memory below the worker limit.
func TestSourceCorpora(t *testing.T) {
	var rows []map[string]any
	rules := []string{"structure/tailwind-no-physical-direction", "@eslint-community/eslint-comments/require-description", "@next/next/google-font-display"}
	directories := []string{filepath.Join(root(t), "stage1"), "/tmp/lint-wave1-typescript/src/compiler"}
	if os.Getenv("ADAMIC_WAVE1_CORPUS") == "owned" {
		directories = nil
		for _, slug := range []string{"structure-tailwind-no-physical-direction", "eslint-comments-require-description", "next-google-font-display"} {
			directories = append(directories, filepath.Join(root(t), "stage1/cohere/lint/rules", slug))
		}
		t.Log("owned changed-source corpus only")
	}
	for _, directory := range directories {
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
			source, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			for _, rule := range rules {
				rows = append(rows, map[string]any{"file": path, "source": string(source), "rule": rule, "options": nil})
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	goOracle := oracle(t)
	entry, _ := filepath.Abs("main.a")
	binary, emitted := compile(t, entry)
	runner := filepath.Join(root(t), "oracle/node.mjs")
	for _, rule := range rules {
		if selection := os.Getenv("ADAMIC_WAVE1_RULES"); selection != "" && !strings.Contains(selection, rule) {
			continue
		}
		var subset []map[string]any
		for _, row := range rows {
			if row["rule"] == rule {
				subset = append(subset, row)
			}
		}
		totalBytes := 0
		for start := 0; start < len(subset); start += 10 {
			end := start + 10
			if end > len(subset) {
				end = len(subset)
			}
			encoded, _ := json.Marshal(subset[start:end])
			raw := file(t, "batch.json", encoded)
			want := command(t, "", goOracle, raw)
			adapted := file(t, "adapted.json", command(t, "", goOracle, "--ast", raw))
			for _, got := range runSides(t, entry, binary, emitted, adapted) {
				equal(t, got, want)
			}
			totalBytes += len(want)
		}
		t.Logf("corpus projected rule %s: %d files all three sides identical %d bytes", rule, len(subset), totalBytes)
	}
	// Probe independent parsing separately: source refusal must never be hidden by projection.
	encoded, _ := json.Marshal(rows[:3])
	raw := file(t, "source-probe.json", encoded)
	adapted := file(t, "source-probe-ast.json", command(t, "", goOracle, "--ast", raw))
	want := command(t, "", goOracle, raw)
	for side, cmd := range []*exec.Cmd{exec.Command("node", "--disable-warning=ExperimentalWarning", runner, entry, adapted, "--source"), exec.Command("node", "--disable-warning=ExperimentalWarning", runner, emitted, adapted, "--source"), exec.Command(binary, adapted, "--source")} {
		output, err := os.CreateTemp(t.TempDir(), "source-probe-")
		if err != nil {
			t.Fatal(err)
		}
		cmd.Stdout = output
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		err = cmd.Run()
		output.Close()
		if err != nil {
			t.Logf("independent source corpus probe side %d BLOCKED: %v %s", side, err, stderr.String())
		} else {
			got, _ := os.ReadFile(output.Name())
			equal(t, got, want)
			t.Logf("independent source probe side %d passed; full independent corpus not certified", side)
		}
	}
}

// Not parallel: sanitized native builds share the worker CPU and memory budget.
func TestFixtureRates(t *testing.T) {
	goOracle := oracle(t)
	entry, _ := filepath.Abs("main.a")
	binary, _ := compile(t, entry)
	runner := filepath.Join(root(t), "oracle/node.mjs")
	data, err := os.ReadFile("testdata/cases.json")
	if err != nil {
		t.Fatal(err)
	}
	var rows []map[string]any
	if err = json.Unmarshal(data, &rows); err != nil {
		t.Fatal(err)
	}
	for _, rule := range []string{"structure/tailwind-no-physical-direction", "@eslint-community/eslint-comments/require-description", "@next/next/google-font-display"} {
		var subset []map[string]any
		for _, row := range rows {
			if row["rule"] == rule {
				subset = append(subset, row)
			}
		}
		encoded, _ := json.Marshal(subset)
		raw := file(t, "cases.json", encoded)
		adapted := file(t, "ast.json", command(t, "", goOracle, "--ast", raw))
		want := ""
		for side, cmd := range []*exec.Cmd{exec.Command(goOracle, "--count", raw), exec.Command("node", "--disable-warning=ExperimentalWarning", runner, entry, adapted, "--count"), exec.Command(binary, adapted, "--count")} {
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
			if side == 0 {
				want = string(output)
			} else if string(output) != want {
				t.Fatal("rate count mismatch")
			}
			count, err := strconv.Atoi(strings.TrimSpace(string(output)))
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("fixture rate %s side %d findings=%d elapsed=%.6fs findings/s=%.3f", rule, side, count, seconds, float64(count)/seconds)
		}
	}
}
