package checker

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// Raw nonnullable identity, union members, symbol origins and declared bases.
func (p *Program) interfaceBases(out *fields, c *checker.Checker, question string) (string, error) {
	split := strings.Split(question, "\n")
	if len(split) != 2 {
		return "", fmt.Errorf("interface-bases requires an identity")
	}
	id, err := strconv.ParseUint(split[1], 10, 64)
	if err != nil || id == 0 || id > uint64(len(p.typesByID)) || strconv.FormatUint(id, 10) != split[1] {
		return "", fmt.Errorf("unknown checker type identity")
	}
	t := p.typesByID[id-1]
	g := &graph{program: p, checker: c}
	out.number(uint64(t.Flags()))
	out.number(g.id(c.GetNonNullableType(t)))
	s := t.Symbol()
	flags := uint64(0)
	name := ""
	count := 0
	if s != nil {
		flags = uint64(s.Flags)
		name = s.Name
		count = len(s.Declarations)
	}
	out.number(flags)
	out.text(strings.ToValidUTF8(name, "�"))
	out.number(uint64(count))
	if s != nil {
		for _, d := range s.Declarations {
			f := ast.GetSourceFileOfNode(d)
			out.yes(f != nil && p.Compiler.IsSourceFileDefaultLibrary(f.PathKey()))
		}
	}
	var parts, bases []uint64
	if t.Flags()&checker.TypeFlagsUnion != 0 {
		for _, part := range t.Types() {
			parts = append(parts, g.id(part))
		}
	}
	if s != nil {
		declared := checker.Checker_getDeclaredTypeOfSymbol(c, s)
		if declared != nil && declared.ObjectFlags()&(checker.ObjectFlagsInterface|checker.ObjectFlagsClass) != 0 {
			for _, base := range checker.Checker_getBaseTypes(c, declared) {
				bases = append(bases, g.id(base))
			}
		}
	}
	out.ids(parts)
	out.ids(bases)
	return out.String(), nil
}
