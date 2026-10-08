package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
)

func (e *emitter) jsonParse(p ir.JSONParse) string {
	e.declarations = append(e.declarations, `#include "json_parse.h"`)
	text := e.value(p.Text)
	pinned := e.temporary()
	e.line("adamic_string *%s = %s;", pinned, text)
	callback := "NULL"
	if p.Reviver != nil {
		callback = e.value(p.Reviver)
	}
	if p.IgnoredReviver != nil {
		e.value(p.IgnoredReviver)
	}
	result := e.temporary()
	e.line("adamic_value %s = adamic_json_parse(%s, %s, adamic_json_%s, %d, adamic_json_parse_%s);", result, pinned, callback, p.ReviverKind, p.Takes, p.Mode)
	// On failure the runtime returns no reference. Check before making the result an owned
	// temporary so exception cleanup never releases an uninitialized or transferred slot.
	e.checkThrown()
	value := fmt.Sprintf("%s.%s", result, member(p.Of))
	if p.Of.IsReference() {
		return e.own(p.Of, value)
	}
	return e.snapshot(p.Of, value)
}
