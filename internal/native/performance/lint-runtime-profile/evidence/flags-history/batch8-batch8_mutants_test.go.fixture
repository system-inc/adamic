package lint

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func batch8Mutants(t *testing.T, oracle, directory, originalBinary string) {
	t.Helper()
	sources := []struct{ name, rule, source, from, to string }{
		{"no_octal_escape", "no-octal-escape", `const value = '\251';`, "if(width !== 0)", "if(width < 0)"},
		{"no_unexpected_multiline", "no-unexpected-multiline", "const x = f\n(a);", "context.line(start) !== context.line(context.node(expression).end - 1)", "context.line(start) === context.line(context.node(expression).end - 1)"},
		{"no_unused_private_class_members", "no-unused-private-class-members", "class C { #x=0; m(){this.#x=1;} }", "if(!read)", "if(read)"},
		{"no_useless_constructor", "no-useless-constructor", "class C { constructor(){} }", "if(body < 0)", "if(body < 0 || body >= 0)"},
		{"prefer_template", "prefer-template", "const x = 'a' + value;", "if(!contains(context, top, false))", "if(contains(context, top, false))"},
		{"forward_ref_uses_ref", "react/forward-ref-uses-ref", "const C = forwardRef(props => props);", "if(counted !== 1)", "if(counted === 1)"},
		{"jsx_no_comment_textnodes", "react/jsx-no-comment-textnodes", "const view = <div>// comment</div>;", "if(atLineStart && character === '/'", "if(!atLineStart && character === '/'"},
		{"no_find_dom_node", "react/no-find-dom-node", "ReactDOM['findDOMNode'](node);", "context.node(name).text === 'findDOMNode'", "context.node(name).text !== 'findDOMNode'"},
		{"no_is_mounted", "react/no-is-mounted", "class C { m(){this.isMounted();} }", "context.node(name).text !== 'isMounted'", "context.node(name).text === 'isMounted'"},
		{"no_redundant_should_component_update", "react/no-redundant-should-component-update", "class C extends PureComponent {shouldComponentUpdate(){}}", "if(!pure)", "if(pure)"},
		{"forward_ref_uses_ref", "react/forward-ref-uses-ref", "const C = forwardRef(props => props);", "'forwardRefAddRefParameter'", "'wrongSecondSuggestion'"},
		{"prefer_template", "prefer-template", "const x = 'a' + value;", "new Edit(tokenStart(context, top), context.node(top).end, replacement)", "new Edit(tokenStart(context, top), context.node(top).end, replacement + ' ')"},
	}
	var controls []string
	for i, change := range sources {
		path := filepath.Join(t.TempDir(), fmt.Sprintf("control-%d.ts", i))
		if change.rule == "react/jsx-no-comment-textnodes" {
			path += "x"
		}
		if err := os.WriteFile(path, []byte(change.source), 0644); err != nil {
			t.Fatal(err)
		}
		row := path + "\t" + change.rule
		if change.rule == "react/jsx-no-comment-textnodes" {
			row += "\t" + batch8Spans(t, oracle, path)
		}
		controls = append(controls, row)
	}
	batch8Compare(t, oracle, originalBinary, directory, controls)
	slots := make(chan struct{}, 4)
	for _, change := range sources {
		t.Run(change.name, func(t *testing.T) {
			t.Parallel()
			t.Logf("mutant %s -> %s", change.from, change.to)
			slots <- struct{}{}
			defer func() { <-slots }()
			scratch := t.TempDir()
			entries, err := os.ReadDir(directory)
			if err != nil {
				t.Fatal(err)
			}
			typescript, _ := filepath.Abs("../../typescript")
			for _, entry := range entries {
				if !strings.HasSuffix(entry.Name(), ".ts") {
					continue
				}
				data, err := os.ReadFile(filepath.Join(directory, entry.Name()))
				if err != nil {
					t.Fatal(err)
				}
				source := string(data)
				if entry.Name() == change.name+".ts" {
					if strings.Count(source, change.from) != 1 {
						t.Fatalf("mutant anchor %q count", change.from)
					}
					source = strings.Replace(source, change.from, change.to, 1)
				}
				source = strings.ReplaceAll(source, "../../../typescript", typescript)
				if err := os.WriteFile(filepath.Join(scratch, entry.Name()), []byte(source), 0644); err != nil {
					t.Fatal(err)
				}
			}
			path := filepath.Join(t.TempDir(), "control.ts")
			if change.rule == "react/jsx-no-comment-textnodes" {
				path += "x"
			}
			if err := os.WriteFile(path, []byte(change.source), 0644); err != nil {
				t.Fatal(err)
			}
			row := path + "\t" + change.rule
			if change.rule == "react/jsx-no-comment-textnodes" {
				row += "\t" + batch8Spans(t, oracle, path)
			}
			input := manifest(t, []string{row})
			want := execute(t, "", oracle, "--manifest", input).output
			if !bytes.Contains(want, []byte("\nrange ")) {
				t.Fatal("mutant has no positive Go control")
			}
			binary := batch8Build(t, scratch, true)
			for _, side := range []struct {
				name string
				run  execution
			}{{"Node", batch8Node(t, scratch, input, false)}, {"native", execute(t, "", binary, "--manifest", input)}} {
				if bytes.Equal(side.run.output, want) {
					t.Fatalf("mutant survived on %s", side.name)
				}
				t.Logf("compiled mutant caught on %s: %s", side.name, difference(side.run.output, want))
			}
		})
	}
}
func batch8Throughput(t *testing.T, oracle, directory string) {
	t.Helper()
	var rows []string
	all := batch8Sources(t)
	for _, row := range all {
		if strings.Contains(row, "/src/compiler/") {
			rows = append(rows, row)
		}
	}
	input := manifest(t, rows)
	binary := batch8Build(t, directory, false)
	best := map[string]time.Duration{}
	var answer []byte
	for round := 0; round < 5; round++ {
		for _, name := range []string{"Go", "native", "Node"} {
			var result execution
			switch name {
			case "Go":
				result = execute(t, "", oracle, "--manifest", input, "--count")
			case "native":
				result = execute(t, "", binary, "--manifest", input, "--count")
			case "Node":
				result = batch8Node(t, directory, input, true)
			}
			if answer == nil {
				answer = result.output
			}
			if !bytes.Equal(answer, result.output) {
				t.Fatalf("count differs on %s", name)
			}
			if best[name] == 0 || result.duration < best[name] {
				best[name] = result.duration
			}
			t.Logf("round %d %s %s findings=%s", round+1, name, result.duration, strings.TrimSpace(string(answer)))
		}
	}
	var count int
	if _, err := fmt.Sscan(string(answer), &count); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"Go", "native", "Node"} {
		t.Logf("best of 5 %s: %.6fs, %.2f findings/s (%d files, %d findings)", name, best[name].Seconds(), float64(count)/best[name].Seconds(), len(rows), count)
	}
}
