package lint

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/stage1/cohere/lint/registry"
)

const testMutantsReactJsxNoCommentTextnodesShards = 16

func jsxTextnodesDescriptor(t *testing.T) registry.Descriptor {
	t.Helper()
	for _, d := range prepareRegistry(t, ".") {
		if d.Name == nativeCanaryRule {
			return d
		}
	}
	t.Fatal("missing JSX textnode canary")
	return registry.Descriptor{}
}

// A generated file's temporary directory is not a stable key. Its contents,
// basename and complete options are; distinct rows with the same key remain
// separate occurrences in the live enumeration.
func jsxTextnodesKey(t *testing.T, row string) string {
	t.Helper()
	fields := strings.SplitN(row, "\t", 2)
	data, err := os.ReadFile(fields[0])
	if err != nil {
		t.Fatal(err)
	}
	options := ""
	if len(fields) == 2 {
		options = fields[1]
	}
	return fmt.Sprintf("%s\x00%x\x00%s", filepath.Base(fields[0]), sha256.Sum256(data), options)
}
func jsxTextnodesPartition(t *testing.T, rows []string) [testMutantsReactJsxNoCommentTextnodesShards][]int {
	t.Helper()
	var slices [testMutantsReactJsxNoCommentTextnodesShards][]int
	for i, row := range rows {
		sum := sha256.Sum256([]byte(jsxTextnodesKey(t, row)))
		shard := int(binary.LittleEndian.Uint64(sum[:8]) % testMutantsReactJsxNoCommentTextnodesShards)
		slices[shard] = append(slices[shard], i)
	}
	return slices
}
func jsxTextnodesRows(t *testing.T, oracle string) []string {
	t.Helper()
	d := jsxTextnodesDescriptor(t)
	var rows []string
	for _, row := range generated(t) {
		selected := "all"
		fields := strings.Split(row, "\t")
		if len(fields) > 1 && fields[1] != "" {
			selected = fields[1]
		}
		if selected == "all" || selected == d.Name {
			rows = append(rows, row)
		}
	}
	return append(rows, recoveryRows(t, oracle, ownedWitnessRows(t, ".", d))...)
}

func jsxTextnodesShard(t *testing.T, shard int) {
	t.Helper()
	t.Parallel()
	started := time.Now()
	oracle := goOracle(t)
	rows := jsxTextnodesRows(t, oracle)
	slices := jsxTextnodesPartition(t, rows)
	var selected []string
	for _, i := range slices[shard] {
		selected = append(selected, rows[i])
	}
	port, js, binary := jsxTextnodesProducts(t)
	t.Logf("setup %.3fs; shard-%03d cases=%d", time.Since(started).Seconds(), shard, len(selected))
	path := manifest(t, selected)
	want := execute(t, "", oracle, "--manifest", path).output
	mutated := runJavaScript(t, filepath.Join(port, "main.ts"), path, false).output
	emitted := runJavaScript(t, js, path, false).output
	if diff := difference(emitted, mutated); diff != "" {
		t.Fatalf("emitted mutant differs from Node: %s", diff)
	}
	got := execute(t, "", binary, "--manifest", path).output
	if diff := difference(got, mutated); diff != "" {
		t.Fatalf("sanitized native canary differs from mutated Node: %s", diff)
	}
	t.Logf("shard-%03d %.3fs; mutant detected=%t", shard, time.Since(started).Seconds(), !bytes.Equal(want, mutated))
}

