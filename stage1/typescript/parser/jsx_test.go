package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func jsxManifest(t *testing.T) (string, int) {
	t.Helper()
	cases := []string{
		`const x = <div/>;`, `const x = <div></div>;`, `const x = <></>;`,
		`const x = <><a/><b>text</b></>;`,
		`const x = <Foo.Bar baz data-value="a\nb" xml:lang='en' {...props}/>;`,
		`const x = <ns:tag ns:attr="&amp;">&lt; héllo 😀</ns:tag>;`,
		`const x = <this.Component flag={true} value={a, b}>{...items}{/* comment */}{value}</this.Component>;`,
		`const x = <Foo<T, U> ref={ref}><Foo.Bar<T[]>/></Foo>;`,
		`const x = <Foo<T,>><Bar<A<B<C>>, U,>/></Foo>;`,
		`const x = <div attr=<><i/></> other=<span/>/>;`,
		"const x = <div a=\" first\r\n second\\u1234 &quot;\">\n  \t\r\n</div>;",
		`const x = <div>{/a/.test(x)}{` + "`a${x}b`" + `}{x < y ? <a/> : <b/>}</div>;`,
		`const f = <T,>(x: T) => <span>{x}</span>;`,
		`const f = <T extends {}>(x: T) => <span>{x}</span>;`,
		`const f = <const T, U = string>(x: T) => <span>{x}</span>;`,
		`const f = async <T,>(x: T) => <span>{await x}</span>;`,
		`const x = <T>(x) =&gt; x</T>;`,
		`const x = (<div/>).props; const y = <div/> ? 1 : 2;`,
		`class C { m() { return <this attr={this.x}/>; } }`,
		`const x = <#tag/>;`,
		`const x = <ns:#tag/>;`,
		`const x = <Foo #attr ns:#attr/>;`,
	}
	names := []string{"div", "Foo.Bar", "this.Widget", "ns:item", "custom-element"}
	attributes := []string{"", ` bool`, ` data-x="\123"`, ` ns:key={x < y}`, ` {...props} a={value}`, ` a=<Child/>`}
	children := []string{"", "text 😀", "\n  \n", "{value}", "{/* note */}<Child/>"}
	for _, name := range names {
		for _, attr := range attributes {
			for _, child := range children {
				cases = append(cases, fmt.Sprintf("const x = <%s%s>%s</%s>;", name, attr, child, name))
			}
		}
	}
	directory := t.TempDir()
	var rows []string
	for i, source := range cases {
		for j, variant := range []string{source, "/* é 😀 */\r\n" + strings.ReplaceAll(source, "\n", "\r\n")} {
			path := filepath.Join(directory, fmt.Sprintf("jsx-%d-%d.tsx", i, j))
			if err := os.WriteFile(path, []byte(variant), 0644); err != nil {
				t.Fatal(err)
			}
			rows = append(rows, path)
		}
	}
	for _, extension := range []string{".jsx", ".js", ".mjs", ".cjs"} {
		path := filepath.Join(directory, "javascript"+extension)
		if err := os.WriteFile(path, []byte(`const x = <div data-x='raw\n'>{value}</div>;`), 0644); err != nil {
			t.Fatal(err)
		}
		rows = append(rows, path)
	}
	path := filepath.Join(directory, "manifest")
	if err := os.WriteFile(path, []byte(strings.Join(rows, "\n")+"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	return path, len(rows)
}

func TestJsxNode(t *testing.T) {
	manifest, count := jsxManifest(t)
	directory, _ := filepath.Abs(".")
	oracle := goOracle(t)
	want := execute(t, "", oracle, "--manifest", manifest, "--whole")
	seen := map[string]bool{}
	for _, line := range strings.Split(string(want.output), "\n") {
		fields := strings.Fields(line)
		if len(fields) > 2 {
			seen[fields[1]] = true
		}
	}
	kinds := strings.Fields(string(execute(t, "", oracle, "--jsx-kinds").output))
	for _, kind := range kinds {
		if !seen[kind] {
			t.Fatalf("Go JSX inventory kind not covered: %s", kind)
		}
	}
	t.Logf("all %d Go JSX AST kinds covered", len(kinds))
	got := wholeNode(t, directory, manifest, false)
	if diff := difference(got.output, want.output); diff != "" {
		t.Fatal(diff)
	}
	t.Logf("%d JSX inputs, %d identical whole-tree bytes", count, len(want.output))
}

// Not parallel: release throughput and full native tree comparisons share this corpus.
func TestJsxNative(t *testing.T) {
	manifest, count := jsxManifest(t)
	directory, _ := filepath.Abs(".")
	oracle := goOracle(t)
	want := execute(t, "", oracle, "--manifest", manifest, "--whole")
	for _, sanitize := range []bool{true, false} {
		binary := buildPort(t, directory, sanitize)
		got := execute(t, "", binary, "--manifest", manifest, "--whole")
		if diff := difference(got.output, want.output); diff != "" {
			t.Fatal(diff)
		}
		t.Logf("sanitize=%t: %d JSX inputs, %d identical whole-tree bytes", sanitize, count, len(want.output))
	}
}
