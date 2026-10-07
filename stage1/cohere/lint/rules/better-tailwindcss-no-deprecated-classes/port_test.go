package cores

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

// Not parallel: pins upstream package globals and bounds native compilation memory.
func TestDecisionCores(t *testing.T) {
	repo := root(t)
	cohere := filepath.Join(repo, "cohere")
	export, _ := filepath.Abs("testdata/oracle_test.go")
	virtual := filepath.Join(cohere, "internal/lint/rules/tailwind/zz_adamic_kernels_test.go")
	overlay, _ := json.Marshal(map[string]any{"Replace": map[string]string{virtual: export}})
	overlayPath := file(t, "overlay.json", overlay)
	rows := filepath.Join(t.TempDir(), "rows.json")
	wantPath := filepath.Join(t.TempDir(), "want.txt")
	cmd := exec.Command("go", "test", "-overlay="+overlayPath, "./internal/lint/rules/tailwind", "-run", "^TestAdamicDecisionCores$", "-count=1", "-v")
	cmd.Dir = cohere
	cmd.Env = append(os.Environ(), "ADAMIC_KERNEL_ROWS="+rows, "ADAMIC_KERNEL_WANT="+wantPath)
	log, err := os.CreateTemp(t.TempDir(), "go-")
	if err != nil {
		t.Fatal(err)
	}
	cmd.Stdout = log
	cmd.Stderr = log
	err = cmd.Run()
	log.Close()
	if err != nil {
		output, _ := os.ReadFile(log.Name())
		t.Fatalf("capture: %v %s", err, output)
	}
	want, err := os.ReadFile(wantPath)
	if err != nil {
		t.Fatal(err)
	}
	entry, _ := filepath.Abs("main.a")
	binary, js := compile(t, entry)
	for side, got := range runSides(t, entry, binary, js, rows) {
		equal(t, got, want)
		t.Logf("side %d: %d identical core-contract bytes", side, len(got))
	}
	for _, mutant := range []struct{ slug, from, to string }{{"better-tailwindcss-no-deprecated-classes", "major<entry.major", "major<=entry.major"}, {"better-tailwindcss-no-duplicate-classes", "if(!seen)", "if(seen)"}, {"better-tailwindcss-no-unknown-classes", "if(!systemPresent)", "if(systemPresent)"}} {
		module := filepath.Join(repo, "stage1/cohere/lint/rules", mutant.slug, "core.a")
		data, err := os.ReadFile(module)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Count(string(data), mutant.from) != 1 {
			t.Fatal("mutant anchor")
		}
		mutated := file(t, "core.a", absoluteImports(t, module, []byte(strings.Replace(string(data), mutant.from, mutant.to, 1))))
		source, err := os.ReadFile(entry)
		if err != nil {
			t.Fatal(err)
		}
		variant := file(t, "main.a", []byte(strings.Replace(string(absoluteImports(t, entry, source)), filepath.ToSlash(module), filepath.ToSlash(mutated), -1)))
		nb, nj := compile(t, variant)
		for side, got := range runSides(t, variant, nb, nj, rows) {
			if bytes.Equal(got, want) {
				t.Fatalf("%s mutant survived", mutant.slug)
			}
			t.Logf("%s mutant compiles and executes; output comparison catches side %d", mutant.slug, side)
		}
	}
	var inputs []map[string]any
	data, _ := os.ReadFile(rows)
	json.Unmarshal(data, &inputs)
	t.Logf("%d decision-core observations; full rule listeners not certified", len(inputs))
}

// Not parallel: bounds native compilation and source parsing memory.
func TestListeners(t *testing.T) {
	oracle := oracle(t)
	cases, _ := filepath.Abs("testdata/cases.json")
	want := command(t, "", oracle, cases)
	adapted := file(t, "ast.json", command(t, "", oracle, "--ast", cases))
	entry, _ := filepath.Abs("listeners.a")
	binary, js := compile(t, entry)
	for side, got := range runSides(t, entry, binary, js, adapted) {
		equal(t, got, want)
		t.Logf("projected listener side %d: %d finding/fix/first-plan bytes", side, len(got))
	}
	var projected []map[string]any
	data, err := os.ReadFile(adapted)
	if err != nil {
		t.Fatal(err)
	}
	json.Unmarshal(data, &projected)
	t.Logf("%d supported fixture rows; Go diagnostics with more than one edit excluded", len(projected))
	var rawRows []map[string]any
	data, _ = os.ReadFile(cases)
	json.Unmarshal(data, &rawRows)
	var sourceRows []map[string]any
	for _, row := range rawRows {
		if strings.HasSuffix(row["file"].(string), ".ts") {
			sourceRows = append(sourceRows, row)
		}
	}
	if len(sourceRows) > 0 {
		data, _ = json.Marshal(sourceRows)
		raw := file(t, "source.json", data)
		expected := command(t, "", oracle, raw)
		ast := file(t, "source-ast.json", command(t, "", oracle, "--ast", raw))
		for side, got := range runSides(t, entry, binary, js, ast, "--source") {
			equal(t, got, expected)
			t.Logf("independent side %d: %d source fixture rows", side, len(sourceRows))
		}
	}
	runner := filepath.Join(root(t), "oracle/node.mjs")
	for _, name := range []string{"better-tailwindcss/no-deprecated-classes", "better-tailwindcss/no-duplicate-classes"} {
		var grouped []map[string]any
		for _, row := range rawRows {
			if row["rule"] == name {
				grouped = append(grouped, row)
			}
		}
		data, _ := json.Marshal(grouped)
		raw := file(t, "rate-source.json", data)
		ast := file(t, "rate-ast.json", command(t, "", oracle, "--ast", raw))
		expected := ""
		for side, cmd := range []*exec.Cmd{exec.Command(oracle, "--count", raw), exec.Command("node", "--disable-warning=ExperimentalWarning", runner, entry, ast, "--count"), exec.Command(binary, ast, "--count")} {
			log, err := os.CreateTemp(t.TempDir(), "rate-")
			if err != nil {
				t.Fatal(err)
			}
			cmd.Stdout = log
			var stderr bytes.Buffer
			cmd.Stderr = &stderr
			start := time.Now()
			err = cmd.Run()
			elapsed := time.Since(start).Seconds()
			log.Close()
			output, _ := os.ReadFile(log.Name())
			if err != nil || stderr.Len() != 0 {
				t.Fatalf("rate %v %s", err, stderr.String())
			}
			count := strings.TrimSpace(string(output))
			if side == 0 {
				expected = count
			} else if count != expected {
				t.Fatal("finding count mismatch")
			}
			t.Logf("%s Go/Node/native side %d findings=%s seconds=%.6f", name, side, count, elapsed)
		}
	}

}

