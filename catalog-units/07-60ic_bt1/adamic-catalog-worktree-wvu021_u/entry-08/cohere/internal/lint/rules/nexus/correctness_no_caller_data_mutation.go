package nexus

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/ecmascript/reference"
	"github.com/system-inc/cohere/internal/lint/ecmascript/text"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

const correctnessNoCallerDataMutationId = "callerDataMutation"

// The rule's messages, whose wording lives in `policy/messages/correctness-no-caller-data-mutation.json`.
var (
	correctnessNoCallerDataMutationText             = policy.MessageOf("nexus/correctness-no-caller-data-mutation", correctnessNoCallerDataMutationId)
	correctnessNoCallerDataMutationBareProcessState = policy.MessageOf("nexus/correctness-no-caller-data-mutation", "processStateWithoutReason")
	correctnessNoCallerDataMutationBareMutates      = policy.MessageOf("nexus/correctness-no-caller-data-mutation", "mutatesWithoutReason")
	correctnessNoCallerDataMutationUnknownMutates   = policy.MessageOf("nexus/correctness-no-caller-data-mutation", "mutatesUnknownParameter")
)

// correctnessNoCallerDataMutationProcessStateTag declares a type as state the process shares.
const correctnessNoCallerDataMutationProcessStateTag = "processState"

// correctnessNoCallerDataMutationMutatesTag declares a method's parameter an out-parameter by contract.
const correctnessNoCallerDataMutationMutatesTag = "mutates"

func correctnessNoCallerDataMutationMessage() rule.Message {
	return rule.Message{
		Id:          correctnessNoCallerDataMutationId,
		Description: correctnessNoCallerDataMutationText.Render(nil),
	}
}

