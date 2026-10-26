package worker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
)

// Each 1,000-text case interleaves 250 original texts from each schema. Thus
// every Worker process crosses schemas (and exercises its descriptor cache),
// while forty disjoint cases cover all 40,000 original indices exactly once.
// FNV assigns these case names, fixture names and mutant names.
func workerJSONBehaviorKeys(mode string) []string {
	// Reads: internal/oracle/testdata/json_*.a
	paths, err := filepath.Glob("../oracle/testdata/json_" + mode + "_*.a")
	if err != nil {
		panic(err)
	}
	keys := []string{}
	for _, path := range paths {
		keys = append(keys, "fixture/"+filepath.Base(path))
	}
	if mode == "decode" {
		keys = append(keys, "fixture/fast_risks.a")
	}
	for start := 0; start < 40000; start += 1000 {
		keys = append(keys, fmt.Sprintf("corpus/%05d", start))
	}
	mutants := workerJSONDecodeMutants
	if mode == "encode" {
		mutants = workerJSONEncodeMutants
	}
	for _, mutant := range mutants {
		keys = append(keys, "mutant/"+mutant.name)
	}
	return keys
}
func workerJSONDecodeKeys() []string { return workerJSONBehaviorKeys("decode") }
func workerJSONEncodeKeys() []string { return workerJSONBehaviorKeys("encode") }

const workerJSONProbeVariable = "ADAMIC_WORKER_JSON_PLANTED_CASE"

func workerJSONProbe(family, key string) bool {
	return os.Getenv(workerJSONProbeVariable) == family+"="+key
}
func workerJSONCaseOwns(family string, shard int, key string) bool {
	probe := os.Getenv(workerJSONProbeVariable)
	return familyOwns(family, shard, key) && (probe == "" || probe == family+"="+key)
}
func workerJSONShardHasCases(family string, shard int) bool {
	for _, key := range familyNamed(family).keys() {
		if workerJSONCaseOwns(family, shard, key) {
			return true
		}
	}
	return false
}

// Corpus products contain deterministic input texts and JSON.stringify witnesses,
// never native or generated-Worker verdicts. Every behavior case observes those anew.
func workerJSONCorpusProduct(t *testing.T, mode string, build bool) string {
	t.Helper()
	inputs := buildcache.Inputs{Name: "worker JSON corpus " + mode,
		Files: []string{"internal/worker/json_families_test.go", "internal/worker/json_corpus_product.mjs", "internal/worker/json_" + mode + "_corpus.mjs"},
		Flags: []string{mode}, Toolchain: []string{buildcache.Tool("node", "--version")}}
	if !build {
		return findWorkerProduct(t, inputs)
	}
	return buildcache.Product(t, inputs, func(directory string) error {
		command := exec.CommandContext(t.Context(), "node", "json_corpus_product.mjs", mode, directory)
		if output, err := command.CombinedOutput(); err != nil {
			return fmt.Errorf("corpus inputs: %w\n%s", err, output)
		}
		return nil
	})
}

// Not parallel: generates the immutable decoder input corpus once.
func TestProduct_WorkerDecodeJSONCorpus(t *testing.T) { workerJSONCorpusProduct(t, "decode", true) }

// Not parallel: generates the immutable encoder input corpus once.
func TestProduct_WorkerEncodeJSONCorpus(t *testing.T) { workerJSONCorpusProduct(t, "encode", true) }

func catchWorkerJSONPlantedCase(t *testing.T, family string) {
	t.Helper()
	key := "corpus/00000"
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	command := exec.CommandContext(t.Context(), executable, "-test.run=^"+family+"_[0-9]+$", "-test.v", "-test.parallel=4", "-test.timeout=60s")
	command.Env = append(os.Environ(), workerJSONProbeVariable+"="+family+"="+key)
	output, err := command.CombinedOutput()
	owner := familyCaseName(family, key)
	failed := regexp.MustCompile(`(?m)^--- FAIL: (`+family+`_[0-9]+) `).FindAllSubmatch(output, -1)
	if err == nil || len(failed) != 1 || string(failed[0][1]) != owner || !bytes.Contains(output, []byte("--- FAIL: "+owner+"/"+key+" ")) || !bytes.Contains(output, []byte("native disagreement")) {
		t.Fatalf("planted wrong case must fail exactly %s/%s with native disagreement: %v\n%s", owner, key, err, output)
	}
	t.Logf("planted wrong case caught by %s/%s", owner, key)
}
func TestWorkerDecodeJSONFamiliesCatchAPlantedCase(t *testing.T) {
	t.Parallel()
	catchWorkerJSONPlantedCase(t, "TestWorkerDecodeJSON")
}
func TestWorkerEncodeJSONFamiliesCatchAPlantedCase(t *testing.T) {
	t.Parallel()
	catchWorkerJSONPlantedCase(t, "TestWorkerEncodeJSON")
}

func TestWorkerJSONFamiliesCoverEveryText(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"decode", "encode"} {
		directory := workerJSONCorpusProduct(t, mode, false)
		read := func(name string) []string {
			contents, err := os.ReadFile(filepath.Join(directory, name))
			if err != nil {
				t.Fatal(err)
			}
			var texts []string
			if err := json.Unmarshal(contents, &texts); err != nil {
				t.Fatal(err)
			}
			return texts
		}
		original := read("corpus.json")
		combined := make([]string, len(original))
		seen := make([]bool, len(original))
		familyName := "TestWorkerDecodeJSON"
		if mode == "encode" {
			familyName = "TestWorkerEncodeJSON"
		}
		family := familyNamed(familyName)
		counts := make([]int, family.shards)
		for start := 0; start < 40000; start += 1000 {
			block := read(fmt.Sprintf("%05d.json", start))
			if len(block) != 1000 {
				t.Fatalf("%s/%d has %d texts", mode, start, len(block))
			}
			offsets := workerJSONBlockOffsets(start)
			for index, text := range block {
				originalIndex := offsets[index/250] + index%250
				if seen[originalIndex] {
					t.Fatalf("repeated text index %d", originalIndex)
				}
				seen[originalIndex] = true
				combined[originalIndex] = text
			}
			counts[familyShard(fmt.Sprintf("corpus/%05d", start), family.shards)] += len(block)
		}
		if len(original) != 40000 || !bytes.Equal(mustJSON(t, original), mustJSON(t, combined)) {
			t.Fatal("family blocks omit, repeat or change a corpus text")
		}
		t.Logf("%s: texts per shard=%v; total=%d", mode, counts, len(combined))
	}
}

// Every batch carries all four schemas, including a required-field witness.
func workerJSONMutantBlocks() []int {
	starts := []int{}
	for start := 0; start < 40000; start += 1000 {
		starts = append(starts, start)
	}
	return starts
}
func workerJSONBlockOffsets(start int) []int {
	position := start / 4
	return []int{position, 10000 + position, 20000 + position, 30000 + position}
}
func workerJSONOffsetsArgument(offsets []int) string {
	values := []string{}
	for _, offset := range offsets {
		values = append(values, strconv.Itoa(offset))
	}
	return strings.Join(values, ",")
}
