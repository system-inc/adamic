package compose

// Ported from eemeli/yaml 2.9.0, dist/schema/Schema.js and dist/schema/tags.js: the schema a Document
// builds for its version, with the tags composing resolves against. Only what composing reads is here:
// the tags' tag, collection, default, test, format, nodeClass and resolve. identify, stringify and
// createNode serve stringify and createNode, and sortMapEntries and toStringOptions serve stringify.

import "regexp"

// tagDefault is a tag's `default`, which is true, false or 'key' upstream.
type tagDefault int

const (
	defaultFalse tagDefault = iota
	defaultTrue
	// defaultKey is 'key': the merge tag, tested only at a mapping key.
	defaultKey
)

// Tag is a schema tag, upstream's ScalarTag or CollectionTag.
type Tag struct {
	Tag string
	// Collection is "map" or "seq" for a collection tag, "" for a scalar tag (undefined).
	Collection string
	Default    tagDefault
	// Test is the tag's test regular expression, nil when it has none. It is matched against the
	// asciiView of the value, which every test's ASCII-only classes read the same way.
	Test   *regexp.Regexp
	Format string
	// NodeClass is the collection class the tag constructs, "" when it has none.
	NodeClass string

	// resolveScalar is a scalar tag's resolve. It returns a *Node when upstream's returns a Scalar,
	// otherwise the raw value, and an error where upstream's throws.
	resolveScalar func(value []uint16, onError func(message string), options *Options) (any, error)
	// resolveCollection is a collection tag's resolve.
	resolveCollection func(collection *Node, onError func(message string)) *Node

	// prototypeProperty marks what knownTag finds for a tag named like a property every JavaScript
	// object inherits: something with no tag, collection, test or resolve.
	prototypeProperty bool
}

// test is `tag.test?.test(value)`.
func (tag *Tag) test(value []uint16) bool {
	return tag.Test != nil && tag.Test.MatchString(asciiView(value))
}

// Schema is a Document's schema.
type Schema struct {
	Name string
	// KnownTags are the tags a tag name can reach without being in Tags: the core schema's
	// coreKnownTags, none for YAML 1.1.
	KnownTags map[string]*Tag
	// Tags is this schema's own list, which composing appends known tags to as it meets them.
	Tags []*Tag
}

// Upstream's schema[MAP], schema[SCALAR] and schema[SEQ] are the common map, string and seq tags,
// whatever the schema.
func (schema *Schema) scalarTag() *Tag { return stringTag }

// schemaOptions is the part of upstream's Schema constructor options composing sets.
type schemaOptions struct {
	merge            bool
	resolveKnownTags bool
	schema           string
}

// newSchema is `new Schema({ merge, resolveKnownTags, schema })`. compat, customTags and sortMapEntries
// are never set on this path, so this.compat is null.
func newSchema(options schemaOptions) *Schema {
	name := options.schema
	if name == "" {
		name = "core"
	}
	schema := &Schema{Name: name, KnownTags: map[string]*Tag{}}
	if options.resolveKnownTags {
		schema.KnownTags = coreKnownTags()
	}
	schema.Tags = getTags(name, options.merge)
	return schema
}

// coreKnownTags is tags.js's coreKnownTags.
func coreKnownTags() map[string]*Tag {
	return map[string]*Tag{
		"tag:yaml.org,2002:binary":    binaryTag,
		"tag:yaml.org,2002:merge":     mergeTag,
		"tag:yaml.org,2002:omap":      omapTag,
		"tag:yaml.org,2002:pairs":     pairsTag,
		"tag:yaml.org,2002:set":       setTag,
		"tag:yaml.org,2002:timestamp": timestampTag,
	}
}

// schemaTags is tags.js's schemas map, for the two schemas a Document picks: core for YAML 1.2 and
// yaml-1.1 for YAML 1.1.
func schemaTags(name string) []*Tag {
	switch name {
	case "core":
		return coreSchema
	case "yaml-1.1":
		return yaml11Schema
	}
	panic("compose: unknown schema " + name)
}

// getTags is upstream's for no customTags: a fresh copy of the schema's list, with the merge tag
// appended when asked for and missing.
func getTags(schemaName string, addMergeTag bool) []*Tag {
	tags := schemaTags(schemaName)
	result := make([]*Tag, len(tags), len(tags)+8)
	copy(result, tags)
	if addMergeTag && !containsTag(tags, mergeTag) {
		result = append(result, mergeTag)
	}
	return result
}

func containsTag(tags []*Tag, tag *Tag) bool {
	for _, each := range tags {
		if each == tag {
			return true
		}
	}
	return false
}

// copyTag is `Object.assign({}, tag)`.
func copyTag(tag *Tag) *Tag {
	copied := *tag
	return &copied
}

// The tag names composing compares against.
const (
	mapTagName  = "tag:yaml.org,2002:map"
	seqTagName  = "tag:yaml.org,2002:seq"
	strTagName  = "tag:yaml.org,2002:str"
	omapTagName = "tag:yaml.org,2002:omap"
	setTagName  = "tag:yaml.org,2002:set"
)

// Ported from dist/schema/common/: map.js, seq.js, string.js, null.js.

var mapTag = &Tag{
	Collection: "map",
	Default:    defaultTrue,
	NodeClass:  "YAMLMap",
	Tag:        mapTagName,
	resolveCollection: func(collection *Node, onError func(message string)) *Node {
		if !IsMap(collection) {
			onError("Expected a mapping for this tag")
		}
		return collection
	},
}

