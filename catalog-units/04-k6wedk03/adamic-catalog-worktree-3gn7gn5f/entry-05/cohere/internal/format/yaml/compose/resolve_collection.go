package compose

// Ported from eemeli/yaml 2.9.0, dist/compose/resolve-block-map.js, resolve-block-seq.js,
// resolve-flow-collection.js, util-contains-newline.js and util-map-includes.js.
// util-flow-indent-check.js runs only with the compat schema option, which is never set on this path.

import (
	"strings"

	"github.com/system-inc/cohere/internal/format/yaml/cst"
)

// nodeClass is `tag?.nodeClass ?? fallback`.
func nodeClass(tag *Tag, fallback string) string {
	if tag != nil && tag.NodeClass != "" {
		return tag.NodeClass
	}
	return fallback
}

const startColMsg = "All mapping items must start at the same column"

func resolveBlockMap(ctx *composeContext, bm *cst.Token, onError onErrorFunc, tag *Tag) *Node {
	yamlMap := ctx.newCollection(nodeClass(tag, "YAMLMap"))
	if ctx.atRoot {
		ctx.atRoot = false
	}
	offset := bm.Offset
	// commentEnd is null until set; upstream tests it for truthiness, so 0 counts as unset there.
	commentEnd, commentEndSet := 0, false
	for _, collItem := range bm.Items {
		start, key, sep, value := collItem.Start, collItem.Key, collItem.Sep, collItem.Value
		// key properties
		keyProps := resolveProps(start, resolvePropsOptions{
			indicator:      "explicit-key-ind",
			next:           keyOrFirstSep(key, sep),
			offset:         offset,
			onError:        onError,
			parentIndent:   bm.Indent,
			startOnNewline: true,
		})
		implicitKey := keyProps.found == nil
		if implicitKey {
			if key != nil {
				if key.Type == "block-seq" {
					onError(offset, "BLOCK_AS_IMPLICIT_KEY", "A block sequence may not be used as an implicit map key", false)
				} else if key.HasIndent() && key.Indent != bm.Indent {
					onError(offset, "BAD_INDENT", startColMsg, false)
				}
			}
			if keyProps.anchor == nil && keyProps.tag == nil && sep == nil {
				commentEnd, commentEndSet = keyProps.end, true
				if keyProps.comment != "" {
					if yamlMap.Comment != "" {
						yamlMap.Comment += "\n" + keyProps.comment
					} else {
						yamlMap.Comment = keyProps.comment
					}
				}
				continue
			}
			if keyProps.newlineAfterProp != nil || containsNewline(key) {
				// `key ?? start[start.length - 1]`
				var source any = key
				if key == nil {
					source = start[len(start)-1]
				}
				onError(source, "MULTILINE_IMPLICIT_KEY", "Implicit keys need to be on a single line", false)
			}
		} else if keyProps.found.Indent != bm.Indent {
			onError(offset, "BAD_INDENT", startColMsg, false)
		}
		// key value
		ctx.atKey = true
		keyStart := keyProps.end
		var keyNode *Node
		if key != nil {
			keyNode = composeNode(ctx, key, keyProps, onError)
		} else {
			keyNode = composeEmptyNode(ctx, keyStart, start, keyProps, onError)
		}
		ctx.atKey = false
		if mapIncludes(ctx, yamlMap.Items, keyNode) {
			onError(keyStart, "DUPLICATE_KEY", "Map keys must be unique", false)
		}
		// value properties
		valueProps := resolveProps(sepOrEmpty(sep), resolvePropsOptions{
			indicator:      "map-value-ind",
			next:           value,
			offset:         keyNode.Range[2],
			onError:        onError,
			parentIndent:   bm.Indent,
			startOnNewline: key == nil || key.Type == "block-scalar",
		})
		offset = valueProps.end
		if valueProps.found != nil {
			if implicitKey {
				if value != nil && value.Type == "block-map" && !valueProps.hasNewline {
					onError(offset, "BLOCK_AS_IMPLICIT_KEY", "Nested mappings are not allowed in compact mappings", false)
				}
				if ctx.options.Strict && keyProps.start < valueProps.found.Offset-1024 {
					onError(keyNode.Range, "KEY_OVER_1024_CHARS",
						"The : indicator must be at most 1024 chars after the start of an implicit block mapping key", false)
				}
			}
			// value value
			var valueNode *Node
			if value != nil {
				valueNode = composeNode(ctx, value, valueProps, onError)
			} else {
				valueNode = composeEmptyNode(ctx, offset, sep, valueProps, onError)
			}
			offset = valueNode.Range[2]
			pair := ctx.newPair(keyNode, valueNode)
			if ctx.options.KeepSourceTokens {
				pair.SrcItem = collItem
			}
			yamlMap.Items = append(yamlMap.Items, pair)
		} else {
			// key with no value
			if implicitKey {
				onError(keyNode.Range, "MISSING_CHAR", "Implicit map keys need to be followed by map values", false)
			}
			if valueProps.comment != "" {
				if keyNode.Comment != "" {
					keyNode.Comment += "\n" + valueProps.comment
				} else {
					keyNode.Comment = valueProps.comment
				}
			}
			pair := ctx.newPair(keyNode, nil)
			if ctx.options.KeepSourceTokens {
				pair.SrcItem = collItem
			}
			yamlMap.Items = append(yamlMap.Items, pair)
		}
	}
	if commentEndSet && commentEnd != 0 && commentEnd < offset {
		onError(commentEnd, "IMPOSSIBLE", "Map comment with trailing content", false)
	}
	end := offset
	if commentEndSet {
		end = commentEnd
	}
	yamlMap.Range = []int{bm.Offset, offset, end}
	return yamlMap
}

