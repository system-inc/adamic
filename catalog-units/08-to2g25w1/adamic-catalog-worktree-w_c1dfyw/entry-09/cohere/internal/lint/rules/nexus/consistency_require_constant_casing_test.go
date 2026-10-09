package nexus

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const constantCasingFile = "/repository/source/Thing.ts"
const constantCasingComponentFile = "/repository/source/Thing.tsx"

func TestConsistencyRequireConstantCasingFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		fileName   string
		sourceText string
		wantId     string
	}{
		// The three clauses, each firing on its own failure.
		{
			"camelCase export", constantCasingFile,
			"export const numberAbsentPlaceholder = '-';\n",
			"requirePascalCaseExported",
		},
		{
			"PascalCase local", constantCasingFile,
			"const OrderColumns = [1, 2];\n",
			"requireCamelCaseLocal",
		},
		{
			"PascalCase exported function", constantCasingFile,
			"export const FormatNumber = (value: number) => String(value);\n",
			"requireCamelCaseFunction",
		},
		{
			"PascalCase exported instance", constantCasingFile,
			"class NetworkService {}\nexport const NetworkServiceSingleton = new NetworkService();\n",
			"requireCamelCaseInstance",
		},

		// Kind beats reach: a function stays camelCase even where PascalCase would be right for data.
		{
			"PascalCase local function", constantCasingFile,
			"const FormatNumber = function (value: number) { return String(value); };\n",
			"requireCamelCaseFunction",
		},
		// Usage is the evidence: an ordinary call on the right, proven a function by being invoked.
		{
			"called like a function", constantCasingFile,
			"import { build } from 'x';\nconst Handler = build();\nHandler();\n",
			"requireCamelCaseFunction",
		},
		// The declared return type of a local resolver says the constant holds an instance. A factory
		// returning a function is read from the constant's type, in the typed fixtures below.
		{
			"local resolver returning a class", constantCasingFile,
			"class AgentManager {}\nfunction resolveAgentManager(): AgentManager {\n" +
				"    return new AgentManager();\n}\n" +
				"export const AgentManagerSingleton = resolveAgentManager();\n",
			"requireCamelCaseInstance",
		},
		// The cached-singleton guard, in both operand orders.
		{
			"cached singleton, construction on the right", constantCasingFile,
			"declare const globalForThing: { thing?: Thing };\nclass Thing {}\n" +
				"export const CachedThing = globalForThing.thing ?? new Thing();\n",
			"requireCamelCaseInstance",
		},
		{
			"cached singleton, construction on the left", constantCasingFile,
			"declare const globalForThing: { thing?: Thing };\nclass Thing {}\n" +
				"export const CachedThing = new Thing() || globalForThing.thing;\n",
			"requireCamelCaseInstance",
		},
		// A bare class annotation says what the value holds even when the initializer cannot.
		{
			"class type annotation", constantCasingFile,
			"import { Client } from 'x';\nimport { lookup } from 'y';\n" +
				"export const DiscordClient: Client = lookup();\n",
			"requireCamelCaseInstance",
		},

		// Separator folding: the suggestion has to be a name the rule would accept.
		{
			"snake_case local", constantCasingFile,
			"const user_id = 'x';\n",
			"requireCamelCaseLocal",
		},
		{
			"snake_case export", constantCasingFile,
			"export const apple_root_ca = 'x';\n",
			"requirePascalCaseExported",
		},

		// A PascalCase local in a .tsx file that is never rendered is still a PascalCase local. This
		// is what makes the JSX exemption resolve against usage rather than the file extension.
		{
			"unrendered PascalCase local in tsx", constantCasingComponentFile,
			"const OrderColumns = [1, 2];\n",
			"requireCamelCaseLocal",
		},
		// A name merely ending in Context is judged on its casing; only the call shape exempts.
		{
			"value named Context without the call", constantCasingFile,
			"const ThemeContext = { value: 1 };\n",
			"requireCamelCaseLocal",
		},
		// A lookup table is structurally an as-const object too, but its keys are labels rather than
		// members, so it is data and takes the exported/local split.
		{
			"as const lookup table is not an enum", constantCasingFile,
			"const classNames = { small: 'a', large: 'b' } as const;\nexport const table = classNames;\n",
			"requirePascalCaseExported",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, ConsistencyRequireConstantCasing, testCase.fileName, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, testCase.wantId)
		})
	}
}

func TestConsistencyRequireConstantCasingStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		fileName   string
		sourceText string
	}{
		// The three clauses, satisfied.
		{"PascalCase export", constantCasingFile, "export const OrderColumns = [1, 2];\n"},
		{"camelCase local", constantCasingFile, "const orderColumns = [1, 2];\n"},
		{
			"camelCase exported function", constantCasingFile,
			"export const formatNumber = (value: number) => String(value);\n",
		},
		{
			"camelCase exported instance", constantCasingFile,
			"class NetworkService {}\nexport const networkService = new NetworkService();\n",
		},

		// Only const carries the immutability the convention rests on.
		{"let is not const", constantCasingFile, "let OrderColumns = [1, 2];\n"},
		{"var is not const", constantCasingFile, "var OrderColumns = [1, 2];\n"},

		// Destructuring takes its names from the shape of the value rather than from a choice.
		{"destructured local", constantCasingFile, "const { OrderColumns } = source;\ndeclare const source: any;\n"},
		{"array destructuring", constantCasingFile, "const [First] = list;\ndeclare const list: string[];\n"},

		// A declaration describes something this file does not control, so the upstream spelling is
		// the contract. Lowercasing it would make the script throw while the type-checker stayed
		// happy, because the declaration is all TypeScript ever sees.
		{"declare const", constantCasingFile, "declare const DocumentApp: unknown;\n"},

		// Owned by consistency-no-screaming-snake-case, which has the better message for it.
		{"screaming snake local", constantCasingFile, "const MAX_RETRY_COUNT = 3;\n"},
		{"screaming snake export", constantCasingFile, "export const HTTP_TIMEOUT = 5000;\n"},

		// The trailing underscore escapes a keyword, so recasing produces a name the parser rejects.
		{"reserved word escape", constantCasingFile, "const do_ = 1;\n"},
		{"reserved word escape, exported", constantCasingFile, "export const class_ = 1;\n"},

		// A leading underscore marks a knowingly-unused binding, which is orthogonal to casing.
		{"leading underscore local", constantCasingFile, "const _results = [1];\n"},
		{"leading underscore export", constantCasingFile, "export const _Results = [1];\n"},

		// Built-in containers hold rather than act, so an exported one is exported data.
		{"exported Map keeps PascalCase", constantCasingFile, "export const Conversations = new Map();\n"},
		{"exported Set keeps PascalCase", constantCasingFile, "export const Seen = new Set();\n"},
		{"exported URL keeps PascalCase", constantCasingFile, "export const Endpoint = new URL('https://x');\n"},

		// React shapes, each of which needs its capital for a different reason.
		{"createContext", constantCasingFile, "import { createContext } from 'react';\nconst ThemeContext = createContext(null);\n"},
		{
			"React.createContext", constantCasingFile,
			"import * as React from 'react';\nconst ThemeContext = React.createContext(null);\n",
		},
		{
			"memo-wrapped component", constantCasingComponentFile,
			"import { memo } from 'react';\nexport const MenuSearch = memo(function MenuSearch() { return null; });\n",
		},
		{
			"FC-annotated component", constantCasingComponentFile,
			"import type { FC } from 'react';\nexport const ShareBlock: FC = function () { return null; };\n",
		},
		{
			"React.FC-annotated component", constantCasingComponentFile,
			"import * as React from 'react';\nexport const ShareBlock: React.FC = function () { return null; };\n",
		},
		// JSX owns the capital letter, so a rendered local keeps it.
		{
			"local rendered as an element", constantCasingComponentFile,
			"declare function getIcon(): any;\nexport function Row() {\n" +
				"    const Icon = getIcon();\n    return <Icon />;\n}\n",
		},
		// Handed to something else to render: the receiver will put it in element position.
		{
			"local handed to a JSX attribute", constantCasingComponentFile,
			"declare function getIcon(): any;\ndeclare function Panel(properties: { icon: any }): any;\n" +
				"export function Row() {\n    const Icon = getIcon();\n    return <Panel icon={Icon} />;\n}\n",
		},
		// The capital is what keeps the declaration from referencing itself.
		{
			"shadows a camelCase source", constantCasingComponentFile,
			"declare const processingIcon: any;\ndeclare const BrokenCircleIcon: any;\n" +
				"export function Row() {\n    const ProcessingIcon = processingIcon ?? BrokenCircleIcon;\n" +
				"    return ProcessingIcon;\n}\n",
		},

		// Mirrors an upstream constructor's own spelling, including through a fallback chain where
		// the constructor sits two levels down.
		{"re-binds a global constructor", constantCasingFile, "const NativeMap = globalThis.Map;\n"},
		{
			"constructor picked out of a fallback chain", constantCasingFile,
			"declare const polyfill: any;\n" +
				"const ResizeObserverConstructor = polyfill ?? (typeof window !== 'undefined' ? window.ResizeObserver : undefined);\n",
		},
		// A lowercase property read is not a constructor re-binding.
		{"lowercase property read is unaffected", constantCasingFile, "declare const counts: any;\nconst total = counts.sum;\n"},

		// An import wearing a const; import-require-node-namespace owns this spelling.
		{
			"dynamic import binding", constantCasingFile,
			"export async function run() {\n    const NodeFileSystem = await import('node:fs');\n" +
				"    return NodeFileSystem;\n}\n",
		},
		{
			"dynamic import reaching a default export", constantCasingFile,
			"export async function run() {\n    const Database = (await import('better-sqlite3')).default;\n" +
				"    return Database;\n}\n",
		},

		// consistency-require-type-suffix requires these to end in Kind, which forces PascalCase.
		{
			"const enum shape, string members", constantCasingFile,
			"const ButtonVariantKind = { Default: 'Default', Destructive: 'Destructive' } as const;\n" +
				"export const Used = ButtonVariantKind;\n",
		},
		{
			"const enum shape, numeric members", constantCasingFile,
			"const DirectionKind = { Positive: 1, Negative: -1 } as const;\nexport const Used = DirectionKind;\n",
		},

		// The suffixes our conventions give to shapes rather than to classes, and the generic
		// container that turned 129 exported lookup tables into false reports.
		{
			"Record annotation is data", constantCasingFile,
			"export const OrderColumns: Record<string, number> = { a: 1 };\n",
		},
		// The suffixes our conventions give to shapes rather than to classes. The factory is
		// imported on purpose: a locally declared one whose return type is PascalCase reaches the
		// resolver clause first and is judged an instance, which is correct and a different path.
		{
			"Properties-suffixed annotation is data", constantCasingFile,
			"import type { RowProperties } from 'x';\nimport { read } from 'y';\n" +
				"export const DefaultRow: RowProperties = read();\n",
		},
		{
			"Options-suffixed annotation is data", constantCasingFile,
			"import type { RowOptions } from 'x';\nimport { read } from 'y';\n" +
				"export const DefaultRow: RowOptions = read();\n",
		},
		{
			"object literal outranks a class-shaped annotation", constantCasingFile,
			"import type { Transition } from 'framer-motion';\n" +
				"export const CollapsibleTransition: Transition = { type: 'spring' };\n",
		},

		// A function alias, resolved through the file rather than through a type.
		{
			"alias for a local function", constantCasingFile,
			"function numberCompact(value: number) { return String(value); }\n" +
				"export const formatNumber = numberCompact;\n",
		},
		// An import is read through its type, in the typed fixtures below, not through its casing.
		// Without a checker an alias for one has no function evidence, so PascalCase stays data.
		{
			"alias for a PascalCase import is data", constantCasingFile,
			"import { ProjectRoot } from 'x';\nexport const AhraProjectRoot = ProjectRoot;\n",
		},

		// A construction cast to something else. This is the shape of every generated GraphQL
		// document constant, and unwrapping the cast to find the "new" underneath produced 95
		// findings on a tree whose baseline is clean. A value cast away from its constructor is not
		// being handed to the reader as that class, and this rule has no type information to
		// overrule the author with, so the instance clause reads the initializer raw here.
		{
			"construction cast away from its class", constantCasingFile,
			"export class TypedDocumentString<R, V> { constructor(public value: string) {} }\n" +
				"export const AccountDocument = new TypedDocumentString('query {}') as unknown as " +
				"TypedDocumentString<number, number>;\n",
		},

		{"no declarations", constantCasingFile, "export function run() {\n    return 1;\n}\n"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, ConsistencyRequireConstantCasing, testCase.fileName, testCase.sourceText)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// The framework option exists because a name a library reads verbatim is a contract rather than a
// choice, and the rule cannot know which names those are.
func TestConsistencyRequireConstantCasingFrameworkNames(t *testing.T) {
	t.Parallel()

	source := "export const runtime = 'edge';\n"

	reported := rule_testing.Run(t, ConsistencyRequireConstantCasing, constantCasingFile, source)
	rule_testing.ExpectFindings(t, reported, "requirePascalCaseExported")

	exempt := rule_testing.RunWithOptions(t, ConsistencyRequireConstantCasing, constantCasingFile, source,
		ConsistencyRequireConstantCasingOptions{FrameworkConstantNames: []string{"runtime"}})
	rule_testing.ExpectClean(t, exempt)
}

// The suggestion has to be a name the rule's own predicates accept, or the reader who takes it gets
// the same report again. This is the check that catches a converter handing back its own input.
func TestConsistencyRequireConstantCasingSuggestionsAreAccepted(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantAccept func(string) bool
	}{
		{"snake_case to camelCase", "const user_id = 'x';\n", isCamelCase},
		{"snake_case to PascalCase", "export const apple_root_ca = 'x';\n", isPascalCase},
		{"screaming part folded", "const utm_SOURCE = 'x';\n", isCamelCase},
		{"screaming part folded, exported", "export const utm_SOURCE = 'x';\n", isPascalCase},
		{"camelCase to PascalCase", "export const numberAbsentPlaceholder = '-';\n", isPascalCase},
		{"PascalCase to camelCase", "const OrderColumns = [1];\n", isCamelCase},
		{"leading underscore kept", "const _order_columns = [1];\n", isCamelCase},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, ConsistencyRequireConstantCasing, constantCasingFile, testCase.sourceText)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("want exactly one finding, got %d", len(result.Diagnostics))
			}

			// The message quotes the suggestion in parentheses, which is where the reader takes it
			// from, so that is the string this reads back.
			description := result.Diagnostics[0].Message.Description
			opening := strings.Index(description, `("`)
			if opening < 0 {
				t.Fatalf("no suggestion in message: %s", description)
			}
			rest := description[opening+2:]
			closing := strings.Index(rest, `"`)
			if closing < 0 {
				t.Fatalf("unterminated suggestion in message: %s", description)
			}
			suggestion := rest[:closing]

			if !testCase.wantAccept(suggestion) {
				t.Fatalf("suggestion %q is not a casing the rule accepts", suggestion)
			}
		})
	}
}

