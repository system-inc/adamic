| ID | Origin location | Change | Failed rows |
|---|---|---|---|
| M01 | internal/native/runtime/math.c:15 | `value - floored >= 0.5` -> `value - floored > 0.5` | TestMathAndToFixedMatchJavaScript |
| M02 | internal/native/runtime/math.c:27 | `value > 0 ? 1 : -1` -> `value < 0 ? 1 : -1` | TestMathAndToFixedMatchJavaScript |
| M03 | internal/native/runtime/math.c:35 | `signbit(left) && signbit(right)` -> `signbit(left) \|\| signbit(right)` | TestMathAndToFixedMatchJavaScript |
| M04 | internal/native/runtime/math.c:47 | `left < right ? left : right` -> `left > right ? left : right` | TestMathAndToFixedMatchJavaScript |
| M05 | internal/native/runtime/number.c:117 | `exponent == 0.5` -> `exponent == 0.25` |  |
| M06 | internal/native/runtime/number.c:152 | `exact[keep] >= '5'` -> `exact[keep] > '5'` | TestMathAndToFixedMatchJavaScript |
| M07 | internal/native/runtime/number.c:144 | `bool negative = value < 0;` -> `bool negative = signbit(value);` | TestMathAndToFixedMatchJavaScript |
| M08 | internal/native/runtime/maybe.c:34 | `bits = 0x7ff8000000000000u;` -> `bits = 0x7ff8000000000001u;` | TestMaybeNumbersPackIntoOneDouble |
| M09 | internal/native/runtime/maybe.c:29 | `if (value.present) {` -> `if (!value.present) {` | TestMaybeNumbersPackIntoOneDouble |
| M10 | internal/native/runtime/maybe.c:44 | `bits == ADAMIC_UNDEFINED_BITS` -> `bits != ADAMIC_UNDEFINED_BITS` | TestMaybeNumbersPackIntoOneDouble |
| M11 | internal/native/runtime/normalize.c:165 | `if (mapping == NULL \|\| (mapping->compatibility && !compatibility)) {` -> `if (mapping == NULL \|\| (mapping->compatibility && compatibility)) {` | TestNormalizeRandomMatchesNode, TestNormalizeMatchesNode family |
| M12 | internal/native/runtime/normalize.c:574 | `compatibility = true, composed = true;` -> `compatibility = false, composed = true;` | TestNormalizeRandomMatchesNode, TestNormalizeMatchesNode family |
| M13 | internal/native/runtime/normalize.c:572 | `compatibility = false, composed = false;` -> `compatibility = false, composed = true;` | TestNormalizeRandomMatchesNode, TestNormalizeMatchesNode family |
| M14 | internal/native/runtime/node_buffer.c:132 | `return (int)(c - 'a') + 26;` -> `return (int)(c - 'a') + 25;` | TestNodeBufferRuntimeWithoutDeclarations |
| M15 | internal/native/runtime/node_crypto.c:93 | `hash->slots[1].number = 1;` -> `/* dropped finalized flag assignment */` |  |
