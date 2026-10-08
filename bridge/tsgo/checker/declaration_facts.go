package checker

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

func (p *Program) writeDeclaration(out *fields, declaration *ast.Node) error {
	source := ast.GetSourceFileOfNode(declaration)
	if source == nil {
		return fmt.Errorf("declaration has no source")
	}
	out.text(source.FileName().AsString())
	out.text(strings.TrimPrefix(declaration.Kind.String(), "Kind"))
	out.number(uint64(declaration.Pos()))
	out.number(uint64(declaration.End()))
	out.yes(source.IsDeclarationFile)
	out.yes(p.Compiler.IsSourceFileDefaultLibrary(source.PathKey()))
	parent := declaration.Parent
	parentKind, parentName := "", ""
	first, last := uint64(0), uint64(0)
	if parent != nil {
		parentKind = strings.TrimPrefix(parent.Kind.String(), "Kind")
		first, last = uint64(parent.Pos()), uint64(parent.End())
		if name := parent.Name(); name != nil && (ast.IsPropertyNameLiteral(name) || name.Kind == ast.KindPrivateIdentifier) {
			parentName = name.Text()
		}
	}
	out.text(parentKind)
	out.text(parentName)
	out.number(first)
	out.number(last)
	var tags []*ast.Node
	for _, doc := range declaration.JSDoc(source) {
		if list := doc.AsJSDoc().Tags; list != nil {
			tags = append(tags, list.Nodes...)
		}
	}
	out.number(uint64(len(tags)))
	for _, tag := range tags {
		out.text(strings.TrimPrefix(tag.Kind.String(), "Kind"))
		name := ""
		if tag.TagName() != nil {
			name = tag.TagName().Text()
		}
		out.text(name)
		out.number(uint64(tag.Pos()))
		out.number(uint64(tag.End()))
		out.text(source.Text()[tag.Pos():tag.End()])
	}
	var parameters []*ast.Node
	if ast.IsFunctionLike(declaration) {
		if list := declaration.ParameterList(); list != nil {
			parameters = list.Nodes
		}
	}
	out.number(uint64(len(parameters)))
	for _, parameter := range parameters {
		name := ""
		if declared := parameter.Name(); declared != nil && declared.Kind == ast.KindIdentifier {
			name = declared.Text()
		}
		out.text(name)
	}
	return nil
}
func (p *Program) writeSymbolDetails(out *fields, symbol *ast.Symbol) error {
	out.yes(symbol != nil)
	if symbol == nil {
		return nil
	}
	out.number(p.symbolID(symbol))
	out.number(uint64(symbol.Flags))
	out.text(strings.ToValidUTF8(symbol.Name, "�"))
	out.number(uint64(len(symbol.Declarations)))
	for _, declaration := range symbol.Declarations {
		if err := p.writeDeclaration(out, declaration); err != nil {
			return err
		}
	}
	return nil
}
func (p *Program) declarationFacts(out *fields, c *checker.Checker, node *ast.Node, mode, question string) error {
	switch mode {
	case "node-symbol-details":
		if mode != question {
			return fmt.Errorf("unexpected symbol details suffix")
		}
		return p.writeSymbolDetails(out, c.GetSymbolAtLocation(node))
	case "declaration-details":
		if mode != question {
			return fmt.Errorf("unexpected declaration details suffix")
		}
		return p.writeDeclaration(out, node)
	case "type-symbol-details", "property-declarations":
		split := strings.SplitN(question, "\n", 3)
		count := 2
		if mode == "property-declarations" {
			count = 3
		}
		if len(split) != count {
			return fmt.Errorf("invalid declaration metadata question")
		}
		id, err := strconv.ParseUint(split[1], 10, 64)
		if err != nil || id == 0 || id > uint64(len(p.typesByID)) || strconv.FormatUint(id, 10) != split[1] {
			return fmt.Errorf("unknown checker type identity")
		}
		subject := p.typesByID[id-1]
		if mode == "type-symbol-details" {
			if err := p.writeSymbolDetails(out, subject.Symbol()); err != nil {
				return err
			}
			var alias *ast.Symbol
			if typeAlias := checker.Type_alias(subject); typeAlias != nil {
				alias = typeAlias.Symbol()
			}
			return p.writeSymbolDetails(out, alias)
		} else {
			return p.writeSymbolDetails(out, checker.Checker_getPropertyOfType(c, subject, split[2]))
		}
	}
	return nil
}
