package validation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	registry "github.com/system-inc/adamic/stage1/cohere/lint/registry"
)

var slugs = []string{"func-name-matching", "consistent-this"}

func ruleName(slug string) string { return slug }
func projected(t *testing.T, oracle string, rows []string) []string {
	var result []string
	for _, row := range rows {
		fields := strings.Split(row, "\t")
		data := clean(t, run(t, "", oracle, "--ast", fields[0]))
		path := filepath.Join(t.TempDir(), "ast.json")
		if err := os.WriteFile(path, data, 0644); err != nil {
			t.Fatal(err)
		}
		for len(fields) < 6 {
			fields = append(fields, "")
		}
		fields = append(fields, path)
		result = append(result, strings.Join(fields, "\t"))
	}
	return result
}
func blocked(row string) bool { return false }

type observation struct {
	output   []byte
	stderr   string
	err      error
	duration time.Duration
}

func run(t *testing.T, dir string, args ...string) observation {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	cmd.Dir = dir
	output, err := os.CreateTemp(t.TempDir(), "stdout-")
	if err != nil {
		t.Fatal(err)
	}
	cmd.Stdout = output
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	start := time.Now()
	runErr := cmd.Run()
	elapsed := time.Since(start)
	if err = output.Close(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(output.Name())
	if err != nil {
		t.Fatal(err)
	}
	return observation{data, stderr.String(), runErr, elapsed}
}
func clean(t *testing.T, r observation) []byte {
	t.Helper()
	if r.err != nil || r.stderr != "" {
		t.Fatalf("exit=%v stderr=%s stdout=%s", r.err, r.stderr, r.output)
	}
	return r.output
}
func repo(t *testing.T) string {
	t.Helper()
	r, e := filepath.Abs("../../../../..")
	if e != nil {
		t.Fatal(e)
	}
	return r
}

var imports = regexp.MustCompile(`(?m)(^import\s+[^;]*?\s+from\s+)(['"])([^'"]+)(['"])`)

func copyPort(t *testing.T, mutationSlug, from, to string) string {
	t.Helper()
	root := filepath.Join(repo(t), "stage1/cohere/lint")
	destination := t.TempDir()
	changed := 0
	err := filepath.WalkDir(root, func(path string, e fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relative, _ := filepath.Rel(root, path)
		if e.IsDir() {
			if filepath.Dir(relative) == "rules" {
				owned := false
				for _, slug := range slugs {
					owned = owned || e.Name() == slug
				}
				if !owned {
					return filepath.SkipDir
				}
			}
			if relative == "gaps" || relative == "claims" || relative == ".generated" || e.Name() == "evidence" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".a") && !strings.HasSuffix(path, ".ts") && !strings.HasSuffix(path, ".json") && !strings.HasSuffix(path, "oracle.go") && !strings.HasSuffix(path, ".ts.txt") {
			return nil
		}
		if e.Name() == "rule.json" {
			owned := false
			for _, slug := range slugs {
				owned = owned || relative == filepath.Join("rules", slug, "rule.json")
			}
			if !owned {
				return nil
			}
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		source := string(data)
		if relative == filepath.Join("rules", mutationSlug, "rule.a") && from != "" {
			if strings.Count(source, from) != 1 {
				return fmt.Errorf("mutation anchor count differs in %s", relative)
			}
			source = strings.Replace(source, from, to, 1)
			changed++
		}
		if strings.HasSuffix(path, ".a") || strings.HasSuffix(path, ".ts") {
			source = imports.ReplaceAllStringFunc(source, func(declaration string) string {
				parts := imports.FindStringSubmatch(declaration)
				if !strings.HasPrefix(parts[3], ".") {
					return declaration
				}
				absolute := filepath.Clean(filepath.Join(filepath.Dir(path), parts[3]))
				rel, _ := filepath.Rel(root, absolute)
				if rel != ".." && !strings.HasPrefix(rel, "../") {
					return declaration
				}
				return parts[1] + parts[2] + filepath.ToSlash(absolute) + parts[4]
			})
		}
		target := filepath.Join(destination, relative)
		if err = os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return err
		}
		return os.WriteFile(target, []byte(source), 0644)
	})
	if err != nil {
		t.Fatal(err)
	}
	if from != "" && changed != 1 {
		t.Fatal("mutation did not change exactly one owned site")
	}
	descriptors, err := registry.Generate(destination)
	if err != nil {
		t.Fatal(err)
	}
	for _, slug := range slugs {
		found := false
		for _, d := range descriptors {
			found = found || d.Slug == slug
		}
		if !found {
			t.Fatal("owned rule not registered", slug)
		}
	}
	return destination
}
func goOracle(t *testing.T, port string, sources ...string) string {
	t.Helper()
	root := filepath.Join(repo(t), "cohere")
	dir := t.TempDir()
	replacements := map[string]string{}
	var files []string
	add := func(name, source string) {
		virtual := filepath.Join(root, "adamic_wave104_"+name+".go")
		replacements[virtual] = source
		files = append(files, virtual)
	}
	side, _ := filepath.Abs("validation_oracle.go.txt")
	if len(sources) > 0 {
		side = sources[0]
	}
	add("oracle", side)
	add("registry", filepath.Join(port, ".generated/registry.go"))
	descriptors, err := registry.Discover(port)
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range descriptors {
		add(strings.ReplaceAll(d.Slug, "-", "_"), filepath.Join(port, "rules", d.Slug, "oracle.go"))
	}
	encoded, _ := json.Marshal(map[string]any{"Replace": replacements})
	overlay := filepath.Join(dir, "overlay.json")
	if err = os.WriteFile(overlay, encoded, 0644); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(dir, "oracle")
	args := append([]string{"go", "build", "-overlay=" + overlay, "-o", binary}, files...)
	clean(t, run(t, root, args...))
	return binary
}
func builds(t *testing.T, port string, sanitize bool) (string, string, string) {
	t.Helper()
	entry := filepath.Join(port, "rules", slugs[0], "driver.a")
	program, err := load.Load([]string{entry})
	if err != nil {
		t.Fatal(err)
	}
	lowered, err := lower.Lower(context.Background(), program)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	binary := filepath.Join(dir, "lint")
	if err = native.Build(native.C(lowered), binary, native.Options{Sanitize: sanitize}); err != nil {
		t.Fatal(err)
	}
	js := filepath.Join(dir, "lint.mjs")
	if err = os.WriteFile(js, []byte(javascript.JavaScript(lowered)), 0644); err != nil {
		t.Fatal(err)
	}
	return entry, js, binary
}
func manifest(t *testing.T, rows []string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "manifest")
	if err := os.WriteFile(p, []byte(strings.Join(rows, "\n")+"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	return p
}
func sides(t *testing.T, entry, js, binary, path string, count bool) []struct {
	name string
	r    observation
} {
	runner := filepath.Join(repo(t), "oracle/node.mjs")
	suffix := []string{"--manifest", path}
	if count {
		suffix = append(suffix, "--count")
	}
	var result []struct {
		name string
		r    observation
	}
	for _, command := range []struct {
		name string
		args []string
	}{{"Node", []string{"node", "--disable-warning=ExperimentalWarning", runner, entry}}, {"emitted JavaScript", []string{"node", "--disable-warning=ExperimentalWarning", runner, js}}, {"sanitized native", []string{binary}}} {
		result = append(result, struct {
			name string
			r    observation
		}{command.name, run(t, "", append(command.args, suffix...)...)})
	}
	return result
}
func difference(got, want []byte) string {
	a, b := strings.Split(string(got), "\n"), strings.Split(string(want), "\n")
	for i := 0; i < len(a) || i < len(b); i++ {
		x, y := "<EOF>", "<EOF>"
		if i < len(a) {
			x = a[i]
		}
		if i < len(b) {
			y = b[i]
		}
		if x != y {
			start := i - 7
			if start < 0 {
				start = 0
			}
			end := i + 5
			if end > len(b) {
				end = len(b)
			}
			return fmt.Sprintf("line %d port=%q Go=%q\nGo context:\n%s", i+1, x, y, strings.Join(b[start:end], "\n"))
		}
	}
	return ""
}
func compare(t *testing.T, oracle, entry, js, binary, path string) []byte {
	t.Helper()
	want := clean(t, run(t, "", oracle, "--manifest", path))
	for _, side := range sides(t, entry, js, binary, path, false) {
		got := clean(t, side.r)
		if diff := difference(got, want); diff != "" {
			t.Fatalf("%s %s", side.name, diff)
		}
	}
	t.Logf("Go, Node, emitted JavaScript, ASan/UBSan native: %d identical bytes", len(want))
	return want
}

type captured struct {
	Rule, File, Source string
	Options            json.RawMessage
}

func upstream(t *testing.T) []string {
	t.Helper()
	root := filepath.Join(repo(t), "cohere")
	originalPath := filepath.Join(root, "internal/lint/testing/rule_testing.go")
	data, err := os.ReadFile(originalPath)
	if err != nil {
		t.Fatal(err)
	}
	anchor := "return Result{Diagnostics: diagnostics, SourceFile: sourceFile, capture: captured}"
	if strings.Count(string(data), anchor) != 1 {
		t.Fatal("capture anchor changed")
	}
	scratch := t.TempDir()
	copy := filepath.Join(scratch, "capture.go")
	data = []byte(strings.Replace(string(data), anchor, "result := Result{Diagnostics: diagnostics, SourceFile: sourceFile, capture: captured}\n RecordAssertedCase(t, result)\n return result", 1))
	if err = os.WriteFile(copy, data, 0644); err != nil {
		t.Fatal(err)
	}
	encoded, _ := json.Marshal(map[string]any{"Replace": map[string]string{originalPath: copy}})
	overlay := filepath.Join(scratch, "overlay.json")
	if err = os.WriteFile(overlay, encoded, 0644); err != nil {
		t.Fatal(err)
	}
	capture := filepath.Join(scratch, "records")
	t.Setenv("COHERE_DOCS_CAPTURE", capture)
	for _, spec := range []struct{ pkg, pattern string }{{"core", "^Test(FuncNameMatching|ConsistentThis)"}} {
		clean(t, run(t, root, "go", "test", "-overlay="+overlay, "-count=1", "-timeout=10m", "./internal/lint/rules/"+spec.pkg, "-run", spec.pattern))
	}
	files, err := filepath.Glob(filepath.Join(capture, "*.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	unique := map[string]captured{}
	for _, p := range files {
		data, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		for _, line := range bytes.Split(data, []byte("\n")) {
			if len(line) == 0 {
				continue
			}
			var row captured
			if err = json.Unmarshal(line, &row); err != nil {
				t.Fatal(err)
			}
			if false {
				continue
			}
			key := row.Rule + row.File + string(row.Options) + row.Source
			unique[key] = row
		}
	}
	var keys []string
	for key := range unique {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var rows []string
	var saved []captured
	counts := map[string]int{}
	for i, key := range keys {
		row := unique[key]
		saved = append(saved, row)
		path := filepath.Join(scratch, fmt.Sprintf("case-%03d%s", i, filepath.Ext(row.File)))
		if err = os.WriteFile(path, []byte(row.Source), 0644); err != nil {
			t.Fatal(err)
		}
		options := "null"
		if len(row.Options) > 0 {
			options = string(row.Options)
		}
		rows = append(rows, path+"\t"+row.Rule+"\t\t\tfalse\t"+options)
		counts[row.Rule]++
	}
	if evidence := os.Getenv("ADAMIC_WAVE104_EVIDENCE"); evidence != "" {
		if err := os.MkdirAll(evidence, 0755); err != nil {
			t.Fatal(err)
		}
		data, err := json.MarshalIndent(saved, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(filepath.Join(evidence, "upstream-cases.json"), data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	for _, slug := range slugs {
		name := ruleName(slug)
		if counts[name] == 0 {
			t.Fatal("missing upstream rule", name)
		}
		t.Logf("%s upstream unique configurations=%d", name, counts[name])
	}
	return rows
}
func witnesses(t *testing.T) []string {
	var rows []string
	for _, slug := range slugs {
		files, err := filepath.Glob(filepath.Join(repo(t), "stage1/cohere/lint/rules", slug, "testdata/*.ts.txt"))
		if err != nil {
			t.Fatal(err)
		}
		for _, file := range files {
			data, err := os.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}
			ext := ".ts"
			if slug == "next-google-font-display" {
				ext = ".tsx"
			}
			path := filepath.Join(t.TempDir(), "witness"+ext)
			if err = os.WriteFile(path, data, 0644); err != nil {
				t.Fatal(err)
			}
			rows = append(rows, path+"\t"+ruleName(slug))
		}
	}
	return rows
}

// Not parallel: bound compiler and sanitizer memory while sharing the captured corpus.
func TestRulesAndSuggestions(t *testing.T) {
	port := copyPort(t, "", "", "")
	oracle := goOracle(t, port)
	entry, js, binary := builds(t, port, true)
	var rows []string
	skipped := 0
	for _, row := range append(upstream(t), witnesses(t)...) {
		if blocked(row) {
			skipped++
			continue
		}
		rows = append(rows, row)
	}
	t.Logf("independent parser comparison: %d cases; %d JSX cases explicitly blocked", len(rows), skipped)
	compare(t, oracle, entry, js, binary, manifest(t, rows))
}

// Not parallel: build each compiling semantic mutant independently.
func TestOwnedMutants(t *testing.T) {
	baseline := copyPort(t, "", "", "")
	oracle := goOracle(t, baseline)
	var rows []string
	for _, row := range witnesses(t) {
		if blocked(row) {
			rows = append(rows, projected(t, oracle, []string{row})...)
		} else {
			rows = append(rows, row)
		}
	}
	path := manifest(t, rows)
	want := clean(t, run(t, "", oracle, "--manifest", path))
	for _, slug := range slugs {
		t.Run(slug, func(t *testing.T) {
			var change struct{ Name, File, From, To string }
			data, err := os.ReadFile(filepath.Join(repo(t), "stage1/cohere/lint/rules", slug, "mutant.json"))
			if err != nil {
				t.Fatal(err)
			}
			if err = json.Unmarshal(data, &change); err != nil {
				t.Fatal(err)
			}
			port := copyPort(t, slug, change.From, change.To)
			entry, js, binary := builds(t, port, true)
			for _, side := range sides(t, entry, js, binary, path, false) {
				got := clean(t, side.r)
				if bytes.Equal(got, want) {
					t.Fatal("mutant survived", side.name)
				}
				t.Logf("%s caught by Go comparison on %s: %s", change.Name, side.name, difference(got, want))
			}
		})
	}
}
func corpus(t *testing.T) []string {
	t.Helper()
	source := os.Getenv("ADAMIC_TYPESCRIPT_SOURCE")
	if source == "" {
		t.Fatal("ADAMIC_TYPESCRIPT_SOURCE must name the pinned compiler checkout")
	}
	pin := clean(t, run(t, "", "git", "-C", source, "rev-parse", "HEAD"))
	if strings.TrimSpace(string(pin)) != "050880ce59e30b356b686bd3144efe24f875ebc8" {
		t.Fatalf("wrong compiler pin: %s", pin)
	}
	var files []string
	for _, root := range []string{filepath.Join(source, "src/compiler"), filepath.Join(repo(t), "stage1")} {
		err := filepath.WalkDir(root, func(path string, e fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !e.IsDir() && (strings.HasSuffix(path, ".ts") || strings.HasSuffix(path, ".a")) {
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	sort.Strings(files)
	var rows []string
	for _, file := range files {
		for _, slug := range slugs {
			rows = append(rows, file+"\t"+ruleName(slug))
		}
	}
	t.Logf("whole compiler and stage1 corpus: %d files, %d rule/file pairs", len(files), len(rows))
	return rows
}

// Not parallel: run the full source corpus on each external execution path.
func TestCompilerAndStage1(t *testing.T) {
	port := copyPort(t, "", "", "")
	oracle := goOracle(t, port)
	entry, js, binary := builds(t, port, true)
	compare(t, oracle, entry, js, binary, manifest(t, corpus(t)))
}
func TestThroughput(t *testing.T) {
	if os.Getenv("ADAMIC_LINT_BENCH") != "1" {
		t.Skip("ADAMIC_LINT_BENCH=1")
	}
	port := copyPort(t, "", "", "")
	oracle := goOracle(t, port)
	entry, _, binary := builds(t, port, false)
	all := corpus(t)
	for _, slug := range slugs {
		name := ruleName(slug)
		var rows []string
		for _, row := range all {
			if strings.HasSuffix(row, "\t"+name) {
				rows = append(rows, row)
			}
		}
		path := manifest(t, rows)
		best := map[string]time.Duration{}
		var want []byte
		for round := 0; round < 3; round++ {
			for _, side := range []struct {
				name string
				r    observation
			}{{"Go", run(t, "", oracle, "--manifest", path, "--count")}, {"native", run(t, "", binary, "--manifest", path, "--count")}, {"Node", run(t, "", "node", "--disable-warning=ExperimentalWarning", filepath.Join(repo(t), "oracle/node.mjs"), entry, "--manifest", path, "--count")}} {
				got := clean(t, side.r)
				if want == nil {
					want = got
				}
				if !bytes.Equal(got, want) {
					t.Fatal("throughput count differs", side.name)
				}
				if best[side.name] == 0 || side.r.duration < best[side.name] {
					best[side.name] = side.r.duration
				}
			}
		}
		var count int
		if _, err := fmt.Sscan(string(want), &count); err != nil {
			t.Fatal(err)
		}
		for _, side := range []string{"native", "Node", "Go"} {
			t.Logf("%s %s best-of-3 %d findings / %.6fs = %.2f findings/s", name, side, count, best[side].Seconds(), float64(count)/best[side].Seconds())
		}
	}
}

// Not parallel: prove the explicit integration refusal before and after one
// compiling guard mutant, on every execution path, without editing shared code.

func TestJsxProjectedSemanticsAndParserBlocker(t *testing.T) {
	port := copyPort(t, "", "", "")
	oracle := goOracle(t, port)
	entry, js, binary := builds(t, port, true)
	var rows []string
	for _, row := range upstream(t) {
		if blocked(row) {
			rows = append(rows, row)
		}
	}
	t.Logf("Go AST projection only: %d JSX cases; this does not certify the independent parser", len(rows))
	compare(t, oracle, entry, js, binary, manifest(t, projected(t, oracle, rows)))
	// Each positive JSX fixture must refuse through the unchanged stage1 parser.
	for _, row := range rows {
		fields := strings.Split(row, "\t")
		data, _ := os.ReadFile(fields[0])
		if !bytes.Contains(data, []byte("<link href=")) && !bytes.Contains(data, []byte("<div className=")) {
			continue
		}
		path := manifest(t, []string{row})
		for _, side := range sides(t, entry, js, binary, path, false) {
			exit, ok := side.r.err.(*exec.ExitError)
			if !ok || exit.ExitCode() != 70 || !strings.Contains(side.r.stderr, "parser slice expected GreaterThanToken") {
				t.Fatalf("%s parser blocker changed: %v %s", side.name, side.r.err, side.r.stderr)
			}
		}
	}
	t.Log("Node, emitted JS, ASan/UBSan native: exact parser refusal observed")
}

func TestDecodedOptionCorners(t *testing.T) {
	port := copyPort(t, "", "", "")
	oracle := goOracle(t, port)
	entry, js, binary := builds(t, port, true)
	cases := []struct{ name, source, options string }{
		{"func-name-matching", `Object.defineProperty(o, '', {value: function foo(){}});`, `{"Direction":"always","considerPropertyDescriptor":true}`},
	}
	var rows []string
	for _, c := range cases {
		path := filepath.Join(t.TempDir(), "corner.ts")
		if err := os.WriteFile(path, []byte(c.source), 0644); err != nil {
			t.Fatal(err)
		}
		rows = append(rows, path+"\t"+c.name+"\t\t\tfalse\t"+c.options)
	}
	compare(t, oracle, entry, js, binary, manifest(t, rows))
}
