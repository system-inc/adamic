Built slice 6: exceptions-21-main, inherit-guards, feature-set-link-main and lint-features, preserving main slice 2.
Branch compiler/loom-r-slice-6; one commit on 67733635b0252eb9d3bb3645c72262c7b18361e1; full delivery SHA is in the push receipt and final response.
Commands/results: bounded four-CPU package runs, Linux clang 20.1.8 counts regeneration, guards, formatting, vet and a-check are itemized below.
Mutants: every attempted mutation, catcher and observed status is itemized below; raw failures are retained.
Not fully green or Loom-ready: broad stage1 readers hit memory exhaustion; over-budget and incomplete coverage is listed below.

## Extraction and assumptions

Base: `67733635b0252eb9d3bb3645c72262c7b18361e1`. Source: `26f45b7f2029fae240b0d216ea6ecb477e835efa` on `55347dce1537c0bd05d35eab0dc5d663c4823803`. Method: `git merge-tree --write-tree --merge-base=55347dce15 origin/main 26f45b7f20`, then apply the returned tree and resolve. This retains main infrastructure rather than replaying the old main merge as a reverse diff. Source review and runtime conditions are preserved in review/compiler/chain-slices/slice-6.md and review/compiler/chain-slice-6/.

Conservative choices: keep main async support, OrdinaryToPrimitive semantics, lexical readiness, portable cache recipes and strict WASI product infrastructure. Pending exception payload may be NULL for undefined; both payload and pending bit stay thread-local. Parallel propagation selects by exception_index under the existing scheduler lock and restores both words around helping. New Error-name literals are immutable, not shared mutable state. Main runtime statics registry lists the new TLS and immortals. Runtime clearance is still required for the full diff below, including union.c and the adapted main consumers.

Runtime condition: exactly one line, `Uncaught Name: message` for actual Error instances, otherwise `Uncaught ` plus undefined/null/number/boolean/string/function/object. Escape CR/LF, replace lone surrogates, flush stdout and stderr, release payload, exit 1. JS writes the same bytes synchronously; no Node stack.

Setup: GOPROXY=https://proxy.golang.org|direct. nproc=5, cgroup four CPUs; GOMAXPROCS=4 and test runner affinity restricted to four CPUs. House Go 1.27.1, Node 24.19.0 and clang 20.1.8. setup.sh hit timeout 300 (exit 124) during cold Go warming after successful submodule clone: go 0.039s, node 0.023s, markdown 0.090s, clang 0.245s, submodules 257.237s, shared-cache-off 257.239s. No final setup timing. Sourced its provisioned environment and completed required builds directly. An initial reference clone rejected a shallow reference; normal setup clone succeeded. Cache/product relocation, byte-verified archive deduplication and WASI SDK 27 installation addressed disk pressure; no cohere code was copied into Adamic.

## Conflicts

- `internal/lower/exceptions.go`: Keep main ObjectSpreadMayThrow and add slice 6 NodeBufferCall.MayThrow; preserve readiness-specific propagation.
- `internal/lower/import_cycle_ready_test.go`: Keep main cycle fixtures and native TDZ checks; apply source Node/JavaScript/native exit-1 and one-line uncaught comparisons, removing obsolete imports.
- `internal/lower/lower_test.go`: Keep main async support and unawaited-task refusal. Primitive throws move from old refusal rows to step 21's Node-held dynamic throw tests.
- `internal/lower/refusals.go`: Run main parallelPreflight and then the .a writing-call narrowing refusal; neither check replaces the other.
- `internal/lower/regexp_test.go`: Move the now-catchable replaceAll case to TestStep21LibraryAdmission. Main also admits huge repeat counts, so retain that admission in a registered Node/native agreement rather than restoring the old refusal.
- `internal/native/library.go`: Retain RuntimeRequest, suppliers, portable keys, cache publication and parallel compilation. Export SourceFlags, share main runtimeFeatures, and remove only compileRuntime's header feature injection.
- `internal/native/native.go`: Retain main buildOneWithRuntime, runtime symlinks, prefix maps, platform linker handling and deterministic timestamps. Add source feature flags at its actual compile-flags seam.
- `internal/native/runtime/adamic.h`: Retain main OrdinaryToPrimitive ABI and undefined-message declaration. Use a thread-local heap payload plus thread-local pending bit and source builtin Error class declarations.
- `internal/native/runtime/exceptions.c`: Combine source payload/classes/exact uncaught renderer with main absent-message shapes, stored field tags and lazy OrdinaryToPrimitive. Keep both pending words thread-local; use packed-cache zero initializers; give conversion failures nominal TypeError identity and pending=true.
- `internal/native/wasm_test.go`: Retain main ahead-built, relocated strict-WASI products and SDK compiler identity. Add source's three closure-feature fixtures and three selectable units; key and select six runtime variants from the five observed feature sets plus counted request.
- `internal/oracle/counts.md`: Discard merge contents, take origin/main, then regenerate once on Linux with house clang 20.1.8. No hand merge.
- `internal/oracle/oracle_test.go`: Keep main planted-family and leak-check harness. Pass original source identity to the JavaScript backend for terminal conventions; its generated cache comment uses a portable source label.
- `internal/oracle/wasi_test.go`: Keep main's newer split fixture emission. Its native.Build path already supplies source feature flags; do not restore the removed aggregate TestWASIEmission or absolute-path clang invocation.
- `stage1/cohere/lint/profile_compilation_main_test.go`: Keep main full-registry oracle and product/relocation infrastructure, including ADAMIC_NATIVE_SPLIT in keys. Use SourceFlags in identity and profiled compilation. The dedicated source-aware profile helper delegates to native.Build.
- `stage1/cohere/markdownblocks/list_layout_shards_test.go`: Keep main native.Build/BuildFlags portable recipes and v2 canary identity. Preserve source-aware flags without reinstating raw absolute-path clang commands.
- `stage1/cohere/markdownblocks/malformed_events_independent_test.go`: Keep main native.Build recipe, which already preserves source features and portable output. Source's older raw-clang replacement is superseded; final file equals main.

IR auto-merge checked: preserve SortedMaps, add CheckedLibrary/BuiltinError and MakeError constructor identity. Reader and sorted-map guard results are below. Main newly added object/JSON/parallel exception readers use the pending bit. Main primitive and JSON C witnesses and their embedded mutants were adapted to that ABI.

Own CSS/JSON test names now disambiguate the package. Two aggregate regexp replacement mutant tests were split into five top-level tests (offset/group order/named groups and argument/group type guards); all five mutations are retained. Fifteen existing writing-call admissions remain exact refusal tests under *_refused, with flow/fresh corpus coverage retained as refusal leaves. New native agreement rows use literal registered program names. WASI fixture census expands 35 to 38 and units 36 to 39.

## Counts: final Linux regeneration

`go test -p 1 ./internal/oracle -run TestCountsAreRecorded -count=1 -json -parallel=4 -timeout 3m -args -update-counts` exited 0. Main table was the starting point; no hand merging. An initial regeneration passed, a later attempt after fixture relocation failed for disk space, and a serial replay passed in 55.124s, and the delivered table was regenerated again after the final numeric repair in 87.019s. All failed attempts remain in the evidence. 147 affected rows: 26 added, 9 removed, 112 changed. The following includes every before/after row and member attribution.

### internal/oracle/testdata/047cb0d_n_arrayindex.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/047cb0d_n_arrayindex.a | 1 | 1 | 1 | 2 | 1 | 0 |
after:  None
```

### internal/oracle/testdata/accessor_spread_throw_first.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/accessor_spread_throw_first.a | 23 | 23 | 19 | 26 | 11 | 0 |
after:  | internal/oracle/testdata/accessor_spread_throw_first.a | 23 | 23 | 20 | 27 | 11 | 0 |
```

### internal/oracle/testdata/accessor_spread_throw_last.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/accessor_spread_throw_last.a | 27 | 27 | 21 | 30 | 11 | 0 |
after:  | internal/oracle/testdata/accessor_spread_throw_last.a | 27 | 27 | 22 | 31 | 11 | 0 |
```

### internal/oracle/testdata/accessor_spread_throw_middle.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/accessor_spread_throw_middle.a | 25 | 25 | 20 | 28 | 11 | 0 |
after:  | internal/oracle/testdata/accessor_spread_throw_middle.a | 25 | 25 | 21 | 29 | 11 | 0 |
```

### internal/oracle/testdata/array_from_length.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/array_from_length.a | 5 | 3 | 3 | 7 | 3 | 0 |
after:  | internal/oracle/testdata/array_from_length.a | 6 | 6 | 6 | 12 | 3 | 0 |
```

### internal/oracle/testdata/borrow_chain_coverage_finally.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/borrow_chain_coverage_finally.a | 13 | 13 | 9 | 13 | 6 | 0 |
after:  | internal/oracle/testdata/borrow_chain_coverage_finally.a | 13 | 13 | 12 | 16 | 6 | 0 |
```

### internal/oracle/testdata/borrow_element_throw.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/borrow_element_throw.a | 38 | 38 | 30 | 50 | 8 | 0 |
after:  | internal/oracle/testdata/borrow_element_throw.a | 38 | 38 | 36 | 56 | 8 | 0 |
```

### internal/oracle/testdata/borrow_loop.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/borrow_loop.a | 128 | 126 | 110 | 181 | 14 | 2 |
after:  | internal/oracle/testdata/borrow_loop.a | 128 | 126 | 116 | 187 | 14 | 2 |
```

### internal/oracle/testdata/borrow_loop_calls.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/borrow_loop_calls.a | 263 | 261 | 176 | 342 | 25 | 2 |
after:  | internal/oracle/testdata/borrow_loop_calls.a | 263 | 261 | 179 | 345 | 25 | 2 |
```

### internal/oracle/testdata/class_as_interface.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/class_as_interface.a | 372 | 372 | 303 | 468 | 60 | 0 |
after:  | internal/oracle/testdata/class_as_interface.a | 372 | 372 | 304 | 469 | 60 | 0 |
```

### internal/oracle/testdata/class_features_accessors.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/class_features_accessors.a | 67 | 67 | 47 | 104 | 21 | 0 |
after:  | internal/oracle/testdata/class_features_accessors.a | 67 | 67 | 49 | 106 | 21 | 0 |
```

### internal/oracle/testdata/class_features_static.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/class_features_static.a | 110 | 110 | 55 | 162 | 28 | 0 |
after:  | internal/oracle/testdata/class_features_static.a | 110 | 110 | 58 | 165 | 28 | 0 |
```

### internal/oracle/testdata/class_features_static_private.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/class_features_static_private.a | 42 | 42 | 53 | 94 | 12 | 0 |
after:  | internal/oracle/testdata/class_features_static_private.a | 42 | 42 | 73 | 114 | 12 | 0 |
```

### internal/oracle/testdata/class_inheritance_conditional.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/class_inheritance_conditional.a | 164 | 164 | 146 | 249 | 35 | 0 |
after:  | internal/oracle/testdata/class_inheritance_conditional.a | 164 | 164 | 157 | 260 | 35 | 0 |
```

### internal/oracle/testdata/class_inheritance_exceptions.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/class_inheritance_exceptions.a | 29 | 29 | 19 | 39 | 8 | 0 |
after:  | internal/oracle/testdata/class_inheritance_exceptions.a | 29 | 29 | 23 | 43 | 8 | 0 |
```

### internal/oracle/testdata/class_instance_key_throw.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/class_instance_key_throw.a | 6 | 5 | 5 | 10 | 5 | 1 |
after:  | internal/oracle/testdata/class_instance_key_throw.a | 6 | 5 | 7 | 12 | 5 | 1 |
```

### internal/oracle/testdata/closures_throw.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/closures_throw.a | 418 | 418 | 1566 | 1801 | 142 | 0 |
after:  | internal/oracle/testdata/closures_throw.a | 418 | 418 | 1585 | 1820 | 142 | 0 |
```

### internal/oracle/testdata/closures_throw_uncaught.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/closures_throw_uncaught.a | 31 | 29 | 20 | 36 | 13 | 0 |
after:  | internal/oracle/testdata/closures_throw_uncaught.a | 31 | 31 | 20 | 37 | 13 | 0 |
```

### internal/oracle/testdata/concurrency/accepted/throw_middle.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/concurrency/accepted/throw_middle.a | 9 | 9 | 6 | 13 | 9 | 0 |
after:  | internal/oracle/testdata/concurrency/accepted/throw_middle.a | 9 | 9 | 7 | 14 | 9 | 0 |
```

### internal/oracle/testdata/dead_zone.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/dead_zone.a | 0 | 0 | 0 | 0 | 0 | 0 |
after:  | internal/oracle/testdata/dead_zone.a | 1 | 1 | 2 | 1 | 1 | 0 |
```

### internal/oracle/testdata/debugger_fail.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/debugger_fail.a | 2 | 2 | 4 | 4 | 2 | 0 |
after:  | internal/oracle/testdata/debugger_fail.a | 2 | 2 | 5 | 5 | 2 | 0 |
```

### internal/oracle/testdata/devirtualize.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/devirtualize.a | 43 | 39 | 22 | 62 | 10 | 4 |
after:  | internal/oracle/testdata/devirtualize.a | 43 | 39 | 24 | 64 | 10 | 4 |
```

### internal/oracle/testdata/e4eec87_f2_union_narrow_call.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/e4eec87_f2_union_narrow_call.a | 1 | 0 | 3 | 3 | 1 | 0 |
after:  None
```

### internal/oracle/testdata/e4eec87_f2b_union_narrow_number.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/e4eec87_f2b_union_narrow_number.a | 2 | 1 | 4 | 5 | 2 | 0 |
after:  None
```

### internal/oracle/testdata/entries_record_alias.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/entries_record_alias.a | 65 | 65 | 89 | 118 | 28 | 0 |
after:  | internal/oracle/testdata/entries_record_alias.a | 65 | 65 | 90 | 119 | 28 | 0 |
```

### internal/oracle/testdata/exceptions.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/exceptions.a | 133 | 133 | 163 | 229 | 20 | 0 |
after:  | internal/oracle/testdata/exceptions.a | 133 | 133 | 186 | 252 | 20 | 0 |
```

### internal/oracle/testdata/exceptions_empty.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/exceptions_empty.a | 2 | 1 | 6 | 4 | 2 | 0 |
after:  | internal/oracle/testdata/exceptions_empty.a | 2 | 2 | 8 | 8 | 2 | 0 |
```

### internal/oracle/testdata/exceptions_uncaught.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/exceptions_uncaught.a | 3 | 1 | 3 | 5 | 2 | 0 |
after:  | internal/oracle/testdata/exceptions_uncaught.a | 3 | 3 | 3 | 6 | 2 | 0 |
```

### internal/oracle/testdata/fallthrough_exceptions.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/fallthrough_exceptions.a | 26 | 26 | 10 | 29 | 5 | 0 |
after:  | internal/oracle/testdata/fallthrough_exceptions.a | 26 | 26 | 12 | 31 | 5 | 0 |
```

### internal/oracle/testdata/fill_length.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/fill_length.a | 2 | 2 | 0 | 2 | 2 | 0 |
after:  | internal/oracle/testdata/fill_length.a | 3 | 3 | 3 | 5 | 2 | 0 |
```

### internal/oracle/testdata/finally_leaves.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/finally_leaves.a | 157 | 157 | 59 | 174 | 10 | 0 |
after:  | internal/oracle/testdata/finally_leaves.a | 157 | 157 | 64 | 179 | 10 | 0 |
```

### internal/oracle/testdata/fresh_writes.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/fresh_writes.a | 330 | 330 | 369 | 485 | 132 | 0 |
after:  | internal/oracle/testdata/fresh_writes.a | 330 | 330 | 370 | 486 | 132 | 0 |
```

### internal/oracle/testdata/generic_method_return.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/generic_method_return.a | 23 | 15 | 28 | 36 | 5 | 6 |
after:  | internal/oracle/testdata/generic_method_return.a | 23 | 17 | 30 | 39 | 5 | 6 |
```

### internal/oracle/testdata/getters_census_throwing_object.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/getters_census_throwing_object.a | 3 | 3 | 5 | 8 | 3 | 0 |
after:  | internal/oracle/testdata/getters_census_throwing_object.a | 3 | 3 | 6 | 9 | 3 | 0 |
```

### internal/oracle/testdata/inherited_fields_guard.a (inherit-guards)

```text
before: None
after:  | internal/oracle/testdata/inherited_fields_guard.a | 2 | 2 | 1 | 4 | 2 | 0 |
```

### internal/oracle/testdata/library_array_flat_map.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/library_array_flat_map.a | 50 | 50 | 60 | 109 | 17 | 0 |
after:  | internal/oracle/testdata/library_array_flat_map.a | 50 | 50 | 61 | 110 | 17 | 0 |
```

### internal/oracle/testdata/library_array_spliced.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/library_array_spliced.a | 72 | 72 | 40 | 114 | 12 | 0 |
after:  | internal/oracle/testdata/library_array_spliced.a | 72 | 72 | 41 | 115 | 12 | 0 |
```

### internal/oracle/testdata/library_array_with.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/library_array_with.a | 68 | 68 | 43 | 109 | 10 | 0 |
after:  | internal/oracle/testdata/library_array_with.a | 68 | 68 | 64 | 130 | 10 | 0 |
```

### internal/oracle/testdata/library_function_expressions.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/library_function_expressions.a | 56 | 56 | 49 | 96 | 19 | 0 |
after:  | internal/oracle/testdata/library_function_expressions.a | 56 | 56 | 50 | 97 | 19 | 0 |
```

### internal/oracle/testdata/library_map_set_group_by.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/library_map_set_group_by.a | 115 | 115 | 117 | 201 | 36 | 0 |
after:  | internal/oracle/testdata/library_map_set_group_by.a | 115 | 115 | 118 | 202 | 36 | 0 |
```

### internal/oracle/testdata/library_method_values_dead_zone.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/library_method_values_dead_zone.a | 0 | 0 | 0 | 0 | 0 | 0 |
after:  | internal/oracle/testdata/library_method_values_dead_zone.a | 1 | 1 | 2 | 1 | 1 | 0 |
```

### internal/oracle/testdata/library_object_freeze.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/library_object_freeze.a | 16 | 16 | 10 | 27 | 5 | 0 |
after:  | internal/oracle/testdata/library_object_freeze.a | 16 | 16 | 10 | 28 | 5 | 0 |
```

### internal/oracle/testdata/library_object_freeze_alias.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/library_object_freeze_alias.a | 4 | 1 | 7 | 8 | 3 | 0 |
after:  | internal/oracle/testdata/library_object_freeze_alias.a | 5 | 5 | 9 | 14 | 4 | 0 |
```

### internal/oracle/testdata/library_object_freeze_assign.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/library_object_freeze_assign.a | 4 | 2 | 6 | 8 | 2 | 0 |
after:  | internal/oracle/testdata/library_object_freeze_assign.a | 10 | 10 | 10 | 18 | 8 | 0 |
```

### internal/oracle/testdata/library_object_freeze_write.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/library_object_freeze_write.a | 2 | 1 | 5 | 6 | 2 | 0 |
after:  | internal/oracle/testdata/library_object_freeze_write.a | 3 | 3 | 7 | 11 | 2 | 0 |
```

### internal/oracle/testdata/long_literals.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/long_literals.a | 6117 | 6117 | 85 | 14610 | 8 | 0 |
after:  | internal/oracle/testdata/long_literals.a | 6117 | 6117 | 86 | 14611 | 8 | 0 |
```

### internal/oracle/testdata/lowering_chain_tdz_catch.a (exceptions-21-main)

```text
before: None
after:  | internal/oracle/testdata/lowering_chain_tdz_catch.a | 4 | 4 | 6 | 8 | 2 | 0 |
```

### internal/oracle/testdata/map_foreach_arrays.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/map_foreach_arrays.a | 995 | 995 | 932 | 1355 | 95 | 0 |
after:  | internal/oracle/testdata/map_foreach_arrays.a | 995 | 995 | 933 | 1356 | 95 | 0 |
```

### internal/oracle/testdata/map_foreach_closures.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/map_foreach_closures.a | 816 | 816 | 1044 | 1277 | 87 | 0 |
after:  | internal/oracle/testdata/map_foreach_closures.a | 816 | 816 | 1045 | 1278 | 87 | 0 |
```

### internal/oracle/testdata/map_foreach_fnexpr.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/map_foreach_fnexpr.a | 44 | 44 | 64 | 100 | 15 | 0 |
after:  | internal/oracle/testdata/map_foreach_fnexpr.a | 44 | 44 | 65 | 101 | 15 | 0 |
```

### internal/oracle/testdata/map_foreach_keep.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/map_foreach_keep.a | 516 | 516 | 625 | 931 | 35 | 0 |
after:  | internal/oracle/testdata/map_foreach_keep.a | 516 | 516 | 626 | 932 | 35 | 0 |
```

### internal/oracle/testdata/map_foreach_keys.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/map_foreach_keys.a | 391 | 391 | 363 | 674 | 25 | 0 |
after:  | internal/oracle/testdata/map_foreach_keys.a | 391 | 391 | 364 | 675 | 25 | 0 |
```

### internal/oracle/testdata/map_foreach_named_closures.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/map_foreach_named_closures.a | 117 | 117 | 179 | 219 | 29 | 0 |
after:  | internal/oracle/testdata/map_foreach_named_closures.a | 117 | 117 | 180 | 220 | 29 | 0 |
```

### internal/oracle/testdata/map_foreach_named_more.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/map_foreach_named_more.a | 395 | 395 | 569 | 819 | 36 | 0 |
after:  | internal/oracle/testdata/map_foreach_named_more.a | 395 | 395 | 570 | 820 | 36 | 0 |
```

### internal/oracle/testdata/map_foreach_named_numbers.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/map_foreach_named_numbers.a | 117 | 117 | 73 | 177 | 12 | 0 |
after:  | internal/oracle/testdata/map_foreach_named_numbers.a | 117 | 117 | 74 | 178 | 12 | 0 |
```

### internal/oracle/testdata/map_foreach_named_objects.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/map_foreach_named_objects.a | 165 | 165 | 207 | 278 | 31 | 0 |
after:  | internal/oracle/testdata/map_foreach_named_objects.a | 165 | 165 | 208 | 279 | 31 | 0 |
```

### internal/oracle/testdata/map_foreach_numbers.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/map_foreach_numbers.a | 373 | 373 | 155 | 478 | 14 | 0 |
after:  | internal/oracle/testdata/map_foreach_numbers.a | 373 | 373 | 156 | 479 | 14 | 0 |
```

### internal/oracle/testdata/map_foreach_objects.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/map_foreach_objects.a | 732 | 732 | 834 | 1076 | 63 | 0 |
after:  | internal/oracle/testdata/map_foreach_objects.a | 732 | 732 | 835 | 1077 | 63 | 0 |
```

### internal/oracle/testdata/method_coverage_empty_alias_tdz.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/method_coverage_empty_alias_tdz.a | 0 | 0 | 0 | 0 | 0 | 0 |
after:  | internal/oracle/testdata/method_coverage_empty_alias_tdz.a | 1 | 1 | 2 | 1 | 1 | 0 |
```

### internal/oracle/testdata/method_coverage_format_failure.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/method_coverage_format_failure.a | 1 | 0 | 0 | 1 | 1 | 0 |
after:  | internal/oracle/testdata/method_coverage_format_failure.a | 2 | 2 | 4 | 6 | 2 | 0 |
```

### internal/oracle/testdata/method_coverage_string_repeat.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/method_coverage_string_repeat.a | 2 | 2 | 3 | 6 | 2 | 0 |
after:  | internal/oracle/testdata/method_coverage_string_repeat.a | 2 | 2 | 5 | 8 | 2 | 0 |
```

### internal/oracle/testdata/module_namespace_reads/direct_early.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/module_namespace_reads/direct_early.a | 0 | 0 | 0 | 0 | 0 | 0 |
after:  | internal/oracle/testdata/module_namespace_reads/direct_early.a | 1 | 1 | 2 | 1 | 1 | 0 |
```

### internal/oracle/testdata/module_namespace_reads/early.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/module_namespace_reads/early.a | 0 | 0 | 0 | 0 | 0 | 0 |
after:  | internal/oracle/testdata/module_namespace_reads/early.a | 1 | 1 | 2 | 1 | 1 | 0 |
```

### internal/oracle/testdata/move_throw.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/move_throw.a | 42 | 42 | 22 | 50 | 9 | 0 |
after:  | internal/oracle/testdata/move_throw.a | 42 | 42 | 24 | 52 | 9 | 0 |
```

### internal/oracle/testdata/named_function_values.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/named_function_values.a | 116 | 116 | 127 | 222 | 25 | 0 |
after:  | internal/oracle/testdata/named_function_values.a | 116 | 116 | 128 | 223 | 25 | 0 |
```

### internal/oracle/testdata/narrowed_methods.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/narrowed_methods.a | 3 | 3 | 3 | 6 | 3 | 0 |
after:  None
```

### internal/oracle/testdata/narrowed_numbers.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/narrowed_numbers.a | 2 | 2 | 0 | 2 | 2 | 0 |
after:  None
```

### internal/oracle/testdata/narrowed_reads.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/narrowed_reads.a | 5 | 5 | 3 | 7 | 4 | 0 |
after:  None
```

### internal/oracle/testdata/narrowed_union_valid.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/narrowed_union_valid.a | 9 | 9 | 13 | 23 | 3 | 0 |
after:  None
```

### internal/oracle/testdata/narrowed_writes.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/narrowed_writes.a | 1 | 1 | 2 | 4 | 1 | 0 |
after:  None
```

### internal/oracle/testdata/nested_destructured_tdz.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/nested_destructured_tdz.a | 2 | 0 | 1 | 0 | 2 | 0 |
after:  | internal/oracle/testdata/nested_destructured_tdz.a | 3 | 3 | 3 | 3 | 3 | 0 |
```

### internal/oracle/testdata/nested_tdz.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/nested_tdz.a | 2 | 0 | 1 | 0 | 2 | 0 |
after:  | internal/oracle/testdata/nested_tdz.a | 3 | 3 | 3 | 3 | 3 | 0 |
```

### internal/oracle/testdata/nested_tdz_write.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/nested_tdz_write.a | 2 | 0 | 1 | 0 | 2 | 0 |
after:  | internal/oracle/testdata/nested_tdz_write.a | 3 | 3 | 3 | 3 | 3 | 0 |
```

### internal/oracle/testdata/node_buffer_digest_twice.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/node_buffer_digest_twice.a | 4 | 2 | 4 | 6 | 3 | 0 |
after:  | internal/oracle/testdata/node_buffer_digest_twice.a | 5 | 5 | 7 | 10 | 3 | 0 |
```

### internal/oracle/testdata/node_buffer_finalized.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/node_buffer_finalized.a | 4 | 2 | 4 | 6 | 3 | 0 |
after:  | internal/oracle/testdata/node_buffer_finalized.a | 5 | 5 | 7 | 10 | 3 | 0 |
```

### internal/oracle/testdata/node_fs_file_close.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/node_fs_file_close.a | 27 | 27 | 13 | 30 | 7 | 0 |
after:  | internal/oracle/testdata/node_fs_file_close.a | 27 | 27 | 16 | 33 | 7 | 0 |
```

### internal/oracle/testdata/node_fs_file_mkdir.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/node_fs_file_mkdir.a | 61 | 61 | 24 | 81 | 10 | 0 |
after:  | internal/oracle/testdata/node_fs_file_mkdir.a | 61 | 61 | 29 | 86 | 10 | 0 |
```

### internal/oracle/testdata/node_fs_file_write_buffer.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/node_fs_file_write_buffer.a | 34 | 34 | 11 | 45 | 12 | 0 |
after:  | internal/oracle/testdata/node_fs_file_write_buffer.a | 34 | 34 | 15 | 49 | 12 | 0 |
```

### internal/oracle/testdata/node_fs_file_write_file.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/node_fs_file_write_file.a | 31 | 31 | 11 | 36 | 9 | 0 |
after:  | internal/oracle/testdata/node_fs_file_write_file.a | 31 | 31 | 13 | 38 | 9 | 0 |
```

### internal/oracle/testdata/node_fs_rm_options.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/node_fs_rm_options.a | 36 | 36 | 5 | 37 | 6 | 0 |
after:  | internal/oracle/testdata/node_fs_rm_options.a | 36 | 36 | 7 | 39 | 6 | 0 |
```

### internal/oracle/testdata/normalize_form.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/normalize_form.a | 3 | 2 | 0 | 3 | 3 | 0 |
after:  | internal/oracle/testdata/normalize_form.a | 4 | 4 | 3 | 7 | 3 | 0 |
```

### internal/oracle/testdata/number_formats.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/number_formats.a | 225 | 225 | 55 | 262 | 18 | 0 |
after:  | internal/oracle/testdata/number_formats.a | 225 | 225 | 97 | 304 | 18 | 0 |
```

### internal/oracle/testdata/param_assigned_in_try.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/param_assigned_in_try.a | 46 | 46 | 29 | 64 | 12 | 0 |
after:  | internal/oracle/testdata/param_assigned_in_try.a | 46 | 46 | 31 | 66 | 12 | 0 |
```

### internal/oracle/testdata/precision_range.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/precision_range.a | 4 | 3 | 1 | 4 | 2 | 0 |
after:  | internal/oracle/testdata/precision_range.a | 5 | 5 | 4 | 9 | 2 | 0 |
```

### internal/oracle/testdata/radix_range.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/radix_range.a | 3 | 3 | 0 | 3 | 1 | 0 |
after:  | internal/oracle/testdata/radix_range.a | 4 | 4 | 3 | 6 | 1 | 0 |
```

### internal/oracle/testdata/regexp.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/regexp.a | 437 | 437 | 382 | 417 | 58 | 0 |
after:  | internal/oracle/testdata/regexp.a | 437 | 437 | 381 | 416 | 58 | 0 |
```

### internal/oracle/testdata/regexp_matchall_nonglobal.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/regexp_matchall_nonglobal.a | 1 | 0 | 2 | 0 | 1 | 0 |
after:  | internal/oracle/testdata/regexp_matchall_nonglobal.a | 2 | 2 | 5 | 4 | 2 | 0 |
```

### internal/oracle/testdata/regexp_null_narrowed.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/regexp_null_narrowed.a | 4 | 4 | 8 | 5 | 3 | 0 |
after:  None
```

