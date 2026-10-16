package json

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"

	"github.com/system-inc/adamic/stage1/cohere/json/oracles"
)

// An explicit repository-relative key keeps ownership stable as the corpus grows.
// _000 owns backend agreement and the Go oracle's live answer, _001 the sanitized run: ASan/UBSan with LeakSanitizer.
const jsonPortClassOrderKey = "cohere/internal/lint/rules/tailwind/collapse/testdata/classorder_fixtures.json"

// jsonPortInventoryKey is the corpus's largest case, 17.5 MB, about 17 s of one core's work across the five sides
// (its sanitized run 10.8 s, Golem, Oct 14). Hashed into _0760 with a 3 KB case, it was json#20/28's long pole, and
// under the unit's 8-way contention its own run reached 77.6 s of the 90 s (#21kptp1). One case doesn't split, so it
// runs as class order's does: alone in the unit, before the parallel shards resume (inventory_dedicated_test.go).
const jsonPortInventoryKey = "stage1/cohere/lint/inventory/inventory.json"

// jsonPortDedicatedKeys are the cases taken out of the hash buckets, each owned by tests of its own. Taking one out
// moves no other case: a bucket is chosen by the case's own key.
var jsonPortDedicatedKeys = []string{jsonPortClassOrderKey, jsonPortInventoryKey}

func jsonPortDedicatedPartition(all []textCase) (ordinary []textCase, shards []nativeChunk, dedicated []textCase) {
	for _, item := range all {
		if slices.Contains(jsonPortDedicatedKeys, item.Name) {
			dedicated = append(dedicated, item)
		} else {
			ordinary = append(ordinary, item)
		}
	}
	shards = jsonHashShards(ordinary, testPortMatchesGoCohereShards)
	return
}

// jsonPortDedicatedCase is the dedicated case named key as a list of one, or an empty list where the corpus lacks it.
func jsonPortDedicatedCase(dedicated []textCase, key string) []textCase {
	var items []textCase
	for _, item := range dedicated {
		if item.Name == key {
			items = append(items, item)
		}
	}
	return items
}

func jsonPortDedicatedUnion(all, ordinary []textCase, shards []nativeChunk, dedicated []textCase) error {
	named := map[string]bool{}
	for _, item := range dedicated {
		if named[item.Name] {
			return fmt.Errorf("duplicate dedicated case %s", item.Name)
		}
		named[item.Name] = true
	}
	combined := append(append([]textCase(nil), ordinary...), dedicated...)
	owners := append([]nativeChunk(nil), shards...)
	for index := range dedicated {
		owners = append(owners, nativeChunk{start: len(ordinary) + index, end: len(ordinary) + index + 1})
	}
	if err := jsonPortUnion(combined, owners); err != nil {
		return err
	}
	if len(combined) != len(all) {
		return fmt.Errorf("dedicated union has %d of %d cases", len(combined), len(all))
	}
	expected := make(map[string]textCase, len(all))
	for _, item := range all {
		if _, exists := expected[item.Name]; exists {
			return fmt.Errorf("duplicate input %s", item.Name)
		}
		expected[item.Name] = item
	}
	for i, item := range combined {
		if want, exists := expected[item.Name]; !exists || want != item {
			return fmt.Errorf("dedicated union changed %s", item.Name)
		}
		if i < len(ordinary) && slices.Contains(jsonPortDedicatedKeys, item.Name) {
			return fmt.Errorf("dedicated case %s retained in ordinary shard", item.Name)
		}
		if i >= len(ordinary) && !slices.Contains(jsonPortDedicatedKeys, item.Name) {
			return fmt.Errorf("unexpected dedicated case %s", item.Name)
		}
	}
	return nil
}

