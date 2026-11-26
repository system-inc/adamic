package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// A spread copies own properties only, so { ...instance } has none of its class's methods, getters or
// setters, which are its prototype's. Written on a value of a class type the checker types the copy
// without them, but a spread of a type parameter's value is typed as that type parameter, so once a
// generic function is instantiated with a class the copy is typed as the class, prototype members and all
// (#y4gjeb8): function copy<T extends Weighs>(source: T): T { return { ...source }; } called with
// new Scale() gives a Scale with no weigh, where Node throws at copy(...).weigh() and native called
// Scale's weigh, typeof read function where Node reads undefined, and a setter's name read stopped on a
// compiler bug. At the read nothing is generic, the receiver is plain Scale, so the stop keys on reach, as
// spreadCopyRead does for a getter (spread_copy_reads.go): a read or call of a prototype member of a
// program class stops where its receiver's allocations may hold a generic spread's copy whose source
// class holds that member on its prototype, and where the graph can't close the reach while any generic
// spread in the program may copy such a class, failing closed. A receiver that is this is never judged:
// a this inside a prototype member is reached only through a read of that member, which is. A program
// with no generic spread of a class, or one whose spreads copy only plain objects, lowers as on main.
// This is Runtime's floor; the real fix, a copy with no prototype answering as Node's does, is Lower's.
//
// A spread needn't be generic to hide a class (#r8mazrh): a value typed by an interface or an object type
// may be a program class's instance, and the spread's copy is typed with every member that type declares,
// methods included. function copy(source: Weighs): Weighs { return { ...source }; } over new Scale()
// printed 27 natively where Node throws at .weigh(), since native reuse kept the source object and its
// methods. So such a spread counts as one too, its members those of each program class with a prototype
// member assignable to its source's type, and a setter's write is judged as a read is.

// prototypeMember is a class's method, getter or setter, which lives on its prototype: the class's name
// and which of the three it is, as a stop names them, and copy, how the stop names the spread that may
// have lost it.
type prototypeMember struct {
	class, kind, copy string
}

// prototypeMemberUse is a read lowering made of a program class's method on a receiver other than this
// that leaves no ir.Property behind: a direct call (callOrMethod) or a presence test (methodPresence).
type prototypeMemberUse struct {
	node     *ast.Node
	name     string
	receiver ir.Expression
}

// genericSpreadCopy and hiddenClassSpreadCopy are how a stop names a spread's copy that may lack its
// source's prototype members: a generic spread's, typed by its instantiation, and one whose source's type
// may hide a class instance (#r8mazrh).
const (
	genericSpreadCopy     = "a generic spread's copy"
	hiddenClassSpreadCopy = "a spread's copy, through a type that doesn't name its class,"
)

// noteGenericSpread records a spread's literal, with how a stop names its copy, when the checker types its
// source, before the instantiation's types replace its type parameters, with a type parameter or a type
// that names one, or when the source's type may hide a program class's instance (mayHideClassInstance).
func (l *lowering) noteGenericSpread(source *ast.Node, site int) {
	written := l.checker.GetTypeAtLocation(source)
	copy := ""
	switch {
	case l.mentionsTypeParameter(written, map[*checker.Type]bool{}):
		copy = genericSpreadCopy
	case l.mayHideClassInstance(l.concrete(written)):
		copy = hiddenClassSpreadCopy
	default:
		return
	}
	if l.genericSpreadSites == nil {
		l.genericSpreadSites = map[int]string{}
	}
	l.genericSpreadSites[site] = copy
}

// mayHideClassInstance is whether a value of type of may be a program class's instance that of doesn't
// name as one: of, or a member of its union, is an object or intersection type that isn't a class
// instance, and some program class with a method, getter or setter on its prototype is assignable to it.
func (l *lowering) mayHideClassInstance(of *checker.Type) bool {
	if of == nil {
		return false
	}
	if of.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range of.Types() {
			if l.mayHideClassInstance(member) {
				return true
			}
		}
		return false
	}
	if of.Flags()&(checker.TypeFlagsObject|checker.TypeFlagsIntersection) == 0 || isClassInstance(of) {
		return false
	}
	for _, class := range l.declaredClasses() {
		members := map[string]prototypeMember{}
		l.prototypeMembers(class, members)
		if len(members) != 0 && l.checker.IsTypeAssignableTo(class, of) {
			return true
		}
	}
	return false
}

// hiddenClassMembers adds to members the prototype members a copy of a value of type of may lack: a
// class instance's own, as prototypeMembers reads them, and for any other object type every program
// class assignable to it, each union member's included.
func (l *lowering) hiddenClassMembers(of *checker.Type, members map[string]prototypeMember) {
	if of.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range of.Types() {
			l.hiddenClassMembers(member, members)
		}
		return
	}
	if isClassInstance(of) {
		l.prototypeMembers(of, members)
		return
	}
	for _, class := range l.declaredClasses() {
		if l.checker.IsTypeAssignableTo(class, of) {
			l.prototypeMembers(class, members)
		}
	}
}

