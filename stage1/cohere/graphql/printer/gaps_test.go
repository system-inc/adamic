package printer

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

func TestPrinterConstructorGap(t *testing.T) {
	t.Parallel()
	path, _ := filepath.Abs("gaps/nestedConstructor.ts")
	result := onNode(t, path)
	if result.exitCode != 0 || string(result.stdout) != "ready\n" {
		t.Fatalf("Node: %+v", result)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	_, err = lower.Lower(context.Background(), program)
	var refused *lower.Refused
	if !errors.As(err, &refused) || !strings.Contains(refused.What, "this escaping a constructor before every field is set") {
		t.Fatalf("gap changed; update GAPS.md and remove workaround: %v", err)
	}
	t.Log(err)
}

const testPrinterWhitespaceGapShards = 5

// A fixed set of shards hashes the repository-relative JSON path and entry index.
// Adding entries preserves every existing assignment. ADAMIC_TEST_SHARD=i/n
// selects indices modulo n equal to i; unset runs all. Builds are shared inputs.
// Not parallel: printerBuild writes the shared buildcache directory and native.Build writes the shared adamic/runtime cache.
func TestPrinterWhitespaceGap(t *testing.T) {
	library := os.Getenv("ADAMIC_GRAPHQL_PRETTIER")
	if library == "" {
		t.Skip("set ADAMIC_GRAPHQL_PRETTIER to prettier@3.9.6 and graphql@17.0.2")
	}
	proof, err := os.ReadFile("gaps/whitespace-cases.json")
	if err != nil {
		t.Fatal(err)
	}
	var texts []string
	if err := json.Unmarshal(proof, &texts); err != nil {
		t.Fatal(err)
	}
	cases, answers := printerCases(t, "defaults", printerOracle(t))
	enumeration := enumeratePrinter(t, "defaults", cases, answers)
	byInput := map[string]string{}
	for _, item := range enumeration {
		byInput[item.input] = item.want
	}
	escape := strings.NewReplacer("\\", "\\\\", "\n", "\\n", "\r", "\\r", "\t", "\\t")
	whole := make([]printerCase, len(texts))
	for i, text := range texts {
		input := ">" + escape.Replace(text)
		answer, exists := byInput[input]
		if !exists || !strings.HasPrefix(answer, "error\tSyntax Error:") {
			t.Fatalf("Go whitespace refusal missing: %q", input)
		}
		whole[i] = printerCase{id: whitespaceCaseID(i), input: input, want: answer}
	}
	if len(whole) == 0 {
		t.Fatal("repository whitespace corpus is empty")
	}
	shards := whitespaceShards(whole)
	for i := range shards {
		path := filepath.Join(t.TempDir(), "whitespace-cases.txt")
		var protocol strings.Builder
		for _, item := range shards[i].cases {
			protocol.WriteString(item.input + "\n")
		}
		if err := os.WriteFile(path, []byte(protocol.String()), 0644); err != nil {
			t.Fatal(err)
		}
		shards[i].path = path
	}
	if len(shards) != testPrinterWhitespaceGapShards {
		t.Fatalf("enumerated %d shards, declared %d", len(shards), testPrinterWhitespaceGapShards)
	}
	if err := printerShardUnion(whole, shards); err != nil {
		t.Fatal(err)
	}
	t.Logf("union: %d unique whitespace case ids across %d shards", len(whole), len(shards))
	selected, err := printerShardSelection(os.Getenv("ADAMIC_TEST_SHARD"), len(shards))
	if err != nil {
		t.Fatal(err)
	}
	products := preparePrinterProducts(t, printerDirectory(t, "", "", ""))
	script, _ := filepath.Abs("testdata/prettier.mjs")
	embedded, _ := filepath.Abs(filepath.Join(repository, "cohere/internal/format/prettier/bundles"))
	for number, shard := range shards {
		if !selected[number] {
			continue
		}
		t.Run(fmt.Sprintf("shard-%03d", number), func(t *testing.T) {
			t.Parallel()
			start := time.Now()
			t.Cleanup(func() {
				if elapsed := time.Since(start); elapsed > 30*time.Second {
					t.Errorf("invalid test unit: %.3fs exceeds 30s", elapsed.Seconds())
				}
			})
			t.Logf("whitespace cases: %v", shard.cases)
			if len(shard.cases) == 0 {
				return
			}
			for _, side := range []struct {
				name   string
				result run
			}{
				{"native", execute(t, []string{"ASAN_OPTIONS=detect_leaks=0"}, products.sanitized, "--cases", shard.path)},
				{"Node", onNode(t, products.source, "--cases", shard.path)},
			} {
				if err := printerShardDisagreement(number, shard, side.result); err != nil {
					t.Errorf("%s: %v", side.name, err)
				}
			}
			for _, side := range []struct{ directory, engine string }{{library, "npm"}, {embedded, "embedded"}} {
				result := execute(t, nil, "node", script, side.directory, shard.path, "defaults", side.engine)
				if result.exitCode != 0 || len(result.stderr) != 0 || string(result.stdout) != strings.Repeat("ok\t\n", len(shard.cases)) {
					t.Errorf("shard-%03d %s Prettier %s: %+v", number, fmt.Sprint(shard.cases), side.engine, result)
				}
			}
		})
	}
}

// A real process emits one planted acceptance where the oracle requires refusal;
// only the owning whitespace shard may report the disagreement.
func whitespaceCaseID(index int) string {
	return fmt.Sprintf("stage1/cohere/graphql/printer/gaps/whitespace-cases.json/case-%06d", index)
}

func whitespaceOwner(key string) int {
	hash := sha256.Sum256([]byte(key))
	return int(binary.BigEndian.Uint64(hash[:8]) % testPrinterWhitespaceGapShards)
}

func whitespaceShards(cases []printerCase) []printerShard {
	shards := make([]printerShard, testPrinterWhitespaceGapShards)
	for i := range shards {
		shards[i].mode = "defaults"
	}
	for _, item := range cases {
		owner := whitespaceOwner(item.id)
		shards[owner].cases = append(shards[owner].cases, item)
	}
	return shards
}

func TestPrinterWhitespacePlantedDisagreement(t *testing.T) {
	var whole []printerCase
	for i := range 5 {
		whole = append(whole, printerCase{id: whitespaceCaseID(i), want: "error\trefused"})
	}
	planted := whitespaceCaseID(3)
	caught := 0
	for number, shard := range whitespaceShards(whole) {
		var answer strings.Builder
		for _, item := range shard.cases {
			if item.id == planted {
				answer.WriteString("ok\t\n")
			} else {
				answer.WriteString(item.want + "\n")
			}
		}
		encoded, _ := json.Marshal(answer.String())
		result := execute(t, nil, "node", "-e", "process.stdout.write("+string(encoded)+")")
		err := printerShardDisagreement(number, shard, result)
		if err == nil {
			continue
		}
		if number != whitespaceOwner(planted) || !strings.Contains(err.Error(), planted) {
			t.Fatalf("wrong owner: %v", err)
		}
		t.Log(err)
		caught++
	}
	if caught != 1 {
		t.Fatalf("%d shards caught planted disagreement, want 1", caught)
	}
}

func TestPrinterWhitespaceShardGrowth(t *testing.T) {
	original := []printerCase{{id: whitespaceCaseID(0)}, {id: whitespaceCaseID(1)}}
	grown := append(append([]printerCase(nil), original...), printerCase{id: whitespaceCaseID(2)})
	before, after := whitespaceShards(original), whitespaceShards(grown)
	if len(before) != testPrinterWhitespaceGapShards || len(after) != len(before) {
		t.Fatal("growth changed shard count")
	}
	for _, cases := range [][]printerCase{original, grown} {
		if err := printerShardUnion(cases, whitespaceShards(cases)); err != nil {
			t.Fatal(err)
		}
	}
	for number, shard := range before {
		for _, item := range shard.cases {
			found := false
			for _, candidate := range after[number].cases {
				if candidate.id == item.id {
					found = true
				}
			}
			if !found {
				t.Fatalf("growth moved %s from shard-%03d", item.id, number)
			}
		}
	}
}
