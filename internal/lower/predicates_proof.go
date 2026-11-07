package lower

import (
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// A tag proof certifies membership in the tag set only. Admission must preserve
// a checked view for the target's remaining fields; this is not a shape proof.
type predicateProof struct{ TaggedView bool }

type predicateTruth uint8

const (
	predicateTrue   predicateTruth = 1
	predicateFalse  predicateTruth = 2
	predicateEither predicateTruth = 3
)

type predicateVerifier struct {
	l          *lowering
	field      string
	truthiness bool
	cells      []string
	summaries  map[*ast.Node][]predicateTruth
}

type predicatePath struct {
	cell   int
	locals map[*ast.Symbol]predicateTruth
}

func predicateFailure(l *lowering, node *ast.Node, path string) *Refused {
	where := ""
	if node != nil {
		where = l.program.Where(node)
	}
	return &Refused{
		Where: where,
		What:  "an unproven type predicate",
		Fix:   path + "; return a boolean and narrow at the caller (adamic/no-type-predicate)",
	}
}

// A comparison on a tag that can lie proves nothing about the layout (#drwm5pj), so neither prover
// takes it as proof of a member type.
func predicateTagHazardReason(field string, hazard enumTagHazard) string {
	if hazard == enumTagSharedValueHazard {
		return "the comparison of " +
			field +
			" proves nothing about the layout, since tags from different numeric enums share a value"
	}
	return "the comparison of " +
		field +
		" proves nothing about the layout, since a member whose " +
		field +
		" is an open numeric enum can hold that value"
}

// provePredicate starts with no helper facts. Only independently verified bodies
// become summaries. Iteration terminates when no new body can be discharged;
// cycles and opaque calls never acquire a summary from their annotations.
func (lowering *lowering) provePredicate(node *ast.Node) (predicateProof, error) {
	proof := predicateProof{}
	if node == nil ||
		node.Kind != ast.KindTypePredicate ||
		node.Parent == nil ||
		!ast.IsFunctionLike(node.Parent) ||
		node.Parent.Body() == nil {
		return proof, predicateFailure(lowering, node, "no implementation body proves a return path")
	}
	if node.AsTypePredicateNode().Type == nil {
		if node.AsTypePredicateNode().AssertsModifier != nil {
			return lowering.proveConditionAssertion(node)
		}
		return proof, predicateFailure(lowering, node, "an assertion without a target type is not proved")
	}
	target := lowering.checker.GetTypeAtLocation(node.AsTypePredicateNode().Type)
	field := ""
	if target.Flags()&(checker.TypeFlagsObject|checker.TypeFlagsUnion) != 0 &&
		(&predicateVerifier{l: lowering}).kindTarget(target) {
		field = "kind"
		proof.TaggedView = true
	} else if !predicatePrimitiveTarget(target) {
		return proof, predicateFailure(lowering, node, "the target has an unsupported runtime contract")
	}
	point := lowering.predicateFixedPoint(field)
	if _, ok := point.summaries[node.Parent]; ok {
		return proof, nil
	}
	if failure, ok := point.failures[node.Parent]; ok {
		return predicateProof{}, predicateFailureCopy(failure)
	}
	return predicateProof{}, predicateFailure(lowering, node, "unconverged helper proof")
}

// predicateProofs is one lowering's proof of every type predicate in the program: the declarations with a predicate
// return and a body, in the program's order, and for each verifier mode ("" for the eight typeof cells, "kind" for a
// tag), that mode's fixed point. provePredicate used to walk the program and run a fixed point over all of its
// predicates on every call, 43% of the meter's latent step on tsc's 79 files (#pv935yw). The program and the checker
// are every input to it (the verifier reads the checker, fieldLiteral, symbol, isPanicCall and enumTagHazardOf, never
// typeMapper or other lowering state), so nothing invalidates it. It lives on one lowering, never across lowerings.
type predicateProofs struct {
	declarations []*ast.Node
	modes        map[string]*predicateFixedPoint
	// brandedViews is each branded supertype's view, registered once per source with the error that
	// stopped it, brandedRequests each instance request's verdict by the instance's key, brandedClasses the
	// declared classes assignable to it, classes the classes they're found among, and files the program's
	// files they're declared in (predicates_branded.go).
	brandedViews    map[brandedViewKey]error
	brandedRequests map[string]error
	brandedClasses  map[*checker.Type][]*checker.Type
	classes         []*checker.Type
	files           []*ast.SourceFile
}

// predicateFixedPoint is one mode's verifier run to its end: every declaration's summary, or the error its verify gave
// in the last pass, the pass that proved nothing new. A summary is never dropped and the passes run in the program's
// order, so a summary that provePredicate's early return saw is here too, and a declaration never proved has the
// error from the same final pass it always had.
type predicateFixedPoint struct {
	summaries map[*ast.Node][]predicateTruth
	failures  map[*ast.Node]error
}

// predicateFixedPoint returns mode field's fixed point, running it the first time it's asked.
func (lowering *lowering) predicateFixedPoint(field string) *predicateFixedPoint {
	lowering.proofs()
	if point, ok := lowering.predicateProofs.modes[field]; ok {
		return point
	}
	return lowering.runPredicateFixedPoint(field)
}

// proofs is the lowering's predicateProofs, its declarations found the first time it's asked.
func (lowering *lowering) proofs() *predicateProofs {
	if lowering.predicateProofs == nil {
		proofs := &predicateProofs{
			modes:           map[string]*predicateFixedPoint{},
			brandedViews:    map[brandedViewKey]error{},
			brandedRequests: map[string]error{},
			brandedClasses:  map[*checker.Type][]*checker.Type{},
		}
		var visit ast.Visitor
		visit = func(n *ast.Node) bool {
			if ast.IsFunctionLike(n) && n.Type() != nil && n.Type().Kind == ast.KindTypePredicate && n.Body() != nil {
				proofs.declarations = append(proofs.declarations, n)
			}
			n.ForEachChild(visit)
			return false
		}
		for _, file := range lowering.program.Files() {
			file.AsNode().ForEachChild(visit)
		}
		lowering.predicateProofs = proofs
	}
	return lowering.predicateProofs
}

// runPredicateFixedPoint runs mode field's verifier to its fixed point and keeps it.
func (lowering *lowering) runPredicateFixedPoint(field string) *predicateFixedPoint {
	v := &predicateVerifier{l: lowering, field: field, summaries: map[*ast.Node][]predicateTruth{}}
	if field == "" {
		v.cells = []string{"string", "number", "boolean", "undefined", "object", "function", "symbol", "bigint"}
	} else {
		// A tag's cells are every kind the program's predicates name, the same set whichever predicate asks.
		v.cells = []string{"<other kind>"}
		for _, declaration := range lowering.predicateProofs.declarations {
			v.collectKinds(declaration)
		}
	}
	point := &predicateFixedPoint{summaries: v.summaries, failures: map[*ast.Node]error{}}
	for {
		changed := false
		for _, declaration := range lowering.predicateProofs.declarations {
			if _, ok := v.summaries[declaration]; ok {
				continue
			}
			summary, verifyError := v.verify(declaration)
			if verifyError == nil {
				v.summaries[declaration] = summary
				delete(point.failures, declaration)
				changed = true
			} else {
				point.failures[declaration] = verifyError
			}
		}
		if !changed {
			break
		}
	}
	lowering.predicateProofs.modes[field] = point
	return point
}

// predicateFailureCopy is a memoized failure for one caller: its own copy of a refusal or a not-yet, so no caller's
// change to it reaches another's.
func predicateFailureCopy(sourceError error) error {
	switch failure := sourceError.(type) {
	case *Refused:
		copied := *failure
		return &copied
	case *NotYet:
		copied := *failure
		return &copied
	}
	return sourceError
}

func predicatePrimitiveTarget(t *checker.Type) bool {
	if t.Flags()&checker.TypeFlagsUnion != 0 {
		for _, m := range t.Types() {
			if !predicatePrimitiveTarget(m) {
				return false
			}
		}
		return true
	}
	// Literal refinements need value partitions, not just typeof partitions.
	return t.Flags()&(checker.TypeFlagsString|checker.TypeFlagsNumber|checker.TypeFlagsBoolean|checker.TypeFlagsUndefined) != 0 &&
		t.Flags()&(checker.TypeFlagsLiteral|checker.TypeFlagsAny|checker.TypeFlagsUnknown|checker.TypeFlagsTypeParameter) == 0
}

func (verifier *predicateVerifier) kindTarget(t *checker.Type) bool {
	if t.Flags()&checker.TypeFlagsUnion != 0 {
		for _, m := range t.Types() {
			if !verifier.kindTarget(m) {
				return false
			}
		}
		return true
	}
	return t.Flags()&checker.TypeFlagsObject != 0 && !isClassInstance(t) && verifier.l.fieldLiteral(t, "kind") != nil
}

func predicateLiteral(t *checker.Type) (string, bool) {
	if t.Flags()&(checker.TypeFlagsStringLiteral|checker.TypeFlagsNumberLiteral|checker.TypeFlagsBooleanLiteral) == 0 {
		return "", false
	}
	return fmt.Sprintf("%T:%v", t.AsLiteralType().Value(), t.AsLiteralType().Value()), true
}

func (verifier *predicateVerifier) addKind(t *checker.Type) {
	key, ok := predicateLiteral(t)
	if !ok {
		return
	}
	for _, cell := range verifier.cells {
		if cell == key {
			return
		}
	}
	verifier.cells = append(verifier.cells, key)
}

func (verifier *predicateVerifier) collectKinds(declaration *ast.Node) {
	var addTarget func(*checker.Type)
	addTarget = func(t *checker.Type) {
		if t.Flags()&checker.TypeFlagsUnion != 0 {
			for _, m := range t.Types() {
				addTarget(m)
			}
			return
		}
		if literal := verifier.l.fieldLiteral(t, "kind"); literal != nil {
			verifier.addKind(literal)
		}
	}
	predicate := declaration.Type().AsTypePredicateNode()
	if predicate.Type != nil {
		addTarget(verifier.l.checker.GetTypeAtLocation(predicate.Type))
	}
	var visit ast.Visitor
	visit = func(n *ast.Node) bool {
		if n.Kind == ast.KindBinaryExpression {
			b := n.AsBinaryExpression()
			if b.OperatorToken.Kind == ast.KindEqualsEqualsEqualsToken ||
				b.OperatorToken.Kind == ast.KindExclamationEqualsEqualsToken {
				verifier.addKind(verifier.l.checker.GetTypeAtLocation(b.Left))
				verifier.addKind(verifier.l.checker.GetTypeAtLocation(b.Right))
			}
		}
		n.ForEachChild(visit)
		return false
	}
	declaration.Body().ForEachChild(visit)
}

func (verifier *predicateVerifier) wanted(t *checker.Type, cell string) (bool, bool) {
	if t.Flags()&checker.TypeFlagsUnion != 0 {
		answer := false
		for _, m := range t.Types() {
			yes, ok := verifier.wanted(m, cell)
			if !ok {
				return false, false
			}
			answer = answer || yes
		}
		return answer, true
	}
	if verifier.field != "" {
		if !verifier.kindTarget(t) {
			return false, false
		}
		key, ok := predicateLiteral(verifier.l.fieldLiteral(t, verifier.field))
		return key == cell, ok
	}
	if !predicatePrimitiveTarget(t) {
		return false, false
	}
	kinds := map[string]checker.TypeFlags{
		"string":    checker.TypeFlagsString,
		"number":    checker.TypeFlagsNumber,
		"boolean":   checker.TypeFlagsBoolean,
		"undefined": checker.TypeFlagsUndefined,
	}
	return t.Flags()&kinds[cell] != 0, true
}

func (verifier *predicateVerifier) parameter(declaration *ast.Node) *ast.Symbol {
	predicate := declaration.Type().AsTypePredicateNode()
	for _, p := range declaration.Parameters() {
		if ast.IsIdentifier(p.Name()) && p.Name().Text() == predicate.ParameterName.Text() {
			return verifier.l.checker.GetSymbolAtLocation(p.Name())
		}
	}
	return nil
}

func (verifier *predicateVerifier) verify(declaration *ast.Node) ([]predicateTruth, error) {
	parameter := verifier.parameter(declaration)
	if parameter == nil || len(declaration.TypeParameters()) != 0 {
		return nil, predicateFailure(verifier.l, declaration, "a generic or non-parameter target has no body proof")
	}
	for _, parameter := range declaration.Parameters() {
		declared := parameter.AsParameterDeclaration()
		if declared.Initializer != nil || declared.DotDotDotToken != nil || !ast.IsIdentifier(parameter.Name()) {
			return nil, predicateFailure(
				verifier.l,
				parameter,
				"parameter initialization may mutate the tested input before this return path",
			)
		}
	}
	predicate := declaration.Type().AsTypePredicateNode()
	if predicate.Type == nil {
		return nil, predicateFailure(verifier.l, declaration, "an assertion without a target type is not proved")
	}
	target := verifier.l.checker.GetTypeAtLocation(predicate.Type)
	summary := make([]predicateTruth, len(verifier.cells))
	for i, cell := range verifier.cells {
		wanted, ok := verifier.wanted(target, cell)
		if !ok {
			return nil, predicateFailure(verifier.l, declaration, "the target has an unsupported runtime contract")
		}
		path := predicatePath{cell: i, locals: map[*ast.Symbol]predicateTruth{}}
		check := func(n *ast.Node, result predicateTruth) error {
			if predicate.AssertsModifier != nil {
				if !wanted {
					return predicateFailure(verifier.l, n, "normal return does not establish the asserted target")
				}
			} else {
				if result&predicateTrue != 0 && !wanted {
					return predicateFailure(verifier.l, n, "true return does not establish the target")
				}
				if result&predicateFalse != 0 && wanted {
					return predicateFailure(verifier.l, n, "false return does not exclude the target")
				}
			}
			summary[i] |= result
			return nil
		}
		body := declaration.Body()
		if body.Kind != ast.KindBlock {
			result, expressionError := verifier.expression(body, parameter, path)
			if expressionError != nil {
				return nil, expressionError
			}
			if checkError := check(body, result); checkError != nil {
				return nil, checkError
			}
			continue
		}
		paths, statementsError := verifier.statements(
			body.AsBlock().Statements.Nodes,
			parameter,
			[]predicatePath{path},
			check,
		)
		if statementsError != nil {
			return nil, statementsError
		}
		if len(paths) > 0 {
			if predicate.AssertsModifier == nil {
				return nil, predicateFailure(verifier.l, body, "normal return has no boolean result")
			}
			if checkError := check(body, predicateTrue); checkError != nil {
				return nil, checkError
			}
		}
	}
	return summary, nil
}

func predicateNot(t predicateTruth) predicateTruth {
	return (t&predicateTrue)<<1 | (t&predicateFalse)>>1
}
func predicateCombine(a, b predicateTruth, and bool) predicateTruth {
	var result predicateTruth
	for _, x := range []predicateTruth{predicateTrue, predicateFalse} {
		for _, y := range []predicateTruth{predicateTrue, predicateFalse} {
			if a&x == 0 || b&y == 0 {
				continue
			}
			yes := x == predicateTrue || y == predicateTrue
			if and {
				yes = x == predicateTrue && y == predicateTrue
			}
			if yes {
				result |= predicateTrue
			} else {
				result |= predicateFalse
			}
		}
	}
	return result
}

func (verifier *predicateVerifier) expression(
	n *ast.Node,
	parameter *ast.Symbol,
	path predicatePath,
) (predicateTruth, error) {
	n = ast.SkipParentheses(n)
	switch n.Kind {
	case ast.KindTrueKeyword:
		return predicateTrue, nil
	case ast.KindFalseKeyword:
		return predicateFalse, nil
	case ast.KindIdentifier:
		if verifier.truthiness && verifier.sameParameter(n, parameter) {
			if verifier.cells[path.cell] == "truthy" {
				return predicateTrue, nil
			}
			return predicateFalse, nil
		}
		if value, ok := path.locals[verifier.l.checker.GetSymbolAtLocation(n)]; ok {
			return value, nil
		}
		return predicateEither, nil
	case ast.KindPrefixUnaryExpression:
		u := n.AsPrefixUnaryExpression()
		if u.Operator == ast.KindExclamationToken {
			t, expressionError := verifier.expression(u.Operand, parameter, path)
			return predicateNot(t), expressionError
		}
	case ast.KindBinaryExpression:
		b := n.AsBinaryExpression()
		if ast.IsAssignmentOperator(b.OperatorToken.Kind) {
			return 0, predicateFailure(verifier.l, n, "mutation invalidates the facts before this return path")
		}
		if b.OperatorToken.Kind == ast.KindAmpersandAmpersandToken || b.OperatorToken.Kind == ast.KindBarBarToken {
			a, expressionError := verifier.expression(b.Left, parameter, path)
			if expressionError != nil {
				return 0, expressionError
			}
			z, expressionError := verifier.expression(b.Right, parameter, path)
			if expressionError != nil {
				return 0, expressionError
			}
			return predicateCombine(a, z, b.OperatorToken.Kind == ast.KindAmpersandAmpersandToken), nil
		}
		if b.OperatorToken.Kind == ast.KindEqualsEqualsEqualsToken ||
			b.OperatorToken.Kind == ast.KindExclamationEqualsEqualsToken {
			for _, pair := range [][2]*ast.Node{{b.Left, b.Right}, {b.Right, b.Left}} {
				test := ast.SkipParentheses(pair[0])
				constant := pair[1]
				matched := false
				yes := false
				if verifier.field == "" &&
					test.Kind == ast.KindTypeOfExpression &&
					verifier.sameParameter(test.AsTypeOfExpression().Expression, parameter) &&
					constant.Kind == ast.KindStringLiteral {
					matched = true
					yes = verifier.cells[path.cell] == constant.Text()
				}
				if verifier.field != "" &&
					test.Kind == ast.KindPropertyAccessExpression &&
					test.Name().Text() == verifier.field &&
					verifier.sameParameter(test.AsPropertyAccessExpression().Expression, parameter) {
					// A cell holds every member with that tag value, so a tag that can lie in the
					// parameter's union doesn't partition it by member.
					if hazard := verifier.l.enumTagHazardOf(
						verifier.l.checker.GetTypeOfSymbol(parameter),
						verifier.field,
					); hazard != enumTagNoHazard {
						reason := predicateTagHazardReason(verifier.field, hazard)
						refused := predicateFailure(verifier.l, n, reason)
						refused.lyingTag = reason
						return 0, refused
					}
					key, ok := predicateLiteral(verifier.l.checker.GetTypeAtLocation(constant))
					if ok && verifier.constant(constant) {
						matched = true
						yes = verifier.cells[path.cell] == key
					}
				}
				if matched {
					if b.OperatorToken.Kind == ast.KindExclamationEqualsEqualsToken {
						yes = !yes
					}
					if yes {
						return predicateTrue, nil
					}
					return predicateFalse, nil
				}
			}
		}
		// Pure comparisons outside the known partition provide no narrowing fact.
		if b.OperatorToken.Kind == ast.KindGreaterThanToken || b.OperatorToken.Kind == ast.KindLessThanToken {
			if verifier.pure(b.Left) && verifier.pure(b.Right) {
				return predicateEither, nil
			}
		}
	case ast.KindCallExpression:
		call := n.AsCallExpression()
		if len(call.Arguments.Nodes) != 1 ||
			!verifier.sameParameter(call.Arguments.Nodes[0], parameter) ||
			!ast.IsIdentifier(call.Expression) {
			break
		}
		symbol := verifier.l.symbol(call.Expression)
		if symbol != nil {
			for _, d := range symbol.Declarations {
				if values, ok := verifier.summaries[d]; ok {
					// This summary partitions the helper's predicate parameter,
					// not an arbitrary argument with the same caller symbol.
					claim := predicateOfSignature(verifier.l.checker, verifier.l.checker.GetSignatureFromDeclaration(d))
					if claim == nil || claim.ParameterIndex() != 0 {
						return 0, predicateFailure(
							verifier.l,
							n,
							"helper argument 0 does not occupy its predicate parameter; pass the tested value at the helper's predicate parameter index",
						)
					}
					return values[path.cell], nil
				}
			}
		}
		return 0, predicateFailure(verifier.l, n, "unconverged or opaque helper on this return path")
	}
	return 0, predicateFailure(verifier.l, n, "opaque test or possible mutation on this return path")
}

func (verifier *predicateVerifier) sameParameter(n *ast.Node, parameter *ast.Symbol) bool {
	n = ast.SkipParentheses(n)
	return ast.IsIdentifier(n) && verifier.l.checker.GetSymbolAtLocation(n) == parameter
}

func (verifier *predicateVerifier) constant(n *ast.Node) bool {
	n = ast.SkipParentheses(n)
	switch n.Kind {
	case ast.KindStringLiteral, ast.KindNumericLiteral, ast.KindTrueKeyword, ast.KindFalseKeyword:
		return true
	case ast.KindPropertyAccessExpression:
		symbol := verifier.l.checker.GetSymbolAtLocation(n)
		return symbol != nil &&
			(symbol.Flags&ast.SymbolFlagsEnumMember != 0 || verifier.l.checker.IsReadonlySymbol(symbol))
	}
	return false
}

func (verifier *predicateVerifier) pure(n *ast.Node) bool {
	n = ast.SkipParentheses(n)
	return n.Kind == ast.KindIdentifier || n.Kind == ast.KindNumericLiteral || n.Kind == ast.KindStringLiteral
}

func copyPredicatePath(path predicatePath) predicatePath {
	copy := predicatePath{cell: path.cell, locals: map[*ast.Symbol]predicateTruth{}}
	for symbol, value := range path.locals {
		copy.locals[symbol] = value
	}
	return copy
}

func (verifier *predicateVerifier) statements(
	nodes []*ast.Node,
	parameter *ast.Symbol,
	paths []predicatePath,
	check func(*ast.Node, predicateTruth) error,
) ([]predicatePath, error) {
	for _, n := range nodes {
		var next []predicatePath
		for _, path := range paths {
			switch n.Kind {
			case ast.KindBlock:
				result, statementsError := verifier.statements(
					n.AsBlock().Statements.Nodes,
					parameter,
					[]predicatePath{path},
					check,
				)
				if statementsError != nil {
					return nil, statementsError
				}
				next = append(next, result...)
			case ast.KindReturnStatement:
				result := predicateTrue
				if expression := n.AsReturnStatement().Expression; expression != nil {
					value, expressionError := verifier.expression(expression, parameter, path)
					if expressionError != nil {
						return nil, expressionError
					}
					result = value
				}
				if checkError := check(n, result); checkError != nil {
					return nil, checkError
				}
			case ast.KindIfStatement:
				branch := n.AsIfStatement()
				condition, expressionError := verifier.expression(branch.Expression, parameter, path)
				if expressionError != nil {
					return nil, expressionError
				}
				for _, side := range []struct {
					truth predicateTruth
					body  *ast.Node
				}{{predicateTrue, branch.ThenStatement}, {predicateFalse, branch.ElseStatement}} {
					if condition&side.truth == 0 {
						continue
					}
					p := copyPredicatePath(path)
					if side.body != nil && verifier.neverStatement(side.body, map[*ast.Node]bool{}) {
						continue
					}
					if side.body == nil {
						next = append(next, p)
						continue
					}
					result, statementsError := verifier.statements(
						[]*ast.Node{side.body},
						parameter,
						[]predicatePath{p},
						check,
					)
					if statementsError != nil {
						return nil, statementsError
					}
					next = append(next, result...)
				}
			case ast.KindVariableStatement:
				for _, d := range n.AsVariableStatement().DeclarationList.AsVariableDeclarationList().Declarations.Nodes {
					variable := d.AsVariableDeclaration()
					if !ast.IsIdentifier(d.Name()) || variable.Initializer == nil {
						return nil, predicateFailure(verifier.l, n, "an unproven alias on this return path")
					}
					value, expressionError := verifier.expression(variable.Initializer, parameter, path)
					if expressionError != nil {
						return nil, expressionError
					}
					path.locals[verifier.l.checker.GetSymbolAtLocation(d.Name())] = value
				}
				next = append(next, path)
			case ast.KindExpressionStatement:
				expression := n.AsExpressionStatement().Expression
				if verifier.neverCall(expression, map[*ast.Node]bool{}) {
					continue
				}
				if verifier.l.isPanicCall(expression) {
					call := expression.AsCallExpression()
					for _, argument := range call.Arguments.Nodes {
						if !verifier.pure(argument) {
							return nil, predicateFailure(
								verifier.l,
								argument,
								"opaque panic argument may mutate this path",
							)
						}
					}
					continue
				}
				if verifier.truthiness && expression.Kind == ast.KindCallExpression {
					return nil, predicateFailure(
						verifier.l,
						n,
						"a failure helper can return normally on this normal return path",
					)
				}
				_, expressionError := verifier.expression(expression, parameter, path)
				if expressionError != nil {
					return nil, expressionError
				}
				next = append(next, path)
			default:
				return nil, predicateFailure(verifier.l, n, "unsupported control flow or mutation on this return path")
			}
		}
		paths = next
	}
	return paths, nil
}

// An assertion on a condition has a truthiness postcondition, not a structural
// target. Partition all possible inputs into truthy and falsy values, rather
// than treating unknown as a boolean or trusting the asserts annotation.
func (lowering *lowering) proveConditionAssertion(node *ast.Node) (predicateProof, error) {
	declaration := node.Parent
	verifier := &predicateVerifier{l: lowering, truthiness: true, cells: []string{"truthy", "falsy"}}
	parameter := verifier.parameter(declaration)
	if parameter == nil || len(declaration.TypeParameters()) != 0 {
		return predicateProof{}, predicateFailure(lowering, node, "a non-parameter assertion has no truthiness proof")
	}
	for _, argument := range declaration.Parameters() {
		declared := argument.AsParameterDeclaration()
		if declared.Initializer != nil || declared.DotDotDotToken != nil || !ast.IsIdentifier(argument.Name()) {
			return predicateProof{}, predicateFailure(
				lowering,
				argument,
				"parameter initialization may change the asserted input",
			)
		}
	}
	if declaration.Body().Kind != ast.KindBlock {
		return predicateProof{}, predicateFailure(lowering, node, "an assertion needs a proved normal return path")
	}
	for cell := range verifier.cells {
		check := func(returned *ast.Node, _ predicateTruth) error {
			if verifier.cells[cell] != "truthy" {
				return predicateFailure(
					lowering,
					returned,
					"normal return does not establish the asserted condition's truthiness",
				)
			}
			return nil
		}
		paths, statementsError := verifier.statements(
			declaration.Body().AsBlock().Statements.Nodes,
			parameter,
			[]predicatePath{{cell: cell, locals: map[*ast.Symbol]predicateTruth{}}},
			check,
		)
		if statementsError != nil {
			return predicateProof{}, statementsError
		}
		if len(paths) != 0 {
			if checkError := check(declaration.Body(), predicateTrue); checkError != nil {
				return predicateProof{}, checkError
			}
		}
	}
	return predicateProof{}, nil
}

// A never annotation alone is not evidence. Inspect a direct implementation's
// exits; recursive or bodyless failure helpers have no termination summary.
func (verifier *predicateVerifier) neverCall(node *ast.Node, visiting map[*ast.Node]bool) bool {
	node = ast.SkipParentheses(node)
	if node.Kind != ast.KindCallExpression {
		return false
	}
	if verifier.l.isPanicCall(node) {
		return true
	}
	call := node.AsCallExpression()
	callee := ast.SkipParentheses(call.Expression)
	if !ast.IsIdentifier(callee) && callee.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	symbol := verifier.l.symbol(callee)
	if symbol == nil {
		return false
	}
	for _, declaration := range symbol.Declarations {
		if !ast.IsFunctionLike(declaration) || declaration.Body() == nil || visiting[declaration] {
			continue
		}
		signature := verifier.l.checker.GetSignatureFromDeclaration(declaration)
		if signature == nil ||
			verifier.l.checker.GetReturnTypeOfSignature(signature).Flags()&checker.TypeFlagsNever == 0 {
			continue
		}
		visiting[declaration] = true
		proved := verifier.neverStatement(declaration.Body(), visiting)
		delete(visiting, declaration)
		if proved {
			return true
		}
	}
	return false
}

func (verifier *predicateVerifier) neverStatement(node *ast.Node, visiting map[*ast.Node]bool) bool {
	switch node.Kind {
	case ast.KindThrowStatement:
		return true
	case ast.KindExpressionStatement:
		return verifier.neverCall(node.AsExpressionStatement().Expression, visiting)
	case ast.KindReturnStatement:
		expression := node.AsReturnStatement().Expression
		return expression != nil && verifier.neverCall(expression, visiting)
	case ast.KindIfStatement:
		branch := node.AsIfStatement()
		return branch.ElseStatement != nil &&
			verifier.neverStatement(branch.ThenStatement, visiting) &&
			verifier.neverStatement(branch.ElseStatement, visiting)
	case ast.KindBlock:
		for _, statement := range node.AsBlock().Statements.Nodes {
			if verifier.neverStatement(statement, visiting) {
				return true
			}
			if predicateNormalExit(statement) {
				return false
			}
		}
	}
	return false
}

// Untreated control flow may escape a block or override a pending throw.
// Ordinary statements before a terminal throw are harmless to non-return,
// even if they mutate state: they are never used as surviving narrowing facts.
func predicateNormalExit(node *ast.Node) bool {
	if ast.IsFunctionLike(node) {
		return false
	}
	switch node.Kind {
	case ast.KindReturnStatement, ast.KindBreakStatement, ast.KindContinueStatement, ast.KindTryStatement:
		return true
	}
	found := false
	node.ForEachChild(func(child *ast.Node) bool {
		if predicateNormalExit(child) {
			found = true
		}
		return found
	})
	return found
}

// Overloads share the implementation's control flow, under their admitted input.
// An annotation is never installed on the implementation as a checker fact.
func (lowering *lowering) predicateOverloadProven(implementation, overload *ast.Node) bool {
	outer := lowering.typeMapper
	defer func() { lowering.typeMapper = outer }()
	sources, targets := []*checker.Type{}, []*checker.Type{}
	for index, parameter := range implementation.TypeParameters() {
		if index >= len(overload.TypeParameters()) {
			return false
		}
		sources = append(sources, lowering.checker.GetTypeAtLocation(parameter.Name()))
		targets = append(targets, lowering.checker.GetTypeAtLocation(overload.TypeParameters()[index].Name()))
	}
	if len(sources) > 0 {
		lowering.typeMapper = newTypeMapper(sources, targets)
	}
	claim := predicateOfSignature(lowering.checker, lowering.checker.GetSignatureFromDeclaration(overload))
	if claim == nil || int(claim.ParameterIndex()) >= len(implementation.Parameters()) {
		return false
	}
	index := int(claim.ParameterIndex())
	if index < 0 {
		return false
	}
	parameter := implementation.Parameters()[index]
	if !ast.IsIdentifier(parameter.Name()) ||
		parameter.AsParameterDeclaration().Initializer != nil ||
		parameter.AsParameterDeclaration().DotDotDotToken != nil {
		return false
	}
	admitted, _, rest, censusOverloadParameterError := lowering.censusOverloadParameter(overload.Parameters(), index)
	if censusOverloadParameterError != nil || rest || admitted == nil {
		return false
	}
	changed := false
	var writes ast.Visitor
	writes = func(node *ast.Node) bool {
		if ast.IsIdentifier(node) &&
			lowering.symbol(node) == lowering.symbol(parameter.Name()) &&
			ast.IsAssignmentTarget(node) {
			changed = true
		}
		node.ForEachChild(writes)
		return false
	}
	implementation.Body().ForEachChild(writes)
	if changed {
		return false
	}
	target := claim.Type()
	if target == nil {
		if admitted.Flags()&checker.TypeFlagsBooleanLike == 0 {
			return false
		}
		target = predicateTrueType(lowering.checker)
	}
	proof := predicateFlowProof{
		l:         lowering,
		function:  implementation,
		parameter: parameter,
		declared:  lowering.checker.GetTypeAtLocation(parameter.Name()),
		admitted:  admitted,
		target:    lowering.concrete(target),
		assertion: claim.Kind() == checker.TypePredicateKindAssertsIdentifier,
	}
	return proof.proveBody(overload.Type()) == nil
}

// Reification is limited to independently checkable contracts. Tagged interfaces
// continue through the shared view checker; untagged/callable contracts are not
// silently equated with a primitive typeof test.
func (lowering *lowering) predicateMembership(
	node *ast.Node,
	value ir.Expression,
	source, target *checker.Type,
	local func(string, ir.Type) int,
	depth int,
) ([]ir.Statement, ir.Expression, error) {
	target = lowering.concrete(target)
	source = lowering.concrete(source)
	if depth > 8 {
		return nil, nil, lowering.notYet(node, "a recursive predicate overload target")
	}
	equal := func(a, b ir.Expression) ir.Expression { return ir.Binary{Operator: ir.Equal, Left: a, Right: b} }
	isType := func(name string) ir.Expression {
		return equal(ir.TypeOf{Value: value}, ir.StringConstant{Index: lowering.constant(name)})
	}
	if target.Flags()&checker.TypeFlagsUnion != 0 {
		var setup []ir.Statement
		var result ir.Expression = ir.BooleanConstant{Value: false}
		for _, member := range target.Types() {
			statements, test, predicateMembershipError := lowering.predicateMembership(
				node,
				value,
				source,
				member,
				local,
				depth+1,
			)
			if predicateMembershipError != nil {
				return nil, nil, predicateMembershipError
			}
			setup = append(setup, statements...)
			result = ir.Binary{Operator: ir.Or, Left: result, Right: test}
		}
		return setup, result, nil
	}
	if target.Flags()&checker.TypeFlagsUndefined != 0 {
		return nil, isType("undefined"), nil
	}
	name := ""
	of := ir.Type(0)
	switch {
	case target.Flags()&checker.TypeFlagsStringLike != 0:
		name = "string"
		of = ir.String
	case target.Flags()&checker.TypeFlagsNumberLike != 0:
		name = "number"
		of = ir.Number
	case target.Flags()&checker.TypeFlagsBooleanLike != 0:
		name = "boolean"
		of = ir.Boolean
	}
	if name != "" {
		test := isType(name)
		if literal, _, ok := lowering.literalConstant(target); ok {
			narrowed := value
			if value.Type() == ir.Union {
				narrowed = ir.Narrow{Value: value, To: of}
			} else if value.Type().IsMaybe() {
				narrowed = ir.Unwrap{Value: value}
			}
			if narrowed.Type() != of {
				return nil, nil, lowering.notYet(node, "a predicate overload literal representation")
			}
			test = ir.Binary{Operator: ir.And, Left: test, Right: equal(narrowed, literal)}
		}
		return nil, test, nil
	}
	if lowering.isArrayType(target) && (source.Flags()&checker.TypeFlagsUndefined != 0 ||
		value.Type() == ir.Number ||
		value.Type() == ir.String ||
		value.Type() == ir.Boolean) {
		return nil, ir.BooleanConstant{Value: false}, nil
	}
	if lowering.isArrayType(target) && value.Type() == ir.Array {
		arraySource := source
		if source.Flags()&checker.TypeFlagsUnion != 0 {
			arraySource = nil
			for _, member := range source.Types() {
				if lowering.isArrayType(member) {
					if arraySource != nil {
						return nil, nil, lowering.notYet(
							node,
							"a predicate overload over multiple array representations",
						)
					}
					arraySource = member
				}
			}
		}
		if arraySource == nil || !lowering.isArrayType(arraySource) {
			return nil, nil, lowering.notYet(node, "a predicate overload without a known array representation")
		}
		element := lowering.elementTypeOfArray(arraySource)
		// An identical element contract is already admitted on the source array.
		// The overload strengthens only presence, not its element view.
		if checker.Checker_isTypeIdenticalTo(
			lowering.checker,
			lowering.concrete(element),
			lowering.concrete(lowering.elementTypeOfArray(target)),
		) {
			return nil, censusCondition(value), nil
		}
		held, known := lowering.kept(element)
		if !known || slotless(held) {
			return nil, nil, lowering.notYet(
				node,
				"a checked predicate overload array element "+lowering.checker.TypeToString(element),
			)
		}
		matches := local("array_matches", ir.Boolean)
		counter := local("array_index", ir.Number)
		read := ir.Read{Local: counter, Of: ir.Number}
		statements, test, predicateMembershipError := lowering.predicateMembership(
			node,
			ir.ArrayIndex{Array: value, Index: read, Element: held},
			element,
			lowering.elementTypeOfArray(target),
			local,
			depth+1,
		)
		if predicateMembershipError != nil {
			return nil, nil, predicateMembershipError
		}
		loopBody := append(statements, ir.If{
			Condition: ir.Unary{Operator: ir.Not, Operand: test},
			Then:      []ir.Statement{ir.Assign{Local: matches, Value: ir.BooleanConstant{Value: false}}, ir.Break{}},
		})
		setup := []ir.Statement{
			ir.Declare{Local: matches, Value: ir.BooleanConstant{Value: false}},
			ir.If{Condition: censusCondition(value), Then: []ir.Statement{
				ir.Assign{Local: matches, Value: ir.BooleanConstant{Value: true}},
				ir.Declare{Local: counter, Value: ir.NumberConstant{Value: 0}},
				ir.Loop{
					Condition: ir.Binary{Operator: ir.Less, Left: read, Right: ir.Length{Array: value}},
					Body:      loopBody,
					Update: []ir.Statement{
						ir.Assign{
							Local: counter,
							Value: ir.Binary{Operator: ir.Add, Left: read, Right: ir.NumberConstant{Value: 1}},
						},
					},
				},
			}},
		}
		return setup, ir.Read{Local: matches, Of: ir.Boolean}, nil
	}
	if isClassInstance(target) && target.Symbol() != nil {
		for _, declaration := range target.Symbol().Declarations {
			if declaration.Kind != ast.KindClassDeclaration || len(declaration.TypeParameters()) != 0 {
				continue
			}
			if value.Type() != ir.Object && value.Type() != ir.Union {
				return nil, ir.BooleanConstant{Value: false}, nil
			}
			instance, instantiateError := lowering.instantiate(declaration, target, node)
			if instantiateError != nil {
				return nil, nil, instantiateError
			}
			object := value
			if value.Type() == ir.Union {
				object = ir.Narrow{Value: value, To: ir.Object}
			}
			return nil, ir.Binary{
				Operator: ir.And,
				Left:     isType("object"),
				Right:    ir.InstanceOf{Value: object, Class: instance.class},
			}, nil
		}
	}
	if target.Flags()&checker.TypeFlagsObject != 0 && !isClassInstance(target) && value.Type() == ir.Object {
		tag := lowering.fieldLiteral(target, "kind")
		if tag != nil {
			// The tag chooses the member, and the view checks its layout at every read. A checked
			// predicate's call registers it here, which is what makes an open enum tag safe on the
			// true branch (#k1t2wp9). predicateRefusal registers the same view for a body proven by
			// tag alone, so either covers the other on these programs and no test guards one alone:
			// keep both.
			if _, viewError := lowering.view(node, nil, source, target); viewError != nil {
				return nil, nil, viewError
			}
			// The value enters the view wherever a caller's read is narrowed to it (view_widened.go).
			lowering.viewSources[len(lowering.viewSources)-1].narrows = true
			literal, of, known := lowering.literalConstant(tag)
			if known {
				return nil, ir.Binary{Operator: ir.And, Left: censusCondition(value), Right: equal(ir.Property{
					Object:   value,
					Name:     "kind",
					Of:       of,
					View:     "predicate argument.kind",
					ViewType: lowering.checker.TypeToString(tag),
				}, literal)}, nil
			}
		} else if lowering.brandedSupertype(target) {
			// No kind set can say which kinds a branded supertype holds (#c7w6j0y, ruled Oct 31), so the call
			// checks only that the value is an object, and the views check every field where it's read.
			if brandedSupertypeViewsError := lowering.brandedSupertypeViews(
				node,
				source,
				target,
			); brandedSupertypeViewsError != nil {
				return nil, nil, brandedSupertypeViewsError
			}
			return nil, censusCondition(value), nil
		}
	}
	if cause := lowering.checkedPredicateTargetCause(target); cause != "" {
		return nil, nil, lowering.notYet(node, cause+" ("+lowering.checker.TypeToString(target)+")")
	}
	return nil, nil, lowering.notYet(node, "checked predicate overload target "+lowering.checker.TypeToString(target))
}

// checkedPredicateTargetCause names why an untagged target, one with no kind property, has no check a
// runtime can afford, or "" when a structural check might serve (#c7w6j0y, ruled Oct 30 into group 10,
// with the casts the runtime can't check). Its base, the intersection's named parts, may have no required
// field, as CompilerOptions has only optional fields and an index signature, so nothing identifies the
// value. Or a required field holds an untagged record, alone, in a union or as an array's elements, so
// the check would read structure after structure past the depth bound, as the build infos do. A target
// with a kind property, tagged or a branded supertype, is the tag rules', never this.
func (lowering *lowering) checkedPredicateTargetCause(target *checker.Type) string {
	if target.Flags()&(checker.TypeFlagsObject|checker.TypeFlagsIntersection) == 0 ||
		isClassInstance(target) ||
		lowering.checker.GetPropertyOfType(target, "kind") != nil {
		return ""
	}
	base := []*checker.Type{target}
	if target.Flags()&checker.TypeFlagsIntersection != 0 {
		base = nil
		for _, part := range target.Types() {
			if part.ObjectFlags()&checker.ObjectFlagsAnonymous == 0 {
				base = append(base, part)
			}
		}
		if len(base) == 0 {
			base = target.Types() // Anonymous parts alone are their own base.
		}
	}
	required := false
	for _, part := range base {
		for _, property := range lowering.checker.GetPropertiesOfType(part) {
			required = required || property.Flags&ast.SymbolFlagsOptional == 0
		}
	}
	if !required {
		return "a checked predicate target whose base has no required field to check"
	}
	for _, property := range lowering.checker.GetPropertiesOfType(target) {
		if property.Flags&ast.SymbolFlagsOptional == 0 &&
			lowering.untaggedRecord(lowering.checker.GetTypeOfSymbol(property), map[*checker.Type]bool{}) {
			return "a checked predicate target whose check would read past the depth bound"
		}
	}
	return ""
}

// untaggedRecord is whether t holds an object with no kind property, alone, as a union member or as an
// array's elements: a structure a check could only read field by field. A library type (Map, Date,
// Promise) isn't one, since its check isn't a matter of depth. visited ends a recursive alias, such as
// `type Json = string | number | Json[]`, on its second visit.
func (lowering *lowering) untaggedRecord(t *checker.Type, visited map[*checker.Type]bool) bool {
	if visited[t] {
		return false
	}
	visited[t] = true
	if t.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range t.Types() {
			if lowering.untaggedRecord(member, visited) {
				return true
			}
		}
		return false
	}
	if lowering.isArrayType(t) || checker.IsTupleType(t) {
		for _, element := range lowering.checker.GetTypeArguments(t) {
			if lowering.untaggedRecord(element, visited) {
				return true
			}
		}
		return false
	}
	if symbol := t.Symbol(); symbol != nil && len(symbol.Declarations) > 0 {
		if file := ast.GetSourceFileOfNode(symbol.Declarations[0]); load.IsLibrary(file) || load.IsPrelude(file) {
			return false
		}
	}
	return t.Flags()&checker.TypeFlagsObject != 0 &&
		!isClassInstance(t) &&
		lowering.checker.GetPropertyOfType(t, "kind") == nil &&
		len(lowering.checker.GetSignaturesOfType(t, checker.SignatureKindCall)) == 0
}

