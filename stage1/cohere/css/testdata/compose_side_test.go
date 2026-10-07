// Overlay into cohere's composition package; compare byte offsets, not a
// normalized UTF-16 approximation, and retain dumpGlueValue's range assertion.
package css

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/format/css/postcss"
)

func TestAdamicCompositionCases(t *testing.T) {
	var request struct{ Cases, Answers string }
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
		root, err := parseWithParserName(text, parser)
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
			answer = dumpGlueValue(root, func(index int) any { return index })
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
	t.Logf("composed trees: %d inputs, %d parsed", count, parsed)
}
