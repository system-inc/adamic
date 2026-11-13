package toolchain

import (
	"errors"
	"fmt"
	"testing"
)

func TestPlatformMissingCarriesPlatformAndRequirement(t *testing.T) {
	t.Parallel()
	original := &PlatformMissing{Platform: "darwin", Missing: "Xcode SDK"}
	var missing *PlatformMissing
	if !errors.As(fmt.Errorf("load: %w", original), &missing) || missing.Platform != "darwin" || missing.Missing != "Xcode SDK" || missing.Error() != "darwin platform missing: Xcode SDK" || PlatformMissingExitCode != 5 {
		t.Fatalf("lost typed platform stop: %+v", missing)
	}
}
