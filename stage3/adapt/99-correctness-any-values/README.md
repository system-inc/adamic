# Family 12 value types

The ten bindings in the approved #kbcj5nr table use the types already supplied by their producers or held at their reads. The uninitialized locals retain their initializer and assignment order. Three readers accept values they already narrow, and the config-directory predicate takes the indexed option type it already receives.

The planner matches each declaration by file, function, AST kind and binding name, then accepts its exact original or adapted snapshot. It checks all files before writing, rejects a partially adapted batch and compares stock TypeScript 6.0.3 JavaScript on each changed file. The normal entry point is the full `stage3/apply.sh`.

The retained candidate guards config own-key input inline and guards file-path strings, and holds the source-map parse result as unknown before the existing reader. The companion sealed counter measures the retained runtime checks before acceptance.

Round 4 leaves two original readers in place. The raw-config reader needs a mutable view with both dictionary storage and the optional compileOnSave field. The source-map reader needs field-level checks on an unknown object without changing the existing field-read order. The Round 3 versions of those readers added census stops. Inline own-key narrowing reaches the existing own-property call after clearing the earlier any stop.

Round 5 defers `compareDataObjects` and preserves the upstream function. Preboxed targets fall outside Adamic's admitted `Reflect.get` shape. The admitted inline shape also stops on primitive receivers in both backends. Resuming this site needs an approved representation that preserves primitive getter receivers and the order of repeated reads.

### Historical Round 2 first-error scope

Each unit below already had a first stop that cleared. The table uses the meter’s exact reason as its family label. A bold label marks a finding that remains an addition after position stripping.

| Site | Unit | Meter family and reason |
| --- | --- | --- |
| `compiler/commandLineParser.ts:3878:102` | `compiler/commandLineParser.ts:1885:1`, `compiler/commandLineParser.ts:3882:1` | **a call spreading something other than an array** |
| `compiler/commandLineParser.ts:2624:13` | `compiler/commandLineParser.ts:2621:1` | a field read through a union whose members store .type in different representations (CommandLineOptionOfBooleanType as string, CommandLineOptionOfCustomType as value) |
| `compiler/commandLineParser.ts:3009:13` | `compiler/commandLineParser.ts:3007:1` | a field read through a union whose members store .type in different representations (CommandLineOptionOfBooleanType as string, CommandLineOptionOfCustomType as value) |
| `compiler/commandLineParser.ts:4291:17` | `compiler/commandLineParser.ts:4285:1` | **an unrestricted dictionary write that can violate named property all; use a literal key or a value accepted by every named property** |
| `compiler/commandLineParser.ts:4300:13` | `compiler/commandLineParser.ts:4298:1` | a field read through a union whose members store .type in different representations (CommandLineOptionOfBooleanType as string, CommandLineOptionOfCustomType as value) |
| `compiler/moduleNameResolver.ts:1859:9` | `compiler/moduleNameResolver.ts:1803:1` | Family 12. a value of type any: any has no proven type |
| `compiler/moduleSpecifiers.ts:800:25` | `compiler/moduleSpecifiers.ts:798:1` | **for...of over an object** |
| `compiler/program.ts:4977:13` | `compiler/program.ts:4902:1` | Family 21, inherited library members read as own fields. **inherited library member call read as an own field** |

### Historical Round 3 replay notes

This is the Round 3 replay scope, with its archived offsets and abandoned comparison candidate. Use the paired full-mode inventory below for current caller attribution.

| Site | Unit | Meter family and reason |
| --- | --- | --- |
| `compiler/commandLineParser.ts:3571:30` | `compiler/commandLineParser.ts:3546:1` | an array of unknown with erased element storage (retain its declared element type before reading elements) |
| `compiler/commandLineParser.ts:2621:1` | `compiler/commandLineParser.ts:3596:1` | Unlisted on Nov 21. Route to @system_adamic. a checked predicate target whose base has no required field to check (MapLike<string[]>) |
| `compiler/sourcemap.ts:432:33` | `compiler/sourcemap.ts:430:1` | JSON.parse: its result's type can't be proven from the text |
| `compiler/utilities.ts:7809:17` | `compiler/utilities.ts:7785:1` | a function returning any |
| `compiler/utilities.ts:8148:25` | `compiler/utilities.ts:8147:1` | Unlisted on Nov 21. Route to @system_adamic. Object.keys of a value that may hold an Error, whose own properties Node doesn't enumerate |


### Historical Round 4 first-error scope

| Site | Unit | Meter family and reason |
| --- | --- | --- |
| `compiler/core.ts:1275:12` | `compiler/commandLineParser.ts:3372:1` | hasOwnProperty through an object view (an Error may be hidden by the view) |

### Paired full-mode exposure inventory

These rows come from the pinned full-mode selection in `exposure-selection.json`. `replay-exposures.sh` collects the baseline and candidate records and checks this table. Both MapLike callers lose an earlier any refusal. The November 21 meter omits that predicate reason. Route any first-error instance to @system_adamic. The first-error census remains the daily inventory.

