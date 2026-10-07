Built: no-unsafe-function-type and no-useless-empty-export load through the unchanged registry using Ahra's permitted temporary .ts filenames; third rule remains blocked.
Commits: claim a01b4c3c; rule bodies ef739a97 and d52fd841; original evidence 1dee4361; finishing commits follow.
Commands and outputs: final four-execution parity/mutants/compiler/throughput PASS 173.377s; default registry and vet PASS; setup 30s, nproc 5.
Mutants: wrong-global-type-name and omit-empty-export-fix run cleanly and are killed by comparison on source Node, emitted JavaScript and sanitized native.
Not covered: no-unnecessary-type-constraint implementation/mutant because its suggestion edits a different span; full repository test gate not run. No more claims.

# Ahra's correction

The automatic continuation was premature. No further rules will be claimed.
Both implemented rules now use temporary rule.ts/messages.ts modules as Ahra
explicitly permitted, awaiting the integration codemod to .a. Only their owned
rule directories changed after the correction. Shared generator, harness,
formatter, oracle, parser and compiler production files were not edited.
The earlier claim-directory report is historical evidence from before the
correction; its default .a registration failure is resolved for these two rules
by the allowed temporary filenames, not by changing the foundation.

# Final validation

From Adamic root, after sourcing /workspace/adamic-tools/env.sh:

    go run ./cmd/lint-registry > stage1/cohere/lint/rules/no-unsafe-function-type/testdata/final-registration.log 2>&1
    go vet ./... > stage1/cohere/lint/rules/no-useless-empty-export/testdata/final-vet.log 2>&1

Both exit 0. The generator lists the five existing rules and these two new rules.
The final test supplement lives in this owned rule's testdata/comparison.go.txt.
A Go overlay adds it as a virtual extra test file. It replaces no existing shared
source and does not change the production generator or harness. Its only mapping:

    /workspace/adamic/stage1/cohere/lint/wave06_candidate_test.go
      -> /workspace/adamic/stage1/cohere/lint/rules/no-unsafe-function-type/testdata/comparison.go.txt

The mapping is ordinary JSON under Replace, written to
/tmp/wave1-06-final-owned-test-overlay.json. Recreate it with these absolute
paths for a reproduced run. The extra tests reuse unchanged shared helpers;
Go capture and adapter builds use the foundation's existing scratch overlays.
Unlike the pre-correction test runs, no .a compatibility production-source
overlay is involved in the final check.

    ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/wave1-06-typescript-6.0.3 go test -overlay=/tmp/wave1-06-final-owned-test-overlay.json ./stage1/cohere/lint -run '^TestWave06(Candidates|Mutants|Compiler|CompilerThroughput)$' -count=1 -v -timeout 15m > stage1/cohere/lint/rules/no-unsafe-function-type/testdata/final-parity.log 2>&1

PASS 173.377s. Filename-preserving upstream cases: 89 plus two owned witnesses,
30176 identical output bytes across Go cohere, source Node, emitted JavaScript
and sanitized native. Findings, exact messages/spans, declaration-file gates,
shadowing and every captured safe-fix output are compared. All successful
executions require exit 0 and empty stderr. Native uses ASan/UBSan and Linux leak
checking. Each final mutant succeeds on all three Adamic executions before
comparison kills it; no compile or sanitizer failure is credited as a kill.

Compiler pin: TypeScript v6.0.3 at
050880ce59e30b356b686bd3144efe24f875ebc8, scratch-only checkout. 77 src/compiler
files and 122 stage1 .ts/.a files, with no parse-refusal filtering, produce 398
file/rule rows and 24387660 identical output bytes on all four executions.
TestWave06Compiler PASS 73.52s. The emitted backend check runs after the ordinary
Go/Node/native comparison and is required for the test to pass.

Per-rule compiler throughput is one run including startup/file reads, native
optimized without sanitizers. Both rules have zero findings on all 77 compiler
files, therefore each observed findings/s is 0.00. These are actual completed
runs, not rates for an absent implementation:

| Rule | Go seconds / findings per second | Native seconds / findings per second | Node seconds / findings per second |
|---|---|---|---|
| no-unsafe-function-type | 0.559962 / 0.00 | 2.136227 / 0.00 | 3.618307 / 0.00 |
| no-useless-empty-export | 0.346322 / 0.00 | 1.848366 / 0.00 | 1.256641 / 0.00 |

The previous report separately labels nonzero positive-witness measurements;
they are not whole-compiler throughput. Setup succeeded in 30s on nproc 5,
Go 1.27.1, clang 20.1.8, Node v24.19.0. The earlier filtered uncached external
Node oracle passed 7.150s with six probe misses. Final vet used no compatibility
overlay. The full repository test gate was not run. Every test wrote directly
to a log; none was piped.

# Exact third-rule blocker and stop

NoUnnecessaryTypeConstraint on function data<T extends any>() {} reports the
parameter name T at 14..15, but its suggestion removes the distinct range
15..27, space + extends any. The unchanged upstream Go probe in
../../claims/wave1-06-fourth-evidence/constraint_oracle.go.txt proves this and
constraint.txt records its successful output. It also confirms there is no
unattended fix: the repair is a suggestion.

The retained Finding model stores one start/end pair. The shared oracle rejects
suggestions whose edit range differs from the diagnostic range with panic
unexpected suggestion shape. An owned rule cannot represent these exact
findings and suggestions through that model. Dropping the suggestion, moving
the reported span, or converting it to an unattended fix would change Go's
behavior. No substitute implementation was written and no third-rule mutant
is credited. Completing this rule requires shared separate diagnostic/edit
span support, beyond the announced .a/emitted-JavaScript harness work.

Per Ahra's instruction, stop here instead of editing shared files. The third
claim is not presented as completed. No additional claims or PR were opened.
