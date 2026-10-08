package lower

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

const recordLimit = "a record other than a readonly string index signature holding finite readonly records, nonnullable scalars or arrays of nonnullable scalars"

// Finite readonly record edges cannot point back through a writable slot.
// Cyclic and expanding generic payloads remain outside the cycle proof.
func (l *lowering) recordInfo(t *checker.Type) (*checker.IndexInfo, bool) {
	return l.recordInfoSeen(t, map[*checker.Type]bool{})
}

func (l *lowering) recordInfoSeen(t *checker.Type, path map[*checker.Type]bool) (*checker.IndexInfo, bool) {
	if t == nil {
		return nil, false
	}
	t = l.concrete(t)
	// The depth bound also terminates generics whose payload instantiation grows
	// on every edge, without revisiting an identical checker type.
	if t.Flags()&checker.TypeFlagsObject == 0 || path[t] || len(path) >= 64 {
		return nil, false
	}
	infos := l.checker.GetIndexInfosOfType(t)
	if len(infos) != 1 || !infos[0].IsReadonly() || infos[0].KeyType().Flags() != checker.TypeFlagsString || len(l.checker.GetPropertiesOfType(t)) != 0 || len(l.checker.GetSignaturesOfType(t, checker.SignatureKindCall)) != 0 || len(l.checker.GetSignaturesOfType(t, checker.SignatureKindConstruct)) != 0 {
		return nil, false
	}
	path[t] = true
	defer delete(path, t)
	value := l.concrete(infos[0].ValueType())
	scalar := func(t *checker.Type) bool {
		return t.Flags()&(checker.TypeFlagsNumberLike|checker.TypeFlagsBooleanLike|checker.TypeFlagsStringLike) != 0
	}
	if scalar(value) {
		return infos[0], true
	}
	if l.checker.IsArrayType(value) {
		args := l.typeArguments(value)
		if len(args) == 1 && scalar(l.concrete(args[0])) {
			return infos[0], true
		}
	}
	if _, ok := l.recordInfoSeen(value, path); ok {
		return infos[0], true
	}
	return nil, false
}

func (l *lowering) sameRecordView(from, to *checker.Type) bool {
	fi, ti := l.checker.GetIndexInfosOfType(from), l.checker.GetIndexInfosOfType(to)
	if len(fi) == 0 && len(ti) == 0 {
		return true
	}
	// Arrays and library collections have their own representation/view checks.
	fr, _ := l.representation(from)
	tr, _ := l.representation(to)
	if fr != ir.Record && tr != ir.Record {
		return true
	}
	f, fok := l.recordInfo(from)
	t, tok := l.recordInfo(to)
	if !fok || !tok {
		return false
	}
	fv, _ := l.representation(f.ValueType())
	tv, _ := l.representation(t.ValueType())
	return fv == tv && l.sameKeeping(f.ValueType(), t.ValueType(), map[[2]*checker.Type]bool{}) && l.widened(f.ValueType(), t.ValueType(), map[[2]*checker.Type]bool{}) == nil
}

func (l *lowering) recordUse(node *ast.Node) error {
	node = ast.SkipParentheses(node)
	if node.Kind == ast.KindAsExpression {
		cast := node.AsAsExpression()
		from, to := l.checker.GetTypeAtLocation(cast.Expression), l.checker.GetTypeAtLocation(node)
		if (l.hasStringRecordIndex(from) || l.hasStringRecordIndex(to)) && ast.SkipParentheses(cast.Expression).Kind != ast.KindObjectLiteralExpression && !l.sameKeeping(from, to, map[[2]*checker.Type]bool{}) {
			return l.notYet(node, "a cast between incompatible record and fixed-object views or mutable payload types")
		}
	}
	if node.Kind == ast.KindPropertyAccessExpression {
		receiver := node.AsPropertyAccessExpression().Expression
		if rep, _ := l.representation(l.checker.GetTypeAtLocation(receiver)); rep == ir.Record {
			return l.notYet(node, "a named property or prototype member on a record; read an own string key with brackets")
		}
	}
	return nil
}

func (l *lowering) recordLiteral(node *ast.Node) (ir.Expression, bool, error) {
	t := l.checker.GetContextualType(node, checker.ContextFlagsNone)
	if t == nil || len(l.checker.GetIndexInfosOfType(t)) == 0 {
		return nil, false, nil
	}
	info, ok := l.recordInfo(t)
	if !ok {
		return nil, true, l.notYet(node, recordLimit)
	}
	valueType, _ := l.representation(info.ValueType())
	result := ir.MapNew{Record: true, Key: ir.String, Value: valueType}
	for _, property := range node.AsObjectLiteralExpression().Properties.Nodes {
		if property.Kind != ast.KindPropertyAssignment && property.Kind != ast.KindShorthandPropertyAssignment {
			return nil, true, l.notYet(property, "a record literal with spread, methods or accessors")
		}
		name := property.Name()
		if !ast.IsIdentifier(name) && name.Kind != ast.KindStringLiteral {
			return nil, true, l.notYet(name, "a computed or numeric record literal key")
		}
		if name.Text() == "__proto__" && property.Kind == ast.KindPropertyAssignment {
			return nil, true, l.notYet(property, "__proto__ in a record literal changes JavaScript's prototype")
		}
		var value ir.Expression
		var err error
		if property.Kind == ast.KindPropertyAssignment {
			value, err = l.expression(property.AsPropertyAssignment().Initializer)
		} else {
			value, err = l.shorthand(property)
		}
		if err != nil {
			return nil, true, err
		}
		result.Entries = append(result.Entries, [2]ir.Expression{ir.StringConstant{Index: l.constant(name.Text())}, fit(value, valueType)})
	}
	return result, true, nil
}

func (l *lowering) recordIndex(node *ast.Node) (ir.Expression, bool, error) {
	access := node.AsElementAccessExpression()
	t := l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(access.Expression))
	rep, _ := l.representation(t)
	if rep != ir.Record {
		return nil, false, nil
	}
	info, ok := l.recordInfo(t)
	if !ok {
		return nil, true, l.notYet(node, recordLimit)
	}
	if node.Flags&ast.NodeFlagsOptionalChain != 0 {
		return nil, true, l.notYet(node, "an optional record lookup")
	}
	object, err := l.expression(access.Expression)
	if err != nil {
		return nil, true, err
	}
	key, err := l.expression(access.ArgumentExpression)
	if err != nil {
		return nil, true, err
	}
	if key.Type() != ir.String {
		return nil, true, l.notYet(node, "a record lookup with a non-string key")
	}
	value, _ := l.representation(info.ValueType())
	return l.defined(node, ir.MapGet{Map: object, Key: key, KeyType: ir.String, ValueType: value}), true, nil
}

// Tuples and numeric enum namespaces also expose numeric index information;
// their fixed representations are handled by the existing lowering paths.
func (l *lowering) hasStringRecordIndex(t *checker.Type) bool {
	for _, info := range l.checker.GetIndexInfosOfType(t) {
		// RegExp named groups already have a checked, fixed-object representation.
		if declaration := info.Declaration(); declaration != nil {
			file := ast.GetSourceFileOfNode(declaration)
			if load.IsLibrary(file) && strings.Contains(file.AsSourceFile().FileName().AsString(), ".regexp.") {
				continue
			}
		}
		if info.KeyType().Flags()&checker.TypeFlagsStringLike != 0 {
			return true
		}
	}
	return false
}
