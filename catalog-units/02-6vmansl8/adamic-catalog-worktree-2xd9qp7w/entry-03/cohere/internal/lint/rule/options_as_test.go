package rule

import (
	"strings"
	"testing"
)

type optionsAsTestOptions struct {
	Enabled bool
}

// No options configured is the honest miss: the zero value and false, so the rule's default arm runs.
func TestOptionsAsGivesZeroAndFalseForNoOptions(t *testing.T) {
	t.Parallel()

	settings, ok := OptionsAs[optionsAsTestOptions](nil)
	if ok || settings != (optionsAsTestOptions{}) {
		t.Fatalf("expected the zero value and false, got %+v and %v", settings, ok)
	}
}

// The type the decoder returns comes back with true. DecodeOptionsInto is used, because it is the
// decoder 93 registrations carry and the one whose value type the unified-signatures bug missed.
func TestOptionsAsReadsTheDecodedValue(t *testing.T) {
	t.Parallel()

	decoded, err := DecodeOptionsInto[optionsAsTestOptions]()([]byte(`{"enabled": true}`))
	if err != nil {
		t.Fatal(err)
	}
	settings, ok := OptionsAs[optionsAsTestOptions](decoded)
	if !ok || !settings.Enabled {
		t.Fatalf("expected the decoded value and true, got %+v and %v", settings, ok)
	}
}

// The defect itself: a pointer read over the value the decoder returns must fail loudly, naming both
// types, where the comma-ok assertion it replaces returned nil and false and the options vanished.
func TestOptionsAsPanicsWhenTheDecoderAndTheRuleDisagree(t *testing.T) {
	t.Parallel()

	decoded, err := DecodeOptionsInto[optionsAsTestOptions]()([]byte(`{"enabled": true}`))
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		recovered := recover()
		message, isString := recovered.(string)
		if !isString {
			t.Fatalf("expected a panic with a message, got %v", recovered)
		}
		for _, name := range []string{"*rule.optionsAsTestOptions", "rule.optionsAsTestOptions"} {
			if !strings.Contains(message, name) {
				t.Fatalf("expected the panic to name %s, got: %s", name, message)
			}
		}
	}()
	OptionsAs[*optionsAsTestOptions](decoded)
}
