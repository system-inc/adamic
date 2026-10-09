package tailwind

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

// settingsSamples is two values for every key a better-tailwindcss rule declares: one written in
// settings["better-tailwindcss"], and one in the rule's own options that must win over it.
//
// TestEveryDeclaredKeyHasASettingsCase fails on a declared key with no entry here, so an option that
// lands later gets its settings case in the same commit.
var settingsSamples = map[string][2]string{
	"entryPoint":             {`"./settings.css"`, `"./options.css"`},
	"tailwindConfig":         {`"./settings.css"`, `"./options.css"`},
	"cwd":                    {`"settings"`, `"options"`},
	"attributes":             {`["fromSettings"]`, `["fromOptions"]`},
	"callees":                {`["fromSettings"]`, `["fromOptions"]`},
	"variables":              {`["fromSettings"]`, `["fromOptions"]`},
	"ignore":                 {`["^settings$"]`, `["^options$"]`},
	"collapse":               {`false`, `true`},
	"logical":                {`false`, `true`},
	"order":                  {`"asc"`, `"desc"`},
	"unknownClassPosition":   {`"start"`, `"end"`},
	"unknownClassOrder":      {`"asc"`, `"desc"`},
	"componentClassPosition": {`"start"`, `"end"`},
	"componentClassOrder":    {`"asc"`, `"desc"`},
	"position":               {`"legacy"`, `"recommended"`},
	"syntax":                 {`"variable"`, `"shorthand"`},
	"allowMultiline":         {`false`, `true`},
}

// tailwindDecoders is every better-tailwindcss rule's registered decoder, by name.
func tailwindDecoders(t *testing.T) map[string]func(raw []byte, base rule.OptionsBase) (any, error) {
	t.Helper()
	decoders := map[string]func(raw []byte, base rule.OptionsBase) (any, error){}
	for _, registration := range rule.Registered() {
		if _, isTailwind := tailwindOptionKeys[registration.Rule.Name]; isTailwind {
			decoders[registration.Rule.Name] = registration.DecodeAt
		}
	}
	if len(decoders) != 12 {
		t.Fatalf("found %d better-tailwindcss decoders, want 12", len(decoders))
	}
	return decoders
}

// decodedKey re-encodes decoded options and returns one key's value, compacted.
func decodedKey(t *testing.T, decoded any, key string) string {
	t.Helper()
	encoded, err := json.Marshal(decoded)
	if err != nil {
		t.Fatalf("encoding %T: %v", decoded, err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &fields); err != nil {
		t.Fatalf("decoding %T's fields: %v", decoded, err)
	}
	return compactJsonForTest(t, fields[key])
}

func compactJsonForTest(t *testing.T, raw json.RawMessage) string {
	t.Helper()
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return string(raw)
	}
	compacted, _ := json.Marshal(value)
	return string(compacted)
}

func settingsBase(block string) rule.OptionsBase {
	return rule.OptionsBase{
		ConfigDirectory: "/repo",
		Settings:        map[string]json.RawMessage{"better-tailwindcss": json.RawMessage(block)},
	}
}

func TestEveryDeclaredKeyHasASettingsCase(t *testing.T) {
	t.Parallel()
	for ruleName, keys := range tailwindOptionKeys {
		for key := range keys {
			if _, sampled := settingsSamples[key]; !sampled {
				t.Errorf("%s declares %q and settingsSamples has no case for it", ruleName, key)
			}
		}
	}
}

