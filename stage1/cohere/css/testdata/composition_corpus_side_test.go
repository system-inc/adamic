package postcss

import (
	"encoding/json"
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
	var request struct{ Cases, Answers, Repository, Fixtures string }
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
	files := map[string]int{}
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
			files[ext]++
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	t.Logf("corpus files: %v", files)
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
	var cases strings.Builder
	for _, text := range texts {
		if !utf8.ValidString(text) {
			t.Fatal("invalid UTF-8 in generated corpus")
		}
		for _, mode := range []string{"C", "S"} {
			cases.WriteString(">" + mode + escape.Replace(text) + "\n")
		}
	}
	if err := os.WriteFile(request.Cases, []byte(cases.String()), 0644); err != nil {
		t.Fatal(err)
	}
	t.Logf("enumerated %d cases", strings.Count(cases.String(), "\n"))
}
