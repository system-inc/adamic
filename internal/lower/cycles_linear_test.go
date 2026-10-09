package lower

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime/pprof"
	"strings"
	"testing"
	"time"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/flow"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// cyclesProbeLower follows Lower through the finder boundary, exposing the
// checked program and lowered IR to the corpus and focused finder tests.
func cyclesProbeLower(ctx context.Context, program *load.Program, run func(*lowering, []*ast.SourceFile) error) (*ir.Program, error) {
	files := program.Files()
	if len(files) != 1 {
		return nil, fmt.Errorf("lower: stage 0 compiles a program from one entry file, got %d", len(files))
	}
	entry := files[0]
	// Stage 0 checks single-threaded, so one checker answers for every file.
	typeChecker, release := program.Checker(ctx, entry)
	defer release()

	lowering := &lowering{program: program, checker: typeChecker, result: &ir.Program{}, this: -1, functionIndex: -1}
	// The base name only, so the same program emits the same C on every machine.
	lowering.result.Source = filepath.Base(program.FileName(entry))
	modules, err := lowering.moduleOrder(entry)
	if err != nil {
		return nil, err
	}
	lowering.noteInheritance(modules)
	if err := lowering.enumInitialization(modules); err != nil {
		return nil, err
	}
	lowering.noteAccessorNames(modules)
	for _, module := range modules {
		if err := lowering.refuse(module); err != nil {
			return nil, err
		}
	}
	// Link every declaration before lowering any function body, including across back edges.
	var declarations []*ast.Node
	for _, module := range modules {
		declarations = append(declarations, module.Statements.Nodes...)
	}
	if err := lowering.declareModule(declarations); err != nil {
		return nil, err
	}
	for _, module := range modules {
		body, err := lowering.statements(module.Statements.Nodes)
		if err != nil {
			return nil, err
		}
		lowering.result.Main = append(lowering.result.Main, body...)
	}
	if lowering.unlowerable != nil {
		return nil, lowering.unlowerable
	}
	lowering.result.Main = append(lowering.forwarderValues, lowering.result.Main...)
	lowering.finishClassCalls()
	if err := lowering.finishAccessors(); err != nil {
		return nil, err
	}
	if err := lowering.exceptions(); err != nil {
		return nil, err
	}
	if err := lowering.checkAccessorSpreads(); err != nil {
		return nil, err
	}
	if err := flow.NormalizeAsync(lowering.result); err != nil {
		return nil, &NotYet{Where: lowering.program.Where(entry.AsNode()), What: err.Error()}
	}
	if err := run(lowering, modules); err != nil {
		return nil, err
	}
	borrow(lowering.result)
	counters(lowering.result)
	return lowering.result, nil
}

