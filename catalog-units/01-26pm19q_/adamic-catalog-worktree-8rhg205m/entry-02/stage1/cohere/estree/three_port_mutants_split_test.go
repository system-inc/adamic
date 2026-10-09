package estree

import (
	"bytes"
	"context"
	"fmt"
	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"
)

const testThreePortMutantsShards = 12
const threePortCasesPerShard = 20

type threePortMutation struct{ name, file, from, to, witness string }

func threePortMutations() []threePortMutation {
	return []threePortMutation{
		{"member-computed", "convert.ts", "boolValue(node.kind === 'ElementAccessExpression')", "boolValue(node.kind === 'PropertyAccessExpression')", "a.b; a[b];"},
		{"logical-rebalance", "postprocess.ts", "completed.set(id, this.rebalance(id));", "completed.set(id, id);", "a && (b && (c && d));"},
		{"merged-jsdoc-value", "postprocess.ts", "*//*", "*/ /*", "x; /**\n * one\n *//**\n * two\n */ y;"},
	}
}

type threePortCase struct {
	mutant, index int
	witness       bool
}
type threePortShard struct {
	name  string
	cases []threePortCase
}

func threePortShards(t *testing.T) []threePortShard {
	file, err := parser.ParseFile(token.NewFileSet(), "three_port_mutants_split_test.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	names := map[string]bool{}
	for _, decl := range file.Decls {
		if f, ok := decl.(*ast.FuncDecl); ok && strings.HasPrefix(f.Name.Name, "TestThreePortMutants_") && f.Name.Name != "TestThreePortMutants_ShardProof" {
			names[f.Name.Name] = true
		}
	}
	if len(names) != testThreePortMutantsShards {
		t.Fatalf("top-level enumeration: %d, declared %d", len(names), testThreePortMutantsShards)
	}
	for i := 0; i < testThreePortMutantsShards; i++ {
		if !names[fmt.Sprintf("TestThreePortMutants_%03d", i)] {
			t.Fatalf("missing shard %03d", i)
		}
	}
	return threePortShardPlan(t, generated(), threePortMutations(), testThreePortMutantsShards)
}

func threePortShardPlan(t *testing.T, cases []string, mutations []threePortMutation, declared int) []threePortShard {
	t.Helper()
	var shards []threePortShard
	var unsplit []threePortCase
	for m, mutation := range mutations {
		witnesses := 0
		var pairs []threePortCase
		for i, text := range cases {
			pair := threePortCase{m, i, strings.Contains(text, mutation.witness)}
			if pair.witness {
				witnesses++
			}
			pairs = append(pairs, pair)
			unsplit = append(unsplit, pair)
		}
		if witnesses != 1 {
			t.Fatalf("%s: expected one witness, got %d", mutation.name, witnesses)
		}
		for start := 0; start < len(pairs); start += threePortCasesPerShard {
			end := min(start+threePortCasesPerShard, len(pairs))
			shards = append(shards, threePortShard{fmt.Sprintf("shard-%03d", len(shards)), pairs[start:end]})
		}
	}
	if len(shards) != declared {
		t.Fatalf("enumerated %d shards, declared %d", len(shards), declared)
	}
	want := make(map[[2]int]bool)
	for _, pair := range unsplit {
		id := [2]int{pair.mutant, pair.index}
		if want[id] {
			t.Fatalf("duplicate unsplit case %v", id)
		}
		want[id] = true
	}
	seen := make(map[[2]int]bool)
	count := 0
	for _, shard := range shards {
		for _, pair := range shard.cases {
			id := [2]int{pair.mutant, pair.index}
			if !want[id] || seen[id] {
				t.Fatalf("%s: extra or repeated case %v", shard.name, id)
			}
			seen[id] = true
			count++
		}
	}
	if count != len(unsplit) || len(seen) != len(want) {
		t.Fatalf("shard union: %d/%d cases, %d/%d ids", count, len(unsplit), len(seen), len(want))
	}
	for id := range want {
		if !seen[id] {
			t.Fatalf("shard union missing %v", id)
		}
	}
	t.Logf("shard union: %d pairs, %d distinct ids, %d mutants x %d cases", count, len(seen), len(mutations), len(cases))
	return shards
}

// Canonical output ends every whole-file record with one escaped "stripped"
// line. Do not divide a file's AST, comments, or stripped-source checks.
func threePortCanonicalRecords(t *testing.T, data []byte, count int) [][]byte {
	t.Helper()
	var records [][]byte
	start, offset := 0, 0
	for _, line := range bytes.SplitAfter(data, []byte("\n")) {
		offset += len(line)
		if bytes.HasPrefix(line, []byte("stripped ")) {
			records = append(records, data[start:offset])
			start = offset
		}
	}
	if start != len(data) || len(records) != count {
		t.Fatalf("canonical records: %d, want %d; trailing bytes %d", len(records), count, len(data)-start)
	}
	return records
}

func checkThreePortCase(pair threePortCase, want, source, native []byte) error {
	sourceDifference := firstDifference(want, source)
	nativeDifference := firstDifference(want, native)
	if pair.witness {
		if sourceDifference == "" {
			return fmt.Errorf("case %03d: source Node mutant survived", pair.index)
		}
		if nativeDifference == "" {
			return fmt.Errorf("case %03d: sanitized native mutant survived", pair.index)
		}
	}
	if (sourceDifference == "") != (nativeDifference == "") {
		return fmt.Errorf("case %03d: mutant ports differ in agreement with Go", pair.index)
	}
	if diff := firstDifference(source, native); diff != "" {
		return fmt.Errorf("case %03d: mutant ports disagree: %s", pair.index, diff)
	}
	return nil
}

// Products are addressed by live source content and built once across shard processes.
func threePortProduct(t *testing.T, mutation threePortMutation) (string, string) {
	t.Helper()
	inputs := buildcache.Inputs{Name: "three-port-mutant-lowered-" + mutation.name,
		Files:     []string{"stage1/cohere/estree", "stage1/typescript", "internal", "cohere/TypeScript/tsc", "cohere/TypeScript-shim", "go.mod", "go.work"},
		Flags:     []string{mutation.file, mutation.from, mutation.to, "repository=" + root(t)},
		Toolchain: []string{runtime.Version(), runtime.GOOS, runtime.GOARCH}}
	lowered := buildcache.Product(t, inputs, func(dir string) error {
		source := filepath.Join(dir, "source")
		if err := os.Mkdir(source, 0755); err != nil {
			return err
		}
		files, err := filepath.Glob("*.ts")
		if err != nil {
			return err
		}
		for _, name := range files {
			data, err := os.ReadFile(name)
			if err != nil {
				return err
			}
			text := string(data)
			if name == mutation.file {
				if strings.Count(text, mutation.from) != 1 {
					return fmt.Errorf("mutant anchor count changed")
				}
				text = strings.Replace(text, mutation.from, mutation.to, 1)
			}
			text = strings.ReplaceAll(text, "'../../typescript/", "'"+filepath.ToSlash(filepath.Join(root(t), "stage1/typescript"))+"/")
			if err := os.WriteFile(filepath.Join(source, name), []byte(text), 0644); err != nil {
				return err
			}
		}
		program, err := load.Load([]string{filepath.Join(source, "main.ts")})
		if err != nil {
			return err
		}
		ir, err := lower.Lower(context.Background(), program)
		if err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(dir, "program.c"), []byte(native.C(ir)), 0644)
	})
	inputs.Name = "three-port-mutant-native-" + mutation.name
	inputs.Flags = append(inputs.Flags, "split=true")
	inputs.Flags = append(inputs.Flags, native.Flags(native.Options{Sanitize: true, Split: true})...)
	inputs.Flags = append(inputs.Flags, "ADAMIC_NATIVE_SPLIT="+os.Getenv("ADAMIC_NATIVE_SPLIT"), "ADAMIC_NATIVE_JOBS="+os.Getenv("ADAMIC_NATIVE_JOBS"), "ADAMIC_GATE_UNCACHED="+os.Getenv("ADAMIC_GATE_UNCACHED"))
	inputs.Toolchain = append(inputs.Toolchain, buildcache.Tool("clang", "--version"))
	binary := buildcache.Product(t, inputs, func(dir string) error {
		c, err := os.ReadFile(filepath.Join(lowered, "program.c"))
		if err != nil {
			return err
		}
		return native.Build(string(c), filepath.Join(dir, "port"), native.Options{Sanitize: true, Split: true})
	})
	return filepath.Join(lowered, "source", "main.ts"), filepath.Join(binary, "port")
}

