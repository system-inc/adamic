package lower

import (
	"strconv"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

func (l *lowering) staticBase(declaration *ast.Node) *ast.Node {
	for _, clause := range nodesOf(declaration.AsClassDeclaration().HeritageClauses) {
		if clause.AsHeritageClause().Token != ast.KindExtendsKeyword {
			continue
		}
		for _, node := range clause.AsHeritageClause().Types.Nodes {
			symbol := l.symbol(ast.SkipParentheses(node.AsExpressionWithTypeArguments().Expression))
			if symbol != nil && len(symbol.Declarations) == 1 && symbol.Declarations[0].Kind == ast.KindClassDeclaration {
				return symbol.Declarations[0]
			}
		}
	}
	return nil
}
func (l *lowering) needsStatics(declaration *ast.Node) bool {
	for _, member := range declaration.Members() {
		if member.Kind == ast.KindClassStaticBlockDeclaration || ast.HasSyntacticModifier(member, ast.ModifierFlagsStatic) {
			return true
		}
	}
	base := l.staticBase(declaration)
	return base != nil && l.needsStatics(base)
}
func (l *lowering) staticClass(proven *checker.Type, fallback *ast.Node) *ast.Node {
	symbol := proven.Symbol()
	if symbol != nil {
		for _, declaration := range symbol.Declarations {
			if declaration.Kind == ast.KindClassDeclaration {
				return declaration
			}
		}
	}
	return fallback
}

// Static layouts belong to source classes, independent of instance type arguments.
func (l *lowering) staticInstance(declaration *ast.Node) (*instance, error) {
	if err := l.staticClassDeclaration(declaration); err != nil {
		return nil, err
	}
	symbol := l.symbol(declaration.Name())
	if existing := l.statics[symbol]; existing != nil {
		return existing, nil
	}
	if l.statics == nil {
		l.statics = map[*ast.Symbol]*instance{}
	}
	if l.staticGlobals == nil {
		l.staticGlobals = map[*ast.Symbol]int{}
	}
	if _, exists := l.staticGlobals[symbol]; !exists {
		local, err := l.declareLocal(declaration.Name())
		if err != nil {
			return nil, err
		}
		l.result.Locals[local].Global = true
		l.staticStorage(declaration)
	}
	var base *instance
	if parent := l.staticBase(declaration); parent != nil {
		var err error
		base, err = l.staticInstance(parent)
		if err != nil {
			return nil, err
		}
	}
	if base != nil {
		if err := l.checkMemberOverrides(declaration, l.checker.GetTypeOfSymbol(symbol), l.checker.GetTypeOfSymbol(l.symbol(l.staticBase(declaration).Name())), true); err != nil {
			return nil, err
		}
	}
	lowered := &instance{static: true, constructor: -1, initializer: -1, superReady: -1, class: len(l.result.Classes) + 1, base: base, methods: map[string]int{}, slots: map[string]int{}}
	metadata := ir.Class{Name: declaration.Name().Text() + "_static", Constructor: -1, Static: true}
	if base != nil {
		parent := l.result.Classes[base.class-1]
		metadata.Base = base.class
		metadata.Fields = append(metadata.Fields, parent.Fields...)
		metadata.Methods = append(metadata.Methods, parent.Methods...)
		metadata.StaticFlags = append(metadata.StaticFlags, parent.StaticFlags...)
		for name, function := range base.methods {
			lowered.methods[name] = function
		}
		for name, slot := range base.slots {
			lowered.slots[name] = slot
		}
	}
	metadata.OwnStart = len(metadata.Fields)
	if base != nil {
		metadata.Fields = append(metadata.Fields, ir.Field{Name: "#static-parent:" + strconv.Itoa(lowered.class), Value: ir.Undefined{Of: ir.Object}, Private: true})
		metadata.StaticFlags = append(metadata.StaticFlags, 0)
		metadata.StaticParent = len(metadata.Fields)
	}
	for _, member := range declaration.Members() {
		if !ast.HasSyntacticModifier(member, ast.ModifierFlagsStatic) {
			continue
		}
		if member.Name() == nil || (!ast.IsIdentifier(member.Name()) && member.Name().Kind != ast.KindPrivateIdentifier) {
			return nil, l.notYet(member, "a static member without an identifier name")
		}
		if member.Kind == ast.KindPropertyDeclaration {
			if ast.HasSyntacticModifier(member, ast.ModifierFlagsAmbient|ast.ModifierFlagsAbstract) {
				return nil, l.notYet(member, "a declare or abstract static field")
			}
			of, err := l.typeOf(member.Name())
			if err != nil {
				return nil, err
			}
			if slotless(of) {
				return nil, l.notYet(member, "a static field without a native slot")
			}
			if member.AsPropertyDeclaration().Initializer == nil && !l.includesUndefined(l.checker.GetTypeAtLocation(member.Name())) {
				return nil, l.notYet(member, "an uninitialized nonnullable static field; initialize it at its declaration")
			}
			value := zeroValue(of)
			if of.IsReference() {
				value = ir.Undefined{Of: of}
			}
			field := ir.Field{Name: memberKey(member.Name(), lowered.class), Value: value, Private: member.Name().Kind == ast.KindPrivateIdentifier}
			slot := -1
			for i, previous := range metadata.Fields {
				if previous.Name == field.Name {
					slot = i
					if previous.Value.Type() != of {
						return nil, l.notYet(member, "a static field override with a different native representation")
					}
					break
				}
			}
			if slot < 0 {
				slot = len(metadata.Fields)
				metadata.Fields = append(metadata.Fields, field)
				metadata.StaticFlags = append(metadata.StaticFlags, 0)
				metadata.Fields = append(metadata.Fields, ir.Field{Name: "#static-present:" + field.Name, Value: ir.NumberConstant{}, Private: true})
				metadata.StaticFlags = append(metadata.StaticFlags, 0)
				if !field.Private {
					metadata.StaticFlags[slot] = len(metadata.Fields)
				}
			}
			if !field.Private {
				metadata.PublicFields = append(metadata.PublicFields, field)
			}
		}
	}
	l.result.Classes = append(l.result.Classes, metadata)
	l.statics[symbol] = lowered
	for _, member := range declaration.Members() {
		if !classFunction(member) || !ast.HasSyntacticModifier(member, ast.ModifierFlagsStatic) {
			continue
		}
		key := methodKey(member, lowered.class)
		function := len(l.result.Functions)
		lowered.methods[key] = function
		meta := &l.result.Classes[lowered.class-1]
		slot, exists := lowered.slots[key]
		if !exists {
			slot = len(meta.Methods)
			lowered.slots[key] = slot
			meta.Methods = append(meta.Methods, -1)
		}
		meta.Methods[slot] = function
		if accessorMember(member) {
			l.registerAccessor(lowered.class, member, function)
		}
		l.result.Functions = append(l.result.Functions, ir.Function{Name: metadata.Name + "_" + key})
	}
	outerInstance, outerType, outerNode, outerLocals := l.instance, l.classType, l.classNode, l.locals
	outerSubstitution, outerMapper := l.substitution, l.typeMapper
	l.instance, l.classType, l.classNode = lowered, l.checker.GetTypeOfSymbol(symbol), declaration.Name()
	l.substitution, l.typeMapper = nil, nil
	l.locals = map[*ast.Symbol]int{}
	for name, local := range outerLocals {
		if l.result.Locals[local].Global {
			l.locals[name] = local
		}
	}
	defer func() {
		l.instance, l.classType, l.classNode, l.locals = outerInstance, outerType, outerNode, outerLocals
		l.substitution, l.typeMapper = outerSubstitution, outerMapper
	}()
	for _, member := range declaration.Members() {
		if !classFunction(member) || !ast.HasSyntacticModifier(member, ast.ModifierFlagsStatic) {
			continue
		}
		function := lowered.methods[methodKey(member, lowered.class)]
		if err := l.signature(function, member, l.thisLocal(function)); err != nil {
			return nil, err
		}
		if err := l.validateAccessorSignature(member, function); err != nil {
			return nil, err
		}
	}
	for _, member := range declaration.Members() {
		if !classFunction(member) || !ast.HasSyntacticModifier(member, ast.ModifierFlagsStatic) {
			continue
		}
		if err := l.lowerFunction(lowered.methods[methodKey(member, lowered.class)], member, -1); err != nil {
			return nil, err
		}
	}
	return lowered, nil
}

// Allocation installs methods and accessors before the ordered field and block initializers run.
func (l *lowering) staticDeclaration(declaration *ast.Node) ([]ir.Statement, error) {
	if err := l.staticClassDeclaration(declaration); err != nil {
		return nil, err
	}
	if !l.needsStatics(declaration) && l.statics[l.symbol(declaration.Name())] == nil {
		return nil, nil
	}
	lowered, err := l.staticInstance(declaration)
	if err != nil {
		return nil, err
	}
	fields := append([]ir.Field{}, l.result.Classes[lowered.class-1].Fields...)
	if lowered.base != nil {
		parent := l.staticBase(declaration)
		slot := l.result.Classes[lowered.class-1].StaticParent - 1
		fields[slot].Value = ir.Read{Local: l.staticGlobals[l.symbol(parent.Name())], Of: ir.Object, Checked: true}
	}
	object := ir.Read{Local: l.staticGlobals[l.symbol(declaration.Name())], Of: ir.Object}
	statements := []ir.Statement{ir.Declare{Local: object.Local, Value: ir.ObjectLiteral{Fields: fields, Class: lowered.class}}}
	outerInstance, outerType, outerNode := l.instance, l.classType, l.classNode
	l.instance, l.classType, l.classNode = lowered, l.checker.GetTypeOfSymbol(l.symbol(declaration.Name())), declaration.Name()
	defer func() { l.instance, l.classType, l.classNode = outerInstance, outerType, outerNode }()
	available := map[string]bool{}
	if lowered.base != nil {
		for _, field := range l.result.Classes[lowered.base.class-1].Fields {
			available[field.Name] = true
		}
	}
	for _, member := range declaration.Members() {
		if member.Kind != ast.KindClassStaticBlockDeclaration && !(member.Kind == ast.KindPropertyDeclaration && ast.HasSyntacticModifier(member, ast.ModifierFlagsStatic)) {
			continue
		}
		{
			var body *ast.Node
			if member.Kind == ast.KindClassStaticBlockDeclaration {
				body = member.AsClassStaticBlockDeclaration().Body
			} else {
				body = member.AsPropertyDeclaration().Initializer
			}
			if err := l.staticInitializationReads(body, declaration, available, map[*ast.Node]uint8{}, true); err != nil {
				return nil, err
			}
		}
		index := len(l.result.Functions)
		l.result.Functions = append(l.result.Functions, ir.Function{Name: declaration.Name().Text() + "_static_initialize"})
		self := l.thisLocal(index)
		l.result.Functions[index].Parameters = []int{self}
		outerThis, outerFunction, outerIndex, outerLocals := l.this, l.function, l.functionIndex, l.locals
		function := l.result.Functions[index]
		l.this, l.function, l.functionIndex = self, &function, index
		l.locals = map[*ast.Symbol]int{}
		for symbol, local := range outerLocals {
			if l.result.Locals[local].Global {
				l.locals[symbol] = local
			}
		}
		if member.Kind == ast.KindClassStaticBlockDeclaration {
			function.Body, err = l.statements(member.AsClassStaticBlockDeclaration().Body.AsBlock().Statements.Nodes)
		} else {
			of, _ := l.typeOf(member.Name())
			value := zeroValue(of)
			initializer := member.AsPropertyDeclaration().Initializer
			if initializer != nil {
				value, err = l.expression(initializer)
				value = fit(value, of)
			} else if of.IsReference() {
				value = ir.Undefined{Of: of}
			}
			if err == nil {
				function.Body = []ir.Statement{ir.SetProperty{Object: ir.Read{Local: self, Of: ir.Object}, Name: memberKey(member.Name(), lowered.class), Value: value, Define: true, Site: l.staticWriteSite(declaration.Name())}}
			}
			available[member.Name().Text()] = true
		}
		l.this, l.function, l.functionIndex, l.locals = outerThis, outerFunction, outerIndex, outerLocals
		if err != nil {
			return nil, err
		}
		l.result.Functions[index] = function
		statements = append(statements, ir.Evaluate{Value: ir.Call{Function: index, Arguments: []ir.Expression{object}}})
	}
	statements = append(statements, ir.Declare{Local: l.locals[l.symbol(declaration.Name())], Value: object})
	return statements, nil
}

func (l *lowering) isStaticType(proven *checker.Type) bool {
	return proven != nil && len(l.checker.GetSignaturesOfType(proven, checker.SignatureKindConstruct)) > 0 && l.staticClass(proven, nil) != nil
}
func (l *lowering) staticWriteSite(node *ast.Node) int {
	l.writeSites = append(l.writeSites, writeSite{holder: l.classType, node: node})
	return len(l.writeSites)
}

// JavaScript leaves future fields undefined. A nonnullable native slot cannot pretend that
// value is its declared type, including when an initializer calls a helper that reads it.
func (l *lowering) staticInitializationReads(node, declaration *ast.Node, available map[string]bool, seen map[*ast.Node]uint8, receiver bool) error {
	pending := false
	for _, member := range declaration.Members() {
		if member.Kind == ast.KindPropertyDeclaration && ast.HasSyntacticModifier(member, ast.ModifierFlagsStatic) && !available[member.Name().Text()] {
			pending = true
		}
	}
	state := uint8(1)
	if receiver {
		state = 2
	}
	if node == nil || seen[node]&state != 0 {
		return nil
	}
	seen[node] |= state
	var failure error
	var visit ast.Visitor
	visit = func(child *ast.Node) bool {
		if failure != nil {
			return true
		}
		if ast.IsPartOfTypeNode(child) {
			return false
		}
		if child.Kind == ast.KindPropertyAccessExpression {
			access := child.AsPropertyAccessExpression()
			object := ast.SkipParentheses(access.Expression)
			current := receiver && object.Kind == ast.KindThisKeyword
			if ast.IsIdentifier(object) {
				current = l.symbol(object) == l.symbol(declaration.Name())
			}
			if current {
				symbol := l.checker.GetSymbolAtLocation(child.Name())
				if symbol != nil && len(symbol.Declarations) > 0 {
					member := symbol.Declarations[0]
					if member.Kind == ast.KindPropertyDeclaration && !available[member.Name().Text()] {
						failure = &Refused{Where: l.program.Where(child), What: "a static field read before its initializer has run", Fix: "declare the field earlier, or move the read after that field's initialization"}
						return true
					}
					if accessorMember(member) {
						for _, candidate := range symbol.Declarations {
							if accessorMember(candidate) {
								failure = l.staticInitializationReads(candidate.Body(), declaration, available, seen, true)
								if failure != nil {
									break
								}
							}
						}
						if failure != nil {
							return true
						}
					}
				}
			}
		}
		if ast.IsIdentifier(child) && l.symbol(child) == l.symbol(declaration.Name()) && !ast.IsDeclarationName(child) {
			if !insideClass(child, declaration) {
				failure = &Refused{Where: l.program.Where(child), What: "an outside helper reading a class during its declaration (the temporal dead zone)", Fix: "run the helper after the class declaration, or read this inside a static method"}
				return true
			}
			parent := child.Parent
			if pending && (parent == nil || parent.Kind != ast.KindPropertyAccessExpression || parent.AsPropertyAccessExpression().Expression != child) {
				failure = &Refused{Where: l.program.Where(child), What: "a constructor object escaping before static fields are initialized", Fix: "publish the class value after the last static field initializer"}
				return true
			}
		}
		if child.Kind == ast.KindThisKeyword && receiver {
			parent := child.Parent
			if pending && (parent == nil || parent.Kind != ast.KindPropertyAccessExpression || parent.AsPropertyAccessExpression().Expression != child) {
				failure = &Refused{Where: l.program.Where(child), What: "a constructor object escaping before static fields are initialized", Fix: "publish this after the last static field initializer"}
				return true
			}
		}
		if child.Kind == ast.KindCallExpression {
			callee := ast.SkipParentheses(child.AsCallExpression().Expression)
			symbol := l.checker.GetSymbolAtLocation(callee)
			if ast.IsIdentifier(callee) {
				symbol = l.symbol(callee)
			}
			known := false
			if symbol != nil && len(symbol.Declarations) > 0 {
				target := symbol.Declarations[0]
				known = load.IsLibrary(ast.GetSourceFileOfNode(target)) || load.IsPrelude(ast.GetSourceFileOfNode(target))
				if target.Kind == ast.KindFunctionDeclaration {
					known = true
					failure = l.staticInitializationReads(target.Body(), declaration, available, seen, false)
				}
				if target.Kind == ast.KindMethodDeclaration && ast.HasSyntacticModifier(target, ast.ModifierFlagsStatic) {
					known = true
					current := false
					if callee.Kind == ast.KindPropertyAccessExpression {
						object := ast.SkipParentheses(callee.AsPropertyAccessExpression().Expression)
						current = (receiver && (object.Kind == ast.KindThisKeyword || object.Kind == ast.KindSuperKeyword)) || (ast.IsIdentifier(object) && l.symbol(object) == l.symbol(declaration.Name()))
					}
					failure = l.staticInitializationReads(target.Body(), declaration, available, seen, current)
				}
			}
			if known && symbol != nil && len(symbol.Declarations) > 0 && (load.IsLibrary(ast.GetSourceFileOfNode(symbol.Declarations[0])) || load.IsPrelude(ast.GetSourceFileOfNode(symbol.Declarations[0]))) {
				for _, argument := range nodesOf(child.AsCallExpression().Arguments) {
					if len(l.checker.GetSignaturesOfType(l.checker.GetTypeAtLocation(argument), checker.SignatureKindCall)) == 0 {
						continue
					}
					if ast.IsFunctionLike(ast.SkipParentheses(argument)) {
						continue
					}
					target := l.symbol(ast.SkipParentheses(argument))
					if target != nil && len(target.Declarations) == 1 && target.Declarations[0].Kind == ast.KindFunctionDeclaration {
						failure = l.staticInitializationReads(target.Declarations[0].Body(), declaration, available, seen, false)
					} else {
						failure = l.notYet(argument, "an indirect callback before static fields are initialized; move it after the field initializers")
					}
					if failure != nil {
						break
					}
				}
			}
			if !known && failure == nil {
				failure = l.notYet(child, "an indirect call before static fields are initialized; move it after the field initializers")
			}
			if failure != nil {
				return true
			}
		}
		return child.ForEachChild(visit)
	}
	visit(node)
	return failure
}

// Constructor built-ins have function semantics, not ordinary class storage.
func (l *lowering) staticProperty(node *ast.Node) error {
	receiver := node.AsPropertyAccessExpression().Expression
	if !l.isStaticType(l.checker.GetTypeAtLocation(receiver)) && !(l.instance != nil && l.instance.static && ast.SkipParentheses(receiver).Kind == ast.KindThisKeyword) {
		return nil
	}
	symbol := l.checker.GetSymbolAtLocation(node.Name())
	if symbol != nil {
		for _, member := range symbol.Declarations {
			if member.Parent != nil && member.Parent.Kind == ast.KindClassDeclaration && ast.HasSyntacticModifier(member, ast.ModifierFlagsStatic) {
				return nil
			}
		}
	}
	return l.notYet(node, "an implicit constructor property such as prototype, name or length")
}

func (l *lowering) staticClassDeclaration(declaration *ast.Node) error {
	if ast.HasSyntacticModifier(declaration, ast.ModifierFlagsAmbient) {
		return l.notYet(declaration, "a declare class without a runtime definition")
	}
	for _, clause := range nodesOf(declaration.AsClassDeclaration().HeritageClauses) {
		if clause.AsHeritageClause().Token == ast.KindExtendsKeyword && l.staticBase(declaration) == nil {
			return l.notYet(clause, "a computed class base; name the base class directly")
		}
	}
	return nil
}

func (l *lowering) staticStorage(declaration *ast.Node) {
	symbol := l.symbol(declaration.Name())
	if _, exists := l.staticGlobals[symbol]; exists {
		return
	}
	local := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: declaration.Name().Text() + "_class", Type: ir.Object, Function: -1, Global: true})
	l.staticGlobals[symbol] = local
	l.noteLocal(local, l.checker.GetTypeOfSymbol(symbol), declaration.Name())
}
func insideClass(node, declaration *ast.Node) bool {
	for parent := node.Parent; parent != nil; parent = parent.Parent {
		if parent.Kind == ast.KindClassDeclaration {
			return parent == declaration
		}
	}
	return false
}
func (l *lowering) staticClassRead(node *ast.Node) (ir.Expression, bool) {
	symbol := l.symbol(node)
	declaration := l.classes[symbol]
	local, exists := l.staticGlobals[symbol]
	if !exists || !insideClass(node, declaration) {
		return nil, false
	}
	return ir.Read{Local: local, Of: ir.Object, Checked: true}, true
}

