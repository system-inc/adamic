package compose

// Ported from eemeli/yaml 2.9.0, dist/compose/compose-doc.js, compose-node.js and
// util-empty-scalar-position.js.

import "github.com/system-inc/cohere/internal/format/yaml/cst"

// composeContext is upstream's ComposeContext, one object shared and mutated down the whole document:
// atKey and atRoot are switched as composing descends.
type composeContext struct {
	atKey      bool
	atRoot     bool
	directives *Directives
	options    *Options
	schema     *Schema
}

func composeDoc(options Options, directives *Directives, token *cst.Token, onError onErrorFunc) *Document {
	offset, start, value, end := token.Offset, token.Start, token.Value, token.End
	doc := newDocument(directives, options)
	ctx := &composeContext{
		atKey:      false,
		atRoot:     true,
		directives: doc.Directives,
		options:    doc.Options,
		schema:     doc.Schema,
	}
	// `value ?? end?.[0]`
	next := value
	if next == nil && len(end) > 0 {
		next = end[0]
	}
	props := resolveProps(start, resolvePropsOptions{
		indicator:      "doc-start",
		next:           next,
		offset:         offset,
		onError:        onError,
		parentIndent:   0,
		startOnNewline: true,
	})
	if props.found != nil {
		doc.Directives.DocStart = true
		if value != nil && (value.Type == "block-map" || value.Type == "block-seq") && !props.hasNewline {
			onError(props.end, "MISSING_CHAR", "Block collection cannot start on same line with directives-end marker", false)
		}
	}
	if value != nil {
		doc.Contents = composeNode(ctx, value, props, onError)
	} else {
		doc.Contents = composeEmptyNode(ctx, props.end, start, props, onError)
	}
	contentEnd := doc.Contents.Range[2]
	re := resolveEnd(end, contentEnd, false, onError)
	if re.comment != "" {
		doc.Comment = re.comment
	}
	doc.Range = []int{offset, contentEnd, re.offset}
	return doc
}

func composeNode(ctx *composeContext, token *cst.Token, props props, onError onErrorFunc) *Node {
	atKey := ctx.atKey
	spaceBefore, comment, anchor, tag := props.spaceBefore, props.comment, props.anchor, props.tag
	var node *Node
	isSrcToken := true
	switch token.Type {
	case "alias":
		node = composeAlias(ctx, token, onError)
		if anchor != nil || tag != nil {
			onError(token, "ALIAS_PROPS", "An alias node must not specify any properties", false)
		}
	case "scalar", "single-quoted-scalar", "double-quoted-scalar", "block-scalar":
		node = composeScalar(ctx, token, tag, onError)
		if anchor != nil {
			node.SetAnchor(unitsToString(substring(anchor.Source, 1, len(anchor.Source))))
		}
	case "block-map", "block-seq", "flow-collection":
		node = composeCollectionCatching(ctx, token, props, onError)
	default:
		message := "Unsupported token (type: " + token.Type + ")"
		if token.Type == "error" {
			message = token.Message
		}
		onError(token, "UNEXPECTED_TOKEN", message, false)
		isSrcToken = false
	}
	if node == nil {
		node = composeEmptyNode(ctx, token.Offset, nil, props, onError)
	}
	// `node.anchor === ''`: an alias never gets the anchor, so its anchor is undefined, not ''.
	if anchor != nil && node.HasAnchor() && node.Anchor == "" {
		onError(anchor, "BAD_ALIAS", "Anchor cannot be an empty string", false)
	}
	if atKey && ctx.options.StringKeys && (!IsScalar(node) || !isString(node.ScalarValue) ||
		node.Tag != "" && node.Tag != strTagName) {
		msg := "With stringKeys, all keys must be strings"
		var source any = token
		if tag != nil {
			source = tag
		}
		onError(source, "NON_STRING_KEY", msg, false)
	}
	if spaceBefore {
		node.SpaceBefore = true
	}
	if comment != "" {
		if token.Type == "scalar" && len(token.Source) == 0 {
			node.Comment = comment
		} else {
			node.CommentBefore = comment
		}
	}
	if ctx.options.KeepSourceTokens && isSrcToken {
		node.SrcToken = token
	}
	return node
}

