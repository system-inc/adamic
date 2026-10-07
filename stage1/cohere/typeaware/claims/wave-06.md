# Type-aware wave 06 claim

Branch: `codex/typeaware-wave-06`.
Base: `origin/codex/tsgo-c-library`, `0d540f413625f016f20fea39761c7b184f335de6`.

Claimed rules, positions 16, 17 and 18 after excluding the base branch's
26 ports from VOLUME_REPORT.md's combined checker-dependent volume counts:

| Position | Rule | Compiler | Repository | Total |
| --- | --- | ---: | ---: | ---: |
| 16 | adamic/nominal-class | 18 | 2 | 20 |
| 17 | @typescript-eslint/no-duplicate-type-constituents | 19 | 0 | 19 |
| 18 | no-useless-return | 16 | 0 | 16 |

Checked all 268 fetched origin references, including claim paths and native
rule source mentions, before implementation. No prior port or claim found.
Config set strings and inventory/count records are not implementations.
No rules skipped.

## Continuation claim

The original three rules are completed and pushed in `027adc366b9f2fbd379df517779a2dd2dfb6ee82`.
Fetched all origin heads again, checking 327 remote refs and every distinct
claim document plus native source names on origin/main and
origin/codex/tsgo-c-library. The first three available entries in the combined
volume ranking, descending count with lexical ties, are now reserved here:

1. `nexus/correctness-require-child-process-error-listener` (0 compiler, 0 repository)
2. `nexus/correctness-require-response-status-check` (0 compiler, 0 repository)
3. `nexus/performance-no-independent-await-in-loop` (0 compiler, 0 repository)

The initially available process-exit, race-timeout and blocking-stream candidates
became claimed on other origin branches during the final fetch and were skipped.
No continuation implementation preceded this claim commit and push.
New implementation and validation files will stay in this unit's own rule
directories; existing shared harness and registration generator stay untouched.

The continuation rules are complete; native implementations and complete byte
comparison evidence are in `wave_06_next/REPORT.md` and `wave_06_next/validation/`.
No further rules are reserved by this claim.

## Second continuation claim

All six earlier rules are complete, tested and pushed through
`d7f911f4ac3b789ddd26d1e0cd974cadf8e787e0`. A fresh all-heads origin fetch
examined 341 remote refs, 33 distinct Markdown claim blobs, 197 ranked rules,
117 claimed rules and 25 ranked ports on main or the bridge base.
The first three remaining entries in descending combined volume with lexical
ties are reserved here before implementation:

1. `no-new-func` (0 compiler, 0 repository)
2. `no-new-native-nonconstructor` (0 compiler, 0 repository)
3. `no-new-wrappers` (0 compiler, 0 repository)

Implementations and validation will remain in an owned `wave_06_constructors/`
directory. No additional rules are claimed.

The second continuation is complete. Implementations and full validation
evidence are in `wave_06_constructors/REPORT.md` and its `validation/` directory.

## Third continuation claim

The earlier nine rules are complete and pushed through
`7e889b141f26d239a2797050e1225fedb93bc44c`. An all-heads fetch checked
347 origin refs, 33 unique claim documents, 126 claimed ranked rules and 25
ranked ports on origin/main or origin/codex/tsgo-c-library. The next three
entries (descending combined volume, lexical ties) are reserved before code:

1. `no-throw-literal` (0 compiler, 0 repository)
2. `no-useless-backreference` (0 compiler, 0 repository)
3. `prefer-arrow-callback` (0 compiler, 0 repository)

The intervening constructor and other core entries are now claimed elsewhere.
Native implementations and evidence belong in `wave_06_callbacks/`.

The third continuation is complete; its native rules and complete comparison
evidence live in `wave_06_callbacks/REPORT.md` and `wave_06_callbacks/validation/`.

## Fourth continuation claim

All twelve earlier rules are complete and pushed through
`d0dc8c2c7a88ab3ddfb3b49d4ffc1e07c8be8f5e`. A fresh all-heads fetch
checked 389 origin refs, 33 unique claim documents, 142 claimed ranked rules
and 25 ranked ports on main or the bridge base. The first three remaining
entries in descending combined volume with lexical ties are reserved before code:

1. `react-hooks/set-state-in-effect` (0 compiler, 0 repository)
2. `react-hooks/set-state-in-render` (0 compiler, 0 repository)
3. `react-hooks/static-components` (0 compiler, 0 repository)

Implementations and evidence belong in `wave_06_react_state/`. No additional
rules are reserved by this claim.

