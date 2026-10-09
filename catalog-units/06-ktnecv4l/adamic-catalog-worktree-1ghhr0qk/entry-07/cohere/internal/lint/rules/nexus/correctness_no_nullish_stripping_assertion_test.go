package nexus

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/rules/typescript"
	"github.com/system-inc/cohere/internal/lint/testing"
)

const correctnessNoNullishStrippingAssertionFile = "/repository/source/CorrectnessNoNullishStrippingAssertion.ts"

// correctnessNoNullishStrippingAssertionTsConfig is the fixture program's compiler options, matching
// ahra's on the two that decide this rule's verdicts: `noUncheckedIndexedAccess`, which is what puts
// `undefined` into `options.days` and `parts[1]`, and the DOM library, which types `event.target` as
// `EventTarget | null`. The harness's default has neither, and without them every real site below
// would be silent for a reason that is not the rule's.
const correctnessNoNullishStrippingAssertionTsConfig = `{
	"compilerOptions": {
		"strict": true,
		"noUncheckedIndexedAccess": true,
		"target": "ES2022",
		"lib": ["ES2022", "DOM"],
		"moduleDetection": "force",
		"types": []
	},
	"include": ["**/*.ts", "**/*.tsx"]
}`

// correctnessNoNullishStrippingAssertionPrelude declares every name the cases use, so the checker
// resolves each one and no verdict rests on an unresolved identifier. The flag types model Structure's
// `source/command-line/CommandArguments.ts` on 2026-10-03.
var correctnessNoNullishStrippingAssertionPrelude = strings.Join([]string{
	"type PendingCommandArgumentFlagValueType = { pendingCommandArgumentFlagValue: true; candidateValue: string; candidateIndex: number };",
	"type CommandArgumentFlagsType = Record<string, string | boolean | PendingCommandArgumentFlagValueType>;",
	"declare function numberFlag(flags: CommandArgumentFlagsType, key: string, fallback: number): number;",
	"declare function required<T>(value: T, message: string): NonNullable<T>;",
	"declare function use(value: unknown): void;",
	"declare const maybe: string | undefined;",
	"declare const nullable: string | null;",
	"declare const cache: Map<string, string>;",
	"declare const parts: string[];",
	"declare const record: Record<string, unknown>;",
	"declare const formData: FormData;",
	"",
}, "\n")

func correctnessNoNullishStrippingAssertionSource(lines ...string) string {
	return correctnessNoNullishStrippingAssertionPrelude + strings.Join(lines, "\n") + "\n"
}

// correctnessNoNullishStrippingAssertionRunRule builds a one-file program under the fixture tsconfig
// and runs one rule on it. A program per case, so no two cases share interned types: the partition
// with the style rule is decided by type identity, and identity across files is the instrument bug
// that rule's doc comment records.
func correctnessNoNullishStrippingAssertionRunRule(t *testing.T, subject rule.Rule, sourceText string) rule_testing.Result {
	t.Helper()
	return rule_testing.RunTypedFilesWithSetup(t, subject, map[string]string{
		correctnessNoNullishStrippingAssertionFile: sourceText,
	}, correctnessNoNullishStrippingAssertionFile, func(directory string) {
		configPath := filepath.Join(directory, "tsconfig.json")
		if err := os.WriteFile(configPath, []byte(correctnessNoNullishStrippingAssertionTsConfig), 0o644); err != nil {
			t.Fatalf("writing the fixture tsconfig: %v", err)
		}
	})
}

func correctnessNoNullishStrippingAssertionRun(t *testing.T, sourceText string) rule_testing.Result {
	t.Helper()
	return correctnessNoNullishStrippingAssertionRunRule(t, CorrectnessNoNullishStrippingAssertion, sourceText)
}

// correctnessNoNullishStrippingAssertionSpans is every finding's text in source order. The harness
// trims the subject file before writing it, so spans are read from the program's own text.
func correctnessNoNullishStrippingAssertionSpans(result rule_testing.Result) []string {
	diagnostics := append(result.Diagnostics[:0:0], result.Diagnostics...)
	sort.SliceStable(diagnostics, func(left int, right int) bool {
		return diagnostics[left].Range.Pos() < diagnostics[right].Range.Pos()
	})
	text := result.SourceFile.Text()
	spans := make([]string, 0, len(diagnostics))
	for _, diagnostic := range diagnostics {
		spans = append(spans, text[diagnostic.Range.Pos():diagnostic.Range.End()])
	}
	return spans
}

