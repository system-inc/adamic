// Frozen scanning reference from 212b8d1de1e387c2a3e5764ff8a9b333e0728f3b.
// Methods are copied verbatim with only receiver/type and entry-point names renamed.
// Graph, Program, closure, async-frame, iteration-cell and ownership-query rules
// remain those of the base. Do not replace them with the indexed implementations.
package lower

import (
	"context"
	"errors"
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/flow"
	"github.com/system-inc/adamic/internal/fresh"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"maps"
	"path/filepath"
	"reflect"
	"slices"
	"sort"
	"strconv"
	"strings"
)

type scanCycleFinder struct {
	l *lowering

	// seen is every type the program uses, in the order first met, with where it was met.
	seen  []*checker.Type
	where map[*checker.Type]*ast.Node

	// shapes is every seen object type that isn't an array, a map or a function: what a value seen as
	// some object type may really be. shaped are the literals' own types among them.
	shapes []*checker.Type
	shaped map[*checker.Type]bool

	// scalarShapes shares identical plain literal shapes. These have no hidden
	// members, signatures, or reference fields; their complete structural types
	// are the same, so repeating them adds no reachability edge.
	scalarShapes map[string]bool

	// writes is every write in the program, judged by the relaxation for fresh writes (fresh.go),
	// worked out the first time a slot needs it.
	writes []fresh.Write

	libraryIterators []libraryIteratorCapture
}

func (l *lowering) scanFindCycles(modules []*ast.SourceFile) error {
	// Generated frame/Promise/reaction layouts are IR identities, never checker declarations.
	// Only their runtime-owned protocol edges are exempt. Source types, including a class
	// called adamic_async_frame, continue through slotsOf with no name-based escape hatch.
	if l.result.HasAsync() {
		for _, generated := range l.result.Generated {
			if !fresh.RuntimeBreaksCycles(generated) {
				return fmt.Errorf("lower: unaudited generated cycle identity")
			}
		}
	}
	finder := &scanCycleFinder{l: l, where: map[*checker.Type]*ast.Node{}}
	for _, module := range modules {
		var visit ast.Visitor
		visit = func(node *ast.Node) bool {
			if ast.IsPartOfTypeNode(node) {
				return false
			}
			finder.libraryIteratorMade(node)
			switch {
			case node.Kind == ast.KindObjectLiteralExpression:
				// A literal is what a value may really be, and nothing is written through its own
				// type: it's seen as the type it's written into, or kept in a variable whose type
				// is used where it's declared.
				finder.shape(l.checker.GetTypeAtLocation(node), node)
			case node.Kind == ast.KindArrayLiteralExpression || node.Kind == ast.KindCallExpression || node.Kind == ast.KindNewExpression:
				// A value just made, by a literal, a call or new, is written through only once it's
				// kept somewhere, and where it's kept has a type of its own, used there. Until then
				// it's what a value may really be.
				finder.made(l.checker.GetTypeAtLocation(node), node)
			case ast.IsIdentifier(node) || node.Kind == ast.KindThisKeyword || node.Kind == ast.KindPropertyAccessExpression || node.Kind == ast.KindElementAccessExpression:
				// Where a value is named or reached, and so where a write can go through its type.
				if ast.IsIdentifier(node) && !ast.IsDeclarationName(node) && !ast.IsExpressionNode(node) {
					break
				}
				finder.use(l.checker.GetTypeAtLocation(node), node)
			}
			return node.ForEachChild(visit)
		}
		module.AsNode().ForEachChild(visit)
	}
	// Every instantiation lowering made, Box<Node> inside Maker<Node> included, which the source may
	// never write.
	for _, proven := range l.instantiated {
		finder.use(proven, l.classNodeFor(proven))
	}
	// Program membership changes ownership only after every ordinary graph check passes.
	if err := finder.graphTypes(modules); err != nil {
		return err
	}
	if l.programDiscovery {
		l.resolveCountTypes(modules)
		l.programPlan = l.scanSelectProgramMembership(finder, modules)
		return l.checkLazyViewReads()
	}
	l.resolveCountTypes(modules)
	return l.checkLazyViewReads()
}

func (f *scanCycleFinder) made(proven *checker.Type, where *ast.Node) {
	switch {
	case proven == nil:
	case proven.Flags()&(checker.TypeFlagsUnion|checker.TypeFlagsIntersection) != 0:
		for _, member := range proven.Types() {
			f.made(member, where)
		}
	case proven.Flags()&checker.TypeFlagsObject == 0 || f.isFunction(proven):
		f.use(proven, where)
	case f.l.recordElement(proven) != nil:
		f.use(f.l.recordElement(proven), where)
	case f.l.checker.IsArrayType(proven) || checker.IsTupleType(proven) || f.l.isLibraryType(proven, "Map", "ReadonlyMap", "Set", "ReadonlySet", "Promise"):
		for _, argument := range f.l.checker.GetTypeArguments(proven) {
			f.use(argument, where)
		}
	default:
		f.shape(proven, where)
	}
}

func (f *scanCycleFinder) shape(proven *checker.Type, where *ast.Node) {
	if f.shaped[proven] {
		return
	}
	if f.shaped == nil {
		f.shaped = map[*checker.Type]bool{}
	}
	f.shaped[proven] = true
	if key, ok := f.scalarShapeKey(proven, where); ok {
		if f.scalarShapes[key] {
			return
		}
		if f.scalarShapes == nil {
			f.scalarShapes = map[string]bool{}
		}
		f.scalarShapes[key] = true
	}
	f.shapes = append(f.shapes, proven)
	for _, field := range f.fields(proven) {
		f.use(f.l.checker.GetTypeOfSymbol(field), declaredAt(field, where))
	}
}

func (f *scanCycleFinder) scalarShapeKey(proven *checker.Type, where *ast.Node) (string, bool) {
	if where == nil || where.Kind != ast.KindObjectLiteralExpression ||
		proven.Flags() != checker.TypeFlagsObject ||
		proven.ObjectFlags()&checker.ObjectFlagsFreshLiteral == 0 ||
		len(f.l.checker.GetIndexInfosOfType(proven)) != 0 ||
		len(f.l.checker.GetSignaturesOfType(proven, checker.SignatureKindCall)) != 0 ||
		len(f.l.checker.GetSignaturesOfType(proven, checker.SignatureKindConstruct)) != 0 {
		return "", false
	}
	properties := where.AsObjectLiteralExpression().Properties.Nodes
	fields := f.l.checker.GetPropertiesOfType(proven)
	if len(properties) != len(fields) {
		return "", false
	}
	var key strings.Builder
	for i, property := range properties {
		if property.Kind != ast.KindPropertyAssignment || property.Name() == nil ||
			property.Name().Kind != ast.KindIdentifier {
			return "", false
		}
		field := fields[i]
		if field.Name != property.Name().Text() {
			return "", false
		}
		value := f.l.checker.GetTypeOfSymbol(field)
		// Exact checker identities retain distinct scalar literal values.
		flags := value.Flags()
		boolean := flags&checker.TypeFlagsBoolean != 0 && flags & ^(checker.TypeFlagsBoolean|checker.TypeFlagsUnion) == 0
		if !boolean && flags & ^(checker.TypeFlagsNumber|checker.TypeFlagsString|checker.TypeFlagsBooleanLiteral|checker.TypeFlagsNumberLiteral|checker.TypeFlagsStringLiteral|checker.TypeFlagsNull|checker.TypeFlagsUndefined) != 0 {
			return "", false
		}
		key.WriteString(strconv.Itoa(len(field.Name)))
		key.WriteByte(':')
		key.WriteString(field.Name)
		fmt.Fprintf(&key, ":%p:%d:%t;", value, field.Flags, f.l.checker.IsReadonlySymbol(field))
	}
	return key.String(), true
}

