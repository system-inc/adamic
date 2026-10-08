package oracle

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// The wrapper interface inherits the entire original intersection. Its existing
// checked view adds no earlier descendant read before the original helper.
func TestCheckedViewIntersectionOriginalIsolatedMembers(t *testing.T) {
	declarations, manifest := intersectionOriginalInputs(t)
	for _, sample := range []struct{ name, good, bad, receiver, value, field string }{
		{"9454", "bindable-element-good", "bindable-element-argument-wrong", "BindableStaticElementAccessExpression", "left", "argumentExpression"},
		{"9657", "class-augments-good", "class-augments-name", "JSDocAugmentsTag['class']", "heritage", "expression"},
	} {
		for _, variant := range []string{"good", "wrong"} {
			t.Run(sample.name+"/"+variant, func(t *testing.T) {
				fixture := sample.good
				if variant == "wrong" {
					fixture = sample.bad
				}
				input, err := os.ReadFile(checkedViewFixturePath("../../stage3/interface-downcasts/lane7/original/" + fixture + ".a"))
				if err != nil {
					t.Fatal(err)
				}
				source := strings.SplitN(string(input), "function ", 2)[0]
				source += fmt.Sprintf("type Receiver = %s; interface CheckedReceiver extends Receiver {}\nconst receiverBase: Base = %s;\nfunction helper(node: Receiver): boolean { const member = node.%s; return member.end > 0; }\nconsole.log(`${helper(receiverBase as CheckedReceiver)}`);\n", sample.receiver, sample.value, sample.field)
				source = strings.Replace(source, "'original-tsc-types'", fmt.Sprintf("%q", filepath.ToSlash(filepath.Join(declarations, "compiler/types.d.ts"))), 1)
				path := filepath.Join(t.TempDir(), "isolated-"+sample.name+"-"+variant+".a")
				if err := os.WriteFile(path, []byte(source), 0600); err != nil {
					t.Fatal(err)
				}
				if difference := disagreement(run{stdout: []byte("true\n")}, onNode(t, path)); difference != "" {
					t.Fatal("Node: " + difference)
				}
				program, err := lowered(t, path)
				if err != nil {
					t.Fatal(err)
				}
				requireIntersectionOriginalComplete(t, program, manifest, []string{"Node", "Identifier"})
				var receiverFields, memberFields []string
				pairID, _ := strconv.Atoi(sample.name)
				for _, pair := range manifest.Pairs {
					if pair.ID == pairID {
						receiverFields = pair.ReceiverFields
						memberFields = pair.PresentFields
					}
				}
				complete := false
				for _, contract := range program.ViewContracts {
					if contract.Name == "CheckedReceiver" {
						fields := []string{}
						for _, field := range contract.Fields {
							fields = append(fields, field.Name)
						}
						slices.Sort(fields)
						complete = complete || len(fields) > 0 && slices.Equal(fields, receiverFields)
					}
				}
				if !complete {
					t.Fatal("wrapper omitted original receiver fields")
				}
				recordIntersectionOriginalCounts(t, path, sample.name+"-"+variant)
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
						if !ok || property.View != "node."+sample.field {
							continue
						}
						found = true
						if property.ViewContract == 0 || program.ViewContracts[property.ViewContract-1].Unsupported != "" {
							t.Fatal("original helper contract missing")
						}
						selected := program.ViewContracts[property.ViewContract-1]
						fields := []string{}
						for _, field := range selected.Fields {
							fields = append(fields, field.Name)
						}
						slices.Sort(fields)
						if !selected.IntersectionBounded || !slices.Equal(fields, memberFields) {
							t.Fatal("original member obligations omitted")
						}
						if os.Getenv("ADAMIC_INTERSECTION_ISOLATED_MUTANT") == sample.name {
							property.ViewContract = 0
							declare.Value = property
							function.Body[j] = declare
						}
					}
				}
				if !found {
					t.Fatal("isolated original member read missing")
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
							t.Error(difference)
						}
					} else if got.exitCode != 70 || !strings.Contains(string(got.stderr), "node."+sample.field) {
						t.Errorf("missing named member refusal: exit %d stdout %q stderr %q", got.exitCode, got.stdout, got.stderr)
					}
				}
			})
		}
	}
}
