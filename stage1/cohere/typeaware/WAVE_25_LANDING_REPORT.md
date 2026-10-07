Rebased all 13 wave-25 commits onto main e011f8f6, preserving every patch; no new rules claimed.
Original pushed tip 16a717e8 became f0cfe281; this accompanying commit preserves the first landing gate and evidence.
Wave gate PASS 408.470s: 840 valid controls, 527 findings, 117 fixes; dependency PASS 542.007s over controls and both corpora, with sanitizers and released handles.
All 14 wave mutants and ten dependency mutants are caught by Go bytes; bridge fault injections and the uncached Node one-byte mutant also pass.
Main advanced to e8ba3d5d during validation; another rebase and gate are required before pushing. Boolean naming is still partial; no new claims.

## First landing attempt

The only branch pushed by this worker is codex/typeaware-wave-25. It was not on
main. Rebase onto e011f8f60899586d6373a5ccb07335ad82cfbf3c completed without
conflicts. Range-diff marks every one of the 13 carried commits equal.
The original pushed tip is 16a717e87985d6ed37411707d4c0a550e7198350; its
rebased counterpart is f0cfe2811ea3ce6043c5e117b822bb247c7d1520.
No shared harness, registration generator or protected compiler file was edited
by this landing work. Two own tests now explain why they run serially.

All five wave suites ran with the frozen 287-file repository manifest and
77-file TypeScript v6.0.3 manifest, normal and under ASan/UBSan/LSan. The 840
valid controls produce 527 findings and 117 ordered fixes, with no suggestions.
The fourth batch excludes the same three parse-invalid controls. The fifth
batch also compares nil boolean options, producing 84 findings instead of 141.
The carried ten-rule bridge dependency ran 36 controls with 53 findings,
6 fixes and 10 suggestions. Its repository corpus produces 180 findings;
its compiler corpus produces 16,589 findings, 2,222 fixes and 152 suggestions.
Every complete stream matches independent production Go cohere, preserving
ordered fixes, suggestions and duplicate findings. No source expectations are
imported into the oracle.

The complete bridge suite passes in 69.845s, checker in 0.242s, and Go vet
passes. The uncached filtered Node oracle passes in 7.913s on eight fixtures
plus TestTheOracleCatchesOneByte; cache statistics show zero hits. Released
handles refuse with panic 70. The bridge catches off-by-one input/output
lengths with ASan, removed frees and region heap allocations with LSan,
stale handles with assertions, wrong source positions with oracle bytes,
and removed link opt-in with its refusal check. Every individual wave and
dependency rule mutant and its first differing byte is in the retained logs.

Three isolated interleaved full-output samples of the fifth batch match bytes:
compiler native median 3.945301s vs Go 0.389238s (10.136x); repository native
0.561033s vs Go 0.135577s (4.138x). These measure load, queries, traversal and
serialization. Boolean naming declines with nil options on these corpora,
just as Go does; they do not measure its enabled mode.

Setup was rerun: go ready 0s; clang ready 0s; node ready 0s; submodules ready
0s; build cache warm 101s; done 101s on 5 processors, cpu.max 400000 100000,
17.6 GB. nproc is 5. Source /workspace/adamic-tools/env.sh for every build shell.
Go 1.27.1, clang 20.1.8, Node v24.19.0. All test output went directly to logs.
The exact gate commands, complete output hashes and compressed streams are in
validation-wave-25-landing, with source inputs and benchmark samples.

Current main now supports a constant new RegExp probe, which prints true.
A runtime-pattern probe still exits 1 at 3:31 with
`stage 0 can't lower RegExp with a nonconstant pattern yet`. This supersedes
the old fifth-batch observation that even the constant constructor was refused.
Boolean naming remains partial: arbitrary runtime regex options need compiler
support, while cross-file props annotations and typed wrapper component paths
remain unimplemented. Dedicated refusal tests exit 70 before findings for
those three paths. No new rules are claimed.

A final fetch found main had advanced to e8ba3d5d81de4d3773c723914fccd4c76248b965,
adding call-target routing, devirtualization and memory-analysis changes.
The completed first gate does not establish agreement with that compiler.
It is preserved here while the branch is rebased and revalidated again.
Not covered: full repository gate, nondefault option matrix, full shared-profile
integration, emitted-JavaScript comparison of these rule ports and CLI fixes.