// notePrototypeMemberUse records a class method's call or presence test for checkGenericSpreadCopyReads.
func (l *lowering) notePrototypeMemberUse(access *ast.Node, receiver ir.Expression) {
	written := ast.SkipParentheses(access.AsPropertyAccessExpression().Expression)
	if written.Kind == ast.KindThisKeyword || written.Kind == ast.KindSuperKeyword {
		return
	}
	l.prototypeMemberUses = append(l.prototypeMemberUses, prototypeMemberUse{node: access, name: l.fieldName(access.Name()), receiver: receiver})
}

// genericSpreadMembers is every prototype member a generic spread's copy may lack though its type says
// it has it: the source's class's methods, getters and setters, as the instantiation types the source.
// A source the instantiation left generic, a polymorphic this among them, may be any program class
// assignable to its constraint, so each such class's count, failing closed.
func (l *lowering) genericSpreadMembers(literal ir.ObjectLiteral) map[string]prototypeMember {
	if literal.Spread == nil || l.genericSpreadSites[literal.SpreadSite] == "" || literal.SpreadSite > len(l.spreadSources) {
		return nil
	}
	members := map[string]prototypeMember{}
	source := l.spreadSources[literal.SpreadSite-1]
	if source == nil || l.mentionsTypeParameter(source, map[*checker.Type]bool{}) {
		var constraint *checker.Type
		if source != nil {
			constraint = l.checker.GetBaseConstraintOfType(source)
		}
		for _, class := range l.declaredClasses() {
			if constraint == nil || l.checker.IsTypeAssignableTo(class, constraint) {
				l.prototypeMembers(class, members)
			}
		}
		return members
	}
	l.hiddenClassMembers(source, members)
	return members
}

// prototypeMembers adds to members each method, getter and setter a program class declares, not
// static, that type holds, each union member's included.
func (l *lowering) prototypeMembers(of *checker.Type, members map[string]prototypeMember) {
	if of.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range of.Types() {
			l.prototypeMembers(member, members)
		}
		return
	}
	for _, property := range l.checker.GetPropertiesOfType(of) {
		for _, root := range l.checker.GetRootSymbols(property) {
			for _, declaration := range root.Declarations {
				kind := ""
				switch declaration.Kind {
				case ast.KindMethodDeclaration:
					kind = "method"
				case ast.KindGetAccessor:
					kind = "getter"
				case ast.KindSetAccessor:
					kind = "setter"
				default:
					continue
				}
				class := declaration.Parent
				if class == nil || (class.Kind != ast.KindClassDeclaration && class.Kind != ast.KindClassExpression) || ast.HasSyntacticModifier(declaration, ast.ModifierFlagsStatic) {
					continue
				}
				file := ast.GetSourceFileOfNode(declaration)
				if load.IsLibrary(file) || load.IsNodeLibrary(file) || load.IsPrelude(file) {
					continue
				}
				name := "class expression"
				if class.Name() != nil {
					name = class.Name().Text()
				}
				// A getter names an accessor pair, since a read meets it first.
				if known, seen := members[root.Name]; !seen || (known.kind == "setter" && kind == "getter") {
					members[root.Name] = prototypeMember{class: name, kind: kind}
				}
			}
		}
	}
}

