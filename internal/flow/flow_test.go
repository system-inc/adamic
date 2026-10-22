package flow

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// Lower is lower.Lower, set by lower_test.go: lower depends on this package (the cycle finder's
// proof of fresh writes is built on the graph), so only an external test can import it.
var Lower func(context.Context, *load.Program) (*ir.Program, error)

// programs is every program the oracle runs that lowers, plus this package's own: the graph is built
// for each of their functions and their top level.
func programs(t *testing.T) []string {
	t.Helper()
	var paths []string
	for _, pattern := range []string{
		"../../dedication/dedication.a",
		"../load/testdata/0.1/compile/*.ts",
		"../load/testdata/0.1/compile/07_modules/main.ts",
		"../oracle/testdata/*.a",
		"../oracle/testdata/non_null*.ts",
		"../oracle/testdata/modules/main.a",
		"testdata/*.a",
	} {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			t.Fatal(err)
		}
		paths = append(paths, matches...)
	}
	// A program that never ends on its own (stopped from outside by a signal, output_test.go) has no
	// trace to walk. size_class_churn.a churns 675,000 values through the allocator, and its trace of
	// every point its loops pass doesn't finish within the trace's five minutes even alone; its loops
	// are shapes the other programs trace, and its size is what it's for. bitwise_sweep.a records
	// over a million points and exceeds the same limit in the full gate; bitwise.a covers its loop
	// and operator shapes here, and the oracle still runs the full sweep. The large typed-array
	// sieve prints a 100,001-element buffer at every traced point and exceeds that limit;
	// typed_arrays_primes.a traces the same sieve through limit 100, and the oracle runs both.
	// normalize_coverage_long.a observes every point of million-unit normalized strings.
	// Its trace would record hundreds of millions of instructions; the smaller normalization
	// fixtures cover the same loop shapes here, and the oracle still runs the long program.
	paths = slices.DeleteFunc(paths, func(path string) bool {
		return runtimeBoundaryLoweringReason(path) != "" || uninstrumentedAsyncFixture(path) || refusedNonNullFixture(path) || refusedConstructorFixture(path) || refusedGettersCensusFixture(path) || filepath.Base(path) == "killed_after_output.a" || filepath.Base(path) == "size_class_churn.a" || filepath.Base(path) == "bitwise_sweep.a" || filepath.Base(path) == "typed_arrays_primes_large.a" || filepath.Base(path) == "normalize_coverage_long.a"
	})
	if len(paths) < 60 {
		t.Fatalf("found only %d programs: the globs no longer find the fixtures", len(paths))
	}
	return paths
}

// Package tests use immutable, hash-pinned source inputs. Each program is lowered
// once per package process; graph construction never changes this shared IR.
type loweredProgram struct {
	once    sync.Once
	program *ir.Program
	err     error
	phase   string
}

var loweredPrograms sync.Map

// Only deliberate .a refusal controls are excluded. Runtime assertions use .ts.
func refusedNonNullFixture(path string) bool {
	name := filepath.Base(path)
	if filepath.Ext(path) != ".a" || !(strings.HasPrefix(name, "non_null_refuse_") || strings.HasPrefix(name, "non_null_possible_")) {
		return false
	}
	return true
}

func lowered(t *testing.T, path string) *ir.Program {
	t.Helper()
	absolute, err := filepath.Abs(path)
	if err != nil {
		t.Fatal(err)
	}
	value, _ := loweredPrograms.LoadOrStore(absolute, &loweredProgram{})
	entry := value.(*loweredProgram)
	entry.once.Do(func() {
		entry.phase = "Load"
		program, err := load.Load([]string{absolute})
		entry.err = err
		if err == nil {
			entry.phase = "Lower"
			entry.program, entry.err = Lower(context.Background(), program)
		}
	})
	if entry.err != nil {
		t.Fatalf("%s: %s: %v", path, entry.phase, entry.err)
	}
	return entry.program
}

// Every function of every program goes into single assignment form, and three checks hold it there,
// none of them built from the construction: the verifier's (each value defined once, each use
// dominated by its definition), reaching definitions computed the classic way (each use names
// exactly what reaches it), and a count of the IR's reads made by encoding/json rather than by
// Build's walk (no read went missing).
func checkSingleAssignmentProgram(t *testing.T, path string) {
	t.Helper()
	var totals SSAStats
	var functions int
	program := lowered(t, path)
	for function := -1; function < len(program.Functions); function++ {
		name := "main"
		if function >= 0 {
			name = program.Functions[function].Name
		}
		where := fmt.Sprintf("%s, function %d (%s)", path, function, name)
		checkReads(t, where, program, function)
		reaching := reachingDefinitions(Build(program, function))
		graph := Build(program, function)
		Construct(graph)
		for _, violation := range VerifySSA(graph) {
			t.Errorf("%s: %s", where, violation)
		}
		checkReaching(t, where, graph, reaching)
		stats := CollectSSAStats(graph)
		totals.Phis += stats.Phis
		totals.NamedValues += stats.NamedValues
		totals.Uses += stats.Uses
		functions++
	}
	// An empty list of violations is what a vacuous check returns too.
	if path == "testdata/joins.a" && (totals.Phis == 0 || totals.Uses == 0) {
		t.Errorf("nothing was checked: %+v", totals)
	}
	t.Logf("%d functions: %d phis, %d values, %d uses checked", functions, totals.Phis, totals.NamedValues, totals.Uses)
}

