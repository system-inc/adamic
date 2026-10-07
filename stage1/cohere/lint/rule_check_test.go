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

var ruleWhole = flag.Bool("rule-whole", false, "benchmark the ordinary whole-file compiler")
var ruleHelperChange = flag.Bool("rule-helper-change", false, "benchmark a private snapshot with an additional identity helper")

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
	if *ruleHelperChange {
		path := filepath.Join(directory, "rules", d.Slug, d.Module)
		source := lintBytes(t, path)
		source = strings.Replace(source, "export class Rule", "function lintCheckHelper(value: number): number { return value; }\n\nexport class Rule", 1)
		start := strings.Index(source, "    visit(")
		if start < 0 {
			t.Fatal("helper edit requires visit")
		}
		offset := strings.Index(source[start:], "{\n")
		if offset < 0 || !strings.Contains(source, "function lintCheckHelper(") {
			t.Fatal("helper edit requires a class and visit body")
		}
		body := offset + start + 2
		source = source[:body] + "        index = lintCheckHelper(index);\n" + source[body:]
		if err := os.WriteFile(path, []byte(source), 0644); err != nil {
			t.Fatal(err)
		}
	}
	prepareRegistry(t, directory)
	return directory
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
	inherited := append(generated(t), volumeGenerated(t)...)
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
		path = manifest(t, rows)
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
		if *ruleWhole {
			binary = buildPort(t, directory, true)
			mutantBinary = buildPort(t, mutated, true)
			return
		}
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
	phase("Go Node JavaScript native comparison", func() { compareWithJavaScript(t, oracle, binary, directory, path, module) })
	phase("owned mutant", func() {
		// Keep today's exact mutant corpus, including all-rule inherited rows.
		mutantPath := manifest(t, append(stableRows(t, generated(t)), owned...))
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
