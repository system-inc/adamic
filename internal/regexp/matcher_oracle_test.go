package regexp

import (
	"bytes"
	"compress/gzip"
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
	"unicode/utf16"

	"github.com/system-inc/adamic/internal/childguard"
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

func matcherResult(exp *RegExp, input []uint16) (executionResult, error) {
	match, execError := exp.Exec(input)
	out := executionResult{LastIndex: exp.LastIndex}
	if match != nil {
		out.Captures = make([][]int, len(match.Captures))
		for index, capture := range match.Captures {
			if capture.Start >= 0 {
				out.Captures[index] = []int{capture.Start, capture.End}
			}
		}
		if len(match.Groups) > 0 {
			out.Groups = map[string][]int{}
			for name, value := range match.Groups {
				var indices []int
				if value.Start >= 0 {
					indices = []int{value.Start, value.End}
				}
				out.Groups[name] = indices
			}
		}
	}
	return out, execError
}
func nodeResults(test *testing.T, cases []executionCase) []executionResult {
	test.Helper()
	data, encodeError := json.Marshal(cases)
	if encodeError != nil {
		test.Fatal(encodeError)
	}
	command := exec.Command("node", "-e", `
const cases=JSON.parse(require('fs').readFileSync(0,'utf8'));
if(!process.version.startsWith('v24.'))throw Error('Node 24 required');
const fs=require('fs');fs.writeSync(2,'regex progress: start\n');
const results=cases.map((c,index)=>{
 const pattern=c.patternUnits?String.fromCharCode(...c.patternUnits):c.pattern;
 const r=new RegExp(pattern,c.flags.includes('d')?c.flags:c.flags+'d');
 r.lastIndex=c.lastIndex;const input=String.fromCharCode(...c.input);
 const m=r.exec(input);
 if((index+1)%128===0) fs.writeSync(2,'regex progress: '+(index+1)+'/'+cases.length+'\n');
 return {captures:m?Array.from(m.indices,x=>x??null):null,lastIndex:r.lastIndex,groups:m?.indices.groups?Object.fromEntries(Object.entries(m.indices.groups).map(([k,v])=>[k,v??null])):null};
});process.stdout.write(JSON.stringify(results));`)
	command.Stdin = bytes.NewReader(data)
	var output, progress bytes.Buffer
	command.Stdout, command.Stderr = &output, &progress
	if runError := childguard.Run(command, childguard.Options{}); runError != nil {
		test.Fatalf("Node execution oracle: %v\n%s\n%s", runError, progress.Bytes(), output.Bytes())
	}
	var results []executionResult
	if decodeError := json.Unmarshal(output.Bytes(), &results); decodeError != nil {
		test.Fatalf("Node decode: %v\n%s", decodeError, output.Bytes())
	}
	if len(results) != len(cases) {
		test.Fatal("Node result count")
	}
	return results
}
func compareExecutionCases(test *testing.T, cases []executionCase, external bool) {
	test.Helper()
	var results []executionResult
	if external {
		results = nodeResults(test, cases)
	}
	programs := map[string]*Program{}
	unavailablePrograms := map[string]error{}
	compared, unavailable, regularCases := 0, 0, 0
	for index, executionCase := range cases {
		key := executionCase.Pattern + "\x00" + executionCase.Flags + fmt.Sprint(executionCase.PatternUnits)
		if _, missing := unavailablePrograms[key]; missing {
			unavailable++
			continue
		}
		program := programs[key]
		if program == nil {
			var err error
			if executionCase.PatternUnits != nil {
				program, err = CompileUTF16(executionCase.PatternUnits, executionCase.Flags)
			} else {
				program, err = Compile(executionCase.Pattern, executionCase.Flags)
			}
			if err != nil {
				var unavailableProperty *UnavailablePropertyError
				if !external && errors.As(err, &unavailableProperty) {
					unavailable++
					unavailablePrograms[key] = err
					test.Logf(
						"UNAVAILABLE source=%s pattern=%q flags=%q error=%v",
						executionCase.Source,
						executionCase.Pattern,
						executionCase.Flags,
						err,
					)
					continue
				}
				test.Errorf(
					"compile source=%s pattern=%q flags=%q: %v",
					executionCase.Source,
					executionCase.Pattern,
					executionCase.Flags,
					err,
				)
				continue
			}
			programs[key] = program
		}
		if program.regular != nil {
			regularCases++
		}
		for mode := 0; mode < 3; mode++ {
			exp := program.New()
			exp.LastIndex = executionCase.LastIndex
			if mode == 0 {
				exp.StepLimit = 10_000_000
			}
			exp.DisableRegular = mode == 2
			got, matcherResultError := matcherResult(exp, executionCase.Input)
			want := executionCase.Expected
			if external {
				want = results[index]
			}
			if matcherResultError != nil {
				test.Errorf(
					"execution source=%s pattern=%q flags=%q input=%v lastIndex=%d: %v",
					executionCase.Source,
					executionCase.Pattern,
					executionCase.Flags,
					executionCase.Input,
					executionCase.LastIndex,
					matcherResultError,
				)
				continue
			}
			if mode == 0 {
				compared++
			}
			if !reflect.DeepEqual(got, want) {
				test.Errorf(
					"DISAGREEMENT source=%s pattern=%q flags=%q input=%v lastIndex=%d got=%+v node=%+v",
					executionCase.Source,
					executionCase.Pattern,
					executionCase.Flags,
					executionCase.Input,
					executionCase.LastIndex,
					got,
					want,
				)
			}
		}
	}
	test.Logf(
		"execution totals: %d cases, %d compared, %d property stand-in unavailable",
		len(cases),
		compared,
		unavailable,
	)
	test.Logf("three modes: budgeted VM, regular with recovery, unlimited VM; %d eligible executions", regularCases)
}
func TestMatcherNodeControls(test *testing.T) {
	test.Parallel()
	tests := [][3]string{
		{"[^[\\q{ab|a}]&&[a]]", "v", "abc"},
		{"\\u{D83C}\\u{DF0D}", "u", "🌍"},
		{"\\uD83C\\u{DF0D}", "u", "🌍"},
		{"\\u{D83C}\\uDF0D", "u", "🌍"},
		{"\\c", "", "\\c"},
		{"\\c+", "", "\\ccc"},
		{"[\\c_]", "", "\x1f"},
		{"[\\c0]", "", "\x10"},
		{"\\8", "", "8"},
		{"\\9", "", "9"},
		{"\\400", "", " 0"},
		{"\\777", "", "?7"},
		{"\\1234", "", "S4"},
		{"\\uZ", "", "uZ"},
		{"\\xZ", "", "xZ"},
		{"\\u{3}", "", "uuu"},
		{"\\k", "", "k"},
		{"[🌍]", "", "🌍"},
		{"[a-🌍]", "", "🌍"},
		{"[🌍-🌎]", "u", "🌎"},
		{"[\\-]", "u", "-"},
		{"[\\!]", "v", "!"},
		{"[\\q{a\\-b}]", "v", "a-b"},
		{"[[]", "", "["},
		{"[\\d-a]+", "", "a-123"},
		{"[\\n-\\r]+", "", "\n\r"},
		{"a+", "", "aaa"},
		{"a+?", "", "aaa"},
		{"(a|(b))+", "", "aba"},
		{"(?<=([ab]+)([bc]+))$", "", "abc"},
		{"(?<=\\1(a))b", "", "aab"},
		{"(?=(a+))a*b\\1", "", "baabac"},
		{"(?!((a)))b\\1", "", "b"},
		{"\\1(a)", "", "a"},
		{"(a)?\\1", "", ""},
		{"(a*)*", "", "aaa"},
		{"(a?){2,4}", "", "a"},
		{"(a?){2,4}?", "", "a"},
		{"(?=a)?a", "", "a"},
		{"[[a-z]&&[^aeiou]]+", "v", "aeibcd"},
		{"[\\q{ab|a|}]b", "v", "ab"},
		{"[[\\q{ab|cd}]--[\\q{ab}]]", "v", "abcd"},
		{"[[\\q{ab|cd}]&&[\\q{cd|ef}]]", "v", "abcd"},
		{"k", "i", "K"},
		{"k", "iu", "K"},
		{"s", "i", "ſ"},
		{"s", "iu", "ſ"},
		{"[a-z]", "iu", "K"},
		{"\\W", "iu", "K"},
		{"[^a]", "iv", "A"},
		{"\\p{Lowercase_Letter}", "iu", "A"},
		{"\\P{Lowercase_Letter}", "iu", "a"},
		{"\\P{Lowercase_Letter}", "iv", "a"},
		{".", "", "🌍"},
		{".", "u", "🌍"},
		{"🌍+", "", "🌍🌍"},
		{"\\ud83c\\udf0d", "u", "🌍"},
		{"\\ud83c\\udf0d+", "u", "🌍🌍"},
		{"[\\ud83c\\udf0d]", "u", "🌍"},
		{"[\\ud83c\\udf0d-\\ud83c\\udf0f]", "u", "🌎"},
		{"(?<=.)a", "u", "🌍a"},
		{"^b.$", "ms", "a\nb\n"},
		{"a$", "", "a\n"},
		{"(?i:a)(?-i:b)", "", "Ab"},
		{"(?<x>a)|(?<x>b)", "", "b"},
		{"(?:(?<x>a)|(?<x>b))\\k<x>", "", "bb"},
	}
	var cases []executionCase
	for _, x := range tests {
		cases = append(
			cases,
			executionCase{Pattern: x[0], Flags: x[1], Input: utf16.Encode([]rune(x[2])), Source: "control"},
		)
	}
	for _, f := range []string{"", "g", "y", "gu", "yu"} {
		for _, index := range []uint64{0, 1, 2, 3, 9} {
			cases = append(cases, executionCase{
				Pattern:   ".",
				Flags:     f,
				Input:     utf16.Encode([]rune("🌍a")),
				LastIndex: index,
				Source:    "lastIndex",
			})
		}
	}
	for _, f := range []string{"", "u", "v"} {
		cases = append(
			cases,
			executionCase{Pattern: ".", Flags: f, Input: []uint16{0xd800, 0x61, 0xdc00}, Source: "lone surrogates"},
		)
	}
	compareExecutionCases(test, cases, true)
}
func runMatcherRandomNode(test *testing.T, shard int) {
	test.Helper()
	random := rand.New(rand.NewSource(0x8_7_5))
	atoms := []string{"a", "b", ".", "[a-z]", "[^]", "(a)", "(?:a)", "(?=a)", "(?<=a)", "\\d", "\\p{Letter}"}
	extra := []string{
		"(a|(b))",
		"(a?b)",
		"(?!b)",
		"(?<!b)",
		"[ab]",
		"[^a]",
		"\\w",
		"\\b",
		"^",
		"$",
		"k",
		"s",
		"(a)\\1",
		"\\1(a)",
		"(?<x>a)\\k<x>",
		"(?<=([ab]+)([bc]+))",
		"[\\q{ab|a|}]",
		"[[a-z]&&[^aeiou]]",
	}
	quantifiers := []string{"", "*", "+", "?", "{0}", "{1,}", "{2,4}", "*?", "+?", "??"}
	flags := []string{"", "u", "v", "i", "iu", "iv", "gimsy", "gyu", "ms"}
	inputs := []rune("abc 12\n\réKſ🌍")
	var cases []executionCase
	for len(cases) < 10000 {
		flag := flags[random.Intn(len(flags))]
		pattern := ""
		for index := 0; index < 1+random.Intn(4); index++ {
			atom := atoms[random.Intn(len(atoms))]
			if random.Intn(3) == 0 {
				atom = extra[random.Intn(len(extra))]
				if strings.Contains(atom, "\\q") || strings.Contains(atom, "&&") {
					if !strings.Contains(flag, "v") {
						atom = "[ab]"
					}
				}
			}
			if index > 0 && random.Intn(5) == 0 {
				pattern += "|"
			}
			pattern += atom + quantifiers[random.Intn(len(quantifiers))]
		}
		if _, parseError := Parse(pattern, flag); parseError != nil {
			continue
		}
		var input []rune
		for index := 0; index < random.Intn(10); index++ {
			input = append(input, inputs[random.Intn(len(inputs))])
		}
		cases = append(cases, executionCase{
			Pattern:   pattern,
			Flags:     flag,
			Input:     utf16.Encode(input),
			LastIndex: uint64(random.Intn(5)),
			Source:    "seed 0x875",
		})
	}
	compareExecutionCases(test, matcherExecutionShard(cases, shard), true)
}
func runMatcherTest262Executions(test *testing.T, shard int) {
	test.Helper()
	data, readError := os.ReadFile("testdata/matches.json.gz")
	if readError != nil {
		test.Fatal(readError)
	}
	reader, gzipError := gzip.NewReader(bytes.NewReader(data))
	if gzipError != nil {
		test.Fatal(gzipError)
	}
	defer reader.Close()
	data, readError = io.ReadAll(reader)
	if readError != nil {
		test.Fatal(readError)
	}
	var cases []executionCase
	if decodeError := json.Unmarshal(data, &cases); decodeError != nil {
		test.Fatal(decodeError)
	}
	compareExecutionCases(test, matcherExecutionShard(cases, shard), false)
}

// Probe every generated mapping against the external oracle, including legacy
// mode on Unicode folds that the spec explicitly keeps separate.
func TestMatcherCanonicalizeNode(test *testing.T) {
	test.Parallel()
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
			cases = append(cases, executionCase{
				Pattern: pattern,
				Flags:   flags,
				Input:   utf16.Encode([]rune{pair[1]}),
				Source:  "Unicode 17 simple fold",
			})
		}
	}
	for _, pair := range legacyUppercase {
		if pair[0] > 0xffff {
			continue
		}
		cases = append(cases, executionCase{
			Pattern: fmt.Sprintf("\\u%04X", pair[0]),
			Flags:   "i",
			Input:   utf16.Encode([]rune{pair[1]}),
			Source:  "Unicode 17 legacy uppercase",
		})
	}
	for _, character := range []rune{0xdf, 0x1f80, 0x1fb3, 0xfb00, 0xfb13} {
		cases = append(cases, executionCase{
			Pattern: fmt.Sprintf("\\u%04X", character),
			Flags:   "i",
			Input:   utf16.Encode([]rune{character + 8}),
			Source:  "full uppercase expansions",
		})
	}
	compareExecutionCases(test, cases, true)
}

