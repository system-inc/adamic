635f2410: lower optional void calls for effects and pass proven boxed union arguments through optional method adapters.
Commits: 0ac4e2c6, 2c7e44ea, a2e3d852, 3f620860, e91cbdf6, 9fe50a38, 635f2410; each pushed to codex/notyet-void-or-undefined.
Validation: focused oracle/lower/native tests and counts refresh passed; positive fixtures agree with Node in release native, sanitized native and JavaScript.
Mutants: all nine final mutants failed through their intended fixture assertions, with production sources restored.
Limits: erased void values, unspecialized T, union results and callable-field union parameter declarations remain refused or NotYet.

The requested compiler-area base resolved to b410340dc8f889b5799c3bc519117c63def3aa24. Replay dependency resolved to 9a1f14c5d994aa855625e7cfa295677060348fec and was merged. The specific compiler-area instruction was taken over the generic main-base instruction. No PR was opened; no main or area branch was pushed.

Optional `void | undefined` is recognized only when every member is void or undefined. Statements retain the call and its argument effects. Optional absent receivers skip argument evaluation. This does not prove the value of an erased void callable: Node reports `number` when a callback returning 42 is stored in a `() => void` field, and `undefined` with an absent receiver. Returning undefined for both would silently miscompile. Value use keeps `a void call used as a value`; it needs actual return provenance or a language ruling forbidding that observation.

Union caller slots require one checker signature and a proven Union parameter. Existing IR boxes and native `adamic_value.reference` already represent these arguments. Optional callees require nonnullable signature lookup. The native method table now admits Union parameters, using its existing thunk and ownership handling; Union results remain excluded. No IR, JavaScript, flow walker or runtime C change was needed. New runtime helpers: none.

Two positive union fixtures cover string/object and number/object/undefined arguments, present/absent receivers, and direct primitive arguments requiring fitting. The two original callable-field fixtures reach `a function value taking Name` and `a function value taking Location`. Branch history identifies `functions.go:signature` as changed in fd0e1225 on codex/notyet-generic-returns-t. Its one-line parameter exemption was requested but not approved during this work, so that function remains untouched. A named function converted to a closure also retains its existing parameter guard. This is an ownership limit, not a language-design refusal.

The generic fixture instantiates a callable return T as string and number and agrees with Node. The census nested function has no concrete instantiation/ABI in standalone replay. Its refusal remains rather than guessing a representation. Removing it requires specialization context, not accepting an unrepresented T.

Raw-count observation from codex/stage3-notyet-table at 57b9777c8eb4ee28b1f50220e8c8fb51a2dfadf7, stage3/notyet-table/roots/raw.csv:

| Exact reason | CSV rows | Distinct root sites | Change coverage |
| --- | ---: | ---: | --- |
| a call returning void \| undefined | 46 | 40 | Statement return rule covers all 40; other stops can precede it. |
| passing union of differently held members to a function value | 23 | 23 | Caller rule supports proven boxed slots; method adapters pass, callable-field creation still has the next guard. No claim that all 23 complete lowering. |
| a call returning T | 2 | 1 | 0 census roots removed; concrete instantiations are verified. |

Counts deduplicate `(kind, where, reason, text)` rather than attempts. These are rule counts, not a fresh full-project census.

The same replay command was run before and after, with output redirected to JSON and a log:

```bash
source /workspace/adamic-tools/env.sh
go run ./stage3/census/latent/replay \
  -project /tmp/notyet-void-adapted/src/tsc/tsc.ts \
  -where /tmp/notyet-void-adapted/src/compiler/binder.ts:587:13 \
  -kind NotYet -reason 'a call returning void | undefined' \
  > /tmp/notyet-call-final-binder.ts-587-13.json \
  2> /tmp/notyet-call-final-binder.ts-587-13.log
```

