# What the generator never reaches
Coverage of the compiler (Go: internal/lower, internal/native, internal/ir, internal/flow) and the C runtime (internal/native/runtime) under the randomized test-program generator, against the hand-written oracle fixtures. Written by verify/coverage/measure.sh; REPORT.json beside it holds the same numbers and the longer lists.
- **generator**: seeds 1 to 2000, parallel 6, commit b2bf06e25c+dirty, verdicts crash 0, agreed 1830, finding 46, checked 122, not yet 0, invalid 0, unfit 2, C profiles 2000, exit 1, seconds 528
- **fixtures**: test go test ./internal/oracle -run TestNativeAgreesWithNode, subtests passed 327, subtests failed 1 (internal/oracle/testdata/navigation.a), C profiles 1112, exit 1, seconds 121

Go's coverage counts blocks (each arm of an `if`, each `case`, is its own block) and the statements in them; Go has no branch coverage, so on the Go side the branch share is the block share. clang's source-based coverage counts branch outcomes (each condition's true and its false), regions and lines.

## (b) What the seeds (1 to 2000) reach, and what neither the seeds nor the fixtures reach
The last column is the code no line of defense executes: neither a generated program nor a hand-written fixture.

| Area | Branches | Statements or lines | Covered by neither |
|---|---|---|---|
| internal/native/emit*.go | 84.6% of 980 blocks | 87.5% of 1674 statements | 20 statements (1.2%) |
| internal/lower | 48.4% of 6340 blocks | 50.4% of 8847 statements | 1815 statements (20.5%) |
| internal/native (all Go) | 68.8% of 2169 blocks | 72.5% of 3528 statements | 511 statements (14.5%) |
| internal/native (Go, besides units*.go, tsgo.go, library.go) | 80.1% of 1858 blocks | 83.7% of 3050 statements | 87 statements (2.9%) |
| internal/ir | 60.5% of 195 blocks | 66.3% of 181 statements | 29 statements (16.0%) |
| internal/flow | 24.4% of 888 blocks | 29.3% of 1343 statements | 947 statements (70.5%) |
| internal/native/runtime (C) | 62.4% of 4560 branch outcomes | 70.0% of 8510 lines; 73.9% of 468 functions; 69.2% of 6617 regions | 1048 lines (12.3%), 678 branch outcomes |

## (c) The oracle fixtures, for comparison
| Area | Generator branches | Fixtures branches | Generator statements or lines | Fixtures statements or lines |
|---|---|---|---|---|
| internal/native/emit*.go | 84.6% | 97.8% | 87.5% | 98.7% |
| internal/lower | 48.4% | 73.6% | 50.4% | 79.3% |
| internal/native (all Go) | 68.8% | 82.8% | 72.5% | 85.3% |
| internal/native (Go, besides units*.go, tsgo.go, library.go) | 80.1% | 95.4% | 83.7% | 97.0% |
| internal/ir | 60.5% | 75.4% | 66.3% | 82.9% |
| internal/flow | 24.4% | 24.7% | 29.3% | 29.5% |
| internal/native/runtime (C) | 62.4% | 79.5% | 70.0% | 83.6% |

Of the generator's blind regions, 878 Go regions (1555 statements) and 286 C regions (1002 lines) are blind to the fixtures too.

What one side reaches and the other doesn't:

| Area | Both | Generator only | Fixtures only |
|---|---|---|---|
| internal/native/emit*.go | 827 blocks | 2 blocks, 2 statements | 131 blocks, 190 statements |
| internal/lower | 3050 blocks | 16 blocks, 20 statements | 1619 blocks, 2573 statements |
| internal/native (all Go) | 1484 blocks | 8 blocks, 8 statements | 313 blocks, 460 statements |
| internal/native (Go, besides units*.go, tsgo.go, library.go) | 1484 blocks | 5 blocks, 5 statements | 289 blocks, 409 statements |
| internal/ir | 116 blocks | 2 blocks, 2 statements | 31 blocks, 32 statements |
| internal/flow | 217 blocks | 0 blocks, 0 statements | 2 blocks, 3 statements |
| runtime lines (C) | 5589 | 370 | 1503 |
| runtime functions (C) | 340 | 6 | 83 |
| runtime branch outcomes (C) | 2591 | 256 | 1035 |

### Go the generator reaches and the fixtures don't (top 20 functions, internal/lower and internal/native)

| Statements | Function | Where |
|---|---|---|
| 8 | `lowering.input` | internal/lower/input.go:13-94 |
| 6 | `lowering.cycleFieldMatches` | internal/lower/class_inheritance.go:767-795 |
| 3 | `lowering.libraryFailure` | internal/lower/exceptions.go:184-259 |
| 3 | `UsesTSGo` | internal/native/tsgo.go:15-22 |
| 2 | `cycleFinder.reaches` | internal/lower/cycles.go:351-445 |
| 1 | `markCounter` | internal/lower/counters.go:76-157 |
| 1 | `unwrap` | internal/native/emit_branches.go:100-117 |
| 1 | `emitter.evaluate` | internal/native/emit_expressions.go:13-642 |
| 1 | `regionPlan.escaping` | internal/native/region.go:167-246 |
| 1 | `emitter.spreadArray` | internal/native/reuse.go:718-745 |
| 1 | `converted` | internal/native/union.go:12-32 |

### Go the fixtures reach and the generator doesn't (top 20 functions, internal/lower and internal/native)

| Statements | Function | Where |
|---|---|---|
| 103 | `lowering.staticInstance` | internal/lower/class_static.go:48-212 |
| 100 | `emitter.evaluate` | internal/native/emit_expressions.go:13-642 |
| 66 | `lowering.librarySetMethod` | internal/lower/library_map_set.go:13-107 |
| 61 | `lowering.staticInitializationReads` | internal/lower/class_static.go:309-436 |
| 58 | `lowering.collectIteration` | internal/lower/iteration_consume.go:11-96 |
| 56 | `lowering.optionalRelationFailure` | internal/lower/proven_relations.go:33-163 |
| 53 | `lowering.staticDeclaration` | internal/lower/class_static.go:215-297 |
| 50 | `lowering.libraryArrayForOf` | internal/lower/library_array.go:481-580 |
| 50 | `lowering.libraryMathNumberCall` | internal/lower/library_math_number.go:67-181 |
| 46 | `lowering.nonObjectOwnProperty` | internal/lower/prototype_own.go:11-87 |
| 44 | `lowering.libraryStringMethod` | internal/lower/library_string.go:151-230 |
| 42 | `lowering.libraryArrayCopyWithin` | internal/lower/library_array.go:262-311 |
| 37 | `lowering.libraryArrayFlatMap` | internal/lower/library_array.go:348-408 |
| 36 | `lowering.functionExpression` | internal/lower/library_function_expressions.go:10-70 |
| 35 | `lowering.functionValue` | internal/lower/expression.go:947-992 |
| 35 | `lowering.provePredicate` | internal/lower/predicates.go:38-103 |
| 35 | `lowering.objectPrototypeCall` | internal/lower/prototype.go:77-171 |
| 34 | `lowering.cast` | internal/lower/cast.go:15-75 |
| 34 | `lowering.libraryArrayFlat` | internal/lower/library_array.go:205-260 |
| 34 | `lowering.prototypeHazard` | internal/lower/prototype.go:176-256 |

### C the generator reaches and the fixtures don't (top 20 functions)

| Lines | Function | Where |
|---|---|---|
| 38 | `decode_step` | internal/native/runtime/input.c:35-89 |
| 36 | `adamic_math_atan2` | internal/native/runtime/ieee754.c:1248-1339 |
| 26 | `adamic_file_status` | internal/native/runtime/directory.c:132-161 |
| 25 | `decode` | internal/native/runtime/input.c:98-131 |
| 21 | `adamic_math_acos` | internal/native/runtime/ieee754.c:877-936 |
| 20 | `adamic_math_expm1` | internal/native/runtime/ieee754.c:2237-2355 |
| 20 | `failure` | internal/native/runtime/input.c:168-194 |
| 18 | `adamic_math_log1p` | internal/native/runtime/ieee754.c:1804-1901 |
| 18 | `failure` | internal/native/runtime/directory.c:37-61 |
| 15 | `adamic_math_atan` | internal/native/runtime/ieee754.c:1135-1219 |
| 14 | `adamic_math_asin` | internal/native/runtime/ieee754.c:1008-1073 |
| 14 | `adamic_math_cosh` | internal/native/runtime/ieee754.c:2578-2619 |
| 13 | `adamic_math_exp` | internal/native/runtime/ieee754.c:1465-1556 |
| 8 | `adamic_math_sinh` | internal/native/runtime/ieee754.c:2642-2673 |
| 8 | `read_all` | internal/native/runtime/input.c:227-260 |
| 7 | `adamic_math_tanh` | internal/native/runtime/ieee754.c:2699-2733 |
| 7 | `adamic_math_acosh` | internal/native/runtime/ieee754.c:952-977 |
| 7 | `adamic_math_log` | internal/native/runtime/ieee754.c:1658-1737 |
| 6 | `adamic_math_log2` | internal/native/runtime/ieee754.c:1995-2073 |
| 6 | `adamic_math_log10` | internal/native/runtime/ieee754.c:2102-2140 |

### C the fixtures reach and the generator doesn't (top 20 functions)

| Lines | Function | Where |
|---|---|---|
| 149 | `regex_run` | internal/native/runtime/regexp.c:172-401 |
| 42 | `adamic_string_normalize` | internal/native/runtime/normalize.c:242-290 |
| 40 | `regex_substitution` | internal/native/runtime/regexp.c:829-917 |
| 32 | `substitution` | internal/native/runtime/string_replace_impl.h:48-83 |
| 31 | `power_of_two` | internal/native/runtime/parse.c:69-123 |
| 31 | `regex_clone` | internal/native/runtime/regexp.c:118-155 |
| 29 | `insert` | internal/native/runtime/weak.c:67-101 |
| 28 | `adamic_regex_match` | internal/native/runtime/regexp.c:673-700 |
| 27 | `compose` | internal/native/runtime/normalize.c:168-194 |
| 26 | `cased_beyond` | internal/native/runtime/case.c:134-160 |
| 22 | `adamic_map_clear` | internal/native/runtime/map.c:293-315 |
| 22 | `composite` | internal/native/runtime/normalize.c:79-100 |
| 22 | `encode` | internal/native/runtime/normalize.c:214-235 |
| 21 | `adamic_regex_iterator_step` | internal/native/runtime/regexp.c:719-739 |
| 21 | `adamic_math_atanh` | internal/native/runtime/ieee754.c:1576-1605 |
| 20 | `adamic_map_entries` | internal/native/runtime/map.c:271-291 |
| 19 | `adamic_write_text_file` | internal/native/runtime/input.c:322-349 |
| 19 | `reorder` | internal/native/runtime/normalize.c:145-163 |
| 19 | `decompose` | internal/native/runtime/normalize.c:124-142 |
| 17 | `adamic_union_typeof` | internal/native/runtime/union.c:62-78 |

## (a) The blind map: what no generated program executes

### Go, internal/lower and internal/native (top 40 of 1366 untouched regions, by statements)

A function no program enters is one region; inside a function that ran, a region is a run of blocks that never ran, cut at each `case` so every switch arm stands alone. The First line column names the case a region opens, or the case it sits inside.

Left out of the ranking, as unreachable by construction (23 regions, 571 statements): the fuzzer compiles through `adamic c` and `adamic js` and runs clang itself, so native.Build and Flags, the split build (units.go, units_tsgo.go), the runtime cache (library.go) and tsgo (tsgo.go in native and lower, which `adamic c` refuses) never run in the measured binary. REPORT.json lists them under blindMap.goStructural.

The last column is how many of the region's statements the oracle fixtures run; 0 means nothing in the fixtures reaches it either.

| # | Statements | Kind | Function | Where | First line | Fixtures run |
|---|---|---|---|---|---|---|
| 1 | 120 | function never entered | `lowering.staticInstance` | internal/lower/class_static.go:48-212 |  | 103 |
| 2 | 89 | function never entered | `lowering.staticInitializationReads` | internal/lower/class_static.go:309-436 |  | 61 |
| 3 | 84 | function never entered | `lowering.optionalRelationFailure` | internal/lower/proven_relations.go:33-163 |  | 56 |
| 4 | 72 | blocks never run | `lowering.librarySetMethod` | internal/lower/library_map_set.go:19-106 |  | 66 |
| 5 | 68 | function never entered | `lowering.collectIteration` | internal/lower/iteration_consume.go:11-96 |  | 58 |
| 6 | 62 | blocks never run | `lowering.libraryArrayForOf` | internal/lower/library_array.go:499-579 |  | 49 |
| 7 | 58 | blocks never run | `lowering.staticDeclaration` | internal/lower/class_static.go:222-296 |  | 53 |
| 8 | 54 | function never entered | `lowering.prototypeHazard` | internal/lower/prototype.go:176-256 |  | 34 |
| 9 | 53 | function never entered | `lowering.libraryStringMethod` | internal/lower/library_string.go:151-230 |  | 44 |
| 10 | 52 | function never entered | `lowering.nonObjectOwnProperty` | internal/lower/prototype_own.go:11-87 |  | 46 |
| 11 | 45 | function never entered | `lowering.libraryArrayCopyWithin` | internal/lower/library_array.go:262-311 |  | 42 |
| 12 | 45 | function never entered | `lowering.libraryArrayFlatMap` | internal/lower/library_array.go:348-408 |  | 37 |
| 13 | 45 | function never entered | `lowering.functionExpression` | internal/lower/library_function_expressions.go:10-70 |  | 36 |
| 14 | 44 | function never entered | `constructorSuperFlow.loop` | internal/lower/class_super.go:218-282 |  | 32 |
| 15 | 44 | function never entered | `lowering.provePredicate` | internal/lower/predicates.go:38-103 |  | 35 |
| 16 | 44 | blocks never run | `lowering.objectPrototypeCall` | internal/lower/prototype.go:102-170 |  | 34 |
| 17 | 43 | function never entered | `lowering.cast` | internal/lower/cast.go:15-75 |  | 34 |
| 18 | 41 | function never entered | `lowering.libraryArrayFlat` | internal/lower/library_array.go:205-260 |  | 34 |
| 19 | 38 | function never entered | `constructorSuperFlow.statement` | internal/lower/class_super.go:86-149 |  | 26 |
| 20 | 38 | function never entered | `lowering.functionValue` | internal/lower/expression.go:947-992 |  | 35 |
| 21 | 37 | function never entered | `lowering.stringRaw` | internal/lower/library_string.go:280-329 |  | 31 |
| 22 | 36 | function never entered | `predicateProof.statement` | internal/lower/predicates.go:172-231 |  | 20 |
| 23 | 35 | function never entered | `lowering.libraryForEachThisArg` | internal/lower/library_map_set.go:162-211 |  | 28 |
| 24 | 32 | blocks never run | `lowering.classGenericCall` | internal/lower/class_generic_calls.go:31-70 |  | 28 |
| 25 | 32 | function never entered | `lowering.updateIndex` | internal/lower/object.go:1889-1936 |  | 24 |
| 26 | 31 | blocks never run | `uniformFieldOffsets` | internal/native/fields.go:38-89 |  | 31 |
| 27 | 29 | function never entered | `lowering.pairTypes` | internal/lower/collections.go:145-184 |  | 19 |
| 28 | 29 | function never entered | `lowering.libraryArrayJoin` | internal/lower/library_array.go:591-630 |  | 23 |
| 29 | 29 | function never entered | `lowering.spreadNumbers` | internal/lower/object.go:1770-1811 |  | 23 |
| 30 | 29 | function never entered | `predicateProof.returned` | internal/lower/predicates.go:125-170 |  | 23 |
| 31 | 28 | blocks never run | `lowering.libraryArrayMethod` | internal/lower/library_array.go:33-79 | `in case "join":` | 21 |
| 32 | 26 | blocks never run | `lowering.userMethodCall` | internal/lower/iteration.go:434-466 |  | 18 |
| 33 | 26 | function never entered | `lowering.libraryForOfIterator` | internal/lower/library_map_set.go:261-298 |  | 9 |
| 34 | 24 | blocks never run | `lowering.input` | internal/lower/input.go:33-66 | `in case l.isPreludeFunction(callee, "writeTextFile"):` | 17 |
| 35 | 24 | function never entered | `lowering.libraryArraySpliced` | internal/lower/library_array.go:168-203 |  | 20 |
| 36 | 24 | function never entered | `lowering.libraryArrayWith` | internal/lower/library_array.go:313-346 |  | 20 |
| 37 | 24 | blocks never run | `lowering.newExpression` | internal/lower/object.go:1315-1348 |  | 19 |
| 38 | 23 | blocks never run | `lowering.conditionalSuper` | internal/lower/class_super.go:184-213 |  | 18 |
| 39 | 23 | function never entered | `constructorSuperFlow.switchCases` | internal/lower/class_super.go:284-314 |  | 17 |
| 40 | 23 | function never entered | `lowering.collectionPart` | internal/lower/collections.go:86-120 |  | 0 |

### Go, internal/native/emit*.go alone (top 40 of 94 untouched regions, by statements)

The emitter's regions are small, so few reach the table above; here they are on their own.

| # | Statements | Kind | Function | Where | First line | Fixtures run |
|---|---|---|---|---|---|---|
| 1 | 13 | blocks never run | `emitter.evaluate` | internal/native/emit_expressions.go:257-271 | `in case ir.CallClosure:` | 13 |
| 2 | 13 | blocks never run | `emitter.evaluate` | internal/native/emit_expressions.go:320-335 | `in case ir.ArrayFill:` | 13 |
| 3 | 13 | blocks never run | `emitter.evaluate` | internal/native/emit_expressions.go:351-365 | `in case ir.CheckedCast:` | 13 |
| 4 | 9 | case never taken | `emitter.evaluate` | internal/native/emit_expressions.go:272-286 | `case expression.Returns == 0:` | 9 |
| 5 | 8 | function never entered | `cBytes` | internal/native/emit.go:108-120 |  | 8 |
| 6 | 8 | blocks never run | `emitter.coalesce` | internal/native/emit_branches.go:130-138 |  | 7 |
| 7 | 7 | blocks never run | `emitter.returnStatement` | internal/native/emit_functions.go:95-103 |  | 6 |
| 8 | 6 | blocks never run | `emitter.evaluate` | internal/native/emit_expressions.go:196-206 | `in case ir.WeakOf:` | 6 |
| 9 | 5 | function never entered | `emitter.comparator` | internal/native/emit_arrays.go:202-210 |  | 5 |
| 10 | 5 | blocks never run | `emitter.evaluate` | internal/native/emit_expressions.go:420-426 | `in case ir.ArraySort:` | 5 |
| 11 | 5 | blocks never run | `emitter.evaluate` | internal/native/emit_expressions.go:513-520 | `in case ir.ProgramArguments:` | 3 |
| 12 | 5 | blocks never run | `emitter.arguments` | internal/native/emit_functions.go:187-192 |  | 5 |
| 13 | 4 | blocks never run | `equality` | internal/native/emit_arrays.go:178-184 | `in case ir.Boolean:` | 4 |
| 14 | 4 | blocks never run | `emitter.evaluate` | internal/native/emit_expressions.go:455-460 | `in case ir.MapKeys:` | 2 |
| 15 | 4 | blocks never run | `emitter.mathCall` | internal/native/emit_numbers.go:69-74 | `in case 0:` | 3 |
| 16 | 3 | blocks never run | `emitter.evaluate` | internal/native/emit_expressions.go:23-25 | `in case ir.RegExpGroup:` | 3 |
| 17 | 3 | blocks never run | `emitter.evaluate` | internal/native/emit_expressions.go:524-526 | `in case ir.WriteTextFile:` | 3 |
| 18 | 3 | blocks never run | `emitter.mapForEach` | internal/native/emit_maps.go:59-62 |  | 3 |
| 19 | 3 | blocks never run | `emitter.fieldSlot` | internal/native/emit_objects.go:26-29 |  | 3 |
| 20 | 3 | blocks never run | `emitter.statement` | internal/native/emit_statements.go:168-170 | `in case ir.Panic:` | 3 |
| 21 | 2 | blocks never run | `C` | internal/native/emit.go:34-35 |  | 2 |
| 22 | 2 | blocks never run | `emitter.arrayVisit` | internal/native/emit_arrays.go:86-88 | `in case "filter":` | 2 |
| 23 | 2 | blocks never run | `emitter.arrayReduce` | internal/native/emit_arrays.go:162-164 |  | 2 |
| 24 | 2 | blocks never run | `emitter.evaluate` | internal/native/emit_expressions.go:32-34 | `in case ir.IsNull:` | 2 |
| 25 | 2 | blocks never run | `emitter.evaluate` | internal/native/emit_expressions.go:93-95 | `in case ir.InstanceOf:` | 0 |
| 26 | 2 | blocks never run | `emitter.evaluate` | internal/native/emit_expressions.go:225-227 | `in case ir.StringLength:` | 2 |
| 27 | 2 | blocks never run | `emitter.evaluate` | internal/native/emit_expressions.go:398-399 | `in case ir.MapEntries:` | 2 |
| 28 | 2 | blocks never run | `emitter.evaluate` | internal/native/emit_expressions.go:580-582 | `in case ir.Length:` | 2 |
| 29 | 2 | blocks never run | `emitter.binary` | internal/native/emit_expressions.go:674-676 | `in case operandType == ir.Union && operator == ir.Equal:` | 2 |
| 30 | 2 | blocks never run | `emitter.objectLiteral` | internal/native/emit_objects.go:123-124 |  | 2 |
| 31 | 2 | blocks never run | `emitter.forOf` | internal/native/emit_statements.go:307-309 |  | 2 |
| 32 | 2 | blocks never run | `cType` | internal/native/emit_values.go:28-30 | `in case ir.Union:` | 2 |
| 33 | 1 | blocks never run | `emitter.arrayVisit` | internal/native/emit_arrays.go:32-33 | `in case "find", "findLast":` | 1 |
| 34 | 1 | blocks never run | `emitter.arrayVisit` | internal/native/emit_arrays.go:38-39 | `in case "find", "findLast":` | 1 |
| 35 | 1 | blocks never run | `emitter.arrayVisit` | internal/native/emit_arrays.go:56-57 | `in case "find", "findLast":` | 1 |
| 36 | 1 | blocks never run | `emitter.arrayVisit` | internal/native/emit_arrays.go:66-67 | `in case "find", "findLast":` | 1 |
| 37 | 1 | blocks never run | `emitter.arrayVisit` | internal/native/emit_arrays.go:72-73 | `in case "find", "findLast":` | 1 |
| 38 | 1 | blocks never run | `emitter.arrayVisit` | internal/native/emit_arrays.go:79-80 | `in case "forEach":` | 1 |
| 39 | 1 | blocks never run | `emitter.arrayVisit` | internal/native/emit_arrays.go:101-102 | `in case "find", "findLast":` | 1 |
| 40 | 1 | blocks never run | `emitter.arrayReduce` | internal/native/emit_arrays.go:135-136 |  | 1 |

### C, internal/native/runtime (top 40 of 487 untouched regions, by lines that never ran)

A function no program calls is one region; inside a function that ran, a region is an outermost code region (a branch's arm, a loop body, the rest of a function after a return) that never ran.

The last column is how many of the region's lines the oracle fixtures run.

| # | Lines | Kind | Function | Where | First line | Fixtures run |
|---|---|---|---|---|---|---|
| 1 | 63 | function never called | `adamic_read_directory` | internal/native/runtime/directory.c:67-130 |  | 0 |
| 2 | 49 | region never run | `regex_run` | internal/native/runtime/regexp.c:329-377 | `case 9: {` | 39 |
| 3 | 47 | function never called | `adamic_string_normalize` | internal/native/runtime/normalize.c:242-290 |  | 42 |
| 4 | 44 | function never called | `adamic_record_keys` | internal/native/runtime/record.c:149-192 |  | 0 |
| 5 | 40 | function never called | `adamic_plain_object_keys` | internal/native/runtime/library_language.c:28-70 |  | 0 |
| 6 | 39 | region never run | `regex_substitution` | internal/native/runtime/regexp.c:855-906 | `case '<': {` | 36 |
| 7 | 39 | function never called | `adamic_utf8_at` | internal/native/runtime/utf8.c:18-60 |  | 16 |
| 8 | 37 | function never called | `regex_clone` | internal/native/runtime/regexp.c:118-155 |  | 31 |
| 9 | 34 | function never called | `insert` | internal/native/runtime/weak.c:67-101 |  | 29 |
| 10 | 32 | region never run | `substitution` | internal/native/runtime/string_replace_impl.h:52-83 | `pieces list = {NULL, 0, 0};` | 32 |
| 11 | 31 | function never called | `new_chunk` | internal/native/runtime/heap.c:107-137 |  | 0 |
| 12 | 30 | region never run | `power_of_two` | internal/native/runtime/parse.c:85-116 | `if (overflow != 0) {` | 30 |
| 13 | 28 | function never called | `adamic_math_atanh` | internal/native/runtime/ieee754.c:1576-1605 |  | 21 |
| 14 | 28 | function never called | `adamic_regex_match` | internal/native/runtime/regexp.c:673-700 |  | 28 |
| 15 | 27 | function never called | `adamic_write_text_file` | internal/native/runtime/input.c:322-349 |  | 19 |
| 16 | 27 | function never called | `compose` | internal/native/runtime/normalize.c:168-194 |  | 27 |
| 17 | 27 | region never run | `regex_run` | internal/native/runtime/regexp.c:298-324 | `case 7: {` | 27 |
| 18 | 26 | function never called | `cased_beyond` | internal/native/runtime/case.c:134-160 |  | 26 |
| 19 | 26 | function never called | `prototype_member` | internal/native/runtime/record.c:43-68 |  | 0 |
| 20 | 24 | function never called | `take` | internal/native/runtime/heap.c:139-162 |  | 0 |
| 21 | 22 | function never called | `adamic_string_from_code_points` | internal/native/runtime/from_codes.c:90-113 |  | 9 |
| 22 | 22 | region never run | `kernel_rem_pio2` | internal/native/runtime/ieee754.c:627-656 | `case 1:` | 0 |
| 23 | 22 | function never called | `adamic_map_clear` | internal/native/runtime/map.c:293-315 |  | 22 |
| 24 | 22 | function never called | `composite` | internal/native/runtime/normalize.c:79-100 |  | 22 |
| 25 | 22 | function never called | `encode` | internal/native/runtime/normalize.c:214-235 |  | 22 |
| 26 | 21 | function never called | `adamic_regex_iterator_step` | internal/native/runtime/regexp.c:719-739 |  | 21 |
| 27 | 20 | function never called | `adamic_map_entries` | internal/native/runtime/map.c:271-291 |  | 20 |
| 28 | 19 | function never called | `decompose` | internal/native/runtime/normalize.c:124-142 |  | 19 |
| 29 | 19 | function never called | `reorder` | internal/native/runtime/normalize.c:145-163 |  | 19 |
| 30 | 18 | function never called | `array_index` | internal/native/runtime/record.c:119-136 |  | 0 |
| 31 | 18 | region never run | `regex_run` | internal/native/runtime/regexp.c:266-283 | `case 5:` | 18 |
| 32 | 17 | function never called | `adamic_uncaught` | internal/native/runtime/exceptions.c:24-42 |  | 0 |
| 33 | 17 | function never called | `give` | internal/native/runtime/heap.c:164-186 |  | 17 |
| 34 | 17 | function never called | `array_index` | internal/native/runtime/library_language.c:10-26 |  | 0 |
| 35 | 17 | function never called | `library_uint32` | internal/native/runtime/library_math_number.c:12-28 |  | 17 |
| 36 | 17 | region never run | `regex_run` | internal/native/runtime/regexp.c:218-235 | `for (size_t k = 0; k < i->string_count; k++) {` | 17 |
| 37 | 17 | function never called | `adamic_union_to_string` | internal/native/runtime/union.c:42-60 |  | 13 |
| 38 | 17 | function never called | `adamic_union_typeof` | internal/native/runtime/union.c:62-78 |  | 17 |
| 39 | 16 | function never called | `adamic_array_join_nested` | internal/native/runtime/library_array.c:34-49 |  | 16 |
| 40 | 16 | function never called | `listed` | internal/native/runtime/map.c:319-334 |  | 0 |

## How it was measured

Run by `verify/coverage/measure.sh` at b2bf06e25cd61b5c2d0c9c0ea471c4f0120f9112+dirty. Reproduce with `verify/coverage/measure.sh <fresh directory>`
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
