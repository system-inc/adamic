package checker

import (
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// usageShape exposes type topology, declaration identities and signatures.
// Counts, traversal multiplicity and all rule judgments belong to native Adamic.
func (p *Program) usageShape(c *checker.Checker, node *ast.Node, question string) (string, error) {
	if !strings.HasPrefix(question, "usage-shape") {
		return "", fmt.Errorf("unsupported checker question: %s", question)
	}
	if question != "usage-shape" {
		return "", fmt.Errorf("unexpected usage-shape suffix")
	}
	if !ast.IsFunctionLike(node) && node.Kind != ast.KindTypeParameter && node.Kind != ast.KindPropertyDeclaration {
		return "", fmt.Errorf("usage-shape requires a signature, type parameter or property")
	}
	g := &usageGraph{p: p, c: c, seen: map[*checker.Type]int{}}
	root := 0
	if node.Kind == ast.KindCallSignature || node.Kind == ast.KindConstructor {
		root = g.signature(c.GetSignatureFromDeclaration(node))
	} else {
		root = g.add(c.GetTypeAtLocation(node))
	}
	out := &fields{}
	out.number(1)
	out.text("usage-shape")
	out.number(uint64(root))
	out.number(uint64(len(g.records)))
	for _, r := range g.records {
		out.number(r.flags)
		out.number(r.objects)
		out.number(r.declarationStart)
		out.number(r.declarationEnd)
		out.text(r.declarationFile)
		out.number(uint64(r.constraint))
		out.number(uint64(r.defaultType))
		out.yes(r.array)
		out.yes(r.tuple)
		out.yes(r.readonly)
		out.text(r.name)
		for _, list := range [][]int{r.alias, r.parts, r.arguments, r.properties, r.indexes, r.signatures, r.parameters, r.typeParameters, r.returns} {
			out.number(uint64(len(list)))
			for _, id := range list {
				out.number(uint64(id))
			}
		}
	}
	return out.String(), nil
}

type usageRecord struct {
	flags, objects, declarationStart, declarationEnd                                              uint64
	declarationFile, name                                                                         string
	constraint, defaultType                                                                       int
	array, tuple, readonly                                                                        bool
	alias, parts, arguments, properties, indexes, signatures, parameters, typeParameters, returns []int
}
type usageGraph struct {
	p       *Program
	c       *checker.Checker
	seen    map[*checker.Type]int
	records []usageRecord
}

func (g *usageGraph) list(ts []*checker.Type) []int {
	var ids []int
	for _, t := range ts {
		ids = append(ids, g.add(t))
	}
	return ids
}
func (g *usageGraph) signature(s *checker.Signature) int {
	if s == nil {
		return 0
	}
	id := len(g.records) + 1
	g.records = append(g.records, usageRecord{})
	r := usageRecord{}
	if receiver := s.ThisParameter(); receiver != nil {
		r.parameters = append(r.parameters, g.add(g.c.GetTypeOfSymbol(receiver)))
	}
	for _, param := range s.Parameters() {
		r.parameters = append(r.parameters, g.add(g.c.GetTypeOfSymbol(param)))
	}
	r.typeParameters = g.list(s.TypeParameters())
	returned := g.c.GetReturnTypeOfSignature(s)
	if predicate := g.c.GetTypePredicateOfSignature(s); predicate != nil && predicate.Type() != nil {
		returned = predicate.Type()
	}
	r.returns = []int{g.add(returned)}
	g.records[id-1] = r
	return id
}
func (g *usageGraph) add(t *checker.Type) int {
	if t == nil {
		return 0
	}
	if id := g.seen[t]; id != 0 {
		return id
	}
	id := len(g.records) + 1
	g.seen[t] = id
	g.records = append(g.records, usageRecord{})
	r := usageRecord{flags: uint64(t.Flags()), objects: uint64(t.ObjectFlags())}
	if symbol := t.Symbol(); symbol != nil {
		r.name = symbol.Name
	}
	switch {
	case t.IsTypeParameter():
		if s := t.Symbol(); s != nil && len(s.Declarations) > 0 {
			d := s.Declarations[0]
			if d.Kind == ast.KindTypeParameter {
				r.declarationStart = uint64(d.Pos())
				r.declarationEnd = uint64(d.End())
				r.declarationFile = ast.GetSourceFileOfNode(d).FileName().AsString()
				data := d.AsTypeParameterDeclaration()
				if data.Constraint != nil {
					r.constraint = g.add(g.c.GetTypeAtLocation(data.Constraint))
				}
				if data.DefaultType != nil {
					r.defaultType = g.add(g.c.GetTypeAtLocation(data.DefaultType))
				}
			}
		}
	case t.Alias() != nil && len(t.Alias().TypeArguments()) > 0:
		r.alias = g.list(t.Alias().TypeArguments())
	case t.Flags()&checker.TypeFlagsUnionOrIntersection != 0:
		r.parts = g.list(t.Types())
	case t.Flags()&checker.TypeFlagsIndexedAccess != 0:
		r.parts = []int{g.add(t.AsIndexedAccessType().ObjectType()), g.add(t.AsIndexedAccessType().IndexType())}
	case t.Flags()&checker.TypeFlagsObject != 0 && t.ObjectFlags()&checker.ObjectFlagsReference != 0:
		r.arguments = g.list(g.c.GetTypeArguments(t))
		target := t.Target()
		if target != nil {
			r.tuple = target.IsTupleType()
			if r.tuple {
				r.readonly = target.AsTupleType().IsReadonly()
			}
			r.array = g.c.IsArrayType(target)
		}
	case t.Flags()&checker.TypeFlagsTemplateLiteral != 0:
		r.parts = g.list(t.AsTemplateLiteralType().Types())
	case t.Flags()&checker.TypeFlagsConditional != 0:
		r.parts = []int{g.add(t.AsConditionalType().CheckType()), g.add(t.AsConditionalType().ExtendsType())}
	case t.Flags()&checker.TypeFlagsObject != 0:
		for _, symbol := range g.c.GetPropertiesOfType(t) {
			r.properties = append(r.properties, g.add(g.c.GetTypeOfSymbol(symbol)))
		}
		r.indexes = []int{g.add(g.c.GetNumberIndexType(t)), g.add(g.c.GetStringIndexType(t))}
		for _, kind := range []checker.SignatureKind{checker.SignatureKindCall, checker.SignatureKindConstruct} {
			for _, s := range g.c.GetSignaturesOfType(t, kind) {
				r.signatures = append(r.signatures, g.signature(s))
			}
		}
	case t.Flags()&checker.TypeFlagsIndex != 0:
		r.parts = []int{g.add(t.AsIndexType().Target())}
	}
	g.records[id-1] = r
	return id
}
