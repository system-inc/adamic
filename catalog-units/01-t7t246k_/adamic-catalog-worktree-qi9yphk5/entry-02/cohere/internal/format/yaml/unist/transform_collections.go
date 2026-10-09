package unist

// Ported from yaml-unist-parser 3.2.0, dist/transforms/: map.mjs, seq.mjs, flow-map.mjs, flow-seq.mjs
// and pair.mjs.

import (
	"github.com/system-inc/cohere/internal/format/yaml/compose"
	"github.com/system-inc/cohere/internal/format/yaml/cst"
)

// itemAt is `srcToken.items[index]`: nil, upstream's undefined, past the end.
func itemAt(items []*cst.CollectionItem, index int) *cst.CollectionItem {
	if index < len(items) {
		return items[index]
	}
	return nil
}

// createItemFunc is the factory transformPair builds its item with: createMappingItem or
// createFlowMappingItem.
type createItemFunc func(position *position, key *Node, value *Node) *Node

// transformMap is upstream's transformMap.
func (context *context) transformMap(mapNode *compose.Node, props []*cst.Token) *Node {
	srcToken := mapNode.SrcToken
	if srcToken == nil || srcToken.Type != "block-map" {
		throwError("Expected block mapping srcToken")
	}
	mappingItems := make([]*Node, len(mapNode.Items))
	for index, pair := range mapNode.Items {
		srcItem := itemAt(srcToken.Items, index)
		mappingItems[index] = context.transformPair(pair, srcItem, context.createMappingItem)
	}
	if len(mapNode.Items) < len(srcToken.Items) {
		for i := len(mapNode.Items); i < len(srcToken.Items); i++ {
			srcItem := srcToken.Items[i]
			for _, token := range context.extractComments(srcItem.Start) {
				throwError("Unexpected token type in collection item start: %s", token.Type)
			}
		}
	}
	if len(mappingItems) == 0 {
		// mappingItems[0].position
		throwTypeError("undefined", "position")
	}
	return context.createMapping(
		context.createPosition(context.position(mappingItems[0]).start, context.position(mappingItems[len(mappingItems)-1]).end),
		context.transformContentProperties(mapNode, props), mappingItems)
}

// transformSeq is upstream's transformSeq.
func (context *context) transformSeq(seq *compose.Node, props []*cst.Token) *Node {
	srcToken := seq.SrcToken
	if srcToken == nil || srcToken.Type != "block-seq" {
		throwError("Expected block sequence srcToken")
	}
	sequenceItems := make([]*Node, len(seq.Items))
	for index, itemNode := range seq.Items {
		srcItem := itemAt(srcToken.Items, index)
		if srcItem == nil {
			throwTypeError("undefined", "start")
		}
		propTokens := []*cst.Token{}
		var seqItemIndToken *cst.Token
		for _, token := range tokens(srcItem.Start) {
			if maybeContentPropertyToken(token) {
				propTokens = append(propTokens, token)
				continue
			}
			if token.Type == "seq-item-ind" {
				seqItemIndToken = token
				continue
			}
			throwError("Unexpected token type in sequence item start: %s", token.Type)
		}
		item := context.transformItemValue(itemNode, propTokens)
		var start *point
		if seqItemIndToken != nil {
			start = context.transformOffset(seqItemIndToken.Offset)
		} else {
			if item == nil {
				throwTypeError("null", "position")
			}
			start = context.position(item).start
		}
		// `item?.position.end ?? context.transformOffset(seqItemIndToken.offset + ...)`
		var end *point
		if item != nil {
			end = context.position(item).end
		} else {
			end = context.transformOffset(seqItemIndToken.Offset + len(seqItemIndToken.Source))
		}
		sequenceItems[index] = context.createSequenceItem(context.createPosition(start, end), item)
	}
	if len(seq.Items) < len(srcToken.Items) {
		for i := len(seq.Items); i < len(srcToken.Items); i++ {
			srcItem := srcToken.Items[i]
			for _, token := range context.extractComments(srcItem.Start) {
				throwError("Unexpected token type in collection item start: %s", token.Type)
			}
		}
	}
	if len(sequenceItems) == 0 {
		throwTypeError("undefined", "position")
	}
	return context.createSequence(
		context.createPosition(context.position(sequenceItems[0]).start, context.position(sequenceItems[len(sequenceItems)-1]).end),
		context.transformContentProperties(seq, props), sequenceItems)
}

// transformItemValue is upstream's transformItemValue. A pair is an item of a sequence a !!pairs or
// !!omap tag resolved; its srcToken is the collection item of the map it came from, and the mapping
// built around it shares the item's position object, as upstream's does.
func (context *context) transformItemValue(itemNode *compose.Node, props []*cst.Token) *Node {
	if !compose.IsPair(itemNode) {
		if isEmptyNode(itemNode, props) {
			context.extractComments(props)
			return nil
		}
		return context.transformNode(itemNode, props)
	}
	srcItem := itemNode.SrcItem
	mappingItem := context.transformPair(itemNode, srcItem, context.createMappingItem)
	return context.createMapping(context.position(mappingItem), context.transformContentProperties(itemNode.Key, props),
		[]*Node{mappingItem})
}

