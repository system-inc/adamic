package typescript

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// StrictVoidReturnOptions is the rule's one option.
type StrictVoidReturnOptions struct {
	// AllowReturnAny permits a function returning `any` where a void function is expected.
	//
	// Defaults to FALSE. The generic decoder would be correct by coincidence here; the hand-rolled
	// one keeps an absent key distinguishable from an explicit false.
	AllowReturnAny bool
}

// DefaultStrictVoidReturnSettings is upstream's `defaultOptions`.
func DefaultStrictVoidReturnSettings() StrictVoidReturnOptions {
	return StrictVoidReturnOptions{AllowReturnAny: false}
}

type strictVoidReturnRawOptions struct {
	AllowReturnAny *bool `json:"allowReturnAny"`
}

// DecodeStrictVoidReturnOptions reads the rule's configuration.
//
// cohere's config layer strips ESLint's `[severity, options]` tuple, so the bare object arrives.
func DecodeStrictVoidReturnOptions(raw []byte) (any, error) {
	decoded, err := rule.DecodeOptionsInto[strictVoidReturnRawOptions]()(raw)
	if err != nil {
		return DefaultStrictVoidReturnSettings(), err
	}
	wire, _ := decoded.(strictVoidReturnRawOptions)
	options := DefaultStrictVoidReturnSettings()
	if wire.AllowReturnAny != nil {
		options.AllowReturnAny = *wire.AllowReturnAny
	}
	return options, nil
}

