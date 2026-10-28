# Item 8 stage 2: lane4b twins

Base: `typescript/item8-twins-untagged` at `3993a2a3a430160702290b54b29d9fcede51a0fe`. Precedents: `d579ff21` (read in full, including its `TWINS.md`) and `bf3ff0489b`.

The directory matched all 13 requested fixtures, plus five excluded fixtures: `optional-{absent,receiver}.a` and `comment-{boolean,flags-wrong,good}.a`. The excluded files remain unchanged. All 13 requested fixtures and their count records moved with `git mv`. Source bytes are unchanged; count rows change only the path, with all six numbers unchanged.

## Comparison

Each outcome records exit code, stdout and stderr. `\n` is an LF byte. N and J identify native and JavaScript; identical identifiers mean byte-identical stdout/stderr and equal exit codes. Before is `.a`; after is `.ts`.

| File (now .ts) | Native before | Native after | JavaScript before | JavaScript after | Identical | Type predicate | Non-null assertion | Definite assignment |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| diagnostic-boolean.ts | N1 | N1 | J1 | J1 | yes | no | no | no |
| diagnostic-code-wrong.ts | N2 | N2 | J2 | J2 | yes | no | no | no |
| diagnostic-good.ts | N3 | N3 | J3 | J3 | yes | no | no | no |
| diagnostic-wrong.ts | N4 | N4 | J4 | J4 | yes | no | no | no |
| literal-boolean.ts | N5 | N5 | J5 | J5 | yes | no | no | no |
| literal-good.ts | N6 | N6 | J6 | J6 | yes | no | no | no |
| literal-negative-wrong.ts | N7 | N7 | J7 | J7 | yes | no | no | no |
| literal-text-wrong.ts | N8 | N8 | J8 | J8 | yes | no | no | no |
| node-indicator-false.ts | N9 | N9 | J9 | J9 | yes | no | no | no |
| node-indicator-flags-wrong.ts | N10 | N10 | J10 | J10 | yes | no | no | no |
| node-indicator-good.ts | N11 | N11 | J11 | J11 | yes | no | no | no |
| node-indicator-wrong.ts | N12 | N12 | J12 | J12 | yes | no | no | no |
| required-missing.ts | N13 | N13 | J13 | J13 | yes | no | no | no |

No file failed to be a twin. Every native result also equals its JavaScript result. Three fixtures finish with exit 0; ten retain their checked panic with exit 70. None uses a type predicate, non-null assertion or definite assignment.

| Outcomes | Exit | Stdout (JSON string) | Stderr (JSON string) |
| --- | --- | --- | --- |
| N1, J1 | 70 | `""` | `"adamic: panic: field read failed: value.messageText matches no member of string \| Chain; expected string \| Chain, found boolean\n"` |
| N2, J2 | 70 | `""` | `"adamic: panic: field read failed: value.messageText matches no member of string \| Chain; expected string \| Chain, found object\n"` |
| N3, J3 | 0 | `"plain\nnested\n"` | `""` |
| N4, J4 | 70 | `""` | `"adamic: panic: field read failed: value.messageText matches no member of string \| Chain; expected string \| Chain, found object\n"` |
| N5, J5 | 70 | `""` | `"adamic: panic: field read failed: type.value matches no member of string \| number \| PseudoBigInt; expected string \| number \| PseudoBigInt, found boolean\n"` |
| N6, J6 | 0 | `"text\n42\ntrue:123\n43\nfalse:456\n"` | `""` |
| N7, J7 | 70 | `""` | `"adamic: panic: field read failed: type.value matches no member of string \| number \| PseudoBigInt; expected string \| number \| PseudoBigInt, found object\n"` |
| N8, J8 | 70 | `""` | `"adamic: panic: field read failed: member.base10Value is not a string; expected string, found number\n"` |
| N9, J9 | 70 | `""` | `"adamic: panic: field read failed: value.externalModuleIndicator matches no member of true \| Node \| undefined; expected true \| Node \| undefined, found boolean\n"` |
| N10, J10 | 70 | `""` | `"adamic: panic: field read failed: member.flags is not a number; expected number, found boolean\n"` |
| N11, J11 | 0 | `"true\nIdentifier\nabsent\n"` | `""` |
| N12, J12 | 70 | `""` | `"adamic: panic: field read failed: value.externalModuleIndicator matches no member of true \| Node \| undefined; expected true \| Node \| undefined, found object\n"` |
| N13, J13 | 70 | `""` | `"adamic: panic: field read failed: value.value is not initialized; expected true \| Node \| undefined, found missing\n"` |

## Method and readers

Required setup: `export GOPROXY='https://proxy.golang.org|direct'; timeout --kill-after=60s 1800 bash cloud/setup.sh > /tmp/setup.log 2>&1`. PASS, exit 0; completed in 11.694 seconds. Each subsequent build/test shell sourced `/workspace/adamic-tools/env.sh`.

Before any fixture rename, the temporary capture utility built for the untagged migration called `load.Load`, `lower.Lower`, `native.C` / `native.Build` with `native.Options{}`, and `javascript.JavaScript`. The compiler implementation is unchanged. It executed native binaries directly and JavaScript with `node --disable-warning=ExperimentalWarning oracle/node.mjs <generated program.mjs>`, matching `releasedUncached` and `onJavaScriptBackend`. The same binary ran all 13 again after renaming. Separate JSON records saved stdout, stderr and exit code for every file/backend; all 26 records compared equal. Logs: `/tmp/item8-lane4b-before.log`, `/tmp/item8-lane4b-after.log`; records: `/tmp/item8-lane4b-before.json`, `/tmp/item8-lane4b-after.json`. No capture utility enters the commit.

