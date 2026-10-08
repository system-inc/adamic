Before these merges, the shared fix applicator blocked activation of this completed checker migration.

Reproducer: `[1].reduce((a, b) => a + b, 0 as number);`

Production Go's PreferReduceTypeParameter emits two removals around the asserted value and a type-argument insertion. One removal is a zero-width empty edit. The unified finding transfer retains every edit, but stage1/cohere/lint/lint.ts:175 panics with `nonprogressing fix` when applying this legitimate no-op proposal. The Go rule is cohere/internal/lint/rules/typescript/prefer_reduce_type_parameter.go, in its CallExpression listener's fix construction. No edit or guard was dropped to make the migration green. The pending descriptor, node-aware engine, option adapter, witness and mutant remain here. The original standalone oracle and mutant checks remain active in typeaware.

Resolved during the three-fix integration: origin/lint-fix/nonprogressing-fix at b8c259568 rejects and records the empty no-op edit instead of panicking, preserving Go proposal and refusal order. The completed descriptor is now active under rules/. The unused whole-file run method was removed; only the supplied CallExpression is visited. Final validation is recorded in the wave-23 shared-fixes report.

Final rule validation: 42/42 upstream cases byte-identical on Go, sanitized native, Node and emitted JavaScript. Its owned witness passes and its omitted-verdict mutant is caught on all three port runtimes. The full final gate and remaining shared failures are recorded in ../../../typeaware/WAVE_23_SHARED_FIXES_REPORT.md. The broad ErrorTypes import was removed from this rule after scratch compilation refused it; only the required member accessor remains local and argument access uses the supplied node.
