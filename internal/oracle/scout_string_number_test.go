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

func nullableObjectStorageOutcome(test *testing.T, name string) {
	test.Helper()
	path, err := filepath.Abs(
		filepath.Join(repository, "internal/oracle/testdata/scout_nullable_storage_refused", name+".a"),
	)
	if err != nil {
		test.Fatal(err)
	}
	if difference := disagreement(run{stdout: []byte("true\n")}, onNode(test, path)); difference != "" {
		test.Fatal(difference)
	}
	_, err = lowered(test, path)
	if name == "array_spread" {
		var refusal *lower.Refused
		if !errors.As(err, &refusal) ||
			refusal.What != "union-typed array built from a narrower literal" ||
			refusal.Fix != "declare the literal's element type explicitly, or build the array with a mixed literal" {
			test.Fatalf("want the newer narrow-literal layout refusal with its fix, got %v", err)
		}
		test.Log(err)
		return
	}
	if err == nil || !(strings.Contains(err.Error(), "sharing nullable object storage") || name == "lookup" &&
		strings.Contains(err.Error(), "a nullable object lookup without separate null/undefined tags")) {
		test.Fatalf("want explicit storage view stop, got %v", err)
	}
	test.Log(err)
}
func TestNullableObjectStorageArray(test *testing.T) {
	test.Parallel()
	nullableObjectStorageOutcome(test, "array")
}
func TestNullableObjectStorageField(test *testing.T) {
	test.Parallel()
	nullableObjectStorageOutcome(test, "field")
}
func TestNullableObjectStorageCallable(test *testing.T) {
	test.Parallel()
	nullableObjectStorageOutcome(test, "callable")
}
func TestNullableObjectStorageArraySpread(test *testing.T) {
	test.Parallel()
	nullableObjectStorageOutcome(test, "array_spread")
}
func TestNullableObjectStorageObjectSpread(test *testing.T) {
	test.Parallel()
	nullableObjectStorageOutcome(test, "object_spread")
}
func TestNullableObjectStorageLookup(test *testing.T) {
	test.Parallel()
	nullableObjectStorageOutcome(test, "lookup")
}