// transformFlowMap is upstream's transformFlowMap.
func (context *context) transformFlowMap(flowMap *compose.Node, props []*cst.Token) *Node {
	srcToken := flowMap.SrcToken
	if srcToken == nil || srcToken.Type != "flow-collection" {
		throwError("Expected flow-collection CST node for flow map")
	}
	flowMappingItems := make([]*Node, len(flowMap.Items))
	for index, pair := range flowMap.Items {
		srcItem := itemAt(srcToken.Items, index)
		flowMappingItems[index] = context.transformPair(pair, srcItem, context.createFlowMappingItem)
	}
	context.extractTrailingItemComments(len(flowMap.Items), srcToken.Items)
	var flowMapEndToken *cst.Token
	for _, token := range context.extractComments(srcToken.End) {
		if token.Type == "flow-map-end" {
			flowMapEndToken = token
			continue
		}
		throwError("Unexpected token type in flow map end: %s", token.Type)
	}
	if flowMapEndToken == nil {
		throwError("Expected flow-map-end token")
	}
	return context.createFlowMapping(
		context.transformRange(srcToken.FlowStart.Offset, flowMapEndToken.Offset+len(flowMapEndToken.Source)),
		context.transformContentProperties(flowMap, props), flowMappingItems)
}

// extractTrailingItemComments is the loop transformFlowMap and transformFlowSeq share: the source items
// past the composed ones hold only comments and commas.
func (context *context) extractTrailingItemComments(composed int, items []*cst.CollectionItem) {
	if composed < len(items) {
		for i := composed; i < len(items); i++ {
			srcItem := items[i]
			for _, token := range context.extractComments(srcItem.Start) {
				if token.Type == "comma" {
					continue
				}
				throwError("Unexpected token type in collection item start: %s", token.Type)
			}
		}
	}
}

// transformFlowSeq is upstream's transformFlowSeq.
func (context *context) transformFlowSeq(flowSeq *compose.Node, props []*cst.Token) *Node {
	srcToken := flowSeq.SrcToken
	if srcToken == nil || srcToken.Type != "flow-collection" {
		throwError("Expected flow-collection CST node for flow sequence")
	}
	flowSequenceItems := make([]*Node, len(flowSeq.Items))
	for index, item := range flowSeq.Items {
		srcItem := itemAt(srcToken.Items, index)
		if isBlockMappingOfImmediateChildOfFlowSequence(item, srcItem) {
			flowSequenceItems[index] = context.transformPair(item.Items[len(item.Items)-1], srcItem, context.createFlowMappingItem)
			continue
		}
		if !compose.IsPair(item) {
			if srcItem == nil {
				throwTypeError("undefined", "start")
			}
			propTokens := []*cst.Token{}
			for _, token := range tokens(srcItem.Start) {
				if maybeContentPropertyToken(token) {
					propTokens = append(propTokens, token)
					continue
				}
				if token.Type == "comma" {
					continue
				}
				throwError("Unexpected token type in sequence item start: %s", token.Type)
			}
			node := context.transformNode(item, propTokens)
			if node == nil {
				throwTypeError("null", "position")
			}
			flowSequenceItems[index] = context.createFlowSequenceItem(
				context.createPosition(context.position(node).start, context.position(node).end), node)
		} else {
			flowSequenceItems[index] = context.transformPair(item, srcItem, context.createFlowMappingItem)
		}
	}
	context.extractTrailingItemComments(len(flowSeq.Items), srcToken.Items)
	var flowSeqEndToken *cst.Token
	for _, token := range tokens(srcToken.End) {
		if token.Type == "comment" {
			context.transformComment(token)
			continue
		}
		if token.Type == "flow-seq-end" {
			flowSeqEndToken = token
			continue
		}
		throwError("Unexpected token type in flow seq end: %s", token.Type)
	}
	if flowSeqEndToken == nil {
		throwError("Expected flow-seq-end token")
	}
	return context.createFlowSequence(
		context.transformRange(srcToken.FlowStart.Offset, flowSeqEndToken.Offset+len(flowSeqEndToken.Source)),
		context.transformContentProperties(flowSeq, props), flowSequenceItems)
}

// isBlockMappingOfImmediateChildOfFlowSequence is upstream's: the map the composer wraps around a pair
// in a flow sequence, `[ key: value ]`, which has no source token and one pair whose source is this
// item. Both sides of the last comparison can be undefined, and undefined === undefined.
func isBlockMappingOfImmediateChildOfFlowSequence(item *compose.Node, srcItem *cst.CollectionItem) bool {
	if item.SrcToken != nil {
		return false
	}
	if !compose.IsMap(item) {
		return false
	}
	if len(item.Items) != 1 {
		return false
	}
	return item.Items[0].SrcItem == srcItem
}

