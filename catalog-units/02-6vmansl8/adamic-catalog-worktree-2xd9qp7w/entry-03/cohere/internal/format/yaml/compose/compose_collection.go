package compose

// Ported from eemeli/yaml 2.9.0, dist/compose/compose-collection.js and compose-scalar.js.

import "github.com/system-inc/cohere/internal/format/yaml/cst"

func resolveCollection(ctx *composeContext, token *cst.Token, onError onErrorFunc, tagName string, tag *Tag) *Node {
	var coll *Node
	switch token.Type {
	case "block-map":
		coll = resolveBlockMap(ctx, token, onError, tag)
	case "block-seq":
		coll = resolveBlockSeq(ctx, token, onError, tag)
	default:
		coll = resolveFlowCollection(ctx, token, onError, tag)
	}
	// If we got a tagName matching the class, or the tag name is '!',
	// then use the tagName from the node class used to create it.
	if tagName == "!" || tagName == collectionTagName(coll) {
		coll.Tag = collectionTagName(coll)
		return coll
	}
	if tagName != "" {
		coll.Tag = tagName
	}
	return coll
}

func composeCollection(ctx *composeContext, token *cst.Token, props props, onError onErrorFunc) *Node {
	tagToken := props.tag
	tagName := ""
	if tagToken != nil {
		tagName = ctx.directives.tagName(tagToken.Source, func(msg string) {
			onError(tagToken, "TAG_RESOLVE_FAILED", msg, false)
		})
	}
	if token.Type == "block-seq" {
		anchor, nl := props.anchor, props.newlineAfterProp
		var lastProp *cst.Token
		if anchor != nil && tagToken != nil {
			if anchor.Offset > tagToken.Offset {
				lastProp = anchor
			} else {
				lastProp = tagToken
			}
		} else if anchor != nil {
			lastProp = anchor
		} else {
			lastProp = tagToken
		}
		if lastProp != nil && (nl == nil || nl.Offset < lastProp.Offset) {
			message := "Missing newline after block sequence props"
			onError(lastProp, "MISSING_CHAR", message, false)
		}
	}
	expType := "seq"
	if token.Type == "block-map" || token.Type == "flow-collection" && equalsASCII(token.FlowStart.Source, "{") {
		expType = "map"
	}
	// shortcut: check if it's a generic YAMLMap or YAMLSeq
	// before jumping into the custom tag logic.
	if tagToken == nil || tagName == "" || tagName == "!" || tagName == mapTagName && expType == "map" ||
		tagName == seqTagName && expType == "seq" {
		return resolveCollection(ctx, token, onError, tagName, nil)
	}
	var tag *Tag
	for _, t := range ctx.schema.Tags {
		if t.Tag == tagName && t.Collection == expType {
			tag = t
			break
		}
	}
	if tag == nil {
		kt := knownTag(ctx.schema, tagName)
		if kt != nil && kt.Collection == expType {
			copied := copyTag(kt)
			copied.Default = defaultFalse
			ctx.schema.Tags = append(ctx.schema.Tags, copied)
			tag = kt
		} else {
			if kt != nil {
				expects := kt.Collection
				if expects == "" {
					expects = "scalar"
				}
				onError(tagToken, "BAD_COLLECTION_TYPE", kt.tagString()+" used for "+expType+
					" collection, but expects "+expects, true)
			} else {
				onError(tagToken, "TAG_RESOLVE_FAILED", "Unresolved tag: "+tagName, true)
			}
			return resolveCollection(ctx, token, onError, tagName, nil)
		}
	}
	coll := resolveCollection(ctx, token, onError, tagName, tag)
	// `tag.resolve?.(coll, onError, ctx.options) ?? coll`: every collection tag has a resolve, and
	// each returns a node, so `isNode(res) ? res : new Scalar(res)` is res.
	node := coll
	if tag.resolveCollection != nil {
		node = tag.resolveCollection(coll, func(msg string) {
			onError(tagToken, "TAG_RESOLVE_FAILED", msg, false)
		})
	}
	node.Range = coll.Range
	node.Tag = tagName
	if tag.Format != "" {
		node.Format = tag.Format
	}
	return node
}

// objectPrototypeKeys are the properties every plain JavaScript object inherits. Upstream looks a tag
// name up in knownTags, a plain object, so a tag that resolves to one of these names finds the inherited
// function (or, for __proto__, Object.prototype) instead of nothing.
var objectPrototypeKeys = map[string]bool{
	"__proto__": true, "__defineGetter__": true, "__defineSetter__": true, "__lookupGetter__": true,
	"__lookupSetter__": true, "constructor": true, "hasOwnProperty": true, "isPrototypeOf": true,
	"propertyIsEnumerable": true, "toLocaleString": true, "toString": true, "valueOf": true,
}

