package regexp

import (
	"bytes"
	"context"
	"encoding/json"
	"os/exec"
	"reflect"
	"testing"
	"time"
	"unicode/utf16"
)

// These cases combine features that the extracted test262 executions seldom combine.
func oct6Cases() []executionCase {
	tests := [][3]string{
		{`^((a(b)?)+)+$`, "", "aba"},
		{`^((a(b)?)+)+$`, "", "abab"},
		{`^((a(b)?)+)+$`, "", "abb"},
		{`^((a|ab)+?)(b?)$`, "", "aab"},
		{`^((a?){2,3}){2}$`, "", "aa"},
		{`^((a*)(b?))+c$`, "", "aabc"},
		{`^((ab|a){1,3}){2,3}?$`, "", "abaab"},
		{`(?<=((a)|(ba)))c`, "", "bac"},
		{`(?<=((ba)|(a)))c`, "", "bac"},
		{`(?<=((ab)+|(cb)+))d`, "", "ababd"},
		{`(?<=((ab)+|(cb)+))d`, "", "cbcbd"},
		{`(?<!ab|bc)c`, "", "abc bcc ac"},
		{`(?<=((a+)|(ba+)))b`, "", "baaab"},
		{`(?<=((a)|(🌍)))b`, "u", "🌍b"},
		{`^((a|b)\2)+$`, "", "aabb"},
		{`^((a|b)\2)+$`, "", "aaba"},
		{`^(?:(a|(b))\2)+$`, "", "bba"},
		{`^((a)?b\2)+$`, "", "abab"},
		{`^((a)?b\2)+$`, "", "ababb"},
		{`^((a|ab)\2)+?c$`, "", "aababc"},
		{`^(?:(?<letter>a|b)\k<letter>)+$`, "", "bbaa"},
		{`^(é+)\1$`, "i", "éÉÉé"},
		{`^([σς]+)\1$`, "iu", "ΣςσΣ"},
		{`^(я+)\1$`, "i", "яЯЯя"},
		{`^(ſ+)\1$`, "i", "ſs"},
		{`^(ſ+)\1$`, "iu", "ſs"},
		{`^([kK]+)\1$`, "i", "Kk"},
		{`^([kK]+)\1$`, "iv", "Kk"},
		{`^(ß)\1$`, "iu", "ßSS"},
		{`^(𐐀+)\1$`, "iu", "𐐀𐐨"},
	}
	cases := make([]executionCase, 0, len(tests))
	for _, test := range tests {
		cases = append(cases, executionCase{Pattern: test[0], Flags: test[1], Input: utf16.Encode([]rune(test[2])), Source: "oct6 combinations"})
	}
	return cases
}

func TestMatcherOct6Node(t *testing.T) {
	t.Parallel()
	compareExecutionCases(t, oct6Cases(), true)
}

type oct6Loop struct {
	Name      string   `json:"name"`
	Pattern   string   `json:"pattern"`
	Flags     string   `json:"flags"`
	Input     []uint16 `json:"input"`
	LastIndex uint64   `json:"lastIndex"`
	Steps     int      `json:"steps"`
	// Exec does not advance an empty match. Some loops deliberately repeat it;
	// others model a caller advancing by UTF-16 code unit or Unicode code point.
	AdvanceEmpty bool `json:"advanceEmpty"`
}

