package yaml

import (
	"bytes"
	"fmt"
	"hash/fnv"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// yamlShards gives each case to the shard the FNV-64a hash of its complete encoded line names, chunk size included, so
// repeated cases keep their multiplicity and inserting a case can't move another's owner.
func yamlShards(cases []string, count int) [][]int {
	shards := make([][]int, count)
	for index, input := range cases {
		hash := fnv.New64a()
		_, _ = hash.Write([]byte(input))
		owner := int(hash.Sum64() % uint64(count))
		shards[owner] = append(shards[owner], index)
	}
	return shards
}

// yamlMatchGo is a comparison with Go too long for one test, its corpus split among shards by case content
// (#9wyp4tt): every shard runs Go, native under the sanitizers, Node and the emitted JavaScript over its own cases.
type yamlMatchGo struct {
	family    string
	shards    int
	cases     func(t *testing.T) (string, int, int)
	oracle    func(t *testing.T) string
	arguments []string
	build     yamlBuild
	entry     string
}

// corpus is the family's whole corpus, one encoded case a line, and its count of repository files.
func (comparison yamlMatchGo) corpus(t *testing.T) ([]string, int) {
	t.Helper()
	path, files, count := comparison.cases(t)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	cases := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
	if len(cases) != count {
		t.Fatalf("enumerated %d cases; the generator counted %d", len(cases), count)
	}
	return cases, files
}

// yamlShardRun is one shard's cases and what each side answered for them.
type yamlShardRun struct {
	all                             []string
	indices                         []int
	files                           int
	path, entry                     string
	expected, native, node, emitted []byte
}

// run fetches the products, then runs Go and the three sides over the shard's cases, in the order the whole test ran
// them.
func (comparison yamlMatchGo) run(t *testing.T, shard int) yamlShardRun {
	t.Helper()
	oracle := comparison.oracle(t)
	binary := comparison.build.binary(t, comparison.entry)
	emitted := comparison.build.emitted(t, comparison.entry)
	all, files := comparison.corpus(t)
	indices := yamlShards(all, comparison.shards)[shard]
	if len(indices) == 0 {
		t.Fatal("empty shard")
	}
	cases := make([]string, len(indices))
	for i, index := range indices {
		cases[i] = all[index]
	}
	path := filepath.Join(t.TempDir(), comparison.family+"-cases.txt")
	if err := os.WriteFile(path, []byte(strings.Join(cases, "\n")+"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	entry, err := filepath.Abs(comparison.entry)
	if err != nil {
		t.Fatal(err)
	}
	runner := filepath.Join(root, "oracle/node.mjs")
	arguments := append(append([]string(nil), comparison.arguments...), path)
	result := yamlShardRun{all: all, indices: indices, files: files, path: path, entry: entry}
	result.expected = run(t, "", nil, oracle, arguments...)
	result.native = run(t, "", []string{"ASAN_OPTIONS=detect_leaks=1"}, binary, arguments...)
	result.node = run(t, "", nil, "node", append([]string{"--disable-warning=ExperimentalWarning", runner, entry}, arguments...)...)
	result.emitted = run(t, "", nil, "node", append([]string{"--disable-warning=ExperimentalWarning", runner, emitted}, arguments...)...)
	t.Logf("%s shard %03d: %d of %d cases", comparison.family, shard, len(indices), len(all))
	return result
}

// diagnostic is where a failing shard leaves its bytes, one name per family, shard and side.
func (comparison yamlMatchGo) diagnostic(shard int, side string) string {
	return fmt.Sprintf("/tmp/stage1-yaml-%s-%03d-%s.txt", comparison.family, shard, side)
}

// union holds every case to exactly one shard.
func (comparison yamlMatchGo) union(t *testing.T) []string {
	t.Helper()
	cases, _ := comparison.corpus(t)
	visits := make([]int, len(cases))
	for _, indices := range yamlShards(cases, comparison.shards) {
		for _, index := range indices {
			visits[index]++
		}
	}
	for index, count := range visits {
		if count != 1 {
			t.Fatalf("case %d visited %d times", index, count)
		}
	}
	t.Logf("counted union: all %d %s cases visited exactly once over %d shards", len(cases), comparison.family, comparison.shards)
	return cases
}

// planted runs the first case through Go and native, plants a byte in native's real stdout, and requires exactly one
// shard to own the disagreement.
func (comparison yamlMatchGo) planted(t *testing.T) {
	t.Helper()
	oracle := comparison.oracle(t)
	binary := comparison.build.binary(t, comparison.entry)
	cases, _ := comparison.corpus(t)
	path := filepath.Join(t.TempDir(), "cases.txt")
	if err := os.WriteFile(path, []byte(cases[0]+"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	arguments := append(append([]string(nil), comparison.arguments...), path)
	expected := run(t, "", nil, oracle, arguments...)
	actual := run(t, "", []string{"ASAN_OPTIONS=detect_leaks=1"}, binary, arguments...)
	if !bytes.Equal(actual, expected) {
		t.Fatal("native control disagrees before planting")
	}
	actual = append(actual, '!')
	var caught []int
	for shard, indices := range yamlShards(cases, comparison.shards) {
		for _, index := range indices {
			output := expected
			if index == 0 {
				output = actual
			}
			if !bytes.Equal(output, expected) {
				caught = append(caught, shard)
			}
		}
	}
	if len(caught) != 1 {
		t.Fatalf("planted stdout disagreement caught by %v; want one owner", caught)
	}
	t.Logf("planted real native stdout disagreement caught by the %s shard %03d only", comparison.family, caught[0])
}