// correctnessNoNullishStrippingAssertionExpect asserts the findings by message id and by the exact
// text each one points at, in source order, and that none carries a fix.
func correctnessNoNullishStrippingAssertionExpect(t *testing.T, result rule_testing.Result, wantSpans ...string) {
	t.Helper()
	wantIds := make([]string, 0, len(wantSpans))
	for range wantSpans {
		wantIds = append(wantIds, correctnessNoNullishStrippingAssertionId)
	}
	rule_testing.ExpectFindings(t, result, wantIds...)
	for _, diagnostic := range result.Diagnostics {
		if len(diagnostic.Fixes) != 0 || len(diagnostic.Suggestions) != 0 {
			t.Fatalf("expected no fix and no suggestion, got %d and %d", len(diagnostic.Fixes), len(diagnostic.Suggestions))
		}
	}
	got := correctnessNoNullishStrippingAssertionSpans(result)
	if strings.Join(got, "\x00") != strings.Join(wantSpans, "\x00") {
		t.Fatalf("findings point at %q, want %q", got, wantSpans)
	}
}

// correctnessNoNullishStrippingAssertionAnalytics models `getDateRange` and `getDaysLabel` in ahra's
// `modules/google/analytics/AnalyticsApi.ts` on 2026-10-03, the site the research cited. `fixed` reads
// the day count through `numberFlag`, which is what the flag helpers exist for. The `options.start as
// string` reads stay in both: they sit under `if(options.start && options.end)`, which has already
// removed the undefined, so they are the nearest legitimate shape and must stay silent.
func correctnessNoNullishStrippingAssertionAnalytics(fixed bool) string {
	days := "    const days = parseInt(options.days as string, 10) || defaultDays;"
	if fixed {
		days = "    const days = numberFlag(options, 'days', defaultDays);"
	}
	return correctnessNoNullishStrippingAssertionSource(
		"type AnalyticsOptionsType = CommandArgumentFlagsType;",
		"interface AnalyticsDateRangeInterface { startDate: string; endDate: string }",
		"export function getDateRange(options: AnalyticsOptionsType, defaultDays = 7): AnalyticsDateRangeInterface {",
		"    if(options.start && options.end) {",
		"        return { startDate: options.start as string, endDate: options.end as string };",
		"    }",
		days,
		"    return { startDate: `${days}daysAgo`, endDate: 'today' };",
		"}",
		"export function getDaysLabel(options: AnalyticsOptionsType, defaultDays = 7): string {",
		"    if(options.start && options.end) return `${options.start as string} to ${options.end as string}`;",
		days,
		"    return `last ${days} days`;",
		"}",
	)
}

// correctnessNoNullishStrippingAssertionReactions models the assignee block of `stampTemplate` in
// ahra's `modules/os/sensation/AhraOsReactions.ts` on 2026-10-03. The outer read already rules out
// `null` with a truthiness test and stays silent in both forms; the inner one tests only `typeof ===
// 'object'`, which `null` passes, so a nerve configured with `configuration: null` hands `null` on as
// a record. `fixed` adds the missing `!== null`.
func correctnessNoNullishStrippingAssertionReactions(fixed bool) string {
	inner := "            typeof assigneeFromConfiguration.configuration === 'object'"
	if fixed {
		inner += " && assigneeFromConfiguration.configuration !== null"
	}
	return correctnessNoNullishStrippingAssertionSource(
		"export function stampTemplate(configuration: Record<string, unknown>) {",
		"    const assigneeFromConfiguration =",
		"        configuration.assigneeFrom && typeof configuration.assigneeFrom === 'object'",
		"            ? (configuration.assigneeFrom as Record<string, unknown>)",
		"            : null;",
		"    if(assigneeFromConfiguration) {",
		"        const resolverConfiguration =",
		inner,
		"                ? (assigneeFromConfiguration.configuration as Record<string, unknown>)",
		"                : assigneeFromConfiguration;",
		"        return resolverConfiguration;",
		"    }",
		"    return null;",
		"}",
	)
}

