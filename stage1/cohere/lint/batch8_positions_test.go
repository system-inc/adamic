package lint

import (
	"os"
	"path/filepath"
	"testing"
)

// Not parallel: release observations are taken after the primary suite's timed builds.
func TestBatch8PositionsAndRelease(t *testing.T) {
	directory := batch8Snapshot(t)
	oracle := batch8Oracle(t)
	cases := []struct{ rule, source string }{
		{"no-octal-escape", "const é='😀'; const x='\\1234\\77';"},
		{"no-unexpected-multiline", "const é='😀';\r\nconst x=f\u2028(args); foo\r/bar/g; tag<T>\n`x`;"},
		{"no-unused-private-class-members", "class É { #x='😀'; #y=0; m(){this.#x='β';this.#y++;} }"},
		{"no-useless-constructor", "class É { field='😀'\nconstructor(){}\n[0](){} }"},
		{"prefer-template", "const é='😀';const x='β` ${' + value;\r\nfoo()\n'a' + bar;"},
		{"react/forward-ref-uses-ref", "const é='😀';const C=forwardRef(( /* β */ props: P, )=>props);forwardRef(function(props){return props});"},
		{"react/jsx-no-comment-textnodes", "const é='😀';const C=<div>\n // β 😀\n text // clean\n /* next */</div>;"},
		{"react/no-find-dom-node", "const é='😀';((ReactDOM))['findDOMNode'](é);ReactDom[`findDOMNode`](é);"},
		{"react/no-is-mounted", "const é='😀';const c={m(){this['isMounted']();},field:()=>this.isMounted()};"},
		{"react/no-redundant-should-component-update", "const é='😀';const C=class extends (React.PureComponent){#shouldComponentUpdate(){}};"},
	}
	cases = append(cases,
		struct{ rule, source string }{"no-unused-private-class-members", "class C{#x; async m(){for await(this.#x of values){}}}"},
		struct{ rule, source string }{"no-useless-constructor", "class C extends Base{constructor(p: A<B<C>> = make()){super(p)}}"},
		struct{ rule, source string }{"no-useless-constructor", "class C{field: <T=string>()=>T\nconstructor(){}\n[0](){}}"},
		struct{ rule, source string }{"no-unexpected-multiline", "foo\n/bar/g\\u0061; foo\n/bar/gא;"},
	)
	var rows []string
	for i, item := range cases {
		path := filepath.Join(t.TempDir(), "unicode.ts")
		if i == 6 {
			path += "x"
		}
		if err := os.WriteFile(path, []byte(item.source), 0644); err != nil {
			t.Fatal(err)
		}
		row := path + "\t" + item.rule
		rows = append(rows, row)
	}
	sanitized := batch8Build(t, directory, true)
	batch8Compare(t, oracle, sanitized, directory, rows)
	release := batch8Build(t, directory, false)
	batch8Compare(t, oracle, release, directory, rows)
	t.Run("source_corpus", func(t *testing.T) { batch8Compare(t, oracle, release, directory, batch8Sources(t)) })
}
