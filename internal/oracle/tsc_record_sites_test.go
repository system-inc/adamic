package oracle

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

const tscRecordSitesPath = "stage3/maplike-records/sites/results.json"

var updateTSCRecordSites = flag.Bool("update-tsc-record-sites", false, "record initial tsc record site outcomes or newly compiling reductions")

type tscRecordObservation struct {
	Stdout string `json:"stdout"`
	Stderr string `json:"stderr"`
	Exit   int    `json:"exit"`
}

type tscRecordSite struct {
	Rank       int                   `json:"rank"`
	Batch      int                   `json:"batch"`
	Source     string                `json:"source"`
	Function   string                `json:"function"`
	Calls      int                   `json:"static_calls"`
	Fixture    string                `json:"fixture"`
	Adaptation string                `json:"adaptation"`
	Node       *tscRecordObservation `json:"node,omitempty"`
	Outcome    string                `json:"outcome"`
	Diagnostic string                `json:"diagnostic"`
}

func readTSCRecordSites() ([]tscRecordSite, error) {
	data, err := os.ReadFile(filepath.Join(repository, tscRecordSitesPath))
	if err != nil {
		return nil, err
	}
	var sites []tscRecordSite
	err = json.Unmarshal(data, &sites)
	return sites, err
}

func init() {
	sites, err := readTSCRecordSites()
	if err != nil {
		panic(err)
	}
	seen := map[string]bool{}
	for _, site := range sites {
		if site.Outcome == "Compiles" && !seen[site.Fixture] {
			seen[site.Fixture] = true
			fixtures = append(fixtures, struct {
				path            string
				lowers, checked bool
			}{site.Fixture, true, false})
		}
	}
}

// Not parallel: initial recording writes a shared manifest. Source observations
// are immutable after recording, including when a compiler stop later closes.
func TestTSCRecordSites(t *testing.T) {
	sites, err := readTSCRecordSites()
	if err != nil {
		t.Fatal(err)
	}
	for index := range sites {
		site := &sites[index]
		t.Run(fmt.Sprintf("batch_%02d/site_%02d", site.Batch, site.Rank), func(t *testing.T) {
			absolute, err := filepath.Abs(filepath.Join(repository, site.Fixture))
			if err != nil {
				t.Fatal(err)
			}
			// Run source directly: some reductions deliberately stop in the checker,
			// so Node must not pass through the checked import-cache loader.
			node := execute(t, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle", "node.mjs"), absolute)
			observed := tscRecordObservation{string(node.stdout), string(node.stderr), node.exitCode}
			if observed.Exit != 0 || observed.Stderr != "" {
				t.Fatalf("Node reduction did not finish cleanly: %+v", observed)
			}
			if site.Node != nil && *site.Node != observed {
				t.Fatalf("recorded Node observation changed: was %+v, now %+v", *site.Node, observed)
			}
			if site.Node == nil {
				if !*updateTSCRecordSites {
					t.Fatal("missing recorded Node observation")
				}
				site.Node = &observed
			}
			outcome, diagnostic := "Compiles", ""
			program, loadErr := load.Load([]string{absolute})
			if loadErr != nil {
				outcome, diagnostic = "Checker", loadErr.Error()
			} else {
				loweredProgram, lowerErr := lower.Lower(context.Background(), program)
				if lowerErr != nil {
					var refused *lower.Refused
					var notYet *lower.NotYet
					switch {
					case errors.As(lowerErr, &refused):
						outcome = "Refused"
					case errors.As(lowerErr, &notYet):
						outcome = "NotYet"
					default:
						t.Fatalf("unexpected lowering error: %v", lowerErr)
					}
					diagnostic = lowerErr.Error()
				} else {
					native, binary := nativelyUncached(t, loweredProgram)
					for backend, got := range map[string]run{"native": native, "javascript": onJavaScriptBackend(t, loweredProgram)} {
						if diff := disagreement(node, got); diff != "" {
							t.Fatalf("%s %s: Node %+v; backend %+v", backend, diff, node, got)
						}
					}
					if report := leaksUncached(t, loweredProgram, binary); report != "" {
						t.Fatal(report)
					}
				}
			}
			root, err := filepath.Abs(repository)
			if err != nil {
				t.Fatal(err)
			}
			diagnostic = strings.ReplaceAll(diagnostic, root+string(filepath.Separator), "")
			if site.Outcome != outcome || site.Diagnostic != diagnostic {
				if !*updateTSCRecordSites || (site.Outcome != "" && outcome != "Compiles") {
					t.Fatalf("first stop changed: recorded %s %q; got %s %q", site.Outcome, site.Diagnostic, outcome, diagnostic)
				}
				site.Outcome, site.Diagnostic = outcome, diagnostic
			}
			t.Logf("%s:%s %s %s", site.Source, site.Function, outcome, diagnostic)
		})
	}
	if *updateTSCRecordSites && !t.Failed() {
		data, err := json.MarshalIndent(sites, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(repository, tscRecordSitesPath), append(data, '\n'), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

// Change real dictionary input while preserving a valid program and ownership.
// Removing one integer key must be caught only by the Node stdout comparison.
func TestTSCRecordSiteKeyMutant(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "stage3/maplike-records/sites/fixtures/getOwnKeys.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	changed := false
	mutateStringExpressions(reflect.ValueOf(&program.Main).Elem(), func(expression ir.Expression) ir.Expression {
		literal, ok := expression.(ir.RecordLiteral)
		if !ok || changed || len(literal.Entries) == 0 {
			return expression
		}
		literal.Entries = literal.Entries[1:]
		changed = true
		return literal
	})
	if !changed {
		t.Fatal("mutant removed no dictionary input")
	}
	want := execute(t, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle", "node.mjs"), path)
	native, binary := nativelyUncached(t, program)
	for backend, got := range map[string]run{"native": native, "javascript": onJavaScriptBackend(t, program)} {
		if got.exitCode != 0 || len(got.stderr) != 0 || disagreement(want, got) != "stdout differs" {
			t.Fatalf("%s key-removal mutant was not caught only by stdout: %+v", backend, got)
		}
		t.Logf("%s key-removal mutant caught by Node stdout; valid build, exit 0, clean sanitizers; mutant stdout %q", backend, got.stdout)
	}
	if report := leaksUncached(t, program, binary); report != "" {
		t.Fatal(report)
	}
}
