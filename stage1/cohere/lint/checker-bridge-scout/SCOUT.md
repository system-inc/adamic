# Step 27: measure the production checker crossing before batching files

Branch: `scout/27-checker-bridge`. Base: area/stage1-lint
`9156bf5c579a44d687c9955d13e44f9ad8bbb6f8`; cohere
`7945d102a6c18dd36adf9114a758ce646e8b2359`; its TypeScript checker
`d92d9bfee114c80be2c375d72edae966176e3a4f`.

## Five findings

1. All 23 supplied public pins were shallow-fetched. Their trees contain 124,348 non-declaration `.ts` paths; the measured sample is 23 files / 316,107 bytes, not Kirk's complete Mac corpus.
2. The package gate holds 84 isolated cases and 23 shared-project rows to unchanged Go rules, native Adamic and source Node: 39 findings with exact proposed fixes and 194 logical pilot asks. Eight additional short reductions also agree.
3. The production path is an in-process C archive. Cold, repeated, per-file and single-program measurements are below; profiling is separate from throughput.
4. The bridge forces one checker and serializes all handles. AST selectors and semantic checker work are cached; `Checker.ask` and serialized fact answers are not memoized.
5. `askFile` refuses an undeclared supplied read kind, but raw `ask` has no corresponding permission check. Question classification, batch errors, ownership and N-checker identities need explicit contracts.

## Scope and evidence

Read before selecting the implementation: `docs/0.1.md`, `docs/memory.md`,
`docs/escape-hatches.md`, `docs/lint-registration.md`, `docs/stage1-progress.md`,
`bridge/tsgo/README.md`, `bridge/tsgo/facts.md`, the lint README/GAPS/CHECKER_REPORT,
typeaware README/GAPS/PROFILE_REPORT/VOLUME_PROFILE_REPORT and parser GAPS. Their historical refusal reports are not
substitutes for checking today's source. `checker_bridge.a:5` now wraps scalar
builtin results, from f0360c73; the reported scalar/envelope compilation blocker
in CHECKER_REPORT no longer applies. Native compilation and the gate prove that.
The scalar adapter still panics on C errors; wrapping a successful return in Ok
has not implemented errors-as-values.

The facts writer's branch was inspected without merging:
`lint-checker/facts` at `d845dccde413c89643293e808626344d12e3f023`.
Its `stage1/cohere/lint/CHECKER_FACTS.md:395` prioritizes complete symbols and
provenance; `:479` routes same-file selectors through guarded `askFile` and
leaves alias, module and signature facets separate. Its
`bridge/tsgo/checker/declaration_facts.md:20` retains exact selectors and
same-file suffixes, declares default-library consumers separately and promises
no foreign-file parsing in the harness. This scout neither adopts that branch's
facts nor creates a second writer or question registry.

The task records #45rq89s and #mpg3abq are not in these trees. Repository text
search and GitHub issue searches returned no matching notes. The mapping below
is evidence from code, not an assertion about the unavailable notes. Their
location and any additional requirements remain open to @system_adamic.

### Where the step bites

| Evidence | Behavior / count |
| --- | --- |
| `cohere/internal/lint/rules/typescript/no_unnecessary_boolean_literal_compare.go:139`, `:163`, `:305`, `:322` | The pilot needs a checker, declares ReadsCompilerOptions, checks strict null options and asks the operand type / generic constraint. Its production Run is unmodified. Defaults permit nullable comparisons. |
| `stage1/cohere/lint/rules/no-unnecessary-boolean-literal-compare/rule.a:15`, `:41` | One options askFile per selected file, then one type-shape ask per equality/inequality having a boolean literal. Right-hand literal wins when both sides are literals; parentheses are unwrapped. Judgment, messages and edits stay in Adamic. |
| `stage1/cohere/lint/checker.a:64`, `:71`, `:75`, `:88`, `:98` | The declared-kind guard precedes askFile delegation; ask computes exact UTF-8 spans. Live calls do not build transcript keys unless recording. Replay checks key, outcome, payload and complete consumption. No answer cache lookup exists. |
| `stage1/cohere/lint/checker_bridge.a:5`, `:8`, `:11` | Actual production scalar ABI adapter, imported unchanged by pilot.ts. A C error panics before the envelope wrapper returns. |
| `internal/native/runtime/tsgo.c:127` | Three input views are allocated; the C call returns an owned buffer; the adapter copies/decodes into an Adamic string and frees temporary views and the C output. |
| `bridge/tsgo/archive/boundary.c:13`, `:63`; `bridge/tsgo/archive/main.go:110` | C copies three views again, cgo copies their text into Go, the global program mutex serializes every inspect and the output is a C allocation. Successful nonempty answers incur seven C malloc/free pairs before counting Adamic string/object allocation: 3 adapter views, 3 boundary copies, 1 output. |
| `bridge/tsgo/checker/facts.go:61`, `:194`, `:199` | Validate exact file/span/kind; acquire the file's checker for each question; compute and frame facts. Even options acquires a checker. |
| Non-test Go rule-source census | 198 `NeedsTypeChecker: true` sites in 198 files. This is a declaration-site census, not 198 verified ports. |
| Integrated lint rule-source census | 7 textual `.ask`/`.askFile` sites in 4 files, spanning 3 typed rules. One site is a wrapper call in no-redundant-type-constituents. The separate sixteen-rule typeaware suite is not counted as this harness's coverage. |

