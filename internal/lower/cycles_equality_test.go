package lower

import (
	"context"
	"errors"
	"fmt"
	"hash/fnv"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/native"
)

var cycleComparisonInputs = sync.OnceValues(func() ([]string, error) {
	var paths []string
	err := filepath.WalkDir("../..", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relative, err := filepath.Rel("../..", path)
		if err != nil {
			return err
		}
		relative = filepath.ToSlash(relative)
		if entry.IsDir() {
			if entry.Name() == ".git" || relative == "cohere" {
				return filepath.SkipDir
			}
			return nil
		}
		// The scanner reference for this size does not fit in a house box's
		// memory. The four other large stress inputs from the same generators
		// are compared; keep this exclusion explicit and limited to this name.
		if relative == "review/compiler/loom-implied-target/followup/objects-4096.a" {
			return nil
		}
		fixture := strings.HasPrefix(relative, "internal/oracle/testdata/") || strings.HasPrefix(relative, "internal/lower/testdata/") || strings.HasPrefix(relative, "review/") || strings.HasPrefix(relative, "stage3/fixtures/cycles/")
		extension := filepath.Ext(path)
		if extension == ".a" || fixture && (extension == ".ts" || extension == ".tsx") {
			paths = append(paths, relative)
		}
		return nil
	})
	slices.Sort(paths)
	return paths, err
})

const cycleEqualityShardCount = 128

func cycleComparisonShard(path string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(path))
	return h.Sum64() % cycleEqualityShardCount
}

// Compare the entire IR as well as refusal identity, rather than just admission.
// This includes every graph/Program flag, membership-by-type map, frame, cell,
// weak field and ownership-query consequence carried by the lowered program.
// These full Program audits exceed the loaded unit budget. They remain
// reproducible with ADAMIC_CYCLE_PROGRAM_STRESS=1, -test.timeout=15m and an
// independent 900-second limit per input. Ordinary mode stays in the unit shards.
var cycleProgramStressInputs = map[string]bool{
	"review/compiler/loom-implied-target/followup/objects-2048.a": true,
	"review/compiler/loom-implied-target/objects.a":               true,
	"review/compiler/loom-implied-target/objects-sequential.a":    true,
	"internal/oracle/testdata/moves/accepted/objects.a":           true,
}

func cycleCompareProgram(t *testing.T, path string) {
	t.Helper()
	regions := []bool{false, true}
	if cycleProgramStressInputs[path] && os.Getenv("ADAMIC_CYCLE_PROGRAM_STRESS") != "1" {
		regions = []bool{false}
		t.Log("budget: Program stress audit is opt-in with ADAMIC_CYCLE_PROGRAM_STRESS=1 and a per-input 900-second limit")
	}
	cycleCompareProgramModes(t, path, regions...)
}

func cycleCompareProgramModes(t *testing.T, path string, regions ...bool) {
	t.Helper()
	absolute, err := filepath.Abs(filepath.Join("../..", path))
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := load.Load([]string{absolute})
	if err != nil {
		t.Logf("%s: common loader refusal: %v", path, err)
		return
	}
	for mode, region := range regions {
		// Use fresh, independently initialized checkers. Reusing one checker changes
		// fresh literal type IDs on the second lowering, even for scanner vs scanner.
		if mode > 0 {
			loaded, err = load.Load([]string{absolute})
			if err != nil {
				t.Fatal(err)
			}
		}
		options := Options{ProgramRegion: region, OwnershipQuery: true}
		before, scanTypes, scanErr := cycleLowerForComparison(loaded, options, false)
		// Only the IR and resolved strings are compared below. Do not retain the
		// scanning checker's quadratic relation cache while building the indexed copy.
		loaded = nil
		if cycleProgramStressInputs[path] {
			t.Logf("%s: scanning snapshot ready, ProgramRegion=%t; release its checker", path, region)
			runtime.GC()
		}
		indexedInput, err := load.Load([]string{absolute})
		if err != nil {
			t.Fatal(err)
		}
		after, indexTypes, indexErr := cycleLowerForComparison(indexedInput, options, true)
		if cycleProgramStressInputs[path] {
			t.Logf("%s: indexed snapshot ready, ProgramRegion=%t", path, region)
		}
		text := func(err error) string {
			if err == nil {
				return ""
			}
			return fmt.Sprintf("%T: %v", err, err)
		}
		var scanRefused, indexRefused *Refused
		errors.As(scanErr, &scanRefused)
		errors.As(indexErr, &indexRefused)
		cycle := func(r *Refused) bool { return r != nil && r.cycle }
		if text(scanErr) != text(indexErr) || cycle(scanRefused) != cycle(indexRefused) {
			t.Fatalf("cycle equality difference: %s ProgramRegion=%t\nscanning: %s (cycle=%t)\nindexed: %s (cycle=%t)", path, region, text(scanErr), cycle(scanRefused), text(indexErr), cycle(indexRefused))
		}
		if difference := cycleFirstDifferenceResolved(reflect.ValueOf(before), reflect.ValueOf(after), "program", scanTypes, indexTypes); difference != "" {
			t.Fatalf("cycle equality difference: %s ProgramRegion=%t: %s", path, region, difference)
		}
		if scanErr == nil && native.C(before) != native.C(after) {
			t.Fatalf("cycle equality difference: %s ProgramRegion=%t: emitted C differs", path, region)
		}
	}
	t.Logf("%s: equal scanning/indexed outcomes in modes %v", path, regions)
}

