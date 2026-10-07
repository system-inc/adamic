package checker

import (
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"strconv"
	"strings"
)

func wave24UnionMembers(t *checker.Type) []*checker.Type {
	if t == nil {
		return nil
	}
	if t.IsUnion() {
		return t.Types()
	}
	return []*checker.Type{t}
}

// thenCallbackSignatures exposes call-signature counts on first callback
// parameters of then properties. The native side decides thenability.
func (p *Program) thenCallbackSignatures(out *fields, c *checker.Checker, node *ast.Node, question string) error {
	parts := strings.Split(question, "\n")
	if len(parts) != 2 {
		return fmt.Errorf("then-callback-signatures requires a type identity")
	}
	id, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil || id == 0 || id > uint64(len(p.typesByID)) || strconv.FormatUint(id, 10) != parts[1] {
		return fmt.Errorf("unknown checker type identity")
	}
	var counts []uint64
	for _, t := range wave24UnionMembers(checker.Checker_getApparentType(c, p.typesByID[id-1])) {
		property := checker.Checker_getPropertyOfType(c, t, "then")
		if property == nil {
			continue
		}
		for _, method := range wave24UnionMembers(c.GetTypeOfSymbolAtLocation(property, node)) {
			for _, signature := range c.GetSignaturesOfType(method, checker.SignatureKindCall) {
				parameters := checker.Signature_parameters(signature)
				if len(parameters) == 0 {
					counts = append(counts, 0)
					continue
				}
				parameter := parameters[0]
				callback := checker.Checker_getApparentType(c, c.GetTypeOfSymbolAtLocation(parameter, node))
				if d := parameter.ValueDeclaration; d != nil && d.Kind == ast.KindParameter && d.AsParameterDeclaration().DotDotDotToken != nil {
					callback = checker.Checker_getIndexTypeOfType(c, callback, checker.Checker_numberType(c))
				}
				count := uint64(0)
				for _, part := range wave24UnionMembers(callback) {
					count += uint64(len(c.GetSignaturesOfType(part, checker.SignatureKindCall)))
				}
				counts = append(counts, count)
			}
		}
	}
	out.ids(counts)
	return nil
}
