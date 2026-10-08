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
	"unicode/utf16"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/cohere/internal/lint/rule"
	react_conformance "github.com/system-inc/cohere/internal/lint/rules/react/conformance"
	rule_testing "github.com/system-inc/cohere/internal/lint/testing"
)

type constructionRecord struct {
	Key        string   `json:"key"`
	Source     string   `json:"source"`
	Start      int      `json:"start"`
	End        int      `json:"end"`
	Checker    bool     `json:"checker"`
	Dump       string   `json:"dump"`
	Functions  int      `json:"functions"`
	Eligible   bool     `json:"eligible"`
	Calls      []string `json:"calls"`
	Symbols    string   `json:"symbols"`
	RootStart  int      `json:"rootStart"`
	RootEnd    int      `json:"rootEnd"`
	NestedPath string   `json:"nestedPath"`
	Excluded   string   `json:"excluded"`
}

var constructionRecords = map[string]*constructionRecord{}
var constructionMutex sync.Mutex

func constructionExpression(n *ast.Node) bool {
	if n == nil {
		return false
	}
	switch n.Kind {
	case ast.KindFunctionExpression, ast.KindArrowFunction:
		return constructionEligible(n)
	case ast.KindIdentifier, ast.KindNumericLiteral, ast.KindStringLiteral, ast.KindBigIntLiteral, ast.KindNoSubstitutionTemplateLiteral, ast.KindTrueKeyword, ast.KindFalseKeyword, ast.KindNullKeyword:
		return true
	case ast.KindParenthesizedExpression:
		return constructionExpression(n.AsParenthesizedExpression().Expression)
	case ast.KindTypeOfExpression:
		return constructionExpression(n.AsTypeOfExpression().Expression)
	case ast.KindVoidExpression:
		return constructionExpression(n.AsVoidExpression().Expression)
	case ast.KindConditionalExpression:
		x := n.AsConditionalExpression()
		return constructionExpression(x.Condition) && constructionExpression(x.WhenTrue) && constructionExpression(x.WhenFalse)
	case ast.KindPropertyAccessExpression:
		x := n.AsPropertyAccessExpression()
		return x.QuestionDotToken == nil && constructionExpression(x.Expression)
	case ast.KindElementAccessExpression:
		x := n.AsElementAccessExpression()
		return x.QuestionDotToken == nil && constructionExpression(x.Expression) && constructionExpression(x.ArgumentExpression)
	case ast.KindCallExpression, ast.KindNewExpression:
		var callee *ast.Node
		var arguments *ast.NodeList
		if n.Kind == ast.KindCallExpression {
			x := n.AsCallExpression()
			callee = x.Expression
			arguments = x.Arguments
		} else {
			x := n.AsNewExpression()
			callee = x.Expression
			arguments = x.Arguments
		}
		if !constructionExpression(callee) {
			return false
		}
		if arguments != nil {
			for _, arg := range arguments.Nodes {
				if arg.Kind == ast.KindSpreadElement {
					arg = arg.AsSpreadElement().Expression
				}
				if !constructionExpression(arg) {
					return false
				}
			}
		}
		return true
	case ast.KindPostfixUnaryExpression:
		return constructionTarget(n.AsPostfixUnaryExpression().Operand)
	case ast.KindPrefixUnaryExpression:
		x := n.AsPrefixUnaryExpression()
		if x.Operator == ast.KindPlusPlusToken || x.Operator == ast.KindMinusMinusToken {
			return constructionTarget(x.Operand)
		}
		return constructionExpression(x.Operand)
	case ast.KindBinaryExpression:
		x := n.AsBinaryExpression()
		op := x.OperatorToken.Kind
		if op == ast.KindEqualsToken || isCompoundAssignment(op) {
			return constructionTarget(x.Left) && constructionExpression(x.Right)
		}
		return constructionExpression(x.Left) && constructionExpression(x.Right)
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
	return constructionStatement(body)
}
func constructionStatement(statement *ast.Node) bool {
	if statement == nil {
		return false
	}
	switch statement.Kind {
	case ast.KindFunctionDeclaration:
		return constructionEligible(statement)
	case ast.KindBlock:
		if list := statement.AsBlock().Statements; list != nil {
			for _, n := range list.Nodes {
				if !constructionStatement(n) {
					return false
				}
			}
		}
		return true
	case ast.KindIfStatement:
		x := statement.AsIfStatement()
		return constructionExpression(x.Expression) && constructionStatement(x.ThenStatement) && (x.ElseStatement == nil || constructionStatement(x.ElseStatement))
	case ast.KindWhileStatement:
		x := statement.AsWhileStatement()
		return constructionExpression(x.Expression) && constructionStatement(x.Statement)
	case ast.KindBreakStatement:
		return statement.AsBreakStatement().Label == nil
	case ast.KindContinueStatement:
		return statement.AsContinueStatement().Label == nil
	case ast.KindVariableStatement:
		list := statement.AsVariableStatement().DeclarationList.AsVariableDeclarationList()
		for _, decl := range list.Declarations.Nodes {
			d := decl.AsVariableDeclaration()
			if d.Name().Kind != ast.KindIdentifier || d.Type != nil || (d.Initializer != nil && !constructionExpression(d.Initializer)) {
				return false
			}
		}
		return true
	case ast.KindExpressionStatement:
		return constructionExpression(statement.AsExpressionStatement().Expression)
	case ast.KindReturnStatement:
		x := statement.AsReturnStatement()
		return x.Expression == nil || constructionExpression(x.Expression)
	case ast.KindThrowStatement:
		return constructionExpression(statement.AsThrowStatement().Expression)
	case ast.KindEmptyStatement:
		return true
	}
	return false
}

func constructionObserve(f *Function, typeChecker *checker.Checker, caller string, constructed bool) {
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
	constructionStore(clone, typeChecker != nil, constructionSymbols(source, typeChecker), caller, clone.Node, "")
}
func constructionStore(clone *Function, checked bool, symbols string, caller string, root *ast.Node, path string) {
	source := ast.GetSourceFileOfNode(clone.Node)
	if source == nil {
		return
	}
	f := clone
	excluded := ""
	if (react_conformance.Fixture{Source: source.Text()}).RequiresFlow() {
		excluded = "Flow"
	}
	dump := oracleDump(clone)
	if oracleDump(clone) != dump {
		panic("hir-v1 dump is nondeterministic")
	}
	key := fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprintf("%s:%d:%d:%t:%s", source.Text(), f.Node.Pos(), f.Node.End(), checked, dump))))
	eligible := excluded == "" && constructionEligible(f.Node)
	rootStart, rootEnd, nestedPath := f.Node.Pos(), f.Node.End(), ""
	if len(f.Context) > 0 {
		eligible = eligible && constructionEligible(root)
		rootStart, rootEnd, nestedPath = root.Pos(), root.End(), path
	}
	constructionMutex.Lock()
	if old := constructionRecords[key]; old != nil {
		old.Calls = append(old.Calls, caller)
		if eligible && !old.Eligible {
			old.Eligible = true
			old.RootStart = rootStart
			old.RootEnd = rootEnd
			old.NestedPath = nestedPath
		}
	} else {
		constructionRecords[key] = &constructionRecord{Key: key, Source: source.Text(), Start: f.Node.Pos(), End: f.Node.End(), Checker: checked, Dump: dump, Functions: 1, Eligible: eligible, RootStart: rootStart, RootEnd: rootEnd, NestedPath: nestedPath, Calls: []string{caller}, Symbols: symbols, Excluded: excluded}
	}
	constructionMutex.Unlock()
	for index, nested := range clone.Functions {
		next := fmt.Sprint(index)
		if path != "" {
			next = path + "," + next
		}
		constructionStore(nested, checked, symbols, caller+"/nested", root, next)
	}
}