// knownTag is `schema.knownTags[tagName]`. An inherited property is a Tag with prototypeProperty set:
// no tag, no collection, no resolve.
func knownTag(schema *Schema, tagName string) *Tag {
	if tag, found := schema.KnownTags[tagName]; found {
		return tag
	}
	if objectPrototypeKeys[tagName] {
		return &Tag{prototypeProperty: true}
	}
	return nil
}

// tagString is `${tag.tag}`: "undefined" for an inherited property, which has no tag.
func (tag *Tag) tagString() string {
	if tag.prototypeProperty {
		return "undefined"
	}
	return tag.Tag
}

type scalarResolution struct {
	value      []uint16
	scalarType string
	comment    string
	valueRange []int
}

func composeScalar(ctx *composeContext, token *cst.Token, tagToken *cst.Token, onError onErrorFunc) *Node {
	var resolved scalarResolution
	if token.Type == "block-scalar" {
		resolved = resolveBlockScalar(ctx, token, onError)
	} else {
		resolved = resolveFlowScalar(token, ctx.options.Strict, onError)
	}
	value := resolved.value
	tagName := ""
	if tagToken != nil {
		tagName = ctx.directives.tagName(tagToken.Source, func(msg string) {
			onError(tagToken, "TAG_RESOLVE_FAILED", msg, false)
		})
	}
	var tag *Tag
	if ctx.options.StringKeys && ctx.atKey {
		tag = ctx.schema.scalarTag()
	} else if tagName != "" {
		tag = findScalarTagByName(ctx.schema, value, tagName, tagToken, onError)
	} else if token.Type == "scalar" {
		tag = findScalarTagByTest(ctx, value, token, onError)
	} else {
		tag = ctx.schema.scalarTag()
	}
	// `tagToken ?? token`
	var errorSource any = token
	if tagToken != nil {
		errorSource = tagToken
	}
	var scalar *Node
	if tag.resolveScalar == nil {
		// An inherited property found by knownTag: `tag.resolve(...)` throws a TypeError.
		onError(errorSource, "TAG_RESOLVE_FAILED", "tag.resolve is not a function", false)
		scalar = ctx.newScalar(unitsToString(value))
	} else {
		res, err := tag.resolveScalar(value, func(msg string) {
			onError(errorSource, "TAG_RESOLVE_FAILED", msg, false)
		}, ctx.options)
		if err != nil {
			onError(errorSource, "TAG_RESOLVE_FAILED", err.Error(), false)
			scalar = ctx.newScalar(unitsToString(value))
		} else if node, isNode := res.(*Node); isNode && IsScalar(node) {
			scalar = node
		} else {
			scalar = ctx.newScalar(res)
		}
	}
	scalar.Range = resolved.valueRange
	scalar.Source = unitsToString(value)
	scalar.sourceSet = true
	if resolved.scalarType != "" {
		scalar.Type = resolved.scalarType
	}
	if tagName != "" {
		scalar.Tag = tagName
	}
	if tag.Format != "" {
		scalar.Format = tag.Format
	}
	if resolved.comment != "" {
		scalar.Comment = resolved.comment
	}
	return scalar
}

func findScalarTagByName(schema *Schema, value []uint16, tagName string, tagToken *cst.Token, onError onErrorFunc) *Tag {
	if tagName == "!" {
		return schema.scalarTag() // non-specific tag
	}
	var matchWithTest []*Tag
	for _, tag := range schema.Tags {
		if tag.Collection == "" && !tag.prototypeProperty && tag.Tag == tagName {
			if tag.Default != defaultFalse && tag.Test != nil {
				matchWithTest = append(matchWithTest, tag)
			} else {
				return tag
			}
		}
	}
	for _, tag := range matchWithTest {
		if tag.test(value) {
			return tag
		}
	}
	kt := knownTag(schema, tagName)
	if kt != nil && kt.Collection == "" {
		// Ensure that the known tag is available for stringifying,
		// but does not get used by default.
		copied := copyTag(kt)
		copied.Default = defaultFalse
		copied.Test = nil
		schema.Tags = append(schema.Tags, copied)
		return kt
	}
	onError(tagToken, "TAG_RESOLVE_FAILED", "Unresolved tag: "+tagName, tagName != strTagName)
	return schema.scalarTag()
}

func findScalarTagByTest(ctx *composeContext, value []uint16, _ *cst.Token, _ onErrorFunc) *Tag {
	// The view every test reads is made once, at the first tag with a test, rather than per tag (#vbjv3d6).
	view, viewMade := "", false
	for _, tag := range ctx.schema.Tags {
		if (tag.Default == defaultTrue || ctx.atKey && tag.Default == defaultKey) && tag.Test != nil {
			if !viewMade {
				view, viewMade = asciiView(value), true
			}
			if tag.Test.MatchString(view) {
				return tag
			}
		}
	}
	// schema.compat is never set on this path, so there is nothing to warn about.
	return ctx.schema.scalarTag()
}