// checkGenericSpreadCopyReads stops a read or call of a prototype member whose receiver may be a
// generic spread's copy without it, as the comment above says. It runs before finishAccessors, so a
// getter's or setter's read is still the ir.Property lowering made.
func (l *lowering) checkGenericSpreadCopyReads() error {
	if len(l.genericSpreadSites) == 0 {
		return nil
	}
	// A copy of the program numbers its allocation sites, as newAccessorReach's does, so the program
	// itself is left as lowering made it.
	numbered := *l.result
	numbered.Functions = append([]ir.Function(nil), l.result.Functions...)
	assignAllocationSites(&numbered)
	graph := l.flowGraph(&numbered)
	records := graph.projectionIndex().records
	// plainSource is whether a spread that isn't generic copies only plain object literals its program
	// built, by the graph: each allocation reaching its source a literal of no class that spreads nothing
	// itself, so no class instance's prototype was hidden by the source's type (#r8mazrh). A reach the
	// graph can't close is no proof. A generic spread is judged by its instantiation alone, as at e1b14c0e.
	plainSource := func(literal ir.ObjectLiteral) bool {
		if l.genericSpreadSites[literal.SpreadSite] == genericSpreadCopy {
			return false
		}
		reach := graph.ReachingAllocations(literal.Spread)
		if reach.Unknown || len(reach.Sites) == 0 {
			return false
		}
		for _, site := range reach.Sites {
			source, built := records[site]
			if !built || source.Class != 0 || source.Spread != nil {
				return false
			}
		}
		return true
	}
	copied := map[int]map[string]prototypeMember{}
	anywhere := map[string]prototypeMember{}
	collect := func(node any) bool {
		if literal, ok := node.(ir.ObjectLiteral); ok && literal.Spread != nil && l.genericSpreadSites[literal.SpreadSite] != "" {
			if _, known := copied[literal.SpreadSite]; !known {
				if plainSource(literal) {
					copied[literal.SpreadSite] = nil
					return true
				}
				copied[literal.SpreadSite] = l.genericSpreadMembers(literal)
				for name, member := range copied[literal.SpreadSite] {
					// A generic spread's entry wins a name both kinds hold, since a hidden class's leaves
					// a getter's read to spreadCopyRead (leftToSpreadCopyRead) and a generic spread's doesn't.
					if known, seen := anywhere[name]; !seen || known.copy == hiddenClassSpreadCopy {
						member.copy = l.genericSpreadSites[literal.SpreadSite]
						anywhere[name] = member
					}
				}
			}
		}
		return true
	}
	walk(numbered.Main, collect)
	for _, function := range numbered.Functions {
		walk(function.Body, collect)
	}
	if len(anywhere) == 0 {
		return nil
	}
	// leftToSpreadCopyRead is whether a use of member from a copy named copy is spreadCopyRead's to judge
	// rather than this pass's: a getter's read from a spread that isn't generic. spreadCopyRead, inside
	// finishAccessors, judges a getter read on any spread's copy by reach, so a hidden class's spread
	// leaves it there, and its stop's words are the ones those programs already pin (#7mjgxkc).
	leftToSpreadCopyRead := func(member prototypeMember, copy, use string) bool {
		return use == "a read" && member.kind == "getter" && copy == hiddenClassSpreadCopy
	}
	judge := func(where, use, name string, receiver ir.Expression) error {
		allocations := graph.ReachingAllocations(receiver)
		if allocations.Unknown {
			if member, ok := anywhere[name]; ok && !leftToSpreadCopyRead(member, member.copy, use) {
				return &NotYet{Where: where, What: use + " of " + name + " whose receiver may be " + member.copy + " of a " + member.class + " instance, whose " + name + " is a " + member.kind + " on its prototype: a spread copies own properties only, so the copy has no " + name + " though its type says it does (#y4gjeb8)"}
			}
			return nil
		}
		for _, site := range allocations.Sites {
			literal, ok := records[site]
			if !ok || literal.Spread == nil {
				continue
			}
			if member, left := copied[literal.SpreadSite][name]; left && l.genericSpreadSites[literal.SpreadSite] != "" && !leftToSpreadCopyRead(member, l.genericSpreadSites[literal.SpreadSite], use) {
				spread := literal.SpreadReadiness
				if spread == "" {
					spread = "an object spread"
				}
				return &NotYet{Where: where, What: use + " of " + name + " from " + spread + ", which may be " + l.genericSpreadSites[literal.SpreadSite] + " of a " + member.class + " instance, whose " + name + " is a " + member.kind + " on its prototype: a spread copies own properties only, so the copy has no " + name + " though its type says it does (#y4gjeb8)"}
			}
		}
		return nil
	}
	var failure error
	// self is the receiver parameter of the function being walked, -1 outside a method.
	self := -1
	inspect := func(node any) bool {
		if failure != nil {
			return false
		}
		// A setter's write is judged as a read is (#r8mazrh): a copy has no setter, so Node defines an own
		// field where lowering called the class's setter. A class field's define and a record's own-key
		// write are no setter's.
		use, name, receiver, where := "a read", "", ir.Expression(nil), ""
		switch access := node.(type) {
		case ir.Property:
			name, receiver, where = access.Name, access.Object, access.ViewWhere
		case ir.SetProperty:
			if access.Define || access.Record {
				return true
			}
			use, name, receiver = "a write", access.Name, access.Object
		default:
			return true
		}
		if _, member := anywhere[name]; !member {
			return true
		}
		if read, isRead := receiver.(ir.Read); isRead && read.Local == self {
			return true
		}
		if where == "" {
			where = l.result.Source
		}
		failure = judge(where, use, name, receiver)
		return failure == nil
	}
	walk(l.result.Main, inspect)
	for _, function := range l.result.Functions {
		self = -1
		if function.Receiver && len(function.Parameters) > 0 {
			self = function.Parameters[0]
		}
		walk(function.Body, inspect)
		if failure != nil {
			return failure
		}
	}
	if failure != nil {
		return failure
	}
	for _, use := range l.prototypeMemberUses {
		if l.firstLinkSkips(use.node) {
			continue
		}
		if err := judge(l.program.Where(use.node), "a read", use.name, use.receiver); err != nil {
			return err
		}
	}
	return nil
}
