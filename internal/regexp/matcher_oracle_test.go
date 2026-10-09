package regexp

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"
	"time"
	"unicode/utf16"
)

type executionCase struct {
	Pattern      string          `json:"pattern"`
	PatternUnits []uint16        `json:"patternUnits,omitempty"`
	Flags        string          `json:"flags"`
	Input        []uint16        `json:"input"`
	LastIndex    uint64          `json:"lastIndex"`
	Source       string          `json:"source"`
	Expected     executionResult `json:"expected"`
}
type executionResult struct {
	Captures  [][]int          `json:"captures"`
	LastIndex uint64           `json:"lastIndex"`
	Groups    map[string][]int `json:"groups"`
}

func matcherResult(r *RegExp, input []uint16) (executionResult, error) {
	m, err := r.Exec(input)
	out := executionResult{LastIndex: r.LastIndex}
	if m != nil {
		out.Captures = make([][]int, len(m.Captures))
		for i, c := range m.Captures {
			if c.Start >= 0 {
				out.Captures[i] = []int{c.Start, c.End}
			}
		}
		if len(m.Groups) > 0 {
			out.Groups = map[string][]int{}
			for name, c := range m.Groups {
				var indices []int
				if c.Start >= 0 {
					indices = []int{c.Start, c.End}
				}
				out.Groups[name] = indices
			}
		}
	}
	return out, err
}
func nodeResults(t *testing.T, cases []executionCase) []executionResult {
	t.Helper()
	data, err := json.Marshal(cases)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, "node", "-e", `
const cases=JSON.parse(require('fs').readFileSync(0,'utf8'));
if(!process.version.startsWith('v24.'))throw Error('Node 24 required');
const results=cases.map(c=>{
 const pattern=c.patternUnits?String.fromCharCode(...c.patternUnits):c.pattern;
 const r=new RegExp(pattern,c.flags.includes('d')?c.flags:c.flags+'d');
 r.lastIndex=c.lastIndex;const input=String.fromCharCode(...c.input);
 const m=r.exec(input);return {captures:m?Array.from(m.indices,x=>x??null):null,lastIndex:r.lastIndex,groups:m?.indices.groups?Object.fromEntries(Object.entries(m.indices.groups).map(([k,v])=>[k,v??null])):null};
});process.stdout.write(JSON.stringify(results));`)
	command.Stdin = bytes.NewReader(data)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("Node execution oracle: %v\n%s", err, output)
	}
	var results []executionResult
	if err = json.Unmarshal(output, &results); err != nil {
		t.Fatalf("Node decode: %v\n%s", err, output)
	}
	if len(results) != len(cases) {
		t.Fatal("Node result count")
	}
	return results
}
func compareExecutionCases(t *testing.T, cases []executionCase, external bool) {
	t.Helper()
	var results []executionResult
	if external {
		results = nodeResults(t, cases)
	}
	programs := map[string]*Program{}
	unavailablePrograms := map[string]error{}
	compared, unavailable := 0, 0
	for i, c := range cases {
		key := c.Pattern + "\x00" + c.Flags + fmt.Sprint(c.PatternUnits)
		if _, missing := unavailablePrograms[key]; missing {
			unavailable++
			continue
		}
		p := programs[key]
		if p == nil {
			var err error
			if c.PatternUnits != nil {
				p, err = CompileUTF16(c.PatternUnits, c.Flags)
			} else {
				p, err = Compile(c.Pattern, c.Flags)
			}
			if err != nil {
				var unavailableProperty *UnavailablePropertyError
				if !external && errors.As(err, &unavailableProperty) {
					unavailable++
					unavailablePrograms[key] = err
					t.Logf("UNAVAILABLE source=%s pattern=%q flags=%q error=%v", c.Source, c.Pattern, c.Flags, err)
					continue
				}
				t.Errorf("compile source=%s pattern=%q flags=%q: %v", c.Source, c.Pattern, c.Flags, err)
				continue
			}
			programs[key] = p
		}
		r := p.New()
		r.LastIndex = c.LastIndex
		r.StepLimit = 10_000_000
		got, err := matcherResult(r, c.Input)
		want := c.Expected
		if external {
			want = results[i]
		}
		if err != nil {
			t.Errorf("execution source=%s pattern=%q flags=%q input=%v lastIndex=%d: %v", c.Source, c.Pattern, c.Flags, c.Input, c.LastIndex, err)
			continue
		}
		compared++
		if !reflect.DeepEqual(got, want) {
			t.Errorf("DISAGREEMENT source=%s pattern=%q flags=%q input=%v lastIndex=%d got=%+v node=%+v", c.Source, c.Pattern, c.Flags, c.Input, c.LastIndex, got, want)
		}
	}
	t.Logf("execution totals: %d cases, %d compared, %d property stand-in unavailable", len(cases), compared, unavailable)
}
func TestMatcherNodeControls(t *testing.T) {
	t.Parallel()
	tests := [][3]string{
		{"[^[\\q{ab|a}]&&[a]]", "v", "abc"},
		{"\\u{D83C}\\u{DF0D}", "u", "🌍"}, {"\\uD83C\\u{DF0D}", "u", "🌍"}, {"\\u{D83C}\\uDF0D", "u", "🌍"},
		{"\\c", "", "\\c"}, {"\\c+", "", "\\ccc"}, {"[\\c_]", "", "\x1f"}, {"[\\c0]", "", "\x10"},
		{"\\8", "", "8"}, {"\\9", "", "9"}, {"\\400", "", " 0"}, {"\\777", "", "?7"}, {"\\1234", "", "S4"},
		{"\\uZ", "", "uZ"}, {"\\xZ", "", "xZ"}, {"\\u{3}", "", "uuu"}, {"\\k", "", "k"},
		{"[🌍]", "", "🌍"}, {"[a-🌍]", "", "🌍"}, {"[🌍-🌎]", "u", "🌎"}, {"[\\-]", "u", "-"}, {"[\\!]", "v", "!"}, {"[\\q{a\\-b}]", "v", "a-b"}, {"[[]", "", "["}, {"[\\d-a]+", "", "a-123"}, {"[\\n-\\r]+", "", "\n\r"},
		{"a+", "", "aaa"}, {"a+?", "", "aaa"}, {"(a|(b))+", "", "aba"},
		{"(?<=([ab]+)([bc]+))$", "", "abc"}, {"(?<=\\1(a))b", "", "aab"},
		{"(?=(a+))a*b\\1", "", "baabac"}, {"(?!((a)))b\\1", "", "b"},
		{"\\1(a)", "", "a"}, {"(a)?\\1", "", ""}, {"(a*)*", "", "aaa"},
		{"(a?){2,4}", "", "a"}, {"(a?){2,4}?", "", "a"}, {"(?=a)?a", "", "a"},
		{"[[a-z]&&[^aeiou]]+", "v", "aeibcd"}, {"[\\q{ab|a|}]b", "v", "ab"},
		{"[[\\q{ab|cd}]--[\\q{ab}]]", "v", "abcd"}, {"[[\\q{ab|cd}]&&[\\q{cd|ef}]]", "v", "abcd"},
		{"k", "i", "K"}, {"k", "iu", "K"}, {"s", "i", "ſ"}, {"s", "iu", "ſ"},
		{"[a-z]", "iu", "K"}, {"\\W", "iu", "K"}, {"[^a]", "iv", "A"},
		{"\\p{Lowercase_Letter}", "iu", "A"}, {"\\P{Lowercase_Letter}", "iu", "a"}, {"\\P{Lowercase_Letter}", "iv", "a"},
		{".", "", "🌍"}, {".", "u", "🌍"}, {"🌍+", "", "🌍🌍"},
		{"\\ud83c\\udf0d", "u", "🌍"}, {"\\ud83c\\udf0d+", "u", "🌍🌍"}, {"[\\ud83c\\udf0d]", "u", "🌍"}, {"[\\ud83c\\udf0d-\\ud83c\\udf0f]", "u", "🌎"}, {"(?<=.)a", "u", "🌍a"},
		{"^b.$", "ms", "a\nb\n"}, {"a$", "", "a\n"}, {"(?i:a)(?-i:b)", "", "Ab"},
		{"(?<x>a)|(?<x>b)", "", "b"}, {"(?:(?<x>a)|(?<x>b))\\k<x>", "", "bb"},
	}
	var cases []executionCase
	for _, x := range tests {
		cases = append(cases, executionCase{Pattern: x[0], Flags: x[1], Input: utf16.Encode([]rune(x[2])), Source: "control"})
	}
	for _, f := range []string{"", "g", "y", "gu", "yu"} {
		for _, index := range []uint64{0, 1, 2, 3, 9} {
			cases = append(cases, executionCase{Pattern: ".", Flags: f, Input: utf16.Encode([]rune("🌍a")), LastIndex: index, Source: "lastIndex"})
		}
	}
	for _, f := range []string{"", "u", "v"} {
		cases = append(cases, executionCase{Pattern: ".", Flags: f, Input: []uint16{0xd800, 0x61, 0xdc00}, Source: "lone surrogates"})
	}
	compareExecutionCases(t, cases, true)
}
func TestMatcherRandomNode(t *testing.T) {
	t.Parallel()
	random := rand.New(rand.NewSource(0x8_7_5))
	atoms := []string{"a", "b", ".", "[a-z]", "[^]", "(a)", "(?:a)", "(?=a)", "(?<=a)", "\\d", "\\p{Letter}"}
	extra := []string{"(a|(b))", "(a?b)", "(?!b)", "(?<!b)", "[ab]", "[^a]", "\\w", "\\b", "^", "$", "k", "s", "(a)\\1", "\\1(a)", "(?<x>a)\\k<x>", "(?<=([ab]+)([bc]+))", "[\\q{ab|a|}]", "[[a-z]&&[^aeiou]]"}
	quantifiers := []string{"", "*", "+", "?", "{0}", "{1,}", "{2,4}", "*?", "+?", "??"}
	flags := []string{"", "u", "v", "i", "iu", "iv", "gimsy", "gyu", "ms"}
	inputs := []rune("abc 12\n\réKſ🌍")
	var cases []executionCase
	for len(cases) < 10000 {
		flag := flags[random.Intn(len(flags))]
		pattern := ""
		for j := 0; j < 1+random.Intn(4); j++ {
			atom := atoms[random.Intn(len(atoms))]
			if random.Intn(3) == 0 {
				atom = extra[random.Intn(len(extra))]
				if strings.Contains(atom, "\\q") || strings.Contains(atom, "&&") {
					if !strings.Contains(flag, "v") {
						atom = "[ab]"
					}
				}
			}
			if j > 0 && random.Intn(5) == 0 {
				pattern += "|"
			}
			pattern += atom + quantifiers[random.Intn(len(quantifiers))]
		}
		if _, err := Parse(pattern, flag); err != nil {
			continue
		}
		var input []rune
		for j := 0; j < random.Intn(10); j++ {
			input = append(input, inputs[random.Intn(len(inputs))])
		}
		cases = append(cases, executionCase{Pattern: pattern, Flags: flag, Input: utf16.Encode(input), LastIndex: uint64(random.Intn(5)), Source: "seed 0x875"})
	}
	compareExecutionCases(t, cases, true)
}
func TestMatcherTest262Executions(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("testdata/matches.json.gz")
	if err != nil {
		t.Fatal(err)
	}
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	data, err = io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	var cases []executionCase
	if err = json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	compareExecutionCases(t, cases, false)
}
func TestMatcherStepLimit(t *testing.T) {
	t.Parallel()
	p, err := Compile("(a+)+$", "")
	if err != nil {
		t.Fatal(err)
	}
	r := p.New()
	r.StepLimit = 1000
	_, err = r.ExecString(strings.Repeat("a", 30) + "b")
	if !errors.Is(err, ErrStepLimit) {
		t.Fatalf("catastrophic backtracking returned %v", err)
	}
}

