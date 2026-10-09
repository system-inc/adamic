package estree

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/system-inc/adamic/internal/buildcache"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func acceptanceGrammar() []string {
	return []string{
		"public abstract class C {}", "({public x:1, async y:2, readonly z:3});", "function* f(){(yield);[yield];f(yield);}", "class implements {}", "try {} catch(e=1){}", "function f(){throw\nx;}", "tag?.`hello`;", "class C extends (A) implements (B) {}", "type T=typeof obj.#x;", "interface I extends A extends B {}", "for(using of=0;;){}",
	}
}

const testAcceptanceGrammarShards = 3

// TestAcceptanceGrammar covers all eleven inputs on source Node, sanitized
// native and emitted JS. ADAMIC_TEST_SHARD=i/n selects zero-based shard indices
// modulo n; unset runs all. The gate selects TestAcceptanceGrammar/shard-NNN.
func TestAcceptanceGrammar(t *testing.T) {
	cases := acceptanceGrammar()
	runAgreementShards(t, cases, agreementShardPlan(t, cases, 4, testAcceptanceGrammarShards))
}

func TestAcceptanceGrammarShardProof(t *testing.T) {
	cases := acceptanceGrammar()
	proveAgreementShard(t, cases, agreementShardPlan(t, cases, 4, testAcceptanceGrammarShards), 4)
}

const testAcceptanceMutantsShards = 2

func acceptanceMutations() []portMutation {
	return []portMutation{
		{"catch-initializer", "convert.ts", "this.separated(this.child(declaration, 0), this.child(declaration, 1), 'ColonToken')", "true", "try {} catch(e=1){}"},
		{"class-keyword-name", "sourceStatements.ts", "!(this.parser.peek() === 'Identifier' || this.parser.peek().endsWith('Keyword'))", "false", "class implements {}"},
	}
}

// TestAcceptanceMutants runs all 22 mutant/case pairs on source Node and
// sanitized native. ADAMIC_TEST_SHARD=i/n selects zero-based shard indices
// modulo n; unset runs both. Shared products are prepared once per invocation.
func TestAcceptanceMutantsUnion(t *testing.T) {
	cases, mutations := acceptanceGrammar(), acceptanceMutations()
	mutantShardPlan(t, cases, mutations, testAcceptanceMutantsShards)
}

func TestAcceptanceMutantShardProof(t *testing.T) {
	cases, mutations := acceptanceGrammar(), acceptanceMutations()
	proveMutantShards(t, mutations, mutantShardPlan(t, cases, mutations, testAcceptanceMutantsShards))
}

func TestAcceptanceDiagnostics(t *testing.T) {
	sources := []string{"++await 42;", "++delete foo.bar;", "--ANY1--;", "type T = A | () => B;", "new obj?.member();", "import { 'a' } from 'm';", "import A from 'm' assert {type:'json'};", "super<T>();"}
	list := manifest(t, sources)
	statuses := string(acceptanceAudit(t, sources))
	if strings.Count(statuses, `"status":"error"`) != len(sources) {
		t.Fatal(statuses)
	}
	main, _ := filepath.Abs("main.ts")
	binary, script := build(t, main, true)
	paths, err := os.ReadFile(list)
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range strings.Fields(string(paths)) {
		for _, argv := range [][]string{{"node", "--disable-warning=ExperimentalWarning", filepath.Join(root(t), "oracle/node.mjs"), main, path}, {binary, path}, {"node", "--disable-warning=ExperimentalWarning", filepath.Join(root(t), "oracle/node.mjs"), script, path}} {
			refusedBeforeDeadline(t, argv, "ESTree parser")
		}
	}
	t.Logf("%d Go refusals explicitly refused before deadline on all builds", len(sources))
}
func TestAcceptanceDiagnosticControl(t *testing.T) {
	list := manifest(t, []string{"++await 42;"})
	statuses := string(acceptanceAudit(t, []string{"++await 42;"}))
	if !strings.Contains(statuses, `"status":"error"`) {
		t.Fatal(statuses)
	}
	main := mutantPort(t, "pipeline.ts", "if(syntax !== '')", "if(false)")
	binary, _ := build(t, main, true)
	for name, got := range map[string][]byte{"Node": onNode(t, main, "--manifest", list), "native": execute(t, "", binary, "--manifest", list)} {
		if !strings.Contains(string(got), "0 Program ") {
			t.Fatal(name + " control did not accept")
		}
		t.Log(name + ": disabled syntax validation accepts Go-refused update operand; acceptance check catches it")
	}
}

