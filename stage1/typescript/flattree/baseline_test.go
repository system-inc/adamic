package flattree

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// This builds only the unchanged parser and scout transport. Typed arrays are
// deliberately not represented by ordinary arrays in the native probe.
func TestNativeBaselineAndTypedArrayGap(t *testing.T) {
	t.Parallel()
	driver, _ := filepath.Abs("snapshot.ts")
	program, e := load.Load([]string{driver})
	if e != nil {
		t.Fatal(e)
	}
	ir, e := lower.Lower(context.Background(), program)
	if e != nil {
		t.Fatal(e)
	}
	c := native.C(ir)
	d := t.TempDir()
	binary := filepath.Join(d, "snapshot")
	if e := native.Build(c, binary, native.Options{Count: true, Sanitize: true}); e != nil {
		t.Fatal(e)
	}
	path := filepath.Join(d, "source.ts")
	if e := os.WriteFile(path, []byte("x;"), 0644); e != nil {
		t.Fatal(e)
	}
	manifest := filepath.Join(d, "manifest")
	if e := os.WriteFile(manifest, []byte(path+"\n"), 0644); e != nil {
		t.Fatal(e)
	}
	runner, _ := filepath.Abs(filepath.Join(repo, "oracle/node.mjs"))
	want := run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, driver, manifest)
	observe := func(args ...string) ([]byte, string) {
		cmd := exec.Command(binary, args...)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		if e := cmd.Run(); e != nil {
			t.Fatalf("native: %v\n%s", e, &stderr)
		}
		if strings.Contains(stderr.String(), "Sanitizer") {
			t.Fatal(stderr.String())
		}
		return stdout.Bytes(), stderr.String()
	}
	got, counts := observe(manifest)
	if !bytes.Equal(got, want) {
		t.Fatal("native snapshot differs from Node")
	}
	t.Logf("native x; export counts: %s", strings.TrimSpace(counts))
	for _, passes := range []int{0, 1, 10000} {
		got, counts := observe(manifest, strconv.Itoa(passes), "--walk")
		expected := run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, driver, manifest, strconv.Itoa(passes), "--walk")
		if !bytes.Equal(got, expected) {
			t.Fatalf("walk passes=%d native=%q Node=%q", passes, got, expected)
		}
		t.Logf("native x; passes=%d count=%s costs=%s", passes, strings.TrimSpace(string(got)), strings.TrimSpace(counts))
	}
	compiler := os.Getenv("ADAMIC_TYPESCRIPT_SOURCE")
	if compiler == "" {
		t.Fatal("ADAMIC_TYPESCRIPT_SOURCE is required")
	}
	largeManifest := filepath.Join(d, "compiler-manifest")
	if e := os.WriteFile(largeManifest, []byte(filepath.Join(compiler, "src/compiler/parser.ts")+"\n"), 0644); e != nil {
		t.Fatal(e)
	}
	for _, passes := range []int{0, 1, 10} {
		got, counts := observe(largeManifest, strconv.Itoa(passes), "--walk")
		expected := run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, driver, largeManifest, strconv.Itoa(passes), "--walk")
		if !bytes.Equal(got, expected) {
			t.Fatal("compiler walk counts differ")
		}
		t.Logf("native compiler/parser.ts passes=%d count=%s costs=%s", passes, strings.TrimSpace(string(got)), strings.TrimSpace(counts))
	}
	gap, _ := filepath.Abs("gaps/typed_arrays.ts")
	got = run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, gap)
	if string(got) != "0\n" {
		t.Fatalf("Node gap=%q", got)
	}
	p, e := load.Load([]string{gap})
	if e != nil {
		t.Fatal(e)
	}
	_, e = lower.Lower(context.Background(), p)
	var notYet *lower.NotYet
	if !errors.As(e, &notYet) || notYet.What != "new an Identifier" {
		t.Fatalf("typed-array gap changed: %v", e)
	}
	t.Logf("Node Uint32Array=0; stage 0: %v", e)
}

func TestReaderSemanticMutantsOnNode(t *testing.T) {
	t.Parallel()
	tree := sample()
	data, e := Encode(tree)
	if e != nil {
		t.Fatal(e)
	}
	want := snapshot(tree, 0)
	d := t.TempDir()
	path := filepath.Join(d, "tree.flat")
	for _, column := range []int{0, 1, 2, 3, 4, 5, 8, 9, 10, 11} {
		x := bytes.Clone(data)
		node := 0
		if column == 9 {
			node = 1
		}
		if column >= 10 {
			node = 2
		}
		off := headerSize + (column*3+node)*4
		value := uint32(1)
		if column == 0 || column >= 8 {
			value = 0
		}
		if column == 2 {
			value = 2
		}
		if column == 3 {
			value = 32
		}
		binaryPut(x[off:], value)
		if _, e := Open(x); e != nil {
			t.Fatal(e)
		}
		if e := os.WriteFile(path, x, 0644); e != nil {
			t.Fatal(e)
		}
		got := run(t, "", "node", "reference.mjs", path)
		if bytes.Equal(got, want) {
			t.Fatalf("Node semantic mutant column %d survived", column)
		}
	}
}
func binaryPut(b []byte, v uint32) {
	b[0] = byte(v)
	b[1] = byte(v >> 8)
	b[2] = byte(v >> 16)
	b[3] = byte(v >> 24)
}

