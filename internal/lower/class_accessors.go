package lower

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

func accessorMember(node *ast.Node) bool {
	return node.Kind == ast.KindGetAccessor || node.Kind == ast.KindSetAccessor
}
func classFunction(node *ast.Node) bool {
	return node.Kind == ast.KindMethodDeclaration || accessorMember(node)
}
func (l *lowering) methodKey(node *ast.Node, class int) string {
	if node.Name().Kind == ast.KindComputedPropertyName {
		return iteratorSlot
	}
	name := l.memberKey(node.Name(), class)
	if node.Kind == ast.KindGetAccessor {
		return "$get:" + name
	}
	if node.Kind == ast.KindSetAccessor {
		return "$set:" + name
	}
	return name
}

func (l *lowering) registerAccessor(class int, node *ast.Node, function int) {
	name := l.memberKey(node.Name(), class)
	metadata := &l.result.Classes[class-1]
	index := -1
	for i, accessor := range metadata.Accessors {
		if accessor.Name == name {
			index = i
		}
	}
	if index < 0 {
		index = len(metadata.Accessors)
		metadata.Accessors = append(metadata.Accessors, ir.Accessor{Name: name, Getter: -1, Setter: -1})
	}
	if node.Kind == ast.KindGetAccessor {
		metadata.Accessors[index].Getter = function
		if l.getterDeclarations == nil {
			l.getterDeclarations = map[int]*ast.Node{}
		}
		l.getterDeclarations[function] = node
	} else {
		metadata.Accessors[index].Setter = function
	}
}

func (l *lowering) superAccessor(target *ast.Node, valueNode *ast.Node) (ir.Expression, bool, error) {
	receiver := ast.SkipParentheses(target.AsPropertyAccessExpression().Expression)
	if receiver.Kind != ast.KindSuperKeyword {
		return nil, false, nil
	}
	if l.instance == nil || l.instance.base == nil {
		return nil, true, l.notYet(target, "super outside a derived accessor or method")
	}
	if useOfThisError := l.useOfThis(receiver); useOfThisError != nil {
		return nil, true, useOfThisError
	}
	prefix := "$get:"
	if valueNode != nil {
		prefix = "$set:"
	}
	index, exists := l.instance.base.methods[prefix+l.fieldName(target.Name())]
	if !exists {
		if l.instance.static {
			symbol := l.checker.GetSymbolAtLocation(target.Name())
			if symbol != nil &&
				len(symbol.Declarations) > 0 &&
				symbol.Declarations[0].Kind == ast.KindPropertyDeclaration {
				of, typeOfSymbolError := l.typeOfSymbol(target, symbol)
				if typeOfSymbolError != nil {
					return nil, true, typeOfSymbolError
				}
				if valueNode == nil {
					parent := l.staticBase(l.classNode.Parent)
					return ir.Property{
						Object: ir.Read{Local: l.staticGlobals[l.symbol(parent.Name())], Of: ir.Object, Checked: true},
						Name:   l.fieldName(target.Name()),
						Of:     of,
					}, true, nil
				}
				value, expressionError := l.expression(valueNode)
				if expressionError != nil {
					return nil, true, expressionError
				}
				index := len(l.result.Functions)
				self := len(l.result.Locals)
				l.result.Locals = append(
					l.result.Locals,
					ir.Local{Name: "this", Type: ir.Object, Function: index},
					ir.Local{Name: "value", Type: of, Function: index},
				)
				l.result.Functions = append(l.result.Functions, ir.Function{
					Name:       "super_static_set",
					Parameters: []int{self, self + 1},
					Body: []ir.Statement{ir.SetProperty{
						Object: ir.Read{Local: self, Of: ir.Object},
						Name:   l.fieldName(target.Name()),
						Value:  ir.Read{Local: self + 1, Of: of},
						Site:   l.staticWriteSite(target),
					}},
				})
				return ir.Call{
					Function:  index,
					Arguments: []ir.Expression{ir.Read{Local: l.this, Of: ir.Object}, fit(value, of)},
				}, true, nil
			}
		}
		return nil, true, l.notYet(target, "super of a member without the requested accessor")
	}
	arguments := []ir.Expression{ir.Read{Local: l.this, Of: ir.Object}}
	if valueNode != nil {
		value, expressionError := l.expression(valueNode)
		if expressionError != nil {
			return nil, true, expressionError
		}
		value = fit(value, l.result.Locals[l.result.Functions[index].Parameters[1]].Type)
		arguments = append(arguments, value)
	}
	return ir.Call{Function: index, Arguments: arguments, Returns: l.result.Functions[index].Returns}, true, nil
}

