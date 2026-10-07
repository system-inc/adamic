package main

// The names are reviewed source boundaries, not statuses inferred from directory
// names. GAPS.md and the production/test files must exist in the measured tree.
// Partial functions are deliberately not credited. New slices fail closed until
// their upstream mapping is reviewed. Full files include comments and blank lines;
// declaration spans include their attached documentation. See the methodology.
var slices = []slice{
	{
		Directory: "gitignore", Required: []string{"main.ts", "gitignore_test.go"},
		Coverage: files("internal/gitignore", "gitignore.go", "glob.go"),
		Note:     "Matcher and glob implementation; recorded native/Node/Go corpus, UTF-8 scope.",
	},
	{
		Directory: "mediaquery", Required: []string{"main.ts", "mediaquery_test.go"},
		Coverage: files("internal/format/css/mediaquery", "index.go", "nodes.go", "parsers.go", "whitespace.go"),
		Note:     "Media-query parser; recorded canonical tree parity, not full CSS formatting.",
	},
	{
		Directory: "values", Required: []string{"main.ts", "values_test.go"},
		Coverage: files("internal/format/css/values", "values.go", "nodes.go", "tokenize.go", "parser.go"),
		Note:     "Value parser in both supported modes; recorded canonical tree parity.",
	},
	{
		Directory: "graphql", Required: []string{"main.ts", "graphql_test.go"},
		Coverage: files("internal/format/graphql", "lexer.go", "parser.go", "token.go", "block_string.go", "character_classes.go"),
		Partial:  []string{"internal/format/graphql"},
		Note:     "Lexer/parser only; format.go and GraphQL printers remain unported.",
	},
	{
		Directory: "suppression", Required: []string{"main.ts", "suppression_test.go"},
		Coverage: append(files("internal/lint/ecmascript/directives", "directives.go", "subject.go"), files("internal/lint/suppression", "suppression.go", "scan.go", "parse.go", "directive_subject.go")...),
		Packages: []string{"internal/lint/suppression"}, Partial: []string{"internal/lint/suppression"},
		Note: "Directive parsing and suppression index; nil-Index and concurrency scope limits remain in GAPS.md.",
	},
	{
		Directory: "selector", Required: []string{"main.ts", "selector_test.go"},
		Coverage: files("internal/format/css/selector", "nodes.go", "tokenize.go"),
		Packages: []string{"internal/format/css/selector"}, Partial: []string{"internal/format/css/selector"},
		Note: "Default lossless selector parsing; parser/options and rendering boundary not credited as complete files.",
	},
	{
		Directory: "formatfiles", Required: []string{"main.ts", "formatfiles_test.go"},
		Coverage: []coverage{{"internal/format/formatfiles/enumerate.go", []string{"repositoryBoundaryBetween", "NestedRepositoriesBelow", "HasOwnRepository", "NestedRepositoryContaining", "parentOf"}}},
		Packages: []string{"internal/format/formatfiles"}, Partial: []string{"internal/format/formatfiles"},
		Note: "Repository-boundary helpers credited; Enumerate requires supplied settings and excludes configured glob cases.",
	},
	{
		Directory: "cssstrings", Required: []string{"strings.ts", "port_test.go"},
		Coverage: []coverage{{"internal/format/css/print_misc.go", []string{"matchStringAt", "adjustStrings", "getPreferredQuote", "makeString", "printString"}}},
		Partial:  []string{"internal/format/css"}, Note: "CSS quote sub-printer only; shared source lines are unioned with the numeric slice.",
	},
	{
		Directory: "cssnumbers", Required: []string{"numbers.ts", "port_test.go"},
		Coverage: append(files("internal/format/css", "css_units.go"), coverage{"internal/format/css/print_misc.go", []string{"printUnit", "isDigit", "isLetter", "skipDigits", "matchNumberAt", "isWordPartStart", "isWordPartCharacter", "matchNumberWithUnitAt", "adjustNumbers", "printCssNumber"}}),
		Partial:  []string{"internal/format/css"}, Note: "Numeric adjustment and units; arbitrary printNumber inputs and full CSS layout excluded.",
	},
	{
		Directory: "json", Required: []string{"main.ts", "port_test.go"},
		Coverage: append(files("internal/format/estree", "parse_json.go"), files("internal/format/javascript", "print_json.go")...),
		Packages: []string{"internal/format/doc"}, Partial: []string{"internal/format/estree", "internal/format/javascript", "internal/format/doc"},
		Note: "Default-options JSON parser/stringify; shared JS/doc functions are partial and conservatively receive no whole-function credit.",
	},
	{
		Directory: "lint", Required: []string{"main.ts", "lint_test.go"},
		Coverage: files("internal/lint/rules/core", "no_debugger.go", "no_empty.go", "eqeqeq.go", "no_var.go", "no_duplicate_case.go"),
		Partial:  []string{"internal/lint/rules/core"}, Note: "Thirty implemented syntax rules, including one with partial recovery coverage; whole-file credit remains the five reviewed baseline rules. General configuration/fix integration excluded.",
	},
	{
		Directory: "typeaware", Required: []string{"main.ts", "suite.ts", "volume_suite.ts", "typeaware_test.go", "suite_test.go", "volume_test.go"},
		Packages: []string{"internal/lint/rules/typescript"}, Partial: []string{"internal/lint/rules/typescript"},
		Note: "Sixteen default-option rules using the external Go checker; configuration, suppression and edit-engine integration excluded. Mixed upstream rule files conservatively receive no line credit.",
	},
	{
		Directory: "css", Required: []string{"input.ts", "parser.ts", "css_test.go"},
		Coverage: files("internal/format/css/postcss", "input.go", "tokenize.go", "scss_tokenize.go", "parser.go", "scss_parser.go", "parse.go", "scss_parse.go", "node.go", "nested_declaration.go"),
		Note:     "Raw CSS/SCSS parser; raw native parity is independent of the composition gap.",
	},
	{
		Directory: "css", Required: []string{"compose.ts", "compose_main.ts"}, NativePhrase: "The former `ir.RegExpCall` cycle-proof refusal is closed",
		Coverage: files("internal/format/css", "parser.go", "location.go", "parse_media_query.go", "parse_selector.go", "parse_value.go", "parse_utilities.go"),
		Packages: []string{"internal/format/css"}, Partial: []string{"internal/format/css"},
		Note: "Composed parser credited only where this tree records native gap closure; Node-only composition receives no native credit.",
	},
	{
		Directory: "css", Required: []string{"print.ts", "print_test.go"}, NativePhrase: "The former `ir.RegExpNew` cycle-proof refusal is closed",
		Coverage: files("internal/format/css", "print.go", "print_helpers.go", "print_comma_separated_value_group.go", "print_parenthesized_value_group.go", "print_sequence.go"),
		Packages: []string{"internal/format/css"}, Partial: []string{"internal/format/css"},
		Note: "CSS/SCSS printer native parity with documented entry-point boundaries; remaining utilities/wrappers conservatively excluded.",
	},
	{
		Directory: "markdowninline", Required: []string{"inline.ts", "port_test.go"},
		Coverage: []coverage{{"internal/format/markdown/word.go", []string{"escapeDelimiterRuns", "allBackslashes", "isLineTerminatorUnit", "matchDelimiterRun", "canOpenOrCloseStrongOrEmphasis"}}, {"internal/format/markdown/print.go", []string{"encodeURL", "printURL", "printTitle", "getPreferredQuote", "getMaxContinuousCount"}}},
		Packages: []string{"internal/format/markdown"}, Partial: []string{"internal/format/markdown"},
		Note: "Inline leaf primitives; AstPath/context supplied by fixtures, no Markdown-file formatter.",
	},
	{
		Directory: "markdownblocks", Required: []string{"preprocess.ts"},
		Coverage: []coverage{{"internal/format/markdown/micromark/parse.go", []string{"preprocess"}}},
		Partial:  []string{"internal/format/markdown/micromark"}, Note: "Source preprocessing only; complete tokenizer/parser remains unfinished.",
	},
	{
		Directory: "markdownblocks", Required: []string{"frontmatter.ts"},
		Coverage: []coverage{{"internal/format/markdown/mdast/parse_markdown.go", []string{"ParseFrontMatter", "getFrontMatter"}}},
		Partial:  []string{"internal/format/markdown/mdast"}, Note: "Front-matter parser stage; no native complete mdast construction.",
	},
	{
		Directory: "markdownblocks", Required: []string{"lists.ts", "quotes.ts", "tables.ts", "codeblocks.ts", "htmlblocks.ts"},
		Coverage: files("internal/format/markdown", "list.go"),
		Packages: []string{"internal/format/markdown", "internal/format/doc"}, Partial: []string{"internal/format/markdown", "internal/format/doc"},
		Note: "Native layout on Go trees/widths; lists credited, mixed table/print/doc declarations conservatively excluded; full parser/formatter unfinished.",
	},
	{
		Directory: "yaml", Required: []string{"main.ts"},
		Note: "Baseline audit only: GAPS.md records no Adamic parser, printer or driver. YAML packages remain not started.",
	},
}

func files(directory string, names ...string) []coverage {
	var result []coverage
	for _, name := range names {
		result = append(result, coverage{File: directory + "/" + name})
	}
	return result
}