### internal/oracle/testdata/regexp_replace.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/regexp_replace.a | 625 | 625 | 206 | 336 | 113 | 0 |
after:  | internal/oracle/testdata/regexp_replace.a | 625 | 625 | 205 | 335 | 113 | 0 |
```

### internal/oracle/testdata/regexp_replace/throw.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/regexp_replace/throw.a | 13 | 13 | 11 | 14 | 11 | 0 |
after:  | internal/oracle/testdata/regexp_replace/throw.a | 13 | 13 | 12 | 15 | 11 | 0 |
```

### internal/oracle/testdata/regexp_replaceall_nonglobal.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/regexp_replaceall_nonglobal.a | 1 | 0 | 2 | 0 | 1 | 0 |
after:  | internal/oracle/testdata/regexp_replaceall_nonglobal.a | 2 | 2 | 5 | 4 | 2 | 0 |
```

### internal/oracle/testdata/regions_throw.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/regions_throw.a | 124 | 95 | 9 | 99 | 31 | 29 |
after:  | internal/oracle/testdata/regions_throw.a | 124 | 95 | 15 | 105 | 31 | 29 |
```

### internal/oracle/testdata/reuse_throw.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/reuse_throw.a | 26 | 26 | 24 | 41 | 8 | 0 |
after:  | internal/oracle/testdata/reuse_throw.a | 26 | 26 | 27 | 44 | 8 | 0 |
```

### internal/oracle/testdata/review/agree/fxspptb_57f2d04_with_frozen.a (exceptions-21-main)

```text
before: None
after:  | internal/oracle/testdata/review/agree/fxspptb_57f2d04_with_frozen.a | 6 | 6 | 2 | 10 | 4 | 0 |
```

### internal/oracle/testdata/review/agree/fxspptb_903f25b_try_assign.a (exceptions-21-main)

```text
before: None
after:  | internal/oracle/testdata/review/agree/fxspptb_903f25b_try_assign.a | 8 | 8 | 7 | 14 | 6 | 0 |
```

### internal/oracle/testdata/review/agree/fxspptb_9984394_lib_dispatch.a (exceptions-21-main)

```text
before: None
after:  | internal/oracle/testdata/review/agree/fxspptb_9984394_lib_dispatch.a | 4 | 4 | 8 | 11 | 2 | 0 |
```

### internal/oracle/testdata/route_targets_try_loop.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/route_targets_try_loop.a | 68 | 66 | 41 | 85 | 9 | 2 |
after:  | internal/oracle/testdata/route_targets_try_loop.a | 68 | 66 | 44 | 88 | 9 | 2 |
```

### internal/oracle/testdata/search_from_sweep.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/search_from_sweep.a | 133 | 133 | 33 | 154 | 10 | 0 |
after:  | internal/oracle/testdata/search_from_sweep.a | 133 | 133 | 34 | 155 | 10 | 0 |
```

### internal/oracle/testdata/set_foreach_arrays.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/set_foreach_arrays.a | 516 | 516 | 524 | 713 | 58 | 0 |
after:  | internal/oracle/testdata/set_foreach_arrays.a | 516 | 516 | 525 | 714 | 58 | 0 |
```

### internal/oracle/testdata/set_foreach_closures.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/set_foreach_closures.a | 411 | 411 | 634 | 705 | 39 | 0 |
after:  | internal/oracle/testdata/set_foreach_closures.a | 411 | 411 | 635 | 706 | 39 | 0 |
```

### internal/oracle/testdata/set_foreach_keep.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/set_foreach_keep.a | 290 | 290 | 385 | 520 | 18 | 0 |
after:  | internal/oracle/testdata/set_foreach_keep.a | 290 | 290 | 386 | 521 | 18 | 0 |
```

### internal/oracle/testdata/set_foreach_keys.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/set_foreach_keys.a | 285 | 285 | 242 | 477 | 16 | 0 |
after:  | internal/oracle/testdata/set_foreach_keys.a | 285 | 285 | 243 | 478 | 16 | 0 |
```

### internal/oracle/testdata/set_foreach_named_closures.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/set_foreach_named_closures.a | 95 | 95 | 155 | 187 | 28 | 0 |
after:  | internal/oracle/testdata/set_foreach_named_closures.a | 95 | 95 | 156 | 188 | 28 | 0 |
```

### internal/oracle/testdata/set_foreach_named_more.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/set_foreach_named_more.a | 179 | 179 | 297 | 378 | 26 | 0 |
after:  | internal/oracle/testdata/set_foreach_named_more.a | 179 | 179 | 298 | 379 | 26 | 0 |
```

### internal/oracle/testdata/set_foreach_named_objects.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/set_foreach_named_objects.a | 87 | 87 | 110 | 149 | 22 | 0 |
after:  | internal/oracle/testdata/set_foreach_named_objects.a | 87 | 87 | 111 | 150 | 22 | 0 |
```

### internal/oracle/testdata/set_foreach_numbers.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/set_foreach_numbers.a | 214 | 214 | 140 | 317 | 11 | 0 |
after:  | internal/oracle/testdata/set_foreach_numbers.a | 214 | 214 | 141 | 318 | 11 | 0 |
```

### internal/oracle/testdata/set_foreach_objects.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/set_foreach_objects.a | 368 | 368 | 515 | 586 | 30 | 0 |
after:  | internal/oracle/testdata/set_foreach_objects.a | 368 | 368 | 516 | 587 | 30 | 0 |
```

### internal/oracle/testdata/statements_small_stopped/structural_error.a (exceptions-21-main)

```text
before: None
after:  | internal/oracle/testdata/statements_small_stopped/structural_error.a | 2 | 2 | 5 | 5 | 2 | 0 |
```

### internal/oracle/testdata/statements_small_throw.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/statements_small_throw.a | 5 | 5 | 9 | 13 | 4 | 0 |
after:  | internal/oracle/testdata/statements_small_throw.a | 5 | 5 | 16 | 20 | 4 | 0 |
```

### internal/oracle/testdata/step21_builtin_narrow_terminal.ts (exceptions-21-main)

```text
before: None
after:  | internal/oracle/testdata/step21_builtin_narrow_terminal.ts | 6 | 6 | 10 | 13 | 5 | 0 |
```

### internal/oracle/testdata/step21_catch_callback.a (exceptions-21-main)

```text
before: None
after:  | internal/oracle/testdata/step21_catch_callback.a | 17 | 17 | 12 | 24 | 6 | 0 |
```

### internal/oracle/testdata/step21_dynamic.a (exceptions-21-main)

```text
before: None
after:  | internal/oracle/testdata/step21_dynamic.a | 65 | 65 | 49 | 100 | 6 | 0 |
```

### internal/oracle/testdata/step21_dynamic_uncaught.a (exceptions-21-main)

```text
before: None
after:  | internal/oracle/testdata/step21_dynamic_uncaught.a | 0 | 0 | 1 | 2 | 0 | 0 |
```

### internal/oracle/testdata/step21_error_subclasses.a (exceptions-21-main)

```text
before: None
after:  | internal/oracle/testdata/step21_error_subclasses.a | 16 | 16 | 36 | 44 | 5 | 0 |
```

### internal/oracle/testdata/step21_finally_callback.a (exceptions-21-main)

```text
before: None
after:  | internal/oracle/testdata/step21_finally_callback.a | 19 | 19 | 18 | 34 | 8 | 0 |
```

### internal/oracle/testdata/step21_finally_completion.a (exceptions-21-main)

```text
before: None
after:  | internal/oracle/testdata/step21_finally_completion.a | 17 | 17 | 6 | 19 | 5 | 0 |
```

### internal/oracle/testdata/step21_library_failures.a (exceptions-21-main)

```text
before: None
after:  | internal/oracle/testdata/step21_library_failures.a | 35 | 35 | 54 | 80 | 4 | 0 |
```

### internal/oracle/testdata/step21_library_host.a (exceptions-21-main)

```text
before: None
after:  | internal/oracle/testdata/step21_library_host.a | 25 | 25 | 31 | 48 | 7 | 0 |
```

### internal/oracle/testdata/step21_library_types.a (exceptions-21-main)

```text
before: None
after:  | internal/oracle/testdata/step21_library_types.a | 22 | 22 | 35 | 50 | 9 | 0 |
```

### internal/oracle/testdata/step21_liveness.a (exceptions-21-main)

```text
before: None
after:  | internal/oracle/testdata/step21_liveness.a | 12 | 12 | 9 | 17 | 5 | 0 |
```

### internal/oracle/testdata/step21_narrow_flags_use.ts (exceptions-21-main)

```text
before: None
after:  | internal/oracle/testdata/step21_narrow_flags_use.ts | 13 | 13 | 9 | 21 | 6 | 0 |
```

### internal/oracle/testdata/step21_narrow_length_use.ts (exceptions-21-main)

```text
before: None
after:  | internal/oracle/testdata/step21_narrow_length_use.ts | 6 | 6 | 5 | 12 | 3 | 0 |
```

### internal/oracle/testdata/step21_narrow_member_use.ts (exceptions-21-main)

```text
before: None
after:  | internal/oracle/testdata/step21_narrow_member_use.ts | 6 | 4 | 11 | 12 | 5 | 0 |
```

### internal/oracle/testdata/step21_narrow_parameter_use.ts (exceptions-21-main)

```text
before: None
after:  | internal/oracle/testdata/step21_narrow_parameter_use.ts | 6 | 4 | 4 | 6 | 5 | 0 |
```

### internal/oracle/testdata/step21_narrow_wider_parameter.ts (exceptions-21-main)

```text
before: None
after:  | internal/oracle/testdata/step21_narrow_wider_parameter.ts | 6 | 6 | 10 | 13 | 5 | 0 |
```

### internal/oracle/testdata/step21_object_uncaught.a (exceptions-21-main)

```text
before: None
after:  | internal/oracle/testdata/step21_object_uncaught.a | 3 | 3 | 2 | 4 | 3 | 0 |
```

### internal/oracle/testdata/step21_region_payload.a (exceptions-21-main)

```text
before: None
after:  | internal/oracle/testdata/step21_region_payload.a | 3 | 3 | 3 | 5 | 3 | 0 |
```

### internal/oracle/testdata/step21_rethrow.a (exceptions-21-main)

```text
before: None
after:  | internal/oracle/testdata/step21_rethrow.a | 13 | 13 | 22 | 32 | 4 | 0 |
```

### internal/oracle/testdata/step21_saved_error.a (exceptions-21-main)

```text
before: None
after:  | internal/oracle/testdata/step21_saved_error.a | 20 | 20 | 37 | 47 | 7 | 0 |
```

### internal/oracle/testdata/string_too_long.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/string_too_long.a | 0 | 0 | 0 | 0 | 0 | 0 |
after:  | internal/oracle/testdata/string_too_long.a | 1 | 1 | 3 | 3 | 1 | 0 |
```

### internal/oracle/testdata/string_views_throw.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/string_views_throw.a | 55 | 55 | 62 | 76 | 14 | 0 |
after:  | internal/oracle/testdata/string_views_throw.a | 55 | 55 | 65 | 79 | 14 | 0 |
```

### internal/oracle/testdata/switch_case_declarations/dead_zone.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/switch_case_declarations/dead_zone.a | 8 | 5 | 7 | 8 | 5 | 0 |
after:  | internal/oracle/testdata/switch_case_declarations/dead_zone.a | 9 | 9 | 9 | 13 | 5 | 0 |
```

### internal/oracle/testdata/switch_case_declarations/dead_zone_direct.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/switch_case_declarations/dead_zone_direct.a | 3 | 3 | 3 | 5 | 2 | 0 |
after:  | internal/oracle/testdata/switch_case_declarations/dead_zone_direct.a | 4 | 4 | 5 | 7 | 2 | 0 |
```

### internal/oracle/testdata/switch_case_declarations/dead_zone_initializer.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/switch_case_declarations/dead_zone_initializer.a | 3 | 0 | 3 | 1 | 3 | 0 |
after:  | internal/oracle/testdata/switch_case_declarations/dead_zone_initializer.a | 4 | 4 | 5 | 6 | 4 | 0 |
```

### internal/oracle/testdata/switch_case_declarations/dead_zone_write.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/switch_case_declarations/dead_zone_write.a | 5 | 1 | 3 | 2 | 5 | 0 |
after:  | internal/oracle/testdata/switch_case_declarations/dead_zone_write.a | 6 | 6 | 5 | 7 | 5 | 0 |
```

### internal/oracle/testdata/taste_comma.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/taste_comma.a | 51 | 51 | 5 | 54 | 4 | 0 |
after:  | internal/oracle/testdata/taste_comma.a | 51 | 51 | 6 | 55 | 4 | 0 |
```

### internal/oracle/testdata/taste_void.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/taste_void.a | 39 | 39 | 8 | 44 | 4 | 0 |
after:  | internal/oracle/testdata/taste_void.a | 39 | 39 | 9 | 45 | 4 | 0 |
```

### internal/oracle/testdata/typed_arrays_order.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/typed_arrays_order.a | 34 | 34 | 12 | 50 | 8 | 0 |
after:  | internal/oracle/testdata/typed_arrays_order.a | 34 | 34 | 12 | 51 | 8 | 0 |
```

### internal/oracle/testdata/typed_arrays_stats.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/typed_arrays_stats.a | 54 | 54 | 14 | 61 | 10 | 0 |
after:  | internal/oracle/testdata/typed_arrays_stats.a | 54 | 54 | 14 | 66 | 10 | 0 |
```

### internal/oracle/testdata/typed_arrays_views.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/typed_arrays_views.a | 83 | 83 | 40 | 111 | 13 | 0 |
after:  | internal/oracle/testdata/typed_arrays_views.a | 83 | 83 | 40 | 115 | 13 | 0 |
```

### internal/oracle/testdata/typed_arrays_workers_stats.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/typed_arrays_workers_stats.a | 513 | 513 | 82 | 584 | 23 | 0 |
after:  | internal/oracle/testdata/typed_arrays_workers_stats.a | 513 | 513 | 82 | 600 | 23 | 0 |
```

### internal/oracle/testdata/unknown_narrowing.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/unknown_narrowing.a | 31 | 31 | 42 | 63 | 3 | 0 |
after:  | internal/oracle/testdata/unknown_narrowing.a | 31 | 31 | 43 | 64 | 3 | 0 |
```

### internal/oracle/testdata/user_iterators.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/user_iterators.a | 663 | 663 | 455 | 904 | 64 | 0 |
after:  | internal/oracle/testdata/user_iterators.a | 663 | 663 | 466 | 915 | 64 | 0 |
```

### internal/oracle/testdata/user_iterators_rest_tdz.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/user_iterators_rest_tdz.a | 5 | 0 | 5 | 3 | 5 | 0 |
after:  | internal/oracle/testdata/user_iterators_rest_tdz.a | 6 | 6 | 7 | 10 | 6 | 0 |
```

### internal/oracle/testdata/walk.a (exceptions-21-main)

```text
before: | internal/oracle/testdata/walk.a | 222 | 222 | 117 | 288 | 36 | 0 |
after:  | internal/oracle/testdata/walk.a | 222 | 222 | 131 | 302 | 36 | 0 |
```

### stage3/drivers/scanner/probes/cyclic-premature-value/main.a (exceptions-21-main)

```text
before: | stage3/drivers/scanner/probes/cyclic-premature-value/main.a | 0 | 0 | 0 | 0 | 0 | 0 |
after:  | stage3/drivers/scanner/probes/cyclic-premature-value/main.a | 1 | 1 | 2 | 1 | 1 | 0 |
```

### stage3/fixtures/taste/24_proportional_conditions.a (exceptions-21-main)

```text
before: | stage3/fixtures/taste/24_proportional_conditions.a | 43 | 43 | 89 | 131 | 14 | 0 |
after:  | stage3/fixtures/taste/24_proportional_conditions.a | 43 | 43 | 92 | 134 | 14 | 0 |
```

## Tests and commands

Exact commands, process seconds and failures follow. Per-leaf elapsed seconds and skips are in each *-seconds.json.txt; raw output is in matching .jsonl. A runner exit alone is not used as evidence of a package pass.

