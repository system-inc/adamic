package lower

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAbfe962OverrideDefaultAddedRefusesNativeSignature(test *testing.T) {
	test.Parallel()
	assertOverrideParameterRefusal(test, "abfe962_override_default_added.a", "factor")
}

func TestOverrideOptionalNumberRefusesNativeSignature(test *testing.T) {
	test.Parallel()
	assertOverrideParameterRefusal(test, "override_optional_number.a", "factor")
}

func TestOverrideOptionalRefusesNativeSignature(test *testing.T) {
	test.Parallel()
	assertOverrideParameterRefusal(test, "override_optional.a", "factor")
}

func assertOverrideParameterRefusal(test *testing.T, fixture, parameter string) {
	test.Helper()
	source, readError := os.ReadFile(filepath.Join("testdata", "override_representation", fixture))
	if readError != nil {
		test.Fatal(readError)
	}
	_, lowerSourceError := lowerSource(test, string(source))
	if lowerSourceError == nil {
		test.Fatalf(
			"%s: lowering accepted the program; want NotYet naming parameter %q and its repair",
			fixture,
			parameter,
		)
	}
	var gap *NotYet
	if !errors.As(lowerSourceError, &gap) ||
		!strings.Contains(lowerSourceError.Error(), `parameter "`+parameter+`"`) ||
		!strings.Contains(lowerSourceError.Error(), "keep the base method's parameter form") {
		test.Fatalf("%s: want NotYet naming parameter %q and its repair, got %v", fixture, parameter, lowerSourceError)
	}
}
