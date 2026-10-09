package micromark

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf16"
)

// AdamicPrefixFixture exposes actual private preprocessing and construct attempts.
func AdamicPrefixFixture(text string) string {
	chunks := preprocess(SourceUnits(text))
	encode := strings.NewReplacer(`\`, `\\`, "\n", `\n`, "\r", `\r`, "\t", `\t`)
	fields := []string{}
	for _, chunk := range chunks {
		if chunk.IsText {
			fields = append(fields, "t:"+encode.Replace(string(utf16.Decode(chunk.Text))))
		} else {
			fields = append(fields, "c:"+strconv.Itoa(int(chunk.Code)))
		}
	}
	result := encode.Replace(strings.Join(fields, "\t"))
	for mode := 0; mode < 8; mode++ {
		continuation, initialOpen, unlimited := mode&1 != 0, mode&2 != 0, mode&4 != 0
		constructs := *MarkdownConstructs()
		if unlimited {
			constructs.Disable = append(append([]string(nil), constructs.Disable...), "codeIndented")
		}
		parser := &ParseContext{Constructs: &constructs, Lazy: map[int]bool{}}
		var matched bool
		var point Point
		finalOpen := initialOpen
		initial := &InitialConstruct{Tokenize: func(self *Self, effects *Effects) State {
			self.ContainerState.open = initialOpen
			var drain State
			drain = func(code Code) State {
				effects.Consume(code)
				if code == CodeEof {
					return nil
				}
				return drain
			}
			finish := func(success bool) State {
				return func(code Code) State {
					matched = success
					point = self.Now()
					finalOpen = self.ContainerState.open
					return drain(code)
				}
			}
			construct := blockQuote
			if continuation {
				construct = blockQuote.Continuation
			}
			return effects.Attempt(construct, finish(true), finish(false))
		}}
		context := createTokenizer(parser, initial, nil)
		context.Write(chunks)
		consumed := 0
		for i := 0; i < point.index; i++ {
			if chunks[i].IsText {
				consumed += len(chunks[i].Text)
			} else {
				consumed++
			}
		}
		if point.bufferIndex > 0 {
			consumed += point.bufferIndex
		}
		flag, opened := 0, 0
		if matched {
			flag = 1
		}
		if finalOpen {
			opened = 1
		}
		result += fmt.Sprintf("\t%d,%d,%d,%d,%d", flag, consumed, point.Offset, point.Column, opened)
	}
	return result
}
