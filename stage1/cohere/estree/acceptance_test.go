package estree

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/system-inc/adamic/internal/buildcache"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func acceptanceGrammar() []string {
	return []string{
		"public abstract class C {}", "({public x:1, async y:2, readonly z:3});", "function* f(){(yield);[yield];f(yield);}", "class implements {}", "try {} catch(e=1){}", "function f(){throw\nx;}", "tag?.`hello`;", "class C extends (A) implements (B) {}", "type T=typeof obj.#x;", "interface I extends A extends B {}", "for(using of=0;;){}",
	}
}
func TestAcceptanceGrammar(t *testing.T) {
	list := manifest(t, acceptanceGrammar())
	want := execute(t, "", goOracle(t), "--manifest", list)
	main, _ := filepath.Abs("main.ts")
	binary, script := build(t, main, true)
	for name, got := range map[string][]byte{"Node": onNode(t, main, "--manifest", list), "native": execute(t, "", binary, "--manifest", list), "emitted": onNode(t, script, "--manifest", list)} {
		if diff := firstDifference(want, got); diff != "" {
			t.Fatal(name + ": " + diff)
		}
	}
	t.Logf("%d acceptance grammar cases, %d identical canonical bytes", len(acceptanceGrammar()), len(want))
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
func TestAcceptanceMutants(t *testing.T) {
	cases, mutations := acceptanceGrammar(), acceptanceMutations()
	runMutantShards(t, cases, mutations, mutantShardPlan(t, cases, mutations, testAcceptanceMutantsShards))
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
