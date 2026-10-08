package oracle

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

type callableShareAMember struct {
	Rank        int    `json:"rank"`
	Directory   string `json:"directory"`
	Field       string `json:"field"`
	Read        string `json:"read"`
	Expected    string `json:"expected"`
	Stdout      string `json:"stdout"`
	MutantArity int    `json:"mutantArity"`
}

func callableShareAMembers(t *testing.T) []callableShareAMember {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(repository, "stage3/interface-downcasts/lane5/share-a/witnesses.json"))
	if err != nil {
		t.Fatal(err)
	}
	var evidence struct {
		Members []callableShareAMember `json:"members"`
	}
	if err := json.Unmarshal(data, &evidence); err != nil {
		t.Fatal(err)
	}
	if len(evidence.Members) == 0 {
		t.Fatal("no share-a evidence")
	}
	for _, member := range evidence.Members {
		if member.Rank%3 != 0 {
			t.Fatalf("rank %d belongs to another share", member.Rank)
		}
	}
	return evidence.Members
}

func TestCheckedViewCallableShareA(t *testing.T) {
	for _, member := range callableShareAMembers(t) {
		t.Run(member.Directory, func(t *testing.T) {
			for _, variant := range []string{"good", "wrong-arity", "wrong-value"} {
				t.Run(variant, func(t *testing.T) {
					program, path := interfaceFixture(t, "lane5/share-a/"+member.Directory+"/"+variant)
					truth := onNode(t, path)
					if variant == "wrong-value" {
						if truth.exitCode != 70 || !strings.Contains(string(truth.stderr), "TypeError:") {
							t.Fatalf("source Node: %#v", truth)
						}
					} else if truth.exitCode != 0 || string(truth.stdout) != member.Stdout {
						t.Fatalf("source Node: %#v want %q", truth, member.Stdout)
					}
					sanitized, binary := nativelyUncached(t, program)
					for index, got := range []run{releasedUncached(t, program), sanitized, onJavaScriptBackend(t, program)} {
						if variant == "good" {
							if difference := disagreement(truth, got); difference != "" {
								t.Fatalf("backend%d: %s", index, difference)
							}
							continue
						}
						found := fmt.Sprintf("function with arity %d", member.MutantArity)
						if variant == "wrong-value" {
							found = "number"
						}
						message := "adamic: panic: field read failed: " + member.Read + " expected " + member.Expected + ", found " + found + "\n"
						if variant == "wrong-value" && index < 2 {
							message = "adamic: panic: field read failed: " + member.Read + " is not a " + member.Expected + "; expected " + member.Expected + ", found number\n"
						}
						if got.exitCode != 70 || len(got.stdout) != 0 || string(got.stderr) != message {
							t.Fatalf("backend%d exit %d stdout %q stderr %q want %q", index, got.exitCode, got.stdout, got.stderr, message)
						}
					}
					if variant == "good" {
						if report := leaksUncached(t, program, binary); report != "" {
							t.Fatal(report)
						}
					}
					if variant == "wrong-arity" {
						// Alter only this read's expected arity in the loaded test IR. The
						// independently recorded producer descriptor and invocation stay intact.
						changed := 0
						for i := range program.ViewContracts {
							contract := &program.ViewContracts[i]
							if contract.Kind != ir.ViewCallable || contract.Name != member.Expected {
								continue
							}
							contract.Parameters = nil
							if member.MutantArity == 1 {
								program.ViewContracts = append(program.ViewContracts, ir.ViewContract{Kind: ir.ViewScalar, Name: "number", Of: ir.Number})
								program.ViewContracts[i].Parameters = []ir.ViewContractID{ir.ViewContractID(len(program.ViewContracts))}
							}
							changed++
						}
						if changed != 1 {
							t.Fatalf("mutant expected one signature, changed %d", changed)
						}
						for index, got := range []run{releasedUncached(t, program), onJavaScriptBackend(t, program)} {
							if got.exitCode != 0 || len(got.stderr) != 0 || string(got.stdout) != member.Stdout {
								t.Fatalf("mutant%d did not execute independently: %#v", index, got)
							}
							t.Logf("rank %d backend%d arity mutant caught by pinned exit/message: mutant exit 0 stdout %q", member.Rank, index, got.stdout)
						}
					}
				})
			}
		})
	}
}

