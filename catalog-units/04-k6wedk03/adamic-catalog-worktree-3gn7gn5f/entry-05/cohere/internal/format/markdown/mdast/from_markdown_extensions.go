package mdast

import (
	"strings"

	"github.com/system-inc/cohere/internal/format/markdown/micromark"
)

// The mdast extensions the fork's parseMarkdown configures, each from its upstream package. Only the
// from-markdown halves; the to-markdown halves are serializers Prettier never calls.

type extension = struct {
	canContainEols []string
	enter          map[string]handle
	exit           map[string]handle
}

// gfmAutolinkLiteralFromMarkdown is mdast-util-gfm-autolink-literal's, without its transform: the fork's
// src/language-markdown/parse/micromark/mdast-util-gfm.js empties `transforms`, because the
// find-and-replace it runs makes nodes without positions.
func gfmAutolinkLiteralFromMarkdown() extension {
	return extension{
		enter: map[string]handle{
			"literalAutolink":      opener(newLink, nil),
			"literalAutolinkEmail": enterLiteralAutolinkValue,
			"literalAutolinkHttp":  enterLiteralAutolinkValue,
			"literalAutolinkWww":   enterLiteralAutolinkValue,
		},
		exit: map[string]handle{
			"literalAutolink": func(context *compileContext, token *micromark.Token) { context.exit(token, nil) },
			"literalAutolinkEmail": func(context *compileContext, token *micromark.Token) {
				context.config.exit["autolinkEmail"](context, token)
			},
			"literalAutolinkHttp": func(context *compileContext, token *micromark.Token) {
				context.config.exit["autolinkProtocol"](context, token)
			},
			"literalAutolinkWww": exitLiteralAutolinkWww,
		},
	}
}

func enterLiteralAutolinkValue(context *compileContext, token *micromark.Token) {
	context.config.enter["autolinkProtocol"](context, token)
}

func exitLiteralAutolinkWww(context *compileContext, token *micromark.Token) {
	context.config.exit["data"](context, token)
	context.top().URL = "http://" + context.sliceSerialize(token)
}

// gfmFootnoteFromMarkdown is mdast-util-gfm-footnote's.
func gfmFootnoteFromMarkdown() extension {
	return extension{
		enter: map[string]handle{
			"gfmFootnoteCallString": buffer,
			"gfmFootnoteCall": func(context *compileContext, token *micromark.Token) {
				empty := ""
				context.enter(context.nodes.New(Node{NodeType: "footnoteReference", Label: &empty}), token, nil)
			},
			"gfmFootnoteDefinitionLabelString": buffer,
			"gfmFootnoteDefinition": func(context *compileContext, token *micromark.Token) {
				empty := ""
				context.enter(context.nodes.New(Node{NodeType: "footnoteDefinition", Label: &empty, IsParent: true, Children: []*Node{}}), token, nil)
			},
		},
		exit: map[string]handle{
			"gfmFootnoteCallString":            exitFootnoteLabelString,
			"gfmFootnoteCall":                  func(context *compileContext, token *micromark.Token) { context.exit(token, nil) },
			"gfmFootnoteDefinitionLabelString": exitFootnoteLabelString,
			"gfmFootnoteDefinition":            func(context *compileContext, token *micromark.Token) { context.exit(token, nil) },
		},
	}
}

func exitFootnoteLabelString(context *compileContext, token *micromark.Token) {
	label := context.resume()
	node := context.top()
	node.Identifier = strings.ToLower(micromark.NormalizeIdentifier(context.sliceSerialize(token)))
	node.Label = &label
}

// gfmStrikethroughFromMarkdown is mdast-util-gfm-strikethrough's.
func gfmStrikethroughFromMarkdown() extension {
	return extension{
		canContainEols: []string{"delete"},
		enter: map[string]handle{
			"strikethrough": func(context *compileContext, token *micromark.Token) {
				context.enter(context.nodes.New(Node{NodeType: "delete", IsParent: true, Children: []*Node{}}), token, nil)
			},
		},
		exit: map[string]handle{
			"strikethrough": func(context *compileContext, token *micromark.Token) { context.exit(token, nil) },
		},
	}
}