// A direction is observable whenever the checker still narrows a value read in
// its flow region. This includes aliases of the boolean result and continuation
// flows after return/throw; lexical then/else scanning would miss those reads.
func (lowering *lowering) predicateUseDirections(call *ast.CallExpression) predicateTruth {
	claim := predicateOfSignature(lowering.checker, lowering.checker.GetResolvedSignature(call.AsNode()))
	if claim == nil || int(claim.ParameterIndex()) < 0 || int(claim.ParameterIndex()) >= len(call.Arguments.Nodes) {
		return predicateEither
	}
	argument := ast.SkipParentheses(call.Arguments.Nodes[claim.ParameterIndex()])
	references := []*ast.Node{argument}
	for container := argument; container.Flags&ast.NodeFlagsOptionalChain != 0; {
		switch container.Kind {
		case ast.KindPropertyAccessExpression:
			container = ast.SkipParentheses(container.AsPropertyAccessExpression().Expression)
		case ast.KindElementAccessExpression:
			container = ast.SkipParentheses(container.AsElementAccessExpression().Expression)
		default:
			container = nil
		}
		if container == nil {
			break
		}
		references = append(references, container)
	}
	var directions predicateTruth
	for _, argument := range references {
		directions |= lowering.predicateReferenceDirections(call, argument)
	}
	return directions
}

