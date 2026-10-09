package graphql

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/format/doc"
	"github.com/system-inc/cohere/internal/format/formatoptions"
	"github.com/system-inc/cohere/internal/format/prettier"
)

// Each snippet goes through this printer and through the embedded Prettier fork, and the two outputs must
// be equal. Snippets are written unformatted, so the oracle rewrites each one and an identity printer
// cannot pass; contains spells out what the case is about, so a printer and an oracle that both lost it
// cannot agree their way past it. Every case runs under every option set in formatOptionSets.

type formatCase struct {
	name     string
	source   string
	contains []string
}

var formatCases = []formatCase{
	{
		name: "operations: named, anonymous, shorthand, every operation type",
		source: "query   Named($id:ID!,$first:Int=10 @deprecated){user(id:$id){id}}\n" +
			"query ($a:Int){a}\nquery{b}\n{c}\nmutation   M{d}\nsubscription S @live{e}\n" +
			"mutation($b:[String!]!=[\"x\"]){f}\n",
		contains: []string{"query Named($id: ID!, $first: Int = 10 @deprecated) {", "query ($a: Int) {", "query {", "\n{\n", "subscription S @live {"},
	},
	{
		name: "descriptions on executable definitions and variables, block and plain",
		source: "\"Plain operation description\" query Q(\"variable description\" $a:Int){a}\n" +
			"\"\"\"\n  Block fragment description\n  second line\n\"\"\" fragment F on T{a}\n",
		contains: []string{"\"Plain operation description\"\nquery Q(", "\"\"\"\nBlock fragment description\nsecond line\n\"\"\"\nfragment F on T {"},
	},
	{
		name:     "fields: aliases, arguments, directives, nested selections, fragments",
		source:   "query{alias:field(a:1,b:\"two\",c:THREE,d:null,e:true,f:false,g:1.5e3,h:$v,i:[1,2],j:{k:1,l:{m:2}})@include(if:$x)@skip(if:false){...Spread @defer ...on Type{x} ... @include(if:true){y} ...{z}}}\n",
		contains: []string{"alias: field(", "@include(if: $x) @skip(if: false) {", "...Spread @defer", "... on Type {", "... @include(if: true) {", "... {"},
	},
	{
		name:     "strings: escapes, unicode, and every block string shape",
		source:   "query{a(s:\"quote\\\" backslash\\\\ newline\\n tab\\t unicode\\u00e9 slash\\/\",t:\"\"\"   one line   \"\"\",u:\"\"\"\n\n\"\"\",v:\"\"\"\n    first\n      indented\n    has \\\"\"\" triple\n\"\"\",w:\"\")}\n",
		contains: []string{`\"`, `\\`, `\n`, "one line\n", "u: \"\"\"\n", `\"""`},
	},
	{
		name:     "values: empty list and object, nested, with bracket spacing",
		source:   "query{a(l:[],o:{},n:{a:[{b:1}],c:{}},m:[[1],[]])}\n",
		contains: []string{"l: []", "o: {}"},
	},
	{
		name: "schema: schema, scalar, type, input, interface, enum, union, directive, and every extension",
		source: "\"schema description\" schema @a{query:Query mutation:Mutation subscription:Subscription}\n" +
			"extend schema @b\nextend schema{query:Q}\nscalar   Date @specifiedBy(url:\"https://example.com\")\nextend scalar Date @c\n" +
			"type Query implements A&B @key(fields:\"id\"){\"field description\" field(\"argument description\" a:Int=1 @d,b:[String!]!):String! @deprecated(reason:\"no\") other:Int}\n" +
			"extend type Query implements C{added:Int}\nextend type Query @e\ntype Empty\n" +
			"input In{a:Int=1 b:In2! @f}\nextend input In{c:String}\ninterface I implements J{a:Int}\nextend interface I @g{b:Int}\n" +
			"enum E{\"value description\" A @h B}\nextend enum E{C}\nenum Bare\nunion U=|A|B\nextend union U=C\nunion Lone\n" +
			"\"\"\"directive description\"\"\" directive @dir(a:Int=1,b:String) repeatable on FIELD|FRAGMENT_SPREAD|INLINE_FRAGMENT\n" +
			"directive @plain on QUERY\nextend directive @dir @i\n",
		contains: []string{
			"schema @a {", "extend schema @b\n", "type Query implements A & B @key(", "\"argument description\" a: Int = 1 @d",
			"type Empty\n", "union U = A | B", "repeatable on FIELD | FRAGMENT_SPREAD | INLINE_FRAGMENT",
			"extend directive @dir @i",
		},
	},
	{
		name:     "input value with a block description breaks the argument list",
		source:   "type T{f(\"\"\"block\"\"\" a:Int,b:Int):Int}\n",
		contains: []string{"f(\n"},
	},
	{
		name:     "fragment arguments on definitions and spreads",
		source:   "fragment F($a:Int=1,$b:String @deprecated) on T @dir{f(a:$a)}\nquery{...F(a:2,b:\"x\") @skip(if:false)}\n",
		contains: []string{"fragment F($a: Int = 1, $b: String @deprecated) on T @dir {", "...F(a: 2, b: \"x\") @skip(if: false)"},
	},
	{
		name:     "blank lines between definitions and members are kept, runs collapse to one",
		source:   "type A{a:Int\n\n\n\nb:Int\nc:Int}\n\n\n\ntype B{a:Int}\nquery{a\n\nb}\nenum E{A\n\nB}\n",
		contains: []string{"a: Int\n\n", "b: Int\n", "}\n\ntype B"},
	},
	{
		name: "comments: own line, end of line, between arguments, trailing spaces, end of file",
		source: "# leading the file   \n# ends in JavaScript-only whitespace\u00a0\f\ufeff\n\n# before a definition\ntype T { # after the brace\n  # own line before a field\n  a: Int # end of line\n" +
			"  b(\n  # before an argument\n  x: Int # after an argument\n  ): Int\n}\nquery { a # after a selection\n  # between selections\n  b }\n# end of file\n",
		contains: []string{"# leading the file\n", "# ends in JavaScript-only whitespace\n", "# end of line", "# end of file"},
	},
	{
		name: "comments: in variable definitions, object fields, list values, before a closing brace",
		source: "query Q(\n# before a variable\n$a:Int # after a variable\n$b:Int) @dir # after a directive\n{a(o:{\n# before a field\nx:1 # after a field\ny:[1, # in a list\n2]})\n" +
			"# before the closing brace\n}\nenum E { A # after a value\n# own line before the end\n}\nunion U = # after the equals\nA | B\n",
		contains: []string{"# before a variable", "# after a field", "# before the closing brace", "# after the equals"},
	},
	{
		name:     "comments dangling in an empty list and an empty object",
		source:   "query{a(l:[ # dangling in a list\n],o:{ # dangling in an object\n})}\n",
		contains: []string{"# dangling in a list", "# dangling in an object"},
	},
	{
		name: "long argument and variable lists break at the print width",
		source: "query LongQuery($firstVariable:String,$secondVariable:Int,$thirdVariable:Boolean,$fourthVariable:[String!]!,$fifthVariable:ID){" +
			"someField(firstArgument:$firstVariable,secondArgument:$secondVariable,thirdArgument:$thirdVariable,fourth:$fourthVariable){id}" +
			"short(a:1)}\ntype T{field(firstArgument:String,secondArgument:Int,thirdArgument:Boolean,fourthArgument:Float,fifth:ID):String}\n" +
			"directive @long(firstArgument:String,secondArgument:Int,thirdArgument:Boolean,fourthArgument:Float,fifth:ID) on FIELD\n" +
			"query{f(o:{firstKey:\"a long value here\",secondKey:\"another long value\",thirdKey:\"and a third long value to break\"},l:[\"one long string value\",\"two long string value\",\"three long string value\"])}\n" +
			"query{f @firstDirective(argument:\"value\") @secondDirective(argument:\"value\") @thirdDirective(argument:\"value\") @fourth}\n" +
			"query Q @firstDirective(argument:\"value\") @secondDirective(argument:\"value\") @thirdDirective(argument:\"value\") @fourth{a}\n",
		contains: []string{"query LongQuery(\n"},
	},
	{
		name: "interfaces with ampersands and unions with pipes that break",
		source: "type Implementing implements FirstInterface&SecondInterface&ThirdInterface&FourthInterface&FifthInterface&SixthInterface{a:Int}\n" +
			"union LongUnion=FirstMemberType|SecondMemberType|ThirdMemberType|FourthMemberType|FifthMemberType|SixthMemberType|Seventh\n" +
			"\"described\" union DescribedUnion @dir=FirstMemberType|SecondMemberType|ThirdMemberType|FourthMemberType|FifthMemberType|Sixth\n",
		contains: []string{"FirstInterface &", "| FirstMemberType"},
	},
	{
		name:     "prettier-ignore keeps a definition and a field as written",
		source:   "# prettier-ignore\ntype   Kept{a:Int}\ntype Formatted{\n  # prettier-ignore\n  kept  (a:Int)  :  Int\n  b:Int}\nquery{a   b}\n# prettier-ignore\u00a0\ufeff\nscalar   AlsoKept\n",
		contains: []string{"type   Kept{a:Int}", "kept  (a:Int)  :  Int", "scalar   AlsoKept"},
	},
}

