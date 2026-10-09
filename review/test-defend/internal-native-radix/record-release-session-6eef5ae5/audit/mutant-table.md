| ID | Origin file:line | Change | Failed production rows |
|---|---|---|---|
| M01 | internal/native/runtime/radix.c:69 | 0.5 * (radix_next_double(value) - value) → 1.0 * (radix_next_double(value) - value) | TestToStringWithARadixMatchesNode |
| M02 | internal/native/runtime/radix.c:151 | radix_result("0", 1) → radix_result("1", 1) | TestToStringWithARadixMatchesNode, TestToStringWithARadixOutOfRangePanics |
| M03 | internal/native/runtime/radix.c:143 | radix > 36 → radix > 35 | TestToStringWithARadixMatchesNode, TestToStringWithARadixOutOfRangePanics |
| M04 | internal/native/runtime/record.c:111 | return true; }  size_t adamic_record_size → return false; }  size_t adamic_record_size | TestRecordsAgainstNode |
| M05 | internal/native/runtime/record.c:132 | number >= UINT32_MAX → number > UINT32_MAX | TestRecordsAgainstNode |
| M06 | internal/native/runtime/record.c:147 | return a < b ? -1 : a > b ? 1 : 0; → return a < b ? 1 : a > b ? -1 : 0; | TestRecordsAgainstNode |
| M07 | internal/native/runtime/regexp.c:9 | regex_step_limit = limit; → regex_step_limit = 1; | TestRegExpBytecodePatternUnits, TestRegExpBytecodeRandomNode family, TestRegExpBytecodeTest262, TestRegExpLintPatternsNode, TestRegExpSearchNode |
| M08 | internal/native/runtime/regexp.c:82 | i->ranges[first].first <= c → i->ranges[first].first < c | TestRegExpBytecodePatternUnits, TestRegExpBytecodeRandomNode family, TestRegExpBytecodeTest262, TestRegExpLintPatternsNode, TestRegExpNativeStepLimit, TestRegExpSearchNode |
| M09 | internal/native/runtime/regexp.c:581 | strcmp(object->shape->names[k], "done") → strcmp(object->shape->names[k], "value") | TestRegExpIteratorResultShape |
| M10 | internal/native/runtime/string_build_impl.h:127 | memcmp(left->bytes, right->bytes, left->length) == 0 → memcmp(left->bytes, right->bytes, left->length) != 0 | TestRecordBenchmark, TestRecordsAgainstNode, TestRegExpBytecodeRandomNode family, TestRegExpBytecodeTest262, TestRegExpLintPatternsNode, TestRegExpSearchNode, TestRuntimeStringEquality |
| M11 | internal/native/runtime/heap.c:375 | release_last(value); → (void)value; | TestRecordsAgainstNode, TestRegExpBytecodePatternUnits, TestRegExpBytecodeRandomNode family, TestRegExpBytecodeTest262, TestRegExpIteratorResultShape, TestRegExpLintPatternsNode, TestRegExpSearchNode, TestRuntimeReleasePaths, TestRuntimeStringEquality, TestToStringWithARadixMatchesNode |
| M12 | internal/native/runtime/regexp.c:187 | *steps >= regex_step_limit → *steps > regex_step_limit |  |