func runThreePortShard(t *testing.T, index int) {
	t.Helper()
	shards := threePortShards(t)
	shard := shards[index]
	if planted := os.Getenv("ADAMIC_THREE_PORT_PROOF"); planted != "" {
		m, err := strconv.Atoi(planted)
		if err != nil {
			t.Fatal(err)
		}
		for _, pair := range shard.cases {
			want, source, native := []byte("oracle"), []byte("oracle"), []byte("oracle")
			if pair.witness {
				source, native = []byte("mutant"), []byte("mutant")
			}
			if pair.mutant == m && pair.witness {
				native = want
			}
			if err := checkThreePortCase(pair, want, source, native); err != nil {
				t.Fatal(err)
			}
		}
		return
	}
	mutation := threePortMutations()[shard.cases[0].mutant]
	start := time.Now()
	// GoBuild is not on this base; retain the original overlay Go build.
	oracle := goOracle(t)
	sourcePath, binary := threePortProduct(t, mutation)
	t.Logf("TestThreePortMutants (setup): %.3fs", time.Since(start).Seconds())
	texts := make([]string, len(shard.cases))
	for i, pair := range shard.cases {
		texts[i] = generated()[pair.index]
	}
	list := manifest(t, texts)
	want := threePortCanonicalRecords(t, execute(t, "", oracle, "--manifest", list), len(texts))
	source := threePortCanonicalRecords(t, onNode(t, sourcePath, "--manifest", list), len(texts))
	native := threePortCanonicalRecords(t, execute(t, "", binary, "--manifest", list), len(texts))
	for i, pair := range shard.cases {
		if err := checkThreePortCase(pair, want[i], source[i], native[i]); err != nil {
			t.Fatal(err)
		}
	}
	t.Logf("%s: %d cases", mutation.name, len(texts))
}

