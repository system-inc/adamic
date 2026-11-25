# Family 91 diagnostic chain array parameter

Adamic ruled this adaptation's shape on Nov 22 2026 (#bbssdcy, family 91, 'yes as sketched ... Mine wherever JavaScript moves'); she accepts the head after its review and census. At
`79aba3a7100e10e1ea6db7c630dc94a2f028b4d9`, item 1 returned (b).
The entry census recorded the first stop at `src/compiler/utilities.ts:8764:54`
in unit `src/compiler/utilities.ts:8753:1`.

> stage 0 can't lower Array.isArray on a tuple or an erased object/any/unknown view yet

The internal function accepts `DiagnosticMessageChain[] | undefined` and
assigns it directly to `next`. Existing arrays retain their identity. Known
single chains are wrapped once. The computed optional chain is evaluated
once, and the mixed expression wraps its single-chain branch at the caller.
The computed local's name must be absent from the file as both a decoded
identifier and a raw spelling so the inserted `const` cannot capture an existing reference.

The adapter identifies every call in the audited src and scripts roots by file, lexical declaration and ordinal,
and pins the exact first argument before writing any file. The test reads
the adapter's own snapshots, checks the complete set of kept and edited calls,
and exercises return values, array identity and computed evaluation.

The adapter parses every source file listed under the two roots, checks `Identifier`, string literal and no-substitution template literal nodes whose decoded text matches the helper name against the listed calls, declaration and pinned imports, then checks raw spellings throughout the file outside comments before writing.
The pinned imports are one plain `chainDiagnosticMessages` specifier from `./_namespaces/ts.js` in each of `src/compiler/checker.ts`, `src/compiler/program.ts`, `src/compiler/programDiagnostics.ts` and `src/compiler/watch.ts`.

`adapt.cjs` reads the `src` and `scripts` roots (lines 342 to 343).
The scan excludes `lib/` because it contains the prebuilt stock compiler that the build runs.

The guard's inventory is tsc's pinned source under those roots, where a new static or escaped spelling of the helper is refused at apply; a key or `eval` string assembled at run time cannot be checked statically, and the audit found none reaching the helper.
An outside Node caller that passes a single chain to the runtime export `chainDiagnosticMessages` gets a non-array `next`, which is outside the contract by Adamic's Nov 23 rule (a difference only a caller outside the public declarations can reach), the same shape as family 69's `setPackageJsonInfo`.

## Governing rows at the pin

### Row 24

- **Array membership in checked views waits for views v3; inferring physical storage from the asserted target is unsound; fixed object-backed tuples are admitted; an unused optional view is admitted lazily and reading its optional alias is refused.** NotYet and Refuse. Undated, compiler ruling (unattributed); source: docs/views-v2-dispatch-review.md:5-7,23,43-45. Spec: not in the spec. Test: internal/oracle/checked_views_v2_source_test.go:199 TestCheckedViewUntaggedArrayPending.

### Row 25

