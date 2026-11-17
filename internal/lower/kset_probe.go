package lower

// A probe for the proposed kind-domain rule, not for landing: with ADAMIC_PROBE_KSET set to a file, the
// first lowering in the process appends one JSON line per fact and changes no decision. A "kset" row is a
// declared interface, type alias or class assignable to one of the branded supertypes ksetProbeTargets
// names, with its kind's type; a "build" row is a place the program builds a value whose kind isn't one
// literal; a "summary" row closes each supertype's K(S). With the probe on, its checker queries run before
// the lowering's own, so type ids, and so that run's refusals, may differ from a run without it.

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/load"
)

var (
	ksetProbePath  = os.Getenv("ADAMIC_PROBE_KSET")
	ksetProbeOnce  sync.Once
	ksetProbeMutex sync.Mutex
)

// ksetProbeTargets are the branded supertypes the enum-tag walk refuses narrowing over on tsc, by declared name.
var ksetProbeTargets = []string{"Expression", "Node", "JSDocPropertyLikeTag", "LiteralExpression", "Statement"}

// ksetProbe holds one run's rows, so a panic partway still writes what came before it.
type ksetProbe struct {
	l    *lowering
	rows []map[string]any
}

// ksetDeclared is one interface, type alias or class a source file declares, by its symbol, so two aliases
// of one type are two rows; typeId lets a reader merge them.
type ksetDeclared struct {
	node     *ast.Node
	declared *checker.Type
	origin   string
	generic  bool
}

// ksetKindShape is a kind type judged: Literal, LiteralUnion, OpenEnum, TypeParameter, Other or Missing.
type ksetKindShape struct {
	text            string
	class           string
	members         []string
	constraint      string
	constraintClass string
	optional        bool
}

// probeKindSets runs the probe once per process, on the first lowering that reaches it.
func (l *lowering) probeKindSets() {
	if ksetProbePath == "" {
		return
	}
	ksetProbeOnce.Do(func() {
		probe := &ksetProbe{l: l}
		defer func() {
			if recovered := recover(); recovered != nil {
				probe.rows = append(probe.rows, map[string]any{"row": "panic", "error": fmt.Sprint(recovered)})
			}
			probe.write()
		}()
		declared := probe.declaredTypes()
		probe.kindSets(declared)
		probe.builds()
	})
}

// ksetOrigin names where a file comes from: the program's own, or a library the probe skips as declarations.
func ksetOrigin(file *ast.SourceFile) string {
	switch {
	case load.IsPrelude(file):
		return "Prelude"
	case load.IsLibrary(file):
		return "Library"
	case load.IsNodeLibrary(file):
		return "NodeLibrary"
	case file.IsDeclarationFile:
		return "DeclarationFile"
	}
	return "Program"
}

// declaredTypes is every interface, type alias and class every source file declares, anywhere, a
// namespace's and a function's included, as change B's declaredTypes walked them (57a243f2).
func (p *ksetProbe) declaredTypes() []ksetDeclared {
	found := []ksetDeclared{}
	seen := map[*ast.Symbol]bool{}
	for _, file := range p.l.program.CompilerProgram().GetSourceFiles() {
		origin := ksetOrigin(file)
		var visit ast.Visitor
		visit = func(n *ast.Node) bool {
			switch n.Kind {
			case ast.KindInterfaceDeclaration, ast.KindTypeAliasDeclaration, ast.KindClassDeclaration, ast.KindClassExpression:
				if symbol := p.l.checker.GetSymbolOfDeclaration(n); symbol != nil && !seen[symbol] {
					seen[symbol] = true
					if declared := p.l.checker.GetDeclaredTypeOfSymbol(symbol); declared != nil {
						found = append(found, ksetDeclared{node: n, declared: declared, origin: origin, generic: len(n.TypeParameters()) != 0})
					}
				}
			}
			n.ForEachChild(visit)
			return false
		}
		file.AsNode().ForEachChild(visit)
	}
	return found
}

