package oracle

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// Historical ordinary assertions are refusal controls under the .a ruling.
// Declaration assertions are distinct syntax and keep their readiness checks.
func refusedAdamicNonNullFixture(path string) bool {
	name := filepath.Base(path)
	if name == "09_checker_constituent_recursion.a" && strings.HasSuffix(filepath.ToSlash(path), "stage3/fixtures/nested-functions/"+name) {
		return true
	}
	if filepath.Ext(path) != ".a" || !strings.HasPrefix(name, "non_null") {
		return false
	}
	switch name {
	case "non_null_definite.a", "non_null_definite_local.a", "non_null_definite_field.a":
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
