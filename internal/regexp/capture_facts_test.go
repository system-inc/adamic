package regexp

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
)

type captureLiteral struct {
	File    string               `json:"file"`
	Line    int                  `json:"line"`
	Pattern string               `json:"pattern"`
	Flags   string               `json:"flags"`
	Facts   CaptureParticipation `json:"facts"`
}

type captureCase struct {
	Pattern, Flags, Input, Source string
	LastIndex                     int
	InputUnits                    []uint16
}
type captureObservation struct {
	Match   bool
	Indices []bool
	Names   map[string]bool
}

func captureNode(t *testing.T, cases []captureCase) []captureObservation {
	t.Helper()
	data, err := json.Marshal(cases)
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("node", "-e", `if(process.version !== 'v24.19.0') throw Error('requires Node 24.19.0'); const cs=JSON.parse(require('fs').readFileSync(0,'utf8')); console.log(JSON.stringify(cs.map(c=>{const r=new RegExp(c.Pattern,c.Flags);r.lastIndex=c.LastIndex;const input=c.InputUnits===null?c.Input:String.fromCharCode(...c.InputUnits);const m=r.exec(input);return m?{Match:true,Indices:Array.from(m,x=>x!==undefined),Names:Object.fromEntries(Object.entries(m.groups||{}).map(([k,v])=>[k,v!==undefined]))}:{Match:false};})));`)
	cmd.Stdin = bytes.NewReader(data)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Node: %v %s", err, out)
	}
	var observations []captureObservation
	if err = json.Unmarshal(out, &observations); err != nil {
		t.Fatal(err)
	}
	if len(observations) != len(cases) {
		t.Fatal("incomplete Node response")
	}
	return observations
}

func assertCaptureFacts(t *testing.T, cases []captureCase) []captureObservation {
	t.Helper()
	observed := captureNode(t, cases)
	matches := 0
	for i, c := range cases {
		facts, err := CaptureFacts(c.Pattern, c.Flags)
		if err != nil {
			t.Fatalf("%s /%s/%s: %v", c.Source, c.Pattern, c.Flags, err)
		}
		m := observed[i]
		if !m.Match {
			continue
		}
		matches++
		if len(m.Indices) != len(facts.Indices) {
			t.Fatalf("capture count /%s/: facts=%d Node=%d", c.Pattern, len(facts.Indices), len(m.Indices))
		}
		for j, always := range facts.Indices {
			if always && !m.Indices[j] {
				t.Errorf("Node undefined: %s /%s/%s input %q capture %d marked always", c.Source, c.Pattern, c.Flags, c.Input, j)
			}
		}
		for name, always := range facts.Names {
			if always && !m.Names[name] {
				t.Errorf("Node undefined: %s /%s/%s input %q name %s marked always", c.Source, c.Pattern, c.Flags, c.Input, name)
			}
		}
	}
	t.Logf("Node: %d inputs, %d successful matches", len(cases), matches)
	return observed
}

func TestCaptureFactsNodeBranches(t *testing.T) {
	t.Parallel()
	var cases []captureCase
	for _, pattern := range []string{`(a)?(b)`, `(a)|(b)`, `((a)|(b))+`, `(a?)(b*)`, `(?=(a))a`, `(?!(a))b`, `(?<=(a))b`, `(?<!(a))b`, `(a)?\1`, `\1(a)`, `(?<x>a)|(?<x>b)`, `(?:(?<x>a)|b)+`, `((a)?){2}`, `(?=(a|))`, `(?:(?=(a)))*`, `((a)?)+?`} {
		for _, input := range []string{"", "a", "b", "ab", "ba", "aa", "bb", "aba", "bab"} {
			cases = append(cases, captureCase{Pattern: pattern, Input: input, Source: "branch/reset/backreference witness"})
		}
	}
	assertCaptureFacts(t, cases)
}

