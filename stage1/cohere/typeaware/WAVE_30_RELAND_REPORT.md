Built: cleanly rebased wave-30 onto current origin/main and re-greened all its oracle gates; no new claims or implementation changes.
Commits: previous pushed tip 990e54b8, validated rebased tip 609b0876, current main base f8013f0b; 19 commits mapped in validation-wave-30-relanding/rebase.json.
Commands and outputs: all ten wave-30 gates PASS 408.291s; bridge PASS 73.438s; fifteen filtered Node oracle fixtures PASS 14.292s; vet PASS; setup 85s, nproc 5.
Mutants: eight listener mutants and thirteen rule, graph, state and React-component mutants again exit successfully and fail independent Go byte comparison.
Not covered: complete React analyses, numeric handed-node dispatch, full repository gate or full oracle fixture matrix; no further claims while retained rules remain partial.

Landing-first is this unit's work because main advanced from e8ba3d5d to
f8013f0b. The branch replayed 19 commits without conflicts. The new main changes
lowering, Map/Set runtime behavior, inheritance and narrowing; old native results
were not used as evidence against this new compiler. Source implementations and
shared files were not edited to resolve the rebase. Only codex/typeaware-wave-30
is pushed by this unit, using the exact prior remote tip as a force-with-lease
condition under the user's explicit rebase authorization. Main and area/ branches
are never push targets.

After the settled rebase, ADAMIC_TOOLS=/workspace/adamic-tools bash cloud/setup.sh
passes. Timing lines: Go 0s, clang 1s, Node 1s, submodules 1s, cache warm 85s,
total 85s; nproc 5, cgroup quota 4 cores. Test commands source
/workspace/adamic-tools/env.sh first and write exclusively to logs:

- go test ./stage1/cohere/typeaware -run '^TestWave30' -count=1 -timeout 30m -v.
  All ten gates execute. ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-30-typescript;
  each WAVE_30, NEXT, THIRD and PROCESS compiler/repository manifest variable
  points to /workspace/wave-30-compiler.manifest or /workspace/wave-30-repository.manifest.
  Neither corpus is skipped: 77 compiler roots and 287 repository roots.
- go test ./bridge/tsgo/... -count=1 -timeout 30m.
- go vet ./bridge/tsgo/... ./stage1/cohere/typeaware.
- go test ./internal/oracle -run '^TestNativeAgreesWithNode$/internal/oracle/testdata/(047cb0d_n_.*|library_map_set_iterator_(number_hash|exhausted).*|override_same_representation.*)$'
  -count=1 -timeout 30m -v. All fifteen selected fixtures pass. This ordinary
  worker gate uses the permitted cache: native hits 11/misses 30, Node hits
  0/misses 30. It is not the full uncached integration gate.
- git diff --check; git ls-remote origin refs/heads/main refs/heads/codex/typeaware-wave-30.

All eight completed behavior ports still match Go cohere's findings, fixes and
suggestions byte for byte on controls and both corpora, with sanitized native
agreement and released-handle refusal checks. All eight have zero findings on
these frozen corpora; positive controls prove their nonzero judgments. The
blocking-stream gate checks 64 programs, 30 reporting; process-exit controls
report 39 findings. The native graph and catch-sensitive state gates also pass.
Listener output matches 348 Go bytes read independently from production listener
maps. React components match 279296 refs bytes, 6523 purity bytes and 900 memo
scope bytes against Go, sanitized native and emitted JavaScript on Node.

All twenty-one successful-exit comparison mutants are recorded in oracles.log.
The eight listener mutants change the first numeric subscription and are caught
at bytes 37, 86, 123, 168, 209, 251, 296 and 346. Existing mutants are collection
boundary (430), outcome finding range (27672), pure finding range (1119), throwable
graph fork (988), process-exit output state (78), blocking-stream output state
(107), refs convergence identity (86865), purity container guard (1368), memo
pruning condition (435), ISO range (80), callback range (4166), timer range (89)
and catch-sensitive state (5). Each compiled native mutant exits 0 and only
comparison catches the changed judgment/output.

Whole-process timings observed on this base, including checker program loading:

| Runner | Corpus | Native seconds | Go seconds |
| --- | --- | ---: | ---: |
| collection and discarded rules | compiler | 2.686261 | 0.791230 |
| collection and discarded rules | repository | 0.378809 | 0.157618 |
| ISO and callback | compiler | 1.760513 | 0.288526 |
| ISO and callback | repository | 0.262526 | 0.124392 |
| race timeout | compiler | 1.581732 | 0.282016 |
| race timeout | repository | 0.253282 | 0.118169 |

These are single observations, not benchmark medians or speedup claims.

Shared prerequisites remain unchanged on this main base. ParseNode exposes only
kind: string; the handed-node numeric driver is absent. Numeric declarations are
present, but legacy implementations still use string kinds and node retrievals;
full compliance with the dispatch speed requirement is not claimed. The rule.json
registry remains on the separate lint-harness-dot-a branch, whose current visitor
signature takes indexes. No incompatible descriptor or shared registration edit
was introduced.

The retained complete React rules remain blocked by native JSX and native React
HIR/SSA/reactive memoization infrastructure. The renewed prerequisite probe shows
Go reports on valid TSX while native purity parsing fails at DotToken 38, manual
memo parsing fails at Identifier 118, and refs JSX becomes a TypeAssertionExpression
with no JSX node. The passing probe reproduces missing prerequisites; it does not
prove full-rule parity. No new claims are made and these three claims are not
presented as completed ports. The shared batch-8 finding-model integration is not
part of this main update.

To make room for the new validation, only 73 verified ELF executables/ar archives
were deleted from the completed /workspace/wave-30-landing-final scratch tree,
freeing 2730528039 bytes. Source, logs, output streams and the committed earlier
full byte evidence remain intact. This pass preserves logs, checksums and the
19-commit rebase mapping. Previous reports cite historical SHAs; the mapping makes
their new rebased identities explicit. No shared parser, generator, harness or
protected compiler file was edited by this unit.