// checkReads counts the tracked reads in an IR function from fmt's %#v of it, where each read is
// printed as ir.Read{Local:N, ...}, and holds Build's uses to that count. fmt walks the IR with its
// own code, not Build's, so a read Build's walk missed shows as a difference.
func checkReads(t *testing.T, where string, program *ir.Program, function int) {
	t.Helper()
	statements := program.Main
	if function >= 0 {
		statements = program.Functions[function].Body
	}
	want := 0
	for _, match := range readPrinted.FindAllStringSubmatch(fmt.Sprintf("%#v", statements), -1) {
		local, err := strconv.Atoi(match[1])
		if err != nil {
			t.Fatal(err)
		}
		if !program.Locals[local].Global && !program.Locals[local].Captured && !program.Locals[local].ExpressionAssigned {
			want++
		}
	}
	graph := Build(program, function)
	got := 0
	for _, instruction := range graph.Instructions {
		got += len(instruction.Uses)
	}
	// The graph keeps only what's reachable, so a read in code after a return would be in the IR and
	// not the graph. The programs here have no such code, so the counts must be equal.
	if got != want {
		t.Errorf("%s: the graph reads tracked locals %d times, the IR %d", where, got, want)
	}
}

var readPrinted = regexp.MustCompile(`ir\.Read\{Local:(\d+),`)

// site is where a variable gets a value: a parameter (instruction -1, its index) or an instruction's
// definition (its id, its index among the instruction's definitions).
type site struct {
	instruction int
	index       int
}

// reaching is, for each use (instruction id, index among its uses), the sites whose value can reach
// it, computed by the textbook iterative dataflow over the graph before construction.
type reaching map[site][]site

func reachingDefinitions(graph *Function) reaching {
	// What each block holds on exit, per variable: the sites that reach there.
	type state map[DeclarationId][]site
	declaration := func(place Place) DeclarationId { return graph.Identifiers[place.Identifier].Declaration }
	merge := func(into state, from state) bool {
		changed := false
		for variable, sites := range from {
			for _, each := range sites {
				if !slices.Contains(into[variable], each) {
					into[variable] = append(into[variable], each)
					changed = true
				}
			}
		}
		return changed
	}
	entryState := state{}
	for index, parameter := range graph.Params {
		entryState[declaration(parameter)] = []site{{instruction: -1, index: index}}
	}
	exits := map[BlockId]state{}
	// thrown is what a block ending in MayThrow hands its handler: what reached its last instruction,
	// since a throw out of it never makes its definitions.
	thrown := map[BlockId]state{}
	result := reaching{}
	for changed := true; changed; {
		changed = false
		for _, block := range graph.Blocks {
			current := state{}
			if block.Id == graph.Entry {
				merge(current, entryState)
			}
			for _, predecessor := range block.Predecessors {
				from, _ := graph.Block(predecessor)
				if throws, ok := from.Terminal.(*MayThrow); ok && throws.Handler == block.Id {
					merge(current, thrown[predecessor])
					continue
				}
				merge(current, exits[predecessor])
			}
			_, endsInThrow := block.Terminal.(*MayThrow)
			for position, id := range block.Instructions {
				instruction := graph.Instructions[id]
				for index, use := range instruction.Uses {
					result[site{int(id), index}] = slices.Clone(current[declaration(use)])
				}
				if endsInThrow && position == len(block.Instructions)-1 {
					if thrown[block.Id] == nil {
						thrown[block.Id] = state{}
					}
					if merge(thrown[block.Id], current) {
						changed = true
					}
				}
				for index, define := range instruction.Defines {
					current[declaration(define)] = []site{{int(id), index}}
				}
			}
			// A block's exit is monotone in its entry, and entries only grow, so merging the new exit
			// into the old one is the new exit, and the loop reaches a fixed point.
			if exits[block.Id] == nil {
				exits[block.Id] = state{}
			}
			if merge(exits[block.Id], current) {
				changed = true
			}
		}
	}
	return result
}

// checkReaching holds a constructed graph to reaching definitions: a use that one site reaches names
// that site's value, and a use several reach names a phi whose operands, followed through any phis
// they name, come to exactly those sites.
func checkReaching(t *testing.T, where string, graph *Function, want reaching) {
	t.Helper()
	sitesOf := definitionSites(graph)
	order := func(sites []site) string {
		text := make([]string, len(sites))
		for index, each := range sites {
			text[index] = fmt.Sprint(each)
		}
		slices.Sort(text)
		return strings.Join(text, " ")
	}
	for _, block := range graph.Blocks {
		for _, id := range block.Instructions {
			for index, use := range graph.Instructions[id].Uses {
				expected := want[site{int(id), index}]
				if len(expected) == 0 {
					t.Errorf("%s: %s in instruction %d reads a variable nothing defines", where, graph.PlaceString(use), id)
					continue
				}
				isPhi := sitesOf.isPhi(use.Identifier)
				got := sitesOf.of(use.Identifier)
				switch {
				case len(expected) == 1 && (isPhi || order(got) != order(expected)):
					t.Errorf("%s: %s in instruction %d names %v, want the one definition that reaches it, %v", where, graph.PlaceString(use), id, got, expected)
				case len(expected) > 1 && (!isPhi || order(got) != order(expected)):
					t.Errorf("%s: %s in instruction %d names %v (a phi: %t), want a phi over what reaches it, %v", where, graph.PlaceString(use), id, got, isPhi, expected)
				}
			}
		}
	}
}