// acceptanceAudit caches both status records and any canonical answer files.
// Paths in stored records are relative placeholders so a fetched product stays
// read-only and its answer paths are resolved in the receiving directory.
func acceptanceAudit(t *testing.T, sources []string) []byte {
	t.Helper()
	oracle := goOracle(t)
	binary, err := os.ReadFile(oracle)
	if err != nil {
		t.Fatal(err)
	}
	inputs := buildcache.Inputs{Name: "estree-acceptance-audit", Files: []string{"stage1/cohere/estree/acceptance_test.go"}, Flags: []string{fmt.Sprintf("oracle-sha256=%x", sha256.Sum256(binary)), "--audit"}}
	for i, source := range sources {
		inputs.Flags = append(inputs.Flags, fmt.Sprintf("case-%03d=%x", i, sha256.Sum256([]byte(source))))
	}
	dir := buildcache.Product(t, inputs, func(dir string) error {
		var listing strings.Builder
		for i, text := range sources {
			path := filepath.Join(dir, fmt.Sprintf("%04d.ts", i))
			if err := os.WriteFile(path, []byte(text), 0644); err != nil {
				return err
			}
			listing.WriteString(path + "\n")
		}
		manifest := filepath.Join(dir, "manifest")
		if err := os.WriteFile(manifest, []byte(listing.String()), 0644); err != nil {
			return err
		}
		output := executeUncached(t, "", oracle, "--audit", manifest, filepath.Join(dir, "answers"))
		// JSON escapes paths, so normalize decoded strings rather than bytes.
		var normalized strings.Builder
		decoder := json.NewDecoder(strings.NewReader(string(output)))
		for range sources {
			record := map[string]any{}
			if err := decoder.Decode(&record); err != nil {
				return err
			}
			for key, value := range record {
				if text, ok := value.(string); ok {
					record[key] = strings.ReplaceAll(text, dir, "${ESTREE_PRODUCT}")
				}
			}
			encoded, err := json.Marshal(record)
			if err != nil {
				return err
			}
			normalized.Write(encoded)
			normalized.WriteByte('\n')
		}
		var extra any
		if err := decoder.Decode(&extra); err != io.EOF {
			return fmt.Errorf("unexpected trailing audit record: %v", err)
		}
		// The input manifest is needed only during the build; its absolute paths
		// must not make an otherwise identical product fail the store audit.
		if err := os.Remove(manifest); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(dir, "statuses"), []byte(normalized.String()), 0644)
	})
	data, err := os.ReadFile(filepath.Join(dir, "statuses"))
	if err != nil {
		t.Fatal(err)
	}
	var output strings.Builder
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	for range sources {
		record := map[string]any{}
		if err := decoder.Decode(&record); err != nil {
			t.Fatal(err)
		}
		for key, value := range record {
			if text, ok := value.(string); ok {
				record[key] = strings.ReplaceAll(text, "${ESTREE_PRODUCT}", dir)
			}
		}
		encoded, err := json.Marshal(record)
		if err != nil {
			t.Fatal(err)
		}
		output.Write(encoded)
		output.WriteByte('\n')
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		t.Fatalf("unexpected trailing cached audit record: %v", err)
	}
	return []byte(output.String())
}

func TestAcceptanceOracleProducts(t *testing.T) {
	sources := []string{"let x=1;", "++await 42;"}
	want := executeUncached(t, "", goOracle(t), "--manifest", manifest(t, sources[:1]))
	for range 2 {
		decoder := json.NewDecoder(strings.NewReader(string(acceptanceAudit(t, sources))))
		for i := range sources {
			var record struct{ Status, Answer string }
			if err := decoder.Decode(&record); err != nil {
				t.Fatal(err)
			}
			if i == 0 {
				if record.Status != "ok" || record.Answer == "" {
					t.Fatalf("accepted oracle input lost its answer: %+v", record)
				}
				answer, err := os.ReadFile(record.Answer)
				if err != nil {
					t.Fatal(err)
				}
				if diff := firstDifference(want, answer); diff != "" {
					t.Fatal("cached canonical answer: " + diff)
				}
			} else if record.Status != "error" || record.Answer != "" {
				t.Fatalf("refused oracle input changed: %+v", record)
			}
		}
	}
}

type agreementShard struct {
	name  string
	cases []int
}

func agreementShardPlan(t *testing.T, cases []string, perShard, declared int) []agreementShard {
	t.Helper()
	if perShard < 1 || declared < 1 || declared > 1000 {
		t.Fatal("invalid agreement shard dimensions")
	}
	var shards []agreementShard
	for begin := 0; begin < len(cases); begin += perShard {
		shard := agreementShard{name: fmt.Sprintf("shard-%03d", len(shards))}
		for id := begin; id < begin+perShard && id < len(cases); id++ {
			shard.cases = append(shard.cases, id)
		}
		shards = append(shards, shard)
	}
	if len(shards) != declared {
		t.Fatalf("enumerated %d shards, declared %d", len(shards), declared)
	}
	// The unsplit enumeration's ids are exactly [0,len(cases)). Check both
	// cardinality and set equality, independently of selection by -run or env.
	seen := make(map[int]bool, len(cases))
	total := 0
	for index, shard := range shards {
		if shard.name != fmt.Sprintf("shard-%03d", index) {
			t.Fatalf("unstable shard name %q", shard.name)
		}
		for _, id := range shard.cases {
			if id < 0 || id >= len(cases) || seen[id] {
				t.Fatalf("invalid or repeated case %03d in %s", id, shard.name)
			}
			seen[id] = true
			total++
		}
	}
	if total != len(cases) || len(seen) != len(cases) {
		t.Fatalf("union has %d cases, %d ids; want %d", total, len(seen), len(cases))
	}
	for id := range cases {
		if !seen[id] {
			t.Fatalf("union missing case %03d", id)
		}
	}
	t.Logf("exact union: %d cases, %d unique ids, %d shards", total, len(seen), len(shards))
	return shards
}

