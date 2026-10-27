# Item 8 stage 2: fx7 twins

Base: `origin/main` at `d579ff21e10791f78ab269628849513f7db5a39c`. Precedents: `d579ff21` (read in full, including v2/TWINS.md) and `bf3ff0489b`.

The directory matched exactly the ten requested fixtures. All ten fixtures were moved with `git mv`; their source bytes are unchanged. Each count record was moved with its fixture, changing only the path in its row; all six numbers are unchanged.

## Comparison

Each outcome below records exit code, stdout and stderr. `\n` is an LF byte. N and J identify native and JavaScript runs; identical identifiers mean byte-identical stdout/stderr and equal exit codes. Before is `.a`; after is `.ts`.

| File (now .ts) | Native before | Native after | JavaScript before | JavaScript after | Identical | Type predicate | Non-null assertion | Definite assignment |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| depth-200.ts | N1 | N1 | J1 | J1 | yes | no | no | no |
| depth-2000.ts | N2 | N2 | J2 | J2 | yes | no | no | no |
| depth-65.ts | N3 | N3 | J3 | J3 | yes | no | no | no |
| p09.ts | N4 | N4 | J4 | J4 | yes | no | no | no |
| p32.ts | N5 | N5 | J5 | J5 | yes | no | no | no |
| p33.ts | N6 | N6 | J6 | J6 | yes | no | no | no |
| p67.ts | N7 | N7 | J7 | J7 | yes | no | no | no |
| p72.ts | N8 | N8 | J8 | J8 | yes | no | no | no |
| scalar-static.ts | N9 | N9 | J9 | J9 | yes | no | no | no |
| scalar-tuple.ts | N10 | N10 | J10 | J10 | yes | no | no | no |

No file failed to be a twin. Every native result also equals its JavaScript result.

| Outcomes | Exit | Stdout (JSON string) | Stderr (JSON string) |
| --- | --- | --- | --- |
| N1, J1 | 0 | `"true\n"` | `""` |
| N2, J2 | 0 | `"true\n"` | `""` |
| N3, J3 | 0 | `"true\n"` | `""` |
| N4, J4 | 70 | `""` | `"adamic: panic: field read failed: viewed.value matches no member of readonly [string, number]; expected readonly [string, number], found array\n"` |
| N5, J5 | 70 | `"3\n"` | `"adamic: panic: field read failed: viewed.name is not a string; expected string, found number\n"` |
| N6, J6 | 70 | `"3\n"` | `"adamic: panic: field read failed: value.name is not a string; expected string, found number\n"` |
| N7, J7 | 70 | `""` | `"adamic: panic: field read failed: viewed.value matches no member of number \| Inner; expected number \| Inner, found function\n"` |
| N8, J8 | 70 | `""` | `"adamic: panic: field read failed: viewed.value matches no member of readonly [string, number]; expected readonly [string, number], found array\n"` |
| N9, J9 | 70 | `""` | `"adamic: panic: field read failed: viewed.value is not a boolean; expected boolean, found function\n"` |
| N10, J10 | 70 | `""` | `"adamic: panic: field read failed: viewed.value is not a boolean; expected boolean, found array\n"` |

## Method and readers

Required setup: `export GOPROXY='https://proxy.golang.org|direct'; timeout --kill-after=60s 1800 bash cloud/setup.sh > /tmp/setup.log 2>&1` (exit 0). Subsequent Go commands source `/workspace/adamic-tools/env.sh`.

Before any fixture rename, a temporary capture program called `load.Load`, `lower.Lower`, `native.C` / `native.Build` with `native.Options{}`, and `javascript.JavaScript`. It executed native binaries directly and JavaScript with `node --disable-warning=ExperimentalWarning oracle/node.mjs <generated program.mjs>`, matching `releasedUncached` and `onJavaScriptBackend`. The same binary ran all ten again after the rename. Separate JSON records saved stdout, stderr and exit code for each file/backend; all 20 records compared equal, including a byte comparison of the complete JSON files. Logs: `/tmp/fx7-before.log`, `/tmp/fx7-after.log`; records: `/tmp/fx7-before.json`, `/tmp/fx7-after.json`. The temporary capture source is removed before commit.

`git grep -n` for each of the ten basenames across the whole repository (saved in `/tmp/fx7-readers.log`), supplemented with `git grep -n 'interface-downcasts/fx7'` and indirect-reader searches, found one generated fixture registration path in `internal/oracle/fx7_depth_texts_test.go`, its runtime tests using `interfaceFixture`, and ten counts records. Only the registration's `.a` suffix changes to `.ts`; `interfaceFixturePath` already resolves the unique on-disk extension. No Go logic or expected outputs changed, and no tests were added.

