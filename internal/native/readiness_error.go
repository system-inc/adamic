package native

import "github.com/system-inc/adamic/internal/ir"

// A lexical dead-zone read uses the same owned exception payload as an explicit throw.
func (e *emitter) readinessError(message string) {
	e.line("static adamic_string error_message = ADAMIC_STRING(%s);", cString(message))
	thrown, _ := converted(ir.Object, ir.Union, "adamic_error_new_kind(&error_message, \"ReferenceError\")")
	e.line("adamic_thrown = %s;", thrown)
	e.line("adamic_exception_pending = true;")
	e.checkThrown()
}