// kindOf judges a kind type. A literal under null and undefined is still one literal, marked optional by
// the caller; a union naming a whole numeric enum is OpenEnum, as openNumericEnumType judges it.
func (p *ksetProbe) kindOf(kind *checker.Type) ksetKindShape {
	if kind == nil {
		return ksetKindShape{class: "Missing"}
	}
	shape := ksetKindShape{text: p.l.checker.TypeToString(kind)}
	proven := p.l.checker.GetNonNullableType(kind)
	switch {
	case proven.Flags()&checker.TypeFlagsUnion == 0 && proven.Flags()&checker.TypeFlagsLiteral != 0:
		shape.class = "Literal"
		shape.members = []string{p.l.checker.TypeToString(proven)}
	case p.l.openNumericEnumType(proven):
		shape.class = "OpenEnum"
	case proven.Flags()&checker.TypeFlagsUnion != 0 && ksetLiteralUnion(proven):
		shape.class = "LiteralUnion"
		for _, member := range proven.Types() {
			shape.members = append(shape.members, p.l.checker.TypeToString(member))
		}
	case proven.Flags()&checker.TypeFlagsTypeParameter != 0:
		shape.class = "TypeParameter"
		if constraint := p.l.checker.GetBaseConstraintOfType(proven); constraint != nil && constraint.Flags()&checker.TypeFlagsTypeParameter == 0 {
			shape.constraint = p.l.checker.TypeToString(constraint)
			shape.constraintClass = p.kindOf(constraint).class
		}
	default:
		shape.class = "Other"
	}
	return shape
}

// ksetLiteralUnion is whether every member of a union is one literal.
func ksetLiteralUnion(union *checker.Type) bool {
	for _, member := range union.Types() {
		if member.Flags()&checker.TypeFlagsLiteral == 0 {
			return false
		}
	}
	return true
}

// propertyKind is the shape of a type's kind property, marked optional when the property is.
func (p *ksetProbe) propertyKind(owner *checker.Type) ksetKindShape {
	property := p.l.checker.GetPropertyOfType(owner, "kind")
	if property == nil {
		return ksetKindShape{class: "Missing"}
	}
	shape := p.kindOf(p.l.checker.GetTypeOfSymbol(property))
	shape.optional = property.Flags&ast.SymbolFlagsOptional != 0
	return shape
}

// ksetShapeName is a declared type's own shape, so a reader can set aside an alias of a union, never or any.
func ksetShapeName(declared *checker.Type) string {
	switch {
	case declared.Flags()&checker.TypeFlagsObject != 0:
		return "Object"
	case declared.Flags()&checker.TypeFlagsIntersection != 0:
		return "Intersection"
	case declared.Flags()&checker.TypeFlagsUnion != 0:
		return "Union"
	}
	return "Other"
}

// ksetDeclarationName names a declaration's syntax.
func ksetDeclarationName(node *ast.Node) string {
	switch node.Kind {
	case ast.KindInterfaceDeclaration:
		return "Interface"
	case ast.KindTypeAliasDeclaration:
		return "TypeAlias"
	case ast.KindClassDeclaration:
		return "Class"
	}
	return "ClassExpression"
}

// brands is target's brand members as brandedSupertype finds them: required, not symbol-keyed, phantom.
func (p *ksetProbe) brands(target *checker.Type) []string {
	names := []string{}
	for _, property := range p.l.checker.GetPropertiesOfType(target) {
		if property.Flags&ast.SymbolFlagsOptional != 0 || checker.IsKnownSymbol(property) {
			continue
		}
		if phantomField(p.l.checker.GetTypeOfSymbol(property)) {
			names = append(names, property.Name)
		}
	}
	return names
}

// kindSets writes K(S) for each declaration the program gives a target's name, then its summary. Library
// declarations aren't S, but one assignable to S is still a row, and the summary lists it apart.
func (p *ksetProbe) kindSets(declared []ksetDeclared) {
	for _, name := range ksetProbeTargets {
		targets := []ksetDeclared{}
		for _, candidate := range declared {
			if candidate.origin != "Program" || candidate.node.Kind == ast.KindClassExpression {
				continue
			}
			if declarationName := candidate.node.Name(); declarationName != nil && ast.IsIdentifier(declarationName) && declarationName.Text() == name {
				targets = append(targets, candidate)
			}
		}
		if len(targets) == 0 {
			p.rows = append(p.rows, map[string]any{"row": "summary", "s": name, "found": false})
			continue
		}
		for _, target := range targets {
			p.kindSet(name, target, declared)
		}
	}
}

