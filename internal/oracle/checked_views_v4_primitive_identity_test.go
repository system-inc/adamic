package oracle

import (
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/javascript"
)

func TestV4PrimitiveUnionIdentityEmission(t *testing.T) {
	t.Parallel()
	source := `function same(left:number|string|boolean|null|undefined,right:number|string|boolean|null|undefined):boolean { return left===right; }
function objectSame(left:number|string|boolean|null|undefined,right:number|string|boolean|null|undefined):boolean { return Object.is(left,right); }
console.log(` + "`" + `${same(7,7)}` + "`" + `);console.log(` + "`" + `${same('ok','bad')}` + "`" + `);console.log(` + "`" + `${same(true,true)}` + "`" + `);console.log(` + "`" + `${same(null,null)}` + "`" + `);console.log(` + "`" + `${same(undefined,undefined)}` + "`" + `);console.log(` + "`" + `${objectSame(0,-0)}` + "`" + `);console.log(` + "`" + `${objectSame(NaN,NaN)}` + "`" + `);`
	program, truth := v4IdentityProgram(t, source, "true\nfalse\ntrue\ntrue\ntrue\nfalse\ntrue\n")
	code := javascript.JavaScript(program)
	// Only the helper declaration and adapter runtime may normalize roots; neither
	// comparison of primitive-only union parameters should do so.
	for _, line := range strings.Split(code, "\n") {
		if strings.Contains(line, "return ") && strings.Contains(line, "adamicViewAdapterUnderlying(") && (strings.Contains(line, " === ") || strings.Contains(line, "Object.is(")) {
			t.Fatalf("primitive union still normalizes identity: %s", line)
		}
	}
	v4EscapeAdmitted(t, program, truth)
}
