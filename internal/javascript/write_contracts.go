package javascript

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

func (e *emitter) writeContract(contract *ir.FieldContract) string {
	integers := func(values []int) string {
		parts := []string{}
		for _, value := range values {
			parts = append(parts, fmt.Sprint(value))
		}
		return strings.Join(parts, ", ")
	}
	fields := []string{}
	for _, field := range contract.Fields {
		fields = append(fields, fmt.Sprintf("{name: %s, optional: %t, contract: %s}", quote(field.Name), field.Optional, e.writeContract(field.Contract)))
	}
	return fmt.Sprintf("{kind: %d, declared: %s, nullable: %t, allowed: [%s], typeID: %d, reference: %t, structural: %t, provenWrites: [%s], provenFields: [%s], fields: [%s]}", contract.Kind, quote(contract.Declared), contract.Nullable, e.values(contract.Allowed), contract.TypeID, contract.Reference, contract.Structural, integers(contract.ProvenWrites), integers(contract.ProvenFields), strings.Join(fields, ", "))
}
