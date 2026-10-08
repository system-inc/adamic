package lower

import (
	"errors"
	"strings"
	"testing"
)

func TestScoutClassBooleanKeepsUnsafeViewsRefused(t *testing.T) {
	_, err := lowerSource(t, `class Base { flag: boolean | undefined = undefined; } class Child extends Base { override flag: boolean = true; } const value: Base = new Child(); value.flag = undefined;`)
	var refused *Refused
	if !errors.As(err, &refused) || !strings.Contains(err.Error(), "invariant-mutable") {
		t.Fatalf("want invariant mutable refusal, got %v", err)
	}
	for _, source := range []string{
		`class Flags { flag?: boolean; } const flags = new Flags();`,
		`class Flags { flag: string | boolean | undefined = undefined; } const flags = new Flags();`,
	} {
		if _, err := lowerSource(t, source); err == nil || !strings.Contains(err.Error(), "a field of type") {
			t.Fatalf("neighbor must remain stopped, got %v", err)
		}
	}
}
