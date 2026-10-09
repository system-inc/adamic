package nexus

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const consistencyNoIsoStringDateCutFile = "/repository/source/ConsistencyNoIsoStringDateCut.ts"

const consistencyNoIsoStringDateCutFormatTimeFile = "/repository/libraries/nexus/source/time/FormatTime.ts"

const consistencyNoIsoStringDateCutTimeZonesFile = "/repository/libraries/nexus/source/time/TimeZones.ts"

// consistencyNoIsoStringDateCutFormatTime models `dateIso8601` in Structure's
// `libraries/nexus/source/time/FormatTime.ts` on 2026-10-03. `fixed` gives the no-zone branch the zone
// too; unfixed, it is the file as it stands, which cuts the ISO string itself.
func consistencyNoIsoStringDateCutFormatTime(fixed bool) string {
	fallback := "    return date.toISOString().split('T')[0] || '';"
	if fixed {
		fallback = "    return dateIso8601(date, 'UTC');"
	}
	return strings.Join([]string{
		"export function dateIso8601(date: Date, timeZone?: string): string {",
		"    if(timeZone) {",
		"        return new Intl.DateTimeFormat('en-CA', { timeZone, year: 'numeric', month: '2-digit', day: '2-digit' }).format(date);",
		"    }",
		fallback,
		"}",
	}, "\n")
}

// consistencyNoIsoStringDateCutTimeZones models `userTimeZone` in Structure's
// `libraries/nexus/source/time/TimeZones.ts`.
var consistencyNoIsoStringDateCutTimeZones = strings.Join([]string{
	"export function userTimeZone(): string {",
	"    return Intl.DateTimeFormat().resolvedOptions().timeZone;",
	"}",
}, "\n")

// consistencyNoIsoStringDateCutPrelude imports the fix and declares every name the cases use, so the
// checker resolves each one and no verdict rests on an unresolved identifier.
var consistencyNoIsoStringDateCutPrelude = strings.Join([]string{
	"import { dateIso8601 } from '../libraries/nexus/source/time/FormatTime';",
	"import { userTimeZone } from '../libraries/nexus/source/time/TimeZones';",
	"declare const days: number;",
	"declare const isoDate: string;",
	"declare function use(...values: unknown[]): void;",
	"interface DiscordMessageInterface { createdAt: Date; content: string }",
	"",
}, "\n")

func consistencyNoIsoStringDateCutSource(lines ...string) string {
	return consistencyNoIsoStringDateCutPrelude + strings.Join(lines, "\n") + "\n"
}

func consistencyNoIsoStringDateCutRun(t *testing.T, sourceText string) rule_testing.Result {
	t.Helper()
	return rule_testing.RunTypedFiles(t, ConsistencyNoIsoStringDateCut, map[string]string{
		consistencyNoIsoStringDateCutFile:           sourceText,
		consistencyNoIsoStringDateCutFormatTimeFile: consistencyNoIsoStringDateCutFormatTime(true),
		consistencyNoIsoStringDateCutTimeZonesFile:  consistencyNoIsoStringDateCutTimeZones,
	}, consistencyNoIsoStringDateCutFile)
}

// consistencyNoIsoStringDateCutExpect asserts the findings' spans in source order. The harness trims
// the subject file before writing it, so spans are read from the program's own text.
func consistencyNoIsoStringDateCutExpect(t *testing.T, result rule_testing.Result, want []string) {
	t.Helper()
	diagnostics := append(result.Diagnostics[:0:0], result.Diagnostics...)
	sort.SliceStable(diagnostics, func(left int, right int) bool {
		return diagnostics[left].Range.Pos() < diagnostics[right].Range.Pos()
	})
	wantIds := make([]string, 0, len(want))
	for range want {
		wantIds = append(wantIds, consistencyNoIsoStringDateCutId)
	}
	sorted := result
	sorted.Diagnostics = diagnostics
	rule_testing.ExpectFindings(t, sorted, wantIds...)
	text := result.SourceFile.Text()
	var got []string
	for _, diagnostic := range diagnostics {
		got = append(got, text[diagnostic.Range.Pos():diagnostic.Range.End()])
		if len(diagnostic.Fixes) != 0 || len(diagnostic.Suggestions) != 0 {
			t.Fatalf("expected no fix and no suggestion, got %d and %d", len(diagnostic.Fixes), len(diagnostic.Suggestions))
		}
	}
	if len(got) != len(want) {
		t.Fatalf("expected %d findings %v, got %d %v", len(want), want, len(got), got)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("finding %d is %q, want %q (all: %q)", index, got[index], want[index], got)
		}
	}
}