| Run | Package | Exit | Seconds | Failed tests |
| --- | --- | ---: | ---: | --- |
| guards | cmd/adamic-gate | 1 | 1.121 | TestEveryTestHasOneDescriptiveName |
| guards2 | cmd/adamic-gate | 0 | 0.767 |  |
| guards2 | internal/ir | 0 | 1.619 |  |
| guards2 | internal/native | 1 | 1.875 | TestRuntimeStaticsAreListed |
| guards3 | internal/native | 0 | 1.925 |  |
| guards3 | internal/lower | 0 | 8.273 |  |
| guards3 | internal/oracle | 0 | 2.932 |  |
| own-lower | internal/lower/... | 1 | 108.713 | TestRegExpNativeRefusals |
| own-lower-final | internal/lower/... | 0 | 55.143 |  |
| prepare-archive | internal/native | 0 | 14.674 |  |
| full-native | internal/native/... | 1 | 265.736 | TestSplitTSGoAgreesUnit00, TestRegExpBytecodeTest262Bucket25Of32, TestRegExpBytecodeTest262Bucket24Of32, TestRegExpBytecodeTest262Bucket22Of32, TestRegExpBytecodeTest262Bucket23Of32, TestRegExpBytecodeTest262Bucket19Of32, TestRegExpBytecodeTest262Bucket20Of32, TestRegExpBytecodeTest262Bucket21Of32, TestRegExpBytecodeTest262Bucket17Of32, TestRegExpBytecodeTest262Bucket18Of32, TestRegExpBytecodeTest262Bucket16Of32, TestRegExpBytecodeTest262Bucket14Of32, TestRegExpBytecodeTest262Bucket15Of32, TestRegExpBytecodeTest262Bucket13Of32, TestRegExpBytecodeTest262Bucket12Of32, TestRegExpBytecodeTest262Bucket11Of32, TestRegExpBytecodeTest262Bucket10Of32, TestRegExpBytecodeTest262Bucket9Of32, TestRegExpBytecodeTest262Bucket8Of32, TestRegExpBytecodeTest262Bucket7Of32, TestRegExpBytecodeTest262Bucket6Of32, TestRegExpBytecodeTest262Bucket5Of32, TestRegExpBytecodeTest262Bucket4Of32, TestRegExpBytecodeTest262Bucket3Of32, TestRegExpBytecodeTest262Bucket2Of32, TestRegExpBytecodeTest262Bucket1Of32, TestRegExpBytecodeTest262Bucket0Of32, TestRegExpLintPatternsNode, TestRegExpSearchNode, TestProduct_LeanCensusExports, TestTheLeanCensusFollowsTheBuildAndFailsClosed |
| native-remaining | internal/native/... | 1 | 289.528 | TestSplitAttributedBindingsAgreeWithNode/wasm32-wasi, TestSplitAttributedBindingsAgreeWithNode, TestOrdinaryToPrimitiveMutants/missing-TypeError, TestOrdinaryToPrimitiveMutants, TestRetainedTopLevelCoverage, TestRetainedSplitCoverage/WASI, TestRetainedSplitCoverage |
| native-integration | internal/native | 0 | 35.37 |  |
| prepare-runtimes | internal/oracle | 0 | 64.6 |  |
| own-exceptions | internal/oracle | 0 | 9.573 |  |
| spread-pending | internal/oracle | 1 | 9.416 | TestSpreadUndefinedThrowStopsBeforeNextGetter |
| spread-pending2 | internal/oracle | 0 | 11.74 |  |
| ancestry-control | internal/oracle | 0 | 10.003 |  |
| completion-core | internal/flow/... | 1 | 115.013 | TestFlowProgram____oracle_testdata_regexp_null_narrowed_a_4ec2434b2575, TestFlowProgram____oracle_testdata_narrowed_writes_a_8d04a57f2196, TestFlowProgram____oracle_testdata_narrowed_union_valid_a_c203ac498591, TestFlowProgram____oracle_testdata_narrowed_reads_a_52f71cc82d23, TestFlowProgram____oracle_testdata_narrowed_numbers_a_27bdbba03a02, TestFlowProgram____oracle_testdata_narrowed_methods_a_0adbd2c74bec, TestFlowProgram____oracle_testdata_e4eec87_f2b_union_narrow_number_a_d4b3506dc3af, TestFlowProgram____oracle_testdata_e4eec87_f2_union_narrow_call_a_3342f5a4339a, TestFlowProgram____oracle_testdata_047cb0d_n_arrayindex_a_961148726e38 |
| guards-final | cmd/adamic-gate | 0 | 3.232 |  |
| guards-final | internal/ir | 0 | 10.266 |  |
| guards-final | internal/native | 0 | 3.346 |  |
| guards-final | internal/lower | 0 | 17.088 |  |
| guards-final | internal/oracle | 0 | 5.85 |  |
| completion-core-fixed | internal/flow/... | 0 | 143.882 |  |
| completion-core-fixed | internal/fresh/... | 1 | 57.418 | TestFreshWrites____oracle_testdata_step21_saved_error_a_18d8dd7d556e, TestFreshWrites____oracle_testdata_step21_error_subclasses_a_0975ec294701 |
| completion-other | internal/javascript/... | 0 | 0.77 |  |
| completion-other | internal/fuzz/... | 1 | 49.426 | TestParallelRunnerStringsAgree, TestOwnershipShapes, TestGeneratedProgramsCheckAndLower/seeds-001-005, TestGeneratedProgramsCheckAndLower/seeds-031-035, TestGeneratedProgramsCheckAndLower/seeds-056-060, TestGeneratedProgramsCheckAndLower/seeds-041-045, TestGeneratedProgramsCheckAndLower/seeds-016-020, TestGeneratedProgramsCheckAndLower/seeds-036-040, TestGeneratedProgramsCheckAndLower/seeds-026-030, TestGeneratedProgramsCheckAndLower/seeds-021-025, TestGeneratedProgramsCheckAndLower/seeds-011-015, TestGeneratedProgramsCheckAndLower/seeds-006-010, TestGeneratedProgramsCheckAndLower, TestUndefinedNumbersShapes/seeds-021-025, TestUndefinedNumbersShapes/seeds-036-040, TestUndefinedNumbersShapes/seeds-031-035, TestUndefinedNumbersShapes/seeds-026-030, TestUndefinedNumbersShapes/seeds-011-015, TestUndefinedNumbersShapes/seeds-016-020, TestUndefinedNumbersShapes/seeds-006-010, TestUndefinedNumbersShapes/seeds-001-005, TestUndefinedNumbersShapes, TestFuzzerFeatureMismatchStopsAtLink |
| completion-stage1 | stage1/cohere/css/... | 1 | 46.685 | TestProduct_CSSParserCorpus, TestProduct_CSSPrinterParserOracle, TestProduct_CSSPrinterOracle |
| completion-commands | cmd/adamic/... | 0 | 27.524 |  |
| completion-commands | cmd/adamic-test262/... | 0 | 35.532 |  |
| completion-commands | stage3/fixtures/... | 0 | 114.907 |  |
| completion-oracle | internal/oracle/... | 1 | 656.348 | TestSwitchCaseDeadZoneStop, TestSmokeEntriesAreRegisteredCases, TestReviewProgramsNoLooseFiles, TestReviewProgramsRefuse/fxspptb_57f2d04_with_frozen.a, TestReviewProgramsRefuse/fxspptb_903f25b_try_assign.a, TestReviewProgramsRefuse/fxspptb_9984394_lib_dispatch.a, TestReviewProgramsRefuse, TestParserNamespaceClassRegistrationMutant, TestNodeFSRmForceUndefinedReproducer, TestModuleNamespaceReadinessMutants/early, TestModuleNamespaceReadinessMutants/direct_early, TestModuleNamespaceReadinessMutants, TestLibraryMethodAliasReadMutant, TestReviewProgramsAgreeWithNode_009/fxspptb_047cb0d_n_element_plain.a, TestReviewProgramsAgreeWithNode_010/fxspptb_047cb0d_n_element.a, TestReviewProgramsAgreeWithNode_010, TestReviewProgramsAgreeWithNode_009, TestReviewProgramsAgreeWithNode_004/fxspptb_047cb0d_n_element_method.a, TestReviewProgramsAgreeWithNode_004, TestEnumCleanupMutant, TestSwitchCaseDeclarationMutants/function_not_hoisted, TestSwitchCaseDeclarationMutants, TestLibraryMethodValueMutants_002/parseFloat, TestLibraryMethodValueMutants_002/map_String, TestLibraryMethodValueMutants_002, TestNativeAgreesWithNode_127/internal/oracle/testdata/async_coverage_reject_nested.a, TestNativeAgreesWithNode_127, TestNativeAgreesWithNode_091/internal/oracle/testdata/namespaces_unknown_write_before.a, TestNativeAgreesWithNode_091, TestNativeAgreesWithNode_087/internal/oracle/testdata/element_access_presence.a, TestNativeAgreesWithNode_087, TestNativeAgreesWithNode_078/internal/oracle/testdata/namespaces_unknown_void_before.a, TestNativeAgreesWithNode_078, TestNativeAgreesWithNode_066/internal/oracle/testdata/async_coverage_reject_eager.a, TestNativeAgreesWithNode_066, TestNativeAgreesWithNode_059/internal/oracle/testdata/async_coverage_reject_empty.a, TestNativeAgreesWithNode_059, TestNativeAgreesWithNode_039/internal/oracle/testdata/namespaces_unknown_enum_before.a, TestNativeAgreesWithNode_039, TestNativeAgreesWithNode_031/internal/oracle/testdata/namespaces_unknown_before.a, TestNativeAgreesWithNode_031, TestNativeAgreesWithNode_021/internal/oracle/testdata/async_throw.a, TestNativeAgreesWithNode_020/internal/oracle/testdata/namespaces_unknown_function_before.a, TestNativeAgreesWithNode_021, TestNativeAgreesWithNode_020, TestEnumInitializationUnknownPinned/unknown-before, TestEnumInitializationUnknownPinned/unknown-callback, TestEnumInitializationUnknownPinned/unknown-property, TestEnumInitializationUnknownPinned, TestAsyncFixturesAgreeWithNode/async_throw.a, TestCheckedViewFlagDowncast/tsc_source, TestAsyncFixturesAgreeWithNode/async_coverage_reject_nested.a, TestAsyncFixturesAgreeWithNode/async_coverage_reject_empty.a, TestAsyncFixturesAgreeWithNode/async_coverage_reject_eager.a, TestAsyncFixturesAgreeWithNode, TestCheckedViewFlagDowncast |
| completion-commands | internal/ir/... | 1 | 63.656 | TestCallTargetExportInputsNative, TestCallTargetExportInputsLower, TestCallTargetExportInputsFresh, TestCallTargetExportInputsFlow, TestProduct_CallTargetReaderArchivesNative, TestProduct_CallTargetReaderArchivesLower, TestCallTargetReadersLower, TestProduct_CallTargetReaderArchivesFresh, TestCallTargetReadersFlow, TestCallTargetReadersNative, TestProduct_CallTargetReaderArchivesFlow, TestCallTargetReadersFresh |
| integration-focused | internal/lower | 0 | 8.217 |  |
| integration-focused | internal/oracle | 0 | 9.897 |  |
| integration-focused | internal/fresh | 0 | 2.026 |  |
| integration-core | internal/fresh/... | 0 | 43.753 |  |
| integration-oracle-own | internal/oracle | 0 | 75.001 |  |
| integration-core | internal/fuzz/... | 1 | 125.344 | TestMovesCampaign_2, TestFuzzerFeatureMismatchStopsAtLink |
| integration-lower | internal/lower/... | 1 | 211.223 |  |
| final-ir-complete | internal/ir/... | 0 | 3.196 |  |
| final-native | internal/native/... | 1 | 291.573 | TestTheLeanCensusFollowsTheBuildAndFailsClosed, TestProduct_LeanCensusExports, TestTheRuntimeBuildReachesOnlyItsDeclaredFiles |
| final-core | internal/lower/... | 1 | 85.178 |  |
| final-core | internal/flow/... | 0 | 72.075 |  |
| final-core | internal/fresh/... | 0 | 13.91 |  |
| final-core | internal/fuzz/... | 0 | 32.882 |  |
| final-provenance | internal/native/runtimeproducts/... | 0 | 8.713 |  |
| final-cycleledger | internal/lower/cycleledger/... | 0 | 21.604 |  |
| final-stage1 | stage1/cohere/css/... | 1 | 99.882 | TestCompositionMatchesGo, TestTheCanonicalRangeChecksCanFail, TestClosedParserRegexGap, TestCSSPrinterOptimizedMatchesGo_Default, TestCSSPrinterOptimizedMatchesGo_Narrow, TestCSSParserOptimizedMatchesNode, TestCompositionMatchesGo_Setup, TestThePortParsesAsGoCohereDoes_290, TestThePortParsesAsGoCohereDoes_387, TestThePortParsesAsGoCohereDoes_382, TestThePortParsesAsGoCohereDoes_381, TestThePortParsesAsGoCohereDoes_380, TestThePortParsesAsGoCohereDoes_379, TestThePortParsesAsGoCohereDoes_378, TestThePortParsesAsGoCohereDoes_377, TestThePortParsesAsGoCohereDoes_376, TestThePortParsesAsGoCohereDoes_375, TestThePortParsesAsGoCohereDoes_374, TestThePortParsesAsGoCohereDoes_373, TestThePortParsesAsGoCohereDoes_372, TestThePortParsesAsGoCohereDoes_371, TestThePortParsesAsGoCohereDoes_370, TestThePortParsesAsGoCohereDoes_369, TestThePortParsesAsGoCohereDoes_368, TestThePortParsesAsGoCohereDoes_367, TestThePortParsesAsGoCohereDoes_366, TestThePortParsesAsGoCohereDoes_365, TestThePortParsesAsGoCohereDoes_364, TestThePortParsesAsGoCohereDoes_363, TestThePortParsesAsGoCohereDoes_218, TestThePortParsesAsGoCohereDoes_362, TestThePortParsesAsGoCohereDoes_361, TestThePortParsesAsGoCohereDoes_360, TestThePortParsesAsGoCohereDoes_359, TestThePortParsesAsGoCohereDoes_358, TestThePortParsesAsGoCohereDoes_357, TestThePortParsesAsGoCohereDoes_356, TestThePortParsesAsGoCohereDoes_355, TestThePortParsesAsGoCohereDoes_354, TestThePortParsesAsGoCohereDoes_353, TestThePortParsesAsGoCohereDoes_352, TestThePortParsesAsGoCohereDoes_351, TestThePortParsesAsGoCohereDoes_350, TestThePortParsesAsGoCohereDoes_349, TestThePortParsesAsGoCohereDoes_348, TestThePortParsesAsGoCohereDoes_347, TestThePortParsesAsGoCohereDoes_346, TestThePortParsesAsGoCohereDoes_345, TestThePortParsesAsGoCohereDoes_344, TestThePortParsesAsGoCohereDoes_343, TestThePortParsesAsGoCohereDoes_342, TestThePortParsesAsGoCohereDoes_341, TestThePortParsesAsGoCohereDoes_340, TestThePortParsesAsGoCohereDoes_339, TestThePortParsesAsGoCohereDoes_338, TestThePortParsesAsGoCohereDoes_337, TestThePortParsesAsGoCohereDoes_336, TestThePortParsesAsGoCohereDoes_335, TestThePortParsesAsGoCohereDoes_334, TestThePortParsesAsGoCohereDoes_333, TestThePortParsesAsGoCohereDoes_332, TestThePortParsesAsGoCohereDoes_331, TestThePortParsesAsGoCohereDoes_330, TestThePortParsesAsGoCohereDoes_329, TestThePortParsesAsGoCohereDoes_328, TestThePortParsesAsGoCohereDoes_327, TestThePortParsesAsGoCohereDoes_326, TestThePortParsesAsGoCohereDoes_325, TestThePortParsesAsGoCohereDoes_324, TestThePortParsesAsGoCohereDoes_323, TestThePortParsesAsGoCohereDoes_322, TestThePortParsesAsGoCohereDoes_321, TestThePortParsesAsGoCohereDoes_320, TestThePortParsesAsGoCohereDoes_319, TestThePortParsesAsGoCohereDoes_318, TestThePortParsesAsGoCohereDoes_317, TestThePortParsesAsGoCohereDoes_316, TestThePortParsesAsGoCohereDoes_315, TestThePortParsesAsGoCohereDoes_314, TestThePortParsesAsGoCohereDoes_313, TestThePortParsesAsGoCohereDoes_312, TestThePortParsesAsGoCohereDoes_311, TestThePortParsesAsGoCohereDoes_310, TestThePortParsesAsGoCohereDoes_309, TestThePortParsesAsGoCohereDoes_308, TestThePortParsesAsGoCohereDoes_307, TestThePortParsesAsGoCohereDoes_306, TestThePortParsesAsGoCohereDoes_305, TestThePortParsesAsGoCohereDoes_304, TestThePortParsesAsGoCohereDoes_303, TestThePortParsesAsGoCohereDoes_217, TestThePortParsesAsGoCohereDoes_302, TestThePortParsesAsGoCohereDoes_301, TestThePortParsesAsGoCohereDoes_300, TestThePortParsesAsGoCohereDoes_299, TestThePortParsesAsGoCohereDoes_298, TestThePortParsesAsGoCohereDoes_297, TestThePortParsesAsGoCohereDoes_296, TestThePortParsesAsGoCohereDoes_295, TestThePortParsesAsGoCohereDoes_294, TestThePortParsesAsGoCohereDoes_293, TestThePortParsesAsGoCohereDoes_292, TestThePortParsesAsGoCohereDoes_291, TestThePortParsesAsGoCohereDoes_216, TestThePortParsesAsGoCohereDoes_215, TestThePortParsesAsGoCohereDoes_214, TestThePortParsesAsGoCohereDoes_213, TestThePortParsesAsGoCohereDoes_212, TestThePortParsesAsGoCohereDoes_211, TestThePortParsesAsGoCohereDoes_210, TestThePortParsesAsGoCohereDoes_209, TestThePortParsesAsGoCohereDoes_208, TestThePortParsesAsGoCohereDoes_207, TestThePortParsesAsGoCohereDoes_206, TestThePortParsesAsGoCohereDoes_205, TestThePortParsesAsGoCohereDoes_204, TestThePortParsesAsGoCohereDoes_203, TestThePortParsesAsGoCohereDoes_202, TestThePortParsesAsGoCohereDoes_201, TestThePortParsesAsGoCohereDoes_200, TestThePortParsesAsGoCohereDoes_199, TestThePortParsesAsGoCohereDoes_198, TestThePortParsesAsGoCohereDoes_197, TestThePortParsesAsGoCohereDoes_196, TestThePortParsesAsGoCohereDoes_195, TestThePortParsesAsGoCohereDoes_194, TestThePortParsesAsGoCohereDoes_193, TestThePortParsesAsGoCohereDoes_192, TestThePortParsesAsGoCohereDoes_191, TestThePortParsesAsGoCohereDoes_190, TestThePortParsesAsGoCohereDoes_189, TestThePortParsesAsGoCohereDoes_188, TestThePortParsesAsGoCohereDoes_187, TestThePortParsesAsGoCohereDoes_186, TestThePortParsesAsGoCohereDoes_185, TestThePortParsesAsGoCohereDoes_184, TestThePortParsesAsGoCohereDoes_183, TestThePortParsesAsGoCohereDoes_182, TestThePortParsesAsGoCohereDoes_181, TestThePortParsesAsGoCohereDoes_180, TestThePortParsesAsGoCohereDoes_179, TestThePortParsesAsGoCohereDoes_178, TestThePortParsesAsGoCohereDoes_177, TestThePortParsesAsGoCohereDoes_176, TestThePortParsesAsGoCohereDoes_175, TestThePortParsesAsGoCohereDoes_174, TestThePortParsesAsGoCohereDoes_173, TestThePortParsesAsGoCohereDoes_172, TestThePortParsesAsGoCohereDoes_171, TestThePortParsesAsGoCohereDoes_170, TestThePortParsesAsGoCohereDoes_169, TestThePortParsesAsGoCohereDoes_162, TestThePortParsesAsGoCohereDoes_168, TestThePortParsesAsGoCohereDoes_167, TestThePortParsesAsGoCohereDoes_166, TestThePortParsesAsGoCohereDoes_165, TestThePortParsesAsGoCohereDoes_164, TestThePortParsesAsGoCohereDoes_163, TestThePortParsesAsGoCohereDoes_161, TestThePortParsesAsGoCohereDoes_160, TestThePortParsesAsGoCohereDoes_159, TestThePortParsesAsGoCohereDoes_158, TestThePortParsesAsGoCohereDoes_157, TestThePortParsesAsGoCohereDoes_156, TestThePortParsesAsGoCohereDoes_155, TestThePortParsesAsGoCohereDoes_154, TestThePortParsesAsGoCohereDoes_152, TestThePortParsesAsGoCohereDoes_151, TestThePortParsesAsGoCohereDoes_150, TestThePortParsesAsGoCohereDoes_149, TestThePortParsesAsGoCohereDoes_148, TestThePortParsesAsGoCohereDoes_147, TestThePortParsesAsGoCohereDoes_146, TestThePortParsesAsGoCohereDoes_145, TestThePortParsesAsGoCohereDoes_144, TestThePortParsesAsGoCohereDoes_143, TestThePortParsesAsGoCohereDoes_142, TestThePortParsesAsGoCohereDoes_141, TestThePortParsesAsGoCohereDoes_140, TestThePortParsesAsGoCohereDoes_139, TestThePortParsesAsGoCohereDoes_138, TestThePortParsesAsGoCohereDoes_137, TestThePortParsesAsGoCohereDoes_136, TestThePortParsesAsGoCohereDoes_135, TestThePortParsesAsGoCohereDoes_134, TestThePortParsesAsGoCohereDoes_133, TestThePortParsesAsGoCohereDoes_132, TestThePortParsesAsGoCohereDoes_131, TestThePortParsesAsGoCohereDoes_001, TestThePortParsesAsGoCohereDoes_129, TestThePortParsesAsGoCohereDoes_128, TestThePortParsesAsGoCohereDoes_127, TestThePortParsesAsGoCohereDoes_126, TestThePortParsesAsGoCohereDoes_125, TestThePortParsesAsGoCohereDoes_124, TestThePortParsesAsGoCohereDoes_123, TestThePortParsesAsGoCohereDoes_122, TestThePortParsesAsGoCohereDoes_121, TestThePortParsesAsGoCohereDoes_120, TestThePortParsesAsGoCohereDoes_119, TestThePortParsesAsGoCohereDoes_118, TestThePortParsesAsGoCohereDoes_117, TestThePortParsesAsGoCohereDoes_116, TestThePortParsesAsGoCohereDoes_115, TestThePortParsesAsGoCohereDoes_114, TestThePortParsesAsGoCohereDoes_113, TestThePortParsesAsGoCohereDoes_112, TestThePortParsesAsGoCohereDoes_111, TestThePortParsesAsGoCohereDoes_110, TestThePortParsesAsGoCohereDoes_109, TestThePortParsesAsGoCohereDoes_108, TestThePortParsesAsGoCohereDoes_107, TestThePortParsesAsGoCohereDoes_089, TestThePortParsesAsGoCohereDoes_106, TestThePortParsesAsGoCohereDoes_385, TestThePortParsesAsGoCohereDoes_104, TestThePortParsesAsGoCohereDoes_103, TestThePortParsesAsGoCohereDoes_102, TestThePortParsesAsGoCohereDoes_101, TestThePortParsesAsGoCohereDoes_100, TestThePortParsesAsGoCohereDoes_099, TestThePortParsesAsGoCohereDoes_098, TestThePortParsesAsGoCohereDoes_097, TestThePortParsesAsGoCohereDoes_096, TestThePortParsesAsGoCohereDoes_095, TestThePortParsesAsGoCohereDoes_094, TestThePortParsesAsGoCohereDoes_093, TestThePortParsesAsGoCohereDoes_092, TestThePortParsesAsGoCohereDoes_091, TestThePortParsesAsGoCohereDoes_090, TestThePortParsesAsGoCohereDoes_088, TestThePortParsesAsGoCohereDoes_087, TestThePortParsesAsGoCohereDoes_086, TestThePortParsesAsGoCohereDoes_085, TestThePortParsesAsGoCohereDoes_084, TestThePortParsesAsGoCohereDoes_083, TestThePortParsesAsGoCohereDoes_082, TestThePortParsesAsGoCohereDoes_081, TestThePortParsesAsGoCohereDoes_080, TestThePortParsesAsGoCohereDoes_079, TestThePortParsesAsGoCohereDoes_078, TestThePortParsesAsGoCohereDoes_077, TestThePortParsesAsGoCohereDoes_076, TestThePortParsesAsGoCohereDoes_075, TestThePortParsesAsGoCohereDoes_074, TestThePortParsesAsGoCohereDoes_073, TestThePortParsesAsGoCohereDoes_072, TestThePortParsesAsGoCohereDoes_071, TestThePortParsesAsGoCohereDoes_070, TestThePortParsesAsGoCohereDoes_069, TestThePortParsesAsGoCohereDoes_068, TestThePortParsesAsGoCohereDoes_067, TestThePortParsesAsGoCohereDoes_066, TestThePortParsesAsGoCohereDoes_065, TestThePortParsesAsGoCohereDoes_064, TestThePortParsesAsGoCohereDoes_063, TestThePortParsesAsGoCohereDoes_062, TestThePortParsesAsGoCohereDoes_061, TestThePortParsesAsGoCohereDoes_060, TestThePortParsesAsGoCohereDoes_059, TestThePortParsesAsGoCohereDoes_058, TestThePortParsesAsGoCohereDoes_057, TestThePortParsesAsGoCohereDoes_056, TestThePortParsesAsGoCohereDoes_055, TestThePortParsesAsGoCohereDoes_054, TestThePortParsesAsGoCohereDoes_053, TestThePortParsesAsGoCohereDoes_052, TestThePortParsesAsGoCohereDoes_051, TestThePortParsesAsGoCohereDoes_050, TestThePortParsesAsGoCohereDoes_049, TestThePortParsesAsGoCohereDoes_048, TestThePortParsesAsGoCohereDoes_047, TestThePortParsesAsGoCohereDoes_046, TestThePortParsesAsGoCohereDoes_035, TestThePortParsesAsGoCohereDoes_045, TestThePortParsesAsGoCohereDoes_044, TestThePortParsesAsGoCohereDoes_043, TestThePortParsesAsGoCohereDoes_042, TestThePortParsesAsGoCohereDoes_041, TestThePortParsesAsGoCohereDoes_040, TestThePortParsesAsGoCohereDoes_039, TestThePortParsesAsGoCohereDoes_038, TestThePortParsesAsGoCohereDoes_037, TestThePortParsesAsGoCohereDoes_036, TestThePortParsesAsGoCohereDoes_034, TestThePortParsesAsGoCohereDoes_033, TestThePortParsesAsGoCohereDoes_032, TestThePortParsesAsGoCohereDoes_031, TestThePortParsesAsGoCohereDoes_030, TestThePortParsesAsGoCohereDoes_029, TestThePortParsesAsGoCohereDoes_028, TestThePortParsesAsGoCohereDoes_027, TestThePortParsesAsGoCohereDoes_026, TestThePortParsesAsGoCohereDoes_025, TestThePortParsesAsGoCohereDoes_024, TestThePortParsesAsGoCohereDoes_023, TestThePortParsesAsGoCohereDoes_022, TestThePortParsesAsGoCohereDoes_021, TestThePortParsesAsGoCohereDoes_020, TestThePortParsesAsGoCohereDoes_019, TestThePortParsesAsGoCohereDoes_018, TestThePortParsesAsGoCohereDoes_017, TestThePortParsesAsGoCohereDoes_016, TestThePortParsesAsGoCohereDoes_015, TestThePortParsesAsGoCohereDoes_014, TestThePortParsesAsGoCohereDoes_013, TestThePortParsesAsGoCohereDoes_012, TestThePortParsesAsGoCohereDoes_011, TestThePortParsesAsGoCohereDoes_010, TestThePortParsesAsGoCohereDoes_009, TestThePortParsesAsGoCohereDoes_008, TestThePortParsesAsGoCohereDoes_007, TestThePortParsesAsGoCohereDoes_006, TestThePortParsesAsGoCohereDoes_005, TestThePortParsesAsGoCohereDoes_004, TestThePortParsesAsGoCohereDoes_003, TestThePortParsesAsGoCohereDoes_002, TestThePortParsesAsGoCohereDoes_105, TestThePortParsesAsGoCohereDoes_386, TestThePortParsesAsGoCohereDoes_000, TestThePortParsesAsGoCohereDoes_383, TestThePortParsesAsGoCohereDoes_384, TestThePortParsesAsGoCohereDoes_Setup, TestProduct_CSSCompositionNative, TestProduct_CSSCompositionLowered, TestProduct_CSSParserClosingBraceMutantNative, TestProduct_CSSParserClosingBraceMutantLowered, TestProduct_CSSParserCommentMutantNative, TestProduct_CSSParserCustomPropertyMutantNative, TestProduct_CSSParserCustomPropertyMutantLowered, TestProduct_CSSParserParserNative, TestProduct_CSSParserParserLowered, TestProduct_CSSParserCommentMutantLowered, TestCSSPrinterMutantThatDiffersNowhereFails, TestCSSPrinterMutantWitnessesRekey, TestProduct_CSSPrinterIndentMutantWitnesses, TestProduct_CSSPrinterWidthMutantWitnesses, TestProduct_CSSPrinterSemicolonMutantWitnesses, TestCSSPrinterShardingCatchesDisagreement, TestCSSPrinterByteSlicesMatchGo, TestCSSPrinterAgreesWithGo_087, TestCSSPrinterAgreesWithGo_070, TestProduct_CSSPrinterWidthMutant, TestProduct_CSSPrinterIndentMutant, TestProduct_CSSPrinterSemicolonMutant, TestProduct_CSSPrinterSanitizedAndLowered, TestCSSParserMutantThatDiffersNowhereFails, TestCSSPrinterPreflightCost, TestProduct_CSSParserClosingBraceMutantWitnesses, TestProduct_CSSParserCommentMutantWitnesses, TestProduct_CSSParserCustomPropertyMutantWitnesses, TestCSSParserMutantWitnessesRekey, TestThePortParsesAsGoCohereDoesUnion, TestCSSParserSanitizedRunCatchesALeak, TestProduct_CSSParserReleaseMutant, TestCSSPrinterPreflightCostMutant, TestProduct_CSSPrinterPrefixCostMutant, TestProduct_CSSParserOptimizedNative, TestProduct_CSSPrinterOptimizedNative, TestCSSParserPlantedDisagreement, TestComposedMemoryChecksCanFail_LeakSanitizer, TestProduct_CSSComposedMemoryLeakMutant, TestComposedMemoryChecksCanFail_UndefinedBehaviorSanitizer, TestProduct_CSSComposedMemoryUndefinedBehaviorMutant, TestComposedMemoryChecksCanFail_AddressSanitizer, TestCSSPrinterAgreesWithGo_133, TestProduct_CSSComposedMemoryAddressMutant, TestCSSPrinterAgreesWithGo_131, TestCSSPrinterAgreesWithGo_130, TestCSSPrinterAgreesWithGo_132, TestCSSPrinterAgreesWithGo_129, TestCSSPrinterAgreesWithGo_128, TestCSSPrinterAgreesWithGo_126, TestCSSPrinterAgreesWithGo_127, TestCSSPrinterAgreesWithGo_125, TestCSSPrinterAgreesWithGo_123, TestCSSPrinterAgreesWithGo_124, TestCSSPrinterAgreesWithGo_122, TestCSSPrinterAgreesWithGo_120, TestCSSPrinterAgreesWithGo_121, TestCSSPrinterAgreesWithGo_119, TestCSSPrinterAgreesWithGo_116, TestCSSPrinterAgreesWithGo_118, TestCSSPrinterAgreesWithGo_117, TestCSSPrinterAgreesWithGo_115, TestCSSPrinterAgreesWithGo_113, TestCSSPrinterAgreesWithGo_114, TestCSSPrinterAgreesWithGo_111, TestCSSPrinterAgreesWithGo_112, TestCSSPrinterAgreesWithGo_109, TestCSSPrinterAgreesWithGo_110, TestCSSPrinterAgreesWithGo_107, TestCSSPrinterAgreesWithGo_106, TestCSSPrinterAgreesWithGo_105, TestCSSPrinterAgreesWithGo_108, TestCSSPrinterAgreesWithGo_104, TestCSSPrinterAgreesWithGo_103, TestCSSPrinterAgreesWithGo_102, TestCSSPrinterAgreesWithGo_101, TestCSSPrinterAgreesWithGo_099, TestCSSPrinterAgreesWithGo_100, TestCSSPrinterAgreesWithGo_098, TestCSSPrinterAgreesWithGo_097, TestCSSPrinterAgreesWithGo_096, TestCSSPrinterAgreesWithGo_095, TestCSSPrinterAgreesWithGo_093, TestCSSPrinterAgreesWithGo_094, TestCSSPrinterAgreesWithGo_092, TestCSSPrinterAgreesWithGo_090, TestCSSPrinterAgreesWithGo_091, TestCSSPrinterAgreesWithGo_089, TestCSSPrinterAgreesWithGo_088, TestCSSPrinterAgreesWithGo_086, TestCSSPrinterAgreesWithGo_030, TestCSSPrinterAgreesWithGo_085, TestCSSPrinterAgreesWithGo_082, TestCSSPrinterAgreesWithGo_084, TestCSSPrinterAgreesWithGo_083, TestCSSPrinterAgreesWithGo_080, TestCSSPrinterAgreesWithGo_081, TestCSSPrinterAgreesWithGo_079, TestCSSPrinterAgreesWithGo_076, TestCSSPrinterAgreesWithGo_077, TestCSSPrinterAgreesWithGo_078, TestCSSPrinterAgreesWithGo_075, TestCSSPrinterAgreesWithGo_074, TestCSSPrinterAgreesWithGo_072, TestCSSPrinterAgreesWithGo_071, TestCSSPrinterAgreesWithGo_073, TestCSSPrinterAgreesWithGo_069, TestCSSPrinterAgreesWithGo_066, TestCSSPrinterAgreesWithGo_065, TestCSSPrinterAgreesWithGo_067, TestCSSPrinterAgreesWithGo_068, TestCSSPrinterAgreesWithGo_064, TestCSSPrinterAgreesWithGo_063, TestCSSPrinterAgreesWithGo_062, TestCSSPrinterAgreesWithGo_061, TestCSSPrinterAgreesWithGo_060, TestCSSPrinterAgreesWithGo_057, TestCSSPrinterAgreesWithGo_059, TestCSSPrinterAgreesWithGo_058, TestCSSPrinterAgreesWithGo_054, TestCSSPrinterAgreesWithGo_056, TestCSSPrinterAgreesWithGo_055, TestCSSPrinterAgreesWithGo_053, TestCSSPrinterAgreesWithGo_052, TestCSSPrinterAgreesWithGo_049, TestCSSPrinterAgreesWithGo_050, TestCSSPrinterAgreesWithGo_048, TestCSSPrinterAgreesWithGo_051, TestCSSPrinterAgreesWithGo_047, TestCSSPrinterAgreesWithGo_046, TestCSSPrinterAgreesWithGo_043, TestCSSPrinterAgreesWithGo_044, TestCSSPrinterAgreesWithGo_045, TestCSSPrinterAgreesWithGo_042, TestCSSPrinterAgreesWithGo_040, TestCSSPrinterAgreesWithGo_038, TestCSSPrinterAgreesWithGo_039, TestCSSPrinterAgreesWithGo_041, TestCSSPrinterAgreesWithGo_036, TestCSSPrinterAgreesWithGo_034, TestCSSPrinterAgreesWithGo_035, TestCSSPrinterAgreesWithGo_033, TestCSSPrinterAgreesWithGo_037, TestCSSPrinterAgreesWithGo_002, TestCSSPrinterAgreesWithGo_032, TestCSSPrinterAgreesWithGo_029, TestCSSPrinterAgreesWithGo_028, TestCSSPrinterAgreesWithGo_031, TestCSSPrinterAgreesWithGo_026, TestCSSPrinterAgreesWithGo_025, TestCSSPrinterAgreesWithGo_024, TestCSSPrinterAgreesWithGo_027, TestCSSPrinterAgreesWithGo_023, TestCSSPrinterAgreesWithGo_022, TestCSSPrinterAgreesWithGo_020, TestCSSPrinterAgreesWithGo_021, TestCSSPrinterAgreesWithGo_016, TestCSSPrinterAgreesWithGo_017, TestCSSPrinterAgreesWithGo_019, TestCSSPrinterAgreesWithGo_018, TestCSSPrinterAgreesWithGo_013, TestCSSPrinterAgreesWithGo_014, TestCSSPrinterAgreesWithGo_015, TestCSSPrinterAgreesWithGo_012, TestCSSPrinterAgreesWithGo_010, TestCSSPrinterAgreesWithGo_011, TestCSSPrinterAgreesWithGo_009, TestCSSPrinterAgreesWithGo_008, TestCSSPrinterAgreesWithGo_006, TestCSSPrinterAgreesWithGo_007, TestCompositionMatchesGo_008, TestCSSPrinterAgreesWithGo_004, TestCSSPrinterAgreesWithGo_005, TestCSSPrinterAgreesWithGoUnion, TestCSSPrinterAgreesWithGo_003, TestCSSPrinterAgreesWithGo_001, TestCSSPrinterSanitizedRunNamesEachFailure, TestCompositionMatchesGo_015, TestCompositionMatchesGo_014, TestCompositionMatchesGo_013, TestCompositionMatchesGo_012, TestCompositionMatchesGo_011, TestCompositionMatchesGo_010, TestCompositionMatchesGo_009, TestThePortParsesAsGoCohereDoes_255, TestThePortParsesAsGoCohereDoes_289, TestThePortParsesAsGoCohereDoes_288, TestThePortParsesAsGoCohereDoes_287, TestThePortParsesAsGoCohereDoes_286, TestThePortParsesAsGoCohereDoes_285, TestThePortParsesAsGoCohereDoes_284, TestThePortParsesAsGoCohereDoes_283, TestThePortParsesAsGoCohereDoes_282, TestThePortParsesAsGoCohereDoes_281, TestThePortParsesAsGoCohereDoes_280, TestThePortParsesAsGoCohereDoes_279, TestThePortParsesAsGoCohereDoes_278, TestThePortParsesAsGoCohereDoes_277, TestThePortParsesAsGoCohereDoes_276, TestThePortParsesAsGoCohereDoes_275, TestThePortParsesAsGoCohereDoes_274, TestThePortParsesAsGoCohereDoes_273, TestThePortParsesAsGoCohereDoes_272, TestThePortParsesAsGoCohereDoes_271, TestThePortParsesAsGoCohereDoes_270, TestThePortParsesAsGoCohereDoes_269, TestThePortParsesAsGoCohereDoes_268, TestThePortParsesAsGoCohereDoes_267, TestThePortParsesAsGoCohereDoes_266, TestThePortParsesAsGoCohereDoes_265, TestThePortParsesAsGoCohereDoes_264, TestThePortParsesAsGoCohereDoes_263, TestThePortParsesAsGoCohereDoes_262, TestThePortParsesAsGoCohereDoes_261, TestThePortParsesAsGoCohereDoes_260, TestThePortParsesAsGoCohereDoes_130, TestThePortParsesAsGoCohereDoes_259, TestThePortParsesAsGoCohereDoes_258, TestThePortParsesAsGoCohereDoes_257, TestThePortParsesAsGoCohereDoes_256, TestThePortParsesAsGoCohereDoes_237, TestThePortParsesAsGoCohereDoes_254, TestThePortParsesAsGoCohereDoes_253, TestThePortParsesAsGoCohereDoes_252, TestThePortParsesAsGoCohereDoes_251, TestThePortParsesAsGoCohereDoes_250, TestThePortParsesAsGoCohereDoes_249, TestThePortParsesAsGoCohereDoes_248, TestThePortParsesAsGoCohereDoes_247, TestThePortParsesAsGoCohereDoes_246, TestThePortParsesAsGoCohereDoes_245, TestThePortParsesAsGoCohereDoes_244, TestThePortParsesAsGoCohereDoes_243, TestThePortParsesAsGoCohereDoes_242, TestThePortParsesAsGoCohereDoes_241, TestThePortParsesAsGoCohereDoes_240, TestThePortParsesAsGoCohereDoes_239, TestThePortParsesAsGoCohereDoes_238, TestThePortParsesAsGoCohereDoes_228, TestThePortParsesAsGoCohereDoes_236, TestThePortParsesAsGoCohereDoes_235, TestThePortParsesAsGoCohereDoes_234, TestThePortParsesAsGoCohereDoes_233, TestThePortParsesAsGoCohereDoes_232, TestThePortParsesAsGoCohereDoes_231, TestThePortParsesAsGoCohereDoes_230, TestThePortParsesAsGoCohereDoes_229, TestThePortParsesAsGoCohereDoes_223, TestThePortParsesAsGoCohereDoes_227, TestThePortParsesAsGoCohereDoes_226, TestThePortParsesAsGoCohereDoes_225, TestThePortParsesAsGoCohereDoes_224, TestThePortParsesAsGoCohereDoes_221, TestThePortParsesAsGoCohereDoes_222, TestThePortParsesAsGoCohereDoes_220, TestThePortParsesAsGoCohereDoes_219, TestCompositionMatchesGo_001, TestCompositionMatchesGo_007, TestCompositionMatchesGo_006, TestCompositionMatchesGo_005, TestCompositionMatchesGo_004, TestCompositionMatchesGo_003, TestCompositionMatchesGo_002, TestCompositionMatchesGoCatchesMutant_2, TestCompositionMatchesGo_000, TestCompositionMatchesGoCatchesMutant_0, TestCompositionMatchesGoCatchesMutant_1, TestCompositionMatchesGoUnion, TestThePortParsesAsGoCohereDoes_153, TestCSSPrinterAgreesWithGo_000 |
| final-stage1 | stage1/cohere/graphql/... | 1 | 53.343 | TestThePortParsesAsGoCohereDoes, TestProduct_GraphQLFindsGoOracle, TestPrinterMutants_002, TestPrinterShardPlantedDisagreement, TestPrinterMutants_001, TestPrinterMutants_000, TestProduct_GraphQLPrinterFindsGoOracle, TestPrinterAsGoCohere_011, TestPrinterAsGoCohere_007, TestPrinterAsGoCohere_010, TestPrinterAsGoCohere_009, TestPrinterAsGoCohere_008, TestPrinterAsGoCohere_005, TestPrinterAsGoCohere_006, TestPrinterMutantUnion, TestPrinterAsGoCohere_004, TestProduct_GraphQLPrinterFindsNarrow, TestPrinterAsGoCohere_002, TestPrinterAsGoCohere_001, TestPrinterAsGoCohere_000, TestPrinterAsGoCohere_Union, TestProduct_GraphQLPrinterFindsTabs, TestProduct_GraphQLPrinterFindsTight, TestProduct_GraphQLPrinterFindsDefaults, TestPrinterAsGoCohere_003 |
| final-stage1 | stage1/cohere/json/... | 1 | 65.244 | TestPortMatchesGoCohereClassOrder_000, TestPortMatchesGoCohereClassOrder_001, TestPortMatchesGoCohereInventory, TestCachedWidthMutantIsCaught, TestCorpusPinBeforeSampling, TestPortMatchesGoCohere_0118, TestPortMatchesGoCohere_0117, TestPortMatchesGoCohere_0116, TestPortMatchesGoCohere_0115, TestPortMatchesGoCohere_0114, TestPortMatchesGoCohere_0113, TestPortMatchesGoCohere_0112, TestPortMatchesGoCohere_0111, TestPortMatchesGoCohere_0110, TestPortMatchesGoCohere_0109, TestPortMatchesGoCohere_0108, TestPortMatchesGoCohere_0107, TestPortMatchesGoCohere_0106, TestPortMatchesGoCohere_0105, TestPortMatchesGoCohere_0104, TestPortMatchesGoCohere_0103, TestPortMatchesGoCohere_0102, TestPortMatchesGoCohere_0101, TestPortMatchesGoCohere_0100, TestPortMatchesGoCohere_0099, TestPortMatchesGoCohere_0098, TestPortMatchesGoCohere_0097, TestPortMatchesGoCohere_0096, TestPortMatchesGoCohere_0095, TestPortMatchesGoCohere_0094, TestPortMatchesGoCohere_0093, TestPortMatchesGoCohere_0092, TestPortMatchesGoCohere_0091, TestPortMatchesGoCohere_0090, TestPortMatchesGoCohere_0089, TestProduct_JsonFindsUpstreamOracle, TestPortMatchesGoCohere_0088, TestPortMatchesGoCohere_0087, TestPortMatchesGoCohere_0086, TestPortMatchesGoCohere_0085, TestPortMatchesGoCohere_0084, TestPortMatchesGoCohere_0083, TestPortMatchesGoCohere_0082, TestPortMatchesGoCohere_0081, TestPortMatchesGoCohere_0080, TestPortMatchesGoCohere_0079, TestPortMatchesGoCohere_0078, TestPortMatchesGoCohere_0077, TestPortMatchesGoCohere_0076, TestPortMatchesGoCohere_0075, TestPortMatchesGoCohere_0074, TestPortMatchesGoCohere_0073, TestPortMatchesGoCohere_0072, TestPortMatchesGoCohere_0071, TestPortMatchesGoCohere_0070, TestPortMatchesGoCohere_0069, TestPortMatchesGoCohere_0068, TestPortMatchesGoCohere_0067, TestPortMatchesGoCohere_0066, TestPortMatchesGoCohere_0065, TestPortMatchesGoCohere_0064, TestPortMatchesGoCohere_0063, TestPortMatchesGoCohere_0062, TestPortMatchesGoCohere_0061, TestPortMatchesGoCohere_0060, TestPortMatchesGoCohere_0059, TestPortMatchesGoCohere_0058, TestPortMatchesGoCohere_0057, TestPortMatchesGoCohere_0056, TestPortMatchesGoCohere_0055, TestPortMatchesGoCohere_0054, TestPortMatchesGoCohere_0053, TestPortMatchesGoCohere_0052, TestPortMatchesGoCohere_0051, TestPortMatchesGoCohere_0050, TestPortMatchesGoCohere_0049, TestPortMatchesGoCohere_0048, TestPortMatchesGoCohere_0047, TestPortMatchesGoCohere_0046, TestPortMatchesGoCohere_0045, TestPortMatchesGoCohere_0044, TestPortMatchesGoCohere_0043, TestPortMatchesGoCohere_0042, TestPortMatchesGoCohere_0041, TestPortMatchesGoCohere_0040, TestPortMatchesGoCohere_0039, TestPortMatchesGoCohere_0038, TestPortMatchesGoCohere_0037, TestPortMatchesGoCohere_0036, TestPortMatchesGoCohere_0035, TestPortMatchesGoCohere_0034, TestPortMatchesGoCohere_0033, TestPortMatchesGoCohere_0032, TestPortMatchesGoCohere_0031, TestPortMatchesGoCohere_0030, TestPortMatchesGoCohere_0029, TestPortMatchesGoCohere_0028, TestPortMatchesGoCohere_0027, TestPortMatchesGoCohere_0026, TestPortMatchesGoCohere_0025, TestPortMatchesGoCohere_0024, TestPortMatchesGoCohere_0023, TestPortMatchesGoCohere_0022, TestPortMatchesGoCohere_0021, TestPortMatchesGoCohere_0020, TestPortMatchesGoCohere_0019, TestPortMatchesGoCohere_0018, TestPortMatchesGoCohere_0017, TestPortMatchesGoCohere_0016, TestPortMatchesGoCohere_0015, TestPortMatchesGoCohere_0014, TestPortMatchesGoCohere_0013, TestPortMatchesGoCohere_0012, TestPortMatchesGoCohere_0011, TestPortMatchesGoCohere_0010, TestPortMatchesGoCohere_0009, TestPortMatchesGoCohere_0008, TestPortMatchesGoCohere_0007, TestPortMatchesGoCohere_0006, TestPortMatchesGoCohere_0005, TestPortMatchesGoCohere_0004, TestPortMatchesGoCohere_0003, TestPortMatchesGoCohere_0002, TestPortMatchesGoCohere_0001, TestPortMatchesGoCohere_0000, TestPortMatchesGoCohere_1923, TestPortMatchesGoCohere_1922, TestPortMatchesGoCohere_1921, TestPortMatchesGoCohere_1920, TestPortMatchesGoCohere_1919, TestPortMatchesGoCohere_1918, TestPortMatchesGoCohere_1917, TestPortMatchesGoCohere_1916, TestPortMatchesGoCohere_1915, TestPortMatchesGoCohere_1914, TestPortMatchesGoCohere_1913, TestPortMatchesGoCohere_1912, TestPortMatchesGoCohere_1911, TestPortMatchesGoCohere_1910, TestPortMatchesGoCohere_1909, TestPortMatchesGoCohere_1908, TestPortMatchesGoCohere_1907, TestPortMatchesGoCohere_1906, TestPortMatchesGoCohere_1905, TestPortMatchesGoCohere_1904, TestPortMatchesGoCohere_1903, TestPortMatchesGoCohere_1902, TestPortMatchesGoCohere_1901, TestPortMatchesGoCohere_1900, TestPortMatchesGoCohere_1899, TestPortMatchesGoCohere_1898, TestPortMatchesGoCohere_1897, TestPortMatchesGoCohere_1896, TestPortMatchesGoCohere_1895, TestPortMatchesGoCohere_1894, TestPortMatchesGoCohere_1893, TestPortMatchesGoCohere_1892, TestPortMatchesGoCohere_1891, TestPortMatchesGoCohere_1890, TestPortMatchesGoCohere_1889, TestPortMatchesGoCohere_1888, TestPortMatchesGoCohere_1887, TestPortMatchesGoCohere_1886, TestPortMatchesGoCohere_1885, TestPortMatchesGoCohere_1884, TestPortMatchesGoCohere_1883, TestPortMatchesGoCohere_1882, TestPortMatchesGoCohere_1881, TestPortMatchesGoCohere_1880, TestPortMatchesGoCohere_1795, TestPortMatchesGoCohere_1879, TestPortMatchesGoCohere_1878, TestPortMatchesGoCohere_1877, TestPortMatchesGoCohere_1876, TestPortMatchesGoCohere_1875, TestPortMatchesGoCohere_1874, TestPortMatchesGoCohere_1873, TestPortMatchesGoCohere_1872, TestPortMatchesGoCohere_1871, TestPortMatchesGoCohere_1870, TestPortMatchesGoCohere_1869, TestPortMatchesGoCohere_1868, TestPortMatchesGoCohere_1867, TestPortMatchesGoCohere_1866, TestPortMatchesGoCohere_1865, TestPortMatchesGoCohere_1864, TestPortMatchesGoCohere_1863, TestPortMatchesGoCohere_1862, TestPortMatchesGoCohere_1861, TestPortMatchesGoCohere_1860, TestPortMatchesGoCohere_1859, TestPortMatchesGoCohere_1858, TestPortMatchesGoCohere_1857, TestPortMatchesGoCohere_1856, TestPortMatchesGoCohere_1855, TestPortMatchesGoCohere_1854, TestPortMatchesGoCohere_1853, TestPortMatchesGoCohere_1852, TestPortMatchesGoCohere_1851, TestPortMatchesGoCohere_1850, TestPortMatchesGoCohere_1849, TestPortMatchesGoCohere_1848, TestPortMatchesGoCohere_1847, TestPortMatchesGoCohere_1846, TestPortMatchesGoCohere_1845, TestPortMatchesGoCohere_1844, TestPortMatchesGoCohere_1843, TestPortMatchesGoCohere_1842, TestPortMatchesGoCohere_1841, TestPortMatchesGoCohere_1840, TestPortMatchesGoCohere_1839, TestPortMatchesGoCohere_1838, TestPortMatchesGoCohere_1837, TestPortMatchesGoCohere_1836, TestPortMatchesGoCohere_1835, TestPortMatchesGoCohere_1834, TestPortMatchesGoCohere_1833, TestPortMatchesGoCohere_1832, TestPortMatchesGoCohere_1831, TestPortMatchesGoCohere_1830, TestPortMatchesGoCohere_1829, TestPortMatchesGoCohere_1828, TestPortMatchesGoCohere_1827, TestPortMatchesGoCohere_1826, TestPortMatchesGoCohere_1825, TestPortMatchesGoCohere_1824, TestPortMatchesGoCohere_1823, TestPortMatchesGoCohere_1822, TestPortMatchesGoCohere_1821, TestPortMatchesGoCohere_1820, TestPortMatchesGoCohere_1794, TestPortMatchesGoCohere_1819, TestPortMatchesGoCohere_1818, TestPortMatchesGoCohere_1817, TestPortMatchesGoCohere_1816, TestPortMatchesGoCohere_1815, TestPortMatchesGoCohere_1814, TestPortMatchesGoCohere_1813, TestPortMatchesGoCohere_1812, TestPortMatchesGoCohere_1811, TestPortMatchesGoCohere_1810, TestPortMatchesGoCohere_1809, TestPortMatchesGoCohere_1808, TestPortMatchesGoCohere_1807, TestPortMatchesGoCohere_1806, TestPortMatchesGoCohere_1805, TestPortMatchesGoCohere_1804, TestPortMatchesGoCohere_1803, TestPortMatchesGoCohere_1802, TestPortMatchesGoCohere_1801, TestPortMatchesGoCohere_1800, TestPortMatchesGoCohere_1799, TestPortMatchesGoCohere_1798, TestPortMatchesGoCohere_1797, TestPortMatchesGoCohere_1667, TestPortMatchesGoCohere_1793, TestPortMatchesGoCohere_1792, TestPortMatchesGoCohere_1791, TestPortMatchesGoCohere_1790, TestPortMatchesGoCohere_1789, TestPortMatchesGoCohere_1788, TestPortMatchesGoCohere_1787, TestPortMatchesGoCohere_1786, TestPortMatchesGoCohere_1785, TestPortMatchesGoCohere_1784, TestPortMatchesGoCohere_1783, TestPortMatchesGoCohere_1782, TestPortMatchesGoCohere_1781, TestPortMatchesGoCohere_1780, TestPortMatchesGoCohere_1779, TestPortMatchesGoCohere_1778, TestPortMatchesGoCohere_1777, TestPortMatchesGoCohere_1776, TestPortMatchesGoCohere_1775, TestPortMatchesGoCohere_1774, TestPortMatchesGoCohere_1773, TestPortMatchesGoCohere_1772, TestPortMatchesGoCohere_1771, TestPortMatchesGoCohere_1770, TestPortMatchesGoCohere_1769, TestPortMatchesGoCohere_1768, TestPortMatchesGoCohere_1767, TestPortMatchesGoCohere_1766, TestPortMatchesGoCohere_1765, TestPortMatchesGoCohere_1764, TestPortMatchesGoCohere_1763, TestPortMatchesGoCohere_1762, TestPortMatchesGoCohere_1761, TestPortMatchesGoCohere_1760, TestPortMatchesGoCohere_1759, TestPortMatchesGoCohere_1758, TestPortMatchesGoCohere_1665, TestPortMatchesGoCohere_1757, TestPortMatchesGoCohere_1756, TestPortMatchesGoCohere_1755, TestPortMatchesGoCohere_1754, TestPortMatchesGoCohere_1753, TestPortMatchesGoCohere_1752, TestPortMatchesGoCohere_1751, TestPortMatchesGoCohere_1750, TestPortMatchesGoCohere_1749, TestPortMatchesGoCohere_1748, TestPortMatchesGoCohere_1747, TestPortMatchesGoCohere_1746, TestPortMatchesGoCohere_1745, TestPortMatchesGoCohere_1744, TestPortMatchesGoCohere_1743, TestPortMatchesGoCohere_1742, TestPortMatchesGoCohere_1741, TestPortMatchesGoCohere_1740, TestPortMatchesGoCohere_1739, TestPortMatchesGoCohere_1738, TestPortMatchesGoCohere_1737, TestPortMatchesGoCohere_1736, TestPortMatchesGoCohere_1735, TestPortMatchesGoCohere_1734, TestPortMatchesGoCohere_1733, TestPortMatchesGoCohere_1732, TestPortMatchesGoCohere_1731, TestPortMatchesGoCohere_1730, TestPortMatchesGoCohere_1729, TestPortMatchesGoCohere_1728, TestPortMatchesGoCohere_1727, TestPortMatchesGoCohere_1726, TestPortMatchesGoCohere_1725, TestPortMatchesGoCohere_1724, TestPortMatchesGoCohere_1723, TestPortMatchesGoCohere_1722, TestPortMatchesGoCohere_1721, TestPortMatchesGoCohere_1720, TestPortMatchesGoCohere_1719, TestPortMatchesGoCohere_1718, TestPortMatchesGoCohere_1717, TestPortMatchesGoCohere_1716, TestPortMatchesGoCohere_1715, TestPortMatchesGoCohere_1714, TestPortMatchesGoCohere_1713, TestPortMatchesGoCohere_1712, TestPortMatchesGoCohere_1711, TestPortMatchesGoCohere_1710, TestPortMatchesGoCohere_1709, TestPortMatchesGoCohere_1708, TestPortMatchesGoCohere_1707, TestPortMatchesGoCohere_1706, TestPortMatchesGoCohere_1705, TestPortMatchesGoCohere_1704, TestPortMatchesGoCohere_1703, TestPortMatchesGoCohere_1702, TestPortMatchesGoCohere_1701, TestPortMatchesGoCohere_1700, TestPortMatchesGoCohere_1699, TestPortMatchesGoCohere_1698, TestPortMatchesGoCohere_1664, TestPortMatchesGoCohere_1697, TestPortMatchesGoCohere_1696, TestPortMatchesGoCohere_1695, TestPortMatchesGoCohere_1694, TestPortMatchesGoCohere_1693, TestPortMatchesGoCohere_1692, TestPortMatchesGoCohere_1691, TestPortMatchesGoCohere_1690, TestPortMatchesGoCohere_1689, TestPortMatchesGoCohere_1688, TestPortMatchesGoCohere_1687, TestPortMatchesGoCohere_1686, TestPortMatchesGoCohere_1685, TestPortMatchesGoCohere_1684, TestPortMatchesGoCohere_1683, TestPortMatchesGoCohere_1682, TestPortMatchesGoCohere_1681, TestPortMatchesGoCohere_1680, TestPortMatchesGoCohere_1679, TestPortMatchesGoCohere_1678, TestPortMatchesGoCohere_1677, TestPortMatchesGoCohere_1676, TestPortMatchesGoCohere_1675, TestPortMatchesGoCohere_1674, TestPortMatchesGoCohere_1673, TestPortMatchesGoCohere_1672, TestPortMatchesGoCohere_1671, TestPortMatchesGoCohere_1670, TestPortMatchesGoCohere_1669, TestPortMatchesGoCohere_1668, TestPortMatchesGoCohere_1538, TestPortMatchesGoCohere_1666, TestPortMatchesGoCohere_1663, TestPortMatchesGoCohere_1662, TestPortMatchesGoCohere_1661, TestPortMatchesGoCohere_1660, TestPortMatchesGoCohere_1659, TestPortMatchesGoCohere_1658, TestPortMatchesGoCohere_1657, TestPortMatchesGoCohere_1656, TestPortMatchesGoCohere_1655, TestPortMatchesGoCohere_1654, TestPortMatchesGoCohere_1653, TestPortMatchesGoCohere_1652, TestPortMatchesGoCohere_1651, TestPortMatchesGoCohere_1650, TestPortMatchesGoCohere_1649, TestPortMatchesGoCohere_1648, TestPortMatchesGoCohere_1647, TestPortMatchesGoCohere_1646, TestPortMatchesGoCohere_1645, TestPortMatchesGoCohere_1644, TestPortMatchesGoCohere_1643, TestPortMatchesGoCohere_1642, TestPortMatchesGoCohere_1641, TestPortMatchesGoCohere_1640, TestPortMatchesGoCohere_1639, TestPortMatchesGoCohere_1638, TestPortMatchesGoCohere_1637, TestPortMatchesGoCohere_1636, TestPortMatchesGoCohere_1535, TestPortMatchesGoCohere_1635, TestPortMatchesGoCohere_1634, TestPortMatchesGoCohere_1633, TestPortMatchesGoCohere_1632, TestPortMatchesGoCohere_1631, TestPortMatchesGoCohere_1630, TestPortMatchesGoCohere_1629, TestPortMatchesGoCohere_1628, TestPortMatchesGoCohere_1627, TestPortMatchesGoCohere_1626, TestPortMatchesGoCohere_1625, TestPortMatchesGoCohere_1624, TestPortMatchesGoCohere_1623, TestPortMatchesGoCohere_1622, TestPortMatchesGoCohere_1621, TestPortMatchesGoCohere_1620, TestPortMatchesGoCohere_1619, TestPortMatchesGoCohere_1618, TestPortMatchesGoCohere_1617, TestPortMatchesGoCohere_1616, TestPortMatchesGoCohere_1615, TestPortMatchesGoCohere_1614, TestPortMatchesGoCohere_1613, TestPortMatchesGoCohere_1612, TestPortMatchesGoCohere_1611, TestPortMatchesGoCohere_1610, TestPortMatchesGoCohere_1609, TestPortMatchesGoCohere_1608, TestPortMatchesGoCohere_1607, TestPortMatchesGoCohere_1606, TestPortMatchesGoCohere_1605, TestPortMatchesGoCohere_1604, TestPortMatchesGoCohere_1603, TestPortMatchesGoCohere_1602, TestPortMatchesGoCohere_1601, TestPortMatchesGoCohere_1600, TestPortMatchesGoCohere_1599, TestPortMatchesGoCohere_1598, TestPortMatchesGoCohere_1597, TestPortMatchesGoCohere_1596, TestPortMatchesGoCohere_1595, TestPortMatchesGoCohere_1594, TestPortMatchesGoCohere_1593, TestPortMatchesGoCohere_1592, TestPortMatchesGoCohere_1591, TestPortMatchesGoCohere_1590, TestPortMatchesGoCohere_1589, TestPortMatchesGoCohere_1588, TestPortMatchesGoCohere_1587, TestPortMatchesGoCohere_1586, TestPortMatchesGoCohere_1585, TestPortMatchesGoCohere_1584, TestPortMatchesGoCohere_1583, TestPortMatchesGoCohere_1582, TestPortMatchesGoCohere_1581, TestPortMatchesGoCohere_1580, TestPortMatchesGoCohere_1579, TestPortMatchesGoCohere_1578, TestPortMatchesGoCohere_1577, TestPortMatchesGoCohere_1532, TestPortMatchesGoCohere_1531, TestPortMatchesGoCohere_1533, TestPortMatchesGoCohere_1529, TestPortMatchesGoCohere_1534, TestPortMatchesGoCohere_1527, TestPortMatchesGoCohere_1528, TestPortMatchesGoCohere_1526, TestPortMatchesGoCohere_1530, TestPortMatchesGoCohere_1524, TestPortMatchesGoCohere_1523, TestPortMatchesGoCohere_1522, TestPortMatchesGoCohere_1521, TestPortMatchesGoCohere_1519, TestPortMatchesGoCohere_1518, TestPortMatchesGoCohere_1525, TestPortMatchesGoCohere_1517, TestPortMatchesGoCohere_1516, TestPortMatchesGoCohere_1520, TestPortMatchesGoCohere_1515, TestPortMatchesGoCohere_1513, TestPortMatchesGoCohere_1511, TestPortMatchesGoCohere_1512, TestPortMatchesGoCohere_1514, TestPortMatchesGoCohere_1509, TestPortMatchesGoCohere_1507, TestPortMatchesGoCohere_1508, TestPortMatchesGoCohere_1506, TestPortMatchesGoCohere_1505, TestPortMatchesGoCohere_1504, TestPortMatchesGoCohere_1503, TestPortMatchesGoCohere_1510, TestPortMatchesGoCohere_1501, TestPortMatchesGoCohere_1500, TestPortMatchesGoCohere_1499, TestPortMatchesGoCohere_1498, TestPortMatchesGoCohere_1496, TestPortMatchesGoCohere_1502, TestPortMatchesGoCohere_1497, TestPortMatchesGoCohere_1493, TestPortMatchesGoCohere_1492, TestPortMatchesGoCohere_1491, TestPortMatchesGoCohere_1495, TestPortMatchesGoCohere_1489, TestPortMatchesGoCohere_1494, TestPortMatchesGoCohere_1487, TestPortMatchesGoCohere_1486, TestPortMatchesGoCohere_1485, TestPortMatchesGoCohere_1484, TestPortMatchesGoCohere_1488, TestPortMatchesGoCohere_1483, TestPortMatchesGoCohere_1482, TestPortMatchesGoCohere_1490, TestPortMatchesGoCohere_1480, TestPortMatchesGoCohere_1406, TestPortMatchesGoCohere_1479, TestPortMatchesGoCohere_1478, TestPortMatchesGoCohere_1476, TestPortMatchesGoCohere_1475, TestPortMatchesGoCohere_1477, TestPortMatchesGoCohere_1474, TestPortMatchesGoCohere_1481, TestPortMatchesGoCohere_1473, TestPortMatchesGoCohere_1471, TestPortMatchesGoCohere_1472, TestPortMatchesGoCohere_1468, TestPortMatchesGoCohere_1467, TestPortMatchesGoCohere_1466, TestPortMatchesGoCohere_1470, TestPortMatchesGoCohere_1469, TestPortMatchesGoCohere_1465, TestPortMatchesGoCohere_1464, TestPortMatchesGoCohere_1462, TestPortMatchesGoCohere_1461, TestPortMatchesGoCohere_1463, TestPortMatchesGoCohere_1459, TestPortMatchesGoCohere_1458, TestPortMatchesGoCohere_1457, TestPortMatchesGoCohere_1456, TestPortMatchesGoCohere_1455, TestPortMatchesGoCohere_1454, TestPortMatchesGoCohere_1460, TestPortMatchesGoCohere_1453, TestPortMatchesGoCohere_1451, TestPortMatchesGoCohere_1449, TestPortMatchesGoCohere_1452, TestPortMatchesGoCohere_1447, TestPortMatchesGoCohere_1448, TestPortMatchesGoCohere_1446, TestPortMatchesGoCohere_1450, TestPortMatchesGoCohere_1445, TestPortMatchesGoCohere_1443, TestPortMatchesGoCohere_1442, TestPortMatchesGoCohere_1441, TestPortMatchesGoCohere_1440, TestPortMatchesGoCohere_1439, TestPortMatchesGoCohere_1437, TestPortMatchesGoCohere_1436, TestPortMatchesGoCohere_1435, TestPortMatchesGoCohere_1434, TestPortMatchesGoCohere_1433, TestPortMatchesGoCohere_1432, TestPortMatchesGoCohere_1431, TestPortMatchesGoCohere_1430, TestPortMatchesGoCohere_1429, TestPortMatchesGoCohere_1438, TestPortMatchesGoCohere_1444, TestPortMatchesGoCohere_1426, TestPortMatchesGoCohere_1425, TestPortMatchesGoCohere_1424, TestPortMatchesGoCohere_1423, TestPortMatchesGoCohere_1422, TestPortMatchesGoCohere_1421, TestPortMatchesGoCohere_1420, TestPortMatchesGoCohere_1419, TestPortMatchesGoCohere_1405, TestPortMatchesGoCohere_1418, TestPortMatchesGoCohere_1417, TestPortMatchesGoCohere_1416, TestPortMatchesGoCohere_1415, TestPortMatchesGoCohere_1414, TestPortMatchesGoCohere_1413, TestPortMatchesGoCohere_1412, TestPortMatchesGoCohere_1411, TestPortMatchesGoCohere_1410, TestPortMatchesGoCohere_1280, TestPortMatchesGoCohere_1408, TestPortMatchesGoCohere_1407, TestPortMatchesGoCohere_1404, TestPortMatchesGoCohere_1403, TestPortMatchesGoCohere_1402, TestPortMatchesGoCohere_1401, TestPortMatchesGoCohere_1400, TestPortMatchesGoCohere_1399, TestPortMatchesGoCohere_1398, TestPortMatchesGoCohere_1397, TestPortMatchesGoCohere_1396, TestPortMatchesGoCohere_1395, TestPortMatchesGoCohere_1394, TestPortMatchesGoCohere_1393, TestPortMatchesGoCohere_1392, TestPortMatchesGoCohere_1391, TestPortMatchesGoCohere_1390, TestPortMatchesGoCohere_1389, TestPortMatchesGoCohere_1388, TestPortMatchesGoCohere_1387, TestPortMatchesGoCohere_1386, TestPortMatchesGoCohere_1385, TestPortMatchesGoCohere_1384, TestPortMatchesGoCohere_1383, TestPortMatchesGoCohere_1382, TestPortMatchesGoCohere_1381, TestPortMatchesGoCohere_1380, TestPortMatchesGoCohere_1379, TestPortMatchesGoCohere_1378, TestPortMatchesGoCohere_1377, TestPortMatchesGoCohere_1376, TestPortMatchesGoCohere_1375, TestPortMatchesGoCohere_1374, TestPortMatchesGoCohere_1373, TestPortMatchesGoCohere_1372, TestPortMatchesGoCohere_1371, TestPortMatchesGoCohere_1370, TestPortMatchesGoCohere_1369, TestPortMatchesGoCohere_1368, TestPortMatchesGoCohere_1367, TestPortMatchesGoCohere_1366, TestPortMatchesGoCohere_1365, TestPortMatchesGoCohere_1428, TestPortMatchesGoCohere_1363, TestPortMatchesGoCohere_1362, TestPortMatchesGoCohere_1361, TestPortMatchesGoCohere_1360, TestPortMatchesGoCohere_1359, TestPortMatchesGoCohere_1358, TestPortMatchesGoCohere_1357, TestPortMatchesGoCohere_1276, TestPortMatchesGoCohere_1356, TestPortMatchesGoCohere_1355, TestPortMatchesGoCohere_1354, TestPortMatchesGoCohere_1353, TestPortMatchesGoCohere_1352, TestPortMatchesGoCohere_1351, TestPortMatchesGoCohere_1350, TestPortMatchesGoCohere_1349, TestPortMatchesGoCohere_1348, TestPortMatchesGoCohere_1347, TestPortMatchesGoCohere_1346, TestPortMatchesGoCohere_1345, TestPortMatchesGoCohere_1344, TestPortMatchesGoCohere_1343, TestPortMatchesGoCohere_1342, TestPortMatchesGoCohere_1341, TestPortMatchesGoCohere_1340, TestPortMatchesGoCohere_1339, TestPortMatchesGoCohere_1338, TestPortMatchesGoCohere_1364, TestPortMatchesGoCohere_1427, TestPortMatchesGoCohere_1335, TestPortMatchesGoCohere_1336, TestPortMatchesGoCohere_1334, TestPortMatchesGoCohere_1333, TestPortMatchesGoCohere_1332, TestPortMatchesGoCohere_1330, TestPortMatchesGoCohere_1331, TestPortMatchesGoCohere_1327, TestPortMatchesGoCohere_1326, TestPortMatchesGoCohere_1325, TestPortMatchesGoCohere_1324, TestPortMatchesGoCohere_1323, TestPortMatchesGoCohere_1322, TestPortMatchesGoCohere_1321, TestPortMatchesGoCohere_1320, TestPortMatchesGoCohere_1319, TestPortMatchesGoCohere_1318, TestPortMatchesGoCohere_1317, TestPortMatchesGoCohere_1316, TestPortMatchesGoCohere_1315, TestPortMatchesGoCohere_1314, TestPortMatchesGoCohere_1313, TestPortMatchesGoCohere_1312, TestPortMatchesGoCohere_1311, TestPortMatchesGoCohere_1310, TestPortMatchesGoCohere_1309, TestPortMatchesGoCohere_1308, TestPortMatchesGoCohere_1307, TestPortMatchesGoCohere_1306, TestPortMatchesGoCohere_1305, TestPortMatchesGoCohere_1304, TestPortMatchesGoCohere_1303, TestPortMatchesGoCohere_1302, TestPortMatchesGoCohere_1301, TestPortMatchesGoCohere_1300, TestPortMatchesGoCohere_1299, TestPortMatchesGoCohere_1298, TestPortMatchesGoCohere_1297, TestPortMatchesGoCohere_1275, TestPortMatchesGoCohere_1296, TestPortMatchesGoCohere_1295, TestPortMatchesGoCohere_1294, TestPortMatchesGoCohere_1293, TestPortMatchesGoCohere_1292, TestPortMatchesGoCohere_1291, TestPortMatchesGoCohere_1290, TestPortMatchesGoCohere_1289, TestPortMatchesGoCohere_1288, TestPortMatchesGoCohere_1287, TestPortMatchesGoCohere_1286, TestPortMatchesGoCohere_1285, TestPortMatchesGoCohere_1284, TestPortMatchesGoCohere_1283, TestPortMatchesGoCohere_1282, TestPortMatchesGoCohere_1281, TestPortMatchesGoCohere_1151, TestPortMatchesGoCohere_1279, TestPortMatchesGoCohere_1278, TestPortMatchesGoCohere_1277, TestPortMatchesGoCohere_1274, TestPortMatchesGoCohere_1273, TestPortMatchesGoCohere_1272, TestPortMatchesGoCohere_1271, TestPortMatchesGoCohere_1270, TestPortMatchesGoCohere_1269, TestPortMatchesGoCohere_1268, TestPortMatchesGoCohere_1267, TestPortMatchesGoCohere_1266, TestPortMatchesGoCohere_1265, TestPortMatchesGoCohere_1264, TestPortMatchesGoCohere_1263, TestPortMatchesGoCohere_1262, TestPortMatchesGoCohere_1261, TestPortMatchesGoCohere_1260, TestPortMatchesGoCohere_1259, TestPortMatchesGoCohere_1258, TestPortMatchesGoCohere_1257, TestPortMatchesGoCohere_1256, TestPortMatchesGoCohere_1255, TestPortMatchesGoCohere_1254, TestPortMatchesGoCohere_1253, TestPortMatchesGoCohere_1252, TestPortMatchesGoCohere_1251, TestPortMatchesGoCohere_1250, TestPortMatchesGoCohere_1249, TestPortMatchesGoCohere_1248, TestPortMatchesGoCohere_1247, TestPortMatchesGoCohere_1246, TestPortMatchesGoCohere_1245, TestPortMatchesGoCohere_1244, TestPortMatchesGoCohere_1243, TestPortMatchesGoCohere_1242, TestPortMatchesGoCohere_1241, TestPortMatchesGoCohere_1240, TestPortMatchesGoCohere_1239, TestPortMatchesGoCohere_1238, TestPortMatchesGoCohere_1237, TestPortMatchesGoCohere_1236, TestPortMatchesGoCohere_1235, TestPortMatchesGoCohere_1146, TestPortMatchesGoCohere_1234, TestPortMatchesGoCohere_1233, TestPortMatchesGoCohere_1232, TestPortMatchesGoCohere_1231, TestPortMatchesGoCohere_1230, TestPortMatchesGoCohere_1229, TestPortMatchesGoCohere_1228, TestPortMatchesGoCohere_1227, TestPortMatchesGoCohere_1226, TestPortMatchesGoCohere_1225, TestPortMatchesGoCohere_1224, TestPortMatchesGoCohere_1223, TestPortMatchesGoCohere_1222, TestPortMatchesGoCohere_1221, TestPortMatchesGoCohere_1220, TestPortMatchesGoCohere_1219, TestPortMatchesGoCohere_1218, TestPortMatchesGoCohere_1217, TestPortMatchesGoCohere_1216, TestPortMatchesGoCohere_1215, TestPortMatchesGoCohere_1214, TestPortMatchesGoCohere_1213, TestPortMatchesGoCohere_1212, TestPortMatchesGoCohere_1211, TestPortMatchesGoCohere_1210, TestPortMatchesGoCohere_1209, TestPortMatchesGoCohere_1208, TestPortMatchesGoCohere_1207, TestPortMatchesGoCohere_1206, TestPortMatchesGoCohere_1205, TestPortMatchesGoCohere_1204, TestPortMatchesGoCohere_1203, TestPortMatchesGoCohere_1202, TestPortMatchesGoCohere_1201, TestPortMatchesGoCohere_1200, TestPortMatchesGoCohere_1199, TestPortMatchesGoCohere_1198, TestPortMatchesGoCohere_1197, TestPortMatchesGoCohere_1196, TestPortMatchesGoCohere_1195, TestPortMatchesGoCohere_1194, TestPortMatchesGoCohere_1193, TestPortMatchesGoCohere_1192, TestPortMatchesGoCohere_1191, TestPortMatchesGoCohere_1190, TestPortMatchesGoCohere_1189, TestPortMatchesGoCohere_1188, TestPortMatchesGoCohere_1187, TestPortMatchesGoCohere_1186, TestPortMatchesGoCohere_1185, TestPortMatchesGoCohere_1184, TestPortMatchesGoCohere_1183, TestPortMatchesGoCohere_1182, TestPortMatchesGoCohere_1181, TestPortMatchesGoCohere_1180, TestPortMatchesGoCohere_1179, TestPortMatchesGoCohere_1178, TestPortMatchesGoCohere_1177, TestPortMatchesGoCohere_1176, TestPortMatchesGoCohere_1175, TestPortMatchesGoCohere_1145, TestPortMatchesGoCohere_1174, TestPortMatchesGoCohere_1173, TestPortMatchesGoCohere_1172, TestPortMatchesGoCohere_1171, TestPortMatchesGoCohere_1170, TestPortMatchesGoCohere_1169, TestPortMatchesGoCohere_1168, TestPortMatchesGoCohere_1167, TestPortMatchesGoCohere_1166, TestPortMatchesGoCohere_1165, TestPortMatchesGoCohere_1164, TestPortMatchesGoCohere_1163, TestPortMatchesGoCohere_1162, TestPortMatchesGoCohere_1161, TestPortMatchesGoCohere_1160, TestPortMatchesGoCohere_1159, TestPortMatchesGoCohere_1158, TestPortMatchesGoCohere_1157, TestPortMatchesGoCohere_1156, TestPortMatchesGoCohere_1155, TestPortMatchesGoCohere_1154, TestPortMatchesGoCohere_1153, TestPortMatchesGoCohere_1152, TestPortMatchesGoCohere_1022, TestPortMatchesGoCohere_1150, TestPortMatchesGoCohere_1149, TestPortMatchesGoCohere_1148, TestPortMatchesGoCohere_1147, TestPortMatchesGoCohere_1144, TestPortMatchesGoCohere_1143, TestPortMatchesGoCohere_1142, TestPortMatchesGoCohere_1141, TestPortMatchesGoCohere_1140, TestPortMatchesGoCohere_1139, TestPortMatchesGoCohere_1138, TestPortMatchesGoCohere_1137, TestPortMatchesGoCohere_1136, TestPortMatchesGoCohere_1135, TestPortMatchesGoCohere_1134, TestPortMatchesGoCohere_1133, TestPortMatchesGoCohere_1132, TestPortMatchesGoCohere_1048, TestPortMatchesGoCohere_1131, TestPortMatchesGoCohere_1130, TestPortMatchesGoCohere_1129, TestPortMatchesGoCohere_1128, TestPortMatchesGoCohere_1127, TestPortMatchesGoCohere_1126, TestPortMatchesGoCohere_1125, TestPortMatchesGoCohere_1124, TestPortMatchesGoCohere_1123, TestPortMatchesGoCohere_1122, TestPortMatchesGoCohere_1121, TestPortMatchesGoCohere_1120, TestPortMatchesGoCohere_1119, TestPortMatchesGoCohere_1118, TestPortMatchesGoCohere_1117, TestPortMatchesGoCohere_1116, TestPortMatchesGoCohere_1115, TestPortMatchesGoCohere_1114, TestPortMatchesGoCohere_1016, TestPortMatchesGoCohere_1113, TestPortMatchesGoCohere_1112, TestPortMatchesGoCohere_1329, TestPortMatchesGoCohere_1337, TestPortMatchesGoCohere_1109, TestPortMatchesGoCohere_1108, TestPortMatchesGoCohere_1110, TestPortMatchesGoCohere_1106, TestPortMatchesGoCohere_1105, TestPortMatchesGoCohere_1104, TestPortMatchesGoCohere_1103, TestPortMatchesGoCohere_1102, TestPortMatchesGoCohere_1101, TestPortMatchesGoCohere_1100, TestPortMatchesGoCohere_1099, TestPortMatchesGoCohere_1098, TestPortMatchesGoCohere_1097, TestPortMatchesGoCohere_1096, TestPortMatchesGoCohere_1095, TestPortMatchesGoCohere_1094, TestPortMatchesGoCohere_1093, TestPortMatchesGoCohere_1092, TestPortMatchesGoCohere_1091, TestPortMatchesGoCohere_1090, TestPortMatchesGoCohere_1089, TestPortMatchesGoCohere_1088, TestPortMatchesGoCohere_1087, TestPortMatchesGoCohere_1086, TestPortMatchesGoCohere_1085, TestPortMatchesGoCohere_1084, TestPortMatchesGoCohere_1083, TestPortMatchesGoCohere_1082, TestPortMatchesGoCohere_1111, TestPortMatchesGoCohere_1107, TestPortMatchesGoCohere_1081, TestPortMatchesGoCohere_1328, TestPortMatchesGoCohere_1080, TestPortMatchesGoCohere_1076, TestPortMatchesGoCohere_1075, TestPortMatchesGoCohere_1074, TestPortMatchesGoCohere_1073, TestPortMatchesGoCohere_1078, TestPortMatchesGoCohere_1077, TestPortMatchesGoCohere_1071, TestPortMatchesGoCohere_1070, TestPortMatchesGoCohere_1079, TestPortMatchesGoCohere_1069, TestPortMatchesGoCohere_1068, TestPortMatchesGoCohere_1067, TestPortMatchesGoCohere_1066, TestPortMatchesGoCohere_1063, TestPortMatchesGoCohere_1064, TestPortMatchesGoCohere_1062, TestPortMatchesGoCohere_1061, TestPortMatchesGoCohere_1072, TestPortMatchesGoCohere_1065, TestPortMatchesGoCohere_1057, TestPortMatchesGoCohere_1060, TestPortMatchesGoCohere_1056, TestPortMatchesGoCohere_1054, TestPortMatchesGoCohere_1015, TestPortMatchesGoCohere_1059, TestPortMatchesGoCohere_1055, TestPortMatchesGoCohere_1052, TestPortMatchesGoCohere_1058, TestPortMatchesGoCohere_1051, TestPortMatchesGoCohere_1049, TestPortMatchesGoCohere_1047, TestPortMatchesGoCohere_1050, TestPortMatchesGoCohere_1046, TestPortMatchesGoCohere_1045, TestPortMatchesGoCohere_1043, TestPortMatchesGoCohere_1053, TestPortMatchesGoCohere_1042, TestPortMatchesGoCohere_1044, TestPortMatchesGoCohere_1038, TestPortMatchesGoCohere_1037, TestPortMatchesGoCohere_1036, TestPortMatchesGoCohere_1040, TestPortMatchesGoCohere_1039, TestPortMatchesGoCohere_1041, TestPortMatchesGoCohere_1034, TestPortMatchesGoCohere_1035, TestPortMatchesGoCohere_1030, TestPortMatchesGoCohere_1029, TestPortMatchesGoCohere_1032, TestPortMatchesGoCohere_1033, TestPortMatchesGoCohere_1027, TestPortMatchesGoCohere_1031, TestPortMatchesGoCohere_1026, TestPortMatchesGoCohere_1025, TestPortMatchesGoCohere_1024, TestPortMatchesGoCohere_1023, TestPortMatchesGoCohere_0893, TestPortMatchesGoCohere_1021, TestPortMatchesGoCohere_1020, TestPortMatchesGoCohere_1019, TestPortMatchesGoCohere_1018, TestPortMatchesGoCohere_1017, TestPortMatchesGoCohere_1014, TestPortMatchesGoCohere_1013, TestPortMatchesGoCohere_1012, TestPortMatchesGoCohere_1011, TestPortMatchesGoCohere_1010, TestPortMatchesGoCohere_1009, TestPortMatchesGoCohere_1008, TestPortMatchesGoCohere_1007, TestPortMatchesGoCohere_1006, TestPortMatchesGoCohere_1005, TestPortMatchesGoCohere_1004, TestPortMatchesGoCohere_1003, TestPortMatchesGoCohere_1002, TestPortMatchesGoCohere_1001, TestPortMatchesGoCohere_1000, TestPortMatchesGoCohere_0999, TestPortMatchesGoCohere_0998, TestPortMatchesGoCohere_0997, TestPortMatchesGoCohere_0996, TestPortMatchesGoCohere_0995, TestPortMatchesGoCohere_0994, TestPortMatchesGoCohere_0993, TestPortMatchesGoCohere_0992, TestPortMatchesGoCohere_0991, TestPortMatchesGoCohere_0886, TestPortMatchesGoCohere_0990, TestPortMatchesGoCohere_0989, TestPortMatchesGoCohere_0988, TestPortMatchesGoCohere_0987, TestPortMatchesGoCohere_0986, TestPortMatchesGoCohere_0985, TestPortMatchesGoCohere_0984, TestPortMatchesGoCohere_0983, TestPortMatchesGoCohere_0982, TestPortMatchesGoCohere_0981, TestPortMatchesGoCohere_0980, TestPortMatchesGoCohere_0979, TestPortMatchesGoCohere_0978, TestPortMatchesGoCohere_0977, TestPortMatchesGoCohere_0976, TestPortMatchesGoCohere_0974, TestPortMatchesGoCohere_0973, TestPortMatchesGoCohere_0975, TestPortMatchesGoCohere_0972, TestPortMatchesGoCohere_0969, TestPortMatchesGoCohere_0971, TestPortMatchesGoCohere_0967, TestPortMatchesGoCohere_1028, TestPortMatchesGoCohere_0966, TestPortMatchesGoCohere_0965, TestPortMatchesGoCohere_0964, TestPortMatchesGoCohere_0963, TestPortMatchesGoCohere_0961, TestPortMatchesGoCohere_0960, TestPortMatchesGoCohere_0962, TestPortMatchesGoCohere_0959, TestPortMatchesGoCohere_0958, TestPortMatchesGoCohere_0970, TestPortMatchesGoCohere_0957, TestPortMatchesGoCohere_0956, TestPortMatchesGoCohere_0955, TestPortMatchesGoCohere_0954, TestPortMatchesGoCohere_0953, TestPortMatchesGoCohere_0952, TestPortMatchesGoCohere_0951, TestPortMatchesGoCohere_0950, TestPortMatchesGoCohere_0948, TestPortMatchesGoCohere_0947, TestPortMatchesGoCohere_0946, TestPortMatchesGoCohere_0945, TestPortMatchesGoCohere_0943, TestPortMatchesGoCohere_0944, TestPortMatchesGoCohere_0949, TestPortMatchesGoCohere_0942, TestPortMatchesGoCohere_0939, TestPortMatchesGoCohere_0938, TestPortMatchesGoCohere_0940, TestPortMatchesGoCohere_0941, TestPortMatchesGoCohere_0937, TestPortMatchesGoCohere_0968, TestPortMatchesGoCohere_0935, TestPortMatchesGoCohere_0936, TestPortMatchesGoCohere_0934, TestPortMatchesGoCohere_0932, TestPortMatchesGoCohere_0931, TestPortMatchesGoCohere_0930, TestPortMatchesGoCohere_0885, TestPortMatchesGoCohere_0929, TestPortMatchesGoCohere_0926, TestPortMatchesGoCohere_0928, TestPortMatchesGoCohere_0927, TestPortMatchesGoCohere_0925, TestPortMatchesGoCohere_0924, TestPortMatchesGoCohere_0923, TestPortMatchesGoCohere_0922, TestPortMatchesGoCohere_0920, TestPortMatchesGoCohere_0921, TestPortMatchesGoCohere_0917, TestPortMatchesGoCohere_0916, TestPortMatchesGoCohere_0933, TestPortMatchesGoCohere_0915, TestPortMatchesGoCohere_0919, TestPortMatchesGoCohere_0914, TestPortMatchesGoCohere_0911, TestPortMatchesGoCohere_0912, TestPortMatchesGoCohere_0910, TestPortMatchesGoCohere_0909, TestPortMatchesGoCohere_0913, TestPortMatchesGoCohere_0906, TestPortMatchesGoCohere_0907, TestPortMatchesGoCohere_0918, TestPortMatchesGoCohere_0905, TestPortMatchesGoCohere_0904, TestPortMatchesGoCohere_0902, TestPortMatchesGoCohere_0901, TestPortMatchesGoCohere_0899, TestPortMatchesGoCohere_0898, TestPortMatchesGoCohere_0900, TestPortMatchesGoCohere_0908, TestPortMatchesGoCohere_0896, TestPortMatchesGoCohere_0897, TestPortMatchesGoCohere_0764, TestPortMatchesGoCohere_0892, TestPortMatchesGoCohere_0891, TestPortMatchesGoCohere_0890, TestPortMatchesGoCohere_0903, TestPortMatchesGoCohere_0889, TestPortMatchesGoCohere_0888, TestPortMatchesGoCohere_0894, TestPortMatchesGoCohere_0884, TestPortMatchesGoCohere_0883, TestPortMatchesGoCohere_0882, TestPortMatchesGoCohere_0881, TestPortMatchesGoCohere_0880, TestPortMatchesGoCohere_0887, TestPortMatchesGoCohere_0877, TestPortMatchesGoCohere_0876, TestPortMatchesGoCohere_0875, TestPortMatchesGoCohere_0874, TestPortMatchesGoCohere_0873, TestPortMatchesGoCohere_0872, TestPortMatchesGoCohere_0871, TestPortMatchesGoCohere_0870, TestPortMatchesGoCohere_0869, TestPortMatchesGoCohere_0756, TestPortMatchesGoCohere_0868, TestPortMatchesGoCohere_0867, TestPortMatchesGoCohere_0866, TestPortMatchesGoCohere_0865, TestPortMatchesGoCohere_0864, TestPortMatchesGoCohere_0863, TestPortMatchesGoCohere_0862, TestPortMatchesGoCohere_0861, TestPortMatchesGoCohere_0860, TestPortMatchesGoCohere_0859, TestPortMatchesGoCohere_0858, TestPortMatchesGoCohere_0857, TestPortMatchesGoCohere_0856, TestPortMatchesGoCohere_0855, TestPortMatchesGoCohere_0854, TestPortMatchesGoCohere_0853, TestPortMatchesGoCohere_0852, TestPortMatchesGoCohere_0851, TestPortMatchesGoCohere_0850, TestPortMatchesGoCohere_0849, TestPortMatchesGoCohere_0848, TestPortMatchesGoCohere_0847, TestPortMatchesGoCohere_0879, TestPortMatchesGoCohere_0845, TestPortMatchesGoCohere_0844, TestPortMatchesGoCohere_0895, TestPortMatchesGoCohere_0843, TestPortMatchesGoCohere_0846, TestPortMatchesGoCohere_0840, TestPortMatchesGoCohere_0842, TestPortMatchesGoCohere_0838, TestPortMatchesGoCohere_0837, TestPortMatchesGoCohere_0836, TestPortMatchesGoCohere_0835, TestPortMatchesGoCohere_0834, TestPortMatchesGoCohere_0833, TestPortMatchesGoCohere_0832, TestPortMatchesGoCohere_0831, TestPortMatchesGoCohere_0830, TestPortMatchesGoCohere_0829, TestPortMatchesGoCohere_0828, TestPortMatchesGoCohere_0827, TestPortMatchesGoCohere_0826, TestPortMatchesGoCohere_0825, TestPortMatchesGoCohere_0824, TestPortMatchesGoCohere_0823, TestPortMatchesGoCohere_0822, TestPortMatchesGoCohere_0821, TestPortMatchesGoCohere_0820, TestPortMatchesGoCohere_0819, TestPortMatchesGoCohere_0818, TestPortMatchesGoCohere_0817, TestPortMatchesGoCohere_0816, TestPortMatchesGoCohere_0815, TestPortMatchesGoCohere_0814, TestPortMatchesGoCohere_0813, TestPortMatchesGoCohere_0812, TestPortMatchesGoCohere_0811, TestPortMatchesGoCohere_0810, TestPortMatchesGoCohere_0809, TestPortMatchesGoCohere_0755, TestPortMatchesGoCohere_0808, TestPortMatchesGoCohere_0807, TestPortMatchesGoCohere_0806, TestPortMatchesGoCohere_0805, TestPortMatchesGoCohere_0804, TestPortMatchesGoCohere_0803, TestPortMatchesGoCohere_0802, TestPortMatchesGoCohere_0801, TestPortMatchesGoCohere_0800, TestPortMatchesGoCohere_0799, TestPortMatchesGoCohere_0798, TestPortMatchesGoCohere_0797, TestPortMatchesGoCohere_0796, TestPortMatchesGoCohere_0795, TestPortMatchesGoCohere_0794, TestPortMatchesGoCohere_0793, TestPortMatchesGoCohere_0792, TestPortMatchesGoCohere_0791, TestPortMatchesGoCohere_0790, TestPortMatchesGoCohere_0789, TestPortMatchesGoCohere_0788, TestPortMatchesGoCohere_0787, TestPortMatchesGoCohere_0786, TestPortMatchesGoCohere_0785, TestPortMatchesGoCohere_0784, TestPortMatchesGoCohere_0783, TestPortMatchesGoCohere_0782, TestPortMatchesGoCohere_0781, TestPortMatchesGoCohere_0780, TestPortMatchesGoCohere_0779, TestPortMatchesGoCohere_0778, TestPortMatchesGoCohere_0878, TestPortMatchesGoCohere_0841, TestPortMatchesGoCohere_0775, TestPortMatchesGoCohere_0774, TestPortMatchesGoCohere_0773, TestPortMatchesGoCohere_0772, TestPortMatchesGoCohere_0771, TestPortMatchesGoCohere_0770, TestPortMatchesGoCohere_0777, TestPortMatchesGoCohere_0768, TestPortMatchesGoCohere_0767, TestPortMatchesGoCohere_0766, TestPortMatchesGoCohere_0765, TestPortMatchesGoCohere_0635, TestPortMatchesGoCohere_0763, TestPortMatchesGoCohere_0762, TestPortMatchesGoCohere_0761, TestPortMatchesGoCohere_0839, TestPortMatchesGoCohere_0759, TestPortMatchesGoCohere_0758, TestPortMatchesGoCohere_0769, TestPortMatchesGoCohere_0776, TestPortMatchesGoCohere_0754, TestPortMatchesGoCohere_0753, TestPortMatchesGoCohere_0752, TestPortMatchesGoCohere_0751, TestPortMatchesGoCohere_0750, TestPortMatchesGoCohere_0749, TestPortMatchesGoCohere_0748, TestPortMatchesGoCohere_0747, TestPortMatchesGoCohere_0746, TestPortMatchesGoCohere_0626, TestPortMatchesGoCohere_0745, TestPortMatchesGoCohere_0744, TestPortMatchesGoCohere_0743, TestPortMatchesGoCohere_0742, TestPortMatchesGoCohere_0741, TestPortMatchesGoCohere_0760, TestPortMatchesGoCohere_0740, TestPortMatchesGoCohere_0739, TestPortMatchesGoCohere_0738, TestPortMatchesGoCohere_0736, TestPortMatchesGoCohere_0735, TestPortMatchesGoCohere_0734, TestPortMatchesGoCohere_0733, TestPortMatchesGoCohere_0732, TestPortMatchesGoCohere_0731, TestPortMatchesGoCohere_0730, TestPortMatchesGoCohere_0729, TestPortMatchesGoCohere_0728, TestPortMatchesGoCohere_0727, TestPortMatchesGoCohere_0726, TestPortMatchesGoCohere_0737, TestPortMatchesGoCohere_0724, TestPortMatchesGoCohere_0723, TestPortMatchesGoCohere_0722, TestPortMatchesGoCohere_0721, TestPortMatchesGoCohere_0720, TestPortMatchesGoCohere_0719, TestPortMatchesGoCohere_0757, TestPortMatchesGoCohere_0717, TestPortMatchesGoCohere_0716, TestPortMatchesGoCohere_0715, TestPortMatchesGoCohere_0714, TestPortMatchesGoCohere_0725, TestPortMatchesGoCohere_0711, TestPortMatchesGoCohere_0710, TestPortMatchesGoCohere_0709, TestPortMatchesGoCohere_0708, TestPortMatchesGoCohere_0707, TestPortMatchesGoCohere_0706, TestPortMatchesGoCohere_0705, TestPortMatchesGoCohere_0713, TestPortMatchesGoCohere_0704, TestPortMatchesGoCohere_0703, TestPortMatchesGoCohere_0702, TestPortMatchesGoCohere_0718, TestPortMatchesGoCohere_0701, TestPortMatchesGoCohere_0712, TestPortMatchesGoCohere_0698, TestPortMatchesGoCohere_0697, TestPortMatchesGoCohere_0700, TestPortMatchesGoCohere_0695, TestPortMatchesGoCohere_0694, TestPortMatchesGoCohere_0693, TestPortMatchesGoCohere_0692, TestPortMatchesGoCohere_0691, TestPortMatchesGoCohere_0690, TestPortMatchesGoCohere_0689, TestPortMatchesGoCohere_0687, TestPortMatchesGoCohere_0696, TestPortMatchesGoCohere_0688, TestPortMatchesGoCohere_0686, TestPortMatchesGoCohere_0685, TestPortMatchesGoCohere_0684, TestPortMatchesGoCohere_0683, TestPortMatchesGoCohere_0682, TestPortMatchesGoCohere_0699, TestPortMatchesGoCohere_0681, TestPortMatchesGoCohere_0679, TestPortMatchesGoCohere_0625, TestPortMatchesGoCohere_0677, TestPortMatchesGoCohere_0680, TestPortMatchesGoCohere_0675, TestPortMatchesGoCohere_0674, TestPortMatchesGoCohere_0678, TestPortMatchesGoCohere_0673, TestPortMatchesGoCohere_0670, TestPortMatchesGoCohere_0669, TestPortMatchesGoCohere_0668, TestPortMatchesGoCohere_0667, TestPortMatchesGoCohere_0666, TestPortMatchesGoCohere_0665, TestPortMatchesGoCohere_0664, TestPortMatchesGoCohere_0663, TestPortMatchesGoCohere_0662, TestPortMatchesGoCohere_0661, TestPortMatchesGoCohere_0660, TestPortMatchesGoCohere_0659, TestPortMatchesGoCohere_0658, TestPortMatchesGoCohere_0657, TestPortMatchesGoCohere_0656, TestPortMatchesGoCohere_0655, TestPortMatchesGoCohere_0654, TestPortMatchesGoCohere_0653, TestPortMatchesGoCohere_0652, TestPortMatchesGoCohere_0651, TestPortMatchesGoCohere_0650, TestPortMatchesGoCohere_0649, TestPortMatchesGoCohere_0648, TestPortMatchesGoCohere_0647, TestPortMatchesGoCohere_0646, TestPortMatchesGoCohere_0645, TestPortMatchesGoCohere_0644, TestPortMatchesGoCohere_0643, TestPortMatchesGoCohere_0642, TestPortMatchesGoCohere_0641, TestPortMatchesGoCohere_0640, TestPortMatchesGoCohere_0639, TestPortMatchesGoCohere_0638, TestPortMatchesGoCohere_0637, TestPortMatchesGoCohere_0636, TestPortMatchesGoCohere_0506, TestPortMatchesGoCohere_0634, TestPortMatchesGoCohere_0633, TestPortMatchesGoCohere_0632, TestPortMatchesGoCohere_0631, TestPortMatchesGoCohere_0630, TestPortMatchesGoCohere_0629, TestPortMatchesGoCohere_0628, TestPortMatchesGoCohere_0627, TestPortMatchesGoCohere_0624, TestPortMatchesGoCohere_0623, TestPortMatchesGoCohere_0622, TestPortMatchesGoCohere_0621, TestPortMatchesGoCohere_0620, TestPortMatchesGoCohere_0619, TestPortMatchesGoCohere_0618, TestPortMatchesGoCohere_0617, TestPortMatchesGoCohere_0616, TestPortMatchesGoCohere_0615, TestPortMatchesGoCohere_0614, TestPortMatchesGoCohere_0613, TestPortMatchesGoCohere_0612, TestPortMatchesGoCohere_0611, TestPortMatchesGoCohere_0610, TestPortMatchesGoCohere_0609, TestPortMatchesGoCohere_0608, TestPortMatchesGoCohere_0607, TestPortMatchesGoCohere_0606, TestPortMatchesGoCohere_0605, TestPortMatchesGoCohere_0604, TestPortMatchesGoCohere_0603, TestPortMatchesGoCohere_0602, TestPortMatchesGoCohere_0601, TestPortMatchesGoCohere_0600, TestPortMatchesGoCohere_0599, TestPortMatchesGoCohere_0598, TestPortMatchesGoCohere_0597, TestPortMatchesGoCohere_0596, TestPortMatchesGoCohere_0595, TestPortMatchesGoCohere_0594, TestPortMatchesGoCohere_0593, TestPortMatchesGoCohere_0592, TestPortMatchesGoCohere_0591, TestPortMatchesGoCohere_0590, TestPortMatchesGoCohere_0589, TestPortMatchesGoCohere_0588, TestPortMatchesGoCohere_0587, TestPortMatchesGoCohere_0586, TestPortMatchesGoCohere_0585, TestPortMatchesGoCohere_0584, TestPortMatchesGoCohere_0583, TestPortMatchesGoCohere_0582, TestPortMatchesGoCohere_0581, TestPortMatchesGoCohere_0580, TestPortMatchesGoCohere_0579, TestPortMatchesGoCohere_0578, TestPortMatchesGoCohere_0577, TestPortMatchesGoCohere_0576, TestPortMatchesGoCohere_0575, TestPortMatchesGoCohere_0574, TestPortMatchesGoCohere_0573, TestPortMatchesGoCohere_0572, TestPortMatchesGoCohere_0571, TestPortMatchesGoCohere_0570, TestPortMatchesGoCohere_0569, TestPortMatchesGoCohere_0568, TestPortMatchesGoCohere_0567, TestPortMatchesGoCohere_0566, TestPortMatchesGoCohere_0565, TestPortMatchesGoCohere_0564, TestPortMatchesGoCohere_0496, TestPortMatchesGoCohere_0563, TestPortMatchesGoCohere_0562, TestPortMatchesGoCohere_0561, TestPortMatchesGoCohere_0560, TestPortMatchesGoCohere_0559, TestPortMatchesGoCohere_0558, TestPortMatchesGoCohere_0557, TestPortMatchesGoCohere_0556, TestPortMatchesGoCohere_0555, TestPortMatchesGoCohere_0554, TestPortMatchesGoCohere_0553, TestPortMatchesGoCohere_0552, TestPortMatchesGoCohere_0551, TestPortMatchesGoCohere_0671, TestPortMatchesGoCohere_0672, TestPortMatchesGoCohere_0550, TestPortMatchesGoCohere_0549, TestPortMatchesGoCohere_0676, TestPortMatchesGoCohere_0547, TestPortMatchesGoCohere_0544, TestPortMatchesGoCohere_0543, TestPortMatchesGoCohere_0542, TestPortMatchesGoCohere_0548, TestPortMatchesGoCohere_0540, TestPortMatchesGoCohere_0545, TestPortMatchesGoCohere_0539, TestPortMatchesGoCohere_0546, TestPortMatchesGoCohere_0538, TestPortMatchesGoCohere_0536, TestPortMatchesGoCohere_0541, TestPortMatchesGoCohere_0534, TestPortMatchesGoCohere_0533, TestPortMatchesGoCohere_0532, TestPortMatchesGoCohere_0531, TestPortMatchesGoCohere_0530, TestPortMatchesGoCohere_0529, TestPortMatchesGoCohere_0528, TestPortMatchesGoCohere_0527, TestPortMatchesGoCohere_0526, TestPortMatchesGoCohere_0525, TestPortMatchesGoCohere_0535, TestPortMatchesGoCohere_0523, TestPortMatchesGoCohere_0522, TestPortMatchesGoCohere_0521, TestPortMatchesGoCohere_0520, TestPortMatchesGoCohere_0524, TestPortMatchesGoCohere_0537, TestPortMatchesGoCohere_0517, TestPortMatchesGoCohere_0515, TestPortMatchesGoCohere_0514, TestPortMatchesGoCohere_0513, TestPortMatchesGoCohere_0512, TestPortMatchesGoCohere_0511, TestPortMatchesGoCohere_0519, TestPortMatchesGoCohere_0510, TestPortMatchesGoCohere_0508, TestPortMatchesGoCohere_0509, TestPortMatchesGoCohere_0507, TestPortMatchesGoCohere_0518, TestPortMatchesGoCohere_0505, TestPortMatchesGoCohere_0504, TestPortMatchesGoCohere_0495, TestPortMatchesGoCohere_0503, TestPortMatchesGoCohere_0501, TestPortMatchesGoCohere_0500, TestPortMatchesGoCohere_0499, TestPortMatchesGoCohere_0498, TestPortMatchesGoCohere_0497, TestPortMatchesGoCohere_0494, TestPortMatchesGoCohere_0516, TestPortMatchesGoCohere_0502, TestPortMatchesGoCohere_0493, TestPortMatchesGoCohere_0492, TestPortMatchesGoCohere_0491, TestPortMatchesGoCohere_0488, TestPortMatchesGoCohere_0489, TestPortMatchesGoCohere_0487, TestPortMatchesGoCohere_0490, TestPortMatchesGoCohere_0485, TestPortMatchesGoCohere_0484, TestPortMatchesGoCohere_0483, TestPortMatchesGoCohere_0482, TestPortMatchesGoCohere_0481, TestPortMatchesGoCohere_0480, TestPortMatchesGoCohere_0479, TestPortMatchesGoCohere_0478, TestPortMatchesGoCohere_0486, TestPortMatchesGoCohere_0476, TestPortMatchesGoCohere_0377, TestPortMatchesGoCohere_0474, TestPortMatchesGoCohere_0473, TestPortMatchesGoCohere_0475, TestPortMatchesGoCohere_0471, TestPortMatchesGoCohere_0470, TestPortMatchesGoCohere_0469, TestPortMatchesGoCohere_0468, TestPortMatchesGoCohere_0467, TestPortMatchesGoCohere_0466, TestPortMatchesGoCohere_0477, TestPortMatchesGoCohere_0464, TestPortMatchesGoCohere_0463, TestPortMatchesGoCohere_0462, TestPortMatchesGoCohere_0461, TestPortMatchesGoCohere_0460, TestPortMatchesGoCohere_0459, TestPortMatchesGoCohere_0458, TestPortMatchesGoCohere_0457, TestPortMatchesGoCohere_0456, TestPortMatchesGoCohere_0455, TestPortMatchesGoCohere_0454, TestPortMatchesGoCohere_0453, TestPortMatchesGoCohere_0452, TestPortMatchesGoCohere_0451, TestPortMatchesGoCohere_0450, TestPortMatchesGoCohere_0449, TestPortMatchesGoCohere_0448, TestPortMatchesGoCohere_0447, TestPortMatchesGoCohere_0446, TestPortMatchesGoCohere_0445, TestPortMatchesGoCohere_0444, TestPortMatchesGoCohere_0443, TestPortMatchesGoCohere_0442, TestPortMatchesGoCohere_0399, TestPortMatchesGoCohere_0441, TestPortMatchesGoCohere_0440, TestPortMatchesGoCohere_0439, TestPortMatchesGoCohere_0438, TestPortMatchesGoCohere_0436, TestPortMatchesGoCohere_0465, TestPortMatchesGoCohere_0435, TestPortMatchesGoCohere_0434, TestPortMatchesGoCohere_0433, TestPortMatchesGoCohere_0432, TestPortMatchesGoCohere_0431, TestPortMatchesGoCohere_0430, TestPortMatchesGoCohere_0429, TestPortMatchesGoCohere_0428, TestPortMatchesGoCohere_0427, TestPortMatchesGoCohere_0426, TestPortMatchesGoCohere_0425, TestPortMatchesGoCohere_0424, TestPortMatchesGoCohere_0423, TestPortMatchesGoCohere_0422, TestPortMatchesGoCohere_0421, TestPortMatchesGoCohere_0420, TestPortMatchesGoCohere_0419, TestPortMatchesGoCohere_0437, TestPortMatchesGoCohere_0417, TestPortMatchesGoCohere_0472, TestPortMatchesGoCohere_0415, TestPortMatchesGoCohere_0414, TestPortMatchesGoCohere_0413, TestPortMatchesGoCohere_0412, TestPortMatchesGoCohere_0418, TestPortMatchesGoCohere_0410, TestPortMatchesGoCohere_0411, TestPortMatchesGoCohere_0408, TestPortMatchesGoCohere_0407, TestPortMatchesGoCohere_0406, TestPortMatchesGoCohere_0405, TestPortMatchesGoCohere_0404, TestPortMatchesGoCohere_0402, TestPortMatchesGoCohere_0401, TestPortMatchesGoCohere_0400, TestPortMatchesGoCohere_0398, TestPortMatchesGoCohere_0397, TestPortMatchesGoCohere_0396, TestPortMatchesGoCohere_0395, TestPortMatchesGoCohere_0394, TestPortMatchesGoCohere_0416, TestPortMatchesGoCohere_0393, TestPortMatchesGoCohere_0391, TestPortMatchesGoCohere_0390, TestPortMatchesGoCohere_0389, TestPortMatchesGoCohere_0388, TestPortMatchesGoCohere_0387, TestPortMatchesGoCohere_0386, TestPortMatchesGoCohere_0385, TestPortMatchesGoCohere_0384, TestPortMatchesGoCohere_0383, TestPortMatchesGoCohere_0382, TestPortMatchesGoCohere_0381, TestPortMatchesGoCohere_0327, TestPortMatchesGoCohere_0380, TestPortMatchesGoCohere_0379, TestPortMatchesGoCohere_0378, TestPortMatchesGoCohere_0248, TestPortMatchesGoCohere_0376, TestPortMatchesGoCohere_0375, TestPortMatchesGoCohere_0374, TestPortMatchesGoCohere_0373, TestPortMatchesGoCohere_0372, TestPortMatchesGoCohere_0371, TestPortMatchesGoCohere_0370, TestPortMatchesGoCohere_0409, TestPortMatchesGoCohere_0369, TestPortMatchesGoCohere_0367, TestPortMatchesGoCohere_0368, TestPortMatchesGoCohere_0366, TestPortMatchesGoCohere_0365, TestPortMatchesGoCohere_0364, TestPortMatchesGoCohere_0363, TestPortMatchesGoCohere_0362, TestPortMatchesGoCohere_0361, TestPortMatchesGoCohere_0360, TestPortMatchesGoCohere_0359, TestPortMatchesGoCohere_0358, TestPortMatchesGoCohere_0357, TestPortMatchesGoCohere_0356, TestPortMatchesGoCohere_0355, TestPortMatchesGoCohere_0354, TestPortMatchesGoCohere_0353, TestPortMatchesGoCohere_0352, TestPortMatchesGoCohere_0351, TestPortMatchesGoCohere_0350, TestPortMatchesGoCohere_0349, TestPortMatchesGoCohere_0348, TestPortMatchesGoCohere_0347, TestPortMatchesGoCohere_0346, TestPortMatchesGoCohere_0345, TestPortMatchesGoCohere_0344, TestPortMatchesGoCohere_0343, TestPortMatchesGoCohere_0342, TestPortMatchesGoCohere_0341, TestPortMatchesGoCohere_0340, TestPortMatchesGoCohere_0339, TestPortMatchesGoCohere_0338, TestPortMatchesGoCohere_0337, TestPortMatchesGoCohere_0336, TestPortMatchesGoCohere_0335, TestPortMatchesGoCohere_0333, TestPortMatchesGoCohere_0334, TestPortMatchesGoCohere_0332, TestPortMatchesGoCohere_0331, TestPortMatchesGoCohere_0330, TestPortMatchesGoCohere_0329, TestPortMatchesGoCohere_0328, TestPortMatchesGoCohere_0326, TestPortMatchesGoCohere_0324, TestPortMatchesGoCohere_0325, TestPortMatchesGoCohere_0323, TestPortMatchesGoCohere_0322, TestPortMatchesGoCohere_0321, TestPortMatchesGoCohere_0320, TestPortMatchesGoCohere_0273, TestPortMatchesGoCohere_0319, TestPortMatchesGoCohere_0318, TestPortMatchesGoCohere_0317, TestPortMatchesGoCohere_0316, TestPortMatchesGoCohere_0315, TestPortMatchesGoCohere_0314, TestPortMatchesGoCohere_0313, TestPortMatchesGoCohere_0312, TestPortMatchesGoCohere_0311, TestPortMatchesGoCohere_0310, TestPortMatchesGoCohere_0309, TestPortMatchesGoCohere_0308, TestPortMatchesGoCohere_0307, TestPortMatchesGoCohere_0306, TestPortMatchesGoCohere_0305, TestPortMatchesGoCohere_0304, TestPortMatchesGoCohere_0303, TestPortMatchesGoCohere_0302, TestPortMatchesGoCohere_0301, TestPortMatchesGoCohere_0300, TestPortMatchesGoCohere_0299, TestPortMatchesGoCohere_0298, TestPortMatchesGoCohere_0297, TestPortMatchesGoCohere_0296, TestPortMatchesGoCohere_0295, TestPortMatchesGoCohere_0294, TestPortMatchesGoCohere_0293, TestPortMatchesGoCohere_0292, TestPortMatchesGoCohere_0291, TestPortMatchesGoCohere_0290, TestPortMatchesGoCohere_0289, TestPortMatchesGoCohere_0288, TestPortMatchesGoCohere_0287, TestPortMatchesGoCohere_0286, TestPortMatchesGoCohere_0285, TestPortMatchesGoCohere_0284, TestPortMatchesGoCohere_0283, TestPortMatchesGoCohere_0282, TestPortMatchesGoCohere_0281, TestPortMatchesGoCohere_0280, TestPortMatchesGoCohere_0278, TestPortMatchesGoCohere_0392, TestPortMatchesGoCohere_0279, TestPortMatchesGoCohere_0276, TestPortMatchesGoCohere_0275, TestPortMatchesGoCohere_0272, TestPortMatchesGoCohere_0271, TestPortMatchesGoCohere_0270, TestPortMatchesGoCohere_0269, TestPortMatchesGoCohere_0268, TestPortMatchesGoCohere_0267, TestPortMatchesGoCohere_0266, TestPortMatchesGoCohere_0265, TestPortMatchesGoCohere_0274, TestPortMatchesGoCohere_0264, TestPortMatchesGoCohere_0277, TestPortMatchesGoCohere_0262, TestPortMatchesGoCohere_0403, TestPortMatchesGoCohere_0261, TestPortMatchesGoCohere_0259, TestPortMatchesGoCohere_0263, TestPortMatchesGoCohere_0233, TestPortMatchesGoCohere_0256, TestPortMatchesGoCohere_0257, TestPortMatchesGoCohere_0255, TestPortMatchesGoCohere_0254, TestPortMatchesGoCohere_0253, TestPortMatchesGoCohere_0252, TestPortMatchesGoCohere_0251, TestPortMatchesGoCohere_0250, TestPortMatchesGoCohere_0249, TestPortMatchesGoCohere_0119, TestPortMatchesGoCohere_0246, TestPortMatchesGoCohere_0245, TestPortMatchesGoCohere_0244, TestPortMatchesGoCohere_0243, TestPortMatchesGoCohere_0242, TestPortMatchesGoCohere_0241, TestPortMatchesGoCohere_0240, TestPortMatchesGoCohere_0239, TestPortMatchesGoCohere_0238, TestPortMatchesGoCohere_0237, TestPortMatchesGoCohere_0236, TestPortMatchesGoCohere_0235, TestPortMatchesGoCohere_0234, TestPortMatchesGoCohere_0232, TestPortMatchesGoCohere_0231, TestPortMatchesGoCohere_0230, TestPortMatchesGoCohere_0229, TestPortMatchesGoCohere_0228, TestPortMatchesGoCohere_0227, TestPortMatchesGoCohere_0226, TestPortMatchesGoCohere_0225, TestPortMatchesGoCohere_0224, TestPortMatchesGoCohere_0223, TestPortMatchesGoCohere_0222, TestPortMatchesGoCohere_0221, TestPortMatchesGoCohere_0220, TestPortMatchesGoCohere_0219, TestPortMatchesGoCohere_0218, TestPortMatchesGoCohere_0217, TestPortMatchesGoCohere_0216, TestPortMatchesGoCohere_0215, TestPortMatchesGoCohere_0214, TestPortMatchesGoCohere_0213, TestPortMatchesGoCohere_0212, TestPortMatchesGoCohere_0211, TestPortMatchesGoCohere_0210, TestPortMatchesGoCohere_0209, TestPortMatchesGoCohere_0208, TestPortMatchesGoCohere_0207, TestPortMatchesGoCohere_0206, TestPortMatchesGoCohere_0205, TestPortMatchesGoCohere_0204, TestPortMatchesGoCohere_0203, TestPortMatchesGoCohere_0202, TestPortMatchesGoCohere_0201, TestPortMatchesGoCohere_0200, TestPortMatchesGoCohere_0199, TestPortMatchesGoCohere_0198, TestPortMatchesGoCohere_0181, TestPortMatchesGoCohere_0197, TestPortMatchesGoCohere_0196, TestPortMatchesGoCohere_0195, TestPortMatchesGoCohere_0194, TestPortMatchesGoCohere_0193, TestPortMatchesGoCohere_0192, TestPortMatchesGoCohere_0191, TestPortMatchesGoCohere_0190, TestPortMatchesGoCohere_0189, TestPortMatchesGoCohere_0188, TestPortMatchesGoCohere_0187, TestPortMatchesGoCohere_0186, TestPortMatchesGoCohere_0185, TestPortMatchesGoCohere_0184, TestPortMatchesGoCohere_0183, TestPortMatchesGoCohere_0182, TestPortMatchesGoCohere_0180, TestPortMatchesGoCohere_0179, TestPortMatchesGoCohere_0178, TestPortMatchesGoCohere_0177, TestPortMatchesGoCohere_0176, TestPortMatchesGoCohere_0175, TestPortMatchesGoCohere_0174, TestPortMatchesGoCohere_0173, TestPortMatchesGoCohere_0172, TestPortMatchesGoCohere_0171, TestPortMatchesGoCohere_0170, TestPortMatchesGoCohere_0169, TestPortMatchesGoCohere_0168, TestPortMatchesGoCohere_0167, TestPortMatchesGoCohere_0166, TestPortMatchesGoCohere_0165, TestPortMatchesGoCohere_0164, TestPortMatchesGoCohere_0163, TestPortMatchesGoCohere_0162, TestPortMatchesGoCohere_0161, TestPortMatchesGoCohere_0160, TestPortMatchesGoCohere_0159, TestPortMatchesGoCohere_0158, TestPortMatchesGoCohere_0157, TestPortMatchesGoCohere_0156, TestPortMatchesGoCohere_0155, TestPortMatchesGoCohere_0154, TestPortMatchesGoCohere_0153, TestPortMatchesGoCohere_0152, TestPortMatchesGoCohere_0151, TestPortMatchesGoCohere_0150, TestPortMatchesGoCohere_0149, TestPortMatchesGoCohere_0148, TestPortMatchesGoCohere_0147, TestPortMatchesGoCohere_0146, TestPortMatchesGoCohere_0145, TestPortMatchesGoCohere_0144, TestPortMatchesGoCohere_0143, TestPortMatchesGoCohere_0142, TestPortMatchesGoCohere_0141, TestPortMatchesGoCohere_0140, TestPortMatchesGoCohere_0139, TestPortMatchesGoCohere_0138, TestPortMatchesGoCohere_0137, TestPortMatchesGoCohere_0136, TestPortMatchesGoCohere_0135, TestPortMatchesGoCohere_0134, TestPortMatchesGoCohere_0133, TestPortMatchesGoCohere_0132, TestPortMatchesGoCohere_0131, TestPortMatchesGoCohere_0130, TestPortMatchesGoCohere_0129, TestPortMatchesGoCohere_0128, TestPortMatchesGoCohere_0127, TestPortMatchesGoCohere_0126, TestPortMatchesGoCohere_0125, TestPortMatchesGoCohere_0124, TestPortMatchesGoCohere_0123, TestPortMatchesGoCohere_0122, TestPortMatchesGoCohere_0121, TestPortMatchesGoCohere_0120, TestAdditionalJSONBoundaries, TestPinnedRepositoryCorpusIdentity, TestPortMatchesGoCohere_0260, TestPortMatchesGoCohere_0247, TestPortMatchesGoCohere_0258, TestSingleFileStdoutDriver, TestThreePortMutantsAreCaught, TestProduct_JsonFindsGoOracle, TestPortMatchesGoCohereSplit_Setup, TestCorpusPinRejectsMissingProvisionedInputs, TestPortMatchesGoCohere_1556, TestPortMatchesGoCohere_1576, TestPortMatchesGoCohere_1575, TestPortMatchesGoCohere_1574, TestPortMatchesGoCohere_1573, TestPortMatchesGoCohere_1572, TestPortMatchesGoCohere_1571, TestPortMatchesGoCohere_1570, TestPortMatchesGoCohere_1569, TestPortMatchesGoCohere_1568, TestPortMatchesGoCohere_1567, TestPortMatchesGoCohere_1566, TestPortMatchesGoCohere_1565, TestPortMatchesGoCohere_1564, TestPortMatchesGoCohere_1563, TestPortMatchesGoCohere_1562, TestPortMatchesGoCohere_1561, TestPortMatchesGoCohere_1560, TestPortMatchesGoCohere_1559, TestPortMatchesGoCohere_1558, TestPortMatchesGoCohere_1557, TestProduct_JsonFindsClassOrderAnswer, TestPortMatchesGoCohereUnion, TestPortMatchesGoCohere_1986, TestPortMatchesGoCohere_2047, TestPortMatchesGoCohere_2046, TestPortMatchesGoCohere_2045, TestPortMatchesGoCohere_2044, TestPortMatchesGoCohere_2043, TestPortMatchesGoCohere_2042, TestPortMatchesGoCohere_2041, TestPortMatchesGoCohere_2040, TestPortMatchesGoCohere_2039, TestPortMatchesGoCohere_2038, TestPortMatchesGoCohere_2037, TestPortMatchesGoCohere_2036, TestPortMatchesGoCohere_2035, TestPortMatchesGoCohere_2034, TestPortMatchesGoCohere_2033, TestPortMatchesGoCohere_2032, TestPortMatchesGoCohere_2031, TestPortMatchesGoCohere_2030, TestPortMatchesGoCohere_2029, TestPortMatchesGoCohere_2028, TestPortMatchesGoCohere_2027, TestPortMatchesGoCohere_2026, TestPortMatchesGoCohere_2025, TestPortMatchesGoCohere_2024, TestPortMatchesGoCohere_2023, TestPortMatchesGoCohere_2022, TestPortMatchesGoCohere_2021, TestPortMatchesGoCohere_2020, TestPortMatchesGoCohere_2019, TestPortMatchesGoCohere_2018, TestPortMatchesGoCohere_2017, TestPortMatchesGoCohere_2016, TestPortMatchesGoCohere_2015, TestPortMatchesGoCohere_2014, TestPortMatchesGoCohere_2013, TestPortMatchesGoCohere_2012, TestPortMatchesGoCohere_2011, TestPortMatchesGoCohere_2010, TestPortMatchesGoCohere_2009, TestPortMatchesGoCohere_2008, TestPortMatchesGoCohere_2007, TestPortMatchesGoCohere_2006, TestPortMatchesGoCohere_2005, TestPortMatchesGoCohere_2004, TestPortMatchesGoCohere_2003, TestPortMatchesGoCohere_2002, TestPortMatchesGoCohere_2001, TestPortMatchesGoCohere_2000, TestPortMatchesGoCohere_1999, TestPortMatchesGoCohere_1998, TestPortMatchesGoCohere_1997, TestPortMatchesGoCohere_1996, TestPortMatchesGoCohere_1995, TestPortMatchesGoCohere_1994, TestPortMatchesGoCohere_1993, TestPortMatchesGoCohere_1992, TestPortMatchesGoCohere_1991, TestPortMatchesGoCohere_1990, TestPortMatchesGoCohere_1989, TestPortMatchesGoCohere_1988, TestPortMatchesGoCohere_1987, TestProduct_JsonFindsPrinterMutant_2, TestPortMatchesGoCohere_1954, TestPortMatchesGoCohere_1985, TestPortMatchesGoCohere_1984, TestPortMatchesGoCohere_1983, TestPortMatchesGoCohere_1982, TestPortMatchesGoCohere_1981, TestPortMatchesGoCohere_1980, TestPortMatchesGoCohere_1979, TestPortMatchesGoCohere_1978, TestPortMatchesGoCohere_1977, TestPortMatchesGoCohere_1976, TestPortMatchesGoCohere_1975, TestPortMatchesGoCohere_1974, TestPortMatchesGoCohere_1973, TestPortMatchesGoCohere_1972, TestPortMatchesGoCohere_1971, TestPortMatchesGoCohere_1970, TestPortMatchesGoCohere_1969, TestPortMatchesGoCohere_1968, TestPortMatchesGoCohere_1967, TestPortMatchesGoCohere_1966, TestPortMatchesGoCohere_1965, TestPortMatchesGoCohere_1964, TestPortMatchesGoCohere_1963, TestPortMatchesGoCohere_1962, TestPortMatchesGoCohere_1961, TestPortMatchesGoCohere_1960, TestPortMatchesGoCohere_1959, TestPortMatchesGoCohere_1958, TestPortMatchesGoCohere_1957, TestPortMatchesGoCohere_1956, TestPortMatchesGoCohere_1955, TestProduct_JsonFindsPrinterMutant_1, TestProduct_JsonFindsPrinterMutant_0, TestPortMatchesGoCohere_1938, TestPortMatchesGoCohere_1953, TestPortMatchesGoCohere_1952, TestPortMatchesGoCohere_1951, TestPortMatchesGoCohere_1950, TestPortMatchesGoCohere_1949, TestPortMatchesGoCohere_1948, TestPortMatchesGoCohere_1947, TestPortMatchesGoCohere_1946, TestPortMatchesGoCohere_1945, TestPortMatchesGoCohere_1944, TestPortMatchesGoCohere_1943, TestPortMatchesGoCohere_1942, TestPortMatchesGoCohere_1941, TestPortMatchesGoCohere_1940, TestPortMatchesGoCohere_1939, TestPortMatchesGoCohere_1546, TestPortMatchesGoCohere_1555, TestPortMatchesGoCohere_1554, TestPortMatchesGoCohere_1553, TestPortMatchesGoCohere_1552, TestPortMatchesGoCohere_1551, TestPortMatchesGoCohere_1550, TestPortMatchesGoCohere_1549, TestPortMatchesGoCohere_1548, TestPortMatchesGoCohere_1547, TestPortMatchesGoCohere_1541, TestPortMatchesGoCohere_1545, TestPortMatchesGoCohere_1544, TestPortMatchesGoCohere_1543, TestPortMatchesGoCohere_1542, TestPortMatchesGoCohere_1409, TestPortMatchesGoCohere_1540, TestPortMatchesGoCohere_1539, TestPortMatchesGoCohere_1930, TestPortMatchesGoCohere_1937, TestPortMatchesGoCohere_1936, TestPortMatchesGoCohere_1935, TestPortMatchesGoCohere_1934, TestPortMatchesGoCohere_1933, TestPortMatchesGoCohere_1932, TestPortMatchesGoCohere_1931, TestPortMatchesGoCohere_1926, TestPortMatchesGoCohere_1929, TestPortMatchesGoCohere_1928, TestPortMatchesGoCohere_1927, TestPortMatchesGoCohere_1537, TestPortMatchesGoCohere_1796, TestPortMatchesGoCohere_1925, TestPortMatchesGoCohere_1536, TestPortMatchesGoCohere_1924 |
| final-oracle | internal/oracle/... | 1 | 616.768 | TestEnumInitializationUnknownPinned/unknown-before, TestEnumInitializationUnknownPinned/unknown-property, TestEnumInitializationUnknownPinned/unknown-callback, TestEnumInitializationUnknownPinned, TestAsyncFixturesAgreeWithNode/async_throw.a, TestAsyncFixturesAgreeWithNode/async_coverage_reject_empty.a, TestAsyncFixturesAgreeWithNode/async_coverage_reject_eager.a, TestAsyncFixturesAgreeWithNode/async_coverage_reject_nested.a, TestAsyncFixturesAgreeWithNode |
| owners-final | stage1/cohere/css/oracles | 0 | 2.708 |  |
| owners-final | stage1/cohere/graphql/oracles | 0 | 0.88 |  |
| owners-final | stage1/cohere/json/oracles | 0 | 2.262 |  |
| oracle-contract-final | internal/oracle | 0 | 33.004 |  |
| owners-final | stage1/cohere/json | 0 | 36.38 |  |
| owners-final | stage1/cohere/markdownblocks/oracles | 0 | 20.583 |  |
| guards-delivery | cmd/adamic-gate | 0 | 2.531 |  |
| guards-delivery | internal/ir | 0 | 10.295 |  |
| guards-delivery | internal/native | 0 | 6.334 |  |
| guards-delivery | internal/lower | 0 | 1.875 |  |
| guards-delivery | internal/oracle | 0 | 2.889 |  |
| lint-own-final | stage1/cohere/lint | 1 | 136.608 | TestCommentFoldMutant, TestPositionIndexMutant, TestProfileCompilation_000, TestProfileCompilationPlantedFailure, TestProfileCompilation_Setup |
| lower-delivery | internal/lower/... | 0 | 107.69 |  |
| reader-final | stage1/cohere/css/... | 1 | 319.752 | TestCSSPrinterPreflightCostMutant |
| missing-owners-final | stage1/cohere/lint/oracles | 1 | 118.726 | TestProduct_jsx_parser, TestProduct_LintGoOracle, TestProduct_YepestaSetup, TestProduct_SuggestionAlongsideGoOracle, TestProduct_ProfileCompilationOracleOutputs, TestProduct_CompilerAgreementCheckerAllAnswer, TestProduct_jsx_membership, TestProduct_RulesAgreeCapture, TestProduct_CompleteSuggestionGoOracle |
| reader-final | stage1/cohere/graphql/... | 1 | 92.555 |  |
| reader-final | stage1/cohere/json/... | 1 | 136.306 | TestPortMatchesGoCohereClassOrder_001, TestPortMatchesGoCohereInventory, TestPortMatchesGoCohere_1643, TestPortMatchesGoCohere_1396 |