type emojiTestProvider struct{}

func (emojiTestProvider) Lookup(property string) (PropertySet, error) {
	return PropertySet{Strings: [][]rune{[]rune("🌍"), []rune("👩‍💻"), []rune("🏳️‍🌈")}}, nil
}
func TestMatcherPropertyProviderStrings(test *testing.T) {
	test.Parallel()
	cases := []executionCase{
		{Pattern: "[[\\p{RGI_Emoji}]&&[\\q{🌍|aa}]]", Flags: "v", Input: utf16.Encode([]rune("🌍"))},
		{Pattern: "[[\\p{RGI_Emoji}]--[\\q{🌍}]]", Flags: "v", Input: utf16.Encode([]rune("🌍👩‍💻"))},
		{Pattern: "(?<=[\\p{RGI_Emoji}])a", Flags: "v", Input: utf16.Encode([]rune("👩‍💻a"))},
		{Pattern: "[\\p{RGI_Emoji}\\q{a|ab|}]b", Flags: "v", Input: utf16.Encode([]rune("ab"))},
	}
	expected := nodeResults(test, cases)
	for index, executionCase := range cases {
		program, compileWithPropertiesError := CompileWithProperties(
			executionCase.Pattern,
			executionCase.Flags,
			emojiTestProvider{},
		)
		if compileWithPropertiesError != nil {
			test.Fatal(compileWithPropertiesError)
		}
		exp := program.New()
		exp.StepLimit = 10000
		got, matcherResultError := matcherResult(exp, executionCase.Input)
		if matcherResultError != nil || !reflect.DeepEqual(got, expected[index]) {
			test.Errorf(
				"property strings %q got=%+v node=%+v error=%v",
				executionCase.Pattern,
				got,
				expected[index],
				matcherResultError,
			)
		}
	}
	_, compileWithPropertiesError := CompileWithProperties("\\p{ASCII}", "u", GoProperties{})
	var unavailable *UnavailablePropertyError
	if !errors.As(compileWithPropertiesError, &unavailable) {
		test.Fatalf("missing property silently compiled: %v", compileWithPropertiesError)
	}
}

