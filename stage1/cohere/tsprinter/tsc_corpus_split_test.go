package tsprinter

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

const testTSCCorpusAgreementShards = 8

// Four deterministic byte-balanced shards per family hold every repository
// fragment. ADAMIC_TEST_SHARD=i/n selects indices modulo n equal to i; unset
// runs all. The gate can select TestTSCCorpusAgreement/shard-NNN directly.
// Builds are prepared once per invocation and shared before t.Parallel.
func TestTSCCorpusAgreement(t *testing.T) {
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	files, err := trackedRootFiles(t, root, "stage3/drivers/tsc/corpus")
	if err != nil {
		t.Fatal(err)
	}
	oracle := tscOracle(t, root)
	plan := &corpusPlan{}
	var answers [][]byte
	var want strings.Builder
	type familyInput struct {
		name, directory string
		cases           []printerCase
		products        tsPrinterProducts
	}
	var families []familyInput
	var owners []int
	for _, statements := range []bool{false, true} {
		family, entry := "expressions", "main.ts"
		if statements {
			family, entry = "statements", "statementsMain.ts"
		}
		cases, directory := tscCorpus(t, root, files, family, statements, oracle)
		port, _ := filepath.Abs(entry)
		products := prepareTSPrinterProducts(t, port, family)
		shards, outputs, whole := tscPartition(t, family, cases, testTSCCorpusAgreementShards/2, len(plan.labels))
		for i, item := range cases {
			plan.labels = append(plan.labels, fmt.Sprintf("%s/case-%06d %s", family, i, item.Label))
		}
		plan.shards = append(plan.shards, shards...)
		answers = append(answers, outputs...)
		want.WriteString(whole)
		for range shards {
			owners = append(owners, len(families))
		}
		families = append(families, familyInput{family, directory, cases, products})
		t.Logf("%s union: %d repository fragment ids across %d shards", family, len(cases), len(shards))
	}
	if len(plan.shards) != testTSCCorpusAgreementShards {
		t.Fatalf("enumerated %d shards, declared %d", len(plan.shards), testTSCCorpusAgreementShards)
	}
	if err := expressionUnion(plan, answers, []byte(want.String())); err != nil {
		t.Fatal(err)
	}
	t.Logf("union: %d unique family/case ids across %d shards", len(plan.labels), len(plan.shards))
	reports := make([][]map[string]string, len(plan.shards))
	// Write caller-requested audits only after every parallel shard has finished.
	t.Cleanup(func() {
		for owner, family := range families {
			report := make([]map[string]string, 0)
			for i, items := range reports {
				if owners[i] == owner {
					report = append(report, items...)
				}
			}
			if keep := os.Getenv("ADAMIC_TSC_PRINTER_AUDIT"); keep != "" {
				if err := os.MkdirAll(keep, 0755); err != nil {
					t.Fatal(err)
				}
				data, err := json.MarshalIndent(report, "", "  ")
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(keep, family.name+".json"), data, 0644); err != nil {
					t.Fatal(err)
				}
				data, err = json.Marshal(family.cases)
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(keep, family.name+"-cases.json"), data, 0644); err != nil {
					t.Fatal(err)
				}
				data, err = os.ReadFile(filepath.Join(family.directory, "coverage.json"))
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(keep, family.name+"-coverage.json"), data, 0644); err != nil {
					t.Fatal(err)
				}
			}
			t.Logf("%s: %d tsc corpus fragments, %d disagreements", family.name, len(family.cases), len(report))
		}
	})
	selected := expressionSelection(t, len(plan.shards))
	for number, shard := range plan.shards {
		if !selected[number] {
			continue
		}
		family := families[owners[number]]
		t.Run(fmt.Sprintf("shard-%03d", number), func(t *testing.T) {
			t.Parallel()
			start := time.Now()
			t.Cleanup(func() {
				if elapsed := time.Since(start); elapsed > 30*time.Second {
					t.Errorf("invalid test unit: %.3fs exceeds 30s", elapsed.Seconds())
				}
			})
			t.Logf("family %s: %d fragments, original case ids %v", family.name, len(shard.indices), shard.indices)
			args := []string{"--cases", shard.text, "80"}
			check := func(name string, result run) {
				if err := expressionDisagreement(plan, number, answers[number], result); err != nil {
					t.Errorf("%s: %v", name, err)
				}
				// Retain every per-case audit comparison, including later disagreements.
				lines := strings.Split(strings.TrimSuffix(string(result.stdout), "\n"), "\n")
				expected := strings.Split(strings.TrimSuffix(string(answers[number]), "\n"), "\n")
				for index, id := range shard.indices {
					got := ""
					if index < len(lines) {
						got = lines[index]
					}
					if got != expected[index] {
						local := id
						if owners[number] == 1 {
							local -= len(families[0].cases)
						}
						item := family.cases[local]
						reports[number] = append(reports[number], map[string]string{"build": name, "file": strings.TrimPrefix(item.Label, root+"/"), "source": item.Source, "port": got, "go": expected[index]})
					}
				}
			}
			check("Node", onNode(t, family.products.source, args...))
			check("native", execute(t, []string{"ASAN_OPTIONS=detect_leaks=0"}, family.products.sanitized, args...))
			check("backend", onNode(t, family.products.backend, args...))
			switch runtime.GOOS {
			case "linux":
				if err := expressionDisagreement(plan, number, answers[number], execute(t, []string{"ASAN_OPTIONS=detect_leaks=1"}, family.products.sanitized, args...)); err != nil {
					t.Errorf("leaks: %v", err)
				}
			case "darwin":
				report := execute(t, nil, "leaks", append([]string{"--atExit", "--", family.products.release}, args...)...)
				if report.exitCode != 0 {
					t.Errorf("leaks: exit %d stdout %s stderr %s", report.exitCode, report.stdout, report.stderr)
				}
			default:
				t.Fatalf("no leak check for %s", runtime.GOOS)
			}
		})
	}
}
