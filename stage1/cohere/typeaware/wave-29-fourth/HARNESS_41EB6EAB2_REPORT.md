Integrated the newly named shared harness 41eb6eab2 while retaining current main b8fb957aa.
Dependency merge a20fae46d43e98ff52f04caac5e89703fc040871 preserves the previously pushed owned work.
Registry PASS 0.021s; 54-source JSX tree oracle PASS 26.804s; setup 33s, nproc 5.
Registry descriptor mutants are rejected; owned rule mutant proofs remain unchanged from the preceding landing.
No new claims; checker-backed source adapters, parked React analysis and dynamic configured RegExp remain blocked.

The exact new harness is 41eb6eab2b6de45ede0a40250765be295ee25fbd. Its two new commits retire batch 8's duplicate runner and capture adapters, migrate its positions witnesses and capture prefixes to the existing registry descriptors, update the JSX corpus test to obtain spans from an independent Go parser, and add DEDUP_LEDGER.md. The ledger was read in full. It explicitly excludes the typeaware waves from its batch dedup decisions and names none of wave 29's twelve claimed rules as losing copies. No owned rule was deleted. These shared changes were inherited unmodified through the merge, with no hand edits to shared files.

Current main remains b8fb957aa839a9e8cb0b54279dd9864fa317bd30, and both main and the exact named harness are ancestors. The previous worker tip is 1c5f4605a706460aa1779a9e092dc897511492a6. The harness was merged onto this already-rebased and re-green branch; no history rewrite or force push is needed for this dependency update.

Fresh commands (all test outputs redirected to logs):

- bash cloud/setup.sh: cache warm 33s, total 33s, tool/submodule phases 0s; nproc 5, cgroup quota 4, 17.6GB.
- source /workspace/adamic-tools/env.sh; go run ./cmd/lint-registry: PASS.
- go test ./stage1/cohere/lint/registry -count=1 -timeout=30m -v: PASS 0.021s. Rejection mutants cover no listener, duplicate public name, unsafe public name, missing class, missing finish hook, missing factory, missing named export, unknown field, missing oracle export and invalid kind. Deterministic regeneration passes.
- go test ./stage1/cohere/lint -run '^(TestJsxLintNode|TestJsxLintTrees)$' -count=1 -timeout=30m -v: only TestJsxLintTrees exists on this revision and actually runs. PASS 26.804s. The capture yields 1987 source/rule/options combinations and exactly 54 JSX sources; Go parser, Node and sanitized native agree on 45527 complete-tree bytes. No nonexistent TestJsxLintNode result is claimed.
- git diff --name-only 1c5f4605a HEAD -- cmd internal bridge oracle stage1/typescript stage1/cohere/typeaware stage1/cohere/lint/context.ts stage1/cohere/lint/registry: empty before this owned evidence commit. The prior owned runtime, compiler, bridge, parser, option profiles, checker context and rule-registration implementation are unchanged. The empty result is retained as unchanged-runtime.txt.

The previous owned full-profile oracle and compiling-mutant evidence therefore remains the exact source evidence from LANDING_B8FB957A_REPORT.md; it was not rerun during this focused shared-harness update. That preceding run passed the original package in 128.285s, both 77-file compiler and 287-file repository corpora, all next-batch option controls, kernel/graph/listener mutants, released-registry checks and ASan/UBSan/LeakSanitizer. Most recent observed native/Go process ratios are original compiler 5.83x/repository 1.87x and next compiler 6.27x/repository 1.92x; no new lint timing measurement was made here.

The shared RuleContext still has no checker program, compiler project identity, file path or symbol bindings. Named kind metadata is supported and no longer a blocker. Full JSX fragments and undef source adapters remain partial; constructed-context-values and the hook rules remain parked for native memo/escape/capture or HIR/SSA integration. Native nonconstant new RegExp(pattern, 'u') and raw Go/JS option dialect differences remain explicit. No hand-written matcher was restored, no complete source-rule parity is claimed for these partial rules, and no new claims are taken. Full repository/shared lint gates were not run; this continuation ran the two focused shared packages/tests above.
