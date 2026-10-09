package javascript

// utilities/function-parameters.js.

// getFunctionParameters is upstream's getFunctionParameters. Upstream memoizes it in a WeakMap; no
// caller mutates the array or compares it by identity, and a cache on settings would need options,
// which upstream's callers (canAttachComment, isTestCall, shouldHugTheOnlyFunctionParameter) do not
// have, so the Go computes it each time and keeps upstream's one-argument signature.
func getFunctionParameters(node Node) []Node {
	var parameters []Node
	if node.Truthy("this") {
		parameters = append(parameters, node.Child("this"))
	}

	parameters = append(parameters, node.List("params")...)

	if node.Truthy("rest") {
		parameters = append(parameters, node.Child("rest"))
	}
	return parameters
}

// iterateFunctionParametersPath is upstream's iterateFunctionParametersPath.
func iterateFunctionParametersPath(path *Path, iteratee func(path *Path, index int)) {
	node := node(path)
	index := 0
	callback := func(path *Path) any {
		iteratee(path, index)
		index++
		return nil
	}
	if node.Truthy("this") {
		call(path, callback, "this")
	}

	each(path, func(path *Path, _ int) { callback(path) }, "params")

	if node.Truthy("rest") {
		call(path, callback, "rest")
	}
}

// hasRestParameter is upstream's hasRestParameter.
func hasRestParameter(node Node) bool {
	if node.Truthy("rest") {
		return true
	}
	parameters := getFunctionParameters(node)
	return len(parameters) > 0 && parameters[len(parameters)-1].Is("RestElement")
}
