Built DesignSystem.VariantKind in variant_kind.a over a live registration-kind map.
Claim 4f3c8f1f was pushed before code on codex/lint-helpers-from-codex/lint-wave1-15.
validate.py passed 7,248 queries and 216,218 identical output bytes on Go, Node, emitted JavaScript and sanitized native; consumer Go tests passed.
A missing-root fallback mutant, static to functional, compiled and was caught by output comparison alone on all three backends.
Not covered: complete CSS loader/registry integration, nil receivers, every possible configuration or full repository gate.

The six consuming rules are better-tailwindcss/enforce-canonical-classes,
better-tailwindcss/enforce-consistent-class-order,
better-tailwindcss/enforce-consistent-variant-order,
better-tailwindcss/enforce-shorthand-classes,
better-tailwindcss/no-conflicting-classes and better-tailwindcss/no-unknown-classes.
This removes six prerequisite entries and zero last recorded helper blockers.
It is helper behavior parity, not six implemented lint rules.

The actual unchanged Go LoadedDesignSystem.VariantKind and VariantRegistry.Get
provide the answer. An owned oracle overlay only adds an exported seam; no
cohere source or shared harness is changed. A non-null registry is required.
The caller supplies its live registration-kind lookup, including empty values.
Unknown/missing roots return static; registered kinds are returned unchanged.

The Go AST collector reads all string literals in every matching consumer test
file, unquotes them with Go strconv and generates missing/registered/replaced/
deleted map states. Distinct literals per family are 131, 196, 48, 159, 130 and
139. These are helper input controls derived from consumer fixtures, not a claim
that each literal is an observed call to VariantKind in the upstream rule run.
The Go rule tests are independently run with their original assertions.

Run from the repository root after sourcing /workspace/adamic-tools/env.sh:

    python3 stage1/cohere/lint/helpers/wave15/validate.py > /tmp/wave15-helper-variant-validation.log 2>&1

All successful backend and mutant runs require exit zero and empty stderr.
Mutants rebuild both emitted JavaScript and ASan/UBSan native from the mutated
source. Node runs source with oracle/node.mjs. The owned build adds a virtual
standalone Go entry through an overlay and uses the existing compiler unchanged.
The complete logs, generated cases and literal coverage are in variant-evidence.
The native sanitizer run checks leaks; no mutant panic or build refusal counts.

The 389-ref selection excluded every helper-claim mention, the implemented and
explicitly reserved comment bundle, and the per-rule strict-options marker.
All remaining named helpers above six consumers were reserved. VariantKind was
chosen from the six-consumer tie. The original rule branch remains pushed; its
previous parser/repair/registration limitations are separate integration work.
