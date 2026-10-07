package micromark

import (
	"fmt"
	"strconv"
	"strings"
)

func adamicEventEscape(s string) string {
	return strings.NewReplacer("\\", "\\\\", "\n", "\\n", "\r", "\\r", "\t", "\\t").Replace(s)
}
func adamicPoint(p Point) string {
	return fmt.Sprintf("%d,%d,%d,%d,%d", p.Line, p.Column, p.Offset, p.index, p.bufferIndex)
}
func AdamicEventTransport(events []Event) string {
	var b strings.Builder
	tokens := map[*Token]int{}
	contexts := map[*TokenizeContext]int{}
	for _, e := range events {
		if _, ok := contexts[e.Context]; !ok {
			contexts[e.Context] = len(contexts)
			parts := []string{}
			for _, c := range e.Context.chunks {
				if c.IsText {
					v := []string{}
					for _, u := range c.Text {
						v = append(v, fmt.Sprint(u))
					}
					parts = append(parts, "T"+strings.Join(v, ","))
				} else {
					parts = append(parts, fmt.Sprint("C", int(c.Code)))
				}
			}
			fmt.Fprintf(&b, "C\t%s\n", strings.Join(parts, ";"))
		}
	}
	for _, e := range events {
		if _, ok := tokens[e.Token]; !ok {
			t := e.Token
			tokens[t] = len(tokens)
			fmt.Fprintf(&b, "T\t%s\t%s\t%s\t%t\t%s\n", t.Type, adamicPoint(t.Start), adamicPoint(t.End), t.Spread, strings.Join(t.Align, ","))
		}
	}
	for _, e := range events {
		fmt.Fprintf(&b, "E\t%t\t%d\t%d\n", e.Enter, tokens[e.Token], contexts[e.Context])
	}
	return b.String()
}

type AdamicEventDocument struct {
	Source string
	Events []Event
}

func AdamicReadEventDocuments(text string) []AdamicEventDocument {
	number := func(s string) int {
		n, e := strconv.Atoi(s)
		if e != nil {
			panic(e)
		}
		return n
	}
	point := func(s string) Point {
		f := strings.Split(s, ",")
		return Point{Line: number(f[0]), Column: number(f[1]), Offset: number(f[2]), index: number(f[3]), bufferIndex: number(f[4])}
	}
	decode := func(s string) string {
		var b strings.Builder
		for i := 0; i < len(s); i++ {
			c := s[i]
			if c == '\\' && i+1 < len(s) {
				i++
				c = s[i]
				switch c {
				case 'n':
					c = '\n'
				case 'r':
					c = '\r'
				case 't':
					c = '\t'
				}
			}
			b.WriteByte(c)
		}
		return b.String()
	}
	documents := []AdamicEventDocument{}
	source := ""
	tokens := []*Token{}
	contexts := []*TokenizeContext{}
	events := []Event{}
	for _, line := range strings.Split(text, "\n") {
		f := strings.Split(line, "\t")
		switch f[0] {
		case "S":
			source = decode(f[1])
		case "C":
			chunks := []Chunk{}
			for _, v := range strings.Split(f[1], ";") {
				if v == "" {
					continue
				}
				if v[0] == 'C' {
					chunks = append(chunks, codeChunk(Code(number(v[1:]))))
				} else {
					units := []uint16{}
					if len(v) > 1 {
						for _, n := range strings.Split(v[1:], ",") {
							units = append(units, uint16(number(n)))
						}
					}
					chunks = append(chunks, textChunk(units))
				}
			}
			contexts = append(contexts, &TokenizeContext{chunks: chunks})
		case "T":
			align := []string{}
			if f[5] != "" {
				align = strings.Split(f[5], ",")
			}
			tokens = append(tokens, &Token{Type: f[1], Start: point(f[2]), End: point(f[3]), Spread: f[4] == "true", Align: align})
		case "E":
			events = append(events, Event{Enter: f[1] == "true", Token: tokens[number(f[2])], Context: contexts[number(f[3])]})
		case "F":
			documents = append(documents, AdamicEventDocument{Source: source, Events: events})
			source = ""
			tokens = nil
			contexts = nil
			events = nil
		case "":
		default:
			panic("event document protocol")
		}
	}
	return documents
}
