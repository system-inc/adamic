Rebased the three retained helpers onto current main c01907a7036a22c2ea7ee686ed5fe4c6cd4bbc06 and reran their four-way comparisons.
Helper code before this evidence commit: a2efa7b9b0a169c3e3f5b58f3cb280c7c9552024; both owned branches remain landing-ready under the rule branch's named shared-harness parking exception.
parseCSSString, parseCSSDeclaration and parseFlags match fresh Go truth on source Node, emitted JavaScript and sanitized native; the filtered external oracle passes in 1.033s.
All three retained helper mutants finish normally with empty stderr and differ only in the output comparison on all three backends.
No new claim: DesignSystemForProgram still lacks Mutex (TS2305); its prior claim stays released. Other helpers remain unclaimed.

Main changed stage3, documentation and internal/oracle/stage3_hook_test.go. No cmd, production internal compiler/runtime, stage1, cohere submodule, cloud or oracle runner source changed between f8013f0b and c01907a7. This is the observed basis for reusing existing built artifacts and oracle executables. Fresh source Node uses the rebased files; all comparisons regenerate Go output. Shared leak-check changes, when landed, will be accepted in the next rebase; none occur in this main delta.

The logged replay command runs each helper's retained /tmp scratch oracle and cases.tsv, then its rebased main.a through oracle/node.mjs, emitted port.mjs and sanitized port. It repeats the corresponding mutant on the same cases, except parseFlags uses its six duplicate-state witnesses. Complete fresh observations remain under /tmp/wave15-c019-helper-runs; the replay summary is retained here.

- parseCSSString: 198,584 cases; 5,648,824 bytes identical. Escape mutant caught on all three Adamic paths.
- parseCSSDeclaration: 8,836 cases; 619,181 bytes identical. Important-range mutant caught on all three paths.
- parseFlags: 1,197,583 cases; 173,833,604 bytes identical. Duplicate-flag mutant caught on all three paths.

These are dependency proofs, not full consumer rule ports. The two CSS helpers each supply six edges to better-tailwindcss/enforce-canonical-classes, enforce-consistent-class-order, enforce-consistent-variant-order, enforce-shorthand-classes, no-conflicting-classes and no-unknown-classes. parseFlags supplies four edges to @next/next/no-html-link-for-pages, @typescript-eslint/no-empty-object-type, no-restricted-exports and no-restricted-imports. Zero final blockers are claimed removed by a helper alone.

ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^(TestInputAgreesWithNode|TestTheOracleCatchesOneByte)$' -count=1 -v passes with one native miss, one Node miss, six probe misses and zero hits.

python3 stage1/cohere/lint/helpers/wave15/design_system_cache/validate_blocker.py again confirms TS2305 missing Mutex, exit one and no native artifact. This remains a prerequisite probe, not a semantic cache mutant or a completed cache helper. No shared harness or compiler implementation was edited, and no hand-written regex matcher was introduced.

bash cloud/setup.sh succeeds: Go 0s, clang 0s, Node 1s, submodules 1s, build cache warm 37s, total 37s. nproc 5; cgroup cpu.max 400000 100000. Environment /workspace/adamic-tools/env.sh, Go 1.27.1, clang 20.1.8, Node 24.19.0. Compile-only warmup does not certify the full repository test gate.