func stage1ObservedLower(node *ast.Node, checker *checker.Checker) *Function {
	f := Lower(node, checker)
	_, file, line, _ := runtime.Caller(1)
	constructionObserve(f, checker, fmt.Sprintf("%s:%d", filepath.Base(file), line), false)
	return f
}
func stage1ObservedForFunction(ctx rule.Context, node *ast.Node) *Function {
	f := ForFunction(ctx, node)
	_, file, line, _ := runtime.Caller(1)
	constructionObserve(f, ctx.TypeChecker, fmt.Sprintf("%s:%d", filepath.Base(file), line), true)
	return f
}

func stage1ObservedForFunctionWithoutManualMemoization(ctx rule.Context, node *ast.Node) *Function {
	f := ForFunctionWithoutManualMemoization(ctx, node)
	_, file, line, _ := runtime.Caller(1)
	if ctx.TypeChecker != nil {
		constructionObserve(Lower(node, ctx.TypeChecker), ctx.TypeChecker, fmt.Sprintf("%s:%d", filepath.Base(file), line), false)
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
					constructionObserve(f, ctx.TypeChecker, "upstream:"+fixture.Name, false)
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
			symbolsPath := filepath.Join(destination, key+".symbols")
			if err := os.WriteFile(symbolsPath, []byte(r.Symbols), 0600); err != nil {
				panic(err)
			}
			if err := os.WriteFile(sourcePath, []byte(r.Source), 0600); err != nil {
				panic(err)
			}
			if err := os.WriteFile(dumpPath, []byte(r.Dump), 0600); err != nil {
				panic(err)
			}
			fmt.Fprintf(&manifest, "%s\t%s\t%d\t%d\t%t\t%s\t%d\t%t\t%s\t%s\t%d\t%d\t%s\n", key, sourcePath, r.Start, r.End, r.Checker, dumpPath, r.Functions, r.Eligible, symbolsPath, r.Excluded, r.RootStart, r.RootEnd, r.NestedPath)
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
		"function Captures(x) { const f = () => x + x; return f; }",
		"function Grandparent(x) { return () => () => x; }",
		"function ContextWrites(flag) { let x = 1; const read = () => x; if (flag) x = 2; return read; }",
		"function InnerWrite() { let x = 1; const write = () => { x = 2; return x; }; return write; }",
		"function Declaration(x) { function Local() { return x; } return Local; }",
		"function Shadow(x) { const read = (x) => x; return read; }",
		"function Calls(fn, value) { fn?.(value); return fn(value, ...value); }",
		"function Methods(obj, key) { obj.method(1); return obj[key](2); }",
		"function Constructors(Ctor, value) { new Ctor; return new Ctor(value, ...value); }",
		"import React, {useState as state} from 'react'; import * as R from 'react'; const alias: unknown = state; function Origins() { state(1); alias(2); React.useEffect(3); return R['useRef'](4); }",
		"function Orphan() { return 1; 2; }",
		"function Branches(flag) { let x = 1; if (flag) { x = 2; } else { x = 3; } return x; }",
		"function Abrupt(flag) { if (flag) { return 1; } else { throw 2; } return 3; }",
		"function Ternary(flag) { return flag ? 1 : 2; }",
		"function Logical(a, b) { a && b; a || b; return a ?? b; }",
		"function Loop(n) { let x = 0; while (x < n) { x++; if (x === 2) continue; if (x === 3) break; } return x; }",
		"function Binding(value) { let local = value; local = 2; return local; }",
		"function Declaration() { let local; local = 3; return local; }",
		"function Updates(value) { ++value; value--; value += 2; return value; }",
		"import Default from './missing'; import * as Namespace from './missing'; import {x as Named} from './missing'; const local = 1; function Globals() { Default; Namespace; Named; local; unknown = 2; return unknown; }",
		"function Properties(object, key) { object.name = 1; object[key] = 2; return object.name + object[key]; }",
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
					constructionObserve(Lower(n, ctx.TypeChecker), ctx.TypeChecker, fmt.Sprintf("probe:%d", index), false)
				})
			}}
		}}
		rule_testing.RunTyped(t, probe, fmt.Sprintf("probe-%d.tsx", index), source)
	}
}