func (f *scanCycleFinder) use(proven *checker.Type, where *ast.Node) {
	if proven == nil {
		return
	}
	if _, isSeen := f.where[proven]; isSeen || f.template(proven) {
		return
	}
	f.where[proven] = where
	if proven.Flags()&checker.TypeFlagsObject != 0 && proven.ObjectFlags()&checker.ObjectFlagsFreshLiteral != 0 {
		// An object literal's own type, as the literal expression has it. Nothing is written through
		// it: written into a declared type, the object is seen only as that, and kept in a variable
		// of its own, the variable's type is the literal's regular type, which is used where it's
		// declared. It's still what a value may really be.
		f.shape(proven, where)
		return
	}
	f.seen = append(f.seen, proven)
	flags := proven.Flags()
	if flags&(checker.TypeFlagsUnion|checker.TypeFlagsIntersection) != 0 {
		for _, member := range proven.Types() {
			f.use(member, where)
		}
		return
	}
	if flags&checker.TypeFlagsObject == 0 {
		return
	}
	if f.isFunction(proven) {
		return
	}
	if element := f.l.recordElement(proven); element != nil {
		f.use(element, where)
		return
	}
	if f.l.checker.IsArrayType(proven) || checker.IsTupleType(proven) || f.l.isLibraryType(proven, "Map", "ReadonlyMap", "Set", "ReadonlySet", "Promise") {
		for _, argument := range f.l.checker.GetTypeArguments(proven) {
			f.use(argument, where)
		}
		return
	}
	if !f.shaped[proven] {
		f.shapes = append(f.shapes, proven)
	}
	for _, field := range f.fields(proven) {
		f.use(f.l.checker.GetTypeOfSymbol(field), declaredAt(field, where))
	}
}

func (f *scanCycleFinder) template(proven *checker.Type) bool {
	if proven.Flags()&(checker.TypeFlagsAny|checker.TypeFlagsTypeParameter) != 0 {
		return true
	}
	if proven.Flags()&checker.TypeFlagsObject == 0 || proven.ObjectFlags()&checker.ObjectFlagsReference == 0 {
		return false
	}
	for _, argument := range f.l.checker.GetTypeArguments(proven) {
		if argument.Flags()&(checker.TypeFlagsAny|checker.TypeFlagsTypeParameter) != 0 {
			return true
		}
	}
	return false
}

func (f *scanCycleFinder) fields(proven *checker.Type) []*ast.Symbol {
	fields := []*ast.Symbol{}
	for _, property := range f.l.checker.GetPropertiesOfType(proven) {
		if (property.Flags&ast.SymbolFlagsMethod == 0 || literalMethod(property)) && !accessorSymbol(property) && !(f.l.isStaticType(proven) && property.Name == "prototype") {
			fields = append(fields, property)
		}
	}
	return fields
}

func (f *scanCycleFinder) isFunction(proven *checker.Type) bool {
	return f.l.isLibraryType(proven, "Function") || len(f.l.checker.GetSignaturesOfType(proven, checker.SignatureKindCall)) > 0 || (len(f.l.checker.GetSignaturesOfType(proven, checker.SignatureKindConstruct)) > 0 && (!f.l.isStaticType(proven) || len(f.l.staticGlobals) == 0))
}

func (f *scanCycleFinder) weak(proven *checker.Type) bool {
	representation, _ := f.l.representation(proven)
	return representation == ir.Weak
}

func (f *scanCycleFinder) slotsOf(holder *checker.Type) error {
	l := f.l
	if holder.Flags()&checker.TypeFlagsObject == 0 || f.isFunction(holder) {
		return nil
	}
	name := l.checker.TypeToString(holder)
	switch {
	case l.checker.IsArrayType(holder) || checker.IsTupleType(holder):
		if l.isLibraryType(holder, "ReadonlyArray") || (checker.IsTupleType(holder) && holder.TargetTupleType().IsReadonly()) {
			return nil
		}
		for _, element := range l.checker.GetTypeArguments(holder) {
			if f.weak(element) || !f.reaches(element, cycleNode{proven: holder}) {
				continue
			}
			write := f.unproven(fresh.WriteElement, holder, "")
			if write == nil {
				continue
			}
			target := f.present(element)
			return &Refused{
				cycle: true,
				Where: l.program.Where(f.where[holder]),
				What:  name + ", an array whose elements can reach back to an array like it: a cycle reference counting can't free, and " + f.writtenAt(write) + " may close one (" + write.Why + ")",
				Fix:   "declare the elements weak, Weak<" + target + ">[] (import type { Weak } from 'adamic'), which don't count and read undefined once what they point to is freed; or make it readonly " + target + "[]; or write into such an array only values this function made, or only into one it made (adamic/cycle-capable)",
			}
		}
	case l.recordElement(holder) != nil:
		element := l.recordElement(holder)
		if !f.weak(element) && f.reaches(element, cycleNode{proven: holder}) {
			if write := f.unproven(fresh.WriteMapEntry, holder, ""); write != nil {
				return &Refused{cycle: true, Where: l.program.Where(f.where[holder]), What: name + ", a record whose values can reach back to its holder: a cycle reference counting cannot free, and " + f.writtenAt(write) + " may close one", Fix: "use weak links or write only values proven unable to reach the record (adamic/cycle-capable)"}
			}
		}
	case l.isLibraryType(holder, "ReadonlyMap"), l.isLibraryType(holder, "ReadonlySet"):
	case l.isLibraryType(holder, "Set"):
		arguments := l.checker.GetTypeArguments(holder)
		if len(arguments) == 1 && f.reaches(arguments[0], cycleNode{proven: holder}) {
			if write := f.unproven(fresh.WriteSetElement, holder, ""); write != nil {
				return &Refused{
					cycle: true,
					Where: l.program.Where(f.where[holder]),
					What:  name + ", a set whose elements can reach back to a set like it: a cycle reference counting can't free, and " + f.writtenAt(write) + " may close one (" + write.Why + ")",
					Fix:   "make it a ReadonlySet, built whole when it's made; or add to such a set only values this function made, or only to one it made (adamic/cycle-capable)",
				}
			}
		}
	case l.isLibraryType(holder, "Map"):
		arguments := l.checker.GetTypeArguments(holder)
		if len(arguments) == 2 && f.reaches(arguments[0], cycleNode{proven: holder}) {
			if write := f.unproven(fresh.WriteMapEntry, holder, ""); write != nil {
				return &Refused{
					cycle: true,
					Where: l.program.Where(f.where[holder]),
					What:  name + ", a map whose keys can reach back to a map like it: a cycle reference counting can't free, and " + f.writtenAt(write) + " may close one (" + write.Why + ")",
					Fix:   "make it a ReadonlyMap, built whole when it's made; or set in such a map only keys this function made, or only in one it made (adamic/cycle-capable)",
				}
			}
		}
		if len(arguments) == 2 && !f.weak(arguments[1]) && f.reaches(arguments[1], cycleNode{proven: holder}) {
			if write := f.unproven(fresh.WriteMapEntry, holder, ""); write != nil {
				return &Refused{
					cycle: true,
					Where: l.program.Where(f.where[holder]),
					What:  name + ", a map whose values can reach back to a map like it: a cycle reference counting can't free, and " + f.writtenAt(write) + " may close one (" + write.Why + ")",
					Fix:   "declare the values weak, Map<" + l.checker.TypeToString(arguments[0]) + ", Weak<" + f.present(arguments[1]) + ">> (import type { Weak } from 'adamic'), or make it a ReadonlyMap; or set in such a map only values this function made, or only in one it made (adamic/cycle-capable)",
				}
			}
		}
	default:
		for _, field := range f.fields(holder) {
			// A readonly field is a slot too: a constructor writes it, after this may have been put
			// somewhere (parent.kids.push(this)), and then that write can close a cycle. Every other
			// way a readonly field gets its value (a literal, an initializer) makes the holder with it,
			// and is no write at all.
			readonly := l.checker.IsReadonlySymbol(field)
			proven := l.checker.GetTypeOfSymbol(field)
			if f.weak(proven) || !f.reaches(proven, cycleNode{proven: holder}) {
				continue
			}
			write := f.unproven(fresh.WriteField, holder, field.Name)
			if write == nil {
				continue
			}
			where := f.where[holder]
			if len(field.Declarations) > 0 {
				where = field.Declarations[0]
			}
			target := l.checker.TypeToString(proven)
			kind, orReadonly := "a mutable field", "; or make it readonly"
			if readonly {
				kind, orReadonly = "a readonly field, which its constructor writes,", ""
			}
			return &Refused{
				cycle: true,
				Where: l.program.Where(where),
				What:  name + "." + l.cycleFieldName(field) + ", " + kind + " of type " + target + ", which can reach back to the " + name + " holding it: a cycle reference counting can't free, and " + f.writtenAt(write) + " may close one (" + write.Why + ")",
				Fix:   "declare it " + l.cycleFieldName(field) + ": Weak<" + f.present(proven) + "> (import type { Weak } from 'adamic'), which doesn't count and reads undefined once what it points to is freed" + orReadonly + "; or write into it only values this function made, or only into what it made (adamic/cycle-capable)",
			}
		}
	}
	return nil
}

