# Type-aware wave 05

Branch: `codex/typeaware-wave-05`.
Base: `origin/codex/tsgo-c-library`, `0d540f413625f016f20fea39761c7b184f335de6`.

Claimed rules, positions 13, 14 and 15 after excluding the 26 existing ports
from the combined checker-dependent ranking recorded by VOLUME_REPORT.md:

| Position | Rule | Compiler | Repository | Total |
| --- | --- | ---: | ---: | ---: |
| 13 | @typescript-eslint/no-unnecessary-type-parameters | 30 | 1 | 31 |
| 14 | no-useless-assignment | 17 | 12 | 29 |
| 15 | @typescript-eslint/restrict-template-expressions | 23 | 0 | 23 |

Selection uses `validation-volume/compiler-all.counts` and
`validation-volume/repository-all.counts`, descending combined count with lexical
ties. The 16 volume-suite and 10 coverage-suite ports are excluded before
indexing. All origin heads were fetched and searched for matching port filenames
and claim contents before this commit. No matching port or claim was found;
configuration references are not implementations. No rules were skipped.

## Continuation 1

After completing and pushing the original three rules through c5403d46,
fetched all origin heads on October 7, 2026. Combined VOLUME_REPORT.md's
197 checker-dependent count rows, sorted by descending combined count with
lexical ties, and excluded ports on origin/main and origin/codex/tsgo-c-library
and names reserved in every origin branch's typeaware claims.

The next three unclaimed rules are reserved here before implementation:

- nexus/correctness-no-process-exit-after-output (0 compiler, 0 repository)
- nexus/correctness-no-uncleared-race-timeout (0 compiler, 0 repository)
- nexus/correctness-require-blocking-standard-streams (0 compiler, 0 repository)

These names have no implementation on main or the bridge branch. Inventory,
configuration and skip-types records are not ports. Positive controls are
required because the recorded corpus counts are zero. Implementation stays in
this wave's own directories, without editing the shared generator or harness.

Continuation 1 status: all three reserved rules are implemented and validated.
The timer rule retains its prior DOM and Node controls. The output rules match
production Go on 30 direct controls and 30 module programs, and all three match
both frozen corpora normally and under ASan, UBSan and LSan. Each rule has a
comparison-only mutant; a CFG-edge mutant, direct bridge-guard mutants and
released-handle checks also pass their assertions.

The previous constructor refusal was overcome inside this wave's directory:
its own forward CFG accepts fully constructed bindings instead of constructing
bindings while its own constructor is still being lowered. No inheritance,
shared compiler/bindings repair, registration generator or shared harness edit
was needed. New raw checker questions each have their own Go and .a files and
one physical switch-registration line. Details are in
`../wave_05_next/OUTPUT_REPORT.md`; the old refusal remains as historical evidence.
No additional rules are claimed in this update.

## Continuation 2

All six prior reservations are complete, tested and pushed through 9d69fd77.
After fetching all origin heads, rechecked the 197-rule combined by-volume
ranking against ports on origin/main and origin/codex/tsgo-c-library and every
origin branch's typeaware claim files. The first three remaining rules are
reserved here before implementation:

- react-hooks/globals (0 compiler, 0 repository)
- react-hooks/immutability (0 compiler, 0 repository)
- react-hooks/no-deriving-state-in-effects (0 compiler, 0 repository)

The scan found 30 unique claim blobs reserving 137 ranked rule names and 25
ranked baseline ports. Seven high-volume apparent candidates were skipped
because their existing native implementations emit through Rules.add rather
than directly constructing Diagnostic. The 26th baseline port,
method-signature-style, is outside this checker-dependent ranking.
These three React hook names have no baseline implementation or existing
origin claim. Implementation remains inside this wave's directories.
Continuation 2 status: blocked and unported. The shared stage-1 parser refuses
ordinary JSX (`<div />`) with `expected GreaterThanToken, got SlashToken`, exit
70, before native rule execution. Independent production Go reports on a
positive JSX control for each of these three rules. A non-JSX control succeeds,
and a skipped-parser probe mutant is rejected by the refusal assertion.
The parser is outside this wave's allowed directories; under Ahra's instruction
to stop on other blockers, no shared frontend repair or further claim is made.
Exact reproduction and limitations are in `../wave_05_react/REPORT.md`.
All six earlier reservations remain complete, tested and pushed.

## React parking and continuation 3

Under Ahra's new parking instruction, continuation 2 is PARKED and counts as
finished for landing-first; it is not represented as implemented:

