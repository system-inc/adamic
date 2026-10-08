package javascript

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
)

// Presence and kind were already checked by the owner. Preserve permitted
// null/undefined and independently certify every present callable snapshot.
func (e *emitter) viewCallableNullishCertificate(property ir.Property, value string) string {
	if property.ViewContract == 0 {
		return value
	}
	contract := e.program.ViewContracts[property.ViewContract-1]
	if contract.Kind != ir.ViewCallable || contract.Result == 0 && !contract.DiscardResult {
		return value
	}
	return fmt.Sprintf("((v) => v == null ? v : %s)(%s)", e.emitViewCallableCertificate(property, "v"), value)
}