func (f *scanCycleFinder) reaches(from *checker.Type, target cycleNode) bool {
	visited := map[cycleNode]bool{}
	queue := []cycleNode{{proven: from}}
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		if visited[node] {
			continue
		}
		visited[node] = true
		if node.cell != 0 {
			if node == target {
				return true
			}
			// An interior async cell retains its owner, not just its declared value.
			// Completion drops private slots; captured slots remain reachable through
			// any escaped cell. Follow every such slot irrespective of signature.
			local := f.l.result.Locals[node.cell-1]
			if local.EnvironmentCell && local.Function >= 0 {
				owner := f.l.result.Functions[local.Function]
				if owner.Async {
					for _, held := range owner.FrameEnvironment {
						if f.l.result.Locals[held].Captured {
							queue = append(queue, cycleNode{cell: held + 1})
						}
					}
				}
			}
			if proven := f.l.localTypes[node.cell-1]; proven != nil {
				queue = append(queue, cycleNode{proven: proven})
			}
			// A this shared by every instantiation of a class held the same way is each of them.
			for _, proven := range f.l.localAlso[node.cell-1] {
				queue = append(queue, cycleNode{proven: proven})
			}
			continue
		}
		proven := node.proven
		flags := proven.Flags()
		if f.weak(proven) || f.template(proven) {
			// A Weak holds nothing.
			continue
		}
		if flags&(checker.TypeFlagsUnion|checker.TypeFlagsIntersection) != 0 {
			for _, member := range proven.Types() {
				queue = append(queue, cycleNode{proven: member})
			}
			continue
		}
		if flags&checker.TypeFlagsObject == 0 {
			continue
		}
		if target.proven != nil && f.related(proven, target.proven) {
			return true
		}
		switch {
		case f.l.isLibraryType(proven, "MapIterator", "SetIterator"):
			queue = append(queue, f.libraryIteratorCaptures(proven)...)
		case f.isFunction(proven):
			// Construct signatures can hide constructor objects behind an interface.
			if f.l.isLibraryType(proven, "Function") || len(f.l.checker.GetSignaturesOfType(proven, checker.SignatureKindConstruct)) > 0 {
				for symbol := range f.l.statics {
					actual := f.l.checker.GetTypeOfSymbol(symbol)
					if f.l.checker.IsTypeAssignableTo(actual, proven) {
						queue = append(queue, cycleNode{proven: actual})
					}
				}
			}
			// What a function value holds is what it captured: the cells of every function value the
			// program makes that can be seen as this type.
			for _, closure := range f.l.closureRecords {
				if !f.l.checker.IsTypeAssignableTo(closure.proven, proven) {
					continue
				}
				for _, local := range f.l.result.Functions[closure.function].Environment {
					queue = append(queue, cycleNode{cell: local + 1})
				}
			}
		case f.l.recordElement(proven) != nil:
			queue = append(queue, cycleNode{proven: f.l.recordElement(proven)})
		case f.l.checker.IsArrayType(proven) || checker.IsTupleType(proven) || f.l.isLibraryType(proven, "Map", "ReadonlyMap", "Set", "ReadonlySet", "Promise"):
			for _, argument := range f.l.checker.GetTypeArguments(proven) {
				queue = append(queue, cycleNode{proven: argument})
			}
		default:
			if f.l.isStaticType(proven) {
				if parent := f.l.staticBase(f.l.staticClass(proven, nil)); parent != nil {
					queue = append(queue, cycleNode{proven: f.l.checker.GetTypeOfSymbol(f.l.symbol(parent.Name()))})
				}
			}
			for _, accessor := range f.l.accessorCaptures {
				if f.l.checker.IsTypeAssignableTo(accessor.holder, proven) {
					for _, local := range f.l.result.Functions[accessor.function].Environment {
						queue = append(queue, cycleNode{cell: local + 1})
					}
				}
			}
			for _, field := range f.fields(proven) {
				queue = append(queue, cycleNode{proven: f.l.checker.GetTypeOfSymbol(field)})
			}
			// A value seen as this type may be any object type the program has that can be seen as
			// it, with fields this type doesn't show.
			for _, shape := range f.shapes {
				if shape != proven && f.l.checker.IsTypeAssignableTo(shape, proven) {
					queue = append(queue, cycleNode{proven: shape})
				}
			}
		}
	}
	return false
}

func (f *scanCycleFinder) present(proven *checker.Type) string {
	if proven.Flags()&checker.TypeFlagsUnion == 0 {
		return f.l.checker.TypeToString(proven)
	}
	members := []string{}
	for _, member := range proven.Types() {
		if member.Flags()&checker.TypeFlagsUndefined == 0 {
			members = append(members, f.l.checker.TypeToString(member))
		}
	}
	return strings.Join(members, " | ")
}

func (f *scanCycleFinder) related(one *checker.Type, other *checker.Type) bool {
	if one == other {
		return true
	}
	if f.isFunction(one) || f.isFunction(other) {
		return false
	}
	return f.l.checker.IsTypeAssignableTo(one, other) || f.l.checker.IsTypeAssignableTo(other, one)
}

func (f *scanCycleFinder) unproven(kind fresh.WriteKind, holder *checker.Type, name string) *fresh.Write {
	if f.writes == nil {
		f.writes = fresh.ProveWrites(f.l.result)
	}
	for index := range f.writes {
		write := &f.writes[index]
		if write.Proven {
			continue
		}
		if write.Kind == fresh.WriteUnknown {
			return write
		}
		if write.Kind != kind || (kind == fresh.WriteField && !f.l.cycleFieldMatches(holder, name, write.Name)) {
			continue
		}
		if write.Site == 0 {
			// A write lowering didn't record: it may be into anything of its kind.
			return write
		}
		written := f.l.writeSites[write.Site-1].holder
		if written == nil || f.template(written) || f.related(holder, written) {
			return write
		}
	}
	return nil
}

func (f *scanCycleFinder) writtenAt(write *fresh.Write) string {
	if write.Site == 0 {
		return "a write in " + f.l.functionName(write.Function)
	}
	return "the write at " + f.l.program.Where(f.l.writeSites[write.Site-1].node)
}

