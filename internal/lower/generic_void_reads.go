package lower

import (
	"fmt"
	"sort"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
)

// refuseGenericVoidReads refuses an instantiation of a generic at void whose body reads the value
// (#p7t8tmh part 2, ruled Oct 27 and 28). Adamic monomorphizes, so `showAll<T>(xs: T[])` called with a
// void[] is a body where every T is void, and a discard released what Node kept: String(x) there
// would print undefined where Node prints 42. Within the body a type parameter bound to void is a void
// value, and one bound to a type holding void is a void position, so the void-container classifier
// judges every expression whose type mentions it, with the same list of moves: one classifier for the
// generic and the concrete case. A body that only moves its T values (stores them in slots of exactly
// T, returns them, reads length) stays legal.
//
// bound maps each declared type parameter to what this instantiation makes it, or nil where it
// couldn't be read back; resolved and target are the call's signature and the declaration's, read for
// an unmapped parameter, which fails closed. Each declaration is judged once per binding.
func (l *lowering) refuseGenericVoidReads(where, declaration *ast.Node, bound map[*checker.Type]*checker.Type, target, resolved *checker.Signature) error {
	voidBound := map[*checker.Type]bool{}
	names := []string{}
	// The memo keys on each binding's type itself, never its rendering: two types that print alike
	// (a local Box in each of two modules) are judged apart.
	identities := []string{}
	unmapped := false
	for _, parameter := range declaration.TypeParameters() {
		parameterType := l.checker.GetTypeAtLocation(parameter.Name())
		concrete, known := bound[parameterType]
		if !known || concrete == nil {
			unmapped = true
			continue
		}
		switch {
		case concrete.Flags() == checker.TypeFlagsVoid:
			voidBound[parameterType] = true
		case l.holdsVoid(concrete):
			voidBound[parameterType] = false
		default:
			continue
		}
		names = append(names, parameter.Name().Text()+" = "+l.checker.TypeToString(concrete))
		identities = append(identities, parameter.Name().Text()+"="+fmt.Sprintf("%p", concrete))
	}
	if unmapped && l.signatureReachesVoid(target, resolved) {
		// A type parameter not read back (it stands only inside an object, { items: T[] }) is taken to
		// be void itself, the strictest binding, when the resolved signature holds void where the
		// declared one is generic.
		for _, parameter := range declaration.TypeParameters() {
			parameterType := l.checker.GetTypeAtLocation(parameter.Name())
			if concrete, known := bound[parameterType]; !known || concrete == nil {
				voidBound[parameterType] = true
				names = append(names, parameter.Name().Text()+" = void (read from the signature)")
				identities = append(identities, parameter.Name().Text()+"=unmapped")
			}
		}
	}
	if len(voidBound) == 0 {
		return nil
	}
	sort.Strings(names)
	sort.Strings(identities)
	key := l.program.Where(declaration) + ";" + strings.Join(identities, ";")
	if verdict, judged := l.genericVoidVerdicts[key]; judged {
		return verdict
	}
	outerBound, outerHolders := l.voidBound, l.voidHolders
	l.voidBound, l.voidHolders = voidBound, map[*checker.Type]bool{}
	read, role := l.genericVoidRead(declaration)
	l.voidBound, l.voidHolders = outerBound, outerHolders
	var verdict error
	if read != nil {
		name := "the generic"
		if declarationName := declaration.Name(); declarationName != nil {
			name = declarationName.Text()
		}
		verdict = &Refused{Where: l.program.Where(where), What: "a generic instantiated at void whose body reads the value: " + name + "'s " + strings.Join(names, ", ") + ", read at " + l.program.Where(read) + " " + role, Fix: "give the function a return type, or don't hand a void container to a generic that reads its elements (adamic/no-void-read)"}
	}
	if l.genericVoidVerdicts == nil {
		l.genericVoidVerdicts = map[string]error{}
	}
	l.genericVoidVerdicts[key] = verdict
	return verdict
}

// signatureReachesVoid is whether a parameter or the result the declaration states generically is,
// as this call resolved it, void or a type holding void.
func (l *lowering) signatureReachesVoid(target, resolved *checker.Signature) bool {
	if target == nil || resolved == nil {
		return true // Nothing to read back: fail closed.
	}
	reaches := func(declared, given *checker.Type) bool {
		return declared != given && given != nil && l.voidAnywhere(l.concrete(given), 0, map[*checker.Type]bool{})
	}
	declared, given := target.Parameters(), resolved.Parameters()
	for index := range declared {
		if index < len(given) && reaches(l.checker.GetTypeOfSymbol(declared[index]), l.checker.GetTypeOfSymbol(given[index])) {
			return true
		}
	}
	return reaches(l.checker.GetReturnTypeOfSignature(target), l.checker.GetReturnTypeOfSignature(resolved))
}

