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

// prototypeMember is a class's method, getter or setter, which lives on its prototype: the class's name
// and which of the three it is, as a stop names them.
type prototypeMember struct {
	class, kind string
}

// prototypeMemberUse is a read lowering made of a program class's method on a receiver other than this
// that leaves no ir.Property behind: a direct call (callOrMethod) or a presence test (methodPresence).
type prototypeMemberUse struct {
	node     *ast.Node
	name     string
	receiver ir.Expression
}

// noteGenericSpread records a spread's literal as a generic spread when the checker types its source,
// before the instantiation's types replace its type parameters, with a type parameter or a type that
// names one.
func (l *lowering) noteGenericSpread(source *ast.Node, site int) {
	if !l.mentionsTypeParameter(l.checker.GetTypeAtLocation(source), map[*checker.Type]bool{}) {
		return
	}
	if l.genericSpreadSites == nil {
		l.genericSpreadSites = map[int]bool{}
	}
	l.genericSpreadSites[site] = true
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
	if literal.Spread == nil || !l.genericSpreadSites[literal.SpreadSite] || literal.SpreadSite > len(l.spreadSources) {
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
	l.prototypeMembers(source, members)
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
	copied := map[int]map[string]prototypeMember{}
	anywhere := map[string]prototypeMember{}
	collect := func(node any) bool {
		if literal, ok := node.(ir.ObjectLiteral); ok && literal.Spread != nil && l.genericSpreadSites[literal.SpreadSite] {
			if _, known := copied[literal.SpreadSite]; !known {
				copied[literal.SpreadSite] = l.genericSpreadMembers(literal)
				for name, member := range copied[literal.SpreadSite] {
					if _, known := anywhere[name]; !known {
						anywhere[name] = member
					}
				}
			}
		}
		return true
	}
	walk(l.result.Main, collect)
	for _, function := range l.result.Functions {
		walk(function.Body, collect)
	}
	if len(anywhere) == 0 {
		return nil
	}
	// A copy of the program numbers its allocation sites, as newAccessorReach's does, so the program
	// itself is left as lowering made it.
	numbered := *l.result
	numbered.Functions = append([]ir.Function(nil), l.result.Functions...)
	assignAllocationSites(&numbered)
	graph := l.flowGraph(&numbered)
	records := graph.projectionIndex().records
	judge := func(where, name string, receiver ir.Expression) error {
		allocations := graph.ReachingAllocations(receiver)
		if allocations.Unknown {
			if member, ok := anywhere[name]; ok {
				return &NotYet{Where: where, What: "a read of " + name + " whose receiver may be a generic spread's copy of a " + member.class + " instance, whose " + name + " is a " + member.kind + " on its prototype: a spread copies own properties only, so the copy has no " + name + " though its type says it does (#y4gjeb8)"}
			}
			return nil
		}
		for _, site := range allocations.Sites {
			literal, ok := records[site]
			if !ok || literal.Spread == nil {
				continue
			}
			if member, left := copied[literal.SpreadSite][name]; left && l.genericSpreadSites[literal.SpreadSite] {
				spread := literal.SpreadReadiness
				if spread == "" {
					spread = "an object spread"
				}
				return &NotYet{Where: where, What: "a read of " + name + " from " + spread + ", which may be a generic spread's copy of a " + member.class + " instance, whose " + name + " is a " + member.kind + " on its prototype: a spread copies own properties only, so the copy has no " + name + " though its type says it does (#y4gjeb8)"}
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
		read, ok := node.(ir.Property)
		if !ok {
			return true
		}
		if _, member := anywhere[read.Name]; !member {
			return true
		}
		if receiver, isRead := read.Object.(ir.Read); isRead && receiver.Local == self {
			return true
		}
		where := read.ViewWhere
		if where == "" {
			where = l.result.Source
		}
		failure = judge(where, read.Name, read.Object)
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
		if err := judge(l.program.Where(use.node), use.name, use.receiver); err != nil {
			return err
		}
	}
	return nil
}
