Measured the whole adapted tsc entry closure: 81 source files, including two outside src/compiler.
Compiler base 89ac4a8c; hidden measurement f1502d13; no compiler feature merges or source edits.
Latent: 1,473 NotYet and 5,164 Refused; outside compiler: 5 NotYet and 2 Refused, zero checker diagnostics.
Hidden: 5,430,761/10,010,532 bytes; outside compiler: 564/712 bytes; independent byte-mask audit agrees.
Two-file known-answer fixture, both outside-root-drop mutants, transformer tests and no-output guard mutants pass; no native or full-gate claim.

[RESULT.json](RESULT.json) retains all per-file counts, exact reason families,
hidden ranges, closure edges and source hashes. [REASONS.md](REASONS.md) gives
complete reason tables separately for compiler, outside compiler and total,
for both the first-error latent census and the full hidden-source census.

| Scope | Files | Latent NotYet | Latent Refused | Source bytes | Hidden bytes |
| --- | ---: | ---: | ---: | ---: | ---: |
| Full entry closure | 81 | 1,473 | 5,164 | 10,010,532 | 5,430,761 |
| src/compiler | 79 | 1,468 | 5,162 | 10,009,820 | 5,430,197 |
| Outside compiler | 2 | 5 | 2 | 712 | 564 |

| Outside file | Bytes | Hidden bytes | NotYet | Refused |
| --- | ---: | ---: | ---: | ---: |
| src/tsc/tsc.ts | 608 | 564 | 5 | 2 |
| src/tsc/_namespaces/ts.ts | 104 | 0 | 0 | 0 |

These are the only source files outside compiler in this pin. Contrary to the
prompt's suggested path, executeCommandLine is reached at
`src/compiler/executeCommandLine.ts`, and is included in the compiler bucket.
`closure.cjs` walks imports, re-exports, literal import types and require/import
calls; stock TypeScript 6.0.3 independently confirms the same 81-file source
program. Unresolved relative imports fail. Host modules and nonliteral require
expressions are listed in the closure artifact and do not create invented
source files. This is a static source closure, not a trace of runtime file reads.

The outside reason table is identical in both measurement modes:

| Kind | Exact reason | Sites |
| --- | --- | ---: |
| NotYet | reading ts | 5 |
| Refused | a method read as a value (setBlocking would lose its object, and this with it) | 1 |
| Refused | a method read as a value (tryEnableSourceMapsForHost would lose its object, and this with it) | 1 |

There are 324 checker diagnostics, all attributed to compiler files. The latent
ledger additionally has 3 SkippedDependency and 3 panic sites. Full mode has
15,288 NotYet, 12,275 Refused, 3 SkippedDependency, 25 panic and 2 ordinary error
sites. Two ordinary full-mode errors have no source location; they are retained
in an explicit unattributed bucket rather than assigned to either file group.
Boundaries used for hidden arithmetic are not counted as lowering reasons.

The hidden-source arithmetic is **7,543,130 blocked union bytes minus 2,112,369
independently examined bytes = 5,430,761 hidden bytes (54.250473%)**. Outside
compiler, 564 bytes are blocked, no independent exposure subtracts from those
spans, and the share is 79.213483%. Only reached source files form this denominator.
The earlier hidden report included all regular compiler-directory files, including
three JSON assets; those unreached assets are not part of this source closure.
Do not compare the two denominator conventions as a compiler improvement.

## Reproduce