// StrictVoidReturn flags a value-returning function passed where a void function is expected.
//
//	valid:   declare function takes(cb: () => void): void; takes(() => {});
//	valid:   declare function takes(cb: () => number): void; takes(() => 1);
//	invalid: declare function takes(cb: () => void): void; takes(() => 1);
//	invalid: declare function takes(cb: () => void): void; takes(async () => {});
//
// A callback typed `() => void` is a promise that its result is ignored. Passing something that
// returns a value there is usually a mistake with consequences the type system will not raise: an
// async callback passed to `addEventListener` produces a floating promise whose rejection nobody
// handles, and a value-returning callback passed to `forEach` reads as though the value matters.
//
// # The whole rule turns on the CONTEXTUAL type
//
// Every listener finds a position where a value is being placed, asks what type is expected there,
// and reports when that expectation is a void-returning function while the value supplied is not.
// That expectation comes from the checker rather than from syntax, which is why this needs types.
//
// `getContextualType` is the load-bearing call and it is reachable here, which was established with
// a compiling probe before this rule was written rather than assumed from a grep. The shim spells
// checker methods `Checker_getX`, so searching for `GetContextualType` finds nothing and reports a
// false absence; the control on that same search also returned zero, which is the only reason the
// absence was not recorded as real.
//
// # Three findings, and they point at three different things
//
//	nonVoidFunc     the function is not a literal, or is a generator, or has a non-void return
//	                annotation. Points at the function, or at the ANNOTATION when there is one.
//	asyncFunc       an async function literal. Points at the function HEAD rather than the whole
//	                function, which is a narrower span than the node.
//	nonVoidReturn   a specific `return` statement inside an otherwise-acceptable function. Points at
//	                the `return` KEYWORD, not the statement and not the returned expression.
//
// Those spans are upstream's and each is measured. A port reporting all three on the node would
// satisfy every message-id assertion while pointing a reader at the wrong token three ways.
//
// # What counts as already-void
//
// `void`, `never` and `undefined` are all acceptable return types, and `any` joins them only when
// `allowReturnAny` is set. The check is over every constituent of every call signature's return
// type, so a union reporting requires every member to be acceptable.
//
// # Cost
//
// Twelve anchors, most of them common, but each exits on a cheap syntactic test before the
// contextual type is requested. The checker is consulted only for an expression sitting in a
// position that could expect a function.
var StrictVoidReturn = rule.Rule{
	Name: "@typescript-eslint/strict-void-return",

	// The contextual type decides every finding.
	NeedsTypeChecker: true,
	TypeReach:        rule.TypeReachShapes,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		settings, isSettings := rule.OptionsAs[StrictVoidReturnOptions](options)
		if !isSettings {
			settings = DefaultStrictVoidReturnSettings()
		}

		state := &strictVoidReturnState{ctx: ctx, settings: settings}

		return rule.Listeners{
			ast.KindCallExpression: state.checkCall,
			ast.KindNewExpression:  state.checkCall,

			ast.KindArrayLiteralExpression: func(node *ast.Node) {
				for _, element := range node.AsArrayLiteralExpression().Elements.Nodes {
					if element.Kind != ast.KindSpreadElement {
						state.checkExpression(element)
					}
				}
			},

			ast.KindArrowFunction: func(node *ast.Node) {
				// A concise arrow body is an expression in a position that may expect a function.
				body := node.AsArrowFunction().Body
				if body != nil && body.Kind != ast.KindBlock {
					state.checkExpression(body)
				}
			},

			ast.KindBinaryExpression: func(node *ast.Node) {
				binary := node.AsBinaryExpression()
				if binary.OperatorToken == nil {
					return
				}
				// Upstream checks every AssignmentExpression and notes that arithmetic compounds
				// like `+=` fall out on their own, because their contextual type is never a
				// function. The LOGICAL compounds do not fall out: `foo ??= cb` places `cb` in a
				// position typed by `foo`, so a void-returning `foo` makes it this rule's question.
				// Measured, three corpus cases turn on exactly that and were silent when this arm
				// accepted only a plain `=`.
				switch binary.OperatorToken.Kind {
				case ast.KindEqualsToken,
					ast.KindQuestionQuestionEqualsToken,
					ast.KindBarBarEqualsToken,
					ast.KindAmpersandAmpersandEqualsToken:
					state.checkExpression(binary.Right)
				}
			},

			ast.KindObjectLiteralExpression: func(node *ast.Node) {
				for _, property := range node.AsObjectLiteralExpression().Properties.Nodes {
					switch property.Kind {
					case ast.KindPropertyAssignment:
						state.checkExpression(property.AsPropertyAssignment().Initializer)
					case ast.KindMethodDeclaration:
						// A shorthand method is a function in a property position, so it is checked
						// like one. `{ cb(n) { return n; } }` against a `(n: number) => void`
						// property reports, and it is not reachable through the assignment arm
						// because there is no initializer expression to check.
						state.reportIfNonVoidFunctionAtProperty(property)
					}
				}
			},

			ast.KindReturnStatement: func(node *ast.Node) {
				if argument := node.AsReturnStatement().Expression; argument != nil {
					state.checkExpression(argument)
				}
			},

			ast.KindVariableDeclaration: func(node *ast.Node) {
				if initializer := node.AsVariableDeclaration().Initializer; initializer != nil {
					state.checkExpression(initializer)
				}
			},

			ast.KindPropertyDeclaration: func(node *ast.Node) {
				// A class property is checked against the same-named member of every base type
				// FIRST, because a subclass narrowing an inherited void callback to a value-
				// returning one is the mistake this arm exists for. Only if no base expects a void
				// function does the ordinary contextual-type path apply.
				if state.checkAgainstBaseTypes(node) {
					return
				}
				if initializer := node.AsPropertyDeclaration().Initializer; initializer != nil {
					state.checkExpression(initializer)
				}
			},

			ast.KindMethodDeclaration: func(node *ast.Node) {
				// A method with no body is an overload signature and declares nothing to check.
				if node.AsMethodDeclaration().Body == nil {
					return
				}
				state.checkAgainstBaseTypes(node)
			},

			ast.KindGetAccessor: func(node *ast.Node) {
				if node.AsGetAccessorDeclaration().Body == nil {
					return
				}
				state.checkAgainstBaseTypes(node)
			},

			ast.KindJsxAttribute: func(node *ast.Node) {
				initializer := node.AsJsxAttribute().Initializer
				if initializer == nil || initializer.Kind != ast.KindJsxExpression {
					return
				}
				// The attribute's expected type resolves the element's props, for every attribute,
				// and a value that is not a non-void function is silent whatever was expected, so
				// the value is asked first (#hekjpw3).
				if inner := initializer.AsJsxExpression().Expression; inner != nil && state.isNonVoidFunctionValue(inner) {
					state.checkExpression(inner)
				}
			},
		}
	},
}