// Literal accessors are closures with an explicit receiver. The receiver is never captured by the
// accessor itself, so storing the closure in its object introduces no implicit self-cycle.
func (l *lowering) literalAccessor(owner, declaration *ast.Node) (ir.Expression, int, error) {
	index := len(l.result.Functions)
	l.result.Functions = append(l.result.Functions, ir.Function{Name: "accessor", Closure: true})
	self := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "this", Type: ir.Object, Function: index})
	l.noteLocal(self, l.concrete(l.checker.GetTypeAtLocation(owner)), owner)
	l.accessorCaptures = append(
		l.accessorCaptures,
		accessorCapture{holder: l.concrete(l.checker.GetTypeAtLocation(owner)), function: index},
	)
	outerInstance, outerUnset := l.instance, l.unsetUntil
	l.instance, l.unsetUntil = nil, 0
	l.closures = append(l.closures, index)
	lowerFunctionError := l.lowerFunction(index, declaration, self)
	l.closures = l.closures[:len(l.closures)-1]
	l.instance, l.unsetUntil = outerInstance, outerUnset
	if lowerFunctionError != nil {
		return nil, 0, lowerFunctionError
	}
	return ir.MakeClosure{Function: index}, index, nil
}

type accessorCapture struct {
	holder   *checker.Type
	function int
}

func (l *lowering) accessorLiteral(node *ast.Node) (ir.Expression, bool, error) {
	properties := node.AsObjectLiteralExpression().Properties.Nodes
	found := false
	for _, property := range properties {
		found = found || accessorMember(property)
	}
	if !found {
		return nil, false, nil
	}
	class := len(l.result.Classes) + 1
	l.result.Classes = append(
		l.result.Classes,
		ir.Class{Source: l.sourceIdentity(node), Name: "literal_accessors", Constructor: -1, Literal: true},
	)
	literal := ir.ObjectLiteral{Class: class}
	keys := map[string]bool{}
	for _, property := range properties {
		if property.Kind == ast.KindSpreadAssignment {
			return nil, true, l.notYet(property, "an accessor literal with a spread")
		}
		name := property.Name()
		if name == nil || (!ast.IsIdentifier(name) && name.Kind != ast.KindStringLiteral) {
			return nil, true, l.notYet(property, "a computed accessor literal name")
		}
		if strings.HasPrefix(name.Text(), "#accessor:") {
			return nil, true, l.notYet(
				name,
				"an accessor literal property name reserved for closure storage; rename the property",
			)
		}
		var value ir.Expression
		var err error
		if accessorMember(property) {
			var function int
			var literalAccessorError error
			value, function, literalAccessorError = l.literalAccessor(node, property)
			if literalAccessorError != nil {
				return nil, true, literalAccessorError
			}
			l.registerAccessor(class, property, function)
			storage := fmt.Sprintf("#accessor:%d", function)
			literal.Fields = append(literal.Fields, ir.Field{Name: storage, Value: value, Private: true})
			if !keys[name.Text()] {
				of, typeOfError := l.typeOf(name)
				if typeOfError != nil || slotless(of) {
					return nil, true, l.notYet(property, "an accessor without a native property representation")
				}
				l.result.Classes[class-1].PublicFields = append(
					l.result.Classes[class-1].PublicFields,
					ir.Field{Name: name.Text(), Value: zeroValue(of)},
				)
				keys[name.Text()] = true
			}
			continue
		}
		switch property.Kind {
		case ast.KindPropertyAssignment:
			value, err = l.expression(property.AsPropertyAssignment().Initializer)
		case ast.KindShorthandPropertyAssignment:
			value, err = l.shorthand(property)
		default:
			return nil, true, l.notYet(property, "a non-accessor method in an accessor literal")
		}
		if err != nil {
			return nil, true, err
		}
		if of := l.declaredField(node, name.Text()); of != 0 {
			value = fit(value, of)
		}
		if slotless(value.Type()) {
			return nil, true, l.notYet(property, "an accessor literal's field without a native slot")
		}
		field := ir.Field{Name: name.Text(), Value: value}
		literal.Fields = append(literal.Fields, field)
		if !keys[name.Text()] {
			l.result.Classes[class-1].PublicFields = append(l.result.Classes[class-1].PublicFields, field)
			keys[name.Text()] = true
		}
	}
	l.result.Classes[class-1].Fields = literal.Fields
	return literal, true, nil
}

// accessorReach is which allocations own an accessor of each name, judged per allocation over every
// module by the may-flow graph the lazy view rule reads (view_lazy.go), never by the name alone
// (Compiler and Lower, Nov 1, #c7w6j0y). It numbers the allocation sites of a copy of the program,
// so the real program's are still graphFlows' to number (cycles.go).
type accessorReach struct {
	graph  *allocationFlowGraph
	owners map[int]map[string]bool
}