- **An element read through a checked view of a readonly array field is checked at each read and never cached, against the element type after instantiation, the same hatch as a field read through a view (#995z2fa ruling a): a lying element stops loudly, uncatchable, exit 70 on both backends with the same words, `element read failed: <expression> is not a <T>; expected <T>, found <kind>`, joining the one checked_boundary form beside `field read failed`. A read of an element type E is checked when E is comparable, either way, to the element type of an array some view holds, and always when E is or holds a type parameter, `any` or `unknown`, so a narrowing (`filter(isIdentifier)`, `as Identifier[]`) can't read viewed elements raw. A read-only pre-pass collects every view's array contract before any array read lowers, and a contract lowering makes that the pre-pass missed relowers the program with every read checked. Copies and element reads with no handout (slice, spread, flat, includes, indexOf, entries, a Set or Map built from it) are NotYet in the view's own words; a write through the view and the cast to a mutable array are NotYet by reach. A program that views no array gets no element check, its C unchanged.** Check, NotYet; on main since #6wfvfhm landed as 3edbe6404c. Nov 5 and 6 2026, @system_adamic and @system_adamic_compiler_lower; source: #6wfvfhm (the design, readonly fields and per-read checks; Lower's Nov 5 rulings, the key related either way and the pre-pass guard; Adamic's Nov 6 10:54Z ruling (b), one checked_boundary form, the regex exactly `(field|element) read failed: ... is not a <T>; expected <T>, found <kind>`, each lie fixture listed with a repair whose copy agrees with Node on both backends). Lifts the entry above's first clause for readonly fields; mutable arrays read from a record's union slot wait on #y852c3c. Spec: not in the spec. Test: internal/oracle/view_arrays_test.go:13 TestViewArrays over testdata/views_v3 (the good fixtures agree with Node on both backends; each lie stops at exit 70 in the element-read words; try-lie.ts keeps its exact bytes and stays uncatchable on native, sanitized native and JavaScript, :65); internal/lower/view_arrays_test.go:343 TestViewElementKeyAgreesWithLowering, :479 TestViewElementKeySeesInferredPredicateWrapper, :413 TestViewElementUnrelatedButReachableStops, :447 TestViewCopyRoadsStopAtTheirRead (slice, spread, new Set and new Map stop at their line and column), :562 TestViewWritableCastLeavesBrandMints; internal/oracle/checked_views_v2_migration_test.go:40 TestCheckedViewV2ArrayArmBoundary. The 27 lies and their repairs are in cloud/admission-corpus/negative-witnesses.json, and the mutants in internal/lower/view_elements.mutants.md. Language's 17 roads (views-v3-road-probes.tgz on #6wfvfhm) ran at the head and aren't pinned one by one.

### Row 260

- **`Array.isArray` is a library axiom exact about storage in both directions: it answers true exactly when the value is stored as an array, so it proves a union's arms (family 1's U4). On its false side, the arms tsc removes are proven absent when every removed arm is array-stored (a declared array, tuple or NodeArray type) and no kept arm could be (an `object`, `unknown` or `{}` arm); the kept arms are read as their own types, under their own rules. On its true side, element reads are proven only where every array arm tsc keeps came from a declared array type in the union, whose elements invariance already holds; an `any[]` arm tsc builds from an `object`, `unknown` or `{}` member leaves the elements unknown, and that site stays refused. An interface arm is in the domain (arrayPredicateDomain) only when it declares a required property that no array-stored type the program can hold has: not numeric-named, not `length`, not a member of lib's Array or ReadonlyArray, and not declared by any interface, class or intersection in the closed world whose storage is an array (NodeArray's `pos`, `end` and `hasTrailingComma` among them). A function whose whole body is `return Array.isArray(p)` on an untouched first parameter is read as the call on its argument (:253 applied), and tsc's `isArray` takes `value: unknown` by Correctness's adaptation. `isNodeArray` is a data predicate, not this axiom, and stays with its own family. The premise, named: a value at an arm is something assignable to it (Adamic proves types), and a property every value at the arm must have and no array can have rules out an array there; the list of array types can't outgrow that test, where testing assignability from `readonly unknown[]` failed open (a tuple, an element-typed interface and tsc's TextRange each take an array it doesn't). Today nothing builds an array that isn't one, or an array-shaped object that is one: `class X extends Array`, a NodeArray by `Object.assign` or cast, a Proxy, `arguments`, `ArrayLike`, typed arrays and a frozen array each stop at their own construction, and the subclass half of the domain leans on that NotYet.** Accept, Refuse otherwise; ruled and not built. Nov 22 2026, @system_adamic on #s8gy89r (the false side and the true side, with Language attacking the axiom's premise; Q1's domain, Language's form replacing Adamic's, which failed open; Q2 (a) and (b)). Language's attack on the premise held at main 0efbdb08 (#mzc2mmn, J-yv5vsff and J-32s3rqw: every road that could split the answer from storage stops before the test, and split, Object.keys, Set and Map spreads, a rest parameter and a class-or-array union agree with Node on both backends). Built only if Seats' first-stop table puts a U4 site in front (Adamic's order, Nov 22), by @system_adamic_platforms_cloudflare_workers. Spec: not in the spec. Test: owed by the build: one site of each read kind beside Node, and a union with an `object` arm, which stays refused; `Pair | Tagged` handed a tuple, `StringAt | Tagged` handed a `string[]` and `TextRange | Tagged` handed `Object.assign([...], { pos, end })`, each out of the domain and agreeing with Node or stopping; the mutant testing `readonly unknown[]` alone turns the TextRange row red; and the head reruns Language's isArray probes (attached to #mzc2mmn) the day any of their constructions lifts.

### Row 406