// Probe every generated mapping against the external oracle, including legacy
// mode on Unicode folds that the spec explicitly keeps separate.
func TestMatcherCanonicalizeNode(t *testing.T) {
	t.Parallel()
	var cases []executionCase
	for _, pair := range simpleCaseFold {
		for _, flags := range []string{"i", "iu", "iv"} {
			pattern := fmt.Sprintf("\\u{%X}", pair[0])
			if flags == "i" {
				if pair[0] > 0xffff {
					continue
				}
				pattern = fmt.Sprintf("\\u%04X", pair[0])
			}
			cases = append(cases, executionCase{Pattern: pattern, Flags: flags, Input: utf16.Encode([]rune{pair[1]}), Source: "Unicode 17 simple fold"})
		}
	}
	for _, pair := range legacyUppercase {
		if pair[0] > 0xffff {
			continue
		}
		cases = append(cases, executionCase{Pattern: fmt.Sprintf("\\u%04X", pair[0]), Flags: "i", Input: utf16.Encode([]rune{pair[1]}), Source: "Unicode 17 legacy uppercase"})
	}
	for _, c := range []rune{0xdf, 0x1f80, 0x1fb3, 0xfb00, 0xfb13} {
		cases = append(cases, executionCase{Pattern: fmt.Sprintf("\\u%04X", c), Flags: "i", Input: utf16.Encode([]rune{c + 8}), Source: "full uppercase expansions"})
	}
	compareExecutionCases(t, cases, true)
}

