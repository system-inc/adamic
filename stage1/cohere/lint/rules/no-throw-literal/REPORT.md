# no-throw-literal

Landing base: `ad7bd06632f119abc7680719ad3a7d3b71100f58`. A typed node listener for `ThrowStatement`, verbatim messages, and an unchanged Go rule adapter that rejects unsupported options. `TestNoThrowLiteral` captures all seven upstream test functions.

- 48 upstream cases match byte for byte on Go, Node, emitted JavaScript and ASan/UBSan native, including findings, fixed source and empty fix/suggestion payloads.
- Three firing witnesses and nine additional controls agree. Controls include imported/shadowed undefined, Unicode, nested scopes, `.d.ts`/`.d.mts`/`.d.cts` aliases and no-program coverage. Declaration-file controls use real filenames separately because the witness copier retains only the final extension.
- `non Error throw suppressed` compiles and is caught by ordinary comparison on sanitized native, Node and emitted JavaScript.
- Only the shared live checker is used (`symbol-origin`, `scope-locals`). No private checker, hand-written answers, bridge changes or shared edits.
- Current 48-case process totals: sanitized native with recording 4.271s; Go 3.320s, under concurrent gate load. The prior supplemental typed corpus had 558 `.ts` files and 14,670,601 identical bytes; its logs remain below.
- Current package corpus: 814 files including `.a`, 30,871,050 identical bytes. Its shared typed-rule coverage is syntax-only; upstream replay supplies live checking.

Full all-input lint package passed: 152 pass, 0 fail, 1 skip (top level 46/0/1), all 94 mutants caught, 1333.147s wall, nproc 5, ending load 2.224/2.978/4.082. The sole skip, `TestCheckerBridgeRefusalPending`, awaits `tsgoInspect` returning `TSGoError`. Registry, gofmt and vet passed. Command: `go test -json -count=1 -timeout=3h ./stage1/cohere/lint`; all inputs are in `inputs.json` beside the log. Current logs: `/workspace/wave17-area-unit/clean-green/`; prior supplemental live-checker comparisons: `/workspace/wave17-clean-unit/no-throw-replay/`. Logs and wave suites are absent from the landing diff.

## Captured compiler options

Merged `22790cacf` and its typed-compiler-options parent `72bce0111`. The 48 distinct upstream cases below add 47 captured programs to the existing 261: `capturedTypedCases = 308`. The plain-harness guard from `TestNoThrowLiteralNeedsTheTypedHarness` adds the third `strictAloneTypedCases` entry. It is intentionally distinct from the typed source with its trailing newline. No compiler options or finding-count checks are relaxed.

Each added case is named by its exact captured source below; JSON escapes retain newlines. All captured programs use the upstream default compiler options.

| Replay | Exact source |
| --- | --- |
| Captured program | `"async function foo() { throw await bar; }\n"` |
| Captured program | `"class C { #field; foo() { throw foo.#field; } }\n"` |
| Captured program | `"declare const foo: unknown;\nthrow 'a' ?? 'b';\n"` |
| Captured program | `"declare const foo: unknown;\nthrow new Error() ?? 'literal';\n"` |
| Captured program | `"declare let foo: unknown;\nthrow foo ??= 'literal';\n"` |
| Captured program | `"function f() { throw; }\n"` |
| Captured program | `"function foo() { throw undefined; }\n"` |
| Captured program | `"function foo(undefined) { throw undefined; }\n"` |
| Captured program | `"function* foo() { var index = 0; throw yield index++; }\n"` |
| Captured program | `"throw 'a' + 'b';\n"` |
| Strict alone | `"throw 'error';"` |
| Captured program | `"throw 'error';\n"` |
| Captured program | `"throw 'literal' && 'not an Error';\n"` |
| Captured program | `"throw 'literal' && new Error();\n"` |
| Captured program | `"throw 0;\n"` |
| Captured program | `"throw 1, 2, new Error();\n"` |
| Captured program | `"throw Error('error');\n"` |
| Captured program | `"throw &#96;${err}&#96;;\n"` |
| Captured program | `"throw a;\n"` |
| Captured program | `"throw false;\n"` |
| Captured program | `"throw foo && 'literal'\n"` |
| Captured program | `"throw foo &&= 'literal'\n"` |
| Captured program | `"throw foo &= new Error();\n"` |
| Captured program | `"throw foo += new Error();\n"` |
| Captured program | `"throw foo = 'error';\n"` |
| Captured program | `"throw foo = new Error();\n"` |
| Captured program | `"throw foo ? 'literal' : new Error();\n"` |
| Captured program | `"throw foo ? 'not an Error' : 'literal';\n"` |
| Captured program | `"throw foo ? new Error() : 'literal';\n"` |
| Captured program | `"throw foo();\n"` |
| Captured program | `"throw foo.bar &#124;&#124;= 'literal'\n"` |
| Captured program | `"throw foo.bar;\n"` |
| Captured program | `"throw foo[bar] ??= 'literal'\n"` |
| Captured program | `"throw foo[bar];\n"` |
| Captured program | `"throw new Error('error');\n"` |
| Captured program | `"throw new Error() &#124;&#124; 'literal';\n"` |
| Captured program | `"throw new Error(), 1, 2, 3;\n"` |
| Captured program | `"throw new Error();\n"` |
| Captured program | `"throw new foo();\n"` |
| Captured program | `"throw null;\n"` |
| Captured program | `"throw obj?.foo\n"` |
| Captured program | `"throw obj?.foo()\n"` |
| Captured program | `"throw tag &#96;${foo}&#96;;\n"` |
| Captured program | `"throw undefined;\n"` |
| Captured program | `"throw {};\n"` |
| Captured program | `"try {throw new Error();} catch (e) {throw e;};\n"` |
| Captured program | `"var b = new Error(); throw 'a' + b;\n"` |
| Captured program | `"var e = new Error(); throw e;\n"` |