- **Adaptations to tsc's public API may only add truth (replace `any` with the type proven at every owner, add `| undefined`, add `readonly`); brand fields typed `any` become `undefined`; turning public host-interface methods into optional properties is not sanctioned; any other change needs a ruling.** Meaning. Oct 6 and 7 2026, @system_adamic; source: #bw3xg7c (hf9kqp5b, gqxzhnwv). Spec: not in the spec. **Extended, by name: removing a `readonly` that tsc's own public function disproves is adding truth, for Node's and NodeArray's `pos` and `end`, so both extend TextRange in place of ReadonlyTextRange, type-only.** Meaning. Nov 17 2026, @system_adamic, on #aryr7ea (00:17Z; the condition sharpened at 00:31Z). The premise: tsc's public `setTextRange` writes `pos` and `end` on a node after the factory made it, so readonly on Node was a promise to consumers, never a fact about the value. :93 stands whole, and every other removal still goes to Adamic with its count. The condition: no type assignable to Node or NodeArray redeclares `pos` or `end` narrower than `number`; none does, and SynthesizedComment's `-1`s are outside the ruling and stay refused. The cost: a consumer may write `node.pos` or `array.end` with no cast where upstream's types forbid it. Built as 17a20029 (stage3/adapt/80-writable-node-range; the family went from 550 sites to 2). Extended again Nov 18 2026, @system_adamic, on #aryr7ea (08:26Z, by name, for setTextRangePos and setTextRangeEnd; 09:37Z, setTextRangePosEnd and setTextRangePosWidth, which write through them): each setter's `T` is constrained to TextRange in place of ReadonlyTextRange, since each writes `pos` or `end`. All four are `@internal`, so no published line moves and the narrower-redeclaration condition falls away. Built as fbe5b72b (stage3/adapt/85-writable-range-setters; the family went from 2 to 0). Recorded: 80's two heritage clauses in its api-sanctions.json, regenerated into stage3/lane/sanctioned-api.json as `InterfaceDeclaration:["ts","Node"]#0` and `InterfaceDeclaration:["ts","NodeArray"]#0`; 85 has no sanction. Test: stage3/lane/test_check.py:112 LandingLaneTests.test_missing_sanctioned_line (80); NONE (implemented at stage3/adapt/85-writable-range-setters/adapt.cjs) for 85, whose mutant is a recorded run (J-x7nyzap). Applied as written, with no new ruling: `readonly useCaseSensitiveFileNames` on ModuleResolutionHost (#hes7595, ae2094f2, stage3/adapt/73-readonly-case-sensitivity), and InterfaceTypeWithDeclaredMembers' `declaredCallSignatures`, `declaredConstructSignatures` and `declaredIndexInfos` (#r11wppy, 1e3ff96a, stage3/adapt/96-readonly-declared-members), a readonly public line earned only by a listed site it clears (Adamic, #r11wppy, Nov 17 13:07Z).

# Family 91 caller audit

Adamic `79aba3a7100e10e1ea6db7c630dc94a2f028b4d9`, TypeScript `050880ce59e30b356b686bd3144efe24f875ebc8`.

The checker language service supplied the call references across src and scripts. Every result joined to a syntax call. Each identity includes the lexical declaration path, a declaration occurrence number for equal names, and the call's one-based ordinal in that declaration.

The checker found 102 calls and five non-call references. Shapes: array 2, array or undefined 1, mixed 1, single chain 23, single chain or undefined 20, undefined 55.

