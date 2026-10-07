Built: reporting-only portions of three newly claimed React rules; the earlier six rule ports remain complete and pushed.
Commits: completed correctness ports `99c9e8d3` / evidence `8673be2b`; React claim `f3b2e6f9`; partial reporters `758b9c60`.
Commands and outputs: Go controls PASS with 2 memo, 3 purity and 1 refs findings; 11 reporting records match 5948 bytes across Go, native, ASan/UBSan/leaks, source Node and emitted JavaScript.
Mutants: three renderer end+1 mutants exit 0 and fail only byte comparison; three removed-refusal mutants exit 0 and fail the expected NotYet panic check.
Not covered: React analysis verdicts, corpus parity, full-rule mutants, checker-handle checks or timings for these three; missing native React HIR and JSX parsing block them.

## Claimed rules and honest status

- `react-hooks/preserve-manual-memoization`
- `react-hooks/purity`
- `react-hooks/refs`

These were the first three unported/unclaimed names after the previous three
were pushed. Selection fetched all 358 origin references, used the 197-rule
combined volume ranking, excluded the 25 ranked baseline ports and 139 names
in remote claims, and preserved the search proof in
[wave-04-selection-2.json.gz](../claims/wave-04-selection-2.json.gz).
The claim was pushed in `f3b2e6f9` before writing any implementation.

These three are **unfinished and remain claimed**. Their separate `.a` files
implement message selection, formatting, span transport and zero fix/suggestion
fields. They do not implement a lint analysis. Calling `analyze()` gives an
explicit `NotYet` panic rather than silently returning an empty finding list.
The reporter is deliberately separate from the verdict so it can be integrated
when the prerequisite native substrate exists.

## Exact blockers

The purity control is the production rule's original vendored fixture. Go reports
three findings, one for each of `Date.now`, `performance.now` and `Math.random`.
The unchanged shared native parser exits 70 before a rule can run:

```text
adamic: panic: parser slice expected GreaterThanToken, got Identifier at 166 in /workspace/typeaware-wave-04-react/final-2/purity.tsx
```

The memoization and ref controls parse natively, so JSX is not the only blocker.
All three production analyses consume
`cohere/internal/lint/ecmascript/high_level_intermediate_representation`:

- Memoization calls `CloneFunction(ForFunction(...))` and
  `AnalyzePreservedManualMemoization`. The latter outlines and inlines functions,
  infers reactivity and mutable ranges, builds SSA/reactive scopes, aligns and
  merges scopes, collects dependencies, rebuilds the reactive tree, prunes scopes
  and validates whether source memoization survived. A syntax-only approximation
  cannot reproduce that comparison.
- Purity walks HIR instructions and phis, propagates builtin aliases, carries
  captured values across closures and re-raises nested calls at their true spans.
- Refs consumes `ForFunctionWithoutManualMemoization`, SSA/phi values, captures,
  nominal checker types and a six-element lattice with an exact ten-round bound.

There is no native React HIR implementation under stage 1 on this branch, and
no checker bridge question supplies it. Scoped source searches return exit 1
with zero stdout and stderr; their commands and outputs are preserved in
[evidence](evidence). The existing syntax CFG from the completed correctness
ports provides control-flow edges, not these typed SSA values, capture spaces,
reactive scopes or memoization passes.

Completing these analyses requires a native port of that shared substrate, plus
JSX support for the corpus. Running the Go validators in a new bridge question
would pass Go lint verdicts across the bridge and would not be a native rule
port. No such shortcut was added.

Ahra's correction says "Keep your changes inside your own rule directories" and
"If anything else blocks you, say exactly what it is and stop, rather than
editing shared files." The shared parser, lowering, harness and registration
files were left untouched. Work stops here with the missing dependencies
recorded; this is not an automatic approval rejection or a request for permission.
No further rules were claimed.

## What the partial checks establish

An isolated Go test overlay imports the production React rule package and runs
its unmodified rules through the typed test API. It exports positive controls:
`reassignedContextCapture`, `purityImpureFunctionsInRender` and the ref-read
fixture. They report 2, 3 and 1 findings respectively. The overlay also exports
all five `refsMessageFor` arms, including the non-convergence condition.

