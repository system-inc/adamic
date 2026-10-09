package notices

// Reach is where an upstream's work reaches a user.
type Reach string

const (
	// Compiled is code ported or vendored into cohere's source and compiled into the binary.
	Compiled Reach = "Compiled"

	// RepositoryOnly is kept in the repository for tests and tools, and ships in no package.
	RepositoryOnly Reach = "RepositoryOnly"
)

// Upstream is a project whose code or data cohere carries in its own source, which no module graph
// can find: a port, a vendored copy, or data captured from it.
type Upstream struct {
	Name    string
	URL     string
	License string

	// Version is the version ported, when the port names one. Empty for code ported from the project
	// over time rather than from one release.
	Version string

	// Carries says what in cohere holds the work.
	Carries string

	Reach Reach

	// LicenseFile is the license text under licenses/, taken at Version when there is one, since a
	// project can relicense later releases (postcss-values-parser moved from MIT to MPL-2.0 at 3.0).
	LicenseFile string

	// Note stands in for the text when the project publishes none, saying what it declares and who it
	// names. Exactly one of LicenseFile and Note is set; a copyright line nobody published is never made up.
	Note string
}

// Upstreams are the curated projects, in the order they are written: vendored code first, then the
// formatter's ports and captured data, then the linter's. The list was checked against a sweep of every
// shipped package's attribution comments (#8p2v6j6).
var Upstreams = []Upstream{
	{
		Name: "tsgolint", URL: "https://github.com/typescript-eslint/tsgolint", License: "MIT",
		Carries: "The shim generator behind TypeScript-shim/ and the checker helpers vendored in internal/lint/checking.",
		Reach:   Compiled, LicenseFile: "tsgolint.txt",
	},
	{
		Name: "rslint", URL: "https://github.com/web-infra-dev/rslint", License: "MIT",
		Carries: "The control-flow graph vendored in internal/lint/ecmascript/control_flow_graph.",
		Reach:   Compiled, LicenseFile: "rslint.txt",
	},
	{
		Name: "Prettier", URL: "https://github.com/prettier/prettier", License: "MIT",
		Carries: "The formatter in internal/format: its document model and printers are ported from Prettier's.",
		Reach:   Compiled, LicenseFile: "prettier.txt",
	},
	{
		Name: "yaml", URL: "https://github.com/eemeli/yaml", License: "ISC", Version: "2",
		Carries: "The YAML parser in internal/format/yaml, ported.",
		Reach:   Compiled, LicenseFile: "yaml.txt",
	},
	{
		Name: "yaml-unist-parser", URL: "https://github.com/prettier/yaml-unist-parser", License: "MIT", Version: "3",
		Carries: "The YAML syntax tree in internal/format/yaml/unist, ported.",
		Reach:   Compiled, LicenseFile: "yaml-unist-parser.txt",
	},
	{
		Name: "postcss-values-parser", URL: "https://github.com/shellscape/postcss-values-parser", License: "MIT", Version: "2.0.1",
		Carries: "The CSS value parser in internal/format/css/values, ported from 2.0.1, which was MIT.",
		Reach:   Compiled, LicenseFile: "postcss-values-parser.txt",
	},
	{
		Name: "micromark", URL: "https://github.com/micromark/micromark", License: "MIT",
		Carries: "The Markdown tokenizer in internal/format/markdown/micromark, ported.",
		Reach:   Compiled, LicenseFile: "micromark.txt",
	},
	{
		Name: "mdast-util-from-markdown", URL: "https://github.com/syntax-tree/mdast-util-from-markdown", License: "MIT",
		Carries: "The Markdown syntax tree builder in internal/format/markdown/mdast, ported.",
		Reach:   Compiled, LicenseFile: "mdast-util-from-markdown.txt",
	},
	{
		Name: "emoji-regex", URL: "https://github.com/mathiasbynens/emoji-regex", License: "MIT", Version: "10.6.0",
		Carries: "Emoji ranges captured into internal/format/doc/string_width_generated.go.",
		Reach:   Compiled, LicenseFile: "emoji-regex.txt",
	},
	{
		Name: "get-east-asian-width", URL: "https://github.com/sindresorhus/get-east-asian-width", License: "MIT", Version: "1.6.0",
		Carries: "Character widths captured into internal/format/doc/string_width_generated.go.",
		Reach:   Compiled, LicenseFile: "get-east-asian-width.txt",
	},
	{
		Name: "narrow-emojis", URL: "https://github.com/fisker/narrow-emojis", License: "MIT", Version: "0.0.3",
		Carries: "Narrow emoji ranges captured into internal/format/doc/string_width_generated.go.",
		Reach:   Compiled, LicenseFile: "narrow-emojis.txt",
	},
	{
		Name: "PostCSS", URL: "https://github.com/postcss/postcss", License: "MIT", Version: "8.5.16",
		Carries: "The CSS parser and tokenizer in internal/format/css/postcss, ported.",
		Reach:   Compiled, LicenseFile: "postcss.txt",
	},
	{
		Name: "postcss-scss", URL: "https://github.com/postcss/postcss-scss", License: "MIT", Version: "4.0.9",
		Carries: "The SCSS parser in internal/format/css/postcss, ported.",
		Reach:   Compiled, LicenseFile: "postcss-scss.txt",
	},
	{
		Name: "postcss-selector-parser", URL: "https://github.com/postcss/postcss-selector-parser", License: "MIT", Version: "2.2.3",
		Carries: "The selector parser in internal/format/css/selector, ported.",
		Reach:   Compiled, LicenseFile: "postcss-selector-parser.txt",
	},
	{
		Name: "flatten", URL: "https://github.com/mk-pmb/flatten-js", License: "MIT", Version: "1.0.3",
		Carries: "A helper of the selector parser, ported in internal/format/css/selector.",
		Reach:   Compiled, LicenseFile: "flatten.txt",
	},
	{
		Name: "indexes-of", URL: "https://github.com/dominictarr/indexes-of", License: "MIT", Version: "1.0.1",
		Carries: "A helper of the selector parser, ported in internal/format/css/selector.",
		Reach:   Compiled, LicenseFile: "indexes-of.txt",
	},
	{
		Name: "uniq", URL: "https://github.com/mikolalysenko/uniq", License: "MIT", Version: "1.0.1",
		Carries: "A helper of the selector parser, ported in internal/format/css/selector.",
		Reach:   Compiled, LicenseFile: "uniq.txt",
	},
	{
		Name: "postcss-media-query-parser", URL: "https://github.com/dryoma/postcss-media-query-parser", License: "MIT", Version: "0.2.3",
		Carries: "The media query parser in internal/format/css/mediaquery, ported.",
		Reach:   Compiled, Note: "No license text is published with this package or in its repository. Its package.json declares MIT, and names dryoma as its author.",
	},
	{
		Name: "css-units-list", URL: "https://github.com/fisker/css-units-list", License: "MIT", Version: "2.1.0",
		Carries: "The CSS unit list captured into internal/format/css/css_units.go.",
		Reach:   Compiled, LicenseFile: "css-units-list.txt",
	},
	{
		Name: "GraphQL.js", URL: "https://github.com/graphql/graphql-js", License: "MIT", Version: "17.0.2",
		Carries: "The GraphQL lexer and parser in internal/format/graphql, ported.",
		Reach:   Compiled, LicenseFile: "graphql-js.txt",
	},
	{
		Name: "typescript-estree", URL: "https://github.com/typescript-eslint/typescript-eslint/tree/main/packages/typescript-estree", License: "MIT", Version: "8.65.0",
		Carries: "The TypeScript-to-ESTree conversion in internal/format/estree, ported, and its XHTML entity table, captured.",
		Reach:   Compiled, LicenseFile: "typescript-estree.txt",
	},
	{
		Name: "@typescript-eslint/visitor-keys", URL: "https://github.com/typescript-eslint/typescript-eslint/tree/main/packages/visitor-keys", License: "MIT", Version: "8.65.0",
		Carries: "Visitor keys captured, through Prettier's table, into internal/format/estree/visitor_keys_generated.go.",
		Reach:   Compiled, LicenseFile: "typescript-eslint-visitor-keys.txt",
	},
	{
		Name: "@babel/types", URL: "https://github.com/babel/babel/tree/main/packages/babel-types", License: "MIT", Version: "8.0.0",
		Carries: "Visitor keys captured, through Prettier's table, into internal/format/estree/visitor_keys_generated.go.",
		Reach:   Compiled, LicenseFile: "babel-types.txt",
	},
	{
		Name: "angular-estree-parser", URL: "https://github.com/prettier/angular-estree-parser", License: "MIT", Version: "15.5.0",
		Carries: "Visitor keys captured, through Prettier's table, into internal/format/estree/visitor_keys_generated.go.",
		Reach:   Compiled, LicenseFile: "angular-estree-parser.txt",
	},
	{
		Name: "Flow (flow-parser)", URL: "https://github.com/facebook/flow", License: "MIT", Version: "0.322.0",
		Carries: "Visitor keys captured, through Prettier's table, into internal/format/estree/visitor_keys_generated.go. flow-parser ships no license file, so this is Flow's, at its v0.322.0 tag.",
		Reach:   Compiled, LicenseFile: "flow.txt",
	},
	{
		Name: "micromark-extension-gfm-autolink-literal", URL: "https://github.com/micromark/micromark-extension-gfm-autolink-literal", License: "MIT", Version: "2.1.0",
		Carries: "Ported in internal/format/markdown/micromark.",
		Reach:   Compiled, LicenseFile: "micromark-extension-gfm-autolink-literal.txt",
	},
	{
		Name: "micromark-extension-gfm-footnote", URL: "https://github.com/micromark/micromark-extension-gfm-footnote", License: "MIT", Version: "2.1.0",
		Carries: "Ported in internal/format/markdown/micromark.",
		Reach:   Compiled, LicenseFile: "micromark-extension-gfm-footnote.txt",
	},
	{
		Name: "micromark-extension-gfm-strikethrough", URL: "https://github.com/micromark/micromark-extension-gfm-strikethrough", License: "MIT", Version: "2.1.0",
		Carries: "Ported in internal/format/markdown/micromark.",
		Reach:   Compiled, LicenseFile: "micromark-extension-gfm-strikethrough.txt",
	},
	{
		Name: "micromark-extension-gfm-table", URL: "https://github.com/micromark/micromark-extension-gfm-table", License: "MIT", Version: "2.1.1",
		Carries: "Ported in internal/format/markdown/micromark.",
		Reach:   Compiled, LicenseFile: "micromark-extension-gfm-table.txt",
	},
	{
		Name: "micromark-extension-gfm-task-list-item", URL: "https://github.com/micromark/micromark-extension-gfm-task-list-item", License: "MIT", Version: "2.1.0",
		Carries: "Ported in internal/format/markdown/micromark.",
		Reach:   Compiled, LicenseFile: "micromark-extension-gfm-task-list-item.txt",
	},
	{
		Name: "micromark-extension-math", URL: "https://github.com/micromark/micromark-extension-math", License: "MIT", Version: "3.1.0",
		Carries: "Ported in internal/format/markdown/micromark.",
		Reach:   Compiled, LicenseFile: "micromark-extension-math.txt",
	},
	{
		Name: "@braindb/micromark-extension-wiki-link", URL: "https://github.com/stereobooster/braindb", License: "MIT", Version: "0.1.0",
		Carries: "Ported in internal/format/markdown/micromark/wiki_link.go.",
		Reach:   Compiled, Note: "No license text is published with this package or in its repository. Its package.json declares MIT. Contributors: Mark Hudnall and stereobooster.",
	},
	{
		Name: "mdast-util-gfm-autolink-literal", URL: "https://github.com/syntax-tree/mdast-util-gfm-autolink-literal", License: "MIT", Version: "2.0.1",
		Carries: "Ported in internal/format/markdown/mdast.",
		Reach:   Compiled, LicenseFile: "mdast-util-gfm-autolink-literal.txt",
	},
	{
		Name: "mdast-util-gfm-footnote", URL: "https://github.com/syntax-tree/mdast-util-gfm-footnote", License: "MIT", Version: "2.1.0",
		Carries: "Ported in internal/format/markdown/mdast.",
		Reach:   Compiled, LicenseFile: "mdast-util-gfm-footnote.txt",
	},
	{
		Name: "mdast-util-gfm-strikethrough", URL: "https://github.com/syntax-tree/mdast-util-gfm-strikethrough", License: "MIT", Version: "2.0.0",
		Carries: "Ported in internal/format/markdown/mdast.",
		Reach:   Compiled, LicenseFile: "mdast-util-gfm-strikethrough.txt",
	},
	{
		Name: "mdast-util-gfm-table", URL: "https://github.com/syntax-tree/mdast-util-gfm-table", License: "MIT", Version: "2.0.0",
		Carries: "Ported in internal/format/markdown/mdast.",
		Reach:   Compiled, LicenseFile: "mdast-util-gfm-table.txt",
	},
	{
		Name: "mdast-util-gfm-task-list-item", URL: "https://github.com/syntax-tree/mdast-util-gfm-task-list-item", License: "MIT", Version: "2.0.0",
		Carries: "Ported in internal/format/markdown/mdast.",
		Reach:   Compiled, LicenseFile: "mdast-util-gfm-task-list-item.txt",
	},
	{
		Name: "mdast-util-math", URL: "https://github.com/syntax-tree/mdast-util-math", License: "MIT", Version: "3.0.0",
		Carries: "Ported in internal/format/markdown/mdast.",
		Reach:   Compiled, LicenseFile: "mdast-util-math.txt",
	},
	{
		Name: "mdast-util-to-string", URL: "https://github.com/syntax-tree/mdast-util-to-string", License: "MIT", Version: "4.0.0",
		Carries: "Ported in internal/format/markdown/mdast.",
		Reach:   Compiled, LicenseFile: "mdast-util-to-string.txt",
	},
	{
		Name: "@braindb/mdast-util-wiki-link", URL: "https://github.com/stereobooster/braindb", License: "MIT", Version: "0.2.0",
		Carries: "Ported in internal/format/markdown/mdast.",
		Reach:   Compiled, Note: "No license text is published with this package or in its repository. Its package.json declares MIT. Contributors: Mark Hudnall and stereobooster.",
	},
	{
		Name: "decode-named-character-reference", URL: "https://github.com/wooorm/decode-named-character-reference", License: "MIT", Version: "1.3.0",
		Carries: "Ported in internal/format/markdown/micromark/decode.go.",
		Reach:   Compiled, LicenseFile: "decode-named-character-reference.txt",
	},
	{
		Name: "character-entities", URL: "https://github.com/wooorm/character-entities", License: "MIT", Version: "2.0.2",
		Carries: "The named character references captured into internal/format/markdown/micromark/entities_generated.go.",
		Reach:   Compiled, LicenseFile: "character-entities.txt",
	},
	{
		Name: "linguist-languages", URL: "https://github.com/ikatyang/linguist-languages", License: "MIT", Version: "9.4.0",
		Carries: "GitHub Linguist's language list, captured through Prettier into internal/format/markdown/languages_generated.go.",
		Reach:   Compiled, LicenseFile: "linguist-languages.txt",
	},
	{
		Name: "cjk-regex", URL: "https://github.com/ikatyang/cjk-regex", License: "MIT", Version: "3.4.0",
		Carries: "Character classes captured through Prettier into internal/format/markdown/classes_generated.go.",
		Reach:   Compiled, LicenseFile: "cjk-regex.txt",
	},
	{
		Name: "unicode-regex", URL: "https://github.com/ikatyang/unicode-regex", License: "MIT", Version: "4.2.0",
		Carries: "Character classes captured through Prettier into internal/format/markdown/classes_generated.go.",
		Reach:   Compiled, LicenseFile: "unicode-regex.txt",
	},
	{
		Name: "Unicode Character Database", URL: "https://www.unicode.org/ucd/", License: "Unicode-3.0",
		Carries: "Character properties behind the generated tables in internal/format/doc and internal/format/markdown.",
		Reach:   Compiled, LicenseFile: "unicode.txt",
	},
	{
		Name: "ESLint", URL: "https://github.com/eslint/eslint", License: "MIT",
		Carries: "Core lint rules in internal/lint/rules/core, ported.",
		Reach:   Compiled, LicenseFile: "eslint.txt",
	},
	{
		Name: "eslint-utils", URL: "https://github.com/eslint-community/eslint-utils", License: "MIT",
		Carries: "The reference tracker in internal/lint/ecmascript/reference, ported.",
		Reach:   Compiled, LicenseFile: "eslint-utils.txt",
	},
	{
		Name: "typescript-eslint", URL: "https://github.com/typescript-eslint/typescript-eslint", License: "MIT",
		Carries: "TypeScript lint rules in internal/lint/rules/typescript, ported.",
		Reach:   Compiled, LicenseFile: "typescript-eslint.txt",
	},
	{
		Name: "oxc", URL: "https://github.com/oxc-project/oxc", License: "MIT",
		Carries: "Lint rules ported from oxc's, across internal/lint/rules.",
		Reach:   Compiled, LicenseFile: "oxc.txt",
	},
	{
		Name: "eslint-plugin-react", URL: "https://github.com/jsx-eslint/eslint-plugin-react", License: "MIT",
		Carries: "React lint rules in internal/lint/rules/react, ported.",
		Reach:   Compiled, LicenseFile: "eslint-plugin-react.txt",
	},
	{
		Name: "React", URL: "https://github.com/facebook/react", License: "MIT",
		Carries: "eslint-plugin-react-hooks and the React Compiler's rules, in internal/lint/rules/react and internal/lint/ecmascript/high_level_intermediate_representation, ported.",
		Reach:   Compiled, LicenseFile: "react.txt",
	},
	{
		Name: "Next.js", URL: "https://github.com/vercel/next.js", License: "MIT",
		Carries: "@next/eslint-plugin-next's rules in internal/lint/rules/next, ported.",
		Reach:   Compiled, LicenseFile: "next.js.txt",
	},
	{
		Name: "Tailwind CSS", URL: "https://github.com/tailwindlabs/tailwindcss", License: "MIT",
		Carries: "Tailwind's class model in internal/lint/rules/tailwind, ported.",
		Reach:   Compiled, LicenseFile: "tailwindcss.txt",
	},
	{
		Name: "prettier-plugin-tailwindcss", URL: "https://github.com/tailwindlabs/prettier-plugin-tailwindcss", License: "MIT",
		Carries: "Tailwind class ordering in internal/lint/rules/tailwind, ported.",
		Reach:   Compiled, LicenseFile: "prettier-plugin-tailwindcss.txt",
	},
	{
		Name: "eslint-plugin-better-tailwindcss", URL: "https://github.com/schoero/eslint-plugin-better-tailwindcss", License: "MIT",
		Carries: "Tailwind lint rules in internal/lint/rules/tailwind, ported.",
		Reach:   Compiled, LicenseFile: "eslint-plugin-better-tailwindcss.txt",
	},
	{
		Name: "ts-api-utils", URL: "https://github.com/JoshuaKGoldberg/ts-api-utils", License: "MIT",
		Carries: "Type predicates in internal/lint/checking, through tsgolint's port of them.",
		Reach:   Compiled, LicenseFile: "ts-api-utils.txt",
	},
	{
		Name: "tsutils", URL: "https://github.com/ajafff/tsutils", License: "MIT",
		Carries: "A function in internal/lint/rules/typescript/no_floating_promises.go, ported and modified.",
		Reach:   Compiled, LicenseFile: "tsutils.txt",
	},
	{
		Name: "TSLint", URL: "https://github.com/palantir/tslint", License: "Apache-2.0",
		Carries: "A regular expression in internal/lint/rules/typescript/ban_tslint_comment.go, copied.",
		Reach:   Compiled, LicenseFile: "tslint.txt",
	},
	{
		Name: "eslint-plugin-boundaries", URL: "https://github.com/javierbrea/eslint-plugin-boundaries", License: "MIT",
		Carries: "The boundary rules in internal/lint/rules/boundaries, ported.",
		Reach:   Compiled, LicenseFile: "eslint-plugin-boundaries.txt",
	},
	{
		Name: "eslint-plugin-eslint-comments", URL: "https://github.com/eslint-community/eslint-plugin-eslint-comments", License: "MIT",
		Carries: "A rule in internal/lint/rules/core/require_description.go, ported.",
		Reach:   Compiled, LicenseFile: "eslint-plugin-eslint-comments.txt",
	},
	{
		Name: "globals", URL: "https://github.com/sindresorhus/globals", License: "MIT", Version: "16.4.0",
		Carries: "Global names captured into internal/lint/rules/core/no_extend_native.go.",
		Reach:   Compiled, LicenseFile: "globals.txt",
	},
	{
		Name: "confusing-browser-globals", URL: "https://github.com/facebook/create-react-app/tree/main/packages/confusing-browser-globals", License: "MIT",
		Carries: "The restricted globals list embedded in internal/lint/configuration/sets/typescript.json.",
		Reach:   Compiled, LicenseFile: "confusing-browser-globals.txt",
	},
	{
		Name: "Prettier's bundled builds", URL: "https://github.com/system-inc/prettier", License: "MIT",
		Carries: "Built Prettier bundles vendored in internal/format/prettier/bundles, which tests compare the formatter against. They bundle many other projects, whose licenses are in internal/format/prettier/bundles/THIRD-PARTY-NOTICES.md.",
		Reach:   RepositoryOnly, LicenseFile: "prettier.txt",
	},
}

