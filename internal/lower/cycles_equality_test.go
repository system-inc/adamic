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
	walkError := filepath.WalkDir("../..", func(path string, entry fs.DirEntry, pathError error) error {
		if pathError != nil {
			return pathError
		}
		relative, relativePathError := filepath.Rel("../..", path)
		if relativePathError != nil {
			return relativePathError
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
		fixture := strings.HasPrefix(relative, "internal/oracle/testdata/") ||
			strings.HasPrefix(relative, "internal/lower/testdata/") ||
			strings.HasPrefix(relative, "review/") ||
			strings.HasPrefix(relative, "stage3/fixtures/cycles/")
		extension := filepath.Ext(path)
		if extension == ".a" || fixture && (extension == ".ts" || extension == ".tsx") {
			paths = append(paths, relative)
		}
		return nil
	})
	slices.Sort(paths)
	return paths, walkError
})

const cycleEqualityShardCount = 128

func cycleComparisonShard(path string) uint64 {
	hash64 := fnv.New64a()
	hash64.Write([]byte(path))
	return hash64.Sum64() % cycleEqualityShardCount
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

func cycleCompareProgram(test *testing.T, path string) {
	test.Helper()
	regions := []bool{false, true}
	if cycleProgramStressInputs[path] && os.Getenv("ADAMIC_CYCLE_PROGRAM_STRESS") != "1" {
		regions = []bool{false}
		test.Log(
			"budget: Program stress audit is opt-in with ADAMIC_CYCLE_PROGRAM_STRESS=1 and a per-input 900-second limit",
		)
	}
	cycleCompareProgramModes(test, path, regions...)
}

// cycleFrozenGeneratorGuard is the frozen scanning reference's stop for a Program region in a program
// with generators, a guard the real pipeline lifted (#8xcz03h). With the region on, the reference
// still stops there, so that mode has nothing to compare against.
const cycleFrozenGeneratorGuard = "a Program region in a program with generator functions (#qk8rztp)"

func cycleCompareProgramModes(test *testing.T, path string, regions ...bool) {
	test.Helper()
	absolute, absolutePathError := filepath.Abs(filepath.Join("../..", path))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	loaded, loadError := load.Load([]string{absolute})
	if loadError != nil {
		test.Logf("%s: common loader refusal: %v", path, loadError)
		return
	}
	for mode, region := range regions {
		// Use fresh, independently initialized checkers. Reusing one checker changes
		// fresh literal type IDs on the second lowering, even for scanner vs scanner.
		if mode > 0 {
			loaded, loadError = load.Load([]string{absolute})
			if loadError != nil {
				test.Fatal(loadError)
			}
		}
		options := Options{ProgramRegion: region, OwnershipQuery: true}
		before, scanTypes, scanError := cycleLowerForComparison(loaded, options, false)
		// Only the IR and resolved strings are compared below. Do not retain the
		// scanning checker's quadratic relation cache while building the indexed copy.
		loaded = nil
		var frozenGuard *NotYet
		if region && errors.As(scanError, &frozenGuard) && frozenGuard.What == cycleFrozenGeneratorGuard {
			test.Logf(
				"%s: ProgramRegion=true not compared: the frozen scanning reference refuses a Program region with generators, a guard the real pipeline lifted in #8xcz03h",
				path,
			)
			continue
		}
		if cycleProgramStressInputs[path] {
			test.Logf("%s: scanning snapshot ready, ProgramRegion=%t; release its checker", path, region)
			runtime.GC()
		}
		indexedInput, loadError := load.Load([]string{absolute})
		if loadError != nil {
			test.Fatal(loadError)
		}
		after, indexTypes, indexError := cycleLowerForComparison(indexedInput, options, true)
		if cycleProgramStressInputs[path] {
			test.Logf("%s: indexed snapshot ready, ProgramRegion=%t", path, region)
		}
		text := func(sourceError error) string {
			if sourceError == nil {
				return ""
			}
			return fmt.Sprintf("%T: %v", sourceError, sourceError)
		}
		var scanRefused, indexRefused *Refused
		errors.As(scanError, &scanRefused)
		errors.As(indexError, &indexRefused)
		cycle := func(refused *Refused) bool { return refused != nil && refused.cycle }
		if text(scanError) != text(indexError) || cycle(scanRefused) != cycle(indexRefused) {
			test.Fatalf(
				"cycle equality difference: %s ProgramRegion=%t\nscanning: %s (cycle=%t)\nindexed: %s (cycle=%t)",
				path,
				region,
				text(scanError),
				cycle(scanRefused),
				text(indexError),
				cycle(indexRefused),
			)
		}
		if difference := cycleFirstDifferenceResolved(
			reflect.ValueOf(before),
			reflect.ValueOf(after),
			"program",
			scanTypes,
			indexTypes,
		); difference != "" {
			test.Fatalf("cycle equality difference: %s ProgramRegion=%t: %s", path, region, difference)
		}
		if scanError == nil && native.C(before) != native.C(after) {
			test.Fatalf("cycle equality difference: %s ProgramRegion=%t: emitted C differs", path, region)
		}
	}
	test.Logf("%s: equal scanning/indexed outcomes in modes %v", path, regions)
}

// Checker IDs are assigned lazily, and indexing changes query order. Compare
// GraphTypes, ProgramTypes and the checker-keyed function target table by
// resolved type and declaration identity, never ID.
func cycleLowerForComparison(
	program *load.Program,
	options Options,
	indexed bool,
) (*ir.Program, map[int]string, error) {
	entry := program.Files()[0]
	checked, release := program.Checker(context.Background(), entry)
	defer release()
	types := map[int]*checker.Type{}
	lower := scanLowerCheckedWithCycleTypes
	if indexed {
		// The frozen scanning reference compares unmarked IR.
		lower = func(
			program *load.Program,
			checked *checker.Checker,
			entry *ast.SourceFile,
			plan *programMembership,
			discovery bool,
			types map[int]*checker.Type,
			ownershipQuery ...bool,
		) (*lowering, error) {
			return lowerCheckedWithCycleTypes(program, checked, entry, plan, discovery, false, types, ownershipQuery...)
		}
	}
	var plan *programMembership
	if options.ProgramRegion {
		discovery, lowerError := lower(program, checked, entry, nil, true, types)
		if lowerError != nil {
			return nil, nil, lowerError
		}
		plan = discovery.programPlan
	}
	built, lowerError := lower(program, checked, entry, plan, false, types, options.OwnershipQuery)
	if lowerError != nil {
		return nil, nil, lowerError
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
	if strings.HasSuffix(path, ".GraphTypes") ||
		strings.HasSuffix(path, ".ProgramTypes") ||
		strings.HasSuffix(path, ".FunctionTypeTargets") {
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
					if addError := add(
						int(key.Int()),
						fmt.Sprintf("=%#v", value.MapIndex(key).Interface()),
					); addError != nil {
						return nil, addError
					}
				}
			} else {
				for index := 0; index < value.Len(); index++ {
					if addError := add(int(value.Index(index).Int()), ""); addError != nil {
						return nil, addError
					}
				}
			}
			slices.Sort(result)
			return result, nil
		}
		left, resolveError := resolve(a, scanTypes)
		if resolveError != nil {
			return path + ": scanning " + resolveError.Error()
		}
		right, resolveError := resolve(b, indexTypes)
		if resolveError != nil {
			return path + ": indexed " + resolveError.Error()
		}
		if !slices.Equal(left, right) {
			return fmt.Sprintf("%s: scanning=%v indexed=%v", path, left, right)
		}
		return ""
	}
	if a.Kind() != reflect.Pointer &&
		a.Kind() != reflect.Interface &&
		a.Kind() != reflect.Struct &&
		a.Kind() != reflect.Slice &&
		a.Kind() != reflect.Array &&
		a.Kind() != reflect.Map &&
		a.CanInterface() &&
		b.CanInterface() &&
		reflect.DeepEqual(a.Interface(), b.Interface()) {
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
		for index := 0; index < a.NumField(); index++ {
			if !a.Field(index).CanInterface() {
				continue
			}
			if d := cycleFirstDifferenceResolved(
				a.Field(index),
				b.Field(index),
				path+"."+a.Type().Field(index).Name,
				scanTypes,
				indexTypes,
			); d != "" {
				return d
			}
		}
	case reflect.Slice, reflect.Array:
		if a.Len() != b.Len() {
			return fmt.Sprintf("%s: scanning length=%d indexed length=%d", path, a.Len(), b.Len())
		}
		for index := 0; index < a.Len(); index++ {
			if d := cycleFirstDifferenceResolved(
				a.Index(index),
				b.Index(index),
				fmt.Sprintf("%s[%d]", path, index),
				scanTypes,
				indexTypes,
			); d != "" {
				return d
			}
		}
	case reflect.Map:
		if a.Len() != b.Len() {
			return fmt.Sprintf("%s: scanning=%v indexed=%v", path, a.Interface(), b.Interface())
		}
		for _, key := range a.MapKeys() {
			if d := cycleFirstDifferenceResolved(
				a.MapIndex(key),
				b.MapIndex(key),
				fmt.Sprintf("%s[%v]", path, key),
				scanTypes,
				indexTypes,
			); d != "" {
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

func runCycleFinderEqualityShard(test *testing.T, shard uint64) {
	test.Helper()
	paths, cycleComparisonInputsError := cycleComparisonInputs()
	if cycleComparisonInputsError != nil {
		test.Fatal(cycleComparisonInputsError)
	}
	if len(paths) < 3500 {
		test.Fatalf("cycle comparison inventory shrank: %d", len(paths))
	}
	for _, path := range cycleComparisonShardPaths(paths, shard) {
		if !test.Run(path, func(test *testing.T) { cycleCompareProgramRecorded(test, path) }) {
			return
		}
	}
}

func TestCycleIndexCandidateEdgeEquality(test *testing.T) {
	test.Parallel()
	indexed, types := cycleIndexTypes(
		test,
		`declare const view:{visible:number};declare const actual:{visible:number;hidden:{target:true}};declare const target:{target:true};`,
	)
	indexed.shapes = []*checker.Type{types["actual"]}
	scanning := &scanCycleFinder{lowering: indexed.lowering, where: indexed.where, shapes: indexed.shapes}
	target := cycleNode{proven: types["target"]}
	before, after := scanning.reaches(types["view"], target), indexed.reaches(types["view"], target)
	if !before || cycleFirstDifferenceResolved(
		reflect.ValueOf(before),
		reflect.ValueOf(after),
		"candidate edge",
		nil,
		nil,
	) != "" {
		test.Fatalf("candidate-edge equality: scanning=%t indexed=%t", before, after)
	}
}

func TestCycleIndexArgumentsFunctionEquality(test *testing.T) {
	test.Parallel()
	cycleCompareProgram(test, "internal/lower/testdata/cycle_index_arguments_function.a")
}

func TestCycleIndexLibraryFunctionLinks(test *testing.T) {
	test.Parallel()
	absolute, absolutePathError := filepath.Abs("testdata/cycle_index_arguments_function.a")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	loaded, loadError := load.Load([]string{absolute})
	if loadError != nil {
		test.Fatal(loadError)
	}
	checked, release := loaded.Checker(context.Background(), loaded.Files()[0])
	defer release()
	built, scanLowerCheckedError := scanLowerChecked(loaded, checked, loaded.Files()[0], nil, true)
	if scanLowerCheckedError != nil {
		test.Fatal(scanLowerCheckedError)
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
		test.Fatal("missing arguments reference")
	}
	holder := checked.GetTypeAtLocation(arguments)
	callee := checked.GetPropertyOfType(holder, "callee")
	if callee == nil {
		test.Fatal("IArguments.callee is missing")
	}
	function := checked.GetTypeOfSymbol(callee)
	indexed := &cycleFinder{lowering: built, where: map[*checker.Type]*ast.Node{}}
	scanning := &scanCycleFinder{lowering: built, where: map[*checker.Type]*ast.Node{}}
	indexed.use(holder, arguments)
	scanning.use(holder, arguments)
	test.Logf(
		"Function shape candidate: scanning=%t indexed=%t; cached function classification=%t",
		slices.Contains(scanning.shapes, function),
		slices.Contains(indexed.shapes, function),
		indexed.functionCache[function],
	)
	calls := len(checked.GetSignaturesOfType(function, checker.SignatureKindCall))
	constructs := len(checked.GetSignaturesOfType(function, checker.SignatureKindConstruct))
	test.Logf(
		"arguments at %s -> %s.callee -> %s; call signatures=%d construct signatures=%d",
		loaded.Where(arguments),
		checked.TypeToString(holder),
		checked.TypeToString(function),
		calls,
		constructs,
	)
	if calls != 0 || constructs != 0 || !built.isLibraryType(function, "Function") {
		test.Fatal("fixture lost its signature-free library Function")
	}
	if _, ok := indexed.where[function]; !ok {
		test.Fatal("indexed discovery omitted IArguments.callee")
	}
	if _, ok := scanning.where[function]; !ok {
		test.Fatal("scanning discovery omitted IArguments.callee")
	}
	for _, field := range indexed.fields(function) {
		test.Logf("Function field %s -> %s", field.Name, checked.TypeToString(checked.GetTypeOfSymbol(field)))
	}
	before, after := scanning.programOwningLinks(
		cycleNode{proven: function},
	), indexed.programOwningLinks(cycleNode{proven: function})
	test.Logf(
		"Function classified callable: scanning=%t indexed=%t; owning links: scanning=%d indexed=%d",
		scanning.isFunction(function),
		indexed.isFunction(function),
		len(before),
		len(after),
	)
	for _, link := range after {
		if link.proven != nil {
			test.Logf("indexed Function owns %s", checked.TypeToString(link.proven))
		}
	}
	if !scanning.isFunction(function) ||
		indexed.isFunction(function) != scanning.isFunction(function) ||
		len(before) != 0 ||
		!slices.Equal(before, after) {
		test.Fatal("library Function must use the scanning finder's closure-only ownership path")
	}
}

func TestCycleIndexCountTargetsEquality(test *testing.T) {
	test.Parallel()
	cycleCompareProgram(test, "internal/oracle/testdata/miscompile_2a/controls.a")
}

func TestCycleIndexResolvedTargetsKeepAnswers(test *testing.T) {
	test.Parallel()
	type program struct{ FunctionTypeTargets map[int][]int }
	scanning := program{map[int][]int{1: {0, 1}}}
	indexed := program{map[int][]int{2: {0, 1}}}
	before, after := map[int]string{
		1: "resolved function declaration",
	}, map[int]string{2: "resolved function declaration"}
	compare := func() string {
		return cycleFirstDifferenceResolved(
			reflect.ValueOf(scanning),
			reflect.ValueOf(indexed),
			"program",
			before,
			after,
		)
	}
	if difference := compare(); difference != "" {
		test.Fatal(difference)
	}
	indexed.FunctionTypeTargets[2] = []int{0}
	if compare() == "" {
		test.Fatal("resolving type keys hid a changed function target set")
	}
}

func TestCycleIndexIteratorTargetsEquality(test *testing.T) {
	test.Parallel()
	cycleCompareProgram(test, "internal/oracle/testdata/user_iterators.a")
}