func TestCyclesProbe(t *testing.T) {
	path := os.Getenv("CYCLES_ENTRY")
	if path == "" {
		path = "../../stage1/cohere/markdownblocks/testdata/list_probe.ts"
	}
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	_, err = cyclesProbeLower(context.Background(), program, func(l *lowering, modules []*ast.SourceFile) error {
		if path := os.Getenv("CYCLES_PROFILE"); path != "" {
			file, e := os.Create(path)
			if e != nil {
				t.Fatal(e)
			}
			defer file.Close()
			pprof.StartCPUProfile(file)
			defer pprof.StopCPUProfile()
		}
		start := time.Now()
		err := l.findCycles(modules)
		t.Logf("findCycles: %s", time.Since(start))
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
}

// Getter/setter properties and stored fields can have identical value types.
// Their strong slots differ even when both properties are mutable.
func TestCyclesLiteralAccessorsKeepDataSlots(t *testing.T) {
	path := filepath.Join(t.TempDir(), "main.a")
	source := `interface Item { peer: Holder }
class Holder { item: typeof accessor | undefined = undefined; }
const holder = new Holder();
const item: Item = { peer: holder };
const accessor = {
 get value(): Item { return item; },
 set value(next: Item) { console.log(next.peer === holder ? 'yes' : 'no'); }
};
function supply(): Item { return item; }
const plain = { value: supply() };
holder.item = accessor;
console.log(plain.value.peer === holder ? 'yes' : 'no');
`
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	_, err = cyclesProbeLower(context.Background(), program, func(l *lowering, modules []*ast.SourceFile) error { return l.findCycles(modules) })
	if err != nil {
		t.Fatal(err)
	}
}

// These literals have the same stored value type. Their modifiers must survive
// canonicalization even when fresh writes give their allocations different answers.
func TestCyclesLiteralModifiers(t *testing.T) {
	for _, modifier := range []string{"readonly", "optional"} {
		t.Run(modifier, func(t *testing.T) {
			for _, closes := range []bool{false, true} {
				t.Run(fmt.Sprintf("closes=%t", closes), func(t *testing.T) {
					closing := `{ peer: holder as Holder | undefined }`
					other := `{ peer: new Holder() as Holder | undefined } as const`
					if modifier == "optional" {
						closing = `{ ...requiredSeed }`
						other = `{ ...seed }`
					}
					write := `holder.item = safe;`
					if closes {
						write = `holder.item = closing;`
					}
					path := filepath.Join(t.TempDir(), "main.a")
					source := `class Holder { item: { readonly peer?: Holder | undefined } | undefined = undefined; }
function run(): void {
 const holder = new Holder();
 const seed: { peer?: Holder | undefined } = { peer: new Holder() };
 const requiredSeed: { peer: Holder | undefined } = { peer: holder };
 const closing = ` + closing + `;
 const safe = ` + other + `;
 ` + write + `
 console.log(holder.item === undefined ? 'empty' : 'set');
}
run();
`
					if err := os.WriteFile(path, []byte(source), 0600); err != nil {
						t.Fatal(err)
					}
					program, err := load.Load([]string{path})
					if err != nil {
						t.Fatal(err)
					}
					_, err = cyclesProbeLower(context.Background(), program, func(l *lowering, modules []*ast.SourceFile) error {
						var pair [2]*checker.Type
						var holder *checker.Type
						var visit ast.Visitor
						visit = func(node *ast.Node) bool {
							if node.Kind == ast.KindVariableDeclaration {
								name := node.Name().Text()
								if name == "holder" {
									holder = l.checker.GetTypeAtLocation(node.Name())
								}
								if name == "closing" {
									pair[0] = l.checker.GetTypeAtLocation(node.AsVariableDeclaration().Initializer)
								}
								if name == "safe" {
									pair[1] = l.checker.GetTypeAtLocation(node.AsVariableDeclaration().Initializer)
								}
							}
							return node.ForEachChild(visit)
						}
						modules[0].AsNode().ForEachChild(visit)
						finder := l.cycleTypes(modules)
						for _, shape := range pair {
							if shape == nil {
								t.Fatal("fixture is missing a literal type")
							}
							if shape.ObjectFlags()&checker.ObjectFlagsObjectLiteral == 0 {
								t.Fatalf("fixture did not produce literal types: %s flags=%v", l.checker.TypeToString(shape), shape.ObjectFlags())
							}
						}
						a, b := l.checker.GetPropertiesOfType(pair[0]), l.checker.GetPropertiesOfType(pair[1])
						if len(a) != 1 || len(b) != 1 || a[0].Name != b[0].Name || l.checker.GetTypeOfSymbol(a[0]) != l.checker.GetTypeOfSymbol(b[0]) || pair[0].ObjectFlags() != pair[1].ObjectFlags() {
							t.Fatal("fixture differs beyond property modifiers")
						}
						if modifier == "readonly" {
							if a[0].Flags != b[0].Flags || l.checker.IsReadonlySymbol(a[0]) || !l.checker.IsReadonlySymbol(b[0]) {
								t.Fatal("fixture must differ only in readonly")
							}
						} else if a[0].Flags^b[0].Flags != ast.SymbolFlagsOptional || l.checker.IsReadonlySymbol(a[0]) != l.checker.IsReadonlySymbol(b[0]) {
							t.Fatalf("fixture must differ only in optionality: flags=%v/%v readonly=%t/%t", a[0].Flags, b[0].Flags, l.checker.IsReadonlySymbol(a[0]), l.checker.IsReadonlySymbol(b[0]))
						}
						if finder.literalType(pair[0]) == finder.literalType(pair[1]) {
							t.Fatalf("%s literal shapes share a canonical type", modifier)
						}
						// Pin the Weak requirement before findCycles applies the existing
						// graph-region relaxation to a cycle-capable allocation.
						slotErr := finder.slotsOf(holder)
						if closes {
							refused, ok := slotErr.(*Refused)
							if !ok || !strings.Contains(refused.Fix, "Weak<") {
								t.Fatalf("want cycle-closing write to require Weak, got %v", slotErr)
							}
						} else if slotErr != nil {
							t.Fatalf("independent shape refused: %v", slotErr)
						}
						return l.findCycles(modules)
					})
					if err != nil {
						t.Fatal(err)
					}
				})
			}
		})
	}
}
