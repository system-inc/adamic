package lower

import (
	"errors"
	"strings"
	"testing"
)

func TestScoutClassBooleanKeepsUnsafeViewsRefused(test *testing.T) {
	test.Parallel()
	_, lowerSourceError := lowerSource(
		test,
		`class Base { flag: boolean | undefined = undefined; } class Child extends Base { override flag: boolean = true; } const value: Base = new Child(); value.flag = undefined;`,
	)
	var refused *Refused
	if !errors.As(lowerSourceError, &refused) || !strings.Contains(lowerSourceError.Error(), "invariant-mutable") {
		test.Fatalf("want invariant mutable refusal, got %v", lowerSourceError)
	}
	for _, source := range []string{
		`class Flags { flag?: boolean; } const flags = new Flags();`,
		`class Flags { flag: string | boolean | undefined = undefined; } const flags = new Flags();`,
	} {
		if _, lowerSourceError := lowerSource(
			test,
			source,
		); lowerSourceError == nil || !strings.Contains(lowerSourceError.Error(), "a field of type") {
			test.Fatalf("neighbor must remain stopped, got %v", lowerSourceError)
		}
	}
}