func newAccessorReach(program *ir.Program) *accessorReach {
	numbered := *program
	numbered.Functions = append([]ir.Function(nil), program.Functions...)
	assignAllocationSites(&numbered)
	reach := &accessorReach{graph: newAllocationFlowGraph(&numbered), owners: map[int]map[string]bool{}}
	own := func(node any) bool {
		literal, isLiteral := node.(ir.ObjectLiteral)
		if !isLiteral || literal.Class == 0 || len(literal.GraphTypes) == 0 {
			return true
		}
		site := literal.GraphTypes[len(literal.GraphTypes)-1]
		// An object owns its class's accessors and every base's, as adamic_accessor_find finds them.
		for class := literal.Class; class != 0; class = program.Classes[class-1].Base {
			for _, accessor := range program.Classes[class-1].Accessors {
				if reach.owners[site] == nil {
					reach.owners[site] = map[string]bool{}
				}
				reach.owners[site][accessor.Name] = true
			}
		}
		return true
	}
	walk(numbered.Main, own)
	for _, function := range numbered.Functions {
		walk(function.Body, own)
	}
	return reach
}

// owner is whether one of the allocations owns an accessor named name. Unknown reach counts as
// reaching one, so it fails closed.
func (reach *accessorReach) owner(allocations ir.AllocationSet, name string) bool {
	if allocations.Unknown {
		return true
	}
	for _, site := range allocations.Sites {
		if reach.owners[site][name] {
			return true
		}
	}
	return false
}

// contractGetters is a view's origin with every getter in the program named as one of its contract's
// fields. A field read of an object a getter owns reads the getter's slot, never its value, so the
// view is refused when an allocation owning one of them can reach the origin, through the fields its
// contract checks too, or when the reach is unknown (Compiler and Lower, Nov 1, #c7w6j0y). Judged
// once lowering is done, since what reaches the origin is the whole program's to say. A getter no
// class or accessor literal owns (an interface's) is on no allocation, so no reach brings it.
type contractGetters struct {
	origin  ir.Expression
	getters []*ast.Node
}

func (l *lowering) refuseReachedContractGetters(reach *accessorReach, owned map[string]bool) error {
	for _, contract := range l.contractGetters {
		viewed, unknown := reach.graph.viewedFrom([]ir.Expression{contract.origin}, true)
		for _, getter := range contract.getters {
			if !owned[getter.Name().Text()] {
				continue
			}
			reached := unknown
			for site := range viewed {
				reached = reached || reach.owners[site][getter.Name().Text()]
			}
			if reached {
				return l.notYet(getter, "a getter in a checked field contract")
			}
		}
	}
	return nil
}