// CorrectnessNoCallerDataMutation reports a write through a parameter into data declared in project
// code: a property assignment, an update, a `delete`, or a call of a mutating collection method.
//
//	invalid: function addUsage(into: UsageInterface, from: UsageInterface) { into.inputTokens += from.inputTokens; }
//	invalid: function check(diffs: DiffInterface[]) { diffs.push({ path }); }
//	invalid: function merge(messages: MessageInterface[]) { return messages.sort(byDate); }
//	invalid: function sign(request: { headers: Record<string, string> }) { request.headers['host'] = host; }
//	valid:   function paint(context: CanvasRenderingContext2D) { context.fillStyle = 'red'; }
//	valid:   function listen(port: MessagePort) { port.onmessage = handle; }
//	valid:   rows.forEach(function(row) { row.seen = true; });
//
// # Where it came from
//
// Kirk's ruling on `#e3yxyx7`, 2026-10-01. Upstream's `no-param-reassign` with `props: true` adds 143
// findings on ahra, and they are two different things. About 80 are writes into a caller's data object
// (`into.cacheWriteTokens +=`, `hub.lastEmittedHadTeam =`, `entry.position =`,
// `opsNavigationLink.active =`, `packageJson.*`), which are hidden side effects. About 55 are writes to
// host objects whose API is setting properties (a canvas context's `fillStyle`, a DOM event or element,
// `messagePort.onmessage`), which are how those APIs work. Upstream can only tell them apart with an
// allowance list of parameter names (`ignorePropertyModificationsFor`), which the doctrine forbids, so
// this rule draws the line where the type does.
//
// # The shape, exactly
//
//  1. A write. The left side of any assignment operator, the operand of `++` or `--`, the operand of
//     `delete`, or the receiver of a call of a mutating collection method: for an array `push`,
//     `unshift`, `pop`, `shift`, `splice`, `sort`, `reverse`, `fill`, `copyWithin`; for a Map or
//     WeakMap `set`, `delete`, and a Map's `clear`; for a Set or WeakSet `add`, `delete`, and a Set's
//     `clear`. Each method is resolved to the default library's declaration, so a project type with a
//     `push` of its own is not read as one. `sort` and `reverse` count although they return the array:
//     they sort it in place, and `return messages.sort(...)` hands the caller back its own list
//     reordered.
//  2. Reached from a parameter. The written expression is a chain of property accesses, element
//     accesses and non-null assertions whose root is an identifier resolving, by symbol, to a parameter
//     of a function in this file, or to a name destructured from one (`function f({ entry })`).
//  3. Into project data. The written property is declared in a source file of this program rather
//     than in a declaration file (`lib.*.d.ts`, `lib.dom.d.ts`, `@types/*`, a package's own types). A
//     write with no declared property (an element of an array or tuple, or an index-signature key)
//     counts when the receiver is the caller's data by this same test and its type is an array or
//     tuple, one of the default library's data-shaped mapped types (`Record`, `Partial`, `Required`,
//     `Pick`, `Omit`), or a type a source file declares. Only a property name or a literal key is
//     resolved as a property: `buffer[offset]` resolves `offset` to the variable, which a source file
//     declares. For a collection call, the collection is either the parameter itself or a property
//     reached from it that passes the same test. `param.canvas.width = 10` is silent: `width` is the
//     DOM's, and so is `element.dataset['x']`, a Record the DOM declares.
//
// # What it declines
//
// Each is a missed finding at worst, never a false one.
//
//   - A callback's parameter: a function expression or arrow passed straight to a call or `new`, or
//     invoked on the spot.
//     `reduce`'s accumulator is the reduce's own value, made at the call, and an element `forEach`
//     hands over belongs to the code iterating it, right there on the page. Measured on ahra
//     2026-10-03: 31 of 175 findings were callback parameters, every one a reduce accumulator, a
//     `forEach`/`map` element, or a syntax-tree visitor's node. The cost is the case where the iterated
//     data is itself the caller's: `Object.entries(byGroup).map(function([group, items]) { return
//     items.sort(...); })` reorders the caller's arrays, and is not reported.
//   - A setter. `param.value = x` through a project class's `set value()` calls code the class wrote to
//     be called that way, which is an API rather than a write into data.
//   - A parameter the function reassigns anywhere (`options = { ...options }`): a write may land in the
//     copy, and telling which is flow analysis this rule does not do.
//   - A write through `this`, through a call's result, or through an `as` cast: the chain must reach the
//     parameter by plain access.
//   - A write whose target type is `any`: no declaration, and no receiver type to read.
//   - A write through a parameter whose declared type is process state (below).
//   - A write through a method's parameter that the method, or a method it overrides or implements,
//     declares an out-parameter (below).
//
// # Process state: `@processState <why>`
//
// Some objects are not a caller's data but state the whole process shares: a long-lived hub on
// `globalThis` that every helper is handed, an observer, a registry. Writing through one is what it is
// for, and suppressing each write says the same sentence at every site. So a type can declare it once,
// at its own declaration, with a doc tag and the reason:
//
//	/** The kingdom's live observer, one per process. @processState every helper updates its counters */
//	export interface AhraOsObserverInterface { ... }
//
// A write through a parameter whose declared type is that type is then exempt. The match is exact:
// the parameter's declared type, after alias resolution, is the tagged symbol itself (an interface,
// a class, or a type alias). A union containing it, `Partial<Hub>`, `Readonly<Hub>`, `Hub[]`, and an
// interface extending it are each a different type and are not exempt, so a tag cannot spread past the
// one type that says it. An optional parameter is a union with `undefined` and is not exempt either.
//
// The reason is required. A bare `@processState` is reported where it is written, and it exempts
// nothing, so a tag can never hide a write without saying why.
//
// # An out-parameter by contract: `@mutates <parameter> <why>`
//
// Some methods are handed a record to fill in, and that is their contract rather than a side effect.
// Base's `OrmPersistedScheduledExecutable.run(entity, context)` is the case: the scheduler persists
// `entity` after `run` returns, so every override writes its outcome onto it. Stating that at each of
// api's 26 overrides says one sentence 26 times. So the method that defines the contract says it once:
//
//	/** @mutates entity the scheduler persists the record after run returns */
//	protected abstract run(entity: JobRecordInterface, context: ContextInterface): Promise<void>;
//
// A write through that parameter is then exempt in the tagged method and in every method that
// overrides or implements it, at any depth. The override is found through the checker: the class's
// and interface's base types are asked for a member of the method's name, and that member's own
// declarations are read, so a method that only shares the name, or the parameter's name, with a tagged
// one overrides nothing and still reports. An override may rename the parameter, so it is matched by
// position in the tagged declaration's list, never by name. Every other parameter of the method still
// reports, and so does a plain function, which has nothing to override.
//
// The tag sits on a method declaration or an interface's method signature. A tag without a reason is
// reported where it is written and exempts nothing, as a bare `@processState` is, and so is one whose
// first word names none of its method's parameters.
//
// # No fix
//
// The repair is a contract change: return the new value, build a copy, or rename the parameter to say
// it is an out-parameter and suppress with that reason. Which one is the author's call.
var CorrectnessNoCallerDataMutation = rule.Rule{
	Name: "nexus/correctness-no-caller-data-mutation",

	// Symbol identity ties the chain's root to a parameter through shadowing, and says where each
	// written property and each mutating method is declared.
	NeedsTypeChecker: true,

	// Whether `Record` and its kin and the mutating methods are the default library's, through
	// type_checking.
	ProgramReads: rule.ReadsCompilerOptions | rule.ReadsDefaultLibrary,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.TypeChecker == nil {
			return nil
		}
		reportBareTags := func(node *ast.Node) {
			for _, tag := range correctnessNoCallerDataMutationProcessStateTags(ctx.SourceFile, node) {
				if correctnessNoCallerDataMutationTagReason(ctx.SourceFile, tag) == "" {
					ctx.ReportNode(tag, rule.Message{
						Id:          correctnessNoCallerDataMutationBareProcessState.Id,
						Description: correctnessNoCallerDataMutationBareProcessState.Render(nil),
					})
				}
			}
		}
		reportMutatesTags := func(node *ast.Node) {
			for _, tag := range correctnessNoCallerDataMutationTagsOf(ctx.SourceFile, node, correctnessNoCallerDataMutationMutatesTag) {
				parameter, reason := correctnessNoCallerDataMutationMutatesTagParts(ctx.SourceFile, tag)
				message := correctnessNoCallerDataMutationBareMutates
				switch {
				case correctnessNoCallerDataMutationParameterIndex(node, parameter) < 0:
					message = correctnessNoCallerDataMutationUnknownMutates
				case reason != "":
					continue
				}
				ctx.ReportNode(tag, rule.Message{Id: message.Id, Description: message.Render(nil)})
			}
		}
		return rule.Listeners{
			ast.KindInterfaceDeclaration: reportBareTags,
			ast.KindTypeAliasDeclaration: reportBareTags,
			ast.KindClassDeclaration:     reportBareTags,
			ast.KindMethodDeclaration:    reportMutatesTags,
			ast.KindMethodSignature:      reportMutatesTags,
			ast.KindBinaryExpression: func(node *ast.Node) {
				binary := node.AsBinaryExpression()
				if !ast.IsAssignmentOperator(binary.OperatorToken.Kind) {
					return
				}
				correctnessNoCallerDataMutationCheckTarget(ctx, binary.Left)
			},
			ast.KindPrefixUnaryExpression: func(node *ast.Node) {
				prefix := node.AsPrefixUnaryExpression()
				if prefix.Operator != ast.KindPlusPlusToken && prefix.Operator != ast.KindMinusMinusToken {
					return
				}
				correctnessNoCallerDataMutationCheckTarget(ctx, prefix.Operand)
			},
			ast.KindPostfixUnaryExpression: func(node *ast.Node) {
				correctnessNoCallerDataMutationCheckTarget(ctx, node.AsPostfixUnaryExpression().Operand)
			},
			ast.KindDeleteExpression: func(node *ast.Node) {
				correctnessNoCallerDataMutationCheckTarget(ctx, node.AsDeleteExpression().Expression)
			},
			ast.KindCallExpression: func(node *ast.Node) {
				correctnessNoCallerDataMutationCheckCall(ctx, node)
			},
		}
	},
}

