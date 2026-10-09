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

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	tsparser "github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
)

type captureLiteral struct {
	File    string               `json:"file"`
	Line    int                  `json:"line"`
	Pattern string               `json:"pattern"`
	Flags   string               `json:"flags"`
	Facts   CaptureParticipation `json:"facts"`
}

// This fixed checker contract is read by the test, not a measurement dump.
// Parse every pinned compiler source again so omissions and new literals fail.
// Not parallel: writes testdata/capture-tsc.json when ADAMIC_UPDATE_CAPTURE_TABLE=1.
func TestCaptureFactsCompilerTable(t *testing.T) {
	data, err := os.ReadFile("testdata/capture-tsc.json")
	if err != nil {
		t.Fatal(err)
	}
	var table []captureLiteral
	if err = json.Unmarshal(data, &table); err != nil {
		t.Fatal(err)
	}
	root := "../../cohere/TypeScript/tsc/testdata/fixtures/compiler"
	var actual []captureLiteral
	err = filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".ts") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		text := string(data)
		absolute, err := filepath.Abs(path)
		if err != nil {
			return err
		}
		file := tsparser.ParseSourceFile(ast.SourceFileParseOptions{FileName: tspath.RootedFilePathFromAbsolute(absolute)}, text, core.ScriptKindTS)
		var visit func(*ast.Node) bool
		visit = func(n *ast.Node) bool {
			if n.Kind == ast.KindRegularExpressionLiteral {
				literal := n.Text()
				end := strings.LastIndex(literal, "/")
				facts, err := CaptureFacts(literal[1:end], literal[end+1:])
				if err != nil {
					t.Errorf("%s: %v", path, err)
				}
				pos := strings.Index(text[n.Pos():], literal) + n.Pos()
				actual = append(actual, captureLiteral{File: strings.TrimPrefix(filepath.ToSlash(path), "../../"), Line: strings.Count(text[:pos], "\n") + 1, Pattern: literal[1:end], Flags: literal[end+1:], Facts: facts})
			}
			n.ForEachChild(visit)
			return false
		}
		visit(file.AsNode())
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if os.Getenv("ADAMIC_UPDATE_CAPTURE_TABLE") == "1" {
		data, err := json.MarshalIndent(actual, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile("testdata/capture-tsc.json", append(data, '\n'), 0644); err != nil {
			t.Fatal(err)
		}
		return
	}
	if !reflect.DeepEqual(actual, table) {
		t.Fatal("compiler literal facts changed; regenerate and review capture-tsc.json")
	}
	if len(actual) != 88 {
		t.Fatalf("literal coverage: %d", len(actual))
	}
	t.Logf("checked %d compiler literal sites", len(actual))
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

func TestCaptureFactsOptionalMutant(t *testing.T) {
	t.Parallel()
	source, err := os.ReadFile("capture_facts.go")
	if err != nil {
		t.Fatal(err)
	}
	old := "if n.Min.Sign() > 0 {"
	if strings.Count(string(source), old) != 1 {
		t.Fatal("mutation site changed")
	}
	dir := t.TempDir()
	replacement := filepath.Join(dir, "capture_facts.go")
	if err = os.WriteFile(replacement, []byte(strings.Replace(string(source), old, "if n.Min.Sign() >= 0 {", 1)), 0600); err != nil {
		t.Fatal(err)
	}
	original, err := filepath.Abs("capture_facts.go")
	if err != nil {
		t.Fatal(err)
	}
	overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{original: replacement}})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "overlay.json")
	if err = os.WriteFile(path, overlay, 0600); err != nil {
		t.Fatal(err)
	}
	command := exec.Command("go", "test", "-overlay", path, ".", "-run", "^TestCaptureFactsNodeBranches$", "-count=1", "-v")
	output, err := command.CombinedOutput()
	if err == nil || bytes.Contains(output, []byte("build failed")) || !bytes.Contains(output, []byte("Node undefined:")) {
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
