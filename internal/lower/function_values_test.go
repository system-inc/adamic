package lower

import (
	"errors"
	"testing"
)

func TestFunctionValueUnionViewsStayNotYet(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"function show(value: string | number): string { return `${value}`; }\nconst numeric: (value: number) => string = show;\nconsole.log(numeric(7));\n",
		"const show = (value: string | number): string => `${value}`;\nconst numeric: (value: number) => string = show;\nconsole.log(numeric(7));\n",
		"function number(): number { return 7; }\nconst mixed: () => string | number = number;\nconsole.log(`${mixed()}`);\n",
		"function show(value: string | number): string { return `${value}`; }\nconst original = { show };\nconst holder: { readonly show: (value: number) => string } = original;\nconsole.log(holder.show(7));\n",
	} {
		_, err := lowerSource(t, source)
		var notYet *NotYet
		if !errors.As(err, &notYet) {
			t.Errorf("callable view requires a union slot adapter, got %v", err)
		}
	}
}
