package command

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

type row struct{ Name, Source string }
type answer struct{ Text, Error string }

func read(t *testing.T, path string) []byte {
	t.Helper()
	data, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	return data
}
func write(t *testing.T, path string, data []byte) {
	t.Helper()
	if e := os.MkdirAll(filepath.Dir(path), 0755); e != nil {
		t.Fatal(e)
	}
	if e := os.WriteFile(path, data, 0644); e != nil {
		t.Fatal(e)
	}
}
func run(t *testing.T, directory, command string, args ...string) []byte {
	t.Helper()
	c := exec.Command(command, args...)
	c.Dir = directory
	var stderr bytes.Buffer
	c.Stderr = &stderr
	out, e := c.Output()
	if e != nil {
		t.Fatalf("%s: %v\nstdout=%s\nstderr=%s", command, e, out, stderr.Bytes())
	}
	if stderr.Len() != 0 {
		t.Fatalf("%s stderr: %s", command, stderr.Bytes())
	}
	return out
}
func build(t *testing.T, root, entry, directory string, sanitize bool) string {
	t.Helper()
	start := time.Now()
	p, e := load.Load([]string{entry})
	if e != nil {
		t.Fatal(e)
	}
	ir, e := lower.Lower(context.Background(), p)
	if e != nil {
		t.Fatal(e)
	}
	source := native.C(ir)
	t.Logf("lower+emit=%s C=%d bytes", time.Since(start), len(source))
	options := native.Options{Sanitize: sanitize}
	lib, e := native.RuntimeLibrary("", options)
	if e != nil {
		t.Fatal(e)
	}
	c := filepath.Join(directory, "command.c")
	write(t, c, []byte(source))
	object := filepath.Join(directory, "command.o")
	archive := filepath.Join(directory, "cohere.a")
	binary := filepath.Join(directory, "cohere")
	flags := native.Flags(options)
	start = time.Now()
	run(t, root, "clang", append(append([]string{}, flags...), "-I", filepath.Dir(lib), "-c", c, "-o", object)...)
	t.Logf("C compile=%s", time.Since(start))
	run(t, root, "ar", "rcs", archive, object)
	args := append(append([]string{}, flags...), "-o", binary, archive)
	args = append(args, native.RuntimeLinkFlags(lib)...)
	args = append(args, "-lm")
	start = time.Now()
	run(t, root, "clang", args...)
	elapsed := time.Since(start)
	info, e := os.Stat(binary)
	if e != nil {
		t.Fatal(e)
	}
	a, e := os.Stat(archive)
	if e != nil {
		t.Fatal(e)
	}
	t.Logf("archive=%d binary=%d link=%s sanitized=%v", a.Size(), info.Size(), elapsed, sanitize)
	if keep := os.Getenv("ADAMIC_COMMAND_KEEP"); keep != "" && !sanitize && strings.HasSuffix(directory, "release") {
		write(t, filepath.Join(keep, "cohere"), read(t, binary))
		if e := os.Chmod(filepath.Join(keep, "cohere"), 0755); e != nil {
			t.Fatal(e)
		}
		write(t, filepath.Join(keep, "cohere.a"), read(t, archive))
		write(t, filepath.Join(keep, "command.c"), []byte(source))
	}
	return binary
}
func TestLinkedCommand(t *testing.T) {
	t.Parallel()
	root, e := filepath.Abs("../../..")
	if e != nil {
		t.Fatal(e)
	}
	temp := t.TempDir()
	var manifest []struct {
		Status string
		Files  []struct{ Fixture, SHA256 string }
	}
	if e = json.Unmarshal(read(t, "testdata/corpus.json"), &manifest); e != nil {
		t.Fatal(e)
	}
	if len(manifest) != 23 {
		t.Fatalf("pins=%d", len(manifest))
	}
	rows := []row{}
	proving := []row{}
	quiet := 0
	tsCases := 0
	for _, pin := range manifest {
		if pin.Status != "fetched" {
			t.Fatalf("corpus pin failed: %s", pin.Status)
		}
		for _, f := range pin.Files {
			data := read(t, filepath.Join("testdata", f.Fixture))
			if fmt.Sprintf("%x", sha256.Sum256(data)) != f.SHA256 {
				t.Fatalf("fixture hash %s", f.Fixture)
			}
			if strings.HasSuffix(f.Fixture, ".ts") {
				tsCases++
				proving = append(proving, row{filepath.Base(f.Fixture), string(data)})
				continue
			}
			rows = append(rows, row{filepath.Base(f.Fixture), string(data)})
			quiet++
		}
	}
	// Unchanged texts from cohere's native/core_test.go; its format wrapper owns normalization.
	controls := []row{{"probe.json", "{ \"a\": [1, 2],\n\"b\": {} }\n"}, {"package.json", "{\"name\":\"probe\",\"files\":[]}\n"}, {"probe.css", "a{color:red;\nmargin:0}\n/* comment */\n"}, {"probe.graphql", "query Q { a\n b(c: 1) }\n"}, {"probe.yaml", "a:   1\n# comment\nb: [ x,y ]\nc: |\n  line\n  line\n"}, {"probe.yml", "- a\n-   b: 'c'\n"}}
	for _, r := range controls {
		rows = append(rows, r, row{r.Name, "\ufeff" + r.Source}, row{r.Name, strings.ReplaceAll(r.Source, "\n", "\r\n")}, row{r.Name, "\ufeff" + strings.ReplaceAll(r.Source, "\n", "\r")})
	}
	rows = append(rows, row{"probe.json", "{}"}, row{"probe.css", "a{color:red}"}, row{"probe.yaml", "a: 1"}, row{"probe.graphql", "{a}"}, row{"probe.ts", "1+2;"}, row{"probe.a", "1+2;"})
	data, e := json.Marshal(rows)
	if e != nil {
		t.Fatal(e)
	}
	cases := filepath.Join(temp, "cases.json")
	write(t, cases, data)
	adapter := filepath.Join(root, "stage1/cohere/command/testdata/oracle.go")
	overlay, _ := json.Marshal(map[string]any{"Replace": map[string]string{filepath.Join(root, "cohere/command/formatter_comparison/main.go"): adapter}})
	op := filepath.Join(temp, "overlay.json")
	write(t, op, overlay)
	oracle := filepath.Join(temp, "oracle")
	run(t, filepath.Join(root, "cohere"), "go", "build", "-overlay", op, "-o", oracle, "./command/formatter_comparison")
	var expected []answer
	if e = json.Unmarshal(run(t, root, oracle, cases), &expected); e != nil {
		t.Fatal(e)
	}
	if len(expected) != len(rows) {
		t.Fatal("oracle lost rows")
	}
	for i, r := range expected {
		if r.Error != "" {
			t.Fatalf("oracle refused %s: %s", rows[i].Name, r.Error)
		}
	}
	t.Logf("quiet files=%d, retained TS proving files=%d, Go controls=%d, total=%d", quiet, tsCases, len(controls)*4+6, len(rows))
	binary := build(t, root, filepath.Join(root, "stage1/cohere/command/main.a"), filepath.Join(temp, "release"), false)
	check := func(t *testing.T, executable string, prefix []string, wantMismatch bool) {
		t.Helper()
		directory := t.TempDir()
		args := append(append([]string{}, prefix...), "--format-only", "--format-all", "--no-cache", "--directory", directory)
		for i, r := range rows {
			name := fmt.Sprintf("%03d/%s", i, r.Name)
			write(t, filepath.Join(directory, name), []byte(r.Source))
			args = append(args, name)
		}
		run(t, root, executable, args...)
		differences := 0
		for i, r := range rows {
			got := read(t, filepath.Join(directory, fmt.Sprintf("%03d/%s", i, r.Name)))
			if !bytes.Equal(got, []byte(expected[i].Text)) {
				differences++
				if !wantMismatch {
					t.Fatalf("case %d %s differs: got %q want %q", i, r.Name, got, expected[i].Text)
				}
			}
		}
		if wantMismatch && differences == 0 {
			t.Fatal("mutant survived")
		}
		t.Logf("compared %d files differences=%d", len(rows), differences)
	}
	check(t, binary, nil, false)
	// Per-package executables cannot be linked together: both export main.
	second := filepath.Join(temp, "second.a")
	write(t, second, []byte("console.log('');\n"))
	build(t, root, second, filepath.Join(temp, "second"), false)
	library, err := native.RuntimeLibrary("", native.Options{})
	if err != nil {
		t.Fatal(err)
	}
	collisionArgs := []string{filepath.Join(temp, "release/command.o"), filepath.Join(temp, "second/command.o")}
	collisionArgs = append(collisionArgs, native.RuntimeLinkFlags(library)...)
	collisionArgs = append(collisionArgs, "-lm", "-o", filepath.Join(temp, "collision"))
	collision, err := exec.Command("clang", collisionArgs...).CombinedOutput()
	if err == nil || !strings.Contains(string(collision), "main") {
		t.Fatalf("expected duplicate main: %v %s", err, collision)
	}
	t.Logf("independent entry objects refuse to link: %s", collision)

	cli := filepath.Join(temp, "go-cohere")
	run(t, filepath.Join(root, "cohere"), "go", "build", "-o", cli, "./command/cohere")
	goFiles := t.TempDir()
	write(t, filepath.Join(goFiles, "tsconfig.json"), []byte(`{"files":["057/probe.ts","058/probe.a"]}`))
	write(t, filepath.Join(goFiles, "NexusCohereSettings.json"), []byte(`{"format":{}}`))
	write(t, filepath.Join(goFiles, "CohereSettings.json"), []byte(`{"extends":"./NexusCohereSettings.json"}`))
	cliArgs := []string{"--format-only", "--format-all", "--no-cache", "--directory", goFiles}
	for i, r := range rows {
		name := fmt.Sprintf("%03d/%s", i, r.Name)
		write(t, filepath.Join(goFiles, name), []byte(r.Source))
		cliArgs = append(cliArgs, name)
	}
	cliArgs = append([]string{"--json"}, cliArgs...)
	cliRun := exec.Command(cli, cliArgs...)
	cliRun.Dir = root
	var cliError bytes.Buffer
	cliRun.Stderr = &cliError
	summary, cliErr := cliRun.Output()
	cliExit, ok := cliErr.(*exec.ExitError)
	if !ok || cliExit.ExitCode() != 1 || cliError.Len() != 0 {
		t.Fatalf("Go CLI: %v stdout=%s stderr=%s", cliErr, summary, cliError.Bytes())
	}
	findings := []map[string]any{}
	decoder := json.NewDecoder(bytes.NewReader(summary))
	for decoder.More() {
		var record map[string]any
		if e := decoder.Decode(&record); e != nil {
			t.Fatal(e)
		}
		if record["kind"] == "finding" {
			for _, key := range []string{"path", "message"} {
				value, ok := record[key].(string)
				if !ok {
					t.Fatal("CLI finding field")
				}
				record[key] = strings.ReplaceAll(value, goFiles, "$ROOT")
			}
			findings = append(findings, record)
		}
	}
	var wantFindings []map[string]any
	if e := json.Unmarshal(read(t, "testdata/cli-refusals.json"), &wantFindings); e != nil {
		t.Fatal(e)
	}
	gotFindings, _ := json.Marshal(findings)
	pinnedFindings, _ := json.Marshal(wantFindings)
	if !bytes.Equal(gotFindings, pinnedFindings) {
		t.Fatalf("CLI refusals changed: %s", gotFindings)
	}
	refused := map[int]bool{5: true, 8: true, 24: true, 34: true, 36: true}
	for i, r := range rows {
		got := read(t, filepath.Join(goFiles, fmt.Sprintf("%03d/%s", i, r.Name)))
		want := expected[i].Text
		if refused[i] {
			want = r.Source
		}
		if !bytes.Equal(got, []byte(want)) {
			t.Fatalf("Go CLI case %d %s differs", i, r.Name)
		}
	}
	t.Logf("Go command flags: 54 exact formatted outputs, 5 exact findings with unchanged input; all 59 held")
	gateFile := filepath.Join(goFiles, "gate.json")
	write(t, gateFile, []byte("0"))
	gate := exec.Command(cli, "--format-only", "--format-all", "--no-cache", "--no-fix", "--directory", goFiles, gateFile)
	gateOut, gateErr := gate.CombinedOutput()
	exit, ok := gateErr.(*exec.ExitError)
	if !ok || exit.ExitCode() != 1 {
		t.Fatalf("Go no-fix gate: %v %q", gateErr, gateOut)
	}
	if string(read(t, gateFile)) != "0" {
		t.Fatal("Go no-fix mutated input")
	}
	t.Logf("Go no-fix shortest input 0: exit=1, source unchanged, output=%q", gateOut)

	entry := filepath.Join(root, "stage1/cohere/command/main.a")
	node := []string{"--disable-warning=ExperimentalWarning", filepath.Join(root, "oracle/node.mjs"), entry}
	check(t, "node", node, false)
	// Whole compiler cases are held as explicit refusals, never credited as formats.
	proving = append(proving, row{"comment.ts", "//"})
	for _, r := range proving {
		file := filepath.Join(t.TempDir(), r.Name)
		write(t, file, []byte(r.Source))
		for _, invocation := range [][]string{{binary, "--format-only", file}, append(append([]string{"node"}, node...), "--format-only", file)} {
			cmd := exec.Command(invocation[0], invocation[1:]...)
			out, err := cmd.CombinedOutput()
			exit, ok := err.(*exec.ExitError)
			if !ok || exit.ExitCode() != 70 || string(out) != "adamic: panic: comment-attachment\n" {
				t.Fatalf("proving %s: %v %q", r.Name, err, out)
			}
		}
	}
	proofJSON, err := json.Marshal(proving)
	if err != nil {
		t.Fatal(err)
	}
	proofPath := filepath.Join(temp, "proving.json")
	write(t, proofPath, proofJSON)
	var proofAnswers []answer
	if err = json.Unmarshal(run(t, root, oracle, proofPath), &proofAnswers); err != nil {
		t.Fatal(err)
	}
	pinnedProof := read(t, "testdata/proof-go.json")
	actualProof, err := json.Marshal(proofAnswers)
	if err != nil {
		t.Fatal(err)
	}
	var wantProof []answer
	if err = json.Unmarshal(pinnedProof, &wantProof); err != nil {
		t.Fatal(err)
	}
	wantBytes, err := json.Marshal(wantProof)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(actualProof, wantBytes) {
		t.Fatalf("Go proving outcomes changed: %s", actualProof)
	}
	t.Logf("whole-file unsupported proofs=%d; Go formats 2 and refuses 2 raw multi-file compiler cases; Node/native explicitly refuse all 4", len(proving))
	sanitized := build(t, root, entry, filepath.Join(temp, "sanitized"), true)
	check(t, sanitized, nil, false)
	// Each mutation changes only this lane's composition and must build and finish successfully.
	original := string(read(t, "format.a"))
	mutations := [][2]string{{"(bom ? '\\ufeff' : '') + formatted", "formatted"}, {"jsonFormat(name, text)", "jsonFormat('ordinary.json', text)"}, {"const result = cssFormat(text);", "const result = cssFormat(text, false, { width: 80, tabWidth: 4, tabs: false, single: false, trailingComma: 'all' });"}}
	for i, m := range mutations {
		t.Run(fmt.Sprintf("mutant%d", i), func(t *testing.T) {
			if strings.Count(original, m[0]) != 1 {
				t.Fatal("mutant anchor")
			}
			dir := t.TempDir() // Absolute imports preserve the original package graph without copying shared modules.
			source := strings.Replace(original, m[0], m[1], 1)
			relative, err := filepath.Rel(dir, filepath.Join(root, "stage1/cohere"))
			if err != nil {
				t.Fatal(err)
			}
			source = strings.ReplaceAll(source, "from '../", "from '"+filepath.ToSlash(relative)+"/")
			write(t, filepath.Join(dir, "format.a"), []byte(source))
			write(t, filepath.Join(dir, "main.a"), read(t, "main.a"))
			mutant := build(t, root, filepath.Join(dir, "main.a"), filepath.Join(dir, "build"), false)
			check(t, mutant, nil, true)
			check(t, "node", []string{"--disable-warning=ExperimentalWarning", filepath.Join(root, "oracle/node.mjs"), filepath.Join(dir, "main.a")}, true)
		})
	}
}
