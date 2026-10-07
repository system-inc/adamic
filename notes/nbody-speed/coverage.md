# Field coverage for 3c6818554dde639879bb3dfa4a0fa13f94e7ad59

Read CLAUDE.md, README.md, docs/0.1.md, docs/memory.md, the four branch commit messages, the comparison to refreshed origin/main (e8ba3d5), and the changed files. The checkout initially had stale main (f580438); fetching main reduced the diff to 11 files.

Existing programs below are under internal/oracle/testdata. New names omit the nbody_ prefix and .a suffix in the new coverage column. This is source and emitted-C coverage, not a claim that every defensive IR-only branch is reachable.

| Condition or use | Existing source evidence | New coverage |
|---|---|---|
| Required numeric field with a uniform offset | field_access_paths.a: uniform | field_values, base_writes |
| Guarded uniform writes, both literal orders and repeated layouts | field_write_paths.a: update, uniformWrite | collection_fields |
| Conflicting offsets require cached data lookup | field_access_paths.a: Pair, Reordered; class_layouts.a | spread_fields, collection_fields |
| Non-C-name receiver falls back without repeating evaluation | No narrowed heterogeneous object receiver field write found | narrowed_receiver: emitted write uses adamic_object_write_field on a C cast |
| Exact class constructor shape vs other subclass shape | class_layouts.a, class_inheritance.a | base_writes with direct, base-typed virtual and super calls |
| Structural interface, literal, extra fields, class as interface | class_layouts.a, class_as_interface.a | collection_fields |
| Runtime layout with a uniform offset and zero matching literal guards | Error reads in exceptions.a; no Error field writes found | runtime_fields: name and message stores, checked fallback confirmed in emitted C |
| Spreads are skipped when collecting write guards | spreads.a, reuse_spread_method.a | spread_fields, private field omitted from public copy |
| Duplicate literal shapes are deduplicated; multiple distinct shapes guard a name | field_write_paths.a | collection_fields, optional_references |
| Static field name appears in a constructor layout | field_access_paths.a, inherited_static_field_read.a | static_collision |
| Static name shared by instance and literal fields; class-null and nonstatic descriptors | No combined static/instance/literal collision found | static_collision: shared and flag |
| Live parent reads before shadowing, reads after parent writes | inherited_static_field_read.a | static_collision through constructor parameter, Child and Grandchild |
| Parent and child writes, own-write flag and key order | field_write_paths.a, class_features_static.a | static_collision: constructor parameter writes, boolean and reference fields |
| Multi-level static parent delegation, shadow at each level | class_features_static.a: text | static_collision: shared and flag |
| Static own field vs inherited, static blocks, this and super access | class_features_static.a, class_features_static_private.a | Existing coverage retained |
| Required data cache misses, hits and shape changes | field_access_paths.a: update | collection_fields: array and Map values at one access site |
| Optional cache miss with present/absent field | literal_optional_shapes.a, field_access_paths.a | optional_references |
| Optional cache hit on present and absent shape, alternating back again | literal_optional_shapes.a: repeated empty objects | optional_references: repeated string and object presence and absence |
| Present optional property written after narrowing | field_write_paths.a: extra | optional_references: string and object replacement |
| Absent optional property written without presence proof | No oracle program creates a missing optional slot; native/field_write_absent.a is a safety test | optional_write_number.a, optional_write_string.a and optional_write_unknown.a differ, kept here |
| Optional chaining on undefined receiver vs absent property | literal_optional_shapes.a, optional_class_method.a | Existing coverage retained |
| Numeric representation: number, NaN, signed zero, number or undefined, undefined-only readonly view | maybe_number_slots.a, literal_optional_shapes.a, e4eec87_u01_undefined_field_widened.a | field_values: class writes undefined then -0 |
| Boolean stores and reads | maybe_booleans.a: Lamp | field_values, slot_kinds, static_collision |
| String reference replacement, retain before release, old reference NULL/non-NULL | undefined_references.a, class_layouts.a | field_values, optional_references, runtime_fields |
| Object, array, Map reference fields | class_layouts.a: tags, next; fresh_writes.a: named | field_values: replacement and self-assignment of all three |
| Function field replacement and self-assignment, calling stored closure | Functions in class_layouts.a, but no such replacement sequence found | slot_kinds on class and reordered literal |
| Weak field stores and reads, undefined Weak | weak_parent.a, doubly_linked.a, class_layouts.a | Existing coverage retained |
| Discriminated object unions, literal types | field_write_paths.a: Shape; narrowed_fields.a | Existing coverage retained |
| Heterogeneous scalar union field and optional boolean field | No supported field program found | Lowering refuses; probes here |
| Direct assignment, compound assignment, prefix/postfix increment/decrement | updates.a, field_write_paths.a | write_read_order, base_writes |
| Named function, method, override, super, captured closure calls | updates.a, class_inheritance.a | base_writes, collection_fields callback, slot_kinds |
| Array-held object field writes and reads | class_layouts.a: points loop | collection_fields with alias visibility |
| Map-held object field writes and reads | tuples_kept.a reads tuples; fresh_writes.a graph operations | collection_fields: values iterator and get result |
| Receiver evaluated once, RHS evaluated before store | updates.a: make, interfere; narrowed_writes.a | write_read_order: receiver() counter |
| Read one reference while writing another, RHS mutates original source | lent_reads.a reads across calls; read_order.a | write_read_order: target = source + change() |
| Undefined receiver check after RHS; RHS throws before store | narrowed_writes.a, throw_in_writes.a | Existing coverage retained |
| Frozen true and false branch of inline write check | library_object_freeze_write.a, library_object_freeze_alias.a, ordinary writes | Existing coverage retained |
| Regex presence disables uniform offsets, runtime named group layouts | regexp.a, regexp_exec.a, native/fields_test.go | Existing coverage retained |