func (f *scanCycleFinder) graphTypes(modules []*ast.SourceFile) error {
	// Graph regions don't follow record values, so a cycle through a record stays refused rather than seeded.
	for _, proven := range f.seen {
		if f.l.recordElement(proven) != nil {
			if err := f.slotsOf(proven); err != nil {
				return err
			}
		}
	}
	seeds := map[cycleNode]bool{}
	for _, proven := range f.seen {
		seed, err := graphSeed(f.slotsOf(proven))
		if err != nil {
			return err
		}
		if seed {
			seeds[cycleNode{proven: proven}] = true
		}
	}
	for local, declared := range f.l.result.Locals {
		proven := f.l.localTypes[local]
		// A closed frame input (closed_frame_inputs.go) is proven not to close a cycle, so its cell
		// stays counted rather than seeding a graph region.
		if declared.Captured && !declared.Global && proven != nil && !f.weak(proven) && f.reaches(proven, cycleNode{cell: local + 1}) && !f.l.closedFrameInput(local) {
			// An async frame owns its captured cells and is never a graph member, so a cycle through
			// one stays refused, as it was before graph regions.
			if declared.Function >= 0 && f.l.result.Functions[declared.Function].Async {
				return &Refused{
					Where: f.l.program.Where(f.l.localNodes[local]),
					What:  "async frame capture cycle: '" + declared.Name + "', a variable a function value captures and can be reached from what it holds, so the function holds the variable and the variable holds the function: a cycle reference counting can't free",
					Fix:   "remove the captured strong back-reference, use a module function declaration that captures nothing, or declare the variable Weak<...> and keep the function somewhere strong (adamic/cycle-capable)",
				}
			}
			seeds[cycleNode{cell: local + 1}] = true
		}
	}
	if len(seeds) == 0 {
		return nil
	}
	// Literals may have narrower types than the mutable view through which they
	// are later written. Include those allocation identities without creating seeds.
	for _, module := range modules {
		var visit ast.Visitor
		visit = func(node *ast.Node) bool {
			if ast.IsPartOfTypeNode(node) {
				return false
			}
			if ast.IsExpressionNode(node) {
				f.use(f.l.checker.GetTypeAtLocation(node), node)
				if contextual := f.l.checker.GetContextualType(node, checker.ContextFlagsNone); contextual != nil {
					f.use(contextual, node)
				}
			}
			return node.ForEachChild(visit)
		}
		module.AsNode().ForEachChild(visit)
	}
	for _, closure := range f.l.closureRecords {
		f.use(closure.proven, closure.node)
	}
	candidates := []cycleNode{}
	for proven := range f.where {
		candidates = append(candidates, cycleNode{proven: proven})
	}
	for _, shape := range f.shapes {
		candidates = append(candidates, cycleNode{proven: shape})
	}
	for local, declared := range f.l.result.Locals {
		if declared.Captured && !declared.Global {
			candidates = append(candidates, cycleNode{cell: local + 1})
		}
	}
	edges := map[cycleNode][]cycleNode{}
	queue := append([]cycleNode{}, candidates...)
	for len(queue) != 0 {
		node := queue[len(queue)-1]
		queue = queue[:len(queue)-1]
		if _, seen := edges[node]; seen {
			continue
		}
		links := f.graphLinks(node)
		// The finder's related relation treats compatible structural views as the
		// same possible allocation. Weak views never become strong edges.
		if node.proven != nil && node.proven.Flags()&checker.TypeFlagsObject != 0 && !f.weak(node.proven) && !f.template(node.proven) && !f.isFunction(node.proven) {
			for _, view := range candidates {
				if view.proven != nil && view.proven.Flags()&checker.TypeFlagsObject != 0 && !f.weak(view.proven) && !f.template(view.proven) && f.related(node.proven, view.proven) {
					links = append(links, view)
				}
			}
		}
		edges[node] = links
		queue = append(queue, links...)
	}
	// Tarjan's components, with a component selected iff it contains a seed.
	next := 0
	index, low := map[cycleNode]int{}, map[cycleNode]int{}
	active := map[cycleNode]bool{}
	stack := []cycleNode{}
	graph := map[cycleNode]bool{}
	var throughPromise []cycleNode
	var visit func(cycleNode)
	visit = func(node cycleNode) {
		next++
		index[node], low[node] = next, next
		stack = append(stack, node)
		active[node] = true
		for _, link := range edges[node] {
			if index[link] == 0 {
				visit(link)
				if low[link] < low[node] {
					low[node] = low[link]
				}
			} else if active[link] && index[link] < low[node] {
				low[node] = index[link]
			}
		}
		if low[node] != index[node] {
			return
		}
		component := []cycleNode{}
		selected := false
		for {
			last := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			active[last] = false
			component = append(component, last)
			selected = selected || seeds[last]
			if last == node {
				break
			}
		}
		if selected {
			for _, each := range component {
				graph[each] = true
				if each.proven != nil && f.l.isLibraryType(each.proven, "Promise") && throughPromise == nil {
					throughPromise = component
				}
			}
		}
	}
	for node := range edges {
		if index[node] == 0 {
			visit(node)
		}
	}
	// A region frees a cycle only when every value on it is a member. A Promise never is (its
	// payload is a counted value outside any region), so a cycle through one would keep its region
	// alive forever: refuse it, as before graph regions.
	if throughPromise != nil {
		var where *ast.Node
		for _, each := range throughPromise {
			if each.proven != nil && f.where[each.proven] != nil {
				where = f.where[each.proven]
				break
			}
		}
		if where == nil {
			where = modules[0].AsNode()
		}
		return &Refused{
			Where: f.l.program.Where(where),
			What:  "a Promise whose payload can reach back to what holds it: a cycle reference counting can't free, and a Promise can't join a graph region",
			Fix:   "don't keep a Promise in a value its result can reach, or declare the field Weak<...> (adamic/cycle-capable)",
		}
	}
	// An otherwise acyclic cache of graph elements belongs inside their graph too.
	for changed := true; changed; {
		changed = false
		for node := range edges {
			if graph[node] || node.proven == nil || f.weak(node.proven) {
				continue
			}
			proven := node.proven
			if f.l.checker.IsArrayType(proven) || checker.IsTupleType(proven) || f.l.isLibraryType(proven, "Map", "ReadonlyMap", "Set", "ReadonlySet") {
				for _, argument := range f.graphArguments(proven) {
					if graph[cycleNode{proven: argument}] {
						graph[node] = true
						changed = true
					}
					if argument.Flags()&checker.TypeFlagsUnion != 0 {
						for _, member := range argument.Types() {
							if graph[cycleNode{proven: member}] {
								graph[node] = true
								changed = true
							}
						}
					}
				}
			}
		}
	}
	f.l.result.GraphTypes = map[int]bool{}
	for node := range graph {
		if node.proven != nil && !f.weak(node.proven) {
			f.l.result.GraphTypes[int(node.proven.Id())] = true
		}
		if node.cell != 0 {
			f.l.result.Locals[node.cell-1].GraphCell = true
		}
	}
	for _, closure := range f.l.closureRecords {
		if f.l.result.GraphTypes[int(closure.proven.Id())] {
			f.l.result.Functions[closure.function].GraphClosure = true
		}
	}
	// One frame allocation is indivisible, even if siblings capture disjoint slots.
	for changed := true; changed; {
		changed = false
		for index := range f.l.result.Functions {
			function := &f.l.result.Functions[index]
			for _, local := range function.Environment {
				if f.l.result.Locals[local].GraphCell && !function.GraphClosure {
					function.GraphClosure = true
					changed = true
				}
			}
			if function.GraphClosure {
				for _, local := range function.Environment {
					if !f.l.result.Locals[local].GraphCell {
						f.l.result.Locals[local].GraphCell = true
						changed = true
					}
				}
			}
			frame := false
			for _, local := range function.FrameEnvironment {
				frame = frame || f.l.result.Locals[local].GraphCell
			}
			if frame {
				for _, local := range function.FrameEnvironment {
					if !f.l.result.Locals[local].GraphCell {
						f.l.result.Locals[local].GraphCell = true
						changed = true
					}
				}
			}
		}
	}
	// Forwarded nested callbacks retain their parent frame through a path outside
	// the direct environment ownership proof. Keep that capture cycle refused.
	for _, closure := range f.l.closureRecords {
		function := f.l.result.Functions[closure.function]
		if function.GraphClosure && function.ForwardedNestedParent != 0 {
			return &Refused{
				Where: f.l.program.Where(closure.node),
				What:  "a nested callback retains a cyclic parent environment through a forwarded frame",
				Fix:   "remove the captured strong back-reference, use a module function declaration that captures nothing, or declare the variable Weak<...> (adamic/cycle-capable)",
			}
		}
	}
	for _, instance := range f.l.instances {
		for _, local := range instance.thisLocals {
			types := append([]*checker.Type{f.l.localTypes[local]}, f.l.localAlso[local]...)
			for _, proven := range types {
				if proven != nil && f.l.result.GraphTypes[int(proven.Id())] {
					f.l.result.Classes[instance.class-1].Graph = true
				}
			}
		}
	}
	for symbol, instance := range f.l.statics {
		if f.l.result.GraphTypes[int(f.l.checker.GetTypeOfSymbol(symbol).Id())] {
			f.l.result.Classes[instance.class-1].Graph = true
		}
	}
	f.graphFlows()
	return f.graphAdoptionBoundary()
}

