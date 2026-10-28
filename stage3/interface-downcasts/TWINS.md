# Item 8 stage 2: loose twins

Base: `e33e6cb67613be7f7d7aa44d3a724aa8f82549f6`. Precedent: `d579ff21e10791f78ab269628849513f7db5a39c` (read in full, including its TWINS.md), and `bf3ff0489b97ca92e4d05863d2d4833ad2399986`.

All 8 requested fixtures were moved with `git mv`, with unchanged source bytes. Each existing count record was moved with its fixture, changing only the path in its row. All six numbers in every moved count record compare equal to the base.

## Comparison

Each outcome records exit code, stdout and stderr. `\n` is an LF byte. N and J identify native and JavaScript runs. Repeated identifiers mean byte-identical stdout/stderr and equal exit codes. Before is `.a`; after is `.ts`.

| File (now .ts) | Native before | Native after | JavaScript before | JavaScript after | Identical | Type predicate | Non-null assertion | Definite assignment |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| default-boxed-string.ts | N1 | N1 | J1 | J1 | yes | no | no | no |
| default-destructure.ts | N2 | N2 | J2 | J2 | yes | no | no | no |
| default-literal.ts | N3 | N3 | J3 | J3 | yes | no | no | no |
| default-wrong-boolean.ts | N4 | N4 | J4 | J4 | yes | no | no | no |
| default-wrong-type.ts | N5 | N5 | J5 | J5 | yes | no | no | no |
| missing-name.ts | N6 | N6 | J6 | J6 | yes | no | no | no |
| visitor.ts | N7 | N7 | J7 | J7 | yes | no | no | no |
| wrong-kind.ts | N8 | N8 | J8 | J8 | yes | no | no | no |

No file failed to be a twin. Every native result also equals its JavaScript result.

| Outcomes | Exit | Stdout (JSON string) | Stderr (JSON string) |
| --- | --- | --- | --- |
| N1, J1 | 0 | `"okok\n"` | `""` |
| N2, J2 | 0 | `"okok\n"` | `""` |
| N3, J3 | 70 | `""` | `"adamic: panic: field read failed: (node as Identifier).name expected \"wanted\", found string other\n"` |
| N4, J4 | 70 | `""` | `"adamic: panic: field read failed: (node as Identifier).ready is not a boolean; expected boolean, found number\n"` |
| N5, J5 | 70 | `""` | `"adamic: panic: field read failed: (node as Identifier).name is not a string; expected string, found number\n"` |
| N6, J6 | 70 | `""` | `"adamic: panic: field read failed: identifier(malformed).name is not initialized; expected string, found missing\n"` |
| N7, J7 | 0 | `"idid\n42\ntexttext\n5\ntrue\nonce2\ncalls 2\n"` | `""` |
| N8, J8 | 70 | `"casting\n"` | `"adamic: panic: cast failed: this Node is not a Identifier\n"` |

## Method and readers

The required setup command was `export GOPROXY='https://proxy.golang.org|direct'; timeout --kill-after=60s 1800 bash cloud/setup.sh > /tmp/setup.log 2>&1`. Tool environment: `source /workspace/adamic-tools/env.sh`. The full suite also used the pinned TypeScript v6.0.3 checkout at `/tmp/item8-typescript` (verified HEAD `050880ce59e30b356b686bd3144efe24f875ebc8`) and `npm ci --prefix stage3/api`; no tracked dependency files changed.

Before any fixture rename, a temporary non-test capture program called `load.Load`, `lower.Lower`, `native.C` / `native.Build` with `native.Options{}`, and `javascript.JavaScript`. Native binaries ran directly; JavaScript ran with `node --disable-warning=ExperimentalWarning oracle/node.mjs <generated program.mjs>`, matching `releasedUncached` and `onJavaScriptBackend`. The same capture binary ran again after the rename. Separate JSON records save stdout, stderr and exit for each file/backend. The complete before/after maps compare equal. No tests were added, and no expected outputs changed.

