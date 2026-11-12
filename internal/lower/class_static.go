package lower

import (
	"strconv"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

func (l *lowering) staticBase(declaration *ast.Node) *ast.Node {
	for _, clause := range nodesOf(declaration.ClassLikeData().HeritageClauses) {
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
	// A class is in the temporal dead zone even when it has no static members.
	// Cyclic graphs and namespace classes need the same ready storage for
	// construction and extends, including classes without static members.
	if l.cyclicModules || declaration.Parent != nil && declaration.Parent.Kind == ast.KindModuleBlock {
		return true
	}
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
	metadata := ir.Class{Source: l.functionSource(declaration), Name: declaration.Name().Text() + "_static", Constructor: -1, Static: true}
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
			if slotless(of) && l.placeholderOrigin(member.Name()) == "" {
				return nil, l.notYet(member, "a static field without a native slot")
			}
			if member.AsPropertyDeclaration().Initializer == nil && !l.uninitializedDeclaration(member) && !l.includesUndefined(l.checker.GetTypeAtLocation(member.Name())) {
				return nil, l.notYet(member, "an uninitialized nonnullable static field; initialize it at its declaration")
			}
			value := zeroValue(of)
			if of.IsReference() {
				value = ir.Undefined{Of: of}
			}
			field := ir.Field{Name: l.memberKey(member.Name(), lowered.class), Value: value, Private: member.Name().Kind == ast.KindPrivateIdentifier, Uninitialized: l.uninitializedDeclaration(member), Unset: l.placeholderOrigin(member.Name()) != ""}
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
		key := l.methodKey(member, lowered.class)
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
		function := lowered.methods[l.methodKey(member, lowered.class)]
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
		if err := l.lowerFunction(lowered.methods[l.methodKey(member, lowered.class)], member, -1); err != nil {
			return nil, err
		}
	}
	return lowered, nil
}

// Allocation installs methods and accessors before the ordered field and block initializers run.
// The named map uses this static side's implementations, including inherited ones, while
// the class's virtual slots still serve nominal calls. Both keep the actual receiver as this.
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
		// Extends reads the public binding, whose name and ready point belong to
		// the source class, rather than the internal storage initialized earlier.
		checked := true
		for _, clause := range nodesOf(declaration.ClassLikeData().HeritageClauses) {
			if clause.AsHeritageClause().Token != ast.KindExtendsKeyword {
				continue
			}
			for _, element := range clause.AsHeritageClause().Types.Nodes {
				expression := ast.SkipParentheses(element.AsExpressionWithTypeArguments().Expression)
				if l.symbol(expression) == l.symbol(parent.Name()) {
					checked = !l.provenModuleReads[expression]
				}
			}
		}
		fields[slot].Value = ir.Read{Local: l.locals[l.symbol(parent.Name())], Of: ir.Object, Checked: checked}
	}
	object := ir.Read{Local: l.staticGlobals[l.symbol(declaration.Name())], Of: ir.Object}
	statements := []ir.Statement{ir.Declare{Local: object.Local, Value: ir.ObjectLiteral{ProgramRegion: l.programClass(lowered.class), Fields: fields, Class: lowered.class, Methods: lowered.methodList()}}}
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
			if err := l.staticInitializationReads(body, declaration, available, map[*ast.Node]uint8{}, true, false); err != nil {
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
			if err == nil {
				l.noteClassInitializer(declaration, member.AsClassStaticBlockDeclaration().Body)
			}
		} else {
			of, _ := l.typeOf(member.Name())
			value := zeroValue(of)
			initializer := member.AsPropertyDeclaration().Initializer

			if l.uninitializedDeclaration(member) {
				if of.IsReference() {
					value = ir.Undefined{Of: of}
				}
				if l.placeholderOrigin(member.Name()) != "" {
					value = l.placeholderInitialValue(initializer, of)
				}
			} else if initializer != nil {
				value, err = l.expression(initializer)
				if err == nil {
					l.noteClassInitializer(declaration, initializer)
				}
				value = fit(value, of)
			} else if of.IsReference() {
				value = ir.Undefined{Of: of}
			}
			if err == nil {
				function.Body = []ir.Statement{ir.SetProperty{Object: ir.Read{Local: self, Of: ir.Object}, Name: l.memberKey(member.Name(), lowered.class), Value: value, Nullish: l.fieldNullish(initializer, value), Define: true, Uninitialized: l.uninitializedDeclaration(member), Unset: l.placeholderOrigin(member.Name()) != "", Site: l.staticWriteSite(declaration.Name())}}
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
// judged says node is the body or a parameter of a function every call edge reaching it in this
// walk has judged (staticParameterCalls), so a call of that function's own parameter is known.
func (l *lowering) staticInitializationReads(node, declaration *ast.Node, available map[string]bool, seen map[*ast.Node]uint8, receiver, judged bool) error {
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
	if judged {
		state <<= 2
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
						failure = &Refused{Where: l.program.Where(child), What: "a static field read before its initializer has run: " + declaration.Name().Text() + "." + member.Name().Text(), Fix: "declare the field earlier, or move the read after that field's initialization"}
						return true
					}
					if accessorMember(member) {
						for _, candidate := range symbol.Declarations {
							if accessorMember(candidate) {
								failure = l.staticInitializationReads(candidate.Body(), declaration, available, seen, true, false)
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
			// An instance accessor of this class runs whenever an instance it constructed
			// during static initialization is read or written, so its bodies are reached too.
			if !current {
				if symbol := l.checker.GetSymbolAtLocation(child.Name()); symbol != nil {
					for _, candidate := range symbol.Declarations {
						if accessorMember(candidate) && candidate.Parent == declaration && !ast.HasSyntacticModifier(candidate, ast.ModifierFlagsStatic) {
							failure = l.staticInitializationReads(candidate.Body(), declaration, available, seen, false, false)
							if failure != nil {
								return true
							}
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
				if construction := selfConstruction(child); construction != nil && !extendsClass(declaration) {
					// The class constructing itself is judged by what the construction reads, as
					// the future-read ruling judges any initializer (#1451npc): the constructor and
					// everything it reaches, against the same available fields. The arguments are
					// the new expression's own children and are walked after this.
					failure = l.staticConstructionReads(construction, declaration, available, seen)
				} else {
					failure = &Refused{Where: l.program.Where(child), What: "a constructor object escaping before static fields are initialized", Fix: "publish the class value after the last static field initializer"}
				}
				if failure != nil {
					return true
				}
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
					failure = l.staticParameterCalls(target, child, declaration, available, seen)
					if failure == nil {
						failure = l.staticInitializationReads(target.Body(), declaration, available, seen, false, true)
					}
				}
				if target.Kind == ast.KindMethodDeclaration && ast.HasSyntacticModifier(target, ast.ModifierFlagsStatic) {
					known = true
					current := false
					if callee.Kind == ast.KindPropertyAccessExpression {
						object := ast.SkipParentheses(callee.AsPropertyAccessExpression().Expression)
						current = (receiver && (object.Kind == ast.KindThisKeyword || object.Kind == ast.KindSuperKeyword)) || (ast.IsIdentifier(object) && l.symbol(object) == l.symbol(declaration.Name()))
					}
					failure = l.staticParameterCalls(target, child, declaration, available, seen)
					if failure == nil {
						failure = l.staticInitializationReads(target.Body(), declaration, available, seen, current, true)
					}
				}
				// A parameter of the function this walk is in was judged at every edge that reached it.
				if judged && target.Kind == ast.KindParameter && target.Parent == node.Parent && !l.assignedSymbol(symbol) {
					known = true
				}
			}
			if !known && l.staticLibraryCallsNothing(callee) {
				known = true
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
						failure = l.staticInitializationReads(target.Declarations[0].Body(), declaration, available, seen, false, false)
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

// selfConstruction is the new expression name, a read of the class's own name, is the constructor
// of, `new Version(...)`, or nil where the class object is handed somewhere.
func selfConstruction(name *ast.Node) *ast.Node {
	parent := name.Parent
	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		parent = parent.Parent
	}
	if parent != nil && parent.Kind == ast.KindNewExpression && ast.SkipParentheses(parent.AsNewExpression().Expression) == name {
		return parent
	}
	return nil
}

// extendsClass reports whether declaration has a base, whose constructor's reach the static
// initialization walk doesn't follow.
func extendsClass(declaration *ast.Node) bool {
	for _, clause := range nodesOf(declaration.ClassLikeData().HeritageClauses) {
		if clause.AsHeritageClause().Token == ast.KindExtendsKeyword {
			return true
		}
	}
	return false
}

// Constructing a class without a base runs, in order, the implementation's parameter defaults,
// the instance field initializers and the constructor body, each with this the new instance
// rather than the class. Overload signatures have no body and run nothing. construction, the new
// expression, is the call edge the constructor's own parameter calls are judged by.
func (l *lowering) staticConstructionReads(construction, declaration *ast.Node, available map[string]bool, seen map[*ast.Node]uint8) error {
	var implementation *ast.Node
	for _, member := range declaration.Members() {
		if member.Kind == ast.KindConstructor && member.Body() != nil {
			implementation = member
		}
	}
	if implementation != nil {
		if err := l.staticParameterCalls(implementation, construction, declaration, available, seen); err != nil {
			return err
		}
		for _, parameter := range implementation.Parameters() {
			if err := l.staticInitializationReads(parameter, declaration, available, seen, false, true); err != nil {
				return err
			}
		}
	}
	for _, member := range declaration.Members() {
		if member.Kind == ast.KindPropertyDeclaration && !ast.HasSyntacticModifier(member, ast.ModifierFlagsStatic) {
			if err := l.staticInitializationReads(member.AsPropertyDeclaration().Initializer, declaration, available, seen, false, false); err != nil {
				return err
			}
		}
	}
	if implementation != nil {
		return l.staticInitializationReads(implementation.Body(), declaration, available, seen, false, true)
	}
	return nil
}

// A parameter a function calls runs what the call edge passed for it, or its default, so each edge
// reaching the function in a static initialization walk judges its own (#1451npc, Adamic's ruling):
// with neither, the parameter is undefined and the call runs no program code (debug.ts's assert,
// reached with no verboseDebugInfo); a value nothing can call runs none either; a function written
// in place, or a function declaration, is walked; anything else can't be followed and keeps the stop.
func (l *lowering) staticParameterCalls(function, edge, declaration *ast.Node, available map[string]bool, seen map[*ast.Node]uint8) error {
	if function.Body() == nil {
		return nil
	}
	var arguments []*ast.Node
	switch edge.Kind {
	case ast.KindCallExpression:
		arguments = nodesOf(edge.AsCallExpression().Arguments)
	case ast.KindNewExpression:
		arguments = nodesOf(edge.AsNewExpression().Arguments)
	}
	for index, parameter := range function.Parameters() {
		symbol := l.symbol(parameter.Name())
		if symbol == nil || parameter.AsParameterDeclaration().DotDotDotToken != nil || !l.callsParameter(function, symbol) {
			continue
		}
		var argument *ast.Node
		for position, candidate := range arguments {
			if candidate.Kind == ast.KindSpreadElement && position <= index {
				return l.notYet(candidate, "an indirect callback before static fields are initialized; move it after the field initializers")
			}
			if position == index {
				argument = candidate
			}
		}
		if argument == nil {
			argument = parameter.AsParameterDeclaration().Initializer
		}
		if argument == nil || !l.mayBeCallable(l.checker.GetTypeAtLocation(argument)) {
			continue
		}
		inner := ast.SkipParentheses(argument)
		if ast.IsFunctionLike(inner) {
			if err := l.staticInitializationReads(inner.Body(), declaration, available, seen, false, false); err != nil {
				return err
			}
			continue
		}
		target := l.symbol(inner)
		if target != nil && len(target.Declarations) == 1 && target.Declarations[0].Kind == ast.KindFunctionDeclaration {
			if err := l.staticInitializationReads(target.Declarations[0].Body(), declaration, available, seen, false, false); err != nil {
				return err
			}
			continue
		}
		return l.notYet(argument, "an indirect callback before static fields are initialized; move it after the field initializers")
	}
	return nil
}

// callsParameter reports whether function, its parameters or its body, calls parameter by name.
func (l *lowering) callsParameter(function *ast.Node, parameter *ast.Symbol) bool {
	found := false
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if found || ast.IsPartOfTypeNode(node) {
			return found
		}
		if node.Kind == ast.KindCallExpression {
			callee := ast.SkipParentheses(node.AsCallExpression().Expression)
			if ast.IsIdentifier(callee) && l.symbol(callee) == parameter {
				found = true
				return true
			}
		}
		return node.ForEachChild(visit)
	}
	function.ForEachChild(visit)
	return found
}

// mayBeCallable reports whether any member of a value's type can be called: string | (() => string)
// can, though the union has no call signature of its own.
func (l *lowering) mayBeCallable(proven *checker.Type) bool {
	if proven.Flags()&(checker.TypeFlagsAny|checker.TypeFlagsUnknown) != 0 {
		return true
	}
	for _, member := range castMembers(proven) {
		if len(l.checker.GetSignaturesOfType(member, checker.SignatureKindCall)) > 0 {
			return true
		}
	}
	return false
}

// libraryMembersCallingNothing are the library members whose call runs no program code: Node's
// Error.captureStackTrace sets its target's stack and calls only an installed
// Error.prepareStackTrace, and every road to installing one, or to replacing captureStackTrace
// itself, stays stopped (TestPrepareStackTraceStaysStopped). Its constructorOpt is a frame marker V8 never calls. A member
// that calls its arguments, as Array.from does, isn't one.
var libraryMembersCallingNothing = map[string]bool{nodeCaptureStackTraceMember: true}

// staticLibraryCallsNothing reports whether callee is a member libraryMembersCallingNothing names,
// read off a library global. Both resolve by symbol, never by name: the receiver, past parentheses
// and casts such as tsc's (Error as any), is a global every declaration of which is a library's, and
// the member is that global's own, found on its declared type.
func (l *lowering) staticLibraryCallsNothing(callee *ast.Node) bool {
	callee = ast.SkipOuterExpressions(callee, ast.OEKParentheses|ast.OEKAssertions)
	if callee.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	receiver := ast.SkipOuterExpressions(callee.AsPropertyAccessExpression().Expression, ast.OEKParentheses|ast.OEKAssertions)
	if !ast.IsIdentifier(receiver) {
		return false
	}
	global := l.checker.GetSymbolAtLocation(receiver)
	if !librarySymbol(global) {
		return false
	}
	member := l.checker.GetPropertyOfType(l.checker.GetTypeOfSymbol(global), callee.Name().Text())
	return librarySymbol(member) && libraryMembersCallingNothing[nodeLibraryName(member)]
}

// librarySymbol reports whether symbol is declared, every declaration of it, by TypeScript's bundled
// library or Node's, so no program declaration merges into it.
func librarySymbol(symbol *ast.Symbol) bool {
	if symbol == nil || len(symbol.Declarations) == 0 {
		return false
	}
	for _, declaration := range symbol.Declarations {
		file := ast.GetSourceFileOfNode(declaration)
		if !load.IsLibrary(file) && !load.IsNodeLibrary(file) {
			return false
		}
	}
	return true
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
	for _, clause := range nodesOf(declaration.ClassLikeData().HeritageClauses) {
		if clause.AsHeritageClause().Token == ast.KindExtendsKeyword && l.staticBase(declaration) == nil {
			return l.notYet(clause, "a computed class base; name the base class directly")
		}
	}
	// Instance-private names belong to instance storage, even in the declaring class.
	// Diagnose this shape before registering static bodies: the wrong method key can
	// otherwise reach call-target analysis as an internal compiler error.
	for _, member := range declaration.Members() {
		if member.Kind != ast.KindClassStaticBlockDeclaration && !ast.HasSyntacticModifier(member, ast.ModifierFlagsStatic) {
			continue
		}
		var failure error
		var visit ast.Visitor
		visit = func(node *ast.Node) bool {
			if node.Kind == ast.KindPropertyAccessExpression && node.Name().Kind == ast.KindPrivateIdentifier {
				symbol := l.checker.GetSymbolAtLocation(node.Name())
				if symbol != nil && len(symbol.Declarations) > 0 {
					target := symbol.Declarations[0]
					if target.Parent == declaration && !ast.HasSyntacticModifier(target, ast.ModifierFlagsStatic) {
						failure = &Refused{Where: l.program.Where(node), What: "instance private storage reached from a static member: " + declaration.Name().Text() + "." + node.Name().Text() + " (adamic/private-instance-from-static)", Fix: "move the private access into an instance method; call that instance method from the static member"}
						return true
					}
				}
			}
			return node.ForEachChild(visit)
		}
		member.ForEachChild(visit)
		if failure != nil {
			return failure
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
	return ir.Read{Local: local, Of: ir.Object, Checked: !l.provenModuleReads[node]}, true
}

// Looking up the constructor precedes evaluating new's arguments. Instance allocators
// remain independent functions, so make that declaration-readiness check explicit here.
func (l *lowering) staticConstruct(node, declaration *ast.Node, constructor int, arguments []ir.Expression) ir.Expression {
	symbol := l.symbol(declaration.Name())
	_, hasStorage := l.staticGlobals[symbol]
	if !hasStorage {
		return ir.Call{Function: constructor, Arguments: arguments, Returns: ir.Object}
	}
	expression := ast.SkipParentheses(node.AsNewExpression().Expression)
	receiver := ir.Expression(ir.Read{Local: l.locals[symbol], Of: ir.Object, Checked: !l.provenModuleReads[expression]})
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
