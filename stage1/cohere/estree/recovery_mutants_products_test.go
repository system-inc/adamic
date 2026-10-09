package estree

import "testing"

// The build phase uses the shards' exact Product recipe and key. Each unit can
// run alone; the shards retain their own once-per-process preparation as well.
func TestProduct_RecoveryMutantFirstAccessibility(t *testing.T) {
	t.Parallel()
	recoveryMutantPrepared(t, "first-accessibility")
}

func TestProduct_RecoveryMutantEmptyTypeListRange(t *testing.T) {
	t.Parallel()
	recoveryMutantPrepared(t, "empty-type-list-range")
}

func TestProduct_RecoveryMutantModuleAwait(t *testing.T) {
	t.Parallel()
	recoveryMutantPrepared(t, "module-await")
}
