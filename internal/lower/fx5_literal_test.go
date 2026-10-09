package lower

import (
	"errors"
	"strings"
	"testing"
)

const fx5NumberLiteral = "const items: (string | number)[] = [1, 2];\nconsole.log(`${typeof items[0]} ${items[0]}`);\nconsole.log(`${typeof items[1]} ${items[1]}`);\nitems.push('a');\nconsole.log(`${typeof items[0]} ${items[0]}`);\nconsole.log(`${typeof items[1]} ${items[1]}`);\nconsole.log(`${typeof items[2]} ${items[2]}`);\n"

func TestFX5NumberLiteralAgrees(t *testing.T) {
	t.Parallel()
	lowersAndAgreesWithNode(t, fx5NumberLiteral)
	// Only native observes the array's physical element layout.
	lowersAndAgreesWithNodeNative(t, fx5NumberLiteral)
}

func TestFX5AssignedNumberLiteralAgrees(t *testing.T) {
	t.Parallel()
	source := strings.Replace(fx5NumberLiteral, "const items: (string | number)[] = [1, 2];", "let items: (string | number)[] = [0, 'seed'];\nitems = [1, 2];", 1)
	lowersAndAgreesWithNode(t, source)
	// Only native observes whether assignment builds the destination's union slots.
	lowersAndAgreesWithNodeNative(t, source)
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

func TestFX5NestedNumberLiteralAgrees(t *testing.T) {
	t.Parallel()
	source := "const rows: (string | number)[][] = [[1, 2]];\nconst items = rows[0];\nif (items !== undefined) {\nconsole.log(`${typeof items[0]} ${items[0]}`);\nconsole.log(`${typeof items[1]} ${items[1]}`);\nitems.push('a');\nconsole.log(`${typeof items[0]} ${items[0]}`);\nconsole.log(`${typeof items[1]} ${items[1]}`);\nconsole.log(`${typeof items[2]} ${items[2]}`);\n}\n"
	lowersAndAgreesWithNode(t, source)
	// Only native observes whether the nested literal uses boxed union slots.
	lowersAndAgreesWithNodeNative(t, source)
}

func TestFX5ConstTupleUsesSeparatePath(t *testing.T) {
	t.Parallel()
	source := "const items: readonly (string | number)[] = [1, 2] as const;\nconsole.log(`${typeof items[0]} ${items[0]}`);\nconsole.log(`${typeof items[1]} ${items[1]}`);\n"
	_, err := lowerSource(t, source)
	var gap *NotYet
	if !errors.As(err, &gap) || gap.What != "a tuple where an array goes (as readonly (string | number)[])" || !strings.HasSuffix(gap.Where, "/main.a:1:45") {
		t.Fatalf("want separate tuple-to-array boundary at main.a:1:45, got %v", err)
	}
}