```text
go test -p 1 -count=1 -json -parallel=4 -timeout 11m ./cmd/adamic-gate -run ^(TestEveryTestHasOneDescriptiveName|TestEveryTestIsParallelOrSaysWhy|TestEveryMapAProgramHoldsIsSorted|TestCallTargetReaders|TestRuntimeStaticsAreListed|TestNativeAgreementRowsNameRegisteredPrograms|TestRuntimeSetsAreEachUsed)$
go test -p 1 -count=1 -json -parallel=4 -timeout 90s ./cmd/adamic-gate -run ^(TestEveryTestHasOneDescriptiveName|TestEveryTestIsParallelOrSaysWhy|TestEveryMapAProgramHoldsIsSorted|TestCallTargetReaders|TestRuntimeStaticsAreListed|TestNativeAgreementRowsNameRegisteredPrograms|TestRuntimeSetsAreEachUsed)$
go test -p 1 -count=1 -json -parallel=4 -timeout 90s ./internal/ir -run ^(TestEveryTestHasOneDescriptiveName|TestEveryTestIsParallelOrSaysWhy|TestEveryMapAProgramHoldsIsSorted|TestCallTargetReaders|TestRuntimeStaticsAreListed|TestNativeAgreementRowsNameRegisteredPrograms|TestRuntimeSetsAreEachUsed)$
go test -p 1 -count=1 -json -parallel=4 -timeout 90s ./internal/native -run ^(TestEveryTestHasOneDescriptiveName|TestEveryTestIsParallelOrSaysWhy|TestEveryMapAProgramHoldsIsSorted|TestCallTargetReaders|TestRuntimeStaticsAreListed|TestNativeAgreementRowsNameRegisteredPrograms|TestRuntimeSetsAreEachUsed)$
go test -p 1 -count=1 -json -parallel=4 -timeout 90s ./internal/native -run ^(TestRuntimeStaticsAreListed|TestNativeAgreementRowsNameRegisteredPrograms|TestRuntimeSetsAreEachUsed)$
go test -p 1 -count=1 -json -parallel=4 -timeout 90s ./internal/lower -run ^(TestRuntimeStaticsAreListed|TestNativeAgreementRowsNameRegisteredPrograms|TestRuntimeSetsAreEachUsed)$
go test -p 1 -count=1 -json -parallel=4 -timeout 90s ./internal/oracle -run ^(TestRuntimeStaticsAreListed|TestNativeAgreementRowsNameRegisteredPrograms|TestRuntimeSetsAreEachUsed)$
go test -p 1 -count=1 -json -parallel=4 -timeout 11m ./internal/lower/...
go test -p 1 -count=1 -json -parallel=4 -timeout 11m ./internal/lower/...
go test -p 1 -count=1 -json -parallel=4 -timeout 11m ./internal/native -run ^TestProduct_TSGoArchive$
go test -p 1 -count=1 -json -parallel=4 -timeout 11m ./internal/native/...
go test -p 1 -count=1 -json -parallel=4 -timeout 11m ./internal/native/... -skip ^TestSplitTSGoAgreesUnit00$
go test -p 1 -count=1 -json -parallel=4 -timeout 11m ./internal/native -run ^(TestOrdinaryToPrimitive.*|TestProduct_PrimitiveMutants|TestRetained.*Coverage|TestSplitAttributedBindingsAgreeWithNode|TestParallelUndefinedException.*|TestRuntimeFeature.*|TestSplitTSGoRuntimeFeatures)$
go test -p 1 -count=1 -json -parallel=4 -timeout 11m ./internal/oracle -run ^TestProduct_NativeRuntimes(Sanitized|Release|Counted)
go test -p 1 -count=1 -json -parallel=4 -timeout 11m ./internal/oracle -run ^(TestStep21|TestNarrowedUnion|TestInheritedFields)
go test -p 1 -count=1 -json -parallel=4 -timeout 11m ./internal/oracle -run ^TestSpreadUndefinedThrowStopsBeforeNextGetter$
go test -p 1 -count=1 -json -parallel=4 -timeout 11m ./internal/oracle -run ^TestSpreadUndefinedThrowStopsBeforeNextGetter$
go test -p 1 -count=1 -json -parallel=4 -timeout 11m ./internal/oracle -run ^TestErrorSubclassAncestryAgreesWithNode$
go test -p 1 -count=1 -json -parallel=4 -timeout 11m ./internal/flow/...
go test -p 1 -count=1 -json -parallel=4 -timeout 90s ./cmd/adamic-gate -run ^(TestEveryTestHasOneDescriptiveName|TestEveryTestIsParallelOrSaysWhy|TestCallTargetReaders|TestEveryMapAProgramHoldsIsSorted|TestRuntimeStaticsAreListed|TestNativeAgreementRowsNameRegisteredPrograms|TestRuntimeSetsAreEachUsed)$
go test -p 1 -count=1 -json -parallel=4 -timeout 90s ./internal/ir -run ^(TestEveryTestHasOneDescriptiveName|TestEveryTestIsParallelOrSaysWhy|TestCallTargetReaders|TestEveryMapAProgramHoldsIsSorted|TestRuntimeStaticsAreListed|TestNativeAgreementRowsNameRegisteredPrograms|TestRuntimeSetsAreEachUsed)$
go test -p 1 -count=1 -json -parallel=4 -timeout 90s ./internal/native -run ^(TestEveryTestHasOneDescriptiveName|TestEveryTestIsParallelOrSaysWhy|TestCallTargetReaders|TestEveryMapAProgramHoldsIsSorted|TestRuntimeStaticsAreListed|TestNativeAgreementRowsNameRegisteredPrograms|TestRuntimeSetsAreEachUsed)$
go test -p 1 -count=1 -json -parallel=4 -timeout 90s ./internal/lower -run ^(TestEveryTestHasOneDescriptiveName|TestEveryTestIsParallelOrSaysWhy|TestCallTargetReaders|TestEveryMapAProgramHoldsIsSorted|TestRuntimeStaticsAreListed|TestNativeAgreementRowsNameRegisteredPrograms|TestRuntimeSetsAreEachUsed)$
go test -p 1 -count=1 -json -parallel=4 -timeout 90s ./internal/oracle -run ^(TestEveryTestHasOneDescriptiveName|TestEveryTestIsParallelOrSaysWhy|TestCallTargetReaders|TestEveryMapAProgramHoldsIsSorted|TestRuntimeStaticsAreListed|TestNativeAgreementRowsNameRegisteredPrograms|TestRuntimeSetsAreEachUsed)$
go test -p 1 -count=1 -json -parallel=4 -timeout 11m ./internal/flow/...
go test -p 1 -count=1 -json -parallel=4 -timeout 11m ./internal/fresh/...
go test -p 1 -count=1 -json -parallel=4 -timeout 11m ./internal/javascript/...
go test -p 1 -count=1 -json -parallel=4 -timeout 11m ./internal/fuzz/...
go test -p 1 -count=1 -json -parallel=4 -timeout 11m ./stage1/cohere/css/...
go test -p 1 -count=1 -json -parallel=4 -timeout 11m ./cmd/adamic/...
go test -p 1 -count=1 -json -parallel=4 -timeout 11m ./cmd/adamic-test262/...
go test -p 1 -count=1 -json -parallel=4 -timeout 11m ./stage3/fixtures/...
go test -p 1 -count=1 -json -parallel=4 -timeout 11m ./internal/oracle/...
go test -p 1 -count=1 -json -parallel=4 -timeout 11m ./internal/ir/...
go test -p 1 -count=1 -json -parallel=4 -timeout 11m ./internal/lower -run ^(TestStep21|TestSwitchCase|TestSmoke|TestReviewPrograms|TestParserNamespaceClassRegistrationMutant|TestNodeFSRmForceUndefinedReproducer|TestModuleNamespaceReadinessMutants|TestLibraryMethodAliasReadMutant|TestEnumCleanupMutant|TestFreshWrites____oracle_testdata_step21_.*)$
go test -p 1 -count=1 -json -parallel=4 -timeout 11m ./internal/oracle -run ^(TestStep21|TestSwitchCase|TestSmoke|TestReviewPrograms|TestParserNamespaceClassRegistrationMutant|TestNodeFSRmForceUndefinedReproducer|TestModuleNamespaceReadinessMutants|TestLibraryMethodAliasReadMutant|TestEnumCleanupMutant|TestFreshWrites____oracle_testdata_step21_.*)$
go test -p 1 -count=1 -json -parallel=4 -timeout 11m ./internal/fresh -run ^(TestStep21|TestSwitchCase|TestSmoke|TestReviewPrograms|TestParserNamespaceClassRegistrationMutant|TestNodeFSRmForceUndefinedReproducer|TestModuleNamespaceReadinessMutants|TestLibraryMethodAliasReadMutant|TestEnumCleanupMutant|TestFreshWrites____oracle_testdata_step21_.*)$
go test -p 1 -count=1 -json -parallel=4 -timeout 11m ./internal/fresh/...
go test -p 1 -count=1 -json -parallel=4 -timeout 11m ./internal/oracle -run ^(TestStep21|TestNarrowed|TestInheritedFields|TestSmoke|TestReviewPrograms)
go test -p 1 -count=1 -json -parallel=4 -timeout 11m ./internal/fuzz/...
go test -p 1 -count=1 -json -parallel=4 -timeout 11m ./internal/lower/...
go test -p 1 -count=1 -json -parallel=4 -timeout 11m ./internal/ir/...
go test -p 1 -count=1 -json -parallel=4 -timeout 11m ./internal/native/... -skip ^TestSplitTSGoAgreesUnit00$
go test -p 1 -count=1 -json -parallel=4 -timeout 11m ./internal/lower/...
go test -p 1 -count=1 -json -parallel=4 -timeout 11m ./internal/flow/...
go test -p 1 -count=1 -json -parallel=4 -timeout 11m ./internal/fresh/...
go test -p 1 -count=1 -json -parallel=4 -timeout 11m ./internal/fuzz/...
go test -p 1 -count=1 -json -parallel=4 -timeout 11m ./internal/native/runtimeproducts/...
go test -p 1 -count=1 -json -parallel=4 -timeout 11m ./internal/lower/cycleledger/...
go test -p 1 -count=1 -json -parallel=4 -timeout 11m ./stage1/cohere/css/...
go test -p 1 -count=1 -json -parallel=4 -timeout 11m ./stage1/cohere/graphql/...
go test -p 1 -count=1 -json -parallel=4 -timeout 11m ./stage1/cohere/json/...
go test -p 1 -count=1 -json -parallel=4 -timeout 11m ./internal/oracle/...
go test -p 1 -count=1 -json -parallel=4 -timeout 11m ./stage1/cohere/css/oracles -run ^TestProduct_
go test -p 1 -count=1 -json -parallel=4 -timeout 11m ./stage1/cohere/graphql/oracles -run ^TestProduct_
go test -p 1 -count=1 -json -parallel=4 -timeout 11m ./stage1/cohere/json/oracles -run ^TestProduct_
go test -p 1 -count=1 -json -parallel=4 -timeout 11m ./internal/oracle -run ^(TestStep21|TestNarrowed|TestInheritedFields|TestEnumInitializationUnknownPinned|TestAsync|TestSmoke)
go test -p 1 -count=1 -json -parallel=4 -timeout 11m ./stage1/cohere/json -run ^TestProduct_
go test -p 1 -count=1 -json -parallel=4 -timeout 11m ./stage1/cohere/markdownblocks/oracles -run ^TestProduct_ -skip ^(TestProduct_WholeDocumentPreflightGoProducts|TestProduct_MarkdownblocksTextFormatter|TestProduct_GrainMarkdownFormatter)$
go test -p 1 -count=1 -json -parallel=4 -timeout 90s ./cmd/adamic-gate -run ^(TestEveryTestHasOneDescriptiveName|TestEveryTestIsParallelOrSaysWhy|TestCallTargetReaders|TestEveryMapAProgramHoldsIsSorted|TestRuntimeStaticsAreListed|TestNativeAgreementRowsNameRegisteredPrograms|TestRuntimeSetsAreEachUsed)$
go test -p 1 -count=1 -json -parallel=4 -timeout 90s ./internal/ir -run ^(TestEveryTestHasOneDescriptiveName|TestEveryTestIsParallelOrSaysWhy|TestCallTargetReaders|TestEveryMapAProgramHoldsIsSorted|TestRuntimeStaticsAreListed|TestNativeAgreementRowsNameRegisteredPrograms|TestRuntimeSetsAreEachUsed)$
go test -p 1 -count=1 -json -parallel=4 -timeout 90s ./internal/native -run ^(TestEveryTestHasOneDescriptiveName|TestEveryTestIsParallelOrSaysWhy|TestCallTargetReaders|TestEveryMapAProgramHoldsIsSorted|TestRuntimeStaticsAreListed|TestNativeAgreementRowsNameRegisteredPrograms|TestRuntimeSetsAreEachUsed)$
go test -p 1 -count=1 -json -parallel=4 -timeout 90s ./internal/lower -run ^(TestEveryTestHasOneDescriptiveName|TestEveryTestIsParallelOrSaysWhy|TestCallTargetReaders|TestEveryMapAProgramHoldsIsSorted|TestRuntimeStaticsAreListed|TestNativeAgreementRowsNameRegisteredPrograms|TestRuntimeSetsAreEachUsed)$
go test -p 1 -count=1 -json -parallel=4 -timeout 90s ./internal/oracle -run ^(TestEveryTestHasOneDescriptiveName|TestEveryTestIsParallelOrSaysWhy|TestCallTargetReaders|TestEveryMapAProgramHoldsIsSorted|TestRuntimeStaticsAreListed|TestNativeAgreementRowsNameRegisteredPrograms|TestRuntimeSetsAreEachUsed)$
go test -p 1 -count=1 -json -parallel=4 -timeout 11m ./stage1/cohere/lint -run ^(TestProfile|TestCommentFoldMutant|TestPositionIndexMutant)
go test -p 1 -count=1 -json -parallel=4 -timeout 11m ./internal/lower/...
go test -p 1 -count=1 -json -parallel=4 -timeout 11m ./stage1/cohere/css/...
go test -p 1 -count=1 -json -parallel=4 -timeout 11m ./stage1/cohere/lint/oracles -run ^TestProduct_
go test -p 1 -count=1 -json -parallel=4 -timeout 11m ./stage1/cohere/graphql/...
go test -p 1 -count=1 -json -parallel=4 -timeout 11m ./stage1/cohere/json/...
```

