package typescript

import (
	"strings"
	"testing"
)

// TestOptionDecodersRefuseAnUnknownKey replaces three table rows that asserted `{"somethingElse": 1}`
// decodes to the defaults. Upstream's schema for each of these rules sets `additionalProperties:
// false` (checked in the installed typescript-eslint), so ESLint refuses the key, and a config that
// loads clean while an option does nothing is the silent drop #4n972g9 closed.
func TestOptionDecodersRefuseAnUnknownKey(t *testing.T) {
	t.Parallel()
	decoders := map[string]func([]byte) (any, error){
		"no-meaningless-void-operator": DecodeNoMeaninglessVoidOperatorOptions,
		"require-array-sort-compare":   DecodeRequireArraySortCompareOptions,
		"strict-void-return":           DecodeStrictVoidReturnOptions,
	}
	for name, decode := range decoders {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if _, err := decode([]byte(`{}`)); err != nil {
				t.Fatalf("baseline: an empty options object was refused: %v", err)
			}
			_, err := decode([]byte(`{"somethingElse": 1}`))
			if err == nil || !strings.Contains(err.Error(), "somethingElse") {
				t.Fatalf("want a refusal naming the key, got %v", err)
			}
		})
	}
}
