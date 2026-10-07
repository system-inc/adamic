Built stylesheetCollector.ingest, Builder.nestedFunction and Builder.variableDeclarationList as three .a helpers; fourteen prerequisite entries across ten rules, zero final blockers.
Commits: initial claim 6bd6943b, replacement claim 3c9f6b2a; base area d65a8f93 containing current main 39638d9e; implementation ead51c7848283edcf377122668cc6dd68ee8033e; branch codex/lint-helpers-02 only.
Commands: setup PASS 65s/nproc 5; corrected isolated oracle PASS 86.824s; complete regression PASS 771.866s; supporting vet/inventory/input oracle recorded in evidence.
Mutants: all nineteen final new compiling semantic mutants caught against actual Go, plus all 102 prior, 121 total.
Not covered: full repository gate, integrated rule findings/fixes/suggestions, full CFG builder, external dependency wiring, malformed adapters, callback failures/concurrency, inherited comments and prior external Tailwind live/corpus gaps.

# Landing and ownership

The prior thirty helpers were complete, rebased onto area d65a8f93 containing main 39638d9e, green and pushed at d8515ce7 before this claim. That complete helper gate passed in 650.727s with all 102 semantic mutants; six inherited harness checks passed in 294.490s. This branch is the only branch this worker has published. Neither main nor any area branch is pushed.

An explicit wildcard fetch audited every claims file across twenty origin codex/lint-helpers* branches. Shared comments are already inherited. The generic strict-option ledger label remains rule-local decoding, not a named Go helper. stylesheetCollector.ingest has six consumers and is the largest unclaimed concrete count; the other two tie the next count at four. Claim 6bd6943b was pushed before code. Its makeReturn/makeThrow tie candidates were incorrectly selected despite their earlier slot 03 reservations b06c5170 at 06:32:02 UTC. Both are withdrawn, their sources removed and their results receive no delivery or mutant credit. Replacement claim 3c9f6b2a was pushed before nestedFunction/variableDeclarationList code. No fourth retained helper is claimed.

No shared registration, harness, compiler or inventory file is edited. All new Adamic files are .a. This batch adds no rule, kind dispatcher, finding-model change or regexp matcher. The inherited harness/finding model, profiles and comments are retained through the requested area base. No allocator leak-check replacement is reverted.

# Observation and inference

Observed: 2,276 previously captured actual consumer inputs, all six Tailwind and four CFG consumers. Go re-parses these sources and the existing CSS unit-fixture strings. Actual CSS parser nodes produce 7,833 replay cases. The corrected CFG run records 900 actual graph-method entries plus one factory-created list projection, yielding 3,612 replays. The log's Calls value 901 includes that synthetic projection. Exact counts are recorded in evidence/helpers-final.log; they include graph roots beyond each rule's filtered root selection and are helper coverage rather than every consumer's exact invocation replay.

CSS dependency failures are injected at call positions zero through six. Go verifies that every returned error is the original sentinel pointer. The adapter preserves this identity as token 7 and nil as -1. Imports receive resolved path strings, each dependency receives original node identity and path, and skips preserve name/params/path. CSS recursion and dispatch keep original method bodies intact under private dependency overlays. Node classification projects Go Kind/IsContainer.

CFG helper-entry instrumentation preserves original bodies. Header and declaration dependencies record original builder/node pointers, mutate the same builder's current block with Emit, and optionally change the current block before another emit. Replays cover two builder identities and two callback modes. A factory-created nil declaration list and nil-node controls supplement runtime inputs. Go's wrong-kind AsVariableDeclarationList panics; the Adamic helper explicitly refuses that invalid private precondition. The initial claim's wrong-kind guard wording must not be read as Go accepting arbitrary wrong-kind nodes.