// Property reads and writes become dispatch functions only where their receiver can reach an
// allocation owning an accessor of that name, or its reach is unknown (accessorReach). A structural
// view still invokes the actual descriptor; plain fields keep their load and their view.
func (l *lowering) finishAccessors() error {
	names := map[string]bool{}
	for _, class := range l.result.Classes {
		for _, accessor := range class.Accessors {
			names[accessor.Name] = true
		}
	}
	if len(names) == 0 {
		return nil
	}
	reach := newAccessorReach(l.result)
	if refuseReachedContractGettersError := l.refuseReachedContractGetters(
		reach,
		names,
	); refuseReachedContractGettersError != nil {
		return refuseReachedContractGettersError
	}
	groups := map[string][]int{}
	for _, class := range l.result.Classes {
		for _, accessor := range class.Accessors {
			if accessor.Getter >= 0 {
				groups["get:"+accessor.Name] = append(groups["get:"+accessor.Name], accessor.Getter)
			}
			if accessor.Setter >= 0 {
				groups["set:"+accessor.Name] = append(groups["set:"+accessor.Name], accessor.Setter)
			}
		}
	}
	dispatchers := map[string]int{}
	var failure error
	dispatch := func(name string, of ir.Type, setter bool, site int) int {
		kind := "get:"
		if setter {
			kind = "set:"
		}
		key := fmt.Sprintf("%s%s:%d:%d", kind, name, of, site)
		if index, exists := dispatchers[key]; exists {
			return index
		}
		for _, class := range l.result.Classes {
			for _, accessor := range class.Accessors {
				if accessor.Name == name && ((setter && accessor.Setter < 0) || (!setter && accessor.Getter < 0)) {
					failure = &NotYet{
						Where: l.result.Source,
						What:  "a property name shared with a descriptor missing the requested getter or setter",
					}
					return 0
				}
			}
		}
		targets := groups[kind+name]
		if len(targets) == 0 {
			failure = &NotYet{
				Where: l.result.Source,
				What:  "reading a setter-only property or writing a getter-only property",
			}
			return 0
		}
		for _, target := range targets {
			actual := l.result.Functions[target].Returns
			if setter {
				actual = l.result.Locals[l.result.Functions[target].Parameters[1]].Type
			}
			if actual != of {
				failure = &NotYet{
					Where: l.result.Source,
					What:  "accessors sharing a name with different native representations",
				}
				return 0
			}
		}
		index := len(l.result.Functions)
		dispatchers[key] = index
		function := ir.Function{Name: "property_" + kind + name, Returns: of}
		self := len(l.result.Locals)
		l.result.Locals = append(l.result.Locals, ir.Local{Name: "object", Type: ir.Object, Function: index})
		function.Parameters = []int{self}
		object := ir.Read{Local: self, Of: ir.Object}
		call := ir.Call{
			Function:  targets[0],
			Arguments: []ir.Expression{object},
			Returns:   of,
			Virtual:   -1,
			Accessor:  name,
		}
		l.result.MethodTargets[targets[0]] = targets
		if setter {
			function.Returns, call.Returns, call.Setter = 0, 0, true
			incoming := len(l.result.Locals)
			l.result.Locals = append(l.result.Locals, ir.Local{Name: "value", Type: of, Function: index})
			function.Parameters = append(function.Parameters, incoming)
			value := ir.Read{Local: incoming, Of: of}
			call.Arguments = append(call.Arguments, value)
			function.Body = []ir.Statement{ir.If{
				Condition: ir.HasAccessor{Object: object, Name: name},
				Then:      []ir.Statement{ir.Evaluate{Value: call}},
				Else:      []ir.Statement{ir.SetProperty{Object: object, Name: name, Value: value, Site: site}},
			}}
		} else {
			function.Body = []ir.Statement{ir.If{
				Condition: ir.HasAccessor{Object: object, Name: name},
				Then:      []ir.Statement{ir.Return{Value: call}},
			}, ir.Return{Value: ir.Property{Object: object, Name: name, Of: of}}}
		}
		l.result.Functions = append(l.result.Functions, function)
		return index
	}
	var transform func(reflect.Value) reflect.Value
	transform = func(value reflect.Value) reflect.Value {
		if !value.IsValid() {
			return value
		}
		if value.Kind() == reflect.Interface {
			if value.IsNil() {
				return value
			}
			// The receiver as lowering made it, before a read inside it became a dispatcher call
			// the flow graph has no function for.
			original := value.Elem().Interface()
			mapped := transform(value.Elem())
			node := mapped.Interface()
			switch expression := node.(type) {
			case ir.ObjectLiteral:
				if expression.Spread != nil {
					expression.NoReuse = true
					node = expression
				}
			case ir.Property:
				if names[expression.Name] &&
					reach.owner(reach.graph.ReachingAllocations(original.(ir.Property).Object), expression.Name) {
					if expression.Optional {
						failure = &NotYet{Where: l.result.Source, What: "an optional accessor read"}
						return value
					}
					// The dispatcher's fallback is an unchecked load, so a checked read can't go
					// through it: refuse at the read rather than drop its check (#c7w6j0y).
					if expression.View != "" && l.result.CheckedField(expression.ViewReceiverTypeID, expression.Name) {
						failure = &NotYet{
							Where: expression.ViewWhere,
							What:  "a checked view read whose receiver may own an accessor of the same name",
						}
						return value
					}
					index := dispatch(expression.Name, expression.Of, false, 0)
					node = ir.Call{
						Function:  index,
						Arguments: []ir.Expression{expression.Object},
						Returns:   expression.Of,
					}
				}
			case ir.SetProperty:
				if names[expression.Name] &&
					!expression.Define &&
					reach.owner(reach.graph.ReachingAllocations(original.(ir.SetProperty).Object), expression.Name) {
					// As a read: the fallback store drops the checked write's conversion. The write is
					// located at its receiver when lowering recorded its site (fresh.go's writeSite).
					if l.result.CheckedField(expression.ViewReceiverTypeID, expression.Name) {
						where := l.result.Source
						if expression.Site > 0 && expression.Site <= len(l.writeSites) {
							where = l.program.Where(l.writeSites[expression.Site-1].node)
						}
						failure = &NotYet{Where: where, What: "a checked view write of " +
							expression.Name +
							" whose receiver may own an accessor of the same name"}
						return value
					}
					index := dispatch(expression.Name, expression.Value.Type(), true, expression.Site)
					node = ir.Evaluate{Value: ir.Call{
						Function:  index,
						Arguments: []ir.Expression{expression.Object, expression.Value},
					}}
				}
			}
			result := reflect.New(value.Type()).Elem()
			result.Set(reflect.ValueOf(node))
			return result
		}
		switch value.Kind() {
		case reflect.Struct:
			result := reflect.New(value.Type()).Elem()
			for i := 0; i < value.NumField(); i++ {
				result.Field(i).Set(transform(value.Field(i)))
			}
			return result
		case reflect.Slice:
			if value.IsNil() {
				return value
			}
			result := reflect.MakeSlice(value.Type(), value.Len(), value.Len())
			for i := 0; i < value.Len(); i++ {
				result.Index(i).Set(transform(value.Index(i)))
			}
			return result
		}
		return value
	}
	// Generated fallback loads must not be rewritten into their own dispatchers.
	count := len(l.result.Functions)
	l.result.Main = transform(reflect.ValueOf(l.result.Main)).Interface().([]ir.Statement)
	for index := 0; index < count; index++ {
		l.result.Functions[index].Body = transform(
			reflect.ValueOf(l.result.Functions[index].Body),
		).Interface().([]ir.Statement)
	}
	return failure
}