// composeCollectionCatching is upstream's try/catch around composeCollection: an exception becomes a
// RESOURCE_EXHAUSTION error and no node. Upstream expects a stack overflow there ("almost certainly");
// Go's stack grows instead of overflowing at V8's depth, so the port composes what upstream would have
// given up on. The one exception the port does raise is the TypeError resolveFlowCollection meets on a
// comment that belongs to a previous item that was never composed.
func composeCollectionCatching(ctx *composeContext, token *cst.Token, props props, onError onErrorFunc) (node *Node) {
	defer func() {
		if recovered := recover(); recovered != nil {
			message, isTypeError := recovered.(typeError)
			if !isTypeError {
				panic(recovered)
			}
			onError(token, "RESOURCE_EXHAUSTION", string(message), false)
			node = nil
		}
	}()
	node = composeCollection(ctx, token, props, onError)
	if props.anchor != nil {
		node.SetAnchor(unitsToString(substring(props.anchor.Source, 1, len(props.anchor.Source))))
	}
	return node
}

func isString(value any) bool {
	_, ok := value.(string)
	return ok
}

// composeEmptyNode composes the empty scalar where a node is missing. before is nil for upstream's
// undefined; upstream's pos argument is always null, so it is left out.
func composeEmptyNode(ctx *composeContext, offset int, before []*cst.Token, props props, onError onErrorFunc) *Node {
	token := &cst.Token{
		Type:   "scalar",
		Offset: emptyScalarPosition(offset, before),
		Indent: -1,
		Source: []uint16{},
	}
	node := composeScalar(ctx, token, props.tag, onError)
	if props.anchor != nil {
		node.SetAnchor(unitsToString(substring(props.anchor.Source, 1, len(props.anchor.Source))))
		if node.Anchor == "" {
			onError(props.anchor, "BAD_ALIAS", "Anchor cannot be an empty string", false)
		}
	}
	if props.spaceBefore {
		node.SpaceBefore = true
	}
	if props.comment != "" {
		node.Comment = props.comment
		node.Range[2] = props.end
	}
	return node
}

func composeAlias(ctx *composeContext, token *cst.Token, onError onErrorFunc) *Node {
	offset, source, end := token.Offset, token.Source, token.End
	alias := newAlias(unitsToString(substring(source, 1, len(source))))
	if alias.Source == "" {
		onError(offset, "BAD_ALIAS", "Alias cannot be an empty string", false)
	}
	if characterAt(source, len(source)-1) == ':' && len(source) > 1 {
		onError(offset+len(source)-1, "BAD_ALIAS", "Alias ending in : is ambiguous", true)
	}
	valueEnd := offset + len(source)
	re := resolveEnd(end, valueEnd, ctx.options.Strict, onError)
	alias.Range = []int{offset, valueEnd, re.offset}
	if re.comment != "" {
		alias.Comment = re.comment
	}
	return alias
}

// emptyScalarPosition is upstream's with pos null: the offset after the last non-empty token before,
// moved forward past any spaces that follow it.
func emptyScalarPosition(offset int, before []*cst.Token) int {
	if before != nil {
		pos := len(before)
		for i := pos - 1; i >= 0; i-- {
			st := before[i]
			switch st.Type {
			case "space", "comment", "newline":
				offset -= len(st.Source)
				continue
			}
			// Technically, an empty scalar is immediately after the last non-empty
			// node, but it's more useful to place it after any whitespace.
			i++
			for i < len(before) && before[i].Type == "space" {
				offset += len(before[i].Source)
				i++
			}
			break
		}
	}
	return offset
}
