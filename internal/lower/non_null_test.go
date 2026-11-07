package lower

import (
	"context"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

func TestNonNullAssertionLowersToNullishPanic(test *testing.T) {
	test.Parallel()
	program, lowerTypeScriptAssertionSourceError := lowerTypeScriptAssertionSource(
		test,
		"const map = new Map<string, number>();\nconsole.log(`${map.get('a')!}`);\n",
	)
	if lowerTypeScriptAssertionSourceError != nil {
		test.Fatal(lowerTypeScriptAssertionSourceError)
	}
	var check ir.Coalesce
	found := false
	walk(program.Main, func(node any) bool {
		if value, ok := node.(ir.Coalesce); ok && value.Panic != nil {
			check = value
			found = true
		}
		return true
	})
	ok := found
	if !ok || check.Type() != ir.Number || check.Value.Type() != ir.MaybeNumber {
		test.Fatalf("want checked maybe-number unwrap, got %#v", program.Main)
	}
	message := check.Panic.(ir.StringConstant)
	if got := program.Strings[message.Index]; !strings.HasSuffix(got, ": map.get('a')! is null or undefined") ||
		!strings.HasPrefix(got, "non-null assertion failed at ") {
		test.Fatalf("unexpected panic text %q", got)
	}
}

func TestNonNullAssertionOnPresentTypeIsErased(test *testing.T) {
	test.Parallel()
	program, lowerTypeScriptAssertionSourceError := lowerTypeScriptAssertionSource(test, "const value = 0!;\n")
	if lowerTypeScriptAssertionSourceError != nil {
		test.Fatal(lowerTypeScriptAssertionSourceError)
	}
	if _, ok := program.Main[0].(ir.Declare).Value.(ir.NumberConstant); !ok {
		test.Fatalf("want a plain number, got %#v", program.Main[0])
	}
	if len(program.Strings) != 0 {
		test.Fatalf("redundant assertion emitted panic text: %q", program.Strings)
	}
}

func lowerTypeScriptAssertionSource(test *testing.T, source string) (*ir.Program, error) {
	test.Helper()
	program, loadError := load.Load([]string{writeAgreementSource(test, "main.ts", source)})
	if loadError != nil {
		test.Fatal(loadError)
	}
	return Lower(context.Background(), program)
}