// jsonClassOrderAnswerInputs keys the Go oracle's answer to the class-order case by the oracle binary's bytes and the
// case's, so a different oracle or a changed fixture is a different product. The binary's digest is stricter than the
// oracle product's key: an oracle rebuilt to the same bytes keeps the answer, which is the same answer.
func jsonClassOrderAnswerInputs(oracle []byte, item textCase) buildcache.Inputs {
	return oracles.ClassOrderAnswerInputs(oracle, oracles.ComparisonCase(item))
}

// buildJSONClassOrderAnswer runs the Go oracle once over the class-order case, as _000 runs it live: the product holds
// the case it was asked (cases.txt) and its answer (answer.txt).

// jsonClassOrderAnswerProduct is the Go oracle's answer to the class-order case, built once and read by the two
// sanitizer passes, which compare against it instead of running the oracle and release again (#f42kzjr).
func jsonClassOrderAnswerProduct(t *testing.T) string { return oracles.ClassOrderAnswer(t) }

func TestProduct_JsonFindsClassOrderAnswer(t *testing.T) {
	t.Parallel()
	jsonClassOrderAnswerProduct(t)
}

// The answer's key moves with the oracle's bytes and with the case's, and with nothing else it was given.
func TestJSONClassOrderAnswerKey(t *testing.T) {
	t.Parallel()
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	key := func(oracle string, item textCase) string {
		declared, err := buildcache.Key(root, jsonClassOrderAnswerInputs([]byte(oracle), item))
		if err != nil {
			t.Fatal(err)
		}
		return declared
	}
	item := textCase{Name: jsonPortClassOrderKey, Text: `[{"classes":"b a"}]`}
	base := key("oracle binary", item)
	if again := key("oracle binary", item); again != base {
		t.Fatalf("the same oracle and case keyed %s and %s", base[:12], again[:12])
	}
	if moved := key("oracle binary rebuilt", item); moved == base {
		t.Fatalf("a different oracle binary kept the key %s", base[:12])
	}
	if moved := key("oracle binary", textCase{Name: item.Name, Text: item.Text + " "}); moved == base {
		t.Fatalf("a changed case kept the key %s", base[:12])
	}
	if moved := key("oracle binary", textCase{Name: "elsewhere.json", Text: item.Text}); moved == base {
		t.Fatalf("a renamed case kept the key %s", base[:12])
	}
}