// keyOrFirstSep is `key ?? sep?.[0]`.
func keyOrFirstSep(key *cst.Token, sep []*cst.Token) *cst.Token {
	if key != nil {
		return key
	}
	if len(sep) > 0 {
		return sep[0]
	}
	return nil
}

// sepOrEmpty is `sep ?? []`.
func sepOrEmpty(sep []*cst.Token) []*cst.Token {
	if sep == nil {
		return []*cst.Token{}
	}
	return sep
}

func resolveBlockSeq(ctx *composeContext, bs *cst.Token, onError onErrorFunc, tag *Tag) *Node {
	seq := ctx.newCollection(nodeClass(tag, "YAMLSeq"))
	if ctx.atRoot {
		ctx.atRoot = false
	}
	if ctx.atKey {
		ctx.atKey = false
	}
	offset := bs.Offset
	commentEnd, commentEndSet := 0, false
	for _, item := range bs.Items {
		start, value := item.Start, item.Value
		props := resolveProps(start, resolvePropsOptions{
			indicator:      "seq-item-ind",
			next:           value,
			offset:         offset,
			onError:        onError,
			parentIndent:   bs.Indent,
			startOnNewline: true,
		})
		if props.found == nil {
			if props.anchor != nil || props.tag != nil || value != nil {
				if value != nil && value.Type == "block-seq" {
					onError(props.end, "BAD_INDENT", "All sequence items must start at the same column", false)
				} else {
					onError(offset, "MISSING_CHAR", "Sequence item without - indicator", false)
				}
			} else {
				commentEnd, commentEndSet = props.end, true
				if props.comment != "" {
					seq.Comment = props.comment
				}
				continue
			}
		}
		var node *Node
		if value != nil {
			node = composeNode(ctx, value, props, onError)
		} else {
			node = composeEmptyNode(ctx, props.end, start, props, onError)
		}
		offset = node.Range[2]
		seq.Items = append(seq.Items, node)
	}
	end := offset
	if commentEndSet {
		end = commentEnd
	}
	seq.Range = []int{bs.Offset, offset, end}
	return seq
}

const blockMsg = "Block collections are not allowed within flow collections"

func isBlock(token *cst.Token) bool {
	return token != nil && (token.Type == "block-map" || token.Type == "block-seq")
}

// typeError is a TypeError upstream throws by reading a property of undefined. composeNode catches it
// where upstream's try/catch around composeCollection does.
type typeError string