func (p *ksetProbe) kindSet(name string, target ksetDeclared, declared []ksetDeclared) {
	sSite := p.l.program.Where(target.node)
	literal, literalUnion, generic, class := 0, 0, 0, 0
	open, unions, others, library := []string{}, []string{}, []string{}, []string{}
	for _, candidate := range declared {
		if !p.l.checker.IsTypeAssignableTo(candidate.declared, target.declared) {
			continue
		}
		typeText := p.l.checker.TypeToString(candidate.declared)
		shapeName := ksetShapeName(candidate.declared)
		isClass := candidate.node.Kind == ast.KindClassDeclaration || candidate.node.Kind == ast.KindClassExpression || isClassInstance(candidate.declared)
		kind := p.propertyKind(candidate.declared)
		p.rows = append(p.rows, map[string]any{
			"row":             "kset",
			"s":               name,
			"sSite":           sSite,
			"type":            typeText,
			"typeId":          int(candidate.declared.Id()),
			"site":            p.l.program.Where(candidate.node),
			"origin":          candidate.origin,
			"declaration":     ksetDeclarationName(candidate.node),
			"shape":           shapeName,
			"self":            candidate.declared == target.declared,
			"generic":         candidate.generic,
			"class":           isClass,
			"branded":         p.l.brandedSupertype(candidate.declared),
			"kindType":        kind.text,
			"kindClass":       kind.class,
			"kindMembers":     kind.members,
			"kindOptional":    kind.optional,
			"constraint":      kind.constraint,
			"constraintClass": kind.constraintClass,
		})
		switch {
		case candidate.origin != "Program":
			library = append(library, typeText+" ("+candidate.origin+")")
			continue
		case shapeName == "Union":
			unions = append(unions, typeText)
			continue
		case shapeName == "Other":
			others = append(others, typeText)
			continue
		}
		switch kind.class {
		case "Literal":
			literal++
		case "LiteralUnion":
			literalUnion++
		default:
			open = append(open, typeText)
		}
		if candidate.generic {
			generic++
		}
		if isClass {
			class++
		}
	}
	kind := p.propertyKind(target.declared)
	// Counts are over the program's object and intersection types; union aliases, other shapes and
	// library types are named apart.
	p.rows = append(p.rows, map[string]any{
		"row":          "summary",
		"s":            name,
		"found":        true,
		"sSite":        sSite,
		"sKindType":    kind.text,
		"sKindClass":   kind.class,
		"branded":      p.l.brandedSupertype(target.declared),
		"brands":       p.brands(target.declared),
		"literal":      literal,
		"literalUnion": literalUnion,
		"open":         open,
		"generic":      generic,
		"class":        class,
		"unionAliases": unions,
		"otherShapes":  others,
		"library":      library,
	})
}

// build writes one build row when kind isn't one literal.
func (p *ksetProbe) build(site *ast.Node, construct string, built *checker.Type, kind ksetKindShape) {
	if kind.class == "Literal" || kind.class == "Missing" {
		return
	}
	typeText := ""
	if built != nil {
		typeText = p.l.checker.TypeToString(built)
	}
	p.rows = append(p.rows, map[string]any{
		"row":             "build",
		"site":            p.l.program.Where(site),
		"construct":       construct,
		"type":            typeText,
		"kindType":        kind.text,
		"kindClass":       kind.class,
		"kindMembers":     kind.members,
		"kindOptional":    kind.optional,
		"constraint":      kind.constraint,
		"constraintClass": kind.constraintClass,
	})
}

