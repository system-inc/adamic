package lint

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/system-inc/adamic/internal/buildcache"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
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
	started := time.Now()
	bundle, _ := yepestaFetch(t)
	oracle := bundle.Oracle
	rows := bundle.Rows
	slices := jsxTextnodesPartition(t, rows)
	var selected []string
	for _, i := range slices[shard] {
		selected = append(selected, rows[i])
	}
	port, js, binary := bundle.Port, bundle.JS, bundle.Binary
	t.Logf("setup %.3fs; shard-%03d cases=%d", time.Since(started).Seconds(), shard, len(selected))
	path := manifest(t, selected)
	want := yepestaExecute(t, "", oracle, "--manifest", path).output
	mutated := yepestaJavaScript(t, filepath.Join(port, "main.ts"), path, false).output
	emitted := yepestaJavaScript(t, js, path, false).output
	if diff := difference(emitted, mutated); diff != "" {
		t.Fatalf("emitted mutant differs from Node: %s", diff)
	}
	got := yepestaExecute(t, "", binary, "--manifest", path).output
	if diff := difference(got, mutated); diff != "" {
		t.Fatalf("sanitized native canary differs from mutated Node: %s", diff)
	}
	t.Logf("shard-%03d %.3fs; mutant detected=%t", shard, time.Since(started).Seconds(), !bytes.Equal(want, mutated))
}