type emojiTestProvider struct{}

func (emojiTestProvider) Lookup(property string) (PropertySet, error) {
	return PropertySet{Strings: [][]rune{[]rune("🌍"), []rune("👩‍💻"), []rune("🏳️‍🌈")}}, nil
}
func TestMatcherPropertyProviderStrings(t *testing.T) {
	t.Parallel()
	cases := []executionCase{
		{Pattern: "[[\\p{RGI_Emoji}]&&[\\q{🌍|aa}]]", Flags: "v", Input: utf16.Encode([]rune("🌍"))},
		{Pattern: "[[\\p{RGI_Emoji}]--[\\q{🌍}]]", Flags: "v", Input: utf16.Encode([]rune("🌍👩‍💻"))},
		{Pattern: "(?<=[\\p{RGI_Emoji}])a", Flags: "v", Input: utf16.Encode([]rune("👩‍💻a"))},
		{Pattern: "[\\p{RGI_Emoji}\\q{a|ab|}]b", Flags: "v", Input: utf16.Encode([]rune("ab"))},
	}
	expected := nodeResults(t, cases)
	for i, c := range cases {
		p, err := CompileWithProperties(c.Pattern, c.Flags, emojiTestProvider{})
		if err != nil {
			t.Fatal(err)
		}
		r := p.New()
		r.StepLimit = 10000
		got, err := matcherResult(r, c.Input)
		if err != nil || !reflect.DeepEqual(got, expected[i]) {
			t.Errorf("property strings %q got=%+v node=%+v error=%v", c.Pattern, got, expected[i], err)
		}
	}
	_, err := CompileWithProperties("\\p{ASCII}", "u", GoProperties{})
	var unavailable *UnavailablePropertyError
	if !errors.As(err, &unavailable) {
		t.Fatalf("missing property silently compiled: %v", err)
	}
}

