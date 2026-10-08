The shared fix applicator blocks activation of this completed checker migration.

Reproducer: `[1].reduce((a, b) => a + b, 0 as number);`

Production Go's PreferReduceTypeParameter emits two removals around the asserted value and a type-argument insertion. One removal is a zero-width empty edit. The unified finding transfer retains every edit, but stage1/cohere/lint/lint.ts:175 panics with `nonprogressing fix` when applying this legitimate no-op proposal. The Go rule is cohere/internal/lint/rules/typescript/prefer_reduce_type_parameter.go, in its CallExpression listener's fix construction. No edit or guard was dropped to make the migration green. The pending descriptor, node-aware engine, option adapter, witness and mutant remain here. The original standalone oracle and mutant checks remain active in typeaware.
