Built: structure.parameterTypeNode in one .a file, removing one prerequisite from four structure rules.
Commits: claim bd2da22d8 pushed before implementation; preceding helpers be2775d0e and rule branch b833f85ab are pushed and contain current main c01907a70.
Commands and outputs: all four original Go consumer suites PASS; 90 actual calls, five distinct metadata inputs; 4,514 cases and 88,662 canonical bytes identical on source Node, emitted JavaScript and sanitized native.
Mutant: parameter_kind_guard_lost compiles and exits successfully with empty stderr on all three runtimes; only byte comparison catches the wrong results.
Not covered: complete consumer-rule ports, native Program loading, arbitrary invalid node coordinates or generic nullable values; no complete rule readiness claimed.

Consumers: structure/network-require-hook-options-parameter (14 calls), structure/network-require-hook-variables-type (19), structure/next-require-api-parameter-name (13), structure/react-component-require-properties-type-suffix (44). The original Go fixture assertions run unchanged. An overlay records inputs while preserving the original private helper body; another exports it to an independent Go oracle. Expected result fields remain only in captured evidence. Adamic receives presence, numeric SyntaxKind and input annotation coordinates, never the Go answer. The Go oracle reconstructs parameter nodes with the real NodeFactory and checks that any returned pointer is exactly its supplied annotation.

The pinned parser's Parameter kind is 170. The accessor returns -1 for nil or another kind and otherwise preserves the supplied annotation coordinate, including zero. Controls cover kinds 0 through 500, both presence flags and absent/zero/nonzero annotations. Each possible annotation kind is tested without changing annotation identity. Wrong-kind controls retain stale annotation coordinates to prove the guard rather than relying on an input adapter to discard them.

The initial generic T | null version was explicitly refused by native lowering: stage 0 cannot lower number | null. The final implementation uses the same numeric node-coordinate contract as the earlier identity-preserving accessor helper. This changes the adapter representation, not Go behavior; no refusal earns comparison or mutant credit. No shared harness file or cohere source is changed, and there is no regex upstream.

Reproduce: source /workspace/adamic-tools/env.sh; python3 stage1/cohere/lint/helpers/from_wave1_11/validate_parameter.py --scratch <directory>, redirecting output to a log. TestParameterTypeNodeMatchesGoWithMutant replays retained inputs against the actual Go helper and rebuilds sanitized native and emitted JavaScript.

Owned replay test PASS in 3.981s; owned package vet PASS.