// TestSettingsSitBetweenDefaultsAndOptions is a case per merged key, on every rule that declares it:
// a key written only in settings reaches the rule, and the rule's own options win over it, as
// upstream's createRule merges `{...defaults, ...settings, ...options}`.
func TestSettingsSitBetweenDefaultsAndOptions(t *testing.T) {
	t.Parallel()
	for ruleName, decode := range tailwindDecoders(t) {
		for key := range tailwindOptionKeys[ruleName] {
			sample := settingsSamples[key]
			base := settingsBase(`{"` + key + `": ` + sample[0] + `}`)

			fromSettings, err := decode(nil, base)
			if err != nil {
				t.Fatalf("%s, %s in settings: %v", ruleName, key, err)
			}
			if got := decodedKey(t, fromSettings, key); got != compactJsonForTest(t, json.RawMessage(sample[0])) {
				t.Errorf("%s: %s written in settings as %s decoded as %s", ruleName, key, sample[0], got)
			}

			fromBoth, err := decode([]byte(`{"`+key+`": `+sample[1]+`}`), base)
			if err != nil {
				t.Fatalf("%s, %s in settings and options: %v", ruleName, key, err)
			}
			if got := decodedKey(t, fromBoth, key); got != compactJsonForTest(t, json.RawMessage(sample[1])) {
				t.Errorf("%s: %s in options (%s) did not win over settings (%s): decoded %s",
					ruleName, key, sample[1], sample[0], got)
			}
		}
	}
}

// A key one rule declares and another does not reaches only the rule that declares it, and the merge
// is shallow: an options key replaces the settings key whole, never merging into it.
func TestSettingsReachOnlyTheRulesThatDeclareThem(t *testing.T) {
	t.Parallel()
	decoders := tailwindDecoders(t)
	base := settingsBase(`{"order": "desc", "callees": ["a", "b"]}`)

	decoded, err := decoders[NoDuplicateClasses.Name](nil, base)
	if err != nil {
		t.Fatalf("no-duplicate-classes read a key it does not declare instead of leaving it to its sibling: %v", err)
	}
	if callees := decoded.(NoDuplicateClassesOptions).Callees; strings.Join(callees, ",") != "a,b" {
		t.Fatalf("no-duplicate-classes' callees = %v, want settings' [a b]", callees)
	}

	ordered, err := decoders[EnforceConsistentClassOrder.Name]([]byte(`{"callees": ["c"]}`), base)
	if err != nil {
		t.Fatal(err)
	}
	options := ordered.(EnforceConsistentClassOrderOptions)
	if options.Order != "desc" || strings.Join(options.Callees, ",") != "c" {
		t.Fatalf("class order decoded order %q and callees %v, want settings' desc and options' [c] alone",
			options.Order, options.Callees)
	}
}

// The rule's own element is decoded strictly before anything merges, so its error names its own key.
func TestARulesOwnOptionsAreDecodedBeforeSettingsMerge(t *testing.T) {
	t.Parallel()
	decode := tailwindDecoders(t)[NoDuplicateClasses.Name]
	_, err := decode([]byte(`{"callee": ["cn"]}`), settingsBase(`{"callees": ["mergeClassNames"]}`))
	if err == nil || !strings.Contains(err.Error(), "callee") {
		t.Fatalf("a misspelled key in the rule's own options was not refused by name: %v", err)
	}
}

func TestSplitTailwindSettingsRefusesWhatNoRuleReads(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name, block, refused string
	}{
		{"an upstream option not ported yet", `{"selectors": []}`, `"selectors"`},
		{"another not ported yet", `{"tsconfig": "tsconfig.json"}`, `"tsconfig"`},
		{"no option at all", `{"callees": ["cn"], "calees": ["cn"]}`, `"calees"`},
		{"not an object", `["cn"]`, "expected an object"},
	}
	for _, testCase := range testCases {
		_, err := splitTailwindSettings([]byte(testCase.block))
		if err == nil || !strings.Contains(err.Error(), testCase.refused) {
			t.Errorf("%s: %s was not refused naming %s: %v", testCase.name, testCase.block, testCase.refused, err)
		}
	}

	parts, err := splitTailwindSettings([]byte(`{"entryPoint": "./theme.css", "order": "desc"}`))
	if err != nil {
		t.Fatal(err)
	}
	if len(parts) != 12 {
		t.Fatalf("entryPoint reached %d rules, want all 12", len(parts))
	}
	if !strings.Contains(string(parts[EnforceConsistentClassOrder.Name]), "order") ||
		strings.Contains(string(parts[NoDuplicateClasses.Name]), "order") {
		t.Fatalf("order went to the wrong rules: %v", parts)
	}
}

