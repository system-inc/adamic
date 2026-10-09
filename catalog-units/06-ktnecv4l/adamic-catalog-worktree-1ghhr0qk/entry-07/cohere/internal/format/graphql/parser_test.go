package graphql

import (
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/format/estree"
)

// These run without node. Each expectation is what graphql-js 17.0.2 gives for the input, read from
// graphql-js when the test was written, with its UTF-16 offsets converted to bytes.

// outline renders a tree as kind[start,end] with its fields in order, so a whole shape asserts as one
// string: a nil field prints as key=undefined, an absent one does not print.
func outline(node *estree.Node) string {
	var builder strings.Builder
	var write func(value any)
	write = func(value any) {
		switch typed := value.(type) {
		case nil:
			builder.WriteString("undefined")
		case *estree.Node:
			builder.WriteString(typed.Type())
			builder.WriteString("[")
			builder.WriteString(strconv.Itoa(typed.Range[0]) + "," + strconv.Itoa(typed.Range[1]))
			builder.WriteString("]")
			if keys := typed.Keys(); len(keys) > 0 {
				builder.WriteString("{")
				for index, key := range keys {
					if index > 0 {
						builder.WriteString(" ")
					}
					builder.WriteString(key + "=")
					write(typed.Get(key))
				}
				builder.WriteString("}")
			}
		case []*estree.Node:
			builder.WriteString("(")
			for index, child := range typed {
				if index > 0 {
					builder.WriteString(" ")
				}
				write(child)
			}
			builder.WriteString(")")
		case string:
			builder.WriteString(`"` + typed + `"`)
		case bool:
			if typed {
				builder.WriteString("true")
			} else {
				builder.WriteString("false")
			}
		default:
			builder.WriteString("?")
		}
	}
	write(node)
	return builder.String()
}

func TestParseShapes(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		text string
		want string
	}{
		{
			name: "shorthand query writes its undefined fields",
			text: "{ a }",
			want: `Document[0,5]{definitions=(OperationDefinition[0,5]{operation="query" description=undefined name=undefined ` +
				`variableDefinitions=undefined directives=undefined selectionSet=SelectionSet[0,5]{selections=(` +
				`Field[2,3]{alias=undefined name=Name[2,3]{value="a"} arguments=undefined directives=undefined selectionSet=undefined})}})}`,
		},
		{
			name: "variables, alias, arguments, directive, list and non-null types",
			text: "query Q($a: [Int!] = [1]) { b: f(x: $a) @d }",
			want: `Document[0,44]{definitions=(OperationDefinition[0,44]{operation="query" description=undefined name=Name[6,7]{value="Q"} ` +
				`variableDefinitions=(VariableDefinition[8,24]{description=undefined variable=Variable[8,10]{name=Name[9,10]{value="a"}} ` +
				`type=ListType[12,18]{type=NonNullType[13,17]{type=NamedType[13,16]{name=Name[13,16]{value="Int"}}}} ` +
				`defaultValue=ListValue[21,24]{values=(IntValue[22,23]{value="1"})} directives=undefined}) directives=undefined ` +
				`selectionSet=SelectionSet[26,44]{selections=(Field[28,42]{alias=Name[28,29]{value="b"} name=Name[31,32]{value="f"} ` +
				`arguments=(Argument[33,38]{name=Name[33,34]{value="x"} value=Variable[36,38]{name=Name[37,38]{value="a"}}}) ` +
				`directives=(Directive[40,42]{name=Name[41,42]{value="d"} arguments=undefined}) selectionSet=undefined})}})}`,
		},
		{
			name: "a fragment spread has arguments only when they are written",
			text: "{ ...F(x: 1) ...G }",
			want: `Document[0,19]{definitions=(OperationDefinition[0,19]{operation="query" description=undefined name=undefined ` +
				`variableDefinitions=undefined directives=undefined selectionSet=SelectionSet[0,19]{selections=(` +
				`FragmentSpread[2,12]{name=Name[5,6]{value="F"} arguments=(FragmentArgument[7,11]{name=Name[7,8]{value="x"} value=IntValue[10,11]{value="1"}}) directives=undefined} ` +
				`FragmentSpread[13,17]{name=Name[16,17]{value="G"} directives=undefined})}})}`,
		},
		{
			name: "directive definition with repeatable, empty list and object",
			text: "directive @d(a: [Int] = [], b: I = {}) repeatable on | FIELD",
			want: `Document[0,60]{definitions=(DirectiveDefinition[0,60]{description=undefined name=Name[11,12]{value="d"} arguments=(` +
				`InputValueDefinition[13,26]{description=undefined name=Name[13,14]{value="a"} type=ListType[16,21]{type=NamedType[17,20]{name=Name[17,20]{value="Int"}}} ` +
				`defaultValue=ListValue[24,26]{values=()} directives=undefined} ` +
				`InputValueDefinition[28,37]{description=undefined name=Name[28,29]{value="b"} type=NamedType[31,32]{name=Name[31,32]{value="I"}} ` +
				`defaultValue=ObjectValue[35,37]{fields=()} directives=undefined}) directives=undefined repeatable=true locations=(Name[55,60]{value="FIELD"})})}`,
		},
		{
			name: "offsets are bytes, after multi-byte characters",
			text: "\"é😀\" scalar S",
			want: `Document[0,17]{definitions=(ScalarTypeDefinition[0,17]{description=StringValue[0,8]{value="é😀" block=false} name=Name[16,17]{value="S"} directives=undefined})}`,
		},
		{
			name: "interface implementing interfaces, extension",
			text: "extend interface I implements & J & K",
			want: `Document[0,37]{definitions=(InterfaceTypeExtension[0,37]{name=Name[17,18]{value="I"} interfaces=(` +
				`NamedType[32,33]{name=Name[32,33]{value="J"}} NamedType[36,37]{name=Name[36,37]{value="K"}}) directives=undefined fields=undefined})}`,
		},
	}
	for _, each := range cases {
		document, _, err := Parse(each.text)
		if err != nil {
			t.Errorf("%s: %v", each.name, err)
			continue
		}
		if got := outline(document); got != each.want {
			t.Errorf("%s:\n got %s\nwant %s", each.name, got, each.want)
		}
	}
}