Node runs the unchanged Parser, Checker, RuleContext and pilot Rule, with native
facts replayed. It has no live checker bridge. Every recorded selector, order,
question and payload is checked, and both missing and unused entries fail. The
independent Go oracle opens its own cohere Graph and invokes the unchanged
production rule through `linter.LintFile`, including ProgramView and FileCache.
The C library returns compiler facts, never a lint verdict. Complete comparison
includes rule name, message ID/text, finding byte bounds and the automatic edit's
byte bounds/text; both clients assert the pilot's one-fix/no-suggestion shape.
Formatting and the converging edit engine are not implemented by this driver;
this pilot proposes edits. No full CLI formatting claim is made.

### Cache and program/checker map

* `cohere/internal/lint/linter/linter.go:48` creates one FileCache shared by the file's rules; `rule/rule.go:177` caches derived work by a typed key. This is distinct from compiler semantic caches and persistent findings caches.
* `bridge/tsgo/checker/facts.go:55` indexes immutable AST ranges by source-file identity, both byte bounds and kind. SourceFile metadata bypasses the descendant index at `:75`. `facts.go:121` interns type pointers into program-scoped safe integer IDs, retaining the reverse table. These tables do not cache serialized answers.
* `cohere/TypeScript/tsc/internal/checker/checker.go:687`, `:5616`, `:7728`, `:32717` show node-link/value-symbol resolved-type caches behind GetTypeAtLocation. Repeated bridge calls still perform input copies, lookup, lease, graph serialization, output allocation and decode.
* `bridge/tsgo/checker/program.go:51`, `:101`, `:112` creates a program with cached FS/bundled libraries, forces SingleThreaded and initializes its pool. Types stay lazy. Changing tsconfig's checker count or GOMAXPROCS does not make this bridge allocate N checkers.
* The compiler pool at `cohere/TypeScript/tsc/internal/compiler/checkerpool.go:309` normally defaults to four, honors Checkers when not single-threaded, and clamps to `[1,min(fileCount,256)]`. `:351` associates each file with its owning checker and locks it exclusively. The same checker must be passed into nested operations: acquisitions are not reentrant.
* Go cohere `internal/types/program/program.go:547`, `:701`, `:770`, `:828` sets its worker count, leases the file's owner, and reports the actual single-threaded/clamped worker count. Production defaults track GOMAXPROCS up to the ceiling; the scout Go oracle deliberately requests one checker to compare the bridge's policy. Full cohere's N-worker throughput is not simulated here.
* The normal harness `stage1/cohere/lint/main.ts:199`, `:221`, `:238` creates one program per manifest, creates file views per row, and releases once. `pilot.ts` supports the same single-program lifetime with --manifest. The per-file experiment deliberately recreates that complete program to isolate first-file costs; its loads must not be summed as the production lifecycle.
* Creating N bridge handles duplicates program/checker/cache state and still hits the global mutex at `bridge/tsgo/archive/main.go:21`. It does not provide parallel checker throughput. N checkers in one future program also require type identities to retain their owning checker: Go cohere's `program.go:692` explicitly forbids mixing types from different checkers.