func (f *scanCycleFinder) graphLinks(node cycleNode) []cycleNode {
	if node.cell != 0 {
		links := []cycleNode{}
		if proven := f.l.localTypes[node.cell-1]; proven != nil {
			links = append(links, cycleNode{proven: proven})
		}
		for _, proven := range f.l.localAlso[node.cell-1] {
			links = append(links, cycleNode{proven: proven})
		}
		if f.l.result.Locals[node.cell-1].IterationCell {
			return links
		}
		// A capture into one interior slot owns the entire environment allocation.
		for _, function := range f.l.result.Functions {
			found := false
			for _, local := range function.FrameEnvironment {
				found = found || local == node.cell-1
			}
			if found {
				for _, local := range function.FrameEnvironment {
					if local != node.cell-1 {
						links = append(links, cycleNode{cell: local + 1})
					}
				}
			}
		}
		return links
	}
	proven := node.proven
	if proven == nil || f.weak(proven) || f.template(proven) {
		return nil
	}
	links := []cycleNode{}
	if proven.Flags()&(checker.TypeFlagsUnion|checker.TypeFlagsIntersection) != 0 {
		for _, member := range proven.Types() {
			links = append(links, cycleNode{proven: member})
		}
		return links
	}
	if proven.Flags()&checker.TypeFlagsObject == 0 {
		return nil
	}
	switch {
	case f.l.isLibraryType(proven, "MapIterator", "SetIterator"):
		links = append(links, f.libraryIteratorCaptures(proven)...)
	case f.isFunction(proven):
		if len(f.l.checker.GetSignaturesOfType(proven, checker.SignatureKindConstruct)) > 0 {
			for symbol := range f.l.statics {
				actual := f.l.checker.GetTypeOfSymbol(symbol)
				if f.l.checker.IsTypeAssignableTo(actual, proven) {
					links = append(links, cycleNode{proven: actual})
				}
			}
		}
		for _, closure := range f.l.closureRecords {
			if f.l.checker.IsTypeAssignableTo(closure.proven, proven) {
				for _, local := range f.l.result.Functions[closure.function].Environment {
					links = append(links, cycleNode{cell: local + 1})
				}
			}
		}
	case f.l.checker.IsArrayType(proven) || checker.IsTupleType(proven) || f.l.isLibraryType(proven, "Map", "ReadonlyMap", "Set", "ReadonlySet", "Promise"):
		for _, argument := range f.graphArguments(proven) {
			links = append(links, cycleNode{proven: argument})
		}
	default:
		if f.l.isStaticType(proven) {
			if parent := f.l.staticBase(f.l.staticClass(proven, nil)); parent != nil {
				links = append(links, cycleNode{proven: f.l.checker.GetTypeOfSymbol(f.l.symbol(parent.Name()))})
			}
		}
		for _, field := range f.fields(proven) {
			links = append(links, cycleNode{proven: f.l.checker.GetTypeOfSymbol(field)})
		}
		for _, shape := range f.shapes {
			if shape != proven && f.l.checker.IsTypeAssignableTo(shape, proven) {
				links = append(links, cycleNode{proven: shape})
			}
		}
		for _, accessor := range f.l.accessorCaptures {
			if f.l.checker.IsTypeAssignableTo(accessor.holder, proven) {
				for _, local := range f.l.result.Functions[accessor.function].Environment {
					links = append(links, cycleNode{cell: local + 1})
				}
			}
		}
	}
	return links
}

func (f *scanCycleFinder) graphArguments(proven *checker.Type) []*checker.Type {
	if proven == nil || proven.Flags()&checker.TypeFlagsObject == 0 || proven.ObjectFlags()&checker.ObjectFlagsReference == 0 {
		return nil
	}
	return f.l.checker.GetTypeArguments(proven)
}

func (f *scanCycleFinder) graphAdoptionBoundary() error {
	reason := ""
	var site *ast.Node
	// Legacy tagged-union views record checked fields without a ViewOrigin.
	if len(f.l.result.ViewOrigins) != 0 || len(f.l.result.CheckedFields) != 0 {
		reason = "a checked-view adapter in a graph program"
		site = f.l.graphViewSite
	}
	inspect := func(node any) bool {
		switch value := node.(type) {
		case ir.ObjectLiteral:
			if value.Record {
				reason = "a dynamic-shape object in a graph program"
				site = f.l.graphDynamicSite
			}
		case ir.DynamicProperty:
			reason = "a dynamic-shape object in a graph program"
			site = f.l.graphDynamicSite
		case ir.ObjectCall:
			if value.Checked {
				reason = "a dynamic-shape object in a graph program"
				site = f.l.graphDynamicSite
			}
		}
		return true
	}
	walk(f.l.result.Main, inspect)
	for _, function := range f.l.result.Functions {
		walk(function.Body, inspect)
	}
	if reason != "" {
		return &Refused{Where: f.l.program.Where(site), What: reason, Fix: "keep checked views and dynamic shapes outside graph adoption (adamic/cycle-capable)"}
	}
	return nil
}

func (f *scanCycleFinder) graphFlows() {
	program := f.l.result
	next := 0
	program.Main = graphAllocationSites(reflect.ValueOf(program.Main), &next).Interface().([]ir.Statement)
	for i := range program.Functions {
		program.Functions[i].Body = graphAllocationSites(reflect.ValueOf(program.Functions[i].Body), &next).Interface().([]ir.Statement)
	}

	var graphType func(*checker.Type) bool
	graphType = func(proven *checker.Type) bool {
		if proven == nil || f.weak(proven) {
			return false
		}
		if program.GraphTypes[int(proven.Id())] {
			return true
		}
		if proven.Flags()&(checker.TypeFlagsUnion|checker.TypeFlagsIntersection) != 0 {
			for _, member := range proven.Types() {
				if graphType(member) {
					return true
				}
			}
		}
		return false
	}
	// Positive nodes are locals followed by function results.
	resultNode := func(function int) int { return len(program.Locals) + function + 1 }
	sources := map[int][]ir.Expression{}
	wanted := map[int]bool{}
	queue := []int{}
	demand := func(node int) {
		if !wanted[node] {
			wanted[node] = true
			queue = append(queue, node)
		}
	}
	for local, proven := range f.l.localTypes {
		if graphType(proven) {
			demand(local + 1)
		}
	}
	for local, types := range f.l.localAlso {
		for _, proven := range types {
			if graphType(proven) {
				demand(local + 1)
			}
		}
	}
	pending := []ir.Expression{}
	// Site retains the checker view of the actual slot being written.
	slot := func(site int, name string, argument int) bool {
		if site <= 0 || site > len(f.l.writeSites) {
			return false
		}
		holder := f.l.writeSites[site-1].holder
		if holder == nil {
			return false
		}
		if name != "" {
			for _, field := range f.fields(holder) {
				if f.l.cycleFieldMatches(holder, field.Name, name) {
					return graphType(f.l.checker.GetTypeOfSymbol(field))
				}
			}
			return false
		}
		args := f.graphArguments(holder)
		return argument < len(args) && graphType(args[argument])
	}
	// A counted aggregate can have graph-typed fields/elements. Seed those
	// initializer slots individually, rather than promoting its whole payload.
	returns := map[int]*checker.Type{}
	for symbol, index := range f.l.functions {
		signatures := f.l.checker.GetSignaturesOfType(f.l.checker.GetTypeOfSymbol(symbol), checker.SignatureKindCall)
		if len(signatures) > 0 {
			returns[index] = f.l.checker.GetReturnTypeOfSignature(signatures[0])
		}
	}
	for _, closure := range f.l.closureRecords {
		signatures := f.l.checker.GetSignaturesOfType(closure.proven, checker.SignatureKindCall)
		if len(signatures) > 0 {
			returns[closure.function] = f.l.checker.GetReturnTypeOfSignature(signatures[0])
		}
	}
	var members func(ir.Expression, *checker.Type)
	members = func(expression ir.Expression, proven *checker.Type) {
		if expression == nil || proven == nil || f.weak(proven) {
			return
		}
		if graphType(proven) {
			pending = append(pending, expression)
		}
		switch value := expression.(type) {
		case ir.ObjectLiteral:
			for _, field := range f.fields(proven) {
				for _, initializer := range value.Fields {
					if f.l.cycleFieldMatches(proven, field.Name, initializer.Name) {
						members(initializer.Value, f.l.checker.GetTypeOfSymbol(field))
					}
				}
			}
		case ir.ArrayLiteral:
			arguments := f.graphArguments(proven)
			if len(arguments) > 0 {
				for _, element := range value.Elements {
					members(element, arguments[0])
				}
			}
		case ir.Conditional:
			members(value.WhenTrue, proven)
			members(value.WhenNot, proven)
		}
	}
	collect := func(body []ir.Statement, function int) {
		walk(body, func(node any) bool {
			switch value := node.(type) {
			case ir.Declare:
				members(value.Value, f.l.localTypes[value.Local])
				sources[value.Local+1] = append(sources[value.Local+1], value.Value)
			case ir.Assign:
				members(value.Value, f.l.localTypes[value.Local])
				sources[value.Local+1] = append(sources[value.Local+1], value.Value)
			case ir.Return:
				members(value.Value, returns[function])
				if function >= 0 {
					sources[resultNode(function)] = append(sources[resultNode(function)], value.Value)
				}
			case ir.Call:
				for _, target := range program.CallTargets(value) {
					if target < 0 || target >= len(program.Functions) {
						continue
					}
					for i, param := range program.Functions[target].Parameters {
						if i < len(value.Arguments) {
							sources[param+1] = append(sources[param+1], value.Arguments[i])
						}
					}
				}
			case ir.CallClosure:
				targets := program.ClosureTargets(value)
				if targets.Unknown {
					// Opaque function values are a documented leak-only frontier.
					break
				}
				for _, target := range targets.Functions {
					for i, param := range program.Functions[target].Parameters {
						if i < len(value.Arguments) {
							sources[param+1] = append(sources[param+1], value.Arguments[i])
						}
					}
				}
			case ir.SetProperty:
				if slot(value.Site, value.Name, 0) {
					pending = append(pending, value.Value)
				}
			case ir.SetIndex:
				if slot(value.Site, "", 0) {
					pending = append(pending, value.Value)
				}
			case ir.ArrayPush:
				if slot(value.Site, "", 0) {
					pending = append(pending, value.Value)
				}
			case ir.MapSet:
				if slot(value.Site, "", 0) {
					pending = append(pending, value.Key)
				}
				if slot(value.Site, "", 1) {
					pending = append(pending, value.Value)
				}
			case ir.SetAdd:
				if slot(value.Site, "", 0) {
					pending = append(pending, value.Value)
				}
			}
			return true
		})
	}
	collect(program.Main, -1)
	for i, function := range program.Functions {
		collect(function.Body, i)
	}
	var follow func(ir.Expression)
	follow = func(expression ir.Expression) {
		if expression == nil {
			return
		}
		value := reflect.ValueOf(expression)
		if field := value.FieldByName("GraphTypes"); field.IsValid() {
			ids := field.Interface().([]int)
			program.GraphTypes[ids[len(ids)-1]] = true
			return
		}
		switch value := expression.(type) {
		case ir.Read:
			demand(value.Local + 1)
		case ir.Call:
			for _, target := range program.CallTargets(value) {
				demand(resultNode(target))
			}
		case ir.CallClosure:
			targets := program.ClosureTargets(value)
			if targets.Unknown {
				return
			}
			for _, target := range targets.Functions {
				demand(resultNode(target))
			}
		case ir.Conditional:
			follow(value.WhenTrue)
			follow(value.WhenNot)
		case ir.Box:
			follow(value.Value)
		case ir.Narrow:
			follow(value.Value)
		case ir.Unwrap:
			follow(value.Value)
		case ir.CheckedCast:
			follow(value.Value)
		}
	}
	for _, expression := range pending {
		follow(expression)
	}
	for len(queue) > 0 {
		node := queue[len(queue)-1]
		queue = queue[:len(queue)-1]
		for _, expression := range sources[node] {
			follow(expression)
		}
	}
}

