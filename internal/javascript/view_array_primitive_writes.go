package javascript

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

// Reify only complete declared primitive domains, not physical pointer tags.
func primitiveArrayWriteRuntime(program *ir.Program) string {
	entries := []string{}
	for index, contract := range program.ViewContracts {
		if contract.Of != ir.Union {
			continue
		}
		members, complete := ir.PrimitiveViewMembers(program, ir.ViewContractID(index+1))
		if !complete {
			continue
		}
		tests := []string{}
		for _, member := range members {
			if member.Kind == ir.ViewUndefined {
				tests = append(tests, "v === undefined")
				continue
			}
			if len(member.Allowed) == 0 {
				tests = append(tests, "typeof v === "+quote(map[ir.Type]string{ir.Number: "number", ir.Boolean: "boolean", ir.String: "string"}[member.Of]))
				continue
			}
			for _, literal := range member.Allowed {
				switch literal.Of {
				case ir.Number:
					tests = append(tests, fmt.Sprintf("v === %g", literal.Number))
				case ir.String:
					tests = append(tests, "v === "+quote(literal.String))
				case ir.Boolean:
					tests = append(tests, fmt.Sprintf("v === %t", literal.Boolean))
				}
			}
		}
		entries = append(entries, fmt.Sprintf("%d: v => (%s)", index+1, strings.Join(tests, " || ")))
	}
	return "const adamicArrayPrimitiveWrites = {" + strings.Join(entries, ",") + "};\n"
}
