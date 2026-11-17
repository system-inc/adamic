package lower

import (
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// CheckWorkerEnvironment checks the bridge's projected environment, after the Worker
// signature is checked and before lowering or writing output. All imported modules
// share the entry interface's symbol. No checker fact escapes into IR or a cache.
func CheckWorkerEnvironment(program *load.Program, types *checker.Checker, environment *checker.Type, capabilitySites ...map[ir.SourcePosition]bool) error {
	// The checked Worker contract owns discovery. Never rediscover it here: a
	// missed export must not silently disable the environment boundary.
	if environment == nil {
		return fmt.Errorf("worker: compiler bug: two-parameter handle has no checked environment type")
	}
	return checkWorkerBindings(program, types, environment, capabilitySites...)
}

// One shared walk, with the environment absent for a one-parameter handler.
func CheckWorkerCapabilities(program *load.Program, types *checker.Checker, sites map[ir.SourcePosition]bool) error {
	return checkWorkerBindings(program, types, nil, sites)
}
func checkWorkerBindings(program *load.Program, types *checker.Checker, environment *checker.Type, capabilitySites ...map[ir.SourcePosition]bool) error {
	entry := program.Files()[0]
	l := &lowering{program: program, checker: types}
	rules := workerEnvironmentRules{l: l, environment: environment}
	if len(capabilitySites) != 0 {
		rules.capabilitySites = capabilitySites[0]
	}
	modules, _ := ESMModuleOrder(types, entry)
	for _, module := range modules {
		var found error
		var visit ast.Visitor
		visit = func(node *ast.Node) bool {
			found = rules.workerEnvironmentCheck(node)
			if found != nil {
				return true
			}
			return node.ForEachChild(visit)
		}
		module.AsNode().ForEachChild(visit)
		if found != nil {
			return found
		}
	}
	return nil
}

type workerEnvironmentRules struct {
	l               *lowering
	environment     *checker.Type
	exhausted       bool
	capabilitySites map[ir.SourcePosition]bool
}

func (r *workerEnvironmentRules) exact(t *checker.Type) bool {
	return r.environment != nil && t != nil && t.Symbol() != nil && t.Symbol() == r.environment.Symbol() && t.Flags()&checker.TypeFlagsObject != 0
}

// A predicate or assertion can narrow the environment to an intersection without
// its own symbol. Such a value still observes the projected environment. Unions
// are deliberately excluded here: the declaration slot rule refuses them.
func (r *workerEnvironmentRules) environmentValue(t *checker.Type) bool {
	if r.exact(t) {
		return true
	}
	if t != nil && t.Flags()&checker.TypeFlagsIntersection != 0 {
		for _, part := range t.Types() {
			if r.environmentValue(part) {
				return true
			}
		}
	}
	return false
}

// Lower's #bxaz0p1 slot rule: the environment interface is admitted only as the
// handler's parameter, a const local or a parameter of that exact interface.
// A class field or a this holding it refuses. Walk every type position, not names
// of bindings: aliases, unions, containers and callable views cannot hide it.
func (r *workerEnvironmentRules) holds(t *checker.Type, seen map[*checker.Type]bool, depth int, kvOnly ...bool) bool {
	if t == nil || seen[t] {
		return false
	}
	if (len(kvOnly) == 0 && r.exact(t)) || r.kv(t) {
		return true
	}
	seen[t] = true
	if depth >= 64 {
		r.exhausted = true
		return true
	}
	if alias := t.Alias(); alias != nil {
		for _, argument := range alias.TypeArguments() {
			if r.holds(argument, seen, depth+1, kvOnly...) {
				return true
			}
		}
	}
	var next []*checker.Type
	switch {
	case t.Flags()&(checker.TypeFlagsUnion|checker.TypeFlagsIntersection) != 0:
		next = t.Types()
	case t.Flags()&checker.TypeFlagsInstantiableNonPrimitive != 0:
		next = append(next, r.l.checker.GetBaseConstraintOfType(t))
	case t.Flags()&checker.TypeFlagsObject != 0:
		if t.ObjectFlags()&checker.ObjectFlagsReference != 0 {
			next = append(next, r.l.checker.GetTypeArguments(t)...)
		}
		// As the host-view walk does, inspect a library container's payloads,
		// not Promise.then's recursively instantiated library signatures.
		if !r.l.nodeLibraryContainer(t) {
			for _, field := range r.l.checker.GetPropertiesOfType(t) {
				next = append(next, r.l.checker.GetTypeOfSymbol(field))
			}
			for _, index := range r.l.checker.GetIndexInfosOfType(t) {
				next = append(next, index.ValueType())
			}
			for _, kind := range []checker.SignatureKind{checker.SignatureKindCall, checker.SignatureKindConstruct} {
				for _, signature := range r.l.checker.GetSignaturesOfType(t, kind) {
					next = append(next, r.l.checker.GetReturnTypeOfSignature(signature))
					for _, parameter := range signature.Parameters() {
						next = append(next, r.l.checker.GetTypeOfSymbol(parameter))
					}
				}
			}
		}
	}
	for _, inside := range next {
		if r.holds(inside, seen, depth+1, kvOnly...) {
			return true
		}
	}
	return false
}

func (r *workerEnvironmentRules) refused(node *ast.Node, form string) error {
	return &Refused{Where: r.l.program.Where(node), What: "a worker environment in " + form,
		Fix: "keep the exact environment interface in a const local or parameter and read its declared fields"}
}

// Adamic, #bxaz0p1, 14:45Z: these are valid programs Workers has an answer
// for. Shape observations and widenings are NotYet, never Refused.
const workerEnvironmentLift = "; the environment holds only the declared bindings here; observing its shape waits on a bridge that passes Workers' own env"

func (r *workerEnvironmentRules) observation(node *ast.Node, form string) error {
	return r.l.notYet(node, "worker environment "+form+workerEnvironmentLift)
}

func workerEnvironmentFunction(node *ast.Node) *ast.Node {
	for node = node.Parent; node != nil; node = node.Parent {
		if ast.IsFunctionLike(node) {
			return node
		}
	}
	return nil
}

func (r *workerEnvironmentRules) slot(node *ast.Node) error {
	var t *checker.Type
	allowed := false
	if (node.Kind == ast.KindVariableDeclaration || node.Kind == ast.KindParameter) && node.Name().Kind == ast.KindObjectBindingPattern && r.exact(r.l.checker.GetTypeAtLocation(node)) {
		for _, element := range node.Name().AsBindingPattern().Elements.Nodes {
			if element.AsBindingElement().DotDotDotToken != nil {
				return r.observation(element, "spread")
			}
		}
	}
	switch node.Kind {
	case ast.KindVariableDeclaration:
		if node.Name().Kind != ast.KindIdentifier {
			return nil
		} // Destructuring reads fields; it holds no environment slot.
		t = r.l.checker.GetTypeAtLocation(node.Name())
		allowed = node.Parent.Flags&ast.NodeFlagsConst != 0 && workerEnvironmentFunction(node) != nil && r.exact(t)
	case ast.KindParameter:
		t = r.l.checker.GetTypeAtLocation(node)
		allowed = r.exact(t) && !ast.IsThisIdentifier(node.Name()) && node.AsParameterDeclaration().QuestionToken == nil && node.AsParameterDeclaration().DotDotDotToken == nil && !ast.HasSyntacticModifier(node, ast.ModifierFlagsParameterPropertyModifier)
	case ast.KindPropertyDeclaration, ast.KindPropertySignature, ast.KindBindingElement:
		t = r.l.checker.GetTypeAtLocation(node)
	case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction, ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindCallSignature, ast.KindConstructSignature, ast.KindFunctionType, ast.KindConstructorType:
		if signature := r.l.checker.GetSignatureFromDeclaration(node); signature != nil {
			t = r.l.checker.GetReturnTypeOfSignature(signature)
		}
	case ast.KindThisKeyword:
		t = r.l.checker.GetTypeAtLocation(node)
	}
	// Lower #bxaz0p1 rule 1: the environment remains its designated container;
	// only its declared KV field, an exact const local or an exact parameter adds
	// another KV slot. The same holds walk judges both symbols.
	allowed = allowed || r.kvSlot(node, t)
	r.noteCapability(node, t)
	r.exhausted = false
	if t != nil && !allowed && r.holds(t, map[*checker.Type]bool{}, 0) {
		if r.exhausted {
			return r.l.notYet(node, "proving a worker environment slot free of its interface beyond 64 type positions")
		}
		if r.holds(t, map[*checker.Type]bool{}, 0, true) {
			return r.kvRefused(node, node.Kind.String()+" slot")
		}
		return r.refused(node, node.Kind.String()+" slot")
	}
	return nil
}

func (r *workerEnvironmentRules) parameterField(node *ast.Node) bool {
	if node.Kind != ast.KindIdentifier {
		return false
	}
	symbol := r.l.checker.GetSymbolAtLocation(node)
	if symbol == nil {
		return false
	}
	for _, d := range symbol.Declarations {
		if d.Kind != ast.KindBindingElement {
			continue
		}
		for parent := d.Parent; parent != nil; parent = parent.Parent {
			if parent.Kind == ast.KindParameter {
				return r.exact(r.l.checker.GetTypeAtLocation(parent))
			}
			if ast.IsFunctionLike(parent) {
				break
			}
		}
	}
	return false
}

// CheckSignature admits only required readonly data properties of type exactly
// string or the registered KV namespace. A declared name can therefore never be a method such as hasOwnProperty
// or a constructor: Function field (Language's e03/e04).
func (r *workerEnvironmentRules) declared(access *ast.Node) bool {
	var name string
	switch access.Kind {
	case ast.KindPropertyAccessExpression:
		name = access.Name().Text()
	case ast.KindElementAccessExpression:
		key := ast.SkipParentheses(access.AsElementAccessExpression().ArgumentExpression)
		if key.Kind != ast.KindStringLiteral {
			return false
		}
		name = key.Text()
	default:
		return false
	}
	for _, field := range r.l.checker.GetPropertiesOfType(r.environment) {
		if field.Name == name {
			return true
		}
	}
	return false
}

func workerEnvironmentUse(node *ast.Node) string {
	switch node.Kind {
	case ast.KindCallExpression:
		callee := ast.SkipParentheses(node.AsCallExpression().Expression)
		if callee.Kind == ast.KindPropertyAccessExpression {
			if receiver := callee.Expression(); receiver.Kind == ast.KindIdentifier {
				return receiver.Text() + "." + callee.Name().Text()
			}
			return "call to ." + callee.Name().Text()
		}
		if callee.Kind == ast.KindIdentifier {
			return "argument to " + callee.Text()
		}
	case ast.KindBinaryExpression:
		if node.AsBinaryExpression().OperatorToken.Kind == ast.KindInKeyword {
			return "in"
		}
	case ast.KindForInStatement:
		return "for...in"
	case ast.KindSpreadAssignment, ast.KindSpreadElement:
		return "spread"
	}
	return node.Kind.String()
}

func (r *workerEnvironmentRules) workerEnvironmentCheck(node *ast.Node) error {
	if err := r.slot(node); err != nil {
		return err
	}
	if err := r.kvUse(node); err != nil {
		return err
	}
	if ast.IsAssignmentTarget(node) && r.parameterField(node) {
		return r.refused(node, "a write to a destructured parameter field")
	}
	if node.Kind == ast.KindParenthesizedExpression || !r.l.isExpression(node) || ast.IsDeclarationName(node) || hasTypeAncestor(node) {
		return nil
	}
	if node.Parent != nil && node.Parent.Kind == ast.KindPropertyAccessExpression && node.Parent.Name() == node {
		return nil
	}
	name, source, target := r.l.relationSite(node)
	if r.environmentValue(source) && target != nil && !r.exact(target) {
		return r.observation(node, "widening at "+name+" to "+r.l.checker.TypeToString(target)+" ("+workerEnvironmentUse(node.Parent)+")")
	}
	// Literal elements and fields are slots too, even with no binding declaration.
	if r.exact(source) && (name == "array element" || name == "object field") {
		return r.refused(node, name+" slot")
	}
	if !r.environmentValue(r.l.checker.GetTypeAtLocation(node)) {
		return nil
	}
	site := node
	for site.Parent != nil && site.Parent.Kind == ast.KindParenthesizedExpression {
		site = site.Parent
	}
	parent := site.Parent
	if parent == nil {
		return r.observation(node, "value")
	}
	if (parent.Kind == ast.KindPropertyAccessExpression || parent.Kind == ast.KindElementAccessExpression) && parent.Expression() == site {
		if r.declared(parent) {
			if ast.IsAssignmentTarget(parent) {
				return r.refused(parent, "a write to a declared field")
			}
			// Lower, 14:49Z: ?. is admitted on a non-optional environment. The
			// optional slot itself is refused above, before a narrowed read can hide it.
			return nil
		}
		if parent.Kind == ast.KindPropertyAccessExpression {
			return r.observation(node, "member "+parent.Name().Text())
		}
		return r.observation(node, "computed field read")
	}
	if target != nil && r.exact(target) {
		if parent.Kind == ast.KindCallExpression || parent.Kind == ast.KindNewExpression {
			// The value's slot was read through relationSite above (including
			// slotSet). This lookup finds the original declaration's nominal
			// contract only; an instantiated T is not that declared interface.
			signature := r.l.checker.GetResolvedSignature(parent)
			if signature != nil {
				parameters := signature.Parameters()
				var arguments []*ast.Node
				if parent.Kind == ast.KindCallExpression {
					arguments = parent.AsCallExpression().Arguments.Nodes
				} else if parent.AsNewExpression().Arguments != nil {
					arguments = parent.AsNewExpression().Arguments.Nodes
				}
				for i, argument := range arguments {
					if argument != site || len(parameters) == 0 {
						continue
					}
					parameter := parameters[min(i, len(parameters)-1)]
					if len(parameter.Declarations) == 0 {
						return r.observation(node, "an argument whose original parameter cannot be read")
					}
					for _, declaration := range parameter.Declarations {
						if declaration.Kind == ast.KindParameter && !r.exact(r.l.checker.GetTypeAtLocation(declaration)) {
							return r.observation(node, "widening through a generic argument")
						}
					}
				}
			}
		}
		return nil
	}
	// A destructuring binding's initializer is a field read, as is a handler's
	// parameter pattern. Rest elements observe shape and are not destructuring reads.
	if parent.Kind == ast.KindVariableDeclaration && parent.Name().Kind == ast.KindObjectBindingPattern {
		for _, element := range parent.Name().AsBindingPattern().Elements.Nodes {
			if element.AsBindingElement().DotDotDotToken != nil {
				return r.observation(element, "spread")
			}
		}
		return nil
	}
	return r.observation(node, workerEnvironmentUse(parent))
}