A denser historical production suite is documented in
`stage1/cohere/typeaware/VOLUME_PROFILE_REPORT.md:65`: sixteen default rules on
the 77-file compiler corpus made 854,525 queries and returned 144,896,637 UTF-8
fact bytes. Its post-change profiled C-call minus Go-body residual was 0.353 µs
per query; its decoder interval was 4.729036 s. This is prior evidence at the
older cohere/typechecker pins stated in the report, not a rerun at this scout's
pin. The six-rule PROFILE_REPORT likewise attributes its major costs to fact
rendering/decoding and allocations. Neither prior result establishes a batch
speedup, and neither is substituted for this scout's public-sample measurements.

### Read contract

Go cohere's `internal/lint/rule/program.go:185` checks every Program method
against the active rule's declarations. Default-library methods require
ReadsDefaultLibrary at `:210`/`:215`; foreign lookups, source enumeration and FS
require ReadsOtherFiles at `:238`/`:253`/`:258`/`:263`. Resolving a foreign file's
imports adds ReadsOtherFiles at `:222`.

Cache policy at `cohere/internal/types/program/lint_cache.go:446`/`:474` keys
options through config, libraries through the binary and imported types through
the type fingerprint. ReadsOtherFiles is uncacheable unless the rule supplies
its ProgramFingerprint; it must not be made cacheable merely because facts were
batched. ReadsDesignSystem and type reads have their separate exclusion.

The port registry validates names and rejects ProgramReads on untyped
descriptors at `stage1/cohere/lint/registry/registry.go:154`/`:164`.
Checker.askFile verifies one supplied kind against `enter`'s descriptor list.
Native and Node controls prove ReadsOtherFiles and ReadsDefaultLibrary refuse
when absent. But `ask(index, 'options')` has no equivalent check, and the facade
does not map question text to required categories. The ABI receives no rule or
read mask. `bridge/tsgo/checker/metadata.go:23` reads default-library provenance
directly from Compiler, outside Go's ProgramView. Even an askFile caller could
supply a category unrelated to its text. These are source-visible enforcement
limits, not permission to exploit the raw route or to weaken declarations.

The upcoming facts lane's declared same-file selectors and provenance flags
must retain those requirements in a batch. Asking a foreign fact is different
from allowing Adamic to open or parse foreign source. This scout does neither.

## Corpus and fixtures

Kirk clarified that the quiet hundred lives on his Mac, not in a repository.
HTTPS GitHub access worked; all 23 supplied pins were shallow-fetched with
blob filtering and no installs or repository scripts. SSH was unavailable and
HTTPS used the configured proxy. The inferred public owner names for the short
labels are recorded in `testdata/public-pins.json` beside each supplied SHA.
Their trees were enumerated, not all linted. `fetch_public.py` and
`sample_public.py` reproduce the sample: first boolean-literal comparison in the
first 40 eligible lexically ordered non-declaration TS paths, excluding common
test/fixture paths and limiting candidates to 40,000 bytes; otherwise the first
eligible small file. TypeScript uses compiler/core.ts explicitly, with its moved
fixture path at 50d70a3f. Both TypeScript core files have identical bytes.

The controlled sample config is exactly strict=true, target=ESNext,
noResolve=true, roots=all 23 sampled paths. It does not claim original project
configurations or installed dependency types. Missing-import/type errors are
allowed by the production bridge and oracle; parse diagnostics fail. All files
remain unchanged during a run. Consequently these finding counts and costs are
not a survey of all findings in the 23 repositories or the complete quiet hundred.

| Fixture population | Cases / assertions |
| --- | --- |
| Go cohere's own pilot tests | 53 unique literal sourceText values with literal empty configuration; same TrimSpace+LF convention as RunTyped. Exact source, SHA-256 and file:line are in `testdata/cohere-fixtures.json`. Nondefault-option assertions are outside this scout's fixed-default scope. |
| Existing pilot witness | `stage1/cohere/lint/rules/no-unnecessary-boolean-literal-compare/testdata/witness.ts.txt`; 1 case. |
| Reductions in package gate | Constrained generic, nullable boolean, astral comment with nested parentheses; 3 cases. |
| TypeScript 6.0.3 | Supplied pin 050880ce59e30b356b686bd3144efe24f875ebc8, exact src/compiler/core.ts; 1 case. |
| TypeScript tests at cohere's pin | Actual directory is `cohere/TypeScript/tsc/testdata/tests/cases`, not the suggested top-level path. booleanLiteralTypes2.ts, booleanAssignment.ts and booleanPropertyAccess.ts; 3 cases, with original CRLF retained in `testdata/typescript-fixtures.json`. |
| Public sample | 23 isolated cases plus the same 23 files in one controlled shared project. Every path/SHA/size is in `testdata/public-fixtures.json`. |
| Whole-manifest lifetime | One native program reused across all 23 files agrees with Go's one-program run; 38 asks, 2 findings. |