// correctnessNoCallerDataMutationWriters are the mutating methods of the default library's
// collections, by the interface that declares them.
var correctnessNoCallerDataMutationWriters = map[string]map[string]bool{
	"Array": {"push": true, "unshift": true, "pop": true, "shift": true, "splice": true, "sort": true,
		"reverse": true, "fill": true, "copyWithin": true},
	"Map":     {"set": true, "delete": true, "clear": true},
	"Set":     {"add": true, "delete": true, "clear": true},
	"WeakMap": {"set": true, "delete": true},
	"WeakSet": {"add": true, "delete": true},
}

// correctnessNoCallerDataMutationDataShapes are the default library's mapped types that only reshape
// data: an object typed through one of them is a record of keys and values, never a host object.
// `Readonly` is not here, since nothing can be written through it.
var correctnessNoCallerDataMutationDataShapes = map[string]bool{
	"Record": true, "Partial": true, "Required": true, "Pick": true, "Omit": true,
}

// correctnessNoCallerDataMutationWriterNames is every writer's name, the cheap gate a call passes
// before any checker question.
var correctnessNoCallerDataMutationWriterNames = func() map[string]bool {
	names := map[string]bool{}
	for _, writers := range correctnessNoCallerDataMutationWriters {
		for name := range writers {
			names[name] = true
		}
	}
	return names
}()

