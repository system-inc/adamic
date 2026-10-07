Built Theme.Resolve, Builder.variableDeclaration and declineListeners as three .a helpers; twelve prerequisite entries across ten rules, zero final blockers.
Commits: claim 94da62211eecf0be159edc5571fc3c354051aa72; rebased area b84a9d93 containing main c7991b90; implementation facef6102da2dbf3fb0016213813a98176a3b147; publication branch codex/lint-helpers-02 only.
Commands: setup PASS 67s/nproc 5; complete helper regression PASS 1075.396s; required real TypeScript comparison PASS 185.010s, 399 files/20,748,406 identical bytes; vet and supporting checks passed.
Mutants: all twenty-one new compiling semantic variants caught against actual Go, plus all 121 prior variants, 142 total with fresh independent mismatches.
Not covered: full repository gate, the other sixteen external correctness checks, integrated rule findings/fixes/suggestions, full CFG semantics, dependency wiring and prior external Tailwind gaps.

# Landing and ownership

The prior thirty-three helpers were complete, green and pushed at 1e3584ec39641d09116552a790962f804f386d7c before this claim. That complete gate passed in 771.866s with all 121 semantic mutants. This branch is the only branch this worker has published. Main and area are integration-owned and never pushed here.

An explicit wildcard fetch inspected every claim across twenty origin codex/lint-helpers* branches and nineteen claim files. These three symbols tie the largest unclaimed concrete consumer count at four. Previously reserved parameter/updateExpression/memberHeader/bindWithDefault symbols were rejected before any new claim. Claim 94da6221 was pushed before code. No fourth helper is claimed.

Only this slot's claims, helper files, standalone oracle and evidence are edited. New Adamic sources use .a. No shared harness, registration generator, inventory, compiler, rule dispatcher or regex implementation is changed. The requested area harness/finding model is accepted through ancestry; no inherited allocator changes are reverted.

# Observations and inferences

Observed: the unchanged actual consumer-source captures contain 2,276 inputs covering every consumer of these three helpers. They are reparsed by pinned Go cohere 715ba94f3608a6500086b1076ce5cb7e51b836db. Existing real Go theme/framework-variant/utility fixture strings and parsed TypeScript literal texts supplement candidate inputs. The corrected Go corpus has 5,622 theme replays and 366 actual CFG method entries; including the nil-node control and two builders/two callback modes yields 1,468 CFG replays. Exact observations are in evidence/helpers-final.log. Captures are reused, not credited as a fresh upstream package capture pass; evidence/capture-manifest.json links their hashes to batch6/batch10 provenance.

Theme.Resolve retains its original Go body under overlays of only the externally owned resolveKey/variableReference calls. Every deduplicated candidate and both presence values are retained. Caller options 0..3 and empty/tw prefixes vary for every candidate actually resolved by Go, plus empty and missing controls. Missing candidate combinations that return before inspecting options are not multiplied across all flags/prefixes. Inline-option union, empty inline literals, missing keys, ordered unchanged key-slice identity and original theme/candidate/presence arguments are observable. The separate key/reference implementations are actual Go dependencies, not reimplemented matchers.

Actual CFG builder execution captures variableDeclaration entries. Replays preserve original method bodies and node identities. External expression/bind/read/write callbacks observe order and mutate the same builder's current block; two builders and two callback modes distinguish state and identity. Nil-node controls and real parsed declarations with multiple entries, annotations, uninitialized names and object/array binding patterns supplement runtime coverage. Nonnil callers must project a real VariableDeclaration; arbitrary wrong-kind values violate Go's AsVariableDeclaration private precondition.

Each actual Go declineListeners factory receives a value-copy context and design-system result. The Go map is checked to contain only the named SourceFile subscription. Two fresh factories per parsed source receive repeated calls and a bounded reentrant description callback. Traces preserve rule name, source-file/node identity, message id, original system entry and actual DesignSystemDeclineMessage output. Adamic's opaque context/system handles must refer to stable snapshots of those Go value arguments, preserving contained pointer identities. Production adapter wiring is not implemented here.

The first returned-closure attempt was refused because its captured dependency record is cycle-capable. A concrete DeclineListener avoids that closure capture. Returning/storing it as the structural DeclineListeners interface erased method prototype origin and was refused too. Keeping the concrete class type at return/storage boundaries passes without compiler edits. These refused attempts get no parity or mutant credit.

Expected Want is removed from fixture inputs before Adamic comparisons. Source Node, emitted JavaScript and ASan/UBSan native must finish cleanly and agree byte for byte before actual Go mismatches receive mutant credit. No sanitizer/compile/runtime failure counts as a semantic mutant kill.

Inferred: twelve listed prerequisites disappear across ten distinct rules; none becomes fully helper-ready. RULES.md names every consumer. readiness.json subtracts only this slot's thirty-six retained helpers and remains conditional on the inventory's common AST adapter, without assuming other workers' unlanded work.

# Mutants and failed runs

