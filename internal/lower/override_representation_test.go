package lower

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAbfe962OverrideDefaultAddedRefusesNativeSignature(t *testing.T) {
	t.Parallel()
	assertOverrideParameterRefusal(t, "abfe962_override_default_added.a", "factor")
}

func TestOverrideOptionalNumberRefusesNativeSignature(t *testing.T) {
	t.Parallel()
	assertOverrideParameterRefusal(t, "override_optional_number.a", "factor")
}

func TestOverrideOptionalRefusesNativeSignature(t *testing.T) {
	t.Parallel()
	assertOverrideParameterRefusal(t, "override_optional.a", "factor")
}

func assertOverrideParameterRefusal(t *testing.T, fixture, parameter string) {
	t.Helper()
	source, err := os.ReadFile(filepath.Join("testdata", "override_representation", fixture))
	if err != nil {
		t.Fatal(err)
	}
	_, err = lowerSource(t, string(source))
	if err == nil {
		t.Fatalf("%s: lowering accepted the program; want NotYet naming parameter %q and its repair", fixture, parameter)
	}
	var gap *NotYet
	if !errors.As(err, &gap) || !strings.Contains(err.Error(), `parameter "`+parameter+`"`) || !strings.Contains(err.Error(), "keep the base method's parameter form") {
		t.Fatalf("%s: want NotYet naming parameter %q and its repair, got %v", fixture, parameter, err)
	}
}
