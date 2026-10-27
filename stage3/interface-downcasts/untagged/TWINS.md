# Item 8 stage 2: untagged twins

Base: `origin/main` at `d579ff21e10791f78ab269628849513f7db5a39c`. Precedents: `d579ff21` (read in full, including its `TWINS.md`) and `bf3ff0489b`.

The directory matched all 26 requested fixtures, plus ten excluded fixtures: `array-union-{empty,good,mixed,nested,wrong}.a`, `{binding-name,option-element,structural}-source-absent.a`, and `callable-union-{nested,wrong}.a`. The excluded files remain unchanged. All 26 requested fixtures and their count records moved with `git mv`. Source bytes are unchanged; count rows change only the path, with all six numbers unchanged.

## Comparison

Each outcome records exit code, stdout and stderr. `\n` is an LF byte. N and J identify native and JavaScript; identical identifiers mean byte-identical stdout/stderr and equal exit codes. Before is `.a`; after is `.ts`.

| File (now .ts) | Native before | Native after | JavaScript before | JavaScript after | Identical | Type predicate | Non-null assertion | Definite assignment |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| binding-name-source-good.ts | N1 | N1 | J1 | J1 | yes | no | no | no |
| binding-name-source-nested.ts | N2 | N2 | J2 | J2 | yes | no | no | no |
| binding-name-source-wrong.ts | N3 | N3 | J3 | J3 | yes | no | no | no |
| callable-union-good-number.ts | N4 | N4 | J4 | J4 | yes | no | no | no |
| callable-union-good-string.ts | N5 | N5 | J5 | J5 | yes | no | no | no |
| callable-union-optional-boundary.ts | N6 | N6 | J6 | J6 | yes | no | no | no |
| class-data-good.ts | N7 | N7 | J7 | J7 | yes | no | no | no |
| class-data-wrong.ts | N8 | N8 | J8 | J8 | yes | no | no | no |
| flow-callback-good.ts | N9 | N9 | J9 | J9 | yes | no | no | no |
| flow-callback-wrong.ts | N10 | N10 | J10 | J10 | yes | no | no | no |
| flow-generic-good.ts | N11 | N11 | J11 | J11 | yes | no | no | no |
| flow-generic-wrong.ts | N12 | N12 | J12 | J12 | yes | no | no | no |
| flow-helper-good.ts | N13 | N13 | J13 | J13 | yes | no | no | no |
| flow-helper-wrong.ts | N14 | N14 | J14 | J14 | yes | no | no | no |
| flow-stored-good.ts | N15 | N15 | J15 | J15 | yes | no | no | no |
| flow-stored-wrong.ts | N16 | N16 | J16 | J16 | yes | no | no | no |
| option-element-source-good.ts | N17 | N17 | J17 | J17 | yes | no | no | no |
| option-element-source-nested.ts | N18 | N18 | J18 | J18 | yes | no | no | no |
| option-element-source-wrong.ts | N19 | N19 | J19 | J19 | yes | no | no | no |
| recursive-absent.ts | N20 | N20 | J20 | J20 | yes | no | no | no |
| recursive-good.ts | N21 | N21 | J21 | J21 | yes | no | no | no |
| recursive-nested.ts | N22 | N22 | J22 | J22 | yes | no | no | no |
| recursive-wrong.ts | N23 | N23 | J23 | J23 | yes | no | no | no |
| structural-source-good.ts | N24 | N24 | J24 | J24 | yes | no | no | no |
| structural-source-nested.ts | N25 | N25 | J25 | J25 | yes | no | no | no |
| structural-source-wrong.ts | N26 | N26 | J26 | J26 | yes | no | no | no |

No file failed to be a twin. Every native result also equals its JavaScript result. Thirteen fixtures finish with exit 0, and thirteen retain their checked panic with exit 70. None uses a type predicate, non-null assertion or definite assignment.