func resolveFlowCollection(ctx *composeContext, fc *cst.Token, onError onErrorFunc, tag *Tag) *Node {
	isMap := equalsASCII(fc.FlowStart.Source, "{")
	fcName := "flow sequence"
	fallback := "YAMLSeq"
	if isMap {
		fcName = "flow map"
		fallback = "YAMLMap"
	}
	coll := ctx.newCollection(nodeClass(tag, fallback))
	coll.Flow = true
	atRoot := ctx.atRoot
	if atRoot {
		ctx.atRoot = false
	}
	if ctx.atKey {
		ctx.atKey = false
	}
	offset := fc.Offset + len(fc.FlowStart.Source)
	for i := 0; i < len(fc.Items); i++ {
		collItem := fc.Items[i]
		start, key, sep, value := collItem.Start, collItem.Key, collItem.Sep, collItem.Value
		props := resolveProps(start, resolvePropsOptions{
			flow:           fcName,
			indicator:      "explicit-key-ind",
			next:           keyOrFirstSep(key, sep),
			offset:         offset,
			onError:        onError,
			parentIndent:   fc.Indent,
			startOnNewline: false,
		})
		if props.found == nil {
			if props.anchor == nil && props.tag == nil && sep == nil && value == nil {
				if i == 0 && props.comma != nil {
					onError(props.comma, "UNEXPECTED_TOKEN", "Unexpected , in "+fcName, false)
				} else if i < len(fc.Items)-1 {
					onError(props.start, "UNEXPECTED_TOKEN", "Unexpected empty item in "+fcName, false)
				}
				if props.comment != "" {
					if coll.Comment != "" {
						coll.Comment += "\n" + props.comment
					} else {
						coll.Comment = props.comment
					}
				}
				offset = props.end
				continue
			}
			if !isMap && ctx.options.Strict && containsNewline(key) {
				onError(key, // checked by containsNewline()
					"MULTILINE_IMPLICIT_KEY", "Implicit keys of flow sequence pairs need to be on a single line", false)
			}
		}
		if i == 0 {
			if props.comma != nil {
				onError(props.comma, "UNEXPECTED_TOKEN", "Unexpected , in "+fcName, false)
			}
		} else {
			if props.comma == nil {
				onError(props.start, "MISSING_CHAR", "Missing , between "+fcName+" items", false)
			}
			if props.comment != "" {
				prevItemComment := []uint16{}
			loop:
				for _, st := range start {
					switch st.Type {
					case "comma", "space":
					case "comment":
						prevItemComment = substring(st.Source, 1, len(st.Source))
						break loop
					default:
						break loop
					}
				}
				if len(prevItemComment) > 0 {
					if len(coll.Items) == 0 {
						// `coll.items[coll.items.length - 1]` is undefined.
						panic(typeError("Cannot read properties of undefined (reading 'comment')"))
					}
					prev := coll.Items[len(coll.Items)-1]
					if IsPair(prev) {
						// `prev.value ?? prev.key`
						if prev.Value != nil {
							prev = prev.Value
						} else {
							prev = prev.Key
						}
					}
					if prev.Comment != "" {
						prev.Comment += "\n" + unitsToString(prevItemComment)
					} else {
						prev.Comment = unitsToString(prevItemComment)
					}
					props.comment = unitsToString(substring(stringToUnits(props.comment), len(prevItemComment)+1,
						UnitLength(props.comment)))
				}
			}
		}
		if !isMap && sep == nil && props.found == nil {
			// item is a value in a seq
			// → key & sep are empty, start does not include ? or :
			var valueNode *Node
			if value != nil {
				valueNode = composeNode(ctx, value, props, onError)
			} else {
				valueNode = composeEmptyNode(ctx, props.end, sep, props, onError)
			}
			coll.Items = append(coll.Items, valueNode)
			offset = valueNode.Range[2]
			if isBlock(value) {
				onError(valueNode.Range, "BLOCK_IN_FLOW", blockMsg, false)
			}
		} else {
			// item is a key+value pair
			// key value
			ctx.atKey = true
			keyStart := props.end
			var keyNode *Node
			if key != nil {
				keyNode = composeNode(ctx, key, props, onError)
			} else {
				keyNode = composeEmptyNode(ctx, keyStart, start, props, onError)
			}
			if isBlock(key) {
				onError(keyNode.Range, "BLOCK_IN_FLOW", blockMsg, false)
			}
			ctx.atKey = false
			// value properties
			valueProps := resolveProps(sepOrEmpty(sep), resolvePropsOptions{
				flow:           fcName,
				indicator:      "map-value-ind",
				next:           value,
				offset:         keyNode.Range[2],
				onError:        onError,
				parentIndent:   fc.Indent,
				startOnNewline: false,
			})
			if valueProps.found != nil {
				if !isMap && props.found == nil && ctx.options.Strict {
					for _, st := range sep {
						if st == valueProps.found {
							break
						}
						if st.Type == "newline" {
							onError(st, "MULTILINE_IMPLICIT_KEY", "Implicit keys of flow sequence pairs need to be on a single line", false)
							break
						}
					}
					if props.start < valueProps.found.Offset-1024 {
						onError(valueProps.found, "KEY_OVER_1024_CHARS",
							"The : indicator must be at most 1024 chars after the start of an implicit flow sequence key", false)
					}
				}
			} else if value != nil {
				if value.HasSource() && characterAt(value.Source, 0) == ':' {
					onError(value, "MISSING_CHAR", "Missing space after : in "+fcName, false)
				} else {
					onError(valueProps.start, "MISSING_CHAR", "Missing , or : between "+fcName+" items", false)
				}
			}
			// value value
			var valueNode *Node
			if value != nil {
				valueNode = composeNode(ctx, value, valueProps, onError)
			} else if valueProps.found != nil {
				valueNode = composeEmptyNode(ctx, valueProps.end, sep, valueProps, onError)
			}
			if valueNode != nil {
				if isBlock(value) {
					onError(valueNode.Range, "BLOCK_IN_FLOW", blockMsg, false)
				}
			} else if valueProps.comment != "" {
				if keyNode.Comment != "" {
					keyNode.Comment += "\n" + valueProps.comment
				} else {
					keyNode.Comment = valueProps.comment
				}
			}
			pair := ctx.newPair(keyNode, valueNode)
			if ctx.options.KeepSourceTokens {
				pair.SrcItem = collItem
			}
			if isMap {
				if mapIncludes(ctx, coll.Items, keyNode) {
					onError(keyStart, "DUPLICATE_KEY", "Map keys must be unique", false)
				}
				coll.Items = append(coll.Items, pair)
			} else {
				yamlMap := ctx.newCollection("YAMLMap")
				yamlMap.Flow = true
				yamlMap.Items = append(yamlMap.Items, pair)
				endRange := keyNode.Range
				if valueNode != nil {
					endRange = valueNode.Range
				}
				yamlMap.Range = []int{keyNode.Range[0], endRange[1], endRange[2]}
				coll.Items = append(coll.Items, yamlMap)
			}
			if valueNode != nil {
				offset = valueNode.Range[2]
			} else {
				offset = valueProps.end
			}
		}
	}
	expectedEnd := "]"
	if isMap {
		expectedEnd = "}"
	}
	// `const [ce, ...ee] = fc.end`
	var ce *cst.Token
	ee := []*cst.Token{}
	if len(fc.End) > 0 {
		ce = fc.End[0]
		ee = append(ee, fc.End[1:]...)
	}
	cePos := offset
	if ce != nil && equalsASCII(ce.Source, expectedEnd) {
		cePos = ce.Offset + len(ce.Source)
	} else {
		name := strings.ToUpper(fcName[:1]) + fcName[1:]
		code := "BAD_INDENT"
		msg := name + " in block collection must be sufficiently indented and end with a " + expectedEnd
		if atRoot {
			code = "MISSING_CHAR"
			msg = name + " must end with a " + expectedEnd
		}
		onError(offset, code, msg, false)
		if ce != nil && len(ce.Source) != 1 {
			ee = append([]*cst.Token{ce}, ee...)
		}
	}
	if len(ee) > 0 {
		end := resolveEnd(ee, cePos, ctx.options.Strict, onError)
		if end.comment != "" {
			if coll.Comment != "" {
				coll.Comment += "\n" + end.comment
			} else {
				coll.Comment = end.comment
			}
		}
		coll.Range = []int{fc.Offset, cePos, end.offset}
	} else {
		coll.Range = []int{fc.Offset, cePos, cePos}
	}
	return coll
}