func checkAgreementCase(id int, want, source, native, emitted []byte) error {
	for _, side := range []struct {
		name string
		got  []byte
	}{{"source Node", source}, {"sanitized native", native}, {"emitted JS", emitted}} {
		if diff := firstDifference(want, side.got); diff != "" {
			return fmt.Errorf("case %03d %s: %s", id, side.name, diff)
		}
	}
	return nil
}

func runAgreementShards(t *testing.T, cases []string, shards []agreementShard) {
	t.Helper()
	start := time.Now()
	main, err := filepath.Abs("main.ts")
	if err != nil {
		t.Fatal(err)
	}
	oracle, binary, script := "", "", ""
	needed := false
	setup := time.Duration(0)
	t.Cleanup(func() { t.Logf("setup excluding parallel shard logic: %.3fs", setup.Seconds()) })
	for index, shard := range shards {
		if !selectedShard(t, index) {
			continue
		}
		t.Run(shard.name, func(t *testing.T) {
			needed = true
			t.Parallel()
			texts := make([]string, len(shard.cases))
			for i, id := range shard.cases {
				texts[i] = cases[id]
			}
			list := manifest(t, texts)
			want := canonicalRecords(t, execute(t, "", oracle, "--manifest", list), len(texts))
			source := canonicalRecords(t, onNode(t, main, "--manifest", list), len(texts))
			native := canonicalRecords(t, execute(t, "", binary, "--manifest", list), len(texts))
			emitted := canonicalRecords(t, onNode(t, script, "--manifest", list), len(texts))
			for i, id := range shard.cases {
				if err := checkAgreementCase(id, want[i], source[i], native[i], emitted[i]); err != nil {
					t.Fatal(err)
				}
			}
			t.Logf("cases [%d,%d): %d inputs; Go/source Node/sanitized native/emitted JS checked", shard.cases[0], shard.cases[len(shard.cases)-1]+1, len(texts))
		})
	}
	if needed {
		oracle = goOracle(t)
		binary, script = build(t, main, true)
	}
	setup = time.Since(start)
}

func proveAgreementShard(t *testing.T, cases []string, shards []agreementShard, planted int) {
	t.Helper()
	parent := t.Name()
	if value := os.Getenv("ADAMIC_ESTREE_PLANTED_AGREEMENT"); value != "" {
		id, err := strconv.Atoi(value)
		if err != nil || id < 0 || id >= len(cases) {
			t.Fatal("invalid planted agreement case")
		}
		for _, shard := range shards {
			t.Run(shard.name, func(t *testing.T) {
				t.Parallel()
				for _, caseID := range shard.cases {
					want := []byte("oracle")
					emitted := want
					if caseID == id {
						emitted = []byte("planted disagreement")
					}
					if err := checkAgreementCase(caseID, want, want, want, emitted); err != nil {
						t.Fatal(err)
					}
				}
			})
		}
		return
	}
	expected := ""
	for _, shard := range shards {
		for _, id := range shard.cases {
			if id == planted {
				expected = shard.name
			}
		}
	}
	if expected == "" {
		t.Fatal("planted case absent from union")
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	command := exec.Command(executable, "-test.run=^"+parent+"$", "-test.v", "-test.count=1")
	command.Env = append(os.Environ(), fmt.Sprintf("ADAMIC_ESTREE_PLANTED_AGREEMENT=%d", planted))
	output, err := command.CombinedOutput()
	exit, ok := err.(*exec.ExitError)
	if !ok || exit.ExitCode() != 1 {
		t.Fatalf("planted failure: exit=%v; %s", err, output)
	}
	failed, passed := 0, 0
	for _, line := range strings.Split(string(output), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "--- FAIL: "+parent+"/shard-") {
			failed++
			if !strings.HasPrefix(line, "--- FAIL: "+parent+"/"+expected+" (") {
				t.Fatalf("wrong shard caught failure: %s", line)
			}
		}
		if strings.HasPrefix(line, "--- PASS: "+parent+"/shard-") {
			passed++
		}
	}
	diagnostic := fmt.Sprintf("case %03d emitted JS:", planted)
	if failed != 1 || passed != len(shards)-1 || !strings.Contains(string(output), diagnostic) {
		t.Fatalf("planted failure: failed=%d passed=%d; want 1/%d and %q; %s", failed, passed, len(shards)-1, diagnostic, output)
	}
	t.Logf("planted emitted-JS disagreement case %03d caught by %s; exactly one shard failed, %d passed", planted, expected, passed)
}