| File, declaration and call ordinal | Argument and location | Static type | Shape |
| --- | --- | --- | --- |
| `src/compiler/checker.ts` · `createTypeChecker #1 / resolveExternalModule #1` · 1 | `diagnosticDetails` at `src/compiler/checker.ts:4856:57` | `DiagnosticMessageChain \| undefined` | Single chain or undefined |
| `src/compiler/checker.ts` · `createTypeChecker #1 / resolveExternalModule #1` · 2 | `errorInfo` at `src/compiler/checker.ts:4941:96` | `DiagnosticMessageChain` | Single chain |
| `src/compiler/checker.ts` · `createTypeChecker #1 / errorOnImplicitAnyModule #1` · 1 | `errorInfo` at `src/compiler/checker.ts:4981:17` | `DiagnosticMessageChain \| undefined` | Single chain or undefined |
| `src/compiler/checker.ts` · `createTypeChecker #1 / resolveBaseTypesOfClass #1 / diagnostic #1` · 1 | `elaboration` at `src/compiler/checker.ts:13366:56` | `DiagnosticMessageChain \| undefined` | Single chain or undefined |
| `src/compiler/checker.ts` · `createTypeChecker #1 / elaborateNeverIntersection #1` · 1 | `errorInfo` at `src/compiler/checker.ts:15885:48` | `DiagnosticMessageChain \| undefined` | Single chain or undefined |
| `src/compiler/checker.ts` · `createTypeChecker #1 / elaborateNeverIntersection #1` · 2 | `errorInfo` at `src/compiler/checker.ts:15889:48` | `DiagnosticMessageChain \| undefined` | Single chain or undefined |
| `src/compiler/checker.ts` · `createTypeChecker #1 / getPropertyTypeForIndexType #1` · 1 | `undefined` at `src/compiler/checker.ts:19379:85` | `undefined` | Undefined |
| `src/compiler/checker.ts` · `createTypeChecker #1 / getPropertyTypeForIndexType #1` · 2 | `undefined` at `src/compiler/checker.ts:19383:85` | `undefined` | Undefined |
| `src/compiler/checker.ts` · `createTypeChecker #1 / getPropertyTypeForIndexType #1` · 3 | `undefined` at `src/compiler/checker.ts:19386:85` | `undefined` | Undefined |
| `src/compiler/checker.ts` · `createTypeChecker #1 / getPropertyTypeForIndexType #1` · 4 | `undefined` at `src/compiler/checker.ts:19389:85` | `undefined` | Undefined |
| `src/compiler/checker.ts` · `createTypeChecker #1 / getPropertyTypeForIndexType #1` · 5 | `undefined` at `src/compiler/checker.ts:19392:85` | `undefined` | Undefined |
| `src/compiler/checker.ts` · `createTypeChecker #1 / getPropertyTypeForIndexType #1` · 6 | `errorInfo` at `src/compiler/checker.ts:19396:37` | `DiagnosticMessageChain \| undefined` | Single chain or undefined |
| `src/compiler/checker.ts` · `createTypeChecker #1 / checkTypeRelatedTo #1 / reportError #1` · 1 | `errorInfo` at `src/compiler/checker.ts:22622:53` | `DiagnosticMessageChain \| undefined` | Single chain or undefined |
| `src/compiler/checker.ts` · `createTypeChecker #1 / checkJsxReturnAssignableToAppropriateBound #1 / generateInitialErrorChain #1` · 1 | `undefined` at `src/compiler/checker.ts:34266:56` | `undefined` | Undefined |
| `src/compiler/checker.ts` · `createTypeChecker #1 / checkJsxOpeningLikeElementOrOpeningFragment #1` · 1 | `undefined` at `src/compiler/checker.ts:34382:64` | `undefined` | Undefined |
| `src/compiler/checker.ts` · `createTypeChecker #1 / reportNonexistentProperty #1` · 1 | `errorInfo` at `src/compiler/checker.ts:35178:57` | `DiagnosticMessageChain \| undefined` | Single chain or undefined |
| `src/compiler/checker.ts` · `createTypeChecker #1 / reportNonexistentProperty #1` · 2 | `errorInfo` at `src/compiler/checker.ts:35186:49` | `DiagnosticMessageChain \| undefined` | Single chain or undefined |
| `src/compiler/checker.ts` · `createTypeChecker #1 / reportNonexistentProperty #1` · 3 | `errorInfo` at `src/compiler/checker.ts:35191:53` | `DiagnosticMessageChain \| undefined` | Single chain or undefined |
| `src/compiler/checker.ts` · `createTypeChecker #1 / reportNonexistentProperty #1` · 4 | `errorInfo` at `src/compiler/checker.ts:35199:57` | `DiagnosticMessageChain \| undefined` | Single chain or undefined |
| `src/compiler/checker.ts` · `createTypeChecker #1 / reportNonexistentProperty #1` · 5 | `errorInfo` at `src/compiler/checker.ts:35206:61` | `DiagnosticMessageChain \| undefined` | Single chain or undefined |
| `src/compiler/checker.ts` · `createTypeChecker #1 / reportNonexistentProperty #1` · 6 | `elaborateNeverIntersection(errorInfo, containingType)` at `src/compiler/checker.ts:35213:61` | `DiagnosticMessageChain \| undefined` | Single chain or undefined |
| `src/compiler/checker.ts` · `createTypeChecker #1 / checkTypeArguments #1 / errorInfo #1` · 1 | `undefined` at `src/compiler/checker.ts:36009:108` | `undefined` | Undefined |
| `src/compiler/checker.ts` · `createTypeChecker #1 / getArgumentArityError #1 / chain #1` · 1 | `undefined` at `src/compiler/checker.ts:36468:65` | `undefined` | Undefined |
| `src/compiler/checker.ts` · `createTypeChecker #1 / getArgumentArityError #1` · 1 | `chain` at `src/compiler/checker.ts:36469:49` | `DiagnosticMessageChain` | Single chain |
| `src/compiler/checker.ts` · `createTypeChecker #1 / getArgumentArityError #1 / chain #2` · 1 | `undefined` at `src/compiler/checker.ts:36478:65` | `undefined` | Undefined |
| `src/compiler/checker.ts` · `createTypeChecker #1 / getArgumentArityError #1` · 2 | `chain` at `src/compiler/checker.ts:36479:49` | `DiagnosticMessageChain` | Single chain |
| `src/compiler/checker.ts` · `createTypeChecker #1 / getArgumentArityError #1 / chain #3` · 1 | `undefined` at `src/compiler/checker.ts:36505:65` | `undefined` | Undefined |
| `src/compiler/checker.ts` · `createTypeChecker #1 / getArgumentArityError #1` · 3 | `chain` at `src/compiler/checker.ts:36506:49` | `DiagnosticMessageChain` | Single chain |
| `src/compiler/checker.ts` · `createTypeChecker #1 / getTypeArgumentArityError #1 / chain #1` · 1 | `undefined` at `src/compiler/checker.ts:36521:65` | `undefined` | Undefined |
| `src/compiler/checker.ts` · `createTypeChecker #1 / getTypeArgumentArityError #1` · 1 | `chain` at `src/compiler/checker.ts:36522:49` | `DiagnosticMessageChain` | Single chain |
| `src/compiler/checker.ts` · `createTypeChecker #1 / getTypeArgumentArityError #1 / chain #2` · 1 | `undefined` at `src/compiler/checker.ts:36542:65` | `undefined` | Undefined |
| `src/compiler/checker.ts` · `createTypeChecker #1 / getTypeArgumentArityError #1` · 2 | `chain` at `src/compiler/checker.ts:36543:49` | `DiagnosticMessageChain` | Single chain |
| `src/compiler/checker.ts` · `createTypeChecker #1 / getTypeArgumentArityError #1 / chain #3` · 1 | `undefined` at `src/compiler/checker.ts:36549:61` | `undefined` | Undefined |
| `src/compiler/checker.ts` · `createTypeChecker #1 / getTypeArgumentArityError #1` · 3 | `chain` at `src/compiler/checker.ts:36550:45` | `DiagnosticMessageChain` | Single chain |
| `src/compiler/checker.ts` · `createTypeChecker #1 / resolveCall #1` · 1 | `chain` at `src/compiler/checker.ts:36700:57` | `DiagnosticMessageChain \| undefined` | Single chain or undefined |
| `src/compiler/checker.ts` · `createTypeChecker #1 / resolveCall #1` · 2 | `chain` at `src/compiler/checker.ts:36701:57` | `DiagnosticMessageChain` | Single chain |
| `src/compiler/checker.ts` · `createTypeChecker #1 / resolveCall #1` · 3 | `chain` at `src/compiler/checker.ts:36704:57` | `DiagnosticMessageChain \| undefined` | Single chain or undefined |
| `src/compiler/checker.ts` · `createTypeChecker #1 / resolveCall #1 / chain #1` · 1 | `undefined` at `src/compiler/checker.ts:36727:81` | `undefined` | Undefined |
| `src/compiler/checker.ts` · `createTypeChecker #1 / resolveCall #1 / chain #2` · 1 | `map(diags, createDiagnosticMessageChainFromDiagnostic)` at `src/compiler/checker.ts:36746:25` | `DiagnosticMessageChain[]` | Array |
| `src/compiler/checker.ts` · `createTypeChecker #1 / resolveCall #1` · 4 | `chain` at `src/compiler/checker.ts:36750:57` | `DiagnosticMessageChain` | Single chain |
| `src/compiler/checker.ts` · `createTypeChecker #1 / invocationErrorDetails #1` · 1 | `errorInfo` at `src/compiler/checker.ts:37319:29` | `undefined` | Undefined |
| `src/compiler/checker.ts` · `createTypeChecker #1 / invocationErrorDetails #1` · 2 | `errorInfo` at `src/compiler/checker.ts:37326:29` | `DiagnosticMessageChain` | Single chain |
| `src/compiler/checker.ts` · `createTypeChecker #1 / invocationErrorDetails #1` · 3 | `undefined` at `src/compiler/checker.ts:37341:33` | `undefined` | Undefined |
| `src/compiler/checker.ts` · `createTypeChecker #1 / invocationErrorDetails #1` · 4 | `errorInfo` at `src/compiler/checker.ts:37350:21` | `undefined` | Undefined |
| `src/compiler/checker.ts` · `createTypeChecker #1 / invocationErrorDetails #1` · 5 | `errorInfo` at `src/compiler/checker.ts:37360:17` | `DiagnosticMessageChain \| undefined` | Single chain or undefined |
| `src/compiler/checker.ts` · `createTypeChecker #1 / invocationErrorDetails #1` · 6 | `errorInfo` at `src/compiler/checker.ts:37379:51` | `DiagnosticMessageChain` | Single chain |
| `src/compiler/checker.ts` · `createTypeChecker #1 / resolveDecorator #1 / messageChain #1` · 1 | `errorDetails.messageChain` at `src/compiler/checker.ts:37493:58` | `DiagnosticMessageChain` | Single chain |
| `src/compiler/checker.ts` · `createTypeChecker #1 / checkTypePredicate #1 / leadingError #1` · 1 | `undefined` at `src/compiler/checker.ts:42076:88` | `undefined` | Undefined |
| `src/compiler/checker.ts` · `createTypeChecker #1 / getAwaitedTypeNoAlias #1` · 1 | `chain` at `src/compiler/checker.ts:43832:53` | `DiagnosticMessageChain \| undefined` | Single chain or undefined |
| `src/compiler/checker.ts` · `createTypeChecker #1 / getAwaitedTypeNoAlias #1` · 2 | `chain` at `src/compiler/checker.ts:43834:49` | `DiagnosticMessageChain \| undefined` | Single chain or undefined |
| `src/compiler/checker.ts` · `createTypeChecker #1 / checkAsyncFunctionReturnType #1 / errorInfo #1` · 1 | `undefined` at `src/compiler/checker.ts:43927:130` | `undefined` | Undefined |
| `src/compiler/checker.ts` · `createTypeChecker #1 / issueMemberSpecificError #1 / rootChain #1` · 1 | `undefined` at `src/compiler/checker.ts:47442:41` | `undefined` | Undefined |
| `src/compiler/checker.ts` · `createTypeChecker #1 / checkInheritedPropertiesAreIdentical #1 / errorInfo #1` · 1 | `undefined` at `src/compiler/checker.ts:47765:77` | `undefined` | Undefined |
| `src/compiler/checker.ts` · `createTypeChecker #1 / checkInheritedPropertiesAreIdentical #1` · 1 | `errorInfo` at `src/compiler/checker.ts:47766:61` | `DiagnosticMessageChain` | Single chain |
| `src/compiler/program.ts` · `createProgram #3 / verifyCompilerOptions #1` · 1 | `chainDiagnosticMessages(/*details*/ undefined, Diagnostics.Visit_https_Colon_Slash_Slashaka_ms_Slashts6_for_migration_information)` at `src/compiler/program.ts:4297:25` | `DiagnosticMessageChain` | Single chain |
| `src/compiler/program.ts` · `createProgram #3 / verifyCompilerOptions #1` · 2 | `undefined` at `src/compiler/program.ts:4297:61` | `undefined` | Undefined |
| `src/compiler/program.ts` · `createProgram #3 / verifyCompilerOptions #1 / verifyEmitFilePath #1` · 1 | `undefined` at `src/compiler/program.ts:4427:69` | `undefined` | Undefined |
| `src/compiler/program.ts` · `createProgram #3 / verifyCompilerOptions #1 / verifyEmitFilePath #1` · 2 | `chain` at `src/compiler/program.ts:4429:53` | `DiagnosticMessageChain \| undefined` | Single chain or undefined |
| `src/compiler/program.ts` · `createProgram #3 / verifyDeprecatedCompilerOptions #1 / createDiagnostic #1 / details #1` · 1 | `undefined` at `src/compiler/program.ts:4496:67` | `undefined` | Undefined |
| `src/compiler/program.ts` · `createProgram #3 / verifyDeprecatedCompilerOptions #1 / createDiagnostic #1` · 1 | `details` at `src/compiler/program.ts:4498:55` | `DiagnosticMessageChain` | Single chain |
| `src/compiler/program.ts` · `createProgram #3 / verifyDeprecatedCompilerOptions #1 / createDiagnostic #1 / chain #1` · 1 | `details` at `src/compiler/program.ts:4500:55` | `DiagnosticMessageChain` | Single chain |
| `src/compiler/program.ts` · `createProgram #3 / verifyDeprecatedCompilerOptions #1 / createDiagnostic #1` · 2 | `undefined` at `src/compiler/program.ts:4506:67` | `undefined` | Undefined |
| `src/compiler/program.ts` · `createProgram #3 / verifyDeprecatedCompilerOptions #1 / createDiagnostic #1 / chain #2` · 1 | `details` at `src/compiler/program.ts:4508:55` | `DiagnosticMessageChain \| undefined` | Single chain or undefined |
| `src/compiler/programDiagnostics.ts` · `createProgramDiagnostics #1 / createDiagnosticExplainingFile #1` · 1 | `cachedChain.details` at `src/compiler/programDiagnostics.ts:240:49` | `DiagnosticMessageChain[]` | Array |
| `src/compiler/programDiagnostics.ts` · `createProgramDiagnostics #1 / createDiagnosticExplainingFile #1` · 2 | `fileIncludeReasons` at `src/compiler/programDiagnostics.ts:263:110` | `DiagnosticMessageChain[] \| undefined` | Array or undefined |
| `src/compiler/programDiagnostics.ts` · `createProgramDiagnostics #1 / createDiagnosticExplainingFile #1` · 3 | `redirectInfo ? fileIncludeReasonDetails ? [fileIncludeReasonDetails, ...redirectInfo] : redirectInfo : fileIncludeReasonDetails` at `src/compiler/programDiagnostics.ts:265:17` | `DiagnosticMessageChain \| DiagnosticMessageChain[] \| undefined` | Mixed |
| `src/compiler/utilities.ts` · `createModuleNotFoundChain #1 / result #1` · 1 | `undefined` at `src/compiler/utilities.ts:840:29` | `undefined` | Undefined |
| `src/compiler/utilities.ts` · `createModuleNotFoundChain #1 / result #1` · 2 | `undefined` at `src/compiler/utilities.ts:845:29` | `undefined` | Undefined |
| `src/compiler/utilities.ts` · `createModuleNotFoundChain #1 / result #1` · 3 | `undefined` at `src/compiler/utilities.ts:852:25` | `undefined` | Undefined |
| `src/compiler/utilities.ts` · `createModuleNotFoundChain #1 / result #1` · 4 | `undefined` at `src/compiler/utilities.ts:859:25` | `undefined` | Undefined |
| `src/compiler/utilities.ts` · `createModuleNotFoundChain #1 / result #1` · 5 | `undefined` at `src/compiler/utilities.ts:865:25` | `undefined` | Undefined |
| `src/compiler/utilities.ts` · `createModeMismatchDetails #1 / result #1` · 1 | `undefined` at `src/compiler/utilities.ts:882:29` | `undefined` | Undefined |
| `src/compiler/utilities.ts` · `createModeMismatchDetails #1 / result #1` · 2 | `undefined` at `src/compiler/utilities.ts:888:29` | `undefined` | Undefined |
| `src/compiler/utilities.ts` · `createModeMismatchDetails #1 / result #1` · 3 | `undefined` at `src/compiler/utilities.ts:894:25` | `undefined` | Undefined |
| `src/compiler/utilities.ts` · `createModeMismatchDetails #1 / result #1` · 4 | `undefined` at `src/compiler/utilities.ts:899:25` | `undefined` | Undefined |
| `src/compiler/watch.ts` · `explainIfFileIsRedirectAndImpliedFormat #1` · 1 | `undefined` at `src/compiler/watch.ts:369:25` | `undefined` | Undefined |
| `src/compiler/watch.ts` · `explainIfFileIsRedirectAndImpliedFormat #1` · 2 | `undefined` at `src/compiler/watch.ts:376:25` | `undefined` | Undefined |
| `src/compiler/watch.ts` · `explainIfFileIsRedirectAndImpliedFormat #1` · 3 | `undefined` at `src/compiler/watch.ts:386:37` | `undefined` | Undefined |
| `src/compiler/watch.ts` · `explainIfFileIsRedirectAndImpliedFormat #1` · 4 | `undefined` at `src/compiler/watch.ts:395:37` | `undefined` | Undefined |
| `src/compiler/watch.ts` · `explainIfFileIsRedirectAndImpliedFormat #1` · 5 | `undefined` at `src/compiler/watch.ts:404:37` | `undefined` | Undefined |
| `src/compiler/watch.ts` · `fileIncludeReasonToDiagnostics #1` · 1 | `undefined` at `src/compiler/watch.ts:487:25` | `undefined` | Undefined |
| `src/compiler/watch.ts` · `fileIncludeReasonToDiagnostics #1` · 2 | `undefined` at `src/compiler/watch.ts:496:98` | `undefined` | Undefined |
| `src/compiler/watch.ts` · `fileIncludeReasonToDiagnostics #1` · 3 | `undefined` at `src/compiler/watch.ts:499:76` | `undefined` | Undefined |
| `src/compiler/watch.ts` · `fileIncludeReasonToDiagnostics #1` · 4 | `undefined` at `src/compiler/watch.ts:503:33` | `undefined` | Undefined |
| `src/compiler/watch.ts` · `fileIncludeReasonToDiagnostics #1` · 5 | `undefined` at `src/compiler/watch.ts:510:33` | `undefined` | Undefined |
| `src/compiler/watch.ts` · `fileIncludeReasonToDiagnostics #1` · 6 | `undefined` at `src/compiler/watch.ts:520:29` | `undefined` | Undefined |
| `src/compiler/watch.ts` · `fileIncludeReasonToDiagnostics #1` · 7 | `undefined` at `src/compiler/watch.ts:540:56` | `undefined` | Undefined |
| `src/compiler/watch.ts` · `fileIncludeReasonToDiagnostics #1` · 8 | `undefined` at `src/compiler/watch.ts:543:88` | `undefined` | Undefined |
| `src/compiler/watch.ts` · `fileIncludeReasonToDiagnostics #1` · 9 | `undefined` at `src/compiler/watch.ts:546:56` | `undefined` | Undefined |
| `src/testRunner/unittests/diagnosticCollection.ts` · `chain #1` · 1 | `ts.chainDiagnosticMessages(/*details*/ undefined, ts.Diagnostics.Did_you_mean_0, "x")` at `src/testRunner/unittests/diagnosticCollection.ts:16:17` | `ts.DiagnosticMessageChain` | Single chain |
| `src/testRunner/unittests/diagnosticCollection.ts` · `chain #1` · 2 | `undefined` at `src/testRunner/unittests/diagnosticCollection.ts:16:56` | `undefined` | Undefined |
| `src/testRunner/unittests/diagnosticCollection.ts` · `chain #2` · 1 | `ts.chainDiagnosticMessages(/*details*/ undefined, ts.Diagnostics.Did_you_mean_0, "a")` at `src/testRunner/unittests/diagnosticCollection.ts:36:17` | `ts.DiagnosticMessageChain` | Single chain |
| `src/testRunner/unittests/diagnosticCollection.ts` · `chain #2` · 2 | `undefined` at `src/testRunner/unittests/diagnosticCollection.ts:36:56` | `undefined` | Undefined |
| `src/testRunner/unittests/diagnosticCollection.ts` · `chainBetter #1` · 1 | `ts.chainDiagnosticMessages( ts.chainDiagnosticMessages(/*details*/ undefined, ts.Diagnostics.Did_you_mean_0, "x"), ts.Diagnostics.Did_you_mean_0, "b", )` at `src/testRunner/unittests/diagnosticCollection.ts:44:17` | `ts.DiagnosticMessageChain` | Single chain |
| `src/testRunner/unittests/diagnosticCollection.ts` · `chainBetter #1` · 2 | `ts.chainDiagnosticMessages(/*details*/ undefined, ts.Diagnostics.Did_you_mean_0, "x")` at `src/testRunner/unittests/diagnosticCollection.ts:45:21` | `ts.DiagnosticMessageChain` | Single chain |
| `src/testRunner/unittests/diagnosticCollection.ts` · `chainBetter #1` · 3 | `undefined` at `src/testRunner/unittests/diagnosticCollection.ts:45:60` | `undefined` | Undefined |
| `src/testRunner/unittests/diagnosticCollection.ts` · `chain #3` · 1 | `ts.chainDiagnosticMessages(/*details*/ undefined, ts.Diagnostics.Did_you_mean_0, "x")` at `src/testRunner/unittests/diagnosticCollection.ts:65:17` | `ts.DiagnosticMessageChain` | Single chain |
| `src/testRunner/unittests/diagnosticCollection.ts` · `chain #3` · 2 | `undefined` at `src/testRunner/unittests/diagnosticCollection.ts:65:56` | `undefined` | Undefined |
| `src/testRunner/unittests/diagnosticCollection.ts` · `chain #4` · 1 | `ts.chainDiagnosticMessages(/*details*/ undefined, ts.Diagnostics.Did_you_mean_0, "x")` at `src/testRunner/unittests/diagnosticCollection.ts:91:17` | `ts.DiagnosticMessageChain` | Single chain |
| `src/testRunner/unittests/diagnosticCollection.ts` · `chain #4` · 2 | `undefined` at `src/testRunner/unittests/diagnosticCollection.ts:91:56` | `undefined` | Undefined |
| `src/testRunner/unittests/diagnosticCollection.ts` · `chain #5` · 1 | `ts.chainDiagnosticMessages(/*details*/ undefined, ts.Diagnostics.Did_you_mean_0, "x")` at `src/testRunner/unittests/diagnosticCollection.ts:112:17` | `ts.DiagnosticMessageChain` | Single chain |
| `src/testRunner/unittests/diagnosticCollection.ts` · `chain #5` · 2 | `undefined` at `src/testRunner/unittests/diagnosticCollection.ts:112:56` | `undefined` | Undefined |

