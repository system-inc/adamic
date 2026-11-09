package markdown

import (
	"fmt"
	"github.com/system-inc/cohere/internal/format/arena"
	"github.com/system-inc/cohere/internal/format/markdown/mdast"
	"github.com/system-inc/cohere/internal/format/printing"
	"strings"
)

func AdamicPathFixture(input string) (string, string, error) {
	source := strings.TrimPrefix(strings.ReplaceAll(strings.ReplaceAll(input, "\r\n", "\n"), "\r", "\n"), "\ufeff")
	memory := &arena.Arena[Node]{}
	root, err := mdast.ParseMarkdown(source, memory)
	if err != nil {
		return "", "", err
	}
	root = preprocess(root, source, 4, memory)
	return "A\t\t4\n" + adamicAstSerialize(root, source) + "\nF\n", adamicPathObserve(root), nil
}
func adamicPathObserve(root *Node) string {
	ids := map[*Node]int{}
	var index func(*Node)
	index = func(node *Node) {
		ids[node] = len(ids)
		for _, child := range node.Children {
			index(child)
		}
	}
	index(root)
	number := func(node *Node, ok bool) int {
		if !ok || node == nil {
			return -1
		}
		return ids[node]
	}
	frame := func(value any, owner int) string {
		switch v := value.(type) {
		case *Node:
			if v == nil {
				return "U"
			}
			return fmt.Sprint("N", ids[v])
		case []*Node:
			return fmt.Sprint("A", owner)
		case int:
			return fmt.Sprint("I", v)
		case string:
			return "P" + v
		default:
			return "U"
		}
	}
	even := func(node *Node) bool { return ids[node]%2 == 0 }
	some := func(value printing.Value[*Node], name string, index int, hasIndex bool) bool {
		_, node := value.Node()
		return node && (!hasIndex || index >= -1)
	}
	never := func(value printing.Value[*Node], name string, index int, hasIndex bool) bool {
		_, node := value.Node()
		return node && name != "" && hasIndex && index == -9
	}
	var snapshot func(*astPath) string
	snapshot = func(path *astPath) string {
		stack := []string{}
		owner := -1
		for _, value := range adamicPathStack(path) {
			if node, ok := value.(*Node); ok && node != nil {
				owner = ids[node]
			}
			stack = append(stack, frame(value, owner))
		}
		key, _ := path.Key()
		index, present := path.Index()
		if !present {
			index = -1
		}
		ancestors := []string{}
		for _, node := range path.Ancestors() {
			ancestors = append(ancestors, fmt.Sprint(ids[node]))
		}
		nodes := []string{}
		for count := 0; count < 5; count++ {
			nodes = append(nodes, fmt.Sprint(number(path.GetNode(count))))
		}
		return strings.Join([]string{strings.Join(stack, ","), frame(adamicPathValue(path.Value()), owner), key, fmt.Sprint(index), frame(adamicPathName(path), owner),
			fmt.Sprint(number(path.Node())), fmt.Sprint(number(path.Parent())), fmt.Sprint(number(path.Grandparent())), fmt.Sprint(number(path.Root())),
			fmt.Sprint(number(path.Next())), fmt.Sprint(number(path.Previous())), fmt.Sprint(path.IsRoot()), fmt.Sprint(path.IsInArray()), fmt.Sprint(path.IsFirst()), fmt.Sprint(path.IsLast()),
			strings.Join(ancestors, ","), strings.Join(nodes, ","), fmt.Sprint(number(path.GetParentNode(1))), fmt.Sprint(number(path.FindAncestor(even))), fmt.Sprint(path.HasAncestor(even)),
			fmt.Sprint(path.Match()), fmt.Sprint(path.Match(nil)), fmt.Sprint(path.Match(some, nil)), fmt.Sprint(path.Match(some, some)), fmt.Sprint(path.Match(never))}, "|")
	}
	identity := func(path *astPath, i int) int {
		node, _ := path.Node()
		owner := -1
		for _, value := range adamicPathStack(path) {
			if n, ok := value.(*Node); ok && n != nil && n != node {
				owner = ids[n]
			}
		}
		return (ids[node]+1)*1000000 + (i+1)*1000 + owner
	}
	observe := func(path *astPath) string {
		rows := []string{snapshot(path), printing.Call(path, snapshot, "children"), printing.CallElement(path, snapshot, -1, "children"), printing.CallElement(path, snapshot, 999999, "children"), printing.Call(path, snapshot, "absent")}
		for count := 0; count < min(3, len(path.Ancestors())); count++ {
			rows = append(rows, printing.CallParent(path, snapshot, count), fmt.Sprint(printing.CallParent(path, func(p *astPath) int { return number(p.Node()) }, count)), fmt.Sprint(printing.CallParent(path, func(p *astPath) bool { return p.IsFirst() }, count)))
		}
		rows = append(rows, fmt.Sprint(printing.CallElement(path, func(p *astPath) int { return number(p.Node()) }, 0, "children")), snapshot(path))
		values := printing.Map(path, identity, "children")
		text := []string{}
		for _, v := range values {
			text = append(text, fmt.Sprint(v))
		}
		rows = append(rows, strings.Join(text, ","))
		text = nil
		path.Each(func(p *astPath, i int) { text = append(text, fmt.Sprint(identity(p, i))) }, "children")
		rows = append(rows, strings.Join(text, ","), snapshot(path))
		return strings.Join(rows, "\n")
	}
	rows := []string{}
	path := printing.NewAstPath(root, (*settings)(nil))
	var walk func(*astPath)
	walk = func(p *astPath) {
		node, _ := p.Node()
		if len(node.Children) > 0 || p.IsRoot() {
			rows = append(rows, observe(p))
		} else {
			rows = append(rows, snapshot(p))
		}
		p.Each(func(p *astPath, _ int) { walk(p) }, "children")
	}
	walk(path)
	return strings.Join(rows, "\n")
}

func AdamicPathKeyWitness() string {
	path := printing.NewAstPath(&Node{NodeType: "root", IsParent: true, Children: []*Node{}}, (*settings)(nil))
	return printing.CallElement(path, func(p *astPath) string {
		key, present := p.Key()
		return fmt.Sprintf("key=%q present=%t\n", key, present)
	}, 0, "absent", "children")
}

// Keep the transport's alternating name/value stack while cohere stores typed frames.
func adamicPathValue(value printing.Value[*Node]) any {
	if node, ok := value.Node(); ok {
		return node
	}
	if nodes, ok := value.Nodes(); ok {
		return nodes
	}
	if leaf, ok := value.Leaf(); ok {
		return leaf
	}
	return nil
}
func adamicPathFrameName(name printing.PathName) any {
	if name.IsIndex {
		return name.Index
	}
	return name.Property
}
func adamicPathStack(path *astPath) []any {
	var stack []any
	for index, entry := range path.Stack() {
		if index > 0 {
			stack = append(stack, adamicPathFrameName(entry.Name))
		}
		stack = append(stack, adamicPathValue(entry.Value))
	}
	return stack
}
func adamicPathName(path *astPath) any {
	stack := path.Stack()
	if len(stack) == 1 {
		return nil
	}
	return adamicPathFrameName(stack[len(stack)-1].Name)
}
