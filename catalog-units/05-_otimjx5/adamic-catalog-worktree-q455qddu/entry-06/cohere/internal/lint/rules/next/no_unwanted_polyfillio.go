package next

import (
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/imports"
	"github.com/system-inc/cohere/internal/lint/ecmascript/jsx"
	"github.com/system-inc/cohere/internal/lint/ecmascript/text"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageNoUnwantedPolyfillioSecurity = rule.Message{
	Id: "noUnwantedPolyfillioSecurity",
	Description: "This script loads from polyfill.io, whose domain was sold in 2024 and then used " +
		"to serve malicious code to the sites embedding it. Every visitor runs whatever that " +
		"domain returns, so this is a live supply chain risk rather than a stale dependency. " +
		"Point it at https://cdnjs.cloudflare.com/polyfill/ or drop the polyfill and use the " +
		"browser feature directly.",
}

// messageNoUnwantedPolyfillioDuplicate is rendered per finding because it names the features it
// found, so the Description here is the shape and the report site fills it in. The Id is what
// fixtures assert and it stays constant.
var messageNoUnwantedPolyfillioDuplicate = rule.Message{
	Id: "noUnwantedPolyfillioDuplicate",
	Description: "This script asks a polyfill service for features Next.js already ships, so the " +
		"bytes are downloaded and parsed to redefine what is present.",
}

// NoUnwantedPolyfillio flags a script that loads polyfills from polyfill.io or duplicates one.
//
//	valid:   <script src='https://cdnjs.cloudflare.com/polyfill/v3/polyfill.min.js?features=AbortController' />
//	invalid: <script src='https://polyfill.io/v3/polyfill.min.js' />
//	invalid: <script src='https://cdnjs.cloudflare.com/polyfill/v3/polyfill.min.js?features=Promise' />
//
// Ported from `@next/next/no-unwanted-polyfillio`, read against oxc's `no_unwanted_polyfillio.rs`.
//
// # Two rules under one name, and only the second matches the name
//
// The first arm is a pure blocklist on the URL prefix, for the 2024 compromise, and it never looks
// at the feature list at all. It reports on the domain alone and returns, so
// `https://cdn.polyfill.io/v2/polyfill.min.js?features=Promise` says "security risk" and never says
// "Promise is already shipped" even though it would have qualified. Two of upstream's four failing
// cases exercise only this arm, so a port that folded the arms into one code path and lost the
// distinct message would still pass half the corpus.
//
// The second arm is the duplicate check, and it runs only on the three domains that replaced
// polyfill.io after the compromise.
//
// # What the corpus does not pin, measured on the release binary instead
//
// Upstream tests each gate broken open and none broken shut, so several of the decisions below have
// no fixture upstream and were established by running oxlint rather than by reading. Each was run
// with `no-sync-scripts` alongside as a control, because a config missing its `plugins` key makes
// every rule silent and a silent probe reads exactly like a decline.
//
// Feature matching is exact, case-sensitive equality on a whole comma-separated token:
// `features=promise` is silent. An unknown feature sitting beside a known one still reports, and the
// message names only the known one. `Array.prototype.@@iterator` carries a literal double at sign.
//
// Only %2C is decoded, uppercase only, and by a plain string replacement rather than a percent
// decode. `features=Promise%2cSet` is silent upstream because neither token survives as a whole
// name. That is a defect a port using Go's `url.QueryUnescape` would silently repair, which would be
// a divergence rather than an improvement, so the replacement is reproduced exactly.
//
// The prefixes are matched with the scheme attached, so `http://polyfill.io/v3/`, a
// protocol-relative `//cdn.polyfill.io/v2/`, and a path-only URL are all silent.
//
// # Resolving the tag name, and the shelf helper that would have got it wrong
//
// A lowercase `script` matches directly. Any other tag matches only if the file imports
// `next/script` and binds it to that name. oxc reads its module record's import entries and takes
// the local name of the first entry for that module, whatever kind of binding it is, so a default
// import, a named `import {Script}`, and a namespace `import * as S` all qualify. All three were
// confirmed reporting on the release binary, and `<Script />` with no such import was confirmed
// silent.
//
// `imports.LocalNameOfDefaultImport` is the obvious shelf reach here and it is the wrong one. Its
// doc comment invites a port to resolve the import properly and calls upstream's looser reading a
// defect, which is sound advice for the sibling rule it was written for and a behaviour change
// here: it answers only for a default binding, so the named and namespace forms above would go
// silent against a measured upstream that reports on both. `imports.BindingsOf` is used instead,
// which returns all three shapes, and the first-entry rule is reproduced by preferring the default
// binding and falling back in source order.
var NoUnwantedPolyfillio = rule.Rule{
	// No family prefix. The config writes `nextjs/no-unwanted-polyfillio` and matching strips the
	// namespace on a `/` boundary, so a prefixed name matches nothing and runs on no files while
	// its own tests pass.
	Name: "@next/next/no-unwanted-polyfillio",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		// The local name `next/script` is bound to in this file, empty when it is not imported.
		//
		// Filled by the import listener and read by the JSX listener. The walk is a single
		// pre-order pass, so an import at the top of the file is visited before the JSX below it.
		// An import written after the JSX would not be seen; upstream reads a prebuilt module
		// record and does not have that hazard, but it is unreachable in practice because an
		// import declaration is only legal at the top level and JSX referencing the binding before
		// it would be a use before definition.
		nextScriptLocalName := ""

		recordImport := func(node *ast.Node) {
			if nextScriptLocalName != "" {
				return
			}
			declaration := node.AsImportDeclaration()
			if declaration == nil || declaration.ModuleSpecifier == nil {
				return
			}
			if !ast.IsStringLiteralLike(declaration.ModuleSpecifier) {
				return
			}
			if declaration.ModuleSpecifier.Text() != "next/script" {
				return
			}

			// oxc takes the first import entry for the module whatever its kind. A single
			// declaration writes its default binding first, then its named or namespace ones, so
			// preferring the default and otherwise taking the first in source order is that order.
			bindings := imports.BindingsOf(node)
			candidates := append([]*ast.Node{bindings.Default, bindings.Namespace}, bindings.Named...)
			for _, binding := range candidates {
				name := localBindingName(binding)
				if name != "" {
					nextScriptLocalName = name
					return
				}
			}
		}

		report := func(node *ast.Node) {
			tagName, attributes := jsx.ElementParts(node)
			if !jsx.IsIntrinsicElementNamed(tagName, "script") {
				// A component tag qualifies only by matching the local name of a `next/script`
				// import. With no such import there is nothing to match, and an empty name must not
				// match an unnamed tag, so the check is written to decline first.
				if nextScriptLocalName == "" {
					return
				}
				if tagName == nil || tagName.Kind != ast.KindIdentifier || tagName.Text() != nextScriptLocalName {
					return
				}
			}

			source, sourceNode := stringAttributeWithNode(attributes, "src")
			if sourceNode == nil {
				return
			}

			// The compromised domains, checked before anything else and reported on their own. The
			// feature list is never consulted on these, which is why the arm returns rather than
			// falling through.
			if strings.HasPrefix(source, "https://cdn.polyfill.io/v2/") || strings.HasPrefix(source, "https://polyfill.io/v3/") {
				ctx.ReportNode(sourceNode, messageNoUnwantedPolyfillioSecurity)
				return
			}

			if !strings.HasPrefix(source, "https://polyfill-fastly.net/") &&
				!strings.HasPrefix(source, "https://polyfill-fastly.io/") &&
				!strings.HasPrefix(source, "https://cdnjs.cloudflare.com/polyfill/") {
				return
			}

			features, hasFeatures := urlQueryValue(source, "features")
			if !hasFeatures {
				return
			}

			// Upstream replaces this one sequence rather than percent-decoding, so a lowercase
			// %2c survives into the token and matches nothing. Reproduced deliberately; see the
			// doc comment above.
			features = strings.ReplaceAll(features, "%2C", ",")

			var unwanted []string
			for _, feature := range strings.Split(features, ",") {
				if _, shipped := nextPolyfilledFeatures[feature]; shipped {
					unwanted = append(unwanted, feature)
				}
			}
			if len(unwanted) == 0 {
				return
			}

			// Order follows the URL rather than the feature set, and the verb agrees with the
			// count. Both are asserted by fixtures because both are visible in upstream's snapshot.
			verb := "is"
			if len(unwanted) > 1 {
				verb = "are"
			}
			message := messageNoUnwantedPolyfillioDuplicate
			named := strings.Join(unwanted, ", ")
			message.Description = fmt.Sprintf(
				"This script asks a polyfill service for %s, which Next.js already ships, so the "+
					"bytes are downloaded and parsed to redefine what %s already there. Drop %s "+
					"from the `features` list.",
				named, verb, named,
			)
			ctx.ReportNode(sourceNode, message)
		}

		return rule.Listeners{
			ast.KindImportDeclaration:     recordImport,
			ast.KindJsxOpeningElement:     report,
			ast.KindJsxSelfClosingElement: report,
		}
	},
}

