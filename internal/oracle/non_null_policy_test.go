package oracle

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// Only deliberate .a refusal pins are excluded from runnable fixtures.
func refusedAdamicNonNullFixture(path string) bool {
	name := filepath.Base(path)
	if filepath.Ext(path) != ".a" || !(strings.HasPrefix(name, "non_null_refuse_") || strings.HasPrefix(name, "non_null_possible_")) {
		return false
	}
	return true
}

func assertAdamicNonNullRefusal(t *testing.T, err error) {
	t.Helper()
	var refused *lower.Refused
	if !errors.As(err, &refused) || refused.What != "the non-null assertion !" || refused.Fix != "write ?? panic('why it can't be missing'), or narrow and handle the missing case" {
		t.Fatalf("want .a non-null refusal with its unchanged fix, got %v", err)
	}
}