// moduleLicenses is each Go module's license, read from its license file, for every module in the build
// graph or go.mod. A module missing here stops the generator by name, so a new dependency is read
// before it is credited.
var moduleLicenses = map[string]string{
	"github.com/Microsoft/go-winio":     "MIT",
	"github.com/klauspost/cpuid/v2":     "MIT",
	"github.com/dlclark/regexp2/v2":     "MIT",
	"github.com/zeebo/xxh3":             "BSD-2-Clause",
	"golang.org/x/sync":                 "BSD-3-Clause",
	"golang.org/x/sys":                  "BSD-3-Clause",
	"golang.org/x/text":                 "BSD-3-Clause",
	"github.com/dop251/goja":            "MIT",
	"github.com/go-sourcemap/sourcemap": "BSD-2-Clause",
	"github.com/google/pprof":           "Apache-2.0",
	"golang.org/x/mod":                  "BSD-3-Clause",
	"golang.org/x/tools":                "BSD-3-Clause",
}

// swiftLicense is a Swift package's license expression and its text under licenses/.
type swiftLicense struct {
	Expression string
	File       string
}

// swiftLicenses are the Swift engine's packages, by their identity in swift/Package.resolved, each
// license taken at the pinned revision.
var swiftLicenses = map[string]swiftLicense{
	"swift-argument-parser": {Expression: "Apache-2.0 WITH Swift-exception", File: "swift-argument-parser.txt"},
	"swift-cmark":           {Expression: "BSD-2-Clause AND MIT", File: "swift-cmark.txt"},
	"swift-format":          {Expression: "Apache-2.0 WITH Swift-exception", File: "swift-format.txt"},
	"swift-markdown":        {Expression: "Apache-2.0 WITH Swift-exception", File: "swift-markdown.txt"},
	"swift-syntax":          {Expression: "Apache-2.0 WITH Swift-exception", File: "swift-syntax.txt"},
}
