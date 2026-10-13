package json

import (
	"crypto/sha256"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
)

// An explicit repository-relative key keeps ownership stable as the corpus grows.
// _000 owns backend agreement and the Go oracle's live answer, _001 the leak pass, and _002 ASan/UBSan.
const jsonPortClassOrderKey = "cohere/internal/lint/rules/tailwind/collapse/testdata/classorder_fixtures.json"

func jsonPortClassOrderPartition(all []textCase) (ordinary []textCase, shards []nativeChunk, dedicated []textCase) {
	for _, item := range all {
		if item.Name == jsonPortClassOrderKey {
			dedicated = append(dedicated, item)
		} else {
			ordinary = append(ordinary, item)
		}
	}
	shards = jsonHashShards(ordinary, testPortMatchesGoCohereShards)
	return
}

func jsonPortClassOrderUnion(all, ordinary []textCase, shards []nativeChunk, dedicated []textCase) error {
	if len(dedicated) > 1 {
		return fmt.Errorf("duplicate dedicated class-order case")
	}
	combined := append(append([]textCase(nil), ordinary...), dedicated...)
	owners := append([]nativeChunk(nil), shards...)
	owners = append(owners, nativeChunk{start: len(ordinary), end: len(combined)})
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
		if i < len(ordinary) && item.Name == jsonPortClassOrderKey {
			return fmt.Errorf("dedicated case retained in ordinary shard")
		}
		if i >= len(ordinary) && item.Name != jsonPortClassOrderKey {
			return fmt.Errorf("unexpected dedicated case %s", item.Name)
		}
	}
	return nil
}

// jsonClassOrderAnswerInputs keys the Go oracle's answer to the class-order case by the oracle binary's bytes and the
// case's, so a different oracle or a changed fixture is a different product. The binary's digest is stricter than the
// oracle product's key: an oracle rebuilt to the same bytes keeps the answer, which is the same answer.
func jsonClassOrderAnswerInputs(oracle []byte, item textCase) buildcache.Inputs {
	return buildcache.Inputs{
		Name:  "json class-order Go answer",
		Files: []string{"stage1/cohere/json/class_order_grain30_test.go", "stage1/cohere/json/port_test.go"},
		Flags: []string{fmt.Sprintf("Go oracle binary sha256 %x", sha256.Sum256(oracle)), fmt.Sprintf("case %s sha256 %x", item.Name, sha256.Sum256([]byte(item.Text))), "--cases"},
	}
}

// buildJSONClassOrderAnswer runs the Go oracle once over the class-order case, as _000 runs it live: the product holds
// the case it was asked (cases.txt) and its answer (answer.txt).
func buildJSONClassOrderAnswer(oracle string, item textCase) func(string) error {
	return func(directory string) error {
		input, _ := protocol([]textCase{item}, make([]answer, 1))
		cases := filepath.Join(directory, "cases.txt")
		if err := os.WriteFile(cases, []byte(input), 0644); err != nil {
			return err
		}
		reference, err := executeResult(nil, oracle, "--cases", cases)
		if err != nil {
			return err
		}
		if reference.exitCode != 0 || len(reference.stderr) != 0 {
			return fmt.Errorf("Go oracle: exit %d stderr %s", reference.exitCode, reference.stderr)
		}
		if _, err := jsonPortAnswers(string(reference.stdout), 1); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(directory, "answer.txt"), reference.stdout, 0644)
	}
}

var jsonClassOrderAnswer struct {
	once   sync.Once
	answer string
}

// jsonClassOrderAnswerProduct is the Go oracle's answer to the class-order case, built once and read by the two
// sanitizer passes, which compare against it instead of running the oracle and release again (#f42kzjr).
func jsonClassOrderAnswerProduct(t *testing.T) string {
	t.Helper()
	jsonClassOrderAnswer.once.Do(func() {
		oracle := portMatchesProduct(t, "oracle")
		binary, err := os.ReadFile(oracle)
		if err != nil {
			t.Fatal(err)
		}
		all, _ := jsonTopCorpus(t)
		_, _, items := jsonPortClassOrderPartition(all)
		if len(items) != 1 {
			t.Fatalf("dedicated fixture count %d, want 1", len(items))
		}
		directory := buildcache.Product(t, jsonClassOrderAnswerInputs(binary, items[0]), buildJSONClassOrderAnswer(oracle, items[0]))
		stored, err := os.ReadFile(filepath.Join(directory, "answer.txt"))
		if err != nil {
			t.Fatal(err)
		}
		jsonClassOrderAnswer.answer = string(stored)
	})
	if jsonClassOrderAnswer.answer == "" {
		t.Fatal("the class-order Go answer failed earlier in this process")
	}
	return jsonClassOrderAnswer.answer
}

