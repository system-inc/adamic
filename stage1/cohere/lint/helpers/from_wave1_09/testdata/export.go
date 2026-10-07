package tailwind

// Oracle-only export of the unchanged Go helper.
func AdamicStaticNodes(input []StaticDeclaration) []*Node {
	return nodesFromStaticDeclarations(input)
}

func AdamicPropertySort(nodes []*Node) (Sort, []string) {
	return propertySort(nodes)
}

func AdamicValueSeparators() []int {
	values := []int{}
	for value := 0; value < 256; value++ {
		if isValueSeparator(byte(value)) {
			values = append(values, value)
		}
	}
	return values
}