type strictVoidReturnState struct {
	ctx      rule.Context
	settings StrictVoidReturnOptions
}

// isVoidReturningFunctionType answers upstream's helper of the same name.
//
// Every call signature's return type is split into union constituents, and the whole thing counts
// as void-returning only when there is at least one constituent and every one of them is `void`.
// The emptiness test matters: a type with no call signatures is not a void-returning function, and
// without it every non-function contextual type would qualify vacuously.
func (s *strictVoidReturnState) isVoidReturningFunctionType(subject *checker.Type) bool {
	if subject == nil {
		return false
	}
	// The contextual type is split into union constituents BEFORE signatures are read, and leaving
	// that out was a real defect rather than a nicety. `let foo: (() => void) | null` gives a
	// contextual type of `(() => void) | null`, and asking a union for its own call signatures
	// returns none, so the whole rule went silent on every optional or nullable callback position.
	// Measured: three corpus cases turn on exactly that shape.
	//
	// `tsutils.getCallSignaturesOfType`, which upstream calls, does this splitting itself. The shim
	// exposes the raw checker method, which does not.
	found := 0
	for part := range type_checking.UnionTypePartsSeq(subject) {
		for _, signature := range checker.Checker_getSignaturesOfType(s.ctx.TypeChecker, part, checker.SignatureKindCall) {
			returnType := checker.Checker_getReturnTypeOfSignature(s.ctx.TypeChecker, signature)
			for constituent := range type_checking.UnionTypePartsSeq(returnType) {
				found++
				if !type_checking.IsTypeFlagSet(constituent, checker.TypeFlagsVoid) {
					return false
				}
			}
		}
	}
	return found > 0
}

// checkAgainstBaseTypes answers upstream's `getBaseTypesOfClassMember` loop.
//
// A class member that overrides or implements one declared elsewhere is judged against what the
// BASE declared rather than against a contextual type, because there is no contextual position: the
// member simply sits in a class body. `class Bar extends Foo { cb = Math.random }` reports when
// `Foo.cb` is typed `() => void`.
//
// Upstream reports at most one finding per member however many bases match, and that is reproduced:
// two base types both expecting void would otherwise report the same member twice.
//
// Returns whether a base was found expecting a void function, so the caller can skip the ordinary
// contextual path rather than reporting the same member through two routes.
func (s *strictVoidReturnState) checkAgainstBaseTypes(member *ast.Node) bool {
	if s.ctx.TypeChecker == nil {
		return false
	}
	name := member.Name()
	if name == nil {
		return false
	}
	memberSymbol := s.ctx.TypeChecker.GetSymbolAtLocation(name)
	if memberSymbol == nil {
		return false
	}

	class := member.Parent
	if class == nil {
		return false
	}
	heritage := strictVoidReturnHeritageClauses(class)
	if heritage == nil {
		return false
	}

	for _, clause := range heritage.Nodes {
		types := clause.AsHeritageClause().Types
		if types == nil {
			continue
		}
		for _, baseTypeNode := range types.Nodes {
			baseType := s.ctx.TypeChecker.GetTypeAtLocation(baseTypeNode)
			if baseType == nil {
				continue
			}
			baseMemberSymbol := checker.Checker_getPropertyOfType(s.ctx.TypeChecker, baseType, memberSymbol.Name)
			if baseMemberSymbol == nil {
				continue
			}
			baseMemberType := s.ctx.TypeChecker.GetTypeOfSymbolAtLocation(baseMemberSymbol, member)
			if baseMemberType == nil || !s.isVoidReturningFunctionType(baseMemberType) {
				continue
			}
			// The value judged is the member's own function: a property's initializer, or the
			// method itself.
			subject := member
			if member.Kind == ast.KindPropertyDeclaration {
				initializer := member.AsPropertyDeclaration().Initializer
				if initializer == nil {
					return true
				}
				subject = initializer
			}
			s.reportIfNonVoidFunction(subject)
			// At most one finding per member, however many bases match.
			return true
		}
	}
	return false
}

