package unist

// Ported from yaml-unist-parser 3.2.0, dist/attach.mjs, dist/utils/define-parents.mjs and
// dist/utils/update-positions.mjs.

// nodeTableEntry is one line's entry of attach.mjs's node table.
type nodeTableEntry struct {
	comment                *Node
	leadingAttachableNode  *Node
	trailingAttachableNode *Node
	trailingNode           *Node
}

// attachComments is upstream's attachComments: every comment that no transform placed goes to the node
// that owns it, as a trailing, end or leading comment.
func (context *context) attachComments(root *Node) {
	defineParents(root, nil)
	nodeTable := context.createNodeTable(root)
	restDocuments := append([]*Node{}, root.Children...)
	// The filter runs whole before the first comment is attached.
	var unattached []*Node
	for _, comment := range root.Comments {
		if comment.Parent == nil {
			unattached = append(unattached, comment)
		}
	}
	for _, comment := range unattached {
		for len(restDocuments) > 1 && context.position(comment).start.line > context.position(restDocuments[0]).end.line {
			restDocuments = restDocuments[1:]
		}
		context.attachComment(comment, nodeTable, restDocuments[0])
	}
}

// createNodeTable is upstream's createNodeTable: for each line, its comment and the nodes a comment there
// could attach to.
func (context *context) createNodeTable(root *Node) []nodeTableEntry {
	nodeTable := make([]nodeTableEntry, context.position(root).end.line)
	for _, comment := range root.Comments {
		nodeTable[context.position(comment).start.line-1].comment = comment
	}
	context.initNodeTable(nodeTable, root)
	return nodeTable
}

// initNodeTable is upstream's initNodeTable.
func (context *context) initNodeTable(nodeTable []nodeTableEntry, node *Node) {
	position := context.position(node)
	if position.start.offset == position.end.offset {
		return
	}
	if hasLeadingCommentsField(node.NodeType) {
		start := position.start
		leadingAttachableNode := nodeTable[start.line-1].leadingAttachableNode
		if leadingAttachableNode == nil || start.column < context.position(leadingAttachableNode).start.column {
			nodeTable[start.line-1].leadingAttachableNode = node
		}
	}
	if hasTrailingCommentField(node.NodeType) && position.end.column > 1 && node.NodeType != "document" &&
		node.NodeType != "documentHead" {
		end := position.end
		trailingAttachableNode := nodeTable[end.line-1].trailingAttachableNode
		if trailingAttachableNode == nil || end.column >= context.position(trailingAttachableNode).end.column {
			nodeTable[end.line-1].trailingAttachableNode = node
		}
	}
	if node.NodeType != "root" && node.NodeType != "document" && node.NodeType != "documentHead" &&
		node.NodeType != "documentBody" {
		start, end := position.start, position.end
		// [end.line].concat(start.line === end.line ? [] : start.line)
		lines := []int{end.line}
		if start.line != end.line {
			lines = append(lines, start.line)
		}
		for _, line := range lines {
			currentEndNode := nodeTable[line-1].trailingNode
			if currentEndNode == nil || end.column >= context.position(currentEndNode).end.column {
				nodeTable[line-1].trailingNode = node
			}
		}
	}
	if hasChildrenField(node.NodeType) {
		for _, child := range node.Children {
			context.initNodeTable(nodeTable, child)
		}
	}
}

// attachComment is upstream's attachComment.
func (context *context) attachComment(comment *Node, nodeTable []nodeTableEntry, document *Node) {
	commentLine := context.position(comment).start.line
	trailingAttachableNode := nodeTable[commentLine-1].trailingAttachableNode
	if trailingAttachableNode != nil {
		if trailingAttachableNode.TrailingComment != nil {
			throwError("Unexpected multiple trailing comment at %s", getPointText(context.position(comment).start))
		}
		defineParents(comment, trailingAttachableNode)
		trailingAttachableNode.TrailingComment = comment
		return
	}
	for line := commentLine; line >= context.position(document).start.line; line-- {
		trailingNode := nodeTable[line-1].trailingNode
		var currentNode *Node
		if trailingNode == nil {
			// a:
			//   b:
			//    #b
			//  #a
			//
			// a:
			//   b:
			//  #a
			//    #a
			if line != commentLine && nodeTable[line-1].comment != nil {
				currentNode = nodeTable[line-1].comment.Parent
			} else {
				continue
			}
		} else {
			currentNode = trailingNode
		}
		if currentNode == nil {
			throwTypeError("undefined", "type")
		}
		if currentNode.NodeType == "sequence" || currentNode.NodeType == "mapping" {
			currentNode = currentNode.Children[0]
		}
		if currentNode.NodeType == "mappingItem" {
			mappingKey, mappingValue := currentNode.Children[0], currentNode.Children[1]
			if context.isExplicitMappingKey(mappingKey) {
				currentNode = mappingKey
			} else {
				currentNode = mappingValue
			}
		}
		for {
			if context.shouldOwnEndComment(currentNode, comment) {
				defineParents(comment, currentNode)
				currentNode.EndComments = append(currentNode.EndComments, comment)
				return
			}
			if currentNode.Parent == nil {
				break
			}
			currentNode = currentNode.Parent
		}
		break
	}
	for line := commentLine + 1; line <= context.position(document).end.line; line++ {
		leadingAttachableNode := nodeTable[line-1].leadingAttachableNode
		if leadingAttachableNode != nil {
			defineParents(comment, leadingAttachableNode)
			leadingAttachableNode.LeadingComments = append(leadingAttachableNode.LeadingComments, comment)
			return
		}
	}
	documentBody := document.Children[1]
	defineParents(comment, documentBody)
	documentBody.EndComments = append(documentBody.EndComments, comment)
}