The 84 isolated cases report 37 findings and make 156 logical asks. The 23
shared-project rows add 2 findings and 38 asks. All 38 shared-project keys are
distinct; there is no same-selector duplicate to eliminate in this pilot.
Seven files ask operand types; the other sixteen ask only options. The sample
has 23 options asks and 15 type-shape asks. Native timing counts must agree with
the live port's recorded sequence; the Go AST planner cannot silently omit an
ask or count a generated manifest as a real rule run.

### Short inputs and hard cases

`evidence/minimal.json` holds independent Go/native/Node outputs and actual
native transcripts for these additional verified reductions. Sources are inputs
to the checker, not Adamic implementation source; the no-any/subset contract
still applies to the drivers themselves.

| Input, LF implicit unless empty | Oracle and Node/Native behavior |
| --- | --- |
| empty string | 0 findings, 1 options ask; exact empty SourceFile range 0:0 is accepted. |
| `true===false` | 1 finding; right literal is selected, fix is `!true`; 2 asks. |
| `let b=true;b===true` | 1 finding and exact `b` edit; 2 asks. Querying the comparison result would hide operand distinctions in other cases. |
| `function f<T>(b:T){return b===true}` | 0 findings; unconstrained parameter is declined; 2 asks. |
| `function f<T extends boolean>(b:T){return b===true}` | 1 finding with `b` edit; constraint, not raw type-parameter flags, decides; 2 asks. |
| `declare let b:boolean\|undefined;b===true` | 0 findings under the production nullable defaults; 2 asks. |
| `//🌍` then `let b=true;b===false` | 1 finding with `!b` edit; UTF-16 parser spans map to UTF-8 byte bounds; 2 asks. |
| `1===true` | Semantic type error is allowed, 0 lint findings; 2 asks. |

Boundary hard cases remain explicit: exact trivia-inclusive selectors cannot be
replaced by nearest-token positions; per-program type IDs cannot be reused after
release; root option asks must not pay to index descendants. Name/assignability
questions can depend on IDs returned by earlier answers, so one eager batch
cannot indiscriminately prefetch every future question. An unsupported question
returns an error in Go but currently panics at the native scalar adapter;
errors-as-values is not claimed. The production same-file `askFile` grammar on
the facts writer branch is not silently widened to foreign paths.

### Controls that can fail

The package uses t.Parallel in its new top-level test and supplies every corpus
input; no test skips. Required environment variables fail loudly if absent.
Native clients and the archive C boundary are built with ASan/UBSan; leak
checking is enabled. This covers C/Adamic allocations, not Go heap reclamation.

* A wrong expected fact makes the native byte comparison panic, rather than passing a count-only test.
* A repair-output mutant builds/runs on native and Node and disagrees with Go's full fix bytes while retaining ordinary source/rule execution.
* Missing, extra and wrong-question transcripts fail the actual Node Checker replay contract.
* Undeclared ReadsOtherFiles and ReadsDefaultLibrary calls fail with the named guard and Adamic exit 70 on both native and Node.
* Every recorded native ask is compared to the planner transcript; every timing counter is checked against the same actual ask count times repetitions.

No mutant edits the bridge, parser, harness, compiler, runtime or another rule.
The output mutant is in this package's presentation client, not a claim that a
production rule-predicate mutant was added.

## Measurements

The controlled project loads 122 source files: 23 roots plus bundled libraries. Cold and repeated per-file runs each have 69 records (23 files × 3 rounds). Profiling has 23 separate records. Source hashes and all counters are retained in evidence/.

| Whole program, median of 3 release runs | Native pilot | Independent Go rule |
| --- | ---: | ---: |
| One pass, process | 139.623 ms | 88.433 ms |
| Create/pool initialization | 76.634 ms | not separately instrumented |
| 38 question adapters, including lazy Go work | 3.447 ms | not separately instrumented |
| After-create run | 58.140 ms | not separately instrumented |