func (lowering *lowering) predicateReferenceDirections(call *ast.CallExpression, argument *ast.Node) predicateTruth {
	if argument.Kind != ast.KindIdentifier &&
		argument.Kind != ast.KindPropertyAccessExpression &&
		argument.Kind != ast.KindElementAccessExpression &&
		argument.Kind != ast.KindThisKeyword {
		// Temporary values cannot be narrowed by a later reference read.
		return 0
	}
	declared := lowering.checker.GetTypeAtLocation(argument)
	if argument.Kind == ast.KindIdentifier {
		declared = lowering.checker.GetTypeOfSymbol(lowering.symbol(argument))
	}
	var directions predicateTruth
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if predicateCannotMatch(node, argument) {
			node.ForEachChild(visit)
			return false
		}
		if lowering.predicateSameReference(node, argument) &&
			!ast.IsDeclarationName(node) &&
			!predicateWriteOnly(node) &&
			!ast.IsPartOfTypeNode(node) {
			actual := lowering.checker.GetTypeAtLocation(node)
			if !checker.Checker_isTypeIdenticalTo(lowering.checker, actual, declared) {
				if flow := node.FlowNodeData(); flow != nil {
					candidates := lowering.predicateFlowDirections(
						flow.FlowNode,
						call.AsNode(),
						argument,
						map[*ast.FlowNode]bool{},
					)
					for _, direction := range []predicateTruth{predicateTrue, predicateFalse} {
						if candidates&direction == 0 {
							continue
						}
						without := lowering.predicateFlowWithout(flow.FlowNode, call.AsNode(), node, direction)
						unclaimed := predicateFlowType(lowering.checker, node, declared, declared, nil, without)
						if !checker.Checker_isTypeIdenticalTo(lowering.checker, actual, unclaimed) {
							directions |= direction
						}
					}
				}
			}
		}
		node.ForEachChild(visit)
		return false
	}
	for _, file := range lowering.program.Files() {
		file.AsNode().ForEachChild(visit)
	}
	return directions
}