| Site | Unit | Verdict | Reason | November 21 meter |
| --- | --- | --- | --- | --- |
| `compiler/commandLineParser.ts:2624:13` | `compiler/commandLineParser.ts:2621:1` | NotYet | a field read through a union whose members store .type in different representations (CommandLineOptionOfBooleanType as string, CommandLineOptionOfCustomType as value) | Exact match |
| `compiler/commandLineParser.ts:2627:13` | `compiler/commandLineParser.ts:2621:1` | NotYet | a field read through a union whose members store .type in different representations (CommandLineOptionOfBooleanType as string, CommandLineOptionOfCustomType as value) | Exact match |
| `compiler/commandLineParser.ts:2630:39` | `compiler/commandLineParser.ts:2621:1` | NotYet | a field read through a union whose members store .type in different representations (CommandLineOptionOfBooleanType as string, CommandLineOptionOfCustomType as value) | Exact match |
| `compiler/commandLineParser.ts:2631:33` | `compiler/commandLineParser.ts:2621:1` | NotYet | reading expectedType | Exact reason absent |
| `compiler/commandLineParser.ts:3634:21` | `compiler/commandLineParser.ts:3623:5` | Refused | a value of type any: any has no proven type | Exact match |
| `compiler/commandLineParser.ts:3635:66` | `compiler/commandLineParser.ts:3623:5` | NotYet | assigning to an Identifier | Exact reason absent |
| `compiler/commandLineParser.ts:3636:68` | `compiler/commandLineParser.ts:3623:5` | NotYet | assigning to an Identifier | Exact reason absent |
| `compiler/commandLineParser.ts:3637:71` | `compiler/commandLineParser.ts:3623:5` | NotYet | assigning to an Identifier | Exact reason absent |
| `compiler/commandLineParser.ts:3640:17` | `compiler/commandLineParser.ts:3623:5` | NotYet | reading currentOption | Exact reason absent |
| `compiler/commandLineParser.ts:2621:1` | `compiler/commandLineParser.ts:3623:5` | NotYet | a checked predicate target whose base has no required field to check (MapLike<string[]>) | Exact reason absent · @system_adamic |
| `compiler/commandLineParser.ts:2621:1` | `compiler/commandLineParser.ts:3543:1` | NotYet | a checked predicate target whose base has no required field to check (MapLike<string[]>) | Exact reason absent · @system_adamic |
| `compiler/program.ts:4977:13` | `compiler/program.ts:4902:1` | Refused | inherited library member call read as an own field | Family 21, member variant |
| `compiler/program.ts:4988:39` | `compiler/program.ts:4902:1` | Refused | inherited library member call read as an own field | Family 21, member variant |
| `compiler/program.ts:4996:13` | `compiler/program.ts:4902:1` | NotYet | a Map or Set going where a plain object type is held | Exact match |
| `compiler/program.ts:5015:36` | `compiler/program.ts:4902:1` | Refused | inherited library member call read as an own field | Family 21, member variant |
| `compiler/program.ts:5018:13` | `compiler/program.ts:4902:1` | NotYet | reading real | Exact reason absent |
| `compiler/program.ts:5027:52` | `compiler/program.ts:4902:1` | NotYet | reading real | Exact reason absent |
| `compiler/program.ts:5058:86` | `compiler/program.ts:4902:1` | NotYet | RegExp with a nonconstant pattern | Exact match |
| `compiler/program.ts:4919:27` | `compiler/program.ts:4902:1` | NotYet | an assignment value to a member | Exact match |
| `compiler/program.ts:4952:81` | `compiler/program.ts:4902:1` | Refused | inherited library member call read as an own field | Family 21, member variant |
| `compiler/program.ts:4960:13` | `compiler/program.ts:4902:1` | NotYet | an optional Map get through a structural receiver | Exact match |
| `compiler/utilities.ts:7809:17` | `compiler/utilities.ts:7785:1` | NotYet | a function returning any | Exact match |

### Replay limits

The pinned redirect replay reports `4919:27` at the member assignment. The review's fetched-main `4920:17` reading was not reproduced at the pin. The pinned strict-parser reason is `NotYet` “a function returning any” in the reader unit. The review reports a different verdict and reason on fetched main.

The reader guard also reports “reading result” at `utilities.ts:7799:36` in full mode. The baseline reader reports the same reason at `:7793:9` and `:7799:12`. Its first-error producer attribution is authoritative for the daily inventory. The comparison Object.keys row above is historical after that candidate was deferred.

### Round 7 comparison view

The sole call in `isProgramUptoDate` passes complete `CompilerOptions` objects. Their index signature includes `TsConfigSourceFile`, and `setConfigFileInOptions` installs that metadata non-enumerably. Stock TypeScript rejects either complete argument as `CompilerOptionsValue` or `AdamicJsonRecoveryValue`. The owner's fallback therefore needs a checked view of the enumerable options data at the caller.

A cast-free prototype test at the adapted caller produced a new `Refused` at `program.ts:1266:39` in unit `program.ts:1234:1`. The reason is `Object.getPrototypeOf`. Descriptor reflection also has an explicit refusal in the pinned compiler. The comparison stays deferred pending a supported way to check the enumerable data without replacing the objects or invoking accessors ahead of upstream. The sealed measurement observes the proposed data envelope on the original root and recursive calls.
