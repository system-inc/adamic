Built: a deterministic 300-case real-program corpus, exact CLI harness, native slot, and three-file day-4 project.
Base: ef3d907; TypeScript: 050880ce59e30b356b686bd3144efe24f875ebc8; prerequisite SHAs are in evidence/build-provenance.json.
Observed: Node expectations 301/301 and ordinary golden comparison 301/301; 60 clean corpus cases, 285 codes.
Mutants: one diagnostic column, stderr and exit each caught by its exact comparison; eight provenance mutants caught.
Not covered: actual native execution, full upstream/Adamic gates, and adapted-source typechecking; standard build failed, targeted Node bundle succeeded.

## What ran

All committed changes are under `stage3/drivers/tsc/`. No compiler implementation,
source adaptation, other fixture bucket, or shared fixture test was edited.
The branch starts from current `origin/main` (`ef3d907`). The named prerequisite
branches were fetched, and their stage3 pipeline/adapters extracted into
`/tmp/diagnostics-base`. `stage3/apply.sh /tmp/diagnostics-adapted` pinned and
adapted TypeScript with adapters 00, 10 and 20. The adapted checkout remains
outside Adamic. Adapter diffs are summarized in `evidence/patch-set.md`; every
adapted compiler/tsc source, standard library and measured CLI artifact has a
SHA256 in `evidence/build-provenance.json`.

`bash cloud/setup.sh > /tmp/diagnostics-setup.log 2>&1` succeeded; the environment
file is `/workspace/adamic-tools/env.sh`. Timing lines: Go 0s, clang 0s, Node 0s,
submodules 0s, build-cache warm 157s, total 157s. `nproc` printed 5; CPU quota is
4 (`cpu.max: 400000 100000`). Go is go1.27.1, clang 20.1.8, Node v24.19.0.
The complete timing output is saved in `evidence/setup.log`.

The standard `npm run build:compiler` after all adaptations **failed**, with
TS2322 in `src/compiler/tsbuildPublic.ts:787`, and TS2322, TS2345 and TS18048 in
`src/services/services.ts`. `evidence/standard-build-failed.log` retains the
exact diagnostic chains. Adapter 20 widens optional contracts and exposes
consumer errors; this unit does not repair those owners/consumers.
Even `npm run build:compiler -- --no-typecheck` still invokes the services
build for declarations and fails. The focused upstream task
`npx hereby tsc --no-typecheck > /tmp/diagnostics-tsc-bundle.log 2>&1` **succeeded**
(exit 0), producing `built/local/tsc.js` and `built/local/_tsc.js` from adapted
source in 1.5s. This is a runtime oracle bundle, not a successful adapted-source
checker gate. No Adamic checker option was changed.

An exploratory build overlapped apply; its success is not counted as evidence.
The recorded bundle was rebuilt only after apply completed. The standard-build
failure and focused-bundle success are both preserved, so the runnable CLI is
not confused with a passing source checker.

## Verification

These commands redirected complete output to logs, then those logs were read:

```sh
python3 stage3/drivers/tsc/corpus.py /tmp/diagnostics-adapted /home/agent/.cache/adamic-stage3/api/node_modules/typescript/lib/typescript.js > /tmp/diagnostics-selection.log 2>&1
TSC_RESULTS=/tmp/diagnostics-golden stage3/drivers/tsc/run.sh --record node /tmp/diagnostics-adapted/built/local/tsc.js > /tmp/diagnostics-golden.log 2>&1
TSC_RESULTS=/tmp/diagnostics-compare stage3/drivers/tsc/run.sh node /tmp/diagnostics-adapted/built/local/tsc.js > /tmp/diagnostics-compare.log 2>&1
python3 stage3/drivers/tsc/audit.py > /tmp/diagnostics-audit-final.log 2>&1
python3 stage3/drivers/tsc/mutants.py node /tmp/diagnostics-adapted/built/local/tsc.js > /tmp/diagnostics-mutants-final.log 2>&1
```

Selection printed `selected 300 from 3480; clean=60; codes=285`. The written rule
and ordered selection are in `README.md`, `SELECTION.md` and `selection.json`.
The expected diagnostic summaries are derived from upstream baselines, not from
the adapted CLI. `--record` passed **301/301**, exit 0, 75.843s, and saved the
actual bytes as goldens. Ordinary golden comparison then passed **301/301**,
exit 0, 72.626s. The tiny project's stdout is exactly:

```text
argument.ts(2,6): error TS2345: Argument of type 'string' is not assignable to parameter of type 'number'.
assignment.ts(1,14): error TS2322: Type 'string' is not assignable to type 'number'.
```

Its stderr is empty and CLI exit is 2. The third file has no diagnostic.
The two report JSONs and full run logs are committed in `evidence/`.
Raw materializations and per-case captures remain in the named scratch result
directories; their matching bytes are committed as per-case goldens.

The audit passed source/baseline hashes, the 300-case/60-clean population,
header options, diagnostic codes, upstream summary derivation and all golden
streams/exits. An independent re-selection using the stock parser returned the
same 300 ordered programs and 3,480 candidates. Shell syntax, Python parsing and
`git diff --check` were also checked. No Go packages were changed, so no Go test
run or full uncached integration gate was used for this harness-only unit.

## Mutants and their catches

Each mutant changes one copied artifact; reviewed goldens stay intact. The three
CLI probes rerun the real tsc against the tiny project, using the same comparison
function as all corpus cases. A mutant killed by an unrelated error does not count.

| Mutant | Dedicated catch | Result |
|---|---|---|
| Golden stdout first column: `(2,6)` becomes `(2,7)` | stdout bytes | `FAIL tiny: stdout`, harness exit 1 |
| Golden stderr gains `changed stderr` | stderr bytes | `FAIL tiny: stderr`, harness exit 1 |
| Golden CLI exit becomes 0 instead of 2 | exit bytes | `FAIL tiny: exit`, harness exit 1 |
| Input source gains a newline | source SHA256 | caught |
| Upstream baseline gains a newline | baseline SHA256 | caught |
| Derived expectation changes `error TS` to `error XX` | baseline summary derivation | caught |
| Expected empty stderr gains a line | expected stderr | caught |
| Clean expected exit becomes 2 | expected exit | caught |
| Selection loses one entry | case population | caught |
| Manifest option differs from the source header | header options | caught |
| Manifest adds TS99999 | baseline codes | caught |

`evidence/mutants.log` and the three dedicated diff logs retain these catches.
The `NATIVE_TSC` slot also passed a smoke run using a transparent executable
that forwards its argv to Node. It ran the tiny project twice and wrote a
separate `native/report.json`. That establishes slot plumbing only; it provides
no evidence of a native compiler.

## Observations and limits

Early corpus attempts exposed genuine harness-domain differences: the API
baseline combines syntax and semantic diagnostics whereas the CLI suppresses
semantic checking after syntax errors; TS18027 is emit-only; `lib` headers need
array parsing; semantic noEmit CLI errors exit 2. Those observations informed
the explicit scope/option rules, not changed expected messages or positions.
The final rule selects without running the measured compiler or using match
results. All selected input and reference bytes still match their upstream
hashes. Actual output receives no path, newline, location or text normalization.

A native binary was not supplied, so native execution is intentionally pending.
The full upstream suite and Adamic's native/ownership gates were not run.
The optional-adapted source still fails its standard checker build, as recorded
above. JSX, option variants, virtual/multi-file corpus cases, emission,
watch/incremental modes and Windows formatting are outside this bounded oracle.
No native correctness or silent-miscompile conclusion is inferred from Node
agreement. The harness is ready to measure native tsc when its path is provided.
