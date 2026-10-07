package tailwind

// Oracle-only export of the unchanged Go helper.
func AdamicStaticNodes(input []StaticDeclaration) []*Node {
	return nodesFromStaticDeclarations(input)
}
