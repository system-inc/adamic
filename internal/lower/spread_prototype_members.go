package lower

import (
	"slices"
	"sort"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// A spread copies only own enumerable data, and a class's getter, setter or method lives on its
// prototype, so { ...instance } never holds one. The checker still types the copy with such a member
// when the source is seen through a type that declares it in any form, a field, a method signature, a
// getter or setter signature or an object literal's accessor, or through a type parameter, which types
// the copy as what it stands for, every member included. So every prototype member the copy's type still
// claims is a lie about the copy, whatever form the claim takes, and by Adamic's amendment of ruling 91
// (#vjvcnnj, her ruling (1)) an Adamic class stays admitted unless its prototype holds a member the
// spread's copy is typed with, and such a spread is refused in .a, naming the class, the member and the
// fix. A method or accessor a class itself declares is one the checker leaves out of a spread's type
// (isSpreadableProperty), so a view that is a class type claims none of those. In .ts the spread stays
// and a read of the missing member stops instead (spread_copy_reads.go).
//
// It runs before finishAccessors, whose stop at such a read (spreadCopyRead) would otherwise answer
// first for a program the language refuses at the spread.

// prototypeMember is a member a class reaching a spread holds on its prototype, under a name the
// spread's copy is typed with.
type prototypeMember struct {
	// class reaches the spread; holder declares the member, the class itself or one of its bases.
	class, holder, name, kind string
	// claim is how the copy's type comes to hold name, said after "whose type".
	claim string
}

// viewPart is one constituent of a spread's view, as the instantiation being lowered makes it, with
// the names a copy of it is typed with.
type viewPart struct {
	part  *checker.Type
	names []string
	// claim is prototypeMember's, for a name of this part.
	claim func(name string) string
}

// checkSpreadPrototypeMembers refuses, in .a, each object spread whose copy is typed with a member that a
// class reaching it holds on its prototype (spreadViewParts). The closed world is every program class
// instance an expression in the modules has, as librarySpreadCandidates finds library types. A spread
// whose producers the allocation graph proves are all classless literals, accessor literals or
// instances of classes holding none of the copy's names on their prototypes reaches no such class.
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
		parts := l.spreadViewParts(spread.written, spread.mapper)
		member, found := l.spreadPrototypeMember(spread.view, parts, classes)
		if !found {
			continue
		}
		// Built only once a type might refuse, so a program with no such spread builds no graph this
		// early. An Adamic class instance is proof here only when its chain holds none of the copy's
		// names on its prototype, unlike the library rule's, where every one is.
		if graph == nil {
			graph = l.flowGraph(l.result)
		}
		fields := map[string]bool{}
		for _, part := range parts {
			for _, name := range part.names {
				fields[name] = true
			}
		}
		if graph.plainSpreadSources(spread.source, map[int]bool{}, graph.prototypeFree(fields)) {
			continue
		}
		inherited := ""
		if member.holder != member.class {
			inherited = ", inherited from " + member.holder
		}
		return &Refused{
			Where: l.program.Where(spread.node),
			What:  "a spread of " + sourceExpression(spread.node.AsSpreadAssignment().Expression) + ", whose type " + member.claim + ", while class " + member.class + ", which may reach it, holds " + member.name + " as a " + member.kind + " on its prototype" + inherited + ": a spread copies own properties only, so the copy has no " + member.name + " though its type says it does",
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

// spreadViewParts is what a spread's copy is typed with, constituent by constituent of its view as
// the source writes it (written), each part as the instantiation being lowered makes it. A union view
// types the copy with any member's names, which a narrowing reaches, so each member is a part of its
// own: the union's properties would be only the names every member has. undefined and null add
// nothing to a copy. A union of one type with undefined or null spreads as that type made partial (the
// checker's tryMergeUnionOfObjectTypeAndEmptyObject), so its copy claims no name present and has no
// part. A type parameter types the copy as itself, so every member of what it stands for is claimed,
// a method and an accessor included; any other part claims the names its type declares in any form the
// copy keeps (copyClaim).
// mapper is the spread's instantiation's, as librarySpread keeps it.
func (l *lowering) spreadViewParts(written *checker.Type, mapper *typeMapper) []viewPart {
	if written == nil {
		return nil
	}
	concrete := func(proven *checker.Type) *checker.Type {
		if mapper == nil {
			return proven
		}
		return instantiateType(l.checker, proven, mapper)
	}
	if written.Flags()&checker.TypeFlagsUnion != 0 {
		nullish, others := false, 0
		for _, member := range written.Types() {
			if member.Flags()&(checker.TypeFlagsUndefined|checker.TypeFlagsNull) != 0 {
				nullish = true
			} else {
				others++
			}
		}
		if nullish && others == 1 {
			return nil
		}
	}
	var parts []viewPart
	var walk func(*checker.Type)
	walk = func(written *checker.Type) {
		if written.Flags()&(checker.TypeFlagsUndefined|checker.TypeFlagsNull) != 0 {
			return
		}
		if written.Flags()&(checker.TypeFlagsUnion|checker.TypeFlagsIntersection) != 0 {
			for _, member := range written.Types() {
				walk(member)
			}
			return
		}
		if written.Flags()&checker.TypeFlagsTypeParameter != 0 {
			parameter := l.checker.TypeToString(written)
			stands := []*checker.Type{concrete(written)}
			if stands[0].Flags()&checker.TypeFlagsUnion != 0 {
				stands = stands[0].Types()
			}
			for _, standing := range stands {
				if standing.Flags()&(checker.TypeFlagsUndefined|checker.TypeFlagsNull) != 0 {
					continue
				}
				var names []string
				for _, property := range l.checker.GetPropertiesOfType(standing) {
					names = append(names, property.Name)
				}
				shown := l.checker.TypeToString(standing)
				parts = append(parts, viewPart{part: standing, names: names, claim: func(name string) string {
					return "is " + parameter + ", which stands for " + shown + " here, so the copy is typed with " + name
				}})
			}
			return
		}
		instantiated := concrete(written)
		var names []string
		forms := map[string]string{}
		for _, property := range l.checker.GetPropertiesOfType(instantiated) {
			if form, claimed := copyClaim(property); claimed {
				names = append(names, property.Name)
				forms[property.Name] = form
			}
		}
		parts = append(parts, viewPart{part: instantiated, names: names, claim: func(name string) string {
			return "declares " + name + " " + forms[name]
		}})
	}
	walk(written)
	return parts
}

// spreadPrototypeMember is the first, by class then member name, of the prototype members a candidate
// class assignable to view, and to the part that names it, holds under a name the copy is typed with.
func (l *lowering) spreadPrototypeMember(view *checker.Type, parts []viewPart, classes map[*checker.Type]string) (prototypeMember, bool) {
	var found []prototypeMember
	for candidate, class := range classes {
		if !l.checker.IsTypeAssignableTo(candidate, view) {
			continue
		}
		for _, part := range parts {
			if len(part.names) == 0 || !l.checker.IsTypeAssignableTo(candidate, part.part) {
				continue
			}
			for _, name := range part.names {
				holder, kind := prototypeHeld(l.checker.GetPropertyOfType(candidate, name))
				if holder == nil {
					continue
				}
				member := prototypeMember{class: class, holder: class, name: name, kind: kind, claim: part.claim(name)}
				if !slices.Contains(candidate.Symbol().Declarations, holder) {
					member.holder = "an anonymous class"
					if holderName := holder.Name(); holderName != nil {
						member.holder = holderName.Text()
					}
				}
				found = append(found, member)
			}
		}
	}
	if len(found) == 0 {
		return prototypeMember{}, false
	}
	// Stable, so of one class and name met through several parts the first part's claim is said.
	sort.SliceStable(found, func(i, j int) bool {
		if found[i].class != found[j].class {
			return found[i].class < found[j].class
		}
		return found[i].name < found[j].name
	})
	return found[0], true
}

// copyClaim is the form in which a view's property types a spread's copy with it, as the refusal says
// it: a field (a property signature or declaration, an object literal's property, a parameter property,
// or a mapped type's property with no declaration at all), a method (a method signature, or an object
// literal type's method), a getter, a setter, or a getter and setter. Every form counts, by Adamic's
// ruling (1). It claims nothing only for a method or an accessor a class declares, which the checker
// leaves out of a spread's type. An intersection's property takes any member's form; a union's members
// are each judged on their own (spreadViewParts).
func copyClaim(property *ast.Symbol) (string, bool) {
	if len(property.Declarations) == 0 {
		return "a field", true
	}
	field, method, getter, setter := false, false, false, false
	for _, declaration := range property.Declarations {
		switch declaration.Kind {
		case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
			if declaration.Parent != nil && ast.IsClassLike(declaration.Parent) {
				continue
			}
		}
		switch declaration.Kind {
		case ast.KindMethodDeclaration, ast.KindMethodSignature:
			method = true
		case ast.KindGetAccessor:
			getter = true
		case ast.KindSetAccessor:
			setter = true
		default:
			field = true
		}
	}
	if field {
		return "a field", true
	}
	switch {
	case getter && setter:
		return "a getter and setter", true
	case getter:
		return "a getter", true
	case setter:
		return "a setter", true
	case method:
		return "a method", true
	}
	return "", false
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