// definitions says, of a constructed graph's values, which sites each comes to: a parameter's or an
// instruction's value is its own site, and a phi's is every site its operands come to, followed
// through any phis they name.
type definitions struct {
	definedBy map[IdentifierId]site
	phis      map[IdentifierId]*Phi
}

func definitionSites(graph *Function) definitions {
	definedBy := map[IdentifierId]site{}
	for index, parameter := range graph.Params {
		definedBy[parameter.Identifier] = site{-1, index}
	}
	phis := map[IdentifierId]*Phi{}
	for _, block := range graph.Blocks {
		for _, phi := range block.Phis {
			phis[phi.Place.Identifier] = phi
		}
		for _, id := range block.Instructions {
			for index, define := range graph.Instructions[id].Defines {
				definedBy[define.Identifier] = site{int(id), index}
			}
		}
	}
	return definitions{definedBy: definedBy, phis: phis}
}

func (s definitions) isPhi(value IdentifierId) bool {
	_, ok := s.phis[value]
	return ok
}

// of is every site a value comes to.
func (s definitions) of(value IdentifierId) []site {
	return s.walk(value, map[IdentifierId]bool{})
}

func (s definitions) walk(value IdentifierId, seen map[IdentifierId]bool) []site {
	if each, ok := s.definedBy[value]; ok {
		return []site{each}
	}
	phi, ok := s.phis[value]
	if !ok || seen[value] {
		return nil
	}
	seen[value] = true
	var found []site
	for _, predecessor := range PhiOperandsInOrder(phi) {
		for _, each := range s.walk(phi.Operands[predecessor].Identifier, seen) {
			if !slices.Contains(found, each) {
				found = append(found, each)
			}
		}
	}
	return found
}

// These constructor fixtures deliberately assert NotYet and never enter the oracle's runnable set.
func refusedConstructorFixture(path string) bool {
	switch filepath.Base(path) {
	case "new_expression_class_cache_capture_notyet.a":
		return true
	}
	return false
}

// These accessor census witnesses pin where lowering stops (internal/oracle/getters_census_test.go
// asserts each one's refusal), so they never lower and have no graph. The census's three that lower,
// class_map, throwing_object and binary_inline, stay in the corpus.
func refusedGettersCensusFixture(path string) bool {
	switch filepath.Base(path) {
	case "getters_census_binary_function.a", "getters_census_callback_pair_methods.a",
		"getters_census_closure_number_methods.a", "getters_census_closure_number_next.a",
		"getters_census_closure_union_next.a", "getters_census_higher_callback_pair_methods.a",
		"getters_census_lazy_object.a", "getters_census_optional_boolean_function.a",
		"getters_census_optional_comment_function.a", "getters_census_optional_type_function.a",
		"getters_census_overloaded_function.a", "getters_census_primary_function.a",
		"getters_census_snapshot_next.a", "getters_census_unary_function.a",
		"getters_census_unary_node_function.a", "getters_census_update_comment_function.a",
		"getters_census_update_node_function.a", "getters_census_update_type_function.a":
		return true
	}
	return false
}

// Async unit 1's JavaScript backend has no flow instrumentation; unit 2 adds it.
// Name only its current witnesses so later async fixtures still enter the corpus.
func uninstrumentedAsyncFixture(path string) bool {
	switch filepath.Base(path) {
	case "async_coverage_discard.a", "async_coverage_parameters.a",
		"async_coverage_reject_eager.a", "async_coverage_reject_empty.a",
		"async_coverage_reject_nested.a", "async_coverage_typeof.a",
		"async_coverage_unions.a", "async_coverage_values.a",
		"async_nested.a", "async_plain.a", "async_three.a", "async_throw.a", "async_typeof.a":
		return true
	}
	return false
}

// These boundary witnesses pin lowering refusals in the oracle and have no flow graph.
func runtimeBoundaryLoweringReason(path string) string {
	switch filepath.Base(path) {
	case "coverage_regexp_call.a", "coverage_regexp_literal.a", "coverage_regexp_new.a":
		return "stage 0 can't lower spreading a RegExp or its iterator yet"
	case "coverage_view_array.a", "coverage_view_boolean.a", "coverage_view_closure.a", "coverage_view_number.a", "coverage_view_string.a":
		return "Adamic 0.1 refuses an object spread through a view that may hold library class Array (its internal fields are not own enumerable properties); copy the fields you mean by name"
	}
	return ""
}
