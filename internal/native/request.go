package native

import (
	"fmt"
	"strings"

	"github.com/system-inc/adamic/internal/ir"
)

// WASI emits the request ABI around one validated function, or ordinary command C for handler -1.
func WASI(program *ir.Program, handler int) (string, error) {
	if handler == -1 {
		return C(program), nil
	}
	if handler < 0 || handler >= len(program.Functions) {
		return "", fmt.Errorf("native: invalid request handler index %d", handler)
	}
	function := program.Functions[handler]
	if function.Closure || function.Returns != ir.String || len(function.Parameters) != 1 || program.Locals[function.Parameters[0]].Type != ir.String {
		return "", fmt.Errorf("native: request handler must take one string and return a string")
	}
	return cProgram(program, handler), nil
}

func (e *emitter) requestABI(handler int) string {
	// Reuse consumes object and array parameters, never strings. The decoded string stays
	// owned here; the ordinary string parameter convention retains when reassignment needs it.
	var builder strings.Builder
	builder.WriteString(`
// The host calls WASI _initialize once, which runs runtime and module constructors.
__attribute__((constructor)) static void adamic_module_start(void) {
 (void)main(0, NULL);
 adamic_output_flush();
}

// Borrowed UTF-8 input; the returned string is owned by the host until adamic_release.
adamic_string *adamic_request(const unsigned char *bytes, size_t length) {
 adamic_string *request = adamic_decode_utf8(bytes, length);
`)
	fmt.Fprintf(&builder, " adamic_string *response = %s(%s);\n", e.functionName(handler), "request")
	builder.WriteString(" adamic_release(request);\n")
	if e.program.Functions[handler].MayThrow {
		builder.WriteString(" if (adamic_thrown != NULL) adamic_uncaught();\n")
	}
	builder.WriteString(` adamic_output_flush();
 return response;
}

const char *adamic_response_bytes(const adamic_string *response) {
 return response->bytes;
}

size_t adamic_response_length(const adamic_string *response) {
 return response->length;
}

#ifdef ADAMIC_COUNT
#include "count.h"
size_t adamic_live(void) { return adamic_counted.live; }
size_t adamic_regions(void) { return adamic_counted.regions; }
#endif
`)
	return builder.String()
}