func (f *scanCycleFinder) libraryIteratorMade(node *ast.Node) {
	if node.Kind != ast.KindCallExpression {
		return
	}
	callee := ast.SkipParentheses(node.AsCallExpression().Expression)
	if callee.Kind != ast.KindPropertyAccessExpression {
		return
	}
	access := callee.AsPropertyAccessExpression()
	iterator := f.l.checker.GetTypeAtLocation(node)
	collection := f.l.checker.GetTypeAtLocation(access.Expression)
	if f.l.isLibraryType(iterator, "MapIterator", "SetIterator") &&
		f.l.isLibraryType(collection, "Map", "ReadonlyMap", "Set", "ReadonlySet") {
		f.libraryIterators = append(f.libraryIterators, libraryIteratorCapture{iterator, collection})
	}
}

func (f *scanCycleFinder) libraryIteratorCaptures(proven *checker.Type) []cycleNode {
	var held []cycleNode
	for _, capture := range f.libraryIterators {
		if f.l.checker.IsTypeAssignableTo(capture.iterator, proven) {
			held = append(held, cycleNode{proven: capture.collection})
		}
	}
	return held
}

type scanProgramShapeIndex struct {
	finder *scanCycleFinder
	groups []*programShape
	types  map[*checker.Type]*programShape
	cache  map[string][]cycleNode
}

func (f *scanCycleFinder) programShapes(candidates []cycleNode) *scanProgramShapeIndex {
	index := &scanProgramShapeIndex{finder: f, types: map[*checker.Type]*programShape{}, cache: map[string][]cycleNode{}}
	groups := map[string]*programShape{}
	for _, candidate := range candidates {
		t := candidate.proven
		if t == nil || t.Flags()&checker.TypeFlagsObject == 0 || f.weak(t) || f.template(t) || f.isFunction(t) || f.programScalarStorage(t) {
			continue
		}
		shape := index.shape(t)
		group := groups[shape.key]
		if group == nil {
			group = shape
			groups[shape.key] = group
			index.groups = append(index.groups, group)
		}
		group.candidates = append(group.candidates, candidate)
	}
	return index
}

func (index *scanProgramShapeIndex) shape(t *checker.Type) *programShape {
	if shape := index.types[t]; shape != nil {
		return shape
	}
	shape := &programShape{all: map[string]bool{}, required: map[string]bool{}}
	all, required := []string{}, []string{}
	for _, field := range index.finder.l.checker.GetPropertiesOfType(t) {
		shape.all[field.Name] = true
		all = append(all, field.Name)
		if field.Flags&ast.SymbolFlagsOptional == 0 {
			shape.required[field.Name] = true
			required = append(required, field.Name)
		}
	}
	sort.Strings(all)
	sort.Strings(required)
	shape.key = programShapeKey(all, required)
	index.types[t] = shape
	return shape
}

func (index *scanProgramShapeIndex) relatedCandidates(t *checker.Type) []cycleNode {
	shape := index.shape(t)
	if candidates, known := index.cache[shape.key]; known {
		return candidates
	}
	candidates := []cycleNode{}
	for _, group := range index.groups {
		if programShapeSubset(shape.required, group.all) || programShapeSubset(group.required, shape.all) {
			candidates = append(candidates, group.candidates...)
		}
	}
	index.cache[shape.key] = candidates
	return candidates
}

func (f *scanCycleFinder) programRegionSelection() map[cycleNode]bool {
	return f.programRegionSelectionMode(true)
}

func (f *scanCycleFinder) programRegionSelectionMode(indexed bool) map[cycleNode]bool {
	candidates := []cycleNode{}
	for proven := range f.where {
		candidates = append(candidates, cycleNode{proven: proven})
	}
	for _, shape := range f.shapes {
		candidates = append(candidates, cycleNode{proven: shape})
	}
	for local, declared := range f.l.result.Locals {
		if declared.Captured && !declared.Global {
			candidates = append(candidates, cycleNode{cell: local + 1})
		}
	}
	slices.SortFunc(candidates, func(a, b cycleNode) int {
		if a.proven == nil {
			if b.proven == nil {
				return a.cell - b.cell
			}
			return -1
		}
		if b.proven == nil {
			return 1
		}
		return int(a.proven.Id()) - int(b.proven.Id())
	})
	var shapes *scanProgramShapeIndex
	if indexed {
		shapes = f.programShapes(candidates)
	}
	edges, owning := map[cycleNode][]cycleNode{}, map[cycleNode][]cycleNode{}
	queue := append([]cycleNode{}, candidates...)
	for len(queue) != 0 {
		node := queue[len(queue)-1]
		queue = queue[:len(queue)-1]
		if _, seen := edges[node]; seen {
			continue
		}
		if node.proven != nil && f.programScalarStorage(node.proven) {
			edges[node] = nil
			owning[node] = nil
			continue
		}
		owned := f.programOwningLinks(node)
		links := append([]cycleNode{}, owned...)
		if node.proven != nil && node.proven.Flags()&checker.TypeFlagsObject != 0 && !f.weak(node.proven) && !f.template(node.proven) && !f.isFunction(node.proven) {
			views := candidates
			if shapes != nil {
				views = shapes.relatedCandidates(node.proven)
			}
			for _, view := range views {
				if view == node {
					continue
				}
				// A view can identify the same storage only when it has the same
				// owning field targets. Mere assignability cannot close a cycle.
				if view.proven != nil && !f.programScalarStorage(view.proven) && f.related(node.proven, view.proven) {
					other := f.programOwningLinks(view)
					if sameProgramLinks(owned, other) {
						links = append(links, view)
					}
				}
			}
		}
		owning[node] = owned
		edges[node] = links
		queue = append(queue, links...)
	}
	selected := ProgramRegionSCC(edges, owning)
	// Compiler's container rule is applied after SCC selection, to concrete
	// allocation-site types after monomorphization. Views cannot admit a container
	// whose elements/keys/values are outside the selected component.
	for node := range selected {
		if node.proven != nil && f.programContainer(node.proven) {
			delete(selected, node)
		}
	}
	for node := range edges {
		if node.proven != nil && f.programContainerMember(node.proven, selected) {
			selected[node] = true
		}
	}
	return selected
}

