Built: native process-exit, uncleared-race-timeout and blocking-standard-streams rules, with raw checker facts.
Commits: claim 19e28bf4; implementation and validation 7d1f1fcc (pushed).
Commands/output: TestWave21ProcessRules PASS 75.790s, 174 control findings, compiler 77 and repository 287 roots byte-identical; checker PASS 0.100s; Node PASS 14.748s; vet and gofmt clean.
Mutants: timer reversed handle-loss test caught at byte 81; process-exit reversed output-state test at 6038; blocking reversed imported-entry test at 15314; retained released handles caught for all four question variants.
Not covered: full repository gate, shared profile registration/JavaScript comparison, or every possible syntax combination; no further claims before this batch was pushed.

All three previously reserved rules are now implemented in their own .a files:
`nexus/correctness-no-process-exit-after-output`,
`nexus/correctness-no-uncleared-race-timeout`, and
`nexus/correctness-require-blocking-standard-streams`.
No rule source, registration generator, shared harness or protected compiler file was edited.
The only existing bridge file change is three dispatch arms in facts.go (six Go lines).

The raw bridge questions live in node_symbol_context.go/.a,
program_modules.go/.a, and resolved_declaration.go/.a. They supply symbol identities,
declaration ancestry, source text, resolved import edges, signature declarations
and flags. Go supplies no lint verdict. Native code performs timer-reference checks,
entry selection, module reachability, blocking analysis, local-call following,
control-flow traversal, try-chain state and every report decision. The native graph
is adapted from wave 05 and includes optional-chain joins and generator suspension.
Go cohere deliberately routes normal try completion into catch; the port preserves that.
The private driver sets the parser's top-level await context for these module corpora.

The independent Go oracle imports the unmodified production registry and no bridge code.
The comparison includes byte ranges, rule/message identities and exact messages,
ordered fixes and suggestions. These rules intentionally produce no fixes or suggestions;
every serialized finding carries zero for both, checked against Go.
The controls include 18 timer fixture sources, 48 process-exit fixture sources plus
the imported module and global-script helpers, 10 real-site before/after programs,
all 44 upstream blocking-streams programs (151 source files including the real sites),
a DOM timer control, 10 timer-binding boundaries and five control-flow boundaries.
Their 174 findings serialize to 117,021 identical bytes. Real-site builders are exported
unchanged as fixture data; separate directory trees prevent module graphs colliding.
The compiler and frozen repository manifests are the branch's existing populations:
77 compiler roots at TypeScript 050880ce59e30b356b686bd3144efe24f875ebc8, and 287 repository
roots. Both produce zero findings for this batch, including all file headings:
5,087 compiler bytes and 18,485 repository bytes.

Normal and ASan/UBSan/LeakSanitizer binaries agree on all three populations;
sanitizer stderr is empty. Each rule mutant compiles, exits 0 and has empty stderr,
then fails only the comparison. The process-exit mutant dependency graph uses
unique class names so it cannot be rejected because two copies of the same native
class are loaded. An earlier duplicate-module mutant was rejected by the compiler
and does not count as a caught judgment mutant.
After release, node-symbol-context, its follow-alias variant, program-modules and
resolved-declaration each panic with exit 70 and the exact invalid/released-handle
message. An overlay mutant retaining the live registry entry exits 0 for each,
so the released-handle assertion catches it. Live queries pass checker tests and
all sanitizer populations.

Quiet end-to-end wall time, including checker program load and native parsing:

| Population | Native | Go cohere |
| --- | ---: | ---: |
| Compiler 77 roots | 2.912723 s | 0.332955 s |
| Repository 287 roots | 0.337311 s | 0.117830 s |

The native implementation is slower on these populations. These are observations
from one run, not a benchmark distribution.

Commands (each redirected to a log file):

```
ADAMIC_WAVE21_PROCESS_ARTIFACTS=/workspace/wave21-process-final ADAMIC_WAVE21_COMPILER_CONFIG=/workspace/wave21-compiler/src/compiler/tsconfig.json ADAMIC_WAVE21_COMPILER_MANIFEST=/workspace/wave21-compiler.manifest ADAMIC_WAVE21_REPOSITORY_MANIFEST=/workspace/wave21-repository.manifest go test ./stage1/cohere/typeaware -run '^TestWave21ProcessRules$' -count=1 -v -timeout=10m
go test ./bridge/tsgo/checker -count=1
go vet ./bridge/tsgo/checker ./stage1/cohere/typeaware
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(maps_and_text|sorting|string_index|lone_surrogates|functions|closures|method_closures|generic_functions|number_parsing)\.a$' -count=1 -timeout=10m -v
```

The filtered Node gate runs nine fixtures through source Node, native sanitizers
and emitted JavaScript, plus its independent one-byte mutant. The new checker-backed
runner is validated natively against Go; shared emitted-JavaScript/profile support
belongs to codex/lint-harness-dot-a. No shared harness edits were made.
Original cloud setup timing: total 81 seconds, tools/submodules ready 0 seconds,
cache warmup 81 seconds; nproc is still 5. The existing tool environment is
/workspace/adamic-tools/env.sh. The workspace restarted during this turn and all
source changes and validation artifacts survived.

Complete logs and canonical outputs are in validation-wave-21-process, with SHA256SUMS.
