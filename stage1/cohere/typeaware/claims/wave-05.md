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
