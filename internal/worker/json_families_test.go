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
	paths, globError := filepath.Glob("../oracle/testdata/json_" + mode + "_*.a")
	if globError != nil {
		panic(globError)
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
func workerJSONCorpusProduct(test *testing.T, mode string, build bool) string {
	test.Helper()
	inputs := buildcache.Inputs{Name: "worker JSON corpus " + mode,
		Files: []string{
			"internal/worker/json_families_test.go",
			"internal/worker/json_corpus_product.mjs",
			"internal/worker/json_" + mode + "_corpus.mjs",
		},
		Flags: []string{mode}, Toolchain: []string{buildcache.Tool("node", "--version")}}
	if !build {
		return findWorkerProduct(test, inputs)
	}
	return buildcache.Product(test, inputs, func(directory string) error {
		command := exec.CommandContext(test.Context(), "node", "json_corpus_product.mjs", mode, directory)
		if output, combinedOutputError := command.CombinedOutput(); combinedOutputError != nil {
			return fmt.Errorf("corpus inputs: %w\n%s", combinedOutputError, output)
		}
		return nil
	})
}

// Not parallel: generates the immutable decoder input corpus once.
func TestProduct_WorkerDecodeJSONCorpus(test *testing.T) {
	workerJSONCorpusProduct(test, "decode", true)
}

// Not parallel: generates the immutable encoder input corpus once.
func TestProduct_WorkerEncodeJSONCorpus(test *testing.T) {
	workerJSONCorpusProduct(test, "encode", true)
}

func catchWorkerJSONPlantedCase(test *testing.T, family string) {
	test.Helper()
	key := "corpus/00000"
	executable, executableError := os.Executable()
	if executableError != nil {
		test.Fatal(executableError)
	}
	command := exec.CommandContext(
		test.Context(),
		executable,
		"-test.run=^"+family+"_[0-9]+$",
		"-test.v",
		"-test.parallel=4",
		"-test.timeout=60s",
	)
	command.Env = append(os.Environ(), workerJSONProbeVariable+"="+family+"="+key)
	output, combinedOutputError := command.CombinedOutput()
	owner := familyCaseName(family, key)
	failed := regexp.MustCompile(`(?m)^--- FAIL: (`+family+`_[0-9]+) `).FindAllSubmatch(output, -1)
	if combinedOutputError == nil ||
		len(failed) != 1 ||
		string(failed[0][1]) != owner ||
		!bytes.Contains(output, []byte("--- FAIL: "+owner+"/"+key+" ")) ||
		!bytes.Contains(output, []byte("native disagreement")) {
		test.Fatalf(
			"planted wrong case must fail exactly %s/%s with native disagreement: %v\n%s",
			owner,
			key,
			combinedOutputError,
			output,
		)
	}
	test.Logf("planted wrong case caught by %s/%s", owner, key)
}
func TestWorkerDecodeJSONFamiliesCatchAPlantedCase(test *testing.T) {
	test.Parallel()
	catchWorkerJSONPlantedCase(test, "TestWorkerDecodeJSON")
}
func TestWorkerEncodeJSONFamiliesCatchAPlantedCase(test *testing.T) {
	test.Parallel()
	catchWorkerJSONPlantedCase(test, "TestWorkerEncodeJSON")
}

func TestWorkerJSONFamiliesCoverEveryText(test *testing.T) {
	test.Parallel()
	for _, mode := range []string{"decode", "encode"} {
		directory := workerJSONCorpusProduct(test, mode, false)
		read := func(name string) []string {
			contents, readError := os.ReadFile(filepath.Join(directory, name))
			if readError != nil {
				test.Fatal(readError)
			}
			var texts []string
			if decodeError := json.Unmarshal(contents, &texts); decodeError != nil {
				test.Fatal(decodeError)
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
				test.Fatalf("%s/%d has %d texts", mode, start, len(block))
			}
			offsets := workerJSONBlockOffsets(start)
			for index, text := range block {
				originalIndex := offsets[index/250] + index%250
				if seen[originalIndex] {
					test.Fatalf("repeated text index %d", originalIndex)
				}
				seen[originalIndex] = true
				combined[originalIndex] = text
			}
			counts[familyShard(fmt.Sprintf("corpus/%05d", start), family.shards)] += len(block)
		}
		if len(original) != 40000 || !bytes.Equal(mustJSON(test, original), mustJSON(test, combined)) {
			test.Fatal("family blocks omit, repeat or change a corpus text")
		}
		test.Logf("%s: texts per shard=%v; total=%d", mode, counts, len(combined))
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
