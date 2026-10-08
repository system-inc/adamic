package lower

import (
	"errors"
	"strings"
	"testing"
)

func TestStringBrandSignatureGuards(t *testing.T) {
	t.Parallel()
	for _, brand := range []string{
		"{ __escapedIdentifier: number }",
		"{ __escapedIdentifier: never }",
		"{ toString: void }",
		"{ '0': void }",
		"{ __escapedIdentifier: void; (): string }",
		"{ __escapedIdentifier: void; [key: string]: void }",
	} {
		t.Run(brand, func(t *testing.T) {
			source := `enum InternalSymbolName { Call = "__call" }
 type Bad = (string & ` + brand + `) | InternalSymbolName;
 function invalid(): Bad { return InternalSymbolName.Call; }
 console.log(typeof invalid());`
			_, err := lowerSource(t, source)
			if strings.Contains(brand, "[key:") {
				var refused *Refused
				if !errors.As(err, &refused) || refused.What != "an index signature" {
					t.Fatalf("index brand must retain its existing refusal, got %v", err)
				}
				return
			}
			var notYet *NotYet
			if !errors.As(err, &notYet) || !strings.Contains(notYet.What, "a function returning Bad") {
				t.Fatalf("unproven branded signature must stay NotYet, got %v", err)
			}
		})
	}
}
