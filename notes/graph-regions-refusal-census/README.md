# Graph-region recorded-outcome census

Candidate: `749e668e1d7aa32b0ed716b3f17ab6cba94b6f13`.
Exact main baseline: `9a941a49cb2c20ae90130348e5271c0f8cb7f08a`.
Linux, clang 20.1.8, Go 1.27.1, Node 24.19.0, WASI SDK 27.
`internal/fuzz` belongs to another change and is not modified.

The complete retained census checked **771** file-backed promises. Exactly **63**
change kind, all from cycle-capable Refused to Compiles; there are no other kind
transitions. Seven needed live record corrections; 56 were already admitted by
the batch's oracle guards. All 63 pass sanitizers, both backends and counted heap
balance.

## Harness inventory

- `stage3/fixtures`: fifteen live `*/status.json` manifests, 198 cases.
  The independent lowering census checked every one of the 46 Refused and
  38 NotYet rows. Only host 14 and objects 29 became Compiles. The historical
  nested-function status snapshots and real/bytes-revealed report are evidence,
  not live manifests. The whole package also checks the directory/test inventory.
- `internal/acheck`: `known.txt` plus `// a-check:` first-line promises for
  every program outside a Go package; the whole promise test checks additions,
  changed reasons, stale known entries and vanished paths.
- `internal/oracle`: the fixture table (lowered/not-yet/checked), shared Refused
  handling, dedicated refusal tables and inline diagnostics; review `refuse/`,
  `agree/`, `.pending`, `.intended`, `.expected.json`, `.what` and `.fix` sidecars;
  async review, concurrency and moves refusal families; getter/refusal censuses.
  All 231 file-backed pending/refused sidecar programs were also compared using
  independently built main and candidate compilers, so ordinary pending skips
  cannot conceal an admission. Only memory_examples/refused/{closure,tree}.a
  changed in that first selection; batch 8 already passes those through checkGraphAdmission.
  Expanding to every refuse/refused/notyet-named directory and fixture checks
  488 programs, finding 54 fresh_refused admissions plus those two memory
  examples, all already covered. See already-covered-admissions.md for every
  program, proving test and Node output.
- `internal/lower`: inline Refused/NotYet assertions and named refusal probes,
  including nullable cycles, record cycles and Promise cycles. Run whole.
- `internal/native`: OptionalWriteMissingSlot and PrototypeDebugSiteRefusals /
  PrototypeMixedCountedCycleProbe pin lowering outcomes. Runtime refusal tests
  (graph concurrency, JSON cycles, record __proto__) assert runtime contracts,
  rather than recorded front-end outcomes.
- `internal/conformance`: `// Spec:` and `// Expect:` first-line contracts,
  judged by TestEveryCaseKeepsTheSpecsPromise. Run whole.
- `stage3/meter`: compiler selection, entry and report tests use synthetic
  compilers/findings; 42 Python tests pass, four prerequisite-dependent real
  probes skip. They do not pin a current compiler cycle refusal.
- The standalone stage3 fixture recorders/checkers consume these same manifests.
  The objects audit also consumes native.json, now regenerated and guarded.
  Original source-ledger audits and real/bytes-revealed/check.py are frozen
  provenance checks (the latter explicitly requires compiler ed6e2975), not
  current-compiler verdict gates. Historical runs under stage3/census and
  stage3/meter/runs are measurements, not evergreen expected counts.
- typeaware request refusals and lint helper/comment parser refusals describe
  runtime input validation, not lower.Refused/NotYet; they are excluded from
  the compiler-outcome list. The lint regex dynamic-pattern gap is included.
- `cmd/adamic-meter`: diagnostic-kind corpus, including refused.ts and notyet.ts;
  TestFixtureCorpusCountsEveryDiagnosticKind and TestJSONIsMachineReadable.
- `cmd/adamic-test262`: classifier and mini-runner corpus expectations, adapter
  refusals and runtime comparisons. Reports under evidence/ are historical
  measurements, not a checked-in list of expected compiler verdicts.
- `cmd/adamic-stage1-progress`: gap/evidence inventory, not compiler verdicts.
- `internal/flow`: excluded constructor/getter refusal fixtures point to the
  oracle's active refusal assertions; they are not separate expected verdicts.
- `internal/fuzz`: moves and captured-let refusal properties; owned by the other
  worker and deliberately neither changed nor rerun here. The oracle moves and
  concurrency families check the corresponding recorded fixture refusals.
- `internal/dry`, buildcache tests, admission-delta tests and latent/refusalrewrite
  tests use synthetic diagnostic strings or generated miniature repositories;
  they do not record outcomes for programs affected by graph lowering.