// Acceptance is necessary and not sufficient, which is the whole reason this second test exists.
//
// isCamelCase accepts any letters after the first, so the un-folded "utmSOURCE" passes it while
// being exactly the shape the folding exists to prevent. A test that only asked "is the suggestion
// acceptable" stayed green with the fold removed, so it was measuring the wrong property. These
// assert the exact string, which is the only thing that pins the fold down.
func TestConsistencyRequireConstantCasingSuggestionText(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		want       string
	}{
		// An all-uppercase part has no internal boundaries to preserve, so it is lowered whole.
		// Leaving it produces "utmSOURCE", which is neither casing.
		{"all-uppercase part is folded whole", "const utm_SOURCE = 'x';\n", "utmSource"},
		{"all-uppercase part is folded whole, exported", "export const utm_SOURCE = 'x';\n", "UtmSource"},
		{"underscores become word boundaries", "const user_id = 'x';\n", "userId"},
		{"underscores become word boundaries, exported", "export const apple_root_ca = 'x';\n", "AppleRootCa"},
		// The leading underscore is the unused-binding convention and is not this rule's to remove.
		{"leading underscore survives", "const _order_columns = [1];\n", "_orderColumns"},
		{"internal boundaries survive", "export const numberAbsentPlaceholder = '-';\n", "NumberAbsentPlaceholder"},
		{"internal boundaries survive, lowered", "const OrderColumns = [1];\n", "orderColumns"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, ConsistencyRequireConstantCasing, constantCasingFile, testCase.sourceText)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("want exactly one finding, got %d", len(result.Diagnostics))
			}
			if suggestion := suggestionFrom(t, result.Diagnostics[0].Message.Description); suggestion != testCase.want {
				t.Fatalf("suggestion = %q, want %q", suggestion, testCase.want)
			}
		})
	}
}

// suggestionFrom reads the name a message offers, which is quoted in parentheses and is where the
// reader takes it from.
func suggestionFrom(t *testing.T, description string) string {
	t.Helper()

	opening := strings.Index(description, `("`)
	if opening < 0 {
		t.Fatalf("no suggestion in message: %s", description)
	}
	rest := description[opening+2:]
	closing := strings.Index(rest, `"`)
	if closing < 0 {
		t.Fatalf("unterminated suggestion in message: %s", description)
	}
	return rest[:closing]
}