// correctnessNoCallerDataMutationCheckTarget judges the target of an assignment, an update or a
// `delete`.
func correctnessNoCallerDataMutationCheckTarget(ctx rule.Context, target *ast.Node) {
	target = ast.SkipParentheses(target)
	if target.Kind != ast.KindPropertyAccessExpression && target.Kind != ast.KindElementAccessExpression {
		return
	}
	parameter := correctnessNoCallerDataMutationParameterReached(ctx, target)
	if parameter == nil {
		return
	}
	if !correctnessNoCallerDataMutationWritesData(ctx, target) {
		return
	}
	if tagged, isProcessState := correctnessNoCallerDataMutationProcessStateOf(ctx, parameter); isProcessState {
		ctx.Note(tagged)
		return
	}
	if contract, isOutParameter := correctnessNoCallerDataMutationOutParameterOf(ctx, parameter); isOutParameter {
		ctx.Note(contract)
		return
	}
	ctx.ReportNode(target, correctnessNoCallerDataMutationMessage())
}

// correctnessNoCallerDataMutationCheckCall judges a call of a mutating collection method.
func correctnessNoCallerDataMutationCheckCall(ctx rule.Context, call *ast.Node) {
	callee := ast.SkipParentheses(call.AsCallExpression().Expression)
	if callee.Kind != ast.KindPropertyAccessExpression {
		return
	}
	access := callee.AsPropertyAccessExpression()
	method := access.Name()
	if method.Kind != ast.KindIdentifier || !correctnessNoCallerDataMutationWriterNames[method.Text()] {
		return
	}
	receiver := ast.SkipParentheses(access.Expression)
	parameter := correctnessNoCallerDataMutationParameterReached(ctx, receiver)
	if parameter == nil {
		return
	}
	if !correctnessNoCallerDataMutationIsWriter(ctx, method) {
		return
	}
	// The collection is the parameter itself, or data reached from it.
	if receiver.Kind != ast.KindIdentifier && !correctnessNoCallerDataMutationWritesData(ctx, receiver) {
		return
	}
	if tagged, isProcessState := correctnessNoCallerDataMutationProcessStateOf(ctx, parameter); isProcessState {
		ctx.Note(tagged)
		return
	}
	if contract, isOutParameter := correctnessNoCallerDataMutationOutParameterOf(ctx, parameter); isOutParameter {
		ctx.Note(contract)
		return
	}
	ctx.ReportNode(callee, correctnessNoCallerDataMutationMessage())
}

