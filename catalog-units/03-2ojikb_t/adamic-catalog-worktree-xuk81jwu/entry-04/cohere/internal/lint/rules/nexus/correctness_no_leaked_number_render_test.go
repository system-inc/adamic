package nexus

import (
	"sort"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// A `.tsx` name because every case is JSX.
const correctnessNoLeakedNumberRenderFile = "/repository/source/CorrectnessNoLeakedNumberRender.tsx"

// correctnessNoLeakedNumberRenderPrelude declares every name the cases use, so the checker resolves
// each one and no verdict rests on an unresolved identifier or an `any`.
var correctnessNoLeakedNumberRenderPrelude = strings.Join([]string{
	"declare namespace JSX {",
	"    interface Element {}",
	"    interface IntrinsicElements { [name: string]: Record<string, unknown> }",
	"}",
	"type ReactNode = string | number | boolean | null | undefined | JSX.Element;",
	"declare function Badge(properties: { visible?: unknown }): JSX.Element;",
	"declare function addCommas(value: number): string;",
	"declare const count: number;",
	"declare const optionalCount: number | undefined;",
	"declare const nullableCount: number | null;",
	"declare const big: bigint;",
	"declare const flag: boolean;",
	"declare const label: string;",
	"declare const labelOrCount: string | number;",
	"declare const children: ReactNode;",
	"declare const columns: 1 | 2 | 3;",
	"declare const zeroOrOne: 0 | 1;",
	"declare const countOrFalse: number | false;",
	"declare const items: string[];",
	"declare const loose: any;",
	"declare const opaque: unknown;",
	"type Pixels = number & { readonly unit: 'Pixels' };",
	"declare const width: Pixels;",
	"enum Level { Off, Low, High }",
	"enum Size { Small = 1, Large = 2 }",
	"declare const level: Level;",
	"declare const size: Size;",
	"",
}, "\n")

func correctnessNoLeakedNumberRenderSource(lines ...string) string {
	return correctnessNoLeakedNumberRenderPrelude + strings.Join(lines, "\n") + "\n"
}

func correctnessNoLeakedNumberRenderRun(t *testing.T, sourceText string) rule_testing.Result {
	t.Helper()
	return rule_testing.RunTyped(t, CorrectnessNoLeakedNumberRender, correctnessNoLeakedNumberRenderFile, sourceText)
}

// correctnessNoLeakedNumberRenderExpect asserts the findings in source order by the text each one
// points at. The harness trims the subject file before writing it, so spans are read from the
// program's own text.
func correctnessNoLeakedNumberRenderExpect(t *testing.T, result rule_testing.Result, want ...string) {
	t.Helper()
	diagnostics := append(result.Diagnostics[:0:0], result.Diagnostics...)
	sort.SliceStable(diagnostics, func(left int, right int) bool {
		return diagnostics[left].Range.Pos() < diagnostics[right].Range.Pos()
	})
	wantIds := make([]string, 0, len(want))
	for range want {
		wantIds = append(wantIds, correctnessNoLeakedNumberRenderId)
	}
	rule_testing.ExpectFindings(t, result, wantIds...)
	text := result.SourceFile.Text()
	var got []string
	for _, diagnostic := range diagnostics {
		got = append(got, text[diagnostic.Range.Pos():diagnostic.Range.End()])
		if len(diagnostic.Fixes) != 0 || len(diagnostic.Suggestions) != 0 {
			t.Fatalf("expected no fix and no suggestion, got %d and %d", len(diagnostic.Fixes), len(diagnostic.Suggestions))
		}
	}
	if len(got) != len(want) {
		t.Fatalf("expected %d findings %q, got %d %q", len(want), want, len(got), got)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("finding %d is %q, want %q (all: %q)", index, got[index], want[index], got)
		}
	}
}

// correctnessNoLeakedNumberRenderPagination models the records count in Structure's
// `source/components/navigation/pagination/PaginationControls.tsx` on 2026-10-03, trimmed to the
// property it reads and the block around it. `condition` is what guards the records line: the file
// as it stands passes `properties.itemsTotal`.
func correctnessNoLeakedNumberRenderPagination(condition string) string {
	return correctnessNoLeakedNumberRenderSource(
		"export interface PaginationControlsBaseProperties {",
		"    page: number;",
		"    itemsTotal?: number;",
		"    itemsPerPage?: number;",
		"}",
		"export function PaginationControls(properties: PaginationControlsBaseProperties) {",
		"    return (",
		"        <div className=\"flex items-center gap-2\">",
		"            {"+condition+" && (",
		"                <p className=\"me-10\">",
		"                    {addCommas(properties.itemsTotal ?? 0)} records",
		"                </p>",
		"            )}",
		"            <p>Show</p>",
		"        </div>",
		"    );",
		"}",
	)
}

