package micromark

import (
	"fmt"
	"strings"
)

func AdamicKernelEvents(units []uint16) string {
	chunks := preprocess(units)
	before := &Construct{Name: "before"}
	changed := &Construct{Name: "changed"}
	fields := &Token{Spread: true, Align: []string{"left"}}
	check := &Construct{Partial: true, Tokenize: func(self *Self, e *Effects, ok, nok State) State {
		e.Enter("rollback", nil)
		self.CurrentConstruct = changed
		return func(code Code) State { e.Consume(code); e.Exit("rollback"); return ok }
	}}
	initial := &InitialConstruct{Tokenize: func(self *Self, e *Effects) State {
		self.DefineSkip(Point{Line: 2, Column: 3})
		self.DefineSkip(Point{Line: 3, Column: 5})
		self.CurrentConstruct = before
		e.Enter("document", nil)
		data := false
		var state, after State
		after = func(code Code) State {
			if code <= 0 && data {
				e.Exit("data")
				data = false
			}
			if code < 0 {
				e.Enter("virtual", fields)
				e.Consume(code)
				e.Exit("virtual")
			} else if code > 0 {
				if !data {
					e.Enter("data", nil)
					data = true
				}
				e.Consume(code)
			} else {
				e.Exit("document")
				e.Consume(code)
			}
			return state
		}
		state = func(code Code) State {
			if code < 0 {
				return e.Check(check, after, after)(code)
			}
			return after(code)
		}
		return state
	}}
	ctx := createTokenizer(&ParseContext{Constructs: Combine(nil)}, initial, nil)
	events := ctx.Write(chunks)
	rows := []string{}
	for _, e := range events {
		t := e.Token
		text, expanded := "-", "-"
		if t.Start.index != t.End.index || (t.Start.bufferIndex >= 0 && t.End.bufferIndex >= 0) {
			text = ctx.SliceSerialize(t, false)
			expanded = ctx.SliceSerialize(t, true)
		}
		rows = append(rows, fmt.Sprintf("%t\t%s\t%s\t%s\t%t\t%s\t%s\t%s", e.Enter, t.Type, adamicPoint(t.Start), adamicPoint(t.End), t.Spread, strings.Join(t.Align, ","), adamicEventEscape(text), adamicEventEscape(expanded)))
	}
	construct := 9
	if ctx.CurrentConstruct == before {
		construct = 7
	}
	rows = append(rows, fmt.Sprintf("now\t%s\t%d\t%d\t%t\t%d", adamicPoint(ctx.Now()), ctx.Previous, construct, ctx.consumed, len(ctx.stack)))
	return adamicEventEscape(strings.Join(rows, "\n"))
}