The 100-pass whole-program native run makes 3,800 asks; its Go comparison executes one pass to check the first-pass output. Their repeated process times are not comparable throughput figures.

| Per-file experiment, all 23 files × 3 rounds | Raw native ABI | Production Checker + rule |
| --- | ---: | ---: |
| Median complete-program create | 78.737 ms | 79.549 ms |
| Median cold question time per file (µs) | 16.946 | 16.785 |
| Median 100-pass question time / file / pass (µs) | 1.757 | 5.484 |

Raw ABI: complete first-pass subtraction yields **2.198 µs per remaining question** (aggregate measured difference, separate cold/repeated runs). This includes Go work and scheduling. It is not a pure crossing latency or a proposed batch speedup.

Production pilot: complete first-pass subtraction yields **4.191 µs per remaining question** (aggregate measured difference, separate cold/repeated runs). This includes Go work and scheduling. It is not a pure crossing latency or a proposed batch speedup.

| Separate profiled run, weighted ns/question | Raw native ABI | Production pilot |
| --- | ---: | ---: |
| Input conversion | 152.8 | 165.0 |
| Output decode/free | 115.1 | 137.9 |
| C-call minus timed Go inspection | 189.5 | 6332.8 |
| Timed Go inspection | 2942.9 | 5938.3 |

The larger production residual can include registry-lock/scheduling/Go-runtime effects while the surrounding Adamic rule allocates and executes. It must not be attributed entirely to cgo. Program creation and surrounding rule work are visible separately.

| Public sample (full pins/paths in testdata/public-fixtures.json) | Asks | Findings | Median cold adapter time/file (µs) | Remaining-question difference (µs/q) |
| --- | ---: | ---: | ---: | ---: |
| `actualbudget__actual-9732a446/packages/ci-actions/src/news-feed/parse.ts` | 4 | 0 | 646.877 | 3.733 |
| `angular__angular-c0dc8c4b/adev/shared-docs/components/algolia-icon/algolia-icon.component.ts` | 1 | 0 | 14.973 | 1.503 |
| `babel__babel-f67453d5/Makefile.source.ts` | 1 | 0 | 16.095 | 6.942 |
| `backstage__backstage-53293124/.storybook/main.ts` | 1 | 0 | 15.484 | 1.792 |
| `calcom__cal.diy-54343aa6/apps/api/v2/src/env.ts` | 2 | 0 | 335.316 | 3.775 |
| `date-fns__date-fns-717ce0a8/pkgs/core/examples/cts/test.ts` | 1 | 0 | 14.312 | 1.446 |
| `excalidraw__excalidraw-ed10ac7d/packages/common/src/editorInterface.ts` | 3 | 0 | 277.138 | 3.413 |
| `grafana__grafana-c7a7b797/apps/alerting/rules/plugin/src/generated/alertrule/v0alpha1/alertrule_object_gen.ts` | 1 | 0 | 15.924 | 1.482 |
| `microsoft__playwright-2a8ba77a/packages/injected/src/ariaSnapshot.ts` | 4 | 0 | 632.035 | 4.343 |
| `n8n-io__n8n-e77e30c7/packages/@n8n/agents/src/runtime/loop/stream-sink.ts` | 2 | 0 | 1330.038 | 1.264 |
| `nestjs__nest-35142c3e/integration/_support/register-local-packages.ts` | 1 | 0 | 14.351 | 1.707 |
| `outline__outline-478e8121/app/actions/index.ts` | 4 | 2 | 880.450 | 5.420 |
| `prisma__prisma-c882b033/apps/lsp-playground/src/bridge.ts` | 1 | 0 | 15.313 | 2.000 |
| `shadcn-ui__ui-a2e305c2/apps/v4/app/(app)/(create)/hooks/use-action-menu.ts` | 1 | 0 | 15.844 | 3.308 |
| `supabase__supabase-87681812/apps/design-system/config/docs.ts` | 1 | 0 | 15.513 | 5.731 |
| `TanStack__query-eaa75f4f/examples/angular/auto-refetching/src/app/app.component.ts` | 1 | 0 | 15.393 | 1.780 |
| `tldraw__tldraw-db1c86ea/apps/analytics-worker/src/worker.ts` | 1 | 0 | 14.422 | 1.633 |
| `trpc__trpc-d756e591/examples/.experimental/next-app-dir/next.config.ts` | 1 | 0 | 14.562 | 1.707 |
| `twentyhq__twenty-ddd16625/packages/create-twenty-app/src/cli.ts` | 1 | 0 | 15.092 | 2.051 |
| `typeorm__typeorm-c64a1f05/docs/docusaurus.config.ts` | 1 | 0 | 15.333 | 6.851 |
| `microsoft__TypeScript-50d70a3f/tsc/testdata/fixtures/compiler/core.ts` | 1 | 0 | 36.085 | 30.469 |
| `microsoft__TypeScript-050880ce/src/compiler/core.ts` | 1 | 0 | 40.030 | 13.798 |
| `vuejs__core-4ab865a8/packages/compiler-core/src/compat/compatConfig.ts` | 3 | 0 | 282.897 | 2.893 |


