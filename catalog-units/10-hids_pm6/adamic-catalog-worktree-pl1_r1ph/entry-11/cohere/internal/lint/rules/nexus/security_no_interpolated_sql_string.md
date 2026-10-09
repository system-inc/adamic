# `nexus/security-no-interpolated-sql-string`

| | |
|---|---|
| **Recommendation** | **Yes**, at `error` once ahra's 33 findings are fixed. Register it `off` until then: six are real bugs, and the other 27 need a narrower type or a bound parameter each, a per-site call |
| Findings | **ahra 33, www-phi-health 0, www-connected-app 0**; also **api-phi-health 14**, **api-connected-app 0** (measured 2026-10-02) |
| Measured precision | 47 of 47 true: every finding is a `string`-typed value between the single quotes of a SQL string. 6 real bugs, 41 true-but-harmless (the value cannot carry a quote today, but its type says it could), 0 false |
| Research count | 33 sites in 10 files (`#tevhg3f`, `research_js_catalogs.md` entry 3); cohere finds the same 33 sites in the same 10 files |
| Plugin | cohere-native, `nexus` |
| Auto-fixable | no |
| Needs type information | yes, to tell a `string` from a number or a literal union; without a checker the rule registers nothing |

## What it checks

A template literal that reads as SQL, with an interpolation between the single quotes of one of its
strings, where the interpolated value's type can hold a quote. The rule's doc comment carries the
precise version: the two shapes that make a template SQL (a statement opening with an uppercase
keyword, or a fragment's `LIKE '...'`), the quote-state reading that gives up rather than guess, the
types that count as quote-capable, and the one escape it accepts.

## Where it came from

`modules/apple/contacts/ContactsApi.ts:62` in ahra, the contact search: `LIKE '%${word}%'` with each
search word, so a name with an apostrophe returns nothing. And
`modules/os/sensation/AhraOsTriggers.ts:1589`, which doubles quotes but not backslashes in a query that
runs on MySQL. Found by the new-rules sweep (`#tevhg3f`), built as `#3nvwzjs`.

## Reconciling with the research count

The research probe and the rule agree site for site: 33 in the same 10 files, the line numbers shifted
by edits since the probe ran. The first build found 32: it asked every statement for a second uppercase
keyword, and `AdsRegistry.ts:186` is `WHERE campaign.status = '${...}'` with only one. An uppercase
`WHERE` or `HAVING` now confirms itself, and the 33rd site came back.

The two definitions are not the same, and they land on the same set only because ahra's SQL is written
in the shapes both accept. The probe's SQL test was any of seven keywords anywhere in the template, and
its quote test looked at one text on each side of the interpolation; the rule reads the whole template
as SQL from its first character. On a template like `'a' = ${x} AND b = 'c'` the probe would see a
quote before and after `${x}` and report it, and the rule would not, because `${x}` is between two
strings, not inside one. Where they differ, the rule is the exact one.

## The findings, read one by one

**Real bugs (6)**

| Site | Shape | Reading |
|---|---|---|
| `modules/apple/contacts/ContactsApi.ts:62` (×3) | `LIKE '%${word}%'`, one per name column | **Bug.** `word` is the user's search text; `O'Brien` ends the string, the query fails inside a `try`, and the search finds nothing |
| `modules/os/sensation/AhraOsTriggers.ts:1589` (×2) | `viewIdentifier = '${viewIdentifier}' OR ... LIKE '${viewIdentifier}?%'`, quotes doubled | **Bug.** The query runs on MySQL, where a trailing `\` escapes the closing quote. The identifier comes from a trigger's parameters |
| `modules/phi/commerce/PhiCustomerApi.ts:282` | `emailAddress = '${escapedEmail}'`, quotes doubled | **Bug**, the same one: a PlanetScale (MySQL) query with only quotes doubled, on an email the operator types |

**True but harmless today (27 in ahra)**: the value is a `string` that cannot hold a quote as the code
stands, and narrowing its type (or binding it) silences the finding.

| Site | Value | Why it cannot hold a quote |
|---|---|---|
| `modules/apple/contacts/ContactsApi.ts:227` | `digits` | `phoneNumber.replace(/\D/g, '')` |
| `modules/apple/imessage/iMessageApi.ts:749`, `:832`, `:1189` | the phone number, inside a Python script's SQL | `replace(/[^0-9+]/g, '')` |
| `modules/phi/commerce/PhiCustomerApi.ts:298`, `:365` | `accountId` | read back from the `Account` row the previous query found |
| `modules/data/DataTrimmedRowsDelete.ts:294` (×2), `:335` | `cutoff`, `ceiling` | timestamps the module computes |
| `modules/data/DataTrimmedRowsDelete.ts:569`, `:596` | the table name, quotes doubled | MySQL, so the escape is incomplete, but the value is the operator's `--table` |
| `modules/finance/connections/QuickBooksAdapter.ts:245`, `:277`, `:452` | `sinceDate` | a date string; QuickBooks queries take no parameters, so the fix is a date-shaped type |
| `modules/finance/connections/QuickBooksAdapter.ts:330` | `accountType` | every caller passes a constant; the fix is a literal union |
| `modules/google/ads/GoogleAdsKeywordApi.ts:127`, `GoogleAdsReportingApi.ts:52`, `:113`, `:159`, `:208` | `date` | a date string; GAQL takes no parameters |
| `modules/ads/AdsRegistry.ts:186` | `mapUniversalToGoogleStatus(...)` | returns one of three constants, declared `string` |
| `modules/ads/AdsRegistry.ts:255` (×2) | `start`, `end` | `resolveDateRange` dates |
| `modules/planetscale/reports/i18n-conversion.ts:355` (×2), `:367` (×2) | `window.start`, `window.end` | dates from `addDays` |

**api-phi-health (14), true but harmless**:
`workers/chat/internal/service/metrics/ChatMetricsRpcService.ts:70`, `:76`, `:84`, `:107`, `:115`,
`:125`, `:218`, two each, `createdAt >= '${start}' AND createdAt <= '${end}'`, where both are
`new Date(...).toISOString()`. The `${ChatMessageStatusKind.Active}`-style enum values beside them in
the same queries are literal types and stay silent, which is the type gate working on a real tree.

**www-phi-health, www-connected-app, api-connected-app: 0.** The control: neither front end has a
line with an uppercase `SELECT`, `WHERE`, `LIKE`, `INSERT`, `UPDATE` or `DELETE` and a `'${` (the same
grep finds 28 such lines in ahra), and api-connected-app has 22 files with SQL keywords but no quoted
interpolation in them.

