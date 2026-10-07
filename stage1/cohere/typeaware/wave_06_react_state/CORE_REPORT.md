Built: native prepared-HIR validators for the three React claims, reusing wave 21 cores inside wave 06's owned directory; source analysis still explicitly refuses.
Commits: parent 48c6c7e076a9474db298a4fa294444c122ea96d2 remains rebased onto main e8ba3d5d; upstream core snapshot e9ec024c3213ec7c9ad967d39029e4dd004abb9a; implementation is the commit adding this report.
Checks: 91 valid controls, 48 findings, 36,582 canonical bytes match Go, native, ASan/UBSan, source Node and emitted JavaScript; both prepared-input corpora match.
Mutants: three semantic rule mutants and predecessor-row alias mutant caught only by Go bytes; three reporting mutants caught by Go and three removed-refusal mutants by panic 70.
Not covered: native source lowering/SSA, source unit gates, memo-scope resolution, production harness integration, end-to-end React timings, new bridge handle checks and deterministic Go two-creator phi text.

# Current status

The retained claims are `react-hooks/set-state-in-effect`,
`react-hooks/set-state-in-render` and `react-hooks/static-components`. Each
owned rule now offers `runPrepared(HirFunction): Diagnostic[]` and retains its
reporting methods. Calling `analyze()` still refuses before output. These are
unfinished end-to-end source ports, not a new batch of claims.

Fetching all origin heads found a newly available native prepared-HIR shelf on
wave 21. Six `.a` files are adapted from its pinned snapshot, with only the
relative Diagnostic import changed and provenance comments added. They provide
value/place and phi models, capture/context mapping inputs, native post-dominance,
setter/ref taint propagation, and static creation-site propagation. Native code
computes each finding. The owned public wrappers call those validators.
All new Adamic files are `.a`. No shared harness, registration, parser, protected
compiler file, bridge source or submodule pin changes.

The complete earlier twelve ports remain green on main e8ba3d5d and were already
pushed in 48c6c7e0. The current fetch and final remote check show that main has
not advanced. Their full source/corpus and released-handle evidence remains in
LANDING_REPORT.md. Only the worker's own branch is pushed; no main or area branch
is touched.

# Independent comparison and boundaries

The private Go fixture producer is adapted from wave 21 into
`testdata/cores_oracle.go`. In fixture-production mode it exports actual Go
Lower/Construct results, capture/context mappings, source spans, raw type alias
and symbol names, compilation-unit selection and the memo-erased graph variant.
It emits `.a` fixture programs. This is a test input producer, not a production
checker question or a native source frontend.

Expected findings come from a separate invocation of unchanged Go production
registry rules and their independent source loader. No Go finding verdict is
passed into the native validators. Canonical records include every finding,
fix and suggestion field; all three rules emit zero fixes and suggestions in
these cases. Positive findings in all three rules guard against count-only or
empty-output agreement.

The copied source-control evidence contains 100 strings. The independent Go
parser admits 91; all 91 admitted inputs are retained in twelve private batches.
They produce 48 findings and 36,582 identical bytes. Normal native and sanitized
native match Go in every batch; sanitizer stderr is empty. Each generated batch
also runs from its original source on Node, using the existing `oracle/node.mjs`
loader, and from the Adamic JavaScript backend. Both match complete Go bytes.
Source Node is independent of Adamic lowering.

The 77 frozen compiler roots match 5,318 bytes and the 287 repository roots match
18,485 bytes, with zero findings. Normal native, sanitized native, source Node
and emitted JavaScript match Go on both. These remain prepared-input corpus
comparisons: source-to-HIR work is done by the fixture producer, so they do not
establish native source parity. Source unit gates and the direct-useMemo fixture
annotation remain Go-provider dependencies; that annotation does not implement
full production memo-shadow resolution.

Each rule mutant and the graph mutation compiles, exits zero with empty stderr
and is caught only by complete-byte comparison:

| Mutant | Independent failure detector |
| --- | --- |
| Reverse effect's typed-setter guard | Go finding bytes |
| Reverse render's unconditional-block membership guard | Go finding bytes |
| Remove call/new static creator propagation | Go finding bytes |
| Reuse one mutable predecessor row | Go finding bytes on loop-after-throw control |

