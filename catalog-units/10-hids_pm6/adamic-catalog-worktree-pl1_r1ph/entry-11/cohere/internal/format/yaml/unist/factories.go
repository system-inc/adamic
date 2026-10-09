package unist

// Ported from yaml-unist-parser 3.2.0, dist/factories/*.mjs.
//
// Upstream's factories return object literals, some spreading a content object ({ anchor, tag,
// middleComments }) or another factory's result into them. Here each builds a *Node and registers its
// position with the context, since positions are shared and reassigned during the build (see point and
// position in context.go). A field upstream's literal has as an empty array is a non-nil empty slice
// here, as the printer's tree loader makes it from the JSON; a field it has as null is nil.

// content is createContent's { anchor, tag, middleComments }.
type content struct {
	tag            *Node
	anchor         *Node
	middleComments []*Node
}

// createContent is upstream's createContent.
func createContent(tag *Node, anchor *Node, middleComments []*Node) content {
	return content{tag: tag, anchor: anchor, middleComments: middleComments}
}

// createPosition is upstream's createPosition: a new position object over the two point objects.
func (context *context) createPosition(start *point, end *point) *position {
	return context.memory.positions.New(position{start: start, end: end})
}

// createEmptyPosition is upstream's createEmptyPosition: start and end are one point object.
func (context *context) createEmptyPosition(at *point) *position {
	return context.memory.positions.New(position{start: at, end: at})
}

// newNode makes a node of the type and registers its position object.
func (context *context) newNode(nodeType string, position *position) *Node {
	node := context.nodes.New(Node{NodeType: nodeType})
	context.positions[node] = position
	return node
}

// withContent is `...content`.
func (node *Node) withContent(content content) *Node {
	node.Anchor = content.anchor
	node.Tag = content.tag
	node.MiddleComments = content.middleComments
	return node
}

// createAlias is upstream's createAlias.
func (context *context) createAlias(position *position, content content, value string) *Node {
	node := context.newNode("alias", position)
	node.LeadingComments = []*Node{}
	node.withContent(content)
	node.Value = value
	return node
}

// createAnchor is upstream's createAnchor.
func (context *context) createAnchor(position *position, value string) *Node {
	node := context.newNode("anchor", position)
	node.Value = value
	return node
}

// createBlockValue is upstream's createBlockValue, with createBlockFolded's and createBlockLiteral's
// type, which they set over the spread blockValue.
func (context *context) createBlockValue(nodeType string, position *position, content content, chomping string, indent *int, value string, indicatorComment *Node) *Node {
	node := context.newNode(nodeType, position)
	node.LeadingComments = []*Node{}
	node.withContent(content)
	node.Chomping = chomping
	node.Indent = indent
	node.Value = value
	node.IndicatorComment = indicatorComment
	return node
}

// createComment is the object transformComment returns.
func (context *context) createComment(position *position, value string) *Node {
	node := context.newNode("comment", position)
	node.Value = value
	return node
}

// createDirective is upstream's createDirective.
func (context *context) createDirective(position *position, name string, parameters []string) *Node {
	node := context.newNode("directive", position)
	node.LeadingComments = []*Node{}
	node.Name = name
	node.Parameters = parameters
	return node
}

// createDocument is upstream's createDocument.
func (context *context) createDocument(position *position, directivesEndMarker bool, documentEndMarker bool, head *Node, body *Node, trailingComment *Node) *Node {
	node := context.newNode("document", position)
	node.TrailingComment = trailingComment
	node.DirectivesEndMarker = directivesEndMarker
	node.DocumentEndMarker = documentEndMarker
	node.Children = []*Node{head, body}
	return node
}

// createDocumentBody is upstream's createDocumentBody.
func (context *context) createDocumentBody(position *position, content *Node, endComments []*Node) *Node {
	node := context.newNode("documentBody", position)
	node.EndComments = endComments
	node.Children = optionalChildren(content)
	return node
}

// createDocumentHead is upstream's createDocumentHead.
func (context *context) createDocumentHead(position *position, children []*Node, endComments []*Node, trailingComment *Node) *Node {
	node := context.newNode("documentHead", position)
	node.EndComments = endComments
	node.TrailingComment = trailingComment
	node.Children = children
	return node
}

// createFlowCollection is upstream's createFlowCollection, with createFlowMapping's and
// createFlowSequence's type.
func (context *context) createFlowCollection(nodeType string, position *position, content content, children []*Node) *Node {
	node := context.newNode(nodeType, position)
	node.LeadingComments = []*Node{}
	node.EndComments = []*Node{}
	node.withContent(content)
	node.Children = children
	return node
}

// createFlowMapping is upstream's createFlowMapping.
func (context *context) createFlowMapping(position *position, content content, children []*Node) *Node {
	return context.createFlowCollection("flowMapping", position, content, children)
}

