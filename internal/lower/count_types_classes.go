package lower

import (
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"sync/atomic"
	_ "unsafe"

	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// resolveCountTypes' FunctionTypeTargets asked the checker whether every closure record's type is assignable to
// every count target: 9.3M questions on a 23K-line program, 64% of the process's CPU, where most targets are the
// records' own function types and the rows come to a handful. Types the checker holds identical get one answer from
// assignability, so the targets are put in identity classes and so are the records, the checker is asked once per
// pair of classes, and each answer is spread to every record of its class, in record order, for every target of its
// class (#mh5kdde). Identity doesn't see a fresh object literal's freshness, which assignability's excess-property
// check reads, so a fresh type is a class of its own on both sides (countTypeFresh).

// countTypeClassesTally counts, while a test sets it, what countTypeTargets did: the checker's assignability answers
// it asked for, the class pairs they should be, and the records and targets with the classes they made, so a pin holds
// the asks to one per pair of classes. It's nil outside those tests (#mh5kdde).
var countTypeClassesTally atomic.Pointer[countTypeClassesCounts]

// countTypeClassesCounts is the tally's counts, summed over every countTypeTargets the lowerings made while it was set
// (#mh5kdde).
type countTypeClassesCounts struct {
	runs, asks, pairs, records, sourceClasses, targets, targetClasses atomic.Int64
}

// resolveCountTypesCompare, set only by tests, asks the checker for the whole matrix of targets and records again
// beside the shared answers, and panics where a target's row or a pair's relation overflow differs from its classes'
// (countTypeTargetsCompared). Tests that set it aren't parallel, and no parallel test runs beside a sequential one
// (#mh5kdde).
var resolveCountTypesCompare atomic.Bool

// countTypeTargets fills FunctionTypeTargets for types, a target's row being every closure record whose type is
// assignable to it, in record order, as asking each record of each target would (#mh5kdde).
func (l *lowering) countTypeTargets(types map[int]*checker.Type) {
	l.result.FunctionTypeTargets = map[int][]int{}
	ids := make([]int, 0, len(types))
	targets := make([]*checker.Type, 0, len(types))
	for id, proven := range types {
		l.cycleTypeID(proven) // Retain transient count-target types for differential tests.
		ids = append(ids, id)
		targets = append(targets, proven)
	}
	sources := make([]*checker.Type, len(l.closureRecords))
	for index, record := range l.closureRecords {
		sources[index] = record.proven
	}
	targetClassOf, targetRepresentatives := countTypeClasses(l.checker, targets)
	sourceClassOf, sourceRepresentatives := countTypeClasses(l.checker, sources)
	tally := countTypeClassesTally.Load()
	compare := resolveCountTypesCompare.Load()
	// overflows is each class pair's relation overflow, kept only for the comparison (#mh5kdde).
	var overflows [][]bool
	if compare {
		overflows = make([][]bool, len(targetRepresentatives))
	}
	rows := make([][]int, len(targetRepresentatives))
	for targetClass, target := range targetRepresentatives {
		assignable := make([]bool, len(sourceRepresentatives))
		if compare {
			overflows[targetClass] = make([]bool, len(sourceRepresentatives))
		}
		for sourceClass, source := range sourceRepresentatives {
			assignable[sourceClass] = l.checker.IsTypeAssignableTo(source, target)
			if tally != nil {
				tally.asks.Add(1)
			}
			if compare {
				overflows[targetClass][sourceClass] = countTypeOverflowed(l.checker, source, target)
			}
		}
		rows[targetClass] = []int{}
		for index, record := range l.closureRecords {
			if assignable[sourceClassOf[index]] {
				rows[targetClass] = append(rows[targetClass], record.function)
			}
		}
	}
	// Each target holds a row of its own, as each did when it built its own, so nothing written through one target's
	// row reaches another's (#mh5kdde).
	for index, id := range ids {
		l.result.FunctionTypeTargets[id] = slices.Clone(rows[targetClassOf[index]])
	}
	if tally != nil {
		tally.runs.Add(1)
		tally.pairs.Add(int64(len(sourceRepresentatives) * len(targetRepresentatives)))
		tally.records.Add(int64(len(sources)))
		tally.sourceClasses.Add(int64(len(sourceRepresentatives)))
		tally.targets.Add(int64(len(targets)))
		tally.targetClasses.Add(int64(len(targetRepresentatives)))
	}
	if compare {
		l.countTypeTargetsCompared(ids, targets, targetClassOf, sourceClassOf, overflows)
	}
}

// countTypeTargetsCompared asks the checker every record of every target, as resolveCountTypes did before the
// classes, and panics where a target's row isn't the one its class gave it, or where a record's relation to a target
// overflowed and its classes' didn't, or the other way: a shared answer an overflow made would be read as a fact for
// a pair that never overflowed (#mh5kdde).
func (l *lowering) countTypeTargetsCompared(ids []int, targets []*checker.Type, targetClassOf, sourceClassOf []int, overflows [][]bool) {
	for index, id := range ids {
		target := targets[index]
		row := []int{}
		for recordIndex, record := range l.closureRecords {
			if l.checker.IsTypeAssignableTo(record.proven, target) {
				row = append(row, record.function)
			}
			shared := overflows[targetClassOf[index]][sourceClassOf[recordIndex]]
			if own := countTypeOverflowed(l.checker, record.proven, target); own != shared {
				panic(fmt.Sprintf("count type classes comparison (#mh5kdde): record %d's type %s against target %s overflowed %t asked alone, %t asked through its classes", recordIndex, l.checker.TypeToString(record.proven), l.checker.TypeToString(target), own, shared))
			}
		}
		if !slices.Equal(row, l.result.FunctionTypeTargets[id]) {
			panic(fmt.Sprintf("count type classes comparison (#mh5kdde): target %s (type %d) has row %v asked record by record, %v through its classes", l.checker.TypeToString(target), id, row, l.result.FunctionTypeTargets[id]))
		}
	}
}

// countTypeClasses puts types in identity classes: classOf is each type's class, and representatives each class's
// first type, which the checker is asked about for every type of its class. A type is first bucketed by its call
// signatures' shape (countTypeShape), which identical types share, so it's held against only the representatives of
// its own bucket. A fresh type is a class of its own (countTypeFresh) (#mh5kdde).
func countTypeClasses(typeChecker *checker.Checker, types []*checker.Type) (classOf []int, representatives []*checker.Type) {
	classOf = make([]int, len(types))
	buckets := map[string][]int{}
	for index, proven := range types {
		if countTypeFresh(proven) {
			classOf[index] = len(representatives)
			representatives = append(representatives, proven)
			continue
		}
		shape := countTypeShape(typeChecker, proven)
		class := -1
		for _, held := range buckets[shape] {
			if checker.Checker_isTypeIdenticalTo(typeChecker, representatives[held], proven) {
				class = held
				break
			}
		}
		if class < 0 {
			class = len(representatives)
			representatives = append(representatives, proven)
			buckets[shape] = append(buckets[shape], class)
		}
		classOf[index] = class
	}
	return classOf, representatives
}

// countTypeShape is a type's call signatures as identity has to agree on them: how many, and for each its parameter
// count, minimum argument count, rest parameter and type parameter count. Two types it splits are only asked about
// apart, never answered wrongly, so it needs to be cheap more than exact (#mh5kdde).
func countTypeShape(typeChecker *checker.Checker, proven *checker.Type) string {
	signatures := typeChecker.GetSignaturesOfType(proven, checker.SignatureKindCall)
	shape := strconv.AppendInt(nil, int64(len(signatures)), 10)
	for _, signature := range signatures {
		shape = fmt.Appendf(shape, ";%d,%d,%t,%d", len(signature.Parameters()), signature.MinArgumentCount(), signature.HasRestParameter(), len(signature.TypeParameters()))
	}
	return string(shape)
}

// countTypeFresh is whether a type is fresh, or a union or intersection with a fresh member: an object literal's own
// type, whose excess properties assignability judges and identity doesn't see, or a fresh literal type, as the
// checker's isFreshLiteralType tells one. Such a type is never shared with an identical one (#mh5kdde).
func countTypeFresh(proven *checker.Type) bool {
	flags := proven.Flags()
	if flags&checker.TypeFlagsObject != 0 && proven.ObjectFlags()&checker.ObjectFlagsFreshLiteral != 0 {
		return true
	}
	if flags&checker.TypeFlagsFreshable != 0 && proven.AsLiteralType().FreshType() == proven {
		return true
	}
	if flags&(checker.TypeFlagsUnion|checker.TypeFlagsIntersection) != 0 {
		return slices.ContainsFunc(proven.Types(), countTypeFresh)
	}
	return false
}

// assignableRelationKey is the key the checker's assignable relation caches a pair under, its own, pulled from its
// package as class.go pulls isTypeIdenticalTo. A pair whose comparison spent its budget is cached under it marked as
// overflowed (relater.go's checkTypeRelatedToEx) (#mh5kdde).
//
//go:linkname assignableRelationKey github.com/microsoft/TypeScript/tsc/internal/checker.getRelationKey
func assignableRelationKey(source *checker.Type, target *checker.Type, intersectionState checker.IntersectionState, isIdentity bool, ignoreConstraints bool) (checker.CacheHashKey, bool)

// countTypeOverflowed is whether the checker's assignable relation holds source and target marked as overflowed. The
// relation is unexported, so it's read by name as load.RelationOverflowMarked reads it, and a relation that isn't
// where it's expected panics, since only the comparison asks and a comparison that can't read it compares nothing
// (#mh5kdde).
func countTypeOverflowed(typeChecker *checker.Checker, source, target *checker.Type) bool {
	// isTypeRelatedTo relates a fresh literal's regular type, and caches the pair under that (#mh5kdde).
	if source.Flags()&checker.TypeFlagsFreshable != 0 && source.AsLiteralType().FreshType() == source {
		source = source.AsLiteralType().RegularType()
	}
	if target.Flags()&checker.TypeFlagsFreshable != 0 && target.AsLiteralType().FreshType() == target {
		target = target.AsLiteralType().RegularType()
	}
	key, _ := assignableRelationKey(source, target, checker.IntersectionStateNone, false, false)
	relation := reflect.ValueOf(typeChecker).Elem().FieldByName("assignableRelation")
	if !relation.IsValid() || relation.Kind() != reflect.Pointer {
		panic("count type classes comparison (#mh5kdde): the checker's assignableRelation isn't readable, so an overflow can't be compared")
	}
	if relation.IsNil() {
		return false
	}
	results := relation.Elem().FieldByName("results")
	if !results.IsValid() || results.Kind() != reflect.Map || results.Type().Key() != reflect.TypeOf(key) {
		panic("count type classes comparison (#mh5kdde): the assignable relation's results aren't readable, so an overflow can't be compared")
	}
	entry := results.MapIndex(reflect.ValueOf(key))
	if !entry.IsValid() {
		return false
	}
	switch entry.Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
	default:
		panic("count type classes comparison (#mh5kdde): the assignable relation's result isn't a flag word, so an overflow can't be compared")
	}
	return entry.Uint()&uint64(checker.RelationComparisonResultComplexityOverflow) != 0
}
