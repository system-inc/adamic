# Family 62, diagnostic append callbacks

This replaces the method-only views introduced by `71-writable-views` on
`convertConfigFileToObject` and `convertToJson`. It covers the two
`commandLineParser.ts` sites reported at 2281:59 and 2483:76 by the scout.
The Nov 22 ruling on #bbssdcy authorizes the JavaScript change. Rulings 89
and 95 govern represented arrays and mutable slots, and 126 covers element
checks for Array writes. Rulings 400 and 401 govern the public declaration
boundary. The public `convertToObject` signature still takes `Diagnostic[]`.

## Value identity

Each external caller passes a closure with the same receiver and returns the
result of `push`. Both callers using a source file's `parseDiagnostics`
read the same data field on their newly parsed file. Each takes a local
copy of that array at the call, so later appends continue to use the original
receiver. The parser copies the root expression first to preserve argument
evaluation order. Those calls supply no notifier. The config worker passes
the callback onward in the two branches that formerly forwarded the same
receiver. Each append calls `push` once on the same array with the same
diagnostic, in the same order. The `push` lookup now follows diagnostic
construction. Observing that difference requires caller-defined accessors
where the declarations describe a method or data field. Adamic's Nov 23
answer on #bbssdcy puts those accessors outside the contract for tsc's own
producers. All seven calls discard the result. There are no other reads,
stores, comparisons or escapes of the converted parameters. In
`parseConfigFileTextToJson`, the first diagnostic is still read after the
conversion.

The guard checks the exact old `errors` references. Across each entire
consumer declaration, it also requires zero `appendDiagnostic` identifiers
before the edit and exactly the identifiers inside the pinned nodes after
the edit. The `d` in each callback is bound by that arrow's own innermost
parameter. The new `parseDiagnostics` and `rootExpression` locals and their
uses are inside the exact whole-block pins for `parseConfigFileTextToJson`
and `parseSourceFile`. Inserting code between those declarations and uses
changes the pinned block text and refuses the input.

The checker trace below resolves both declaration symbols and aliases in a
TypeScript 6.0.3 Program over the adapted `src`. It also visits every
reference to each parameter symbol, including uses in nested functions.

```jsonl
{"kind":"binding","name":"convertToJson","at":"src/compiler/parser.ts:54:5","text":"convertToJson"}
{"kind":"caller","name":"convertToJson","at":"src/compiler/parser.ts:1589:13","text":"convertToJson(result, result.statements[0]?.expression, result.parseDiagnostics, /*returnValue*/ false, /*jsonConversionNotifier*/ undefined)"}
{"kind":"caller","name":"convertConfigFileToObject","at":"src/compiler/commandLineParser.ts:2281:17","text":"convertConfigFileToObject(jsonSourceFile, jsonSourceFile.parseDiagnostics, /*jsonConversionNotifier*/ undefined)"}
{"kind":"binding","name":"convertConfigFileToObject","at":"src/compiler/commandLineParser.ts:2452:10","text":"convertConfigFileToObject"}
{"kind":"unused one-argument push","name":"convertConfigFileToObject","at":"src/compiler/commandLineParser.ts:2459:9","text":"errors.push(createDiagnosticForNodeInSourceFile(\r\n            sourceFile,\r\n            rootExpression,\r\n            Diagnostics.The_root_value_of_a_0_file_must_be_an_object,\r\n            getBaseFileName(sourceFile.fileName) === \"jsconfig.json\" ? \"jsconfig.json\" : \"tsconfig.json\",\r\n        ))"}
{"kind":"caller","name":"convertToJson","at":"src/compiler/commandLineParser.ts:2471:24","text":"convertToJson(sourceFile, firstObject, errors, /*returnValue*/ true, jsonConversionNotifier)"}
{"kind":"forward to convertToJson","name":"convertConfigFileToObject","at":"src/compiler/commandLineParser.ts:2471:24","text":"convertToJson(sourceFile, firstObject, errors, /*returnValue*/ true, jsonConversionNotifier)"}
{"kind":"caller","name":"convertToJson","at":"src/compiler/commandLineParser.ts:2476:12","text":"convertToJson(sourceFile, rootExpression, errors, /*returnValue*/ true, jsonConversionNotifier)"}
{"kind":"forward to convertToJson","name":"convertConfigFileToObject","at":"src/compiler/commandLineParser.ts:2476:12","text":"convertToJson(sourceFile, rootExpression, errors, /*returnValue*/ true, jsonConversionNotifier)"}
{"kind":"caller","name":"convertToJson","at":"src/compiler/commandLineParser.ts:2483:12","text":"convertToJson(sourceFile, sourceFile.statements[0]?.expression, errors, /*returnValue*/ true, /*jsonConversionNotifier*/ undefined)"}
{"kind":"binding","name":"convertToJson","at":"src/compiler/commandLineParser.ts:2493:17","text":"convertToJson"}
{"kind":"unused one-argument push","name":"convertToJson","at":"src/compiler/commandLineParser.ts:2513:17","text":"errors.push(createDiagnosticForNodeInSourceFile(sourceFile, element, Diagnostics.Property_assignment_expected))"}
{"kind":"unused one-argument push","name":"convertToJson","at":"src/compiler/commandLineParser.ts:2518:17","text":"errors.push(createDiagnosticForNodeInSourceFile(sourceFile, element.questionToken, Diagnostics.The_0_modifier_can_only_be_used_in_TypeScript_files, \"?\"))"}
{"kind":"unused one-argument push","name":"convertToJson","at":"src/compiler/commandLineParser.ts:2521:17","text":"errors.push(createDiagnosticForNodeInSourceFile(sourceFile, element.name, Diagnostics.String_literal_with_double_quotes_expected))"}
{"kind":"unused one-argument push","name":"convertToJson","at":"src/compiler/commandLineParser.ts:2566:21","text":"errors.push(createDiagnosticForNodeInSourceFile(sourceFile, valueExpression, Diagnostics.String_literal_with_double_quotes_expected))"}
{"kind":"unused one-argument push","name":"convertToJson","at":"src/compiler/commandLineParser.ts:2599:13","text":"errors.push(createDiagnosticForNodeInSourceFile(sourceFile, valueExpression, Diagnostics.Compiler_option_0_requires_a_value_of_type_1, option.name, getCompilerOptionValueTypeString(option)))"}
{"kind":"unused one-argument push","name":"convertToJson","at":"src/compiler/commandLineParser.ts:2602:13","text":"errors.push(createDiagnosticForNodeInSourceFile(sourceFile, valueExpression, Diagnostics.Property_value_can_only_be_string_literal_numeric_literal_true_false_null_object_literal_or_array_literal))"}
{"kind":"caller","name":"convertConfigFileToObject","at":"src/compiler/commandLineParser.ts:3603:18","text":"convertConfigFileToObject(\r\n        sourceFile,\r\n        errors,\r\n        { rootOptions, onPropertySet },\r\n    )"}
```

## Output test

With the locked `stage3/api` packages on `NODE_PATH`, run
`node stage3/adapt/99-seats-build22-62-append-callbacks/test.cjs <tree>`.
The test takes a tree written by `stage3/apply.sh` and reads only this
adapter's before and after refs under `refs/stage3/`. Later adapters can
edit the working tree without affecting the test. It emits the SDK from
the before snapshot and from the adapter's scratch output, then compares
diagnostics, values, and diagnostic array identity and order across all
seven append sites. It verifies zero edits and identical bytes on reapply.
Dropping the property-name append, prepending in the public caller,
accepting an extra reference to the old parameter, or accepting a binding
named `appendDiagnostic` inside a consumer each makes it fail.