func TestMutantsReactJsxNoCommentTextnodesUnion(t *testing.T) {
	t.Parallel()
	bundle, _ := yepestaFetch(t)
	rows := bundle.Rows
	live := jsxTextnodesRows(t, bundle.Oracle)
	counts := map[string]int{}
	for _, row := range live {
		counts[jsxTextnodesKey(t, row)]++
	}
	for _, row := range rows {
		counts[jsxTextnodesKey(t, row)]--
	}
	for key, count := range counts {
		if count != 0 {
			t.Fatalf("live enumeration differs for %q: %d", key, count)
		}
	}
	slices := jsxTextnodesPartition(t, rows)
	if len(slices) != testMutantsReactJsxNoCommentTextnodesShards {
		t.Fatal("shard enumeration mismatch")
	}
	// Check the gate-visible top-level enumeration, not merely the array size.
	file, err := parser.ParseFile(token.NewFileSet(), "yepesta_shards_test.go", nil, 0)
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
		if !ok || suffix == "Setup" {
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
}

func TestMutantsReactJsxNoCommentTextnodesMutantKilled(t *testing.T) {
	t.Parallel()
	bundle, _ := yepestaFetch(t)
	oracle := bundle.Oracle
	// Preserve the original existential mutant-must-fail checks, with the owned
	// witnesses alone: these must kill both Node and emitted JavaScript mutants.
	witnesses := bundle.Witnesses
	path := manifest(t, witnesses)
	want := yepestaExecute(t, "", oracle, "--manifest", path).output
	port, js := bundle.Port, bundle.JS
	for _, side := range []struct {
		name   string
		output []byte
	}{{"Node", yepestaJavaScript(t, filepath.Join(port, "main.ts"), path, false).output}, {"emitted JavaScript", yepestaJavaScript(t, js, path, false).output}} {
		if bytes.Equal(side.output, want) {
			t.Fatalf("JSX textnode mutant survived on %s", side.name)
		}
		t.Logf("mutant caught on %s: %s", side.name, difference(side.output, want))
	}
}
func TestMutantsReactJsxNoCommentTextnodesPlantedFailure(t *testing.T) {
	t.Parallel()
	bundle, _ := yepestaFetch(t)
	rows := bundle.Rows
	slices := jsxTextnodesPartition(t, rows)
	if len(rows) == 0 {
		t.Fatal("empty corpus")
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
		path := manifest(t, selected)
		got := yepestaExecute(t, "", bundle.Oracle, "--manifest", path).output
		want := append([]byte(nil), got...)
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
func TestMutantsReactJsxNoCommentTextnodes_000(t *testing.T) { t.Parallel(); jsxTextnodesShard(t, 0) }
func TestMutantsReactJsxNoCommentTextnodes_001(t *testing.T) { t.Parallel(); jsxTextnodesShard(t, 1) }
func TestMutantsReactJsxNoCommentTextnodes_002(t *testing.T) { t.Parallel(); jsxTextnodesShard(t, 2) }
func TestMutantsReactJsxNoCommentTextnodes_003(t *testing.T) { t.Parallel(); jsxTextnodesShard(t, 3) }
func TestMutantsReactJsxNoCommentTextnodes_004(t *testing.T) { t.Parallel(); jsxTextnodesShard(t, 4) }
func TestMutantsReactJsxNoCommentTextnodes_005(t *testing.T) { t.Parallel(); jsxTextnodesShard(t, 5) }
func TestMutantsReactJsxNoCommentTextnodes_006(t *testing.T) { t.Parallel(); jsxTextnodesShard(t, 6) }
func TestMutantsReactJsxNoCommentTextnodes_007(t *testing.T) { t.Parallel(); jsxTextnodesShard(t, 7) }
func TestMutantsReactJsxNoCommentTextnodes_008(t *testing.T) { t.Parallel(); jsxTextnodesShard(t, 8) }
func TestMutantsReactJsxNoCommentTextnodes_009(t *testing.T) { t.Parallel(); jsxTextnodesShard(t, 9) }
func TestMutantsReactJsxNoCommentTextnodes_010(t *testing.T) { t.Parallel(); jsxTextnodesShard(t, 10) }
func TestMutantsReactJsxNoCommentTextnodes_011(t *testing.T) { t.Parallel(); jsxTextnodesShard(t, 11) }
func TestMutantsReactJsxNoCommentTextnodes_012(t *testing.T) { t.Parallel(); jsxTextnodesShard(t, 12) }
func TestMutantsReactJsxNoCommentTextnodes_013(t *testing.T) { t.Parallel(); jsxTextnodesShard(t, 13) }
func TestMutantsReactJsxNoCommentTextnodes_014(t *testing.T) { t.Parallel(); jsxTextnodesShard(t, 14) }
func TestMutantsReactJsxNoCommentTextnodes_015(t *testing.T) { t.Parallel(); jsxTextnodesShard(t, 15) }

// Not parallel: builds the shared Go oracle before immutable fixtures are published.
func TestMutantsReactJsxNoCommentTextnodesOracleSetup(t *testing.T) {
	started := time.Now()
	yepestaOracle(t, true)
	if time.Since(started) >= 60*time.Second {
		t.Fatal("cooked: oracle setup exceeds 60s")
	}
}

// Not parallel: lowers the shared mutant before its native product is built.
func TestMutantsReactJsxNoCommentTextnodesLoweredSetup(t *testing.T) {
	started := time.Now()
	jsxTextnodesProducts(t, 1)
	if time.Since(started) >= 60*time.Second {
		t.Fatal("cooked: lowered setup exceeds 60s")
	}
}

// Not parallel: builds the sanitized native product before fixture publication.
func TestMutantsReactJsxNoCommentTextnodesNativeSetup(t *testing.T) {
	started := time.Now()
	jsxTextnodesProducts(t, 2)
	if time.Since(started) >= 60*time.Second {
		t.Fatal("cooked: native setup exceeds 60s")
	}
}
func yepestaOracle(t *testing.T, build bool) string {
	t.Helper()
	inputs := yepestaInputs()
	inputs.Name = "yepesta-oracle-v1"
	directory := buildcache.Product(t, inputs, func(out string) error {
		if !build {
			return fmt.Errorf("run TestMutantsReactJsxNoCommentTextnodesOracleSetup first")
		}
		_, err := goOracleIn(".", out)
		return err
	})
	return filepath.Join(directory, "oracle")
}

// Not parallel: publishes immutable products before parallel shard execution.
func TestMutantsReactJsxNoCommentTextnodes_Setup(t *testing.T) {
	started := time.Now()
	directory := buildcache.Product(t, yepestaInputs(), func(out string) error {
		oracle := yepestaOracle(t, false)
		rows := jsxTextnodesRows(t, oracle)
		witnesses := recoveryRows(t, oracle, ownedWitnessRows(t, ".", jsxTextnodesDescriptor(t)))
		port, js, binary := jsxTextnodesProducts(t, 0)
		bundle := yepestaBundle{Oracle: filepath.Join(out, "oracle"), Port: port, JS: js, Binary: binary}
		data, err := os.ReadFile(oracle)
		if err != nil {
			return err
		}
		if err := os.WriteFile(bundle.Oracle, data, 0755); err != nil {
			return err
		}
		persist := func(rows []string, prefix string) ([]string, error) {
			var result []string
			for i, row := range rows {
				fields := strings.Split(row, "\t")
				data, err := os.ReadFile(fields[0])
				if err != nil {
					return nil, err
				}
				name := filepath.Join(prefix, fmt.Sprintf("%03d", i), filepath.Base(fields[0]))
				if err := os.MkdirAll(filepath.Dir(filepath.Join(out, name)), 0755); err != nil {
					return nil, err
				}
				if err := os.WriteFile(filepath.Join(out, name), data, 0644); err != nil {
					return nil, err
				}
				fields[0] = name
				result = append(result, strings.Join(fields, "\t"))
			}
			return result, nil
		}
		bundle.Rows, err = persist(rows, "rows")
		if err != nil {
			return err
		}
		bundle.Witnesses, err = persist(witnesses, "witnesses")
		if err != nil {
			return err
		}
		bundle.Oracle = "oracle"
		data, err = json.Marshal(bundle)
		if err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(out, "bundle.json"), data, 0644)
	})
	t.Logf("setup product %s: %.3fs", directory, time.Since(started).Seconds())
	if time.Since(started) >= 60*time.Second {
		t.Fatal("cooked: setup exceeds 60s")
	}
}

type yepestaBundle struct {
	Oracle, Port, JS, Binary string
	Rows, Witnesses          []string
}

func yepestaInputs() buildcache.Inputs {
	return buildcache.Inputs{Name: "yepesta-setup-v1", Files: []string{"go.mod", "cohere", "internal", "stage1/typescript", "oracle", "stage1/cohere/lint"}, Flags: []string{packageDirectory, "shards=16", os.Getenv("ADAMIC_NATIVE_SPLIT"), os.Getenv("ADAMIC_NATIVE_JOBS")}, Toolchain: []string{buildcache.Tool("go", "version"), buildcache.Tool("clang", "--version")}}
}
func yepestaFetch(t *testing.T) (yepestaBundle, string) {
	t.Helper()
	directory := buildcache.Product(t, yepestaInputs(), func(string) error {
		return fmt.Errorf("run TestMutantsReactJsxNoCommentTextnodes_Setup first; shards never build")
	})
	data, err := os.ReadFile(filepath.Join(directory, "bundle.json"))
	if err != nil {
		t.Fatal(err)
	}
	var bundle yepestaBundle
	if err := json.Unmarshal(data, &bundle); err != nil {
		t.Fatal(err)
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

func yepestaExecute(t *testing.T, directory, name string, args ...string) execution {
	t.Helper()
	deadline := time.Now().Add(90 * time.Second)
	if end, ok := t.Deadline(); ok && end.Before(deadline) {
		deadline = end.Add(-time.Second)
	}
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = directory
	out, err := os.CreateTemp(t.TempDir(), "stdout-")
	if err != nil {
		t.Fatal(err)
	}
	defer out.Close()
	cmd.Stdout = out
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	started := time.Now()
	err = cmd.Run()
	if err != nil || len(commandDiagnostics(name, stderr.Bytes())) != 0 {
		t.Fatalf("%s %v: %v\n%s", name, args, err, &stderr)
	}
	data, err := os.ReadFile(out.Name())
	if err != nil {
		t.Fatal(err)
	}
	return execution{data, time.Since(started)}
}
func yepestaJavaScript(t *testing.T, module, path string, count bool) execution {
	t.Helper()
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	args := []string{"--disable-warning=ExperimentalWarning", runner, module, "--manifest", path}
	if count {
		args = append(args, "--count")
	}
	return yepestaExecute(t, "", "node", args...)
}
