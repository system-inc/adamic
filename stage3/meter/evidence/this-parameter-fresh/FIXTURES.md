# Receiver freshness fixtures (#69vkn0j)

**Native divergence.** Both guarded f2 fixtures and string-wrapped f2.ts hit a null string member access under UBSan (exit 1). Node and the JavaScript backend raise a TypeError (exit 70).

Origin/main was `491e020de2aa138d7b608611c40a641571a5a1eb`. The runs used Node v24.19.0, Go 1.27.1 and Clang 20.1.8, with recursive submodules at the baseline's gitlinks. A temporary test used the oracle's source runner, `lowered`, `onJavaScriptBackend`, `nativelyUncached` and `disagreement`. The native build enabled address and undefined-behavior sanitizers. The comparison run disabled leak detection, as in the oracle, and a separate leak check for each f1 and f3 run was clean.

| Fixture | Engine | stdout | stderr | Exit |
| --- | --- | --- | --- | ---: |
| f1.a, f1.ts | Node, JavaScript, native | `"undefined\n"` | `""` | 0 |
| f2.a, f2.ts | Node | `""` | TypeError below | 70 |
| f2.a, f2.ts | JavaScript, native | — | Checker reports TS2345 at 5:13 | — |
| f3.a, f3.ts | Node, JavaScript, native | `"a\nundefined\n"` | `""` | 0 |
| f2-string.a | Node | `""` | TypeError below | 70 |
| f2-string.a | JavaScript, native | — | Lowering refuses `!` at 5:20 | — |
| f2-string.ts, f2-guarded.a, f2-guarded.ts | Node, JavaScript | `""` | TypeError below | 70 |
| f2-string.ts, f2-guarded.a, f2-guarded.ts | native | `""` | UBSan report below | 1 |

The original f2 passes a number to Adamic's string-only `console.log`. Its two backend executables could not be built. `f2-string` converts the result after evaluating the length. In `.a`, the existing non-null assertion rule then refuses the unproven `!`. `f2-guarded` uses a presence guard for the array element in both formats. Once that guard succeeds, it reads the same `tag.length` field.

TypeError stderr was `"adamic: panic: TypeError: Cannot read properties of undefined (reading 'length')\n"`.

UBSan stderr was `"adamic.h:1394:14: runtime error: member access within null pointer of type 'const adamic_string' (aka 'const struct adamic_string')\nSUMMARY: UndefinedBehaviorSanitizer: undefined-behavior adamic.h:1394:14 \n"`.

The checker diagnostic was `error TS2345: Argument of type 'number' is not assignable to parameter of type 'string'.` The f2-string.a diagnostic was `Adamic 0.1 refuses the non-null assertion !; write ?? panic('why it can't be missing'), or narrow and handle the missing case`.