// createFlowSequence is upstream's createFlowSequence.
func (context *context) createFlowSequence(position *position, content content, children []*Node) *Node {
	return context.createFlowCollection("flowSequence", position, content, children)
}

// createFlowMappingItem is upstream's createFlowMappingItem.
func (context *context) createFlowMappingItem(position *position, key *Node, value *Node) *Node {
	node := context.newNode("flowMappingItem", position)
	node.LeadingComments = []*Node{}
	node.Children = []*Node{key, value}
	return node
}

// createFlowSequenceItem is upstream's createFlowSequenceItem.
func (context *context) createFlowSequenceItem(position *position, content *Node) *Node {
	node := context.newNode("flowSequenceItem", position)
	node.Children = []*Node{content}
	return node
}

// createMappingItem is upstream's createMappingItem.
func (context *context) createMappingItem(position *position, key *Node, value *Node) *Node {
	node := context.newNode("mappingItem", position)
	node.LeadingComments = []*Node{}
	node.Children = []*Node{key, value}
	return node
}

// createMappingKey is upstream's createMappingKey.
func (context *context) createMappingKey(position *position, content *Node) *Node {
	node := context.newNode("mappingKey", position)
	node.EndComments = []*Node{}
	node.Children = optionalChildren(content)
	return node
}

// createMappingValue is upstream's createMappingValue.
func (context *context) createMappingValue(position *position, content *Node) *Node {
	node := context.newNode("mappingValue", position)
	node.LeadingComments = []*Node{}
	node.EndComments = []*Node{}
	node.Children = optionalChildren(content)
	return node
}

// createMapping is upstream's createMapping.
func (context *context) createMapping(position *position, content content, children []*Node) *Node {
	node := context.newNode("mapping", position)
	node.LeadingComments = []*Node{}
	node.withContent(content)
	node.Children = children
	return node
}

// createPlain is upstream's createPlain.
func (context *context) createPlain(position *position, content content, value string) *Node {
	node := context.newNode("plain", position)
	node.LeadingComments = []*Node{}
	node.withContent(content)
	node.Value = value
	return node
}

// createQuoteValue is upstream's createQuoteValue, with createQuoteDouble's and createQuoteSingle's
// type.
func (context *context) createQuoteValue(nodeType string, position *position, content content, value string) *Node {
	node := context.newNode(nodeType, position)
	node.withContent(content)
	node.LeadingComments = []*Node{}
	node.Value = value
	return node
}

// createRoot is upstream's createRoot.
func (context *context) createRoot(position *position, children []*Node, comments []*Node) *Node {
	node := context.newNode("root", position)
	node.Children = children
	node.Comments = comments
	return node
}

// createSequenceItem is upstream's createSequenceItem.
func (context *context) createSequenceItem(position *position, content *Node) *Node {
	node := context.newNode("sequenceItem", position)
	node.LeadingComments = []*Node{}
	node.EndComments = []*Node{}
	node.Children = optionalChildren(content)
	return node
}

// createSequence is upstream's createSequence.
func (context *context) createSequence(position *position, content content, children []*Node) *Node {
	node := context.newNode("sequence", position)
	node.LeadingComments = []*Node{}
	node.EndComments = []*Node{}
	node.withContent(content)
	node.Children = children
	return node
}

// createTag is upstream's createTag.
func (context *context) createTag(position *position, value string) *Node {
	node := context.newNode("tag", position)
	node.Value = value
	return node
}

// optionalChildren is `!content ? [] : [content]`.
func optionalChildren(content *Node) []*Node {
	if content == nil {
		return []*Node{}
	}
	return []*Node{content}
}

// The `"field" in node` checks attach.mjs and update-positions.mjs make, by the fields each factory's
// literal has. A nil pointer or slice cannot say whether upstream's field is null or absent, so the
// checks that read presence alone go through these.

// hasLeadingCommentsField is `"leadingComments" in node`.
func hasLeadingCommentsField(nodeType string) bool {
	switch nodeType {
	case "alias", "blockFolded", "blockLiteral", "directive", "flowMapping", "flowSequence", "flowMappingItem",
		"mappingItem", "mappingValue", "mapping", "plain", "quoteDouble", "quoteSingle", "sequenceItem", "sequence":
		return true
	}
	return false
}

// hasTrailingCommentField is `"trailingComment" in node`.
func hasTrailingCommentField(nodeType string) bool {
	switch nodeType {
	case "alias", "directive", "document", "documentHead", "flowMapping", "flowSequence", "mappingKey",
		"mappingValue", "plain", "quoteDouble", "quoteSingle", "sequenceItem":
		return true
	}
	return false
}

// hasChildrenField is `"children" in node`.
func hasChildrenField(nodeType string) bool {
	switch nodeType {
	case "root", "document", "documentHead", "documentBody", "flowMapping", "flowSequence", "flowMappingItem",
		"flowSequenceItem", "mappingItem", "mappingKey", "mappingValue", "mapping", "sequenceItem", "sequence":
		return true
	}
	return false
}