// strictVoidReturnPropertyName reads a property key that names something a type can carry.
//
// A numeric key counts: `{ 1234(n) { return n } }` against a `{ 1234: (n: number) => void }` type
// reports, and the property symbol is spelled with the digits.
func strictVoidReturnPropertyName(name *ast.Node) (string, bool) {
	switch name.Kind {
	case ast.KindIdentifier, ast.KindStringLiteral, ast.KindNumericLiteral,
		ast.KindNoSubstitutionTemplateLiteral:
		return name.Text(), true
	}
	return "", false
}

// strictVoidReturnHeritageClauses reads the heritage clauses of a class-like declaration.
func strictVoidReturnHeritageClauses(node *ast.Node) *ast.NodeList {
	switch node.Kind {
	case ast.KindClassDeclaration:
		return node.AsClassDeclaration().HeritageClauses
	case ast.KindClassExpression:
		return node.AsClassExpression().HeritageClauses
	}
	return nil
}

// reportIfNonVoidFunctionAtProperty checks a shorthand method against its property's expected type.
//
// A method has no initializer expression, so `getContextualType` has nothing to ask about. The
// expected type is read from the property NAME instead, which the checker resolves against the
// contextual object type, and the method itself is then judged exactly as a function literal would
// be. Upstream reaches the same place through its own object-property path.
func (s *strictVoidReturnState) reportIfNonVoidFunctionAtProperty(property *ast.Node) {
	if s.ctx.TypeChecker == nil {
		return
	}
	name := property.Name()
	if name == nil {
		return
	}
	// The property's expected type comes from the enclosing OBJECT's contextual type, looked up by
	// name. Asking the method's own name for a contextual type answers nil, measured, because a
	// method name is a declaration rather than a value position.
	object := property.Parent
	if object == nil || object.Kind != ast.KindObjectLiteralExpression {
		return
	}
	objectExpected := checker.Checker_getContextualType(s.ctx.TypeChecker, object, 0)
	if objectExpected == nil {
		return
	}
	propertyName, readable := strictVoidReturnPropertyName(name)
	if !readable {
		return
	}
	var expected *checker.Type
	for part := range type_checking.UnionTypePartsSeq(objectExpected) {
		symbol := checker.Checker_getPropertyOfType(s.ctx.TypeChecker, part, propertyName)
		if symbol == nil {
			continue
		}
		if candidate := s.ctx.TypeChecker.GetTypeOfSymbolAtLocation(symbol, property); candidate != nil {
			expected = candidate
			break
		}
	}
	if expected == nil || !s.isVoidReturningFunctionType(expected) {
		return
	}
	s.reportIfNonVoidFunction(property)
}

// checkExpression is upstream's `checkExpressionNode`, and reports whether the position expected a
// void function.
func (s *strictVoidReturnState) checkExpression(node *ast.Node) bool {
	if node == nil || s.ctx.TypeChecker == nil {
		return false
	}
	expected := checker.Checker_getContextualType(s.ctx.TypeChecker, node, 0)
	if expected == nil || !s.isVoidReturningFunctionType(expected) {
		return false
	}
	s.reportIfNonVoidFunction(node)
	return true
}

// allowedReturnFlags is the set of return types a void position tolerates.
func (s *strictVoidReturnState) allowedReturnFlags() checker.TypeFlags {
	allowed := checker.TypeFlagsVoid | checker.TypeFlagsNever | checker.TypeFlagsUndefined
	if s.settings.AllowReturnAny {
		allowed |= checker.TypeFlagsAny
	}
	return allowed
}