Every stage 1 GAPS.md inspected (the executable gap harnesses are beside them):
- `stage1/cohere/css/GAPS.md`
- `stage1/cohere/cssnumbers/GAPS.md`
- `stage1/cohere/cssstrings/GAPS.md`
- `stage1/cohere/estree/GAPS.md`
- `stage1/cohere/formatfiles/GAPS.md`
- `stage1/cohere/gitignore/GAPS.md`
- `stage1/cohere/graphql/GAPS.md`
- `stage1/cohere/graphql/printer/GAPS.md`
- `stage1/cohere/json/GAPS.md`
- `stage1/cohere/lint/GAPS.md`
- `stage1/cohere/markdownblocks/GAPS.md`
- `stage1/cohere/markdowninline/GAPS.md`
- `stage1/cohere/mediaquery/GAPS.md`
- `stage1/cohere/selector/GAPS.md`
- `stage1/cohere/static_single_assignment/GAPS.md`
- `stage1/cohere/suppression/GAPS.md`
- `stage1/cohere/tsprinter/GAPS.md`
- `stage1/cohere/typeaware/GAPS.md`
- `stage1/cohere/values/GAPS.md`
- `stage1/cohere/yaml/GAPS.md`
- `stage1/typescript/parser/GAPS.md`
- `stage1/typescript/scanner/GAPS.md`

The cssnumbers report delegates its only compiler gap to cssstrings. formatfiles
records already-closed filesystem gaps; typeaware records product scope and
unsupported checker-runtime boundaries. Historical explanations of parser
workarounds are not newly changed verdicts. Gap tests, not prose word matches,
decide whether an outcome changed. See harness-files.txt for the complete source
inventory, including tests with synthetic Refused/NotYet inputs.

## Newly updated records and their proof

Each proof compares exit, stdout and stderr byte for byte to source Node, on
emitted JavaScript and native ASan/UBSan, then runs the ordinary leak check and a counted heap-balance check.
The native path is rebuilt uncached. Each has its own top-level oracle test.

| Program | Expected stdout | Proof test suffix |
| --- | --- | --- |
| stage3/fixtures/objects/29_parameter_default.a | `true\nfalse\nfalse\n` | ParameterDefault |
| stage3/fixtures/host/14_getCurrentDirectory.a | `true\ntrue\n` | HostCurrentDirectory |
| stage3/adapt/48-memoize/a-explicit.a | `42 42 1\n` | MemoizeExplicitPayload |
| stage3/adapt/48-memoize/a.a | `42 42 1\n` | MemoizePayload |
| stage3/adapt/48-memoize/b-explicit.a | `42 42 1\n` | MemoizeExplicitPending |
| stage3/adapt/48-memoize/b.a | `42 42 1\n` | MemoizePending |
| stage3/namespace-init-sys/cwd.a | `true\n` | NamespaceCurrentDirectory |

All have exit 0, empty stderr and clean leak checks. Prefix each suffix with
`TestRecordedGraphAdmission`. The stage 3 update generator independently reruns
Node and its sanitizer/leak hook before replacing only stage0. Recorded Node
stdout stays unchanged. Four memoize known entries are removed; cwd and host 14 now promise `checked`.
The host header was historical inside a Go-owned fixture, so acheck did not read
it; its graph-admission proof now pins the checked header as well.

The objects Python audit also consumes native.json. Its checked generator
re-measured all six compiled cases, adding objects 29 plus the already-compiled
14_comment_pending.a and 28_debugger.a, which the old summary omitted. No extra
status outcome changes. Default tests check summary completeness cheaply; the
update flag performs all backend, sanitizer and counted proofs before writing.

No runtime, compiler, count record, family wrapper or internal/fuzz change is
needed for these seven corrections. The compiler changes remain batch 8's own.


## Reproduce

Build the before compiler from exact main 9a941a49 and the after compiler from
this batch, using cloud/setup.sh's environment and identical pinned submodules.
The saved compiler-pins.json identifies both binaries used for this audit.

```sh
timeout 600s python3 notes/graph-regions-refusal-census/census.py \
  --before /tmp/wb8-main-adamic --after /tmp/wb8-batch-adamic \
  --records-ref 749e668e1d7aa32b0ed716b3f17ab6cba94b6f13 \
  --output /tmp/graph-refusal-census.jsonl
```

Inline promises are checked by the whole lower/oracle suites and the stage 1
gap harnesses. File-backed comparison is not a substitute for those tests.

The objects native-summary generator writes only after every compiled object
case passes recorded Node, JavaScript, ASan/UBSan/LSan and counted heap balance:

```sh
timeout 300s go test ./internal/oracle \
  -run '^TestObjectFixtureNativeSummaryMatchesCompiledCases$' \
  -count=1 -timeout 4m -args -update-graph-object-native
```

The normal summary guard only compares records; the ordinary fixture tests hold
actual execution to Node, and the new admission tests always run the counted
check. These additional recorded-outcome guards are test-only.