// Checker IDs are assigned lazily, and indexing changes query order. Compare
// GraphTypes, ProgramTypes and the checker-keyed function target table by
// resolved type and declaration identity, never ID.
func cycleLowerForComparison(program *load.Program, options Options, indexed bool) (*ir.Program, map[int]string, error) {
	entry := program.Files()[0]
	checked, release := program.Checker(context.Background(), entry)
	defer release()
	types := map[int]*checker.Type{}
	lower := scanLowerCheckedWithCycleTypes
	if indexed {
		// The frozen scanning reference compares unmarked IR.
		lower = func(program *load.Program, checked *checker.Checker, entry *ast.SourceFile, plan *programMembership, discovery bool, types map[int]*checker.Type, ownershipQuery ...bool) (*lowering, error) {
			return lowerCheckedWithCycleTypes(program, checked, entry, plan, discovery, false, types, ownershipQuery...)
		}
	}
	var plan *programMembership
	if options.ProgramRegion {
		discovery, err := lower(program, checked, entry, nil, true, types)
		if err != nil {
			return nil, nil, err
		}
		plan = discovery.programPlan
	}
	built, err := lower(program, checked, entry, plan, false, types, options.OwnershipQuery)
	if err != nil {
		return nil, nil, err
	}
	resolved := map[int]string{}
	for id, proven := range types {
		identity := checked.TypeToStringEx(proven, nil, checker.TypeFormatFlagsNoTruncation, nil)
		if symbol := proven.Symbol(); symbol != nil {
			name := ast.SymbolName(symbol)
			// Computed names carry checker-local unique-symbol numbers. Their
			// source spelling and declaration spans identify the same symbol across checkers.
			if len(symbol.Declarations) != 0 {
				if declared := symbol.Declarations[0].Name(); declared != nil {
					if source := ast.GetSourceFileOfNode(declared); source != nil {
						name = strings.TrimSpace(source.Text()[declared.Pos():declared.End()])
					}
				}
			}
			identity += " symbol=" + name
			for _, declaration := range symbol.Declarations {
				identity += fmt.Sprintf(" %s:%d:%d", program.Where(declaration), declaration.Pos(), declaration.End())
			}
		}
		resolved[id] = identity
	}
	return built.result, resolved, nil
}

