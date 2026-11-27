package lower

import (
	"sort"
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// The finder's own relation for selection (#nxb3xh1 head 6, Compiler's ruling with Memory's conditions).
// graphTypes' view sweep and graphLinks' shapes road asked the checker whether every object node relates to
// every candidate: on tsc's first link 57,846 candidates, every answer a pair in the checker's relation cache,
// whose budget was spent within 9,000 nodes, after which every query reported and the run never ended. Those two
// sweeps only ask whether two values could be one allocation, to select graph members, and there relating more
// than the checker would is sound: an extra member costs its header, a region record and merge searches, never a
// wrong free (Memory's reading of graph_regions.c), and a cycle is found more often, never less. So they ask
// mayShareAllocation, which never asks the checker's relation, and answers "not related" only where the checker
// certainly would (Compiler's condition 1).
//
// What the checker certainly refuses: a target's required property the source lacks. The relation finds it by
// getPropertyOfType on the source (relater.go's getUnmatchedPropertiesWorker), which falls back to the global
// Object type's members for every object type, so a required toString or hasOwnProperty is held by every object,
// and those names are dropped from every required set here (allocationUniversal). An optional property never
// counts as missing. Two types are related when either one's required names are all held by the other.
//
// That test reads only what a type lists, so it's asked only of plain types (allocationPlain): an object type
// with no call or construct signature, which isn't mapped, reverse-mapped or generic (the checker's
// isGenericObjectType) and isn't a template. A library type, a reference with concrete type arguments and a type
// with an index signature are plain (Compiler's ruling at 14:16 on #nxb3xh1, replacing her first list): a
// reference lists its target's members, since instantiation maps their types and never their names, and an index
// signature on either side adds a requirement but never stands in for a missing required property. Anything else
// is related to everything, failing closed: unions and intersections, which are never judged as one object, and
// any, unknown, type parameters, primitives and literals, which the checker reads through their apparent types,
// none of them object types; and a callable or constructable type, whose missing names the checker also finds on
// Function. Were one of these a candidate, every node would link to it and it to every node, so a run that meets
// one selects every object type the seeds reach: a cost, never a missed cycle.
//
// Everything else in the finder keeps the checker's exact relation (related, deriveRelated): viewAlsoTable skips an
// edge for a pair related exactly, on the premise the finder walks that pair as one allocation, and reachLinks'
// shapes road walks by IsTypeAssignableTo, so widening its relation would leave a skipped pair walked nowhere
// (Memory's condition A, pinned by the viewAlsoTable mutant in shared_allocation.mutants.md); viewPairFailure's
// tests stay exact for the same premise (condition B); unproven writes and reaching stay exact too.

// mayShareAllocation reports whether one value could be seen as both types, for selection only: true unless both
// are plain and neither's required names are all held by the other.
func (f *cycleFinder) mayShareAllocation(one, other *checker.Type) bool {
	if one == other {
		return true
	}
	oneShape, otherShape := f.allocationShapeOf(one), f.allocationShapeOf(other)
	if !oneShape.plain || !otherShape.plain {
		return true
	}
	return allocationHolds(oneShape, otherShape) || allocationHolds(otherShape, oneShape)
}

// mayBeSeenAs is mayShareAllocation's one-way form: whether a value of source could be seen as target, true unless
// both are plain and source lacks one of target's required names.
func (f *cycleFinder) mayBeSeenAs(source, target *checker.Type) bool {
	if source == target {
		return true
	}
	sourceShape, targetShape := f.allocationShapeOf(source), f.allocationShapeOf(target)
	if !sourceShape.plain || !targetShape.plain {
		return true
	}
	return allocationHolds(sourceShape, targetShape)
}

// allocationShape is what the relation reads of a type, by its relatedShape (a fresh literal by its regular widened
// type, as deriveRelated reads one): whether it's plain, and for a plain type the names it holds, its required
// names without Object's members, sorted, and a key naming both, so types with the same names share one group.
type allocationShape struct {
	plain    bool
	held     map[string]bool
	required []string
	key      string
}

// allocationHolds is whether source holds every required name of target.
func allocationHolds(source, target *allocationShape) bool {
	for _, name := range target.required {
		if !source.held[name] {
			return false
		}
	}
	return true
}

// allocationShapeOf reads a type's allocationShape once per finder.
func (f *cycleFinder) allocationShapeOf(proven *checker.Type) *allocationShape {
	if shape := f.allocationShapes[proven]; shape != nil {
		return shape
	}
	if f.allocationShapes == nil {
		f.allocationShapes = map[*checker.Type]*allocationShape{}
	}
	shape := &allocationShape{}
	read := f.l.relatedShape(proven)
	if f.allocationPlain(read) {
		shape.plain = true
		shape.held = map[string]bool{}
		universal := f.allocationUniversal()
		held := []string{}
		for _, property := range f.l.checker.GetPropertiesOfType(read) {
			shape.held[property.Name] = true
			held = append(held, property.Name)
			if property.Flags&ast.SymbolFlagsOptional == 0 && !universal[property.Name] {
				shape.required = append(shape.required, property.Name)
			}
		}
		sort.Strings(held)
		sort.Strings(shape.required)
		quote := func(names []string) string {
			encoded := make([]string, len(names))
			for index, name := range names {
				encoded[index] = strconv.Quote(name)
			}
			return strings.Join(encoded, ",")
		}
		shape.key = quote(held) + "|" + quote(shape.required)
	}
	f.allocationShapes[proven] = shape
	return shape
}

// allocationUniversal is the global Object type's member names, which getPropertyOfType finds on every object type
// that doesn't list them itself, so no object type lacks one. With no Object in the program's library the checker's
// global Object type is empty, and so is this.
func (f *cycleFinder) allocationUniversal() map[string]bool {
	if f.allocationUniversalNames != nil {
		return f.allocationUniversalNames
	}
	f.allocationUniversalNames = map[string]bool{}
	if symbol := f.l.checker.GetGlobalSymbol("Object", ast.SymbolFlagsType, nil); symbol != nil {
		for _, property := range f.l.checker.GetPropertiesOfType(f.l.checker.GetDeclaredTypeOfSymbol(symbol)) {
			f.allocationUniversalNames[property.Name] = true
		}
	}
	return f.allocationUniversalNames
}

// allocationPlain is whether a type's listed properties are all the relation needs to read (the comment above).
func (f *cycleFinder) allocationPlain(proven *checker.Type) bool {
	if proven.Flags()&checker.TypeFlagsObject == 0 || f.template(proven) {
		return false
	}
	if proven.ObjectFlags()&(checker.ObjectFlagsMapped|checker.ObjectFlagsReverseMapped) != 0 {
		return false
	}
	// isGenericObjectType, for an object type that isn't mapped: a tuple with a variadic element.
	if checker.IsTupleType(proven) {
		for _, element := range proven.TargetTupleType().ElementFlags() {
			if element&checker.ElementFlagsVariadic != 0 {
				return false
			}
		}
	}
	return len(f.l.checker.GetSignaturesOfType(proven, checker.SignatureKindCall)) == 0 && len(f.l.checker.GetSignaturesOfType(proven, checker.SignatureKindConstruct)) == 0
}

// allocationIndex is one findCycles run's index for the two sweeps, so a node asks only the candidates whose names
// it can satisfy, or that can satisfy its own, never a scan of every candidate (Compiler's condition 2). Plain
// candidates are grouped by their names (allocationShape.key), each group indexed by the names it requires and the
// names it holds; a candidate that isn't plain is related to every node, so it's kept in a list of its own.
type allocationIndex struct {
	// views are the sweep's candidates, every one a node that isn't plain links to, and shapes the shapes road's.
	views  []cycleNode
	shapes []*checker.Type
	// shapeCount is len(f.shapes) when the index was built, so the shapes road rebuilds it if shapes grows.
	shapeCount int
	groups     []*allocationGroup
	groupOf    map[string]int
	// requiring is, by name, the groups that require it; holding the groups that hold it; unrequired the groups
	// that require nothing, which every plain node's names satisfy.
	requiring  map[string][]int
	holding    map[string][]int
	unrequired []int
	// looseViews and looseShapes are the candidates and shapes that aren't plain.
	looseViews  []cycleNode
	looseShapes []*checker.Type
	// related caches, by a plain node's key, the groups mayShareAllocation relates it to, and seen the groups a
	// plain target's required names are all held by, for the shapes road.
	related map[string][]int
	seen    map[string][]int
	// counts and marks are one query's scratch, by group: how many of a group's required names the node holds, and
	// the query a group was last added in.
	counts []int
	marks  []int
	query  int
}

// allocationGroup is the plain candidates and shapes sharing one set of names.
type allocationGroup struct {
	shape  *allocationShape
	views  []cycleNode
	shapes []*checker.Type
}

// allocationView is whether the sweep may link a candidate at all: an object type that isn't weak, a template or a
// function's. deriveRelated answers false for any pair with a function's type in it, and so the sweep linked none,
// so a function's type stays out here too and the sweep relates a superset of the pairs it related before.
func (f *cycleFinder) allocationView(proven *checker.Type) bool {
	return proven != nil && proven.Flags()&checker.TypeFlagsObject != 0 && !f.weak(proven) && !f.template(proven) && !f.isFunction(proven)
}

// allocationIndexOf builds the index for graphTypes' candidates and the finder's shapes.
func (f *cycleFinder) allocationIndexOf(candidates []cycleNode) *allocationIndex {
	index := &allocationIndex{shapeCount: len(f.shapes), groupOf: map[string]int{}, requiring: map[string][]int{}, holding: map[string][]int{}, related: map[string][]int{}, seen: map[string][]int{}}
	group := func(shape *allocationShape) *allocationGroup {
		if at, known := index.groupOf[shape.key]; known {
			return index.groups[at]
		}
		at := len(index.groups)
		index.groupOf[shape.key] = at
		index.groups = append(index.groups, &allocationGroup{shape: shape})
		for _, name := range shape.required {
			index.requiring[name] = append(index.requiring[name], at)
		}
		if len(shape.required) == 0 {
			index.unrequired = append(index.unrequired, at)
		}
		for name := range shape.held {
			index.holding[name] = append(index.holding[name], at)
		}
		return index.groups[at]
	}
	// A type is a candidate once for where and once more for shapes when use takes it in as both, so each is
	// indexed once, and a node's links name each view once.
	indexed := map[cycleNode]bool{}
	for _, candidate := range candidates {
		if !f.allocationView(candidate.proven) || indexed[candidate] {
			continue
		}
		indexed[candidate] = true
		index.views = append(index.views, candidate)
		if shape := f.allocationShapeOf(candidate.proven); shape.plain {
			group(shape).views = append(group(shape).views, candidate)
		} else {
			index.looseViews = append(index.looseViews, candidate)
		}
	}
	shaped := map[*checker.Type]bool{}
	for _, shape := range f.shapes {
		if shaped[shape] {
			continue
		}
		shaped[shape] = true
		index.shapes = append(index.shapes, shape)
		if read := f.allocationShapeOf(shape); read.plain {
			group(read).shapes = append(group(read).shapes, shape)
		} else {
			index.looseShapes = append(index.looseShapes, shape)
		}
	}
	index.counts = make([]int, len(index.groups))
	index.marks = make([]int, len(index.groups))
	return index
}

// allocationViews is every sweep candidate mayShareAllocation relates to proven: all of them when proven isn't
// plain; else the groups whose required names proven holds, the groups holding proven's required names, and every
// candidate that isn't plain.
func (f *cycleFinder) allocationViews(proven *checker.Type) []cycleNode {
	index := f.allocation
	shape := f.allocationShapeOf(proven)
	if !shape.plain {
		return index.views
	}
	groups, known := index.related[shape.key]
	if !known {
		groups = index.holdersOf(f, shape, index.requiredBy(f, shape, nil))
		index.related[shape.key] = groups
	}
	views := append([]cycleNode{}, index.looseViews...)
	for _, at := range groups {
		views = append(views, index.groups[at].views...)
	}
	return views
}

// allocationShapesSeenAs is every shape but proven itself that mayBeSeenAs proven, for graphLinks' shapes road: all
// of them when proven isn't plain; else the shapes whose names hold proven's required names, and every shape that
// isn't plain.
func (f *cycleFinder) allocationShapesSeenAs(proven *checker.Type) []*checker.Type {
	if f.allocation == nil || f.allocation.shapeCount != len(f.shapes) {
		var views []cycleNode
		if f.allocation != nil {
			views = f.allocation.views
		}
		f.allocation = f.allocationIndexOf(views)
	}
	index := f.allocation
	shapes := []*checker.Type{}
	add := func(shape *checker.Type) {
		if shape != proven {
			shapes = append(shapes, shape)
		}
	}
	shape := f.allocationShapeOf(proven)
	if !shape.plain {
		for _, each := range index.shapes {
			add(each)
		}
		return shapes
	}
	groups, known := index.seen[shape.key]
	if !known {
		groups = index.holdersOf(f, shape, nil)
		index.seen[shape.key] = groups
	}
	for _, each := range index.looseShapes {
		add(each)
	}
	for _, at := range groups {
		for _, each := range index.groups[at].shapes {
			add(each)
		}
	}
	return shapes
}

// requiredBy adds to groups, once each, every group whose required names shape holds, counting, for each name shape
// holds, the groups requiring it: a group is reached when the count meets its required names.
func (index *allocationIndex) requiredBy(f *cycleFinder, shape *allocationShape, groups []int) []int {
	index.query++
	touched := []int{}
	for name := range shape.held {
		for _, at := range index.requiring[name] {
			f.allocationTests++
			if index.counts[at] == 0 {
				touched = append(touched, at)
			}
			index.counts[at]++
			if index.counts[at] == len(index.groups[at].shape.required) && index.marks[at] != index.query {
				index.marks[at] = index.query
				groups = append(groups, at)
			}
		}
	}
	for _, at := range touched {
		index.counts[at] = 0
	}
	for _, at := range index.unrequired {
		if index.marks[at] != index.query {
			index.marks[at] = index.query
			groups = append(groups, at)
		}
	}
	return groups
}

// holdersOf adds to groups every group holding all of shape's required names, not already among them, read from
// the fewest groups holding one of those names.
func (index *allocationIndex) holdersOf(f *cycleFinder, shape *allocationShape, groups []int) []int {
	index.query++
	for _, at := range groups {
		index.marks[at] = index.query
	}
	add := func(at int) {
		if index.marks[at] != index.query {
			index.marks[at] = index.query
			groups = append(groups, at)
		}
	}
	if len(shape.required) == 0 {
		for at := range index.groups {
			add(at)
		}
		return groups
	}
	fewest := index.holding[shape.required[0]]
	for _, name := range shape.required[1:] {
		if len(index.holding[name]) < len(fewest) {
			fewest = index.holding[name]
		}
	}
	for _, at := range fewest {
		f.allocationTests++
		if allocationHolds(index.groups[at].shape, shape) {
			add(at)
		}
	}
	return groups
}