The original cold TestSplitTSGoAgreesUnit00 took 93.07s and failed while linking the checker archive with no space left on device. It was not rerun. Later native suites explicitly exclude only that leaf. Other initial failures and their fixes remain visible in the logs (new runtime source whitelist, pending-aware TypeError mutant, WASI SDK and coverage census). Tests skipped by their own opt-in/environment contracts are not claimed executed.

## Mutant replay

| Mutant | Recorded caught | Exit | Seconds | Expected catcher |
| --- | --- | ---: | ---: | --- |
| missing-exception-edge | True | 1 | 16.923 | runtime error: |
| assignment-kills-handler-value | True | 1 | 13.255 | old text is dead |
| throw-temporaries-leak | True | 1 | 10.881 | leaks: |
| pending-finally-leak | True | 1 | 10.541 | leaks: |
| error-message-not-retained | True | 1 | 25.133 | AddressSanitizer: heap-use-after-free |
| range-error-is-error | True | 1 | 15.773 | stdout differs |
| wrong-range-message | True | 1 | 15.559 | stdout differs |
| type-error-is-error | True | 1 | 15.361 | stdout differs |
| hash-failure-is-panic | True | 1 | 24.551 | exit codes differ |
| host-type-error-is-error | True | 1 | 24.103 | stdout differs |
| soundness-check-is-catchable | True | 1 | 15.002 | per-use stop |
| error-ancestry-missing | False | 1 | 10.641 | stdout differs |
| error-default-name-wrong | True | 1 | 16.078 | stdout differs |
| error-prefix-released-twice | True | 1 | 15.109 | AddressSanitizer: heap-use-after-free |
| error-fields-out-of-order | True | 1 | 15.901 | stdout differs |
| implicit-string-conversion | False | 1 | 23.72 | uncaught diagnostic |
| pending-undefined-lost | True | 1 | 9.981 | stdout differs |
| catch-assumes-error | True | 1 | 15.768 | exit codes differ |
| typeof-boolean-is-number | True | 1 | 24.356 | stdout differs |
| null-is-undefined | True | 1 | 10.649 | stdout differs |
| finally-releases-twice | True | 1 | 10.33 | AddressSanitizer: heap-use-after-free |
| region-payload-freed | True | 1 | 9.986 | AddressSanitizer: heap-use-after-free |
| uncaught-is-panic | True | 1 | 23.718 | stdout/exit differ |
| changed-source-keeps-terminal-mode | True | 1 | 1.923 | changed source borrowed terminal convention |
| changed-dependency-keeps-terminal-mode | True | 1 | 1.874 | changed dependency borrowed terminal convention |
| backend-drops-entry-source | True | 1 | 9.679 | JavaScript backend: exit codes differ |
| readiness-old-panic | True | 1 | 10.017 | JavaScript language TDZ |
| rethrow-wrapped | True | 1 | 12.903 | stdout differs |
| builtin-narrow-forgets-subtype | True | 1 | 17.267 | per-use stop |
| saved-error-wrapped | True | 1 | 12.699 | stdout differs |
| backend-renders-stack | True | 1 | 9.668 | uncaught diagnostic |
| backend-drops-line | True | 1 | 9.778 | uncaught diagnostic |
| native-renders-stack | True | 1 | 23.793 | uncaught diagnostic |
| native-drops-line | True | 1 | 23.846 | uncaught diagnostic |
| drop-write-set-test | True | 1 | 7.718 | narrow again after the call |
| admit-writing-call | True | 1 | 7.959 | writing-call refusal: got <nil> |
| whole-narrowed-type | True | 1 | 14.506 | .name agreement |
| native-string-length | True | 1 | 23.479 | primitive length |
| native-scalar-properties | False | 1 | 24.336 | stale narrowing use failed |
| numeric-enum-as-finite | True | 1 | 17.393 | stored flags |
| M02 | True | 1 | 12.183 | B.Fields |
| M07 | True | 1 | 13.256 | want NotYet |
| P_LOWER | True | 1 | 13.559 | accepted inheritance must produce |
| readiness-old-effects | True | 1 | 25.573 | stdout differs |
| unknown-authorizes-assertion | True | 1 | 22.483 | want Refused, got <nil> |
| fixed-symbol | True | 1 | 23.417 | mismatched runtime linked successfully |
| drop-reference | True | 1 | 14.747 | mismatched runtime linked successfully |
| drop-retain | True | 1 | 19.738 | mismatched runtime linked successfully |
| header-forces-runtime-features | True | 1 | 36.791 | mismatched runtime linked successfully |
| split-runtime-default-flags | True | 1 | 28.257 | undefined reference |
| lint-profile-default-runtime | True | 1 | 14.421 | undefined reference |
| parallel-undefined-winner-lost | True | 1 | 52.427 | signal: aborted |
| error-ancestry-missing-independent | True | 1 | 21.723 | ancestry: |
| primitive-missing-nominal-typeerror | True | 1 | 44.398 | C differs from Node |
| spread-undefined-pending-lost | True | 1 | 55.325 | spread: |
| drop-fresh-nullish-proof | True | 1 | 37.049 | Lower refused acceptance row |
| nullish-guard-claims-other-narrowings | True | 1 | 25.308 | nullish guard restored an unsupported narrowing |
| builtin-error-write-unrecorded | False | 1 | 24.678 | a write lowering didn't record |
| builtin-error-write-unrecorded-replay | True | 1 | 10.873 | a write lowering didn't record |
| fuzzer-reuses-featureless-runtime | True | 1 | 10.334 | program runtime features did not agree |
| unconditional-wider-storage-read | True | 1 | 13.214 | Lower refused acceptance row |