// isNonVoidFunctionValue answers whether the value is a function with a return this rule does not
// accept, from the value's own type alone. It is the first step of reportIfNonVoidFunction, which
// is silent whenever this is false, so a position whose expected type is costly to resolve asks it
// before asking the position (#hekjpw3).
func (s *strictVoidReturnState) isNonVoidFunctionValue(node *ast.Node) bool {
	allowed := s.allowedReturnFlags()

	actual := checker.Checker_getApparentType(s.ctx.TypeChecker, s.ctx.TypeChecker.GetTypeAtLocation(node))
	if actual == nil {
		return false
	}

	// Already void: every constituent of every call signature's return type is acceptable.
	acceptable := true
	sawOne := false
	for part := range type_checking.UnionTypePartsSeq(actual) {
		for _, signature := range checker.Checker_getSignaturesOfType(s.ctx.TypeChecker, part, checker.SignatureKindCall) {
			returnType := checker.Checker_getReturnTypeOfSignature(s.ctx.TypeChecker, signature)
			for constituent := range type_checking.UnionTypePartsSeq(returnType) {
				sawOne = true
				if !type_checking.IsTypeFlagSet(constituent, allowed) {
					acceptable = false
				}
			}
		}
	}
	// No call signatures means nothing to complain about here, and upstream's `every` over an
	// empty list is vacuously true, so both land on the same silence.
	return sawOne && !acceptable
}

// reportIfNonVoidFunction is the decision tree, and its order is upstream's.
func (s *strictVoidReturnState) reportIfNonVoidFunction(node *ast.Node) {
	if !s.isNonVoidFunctionValue(node) {
		return
	}
	allowed := s.allowedReturnFlags()

	isArrow := node.Kind == ast.KindArrowFunction
	isFunctionExpression := node.Kind == ast.KindFunctionExpression
	// A shorthand method reaches here from the object-property arm and is a function literal for
	// every purpose below: it can be async, it can carry a return annotation, and it has a block
	// body whose returns are its own.
	isMethod := node.Kind == ast.KindMethodDeclaration || node.Kind == ast.KindGetAccessor
	if !isArrow && !isFunctionExpression && !isMethod {
		// Not a function literal, so there is nothing to point inside. Upstream reports the node.
		s.ctx.ReportNode(node, strictVoidReturnNonVoidFuncMessage())
		return
	}

	if strictVoidReturnIsGenerator(node) {
		s.ctx.ReportRange(strictVoidReturnHeadRange(s.ctx, node), strictVoidReturnNonVoidFuncMessage())
		return
	}

	if strictVoidReturnIsAsync(node) {
		s.ctx.ReportRange(strictVoidReturnHeadRange(s.ctx, node), strictVoidReturnAsyncFuncMessage())
		return
	}

	// A sync function literal.

	// A concise arrow body is its own return, and it reports on the BODY rather than on the
	// function or on a `return` keyword there is none of. This branch sits before the annotation
	// check, which matters: a shorthand arrow with a non-void annotation reports here rather than
	// at the annotation. Missing it cost 54 of 107 reporting cases on the first writing, all of
	// them this shape, because the block-body walk below has nothing to walk.
	//
	// The parentheses are unwrapped for the span. ESTree has no parenthesis node, so upstream's body
	// in `(chunk) => (total += chunk)` is the assignment and its finding starts at `total`; reporting
	// our `ParenthesizedExpression` pointed one column early on eight ahra sites.
	if body := strictVoidReturnBody(node); body != nil && body.Kind != ast.KindBlock {
		s.ctx.ReportNode(ast.SkipParentheses(body), strictVoidReturnNonVoidReturnMessage())
		return
	}

	// An explicit non-void return annotation is reported at the ANNOTATION rather than at the
	// function.
	if annotation := strictVoidReturnTypeAnnotation(node); annotation != nil {
		if annotation.Kind != ast.KindVoidKeyword {
			s.ctx.ReportNode(annotation, strictVoidReturnNonVoidFuncMessage())
			return
		}
	}

	// Otherwise every offending return statement is reported, at its `return` keyword.
	body := strictVoidReturnBody(node)
	if body == nil || body.Kind != ast.KindBlock {
		return
	}
	for _, statement := range strictVoidReturnWalkStatements(body) {
		argument := statement.AsReturnStatement().Expression
		if argument == nil {
			continue
		}
		returnType := s.ctx.TypeChecker.GetTypeAtLocation(argument)
		if type_checking.IsTypeFlagSet(returnType, allowed) {
			continue
		}
		s.ctx.ReportRange(scanner.GetRangeOfTokenAtPosition(s.ctx.SourceFile, statement.Pos()),
			strictVoidReturnNonVoidReturnMessage())
	}
}

