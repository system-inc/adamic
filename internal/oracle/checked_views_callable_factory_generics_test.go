package oracle

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckedViewCallableFactoryGenerics(t *testing.T) {
	for _, pair := range []struct {
		rank                         int
		field, conversion, signature string
	}{
		{12, "cloneNode", "7\nok\ntrue\n9\nundefined\n", "<T extends Node | undefined>(node: T) => T"},
		{174, "createModifier", "7\n1\n2\n", "<T extends ModifierSyntaxKind>(kind: T) => ModifierToken<T>"},
	} {
		t.Run(pair.field, func(t *testing.T) {
			for _, variant := range []string{"good", "conversion", "wrong-producer"} {
				t.Run(variant, func(t *testing.T) {
					program, path := interfaceFixture(t, fmt.Sprintf("lane5/share-factory/rank-%d/%s", pair.rank, variant))
					if variant == "conversion" {
						instances := map[int]bool{}
						collect := func(value any) any {
							if call, ok := value.(ir.CallClosure); ok {
								for _, pair := range call.GenericInstances {
									if pair[1] >= 0 {
										instances[pair[1]] = true
									}
								}
							}
							return value
						}
						rewriteUnionTargetStatements(program.Main, collect)
						for _, function := range program.Functions {
							rewriteUnionTargetStatements(function.Body, collect)
						}
						wantInstances := 2
						if pair.rank == 12 {
							wantInstances = 4
						}
						if len(instances) != wantInstances {
							t.Fatalf("generic instantiations collapsed: got %d bodies, want %d", len(instances), wantInstances)
						}
						t.Logf("distinct instantiated bodies: %d", len(instances))
					}
					truth := onNode(t, path)
					want := "7\n"
					if variant == "conversion" {
						want = pair.conversion
					}
					if truth.exitCode != 0 || string(truth.stdout) != want || len(truth.stderr) != 0 {
						t.Fatalf("source Node: %#v want %q", truth, want)
					}
					sanitized, binary := nativelyUncached(t, program)
					for index, got := range []run{releasedUncached(t, program), sanitized, onJavaScriptBackend(t, program)} {
						if variant != "wrong-producer" {
							if difference := disagreement(truth, got); difference != "" {
								t.Fatalf("backend%d: %s", index, difference)
							}
						} else {
							message := fmt.Sprintf("adamic: panic: field read failed: factory.%s expected %s, found function with unknown signature\n", pair.field, pair.signature)
							if got.exitCode != 70 || len(got.stdout) != 0 || string(got.stderr) != message {
								t.Fatalf("generic negative%d did not stop: %#v", index, got)
							}
							t.Logf("backend%d pinned generic declaration rejection: %q", index, got.stderr)
						}
					}
					if variant != "wrong-producer" {
						if report := leaksUncached(t, program, binary); report != "" {
							t.Fatal(report)
						}
						return
					}
					producer := -1
					for index, function := range program.Functions {
						if function.Closure && !function.GenericTemplate && len(function.Parameters) == 1 {
							if producer != -1 {
								t.Fatal("ambiguous mutant producer")
							}
							producer = index
						}
					}
					if producer < 0 {
						t.Fatal("missing ordinary mutant producer")
					}
					changed := 0
					for index := range program.ViewContracts {
						contract := &program.ViewContracts[index]
						if contract.Generic {
							contract.Functions = append(contract.Functions, producer)
							changed++
						}
					}
					if changed != 1 {
						t.Fatalf("generic admission mutant changed %d", changed)
					}
					rewrite := func(value any) any {
						if call, ok := value.(ir.CallClosure); ok && len(call.GenericInstances) != 0 {
							call.GenericInstances = [][2]int{{producer, producer}}
							return call
						}
						return value
					}
					program.Main = rewriteUnionTargetStatements(program.Main, rewrite)
					for index := range program.Functions {
						program.Functions[index].Body = rewriteUnionTargetStatements(program.Functions[index].Body, rewrite)
					}
					mutantSanitized, mutantBinary := nativelyUncached(t, program)
					for index, got := range []run{releasedUncached(t, program), mutantSanitized, onJavaScriptBackend(t, program)} {
						if difference := disagreement(truth, got); difference != "" {
							t.Fatalf("mutant%d did not execute cleanly: %s", index, difference)
						}
						t.Logf("backend%d non-generic admission mutant caught: stdout %q instead of exit 70", index, got.stdout)
					}
					if report := leaksUncached(t, program, mutantBinary); report != "" {
						t.Fatal(report)
					}
				})
			}
		})
	}
}

func TestCheckedViewCallableFactoryGenericResultProof(t *testing.T) {
	for _, rank := range []int{12, 174} {
		t.Run(fmt.Sprintf("rank-%d", rank), func(t *testing.T) {
			fixture := "good"
			if rank == 174 {
				fixture = "conversion"
			}
			source, readErr := os.ReadFile(filepath.Join(repository, fmt.Sprintf("stage3/interface-downcasts/lane5/share-factory/rank-%d/%s.a", rank, fixture)))
			if readErr != nil {
				t.Fatal(readErr)
			}
			text := string(source)
			if rank == 12 {
				text = strings.Replace(text, "return node;", "return {value:0} as T;", 1)
			} else {
				text = strings.Replace(text, "return {value:7,kind};", "return {value:7,kind:0} as ModifierToken<T>;", 1)
			}
			path := filepath.Join(t.TempDir(), "wrong-instantiation.a")
			if err := os.WriteFile(path, []byte(text), 0600); err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			if truth.exitCode != 0 || len(truth.stderr) != 0 {
				t.Fatalf("source Node: %#v", truth)
			}
			_, err := lowered(t, path)
			if err == nil || !strings.HasSuffix(err.Error(), "Adamic 0.1 refuses a cast the runtime can't check; use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast)") {
				t.Fatalf("asserted generic result admitted: %v", err)
			}
			t.Logf("pinned result proof refusal: %v", err)
		})
	}
}

func TestCheckedViewCallableFactoryGenericMethodDeclaration(t *testing.T) {
	source, err := os.ReadFile(filepath.Join(repository, "stage3/interface-downcasts/lane5/share-factory/rank-12/wrong-producer.a"))
	if err != nil {
		t.Fatal(err)
	}
	text := strings.Replace(string(source), "probe({cloneNode:(node:Node):Node=>node});", "class Source { cloneNode(node:Node):Node {return node;} }\nprobe(new Source());", 1)
	path := filepath.Join(t.TempDir(), "generic-method.a")
	if err := os.WriteFile(path, []byte(text), 0600); err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	if truth.exitCode != 0 || string(truth.stdout) != "7\n" || len(truth.stderr) != 0 {
		t.Fatalf("source Node: %#v", truth)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	sanitized, _ := nativelyUncached(t, program)
	t.Logf("sanitized generic method exit=%d stdout=%q stderr=%q", sanitized.exitCode, sanitized.stdout, sanitized.stderr)
	message := "adamic: panic: field read failed: factory.cloneNode expected <T extends Node | undefined>(node: T) => T, found function with unknown signature\n"
	for index, got := range []run{releasedUncached(t, program), sanitized, onJavaScriptBackend(t, program)} {
		t.Logf("generic method negative exit=%d stdout=%q stderr=%q", got.exitCode, got.stdout, got.stderr)
		if got.exitCode != 70 || len(got.stdout) != 0 || string(got.stderr) != message {
			t.Fatalf("ordinary method admitted by generic contract%d: %#v", index, got)
		}
	}
}
