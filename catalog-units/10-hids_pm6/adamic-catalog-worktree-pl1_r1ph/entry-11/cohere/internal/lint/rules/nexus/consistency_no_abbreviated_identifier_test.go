package nexus

import (
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/system-inc/cohere/internal/lint/testing"
)

const abbreviatedFile = "/repository/source/Thing.tsx"

func TestConsistencyNoAbbreviatedIdentifierFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		// Whole-word abbreviations.
		{"prop", "const prop = 1;\n", []string{"abbreviatedIdentifier"}},
		{"props", "const props = 1;\n", []string{"abbreviatedIdentifier"}},
		{"param", "const param = 1;\n", []string{"abbreviatedIdentifier"}},
		{"params", "const params = 1;\n", []string{"abbreviatedIdentifier"}},
		{"ref", "const ref = 1;\n", []string{"abbreviatedIdentifier"}},
		{"config", "const config = 1;\n", []string{"abbreviatedIdentifier"}},
		{"idx", "const idx = 0;\n", []string{"abbreviatedIdentifier"}},
		{"arg", "const arg = 1;\n", []string{"abbreviatedIdentifier"}},
		{"args", "const args = 1;\n", []string{"abbreviatedIdentifier"}},
		{"acc", "const acc = 0;\n", []string{"abbreviatedIdentifier"}},
		{"char", "const char = 'a';\n", []string{"abbreviatedIdentifier"}},
		{"fn", "const fn = () => 1;\n", []string{"abbreviatedIdentifier"}},
		{"str", "const str = 'a';\n", []string{"abbreviatedIdentifier"}},
		{"val", "const val = 1;\n", []string{"abbreviatedIdentifier"}},
		{"arr", "const arr = [];\n", []string{"abbreviatedIdentifier"}},
		{"obj", "const obj = {};\n", []string{"abbreviatedIdentifier"}},
		{"num", "const num = 1;\n", []string{"abbreviatedIdentifier"}},
		{"res", "const res = 1;\n", []string{"abbreviatedIdentifier"}},
		{"err", "const err = 1;\n", []string{"abbreviatedIdentifier"}},
		{"req", "const req = 1;\n", []string{"abbreviatedIdentifier"}},
		{"msg", "const msg = 1;\n", []string{"abbreviatedIdentifier"}},
		{"min", "const min = 1;\n", []string{"abbreviatedIdentifier"}},
		{"max", "const max = 1;\n", []string{"abbreviatedIdentifier"}},
		{"ctx", "const ctx = 1;\n", []string{"abbreviatedIdentifier"}},
		{"db", "const db = 1;\n", []string{"abbreviatedIdentifier"}},
		{"tx", "const tx = 1;\n", []string{"abbreviatedIdentifier"}},
		{"opts", "const opts = {};\n", []string{"abbreviatedIdentifier"}},
		{"cur", "const cur = 1;\n", []string{"abbreviatedIdentifier"}},
		{"pct", "const pct = 1;\n", []string{"abbreviatedIdentifier"}},
		{"prev", "const prev = 1;\n", []string{"abbreviatedIdentifier"}},

		// camelCase prefixes.
		{"ctx prefix", "const ctxValue = 1;\n", []string{"abbreviatedIdentifier"}},
		{"db prefix", "const dbCode = 1;\n", []string{"abbreviatedIdentifier"}},
		{"tx prefix", "const txHash = 1;\n", []string{"abbreviatedIdentifier"}},
		{"opts prefix", "const optsForRun = 1;\n", []string{"abbreviatedIdentifier"}},
		{"cur prefix", "const curStep = 1;\n", []string{"abbreviatedIdentifier"}},
		{"pct prefix", "const pctComplete = 1;\n", []string{"abbreviatedIdentifier"}},
		{"prev prefix", "const prevStep = 1;\n", []string{"abbreviatedIdentifier"}},
		{"idx prefix", "const idxStart = 0;\n", []string{"abbreviatedIdentifier"}},
		{"config prefix", "const configValue = 1;\n", []string{"abbreviatedIdentifier"}},
		{"prop prefix", "const propName = 1;\n", []string{"abbreviatedIdentifier"}},
		{"params prefix", "const paramsText = 1;\n", []string{"abbreviatedIdentifier"}},
		{"ref prefix", "const refCount = 1;\n", []string{"abbreviatedIdentifier"}},
		{"arg prefix", "const argCount = 1;\n", []string{"abbreviatedIdentifier"}},
		{"args prefix", "const argsList = 1;\n", []string{"abbreviatedIdentifier"}},
		{"char prefix", "const charSet = 1;\n", []string{"abbreviatedIdentifier"}},
		{"fn prefix", "const fnCache = 1;\n", []string{"abbreviatedIdentifier"}},
		{"str prefix", "const strValue = 1;\n", []string{"abbreviatedIdentifier"}},
		{"max prefix", "const maxAge = 1;\n", []string{"abbreviatedIdentifier"}},
		{"err prefix", "const errCount = 1;\n", []string{"abbreviatedIdentifier"}},

		// Suffixes.
		{"Prop suffix", "const rowProp = 1;\n", []string{"abbreviatedSuffix"}},
		{"Props suffix", "const rowProps = 1;\n", []string{"abbreviatedSuffix"}},
		{"Param suffix", "const queryParam = 1;\n", []string{"abbreviatedSuffix"}},
		{"Params suffix", "const searchParams = 1;\n", []string{"abbreviatedSuffix"}},
		{"Ref suffix", "const nodeRef = 1;\n", []string{"abbreviatedSuffix"}},
		{"Config suffix", "const buildConfig = 1;\n", []string{"abbreviatedSuffix"}},
		{"Idx suffix", "const rowIdx = 1;\n", []string{"abbreviatedSuffix"}},
		{"Arg suffix", "const firstArg = 1;\n", []string{"abbreviatedSuffix"}},
		{"Args suffix", "const esBuildArgs = 1;\n", []string{"abbreviatedSuffix"}},
		{"Char suffix", "const lastChar = 1;\n", []string{"abbreviatedSuffix"}},
		{"Fn suffix", "const compareFn = 1;\n", []string{"abbreviatedSuffix"}},
		{"Str suffix", "const queryStr = 1;\n", []string{"abbreviatedSuffix"}},
		{"Val suffix", "const inputVal = 1;\n", []string{"abbreviatedSuffix"}},
		{"Arr suffix", "const itemArr = 1;\n", []string{"abbreviatedSuffix"}},
		{"Obj suffix", "const targetObj = 1;\n", []string{"abbreviatedSuffix"}},
		{"Num suffix", "const pageNum = 1;\n", []string{"abbreviatedSuffix"}},
		{"Res suffix", "const fetchRes = 1;\n", []string{"abbreviatedSuffix"}},
		{"Err suffix", "const parseErr = 1;\n", []string{"abbreviatedSuffix"}},
		{"Req suffix", "const httpReq = 1;\n", []string{"abbreviatedSuffix"}},
		{"Msg suffix", "const errorMsg = 1;\n", []string{"abbreviatedSuffix"}},
		{"Min suffix", "const pageMin = 1;\n", []string{"abbreviatedSuffix"}},
		{"Max suffix", "const pageMax = 1;\n", []string{"abbreviatedSuffix"}},

		// Milliseconds, matched wherever the word sits rather than only at the end. Anchoring to
		// the end is how `maximumAgeMsHalf` sat unreported while every sibling was caught.
		{"Ms at the end", "const maximumAgeMs = 1;\n", []string{"millisecondSuffix"}},
		{"Ms in the middle", "const maximumAgeMsHalf = 1;\n", []string{"millisecondSuffix"}},

		// Word segments, matched wherever they sit. The anchored branches miss these.
		{"Cwd segment", "const originalInitCwd = 1;\n", []string{"abbreviatedWordSegment"}},
		{"Dir segment", "const outputDirPath = 1;\n", []string{"abbreviatedWordSegment"}},
		{"Env segment", "const buildEnvName = 1;\n", []string{"abbreviatedWordSegment"}},
		{"Cli segment", "const runCliCommand = 1;\n", []string{"abbreviatedWordSegment"}},
		{"Len segment", "const bufferLen = 1;\n", []string{"abbreviatedWordSegment"}},
		{"Seq segment", "const eventSeq = 1;\n", []string{"abbreviatedWordSegment"}},
		{"Db segment", "const primaryDbHost = 1;\n", []string{"abbreviatedWordSegment"}},
		{"Tx segment", "const pendingTxHash = 1;\n", []string{"abbreviatedWordSegment"}},
		{"Vars segment", "const themeVars = 1;\n", []string{"abbreviatedWordSegment"}},
		{"Var segment", "const themeVarName = 1;\n", []string{"abbreviatedWordSegment"}},

		// A bare `queryFn` variable still fires; only the object-literal key is exempt. It lands on
		// the `Fn` suffix branch rather than the `fn` prefix one, which is the order the original
		// runs them in and the message a reader of `queryFn` needs: the repair is a role name.
		{"bare queryFn", "const queryFn = () => 1;\n", []string{"abbreviatedSuffix"}},
		{"bare mutationFn", "const mutationFn = () => 1;\n", []string{"abbreviatedSuffix"}},

		// A destructure binds a name this file owns, unlike an object-literal key.
		{"destructured binding", "const { props } = source;\n", []string{"abbreviatedIdentifier"}},

		// `this.props` is ours: the declaration is a class property no member check protects, so a
		// blind skip renames the declaration and leaves every read pointing at the old name.
		{"this member read", "class Thing {\n    read() {\n        return this.props;\n    }\n}\n", []string{"abbreviatedIdentifier"}},

		// Every reference reports, not just the declaration.
		{"declaration and reference", "const idx = 0;\nuse(idx);\n", []string{"abbreviatedIdentifier", "abbreviatedIdentifier"}},

		// `Msg` continues in lowercase, so it is not the millisecond unit. It is still the `msg`
		// prefix, which is a different branch with a different message.
		{"msgText is the msg prefix, not the Ms unit", "const msgText = 1;\n", []string{"abbreviatedIdentifier"}},

		// A plain parameter or catch binding named `args` reports where it is declared and at every
		// use. Once the file binds one, the rest-only exemption is off for the whole file, since
		// without scope analysis a use cannot be traced to its binding (#e000k8d): the rest's own
		// declaration stays exempt and its use reports.
		{"a plain args parameter", "function run(args: string[]) {\n    use(args);\n}\n", []string{"abbreviatedIdentifier", "abbreviatedIdentifier"}},
		{"a catch binding named args", "try {\n    run();\n} catch (args) {\n    use(args);\n}\n", []string{"abbreviatedIdentifier", "abbreviatedIdentifier"}},
		// A file that binds no `args` at all has no rest to agree with, so a free reference reports.
		{"an args the file never binds", "use(args);\n", []string{"abbreviatedIdentifier"}},
		{"a rest beside a plain args", "function run(...args: string[]) {\n    use(args);\n}\nfunction other(args: string[]) {\n    return args;\n}\n", []string{"abbreviatedIdentifier", "abbreviatedIdentifier", "abbreviatedIdentifier"}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, ConsistencyNoAbbreviatedIdentifier, abbreviatedFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

func TestConsistencyNoAbbreviatedIdentifierStaysSilent(t *testing.T) {
	t.Parallel()

	// The exemptions are the rule. Without them it demands renames the author cannot perform, which
	// is how a naming rule gets switched off.
	cases := []struct {
		name       string
		sourceText string
	}{
		// Already spelled out. Each of these shares a prefix or suffix with an abbreviation, which
		// is exactly what the already-spelled guards exist for.
		{"spelled out names", "const properties = 1;\nconst parameters = 2;\nconst configuration = 3;\n"},
		{"property and parameter", "const property = 1;\nconst parameter = 2;\n"},
		{"context, database, transaction", "const context = 1;\nconst database = 2;\nconst transaction = 3;\n"},
		{"current and previous", "const current = 1;\nconst previous = 2;\n"},
		{"options and percent", "const options = 1;\nconst percent = 2;\n"},
		{"index and character", "const index = 0;\nconst character = 'a';\n"},
		{"value, array, object, number", "const value = 1;\nconst array = [];\nconst object = {};\nconst number = 1;\n"},
		{"response, error, request, message", "const response = 1;\nconst error = 2;\nconst request = 3;\nconst message = 4;\n"},
		{"minimum and maximum", "const minimum = 1;\nconst maximum = 2;\n"},
		{"reference and string", "const reference = 1;\nconst string = 'a';\n"},
		{"spelled suffixes", "const rowProperty = 1;\nconst inputValue = 2;\nconst pageNumber = 3;\n"},
		{"spelled Argument suffix", "const firstArgument = 1;\nconst commandArguments = [];\n"},
		{"spelled Character suffix", "const lastCharacter = 'a';\n"},
		{"spelled Function suffix", "const compareFunction = () => 1;\n"},
		{"spelled String suffix", "const queryString = 'a';\n"},
		{"spelled milliseconds", "const durationInMilliseconds = 500;\n"},
		{"spelled word segments", "const workingDirectory = 1;\nconst environment = 2;\nconst length = 3;\n"},

		// Words that merely contain an abbreviation's letters.
		{"Direction is not Dir", "const scrollDirection = 1;\n"},
		{"Environment is not Env", "const buildEnvironment = 1;\n"},
		{"SMS is not Ms", "const sendSMS = 1;\n"},
		{"LLMs is not Ms", "const availableLLMs = 1;\n"},
		{"InnoDb is a proper noun", "const InnoDbMaximumIndexKeyBytes = 1;\n"},
		{"URLSearchParams is a Web API", "const parsed = new URLSearchParams();\n"},

		// Names somebody else chose. Renaming any of these changes what the code reaches for.
		{"a property read", "const value = thing.props;\n"},
		{"a chained property read", "const value = thing.config.idx;\n"},
		{"an object literal key", "const options = { config: 1, idx: 2 };\n"},
		{"an import specifier", "import { config } from 'external';\n"},
		{"a namespace import", "import * as config from 'external';\n"},
		{"a default import", "import params from 'external';\n"},
		// A type property signature is NOT here, and it used to be. The original reports one: its
		// `reportWithTypeKeyGuard` withholds the AUTOFIX and calls `context.report` either way, on
		// the reasoning that a blind rename touches the declaration while every caller keeps the old
		// spelling through the object-literal-key skip. This port ships no fixer, so the guard has
		// nothing to withhold and collapses to a plain report.
		//
		// Measured rather than argued: `modules/pensieve/PensieveBootstrap.ts` lines 40, 41, 562,
		// 662 and 663 are `maxAgentsBytes` and `maxBootBytes` written as interface members, and a
		// full eslint run over the ahra tree reports all five as `Identifier "max" should not be
		// abbreviated`. This port reported none, which was that rule's entire parity gap.
		//
		// See `TestConsistencyNoAbbreviatedIdentifierJudgesTypeMemberKeys` for the positive case.
		{"a qualified type name member", "type Alias = Namespace.Config;\n"},
		{"a type reference", "let value: SomeConfig = load();\n"},

		// JSX. A tag is named by HTML or by the component's author, an attribute by whatever
		// component declares it. Both are plain identifiers in typescript-go and a distinct node
		// type in the TypeScript parser, so the original never needed these and this port does.
		// Their absence cost the sibling rule 3,081 false findings on the real tree.
		{"a jsx element named for a component", "export const Thing = () => <Props />;\n"},
		{"a paired jsx element", "export const Thing = () => <Config>hi</Config>;\n"},
		{"a jsx attribute name", "export const Thing = () => <div config=\"1\" idx=\"2\" />;\n"},
		{"a hyphenated jsx attribute", "export const Thing = () => <div data-config=\"1\" />;\n"},

		// The conventional shapes.
		{"a rest parameter declaration", "function run(...args: string[]) {\n    return 1;\n}\n"},
		{"a rest binding element", "const { first, ...args } = source;\n"},
		// The uses agree with the rest that named them, when every `args` the file binds is a rest
		// (#e000k8d): the finding would otherwise land on the spread, away from where the name was
		// chosen.
		{"a use of a rest parameter", "function run(...args: string[]) {\n    use(args);\n}\n"},
		{"a spread of a rest parameter", "function forward(...args: unknown[]) {\n    return target(...args);\n}\n"},
		{"a use of a destructured rest", "const [first, ...args] = source;\nuse(first, args);\n"},
		{"two rests in one file", "function run(...args: string[]) {\n    use(args);\n}\nconst other = (...args: number[]) => args.length;\n"},
		{"queryFn as an object key", "const query = { queryFn: load, mutationFn: save };\n"},

		// React 19 made `ref` a regular property, so the canonical spelling is load-bearing.
		{"a ref interface member", "interface ThingProperties {\n    ref?: Reference;\n}\n"},
		{"a ref shorthand", "const forwarded = { ref };\n"},
		{"a destructured ref forwarded onward", "export const Thing = ({ ref }) => <Child ref={ref} />;\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, ConsistencyNoAbbreviatedIdentifier, abbreviatedFile, testCase.sourceText)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// The framework exemptions are the half that needs options, and a rule that reads options has a
// failure mode an unconfigured one does not: it can guard everything or nothing. Both directions
// belong in fixtures.
func TestConsistencyNoAbbreviatedIdentifierFrameworkExemptions(t *testing.T) {
	t.Parallel()

	options := ConsistencyNoAbbreviatedIdentifierOptions{
		FrameworkParameterFilePatterns:  []string{"/app/"},
		FrameworkConstantFilePatterns:   []string{"/middleware.ts"},
		FrameworkParameterScopeNames:    []string{"generateMetadata", "generateStaticParams"},
		FrameworkParameterScopeSuffixes: []string{"PageRoute"},
	}

	cases := []struct {
		name       string
		fileName   string
		sourceText string
		wantIds    []string
	}{
		{
			"params in a page file is mandated by the framework",
			"/repository/app/blog/page.tsx",
			"export default function Page({ params }) {\n    return params;\n}\n",
			nil,
		},
		{
			"searchParams in a page file is mandated too",
			"/repository/app/blog/page.tsx",
			"export default function Page({ searchParams }) {\n    return searchParams;\n}\n",
			nil,
		},
		{
			"the same name outside a framework file is judged on its merits",
			"/repository/source/Thing.tsx",
			"export function read({ params }) {\n    return params;\n}\n",
			[]string{"abbreviatedIdentifier", "abbreviatedIdentifier"},
		},
		{
			"config in a middleware file is mandated",
			"/repository/middleware.ts",
			"export const config = { matcher: '/' };\n",
			nil,
		},
		{
			"config elsewhere still fires",
			"/repository/source/Thing.tsx",
			"export const config = { matcher: '/' };\n",
			[]string{"abbreviatedIdentifier"},
		},
		{
			"a framework scope name exempts wherever the file lives",
			"/repository/source/Thing.tsx",
			"export function generateMetadata({ params }) {\n    return params;\n}\n",
			nil,
		},
		{
			"a framework scope suffix exempts a component named for the route",
			"/repository/source/Thing.tsx",
			"export function BlogPageRoute({ params }) {\n    return params;\n}\n",
			nil,
		},
		{
			// The role suffix has to be stripped before the scope suffix can match, so this is the
			// case that separates a `PageRoute` suffix from `PageRouteProperties`. It is deliberately
			// a function parameter rather than an interface key: a key is already exempt as a
			// property signature, which is how an earlier version of this fixture passed without
			// ever reaching the stripping it claimed to test.
			"a scope suffix matches after the role suffix is stripped",
			"/repository/source/Thing.tsx",
			"export function BlogPageRouteProperties({ params }) {\n    return params;\n}\n",
			nil,
		},
		{
			"the stripping covers every role suffix our conventions append",
			"/repository/source/Thing.tsx",
			"export const BlogPageRouteOptions = ({ params }) => params;\n",
			nil,
		},
		{
			// An interface key is JUDGED, and the framework options do not change that. A type
			// member often shapes an external surface, which is why the original withholds its
			// autofix there, but it still reports: the name is one this file declares and can
			// choose. The framework exemptions above are about a name the framework MANDATES,
			// which is a different question and the reason these two are stated separately.
			"an interface key is judged as a property signature",
			"/repository/source/Thing.tsx",
			"interface ThingProperties {\n    params: string;\n}\n",
			[]string{"abbreviatedIdentifier"},
		},
		{
			"a non-framework function in the same file is still judged",
			"/repository/source/Thing.tsx",
			"export function readThing({ params }) {\n    return params;\n}\n",
			[]string{"abbreviatedIdentifier", "abbreviatedIdentifier"},
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, ConsistencyNoAbbreviatedIdentifier,
				testCase.fileName, testCase.sourceText, options)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// Without options the framework exemptions cannot apply, and the rule must still run rather than
// guard everything. `boundary-no-project-import` was enabled and inert for months because a missing
// option made it decline every file, which looks exactly like a clean run.
func TestConsistencyNoAbbreviatedIdentifierRunsWithoutOptions(t *testing.T) {
	t.Parallel()

	result := rule_testing.Run(t, ConsistencyNoAbbreviatedIdentifier,
		"/repository/app/blog/page.tsx", "export default function Page({ params }) {\n    return params;\n}\n")
	rule_testing.ExpectFindings(t, result, "abbreviatedIdentifier", "abbreviatedIdentifier")
}

// The suggested name travels in the message, since there is no fix to carry it. A message naming
// the wrong replacement is worse than no message, so the text is asserted rather than just the id.
func TestConsistencyNoAbbreviatedIdentifierNamesTheReplacement(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantText   string
	}{
		{"exact expansion", "const idx = 0;\n", `Use "index"`},
		{"prefix expansion keeps the tail", "const ctxValue = 1;\n", `Use "contextValue"`},
		{"suffix expansion keeps the head", "const rowProps = 1;\n", `Use "rowProperties"`},
		{"milliseconds at the end", "const maximumAgeMs = 1;\n", `Use "maximumAgeInMilliseconds"`},
		{"milliseconds in the middle", "const maximumAgeMsHalf = 1;\n", `Use "maximumAgeInMillisecondsHalf"`},
		{"a word segment in the middle", "const outputDirPath = 1;\n", `Use "outputDirectoryPath"`},
		{"a word segment at the end", "const originalInitCwd = 1;\n", `Use "originalInitWorkingDirectory"`},
		{"the longer word wins", "const themeVars = 1;\n", `Use "themeVariables"`},
		// One occurrence at a time, which is what JavaScript's non-global String.replace does and
		// what the original means. Expanding every occurrence at once would hand a reader a name
		// they did not ask for in a position the rule never reported.
		{"a repeated segment expands once", "const srcDirDstDir = 1;\n", `Use "srcDirectoryDstDir"`},
		{"a repeated millisecond unit expands once", "const ageMsHalfMs = 1;\n", `Use "ageInMillisecondsHalfMs"`},
		// `arguments` is a reserved binding, so a bare `args` names no replacement and says why.
		{"args offers alternatives instead of one name", "const args = 1;\n", `"commandLineArguments"`},
		{"Fn asks for a role rather than an expansion", "const compareFn = () => 1;\n", `like "Handler"`},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, ConsistencyNoAbbreviatedIdentifier, abbreviatedFile, testCase.sourceText)
			if len(result.Diagnostics) == 0 {
				t.Fatalf("expected a finding for %q", testCase.sourceText)
			}
			description := result.Diagnostics[0].Message.Description
			if !strings.Contains(description, testCase.wantText) {
				t.Fatalf("expected the message to contain %q, got %q", testCase.wantText, description)
			}
		})
	}
}

// The candidate gate is the speed story: it decides whether a name is worth spending the rest of the
// rule on. It must stay a superset of the branches, because a false negative here silently stops the
// rule firing and looks exactly like a clean run.
func TestConsistencyNoAbbreviatedIdentifierCandidateGateCoversEveryBranch(t *testing.T) {
	t.Parallel()

	// Every name any branch below the gate can report, one per branch family.
	reportable := []string{
		"prop", "props", "param", "params", "ref", "config", "idx", "arg", "args", "acc",
		"char", "fn", "str", "val", "arr", "obj", "num", "res", "err", "req", "msg", "min", "max",
		"prev", "cur", "pct", "opts", "ctx", "db", "tx",
		"ctxValue", "dbCode", "txHash", "optsForRun", "curStep", "pctComplete", "prevStep",
		"idxStart", "configValue", "propName", "propsList", "paramName", "paramsText", "refCount",
		"argCount", "argsList", "charSet", "fnCache", "strValue", "valInput", "arrItems",
		"objTarget", "numPages", "resBody", "errCount", "reqBody", "msgText", "minAge", "maxAge",
		"rowProp", "rowProps", "queryParam", "searchParams", "nodeRef", "buildConfig", "rowIdx",
		"firstArg", "esBuildArgs", "lastChar", "compareFn", "queryStr", "inputVal", "itemArr",
		"targetObj", "pageNum", "fetchRes", "parseErr", "httpReq", "errorMsg", "pageMin", "pageMax",
		"maximumAgeMs", "maximumAgeMsHalf",
		"originalInitCwd", "outputDirPath", "buildEnvName", "runCliCommand", "bufferLen",
		"eventSeq", "primaryDbHost", "pendingTxHash", "themeVars", "themeVarName",
	}
	for _, name := range reportable {
		if !isAbbreviationCandidate(name) {
			t.Errorf("the gate rejects %q, which a branch below would have reported", name)
		}
	}

	// The gate has to reject the ordinary name, or it buys nothing.
	for _, name := range []string{
		"renderNodeOrComponent", "identifier", "sourceFile", "handleSubmit", "useMediaQuery",
	} {
		if isAbbreviationCandidate(name) {
			t.Errorf("the gate accepts %q, which no branch would report", name)
		}
	}
}

// A finding has to land on the line the author can suppress.
//
// A node's Pos() sits before its leading trivia, so a binding preceded by a comment anchors at the
// comment rather than at the name. That is the one line an `eslint-disable-next-line` above it
// cannot cover, since the directive matches the line after itself. Measured on the real tree:
// `params` in McpApi.ts reported at line 240 while its suppression sat on 241 covering 242, so a
// correctly-suppressed identifier still produced a finding nothing could silence.
//
// The rule anchors on rule.TokenRange explicitly rather than leaning on ReportNode to do it. Both
// trim today, so this asserts the position rather than the call: the guarantee that matters is where
// the finding lands, and a fixture that tested which helper was called would go on passing if the
// helper changed underneath it.
//
// The column is asserted alongside the line because trivia swallowed within a single line moves only
// the column, and a line-only assertion cannot see it.
func TestConsistencyNoAbbreviatedIdentifierReportsAtTheIdentifier(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantLine   int
		wantColumn int
	}{
		{
			// The real-tree shape. Leading trivia spans two lines here, so an untrimmed range lands
			// on `method` two lines above the name.
			"a destructured binding under a comment",
			"function handleRequest(message: M): void {\n" +
				"    const {\n" +
				"        id,\n" +
				"        method,\n" +
				"        // eslint-disable-next-line nexus/consistency-no-abbreviated-identifier\n" +
				"        params,\n" +
				"    } = message;\n" +
				"}\n",
			6, 9,
		},
		{
			// A parameter's trivia reaches back to the open parenthesis on the signature line.
			"a parameter under a comment",
			"function handleRequest(\n    // eslint-disable-next-line\n    params: string,\n) {}\n",
			3, 5,
		},
		{
			// A class member, whose trivia reaches back past the comment to the previous member.
			"a class property under a comment",
			"class Thing {\n    id = 1;\n    // eslint-disable-next-line\n    paramsText = 2;\n}\n",
			4, 5,
		},
		{
			// Within one line the whitespace after `const` is the trivia, so only the column moves.
			"a declaration with no comment above it",
			"const params = 1;\n",
			1, 7,
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, ConsistencyNoAbbreviatedIdentifier, "/repository/source/Thing.ts", testCase.sourceText)
			if len(result.Diagnostics) == 0 {
				t.Fatalf("expected a finding, got none")
			}

			line, column := scanner.GetECMALineAndByteOffsetOfPosition(result.SourceFile, result.Diagnostics[0].Range.Pos())
			if line+1 != testCase.wantLine || column+1 != testCase.wantColumn {
				t.Fatalf("expected the finding at %d:%d, got %d:%d",
					testCase.wantLine, testCase.wantColumn, line+1, column+1)
			}
		})
	}
}

