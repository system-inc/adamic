package lower

import (
	"fmt"
	"hash/fnv"
	"os"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/load"
)

// relationSizeTrace is a probe, never for main (#nxb3xh1): each relation cache's size, read by name as
// load.RelationOverflowMarked reads the caches, printed on stderr as one RELSIZE line.
func relationSizeTrace(typeChecker *checker.Checker, label string) {
	value := reflect.ValueOf(typeChecker).Elem()
	var parts []string
	total := 0
	for _, name := range []string{"subtypeRelation", "strictSubtypeRelation", "assignableRelation", "comparableRelation", "identityRelation"} {
		relation := value.FieldByName(name)
		size := -1
		if relation.IsValid() && relation.Kind() == reflect.Pointer && !relation.IsNil() {
			if results := relation.Elem().FieldByName("results"); results.IsValid() && results.Kind() == reflect.Map {
				size = results.Len()
				total += size
			}
		}
		parts = append(parts, fmt.Sprintf("%s=%d", strings.TrimSuffix(name, "Relation"), size))
	}
	if enum := value.FieldByName("enumRelation"); enum.IsValid() && enum.Kind() == reflect.Map {
		parts = append(parts, fmt.Sprintf("enum=%d", enum.Len()))
	}
	fmt.Fprintf(os.Stderr, "RELSIZE %s %s total=%d %s\n", time.Now().UTC().Format("15:04:05"), label, total, strings.Join(parts, " "))
}

// armPairs and armCalls are the probe's count, per arm of graphLinks, of assignable pairs added and calls made,
// printed and reset at each findCycles end (#nxb3xh1, Compiler's (b)).
var (
	graphLinksCalls int
	armPairs = map[string]int{}
	armCalls = map[string]int{}
)

func assignableSize(typeChecker *checker.Checker) int {
	relation := reflect.ValueOf(typeChecker).Elem().FieldByName("assignableRelation")
	if !relation.IsValid() || relation.Kind() != reflect.Pointer || relation.IsNil() {
		return -1
	}
	results := relation.Elem().FieldByName("results")
	if !results.IsValid() || results.Kind() != reflect.Map {
		return -1
	}
	return results.Len()
}

func armTrace(label string) {
	keys := make([]string, 0, len(armPairs))
	for key := range armCalls {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		fmt.Fprintf(os.Stderr, "ARM %s %s pairs=%d calls=%d\n", label, key, armPairs[key], armCalls[key])
	}
	armPairs = map[string]int{}
	armCalls = map[string]int{}
}

// graphLinksProgress counts graphLinks' calls and prints the relation sizes every 5000.
func graphLinksProgress(typeChecker *checker.Checker) {
	graphLinksCalls++
	if graphLinksCalls%5000 == 0 {
		relationSizeTrace(typeChecker, fmt.Sprintf("graphLinks call %d", graphLinksCalls))
	}
}

// edgeLoopTrace prints graphTypes' edge loop's progress: candidates, shapes, where, nodes linked and related calls.
func edgeLoopTrace(typeChecker *checker.Checker, label string, candidates, shapes, where, nodes, related int) {
	fmt.Fprintf(os.Stderr, "EDGES %s %s candidates=%d shapes=%d where=%d nodes=%d related=%d assignable=%d\n", time.Now().UTC().Format("15:04:05"), label, candidates, shapes, where, nodes, related, assignableSize(typeChecker))
}

// probeNames is a type's property names and its required ones, the probe's cut B (#nxb3xh1): a required property of
// the target the source lacks makes the source not assignable, so a pair failing that is never asked of the checker.
type probeNames struct {
	all      map[string]bool
	required []string
}

var probeNamesOf = map[*checker.Type]*probeNames{}

func probeNamesFor(f *cycleFinder, proven *checker.Type) *probeNames {
	if names, found := probeNamesOf[proven]; found {
		return names
	}
	names := &probeNames{all: map[string]bool{}}
	for _, property := range f.l.checker.GetPropertiesOfType(proven) {
		names.all[property.Name] = true
		if property.Flags&ast.SymbolFlagsOptional == 0 {
			names.required = append(names.required, property.Name)
		}
	}
	probeNamesOf[proven] = names
	return names
}

