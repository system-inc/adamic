// Overlay into cohere's composition package; compare byte offsets, not a
// normalized UTF-16 approximation, and retain dumpGlueValue's range assertion.
package css

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/system-inc/cohere/internal/format/formatoptions"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
	"unicode/utf16"

	"github.com/system-inc/cohere/internal/format/css/postcss"
)

// Not parallel: writes the fixed output paths supplied by the environment request.
func TestAdamicPrinterCases(t *testing.T) {
	var request struct{ Cases, Answers, Mode string }
	path := os.Getenv("ADAMIC_PORT_REQUEST")
	if path == "" {
		t.Skip("run by Adamic CSS slice")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &request); err != nil {
		t.Fatal(err)
	}
	data, err = os.ReadFile(request.Cases)
	if err != nil {
		t.Fatal(err)
	}
	var answers strings.Builder
	count, parsed := 0, 0
	started := time.Now()
	for _, line := range strings.Split(string(data), "\n") {
		if line == "" {
			continue
		}
		text, err := strconv.Unquote(`"` + strings.ReplaceAll(line[2:], `"`, `\"`) + `"`)
		if err != nil {
			t.Fatal(err)
		}
		parser := "css"
		if line[1] == 'S' {
			parser = "scss"
		}
		options := formatoptions.PrettierDefaults()
		if request.Mode == "narrow" {
			options.PrintWidth = 24
			options.TabWidth = 4
			options.UseTabs = true
			options.SingleQuote = true
			options.TrailingComma = "none"
		}
		formatted, err := formatWithParser(text, options, parser)
		fmt.Fprintf(&answers, "case %d\n", count)
		count++
		var answer any
		if err != nil {
			answers.WriteString("error ")
			var syntax *postcss.CssSyntaxError
			if errors.As(err, &syntax) {
				fields := map[string]any{"name": syntax.Name(), "reason": syntax.Reason, "line": syntax.Line, "column": syntax.Column, "offset": syntax.Offset}
				if syntax.HasEnd {
					fields["endLine"] = syntax.EndLine
					fields["endColumn"] = syntax.EndColumn
					fields["endOffset"] = syntax.EndOffset
				}
				answer = fields
			} else {
				answer = err.Error()
			}
		} else {
			answer = formatted
			parsed++
		}
		encoder := json.NewEncoder(&answers)
		encoder.SetEscapeHTML(false)
		if err := encoder.Encode(answer); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(request.Answers, []byte(answers.String()), 0644); err != nil {
		t.Fatal(err)
	}
	t.Logf("printed stylesheets: %d inputs, %d formatted, elapsed %s", count, parsed, time.Since(started))
}

// Not parallel: the formatter's throughput is measured without a competing oracle.
func TestAdamicPrinterThroughput(t *testing.T) {
	var request struct {
		Cases, Repeat string
		Rounds        int
	}
	data, err := os.ReadFile(os.Getenv("ADAMIC_PORT_REQUEST"))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &request); err != nil {
		t.Fatal(err)
	}
	data, err = os.ReadFile(request.Cases)
	if err != nil {
		t.Fatal(err)
	}
	var texts, parsers []string
	for _, line := range strings.Split(string(data), "\n") {
		if line == "" {
			continue
		}
		text, err := strconv.Unquote(`"` + strings.ReplaceAll(line[2:], `"`, `\"`) + `"`)
		if err != nil {
			t.Fatal(err)
		}
		parser := "css"
		if line[1] == 'S' {
			parser = "scss"
		}
		texts = append(texts, text)
		parsers = append(parsers, parser)
	}
	options := formatoptions.PrettierDefaults()
	repetitions := 10
	if request.Repeat == "once" {
		repetitions = 1
	}
	rounds := request.Rounds
	if rounds == 0 {
		rounds = 3
	}
	for round := 0; round < rounds; round++ {
		count, units := 0, 0
		started := time.Now()
		for repeat := 0; repeat < repetitions; repeat++ {
			for index, text := range texts {
				formatted, err := formatWithParser(text, options, parsers[index])
				if err != nil {
					t.Fatal(err)
				}
				count++
				units += len(utf16.Encode([]rune(formatted)))
			}
		}
		elapsed := time.Since(started)
		t.Logf("Go round %d: %d stylesheets in %s, %.0f stylesheets/s; %d units", round+1, count, elapsed, float64(count)/elapsed.Seconds(), units)
	}
}
