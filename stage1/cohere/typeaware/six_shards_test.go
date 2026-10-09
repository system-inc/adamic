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
	i, a := strconv.Atoi(parts[0])
	n, b := strconv.Atoi(parts[1])
	if a != nil || b != nil || n < 1 || i < 0 || i >= n {
		return 0, 0, fmt.Errorf("invalid ADAMIC_TEST_SHARD %q", value)
	}
	return i, n, nil
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

func sixIDs(prefix string, paths []string) []string {
	ids := make([]string, len(paths))
	for i, path := range paths {
		name := filepath.Base(path)
		// Compiler namespaces intentionally repeat basenames. Retain their full
		// corpus-relative identity, independent of the checkout's location.
		if _, relative, ok := strings.Cut(filepath.ToSlash(path), "/src/compiler/"); ok {
			name = relative
		} else if repository, err := filepath.Abs("../../.."); err == nil {
			// Repository corpora also repeat basenames across directories.
			if relative, err := filepath.Rel(repository, path); err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
				name = filepath.ToSlash(relative)
			}
		}
		ids[i] = prefix + "/" + name
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

func sixRunShards(t *testing.T, h *harness, expected []string, shards []sixShard, value string, required int) {
	t.Helper()
	if len(shards) != required {
		t.Fatalf("enumerated shard count %d differs from declared %d", len(shards), required)
	}
	if err := sixUnion(expected, shards); err != nil {
		t.Fatal(err)
	}
	index, count, err := sixSelection(value)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("six-union count=%d unique=%d shards=%d", len(expected), len(expected), len(shards))
	started := h.setupStarted
	if started.IsZero() {
		started = time.Now()
	}
	var mu sync.Mutex
	var admittedIDs []string
	var completed []sixShard
	var admitted int
	var setupBefore time.Duration
	t.Cleanup(func() {
		cleanupStarted := time.Now()
		if err := sixUnion(admittedIDs, completed); err != nil {
			t.Errorf("executed shard union: %v", err)
		}
		t.Logf("six-executed-union count=%d admitted_shards=%d selector=%d/%d", len(admittedIDs), admitted, index, count)
		t.Logf("split-setup elapsed_s=%.6f builds=shared-once", (setupBefore + time.Since(cleanupStarted)).Seconds())
	})
	width := max(3, len(strconv.Itoa(required-1)))
	for i, shard := range shards {
		if i%count != index {
			continue
		}
		unit := fmt.Sprintf("shard-%0*d", width, i)
		t.Run(unit, func(t *testing.T) {
			// Only matching Go -run subtests enter here. Validate their executed union,
			// while the independent complete-plan union above always covers every case.
			admittedIDs = append(admittedIDs, shard.ids...)
			admitted++
			t.Parallel()
			started := time.Now()
			defer func() {
				elapsed := time.Since(started)
				t.Logf("six-shard name=%s content=%s cases=%d elapsed_s=%.6f", unit, shard.name, len(shard.ids), elapsed.Seconds())
				if elapsed > 30*time.Second {
					t.Errorf("shard %s (%s) exceeded 30 s: %s", unit, shard.name, elapsed)
				}
				mu.Lock()
				completed = append(completed, shard)
				mu.Unlock()
			}()
			directory := filepath.Join(h.directory, "shards", unit)
			if err := os.MkdirAll(directory, 0755); err != nil {
				t.Fatal(err)
			}
			local := &harness{t: t, repository: h.repository, directory: directory, parallel: true}
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
		i := firstDifference(got, want)
		return fmt.Errorf("shard %s mismatch byte %d: native %q Go %q", name, i, got[max(0, i-50):min(len(got), i+250)], want[max(0, i-50):min(len(want), i+250)])
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
	for i, p := range paths {
		expected[i] = sixWritten(p)
	}
	return sixUnion(expected, []sixShard{{name: "finding-output", ids: ids}})
}

func (h *harness) sixCompare(name, oracle, binary, config, roots string, paths []string) result {
	h.t.Helper()
	selected := h.write("selected.manifest", strings.Join(paths, "\n")+"\n")
	args := []string{config, roots, "--files", selected}
	want := h.must(name+"-go", exec.Command(oracle, args...))
	got := h.must(name+"-native", exec.Command(binary, args...))
	if err := sixOutputError(h.t.Name(), want.stdout, got.stdout, got.stderr); err != nil {
		h.t.Fatal(err)
	}
	if err := sixFindingUnion(want.stdout, paths); err != nil {
		h.t.Fatalf("shard %s oracle case union: %v", h.t.Name(), err)
	}
	if err := sixFindingUnion(got.stdout, paths); err != nil {
		h.t.Fatalf("shard %s native case union: %v", h.t.Name(), err)
	}
	h.t.Logf("%s: %d identical finding bytes; %s", name, len(want.stdout), summary(want.stdout))
	return want
}

func TestSixShardUnionRejectsLossAndDuplication(t *testing.T) {
	t.Parallel()
	ids := sixIDs("compiler", []string{"/checkout/src/compiler/utilities.ts", "/checkout/src/compiler/_namespaces/utilities.ts"})
	if err := sixUnion([]string{"compiler/utilities.ts", "compiler/_namespaces/utilities.ts"}, []sixShard{{name: "namespaces", ids: ids}}); err != nil {
		t.Fatal(err)
	}
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	repositoryIDs := sixIDs("repository", []string{filepath.Join(repository, "fixtures/main.ts"), filepath.Join(repository, "stage1/main.ts")})
	if err := sixUnion([]string{"repository/fixtures/main.ts", "repository/stage1/main.ts"}, []sixShard{{name: "repository", ids: repositoryIDs}}); err != nil {
		t.Fatal(err)
	}
	expected := []string{"a", "b", "c"}
	for _, shards := range [][]sixShard{
		{{name: "left", ids: []string{"a"}}, {name: "right", ids: []string{"c"}}},
		{{name: "left", ids: []string{"a", "b"}}, {name: "right", ids: []string{"b", "c"}}},
		{{name: "left", ids: []string{"a", "b", "c", "extra"}}},
	} {
		if err := sixUnion(expected, shards); err == nil {
			t.Fatalf("invalid union accepted: %v", shards)
		}
	}
	for _, value := range []string{"1", "-1/2", "2/2", "0/0", "x/2", "0/2/3"} {
		if _, _, err := sixSelection(value); err == nil {
			t.Fatalf("invalid selector accepted: %s", value)
		}
	}
}

// The child executes the real scheduler and production output comparison. Only
// planted/case-5 is corrupted; exactly selection 2/4 must fail in shard-002.
const testSixShardPlantedDisagreementShards = 4

// Not parallel: parent and child share the planted-case subprocess protocol.
func TestSixShardPlantedDisagreement(t *testing.T) {
	if os.Getenv("ADAMIC_SIX_PLANTED_CHILD") == "1" {
		var shards []sixShard
		var expected []string
		for i := 0; i < testSixShardPlantedDisagreementShards; i++ {
			ids := []string{fmt.Sprintf("planted/case-%d", i*2), fmt.Sprintf("planted/case-%d", i*2+1)}
			expected = append(expected, ids...)
			shards = append(shards, sixShard{name: fmt.Sprintf("s-%02d", i), ids: ids, run: func(h *harness) {
				for _, id := range ids {
					want := []byte("file\t" + id + "\n0\t1\t@typescript-eslint/no-unsafe-unary-minus\tunsafeUnaryMinus\nfindings 1\n")
					got := bytes.Clone(want)
					if id == "planted/case-5" {
						got[0] = 'X'
					}
					if err := sixOutputError(h.t.Name(), want, got, nil); err != nil {
						h.t.Fatal(err)
					}
				}
			}})
		}
		sixRunShards(t, &harness{directory: t.TempDir()}, expected, shards, os.Getenv("ADAMIC_TEST_SHARD"), testSixShardPlantedDisagreementShards)
		return
	}
	t.Parallel()
	for i := 0; i < testSixShardPlantedDisagreementShards; i++ {
		t.Run(fmt.Sprintf("shard-%03d", i), func(t *testing.T) {
			t.Parallel()
			for _, gate := range []bool{false, true} {
				pattern := "^TestSixShardPlantedDisagreement$"
				selector := fmt.Sprintf("%d/4", i)
				if gate {
					pattern += fmt.Sprintf("/^shard-%03d$", i)
					selector = ""
				}
				cmd := exec.Command(os.Args[0], "-test.run="+pattern, "-test.v")
				cmd.Env = append(os.Environ(), "ADAMIC_SIX_PLANTED_CHILD=1", "ADAMIC_TEST_SHARD="+selector)
				text, err := cmd.CombinedOutput()
				if i == 2 {
					if err == nil || !bytes.Contains(text, []byte("shard TestSixShardPlantedDisagreement/shard-002 mismatch")) {
						t.Fatalf("holding shard failed to catch planted case: %v\n%s", err, text)
					}
					if bytes.Count(text, []byte("--- FAIL: TestSixShardPlantedDisagreement/shard-")) != 1 {
						t.Fatalf("expected exactly one failing shard:\n%s", text)
					}
					t.Log("planted/case-5 caught exactly by shard-002 (selection 2/4)")
				} else if err != nil {
					t.Fatalf("nonholding shard caught planted disagreement: %v\n%s", err, text)
				}
			}
		})
	}
}

// Partition by source size with deterministic ties, then restore corpus order.
func sixCompilerRanges(t *testing.T, paths []string, count int) [][]string {
	t.Helper()
	type file struct {
		index int
		size  int64
	}
	files := make([]file, len(paths))
	for i, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			t.Fatal(err)
		}
		files[i] = file{i, info.Size()}
	}
	sort.Slice(files, func(i, j int) bool {
		if files[i].size == files[j].size {
			return files[i].index < files[j].index
		}
		return files[i].size > files[j].size
	})
	result := make([][]string, min(count, len(paths)))
	loads := make([]int64, len(result))
	indices := make([][]int, len(result))
	for _, f := range files {
		shard := 0
		for i := range loads {
			if loads[i] < loads[shard] {
				shard = i
			}
		}
		loads[shard] += f.size
		indices[shard] = append(indices[shard], f.index)
	}
	for i := range indices {
		sort.Ints(indices[i])
		for _, index := range indices[i] {
			result[i] = append(result[i], paths[index])
		}
	}
	return result
}

func globalManifestText(h *harness, path string) string {
	h.t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		h.t.Fatal(err)
	}
	return string(data)
}

func sixRuleControlIDs() []string {
	return []string{"reporting/no-unsafe-unary-minus", "reporting/related-getter-setter-pairs", "reporting/no-unsafe-declaration-merging", "reporting/no-unsafe-argument", "reporting/restrict-plus-operands", "reporting/no-unnecessary-boolean-literal-compare"}
}

// This enumerates the original checks, independently of the shard partitioner.
func sixUnsplitIDs(generated, global, compiler []string, rounds int, benchmark bool) []string {
	ids := sixIDs("agreement/generated", generated)
	ids = append(ids, sixRuleControlIDs()...)
	ids = append(ids, sixIDs("agreement/nonstrict", generated[:1])...)
	ids = append(ids, sixIDs("agreement/cross-file", global)...)
	ids = append(ids, sixIDs("mutant/declaration-source", global)...)
	ids = append(ids, "refusal/unlinked", "control/released-handle", "mutant/retained-handle")
	for _, name := range []string{"wrong-node", "last-declaration", "nullable-default", "union-members", "type-name", "raw-shape-constraint", "assignability-direction", "resolved-signature", "facts-length-asan", "facts-empty-bad", "facts-integer-bad", "facts-length-bad", "facts-version-bad"} {
		ids = append(ids, sixIDs("mutant/"+name, generated)...)
	}
	for _, name := range []string{"assignable", "declarations", "signature", "raw-type", "nullable", "union", "signature-shape", "raw-shape", "type-shape", "options"} {
		for round := 1; round <= rounds; round++ {
			ids = append(ids, fmt.Sprintf("cost/%s-%d", name, round))
		}
	}
	ids = append(ids, sixIDs("agreement/compiler", compiler)...)
	if benchmark {
		for round := 1; round <= 3; round++ {
			ids = append(ids, sixIDs(fmt.Sprintf("bench/compiler-%d", round), compiler)...)
		}
	}
	return ids
}
