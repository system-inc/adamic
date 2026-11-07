package lint

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"

	"github.com/system-inc/adamic/stage1/cohere/lint/oracles"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/system-inc/adamic/stage1/cohere/lint/registry"
)

const testMutantsReactJsxNoCommentTextnodesShards = 16

func jsxTextnodesDescriptor(test *testing.T) registry.Descriptor {
	test.Helper()
	for _, d := range prepareRegistry(test, ".") {
		if d.Name == nativeCanaryRule {
			return d
		}
	}
	test.Fatal("missing JSX textnode canary")
	return registry.Descriptor{}
}

// A generated file's temporary directory is not a stable key. Its contents,
// basename and complete options are; distinct rows with the same key remain
// separate occurrences in the live enumeration.
func jsxTextnodesKey(test *testing.T, row string) string {
	test.Helper()
	fields := strings.SplitN(row, "\t", 2)
	data, err := os.ReadFile(fields[0])
	if err != nil {
		test.Fatal(err)
	}
	options := ""
	if len(fields) == 2 {
		options = fields[1]
	}
	return fmt.Sprintf("%s\x00%x\x00%s", filepath.Base(fields[0]), sha256.Sum256(data), options)
}
func jsxTextnodesPartition(test *testing.T, rows []string) [testMutantsReactJsxNoCommentTextnodesShards][]int {
	test.Helper()
	var slices [testMutantsReactJsxNoCommentTextnodesShards][]int
	for i, row := range rows {
		sum := sha256.Sum256([]byte(jsxTextnodesKey(test, row)))
		shard := int(binary.LittleEndian.Uint64(sum[:8]) % testMutantsReactJsxNoCommentTextnodesShards)
		slices[shard] = append(slices[shard], i)
	}
	return slices
}
func jsxTextnodesRows(test *testing.T, oracle string) []string {
	test.Helper()
	d := jsxTextnodesDescriptor(test)
	var rows []string
	for _, row := range generated(test) {
		selected := "all"
		fields := strings.Split(row, "\t")
		if len(fields) > 1 && fields[1] != "" {
			selected = fields[1]
		}
		if selected == "all" || selected == d.Name {
			rows = append(rows, row)
		}
	}
	return append(rows, recoveryRows(test, oracle, ownedWitnessRows(test, ".", d))...)
}

func jsxTextnodesShard(test *testing.T, shard int) {
	test.Helper()
	bundle, _ := yepestaFetch(test)
	started := time.Now()
	oracle := bundle.Oracle
	rows := bundle.Rows
	slices := jsxTextnodesPartition(test, rows)
	var selected []string
	for _, i := range slices[shard] {
		selected = append(selected, rows[i])
	}
	port, js, binary := jsxTextnodesProducts(test)
	test.Logf("shard-%03d cases=%d", shard, len(selected))
	path := manifest(test, selected)
	want := yepestaExecute(test, "", oracle, "--manifest", path).output
	mutated := yepestaJavaScript(test, filepath.Join(port, "main.ts"), path, false).output
	emitted := yepestaJavaScript(test, js, path, false).output
	if diff := difference(emitted, mutated); diff != "" {
		test.Fatalf("emitted mutant differs from Node: %s", diff)
	}
	got := yepestaExecute(test, "", binary, "--manifest", path).output
	if diff := difference(got, mutated); diff != "" {
		test.Fatalf("sanitized native canary differs from mutated Node: %s", diff)
	}
	test.Logf("shard-%03d %.3fs; mutant detected=%t", shard, time.Since(started).Seconds(), !bytes.Equal(want, mutated))
}