// An identifier argument matches only an identifier resolving to its symbol, so one spelled the
// same, or a declaration of it; the checker resolves each other identifier it's asked about, and
// a property name like the b in a.b resolves to nothing, so it then computes spelling suggestions
// it throws away: 94% of tsc's mapToTypeNodes. A declaration of the symbol isn't an identifier, so
// it still reaches the checker.
func predicateCannotMatch(node, argument *ast.Node) bool {
	return argument.Kind == ast.KindIdentifier && node.Kind == ast.KindIdentifier && node.Text() != argument.Text()
}

func (lowering *lowering) predicateSameReference(a, b *ast.Node) bool {
	return a != nil && b != nil && predicateMatchingReference(lowering.checker, b, a)
}

// Follow the checker's existing graph rather than inventing branch regions.
func (lowering *lowering) predicateFlowDirections(
	flow *ast.FlowNode,
	call, reference *ast.Node,
	seen map[*ast.FlowNode]bool,
) predicateTruth {
	if flow == nil || seen[flow] {
		return 0
	}
	seen[flow] = true
	// A replacement kills earlier predicate facts for this reference. Compound
	// assignments also read the old narrowed value and must keep those checks.
	if flow.Flags&ast.FlowFlagsAssignment != 0 &&
		lowering.predicateSameReference(flow.Node, reference) &&
		predicateWriteOnly(flow.Node) {
		return 0
	}
	var result predicateTruth
	if flow.Flags&ast.FlowFlagsCondition != 0 {
		truth := flow.Flags&ast.FlowFlagsTrueCondition != 0
		result |= lowering.predicateConditionDirection(flow.Node, call, truth, 0)
	}
	if flow.Flags&ast.FlowFlagsSwitchClause != 0 {
		result |= lowering.predicateSwitchDirection(flow.Node.AsFlowSwitchClauseData(), call)
	}
	if flow.Flags&ast.FlowFlagsCall != 0 && flow.Node == call {
		result |= predicateTrue
	}
	result |= lowering.predicateFlowDirections(flow.Antecedent, call, reference, seen)
	for list := flow.Antecedents; list != nil; list = list.Next {
		result |= lowering.predicateFlowDirections(list.Flow, call, reference, seen)
	}
	if flow.Flags&ast.FlowFlagsStart != 0 && flow.Node != nil {
		if outer := flow.Node.FlowNodeData(); outer != nil {
			result |= lowering.predicateFlowDirections(outer.FlowNode, call, reference, seen)
		}
	}
	return result
}