// correctnessNoLeakedNumberRenderCheckboxGrid models the cap note at the foot of Structure's
// `source/components/forms/fields/multiple-checkbox-grid/FieldInputMultipleCheckboxGrid.tsx` on
// 2026-10-03.
func correctnessNoLeakedNumberRenderCheckboxGrid(condition string) string {
	return correctnessNoLeakedNumberRenderSource(
		"export interface FieldInputMultipleCheckboxGridProperties {",
		"    maximumSelectionsPerRow?: number;",
		"}",
		"export function FieldInputMultipleCheckboxGrid(properties: FieldInputMultipleCheckboxGridProperties) {",
		"    return (",
		"        <div>",
		"            <table />",
		"            {"+condition+" && (",
		"                <p className=\"text-xs content--4\">",
		"                    Maximum {properties.maximumSelectionsPerRow} selection",
		"                    {properties.maximumSelectionsPerRow !== 1 ? 's' : ''} per row",
		"                </p>",
		"            )}",
		"        </div>",
		"    );",
		"}",
	)
}

// correctnessNoLeakedNumberRenderWebSockets models the statistics block of Structure's
// `source/ops/developers/web-sockets/WebSocketsPage.tsx` on 2026-10-03, with `connectedAt` typed as
// `WebSocketSharedWorkerTypes.ts` types it.
func correctnessNoLeakedNumberRenderWebSockets(condition string) string {
	return correctnessNoLeakedNumberRenderSource(
		"interface WebSocketStatisticsInterface { messagesSent: number; connectedAt: number | null }",
		"declare const information: { statistics: WebSocketStatisticsInterface | null };",
		"declare const formattedTimes: { connectedSince: string };",
		"export function WebSocketsPage() {",
		"    return (",
		"        <div>",
		"            {information.statistics && (",
		"                <>",
		"                    {"+condition+" && (",
		"                        <>",
		"                            <p className=\"font-bold\">Connected Since</p>",
		"                            <p>{formattedTimes.connectedSince}</p>",
		"                        </>",
		"                    )}",
		"                </>",
		"            )}",
		"        </div>",
		"    );",
		"}",
	)
}

// The three real sites as they stand: each operand is reported, and nothing else is.
func TestCorrectnessNoLeakedNumberRenderFiresOnTheRealSites(t *testing.T) {
	t.Parallel()

	correctnessNoLeakedNumberRenderExpect(t,
		correctnessNoLeakedNumberRenderRun(t, correctnessNoLeakedNumberRenderPagination("properties.itemsTotal")),
		"properties.itemsTotal")
	correctnessNoLeakedNumberRenderExpect(t,
		correctnessNoLeakedNumberRenderRun(t, correctnessNoLeakedNumberRenderCheckboxGrid("properties.maximumSelectionsPerRow")),
		"properties.maximumSelectionsPerRow")
	correctnessNoLeakedNumberRenderExpect(t,
		correctnessNoLeakedNumberRenderRun(t, correctnessNoLeakedNumberRenderWebSockets("information.statistics.connectedAt")),
		"information.statistics.connectedAt")
}

// The same three sites fixed: each condition is a boolean, and nothing fires. The outer
// `information.statistics &&` in the WebSockets page is an object and stays silent in both.
func TestCorrectnessNoLeakedNumberRenderStaysSilentOnTheFixedSites(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectClean(t, correctnessNoLeakedNumberRenderRun(t,
		correctnessNoLeakedNumberRenderPagination("properties.itemsTotal !== undefined")))
	rule_testing.ExpectClean(t, correctnessNoLeakedNumberRenderRun(t,
		correctnessNoLeakedNumberRenderCheckboxGrid("properties.maximumSelectionsPerRow !== undefined && properties.maximumSelectionsPerRow > 0")))
	rule_testing.ExpectClean(t, correctnessNoLeakedNumberRenderRun(t,
		correctnessNoLeakedNumberRenderWebSockets("information.statistics.connectedAt !== null")))
}

