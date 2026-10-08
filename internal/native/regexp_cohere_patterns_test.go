package native

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	regex "github.com/system-inc/adamic/internal/regexp"
)

type cohereRegexInput struct {
	Text, Source, Kind string
	Units              []uint16
	Traces             []string
}
type cohereRegexPattern struct {
	PatternUnits     []uint16
	Pattern, Flags   string
	Sources, Engines []string
	Inputs           []cohereRegexInput
}
type cohereRegexCorpus struct {
	CohereCommit, Scope string
	Patterns            []cohereRegexPattern
}

func cohereRegexPatterns(t *testing.T) cohereRegexCorpus {
	t.Helper()
	path := os.Getenv("ADAMIC_COHERE_REGEX_CORPUS")
	if path == "" {
		path = "testdata/regexp_cohere_patterns.json.gz"
	}
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	var corpus cohereRegexCorpus
	if strings.HasSuffix(path, ".gz") {
		reader, err := gzip.NewReader(file)
		if err != nil {
			t.Fatal(err)
		}
		defer reader.Close()
		err = json.NewDecoder(reader).Decode(&corpus)
	} else {
		err = json.NewDecoder(file).Decode(&corpus)
	}
	if err != nil {
		t.Fatal(err)
	}
	if len(corpus.Patterns) == 0 || corpus.CohereCommit == "" {
		t.Fatal("empty or unpinned cohere corpus")
	}
	return corpus
}

// Candidate strings retain fixture provenance. They include identifiers,
// comments and literal values a rule consumes, plus complete source and lines;
// this is a conservative superset, not a claim of per-call runtime tracing.
func TestRegExpCoherePatterns(t *testing.T) {
	t.Parallel()
	corpus := cohereRegexPatterns(t)
	type row struct {
		Pattern, Flags string
		PatternUnits   []uint16
		Error          string
	}
	rows := make([]row, len(corpus.Patterns))
	for i, p := range corpus.Patterns {
		rows[i] = row{Pattern: p.Pattern, Flags: p.Flags, PatternUnits: p.PatternUnits}
	}
	data, err := json.Marshal(rows)
	if err != nil {
		t.Fatal(err)
	}
	command := exec.Command("node", "-e", `
if(process.version!=='v24.19.0')throw Error('Node 24.19.0 required');
const rows=JSON.parse(require('fs').readFileSync(0,'utf8'));
for(const r of rows){try{let pattern='';for(let i=0;i<r.PatternUnits.length;i+=8192)pattern+=String.fromCharCode(...r.PatternUnits.slice(i,i+8192));new RegExp(pattern,r.Flags)}catch(e){if(!(e instanceof SyntaxError))throw e;r.Error=e.message;continue;}}
process.stdout.write(JSON.stringify(rows.map(r=>r.Error??'')));`)
	command.Stdin = bytes.NewReader(data)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("Node syntax: %v %s", err, output)
	}
	var syntax []string
	if err := json.Unmarshal(output, &syntax); err != nil {
		t.Fatal(err)
	}
	var cases []regexCase
	compiled, refused, invalid, withoutInputs := 0, 0, 0, 0
	for i, p := range corpus.Patterns {
		program, err := regex.CompileUTF16(p.PatternUnits, p.Flags)
		if syntax[i] != "" {
			invalid++
			var syntaxError *regex.SyntaxError
			if !errors.As(err, &syntaxError) {
				t.Errorf("DISAGREEMENT syntax /%s/%s: Node %s", p.Pattern, p.Flags, syntax[i])
			}
			t.Logf("NODE_SYNTAX pattern=%d /%s/%s: %s", i, p.Pattern, p.Flags, syntax[i])
			continue
		}
		if err == nil {
			err = program.NativeCompatibility()
		}
		if err != nil {
			refused++
			t.Logf("REFUSED pattern=%d /%s/%s source=%s: %v", i, p.Pattern, p.Flags, p.Sources[0], err)
			continue
		}
		if _, encodeErr := program.NativeDeclarations(fmt.Sprintf("cohere_encode_%d", i)); encodeErr != nil {
			t.Fatalf("native encoding pattern %d: %v", i, encodeErr)
		}
		compiled++
		if len(p.Inputs) == 0 {
			withoutInputs++
			t.Logf("NO_INPUT pattern=%d /%s/%s source=%s", i, p.Pattern, p.Flags, p.Sources[0])
			continue
		}
		for _, input := range p.Inputs {
			cases = append(cases, regexCase{Pattern: p.Pattern, PatternUnits: p.PatternUnits, Flags: p.Flags, Input: input.Units})
		}
	}
	t.Logf("cohere totals: patterns=%d compiled=%d refused=%d Node-invalid=%d without-fixture-input=%d executions=%d pin=%s", len(corpus.Patterns), compiled, refused, invalid, withoutInputs, len(cases), corpus.CohereCommit)
	if refused != 0 {
		t.Fatalf("%d unexpected Node-valid refusals", refused)
	}
	// The existing native oracle compares test, exec, capture text, capture spans,
	// named groups and lastIndex in bounded, unbounded and forced-VM modes.
	const batch = 8000
	for start := 0; start < len(cases); start += batch {
		end := start + batch
		if end > len(cases) {
			end = len(cases)
		}
		chunk := cases[start:end]
		data, err := json.Marshal(chunk)
		if err != nil {
			t.Fatal(err)
		}
		command := exec.Command("node", "-e", `
const rows=JSON.parse(require('fs').readFileSync(0,'utf8'));
for(const c of rows){let pattern='';for(let i=0;i<c.patternUnits.length;i+=8192)pattern+=String.fromCharCode(...c.patternUnits.slice(i,i+8192));const r=new RegExp(pattern,c.flags.includes('d')?c.flags:c.flags+'d');let input='';for(let i=0;i<c.input.length;i+=8192)input+=String.fromCharCode(...c.input.slice(i,i+8192));r.lastIndex=c.lastIndex;const m=r.exec(input);c.expected={captures:m?Array.from(m.indices,x=>x??null):null,lastIndex:r.lastIndex,groups:m?.indices.groups?Object.fromEntries(Object.entries(m.indices.groups).map(([k,v])=>[k,v??null])):null};}process.stdout.write(JSON.stringify(rows));`)
		command.Stdin = bytes.NewReader(data)
		output, err := command.CombinedOutput()
		if err != nil {
			t.Fatalf("Node matching batch %d: %v %s", start, err, output)
		}
		if err := json.Unmarshal(output, &chunk); err != nil {
			t.Fatal(err)
		}
		t.Run(fmt.Sprintf("inputs_%d_%d", start, end), func(t *testing.T) { runRegexCases(t, chunk) })
	}
}

