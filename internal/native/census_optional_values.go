package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
)

// A call's declared type can omit more parameters than the actual closure. The callee
// supplies its own undefined representation, before the existing default prologue runs.
func closureArgument(of ir.Type, index int) string {
	value := unslotted(of, fmt.Sprintf("arguments[%d].%s", index, member(of)))
	missing := zero(of)
	if of.IsReference() {
		missing = "NULL"
	}
	return fmt.Sprintf("(argument_count > %d ? %s : %s)", index, value, missing)
}
