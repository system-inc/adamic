package lower

import (
	"slices"
	"sort"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// A spread copies own enumerable properties only, and a class's getter, setter or method lives on its
// prototype, so { ...instance } never holds one. The checker still types the copy with such a member
// when the source is seen through a type that declares it as a field, and in .a a field typed number
// never holds undefined. So by Adamic's amendment of ruling 91 (#vjvcnnj) an Adamic class stays
// admitted unless its prototype holds a member that the spread's source type declares as a field, and
// such a spread is refused in .a, naming the class, the member and the fix. In .ts the spread stays and
// a read of the missing member stops instead (spread_copy_reads.go).
//
// It runs before finishAccessors, whose stop at such a read (spreadCopyRead) would otherwise answer
// first for a program the language refuses at the spread.

// prototypeMember is a member a class reaching a spread holds on its prototype, under a name the
// spread's view declares as a field.
type prototypeMember struct {
	// class reaches the spread; holder declares the member, the class itself or one of its bases.
	class, holder, name, kind string
}

// checkSpreadPrototypeMembers refuses, in .a, each object spread whose view declares as a field a
// member that a class reaching it holds on its prototype. The closed world is every program class
// instance an expression in the modules has, as librarySpreadCandidates finds library types. A spread
// whose producers the allocation graph proves are all classless literals reaches no class.
func (l *lowering) checkSpreadPrototypeMembers(modules []*ast.SourceFile) error {
	var classes map[*checker.Type]string
	var graph *allocationFlowGraph
	for _, spread := range l.librarySpreads {
		if l.firstLinkSkips(spread.node) || !strings.HasSuffix(l.program.FileName(ast.GetSourceFileOfNode(spread.node)), ".a") {
			continue
		}
		if classes == nil {
			classes = l.spreadClassCandidates(modules)
		}
		member, found := l.spreadPrototypeMember(spread.view, classes)
		if !found {
			continue
		}
		// Built only once a type might refuse, so a program with no such spread builds no graph this
		// early. An Adamic class instance is no proof here, unlike the library rule's.
		if graph == nil {
			graph = l.flowGraph(l.result)
		}
		if graph.plainSpreadSources(spread.source, map[int]bool{}, false) {
			continue
		}
		inherited := ""
		if member.holder != member.class {
			inherited = ", inherited from " + member.holder
		}
		return &Refused{
			Where: l.program.Where(spread.node),
			What:  "a spread of " + sourceExpression(spread.node.AsSpreadAssignment().Expression) + ", whose type declares " + member.name + " a field, while class " + member.class + ", which may reach it, holds " + member.name + " as a " + member.kind + " on its prototype" + inherited + ": a spread copies own properties only, so the copy has no " + member.name + " though its type says it does",
			Fix:   "copy the fields you mean explicitly, or declare " + member.name + " a field of " + member.holder,
		}
	}
	return nil
}

// spreadClassCandidates is every instance type of a class the program declares that some expression in
// the modules has, with its class's name. A library class is the library rule's (librarySpreadRefusal).
func (l *lowering) spreadClassCandidates(modules []*ast.SourceFile) map[*checker.Type]string {
	candidates := map[*checker.Type]string{}
	var note func(*checker.Type)
	note = func(proven *checker.Type) {
		if proven == nil {
			return
		}
		if proven.Flags()&(checker.TypeFlagsUnion|checker.TypeFlagsIntersection) != 0 {
			for _, member := range proven.Types() {
				note(member)
			}
			return
		}
		symbol := proven.Symbol()
		if !isClassInstance(proven) || symbol == nil || libraryDeclared(symbol) {
			return
		}
		candidates[proven] = symbol.Name
	}
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if ast.IsExpression(node) {
			note(l.checker.GetTypeAtLocation(node))
		}
		return node.ForEachChild(visit)
	}
	for _, module := range modules {
		module.AsNode().ForEachChild(visit)
	}
	return candidates
}

// spreadPrototypeMember is the first, by class then member name, of the prototype members a candidate
// class assignable to view holds under a name view declares as a field.
func (l *lowering) spreadPrototypeMember(view *checker.Type, classes map[*checker.Type]string) (prototypeMember, bool) {
	var fields []string
	for _, property := range l.checker.GetPropertiesOfType(view) {
		if fieldDeclared(property) {
			fields = append(fields, property.Name)
		}
	}
	if len(fields) == 0 {
		return prototypeMember{}, false
	}
	var found []prototypeMember
	for candidate, class := range classes {
		if !l.checker.IsTypeAssignableTo(candidate, view) {
			continue
		}
		for _, name := range fields {
			holder, kind := prototypeHeld(l.checker.GetPropertyOfType(candidate, name))
			if holder == nil {
				continue
			}
			member := prototypeMember{class: class, holder: class, name: name, kind: kind}
			if !slices.Contains(candidate.Symbol().Declarations, holder) {
				member.holder = "an anonymous class"
				if holderName := holder.Name(); holderName != nil {
					member.holder = holderName.Text()
				}
			}
			found = append(found, member)
		}
	}
	if len(found) == 0 {
		return prototypeMember{}, false
	}
	sort.Slice(found, func(i, j int) bool {
		if found[i].class != found[j].class {
			return found[i].class < found[j].class
		}
		return found[i].name < found[j].name
	})
	return found[0], true
}

// fieldDeclared reports whether a view's property is declared as a field: a property signature or
// declaration, an object literal's property, a parameter property, or a mapped type's property with no
// declaration at all, rather than a method or an accessor. A union's property is a field if any of its
// members declares it one, since the copy of that member is typed with it.
func fieldDeclared(property *ast.Symbol) bool {
	if len(property.Declarations) == 0 {
		return true
	}
	for _, declaration := range property.Declarations {
		switch declaration.Kind {
		case ast.KindMethodDeclaration, ast.KindMethodSignature, ast.KindGetAccessor, ast.KindSetAccessor:
		default:
			return true
		}
	}
	return false
}

// prototypeHeld reads a class instance's property as a member on a class's prototype: a getter, a
// setter or a method, with the class declaration holding it and the member's kind. An own data field,
// and a member no class declares, hold none.
func prototypeHeld(property *ast.Symbol) (*ast.Node, string) {
	if property == nil {
		return nil, ""
	}
	getter, setter, method := false, false, false
	var holder *ast.Node
	for _, declaration := range property.Declarations {
		if declaration.Parent == nil || !ast.IsClassLike(declaration.Parent) {
			continue
		}
		switch declaration.Kind {
		case ast.KindGetAccessor:
			getter = true
		case ast.KindSetAccessor:
			setter = true
		case ast.KindMethodDeclaration:
			method = true
		default:
			continue
		}
		if holder == nil {
			holder = declaration.Parent
		}
	}
	switch {
	case getter && setter:
		return holder, "getter and setter"
	case getter:
		return holder, "getter"
	case setter:
		return holder, "setter"
	case method:
		return holder, "method"
	}
	return nil, ""
}