func probeMayAssign(f *cycleFinder, source, target *checker.Type) bool {
	if source.Flags()&checker.TypeFlagsObject == 0 || target.Flags()&checker.TypeFlagsObject == 0 {
		return true
	}
	held := probeNamesFor(f, source)
	for _, name := range probeNamesFor(f, target).required {
		if !held.all[name] {
			return false
		}
	}
	return true
}

func probeMayRelate(f *cycleFinder, one, other *checker.Type) bool {
	if one == other {
		return true
	}
	one, other = f.l.relatedShape(one), f.l.relatedShape(other)
	return probeMayAssign(f, one, other) || probeMayAssign(f, other, one)
}

// probeCutCheck turns on allocationCutCheck in the probe that asks it (#nxb3xh1 head 6, Compiler's condition 3).
var probeCutCheck = true

// allocationCutCheck asks the checker, by deriveRelated, of the pairs graphTypes' sweep cut over the seeds' closure:
// a fixed seeded 1% of them by their type ids, and every one involving a type with an index signature or one the
// library, the prelude or Node's types declare, up to cutCheckCap checks. Each disagreement prints as a CUTCHECK
// line naming both types, and the counts as one more.
func allocationCutCheck(f *cycleFinder, edges map[cycleNode][]cycleNode) {
	const cutCheckCap = 2000000
	start := time.Now()
	views := f.allocation.views
	position := map[cycleNode]int{}
	for at, view := range views {
		position[view] = at
	}
	marks := make([]int, len(views))
	special := map[*checker.Type]bool{}
	isSpecial := func(proven *checker.Type) bool {
		if value, known := special[proven]; known {
			return value
		}
		value := len(f.l.checker.GetIndexInfosOfType(proven)) != 0
		symbol := proven.Symbol()
		if symbol == nil && proven.ObjectFlags()&checker.ObjectFlagsReference != 0 && proven.Target() != nil {
			symbol = proven.Target().Symbol()
		}
		if symbol != nil {
			for _, declaration := range symbol.Declarations {
				if file := ast.GetSourceFileOfNode(declaration); file != nil && (load.IsLibrary(file) || load.IsPrelude(file) || load.IsNodeLibrary(file)) {
					value = true
				}
			}
		}
		special[proven] = value
		return value
	}
	nodes, cut, sampled, specials, checked, disagreements, capped := 0, 0, 0, 0, 0, 0, 0
	query := 0
	for node := range edges {
		if node.proven == nil || node.proven.Flags()&checker.TypeFlagsObject == 0 || f.weak(node.proven) || f.template(node.proven) || f.isFunction(node.proven) {
			continue
		}
		nodes++
		query++
		for _, view := range f.allocationViews(node.proven) {
			if at, known := position[view]; known {
				marks[at] = query
			}
		}
		for at, view := range views {
			if marks[at] == query || view.proven == node.proven {
				continue
			}
			cut++
			hash := fnv.New32a()
			fmt.Fprintf(hash, "nxb6:%d:%d", node.proven.Id(), view.proven.Id())
			inSample := hash.Sum32()%100 == 0
			isOne := isSpecial(node.proven) || isSpecial(view.proven)
			if inSample {
				sampled++
			}
			if isOne {
				specials++
			}
			if !inSample && !isOne {
				continue
			}
			if checked >= cutCheckCap {
				capped++
				continue
			}
			checked++
			if f.deriveRelated(node.proven, view.proven) {
				disagreements++
				if disagreements <= 100 {
					fmt.Fprintf(os.Stderr, "CUTCHECK disagree %s | %s\n", f.l.checker.TypeToString(node.proven), f.l.checker.TypeToString(view.proven))
				}
			}
		}
	}
	fmt.Fprintf(os.Stderr, "CUTCHECK nodes=%d views=%d cut=%d sampled=%d special=%d checked=%d capped=%d disagreements=%d wall=%s assignable=%d\n", nodes, len(views), cut, sampled, specials, checked, capped, disagreements, time.Since(start).Round(time.Second), assignableSize(f.l.checker))
}