Source Node, emitted JavaScript and ASan/UBSan native must agree and finish without stderr before their byte output is compared to actual Go. Expected Want is removed from the adapter input. Pinned cohere is 715ba94f3608a6500086b1076ce5cb7e51b836db. No cohere worktree is changed. evidence/capture-manifest.json records hashes of the reused, unchanged batch6/batch10 source captures; their original provenance and upstream package results remain in those batches. No fresh capture-package pass is claimed.

Inferred: six plus four plus four prerequisite entries disappear across ten distinct rules, but none becomes completely helper-ready. RULES.md names them. readiness.json subtracts this slot's thirty-three retained helpers without assuming other workers' unlanded work and remains conditional on the inventory's common AST adapter.

# Commands and evidence

All test output goes directly to files. Source /workspace/adamic-tools/env.sh before Go commands. Go 1.27.1, clang 20.1.8, Node 24.19.0.

- bash cloud/setup.sh > evidence/setup.log 2>&1; nproc > evidence/nproc.log: PASS 65s, nproc 5. Go ready 0s; clang/Node/submodules ready 1s; cache warm 65s; done 65s.
- Corrected isolated go test ./stage1/cohere/lint/helpers -run '^TestSlot02Batch11$' -count=1 -v -timeout=20m > evidence/isolated.log 2>&1: PASS 86.824s, seventeen retained compiling semantic mutants.
- ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers -count=1 -v -timeout=20m > evidence/helpers-final.log 2>&1: PASS 771.866s, all nineteen new variants and 102 previous variants caught.
- go vet ./stage1/cohere/lint/helpers ./stage1/cohere/lint/inventory > evidence/vet.log 2>&1: exit zero, empty log.
- go test ./stage1/cohere/lint/inventory -count=1 -v > evidence/inventory.log 2>&1: PASS 7.340s.
- ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v > evidence/input-oracle.log 2>&1: PASS 1.994s, six fixtures, zero cache hits/six probe misses.

# Mutants

Every new retained mutation is listed with its first fresh independent Go mismatch in evidence/mutant-witnesses.log after the complete gate. The stylesheet variants change resolved-path loading, resolve/theme/load/utility error propagation, utility/custom dispatch, skip path and ordinary/unknown-at-rule recursion. The CFG variants change nil-node/nil-declarations guards, header node/builder identity and call count, and declaration order/node/builder identity and call count. None relies on a compile failure, panic or sanitizer finding.

The first declaration-order mutant survived because observed lists contained only one declaration. Its failed run is preserved in evidence/superseded-missing-order-control.log and receives no credit. A real parsed multi-declaration control now kills that mutant. Initial driver type-import and non-null-assertion failures were corrected locally and are not evidence of parity. The replacement implementations are the only CFG helpers retained. The isolated run's theme variant omitted the call; the final stronger variant retains the call and specifically ignores its error, with separate load/utility error variants.

The original inherited option/policy and first slot02 tests compare Go/source Node/native. Batch2 onward additionally compares emitted JavaScript. Earlier reports' rewritten SHAs remain historical identifiers. No extra emitted-JavaScript coverage is claimed for the original suites.

# Limits

The full repository gate and cohere style/format CLI are not run. Whole-rule findings/fixes/suggestions, final AST adapter wiring, arbitrary malformed/cyclic arenas, mutations of traversal node lists by dependencies, callback exceptions/panics/concurrency and full CFG semantics remain outside this gate. CSS parser rejection is outside ingest's parsed-tree precondition. The separate inherited comments package and eight previously documented unavailable external Tailwind live/corpus gates remain excluded. No shared-harness edit or blocker is needed for retained helpers. Final fetch and ancestry checks precede publication; if upstream advances, the branch must rebase and re-green before push.

Final wildcard fetch confirms unchanged main 39638d9e and area d65a8f93, both ancestors of verified source ead51c78. Twenty branches and nineteen complete claim files were inspected; the three retained symbols appear only on this branch. All thirty-three source hashes match the manifest and all 102 prior mutant names match the previous completed gate. Publication is a fast-forward to this branch only. Working-tree and whitespace checks precede push.