func (lowering *lowering) predicateConditionDirection(node, call *ast.Node, truth bool, depth int) predicateTruth {
	if node == nil || depth > 16 {
		return 0
	}
	node = ast.SkipParentheses(node)
	if node == call {
		if truth {
			return predicateTrue
		}
		return predicateFalse
	}
	if node.Kind == ast.KindPrefixUnaryExpression &&
		node.AsPrefixUnaryExpression().Operator == ast.KindExclamationToken {
		return lowering.predicateConditionDirection(node.AsPrefixUnaryExpression().Operand, call, !truth, depth+1)
	}
	if node.Kind == ast.KindBinaryExpression {
		binary := node.AsBinaryExpression()
		switch binary.OperatorToken.Kind {
		case ast.KindAmpersandAmpersandToken, ast.KindBarBarToken:
			precise := truth == (binary.OperatorToken.Kind == ast.KindAmpersandAmpersandToken)
			if precise {
				return lowering.predicateConditionDirection(
					binary.Left,
					call,
					truth,
					depth+1,
				) | lowering.predicateConditionDirection(binary.Right, call, truth, depth+1)
			}
			return lowering.predicateConditionDirection(
				binary.Left,
				call,
				true,
				depth+1,
			) | lowering.predicateConditionDirection(
				binary.Left,
				call,
				false,
				depth+1,
			) | lowering.predicateConditionDirection(
				binary.Right,
				call,
				true,
				depth+1,
			) | lowering.predicateConditionDirection(binary.Right, call, false, depth+1)
		case ast.KindEqualsToken, ast.KindBarBarEqualsToken, ast.KindAmpersandAmpersandEqualsToken, ast.KindQuestionQuestionEqualsToken, ast.KindCommaToken:
			return lowering.predicateConditionDirection(binary.Right, call, truth, depth+1)
		case ast.KindEqualsEqualsToken, ast.KindExclamationEqualsToken, ast.KindEqualsEqualsEqualsToken, ast.KindExclamationEqualsEqualsToken:
			// As the checker's narrowTypeByBooleanComparison: compared with a boolean literal, either
			// way round, the other operand narrows as itself or its negation.
			operand, literal := binary.Left, ast.SkipParentheses(binary.Right)
			if !ast.IsBooleanLiteral(literal) {
				operand, literal = binary.Right, ast.SkipParentheses(binary.Left)
			}
			if ast.IsBooleanLiteral(literal) && !ast.IsAccessExpression(ast.SkipParentheses(operand)) {
				equal := binary.OperatorToken.Kind == ast.KindEqualsEqualsToken ||
					binary.OperatorToken.Kind == ast.KindEqualsEqualsEqualsToken
				return lowering.predicateConditionDirection(
					operand,
					call,
					(truth == (literal.Kind == ast.KindTrueKeyword)) == equal,
					depth+1,
				)
			}
		}
	}
	if node.Kind == ast.KindNonNullExpression || node.Kind == ast.KindSatisfiesExpression {
		return lowering.predicateConditionDirection(node.Expression(), call, truth, depth+1)
	}
	if node.Kind == ast.KindIdentifier {
		if symbol := lowering.symbol(node); symbol != nil {
			for _, declaration := range symbol.Declarations {
				if declaration.Kind == ast.KindVariableDeclaration && declaration.Parent.Flags&ast.NodeFlagsConst != 0 {
					return lowering.predicateConditionDirection(
						declaration.AsVariableDeclaration().Initializer,
						call,
						truth,
						depth+1,
					)
				}
			}
		}
	}
	return 0
}