func (f *scanCycleFinder) programContainer(proven *checker.Type) bool {
	return f.l.checker.IsArrayType(proven) || checker.IsTupleType(proven) || f.l.isLibraryType(proven, "Map", "ReadonlyMap", "Set", "ReadonlySet")
}

func (f *scanCycleFinder) programContainerMember(proven *checker.Type, selected map[cycleNode]bool) bool {
	if !f.programContainer(proven) {
		return false
	}
	active := map[*checker.Type]bool{}
	var member func(*checker.Type) bool
	member = func(value *checker.Type) bool {
		if active[value] {
			return false
		}
		active[value] = true
		defer delete(active, value)
		if value.Flags()&(checker.TypeFlagsUnion|checker.TypeFlagsIntersection) != 0 {
			for _, part := range value.Types() {
				if member(part) {
					return true
				}
			}
			return false
		}
		if f.programContainer(value) {
			for _, argument := range f.graphArguments(value) {
				if member(argument) {
					return true
				}
			}
			return false
		}
		return !f.template(value) && !f.weak(value) && selected[cycleNode{proven: value}]
	}
	return member(proven)
}

func (f *scanCycleFinder) programScalarStorage(proven *checker.Type) bool {
	if !(f.l.checker.IsArrayType(proven) || checker.IsTupleType(proven) || f.l.isLibraryType(proven, "Map", "ReadonlyMap", "Set", "ReadonlySet")) {
		return false
	}
	var scalar func(*checker.Type) bool
	scalar = func(value *checker.Type) bool {
		if value.Flags()&checker.TypeFlagsUnion != 0 {
			for _, member := range value.Types() {
				if !scalar(member) {
					return false
				}
			}
			return true
		}
		return value.Flags()&(checker.TypeFlagsStringLike|checker.TypeFlagsNumberLike|checker.TypeFlagsBooleanLike|checker.TypeFlagsUndefined|checker.TypeFlagsNull|checker.TypeFlagsVoid|checker.TypeFlagsNever) != 0
	}
	for _, argument := range f.graphArguments(proven) {
		if !scalar(argument) {
			return false
		}
	}
	return true
}

func (f *scanCycleFinder) programOwningLinks(node cycleNode) []cycleNode {
	if node.cell != 0 {
		links := []cycleNode{}
		if proven := f.l.localTypes[node.cell-1]; proven != nil {
			links = append(links, cycleNode{proven: proven})
		}
		for _, proven := range f.l.localAlso[node.cell-1] {
			links = append(links, cycleNode{proven: proven})
		}
		// A capture into one interior slot owns the entire environment allocation.
		for _, function := range f.l.result.Functions {
			found := false
			for _, local := range function.FrameEnvironment {
				found = found || local == node.cell-1
			}
			if found {
				for _, local := range function.FrameEnvironment {
					if local != node.cell-1 {
						links = append(links, cycleNode{cell: local + 1})
					}
				}
			}
		}
		return links
	}
	proven := node.proven
	if proven == nil || f.weak(proven) || f.template(proven) {
		return nil
	}
	links := []cycleNode{}
	if proven.Flags()&(checker.TypeFlagsUnion|checker.TypeFlagsIntersection) != 0 {
		for _, member := range proven.Types() {
			links = append(links, cycleNode{proven: member})
		}
		return links
	}
	if proven.Flags()&checker.TypeFlagsObject == 0 {
		return nil
	}
	switch {
	case f.l.isLibraryType(proven, "MapIterator", "SetIterator"):
		links = append(links, f.libraryIteratorCaptures(proven)...)
	case f.isFunction(proven):
		if len(f.l.checker.GetSignaturesOfType(proven, checker.SignatureKindConstruct)) > 0 {
			for symbol := range f.l.statics {
				actual := f.l.checker.GetTypeOfSymbol(symbol)
				if f.l.checker.IsTypeAssignableTo(actual, proven) {
					links = append(links, cycleNode{proven: actual})
				}
			}
		}
		for _, closure := range f.l.closureRecords {
			if f.l.checker.IsTypeAssignableTo(closure.proven, proven) {
				for _, local := range f.l.result.Functions[closure.function].Environment {
					links = append(links, cycleNode{cell: local + 1})
				}
			}
		}
	case f.l.checker.IsArrayType(proven) || checker.IsTupleType(proven) || f.l.isLibraryType(proven, "Map", "ReadonlyMap", "Set", "ReadonlySet", "Promise"):
		for _, argument := range f.graphArguments(proven) {
			links = append(links, cycleNode{proven: argument})
		}
	default:
		if f.l.isStaticType(proven) {
			if parent := f.l.staticBase(f.l.staticClass(proven, nil)); parent != nil {
				links = append(links, cycleNode{proven: f.l.checker.GetTypeOfSymbol(f.l.symbol(parent.Name()))})
			}
		}
		for _, field := range f.fields(proven) {
			links = append(links, cycleNode{proven: f.l.checker.GetTypeOfSymbol(field)})
		}
		for _, accessor := range f.l.accessorCaptures {
			if f.l.checker.IsTypeAssignableTo(accessor.holder, proven) {
				for _, local := range f.l.result.Functions[accessor.function].Environment {
					links = append(links, cycleNode{cell: local + 1})
				}
			}
		}
	}
	return links
}