func oct6Loops() []oct6Loop {
	var loops []oct6Loop
	add := func(name, pattern, flags, input string, index uint64, steps int, advance bool) {
		loops = append(loops, oct6Loop{name, pattern, flags, utf16.Encode([]rune(input)), index, steps, advance})
	}
	for _, flags := range []string{"g", "y", "gy"} {
		add("gap "+flags, `(a)|(b)`, flags, "ab ab", 0, 7, false)
		add("initial gap "+flags, `a+`, flags, " a aa", 1, 6, false)
		add("past end "+flags, `a`, flags, "a", 9, 3, false)
		add("empty unchanged "+flags, `(?:)`, flags, "ab", 1, 3, false)
		add("empty at end "+flags, `$`, flags, "ab", 2, 3, false)
		add("empty advanced "+flags, `(a?)`, flags, "ba", 0, 6, true)
	}
	for _, flags := range []string{"g", "gu", "gv", "y", "yu", "yv"} {
		add("astral empty "+flags, `(?:)`, flags, "🌍a", 0, 6, true)
		add("astral consuming "+flags, `(.)`, flags, "🌍a", 0, 5, false)
		add("inside pair "+flags, `(?:)`, flags, "🌍a", 1, 3, false)
	}
	add("lookahead empties", `(?=(a))`, "g", "aba", 0, 6, true)
	add("lookbehind empties", `(?<=(a|b))`, "gy", "ab", 1, 5, true)
	return loops
}

func oct6NodeLoops(t *testing.T, loops []oct6Loop) [][]executionResult {
	t.Helper()
	data, err := json.Marshal(loops)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, "node", "-e", `
if(!process.version.startsWith('v24.'))throw Error('Node 24 required');
const cases=JSON.parse(require('fs').readFileSync(0,'utf8'));
const results=cases.map(c=>{
 const r=new RegExp(c.pattern,c.flags+'d'), input=String.fromCharCode(...c.input);
 r.lastIndex=c.lastIndex;const results=[];
 for(let step=0;step<c.steps;step++){
  const m=r.exec(input);
  results.push({captures:m?Array.from(m.indices,x=>x??null):null,lastIndex:r.lastIndex,groups:m?.indices.groups??null});
  if(c.advanceEmpty&&m&&m[0].length===0){
   const at=r.lastIndex, unicode=c.flags.includes('u')||c.flags.includes('v');
   const first=input.charCodeAt(at), second=input.charCodeAt(at+1);
   r.lastIndex=at+(unicode&&first>=0xd800&&first<=0xdbff&&second>=0xdc00&&second<=0xdfff?2:1);
  }
 }
 return results;
});process.stdout.write(JSON.stringify(results));`)
	command.Stdin = bytes.NewReader(data)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("Node loop oracle: %v\n%s", err, output)
	}
	var results [][]executionResult
	if err := json.Unmarshal(output, &results); err != nil {
		t.Fatal(err)
	}
	if len(results) != len(loops) {
		t.Fatal("Node loop result count")
	}
	return results
}

func TestMatcherOct6LoopsNode(t *testing.T) {
	t.Parallel()
	loops := oct6Loops()
	expected := oct6NodeLoops(t, loops)
	executions := 0
	for index, loop := range loops {
		t.Run(loop.Name, func(t *testing.T) {
			program, err := Compile(loop.Pattern, loop.Flags)
			if err != nil {
				t.Fatal(err)
			}
			matcher := program.New()
			matcher.LastIndex = loop.LastIndex
			matcher.StepLimit = 100000
			if len(expected[index]) != loop.Steps {
				t.Fatal("Node step count")
			}
			for step := 0; step < loop.Steps; step++ {
				got, err := matcherResult(matcher, loop.Input)
				if err != nil {
					t.Fatal(err)
				}
				executions++
				if !reflect.DeepEqual(got, expected[index][step]) {
					t.Errorf("step %d pattern=%q flags=%q: matcher=%+v Node=%+v", step, loop.Pattern, loop.Flags, got, expected[index][step])
				}
				if loop.AdvanceEmpty && got.Captures != nil && got.Captures[0][0] == got.Captures[0][1] {
					at := matcher.LastIndex
					matcher.LastIndex++
					if unicodeMode(program.flags) && at+1 < uint64(len(loop.Input)) && high(loop.Input[at]) && low(loop.Input[at+1]) {
						matcher.LastIndex++
					}
				}
			}
		})
	}
	t.Logf("%d stateful loops, %d executions compared with Node", len(loops), executions)
}
