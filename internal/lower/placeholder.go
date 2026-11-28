package lower

import (
	"maps"
	"sort"
	"strings"
	"sync/atomic"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// placeholderDeclaration admits only literal slot initializers. A shadowed
// undefined binding is an ordinary assertion, as are calls and standalone values.
func (l *lowering) placeholderDeclaration(node *ast.Node) bool {
	if !l.uninitializedInitializer(node) {
		return false
	}
	for node.Parent != nil && (node.Parent.Kind == ast.KindParenthesizedExpression || node.Parent.Kind == ast.KindAsExpression) {
		node = node.Parent
	}
	if node.Parent == nil {
		return false
	}
	switch parent := node.Parent; parent.Kind {
	case ast.KindVariableDeclaration:
		return parent.AsVariableDeclaration().Initializer == node
	case ast.KindPropertyDeclaration:
		return parent.AsPropertyDeclaration().Initializer == node
	case ast.KindParameter:
		return parent.AsParameterDeclaration().Initializer == node
	case ast.KindPropertyAssignment:
		return parent.AsPropertyAssignment().Initializer == node
	case ast.KindBinaryExpression:
		binary := parent.AsBinaryExpression()
		return binary.OperatorToken.Kind == ast.KindEqualsToken && binary.Right == node
	}
	return false
}

// placeholderZero is the observable undefined payload of an unset slot.
func placeholderZero(of ir.Type) ir.Expression {
	if of.IsReference() {
		return ir.Undefined{Of: of}
	}
	if of == ir.MaybeBoolean {
		return ir.MaybeOf{Of: of}
	}
	return zeroValue(of)
}

// Placeholder storage uses the existing counted union: NULL is undefined,
// adamic_null is null, and each present arm retains its ordinary ownership.
func placeholderStorage(of ir.Type) ir.Type {
	if of == 0 {
		return 0
	}
	return ir.Union
}

func (l *lowering) placeholderInitialValue(node *ast.Node, of ir.Type) ir.Expression {
	if node != nil && l.uninitializedInitializer(node) {
		node = ast.SkipParentheses(node)
		for node.Kind == ast.KindAsExpression {
			node = ast.SkipParentheses(node.AsAsExpression().Expression)
		}
		if ast.SkipParentheses(node.AsNonNullExpression().Expression).Kind == ast.KindNullKeyword {
			return fit(ir.Null{}, of)
		}
	}
	return placeholderZero(of)
}

// placeholderResetRefusal refuses a reset to a placeholder (x = undefined!) of a local whose
// readiness flag the backends never declare where the reset can reach it. The flag is declared
// with its local's declaration (ir.Declare) or held in its cell. A parameter without a default and
// a for-of or for-in variable have no declaration, unless a closure captures them into a cell; an
// async frame's local declares its flag in the resume block it starts in, and a reset after an
// await is in another. Each such reset wrote an undeclared flag: C that does not compile,
// JavaScript that throws a ReferenceError. A defaulted or destructured parameter is declared from
// its incoming value, so it has its flag.
func (l *lowering) placeholderResetRefusal(target *ast.Node, local int) error {
	what, fix := "", ""
	var declaration *ast.Node
	if symbol := l.symbol(target); symbol != nil {
		declaration = symbol.ValueDeclaration
	}
	root := declaration
	for root != nil && root.Kind == ast.KindBindingElement {
		root = root.Parent.Parent
	}
	switch {
	case l.result.Locals[local].Global:
		// A global's flag is static, declared once for the whole program.
	case !l.result.Locals[local].Captured && declaration != nil && declaration.Kind == ast.KindParameter && declaration.AsParameterDeclaration().Initializer == nil:
		what, fix = "a parameter", "declare a local with let and assign the placeholder to it, or give the parameter a value of its type"
	case !l.result.Locals[local].Captured && root != nil && root.Kind == ast.KindVariableDeclaration && root.Parent.Parent.Kind == ast.KindForOfStatement:
		what, fix = "a for-of loop variable", "declare a local with let in the loop body and assign the placeholder to it, or give the loop variable a value of its type"
	case !l.result.Locals[local].Captured && root != nil && root.Kind == ast.KindVariableDeclaration && root.Parent.Parent.Kind == ast.KindForInStatement:
		what, fix = "a for-in loop variable", "declare a local with let in the loop body and assign the placeholder to it, or give the loop variable a value of its type"
	case l.result.Locals[local].Function >= 0 && l.result.Functions[l.result.Locals[local].Function].Generator:
		what, fix = "a local of a generator function", "give the local a value of its type, or hold the placeholder in a local of a function that isn't a generator"
	case l.result.Locals[local].Function >= 0 && l.result.Functions[l.result.Locals[local].Function].Async:
		what, fix = "a local of an async function", "give the local a value of its type, or hold the placeholder in a local of a synchronous function"
	case l.result.Locals[local].Function < 0 && moduleAwaits(ast.GetSourceFileOfNode(target).AsNode()):
		what, fix = "a block local of a module that awaits at its top level", "give the local a value of its type, or hold the placeholder in a top-level let"
	}
	if what == "" {
		return nil
	}
	return &Refused{Where: l.program.Where(target), What: "a placeholder reset of " + what, Fix: fix}
}

// moduleAwaits is whether a file awaits outside every function, which makes its top level an
// async function after lowering (flow's async normalization), with its block locals in frame cells.
func moduleAwaits(file *ast.Node) bool {
	var awaits ast.Visitor
	awaits = func(node *ast.Node) bool {
		if ast.IsFunctionLike(node) {
			return false
		}
		if node.Kind == ast.KindAwaitExpression || node.Kind == ast.KindForOfStatement && node.AsForInOrOfStatement().AwaitModifier != nil {
			return true
		}
		return node.ForEachChild(awaits)
	}
	return file.ForEachChild(awaits)
}

func (l *lowering) placeholderOptional(of *checker.Type) bool {
	return l.includesUndefined(of) || l.includesNull(of)
}

// Only a placeholder compared with literal null gets the ruling's loose
// nullish comparison. General coercing equality retains its refusal.
func (l *lowering) placeholderNullishTest(node *ast.Node) bool {
	if node == nil {
		return false
	}
	if node.Kind != ast.KindBinaryExpression {
		node = node.Parent
	}
	if node == nil || node.Kind != ast.KindBinaryExpression {
		return false
	}
	binary := node.AsBinaryExpression()
	if binary.OperatorToken.Kind != ast.KindEqualsEqualsToken && binary.OperatorToken.Kind != ast.KindExclamationEqualsToken {
		return false
	}
	left, right := ast.SkipParentheses(binary.Left), ast.SkipParentheses(binary.Right)
	return left.Kind == ast.KindNullKeyword && l.placeholderReadOrigin(right) != "" || right.Kind == ast.KindNullKeyword && l.placeholderReadOrigin(left) != ""
}

func (l *lowering) placeholderKey(symbol *ast.Symbol) string {
	if symbol == nil || len(symbol.Declarations) == 0 {
		return ""
	}
	declaration := symbol.Declarations[0]
	return l.program.Where(declaration) + ":" + symbol.Name
}

func (l *lowering) placeholderSymbolOrigin(symbol *ast.Symbol) string {
	return l.result.PlaceholderSources[l.placeholderKey(symbol)]
}

func (l *lowering) placeholderOrigin(node *ast.Node) string {
	if node == nil {
		return ""
	}
	return l.placeholderSymbolOrigin(l.symbol(ast.SkipParentheses(node)))
}

// notePlaceholderSlots' answer, PlaceholderSources and PlaceholderViews, reads only the checker and the
// program, unless a type mapper is in force (placeholderOptional reaches concrete) or a caller kept a slot
// (slotType reads slots.explicit). The latent census makes a fresh lowering for each of its units, and each
// walked every module of the program again: 22% of the tsc entry census's CPU (#6xt6p1f). So a census run
// hands its lowerings one placeholderSlotCache, and a lowering under neither takes a copy of the answer the
// first such lowering over its checker and program walked.

// placeholderSlotCache is notePlaceholderSlots' answer for each checker and program a run's lowerings
// walked under no type mapper and no kept slot. Production makes none and walks; the latent census makes
// one a run (lower.go.txt's latentRun, units.go.txt's latentFullSelected), which statecopy shares. A run
// lowers one unit at a time, so it takes no lock.
type placeholderSlotCache struct {
	answers map[placeholderSlotKey]placeholderSlotAnswer
}

// placeholderSlotKey is what the answer is a function of. A run asks the program for a checker per file,
// and the program holds those checkers itself, so the key keeps nothing alive the program doesn't.
type placeholderSlotKey struct {
	checker *checker.Checker
	program *load.Program
}

// placeholderSlotAnswer is one walk's two maps, which no lowering holds: each takes clones of them.
type placeholderSlotAnswer struct {
	sources ir.SortedMap[string, string]
	views   ir.SortedMap[string, string]
}

// placeholderSlotWalks, when a test sets it, counts the walks notePlaceholderSlots makes, so a pin holds a
// run's lowerings to one walk per checker and program. It's nil outside those tests.
var placeholderSlotWalks atomic.Pointer[atomic.Int64]

// placeholderSlotCompare, set only by tests, walks the program again beside every answer taken from the
// cache and panics where the two differ (placeholderSlotsCompared). Tests that set it aren't parallel, and
// no parallel test runs beside a sequential one.
var placeholderSlotCompare atomic.Bool

// notePlaceholderSlots fills PlaceholderSources and PlaceholderViews once a lowering: a copy of the cached
// answer where this lowering may take one, else its own walk.
func (l *lowering) notePlaceholderSlots() {
	// The refusal pass is also used independently of IR construction.
	if l.result == nil {
		l.result = &ir.Program{}
	}
	if l.result.PlaceholderSources != nil {
		return
	}
	// A type mapper or a kept slot can give this lowering an answer of its own, so it walks, and neither
	// reads the cache nor fills it. Nothing calls this under either today, so the guard is held by a pin alone.
	if l.placeholderSlots == nil || l.typeMapper != nil || len(l.slots.explicit) != 0 {
		l.walkPlaceholderSlots()
		return
	}
	key := placeholderSlotKey{checker: l.checker, program: l.program}
	if answer, found := l.placeholderSlots.answers[key]; found {
		if placeholderSlotCompare.Load() {
			l.placeholderSlotsCompared(answer)
		}
		// Clones, since a lowering writes its own maps and every later lowering's copy comes from these.
		l.result.PlaceholderSources, l.result.PlaceholderViews = maps.Clone(answer.sources), maps.Clone(answer.views)
		return
	}
	l.walkPlaceholderSlots()
	if l.placeholderSlots.answers == nil {
		l.placeholderSlots.answers = map[placeholderSlotKey]placeholderSlotAnswer{}
	}
	l.placeholderSlots.answers[key] = placeholderSlotAnswer{sources: maps.Clone(l.result.PlaceholderSources), views: maps.Clone(l.result.PlaceholderViews)}
}

// placeholderSlotsCompared walks the program into a scratch result beside an answer taken from the cache,
// and panics where the fresh maps aren't the cached ones. The lowering goes on from the cached answer, as
// it would without the comparison.
func (l *lowering) placeholderSlotsCompared(answer placeholderSlotAnswer) {
	outer := l.result
	l.result = &ir.Program{}
	l.walkPlaceholderSlots()
	walked := placeholderSlotAnswer{sources: l.result.PlaceholderSources, views: l.result.PlaceholderViews}
	l.result = outer
	var differences []string
	if !maps.Equal(answer.sources, walked.sources) {
		differences = append(differences, "PlaceholderSources: only cached "+placeholderSlotsOnly(answer.sources, walked.sources)+", only fresh "+placeholderSlotsOnly(walked.sources, answer.sources))
	}
	if !maps.Equal(answer.views, walked.views) {
		differences = append(differences, "PlaceholderViews: only cached "+placeholderSlotsOnly(answer.views, walked.views)+", only fresh "+placeholderSlotsOnly(walked.views, answer.views))
	}
	if len(differences) != 0 {
		panic("placeholder slot comparison (#6xt6p1f): the cached answer for this checker and program isn't a fresh walk's; " + strings.Join(differences, "; "))
	}
}

// placeholderSlotsOnly is every entry of one map that the other lacks or holds with another value, as
// key=value, sorted.
func placeholderSlotsOnly(one, other ir.SortedMap[string, string]) string {
	entries := []string{}
	for key, value := range one {
		if held, found := other[key]; !found || held != value {
			entries = append(entries, key+"="+value)
		}
	}
	sort.Strings(entries)
	return "[" + strings.Join(entries, " ") + "]"
}

// walkPlaceholderSlots walks every module of the program for its placeholder slots, into l.result.
// The transfer graph is conservative across branches and loops. A copied local
// never loses its origin when another path writes it. Readiness still proves
// individual writes using the existing CFG; every remaining use as T is checked.
func (l *lowering) walkPlaceholderSlots() {
	if counter := placeholderSlotWalks.Load(); counter != nil {
		counter.Add(1)
	}
	l.result.PlaceholderSources = map[string]string{}
	l.result.PlaceholderViews = map[string]string{}
	type transfer struct {
		to   string
		from *ast.Node
	}
	var transfers []transfer
	var views []transfer
	// A destructured name's transfer: the field it's read from (placeholderBindingField).
	type fieldTransfer struct {
		to    string
		field *ast.Symbol
	}
	var fieldTransfers []fieldTransfer
	inferred := map[string]bool{}
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if ast.IsPartOfTypeNode(node) {
			return false
		}
		var target, value *ast.Node
		switch node.Kind {
		case ast.KindVariableDeclaration:
			target, value = node.Name(), node.AsVariableDeclaration().Initializer
			if ast.IsIdentifier(target) {
				key := l.placeholderKey(l.symbol(target))
				inferred[key] = node.AsVariableDeclaration().Type == nil || l.placeholderOptional(l.checker.GetTypeOfSymbol(l.symbol(target)))
				if value != nil {
					transfers = append(transfers, transfer{key, value})
				}
			} else if target.Kind == ast.KindObjectBindingPattern && node.AsVariableDeclaration().Type == nil && value != nil {
				// const { parent } = node is const parent = node.parent (#s59xxn3): each plain name the pattern
				// reads a field into is a slot of the field's placeholder, its read flowing unset and its first
				// use checked. A typed pattern, a default or a nested pattern stays checked at the binding.
				destructured := l.checker.GetTypeAtLocation(value)
				for _, binding := range target.AsBindingPattern().Elements.Nodes {
					if field := l.placeholderBindingField(binding, destructured); field != nil {
						key := l.placeholderKey(l.symbol(binding.Name()))
						inferred[key] = true
						fieldTransfers = append(fieldTransfers, fieldTransfer{key, field})
					}
				}
			}
		case ast.KindPropertyDeclaration:
			target, value = node.Name(), node.AsPropertyDeclaration().Initializer
			key := l.placeholderKey(l.symbol(target))
			inferred[key] = l.placeholderOptional(l.checker.GetTypeOfSymbol(l.symbol(target)))
			if value != nil {
				transfers = append(transfers, transfer{key, value})
			}
		case ast.KindParameter:
			target, value = node.Name(), node.AsParameterDeclaration().Initializer
		case ast.KindCallExpression:
			if signature := l.checker.GetResolvedSignature(node); signature != nil {
				parameters := signature.Parameters()
				for index, argument := range nodesOf(node.AsCallExpression().Arguments) {
					if index < len(parameters) && l.placeholderOptional(l.checker.GetTypeOfSymbol(parameters[index])) {
						key := l.placeholderKey(parameters[index])
						inferred[key] = true
						transfers = append(transfers, transfer{key, argument})
					}
				}
			}
		case ast.KindPropertyAssignment:
			target, value = node.Name(), node.AsPropertyAssignment().Initializer
			key := l.placeholderKey(l.symbol(target))
			inferred[key] = l.placeholderOptional(l.checker.GetTypeOfSymbol(l.symbol(target)))
			transfers = append(transfers, transfer{key, value})
		case ast.KindBinaryExpression:
			binary := node.AsBinaryExpression()
			if binary.OperatorToken.Kind == ast.KindEqualsToken {
				target, value = ast.SkipParentheses(binary.Left), binary.Right
				if ast.IsIdentifier(target) || target.Kind == ast.KindPropertyAccessExpression {
					if target.Kind == ast.KindPropertyAccessExpression {
						// An any-valued receiver has no declared property symbol. It cannot carry a placeholder slot.
						if symbol := l.symbol(target); symbol != nil {
							key := l.placeholderKey(symbol)
							inferred[key] = l.placeholderOptional(l.checker.GetTypeOfSymbol(symbol))
						}
					}
					transfers = append(transfers, transfer{l.placeholderKey(l.symbol(target)), value})
				}
			}
		}
		if node.Kind == ast.KindObjectLiteralExpression {
			properties := node.AsObjectLiteralExpression().Properties.Nodes
			if len(properties) > 0 && properties[0].Kind == ast.KindSpreadAssignment {
				source := l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(properties[0].AsSpreadAssignment().Expression))
				target := l.checker.GetTypeAtLocation(node)
				for _, field := range l.checker.GetPropertiesOfType(target) {
					if from := l.checker.GetPropertyOfType(source, field.Name); from != nil && len(from.Declarations) > 0 {
						views = append(views, transfer{l.placeholderKey(field), from.Declarations[0].Name()})
					}
				}
			}
		}
		if node.Parent != nil && ast.IsExpression(node) && !ast.IsDeclarationName(node) && !ast.IsPartOfTypeNode(node) {
			if targetType := l.slotType(node); targetType != nil {
				sourceType := l.checker.GetTypeAtLocation(node)
				if sourceType.Flags()&checker.TypeFlagsObject != 0 && targetType.Flags()&checker.TypeFlagsObject != 0 {
					for _, field := range l.checker.GetPropertiesOfType(targetType) {
						if source := l.checker.GetPropertyOfType(sourceType, field.Name); source != nil && len(source.Declarations) > 0 {
							views = append(views, transfer{l.placeholderKey(field), source.Declarations[0].Name()})
						}
					}
				}
			}
		}
		if target != nil && l.uninitializedInitializer(value) {
			key := l.placeholderKey(l.symbol(target))
			if key != "" && l.result.PlaceholderSources[key] == "" {
				l.result.PlaceholderSources[key] = sourceExpression(target)
			}
		}
		return node.ForEachChild(visit)
	}
	for _, file := range l.program.CompilerProgram().GetSourceFiles() {
		if !load.IsLibrary(file) && !load.IsPrelude(file) {
			visit(file.AsNode())
		}
	}
	var origins func(*ast.Node) string
	origins = func(node *ast.Node) string {
		if node == nil {
			return ""
		}
		node = ast.SkipParentheses(node)
		if node.Kind == ast.KindConditionalExpression {
			conditional := node.AsConditionalExpression()
			if origin := origins(conditional.WhenTrue); origin != "" {
				return origin
			}
			return origins(conditional.WhenFalse)
		}
		if node.Kind == ast.KindAsExpression {
			return origins(node.AsAsExpression().Expression)
		}
		if ast.IsIdentifier(node) || node.Kind == ast.KindPropertyAccessExpression {
			return l.placeholderReadOrigin(node)
		}
		return ""
	}
	for changed := true; changed; {
		changed = false
		for _, edge := range views {
			if edge.to == "" || l.result.PlaceholderSources[edge.to] != "" || l.result.PlaceholderViews[edge.to] != "" {
				continue
			}
			if origin := origins(edge.from); origin != "" {
				l.result.PlaceholderViews[edge.to] = origin
				changed = true
			}
		}
		for _, edge := range transfers {
			if edge.to == "" || l.result.PlaceholderSources[edge.to] != "" || !inferred[edge.to] {
				continue
			}
			if origin := origins(edge.from); origin != "" {
				l.result.PlaceholderSources[edge.to] = origin
				changed = true
			}
		}
		for _, edge := range fieldTransfers {
			if edge.to == "" || l.result.PlaceholderSources[edge.to] != "" || !inferred[edge.to] {
				continue
			}
			// The field's own placeholder, or the one a view of it holds, as placeholderDestructuredRead reads it.
			origin := l.placeholderSymbolOrigin(edge.field)
			if origin == "" {
				origin = l.result.PlaceholderViews[l.placeholderKey(edge.field)]
			}
			if origin != "" {
				l.result.PlaceholderSources[edge.to] = origin
				changed = true
			}
		}
	}
}