// correctnessNoCallerDataMutationParameterReached walks an access chain to its root and returns the
// root's symbol when it is a parameter this rule judges (not a callback's, and never reassigned), or
// nil.
func correctnessNoCallerDataMutationParameterReached(ctx rule.Context, expression *ast.Node) *ast.Symbol {
	root := expression
	for {
		root = ast.SkipParentheses(root)
		switch root.Kind {
		case ast.KindPropertyAccessExpression:
			root = root.AsPropertyAccessExpression().Expression
			continue
		case ast.KindElementAccessExpression:
			root = root.AsElementAccessExpression().Expression
			continue
		case ast.KindNonNullExpression:
			root = root.AsNonNullExpression().Expression
			continue
		}
		break
	}
	if root.Kind != ast.KindIdentifier {
		return nil
	}
	symbol := ctx.TypeChecker.GetSymbolAtLocation(root)
	parameter := correctnessNoCallerDataMutationParameterOf(ctx, symbol)
	if parameter == nil {
		return nil
	}
	function := parameter.Parent
	if function == nil || correctnessNoCallerDataMutationIsCallback(function) {
		return nil
	}
	if correctnessNoCallerDataMutationIsReassigned(ctx, function, symbol) {
		return nil
	}
	return symbol
}

// correctnessNoCallerDataMutationParameterOf returns the parameter a symbol is declared by, directly or
// through a destructuring pattern, or nil.
func correctnessNoCallerDataMutationParameterOf(ctx rule.Context, symbol *ast.Symbol) *ast.Node {
	for _, declaration := range rule.DeclarationsIn(ctx.SourceFile, symbol) {
		switch declaration.Kind {
		case ast.KindParameter:
			return declaration
		case ast.KindBindingElement:
			ancestor := declaration.Parent
			for ancestor != nil && (ancestor.Kind == ast.KindObjectBindingPattern ||
				ancestor.Kind == ast.KindArrayBindingPattern || ancestor.Kind == ast.KindBindingElement) {
				ancestor = ancestor.Parent
			}
			if ancestor != nil && ancestor.Kind == ast.KindParameter {
				return ancestor
			}
		}
	}
	return nil
}

// correctnessNoCallerDataMutationIsCallback says whether a function is a function expression or arrow
// handed straight to a call or `new`. An immediately invoked one counts too: its argument is written
// at the same place, as visible as a callback's.
func correctnessNoCallerDataMutationIsCallback(function *ast.Node) bool {
	if function.Kind != ast.KindFunctionExpression && function.Kind != ast.KindArrowFunction {
		return false
	}
	outer := function.Parent
	for outer != nil && outer.Kind == ast.KindParenthesizedExpression {
		outer = outer.Parent
	}
	return outer != nil && (outer.Kind == ast.KindCallExpression || outer.Kind == ast.KindNewExpression)
}

// correctnessNoCallerDataMutationIsReassigned says whether any reference to the parameter, in its
// function's parameter list or body, writes the binding itself.
func correctnessNoCallerDataMutationIsReassigned(ctx rule.Context, function *ast.Node, symbol *ast.Symbol) bool {
	name := symbol.Name
	reassigned := false
	var visit func(node *ast.Node) bool
	visit = func(node *ast.Node) bool {
		if reassigned {
			return true
		}
		if node.Kind == ast.KindIdentifier && node.Text() == name && reference.WritesToBinding(node) {
			resolved := ctx.TypeChecker.GetSymbolAtLocation(node)
			if parent := node.Parent; parent != nil && parent.Kind == ast.KindShorthandPropertyAssignment {
				if valueSymbol := ctx.TypeChecker.GetShorthandAssignmentValueSymbol(parent); valueSymbol != nil {
					resolved = valueSymbol
				}
			}
			if resolved == symbol {
				reassigned = true
				return true
			}
		}
		node.ForEachChild(visit)
		return false
	}
	for _, parameter := range function.Parameters() {
		if initializer := parameter.Initializer(); initializer != nil {
			visit(initializer)
		}
	}
	if body := function.Body(); body != nil {
		visit(body)
	}
	return reassigned
}

