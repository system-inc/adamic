// Overlaid into cohere/internal/format/graphql, without modifying the submodule.
package graphql

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/format/estree"
	"github.com/system-inc/cohere/internal/format/formatoptions"
)

func TestAdamicPrinter(t *testing.T) {
	path := os.Getenv("ADAMIC_PRINTER_REQUEST")
	if path == "" {
		// census: required-input ADAMIC_PRINTER_REQUEST: stage1/cohere/graphql/printer/printer_test.go printerCases writes the Sources/Cases/Answers/Mode/Coverage JSON request and exports its path to the pinned cohere Go overlay.
		t.Skip("stage 1 overlay")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var request struct {
		Sources  []string
		Answers  string
		Cases    string
		Mode     string
		Coverage string
	}
	if err = json.Unmarshal(data, &request); err != nil {
		t.Fatal(err)
	}
	// Exhaust the width boundary, argument counts, nesting and unicode widths.
	for width := 70; width <= 140; width++ {
		for _, value := range []string{"x", "é", "😀", "漢", "e\u0301", "\u2028", "\u0085"} {
			request.Sources = append(request.Sources, "query{f(a:\""+strings.Repeat(value, width)+"\",b:[1 2 3],c:{x:1 y:2})}")
		}
	}
	for count := 1; count <= 40; count++ {
		var args []string
		for i := 0; i < count; i++ {
			args = append(args, fmt.Sprintf(`argument%d: "value%d"`, i, i))
		}
		request.Sources = append(request.Sources, "query{f("+strings.Join(args, ",")+")}")
	}
	for _, item := range formatCases {
		request.Sources = append(request.Sources, item.source)
	}
	generator := &adamicGenerator{random: rand.New(rand.NewSource(20261006))}
	for count := 0; count < 3000; count++ {
		request.Sources = append(request.Sources, generator.document())
	}
	request.Sources = append(request.Sources, adamicDeepTexts()...)
	coverage := map[string]int{}
	var walk func(*estree.Node)
	walk = func(node *estree.Node) {
		if node == nil {
			return
		}
		coverage[node.Type()]++
		for _, key := range getVisitorKeys(node) {
			switch value := node.Get(key).(type) {
			case *estree.Node:
				walk(value)
			case []*estree.Node:
				for _, child := range value {
					walk(child)
				}
			}
		}
	}
	var answers, cases strings.Builder
	options := formatoptions.Default()
	switch request.Mode {
	case "narrow":
		options.PrintWidth = 80
		options.TabWidth = 2
	case "tabs":
		options.UseTabs = true
	case "tight":
		options.BracketSpacing = false
	}
	for _, source := range request.Sources {
		normalized := strings.TrimPrefix(source, "\ufeff")
		normalized = strings.ReplaceAll(strings.ReplaceAll(normalized, "\r\n", "\n"), "\r", "\n")
		document, _, parseErr := Parse(normalized)
		if parseErr == nil {
			walk(document)
		}
		formatted, err := Format(normalized, options)
		if err == nil && strings.HasPrefix(source, "\ufeff") {
			formatted = "\ufeff" + formatted
		}
		if err != nil {
			fmt.Fprintln(&answers, "error\t"+escapePrinter(err.Error()))
		} else {
			fmt.Fprintln(&answers, "ok\t"+escapePrinter(formatted))
		}
		fmt.Fprintln(&cases, ">"+escapePrinter(source))
	}
	for kind := range visitorKeys {
		if coverage[kind] == 0 {
			t.Errorf("corpus never reaches %s", kind)
		}
	}
	encoded, err := json.MarshalIndent(coverage, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if request.Coverage != "" {
		if err = os.WriteFile(request.Coverage, append(encoded, '\n'), 0644); err != nil {
			t.Fatal(err)
		}
	}
	t.Logf("%d upstream format fixtures, %d AST kinds covered", len(formatCases), len(coverage))
	if err = os.WriteFile(request.Answers, []byte(answers.String()), 0644); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(request.Cases, []byte(cases.String()), 0644); err != nil {
		t.Fatal(err)
	}
}
func escapePrinter(text string) string {
	return strings.NewReplacer("\\", "\\\\", "\n", "\\n", "\r", "\\r", "\t", "\\t").Replace(text)
}
