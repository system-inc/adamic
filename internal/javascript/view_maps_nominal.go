package javascript

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

func (e *emitter) mapEntryNominalCertificate(id ir.ViewContractID, value, where string) string {
	if id == 0 {
		return value
	}
	contract := e.program.ViewContracts[id-1]
	nullable, undefined := contract.Null, contract.Undefined
	if contract.Kind == ir.ViewNullable {
		if contract.Element == 0 {
			return value
		}
		contract = e.program.ViewContracts[contract.Element-1]
	}
	if contract.NominalClass == 0 {
		return value
	}
	accepted := []string{fmt.Sprintf("adamicInstanceOf(entry, %d)", contract.NominalClass)}
	if nullable {
		accepted = append(accepted, "entry === null")
	}
	if undefined || contract.Undefined {
		accepted = append(accepted, "entry === undefined")
	}
	message := quote("Map nominal producer failed: " + where + " expected " + contract.Name + ", found value without its class identity")
	return "((entry) => (" + strings.Join(accepted, " || ") + ") ? entry : panic(" + message + "))(" + value + ")"
}