func cycleFirstDifferenceResolved(a, b reflect.Value, path string, scanTypes, indexTypes map[int]string) string {
	if !a.IsValid() || !b.IsValid() {
		return fmt.Sprintf("%s: validity %t/%t", path, a.IsValid(), b.IsValid())
	}
	if a.Type() != b.Type() {
		return fmt.Sprintf("%s: types %v/%v", path, a.Type(), b.Type())
	}
	if strings.HasSuffix(path, ".GraphTypes") || strings.HasSuffix(path, ".ProgramTypes") || strings.HasSuffix(path, ".FunctionTypeTargets") {
		resolve := func(value reflect.Value, types map[int]string) ([]string, error) {
			var result []string
			add := func(id int, suffix string) error {
				// Negative values are deterministic IR allocation-site identities,
				// not lazily numbered checker types; preserve them exactly.
				if id < 0 {
					result = append(result, fmt.Sprintf("allocation-site:%d%s", id, suffix))
					return nil
				}
				identity, ok := types[id]
				if !ok {
					return fmt.Errorf("unresolved checker type %d", id)
				}
				result = append(result, identity+suffix)
				return nil
			}
			if value.Kind() == reflect.Map {
				for _, key := range value.MapKeys() {
					if err := add(int(key.Int()), fmt.Sprintf("=%#v", value.MapIndex(key).Interface())); err != nil {
						return nil, err
					}
				}
			} else {
				for i := 0; i < value.Len(); i++ {
					if err := add(int(value.Index(i).Int()), ""); err != nil {
						return nil, err
					}
				}
			}
			slices.Sort(result)
			return result, nil
		}
		left, err := resolve(a, scanTypes)
		if err != nil {
			return path + ": scanning " + err.Error()
		}
		right, err := resolve(b, indexTypes)
		if err != nil {
			return path + ": indexed " + err.Error()
		}
		if !slices.Equal(left, right) {
			return fmt.Sprintf("%s: scanning=%v indexed=%v", path, left, right)
		}
		return ""
	}
	if a.Kind() != reflect.Pointer && a.Kind() != reflect.Interface && a.Kind() != reflect.Struct && a.Kind() != reflect.Slice && a.Kind() != reflect.Array && a.Kind() != reflect.Map && a.CanInterface() && b.CanInterface() && reflect.DeepEqual(a.Interface(), b.Interface()) {
		return ""
	}
	switch a.Kind() {
	case reflect.Pointer, reflect.Interface:
		if a.IsNil() && b.IsNil() {
			return ""
		}
		if a.IsNil() || b.IsNil() {
			return fmt.Sprintf("%s: nil scanning=%t indexed=%t", path, a.IsNil(), b.IsNil())
		}
		return cycleFirstDifferenceResolved(a.Elem(), b.Elem(), path, scanTypes, indexTypes)
	case reflect.Struct:
		for i := 0; i < a.NumField(); i++ {
			if !a.Field(i).CanInterface() {
				continue
			}
			if d := cycleFirstDifferenceResolved(a.Field(i), b.Field(i), path+"."+a.Type().Field(i).Name, scanTypes, indexTypes); d != "" {
				return d
			}
		}
	case reflect.Slice, reflect.Array:
		if a.Len() != b.Len() {
			return fmt.Sprintf("%s: scanning length=%d indexed length=%d", path, a.Len(), b.Len())
		}
		for i := 0; i < a.Len(); i++ {
			if d := cycleFirstDifferenceResolved(a.Index(i), b.Index(i), fmt.Sprintf("%s[%d]", path, i), scanTypes, indexTypes); d != "" {
				return d
			}
		}
	case reflect.Map:
		if a.Len() != b.Len() {
			return fmt.Sprintf("%s: scanning=%v indexed=%v", path, a.Interface(), b.Interface())
		}
		for _, key := range a.MapKeys() {
			if d := cycleFirstDifferenceResolved(a.MapIndex(key), b.MapIndex(key), fmt.Sprintf("%s[%v]", path, key), scanTypes, indexTypes); d != "" {
				return d
			}
		}
	case reflect.Float32, reflect.Float64:
		if math.IsNaN(a.Float()) && math.IsNaN(b.Float()) {
			return ""
		}
		return fmt.Sprintf("%s: scanning=%v indexed=%v", path, a.Float(), b.Float())
	default:
		if a.CanInterface() && b.CanInterface() {
			return fmt.Sprintf("%s: scanning=%#v indexed=%#v", path, a.Interface(), b.Interface())
		}
	}
	return "" // Unexported IR analysis caches do not carry ownership answers.
}

func cycleComparisonShardPaths(paths []string, shard uint64) []string {
	var selected []string
	for _, path := range paths {
		if cycleComparisonShard(path) == shard {
			selected = append(selected, path)
		}
	}
	return selected
}

func runCycleFinderEqualityShard(t *testing.T, shard uint64) {
	t.Helper()
	paths, err := cycleComparisonInputs()
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) < 3500 {
		t.Fatalf("cycle comparison inventory shrank: %d", len(paths))
	}
	for _, path := range cycleComparisonShardPaths(paths, shard) {
		if !t.Run(path, func(t *testing.T) { cycleCompareProgramRecorded(t, path) }) {
			return
		}
	}
}

func TestCycleIndexCandidateEdgeEquality(t *testing.T) {
	t.Parallel()
	indexed, types := cycleIndexTypes(t, `declare const view:{visible:number};declare const actual:{visible:number;hidden:{target:true}};declare const target:{target:true};`)
	indexed.shapes = []*checker.Type{types["actual"]}
	scanning := &scanCycleFinder{l: indexed.l, where: indexed.where, shapes: indexed.shapes}
	target := cycleNode{proven: types["target"]}
	before, after := scanning.reaches(types["view"], target), indexed.reaches(types["view"], target)
	if !before || cycleFirstDifferenceResolved(reflect.ValueOf(before), reflect.ValueOf(after), "candidate edge", nil, nil) != "" {
		t.Fatalf("candidate-edge equality: scanning=%t indexed=%t", before, after)
	}
}

