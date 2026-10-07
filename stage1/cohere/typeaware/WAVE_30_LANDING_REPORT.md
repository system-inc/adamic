Built: rebased the single pushed wave-30 branch onto current origin/main; existing complete ports and partial React components re-greened, with no new claims.
Commits: prior pushed tip 5cc0b875; validated rebased code tip da9164d5; base e8ba3d5d; landing evidence committed separately before the lease-protected push.
Commands and outputs: all nine wave-30 gates PASS 392.386s; inherited controls/corpora PASS 567.700s; bridge PASS 65.989s; vet PASS; stable setup 114s, nproc 5.
Mutants: eight complete-rule comparisons, graph/state comparisons and three React-component comparisons all caught successful-exit mutants; earlier inherited mutant proofs also pass.
Not covered: complete React analyses, full repository gate or full oracle fixture matrix; no new claims while these remain partial.

Only codex/typeaware-wave-30 was pushed by this unit. It was not on main, so
landing-first became this unit's work. No PR was opened and no other branch was
rewritten. Rebase authorization came from the user's explicit instruction; the
push uses the exact prior wave-30 remote tip as its lease.

The first clean rebase targeted e011f8f6 and completed 15 commits without conflicts.
Its nine wave-30 gates passed in 506.359s, the inherited five-test run passed in
1116.394s, the changed six-rule integration oracle passed in 218.940s, and the
bridge passed in 149.365s. The first setup cache-warm step overlapped that rebase
and failed to compile lint_test.go at lines 312/316 with undefined volumeGenerated
and checkRecoveryRefusal. Both helpers were present after the checkout settled;
rerunning setup on the stable tree passed. Tool timing lines: Go 0s, clang 0s,
Node 0s, submodules 0s, cache warm 114s, total 114s, nproc 5, quota 4 cores.
The failure log and stable rerun are both preserved.

Main then advanced to e8ba3d5d with devirtualization and call-target changes in
the native compiler. The earlier results were not used to claim readiness
against that new compiler. After all earlier processes finished, a second clean
rebase replayed the same 15 commits. All nine wave-30 gates, both inherited
compiler/repository comparisons and bridge checks were rerun on that base.
The remote main tip was verified again after those gates and still equals e8ba3d5d.

Current-main commands, after source /workspace/adamic-tools/env.sh, output to logs:

- go test ./stage1/cohere/typeaware -run '^TestWave30' -count=1 -timeout 30m -v,
  with every WAVE_30, NEXT, THIRD and PROCESS compiler/repository manifest
  variable set to the frozen corpora, ADAMIC_TYPESCRIPT_SOURCE pointing to the
  TypeScript 6.0.3 checkout, and per-suite artifact roots in wave-30-landing-final.
- go test ./stage1/cohere/typeaware -run '^Test(CoverageAgreementAndMutants|VolumeAgreementAndMutants)$/^$' -count=1 -timeout 30m -v,
  with VOLUME and COVERAGE artifact/manifest variables and the same compiler pin.
  The leaf selector intentionally skips inherited nested mutants on this second
  pass; their complete first-pass proofs are preserved. Parent control, corpus,
  sanitizer and released-handle bodies run and pass. Go prints a no-tests-to-run
  leaf warning despite those executed parent comparisons; the logs record both.
- go test ./bridge/tsgo/... -count=1 -timeout 30m; go vet ./bridge/tsgo/... ./stage1/cohere/typeaware.
- git diff --check; git ls-remote origin refs/heads/main refs/heads/codex/typeaware-wave-30.

Every complete wave-30 rule again agrees byte for byte with production Go on its
controls, findings, fixes and suggestions, the 77-root compiler corpus and the
287-root repository corpus. All eight rules have zero findings on those two
frozen corpora; positive controls establish the nonzero judgments. Sanitized
native runs agree and released handles refuse with the required panic text and
exit 70. The inherited coverage runner matches 16589 compiler findings and 180
repository findings; the inherited volume runner matches 14232 compiler findings
and 47 repository findings, with complete serialized bytes and sanitizer runs.

Current-main comparison mutants, all native exit 0:

| Check | Changed decision/output | First differing byte |
| --- | --- | ---: |
| collection misuse | collection boundary | 403 |
| discarded outcome | finding end | 26916 |
| discarded pure result | finding end | 1092 |
| process graph | throwable fork | 916 |
| process exit after output | empty write state | 54 |
| blocking standard streams | empty write state | 83 |
| refs component | convergence compares ref identity | 86865 |
| purity component | container-kind guard reversed | 1368 |
| manual memo component | pruning applied to value condition | 435 |
| ISO date cut | finding end | 58 |
| callback parse try | finding end | 3968 |
| uncleared race timeout | finding end | 62 |
| process output state | catch-sensitive state | 5 |

The React component gate still compares 279296 ref-lattice bytes, 6523 purity
property bytes and 900 memo-scope bytes against actual Go helpers, sanitized
native and emitted JavaScript on Node. These remain partial components, not full
rules. The prerequisite probe again confirms Go reports on three valid TSX
controls while the native parser rejects two and parses the third as a type
assertion rather than JSX. Native React HIR/SSA and the reactive memoization
pipeline are still missing from this branch. JSX support exists on the separate
stage1-jsx-lint branch but requires shared integration. No further rules were
claimed under the landing cap or while these three retained claims remain partial.

Single measured whole-process comparisons from the current-main gates follow.
They include checker load and competed with inherited validation; they are
observations, not quiet benchmark medians or a speedup claim:

| Runner | Corpus | Native | Go |
| --- | --- | ---: | ---: |
| initial two rules | compiler | 2.003110s | 0.321438s |
| initial two rules | repository | 0.278235s | 0.133909s |
| next three rules | compiler | 2.828645s | 0.869215s |
| next three rules | repository | 0.410034s | 0.165438s |
| race timeout | compiler | 1.690133s | 0.346149s |
| race timeout | repository | 0.265250s | 0.139682s |

All implementation changes from this unit remain in their existing files;
no conflict resolution or new shared-file edit was necessary. The shared
suite_test.go nominal-clone fix was received from main, and its six-rule oracle
was checked in the earlier pass. Whole-rule React timing and diagnostic parity
remain unavailable until their missing analysis inputs exist.

validation-wave-30-landing contains logs, exact gzip streams, SHA-256 indexes,
test summaries and the 15-commit old/new mapping. To make room for the final pass,
1121 earlier-base streams were preserved before deleting only 81 verified ELF
executables/ar archives from the completed earlier landing scratch directory,
freeing 3064004206 bytes. Authored source and active final-pass outputs were kept.
Prior reports' historical SHAs remain linked by rebase.json to current commits.
