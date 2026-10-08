package control_flow_graph

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func cfgRoot(t *testing.T) string {
	t.Helper()
	p, e := filepath.Abs("../../../../..")
	if e != nil {
		t.Fatal(e)
	}
	return p
}
func cfgRun(t *testing.T, dir string, env []string, name string, args ...string) []byte {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()
	c := exec.CommandContext(ctx, name, args...)
	c.Dir = dir
	c.Env = append(os.Environ(), env...)
	var errout bytes.Buffer
	c.Stderr = &errout
	out, e := c.Output()
	if e != nil || errout.Len() != 0 {
		t.Fatalf("%s: %v\n%s", name, e, errout.String())
	}
	return out
}
func cfgInflate(t *testing.T, path string) []byte {
	t.Helper()
	f, e := os.Open(path)
	if e != nil {
		t.Fatal(e)
	}
	defer f.Close()
	r, e := gzip.NewReader(f)
	if e != nil {
		t.Fatal(e)
	}
	defer r.Close()
	b, e := io.ReadAll(r)
	if e != nil {
		t.Fatal(e)
	}
	return b
}
func cfgWrite(t *testing.T, path string, b []byte) {
	t.Helper()
	if e := os.WriteFile(path, b, 0644); e != nil {
		t.Fatal(e)
	}
}
func cfgCorpus(t *testing.T) (string, []byte) {
	t.Helper()
	p := filepath.Join(t.TempDir(), "cases.json")
	cfgWrite(t, p, cfgInflate(t, "testdata/cases.json.gz"))
	return p, cfgInflate(t, "testdata/expected.txt.gz")
}
func cfgBuild(t *testing.T, entry string) (string, string) {
	t.Helper()
	program, e := load.Load([]string{entry})
	if e != nil {
		t.Fatal(e)
	}
	ir, e := lower.Lower(context.Background(), program)
	if e != nil {
		t.Fatal(e)
	}
	temp := t.TempDir()
	binary := filepath.Join(temp, "cfg")
	if e := native.Build(native.C(ir), binary, native.Options{Sanitize: true}); e != nil {
		t.Fatal(e)
	}
	js := filepath.Join(temp, "cfg.js")
	cfgWrite(t, js, []byte(javascript.JavaScript(ir)))
	return binary, js
}
func cfgCompare(t *testing.T, want, got []byte) {
	t.Helper()
	if bytes.Equal(want, got) {
		return
	}
	w, g := bytes.Split(want, []byte("\n")), bytes.Split(got, []byte("\n"))
	for i := 0; i < len(w) && i < len(g); i++ {
		if !bytes.Equal(w[i], g[i]) {
			t.Fatalf("first disagreement at line %d\nGo: %s\nport: %s", i+1, w[i], g[i])
		}
	}
	t.Fatalf("length disagreement %d / %d", len(want), len(got))
}
func TestFreshConsumingRuleCapture(t *testing.T) {
	t.Parallel()
	temp := t.TempDir()
	env := []string{"ADAMIC_CFG_FIXTURE_OUTPUT=" + temp}
	root := cfgRoot(t)
	base := filepath.Join(root, "stage1/cohere/lint/helpers/control_flow_graph/testdata")
	t.Log(string(cfgRun(t, root, env, "python3", filepath.Join(base, "capture.py"))))
	cfgRun(t, root, env, "python3", filepath.Join(base, "expected.py"))
	for _, name := range []string{"cases.json.gz", "expected.txt.gz", "capture_counts.json"} {
		a, e := os.ReadFile(filepath.Join("testdata", name))
		if e != nil {
			t.Fatal(e)
		}
		b, e := os.ReadFile(filepath.Join(temp, name))
		if e != nil {
			t.Fatal(e)
		}
		if !bytes.Equal(a, b) {
			t.Fatalf("fresh Go capture drift: %s", name)
		}
	}
}
func TestArenaHelpersAgreeWithGo(t *testing.T) {
	t.Parallel()
	cases, want := cfgCorpus(t)
	root := cfgRoot(t)
	entry, _ := filepath.Abs("main.a")
	runner := filepath.Join(root, "oracle/node.mjs")
	cfgCompare(t, want, cfgRun(t, root, nil, "node", "--disable-warning=ExperimentalWarning", runner, entry, cases))
	binary, js := cfgBuild(t, entry)
	cfgCompare(t, want, cfgRun(t, root, nil, "node", "--disable-warning=ExperimentalWarning", runner, js, cases))
	cfgCompare(t, want, cfgRun(t, root, nil, binary, cases))
	t.Logf("%d Go output lines / %d identical bytes on Node, emitted JavaScript and sanitized native", bytes.Count(want, []byte("\n")), len(want))
}

type cfgMutation struct{ Symbol, File, From, To string }

