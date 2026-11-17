package lower

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

// A scratch probe for #ha4hbqt, never landed: the four "plain" member casts of an element read, under Adamic's own
// options (noUncheckedIndexedAccess), modeled on tsc's declarations. If an element read types as T | undefined, the
// absent arm (main cf468df0) should take them.
func TestCastElementReadProbe(t *testing.T) {
	output := os.Getenv("ADAMIC_ELEM_PROBE_OUT")
	if output == "" {
		t.Skip("set ADAMIC_ELEM_PROBE_OUT")
	}
	model := `enum SK { Identifier = 1, PropertyAccess = 2, ElementAccess = 3, Parameter = 4, JSDocParameterTag = 5, JsxText = 6, JsxExpression = 7, JsxElement = 8 }
interface Node { readonly kind: SK; readonly parent: Node | undefined }
interface Identifier extends Node { readonly kind: SK.Identifier; readonly text: string }
interface PropertyAccessExpression extends Node { readonly kind: SK.PropertyAccess; readonly name: string }
interface ElementAccessExpression extends Node { readonly kind: SK.ElementAccess; readonly index: number }
interface ParameterDeclaration extends Node { readonly kind: SK.Parameter; readonly name: string }
interface JSDocParameterTag extends Node { readonly kind: SK.JSDocParameterTag; readonly comment: string }
interface JsxText extends Node { readonly kind: SK.JsxText; readonly text: string }
interface JsxExpression extends Node { readonly kind: SK.JsxExpression; readonly expression: string }
interface JsxElement extends Node { readonly kind: SK.JsxElement; readonly tag: string }
type BindableStaticNameExpression = Identifier | PropertyAccessExpression | ElementAccessExpression;
type JsxChild = JsxText | JsxExpression | JsxElement;
`
	lines := []string{}
	for _, probe := range []struct{ name, source string }{
		{"binder.ts:3372 arguments[0]", model + "interface Call { readonly arguments: readonly BindableStaticNameExpression[] } function f(node: Call): string { return (node.arguments[0] as PropertyAccessExpression).name; }"},
		{"checker.ts:16527 parameters[0]", model + "interface Sig { readonly parameters: readonly (ParameterDeclaration | JSDocParameterTag)[] } function f(declaration: Sig): string { return (declaration.parameters[0] as ParameterDeclaration).name; }"},
		{"jsx.ts:307 nonWhitespaceChildren[0]", model + "function f(nonWhitespaceChildren: JsxChild[]): string { return (nonWhitespaceChildren[0] as JsxExpression).expression; }"},
		{"jsx.ts:341 nonWhitespaceChildren[0], held", model + "function f(nonWhitespaceChildren: readonly JsxChild[]): string { const child = nonWhitespaceChildren[0] as JsxExpression; return child.expression; }"},
	} {
		_, err := lowerSourceAs(t, probe.source, ".ts")
		verdict := "lowers"
		if err != nil {
			verdict = strings.ReplaceAll(err.Error(), "\n", " ")
			if at := strings.Index(verdict, "main.ts:"); at >= 0 {
				verdict = verdict[at:]
			}
			if len(verdict) > 220 {
				verdict = verdict[:220]
			}
		}
		lines = append(lines, fmt.Sprintf("PROBE %s: %s", probe.name, verdict))
	}
	if err := os.WriteFile(output, []byte(strings.Join(lines, "\n")+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}
