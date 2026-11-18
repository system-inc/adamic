package platform

import (
	"errors"
	"fmt"
	"testing"
)

func TestMissingCarriesPlatformAndRequirement(t *testing.T) {
	t.Parallel()
	original := &Missing{Platform: "darwin", Requirement: "Xcode SDK"}
	var missing *Missing
	if !errors.As(fmt.Errorf("load: %w", original), &missing) || missing.Platform != "darwin" || missing.Requirement != "Xcode SDK" || missing.Error() != "darwin platform missing: Xcode SDK" || MissingExitCode != 5 {
		t.Fatalf("lost typed platform stop: %+v", missing)
	}
}
