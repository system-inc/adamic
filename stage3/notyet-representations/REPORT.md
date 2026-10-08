f78fa9a5e978149a168afdaaf999e53e853a011d: Prepared replay tooling and proved the unknown null/undefined blocker; no lowering rule added.
Base b410340dc8f889b5799c3bc519117c63def3aa24; merged replay 9a1f14c5d994aa855625e7cfa295677060348fec.
Setup exited 0; source hashes matched 81/81; final replays exited 1, 1, 0; reduction compilation exited 1.
The unknown-to-Union scratch mutant built with ASan/UBSan and was caught by native stdout differing from Node.
Covered root sites: 0/299; all 116 kinds are recorded as skipped in status.csv; no registered fixture or production change.

The unit-specific area/compiler base takes precedence over the generic main-base instruction. The only tracked changes before this report are the explicitly requested replay-tooling merge. Compiler and backend sources were not changed. No claim is made that old census signatures which do not reproduce are fixed or proven echoes.

The exact adaptation snapshot was reconstructed from origin/codex/stage3-latent-full. Every file in the recorded 81-file tsc source manifest matched both byte length and SHA-256 after apply completed. An initial replay started before adaptation completed; its result was discarded and both examples were replayed after the hash check. Final observations are preserved in evidence/first-final.json, second.json and unknown.json.

The largest kind, SolutionBuilderState<T> (22 listed roots), did not reproduce at either example. At tsbuildPublic.ts:555:43 the selected unit stopped at `a function returning Path` at 555:10. At 608:53 the selected unit instead reported unsupported Map keys at 609:28 and 610:28, an array of ResolvedConfigFileName at 614:24, a BinaryExpression with a value and a value at 619:23, and a nested closure at 622:5. These are observations on the required current base, not proof of successful whole-program lowering.

The next kind, unknown (20 listed roots), reproduced at binder.ts:495:50. Its reduction is evidence/unknown_null_undefined.a. Node prints `object` and `undefined`; the production compiler stops at the unknown parameter at 1:19. Unknown is in the language's type list; this is an implementation gap, not a design refusal needing a ruling.

A scratch Go overlay added only `if proven.Flags()&checker.TypeFlagsUnknown != 0 { return ir.Union, true }` at the start of representation. The mutant built successfully with --sanitize, ran with exit 0 and no sanitizer diagnostics, and printed `undefined` twice. The Node stdout comparison killed it. Existing native union storage uses NULL for undefined; native null emission also uses NULL, and union typeof uses a static boolean to interpret that pointer. A dynamic unknown parameter can hold either, so this rule loses information. Exact support requires changes to IR/native runtime representation, conversions and observations, outside this unit's typeOf territory and minimal shared-hook allowance. No such changes were retained.

The generated JavaScript initially could not resolve its unused adamic runtime import. For the standalone reduction only, that unused import line was removed from a scratch copy; Node then printed the correct `object` and `undefined`. This is not a registered oracle run and is not counted as completed fixture validation.

Commands run (all outputs written directly to logs):

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/notyet-representations-setup.log 2>&1
source /workspace/adamic-tools/env.sh
bash /tmp/notyet-representations-input/stage3/apply.sh /tmp/notyet-representations-adapted > /tmp/notyet-representations-apply.log 2>&1
go run ./stage3/census/latent/replay -project /tmp/notyet-representations-adapted/src/tsc/tsc.ts -where /tmp/notyet-representations-adapted/src/compiler/tsbuildPublic.ts:555:43 -kind NotYet -reason 'a value of type SolutionBuilderState<T>' > /tmp/notyet-representations-first-final.json 2> /tmp/notyet-representations-first-final.log
go run ./stage3/census/latent/replay -project /tmp/notyet-representations-adapted/src/tsc/tsc.ts -where /tmp/notyet-representations-adapted/src/compiler/tsbuildPublic.ts:608:53 -kind NotYet -reason 'a value of type SolutionBuilderState<T>' > /tmp/notyet-representations-second.json 2> /tmp/notyet-representations-second.log
go run ./stage3/census/latent/replay -project /tmp/notyet-representations-adapted/src/tsc/tsc.ts -where /tmp/notyet-representations-adapted/src/compiler/binder.ts:495:50 -kind NotYet -reason 'a value of type unknown' > /tmp/notyet-representations-unknown.json 2> /tmp/notyet-representations-unknown.log
node --input-type=module-typescript < /tmp/notyet-representations-probes/unknown_null_undefined.a > /tmp/notyet-representations-node.log 2>&1
go run ./cmd/adamic c /tmp/notyet-representations-probes/unknown_null_undefined.a > /tmp/notyet-representations-probe.log 2>&1
go run -overlay=/tmp/notyet-representations-probes/overlay.json ./cmd/adamic build /tmp/notyet-representations-probes/unknown_null_undefined.a -o /tmp/notyet-representations-probes/unknown-mutant --sanitize > /tmp/notyet-representations-mutant-build.log 2>&1
/tmp/notyet-representations-probes/unknown-mutant > /tmp/notyet-representations-mutant-native.log 2>&1
go run -overlay=/tmp/notyet-representations-probes/overlay.json ./cmd/adamic js /tmp/notyet-representations-probes/unknown_null_undefined.a > /tmp/notyet-representations-probes/unknown-mutant.js 2> /tmp/notyet-representations-mutant-js-build.log
node /tmp/notyet-representations-probes/unknown-mutant-standalone.js > /tmp/notyet-representations-mutant-js-standalone.log 2>&1
```

Setup timings: Node 0.057s; Go 0.080s; clang 0.491s; markdown dependencies 1.231s; submodules 17.582s; Go build 221.418s; cache warm 221.514s; total 221.547s. nproc=5, cgroup cpu.max=400000 100000. The printed /workspace/adamic-tools/env.sh was sourced. A merge attempted during submodule setup met an index lock; retry after setup released the lock merged cleanly without removing the lock.

No whole-package tests, full gate or counts refresh were run. No fixture was added to the oracle registry or testdata; the reduction remains evidence because the production compiler cannot lower it. The remaining 114 kinds were not replayed or implemented after this territory blocker. status.csv accounts for every supplied kind and its listed count, separately from the zero actually covered roots. No kinds are declared lowered, refused for a ruling, or cancelled as census echoes.