// placeholderBindingField is the field a declaration's object pattern reads into binding when the binding can
// be a placeholder slot as a plain name is (#s59xxn3): a plain name at the pattern's top level, with no default
// and no rest, read from a field destructured names, by an identifier or a string. Anything else is nil.
func (l *lowering) placeholderBindingField(binding *ast.Node, destructured *checker.Type) *ast.Symbol {
	if binding.Kind != ast.KindBindingElement || binding.Name() == nil || !ast.IsIdentifier(binding.Name()) {
		return nil
	}
	element := binding.AsBindingElement()
	if element.DotDotDotToken != nil || element.Initializer != nil {
		return nil
	}
	field := binding.Name().Text()
	if element.PropertyName != nil {
		if !ast.IsIdentifier(element.PropertyName) && element.PropertyName.Kind != ast.KindStringLiteral {
			return nil
		}
		field = element.PropertyName.Text()
	}
	return l.checker.GetPropertyOfType(destructured, field)
}

// placeholderAllowsUnset is deliberately narrower than acceptsUndefined: a
// template or a container element is a use as T, even when Node tolerates it.
func (l *lowering) placeholderAllowsUnset(node *ast.Node) bool {
	if comparedWithUndefined(node) || l.placeholderNullishTest(node.Parent) {
		return true
	}
	original := node
	for node.Parent != nil && (node.Parent.Kind == ast.KindParenthesizedExpression || node.Parent.Kind == ast.KindAsExpression) {
		node = node.Parent
	}
	parent := node.Parent
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindPropertyAccessExpression:
		// The receiver of a ?. is a presence test (rulings.md:55): the ?. reads undefined where the
		// placeholder is still unset, as Node does (#0d56xpd). Only ?. exactly: a plain read through the
		// placeholder is a use as T, checked.
		if l.placeholderChainObject(original) {
			return true
		}
	case ast.KindIfStatement:
		return parent.AsIfStatement().Expression == node
	case ast.KindWhileStatement:
		return parent.AsWhileStatement().Expression == node
	case ast.KindDoStatement:
		return parent.AsDoStatement().Expression == node
	case ast.KindTypeOfExpression:
		return true
	case ast.KindPrefixUnaryExpression:
		return parent.AsPrefixUnaryExpression().Operator == ast.KindExclamationToken
	case ast.KindConditionalExpression:
		if parent.AsConditionalExpression().Condition == node {
			return true
		}
		return l.placeholderAllowsUnset(parent)
	case ast.KindBinaryExpression:
		binary := parent.AsBinaryExpression()
		if binary.OperatorToken.Kind == ast.KindEqualsToken && binary.Right == node && ast.SkipParentheses(binary.Left).Kind == ast.KindElementAccessExpression {
			return false
		}
		if binary.Left == node && (binary.OperatorToken.Kind == ast.KindQuestionQuestionToken || binary.OperatorToken.Kind == ast.KindAmpersandAmpersandToken || binary.OperatorToken.Kind == ast.KindBarBarToken) {
			return true
		}
		if binary.OperatorToken.Kind == ast.KindEqualsToken && binary.Right == node && l.placeholderOrigin(binary.Left) != "" {
			return true
		}
		if binary.OperatorToken.Kind == ast.KindAmpersandAmpersandToken || binary.OperatorToken.Kind == ast.KindBarBarToken {
			return l.placeholderAllowsUnset(parent)
		}
	case ast.KindVariableDeclaration:
		return parent.AsVariableDeclaration().Initializer == node && l.placeholderOrigin(parent.Name()) != ""
	case ast.KindPropertyAssignment:
		return parent.AsPropertyAssignment().Initializer == node && l.placeholderOrigin(parent.Name()) != ""
	case ast.KindCallExpression:
		callee := ast.SkipParentheses(parent.AsCallExpression().Expression)
		if callee.Kind == ast.KindPropertyAccessExpression {
			receiver := l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(callee.AsPropertyAccessExpression().Expression))
			if l.isArrayType(receiver) || l.isLibraryType(receiver, "Map", "Set") {
				switch callee.Name().Text() {
				case "set", "add", "push", "unshift", "splice", "fill":
					return false
				}
			}
		}
	case ast.KindArrayLiteralExpression, ast.KindSpreadElement:
		return false
	}
	contextual := l.slotType(node)
	return contextual != nil && l.placeholderOptional(contextual)
}

