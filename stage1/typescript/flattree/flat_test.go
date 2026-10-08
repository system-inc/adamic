package flattree

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

const repo = "../../.."
const pin = "050880ce59e30b356b686bd3144efe24f875ebc8"

func run(t *testing.T, dir, name string, args ...string) []byte {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	c := exec.CommandContext(ctx, name, args...)
	c.Dir = dir
	// File stdout holds Node's entire large writev output before reading it.
	f, e := os.CreateTemp(t.TempDir(), "output")
	if e != nil {
		t.Fatal(e)
	}
	defer f.Close()
	c.Stdout = f
	var stderr bytes.Buffer
	c.Stderr = &stderr
	if e := c.Run(); e != nil {
		t.Fatalf("%s: %v\n%s", name, e, &stderr)
	}
	if stderr.Len() != 0 {
		t.Fatalf("%s stderr: %s", name, &stderr)
	}
	b, e := os.ReadFile(f.Name())
	if e != nil {
		t.Fatal(e)
	}
	return b
}
func integer(t *testing.T, s string) int {
	t.Helper()
	v, e := strconv.Atoi(s)
	if e != nil {
		t.Fatal(e)
	}
	return v
}
func indices(t *testing.T, s string) []int {
	t.Helper()
	var r []int
	if s == "" {
		return r
	}
	for _, v := range strings.Split(s, ",") {
		r = append(r, integer(t, v))
	}
	return r
}
func snapshots(t *testing.T, b []byte) []Tree {
	t.Helper()
	var trees []Tree
	for _, line := range strings.Split(strings.TrimSuffix(string(b), "\n"), "\n") {
		if strings.HasPrefix(line, "case ") {
			v := strings.Split(line, " ")
			if len(v) != 4 {
				t.Fatalf("header %q", line)
			}
			trees = append(trees, Tree{Root: integer(t, v[2]), Roots: indices(t, v[3])})
			continue
		}
		f := strings.Split(line, "\t")
		if len(f) != 13 || len(trees) == 0 {
			t.Fatalf("node %q", line)
		}
		n := Node{Kind: f[0], Pos: integer(t, f[1]), End: integer(t, f[2]), Flags: integer(t, f[3]), LiteralFlags: integer(t, f[4]), List: integer(t, f[5]), Trailing: f[6] == "1", MultiLine: f[7] == "1", Operator: f[8], Text: f[9], Raw: f[10], Semantic: f[11], Children: indices(t, f[12])}
		i := len(trees) - 1
		trees[i].Nodes = append(trees[i].Nodes, n)
	}
	return trees
}
func escapedUnits(b []byte, escaped bool) string {
	var out strings.Builder
	for i := 0; i < len(b); i += 2 {
		u := binary.LittleEndian.Uint16(b[i:])
		if escaped && (u < 32 || u > 126 || u == 92) {
			fmt.Fprintf(&out, `\u%04x`, u)
		} else {
			out.WriteRune(rune(u))
		}
	}
	return out.String()
}
func restore(r Reader) Tree {
	t := Tree{Root: int(r.Root())}
	for j := uint32(0); j < r.roots; j++ {
		t.Roots = append(t.Roots, int(binary.LittleEndian.Uint32(r.data[r.rootOff+int(j)*4:])))
	}
	for i := uint32(0); i < r.NodeCount(); i++ {
		s := func(c int, escaped bool) string { return escapedUnits(r.StringUnits(r.Value(c, i)), escaped) }
		flags := r.Value(3, i)
		n := Node{Kind: s(0, false), Pos: int(r.Value(1, i)), End: int(r.Value(2, i)), Flags: int(flags & 32), LiteralFlags: int(r.Value(4, i)), List: int(int32(r.Value(5, i))), Trailing: flags&(1<<30) != 0, MultiLine: flags&(1<<29) != 0, Text: s(8, true), Raw: s(9, true), Operator: s(10, false), Semantic: s(11, false)}
		for j := uint32(0); j < r.Value(7, i); j++ {
			n.Children = append(n.Children, int(r.Child(i, j)))
		}
		t.Nodes = append(t.Nodes, n)
	}
	return t
}
func snapshot(t Tree, caseNumber int) []byte {
	var b strings.Builder
	fmt.Fprintf(&b, "case %d %d %s\n", caseNumber, t.Root, join(t.Roots))
	for _, n := range t.Nodes {
		fmt.Fprintf(&b, "%s\t%d\t%d\t%d\t%d\t%d\t%d\t%d\t%s\t%s\t%s\t%s\t%s\n", n.Kind, n.Pos, n.End, n.Flags, n.LiteralFlags, n.List, bit(n.Trailing), bit(n.MultiLine), n.Operator, n.Text, n.Raw, n.Semantic, join(n.Children))
	}
	return []byte(b.String())
}
func bit(v bool) int {
	if v {
		return 1
	}
	return 0
}
func join(v []int) string {
	s := make([]string, len(v))
	for i, n := range v {
		s[i] = strconv.Itoa(n)
	}
	return strings.Join(s, ",")
}
func offsets(source string) []int {
	r := []int{0}
	bytes := 0
	for _, c := range source {
		if c > 65535 {
			r = append(r, bytes)
		}
		bytes += len(string(c))
		r = append(r, bytes)
	}
	return r
}
func canonical(t Tree, source string, caseNumber int) []byte {
	var b strings.Builder
	fmt.Fprintf(&b, "case %d\nfile\n", caseNumber)
	o := offsets(source)
	var visit func(int, int)
	visit = func(id, depth int) {
		n := t.Nodes[id]
		fmt.Fprintf(&b, "%d %s %d %d %d %d %d %d %d\t%s\t%s\t%s\t%s\n", depth, n.Kind, o[n.Pos], o[n.End], n.Flags, n.LiteralFlags, n.List, bit(n.Trailing), bit(n.MultiLine), n.Operator, n.Text, n.Raw, n.Semantic)
		for _, c := range n.Children {
			visit(c, depth+1)
		}
	}
	visit(t.Root, 0)
	return []byte(b.String())
}
func oracle(t *testing.T) string {
	t.Helper()
	root, _ := filepath.Abs(filepath.Join(repo, "cohere/TypeScript/tsc"))
	source, _ := filepath.Abs("../parser/testdata/oracle.go")
	virtual := filepath.Join(root, "flat_scout_oracle.go")
	b, _ := json.Marshal(map[string]any{"Replace": map[string]string{virtual: source}})
	d := t.TempDir()
	p := filepath.Join(d, "overlay.json")
	if e := os.WriteFile(p, b, 0644); e != nil {
		t.Fatal(e)
	}
	binary := filepath.Join(d, "oracle")
	run(t, root, "go", "build", "-overlay="+p, "-o", binary, virtual)
	return binary
}
func corpus(t *testing.T) ([]string, string) {
	t.Helper()
	root := os.Getenv("ADAMIC_TYPESCRIPT_SOURCE")
	if root == "" {
		t.Fatal("ADAMIC_TYPESCRIPT_SOURCE must name the pinned TypeScript checkout; no skipped inputs")
	}
	if strings.TrimSpace(string(run(t, "", "git", "-C", root, "rev-parse", "HEAD"))) != pin {
		t.Fatal("wrong TypeScript pin")
	}
	var files []string
	e := filepath.WalkDir(filepath.Join(root, "src/compiler"), func(path string, d os.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if !d.IsDir() && strings.HasSuffix(path, ".ts") {
			files = append(files, path)
		}
		return nil
	})
	if e != nil {
		t.Fatal(e)
	}
	sort.Strings(files)
	if len(files) != 77 {
		t.Fatalf("compiler files=%d, want 77", len(files))
	}
	// These are raw witnesses extracted from Go cohere tests, with provenance.
	fixtures, e := filepath.Glob("testdata/upstream/*.ts.txt")
	if e != nil || len(fixtures) != 23 {
		t.Fatalf("upstream fixtures=%d err=%v", len(fixtures), e)
	}
	for _, p := range fixtures {
		a, _ := filepath.Abs(p)
		files = append(files, a)
	}

	for _, family := range []string{"public", "typescript", "edges"} {
		fixturePaths, err := filepath.Glob("testdata/" + family + "/*.ts.txt")
		if err != nil || len(fixturePaths) == 0 {
			t.Fatalf("missing family %s: %v", family, err)
		}
		if family == "public" && len(fixturePaths) != 23 {
			t.Fatalf("public pins=%d", len(fixturePaths))
		}
		for _, path := range fixturePaths {
			absolute, _ := filepath.Abs(path)
			files = append(files, absolute)
		}
	}
	for _, path := range []string{"../parser/nodes.ts", "../parser/main.ts", "../../cohere/lint/lint.ts", "../../cohere/tsprinter/expressions.ts"} {
		absolute, _ := filepath.Abs(path)
		files = append(files, absolute)
	}
	manifest := filepath.Join(t.TempDir(), "manifest")
	if e := os.WriteFile(manifest, []byte(strings.Join(files, "\n")+"\n"), 0644); e != nil {
		t.Fatal(e)
	}
	return files, manifest
}
func TestCorpusRoundTrip(t *testing.T) {
	t.Parallel()
	files, manifest := corpus(t)
	runner, _ := filepath.Abs(filepath.Join(repo, "oracle/node.mjs"))
	driver, _ := filepath.Abs("snapshot.ts")
	raw := run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, driver, manifest)
	trees := snapshots(t, raw)
	if len(trees) != len(files) {
		t.Fatalf("trees=%d files=%d", len(trees), len(files))
	}
	want := run(t, "", oracle(t), "--manifest", manifest, "--whole")
	var got, round bytes.Buffer
	nodes, total, children, stringsCount := 0, 0, 0, 0
	for i, tree := range trees {
		source, e := os.ReadFile(files[i])
		if e != nil {
			t.Fatal(e)
		}
		b, e := Encode(tree)
		if e != nil {
			t.Fatal(e)
		}
		r, e := Open(b)
		if e != nil {
			t.Fatal(e)
		}
		back := restore(r)
		again, e := Encode(back)
		if e != nil || !bytes.Equal(again, b) {
			t.Fatalf("binary round-trip case %d: %v", i, e)
		}
		if !bytes.Equal(snapshot(back, i), snapshot(tree, i)) {
			t.Fatalf("transport round-trip case %d", i)
		}
		got.Write(canonical(back, string(source), i))
		round.Write(snapshot(back, i))
		path := filepath.Join(t.TempDir(), "tree.flat")
		if e := os.WriteFile(path, b, 0644); e != nil {
			t.Fatal(e)
		}
		m, e := Map(path)
		if e != nil {
			t.Fatal(e)
		}
		if m.Reader.Walk(m.Reader.Root()) != r.Walk(r.Root()) {
			t.Fatal("mmap walk differs")
		}
		if e := m.Close(); e != nil {
			t.Fatal(e)
		}
		if node := run(t, "", "node", "reference.mjs", path, strconv.Itoa(i)); !bytes.Equal(node, snapshot(tree, i)) {
			t.Fatalf("Node flat reader case %d", i)
		}
		if destination := os.Getenv("ADAMIC_FLAT_ARTIFACTS"); destination != "" && strings.HasSuffix(files[i], "/src/compiler/parser.ts") {
			if e := os.MkdirAll(destination, 0755); e != nil {
				t.Fatal(e)
			}
			if e := os.WriteFile(filepath.Join(destination, "parser.flat"), b, 0644); e != nil {
				t.Fatal(e)
			}
			if e := os.WriteFile(filepath.Join(destination, "parser.snapshot"), snapshot(tree, 0), 0644); e != nil {
				t.Fatal(e)
			}
		}
		nodes += len(tree.Nodes)
		total += len(b)
		children += int(r.children)
		stringsCount += int(r.strings)
	}
	if !bytes.Equal(round.Bytes(), raw) {
		t.Fatal("full snapshot bytes differ")
	}
	// Keep the known shared-parser defect explicit and pinned. It is never
	// counted as oracle agreement or silently dropped from round-trip coverage.
	starts := regexp.MustCompile(`(?m)^case [0-9]+\n`).FindAllIndex(want, -1)
	if len(starts) != len(trees) {
		t.Fatal("oracle case count")
	}
	boundaryBytes, e := os.ReadFile("testdata/parser-boundary.json")
	if e != nil {
		t.Fatal(e)
	}
	var boundary struct {
		Fixture string `json:"fixture"`
		Node    string `json:"node_sha256"`
		Go      string `json:"go_sha256"`
	}
	if e := json.Unmarshal(boundaryBytes, &boundary); e != nil {
		t.Fatal(e)
	}
	mismatches := 0
	for i, tree := range trees {
		end := len(want)
		if i+1 < len(starts) {
			end = starts[i+1][0]
		}
		reference := want[starts[i][1]:end]
		source, e := os.ReadFile(files[i])
		if e != nil {
			t.Fatal(e)
		}
		actual := canonical(tree, string(source), i)
		actual = actual[bytes.IndexByte(actual, '\n')+1:]
		if strings.HasSuffix(files[i], "/"+boundary.Fixture) {
			if fmt.Sprintf("%x", sha256.Sum256(actual)) != boundary.Node || fmt.Sprintf("%x", sha256.Sum256(reference)) != boundary.Go || bytes.Equal(actual, reference) {
				t.Fatal("shared parser boundary changed; remeasure it")
			}
			mismatches++
			t.Logf("BLOCKED parser oracle parity: %s Node=%d Go=%d bytes", files[i], len(actual), len(reference))
		} else if !bytes.Equal(actual, reference) {
			t.Fatalf("unclassified Go/Node difference: %s", files[i])
		}
	}
	if mismatches != 1 {
		t.Fatalf("parser boundaries=%d, want 1", mismatches)
	}
	if destination := os.Getenv("ADAMIC_FLAT_ARTIFACTS"); destination != "" {
		os.WriteFile(filepath.Join(destination, "manifest"), []byte(strings.Join(files, "\n")), 0644)
	}

	t.Logf("files=%d table_nodes=%d children=%d per-file_strings=%d flat_bytes=%d bytes/table_node=%.3f canonical_bytes=%d snapshot_bytes=%d", len(files), nodes, children, stringsCount, total, float64(total)/float64(nodes), len(want), len(raw))
}
func sample() Tree {
	return Tree{Root: 2, Roots: []int{2}, Nodes: []Node{{Kind: "Identifier", End: 1, List: -1, Text: "x"}, {Kind: "StringLiteral", Pos: 2, End: 10, List: -1, Text: `\ud800`, Raw: `\u005cud800`}, {Kind: "BinaryExpression", End: 10, List: 2, Flags: 32, Trailing: true, MultiLine: true, Operator: "PlusToken", Semantic: "2", Children: []int{0, 1}}}}
}
func TestFormatAndValidation(t *testing.T) {
	t.Parallel()
	tree := sample()
	b, e := Encode(tree)
	if e != nil {
		t.Fatal(e)
	}
	r, e := Open(b)
	if e != nil {
		t.Fatal(e)
	}
	if !bytes.Equal(snapshot(tree, 0), snapshot(restore(r), 0)) {
		t.Fatal("all fields not retained")
	}

	for name, mutate := range map[string]func([]byte){"magic": func(x []byte) { x[0] ^= 1 }, "length": func(x []byte) { x[8] ^= 1 }, "version": func(x []byte) { x[7] ^= 1 }, "root": func(x []byte) { binary.LittleEndian.PutUint32(x[32:], 99) }, "child": func(x []byte) { binary.LittleEndian.PutUint32(x[r.childOff:], 99) }, "range": func(x []byte) { binary.LittleEndian.PutUint32(x[headerSize+(6*3+2)*4:], 99) }, "string": func(x []byte) { binary.LittleEndian.PutUint32(x[headerSize:], 99) }, "pool": func(x []byte) { binary.LittleEndian.PutUint32(x[r.stringOff:], 999) }, "cycle": func(x []byte) { binary.LittleEndian.PutUint32(x[r.childOff:], 2) }, "reserved": func(x []byte) { x[36] = 1 }} {
		t.Run(name, func(t *testing.T) {
			x := bytes.Clone(b)
			mutate(x)
			if _, e := Open(x); e == nil {
				t.Fatal("validation mutant survived")
			}
		})
	}
	for i := 0; i < len(b); i++ {
		if _, e := Open(b[:i]); e == nil {
			t.Fatalf("accepted truncation %d", i)
		}
	}
	// Well-formed semantic mutants must open, execute, and fail equality.
	for _, column := range []int{0, 1, 2, 3, 4, 5, 8, 9, 10, 11} {
		x := bytes.Clone(b)
		node := 0
		if column == 9 {
			node = 1
		}
		if column == 10 || column == 11 {
			node = 2
		}
		off := headerSize + (column*3+node)*4
		if column == 0 || column >= 8 {
			binary.LittleEndian.PutUint32(x[off:], 0)
		} else if column == 3 {
			binary.LittleEndian.PutUint32(x[off:], 32)
		} else if column == 2 {
			binary.LittleEndian.PutUint32(x[off:], 2)
		} else {
			binary.LittleEndian.PutUint32(x[off:], 1)
		}
		m, e := Open(x)
		if e != nil {
			t.Fatalf("semantic mutant %d invalid: %v", column, e)
		}
		if bytes.Equal(snapshot(restore(m), 0), snapshot(tree, 0)) {
			t.Fatalf("semantic mutant column %d survived", column)
		}
	}
}
func BenchmarkWalk(b *testing.B) {
	tree := sample()
	data, e := Encode(tree)
	if e != nil {
		b.Fatal(e)
	}
	r, e := Open(data)
	if e != nil {
		b.Fatal(e)
	}
	var walk func(int) uint64
	walk = func(i int) uint64 {
		n := tree.Nodes[i]
		sum := uint64(n.Pos+n.End+n.Flags) + 1
		if n.Trailing {
			sum += 1 << 30
		}
		if n.MultiLine {
			sum += 1 << 29
		}
		for _, c := range n.Children {
			sum += walk(c)
		}
		return sum
	}
	want := walk(tree.Root)
	if r.Walk(r.Root()) != want {
		b.Fatal("walk checksum differs")
	}
	b.Run("current-snapshot", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			if walk(tree.Root) != want {
				b.Fatal("checksum")
			}
		}
	})
	b.Run("flat-columns", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			if r.Walk(r.Root()) != want {
				b.Fatal("checksum")
			}
		}
	})
}