// As the checker's narrowTypeBySwitchOnTrue: under switch (true), the case clauses before this
// clause set, and after it when the set holds the default, narrow by their false branch; the set's
// own case clauses by their true branch.
func (lowering *lowering) predicateSwitchDirection(data *ast.FlowSwitchClauseData, call *ast.Node) predicateTruth {
	statement := data.SwitchStatement.AsSwitchStatement()
	if ast.SkipParentheses(statement.Expression).Kind != ast.KindTrueKeyword {
		return 0
	}
	clauses := statement.CaseBlock.AsCaseBlock().Clauses.Nodes
	start, end := int(data.ClauseStart), int(data.ClauseEnd)
	holdsDefault := start == end
	for index := start; index < end; index++ {
		if clauses[index].Kind == ast.KindDefaultClause {
			holdsDefault = true
		}
	}
	var result predicateTruth
	for index, clause := range clauses {
		if clause.Kind != ast.KindCaseClause {
			continue
		}
		switch {
		case index < start || (holdsDefault && index >= end):
			result |= lowering.predicateConditionDirection(clause.Expression(), call, false, 0)
		case index < end && !holdsDefault:
			result |= lowering.predicateConditionDirection(clause.Expression(), call, true, 0)
		}
	}
	return result
}

func predicateWriteOnly(node *ast.Node) bool {
	if node == nil || !ast.IsAssignmentTarget(node) {
		return false
	}
	parent := node.Parent
	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		parent = parent.Parent
	}
	return parent != nil &&
		parent.Kind == ast.KindBinaryExpression &&
		parent.AsBinaryExpression().OperatorToken.Kind == ast.KindEqualsToken
}