// placeholderChainObject reports whether node, a placeholder's read, is the receiver a property access
// tests with its ?., through parentheses and assertions, and holds an object when it's set (#0d56xpd). The
// read is the placeholder's storage, a union that may hold undefined or null, which placeholderRead hands
// the ?. as the nullable object pointer it tests. A receiver of any other representation keeps its checked
// use, as before. The ?. is questionDot's: a guarded chain's step runs as a plain read (#cdpxjmb), and so
// keeps the placeholder's check.
func (l *lowering) placeholderChainObject(node *ast.Node) bool {
	at := node
	for at.Parent != nil && (at.Parent.Kind == ast.KindParenthesizedExpression || at.Parent.Kind == ast.KindAsExpression) {
		at = at.Parent
	}
	access := at.Parent
	if access == nil || access.Kind != ast.KindPropertyAccessExpression || access.AsPropertyAccessExpression().Expression != at || !l.questionDot(access) {
		return false
	}
	present, known := l.representation(l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(node)))
	return known && present == ir.Object
}

func (l *lowering) placeholderRead(node *ast.Node, value ir.Expression, origin string) ir.Expression {
	if l.placeholderAllowsUnset(node) {
		if !l.placeholderChainObject(node) {
			return value
		}
		// An unset placeholder reads undefined at the ?., and a set one is the object it holds. A storage
		// that isn't a union has no undefined for the ?. to test, so it keeps the placeholder's check below
		// and fails closed (Lower's should on #0d56xpd); placeholderStorage makes every one a union today.
		if value.Type() == ir.Union {
			return l.optionalNullableObject(value)
		}
	}
	use := "use as T"
	at := node
	for at.Parent != nil && at.Parent.Kind == ast.KindParenthesizedExpression {
		at = at.Parent
	}
	if parent := at.Parent; parent != nil {
		switch parent.Kind {
		case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
			use = "property read"
		case ast.KindCallExpression:
			use = "argument"
			if parent.AsCallExpression().Expression == at {
				use = "call"
			}
		case ast.KindReturnStatement:
			use = "return"
		case ast.KindVariableDeclaration, ast.KindBinaryExpression, ast.KindPropertyAssignment:
			use = "assignment"
		case ast.KindArrayLiteralExpression:
			use = "container element"
		}
	}
	of, known := l.representation(l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(node)))
	if !known {
		of = ir.Union
	}
	checkArm := false
	if symbol := l.symbol(node); symbol != nil {
		declared, _ := l.representation(l.checker.GetNonNullableType(l.checker.GetTypeOfSymbol(symbol)))
		checkArm = declared == ir.Union && of != ir.Union
		if checkArm && (of == ir.Weak || of.IsTypedArray()) {
			l.holdUnlowerable(l.notYet(node, "a placeholder narrowing without a present-arm storage check"))
		}
	}
	return ir.PlaceholderUse{Value: value, Origin: origin, Use: use, Path: sourceExpression(node), Where: l.program.Where(node), CheckArm: checkArm, Of: of}
}