// Not parallel: -update-counts appends only this share's measured rows.
func TestCheckedViewCallableShareACounts(t *testing.T) {
	path := filepath.Join(repository, "internal/oracle/counts.md")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, member := range callableShareAMembers(t) {
		for _, variant := range []string{"good", "wrong-arity", "wrong-value"} {
			row := counted(t, checkedViewFixturePath("stage3/interface-downcasts/lane5/share-a/"+member.Directory+"/"+variant+".a"), false, nil, false, false)
			if strings.Contains(text, row+"\n") {
				continue
			}
			key := strings.Split(row, " | ")[0] + " | "
			if strings.Contains(text, key) {
				t.Fatalf("existing share-a count changed: %s", row)
			}
			if !*updateCounts {
				t.Errorf("unrecorded share-a counts: %s", row)
				continue
			}
			text += row + "\n"
		}
	}
	if *updateCounts {
		if err := os.WriteFile(path, []byte(text), 0644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestCheckedViewCallableShareAFrontiers(t *testing.T) {
	data, err := os.ReadFile(filepath.Join(repository, "stage3/interface-downcasts/lane5/share-a/gaps.json"))
	if err != nil {
		t.Fatal(err)
	}
	var evidence struct {
		Members []struct {
			Rank      int    `json:"rank"`
			Directory string `json:"directory"`
			Field     string `json:"field"`
			Stdout    string `json:"stdout"`
		} `json:"members"`
	}
	if err := json.Unmarshal(data, &evidence); err != nil {
		t.Fatal(err)
	}
	for _, member := range evidence.Members {
		t.Run(member.Directory, func(t *testing.T) {
			path, err := filepath.Abs(checkedViewFixturePath(filepath.Join(repository, "stage3/interface-downcasts/lane5/share-a", member.Directory, "good.a")))
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			if truth.exitCode != 0 || string(truth.stdout) != member.Stdout {
				t.Fatalf("source Node: %#v", truth)
			}
			program, lowerError := lowered(t, path)
			err = lowerError
			if err == nil {
				sanitized, binary := nativelyUncached(t, program)
				switch member.Rank {
				case 3, 9, 24:
					// These original controls now also have negative and mutant certificates
					// in TestCheckedViewCallableFactory. Keep the frontier's source control.
					for _, got := range []run{releasedUncached(t, program), sanitized, onJavaScriptBackend(t, program)} {
						if difference := disagreement(truth, got); difference != "" {
							t.Fatal(difference)
						}
					}
					if report := leaksUncached(t, program, binary); report != "" {
						t.Fatal(report)
					}
				case 186, 189:
					// The reduced producer in these original controls serves only one arm.
					// A full-set runtime check must stop before the selected call executes.
					expected := "{ (phaseModifier: number | undefined, name: Identifier | undefined, namedBindings: NamedImportBindings | undefined): ImportClause; (isTypeOnly: boolean, name: Identifier | undefined, namedBindings: NamedImportBindings | undefined): ImportClause; }"
					read := "nodeFactory.createImportClause"
					if member.Rank == 189 {
						expected = "{ (asteriskToken: AsteriskToken, expression: Expression): YieldExpression; (asteriskToken: undefined, expression: Expression | undefined): YieldExpression; (asteriskToken: AsteriskToken | undefined, expression: Expression | undefined): YieldExpression; }"
						read = "factory.createYieldExpression"
					}
					message := "adamic: panic: field read failed: " + read + " expected " + expected + ", found function with incompatible overload signature 1\n"
					for _, got := range []run{releasedUncached(t, program), sanitized, onJavaScriptBackend(t, program)} {
						if got.exitCode != 70 || len(got.stdout) != 0 || string(got.stderr) != message {
							t.Fatalf("full overload-set frontier: %#v want %q", got, message)
						}
					}
				default:
					t.Fatal("frontier lowered; needs complete positive/negative/mutant certification")
				}
				return
			}
			t.Logf("rank %d original read refusal: %v", member.Rank, err)
			expected := "Adamic 0.1 refuses checked view read of field " + member.Field + " with unsupported callable contract; prove or implement the callable contract before reading this field"
			if !strings.HasSuffix(err.Error(), expected) {
				t.Fatalf("refusal does not name original member: %v", err)
			}
		})
	}
}
