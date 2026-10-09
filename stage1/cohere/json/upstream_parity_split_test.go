package json

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/gatesample"
)

const testUpstreamRepositoryCorpusParitySplitShards = 16

var upstreamParityReports struct {
	sync.Mutex
	blocks    map[string]string
	completed int
}

var upstreamParitySetup struct {
	once   sync.Once
	cases  []textCase
	shards [][]textCase
	known  map[string]string
	oracle string
}

func upstreamParityOrdinal(id string) int {
	sum := sha256.Sum256([]byte(id))
	return int(binary.BigEndian.Uint64(sum[:8]) % testUpstreamRepositoryCorpusParitySplitShards)
}

func upstreamParityUnion(cases []textCase, shards [][]textCase) error {
	expected := map[string]textCase{}
	for _, item := range cases {
		if _, exists := expected[item.Name]; exists {
			return fmt.Errorf("duplicate live case %s", item.Name)
		}
		expected[item.Name] = item
	}
	seen := map[string]bool{}
	for ordinal, items := range shards {
		for _, item := range items {
			if want, ok := expected[item.Name]; !ok || want != item || seen[item.Name] || upstreamParityOrdinal(item.Name) != ordinal {
				return fmt.Errorf("invalid shard %03d case %s", ordinal, item.Name)
			}
			seen[item.Name] = true
		}
	}
	if len(seen) != len(expected) {
		return fmt.Errorf("union has %d of %d live cases", len(seen), len(expected))
	}
	return nil
}

