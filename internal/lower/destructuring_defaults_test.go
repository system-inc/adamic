package lower

import (
	"strings"
	"testing"
)

// A default on a field whose type holds null, or is unknown, stops before the default is judged (#smqr4vn):
// the field is a union with a null member, or unknown's boxed value, which destructuring doesn't read yet, so no program asks the default to
// tell a present null from undefined, where ?? would take the default and Node keeps null.
func TestDestructuringDefaultNullStops(t *testing.T) {
	t.Parallel()
	const want = "stage 0 can't lower a destructured name held otherwise than its field yet"
	for _, row := range []struct{ name, source, where string }{
		{"a string or null field", "interface Holder {\n\ttext?: string | null | undefined;\n}\nfunction show(holder: Holder): string {\n\tconst { text = 'd' } = holder;\n\treturn text === null ? 'null' : text;\n}\nconsole.log(show({ text: null }));\n", "5:10"},
		{"an unknown field", "interface Holder {\n\tvalue?: unknown;\n}\nfunction show(holder: Holder): string {\n\tconst { value = 1 } = holder;\n\treturn value === null ? 'null' : typeof value;\n}\nconsole.log(show({ value: null }));\n", "5:10"},
		{"a number or null field", "interface Holder {\n\tcount?: number | null | undefined;\n}\nfunction show(holder: Holder): string {\n\tconst { count = 5 } = holder;\n\treturn count === null ? 'null' : `${count}`;\n}\nconsole.log(show({ count: null }));\n", "5:10"},
	} {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, row.source)
			if err == nil || !strings.HasSuffix(err.Error(), "/main.a:"+row.where+": "+want) {
				t.Errorf("got %v, want main.a:%s: %s", err, row.where, want)
			}
		})
	}
}