func TestProduct_JSONClassOrderAnswer(t *testing.T) {
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
	leak := mode == 1
	setup := time.Now()
	portMatchesPrepare(t)
	all, _ := jsonTopCorpus(t)
	ordinary, shards, items := jsonPortClassOrderPartition(all)
	if err := jsonPortClassOrderUnion(all, ordinary, shards, items); err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("dedicated fixture count %d, want 1", len(items))
	}
	stored := jsonClassOrderAnswerProduct(t)
	t.Logf("shared setup: %.3fs; exact union: %d ordinary + %d dedicated", time.Since(setup).Seconds(), len(ordinary), len(items))
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
		label := "ASan/UBSan"
		options := "ASAN_OPTIONS=detect_leaks=0"
		if leak {
			label = "LeakSanitizer"
			options = "ASAN_OPTIONS=detect_leaks=1"
		}
		if leak && runtime.GOOS != "linux" {
			result := execute(t, nil, "leaks", "--atExit", "--", portMatchesShared.release, "--cases", path)
			if result.exitCode != 0 {
				t.Fatalf("leaks: %s", result.stderr)
			}
		} else {
			began := time.Now()
			compare(t, label, execute(t, []string{options}, portMatchesShared.sanitized, "--cases", path), expected, items)
			t.Logf("%s: %.3fs", label, time.Since(began).Seconds())
		}
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
// _001 reached 61.5 s.

// Not parallel: runs alone in the unit, before the parallel shards resume, so the indivisible fixture has its CPUs.
func TestPortMatchesGoCohereClassOrder_000(t *testing.T) {
	jsonPortClassOrderRun(t, 0)
}

// Not parallel: runs alone in the unit, before the parallel shards resume, so the indivisible fixture has its CPUs.
func TestPortMatchesGoCohereClassOrder_001(t *testing.T) {
	jsonPortClassOrderRun(t, 1)
}

// Not parallel: runs alone in the unit, before the parallel shards resume, so the indivisible fixture has its CPUs.
func TestPortMatchesGoCohereClassOrder_002(t *testing.T) {
	jsonPortClassOrderRun(t, 2)
}

func TestJSONClassOrderShardOwnership(t *testing.T) {
	t.Parallel()
	for _, group := range [][]textCase{nil, {{Name: "small.json", Text: "{}"}}, {{Name: jsonPortClassOrderKey, Text: "[]"}}} {
		ordinary, shards, dedicated := jsonPortClassOrderPartition(group)
		if err := jsonPortClassOrderUnion(group, ordinary, shards, dedicated); err != nil {
			t.Fatal(err)
		}
	}
	all := []textCase{{Name: "small.json", Text: "{}"}, {Name: jsonPortClassOrderKey, Text: "[]"}}
	ordinary, shards, dedicated := jsonPortClassOrderPartition(all)
	if err := jsonPortClassOrderUnion(all, ordinary, shards, dedicated); err != nil {
		t.Fatal(err)
	}
	if len(dedicated) != 1 || len(ordinary) != 1 {
		t.Fatal("wrong dedicated ownership")
	}
	owner := jsonCaseShard("small.json", testPortMatchesGoCohereShards)
	if shards[owner].end-shards[owner].start != 1 || ordinary[shards[owner].start].Name != "small.json" {
		t.Fatal("ordinary hash owner changed")
	}
	if err := jsonPortClassOrderUnion(all, ordinary, shards, nil); err == nil {
		t.Fatal("missing dedicated case survived")
	}
	if err := jsonPortClassOrderUnion(all, ordinary, shards, append(dedicated, dedicated...)); err == nil {
		t.Fatal("duplicate dedicated case survived")
	}
	// Agreement ownership includes the dedicated case exactly once: _000 runs the Go oracle, release, Node and the
	// JavaScript backend, and the two sanitizer passes compare against its stored answer.
	owners := [][]textCase{}
	for _, shard := range shards {
		owners = append(owners, ordinary[shard.start:shard.end])
	}
	owners = append(owners, dedicated)
	caught := 0
	for _, items := range owners {
		answers := make([]answer, len(items))
		_, want := protocol(items, answers)
		for i, item := range items {
			if item.Name == jsonPortClassOrderKey {
				answers[i].Output = "planted disagreement"
			}
		}
		_, got := protocol(items, answers)
		if comparisonError("planted", run{stdout: []byte(got)}, want, items) != nil {
			caught++
		}
	}
	if caught != 1 {
		t.Fatalf("planted disagreement caught by %d owners", caught)
	}
	for child := 0; child < 3; child++ {
		answers := make([]answer, len(dedicated))
		_, want := protocol(dedicated, answers)
		answers[0].Output = "planted disagreement"
		_, got := protocol(dedicated, answers)
		if comparisonError(fmt.Sprintf("child-%03d", child), run{stdout: []byte(got)}, want, dedicated) == nil {
			t.Fatalf("child %d missed planted failure", child)
		}
	}
	tree, err := parser.ParseFile(token.NewFileSet(), "class_order_grain30_test.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	declared := map[string]bool{}
	for _, decl := range tree.Decls {
		if f, ok := decl.(*ast.FuncDecl); ok && strings.HasPrefix(f.Name.Name, "TestPortMatchesGoCohereClassOrder_") {
			declared[f.Name.Name] = true
		}
	}
	if len(declared) != 3 || !declared["TestPortMatchesGoCohereClassOrder_002"] || !declared["TestPortMatchesGoCohereClassOrder_000"] || !declared["TestPortMatchesGoCohereClassOrder_001"] {
		t.Fatalf("dedicated shard declarations: %v", declared)
	}
	t.Log("dedicated union intact; planted disagreement caught by exactly one agreement shard")
}