func upstreamParityPrepare(t *testing.T) {
	t.Helper()
	if err := gatesample.Validate(); err != nil {
		t.Fatal(err)
	}
	if os.Getenv("ADAMIC_JSON_PRETTIER") == "" {
		t.Skip("set ADAMIC_JSON_PRETTIER for the separate upstream report")
	}
	upstreamParitySetup.once.Do(func() {
		started := time.Now()
		defer func() { t.Logf("TestUpstreamRepositoryCorpusParity (setup): %.3fs", time.Since(started).Seconds()) }()
		tree, err := parser.ParseFile(token.NewFileSet(), "upstream_parity_split_test.go", nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		enumerated := map[string]bool{}
		for _, decl := range tree.Decls {
			if f, ok := decl.(*ast.FuncDecl); ok && strings.HasPrefix(f.Name.Name, "TestUpstreamRepositoryCorpusParity_") {
				enumerated[f.Name.Name] = true
			}
		}
		if len(enumerated) != testUpstreamRepositoryCorpusParitySplitShards {
			t.Fatalf("enumerated %d tests, declared %d", len(enumerated), testUpstreamRepositoryCorpusParitySplitShards)
		}
		for i := 0; i < testUpstreamRepositoryCorpusParitySplitShards; i++ {
			if !enumerated[fmt.Sprintf("TestUpstreamRepositoryCorpusParity_%03d", i)] {
				t.Fatalf("missing shard %03d", i)
			}
		}
		cases := sampledCorpusCases(t, 8)
		shards := make([][]textCase, testUpstreamRepositoryCorpusParitySplitShards)
		for _, item := range cases {
			ordinal := upstreamParityOrdinal(item.Name)
			shards[ordinal] = append(shards[ordinal], item)
		}
		if err := upstreamParityUnion(cases, shards); err != nil {
			t.Fatal(err)
		}
		encoded, err := os.ReadFile("known-upstream-differences.txt")
		if err != nil {
			t.Fatal(err)
		}
		lines := strings.Split(strings.TrimSuffix(string(encoded), "\n"), "\n")
		if len(lines) != 27 {
			t.Fatalf("known report has %d lines, want 27", len(lines))
		}
		known := map[string]string{}
		for i := 0; i < len(lines); i += 3 {
			if _, exists := known[lines[i]]; exists {
				t.Fatalf("duplicate known difference %s", lines[i])
			}
			known[lines[i]] = strings.Join(lines[i:i+3], "\n") + "\n"
		}
		ids := map[string]bool{}
		for _, item := range cases {
			ids[item.Name] = true
		}
		for id := range known {
			if !ids[id] {
				t.Fatalf("known difference missing: %s", id)
			}
		}
		cohere, err := filepath.Abs("../../../cohere")
		if err != nil {
			t.Fatal(err)
		}
		side, err := filepath.Abs("testdata/cohere_side_test.go")
		if err != nil {
			t.Fatal(err)
		}
		flags := []string{"go test -c ./internal/format/javascript", "overlay=cohere_side_test.go"}
		for _, name := range []string{"GOFLAGS", "GOTOOLCHAIN", "GOOS", "GOARCH", "CGO_ENABLED", "GOEXPERIMENT", "GOWORK", "GOENV"} {
			flags = append(flags, name+"="+os.Getenv(name))
		}
		product := buildcache.Product(t, buildcache.Inputs{Name: "json upstream parity audit oracle", Files: []string{"cohere", "stage1/cohere/json/testdata/cohere_side_test.go"}, Flags: flags, Toolchain: []string{buildcache.Tool("go", "version")}}, func(dir string) error {
			overlay := filepath.Join(dir, "overlay.json")
			encoded, err := json.Marshal(map[string]any{"Replace": map[string]string{filepath.Join(cohere, "internal/format/javascript/adamic_json_audit_test.go"): side}})
			if err != nil {
				return err
			}
			if err := os.WriteFile(overlay, encoded, 0644); err != nil {
				return err
			}
			command := exec.Command("go", "test", "-c", "-overlay="+overlay, "-o", filepath.Join(dir, "oracle"), "./internal/format/javascript")
			command.Dir = cohere
			output, err := childguard.CombinedOutput(command, jsonGuard)
			if err != nil {
				return fmt.Errorf("Go oracle build: %w\n%s", err, output)
			}
			return nil
		})
		upstreamParitySetup.cases, upstreamParitySetup.shards, upstreamParitySetup.known, upstreamParitySetup.oracle = cases, shards, known, filepath.Join(product, "oracle")
		t.Logf("exact union: %d cases; all nine known differences assigned exactly once", len(cases))
	})
}

func upstreamParityBlock(id string, left, right answer) string {
	if sameAnswer(left, right) {
		return ""
	}
	return fmt.Sprintf("%s\nGo: %q error=%q\nPrettier: %q error=%q\n", id, left.Output, left.Error, right.Output, right.Error)
}

func upstreamParityBlockCheck(shard, id, got, want string) error {
	if got != want {
		return fmt.Errorf("%s case %s: upstream difference identity or answer changed; update the checked-in report", shard, id)
	}
	return nil
}

func upstreamParityRun(t *testing.T, ordinal int) {
	t.Helper()
	upstreamParityPrepare(t)
	cases := upstreamParitySetup.shards[ordinal]
	dir := t.TempDir()
	casesPath := filepath.Join(dir, "cases.json")
	writeJSON(t, casesPath, cases)
	readAnswers := func(path string) []answer {
		encoded, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		var result []answer
		if err := json.Unmarshal(encoded, &result); err != nil {
			t.Fatal(err)
		}
		if len(result) != len(cases) {
			t.Fatalf("%s answered %d of %d", path, len(result), len(cases))
		}
		return result
	}
	goPath := filepath.Join(dir, "go.json")
	command := bounded(t, upstreamParitySetup.oracle, "-test.v", "-test.run=^TestAdamicJSONAudit$")
	command.Env = append(os.Environ(), "ADAMIC_JSON_CASES="+casesPath, "ADAMIC_JSON_ANSWERS="+goPath)
	if output, err := childguard.CombinedOutput(command, jsonGuard); err != nil {
		t.Fatalf("Go cohere: %v\n%s", err, output)
	}
	prettierPath := filepath.Join(dir, "prettier.json")
	command = bounded(t, "node", "--stack-size=4096", "testdata/library.mjs", os.Getenv("ADAMIC_JSON_PRETTIER"), casesPath, prettierPath)
	if output, err := childguard.CombinedOutput(command, jsonGuard); err != nil {
		t.Fatalf("Prettier: %v\n%s", err, output)
	}
	left, right := readAnswers(goPath), readAnswers(prettierPath)
	blocks := make(map[string]string, len(cases))
	for i, item := range cases {
		block := upstreamParityBlock(item.Name, left[i], right[i])
		blocks[item.Name] = block
		allowed := item.Name == "stage1/cohere/json/gaps/numeric-separators.json" || strings.HasPrefix(item.Name, "generated/12/") || strings.HasPrefix(item.Name, "generated/13/") || strings.HasPrefix(item.Name, "generated/14/") || strings.HasPrefix(item.Name, "generated/16/")
		if block != "" && !allowed {
			t.Errorf("unexpected upstream difference: %s", item.Name)
		}
		if err := upstreamParityBlockCheck(t.Name(), item.Name, block, upstreamParitySetup.known[item.Name]); err != nil {
			t.Error(err)
		}
	}
	upstreamParityReports.Lock()
	defer upstreamParityReports.Unlock()
	if upstreamParityReports.blocks == nil {
		upstreamParityReports.blocks = map[string]string{}
	}
	for id, block := range blocks {
		upstreamParityReports.blocks[id] = block
	}
	upstreamParityReports.completed++
	if path := os.Getenv("ADAMIC_JSON_REPORT"); path != "" && upstreamParityReports.completed == testUpstreamRepositoryCorpusParitySplitShards {
		var report strings.Builder
		for _, item := range upstreamParitySetup.cases {
			report.WriteString(upstreamParityReports.blocks[item.Name])
		}
		if err := os.WriteFile(path, []byte(report.String()), 0644); err != nil {
			t.Fatal(err)
		}
	}
	t.Logf("shard %03d: %d cases", ordinal, len(cases))
}

func TestUpstreamRepositoryCorpusParity(t *testing.T) {
	t.Parallel()
	upstreamParityPrepare(t)
}

func TestUpstreamRepositoryCorpusParityShardProof(t *testing.T) {
	t.Parallel()
	cases := sampledCorpusCases(t, 8)
	if len(cases) == 0 {
		t.Fatal("empty live corpus")
	}
	planted := cases[0].Name
	shards := make([][]textCase, testUpstreamRepositoryCorpusParitySplitShards)
	for _, item := range cases {
		i := upstreamParityOrdinal(item.Name)
		shards[i] = append(shards[i], item)
	}
	if err := upstreamParityUnion(cases, shards); err != nil {
		t.Fatal(err)
	}
	caught := 0
	for i, items := range shards {
		for _, item := range items {
			left, right := answer{Output: item.Text}, answer{Output: item.Text}
			if item.Name == planted {
				right.Output = "{X}"
			}
			if err := upstreamParityBlockCheck(fmt.Sprintf("%03d", i), item.Name, upstreamParityBlock(item.Name, left, right), ""); err != nil {
				caught++
				t.Log(err)
			}
		}
	}
	if caught != 1 {
		t.Fatalf("planted failure caught by %d shards", caught)
	}
	shards[upstreamParityOrdinal(planted)] = nil
	if upstreamParityUnion(cases, shards) == nil {
		t.Fatal("missing case survived")
	}
}

func TestUpstreamRepositoryCorpusParity_000(t *testing.T) {
	t.Parallel()
	upstreamParityRun(t, 0)
}

func TestUpstreamRepositoryCorpusParity_001(t *testing.T) {
	t.Parallel()
	upstreamParityRun(t, 1)
}

func TestUpstreamRepositoryCorpusParity_002(t *testing.T) {
	t.Parallel()
	upstreamParityRun(t, 2)
}

func TestUpstreamRepositoryCorpusParity_003(t *testing.T) {
	t.Parallel()
	upstreamParityRun(t, 3)
}

func TestUpstreamRepositoryCorpusParity_004(t *testing.T) {
	t.Parallel()
	upstreamParityRun(t, 4)
}

func TestUpstreamRepositoryCorpusParity_005(t *testing.T) {
	t.Parallel()
	upstreamParityRun(t, 5)
}

func TestUpstreamRepositoryCorpusParity_006(t *testing.T) {
	t.Parallel()
	upstreamParityRun(t, 6)
}

func TestUpstreamRepositoryCorpusParity_007(t *testing.T) {
	t.Parallel()
	upstreamParityRun(t, 7)
}

func TestUpstreamRepositoryCorpusParity_008(t *testing.T) {
	t.Parallel()
	upstreamParityRun(t, 8)
}

func TestUpstreamRepositoryCorpusParity_009(t *testing.T) {
	t.Parallel()
	upstreamParityRun(t, 9)
}

func TestUpstreamRepositoryCorpusParity_010(t *testing.T) {
	t.Parallel()
	upstreamParityRun(t, 10)
}

func TestUpstreamRepositoryCorpusParity_011(t *testing.T) {
	t.Parallel()
	upstreamParityRun(t, 11)
}

func TestUpstreamRepositoryCorpusParity_012(t *testing.T) {
	t.Parallel()
	upstreamParityRun(t, 12)
}

func TestUpstreamRepositoryCorpusParity_013(t *testing.T) {
	t.Parallel()
	upstreamParityRun(t, 13)
}

func TestUpstreamRepositoryCorpusParity_014(t *testing.T) {
	t.Parallel()
	upstreamParityRun(t, 14)
}

func TestUpstreamRepositoryCorpusParity_015(t *testing.T) {
	t.Parallel()
	upstreamParityRun(t, 15)
}