// checkCall is upstream's `checkFunctionCallNode`.
//
// It does not simply ask the contextual type of each argument, and the reason is overloads:
// `getContextualType` resolves to the FIRST overload's parameter type even when another overload is
// the one that matches, so a call like `addEventListener` whose overloads differ in callback return
// type would be judged against the wrong one. Upstream therefore collects the expected return type
// from every signature and only trusts the contextual type when there is a single signature or when
// all of them agree that the position is void.
func (s *strictVoidReturnState) checkCall(node *ast.Node) {
	if s.ctx.TypeChecker == nil {
		return
	}

	var arguments *ast.NodeList
	var callee *ast.Node
	signatureKind := checker.SignatureKindCall
	if node.Kind == ast.KindCallExpression {
		call := node.AsCallExpression()
		arguments, callee = call.Arguments, call.Expression
	} else {
		newExpression := node.AsNewExpression()
		arguments, callee = newExpression.Arguments, newExpression.Expression
		signatureKind = checker.SignatureKindConstruct
	}
	if arguments == nil || len(arguments.Nodes) == 0 || callee == nil {
		return
	}

	calleeType := s.ctx.TypeChecker.GetTypeAtLocation(callee)
	if calleeType == nil {
		return
	}
	signatures := []*checker.Signature{}
	for constituent := range type_checking.UnionTypePartsSeq(calleeType) {
		signatures = append(signatures,
			checker.Checker_getSignaturesOfType(s.ctx.TypeChecker, constituent, signatureKind)...)
	}

	for index, argument := range arguments.Nodes {
		if argument.Kind == ast.KindSpreadElement {
			continue
		}

		expectedReturns := []*checker.Type{}
		for _, signature := range signatures {
			parameters := checker.Signature_parameters(signature)
			if index >= len(parameters) {
				continue
			}
			parameterType := s.ctx.TypeChecker.GetTypeOfSymbolAtLocation(parameters[index], callee)
			if parameterType == nil {
				continue
			}
			for constituent := range type_checking.UnionTypePartsSeq(parameterType) {
				for _, parameterSignature := range checker.Checker_getSignaturesOfType(
					s.ctx.TypeChecker, constituent, checker.SignatureKindCall) {
					expectedReturns = append(expectedReturns,
						checker.Checker_getReturnTypeOfSignature(s.ctx.TypeChecker, parameterSignature))
				}
			}
		}

		allReturnVoid := true
		for _, returnType := range expectedReturns {
			// A type parameter is treated like void, because `getTypeOfSymbolAtLocation` hands back
			// an unresolved `T` even for overloads that do match, and there is no way to tell which
			// generic overload applies.
			if !strictVoidReturnIsVoid(returnType) &&
				!strictVoidReturnIsNullishOrAny(returnType) &&
				!type_checking.IsTypeParameter(returnType) {
				allReturnVoid = false
				break
			}
		}

		if (len(signatures) == 1 || allReturnVoid) && s.checkExpression(argument) {
			continue
		}

		anyVoid := false
		everyNullishOrAny := true
		for _, returnType := range expectedReturns {
			if strictVoidReturnIsVoid(returnType) {
				anyVoid = true
			}
			if !strictVoidReturnIsNullishOrAny(returnType) {
				everyNullishOrAny = false
			}
		}
		if anyVoid && everyNullishOrAny {
			s.reportIfNonVoidFunction(argument)
		}
	}
}

func strictVoidReturnIsVoid(subject *checker.Type) bool {
	return type_checking.IsTypeFlagSet(subject, checker.TypeFlagsVoid)
}

func strictVoidReturnIsNullishOrAny(subject *checker.Type) bool {
	return type_checking.IsTypeFlagSet(subject,
		checker.TypeFlagsVoidLike|checker.TypeFlagsUndefined|checker.TypeFlagsNull|
			checker.TypeFlagsAny|checker.TypeFlagsNever)
}

