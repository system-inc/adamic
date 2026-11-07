package lower

import (
	"strconv"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

func (lowering *lowering) staticBase(declaration *ast.Node) *ast.Node {
	for _, clause := range nodesOf(declaration.ClassLikeData().HeritageClauses) {
		if clause.AsHeritageClause().Token != ast.KindExtendsKeyword {
			continue
		}
		for _, node := range clause.AsHeritageClause().Types.Nodes {
			symbol := lowering.symbol(ast.SkipParentheses(node.AsExpressionWithTypeArguments().Expression))
			if symbol != nil &&
				len(symbol.Declarations) == 1 &&
				symbol.Declarations[0].Kind == ast.KindClassDeclaration {
				return symbol.Declarations[0]
			}
		}
	}
	return nil
}
func (lowering *lowering) needsStatics(declaration *ast.Node) bool {
	// A class is in the temporal dead zone even when it has no static members.
	// Cyclic graphs and namespace classes need the same ready storage for
	// construction and extends, including classes without static members.
	if lowering.cyclicModules || declaration.Parent != nil && declaration.Parent.Kind == ast.KindModuleBlock {
		return true
	}
	for _, member := range declaration.Members() {
		if member.Kind == ast.KindClassStaticBlockDeclaration ||
			ast.HasSyntacticModifier(member, ast.ModifierFlagsStatic) {
			return true
		}
	}
	base := lowering.staticBase(declaration)
	return base != nil && lowering.needsStatics(base)
}
func (lowering *lowering) staticClass(proven *checker.Type, fallback *ast.Node) *ast.Node {
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
func (lowering *lowering) staticInstance(declaration *ast.Node) (*instance, error) {
	if staticClassDeclarationError := lowering.staticClassDeclaration(declaration); staticClassDeclarationError != nil {
		return nil, staticClassDeclarationError
	}
	symbol := lowering.symbol(declaration.Name())
	if existing := lowering.statics[symbol]; existing != nil {
		return existing, nil
	}
	if lowering.statics == nil {
		lowering.statics = map[*ast.Symbol]*instance{}
	}
	if lowering.staticGlobals == nil {
		lowering.staticGlobals = map[*ast.Symbol]int{}
	}
	if _, exists := lowering.staticGlobals[symbol]; !exists {
		local, declareLocalError := lowering.declareLocal(declaration.Name())
		if declareLocalError != nil {
			return nil, declareLocalError
		}
		lowering.result.Locals[local].Global = true
		lowering.staticStorage(declaration)
	}
	var base *instance
	if parent := lowering.staticBase(declaration); parent != nil {
		var staticInstanceError error
		base, staticInstanceError = lowering.staticInstance(parent)
		if staticInstanceError != nil {
			return nil, staticInstanceError
		}
	}
	if base != nil {
		if checkMemberOverridesError := lowering.checkMemberOverrides(
			declaration,
			lowering.checker.GetTypeOfSymbol(symbol),
			lowering.checker.GetTypeOfSymbol(lowering.symbol(lowering.staticBase(declaration).Name())),
			true,
		); checkMemberOverridesError != nil {
			return nil, checkMemberOverridesError
		}
	}
	lowered := &instance{
		static:      true,
		constructor: -1,
		initializer: -1,
		superReady:  -1,
		class:       len(lowering.result.Classes) + 1,
		base:        base,
		methods:     map[string]int{},
		slots:       map[string]int{},
	}
	metadata := ir.Class{
		Source:      lowering.functionSource(declaration),
		Name:        declaration.Name().Text() + "_static",
		Constructor: -1,
		Static:      true,
	}
	if base != nil {
		parent := lowering.result.Classes[base.class-1]
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
		metadata.Fields = append(metadata.Fields, ir.Field{
			Name:    "#static-parent:" + strconv.Itoa(lowered.class),
			Value:   ir.Undefined{Of: ir.Object},
			Private: true,
		})
		metadata.StaticFlags = append(metadata.StaticFlags, 0)
		metadata.StaticParent = len(metadata.Fields)
	}
	for _, member := range declaration.Members() {
		if !ast.HasSyntacticModifier(member, ast.ModifierFlagsStatic) {
			continue
		}
		if member.Name() == nil ||
			(!ast.IsIdentifier(member.Name()) && member.Name().Kind != ast.KindPrivateIdentifier) {
			return nil, lowering.notYet(member, "a static member without an identifier name")
		}
		if member.Kind == ast.KindPropertyDeclaration {
			if ast.HasSyntacticModifier(member, ast.ModifierFlagsAmbient|ast.ModifierFlagsAbstract) {
				return nil, lowering.notYet(member, "a declare or abstract static field")
			}
			of, typeOfError := lowering.typeOf(member.Name())
			if typeOfError != nil {
				return nil, typeOfError
			}
			if slotless(of) && lowering.placeholderOrigin(member.Name()) == "" {
				return nil, lowering.notYet(member, "a static field without a native slot")
			}
			if member.AsPropertyDeclaration().Initializer == nil &&
				!lowering.uninitializedDeclaration(member) &&
				!lowering.includesUndefined(lowering.checker.GetTypeAtLocation(member.Name())) {
				return nil, lowering.notYet(
					member,
					"an uninitialized nonnullable static field; initialize it at its declaration",
				)
			}
			value := zeroValue(of)
			if of.IsReference() {
				value = ir.Undefined{Of: of}
			}
			field := ir.Field{
				Name:          lowering.memberKey(member.Name(), lowered.class),
				Value:         value,
				Private:       member.Name().Kind == ast.KindPrivateIdentifier,
				Uninitialized: lowering.uninitializedDeclaration(member),
				Unset:         lowering.placeholderOrigin(member.Name()) != "",
			}
			slot := -1
			for index, previous := range metadata.Fields {
				if previous.Name == field.Name {
					slot = index
					if previous.Value.Type() != of {
						return nil, lowering.notYet(
							member,
							"a static field override with a different native representation",
						)
					}
					break
				}
			}
			if slot < 0 {
				slot = len(metadata.Fields)
				metadata.Fields = append(metadata.Fields, field)
				metadata.StaticFlags = append(metadata.StaticFlags, 0)
				metadata.Fields = append(
					metadata.Fields,
					ir.Field{Name: "#static-present:" + field.Name, Value: ir.NumberConstant{}, Private: true},
				)
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
	lowering.result.Classes = append(lowering.result.Classes, metadata)
	lowering.statics[symbol] = lowered
	for _, member := range declaration.Members() {
		if !classFunction(member) || !ast.HasSyntacticModifier(member, ast.ModifierFlagsStatic) {
			continue
		}
		key := lowering.methodKey(member, lowered.class)
		function := len(lowering.result.Functions)
		lowered.methods[key] = function
		meta := &lowering.result.Classes[lowered.class-1]
		slot, exists := lowered.slots[key]
		if !exists {
			slot = len(meta.Methods)
			lowered.slots[key] = slot
			meta.Methods = append(meta.Methods, -1)
		}
		meta.Methods[slot] = function
		if accessorMember(member) {
			lowering.registerAccessor(lowered.class, member, function)
		}
		lowering.result.Functions = append(lowering.result.Functions, ir.Function{Name: metadata.Name + "_" + key})
	}
	outerInstance, outerType, outerNode, outerLocals := lowering.instance, lowering.classType, lowering.classNode, lowering.locals
	outerSubstitution, outerMapper := lowering.substitution, lowering.typeMapper
	lowering.instance, lowering.classType, lowering.classNode = lowered, lowering.checker.GetTypeOfSymbol(
		symbol,
	), declaration.Name()
	lowering.substitution, lowering.typeMapper = nil, nil
	lowering.locals = map[*ast.Symbol]int{}
	for name, local := range outerLocals {
		if lowering.result.Locals[local].Global {
			lowering.locals[name] = local
		}
	}
	defer func() {
		lowering.instance, lowering.classType, lowering.classNode, lowering.locals = outerInstance, outerType, outerNode, outerLocals
		lowering.substitution, lowering.typeMapper = outerSubstitution, outerMapper
	}()
	for _, member := range declaration.Members() {
		if !classFunction(member) || !ast.HasSyntacticModifier(member, ast.ModifierFlagsStatic) {
			continue
		}
		function := lowered.methods[lowering.methodKey(member, lowered.class)]
		if signatureError := lowering.signature(function, member, lowering.thisLocal(function)); signatureError != nil {
			return nil, signatureError
		}
		if validateAccessorSignatureError := lowering.validateAccessorSignature(
			member,
			function,
		); validateAccessorSignatureError != nil {
			return nil, validateAccessorSignatureError
		}
	}
	for _, member := range declaration.Members() {
		if !classFunction(member) || !ast.HasSyntacticModifier(member, ast.ModifierFlagsStatic) {
			continue
		}
		if lowerFunctionError := lowering.lowerFunction(
			lowered.methods[lowering.methodKey(member, lowered.class)],
			member,
			-1,
		); lowerFunctionError != nil {
			return nil, lowerFunctionError
		}
	}
	return lowered, nil
}

// Allocation installs methods and accessors before the ordered field and block initializers run.
// The named map uses this static side's implementations, including inherited ones, while
// the class's virtual slots still serve nominal calls. Both keep the actual receiver as this.
func (lowering *lowering) staticDeclaration(declaration *ast.Node) ([]ir.Statement, error) {
	if staticClassDeclarationError := lowering.staticClassDeclaration(declaration); staticClassDeclarationError != nil {
		return nil, staticClassDeclarationError
	}
	if !lowering.needsStatics(declaration) && lowering.statics[lowering.symbol(declaration.Name())] == nil {
		return nil, nil
	}
	lowered, err := lowering.staticInstance(declaration)
	if err != nil {
		return nil, err
	}
	fields := append([]ir.Field{}, lowering.result.Classes[lowered.class-1].Fields...)
	if lowered.base != nil {
		parent := lowering.staticBase(declaration)
		slot := lowering.result.Classes[lowered.class-1].StaticParent - 1
		// Extends reads the public binding, whose name and ready point belong to
		// the source class, rather than the internal storage initialized earlier.
		checked := true
		for _, clause := range nodesOf(declaration.ClassLikeData().HeritageClauses) {
			if clause.AsHeritageClause().Token != ast.KindExtendsKeyword {
				continue
			}
			for _, element := range clause.AsHeritageClause().Types.Nodes {
				expression := ast.SkipParentheses(element.AsExpressionWithTypeArguments().Expression)
				if lowering.symbol(expression) == lowering.symbol(parent.Name()) {
					checked = !lowering.provenModuleReads[expression]
				}
			}
		}
		fields[slot].Value = ir.Read{
			Local:   lowering.locals[lowering.symbol(parent.Name())],
			Of:      ir.Object,
			Checked: checked,
		}
	}
	object := ir.Read{Local: lowering.staticGlobals[lowering.symbol(declaration.Name())], Of: ir.Object}
	statements := []ir.Statement{ir.Declare{Local: object.Local, Value: ir.ObjectLiteral{
		ProgramRegion: lowering.programClass(lowered.class),
		Fields:        fields,
		Class:         lowered.class,
		Methods:       lowered.methodList(),
	}}}
	outerInstance, outerType, outerNode := lowering.instance, lowering.classType, lowering.classNode
	lowering.instance, lowering.classType, lowering.classNode = lowered, lowering.checker.GetTypeOfSymbol(
		lowering.symbol(declaration.Name()),
	), declaration.Name()
	defer func() {
		lowering.instance, lowering.classType, lowering.classNode = outerInstance, outerType, outerNode
	}()
	available := map[string]bool{}
	if lowered.base != nil {
		for _, field := range lowering.result.Classes[lowered.base.class-1].Fields {
			available[field.Name] = true
		}
	}
	for _, member := range declaration.Members() {
		if member.Kind != ast.KindClassStaticBlockDeclaration &&
			!(member.Kind == ast.KindPropertyDeclaration && ast.HasSyntacticModifier(member, ast.ModifierFlagsStatic)) {
			continue
		}
		{
			var body *ast.Node
			if member.Kind == ast.KindClassStaticBlockDeclaration {
				body = member.AsClassStaticBlockDeclaration().Body
			} else {
				body = member.AsPropertyDeclaration().Initializer
			}
			if staticInitializationReadsError := lowering.staticInitializationReads(
				body,
				declaration,
				available,
				map[*ast.Node]uint8{},
				true,
			); staticInitializationReadsError != nil {
				return nil, staticInitializationReadsError
			}
		}
		index := len(lowering.result.Functions)
		lowering.result.Functions = append(
			lowering.result.Functions,
			ir.Function{Name: declaration.Name().Text() + "_static_initialize"},
		)
		self := lowering.thisLocal(index)
		lowering.result.Functions[index].Parameters = []int{self}
		outerThis, outerFunction, outerIndex, outerLocals := lowering.this, lowering.function, lowering.functionIndex, lowering.locals
		function := lowering.result.Functions[index]
		lowering.this, lowering.function, lowering.functionIndex = self, &function, index
		lowering.locals = map[*ast.Symbol]int{}
		for symbol, local := range outerLocals {
			if lowering.result.Locals[local].Global {
				lowering.locals[symbol] = local
			}
		}
		if member.Kind == ast.KindClassStaticBlockDeclaration {
			function.Body, err = lowering.statements(
				member.AsClassStaticBlockDeclaration().Body.AsBlock().Statements.Nodes,
			)
		} else {
			of, _ := lowering.typeOf(member.Name())
			value := zeroValue(of)
			initializer := member.AsPropertyDeclaration().Initializer

			if lowering.uninitializedDeclaration(member) {
				if of.IsReference() {
					value = ir.Undefined{Of: of}
				}
				if lowering.placeholderOrigin(member.Name()) != "" {
					value = lowering.placeholderInitialValue(initializer, of)
				}
			} else if initializer != nil {
				value, err = lowering.expression(initializer)
				value = fit(value, of)
			} else if of.IsReference() {
				value = ir.Undefined{Of: of}
			}
			if err == nil {
				function.Body = []ir.Statement{ir.SetProperty{
					Object:        ir.Read{Local: self, Of: ir.Object},
					Name:          lowering.memberKey(member.Name(), lowered.class),
					Value:         value,
					Define:        true,
					Uninitialized: lowering.uninitializedDeclaration(member),
					Unset:         lowering.placeholderOrigin(member.Name()) != "",
					Site:          lowering.staticWriteSite(declaration.Name()),
				}}
			}
			available[member.Name().Text()] = true
		}
		lowering.this, lowering.function, lowering.functionIndex, lowering.locals = outerThis, outerFunction, outerIndex, outerLocals
		if err != nil {
			return nil, err
		}
		lowering.result.Functions[index] = function
		statements = append(
			statements,
			ir.Evaluate{Value: ir.Call{Function: index, Arguments: []ir.Expression{object}}},
		)
	}
	statements = append(
		statements,
		ir.Declare{Local: lowering.locals[lowering.symbol(declaration.Name())], Value: object},
	)
	return statements, nil
}

func (lowering *lowering) isStaticType(proven *checker.Type) bool {
	return proven != nil &&
		len(lowering.checker.GetSignaturesOfType(proven, checker.SignatureKindConstruct)) > 0 &&
		lowering.staticClass(proven, nil) != nil
}
func (lowering *lowering) staticWriteSite(node *ast.Node) int {
	lowering.writeSites = append(lowering.writeSites, writeSite{holder: lowering.classType, node: node})
	return len(lowering.writeSites)
}

// JavaScript leaves future fields undefined. A nonnullable native slot cannot pretend that
// value is its declared type, including when an initializer calls a helper that reads it.
func (lowering *lowering) staticInitializationReads(
	node, declaration *ast.Node,
	available map[string]bool,
	seen map[*ast.Node]uint8,
	receiver bool,
) error {
	pending := false
	for _, member := range declaration.Members() {
		if member.Kind == ast.KindPropertyDeclaration &&
			ast.HasSyntacticModifier(member, ast.ModifierFlagsStatic) &&
			!available[member.Name().Text()] {
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
				current = lowering.symbol(object) == lowering.symbol(declaration.Name())
			}
			if current {
				symbol := lowering.checker.GetSymbolAtLocation(child.Name())
				if symbol != nil && len(symbol.Declarations) > 0 {
					member := symbol.Declarations[0]
					if member.Kind == ast.KindPropertyDeclaration && !available[member.Name().Text()] {
						failure = &Refused{
							Where: lowering.program.Where(child),
							What:  "a static field read before its initializer has run",
							Fix:   "declare the field earlier, or move the read after that field's initialization",
						}
						return true
					}
					if accessorMember(member) {
						for _, candidate := range symbol.Declarations {
							if accessorMember(candidate) {
								failure = lowering.staticInitializationReads(
									candidate.Body(),
									declaration,
									available,
									seen,
									true,
								)
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
		if ast.IsIdentifier(child) &&
			lowering.symbol(child) == lowering.symbol(declaration.Name()) &&
			!ast.IsDeclarationName(child) {
			if !insideClass(child, declaration) {
				failure = &Refused{
					Where: lowering.program.Where(child),
					What:  "an outside helper reading a class during its declaration (the temporal dead zone)",
					Fix:   "run the helper after the class declaration, or read this inside a static method",
				}
				return true
			}
			parent := child.Parent
			if pending && (parent == nil ||
				parent.Kind != ast.KindPropertyAccessExpression ||
				parent.AsPropertyAccessExpression().Expression != child) {
				failure = &Refused{
					Where: lowering.program.Where(child),
					What:  "a constructor object escaping before static fields are initialized",
					Fix:   "publish the class value after the last static field initializer",
				}
				return true
			}
		}
		if child.Kind == ast.KindThisKeyword && receiver {
			parent := child.Parent
			if pending && (parent == nil ||
				parent.Kind != ast.KindPropertyAccessExpression ||
				parent.AsPropertyAccessExpression().Expression != child) {
				failure = &Refused{
					Where: lowering.program.Where(child),
					What:  "a constructor object escaping before static fields are initialized",
					Fix:   "publish this after the last static field initializer",
				}
				return true
			}
		}
		if child.Kind == ast.KindCallExpression {
			callee := ast.SkipParentheses(child.AsCallExpression().Expression)
			symbol := lowering.checker.GetSymbolAtLocation(callee)
			if ast.IsIdentifier(callee) {
				symbol = lowering.symbol(callee)
			}
			known := false
			if symbol != nil && len(symbol.Declarations) > 0 {
				target := symbol.Declarations[0]
				known = load.IsLibrary(ast.GetSourceFileOfNode(target)) ||
					load.IsPrelude(ast.GetSourceFileOfNode(target))
				if target.Kind == ast.KindFunctionDeclaration {
					known = true
					failure = lowering.staticInitializationReads(target.Body(), declaration, available, seen, false)
				}
				if target.Kind == ast.KindMethodDeclaration &&
					ast.HasSyntacticModifier(target, ast.ModifierFlagsStatic) {
					known = true
					current := false
					if callee.Kind == ast.KindPropertyAccessExpression {
						object := ast.SkipParentheses(callee.AsPropertyAccessExpression().Expression)
						current = (receiver &&
							(object.Kind == ast.KindThisKeyword || object.Kind == ast.KindSuperKeyword)) ||
							(ast.IsIdentifier(object) && lowering.symbol(object) == lowering.symbol(declaration.Name()))
					}
					failure = lowering.staticInitializationReads(target.Body(), declaration, available, seen, current)
				}
			}
			if known && symbol != nil && len(symbol.Declarations) > 0 && (load.IsLibrary(
				ast.GetSourceFileOfNode(symbol.Declarations[0]),
			) || load.IsPrelude(ast.GetSourceFileOfNode(symbol.Declarations[0]))) {
				for _, argument := range nodesOf(child.AsCallExpression().Arguments) {
					if len(
						lowering.checker.GetSignaturesOfType(
							lowering.checker.GetTypeAtLocation(argument),
							checker.SignatureKindCall,
						),
					) == 0 {
						continue
					}
					if ast.IsFunctionLike(ast.SkipParentheses(argument)) {
						continue
					}
					target := lowering.symbol(ast.SkipParentheses(argument))
					if target != nil &&
						len(target.Declarations) == 1 &&
						target.Declarations[0].Kind == ast.KindFunctionDeclaration {
						failure = lowering.staticInitializationReads(
							target.Declarations[0].Body(),
							declaration,
							available,
							seen,
							false,
						)
					} else {
						failure = lowering.notYet(
							argument,
							"an indirect callback before static fields are initialized; move it after the field initializers",
						)
					}
					if failure != nil {
						break
					}
				}
			}
			if !known && failure == nil {
				failure = lowering.notYet(
					child,
					"an indirect call before static fields are initialized; move it after the field initializers",
				)
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
func (lowering *lowering) staticProperty(node *ast.Node) error {
	receiver := node.AsPropertyAccessExpression().Expression
	if !lowering.isStaticType(lowering.checker.GetTypeAtLocation(receiver)) &&
		!(lowering.instance != nil &&
			lowering.instance.static &&
			ast.SkipParentheses(receiver).Kind == ast.KindThisKeyword) {
		return nil
	}
	symbol := lowering.checker.GetSymbolAtLocation(node.Name())
	if symbol != nil {
		for _, member := range symbol.Declarations {
			if member.Parent != nil &&
				member.Parent.Kind == ast.KindClassDeclaration &&
				ast.HasSyntacticModifier(member, ast.ModifierFlagsStatic) {
				return nil
			}
		}
	}
	return lowering.notYet(node, "an implicit constructor property such as prototype, name or length")
}

func (lowering *lowering) staticClassDeclaration(declaration *ast.Node) error {
	if ast.HasSyntacticModifier(declaration, ast.ModifierFlagsAmbient) {
		return lowering.notYet(declaration, "a declare class without a runtime definition")
	}
	for _, clause := range nodesOf(declaration.ClassLikeData().HeritageClauses) {
		if clause.AsHeritageClause().Token == ast.KindExtendsKeyword && lowering.staticBase(declaration) == nil {
			return lowering.notYet(clause, "a computed class base; name the base class directly")
		}
	}
	// Instance-private names belong to instance storage, even in the declaring class.
	// Diagnose this shape before registering static bodies: the wrong method key can
	// otherwise reach call-target analysis as an internal compiler error.
	for _, member := range declaration.Members() {
		if member.Kind != ast.KindClassStaticBlockDeclaration &&
			!ast.HasSyntacticModifier(member, ast.ModifierFlagsStatic) {
			continue
		}
		var failure error
		var visit ast.Visitor
		visit = func(node *ast.Node) bool {
			if node.Kind == ast.KindPropertyAccessExpression && node.Name().Kind == ast.KindPrivateIdentifier {
				symbol := lowering.checker.GetSymbolAtLocation(node.Name())
				if symbol != nil && len(symbol.Declarations) > 0 {
					target := symbol.Declarations[0]
					if target.Parent == declaration && !ast.HasSyntacticModifier(target, ast.ModifierFlagsStatic) {
						failure = &Refused{
							Where: lowering.program.Where(node),
							What: "instance private storage reached from a static member: " +
								declaration.Name().Text() +
								"." +
								node.Name().Text() +
								" (adamic/private-instance-from-static)",
							Fix: "move the private access into an instance method; call that instance method from the static member",
						}
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

func (lowering *lowering) staticStorage(declaration *ast.Node) {
	symbol := lowering.symbol(declaration.Name())
	if _, exists := lowering.staticGlobals[symbol]; exists {
		return
	}
	local := len(lowering.result.Locals)
	lowering.result.Locals = append(
		lowering.result.Locals,
		ir.Local{Name: declaration.Name().Text() + "_class", Type: ir.Object, Function: -1, Global: true},
	)
	lowering.staticGlobals[symbol] = local
	lowering.noteLocal(local, lowering.checker.GetTypeOfSymbol(symbol), declaration.Name())
}
func insideClass(node, declaration *ast.Node) bool {
	for parent := node.Parent; parent != nil; parent = parent.Parent {
		if parent.Kind == ast.KindClassDeclaration {
			return parent == declaration
		}
	}
	return false
}
func (lowering *lowering) staticClassRead(node *ast.Node) (ir.Expression, bool) {
	symbol := lowering.symbol(node)
	declaration := lowering.classes[symbol]
	local, exists := lowering.staticGlobals[symbol]
	if !exists || !insideClass(node, declaration) {
		return nil, false
	}
	return ir.Read{Local: local, Of: ir.Object, Checked: !lowering.provenModuleReads[node]}, true
}

// Looking up the constructor precedes evaluating new's arguments. Instance allocators
// remain independent functions, so make that declaration-readiness check explicit here.
func (lowering *lowering) staticConstruct(
	node, declaration *ast.Node,
	constructor int,
	arguments []ir.Expression,
) ir.Expression {
	symbol := lowering.symbol(declaration.Name())
	_, hasStorage := lowering.staticGlobals[symbol]
	if !hasStorage {
		return ir.Call{Function: constructor, Arguments: arguments, Returns: ir.Object}
	}
	expression := ast.SkipParentheses(node.AsNewExpression().Expression)
	receiver := ir.Expression(
		ir.Read{Local: lowering.locals[symbol], Of: ir.Object, Checked: !lowering.provenModuleReads[expression]},
	)
	if inner, handled := lowering.staticClassRead(ast.SkipParentheses(node.AsNewExpression().Expression)); handled {
		receiver = inner
	}
	index := len(lowering.result.Functions)
	self := len(lowering.result.Locals)
	lowering.result.Locals = append(lowering.result.Locals, ir.Local{Name: "class", Type: ir.Object, Function: index})
	function := ir.Function{
		Name:              "class_ready_new",
		Parameters:        []int{self},
		Returns:           ir.Object,
		Receiver:          true,
		ForwardsArguments: constructor + 1,
		RestElement:       lowering.result.Functions[constructor].RestElement,
	}
	forwarded := []ir.Expression{}
	for _, parameter := range lowering.result.Functions[constructor].Parameters {
		local := len(lowering.result.Locals)
		incoming := lowering.result.Locals[parameter]
		incoming.Function, incoming.Global, incoming.Captured, incoming.Borrowed = index, false, false, false
		lowering.result.Locals = append(lowering.result.Locals, incoming)
		function.Parameters = append(function.Parameters, local)
		forwarded = append(forwarded, ir.Read{Local: local, Of: incoming.Type})
	}
	function.Body = []ir.Statement{ir.Return{Value: ir.Call{
		Function:     constructor,
		Arguments:    forwarded,
		Returns:      ir.Object,
		ForwardCount: true,
		RestPacked:   true,
	}}}
	lowering.result.Functions = append(lowering.result.Functions, function)
	return ir.Call{Function: index, Arguments: append([]ir.Expression{receiver}, arguments...), Returns: ir.Object}
}