// Each real ahra site on 2026-10-03, before the fix (`broken`) and after (`fixed`), one per form.
var consistencyNoIsoStringDateCutRealSites = []struct {
	name   string
	broken []string
	want   []string
	fixed  []string
}{
	{
		// `modules/finance/FinanceCommandLineInterfaceShared.ts:56`, the default as-of for every dated
		// write verb: the local day is meant, and an evening write gets tomorrow.
		name:   "slice: the finance today() helper",
		broken: []string{"export const today = (): string => new Date().toISOString().slice(0, 10);"},
		want:   []string{"new Date().toISOString().slice(0, 10)"},
		fixed:  []string{"export const today = (): string => dateIso8601(new Date(), userTimeZone());"},
	},
	{
		// `app/api/finance/route.ts:48`, local-day arithmetic read as a UTC day.
		name: "slice: the finance feed window, stepped in local days",
		broken: []string{
			"export function recentWindowSinceIso(days: number): string {",
			"    const since = new Date();",
			"    since.setDate(since.getDate() - days);",
			"    return since.toISOString().slice(0, 10);",
			"}",
		},
		want: []string{"since.toISOString().slice(0, 10)"},
		fixed: []string{
			"export function recentWindowSinceIso(days: number): string {",
			"    const since = new Date();",
			"    since.setDate(since.getDate() - days);",
			"    return dateIso8601(since, userTimeZone());",
			"}",
		},
	},
	{
		// `modules/openai/ads/OpenAiAdsInsightsApi.ts:192`, documented as the UTC day on purpose.
		name: "substring: the OpenAI ads UTC day",
		broken: []string{
			"export function isoDateUtc(date: Date): string {",
			"    return date.toISOString().substring(0, 10);",
			"}",
		},
		want: []string{"date.toISOString().substring(0, 10)"},
		fixed: []string{
			"export function isoDateUtc(date: Date): string {",
			"    return dateIso8601(date, 'UTC');",
			"}",
		},
	},
	{
		// `modules/discord/DiscordApi.ts:192`, a message's day for a history listing, zone unstated.
		name: "split index: the Discord history line",
		broken: []string{
			"export function historyLine(message: DiscordMessageInterface): string {",
			"    const date = message.createdAt.toISOString().split('T')[0];",
			"    return `[${date}] ${message.content}`;",
			"}",
		},
		want: []string{"message.createdAt.toISOString().split('T')[0]"},
		fixed: []string{
			"export function historyLine(message: DiscordMessageInterface): string {",
			"    const date = dateIso8601(message.createdAt, userTimeZone());",
			"    return `[${date}] ${message.content}`;",
			"}",
		},
	},
	{
		// `modules/pensieve/PensieveWeeklies.ts:147`, UTC on purpose, anchored at UTC noon.
		name: "destructured split: the Pensieve week start",
		broken: []string{
			"export function getWeekStart(dateString: string): string {",
			"    const date = new Date(dateString + 'T12:00:00Z');",
			"    const monday = new Date(date);",
			"    monday.setUTCDate(date.getUTCDate() - ((date.getUTCDay() + 6) % 7));",
			"    const [mondayDate = ''] = monday.toISOString().split('T');",
			"    return mondayDate;",
			"}",
		},
		want: []string{"monday.toISOString().split('T')"},
		fixed: []string{
			"export function getWeekStart(dateString: string): string {",
			"    const date = new Date(dateString + 'T12:00:00Z');",
			"    const monday = new Date(date);",
			"    monday.setUTCDate(date.getUTCDate() - ((date.getUTCDay() + 6) % 7));",
			"    return dateIso8601(monday, 'UTC');",
			"}",
		},
	},
	{
		// `modules/os/migrations/NewMigration.ts:31`, a migration file stamp, UTC date beside UTC time.
		// The fix keeps the time-only slice, which is not the date part.
		name: "const-held ISO string: the migration stamp",
		broken: []string{
			"export function migrationStamp(): string {",
			"    const iso = new Date().toISOString();",
			"    const datePart = iso.slice(0, 10);",
			"    const timePart = iso.slice(11, 19).replace(/:/g, '-');",
			"    return `${datePart}-${timePart}`;",
			"}",
		},
		want: []string{"iso.slice(0, 10)"},
		fixed: []string{
			"export function migrationStamp(): string {",
			"    const now = new Date();",
			"    const iso = now.toISOString();",
			"    const datePart = dateIso8601(now, 'UTC');",
			"    const timePart = iso.slice(11, 19).replace(/:/g, '-');",
			"    return `${datePart}-${timePart}`;",
			"}",
		},
	},
}

