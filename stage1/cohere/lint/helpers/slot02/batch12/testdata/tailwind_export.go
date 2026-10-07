package tailwind

import (
	"errors"
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
	"strings"
)

type Slot12DeclineCase struct {
	Node                  int
	Rule, Description     string
	Context, System, Mode int
}

var slot12Sources = []*ast.Node{}

func Slot12Source(n *ast.Node) { slot12Sources = append(slot12Sources, n) }

var slot12DeclineTrace string
var slot12Active int
var slot12Listeners []rule.Listeners
var slot12Reenter, slot12Entered bool

func Slot12Describe(name string, r DesignSystemResult) string {
	slot12DeclineTrace += fmt.Sprintf("describe:%s:%s;", name, r.EntryPoint)
	if slot12Reenter && !slot12Entered {
		slot12Entered = true
		slot12Listeners[slot12Active][ast.KindSourceFile](slot12Sources[0])
	}
	return DesignSystemDeclineMessage(name, r)
}
func Slot12Report(ctx rule.Context, n *ast.Node, m rule.Message) {
	nid, cid := -1, -1
	for i, p := range slot12Sources {
		if p == n {
			nid = i
		}
		if p.AsSourceFile() == ctx.SourceFile {
			cid = i
		}
	}
	slot12DeclineTrace += fmt.Sprintf("report:%d:%d:%s:%s;", cid, nid, m.Id, m.Description)
}
func Slot12Decline(names []string) ([]Slot12DeclineCase, string) {
	cases := []Slot12DeclineCase{}
	var want strings.Builder
	for i, n := range slot12Sources {
		for mode := 0; mode < 2; mode++ {
			slot12DeclineTrace = ""
			slot12Reenter = mode == 1
			slot12Entered = false
			slot12Listeners = nil
			for factory := 0; factory < 2; factory++ {
				ctx := rule.Context{SourceFile: n.AsSourceFile()}
				entry := fmt.Sprintf("entry:%d:%d", i, factory)
				r := DesignSystemResult{EntryPoint: entry, Err: errors.New("fixture é😀 failure")}
				name := names[i%len(names)]
				listeners := declineListeners(ctx, name, r)
				if len(listeners) != 1 || listeners[ast.KindSourceFile] == nil {
					panic("wrong subscription")
				}
				slot12Listeners = append(slot12Listeners, listeners)
				cases = append(cases, Slot12DeclineCase{Node: i, Rule: name, Description: DesignSystemDeclineMessage(name, r), Context: i, System: factory, Mode: mode})
			}
			for factory := 0; factory < 2; factory++ {
				slot12Active = factory
				slot12Listeners[factory][ast.KindSourceFile](n)
				slot12Listeners[factory][ast.KindSourceFile](n)
			}
			want.WriteString("decline:" + slot12DeclineTrace + "\n")
		}
	}
	return cases, want.String()
}
