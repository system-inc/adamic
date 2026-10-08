- Built a hidden-source census with per-file ranges and the ten largest residual regions.
- Measured compiler commit `ed6e29751ee47d86fad450cd1674139883bc0f70`, TypeScript 6.0.3.
- Observed **3,970,749 / 10,615,807 bytes hidden (37.404118%)**.
- Eight arithmetic tests, a real census witness and the independent byte-mask audit pass; nine mutants are caught.
- Scope is ledger-defined lowering coverage; no native output, semantic oracle or full gate was run.

[RESULT.json](RESULT.json) contains every file's bytes, hidden share, source hash and
residual half-open ranges. It retains the top ten's exact stopping reasons and
attempt owners. [evidence/full.jsonl.gz](evidence/full.jsonl.gz) is the fresh full
census; [evidence/stock.json.gz](evidence/stock.json.gz) is the independent stock
TypeScript span catalogue.

The arithmetic is **6,815,024 union bytes minus
2,844,275 independently examined bytes = 3,970,749 hidden bytes**.
There are 17,643 boundary records, 147 checker-skipped declaration bodies,
16 dependency-skip events and 10,544 independent attempts.
The denominator includes all 82 regular files under the adapted `src/compiler`,
including comments, whitespace, generated inputs and three JSON files.
For the 79 TypeScript files alone it is 10,009,820 bytes, giving
**39.668535%** hidden.

`Boundary` findings carry `start`/`end`, including signature/prologue whole-body
selections. `split_checker_body` units carry `body_start`/`body_end` and diagnostics.
`SkippedDependency` events identify additional skipped bodies; stock TypeScript
supplies their spans, including function expressions. All ranges are UTF-8 bytes.
Each independent attempted declaration exposes its span minus its own boundaries
and diagnosed descendants. A skipped ancestor does not suppress that exposure.
Union the blocked spans per file, then subtract the union of these exposures.
A nested attempt's own failed region remains hidden. No spans are added as lengths
before union. The stock catalogue and source hashes are checked before reporting.

The ten largest connected residual regions follow. Lines include the span's leading
trivia; end lines contain the final included byte. Exact diagnostics are in JSON.

| File:lines | Hidden bytes | Stopping reason |
|---|---:|---|
| `visitorPublic.ts:619-1798` | 60,674 | Computed field name, failed statement |
| `parser.ts:504-1136` | 49,243 | Computed field name, failed statement |
| `factory/nodeFactory.ts:492-1166` | 27,831 | `__String` representation, whole-body function boundary |
| `checker.ts:51813-52186` | 26,623 | Checked-view `modifiers` representation; checker-rejected enclosing body |
| `program.ts:1515-1968` | 25,095 | Checker-rejected body, including TS2375 |
| `checker.ts:1486-1956` | 24,907 | Checker-rejected body, including TS2322 |
| `utilities.ts:11518-11904` | 24,676 | Destructured name; `Identifier \| __String` representation |
| `transformers/esDecorators.ts:670-1047` | 22,342 | `ImmediatelyInvokedArrowFunction` result, whole-body function boundary |
| `checker.ts:2046-2412` | 22,159 | Checker-rejected enclosing body, including TS2322 |
| `program.ts:4073-4401` | 21,031 | Checker-rejected enclosing and nested bodies, including TS2375 and TS2345 |

The largest per-file totals are below; RESULT.json lists all 82 files, including zeros.

| File | Hidden bytes | Share of file |
|---|---:|---:|
| `checker.ts` | 1,468,358 | 46.55% |
| `factory/nodeFactory.ts` | 205,100 | 60.92% |
| `utilities.ts` | 184,054 | 35.76% |
| `program.ts` | 161,141 | 59.40% |
| `transformers/es2015.ts` | 132,345 | 57.10% |
| `transformers/classFields.ts` | 98,577 | 62.92% |
| `emitter.ts` | 97,540 | 35.52% |
| `transformers/esDecorators.ts` | 93,640 | 73.51% |
| `visitorPublic.ts` | 80,826 | 91.73% |
| `tsbuildPublic.ts` | 78,413 | 68.39% |

