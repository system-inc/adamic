package oracle

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// Supplemental frozen queue rows are kept separate from the 17-row lazy queue.
func TestCheckedViewIntersectionOriginalRankedImport(t *testing.T) {
	checkIntersectionRankedRead(t, "ranked-intersection-manifest.json", 9245, 14, "literal-import-good.a", "argument")
}

func TestCheckedViewIntersectionOriginalRankedOperands(t *testing.T) {
	for _, id := range []int{8920, 40931} {
		t.Run(fmt.Sprint(id), func(t *testing.T) {
			checkIntersectionRankedRead(t, fmt.Sprintf("ranked-intersection-%d-manifest.json", id), id, 3, fmt.Sprintf("numeric-operand-%d-good.a", id), "operand")
		})
	}
}

func checkIntersectionRankedRead(t *testing.T, manifestFile string, id, reads int, input, field string) {
	t.Helper()
	declarations, original := intersectionOriginalInputs(t)
	data, err := os.ReadFile(filepath.Join(declarations, manifestFile))
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		Commit         string            `json:"upstream_commit"`
		ID             int               `json:"type_id"`
		Reads          int               `json:"read_count"`
		Sites          []json.RawMessage `json:"sites"`
		InventorySHA   string            `json:"inventory_sha256"`
		SitesSHA       string            `json:"sites_sha256"`
		ReceiverFields []string          `json:"receiver_fields"`
		PresentFields  []string          `json:"present_fields"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	if manifest.Commit != original.Commit || manifest.ID != id || manifest.Reads != reads || len(manifest.Sites) != reads {
		t.Fatal("ranked provenance drift")
	}
	for name, digest := range map[string]string{"read-demand-pairs.json.gz": manifest.InventorySHA, "read-demand-sites.json.gz": manifest.SitesSHA} {
		source, err := os.ReadFile("../../stage3/interface-downcasts/lane4/" + name)
		if err != nil {
			t.Fatal(err)
		}
		if fmt.Sprintf("%x", sha256.Sum256(source)) != digest {
			t.Fatal("frozen inventory drift: " + name)
		}
	}
	for _, variant := range []string{"good", "wrong"} {
		t.Run(variant, func(t *testing.T) {
			source, err := os.ReadFile("../../stage3/interface-downcasts/lane7/original/" + input)
			if err != nil {
				t.Fatal(err)
			}
			bound := strings.Replace(string(source), "'original-tsc-types'", fmt.Sprintf("%q", filepath.ToSlash(filepath.Join(declarations, "compiler/types.d.ts"))), 1)
			if variant == "wrong" {
				bound = strings.Replace(bound, "const "+field+"={pos:0,", "const "+field+"={pos:'wrong',", 1)
			}
			fixture := filepath.Join(t.TempDir(), "ranked-import-"+variant+".a")
			if err := os.WriteFile(fixture, []byte(bound), 0600); err != nil {
				t.Fatal(err)
			}
			if difference := disagreement(run{stdout: []byte("true\n")}, onNode(t, fixture)); difference != "" {
				t.Fatal("Node: " + difference)
			}
			program, err := lowered(t, fixture)
			if err != nil {
				t.Fatal(err)
			}
			requireIntersectionOriginalComplete(t, program, original, []string{"Node"})
			fields := func(contract ir.ViewContract) []string {
				result := []string{}
				for _, field := range contract.Fields {
					result = append(result, field.Name)
				}
				slices.Sort(result)
				return result
			}
			complete := false
			for _, contract := range program.ViewContracts {
				if contract.Name == "CheckedReceiver" && slices.Equal(fields(contract), manifest.ReceiverFields) {
					complete = true
				}
			}
			if !complete {
				t.Fatal("complete original receiver fields missing")
			}
			recordIntersectionOriginalCounts(t, fixture, fmt.Sprintf("%d-%s", id, variant))
			found := false
			for i := range program.Functions {
				function := &program.Functions[i]
				if function.Name != "helper" {
					continue
				}
				for j, statement := range function.Body {
					declare, ok := statement.(ir.Declare)
					if !ok {
						continue
					}
					property, ok := declare.Value.(ir.Property)
					if !ok || property.View != "node."+field {
						continue
					}
					found = true
					if property.ViewContract == 0 {
						t.Fatal("original member contract missing")
					}
					contract := program.ViewContracts[property.ViewContract-1]
					if !contract.IntersectionBounded || contract.Unsupported != "" || !slices.Equal(fields(contract), manifest.PresentFields) {
						t.Fatal("complete original intersection member obligations missing")
					}
					if os.Getenv("ADAMIC_INTERSECTION_RANKED_MUTANT") == "1" || os.Getenv("ADAMIC_INTERSECTION_RANKED_MUTANT") == fmt.Sprint(id) {
						property.ViewContract = 0
						declare.Value = property
						function.Body[j] = declare
					}
				}
			}
			if !found {
				t.Fatal("original helper read missing")
			}
			actual, binary := nativelyUncached(t, program)
			if variant == "good" {
				if report := leaks(t, program, binary); report != "" {
					t.Fatal(report)
				}
			}
			for _, got := range []run{actual, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if variant == "good" {
					if difference := disagreement(run{stdout: []byte("true\n")}, got); difference != "" {
						t.Errorf("%s: exit %d stdout %q stderr %q", difference, got.exitCode, got.stdout, got.stderr)
					}
				} else if got.exitCode != 70 || !strings.Contains(string(got.stderr), "node."+field+".pos") {
					t.Errorf("missing named member refusal: exit %d stdout %q stderr %q", got.exitCode, got.stdout, got.stderr)
				}
			}
		})
	}
}