// containsNewline is upstream's, with its null for no key as false: every caller tests it for
// truthiness.
func containsNewline(key *cst.Token) bool {
	if key == nil {
		return false
	}
	switch key.Type {
	case "alias", "scalar", "double-quoted-scalar", "single-quoted-scalar":
		if includesUnit(key.Source, '\n') {
			return true
		}
		for _, st := range key.End {
			if st.Type == "newline" {
				return true
			}
		}
		return false
	case "flow-collection":
		for _, it := range key.Items {
			for _, st := range it.Start {
				if st.Type == "newline" {
					return true
				}
			}
			for _, st := range it.Sep {
				if st.Type == "newline" {
					return true
				}
			}
			if containsNewline(it.Key) || containsNewline(it.Value) {
				return true
			}
		}
		return false
	default:
		return true
	}
}

// mapIncludes is upstream's with uniqueKeys false or true; upstream's custom comparison function is
// not an option here.
func mapIncludes(ctx *composeContext, items []*Node, search *Node) bool {
	if !ctx.options.UniqueKeys {
		return false
	}
	for _, pair := range items {
		a := pair.Key
		if a == search || IsScalar(a) && IsScalar(search) && strictEquals(a.ScalarValue, search.ScalarValue) {
			return true
		}
	}
	return false
}

// strictEquals is `a === b` for scalar values: NaN is not itself, and objects are only themselves.
func strictEquals(a any, b any) bool {
	switch x := a.(type) {
	case float64:
		y, isNumber := b.(float64)
		return isNumber && x == y
	case Date, Binary:
		return false
	}
	switch b.(type) {
	case Date, Binary:
		return false
	}
	return a == b
}
