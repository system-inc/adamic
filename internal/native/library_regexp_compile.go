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

// RuntimeLibraryForSource selects the same opt-in archive for callers that link
// generated C themselves, including the test262 runner using another checkout.
func RuntimeLibraryForSource(directory, source string, options Options) (string, error) {
	flags := []string{}
	if strings.Contains(source, "\n#define ADAMIC_REGEXP_RUNTIME_COMPILER 1\n") {
		flags = append(flags, "-DADAMIC_REGEXP_RUNTIME_OWNER=1")
	}
	if strings.Contains(source, "#define ADAMIC_REGEXP_REPLACE_CALLBACK 1\n") {
		flags = append(flags, "-DADAMIC_REGEXP_REPLACE_CALLBACK=1")
	}
	if len(flags) != 0 {
		return runtimeLibrary(directory, options, flags)
	}
	return RuntimeLibrary(directory, options)
}
