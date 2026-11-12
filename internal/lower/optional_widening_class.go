package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/load"
)

type optionalClassHierarchy struct {
	fields  map[*ast.Symbol]map[string]bool
	unknown map[*ast.Symbol]bool
}

// A nominal view can still hide a subclass's field (#k4kwp9v). Absence is a
// whole-program fact, independent of the property's declared type. Library and
// ambient classes have an open subclass set and cannot establish that fact.
func (l *lowering) classOptionalAbsent(source *checker.Type, property string) bool {
	if !isClassInstance(source) || l.program == nil {
		return false
	}
	if l.optionalClasses == nil {
		l.noteOptionalClasses()
	}
	fields, known := l.optionalClasses.fields[source.Symbol()]
	return known && !l.optionalClasses.unknown[source.Symbol()] && !fields[property]
}

func optionalClassSymbol(symbol *ast.Symbol) bool {
	if symbol != nil {
		for _, declaration := range symbol.Declarations {
			if declaration.Kind == ast.KindClassDeclaration || declaration.Kind == ast.KindClassExpression {
				return true
			}
		}
	}
	return false
}

// Keep this separate from noteInheritance's layout facts: expressions and
// mixins count even when another lowering rule cannot build their bodies yet.
// An unresolved constructor may hold any compatible program class. Mark those
// classes and their ancestors unknown rather than treating a constraint as the
// actual constructor; a checker relation overflow cannot prove incompatibility.
func (l *lowering) noteOptionalClasses() {
	hierarchy := &optionalClassHierarchy{
		fields:  map[*ast.Symbol]map[string]bool{},
		unknown: map[*ast.Symbol]bool{},
	}
	l.optionalClasses = hierarchy
	types := map[*ast.Symbol]*checker.Type{}
	parents := map[*ast.Symbol][]*ast.Symbol{}
	var nodes []*ast.Node
	for _, file := range l.program.CompilerProgram().GetSourceFiles() {
		if file.IsDeclarationFile || load.IsLibrary(file) || load.IsPrelude(file) || load.IsNodeLibrary(file) {
			continue
		}
		var visit ast.Visitor
		visit = func(node *ast.Node) bool {
			if (node.Kind == ast.KindClassDeclaration || node.Kind == ast.KindClassExpression) && !ambientDeclaration(node) {
				proven := l.checker.GetTypeAtLocation(node)
				if node.Kind == ast.KindClassDeclaration && node.Name() != nil {
					proven = l.checker.GetTypeAtLocation(node.Name())
				} else {
					signatures := l.checker.GetSignaturesOfType(proven, checker.SignatureKindConstruct)
					if len(signatures) != 0 {
						proven = l.checker.GetReturnTypeOfSignature(signatures[0])
					}
				}
				if isClassInstance(proven) && optionalClassSymbol(proven.Symbol()) {
					types[proven.Symbol()] = proven
					hierarchy.fields[proven.Symbol()] = map[string]bool{}
					nodes = append(nodes, node)
				}
			}
			return node.ForEachChild(visit)
		}
		file.AsNode().ForEachChild(visit)
	}
	var unresolved []*checker.Type
	unbounded := false
	for _, node := range nodes {
		symbol := l.checker.GetSymbolOfDeclaration(node)
		if _, known := types[symbol]; !known {
			// The checker owns the class expression's symbol through its instance.
			proven := l.checker.GetTypeAtLocation(node)
			for _, signature := range l.checker.GetSignaturesOfType(proven, checker.SignatureKindConstruct) {
				symbol = l.checker.GetReturnTypeOfSignature(signature).Symbol()
				break
			}
		}
		proven := types[symbol]
		if proven == nil {
			continue
		}
		target := proven
		if proven.ObjectFlags()&checker.ObjectFlagsReference != 0 && proven.Target() != nil {
			target = proven.Target()
		}
		var baseSymbols func(*checker.Type) bool
		baseSymbols = func(base *checker.Type) bool {
			if base.Flags()&(checker.TypeFlagsUnion|checker.TypeFlagsIntersection) != 0 {
				known := false
				for _, member := range base.Types() {
					known = baseSymbols(member) || known
				}
				return known
			}
			if optionalClassSymbol(base.Symbol()) {
				parents[symbol] = append(parents[symbol], base.Symbol())
				return true
			}
			return false
		}
		for _, base := range l.checker.GetBaseTypes(target) {
			if !baseSymbols(base) {
				// A structural constructor result cannot identify even its
				// ancestors. Any program class could have an unseen descendant
				// implementing it, with a different constructor signature.
				unbounded = true
			}
		}
		for _, clause := range nodesOf(node.ClassLikeData().HeritageClauses) {
			if clause.AsHeritageClause().Token != ast.KindExtendsKeyword {
				continue
			}
			for _, heritage := range clause.AsHeritageClause().Types.Nodes {
				expression := ast.SkipParentheses(heritage.AsExpressionWithTypeArguments().Expression)
				if !optionalClassSymbol(l.symbol(expression)) {
					hierarchy.unknown[symbol] = true
					unresolved = append(unresolved, l.checker.GetTypeAtLocation(expression))
				}
			}
		}
	}
	// Propagate each subclass's complete instance property set to every ancestor.
	// GetPropertiesOfType includes parameter properties and computed names.
	for symbol, proven := range types {
		var visit func(*ast.Symbol)
		seen := map[*ast.Symbol]bool{}
		visit = func(ancestor *ast.Symbol) {
			if seen[ancestor] {
				return
			}
			seen[ancestor] = true
			if fields := hierarchy.fields[ancestor]; fields != nil {
				for _, property := range l.checker.GetPropertiesOfType(proven) {
					fields[property.Name] = true
				}
			}
			for _, base := range parents[ancestor] {
				visit(base)
			}
		}
		visit(symbol)
	}
	for _, constructor := range unresolved {
		if constructor.Flags()&checker.TypeFlagsTypeParameter != 0 {
			constructor = l.checker.GetBaseConstraintOfType(constructor)
		}
		for symbol := range types {
			if unbounded || constructor == nil || constructor.Flags()&(checker.TypeFlagsAny|checker.TypeFlagsUnknown) != 0 ||
				l.checker.IsTypeAssignableTo(l.checker.GetTypeOfSymbol(symbol), constructor) ||
				l.program.RelationOverflow(l.checker) != "" {
				hierarchy.unknown[symbol] = true
			}
		}
	}
	var mark func(*ast.Symbol, map[*ast.Symbol]bool)
	mark = func(symbol *ast.Symbol, seen map[*ast.Symbol]bool) {
		if seen[symbol] {
			return
		}
		seen[symbol] = true
		hierarchy.unknown[symbol] = true
		for _, base := range parents[symbol] {
			mark(base, seen)
		}
	}
	for symbol, unknown := range hierarchy.unknown {
		if unknown {
			mark(symbol, map[*ast.Symbol]bool{})
		}
	}
}