The normal native archive uses `go build -buildmode=c-archive` with default Go
optimization. Native clients use clang 20.1.8, -std=c11 -O2 -ffp-contract=off
-fno-optimize-sibling-calls and explicit --tsgo linking. Sanitized validation
uses -O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all for both clients
and the archive C boundary; those binaries are not used for timing. Platform:
Linux amd64, AMD EPYC 9V74, 5 visible CPUs, cgroup quota 4 CPUs. All measurement
processes inherit GOMAXPROCS=1; the bridge and Go oracle each use one checker.
Node validation uses Node 24.19.0; Go is 1.27.1.

Per-file request timing uses the same ordered asks with 1 or 100 passes on one
live program. The direct Go bridge is measured alongside the raw native ABI
client; a separate production pilot measures the unchanged Checker and rule.
Ordinary release measurements enable ADAMIC_TSGO_TIMING=1; separate profiles also enable ADAMIC_TSGO_PROFILE. `query_ns` includes all C input/call/output work and lazy checker work, but not
rule traversal or printing. `run_ns` includes post-create rule/client work and
printing until release begins. `load_ns` includes program creation and checker
pool initialization, not a full semantic check. The first query is options,
so first_query_ns alone is not the first semantic-type cost. Repeated-minus-cold
comparisons subtract the complete first pass, not just that metadata query.
Whole-manifest native/Go runs are interleaved in alternating order. Per-file
fact generation necessarily precedes its native check; that order is disclosed.

Profiling adds C input/call/output clocks and a Go inspection interval.
`call_ns - inspect_ns` is a residual containing the boundary copies/cgo entry,
registry-lock waiting and scheduling outside the timed Go body. It is not a pure
cgo toll. Go's timed body includes its input copies, fact serialization and C
output allocation. CPU-profile startup/shutdown also affects process timing;
profiled process time must not be used as ordinary throughput. No measured
improvement from batching is claimed because no batch ABI has been implemented.

## Proposed design, pending the questions below

The first implementation here needs no new language or ABI ruling: driver.ts
calls the existing scalar ABI, pilot.ts calls the existing production Checker
and rule, and the Go driver records counts/costs. All executable Adamic source
is valid TypeScript, uses existing constructs, and builds natively. No language
gap is hidden behind a cast, new syntax, parser change or alternate implementation.

The batching candidate remains a design, not a promised API:

1. Keep one program for the manifest. For the pilot, collect independent exact selectors in preorder: options first, then the same operand selectors the live rule asks. Retain right-literal priority and trivia/UTF-8 spans. Existing transcript equality proves that set and order before a batch implementation is considered.
2. Represent requests and responses as ordinary readonly TypeScript records/arrays, with a file-level outer request and per-entry selector, question, originating rule and declared read requirements. Use discriminated Value/Error responses only if the owner approves that contract. No async, promises, lazy syntax, exceptions, any, expando objects or collector is proposed.
3. Preserve logical request order and per-file checker association. A future C batch could copy the path once, validate every exact selector, lease that file's checker once, and serialize ordered owned responses. It must not hold a checker lease and call today's Inspect inside it: checkerpool.go:351 is not reentrant and Inspect acquires the same lease. Factoring a borrowed-checker helper belongs to the sole facts writer.
4. Make existing synchronous asks consume validated prefetched answers at their original logical call sites. Recording/replay must describe those logical asks, not silently consume a different eager physical sequence. Repeated/dynamic questions and dependent name/assignability questions need later explicit waves; no request may be invented merely to pad a benchmark.
5. Copy returned buffers into owned Adamic values before freeing the batch's C storage. Preserve ordered duplicates, owned answer survival after program release, and borrowed ID invalidation. No Go pointer or external arena view becomes an Adamic reference. Each output/error buffer must have one explicit freeing owner.
6. Start with within-run file batching; do not introduce a persistent answer cache. Persistent findings and fact reuse would require the existing config/library/type-closure/program-fingerprint contracts, not just source path/span. The sample has no duplicate selector keys, and sixteen files have only one ask, so benefits must be demonstrated on broader typed rules rather than inferred from those files.