// correctnessNoCallerDataMutationWritesData says whether a written access lands in data project code
// declares: a property declared in a source file and not a setter, or an element of an array, a tuple,
// a `Record` or its kin, or a type a source file declares.
func correctnessNoCallerDataMutationWritesData(ctx rule.Context, target *ast.Node) bool {
	var key *ast.Node
	var receiver *ast.Node
	switch target.Kind {
	case ast.KindPropertyAccessExpression:
		access := target.AsPropertyAccessExpression()
		key = access.Name()
		receiver = access.Expression
	case ast.KindElementAccessExpression:
		access := target.AsElementAccessExpression()
		key = access.ArgumentExpression
		receiver = access.Expression
	default:
		return false
	}
	// Only a name or a literal key names a property. `buffer[offset]` asked about `offset` answers with
	// the variable, which a source file declares, and would make every element write look like data.
	if target.Kind == ast.KindPropertyAccessExpression || ast.IsStringOrNumericLiteralLike(key) {
		if symbol := ctx.TypeChecker.GetSymbolAtLocation(key); symbol != nil && len(symbol.Declarations) > 0 {
			return symbol.Flags&ast.SymbolFlagsSetAccessor == 0 && rule.IsDeclaredInASourceFile(symbol)
		}
	}
	// No declared property, so the receiver's type decides, and only for a receiver that is itself the
	// caller's data: the parameter, or a property of it that passes this test. `element.dataset['x']` is
	// a Record the host declares.
	owner := receiver
	for {
		owner = ast.SkipParentheses(owner)
		if owner.Kind != ast.KindNonNullExpression {
			break
		}
		owner = owner.AsNonNullExpression().Expression
	}
	if owner.Kind != ast.KindIdentifier && !correctnessNoCallerDataMutationWritesData(ctx, owner) {
		return false
	}
	receiverType := ctx.TypeChecker.GetTypeAtLocation(receiver)
	if receiverType == nil {
		return false
	}
	receiverType = checker.Checker_getApparentType(ctx.TypeChecker, receiverType)
	if checker.Checker_isArrayOrTupleType(ctx.TypeChecker, receiverType) {
		return true
	}
	if alias := checker.Type_alias(receiverType); alias != nil && alias.Symbol() != nil &&
		correctnessNoCallerDataMutationDataShapes[alias.Symbol().Name] &&
		type_checking.IsSymbolFromDefaultLibrary(ctx.Program, alias.Symbol()) {
		return true
	}
	return rule.IsDeclaredInASourceFile(checker.Type_symbol(receiverType))
}

// correctnessNoCallerDataMutationIsWriter says whether a method name resolves to a mutating method of
// a default-library collection, with every declaration there.
func correctnessNoCallerDataMutationIsWriter(ctx rule.Context, name *ast.Node) bool {
	symbol := ctx.TypeChecker.GetSymbolAtLocation(name)
	if symbol == nil || len(symbol.Declarations) == 0 {
		return false
	}
	for _, declaration := range symbol.Declarations {
		owner := declaration.Parent
		if owner == nil || owner.Kind != ast.KindInterfaceDeclaration || owner.Name() == nil {
			return false
		}
		if !correctnessNoCallerDataMutationWriters[owner.Name().Text()][name.Text()] {
			return false
		}
		file := ast.GetSourceFileOfNode(declaration)
		if file == nil || !type_checking.IsSourceFileDefaultLibrary(ctx.Program, file) {
			return false
		}
	}
	return true
}