func cfgCopy(t *testing.T) string {
	t.Helper()
	temp := t.TempDir()
	dir := filepath.Join(temp, "cohere", "lint", "helpers", "control_flow_graph")
	if e := os.MkdirAll(dir, 0755); e != nil {
		t.Fatal(e)
	}
	files, e := filepath.Glob("*.a")
	if e != nil {
		t.Fatal(e)
	}
	for _, file := range files {
		data, e := os.ReadFile(file)
		if e != nil {
			t.Fatal(e)
		}
		cfgWrite(t, filepath.Join(dir, file), data)
	}
	data, e := os.ReadFile("../options_json.ts")
	if e != nil {
		t.Fatal(e)
	}
	cfgWrite(t, filepath.Join(dir, "..", "options_json.ts"), data)
	shared := filepath.Join(temp, "cohere", "arena")
	if e := os.MkdirAll(shared, 0755); e != nil {
		t.Fatal(e)
	}
	index, e := os.ReadFile("../../../arena/arena_index.a")
	if e != nil {
		t.Fatal(e)
	}
	cfgWrite(t, filepath.Join(shared, "arena_index.a"), index)
	return dir
}
func TestEveryHelperMutant(t *testing.T) {
	t.Parallel()
	cases, want := cfgCorpus(t)
	data, e := os.ReadFile("testdata/mutants.json")
	if e != nil {
		t.Fatal(e)
	}
	var mutants []cfgMutation
	if e = json.Unmarshal(data, &mutants); e != nil {
		t.Fatal(e)
	}
	if len(mutants) != 84 {
		t.Fatalf("need all 79 inventory helpers, 4 exposed query helpers and predecessor adapter, got %d", len(mutants))
	}
	root := cfgRoot(t)
	for _, m := range mutants {
		t.Run(m.Symbol, func(t *testing.T) {
			dir := cfgCopy(t)
			file := filepath.Join(dir, m.File)
			data, e := os.ReadFile(file)
			if e != nil {
				t.Fatal(e)
			}
			if !strings.Contains(string(data), m.From) {
				t.Fatal("mutant anchor drift")
			}
			cfgWrite(t, file, []byte(strings.Replace(string(data), m.From, m.To, 1)))
			entry := filepath.Join(dir, "main.a")
			node := cfgRun(t, root, nil, "node", "--disable-warning=ExperimentalWarning", filepath.Join(root, "oracle/node.mjs"), entry, cases)
			if bytes.Equal(node, want) {
				t.Fatal("Node mutant survived")
			}
			witness, witnessWant := cfgMutantWitness(t, cases, want, node)
			witnessNode := cfgRun(t, root, nil, "node", "--disable-warning=ExperimentalWarning", filepath.Join(root, "oracle/node.mjs"), entry, witness)
			if bytes.Equal(witnessNode, witnessWant) {
				t.Fatal("chosen Node witness survived")
			}
			binary, _ := cfgBuild(t, entry)
			got := cfgRun(t, root, nil, binary, witness)
			if bytes.Equal(got, witnessWant) {
				t.Fatal("native mutant survived")
			}
			t.Log("compiling semantic mutant caught on Node and sanitized native")
		})
	}
}
func TestCorruptedIndexStopsOnBothRuntimes(t *testing.T) {
	t.Parallel()
	dir := cfgCopy(t)
	file := filepath.Join(dir, "arena.a")
	data, e := os.ReadFile(file)
	if e != nil {
		t.Fatal(e)
	}
	anchor := "const slot = index.slot;"
	if !strings.Contains(string(data), anchor) {
		t.Fatal("index mutant anchor")
	}
	cfgWrite(t, file, []byte(strings.Replace(string(data), anchor, "const slot = index.slot+1;", 1)))
	entry := filepath.Join(dir, "index_corruption.a")
	cfgWrite(t, entry, []byte("import { BlockArena } from './arena.a';\nconst arena=new BlockArena();const index=arena.allocate();console.log(`${arena.read(index).index.slot}`);\n"))
	binary, _ := cfgBuild(t, entry)
	root := cfgRoot(t)
	for _, cmd := range [][]string{{"node", "--disable-warning=ExperimentalWarning", filepath.Join(root, "oracle/node.mjs"), entry}, {binary}} {
		c := exec.Command(cmd[0], cmd[1:]...)
		out, e := c.CombinedOutput()
		if e == nil || !bytes.Contains(out, []byte("CFG block index out of range")) {
			t.Fatalf("index corruption was not refused: %v %s", e, out)
		}
	}
}
func TestArenaLifetimeStopsOnBothRuntimes(t *testing.T) {
	t.Parallel()
	dir := cfgCopy(t)
	entry := filepath.Join(dir, "freed_arena.a")
	cfgWrite(t, entry, []byte("import { BlockArena } from './arena.a';\nconst arena=new BlockArena();const index=arena.allocate();arena.dispose();console.log(`${arena.read(index).index.slot}`);\n"))
	binary, _ := cfgBuild(t, entry)
	root := cfgRoot(t)
	for _, cmd := range [][]string{{"node", "--disable-warning=ExperimentalWarning", filepath.Join(root, "oracle/node.mjs"), entry}, {binary}} {
		c := exec.Command(cmd[0], cmd[1:]...)
		out, e := c.CombinedOutput()
		if e == nil || !bytes.Contains(out, []byte("freed CFG arena")) {
			t.Fatalf("read after arena disposal was not refused: %v %s", e, out)
		}
	}
}
func TestForeignIndexStopsOnBothRuntimes(t *testing.T) {
	t.Parallel()
	dir := cfgCopy(t)
	entry := filepath.Join(dir, "foreign_arena.a")
	cfgWrite(t, entry, []byte("import { BlockArena } from './arena.a';\nconst owner=new BlockArena();const other=new BlockArena();owner.allocate();const foreign=other.allocate();console.log(`${owner.read(foreign).index.slot}`);\n"))
	binary, _ := cfgBuild(t, entry)
	root := cfgRoot(t)
	for _, args := range [][]string{{"node", "--disable-warning=ExperimentalWarning", filepath.Join(root, "oracle/node.mjs"), entry}, {binary}} {
		out, e := exec.Command(args[0], args[1:]...).CombinedOutput()
		if e == nil || !bytes.Contains(out, []byte("CfgBlockIndex belongs to another arena")) {
			t.Fatalf("foreign index was not refused: %v %s", e, out)
		}
	}
}