// correctnessNoNullishStrippingAssertionDialogMenu models `onOpenAutoFocus` in Structure's
// `source/components/dialogs/DialogMenu.tsx` on 2026-10-03, the shape shared by `PopoverMenu`,
// `ResponsivePopoverDrawerMenu` and `SourcesAccordion`. `fixed` narrows with `instanceof`.
func correctnessNoNullishStrippingAssertionDialogMenu(fixed bool) string {
	body := []string{
		"    const dialogElement = event.target as HTMLElement;",
		"    const inputElement = dialogElement.querySelector('input');",
	}
	if fixed {
		body = []string{
			"    if(!(event.target instanceof HTMLElement)) return;",
			"    const inputElement = event.target.querySelector('input');",
		}
	}
	lines := []string{"export function onOpenAutoFocus(event: Event) {"}
	lines = append(lines, body...)
	lines = append(lines,
		"    setTimeout(function() {",
		"        inputElement?.focus();",
		"    }, 25);",
		"}",
	)
	return correctnessNoNullishStrippingAssertionSource(lines...)
}

// The real analytics flags, before the fix: both day-count reads are reported, the narrowed range
// reads are not.
func TestCorrectnessNoNullishStrippingAssertionFiresOnAnalyticsDays(t *testing.T) {
	t.Parallel()

	result := correctnessNoNullishStrippingAssertionRun(t, correctnessNoNullishStrippingAssertionAnalytics(false))
	correctnessNoNullishStrippingAssertionExpect(t, result, "options.days as string", "options.days as string")
}

func TestCorrectnessNoNullishStrippingAssertionStaysSilentOnFixedAnalyticsDays(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectClean(t, correctnessNoNullishStrippingAssertionRun(t, correctnessNoNullishStrippingAssertionAnalytics(true)))
}

// The real `typeof null === 'object'` read, before the fix: the inner assertion is reported, the outer
// one, already guarded by truthiness, is not.
func TestCorrectnessNoNullishStrippingAssertionFiresOnReactionsAssignee(t *testing.T) {
	t.Parallel()

	result := correctnessNoNullishStrippingAssertionRun(t, correctnessNoNullishStrippingAssertionReactions(false))
	correctnessNoNullishStrippingAssertionExpect(t, result, "assigneeFromConfiguration.configuration as Record<string, unknown>")
}

func TestCorrectnessNoNullishStrippingAssertionStaysSilentOnFixedReactionsAssignee(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectClean(t, correctnessNoNullishStrippingAssertionRun(t, correctnessNoNullishStrippingAssertionReactions(true)))
}

// The real dialog focus handler, before the fix: `event.target` is `EventTarget | null`.
func TestCorrectnessNoNullishStrippingAssertionFiresOnDialogMenu(t *testing.T) {
	t.Parallel()

	result := correctnessNoNullishStrippingAssertionRun(t, correctnessNoNullishStrippingAssertionDialogMenu(false))
	correctnessNoNullishStrippingAssertionExpect(t, result, "event.target as HTMLElement")
}

func TestCorrectnessNoNullishStrippingAssertionStaysSilentOnFixedDialogMenu(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectClean(t, correctnessNoNullishStrippingAssertionRun(t, correctnessNoNullishStrippingAssertionDialogMenu(true)))
}