func placeholderProven(value ir.Expression, program *ir.Program, ready []bool, fields map[fieldReadiness]int) bool {
	if value == nil {
		return false
	}
	switch value := value.(type) {
	case ir.Read:
		return program.Locals[value.Local].Placeholder != "" && ready != nil && ready[value.Local]
	case ir.Property:
		if local, ok := readinessObject(value.Object); ok {
			if slot, found := fields[fieldReadiness{local, value.Name}]; found {
				return ready[slot]
			}
		}
		return false
	case ir.Box:
		return placeholderProven(value.Value, program, ready, fields)
	case ir.MaybeOf:
		return value.Value != nil
	case ir.Undefined, ir.Null:
		return false
	case ir.PlaceholderUse:
		return true // Its checked boundary produces T or stops.
	case ir.Conditional:
		return placeholderProven(value.WhenTrue, program, ready, fields) && placeholderProven(value.WhenNot, program, ready, fields)
	case ir.Coalesce:
		return value.Panic != nil || placeholderProven(value.Fallback, program, ready, fields)
	}
	return !value.Type().IsMaybe() && value.Type() != ir.Union
}

func placeholderUseValue(use ir.PlaceholderUse, program *ir.Program, ready []bool, fields map[fieldReadiness]int) ir.Expression {
	proven := placeholderProven(use.Value, program, ready, fields)
	status := "checked"
	if proven {
		status = "proven"
	}
	program.PlaceholderChecks = append(program.PlaceholderChecks, ir.PlaceholderCheck{Origin: use.Origin, Use: use.Use, Path: use.Path, Where: use.Where, Status: status})
	armMessage := "placeholder '" + use.Origin + "' has the wrong present arm at " + use.Use + " via " + use.Path
	if use.CheckArm {
		program.PlaceholderChecks = append(program.PlaceholderChecks, ir.PlaceholderCheck{Origin: use.Origin, Use: "present arm at " + use.Use, Path: use.Path, Where: use.Where, Status: "checked"})
	}
	presentValue := func(value ir.Expression) ir.Expression {
		if value.Type() == ir.Union && use.Of != ir.Union {
			return ir.Narrow{Value: value, To: use.Of, Checked: use.CheckArm, Message: armMessage}
		}
		return value
	}
	if proven {
		if use.Value.Type().IsMaybe() {
			return ir.Unwrap{Value: use.Value, Proven: true}
		}
		return presentValue(use.Value)
	}
	message := "placeholder '" + use.Origin + "' is unset at " + use.Use + " via " + use.Path
	index := len(program.Strings)
	program.Strings = append(program.Strings, message)
	return presentValue(ir.Coalesce{Value: use.Value, Panic: ir.StringConstant{Index: index}, Of: use.Value.Type().Present()})
}

// Structural views preserve the source slot's optional storage but do not make
// the target's ordinary T fields legal destinations for an unset write.
func (l *lowering) placeholderReadOrigin(node *ast.Node) string {
	if origin := l.placeholderOrigin(node); origin != "" {
		return origin
	}
	return l.result.PlaceholderViews[l.placeholderKey(l.symbol(node))]
}
