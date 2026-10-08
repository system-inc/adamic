package lower

import (
	"context"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
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

// cyclesProbeLower follows Lower through the finder boundary, letting both
// implementations inspect the same checked program and lowered IR.
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
		var err error
		if os.Getenv("CYCLES_OLD") != "" {
			err = l.oldFindCycles(modules)
		} else {
			err = l.findCycles(modules)
		}
		t.Logf("findCycles: %s", time.Since(start))
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
}

// Every fixture is loaded as an entry. Dependencies that are not complete ports
// are still attempted, so adding a new lowering entry needs no inventory edit.
func TestCyclesSameness(t *testing.T) {
	var paths []string
	for _, root := range []string{"../oracle", "../../stage1"} {
		err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() {
				if entry.Name() == "node_modules" {
					return filepath.SkipDir
				}
				return nil
			}
			if strings.HasSuffix(path, ".a") || (strings.HasSuffix(path, ".ts") && !strings.HasSuffix(path, ".d.ts")) {
				paths = append(paths, path)
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	compared, ports := 0, 0
	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			command := exec.Command(os.Args[0], "-test.run=^TestCyclesSamenessFixture$", "-test.v", "-test.timeout=10m")
			command.Env = append(os.Environ(), "CYCLES_FIXTURE="+path, "GOMEMLIMIT=6GiB")
			output, err := command.CombinedOutput()
			if err != nil {
				t.Fatalf("%v\n%s", err, output)
			}
			if strings.Contains(string(output), "cycle finder compared") {
				compared++
				if strings.HasPrefix(path, "../../stage1/") {
					ports++
				}
			}
		})
	}
	if compared == 0 || ports == 0 {
		t.Fatal("corpus did not exercise both fixtures and stage1 ports")
	}
	t.Logf("%d entries attempted, %d finders compared including %d stage1 entries", len(paths), compared, ports)
}

// Each checked entry owns a checker and its relation caches. Isolate entries so
// the frozen quadratic oracle cannot retain caches across the entire corpus.
func TestCyclesSamenessFixture(t *testing.T) {
	path := os.Getenv("CYCLES_FIXTURE")
	if path == "" {
		t.Skip("run through TestCyclesSameness")
	}
	cyclesCompareEntry(t, path, false)
}

func cyclesCompareEntry(t *testing.T, path string, required bool) {
	program, err := load.Load([]string{path})
	if err != nil {
		if required {
			t.Fatal(err)
		}
		t.Logf("does not check: %v", err)
		return
	}

	called := false
	_, err = cyclesProbeLower(context.Background(), program, func(l *lowering, modules []*ast.SourceFile) error {
		called = true
		original := l.result
		l.result = cyclesClone(reflect.ValueOf(original)).Interface().(*ir.Program)
		oldFinder, oldErr := l.oldCycles(modules)
		old := l.result
		l.result = original
		nextFinder := l.cycleTypes(modules)
		nextErr := nextFinder.graphTypes(modules)
		if os.Getenv("CYCLES_MUTANT") == "graph-membership" {
			if l.result.GraphTypes == nil {
				l.result.GraphTypes = map[int]bool{}
			}
			l.result.GraphTypes[-1] = !l.result.GraphTypes[-1]
		}
		if oldFinder != nil && !reflect.DeepEqual(old.GraphTypes, l.result.GraphTypes) {
			if err := cyclesMatchTypeIDs(oldFinder, nextFinder, old, l.result); err != nil {
				return err
			}
		}
		return cyclesSameResult(old, oldErr, l.result, nextErr)
	})
	if !called {
		if required {
			t.Fatalf("required probe did not reach finder: %v", err)
		}
		t.Logf("does not reach finder: %v", err)
		return
	}
	if err != nil {
		t.Fatal(err)
	}
	t.Log("cycle finder compared")
}

func cyclesSameResult(old *ir.Program, oldErr error, next *ir.Program, nextErr error) error {
	if !reflect.DeepEqual(oldErr, nextErr) {
		return fmt.Errorf("finder errors differ: old=%#v new=%#v", oldErr, nextErr)
	}
	if diff := cyclesDifference(reflect.ValueOf(old), reflect.ValueOf(next), "IR"); diff != "" {
		return fmt.Errorf("finder IR differs: %s", diff)
	}
	return nil
}
func TestCyclesSamenessMutant(t *testing.T) {
	command := exec.Command(os.Args[0], "-test.run=^TestCyclesSamenessFixture$", "-test.v", "-test.timeout=10m")
	command.Env = append(os.Environ(), "CYCLES_FIXTURE=../oracle/testdata/fresh_refused/alias_fill.a", "CYCLES_MUTANT=graph-membership")
	output, err := command.CombinedOutput()
	if err == nil || !strings.Contains(string(output), "finder IR differs") {
		t.Fatalf("graph-membership mutant escaped the sameness comparison: %v\n%s", err, output)
	}
}