// correctnessNoCallerDataMutationProcessStateOf says whether a parameter's declared type is a type
// declared as process state: the type's own symbol, or the alias it was written through, carries a
// `@processState` tag with a reason. Exact on purpose, so a union, a mapped or wrapped type, an array
// and a subtype each fail it.
//
// It answers with the note key the exemption is counted under: the tagged type's name and the file
// declaring it, so --coverage can say which type exempted how many writes.
func correctnessNoCallerDataMutationProcessStateOf(ctx rule.Context, parameter *ast.Symbol) (string, bool) {
	declared := checker.Checker_getTypeOfSymbol(ctx.TypeChecker, parameter)
	if declared == nil {
		return "", false
	}
	candidates := []*ast.Symbol{checker.Type_symbol(declared)}
	if alias := checker.Type_alias(declared); alias != nil {
		candidates = append(candidates, alias.Symbol())
	}
	for _, candidate := range candidates {
		if candidate == nil {
			continue
		}
		for _, declaration := range candidate.Declarations {
			switch declaration.Kind {
			case ast.KindInterfaceDeclaration, ast.KindTypeAliasDeclaration, ast.KindClassDeclaration:
			default:
				continue
			}
			file := ast.GetSourceFileOfNode(declaration)
			for _, tag := range correctnessNoCallerDataMutationProcessStateTags(file, declaration) {
				if correctnessNoCallerDataMutationTagReason(file, tag) != "" {
					return candidate.Name + " in " + file.FileName().AsString(), true
				}
			}
		}
	}
	return "", false
}

// correctnessNoCallerDataMutationProcessStateTags returns the `@processState` tags in a declaration's
// doc comments.
func correctnessNoCallerDataMutationProcessStateTags(file *ast.SourceFile, declaration *ast.Node) []*ast.Node {
	return correctnessNoCallerDataMutationTagsOf(file, declaration, correctnessNoCallerDataMutationProcessStateTag)
}

// correctnessNoCallerDataMutationTagsOf returns the tags of one name in a declaration's doc comments.
func correctnessNoCallerDataMutationTagsOf(file *ast.SourceFile, declaration *ast.Node, tagName string) []*ast.Node {
	if file == nil {
		return nil
	}
	var tags []*ast.Node
	for _, doc := range declaration.JSDoc(file) {
		list := doc.AsJSDoc().Tags
		if list == nil {
			continue
		}
		for _, tag := range list.Nodes {
			if tag.Kind != ast.KindJSDocUnknownTag {
				continue
			}
			if name := tag.TagName(); name != nil && name.Text() == tagName {
				tags = append(tags, tag)
			}
		}
	}
	return tags
}

// correctnessNoCallerDataMutationTagReason is the text a tag carries after its name, with the comment's
// line-leading asterisks and the surrounding space removed. Empty for a bare tag.
func correctnessNoCallerDataMutationTagReason(file *ast.SourceFile, tag *ast.Node) string {
	tagSource := file.Text()[tag.Pos():tag.End()]
	_, reason, _ := strings.Cut(tagSource, "@"+tag.TagName().Text())
	reason = strings.TrimSuffix(text.TrimWhitespace(reason), "*/")
	var words []string
	for _, line := range strings.Split(reason, "\n") {
		line = text.TrimWhitespace(strings.TrimLeft(text.TrimWhitespace(line), "*"))
		if line != "" {
			words = append(words, line)
		}
	}
	return strings.Join(words, " ")
}

// correctnessNoCallerDataMutationMutatesTagParts splits a `@mutates` tag into the parameter it names,
// its first word, and the reason after it.
func correctnessNoCallerDataMutationMutatesTagParts(file *ast.SourceFile, tag *ast.Node) (string, string) {
	parameter, reason, _ := strings.Cut(correctnessNoCallerDataMutationTagReason(file, tag), " ")
	return parameter, text.TrimWhitespace(reason)
}

// correctnessNoCallerDataMutationParameterIndex is the position of the parameter a function names
// plainly, or -1.
func correctnessNoCallerDataMutationParameterIndex(function *ast.Node, name string) int {
	for index, parameter := range function.Parameters() {
		if declared := parameter.Name(); declared != nil && declared.Kind == ast.KindIdentifier && declared.Text() == name {
			return index
		}
	}
	return -1
}

