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

type callableShareDMember struct {
	Rank         int    `json:"rank"`
	Directory    string `json:"directory"`
	Field        string `json:"field"`
	Read         string `json:"read"`
	Expected     string `json:"expected"`
	Stdout       string `json:"stdout"`
	MutantArity  int    `json:"mutantArity"`
	Stop         string `json:"stop"`
	StopPosition string `json:"stopPosition"`
}

func callableShareDMembers(t *testing.T, ledger string) []callableShareDMember {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(repository, "stage3/interface-downcasts/lane5/share-d", ledger))
	if err != nil {
		t.Fatal(err)
	}
	var evidence struct {
		Members []callableShareDMember `json:"members"`
	}
	if err := json.Unmarshal(data, &evidence); err != nil {
		t.Fatal(err)
	}
	for _, member := range evidence.Members {
		if member.Rank%3 != 0 {
			t.Fatalf("rank %d belongs to another share", member.Rank)
		}
	}
	return evidence.Members
}

func callableShareDNegative(member callableShareDMember, variant string, backend int) run {
	found := fmt.Sprintf("function with arity %d", member.MutantArity)
	if variant == "wrong-value" {
		found = "number"
	}
	message := "adamic: panic: field read failed: " + member.Read + " expected " + member.Expected + ", found " + found + "\n"
	if variant == "wrong-value" && backend < 2 {
		message = "adamic: panic: field read failed: " + member.Read + " is not a " + member.Expected + "; expected " + member.Expected + ", found number\n"
	}
	return run{exitCode: 70, stderr: []byte(message)}
}

func callableShareDMutateArity(t *testing.T, program *ir.Program, member callableShareDMember) {
	t.Helper()
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
}

func TestCheckedViewCallableShareD(t *testing.T) {
	for _, member := range callableShareDMembers(t, "witnesses.json") {
		t.Run(member.Directory, func(t *testing.T) {
			for _, variant := range []string{"good", "wrong-arity", "wrong-value"} {
				t.Run(variant, func(t *testing.T) {
					program, path := interfaceFixture(t, "lane5/share-d/"+member.Directory+"/"+variant)
					truth := onNode(t, path)
					if variant == "wrong-value" {
						if truth.exitCode != 70 || len(truth.stdout) != 0 || !strings.Contains(string(truth.stderr), "TypeError:") {
							t.Fatalf("source Node: %#v", truth)
						}
					} else if difference := disagreement(run{stdout: []byte(member.Stdout)}, truth); difference != "" {
						t.Fatalf("source Node: %s; got %#v", difference, truth)
					}
					// The external mutant run uses the same negative matcher and must fail.
					if variant == "wrong-arity" && os.Getenv("ADAMIC_CALLABLE_D_MUTANT") == "1" {
						callableShareDMutateArity(t, program, member)
					}
					sanitized, binary := nativelyUncached(t, program)
					for backend, result := range []run{releasedUncached(t, program), sanitized, onJavaScriptBackend(t, program)} {
						want := truth
						if variant != "good" {
							want = callableShareDNegative(member, variant, backend)
						}
						if difference := disagreement(want, result); difference != "" {
							t.Fatalf("rank %d %s backend%d: %s; exit %d stdout %q stderr %q; want exit %d stdout %q stderr %q", member.Rank, variant, backend, difference, result.exitCode, result.stdout, result.stderr, want.exitCode, want.stdout, want.stderr)
						}
					}
					if variant == "good" {
						if report := leaksUncached(t, program, binary); report != "" {
							t.Fatal(report)
						}
					}
					if variant == "wrong-arity" {
						callableShareDMutateArity(t, program, member)
						for backend, result := range []run{releasedUncached(t, program), onJavaScriptBackend(t, program)} {
							if difference := disagreement(run{stdout: []byte(member.Stdout)}, result); difference != "" {
								t.Fatalf("mutant did not execute: %s; got %#v", difference, result)
							}
							difference := disagreement(callableShareDNegative(member, variant, backend*2), result)
							if difference == "" {
								t.Fatal("negative matcher accepted the arity mutant")
							}
							t.Logf("rank %d backend%d arity mutant caught by pinned matcher: %s; mutant exit %d stdout %q stderr %q", member.Rank, backend, difference, result.exitCode, result.stdout, result.stderr)
						}
					}
				})
			}
		})
	}
}

func TestCheckedViewCallableShareDFrontiers(t *testing.T) {
	for _, member := range callableShareDMembers(t, "needs-code.json") {
		t.Run(member.Directory, func(t *testing.T) {
			path, err := filepath.Abs(checkedViewFixturePath(filepath.Join(repository, "stage3/interface-downcasts/lane5/share-d", member.Directory, "good.a")))
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			if difference := disagreement(run{stdout: []byte(member.Stdout)}, truth); difference != "" {
				t.Fatalf("Node: %s; got %#v", difference, truth)
			}
			_, err = lowered(t, path)
			want := path + ":" + member.StopPosition + ": " + member.Stop
			if err == nil || err.Error() != want {
				t.Fatalf("rank %d frontier changed: want %q, got %v", member.Rank, want, err)
			}
			t.Logf("rank %d original callable frontier: %v", member.Rank, err)
		})
	}
}

// Not parallel: the updater appends only this certifier's measured count rows.
func TestCheckedViewCallableShareDCounts(t *testing.T) {
	path := filepath.Join(repository, "internal/oracle/counts.md")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, member := range callableShareDMembers(t, "witnesses.json") {
		for _, variant := range []string{"good", "wrong-arity", "wrong-value"} {
			row := counted(t, checkedViewFixturePath("stage3/interface-downcasts/lane5/share-d/"+member.Directory+"/"+variant+".a"), false, nil, false, false)
			if strings.Contains(text, row+"\n") {
				continue
			}
			key := strings.Split(row, " | ")[0] + " | "
			if strings.Contains(text, key) {
				t.Fatalf("existing share-d count changed: %s", row)
			}
			if !*updateCounts {
				t.Errorf("unrecorded share-d counts: %s", row)
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