// genericVoidRead walks a declaration's body (nested functions and a class's members included) and
// returns the first expression whose use reads a void value or a void position, with its role.
func (l *lowering) genericVoidRead(declaration *ast.Node) (*ast.Node, string) {
	var read *ast.Node
	role := ""
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if read != nil {
			return true
		}
		if found := l.genericVoidUse(node); found != "" {
			read, role = node, found
			return true
		}
		return node.ForEachChild(visit)
	}
	declaration.ForEachChild(visit)
	return read, role
}

// genericVoidUse judges one node in a body where type parameters are bound to void: a void value by
// voidValueUse, a container holding void by the classifier's own rule.
func (l *lowering) genericVoidUse(node *ast.Node) string {
	if node.Kind == ast.KindParenthesizedExpression || !l.isExpression(node) || ast.IsDeclarationName(node) {
		return ""
	}
	if node.Kind == ast.KindExpressionWithTypeArguments {
		return "" // A heritage clause names a type, and holds no value.
	}
	if ast.IsAssignmentTarget(node) {
		// ??=, ||= and &&= read their target before they decide to write it (Lower, on #p7t8tmh).
		site := node
		for site.Parent != nil && site.Parent.Kind == ast.KindParenthesizedExpression {
			site = site.Parent
		}
		if parent := site.Parent; parent != nil && parent.Kind == ast.KindBinaryExpression && parent.AsBinaryExpression().Left == site && logicalAssignment(parent.AsBinaryExpression().OperatorToken.Kind) && l.isVoidValue(l.checker.GetTypeAtLocation(node)) {
			return "as the target of " + scanner.TokenToString(parent.AsBinaryExpression().OperatorToken.Kind)
		}
		return ""
	}
	if node.Parent != nil && node.Parent.Kind == ast.KindPropertyAccessExpression && node.Parent.AsPropertyAccessExpression().Name() == node {
		return ""
	}
	source := l.checker.GetTypeAtLocation(node)
	switch {
	case source == nil:
		return ""
	case l.isVoidValue(source):
		return l.voidValueUse(node, source)
	case l.holdsVoid(source):
		return l.voidContainerUse(node, source)
	}
	return ""
}

// voidValueUse is H2's rule for a value of a type parameter bound to void: a discard keeps it, a slot
// of exactly that parameter keeps it, and flowing on where the whole expression stands keeps it while
// the whole is judged in turn. Every other use reads it, identity included, since === compares the
// value itself.
func (l *lowering) voidValueUse(node *ast.Node, source *checker.Type) string {
	site := node
	for site.Parent != nil && site.Parent.Kind == ast.KindParenthesizedExpression {
		site = site.Parent
	}
	parent := site.Parent
	if parent == nil {
		return "as a value"
	}
	keeps := func(target *checker.Type, role string) string {
		if target == nil || !l.voidKeepsPosition(source, target, 0, map[[2]*checker.Type]bool{}) {
			return "as " + role
		}
		return ""
	}
	switch parent.Kind {
	case ast.KindExpressionStatement, ast.KindVoidExpression:
		return ""
	case ast.KindForStatement:
		statement := parent.AsForStatement()
		if statement.Initializer == site || statement.Incrementor == site {
			return ""
		}
		return "as a loop condition"
	case ast.KindBinaryExpression:
		binary := parent.AsBinaryExpression()
		switch binary.OperatorToken.Kind {
		case ast.KindCommaToken:
			if binary.Left == site {
				return ""
			}
			return l.voidContainerFlows(parent)
		case ast.KindAmpersandAmpersandToken, ast.KindBarBarToken, ast.KindQuestionQuestionToken:
			if binary.Right == site {
				return l.voidContainerFlows(parent)
			}
		case ast.KindEqualsToken:
			return keeps(l.checker.GetTypeAtLocation(binary.Left), "an assigned value")
		}
		return "as an operand of " + scanner.TokenToString(binary.OperatorToken.Kind)
	case ast.KindConditionalExpression:
		if parent.AsConditionalExpression().Condition == site {
			return "as a condition"
		}
		return l.voidContainerFlows(parent)
	case ast.KindArrayLiteralExpression:
		return l.voidContainerFlows(parent)
	case ast.KindPropertyAssignment, ast.KindShorthandPropertyAssignment:
		return l.voidContainerFlows(parent.Parent)
	case ast.KindAwaitExpression, ast.KindNonNullExpression:
		return l.voidContainerFlows(parent)
	case ast.KindAsExpression, ast.KindSatisfiesExpression:
		if role := keeps(l.checker.GetTypeAtLocation(parent), "an assertion's operand"); role != "" {
			return role
		}
		return l.voidContainerFlows(parent)
	case ast.KindVariableDeclaration, ast.KindParameter, ast.KindPropertyDeclaration, ast.KindBindingElement:
		name := parent.Name()
		if name == nil || name.Kind != ast.KindIdentifier {
			return "as a destructured value"
		}
		return keeps(l.checker.GetTypeAtLocation(name), "a binding's initializer")
	case ast.KindReturnStatement, ast.KindArrowFunction:
		function := parent
		if parent.Kind == ast.KindReturnStatement {
			function = ast.GetContainingFunction(parent)
		} else if parent.Body() != site {
			return "as a value"
		}
		if function == nil || !ast.IsFunctionLike(function) {
			return "as a returned value"
		}
		signature := l.checker.GetSignatureFromDeclaration(function)
		if signature == nil {
			return "as a returned value"
		}
		returns := l.checker.GetReturnTypeOfSignature(signature)
		if returns.Flags() == checker.TypeFlagsVoid {
			return "" // #se1zcgr's discard: returned from a function that returns exactly void.
		}
		if ast.HasSyntacticModifier(function, ast.ModifierFlagsAsync) && keeps(checker.Checker_getAwaitedType(l.checker, returns), "") == "" {
			return ""
		}
		return keeps(returns, "a returned value")
	case ast.KindCallExpression, ast.KindNewExpression:
		if parent.Kind == ast.KindCallExpression && parent.AsCallExpression().Expression == site {
			break
		}
		return keeps(l.argumentSlot(parent, site), "an argument")
	case ast.KindTemplateSpan:
		return "as a template literal part"
	case ast.KindTypeOfExpression:
		return "as typeof's operand"
	case ast.KindSpreadElement, ast.KindSpreadAssignment:
		return "as a spread"
	}
	return "as a value"
}