func TestCycleIndexArgumentsFunctionEquality(t *testing.T) {
	t.Parallel()
	cycleCompareProgram(t, "internal/lower/testdata/cycle_index_arguments_function.a")
}

func TestCycleIndexLibraryFunctionLinks(t *testing.T) {
	t.Parallel()
	absolute, err := filepath.Abs("testdata/cycle_index_arguments_function.a")
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := load.Load([]string{absolute})
	if err != nil {
		t.Fatal(err)
	}
	checked, release := loaded.Checker(context.Background(), loaded.Files()[0])
	defer release()
	built, err := scanLowerChecked(loaded, checked, loaded.Files()[0], nil, true)
	if err != nil {
		t.Fatal(err)
	}
	var arguments *ast.Node
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if ast.IsIdentifier(node) && node.Text() == "arguments" {
			arguments = node
		}
		return node.ForEachChild(visit)
	}
	loaded.Files()[0].AsNode().ForEachChild(visit)
	if arguments == nil {
		t.Fatal("missing arguments reference")
	}
	holder := checked.GetTypeAtLocation(arguments)
	callee := checked.GetPropertyOfType(holder, "callee")
	if callee == nil {
		t.Fatal("IArguments.callee is missing")
	}
	function := checked.GetTypeOfSymbol(callee)
	indexed := &cycleFinder{l: built, where: map[*checker.Type]*ast.Node{}}
	scanning := &scanCycleFinder{l: built, where: map[*checker.Type]*ast.Node{}}
	indexed.use(holder, arguments)
	scanning.use(holder, arguments)
	t.Logf("Function shape candidate: scanning=%t indexed=%t; cached function classification=%t", slices.Contains(scanning.shapes, function), slices.Contains(indexed.shapes, function), indexed.functionCache[function])
	calls := len(checked.GetSignaturesOfType(function, checker.SignatureKindCall))
	constructs := len(checked.GetSignaturesOfType(function, checker.SignatureKindConstruct))
	t.Logf("arguments at %s -> %s.callee -> %s; call signatures=%d construct signatures=%d", loaded.Where(arguments), checked.TypeToString(holder), checked.TypeToString(function), calls, constructs)
	if calls != 0 || constructs != 0 || !built.isLibraryType(function, "Function") {
		t.Fatal("fixture lost its signature-free library Function")
	}
	if _, ok := indexed.where[function]; !ok {
		t.Fatal("indexed discovery omitted IArguments.callee")
	}
	if _, ok := scanning.where[function]; !ok {
		t.Fatal("scanning discovery omitted IArguments.callee")
	}
	for _, field := range indexed.fields(function) {
		t.Logf("Function field %s -> %s", field.Name, checked.TypeToString(checked.GetTypeOfSymbol(field)))
	}
	before, after := scanning.programOwningLinks(cycleNode{proven: function}), indexed.programOwningLinks(cycleNode{proven: function})
	t.Logf("Function classified callable: scanning=%t indexed=%t; owning links: scanning=%d indexed=%d", scanning.isFunction(function), indexed.isFunction(function), len(before), len(after))
	for _, link := range after {
		if link.proven != nil {
			t.Logf("indexed Function owns %s", checked.TypeToString(link.proven))
		}
	}
	if !scanning.isFunction(function) || indexed.isFunction(function) != scanning.isFunction(function) || len(before) != 0 || !slices.Equal(before, after) {
		t.Fatal("library Function must use the scanning finder's closure-only ownership path")
	}
}

func TestCycleIndexCountTargetsEquality(t *testing.T) {
	t.Parallel()
	cycleCompareProgram(t, "internal/oracle/testdata/miscompile_2a/controls.a")
}

func TestCycleIndexResolvedTargetsKeepAnswers(t *testing.T) {
	t.Parallel()
	type program struct{ FunctionTypeTargets map[int][]int }
	scanning := program{map[int][]int{1: {0, 1}}}
	indexed := program{map[int][]int{2: {0, 1}}}
	before, after := map[int]string{1: "resolved function declaration"}, map[int]string{2: "resolved function declaration"}
	compare := func() string {
		return cycleFirstDifferenceResolved(reflect.ValueOf(scanning), reflect.ValueOf(indexed), "program", before, after)
	}
	if difference := compare(); difference != "" {
		t.Fatal(difference)
	}
	indexed.FunctionTypeTargets[2] = []int{0}
	if compare() == "" {
		t.Fatal("resolving type keys hid a changed function target set")
	}
}

func TestCycleIndexIteratorTargetsEquality(t *testing.T) {
	t.Parallel()
	cycleCompareProgram(t, "internal/oracle/testdata/user_iterators.a")
}
