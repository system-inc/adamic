// Checker facts for the six-rule native pilot. No lint predicates live here.
package checker

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
)

// Each field is decimal UTF-16 length LF text. The enclosing C buffer is UTF-8.
type fields struct{ strings.Builder }

func (f *fields) text(s string) {
	units := 0
	for _, r := range s {
		units++
		if r > 65535 {
			units++
		}
	}
	var digits [20]byte
	f.Write(strconv.AppendInt(digits[:0], int64(units), 10))
	f.WriteByte('\n')
	f.WriteString(s)
}
func (f *fields) number(n uint64) {
	var digits [20]byte
	value := strconv.AppendUint(digits[:0], n, 10)
	// Numeric fields contain ASCII only; no formatting or rune allocation is needed.
	var length [2]byte
	f.Write(strconv.AppendInt(length[:0], int64(len(value)), 10))
	f.WriteByte('\n')
	f.Write(value)
}
func (f *fields) yes(b bool) {
	if b {
		f.number(1)
	} else {
		f.number(0)
	}
}
func (f *fields) ids(ids []uint64) {
	f.number(uint64(len(ids)))
	for _, id := range ids {
		f.number(id)
	}
}

func (p *Program) exact(file string, start, end uint64, kind string) (*ast.SourceFile, *ast.Node, error) {
	path, err := Path(file)
	if err != nil {
		return nil, nil, err
	}
	source := p.Compiler.GetSourceFile(path)
	if source == nil {
		return nil, nil, fmt.Errorf("file is not in this program: %s", path)
	}
	if (start >= end && !(start == 0 && end == 0 && kind == "SourceFile" && len(source.Text()) == 0)) || end > uint64(len(source.Text())) {
		return nil, nil, fmt.Errorf("invalid node range")
	}
	var node *ast.Node
	var find func(*ast.Node) bool
	find = func(candidate *ast.Node) bool {
		if uint64(candidate.Pos()) > start || uint64(candidate.End()) < end {
			return false
		}
		if uint64(candidate.Pos()) == start && uint64(candidate.End()) == end && candidate.Kind.String() == "Kind"+kind {
			node = candidate
			return true
		}
		return candidate.ForEachChild(find)
	}
	find(source.AsNode())
	if node == nil {
		return nil, nil, fmt.Errorf("no exact %s node at %d:%d in %s", kind, start, end, file)
	}
	return source, node, nil
}

type typeRecord struct {
	id, flags, target, constraint, element uint64
	name                                   string
	errorType, array                       bool
	tupleFlags                             int
	parts, arguments                       []uint64
}
type graph struct {
	program *Program
	checker *checker.Checker
	records []typeRecord
	seen    map[*checker.Type]bool
	names   bool
}

func (g *graph) id(t *checker.Type) uint64 {
	if t == nil {
		return 0
	}
	id := g.program.typeIDs[t]
	if id == 0 {
		id = uint64(len(g.program.typeIDs)) + 1
		if id > (1<<53)-1 {
			panic("type identity space exhausted")
		}
		g.program.typeIDs[t] = id
		g.program.typesByID = append(g.program.typesByID, t)
	}
	return id
}
func (g *graph) add(t *checker.Type) uint64 {
	id := g.id(t)
	if t == nil || g.seen[t] {
		return id
	}
	g.seen[t] = true
	// Reserve before descending so recursive references terminate.
	index := len(g.records)
	g.records = append(g.records, typeRecord{})
	r := typeRecord{id: id, flags: uint64(t.Flags()), tupleFlags: -1}
	if g.names {
		r.name = g.checker.TypeToString(t)
	}
	r.errorType = t.Flags()&checker.TypeFlagsIntrinsic != 0 && t.AsIntrinsicType().IntrinsicName() == "error"
	if checker.IsNonDeferredTypeReference(t) {
		r.target = g.id(t.Target())
		for _, argument := range checker.Checker_getTypeArguments(g.checker, t) {
			r.arguments = append(r.arguments, g.add(argument))
		}
	}
	if t.Flags()&checker.TypeFlagsUnionOrIntersection != 0 {
		for _, part := range t.Types() {
			r.parts = append(r.parts, g.add(part))
		}
	}
	r.array = checker.Checker_isArrayType(g.checker, t)
	if checker.IsTupleType(t) {
		r.tupleFlags = int(checker.TupleType_combinedFlags(t.TargetTupleType()))
	}
	// Only type parameters need a constraint in the wire graph. Rest parameters
	// select their constrained root in Inspect before an index type is requested.
	if t.Flags()&checker.TypeFlagsTypeParameter != 0 {
		r.constraint = g.add(checker.Checker_getBaseConstraintOfType(g.checker, t))
	}
	g.records[index] = r
	return id
}
func (g *graph) write(out *fields) {
	out.number(uint64(len(g.records)))
	for _, r := range g.records {
		out.number(r.id)
		out.number(r.flags)
		out.text(r.name)
		out.yes(r.errorType)
		out.number(r.target)
		out.yes(r.array)
		out.text(strconv.Itoa(r.tupleFlags))
		out.number(r.constraint)
		out.number(r.element)
		out.ids(r.parts)
		out.ids(r.arguments)
	}
}
func constrained(c *checker.Checker, t *checker.Type) *checker.Type {
	if bound := checker.Checker_getBaseConstraintOfType(c, t); bound != nil {
		return bound
	}
	return t
}

