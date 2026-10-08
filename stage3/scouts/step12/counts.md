# Scout witness counts

12 witnesses, 12 source mutants caught, two native-equal controls, ten expected
blocked builds. These are local scout fixtures, not internal/oracle registrations.

| Witness | Node stdout | Build | Input mutation and catcher |
| --- | --- | --- | --- |
| namespace.a | `ok` | expected stop | `Debug.assertEqual(1, 1)` to `Debug.assertEqual(1, 2)`; Node exception/exit |
| mutable-namespace.a | `true` | expected stop | `Debug.isDebugging = true` to `Debug.isDebugging = false`; Node stdout difference |
| diagnostic-cast.a | `1002` | native equal | `diag(1002)` to `diag(1003)`; Node stdout difference |
| string-cast.a | `2 55296` | expected stop | `Worker(0x10000)` to `Worker(0xffff)`; Node stdout difference |
| set-cast.a | `true / true` | expected stop | `new Set(["Latin"])` to `new Set(["Greek"])`; Node stdout difference |
| enum-slot.a | `99` | native equal | `Math.trunc(99)` to `Math.trunc(100)`; Node stdout difference |
| array-holes.a | `4 0 undefined / 3` | expected stop | `i <= s2.length` to `i < s2.length`; Node stdout difference |
| entries.a | `number` | expected stop | `extra: 1` to `extra: "lie"`; Node stdout difference |
| computed-field.a | `constructor` | expected stop | `"" + "constructor"` to `"" + "constructoX"`; Node stdout difference |
| uint16.a | `0 1` | expected stop | `shiftedDigit = 0x10000` to `shiftedDigit = 0xffff`; Node stdout difference |
| generic-return.a | `1` | expected stop | `[["x", 1]]` to `[["x", 0]]`; Node stdout difference |
| error-cast.a | `function` | expected stop | `(Error as any)` to `(String as any)`; Node stdout difference |

Census: 89 code declarations, eight code files, 78 evaluation modules, one namespace,
one mutable namespace export, two Array(length) calls, three Object.entries calls,
one computed name, one Uint16Array call and 16 generic signatures. Assertion syntax:
2,136 expressions, including 2,130 same-type generated diagnostic assertions and
one as const. Five unchecked expressions form Error/String/Set families. The exact
claimed enum-slot count is unavailable, not zero. See RESEARCH.md for qualifications.
