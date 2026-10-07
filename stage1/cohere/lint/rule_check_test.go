package lint

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/system-inc/adamic/stage1/cohere/lint/registry"
)

var ruleSemanticChange = flag.Bool("rule-semantic-change", false, "benchmark one message-ID token edit in a private snapshot")

var ruleByteChange = flag.Bool("rule-byte-change", false, "benchmark a private snapshot with one extra space in the selected module")

var ruleSlug = flag.String("rule", "", "check one registered lint rule")

func selectedDescriptor(t *testing.T, slug string) registry.Descriptor {
	t.Helper()
	for _, d := range prepareRegistry(t, ".") {
		if d.Slug == slug {
			return d
		}
	}
	t.Fatalf("unknown registered rule %q", slug)
	return registry.Descriptor{}
}
func selectedPort(t *testing.T, d registry.Descriptor) string {
	return selectedPortMutation(t, d, "", "", "")
}
func selectedPortMutation(t *testing.T, d registry.Descriptor, from, to, target string) string {
	t.Helper()
	directory := copyPort(t, t.TempDir(), from, to, target)
	if err := os.WriteFile(filepath.Join(directory, ".selected-rule"), []byte(d.Slug), 0600); err != nil {
		t.Fatal(err)
	}
	if *ruleByteChange {
		path := filepath.Join(directory, "rules", d.Slug, d.Module)
		if err := os.WriteFile(path, []byte(lintBytes(t, path)+" "), 0644); err != nil {
			t.Fatal(err)
		}
	}
	if *ruleSemanticChange {
		applySemanticEdit(t, directory, d)
	}
	prepareRegistry(t, directory)
	return directory
}

// Every inherited source and all-rule option variant remains; explicitly named
// variants of another rule belong to that rule's check.
func inheritedRuleRows(rows []string, name string) []string {
	var result []string
	for _, row := range rows {
		fields := strings.Split(row, "\t")
		if len(fields) < 2 || fields[1] == "" || fields[1] == "all" || fields[1] == name {
			result = append(result, row)
		}
	}
	return result
}

func selectedRows(rows []string, name string) []string {
	var selected []string
	for _, row := range rows {
		fields := strings.Split(row, "\t")
		if len(fields) == 1 {
			fields = append(fields, name)
		} else if fields[1] == "" || fields[1] == "all" {
			fields[1] = name
		} else if fields[1] != name {
			continue
		}
		selected = append(selected, strings.Join(fields, "\t"))
	}
	return selected
}

// Inputs are always rewritten, including in uncached mode. These are immutable
// content-addressed corpus snapshots, not cached observations. Stable filenames
// preserve the exact byte protocol, whose diagnostics contain source filenames.
func stableRows(t *testing.T, rows []string) []string {
	var result []string
	for _, row := range rows {
		path, fields, hasFields := strings.Cut(row, "\t")
		source := lintBytes(t, path)
		key := lintKey("input", map[string]string{"name": filepath.Base(path), "source": source})
		destination := filepath.Join(lintCacheRoot(t, "inputs"), key, filepath.Base(path))
		lintPublish(t, destination, []byte(source), 0600)
		if hasFields {
			destination += "\t" + fields
		}
		result = append(result, destination)
	}
	return result
}
func ruleRows(t *testing.T, d registry.Descriptor) ([]string, []string) {
	inherited := inheritedRuleRows(generated(t), d.Name)
	rows := append(inherited, selectedRows(inherited, d.Name)...)
	witnesses := ownedWitnesses(t, ".", d.Slug)
	var owned []string
	for _, source := range witnesses {
		owned = append(owned, source+"\t"+d.Name)
	}
	rows = append(rows, owned...)
	for _, source := range witnesses {
		rows = append(rows, source+"\tall")
	}
	return stableRows(t, rows), stableRows(t, owned)
}

type lintPhase struct {
	Name     string
	Duration time.Duration
}

