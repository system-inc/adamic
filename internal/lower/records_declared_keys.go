package lower

import (
	"errors"
	"reflect"
	"slices"
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/load"
)

// In .a, a declared property's type can't lie through an index signature (#npqr3sx, Adamic's ruling
// on #bq3fvdt). A record that declares a property narrower than its index signature's values, as
// interface Options { outFile?: string; [name: string]: string | number | undefined } does, keeps
// outFile a string only while nothing writes a number under its name, and tsc lets three roads do it:
//   - A wider view: Options seen as { [name: string]: string | number | undefined } can be written 3
//     under outFile. widened holds the declared property invariant against the view's index signature.
//   - A dictionary seen as Options: its entry under outFile, whatever it holds, is read as a string.
//     widened holds the dictionary's values against each declared property they'd be read through.
//   - A write through a key that isn't one literal, options[name] = 3 or { [name]: 3 }, which tsc
//     checks against the index signature alone. Only the value written can land under a declared
//     property, so the value is judged as written into it: refused when a declared property can't
//     take it (declaredKeyMisfit), and otherwise as recordNamedWrite judges it, by the whole registry
//     with the property's type as its slot (namedWriteStop), refused or not yet as that says.
//   - A type argument: Options bound to T extends Dict is seen as Dict wherever the generic's body
//     sees T, which no view site outside it shows (Language's n05 on #npqr3sx). Each binding is
//     judged against its constraint as a view is (declaredKeyBindings).
// .ts keeps the checked read instead (recordReadType), which stops loudly at the declared read.

// declaredKeysFix is what each of the refusals offers.
const declaredKeysFix = "give the record no declared keys, or write the declared property by its name"

// recordDeclaredKeyRefusal refuses, in .a, a write that can store a wider value under a declared
// property through the index signature, and a view widened finds reaching one, a type argument bound
// to its constraint included. It runs before recordStorageView (2000), whose not yet would otherwise
// name the same view a storage gap, and refuses only these views: every other view keeps the stop it
// had.
func (l *lowering) recordDeclaredKeyRefusal(node *ast.Node) error {
	if !strings.HasSuffix(l.program.FileName(ast.GetSourceFileOfNode(node)), ".a") {
		return nil
	}
	if err := l.declaredKeyWrite(node); err != nil {
		return err
	}
	l.declaredKeys = true
	defer func() { l.declaredKeys = false }()
	if err := l.declaredKeyBindings(node); err != nil {
		return err
	}
	if source, target := l.recordViewSite(node); source != nil {
		// recordStorageView's own site and its own relation: the whole value seen as the slot, a
		// fresh array's elements and all, so the view is named where that rule would name it.
		own := l.checker.GetTypeAtLocation(source)
		if found := l.widened(own, target, map[[2]*checker.Type]bool{}); found != nil && found.declaredKey != "" {
			return l.wideningRefusal(node, own, target, found)
		}
	}
	// Every other view site refuseWidening judges: a shorthand property, a spread, a branch's implied
	// target. Only its declared-key refusals are this rule's; the rest it makes again at 3100.
	var refused *Refused
	if err := l.stringWideningRefusal(node); errors.As(err, &refused) && refused.Fix == declaredKeysFix {
		return err
	}
	return nil
}

func init() {
	refusalRules.register(1975, "recordDeclaredKeyRefusal", (*lowering).recordDeclaredKeyRefusal)
}