// Looking up the constructor precedes evaluating new's arguments. Instance allocators
// remain independent functions, so make that declaration-readiness check explicit here.
func (l *lowering) staticConstruct(node, declaration *ast.Node, constructor int, arguments []ir.Expression) ir.Expression {
	symbol := l.symbol(declaration.Name())
	_, hasStorage := l.staticGlobals[symbol]
	if !hasStorage {
		return ir.Call{Function: constructor, Arguments: arguments, Returns: ir.Object}
	}
	receiver := ir.Expression(ir.Read{Local: l.locals[symbol], Of: ir.Object, Checked: true})
	if inner, handled := l.staticClassRead(ast.SkipParentheses(node.AsNewExpression().Expression)); handled {
		receiver = inner
	}
	index := len(l.result.Functions)
	self := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "class", Type: ir.Object, Function: index})
	function := ir.Function{Name: "class_ready_new", Parameters: []int{self}, Returns: ir.Object, Receiver: true, ForwardsArguments: constructor + 1, RestElement: l.result.Functions[constructor].RestElement}
	forwarded := []ir.Expression{}
	for _, parameter := range l.result.Functions[constructor].Parameters {
		local := len(l.result.Locals)
		incoming := l.result.Locals[parameter]
		incoming.Function, incoming.Global, incoming.Captured, incoming.Borrowed = index, false, false, false
		l.result.Locals = append(l.result.Locals, incoming)
		function.Parameters = append(function.Parameters, local)
		forwarded = append(forwarded, ir.Read{Local: local, Of: incoming.Type})
	}
	function.Body = []ir.Statement{ir.Return{Value: ir.Call{Function: constructor, Arguments: forwarded, Returns: ir.Object, ForwardCount: true, RestPacked: true}}}
	l.result.Functions = append(l.result.Functions, function)
	return ir.Call{Function: index, Arguments: append([]ir.Expression{receiver}, arguments...), Returns: ir.Object}
}