Unreached mutants: none.

Classification caveats: error-ancestry-missing originally reached LeakSanitizer (144 bytes/two objects) before stdout, so its original false marker remains. A separate literal-prefix ancestry witness isolates the nominal assertion. implicit-string-conversion failed the actual stdout/exit assertion (source 1, native 70 with owned panic), despite the stale expected log marker. native-scalar-properties failed native stderr agreement (exit 70, unsupported runtime value panic), while its second numeric-scalar subcase passed; its stale marker false is preserved. See raw logs and independent verification records. Build/tool failures do not prove a semantic mutant.

## Runtime clearance: every changed C and header line

This complete base-to-delivery patch includes all additions and removals, including resolution changes beyond the original source. Source-only runtime patch is source-runtime-lines.patch for comparison. Main consumer changes beyond the source occur in json_stringify.c, object.c and parallel.c; OrdinaryToPrimitive adaptation and TLS/pending declarations occur in exceptions.c/adamic.h. No safe-list declaration is used to bypass locking.

```diff
diff --git a/internal/native/runtime/adamic.h b/internal/native/runtime/adamic.h
index 26d417dc..9f201003 100644
--- a/internal/native/runtime/adamic.h
+++ b/internal/native/runtime/adamic.h
@@ -12,6 +12,42 @@
 
 #include <math.h>
 
+// A unit must link to the runtime built with these same layout and callback features.
+// Definedness matches the #ifdef decisions below, including a macro defined as 0.
+#ifdef ADAMIC_CLOSURE_CONVENTION
+#define ADAMIC_FEATURE_CLOSURE_CONVENTION _closure_convention
+#else
+#define ADAMIC_FEATURE_CLOSURE_CONVENTION
+#endif
+#ifdef ADAMIC_CANONICAL_CLOSURES
+#define ADAMIC_FEATURE_CANONICAL_CLOSURES _canonical_closures
+#else
+#define ADAMIC_FEATURE_CANONICAL_CLOSURES
+#endif
+#ifdef ADAMIC_CLOSURE_RECEIVERS
+#define ADAMIC_FEATURE_CLOSURE_RECEIVERS _closure_receivers
+#else
+#define ADAMIC_FEATURE_CLOSURE_RECEIVERS
+#endif
+#ifdef ADAMIC_REGEXP_REPLACE_CALLBACK
+#define ADAMIC_FEATURE_REGEXP_REPLACE_CALLBACK _regexp_replace_callback
+#else
+#define ADAMIC_FEATURE_REGEXP_REPLACE_CALLBACK
+#endif
+#ifdef ADAMIC_NODE_HOST
+#define ADAMIC_FEATURE_NODE_HOST _node_host
+#else
+#define ADAMIC_FEATURE_NODE_HOST
+#endif
+
+#define ADAMIC_FEATURE_NAME_PARTS(a, b, c, d, e) adamic_runtime_features##a##b##c##d##e
+#define ADAMIC_FEATURE_NAME(a, b, c, d, e) ADAMIC_FEATURE_NAME_PARTS(a, b, c, d, e)
+#define ADAMIC_RUNTIME_FEATURES ADAMIC_FEATURE_NAME(ADAMIC_FEATURE_CLOSURE_CONVENTION, ADAMIC_FEATURE_CANONICAL_CLOSURES, ADAMIC_FEATURE_CLOSURE_RECEIVERS, ADAMIC_FEATURE_REGEXP_REPLACE_CALLBACK, ADAMIC_FEATURE_NODE_HOST)
+extern const unsigned char ADAMIC_RUNTIME_FEATURES;
+// used keeps the relocation through optimization; retain keeps it through linker section removal.
+// This is a link requirement only: it allocates nothing and executes no runtime check.
+static const void *const adamic_runtime_feature_reference __attribute__((used, retain)) = &ADAMIC_RUNTIME_FEATURES;
+
 enum adamic_stream {
 	adamic_stdout = 1,
 	adamic_stderr = 2,
@@ -1038,10 +1074,20 @@ typedef struct adamic_primitive_method {
 typedef adamic_primitive_method (*adamic_primitive_get)(adamic_heap *receiver, const char *name);
 adamic_primitive adamic_ordinary_to_primitive(adamic_heap *receiver, enum adamic_primitive_hint hint, adamic_primitive_get get);
 
-// adamic_thrown is the error being thrown, or NULL (exceptions.c): set by a throw, tested after
-// every call that can throw, and taken by the catch that lands it. adamic_error_new is new
-// Error(message), and adamic_uncaught the panic of an error nothing caught.
-extern _Thread_local adamic_object *adamic_thrown;
+// adamic_thrown owns a tagged union payload (NULL is undefined). The separate pending flag
+// distinguishes a thrown undefined from no exception. Every throwing call tests the flag.
+// A catch takes ownership; adamic_uncaught flushes stdout, writes one diagnostic line, releases the payload and exits 1.
+extern _Thread_local adamic_heap *adamic_thrown;
+extern _Thread_local bool adamic_exception_pending;
+extern const adamic_class adamic_error_class;
+extern const adamic_class adamic_host_error_class;
+extern const adamic_class adamic_host_type_error_class;
+extern const adamic_class adamic_host_range_error_class;
+extern const adamic_class adamic_range_error_class;
+extern const adamic_class adamic_type_error_class;
+extern const adamic_class adamic_reference_error_class;
+adamic_object *adamic_error_new_kind(adamic_string *message, const char *kind);
+adamic_object *adamic_error_new_code(adamic_string *message, adamic_string *code);
 // NULL is an omitted or undefined message; its empty fallback is not an own field.
 adamic_object *adamic_error_new(adamic_string *message);
 const char *adamic_error_field_name(const adamic_object *object, const char *name);
diff --git a/internal/native/runtime/class_inheritance.c b/internal/native/runtime/class_inheritance.c
index 809bba97..bac1cc80 100644
--- a/internal/native/runtime/class_inheritance.c
+++ b/internal/native/runtime/class_inheritance.c
@@ -3,7 +3,7 @@
 
 bool adamic_instanceof(const void *value, const adamic_class *wanted) {
 	const adamic_heap *heap = value;
-	if (heap == NULL || heap->kind != adamic_kind_object) { return false; }
+	if (heap == NULL || heap == &adamic_null || heap->kind != adamic_kind_object) { return false; }
 	const adamic_object *object = value;
 	for (const adamic_class *class = object->class; class != NULL; class = class->base) {
 		if (class == wanted || (wanted->definition != 0 && class->definition == wanted->definition)) { return true; }
diff --git a/internal/native/runtime/exceptions.c b/internal/native/runtime/exceptions.c
index 69a8a0e4..d8b7861e 100644
--- a/internal/native/runtime/exceptions.c
+++ b/internal/native/runtime/exceptions.c
@@ -8,7 +8,8 @@
 #include <stdlib.h>
 #include <string.h>
 
-_Thread_local adamic_object *adamic_thrown;
+_Thread_local adamic_heap *adamic_thrown;
+_Thread_local bool adamic_exception_pending;
 
 static const char *const error_names[] = {"name", "message"};
 static const bool error_references[] = {true, true};
@@ -16,23 +17,53 @@ static const adamic_shape error_shape = {2, error_names, error_references, NULL}
 static const char *const absent_message_names[] = {"name", "#message"};
 static const adamic_shape absent_message_shape = {2, absent_message_names, error_references, NULL};
 static adamic_string empty_message = ADAMIC_STRING("");
+const adamic_class adamic_error_class = {NULL, 0, 2, NULL, 0, &error_shape, NULL, 0, false, 0, NULL};
+const adamic_class adamic_host_error_class = {&adamic_error_class, 2, 3, NULL, 0, NULL, NULL, 0, false, 0, NULL};
 static adamic_string error_name = ADAMIC_STRING("Error");
+static adamic_string range_error_name = ADAMIC_STRING("RangeError");
+static adamic_string type_error_name = ADAMIC_STRING("TypeError");
+static adamic_string reference_error_name = ADAMIC_STRING("ReferenceError");
+const adamic_class adamic_reference_error_class = {&adamic_error_class, 2, 2, NULL, 0, &error_shape, NULL, 0, false, 0, NULL};
+const adamic_class adamic_range_error_class = {&adamic_error_class, 2, 2, NULL, 0, &error_shape, NULL, 0, false, 0, NULL};
+const adamic_class adamic_type_error_class = {&adamic_error_class, 2, 2, NULL, 0, &error_shape, NULL, 0, false, 0, NULL};
+const adamic_class adamic_host_type_error_class = {&adamic_type_error_class, 2, 3, NULL, 0, NULL, NULL, 0, false, 0, NULL};
+const adamic_class adamic_host_range_error_class = {&adamic_range_error_class, 2, 3, NULL, 0, NULL, NULL, 0, false, 0, NULL};
 
-adamic_object *adamic_error_new(adamic_string *message) {
-	// Undefined creates no own message property. Keep the inherited empty fallback
-	// in private storage so ordinary and checked field reads share the same slot.
+adamic_object *adamic_error_new_kind(adamic_string *message, const char *kind) {
 	adamic_object *error = adamic_object_new(message == NULL ? &absent_message_shape : &error_shape);
-	error->slots[0].reference = adamic_retain(&error_name);
+	error->class = &adamic_error_class;
+	adamic_string *name = &error_name;
+	if (strcmp(kind, "RangeError") == 0) { error->class = &adamic_range_error_class; name = &range_error_name; }
+	if (strcmp(kind, "TypeError") == 0) { error->class = &adamic_type_error_class; name = &type_error_name; }
+	if (strcmp(kind, "ReferenceError") == 0) { error->class = &adamic_reference_error_class; name = &reference_error_name; }
+	error->slots[0].reference = adamic_retain(name);
 	error->slots[1].reference = adamic_retain(message == NULL ? &empty_message : message);
 	adamic_object_field_types(error)[0] = adamic_rep_string;
 	adamic_object_field_types(error)[1] = adamic_rep_string;
 	return error;
 }
 
+adamic_object *adamic_error_new_code(adamic_string *message, adamic_string *code) {
+	static const char *const names[] = {"name", "message", "code"};
+	static const bool references[] = {true, true, true};
+	static const adamic_shape shape = {3, names, references, NULL};
+	adamic_object *error = adamic_object_new(&shape);
+	error->class = &adamic_host_error_class;
+	error->slots[0].reference = adamic_retain(&error_name);
+	error->slots[1].reference = adamic_retain(message);
+	error->slots[2].reference = adamic_retain(code);
+	adamic_object_field_types(error)[0] = adamic_rep_string;
+	adamic_object_field_types(error)[1] = adamic_rep_string;
+	adamic_object_field_types(error)[2] = adamic_rep_string;
+	return error;
+}
+
+adamic_object *adamic_error_new(adamic_string *message) { return adamic_error_new_kind(message,"Error"); }
+
 // Get must stay inside the loop: the first method or getter can change the second.
 adamic_primitive adamic_ordinary_to_primitive(adamic_heap *receiver, enum adamic_primitive_hint hint, adamic_primitive_get get) {
 	adamic_primitive empty = {adamic_primitive_undefined, {.reference = NULL}};
-	if (adamic_thrown != NULL) { return empty; }
+	if (adamic_exception_pending) { return empty; }
 	if (hint != adamic_hint_default && hint != adamic_hint_number && hint != adamic_hint_string) {
 		static const char message[] = "compiler bug: invalid primitive hint";
 		adamic_panic(message, sizeof message - 1);
@@ -41,11 +72,11 @@ adamic_primitive adamic_ordinary_to_primitive(adamic_heap *receiver, enum adamic
 	if (hint == adamic_hint_string) { names[0] = "toString"; names[1] = "valueOf"; }
 	for (size_t i = 0; i < 2; i++) {
 		adamic_primitive_method method = get(receiver, names[i]);
-		if (adamic_thrown != NULL) { adamic_release(method.owner); return empty; }
+		if (adamic_exception_pending) { adamic_release(method.owner); return empty; }
 		if (method.call == NULL) { adamic_release(method.owner); continue; }
 		adamic_primitive result = method.call(receiver, method.owner);
 		adamic_release(method.owner);
-		if (adamic_thrown != NULL) { return empty; }
+		if (adamic_exception_pending) { return empty; }
 		if (result.kind != adamic_primitive_object) { return result; }
 		adamic_release(result.value.reference);
 	}
@@ -55,13 +86,14 @@ adamic_primitive adamic_ordinary_to_primitive(adamic_heap *receiver, enum adamic
 	static const char message[] = "Cannot convert object to primitive value";
 	adamic_string *text = adamic_string_allocate(sizeof message - 1);
 	memcpy((char *)text->bytes, message, sizeof message - 1);
-	adamic_object *error = adamic_error_new(text);
+	adamic_object *error = adamic_error_new_kind(text, "TypeError");
 	adamic_release(text);
 	adamic_string *type_name = adamic_string_allocate(sizeof name - 1);
 	memcpy((char *)type_name->bytes, name, sizeof name - 1);
 	adamic_release(error->slots[0].reference);
 	error->slots[0].reference = type_name;
-	adamic_thrown = error;
+	adamic_thrown = (adamic_heap *)error;
+	adamic_exception_pending = true;
 	return empty;
 }
 
@@ -70,22 +102,39 @@ const char *adamic_error_field_name(const adamic_object *object, const char *nam
 	return object->shape == &absent_message_shape && strcmp(name, "message") == 0 ? "#message" : name;
 }
 
-_Noreturn void adamic_uncaught(void) {
-	// String(error), as Node's runner reports an error nothing caught: name, and ": " and the message
-	// when there is one.
-	static adamic_slot_cache name_cache, message_cache;
-	const adamic_string *name = adamic_object_field(adamic_thrown, "name", &name_cache)->reference;
-	const adamic_string *message = adamic_object_field(adamic_thrown, "message", &message_cache)->reference;
-	size_t length = name->length + (message->length > 0 ? 2 + message->length : 0);
-	char *text = malloc(length + 1);
-	if (text == NULL) {
-		static const char out_of_memory[] = "out of memory";
-		adamic_panic(out_of_memory, sizeof out_of_memory - 1);
+// Write borrowed string bytes without conversion or allocation. Match Node's UTF-8
+// encoding of lone surrogates and escape CR/LF to keep exactly one diagnostic line.
+static void uncaught_text(const adamic_string *text) {
+	for (size_t index = 0; index < text->length; index++) {
+		unsigned char byte = (unsigned char)text->bytes[index];
+		if (byte == '\r') { fputs("\\r", stderr); }
+		else if (byte == '\n') { fputs("\\n", stderr); }
+		else if (byte == 0xed && index + 2 < text->length && (unsigned char)text->bytes[index + 1] >= 0xa0) {
+			fwrite("\xef\xbf\xbd", 1, 3, stderr);
+			index += 2;
+		} else { fputc(byte, stderr); }
 	}
-	memcpy(text, name->bytes, name->length);
-	if (message->length > 0) {
-		memcpy(text + name->length, ": ", 2);
-		memcpy(text + name->length + 2, message->bytes, message->length);
+}
+
+_Noreturn void adamic_uncaught(void) {
+	adamic_output_flush();
+	fputs("Uncaught ", stderr);
+	if (adamic_instanceof(adamic_thrown, &adamic_error_class) && !((const adamic_object *)adamic_thrown)->class->is_static) {
+		const adamic_object *error = (const adamic_object *)adamic_thrown;
+		adamic_slot_cache name_cache = {0};
+		adamic_slot_cache message_cache = {0};
+		uncaught_text(adamic_object_field(error, "name", &name_cache)->reference);
+		fputs(": ", stderr);
+		uncaught_text(adamic_object_field(error, "message", &message_cache)->reference);
+	} else if (adamic_thrown == &adamic_null) {
+		fputs("null", stderr);
+	} else {
+		uncaught_text(adamic_union_typeof(adamic_thrown, false));
 	}
-	adamic_panic(text, length);
+	fputc('\n', stderr);
+	fflush(stderr);
+	adamic_release(adamic_thrown);
+	adamic_thrown = NULL;
+	adamic_exception_pending = false;
+	exit(1);
 }
diff --git a/internal/native/runtime/features.c b/internal/native/runtime/features.c
new file mode 100644
index 00000000..02a355ec
--- /dev/null
+++ b/internal/native/runtime/features.c
@@ -0,0 +1,4 @@
+// Each runtime defines only its own feature set. adamic.h references it from every unit.
+#include "adamic.h"
+
+const unsigned char ADAMIC_RUNTIME_FEATURES = 0;
diff --git a/internal/native/runtime/json_stringify.c b/internal/native/runtime/json_stringify.c
index f1b292a7..87c6f7a5 100644
--- a/internal/native/runtime/json_stringify.c
+++ b/internal/native/runtime/json_stringify.c
@@ -169,13 +169,13 @@ static adamic_json_result prepare(json_writer *w, adamic_value value, const adam
     return result;
 }
 static void write_field(json_writer *w, const adamic_object *object, const adamic_json_field *field, size_t slots, size_t depth, size_t *written) {
-    if (adamic_thrown != NULL) { return; }
+    if (adamic_exception_pending) { return; }
     if (field->schema == NULL || field->name == NULL || field->slot >= slots) { refuse("JSON object lacks complete field descriptors"); }
     bool owned;
     adamic_json_result result = prepare(w, object->slots[field->slot], field->schema, field->name, &owned);
-    if (adamic_thrown != NULL) { if (owned && reference_result(result.schema)) { adamic_release(result.value.reference); } return; }
+    if (adamic_exception_pending) { if (owned && reference_result(result.schema)) { adamic_release(result.value.reference); } return; }
     json_scalar s = scalar(result.value, result.schema, w->runtime);
-    if (adamic_thrown == NULL && s.kind != adamic_json_undefined && s.kind != adamic_json_function) {
+    if (!adamic_exception_pending && s.kind != adamic_json_undefined && s.kind != adamic_json_function) {
         if ((*written)++ != 0) { ascii(w, ","); }
         indent(w, depth + 1);
         quote(w, field->name);
@@ -186,7 +186,7 @@ static void write_field(json_writer *w, const adamic_object *object, const adami
 }
 static bool write_prepared_body(json_writer *w, adamic_value value, const adamic_json_schema *schema, size_t depth);
 static bool write_prepared(json_writer *w, adamic_value value, const adamic_json_schema *schema, size_t depth) {
-    if (adamic_thrown != NULL) { return false; }
+    if (adamic_exception_pending) { return false; }
     json_scalar s = scalar(value, schema, w->runtime);
     bool container = s.kind == adamic_json_array || s.kind == adamic_json_tuple || s.kind == adamic_json_object || s.kind == adamic_json_toJSON;
     json_ancestor current = {s.value.reference, w->ancestors};
@@ -201,7 +201,7 @@ static bool write_prepared(json_writer *w, adamic_value value, const adamic_json
     return present;
 }
 static bool write_prepared_body(json_writer *w, adamic_value value, const adamic_json_schema *schema, size_t depth) {
-	if (adamic_thrown != NULL) { return false; }
+	if (adamic_exception_pending) { return false; }
 	ADAMIC_CHECK_STACK();
 	json_scalar s = scalar(value, schema, w->runtime);
 	schema = s.schema;
@@ -232,8 +232,8 @@ static bool write_prepared_body(json_writer *w, adamic_value value, const adamic
             adamic_string *key = adamic_string_from_number((double)index);
             bool present = write_value(w, item.value, item.schema, depth + 1, key);
             adamic_release(key);
-            if (!present && adamic_thrown == NULL) { ascii(w, "null"); }
-            if (adamic_thrown != NULL) { return false; }
+            if (!present && !adamic_exception_pending) { ascii(w, "null"); }
+            if (adamic_exception_pending) { return false; }
 		}
 		if (count != 0) { indent(w, depth); }
 		ascii(w, "]");
@@ -270,12 +270,12 @@ static bool write_prepared_body(json_writer *w, adamic_value value, const adamic
 	}
 }
 static bool write_value(json_writer *w, adamic_value value, const adamic_json_schema *schema, size_t depth, const adamic_string *key) {
-    if (adamic_thrown != NULL) { return false; }
+    if (adamic_exception_pending) { return false; }
     bool owned;
     adamic_json_result result = prepare(w, value, schema, key, &owned);
     bool present = write_prepared(w, result.value, result.schema, depth);
     if (owned && reference_result(result.schema)) { adamic_release(result.value.reference); }
-    return present && adamic_thrown == NULL;
+    return present && !adamic_exception_pending;
 }
 static void keys(json_writer *w, adamic_value value, const adamic_json_schema *schema) {
 	if (schema == NULL || (schema->kind != adamic_json_array && schema->kind != adamic_json_tuple) || value.reference == NULL) { return; }
diff --git a/internal/native/runtime/node_crypto.c b/internal/native/runtime/node_crypto.c
index b9af6161..a6f8be4b 100644
--- a/internal/native/runtime/node_crypto.c
+++ b/internal/native/runtime/node_crypto.c
@@ -67,18 +67,20 @@ adamic_object *adamic_node_hash_new(void)
 	hash->slots[0].reference = adamic_array_new(0, false);
 	return hash;
 }
-static void check_finalized(const adamic_object *hash)
+static bool check_finalized(const adamic_object *hash)
 {
-	if (hash->slots[1].number != 0) {
-		static const char message[] =
-			"Error [ERR_CRYPTO_HASH_FINALIZED]: Digest already called";
-		adamic_panic(message, sizeof message - 1);
-	}
+ if (hash->slots[1].number == 0) { return true; }
+ static adamic_string message = ADAMIC_STRING("Digest already called");
+ static adamic_string code = ADAMIC_STRING("ERR_CRYPTO_HASH_FINALIZED");
+ adamic_object *error = adamic_error_new_code(&message, &code);
+ adamic_thrown = &error->heap;
+ adamic_exception_pending = true;
+ return false;
 }
 adamic_object *adamic_node_hash_update(adamic_object *hash,
 									   const adamic_string *text, int encoding)
 {
-	check_finalized(hash);
+	if (!check_finalized(hash)) { return NULL; }
 	adamic_array *input = adamic_node_buffer_from(text, encoding);
 	adamic_array *bytes = hash->slots[0].reference;
 	for (size_t i = 0; i < input->length; i++) {
@@ -89,7 +91,7 @@ adamic_object *adamic_node_hash_update(adamic_object *hash,
 }
 adamic_string *adamic_node_hash_digest(adamic_object *hash)
 {
-	check_finalized(hash);
+	if (!check_finalized(hash)) { return NULL; }
 	hash->slots[1].number = 1;
 	const adamic_array *bytes = hash->slots[0].reference;
 	uint32_t state[8] = {0x6a09e667, 0xbb67ae85, 0x3c6ef372, 0xa54ff53a,
diff --git a/internal/native/runtime/node_fs_file.c b/internal/native/runtime/node_fs_file.c
index 517bbe06..9b134b58 100644
--- a/internal/native/runtime/node_fs_file.c
+++ b/internal/native/runtime/node_fs_file.c
@@ -27,7 +27,10 @@ static void raise_error(const char *name, const char *code, const char *message)
     error->slots[0].reference = text(name);
     error->slots[1].reference = text(message);
     error->slots[2].reference = text(code);
-    adamic_thrown = error;
+    error->class = strcmp(name, "TypeError") == 0 ? &adamic_host_type_error_class :
+        strcmp(name, "RangeError") == 0 ? &adamic_host_range_error_class : &adamic_host_error_class;
+    adamic_thrown = &error->heap;
+    adamic_exception_pending = true;
 }
 
 static void system_error(int error, const char *operation, const char *path) {
@@ -339,7 +342,7 @@ static double write_data(int descriptor, const char *buffer, size_t length, bool
     // finally closes an internally opened fd, even after write/fsync failed.
     // A close failure overrides an earlier exception as Node's finally does.
     if (owned && close(descriptor) < 0 && errno != EINTR) {
-        if (adamic_thrown != NULL) { adamic_release(adamic_thrown); adamic_thrown = NULL; }
+        if (adamic_exception_pending) { adamic_release(adamic_thrown); adamic_thrown = NULL; adamic_exception_pending = false; }
         system_error(errno, "close", NULL);
     }
     return 0;
@@ -574,7 +577,7 @@ static void remove_path(const char *name, bool recursive, bool force) {
     DIR *directory = opendir(name);
     if (directory == NULL) { if (errno != ENOENT) { remove_error(errno, "scandir", name); } return; }
     struct dirent *entry;
-    while (adamic_thrown == NULL) {
+    while (!adamic_exception_pending) {
         errno = 0; entry = readdir(directory);
         if (entry == NULL) { if (errno != 0) { remove_error(errno, "scandir", name); } break; }
         if (strcmp(entry->d_name, ".") == 0 || strcmp(entry->d_name, "..") == 0) { continue; }
@@ -585,7 +588,7 @@ static void remove_path(const char *name, bool recursive, bool force) {
         remove_path(child, true, true); free(child);
     }
     closedir(directory);
-    if (adamic_thrown == NULL && rmdir(name) != 0 && errno != ENOENT) { remove_error(errno, "rmdir", name); }
+    if (!adamic_exception_pending && rmdir(name) != 0 && errno != ENOENT) { remove_error(errno, "rmdir", name); }
 }
 
 double adamic_fs_file_rm(const adamic_string *path, bool recursive, bool force) {
diff --git a/internal/native/runtime/node_host.c b/internal/native/runtime/node_host.c
index 2a924806..c27bff49 100644
--- a/internal/native/runtime/node_host.c
+++ b/internal/native/runtime/node_host.c
@@ -146,10 +146,13 @@ static void host_directory_error(int error, const char *operation, const adamic_
         adamic_release(message);
         message = closed;
     }
-    adamic_thrown = adamic_object_new(&shape);
-    adamic_thrown->slots[0].reference = &error_name;
-    adamic_thrown->slots[1].reference = message;
-    adamic_thrown->slots[2].reference = prefix;
+    adamic_object *raised = adamic_object_new(&shape);
+    raised->class = &adamic_host_error_class;
+    adamic_thrown = &raised->heap;
+    adamic_exception_pending = true;
+    raised->slots[0].reference = &error_name;
+    raised->slots[1].reference = message;
+    raised->slots[2].reference = prefix;
     adamic_release(reason);
     adamic_release(syscall);
 }
@@ -200,7 +203,7 @@ double adamic_fs_file_host_chdir(const adamic_string *directory) {
     adamic_string *target = adamic_decode_utf8((const unsigned char *)path, strlen(path));
     free(path);
     adamic_string *from = adamic_fs_file_host_cwd();
-    if (adamic_thrown != NULL) { adamic_release(target); return 0; }
+    if (adamic_exception_pending) { adamic_release(target); return 0; }
     host_directory_error(error, "chdir", from, target);
     adamic_release(target);
     adamic_release(from);
diff --git a/internal/native/runtime/object.c b/internal/native/runtime/object.c
index 6cf9e0bb..8aa11879 100644
--- a/internal/native/runtime/object.c
+++ b/internal/native/runtime/object.c
@@ -57,7 +57,7 @@ adamic_object *adamic_object_copy_checked(const adamic_object *source, const cha
 			adamic_object_field_types(object)[index] = adamic_object_field_types(owner)[adamic_slot_index(owner, found)];
 		} else {
 			object->slots[index] = adamic_accessor_get((adamic_object *)source, shape->names[index]);
-			if (adamic_thrown != NULL) {
+			if (adamic_exception_pending) {
 				adamic_release(object);
 				return NULL;
 			}
diff --git a/internal/native/runtime/parallel.c b/internal/native/runtime/parallel.c
index bb5b30dc..810d1e03 100644
--- a/internal/native/runtime/parallel.c
+++ b/internal/native/runtime/parallel.c
@@ -35,7 +35,7 @@ typedef struct scope {
 	size_t completed;
 	size_t grain;
 	size_t exception_index;
-	adamic_object *exception;
+	adamic_heap *exception;
 } scope;
 
 typedef struct range {
@@ -190,19 +190,22 @@ static bool claim(scope **scope_out, size_t *from_out, size_t *end_out) {
 }
 
 static void execute_range(scope *scope, size_t from, size_t end) {
-	adamic_object *outer_exception = adamic_thrown;
+	adamic_heap *outer_exception = adamic_thrown;
+	bool outer_pending = adamic_exception_pending;
 	adamic_thrown = NULL;
+	adamic_exception_pending = false;
 	for (size_t index = from; index < end; index++) {
 		adamic_value arguments[] = {scope->items->elements[index], {.number = (double)index}};
 		adamic_value result = adamic_closure_call(scope->work, arguments, 2);
-		if (adamic_thrown != NULL) {
-			adamic_object *error = adamic_thrown;
+		if (adamic_exception_pending) {
+			adamic_heap *error = adamic_thrown;
 			adamic_thrown = NULL;
+			adamic_exception_pending = false;
 			if (scope->results->references) { adamic_release(result.reference); }
 			adamic_share(error);
 			pthread_mutex_lock(&scheduler);
 			if (index < scope->exception_index) {
-				adamic_object *previous = scope->exception;
+				adamic_heap *previous = scope->exception;
 				scope->exception = error;
 				scope->exception_index = index;
 				error = previous;
@@ -215,6 +218,7 @@ static void execute_range(scope *scope, size_t from, size_t end) {
 		}
 	}
 	adamic_thrown = outer_exception;
+	adamic_exception_pending = outer_pending;
 	pthread_mutex_lock(&scheduler);
 	scope->completed += end - from;
 	pthread_cond_broadcast(&changed);
@@ -298,7 +302,7 @@ static adamic_array *parallel_map(adamic_array *items, adamic_closure *work, boo
 		for (size_t index = 0; index < items->length; index++) {
 			adamic_value arguments[] = {items->elements[index], {.number = (double)index}};
 			adamic_value result = adamic_closure_call(work, arguments, 2);
-			if (adamic_thrown != NULL) {
+			if (adamic_exception_pending) {
 				if (results->references) { adamic_release(result.reference); }
 				adamic_release(results); return NULL;
 			}
@@ -324,9 +328,10 @@ static adamic_array *parallel_map(adamic_array *items, adamic_closure *work, boo
 		pthread_mutex_lock(&scheduler);
 	}
 	pthread_mutex_unlock(&scheduler);
-	if (state.exception != NULL) {
+	if (state.exception_index != SIZE_MAX) {
 		adamic_release(results);
 		adamic_thrown = state.exception;
+		adamic_exception_pending = true;
 		return NULL;
 	}
 	return results;
diff --git a/internal/native/runtime/regexp_replace.c b/internal/native/runtime/regexp_replace.c
index fee15c79..5e613f5d 100644
--- a/internal/native/runtime/regexp_replace.c
+++ b/internal/native/runtime/regexp_replace.c
@@ -116,7 +116,7 @@ adamic_string *adamic_regex_replace_callback(adamic_string *input, adamic_object
 			if (parameters[j] != 2)
 				adamic_release(packed[j].reference);
 		adamic_release(rest);
-		if (adamic_thrown != NULL)
+		if (adamic_exception_pending)
 			break;
 		adamic_string *replacement = NULL;
 		switch (return_kind) {
@@ -151,7 +151,7 @@ adamic_string *adamic_regex_replace_callback(adamic_string *input, adamic_object
 	free(packed);
 	adamic_release(matches);
 	adamic_string *result = NULL;
-	if (adamic_thrown == NULL) {
+	if (!adamic_exception_pending) {
 		adamic_array_push(pieces, (adamic_value){.reference = adamic_string_slice(
 													 input, (double)previous,
 													 adamic_string_length(input), true)});
diff --git a/internal/native/runtime/sort.c b/internal/native/runtime/sort.c
index 450b9b7e..0ff7b388 100644
--- a/internal/native/runtime/sort.c
+++ b/internal/native/runtime/sort.c
@@ -41,7 +41,7 @@ typedef struct {
 
 static int order(sort_state *state, adamic_value left, adamic_value right) {
 	int ordered = state->compare(left, right, state->context);
-	if (adamic_thrown != NULL) {
+	if (adamic_exception_pending) {
 		// Only the sort's own frames lie between here and stop, none holding a reference.
 #ifdef ADAMIC_TARGET_WASI
 		return 0;
@@ -55,7 +55,7 @@ static int order(sort_state *state, adamic_value left, adamic_value right) {
 // WASI has no longjmp without experimental exception handling. Propagate the pending
 // error through each sort frame before another comparison or index uses its result.
 #ifdef ADAMIC_TARGET_WASI
-#define STOP_SORT(value) do { if (adamic_thrown != NULL) { return value; } } while (0)
+#define STOP_SORT(value) do { if (adamic_exception_pending) { return value; } } while (0)
 #else
 #define STOP_SORT(value) ((void)0)
 #endif
diff --git a/internal/native/runtime/sort_undefined.c b/internal/native/runtime/sort_undefined.c
index a0cdfe07..08960939 100644
--- a/internal/native/runtime/sort_undefined.c
+++ b/internal/native/runtime/sort_undefined.c
@@ -14,7 +14,7 @@ void adamic_array_sort_undefined_last(adamic_array *array, int (*compare)(adamic
 	}
 	// The comparator may change the array it sorts; what it sorts is this copy, which nothing else sees.
 	adamic_array_sort(defined, compare, context);
-	if (adamic_thrown != NULL) {
+	if (adamic_exception_pending) {
 		// A comparator threw: nothing is written back, as V8 writes nothing back then.
 		adamic_release(defined);
 		return;
diff --git a/internal/native/runtime/union.c b/internal/native/runtime/union.c
index b2fc483e..8911d224 100644
--- a/internal/native/runtime/union.c
+++ b/internal/native/runtime/union.c
@@ -203,5 +203,11 @@ adamic_heap *adamic_dynamic_property(adamic_heap *value, const char *name) {
 		if (array->properties != NULL) return adamic_dynamic_property(&array->properties->heap, name);
 		return NULL;
 	}
+	// A primitive can replace a narrowed object. These data reads follow JavaScript's
+	// temporary boxing; prototype methods remain outside the dynamic-read contract.
+	if (value->kind == adamic_kind_string) {
+		return strcmp(name, "length") == 0 ? adamic_box_number(adamic_string_length((const adamic_string *)value)) : NULL;
+	}
+	if (value->kind == adamic_kind_number || value->kind == adamic_kind_boolean) return NULL;
 	adamic_panic("dynamic property read on an unsupported runtime value", sizeof "dynamic property read on an unsupported runtime value" - 1);
 }

```