| Site | Before | Final |
| --- | --- | --- |
| binder.ts:585:13 | Earlier `a value of type Path` at 585:73 | Same argument stop; requested return stop not reached. |
| binder.ts:587:13 | Exact requested return stop reproduced | No finding/boundary at 587:13; statement passes that stop. Enclosing unit retains other findings. |
| checker.ts:3467:42 | Earlier `a value of type __String` at 3464:78 | Same signature stop. |
| checker.ts:3494:42 | Earlier `a value of type __String` at 3492:74 | Same signature stop. |
| watchUtilities.ts:540:22 | Exact `a call returning T` reproduced | Same stop reproduced. |

Replay exit 1 means the requested signature did not reproduce; watchUtilities exits 0. Replays are observations on a checker-rejected entry-root program, not a whole-program compilation claim. Adaptation command was `bash stage3/apply.sh /tmp/notyet-void-adapted`, exit 0.

Final validation commands (all output redirected to the named log, never piped):

```bash
source /workspace/adamic-tools/env.sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/notyet_|TestOptionalVoidValueStillNotYet|TestUnionCallableSourceObservations' -count=1 -v > /tmp/notyet-call-final-fixtures.log 2>&1
go test ./internal/lower -run 'TestCensusOptional|TestLibraryMethodValue|TestClassInheritance' -count=1 > /tmp/notyet-call-lower.log 2>&1
go test ./internal/native -run 'TestUniformFieldsMatchNode|TestDevirtualizedCalls|TestExactReceiverRejectsAssignments' -count=1 > /tmp/notyet-call-native.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts > /tmp/notyet-call-counts.log 2>&1
python3 internal/oracle/testdata/run-notyet-call-mutants.py > /tmp/notyet-call-mutants.log 2>&1
```

Final oracle: PASS, 1.193s, native hits 0/misses 15, Node hits 0/misses 15. Lower: PASS 0.524s. Native: PASS 0.563s. Counts: PASS 30.255s. Five positive fixtures exercise both backends; three negative fixtures plus value-use and independent Node observations preserve boundaries. Counts were refreshed. No whole-package run or full gate was run, as instructed.

The reusable mutant runner restores expression.go, optional_void.go and emit_objects.go in `finally`. Each mutant is independent and must compile; a build failure is rejected as evidence. Logs are under /tmp/notyet-call-mutants/.

| Mutant | Fixture/assertion that caught it |
| --- | --- |
| reject-void-union | push/pop fixtures fail Lower. |
| drop-void-effects | push/pop Node comparison reports stdout differs. |
| accept-void-value | value-use test loses its named stop; Node observed number/undefined. |
| reject-union-argument | optional method adapter fixtures fail Lower. |
| omit-union-box | direct numeric argument in adapter_location produces native/Node exit-code disagreement. |
| nullable-signature | optional method adapter fixtures fail Lower. |
| ignore-generic-substitution | concrete T fixture fails with `a call returning T`. |
| accept-mixed-void | number/void/undefined negative fixture unexpectedly lowers. |
| omit-union-adapter | native reports the checker-proven method missing while Node succeeds. |

Each final mutant exited 1. Earlier development experiments included the same missing-native-adapter failure; it motivated the native edit. An initial reject-union mutant had an unused loop index and was corrected before being counted. A redundant hasVoid mutant survived because member filtering already excludes nonvoid unions; it is not counted as a rule or a killed mutant. An initial invocation used the wrong Go binary before sourcing the environment and was rerun correctly.

Setup used `export GOPROXY='https://proxy.golang.org|direct'`, then `bash cloud/setup.sh > /tmp/notyet-void-setup.log 2>&1`, then the environment above. Setup exited 0. Timing lines: Node ready 0.050s; Go ready 0.066s; clang ready 0.453s; markdown installed 1.041s and ready 1.203s; submodules ready 15.976s; go build 224.507s; test binaries deferred 224.618s; warm 224.619s; done 224.650s. `nproc` = 5; CPU quota = 4. Tool setup printed /workspace/adamic-tools/env.sh. No setup workaround was needed.
