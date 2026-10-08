package native

import (
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

func (e *emitter) dynamicRegExp(expression ir.RegExpNew) string {
	arguments := make([]string, 0, len(expression.Arguments))
	for _, argument := range expression.Arguments {
		value := e.value(argument)
		arguments = append(arguments, e.own(argument.Type(), "adamic_retain("+value+")"))
	}
	result := e.own(ir.Object, "adamic_regex_compile_new("+strings.Join(arguments, ", ")+")")
	e.checkThrown()
	return result
}

// Ownership hooks are an archive variant, separate from compiler definitions in
// generated source. Its flag participates in the existing runtime cache key.
func runtimeLibraryForSource(source string, options Options) (string, error) {
	return RuntimeLibraryForSource("", source, options)
}
