| ID | origin/main file:line | Change | Failed grouped rows |
|---|---|---|---|
| M01 | internal/native/runtime/input.c:47 | `lead >= 0xc2 -> lead >= 0xc3` | TestDecodeASCII family |
| M02 | internal/native/runtime/input.c:56 | `upper = 0x9f; -> upper = 0xbf;` | TestDecodeASCII family |
| M03 | internal/native/runtime/input.c:129 | `string->units = length + 1; -> string->units = length + 2;` | TestDecodeASCII family |
| M04 | internal/native/runtime/input.c:111 | `bytes[offset] < 0x80 -> bytes[offset] <= 0x80` | TestDecodeASCII family |
| M05 | internal/native/runtime/dtoa.c:975 | `int exponent = decimal_point - 1; -> int exponent = decimal_point;` | TestToExponentialAndToPrecisionMatchNode, TestToExponentialAndToPrecisionOutOfRangePanic |
| M06 | internal/native/runtime/dtoa.c:992 | `exponent < -6 -> exponent < -5` | TestToExponentialAndToPrecisionMatchNode |
| M07 | internal/native/runtime/dtoa.c:1054 | `digits < 0 || -> digits < -1 ||` | TestToExponentialAndToPrecisionOutOfRangePanic |
| M08 | internal/native/runtime/dtoa.c:1072 | `digits < 1 || -> digits < 0 ||` | TestToExponentialAndToPrecisionOutOfRangePanic |
| M09 | internal/native/devirtualize.go:18 | `!class.Static && -> class.Static &&` | TestDevirtualizedCalls |
| M10 | internal/native/devirtualize.go:49 | `declarations == 1 && !written -> declarations == 1 && written` | TestDevirtualizedCalls, TestExactReceiverRejectsAssignments |
| M11 | internal/native/class_inheritance.go:48 | `len(targets) == 1 -> len(targets) == 0` | TestDevirtualizedCalls |
| M12 | internal/native/devirtualize.go:68 | `method.Name == name -> method.Name != name` | TestDevirtualizedCalls |