// shouldOwnEndComment is upstream's shouldOwnEndComment.
func (context *context) shouldOwnEndComment(node *Node, comment *Node) bool {
	position, commentPosition := context.position(node), context.position(comment)
	if position.start.offset < commentPosition.start.offset && position.end.offset > commentPosition.end.offset {
		switch node.NodeType {
		case "flowMapping", "flowSequence":
			return len(node.Children) == 0 ||
				commentPosition.start.line > context.position(node.Children[len(node.Children)-1]).end.line
		}
	}
	if commentPosition.end.offset < position.end.offset {
		return false
	}
	switch node.NodeType {
	case "sequenceItem":
		return commentPosition.start.column > position.start.column
	case "mappingKey", "mappingValue":
		return commentPosition.start.column > context.position(node.Parent).start.column &&
			(len(node.Children) == 0 || len(node.Children) == 1 && node.Children[0].NodeType != "blockFolded" &&
				node.Children[0].NodeType != "blockLiteral") &&
			(node.NodeType == "mappingValue" || context.isExplicitMappingKey(node))
	default:
		return false
	}
}

// isExplicitMappingKey is upstream's isExplicitMappingKey. `node.position.start !== node.position.end`
// compares the point objects, as the pointers do here.
func (context *context) isExplicitMappingKey(node *Node) bool {
	position := context.position(node)
	return position.start != position.end &&
		(len(node.Children) == 0 || position.start.offset != context.position(node.Children[0]).start.offset)
}

// defineParents is upstream's defineParents: Parent for the node and everything under it. Upstream
// defines _parent as a non-writable property, so defining it again with a different parent throws; no
// path does, since attachComments only attaches comments that have none.
func defineParents(node *Node, parent *Node) {
	for _, child := range node.Children {
		defineParents(child, node)
	}
	if node.Anchor != nil {
		defineParents(node.Anchor, node)
	}
	if node.Tag != nil {
		defineParents(node.Tag, node)
	}
	for _, comment := range node.LeadingComments {
		defineParents(comment, node)
	}
	for _, comment := range node.MiddleComments {
		defineParents(comment, node)
	}
	if node.IndicatorComment != nil {
		defineParents(node.IndicatorComment, node)
	}
	if node.TrailingComment != nil {
		defineParents(node.TrailingComment, node)
	}
	for _, comment := range node.EndComments {
		defineParents(comment, node)
	}
	node.Parent = parent
}

// updatePositions is upstream's updatePositions: a node's position grows to cover its children, their
// leading comments, tags and anchors, the last child's trailing comment, and the node's end comments.
func (context *context) updatePositions(node *Node) {
	if node == nil || !hasChildrenField(node.NodeType) {
		return
	}
	children := node.Children
	for _, child := range children {
		context.updatePositions(child)
	}
	if node.NodeType == "document" {
		head, body := context.position(children[0]), context.position(children[1])
		if head.start.offset == head.end.offset {
			// head.position.start = head.position.end = body.position.start
			head.end = body.start
			head.start = body.start
		} else if body.start.offset == body.end.offset {
			body.end = head.end
			body.start = head.end
		}
	}
	host := context.position(node)
	updateStartPoint := createUpdater(host, startPointGetter, startPointSetter, shouldUpdateStartPoint)
	updateEndPoint := createUpdater(host, endPointGetter, endPointSetter, shouldUpdateEndPoint)
	if len(node.EndComments) != 0 {
		updateStartPoint(context.position(node.EndComments[0]).start)
		updateEndPoint(context.position(node.EndComments[len(node.EndComments)-1]).end)
	}
	var nonNullChildren []*Node
	for _, child := range children {
		if child != nil {
			nonNullChildren = append(nonNullChildren, child)
		}
	}
	if len(nonNullChildren) != 0 {
		firstChild := nonNullChildren[0]
		lastChild := nonNullChildren[len(nonNullChildren)-1]
		updateStartPoint(context.position(firstChild).start)
		updateEndPoint(context.position(lastChild).end)
		if len(firstChild.LeadingComments) != 0 {
			updateStartPoint(context.position(firstChild.LeadingComments[0]).start)
		}
		if firstChild.Tag != nil {
			updateStartPoint(context.position(firstChild.Tag).start)
		}
		if firstChild.Anchor != nil {
			updateStartPoint(context.position(firstChild.Anchor).start)
		}
		if lastChild.TrailingComment != nil {
			updateEndPoint(context.position(lastChild.TrailingComment).end)
		}
	}
}

// createUpdater is utils/create-updater.mjs: reduced is read from the host once, when the updater is
// made.
func createUpdater(host *position, getter func(*position) *point, setter func(*position, *point), shouldUpdate func(reduced *point, value *point) bool) func(value *point) {
	reduced := getter(host)
	return func(value *point) {
		if shouldUpdate(reduced, value) {
			reduced = value
			setter(host, value)
		}
	}
}

func startPointGetter(position *position) *point               { return position.start }
func startPointSetter(position *position, at *point)           { position.start = at }
func endPointGetter(position *position) *point                 { return position.end }
func endPointSetter(position *position, at *point)             { position.end = at }
func shouldUpdateStartPoint(reduced *point, value *point) bool { return value.offset < reduced.offset }
func shouldUpdateEndPoint(reduced *point, value *point) bool   { return value.offset > reduced.offset }
