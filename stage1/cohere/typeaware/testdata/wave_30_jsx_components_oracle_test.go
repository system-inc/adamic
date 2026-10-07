// Overlay in production React package; calls the actual private Go helpers.
package react

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf16"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

func wave30JsxWritten(text string) string {
	var out strings.Builder
	for _, r := range text {
		if r >= 32 && r <= 126 && r != 92 {
			out.WriteRune(r)
		} else if r <= 65535 {
			fmt.Fprintf(&out, `\u%04x`, r)
		} else {
			a, b := utf16.EncodeRune(r)
			fmt.Fprintf(&out, `\u%04x\u%04x`, a, b)
		}
	}
	return out.String()
}
func TestWave30JsxComponentOracle(t *testing.T) {
	directory := os.Getenv("ADAMIC_WAVE30_JSX_COMPONENTS")
	if directory == "" {
		t.Fatal("missing component output directory")
	}
	factory := ast.NewNodeFactory(ast.NodeFactoryHooks{})
	var fragments strings.Builder
	for _, objectKind := range []ast.Kind{ast.KindIdentifier, ast.KindStringLiteral} {
		for _, propertyKind := range []ast.Kind{ast.KindIdentifier, ast.KindStringLiteral} {
			for _, objectName := range []string{"", "React", "react", "React.Fragment"} {
				for _, propertyName := range []string{"", "Fragment", "fragment"} {
					object := factory.NewIdentifier(objectName)
					if objectKind == ast.KindStringLiteral {
						object = factory.NewStringLiteral(objectName, 0)
					}
					property := factory.NewIdentifier(propertyName)
					if propertyKind == ast.KindStringLiteral {
						property = factory.NewStringLiteral(propertyName, 0)
					}
					node := factory.NewPropertyAccessExpression(object, nil, property, 0)
					fmt.Fprintf(&fragments, "%d,%d,%s,%s\t%t\n", objectKind, propertyKind, objectName, propertyName, jsxFragmentsNameIsFragment(rule.Context{}, node))
				}
			}
		}
	}
	var undef strings.Builder
	for code := 0; code < 256; code++ {
		for _, suffix := range []string{"", "x", "-x", "x-y"} {
			fmt.Fprintf(&undef, "%d,%s\t%t\n", code, suffix, isComponentName(string(rune(code))+suffix))
		}
	}
	for _, name := range []string{"", "테스트", "😀", "Foo-bar", "$foo", "_foo", "Élément", "div", "A", "x-gif"} {
		fmt.Fprintf(&undef, "%s\t%t\n", wave30JsxWritten(name), isComponentName(name))
	}
	var context strings.Builder
	kinds := []string{"object", "array", "function expression", "function declaration", "class expression", "new expression", "regular expression", "JSX element", "JSX fragment", "assignment expression"}
	for construction, kind := range kinds {
		for _, named := range []bool{false, true} {
			for _, name := range []string{"", "theme", "a\"b", "line\nbreak", "\\", string(rune(7)), string(rune(127)), "테마", "Élément", "😀", "a😀b", "\u0085", "\u00a0", "\u0378", "\u200b", "\u2028", "\ufeff", "\uffff", "\U00010000", "\U0001fffe", "\U0010ffff", "\U000e0001"} {
				input := jsxNoConstructedContextValuesConstruction{kind: kind}
				if named {
					input.usage = factory.NewIdentifier(name)
				}
				message := jsxNoConstructedContextValuesMessage(input, 3, 9, name)
				fmt.Fprintf(&context, "%d,%t,%s\t%s\t%s\n", construction, named, wave30JsxWritten(name), message.Id, wave30JsxWritten(message.Description))
			}
		}
	}
	for name, data := range map[string]string{"fragments": fragments.String(), "undef": undef.String(), "context": context.String()} {
		if err := os.WriteFile(filepath.Join(directory, name+"-go.stdout"), []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
	}
}
