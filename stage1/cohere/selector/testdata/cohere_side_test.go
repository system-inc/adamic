// Laid into Go cohere with go test -overlay. No submodule source is changed.
package selector

import (
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	goparser "go/parser"
	gotoken "go/token"
	"io/fs"
	"math/rand"
	randv2 "math/rand/v2"
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
	path := os.Getenv("ADAMIC_PORT_REQUEST")
	if path == "" {
		// census: required-input stage1/cohere/selector/selector_test.go cohereSide writes ADAMIC_PORT_REQUEST JSON in t.TempDir and runs this overlay in pinned cohere; see docs/gate-inputs.md.
		t.Skip("run by Adamic selector slice")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var request struct {
		Seed                              int64
		Generated                         int
		Cases, Answers, Corpus, TestTexts string
	}
	if err := json.Unmarshal(data, &request); err != nil {
		t.Fatal(err)
	}
	constants := adamicCohereTexts(t)
	if request.TestTexts != "" {
		encoded, err := json.Marshal(constants)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(request.TestTexts, encoded, 0644); err != nil {
			t.Fatal(err)
		}
	}
	texts := constants
	for _, fixture := range selectorFixtures {
		texts = append(texts, fixture.text)
	}
	for _, operator := range []string{"=", "~=", "|=", "^=", "$=", "*="} {
		for _, flag := range []string{"", " i", " I", " s", " S", "\ti ", "\u00a0i", "\u2029i", "\ufeffi"} {
			for _, value := range []string{"b", "'b'", "\"b\"", "", "'a\\'b'", "😀"} {
				texts = append(texts, "[ns|a"+operator+value+flag+"]")
			}
		}
	}
	for _, text := range []string{"a| b", ":is(a|)", "a|@x", "*|)", ".a\\.b", "a.b\\ ", "'a\\'b'", ":is(:not(a))"} {
		texts = append(texts, text)
	}
	for _, depth := range []int{1, 10, 100, 300} {
		text := strings.Repeat(":is(", depth) + ".a" + strings.Repeat(")", depth)
		texts = append(texts, text, text[:len(text)-1])
	}
	// Reproduce cohere's default random oracle corpus exactly, including the
	// unique nonterminating inputs that its upstream comparison skips.
	original := randv2.New(randv2.NewPCG(2, 3))
	seen := map[string]bool{}
	terminating := 0
	for terminating < 20000 {
		var text strings.Builder
		for range 1 + original.IntN(10) {
			text.WriteString(randomSelectorPieces[original.IntN(len(randomSelectorPieces))])
		}
		if seen[text.String()] {
			continue
		}
		seen[text.String()] = true
		texts = append(texts, text.String())
		if _, err := Parse(text.String()); !errors.Is(err, ErrLoopsForever) {
			terminating++
		}
	}
	t.Logf("cohere PCG(2,3) corpus: %d unique inputs, %d terminating", len(seen), terminating)
	random := rand.New(rand.NewSource(request.Seed))
	for range request.Generated {
		var b strings.Builder
		for range 1 + random.Intn(14) {
			b.WriteString(randomSelectorPieces[random.Intn(len(randomSelectorPieces))])
		}
		texts = append(texts, b.String())
	}
	if corpus, err := os.ReadFile(request.Corpus); err == nil {
		for _, line := range strings.Split(string(corpus), "\n") {
			if line == "" {
				continue
			}
			text, err := strconv.Unquote(line)
			if err != nil {
				t.Fatal(err)
			}
			texts = append(texts, text)
		}
	}
	escape := strings.NewReplacer(`\`, `\\`, "\t", `\t`, "\n", `\n`, "\r", `\r`)
	var cases, answers strings.Builder
	for number, text := range texts {
		if !utf8.ValidString(text) {
			t.Fatalf("invalid UTF-8 case %d", number)
		}
		cases.WriteString(">" + escape.Replace(text) + "\n")
		fmt.Fprintf(&answers, "case %d\n", number)
		tree, err := Parse(text)
		if err != nil {
			answers.WriteString("error " + err.Error() + "\n")
			continue
		}
		encoder := json.NewEncoder(&answers)
		encoder.SetEscapeHTML(false)
		if err := encoder.Encode(dumpSelectorValue(tree)); err != nil {
			t.Fatal(err)
		}
	}
	t.Logf("%d CSS test constants plus fixtures and generated selectors; total %d", len(adamicCohereTexts(t)), len(texts))
	if err := os.WriteFile(request.Cases, []byte(cases.String()), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(request.Answers, []byte(answers.String()), 0644); err != nil {
		t.Fatal(err)
	}
	started := time.Now()
	parsed, nodes := 0, 0
	for range 10 {
		for _, text := range texts {
			if tree, err := Parse(text); err == nil {
				parsed++
				nodes += len(tree.List("nodes"))
			}
		}
	}
	t.Logf("Go: %d selectors in %s, %.0f selectors/s; %d parsed, %d root children", len(texts)*10, time.Since(started), float64(len(texts)*10)/time.Since(started).Seconds(), parsed, nodes)
}
