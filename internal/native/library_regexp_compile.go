package native

import (
	"fmt"
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
	return runtimeLibrary(directory, options, regExpRuntimeFlags(source))
}

// The compiler lives in generated C, while its counted storage hooks live in
// regexp.c. Every builder must select the same ownership variant for that source.
func regExpRuntimeFlags(source string) []string {
	if strings.Contains(source, "\n#define ADAMIC_REGEXP_RUNTIME_COMPILER 1\n") {
		return []string{"-DADAMIC_REGEXP_RUNTIME_OWNER=1"}
	}
	return nil
}

// The single-unit backend includes the compiler implementation in generated C.
// Split builds compile it once beside the generated units, which share only its ABI.
const regExpCompilerIncludes = `#include "regexp_compile_parser.c"
#include "regexp_compile_properties.c"
#include "regexp_compile_sets.c"
#include "regexp_compile_bytecode.c"
#include "regexp_compile_v8.c"
#include "regexp_compile_runtime.c"
`

func splitRegExpCompiler(source string) (string, *compilationUnit, error) {
	if len(regExpRuntimeFlags(source)) == 0 {
		return source, nil, nil
	}
	if strings.Count(source, regExpCompilerIncludes) != 1 {
		return "", nil, fmt.Errorf("native: split: missing unique RegExp compiler includes")
	}
	unit := &compilationUnit{name: "regexp_compiler.c", source: "#include \"adamic.h\"\n#define ADAMIC_REGEXP_RUNTIME_COMPILER 1\n" + regExpCompilerIncludes}
	return strings.Replace(source, regExpCompilerIncludes, "#include \"regexp_compile_runtime.h\"\n", 1), unit, nil
}

func regExpCompilerFile(name string) bool {
	return strings.Contains(regExpCompilerIncludes, "#include \""+name+"\"\n")
}
