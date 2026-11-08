package typeaware

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
	"unicode/utf16"
)

type sixShard struct {
	name string
	ids  []string
	run  func(*harness)
}

// i is zero based. Selection distributes complete deterministic units, not cases
// within a mutant: every mutant retains its original complete corpus.
func sixSelection(value string) (int, int, error) {
	if value == "" {
		return 0, 1, nil
	}
	parts := strings.Split(value, "/")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("ADAMIC_TEST_SHARD must be i/n (zero based), got %q", value)
	}
	index, a := strconv.Atoi(parts[0])
	n, b := strconv.Atoi(parts[1])
	if a != nil || b != nil || n < 1 || index < 0 || index >= n {
		return 0, 0, fmt.Errorf("invalid ADAMIC_TEST_SHARD %q", value)
	}
	return index, n, nil
}

func sixUnion(expected []string, shards []sixShard) error {
	want := make(map[string]bool, len(expected))
	for _, id := range expected {
		if want[id] {
			return fmt.Errorf("repeated unsplit case %s", id)
		}
		want[id] = true
	}
	seen := make(map[string]bool, len(expected))
	names := map[string]bool{}
	count := 0
	for _, shard := range shards {
		if names[shard.name] {
			return fmt.Errorf("repeated shard %s", shard.name)
		}
		names[shard.name] = true
		for _, id := range shard.ids {
			count++
			if !want[id] {
				return fmt.Errorf("shard %s has unexpected case %s", shard.name, id)
			}
			if seen[id] {
				return fmt.Errorf("shard %s repeats case %s", shard.name, id)
			}
			seen[id] = true
		}
	}
	for _, id := range expected {
		if !seen[id] {
			return fmt.Errorf("missing case %s", id)
		}
	}
	if count != len(expected) {
		return fmt.Errorf("union count %d, want %d", count, len(expected))
	}
	return nil
}

func sixIds(prefix string, paths []string) []string {
	ids := make([]string, len(paths))
	for index, path := range paths {
		name := filepath.Base(path)
		// Compiler namespaces intentionally repeat basenames. Retain their full
		// corpus-relative identity, independent of the checkout's location.
		if _, relative, ok := strings.Cut(filepath.ToSlash(path), "/src/compiler/"); ok {
			name = relative
		} else if repository, absolutePathError := filepath.Abs("../../.."); absolutePathError == nil {
			// Repository corpora also repeat basenames across directories.
			if relative, relativePathError := filepath.Rel(
				repository,
				path,
			); relativePathError == nil &&
				relative != ".." &&
				!strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
				name = filepath.ToSlash(relative)
			}
		}
		ids[index] = prefix + "/" + name
	}
	return ids
}

// Deterministic contiguous file ranges; keep all root files in both programs.
func sixRanges(paths []string, width int) [][]string {
	var result [][]string
	for start := 0; start < len(paths); start += width {
		end := min(start+width, len(paths))
		result = append(result, paths[start:end])
	}
	return result
}

