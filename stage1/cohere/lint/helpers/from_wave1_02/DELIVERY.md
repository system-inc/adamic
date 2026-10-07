Built CFG markFinal and markThrown; eight dependency entries removed across four rules, zero final blockers removed.
Rule branch parked at 2b4c31771; helper claims 631ea1bb9 and b8b1c7e37; helper branch based on current main f8013f0b; pushed delivery SHA accompanies report.
Checks PASS: actual Go, source Node, emitted JavaScript and sanitized native agree for 3,106 transitions and 98,252 bytes; owned Go adapter vet passes.
Six clean-executing semantic mutants caught only by comparisons: final unreachable/final guards, thrown unreachable/thrown guards, wrong flag, joint wrong list.
Not covered: whole-rule Adamic findings/fixes, full CFG builder, arbitrary arena/reflection semantics or full repository gate; shared harness parking blocker remains named.

Consumers for both helpers:

- array-callback-return
- consistent-return
- no-unreachable-loop
- react-hooks/rules-of-hooks

Each helper removes one frozen readiness dependency for each consumer. Many other CFG helpers remain, so none of these four rules becomes fully helper-ready from this pair. No rule implementation or full findings/fixes parity is claimed by the helper work.

The complete helper runs execute all 31 original Go top-level consumer tests and one control test apiece without skips or failures. markFinal records 2,852 consumer calls plus 32 controls, matching 87,942 bytes. markThrown records 150 consumer calls plus 64 controls, matching 10,074 bytes. The joint check has eight actual Go observations, matching 236 bytes. The sum is 3,106 transitions and 98,252 bytes. Node runs the original .a sources independently of Adamic lowering. JavaScript is independently emitted; native C uses ASan, UBSan and LeakSanitizer through the existing native.Build API. Complete final sources were rerun after the readonly and inheritance changes.

Commands (all output directly to logs):

- python3 stage1/cohere/lint/helpers/from_wave1_02/final/verify.py
- python3 stage1/cohere/lint/helpers/from_wave1_02/thrown/verify.py
- python3 stage1/cohere/lint/helpers/from_wave1_02/verify_joint.py
- go vet ./stage1/cohere/lint/helpers/from_wave1_02/final/build

Final logs: final/evidence/final-delivery-verification.log, thrown/evidence/final-delivery-verification.log, evidence/joint-verification.log and evidence/vet.log. The verifiers contain the exact original Go test commands and create their Go overlays in temporary directories. No cohere source or shared harness file is written. The bounded state replay does not claim to replace full CFG construction or rule finding/fix integration.

Mutants:

1. markFinal removes !block.reachable. Unreachable blocks wrongly become final.
2. markFinal removes block.final. Repeated entries duplicate the final list.
3. markThrown removes !block.reachable. Unreachable blocks wrongly become thrown exits.
4. markThrown removes block.thrown. Repeated entries duplicate the thrown list.
5. markThrown assigns block.final instead of block.thrown. Normal completion state changes and thrown state remains unset.
6. Joint check changes builder.thrown.push(handle) to builder.finals.push(handle). The two lists cease to be independent.

Each mutant passes type checking, JavaScript emission, clang compilation and all three runtime executions without a panic or sanitizer error. Only comparing to the actual Go post-states kills it. The helper source is restored after each mutation.

Arena contract: stable nonnegative integer handles identify blocks, -1 is nil, flags live on block instances, and list entries retain handles. Missing handles panic rather than silently marking a different block. Caller mutation that replaces an existing arena identity, Go reflection and distinctions between nil and empty slices are outside this typed adapter. FinalBlock is the nominal base of ThrownBlock; readonly arena views allow both helpers to access the identical richer block without unsound mutable-array widening. The joint check proves this sharing on both flag values and reachable/unreachable controls. Initial mutable-array and nominal-class refusals are retained in evidence/joint-invariance-refusal.log and joint-nominal-refusal.log; they were corrected, not counted as mutants.

Setup passed on the helper base: Go/clang/Node/submodules 0s each, cache warm 76s, total 76s, nproc 5, cpu.max 400000 100000, 17.6 GB. Environment /workspace/adamic-tools/env.sh. Initial direct nullable-block and CLI sanitizer gaps were handled entirely in owned files, as documented in the per-helper reports. The source models were rechecked on current main f8013f0b.

Ownership: the original enter claim was withdrawn because slot 01's earlier d6fbc135 preceded our 50a839593. Its duplicate sources were removed and no delivered credit is taken. markFinal was selected after reading all 536 origin refs and 19 distinct claims; markThrown after all 547 refs and 19 claims. Post-push audits found no competing claims for the retained helpers. The base HELPERS.md comment reservation was respected. No third helper is reserved.

Parking: codex/lint-wave1-02-land is pushed at 2b4c31771 and based on main f8013f0b. Its preserved main oracle passes in 33.891s; its named missing directory registration blocker awaits the sole harness owner on #zmh9v36. This is the explicit parking exception, not a claim that the owned witness gate passed. When the harness commit is named, resume that landing before further claims. No main or area branch was pushed. There are zero new shared-file hunks in this delivery and no shared-file rebase conflicts on the helper branch.
