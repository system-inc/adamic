Built: no-uncleared-race-timeout in .a, three isolated Go fact modules and Adamic decoders; registration pending, other two rule ports unfinished.
Commits: continuation claim 3fc98576; implementation checkpoint f8ef4017bc3ac617b3bd42668b163bbb8aa72a37; evidence follows separately.
Commands and outputs: timeout agreement PASS 95.918s, checker package PASS 0.254s, repository vet PASS, both pending decoders type-check.
Mutants: timeout judgment caught only by Go bytes at 45; ancestry-global, callee-return and module-target caught by checker contracts; released-registry caught by required panic.
Uncovered: production registration, the two stream-rule algorithms, their findings/fixes/suggestions comparisons and mutants, emitted-JavaScript comparison, DOM-specific timeout fixture and the full repository test suite.

## What this checkpoint proves

The original three rules remain complete and pushed as recorded in WAVE_19_REPORT.md.
No new rules were claimed in this turn. The continuation claim is unchanged.
WAVE_19_NEXT_REPORT.md records the earlier blocked checkpoint; this report records
subsequent implementation and testing without editing shared files.

no_uncleared_race_timeout.a implements the production rule's default-library
Promise/race checks, ambient-global timer recognition, const-bound timeout
promises, executor-local traversal and lost-handle judgment. It distinguishes
unread bindings from real reads by checker binding identity, including shorthand
reads, plain assignment targets, shadows, retained handles and nested callbacks.
Diagnostics preserve the exact pinned policy message and byte ranges. This rule
has no fixes or suggestions; both zero counts are serialized and compared.

The three isolated raw questions are:

- declaration-ancestry: checker symbol identity and flags, declaration source
  metadata, complete ancestor kind/name/span/flags and global-augmentation flags.
- resolved-callee: resolved signature return flags and declaration/body metadata,
  including the declaration source text for native analysis of imported callees.
- program-modules: checker program source files and raw import specifiers with
  resolved targets, including type-only imports, re-exports, import-equals,
  dynamic imports and require calls. Native code must compute any closure.

Each question has its own Go and Adamic file. None calls cohere lint decisions.
The callee/module decoder files type-check, but are not yet exercised by complete
native stream-rule implementations. Direct Go contracts check raw facts against
the actual checker APIs. Their mutants compile and fail those contracts, rather
than being counted as lint-rule byte mutants.

## Registration boundary

The shared registration generator, test harness and existing facts.go are
unchanged. Ahra's correction says to keep changes in owned rule files and to
stop on other blockers rather than editing shared files.

The timeout test builds a temporary Go overlay with one dispatch case to make
its isolated fact module callable. The source-driven test clearly names this as
pending registration and saves the overlay. This proves the implementation in
that test build, not successful production integration. A second archive built
without the overlay exits 70 with `unsupported checker question:
declaration-ancestry` on the same controls. This exact blocker is tested, not
inferred or hidden by skipping findings.

wave_19_next_registration.patch is the concrete, unapplied three-line dispatch
change needed for the fact modules. `git apply --check` passes. It has not been
applied to the working tree. The process-exit and blocking-streams algorithms
remain unfinished beyond their prepared facts; no claim of their parity is made.
No additional batch was claimed while these rules remain pending.

## Byte agreement and sanitizers

The independent oracle invokes unchanged production cohere rule code from the
same pins used by the original wave: cohere 715ba94f3608a6500086b1076ce5cb7e51b836db,
TypeScript-go 8d550c837c90bd1805b047b7eeccc2baac2d5e7a, compiler corpus TypeScript
050880ce59e30b356b686bd3144efe24f875ebc8. It imports no new bridge implementations.

| Population | Inputs | Findings | Identical bytes | Normal and ASan/UBSan/LSan |
| --- | ---: | ---: | ---: | --- |
| Controls | 49 | 24 | 15,417 | PASS |
| Compiler frozen roots | 77 | 0 | 5,241 | PASS |
| Repository frozen roots | 287 | 0 | 18,485 | PASS |

The controls include 31 targeted sources and all 18 reconstructed pinned
production table rows. Only their source lines are extracted, not expected
verdicts. Tests include Node-style timer declarations in a module's declare global,
merged setTimeout namespaces, expression executors, const timeout promises,
void/dropped/unread handles, retained/cleared/chained handles, nested callbacks,
shadows and nondefault Promise classes. The pinned DOM-only direct test and the
dynamically generated PhiSocial source are not separately replayed. Findings,
message bytes, ranges and zero fix/suggestion counts are compared in full.

The timeout mutant suppresses its lost-timer judgment. It compiles and exits 0
with empty stderr; only Go comparison catches byte 45. The ancestry-global
mutant returns false for global augmentation, callee-return returns zero flags,
and module-target drops resolved targets. All three compile and fail their
specific direct-checker contracts. Querying a released handle is refused;
retaining the registry entry makes the mutant exit 0 and fails the required
panic check. The healthy unregistered-archive refusal is also asserted.

## Time against Go

After verification, three alternating count-only rounds used the unchanged
volume_bench.py. Full streams had already matched above. These measure the
one-rule native binary built with the ancestry registration overlay.

| Median process | Go | Native | Native / Go |
| --- | ---: | ---: | ---: |
| Compiler | 0.385141s | 2.161141s | 5.611x |
| Repository | 0.163877s | 0.321178s | 1.960x |

These corpora have zero findings and the timeout rule makes zero checker queries
there. The measurements include loading, parsing and traversal, not findings
throughput. Native is slower. Raw rounds and phase timings are saved.

## Reproduction and evidence

Source /workspace/adamic-tools/env.sh as in the original successful 137s setup
(nproc 5, four-core quota). All test output is redirected to logs:

```bash
ADAMIC_WAVE19_TIMEOUT_ARTIFACTS=/tmp/wave19-timeout-final ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave19-typescript go test ./stage1/cohere/typeaware -run '^TestWave19TimeoutPendingRegistration$' -v -count=1 -timeout=20m > /tmp/wave19-timeout-final.log 2>&1
go test ./bridge/tsgo/checker -v -count=1 > /tmp/wave19-next-checker-all.log 2>&1
python3 stage1/cohere/typeaware/testdata/prove_wave_19_next_facts.py /tmp/wave19-next-fact-mutants > /tmp/wave19-next-fact-mutants.log 2>&1
go vet ./... > /tmp/wave19-next-vet-all.log 2>&1
/tmp/wave19-timeout-full/adamic types stage1/cohere/typeaware/resolved_callee.a > /tmp/wave19-next-resolved-types.log 2>&1
/tmp/wave19-timeout-full/adamic types stage1/cohere/typeaware/program_modules.a > /tmp/wave19-next-modules-types.log 2>&1
python3 bridge/tsgo/profile/volume_bench.py /tmp/wave19-timeout-final/timeout /tmp/wave19-timeout-final/timeout-oracle /tmp/wave19-timeout-bench --corpus compiler /workspace/wave19-typescript/src/compiler/tsconfig.json /tmp/wave19-timeout-final/compiler.manifest --corpus repository /workspace/adamic/tsconfig.json /tmp/wave19-timeout-final/repository.manifest > /tmp/wave19-timeout-bench.log 2>&1
```

validation-wave-19-timeout preserves final logs, mutant failures, type-check
outputs, raw timings, medians, control hashes and all final process stdout/stderr
in a compressed archive. Scratch binaries and controls are in
/tmp/wave19-timeout-final. Early builds exposed a boolean-or-undefined lowering
refusal and absent-node panic; guards and explicit boolean comparison fixed
those before the final full run. They are not counted as mutant kills.