### Open questions for @system_adamic

1. **Read authority (#mpg3abq):** Should every raw and batched question have an owner-registered required read mask, including default-library provenance, with all required bits enforced independently of the caller's label? May raw ask continue to request program metadata? The current guard checks only the supplied kind.
2. **Existing synchronous interface:** May rules gain a planning/prefetch phase with readonly answer tables, or must batching be an explicit new library call? How should logical replay preserve missing/extra-ask checks when physical requests are eager? This is an interface/ordering question, not a proposal for new syntax.
3. **Error contract:** Must a batch preserve the current first-error native panic, or return per-entry Errors? If values are approved, what happens to successful entries after a failure, and which layer reports it? Existing scalar Ok wrappers do not return C errors as values.
4. **Ownership/library surface:** Is an owned array of ordinary readonly responses the intended builtin result, and may the ABI use one temporary C arena copied before return? Who frees partial outputs and error buffers? No borrowed Go pointer or collector in Adamic is proposed.
5. **N checkers (#45rq89s):** Does this stage keep the bridge's forced one checker, or expose N? If N is intended, how are files and type/symbol IDs bound to their owner, and how should the global all-handle lock be replaced without mixing checker caches?
6. **IDs, order and dependencies:** Is deterministic first-seen ID numbering/replay framing part of the observable contract? May a batch deduplicate identical requests or reorder independent ones? How are questions requiring a prior type ID grouped without inventing symbolic-ID semantics?
7. **Reuse/invalidation:** Is reuse restricted to one immutable program lifetime initially? If persistent answers are desired, which approved fingerprint covers each question's library, config, imported/global and other-file reads, including ProgramFingerprint exceptions?
8. **Missing records:** Where are the authoritative #45rq89s and #mpg3abq notes, and do they add constraints beyond the code cited here? Their contents have not been guessed.

No question is answered by this scout. Proposed representations do not authorize
new compiler/builtin semantics; the measurement clients need none of them.

### Owner boundaries and stopping point

No shared source was changed. If batching is approved, the required owner work
would be `stage1/cohere/lint/checker.a` and the shared harness for planning,
permissions, answer consumption and replay; `bridge/tsgo/tsgo.h` and
`archive/boundary.c`/`archive/main.go` for batch ownership/status/locking; the
facts writer's `bridge/tsgo/checker/facts.go` for lease-once evaluation and
question classification; and the library/compiler/native adapter for an approved
new builtin/result. This package stops before all of those edits.

## Reproduction and next scout

From repository root, after sourcing the existing tool environment:

```sh
bash stage1/cohere/lint/checker-bridge-scout/verify.sh /absolute/new/scratch/directory
```

verify.sh shallow-fetches the exact public pins, materializes source fixtures,
builds Go/compiler/archive/release/sanitized clients, builds the independent
oracle via an overlay, runs the entire package with every input supplied, checks
gofmt/vet, then measures cold/repeated per-file and whole-program runs and
separate profiles. It runs no scripts from a fetched repository. Evidence JSON
and gate logs are retained in this package; generated C, archives, executable
binaries and fetched repositories stay in scratch.

The next scout should take the public pin/fixture lists, actual ask transcripts
and the per-file/whole-run baseline to the sole facts writer. First resolve the
read/error/identity questions and obtain the authoritative notes. Then measure
Kirk's complete Mac corpus or explicitly configured full public projects and
rules with dependent/dense questions. After a ruling, the writer can implement
a file batch with lease/permission/error/ownership mutants; this scout package
is the independent scalar and production baseline for that comparison. Do not
land a batching rewrite or change shared files on the strength of this small
pilot's numbers.