func TestRule(t *testing.T) {
	if *ruleSlug == "" {
		t.Skip("use cmd/adamic-lint-check <slug> or -args -rule <slug>")
	}
	started := time.Now()
	var phases []lintPhase
	phase := func(name string, run func()) {
		before := time.Now()
		defer func() { phases = append(phases, lintPhase{name, time.Since(before)}) }()
		run()
	}
	defer func() {
		fmt.Printf("rule %s phase timing (seconds)\n| phase | seconds |\n|---|---:|\n", *ruleSlug)
		for _, p := range phases {
			fmt.Printf("| %s | %.3f |\n", p.Name, p.Duration.Seconds())
		}
		fmt.Printf("| total | %.3f |\n", time.Since(started).Seconds())
	}()
	var d registry.Descriptor
	var directory, mutated, oracle, binary, mutantBinary, module, path string
	var change struct{ Name, File, From, To string }
	var rows, owned, recovery []string
	phase("validate and snapshot", func() {
		d = selectedDescriptor(t, *ruleSlug)
		directory = selectedPort(t, d)
		if err := json.Unmarshal([]byte(lintBytes(t, filepath.Join("rules", d.Slug, "mutant.json"))), &change); err != nil {
			t.Fatal(err)
		}
		if change.File == "" {
			change.File = d.Module
		}
		mutated = selectedPortMutation(t, d, change.From, change.To, filepath.Join("rules", d.Slug, change.File))
		rows, owned = ruleRows(t, d)
	})
	phase("Go oracle", func() { oracle = goOracleFrom(t, directory) })
	phase("upstream capture", func() {
		for _, row := range stableRows(t, lintCapture(t, ".", d)) {
			if strings.HasSuffix(row, "\tunsupported-recovery") {
				recovery = append(recovery, row)
				continue
			}
			rows = append(rows, row)
		}
		path = manifest(t, recoveryRows(t, oracle, rows))
	})
	phase("owned witnesses", func() {
		for _, row := range owned {
			answer := execute(t, "", oracle, "--manifest", manifest(t, []string{row}), "--count")
			if string(answer.output) == "0\n" {
				t.Fatalf("%s witness reports no findings", d.Name)
			}
		}
	})
	phase("sanitized correct and mutant port builds", func() {
		var pending sync.WaitGroup
		pending.Add(2)
		go func() { defer pending.Done(); binary = buildPort(t, directory, true) }()
		go func() { defer pending.Done(); mutantBinary = buildPort(t, mutated, true) }()
		pending.Wait()
		if t.Failed() {
			t.FailNow()
		}
	})
	phase("explicit recovery refusals", func() {
		for _, row := range recovery {
			checkRecoveryRefusal(t, oracle, binary, directory, row)
		}
	})
	phase("emitted JavaScript build", func() { module = emittedJavaScript(t, directory) })
	phase("Go Node JavaScript native comparison", func() {
		if *ruleSemanticChange {
			semanticComparison(t, oracle, binary, directory, path, module)
		}
		compareWithJavaScript(t, oracle, binary, directory, path, module)
	})
	phase("owned mutant", func() {
		// Keep today's exact mutant corpus, including all-rule inherited rows.
		mutantPath := manifest(t, append(stableRows(t, inheritedRuleRows(generated(t), d.Name)), owned...))
		want := execute(t, "", oracle, "--manifest", mutantPath).output
		for _, side := range []struct {
			Name   string
			Result execution
		}{{"Node", node(t, mutated, mutantPath, false)}, {"emitted JavaScript", emittedNode(t, mutated, mutantPath, false)}, {"sanitized native", execute(t, "", mutantBinary, "--manifest", mutantPath)}} {
			if bytes.Equal(side.Result.output, want) {
				t.Fatalf("%s survived on %s", change.Name, side.Name)
			}
			t.Logf("%s caught on %s: %s", change.Name, side.Name, difference(side.Result.output, want))
		}
	})
}

// Not parallel: capture uses process-wide environment state. This checks the
// actual full registration against the restricted build on identical selected
// manifests, including witnesses, inherited options and captured upstream cases.
func TestSelectedRuleParity(t *testing.T) {
	for _, slug := range []string{"no-var", "no-empty", "eqeqeq"} {
		t.Run(slug, func(t *testing.T) {
			d := selectedDescriptor(t, slug)
			rows, _ := ruleRows(t, d)
			rows = selectedRows(rows, d.Name)
			rows = append(rows, stableRows(t, lintCapture(t, ".", d))...)
			path := manifest(t, rows)
			full, err := filepath.Abs(".")
			if err != nil {
				t.Fatal(err)
			}
			oracle := goOracle(t)
			want := compare(t, oracle, buildPort(t, full, true), full, path)
			selected := selectedPort(t, d)
			got := compare(t, oracle, buildPort(t, selected, true), selected, path)
			if !bytes.Equal(got, want) {
				t.Fatal("selected and all-rule findings differ")
			}
			t.Logf("selected and full registration findings identical: %d bytes", len(got))
		})
	}
}

// A clean-running wrong rule must receive the same failing verdict in both
// scopes. Comparing all three port runtimes also prevents a native-only failure
// from passing as the expected rejection.
func TestSelectedFailingRuleParity(t *testing.T) {
	d := selectedDescriptor(t, "eqeqeq")
	rows, _ := ruleRows(t, d)
	rows = selectedRows(rows, d.Name)
	rows = append(rows, stableRows(t, lintCapture(t, ".", d))...)
	path := manifest(t, rows)
	oracle := goOracle(t)
	want := execute(t, "", oracle, "--manifest", path).output
	full := mutant(t, "'unexpected'", "'unexpectee'", filepath.Join("rules", d.Slug, d.Module))
	selected := selectedPortMutation(t, d, "'unexpected'", "'unexpectee'", filepath.Join("rules", d.Slug, d.Module))
	selectedWant := execute(t, "", goOracleFrom(t, selected), "--manifest", path).output
	if !bytes.Equal(selectedWant, want) {
		t.Fatal("selected and full oracle answers differ")
	}
	var reference []byte
	for _, directory := range []string{full, selected} {
		binary := buildPort(t, directory, true)
		module := emittedJavaScript(t, directory)
		for _, side := range []execution{node(t, directory, path, false), runJavaScript(t, module, path, false), execute(t, "", binary, "--manifest", path)} {
			if reference == nil {
				reference = side.output
			}
			if !bytes.Equal(side.output, reference) {
				t.Fatal("failing selected/full outputs differ")
			}
			if bytes.Equal(side.output, want) {
				t.Fatal("wrong message-ID rule earned a passing verdict")
			}
		}
	}
	t.Logf("full and selected verdict FAIL on Node, JavaScript and native; byte-identical findings: %d bytes", len(reference))
}
