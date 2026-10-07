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
	case "ObjectExpression", "Property", "ConditionalExpression", "AssignmentExpression", "SequenceExpression", "UnaryExpression", "UpdateExpression", "BinaryExpression", "LogicalExpression", "MemberExpression", "ArrayExpression", "SpreadElement", "TSNonNullExpression", "ChainExpression":
	case "VariableDeclaration", "VariableDeclarator", "AwaitExpression", "YieldExpression", "TemplateLiteral", "FunctionExpression", "ArrowFunctionExpression", "AssignmentPattern", "RestElement", "BlockStatement", "ExpressionStatement", "ReturnStatement", "ThrowStatement", "EmptyStatement", "DebuggerStatement", "BreakStatement", "ContinueStatement":
	case "CallExpression", "NewExpression", "TaggedTemplateExpression":
		if node.Child("typeArguments") != nil {
			return false
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
				// An object literal extracted from an initializer/return needs its expression context.
				if node.Kind == ast.KindObjectLiteralExpression || node.Kind == ast.KindFunctionExpression {
					fragment = "(" + fragment + ")"
				}
				if node.Kind != ast.KindSpreadElement && node.Kind != ast.KindOmittedExpression && node.Kind != ast.KindPrivateIdentifier && supportedSyntax(node) && coreBoundaries(node, source) && !strings.Contains(fragment, "/*") && !strings.Contains(fragment, "//") && !hasBlankLine(fragment) && !strings.Contains(fragment, "\r") {
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
	for _, item := range []string{"a,b", "a,(b,c)", "(a,b),c", "f((a,b))", "[(a,b),c]", "(a,b).x", "a+(b,c)", "!(a,b)", "(a,b)!", "(a,b)[c]", "a && (b,c)", "(a,b) || c"} {
		add("sequence", item)
	}
	for _, value := range []string{"f?.()", "f?.(x)", "f()", "f(veryLongIdentifierAlpha)", "obj?.x", "obj!.x", "!!x", "++x"} {
		for _, right := range []string{"g(" + value + ")", "g(" + value + ").x"} {
			add("assignment-short-argument-boundary", "veryLongIdentifierAlphaVeryLongIdentifierBetaVeryLongIdentifierGamma="+right)
		}
	}
	for _, body := range []string{
		"const kind=this.node(node.children[offset]??panic('missing method prefix')).kind;",
		"return (this.sequenceBoundaries.has(left)?[left]:this.sequenceParts(left)).concat([this.child(index,2)]);",
		"return (a?b:c).method();", "throw (a?b:c).method();", "const value=(a?b:c).method();",
		"nameIndexToNewNameIndexMap[raw.nameIndex]=newNameIndex=addName(map.names[raw.nameIndex]);",
		"links[id]=graphNode={id,flowNode,edges:[],text:'',lane:-1,endLane:-1,level:-1,circular:false};",
		"memberInfo.memberDescriptorName=descriptorName=createHelperVariable(member,'descriptor');",
		"initializersName=memberInfo.memberInitializersName??=createHelperVariable(member,'initializers');",
		"const x=longIdentifierAlpha=longIdentifierBeta=veryLongFunctionName(veryLongArgumentAlpha,veryLongArgumentBeta);",
	} {
		value := "function named(){" + body + "}"
		for _, context := range []string{"(" + value + ")", "x=" + value, "f(" + value + ")", "({method:" + value + "})"} {
			add("statement-composition-regression", context)
		}
	}
	for _, keyword := range []string{"const", "let", "var", "using"} {
		for _, declarations := range []string{"x=1", "x=a+b", "x=a?b:c", "x={a:1,b:2}", "x=f(a,b,c)", "x=a=b=c", "x=veryLongIdentifierAlpha+veryLongIdentifierBeta+veryLongIdentifierGamma", "x=1,y=2,z=3", "longIdentifierAlpha=foo,longIdentifierBeta=bar", "x=tag`a${b}b`"} {
			body := "{" + keyword + " " + declarations + ";return x;}"
			for _, value := range []string{"x=>" + body, "function named()" + body, "async function named()" + body, "function* named()" + body} {
				for _, context := range []string{"(" + value + ")", "x=" + value, "f(" + value + ")", "({method:" + value + "})"} {
					add("variable-statement-composition", context)
				}
			}
		}
	}
	for _, source := range []string{"(function named(){let x,y,z;})", "(function named(){var veryLongIdentifierAlphaVeryLongIdentifierAlpha,veryLongIdentifierBetaVeryLongIdentifierBeta,veryLongIdentifierGammaVeryLongIdentifierGamma;})", "(async function named(){await using x=f();})", "(function named(){const x=longName(a=b=c);})", "(async function named(){const x=await foo;})", "(function* named(){const x=yield;})"} {
		add("variable-statement-edge", source)
	}
	for _, operand := range []string{"x", "1", "f(x)", "({x:1})", "[1,2]", "a+b", "a?b:c", "veryLongIdentifierAlpha+veryLongIdentifierBeta+veryLongIdentifierGamma", "object.first().second().third()"} {
		for _, body := range []string{"return await " + operand + ";", "return (await " + operand + ").value;", "return (await " + operand + ")(x);", "return await (await " + operand + ").value;", "return !(await " + operand + ");", "return await " + operand + "+x;"} {
			value := "async function named(){" + body + "}"
			for _, context := range []string{"(" + value + ")", "x=" + value, "f(" + value + ")", "({method: " + value + "})"} {
				add("await-composition", context)
			}
		}
		for _, body := range []string{"yield " + operand + ";", "return yield " + operand + ";", "yield* " + operand + ";", "return (yield " + operand + ").value;", "return (yield " + operand + ")+x;"} {
			value := "function* named(){" + body + "}"
			for _, context := range []string{"(" + value + ")", "x=" + value, "f(" + value + ")", "({method: " + value + "})"} {
				add("yield-composition", context)
			}
		}
	}
	for _, source := range []string{"(function* named(){yield;})", "(function* named(){x=yield;})", "(function* named(){x=yield*foo;})", "(function* named(){({longProperty:yield*foo});})", "(async function named(){x=await foo;})", "(async function* named(){yield await x;await (yield x);})", "async x=>await x", "async x=>(await f(x)).value"} {
		add("await-yield-edge", source)
	}
	for _, tag := range []string{"tag", "veryLongTagIdentifierAlphaVeryLongTagIdentifierBeta", "object.tag", "factory().tag", "object[key]", "(a+b)", "(!x)", "(a?b:c)", "(x=>x)", "(function named(){})", "(obj?.tag)", "f()"} {
		for _, template := range []string{"`raw`", "`a${x}b`", "`a\n    ${a+b}\nb`", "`a${(x?.y).z}b`", "`a${veryLongIdentifierAlpha+veryLongIdentifierBeta+veryLongIdentifierGamma}b`", "`a${({method(){return x;}})}b`"} {
			value := tag + template
			for _, context := range []string{value, "x=" + value, "f(" + value + ")", "[" + value + "]", "x=>" + value, "(" + value + ").first().second()", "new (" + value + ")(x)"} {
				add("tagged-template-composition", context)
			}
		}
	}
	for _, expression := range []string{"(x?.y).z", "(fn?.()).x", "(obj?.x)!.y", "(obj?.[key]).value"} {
		for _, template := range []string{"`a${" + expression + "}b`", "`a\n${" + expression + "}b`", "`a${" + expression + "}\nb`"} {
			for _, context := range []string{template, "f(" + template + ")", "value=" + template, "x=>" + template} {
				add("template-optional-boundary-composition", context)
			}
		}
	}
	for _, chain := range []string{"obj?.x", "obj?.[key]", "obj?.method()", "fn?.()", "obj?.x.y", "obj?.x!.y", "obj.method?.(x)", "veryLongIdentifierAlpha?.veryLongPropertyNameBeta.veryLongPropertyNameGamma"} {
		for _, suffix := range []string{".x", "[key]", "()", "(x)", "!.x", "!()", "?.x", "?.(x)", ".first().second().third()", ".veryLongPropertyNameAlpha.veryLongPropertyNameBeta.veryLongPropertyNameGamma"} {
			value := "(" + chain + ")" + suffix
			for _, context := range []string{value, "x=" + value, "f(" + value + ")", "[" + value + "]", "x=>" + value, "((" + value + ")).tail"} {
				add("optional-chain-boundary", context)
			}
		}
	}
	for _, prefix := range []string{"function named", "async function named", "function* named", "async function* named"} {
		for _, parameters := range []string{"()", "(x)", "(x,y)", "(x=1)", "(...items)", "(veryLongParameterNameAlpha,veryLongParameterNameBeta,veryLongParameterNameGamma)"} {
			for _, body := range []string{"{}", "{return x;}", "{f(x);return x;}", "{'use strict';f();'later';}", "{('use strict');return x;}", "{return a?b:c;}", "{return (a,b);}", "{return veryLongIdentifierAlpha+veryLongIdentifierBeta+veryLongIdentifierGamma+veryLongIdentifierDelta;}", "{debugger;throw new Error('failure');}", "{return ({x:1}).x;}"} {
				value := prefix + parameters + body
				for _, context := range []string{"(" + value + ")", "x=" + value, "f(" + value + ")", "f(x," + value + ")", "f(" + value + ",x)", "[" + value + "]", "({key:" + value + "})", "(" + value + ")(x)", "(" + value + ").x", "new (" + value + ")(x)"} {
					add("named-function-composition", context)
				}
			}
		}
	}
	for _, name := range []string{"method", "'escaped-name'", "'é'", "123", "[key]", "[a+b]", "async", "get"} {
		for _, prefix := range []string{"", "async ", "*", "async *"} {
			for _, parameters := range []string{"()", "(x)", "(x=1)", "(...items)", "(veryLongParameterNameAlpha,veryLongParameterNameBeta,veryLongParameterNameGamma)"} {
				for _, body := range []string{"{}", "{return x;}", "{return a?b:c;}", "{f(x);return ({x:1}).x;}"} {
					object := "{" + prefix + name + parameters + body + ",value:1}"
					for _, context := range []string{"(" + object + ")", "f(" + object + ")", "x=" + object, "x=>" + object} {
						if strings.HasPrefix(context, "x=>{") {
							context = "x=>(" + object + ")"
						}
						add("object-method-composition", context)
					}
				}
			}
		}
		for _, member := range []string{"get " + name + "(){return x;}", "set " + name + "(x){f(x);}"} {
			object := "{" + member + ",value:1}"
			for _, context := range []string{"(" + object + ")", "f(" + object + ")", "x=" + object, "x=>(" + object + ")"} {
				add("object-accessor-composition", context)
			}
		}
	}
	for _, key := range []string{"[x]", "[1]", "[-1]", "['a']", "[true]", "[null]", "[long]", "[longer]"} {
		for _, value := range []string{"'" + strings.Repeat("longText", 20) + "'", "object.veryLongPropertyNameAlpha.veryLongPropertyNameBeta.veryLongPropertyNameGamma"} {
			object := "{" + key + ":" + value + "}"
			for _, context := range []string{"(" + object + ")", "f(" + object + ")", "x=" + object, "x=>(" + object + ")"} {
				add("computed-short-key-boundary", context)
			}
		}
	}
	arrowBodies := []string{"a=b", "({x:1}).x", "({x:1})?a:b", "{return a?b:c;}", "{return a=b;}", "{throw a=b;}", "{return (a,b);}", "{throw new Error('failure');}", "{debugger;return x;}", "{return (veryLongIdentifierAlpha,veryLongIdentifierBeta,veryLongIdentifierGamma,veryLongIdentifierDelta);}", "(veryLongIdentifierAlpha,veryLongIdentifierBeta,veryLongIdentifierGamma,veryLongIdentifierDelta)", "{('use strict');f();'other';}", "{'use strict';'use asm';f();'later';}", "{({x:1}).x;return x;}", "x", "x+1", "a?b:c", "(a,b)", "({x:1,y:2})", "[a,b,c]", "f(x)", "`a${x}b`", "`a\nb`", "{return x;}", "{return veryLongIdentifierAlpha + veryLongIdentifierBeta + veryLongIdentifierGamma + veryLongIdentifierDelta;}", "{f(x);return x;}", "{}"}
	for _, parameters := range []string{"x", "()", "(x,y)", "(x=1)", "(...items)", "(veryLongParameterNameAlpha,veryLongParameterNameBeta,veryLongParameterNameGamma)"} {
		for _, body := range arrowBodies {
			for _, prefix := range []string{"", "async "} {
				arrow := prefix + parameters + "=>" + body
				for _, value := range []string{arrow, "x=" + arrow, "f(" + arrow + ")", "f(x," + arrow + ")", "[" + arrow + "]", "(" + arrow + ")(x)", "(" + arrow + ").x", "({key:" + arrow + "})"} {
					add("arrow-composition", value)
				}
			}
		}
	}
	for length := 1; length <= 20; length++ {
		arrow := strings.Repeat("(veryLongParameterNameAlpha,veryLongParameterNameBeta)=>", length) + "veryLongIdentifierAlpha + veryLongIdentifierBeta + veryLongIdentifierGamma"
		for _, value := range []string{arrow, "x=" + arrow, "f(" + arrow + ")", "(" + arrow + ")()"} {
			add("arrow-chain", value)
		}
	}
	templateValues := []string{"x", "a+b+c", "a?b:c", "a??b", "(a,b)", "({x:1})", "{\nx:1,y:2\n}", "[a,b,c]", "f(x)", "obj.x", "obj?.x", "veryLongIdentifierAlpha + veryLongIdentifierBeta + veryLongIdentifierGamma + veryLongIdentifierDelta"}
	quasis := []string{"", "raw", "é😀", "\\u0061", "\\`escaped", "start\n  ", "start\n\t", "start\n"}
	for _, value := range templateValues {
		for _, quasi := range quasis {
			for _, interpolation := range []string{value, "\n" + value + "\n", value + "\n"} {
				literal := "`" + quasi + "${" + interpolation + "}tail`"
				for _, source := range []string{literal, "f(" + literal + ")", "x=" + literal, "[" + literal + "]", "(" + literal + ").x"} {
					add("template-interpolation", source)
				}
			}
		}
	}
	for length := 1; length <= 30; length++ {
		literal := "`head"
		for index := 0; index < length; index++ {
			literal += fmt.Sprintf("${value%d} middle", index)
		}
		literal += "`"
		for _, source := range []string{literal, "f(" + literal + ")", "x=" + literal} {
			add("template-width", source)
		}
	}
	for _, source := range []string{"`a\nb`", "`a\n  b\n c`", "`a\\\nb`", "`a${`b${x}c`}d`", "`a${[1,2,3]}b`", "f(`a\nb`)", "f(\n`a\nb`\n)", "x=`a\n  ${veryLongIdentifierAlpha + veryLongIdentifierBeta + veryLongIdentifierGamma}\n  z`"} {
		add("template-boundary", source)
	}
	for _, source := range []string{"Number()", "Number('17')", "Number('')", "Number('0x11')", "Number('0b101')", "Number('17tail')", "Number(true)", "Number(null)", "Number(undefined)", "Number(-0)"} {
		add("number-constructor", source)
	}
	memberBases := []string{"obj", "this", "Factory", "_", "$$", "longIdentifierName", "namespace.Factory", "f()", "f(x)", "[a,b]", "({a:1})", "(a+b)", "(a?b:c)", "1", "new C()"}
	memberArguments := []string{"", "x", "x,y", "veryLongIdentifierAlpha,veryLongIdentifierBeta,veryLongIdentifierGamma", "{x:1,y:2}", "[1,2,3]", "[x,y,z]", "f(x)", "a+b"}
	for _, base := range memberBases {
		for _, arguments := range memberArguments {
			for _, suffix := range []string{".method(" + arguments + ")", "[key](" + arguments + ")", ".method?.(" + arguments + ")", ".first().second(" + arguments + ").third()", ".first(" + arguments + ")[0].second().third()", ".first!.second(" + arguments + ")"} {
				value := "(" + base + ")" + suffix
				// A parenthesized optional chain remains a separate proving gap.
				for _, source := range []string{value, "x=" + value, "[" + value + "]"} {
					add("member-call-composition", source)
				}
			}
		}
	}
	for length := 1; length <= 30; length++ {
		value := "Factory.create()"
		for index := 0; index < length; index++ {
			value += fmt.Sprintf(".method%d(argument%d)", index, index)
		}
		for _, source := range []string{value, "x=" + value, "f(" + value + ")"} {
			add("member-call-width", source)
		}
	}
	for _, source := range []string{"(PRETTIER_HTML_PLACEHOLDER_0_0_IN_JS)", "f((PRETTIER_HTML_PLACEHOLDER_0_0_IN_JS))", "(PRETTIER_HTML_PLACEHOLDER_x_0_IN_JS)", "f()()", "f(a,b,c)(x)", "f(a,b,c)(x)(y)", "new (f().C)(a,b)", "new (f().C.x)(a,b)", "(let[x])()", "(1).toString()", "({x:1}).method()", "this.a().b().c()", "obj.a().b().c().d()"} {
		add("member-call-boundary", source)
	}
	callValues := []string{"[]", "[1,2]", "[a,b]", "[[1,2],[3,4]]", "[1,,]", "[...items]", "{}", "{a:1}", "{a:x,b:y}", "{\na:x,\nb:y\n}", "{...items}", "veryLongIdentifierAlpha + veryLongIdentifierBeta + veryLongIdentifierGamma"}
	for _, left := range callValues {
		for _, right := range callValues {
			for _, arguments := range []string{left, left + "," + right, "x," + left + "," + right, left + ",'text'"} {
				for _, source := range []string{"f(" + arguments + ")", "f?.(" + arguments + ")", "new C(" + arguments + ")", "x=f(" + arguments + ")"} {
					add("expanded-arguments", source)
				}
			}
		}
	}
	for _, name := range []string{"f", "require", "define", "new C"} {
		for length := 1; length <= 40; length++ {
			arguments := []string{}
			for index := 0; index < length; index++ {
				arguments = append(arguments, fmt.Sprintf("argumentName%d", index))
			}
			add("argument-width", name+"("+strings.Join(arguments, ",")+")")
			add("argument-width", name+"({"+strings.Join(arguments, ",")+"})")
			add("argument-width", name+"(["+strings.Join(arguments, ",")+"])")
		}
	}
	objectValues := []string{"x?.y", "x", "'text'", "a+b+c", "a?b:c", "a??b", "[1,2]", "({nested:1})", "f(x)", "veryLongIdentifierAlpha + veryLongIdentifierBeta + veryLongIdentifierGamma"}
	objectKeys := []string{"x", "longPropertyNameAlphaLongPropertyNameBeta", "'x'", "'with space'", "'é'", "'\\u0061'", "0xAF", "[a+b]", "[a?b:c]"}
	for _, key := range objectKeys {
		for _, value := range objectValues {
			for _, object := range []string{"{" + key + ":" + value + "}", "{\n" + key + ":" + value + ",\nother:x\n}", "{...source," + key + ":" + value + ",shorthand}"} {
				for _, source := range []string{"(" + object + ")", "[" + object + "]", "x=" + object, "(" + object + ").x", "a?" + object + ":fallback", "(" + object + ")+x"} {
					add("object-composition", source)
				}
			}
		}
	}
	for length := 1; length <= 30; length++ {
		properties := []string{}
		for index := 0; index < length; index++ {
			properties = append(properties, fmt.Sprintf("property%d: value%d", index, index))
		}
		for _, source := range []string{"({" + strings.Join(properties, ",") + "})", "({\n" + strings.Join(properties, ",\n") + "\n})"} {
			add("object-width", source)
		}
	}
	conditionalParts := []string{"a", "a+b+c", "a||b||c", "a&&[1,2]", "[1,2]", "f(x)", "obj.x", "(a=b)", "(a,b)", "veryLongIdentifierAlpha + veryLongIdentifierBeta + veryLongIdentifierGamma + veryLongIdentifierDelta"}
	for _, test := range conditionalParts {
		for _, yes := range conditionalParts {
			for _, no := range conditionalParts {
				value := "(" + test + ")?(" + yes + "):(" + no + ")"
				for _, source := range []string{value, "f(" + value + ")", "x=" + value, "(" + value + ").x", "!(" + value + ")"} {
					add("conditional-composition", source)
				}
			}
		}
	}
	for length := 1; length <= 20; length++ {
		alternate := "fallback"
		consequent := "fallback"
		test := "fallback"
		for index := 0; index < length; index++ {
			alternate = "condition" + fmt.Sprint(index) + "?value" + fmt.Sprint(index) + ":" + alternate
			consequent = "condition" + fmt.Sprint(index) + "?(" + consequent + "):value" + fmt.Sprint(index)
			test = "(" + test + ")?value" + fmt.Sprint(index) + ":fallback"
		}
		for _, value := range []string{alternate, consequent, test} {
			for _, source := range []string{value, "f(" + value + ")", "x=(" + value + ").x", "!(" + value + ").x"} {
				add("conditional-nesting", source)
			}
		}
	}
	assignmentOps := []string{"=", "+=", "-=", "*=", "/=", "%=", "**=", "<<=", ">>=", ">>>=", "&=", "|=", "^=", "&&=", "||=", "??="}
	values := []string{"x", "1", "true", "'text'", "`template`", "[1,2]", "a+b+c", "a&&(b||c)", "obj.member", "obj.alpha.beta", "f()", "f(x)", "f(longIdentifierArgumentAlpha, longIdentifierArgumentBeta, longIdentifierArgumentGamma)", "veryLongIdentifierAlpha + veryLongIdentifierBeta + veryLongIdentifierGamma + veryLongIdentifierDelta", "(a,b,c)"}
	for _, op := range assignmentOps {
		for _, left := range []string{"a", "obj.x", "obj.alpha.beta", "obj[index+offset]", "veryLongIdentifierAlpha.veryLongPropertyNameBeta.veryLongPropertyNameGamma"} {
			for _, right := range values {
				value := left + op + right
				for _, source := range []string{value, "f((" + value + "))", "[(" + value + ")]", "!(" + value + ")", "(" + value + ").x"} {
					add("assignment-composition", source)
				}
			}
		}
	}
	for length := 2; length <= 20; length++ {
		parts := []string{}
		for index := 0; index < length; index++ {
			parts = append(parts, fmt.Sprintf("assignmentIdentifier%d", index))
		}
		for _, right := range values {
			value := strings.Join(parts, "=") + "=" + right
			for _, source := range []string{value, "f((" + value + "))", "(" + value + "),x"} {
				add("assignment-chain", source)
			}
		}
	}
	sequenceItems := []string{"a", "a+b+c", "a||b||c", "a*b", "[a,b]", "f(a,b)", "veryLongIdentifierAlpha + veryLongIdentifierBeta + veryLongIdentifierGamma + veryLongIdentifierDelta", "veryLongIdentifierAlpha && veryLongIdentifierBeta && veryLongIdentifierGamma && veryLongIdentifierDelta"}
	for _, left := range sequenceItems {
		for _, right := range sequenceItems {
			for _, value := range []string{left + "," + right, "(" + left + ",b)," + right, left + ",(" + right + ",c)"} {
				for _, source := range []string{value, "f((" + value + "))", "[(" + value + ")]", "(" + value + ").x", "!(" + value + ")"} {
					add("sequence-composition", source)
				}
			}
		}
	}
	for length := 1; length <= 60; length++ {
		parts := []string{}
		for index := 0; index < length; index++ {
			parts = append(parts, fmt.Sprintf("sequenceArgument%d", index))
		}
		value := strings.Join(parts, ",")
		for _, source := range []string{value, "f((" + value + "))", "[(" + value + ")]", "(" + value + ").x"} {
			add("sequence-width", source)
		}
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

// Literal continuation and multiline-template layout is outside the expression core.
func coreBoundaries(node *ast.Node, source string) bool {
	if node.Kind == ast.KindStringLiteral {
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
	case ast.KindObjectLiteralExpression, ast.KindPropertyAssignment, ast.KindComputedPropertyName, ast.KindSpreadAssignment:
		valid := true
		node.ForEachChild(func(child *ast.Node) bool {
			if !supportedSyntax(child) {
				valid = false
			}
			return false
		})
		return valid
	case ast.KindShorthandPropertyAssignment:
		return node.AsShorthandPropertyAssignment().ObjectAssignmentInitializer == nil
	case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
		if node.Parent == nil || node.Parent.Kind != ast.KindObjectLiteralExpression || node.Type() != nil || node.TypeParameterList() != nil {
			return false
		}
		if node.Kind == ast.KindMethodDeclaration && node.AsMethodDeclaration().QuestionToken() != nil {
			return false
		}
		valid := true
		node.ForEachChild(func(child *ast.Node) bool {
			if child.Kind != ast.KindAsteriskToken && child.Kind != ast.KindAsyncKeyword && !supportedSyntax(child) {
				valid = false
			}
			return false
		})
		return valid
	case ast.KindFunctionExpression:
		item := node.AsFunctionExpression()
		if node.Name() == nil || item.Type != nil || item.TypeParameters != nil {
			return false
		}
		valid := true
		node.ForEachChild(func(child *ast.Node) bool {
			if child.Kind != ast.KindAsteriskToken && child.Kind != ast.KindAsyncKeyword && !supportedSyntax(child) {
				valid = false
			}
			return false
		})
		return valid
	case ast.KindArrowFunction:
		if node.AsArrowFunction().Type != nil || node.AsArrowFunction().TypeParameters != nil {
			return false
		}
		valid := true
		node.ForEachChild(func(child *ast.Node) bool {
			if child.Kind != ast.KindEqualsGreaterThanToken && child.Kind != ast.KindAsyncKeyword && !supportedSyntax(child) {
				valid = false
			}
			return false
		})
		return valid
	case ast.KindParameter:
		item := node.AsParameterDeclaration()
		return item.Name().Kind == ast.KindIdentifier && item.Type == nil && item.QuestionToken == nil && supportedSyntax(item.Initializer)
	case ast.KindVariableStatement:
		return node.Modifiers() == nil && supportedSyntax(node.AsVariableStatement().DeclarationList)
	case ast.KindVariableDeclarationList:
		for _, declaration := range node.AsVariableDeclarationList().Declarations.Nodes {
			if !supportedSyntax(declaration) {
				return false
			}
		}
		return true
	case ast.KindVariableDeclaration:
		item := node.AsVariableDeclaration()
		return node.Name().Kind == ast.KindIdentifier && item.Type == nil && item.ExclamationToken == nil && supportedSyntax(item.Initializer)
	case ast.KindBlock, ast.KindExpressionStatement, ast.KindReturnStatement, ast.KindThrowStatement, ast.KindBreakStatement, ast.KindContinueStatement:
		valid := true
		node.ForEachChild(func(child *ast.Node) bool {
			if !supportedSyntax(child) {
				valid = false
			}
			return false
		})
		return valid
	case ast.KindEmptyStatement, ast.KindDebuggerStatement:
		return true
	case ast.KindAwaitExpression, ast.KindYieldExpression:
		return supportedSyntax(node.Expression())
	case ast.KindTaggedTemplateExpression:
		item := node.AsTaggedTemplateExpression()
		return item.TypeArguments == nil && !jestSyntaxTag(item.Tag) && supportedSyntax(item.Tag) && supportedSyntax(item.Template)
	case ast.KindTemplateExpression:
		for _, span := range node.AsTemplateExpression().TemplateSpans.Nodes {
			if !supportedSyntax(span.AsTemplateSpan().Expression) {
				return false
			}
		}
		return true
	case ast.KindConditionalExpression:
		item := node.AsConditionalExpression()
		return supportedSyntax(item.Condition) && supportedSyntax(item.WhenTrue) && supportedSyntax(item.WhenFalse)
	case ast.KindPrefixUnaryExpression:
		return supportedSyntax(node.AsPrefixUnaryExpression().Operand)
	case ast.KindPostfixUnaryExpression:
		return supportedSyntax(node.AsPostfixUnaryExpression().Operand)
	case ast.KindBinaryExpression:
		item := node.AsBinaryExpression()
		if item.OperatorToken.Kind >= ast.KindFirstAssignment && item.OperatorToken.Kind <= ast.KindLastAssignment {
			left := item.Left
			for left.Kind == ast.KindParenthesizedExpression {
				left = left.Expression()
			}
			switch left.Kind {
			case ast.KindIdentifier, ast.KindPropertyAccessExpression, ast.KindElementAccessExpression, ast.KindNonNullExpression:
			default:
				return false
			}
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
		if !supportedSyntax(node.Expression()) || len(node.TypeArguments()) > 0 {
			return false
		}
		for _, child := range node.Arguments() {
			if !supportedSyntax(child) {
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

func jestSyntaxTag(tag *ast.Node) bool {
	for tag.Kind == ast.KindParenthesizedExpression {
		tag = tag.Expression()
	}
	property := func(node *ast.Node) *ast.Node {
		switch node.Kind {
		case ast.KindPropertyAccessExpression:
			return node.Name()
		case ast.KindElementAccessExpression:
			return node.AsElementAccessExpression().ArgumentExpression
		default:
			return nil
		}
	}
	key := property(tag)
	if key == nil || key.Kind != ast.KindIdentifier || key.Text() != "each" {
		return false
	}
	object := tag.Expression()
	for object.Kind == ast.KindParenthesizedExpression {
		object = object.Expression()
	}
	if key := property(object); key != nil {
		if key.Kind != ast.KindIdentifier || (key.Text() != "only" && key.Text() != "skip") {
			return false
		}
		object = object.Expression()
	}
	if object.Kind != ast.KindIdentifier {
		return false
	}
	switch object.Text() {
	case "describe", "it", "test", "fdescribe", "fit", "ftest", "xdescribe", "xit", "xtest":
		return true
	}
	return false
}