func TestThreePortMutants_000(t *testing.T) { t.Parallel(); runThreePortShard(t, 0) }
func TestThreePortMutants_001(t *testing.T) { t.Parallel(); runThreePortShard(t, 1) }
func TestThreePortMutants_002(t *testing.T) { t.Parallel(); runThreePortShard(t, 2) }
func TestThreePortMutants_003(t *testing.T) { t.Parallel(); runThreePortShard(t, 3) }
func TestThreePortMutants_004(t *testing.T) { t.Parallel(); runThreePortShard(t, 4) }
func TestThreePortMutants_005(t *testing.T) { t.Parallel(); runThreePortShard(t, 5) }
func TestThreePortMutants_006(t *testing.T) { t.Parallel(); runThreePortShard(t, 6) }
func TestThreePortMutants_007(t *testing.T) { t.Parallel(); runThreePortShard(t, 7) }
func TestThreePortMutants_008(t *testing.T) { t.Parallel(); runThreePortShard(t, 8) }
func TestThreePortMutants_009(t *testing.T) { t.Parallel(); runThreePortShard(t, 9) }
func TestThreePortMutants_010(t *testing.T) { t.Parallel(); runThreePortShard(t, 10) }
func TestThreePortMutants_011(t *testing.T) { t.Parallel(); runThreePortShard(t, 11) }

// Plant a surviving native witness and execute each actual top-level shard
// independently. This proof performs no builds and never measures the whole test.
func TestThreePortMutants_ShardProof(t *testing.T) {
	t.Parallel()
	shards := threePortShards(t)
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for m, mutation := range threePortMutations() {
		failed, caught, expected, caseID := 0, -1, -1, -1
		for s, shard := range shards {
			for _, pair := range shard.cases {
				if pair.mutant == m && pair.witness {
					expected = s
					caseID = pair.index
				}
			}
		}
		for s := range shards {
			name := fmt.Sprintf("TestThreePortMutants_%03d", s)
			command := exec.Command(executable, "-test.run=^"+name+"$", "-test.timeout=90s", "-test.v")
			command.Env = append(os.Environ(), fmt.Sprintf("ADAMIC_THREE_PORT_PROOF=%d", m))
			output, err := command.CombinedOutput()
			if err != nil {
				exit, ok := err.(*exec.ExitError)
				if !ok || exit.ExitCode() != 1 {
					t.Fatalf("proof child: %v: %s", err, output)
				}
				failed++
				caught = s
				if s != expected || !strings.Contains(string(output), fmt.Sprintf("case %03d: sanitized native mutant survived", caseID)) {
					t.Fatalf("unexpected failure in %s: %s", name, output)
				}
			}
		}
		if failed != 1 || caught != expected {
			t.Fatalf("planted %s: %d failing shards, caught %d want %d", mutation.name, failed, caught, expected)
		}
		t.Logf("planted surviving native %s case %03d caught by TestThreePortMutants_%03d; exactly one shard", mutation.name, caseID, caught)
	}
}
