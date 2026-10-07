Built: rebased wave-30 onto current main, re-greened all twelve gates and audited the new regex requirement; no new claims or implementation changes.
Commits: previous pushed tip 70ffa3fb9, validated rebased tip cc2f5cd6e, main base c01907a70; 24 commits replayed without conflicts.
Commands and outputs: all twelve wave-30 gates PASS 407.595s; vet PASS; setup 36s, nproc 5; collection/discarded compiler timing 2.553932s native versus 0.768265s Go.
Mutants: all 24 successful-exit rule, listener, graph, state and component mutants are caught only by independent Go byte comparison on this base.
Not covered: complete new JSX rule ports/corpora/timings, parked React analyses, numeric handed-node dispatch, full repository gate or full oracle matrix.

Main advanced from f8013f0b to c01907a70. Landing-first therefore became this unit's
work. The clean rebase replayed all 24 branch commits and received main's stage3
work and dormant oracle fixture hook. No shared parser, harness, registration
file, compiler or runtime was changed or reverted by this unit. The expected
13-package leak-helper replacement is not present in this main update; nothing
was reverted to avoid it. Only codex/typeaware-wave-30 is pushed, with the exact
previous remote tip as its lease under the user's explicit rebase instruction.

The new regex instruction was audited against the Go source of all eight completed
behavior ports and the three newly claimed JSX rules, including context-value
stability and imported JSX/React utilities. There is no regexp import or compile
call in these files or utilities. The occurrence of the word regex in the context
rule describes a regular-expression literal as a construction value, not a matcher.
The name-classification predicate reproduces an ordinary Go byte/name predicate;
it does not replace a Go regex with a hand-written matcher. Numeric parsing and
quoted-message formatting likewise do not originate in regexp.

origin/codex/lint-regex at 071fb0128 provides the shared table. Its actual pinned
inventory contains 107 sites; no row belongs to the wave-30 rules. No new regex
literal or constructor migration is needed for this unit's current claims, and
no regex offset conversion was added. regex-audit.json records source hashes,
the cohere pin, the exact table blob and the zero matching rows. This audit does
not claim that arbitrary future option patterns or other workers' regex ports
are complete.

After the settled rebase, ADAMIC_TOOLS=/workspace/adamic-tools bash cloud/setup.sh
passes. Timing lines Go 0s, clang 0s, Node 0s, submodules 0s, cache warm 36s,
total 36s; nproc 5, cgroup quota 4 cores. Commands source the printed env.sh first
and write only to logs:

- go test ./stage1/cohere/typeaware -run '^TestWave30' -count=1 -timeout 30m -v.
  ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-30-typescript. Each WAVE_30, NEXT,
  THIRD and PROCESS compiler/repository manifest variable points to the frozen
  /workspace/wave-30-compiler.manifest or /workspace/wave-30-repository.manifest.
- go vet ./stage1/cohere/typeaware ./bridge/tsgo/...
- git diff --check; git ls-remote origin refs/heads/main refs/heads/codex/typeaware-wave-30.

All twelve gates execute and pass: numeric listeners, collection/discarded rules,
native process graph, process exit, blocking streams, parked React components,
React prerequisites, ISO/callback rules, timer rule, output state, new JSX
components and JSX prerequisites. The eight completed behavior ports match Go
findings/fixes/suggestions on controls and the frozen 77 compiler/287 repository
roots, with native sanitizer agreement and released-handle refusal checks. All
eight have zero findings on those corpora; positive controls exercise the reports.
No stage3 corpus expansion is represented by those frozen manifests.

The 24 mutants comprise eight listener substitutions, eight complete-rule
judgment/range changes, graph and catch-sensitive state changes, three parked
React-component changes and three newly claimed JSX-component changes. Every
mutated native process exits successfully and the independent Go comparator
catches the altered bytes; oracles.log records each first difference. The new
JSX components again match 1215 fragment bytes, 12037 component-name bytes and
46522 message bytes against production Go, sanitized native and emitted
JavaScript. Their mutants are caught at bytes 98, 22 and 13844 respectively.
These are component proofs, not full-rule proofs.

Whole-process measurements observed in this run:

| Runner | Corpus | Native seconds | Go seconds |
| --- | --- | ---: | ---: |
| collection/discarded | compiler | 2.553932 | 0.768265 |
| collection/discarded | repository | 0.349378 | 0.155405 |
| ISO/callback | compiler | 1.761549 | 0.286403 |
| ISO/callback | repository | 0.263750 | 0.135471 |
| timer | compiler | 1.607424 | 0.283832 |
| timer | repository | 0.285366 | 0.123321 |

These are single observations including checker load, not benchmark medians or
speedup claims. Complete JSX rule timings remain unavailable because the native
parser rejects each positive Go TSX control. JSX extraction, symbol scope/bare
fragment resolution and context memo-input stability remain unimplemented beyond
the previously reported components. rule.json metadata marks the three new rules
partial. Their normalized handed-fact components do not read kind strings or
fetch nodes; the actual shared numeric parser/visitor/registration interfaces
have not landed here. Legacy ports still await that dispatch integration.

The prior React HIR/SSA/capture claims stay explicitly PARKED as authorized,
counting as finished for the landing cap. The new JSX claims remain partial,
blocked on shared native JSX and handed-node interfaces. No further rules are
claimed in this landing unit. Full source-rule parity for those three, their
complete fixes/suggestions, arbitrary non-ASCII named construction formatting,
full repository gate and full oracle fixture matrix are not claimed. Unsupported
non-ASCII named construction messages continue to refuse explicitly with NotYet.

Logs, the 24-commit rebase mapping and regex provenance are preserved in
validation-wave-30-regex-landing. Existing exact byte-stream evidence remains
in the earlier validation directories. Bridge and compiler code are unchanged
from the previously validated base, so bridge behavioral tests and unrelated
full gates were not repeated. No new authored Adamic .ts files were added.