// formatOptionSets are the options every case runs under: ahra's defaults (printWidth 120, tabWidth 4),
// printWidth 80 with Prettier's tabWidth 2, bracketSpacing false, and tabs.
func formatOptionSets() map[string]formatoptions.Options {
	narrow := formatoptions.Default()
	narrow.PrintWidth = 80
	narrow.TabWidth = 2
	noBracketSpacing := formatoptions.Default()
	noBracketSpacing.BracketSpacing = false
	tabs := formatoptions.Default()
	tabs.UseTabs = true
	return map[string]formatoptions.Options{
		"defaults":             formatoptions.Default(),
		"printWidth 80":        narrow,
		"bracketSpacing false": noBracketSpacing,
		"useTabs":              tabs,
	}
}

func TestFormatMatchesTheFork(t *testing.T) {
	t.Parallel()
	for optionsName, options := range formatOptionSets() {
		oracle, err := prettier.New(options)
		if err != nil {
			t.Fatal(err)
		}
		for _, testCase := range formatCases {
			expected, err := oracle.Format("Probe.graphql", testCase.source)
			if err != nil {
				t.Fatalf("%s (%s): the oracle failed: %v", testCase.name, optionsName, err)
			}
			if expected == testCase.source {
				t.Fatalf("%s (%s): the oracle left the snippet unchanged, so an identity printer would pass", testCase.name, optionsName)
			}
			for _, substring := range testCase.contains {
				if !strings.Contains(expected, substring) {
					t.Errorf("%s (%s): the fork's output lacks %q:\n%s", testCase.name, optionsName, substring, expected)
				}
			}
			actual, err := Format(testCase.source, options)
			if err != nil {
				t.Errorf("%s (%s): %v", testCase.name, optionsName, err)
				continue
			}
			if actual != expected {
				t.Errorf("%s (%s):\n--- fork\n%s--- native\n%s", testCase.name, optionsName, expected, actual)
			}
		}
	}
}