func TestMatcherUTF16PatternsNode(test *testing.T) {
	test.Parallel()
	var cases []executionCase
	for _, flags := range []string{"", "u", "v"} {
		for _, pattern := range [][]uint16{{0xd800}, {0xdc00}, {0xd83c, 0xdf0d}, {'[', 0xd800, ']'}} {
			for _, input := range [][]uint16{{0xd800}, {0xdc00}, {0xd83c, 0xdf0d}, {0x61, 0xd800, 0xdc00}} {
				cases = append(cases, executionCase{
					PatternUnits: pattern,
					Flags:        flags,
					Input:        input,
					Source:       fmt.Sprintf("UTF16 pattern %v", pattern),
				})
			}
		}
	}
	cases = append(cases,
		executionCase{PatternUnits: []uint16{'\\', 0xe9}, Input: []uint16{0xe9}, Source: "non-ASCII legacy identity"},
		executionCase{
			PatternUnits: []uint16{'\\', 0xd800},
			Input:        []uint16{0xd800},
			Source:       "surrogate legacy identity",
		},
		executionCase{
			PatternUnits: []uint16{'\\', 0xd83c, 0xdf0d, '+'},
			Input:        []uint16{0xd83c, 0xdf0d, 0xdf0d},
			Source:       "astral legacy identity quantifier",
		},
		executionCase{
			PatternUnits: []uint16{'\\', 'u', 'D', '8', '3', 'C', 0xdf0d},
			Flags:        "u",
			Input:        []uint16{0xd83c, 0xdf0d},
			Source:       "escaped lead and raw trail",
		},
		executionCase{
			PatternUnits: []uint16{0xd83c, '\\', 'u', 'D', 'F', '0', 'D'},
			Flags:        "u",
			Input:        []uint16{0xd83c, 0xdf0d},
			Source:       "raw lead and escaped trail",
		},
	)
	compareExecutionCases(test, cases, true)
}