func TestMutantsReactJsxNoCommentTextnodesUnion(t *testing.T) {
	t.Parallel()
	oracle := goOracle(t)
	rows := jsxTextnodesRows(t, oracle)
	slices := jsxTextnodesPartition(t, rows)
	if len(slices) != testMutantsReactJsxNoCommentTextnodesShards {
		t.Fatal("shard enumeration mismatch")
	}
	// Check the gate-visible top-level enumeration, not merely the array size.
	file, err := parser.ParseFile(token.NewFileSet(), "mutants_jsx_textnodes_test.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	functions := map[int]bool{}
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		suffix, ok := strings.CutPrefix(fn.Name.Name, "TestMutantsReactJsxNoCommentTextnodes_")
		if !ok {
			continue
		}
		i, err := strconv.Atoi(suffix)
		if err != nil || suffix != fmt.Sprintf("%03d", i) || i < 0 || i >= len(slices) {
			t.Fatalf("invalid shard function %s", fn.Name.Name)
		}
		functions[i] = true
	}
	if len(functions) != testMutantsReactJsxNoCommentTextnodesShards {
		t.Fatalf("enumerated %d top-level shards, want %d", len(functions), testMutantsReactJsxNoCommentTextnodesShards)
	}
	seen := make([]int, len(rows))
	for _, slice := range slices {
		for _, i := range slice {
			seen[i]++
		}
	}
	for i, count := range seen {
		if count != 1 {
			t.Fatalf("case %d covered %d times", i, count)
		}
	}
	t.Logf("union: %d cases exactly once across %d shards", len(rows), len(slices))
	// Preserve the original existential mutant-must-fail checks, with the owned
	// witnesses alone: these must kill both Node and emitted JavaScript mutants.
	witnesses := recoveryRows(t, oracle, ownedWitnessRows(t, ".", jsxTextnodesDescriptor(t)))
	path := manifest(t, witnesses)
	want := execute(t, "", oracle, "--manifest", path).output
	port, js, _ := jsxTextnodesProducts(t)
	for _, side := range []struct {
		name   string
		output []byte
	}{{"Node", runJavaScript(t, filepath.Join(port, "main.ts"), path, false).output}, {"emitted JavaScript", runJavaScript(t, js, path, false).output}} {
		if bytes.Equal(side.output, want) {
			t.Fatalf("JSX textnode mutant survived on %s", side.name)
		}
		t.Logf("mutant caught on %s: %s", side.name, difference(side.output, want))
	}
}
func TestMutantsReactJsxNoCommentTextnodesPlantedFailure(t *testing.T) {
	t.Parallel()
	rows := jsxTextnodesRows(t, goOracle(t))
	slices := jsxTextnodesPartition(t, rows)
	if len(rows) == 0 {
		t.Fatal("empty corpus")
	}
	planted := len(rows) / 2
	caught := 0
	owner := -1
	port, _, binary := jsxTextnodesProducts(t)
	for shard, slice := range slices {
		var selected []string
		plantPosition := -1
		for position, i := range slice {
			selected = append(selected, rows[i])
			if i == planted {
				plantPosition = position
			}
		}
		path := manifest(t, selected)
		want := runJavaScript(t, filepath.Join(port, "main.ts"), path, false).output
		got := execute(t, "", binary, "--manifest", path).output
		if diff := difference(got, want); diff != "" {
			t.Fatalf("before planting, shard-%03d: %s", shard, diff)
		}
		// Insert one disagreeing finding into exactly the planted case's record.
		// The real leaves' byte comparator must catch only its owning shard.
		if plantPosition >= 0 {
			marker := []byte(fmt.Sprintf("case %d\n", plantPosition))
			if bytes.Count(want, marker) != 1 {
				t.Fatalf("planted case header missing or repeated: %q", marker)
			}
			replacement := append(append([]byte(nil), marker...), []byte("planted disagreement\n")...)
			want = bytes.Replace(want, marker, replacement, 1)
		}
		if difference(got, want) != "" {
			caught++
			owner = shard
		}
	}
	if caught != 1 {
		t.Fatalf("planted disagreement caught by %d shards", caught)
	}
	t.Logf("planted case %d caught only by shard-%03d", planted, owner)
}
func TestMutantsReactJsxNoCommentTextnodes_000(t *testing.T) { jsxTextnodesShard(t, 0) }
func TestMutantsReactJsxNoCommentTextnodes_001(t *testing.T) { jsxTextnodesShard(t, 1) }
func TestMutantsReactJsxNoCommentTextnodes_002(t *testing.T) { jsxTextnodesShard(t, 2) }
func TestMutantsReactJsxNoCommentTextnodes_003(t *testing.T) { jsxTextnodesShard(t, 3) }
func TestMutantsReactJsxNoCommentTextnodes_004(t *testing.T) { jsxTextnodesShard(t, 4) }
func TestMutantsReactJsxNoCommentTextnodes_005(t *testing.T) { jsxTextnodesShard(t, 5) }
func TestMutantsReactJsxNoCommentTextnodes_006(t *testing.T) { jsxTextnodesShard(t, 6) }
func TestMutantsReactJsxNoCommentTextnodes_007(t *testing.T) { jsxTextnodesShard(t, 7) }
func TestMutantsReactJsxNoCommentTextnodes_008(t *testing.T) { jsxTextnodesShard(t, 8) }
func TestMutantsReactJsxNoCommentTextnodes_009(t *testing.T) { jsxTextnodesShard(t, 9) }
func TestMutantsReactJsxNoCommentTextnodes_010(t *testing.T) { jsxTextnodesShard(t, 10) }
func TestMutantsReactJsxNoCommentTextnodes_011(t *testing.T) { jsxTextnodesShard(t, 11) }
func TestMutantsReactJsxNoCommentTextnodes_012(t *testing.T) { jsxTextnodesShard(t, 12) }
func TestMutantsReactJsxNoCommentTextnodes_013(t *testing.T) { jsxTextnodesShard(t, 13) }
func TestMutantsReactJsxNoCommentTextnodes_014(t *testing.T) { jsxTextnodesShard(t, 14) }
func TestMutantsReactJsxNoCommentTextnodes_015(t *testing.T) { jsxTextnodesShard(t, 15) }