// transformPair is upstream's transformPair (pair.mjs).
func (context *context) transformPair(pair *compose.Node, srcItem *cst.CollectionItem, createNode createItemFunc) *Node {
	if srcItem == nil {
		throwTypeError("undefined", "start")
	}
	keyPropTokens := []*cst.Token{}
	var explicitKeyIndToken *cst.Token
	for _, token := range tokens(srcItem.Start) {
		if maybeContentPropertyToken(token) {
			keyPropTokens = append(keyPropTokens, token)
			continue
		}
		if token.Type == "explicit-key-ind" {
			explicitKeyIndToken = token
			continue
		}
		if token.Type == "comma" {
			continue
		}
		throwError("Unexpected token type in collection item start: %s", token.Type)
	}
	valuePropTokens := []*cst.Token{}
	var mapValueIndToken *cst.Token
	for _, token := range tokens(srcItem.Sep) {
		if maybeContentPropertyToken(token) {
			valuePropTokens = append(valuePropTokens, token)
			continue
		}
		if token.Type == "map-value-ind" {
			mapValueIndToken = token
			continue
		}
		throwError("Unexpected token type in collection item sep: %s", token.Type)
	}

	// explicitKeyIndToken?.offset ?? srcItem.key?.offset ?? mapValueIndToken?.offset ?? srcItem.value.offset
	var keyStartOffset int
	switch {
	case explicitKeyIndToken != nil:
		keyStartOffset = explicitKeyIndToken.Offset
	case srcItem.Key != nil:
		keyStartOffset = srcItem.Key.Offset
	case mapValueIndToken != nil:
		keyStartOffset = mapValueIndToken.Offset
	default:
		if srcItem.Value == nil {
			throwTypeError("undefined", "offset")
		}
		keyStartOffset = srcItem.Value.Offset
	}
	keyRange := []int{keyStartOffset, 0}
	switch {
	case srcItem.Key != nil:
		if pair.Key == nil {
			throwTypeError("null", "range")
		}
		if pair.Key.Range == nil {
			throwTypeError("undefined", "1")
		}
		keyRange[1] = pair.Key.Range[1]
	case explicitKeyIndToken != nil:
		keyRange[1] = explicitKeyIndToken.Offset + len(explicitKeyIndToken.Source)
	default:
		keyRange[1] = keyStartOffset
	}
	var valueRange []int
	if pair.Value != nil {
		// mapValueIndToken?.offset ?? srcItem.value?.offset ?? pair.value.range[0]
		var valueStartOffset int
		switch {
		case mapValueIndToken != nil:
			valueStartOffset = mapValueIndToken.Offset
		case srcItem.Value != nil:
			valueStartOffset = srcItem.Value.Offset
		default:
			valueStartOffset = pair.Value.Range[0]
		}
		valueRange = []int{valueStartOffset, 0}
		switch {
		case srcItem.Value != nil:
			valueRange[1] = pair.Value.Range[1]
		case mapValueIndToken != nil:
			valueRange[1] = mapValueIndToken.Offset + len(mapValueIndToken.Source)
		default:
			valueRange[1] = valueStartOffset
		}
	}
	return context.transformAstPair(pair, createNode, keyRange, keyPropTokens, valueRange, valuePropTokens)
}

// transformAstPair is upstream's transformAstPair. Upstream's additionalKeyData and additionalValueData
// are { range, props }: the ranges, nil for null, and the prop tokens.
func (context *context) transformAstPair(pair *compose.Node, createNode createItemFunc, keyRange []int, keyProps []*cst.Token, valueRange []int, valueProps []*cst.Token) *Node {
	var keyContent *Node
	if !isEmptyNode(pair.Key, keyProps) {
		keyContent = context.transformNode(pair.Key, keyProps)
	} else {
		context.extractComments(keyProps)
	}
	var valueContent *Node
	if !isEmptyNode(pair.Value, valueProps) {
		valueContent = context.transformNode(pair.Value, valueProps)
	} else {
		context.extractComments(valueProps)
	}

	// keyRange is never null upstream either; keyContent's start is the fallback it never takes.
	keyEnd := keyRange[1]
	if keyContent != nil {
		keyEnd = context.position(keyContent).end.offset
	}
	mappingKey := context.createMappingKey(context.transformRange(keyRange[0], keyEnd), keyContent)

	// `valueContent || additionalValueData.range ? createMappingValue(...) : null`
	var mappingValue *Node
	if valueContent != nil || valueRange != nil {
		var start, end int
		if valueRange != nil {
			start = valueRange[0]
		} else {
			start = context.position(valueContent).start.offset
		}
		if valueContent != nil {
			end = context.position(valueContent).end.offset
		} else {
			end = valueRange[0] + 1
		}
		mappingValue = context.createMappingValue(context.transformRange(start, end), valueContent)
	}

	itemEnd := context.position(mappingKey).end
	if mappingValue != nil {
		itemEnd = context.position(mappingValue).end
	} else {
		mappingValue = context.createMappingValue(context.createEmptyPosition(context.position(mappingKey).end), nil)
	}
	return createNode(context.createPosition(context.position(mappingKey).start, itemEnd), mappingKey, mappingValue)
}