// localBindingName reads the identifier an import binding introduces into scope.
//
// A namespace import and a default import both carry their local name as the node itself. A named
// specifier carries it in `Name()`, which is the alias when one is written and the imported name
// when it is not, so the aliased and unaliased forms fall out of the same read.
func localBindingName(binding *ast.Node) string {
	if binding == nil {
		return ""
	}
	switch binding.Kind {
	case ast.KindIdentifier:
		return binding.Text()
	case ast.KindNamespaceImport:
		name := binding.AsNamespaceImport().Name()
		if name == nil || name.Kind != ast.KindIdentifier {
			return ""
		}
		return name.Text()
	case ast.KindImportSpecifier:
		name := binding.AsImportSpecifier().Name()
		if name == nil || name.Kind != ast.KindIdentifier {
			return ""
		}
		return name.Text()
	}
	return ""
}

// stringAttributeWithNode returns both the text of a string-literal attribute and the attribute node.
//
// `jsx.StringAttributeValue` answers the text alone, and this rule underlines the attribute rather
// than the element, so it needs the node too. The shelf comment on `jsx.AttributeName` already
// records that a rule reporting on the attribute has no element-level helper to reach for; the name
// guard is shared with the shelf so the two cannot drift on what counts as a named attribute.
//
// A `src` that is present but not a plain string literal stops the search rather than continuing
// past it, matching oxc, which finds the attribute first and only then asks whether its value is a
// string literal.
func stringAttributeWithNode(attributes *ast.Node, name string) (string, *ast.Node) {
	if attributes == nil || attributes.Kind != ast.KindJsxAttributes {
		return "", nil
	}
	properties := attributes.AsJsxAttributes().Properties
	if properties == nil {
		return "", nil
	}

	for _, property := range properties.Nodes {
		attributeName, named := jsx.AttributeName(property)
		if !named || attributeName != name {
			continue
		}
		attribute := property.AsJsxAttribute()
		if attribute.Initializer == nil || attribute.Initializer.Kind != ast.KindStringLiteral {
			return "", nil
		}
		// Decoded as ESLint's parser decodes it, so an entity in the URL reads as the character it
		// names, as it does upstream.
		return text.UnescapeStringLiteralText(attribute.Initializer.Text()), property
	}

	return "", nil
}