// TestConsistencyNoAbbreviatedIdentifierJudgesTypeMemberKeys pins the shape that was silently exempt.
//
// `IsForeignName` answers true for a property signature, and it is right to for
// `consistency-no-ambiguous-identifier`, which shares it and wants the whole family skipped. This
// rule subtracts type member keys back out, because the original only withholds the autofix there
// rather than the finding, and this port has no autofix to withhold.
//
// The `maxAgentsBytes` case is the real one, reduced from `modules/pensieve/PensieveBootstrap.ts`
// where a full eslint run reports five findings this port reported zero of. A method signature is
// included because the original walks `Identifier` without distinguishing the two member kinds, and
// a nested type literal because the skip was keyed on the parent node rather than on depth.
func TestConsistencyNoAbbreviatedIdentifierJudgesTypeMemberKeys(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		{
			"an interface member carrying a prefix abbreviation",
			"interface BootstrapOptionsInterface {\n    maxAgentsBytes?: number;\n}\n",
			[]string{"abbreviatedIdentifier"},
		},
		{
			"two interface members, the shape measured on the real tree",
			"interface BootstrapOptionsInterface {\n    maxAgentsBytes?: number;\n    maxBootBytes?: number;\n}\n",
			[]string{"abbreviatedIdentifier", "abbreviatedIdentifier"},
		},
		{
			"a type literal member",
			"type BootstrapOptionsType = {\n    maxBootBytes?: number;\n};\n",
			[]string{"abbreviatedIdentifier"},
		},
		{
			"a method signature key",
			"interface ReaderInterface {\n    maxDepth(): number;\n}\n",
			[]string{"abbreviatedIdentifier"},
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, ConsistencyNoAbbreviatedIdentifier, abbreviatedFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// TestConsistencyNoAbbreviatedIdentifierStillSkipsTheRestOfTheForeignFamily is the control.
//
// The narrowing above subtracts exactly one member from `IsForeignName`'s set. Without this test a
// mutation widening it to "never skip a foreign name" would pass every assertion in the file, since
// nothing else asserts the remaining exemptions survive.
func TestConsistencyNoAbbreviatedIdentifierStillSkipsTheRestOfTheForeignFamily(t *testing.T) {
	t.Parallel()

	cases := []struct{ name, sourceText string }{
		{"a property read", "const value = thing.maxAgentsBytes;\n"},
		{"an object literal key", "const options = { maxAgentsBytes: 1 };\n"},
		{"an import specifier", "import { maxAgentsBytes } from 'external';\n"},
		{"a namespace import", "import * as maxAgentsBytes from 'external';\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, ConsistencyNoAbbreviatedIdentifier, abbreviatedFile, testCase.sourceText)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestConsistencyNoAbbreviatedIdentifierSkipsNamesTheCodeOnlyReads holds the principle from
// api-phi-health's parity sweep (#dx1vrfm): the rule judges names the code declares, never names it
// reads from elsewhere or renames away from. Each silent row is one of the six reported sites in
// shape; each reporting row is the nearest shape the rule must still catch.
func TestConsistencyNoAbbreviatedIdentifierSkipsNamesTheCodeOnlyReads(t *testing.T) {
	t.Parallel()

	silent := []struct{ name, sourceText string }{
		// ModuleTestDiscovery.ts: tsconfig-paths' loadConfig, read in a type query and a call.
		{"a reference to an imported name", "import { loadConfig } from 'tsconfig-paths';\nlet configuration: ReturnType<typeof loadConfig> | undefined;\nconfiguration = loadConfig('.');\nexport { configuration };\n"},
		// OrmSchemaBuilderDrizzleMySql.ts: drizzle's char, called.
		{"a call of an imported function", "import { char } from 'drizzle-orm/mysql-core';\nexport const column = char('name', { length: 36 });\n"},
		{"a reference to a default import", "import loadConfig from 'config-loader';\nexport const configuration = loadConfig();\n"},
		// MappedJoins.test.ts: drizzle's params key, renamed on destructure.
		{"a destructuring key renamed away", "declare const query: { toSQL(): Record<string, unknown> };\nconst { sql: statement, params: parameters } = query.toSQL();\nexport { statement, parameters };\n"},
		// FileStorageVideoDerivativeService.ts: an external maxBitrate key, renamed on destructure.
		{"an external key renamed away", "declare const entry: Record<string, unknown>;\nconst { maxBitrate: maximumBitrate } = entry;\nexport { maximumBitrate };\n"},
	}
	for _, testCase := range silent {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, ConsistencyNoAbbreviatedIdentifier, abbreviatedFile, testCase.sourceText))
		})
	}

	fires := []struct{ name, sourceText, id string }{
		// A shorthand destructure declares the abbreviated name in this file.
		{"a shorthand destructure", "declare const query: { params: unknown[] };\nconst { params } = query;\nexport { params };\n", "abbreviatedIdentifier"},
		// The local half of a rename is this file's name.
		{"a rename onto an abbreviation", "declare const entry: { maximumBitrate: number };\nconst { maximumBitrate: maxBitrate } = entry;\nexport { maxBitrate };\n", "abbreviatedIdentifier"},
		// A file that shadows an import declares that name too, so it is judged rather than guessed.
		{"a local that shadows an import", "import { char } from 'drizzle-orm/mysql-core';\nexport function build() {\n    const char = 1;\n    return char;\n}\nexport { char as column };\n", "abbreviatedIdentifier"},
	}
	for _, testCase := range fires {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, ConsistencyNoAbbreviatedIdentifier, abbreviatedFile, testCase.sourceText)
			if len(result.Diagnostics) == 0 {
				t.Fatalf("expected a finding, got none")
			}
			if result.Diagnostics[0].Message.Id != testCase.id {
				t.Fatalf("expected %s first, got %v", testCase.id, result.MessageIds())
			}
		})
	}
}