func cyclesDifference(a, b reflect.Value, path string) string {
	if reflect.DeepEqual(a.Interface(), b.Interface()) {
		return ""
	}
	if a.Type() != b.Type() {
		return path + ": different types"
	}
	switch a.Kind() {
	case reflect.Float64, reflect.Float32:
		if math.IsNaN(a.Float()) && math.IsNaN(b.Float()) {
			return ""
		}
	case reflect.Pointer, reflect.Interface:
		if !a.IsNil() && !b.IsNil() {
			return cyclesDifference(a.Elem(), b.Elem(), path)
		}
	case reflect.Struct:
		for i := 0; i < a.NumField(); i++ {
			if d := cyclesDifference(a.Field(i), b.Field(i), path+"."+a.Type().Field(i).Name); d != "" {
				return d
			}
		}
		return ""
	case reflect.Slice, reflect.Array:
		if a.Len() == b.Len() && (a.Kind() != reflect.Slice || a.IsNil() == b.IsNil()) {
			for i := 0; i < a.Len(); i++ {
				if d := cyclesDifference(a.Index(i), b.Index(i), fmt.Sprintf("%s[%d]", path, i)); d != "" {
					return d
				}
			}
		}
		if a.Len() == b.Len() && (a.Kind() != reflect.Slice || a.IsNil() == b.IsNil()) {
			return ""
		}
	}
	return fmt.Sprintf("%s: old=%v new=%v", path, a.Interface(), b.Interface())
}

// Both finders receive the same lowered program and checker identities. Lowering
// twice would create new literal identities and test unrelated lowering state.
func cyclesClone(value reflect.Value) reflect.Value {
	switch value.Kind() {
	case reflect.Interface:
		if value.IsNil() {
			return reflect.Zero(value.Type())
		}
		made := reflect.New(value.Type()).Elem()
		made.Set(cyclesClone(value.Elem()))
		return made
	case reflect.Pointer:
		if value.IsNil() {
			return reflect.Zero(value.Type())
		}
		made := reflect.New(value.Type().Elem())
		made.Elem().Set(cyclesClone(value.Elem()))
		return made
	case reflect.Struct:
		made := reflect.New(value.Type()).Elem()
		made.Set(value)
		for i := 0; i < value.NumField(); i++ {
			if value.Field(i).CanInterface() {
				made.Field(i).Set(cyclesClone(value.Field(i)))
			}
		}
		return made
	case reflect.Slice:
		if value.IsNil() {
			return reflect.Zero(value.Type())
		}
		made := reflect.MakeSlice(value.Type(), value.Len(), value.Len())
		for i := 0; i < value.Len(); i++ {
			made.Index(i).Set(cyclesClone(value.Index(i)))
		}
		return made
	case reflect.Map:
		if value.IsNil() {
			return reflect.Zero(value.Type())
		}
		made := reflect.MakeMapWithSize(value.Type(), value.Len())
		iter := value.MapRange()
		for iter.Next() {
			made.SetMapIndex(iter.Key(), cyclesClone(iter.Value()))
		}
		return made
	}
	return value
}

// Contextual literal views are fresh checker identities on each query. Compare
// those identities one-to-one by declaration location and the checker's identity
// relation, preserving multiplicity and every selected bit. All original IR
// allocation IDs and negative flow IDs are still compared exactly.
func cyclesMatchTypeIDs(old *oldCycleFinder, next *cycleFinder, a, b *ir.Program) error {
	type view struct {
		proven *checker.Type
		where  *ast.Node
	}
	oldTypes, nextTypes := map[int]view{}, map[int]view{}
	collect := func(where map[*checker.Type]*ast.Node, shapes []*checker.Type, links func(cycleNode) []cycleNode, types map[int]view) {
		visited := map[cycleNode]bool{}
		queue := []cycleNode{}
		for proven, node := range where {
			types[int(proven.Id())] = view{proven, node}
			queue = append(queue, cycleNode{proven: proven})
		}
		for _, proven := range shapes {
			queue = append(queue, cycleNode{proven: proven})
		}
		for len(queue) > 0 {
			node := queue[len(queue)-1]
			queue = queue[:len(queue)-1]
			if visited[node] {
				continue
			}
			visited[node] = true
			if node.proven != nil {
				if _, ok := types[int(node.proven.Id())]; !ok {
					types[int(node.proven.Id())] = view{proven: node.proven}
				}
			}
			queue = append(queue, links(node)...)
		}
	}
	collect(old.where, old.shapes, old.graphLinks, oldTypes)
	collect(next.where, next.shapes, next.graphLinks, nextTypes)
	remapped := map[int]bool{}
	used := map[int]bool{}
	processed := map[int]bool{}
	for id, selected := range b.GraphTypes {
		if id <= 0 || a.GraphTypes[id] == selected && oldTypes[id].proven != nil && oldTypes[id].proven == nextTypes[id].proven {
			remapped[id] = selected
			used[id] = true
			processed[id] = true
		}
	}
	for id, selected := range b.GraphTypes {
		if processed[id] {
			continue
		}
		right, ok := nextTypes[id]
		if !ok {
			return fmt.Errorf("new graph identity %d has no type", id)
		}
		matched := false
		for prior, wanted := range a.GraphTypes {
			if used[prior] || wanted != selected {
				continue
			}
			left, ok := oldTypes[prior]
			if !ok {
				continue
			}
			if left.where == right.where && left.proven.Flags() == right.proven.Flags() && left.proven.ObjectFlags() == right.proven.ObjectFlags() && checker.Checker_isTypeIdenticalTo(next.l.checker, left.proven, right.proven) {
				remapped[prior] = selected
				used[prior] = true
				matched = true
				break
			}
		}
		if !matched {
			return fmt.Errorf("new graph type %d (%s) has no identical old view", id, next.l.checker.TypeToString(right.proven))
		}
	}
	if len(remapped) != len(a.GraphTypes) {
		return fmt.Errorf("graph membership counts differ: old=%d new=%d", len(a.GraphTypes), len(remapped))
	}
	b.GraphTypes = remapped
	return nil
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
	cyclesCompareEntry(t, path, true)
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
