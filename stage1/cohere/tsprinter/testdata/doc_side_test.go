// Overlay in cohere's doc package; its existing independent spec generator is the input authority.
package doc

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"testing"
)

// Not parallel: writes docs.txt, answers.txt and docs.json in ADAMIC_TS_DOC_OUTPUT.
func TestAdamicDocuments(t *testing.T) {
	directory := os.Getenv("ADAMIC_TS_DOC_OUTPUT")
	if directory == "" {
		t.Fatal("missing output directory")
	}
	random := rand.New(rand.NewSource(20261001))
	cases := []printCase{}
	options := []Options{}
	for index := 0; index < 5000; index++ {
		generator := &specGenerator{random: random}
		opt := Options{PrintWidth: 8 + random.Intn(40), TabWidth: []int{2, 4}[random.Intn(2)], UseTabs: random.Intn(4) == 0}
		options = append(options, opt)
		cases = append(cases, printCase{Spec: generator.node(), Options: map[string]any{"printWidth": opt.PrintWidth, "tabWidth": opt.TabWidth, "useTabs": opt.UseTabs}})
	}
	specs, boundaries := boundaryCases()
	for index, shape := range specs {
		opt := boundaries[index]
		options = append(options, opt)
		cases = append(cases, printCase{Spec: shape, Options: map[string]any{"printWidth": opt.PrintWidth, "tabWidth": opt.TabWidth, "useTabs": opt.UseTabs}})
	}
	var protocol, answers strings.Builder
	escape := strings.NewReplacer("\\", "\\\\", "\n", "\\n", "\r", "\\r", "\t", "\\t")
	for index, item := range cases {
		opt := options[index]
		fmt.Fprintf(&protocol, "reset\t0x%x\t0x%x\t%d\n", opt.PrintWidth, opt.TabWidth, boolInt(opt.UseTabs))
		count := 0
		groups := map[int]int{}
		row := func(kind, text, key string, number int, flag bool, children []int) int {
			parts := []string{}
			for _, child := range children {
				parts = append(parts, fmt.Sprint(child))
			}
			fmt.Fprintf(&protocol, "%s\t%d\t%d\t%s\t%s\t%s\n", kind, number, boolInt(flag), key, escape.Replace(text), strings.Join(parts, ","))
			result := count
			count++
			return result
		}
		var flatten func(*spec) int
		flatten = func(node *spec) int {
			if node.Kind == "sharedGroup" {
				return groups[node.Ref]
			}
			if node.Kind == "hardline" || node.Kind == "literalline" {
				line := row(node.Kind+"WithoutBreakParent", "", "", 0, false, nil)
				parent := row("breakParent", "", "", 0, false, nil)
				return row("concat", "", "", 0, false, []int{line, parent})
			}
			children := []int{}
			for _, child := range node.Children {
				children = append(children, flatten(child))
			}
			result := row(node.Kind, node.Text, node.ID, node.Number, node.Flag, children)
			if node.Kind == "group" && node.Ref != 0 {
				groups[node.Ref] = result
			}
			return result
		}
		fmt.Fprintf(&protocol, "print\t%d\n", flatten(item.Spec))
		fmt.Fprintf(&answers, "ok\t%s\n", escape.Replace(Print(buildSpec(item.Spec), opt)))
	}
	for name, data := range map[string][]byte{"docs.txt": []byte(protocol.String()), "answers.txt": []byte(answers.String())} {
		if err := os.WriteFile(directory+"/"+name, data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	data, err := json.Marshal(cases)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(directory+"/docs.json", data, 0644); err != nil {
		t.Fatal(err)
	}
	t.Logf("%d docs", len(cases))
}
func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