type storedPropertyProvider struct{ data PropertySet }

func (provider storedPropertyProvider) Lookup(string) (PropertySet, error) { return provider.data, nil }
func TestMatcherProviderSnapshot(test *testing.T) {
	test.Parallel()
	c := executionCase{
		Pattern: "\\p{RGI_Emoji}",
		Flags:   "v",
		Input:   utf16.Encode([]rune("👩‍💻")),
		Source:  "provider snapshot",
	}
	expected := nodeResults(test, []executionCase{c})[0]
	data := PropertySet{Strings: [][]rune{[]rune("👩‍💻")}}
	program, compileWithPropertiesError := CompileWithProperties(c.Pattern, c.Flags, storedPropertyProvider{data})
	if compileWithPropertiesError != nil {
		test.Fatal(compileWithPropertiesError)
	}
	data.Strings[0][0] = 'x'
	exp := program.New()
	exp.StepLimit = 10000
	got, matcherResultError := matcherResult(exp, c.Input)
	if matcherResultError != nil || !reflect.DeepEqual(got, expected) {
		test.Errorf("property provider snapshot got=%+v node=%+v error=%v", got, expected, matcherResultError)
	}
}

// Corpus rows retain their source order and belong to exactly one of four units.
func matcherExecutionShard(cases []executionCase, shard int) []executionCase {
	var selected []executionCase
	for index, executionCase := range cases {
		if index%4 == shard {
			selected = append(selected, executionCase)
		}
	}
	return selected
}

func TestMatcherRandomNodeShard0(test *testing.T) {
	test.Parallel()
	runMatcherRandomNode(test, 0)
}

func TestMatcherRandomNodeShard1(test *testing.T) {
	test.Parallel()
	runMatcherRandomNode(test, 1)
}

func TestMatcherRandomNodeShard2(test *testing.T) {
	test.Parallel()
	runMatcherRandomNode(test, 2)
}

func TestMatcherRandomNodeShard3(test *testing.T) {
	test.Parallel()
	runMatcherRandomNode(test, 3)
}

func TestMatcherTest262ExecutionsShard0(test *testing.T) {
	test.Parallel()
	runMatcherTest262Executions(test, 0)
}

func TestMatcherTest262ExecutionsShard1(test *testing.T) {
	test.Parallel()
	runMatcherTest262Executions(test, 1)
}

func TestMatcherTest262ExecutionsShard2(test *testing.T) {
	test.Parallel()
	runMatcherTest262Executions(test, 2)
}

func TestMatcherTest262ExecutionsShard3(test *testing.T) {
	test.Parallel()
	runMatcherTest262Executions(test, 3)
}