## Integration decisions and final direct checks

Main newly admitted async rejection witnesses and terminal uninitialized-representation checks keep their existing exit-70 contract. They are explicitly source-hash pinned below; changed programs cannot borrow that contract. Synchronous language TDZ failures use exit 1 and exact Uncaught diagnostics. This cut does not claim step 21 support for main’s restricted async rejection representation.

- `internal/oracle/testdata/review/agree/fxspptb_047cb0d_n_element_plain.a`: `1eb94d37614330494f21b0d61e405183d067c9de231941e780b7356f1816c25c` (main terminal representation check).
- `internal/oracle/testdata/review/agree/fxspptb_047cb0d_n_element.a`: `37e4e5c76731a999d9153d9435be8299a0486a45d2c173fe306486085a1d60a8` (main terminal representation check).
- `internal/oracle/testdata/review/agree/fxspptb_047cb0d_n_element_method.a`: `6f32b516bef7bd318bebb855eff6a8186e16f71a8d96acdd421831499006588a` (main terminal representation check).
- `internal/oracle/testdata/async_coverage_reject_nested.a`: `10fa813d2cb2ec049cf3b55944492e3a2db9e890a866486b9b3a5e3cdaea4262` (main terminal native check).
- `internal/oracle/testdata/namespaces_unknown_write_before.a`: `32d4a539c31220adb40189c4a4e1cb0d0f4e4eff5228539f7567c72e03538a52` (main terminal native check).
- `internal/oracle/testdata/element_access_presence.a`: `3056e90354f2749fb53b5fb5ea34c701b1bb4f524b8e677bf6f6005cf00f94f3` (main terminal native check).
- `internal/oracle/testdata/namespaces_unknown_void_before.a`: `79a5b5970dfaf5e59ab89c34066973391d0725625fa61bc12c577b6fe2106d26` (main terminal native check).
- `internal/oracle/testdata/async_coverage_reject_eager.a`: `bc4f5ce41f6731e5e3ffa96ed16047bf22c7314b3ee5119e4c989067a1a3f111` (main terminal native check).
- `internal/oracle/testdata/async_coverage_reject_empty.a`: `7f68067134db38d0405c91ee6fcc620d5eccdce69d786efe76aeb3a64054e9ad` (main terminal native check).
- `internal/oracle/testdata/namespaces_unknown_enum_before.a`: `32388d9f7258d18f4b557bf6ae69fffc4f20a17b8b5c124c7e8dd089666a0892` (main terminal native check).
- `internal/oracle/testdata/namespaces_unknown_before.a`: `4f1be7b872b6d8c72ab38b60a6241ed370a09e77a02fb486b37e0603b657653a` (main terminal native check).
- `internal/oracle/testdata/async_throw.a`: `6c3e00e7233b1131d43ca9e746748fc2b9f035daf6560a37555d03f81c8386fd` (main terminal native check).
- `internal/oracle/testdata/namespaces_unknown_function_before.a`: `78153a5b9e1f38ef774091542042ba449691052c55984e0959ac066e47394b3d` (main terminal native check).