func sixRunShards(test *testing.T, h *harness, expected []string, shards []sixShard, value string, required int) {
	test.Helper()
	if len(shards) != required {
		test.Fatalf("enumerated shard count %d differs from declared %d", len(shards), required)
	}
	if sixUnionError := sixUnion(expected, shards); sixUnionError != nil {
		test.Fatal(sixUnionError)
	}
	index, count, sixSelectionError := sixSelection(value)
	if sixSelectionError != nil {
		test.Fatal(sixSelectionError)
	}
	test.Logf("six-union count=%d unique=%d shards=%d", len(expected), len(expected), len(shards))
	started := h.setupStarted
	if started.IsZero() {
		started = time.Now()
	}
	var mutex sync.Mutex
	var admittedIds []string
	var completed []sixShard
	var admitted int
	var setupBefore time.Duration
	test.Cleanup(func() {
		cleanupStarted := time.Now()
		if sixUnionError := sixUnion(admittedIds, completed); sixUnionError != nil {
			test.Errorf("executed shard union: %v", sixUnionError)
		}
		test.Logf(
			"six-executed-union count=%d admitted_shards=%d selector=%d/%d",
			len(admittedIds),
			admitted,
			index,
			count,
		)
		test.Logf("split-setup elapsed_s=%.6f builds=shared-once", (setupBefore + time.Since(cleanupStarted)).Seconds())
	})
	width := max(3, len(strconv.Itoa(required-1)))
	for shardIndex, shard := range shards {
		if shardIndex%count != index {
			continue
		}
		unit := fmt.Sprintf("shard-%0*d", width, shardIndex)
		test.Run(unit, func(test *testing.T) {
			// Only matching Go -run subtests enter here. Validate their executed union,
			// while the independent complete-plan union above always covers every case.
			admittedIds = append(admittedIds, shard.ids...)
			admitted++
			test.Parallel()
			started := time.Now()
			defer func() {
				elapsed := time.Since(started)
				test.Logf(
					"six-shard name=%s content=%s cases=%d elapsed_s=%.6f",
					unit,
					shard.name,
					len(shard.ids),
					elapsed.Seconds(),
				)
				if elapsed > 30*time.Second {
					test.Errorf("shard %s (%s) exceeded 30 s: %s", unit, shard.name, elapsed)
				}
				mutex.Lock()
				completed = append(completed, shard)
				mutex.Unlock()
			}()
			directory := filepath.Join(h.directory, "shards", unit)
			if createDirectoryError := os.MkdirAll(directory, 0755); createDirectoryError != nil {
				test.Fatal(createDirectoryError)
			}
			local := &harness{t: test, repository: h.repository, directory: directory, parallel: true}
			shard.run(local)
		})
	}
	setupBefore = time.Since(started)
}

func sixOutputError(name string, want, got, stderr []byte) error {
	if len(stderr) != 0 {
		return fmt.Errorf("shard %s sanitizer stderr: %s", name, stderr)
	}
	if !bytes.Equal(want, got) {
		index := firstDifference(got, want)
		return fmt.Errorf(
			"shard %s mismatch byte %d: native %q Go %q",
			name,
			index,
			got[max(0, index-50):min(len(got), index+250)],
			want[max(0, index-50):min(len(want), index+250)],
		)
	}
	return nil
}

func sixWritten(text string) string {
	var out strings.Builder
	for _, r := range text {
		if r >= 32 && r <= 126 && r != 92 {
			out.WriteRune(r)
		} else if r <= 65535 {
			fmt.Fprintf(&out, `\u%04x`, r)
		} else {
			h, l := utf16.EncodeRune(r)
			fmt.Fprintf(&out, `\u%04x\u%04x`, h, l)
		}
	}
	return out.String()
}

func sixFindingUnion(stdout []byte, paths []string) error {
	var ids []string
	for _, line := range strings.Split(string(stdout), "\n") {
		if strings.HasPrefix(line, "file\t") {
			ids = append(ids, strings.TrimPrefix(line, "file\t"))
		}
	}
	expected := make([]string, len(paths))
	for index, p := range paths {
		expected[index] = sixWritten(p)
	}
	return sixUnion(expected, []sixShard{{name: "finding-output", ids: ids}})
}

func (local *harness) sixCompare(name, oracle, binary, configuration, roots string, paths []string) result {
	local.t.Helper()
	selected := local.write("selected.manifest", strings.Join(paths, "\n")+"\n")
	args := []string{configuration, roots, "--files", selected}
	want := local.must(name+"-go", exec.Command(oracle, args...))
	got := local.must(name+"-native", exec.Command(binary, args...))
	if err := sixOutputError(local.t.Name(), want.stdout, got.stdout, got.stderr); err != nil {
		local.t.Fatal(err)
	}
	if sixFindingUnionError := sixFindingUnion(want.stdout, paths); sixFindingUnionError != nil {
		local.t.Fatalf("shard %s oracle case union: %v", local.t.Name(), sixFindingUnionError)
	}
	if sixFindingUnionError := sixFindingUnion(got.stdout, paths); sixFindingUnionError != nil {
		local.t.Fatalf("shard %s native case union: %v", local.t.Name(), sixFindingUnionError)
	}
	local.t.Logf("%s: %d identical finding bytes; %s", name, len(want.stdout), summary(want.stdout))
	return want
}

