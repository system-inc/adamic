package oracle

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/scout_string_number_fields_locals.a", true, false}, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/scout_string_number_stale_field.a", true, true})
}

func nullableObjectStorageOutcome(t *testing.T, name string) {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/scout_nullable_storage_refused", name+".a"))
	if err != nil {
		t.Fatal(err)
	}
	if difference := disagreement(run{stdout: []byte("true\n")}, onNode(t, path)); difference != "" {
		t.Fatal(difference)
	}
	_, err = lowered(t, path)
	if name == "array_spread" {
		var refusal *lower.Refused
		if !errors.As(err, &refusal) || refusal.What != "union-typed array built from a narrower literal" || refusal.Fix != "declare the literal's element type explicitly, or build the array with a mixed literal" {
			t.Fatalf("want the newer narrow-literal layout refusal with its fix, got %v", err)
		}
		t.Log(err)
		return
	}
	if err == nil || !(strings.Contains(err.Error(), "sharing nullable object storage") || name == "lookup" && strings.Contains(err.Error(), "a nullable object lookup without separate null/undefined tags")) {
		t.Fatalf("want explicit storage view stop, got %v", err)
	}
	t.Log(err)
}
func TestNullableObjectStorageArray(t *testing.T) {
	t.Parallel()
	nullableObjectStorageOutcome(t, "array")
}
func TestNullableObjectStorageField(t *testing.T) {
	t.Parallel()
	nullableObjectStorageOutcome(t, "field")
}
func TestNullableObjectStorageCallable(t *testing.T) {
	t.Parallel()
	nullableObjectStorageOutcome(t, "callable")
}
func TestNullableObjectStorageArraySpread(t *testing.T) {
	t.Parallel()
	nullableObjectStorageOutcome(t, "array_spread")
}
func TestNullableObjectStorageObjectSpread(t *testing.T) {
	t.Parallel()
	nullableObjectStorageOutcome(t, "object_spread")
}
func TestNullableObjectStorageLookup(t *testing.T) {
	t.Parallel()
	nullableObjectStorageOutcome(t, "lookup")
}
