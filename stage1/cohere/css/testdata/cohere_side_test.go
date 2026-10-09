package postcss

import (
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
	var request struct{ Cases, Answers, IDs, Repository, Fixtures string }
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
	keys := make([]string, len(texts))
	for i := range keys {
		keys[i] = fmt.Sprintf("pinned/cohere@7945d102a6c18dd36adf9114a758ce646e8b2359/%d", i)
	}
	files := map[string]int{}
	repositoryFiles := 0
	roots := []string{request.Repository}
	if request.Fixtures != "" {
		roots = append(roots, request.Fixtures)
	} else {
		t.Log("Prettier fixtures absent; set ADAMIC_CSS_FIXTURES")
	}
	for _, root := range roots {
		err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() {
				if entry.Name() == ".git" || entry.Name() == "node_modules" {
					return filepath.SkipDir
				}
				return nil
			}
			ext := strings.ToLower(filepath.Ext(path))
			if ext != ".css" && ext != ".scss" && ext != ".less" {
				return nil
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			if !utf8.Valid(data) {
				return fmt.Errorf("%s is not UTF-8", path)
			}
			texts = append(texts, string(data))
			relative, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			prefix := "fixtures/"
			if root == request.Repository {
				prefix = "repository/"
				repositoryFiles++
			}
			keys = append(keys, prefix+filepath.ToSlash(relative)+":0")
			files[ext]++
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	if repositoryFiles == 0 {
		t.Fatal("repository CSS corpus is empty")
	}
	t.Logf("corpus files: %v", files)
	generatedStart := len(texts)
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
	for i := generatedStart; i < len(texts); i++ {
		keys = append(keys, fmt.Sprintf("generated/seed-20261006/%d", i-generatedStart))
	}
	var ids []string
	var cases, answers strings.Builder
	count, parsed := 0, 0
	for textIndex, text := range texts {
		if !utf8.ValidString(text) {
			t.Fatal("invalid UTF-8 in generated corpus")
		}
		for _, mode := range []string{"C", "S"} {
			ids = append(ids, keys[textIndex]+":"+mode)
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
	if request.IDs != "" {
		data, err := json.Marshal(ids)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(request.IDs, data, 0644); err != nil {
			t.Fatal(err)
		}
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
