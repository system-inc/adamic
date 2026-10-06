package regexp

import (
	"compress/gzip"
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

type v8CostRow struct {
	Pattern    string   `json:"pattern"`
	Flags      string   `json:"flags"`
	Reason     string   `json:"reason"`
	Sources    []string `json:"sources"`
	Executions int      `json:"executions"`
	Agree      int      `json:"agree"`
	Disagree   int      `json:"disagree"`
}

func TestV8RefusalCost(t *testing.T) {
	data, err := os.Open("testdata/matches.json.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer data.Close()
	reader, err := gzip.NewReader(data)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	var cases []executionCase
	if err := json.NewDecoder(reader).Decode(&cases); err != nil {
		t.Fatal(err)
	}
	programs := map[string]*Program{}
	rows := map[string]*v8CostRow{}
	sourceSets := map[string]map[string]bool{}
	affectedFiles := map[string]bool{}
	count, agree, disagree := 0, 0, 0
	for _, c := range cases {
		key := c.Pattern + "/" + c.Flags
		p := programs[key]
		if p == nil {
			p, err = Compile(c.Pattern, c.Flags)
			if err != nil {
				t.Fatal(err)
			}
			programs[key] = p
		}
		if p.NativeCompatibility() == nil {
			continue
		}
		if rows[key] == nil {
			rows[key] = &v8CostRow{Pattern: c.Pattern, Flags: c.Flags, Reason: p.NativeCompatibility().Error()}
			sourceSets[key] = map[string]bool{}
		}
		row := rows[key]
		row.Executions++
		count++
		affectedFiles[c.Source] = true
		if !sourceSets[key][c.Source] {
			row.Sources = append(row.Sources, c.Source)
			sourceSets[key][c.Source] = true
		}
		r := p.New()
		r.LastIndex = c.LastIndex
		r.StepLimit = 10000000
		got, err := matcherResult(r, c.Input)
		if err != nil {
			t.Fatal(err)
		}
		if reflect.DeepEqual(got, c.Expected) {
			row.Agree++
			agree++
		} else {
			row.Disagree++
			disagree++
		}
	}
	parserData, err := os.ReadFile("testdata/test262.json")
	if err != nil {
		t.Fatal(err)
	}
	var patterns [][]string
	if err := json.Unmarshal(parserData, &patterns); err != nil {
		t.Fatal(err)
	}
	parserRefusals := []v8CostRow{}
	parserFiles := map[string]bool{}
	for _, row := range patterns {
		p, err := Compile(row[0], row[1])
		if err != nil {
			continue
		}
		if p.NativeCompatibility() != nil {
			parserRefusals = append(parserRefusals, v8CostRow{Pattern: row[0], Flags: row[1], Reason: p.NativeCompatibility().Error(), Sources: []string{row[2]}})
			parserFiles[row[2]] = true
		}
	}
	t.Logf("execution corpus: %d rows, %d pattern/flags; refused %d rows in %d patterns and %d attributed files; %d agree with Node, %d disagree", len(cases), len(programs), count, len(rows), len(affectedFiles), agree, disagree)
	t.Logf("parser corpus: %d rows, %d refused pattern/flags in %d test262 files", len(patterns), len(parserRefusals), len(parserFiles))
	if path := os.Getenv("ADAMIC_V8_COST_REPORT"); path != "" {
		output := struct {
			ExecutionRows          int                   `json:"executionRows"`
			ExecutionPatterns      int                   `json:"executionPatterns"`
			RefusedExecutions      int                   `json:"refusedExecutions"`
			Agree                  int                   `json:"agree"`
			Disagree               int                   `json:"disagree"`
			AffectedExecutionFiles int                   `json:"affectedExecutionFiles"`
			ParserRows             int                   `json:"parserRows"`
			AffectedParserFiles    int                   `json:"affectedParserFiles"`
			Executions             map[string]*v8CostRow `json:"executions"`
			Patterns               []v8CostRow           `json:"patterns"`
		}{len(cases), len(programs), count, agree, disagree, len(affectedFiles), len(patterns), len(parserFiles), rows, parserRefusals}
		text, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, append(text, '\n'), 0644); err != nil {
			t.Fatal(err)
		}
	}
}