func TestBlockAndNodeBrandsAreDistinct(t *testing.T) {
	t.Parallel()
	dir := cfgCopy(t)
	entry := filepath.Join(dir, "bad_brand.a")
	cfgWrite(t, entry, []byte("import { BlockArena } from './arena.a';\nimport { AstArena } from './ast.a';\nconst arena=new BlockArena();const ast=new AstArena();const wrong=ast.allocate();arena.read(wrong);\n"))
	program, e := load.Load([]string{entry})
	if e == nil {
		_, e = lower.Lower(context.Background(), program)
	}
	if e == nil || (!strings.Contains(e.Error(), "not assignable") && !strings.Contains(e.Error(), "nominal ancestry")) {
		t.Fatalf("brands were interchangeable: %v", e)
	}
}
func TestBlockIndexMintingIsPrivate(t *testing.T) {
	t.Parallel()
	dir := cfgCopy(t)
	for name, source := range map[string]string{
		"constructor": "import { CfgBlockIndex } from '../../../arena/arena_index.a';\nnew CfgBlockIndex(0);\n",
		"bare_number": "import { BlockArena } from './arena.a';\nconst arena=new BlockArena();arena.read(0);\n",
	} {
		t.Run(name, func(t *testing.T) {
			entry := filepath.Join(dir, name+".a")
			cfgWrite(t, entry, []byte(source))
			_, e := load.Load([]string{entry})
			if e == nil {
				t.Fatal("unchecked index mint/use accepted")
			}
			if name == "constructor" && !strings.Contains(e.Error(), "private") {
				t.Fatal(e)
			}
			if name == "bare_number" && !strings.Contains(e.Error(), "not assignable") {
				t.Fatal(e)
			}
		})
	}
}

// Each mutation selects its first discriminating original Go capture on Node.
// Native checks that identical witness; the full native corpus is checked by the baseline.
func cfgMutantWitness(t *testing.T, cases string, want, got []byte) (string, []byte) {
	t.Helper()
	w, g := bytes.Split(want, []byte("\n")), bytes.Split(got, []byte("\n"))
	first := 0
	for first < len(w) && first < len(g) && bytes.Equal(w[first], g[first]) {
		first++
	}
	start := first
	if start >= len(w) {
		start = len(w) - 1
	}
	for start > 0 && !bytes.HasPrefix(w[start], []byte("case:")) {
		start--
	}
	header := strings.Split(string(w[start]), ":")
	if len(header) != 3 {
		t.Fatal("mutant witness has no case header")
	}
	var index int
	if _, e := fmt.Sscanf(header[1], "%d", &index); e != nil {
		t.Fatal(e)
	}
	end := start + 1
	for end < len(w) && !bytes.HasPrefix(w[end], []byte("case:")) {
		end++
	}
	data, e := os.ReadFile(cases)
	if e != nil {
		t.Fatal(e)
	}
	var records []json.RawMessage
	if e = json.Unmarshal(data, &records); e != nil {
		t.Fatal(e)
	}
	if index < 0 || index >= len(records) {
		t.Fatal("mutant capture index out of bounds")
	}
	witness := filepath.Join(t.TempDir(), "witness.json")
	cfgWrite(t, witness, append(append([]byte("["), records[index]...), ']'))
	lines := append([][]byte(nil), w[start:end]...)
	lines[0] = []byte("case:0:" + header[2])
	output := bytes.Join(lines, []byte("\n"))
	if len(output) == 0 || output[len(output)-1] != '\n' {
		output = append(output, '\n')
	}
	return witness, output
}
