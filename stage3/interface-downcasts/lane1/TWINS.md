# Item 8 stage 2: lane1 twins

Base: `d579ff21e10791f78ab269628849513f7db5a39c`. Precedent: `d579ff21e10791f78ab269628849513f7db5a39c` (read in full, including its TWINS.md), and `bf3ff0489b97ca92e4d05863d2d4833ad2399986`.

All 11 requested fixtures were moved with `git mv`, with unchanged source bytes. Each existing count record was moved with its fixture, changing only the path in its row. All six numbers in every moved count record compare equal to the base.

## Comparison

Each outcome records exit code, stdout and stderr. `\n` is an LF byte. N and J identify native and JavaScript runs. Repeated identifiers mean byte-identical stdout/stderr and equal exit codes. Before is `.a`; after is `.ts`.

| File (now .ts) | Native before | Native after | JavaScript before | JavaScript after | Identical | Type predicate | Non-null assertion | Definite assignment |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| flag-downcast-wrong.ts | N1 | N1 | J1 | J1 | yes | no | no | no |
| interfaces-good.ts | N2 | N2 | J2 | J2 | yes | no | no | no |
| interfaces-missing-inherited.ts | N3 | N3 | J3 | J3 | yes | no | no | no |
| interfaces-uninitialized-object.ts | N4 | N4 | J4 | J4 | yes | no | no | yes |
| interfaces-wrong-inherited.ts | N5 | N5 | J5 | J5 | yes | no | no | no |
| objects-good.ts | N6 | N6 | J6 | J6 | yes | no | no | no |
| objects-missing-nested.ts | N7 | N7 | J7 | J7 | yes | no | no | no |
| objects-uninitialized-nested.ts | N8 | N8 | J8 | J8 | yes | no | no | yes |
| objects-untagged-good.ts | N9 | N9 | J9 | J9 | yes | no | no | no |
| objects-untagged-wrong.ts | N10 | N10 | J10 | J10 | yes | no | no | no |
| objects-wrong-nested.ts | N11 | N11 | J11 | J11 | yes | no | no | no |

No file failed to be a twin. Every native result also equals its JavaScript result.

| Outcomes | Exit | Stdout (JSON string) | Stderr (JSON string) |
| --- | --- | --- | --- |
| N1, J1 | 70 | `""` | `"adamic: panic: field read failed: (symbol as TransientSymbol).links.checkFlags is not a number; expected number, found boolean\n"` |
| N2, J2 | 0 | `"true:okok\ntrue\n"` | `""` |
| N3, J3 | 70 | `""` | `"adamic: panic: field read failed: view.child.ready is not initialized; expected boolean, found missing\n"` |
| N4, J4 | 70 | `""` | `"adamic: panic: field read failed: view.child.next is not initialized; expected Child, found uninitialized\n"` |
| N5, J5 | 70 | `""` | `"adamic: panic: field read failed: view.child.ready is not a boolean; expected boolean, found number\n"` |
| N6, J6 | 0 | `"true:okok\ntrue\n"` | `""` |
| N7, J7 | 70 | `""` | `"adamic: panic: field read failed: view.child.ready is not initialized; expected boolean, found missing\n"` |
| N8, J8 | 70 | `""` | `"adamic: panic: field read failed: view.child.ready is not initialized; expected boolean, found uninitialized\n"` |
| N9, J9 | 0 | `"true:okok\ntrue:1\n"` | `""` |
| N10, J10 | 70 | `""` | `"adamic: panic: field read failed: view.child.ready is not a boolean; expected boolean, found number\n"` |
| N11, J11 | 70 | `""` | `"adamic: panic: field read failed: view.child.ready is not a boolean; expected boolean, found number\n"` |

## Method and readers

The required setup command was `export GOPROXY='https://proxy.golang.org|direct'; timeout --kill-after=60s 1800 bash cloud/setup.sh > /tmp/setup.log 2>&1`. Tool environment: `source /workspace/adamic-tools/env.sh`. The full suite also used the pinned TypeScript v6.0.3 checkout at `/tmp/item8-typescript` (verified HEAD `050880ce59e30b356b686bd3144efe24f875ebc8`) and `npm ci --prefix stage3/api`; no tracked dependency files changed.

Before any fixture rename, a temporary non-test capture program called `load.Load`, `lower.Lower`, `native.C` / `native.Build` with `native.Options{}`, and `javascript.JavaScript`. Native binaries ran directly; JavaScript ran with `node --disable-warning=ExperimentalWarning oracle/node.mjs <generated program.mjs>`, matching `releasedUncached` and `onJavaScriptBackend`. The same capture binary ran again after the rename. Separate JSON records save stdout, stderr and exit for each file/backend. The complete before/after maps compare equal. No tests were added, and no expected outputs changed.

