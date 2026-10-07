package tailwind

func AdamicWave6Definition(source string, bound bool, mode string) ([][]any, []string) {
	nodes := []*Node{
		{Kind: KindRule, Value: source, ValuePresent: true},
		{Kind: KindDeclaration, Value: source, ValuePresent: true},
		{Kind: KindComment, Value: source, ValuePresent: true},
		{Kind: KindDeclaration, Value: "calc(--modifier(--text --line-height) * 2)", ValuePresent: true},
		{Kind: KindDeclaration, Value: source, ValuePresent: false},
		{Kind: KindDeclaration, Value: "", ValuePresent: true},
	}
	nodes[0].Nodes = []*Node{nodes[1], nodes[2]}
	nodes[1].Nodes = []*Node{nodes[3]}
	nodes[2].Nodes = []*Node{nodes[4]}
	children := [][]int{{1, 2}, {3}, {4}, {}, {}, {}}
	rows := make([][]any, len(nodes))
	for index, node := range nodes {
		single := &Node{Kind: KindDeclaration, Value: node.Value, ValuePresent: true}
		normalizeValueFunctionArguments([]*Node{single})
		rows[index] = []any{string(node.Kind), node.ValuePresent, node.Value, single.Value, children[index]}
	}
	roots := []*Node{nodes[0], nodes[5]}
	normalizeValueFunctionArguments(roots)
	for index, node := range nodes {
		rows[index] = append(rows[index], node.Value)
		node.Value = rows[index][2].(string)
	}
	if bound {
		normalizeUtilityDefinition(&UtilityDefinition{Nodes: roots})
	} else {
		normalizeUtilityDefinition(nil)
	}
	actual := make([]string, len(nodes))
	for index, node := range nodes {
		actual[index] = node.Value
	}
	return rows, actual
}