Reproduce from the repository root, using fresh scratch paths:

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/hidden-setup.log 2>&1
source /workspace/adamic-tools/env.sh
stage3/apply.sh /tmp/hidden-adapted > /tmp/hidden-apply.log 2>&1
python3 stage3/census/latent/make_overlay.py "$PWD" /tmp/hidden-overlay > /tmp/hidden-overlay.log 2>&1
gofmt -w /tmp/hidden-overlay/*.go
go build -buildvcs=false -overlay=/tmp/hidden-overlay/overlay.json -o /tmp/hidden-census ./stage3/census/latent/tool > /tmp/hidden-build.log 2>&1
python3 stage3/census/latent/audit.py /tmp/hidden-census > /tmp/hidden-latent-audit.log 2>&1
LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 /tmp/hidden-census /tmp/hidden-adapted/src/compiler /tmp/hidden-full.jsonl > /tmp/hidden-run.log 2>&1
NODE_PATH="$HOME/.cache/adamic-stage3/api/node_modules" node stage3/census/hidden/units.cjs /tmp/hidden-adapted/src/compiler /tmp/hidden-stock.json > /tmp/hidden-stock.log 2>&1
python3 stage3/census/hidden/hidden.py /tmp/hidden-adapted/src/compiler /tmp/hidden-full.jsonl /tmp/hidden-stock.json /tmp/hidden-result.json --commit "$(git rev-parse HEAD)" > /tmp/hidden-result.log 2>&1
python3 stage3/census/hidden/test_hidden.py > /tmp/hidden-tests.log 2>&1
NODE_PATH="$HOME/.cache/adamic-stage3/api/node_modules" python3 stage3/census/hidden/integration.py /tmp/hidden-census /tmp/hidden-integration > /tmp/hidden-integration.log 2>&1
python3 stage3/census/hidden/audit.py /tmp/hidden-full.jsonl /tmp/hidden-stock.json /tmp/hidden-result.json > /tmp/hidden-result-audit.log 2>&1
python3 stage3/census/hidden/prove.py /tmp/hidden-full.jsonl /tmp/hidden-stock.json /tmp/hidden-result.json /tmp/hidden-proof > /tmp/hidden-proof.log 2>&1
```

All listed commands passed in this run. Integration observed 141 hidden bytes
against an independent byte-set oracle. The arithmetic suite also compares 500
random union/subtraction cases with byte sets. Mutants and their intended catchers:

| Mutant | Catcher |
|---|---|
| Ignore dependency skips | Expression body must retain 17 hidden bytes, with duplicate events counted once; observed 0 |
| Ignore checker-skipped spans | Fake skipped body must grow total by exactly 17 bytes; observed 0 |
| Do not union overlaps | Two overlapping spans must total 30 bytes; observed 40 |
| Drop independent coverage | Nested residual must total 45 bytes; observed 80 |
| Let attempted parent expose skipped child | Child must retain 20 hidden bytes; observed 0 |
| Add 17 to headline total | Independent byte-mask headline assertion |
| Add 1 to a file's total | Independent byte-mask per-file assertion |
| Add 1 to largest region's size | Independent top-ten ranking assertion |
| Add 1 to largest region's start line | Independent line-endpoint assertion |

Each mutant exits 1 at its intended assertion; `prove.py` exits 0 only when all
are caught. Logs are retained in evidence/. The legacy latent audit also passes,
including its body-scope, injected-finding and attribution mutants.

The original measurement's full audit failed because it expected
`a string as a condition` and `a number as a condition` refusals. Those expectations
are now retired: conditions landed with JavaScript truthiness. Both condition
inputs remain in the witness; three distinct `var` sites preserve the continuation
check instead of weakening it to a single refusal. The diagnosed grandchild is
uninstantiated and generic so nested-function hoisting does not veto that witness.

`python3 stage3/census/latent/full_audit.py /tmp/hidden-census /tmp/hidden-full-audit-conditions-final > /tmp/hidden-full-audit-conditions.log 2>&1`
now exits 0. The first-error-only mutant is caught by the three-site assertion;
the failed-state-retention mutant is caught by the poison-binding rollback
assertion. [Passing audit log](evidence/full-audit-conditions.log.txt) records both.
The earlier failure log remains as historical evidence. This audit-only update
does not change RESULT.json or the region tables; the next measurement waits for
the compiler area-next landing SHA on main.

Setup timing: Go 0.056s, Node 0.058s, clang 0.443s, markdown ready 0.742s,
submodules 16.008s, Go build 220.999s, cache warm 221.077s, done 221.114s.
`nproc` is 5, CPU quota is 4; Go 1.27.1, Node 24.19.0, clang 20.1.8.
The base is the requested area branch: origin/main did not contain ed6e2975.
Upstream TypeScript is pinned to 050880ce59e30b356b686bd3144efe24f875ebc8.

This is an exact subtraction of recorded boundaries with independently catalogued
unit coverage, measured on a checker-rejected program. Successful statements
visited before an enclosing construct fails have no separate success spans in the
existing ledger; their exposure is not invented. The global refusal-syntax scan
is not successful lowering coverage. Generic specialization, final module order,
ownership and backend passes remain outside the measurement. No production source
or oracle fixture changed; counts.md was not changed. No whole packages or full gate
were run to confirm this unit.
