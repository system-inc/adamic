// Overlay in the Go JavaScript printer; selection uses Go's own independent ESTree adapter.
package javascript

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/system-inc/cohere/internal/format/estree"
	"github.com/system-inc/cohere/internal/format/formatoptions"
)

type portExpressionCase struct {
	Label  string
	Source string
	Want   string
}

func supportedExpression(node *estree.Node) bool {
	if node == nil {
		return true
	}
	switch node.Type() {
	case "Identifier", "PrivateIdentifier", "Literal", "ThisExpression", "Super":
		return true
	case "UnaryExpression", "UpdateExpression", "BinaryExpression", "LogicalExpression", "MemberExpression", "ArrayExpression", "SpreadElement", "TSNonNullExpression", "ChainExpression":
	case "TemplateLiteral":
		return len(node.List("expressions")) == 0
	case "CallExpression", "NewExpression":
		if !node.Child("callee").Is("Identifier") || node.Child("typeArguments") != nil {
			return false
		}
		for _, arg := range node.List("arguments") {
			if arg.Is("ArrayExpression") {
				return false
			}
		}
	default:
		return false
	}
	for _, child := range estree.ChildNodes(node) {
		if !supportedExpression(child) {
			return false
		}
	}
	return true
}
func TestAdamicExpressionCorpus(t *testing.T) {
	requestData, err := os.ReadFile(os.Getenv("ADAMIC_TS_EXPRESSION_REQUEST"))
	if err != nil {
		t.Fatal(err)
	}
	var request struct {
		Files     []string
		Directory string
		Gaps      string
	}
	if err = json.Unmarshal(requestData, &request); err != nil {
		t.Fatal(err)
	}
	cases := []portExpressionCase{}
	coverage := map[string]int{}
	rejected := map[string]int{}
	filesFailed := map[string]string{}
	add := func(label, source string) { cases = append(cases, portExpressionCase{Label: label, Source: source}) }
	for _, file := range request.Files {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		source := string(data)
		fileTree := estree.ParseSourceFile(file, source)
		if diagnostics := fileTree.Diagnostics(); len(diagnostics) > 0 {
			filesFailed[file] = diagnostics[0].String()
			continue
		}
		var visit func(*ast.Node) bool
		visit = func(node *ast.Node) bool {
			if node.Parent != nil && ast.IsExpressionNode(node) {
				start := scanner.GetTokenPosOfNode(node, fileTree, false)
				fragment := source[start:node.End()]
				if node.Kind != ast.KindSpreadElement && node.Kind != ast.KindOmittedExpression && node.Kind != ast.KindPrivateIdentifier && supportedSyntax(node) && coreBoundaries(node, source) && !strings.Contains(fragment, "/*") && !strings.Contains(fragment, "//") && !hasBlankLine(fragment) && !strings.Contains(fragment, "\r") && !(node.Kind == ast.KindNoSubstitutionTemplateLiteral && strings.Contains(fragment, "\n")) {
					// Property names and other context-only identifiers may not be valid
					// standalone expression statements (for example, obj.delete).
					if _, _, _, err := estree.ParseTypeScript("expression.ts", fragment+";", nil); err != nil {
						rejected["standalone-context"]++
						node.ForEachChild(visit)
						return false
					}
					add(fmt.Sprintf("%s:%d:%s", file, start, node.Kind), fragment)
					coverage[node.Kind.String()]++
					return false
				}
				rejected[node.Kind.String()]++
			}
			node.ForEachChild(visit)
			return false
		}
		visit(fileTree.AsNode())
	}

	operators := []string{"??", "||", "&&", "|", "^", "&", "==", "!=", "===", "!==", "<", ">", "<=", ">=", "in", "instanceof", "<<", ">>", ">>>", "+", "-", "*", "/", "%", "**"}
	for _, left := range operators {
		for _, right := range operators {
			add("operator-pair", "a "+left+" b "+right+" c")
		}
	}
	for _, item := range []string{"x", "this", "true", "false", "null", "0xAB", "1.0000", "1E+003", ".10", "123.", "1_000n", "0xABn", "'é😀'", "\"double\"", "'don\\'t'", "/foo/mi", "`raw\\u{1f600}`", "+x", "-(a+b)", "+ ++x", "- --x", "x++", "typeof a", "void f()", "delete obj.x", "obj?.x", "a?.[b+c]", "a!.x", "a.b.c.d", "[]", "[1,2,3]", "[1,,]", "[...items,]", "[[1,2],[3,4]]", "f()", "f?.(a,b)", "new C", "new C(a,b)", "a && [1,2,3]", "a * (b % c)", "a + (b % c)", "-(a || b)", "(1).toString", "(0xAF).x", "(1e0).x", "(1)[0]", "veryLongIdentifierAlpha.veryLongPropertyNameBeta!.veryLongPropertyNameGamma", "veryLongIdentifierAlpha.veryLongPropertyNameBeta[index]", "+x++", "++x in obj", "Boolean(veryLongIdentifierAlpha && veryLongIdentifierBeta && veryLongIdentifierGamma && veryLongIdentifierDelta)", "(veryLongIdentifierAlpha + veryLongIdentifierBeta + veryLongIdentifierGamma + veryLongIdentifierDelta)[index]"} {
		add("edge", item)
	}
	for length := 1; length <= 60; length++ {
		names := []string{}
		numbers := []string{}
		for index := 0; index < length; index++ {
			names = append(names, fmt.Sprintf("argument%d", index))
			numbers = append(numbers, fmt.Sprint(index))
		}
		add("argument-width", "f("+strings.Join(names, ",")+")")
		add("fill-width", "["+strings.Join(numbers, ",")+"]")
	}
	random := rand.New(rand.NewSource(20261006))
	var generate func(int) string
	generate = func(depth int) string {
		if depth == 0 {
			return []string{"x", "12", "true", "null", "'é😀'", "a.x", "/x/mi"}[random.Intn(7)]
		}
		switch random.Intn(5) {
		case 0:
			return "((" + generate(depth-1) + ") " + operators[random.Intn(len(operators))] + " (" + generate(depth-1) + "))"
		case 1:
			return "f(" + generate(depth-1) + "," + generate(depth-1) + ")"
		case 2:
			return "[" + generate(depth-1) + "," + generate(depth-1) + "]"
		case 3:
			return "!(" + generate(depth-1) + ")"
		default:
			return "(" + generate(depth-1) + ")"
		}
	}
	// Recursive generation can make an expanded array argument, outside this first core.
	for count := 0; count < 500; count++ {
		source := generate(3)
		root, _, _, err := estree.ParseTypeScript("random.ts", source+";", nil)
		if err != nil {
			t.Fatal(err)
		}
		node := root.List("body")[0].Child("expression")
		if supportedExpression(node) {
			add("random", source)
		} else {
			rejected["generated-expanded-argument"]++
		}
	}
	options := formatoptions.Default()
	options.PrintWidth = 80
	var input, answers strings.Builder
	escape := strings.NewReplacer("\\", "\\\\", "\n", "\\n", "\r", "\\r", "\t", "\\t")
	for index := range cases {
		item := &cases[index]
		item.Want, err = Format("expression.ts", item.Source+";", options, nil)
		if err != nil {
			t.Fatalf("case %d %s source %q: %v", index, item.Label, item.Source, err)
		}
		fmt.Fprintln(&input, ">"+escape.Replace(item.Source))
		fmt.Fprintln(&answers, "ok\t"+escape.Replace(item.Want))
	}
	encoded, err := json.Marshal(cases)
	if err != nil {
		t.Fatal(err)
	}
	report, _ := json.MarshalIndent(map[string]any{"files": len(request.Files), "cases": len(cases), "coverage": coverage, "unsupported_candidates": rejected, "file_refusals": filesFailed}, "", "  ")
	for name, data := range map[string][]byte{"cases.txt": []byte(input.String()), "answers.txt": []byte(answers.String()), "cases.json": encoded, "coverage.json": report} {
		if err := os.WriteFile(request.Directory+"/"+name, data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	gapData, err := os.ReadFile(request.Gaps)
	if err != nil {
		t.Fatal(err)
	}
	var gaps []struct{ Source, Reason string }
	if err = json.Unmarshal(gapData, &gaps); err != nil {
		t.Fatal(err)
	}
	gapSpecs := []portExpressionCase{}
	var gapInput, gapAnswers strings.Builder
	for _, gap := range gaps {
		formatted, err := Format("gap.ts", gap.Source+";", options, nil)
		if err != nil {
			t.Fatal(err)
		}
		gapSpecs = append(gapSpecs, portExpressionCase{Label: gap.Reason, Source: gap.Source, Want: formatted})
		fmt.Fprintln(&gapInput, ">"+escape.Replace(gap.Source))
		fmt.Fprintln(&gapAnswers, "notyet\t"+gap.Reason)
	}
	gapJSON, err := json.Marshal(gapSpecs)
	if err != nil {
		t.Fatal(err)
	}
	for name, data := range map[string][]byte{"gaps.txt": []byte(gapInput.String()), "gap-answers.txt": []byte(gapAnswers.String()), "gaps.json": gapJSON} {
		if err := os.WriteFile(request.Directory+"/"+name, data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	t.Logf("%d files, %d full-file parse refusals, %d supported maximal expression fragments", len(request.Files), len(filesFailed), len(cases))
}

func hasBlankLine(source string) bool {
	lines := strings.Split(source, "\n")
	for index := 1; index < len(lines)-1; index++ {
		if strings.TrimSpace(lines[index]) == "" {
			return true
		}
	}
	return false
}
func hasOptionalSyntax(node *ast.Node) bool {
	if node.Kind == ast.KindQuestionDotToken {
		return true
	}
	found := false
	node.ForEachChild(func(child *ast.Node) bool {
		if hasOptionalSyntax(child) {
			found = true
		}
		return false
	})
	return found
}

// Literal continuation and multiline-template layout is outside the expression core.
func coreBoundaries(node *ast.Node, source string) bool {
	if node.Kind == ast.KindParenthesizedExpression && hasOptionalSyntax(node) {
		return false
	}
	if node.Kind == ast.KindStringLiteral || node.Kind == ast.KindNoSubstitutionTemplateLiteral {
		return !strings.Contains(strings.TrimSpace(source[node.Pos():node.End()]), "\n")
	}
	valid := true
	node.ForEachChild(func(child *ast.Node) bool {
		if !coreBoundaries(child, source) {
			valid = false
		}
		return false
	})
	return valid
}

// The independent selector uses the original parser's expression-context predicate.
// ESTree identifiers can include a parameter's type annotation in their range.
func supportedSyntax(node *ast.Node) bool {
	if node == nil {
		return true
	}
	switch node.Kind {
	case ast.KindIdentifier, ast.KindPrivateIdentifier, ast.KindNumericLiteral, ast.KindBigIntLiteral, ast.KindStringLiteral, ast.KindRegularExpressionLiteral, ast.KindNoSubstitutionTemplateLiteral, ast.KindThisKeyword, ast.KindSuperKeyword, ast.KindNullKeyword, ast.KindTrueKeyword, ast.KindFalseKeyword, ast.KindOmittedExpression:
		return true
	case ast.KindPrefixUnaryExpression:
		return supportedSyntax(node.AsPrefixUnaryExpression().Operand)
	case ast.KindPostfixUnaryExpression:
		return supportedSyntax(node.AsPostfixUnaryExpression().Operand)
	case ast.KindBinaryExpression:
		item := node.AsBinaryExpression()
		switch item.OperatorToken.Kind {
		case ast.KindEqualsToken, ast.KindCommaToken, ast.KindPlusEqualsToken, ast.KindMinusEqualsToken, ast.KindAsteriskEqualsToken, ast.KindSlashEqualsToken, ast.KindPercentEqualsToken, ast.KindAsteriskAsteriskEqualsToken, ast.KindLessThanLessThanEqualsToken, ast.KindGreaterThanGreaterThanEqualsToken, ast.KindGreaterThanGreaterThanGreaterThanEqualsToken, ast.KindAmpersandEqualsToken, ast.KindBarEqualsToken, ast.KindCaretEqualsToken, ast.KindAmpersandAmpersandEqualsToken, ast.KindBarBarEqualsToken, ast.KindQuestionQuestionEqualsToken:
			return false
		}
		return supportedSyntax(item.Left) && supportedSyntax(item.Right)
	case ast.KindPropertyAccessExpression:
		return supportedSyntax(node.Expression())
	case ast.KindElementAccessExpression:
		return supportedSyntax(node.Expression()) && supportedSyntax(node.AsElementAccessExpression().ArgumentExpression)
	case ast.KindParenthesizedExpression, ast.KindNonNullExpression, ast.KindSpreadElement, ast.KindDeleteExpression, ast.KindTypeOfExpression, ast.KindVoidExpression:
		return supportedSyntax(node.Expression())
	case ast.KindArrayLiteralExpression:
		for _, child := range node.AsArrayLiteralExpression().Elements.Nodes {
			if !supportedSyntax(child) {
				return false
			}
		}
		return true
	case ast.KindCallExpression, ast.KindNewExpression:
		if node.Expression().Kind != ast.KindIdentifier || len(node.TypeArguments()) > 0 {
			return false
		}
		for _, child := range node.Arguments() {
			if child.Kind == ast.KindArrayLiteralExpression || !supportedSyntax(child) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

// This optional entry makes the same batch protocol available in a compiled Go test binary.
// Input decoding, parsing, printing, encoding and file output are inside each timed process.
func TestAdamicExpressionBenchmark(t *testing.T) {
	path := os.Getenv("ADAMIC_TS_BENCH_CASES")
	if path == "" {
		t.Skip("benchmark-only driver")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	decode := func(text string) string {
		var result strings.Builder
		for index := 0; index < len(text); index++ {
			if text[index] != '\\' {
				result.WriteByte(text[index])
				continue
			}
			index++
			if index >= len(text) {
				t.Fatal("bad escape")
			}
			switch text[index] {
			case 'n':
				result.WriteByte('\n')
			case 'r':
				result.WriteByte('\r')
			case 't':
				result.WriteByte('\t')
			case '\\':
				result.WriteByte('\\')
			default:
				t.Fatal("bad escape")
			}
		}
		return result.String()
	}
	options := formatoptions.Default()
	options.PrintWidth = 80
	if width := os.Getenv("ADAMIC_TS_BENCH_WIDTH"); width != "" {
		options.PrintWidth, err = strconv.Atoi(width)
		if err != nil {
			t.Fatal(err)
		}
	}
	escape := strings.NewReplacer("\\", "\\\\", "\n", "\\n", "\r", "\\r", "\t", "\\t")
	var output strings.Builder
	for _, line := range strings.Split(string(data), "\n") {
		if line == "" {
			continue
		}
		formatted, err := Format("expression.ts", decode(line[1:])+";", options, nil)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Fprintln(&output, "ok\t"+escape.Replace(formatted))
	}
	if err := os.WriteFile(os.Getenv("ADAMIC_TS_BENCH_OUTPUT"), []byte(output.String()), 0644); err != nil {
		t.Fatal(err)
	}
}
