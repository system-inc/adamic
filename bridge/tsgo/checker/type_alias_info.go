package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"strconv"
	"strings"
)

// Alias identity, declaration origins and arguments are compiler facts only.
func (p *Program) typeAliasInfo(out *fields, c *checker.Checker, question string) error {
	parts := strings.Split(question, "\n")
	if len(parts) != 2 {
		return fmt.Errorf("type-alias-info requires a type identity")
	}
	id, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil || id == 0 || id > uint64(len(p.typesByID)) || strconv.FormatUint(id, 10) != parts[1] {
		return fmt.Errorf("unknown checker type identity")
	}
	alias := checker.Type_alias(p.typesByID[id-1])
	g := &graph{program: p, checker: c, seen: make(map[*checker.Type]bool)}
	var roots []uint64
	if alias == nil {
		writeSymbolOrigin(out, p, nil)
	} else {
		writeSymbolOrigin(out, p, alias.Symbol())
		for _, argument := range alias.TypeArguments() {
			roots = append(roots, g.add(argument))
		}
	}
	out.yes(p.Compiler.Options().GetStrictOptionValue(p.Compiler.Options().StrictNullChecks))
	out.yes(alias != nil)
	out.ids(roots)
	out.number(0)
	g.write(out)
	return nil
}

// The existing dispatcher delegates its otherwise unsupported question here.
func (p *Program) inspectTypeAliasInfo(out *fields, c *checker.Checker, node *ast.Node, question string) (string, error) {
	if strings.Split(question, "\n")[0] == "then-callback-parameters" {
		return p.thenCallbackParameters(out, c, node, question)
	}
	if strings.Split(question, "\n")[0] != "type-alias-info" {
		return p.inspectContinuation(out, c, node, question)
	}
	if err := p.typeAliasInfo(out, c, question); err != nil {
		return "", err
	}
	return out.String(), nil
}