func TestSharedParserBoundary(t *testing.T) {
	t.Parallel()
	path, _ := filepath.Abs("testdata/blocked-parser/arrow.ts.txt")
	runner, _ := filepath.Abs(filepath.Join(repo, "oracle/node.mjs"))
	driver, _ := filepath.Abs("../parser/main.ts")
	want := run(t, "", oracle(t), path, "--whole")
	cmd := exec.Command("node", "--disable-warning=ExperimentalWarning", runner, driver, path, "--whole")
	output, e := cmd.CombinedOutput()
	if e == nil || !strings.Contains(string(output), "source has parse diagnostics") {
		t.Fatalf("clean-source boundary changed: err=%v output=%s", e, output)
	}
	got := run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, driver, path, "--whole", "--recovery")
	if bytes.Equal(got, want) || !bytes.Contains(got, []byte("diagnostic ")) {
		t.Fatal("recovered boundary disappeared")
	}
	t.Logf("shortest measured witness %q: Go clean; Node false diagnostics and different tree", "let a=({b:c=>c});")
	manifest := filepath.Join(t.TempDir(), "manifest")
	if e := os.WriteFile(manifest, []byte(path+"\n"), 0644); e != nil {
		t.Fatal(e)
	}
	transport, _ := filepath.Abs("snapshot.ts")
	trees := snapshots(t, run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, transport, manifest))
	b, e := Encode(trees[0])
	if e != nil {
		t.Fatal(e)
	}
	r, e := Open(b)
	if e != nil {
		t.Fatal(e)
	}
	if !bytes.Equal(snapshot(restore(r), 0), snapshot(trees[0], 0)) {
		t.Fatal("recovered tree round-trip differs")
	}
}

func TestFlagAndChildOrderMutants(t *testing.T) {
	t.Parallel()
	tree := sample()
	data, e := Encode(tree)
	if e != nil {
		t.Fatal(e)
	}
	r, e := Open(data)
	if e != nil {
		t.Fatal(e)
	}
	want := snapshot(tree, 0)
	path := filepath.Join(t.TempDir(), "tree.flat")
	for _, name := range []string{"optional", "trailing", "multiline", "child-order", "child-count"} {
		x := bytes.Clone(data)
		switch name {
		case "optional", "trailing", "multiline":
			mask := uint32(32)
			if name == "trailing" {
				mask = 1 << 30
			}
			if name == "multiline" {
				mask = 1 << 29
			}
			binaryPut(x[headerSize+(3*3+2)*4:], r.Value(3, 2)^mask)
		case "child-order":
			binaryPut(x[r.childOff:], 1)
			binaryPut(x[r.childOff+4:], 0)
		case "child-count":
			binaryPut(x[headerSize+(7*3+2)*4:], 1)
		}
		m, e := Open(x)
		if e != nil {
			t.Fatal(e)
		}
		if bytes.Equal(snapshot(restore(m), 0), want) {
			t.Fatalf("Go %s mutant survived", name)
		}
		if e := os.WriteFile(path, x, 0644); e != nil {
			t.Fatal(e)
		}
		if bytes.Equal(run(t, "", "node", "reference.mjs", path), want) {
			t.Fatalf("Node %s mutant survived", name)
		}
	}
}

func TestReducedRecoveryAnswers(t *testing.T) {
	t.Parallel()
	oracleBinary := oracle(t)
	runner, _ := filepath.Abs(filepath.Join(repo, "oracle/node.mjs"))
	driver, _ := filepath.Abs("../parser/main.ts")
	for _, id := range []string{"00", "01", "02"} {
		path, _ := filepath.Abs("testdata/reduced-recovery/" + id + ".ts.txt")
		goAnswer := run(t, "", oracleBinary, path, "--whole", "--recovery")
		nodeAnswer := run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, driver, path, "--whole", "--recovery")
		goWant, e := os.ReadFile("testdata/reduced-recovery/" + id + ".go.txt")
		if e != nil {
			t.Fatal(e)
		}
		nodeWant, e := os.ReadFile("testdata/reduced-recovery/" + id + ".node.txt")
		if e != nil {
			t.Fatal(e)
		}
		if !bytes.Equal(goAnswer, goWant) || !bytes.Equal(nodeAnswer, nodeWant) || bytes.Equal(goAnswer, nodeAnswer) {
			t.Fatalf("reduced parser boundary %s changed", id)
		}
	}
}