// Recompute the read with this call's directional fact removed. Merely finding
// an ancestor condition would attribute a later, independent guard's narrowing
// to every earlier call, even after their branches have rejoined.
// Fresh flow nodes keep the pinned checker's cache and original graph untouched.
func (lowering *lowering) predicateFlowWithout(
	root *ast.FlowNode,
	call, reference *ast.Node,
	direction predicateTruth,
) *ast.FlowNode {
	memo := map[*ast.FlowNode]*ast.FlowNode{}
	var clone func(*ast.FlowNode) *ast.FlowNode
	var list func(*ast.FlowList) *ast.FlowList
	list = func(original *ast.FlowList) *ast.FlowList {
		if original == nil {
			return nil
		}
		return &ast.FlowList{Flow: clone(original.Flow), Next: list(original.Next)}
	}
	clone = func(original *ast.FlowNode) *ast.FlowNode {
		if original == nil {
			return nil
		}
		if previous := memo[original]; previous != nil {
			return previous
		}
		copy := &ast.FlowNode{
			Flags: original.Flags &^ (ast.FlowFlagsShared | ast.FlowFlagsReferenced),
			Node:  original.Node,
		}
		memo[original] = copy
		remove := original.Flags&ast.FlowFlagsCall != 0 && original.Node == call && direction == predicateTrue
		if original.Flags&ast.FlowFlagsCondition != 0 {
			truth := original.Flags&ast.FlowFlagsTrueCondition != 0
			remove = lowering.predicateConditionDirection(original.Node, call, truth, 0)&direction != 0
		}
		// Dropping a whole clause set also drops its other clauses' narrowing, which can only widen
		// the unclaimed type and so report the direction, never hide it.
		if original.Flags&ast.FlowFlagsSwitchClause != 0 {
			remove = lowering.predicateSwitchDirection(original.Node.AsFlowSwitchClauseData(), call)&direction != 0
		}
		if remove {
			copy.Flags, copy.Node = ast.FlowFlagsBranchLabel, nil
			copy.Antecedents = &ast.FlowList{Flow: clone(original.Antecedent)}
			return copy
		}
		copy.Antecedent, copy.Antecedents = clone(original.Antecedent), list(original.Antecedents)
		if original.Flags&ast.FlowFlagsStart != 0 &&
			original.Node != nil &&
			original.Node.FlowNodeData() != nil &&
			original.Node.FlowNodeData().FlowNode != nil {
			// Stable identifier captures inherit the container's flow. Reconstruct
			// that edge explicitly, without altering its shared AST flow node.
			if reference.Kind == ast.KindIdentifier ||
				(reference.Kind == ast.KindThisKeyword && ast.IsArrowFunction(original.Node)) {
				copy.Flags, copy.Node = ast.FlowFlagsBranchLabel, nil
				copy.Antecedents = &ast.FlowList{Flow: clone(original.Node.FlowNodeData().FlowNode)}
			}
		}
		if original.Flags&ast.FlowFlagsReduceLabel != 0 {
			data := original.Node.AsFlowReduceLabelData()
			copy.Node = ast.NewFlowReduceLabelData(clone(data.Target), list(data.Antecedents))
		}
		return copy
	}
	return clone(root)
}

