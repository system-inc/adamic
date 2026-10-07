# Type-aware wave 28 claim

Branch: `codex/typeaware-wave-28`.
Base: `origin/codex/tsgo-c-library`, `0d540f413625f016f20fea39761c7b184f335de6`.

The VOLUME_REPORT.md all-family counts in validation-volume are ranked by
combined compiler and repository volume descending, with lexical ties. Excluding
all 26 ports on the base leaves 172 checker-dependent candidates.

| Remaining position | Rule | Compiler | Repository |
| --- | --- | ---: | ---: |
| 82 | base/correctness-require-verify-optional-parity | 0 | 0 |
| 83 | block-scoped-var | 0 | 0 |
| 84 | getter-return | 0 | 0 |

All three are claimed here before implementation. No matching rule source or
claim was found across 268 fetched origin refs (64 distinct stage1 trees).
Counts, catalogs and configuration references are not implementations.
No rule was skipped. Zero-volume corpora require positive controls and mutants.

## Continuation claim

After the original three ports were tested and pushed at `24c600f017352ef3dd8d7c80f104a26099bc4b85`, origin was fetched again. The first three by combined volume, excluding the 26 base ports, main ports, and canonical rule names in claims across all 325 origin refs, are:

1. `nexus/correctness-no-process-exit-after-output` (0 compiler, 0 repository).
2. `nexus/correctness-no-uncleared-race-timeout` (0 compiler, 0 repository).
3. `nexus/correctness-require-blocking-standard-streams` (0 compiler, 0 repository).

There were 33 distinct Markdown claim blobs naming 96 ranked rules at this snapshot. None of these three is claimed or implemented on main or the bridge branch. This continuation claim is pushed before any implementation. Shared registration generators and shared test harness files remain outside this worker's territory.

## Third batch claim

The six earlier ports are tested and pushed at `4f3a94110b2e2ec55098cef7cf1a913881b3a53b`. After another all-heads fetch, 348 origin refs contain 33 distinct claim blobs naming 126 ranked rules. Excluding those claims and the base/main ports leaves these first three by combined volume:

1. `no-throw-literal` (0 compiler, 0 repository).
2. `no-useless-backreference` (0 compiler, 0 repository).
3. `prefer-arrow-callback` (0 compiler, 0 repository).

No matching source literal occurs in the 34 distinct origin typeaware trees. No selected rule was skipped. This claim is pushed before any third-batch implementation.

## Fourth batch claim

All nine earlier ports are tested and pushed at `431ab666b8d3fc3919f8d6a434e58538a65bc5ac`. After another all-heads fetch, 389 origin refs contain 33 distinct Markdown claim blobs naming 142 ranked rules. Excluding those claims and the base/main ports leaves these first three by combined volume:

1. `react-hooks/set-state-in-effect` (0 compiler, 0 repository).
2. `react-hooks/set-state-in-render` (0 compiler, 0 repository).
3. `react-hooks/static-components` (0 compiler, 0 repository).

No matching implementation literal was found in the 34 distinct origin typeaware trees. No selected rule was skipped. This claim is pushed before fourth-batch implementation.

Fourth-batch status: claimed but unported. Native JSX parsing rejects all three Go-positive controls before rule execution. The required native SSA/control-flow IR is also absent on this branch. See `../wave28_fourth/REPORT.md` and its reproduction evidence. Stopped without editing shared files or claiming another batch.

Resumption audit: JSX support is published on `origin/codex/stage1-jsx-lint` at `a8a62d62`, but is not integrated on this branch/main/bridge/harness. The native SSA and control-flow substrate required by all three remains unavailable in the inspected stage1 sources. The blocker reproduction passes again; all three remain claimed and unported. See the report's resumption audit.

## Parked React analysis claims

Per the explicit parking instruction, react-hooks/set-state-in-effect, react-hooks/set-state-in-render and react-hooks/static-components are **parked** and count as finished for the landing-first cap. They remain unported. Blockers: native high-level IR, single-assignment, translated closure captures, post-dominance and control-dominance analysis. Cohere analysis modules are being ported to Adamic on #dnv6f2c; JSX support is landing on area/stage1-lint. Existing probe/source evidence is in wave28_fourth/REPORT.md. This status does not claim native rule parity. The nine completed ports are green on current main f8013f0b and pushed at 786d33d3c.

## Fifth batch claim

The React high-level-analysis claims are parked and their status is pushed at f7b44295d. The nine completed ports remain green on current main f8013f0b. The all-heads audit inspected 529 origin refs, 33 distinct Markdown claim blobs and 34 distinct typeaware implementation trees. Eighteen ranked candidates remain unclaimed; the first three below require syntax and checker declaration/type questions, not the parked native HIR/SSA/capture substrate:

1. `react/jsx-fragments` (0 compiler, 0 repository).
2. `react/jsx-no-constructed-context-values` (0 compiler, 0 repository).
3. `react/jsx-no-undef` (0 compiler, 0 repository).

No implementation match was found for these three across the inspected origin typeaware trees. Their production Go sources use JSX node listeners and AST/checker walks rather than HIR/SSA/capture lowering. This claim is pushed before implementation or owned blocker checks. JSX parsing remains an integration dependency to inspect.

Fifth-batch status: claimed, blocked and unported. Named rule.json kinds agree with the actual production Go listeners; three Go-positive JSX controls are rejected by current main's native parser before listener execution. JSX integration through area/stage1-lint is the named blocker. See ../wave28_fifth/REPORT.md. These are not parked HIR/SSA claims; no sixth batch is claimed.