func TestConsistencyNoIsoStringDateCutFiresOnRealSites(t *testing.T) {
	t.Parallel()

	for _, site := range consistencyNoIsoStringDateCutRealSites {
		t.Run(site.name, func(t *testing.T) {
			t.Parallel()

			consistencyNoIsoStringDateCutExpect(t, consistencyNoIsoStringDateCutRun(t, consistencyNoIsoStringDateCutSource(site.broken...)), site.want)
		})
	}
}

func TestConsistencyNoIsoStringDateCutStaysSilentOnFixedRealSites(t *testing.T) {
	t.Parallel()

	for _, site := range consistencyNoIsoStringDateCutRealSites {
		t.Run(site.name, func(t *testing.T) {
			t.Parallel()

			rule_testing.ExpectClean(t, consistencyNoIsoStringDateCutRun(t, consistencyNoIsoStringDateCutSource(site.fixed...)))
		})
	}
}

// `dateIso8601` itself cuts the ISO string on its no-zone branch, and the rule reports it like any
// other site; given the zone, the definition is clean.
func TestConsistencyNoIsoStringDateCutReportsTheHelperDefinitionUntilItIsFixed(t *testing.T) {
	t.Parallel()

	run := func(fixed bool) rule_testing.Result {
		return rule_testing.RunTypedFiles(t, ConsistencyNoIsoStringDateCut, map[string]string{
			consistencyNoIsoStringDateCutFormatTimeFile: consistencyNoIsoStringDateCutFormatTime(fixed),
		}, consistencyNoIsoStringDateCutFormatTimeFile)
	}
	consistencyNoIsoStringDateCutExpect(t, run(false), []string{"date.toISOString().split('T')[0]"})
	rule_testing.ExpectClean(t, run(true))
}

func TestConsistencyNoIsoStringDateCutFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		lines []string
		want  []string
	}{
		{"substr, which ahra does not use yet", []string{
			"export const day = new Date().toISOString().substr(0, 10);",
		}, []string{"new Date().toISOString().substr(0, 10)"}},
		{"substr from inside the date part", []string{
			"export const monthAndDay = new Date().toISOString().substr(5, 5);",
		}, []string{"new Date().toISOString().substr(5, 5)"}},
		{"substring with its bounds swapped", []string{
			"export const day = new Date().toISOString().substring(10, 0);",
		}, []string{"new Date().toISOString().substring(10, 0)"}},
		{"the month alone", []string{
			"export const month = new Date().toISOString().slice(0, 7);",
		}, []string{"new Date().toISOString().slice(0, 7)"}},
		{"through parentheses", []string{
			"export const day = (new Date().toISOString()).slice(0, 10);",
		}, []string{"(new Date().toISOString()).slice(0, 10)"}},
		{"split on a template T", []string{
			"export const day = new Date().toISOString().split(`T`)[0];",
		}, []string{"new Date().toISOString().split(`T`)[0]"}},
		{"two cuts on one line (FinanceStatementRangePresets)", []string{
			"export function dayKey(startTime: Date, endTime: Date): string {",
			"    return `${startTime.toISOString().slice(0, 10)}..${endTime.toISOString().slice(0, 10)}`;",
			"}",
		}, []string{"startTime.toISOString().slice(0, 10)", "endTime.toISOString().slice(0, 10)"}},
		{"a Date from a unix timestamp (StripePayoutSplitter)", []string{
			"export function unixSecondsToIsoDate(unixSeconds: number): string {",
			"    return new Date(unixSeconds * 1000).toISOString().slice(0, 10);",
			"}",
		}, []string{"new Date(unixSeconds * 1000).toISOString().slice(0, 10)"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			consistencyNoIsoStringDateCutExpect(t, consistencyNoIsoStringDateCutRun(t, consistencyNoIsoStringDateCutSource(testCase.lines...)), testCase.want)
		})
	}
}

func TestConsistencyNoIsoStringDateCutStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		lines []string
	}{
		// Cuts that are not the date part, each a real ahra shape.
		{"the time part only (FinanceAdapter)", []string{
			"export function isExactUtcMidnight(parsed: Date): boolean {",
			"    const isoUtc = parsed.toISOString();",
			"    return isoUtc.slice(11, 23) === '00:00:00.000';",
			"}",
		}},
		{"date and time to the minute (AhraOsWakeContinuity)", []string{
			"export function stamp(epochInMilliseconds: number): string {",
			"    return new Date(epochInMilliseconds).toISOString().slice(0, 16).replace('T', ' ');",
			"}",
		}},
		{"a cut after a replace (TasksBriefing)", []string{
			"export const time = new Date().toISOString().replace('T', ' ').slice(0, 19);",
		}},
		{"the whole ISO string", []string{
			"export const stampedAt = new Date().toISOString();",
		}},
		{"the time half of the split", []string{
			"export const time = new Date().toISOString().split('T')[1];",
		}},
		{"a split on anything but T", []string{
			"export const day = new Date().toISOString().split('.')[0];",
		}},
		{"a split with a limit", []string{
			"export const day = new Date().toISOString().split('T', 1)[0];",
		}},
		{"a rest element", []string{
			"export const [...halves] = new Date().toISOString().split('T');",
		}},
		{"the second element destructured, the first skipped", []string{
			"export const [, time] = new Date().toISOString().split('T');",
		}},
		{"a cut from the end", []string{
			"export const day = new Date().toISOString().slice(-24, -14);",
		}},
		{"a one-argument slice, the whole tail", []string{
			"export const tail = new Date().toISOString().slice(0);",
		}},
		{"an empty range", []string{
			"export const nothing = new Date().toISOString().slice(10, 10);",
		}},
		{"a computed bound", []string{
			"export const day = new Date().toISOString().slice(0, days);",
		}},

		// Strings that are not the default library's ISO string.
		{"a date string that is not from toISOString", []string{
			"export const month = isoDate.slice(0, 7);",
		}},
		{"a let holding the ISO string", []string{
			"export function stamp(): string {",
			"    let iso = new Date().toISOString();",
			"    use(iso);",
			"    return iso.slice(0, 10);",
			"}",
		}},
		{"another Date method's string (toString is local time)", []string{
			"export const day = new Date().toString().slice(0, 10);",
		}},
		{"a local class named Date", []string{
			"class Date {",
			"    toISOString(): string { return '2026-10-03T00:00:00.000Z'; }",
			"}",
			"export const day = new Date().toISOString().slice(0, 10);",
		}},
		{"a date library object with its own toISOString", []string{
			"interface MomentInterface { toISOString(): string }",
			"declare function moment(): MomentInterface;",
			"export const day = moment().toISOString().slice(0, 10);",
		}},

		// The fix.
		{"dateIso8601 with the UTC zone", []string{
			"export const day = dateIso8601(new Date(), 'UTC');",
		}},
		{"dateIso8601 with the local zone", []string{
			"export const day = dateIso8601(new Date(), userTimeZone());",
		}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			rule_testing.ExpectClean(t, consistencyNoIsoStringDateCutRun(t, consistencyNoIsoStringDateCutSource(testCase.lines...)))
		})
	}
}

// Two global scripts share one scope, so a `const` ISO string can be declared in one file and cut in
// another. The other file is never followed; the same cut with the `const` declared here reports, so
// the silence is the file boundary and not a scope that failed to resolve.
func TestConsistencyNoIsoStringDateCutFollowsOnlyThisFile(t *testing.T) {
	t.Parallel()

	// The harness forces every file to be a module; these two must be scripts to share a scope.
	scripts := func(directory string) {
		config := `{"compilerOptions": {"strict": true, "target": "ES2022", "lib": ["ES2022"], "types": []}, "include": ["**/*.ts"]}`
		if err := os.WriteFile(filepath.Join(directory, "tsconfig.json"), []byte(config), 0o644); err != nil {
			t.Fatalf("writing the tsconfig: %v", err)
		}
	}
	files := map[string]string{
		"/repository/source/Declaring.ts": "const startedIso = new Date().toISOString();",
		"/repository/source/Reading.ts":   "const startedOn = startedIso.slice(0, 10);",
	}
	rule_testing.ExpectClean(t, rule_testing.RunTypedFilesWithSetup(t, ConsistencyNoIsoStringDateCut, files, "/repository/source/Reading.ts", scripts))

	files["/repository/source/Reading.ts"] = "const startedIsoHere = new Date().toISOString();\nconst startedOnHere = startedIsoHere.slice(0, 10);"
	consistencyNoIsoStringDateCutExpect(t, rule_testing.RunTypedFilesWithSetup(t, ConsistencyNoIsoStringDateCut, files, "/repository/source/Reading.ts", scripts),
		[]string{"startedIsoHere.slice(0, 10)"})
}