Capture records: `/tmp/item8-lane1-before.json`, `/tmp/item8-lane1-after.json`; logs: `/tmp/item8-lane1-before.log`, `/tmp/item8-lane1-after.log`. Repository-wide `git grep -n -F` searches for every basename are saved in `/tmp/item8-lane1-readers.txt`. Additional exact-directory and helper searches checked indirect/generated readers. Historical reports, logs, scouts, verification evidence and a-check header records retain the paths actually observed. No status.json or manifest outcome row moves, so no recorded-row mutant is applicable. No stage3/fixtures reader is touched.

The directory contains the eleven requested fixtures plus `optional-read.a` and `optional-read-control.a`; both excluded files remain unchanged. All eleven have runtime count records. Active readers are `checked_views_flag_downcast_test.go`, `checked_views_interfaces_test.go`, `checked_views_objects_test.go`, and `interface_cast_counts_test.go`. The first three already use `interfaceFixturePath`. The sole Go edit is the explicitly authorized substitution of `interfaceFixturePath("lane1/"+name)` for the generated `.a` path in the counts loop, preserving every name, including `optional-read-control` (which still resolves to `.a`). The two definite-assignment fixtures are twins despite using that special construct. No fixture uses a type predicate or non-null assertion.

## Mutants

Each mutation was applied alone and restored before full validation. No mutant enters the commit.

### Counts filename

Only `interfaces-good.ts.counts` was moved back to `.a.counts`; its row still named `.ts`, and the fixture stayed `.ts`.

Command: `go test ./internal/oracle -run '^TestCountsFilesCensus$' -count=1 -json > /tmp/item8-lane1-mutant-counts.jsonl 2> /tmp/item8-lane1-mutant-counts.stderr`. Exit 1; 0 passing and 1 failing test event (`TestCountsFilesCensus`).

```text
count_files_test.go:158: fixture stage3/interface-downcasts/lane1/interfaces-good.ts needs counts file testdata/counts/runtime/stage3/interface-downcasts/lane1/interfaces-good.ts.counts: open testdata/counts/runtime/stage3/interface-downcasts/lane1/interfaces-good.ts.counts: no such file or directory
```

### Hard-coded .a counts reader

Only the helper substitution on `interface_cast_counts_test.go:40` was undone, restoring `"stage3/interface-downcasts/lane1/"+name+".a"`. The renamed fixtures and count records stayed `.ts`.

Command: `ADAMIC_TYPESCRIPT_SOURCE=/tmp/item8-typescript go test ./internal/oracle -run '^TestCountsAreRecorded_017$/stage3/interface-downcasts/lane1/interfaces-good.a$' -count=1 -json -timeout=20m > /tmp/item8-lane1-mutant-helper.jsonl 2> /tmp/item8-lane1-mutant-helper.stderr`. Exit 1; 0 passing and 1 failing test event (`TestCountsAreRecorded_017`). Preparing its counts measurements fails on the renamed `interfaces-good` row before subtests start:

```text
interface_cast_counts_test.go:43: Load: load: no file at /workspace/adamic/stage3/interface-downcasts/lane1/interfaces-good.a
```

## Validation

Full package command: `ADAMIC_TYPESCRIPT_SOURCE=/tmp/item8-typescript go test ./internal/oracle/... -count=1 -json -timeout=60m > /tmp/item8-lane1-oracle.jsonl 2> /tmp/item8-lane1-oracle.stderr`.

PASS, exit 0: 1 package; 1777 top-level tests and 6153 subtests passed (7930 passing test events). 442 top-level tests and 99 subtests skipped (541 skip events). Zero failures. Package elapsed: 1161.916 seconds. All renamed count rows and their active fixture readers pass in this full run.

No `stage3/fixtures` file was touched, so its additional package command is not required. Final audit compares each renamed source byte-for-byte with the base, each moved count row with only its path substituted, and both complete outcome maps. `git diff --check` passes. The commit includes no temporary capture source, mutant edit, dependency installation or test log. The Go diff contains only the authorized helper substitution(s) and, where applicable, `.a` to `.ts` path changes. No tests were added and no expected output changed.

Setup passed with exit 0 in 387.446 seconds; its retained log is `/tmp/item8-lane1-setup.log`. All eleven `.ts` count-row pass events and the excluded `optional-read-control.a` count-row pass event were individually verified.