// genericVoidBinding reads what each of a class's type parameters is in classType, for
// refuseGenericVoidReads at the class's instantiation.
func (l *lowering) genericVoidBinding(declaration *ast.Node, typeArguments []*checker.Type) map[*checker.Type]*checker.Type {
	bound := map[*checker.Type]*checker.Type{}
	for index, parameter := range declaration.TypeParameters() {
		if index < len(typeArguments) {
			bound[l.checker.GetTypeAtLocation(parameter.Name())] = typeArguments[index]
		}
	}
	return bound
}

// voidAnywhere is whether void appears anywhere in t: at a position holdsVoid walks, and also in a
// signature's parameters and result, where holdsVoid doesn't look. Only the unmapped fallback asks:
// `{ make: () => T }` resolved to `{ make: () => void }` hands the body a T through a call (Lower, on
// #p7t8tmh), and a parameter that can't be read back is judged as void when this says so.
func (l *lowering) voidAnywhere(t *checker.Type, depth int, seen map[*checker.Type]bool) bool {
	if t == nil || seen[t] {
		return false
	}
	if t.Flags() == checker.TypeFlagsVoid || depth >= 32 {
		return true // Past the depth bound it fails closed.
	}
	seen[t] = true
	if t.Flags()&(checker.TypeFlagsUnion|checker.TypeFlagsIntersection) != 0 {
		for _, member := range t.Types() {
			if l.voidAnywhere(member, depth+1, seen) {
				return true
			}
		}
		return false
	}
	if t.Flags()&checker.TypeFlagsObject == 0 {
		return false
	}
	if l.holdsVoid(t) {
		return true
	}
	if t.ObjectFlags()&checker.ObjectFlagsReference != 0 && t.Target() != nil {
		for _, argument := range l.checker.GetTypeArguments(t) {
			if l.voidAnywhere(argument, depth+1, seen) {
				return true
			}
		}
	}
	for _, kind := range []checker.SignatureKind{checker.SignatureKindCall, checker.SignatureKindConstruct} {
		for _, signature := range l.checker.GetSignaturesOfType(t, kind) {
			if l.voidAnywhere(l.checker.GetReturnTypeOfSignature(signature), depth+1, seen) {
				return true
			}
			for _, parameter := range signature.Parameters() {
				if l.voidAnywhere(l.checker.GetTypeOfSymbol(parameter), depth+1, seen) {
					return true
				}
			}
		}
	}
	for _, property := range l.checker.GetPropertiesOfType(t) {
		if l.voidAnywhere(l.checker.GetTypeOfSymbol(property), depth+1, seen) {
			return true
		}
	}
	return false
}