func jsonPortClassOrderRun(t *testing.T, mode int) {
	t.Helper()
	setup := time.Now()
	portMatchesPrepare(t)
	all, _ := jsonTopCorpus(t)
	ordinary, shards, dedicated := jsonPortDedicatedPartition(all)
	if err := jsonPortDedicatedUnion(all, ordinary, shards, dedicated); err != nil {
		t.Fatal(err)
	}
	items := jsonPortDedicatedCase(dedicated, jsonPortClassOrderKey)
	if len(items) != 1 {
		t.Fatalf("dedicated fixture count %d, want 1", len(items))
	}
	stored := jsonClassOrderAnswerProduct(t)
	t.Logf("shared setup: %.3fs; exact union: %d ordinary + %d dedicated", time.Since(setup).Seconds(), len(ordinary), len(dedicated))
	deadline := portMatchesDeadline(t.Name())
	defer deadline.Stop()
	start := time.Now()
	input, _ := protocol(items, make([]answer, len(items)))
	path := filepath.Join(t.TempDir(), "cases.txt")
	if err := os.WriteFile(path, []byte(input), 0644); err != nil {
		t.Fatal(err)
	}
	// _000 runs the Go oracle live and holds the stored answer to it, so the sanitizer passes, which compare against
	// the stored answer, compare against the oracle. Release, Node and the JavaScript backend run here alone: the
	// sanitizer passes ran release against the same answer too, the same comparison twice more (#f42kzjr).
	expected := stored
	if mode == 0 {
		reference := execute(t, nil, portMatchesShared.oracle, "--cases", path)
		if reference.exitCode != 0 || len(reference.stderr) != 0 {
			t.Fatalf("Go oracle: exit %d stderr %s", reference.exitCode, reference.stderr)
		}
		expected = string(reference.stdout)
		if _, err := jsonPortAnswers(expected, len(items)); err != nil {
			t.Fatal(err)
		}
		if stored != expected {
			t.Fatal("the stored class-order Go answer differs from the Go oracle's live answer")
		}
		compare(t, "release", execute(t, nil, portMatchesShared.release, "--cases", path), expected, items)
		compare(t, "Node", onNode(t, portMatchesShared.entry, "--cases", path), expected, items)
		compare(t, "JavaScript backend", onNode(t, portMatchesShared.script, "--cases", path), expected, items)
	}
	if mode != 0 {
		// One sanitized run: with LeakSanitizer on (Linux), it reports every ASan/UBSan error a detect_leaks=0 run does,
		// plus leaks, so the separate ASan/UBSan test that ran the case again is gone (#fqvxk72). Off Linux, ASan/UBSan
		// runs without leak detection and macOS's leaks --atExit checks release.
		label, options := "ASan/UBSan/LeakSanitizer", "ASAN_OPTIONS=detect_leaks=1"
		if runtime.GOOS != "linux" {
			label, options = "ASan/UBSan", "ASAN_OPTIONS=detect_leaks=0"
			result := execute(t, nil, "leaks", "--atExit", "--", portMatchesShared.release, "--cases", path)
			if result.exitCode != 0 {
				t.Fatalf("leaks: %s", result.stderr)
			}
		}
		began := time.Now()
		compare(t, label, execute(t, []string{options}, portMatchesShared.sanitized, "--cases", path), expected, items)
		t.Logf("%s: %.3fs", label, time.Since(began).Seconds())
	}
	t.Logf("%s: own work %.3fs", jsonPortClassOrderKey, time.Since(start).Seconds())
}

// The class-order fixture is one indivisible case, and its sanitizer passes are about 28 s of one core's work (_001 28.8 s,
// _002 27.2 s alone on one core, Workshop Oct 11). Run in parallel beside the package's ASan shards, _002 passed its 90 s
// deadline on Home (landable-20, 20:41Z): contention, not work, and no split shortens one case. So the three run as
// sequential tests, which go test finishes before it resumes any parallel one, and each has the unit to itself.
//
// Each sanitizer pass is then one run of the case and nothing else, against the stored Go answer (#f42kzjr). At an
// 8-CPU share on Cloud (Oct 12), LeakSanitizer is 19.4 s and ASan/UBSan 22.4 s, but _001 was 27.5 s and _002 35 to
// 37 s with the oracle and release run again around them; Home runs the same units about 2.1 times slower, where
// _001 reached 61.5 s. Since then _002 is gone: _001's one detect_leaks=1 run is its ASan/UBSan run too (#fqvxk72).

// Not parallel: runs alone in the unit, before the parallel shards resume, so the indivisible fixture has its CPUs.
func TestPortMatchesGoCohereClassOrder_000(t *testing.T) {
	jsonPortClassOrderRun(t, 0)
}

// Not parallel: runs alone in the unit, before the parallel shards resume, so the indivisible fixture has its CPUs.
func TestPortMatchesGoCohereClassOrder_001(t *testing.T) {
	jsonPortClassOrderRun(t, 1)
}