func (lowering *lowering) predicateOverloadOrdinal(implementation, overload *ast.Node) int {
	ordinal := 0
	for _, declaration := range lowering.symbol(implementation.Name()).Declarations {
		if declaration.Kind == ast.KindFunctionDeclaration && declaration.Body() == nil {
			ordinal++
			if declaration == overload {
				return ordinal
			}
		}
	}
	return ordinal
}

// Inferred callback predicates are claims too. Independently prove the resolved
// target from their own body; checker inference alone is not admission evidence.
func (lowering *lowering) proveInferredPredicate(function *ast.Node, target *checker.Type) bool {
	claim := predicateOfSignature(lowering.checker, lowering.checker.GetSignatureFromDeclaration(function))
	if claim == nil ||
		claim.Type() == nil ||
		!checker.Checker_isTypeIdenticalTo(lowering.checker, claim.Type(), target) {
		return false
	}
	index := int(claim.ParameterIndex())
	if index < 0 || index >= len(function.Parameters()) {
		return false
	}
	parameter := function.Parameters()[index]
	if parameter.Name().Kind != ast.KindIdentifier ||
		parameter.AsParameterDeclaration().Initializer != nil ||
		parameter.AsParameterDeclaration().DotDotDotToken != nil {
		return false
	}
	changed := false
	var writes ast.Visitor
	writes = func(node *ast.Node) bool {
		if node.Kind == ast.KindIdentifier &&
			lowering.symbol(node) == lowering.symbol(parameter.Name()) &&
			ast.IsAssignmentTarget(node) {
			changed = true
		}
		node.ForEachChild(writes)
		return false
	}
	function.Body().ForEachChild(writes)
	if changed {
		return false
	}
	proof := predicateFlowProof{
		l:         lowering,
		function:  function,
		parameter: parameter,
		declared:  lowering.checker.GetTypeAtLocation(parameter.Name()),
		target:    target,
	}
	return proof.proveBody(function) == nil
}