func TestMutantsReactJsxNoCommentTextnodesUnion(test *testing.T) {
	test.Parallel()
	bundle, _ := yepestaFetch(test)
	rows := bundle.Rows
	live := jsxTextnodesRows(test, bundle.Oracle)
	counts := map[string]int{}
	for _, row := range live {
		counts[jsxTextnodesKey(test, row)]++
	}
	for _, row := range rows {
		counts[jsxTextnodesKey(test, row)]--
	}
	for key, count := range counts {
		if count != 0 {
			test.Fatalf("live enumeration differs for %q: %d", key, count)
		}
	}
	slices := jsxTextnodesPartition(test, rows)
	if len(slices) != testMutantsReactJsxNoCommentTextnodesShards {
		test.Fatal("shard enumeration mismatch")
	}
	// Check the gate-visible top-level enumeration, not merely the array size.
	file, err := parser.ParseFile(token.NewFileSet(), "yepesta_independent_shards_test.go", nil, 0)
	if err != nil {
		test.Fatal(err)
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
			test.Fatalf("invalid shard function %s", fn.Name.Name)
		}
		functions[i] = true
	}
	if len(functions) != testMutantsReactJsxNoCommentTextnodesShards {
		test.Fatalf(
			"enumerated %d top-level shards, want %d",
			len(functions),
			testMutantsReactJsxNoCommentTextnodesShards,
		)
	}
	seen := make([]int, len(rows))
	for _, slice := range slices {
		for _, i := range slice {
			seen[i]++
		}
	}
	for i, count := range seen {
		if count != 1 {
			test.Fatalf("case %d covered %d times", i, count)
		}
	}
	test.Logf("union: %d cases exactly once across %d shards", len(rows), len(slices))
}

