// Checker facts for the native type-aware rules. No lint predicates live here.
package checker

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
)

// Each field is decimal UTF-16 length LF text. The enclosing C buffer is UTF-8.
type fields struct{ strings.Builder }

func (fields *fields) text(text string) {
	units := 0
	for _, character := range text {
		units++
		if character > 65535 {
			units++
		}
	}
	var digits [20]byte
	fields.Write(strconv.AppendInt(digits[:0], int64(units), 10))
	fields.WriteByte('\n')
	fields.WriteString(text)
}
func (fields *fields) number(n uint64) {
	var digits [20]byte
	value := strconv.AppendUint(digits[:0], n, 10)
	// Numeric fields contain ASCII only; no formatting or rune allocation is needed.
	var length [2]byte
	fields.Write(strconv.AppendInt(length[:0], int64(len(value)), 10))
	fields.WriteByte('\n')
	fields.Write(value)
}
func (fields *fields) yes(b bool) {
	if b {
		fields.number(1)
	} else {
		fields.number(0)
	}
}
func (fields *fields) ids(ids []uint64) {
	fields.number(uint64(len(ids)))
	for _, id := range ids {
		fields.number(id)
	}
}

// Only immutable AST locations are indexed. Checker facts are recomputed on
// every request, and the index belongs to this program's source-file identity.
type nodeRange struct {
	start, end uint64
}

