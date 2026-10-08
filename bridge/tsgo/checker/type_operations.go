package checker

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// typeOperations exposes individual checker operations and declaration syntax.
// Recursive lint predicates and report construction stay in Adamic.
func (p *Program) typeOperations(out *fields, c *checker.Checker, node *ast.Node, question string) (string, error) {
	parts := strings.SplitN(question, "\n", 4)
	if len(parts) < 3 {
		return "", fmt.Errorf("type-operations requires operation and identity")
	}
	operation := parts[1]
	id, err := strconv.ParseUint(parts[2], 10, 64)
	if err != nil || strconv.FormatUint(id, 10) != parts[2] || (operation != "argument" && id > uint64(len(p.typesByID))) {
		return "", fmt.Errorf("invalid type-operations identity")
	}
	var subject *checker.Type
	if id != 0 && operation != "argument" {
		subject = p.typesByID[id-1]
	}
	g := &graph{program: p, checker: c, seen: make(map[*checker.Type]bool), allReferences: true}
	var roots []uint64
	add := func(t *checker.Type) {
		if t != nil {
			roots = append(roots, g.add(t))
		}
	}
	location := node

	switch operation {
	case "raw", "constrained", "contextual", "heritage", "access-name":
		if id != 0 || len(parts) != 3 {
			return "", fmt.Errorf("%s requires zero identity and no suffix", operation)
		}
		switch operation {
		case "raw":
			add(c.GetTypeAtLocation(node))
		case "constrained":
			add(constrained(c, c.GetTypeAtLocation(node)))
		case "contextual":
			add(checker.Checker_getContextualType(c, node, checker.ContextFlagsNone))
		case "access-name":
			if !ast.IsAccessExpression(node) {
				return "", fmt.Errorf("access-name requires an access")
			}
			name, present := checker.Checker_getAccessedPropertyName(c, node)
			out.yes(present)
			out.text(name)
		case "heritage":
			if !ast.IsClassLike(node) && node.Kind != ast.KindInterfaceDeclaration {
				return "", fmt.Errorf("heritage requires a class or interface")
			}
			var clauses *ast.NodeList
			if node.Kind == ast.KindInterfaceDeclaration {
				clauses = node.AsInterfaceDeclaration().HeritageClauses
			} else {
				if node.Kind == ast.KindClassExpression {
					clauses = node.AsClassExpression().HeritageClauses
				} else {
					clauses = node.AsClassDeclaration().HeritageClauses
				}
			}
			if clauses != nil {
				for _, clause := range clauses.Nodes {
					for _, element := range clause.AsHeritageClause().Types.Nodes {
						add(c.GetTypeAtLocation(element))
					}
				}
			}
		}
	case "argument":
		if len(parts) != 3 || (node.Kind != ast.KindCallExpression && node.Kind != ast.KindNewExpression) || id >= uint64(len(node.Arguments())) {
			return "", fmt.Errorf("argument requires a call and valid argument index")
		}
		add(checker.Checker_getContextualTypeForArgumentAtIndex(c, node, int(id)))
	case "apparent", "bases", "property", "member", "iterator", "signatures", "construct", "metadata":
		if subject == nil {
			return "", fmt.Errorf("%s requires a live type identity", operation)
		}
		if operation == "property" || operation == "member" {
			if len(parts) != 4 {
				return "", fmt.Errorf("property requires a name")
			}
		} else if len(parts) != 3 {
			return "", fmt.Errorf("unexpected type-operations suffix")
		}
		switch operation {
		case "apparent":
			add(checker.Checker_getApparentType(c, subject))
		case "bases":
			if symbol := subject.Symbol(); symbol != nil && symbol.Flags&(ast.SymbolFlagsClass|ast.SymbolFlagsInterface) != 0 {
				for _, base := range checker.Checker_getBaseTypes(c, checker.Checker_getDeclaredTypeOfSymbol(c, symbol)) {
					add(base)
				}
			}
		case "property", "member":
			var symbol *ast.Symbol
			if operation == "member" && subject.Symbol() != nil {
				symbol = subject.Symbol().Members[parts[3]]
			}
			if symbol == nil {
				symbol = checker.Checker_getPropertyOfType(c, subject, parts[3])
			}
			if symbol != nil {
				add(c.GetTypeOfSymbolAtLocation(symbol, node))
			}
		case "iterator":
			out.yes(checker.Checker_getPropertyOfType(c, subject, checker.Checker_getPropertyNameForKnownSymbolName(c, "iterator")) != nil)
		case "signatures", "construct":
			kind := checker.SignatureKindCall
			if operation == "construct" {
				kind = checker.SignatureKindConstruct
			}
			signatures := c.GetSignaturesOfType(subject, kind)
			out.number(uint64(len(signatures)))
			for _, signature := range signatures {
				out.number(g.add(c.GetReturnTypeOfSignature(signature)))
				params := checker.Signature_parameters(signature)
				out.number(uint64(len(params)))
				for _, param := range params {
					out.number(g.add(c.GetTypeOfSymbolAtLocation(param, location)))
					decl := param.ValueDeclaration
					out.yes(decl != nil && decl.Kind == ast.KindParameter && decl.AsParameterDeclaration().DotDotDotToken != nil)
				}
			}
		case "metadata":
			out.number(uint64(subject.ObjectFlags()))
			out.number(uint64(len(checker.Checker_getPropertiesOfType(c, subject))))
			out.number(uint64(len(c.GetSignaturesOfType(subject, checker.SignatureKindConstruct))))
			symbol := subject.Symbol()
			out.number(p.symbolID(symbol))
			out.text("")
			if symbol == nil {
				out.text("")
				out.number(0)
			} else {
				out.text(symbol.Name)
				out.number(uint64(symbol.Flags))
			}
			valueKind := ""
			if symbol != nil && symbol.ValueDeclaration != nil {
				valueKind = strings.TrimPrefix(symbol.ValueDeclaration.Kind.String(), "Kind")
			}
			out.text(valueKind)
			count := 0
			if symbol != nil {
				count = len(symbol.Declarations)
			}
			out.number(uint64(count))
			if symbol != nil {
				for _, decl := range symbol.Declarations {
					source := ast.GetSourceFileOfNode(decl)
					if source == nil {
						return "", fmt.Errorf("declaration lacks source")
					}
					out.text(strings.TrimPrefix(decl.Kind.String(), "Kind"))
					out.text(source.FileName().AsString())
					out.yes(source.IsDeclarationFile)
					out.yes(p.Compiler.IsSourceFileDefaultLibrary(source.PathKey()))
					out.number(uint64(decl.Flags))
					out.number(uint64(len(c.GetSignaturesOfType(c.GetTypeOfSymbolAtLocation(symbol, decl), checker.SignatureKindConstruct))))
					var members []*ast.Node
					if ast.IsClassLike(decl) {
						members = decl.Members()
					}
					out.number(uint64(len(members)))
					holders := []*ast.Node{decl}
					for _, member := range members {
						out.text(strings.TrimPrefix(member.Kind.String(), "Kind"))
						out.number(uint64(member.ModifierFlags()))
						nameKind := ""
						if member.Name() != nil {
							nameKind = strings.TrimPrefix(member.Name().Kind.String(), "Kind")
						}
						out.text(nameKind)
						holders = append(holders, member)
						if ast.IsConstructorDeclaration(member) {
							holders = append(holders, member.Parameters()...)
						}
					}
					var decorators []*ast.Node
					for _, holder := range holders {
						if mods := holder.Modifiers(); mods != nil {
							for _, mod := range mods.Nodes {
								if ast.IsDecorator(mod) {
									decorators = append(decorators, mod)
								}
							}
						}
					}
					out.number(uint64(len(decorators)))
					for _, decorator := range decorators {
						target := ast.SkipParentheses(decorator.AsDecorator().Expression)
						if ast.IsCallExpression(target) {
							target = ast.SkipParentheses(target.Expression())
						}
						if ast.IsPropertyAccessExpression(target) {
							target = target.Name()
						}
						var resolved *ast.Symbol
						if ast.IsIdentifier(target) {
							resolved = c.GetSymbolAtLocation(target)
							if resolved != nil {
								resolved = checker.SkipAlias(resolved, c)
							}
						}
						size := 0
						if resolved != nil {
							size = len(resolved.Declarations)
						}
						out.number(uint64(size))
						if resolved != nil {
							for _, origin := range resolved.Declarations {
								file := ast.GetSourceFileOfNode(origin)
								if file == nil {
									return "", fmt.Errorf("decorator declaration lacks source")
								}
								out.text(file.FileName().AsString())
								out.yes(file.IsDeclarationFile)
								out.number(uint64(origin.Flags))
							}
						}
					}
				}
			}
		}
	default:
		return "", fmt.Errorf("unsupported type operation %s", operation)
	}
	// A nested standard graph keeps identity validation identical to existing questions.
	out.number(1)
	out.text("type-operations")
	out.yes(p.Compiler.Options().GetStrictOptionValue(p.Compiler.Options().StrictNullChecks))
	out.yes(len(roots) > 0)
	out.ids(roots)
	out.number(0)
	g.write(out)
	return out.String(), nil
}
