Rebased both retained CSS helpers onto current main f8013f0baac41ddc340d76f83bddde38536a8f07.
Helper code before this evidence commit: 6d31b1b67dcb6971ab193f2c057d7eaf282e40ea.
Fresh Go, Node, emitted JavaScript and sanitized native comparisons passed; helper packages and filtered input oracle passed.
Escape and important-range mutants compiled and finished successfully, then differed only in the output comparison on all three backends.
No new claims; the rule branch still has shared registry, node API and lint build blockers.

Commands, after source /workspace/adamic-tools/env.sh:

- python3 stage1/cohere/lint/helpers/wave15/css_string/validate.py
- python3 stage1/cohere/lint/helpers/wave15/css_declaration/validate.py
- ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers/... -count=1 -v -timeout 20m
- ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v

parseCSSString: 198,584 queries and 5,648,824 identical bytes. parseCSSDeclaration: 8,836 queries and 619,181 identical bytes. Every original consumer fixture file is represented, and upstream Go consumer tests pass. Each helper supplies a dependency for better-tailwindcss/enforce-canonical-classes, enforce-consistent-class-order, enforce-consistent-variant-order, enforce-shorthand-classes, no-conflicting-classes and no-unknown-classes. This does not certify completed ports of those consumers.

Helper packages pass in 127.632s and 170.668s, including semantic and JSX refusal guard mutants. The input oracle passes in 3.482s with six probe misses and zero cache hits. Logs are retained beside this report.

On the rule worktree, cloud/setup.sh reports Go, clang, Node and submodules ready in 0s each, then fails warmup at profile_test.go:32 because portFiles is ranged over without calling it. nproc is 5. Tool versions: Go 1.27.1, clang 20.1.8, Node 24.19.0. Existing standalone builders remain the workaround. The full repository gate and shared rule integration remain unverified.
