package tailwind

// Oracle-only export of the unchanged Go helper.
func AdamicStaticNodes(input []StaticDeclaration) []*Node {
	return nodesFromStaticDeclarations(input)
}

func AdamicPropertySort(nodes []*Node) (Sort, []string) {
	return propertySort(nodes)
}
