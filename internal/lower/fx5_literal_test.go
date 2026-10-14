package lower

import (
	"errors"
	"strings"
	"testing"
)

const fx5NumberLiteral = "const items: (string | number)[] = [1, 2];\nconsole.log(`${typeof items[0]} ${items[0]}`);\n"

func TestFX5NumberLiteralRefused(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, fx5NumberLiteral)
	var refusal *Refused
	if !errors.As(err, &refusal) || refusal.What != "union-typed array built from a narrower literal" || !strings.HasSuffix(refusal.Where, "/main.a:1:36") || refusal.Fix != "declare the literal's element type explicitly, or build the array with a mixed literal" {
		t.Fatalf("want narrow refusal at literal main.a:1:36 with fix, got %v", err)
	}
}

func TestFX5NumberValueRefused(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, "const numbers: number[] = [1, 2];\nconst items: (string | number)[] = numbers;\nconsole.log(`${typeof items[0]} ${items[0]}`);\n")
	var refusal *Refused
	if !errors.As(err, &refusal) || refusal.What != "a value of type number[] seen as (string | number)[], which can write string | number where number is read" || !strings.HasSuffix(refusal.Where, "/main.a:2:36") || !strings.Contains(refusal.Fix, "adamic/invariant-mutable") {
		t.Fatalf("want existing invariance refusal at value main.a:2:36, got %v", err)
	}
}

func TestFX5MixedLiteralAgrees(t *testing.T) {
	t.Parallel()
	source := "const items: (string | number)[] = [1, 'a'];\nconsole.log(`${typeof items[0]} ${items[0]}`);\nconsole.log(`${typeof items[1]} ${items[1]}`);\n"
	lowersAndAgreesWithNode(t, source)
	// Only native observes whether union readers match the literal's physical layout.
	lowersAndAgreesWithNodeNative(t, source)
}
