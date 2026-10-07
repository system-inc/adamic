package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"strconv"
	"strings"
)

// Return compiler JSDoc metadata for declarations and each immediate alias hop.
func documentationTag(out *fields, declaration *ast.Node) {
	var tag *ast.Node
	if declaration != nil && ast.GetCombinedNodeFlags(declaration)&ast.NodeFlagsPossiblyContainsDeprecatedTag != 0 {
		for current := declaration; current != nil; current = current.Parent {
			if current.Flags&ast.NodeFlagsPossiblyContainsDeprecatedTag != 0 {
				tag = ast.GetJSDocDeprecatedTag(current)
				break
			}
		}
	}
	out.yes(tag != nil)
	var comment strings.Builder
	if tag != nil {
		if pieces := tag.AsJSDocDeprecatedTag().Comment; pieces != nil {
			for _, piece := range pieces.Nodes {
				comment.WriteString(piece.Text())
			}
		}
	}
	out.text(comment.String())
}
func documentationSymbol(out *fields, symbol *ast.Symbol) {
	out.yes(symbol != nil)
	if symbol == nil {
		return
	}
	out.number(uint64(symbol.Flags))
	firstKind := ""
	if len(symbol.Declarations) > 0 {
		firstKind = strings.TrimPrefix(symbol.Declarations[0].Kind.String(), "Kind")
	}
	out.text(firstKind)
	out.number(uint64(len(symbol.Declarations)))
	for _, declaration := range symbol.Declarations {
		documentationTag(out, declaration)
	}
}
func (p *Program) symbolDocumentation(out *fields, c *checker.Checker, _ *ast.SourceFile, node *ast.Node, question string) error {
	split := strings.SplitN(question, "\n", 4)
	if len(split) < 2 {
		return fmt.Errorf("symbol-documentation requires an operation")
	}
	if split[1] == "signature" {
		if len(split) != 2 {
			return fmt.Errorf("unexpected signature suffix")
		}
		switch node.Kind {
		case ast.KindCallExpression, ast.KindNewExpression, ast.KindTaggedTemplateExpression, ast.KindJsxOpeningElement, ast.KindJsxSelfClosingElement:
		default:
			return fmt.Errorf("signature documentation requires a call")
		}
		signature := c.GetResolvedSignature(node)
		var declaration *ast.Node
		if signature != nil {
			declaration = signature.Declaration()
		}
		documentationTag(out, declaration)
		return nil
	}
	var symbol *ast.Symbol
	switch split[1] {
	case "node":
		if len(split) != 2 {
			return fmt.Errorf("unexpected node suffix")
		}
		symbol = c.GetSymbolAtLocation(node)
	case "shorthand":
		if len(split) != 2 {
			return fmt.Errorf("unexpected shorthand suffix")
		}
		symbol = c.GetSymbolAtLocation(node)
		if symbol != nil && symbol.ValueDeclaration != nil && symbol.ValueDeclaration.Kind == ast.KindShorthandPropertyAssignment {
			symbol = c.GetShorthandAssignmentValueSymbol(symbol.ValueDeclaration)
		} else {
			symbol = nil
		}
	case "property":
		if len(split) != 4 {
			return fmt.Errorf("property documentation requires type and name")
		}
		id, err := strconv.ParseUint(split[2], 10, 64)
		if err != nil || id == 0 || id > uint64(len(p.typesByID)) || strconv.FormatUint(id, 10) != split[2] {
			return fmt.Errorf("unknown checker type identity")
		}
		symbol = c.GetPropertyOfType(p.typesByID[id-1], split[3])
	default:
		return fmt.Errorf("unknown documentation operation")
	}
	var target *ast.Symbol
	if symbol != nil && symbol.Flags&ast.SymbolFlagsAlias != 0 {
		target = checker.Checker_resolveAlias(c, symbol)
	} else {
		target = symbol
	}
	documentationSymbol(out, target)
	var chain []*ast.Symbol
	seen := map[*ast.Symbol]bool{}
	for symbol != nil && !seen[symbol] {
		seen[symbol] = true
		chain = append(chain, symbol)
		if symbol.Flags&ast.SymbolFlagsAlias == 0 || len(symbol.Declarations) == 0 {
			break
		}
		symbol = checker.Checker_getImmediateAliasedSymbol(c, symbol)
	}
	out.number(uint64(len(chain)))
	for _, hop := range chain {
		out.yes(hop == target)
		documentationSymbol(out, hop)
	}
	return nil
}

func init() { additionalQuestions["symbol-documentation"] = (*Program).symbolDocumentation }
