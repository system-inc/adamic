package printer

import (
	"context"
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

// Each independent whitespace input owns one shard-NNN. ADAMIC_TEST_SHARD=i/n
// selects indices modulo n equal to i; unset runs all. Builds are shared inputs.
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
		whole[i] = printerCase{id: fmt.Sprintf("whitespace/case-%06d", i), input: input, want: answer}
	}
	shards := make([]printerShard, len(whole))
	for i, item := range whole {
		path := filepath.Join(t.TempDir(), "whitespace-cases.txt")
		if err := os.WriteFile(path, []byte(item.input+"\n"), 0644); err != nil {
			t.Fatal(err)
		}
		shards[i] = printerShard{mode: "defaults", path: path, cases: []printerCase{item}}
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
			t.Logf("whitespace case %s", shard.cases[0].id)
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
				if result.exitCode != 0 || len(result.stderr) != 0 || string(result.stdout) != "ok\t\n" {
					t.Errorf("shard-%03d %s Prettier %s: %+v", number, shard.cases[0].id, side.engine, result)
				}
			}
		})
	}
}

// A real process emits one planted acceptance where the oracle requires refusal;
// only the owning whitespace shard may report the disagreement.
func TestPrinterWhitespacePlantedDisagreement(t *testing.T) {
	caught := 0
	for number := 0; number < testPrinterWhitespaceGapShards; number++ {
		shard := printerShard{cases: []printerCase{{id: fmt.Sprintf("whitespace/case-%06d", number), want: "error\trefused"}}}
		answer := "error\trefused\n"
		if number == 3 {
			answer = "ok\t\n"
		}
		encoded, _ := json.Marshal(answer)
		result := execute(t, nil, "node", "-e", "process.stdout.write("+string(encoded)+")")
		err := printerShardDisagreement(number, shard, result)
		if err == nil {
			if number == 3 {
				t.Fatal("planted disagreement escaped shard-003")
			}
			continue
		}
		if number != 3 || !strings.Contains(err.Error(), "shard-003 whitespace/case-000003") {
			t.Fatalf("wrong owner: %v", err)
		}
		t.Log(err)
		caught++
	}
	if caught != 1 {
		t.Fatalf("%d shards caught planted disagreement, want 1", caught)
	}
}