For each of the 13 basenames, `git grep -n -F <basename>` searched the whole repository. Full evidence is `/tmp/item8-lane4b-readers.log`. Supplementary directory and generated-path searches found 13 explicit registrations in `checked_views_v2_counts_test.go`, 13 runtime count records, and the shared generated path in `checked_views_v2_object_primitive_test.go`. No status, manifest, fixture-union, known-list or other recorded-outcome path row needs an edit. Historical `review/compiler` evidence retains the paths actually run; no dated evidence or a-check records change. No stage3 Go package reads these 13 fixtures; no `stage3/fixtures` path is touched.

The 13 explicit Go paths change only `.a` to `.ts`. The sole authorized logic change, per the user's lane4b ruling, replaces the generated path expression with `filepath.Abs(filepath.Join(repository, interfaceFixturePath("lane4b/fixtures/"+sample.name)))`. It passes exactly the same fixture names through the existing helper; the helper prefixes `stage3/interface-downcasts/`, selects the sole extension on disk, and panics unless exactly one of `.a` and `.ts` exists. Excluded optional fixtures therefore keep their `.a` paths, and excluded comment fixtures retain their existing skip. No other Go logic, test or expected output changed.

Every original count record was compared with its renamed record: only the path changes, and all six numeric columns are equal.

## Mutants

Each mutation was applied alone and restored before the full suite. No mutant edits enter the commit.

### Counts filename

Only `diagnostic-good.ts.counts` was moved back to `.a.counts`; its fixture and row remained `.ts`. Command: `go test ./internal/oracle -run '^TestCountsFilesCensus$' -count=1 -v`. Exit 1; 0 passed and 1 failed test events. Log: `/tmp/item8-lane4b-mutant-counts.log`. Red test: `TestCountsFilesCensus`.

```text
count_files_test.go:158: fixture stage3/interface-downcasts/lane4b/fixtures/diagnostic-good.ts needs counts file testdata/counts/runtime/stage3/interface-downcasts/lane4b/fixtures/diagnostic-good.ts.counts: open testdata/counts/runtime/stage3/interface-downcasts/lane4b/fixtures/diagnostic-good.ts.counts: no such file or directory
```

### Recorded counts row

Only the row key in `diagnostic-good.ts.counts` was changed back to `.a`; the filename and fixture remained `.ts`. Command: `go test ./internal/oracle -run '^TestCountsFilesCensus$' -count=1 -v`. Exit 1; 0 passed and 1 failed test events. Log: `/tmp/item8-lane4b-mutant-row.log`. Red test: `TestCountsFilesCensus`.

```text
count_files_test.go:158: fixture stage3/interface-downcasts/lane4b/fixtures/diagnostic-good.ts needs counts file testdata/counts/runtime/stage3/interface-downcasts/lane4b/fixtures/diagnostic-good.ts.counts: testdata/counts/runtime/stage3/interface-downcasts/lane4b/fixtures/diagnostic-good.ts.counts: invalid runtime count row for stage3/interface-downcasts/lane4b/fixtures/diagnostic-good.ts: "| stage3/interface-downcasts/lane4b/fixtures/diagnostic-good.a | 3 | 3 | 15 | 17 | 3 | 0 |"
```

### Hard-coded .a reader

Restored only the original `filepath.Abs("../../stage3/interface-downcasts/lane4b/fixtures/" + sample.name + ".a")` expression. Command: `go test ./internal/oracle -run '^TestCheckedViewObjectPrimitiveSource$/^diagnostic-good$' -count=1 -v`. Exit 1; 0 passed and 2 failed test events (parent and subtest). Log: `/tmp/item8-lane4b-mutant-hard-coded-a.log`. Red witness: `TestCheckedViewObjectPrimitiveSource/diagnostic-good`.

```text
cache_test.go:280: open /workspace/adamic/stage3/interface-downcasts/lane4b/fixtures/diagnostic-good.a: no such file or directory
```

## Validation

The existing pinned upstream checkout `/tmp/item8-typescript` was verified at `050880ce59e30b356b686bd3144efe24f875ebc8`, matching `stage3/source.json`. The locked `stage3/api` dependencies prepared for untagged remain installed. No tracked dependency file changed.

Full package command: `ADAMIC_TYPESCRIPT_SOURCE=/tmp/item8-typescript go test ./internal/oracle/... -count=1 -json -timeout=60m > /tmp/item8-lane4b-oracle.jsonl 2> /tmp/item8-lane4b-oracle.stderr`. PASS, exit 0: 1 package; 1777 top-level tests and 6153 subtests passed (7930 passing test events). 442 top-level tests and 99 subtests skipped (541 skip events). Zero failures. Package elapsed: 904.950 seconds.

All 13 renamed source subtests and both excluded optional `.a` controls passed in `TestCheckedViewObjectPrimitiveSource`; the three excluded comment cases retained their existing skips. The restored counts census and all count-family checks passed. No `stage3/fixtures` file was touched, so its package suite was not required.

Final scope audit verified unchanged source bytes for all 13 renamed fixtures, unchanged numeric columns in all 13 count records, untouched excluded fixtures, exactly the 13 permitted path literal replacements in the counts registrations, and exactly the authorized helper substitution in the mixed-extension reader. No tests or expected outputs changed. `git diff --check` passed. The final commit contains no capture utility, mutant edits, dependency installs or logs.
