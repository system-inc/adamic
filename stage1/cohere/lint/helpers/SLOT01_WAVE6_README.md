# Slot 01 sixth helper batch

One helper per `.a` file, pinned to cohere `715ba94f3608a6500086b1076ce5cb7e51b836db`.

| File | API | Contract |
|---|---|---|
| `collapse_normalize_utility_definition.a` | `collapseNormalizeUtilityDefinition(definition, normalizeArguments)` | Nil flag makes the definition a no-op; otherwise forwards its node list to the separately owned normalizer. |
| `collapse_normalize_value_function_arguments.a` | `collapseNormalizeValueFunctionArguments(nodes, roots, rewrite)` | Child-first traversal, declaration/present/nonempty guards, exact raw function substring bail and in-place value updates. |
| `collapse_register_framework_variants.a` | `collapseRegisterFrameworkVariants(registry, registrations)` | Existing kind replacement preserves name/order; new registrations retain explicit order; lastOrder advances for new maxima; records copied by value. |

A definition's `bound=false` represents a nil Go definition; its other adapter fields are ignored. Present definitions carry their ordered arena root indices. The normalizer callback must implement Go's normalizeValueFunctionArguments over the same arena. The wrapper tests provide actual Go whole-body normalization observations, and independently compare actual Go normalizeUtilityDefinition. Nil controls keep an unrelated arena observable so an erroneous callback invocation cannot hide as a no-op.

Walker nodes carry `kind`, `valuePresent`, mutable `value`, and ordered `children` indices. Acyclic non-null Go trees become a flat identity-preserving arena; nil child/root lists become empty arrays. Invalid node references panic. Children are visited before checking the parent kind, so children of comments and other non-declaration nodes are still reached. Only a nonempty present declaration whose raw value contains `--value(` or `--modifier(` is rewritten. The callback must perform ParseValue, normalizeValueFunctionNodes, then ValueToCss. It does not substitute an approximate argument parser. Tests supply the real Go callback observations and capture the real walker invocation order with an observational Go overlay.

The registry adapter exposes the fields this helper touches: registrations Map and mutable lastOrder. Other registry state belongs to its owner and is not changed. Records contain name, order and kind. Updating an existing registration creates a replacement record with its old name/order and the new kind. New records copy input fields, avoiding aliases that Go's value structs would not create. Negative and shared explicit orders are allowed. Replay ignores an existing root's incoming order when advancing lastOrder. Orders and initial lastOrder must be exact safe integers; unsupported wider Go integers and fractions refuse explicitly with `NotYet: framework variant order outside exact integer range`. The full Go int64 domain is not silently approximated.

Run from repository root after sourcing `/workspace/adamic-tools/env.sh`:

```sh
go test ./stage1/cohere/lint/helpers -run '^TestSlot01Wave6' -count=1 -v -timeout=20m > /tmp/lint-helpers-01-wave6.log 2>&1
```

The slot-owned Go overlay exposes actual private helpers. A second overlay adds only a trace append before the real walker's ParseValue call, preserving all branches and results. No cohere file is modified. Consumer strings retain complete statically evaluable Go concatenations; labels/messages are included, dynamic construction is not evaluated. Tree inputs host those strings in controlled Go Node arenas with nested children, wrong kinds, absent values and nil/present definitions. Registry inputs cover collisions, shared/negative orders, caller mutation, repeat replay, Unicode names and Go's complete actual framework registration table. Output strings are observed as UTF-16 units.

The argument-normalization helper was withdrawn after discovering an earlier worker's claim; its file is not delivered. Its historical passing log is retained only as withdrawn evidence. See [SLOT01_WAVE6_REPORT.md](SLOT01_WAVE6_REPORT.md), `slot01_wave6_readiness.json` and `evidence/slot01-wave6/` for every consumer, remaining dependency, command and mutant.