// gfmTableFromMarkdown is mdast-util-gfm-table's.
func gfmTableFromMarkdown() extension {
	enterCell := func(context *compileContext, token *micromark.Token) {
		context.enter(context.nodes.New(Node{NodeType: "tableCell", IsParent: true, Children: []*Node{}}), token, nil)
	}
	exit := func(context *compileContext, token *micromark.Token) { context.exit(token, nil) }
	return extension{
		enter: map[string]handle{
			"table": func(context *compileContext, token *micromark.Token) {
				align := make([]string, len(token.Align))
				for index, value := range token.Align {
					if value != "none" {
						align[index] = value
					}
				}
				context.enter(context.nodes.New(Node{NodeType: "table", Align: align, IsParent: true, Children: []*Node{}}), token, nil)
				context.data.inTable = true
			},
			"tableData":   enterCell,
			"tableHeader": enterCell,
			"tableRow": func(context *compileContext, token *micromark.Token) {
				context.enter(context.nodes.New(Node{NodeType: "tableRow", IsParent: true, Children: []*Node{}}), token, nil)
			},
		},
		exit: map[string]handle{
			"codeText": func(context *compileContext, token *micromark.Token) {
				value := context.resume()
				if context.data.inTable {
					// value.replace(/\\([\\|])/g, replace), where replace keeps `\\` and unescapes `\|`.
					value = unescapeTablePipes(value)
				}
				context.top().Value = value
				context.exit(token, nil)
			},
			"table": func(context *compileContext, token *micromark.Token) {
				context.exit(token, nil)
				context.data.inTable = false
			},
			"tableData":   exit,
			"tableHeader": exit,
			"tableRow":    exit,
		},
	}
}

// unescapeTablePipes is `.replace(/\\([\\|])/g, ($0, $1) => $1 === '|' ? $1 : $0)`: a backslash before a
// pipe goes, a backslash before a backslash stays, and matches do not overlap.
func unescapeTablePipes(value string) string {
	var builder strings.Builder
	for index := 0; index < len(value); index++ {
		if value[index] == '\\' && index+1 < len(value) && (value[index+1] == '\\' || value[index+1] == '|') {
			if value[index+1] == '|' {
				builder.WriteByte('|')
			} else {
				builder.WriteString(`\\`)
			}
			index++
			continue
		}
		builder.WriteByte(value[index])
	}
	return builder.String()
}

// gfmTaskListItemFromMarkdown is mdast-util-gfm-task-list-item's.
func gfmTaskListItemFromMarkdown() extension {
	exitCheck := func(context *compileContext, token *micromark.Token) {
		node := context.stack[len(context.stack)-2]
		checked := token.Type == "taskListCheckValueChecked"
		node.Checked = &checked
	}
	return extension{
		exit: map[string]handle{
			"taskListCheckValueChecked":   exitCheck,
			"taskListCheckValueUnchecked": exitCheck,
			"paragraph":                   exitParagraphWithTaskListItem,
		},
	}
}

func exitParagraphWithTaskListItem(context *compileContext, token *micromark.Token) {
	var parent *Node
	if len(context.stack) >= 2 {
		parent = context.stack[len(context.stack)-2]
	}

	if parent != nil && parent.NodeType == "listItem" && parent.Checked != nil {
		node := context.top()
		if len(node.Children) > 0 && node.Children[0].NodeType == "text" {
			head := node.Children[0]
			var firstParagraph *Node
			for _, sibling := range parent.Children {
				if sibling.NodeType == "paragraph" {
					firstParagraph = sibling
					break
				}
			}

			if firstParagraph == node {
				// head.value.slice(1): one UTF-16 unit, the space after the check.
				head.Value = sliceOneUnit(head.Value)
				if head.Value == "" {
					node.Children = node.Children[1:]
				} else {
					head.Position.Start.Column++
					// One unit is one byte here: the sliced character is the space after `]`.
					head.Position.Start.Offset++
					start := head.Position.Start
					node.Position.Start = start
				}
			}
		}
	}

	context.exit(token, nil)
}