The 11 records are replayed through the native reporters, producing identical
complete canonical diagnostic lines: 5948 bytes. Normal and sanitized native,
source on Node with types stripped by the existing oracle loader, and emitted
JavaScript on Node all match the independent Go text. This verifies reporting
only; replaying Go's test records does not establish a native analysis verdict.

Each renderer's end+1 mutant compiles, exits 0, produces 11 records and has empty
stderr; only byte comparison kills it. Each analysis-refusal mutant removes its
panic, compiles and exits 0 with empty stderr; the expected exit-70 assertion
kills it. These are partial reporting/refusal mutants, **not full-rule mutants**.
The parser probe's memo and refs inputs exit 0; its purity input exits 70 as above.
The three analysis probes exit 70 with their distinct `NotYet` reasons.

```sh
source /workspace/adamic-tools/env.sh
python3 stage1/cohere/typeaware/wave_04_react/validate_partial.py /workspace/typeaware-wave-04-react/final-2 > /workspace/typeaware-wave-04-react/final-2-validation.log 2>&1
go vet ./... > /workspace/typeaware-wave-04-react/vet.log 2>&1
gofmt -l stage1/cohere/typeaware/wave_04_react/testdata/export_test.go > /workspace/typeaware-wave-04-react/gofmt.log 2>&1
```

Validation ends `PASS partial reporting only`; vet and formatting logs are empty.
All test output went to files. The full repository gate and full-rule corpus
comparisons were not run for these unfinished rules. Native/Go timing comparisons
would be misleading while native analysis cannot run, so none are claimed.

The completed earlier correctness trio's normal/sanitized corpus results, three
comparison-only rule mutants, released-handle registry mutant and timings remain
in [its complete report](../wave_04_next/REPORT.md): compiler medians native
1.815s versus Go 0.335s; repository native 0.266s versus Go 0.118s.

## Dependency refresh on 2026-10-07

Fetched every origin head again and froze all 399 reference SHAs in
[evidence/refresh/origin-inventory.json](evidence/refresh/origin-inventory.json).
The inventory searches native source filenames for HIR, React, memoization and
JSX candidates; it is not an exhaustive semantic search of all source bodies.
Main is `e011f8f6`, the bridge branch `5afbdb83`, and the harness branch
`f4d98cab`. None supplies a native React HIR implementation in this inventory.

There is now a reusable isolated JSX parser on `origin/codex/typeaware-wave-16`,
SHA `58d204d3`, in four `wave16_jsx*.a` dependency files. Its follow-up report
records positive JSX comparison controls. The narrow JSX adapter helpers on the
helper branches describe raw JSX nodes; they do not implement React HIR. Thus
JSX is a potential integration dependency, rather than evidence that no JSX
implementation exists anywhere. This branch still uses the unchanged shared
parser, which rejects the original purity control. No copied parser or shared
parser edits were added while the separate HIR prerequisite remains missing.

The React worker on wave 08 independently records missing HIR lowering, SSA and
instruction transfer. Its partial lattice/derivation kernels do not supply the
three pipelines needed here. Worker reports are frozen beside the inventory.

Rebuilt and reran the owned partial validation:

```sh
source /workspace/adamic-tools/env.sh
python3 stage1/cohere/typeaware/wave_04_react/validate_partial.py /workspace/typeaware-wave-04-react/refresh > /workspace/typeaware-wave-04-react/refresh-validation.log 2>&1
```

PASS: the 11 reporting records still match Go at 5948 bytes in native, sanitized
native, source Node and emitted JavaScript. All three end+1 renderer mutants
exit zero with empty stderr and are caught by byte comparison; all three
removed-refusal mutants are caught by the expected-panic assertion. Memo and
refs parse controls pass; the purity JSX parser probe and all three explicit
analysis refusals still exit 70. These remain partial reporting/refusal checks,
not completed rule ports, full-rule mutants, corpus analysis or rule timings.

No further rules were claimed. The exact remaining prerequisite is native
React HIR lowering and SSA/capture propagation, plus memoization's reactive-scope
passes. The user instruction to stop on prerequisites outside the owned rule
directories still applies. The prior completed six rule ports remain pushed.