func TestParseStringValues(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		text string
		want string
	}{
		{name: "escapes", text: `{ a(s: "q\" b\\ s\/ n\n t\t") }`, want: "q\" b\\ s/ n\n t\t"},
		{name: "variable-width unicode escape", text: `{ a(s: "\u{1F600}x") }`, want: "😀x"},
		{name: "surrogate pair escape", text: "{ a(s: \"\x5cuD83D\x5cuDE00\") }", want: "😀"},
		{name: "block string dedent", text: "{ a(s: \"\"\"\n\n    first\n      indented\n    \\\"\"\" last\n\n\"\"\") }", want: "first\n  indented\n\"\"\" last"},
		{name: "block string first line kept", text: "{ a(s: \"\"\"  one\n    two\"\"\") }", want: "  one\ntwo"},
		{name: "block string CRLF", text: "{ a(s: \"\"\"\r\n  x\r\n  y\r\n\"\"\") }", want: "x\ny"},
	}
	for _, each := range cases {
		document, _, err := Parse(each.text)
		if err != nil {
			t.Errorf("%s: %v", each.name, err)
			continue
		}
		value := document.List("definitions")[0].Child("selectionSet").List("selections")[0].List("arguments")[0].Child("value")
		if got := value.String("value"); got != each.want {
			t.Errorf("%s: got %q, want %q", each.name, got, each.want)
		}
		if want := strings.Contains(each.text, `"""`); value.Bool("block") != want {
			t.Errorf("%s: block is %v", each.name, value.Bool("block"))
		}
	}
}

func TestParseComments(t *testing.T) {
	t.Parallel()
	text := "# one 😀\n{ a # two\n}\n#"
	_, comments, err := Parse(text)
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, comment := range comments {
		got = append(got, comment.Type()+"["+strconv.Itoa(comment.Range[0])+","+strconv.Itoa(comment.Range[1])+"]"+comment.String("value"))
	}
	want := []string{"Comment[0,10] one 😀", "Comment[15,20] two", "Comment[23,24]"}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestParseErrors(t *testing.T) {
	t.Parallel()
	cases := []struct {
		text string
		want string
	}{
		{text: "{ a", want: "Syntax Error: Expected Name, found <EOF>. (1:4)"},
		{text: "{ a(b: \"😀\") ? }", want: `Syntax Error: Unexpected character: "?". (1:14)`},
		{text: "{\r\n a 😀 }", want: "Syntax Error: Unexpected character: U+1F600. (2:4)"},
		{text: "extend type Foo", want: "Syntax Error: Unexpected <EOF>. (1:16)"},
		{text: "enum E { null }", want: `Syntax Error: Name "null" is reserved and cannot be used for an enum value. (1:10)`},
		{text: `{ a(s: "\u{110000}") }`, want: `Syntax Error: Invalid Unicode escape sequence: "\u{110000}". (1:9)`},
		{text: "type T { f(a: Int = $v): Int }", want: `Syntax Error: Unexpected variable "$v" in constant value. (1:21)`},
	}
	for _, each := range cases {
		document, comments, err := Parse(each.text)
		if err == nil || document != nil || comments != nil {
			t.Errorf("%q: parsed, want %s", each.text, each.want)
			continue
		}
		if err.Error() != each.want {
			t.Errorf("%q:\n got %s\nwant %s", each.text, err.Error(), each.want)
		}
	}
}
