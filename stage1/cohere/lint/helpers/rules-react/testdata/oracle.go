//go:build lintoracle

package main

import (
	"encoding/json"
	"fmt"
	"github.com/system-inc/cohere/internal/lint/ecmascript/comments"
	rules "github.com/system-inc/cohere/internal/lint/rules/react"
	"os"
	"strings"
	"unicode"
	"unicode/utf16"
)

func isComponentIdentifierName(name string) bool {
	if name == "" {
		return false
	}
	first := name[0]
	return first >= 'A' && first <= 'Z'
}
func mentionsRef(name string) bool {
	return strings.Contains(name, "ref") || strings.Contains(name, "Ref")
}
func isJavaScriptIdentifier(name string) bool {
	if name == "" {
		return false
	}
	for index := 0; index < len(name); index++ {
		character := name[index]
		switch {
		case character == '_' || character == '$':
		case character >= 'a' && character <= 'z':
		case character >= 'A' && character <= 'Z':
		case index > 0 && character >= '0' && character <= '9':
		default:
			return false
		}
	}
	return true
}
func commentValueOf(comment comments.Comment) string {
	text := comment.Text
	if comment.IsBlock {
		text = strings.TrimPrefix(text, "/*")
		return strings.TrimSuffix(text, "*/")
	}
	// The `//` strip is EQUIVALENT under the current `jsxAnnotationIn` and is kept anyway.
	//
	// Measured: a mutant replacing this line with a bare `return text` survives the whole fixture
	// set, and its inverse (returning the empty string) is caught by five lines, so the arm is
	// reached and the fixtures can see it. The reason no input distinguishes the two is mechanical:
	// `jsxAnnotationIn` searches for `@jsx` at any offset and reads the name from what FOLLOWS it,
	// so a `//` sitting before the marker can never land inside a captured name. The `*/` on the
	// block arm can, which is why that one has a fixture and this one has this paragraph.
	//
	// Kept because it is what ESLint hands a rule, and because the equivalence is a property of the
	// current scanner rather than of the rule: a future `jsxAnnotationIn` anchored at the start of
	// the text would make this line load-bearing again with nothing to announce the change.
	return strings.TrimPrefix(text, "//")
}
func jsxAnnotationIn(text string) (string, bool) {
	const marker = "@jsx"
	for offset := 0; ; {
		index := strings.Index(text[offset:], marker)
		if index < 0 {
			return "", false
		}
		rest := text[offset+index+len(marker):]
		trimmed := strings.TrimLeft(rest, " \t\r\n\f\v")
		// Upstream's `\s+` requires at least one space, so `@jsxFoo` does not match. Comparing
		// lengths is how that requirement survives the trim.
		if len(trimmed) < len(rest) && trimmed != "" {
			name := trimmed
			if end := strings.IndexAny(name, " \t\r\n\f\v"); end >= 0 {
				name = name[:end]
			}
			if dot := strings.IndexByte(name, '.'); dot >= 0 {
				name = name[:dot]
			}
			if name != "" {
				return name, true
			}
		}
		offset += index + len(marker)
	}
}
func isReactComponentBaseName(name string) bool {
	return name == "Component" || name == "PureComponent"
}
func main() {
	var rows []struct {
		Name, Text string
		Block      bool
	}
	b, e := os.ReadFile(os.Args[1])
	if e != nil {
		panic(e)
	}
	if e = json.Unmarshal(b, &rows); e != nil {
		panic(e)
	}
	for _, r := range rows {
		out := ""
		switch r.Name {
		case "DecodeCompilerRuleOptions":
			_, err := rules.DecodeCompilerRuleOptions([]byte(r.Text))
			if err != nil {
				out = err.Error()
			}
		case "DecodeNoMethodSetStateOptions":
			value, err := rules.DecodeNoMethodSetStateOptions([]byte(r.Text))
			out = fmt.Sprint(value.(rules.NoMethodSetStateOptions).DisallowInFunc) + "|"
			if err != nil {
				out += err.Error()
			}
		case "isHookIdentifierName":
			out = fmt.Sprint(isHookIdentifierName(r.Text))
		case "isComponentIdentifierName":
			out = fmt.Sprint(isComponentIdentifierName(r.Text))
		case "mentionsRef":
			out = fmt.Sprint(mentionsRef(r.Text))
		case "isJavaScriptIdentifier":
			out = fmt.Sprint(isJavaScriptIdentifier(r.Text))
		case "commentValueOf":
			out = commentValueOf(comments.Comment{Text: r.Text, IsBlock: r.Block})
		case "jsxAnnotationIn":
			name, found := jsxAnnotationIn(r.Text)
			out = fmt.Sprintf("%t|%s", found, name)
		case "isReactComponentBaseName":
			out = fmt.Sprint(isReactComponentBaseName(r.Text))
		default:
			panic(r.Name)
		}
		fmt.Printf("%s|", r.Name)
		for _, u := range utf16.Encode([]rune(out)) {
			fmt.Printf("%d,", u)
		}
		fmt.Println()
	}
}
func isHookIdentifierName(name string) bool {
	if !strings.HasPrefix(name, "use") {
		return false
	}
	rest := []rune(name[3:])
	if len(rest) == 0 {
		return true
	}
	return unicode.IsUpper(rest[0]) || (rest[0] >= '0' && rest[0] <= '9')
}