var seqTag = &Tag{
	Collection: "seq",
	Default:    defaultTrue,
	NodeClass:  "YAMLSeq",
	Tag:        seqTagName,
	resolveCollection: func(collection *Node, onError func(message string)) *Node {
		if !IsSeq(collection) {
			onError("Expected a sequence for this tag")
		}
		return collection
	},
}

var stringTag = &Tag{
	Default: defaultTrue,
	Tag:     strTagName,
	resolveScalar: func(value []uint16, _ func(string), _ *Options) (any, error) {
		return unitsToString(value), nil
	},
}

var nullTag = &Tag{
	Default: defaultTrue,
	Tag:     "tag:yaml.org,2002:null",
	Test:    regexp.MustCompile(`^(?:~|[Nn]ull|NULL)?$`),
	resolveScalar: func([]uint16, func(string), *Options) (any, error) {
		return newScalar(nil), nil
	},
}

// Ported from dist/schema/core/: bool.js, int.js, float.js, schema.js.

var boolTag = &Tag{
	Default: defaultTrue,
	Tag:     "tag:yaml.org,2002:bool",
	Test:    regexp.MustCompile(`^(?:[Tt]rue|TRUE|[Ff]alse|FALSE)$`),
	resolveScalar: func(value []uint16, _ func(string), _ *Options) (any, error) {
		first := characterAt(value, 0)
		return newScalar(first == 't' || first == 'T'), nil
	},
}

// coreIntResolve is int.js's intResolve: `parseInt(str.substring(offset), radix)`. intAsBigInt is never
// set on this path.
func coreIntResolve(value []uint16, offset int, radix int) float64 {
	return parseInt(substring(value, offset, len(value)), radix)
}

var coreIntOctTag = &Tag{
	Default: defaultTrue,
	Tag:     "tag:yaml.org,2002:int",
	Format:  "OCT",
	Test:    regexp.MustCompile(`^0o[0-7]+$`),
	resolveScalar: func(value []uint16, _ func(string), _ *Options) (any, error) {
		return coreIntResolve(value, 2, 8), nil
	},
}

var coreIntTag = &Tag{
	Default: defaultTrue,
	Tag:     "tag:yaml.org,2002:int",
	Test:    regexp.MustCompile(`^[-+]?[0-9]+$`),
	resolveScalar: func(value []uint16, _ func(string), _ *Options) (any, error) {
		return coreIntResolve(value, 0, 10), nil
	},
}

var coreIntHexTag = &Tag{
	Default: defaultTrue,
	Tag:     "tag:yaml.org,2002:int",
	Format:  "HEX",
	Test:    regexp.MustCompile(`^0x[0-9a-fA-F]+$`),
	resolveScalar: func(value []uint16, _ func(string), _ *Options) (any, error) {
		return coreIntResolve(value, 2, 16), nil
	},
}

// floatNaNResolve is floatNaN's resolve, the same in both schemas.
func floatNaNResolve(value []uint16, _ func(string), _ *Options) (any, error) {
	last := slice(value, -3, len(value))
	if len(last) == 3 && toLowerASCII(last) == "nan" {
		return nan(), nil
	}
	if characterAt(value, 0) == '-' {
		return negativeInfinity(), nil
	}
	return positiveInfinity(), nil
}

// toLowerASCII is `text.toLowerCase()` for text whose comparison is against an ASCII literal: any
// non-ASCII unit stays non-ASCII, so it can never compare equal.
func toLowerASCII(text []uint16) string {
	lowered := []byte(asciiView(text))
	for index, character := range lowered {
		if character >= 'A' && character <= 'Z' {
			lowered[index] = character + 'a' - 'A'
		}
	}
	return string(lowered)
}

var coreFloatNaNTag = &Tag{
	Default:       defaultTrue,
	Tag:           "tag:yaml.org,2002:float",
	Test:          regexp.MustCompile(`^(?:[-+]?\.(?:inf|Inf|INF)|\.nan|\.NaN|\.NAN)$`),
	resolveScalar: floatNaNResolve,
}

var coreFloatExpTag = &Tag{
	Default: defaultTrue,
	Tag:     "tag:yaml.org,2002:float",
	Format:  "EXP",
	Test:    regexp.MustCompile(`^[-+]?(?:\.[0-9]+|[0-9]+(?:\.[0-9]*)?)[eE][-+]?[0-9]+$`),
	resolveScalar: func(value []uint16, _ func(string), _ *Options) (any, error) {
		return parseFloat(value), nil
	},
}

var coreFloatTag = &Tag{
	Default: defaultTrue,
	Tag:     "tag:yaml.org,2002:float",
	Test:    regexp.MustCompile(`^[-+]?(?:\.[0-9]+|[0-9]+\.[0-9]*)$`),
	resolveScalar: func(value []uint16, _ func(string), _ *Options) (any, error) {
		node := newScalar(parseFloat(value))
		dot := indexOfUnit(value, '.')
		if dot != -1 && characterAt(value, len(value)-1) == '0' {
			node.MinFractionDigits = len(value) - dot - 1
		}
		return node, nil
	},
}

func indexOfUnit(text []uint16, character uint16) int {
	for index, unit := range text {
		if unit == character {
			return index
		}
	}
	return -1
}

var coreSchema = []*Tag{
	mapTag,
	seqTag,
	stringTag,
	nullTag,
	boolTag,
	coreIntOctTag,
	coreIntTag,
	coreIntHexTag,
	coreFloatNaNTag,
	coreFloatExpTag,
	coreFloatTag,
}
