//go:build lintoracle

// Observe actual test calls through test-file-only overlays; production Go is unchanged.
package high_level_intermediate_representation

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/cohere/internal/lint/rule"
	react_conformance "github.com/system-inc/cohere/internal/lint/rules/react/conformance"
	rule_testing "github.com/system-inc/cohere/internal/lint/testing"
)

type constructionRecord struct {
	Key       string   `json:"key"`
	Source    string   `json:"source"`
	Start     int      `json:"start"`
	End       int      `json:"end"`
	Checker   bool     `json:"checker"`
	Dump      string   `json:"dump"`
	Functions int      `json:"functions"`
	Eligible  bool     `json:"eligible"`
	Calls     []string `json:"calls"`
}

var constructionRecords = map[string]*constructionRecord{}
var constructionMutex sync.Mutex

func constructionExpression(n *ast.Node) bool {
	if n == nil {
		return false
	}
	switch n.Kind {
	case ast.KindNumericLiteral, ast.KindStringLiteral, ast.KindBigIntLiteral, ast.KindNoSubstitutionTemplateLiteral, ast.KindTrueKeyword, ast.KindFalseKeyword, ast.KindNullKeyword:
		return true
	case ast.KindParenthesizedExpression:
		return constructionExpression(n.AsParenthesizedExpression().Expression)
	case ast.KindTypeOfExpression:
		return constructionExpression(n.AsTypeOfExpression().Expression)
	case ast.KindVoidExpression:
		return constructionExpression(n.AsVoidExpression().Expression)
	case ast.KindPrefixUnaryExpression:
		x := n.AsPrefixUnaryExpression()
		return x.Operator != ast.KindPlusPlusToken && x.Operator != ast.KindMinusMinusToken && constructionExpression(x.Operand)
	case ast.KindBinaryExpression:
		x := n.AsBinaryExpression()
		op := x.OperatorToken.Kind
		return op != ast.KindAmpersandAmpersandToken && op != ast.KindBarBarToken && op != ast.KindQuestionQuestionToken && op != ast.KindEqualsToken && !isCompoundAssignment(op) && constructionExpression(x.Left) && constructionExpression(x.Right)
	}
	return false
}
func constructionEligible(node *ast.Node) bool {
	if node.Kind != ast.KindFunctionDeclaration && node.Kind != ast.KindFunctionExpression && node.Kind != ast.KindArrowFunction {
		return false
	}
	body := functionBody(node)
	if body == nil || node.Type() != nil {
		return false
	}
	if types := node.TypeParameterList(); types != nil && len(types.Nodes) > 0 {
		return false
	}
	if hasModifier(node, ast.KindAsyncKeyword) || functionIsGenerator(node) {
		return false
	}
	if modifiers := node.Modifiers(); modifiers != nil && len(modifiers.Nodes) > 0 {
		return false
	}
	if parameters := functionParameters(node); parameters != nil {
		for _, p := range parameters.Nodes {
			x := p.AsParameterDeclaration()
			if x.Name().Kind != ast.KindIdentifier || x.Type != nil || x.Initializer != nil || x.DotDotDotToken != nil || x.QuestionToken != nil {
				return false
			}
		}
	}
	if body.Kind != ast.KindBlock {
		return node.Kind == ast.KindArrowFunction && constructionExpression(body)
	}
	statements := body.AsBlock().Statements
	if statements != nil {
		for index, statement := range statements.Nodes {
			switch statement.Kind {
			case ast.KindEmptyStatement:
			case ast.KindExpressionStatement:
				if !constructionExpression(statement.AsExpressionStatement().Expression) {
					return false
				}
			case ast.KindReturnStatement:
				if index != len(statements.Nodes)-1 {
					return false
				}
				if expression := statement.AsReturnStatement().Expression; expression != nil && !constructionExpression(expression) {
					return false
				}
			default:
				return false
			}
		}
	}
	return true
}

func constructionObserve(f *Function, checked bool, caller string, constructed bool) {
	if f == nil || f.Node == nil {
		return
	}
	source := ast.GetSourceFileOfNode(f.Node)
	if source == nil {
		return
	}
	clone := CloneFunction(f)
	if !constructed {
		Construct(clone)
	}
	constructionStore(clone, checked, caller)
}
func constructionStore(clone *Function, checked bool, caller string) {
	source := ast.GetSourceFileOfNode(clone.Node)
	if source == nil {
		return
	}
	f := clone
	dump := oracleDump(clone)
	key := fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprintf("%s:%d:%d:%t:%s", source.Text(), f.Node.Pos(), f.Node.End(), checked, dump))))
	constructionMutex.Lock()
	if old := constructionRecords[key]; old != nil {
		old.Calls = append(old.Calls, caller)
	} else {
		constructionRecords[key] = &constructionRecord{Key: key, Source: source.Text(), Start: f.Node.Pos(), End: f.Node.End(), Checker: checked, Dump: dump, Functions: 1, Eligible: constructionEligible(f.Node), Calls: []string{caller}}
	}
	constructionMutex.Unlock()
	for _, nested := range clone.Functions {
		constructionStore(nested, checked, caller+"/nested")
	}
}

func stage1ObservedLower(node *ast.Node, checker *checker.Checker) *Function {
	f := Lower(node, checker)
	_, file, line, _ := runtime.Caller(1)
	constructionObserve(f, checker != nil, fmt.Sprintf("%s:%d", filepath.Base(file), line), false)
	return f
}
func stage1ObservedForFunction(ctx rule.Context, node *ast.Node) *Function {
	f := ForFunction(ctx, node)
	_, file, line, _ := runtime.Caller(1)
	constructionObserve(f, ctx.TypeChecker != nil, fmt.Sprintf("%s:%d", filepath.Base(file), line), true)
	return f
}