No status.json, manifest, known list, a-check header, or recorded-outcome row moves. Other basename matches refer to different fixtures or historical evidence and remain unchanged. No stage3/fixtures files are touched; all readers of these ten fixtures live in internal/oracle.

## Mutant

Only `p09.ts.counts` was temporarily moved back to `p09.a.counts`; its row still named `.ts` and the fixture remained `.ts`. The mutation was restored before the full suite. No mutant change enters the commit.

Command: `go test ./internal/oracle -run '^TestCountsFilesCensus$' -count=1 -json > /tmp/fx7-mutant-counts.jsonl 2> /tmp/fx7-mutant-counts.stderr` (exit 1; 0 passed test events, 1 failed test event). Witness: `TestCountsFilesCensus`.

```text
count_files_test.go:158: fixture stage3/interface-downcasts/fx7/p09.ts needs counts file testdata/counts/runtime/stage3/interface-downcasts/fx7/p09.ts.counts: open testdata/counts/runtime/stage3/interface-downcasts/fx7/p09.ts.counts: no such file or directory
```

No recorded-outcome rows move, so no recorded-row mutant is applicable.

## Validation

Prepared `/tmp/fx7-typescript` from the repository and tag in `stage3/source.json`, and verified HEAD is exactly `050880ce59e30b356b686bd3144efe24f875ebc8`. Installed the locked API dependencies with `npm ci --prefix stage3/api` (exit 0; log `/tmp/fx7-api-setup.log`). No tracked dependency files changed.

Full package command, after sourcing `/workspace/adamic-tools/env.sh`:

```sh
ADAMIC_TYPESCRIPT_SOURCE=/tmp/fx7-typescript go test ./internal/oracle/... -count=1 -json -timeout=60m > /tmp/fx7-oracle.jsonl 2> /tmp/fx7-oracle.stderr
```

PASS, exit 0: 1 package; 1777 top-level tests and 6153 subtests passed (7930 passing test events). 442 top-level tests and 99 subtests skipped (541 skip events). Zero failures. Package elapsed: 1182.074 seconds. Stderr is empty.

All ten `.ts` fixtures passed their `TestCountsAreRecorded_*`, `TestStage0EmitsDeterministically_*`, and `TestNativeAgreesWithNode_*` checks. The FX7 depth and pinned-diagnostic tests passed too. No stage3/fixtures files were touched, so its conditional package run is not applicable.

Final scope audit verified unchanged bodies for all ten renamed fixtures, unchanged numbers in all ten count rows, and exactly one `.a` to `.ts` suffix edit in the sole changed Go file. No Go logic, expected outputs or tests were added or changed beyond that path suffix. `git diff --check` and the Go formatting check passed. Temporary capture sources, mutants, dependency installs and test logs are outside the final diff.

## Counts comparison

Each full before row was compared with its after row after substituting only the fixture path. These are the six numeric columns before and after; every numeric diff is empty.

| Fixture | Before numbers | After numbers | Numeric diff |
| --- | --- | --- | --- |
| depth-200.ts | 201, 201, 205, 209, 201, 0 | 201, 201, 205, 209, 201, 0 | empty |
| depth-2000.ts | 2001, 2001, 2005, 2009, 2001, 0 | 2001, 2001, 2005, 2009, 2001, 0 | empty |
| depth-65.ts | 66, 66, 70, 74, 66, 0 | 66, 66, 70, 74, 66, 0 | empty |
| p09.ts | 2, 0, 6, 5, 2, 0 | 2, 0, 6, 5, 2, 0 | empty |
| p32.ts | 4, 2, 9, 11, 3, 0 | 4, 2, 9, 11, 3, 0 | empty |
| p33.ts | 5, 2, 10, 11, 4, 0 | 5, 2, 10, 11, 4, 0 | empty |
| p67.ts | 2, 0, 8, 6, 2, 0 | 2, 0, 8, 6, 2, 0 | empty |
| p72.ts | 3, 0, 6, 5, 3, 0 | 3, 0, 6, 5, 3, 0 | empty |
| scalar-static.ts | 2, 0, 6, 6, 2, 0 | 2, 0, 6, 6, 2, 0 | empty |
| scalar-tuple.ts | 2, 0, 5, 5, 2, 0 | 2, 0, 5, 5, 2, 0 | empty |
