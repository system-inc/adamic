package main

import (
	"encoding/json"
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/system-inc/cohere/internal/lint/rules/tailwind"
	goast "go/ast"
	goparser "go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type node struct {
	Kind             string `json:"kind"`
	Text             string `json:"text"`
	Expression       int    `json:"expression"`
	Name             int    `json:"name"`
	Initializer      int    `json:"initializer"`
	ArgumentsPresent bool   `json:"argumentsPresent"`
	Arguments        []int  `json:"arguments"`
}
type values struct {
	Literals  []int `json:"literals"`
	Templates []int `json:"templates"`
}
type sample struct {
	Mode       string   `json:"mode"`
	Nodes      []node   `json:"nodes"`
	Queries    []int    `json:"queries"`
	Attributes []values `json:"attributes"`
	Callees    []values `json:"callees"`
	Variables  []values `json:"variables"`
	Collected  []values `json:"collected"`
	Under      []values `json:"under"`
	Patterns   [][]bool `json:"patterns"`
	Names      []string `json:"names"`
	Records    []string `json:"records"`
}

func must(e error) {
	if e != nil {
		panic(e)
	}
}
func main() {
	root, output, mode := os.Args[1], os.Args[2], os.Args[3]
	method := map[string]string{"read": "readClassValues", "callee": "calleeValues", "variable": "variableValues"}[mode]
	symbol := "github.com/system-inc/cohere/internal/lint/rules/tailwind.*ClassLiteralReader." + method
	data, err := os.ReadFile(filepath.Join(root, "stage1/cohere/lint/helpers/readiness.json"))
	must(err)
	var readiness struct {
		Remaining []struct {
			Rule    string   `json:"rule"`
			Helpers []string `json:"remaining_helpers"`
		} `json:"remaining"`
	}
	must(json.Unmarshal(data, &readiness))
	data, err = os.ReadFile(filepath.Join(root, "stage1/cohere/lint/inventory/inventory.json"))
	must(err)
	var inventory struct {
		Rules []struct {
			Name  string `json:"name"`
			Tests struct {
				Files []string `json:"files"`
			} `json:"tests"`
		} `json:"rules"`
	}
	must(json.Unmarshal(data, &inventory))
	consumers := map[string]bool{}
	for _, r := range readiness.Remaining {
		for _, h := range r.Helpers {
			if h == symbol {
				consumers[r.Rule] = true
			}
		}
	}
	if len(consumers) == 0 {
		panic("no consumers")
	}
	sources := map[string]bool{}
	counts := map[string]int{}
	for _, r := range inventory.Rules {
		if !consumers[r.Name] {
			continue
		}
		for _, path := range r.Tests.Files {
			file, err := goparser.ParseFile(token.NewFileSet(), filepath.Join(root, path), nil, 0)
			must(err)
			goast.Inspect(file, func(n goast.Node) bool {
				lit, ok := n.(*goast.BasicLit)
				if !ok || lit.Kind != token.STRING {
					return true
				}
				source, err := strconv.Unquote(lit.Value)
				must(err)
				if len(source) > 0 {
					sources[source] = true
					counts[r.Name]++
				}
				return true
			})
		}
		if counts[r.Name] == 0 {
			panic("no source strings for " + r.Name)
		}
	}
	if len(consumers) != 11 {
		panic("consumer count drift")
	}
	for _, s := range []string{
		"const x = <div className={'flex'} other={'unused'} class={['block', open ? 'p-2' : 'p-4']} />;",
		"mergeClassNames('flex', ['p-2', 'p-4'], open && 'hidden'); theme.mergeClassNames('unused'); (mergeClassNames)('unused'); createVariantClassNames('block');",
		"const buttonClassName = 'flex'; const buttonClassNames = ['p-2','p-4']; const other = 'unused'; const { className } = obj; let className;",
		"const x = mergeClassNames(\u0022flex\u0022, `p-2${open ? 'block' : 'hidden'}`, `flex${`block${open ? ' p-2 ' : ''}`}`);",
		"const ClassName = (('flex')) satisfies string; mergeClassNames(); const z = <div className />;",
		"const x = <div className={open ? `flex${' p-2 '}` : ('block')} />;",
	} {
		sources[s] = true
	}
	ordered := []string{}
	for s := range sources {
		ordered = append(ordered, s)
	}
	sort.Strings(ordered)
	corpus := []sample{}
	var expected strings.Builder
	configurations := []tailwind.ClassLiteralSettings{
		tailwind.DefaultClassLiteralSettings(),
		{AttributeNames: []string{}, CalleeNames: []string{}, VariablePatterns: []string{}},
		{AttributeNames: []string{"other", "class"}, CalleeNames: []string{"cn", "mergeClassNames", "createVariantClassNames"}, VariablePatterns: []string{"[", "^other$", "(?i)classname", "className$"}},
	}
	for _, source := range ordered {
		file := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/probe.tsx", Path: tspath.Path("/probe.tsx")}, source, core.ScriptKindTSX)
		actual := []*ast.Node{}
		indexes := map[*ast.Node]int{}
		var visit func(*ast.Node)
		visit = func(n *ast.Node) {
			if n == nil {
				return
			}
			if _, ok := indexes[n]; ok {
				return
			}
			indexes[n] = len(actual)
			actual = append(actual, n)
			n.ForEachChild(func(c *ast.Node) bool { visit(c); return false })
		}
		visit(file.AsNode())
		index := func(n *ast.Node) int {
			if n == nil {
				return -1
			}
			i, ok := indexes[n]
			if !ok {
				panic("named field absent")
			}
			return i
		}
		for _, config := range configurations {
			reader := tailwind.NewClassLiteralReader(config)
			row := sample{Mode: mode, Nodes: []node{}, Queries: []int{}, Attributes: []values{}, Callees: []values{}, Variables: []values{}, Collected: []values{}, Under: []values{}, Patterns: [][]bool{}, Names: config.CalleeNames, Records: []string{}}
			records := map[string]int{}
			convert := func(v tailwind.AdamicValues) values {
				out := values{[]int{}, []int{}}
				record := func(s string) int {
					i, ok := records[s]
					if !ok {
						i = len(row.Records)
						row.Records = append(row.Records, s)
						records[s] = i
					}
					return i
				}
				for _, s := range v.Literals {
					out.Literals = append(out.Literals, record(s))
				}
				for _, s := range v.Templates {
					out.Templates = append(out.Templates, record(s))
				}
				return out
			}
			for i, n := range actual {
				p := node{Kind: strings.TrimPrefix(n.Kind.String(), "Kind"), Expression: -1, Name: -1, Initializer: -1, Arguments: []int{}}
				if n.Kind == ast.KindIdentifier {
					p.Text = n.Text()
				}
				if n.Kind == ast.KindCallExpression {
					c := n.AsCallExpression()
					p.Expression = index(c.Expression)
					p.ArgumentsPresent = c.Arguments != nil
					if c.Arguments != nil {
						for _, a := range c.Arguments.Nodes {
							p.Arguments = append(p.Arguments, index(a))
						}
					}
				}
				if n.Kind == ast.KindVariableDeclaration {
					d := n.AsVariableDeclaration()
					p.Name = index(d.Name())
					p.Initializer = index(d.Initializer)
				}
				row.Nodes = append(row.Nodes, p)
				empty := values{[]int{}, []int{}}
				a, c, v, collect, under := empty, empty, empty, empty, empty
				if n.Kind == ast.KindJsxAttribute {
					// Go name.Text() panics on JSX namespace names. None of the supported consumer
					// reads rely on those; preserve their refusal rather than manufacture a result.
					name := n.AsJsxAttribute().Name()
					if name != nil && name.Kind != ast.KindIdentifier {
						panic("unsupported namespace attribute in consumer corpus")
					}
					a = convert(tailwind.AdamicSurfaceValues(reader, n, "attribute", index))
				}
				if n.Kind == ast.KindCallExpression {
					c = convert(tailwind.AdamicSurfaceValues(reader, n, "callee", index))
				}
				if n.Kind == ast.KindVariableDeclaration {
					v = convert(tailwind.AdamicSurfaceValues(reader, n, "variable", index))
				}
				collect = convert(tailwind.AdamicSurfaceValues(reader, n, "collect", index))
				under = convert(tailwind.AdamicSurfaceValues(reader, n, "under", index))
				row.Attributes = append(row.Attributes, a)
				row.Callees = append(row.Callees, c)
				row.Variables = append(row.Variables, v)
				row.Collected = append(row.Collected, collect)
				row.Under = append(row.Under, under)
				row.Patterns = append(row.Patterns, tailwind.AdamicPatternMatches(reader, p.Text))
				if mode == "callee" && n.Kind != ast.KindCallExpression {
					continue
				}
				if mode == "variable" && n.Kind != ast.KindVariableDeclaration {
					continue
				}
				result := convert(tailwind.AdamicSurfaceValues(reader, n, mode, index))
				row.Queries = append(row.Queries, i)
				fmt.Fprintf(&expected, "%s:%s\n", join(result.Literals), join(result.Templates))
			}

			if mode == "read" {
				row.Queries = append(row.Queries, -1)
				result := convert(tailwind.AdamicSurfaceValues(reader, nil, mode, index))
				fmt.Fprintf(&expected, "%s:%s\n", join(result.Literals), join(result.Templates))
			} else {
				if !tailwind.AdamicSurfaceRefuses(reader, nil, mode) || !tailwind.AdamicSurfaceRefuses(reader, file.AsNode(), mode) {
					panic("Go refusal contract drift")
				}
			}
			corpus = append(corpus, row)
		}
	}
	if mode != "read" {
		for _, probe := range []struct {
			name  string
			index int
		}{{"nil", -1}, {"kind", 0}} {
			row := corpus[0]
			row.Queries = []int{probe.index}
			data, err := json.Marshal([]sample{row})
			must(err)
			must(os.WriteFile(filepath.Join(output, probe.name+".json"), data, 0644))
		}
	}
	data, err = json.Marshal(corpus)
	must(err)
	must(os.WriteFile(filepath.Join(output, "cases.json"), data, 0644))
	must(os.WriteFile(filepath.Join(output, "want.txt"), []byte(expected.String()), 0644))
	data, err = json.MarshalIndent(counts, "", "  ")
	must(err)
	must(os.WriteFile(filepath.Join(output, "coverage.json"), data, 0644))
	fmt.Printf("%d consumers, %d distinct strings and controls, %d parser/settings cases, %d Go verdicts\n", len(consumers), len(ordered), len(corpus), strings.Count(expected.String(), "\n"))
}
func join(a []int) string {
	s := []string{}
	for _, i := range a {
		s = append(s, strconv.Itoa(i))
	}
	return strings.Join(s, ",")
}