func (program *Program) exact(file string, start, end uint64, kind string) (*ast.SourceFile, *ast.Node, error) {
	path, pathError := Path(file)
	if pathError != nil {
		return nil, nil, pathError
	}
	source := program.Compiler.GetSourceFile(tspath.RootedFilePathFromAbsolute(path))
	if source == nil {
		return nil, nil, fmt.Errorf("file is not in this program: %s", path)
	}
	if (start >= end && !(start == 0 && end == 0 && kind == "SourceFile" && len(source.Text()) == 0)) ||
		end > uint64(len(source.Text())) {
		return nil, nil, fmt.Errorf("invalid node range")
	}
	// Root-only metadata requests need no descendant index. In particular, a
	// large literal table with no operand queries should not pay for one.
	if start == uint64(source.Pos()) && end == uint64(source.End()) && kind == "SourceFile" {
		return source, source.AsNode(), nil
	}
	ranges := program.exactRanges[source]
	if ranges == nil {
		ranges = make(map[nodeRange][]*ast.Node)
		var index func(*ast.Node)
		index = func(candidate *ast.Node) {
			span := nodeRange{uint64(candidate.Pos()), uint64(candidate.End())}
			ranges[span] = append(ranges[span], candidate)
			candidate.ForEachChild(func(child *ast.Node) bool { index(child); return false })
		}
		// Preorder preserves the original first-match semantics for equal spans.
		index(source.AsNode())
		if program.exactRanges == nil {
			program.exactRanges = make(map[*ast.SourceFile]map[nodeRange][]*ast.Node)
		}
		program.exactRanges[source] = ranges
	}
	for _, candidate := range ranges[nodeRange{start, end}] {
		if candidate.Kind.String() == "Kind"+kind {
			return source, candidate, nil
		}
	}
	return nil, nil, fmt.Errorf("no exact %s node at %d:%d in %s", kind, start, end, file)
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

func (graph *graph) id(subjectType *checker.Type) uint64 {
	if subjectType == nil {
		return 0
	}
	id := graph.program.typeIDs[subjectType]
	if id == 0 {
		id = uint64(len(graph.program.typeIDs)) + 1
		if id > (1<<53)-1 {
			panic("type identity space exhausted")
		}
		graph.program.typeIDs[subjectType] = id
		graph.program.typesByID = append(graph.program.typesByID, subjectType)
	}
	return id
}
func (graph *graph) add(subjectType *checker.Type) uint64 {
	id := graph.id(subjectType)
	if subjectType == nil || graph.seen[subjectType] {
		return id
	}
	graph.seen[subjectType] = true
	// Reserve before descending so recursive references terminate.
	index := len(graph.records)
	graph.records = append(graph.records, typeRecord{})
	record := typeRecord{id: id, flags: uint64(subjectType.Flags()), tupleFlags: -1}
	if graph.names {
		record.name = graph.checker.TypeToString(subjectType)
	}
	record.errorType = subjectType.Flags()&checker.TypeFlagsIntrinsic != 0 &&
		subjectType.AsIntrinsicType().IntrinsicName() == "error"
	if checker.IsNonDeferredTypeReference(subjectType) {
		record.target = graph.id(subjectType.Target())
		for _, argument := range checker.Checker_getTypeArguments(graph.checker, subjectType) {
			record.arguments = append(record.arguments, graph.add(argument))
		}
	}
	if subjectType.Flags()&checker.TypeFlagsUnionOrIntersection != 0 {
		for _, part := range subjectType.Types() {
			record.parts = append(record.parts, graph.add(part))
		}
	}
	record.array = checker.Checker_isArrayType(graph.checker, subjectType)
	if checker.IsTupleType(subjectType) {
		record.tupleFlags = int(checker.TupleType_combinedFlags(subjectType.TargetTupleType()))
	}
	// Only type parameters need a constraint in the wire graph. Rest parameters
	// select their constrained root in Inspect before an index type is requested.
	if subjectType.Flags()&checker.TypeFlagsTypeParameter != 0 {
		record.constraint = graph.add(checker.Checker_getBaseConstraintOfType(graph.checker, subjectType))
	}
	graph.records[index] = record
	return id
}
func (graph *graph) write(out *fields) {
	out.number(uint64(len(graph.records)))
	for _, record := range graph.records {
		out.number(record.id)
		out.number(record.flags)
		out.text(record.name)
		out.yes(record.errorType)
		out.number(record.target)
		out.yes(record.array)
		out.text(strconv.Itoa(record.tupleFlags))
		out.number(record.constraint)
		out.number(record.element)
		out.ids(record.parts)
		out.ids(record.arguments)
	}
}
func constrained(c *checker.Checker, subjectType *checker.Type) *checker.Type {
	if bound := checker.Checker_getBaseConstraintOfType(c, subjectType); bound != nil {
		return bound
	}
	return subjectType
}

// Inspect's supported questions are documented in facts.md. Exact nodes and
// opaque per-program type identities are facts, never addresses or owned handles.
func (program *Program) Inspect(file string, start, end uint64, kind, question string) (string, error) {
	source, node, exactError := program.exact(file, start, end, kind)
	if exactError != nil {
		return "", exactError
	}
	c, release := program.Compiler.GetTypeCheckerForFile(context.Background(), source)
	defer release()
	out := &fields{}
	out.number(1)
	mode := strings.Split(question, "\n")[0]
	out.text(mode)
	switch mode {
	case "symbol-origin":
		if question != mode {
			return "", fmt.Errorf("unexpected symbol-origin suffix")
		}
		symbol := c.GetSymbolAtLocation(node)
		file := ""
		if symbol != nil && symbol.ValueDeclaration != nil {
			if sourceFile := ast.GetSourceFileOfNode(symbol.ValueDeclaration); sourceFile != nil {
				file = sourceFile.FileName().AsString()
			}
		}
		out.text(file)
	case "type-origin", "property-info", "call-count":
		split := strings.SplitN(question, "\n", 3)
		if len(split) < 2 ||
			(mode == "property-info" && len(split) != 3) ||
			(mode != "property-info" && len(split) != 2) {
			return "", fmt.Errorf("invalid metadata question")
		}
		id, parseError := strconv.ParseUint(split[1], 10, 64)
		if parseError != nil ||
			id == 0 ||
			id > uint64(len(program.typesByID)) ||
			strconv.FormatUint(id, 10) != split[1] {
			return "", fmt.Errorf("unknown checker type identity")
		}
		subject := program.typesByID[id-1]
		if mode == "call-count" {
			out.number(uint64(len(c.GetSignaturesOfType(subject, checker.SignatureKindCall))))
		} else if mode == "type-origin" {
			writeSymbolOrigin(out, program, subject.Symbol())
		} else {
			writePropertyInfo(out, checker.Checker_getPropertyOfType(c, subject, split[2]))
		}
	case "scope-locals":
		if question != mode || node.Kind != ast.KindSourceFile {
			return "", fmt.Errorf("scope-locals requires a SourceFile")
		}
		scopeTables(out, node)
	case "options", "strict-this":
		if question != mode || node.Kind != ast.KindSourceFile {
			return "", fmt.Errorf("%s requires a SourceFile", mode)
		}
		option := program.Compiler.Options().StrictNullChecks
		if mode == "strict-this" {
			option = program.Compiler.Options().NoImplicitThis
		}
		out.yes(program.Compiler.Options().GetStrictOptionValue(option))
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
		_, target, exactError := program.exact(file, first, last, split[3])
		if exactError != nil {
			return "", exactError
		}
		out.yes(checker.Checker_isTypeAssignableTo(c, c.GetTypeAtLocation(node), c.GetTypeAtLocation(target)))
	case "assignable-types":
		split := strings.Split(question, "\n")
		if len(split) != 3 {
			return "", fmt.Errorf("assignable-types requires two type identities")
		}
		selected := make([]*checker.Type, 2)
		for index := range selected {
			id, parseError := strconv.ParseUint(split[index+1], 10, 64)
			if parseError != nil ||
				id == 0 ||
				id > uint64(len(program.typesByID)) ||
				strconv.FormatUint(id, 10) != split[index+1] {
				return "", fmt.Errorf("unknown checker type identity")
			}
			selected[index] = program.typesByID[id-1]
		}
		out.yes(checker.Checker_isTypeAssignableTo(c, selected[0], selected[1]))
	case "enum-types":
		if question != mode {
			return "", fmt.Errorf("unexpected enum question suffix")
		}
		graph := &graph{program: program, checker: c}
		subject := c.GetTypeAtLocation(node)
		parts := []*checker.Type{subject}
		if subject.Flags()&checker.TypeFlagsUnion != 0 {
			parts = subject.Types()
		}
		var bases []uint64
		for _, part := range parts {
			if part.Flags()&checker.TypeFlagsEnumLiteral == 0 {
				continue
			}
			base := part
			if symbol := part.Symbol(); symbol != nil &&
				symbol.Flags&ast.SymbolFlagsEnumMember != 0 &&
				symbol.ValueDeclaration != nil &&
				symbol.ValueDeclaration.Parent != nil {
				base = c.GetTypeAtLocation(symbol.ValueDeclaration.Parent)
			}
			bases = append(bases, graph.id(base))
		}
		out.ids(bases)
	case "name", "type-symbol":
		split := strings.Split(question, "\n")
		if len(split) != 2 {
			return "", fmt.Errorf("name requires a type identity")
		}
		id, parseError := strconv.ParseUint(split[1], 10, 64)
		if parseError != nil ||
			id == 0 ||
			id > uint64(len(program.typesByID)) ||
			strconv.FormatUint(id, 10) != split[1] {
			return "", fmt.Errorf("unknown checker type identity")
		}
		if mode == "type-symbol" {
			name := ""
			if symbol := program.typesByID[id-1].Symbol(); symbol != nil {
				name = symbol.Name
			}
			out.text(name)
		} else {
			out.text(c.TypeToString(program.typesByID[id-1]))
		}
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
				file := ast.GetSourceFileOfNode(declaration)
				if file == nil {
					panic("declaration has no source")
				}
				out.text(file.FileName().AsString())
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
	case "call-parameters", "apparent-shape", "base-shapes", "call-returns", "property-shape", "contextual-shape", "widened-shape", "raw-type", "type", "base-type", "signature", "raw-shape", "type-shape", "signature-shape":
		if question != mode &&
			mode != "property-shape" &&
			mode != "call-parameters" &&
			mode != "apparent-shape" &&
			mode != "base-shapes" {
			return "", fmt.Errorf("unexpected type question suffix")
		}
		graph := &graph{program: program, checker: c, seen: make(map[*checker.Type]bool), names: !strings.HasSuffix(
			mode,
			"-shape",
		) && mode != "call-returns" && mode != "call-parameters" && mode != "base-shapes"}
		var roots []uint64
		var rest []bool
		present := true
		if mode == "call-parameters" || mode == "apparent-shape" || mode == "base-shapes" {
			split := strings.Split(question, "\n")
			if len(split) != 2 {
				return "", fmt.Errorf("type metadata requires an identity")
			}
			id, parseError := strconv.ParseUint(split[1], 10, 64)
			if parseError != nil ||
				id == 0 ||
				id > uint64(len(program.typesByID)) ||
				strconv.FormatUint(id, 10) != split[1] {
				return "", fmt.Errorf("unknown checker type identity")
			}
			subject := program.typesByID[id-1]
			if mode == "apparent-shape" {
				roots = append(roots, graph.add(checker.Checker_getApparentType(c, subject)))
			} else if mode == "base-shapes" {
				symbol := subject.Symbol()
				if symbol != nil && symbol.Flags&(ast.SymbolFlagsClass|ast.SymbolFlagsInterface) != 0 {
					for _, base := range checker.Checker_getBaseTypes(
						c,
						checker.Checker_getDeclaredTypeOfSymbol(c, symbol),
					) {
						roots = append(roots, graph.add(base))
					}
				}
			} else {
				for _, signature := range c.GetSignaturesOfType(subject, checker.SignatureKindCall) {
					parameters := checker.Signature_parameters(signature)
					if len(parameters) > 0 {
						roots = append(
							roots,
							graph.add(
								checker.Checker_getApparentType(c, c.GetTypeOfSymbolAtLocation(parameters[0], node)),
							),
						)
					}
				}
			}
		} else if mode == "call-returns" {
			for _, signature := range c.GetSignaturesOfType(c.GetTypeAtLocation(node), checker.SignatureKindCall) {
				roots = append(roots, graph.add(c.GetReturnTypeOfSignature(signature)))
			}
		} else if mode == "property-shape" {
			split := strings.SplitN(question, "\n", 3)
			if len(split) != 3 {
				return "", fmt.Errorf("property-shape requires type identity and property name")
			}
			id, parseError := strconv.ParseUint(split[1], 10, 64)
			if parseError != nil ||
				id == 0 ||
				id > uint64(len(program.typesByID)) ||
				strconv.FormatUint(id, 10) != split[1] {
				return "", fmt.Errorf("unknown checker type identity")
			}
			property := checker.Checker_getPropertyOfType(c, program.typesByID[id-1], split[2])
			present = property != nil
			if present {
				roots = append(roots, graph.add(c.GetTypeOfSymbolAtLocation(property, node)))
			}
		} else if mode == "contextual-shape" {
			subjectType := checker.Checker_getContextualType(c, node, checker.ContextFlagsNone)
			present = subjectType != nil
			if present {
				roots = append(roots, graph.add(subjectType))
			}
		} else if mode == "signature" || mode == "signature-shape" {
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
			roots = append(roots, graph.add(c.GetTypeAtLocation(callee)))
			signature := c.GetResolvedSignature(node)
			present = signature != nil
			if signature != nil {
				for _, parameter := range checker.Signature_parameters(signature) {
					subjectType := c.GetTypeOfSymbolAtLocation(parameter, node)
					if subjectType == nil {
						continue
					}
					isRest := len(parameter.Declarations) > 0 &&
						parameter.Declarations[0].Kind == ast.KindParameter &&
						parameter.Declarations[0].AsParameterDeclaration().DotDotDotToken != nil
					if isRest {
						subjectType = constrained(c, subjectType)
						root := graph.add(subjectType)
						// The number index is needed only for a signature's rest parameter.
						element := graph.add(
							checker.Checker_getIndexTypeOfType(c, subjectType, checker.Checker_numberType(c)),
						)
						for index := range graph.records {
							if graph.records[index].id == root {
								graph.records[index].element = element
							}
						}
						roots = append(roots, root)
					} else {
						roots = append(roots, graph.add(subjectType))
					}
					rest = append(rest, isRest)
					if isRest {
						break
					}
				}
			}
		} else {
			subjectType := c.GetTypeAtLocation(node)
			if mode != "raw-type" && mode != "raw-shape" && mode != "widened-shape" {
				subjectType = constrained(c, subjectType)
			}
			if mode == "widened-shape" {
				subjectType = checker.Checker_getWidenedType(c, subjectType)
			}
			if mode == "base-type" {
				subjectType = checker.Checker_getBaseTypeOfLiteralType(c, subjectType)
			}
			roots = append(roots, graph.add(subjectType))
		}
		out.yes(program.Compiler.Options().GetStrictOptionValue(program.Compiler.Options().StrictNullChecks))
		out.yes(present)
		out.ids(roots)
		out.number(uint64(len(rest)))
		for _, r := range rest {
			out.yes(r)
		}
		graph.write(out)
	default:
		return "", fmt.Errorf("unsupported checker question: %s", question)
	}
	return out.String(), nil
}