// nextPolyfilledFeatures is the feature list Next.js already ships, transcribed mechanically from
// oxc's `NEXT_POLYFILLED_FEATURES` so the shapes a hand copy corrupts survive.
//
// Three of those shapes are load-bearing. `Array.prototype.@@iterator` and
// `String.prototype.@@iterator` carry a literal double at sign. `Number.EPSILON` and
// `Number.Epsilon` are both present on purpose, upstream compensating for a real world misspelling,
// so removing the duplicate would change behaviour. The `es5` through `es2019` aliases and `fetch`
// are the only lowercase entries and matching is case-sensitive.
//
// eslint's copy of this list also carries `Object.hasOwn`; oxc's does not, and oxc is what the gate
// compares against, so it is deliberately absent here rather than overlooked.
var nextPolyfilledFeatures = map[string]struct{}{
	"Array.from":                       {},
	"Array.of":                         {},
	"Array.prototype.@@iterator":       {},
	"Array.prototype.at":               {},
	"Array.prototype.copyWithin":       {},
	"Array.prototype.fill":             {},
	"Array.prototype.find":             {},
	"Array.prototype.findIndex":        {},
	"Array.prototype.flat":             {},
	"Array.prototype.flatMap":          {},
	"Array.prototype.includes":         {},
	"Function.prototype.name":          {},
	"Map":                              {},
	"Number.EPSILON":                   {},
	"Number.Epsilon":                   {},
	"Number.MAX_SAFE_INTEGER":          {},
	"Number.MIN_SAFE_INTEGER":          {},
	"Number.isFinite":                  {},
	"Number.isInteger":                 {},
	"Number.isNaN":                     {},
	"Number.isSafeInteger":             {},
	"Number.parseFloat":                {},
	"Number.parseInt":                  {},
	"Object.assign":                    {},
	"Object.entries":                   {},
	"Object.fromEntries":               {},
	"Object.getOwnPropertyDescriptor":  {},
	"Object.getOwnPropertyDescriptors": {},
	"Object.is":                        {},
	"Object.keys":                      {},
	"Object.values":                    {},
	"Promise":                          {},
	"Promise.prototype.finally":        {},
	"Reflect":                          {},
	"Set":                              {},
	"String.fromCodePoint":             {},
	"String.prototype.@@iterator":      {},
	"String.prototype.codePointAt":     {},
	"String.prototype.endsWith":        {},
	"String.prototype.includes":        {},
	"String.prototype.padEnd":          {},
	"String.prototype.padStart":        {},
	"String.prototype.repeat":          {},
	"String.prototype.startsWith":      {},
	"String.prototype.trimEnd":         {},
	"String.prototype.trimStart":       {},
	"String.raw":                       {},
	"Symbol":                           {},
	"Symbol.asyncIterator":             {},
	"URL":                              {},
	"URL.prototype.toJSON":             {},
	"URLSearchParams":                  {},
	"WeakMap":                          {},
	"WeakSet":                          {},
	"es2015":                           {},
	"es2016":                           {},
	"es2017":                           {},
	"es2018":                           {},
	"es2019":                           {},
	"es5":                              {},
	"es6":                              {},
	"es7":                              {},
	"fetch":                            {},
}
