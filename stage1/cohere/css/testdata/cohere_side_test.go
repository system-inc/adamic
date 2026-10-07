package postcss

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	goparser "go/parser"
	gotoken "go/token"
	"io/fs"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
	"unicode/utf8"
)

func adamicCohereTexts(t *testing.T) []string {
	t.Helper()
	var paths []string
	err := filepath.WalkDir("..", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() && strings.HasSuffix(path, "_test.go") && !strings.HasPrefix(entry.Name(), "adamic_") {
			paths = append(paths, path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	var texts []string
	seen := map[string]bool{}
	var constant func(expression ast.Expr) (string, bool)
	constant = func(expression ast.Expr) (string, bool) {
		switch typed := expression.(type) {
		case *ast.BasicLit:
			if typed.Kind != gotoken.STRING {
				return "", false
			}
			text, err := strconv.Unquote(typed.Value)
			return text, err == nil
		case *ast.ParenExpr:
			return constant(typed.X)
		case *ast.BinaryExpr:
			if typed.Op != gotoken.ADD {
				return "", false
			}
			left, isLeft := constant(typed.X)
			right, isRight := constant(typed.Y)
			return left + right, isLeft && isRight
		}
		return "", false
	}
	files := gotoken.NewFileSet()
	for _, name := range paths {
		if !strings.HasSuffix(name, "_test.go") || strings.HasPrefix(name, "adamic_") {
			continue
		}
		file, err := goparser.ParseFile(files, name, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		ast.Inspect(file, func(node ast.Node) bool {
			if _, isImport := node.(*ast.ImportSpec); isImport {
				return false
			}
			expression, isExpression := node.(ast.Expr)
			if !isExpression {
				return true
			}
			text, isConstant := constant(expression)
			if !isConstant {
				return true
			}
			if !seen[text] {
				seen[text] = true
				texts = append(texts, text)
			}
			// A sum is one text; its parts are not texts of their own.
			return false
		})
	}
	return texts
}

func TestAdamicPortCases(t *testing.T) {
	var request struct{ Cases, Answers, Repository, Fixtures, Corpus string }
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
	texts := adamicCohereTexts(t)
	for _, fixture := range parseFixtures {
		texts = append(texts, fixture.text)
	}
	for _, fixture := range scssParseFixtures {
		texts = append(texts, fixture.text)
	}
	if request.Fixtures == "" {
		t.Fatal("ADAMIC_CSS_FIXTURES unset; #xq2ecw6 (setup --gate-inputs) provisions this oracle")
	}
	manifest, err := os.ReadFile(request.Corpus)
	if err != nil {
		t.Fatal(err)
	}
	files := map[string]int{}
	seen := map[string]bool{}
	for index, row := range strings.Split(string(manifest), "\n") {
		if row == "" || strings.HasPrefix(row, "#") {
			continue
		}
		fields := strings.SplitN(row, "\t", 2)
		if len(fields) != 2 || len(fields[0]) != 64 || !filepath.IsLocal(fields[1]) || seen[fields[1]] {
			t.Fatalf("invalid CSS corpus row %d: %q", index+1, row)
		}
		seen[fields[1]] = true
		path := filepath.Join(request.Fixtures, filepath.FromSlash(fields[1]))
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("listed CSS fixture %s: %v", fields[1], err)
		}
		actual := fmt.Sprintf("%x", sha256.Sum256(data))
		if actual != fields[0] {
			t.Fatalf("listed CSS fixture %s: sha256 %s, want %s", fields[1], actual, fields[0])
		}
		if !utf8.Valid(data) {
			t.Fatalf("listed CSS fixture %s is not UTF-8", fields[1])
		}
		texts = append(texts, string(data))
		files[strings.ToLower(filepath.Ext(path))]++
	}
	if len(seen) == 0 {
		t.Fatal("CSS corpus list is empty")
	}
	t.Logf("pinned corpus: %d files; %v; each run in CSS and SCSS modes", len(seen), files)
	parts := []string{"a", ".a", "#a", "&", "@unknown", "@media screen", "@supports (a:b)", "@a", " ", "\n", "\t", "/*x*/", "/* */", "// a\n", ":", ";", "{", "}", "(", ")", "[", "]", "--x", "b", "c", "!important", "! IMPORTANT", "1px", "#{$x}", "\"a\"", "'x'", "url(x)", "url(a(b))", "a\\:b", "\\e9 ", "😀", "é", ",", "!default"}
	random := rand.New(rand.NewSource(20261006))
	for range 6000 {
		var b strings.Builder
		for range 1 + random.Intn(18) {
			b.WriteString(parts[random.Intn(len(parts))])
		}
		texts = append(texts, b.String())
	}
	for _, comment := range []string{"", "/*x*/", "/* */", "// x\n"} {
		for _, value := range []string{"", "a", "(a:b)", "{a:b}", "{a:{b:c}}", "url(a(b))", "'/*x*/'", "[a;b]", "!important", "#{a}", "a:b", "a b:c"} {
			texts = append(texts, "a"+comment+"{"+comment+"--x"+comment+":"+comment+value+comment+";"+comment+"}")
		}
	}
	// Every truncation of the fixture strings, at valid Unicode boundaries.
	for _, fixture := range parseFixtures {
		for index := range fixture.text {
			texts = append(texts, fixture.text[:index])
		}
	}
	escape := strings.NewReplacer(`\`, `\\`, "\t", `\t`, "\n", `\n`, "\r", `\r`)
	var cases, answers strings.Builder
	count, parsed := 0, 0
	for _, text := range texts {
		if !utf8.ValidString(text) {
			t.Fatal("invalid UTF-8 in generated corpus")
		}
		for _, mode := range []string{"C", "S"} {
			cases.WriteString(">" + mode + escape.Replace(text) + "\n")
			fmt.Fprintf(&answers, "case %d\n", count)
			count++
			tree, err := Parse(text)
			if mode == "S" {
				tree, err = ParseSCSS(text)
			}
			var answer any
			if err != nil {
				var syntax *CssSyntaxError
				if errors.As(err, &syntax) {
					fields := map[string]any{"name": syntax.Name(), "reason": syntax.Reason, "line": syntax.Line, "column": syntax.Column, "offset": syntax.Offset}
					if syntax.HasEnd {
						fields["endLine"] = syntax.EndLine
						fields["endColumn"] = syntax.EndColumn
						fields["endOffset"] = syntax.EndOffset
					}
					answer = fields
				} else {
					answer = map[string]any{"notCssSyntaxError": err.Error()}
				}
				answers.WriteString("error ")
			} else {
				answer = dumpParseValue(tree, func(index int) int { return index })
				parsed++
			}
			encoder := json.NewEncoder(&answers)
			encoder.SetEscapeHTML(false)
			if err := encoder.Encode(answer); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := os.WriteFile(request.Cases, []byte(cases.String()), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(request.Answers, []byte(answers.String()), 0644); err != nil {
		t.Fatal(err)
	}
	t.Logf("%d cases, %d parsed", count, parsed)
	started := time.Now()
	checked, nodes := 0, 0
	for range 10 {
		for _, text := range texts {
			for _, scss := range []bool{false, true} {
				parse := Parse
				if scss {
					parse = ParseSCSS
				}
				tree, err := parse(text)
				if err == nil {
					checked++
					nodes += len(tree.List("nodes"))
				}
			}
		}
	}
	t.Logf("Go: %d stylesheets in %s, %.0f stylesheets/s; %d parsed, %d children", len(texts)*20, time.Since(started), float64(len(texts)*20)/time.Since(started).Seconds(), checked, nodes)
}