// Not parallel: bounds sanitizer compilation memory. Missing adapters must never print a clean result.
func TestRefusals(t *testing.T) {
	entry, _ := filepath.Abs("listeners.a")
	binary, js := compile(t, entry)
	runner := filepath.Join(root(t), "oracle/node.mjs")
	for _, control := range []struct{ name, source, message string }{{"better-tailwindcss/no-unknown-classes", "const value=1;", "NotYet: live Tailwind design-system and program adapter"}, {"better-tailwindcss/no-duplicate-classes", "mergeClassNames('flex flex flex');", "NotYet: multiple automatic edits per diagnostic in shared Finding"}, {"better-tailwindcss/no-deprecated-classes", "const arbitrary='flex';", "NotYet: Go-equivalent configured variable regex provider"}} {
		options := "null"
		if strings.Contains(control.message, "regex") {
			options = `{"variables":["^arbitrary$"]}`
		}
		data, _ := json.Marshal([]map[string]any{{"file": "/fixture.ts", "source": control.source, "sourceChunks": []string{}, "rule": control.name, "options": options}})
		input := file(t, "refusal.json", data)
		for side, cmd := range []*exec.Cmd{exec.Command("node", "--disable-warning=ExperimentalWarning", runner, entry, input, "--source"), exec.Command("node", "--disable-warning=ExperimentalWarning", runner, js, input, "--source"), exec.Command(binary, input, "--source")} {
			log, err := os.CreateTemp(t.TempDir(), "refusal-")
			if err != nil {
				t.Fatal(err)
			}
			cmd.Stdout = log
			var stderr bytes.Buffer
			cmd.Stderr = &stderr
			err = cmd.Run()
			log.Close()
			exit, exited := err.(*exec.ExitError)
			firstLine := strings.Split(stderr.String(), "\n")[0]
			output, _ := os.ReadFile(log.Name())
			if !exited || exit.ExitCode() != 70 || firstLine != "adamic: panic: "+control.message || len(output) != 0 {
				t.Fatalf("refusal side %d: %v %s", side, err, stderr.String())
			}
			t.Logf("side %d explicitly refused %s", side, control.message)
		}
	}
}

// Not parallel: compares compiler trees in bounded batches of five source/rule pairs.
func TestListenerCorpus(t *testing.T) {
	var rows []map[string]any
	files := 0
	for _, dir := range []string{filepath.Join(root(t), "stage1"), "/tmp/lint-wave1-typescript/src/compiler"} {
		err := filepath.WalkDir(dir, func(path string, entry os.DirEntry, err error) error {
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
			files++
			for _, name := range []string{"better-tailwindcss/no-deprecated-classes", "better-tailwindcss/no-duplicate-classes"} {
				rows = append(rows, map[string]any{"file": path, "source": string(data), "rule": name, "options": nil})
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	oracle := oracle(t)
	entry, _ := filepath.Abs("listeners.a")
	binary, js := compile(t, entry)
	total := 0
	supported := 0
	for begin := 0; begin < len(rows); begin += 5 {
		end := begin + 5
		if end > len(rows) {
			end = len(rows)
		}
		encoded, _ := json.Marshal(rows[begin:end])
		raw := file(t, "batch.json", encoded)
		want := command(t, "", oracle, raw)
		adapted := file(t, "ast.json", command(t, "", oracle, "--ast", raw))
		for _, got := range runSides(t, entry, binary, js, adapted) {
			equal(t, got, want)
		}
		var selected []map[string]any
		data, _ := os.ReadFile(adapted)
		json.Unmarshal(data, &selected)
		supported += len(selected)
		total += len(want)
	}
	t.Logf("%d files, %d/%d source/rule rows supported; %d finding/fix/first-plan bytes per backend; projected AST", files, supported, len(rows), total)
}