func TestCaptureFactsTest262Node(t *testing.T) {
	t.Parallel()
	f, err := os.Open("testdata/matches.json.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	zip, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	defer zip.Close()
	var rows []struct {
		Pattern, Flags, Source string
		Input                  []uint16
		LastIndex              int
	}
	if err = json.NewDecoder(zip).Decode(&rows); err != nil {
		t.Fatal(err)
	}
	var cases []captureCase
	patterns := map[string]bool{}
	for _, r := range rows {
		if strings.Contains(r.Source, "named-groups/") || strings.Contains(r.Source, "match-indices/") {
			cases = append(cases, captureCase{Pattern: r.Pattern, Flags: r.Flags, Source: r.Source, LastIndex: r.LastIndex, InputUnits: r.Input})
			patterns[r.Pattern+"/"+r.Flags] = true
		}
	}
	if len(cases) == 0 {
		t.Fatal("missing test262 captures corpus")
	}
	t.Logf("test262: %d pattern/flag pairs", len(patterns))
	assertCaptureFacts(t, cases)
}

// All twelve callback sites use these literal patterns; quote escaping has
// three alternatives and JSX escaping two. JSX entities exercise each capture
// alternative (decimal, hexadecimal, word); the other sites have no captures
// except diagnostic substitution's required numeric group.
// Not parallel: reads testdata/capture-tsc.json, which TestCaptureFactsCompilerTable can rewrite.
func TestCaptureFactsTSCCallbacksNode(t *testing.T) {
	var cases []captureCase
	patterns := []struct{ pattern, input, site string }{
		{`\\.`, `\"\a`, `checker.ts:8943`},
		{`[^\u0130\u0131\u00DFa-z0-9\\/:\-_. ]+`, `FILE.TS`, `core.ts:1877`},
		{`\w`, `Ab0_`, `sys.ts:1733`},
		{`&((#((\d+)|x([\da-fA-F]+)))|(\w+));`, `&#65; &#x41; &amp;`, `transformers/jsx.ts:624`},
		{`[\\"\u0000-\u001f\u2028\u2029\u0085]`, `"\n\t\\`, `utilities.ts:6211 double quote`},
		{`[\\'\u0000-\u001f\u2028\u2029\u0085]`, `'\n\t\\`, `utilities.ts:6211 single quote`},
		{"\\r\\n|[\\\\`\\u0000-\\u0009\\u000b-\\u001f\\u2028\\u2029\\u0085]", "\r\n`\t\\", "utilities.ts:6211 backtick"},
		{`[^\u0000-\u007F]`, `é中`, `utilities.ts:6221`},
		{`["\u0000-\u001f\u2028\u2029\u0085]`, `"\n`, `utilities.ts:6252 double quote`},
		{`['\u0000-\u001f\u2028\u2029\u0085]`, `'\n`, `utilities.ts:6252 single quote`},
		{`\{(\d+)\}`, `{0} {12}`, `utilities.ts:8584`},
		{`[^\w\s/]`, `*.?`, `utilities.ts:9598,9746,9761`},
		{`\$`, `$x$`, `utilities.ts:10916`},
	}
	for _, p := range patterns {
		for offset := 0; offset <= len([]rune(p.input)); offset++ {
			cases = append(cases, captureCase{Pattern: p.pattern, Flags: "g", Input: p.input, Source: p.site, LastIndex: offset})
		}
	}
	observations := assertCaptureFacts(t, cases)
	matched := map[string]bool{}
	entityBranches := map[int]bool{}
	for i, m := range observations {
		if !m.Match {
			continue
		}
		matched[cases[i].Pattern] = true
		if cases[i].Source == "transformers/jsx.ts:624" {
			for _, index := range []int{4, 5, 6} {
				if m.Indices[index] {
					entityBranches[index] = true
				}
			}
		}
	}
	data, err := os.ReadFile("testdata/capture-tsc.json")
	if err != nil {
		t.Fatal(err)
	}
	var table []captureLiteral
	if err = json.Unmarshal(data, &table); err != nil {
		t.Fatal(err)
	}
	for _, p := range patterns {
		if !matched[p.pattern] {
			t.Errorf("callback pattern unexercised: %s", p.site)
		}
		sourceLiteral := false
		for _, literal := range table {
			if literal.Pattern == p.pattern && literal.Flags == "g" {
				sourceLiteral = true
			}
		}
		if !sourceLiteral {
			t.Errorf("callback pattern absent from compiler source: %s", p.site)
		}
	}
	if len(entityBranches) != 3 {
		t.Fatal("JSX decimal, hex and word branches were not all exercised")
	}
}

func TestProduct_CaptureFactsOptionalMutant(t *testing.T) {
	t.Parallel()
	captureFactsOptionalMutant(t)
}

func captureFactsOptionalMutant(t testing.TB) string {
	t.Helper()
	return buildcache.GoTestMutated(t, "", "capture-facts-optional.test", regexpMutantSources(t, "capture_facts_test.go"), nil, []buildcache.Mutation{{
		File:   "internal/regexp/capture_facts.go",
		Before: "if n.Min.Sign() > 0 {",
		After:  "if n.Min.Sign() >= 0 {",
	}})
}

func TestCaptureFactsOptionalMutant(t *testing.T) {
	t.Parallel()
	command := exec.Command(captureFactsOptionalMutant(t), "-test.run=^TestCaptureFactsNodeBranches$", "-test.count=1", "-test.v")
	output, err := command.CombinedOutput()
	if err == nil || !bytes.Contains(output, []byte("Node undefined:")) {
		t.Fatalf("mutant did not fail Node comparison: %v %s", err, output)
	}
	t.Logf("optional-as-always mutant caught by Node: %s", output)
}

// Includes syntax-negative patterns, which have no successful match and must
// return a parser error instead of a fabricated participation table.
func TestCaptureFactsTest262Patterns(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("testdata/test262.json")
	if err != nil {
		t.Fatal(err)
	}
	var all [][3]string
	if err = json.Unmarshal(data, &all); err != nil {
		t.Fatal(err)
	}
	var selected [][3]string
	for _, row := range all {
		if strings.Contains(row[2], "named-groups/") || strings.Contains(row[2], "match-indices/") {
			selected = append(selected, row)
		}
	}
	if len(selected) != 264 {
		t.Fatalf("pattern inventory changed: %d", len(selected))
	}
	input, err := json.Marshal(selected)
	if err != nil {
		t.Fatal(err)
	}
	command := exec.Command("node", "-e", `if(process.version !== 'v24.19.0')throw Error('requires Node 24.19.0');console.log(JSON.stringify(JSON.parse(require('fs').readFileSync(0,'utf8')).map(([p,f])=>{try{new RegExp(p,f);return true}catch(e){if(!(e instanceof SyntaxError))throw e;return false}})));`)
	command.Stdin = bytes.NewReader(input)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("Node: %v %s", err, output)
	}
	var valid []bool
	if err = json.Unmarshal(output, &valid); err != nil {
		t.Fatal(err)
	}
	if len(valid) != len(selected) {
		t.Fatal("incomplete Node response")
	}
	accepted, refused := 0, 0
	for i, row := range selected {
		_, err := CaptureFacts(row[0], row[1])
		if (err == nil) != valid[i] {
			t.Errorf("%s /%s/%s: Node validity=%v facts error=%v", row[2], row[0], row[1], valid[i], err)
		}
		if err == nil {
			accepted++
		} else {
			refused++
		}
	}
	t.Logf("test262 pattern sites: %d accepted, %d syntax refused", accepted, refused)
}

