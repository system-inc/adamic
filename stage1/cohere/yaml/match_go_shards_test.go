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
	cases     func(test *testing.T) (string, int, int)
	oracle    func(test *testing.T) string
	arguments []string
	build     yamlBuild
	entry     string
}

// corpus is the family's whole corpus, one encoded case a line, and its count of repository files.
func (comparison yamlMatchGo) corpus(test *testing.T) ([]string, int) {
	test.Helper()
	path, files, count := comparison.cases(test)
	data, readError := os.ReadFile(path)
	if readError != nil {
		test.Fatal(readError)
	}
	cases := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
	if len(cases) != count {
		test.Fatalf("enumerated %d cases; the generator counted %d", len(cases), count)
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
func (comparison yamlMatchGo) run(test *testing.T, shard int) yamlShardRun {
	test.Helper()
	oracle := comparison.oracle(test)
	binary := comparison.build.binary(test, comparison.entry)
	emitted := comparison.build.emitted(test, comparison.entry)
	all, files := comparison.corpus(test)
	indices := yamlShards(all, comparison.shards)[shard]
	if len(indices) == 0 {
		test.Fatal("empty shard")
	}
	cases := make([]string, len(indices))
	for i, index := range indices {
		cases[i] = all[index]
	}
	path := filepath.Join(test.TempDir(), comparison.family+"-cases.txt")
	if writeError := os.WriteFile(path, []byte(strings.Join(cases, "\n")+"\n"), 0644); writeError != nil {
		test.Fatal(writeError)
	}
	root, absolutePathError := filepath.Abs(repository)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	entry, absolutePathError := filepath.Abs(comparison.entry)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	runner := filepath.Join(root, "oracle/node.mjs")
	arguments := append(append([]string(nil), comparison.arguments...), path)
	result := yamlShardRun{all: all, indices: indices, files: files, path: path, entry: entry}
	result.expected = run(test, "", nil, oracle, arguments...)
	result.native = run(test, "", []string{"ASAN_OPTIONS=detect_leaks=1"}, binary, arguments...)
	result.node = run(
		test,
		"",
		nil,
		"node",
		append([]string{"--disable-warning=ExperimentalWarning", runner, entry}, arguments...)...,
	)
	result.emitted = run(
		test,
		"",
		nil,
		"node",
		append([]string{"--disable-warning=ExperimentalWarning", runner, emitted}, arguments...)...,
	)
	test.Logf("%s shard %03d: %d of %d cases", comparison.family, shard, len(indices), len(all))
	return result
}

// diagnostic is where a failing shard leaves its bytes, one name per family, shard and side.
func (comparison yamlMatchGo) diagnostic(shard int, side string) string {
	return fmt.Sprintf("/tmp/stage1-yaml-%s-%03d-%s.txt", comparison.family, shard, side)
}

// union holds every case to exactly one shard.
func (comparison yamlMatchGo) union(test *testing.T) []string {
	test.Helper()
	cases, _ := comparison.corpus(test)
	visits := make([]int, len(cases))
	for _, indices := range yamlShards(cases, comparison.shards) {
		for _, index := range indices {
			visits[index]++
		}
	}
	for index, count := range visits {
		if count != 1 {
			test.Fatalf("case %d visited %d times", index, count)
		}
	}
	test.Logf(
		"counted union: all %d %s cases visited exactly once over %d shards",
		len(cases),
		comparison.family,
		comparison.shards,
	)
	return cases
}

// planted runs the first case through Go and native, plants a byte in native's real stdout, and requires exactly one
// shard to own the disagreement.
func (comparison yamlMatchGo) planted(test *testing.T) {
	test.Helper()
	oracle := comparison.oracle(test)
	binary := comparison.build.binary(test, comparison.entry)
	cases, _ := comparison.corpus(test)
	path := filepath.Join(test.TempDir(), "cases.txt")
	if writeError := os.WriteFile(path, []byte(cases[0]+"\n"), 0644); writeError != nil {
		test.Fatal(writeError)
	}
	arguments := append(append([]string(nil), comparison.arguments...), path)
	expected := run(test, "", nil, oracle, arguments...)
	actual := run(test, "", []string{"ASAN_OPTIONS=detect_leaks=1"}, binary, arguments...)
	if !bytes.Equal(actual, expected) {
		test.Fatal("native control disagrees before planting")
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
		test.Fatalf("planted stdout disagreement caught by %v; want one owner", caught)
	}
	test.Logf("planted real native stdout disagreement caught by the %s shard %03d only", comparison.family, caught[0])
}
