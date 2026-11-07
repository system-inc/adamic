package tsgo_test

import (
	"path/filepath"
	"testing"

	"context"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

func firstDifference(left, right []byte) int {
	for index := 0; index < len(left) && index < len(right); index++ {
		if left[index] != right[index] {
			return index
		}
	}
	return min(len(left), len(right))
}

func TestTSGoRequiresLink(test *testing.T) {
	test.Parallel()
	fixture := filepath.Join("testdata", "queries.a")
	loaded, loadError := load.Load([]string{fixture})
	if loadError != nil {
		test.Fatal(loadError)
	}
	if _, lowerError := lower.Lower(context.Background(), loaded); lowerError == nil {
		test.Fatal("lowering accepted an unlinked checker call")
	}
	loaded.EnableTSGo()
	lowered, lowerError := lower.Lower(context.Background(), loaded)
	if lowerError != nil {
		test.Fatal(lowerError)
	}
	if !native.UsesTSGo(lowered) {
		test.Fatal("opted-in bridge calls lost")
	}
}