func TestCaptureFactsPrecision(t *testing.T) {
	t.Parallel()
	for _, c := range []struct {
		pattern string
		indices []bool
		names   map[string]bool
	}{
		{`(?<x>a)(b?)`, []bool{true, true, true}, map[string]bool{"x": true}},
		{`(?<x>a)?(b)`, []bool{true, false, true}, map[string]bool{"x": false}},
		{`(?<x>a)|(?<x>b)`, []bool{true, false, false}, map[string]bool{"x": true}},
		{`(?!(?<x>a))b`, []bool{true, false}, map[string]bool{"x": false}},
		{`(?<=(?<x>a))b`, []bool{true, true}, map[string]bool{"x": true}},
		{`((a)|(b))+`, []bool{true, true, false, false}, map[string]bool{}},
		{`(?<x>a)?\k<x>`, []bool{true, false}, map[string]bool{"x": false}},
	} {
		facts, err := CaptureFacts(c.pattern, "")
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(facts.Indices, c.indices) || !reflect.DeepEqual(facts.Names, c.names) {
			t.Errorf("/%s/: %+v", c.pattern, facts)
		}
	}
	if _, err := CaptureFacts("(", ""); err == nil {
		t.Fatal("invalid pattern accepted")
	}
	if _, err := CaptureFacts("a", "gg"); err == nil {
		t.Fatal("invalid flags accepted")
	}
}

// Compile the implementation and the named witnesses only. The compiler-source
// inventory uses TypeScript's parser and is unrelated to these Node comparisons.
func regexpMutantSources(t testing.TB, witnesses ...string) string {
	t.Helper()
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	var sources []string
	for _, file := range files {
		if !strings.HasSuffix(file, "_test.go") {
			sources = append(sources, "internal/regexp/"+file)
		}
	}
	for _, file := range witnesses {
		sources = append(sources, "internal/regexp/"+file)
	}
	return strings.Join(sources, " ")
}
