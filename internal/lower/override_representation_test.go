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
	source, err := os.ReadFile(filepath.Join("testdata", "override_representation", fixture))
	if err != nil {
		test.Fatal(err)
	}
	_, err = lowerSource(test, string(source))
	if err == nil {
		test.Fatalf(
			"%s: lowering accepted the program; want NotYet naming parameter %q and its repair",
			fixture,
			parameter,
		)
	}
	var gap *NotYet
	if !errors.As(err, &gap) ||
		!strings.Contains(err.Error(), `parameter "`+parameter+`"`) ||
		!strings.Contains(err.Error(), "keep the base method's parameter form") {
		test.Fatalf("%s: want NotYet naming parameter %q and its repair, got %v", fixture, parameter, err)
	}
}
