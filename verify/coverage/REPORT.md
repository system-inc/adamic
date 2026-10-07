# What the generator never reaches
Coverage of the compiler (Go: internal/lower, internal/native, internal/ir, internal/flow) and the C runtime (internal/native/runtime) under the randomized test-program generator, against the hand-written oracle fixtures. Written by verify/coverage/measure.sh; REPORT.json beside it holds the same numbers and the longer lists.
- **generator**: seeds 1 to 2000, parallel 6, commit 7863216ed4, verdicts agreed 1872, finding 0, checked 126, not yet 0, invalid 0, unfit 2, C profiles 2000, exit 0, seconds 743
- **fixtures**: test go test ./internal/oracle -run TestNativeAgreesWithNode, subtests passed 327, subtests failed 1 (internal/oracle/testdata/navigation.a), C profiles 1112, exit 1, seconds 230

Go's coverage counts blocks (each arm of an `if`, each `case`, is its own block) and the statements in them; Go has no branch coverage, so on the Go side the branch share is the block share. clang's source-based coverage counts branch outcomes (each condition's true and its false), regions and lines.

## (b) What the 2,000 seeds reach
| Area | Branches | Statements or lines |
|---|---|---|
| internal/native/emit*.go | 77.8% of 980 blocks | 80.5% of 1674 statements |
| internal/lower | 36.5% of 6340 blocks | 36.3% of 8847 statements |
| internal/native (all Go) | 62.5% of 2168 blocks | 65.4% of 3527 statements |
| internal/native (Go, besides units*.go, tsgo.go, library.go) | 72.9% of 1857 blocks | 75.6% of 3049 statements |
| internal/ir | 56.4% of 195 blocks | 61.9% of 181 statements |
| internal/flow | 23.4% of 888 blocks | 27.5% of 1343 statements |
| internal/native/runtime (C) | 38.5% of 4560 branch outcomes | 40.6% of 8510 lines; 54.5% of 468 functions; 43.1% of 6617 regions |

## (c) The oracle fixtures, for comparison
| Area | Generator branches | Fixtures branches | Generator statements or lines | Fixtures statements or lines |
|---|---|---|---|---|
| internal/native/emit*.go | 77.8% | 97.8% | 80.5% | 98.7% |
| internal/lower | 36.5% | 73.6% | 36.3% | 79.3% |
| internal/native (all Go) | 62.5% | 82.9% | 65.4% | 85.3% |
| internal/native (Go, besides units*.go, tsgo.go, library.go) | 72.9% | 95.5% | 75.6% | 97.0% |
| internal/ir | 56.4% | 75.4% | 61.9% | 82.9% |
| internal/flow | 23.4% | 24.7% | 27.5% | 29.5% |
| internal/native/runtime (C) | 38.5% | 79.5% | 40.6% | 83.6% |

Of the generator's blind regions, 609 Go regions (1249 statements) and 201 C regions (896 lines) are blind to the fixtures too.

What one side reaches and the other doesn't:

| Area | Both | Generator only | Fixtures only |
|---|---|---|---|
| internal/native/emit*.go | 762 blocks | 0 blocks, 0 statements | 196 blocks, 304 statements |
| internal/lower | 2305 blocks | 8 blocks, 10 statements | 2364 blocks, 3807 statements |
| internal/native (all Go) | 1350 blocks | 6 blocks, 6 statements | 447 blocks, 707 statements |
| internal/native (Go, besides units*.go, tsgo.go, library.go) | 1350 blocks | 3 blocks, 3 statements | 423 blocks, 656 statements |
| internal/ir | 109 blocks | 1 blocks, 1 statements | 38 blocks, 39 statements |
| internal/flow | 208 blocks | 0 blocks, 0 statements | 11 blocks, 27 statements |
| runtime lines (C) | 3438 | 16 | 3654 |
| runtime functions (C) | 255 | 0 | 168 |
| runtime branch outcomes (C) | 1729 | 26 | 1897 |

### Go the generator reaches and the fixtures don't (top 20 functions, internal/lower and internal/native)

| Statements | Function | Where |
|---|---|---|
| 6 | `lowering.cycleFieldMatches` | internal/lower/class_inheritance.go:767-795 |
| 3 | `lowering.libraryFailure` | internal/lower/exceptions.go:184-259 |
| 3 | `UsesTSGo` | internal/native/tsgo.go:15-22 |
| 1 | `markCounter` | internal/lower/counters.go:76-157 |
| 1 | `regionPlan.escaping` | internal/native/region.go:167-246 |
| 1 | `emitter.spreadArray` | internal/native/reuse.go:718-745 |
| 1 | `converted` | internal/native/union.go:12-32 |

### Go the fixtures reach and the generator doesn't (top 20 functions, internal/lower and internal/native)

| Statements | Function | Where |
|---|---|---|
| 125 | `emitter.evaluate` | internal/native/emit_expressions.go:13-642 |
| 103 | `lowering.staticInstance` | internal/lower/class_static.go:48-212 |
| 89 | `lowering.libraryMapGroupBy` | internal/lower/library_map_set.go:302-437 |
| 86 | `lowering.finishAccessors` | internal/lower/class_accessors.go:207-347 |
| 74 | `lowering.instantiateFunction` | internal/lower/generic.go:24-136 |
| 68 | `lowering.destructureIterator` | internal/lower/iteration_consume.go:110-203 |
| 66 | `lowering.librarySetMethod` | internal/lower/library_map_set.go:13-107 |
| 65 | `lowering.objectCall` | internal/lower/library_object.go:11-147 |
| 61 | `lowering.staticInitializationReads` | internal/lower/class_static.go:309-436 |
| 60 | `lowering.planIteration` | internal/lower/iteration.go:162-298 |
| 58 | `lowering.collectIteration` | internal/lower/iteration_consume.go:11-96 |
| 56 | `lowering.optionalRelationFailure` | internal/lower/proven_relations.go:33-163 |
| 54 | `lowering.libraryMathNumberCall` | internal/lower/library_math_number.go:67-181 |
| 53 | `lowering.staticDeclaration` | internal/lower/class_static.go:215-297 |
| 50 | `lowering.libraryArrayForOf` | internal/lower/library_array.go:481-580 |
| 46 | `lowering.nonObjectOwnProperty` | internal/lower/prototype_own.go:11-87 |
| 44 | `lowering.jsonInput` | internal/lower/library_json_stringify.go:72-154 |
| 44 | `lowering.libraryStringMethod` | internal/lower/library_string.go:151-230 |
| 42 | `lowering.libraryArrayCopyWithin` | internal/lower/library_array.go:262-311 |
| 41 | `emitter.accessorDeclarations` | internal/native/class_accessors.go:10-77 |

### C the generator reaches and the fixtures don't (top 20 functions)

| Lines | Function | Where |
|---|---|---|
| 4 | `adamic_math_round` | internal/native/runtime/math.c:7-21 |
| 3 | `adamic_string_code_point_at` | internal/native/runtime/string_slice_impl.h:59-85 |
| 3 | `builder_unit` | internal/native/runtime/string_builder_impl.h:32-47 |
| 2 | `adamic_math_sign` | internal/native/runtime/math.c:23-28 |
| 2 | `adamic_number_parse_int` | internal/native/runtime/parse.c:155-221 |
| 2 | `adamic_math_min` | internal/native/runtime/math.c:40-48 |

### C the fixtures reach and the generator doesn't (top 20 functions)

| Lines | Function | Where |
|---|---|---|
| 149 | `regex_run` | internal/native/runtime/regexp.c:172-401 |
| 121 | `kernel_rem_pio2` | internal/native/runtime/ieee754.c:459-659 |
| 113 | `ieee754_rem_pio2` | internal/native/runtime/ieee754.c:133-285 |
| 91 | `merge_high` | internal/native/runtime/sort.c:314-408 |
| 88 | `kernel_tan` | internal/native/runtime/ieee754.c:748-850 |
| 84 | `adamic_math_expm1` | internal/native/runtime/ieee754.c:2237-2355 |
| 83 | `merge_low` | internal/native/runtime/sort.c:221-311 |
| 66 | `adamic_math_exp` | internal/native/runtime/ieee754.c:1465-1556 |
| 63 | `adamic_math_atan` | internal/native/runtime/ieee754.c:1135-1219 |
| 62 | `adamic_math_log1p` | internal/native/runtime/ieee754.c:1804-1901 |
| 57 | `adamic_math_log` | internal/native/runtime/ieee754.c:1658-1737 |
| 54 | `double_to_radix` | internal/native/runtime/radix.c:53-131 |
| 53 | `write_value` | internal/native/runtime/json_stringify.c:147-203 |
| 47 | `power_of_two` | internal/native/runtime/parse.c:69-123 |
| 46 | `adamic_math_asin` | internal/native/runtime/ieee754.c:1008-1073 |
| 46 | `adamic_math_atan2` | internal/native/runtime/ieee754.c:1248-1339 |
| 45 | `gallop_left` | internal/native/runtime/sort.c:119-167 |
| 45 | `gallop_right` | internal/native/runtime/sort.c:170-218 |
| 42 | `adamic_number_from_string` | internal/native/runtime/library_math_number.c:57-99 |
| 42 | `adamic_string_normalize` | internal/native/runtime/normalize.c:242-290 |

## (a) The blind map: what no generated program executes

### Go, internal/lower and internal/native (top 40 of 1181 untouched regions, by statements)

A function no program enters is one region; inside a function that ran, a region is a run of blocks that never ran, cut at each `case` so every switch arm stands alone. The First line column names the case a region opens, or the case it sits inside.

Left out of the ranking, as unreachable by construction (23 regions, 571 statements): the fuzzer compiles through `adamic c` and `adamic js` and runs clang itself, so native.Build and Flags, the split build (units.go, units_tsgo.go), the runtime cache (library.go) and tsgo (tsgo.go in native and lower, which `adamic c` refuses) never run in the measured binary. REPORT.json lists them under blindMap.goStructural.

The last column is how many of the region's statements the oracle fixtures run; 0 means nothing in the fixtures reaches it either.

| # | Statements | Kind | Function | Where | First line | Fixtures run |
|---|---|---|---|---|---|---|
| 1 | 120 | function never entered | `lowering.staticInstance` | internal/lower/class_static.go:48-212 |  | 103 |
| 2 | 112 | function never entered | `lowering.libraryMapGroupBy` | internal/lower/library_map_set.go:302-437 |  | 89 |
| 3 | 94 | blocks never run | `lowering.finishAccessors` | internal/lower/class_accessors.go:217-346 |  | 85 |
| 4 | 90 | blocks never run | `lowering.planIteration` | internal/lower/iteration.go:168-297 |  | 60 |
| 5 | 90 | function never entered | `lowering.objectCall` | internal/lower/library_object.go:11-147 |  | 65 |
| 6 | 89 | function never entered | `lowering.staticInitializationReads` | internal/lower/class_static.go:309-436 |  | 61 |
| 7 | 85 | function never entered | `lowering.instantiateFunction` | internal/lower/generic.go:24-136 |  | 74 |
| 8 | 84 | function never entered | `lowering.optionalRelationFailure` | internal/lower/proven_relations.go:33-163 |  | 56 |
| 9 | 79 | function never entered | `lowering.destructureIterator` | internal/lower/iteration_consume.go:110-203 |  | 68 |
| 10 | 72 | blocks never run | `lowering.librarySetMethod` | internal/lower/library_map_set.go:19-106 |  | 66 |
| 11 | 68 | function never entered | `lowering.collectIteration` | internal/lower/iteration_consume.go:11-96 |  | 58 |
| 12 | 62 | blocks never run | `lowering.libraryArrayForOf` | internal/lower/library_array.go:499-579 |  | 49 |
| 13 | 58 | blocks never run | `lowering.staticDeclaration` | internal/lower/class_static.go:222-296 |  | 53 |
| 14 | 56 | function never entered | `lowering.jsonInput` | internal/lower/library_json_stringify.go:72-154 |  | 44 |
| 15 | 55 | function never entered | `lowering.plainEnumerableObject` | internal/lower/library_for_in.go:68-153 |  | 38 |
| 16 | 54 | function never entered | `lowering.prototypeHazard` | internal/lower/prototype.go:176-256 |  | 34 |
| 17 | 53 | function never entered | `lowering.libraryStringMethod` | internal/lower/library_string.go:151-230 |  | 44 |
| 18 | 52 | function never entered | `lowering.nonObjectOwnProperty` | internal/lower/prototype_own.go:11-87 |  | 46 |
| 19 | 51 | function never entered | `lowering.destructureFrom` | internal/lower/collections.go:281-355 |  | 39 |
| 20 | 48 | blocks never run | `lowering.input` | internal/lower/input.go:21-91 | `in case l.isPreludeFunction(callee, "readTextFile"):` | 22 |
| 21 | 48 | blocks never run | `emitter.accessorDeclarations` | internal/native/class_accessors.go:14-76 |  | 41 |
| 22 | 46 | blocks never run | `lowering.accessorLiteral` | internal/lower/class_accessors.go:141-202 |  | 36 |
| 23 | 45 | function never entered | `lowering.libraryArrayCopyWithin` | internal/lower/library_array.go:262-311 |  | 42 |
| 24 | 45 | function never entered | `lowering.libraryArrayFlatMap` | internal/lower/library_array.go:348-408 |  | 37 |
| 25 | 45 | function never entered | `lowering.functionExpression` | internal/lower/library_function_expressions.go:10-70 |  | 36 |
| 26 | 44 | function never entered | `constructorSuperFlow.loop` | internal/lower/class_super.go:218-282 |  | 32 |
| 27 | 44 | function never entered | `lowering.provePredicate` | internal/lower/predicates.go:38-103 |  | 35 |
| 28 | 44 | blocks never run | `lowering.objectPrototypeCall` | internal/lower/prototype.go:102-170 |  | 34 |
| 29 | 43 | function never entered | `lowering.cast` | internal/lower/cast.go:15-75 |  | 34 |
| 30 | 41 | function never entered | `lowering.iterationOrigin` | internal/lower/iteration_origin.go:109-171 |  | 29 |
| 31 | 41 | function never entered | `lowering.libraryArrayFlat` | internal/lower/library_array.go:205-260 |  | 34 |
| 32 | 38 | blocks never run | `lowering.classGenericCall` | internal/lower/class_generic_calls.go:23-70 |  | 34 |
| 33 | 38 | function never entered | `constructorSuperFlow.statement` | internal/lower/class_super.go:86-149 |  | 26 |
| 34 | 38 | function never entered | `lowering.functionValue` | internal/lower/expression.go:947-992 |  | 35 |
| 35 | 38 | function never entered | `emitter.mapForEach` | internal/native/emit_maps.go:27-74 |  | 38 |
| 36 | 37 | function never entered | `lowering.forIn` | internal/lower/library_for_in.go:13-66 |  | 28 |
| 37 | 37 | function never entered | `lowering.stringRaw` | internal/lower/library_string.go:280-329 |  | 31 |
| 38 | 36 | function never entered | `predicateProof.statement` | internal/lower/predicates.go:172-231 |  | 20 |
| 39 | 35 | blocks never run | `lowering.superAccessor` | internal/lower/class_accessors.go:58-103 |  | 29 |
| 40 | 35 | function never entered | `lowering.jsonCall` | internal/lower/library_json_stringify.go:12-65 |  | 24 |

### Go, internal/native/emit*.go alone (top 40 of 111 untouched regions, by statements)

The emitter's regions are small, so few reach the table above; here they are on their own.

| # | Statements | Kind | Function | Where | First line | Fixtures run |
|---|---|---|---|---|---|---|
| 1 | 38 | function never entered | `emitter.mapForEach` | internal/native/emit_maps.go:27-74 |  | 38 |
| 2 | 34 | function never entered | `emitter.switchStatement` | internal/native/emit_statements.go:395-435 |  | 34 |
| 3 | 13 | blocks never run | `emitter.evaluate` | internal/native/emit_expressions.go:257-271 | `in case ir.CallClosure:` | 13 |
| 4 | 13 | blocks never run | `emitter.evaluate` | internal/native/emit_expressions.go:320-335 | `in case ir.ArrayFill:` | 13 |
| 5 | 13 | blocks never run | `emitter.evaluate` | internal/native/emit_expressions.go:351-365 | `in case ir.CheckedCast:` | 13 |
| 6 | 12 | blocks never run | `emitter.evaluate` | internal/native/emit_expressions.go:591-607 | `in case ir.StringFromCodes:` | 12 |
| 7 | 11 | blocks never run | `emitter.evaluate` | internal/native/emit_expressions.go:509-526 | `in case ir.HasOwn:` | 8 |
| 8 | 9 | case never taken | `emitter.evaluate` | internal/native/emit_expressions.go:272-286 | `case expression.Returns == 0:` | 9 |
| 9 | 8 | function never entered | `cBytes` | internal/native/emit.go:108-120 |  | 8 |
| 10 | 8 | blocks never run | `emitter.coalesce` | internal/native/emit_branches.go:130-138 |  | 7 |
| 11 | 7 | blocks never run | `emitter.returnStatement` | internal/native/emit_functions.go:95-103 |  | 6 |
| 12 | 6 | blocks never run | `emitter.evaluate` | internal/native/emit_expressions.go:196-206 | `in case ir.WeakOf:` | 6 |
| 13 | 6 | blocks never run | `emitter.forOf` | internal/native/emit_statements.go:352-358 |  | 6 |
| 14 | 5 | function never entered | `emitter.comparator` | internal/native/emit_arrays.go:202-210 |  | 5 |
| 15 | 5 | blocks never run | `emitter.evaluate` | internal/native/emit_expressions.go:420-426 | `in case ir.ArraySort:` | 5 |
| 16 | 5 | blocks never run | `emitter.evaluate` | internal/native/emit_expressions.go:455-462 | `in case ir.MapKeys:` | 3 |
| 17 | 5 | blocks never run | `emitter.arguments` | internal/native/emit_functions.go:187-192 |  | 5 |
| 18 | 4 | blocks never run | `equality` | internal/native/emit_arrays.go:178-184 | `in case ir.Boolean:` | 4 |
| 19 | 4 | blocks never run | `emitter.evaluate` | internal/native/emit_expressions.go:63-67 | `in case ir.Call:` | 4 |
| 20 | 4 | blocks never run | `emitter.mathCall` | internal/native/emit_numbers.go:69-74 | `in case 0:` | 3 |
| 21 | 3 | blocks never run | `emitter.evaluate` | internal/native/emit_expressions.go:23-25 | `in case ir.RegExpGroup:` | 3 |
| 22 | 3 | blocks never run | `emitter.numberFormat` | internal/native/emit_numbers.go:20-23 |  | 3 |
| 23 | 3 | blocks never run | `emitter.mathCall` | internal/native/emit_numbers.go:60-63 |  | 3 |
| 24 | 3 | blocks never run | `emitter.fieldSlot` | internal/native/emit_objects.go:26-29 |  | 3 |
| 25 | 3 | blocks never run | `emitter.statement` | internal/native/emit_statements.go:168-170 | `in case ir.Panic:` | 3 |
| 26 | 2 | blocks never run | `C` | internal/native/emit.go:34-35 |  | 2 |
| 27 | 2 | blocks never run | `emitter.arrayVisit` | internal/native/emit_arrays.go:86-88 | `in case "filter":` | 2 |
| 28 | 2 | blocks never run | `emitter.arrayReduce` | internal/native/emit_arrays.go:162-164 |  | 2 |
| 29 | 2 | blocks never run | `joinKind` | internal/native/emit_arrays.go:193-195 | `in case ir.Boolean:` | 2 |
| 30 | 2 | blocks never run | `emitter.evaluate` | internal/native/emit_expressions.go:32-34 | `in case ir.IsNull:` | 2 |
| 31 | 2 | blocks never run | `emitter.evaluate` | internal/native/emit_expressions.go:93-95 | `in case ir.InstanceOf:` | 0 |
| 32 | 2 | blocks never run | `emitter.evaluate` | internal/native/emit_expressions.go:225-227 | `in case ir.StringLength:` | 2 |
| 33 | 2 | blocks never run | `emitter.evaluate` | internal/native/emit_expressions.go:398-399 | `in case ir.MapEntries:` | 2 |
| 34 | 2 | blocks never run | `emitter.evaluate` | internal/native/emit_expressions.go:580-582 | `in case ir.Length:` | 2 |
| 35 | 2 | blocks never run | `emitter.evaluate` | internal/native/emit_expressions.go:585-587 | `in case ir.JSONStringify:` | 2 |
| 36 | 2 | blocks never run | `emitter.binary` | internal/native/emit_expressions.go:674-676 | `in case operandType == ir.Union && operator == ir.Equal:` | 2 |
| 37 | 2 | blocks never run | `emitter.objectLiteral` | internal/native/emit_objects.go:123-124 |  | 2 |
| 38 | 2 | blocks never run | `emitter.forOf` | internal/native/emit_statements.go:307-309 |  | 2 |
| 39 | 2 | blocks never run | `emitter.forOf` | internal/native/emit_statements.go:311-313 |  | 2 |
| 40 | 2 | blocks never run | `cType` | internal/native/emit_values.go:28-30 | `in case ir.Union:` | 2 |

### C, internal/native/runtime (top 40 of 461 untouched regions, by lines that never ran)

A function no program calls is one region; inside a function that ran, a region is an outermost code region (a branch's arm, a loop body, the rest of a function after a return) that never ran.

The last column is how many of the region's lines the oracle fixtures run.

| # | Lines | Kind | Function | Where | First line | Fixtures run |
|---|---|---|---|---|---|---|
| 1 | 167 | function never called | `kernel_rem_pio2` | internal/native/runtime/ieee754.c:459-659 |  | 121 |
| 2 | 132 | function never called | `ieee754_rem_pio2` | internal/native/runtime/ieee754.c:133-285 |  | 113 |
| 3 | 105 | function never called | `adamic_math_expm1` | internal/native/runtime/ieee754.c:2237-2355 |  | 84 |
| 4 | 93 | function never called | `merge_high` | internal/native/runtime/sort.c:314-408 |  | 91 |
| 5 | 89 | function never called | `kernel_tan` | internal/native/runtime/ieee754.c:748-850 |  | 88 |
| 6 | 89 | function never called | `merge_low` | internal/native/runtime/sort.c:221-311 |  | 83 |
| 7 | 87 | function never called | `adamic_math_log1p` | internal/native/runtime/ieee754.c:1804-1901 |  | 62 |
| 8 | 82 | function never called | `adamic_math_atan2` | internal/native/runtime/ieee754.c:1248-1339 |  | 46 |
| 9 | 79 | function never called | `adamic_math_exp` | internal/native/runtime/ieee754.c:1465-1556 |  | 66 |
| 10 | 78 | function never called | `adamic_math_atan` | internal/native/runtime/ieee754.c:1135-1219 |  | 63 |
| 11 | 76 | function never called | `adamic_math_log` | internal/native/runtime/ieee754.c:1658-1737 |  | 57 |
| 12 | 63 | function never called | `adamic_read_directory` | internal/native/runtime/directory.c:67-130 |  | 0 |
| 13 | 62 | function never called | `adamic_math_asin` | internal/native/runtime/ieee754.c:1008-1073 |  | 46 |
| 14 | 59 | function never called | `adamic_math_acos` | internal/native/runtime/ieee754.c:877-936 |  | 38 |
| 15 | 57 | function never called | `write_value` | internal/native/runtime/json_stringify.c:147-203 |  | 53 |
| 16 | 57 | function never called | `double_to_radix` | internal/native/runtime/radix.c:53-131 |  | 54 |
| 17 | 54 | function never called | `power_of_two` | internal/native/runtime/parse.c:69-123 |  | 47 |
| 18 | 53 | function never called | `decode_step` | internal/native/runtime/input.c:35-89 |  | 0 |
| 19 | 49 | region never run | `regex_run` | internal/native/runtime/regexp.c:329-377 | `case 9: {` | 39 |
| 20 | 49 | function never called | `gallop_left` | internal/native/runtime/sort.c:119-167 |  | 45 |
| 21 | 49 | function never called | `gallop_right` | internal/native/runtime/sort.c:170-218 |  | 45 |
| 22 | 47 | function never called | `adamic_string_normalize` | internal/native/runtime/normalize.c:242-290 |  | 42 |
| 23 | 44 | function never called | `adamic_record_keys` | internal/native/runtime/record.c:149-192 |  | 0 |
| 24 | 42 | function never called | `adamic_math_log2` | internal/native/runtime/ieee754.c:1995-2073 |  | 32 |
| 25 | 42 | function never called | `adamic_number_from_string` | internal/native/runtime/library_math_number.c:57-99 |  | 42 |
| 26 | 40 | function never called | `adamic_plain_object_keys` | internal/native/runtime/library_language.c:28-70 |  | 0 |
| 27 | 39 | function never called | `collection_next` | internal/native/runtime/map_set.c:66-105 |  | 39 |
| 28 | 39 | region never run | `regex_substitution` | internal/native/runtime/regexp.c:855-906 | `case '<': {` | 36 |
| 29 | 39 | function never called | `adamic_utf8_at` | internal/native/runtime/utf8.c:18-60 |  | 16 |
| 30 | 38 | function never called | `adamic_math_cbrt` | internal/native/runtime/ieee754.c:2358-2440 |  | 33 |
| 31 | 37 | function never called | `regex_clone` | internal/native/runtime/regexp.c:118-155 |  | 31 |
| 32 | 34 | function never called | `adamic_math_log10` | internal/native/runtime/ieee754.c:2102-2140 |  | 23 |
| 33 | 34 | function never called | `decode` | internal/native/runtime/input.c:98-131 |  | 9 |
| 34 | 34 | function never called | `read_all` | internal/native/runtime/input.c:227-260 |  | 14 |
| 35 | 34 | function never called | `insert` | internal/native/runtime/weak.c:67-101 |  | 29 |
| 36 | 33 | function never called | `adamic_class_object_keys` | internal/native/runtime/class_features.c:42-74 |  | 27 |
| 37 | 32 | function never called | `adamic_read_text_file` | internal/native/runtime/input.c:262-297 |  | 23 |
| 38 | 32 | region never run | `substitution` | internal/native/runtime/string_replace_impl.h:52-83 | `pieces list = {NULL, 0, 0};` | 32 |
| 39 | 31 | function never called | `new_chunk` | internal/native/runtime/heap.c:107-137 |  | 0 |
| 40 | 31 | function never called | `kernel_cos` | internal/native/runtime/ieee754.c:320-351 |  | 30 |

## How it was measured

Run by `verify/coverage/measure.sh` at 7863216ed419151286c34df3896496dd15f4ec11. Reproduce with `verify/coverage/measure.sh <fresh directory>`
(SEED, COUNT, PARALLEL and FIXTURES in the environment). What it ran:

```
# the coverage runtime: sanitized, -fprofile-instr-generate -fcoverage-mapping, its own cache key
ADAMIC_C_COVERAGE=1 go run -trimpath ./verify/coverage/runtimelibrary
go build -trimpath -o adamic-fuzz ./cmd/adamic-fuzz

# the generator: seeds 1 to 2000, every default family
ADAMIC_C_COVERAGE=1 ADAMIC_C_COVERAGE_DIRECTORY=generator/c GOCOVERDIR=generator/go \
  GOFLAGS="-trimpath -cover -coverpkg=./cmd/adamic,./internal/lower/...,./internal/native/...,./internal/ir/...,./internal/flow/..." \
  adamic-fuzz -root . -seed 1 -count 2000 -parallel 6 -shrink=false -v -work generator/work

# the oracle fixtures, uncached so every binary really runs
ADAMIC_GATE_UNCACHED=1 ADAMIC_C_COVERAGE=1 LLVM_PROFILE_FILE=fixtures/c/oracle-%p-%m.profraw \
  go test -trimpath -count=1 -v -cover -coverpkg=./internal/lower/...,./internal/native/...,./internal/ir/...,./internal/flow/... -coverprofile=fixtures/go.txt \
  -run '^TestNativeAgreesWithNode$' -parallel 6 -timeout 90m ./internal/oracle

go tool covdata textfmt -i=generator/go -o=generator/go.txt
llvm-profdata merge -sparse -f <side>/profraw.list -o <side>/c.profdata
llvm-cov export -format=text <runtime objects> -instr-profile <side>/c.profdata > <side>/c.json
llvm-cov export -format=lcov <runtime objects> -instr-profile <side>/c.profdata > <side>/c.lcov
python3 verify/coverage/analyze.py ...
```

Caveats, as measured:

- Go has block coverage, not branch coverage; the Go "branch" shares are block shares.
- adamic-fuzz runs the compiler twice per program (`adamic c` and `adamic js`), so the Go side
  counts the JavaScript backend's lowering too; lowering is the same for both.
- The C runtime is read through the sanitized runtime's objects. The oracle also runs a release
  build (-O2, the size-class allocator on) and a counted build of each fixture. Their runs are
  credited to the runtime's external functions whose structure matches the sanitized build's;
  their static functions are named after the directory they were compiled in, so they aren't
  credited, and llvm-cov skips any function whose profile doesn't match (export.log counts them).
- The fuzzer runs with -shrink=false, so exactly the seeds' programs are measured.
- cmd/adamic is in the fuzzer's -coverpkg only because this toolchain links the coverage writer
  into a binary only when its main package is covered; the report doesn't read it.
- adamic.h's five static inline functions are counted only for calls from inside the runtime; their
  copies inlined into each program's main.c aren't read.
- A run that ends in a signal (an ASan abort, a deadline kill) writes no C profile.