## Mixed argument and producers

The sole mixed argument is call 3 of `createProgramDiagnostics #1 / createDiagnosticExplainingFile #1`. The checker still reports the same three arms. Its producers are unchanged at the pre-adapter snapshot.

| Producer | Static result |
| --- | --- |
| `src/compiler/programDiagnostics.ts:266:48` | `[fileIncludeReasonDetails, ...redirectInfo]` creates the combined array on the existing both-present branch. |
| `src/compiler/programDiagnostics.ts:266:94` | The existing `redirectInfo` array on the present-redirect branch with no file-include chain. |
| `src/compiler/programDiagnostics.ts:211:13`, `:225:13`, `:229:13` | `redirectInfo` starts undefined, or takes `cachedChain.redirectInfo`, or reads `file && explainIfFileIsRedirectAndImpliedFormat(...)`. An absent file yields undefined. |
| `src/compiler/watch.ts:366:9`, `:368:9`, `:375:9`, `:385:21`, `:394:21`, `:403:21`, `:411:5` | Initializes `result` as undefined. Every writer uses `(result ??= []).push(...)`. The return reads the array or initial absence. |
| `src/compiler/programDiagnostics.ts:216:13`, `:289:17` | The local cache declared at :105 reads entries inserted by :289. The redirect field stores exactly the local `redirectInfo` traced above. |
| `src/compiler/programDiagnostics.ts:210:13`, `:263:13`, `:267:21` | The fallback local starts undefined. The `seenReasons && chainDiagnosticMessages(...)` write yields undefined or the declared single chain at `src/compiler/utilities.ts:8753:1`. |
| `src/compiler/programDiagnostics.ts:245:25`, `:285:21`, `:289:17` | The guarded cache read yields one chain. Both cache writes take their value from the same local traced above. |