// ourSettings is the block our own repositories write in settings["better-tailwindcss"], verbatim.
const ourSettings = `{
	"entryPoint": "./app/_theme/styles/theme.css",
	"attributes": ["class", "className"],
	"callees": ["mergeClassNames", "createVariantClassNames"],
	"variables": [".*[Cc]lassName$", ".*[Cc]lassNames$"]
}`

// TestAnOutsiderIsReadWithUpstreamsDefaults is the zero-config case: a repository with no settings,
// writing shadcn's `cn` and `clsx`, has those strings read and reported, as upstream would. Before
// #gj5nm6e the defaults were ours, so every one of these went unchecked.
func TestAnOutsiderIsReadWithUpstreamsDefaults(t *testing.T) {
	t.Parallel()
	decoded, err := tailwindDecoders(t)[NoDuplicateClasses.Name](nil, rule.OptionsBase{ConfigDirectory: "/repo"})
	if err != nil {
		t.Fatal(err)
	}
	read := []string{
		`const merged = cn('flex flex');`,
		`const merged = clsx('p-2', open && 'p-2 p-2');`,
		`const merged = twMerge('gap-2 gap-2');`,
		`const merged = utils.cn('flex flex');`,
		`const Button = twc.button('px-2 px-2');`,
		`const classNames = 'flex flex';`,
		`const styles = 'flex flex';`,
		`const element = <div className="flex flex" />;`,
		`const element = <div class="flex flex" />;`,
		// An attribute's pattern and name are both lowercased before matching, as upstream's are.
		`const element = <div classname="flex flex" />;`,
	}
	for _, source := range read {
		rule_testing.ExpectFindings(t, rule_testing.RunWithOptions(t, NoDuplicateClasses, "Component.tsx", source, decoded), "duplicateClass")
	}
	unread := []string{
		`const merged = mergeClassNames('flex flex');`,
		`const buttonClassName = 'flex flex';`,
		`const merged = clb({ base: 'flex flex' });`,
	}
	for _, source := range unread {
		rule_testing.ExpectClean(t, rule_testing.RunWithOptions(t, NoDuplicateClasses, "Component.tsx", source, decoded))
	}
}

// TestOurSettingsReadOurMatchers is our consumers' case: their settings name our own callees and
// variables, which then replace upstream's for those kinds, while each kind they leave unwritten keeps
// its defaults.
func TestOurSettingsReadOurMatchers(t *testing.T) {
	t.Parallel()
	decoded, err := tailwindDecoders(t)[NoDuplicateClasses.Name](nil, settingsBase(ourSettings))
	if err != nil {
		t.Fatal(err)
	}
	read := []string{
		`const merged = mergeClassNames('flex flex');`,
		`const variants = createVariantClassNames('flex flex');`,
		`const buttonClassName = 'flex flex';`,
		`const rowClassNames = 'flex flex';`,
		`const element = <div className="flex flex" />;`,
	}
	for _, source := range read {
		rule_testing.ExpectFindings(t, rule_testing.RunWithOptions(t, NoDuplicateClasses, "Component.tsx", source, decoded), "duplicateClass")
	}
	unread := []string{
		`const merged = cn('flex flex');`,
		`const classes = 'flex flex';`,
	}
	for _, source := range unread {
		rule_testing.ExpectClean(t, rule_testing.RunWithOptions(t, NoDuplicateClasses, "Component.tsx", source, decoded))
	}
}