func TestSixShardUnionRejectsLossAndDuplication(test *testing.T) {
	test.Parallel()
	ids := sixIds(
		"compiler",
		[]string{"/checkout/src/compiler/utilities.ts", "/checkout/src/compiler/_namespaces/utilities.ts"},
	)
	if sixUnionError := sixUnion(
		[]string{"compiler/utilities.ts", "compiler/_namespaces/utilities.ts"},
		[]sixShard{{name: "namespaces", ids: ids}},
	); sixUnionError != nil {
		test.Fatal(sixUnionError)
	}
	repository, absolutePathError := filepath.Abs("../../..")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	repositoryIds := sixIds(
		"repository",
		[]string{filepath.Join(repository, "fixtures/main.ts"), filepath.Join(repository, "stage1/main.ts")},
	)
	if sixUnionError := sixUnion(
		[]string{"repository/fixtures/main.ts", "repository/stage1/main.ts"},
		[]sixShard{{name: "repository", ids: repositoryIds}},
	); sixUnionError != nil {
		test.Fatal(sixUnionError)
	}
	expected := []string{"a", "b", "c"}
	for _, shards := range [][]sixShard{
		{{name: "left", ids: []string{"a"}}, {name: "right", ids: []string{"c"}}},
		{{name: "left", ids: []string{"a", "b"}}, {name: "right", ids: []string{"b", "c"}}},
		{{name: "left", ids: []string{"a", "b", "c", "extra"}}},
	} {
		if sixUnionError := sixUnion(expected, shards); sixUnionError == nil {
			test.Fatalf("invalid union accepted: %v", shards)
		}
	}
	for _, value := range []string{"1", "-1/2", "2/2", "0/0", "x/2", "0/2/3"} {
		if _, _, sixSelectionError := sixSelection(value); sixSelectionError == nil {
			test.Fatalf("invalid selector accepted: %s", value)
		}
	}
}

// The child executes the real scheduler and production output comparison. Only
// planted/case-5 is corrupted; exactly selection 2/4 must fail in shard-002.
const testSixShardPlantedDisagreementShards = 4

// Not parallel: parent and child share the planted-case subprocess protocol.
func TestSixShardPlantedDisagreement(test *testing.T) {
	if os.Getenv("ADAMIC_SIX_PLANTED_CHILD") == "1" {
		var shards []sixShard
		var expected []string
		for index := 0; index < testSixShardPlantedDisagreementShards; index++ {
			ids := []string{fmt.Sprintf("planted/case-%d", index*2), fmt.Sprintf("planted/case-%d", index*2+1)}
			expected = append(expected, ids...)
			shards = append(shards, sixShard{name: fmt.Sprintf("s-%02d", index), ids: ids, run: func(harness *harness) {
				for _, id := range ids {
					want := []byte("file\t" +
						id +
						"\n0\t1\t@typescript-eslint/no-unsafe-unary-minus\tunsafeUnaryMinus\nfindings 1\n")
					got := bytes.Clone(want)
					if id == "planted/case-5" {
						got[0] = 'X'
					}
					if err := sixOutputError(harness.t.Name(), want, got, nil); err != nil {
						harness.t.Fatal(err)
					}
				}
			}})
		}
		sixRunShards(
			test,
			&harness{directory: test.TempDir()},
			expected,
			shards,
			os.Getenv("ADAMIC_TEST_SHARD"),
			testSixShardPlantedDisagreementShards,
		)
		return
	}
	test.Parallel()
	for index := 0; index < testSixShardPlantedDisagreementShards; index++ {
		// Not parallel within the parent: a parallel subtest queues behind every waiting test in the unit, and that
		// wait counts as the parent's own time (99.6 s alone on one core, Home, for four subtests of 0.03 s each).
		test.Run(fmt.Sprintf("shard-%03d", index), func(test *testing.T) {
			for _, gate := range []bool{false, true} {
				pattern := "^TestSixShardPlantedDisagreement$"
				selector := fmt.Sprintf("%d/4", index)
				if gate {
					pattern += fmt.Sprintf("/^shard-%03d$", index)
					selector = ""
				}
				command := exec.Command(os.Args[0], "-test.run="+pattern, "-test.v")
				command.Env = append(os.Environ(), "ADAMIC_SIX_PLANTED_CHILD=1", "ADAMIC_TEST_SHARD="+selector)
				text, combinedOutputError := command.CombinedOutput()
				if index == 2 {
					if combinedOutputError == nil ||
						!bytes.Contains(text, []byte("shard TestSixShardPlantedDisagreement/shard-002 mismatch")) {
						test.Fatalf("holding shard failed to catch planted case: %v\n%s", combinedOutputError, text)
					}
					if bytes.Count(text, []byte("--- FAIL: TestSixShardPlantedDisagreement/shard-")) != 1 {
						test.Fatalf("expected exactly one failing shard:\n%s", text)
					}
					test.Log("planted/case-5 caught exactly by shard-002 (selection 2/4)")
				} else if combinedOutputError != nil {
					test.Fatalf("nonholding shard caught planted disagreement: %v\n%s", combinedOutputError, text)
				}
			}
		})
	}
}

