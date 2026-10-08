package lower

import (
	"errors"
	"strings"
	"testing"
)

func TestClockGenericReturnsT04UnsupportedShapes(t *testing.T) {
	for _, extra := range []string{
		"(): string;", "new(): object;", "[key: string]: object;", "extra: Map<string, string>;",
	} {
		t.Run(extra, func(t *testing.T) {
			_, err := lowerSource(t, `interface A { text: string; } interface B { text: string; owner: string; }
interface Result<T> { expression: object; typeArguments: readonly string[]; `+extra+` }
function result<T>(): Result<T> & { expression: A | B } { throw new Error("unused"); }
result<string>();`)
			var stopped *NotYet
			var refused *Refused
			if extra == "[key: string]: object;" && errors.As(err, &refused) && strings.Contains(err.Error(), "index signature") {
				return
			}
			if !errors.As(err, &stopped) || !strings.Contains(err.Error(), "a function returning") {
				t.Fatalf("unsupported intersection must stop: %v", err)
			}
		})
	}
}
