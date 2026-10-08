package javascript

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

func (e *emitter) nominalViewRead(id ir.ViewContractID, value, where string, optional bool) string {
	if id == 0 {
		return value
	}
	c := e.program.ViewContracts[id-1]
	nullable, undefined := c.Null, c.Undefined
	name := c.Name
	if c.Kind == ir.ViewNullable {
		if c.Element == 0 {
			return value
		}
		c = e.program.ViewContracts[c.Element-1]
	}
	if c.NominalClass == 0 {
		return value
	}
	accepted := []string{fmt.Sprintf("adamicInstanceOf(value, %d)", c.NominalClass)}
	if nullable || optional {
		accepted = append(accepted, "value === null")
	}
	if undefined || optional {
		accepted = append(accepted, "value === undefined")
	}
	message := quote("cast failed: field read failed: " + where + "; expected " + name + " class identity, found incompatible object")
	return "((value) => (" + strings.Join(accepted, " || ") + ") ? value : panic(" + message + "))(" + value + ")"
}

func (e *emitter) nominalViewReceiver(property ir.Property) string {
	return e.nominalViewRead(e.program.NominalReadContracts[property.ViewReceiverTypeID], e.value(property.Object), property.View+" receiver", property.Optional)
}

func (e *emitter) nominalArrayContract(id ir.ViewContractID) bool {
	if id == 0 {
		return false
	}
	c := e.program.ViewContracts[id-1]
	if c.Kind == ir.ViewNullable && c.Element != 0 {
		c = e.program.ViewContracts[c.Element-1]
	}
	return c.NominalClass != 0
}