// An empty list replaces its kind's defaults with nothing, as upstream's `callees !== undefined` does.
func TestAnEmptyKindReadsNothingOfIt(t *testing.T) {
	t.Parallel()
	decoded, err := tailwindDecoders(t)[NoDuplicateClasses.Name](nil, settingsBase(`{"callees": []}`))
	if err != nil {
		t.Fatal(err)
	}
	rule_testing.ExpectClean(t, rule_testing.RunWithOptions(t, NoDuplicateClasses, "Component.tsx", `const merged = cn('flex flex');`, decoded))
	rule_testing.ExpectFindings(t, rule_testing.RunWithOptions(t, NoDuplicateClasses, "Component.tsx", `const classNames = 'flex flex';`, decoded), "duplicateClass")
}

// TestNamesMatchAsUpstreamsMatchesName pins utils.js's matchesName: the pattern's first match must
// be the whole name, and an attribute's pattern and name are lowercased first.
func TestNamesMatchAsUpstreamsMatchesName(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		patterns  []string
		lowercase bool
		name      string
		want      bool
	}{
		{[]string{"class"}, false, "class", true},
		{[]string{"class"}, false, "className", false},
		{[]string{".*[Cc]lassName$"}, false, "buttonClassName", true},
		{[]string{"a|ab"}, false, "ab", false}, // the first match is "a", which is not the name
		{[]string{"ab|a"}, false, "ab", true},
		{[]string{"^class(?:Name)?$"}, true, "ClassName", true},
		{[]string{"^class(?:Name)?$"}, false, "ClassName", false},
		{[]string{"(unclosed"}, false, "unclosed", false}, // skipped, never fatal
		{[]string{"x", "y"}, false, "y", true},
	}
	for _, testCase := range testCases {
		patterns := newNamePatterns(testCase.patterns, testCase.lowercase)
		for range 2 { // the second answer is the memo's
			if got := patterns.matches(testCase.name); got != testCase.want {
				t.Errorf("%v (lowercase %v) matching %q = %v, want %v",
					testCase.patterns, testCase.lowercase, testCase.name, got, testCase.want)
			}
		}
	}
}

// TestCalleeNameAndPathIsUpstreamsGetESCalleeName pins both readings of a callee.
func TestCalleeNameAndPathIsUpstreamsGetESCalleeName(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		callee, name, path string
	}{
		{"cn", "cn", "cn"},
		{"utils.cn", "cn", "utils.cn"},
		{"a.b.cn", "cn", "a.b.cn"},
		{"a['cn']", "cn", "a.cn"},
		{"this.cn", "cn", ""},
		{"items[0].cn", "cn", ""},
		{"a?.cn", "cn", "a.cn"},
		{"a[key]", "", ""},
	}
	for _, testCase := range testCases {
		source := "class C extends B { m() { " + testCase.callee + "('flex'); super.cn('flex'); } }"
		fileName := tspath.NormalizePath("/Component.tsx")
		sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: tspath.RootedFilePath(fileName), PathKey: tspath.PathKey(fileName)}, source, core.ScriptKindTSX)
		var calls []*ast.Node
		var walk func(node *ast.Node) bool
		walk = func(node *ast.Node) bool {
			if node.Kind == ast.KindCallExpression {
				calls = append(calls, node.AsCallExpression().Expression)
			}
			node.ForEachChild(walk)
			return false
		}
		walk(sourceFile.AsNode())
		if len(calls) != 2 {
			t.Fatalf("%s: found %d calls, want 2", testCase.callee, len(calls))
		}
		if name, path := calleeNameAndPath(calls[0]); name != testCase.name || path != testCase.path {
			t.Errorf("%s: name %q path %q, want %q and %q", testCase.callee, name, path, testCase.name, testCase.path)
		}
		if name := calleeName(calls[0]); name != testCase.name {
			t.Errorf("%s: calleeName %q, want %q, the name calleeNameAndPath gives", testCase.callee, name, testCase.name)
		}
		if name, path := calleeNameAndPath(calls[1]); name != "" || path != "" || calleeName(calls[1]) != "" {
			t.Errorf("super.cn: name %q path %q, want neither", name, path)
		}
	}
}