// TestConsistencyNoAbbreviatedIdentifierLeavesNextRouteContractsAlone covers the floor
// `nextjs.IsRouteContractExport` sets with no options at all.
//
// `maxDuration` is the case no consumer list carried: it passes the `max[A-Z]` gate, and the rule
// told route authors to write `maximumDuration`, which Next then silently ignores. Each silent case is
// paired with the same text in a file Next does not read, which must still report, so the exemption
// is shown to come from the route file rather than from the name.
func TestConsistencyNoAbbreviatedIdentifierLeavesNextRouteContractsAlone(t *testing.T) {
	t.Parallel()

	const routeFile = "/repository/app/api/report/route.ts"
	const pageFile = "/repository/app/blog/[slug]/page.tsx"
	cases := []struct {
		name       string
		fileName   string
		sourceText string
		wantIds    []string
	}{
		{"maxDuration in a route", routeFile, "export const maxDuration = 60;", nil},
		{"maxDuration outside a route", abbreviatedFile, "export const maxDuration = 60;", []string{"abbreviatedIdentifier"}},
		{"generateStaticParams in a page", pageFile, "export async function generateStaticParams() { return []; }", nil},
		{"generateStaticParams outside a route", abbreviatedFile, "export async function generateStaticParams() { return []; }", []string{"abbreviatedSuffix"}},
		{"dynamicParams in a page", pageFile, "export const dynamicParams = false;", nil},
		{"config in a page", pageFile, "export const config = {};", nil},
		{"a reference to the export is the export", pageFile, "export const maxDuration = 60;\nconsole.log(maxDuration);", nil},
		// A nested binding sharing the spelling is a different variable, judged on its merits. Its
		// declaration reports and its reference does not: references are exempt by spelling, since
		// telling a shadow from the export needs scope resolution this rule does not do. One finding,
		// on the declaration, is enough for the author to rename both.
		{"a nested config shadows the contract", pageFile, "export function load() {\n    const config = 1;\n    return config;\n}", []string{"abbreviatedIdentifier"}},
		{"a parameter named for the contract", pageFile, "export function load(maxDuration: number) {\n    return 1;\n}", []string{"abbreviatedIdentifier"}},
		// Contracts differ by file: a page has no http methods and a route has no metadata, and
		// `params` is no contract export anywhere, so it keeps its own options-driven handling.
		{"a page is not a route handler", routeFile, "export const generateViewport = () => ({});", nil},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, ConsistencyNoAbbreviatedIdentifier, testCase.fileName, testCase.sourceText)
			if len(testCase.wantIds) == 0 {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}