func TestMatcherUTF16PatternsNode(t *testing.T) {
	t.Parallel()
	var cases []executionCase
	for _, flags := range []string{"", "u", "v"} {
		for _, pattern := range [][]uint16{{0xd800}, {0xdc00}, {0xd83c, 0xdf0d}, {'[', 0xd800, ']'}} {
			for _, input := range [][]uint16{{0xd800}, {0xdc00}, {0xd83c, 0xdf0d}, {0x61, 0xd800, 0xdc00}} {
				cases = append(cases, executionCase{PatternUnits: pattern, Flags: flags, Input: input, Source: fmt.Sprintf("UTF16 pattern %v", pattern)})
			}
		}
	}
	cases = append(cases,
		executionCase{PatternUnits: []uint16{'\\', 0xe9}, Input: []uint16{0xe9}, Source: "non-ASCII legacy identity"},
		executionCase{PatternUnits: []uint16{'\\', 0xd800}, Input: []uint16{0xd800}, Source: "surrogate legacy identity"},
		executionCase{PatternUnits: []uint16{'\\', 0xd83c, 0xdf0d, '+'}, Input: []uint16{0xd83c, 0xdf0d, 0xdf0d}, Source: "astral legacy identity quantifier"},
		executionCase{PatternUnits: []uint16{'\\', 'u', 'D', '8', '3', 'C', 0xdf0d}, Flags: "u", Input: []uint16{0xd83c, 0xdf0d}, Source: "escaped lead and raw trail"},
		executionCase{PatternUnits: []uint16{0xd83c, '\\', 'u', 'D', 'F', '0', 'D'}, Flags: "u", Input: []uint16{0xd83c, 0xdf0d}, Source: "raw lead and escaped trail"},
	)
	compareExecutionCases(t, cases, true)
}

type storedPropertyProvider struct{ data PropertySet }

func (p storedPropertyProvider) Lookup(string) (PropertySet, error) { return p.data, nil }
func TestMatcherProviderSnapshot(t *testing.T) {
	t.Parallel()
	c := executionCase{Pattern: "\\p{RGI_Emoji}", Flags: "v", Input: utf16.Encode([]rune("👩‍💻")), Source: "provider snapshot"}
	expected := nodeResults(t, []executionCase{c})[0]
	data := PropertySet{Strings: [][]rune{[]rune("👩‍💻")}}
	p, err := CompileWithProperties(c.Pattern, c.Flags, storedPropertyProvider{data})
	if err != nil {
		t.Fatal(err)
	}
	data.Strings[0][0] = 'x'
	r := p.New()
	r.StepLimit = 10000
	got, err := matcherResult(r, c.Input)
	if err != nil || !reflect.DeepEqual(got, expected) {
		t.Errorf("property provider snapshot got=%+v node=%+v error=%v", got, expected, err)
	}
}