func stage1ObservedForFunctionWithoutManualMemoization(ctx rule.Context, node *ast.Node) *Function {
	f := ForFunctionWithoutManualMemoization(ctx, node)
	_, file, line, _ := runtime.Caller(1)
	if ctx.TypeChecker != nil {
		constructionObserve(Lower(node, ctx.TypeChecker), true, fmt.Sprintf("%s:%d", filepath.Base(file), line), false)
	}
	return f
}

var constructionFixtures int
var constructionFlow []string

func TestStage1AllConstructionFixtures(t *testing.T) {
	fixtures, err := react_conformance.Load("../../rules/react/conformance/testdata/fixtures")
	if err != nil {
		t.Fatal(err)
	}
	constructionFixtures = len(fixtures)
	for _, fixture := range fixtures {
		if fixture.RequiresFlow() {
			constructionFlow = append(constructionFlow, fixture.Name)
			continue
		}
		probe := rule.Rule{Name: "stage1-construction-oracle", NeedsTypeChecker: true, Run: func(ctx rule.Context, _ any) rule.Listeners {
			return rule.Listeners{ast.KindSourceFile: func(node *ast.Node) {
				forEachFunctionLike(node, func(n *ast.Node) {
					f := Lower(n, ctx.TypeChecker)
					constructionObserve(f, true, "upstream:"+fixture.Name, false)
				})
			}}
		}}
		rule_testing.RunTypedFilesWithSetup(t, probe, map[string]string{fixture.Name: fixture.Source}, fixture.Name, func(directory string) {
			config := `{"compilerOptions":{"strict":true,"target":"ES2022","lib":["ES2022"],"moduleDetection":"force","types":[],"allowJs":true,"jsx":"preserve"},"include":["**/*.ts","**/*.tsx","**/*.js","**/*.jsx"]}`
			if err := os.WriteFile(filepath.Join(directory, "tsconfig.json"), []byte(config), 0600); err != nil {
				t.Fatal(err)
			}
		})
	}
}
func TestMain(m *testing.M) {
	code := m.Run()
	destination := os.Getenv("HIR_CENSUS")
	if destination != "" {
		if err := os.MkdirAll(destination, 0755); err != nil {
			panic(err)
		}
		keys := []string{}
		for key := range constructionRecords {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		var manifest strings.Builder
		records := []*constructionRecord{}
		total, eligible := 0, 0
		for _, key := range keys {
			r := constructionRecords[key]
			sort.Strings(r.Calls)
			r.Calls = uniqueConstructionCalls(r.Calls)
			sourcePath := filepath.Join(destination, key+".tsx")
			dumpPath := filepath.Join(destination, key+".dump")
			if err := os.WriteFile(sourcePath, []byte(r.Source), 0600); err != nil {
				panic(err)
			}
			if err := os.WriteFile(dumpPath, []byte(r.Dump), 0600); err != nil {
				panic(err)
			}
			fmt.Fprintf(&manifest, "%s\t%s\t%d\t%d\t%t\t%s\t%d\t%t\n", key, sourcePath, r.Start, r.End, r.Checker, dumpPath, r.Functions, r.Eligible)
			total += r.Functions
			if r.Eligible {
				eligible += r.Functions
			}
			records = append(records, r)
		}
		encoded, err := json.MarshalIndent(struct {
			Records           []*constructionRecord
			Fixtures          int
			Flow              []string
			TotalFunctions    int
			EligibleFunctions int
		}{records, constructionFixtures, constructionFlow, total, eligible}, "", "  ")
		if err != nil {
			panic(err)
		}
		if err := os.WriteFile(filepath.Join(destination, "records.json"), encoded, 0600); err != nil {
			panic(err)
		}
		if err := os.WriteFile(filepath.Join(destination, "manifest.tsv"), []byte(manifest.String()), 0600); err != nil {
			panic(err)
		}
		fmt.Printf("HIR construction census: %d context-distinct functions in %d records; %d eligible; %d fixtures (%d Flow exclusions)\n", total, len(keys), eligible, constructionFixtures, len(constructionFlow))
	}
	os.Exit(code)
}
func uniqueConstructionCalls(values []string) []string {
	out := []string{}
	for _, value := range values {
		if len(out) == 0 || out[len(out)-1] != value {
			out = append(out, value)
		}
	}
	return out
}

func TestStage1ConstructionPathProbes(t *testing.T) {
	sources := []string{
		"function Arithmetic(unused) { 1 + 2 * 3; return 4; }",
		"function Unary() { typeof 1; void 2; +3; ~4; return !false; }",
		"function Comma() { return (1, 2); }",
		"function Strings() { return 'line\\n🙂\\\\'; }",
		"function Bigint() { return 123n; }",
		"function Literal() { return true; }",
		"const Assigned = () => 5 + 6;",
		"const BlockArrow = (unused) => { return -7; };",
		"const Expression = function Named(unused) { return 8; };",
	}
	for index, source := range sources {
		probe := rule.Rule{Name: "stage1-path-probe", NeedsTypeChecker: true, Run: func(ctx rule.Context, _ any) rule.Listeners {
			return rule.Listeners{ast.KindSourceFile: func(node *ast.Node) {
				forEachFunctionLike(node, func(n *ast.Node) {
					constructionObserve(Lower(n, ctx.TypeChecker), true, fmt.Sprintf("probe:%d", index), false)
				})
			}}
		}}
		rule_testing.RunTyped(t, probe, fmt.Sprintf("probe-%d.tsx", index), source)
	}
}