A fresh strict nullish comparison now establishes only the remaining declared union members, branch-sensitively. It does not restore a stale nominal or other scalar contract; positive Node/native and negative refusal regressions have independent mutants. Generated builtin Error initializers record both field write sites for main’s freshness proof. The fuzz feature test preserves main’s matching-archive selection and has a featureless-archive selection mutant. The enum cleanup mutant removes both mutually exclusive normal and exceptional cleanup sites and still requires an actual leak with unchanged Node stdout.

Review programs admitted by catchable initialization moved from refused to agree; five writing-call refusals moved to the existing refused category. Smoke retains narrowed_writes in its exact refusal lane with a full registered subtest path. No guard was relaxed.

Final a-check: 41 changed .a paths, 24 checked, 16 expected refusals and one TS18046 checker error. The branch compiler was rebuilt after the proof repair. No .a body was rewritten to make the checker accept it.

The report distinguishes failures from corrected reruns. A cache eviction during parallel validation invalidated a deferred subpackage import; this was my execution mistake. Later checks run serially without cache eviction. IR GoExports needed more scratch capacity: the obsolete product was retired, current exports protected, the task SDK temporarily archived and restored, and actual owner recipes built one at a time. Only completed byte-identical archives were deduplicated for the successful owner runs. The unsuccessful in-progress dedup experiments and disk failures remain recorded.

- a-check-delivery-build-status.txt: delivery compiler build exit=0
- a-check-final-build-status.txt: a-check compiler build exit=0
- counts-final-replay-status.txt: counts final replay exit=0
- counts-final-update-status.txt: counts final update exit=1
- counts-number-final-status.txt: counts number final exit=0
- counts-update-status.txt: counts update exit=0
- final-ir-replay-status.txt: final IR replay exit=1
- final-ir-status.txt: final IR exit=1
- lean-census-owner-status.txt: lean census owner exit=0
- sdk-restore-status.txt: sdk restore exit=0
- vet-delivery-status.txt: delivery vet exit=0
- vet-driver-delivery-status.txt: delivery driver vet exit=0
- vet-scanner-driver-status.txt: vet scanner driver exit=0
- vet-status.txt: vet exit=0
- wider-number-control-replay-status.txt: wider number replay exit=0
- wider-number-control-status.txt: wider number control exit=1

ir-owner-results.json.txt
```json
[
  {
    "name": "Native",
    "command": [
      "go",
      "test",
      "-p",
      "1",
      "-count=1",
      "-json",
      "-parallel=1",
      "-timeout",
      "90s",
      "./internal/ir",
      "-run",
      "^TestProduct_CallTargetReaderArchivesNative$"
    ],
    "exit": 1,
    "seconds": 14.808
  }
]

```

ir-owner-replay-results.json.txt
```json
[
  {
    "name": "Native",
    "command": [
      "go",
      "test",
      "-p",
      "1",
      "-count=1",
      "-json",
      "-parallel=1",
      "-timeout",
      "90s",
      "./internal/ir",
      "-run",
      "^TestProduct_CallTargetReaderArchivesNative$"
    ],
    "exit": 0,
    "seconds": 4.441
  },
  {
    "name": "Lower",
    "command": [
      "go",
      "test",
      "-p",
      "1",
      "-count=1",
      "-json",
      "-parallel=1",
      "-timeout",
      "90s",
      "./internal/ir",
      "-run",
      "^TestProduct_CallTargetReaderArchivesLower$"
    ],
    "exit": 0,
    "seconds": 4.203
  },
  {
    "name": "Fresh",
    "command": [
      "go",
      "test",
      "-p",
      "1",
      "-count=1",
      "-json",
      "-parallel=1",
      "-timeout",
      "90s",
      "./internal/ir",
      "-run",
      "^TestProduct_CallTargetReaderArchivesFresh$"
    ],
    "exit": 0,
    "seconds": 3.845
  },
  {
    "name": "Flow",
    "command": [
      "go",
      "test",
      "-p",
      "1",
      "-count=1",
      "-json",
      "-parallel=1",
      "-timeout",
      "90s",
      "./internal/ir",
      "-run",
      "^TestProduct_CallTargetReaderArchivesFlow$"
    ],
    "exit": 0,
    "seconds": 3.787
  }
]

```

Each runtime edit is printed exactly in the complete diff above. No runtime C/header changed during the late nullish, freshness-site, review-layout, terminal-pin or enum-mutant integrations.

Counts ownership: the inherited-fields row belongs to inherit-guards. The remaining rows follow exceptions-21-main admissions/refusals, generic payload boxing, checked Error reads and catchable initialization/library paths. feature-set-link-main and lint-features add no independent oracle count rows. Numeric deltas are measured observations; ownership is attributed to these code paths and the retained source count proofs.


## Delivery limitations

Validation is not claimed fully green. The full oracle snapshot had only enum/async harness failures; the corrected focused enum/async and complete step 21 rerun passed. The complete lowering tree passed after the last compiler fix. Native parent tests passed except the forbidden rerun leaf; native runtimeproducts and all IR reader variants passed after storage recovery.

The full stage1 lint run was interrupted after its existing TestCompilerAndStage1AgreeCombinedSanitizerWitness took 66.26 seconds. Its original observation is retained and it was not rerun. Four Markdown leaves previously over budget were explicitly excluded: TestMarkdownListLayout_Setup, TestProduct_WholeDocumentPreflightGoProducts, TestProduct_MarkdownblocksTextFormatter and TestProduct_GrainMarkdownFormatter. These are coverage limits, not green claims. TestCSSPrinterPreflightCostMutant timed out in the broad reader replay. Profile readers initially missed their Go oracle owner; their final state is recorded in the command table and subsequent logs. Other pending/failed reader runs are not claimed passed.

The late CSS failure was a compiler regression: unconditional wider-call storage bypass erased a fresh numeric narrowing in selector/nodes.ts. The fix asks for actual intervening writes, preserving ordinary main narrowing and the source wider-Error/unknown behavior. Its registered TypeScript-mode sanitizer program passes in 1.65s; unconditional-wider-storage-read is caught. Existing registry entries retain their original default language and native options.

All 41 changed .a files pass a-check on the final compiler. Final vet and all structural guards passed. Complete per-function measurements and over-budget observations are in test-time-audit.json.txt. No broad test failure or stale mutant marker was removed from the records.


## Resource stop

{
  "reason": "Stop heavy validation after cgroup memory exhaustion and compiler kills; no green claim for incomplete readers",
  "memory_events": "low 0\nhigh 0\nmax 65430\noom 1711\noom_kill 65\noom_group_kill 0\n",
  "memory_max": "17179869184\n",
  "interrupted": "stage1/cohere/markdownblocks/...; scanner broad replay not reached",
  "css_failed": "TestCSSPrinterPreflightCostMutant timeout",
  "lint_full": "interrupted; existing combined sanitizer witness 66.26 seconds"
}


Lint oracle owner builds and the GraphQL broad replay had compiler processes killed under the 16 GiB cgroup. Later broad Markdown validation was stopped and scanner broad replay was not reached. These are unresolved validation limits. The four-member source changes, 61 attempted mutant replays (with stale markers and one failed build distinguished), focused controls, final Linux counts, a-check, guards and vet are preserved for review. This branch should not be submitted to Loom as fully validated.

