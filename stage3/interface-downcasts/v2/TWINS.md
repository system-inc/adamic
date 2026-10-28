# Item 8 stage 2, wave 1: v2 twins

Base: `origin/main` at `4ac3f04d767ec7d2d4e9b41ae9cdc9c4c9f4f288`. Precedent: `bf3ff0489b97ca92e4d05863d2d4833ad2399986`.

The directory matched the ten requested fixtures, plus `array-arm.a` and `fs-option-boxing.a`, which remain unchanged. All ten fixtures were moved with `git mv`; their source bytes are unchanged. Each count record was moved with its fixture, changing only the path in its row; all six numbers are unchanged.

## Comparison

Each outcome below records exit code, stdout and stderr. `\n` is an LF byte. N and J identify the native and JavaScript runs, respectively; identical outcome identifiers mean byte-identical stdout/stderr and equal exit codes. Before is `.a`; after is `.ts`.

| File (now .ts) | Native before | Native after | JavaScript before | JavaScript after | Identical | Type predicate | Non-null assertion | Definite assignment |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| callable-producer-wrong.ts | N1 | N1 | J1 | J1 | yes | no | no | no |
| fixed-tuple-good.ts | N2 | N2 | J2 | J2 | yes | no | no | no |
| fixed-tuple-wrong.ts | N3 | N3 | J3 | J3 | yes | no | no | no |
| mixed-read-write.ts | N4 | N4 | J4 | J4 | yes | no | no | no |
| null-read-write.ts | N5 | N5 | J5 | J5 | yes | no | no | no |
| representation-read-write.ts | N6 | N6 | J6 | J6 | yes | no | no | no |
| snapshot-maybe-boolean.ts | N7 | N7 | J7 | J7 | yes | no | no | no |
| snapshot-maybe-number.ts | N8 | N8 | J8 | J8 | yes | no | no | no |
| tuple-identity-wrong.ts | N9 | N9 | J9 | J9 | yes | no | no | no |
| undefined-read-write.ts | N10 | N10 | J10 | J10 | yes | no | no | no |

No file failed to be a twin. Every native result also equals its JavaScript result.

| Outcomes | Exit | Stdout (JSON string) | Stderr (JSON string) |
| --- | --- | --- | --- |
| N1, J1 | 0 | `"true\n"` | `""` |
| N2, J2 | 0 | `"first:1\nfirst:1\n"` | `""` |
| N3, J3 | 70 | `""` | `"adamic: panic: field read failed: viewed.value matches no member of readonly [string, number]; expected readonly [string, number], found array\n"` |
| N4, J4 | 0 | `"first\nsecond\n"` | `""` |
| N5, J5 | 0 | `"true\n"` | `""` |
| N6, J6 | 0 | `"ok\ntrue:true\n41:-42:3.5\n"` | `""` |
| N7, J7 | 0 | `"true / true\nfalse / false\nfalse / undefined\n"` | `""` |
| N8, J8 | 0 | `"7\n0\nundefined\n"` | `""` |
| N9, J9 | 70 | `""` | `"adamic: panic: field read failed: viewed.value matches no member of readonly [string, number]; expected readonly [string, number], found object\n"` |
| N10, J10 | 0 | `"true\n"` | `""` |

## Method and readers

Before any fixture rename, a temporary capture program called `load.Load`, `lower.Lower`, `native.C` / `native.Build` with `native.Options{}`, and `javascript.JavaScript`. It executed native binaries directly and JavaScript with `node --disable-warning=ExperimentalWarning oracle/node.mjs <generated program.mjs>`, matching `releasedUncached` and `onJavaScriptBackend`. The same binary ran all ten again after the rename. Separate JSON records saved stdout, stderr and exit code for each file/backend; all 20 records compared equal. Logs: `/tmp/item8-before.log`, `/tmp/item8-after.log`; records: `/tmp/item8-before.json`, `/tmp/item8-after.json`. The temporary capture source is removed before commit.

`git grep -n 'interface-downcasts/v2'` across the whole repository, supplemented with searches for the ten basenames and indirect `interfaceFixture` readers, found the eight explicit registrations in `checked_views_v2_counts_test.go`, the generated snapshot path in `view_snapshot_test.go`, the migration tests using the shared helper, and ten count records. There are no matching a-check headers, `internal/acheck/known.txt` entries or verification JSON entries. Historical `review/compiler` evidence retains the paths actually run, following the precedent commit. No stage3 Go package reads these ten fixtures; their readers all live in `internal/oracle`.

The sole authorized logic change replaces `interfaceFixturePath`'s name list with `os.Stat` of both candidate paths relative to `repository`. Exactly one must exist; otherwise it panics naming both. Every other Go edit changes only `.a` to `.ts` in a path string. No tests were added, and no expected outputs changed.

## Mutant