// The factory both real sites call, as ahra's modules/system/base/BaseGraphQlClient.ts declares it:
// a callable executor that also carries a `requestCapturingSetCookie` method.
const constantCasingGraphQlClientFactory = "export type BaseGraphQlExecutorType = (\n" +
	"    query: string,\n    variables?: Record<string, unknown>,\n) => Promise<Record<string, unknown>>;\n" +
	"export type BaseGraphQlClientType = BaseGraphQlExecutorType & {\n" +
	"    requestCapturingSetCookie: (query: string) => Promise<{ data: Record<string, unknown> }>;\n};\n" +
	"export function createGraphQlClient(options: { name: string; defaultApiUrl: string }): BaseGraphQlClientType {\n" +
	"    const executor = async function (query: string): Promise<Record<string, unknown>> {\n" +
	"        return { query, name: options.name };\n    };\n" +
	"    return Object.assign(executor, {\n" +
	"        requestCapturingSetCookie: async function (query: string) {\n" +
	"            return { data: { query, url: options.defaultApiUrl } };\n        },\n    });\n}\n"

// Whether a constant holds a function is read from its type through the checker, so the factory that
// made it can live in any module and carry any name. The two real sites are ConnectedGraphQl.ts:14
// and PhiGraphQl.ts:57 in ahra, each holding the client `createGraphQlClient` returns and each called
// by name across 16 files. The rule as it stood could not follow the import, so it told both to
// become PascalCase and both carried a suppression saying so.
func TestConsistencyRequireConstantCasingReadsTheRealSitesThroughTheirType(t *testing.T) {
	t.Parallel()

	const factoryPath = "source/system/base/BaseGraphQlClient.ts"
	cases := []struct {
		name    string
		files   map[string]string
		subject string
		wantIds []string
	}{
		{
			"ConnectedGraphQl.ts:14, the client an imported factory returns",
			map[string]string{
				factoryPath: constantCasingGraphQlClientFactory,
				"source/connected/ConnectedGraphQl.ts": "import { createGraphQlClient } from '../system/base/BaseGraphQlClient';\n" +
					"export const connectedGraphQlClient = createGraphQlClient({\n" +
					"    name: 'Connected',\n    defaultApiUrl: 'https://api.connected.app/graphql',\n});\n",
				"source/connected/ConnectedAgents.ts": "import { connectedGraphQlClient } from './ConnectedGraphQl';\n" +
					"export function listAgents(): Promise<Record<string, unknown>> {\n" +
					"    return connectedGraphQlClient('query { agents { id } }');\n}\n",
			},
			"source/connected/ConnectedGraphQl.ts",
			nil,
		},
		{
			"PhiGraphQl.ts:57, the same factory reached from deeper in the tree",
			map[string]string{
				factoryPath: constantCasingGraphQlClientFactory,
				"source/phi/core/PhiGraphQl.ts": "import { createGraphQlClient } from '../../system/base/BaseGraphQlClient';\n" +
					"export const phiGraphQlClient = createGraphQlClient({\n" +
					"    name: 'Phi',\n    defaultApiUrl: 'https://phi.health/api/graphql',\n});\n",
				"source/phi/core/PhiTickets.ts": "import { phiGraphQlClient } from './PhiGraphQl';\n" +
					"export function listTickets(): Promise<Record<string, unknown>> {\n" +
					"    return phiGraphQlClient('query { tickets { id } }');\n}\n",
			},
			"source/phi/core/PhiGraphQl.ts",
			nil,
		},
		{
			// The other direction stays silent: the type only permits camelCase, never requires it, so a
			// PascalCase callable from an imported factory is judged as the constant it may be. Base's
			// Verify* decorators and GraphQlQuery are exactly this shape, used as `@VerifyEmail`.
			"the same client named PascalCase, as a decorator would be",
			map[string]string{
				factoryPath: constantCasingGraphQlClientFactory,
				"source/connected/ConnectedGraphQl.ts": "import { createGraphQlClient } from '../system/base/BaseGraphQlClient';\n" +
					"export const ConnectedGraphQlClient = createGraphQlClient({\n" +
					"    name: 'Connected',\n    defaultApiUrl: 'https://api.connected.app/graphql',\n});\n",
			},
			"source/connected/ConnectedGraphQl.ts",
			nil,
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedFiles(t, ConsistencyRequireConstantCasing, testCase.files, testCase.subject)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// The factories the neighbours below call, one per kind of value a call can hand back.
const constantCasingFactories = "export interface Store {\n    get(): string;\n    set(value: string): void;\n}\n" +
	"export declare function createStore(): Store;\n" +
	"export class Decimal {\n    constructor(public value: number) {}\n}\n" +
	"export declare function createDecimalClass(): typeof Decimal;\n" +
	"export interface CallableConstructor {\n    (value: string): string;\n    new (value: string): object;\n}\n" +
	"export declare function createCallableConstructor(): CallableConstructor;\n" +
	"export declare function createLimit(): number;\n" +
	"export declare function createAnything(): any;\n" +
	"export declare function createLooseHandler(): (value: number) => any;\n" +
	"export declare function createMaybeHandler(): ((value: number) => number) | undefined;\n" +
	"export declare function createHandlerOrLabel(): ((value: number) => number) | string;\n" +
	"export declare function RuleCreator(name: string): (rule: object) => object;\n" +
	"export type HandlerType = (value: number) => number;\n" +
	"export declare function pick(): any;\n" +
	"export function numberCompact(value: number): string {\n    return String(value);\n}\n" +
	"export const defaultSettings = { retries: 3 };\n"

// The element shape a component returns, standing in for React's own declarations. The check reads
// members rather than names, so a fixture does not need the real package installed.
const constantCasingReact = "export interface ReactElement {\n    type: unknown;\n    props: unknown;\n    key: string | null;\n}\n" +
	"export type ReactNode = ReactElement | string | number | boolean | null | undefined;\n" +
	"export interface MemoExoticComponent<P> {\n    (props: P): ReactNode;\n    readonly $$typeof: symbol;\n}\n" +
	"export declare function memo<P>(\n    component: (props: P) => ReactNode,\n" +
	"    compare?: (previous: P, next: P) => boolean,\n): MemoExoticComponent<P>;\n"

// The neighbours of the real sites: what a type with a call signature must also be, or not be, for
// the constant to count as a function. Each is a call to a factory in another module, the shape no
// local evidence can answer.
func TestConsistencyRequireConstantCasingReadsWhatACallHandsBackThroughItsType(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		// An object with methods holds them; it is not called, so it is a constant.
		{"an object with methods, camelCase", "import { createStore } from './Factories';\nexport const settingsStore = createStore();\n", []string{"requirePascalCaseExported"}},
		{"an object with methods, PascalCase", "import { createStore } from './Factories';\nexport const SettingsStore = createStore();\n", nil},

		// A constructor keeps the capital `new` expects, even when it can also be called.
		{"a class", "import { createDecimalClass } from './Factories';\nexport const Decimal = createDecimalClass();\n", nil},
		{"a class, camelCase", "import { createDecimalClass } from './Factories';\nexport const decimal = createDecimalClass();\n", []string{"requirePascalCaseExported"}},
		{"callable and constructible", "import { createCallableConstructor } from './Factories';\nexport const Coerce = createCallableConstructor();\n", nil},

		// A plain value is data, wherever it was made.
		{"a number, camelCase", "import { createLimit } from './Factories';\nexport const retryLimit = createLimit();\n", []string{"requirePascalCaseExported"}},
		{"a number, PascalCase", "import { createLimit } from './Factories';\nexport const RetryLimit = createLimit();\n", nil},

		// `any` has no signatures and proves nothing, so it is judged on the other evidence, which
		// here is none.
		{"any", "import { createAnything } from './Factories';\nexport const anything = createAnything();\n", []string{"requirePascalCaseExported"}},

		// Null and undefined are set aside; a part that is not callable is not.
		{"a function or undefined", "import { createMaybeHandler } from './Factories';\nexport const handler = createMaybeHandler();\n", nil},
		{"a function or a string", "import { createHandlerOrLabel } from './Factories';\nexport const handlerOrLabel = createHandlerOrLabel();\n", []string{"requirePascalCaseExported"}},

		// The factory's name no longer matters: a Creator is read like any other call.
		{"a Creator-named factory, PascalCase", "import { RuleCreator } from './Factories';\nexport const CreateEsLintRule = RuleCreator('a');\n", []string{"requireCamelCaseFunction"}},
		{"a Creator-named factory, camelCase", "import { RuleCreator } from './Factories';\nexport const createEsLintRule = RuleCreator('a');\n", nil},

		// A local factory, annotated or not.
		{
			"a local factory returning a function",
			"function makeMiddleware(): (input: string) => string {\n" +
				"    return function (input: string) { return input; };\n}\n" +
				"export const LocaleMiddleware = makeMiddleware();\n",
			[]string{"requireCamelCaseFunction"},
		},
		{
			"an unannotated local factory returning a function, camelCase",
			"function makeMiddleware() {\n    return function (input: string) { return input; };\n}\n" +
				"export const localeMiddleware = makeMiddleware();\n",
			nil,
		},

		// The annotation is the type, whatever its name ends in.
		{"a function type annotation named Type", "import { pick, type HandlerType } from './Factories';\nexport const handler: HandlerType = pick();\n", nil},

		// An import is followed to what it is, not read from its casing.
		{"an alias for an imported function", "import { numberCompact } from './Factories';\nexport const formatNumber = numberCompact;\n", nil},
		// The rule as it stood reads a camelCase imported name as a function, and that reading is kept.
		{"an alias for an imported object", "import { defaultSettings } from './Factories';\nexport const settings = defaultSettings;\n", nil},

		// A component is a function JSX needs capitalized: the compound and memo shapes from
		// Structure's Dialog.tsx and TableVirtualizedRow.tsx.
		{
			"a compound component",
			"import type { ReactElement } from './react';\n" +
				"declare function DialogRoot(properties: { open: boolean }): ReactElement;\n" +
				"declare function DialogTrigger(properties: { label: string }): ReactElement;\n" +
				"export const Dialog = Object.assign(DialogRoot, { Trigger: DialogTrigger });\n",
			nil,
		},
		{
			"a memo component with a comparator",
			"import { memo, type ReactElement } from './react';\n" +
				"function TableRowInner(properties: { id: string }): ReactElement {\n" +
				"    return { type: 'tr', props: properties, key: null };\n}\n" +
				"export const TableRow = memo(TableRowInner, function (previous, next) {\n" +
				"    return previous.id === next.id;\n});\n",
			nil,
		},
		{
			"a compound component, camelCase",
			"import type { ReactElement } from './react';\n" +
				"declare function DialogRoot(properties: { open: boolean }): ReactElement;\n" +
				"export const dialog = Object.assign(DialogRoot, { Trigger: DialogRoot });\n",
			[]string{"requirePascalCaseExported"},
		},
		// A return the checker cannot see may be an element: what a component's return becomes
		// when the JSX types do not resolve. Such a callable is judged as it was before the checker
		// was asked, so neither casing is reported as a function.
		{
			"a compound component whose element type did not resolve",
			"declare function DialogRoot(properties: { open: boolean }): any;\n" +
				"export const Dialog = Object.assign(DialogRoot, { Trigger: DialogRoot });\n",
			nil,
		},
		{"a function returning any, PascalCase", "import { createLooseHandler } from './Factories';\nexport const LooseHandler = createLooseHandler();\n", nil},
		{"a function returning any, camelCase", "import { createLooseHandler } from './Factories';\nexport const looseHandler = createLooseHandler();\n", []string{"requirePascalCaseExported"}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedFiles(t, ConsistencyRequireConstantCasing, map[string]string{
				"source/Factories.ts": constantCasingFactories,
				"source/react.ts":     constantCasingReact,
				"source/Subject.ts":   testCase.sourceText,
				// Imports every export, so an exported camelCase constant gets the PascalCase advice
				// rather than the drop-the-export one.
				"source/Consumer.ts": "import * as Subject from './Subject';\nexport const Everything = Subject;\n",
			}, "source/Subject.ts")
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// The rename fix, against a real type graph, since references are resolved through the checker.
// The first case is the shape of the pensieve batch that motivated it (PensieveRemember.ts:56,
// `const MaximumDraftBytes = 512 * 1024;`, read further down the file).
func TestConsistencyRequireConstantCasingRenamesAFileLocalConstant(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantId     string
		wantSource string
	}{
		{
			"declaration and every read",
			"const MaximumDraftBytes = 512 * 1024;\n" +
				"export function isTooLong(text: string): boolean {\n" +
				"    return text.length > MaximumDraftBytes;\n}\n" +
				"export const Limits = [MaximumDraftBytes, MaximumDraftBytes / 2];\n",
			"requireCamelCaseLocal",
			"const maximumDraftBytes = 512 * 1024;\n" +
				"export function isTooLong(text: string): boolean {\n" +
				"    return text.length > maximumDraftBytes;\n}\n" +
				"export const Limits = [maximumDraftBytes, maximumDraftBytes / 2];\n",
		},
		{
			// The key is the property's name and is somebody else's contract, so it survives.
			"a shorthand property keeps its key",
			"const OrderColumns = [1, 2];\nexport const Table = { OrderColumns };\n",
			"requireCamelCaseLocal",
			"const orderColumns = [1, 2];\nexport const Table = { OrderColumns: orderColumns };\n",
		},
		{
			// Spelled the same and not the constant: a property read, a type-literal key, an object
			// key. Only the checker can tell these from a reference.
			"same-spelled properties and keys are left alone",
			"const OrderColumns = [1];\n" +
				"export function count(row: { OrderColumns: number }): number {\n" +
				"    return row.OrderColumns + OrderColumns.length;\n}\n" +
				"export const Shape = { OrderColumns: 1 };\n",
			"requireCamelCaseLocal",
			"const orderColumns = [1];\n" +
				"export function count(row: { OrderColumns: number }): number {\n" +
				"    return row.OrderColumns + orderColumns.length;\n}\n" +
				"export const Shape = { OrderColumns: 1 };\n",
		},
		{
			// An inner binding with the same spelling shadows the constant inside its function, so
			// those reads are not the constant's and must not move.
			"a shadowing binding keeps its reads",
			"const OrderColumns = [1];\n" +
				"export function inner(): number {\n" +
				"    let OrderColumns = 2;\n    OrderColumns += 1;\n    return { OrderColumns }.OrderColumns;\n}\n" +
				"export const First = { OrderColumns };\n",
			"requireCamelCaseLocal",
			"const orderColumns = [1];\n" +
				"export function inner(): number {\n" +
				"    let OrderColumns = 2;\n    OrderColumns += 1;\n    return { OrderColumns }.OrderColumns;\n}\n" +
				"export const First = { OrderColumns: orderColumns };\n",
		},
		{
			"a type query is a reference too",
			"const DefaultOptions = { retries: 3 };\nexport type Options = typeof DefaultOptions;\n" +
				"export const Retries = DefaultOptions.retries;\n",
			"requireCamelCaseLocal",
			"const defaultOptions = { retries: 3 };\nexport type Options = typeof defaultOptions;\n" +
				"export const Retries = defaultOptions.retries;\n",
		},
		{
			// The function clause, when the constant is not exported, is file-local too.
			"a file-local function",
			"const FormatNumber = function (value: number) { return String(value); };\n" +
				"export const Formatted = FormatNumber(1);\n",
			"requireCamelCaseFunction",
			"const formatNumber = function (value: number) { return String(value); };\n" +
				"export const Formatted = formatNumber(1);\n",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, ConsistencyRequireConstantCasing, constantCasingFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, testCase.wantId)
			rule_testing.ExpectFixedSource(t, result, testCase.wantSource)
		})
	}
}

// Every refusal still reports. The fix is withheld where the rule cannot prove the rename keeps the
// program's meaning, and the author renames by hand.
func TestConsistencyRequireConstantCasingWithholdsAnUnsafeRename(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		{
			// The new name is taken, so the rename would capture or be captured.
			"the camelCase name already exists",
			"const OrderColumns = [1];\nconst orderColumns = 2;\nexport const Total = OrderColumns.length + orderColumns;\n",
			[]string{"requireCamelCaseLocal"},
		},
		{
			// Taken further in: any identifier spelled that way, even a parameter in another scope.
			"the camelCase name exists in a nested scope",
			"const OrderColumns = [1];\nexport function f(orderColumns: number): number {\n" +
				"    return orderColumns + OrderColumns.length;\n}\n",
			[]string{"requireCamelCaseLocal"},
		},
		{
			// A default export through a clause is not an export of the spelling, so the constant is
			// file-local; the clause still names it, and this fix declines to rewrite a clause.
			"named by a default-export clause",
			"const OrderColumns = [1];\nexport { OrderColumns as default };\n",
			[]string{"requireCamelCaseLocal"},
		},
		{
			// A type alias merged into the same name answers to the old spelling in type positions.
			"a type shares the name",
			"const OrderColumns = [1];\ntype OrderColumns = number[];\n" +
				"export function first(columns: OrderColumns): number[] {\n    return columns.concat(OrderColumns);\n}\n",
			[]string{"requireCamelCaseLocal"},
		},
		{
			// A legal identifier that would redefine a built-in every reader assumes.
			"the camelCase name is a built-in value",
			"const Undefined = 1;\nexport const Total = Undefined + 1;\n",
			[]string{"requireCamelCaseLocal"},
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, ConsistencyRequireConstantCasing, constantCasingFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
			for _, diagnostic := range result.Diagnostics {
				if len(diagnostic.Fixes) != 0 {
					t.Fatalf("expected no fix, got %q", diagnostic.Fixes[0].Text)
				}
			}
		})
	}
}