- react-hooks/globals: native JSX parsing and React compilation-root detection
  depend on the JSX work landing through area/stage1-lint.
- react-hooks/immutability: native React high-level IR, single-assignment and
  capture/frozen-value analysis are absent; dependency is #dnv6f2c, plus JSX.
- react-hooks/no-deriving-state-in-effects: native React high-level IR,
  single-assignment and capture/taint analysis are absent; dependency is
  #dnv6f2c, plus JSX.

Existing probes and production Go controls are pushed in wave_05_react.
No partial rule verdict implementation or React parity is claimed.

All six implemented ports are green and pushed on current main f8013f0b
through abcef242d. After fetching 529 origin refs, scanning 31 unique claim
blobs and the combined 197-rule ranking, the next three eligible rules are:

- require-await (0 compiler, 0 repository)
- symbol-description (0 compiler, 0 repository)
- valid-typeof (0 compiler, 0 repository)

The remaining React names are parked pending JSX/analysis.
require-atomic-updates is skipped because its production implementation
depends on ecmascript/control_flow_graph and captured-binding analysis.
The selected rules do not import high-level IR, SSA or capture-analysis modules.
They have no implementation on the main/bridge baselines or matching origin
claim. Selection evidence is wave_05_core/selection.json. These reservations
are pushed before native implementation. New rule modules use .a, each with
rule.json numeric kinds and callbacks taking the supplied node. The shared generator and harness
remain untouched; raw bridge question registration uses one line per question.

Continuation 3 status: all three rules implemented as numeric handed-node .a
callbacks and validated against production Go on 106 controls and both frozen
corpora, including complete suggestions, sanitizer builds, comparison-only
mutants and released handles. Main advanced to c01907a7; all nine completed
ports were rebased and revalidated before publication. See
`../wave_05_core/REPORT.md`. The three React reservations remain PARKED with
the blockers above; no React verdict parity is claimed. No additional rules
are reserved by this update.

## Shared-harness area landing

Rebased onto origin/area/stage1-lint 7481e032 (includes harness 41eb6eab2
and main 39638d9e). The nine completed ports are revalidated before publication.
The previous native JSX parsing blocker is CLOSED on this base: all three
retained positive JSX probes parse successfully. The three React claims stay
PARKED for native analysis, not JSX: globals needs compilation-root/capture/
reference-write classification; immutability needs HIR/SSA/capture/frozen-value
analysis; no-deriving-state-in-effects needs HIR/SSA/capture/effect-taint analysis.
The shared analysis dependency is #dnv6f2c. No React verdict parity is claimed.
The fresh all-origin ranking has no unclaimed rule. See
`../wave_05_core/AREA_LANDING_REPORT.md`; no additional rules are reserved.

Latest area landing: all nine completed ports were rebased and revalidated
on origin/area/stage1-lint d65a8f93, including current main 39638d9e.
The three React analysis claims stay PARKED with the blockers above.
All ranked names are ported or claimed; no new reservation is made.
Validation evidence is in `../wave_05_core/AREA_2_REPORT.md`.

Latest landing: all nine completed ports were rebased and revalidated on
area b84a9d93, including main c7991b90. No ranked name is unclaimed;
the three React analysis reservations remain PARKED. See
`../wave_05_core/AREA_3_REPORT.md` for the supplied-input filtered gate.

Latest landing: all nine completed ports are rebased and revalidated on
area b4691483, including main c7991b90. No ranked name is unclaimed;
the three React analysis claims remain PARKED. Exact filtered-gate scope
and outputs are in `../wave_05_core/AREA_4_REPORT.md`.

Latest landing: all nine completed ports are rebased and revalidated on
area d3a37422, including main b6b1538b. No ranked name is unclaimed;
the three React analysis claims remain PARKED. Exact filtered-gate scope
and outputs are in `../wave_05_core/AREA_5_REPORT.md`.


## Shared checker merge landing

Merged area c4bdc23fa through 7b908dc74, preserving both sides' checks.
Nine unified node descriptors now use RuleContext.checker and its recording /
replay stream, including same-program foreign-file requests. No private
checker is created by these unified descriptors. Seven ports pass all original
typed programs and witnesses across all four runtimes. Type parameters and
blocking streams retain three exact parser failures. All nine mutants are
caught. The full supplied-input lint gate remains red on retained parser / JSX
inventory failures, plus the named TSGoError dependency skip. React claims
remain PARKED for reference-write / HIR / capture analysis; JSX is no longer
their blocker. See ../wave_05_core/CHECKER_LANDING_REPORT.md and its evidence.
No additional reservation is made.
