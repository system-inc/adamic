package micromark

// liquid is the Prettier fork's liquidSyntax, src/language-markdown/parse/micromark/
// micromark-extension-liquid.js: a `{{ ... }}` or `{% ... %}` span in text, one liquidNode holding data
// and line endings. The mdast side, liquidFromMarkdown, lives in internal/format/markdown/mdast.

// typeLiquidNode is upstream's file-local `nodeType`.
const typeLiquidNode = "liquidNode"

// liquid is the construct upstream's liquidSyntax builds inline, under `{` in text.
var liquid = &Construct{Name: "liquid", Tokenize: tokenizeLiquid}

// liquidExtension is upstream's liquidSyntax().
func liquidExtension() *Extension {
	return &Extension{
		Text: &ConstructRecord{ByCode: map[Code][]*Construct{
			CodeLeftCurlyBrace: {liquid},
		}},
	}
}

func tokenizeLiquid(self *Self, effects *Effects, ok State, nok State) State {
	// closingCode is `}` after `{{` and `%` after `{%`. Upstream's starts undefined; it is always set
	// before inside reads it.
	var closingCode Code
	var start, inside, mayClose State

	start = func(code Code) State {
		effects.Enter(typeLiquidNode, nil)
		effects.Enter(TypeData, nil)
		effects.Consume(code)
		return func(code Code) State {
			switch code {
			case CodePercentSign, CodeLeftCurlyBrace:
				if code == CodePercentSign {
					closingCode = CodePercentSign
				} else {
					closingCode = CodeRightCurlyBrace
				}
				effects.Consume(code)
				return inside
			default:
				return nok(code)
			}
		}
	}

	// Upstream's switch tests closingCode before eof, so the order of the cases is kept.
	inside = func(code Code) State {
		switch code {
		case closingCode:
			effects.Consume(code)
			return mayClose
		case CodeEof:
			return nok(code)
		default:
			if markdownLineEnding(code) {
				effects.Exit(TypeData)
				effects.Enter(TypeLineEnding, nil)
				effects.Consume(code)
				effects.Exit(TypeLineEnding)
				effects.Enter(TypeData, nil)
				return inside
			}
			effects.Consume(code)
			return inside
		}
	}

	// mayClose returns inside without consuming when the code is not `}`, so the tokenizer feeds the same
	// code again, to inside. That leans on the production build's missing consume assertion, which the
	// Go core also leaves out (tokenizer.go, goCode). It means `{{ a }x }}` and `{% a %% }` stay open past
	// the first closing code, and `{{ a }\n}}` takes the line ending into the node.
	mayClose = func(code Code) State {
		if code == CodeRightCurlyBrace {
			effects.Consume(code)
			effects.Exit(TypeData)
			effects.Exit(typeLiquidNode)
			return ok
		}

		return inside
	}

	return start
}
