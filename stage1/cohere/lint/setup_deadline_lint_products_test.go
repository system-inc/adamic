package lint

import (
	"testing"
)

// Independently selectable shared preparation check; it imposes no setup deadline.
func TestRulesAgree_Setup(t *testing.T) {
	t.Parallel()
	rulesAgreeSetup(t)
}