// Corpus input is owned by this test, not a measurement dump. The test consumes
// its full pattern list and retains source/subject provenance for diagnostics.
func TestRegExpCohereCorpusPin(t *testing.T) {
	t.Parallel()
	corpus := cohereRegexPatterns(t)
	if corpus.CohereCommit != "7945d102a6c18dd36adf9114a758ce646e8b2359" {
		t.Fatalf("unexpected pin %s", corpus.CohereCommit)
	}
	for i, p := range corpus.Patterns {
		if len(p.Sources) == 0 {
			t.Errorf("pattern %d lacks a source", i)
		}
		for _, input := range p.Inputs {
			if input.Source == "" || input.Kind == "" {
				t.Errorf("pattern %d lacks input provenance", i)
			}
		}
	}
}

// Source inventory is test input, not a record of Adamic runs. Revalidate the
// earlier pass's source-token rejection distinction before interpreting it as
// a regex-library defect, and check every inventoried source hash.
func TestRegExpCohereInventory(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("../regexp/testdata/cohere/inventory.json")
	if err != nil {
		t.Fatal(err)
	}
	var inventory struct {
		Revisions map[string]string
		Files     []struct{ File, SHA256 string }
		Patterns  []struct {
			File, Pattern, Flags, LiteralToken, SourceError string
			PatternUnits                                    []uint16
		}
		Dynamic []json.RawMessage
	}
	if err := json.Unmarshal(data, &inventory); err != nil {
		t.Fatal(err)
	}
	if inventory.Revisions["cohere"] != "7945d102a6c18dd36adf9114a758ce646e8b2359" || len(inventory.Patterns) != 969 || len(inventory.Dynamic) != 19 {
		t.Fatal("unexpected source inventory pin or size")
	}
	corpus := cohereRegexPatterns(t)
	known := map[string]bool{}
	for _, p := range corpus.Patterns {
		known[fmt.Sprint(p.PatternUnits)+"/"+p.Flags] = true
	}
	for _, p := range inventory.Patterns {
		if p.SourceError == "" && !known[fmt.Sprint(p.PatternUnits)+"/"+p.Flags] {
			t.Errorf("inventory pattern absent from execution corpus: %s /%s/%s", p.File, p.Pattern, p.Flags)
		}
	}
	for _, file := range inventory.Files {
		source, err := os.ReadFile(filepath.Join("../..", file.File))
		if err != nil {
			t.Fatal(err)
		}
		if fmt.Sprintf("%x", sha256.Sum256(source)) != file.SHA256 {
			t.Errorf("source changed since extraction: %s", file.File)
		}
	}
	query := make([]map[string]string, len(inventory.Patterns))
	for i, p := range inventory.Patterns {
		query[i] = map[string]string{"token": p.LiteralToken, "expected": p.SourceError}
	}
	wire, err := json.Marshal(query)
	if err != nil {
		t.Fatal(err)
	}
	command := exec.Command("node", "-e", `const vm=require('node:vm'),rows=JSON.parse(require('fs').readFileSync(0,'utf8'));for(const r of rows){if(!r.token)continue;let message='';try{new vm.Script('('+r.token+')')}catch(e){if(!(e instanceof SyntaxError))throw e;message=e.message}if(message!==r.expected)throw Error('source token expectation changed: '+r.token+' '+message)}`)
	command.Stdin = bytes.NewReader(wire)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("source-token validation: %v %s", err, output)
	}
}
