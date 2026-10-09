Built one candidate from 339c86b3 by merging origin/compiler/fx6-valid-programs f4de16fe.
Merge commit e36095a7 preserves element-key refusal and receiver-type/member registration.
Full lower PASS 94.916s; TestCallTargetReaders PASS 15.756s; counts refresh PASS 130.083s.
All three revert mutants compile and fail their intended behavioral/refusal assertions.
p70 remains outside this candidate and is handled next on compiler/fx6-valid-programs.

Conflict resolution changed only the comment hunk in internal/lower/interface_cast.go. Both refuseViewElementReads calls and receiver-keyed viewSchema registration remain. Counts were regenerated and have no diff.

Commands: go test ./internal/lower -count=1 -timeout 180s; go test ./internal/ir -run TestCallTargetReaders -count=1 -timeout 90s; go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 4m -args -update-counts. Each has a hard outer timeout and output in its named log here.

Run timeout 600 python3 review/compiler/fx6-candidates/run-mutants.py with the toolchain environment sourced. It restores all sources in finally blocks. revert-receiver.diff loses p37's second stdout line (0.13s). revert-destructure.diff loses p53's n5 output (0.18s). revert-element-refusal.diff makes TestViewStringElementReadRefused fail with got <nil> (0.04s). All mutant tests exit 1 after compiling. The initial receiver mutant exposed an unused loop key; its corrected version removes that key and is caught by Node disagreement, not a build error.

The original refusal runner was also run: timeout 400 python3 review/compiler/fx6-key-read/run-mutant.py. Both its refusal assertion and native mismatch witness exit 1. p05 source Node prints NaN/false while release native prints a run-specific tiny number/true; both programs exit 0. See element-native-disagreement.log. The temporary test is removed. Restored focused lower fixtures and element-read tests pass in 0.814s.

Setup: export GOPROXY='https://proxy.golang.org|direct'; timeout 600 bash cloud/setup.sh. Its initial build encountered unresolved merge markers; after resolution setup passed. Timing lines: Go ready 0.078s; Node ready 0.094s; submodules ready 0.190s; markdown ready 0.272s; clang ready 0.428s; Go build ready 73.508s; done 73.681s. Source /workspace/adamic-tools/env.sh. nproc=5, quota=4 CPUs. Lane checks run on the committed candidate before push, with their output recorded separately. No full repository gate was run.
