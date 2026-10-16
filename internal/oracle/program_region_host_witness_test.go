package oracle

import (
	"context"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"path/filepath"
	"strings"
	"testing"
)

func programRegionPending(t *testing.T, path, feature string) {
	t.Helper()
	source, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	_, err = lower.LowerWithOptions(context.Background(), source, lower.Options{ProgramRegion: true})
	if err == nil {
		t.Fatalf("pending feature %s is now available; activate %s", feature, path)
	}
	if !strings.Contains(err.Error(), feature) {
		t.Fatalf("pending %s failed elsewhere: %v", feature, err)
	}
	t.Skipf("registered pending %s: %v", feature, err)
}
func TestProgramRegionHost14AgreesWithNode(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs("../../stage3/fixtures/host/14_getCurrentDirectory.a")
	if err != nil {
		t.Fatal(err)
	}
	counts := checkProgramRegion(t, programRegionLowered(t, path, true), path)
	if counts[5] != 4 {
		t.Fatalf("host members %d, want 4", counts[5])
	}
}