func TestMutantsReactJsxNoCommentTextnodesMutantKilled(test *testing.T) {
	test.Parallel()
	bundle, _ := yepestaFetch(test)
	oracle := bundle.Oracle
	// Preserve the original existential mutant-must-fail checks, with the owned
	// witnesses alone: these must kill both Node and emitted JavaScript mutants.
	witnesses := bundle.Witnesses
	path := manifest(test, witnesses)
	want := yepestaExecute(test, "", oracle, "--manifest", path).output
	port, js, _ := jsxTextnodesProducts(test)
	for _, side := range []struct {
		name   string
		output []byte
	}{
		{"Node", yepestaJavaScript(test, filepath.Join(port, "main.ts"), path, false).output},
		{"emitted JavaScript", yepestaJavaScript(test, js, path, false).output},
	} {
		if bytes.Equal(side.output, want) {
			test.Fatalf("JSX textnode mutant survived on %s", side.name)
		}
		test.Logf("mutant caught on %s: %s", side.name, difference(side.output, want))
	}
}
func TestMutantsReactJsxNoCommentTextnodesPlantedFailure(test *testing.T) {
	test.Parallel()
	bundle, _ := yepestaFetch(test)
	rows := bundle.Rows
	slices := jsxTextnodesPartition(test, rows)
	if len(rows) == 0 {
		test.Fatal("empty corpus")
	}
	planted := len(rows) / 2
	caught := 0
	owner := -1
	for shard, slice := range slices {
		var selected []string
		plantPosition := -1
		for position, i := range slice {
			selected = append(selected, rows[i])
			if i == planted {
				plantPosition = position
			}
		}
		path := manifest(test, selected)
		got := yepestaExecute(test, "", bundle.Oracle, "--manifest", path).output
		want := append([]byte(nil), got...)
		// Insert one disagreeing finding into exactly the planted case's record.
		// The real leaves' byte comparator must catch only its owning shard.
		if plantPosition >= 0 {
			marker := []byte(fmt.Sprintf("case %d\n", plantPosition))
			if bytes.Count(want, marker) != 1 {
				test.Fatalf("planted case header missing or repeated: %q", marker)
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
		test.Fatalf("planted disagreement caught by %d shards", caught)
	}
	test.Logf("planted case %d caught only by shard-%03d", planted, owner)
}
func TestMutantsReactJsxNoCommentTextnodes_000(test *testing.T) {
	test.Parallel()
	jsxTextnodesShard(test, 0)
}
func TestMutantsReactJsxNoCommentTextnodes_001(test *testing.T) {
	test.Parallel()
	jsxTextnodesShard(test, 1)
}
func TestMutantsReactJsxNoCommentTextnodes_002(test *testing.T) {
	test.Parallel()
	jsxTextnodesShard(test, 2)
}
func TestMutantsReactJsxNoCommentTextnodes_003(test *testing.T) {
	test.Parallel()
	jsxTextnodesShard(test, 3)
}
func TestMutantsReactJsxNoCommentTextnodes_004(test *testing.T) {
	test.Parallel()
	jsxTextnodesShard(test, 4)
}
func TestMutantsReactJsxNoCommentTextnodes_005(test *testing.T) {
	test.Parallel()
	jsxTextnodesShard(test, 5)
}
func TestMutantsReactJsxNoCommentTextnodes_006(test *testing.T) {
	test.Parallel()
	jsxTextnodesShard(test, 6)
}
func TestMutantsReactJsxNoCommentTextnodes_007(test *testing.T) {
	test.Parallel()
	jsxTextnodesShard(test, 7)
}
func TestMutantsReactJsxNoCommentTextnodes_008(test *testing.T) {
	test.Parallel()
	jsxTextnodesShard(test, 8)
}
func TestMutantsReactJsxNoCommentTextnodes_009(test *testing.T) {
	test.Parallel()
	jsxTextnodesShard(test, 9)
}
func TestMutantsReactJsxNoCommentTextnodes_010(test *testing.T) {
	test.Parallel()
	jsxTextnodesShard(test, 10)
}
func TestMutantsReactJsxNoCommentTextnodes_011(test *testing.T) {
	test.Parallel()
	jsxTextnodesShard(test, 11)
}
func TestMutantsReactJsxNoCommentTextnodes_012(test *testing.T) {
	test.Parallel()
	jsxTextnodesShard(test, 12)
}
func TestMutantsReactJsxNoCommentTextnodes_013(test *testing.T) {
	test.Parallel()
	jsxTextnodesShard(test, 13)
}
func TestMutantsReactJsxNoCommentTextnodes_014(test *testing.T) {
	test.Parallel()
	jsxTextnodesShard(test, 14)
}
func TestMutantsReactJsxNoCommentTextnodes_015(test *testing.T) {
	test.Parallel()
	jsxTextnodesShard(test, 15)
}

func yepestaPrepare(test *testing.T) string { test.Helper(); return oracles.YepestaSetup(test) }

// yepestaBundle is the shards' shared preparation, every path in it relative to its product's directory: the Go
// oracle and the rows and witnesses it judges. The mutated port, its JavaScript and its native build are products of
// their own (jsxTextnodesProducts), so the bundle names no other product's place in any one machine's cache.
type yepestaBundle struct {
	Oracle          string
	Rows, Witnesses []string
}

// Preparation runs once per process, including when a single shard is selected.
// Product supplies content-addressed reuse across processes: Workshop builds it once
// (TestProduct_LintFindsYepestaSetup), and a runner reading with ADAMIC_BUILD_CACHE=read fails
// naming it when it is missing rather than building it.
var yepestaPrepared struct {
	once      sync.Once
	directory string
}

func yepestaFetch(test *testing.T) (yepestaBundle, string) {
	test.Helper()
	yepestaPrepared.once.Do(func() { yepestaPrepared.directory = yepestaPrepare(test) })
	directory := yepestaPrepared.directory
	if directory == "" {
		test.Fatal("shared preparation failed")
	}
	data, err := os.ReadFile(filepath.Join(directory, "bundle.json"))
	if err != nil {
		test.Fatal(err)
	}
	var bundle yepestaBundle
	if err := json.Unmarshal(data, &bundle); err != nil {
		test.Fatal(err)
	}
	bundle.Oracle = filepath.Join(directory, bundle.Oracle)
	for _, rows := range [][]string{bundle.Rows, bundle.Witnesses} {
		for i, row := range rows {
			fields := strings.Split(row, "\t")
			fields[0] = filepath.Join(directory, fields[0])
			rows[i] = strings.Join(fields, "\t")
		}
	}
	return bundle, directory
}

func yepestaExecute(test *testing.T, directory, name string, args ...string) execution {
	test.Helper()
	deadline := time.Now().Add(90 * time.Second)
	if end, ok := test.Deadline(); ok && end.Before(deadline) {
		deadline = end.Add(-time.Second)
	}
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()
	cmd := lintCommand(ctx, directory, name, args...)
	out, err := os.CreateTemp(test.TempDir(), "stdout-")
	if err != nil {
		test.Fatal(err)
	}
	defer out.Close()
	cmd.Stdout = out
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	started := time.Now()
	err = cmd.Run()
	if err != nil || len(commandDiagnostics(name, stderr.Bytes())) != 0 {
		test.Fatalf("%s %v: %v\n%s", name, args, err, &stderr)
	}
	data, err := os.ReadFile(out.Name())
	if err != nil {
		test.Fatal(err)
	}
	return execution{data, time.Since(started)}
}
func yepestaJavaScript(test *testing.T, module, path string, count bool) execution {
	test.Helper()
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		test.Fatal(err)
	}
	args := []string{"--disable-warning=ExperimentalWarning", runner, module, "--manifest", path}
	if count {
		args = append(args, "--count")
	}
	return yepestaExecute(test, "", "node", args...)
}