// Every dedicated case leaves the hash buckets and is owned exactly once: class order by _000 and _001, inventory.json
// by TestPortMatchesGoCohereInventory. A planted disagreement in either is caught by exactly one owner.
func TestJSONDedicatedShardOwnership(t *testing.T) {
	t.Parallel()
	for _, group := range [][]textCase{nil, {{Name: "small.json", Text: "{}"}}, {{Name: jsonPortClassOrderKey, Text: "[]"}}, {{Name: jsonPortInventoryKey, Text: "{}"}}} {
		ordinary, shards, dedicated := jsonPortDedicatedPartition(group)
		if err := jsonPortDedicatedUnion(group, ordinary, shards, dedicated); err != nil {
			t.Fatal(err)
		}
	}
	all := []textCase{{Name: "small.json", Text: "{}"}, {Name: jsonPortClassOrderKey, Text: "[]"}, {Name: jsonPortInventoryKey, Text: "{}"}}
	ordinary, shards, dedicated := jsonPortDedicatedPartition(all)
	if err := jsonPortDedicatedUnion(all, ordinary, shards, dedicated); err != nil {
		t.Fatal(err)
	}
	if len(dedicated) != 2 || len(ordinary) != 1 {
		t.Fatal("wrong dedicated ownership")
	}
	owner := jsonCaseShard("small.json", testPortMatchesGoCohereShards)
	if shards[owner].end-shards[owner].start != 1 || ordinary[shards[owner].start].Name != "small.json" {
		t.Fatal("ordinary hash owner changed")
	}
	if err := jsonPortDedicatedUnion(all, ordinary, shards, nil); err == nil {
		t.Fatal("missing dedicated cases survived")
	}
	if err := jsonPortDedicatedUnion(all, ordinary, shards, dedicated[:1]); err == nil {
		t.Fatal("a missing dedicated case survived")
	}
	if err := jsonPortDedicatedUnion(all, ordinary, shards, append(dedicated, dedicated[0])); err == nil {
		t.Fatal("duplicate dedicated case survived")
	}
	// Agreement ownership includes each dedicated case exactly once: class order's _000 runs the Go oracle, release,
	// Node and the JavaScript backend, and its sanitizer pass compares against the stored answer; inventory.json's one
	// test runs all five sides.
	owners := [][]textCase{}
	for _, shard := range shards {
		owners = append(owners, ordinary[shard.start:shard.end])
	}
	for _, key := range jsonPortDedicatedKeys {
		owners = append(owners, jsonPortDedicatedCase(dedicated, key))
	}
	for _, key := range jsonPortDedicatedKeys {
		caught := 0
		for _, items := range owners {
			answers := make([]answer, len(items))
			_, want := protocol(items, answers)
			for i, item := range items {
				if item.Name == key {
					answers[i].Output = "planted disagreement"
				}
			}
			_, got := protocol(items, answers)
			if comparisonError("planted", run{stdout: []byte(got)}, want, items) != nil {
				caught++
			}
		}
		if caught != 1 {
			t.Fatalf("planted disagreement in %s caught by %d owners", key, caught)
		}
	}
	classOrder := jsonPortDedicatedCase(dedicated, jsonPortClassOrderKey)
	for child := 0; child < 2; child++ {
		answers := make([]answer, len(classOrder))
		_, want := protocol(classOrder, answers)
		answers[0].Output = "planted disagreement"
		_, got := protocol(classOrder, answers)
		if comparisonError(fmt.Sprintf("child-%03d", child), run{stdout: []byte(got)}, want, classOrder) == nil {
			t.Fatalf("child %d missed planted failure", child)
		}
	}
	declared := map[string]bool{}
	for _, file := range []string{"class_order_grain30_test.go", "inventory_dedicated_test.go"} {
		tree, err := parser.ParseFile(token.NewFileSet(), file, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, decl := range tree.Decls {
			if f, ok := decl.(*ast.FuncDecl); ok && (strings.HasPrefix(f.Name.Name, "TestPortMatchesGoCohereClassOrder_") || strings.HasPrefix(f.Name.Name, "TestPortMatchesGoCohereInventory")) {
				declared[f.Name.Name] = true
			}
		}
	}
	if len(declared) != 3 || !declared["TestPortMatchesGoCohereClassOrder_000"] || !declared["TestPortMatchesGoCohereClassOrder_001"] || !declared["TestPortMatchesGoCohereInventory"] {
		t.Fatalf("dedicated shard declarations: %v", declared)
	}
	t.Log("dedicated union intact; a planted disagreement in each dedicated case caught by exactly one owner")
}