Each mutation was applied alone and restored before the full suite. No mutant changes enter the commit.

### Counts filename

Only `fixed-tuple-good.ts.counts` was moved back to `.a.counts`; the fixture remained `.ts` and the row still named `.ts`.

Command: `go test ./internal/oracle -run '^TestCountsFilesCensus$' -count=1 -v` (exit 1; 0 passed test events, 1 failed test events). Log: `/tmp/item8-mutant-counts.log`.

```text
count_files_test.go:158: fixture stage3/interface-downcasts/v2/fixed-tuple-good.ts needs counts file testdata/counts/runtime/stage3/interface-downcasts/v2/fixed-tuple-good.ts.counts: open testdata/counts/runtime/stage3/interface-downcasts/v2/fixed-tuple-good.ts.counts: no such file or directory
```

### Always .a

Replaced the helper body temporarily with `return "stage3/interface-downcasts/" + name + ".a"`.

Command: `go test ./internal/oracle -run '^TestCheckedViewV2ReadsAfterWrites$' -count=1 -v` (exit 1; 1 passed test events, 6 failed test events). Log: `/tmp/item8-mutant-always-a.log`.

```text
checked_views_v2_migration_test.go:20: Load: load: no file at /workspace/adamic/stage3/interface-downcasts/v2/mixed-read-write.a
```

Witness: `TestCheckedViewV2ReadsAfterWrites/mixed-read-write`. Five renamed subtests fail; unchanged `fs-option-boxing` passes.

### Both extensions

Copied `fixed-tuple-good.ts` to a stray `fixed-tuple-good.a` beside it.

Command: `go test ./internal/oracle -run '^TestCheckedViewV2ReadsAfterWrites$/^fixed-tuple-good$' -count=1 -v` (exit 1; 0 passed test events, 2 failed test events). Log: `/tmp/item8-mutant-both.log`.

```text
panic: expected exactly one interface fixture: stage3/interface-downcasts/v2/fixed-tuple-good.a or stage3/interface-downcasts/v2/fixed-tuple-good.ts [recovered, repanicked]
```

Witness: `TestCheckedViewV2ReadsAfterWrites/fixed-tuple-good`. The panic names both candidate paths.

## Validation

The first full-suite attempt used `go test ./internal/oracle/... -count=1 -json -timeout=60m` without `ADAMIC_TYPESCRIPT_SOURCE`. It was interrupted (signal TERM, exit 143) after 1075 passed, 506 skipped and 2 failed test events; 547 top-level tests had passed. Failures: `TestCountsAreRecorded_async_counts_stable`, `TestCountsAreRecorded_026`. The shared diagnostic was `ADAMIC_TYPESCRIPT_SOURCE must name the pinned upstream TypeScript checkout; Program sources never fetch`. Logs: `/tmp/item8-oracle-unconfigured.jsonl` and `/tmp/item8-oracle-unconfigured.stderr`.

Prepared `/tmp/item8-typescript` from the repository and tag in `stage3/source.json`, and verified HEAD is exactly `050880ce59e30b356b686bd3144efe24f875ebc8`. This changes no repository files or expected outputs.

The second attempt used `ADAMIC_TYPESCRIPT_SOURCE=/tmp/item8-typescript go test ./internal/oracle/... -count=1 -json -timeout=60m`. It was interrupted (signal TERM, exit 143) after 2271 passed, 533 skipped and 12 failed test events; 1243 top-level tests had passed. The 12 assignment-proof tests failed because `stage3/api/node_modules` had not been installed: `Error: Cannot find module 'typescript'`. Logs: `/tmp/item8-oracle-no-api.jsonl` and `/tmp/item8-oracle-no-api.stderr`.

Ran `npm ci --prefix stage3/api` to install its locked TypeScript 6.0.3 and Node type dependencies (log `/tmp/item8-api-setup.log`); no tracked dependency files changed.

Full package final run: `ADAMIC_TYPESCRIPT_SOURCE=/tmp/item8-typescript go test ./internal/oracle/... -count=1 -json -timeout=60m > /tmp/item8-oracle.jsonl 2> /tmp/item8-oracle.stderr`. PASS, exit 0: 1 package; 1773 top-level tests and 6126 subtests passed (7899 passing test events). 442 top-level tests and 99 subtests skipped (541 skip events). Zero failures. Package elapsed: 658.887 seconds.

All ten `.ts` records passed their `TestCountsAreRecorded_*` checks. Both snapshot tests and the migration tests passed. No additional stage3 package reads these fixtures.

Final scope audit verified unchanged bodies for all ten renamed fixtures, unchanged numbers in all ten count rows, untouched excluded fixtures, path-only edits in both reader files, and no change outside the authorized function in `interface_cast_test.go`. `git diff --check` and the formatting check passed. The final commit contains no transient capture sources, mutant edits, dependency installs, or test logs.