func TestCorrectnessNoNullishStrippingAssertionFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		lines []string
		want  []string
	}{
		{"a Map read narrowed to a literal union", []string{
			"export const state = cache.get('job') as 'Ready' | 'Stale';",
		}, []string{"cache.get('job') as 'Ready' | 'Stale'"}},
		{"the angle-bracket spelling", []string{
			"export const state = <'Ready' | 'Stale'>cache.get('job');",
		}, []string{"<'Ready' | 'Stale'>cache.get('job')"}},
		{"a DOM lookup that may find nothing", []string{
			"export const canvas = document.getElementById('canvas') as HTMLCanvasElement;",
		}, []string{"document.getElementById('canvas') as HTMLCanvasElement"}},
		{"a form field, null when absent and a File when uploaded (PostSearch)", []string{
			"export const searchTerm = formData.get('searchTerm') as string;",
		}, []string{"formData.get('searchTerm') as string"}},
		{"an unchecked array element (PhiSocialCommandLineInterface)", []string{
			"export const status = parts[1] as 'Draft' | 'Approved';",
		}, []string{"parts[1] as 'Draft' | 'Approved'"}},
		{"an optional chain (ContactListPage)", []string{
			"declare const urlParameters: { contactListId?: string } | undefined;",
			"export const contactListId = urlParameters?.contactListId as 'list';",
		}, []string{"urlParameters?.contactListId as 'list'"}},
		{"a fallback right after it, the as twin of a non-null asserted coalesce (Markdown)", []string{
			"type ButtonVariantType = 'A' | 'B';",
			"export const variant = (maybe as ButtonVariantType) ?? 'A';",
		}, []string{"maybe as ButtonVariantType"}},
		{"unknown narrowed past null leaves undefined behind (FinanceSqliteDatabase)", []string{
			"export const cents = record.value_cents === null ? null : (record.value_cents as number);",
		}, []string{"record.value_cents as number"}},
		{"null removed, undefined never present", []string{
			"export const value = nullable as 'on' | 'off';",
		}, []string{"nullable as 'on' | 'off'"}},
		{"an asserted type parameter constrained to non-nullish", []string{
			"export function read<KeyType extends string>(): KeyType {",
			"    return maybe as KeyType;",
			"}",
		}, []string{"maybe as KeyType"}},
		{"an assertion read for a member that is then written (AhraOsBplist)", []string{
			"type EncoderObjectType = { kind: 'array'; items: number[] } | { kind: 'string'; value: string };",
			"declare const objects: EncoderObjectType[];",
			"export function fill(self: number) {",
			"    (objects[self] as { kind: 'array'; items: number[] }).items = [];",
			"}",
		}, []string{"objects[self] as { kind: 'array'; items: number[] }"}},
		{"widened past the original while dropping undefined", []string{
			"export const value = maybe as string | number;",
		}, []string{"maybe as string | number"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := correctnessNoNullishStrippingAssertionRun(t, correctnessNoNullishStrippingAssertionSource(testCase.lines...))
			correctnessNoNullishStrippingAssertionExpect(t, result, testCase.want...)
		})
	}
}

func TestCorrectnessNoNullishStrippingAssertionStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		lines []string
	}{
		{"the pure removal, which is the style rule's", []string{
			"export const name = maybe as string;",
		}},
		{"the pure removal of a literal union through an alias", []string{
			"type StateType = 'Ready' | 'Stale';",
			"declare const state: StateType | undefined;",
			"export const ready = state as StateType;",
		}},
		{"undefined kept in the asserted type", []string{
			"export const state = maybe as 'Ready' | undefined;",
		}},
		{"null kept in the asserted type", []string{
			"export const value = nullable as 'on' | null;",
		}},
		{"asserting to any", []string{
			"export const value = maybe as any;",
		}},
		{"asserting to unknown, then onward", []string{
			"export const value = maybe as unknown as number;",
		}},
		{"asserting from unknown (FinanceSqliteDatabase's other columns)", []string{
			"export const cents = record.value_cents as number;",
		}},
		{"asserting from any", []string{
			"export const parsed = JSON.parse('{}') as { name: string };",
		}},
		{"narrowed first, so nothing nullish is left (AnalyticsApi's range)", []string{
			"if(maybe !== undefined) {",
			"    use(maybe as 'Ready');",
			"}",
		}},
		{"truthiness guarding the read (AhraOsReactions' outer read)", []string{
			"export const assigneeFrom = record.assigneeFrom && typeof record.assigneeFrom === 'object'",
			"    ? (record.assigneeFrom as Record<string, unknown>)",
			"    : null;",
		}},
		{"the fix through required", []string{
			"export const state = required(cache.get('job'), 'The job was enqueued above.') as 'Ready' | 'Stale';",
		}},
		{"as const on a nullable, a tsc error the rule must still decline", []string{
			"export const name = maybe as const;",
		}},
		{"an assignment target, which writes rather than reads", []string{
			"declare const holder: { value?: string | number };",
			"(holder.value as string) = 'set';",
		}},
		{"an all-nullish placeholder into a type parameter (LinkedInClient's 204 branch)", []string{
			"export function empty<ResponseType>(): ResponseType {",
			"    return null as ResponseType;",
			"}",
		}},
		{"an unconstrained type parameter, which a caller may instantiate with undefined", []string{
			"declare const store: Map<string, unknown>;",
			"export function read<ValueType>(source: Map<string, ValueType>): ValueType {",
			"    return source.get('key') as ValueType;",
			"}",
		}},
		{"an unconstrained type parameter that is not the original's own", []string{
			"export function read<ValueType>(): ValueType {",
			"    return maybe as ValueType;",
			"}",
		}},
		{"an indexed access with no constraint to read", []string{
			"export function pick<ShapeType, KeyType extends keyof ShapeType>(shape: Partial<ShapeType>, key: KeyType): ShapeType[KeyType] {",
			"    return shape[key] as ShapeType[KeyType];",
			"}",
		}},
		{"as never, the escape hatch (FinancePositionCommandLineInterface)", []string{
			"declare function commandArgumentFlagValue(name: string): string | undefined;",
			"type SourceType = 'Manual' | 'Plaid';",
			"export const position: { source: SourceType } = { source: (commandArgumentFlagValue('--source') as never) ?? 'Manual' };",
		}},
		{"a void result, which is ignored rather than absent", []string{
			"declare function callback(): void | Promise<void>;",
			"export const pending = callback() as Promise<void>;",
		}},
		{"a placeholder made from undefined alone", []string{
			"export const placeholder = undefined as 'Ready';",
		}},
		{"asserting to void", []string{
			"export const ignored = maybe as void;",
		}},
		{"satisfies, which asserts nothing", []string{
			"export const name = maybe satisfies string | undefined;",
		}},
		{"a non-null assertion, which is no-non-null-assertion's", []string{
			"export const name = maybe!;",
		}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, correctnessNoNullishStrippingAssertionRun(t, correctnessNoNullishStrippingAssertionSource(testCase.lines...)))
		})
	}
}