// builds walks the program's own files for every place a value gets a kind that isn't one literal: an
// object literal's kind member, by its value expression's type before the literal widens it; an object
// literal with a spread and no kind member of its own; Object.assign; a new of a class whose instance's
// kind is open, which is how tsc's createBaseNode builds every node; a class whose instances' kind is open;
// and an assignment to a .kind property, which is how tsc's Node function, the constructor createBaseNode
// reaches through objectAllocator, writes it.
func (p *ksetProbe) builds() {
	for _, file := range p.l.program.CompilerProgram().GetSourceFiles() {
		if ksetOrigin(file) != "Program" {
			continue
		}
		var visit ast.Visitor
		visit = func(n *ast.Node) bool {
			switch n.Kind {
			case ast.KindObjectLiteralExpression:
				p.objectLiteral(n)
			case ast.KindNewExpression:
				built := p.l.checker.GetTypeAtLocation(n)
				p.build(n, "New", built, p.propertyKind(built))
			case ast.KindClassDeclaration, ast.KindClassExpression:
				p.class(n)
			case ast.KindCallExpression:
				callee := n.AsCallExpression().Expression
				if callee.Kind == ast.KindPropertyAccessExpression && callee.Name().Text() == "assign" && p.l.isLibraryGlobal(callee.AsPropertyAccessExpression().Expression, "Object") {
					built := p.l.checker.GetTypeAtLocation(n)
					p.build(n, "ObjectAssign", built, p.propertyKind(built))
				}
			case ast.KindBinaryExpression:
				if n.AsBinaryExpression().OperatorToken.Kind == ast.KindEqualsToken {
					left := ast.SkipParentheses(n.AsBinaryExpression().Left)
					if left.Kind == ast.KindPropertyAccessExpression && left.Name().Text() == "kind" {
						right := n.AsBinaryExpression().Right
						p.build(n, "KindAssignment", p.l.checker.GetTypeAtLocation(left.AsPropertyAccessExpression().Expression), p.kindOf(p.l.checker.GetTypeAtLocation(right)))
					}
				}
			}
			n.ForEachChild(visit)
			return false
		}
		file.AsNode().ForEachChild(visit)
	}
}

// objectLiteral judges an object literal by its own kind member's value, or, with none, by the kind its
// spreads give it. One that's a destructuring assignment's pattern builds nothing.
func (p *ksetProbe) objectLiteral(literal *ast.Node) {
	if ast.IsAssignmentTarget(literal) {
		return
	}
	built := p.l.checker.GetTypeAtLocation(literal)
	spread := false
	for _, property := range literal.AsObjectLiteralExpression().Properties.Nodes {
		if property.Kind == ast.KindSpreadAssignment {
			spread = true
			continue
		}
		name := property.Name()
		if name == nil || (!ast.IsIdentifier(name) && name.Kind != ast.KindStringLiteral) || name.Text() != "kind" {
			continue
		}
		switch property.Kind {
		case ast.KindPropertyAssignment:
			p.build(property, "ObjectLiteral", built, p.kindOf(p.l.checker.GetTypeAtLocation(property.AsPropertyAssignment().Initializer)))
		case ast.KindShorthandPropertyAssignment:
			// The value's symbol, as shorthand lowers it (object.go), typed as declared, not as narrowed here.
			symbol := p.l.checker.GetShorthandAssignmentValueSymbol(property)
			if symbol != nil && symbol.Flags&ast.SymbolFlagsAlias != 0 {
				symbol = p.l.checker.GetAliasedSymbol(symbol)
			}
			if symbol == nil {
				p.build(property, "ObjectLiteralShorthand", built, p.propertyKind(built))
			} else {
				p.build(property, "ObjectLiteralShorthand", built, p.kindOf(p.l.checker.GetTypeOfSymbol(p.l.checker.GetExportSymbolOfSymbol(symbol))))
			}
		default:
			p.build(property, "ObjectLiteralAccessor", built, p.propertyKind(built))
		}
		return
	}
	if spread {
		p.build(literal, "ObjectSpread", built, p.propertyKind(built))
	}
}

// class judges a class by its instances' kind, and says whether this class declares the member.
func (p *ksetProbe) class(node *ast.Node) {
	symbol := p.l.checker.GetSymbolOfDeclaration(node)
	if symbol == nil {
		return
	}
	declared := p.l.checker.GetDeclaredTypeOfSymbol(symbol)
	if declared == nil {
		return
	}
	construct := "ClassInherited"
	if property := p.l.checker.GetPropertyOfType(declared, "kind"); property != nil {
		for _, declaration := range property.Declarations {
			if declaration.Parent == node {
				construct = "Class"
			}
		}
	}
	p.build(node, construct, declared, p.propertyKind(declared))
}

// write appends every row as one JSON line.
func (p *ksetProbe) write() {
	ksetProbeMutex.Lock()
	defer ksetProbeMutex.Unlock()
	file, err := os.OpenFile(ksetProbePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer file.Close()
	for _, row := range p.rows {
		line, err := json.Marshal(row)
		if err != nil {
			continue
		}
		_, _ = file.Write(append(line, '\n'))
	}
}
