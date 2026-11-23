package lower

import (
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"sync/atomic"
	_ "unsafe"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// resolveCountTypes' FunctionTypeTargets asked the checker whether every closure record's type is assignable to
// every count target: 9.3M questions on a 23K-line program, 64% of the process's CPU, where most targets are the
// records' own function types and the rows come to a handful. So the targets are put in classes and so are the
// records, the checker is asked once per pair of classes, and each answer is spread to every record of its class, in
// record order, for every target of its class (#mh5kdde).
//
// Identity alone isn't a class: it doesn't see everything assignability reads. Identity calls interface
// View { read(value?: number): string }'s read identical to (value?: number) => string, but compareSignaturesRelated
// relates a method's parameters bivariantly and a function's strictly, so (...items: number[]) => string is
// assignable to the method and not to the function, and classing the two by identity dropped rest from read's row
// (internal/oracle/testdata/arguments_length_no_reader.a, J-89g9jm1). The same reads happen at every depth: a method
// in one return type and a function in the other, the two outer types identical.
//
// So two types share a class only where countTypeShape's keys are equal, and the checker holds them identical as a
// second guard. The key names every type nested in a type by its checker id: its signatures' parameter, this, return,
// predicate and type parameter types, its properties' types, its index infos' key and value types, a union's
// members, a reference's target and arguments. The relater compares nested types through the same relation, so equal
// ids at every nested position make the nested comparisons literally the same questions, answered alike. What's left
// is the top level, the facts the relater reads off the two types themselves, and the key holds each of those
// (countTypeShape lists them). A type whose parts the key can't read, or one the relater reads otherwise (a mapped,
// reverse mapped or evolving array type, or anything not an object, union or intersection), is keyed by its own id,
// so it's a class of its own. A fresh type is one too (countTypeFresh): freshness is a flag on the type, and the
// excess-property check reads it.

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

// countTypeClasses puts types in classes: classOf is each type's class, and representatives each class's first type,
// which the checker is asked about for every type of its class. Two types share a class only where their keys
// (countTypeShape) are equal and the checker holds them identical. A fresh type is a class of its own
// (countTypeFresh) (#mh5kdde).
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

// countTypeCacheFlags are object flags the checker sets as it caches what it worked out about a type, not what the
// type is, so two identical types differ in them by what was asked of each first. The key reads past them, or it
// would split the records it's for by what lowering happened to ask (#mh5kdde).
const countTypeCacheFlags = checker.ObjectFlagsCouldContainTypeVariablesComputed | checker.ObjectFlagsCouldContainTypeVariables | checker.ObjectFlagsMembersResolved

// countTypeObjectCacheFlags are an object type's own cache flags. On a union or intersection the same bits, and every
// bit from IsGenericTypeComputed up, are caches or are worked out from the members, which the key names by id
// (#mh5kdde).
const countTypeObjectCacheFlags = checker.ObjectFlagsIdenticalBaseTypeCalculated | checker.ObjectFlagsIdenticalBaseTypeExists

// countTypeDiscriminantFlags are a property's check flags the checker sets as it caches whether it's a discriminant
// (#mh5kdde).
const countTypeDiscriminantFlags = ast.CheckFlagsIsDiscriminant | ast.CheckFlagsIsDiscriminantComputed

// countTypeShape is a type's key: two types share a class only where their keys are equal (countTypeClasses). It names
// every type nested in this one by its checker id, and holds every fact the relater reads off a type itself, so equal
// keys leave assignability nothing to tell apart (the argument heads this file). A split of two types assignability
// answers alike costs one more question; a merge of two it answers differently gives one of them the other's row. The
// facts, each beside what reads it in cohere/TypeScript's tsc/internal/checker/relater.go (#mh5kdde):
//   - the type flags and object flags, past the checker's caches: the relater's paths branch on them throughout.
//   - the symbol's flags: isObjectTypeWithInferableIndex reads object literal, type literal, enum, module and class
//     to give a type literal an implicit index signature an interface lacks.
//   - an alias's symbol and type arguments: two types of one alias relate by the alias's variances.
//   - an instantiated anonymous type's symbol, and a reference's target: signaturesRelatedTo relates two of one
//     origin signature by signature, erased, and others whole; a reference relates by its target's variances.
//   - each call signature, then each construct signature, in the checker's order, by the facts below.
//   - a signature's declaration kind: compareSignaturesRelated's strictVariance relates parameters bivariantly where
//     the target's declaration is a method or a constructor, and strictly otherwise (arguments_length_no_reader.a).
//   - a signature's declaration accessibility: constructorVisibilitiesAreCompatible reads it on construct signatures.
//   - a signature's flags, whole: signaturesRelatedTo reads abstract, and arity reads the rest parameter and the
//     untyped JS mark; identity reads none of them.
//   - a signature's this parameter's type, or that it has none: identity compares this types only where both have
//     one, compareSignaturesRelated where the target has one, and isTopSignature reads it.
//   - a signature's parameter count, minimum argument count, rest parameter and each parameter's type, a rest
//     parameter's being its array.
//   - a signature's return type, and its type predicate's kind, parameter index and type.
//   - a signature's type parameters, and the generic signature it was instantiated from, which
//     isInstantiatedGenericParameter reads to choose whether a parameter is related as a callback.
//   - each property, in the checker's order: its name, symbol flags (optional among them), check flags past the
//     discriminant cache (partial among them, which getUnmatchedProperties reads), whether it's readonly, and its type.
//   - each index info: its key type, its value type and whether it's readonly.
//
// A union or intersection is its members, and a reference its target and type arguments, all by id: the rest of
// either is worked out from those. The relater reads a type's symbol itself only for its recursion identity
// (getRecursionIdentity), which bounds how deep it follows a type that nests itself; that's left out of the key, since
// keying it would put every function in a class of its own, and resolveCountTypesCompare is what would see the two
// answer apart.
func countTypeShape(typeChecker *checker.Checker, proven *checker.Type) string {
	own := "#" + strconv.FormatUint(uint64(proven.Id()), 10)
	flags := proven.Flags()
	objectFlags := proven.ObjectFlags()
	if flags&(checker.TypeFlagsObject|checker.TypeFlagsUnion|checker.TypeFlagsIntersection) == 0 ||
		flags&checker.TypeFlagsObject != 0 && objectFlags&(checker.ObjectFlagsMapped|checker.ObjectFlagsReverseMapped|checker.ObjectFlagsEvolvingArray) != 0 {
		return own
	}
	key := countTypeKey{}
	key.number(int64(flags))
	if flags&checker.TypeFlagsObject != 0 {
		key.number(int64(objectFlags &^ countTypeCacheFlags &^ countTypeObjectCacheFlags))
	} else {
		key.number(int64(objectFlags & (checker.ObjectFlagsIsGenericTypeComputed - 1) &^ countTypeCacheFlags))
	}
	if symbol := proven.Symbol(); symbol != nil {
		key.number(int64(symbol.Flags))
	} else {
		key.mark('-')
	}
	if alias := proven.Alias(); alias != nil {
		key.mark('a')
		key.pointer(alias.Symbol())
		for _, argument := range alias.TypeArguments() {
			key.typeID(argument)
		}
	}
	switch {
	case flags&(checker.TypeFlagsUnion|checker.TypeFlagsIntersection) != 0:
		key.mark('u')
		for _, member := range proven.Types() {
			key.typeID(member)
		}
	case objectFlags&checker.ObjectFlagsReference != 0:
		key.mark('r')
		key.typeID(proven.Target())
		for _, argument := range typeChecker.GetTypeArguments(proven) {
			key.typeID(argument)
		}
	default:
		if objectFlags&checker.ObjectFlagsInstantiated != 0 {
			key.mark('i')
			key.pointer(proven.Symbol())
		}
		key.signatures(typeChecker, proven)
		key.mark('m')
		for _, property := range typeChecker.GetPropertiesOfType(proven) {
			// The type first: working it out can set a check flag the key reads (HasNeverType) (#mh5kdde).
			propertyType := typeChecker.GetTypeOfSymbol(property)
			key.text(property.Name)
			key.number(int64(property.Flags))
			key.number(int64(property.CheckFlags &^ countTypeDiscriminantFlags))
			key.boolean(typeChecker.IsReadonlySymbol(property))
			key.typeID(propertyType)
		}
		key.mark('x')
		for _, info := range typeChecker.GetIndexInfosOfType(proven) {
			key.typeID(info.KeyType())
			key.typeID(info.ValueType())
			key.boolean(info.IsReadonly())
		}
	}
	if key.unreadable {
		return own
	}
	return string(key.shape)
}

// signatures writes the key's part for each call signature, then each construct signature, as countTypeShape lists
// it (#mh5kdde).
func (key *countTypeKey) signatures(typeChecker *checker.Checker, proven *checker.Type) {
	for _, kind := range []checker.SignatureKind{checker.SignatureKindCall, checker.SignatureKindConstruct} {
		key.mark('s')
		for _, signature := range typeChecker.GetSignaturesOfType(proven, kind) {
			// A synthesized signature has no declaration, which strictVariance reads as strict, as it does a
			// function's; its own mark keeps it apart from every declared kind anyway (#mh5kdde).
			if declaration := signature.Declaration(); declaration != nil {
				key.number(int64(declaration.Kind))
				key.number(int64(declaration.ModifierFlags() & ast.ModifierFlagsNonPublicAccessibilityModifier))
			} else {
				key.mark('-')
			}
			key.number(int64(signature.Flags()))
			if this := signature.ThisParameter(); this != nil {
				key.mark('t')
				key.typeID(typeChecker.GetTypeOfSymbol(this))
			} else {
				key.mark('-')
			}
			key.number(int64(len(signature.Parameters())))
			key.number(int64(signature.MinArgumentCount()))
			key.boolean(signature.HasRestParameter())
			for _, parameter := range signature.Parameters() {
				key.typeID(typeChecker.GetTypeOfSymbol(parameter))
			}
			key.typeID(typeChecker.GetReturnTypeOfSignature(signature))
			if predicate := typeChecker.GetTypePredicateOfSignature(signature); predicate != nil {
				key.mark('p')
				key.number(int64(predicate.Kind()))
				key.number(int64(predicate.ParameterIndex()))
				if predicate.Type() != nil {
					key.typeID(predicate.Type())
				} else {
					key.mark('-')
				}
			}
			key.mark('<')
			for _, parameter := range signature.TypeParameters() {
				key.typeID(parameter)
			}
			key.pointer(signature.Target())
			key.mark(';')
		}
	}
}

// countTypeKey is countTypeShape's key as it's written, and whether a part of it couldn't be read, which makes the type
// a class of its own (#mh5kdde).
type countTypeKey struct {
	shape      []byte
	unreadable bool
}

func (key *countTypeKey) number(value int64) {
	key.shape = strconv.AppendInt(key.shape, value, 10)
	key.shape = append(key.shape, ',')
}

func (key *countTypeKey) boolean(value bool) {
	key.shape = strconv.AppendBool(key.shape, value)
	key.shape = append(key.shape, ',')
}

func (key *countTypeKey) text(value string) {
	key.shape = strconv.AppendQuote(key.shape, value)
	key.shape = append(key.shape, ',')
}

func (key *countTypeKey) mark(value byte) {
	key.shape = append(key.shape, value)
}

// pointer names a symbol or a signature, which have no checker id the key can read, by its address, which the
// classing outlives none of (#mh5kdde).
func (key *countTypeKey) pointer(value any) {
	key.shape = fmt.Appendf(key.shape, "%p,", value)
}

// typeID names a nested type by its checker id. A type the checker didn't give is a part the key can't read
// (#mh5kdde).
func (key *countTypeKey) typeID(proven *checker.Type) {
	if proven == nil {
		key.unreadable = true
		return
	}
	key.shape = strconv.AppendUint(key.shape, uint64(proven.Id()), 10)
	key.shape = append(key.shape, ',')
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