Capture records: `/tmp/item8-loose-before.json`, `/tmp/item8-loose-after.json`; logs: `/tmp/item8-loose-before.log`, `/tmp/item8-loose-after.log`. Repository-wide `git grep -n -F` searches for every basename are saved in `/tmp/item8-loose-readers.txt`. Additional exact-directory and helper searches checked indirect/generated readers. Historical reports, logs, scouts, verification evidence and a-check header records retain the paths actually observed. No status.json or manifest outcome row moves, so no recorded-row mutant is applicable. No stage3/fixtures reader is touched.

The eight requested files were present directly in the directory. Other sources, including `nodes.a` and the existing `.ts` readiness controls, remain unchanged. Seven have count records; `missing-name` has none. Active readers are `interface_cast_test.go`, `interface_cast_counts_test.go`, and `view_fields_test.go`. Existing helper callers require no edits. The two literal visitor/wrong-kind registrations now pass the same names through `interfaceFixturePath`; the explicit missing-name path and the wrong-kind suffix path change only `.a` to `.ts`. The dormant `interfaceSource` function has no callers and remains untouched. No requested fixture uses any of the three special constructs.

## Mutants

Each mutation was applied alone and restored before full validation. No mutant enters the commit.

### Counts filename

Only `default-boxed-string.ts.counts` was moved back to `.a.counts`; its row still named `.ts`, and the fixture stayed `.ts`.

Command: `go test ./internal/oracle -run '^TestCountsFilesCensus$' -count=1 -json > /tmp/item8-loose-mutant-counts.jsonl 2> /tmp/item8-loose-mutant-counts.stderr`. Exit 1; 0 passing and 1 failing test event (`TestCountsFilesCensus`).

```text
count_files_test.go:158: fixture stage3/interface-downcasts/default-boxed-string.ts needs counts file testdata/counts/runtime/stage3/interface-downcasts/default-boxed-string.ts.counts: open testdata/counts/runtime/stage3/interface-downcasts/default-boxed-string.ts.counts: no such file or directory
```

### Hard-coded .a counts readers

Only the helper substitution for `visitor` and `wrong-kind` was undone, restoring the original literal `.a` path list. The renamed fixtures and count records stayed `.ts`.

Command: `ADAMIC_TYPESCRIPT_SOURCE=/tmp/item8-typescript go test ./internal/oracle -run '^TestCountsAreRecorded_011$/stage3/interface-downcasts/visitor.a$' -count=1 -json -timeout=20m > /tmp/item8-loose-mutant-helper.jsonl 2> /tmp/item8-loose-mutant-helper.stderr`. Exit 1; 0 passing and 1 failing test event (`TestCountsAreRecorded_011`). Preparing its counts measurements fails on the renamed `visitor` row before subtests start:

```text
interface_cast_counts_test.go:25: default compiler for stage3/interface-downcasts/visitor.a: exit status 1
```

Setup passed with exit 0 in 10.908 seconds; its retained log is `/tmp/item8-loose-setup.log`.

## Validation

Full package command: `ADAMIC_TYPESCRIPT_SOURCE=/tmp/item8-typescript go test ./internal/oracle/... -count=1 -json -timeout=60m > /tmp/item8-loose-oracle.jsonl 2> /tmp/item8-loose-oracle.stderr`.

PASS, exit 0: 1 package; 1777 top-level tests and 6153 subtests passed (7930 passing test events). 442 top-level tests and 99 subtests skipped (541 skip events). Zero failures. Package elapsed: 904.87 seconds. All renamed count rows and their active fixture readers pass in this full run.

No `stage3/fixtures` file was touched, so its additional package command is not required. Final audit compares each renamed source byte-for-byte with the base, each moved count row with only its path substituted, and both complete outcome maps. `git diff --check` passes. The commit includes no temporary capture source, mutant edit, dependency installation or test log. The Go diff contains only the authorized helper substitution(s) and, where applicable, `.a` to `.ts` path changes. No tests were added and no expected output changed.

All seven renamed `.ts` count-row pass events were individually verified. `TestInterfaceCastOracle`, `TestInterfaceCastChecksMalformedRead`, and `TestDefaultTaggedSourceViews` all passed. The directory mutation scripts select Go test names, not fixture filenames, and require no edit. The final changed-path census contains only the eight fixture renames, seven count-record moves, two authorized reader edits and this evidence file.
