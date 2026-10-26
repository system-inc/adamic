package native

import (
	"fmt"
	"strings"

	"github.com/system-inc/adamic/internal/ir"
)

// An own entry wins. Only a missing prototype key needs the non-string boundary.
// Keep this check here, separate from the runtime's single environment primitive.
func (e *emitter) nodeEnvironmentRead(argument string) string {
	name := e.snapshot(ir.String, argument)
	result := e.snapshot(ir.String, fmt.Sprintf("adamic_host_environment(%s)", name))
	var keys []string
	for _, key := range ir.NodeEnvironmentPrototypeKeys {
		keys = append(keys, fmt.Sprintf("adamic_string_equal(%s, &(adamic_string)ADAMIC_STRING(%s))", name, cString(key)))
	}
	e.line("if (%s == NULL && (%s)) {", result, strings.Join(keys, " || "))
	prefix, message := e.temporary(), e.temporary()
	e.line("static adamic_string %s = ADAMIC_STRING(%s);", prefix, cString(ir.NodeEnvironmentInheritedPanic))
	e.line("adamic_string *%s = adamic_string_concat(2, (adamic_string *const[]){&%s, %s});", message, prefix, name)
	e.line("adamic_panic(%s->bytes, %s->length);", message, message)
	e.line("}")
	return result
}