The twenty-one final new variants change theme miss handling, caller/stored inline flags, empty-string presence, reference key, theme identity, candidate presence and key-slice identity; declaration nil guard, annotation evaluation, destructuring order/route, uninitialized read, initialized order and written-node identity; and listener repeat guard, guard placement before reentrant description, message id, context/node/system identities. Fresh first-mismatch witnesses are saved in evidence/mutant-witnesses.log.

The oversized initial matrix was terminated after a passing baseline and six caught variants because irrelevant repeated miss combinations made it impractical; its partial evidence is preserved in evidence/superseded-oversized-matrix.log and is not a passing run. The next run exposed a surviving uninitialized-read mutant: observed paths lacked that declaration shape. It receives no credit. The complete rerun adds a real parsed TypeScript control instead of weakening the mutant or skipping the check; evidence/superseded-missing-uninitialized-control.log preserves the failure.

The original shared option/policy and first slot02 suites compare actual Go, source Node and sanitized native. Batch2 onward additionally compares emitted JavaScript. No extra emitted-JavaScript coverage is claimed for the earlier suites.

# Rebase

A fresh fetch during validation advanced main from 39638d9e to c7991b90 and area from d65a8f93 to b84a9d93. The inherited changes include lowering and native record-runtime code. The pre-rebase complete regression was terminated and its incomplete log preserved in evidence/superseded-pre-rebase-regression.log, receiving no passing credit. All 54 local commits rebased cleanly onto current area; evidence/rebase.log records the operation. Final helper, compiler and supporting oracles are renewed on facef610, which contains both new upstream heads. Historical claim/report SHAs in older files describe their original publication. No new claim is made after the rebase.

# Required compiler input and commands

All test output goes directly to log files. Source /workspace/adamic-tools/env.sh before Go commands. Go 1.27.1, clang 20.1.8, Node 24.19.0; nproc 5.

The required TypeScript compiler input was initially absent. A public filtered clone at /tmp/slot02-required-typescript was checked out to 050880ce59e30b356b686bd3144efe24f875ebc8 (v6.0.3). Fetch/checkout logs are retained. The pre-implementation compiler comparison passed in 130.543s over 395 files and 20,731,017 output bytes. It receives no final-source credit; evidence/pre-implementation-required-compiler-oracle.log distinguishes it from the final post-implementation run.

- bash cloud/setup.sh > evidence/setup.log 2>&1; nproc > evidence/nproc.log: PASS 67s, ready Go/clang/Node/submodules 0s, cache warm/done 67s, processors 5.
- ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers -count=1 -v -timeout=30m > evidence/helpers-final.log 2>&1: PASS 1075.396s, all 142 mutants. TestSlot02Batch12 PASS 322.13s with all twenty-one new mutants. The larger timeout accommodates the growing complete regression; no check is skipped or relaxed.
- ADAMIC_TYPESCRIPT_SOURCE=/tmp/slot02-required-typescript ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint -run '^TestCompilerAndStage1Agree$' -count=1 -v -timeout=30m > evidence/required-compiler-oracle.log 2>&1: PASS 185.010s (comparison 184.97s), 399 files and identical 20,748,406 output bytes across Go/source Node/emitted JavaScript/native. The pre-rebase comparison passed in 206.821s; that log is retained separately in evidence/pre-rebase-required-compiler-oracle.log.
- go vet ./stage1/cohere/lint/helpers ./stage1/cohere/lint/inventory > evidence/vet.log 2>&1: exit zero, empty log.
- go test ./stage1/cohere/lint/inventory -count=1 -v > evidence/inventory.log 2>&1: PASS 8.671s on the rebased source.
- ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v > evidence/input-oracle.log 2>&1: PASS 20.696s on the rebased source, six probe misses, zero probe cache hits.

# Limits

The full repository gate and other sixteen mandatory external correctness checks were not run; none is represented as skipped-to-green or passed. Whole-rule findings/fixes/suggestions, native production dependency wiring, complete CFG graph semantics, malformed arenas, callback exceptions/concurrency, inherited comments package and eight prior unavailable external Tailwind live/corpus gates remain outside this helper unit. No shared blocker was worked around by editing shared files. A final wildcard fetch, ancestry and overlap audit precedes publication; upstream movement requires rebase and renewed oracle validation before push.

Final verification: main c7991b900362796aefd111474e65eb5398e91953 and area b84a9d9314b65d3d0261ee017e233287b4f071da are ancestors of implementation facef6102da2dbf3fb0016213813a98176a3b147. All thirty-six helper source hashes remain unchanged. Fresh witness names match all 121 previous variants exactly, plus twenty-one new variants. No skipped check appears in the completed helper or required compiler run. Twenty branches and nineteen whole claim files were re-audited after sibling updates; these symbols remain unique to this branch. Publication uses an exact lease against the original remote claim 94da62211eecf0be159edc5571fc3c354051aa72 on codex/lint-helpers-02 only.
