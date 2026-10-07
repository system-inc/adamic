package lower

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

func TestViewCallableShapeContract(t *testing.T) {
	t.Parallel()
	for _, declaration := range []string{
		"type Target = (value: number) => number;",
		"type Target = () => string;",
		"type Target = (value?: number) => number;",
		"type Target = <T>(value: T) => T;",
		"type Target = (...values: number[]) => number;",
		"type Target = { (value: number): number; (value: string): string; };",
	} {
		t.Run(declaration, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "contract.a")
			if err := os.WriteFile(path, []byte(declaration), 0600); err != nil {
				t.Fatal(err)
			}
			program, err := load.Load([]string{path})
			if err != nil {
				t.Fatal(err)
			}
			file := program.Files()[0]
			checked, release := program.Checker(context.Background(), file)
			defer release()
			l := &lowering{program: program, checker: checked, result: &ir.Program{}}
			node := file.Statements.Nodes[0]
			target := checked.GetTypeAtLocation(node.Name())
			calls := 0
			id, err := buildViewCallableContract(l, node, target, func(*checker.Type) (ir.ViewContractID, error) {
				calls++
				return 0, errors.New("cast must not build a signature")
			})
			if err != nil || calls != 0 {
				t.Fatalf("cast admission built signature: %v, calls %d", err, calls)
			}
			if l.result.ViewContracts[id-1].Result != 0 {
				t.Fatal("unknown signature recorded as proven")
			}
			supported := declaration == "type Target = (value: number) => number;" || declaration == "type Target = () => string;"
			sentinel := errors.New("child unavailable")
			err = l.completeViewCallableShapeContract(node, target, id, func(*checker.Type) (ir.ViewContractID, error) { calls++; return 0, sentinel })
			if supported {
				if !errors.Is(err, sentinel) || calls != 1 {
					t.Fatalf("lost child failure: %v calls %d", err, calls)
				}
				calls = 0
				err = l.completeViewCallableShapeContract(node, target, id, func(child *checker.Type) (ir.ViewContractID, error) {
					calls++
					of, known := l.representation(child)
					if !known {
						t.Fatal("control representation unavailable")
					}
					childID := ir.ViewContractID(len(l.result.ViewContracts) + 1)
					l.result.ViewContracts = append(l.result.ViewContracts, ir.ViewContract{Kind: ir.ViewScalar, Of: of})
					return childID, nil
				})
				if err != nil {
					t.Fatal(err)
				}
				contract := l.result.ViewContracts[id-1]
				wantParameters := 1
				if declaration == "type Target = () => string;" {
					wantParameters = 0
				}
				if len(contract.Parameters) != wantParameters || contract.Result == 0 || calls != wantParameters+1 {
					t.Fatalf("incomplete descriptor %#v calls %d", contract, calls)
				}
				if wantParameters != 0 {
					calls = 0
					err = l.completeViewCallableShapeContract(node, target, id, func(*checker.Type) (ir.ViewContractID, error) {
						calls++
						if calls == 1 {
							return 0, nil
						}
						return 2, nil
					})
					if _, ok := err.(*NotYet); !ok {
						t.Fatalf("unknown parameter accepted: %v", err)
					}
				}
				calls = 0
				err = l.completeViewCallableShapeContract(node, target, id, func(*checker.Type) (ir.ViewContractID, error) {
					calls++
					if calls == wantParameters+1 {
						return 0, nil
					}
					return 2, nil
				})
				if _, ok := err.(*NotYet); !ok {
					t.Fatalf("unknown result accepted: %v", err)
				}

			} else {
				if _, ok := err.(*NotYet); !ok || calls != 0 {
					t.Fatalf("unsupported signature must fail at read: %v calls %d", err, calls)
				}
			}
		})
	}
}