## Declined, by decision

- **Lowercase SQL.** Every lowercase keyword before `'${` in ahra is English (`delete refused:
  '${name}' has ...`, `Could not update floor designation for '${accountName}'`).
- **Escaping functions by name.** `escapeSqlString(value)` is reported; only the visible
  backslash-and-quote `replace` chain is accepted. The doc comment gives the reasoning.
- **A fixer.** The fix is a bound parameter, which spans the query and the driver call.

## The bare-value branch (`interpolatedSqlValue`, #3ng4vkk)

A second shape, reported under its own id: in a confirmed statement, an unquoted interpolation right
after a comparison (`=`, `<>`, `!=`, `<`, `>`, `<=`, `>=`) whose type can hold any text. The value is
written straight into the SQL, so quoting or escaping can't make it safe; the fix is a bound parameter,
or, for a language with none such as GAQL, a type that holds only fixed text. A union of literals stays
silent, which is how GAQL writes an enum. Only comparisons count: after `(`, `,`, `AND`, `FROM` or
`ORDER BY`, an interpolation builds SQL from code, such as an `IN` list of placeholders, a condition, or
a table or column name.

Measured on 2026-10-03 before the fixes: **ahra 25, www 0, api 0**, all 25 true.
- 23 were GAQL ids typed `string` from the command line, across 7 files in `modules/google/ads`. They
  were fixed by typing the ids as numbers at the entry points (`#kx29bm3`, ahra 0269aa0).
- 2 were `sqlLiteral(...)` in `modules/data/DataMirrorReconcile.ts`, a helper that doubled only the
  quote, in a query that runs on MySQL. That is the `AhraOsTriggers` bug class, which the quoted branch
  can't see because the quotes are added inside the helper. It was fixed in ahra 7d88414.

After both fixes, ahra reads 0 with a planted string id firing.

Verification: 5 firing and 8 silent fixture groups, plus 8 mutants, each confirmed to compile, all
killed. The mutants: any value token counted rather than only comparisons, the branch removed,
literals counted as text, template holes ignored, branded strings ignored, an open type parameter
treated as fixed, `any` and `unknown` treated as fixed, and the bare value reported under the quoted
id.

## Verification

- Fixtures both ways from the real sites: `ContactsApi.ts` before (4 findings: the three search
  words and the phone digits) and after (bound parameters, silent); `AhraOsTriggers.ts` with quotes
  doubled (2 findings) and with backslashes doubled as well (silent). Then 25 firing shapes and 36
  silent ones, and the no-checker decline.
- Mutation check, 31 mutants, each through `go test -overlay` against a mutated copy, each killed:
  the type gate; the statement gate; the second keyword; `WHERE`/`HAVING` confirming themselves; the
  value position; the string closing in the template; giving up on a backslash, on `#`, on a dollar
  quote and on `--` without a space; a doubled quote read as one; double quotes, backquotes, line and
  block comments read as such; uppercase-only matching (leading word and keyword count, separately);
  the tagged-template skip; the fragment's second keyword, even quote count, backslash and closing
  quote; string literals holding a quote; template literal holes; branded strings; type parameter
  constraints; `any` and `unknown`; and the escape (recognized at all, both characters required,
  `const` only, global regular expression only).
