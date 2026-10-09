# `nexus/consistency-no-iso-string-date-cut`

| | |
|---|---|
| **Recommendation** | **Yes, registered `off` until the sites are fixed.** Turning it on at `error` today puts 80 findings on ahra |
| Findings | **ahra 80** (measured 2026-10-03) |
| Measured precision | 80 of 80: every finding is a `toISOString()` string cut to its date part, which is the violation by definition |
| Plugin | cohere-native, `nexus` |
| Auto-fixable | no |
| Needs type information | yes, to resolve `toISOString` to the default library's `Date` method and a `const` to its one declaration; without a checker the rule registers nothing |

## Name

`consistency-` because the judgment is one spelling for one act: a calendar day is written with its
zone, through `dateIso8601(date, timeZone)`. `no-iso-string-date-cut` because what is reported is the
cut itself. The brief's working title was "`toISOString().slice(0, 10)` used as today".

## What it checks

A string from the default library's `Date.prototype.toISOString`, called directly or held in a `const`
declared in the same file, cut to its date part by any of:

- `.slice(a, b)`, `.substring(a, b)` (either order) or `.substr(a, length)` with non-negative integer
  literals whose range is non-empty and ends at or before index 10;
- `.split('T')[0]`, the split taking exactly that one argument;
- `const [day] = <iso>.split('T')`, the first element bound and not a rest or a hole.

Not reported: time-only and date-and-time cuts (`slice(11, 23)`, `slice(0, 16)`), cuts after anything
else touched the string (`.replace('T', ' ').slice(0, 19)`), the time half of the split, a split with a
limit, negative or computed bounds, a `let`, another file's `const`, any other `Date` method's string
(`toString` is local time), and a `toISOString` that is not the default library's (a local class,
moment). The doc comment carries the precise rules.

The fix it names: `dateIso8601(date, 'UTC')` from `@nexus/source/time/FormatTime` when the UTC day is
meant, `dateIso8601(date, userTimeZone())` with `userTimeZone` from `@nexus/source/time/TimeZones` when
the local day is meant.

## Why a ban on the shape

A bare cut is the UTC day whether or not the author meant one, and in Utah the UTC day turns over at 5
pm in winter and 6 pm in summer. The code cannot tell an intended UTC day from a mistaken one, so a rule
that tries to decide intent is either silent on the bugs or wrong on the correct sites. The ban sidesteps
that: the cut is the violation everywhere, and the fix says the zone on the line. At a site that already
meant UTC the fix is a rename to `dateIso8601(date, 'UTC')`.

## The real bugs behind it

The rule was first built as a correctness rule, reporting only where the code itself proves local
intent. That version found these, and they remain the motivation:

| Site | Bug |
|---|---|
| `app/api/finance/route.ts:48` (`recentWindowSinceIso`) | `setDate(getDate() - 30)` in local days, then the UTC slice: the dashboard's 30-day feed starts a day late in the evening |
| `modules/finance/FinanceApi.ts:474` (`computeMonthlyBurn`) | same; the burn window loses its first local day |
| `modules/finance/FinanceAnalysisWindow.ts:45` | same, on a copy of the caller's `today`; the rolling founder window |
| `modules/finance/FinanceTransactionsCommandLineInterface.ts:270` | same; `--window N` cashflow |
| `modules/finance/FinanceCommandLineInterfaceShared.ts:56` (`today()`) | "the default as-of for every dated write verb": its six callers (`transactions add --date`, four `position` verbs, `holdings`, each `--as-of`) stamp an evening write with tomorrow's date. The intent is in its doc comment, not its code, which is why only a shape ban can reach it |

The site-by-site pass for this rewrite (`sites.tsv` in the build scratch directory) found five more where
the code pairs the UTC day with something local: `AhraOsRehydrate.ts:143` and `AhraOsInbox.ts:109` print
the UTC date beside `toTimeString()`'s local time, so an evening stamp shows tomorrow's date with
tonight's time; `FormatTime.ts:665` (`dateTimeIso8601`) does the same with `formatTime`, and its own doc
example says "UTC date, local time"; `FinanceStatementRangePresets.tsx:91` cuts date-fns
`startOfToday`/`endOfToday` ranges, harmless there because both sides of the match are cut the same way.

## Reconciling with the research count

Research counted **77** (`J_iso_utc_day`), 21 on `new Date()`. cohere counts **80**: 54 `slice(0, 10)`,
10 `substring(0, 10)`, 11 `split('T')[0]`, 4 destructured `split('T')`, and 1 `const`-held ISO string
(`modules/os/migrations/NewMigration.ts:31`). Every finding matches a text search line for line, the
now-based count is 21 on both sides, and no ahra commit since the sweep added or removed a cut. The
research probe's recorded forms cover 75, and the destructured and `const`-held forms make 80. Its
exact definition is not recorded, so why it landed on 77 rather than 75 or 80 cannot be shown.

## The 80, classified

`sites.tsv` lists each as `file:line`, a class and a note:

- **Local, 10**: the 4 finance windows and `today()` above, plus the 5 local pairings.
- **UTC, 22**: provably UTC (`T00:00:00Z` or UTC-noon anchors with `setUTCDate`, `getUTCFullYear` on the
  same Date, a UTC time from the same string) or documented UTC at the site. Fix: `dateIso8601(date,
  'UTC')`. Three of these, the finance `isoDay` view helpers, are documented UTC but fed `new Date()` and
  date-picker ranges, so their end bound is worth a look in the fix pass.
- **Unstated, 48**: the zone is never stated; 20 of them are `new Date()` as an as-of or today, the rest
  API timestamps and listing dates.

`dateIso8601`'s own no-zone branch (`FormatTime.ts:65`) is one of the 22, reported like any site with no
exemption: the fix gives that branch the zone through the same `Intl.DateTimeFormat` the zoned branch uses.

## Verification

- Fixtures from the real sites, one per form, before and after: `today()` (`slice`), the finance feed
  window (`slice`, local arithmetic), `OpenAiAdsInsightsApi` `isoDateUtc` (`substring`), the Discord
  history line (`split('T')[0]`), the Pensieve week start (destructured split), `NewMigration` (`const`-held
  string; the fix keeps its time-only slice). `dateIso8601`'s definition reports unfixed and is clean
  fixed. `substr` has no ahra site and is fixtured from the shape.
- Silent: time-only, date-and-time, after-`replace`, the whole string, the time half, a split on another
  separator, a split with a limit, a rest, a hole, negative, one-argument, empty and computed bounds, a
  non-ISO string, a `let`, `toString`, a local `Date` class, a moment object, both `dateIso8601` calls,
  and another file's `const` (a two-script fixture whose same-file twin reports).
- Mutation check, each mutant a copy through `go test -overlay -count=1`, 15 run, 15 killed: no date-part
  end bound (time-only slice); empty range allowed; substring bounds not swapped; substr read as slice;
  no default-library member check (local `Date` class, moment); `const` not followed (migration stamp);
  `let` followed; other files followed (two-script fixture); split limit allowed; any separator; any
  index (time half); rest counted; hole counted; no destructured split (Pensieve); any `Date` method's
  string, not only `toISOString` (`toString`).
