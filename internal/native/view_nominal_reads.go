package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

// A producer certificate never substitutes for checking an escaped nominal read.
func (e *emitter) nominalViewRead(id ir.ViewContractID, value, where string, optional bool) {
	if id == 0 {
		return
	}
	c, nullable, undefined := mapNominalContract(e.program, id)
	if c.NominalClass == 0 {
		return
	}
	accepted := []string{fmt.Sprintf("adamic_instanceof(%s, &adamic_class_%d)", value, c.NominalClass)}
	if nullable {
		accepted = append(accepted, "(const void *)"+value+" == &adamic_null")
	}
	if undefined || optional {
		accepted = append(accepted, value+" == NULL")
	}
	if optional {
		accepted = append(accepted, "(const void *)"+value+" == &adamic_null")
	}
	message := cString("cast failed: field read failed: " + where + "; expected " + e.program.ViewContracts[id-1].Name + " class identity, found incompatible object")
	e.line("if (!(%s)) adamic_panic(%s, sizeof %s - 1);", strings.Join(accepted, " || "), message, message)
}

func (e *emitter) nominalViewReceiver(property ir.Property) string {
	receiver := e.value(property.Object)
	e.nominalViewRead(e.program.NominalReadContracts[property.ViewReceiverTypeID], receiver, property.View+" receiver", property.Optional)
	return receiver
}