// Partition by source size with deterministic ties, then restore corpus order.
func sixCompilerRanges(test *testing.T, paths []string, count int) [][]string {
	test.Helper()
	type file struct {
		index int
		size  int64
	}
	files := make([]file, len(paths))
	for index, p := range paths {
		info, statError := os.Stat(p)
		if statError != nil {
			test.Fatal(statError)
		}
		files[index] = file{index, info.Size()}
	}
	sort.Slice(files, func(left, right int) bool {
		if files[left].size == files[right].size {
			return files[left].index < files[right].index
		}
		return files[left].size > files[right].size
	})
	result := make([][]string, min(count, len(paths)))
	loads := make([]int64, len(result))
	indices := make([][]int, len(result))
	for _, file := range files {
		shard := 0
		for index := range loads {
			if loads[index] < loads[shard] {
				shard = index
			}
		}
		loads[shard] += file.size
		indices[shard] = append(indices[shard], file.index)
	}
	for indexIndex := range indices {
		sort.Ints(indices[indexIndex])
		for _, index := range indices[indexIndex] {
			result[indexIndex] = append(result[indexIndex], paths[index])
		}
	}
	return result
}

func globalManifestText(harness *harness, path string) string {
	harness.t.Helper()
	data, readError := os.ReadFile(path)
	if readError != nil {
		harness.t.Fatal(readError)
	}
	return string(data)
}

func sixRuleControlIds() []string {
	return []string{
		"reporting/no-unsafe-unary-minus",
		"reporting/related-getter-setter-pairs",
		"reporting/no-unsafe-declaration-merging",
		"reporting/no-unsafe-argument",
		"reporting/restrict-plus-operands",
		"reporting/no-unnecessary-boolean-literal-compare",
	}
}

// This enumerates the original checks, independently of the shard partitioner.
func sixUnsplitIds(generated, global, compiler []string, rounds int, benchmark bool) []string {
	ids := sixIds("agreement/generated", generated)
	ids = append(ids, sixRuleControlIds()...)
	ids = append(ids, sixIds("agreement/nonstrict", generated[:1])...)
	ids = append(ids, sixIds("agreement/cross-file", global)...)
	ids = append(ids, sixIds("mutant/declaration-source", global)...)
	ids = append(ids, "refusal/unlinked", "control/released-handle", "mutant/retained-handle")
	for _, name := range []string{
		"wrong-node",
		"last-declaration",
		"nullable-default",
		"union-members",
		"type-name",
		"raw-shape-constraint",
		"assignability-direction",
		"resolved-signature",
		"facts-length-asan",
		"facts-empty-bad",
		"facts-integer-bad",
		"facts-length-bad",
		"facts-version-bad",
	} {
		ids = append(ids, sixIds("mutant/"+name, generated)...)
	}
	for _, name := range []string{
		"assignable",
		"declarations",
		"signature",
		"raw-type",
		"nullable",
		"union",
		"signature-shape",
		"raw-shape",
		"type-shape",
		"options",
	} {
		for round := 1; round <= rounds; round++ {
			ids = append(ids, fmt.Sprintf("cost/%s-%d", name, round))
		}
	}
	ids = append(ids, sixIds("agreement/compiler", compiler)...)
	if benchmark {
		for round := 1; round <= 3; round++ {
			ids = append(ids, sixIds(fmt.Sprintf("bench/compiler-%d", round), compiler)...)
		}
	}
	return ids
}
