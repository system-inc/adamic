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
	some := func(value any, name any, index int, hasIndex bool) bool {
		_, node := value.(*Node)
		_, property := name.(string)
		return node && (name == nil || property) && (!hasIndex || index >= -1)
	}
	never := func(value any, name any, index int, hasIndex bool) bool {
		_, node := value.(*Node)
		_, property := name.(string)
		return node && property && hasIndex && index == -9
	}
	var snapshot func(*printing.AstPath[*Node]) string
	snapshot = func(path *printing.AstPath[*Node]) string {
		stack := []string{}
		owner := -1
		for _, value := range path.Stack() {
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
		return strings.Join([]string{strings.Join(stack, ","), frame(path.Value(), owner), key, fmt.Sprint(index), frame(path.Name(), owner),
			fmt.Sprint(number(path.Node())), fmt.Sprint(number(path.Parent())), fmt.Sprint(number(path.Grandparent())), fmt.Sprint(number(path.Root())),
			fmt.Sprint(number(path.Next())), fmt.Sprint(number(path.Previous())), fmt.Sprint(path.IsRoot()), fmt.Sprint(path.IsInArray()), fmt.Sprint(path.IsFirst()), fmt.Sprint(path.IsLast()),
			strings.Join(ancestors, ","), strings.Join(nodes, ","), fmt.Sprint(number(path.GetParentNode(1))), fmt.Sprint(number(path.FindAncestor(even))), fmt.Sprint(path.HasAncestor(even)),
			fmt.Sprint(path.Match()), fmt.Sprint(path.Match(nil)), fmt.Sprint(path.Match(some, nil)), fmt.Sprint(path.Match(some, some)), fmt.Sprint(path.Match(never))}, "|")
	}
	identity := func(path *printing.AstPath[*Node], i int, _ any) int {
		node, _ := path.Node()
		owner := -1
		for _, value := range path.Stack() {
			if n, ok := value.(*Node); ok && n != nil && n != node {
				owner = ids[n]
			}
		}
		return (ids[node]+1)*1000000 + (i+1)*1000 + owner
	}
	observe := func(path *printing.AstPath[*Node]) string {
		rows := []string{snapshot(path), printing.Call(path, snapshot, "children"), printing.Call(path, snapshot, "children", -1), printing.Call(path, snapshot, "children", 999999), printing.Call(path, snapshot, "absent")}
		for count := 0; count < min(3, len(path.Ancestors())); count++ {
			rows = append(rows, printing.CallParent(path, snapshot, count), fmt.Sprint(printing.CallParent(path, func(p *printing.AstPath[*Node]) int { return number(p.Node()) }, count)), fmt.Sprint(printing.CallParent(path, func(p *printing.AstPath[*Node]) bool { return p.IsFirst() }, count)))
		}
		rows = append(rows, fmt.Sprint(printing.Call(path, func(p *printing.AstPath[*Node]) int { return number(p.Node()) }, "children", 0)), snapshot(path))
		values := printing.Map(path, identity, "children")
		text := []string{}
		for _, v := range values {
			text = append(text, fmt.Sprint(v))
		}
		rows = append(rows, strings.Join(text, ","))
		text = nil
		path.Each(func(p *printing.AstPath[*Node], i int, a any) { text = append(text, fmt.Sprint(identity(p, i, a))) }, "children")
		rows = append(rows, strings.Join(text, ","), snapshot(path))
		return strings.Join(rows, "\n")
	}
	rows := []string{}
	path := printing.NewAstPath(root)
	var walk func(*printing.AstPath[*Node])
	walk = func(p *printing.AstPath[*Node]) {
		node, _ := p.Node()
		if len(node.Children) > 0 || p.IsRoot() {
			rows = append(rows, observe(p))
		} else {
			rows = append(rows, snapshot(p))
		}
		p.Each(func(p *printing.AstPath[*Node], _ int, _ any) { walk(p) }, "children")
	}
	walk(path)
	return strings.Join(rows, "\n")
}

func AdamicPathKeyWitness() string {
	path := printing.NewAstPath(&Node{NodeType: "root", IsParent: true, Children: []*Node{}})
	return printing.Call(path, func(p *printing.AstPath[*Node]) string {
		key, present := p.Key()
		return fmt.Sprintf("key=%q present=%t\n", key, present)
	}, "absent", "children", 0)
}