// This adapter snapshots compiler facts, independently of the constructed graph.
// IDs reflect real checker symbol equality; no lexical lookup or HIR-derived binding facts.
func constructionTarget(n *ast.Node) bool {
	if n == nil {
		return false
	}
	if n.Kind == ast.KindParenthesizedExpression {
		return constructionTarget(n.Expression())
	}
	return n.Kind == ast.KindIdentifier || (n.Kind == ast.KindPropertyAccessExpression || n.Kind == ast.KindElementAccessExpression) && constructionExpression(n)
}
func constructionFrame(s string) string {
	return fmt.Sprintf("%d\n%s", len(utf16.Encode([]rune(s))), s)
}
func constructionSymbols(source *ast.SourceFile, c *checker.Checker) string {
	if c == nil {
		return ""
	}
	ids := map[*ast.Symbol]int{}
	var records []string
	var walk func(*ast.Node)
	walk = func(node *ast.Node) {
		if node.Kind == ast.KindIdentifier {
			var symbol *ast.Symbol
			if parent := node.Parent; parent != nil && parent.Kind == ast.KindShorthandPropertyAssignment && parent.Name() == node {
				symbol = c.GetShorthandAssignmentValueSymbol(parent)
			}
			if symbol == nil {
				symbol = c.GetSymbolAtLocation(node)
			}
			if symbol == nil {
				symbol = node.Symbol()
			}
			fields := []string{"1", "symbol", "0", "", "0"}
			if symbol != nil {
				id := ids[symbol]
				if id == 0 {
					id = len(ids) + 1
					ids[symbol] = id
				}
				fields[2], fields[3], fields[4] = fmt.Sprint(id), symbol.Name, fmt.Sprint(len(symbol.Declarations))
				for _, d := range symbol.Declarations {
					same := "0"
					if ast.GetSourceFileOfNode(d) == source {
						same = "1"
					}
					name, property, module := "", "", ""
					if n := d.Name(); n != nil && (n.Kind == ast.KindIdentifier || n.Kind == ast.KindPrivateIdentifier || ast.IsStringLiteralLike(n) || n.Kind == ast.KindNumericLiteral) {
						name = n.Text()
					}
					if n := d.PropertyName(); n != nil && (n.Kind == ast.KindIdentifier || n.Kind == ast.KindPrivateIdentifier || ast.IsStringLiteralLike(n) || n.Kind == ast.KindNumericLiteral) {
						property = n.Text()
					}
					switch d.Kind {
					case ast.KindImportSpecifier, ast.KindImportClause, ast.KindNamespaceImport:
						for parent := d.Parent; parent != nil; parent = parent.Parent {
							if parent.Kind == ast.KindImportDeclaration {
								if m := parent.AsImportDeclaration().ModuleSpecifier; m != nil && ast.IsStringLiteralLike(m) {
									module = m.Text()
								}
								break
							}
						}
					}
					fields = append(fields, strings.TrimPrefix(d.Kind.String(), "Kind"), same, fmt.Sprint(d.Pos()), fmt.Sprint(d.End()), name, property, module)
				}
			}
			var wire strings.Builder
			for _, f := range fields {
				wire.WriteString(constructionFrame(f))
			}
			records = append(records, constructionFrame(fmt.Sprint(node.Pos()))+constructionFrame(fmt.Sprint(node.End()))+constructionFrame(wire.String()))
		}
		node.ForEachChild(func(child *ast.Node) bool { walk(child); return false })
	}
	walk(source.AsNode())
	return constructionFrame(fmt.Sprint(len(records))) + strings.Join(records, "")
}