The existing reporting/refusal suite is also rerun after integration. Four
production Go findings match 2,411 bytes normally and under sanitizers. Its three
reporter-ID mutants lose Go equality; its three removed-source-refusal mutants
lose required exit 70. The JSX parser probe still exits 70, and each source
analyzer gives its refined missing-adapter reason. These six mutants are
reporting/refusal checks, not completed source-rule mutants.

The first emitted-JavaScript invocation used direct Node and omitted runtime
resolution, so it failed with ERR_MODULE_NOT_FOUND for `adamic`. The emitted
artifact was unchanged; using the existing oracle runtime loader fixes the
invocation. The failed command, output and error are retained under
`failed-loader-*`, separately from the complete successful backend run.

# Exact remaining blockers

Native source Lower/Construct and SSA generation, source capture namespaces,
compilation-unit selection, JSX lowering, memoization erasure/inlining and
scope-aware memo classification are still missing. The now available native
validators and post-dominance do not supply that source adapter. The shared
native parser still refuses JSX on the owned probe. Isolated JSX support on
other branches does not supply the remaining React lowering passes.

No bridge question is added for these fixture-only validators, and no runtime
checker handles are used. New-question released-handle checks are therefore not
claimed for this continuation. Prior real-bridge ownership checks remain green
on the unchanged main/compiler base. A future source adapter must obtain any
additional raw checker facts through isolated question files.

There is a separate observed oracle ambiguity. Twenty-four independent runs
of unmodified Go on the same two-creator static-component source produce two
canonical messages, frequencies 20 and 4: creation at createA or at createB.
Go's phi operands are a map and the rule takes the first dynamic operand. The
native core preserves its supplied operand ordering. It cannot promise the
same random choice as a separate Go execution. No output is normalized, and
that input is retained as an ambiguity observation, not a passing parity case.
Source, all 24 outputs and commands are preserved in core_evidence.

Under the instruction to stop on dependencies outside owned rule directories,
source integration remains blocked. The portable validators are now implemented,
tested and pushed. The three claims stay unfinished and reserved. No new rules
are claimed and no root test-matrix pass is asserted.

# Timing and reproduction

For the 91 controls, native prepared-input process wall totals
0.021282s, full production Go totals 0.324731s and Go fixture
preparation totals 0.343328s. Builds are excluded. These processes
consume different inputs, so the numbers are not an end-to-end native/Go lint
benchmark or a speedup claim. Comparable React source timings remain uncovered.
The earlier twelve complete suites have comparable measurements in
LANDING_REPORT.md.

```sh
source /workspace/adamic-tools/env.sh
TMPDIR=/workspace python3 stage1/cohere/typeaware/wave_06_react_state/validate_cores.py --scratch /workspace/wave-06-react-cores --compiler /workspace/wave-06-typescript > /tmp/wave-06-react-cores.log 2>&1
TMPDIR=/workspace python3 stage1/cohere/typeaware/wave_06_react_state/validate_core_backends.py --scratch /workspace/wave-06-react-cores > /tmp/wave-06-react-core-backends-final.log 2>&1
TMPDIR=/workspace python3 stage1/cohere/typeaware/wave_06_react_state/validate_partial.py --scratch /workspace/wave-06-react-core-refusals-final > /tmp/wave-06-react-core-refusals-final.log 2>&1
TMPDIR=/workspace go vet ./... > /tmp/wave-06-react-core-vet.log 2>&1
gofmt -l stage1/cohere/typeaware/wave_06_react_state/testdata > /tmp/wave-06-react-core-format.log
```

Every test writes directly to files. Core, backend and partial reporting/refusal
validation pass; vet and Go formatting output are empty, Python syntax passes
and git diff --check passes. Complete streams, generated `.a` fixtures,
commands/exits, controls, ambiguity evidence and hashes are in
[core_evidence](core_evidence).

The existing successful cloud setup is reused: tools 0s, submodules 0s, cache and
total 105s; `nproc` 5, CPU quota 4, memory 17.6 GB. Go 1.27.1, clang 20.1.8 and
Node 24.19.0 remain unchanged. A formatting invocation initially lacked the
sourced toolchain and found no gofmt; it was rerun successfully after sourcing
`/workspace/adamic-tools/env.sh`. No PR is opened.