func (l *lowering) scanSelectProgramMembership(f *scanCycleFinder, modules []*ast.SourceFile) *programMembership {
	// Observe allocation and contextual identities even where all writes passed the
	// fresh proof and the ordinary graph-region pass had no seeds (e.g. Weak trees).
	for _, module := range modules {
		var visit ast.Visitor
		visit = func(node *ast.Node) bool {
			if ast.IsPartOfTypeNode(node) {
				return false
			}
			if ast.IsExpressionNode(node) {
				f.use(f.l.checker.GetTypeAtLocation(node), node)
				if contextual := f.l.checker.GetContextualType(node, checker.ContextFlagsNone); contextual != nil {
					f.use(contextual, node)
				}
			}
			return node.ForEachChild(visit)
		}
		module.AsNode().ForEachChild(visit)
	}
	for _, closure := range f.l.closureRecords {
		f.use(closure.proven, closure.node)
	}
	candidates := slices.Collect(maps.Keys(l.programCandidates))
	slices.SortFunc(candidates, func(a, b *checker.Type) int { return int(a.Id()) - int(b.Id()) })
	for _, proven := range candidates {
		f.use(proven, l.programCandidates[proven])
	}
	for _, local := range slices.Sorted(maps.Keys(l.localTypes)) {
		proven := l.localTypes[local]
		if proven != nil {
			f.use(proven, l.localNodes[local])
		}
	}
	selected := f.programRegionSelection()
	program := f.l.result
	program.ProgramTypes = ir.SortedMap[int, bool]{}
	for node := range selected {
		if node.proven != nil && !f.weak(node.proven) && !f.l.isLibraryType(node.proven, "Promise") {
			program.ProgramTypes[int(node.proven.Id())] = true
		}
		if node.cell != 0 {
			program.Locals[node.cell-1].ProgramRegion = true
		}
	}
	for _, closure := range f.l.closureRecords {
		program.Functions[closure.function].ProgramRegion = program.ProgramTypes[int(closure.proven.Id())]
	}
	// Capture environments are indivisible allocations; retaining an extra sibling
	// costs retention, while treating an interior cell as independently freeable is unsafe.
	for changed := true; changed; {
		changed = false
		for i := range program.Functions {
			function := &program.Functions[i]
			for _, local := range function.Environment {
				if program.Locals[local].ProgramRegion && !function.ProgramRegion {
					function.ProgramRegion = true
					changed = true
				}
			}
			if function.ProgramRegion {
				for _, local := range function.Environment {
					if !program.Locals[local].ProgramRegion {
						program.Locals[local].ProgramRegion = true
						changed = true
					}
				}
			}
			frame := false
			for _, local := range function.FrameEnvironment {
				frame = frame || program.Locals[local].ProgramRegion
			}
			if frame {
				for _, local := range function.FrameEnvironment {
					if !program.Locals[local].ProgramRegion {
						program.Locals[local].ProgramRegion = true
						changed = true
					}
				}
			}
		}
	}
	for _, instance := range f.l.instances {
		for _, local := range instance.thisLocals {
			for _, proven := range append([]*checker.Type{f.l.localTypes[local]}, f.l.localAlso[local]...) {
				if proven != nil && program.ProgramTypes[int(proven.Id())] {
					program.Classes[instance.class-1].ProgramRegion = true
				}
			}
		}
	}
	for symbol, instance := range f.l.statics {
		if program.ProgramTypes[int(f.l.checker.GetTypeOfSymbol(symbol).Id())] {
			program.Classes[instance.class-1].ProgramRegion = true
		}
	}
	// Cached canonical closures cannot be moved after publication. Keep the ordinary
	// graph plan until their entire component can adopt before caching (06c).
	for _, function := range program.Functions {
		if function.ProgramRegion && function.NestedParent > 0 {
			program.ProgramTypes = ir.SortedMap[int, bool]{}
			for i := range program.Locals {
				program.Locals[i].ProgramRegion = false
			}
			for i := range program.Functions {
				program.Functions[i].ProgramRegion = false
			}
			for i := range program.Classes {
				program.Classes[i].ProgramRegion = false
			}
			selected = map[cycleNode]bool{}
			break
		}
	}
	plan := &programMembership{types: program.ProgramTypes, cells: map[int]bool{}, functions: map[int]bool{}, classes: map[int]bool{}, identities: programAllocationIdentities(program)}
	for node := range selected {
		if node.proven != nil && f.l.isLibraryType(node.proven, "Promise") {
			plan.refusal = &Refused{Where: l.result.Source, What: "a Promise whose payload reaches a Program member", Fix: "remove the captured strong back-reference or use Weak (adamic/cycle-capable)"}
		}
	}
	if len(plan.types) != 0 && plan.refusal == nil {
		plan.refusal = f.graphAdoptionBoundary()
	}

	for i, x := range program.Locals {
		plan.cells[i] = x.ProgramRegion
	}
	for i, x := range program.Functions {
		plan.functions[i] = x.ProgramRegion
	}
	for i, x := range program.Classes {
		plan.classes[i+1] = x.ProgramRegion
	}
	return plan
}

func scanLowerWithOptions(ctx context.Context, program *load.Program, options Options) (*ir.Program, error) {
	files := program.Files()
	if len(files) != 1 {
		return nil, fmt.Errorf("lower: stage 0 compiles a program from one entry file, got %d", len(files))
	}
	entry := files[0]
	// Stage 0 checks single-threaded, so one checker answers for every file.
	typeChecker, release := program.Checker(ctx, entry)
	defer release()

	var plan *programMembership
	if options.ProgramRegion {
		discovery, err := scanLowerChecked(program, typeChecker, entry, nil, true)
		if err != nil {
			return nil, err
		}
		plan = discovery.programPlan
	}
	built, err := scanLowerChecked(program, typeChecker, entry, plan, false, options.OwnershipQuery)
	if err != nil {
		return nil, err
	}

	return built.result, nil
}

func scanLowerChecked(program *load.Program, typeChecker *checker.Checker, entry *ast.SourceFile, plan *programMembership, discovery bool, ownershipQuery ...bool) (*lowering, error) {
	lowering := &lowering{program: program, checker: typeChecker, result: &ir.Program{}, this: -1, functionIndex: -1, entry: entry, programPlan: plan, programDiscovery: discovery}
	lowering.result.ProgramRegion = discovery || plan != nil
	if plan != nil {
		lowering.result.ProgramTypes = plan.types
	}

	// The base name only, so the same program emits the same C on every machine.
	lowering.result.Source = filepath.Base(program.FileName(entry))
	modules, err := lowering.moduleOrder(entry)
	if err != nil {
		return nil, err
	}
	lowering.noteInheritance(modules)
	if err := lowering.enumInitialization(modules); err != nil {
		return nil, err
	}
	lowering.noteAccessorNames(modules)
	lowering.noteOmittedOptionals(modules)
	if err := lowering.namespaceInitialization(modules); err != nil {
		return nil, err
	}
	for _, module := range modules {
		if err := lowering.refuse(module); err != nil {
			return nil, err
		}
	}
	// Link every declaration before lowering any function body, including across back edges.
	var declarations []*ast.Node
	for _, module := range modules {
		declarations = append(declarations, module.Statements.Nodes...)
	}
	if err := lowering.declareModule(declarations); err != nil {
		return nil, err
	}
	lowering.recordExports(entry)
	for _, module := range modules {
		body, err := lowering.moduleStatements(module)
		if err != nil {
			return nil, err
		}
		lowering.result.Main = append(lowering.result.Main, body...)
	}
	if lowering.unlowerable != nil {
		return nil, lowering.unlowerable
	}
	lowering.result.Main = append(lowering.forwarderValues, lowering.result.Main...)
	lowering.finishClassCalls()
	if err := lowering.finishAccessors(); err != nil {
		return nil, err
	}
	if err := lowering.exceptions(); err != nil {
		return nil, err
	}
	if err := lowering.checkLibrarySpreads(modules); err != nil {
		return nil, err
	}
	if err := lowering.checkAccessorSpreads(); err != nil {
		return nil, err
	}
	if err := flow.NormalizeAsync(lowering.result); err != nil {
		where := entry.AsNode()
		var located *flow.AsyncError
		if errors.As(err, &located) && located.Site > 0 && located.Site <= len(lowering.asyncSites) {
			where = lowering.asyncSites[located.Site-1]
		}
		return nil, &NotYet{Where: lowering.program.Where(where), What: err.Error()}
	}
	if lowering.result.AsyncEntry > 0 {
		// NormalizeAsync moved module bodies into one scheduler entry. Main now
		// holds only its generated invocation; retain boundaries of the actual Main.
		source := ir.SourceIdentity{Module: lowering.modulePath(entry), Declaration: []string{"module"}, Role: "async_module"}
		lowering.result.Functions[lowering.result.AsyncEntry-1].Source = source
		lowering.result.MainModules = []ir.MainModule{{Module: source.Module, Statements: len(lowering.result.Main)}}
	}
	if err := lowering.scanFindCycles(modules); err != nil {
		return nil, err
	}
	if len(lowering.asyncPayloadFailures) > 0 {
		return nil, lowering.asyncPayloadFailures[0]
	}
	if err := lowering.checkPromiseConsumption(); err != nil {
		return nil, err
	}
	if err := lowering.checkViewLogicalWrites(); err != nil {
		return nil, err
	}
	if discovery || plan != nil {
		if lowering.result.HasAsync() {
			return nil, &Refused{Where: program.Where(entry.AsNode()), What: "Program region with async work", Fix: "keep Program members in the one-shot synchronous CLI"}
		}
	}
	if plan != nil {
		if err := lowering.applyProgramMetadata(plan); err != nil {
			return nil, err
		}
	}
	if len(ownershipQuery) > 0 && ownershipQuery[0] && lowering.hasMovedParallelMap {
		for _, transfer := range fresh.QueryOwnershipTransfers(lowering.result) {
			if transfer.Answer.Verdict != fresh.Proven {
				where := lowering.moveSites[transfer.Site-1]
				return nil, lowering.moveRefused(where, "whole reachable ownership is not proven (ownership query "+transfer.Answer.Verdict.String()+": "+fmt.Sprint(transfer.Answer.Evidence)+")")
			}
		}
	}
	readiness(lowering.result)
	borrow(lowering.result)
	counters(lowering.result)
	// Last, once nothing more changes the program: its walk of the method tables is the one the
	// emitter reads too (ir.Program's MethodsNamed).
	if err := lowering.voidDiscards(); err != nil {
		return nil, err
	}
	return lowering, nil
}