// The partition: on every cast, at most one of this rule and `non-nullable-type-assertion-style`
// reports, and each reports the casts it owns. Both rules run on the same program text, one program per
// rule, with the cast that must be reported named for each.
func TestCorrectnessNoNullishStrippingAssertionNeverReportsTheStyleRulesCast(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		line      string
		span      string
		styleOwns bool
		thisOwns  bool
	}{
		{"pure removal", "export const name = maybe as string;", "maybe as string", true, false},
		{"pure removal of null", "export const name = nullable as string;", "nullable as string", true, false},
		{"pure removal through an alias", "type StateType = 'Ready' | 'Stale'; declare const state: StateType | undefined; export const ready = state as StateType;", "state as StateType", true, false},
		{"removal and narrowing", "export const state = maybe as 'Ready';", "maybe as 'Ready'", false, true},
		{"removal and widening", "export const value = maybe as string | number;", "maybe as string | number", false, true},
		{"a Map read narrowed", "export const state = cache.get('job') as 'Ready' | 'Stale';", "cache.get('job') as 'Ready' | 'Stale'", false, true},
		{"undefined kept", "export const state = maybe as string | undefined;", "", false, false},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			source := correctnessNoNullishStrippingAssertionSource(testCase.line)

			thisResult := correctnessNoNullishStrippingAssertionRun(t, source)
			styleResult := correctnessNoNullishStrippingAssertionRunRule(t, typescript.NonNullableTypeAssertionStyle, source)

			if testCase.thisOwns {
				correctnessNoNullishStrippingAssertionExpect(t, thisResult, testCase.span)
			} else {
				rule_testing.ExpectClean(t, thisResult)
			}

			styleSpans := correctnessNoNullishStrippingAssertionSpans(styleResult)
			wantStyleSpans := []string{}
			if testCase.styleOwns {
				wantStyleSpans = []string{testCase.span}
			}
			if strings.Join(styleSpans, "\x00") != strings.Join(wantStyleSpans, "\x00") {
				t.Fatalf("the style rule reported %q, want %q", styleSpans, wantStyleSpans)
			}
		})
	}
}

// The message gives the repairs in the order `no-non-null-assertion` does, since the two rules report
// the same claim through two doors: guard it, fix the type, then `required` with its reason.
func TestCorrectnessNoNullishStrippingAssertionMessageGivesTheRepairsInOrder(t *testing.T) {
	t.Parallel()

	result := correctnessNoNullishStrippingAssertionRun(t, correctnessNoNullishStrippingAssertionAnalytics(false))
	if len(result.Diagnostics) == 0 {
		t.Fatal("want a finding to read the message from")
	}
	text := result.Diagnostics[0].Message.Description
	rest := text
	for _, part := range []string{"guard it", "fix the type", "`stringFlag` or `numberFlag`", "`required(value, 'why it is present')` from `@nexus/source/errors/Assert`"} {
		index := strings.Index(rest, part)
		if index < 0 {
			t.Fatalf("message lacks %q after the parts before it:\n%s", part, text)
		}
		rest = rest[index+len(part):]
	}
}