Both array branches can retain their original array. The no-redirect branch wraps the present chain once. All 102 call forms fit the round 1 operations. The installed declaration and upstream baseline each contain zero references to the internal helper.

## Diff against round 1

I rebuilt the full round 1 source from commit 0efbdb083a60086fec3a29687a162bbf21105336 and joined its checker references to the same syntax identities. The old table has 101 calls. The new table has 102.

The only changed declaration is `src/compiler/utilities.ts`, `createModuleNotFoundChain #1 / result #1`. Adapter 104, `99-seats-build22-104-module-chain-arguments`, replaces old ordinal 1, which used `alternateResultMessage[0]` and spread `alternateResultMessage[1]`, with the two direct calls below. Both direct calls keep the first argument `undefined`.

| Round 1 | Pin | Shape at details |
| --- | --- | --- |
| 1, tuple forwarding | 1, direct Node10 arguments, call at :839:15 | Undefined |
| 1, tuple forwarding | 2, direct package exports arguments, call at :844:15 | Undefined |
| 2, DefinitelyTyped package fallback | 3, same fallback | Undefined |
| 3, new declaration fallback | 4, same fallback | Undefined |
| 4, install types or declaration fallback | 5, same fallback | Undefined |

The old tuple-forwarding call is gone and two direct calls replace it. On the literal declaration-and-ordinal key there is one added key, ordinal 5, with no missing keys and no changes to the first-argument text or shape of shared keys. The three fallback ordinals shifted because of the split. No other call was added, removed, or changed shape.
