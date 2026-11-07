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
	if filepath.Ext(path) != ".a" ||
		!(strings.HasPrefix(name, "non_null_refuse_") || strings.HasPrefix(name, "non_null_possible_")) {
		return false
	}
	return true
}

func assertAdamicNonNullRefusal(test *testing.T, loweredError error) {
	test.Helper()
	var refused *lower.Refused
	if !errors.As(loweredError, &refused) ||
		refused.What != "the non-null assertion !" ||
		refused.Fix != "write ?? panic('why it can't be missing'), or narrow and handle the missing case" {
		test.Fatalf("want .a non-null refusal with its unchanged fix, got %v", loweredError)
	}
}