// An exported constant is never renamed by the fix: its references cross files.
func TestConsistencyRequireConstantCasingDoesNotRenameAnExport(t *testing.T) {
	t.Parallel()

	result := rule_testing.RunTyped(t, ConsistencyRequireConstantCasing, constantCasingFile,
		"export const FormatNumber = (value: number) => String(value);\n")
	rule_testing.ExpectFindings(t, result, "requireCamelCaseFunction")
	if len(result.Diagnostics[0].Fixes) != 0 {
		t.Fatalf("expected no fix on an export, got %q", result.Diagnostics[0].Fixes[0].Text)
	}
}

// An exported camelCase constant nothing imports is told to drop the export; one something imports,
// or might, keeps the PascalCase advice. The shape is MediumConversationSource.ts:73
// (`defaultMediumArchiveDirectory`, exported and read only in its own file) against
// StructureLinter.ts:66 (`printConfigFlagName`, imported by Structure.ts).
func TestConsistencyRequireConstantCasingTellsAnUnimportedExportToDropTheExport(t *testing.T) {
	t.Parallel()

	const subject = "source/MediumConversationSource.ts"
	const subjectText = "export const defaultMediumArchiveDirectory = '/archive';\n" +
		"export const ArchivePath = defaultMediumArchiveDirectory + '/medium';\n"

	cases := []struct {
		name     string
		consumer string
		wantId   string
	}{
		{"no importer at all", "", "dropUnimportedExport"},
		{"an importer that takes a different name", "import { ArchivePath } from './MediumConversationSource';\nexport const Path = ArchivePath;\n", "dropUnimportedExport"},
		{"an import run only for its effects", "import './MediumConversationSource';\n", "dropUnimportedExport"},
		{"a named import", "import { defaultMediumArchiveDirectory } from './MediumConversationSource';\nexport const Path = defaultMediumArchiveDirectory;\n", "requirePascalCaseExported"},
		{"an aliased import", "import { defaultMediumArchiveDirectory as directory } from './MediumConversationSource';\nexport const Path = directory;\n", "requirePascalCaseExported"},
		{"a namespace import", "import * as Medium from './MediumConversationSource';\nexport const Path = Medium.ArchivePath;\n", "requirePascalCaseExported"},
		{"a named re-export", "export { defaultMediumArchiveDirectory } from './MediumConversationSource';\n", "requirePascalCaseExported"},
		{"a star re-export", "export * from './MediumConversationSource';\n", "requirePascalCaseExported"},
		{"a dynamic import", "export async function load(): Promise<string> {\n    const medium = await import('./MediumConversationSource');\n    return medium.ArchivePath;\n}\n", "requirePascalCaseExported"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			files := map[string]string{subject: subjectText}
			if testCase.consumer != "" {
				files["source/Consumer.ts"] = testCase.consumer
			}
			result := rule_testing.RunTypedFiles(t, ConsistencyRequireConstantCasing, files, subject)
			rule_testing.ExpectFindings(t, result, testCase.wantId)
		})
	}
}

