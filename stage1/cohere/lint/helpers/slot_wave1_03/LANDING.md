# Landing validation

Both original worker branches were absent from main when the landing-first cap
was received. No new helper is claimed. This helper branch rebased cleanly onto
origin/main e8ba3d5d81de4d3773c723914fccd4c76248b965, retaining its required
inventory/options/comments foundations and only the same two owned helpers.
Rebased implementation head before this receipt: 1ded1c2a6750d9e945638c7b6fdb5547f4c7f9a2.

Fresh checks, with logs under evidence/landing:

- bash cloud/setup.sh: PASS, 131s, nproc 5; Go/clang/Node/submodules ready at 0s.
- ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers/slot_wave1_03 -count=1 -v -timeout=20m: PASS 137.099s; Theme.Add 14,780 cases and 7,180,060 identical bytes, FrameworkStaticReading 6,239 cases and 1,085,222 identical bytes. Original Go, source Node, emitted JavaScript and sanitized native all agree.
- All six semantic mutants compile and execute normally, and are caught solely by comparison on source Node, emitted JavaScript and sanitized native.
- go vet ./stage1/cohere/lint/helpers/slot_wave1_03: exit 0, empty log.
- ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v -timeout=20m: PASS 6.621s, six uncached probe misses.

The helper scope remains unchanged: six consumers lose two prerequisites each,
zero final helper blockers removed. Explicit callback integration, whole-rule
parity and unavailable external repository fixtures are not newly certified.
The complete repository gate was not run. No main or area branch is pushed.

## Current main f8013f0b validation

Rebased cleanly onto f8013f0baac41ddc340d76f83bddde38536a8f07. The uncached owned helper package passed in 75.906s: 21,019 cases, unchanged canonical output hashes, all six semantic mutants caught solely by comparison on Node, emitted JavaScript and sanitized native. go vet passed with empty output. nproc is 5. No new helper claim; the rule foundation remains absent on main. Full repository gate was not run.

Current main c01907a7036a22c2ea7ee686ed5fe4c6cd4bbc06: clean rebase, uncached helper oracle PASS 58.059s, all 21,019 cases and six output-only mutants on all three paths. Rule branch c74d385b6 is explicitly parked unrebased for #zmh9v36.

Main b8fb957aa839a9e8cb0b54279dd9864fa317bd30: clean rebase; uncached owned helper package and all six comparison-only mutants passed on all three Adamic paths, canonical hashes unchanged; go vet passed. Strong CFG self-edge remains refused with adamic/cycle-capable. No additional helper claim. Rule branch remains explicitly parked for #zmh9v36; ab70f38d4 is not an ancestor of main or area/stage1-lint. Full gate not run. Raw main-b8fb957a logs preserve exact timings and refusal.

## Harness landed, rule ledger applied

Rule branch b0f179e6a is rebased onto area/stage1-lint 7481e0324 and pushed. Six losing rule copies removed. Shared owned witnesses pass 57,889 identical bytes and all six own mutants pass Node, emitted JavaScript and sanitized native. Full TestRulesAgree remains red on the decorated async executor recovery fixture (AtToken at 12); raw evidence is on that branch under rules/no-async-promise-executor/evidence/landing. No shared harness changed.

This helper branch is cleanly rebased onto current main 39638d9e278d38bb5aeae887f46d55a70e47aaad. Uncached owned helper tests PASS 56.386s, 21,019 cases, all six mutants caught solely by comparison on all three paths, canonical hashes unchanged. go vet passed with empty output. The strong CFG successor primitive still refuses adamic/cycle-capable. No further helper claimed, zero additional rules unblocked; full repository gate and new throughput measurements not run.

## Requested area helper base d65a8f931

This branch now rebases onto origin/area/stage1-lint d65a8f931c98655936ae04c6899f38f14862b73e, which carries the helper foundation and runtime profiling changes. Git skipped seven already-applied foundation commits; all owned helper implementations remain. Uncached helper package PASS 64.664s: 21,019 cases and all six semantic mutants match or fail comparison as intended on Node, emitted JavaScript and sanitized native; canonical hashes unchanged. go vet passed. Strong CFG successor self-edge still refuses adamic/cycle-capable.

Rule branch 9c71d1887 is rebased onto that same area and pushed. Its named-kind descriptors and six retained ports are unchanged; the six losing copies remain dropped. Shared owned witnesses PASS 57,889 bytes in 21.26s; six owned mutants PASS 117.16s. Combined oracle package exits 1 in 184.895s because TestRulesAgree still refuses AtToken at 12 in decorated async executor recovery. No oracle-green landing claim, shared harness edits, new helper claim, new throughput or full gate. Five processors available.
