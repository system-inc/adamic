Built step 1: catch bindings preserve a tagged thrown value and Error identity is tested at runtime.
Base 390af985caf2a81edb239cdf5e3238add50a7fcd; own branch codex/stricter-catch-variables.
Full affected compiler packages, complete uncached oracle, Linux counts and vet passed after the explicit-root fix.
Mutant treating every caught value as Error failed source-Node comparisons in both backends.
Step 2 property checks and the five checker-ledger sites remain to be implemented next.

The ledger is a1a16427a46149435e25f847c45ad136f1f1e55c. Its actual five sites are D108, commandLineParser.ts:2301:91; D128, program.ts:406:25; D132, program.ts:441:25; D152, sys.ts:1281:43; and D153, sys.ts's require result. The first three read e.message; D152 reads e.code; the require result returns e. Four sites are not all inside sys.ts's require.

The pinned base has MakeError and the runtime Error layout rather than error_classes.go. The Error object's name/message layout is unchanged. The exception carrier changes from adamic_object * to adamic_heap *, plus an independent pending flag. Undefined therefore remains NULL as a value while a separate flag records that it was thrown. Catch and finally transfer the carrier's owned count and clear or restore the flag together. Source Error identity is a reserved IR builtin identity, implemented by runtime shape identity natively and instanceof Error in JavaScript. It never treats an ordinary object with a message field as an Error.

catch_values.a is held to source Node, JavaScript output, native release, ASan, UBSan and LeakSanitizer. It throws Error, a built string, a number, an ordinary object and undefined; returns each caught value; checks type, identity and instanceof; and carries undefined through finally. Source Node prints true only for the Error, preserves every identity, and prints the finally line before the outer undefined catch.

The first counts and complete oracle runs uncovered an existing loader restriction on internal/fresh/testdata/regexp_tree.ts. The option audit already accepts production command-line roots, but projectOptionsForRoots refused explicit roots excluded by tsconfig first. Removing that obsolete restriction lets the existing audit check those roots. TestExplicitProjectRootIsAudited proves an ordinary wrong type in such a root is still rejected. The initial gate's only failures were this root rejection in the oracle and counts. The repeated complete uncached oracle passed after the fix.

Commands, with the setup environment sourced and full output in log files:

- bash cloud/setup.sh: 56.993s total. Go 0.049s; Node 0.048s; submodules 0.110s; markdown 0.132s; clang 0.339s; Go build 56.778s; cache warm 56.956s. nproc 5, cpu.max 400000/100000. /tmp/stricter-catch-setup.log.
- go test ./internal/lower ./internal/ir ./internal/javascript: the first run found obsolete number/string throw refusal expectations, replaced by the new differential fixture. The repeated package gate passed.
- ADAMIC_GATE_UNCACHED=1 go test ./internal/load ./internal/lower ./internal/ir ./internal/flow ./internal/fresh ./internal/native ./internal/oracle -count=1 -timeout 30m: load 21.666s, lower 94.048s, ir 4.353s, flow 262.672s, fresh 128.658s, native 502.561s. Initial oracle failed only the excluded explicit root. /tmp/stricter-catch-step1-gate.log.
- go test ./internal/load -count=1 after the root fix: 43.586s, pass. /tmp/stricter-catch-extra-roots.log.
- ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -count=1 -timeout 30m after the root fix: 193.498s, pass. /tmp/stricter-catch-step1-oracle-final.log.
- go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts: 95.751s, pass. /tmp/stricter-catch-step1-counts-final.log. Every changed row is in stricter-catch-counts.md.
- go vet of all affected packages and git diff --check: pass.

The independent mutant replaces the runtime Error identity result with true. The oracle exits 1: source Node exits 0 and prints false for the string; both generated backends print true then stop at the incompatible narrowed read. /tmp/stricter-catch-every-error-mutant.log. The compiler was restored before every final gate.

Only the feature branch is pushed. No main or area branch was pushed or merged into. The entire repository gate, full test262 corpus and stage 3 whole-compiler lowering are not claimed here.
