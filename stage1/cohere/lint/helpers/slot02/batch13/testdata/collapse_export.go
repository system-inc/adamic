package tailwind

import (
	"fmt"
	"reflect"
	"strings"
)

type Slot13Argument struct {
	Kind, Value string
	Nodes       int
}
type Slot13ValueCase struct {
	State, ValueID, Mode, ParseNodes int
	Kind, Value                      string
	Arguments                        []Slot13Argument
}

var Slot13ValueWant []string
var slot13ValueTrace strings.Builder
var slot13StateID, slot13ValueMode int
var slot13ValuePointer *ParsedValue
var slot13Slices = map[uintptr]int{}

func slot13Slice(n []ValueNode) int {
	if n == nil {
		return -1
	}
	p := reflect.ValueOf(n).Pointer()
	if id, ok := slot13Slices[p]; ok {
		return id
	}
	panic("slice identity")
}
func Slot13Parse(s string) []ValueNode {
	fmt.Fprintf(&slot13ValueTrace, "parse:%s;", s)
	r := ParseValue(s)
	if r != nil {
		slot13Slices[reflect.ValueOf(r).Pointer()] = 500
	}
	return r
}
func Slot13Theme(s *utilityEvaluation, a, v string) (string, bool) {
	fmt.Fprintf(&slot13ValueTrace, "theme:%d:%s:%s;", slot13StateID, a, v)
	return "theme é😀", slot13ValueMode != 0
}
func Slot13Bare(s *utilityEvaluation, a string, v *ParsedValue) ([]ValueNode, bool, bool) {
	same := v == slot13ValuePointer
	fmt.Fprintf(&slot13ValueTrace, "bare:%d:%s:%t;", slot13StateID, a, same)
	r := []ValueNode{{Kind: ValueNodeKindWord, Value: "bare"}}
	slot13Slices[reflect.ValueOf(r).Pointer()] = 200
	return r, slot13ValueMode == 2, slot13ValueMode != 0
}
func Slot13Arbitrary(s *utilityEvaluation, a string, v *ParsedValue) ([]ValueNode, bool) {
	fmt.Fprintf(&slot13ValueTrace, "arbitrary:%d:%s:%t;", slot13StateID, a, v == slot13ValuePointer)
	r := []ValueNode{{Kind: ValueNodeKindWord, Value: "arbitrary"}}
	slot13Slices[reflect.ValueOf(r).Pointer()] = 300
	return r, slot13ValueMode != 0
}
func Slot13Values(texts []string) ([]Slot13ValueCase, string) {
	functions := []ValueNode{}
	var walk func([]ValueNode)
	walk = func(ns []ValueNode) {
		for _, n := range ns {
			if n.Kind == ValueNodeKindFunction && (n.Value == "--value" || n.Value == "--modifier") {
				functions = append(functions, n)
			}
			walk(n.Nodes)
		}
	}
	controls := []string{"--value(--default())", "--value(--default(a), --default(b))", "--value('red', --hit-*, integer, [length])", "--value(\"red\", ratio)", "--value('red, integer)", "--value(\", '')", "--value([x], [])", "--value(--missing-*, number)", "--value()", "--value('redx, integer)"}
	for _, s := range append(texts, controls...) {
		walk(ParseValue(s))
	}
	functions = append(functions, ValueNode{Kind: ValueNodeKindFunction, Nodes: []ValueNode{{Kind: ValueNodeKindFunction, Value: "--default", Nodes: nil}}})
	cases := []Slot13ValueCase{}
	var want strings.Builder
	for _, fn := range functions {
		for _, k := range []string{"nil", "named", "arbitrary", "other"} {
			for _, v := range []string{"red", "", "é😀"} {
				for mode := 0; mode < 3; mode++ {
					for state := 0; state < 2; state++ {
						slot13Slices = map[uintptr]int{}
						args := []Slot13Argument{}
						for i, a := range fn.Nodes {
							h := -1
							if a.Nodes != nil {
								p := reflect.ValueOf(a.Nodes).Pointer()
								if old, ok := slot13Slices[p]; ok {
									h = old
								} else {
									h = 100 + i
									slot13Slices[p] = h
								}
							}
							args = append(args, Slot13Argument{string(a.Kind), a.Value, h})
						}
						var value *ParsedValue
						if k != "nil" {
							value = &ParsedValue{Kind: ParsedValueKind(k), Value: v}
						}
						slot13ValuePointer = value
						slot13ValueMode = mode
						slot13StateID = state
						slot13ValueTrace.Reset()
						r, ratio, ok := (&utilityEvaluation{}).resolveValueFunction(value, &fn)
						parseNodes := -1
						if strings.Contains(slot13ValueTrace.String(), "parse:") && r != nil {
							parseNodes = 500
						}
						valueID := -1
						if value != nil {
							valueID = state + 10
						}
						cases = append(cases, Slot13ValueCase{state, valueID, mode, parseNodes, k, v, args})
						line := fmt.Sprintf("value:%s:%d:%t:%t", slot13ValueTrace.String(), slot13Slice(r), ratio, ok)
						want.WriteString(line + "\n")
						Slot13ValueWant = append(Slot13ValueWant, line)
					}
				}
			}
		}
	}
	return cases, want.String()
}