// A file a test runner or framework loads with no import is never told to drop an export, because
// the absence of an importer proves nothing about it.
func TestConsistencyRequireConstantCasingTrustsAFileLoadedWithoutAnImport(t *testing.T) {
	t.Parallel()

	const subject = "source/Medium.test.ts"
	result := rule_testing.RunTypedFiles(t, ConsistencyRequireConstantCasing,
		map[string]string{subject: "export const fixtureDirectory = '/archive';\n"}, subject)
	rule_testing.ExpectFindings(t, result, "requirePascalCaseExported")
}

// A constant exported by a local `export { ... }` clause is exported, matched to the clause through
// the checker. The false finding this fixes: PensieveCommandBoundary.ts declares
// `const IdentitySafePensieveCommandPaths = new Set([...])` and exports it on line 39 with
// `export { IdentitySafePensieveCommandPaths };`, and the rule told it to become camelCase.
func TestConsistencyRequireConstantCasingCountsALocalExportClause(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		{
			"PascalCase exported through a clause is clean",
			"const IdentitySafePensieveCommandPaths = new Set(['a']);\nexport { IdentitySafePensieveCommandPaths };\n",
			nil,
		},
		{
			"PascalCase exported under another name is clean",
			"const OrderColumns = [1];\nexport { OrderColumns as Columns };\n",
			nil,
		},
		{
			// Nothing imports it here, so the exported judgment is the drop-export one.
			"camelCase exported through a clause gets the exported judgment",
			"const orderColumns = [1];\nexport { orderColumns };\n",
			[]string{"dropUnimportedExport"},
		},
		{
			// A default export never reaches anyone by this spelling, so camelCase is right.
			"camelCase named by export default is clean",
			"const orderColumns = [1];\nexport default orderColumns;\n",
			nil,
		},
		{
			"PascalCase named by export default is file-local",
			"const OrderColumns = [1];\nexport default OrderColumns;\n",
			[]string{"requireCamelCaseLocal"},
		},
		{
			// The clause re-exports another module's name; the same spelling here is a different
			// binding and stays file-local.
			"a re-export of another module's same-named binding does not count",
			"import { first } from './Other';\nconst OrderColumns = [first];\nexport const Total = OrderColumns.length;\n" +
				"export { OrderColumns as Columns } from './Other';\n",
			[]string{"requireCamelCaseLocal"},
		},
		{
			// Matched by binding, not spelling: the clause names the outer constant, so the inner
			// same-spelled constant is still file-local.
			"a same-spelled inner binding is not the exported one",
			"const OrderColumns = [1];\nexport function f(): number {\n    const OrderColumns = 2;\n    return OrderColumns;\n}\n" +
				"export { OrderColumns };\n",
			[]string{"requireCamelCaseLocal"},
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedFiles(t, ConsistencyRequireConstantCasing, map[string]string{
				"source/Thing.ts": testCase.sourceText,
				"source/Other.ts": "export const first = 1;\nexport const OrderColumns = [2];\n",
			}, "source/Thing.ts")
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// The importer check reads the name the clause exports, not the local one. Here the consumer imports
// `Columns`, the clause's name for `orderColumns`, so it is imported and keeps the PascalCase advice.
func TestConsistencyRequireConstantCasingChecksTheExportedNameForImporters(t *testing.T) {
	t.Parallel()

	result := rule_testing.RunTypedFiles(t, ConsistencyRequireConstantCasing, map[string]string{
		"source/Thing.ts":    "const orderColumns = [1];\nexport { orderColumns as Columns };\n",
		"source/Consumer.ts": "import { Columns } from './Thing';\nexport const Total = Columns.length;\n",
	}, "source/Thing.ts")
	rule_testing.ExpectFindings(t, result, "requirePascalCaseExported")
}

// TestConsistencyRequireConstantCasingLeavesNextRouteContractsAlone covers the floor
// `nextjs.IsRouteContractExport` sets with no options at all.
//
// Every clause of this rule would advise on a contract export: PascalCase for `revalidate`, camelCase
// for an arrow-function `POST`, and dropping the export from a `metadata` nothing imports. Each silent
// case is paired with the same text where Next does not read it.
func TestConsistencyRequireConstantCasingLeavesNextRouteContractsAlone(t *testing.T) {
	t.Parallel()

	const routeFile = "/repository/app/api/report/route.ts"
	const pageFile = "/repository/app/blog/[slug]/page.tsx"
	cases := []struct {
		name       string
		fileName   string
		sourceText string
		wantIds    []string
	}{
		{"revalidate in a page", pageFile, "export const revalidate = 60;", nil},
		{"revalidate outside a route", constantCasingFile, "export const revalidate = 60;", []string{"requirePascalCaseExported"}},
		{"an arrow POST in a route", routeFile, "export const POST = async () => new Response('');", nil},
		{"an arrow POST outside a route", constantCasingFile, "export const POST = async () => new Response('');", []string{"requireCamelCaseFunction"}},
		{"unstable_instant in a page", pageFile, "export const unstable_instant = false;", nil},
		// Contracts differ by file: a route reads no metadata, so the export there is ours to judge.
		{"metadata in a route is not a contract", routeFile, "export const metadata = { title: 'Title' };", []string{"requirePascalCaseExported"}},
		// Only the export is the contract; a file-local constant with the spelling is judged as one.
		{"an unexported POST is not read by Next", routeFile, "const POST = async () => new Response('');\nconsole.log(POST);", []string{"requireCamelCaseFunction"}},
		{"a local PascalCase is still judged", pageFile, "const Revalidate = 60;\nconsole.log(Revalidate);", []string{"requireCamelCaseLocal"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, ConsistencyRequireConstantCasing, testCase.fileName, testCase.sourceText)
			if len(testCase.wantIds) == 0 {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// A class that extends a built-in container holds data the way the container does, so an exported
// instance of it keeps PascalCase. Structure's GraphQL generator emits one per operation, `export const
// XDocument = new TypedDocumentString(...)`, beside `class TypedDocumentString extends String` (#w26f1b0).
// The heritage is read through the checker for an imported class and from this file's own classes without
// one, transitively; a class that extends anything else is still an instance that acts.
func TestConsistencyRequireConstantCasingExemptsAClassExtendingAContainer(t *testing.T) {
	t.Parallel()

	const document = "export class TypedDocumentString<TResult, TVariables> extends String {\n" +
		"    constructor(private value: string) { super(value); }\n}\n"
	for _, testCase := range []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		{"a same-file class extending String", document + "export const AccountDocument = new TypedDocumentString<1, 2>('query { a }');\n", nil},
		{"a same-file class extending Map", "class Registry extends Map<string, number> {}\nexport const Registries = new Registry();\n", nil},
		{"a container reached through two classes", "class Base extends Set<string> {}\nclass Tags extends Base {}\nexport const AllTags = new Tags();\n", nil},
		{"a class extending a class of ours", "class Service {}\nclass NetworkService extends Service {}\nexport const NetworkServiceSingleton = new NetworkService();\n", []string{"requireCamelCaseInstance"}},
		{"a class extending nothing", "class NetworkService {}\nexport const NetworkServiceSingleton = new NetworkService();\n", []string{"requireCamelCaseInstance"}},
	} {
		t.Run("untyped, "+testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.Run(t, ConsistencyRequireConstantCasing, constantCasingFile, testCase.sourceText), testCase.wantIds...)
		})
		t.Run("typed, "+testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, ConsistencyRequireConstantCasing, constantCasingFile, testCase.sourceText), testCase.wantIds...)
		})
	}

	// Imported, which only the checker can follow.
	for _, testCase := range []struct {
		name    string
		library string
		wantIds []string
	}{
		{"an imported class extending String", document, nil},
		{"an imported class extending a class of ours", "export class Service {}\nexport class TypedDocumentString<TResult, TVariables> extends Service {}\n", []string{"requireCamelCaseInstance"}},
	} {
		t.Run("imported, "+testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedFiles(t, ConsistencyRequireConstantCasing, map[string]string{
				"source/Documents.ts":  testCase.library,
				"source/Operations.ts": "import { TypedDocumentString } from './Documents';\nexport const AccountDocument = new TypedDocumentString<1, 2>('query { a }');\nexport function use(): unknown { return AccountDocument; }\n",
			}, "source/Operations.ts")
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}