// correctnessNoCallerDataMutationOutParameterOf says whether a parameter is an out-parameter by
// contract: its method, or a method its method overrides or implements, carries `@mutates` naming the
// parameter at its position, with a reason.
//
// It answers with the note key the exemption is counted under, the tagged method and parameter and the
// file declaring them, so --coverage can say which contract excused how many writes.
func correctnessNoCallerDataMutationOutParameterOf(ctx rule.Context, parameter *ast.Symbol) (string, bool) {
	declaration := correctnessNoCallerDataMutationParameterOf(ctx, parameter)
	if declaration == nil || declaration.Parent == nil || declaration.Parent.Kind != ast.KindMethodDeclaration {
		return "", false
	}
	method := declaration.Parent
	position := -1
	for index, candidate := range method.Parameters() {
		if candidate == declaration {
			position = index
		}
	}
	for _, contract := range append([]*ast.Node{method}, correctnessNoCallerDataMutationOverriddenMethods(ctx, method)...) {
		file := ast.GetSourceFileOfNode(contract)
		for _, tag := range correctnessNoCallerDataMutationTagsOf(file, contract, correctnessNoCallerDataMutationMutatesTag) {
			named, reason := correctnessNoCallerDataMutationMutatesTagParts(file, tag)
			if reason == "" || correctnessNoCallerDataMutationParameterIndex(contract, named) != position {
				continue
			}
			owner := ""
			if contract.Parent != nil && contract.Parent.Name() != nil {
				owner = contract.Parent.Name().Text() + "."
			}
			return owner + contract.Name().Text() + "(" + named + ") in " + file.FileName().AsString(), true
		}
	}
	return "", false
}

// correctnessNoCallerDataMutationOverriddenMethods returns the method declarations and signatures a
// class method overrides or implements, at any depth.
//
// Each class or interface is asked for its base types (the class it extends, the interfaces it
// extends) and its implemented interfaces, and each base for a member of the method's name. That
// member's declarations are what the method overrides, and their own containers are asked in turn, so
// an override of an override still reaches the tagged declaration.
func correctnessNoCallerDataMutationOverriddenMethods(ctx rule.Context, method *ast.Node) []*ast.Node {
	name := method.Name()
	if name == nil || (name.Kind != ast.KindIdentifier && name.Kind != ast.KindStringLiteral) {
		return nil
	}
	var overridden []*ast.Node
	seen := map[*ast.Node]bool{}
	queue := []*ast.Node{method.Parent}
	for len(queue) > 0 {
		container := queue[0]
		queue = queue[1:]
		if container == nil || seen[container] || container.Symbol() == nil {
			continue
		}
		seen[container] = true
		var bases []*checker.Type
		if declared := checker.Checker_getDeclaredTypeOfSymbol(ctx.TypeChecker, container.Symbol()); declared != nil &&
			checker.Type_objectFlags(declared)&(checker.ObjectFlagsInterface|checker.ObjectFlagsClass) != 0 {
			bases = checker.Checker_getBaseTypes(ctx.TypeChecker, declared)
		}
		for _, implemented := range ast.GetImplementsHeritageClauseElements(container) {
			if implementedType := ctx.TypeChecker.GetTypeAtLocation(implemented.AsNode()); implementedType != nil {
				bases = append(bases, implementedType)
			}
		}
		for _, base := range bases {
			member := checker.Checker_getPropertyOfType(ctx.TypeChecker, base, name.Text())
			if member == nil {
				continue
			}
			for _, declaration := range member.Declarations {
				if declaration.Kind != ast.KindMethodDeclaration && declaration.Kind != ast.KindMethodSignature {
					continue
				}
				overridden = append(overridden, declaration)
				queue = append(queue, declaration.Parent)
			}
		}
	}
	return overridden
}
