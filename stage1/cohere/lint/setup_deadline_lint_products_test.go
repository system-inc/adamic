package lint

import (
	"testing"
)

// Build-phase units call the same recipes and keys as independently selected shards.
// Full-port lowering and its dependent native products remain shard preparation
// until their cold builds meet the build phase grain (#c5k975w, #3he8f8g).
func TestProduct_EmittedMismatchOracle(t *testing.T) {
	t.Parallel()
	emittedMismatchOracle(t)
}

// Independently selectable shared preparation check; it imposes no setup deadline.
func TestRulesAgree_Setup(t *testing.T) {
	t.Parallel()
	rulesAgreeSetup(t)
}