func TestCorrectnessNoLeakedNumberRenderFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		lines []string
		want  []string
	}{
		{"a number", []string{"export const view = <p>{count && <Badge />}</p>;"}, []string{"count"}},
		{"an optional number", []string{"export const view = <p>{optionalCount && 'some'}</p>;"}, []string{"optionalCount"}},
		{"a nullable number", []string{"export const view = <p>{nullableCount && <Badge />}</p>;"}, []string{"nullableCount"}},
		{"a length", []string{"export const view = <ul>{items.length && <li>{items[0]}</li>}</ul>;"}, []string{"items.length"}},
		{"a bigint", []string{"export const view = <p>{big && <Badge />}</p>;"}, []string{"big"}},
		{"a literal union that holds zero", []string{"export const view = <p>{zeroOrOne && <Badge />}</p>;"}, []string{"zeroOrOne"}},
		{"a number or false", []string{"export const view = <p>{countOrFalse && <Badge />}</p>;"}, []string{"countOrFalse"}},
		{"a branded number", []string{"export const view = <p>{width && <Badge />}</p>;"}, []string{"width"}},
		{"a numeric enum with a zero member", []string{"export const view = <p>{level && <Badge />}</p>;"}, []string{"level"}},
		{"in a fragment", []string{"export const view = <>{count && <Badge />}</>;"}, []string{"count"}},
		{"in parentheses", []string{"export const view = <p>{(count) && (<Badge />)}</p>;"}, []string{"count"}},
		{"the second operand of a chain", []string{"export const view = <p>{flag && count && <Badge />}</p>;"}, []string{"count"}},
		{"both operands of a chain", []string{"export const view = <p>{count && optionalCount && <Badge />}</p>;"}, []string{"count", "optionalCount"}},
		{"narrowed to a number by the operand before it", []string{"export const view = <p>{optionalCount !== undefined && optionalCount && <Badge />}</p>;"}, []string{"optionalCount"}},
		{"inside a ternary branch", []string{"export const view = <p>{flag ? count && <Badge /> : null}</p>;"}, []string{"count"}},
		{"behind an or", []string{"export const view = <p>{label || count && <Badge />}</p>;"}, []string{"count"}},
		{"the right side of an or in the operand", []string{"export const view = <p>{(flag || count) && <Badge />}</p>;"}, []string{"count"}},
		{"a ternary as the operand", []string{"export const view = <p>{(flag ? count : false) && <Badge />}</p>;"}, []string{"count"}},
		{"a generic constrained to number", []string{
			"export function View<Total extends number>(properties: { total: Total }) {",
			"    return <p>{properties.total && <Badge />}</p>;",
			"}",
		}, []string{"properties.total"}},
		{"nested inside an attribute's element", []string{"export const view = <Badge visible={<p>{count && 'some'}</p>} />;"}, []string{"count"}},
		{"a number known positive only by a guard the type cannot carry", []string{
			"export function View() {",
			"    if(count > 0) {",
			"        return <p>{count && 'some'}</p>;",
			"    }",
			"    return null;",
			"}",
		}, []string{"count"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := correctnessNoLeakedNumberRenderRun(t, correctnessNoLeakedNumberRenderSource(testCase.lines...))
			correctnessNoLeakedNumberRenderExpect(t, result, testCase.want...)
		})
	}
}

func TestCorrectnessNoLeakedNumberRenderStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		lines []string
	}{
		{"a comparison", []string{"export const view = <p>{count > 0 && <Badge />}</p>;"}},
		{"a length compared", []string{"export const view = <ul>{items.length > 0 && <li>{items[0]}</li>}</ul>;"}},
		{"double negation", []string{"export const view = <p>{!!count && <Badge />}</p>;"}},
		{"Boolean()", []string{"export const view = <p>{Boolean(count) && <Badge />}</p>;"}},
		{"a ternary that renders null", []string{"export const view = <p>{count ? <Badge /> : null}</p>;"}},
		{"a boolean", []string{"export const view = <p>{flag && <Badge />}</p>;"}},
		{"a string, whose empty value renders nothing", []string{"export const view = <p>{label && <span>{label}</span>}</p>;"}},
		{"a string or a number, rendered as text by intent", []string{"export const view = <p>{labelOrCount && <span>{labelOrCount}</span>}</p>;"}},
		{"a ReactNode", []string{"export const view = <p>{children && <div>{children}</div>}</p>;"}},
		{"a number literal union with no zero", []string{"export const view = <p>{columns && <Badge />}</p>;"}},
		{"a numeric enum with no zero member", []string{"export const view = <p>{size && <Badge />}</p>;"}},
		{"a non-zero constant", []string{
			"const three = 3;",
			"export const view = <p>{three && <Badge />}</p>;",
		}},
		{"any", []string{"export const view = <p>{loose && <Badge />}</p>;"}},
		{"unknown", []string{"export const view = <p>{opaque && <Badge />}</p>;"}},
		{"a generic with no constraint", []string{
			"export function View<Value>(properties: { value: Value }) {",
			"    return <p>{properties.value && <Badge />}</p>;",
			"}",
		}},
		{"a number on the right of the and, rendered on purpose", []string{"export const view = <p>{flag && count}</p>;"}},
		{"a number on the left of an or, rendered only when truthy", []string{"export const view = <p>{count || 'none'}</p>;"}},
		{"an and that an or catches, so a zero falls through to the fallback", []string{"export const view = <p>{(count && label) || 'none'}</p>;"}},
		{"a number on the left of an or in the operand, returned only when truthy", []string{"export const view = <p>{(count || flag) && <Badge />}</p>;"}},
		{"a ternary's condition", []string{"export const view = <p>{count && flag ? <Badge /> : null}</p>;"}},
		{"an attribute, not a child", []string{"export const view = <Badge visible={count && true} />;"}},
		{"an and outside JSX", []string{"export const total = count && 'some';"}},
		{"a spread child, which spreads rather than renders the value", []string{"export const view = <p>{...(count && items)}</p>;"}},
		{"a number rendered whole", []string{"export const view = <p>{count}</p>;"}},
		{"a template", []string{"export const view = <p>{`${count && 'x'}`}</p>;"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			rule_testing.ExpectClean(t, correctnessNoLeakedNumberRenderRun(t, correctnessNoLeakedNumberRenderSource(testCase.lines...)))
		})
	}
}
