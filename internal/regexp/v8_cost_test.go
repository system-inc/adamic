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

// Not parallel: writes the process-wide ADAMIC_V8_COST_REPORT path outside test.TempDir.
func TestV8RefusalCost(test *testing.T) {
	data, openError := os.Open("testdata/matches.json.gz")
	if openError != nil {
		test.Fatal(openError)
	}
	defer data.Close()
	reader, gzipError := gzip.NewReader(data)
	if gzipError != nil {
		test.Fatal(gzipError)
	}
	defer reader.Close()
	var cases []executionCase
	if decodeError := json.NewDecoder(reader).Decode(&cases); decodeError != nil {
		test.Fatal(decodeError)
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
			var compileError error
			p, compileError = Compile(c.Pattern, c.Flags)
			if compileError != nil {
				test.Fatal(compileError)
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
		got, matcherResultError := matcherResult(r, c.Input)
		if matcherResultError != nil {
			test.Fatal(matcherResultError)
		}
		if reflect.DeepEqual(got, c.Expected) {
			row.Agree++
			agree++
		} else {
			row.Disagree++
			disagree++
		}
	}
	parserData, readError := os.ReadFile("testdata/test262.json")
	if readError != nil {
		test.Fatal(readError)
	}
	var patterns [][]string
	if decodeError := json.Unmarshal(parserData, &patterns); decodeError != nil {
		test.Fatal(decodeError)
	}
	parserRefusals := []v8CostRow{}
	parserFiles := map[string]bool{}
	for _, row := range patterns {
		p, compileError := Compile(row[0], row[1])
		if compileError != nil {
			continue
		}
		if p.NativeCompatibility() != nil {
			parserRefusals = append(parserRefusals, v8CostRow{
				Pattern: row[0],
				Flags:   row[1],
				Reason:  p.NativeCompatibility().Error(),
				Sources: []string{row[2]},
			})
			parserFiles[row[2]] = true
		}
	}
	test.Logf(
		"execution corpus: %d rows, %d pattern/flags; refused %d rows in %d patterns and %d attributed files; %d agree with Node, %d disagree",
		len(cases),
		len(programs),
		count,
		len(rows),
		len(affectedFiles),
		agree,
		disagree,
	)
	test.Logf(
		"parser corpus: %d rows, %d refused pattern/flags in %d test262 files",
		len(patterns),
		len(parserRefusals),
		len(parserFiles),
	)
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
		}{
			len(cases),
			len(programs),
			count,
			agree,
			disagree,
			len(affectedFiles),
			len(patterns),
			len(parserFiles),
			rows,
			parserRefusals,
		}
		text, encodeError := json.MarshalIndent(output, "", "  ")
		if encodeError != nil {
			test.Fatal(encodeError)
		}
		if writeError := os.WriteFile(path, append(text, '\n'), 0644); writeError != nil {
			test.Fatal(writeError)
		}
	}
}