// Inspect's supported questions are documented in facts.md. Exact nodes and
// opaque per-program type identities are facts, never addresses or owned handles.
func (p *Program) Inspect(file string, start, end uint64, kind, question string) (string, error) {
	source, node, err := p.exact(file, start, end, kind)
	if err != nil {
		return "", err
	}
	c, release := p.Compiler.GetTypeCheckerForFile(context.Background(), source)
	defer release()
	out := &fields{}
	out.number(1)
	mode := strings.Split(question, "\n")[0]
	out.text(mode)
	switch mode {
	case "options":
		if question != mode || node.Kind != ast.KindSourceFile {
			return "", fmt.Errorf("options requires a SourceFile")
		}
		out.yes(p.Compiler.Options().GetStrictOptionValue(p.Compiler.Options().StrictNullChecks))
	case "assignable":
		split := strings.Split(question, "\n")
		if len(split) != 4 {
			return "", fmt.Errorf("assignable requires target start, end and kind")
		}
		first, e1 := strconv.ParseUint(split[1], 10, 64)
		last, e2 := strconv.ParseUint(split[2], 10, 64)
		if e1 != nil || e2 != nil {
			return "", fmt.Errorf("invalid assignability target")
		}
		_, target, e := p.exact(file, first, last, split[3])
		if e != nil {
			return "", e
		}
		out.yes(checker.Checker_isTypeAssignableTo(c, c.GetTypeAtLocation(node), c.GetTypeAtLocation(target)))
	case "name":
		split := strings.Split(question, "\n")
		if len(split) != 2 {
			return "", fmt.Errorf("name requires a type identity")
		}
		id, err := strconv.ParseUint(split[1], 10, 64)
		if err != nil || id == 0 || id > uint64(len(p.typesByID)) || strconv.FormatUint(id, 10) != split[1] {
			return "", fmt.Errorf("unknown checker type identity")
		}
		out.text(c.TypeToString(p.typesByID[id-1]))
	case "declarations":
		if question != mode {
			return "", fmt.Errorf("unexpected declaration question suffix")
		}
		switch node.Kind {
		case ast.KindClassDeclaration, ast.KindInterfaceDeclaration:
		default:
			return "", fmt.Errorf("declarations requires a class or interface")
		}
		var symbol *ast.Symbol
		if name := node.Name(); name != nil {
			symbol = c.GetSymbolAtLocation(name)
		}
		write := func(symbol *ast.Symbol) {
			out.yes(symbol != nil)
			if symbol == nil {
				out.number(0)
				return
			}
			out.number(uint64(len(symbol.Declarations)))
			for _, declaration := range symbol.Declarations {
				out.text(strings.TrimPrefix(declaration.Kind.String(), "Kind"))
				out.number(uint64(declaration.Pos()))
				out.number(uint64(declaration.End()))
				f := ast.GetSourceFileOfNode(declaration)
				if f == nil {
					panic("declaration has no source")
				}
				out.text(f.FileName())
				if name := declaration.Name(); name != nil {
					out.number(uint64(scanner.GetTokenPosOfNode(name, source, false)))
					out.number(uint64(name.End()))
				} else {
					out.number(0)
					out.number(0)
				}
			}
		}
		write(symbol)
		write(node.LocalSymbol())
	case "raw-type", "type", "base-type", "signature", "raw-shape", "type-shape", "signature-shape":
		if question != mode {
			return "", fmt.Errorf("unexpected type question suffix")
		}
		g := &graph{program: p, checker: c, seen: make(map[*checker.Type]bool), names: !strings.HasSuffix(mode, "-shape")}
		var roots []uint64
		var rest []bool
		present := true
		if mode == "signature" || mode == "signature-shape" {
			switch node.Kind {
			case ast.KindCallExpression, ast.KindNewExpression, ast.KindTaggedTemplateExpression:
			default:
				return "", fmt.Errorf("signature requires a call-like node")
			}
			var callee *ast.Node
			if node.Kind == ast.KindTaggedTemplateExpression {
				callee = node.AsTaggedTemplateExpression().Tag
			} else {
				callee = node.Expression()
			}
			roots = append(roots, g.add(c.GetTypeAtLocation(callee)))
			signature := c.GetResolvedSignature(node)
			present = signature != nil
			if signature != nil {
				for _, parameter := range checker.Signature_parameters(signature) {
					t := c.GetTypeOfSymbolAtLocation(parameter, node)
					if t == nil {
						continue
					}
					isRest := len(parameter.Declarations) > 0 && parameter.Declarations[0].Kind == ast.KindParameter && parameter.Declarations[0].AsParameterDeclaration().DotDotDotToken != nil
					if isRest {
						t = constrained(c, t)
						root := g.add(t)
						// The number index is needed only for a signature's rest parameter.
						element := g.add(checker.Checker_getIndexTypeOfType(c, t, checker.Checker_numberType(c)))
						for i := range g.records {
							if g.records[i].id == root {
								g.records[i].element = element
							}
						}
						roots = append(roots, root)
					} else {
						roots = append(roots, g.add(t))
					}
					rest = append(rest, isRest)
					if isRest {
						break
					}
				}
			}
		} else {
			t := c.GetTypeAtLocation(node)
			if mode != "raw-type" && mode != "raw-shape" {
				t = constrained(c, t)
			}
			if mode == "base-type" {
				t = checker.Checker_getBaseTypeOfLiteralType(c, t)
			}
			roots = append(roots, g.add(t))
		}
		out.yes(p.Compiler.Options().GetStrictOptionValue(p.Compiler.Options().StrictNullChecks))
		out.yes(present)
		out.ids(roots)
		out.number(uint64(len(rest)))
		for _, r := range rest {
			out.yes(r)
		}
		g.write(out)
	default:
		return "", fmt.Errorf("unsupported checker question: %s", question)
	}
	return out.String(), nil
}
