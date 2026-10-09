package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
)

func (e *emitter) recordLiteral(literal ir.ObjectLiteral) string {
	constructor := "adamic_record_new(true)"
	if literal.Namespace {
		e.declarations = append(e.declarations, `#include "namespace.h"`)
		constructor = "adamic_namespace_new()"
	}
	object := e.own(ir.Object, constructor)
	for _, field := range literal.Fields {
		value := e.value(field.Value)
		key := e.recordKey(field.Name)
		e.line("adamic_record_define(%s, %s, (adamic_value){.reference = %s});", object, key, e.kept(value))
	}
	return object
}

func (e *emitter) recordKey(name string) string {
	key := e.temporary()
	e.declarations = append(e.declarations, fmt.Sprintf("static adamic_string %s = ADAMIC_STRING(%s);", key, cString(name)))
	return "&" + key
}

func (e *emitter) recordWrite(statement ir.SetProperty) {
	object := e.value(statement.Object)
	value := e.value(statement.Value)
	e.line("if (%s == NULL) adamic_panic(\"record write on undefined\", sizeof \"record write on undefined\" - 1);", object)
	e.line("adamic_object_check_data_write(%s, %s);", object, cString(statement.Name))
	if statement.NamespaceInstall {
		e.declarations = append(e.declarations, `#include "namespace.h"`)
		e.line("adamic_namespace_install(%s, %s, (adamic_value){.reference = %s}, %t);", object, e.recordKey(statement.Name), e.kept(value), statement.NamespaceReadonly)
	} else {
		e.line("adamic_record_set(%s, %s, (adamic_value){.reference = %s});", object, e.recordKey(statement.Name), e.kept(value))
	}
}

func (e *emitter) hasRecordStorage() bool {
	found := false
	walkExpressions(e.program, func(expression ir.Expression) {
		if literal, ok := expression.(ir.ObjectLiteral); ok && literal.Record {
			found = true
		}
	})
	return found
}
