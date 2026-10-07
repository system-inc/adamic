# Type-aware wave 09

Base: origin/codex/tsgo-c-library, 0d540f413625f016f20fea39761c7b184f335de6.
Branch: codex/typeaware-wave-09.

Claimed rules, positions 25 through 27 in descending combined compiler and
repository counts from VOLUME_REPORT.md's validation-volume tables, excluding
the 26 rules already ported on the base, with lexical ties:

| Position | Rule | Compiler | Repository | Total |
| --- | --- | ---: | ---: | ---: |
| 25 | prefer-const | 6 | 1 | 7 |
| 26 | radix | 4 | 3 | 7 |
| 27 | @typescript-eslint/non-nullable-type-assertion-style | 4 | 0 | 4 |

All fetched origin branches were inspected for ports and claims before this
commit. None of these three has an existing stage 1 port or claim. Inventory
entries and config set names are not implementations. No rules were skipped.

## Continuation claim, October 7

The original three rules are implemented, tested and pushed in 03ea146b, with
validation logs in d0e9ba7c. Fetched all origin heads again before selection:
325 remote refs, 33 distinct Markdown claim blobs under typeaware/claims.
Excluded native ports on origin/main and origin/codex/tsgo-c-library, including
abbreviated TypeScript rule names, and every ranked rule named in an origin
claim file. Inventory/config references are not implementations.

The first three remaining rules by combined compiler plus repository volume,
with lexical ties, are reserved on this same branch:

- `nexus/correctness-no-process-exit-after-output` (0 compiler, 0 repository)
- `nexus/correctness-no-uncleared-race-timeout` (0 compiler, 0 repository)
- `nexus/correctness-require-blocking-standard-streams` (0 compiler, 0 repository)

This claim update is pushed before code. New sources use .a. Changes stay in
this unit's rule directories; shared harness and registration generator are
owned by codex/lint-harness-dot-a and will not be edited.

## Continuation withdrawn after concurrent-claim refresh

A second all-head fetch found concurrent claims for all three continuation
rules. Wave 02's claim 958c5c2f was committed at 2026-10-07T00:59:43Z,
before this branch's 24667c85 at 2026-10-07T01:00:59Z. Wave 09 yields all
three continuation reservations and takes no replacements. The original
prefer-const, radix and non-nullable-type-assertion-style ports remain complete.

Work performed before the collision was observed is saved as an explicitly
withdrawn experiment in ../validation-wave-09-continuation/withdrawn-experiment.tar.gz, with no
shared registration edits.
The timeout port passed independent full-byte controls and both corpora using
a temporary checker registration overlay. The other two rules are partial,
not completed ports. See ../WAVE_09_CONTINUATION_REPORT.md for gaps and evidence.

## New continuation claim after withdrawal

The original three rules are complete and pushed. The previous continuation
reservations remain withdrawn. Fetched all origin heads again: 335 remote refs,
33 distinct Markdown claim blobs, 114 ranked rule names named in claims.
Excluded actual native ports on main ef3d907e and bridge branch 5afbdb83,
including abbreviated TypeScript names, and all names in any origin claim.
The first three remaining entries in the combined volume ranking with lexical
ties are reserved on this branch before implementation:

- `no-invalid-regexp` (0 compiler, 0 repository)
- `no-label-var` (0 compiler, 0 repository)
- `no-misleading-character-class` (0 compiler, 0 repository)

This claim update is pushed before code. No shared harness or registration
generator edits. New Adamic files use .a.

## Current continuation status

`no-label-var` is implemented and independently compared on 77 compiler and
287 repository files, plus 34 supported positive/negative controls. One upstream
`undefined:` label fixture is refused by the shared native parser and is recorded,
not counted as covered. The new raw scope-symbol question is delivered inside
wave09_core/label_var/testdata and validated with temporary Go overlays; the
shared registration files remain untouched.

`no-invalid-regexp` and `no-misleading-character-class` are partial native ports.
Flag diagnostics and character-sequence judgments match production Go helpers;
the exact native pattern compiler, class parser, reference tracker and cooked/raw
mapping remain unresolved. These are not complete ports or full-corpus parity.
No additional rules are claimed. See ../wave09_core/REPORT.md.

The final origin refresh found a later duplicate reservation on wave 14:
350d4776 at 01:22:58 UTC. This wave's claim 0a4e2f17 was committed and pushed at
01:20:27 UTC and absent from all other claims in the immediate second refresh.
Both reservations and timestamps are saved for Ahra; this earlier claim remains.

## Regex continuation resumed

The native regex class parser, whole-pattern judgments, literal diagnostics and
suggestions, Go rune quoting and WTF-8 flag behavior are now ported and compared.
Native token diagnostics match Go, including suggestion edits; actual-source
literal comparisons pass on Node over both frozen corpora. The combined native
parser/listener build still times out at 120 seconds in the shared freshness
analysis, with captured stacks. The native pattern compiler and constructor
reference/constant/source-map adapters remain unfinished. No further rules are
claimed. See ../wave09_core/RESUME_REPORT.md for precise scopes and evidence.

## Current regex requirement, supersedes historical component status

The no-hand-rolled-matchers instruction is now the implementation constraint.
The earlier class scanner and regex rewrite helpers are historical isolated
component evidence, not the completion path for these rules. No further custom
matching, pattern parsing or rewriting is being added. Existing flag diagnostics,
scope checks, source mapping and constant-resolution work remain available.

Both regex rules remain incomplete. The requested dynamic JS RegExp constructor
path is blocked by native lowering's explicit "RegExp with a nonconstant
pattern" refusal, reproduced with dynamic_regexp_gap.a and a clean-running
sanitized static-pattern mutant. The named origin/codex/lint-regex shared table
was absent from all fetched origin heads at this check. The string-only parser
and absent numeric handed-node listener contract remain separate shared gaps.
No React analysis exception applies; these claims are not counted as finished
or parked under that exception. No new batch is claimed. See
../wave09_core/REGEX_POLICY_REPORT.md for current scope and retained evidence.

## Landing refresh on c01907a7

Rebased onto current main c01907a7 and reran all original rule, bridge, Node
and seventeen owned component checks, including their mutants and sanitizers.
The shared origin/codex/lint-regex branch now exists at 071fb012. Its 107-row
table has no fixed-pattern compile site for these three claimed rules. Its
own gaps report reproduces the same dynamic constructor refusal. This
supersedes the earlier observation that the branch was absent; the dynamic
RegExp and numeric handed-node blockers remain. No further matcher was added,
no React parking exception was used and no new rules were claimed. See
../wave09_core/LANDING_C019_REPORT.md for fresh verification evidence.

## Named harness review

ab70f38d4 on origin/lint-rules/harness adds the handed-node callback via
"node": true, along with reportNode/reportRange. It is not yet on current
main c01907a7. Its registry still requires string kinds and its parser exposes
string kinds, so the requested numeric rule.json declarations cannot be
decoded by that registry. The remaining speed blocker is the numeric contract
and type-aware checker-context integration, not an absent handed-node hook
on the named harness branch. The separate dynamic RegExp refusal remains.
No new claims or React parking. See ../wave09_core/HARNESS_AB70_REPORT.md.