func TestFixtureProvenance(t *testing.T) {
	t.Parallel()
	for _, family := range []string{"public", "upstream", "typescript", "recovery"} {
		b, e := os.ReadFile("testdata/" + family + "/sources.json")
		if e != nil {
			t.Fatal(e)
		}
		var rows []struct {
			Fixture string `json:"fixture"`
			Hash    string `json:"sha256"`
			Status  string `json:"status"`
		}
		if e := json.Unmarshal(b, &rows); e != nil {
			t.Fatal(e)
		}
		if (family == "public" || family == "upstream") && len(rows) != 23 {
			t.Fatalf("%s rows=%d", family, len(rows))
		}
		for _, row := range rows {
			source, e := os.ReadFile("testdata/" + family + "/" + row.Fixture)
			if e != nil {
				t.Fatal(e)
			}
			if fmt.Sprintf("%x", sha256.Sum256(source)) != row.Hash {
				t.Fatalf("provenance changed: %s/%s", family, row.Fixture)
			}
			if family == "public" && row.Status != "fetched" {
				t.Fatal("unfetched public pin")
			}
		}
	}
}

func TestRecoveryCorpus(t *testing.T) {
	t.Parallel()
	b, e := os.ReadFile("testdata/recovery/answers.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Fixture string `json:"fixture"`
		Go      string `json:"go_sha256"`
		Node    string `json:"node_sha256"`
		Agree   bool   `json:"agree"`
	}
	if e := json.Unmarshal(b, &rows); e != nil {
		t.Fatal(e)
	}
	if len(rows) != 4 {
		t.Fatal("recovery inventory changed")
	}
	binary := oracle(t)
	runner, _ := filepath.Abs(filepath.Join(repo, "oracle/node.mjs"))
	driver, _ := filepath.Abs("../parser/main.ts")
	transport, _ := filepath.Abs("snapshot.ts")
	for _, row := range rows {
		path, _ := filepath.Abs("testdata/recovery/" + row.Fixture)
		goAnswer := run(t, "", binary, path, "--whole", "--recovery")
		nodeAnswer := run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, driver, path, "--whole", "--recovery")
		if fmt.Sprintf("%x", sha256.Sum256(goAnswer)) != row.Go || fmt.Sprintf("%x", sha256.Sum256(nodeAnswer)) != row.Node || bytes.Equal(goAnswer, nodeAnswer) != row.Agree {
			t.Fatalf("recovery boundary changed: %s", row.Fixture)
		}
		manifest := filepath.Join(t.TempDir(), "manifest")
		if e := os.WriteFile(manifest, []byte(path+"\n"), 0644); e != nil {
			t.Fatal(e)
		}
		trees := snapshots(t, run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, transport, manifest))
		data, e := Encode(trees[0])
		if e != nil {
			t.Fatal(e)
		}
		r, e := Open(data)
		if e != nil {
			t.Fatal(e)
		}
		if !bytes.Equal(snapshot(restore(r), 0), snapshot(trees[0], 0)) {
			t.Fatal("recovery flat bytes differ")
		}
		t.Logf("recovery fixture=%s Go/Node-agree=%t flat-roundtrip=identical", row.Fixture, row.Agree)
	}
}