// sliceOneUnit is JavaScript's value.slice(1) for a value whose first character is ASCII, which the
// space after a task list check always is.
func sliceOneUnit(value string) string {
	if value == "" {
		return value
	}
	return value[1:]
}

// mathFromMarkdown is mdast-util-math's. The hast data it attaches is for HTML output and not kept.
func mathFromMarkdown() extension {
	exitMathData := func(context *compileContext, token *micromark.Token) {
		context.config.enter["data"](context, token)
		context.config.exit["data"](context, token)
	}
	return extension{
		enter: map[string]handle{
			"mathFlow": func(context *compileContext, token *micromark.Token) {
				context.enter(context.nodes.New(Node{NodeType: "math", IsLiteral: true}), token, nil)
			},
			"mathFlowFenceMeta": buffer,
			"mathText": func(context *compileContext, token *micromark.Token) {
				context.enter(context.nodes.New(Node{NodeType: "inlineMath", IsLiteral: true}), token, nil)
				buffer(context, token)
			},
		},
		exit: map[string]handle{
			"mathFlow": func(context *compileContext, token *micromark.Token) {
				data := trimOneLineEnding(context.resume(), true, true)
				node := context.top()
				context.exit(token, nil)
				node.Value = data
				context.data.mathFlowInside = false
			},
			"mathFlowFence": func(context *compileContext, token *micromark.Token) {
				if context.data.mathFlowInside {
					return
				}
				buffer(context, token)
				context.data.mathFlowInside = true
			},
			"mathFlowFenceMeta": func(context *compileContext, _ *micromark.Token) {
				data := context.resume()
				context.top().Meta = &data
			},
			"mathFlowValue": exitMathData,
			"mathText": func(context *compileContext, token *micromark.Token) {
				data := context.resume()
				node := context.top()
				context.exit(token, nil)
				node.Value = data
			},
			"mathTextData": exitMathData,
		},
	}
}

// wikiLinkFromMarkdown is @braindb/mdast-util-wiki-link's fromMarkdown. The data it builds for HTML is
// not kept; the printer reads only the value.
func wikiLinkFromMarkdown() extension {
	return extension{
		enter: map[string]handle{
			"wikiLink": func(context *compileContext, token *micromark.Token) {
				context.enter(context.nodes.New(Node{NodeType: "wikiLink", IsLiteral: true, ValueNull: true}), token, nil)
			},
		},
		exit: map[string]handle{
			"wikiLinkTarget": func(context *compileContext, token *micromark.Token) {
				node := context.top()
				node.Value = context.sliceSerialize(token)
				node.ValueNull = false
			},
			"wikiLinkAlias": func(*compileContext, *micromark.Token) {},
			"wikiLink":      func(context *compileContext, token *micromark.Token) { context.exit(token, nil) },
		},
	}
}

// liquidFromMarkdown is the fork's, src/language-markdown/parse/micromark/micromark-extension-liquid.js.
func liquidFromMarkdown() extension {
	return extension{
		canContainEols: []string{"liquidNode"},
		enter: map[string]handle{
			"liquidNode": func(context *compileContext, token *micromark.Token) {
				context.enter(context.nodes.New(Node{NodeType: "liquidNode"}), token, nil)
				buffer(context, token)
			},
		},
		exit: map[string]handle{
			"liquidNode": func(context *compileContext, token *micromark.Token) {
				context.resume()
				node := context.top()
				node.Value = context.sliceSerialize(token)
				node.IsLiteral = true
				context.exit(token, nil)
			},
		},
	}
}