func accessorSymbol(symbol *ast.Symbol) bool {
	for _, declaration := range symbol.Declarations {
		if accessorMember(declaration) {
			return true
		}
	}
	return false
}
func (l *lowering) hasAccessorStorage(proven *checker.Type) bool {
	for _, property := range l.checker.GetPropertiesOfType(l.concrete(proven)) {
		if accessorSymbol(property) {
			return true
		}
	}
	return false
}

// Until spread callbacks are explicit calls in the flow graph, their exception edges must not be lost.
func (l *lowering) checkAccessorSpreads() error {
	setterOnly := false
	for _, class := range l.result.Classes {
		for _, accessor := range class.Accessors {
			setterOnly = setterOnly || (class.Literal && accessor.Getter < 0)
		}
	}
	throwing := false
	for _, class := range l.result.Classes {
		for _, accessor := range class.Accessors {
			if class.Literal &&
				accessor.Getter >= 0 &&
				(l.libraryFailure(l.result.Functions[accessor.Getter].Body, map[int]bool{}) != "") {
				throwing = true
			}
		}
	}
	if !throwing && !setterOnly && len(l.staticGlobals) == 0 {
		return nil
	}
	spread := false
	staticSpread := false
	check := func(node any) bool {
		if literal, ok := node.(ir.ObjectLiteral); ok && literal.Spread != nil {
			spread = true
			staticSpread = staticSpread || !plainDataSpreadSource(literal.Spread)
		}
		return true
	}
	walk(l.result.Main, check)
	for _, function := range l.result.Functions {
		walk(function.Body, check)
	}
	if staticSpread && len(l.staticGlobals) > 0 {
		return &NotYet{Where: l.result.Source, What: "spreading in a program with static constructor objects"}
	}
	if spread && setterOnly {
		return &NotYet{Where: l.result.Source, What: "spreading a setter-only property, whose read value is undefined"}
	}
	if spread && throwing {
		return &NotYet{Where: l.result.Source, What: "spreading an accessor literal whose getter may throw"}
	}
	return nil
}

func (l *lowering) noteAccessorNames(modules []*ast.SourceFile) {
	l.accessorNames = map[string]bool{}
	for _, module := range modules {
		var visit ast.Visitor
		visit = func(node *ast.Node) bool {
			if node.Kind == ast.KindGetAccessor && node.Name() != nil {
				l.accessorNames[node.Name().Text()] = true
			}
			return node.ForEachChild(visit)
		}
		module.AsNode().ForEachChild(visit)
	}
}

func (l *lowering) validateAccessorSignature(member *ast.Node, function int) error {
	signature := l.result.Functions[function]
	if member.Kind == ast.KindGetAccessor && (signature.Returns == 0 || signature.Returns == ir.MaybeBoolean) {
		return l.notYet(member, "a getter without a native adamic_value representation")
	}
	if member.Kind == ast.KindSetAccessor && l.result.Locals[signature.Parameters[1]].Type == ir.MaybeBoolean {
		return l.notYet(member, "a setter whose input needs two native words")
	}
	return nil
}

// A fresh data literal owns exactly these enumerable slots; reading a class field
// while building it does not put the class constructor descriptor in the copy.
func plainDataSpreadSource(source ir.Expression) bool {
	literal, ok := source.(ir.ObjectLiteral)
	return ok && literal.Class == 0 && literal.Spread == nil && !literal.Record
}