// declaredKeyWrite refuses a write through a key that isn't one literal, into a record with a
// writable index signature, of a value a declared property it can land under can't take
// (declaredKeyMisfit), and judges any other value as recordNamedWrite does in .ts (namedWriteStop):
// the registry with each declared property's type as its slot refuses it or stops it not yet, and a
// value only the optional undefined keeps from a property (undefined under outFile?) is not yet. The
// writes are an element access written to (by =, a compound assignment, ++, a destructuring target
// or a for-in or for-of head), and an object literal's computed name, whose entry is written as the
// literal is made, into its own type and the type it's written into. Where the value written isn't
// one expression's, there's no value for the registry to judge, so it's not yet unless its type is
// each declared property's type itself (writtenWholly): a compound assignment's or an increment's is
// what the operator makes, and a destructuring target's or a loop head's whatever the index
// signature holds.
func (l *lowering) declaredKeyWrite(node *ast.Node) error {
	var key *ast.Node
	var receivers []*checker.Type
	var value *ast.Node
	var written *checker.Type
	switch node.Kind {
	case ast.KindElementAccessExpression:
		if !ast.IsAssignmentTarget(node) {
			return nil
		}
		access := node.AsElementAccessExpression()
		key = access.ArgumentExpression
		receivers = []*checker.Type{l.checker.GetTypeAtLocation(access.Expression)}
		value, written = l.writtenValue(node)
	case ast.KindComputedPropertyName:
		member := node.Parent
		if member == nil || member.Parent == nil || member.Parent.Kind != ast.KindObjectLiteralExpression || ast.IsAssignmentTarget(member.Parent) {
			return nil
		}
		// { outFile: 'a', [name]: 3 } is itself a record declaring outFile, whatever it's written into.
		key = node.AsComputedPropertyName().Expression
		receivers = []*checker.Type{l.checker.GetTypeAtLocation(member.Parent)}
		if contextual := l.slotType(member.Parent); contextual != nil {
			receivers = append(receivers, contextual)
		}
		if member.Kind == ast.KindPropertyAssignment {
			value = member.AsPropertyAssignment().Initializer
			written = l.checker.GetTypeAtLocation(value)
		}
	default:
		return nil
	}
	names, single := l.writtenNames(key)
	if single {
		return nil
	}
	for _, receiver := range receivers {
		for _, member := range l.definedMembers(l.checker.GetNonNullableType(receiver)) {
			for _, info := range l.checker.GetIndexInfosOfType(member) {
				if info.KeyType().Flags() != checker.TypeFlagsString || info.IsReadonly() {
					continue
				}
				stored := written
				if stored == nil {
					stored = info.ValueType()
				}
				if property := l.declaredKeyMisfit(member, stored, names); property != nil {
					return &Refused{Where: l.program.Where(node), What: "a write into a record with declared properties through a key that is not one literal, of a value a declared property it can land under cannot take (" + l.checker.TypeToString(stored) + " written into " + l.checker.TypeToString(member) + " by " + sourceExpression(key) + ": " + property.Name + " is declared " + l.checker.TypeToString(l.checker.GetTypeOfSymbol(property)) + ")", Fix: declaredKeysFix}
				}
				if value == nil {
					if property := l.writtenWholly(member, stored, names); property != nil {
						return l.notYet(node, "a write into a record with declared properties through a key that is not one literal, of no one expression's value, whose type isn't declared property "+property.Name+"'s type itself")
					}
					continue
				}
				if err := l.namedWriteStop(node, member, value, names); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// writtenNames is the names a key can be: single when it's one literal, or a value of one literal's
// type, whose write tsc checks against the property it names, or which lands in the index signature
// under a name no declared property takes. Otherwise it's each name a union of literals can be, or
// the name strings joined by + fold to (constantStringKey), which tsc types as string and doesn't
// check, so ["" + "constructor"] is judged as constructor is; or nil, when it can be any name.
func (l *lowering) writtenNames(key *ast.Node) (map[string]bool, bool) {
	key = ast.SkipParentheses(key)
	if key.Kind == ast.KindStringLiteral || key.Kind == ast.KindNoSubstitutionTemplateLiteral || key.Kind == ast.KindNumericLiteral {
		return nil, true
	}
	proven := l.checker.GetTypeAtLocation(key)
	if proven.Flags()&(checker.TypeFlagsStringLiteral|checker.TypeFlagsNumberLiteral|checker.TypeFlagsUniqueESSymbol) != 0 {
		return nil, true
	}
	if folded, known := constantStringKey(key); known {
		return map[string]bool{folded: true}, false
	}
	if proven.Flags()&checker.TypeFlagsUnion == 0 {
		return nil, false
	}
	names := map[string]bool{}
	for _, member := range proven.Types() {
		switch {
		case member.Flags()&checker.TypeFlagsStringLiteral != 0:
			names[member.AsLiteralType().Value().(string)] = true
		case member.Flags()&checker.TypeFlagsNumberLiteral != 0:
			names[strconv.FormatFloat(reflect.ValueOf(member.AsLiteralType().Value()).Float(), 'f', -1, 64)] = true
		default:
			return nil, false
		}
	}
	return names, false
}

// declaredKeyMisfit is the first declared property of receiver, of those names holds when it's given,
// that a value of type written can't be written under in .a: one tsc's own relation doesn't take, its
// optional undefined included, one the record's one storage can't hold as the property does, or a
// readonly one. These are namedWriteStop's own checks, the ones no rule at a direct write judges, but
// for the optional undefined: undefined under an optional property, which may hold it, is left to
// namedWriteStop, not yet as in .ts.
func (l *lowering) declaredKeyMisfit(receiver, written *checker.Type, names map[string]bool) *ast.Symbol {
	for _, property := range l.checker.GetPropertiesOfType(receiver) {
		if names != nil && !names[property.Name] {
			continue
		}
		if l.checker.IsReadonlySymbol(property) || !l.checker.IsTypeAssignableTo(written, l.checker.GetTypeOfSymbol(property)) || !l.sameRecordStorage(written, l.checker.GetNonMissingTypeOfSymbol(property), map[[2]*checker.Type]bool{}) {
			return property
		}
	}
	return nil
}

// writtenWholly is the first declared property of receiver, of those names holds when it's given, whose
// type a value of type written, its literals at their base type, isn't itself, or nil. It judges a write
// whose value is no one expression, which the registry has nothing to run over.
func (l *lowering) writtenWholly(receiver, written *checker.Type, names map[string]bool) *ast.Symbol {
	widened := checker.Checker_getBaseTypeOfLiteralType(l.checker, l.checker.GetWidenedType(written))
	for _, property := range l.checker.GetPropertiesOfType(receiver) {
		if names != nil && !names[property.Name] {
			continue
		}
		if to := l.checker.GetNonMissingTypeOfSymbol(property); widened != to || !l.sameKeeping(widened, to, map[[2]*checker.Type]bool{}) {
			return property
		}
	}
	return nil
}

// writtenValue is what an assignment or an increment writes into its target: the right side of =, or
// of ??=, ||= or &&=, which write it if they write, and its type; or no node and the type of what an
// arithmetic compound assignment or ++ makes; or neither, when the target isn't one's.
func (l *lowering) writtenValue(target *ast.Node) (*ast.Node, *checker.Type) {
	for target.Parent != nil && target.Parent.Kind == ast.KindParenthesizedExpression {
		target = target.Parent
	}
	parent := target.Parent
	switch {
	case parent == nil:
		return nil, nil
	case parent.Kind == ast.KindBinaryExpression && parent.AsBinaryExpression().Left == target && ast.IsAssignmentOperator(parent.AsBinaryExpression().OperatorToken.Kind):
		switch parent.AsBinaryExpression().OperatorToken.Kind {
		case ast.KindEqualsToken, ast.KindQuestionQuestionEqualsToken, ast.KindBarBarEqualsToken, ast.KindAmpersandAmpersandEqualsToken:
			return parent.AsBinaryExpression().Right, l.checker.GetTypeAtLocation(parent.AsBinaryExpression().Right)
		}
		return nil, l.checker.GetTypeAtLocation(parent)
	case parent.Kind == ast.KindPrefixUnaryExpression || parent.Kind == ast.KindPostfixUnaryExpression:
		return nil, l.checker.GetTypeAtLocation(parent)
	}
	return nil, nil
}

// declaredNarrower is the first property declares has, but those skip names, that can't hold a value
// of type value: written under its name through an index signature, value would break it.
func (l *lowering) declaredNarrower(declares, value *checker.Type, skip func(name string) bool) *ast.Symbol {
	for _, property := range l.checker.GetPropertiesOfType(declares) {
		if skip != nil && skip(property.Name) {
			continue
		}
		declared := l.checker.GetTypeOfSymbol(property)
		if !l.checker.IsTypeAssignableTo(value, declared) || !l.enumAssignable(value, declared) {
			return property
		}
	}
	return nil
}

// declaredKeyWidened is the declared property an index signature reaches when a value of type from is
// seen as type to, in either direction, or nil:
//   - from declares it, and to's writable index signature can write under its name what it can't
//     hold. A copy (fresh) only reads from's properties, and isn't judged this way.
//   - to declares it, from's index signature (readonly too, since this direction only reads) holds
//     what it can't hold, and from doesn't declare it, or skip names it as written after a spread.
func (l *lowering) declaredKeyWidened(from, to *checker.Type, skip map[string]bool, fresh bool) *widening {
	if value := l.recordElement(to); value != nil && !fresh && !l.libraryIndexSignature(to) {
		viewDeclares := func(name string) bool { return l.checker.GetPropertyOfType(to, name) != nil }
		if property := l.declaredNarrower(from, value, viewDeclares); property != nil {
			return &widening{source: l.checker.GetTypeOfSymbol(property), target: value, declaredKey: property.Name}
		}
	}
	for _, info := range l.checker.GetIndexInfosOfType(from) {
		if info.KeyType().Flags() != checker.TypeFlagsString {
			continue
		}
		sourceDeclares := func(name string) bool { return skip[name] || l.checker.GetPropertyOfType(from, name) != nil }
		if property := l.declaredNarrower(to, info.ValueType(), sourceDeclares); property != nil {
			return &widening{source: info.ValueType(), target: l.checker.GetTypeOfSymbol(property), declaredKey: property.Name, dictionary: true}
		}
	}
	return nil
}

// libraryIndexSignature reports whether a type's index signature is the library's own, a parameter
// of Object.values or Object.entries, which reads the record it's handed.
func (l *lowering) libraryIndexSignature(proven *checker.Type) bool {
	for _, info := range l.checker.GetIndexInfosOfType(l.concrete(proven)) {
		if declaration := info.Declaration(); declaration != nil && load.IsLibrary(ast.GetSourceFileOfNode(declaration)) {
			return true
		}
	}
	return false
}

// declaredKeyBindings judges each type argument node binds against its type parameter's constraint, as
// instantiated there, by the declared-key rule both ways (widened, with declaredKeys set), keeping only
// a declared-key finding: Opts bound to T extends Dict lets the generic see Opts as Dict, write a
// number under name through it, and return or keep it so (Language's n05, n26, n27 and n28). The
// bindings are a call's or a new's, inferred or written, read from its resolved signature, which
// reaches generic arrows and methods too; a class's extends clause's; and a generic function's or
// class's, instantiated as the value its slot takes (declaredKeyValueBindings), a call's result included.
func (l *lowering) declaredKeyBindings(node *ast.Node) error {
	switch node.Kind {
	case ast.KindCallExpression, ast.KindNewExpression, ast.KindTaggedTemplateExpression:
		// The bindings the call makes, then the value it yields, which may itself be a generic function
		// its slot instantiates: const put: (target: Options) => void = maker() (g05).
		if err := l.declaredKeySignatureBindings(node, l.checker.GetResolvedSignature(node)); err != nil {
			return err
		}
	case ast.KindExpressionWithTypeArguments:
		clause := node.Parent
		if clause == nil || clause.Kind != ast.KindHeritageClause || clause.AsHeritageClause().Token != ast.KindExtendsKeyword || clause.Parent == nil || !ast.IsClassLike(clause.Parent) {
			return nil
		}
		return l.declaredKeyBaseBindings(node, clause.Parent)
	}
	if node.Kind == ast.KindParenthesizedExpression || !l.isExpression(node) || ast.IsDeclarationName(node) {
		return nil
	}
	if node.Parent != nil && node.Parent.Kind == ast.KindPropertyAccessExpression && node.Parent.AsPropertyAccessExpression().Name() == node {
		return nil // The access is the value, never the name it reads.
	}
	return l.declaredKeyValueBindings(node)
}

// declaredKeyValueBindings judges a generic function or class written as a value where its slot's type
// instantiates it, as contextualGenericSignature instantiates a function value: const make: new (held:
// Options) => Store<Options> = Store binds Store's T to Options with no call or new to judge it at. One
// left generic, with no slot or a generic one, binds nothing until it's called or constructed, and is
// judged then. One whose instantiation can't be read, since it or its slot has more than one signature,
// fails closed under a constraint with a writable string index signature (declaredKeyUnread).
func (l *lowering) declaredKeyValueBindings(node *ast.Node) error {
	own := l.checker.GetTypeAtLocation(node)
	for _, kind := range []checker.SignatureKind{checker.SignatureKindCall, checker.SignatureKindConstruct} {
		signatures := l.checker.GetSignaturesOfType(own, kind)
		var parameters []*checker.Type
		for _, signature := range signatures {
			parameters = append(parameters, signature.TypeParameters()...)
		}
		if len(parameters) == 0 {
			continue
		}
		contextual := l.concrete(l.slotType(node))
		if contextual == nil {
			continue
		}
		slots := l.checker.GetSignaturesOfType(l.withoutUndefined(contextual), kind)
		if len(slots) == 0 || len(slots) == 1 && len(slots[0].TypeParameters()) > 0 {
			continue
		}
		if len(signatures) == 1 && len(slots) == 1 {
			if resolved := instantiateSignatureInContextOf(l.checker, signatures[0], slots[0], nil, nil); resolved != nil {
				if err := l.declaredKeySignatureBindings(node, resolved); err != nil {
					return err
				}
				continue
			}
		}
		if err := l.declaredKeyUnread(node, parameters); err != nil {
			return err
		}
	}
	return nil
}

// declaredKeySignatureBindings judges the bindings an instantiated signature made: its target's type
// parameters, each as its mapper maps it, which the checker keeps unexported and classGenericCall reads
// the same way. A signature that isn't an instantiation binds nothing. One whose bindings can't be read
// fails closed under a constraint that has a writable string index signature (declaredKeyUnread).
func (l *lowering) declaredKeySignatureBindings(node *ast.Node, signature *checker.Signature) error {
	if signature == nil {
		return nil
	}
	if len(signature.TypeParameters()) > 0 {
		return l.declaredKeyUnread(node, signature.TypeParameters())
	}
	target := signature.Target()
	if target == nil || len(target.TypeParameters()) == 0 {
		return nil
	}
	field := reflect.ValueOf(signature).Elem().FieldByName("mapper")
	if !field.IsValid() || field.Kind() != reflect.Pointer || field.IsNil() {
		return l.declaredKeyUnread(node, target.TypeParameters())
	}
	mapper := (*typeMapper)(field.UnsafePointer())
	for _, parameter := range target.TypeParameters() {
		constraint := l.checker.GetConstraintOfTypeParameter(parameter)
		if constraint == nil {
			continue
		}
		if err := l.declaredKeyTypeArgument(node, parameter, instantiateType(l.checker, parameter, mapper), instantiateType(l.checker, constraint, mapper)); err != nil {
			return err
		}
	}
	return nil
}

// declaredKeyBaseBindings judges the type arguments a class's extends clause (node) binds to its base
// class's type parameters, as the base type the checker resolved for the class carries them.
func (l *lowering) declaredKeyBaseBindings(node *ast.Node, class *ast.Node) error {
	symbol := l.checker.GetSymbolOfDeclaration(class)
	if symbol == nil {
		return nil
	}
	for _, base := range l.checker.GetBaseTypes(l.checker.GetDeclaredTypeOfSymbol(symbol)) {
		if base.Symbol() == nil {
			continue
		}
		parameters := l.checker.GetLocalTypeParametersOfClassOrInterfaceOrTypeAlias(base.Symbol())
		if len(parameters) == 0 {
			continue
		}
		if base.ObjectFlags()&checker.ObjectFlagsReference == 0 || base.Target() == nil {
			return l.declaredKeyUnread(node, parameters)
		}
		// The generic class's own reference carries its type parameters where the base carries what
		// they're bound to, outer ones first, so each local parameter is found by its place there.
		declared, given := l.checker.GetTypeArguments(base.Target()), l.checker.GetTypeArguments(base)
		arguments := make([]*checker.Type, len(parameters))
		for index, parameter := range parameters {
			place := slices.Index(declared, parameter)
			if place < 0 || place >= len(given) {
				return l.declaredKeyUnread(node, parameters)
			}
			arguments[index] = given[place]
		}
		mapper := newTypeMapper(parameters, arguments)
		for index, parameter := range parameters {
			constraint := l.checker.GetConstraintOfTypeParameter(parameter)
			if constraint == nil {
				continue
			}
			if err := l.declaredKeyTypeArgument(node, parameter, arguments[index], instantiateType(l.checker, constraint, mapper)); err != nil {
				return err
			}
		}
	}
	return nil
}

// declaredKeyUnread stops, not yet, a binding whose type arguments can't be read when one of the type
// parameters is constrained to a writable string index signature, since what it binds could be a record
// with declared properties seen through it.
func (l *lowering) declaredKeyUnread(node *ast.Node, parameters []*checker.Type) error {
	for _, parameter := range parameters {
		if constraint := l.checker.GetConstraintOfTypeParameter(parameter); constraint != nil && l.recordElement(l.checker.GetBaseConstraintOfType(constraint)) != nil {
			return l.notYet(node, "a generic instantiation whose type argument for "+l.checker.TypeToString(parameter)+", constrained to "+l.checker.TypeToString(constraint)+", can't be read")
		}
	}
	return nil
}

// declaredKeyTypeArgument refuses argument bound to parameter when, seen as its constraint, a declared
// property is reached through an index signature, in either direction (declaredKeyWidened through
// widened). Any other finding is no declared-key one, and is left to the rules that judge it.
func (l *lowering) declaredKeyTypeArgument(node *ast.Node, parameter, argument, constraint *checker.Type) error {
	found := l.widened(argument, constraint, map[[2]*checker.Type]bool{})
	if found == nil || found.declaredKey == "" {
		return nil
	}
	bound := l.checker.TypeToString(argument) + " as " + l.checker.TypeToString(parameter) + ", constrained to " + l.checker.TypeToString(constraint)
	if found.dictionary {
		return &Refused{Where: l.program.Where(node), What: "a dictionary bound as a type argument whose constraint reads it through a declared property narrower than its values (" + bound + ": " + found.declaredKey + " is declared " + l.checker.TypeToString(found.target) + ", and the index signature holds " + l.checker.TypeToString(found.source) + ")", Fix: declaredKeysFix}
	}
	return &Refused{Where: l.program.Where(node), What: "a type argument with declared properties bound to a type parameter constrained to a wider index signature, which can write under a declared property what it does not hold (" + bound + ": " + found.declaredKey + " is declared " + l.checker.TypeToString(found.source) + ", and the index signature holds " + l.checker.TypeToString(found.target) + ")", Fix: declaredKeysFix}
}

// declaredKeyRefusal is a declared-key widening's refusal: one reason for each direction. A view found
// in one branch of a conditional names that branch's type (viewed), not the conditional's own, which
// tsc reduced to the wider side (Language's n24).
func (l *lowering) declaredKeyRefusal(node *ast.Node, own, contextual *checker.Type, found *widening) error {
	if found.viewed != nil {
		own = found.viewed
	}
	if found.dictionary {
		return &Refused{Where: l.program.Where(node), What: "a dictionary read through a declared property narrower than its values (" + l.checker.TypeToString(own) + " seen as " + l.checker.TypeToString(contextual) + ": " + found.declaredKey + " is declared " + l.checker.TypeToString(found.target) + ", and the index signature holds " + l.checker.TypeToString(found.source) + ")", Fix: declaredKeysFix}
	}
	return &Refused{Where: l.program.Where(node), What: "a value with declared properties seen through a wider index signature, which can write under a declared property what it does not hold (" + l.checker.TypeToString(own) + " seen as " + l.checker.TypeToString(contextual) + ": " + found.declaredKey + " is declared " + l.checker.TypeToString(found.source) + ", and the index signature holds " + l.checker.TypeToString(found.target) + ")", Fix: declaredKeysFix}
}