Start on this topic's main base, with Node 24.19.0 and TypeScript 6.0.3:

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/closure-setup.log 2>&1
source /workspace/adamic-tools/env.sh
export NODE_PATH=/home/agent/.cache/adamic-stage3/api/node_modules
bash stage3/apply.sh /tmp/closure-adapted > /tmp/closure-apply.log 2>&1
python3 stage3/census/tsc-closure/prepare.py "$PWD" /tmp/closure-tools > /tmp/closure-prepare.log 2>&1
node stage3/census/tsc-closure/closure.cjs /tmp/closure-adapted src/tsc/tsc.ts /tmp/closure.json > /tmp/closure-graph.log 2>&1
LATENT_ROOT_MANIFEST=/tmp/closure.json LATENT_ASSERT_NO_OUTPUT=1 /tmp/closure-tools/census /tmp/closure-adapted/src/compiler /tmp/closure-latent.jsonl > /tmp/closure-latent.log 2>&1
LATENT_ROOT_MANIFEST=/tmp/closure.json LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 /tmp/closure-tools/census /tmp/closure-adapted/src/compiler /tmp/closure-full.jsonl > /tmp/closure-full.log 2>&1
```

The `src/compiler` argument remains for compatibility with the upstream driver;
`LATENT_ROOT_MANIFEST` explicitly replaces that directory's root selection,
including every reached outside source. Missing ledger rows fail coverage.
Create a temporary view containing exactly the manifest's files, preserving their
relative paths and bytes, then catalogue it with the pinned
`closure-tools/tree/stage3/census/hidden/units.cjs`. Pass that stock catalogue,
the two ledgers, manifest and pinned hidden.py to `measure.py`:

```sh
python3 stage3/census/tsc-closure/measure.py /tmp/closure.json /tmp/closure-latent.jsonl /tmp/closure-full.jsonl /tmp/closure-stock.json /tmp/closure-tools/tree/stage3/census/hidden/hidden.py /tmp/closure-result.json > /tmp/closure-result.log 2>&1
python3 stage3/census/tsc-closure/test_fixture.py /tmp/closure-tools/census /tmp/closure-tools/tree/stage3/census/hidden /tmp/closure-fixture > /tmp/closure-fixture.log 2>&1
```

`prepare.py` extracts only the measurement tools from f1502d13 into an uncommitted
main worktree and builds Go overlays against main. It uses main's actual cohere
7945d102 and typescript-go d92d9bfe pins, with canonical dependency paths. The
only driver change selects manifest roots and records them in the header.
No feature branch is merged and production compiler source is untouched.
The full hidden algorithm is reused without arithmetic edits; its denominator
is generalized by supplying stock records for the full source closure.

## Fixture, mutant and checks

The authored fixture has `src/compiler/entry.a` importing `src/outside.a`.
The runner copies them to .ts and replaces the import suffix with .js. That
actual input has 95 entry bytes and 88 outside bytes, **183 total**. The outside
function has exactly one body type error (TS2322), so both ledgers contain two
files, zero NotYet/Refused sites, and one diagnosed outside body. The hidden
answer is the half-open byte range **[33,87)**: exactly **54 hidden bytes**, all
outside compiler. No compiler bytes are hidden.

The real root-selection mutant removes `src/outside.ts` from measurement roots
while keeping its import and source file on disk. The checker's dependency
still exists, but the measurement silently emits only the entry file's row.
`test_fixture.py` separately checks each mutated ledger against the independent
closure and catches both with `closure coverage mismatch: ['src/outside.ts']`.
This exercises the omission that motivated the unit, not merely a changed total.

Passing evidence includes the eight hidden arithmetic tests (including 500 random
byte-set cases), latent continuation/body-scope/attribution audit, transformer Go
tests, and the independent full byte-mask recount. The no-output mutants are
caught with `measurement returned usable IR` and `measurement loader exposed an
output program`. The full-mode witness additionally catches first-error-only and
failed-state-retention mutants. Its pinned initial version expected supported
truthiness and failed on main's extra string/number-condition refusals; the
scratch-only expectation was adjusted to require those two refusals plus all
three var sites. Both logs are retained; no compiler behavior was changed.

Setup timing: Go build 39.969s, cache warm 40.113s, done 40.149s; nproc 5,
CPU quota 4. Raw ledgers and logs are compressed under evidence/. Provenance
includes all source and overlay hashes, binary hash and unchanged compiler diff.
No native executable, JavaScript semantic comparison, adaptation oracle/lane
rerun or full Go gate is claimed. This remains a measurement on a checker-rejected
program, with eligibility and first-error limits retained from the pinned tools.