## Differences

All three programs build successfully. The oracle runs also agreed between sanitized native and release native, and between Node and the JavaScript backend. Native differs from source Node in all three cases. This is the fixed-layout limitation documented by the branch, not evidence of a new regression introduced by the branch.

- optional_write_number.a: Node stdout `1\n2\nundefined\n2\n`, exit 0; native stdout `1\n2\nundefined\n`, stderr `adamic: panic: compiler bug: a field the checker proved is there is missing\n`, exit 70.
- optional_write_unknown.a: Node stdout `undefined\n2\n`, exit 0; native stdout `undefined\n`, the same stderr, exit 70.
- optional_write_string.a: Node stdout `absent\ncreated1\n`, exit 0; native stdout `absent\n`, the same stderr, exit 70.

The uncached oracle also labels the repeated exit-70 panic from its leak-check execution as a failure; that output contains no LeakSanitizer leak report.

The branch routes a nonstatic write to checked data lookup in internal/native/emit_objects.go:80. adamic_object_find in internal/native/runtime/object.c:54 panics instead of extending a missing property's layout. The number probe first writes a present amount slot, then an absent one: emitted C checks the known literal shape and falls back on the empty shape. The unknown probe has no known amount offset and uses checked lookup. The string probe has the runtime-reserved text offset but no matching literal guard and also falls back to lookup. Adding a missing property remains unsupported.

## Cases that cannot be tested as supported source programs

- unsupported_maybe_boolean_field.a: lowering at line 8 rejects a field of type boolean | undefined.
- unsupported_union_field.a: lowering at line 8 rejects a field of type string | number.
- fieldSlot's empty constructor body, non-Declare first statement, nonliteral initializer and spread initializer guards protect unexpected IR. Supported class lowering generates a constructor object declaration. No sound source program was identified that forces these IR shapes without changing lowering.

## Mutation proof

Changed exactly one line in internal/native/emit_objects.go:20:

```go
if e.staticFieldName(name) {
```

to:

```go
if false && e.staticFieldName(name) {
```

Ran only nbody_static_collision.a through the oracle. Go and clang compiled the mutant successfully. The oracle failed: Node exited 0 with the expected six lines; sanitized native exited 1 with UBSan reporting a member access through a null adamic_string pointer in string_build_impl.h:39. Release native terminated by signal. Restored exactly that line; the fixture passed again. The compiler source has no remaining diff.

## Toolchain

Sourced /workspace/adamic-tools/env.sh, the location selected by ADAMIC_TOOLS. Go 1.27.1, clang 20.1.8, Node v24.19.0.

```text
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0s)
setup: node ready (0s)
setup: submodules ready (1s)
setup: build cache warm (97s)
setup: done in 97s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
nproc: 5
```

## Final validation result

Ten new oracle fixtures agree across source Node, native under ASan/UBSan, release native and the JavaScript backend, with leak checks passing. Each standalone build agrees with source Node. The final counts update adds exactly ten rows and changes no existing row. Three notes programs differ at the documented missing-property limitation; two additional notes programs are refused by lowering.

The uncached repository gate exited 1 solely because TestMarkdownUnicodeWidths could not load the Node reference package emoji-regex from /tmp/adamic-markdown-width. Every other package passed, and every other test in markdownblocks finished without a reported failure. Installed the exact reference versions documented in generate_width: emoji-regex 10.6.0, get-east-asian-width 1.6.0 and narrow-emojis 0.0.3, with npm lifecycle scripts disabled. Reran only TestMarkdownUnicodeWidths uncached; it passed in 132.306s. The full command was not repeated after this environment fix. The original failure and successful isolated retry are both recorded in repository-gate.log and width-retry.log.