| Outcomes | Exit | Stdout (JSON string) | Stderr (JSON string) |
| --- | --- | --- | --- |
| N1, J1 | 0 | `"true\ntrue\ntrue\ntrue\ntrue\ntrue\ntrue\ntrue\ntrue\ntrue\ntrue\ntrue\n"` | `""` |
| N2, J2 | 70 | `""` | `"adamic: panic: field read failed: value.payload matches no member of { readonly label: string; } \| { readonly label: string; } \| { readonly label: string; } \| { readonly label: string; } \| { readonly label: string; } \| { readonly label: string; } \| { readonly label: string; } \| ... 4 more ... \| ...; expected { readonly label: string; } \| { readonly label: string; } \| { readonly label: string; } \| { readonly label: string; } \| { readonly label: string; } \| { readonly label: string; } \| { readonly label: string; } \| ... 4 more ... \| ..., found object\n"` |
| N3, J3 | 70 | `""` | `"adamic: panic: field read failed: view.value matches no member of Target; expected Target, found object\n"` |
| N4, J4 | 0 | `"true\n"` | `""` |
| N5, J5 | 0 | `"true\n"` | `""` |
| N6, J6 | 0 | `"true\n"` | `""` |
| N7, J7 | 0 | `"true\n"` | `""` |
| N8, J8 | 70 | `""` | `"adamic: panic: field read failed: view.value matches no member of Target; expected Target, found object\n"` |
| N9, J9 | 0 | `"true\n"` | `""` |
| N10, J10 | 70 | `""` | `"adamic: panic: field read failed: holder.value matches no member of Target; expected Target, found object\n"` |
| N11, J11 | 0 | `"true\n"` | `""` |
| N12, J12 | 70 | `""` | `"adamic: panic: field read failed: holder.value matches no member of Target; expected Target, found object\n"` |
| N13, J13 | 0 | `"true\n"` | `""` |
| N14, J14 | 70 | `""` | `"adamic: panic: field read failed: holder.value matches no member of Target; expected Target, found object\n"` |
| N15, J15 | 0 | `"true\n"` | `""` |
| N16, J16 | 70 | `""` | `"adamic: panic: field read failed: holder.value matches no member of Target; expected Target, found object\n"` |
| N17, J17 | 0 | `"true\ntrue\ntrue\ntrue\ntrue\n"` | `""` |
| N18, J18 | 70 | `""` | `"adamic: panic: field read failed: value.payload matches no member of { readonly label: string; } \| { readonly label: string; } \| { readonly label: string; } \| { readonly label: string; } \| { readonly label: string; }; expected { readonly label: string; } \| { readonly label: string; } \| { readonly label: string; } \| { readonly label: string; } \| { readonly label: string; }, found object\n"` |
| N19, J19 | 70 | `""` | `"adamic: panic: field read failed: view.value matches no member of Target; expected Target, found object\n"` |
| N20, J20 | 0 | `"true\n"` | `""` |
| N21, J21 | 0 | `"true\n"` | `""` |
| N22, J22 | 70 | `""` | `"adamic: panic: field read failed: view.value matches no member of Target; expected Target, found object\n"` |
| N23, J23 | 70 | `""` | `"adamic: panic: field read failed: view.value matches no member of Target; expected Target, found object\n"` |
| N24, J24 | 0 | `"true\ntrue\n"` | `""` |
| N25, J25 | 70 | `""` | `"adamic: panic: field read failed: view.value matches no member of Target; expected Target, found object\n"` |
| N26, J26 | 70 | `""` | `"adamic: panic: field read failed: view.value matches no member of Target; expected Target, found object\n"` |

## Method and readers

Required setup: `export GOPROXY='https://proxy.golang.org|direct'; timeout --kill-after=60s 1800 bash cloud/setup.sh > /tmp/setup.log 2>&1`. PASS, exit 0; completed in 414.336 seconds. Each subsequent build/test shell sourced `/workspace/adamic-tools/env.sh`.

Before any fixture rename, a temporary capture program called `load.Load`, `lower.Lower`, `native.C` / `native.Build` with `native.Options{}`, and `javascript.JavaScript`. It executed native binaries directly and JavaScript with `node --disable-warning=ExperimentalWarning oracle/node.mjs <generated program.mjs>`, matching `releasedUncached` and `onJavaScriptBackend`. The same binary ran all 26 again after renaming. Separate JSON records saved stdout, stderr and exit code for every file/backend; all 52 records compared equal. Logs: `/tmp/item8-untagged-before.log`, `/tmp/item8-untagged-after.log`; records: `/tmp/item8-untagged-before.json`, `/tmp/item8-untagged-after.json`. The temporary capture source is removed before commit.