// The option sets have to change the output somewhere, or running under them proves nothing.
func TestOptionSetsChangeTheOutput(t *testing.T) {
	t.Parallel()
	defaults, err := Format(formatCases[2].source, formatoptions.Default())
	if err != nil {
		t.Fatal(err)
	}
	for optionsName, options := range formatOptionSets() {
		if optionsName == "defaults" {
			continue
		}
		changed := false
		for _, testCase := range formatCases {
			withDefaults, defaultsErr := Format(testCase.source, formatoptions.Default())
			withOptions, optionsErr := Format(testCase.source, options)
			if defaultsErr == nil && optionsErr == nil && withDefaults != withOptions {
				changed = true
				break
			}
		}
		if !changed {
			t.Errorf("%s printed every case as the defaults do", optionsName)
		}
	}
	if !strings.Contains(defaults, "{ k: 1, l: { m: 2 } }") {
		t.Errorf("bracketSpacing true lost its spaces:\n%s", defaults)
	}
}

// What graphql-js refuses, both refuse: an empty selection set, an empty argument list, and a description
// on a shorthand query.
func TestFormatRefusesWhatTheForkRefuses(t *testing.T) {
	t.Parallel()
	oracle, err := prettier.New(formatoptions.Default())
	if err != nil {
		t.Fatal(err)
	}
	for _, source := range []string{"query { }\n", "query { a() }\n", "\"description\" { a }\n", "type T { a: }\n"} {
		if _, err := oracle.Format("Probe.graphql", source); err == nil {
			t.Fatalf("the oracle accepted %q", source)
		}
		if formatted, err := Format(source, formatoptions.Default()); err == nil {
			t.Errorf("%q was printed instead of refused:\n%s", source, formatted)
		}
	}
}

// PrintToDoc is the same doc without its trailing hardline, which is what the TypeScript embed lays out.
func TestPrintToDocIsFormatWithoutTheTrailingHardline(t *testing.T) {
	t.Parallel()
	options := formatoptions.Default()
	for _, testCase := range formatCases {
		formatted, err := Format(testCase.source, options)
		if err != nil {
			t.Fatalf("%s: %v", testCase.name, err)
		}
		document, err := PrintToDoc(testCase.source, options)
		if err != nil {
			t.Fatalf("%s: %v", testCase.name, err)
		}
		printed := doc.Print(document, doc.Options{PrintWidth: options.PrintWidth, TabWidth: options.TabWidth, UseTabs: options.UseTabs})
		if printed != strings.TrimSuffix(formatted, "\n") {
			t.Errorf("%s:\n--- Format\n%s--- PrintToDoc\n%s", testCase.name, formatted, printed)
		}
	}
	if _, err := PrintToDoc("query { }", options); err == nil {
		t.Error("PrintToDoc printed an empty selection set instead of refusing it")
	}
}
