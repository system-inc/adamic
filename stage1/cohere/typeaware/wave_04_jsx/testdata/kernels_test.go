// Owned overlay: expectations call production helpers, never copied predicates.
package react

import (
	"encoding/json"
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"os"
	"strings"
	"testing"
)

func TestWave04JsxKernels(t *testing.T) {
	// Not parallel: this oracle writes one caller-supplied artifact.
	names := []string{"", "div", "App", "app", "_foo", "$foo", "테스트", "É", "Foo-bar", "x-gif", "A-B", "a", "z", "Z", "1", "😀", "this", "React", "Provider", "a\nB"}
	kinds := []ast.Kind{ast.KindObjectLiteralExpression, ast.KindArrayLiteralExpression, ast.KindArrowFunction, ast.KindFunctionExpression, ast.KindClassExpression, ast.KindNewExpression, ast.KindRegularExpressionLiteral, ast.KindJsxElement, ast.KindJsxSelfClosingElement, ast.KindJsxFragment, ast.KindIdentifier, ast.KindCallExpression, ast.KindStringLiteral, ast.KindNumericLiteral, ast.KindNonNullExpression}
	var expected strings.Builder
	for _, name := range names {
		fmt.Fprintf(&expected, "name %t\n", isComponentName(name))
	}
	ids := []int{}
	for _, kind := range kinds {
		ids = append(ids, int(kind))
		fmt.Fprintf(&expected, "construction %s\n", jsxNoConstructedContextValuesConstructionKind(&ast.Node{Kind: kind}))
	}
	for _, name := range []string{"object", "array", "function expression", "function declaration", "class expression", "assignment expression", "JSX fragment", ""} {
		fmt.Fprintf(&expected, "function %t\n", jsxNoConstructedContextValuesIsFunctionKind(name))
	}
	for _, kind := range kinds {
		label := jsxNoConstructedContextValuesConstructionKind(&ast.Node{Kind: kind})
		if label == "" {
			continue
		}
		m := jsxNoConstructedContextValuesMessage(jsxNoConstructedContextValuesConstruction{kind: label}, 3, 0, "")
		fmt.Fprintf(&expected, "4\t17\treact/jsx-no-constructed-context-values\t%s\t%s\t0\t0\n", m.Id, m.Description)
	}

	messages := []struct{ Rule, Id, Text string }{{"react/jsx-fragments", messagePreferFragmentShorthand.Id, messagePreferFragmentShorthand.Description}, {"react/jsx-fragments", messagePreferFragmentPragma.Id, messagePreferFragmentPragma.Description}, {"react/jsx-no-undef", messageJsxIdentifierNotDefined.Id, messageJsxIdentifierNotDefined.Description}}
	for _, m := range messages {
		fmt.Fprintf(&expected, "4\t17\t%s\t%s\t%s\t0\t0\n", m.Rule, m.Id, m.Text)
	}
	listeners := [][]int{{int(ast.KindJsxElement), int(ast.KindJsxSelfClosingElement), int(ast.KindJsxFragment)}, {int(ast.KindJsxOpeningElement), int(ast.KindJsxSelfClosingElement)}, {int(ast.KindJsxOpeningElement), int(ast.KindJsxSelfClosingElement)}}
	listenerNames := [][]string{}
	for _, kinds := range listeners {
		names := []string{}
		for _, kind := range kinds {
			names = append(names, strings.TrimPrefix(ast.Kind(kind).String(), "Kind"))
		}
		listenerNames = append(listenerNames, names)
	}
	data := struct {
		Names         []string
		Kinds         []int
		FunctionNames []string
		Listeners     [][]int
		ListenerNames [][]string
		Expected      string
	}{names, ids, []string{"object", "array", "function expression", "function declaration", "class expression", "assignment expression", "JSX fragment", ""}, listeners, listenerNames, expected.String()}
	encoded, err := json.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(os.Getenv("WAVE04_JSX_KERNELS"), encoded, 0600); err != nil {
		t.Fatal(err)
	}
}