func strictVoidReturnIsGenerator(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindFunctionExpression:
		return node.AsFunctionExpression().AsteriskToken != nil
	case ast.KindMethodDeclaration:
		return node.AsMethodDeclaration().AsteriskToken != nil
	}
	return false
}

func strictVoidReturnIsAsync(node *ast.Node) bool {
	modifiers := node.Modifiers()
	if modifiers == nil {
		return false
	}
	for _, modifier := range modifiers.Nodes {
		if modifier.Kind == ast.KindAsyncKeyword {
			return true
		}
	}
	return false
}

func strictVoidReturnTypeAnnotation(node *ast.Node) *ast.Node {
	switch node.Kind {
	case ast.KindArrowFunction:
		return node.AsArrowFunction().Type
	case ast.KindFunctionExpression:
		return node.AsFunctionExpression().Type
	case ast.KindMethodDeclaration:
		return node.AsMethodDeclaration().Type
	case ast.KindGetAccessor:
		return node.AsGetAccessorDeclaration().Type
	}
	return nil
}

func strictVoidReturnBody(node *ast.Node) *ast.Node {
	switch node.Kind {
	case ast.KindArrowFunction:
		return node.AsArrowFunction().Body
	case ast.KindFunctionExpression:
		return node.AsFunctionExpression().Body
	case ast.KindMethodDeclaration:
		return node.AsMethodDeclaration().Body
	case ast.KindGetAccessor:
		return node.AsGetAccessorDeclaration().Body
	}
	return nil
}

// strictVoidReturnHeadRange is upstream's `getFunctionHeadLoc`.
//
// The reported span for an async or generator function literal is its HEAD rather than the whole
// function, so a multi-line callback underlines its first line instead of its body. The head runs
// from the function's first token through the closing parenthesis of its parameter list.
func strictVoidReturnHeadRange(ctx rule.Context, node *ast.Node) core.TextRange {
	nodeRange := rule.TokenRange(ctx.SourceFile, node)
	body := strictVoidReturnBody(node)
	if body == nil {
		return nodeRange
	}
	// Everything up to the body, with trailing trivia and any arrow trimmed off.
	end := rule.TokenRange(ctx.SourceFile, body).Pos()
	text := ctx.SourceFile.Text()
	for end > nodeRange.Pos() {
		trimmed := end - 1
		switch text[trimmed] {
		case ' ', '\t', '\r', '\n', '>', '=':
			end = trimmed
			continue
		}
		break
	}
	if end <= nodeRange.Pos() {
		return nodeRange
	}
	return core.NewTextRange(nodeRange.Pos(), end)
}

// strictVoidReturnWalkStatements collects the return statements belonging to one function body.
//
// It descends through ordinary statement nesting but never into a nested function, whose returns
// are its own. Upstream's `walkStatements` makes the same distinction.
func strictVoidReturnWalkStatements(body *ast.Node) []*ast.Node {
	found := []*ast.Node{}
	var walk func(*ast.Node)
	walk = func(node *ast.Node) {
		if node == nil {
			return
		}
		switch node.Kind {
		case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction,
			ast.KindClassDeclaration, ast.KindClassExpression:
			return
		case ast.KindReturnStatement:
			found = append(found, node)
			return
		}
		node.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return false
		})
	}
	walk(body)
	return found
}

func strictVoidReturnNonVoidFuncMessage() rule.Message {
	return rule.Message{
		Id:          "nonVoidFunc",
		Description: "Value-returning function used in a context where a void function is expected.",
	}
}

func strictVoidReturnAsyncFuncMessage() rule.Message {
	return rule.Message{
		Id:          "asyncFunc",
		Description: "Async function used in a context where a void function is expected.",
	}
}

func strictVoidReturnNonVoidReturnMessage() rule.Message {
	return rule.Message{
		Id:          "nonVoidReturn",
		Description: "Value returned in a context where a void return is expected.",
	}
}
