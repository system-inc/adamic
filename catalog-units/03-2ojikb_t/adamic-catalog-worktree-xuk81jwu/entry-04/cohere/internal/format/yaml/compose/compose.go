// Package compose is eemeli/yaml 2.9.0's composer, dist/compose/, ported to Go with the node classes it
// builds (dist/nodes/), the parts of dist/doc/ it runs (the Document constructor and setSchema, the
// Directives), and the schemas it resolves tags against (dist/schema/: the core schema, the common map,
// seq, string and null tags, and the YAML 1.1 schema, which a %YAML 1.1 directive selects and six of
// whose tags the core schema reaches as known tags). It turns the CST tokens of
// internal/format/yaml/cst into Documents, the way yaml-unist-parser drives it:
//
//	lineCounter := cst.NewLineCounter()
//	tokens := slices.Collect(cst.NewParser(lineCounter.AddNewLine).Parse(text, false))
//	for document := range compose.NewComposer(compose.UnistParserOptions()).Compose(tokens, true, len(text)) {
//		// yaml-unist-parser throws on document.Errors[0]
//	}
//
// # Offsets are UTF-16
//
// Every offset, range and error position counts UTF-16 code units, as upstream's do and as the cst
// package's do. A node's Range is upstream's three numbers, and its SrcToken is the very *cst.Token it
// was composed from, so yaml-unist-parser's port can walk from a node back into the tokens.
//
// # Strings
//
// The text being composed is a JavaScript string, and values are computed over its code units
// ([]uint16) exactly as upstream computes them, lengths and all. What the package hands out (a scalar's
// Source and string value, comments, tags, anchors, messages) is a Go string: UTF-8, except that a lone
// surrogate, which a double-quoted escape such as "\ud800" can make and UTF-8 cannot hold, is written as
// the three bytes its code point would take (WTF-8), so the JavaScript string can be recovered whole.
// UnitLength measures such a string as JavaScript does.
//
// # Nodes
//
// Upstream's Scalar, Alias, YAMLMap, YAMLSeq and Pair are one struct here, Node, told apart by Kind,
// and the identity checks IsScalar, IsMap and the rest read it. JavaScript's undefined and null
// properties map to Go's zero values where upstream never assigns the zero value itself (see Node).
//
// # Not ported
//
//   - Stringify, toJS and toJSON, the document API (Document's methods past the constructor,
//     Collection's get/set/add, YAMLMap.from, createNode, visit, the anchors helpers), applyReviver,
//     and errors.js's prettifyError: composing never calls them. That includes the merge key's
//     addToJSMap and isMergeKey, which only toJS reads.
//   - The compat schema option and util-flow-indent-check.js, which only it reaches; customTags,
//     sortMapEntries, the json and failsafe schemas, and intAsBigInt (so integers are float64, as
//     upstream's are without it). yaml-unist-parser sets none of these.
//   - uniqueKeys as a comparison function: UniqueKeys is a bool with upstream's default comparison.
//   - Composer.streamInfo and the LOG_STREAM logging.
//
// # Where the port differs
//
//   - Upstream wraps composeCollection in a try/catch for stack overflows and reports one as
//     RESOURCE_EXHAUSTION. Go's stack grows, so a collection nested deeper than V8's stack allows
//     composes here where upstream gives up. The one other exception that path catches, a TypeError
//     on a flow collection comment, is raised and caught the same way.
package compose