The fourth continuation is **parked** under Ahra's explicit React parking instruction. All prepared-HIR validators, reporting, refusals, numeric listener declarations and their oracle/sanitizer evidence are pushed through `5cb1ec2bf7302e7dbe3e56f71d9a19b5fcabdf3d`, rebased and re-green on main `f8013f0baac41ddc340d76f83bddde38536a8f07`.

Parked rules and blockers:

- `react-hooks/set-state-in-effect`: native source-to-HIR lowering, SSA, captures and manual memoization erasure.
- `react-hooks/set-state-in-render`: native source-to-HIR lowering, SSA, compilation-unit selection and scope-aware memo classification.
- `react-hooks/static-components`: native JSX lowering, SSA and compilation-unit selection.

The analysis modules are being ported on #dnv6f2c and JSX support is landing on `area/stage1-lint`. Prepared-HIR validation and post-dominance already exist locally; the missing dependency is the source adapter. See `wave_06_react_state/CORE_REPORT.md` and `LANDING3_REPORT.md`. Parked counts as finished for landing-first only, not completed source parity.

## Fifth continuation claim

A fresh all-heads fetch confirms main remains f8013f0b and the only own pushed branch is landing-ready at 5cb1ec2b. Selection scans 529 origin references, 33 distinct claim blobs and all 197 checker-dependent ranked rules, with 154 claimed ranked rules and 25 main/base source mentions excluded. Combined volume is descending with lexical ties.

Reserved before implementation:

1. `react/jsx-fragments` (0 compiler, 0 repository)
2. `react/jsx-no-undef` (0 compiler, 0 repository)
3. `react/no-adjacent-inline-elements` (0 compiler, 0 repository)

The intervening `react/jsx-no-constructed-context-values` is skipped under the analysis exclusion: its stability extension uses function evaluation and capture/escape checks (`functionEvaluation`, `anyEscapes`, `StaysHome` in jsx_no_constructed_context_values_stability.go). The selected rules do not need that analysis. Implementations and private validation belong in `wave_06_jsx/`. JSX/numeric-driver integration may still be blocked by the shared parser; findings must refuse rather than silently omit unsupported source analyses.

The fifth continuation is partially implemented and tested in `wave_06_jsx/`. The three prepared-node kernels match Go, native, Node and sanitizers on 61 controls with 47 findings; each semantic mutant is caught. Source integration is blocked on the shared numeric JSX parser, handed-node driver and raw syntax/declaration adapter. Source entry points explicitly refuse. These are reserved claims, not completed source ports; see `wave_06_jsx/REPORT.md`.

Named harness follow-up: `ab70f38d4` provides working JSX parsing, .a loading and the shared reporting model. Its isolated parser probe matches Node/native/sanitizers. Remaining blockers are the numeric node-kind contract and raw checker/declaration adapter; reporting and JSX syntax on that branch are no longer blockers. Current main remains c01907a7, and the named harness has not landed there yet. See `wave_06_jsx/HARNESS_REPORT.md`.

## Named-kind source follow-up

The latest ast.Kind-name correction supersedes earlier numeric-driver blockers. All three fifth-continuation rules now have actual handed-source adapters and raw checker wiring, with 91 controls/68 findings matching Go byte for byte under native/sanitizers, plus both frozen corpora. The quoted-import mismatch is fixed by preserving imported node kinds. Named listeners are declared and validated. Shared production registration and parser integration still await ab70f38d4 landing; existing no-argument source entry points refuse until wired. No additional claims are taken. React HIR/SSA/capture claims remain parked. The branch is rebased and re-green on 39638d9e, including 73 fresh prior native rebuilds and 165 oracle replays. See wave_06_jsx/SOURCE_REPORT.md.

Post-push all-heads selection finds no remaining eligible ranked rule: 606 origin references, 33 distinct Markdown claim blobs, 172 ranked names in claims and the other 25 already ported on main/base. No further claims. Evidence is in wave_06_jsx/source_evidence/selection.

## Landed harness follow-up

Rebased onto origin/area/stage1-lint d65a8f931 (including 50a5f105/41eb6eab2), retaining main 39638d9e ancestry. Repository JSX parsing is now available. All three owned handed-source adapters match Go/native/sanitizers on 91 controls/68 findings and both frozen corpora; 73 fresh earlier native builds and 165 replays pass. The parser probe now succeeds and matches Node; its count mutant is caught. Shared integration is still blocked by absent checker handle/filename in RuleContext, absent checker archive in the shared native build, absent TypeChecker in the shared Go oracle, and registry wiring. These shared files are untouched. React HIR/SSA/capture/source adapters remain parked. A scan of 614 origin refs finds no eligible unclaimed ranked rules. See wave_06_jsx/HARNESS_LANDING_REPORT.md and harness_landing_evidence.