For each of the 26 basenames, `git grep -n -F <basename>` searched the whole repository. Full search evidence is `/tmp/item8-untagged-readers.log`. Supplementary directory and generated-path searches found 26 explicit registrations in `checked_views_v2_counts_test.go`, one explicit path in `checked_views_v2_source_test.go`, and the source/migration tests using the already extension-aware `interfaceFixturePath`. Only those 27 path literals change in Go. The generated `.a` path in the absent-only branch remains correct for the three excluded source-absent fixtures. No Go logic, tests or expected outputs changed.

The 26 runtime counts files move with their fixtures. For every file, the original six numeric columns were compared to the renamed row and were equal. `source-refusals.json` keys are family/variant names, not source paths, so it does not move. No status, manifest, fixture-union, known-list or other recorded-outcome path row needs an edit. Historical `review/compiler` evidence retains the paths actually run; no dated evidence or a-check records change. No stage3 Go package reads these 26 fixtures; no `stage3/fixtures` path is touched.

## Mutants

Each mutation was applied alone and restored before the full suite. No mutant edits enter the commit. Both use `go test ./internal/oracle -run '^TestCountsFilesCensus$' -count=1 -v`, and both exit 1 with 0 passed and 1 failed test events.

### Counts filename

Only `binding-name-source-good.ts.counts` was moved back to `.a.counts`; its fixture and row remained `.ts`. Log: `/tmp/item8-untagged-mutant-counts.log`. Red test: `TestCountsFilesCensus`.

```text
count_files_test.go:158: fixture stage3/interface-downcasts/untagged/fixtures/binding-name-source-good.ts needs counts file testdata/counts/runtime/stage3/interface-downcasts/untagged/fixtures/binding-name-source-good.ts.counts: open testdata/counts/runtime/stage3/interface-downcasts/untagged/fixtures/binding-name-source-good.ts.counts: no such file or directory
```

### Recorded counts row

Only the row key in `binding-name-source-good.ts.counts` was changed back to `.a`; the filename and fixture remained `.ts`. Log: `/tmp/item8-untagged-mutant-row.log`. Red test: `TestCountsFilesCensus`.

```text
count_files_test.go:158: fixture stage3/interface-downcasts/untagged/fixtures/binding-name-source-good.ts needs counts file testdata/counts/runtime/stage3/interface-downcasts/untagged/fixtures/binding-name-source-good.ts.counts: testdata/counts/runtime/stage3/interface-downcasts/untagged/fixtures/binding-name-source-good.ts.counts: invalid runtime count row for stage3/interface-downcasts/untagged/fixtures/binding-name-source-good.ts: "| stage3/interface-downcasts/untagged/fixtures/binding-name-source-good.a | 47 | 47 | 84 | 132 | 36 | 0 |"
```

## Validation

The pinned upstream TypeScript checkout `/tmp/item8-typescript` was prepared from `stage3/source.json` and verified at `050880ce59e30b356b686bd3144efe24f875ebc8`. `npm ci --prefix stage3/api` installed the locked API dependencies. Logs: `/tmp/item8-typescript-setup.log`, `/tmp/item8-api-setup.log`. No tracked dependency file changed.

Full package command: `ADAMIC_TYPESCRIPT_SOURCE=/tmp/item8-typescript go test ./internal/oracle/... -count=1 -json -timeout=60m > /tmp/item8-untagged-oracle.jsonl 2> /tmp/item8-untagged-oracle.stderr`. PASS, exit 0: 1 package; 1777 top-level tests and 6153 subtests passed (7930 passing test events). 442 top-level tests and 99 subtests skipped (541 skip events). Zero failures. Package elapsed: 1187.106 seconds.

All untagged source-dispatch, callable, flow, recursive and own-class-data tests passed, along with the restored counts census and full count-family checks. No `stage3/fixtures` file was touched, so its package suite was not required.

Final scope audit verified unchanged source bytes for all 26 renamed fixtures, unchanged numeric columns in all 26 count records, untouched excluded fixtures, exactly the 27 permitted `.a` to `.ts` path literal replacements in the two Go readers, and no unexpected tracked or untracked changes. `git diff --check` passed. The final commit contains no capture utility, mutant edits, dependency installs or logs.
