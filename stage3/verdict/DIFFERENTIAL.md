# Function differential proof

Built from origin/codex/stage3-verdict-harness 22483cd4d147e7ef67a1190c18d3c7dda5af2473,
merging origin/main 45487a809f89885a3fc651cd590e7dabf31362dc first. The scratch
merge is 9646781fddedc6b79699eccc68154f2ca9d502a9. Only stage3/verdict is edited.

The complete saved map holds **13,693 configurations**, each with exercised
function membership and an independent expected stdout/stderr/exit hash.
All 13,693 pass their pinned TypeScript 6.0.3 reference diagnostics and exits;
zero are deferred. There are 614 excluded inputs and 1,274 explicitly excluded
configurations, unchanged from selection.json. The map inventories 11,093
source functions in 80 source files, including generated diagnostics. Function
membership is stored as compressed bitmaps; IDs and source hashes are in its
inventory. evidence/differential/map-summary.json pins the artifact and records
its observed exercised-function count.

The coverage collection uses fresh compiler modules with preserved strict mode
and function directives, and a fixed source snapshot. The root-cwd namespace
capture was initially lost when its private root unmounted. After adding the
private capture mount, the campaign resumed: **12,135** source-identical passing
captures were revalidated, and **1,558** configurations were newly measured.
The continuation's 125.749 seconds is not a full-population duration. The
interrupted command and continuation are recorded separately. This is one
coverage campaign with a validated continuation, not one uninterrupted process.
No TypeScript source or compiled TypeScript implementation is committed.

## One-function measurement

The changed function is binder.ts's
createBinder/getStrictModeBlockScopeFunctionDeclarationMessage. The scratch
source changes only a comment inside its body; its Node CLI is rebuilt from
that checkout. This is a behavior-preserving source-body change, not a claimed
compiler optimization. The full parser inventory and the cached inventory are
identical. The changed function has exactly one configuration in the map.

`run.sh --since 050880ce59e30b356b686bd3144efe24f875ebc8 --tsc BINARY OUTPUT`
selects **one affected configuration plus eight fixed-seed samples**. All nine
pass, 13,684 remain explicitly deferred, and no coverage gaps are reported.
The final measured differential time is **3.157 seconds** including source/map
validation and all nine CLI subprocesses. evidence/differential/proof.json also
records the external wall-clock time, which includes Python startup and argument
parsing. Source binding and compiler rebuilding occur before that measurement.
The supplied binary here is a scratch Node CLI stand-in, not native Adamic tsc;
no native timing or full-scanner lowering claim is made.

## Mutants and named catches

| Mutant | Observation | Check that catches it |
| --- | --- | --- |
| Add an unexpected stdout line in the changed source function | One configuration fails; stderr and exit remain unchanged | Unchanged stdout byte comparison, exit 1 |
| Remove that function from every configuration's map, with a valid recomputed map checksum | Zero affected cases; all eight samples pass | Changed-but-uncovered function gap, exit 2 |
| Replace its source file SHA256 in the map, retaining a valid map checksum | No compiler configuration is executed | Base Git source-hash validation, exit 2 |
| Reuse a compiler module across the two isolation-fixture invocations | The second output changes | Fresh-module output/hit-set comparison in test_differential.py |
| Change a resumed source inventory | Collection refuses before using captures | Resume inventory guard |

The semantic mutant is compiled successfully and executes with the original
exit status. It is not caught by TypeScript, clang, a sanitizer or a crash.
The missing-function mutant proves that the sample's passing result cannot
silently approve a completely unmapped changed function. Its gap is flagged
independently of sampling. A partial mapping omission is not guaranteed to be
sampled; sampling is a deterministic backstop, not proof of unseen branches.

The focused tests also reject independent stdout/stderr/exit hash mutations,
a modified map checksum, a selector that ignores changed functions, new
uncovered functions and changes outside functions. Deleted functions select
old callers. A compressed-map mutation has the same gap result as the readable
association list. The source-hash mutant uses a real Git archive, and the
ignored diagnostics file is recreated with upstream's own generator.

## Commands and retained evidence

Every test, build and compiler invocation writes to logs. Source is the clean
external /workspace/type-imports/pristine checkout at the pin.

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/step34-setup.log 2>&1
source /workspace/adamic-tools/env.sh
# The instrument/instrument-server commands are in README.md.
SLICE_TYPESCRIPT=/workspace/type-imports/npm/node_modules/typescript/lib/typescript.js \
  STAGE3_VERDICT_UPSTREAM=/workspace/type-imports/pristine TSC_JOBS=4 \
  stage3/verdict/run.sh --tsc /tmp/step34-instrumented/tsc \
  --source /workspace/type-imports/pristine --record-coverage /tmp/step34-strict-map.gz \
  /tmp/step34-strict-record > /tmp/step34-strict-record.log 2>&1
# After the root-cwd capture fix, resume only revalidated successful captures.
SLICE_TYPESCRIPT=/workspace/type-imports/npm/node_modules/typescript/lib/typescript.js \
  STAGE3_VERDICT_UPSTREAM=/workspace/type-imports/pristine TSC_JOBS=4 \
  stage3/verdict/run.sh --tsc /tmp/step34-instrumented/tsc \
  --source /workspace/type-imports/pristine --record-coverage /tmp/step34-complete-map.gz \
  --resume-coverage /tmp/step34-strict-record /tmp/step34-complete-record \
  > /tmp/step34-complete-record.log 2>&1
PYTHONDONTWRITEBYTECODE=1 python3 stage3/verdict/prove_differential.py \
  /tmp/step34-complete-map.gz /workspace/type-imports/pristine \
  /tmp/step34-final-proof > /tmp/step34-final-proof.log 2>&1
STAGE3_VERDICT_UPSTREAM=/workspace/type-imports/pristine PYTHONDONTWRITEBYTECODE=1 \
  python3 -m unittest discover -s stage3/verdict -p 'test_*.py' \
  > /tmp/step34-tests-final.log 2>&1
```

The focused checks pass **45/45**, with no skips. The complete coverage report,
resume provenance, differential timings/failures and compressed command logs
are in evidence/differential. The committed coverage-map.json.gz is the default
for --since. Build provenance is created by bind_source.py; rebuilding a binary
requires a new binding, never silently reusing another source's metadata.

Setup passed: Go .017s, Node .019s, submodules .056s, markdown .058s, clang .142s,
build 32.950s, cache 33.091s, total 33.118s; nproc=5, cgroup quota=4 CPUs.
No full package or full gate ran. No new compiler/oracle fixture was added;
counts.md refreshes this standalone harness's population and focused checks.

Limits: the differential covers the existing baseline diagnostic projection,
not emitted JS, types/symbols, acceptance or tiny projects. Those suites retain
their ordinary full-verdict entry point. There is no five-second guarantee for
widely used functions: every mapped caller is selected, however many that is.
Uncovered/new functions and global edits produce explicit gaps. Opaque binaries
without builder source provenance are refused, and the source attestation does
not prove a compiler was honestly built from that tree.
