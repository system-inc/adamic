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

func TestTSGoRequiresLink(t *testing.T) {
	fixture := filepath.Join("testdata", "queries.a")
	loaded, err := load.Load([]string{fixture})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := lower.Lower(context.Background(), loaded); err == nil {
		t.Fatal("lowering accepted an unlinked checker call")
	}
	loaded.EnableTSGo()
	lowered, err := lower.Lower(context.Background(), loaded)
	if err != nil {
		t.Fatal(err)
	}
	if !native.UsesTSGo(lowered) {
		t.Fatal("opted-in bridge calls lost")
	}
}
