# Stage 3 roots that require native runtime work

This is a source-ownership triage, not a new census or a claim that stage 3 runs.
Rank uses **deduplicated root-site count**, not hidden bytes: the latest root
CSV has no per-root byte attribution, and summing overlapping hidden units would
inflate credit. Every one of the 6,703 measured sites is classified below through
its exact reason; 472 reason groups retain the original counts and one example.

## Frozen evidence

- Work branch: `runtime/stage3-runtime-roots`, based on fetched
  `origin/area/runtime` at `c1c6073021e8e5cd6005fcdee59eb3b377395fd8`.
- Root-table branch: fetched `origin/codex/stage3-notyet-table` at
  `e8c283b5ed32477805357b652a170b85a04b2469`. Use
  [rerun-0730/TABLE.md](https://github.com/system-inc/adamic/blob/e8c283b5ed32477805357b652a170b85a04b2469/stage3/notyet-table/rerun-0730/TABLE.md)
  and its `after/roots.csv`, not the older top-level TABLE.md (6,791 sites).
  The newer table reports 6,703 roots, 472 reasons and 108 conservatively retained
  unattributed reads, after removing 4,047 proven echoes from 10,750 sites.
  The 2026-10-08 10:40:28Z after compiler is scratch merge `0e5661e4`;
  these numbers are **not** a measurement of this runtime branch.
- Main: fetched at `89ac4a8c1de0b02d95965be72f7f8bf1c92433d2`.
  Its latest full integration report is
  [20261008T023735Z.XREV1N/integration-report.md](https://github.com/system-inc/adamic/blob/89ac4a8c1de0b02d95965be72f7f8bf1c92433d2/stage3/meter/runs/20261008T023735Z.XREV1N/integration-report.md),
  measured with compiler `132a0ed5`: main/area whole-program 2/79,
  own-file 54/79, 1,468 NotYet and 5,162 Refused findings.
  The previous `20261008T000754Z.hFDR7Q` report and newer
  `20261008T033212Z.debug-watch` diagnosis were also read. The diagnosis attributes
  debug/watch failures to Error host declarations, exact optional properties and
  the loader's composite-project root list, not native runtime C.

Do not add meter findings to the root census: they have different compilers,
source scopes and accounting. The census is checker-rejected (324 diagnostic
sites; 10,551 attempted units; 147 split around diagnosed bodies). It records
lowering observations, not emitter execution or native behavior. In particular,
main's large “a cast the runtime can't check” Refused category is a lowering
proof/policy refusal, not evidence of a missing C function. Non-null, numeric enum
refinement, predicate proof, declarations without bodies, branded strings and
generic substitutions remain compiler work.

## Ownership rule

The **immediate refusal is lowering for all 472 reasons**: that is what this census
measures. “Runtime required” below means lifting the guard soundly needs a missing
C operation or richer native storage; it does **not** mean deleting the guard or
changing C alone will retire the count. Such entries explicitly require lowering,
IR and emitter coordination. Compiler-owned entries lack proof, type substitution,
binding, ABI planning or an IR operation; the current C storage already handles
the concrete values involved. The census has no emitter-execution refusals.
Method replacement is assigned
to emitter/layout ownership because existing C closure fields can represent
replacement callables; its lowering gate and generated dispatch must change.
Other emitter integration points are cited where runtime work needs them.

Runtime-dependent representation entries are ranked separately from claims of a
small runtime-only patch. Conservative guards such as fixed-object enumeration
also admit compiler-only special cases: their full count is an opportunity ceiling,
not guaranteed runtime credit.

## Runtime-required candidates, ranked by root sites

| Rank | Root sites | Exact reason | Missing native support | Refusal / C evidence |
| ---: | ---: | --- | --- | --- |
| 1 | 10 | for...in without a proven fixed plain-object origin (arrays, prototypes and absent synthetic fields cannot be enumerated soundly) | representation: Actual own-property presence, enumerable attributes and prototype traversal; includes a lowering origin-proof guard. | internal/lower/library_for_in.go:22; internal/native/runtime/adamic.h:258 |
| 2 | 8 | regex replacement other than a string | function / callback ABI: Callback replacement; current native replace takes an adamic_string, not a closure. Capture arguments, undefined captures, groups, offset and input need an ABI. | internal/lower/regexp.go:191; internal/native/runtime/regexp.c:920 |
| 3 | 3 | JSON.stringify object references (structural types can hide fields and toJSON; runtime shapes need complete value metadata) | representation: Shapes distinguish references from scalars, not complete JSON types, presence or toJSON behavior for hidden fields. | internal/lower/library_json_stringify.go:182; internal/native/runtime/adamic.h:227 |
| 4 | 3 | RegExp with a nonconstant pattern | runtime parser / compiler: Runtime consumes a compiled program; patterns are compiled in Go by regexConstant, not parsed in C at construction. | internal/lower/regexp.go:39; internal/native/runtime/regexp.h:48 |
| 5 | 3 | lastIndexOf with these arguments | function and emitter call: String lastIndexOf lacks the optional position argument in lowering, emitter and runtime signature; array search already has its separate fromIndex path. | internal/lower/object.go:1539; internal/native/runtime/string_search_impl.h:113 |
| 6 | 2 | a template interpolating an object, an array, a map, a function or undefined | conversion / representation: Generic JS ToPrimitive/ToString needs object hooks, arrays and function source semantics; no blanket reference formatting shortcut is sound. | internal/lower/expression.go:850; internal/native/runtime/adamic.h:258 |
| 7 | 1 | JSON.stringify a union containing containers without runtime element metadata | representation: Runtime container element schemas are needed; a generic tagged container reference does not provide them. | internal/lower/library_json_stringify.go:192; internal/native/runtime/adamic.h:381 |
| 8 | 1 | for...in over an array (holes and own enumerable properties are not represented; use for...of for elements) | representation: Array holes and arbitrary enumerable own properties are absent from the dense array layout. | internal/lower/library_for_in.go:17; internal/native/runtime/adamic.h:381 |

These 8 exact reasons account for 31 historical sites;
another 17 sites are assigned to emitter/layout, and the remaining 6655
sites are classified as lowering-owned below.
These are disjoint reason counts, not predicted compilation gains.

### Why the top candidate is not contained

Stop at the list. The 10 general for...in roots are behind a sound lowering
origin guard, and lifting it for arbitrary represented objects needs a runtime
object model, not one C function plus an emitter call:

- `internal/lower/library_for_in.go:22` requires a fixed plain-object origin
  because structural views may hide arrays, absent properties or prototypes.
  Compiler-only proofs can recover individual cases; these 10 sites are not
  guaranteed retirement from one runtime patch.
- `internal/native/runtime/adamic.h:227` describes fixed shapes with names,
  a reference bitmap and method tables; objects at line 258 have no generic
  prototype link or own-property presence/enumerable descriptors.
- Arrays at line 381 hold dense length/capacity/elements and regex properties,
  not general sparse-hole presence and enumerable property storage.
- Existing `adamic_object_keys` enumerates known own shape names. General
  for...in must preserve JS key order and duplicate suppression, traverse
  prototypes, exclude non-enumerable properties and respect deleted/absent keys.
  This affects layout, allocation/release, mutation, IR and emitter behavior.

Method replacement (17 sites) is specifically **not** in the runtime-ranked
list: `internal/lower/class.go:416` rejects it; the emitter generates immutable
method tables (`emit_objects.go:262`), but existing C closure-valued fields and
`adamic_object_callee` can already hold and call an own replacement. Changing
layout/dispatch and proving receiver ABI is compiler/emitter work. This is not
proof that all current method dispatch can be fixed by dropping the guard.

A contained later candidate is String.lastIndexOf's optional position (three
sites): `internal/native/emit_strings.go:74` currently calls the two-argument
C helper and `internal/lower/object.go:1539` admits only one source argument.
It would need an explicit absent-versus-present position API and JS
ToIntegerOrInfinity/NaN/UTF-16 boundary fixtures. This task does not skip the
highest-ranked item to implement that smaller one.

### Important compiler-owned counterexamples

- Nested named functions (3,239 roots): lowering rejects FunctionDeclaration
  in a body. C already has `adamic_closure`, environments and cells; supporting
  hoisting, recursive binding and capture planning is compiler work.
- Map key refusals (205), Set branded keys and `__String`/Path value refusals:
  `mapTypes` first calls `representation`, and keyability excludes an unproved
  type. Runtime maps already store supported primitive/reference keys. Branded
  string recognition is not a new C key representation.
- `for...of over an object` (124): lower's object fallback and iterator-origin
  planning, not proof that C lacks iteration. Current lowering/native already
  support user iterator methods, library iterators and regex iterator stepping.
  A concrete iterator view and receiver convention must be established first.
- Tuple indexOf/push, unsupported arrays/fields/return types, generic overloads,
  union capture/argument conversions and optional-call guards: IR/layout/ABI
  selection or sound proof in lowering. Do not infer a missing runtime feature
  solely from the phrase “representation”. Runtime tagged values and closure
  storage already exist.
- The historical Uint16Array reason has one site; the current runtime base
  already supports Uint16Array in `typed_arrays.go` and `typed_array.c`.
  It is a stale lowering-census observation, not a runtime backlog item.
- Object.keys/entries/assign origin guards are lower's shape/provenance checks;
  C already enumerates known own shape fields. Complete dynamic object support
  is larger work, but these rows alone do not prove a missing helper.

## Complete reason classification

“L” is lowering-owned; “E” is emitter/layout-owned with a lowering gate;
“R” requires runtime support plus compiler integration.
Every row's immediate refusal is lowering. Evidence for dynamic type strings is
the named lowering family, not an assertion that the full type text occurs as a
literal in Go. Unattributed reads are kept in L conservatively, not promoted to
independent runtime lessons. Emitter/layout-owned: one reason group, 17 sites; immediate emitter refusals
were not measured.

| Root sites | Owner | Exact reason | Refusal family / evidence | First measured root |
| ---: | --- | --- | --- | --- |
| 3239 | L | a function inside a function (a closure) | internal/lower/statements.go | src/compiler/binder.ts:1079:5 |
| 205 | L | a Map whose keys aren't strings, numbers, booleans, objects, arrays, maps or functions | internal/lower/object.go | src/compiler/binder.ts:2413:9 |
| 187 | L | a value of type __String | internal/lower/object.go | src/compiler/binder.ts:1607:30 |
| 172 | L | assigning to an Identifier | internal/lower (syntax / binding / ABI guard) | src/compiler/builder.ts:1954:25 |
| 156 | L | passing union of differently held members to a function value | internal/lower/functions.go | src/compiler/checker.ts:10140:21 |
| 145 | L | a call through ?. (an optional call) | internal/lower (syntax / binding / ABI guard) | src/compiler/builder.ts:1852:30 |
| 131 | L | a value of type Path | internal/lower/object.go | src/compiler/binder.ts:585:73 |
| 124 | L | for...of over an object | internal/lower/object.go:815; internal/lower/iteration.go | src/compiler/binder.ts:1296:36 |
| 100 | L | a function returning T | internal/lower/functions.go:78 | src/compiler/checker.ts:10284:30 |
| 96 | L | a value of type any | internal/lower/object.go | src/compiler/builder.ts:1919:25 |
| 92 | L | a void call used as a value | internal/lower (syntax / binding / ABI guard) | src/compiler/binder.ts:3644:15 |
| 90 | L | an array of never | internal/lower/object.go | src/compiler/builderState.ts:567:38 |
| 89 | L | an ElementAccessExpression | internal/lower (syntax / binding / ABI guard) | src/compiler/binder.ts:1746:21 |
| 76 | L | a BinaryExpression as a statement | internal/lower/expression.go | src/compiler/binder.ts:1956:21 |
| 75 | L | a function returning T &#124; undefined | internal/lower/functions.go:78 | src/compiler/checker.ts:25177:14 |
| 62 | L | a declaration directly in a case (wrap the case in a block) | internal/lower (syntax / binding / ABI guard) | src/compiler/binder.ts:1326:17 |
| 57 | L | reading performance | internal/lower/expression.go | src/compiler/binder.ts:503:5 |
| 52 | L | a value of type T | internal/lower/object.go | src/compiler/binder.ts:1477:71 |
| 49 | L | a call returning void &#124; undefined | internal/lower (syntax / binding / ABI guard) | src/compiler/binder.ts:587:13 |
| 46 | L | a rest array of DiagnosticArguments | internal/lower/object.go | src/compiler/binder.ts:2746:72 |
| 45 | L | a function value returning union of differently held members | internal/lower/functions.go | src/compiler/checker.ts:10979:24 |
| 43 | L | an array of T | internal/lower/object.go | src/compiler/checker.ts:46986:35 |
| 35 | L | this outside a method | internal/lower (syntax / binding / ABI guard) | src/compiler/checker.ts:1456:5 |
| 32 | L | a BinaryExpression with a number and a number | internal/lower/expression.go | src/compiler/scanner.ts:1968:36 |
| 27 | L | a boolean &#124; undefined variable a function value captures | internal/lower (syntax / binding / ABI guard) | src/compiler/checker.ts:22675:26 |
| 26 | L | a call to a PropertyAccessExpression | internal/lower (syntax / binding / ABI guard) | src/compiler/parser.ts:3508:30 |
| 25 | L | a value of type __String &#124; undefined | internal/lower/object.go | src/compiler/binder.ts:755:15 |
| 24 | L | a field of type string &#124; NodeArray&lt;JSDocComment&gt; &#124; undefined | internal/lower/object.go | src/compiler/binder.ts:2121:20 |
| 22 | L | new an Identifier | internal/lower (syntax / binding / ABI guard) | src/compiler/checker.ts:14134:21 |
| 21 | L | an array of any | internal/lower/object.go | src/compiler/checker.ts:11031:21 |
| 21 | L | a generic function as a value | internal/lower/functions.go | src/compiler/core.ts:1370:140 |
| 21 | L | a value of type SolutionBuilderState&lt;T&gt; | internal/lower/object.go | src/compiler/tsbuildPublic.ts:1204:5 |
| 20 | L | a value of type unknown | internal/lower/object.go | src/compiler/binder.ts:495:50 |
| 19 | L | a Set of __String (a Set holds strings, numbers, booleans, objects, arrays, maps or functions so far) | internal/lower/object.go | src/compiler/binder.ts:3615:17 |
| 18 | L | a value of type T["kind"] | internal/lower/object.go | src/compiler/factory/nodeFactory.ts:1209:45 |
| 18 | L | a value of type ResolvedConfigFilePath | internal/lower/object.go | src/compiler/tsbuildPublic.ts:1431:90 |
| 17 | L | optional chaining to .size on a value | internal/lower (syntax / binding / ABI guard) | src/compiler/builder.ts:1796:34 |
| 17 | E | replacing a represented method at runtime | internal/lower/class.go:416; internal/native/emit_objects.go:262 | src/compiler/core.ts:1544:5 |
| 16 | L | a function returning __String &#124; undefined | internal/lower/functions.go:78 | src/compiler/binder.ts:661:14 |
| 16 | L | a BinaryExpression with a value and a value | internal/lower/expression.go | src/compiler/checker.ts:12867:24 |
| 16 | L | a value of type object | internal/lower/object.go | src/compiler/checker.ts:15341:23 |
| 16 | L | a BinaryExpression with a string and a number | internal/lower/expression.go | src/compiler/checker.ts:16886:17 |
| 15 | L | a Set of Path (a Set holds strings, numbers, booleans, objects, arrays, maps or functions so far) | internal/lower/object.go | src/compiler/builder.ts:1429:9 |
| 15 | L | a function returning __String | internal/lower/functions.go:78 | src/compiler/checker.ts:2436:14 |
| 14 | L | an optional chain longer than one step | internal/lower (syntax / binding / ABI guard) | src/compiler/checker.ts:36469:31 |
| 14 | L | a SpreadElement | internal/lower (syntax / binding / ABI guard) | src/compiler/checker.ts:49673:90 |
| 14 | L | a function returning any | internal/lower/functions.go:78 | src/compiler/commandLineParser.ts:2437:10 |
| 13 | L | a BinaryExpression with a number and a string | internal/lower/expression.go | src/compiler/checker.ts:18324:24 |
| 12 | L | a function returning undefined | internal/lower/functions.go:78 | src/compiler/binder.ts:2458:14 |
| 12 | L | a union of differently held members variable a function value captures | internal/lower (syntax / binding / ABI guard) | src/compiler/commandLineParser.ts:2732:30 |
| 12 | L | a value of type T &#124; undefined | internal/lower/object.go | src/compiler/core.ts:2139:72 |
| 12 | L | a value of type ResolvedConfigFileName | internal/lower/object.go | src/compiler/program.ts:1290:19 |
| 11 | L | a field of type true &#124; Node &#124; undefined | internal/lower/object.go | src/compiler/binder.ts:2600:13 |
| 11 | L | a function returning Path | internal/lower/functions.go:78 | src/compiler/builder.ts:2350:14 |
| 11 | L | assigning a field of a value | internal/lower/class.go | src/compiler/checker.ts:1974:9 |
| 10 | L | a field of type string &#124; DiagnosticMessageChain | internal/lower/object.go | src/compiler/builder.ts:1631:173 |
| 10 | L | assigning an element of a value | internal/lower (syntax / binding / ABI guard) | src/compiler/checker.ts:19531:9 |
| 10 | R | for...in without a proven fixed plain-object origin (arrays, prototypes and absent synthetic fields cannot be enumerated soundly) | internal/lower/library_for_in.go:22 | src/compiler/commandLineParser.ts:2788:24 |
| 10 | L | a function returning U &#124; undefined | internal/lower/functions.go:78 | src/compiler/core.ts:33:17 |
| 8 | L | a value of type K | internal/lower/object.go | src/compiler/builderState.ts:171:55 |
| 8 | L | a value of type NonNullable&lt;T&gt; | internal/lower/object.go | src/compiler/checker.ts:20543:23 |
| 8 | L | new a ParenthesizedExpression | internal/lower (syntax / binding / ABI guard) | src/compiler/checker.ts:2940:58 |
| 8 | R | regex replacement other than a string | internal/lower/regexp.go:191 | src/compiler/checker.ts:8943:45 |
| 8 | L | a function returning T[] | internal/lower/functions.go:78 | src/compiler/core.ts:1018:17 |
| 8 | L | a value of type readonly T[] &#124; undefined | internal/lower/object.go | src/compiler/core.ts:145:26 |
| 8 | L | a value of type ((node: Node) =&gt; boolean) &#124; undefined | internal/lower/object.go | src/compiler/debug.ts:283:72 |
| 7 | L | a value of type HasJSDoc &#124; undefined | internal/lower/object.go | src/compiler/binder.ts:2538:19 |
| 7 | L | a value of type BindableStaticNameExpression | internal/lower/object.go | src/compiler/binder.ts:3438:43 |
| 7 | L | a VoidExpression as a statement | internal/lower (syntax / binding / ABI guard) | src/compiler/checker.ts:21958:13 |
| 7 | L | a field of type string &#124; number &#124; undefined | internal/lower/object.go | src/compiler/checker.ts:47931:26 |
| 7 | L | a rest parameter outside a nongeneric named function | internal/lower (syntax / binding / ABI guard) | src/compiler/core.ts:2480:13 |
| 6 | L | an array of boolean &#124; undefined | internal/lower/object.go | src/compiler/binder.ts:1933:17 |
| 6 | L | ?.[] on a value | internal/lower (syntax / binding / ABI guard) | src/compiler/checker.ts:22931:59 |
| 6 | L | a value of type ExpressionWithTypeArguments &amp; { readonly expression: Identifier &#124; PropertyAccessEntityNameExpression; } | internal/lower/object.go | src/compiler/checker.ts:44317:60 |
| 6 | L | an array of U | internal/lower/object.go | src/compiler/core.ts:2524:17 |
| 6 | L | reading ts | internal/lower/expression.go | src/compiler/emitter.ts:985:13 |
| 6 | L | a value of type TOuterState | internal/lower/object.go | src/compiler/factory/utilities.ts:1280:276 |
| 6 | L | a case whose type differs from the switch's | internal/lower (syntax / binding / ABI guard) | src/compiler/program.ts:1164:9 |
| 6 | L | reading currentLexicalScope | internal/lower/expression.go | src/compiler/transformers/ts.ts:1127:125 |
| 5 | L | a value of type HasJSDoc | internal/lower/object.go | src/compiler/checker.ts:16519:50 |
| 5 | L | a narrowed union member whose object tag cannot be checked with typeof; keep differently held object kinds in separately typed variables | internal/lower (syntax / binding / ABI guard) | src/compiler/checker.ts:37210:16 |
| 5 | L | a value of type readonly T[] | internal/lower/object.go | src/compiler/checker.ts:46982:109 |
| 5 | L | a value of type object &#124; undefined | internal/lower/object.go | src/compiler/moduleNameResolver.ts:2256:54 |
| 5 | L | an overloaded function as a value | internal/lower (syntax / binding / ABI guard) | src/compiler/scanner.ts:2187:67 |
| 5 | L | a value of type string &#124; null &#124; undefined | internal/lower/object.go | src/compiler/sourcemap.ts:701:24 |
| 5 | L | an array of ResolvedConfigFileName | internal/lower/object.go | src/compiler/tsbuildPublic.ts:2226:28 |
| 5 | L | reading host | internal/lower/expression.go | src/compiler/watchPublic.ts:1005:35 |
| 4 | L | a function returning CompilerOptionsValue | internal/lower/functions.go:78 | src/compiler/builder.ts:1456:14 |
| 4 | L | a value of type NamedDeclaration &amp; { name: DeclarationName; } | internal/lower/object.go | src/compiler/checker.ts:11157:106 |
| 4 | L | a field of type NodeArray&lt;ParameterDeclaration&gt; &#124; readonly JSDocParameterTag[] | internal/lower/object.go | src/compiler/checker.ts:16202:62 |
| 4 | L | a function returning readonly T[] &#124; undefined | internal/lower/functions.go:78 | src/compiler/checker.ts:20540:14 |
| 4 | L | a for...of destructuring an object | internal/lower (syntax / binding / ABI guard) | src/compiler/checker.ts:26333:18 |
| 4 | L | a value of type string &#124; (void &amp; { __escapedIdentifier: void; }) &#124; (string &amp; { __escapedIdentifier: void; }) | internal/lower/object.go | src/compiler/checker.ts:33795:34 |
| 4 | L | a field of type false &#124; Symbol &#124; undefined | internal/lower/object.go | src/compiler/checker.ts:34057:22 |
| 4 | L | a field of type "boolean" &#124; "list" &#124; "listOrElement" &#124; "number" &#124; "object" &#124; "string" &#124; Map&lt;string, string &#124; number&gt; | internal/lower/object.go | src/compiler/commandLineParser.ts:2599:13 |
| 4 | L | a value of type CompilerOptionsValue | internal/lower/object.go | src/compiler/commandLineParser.ts:2759:46 |
| 4 | L | a function returning T[] &#124; undefined | internal/lower/functions.go:78 | src/compiler/core.ts:871:17 |
| 4 | L | a function returning U | internal/lower/functions.go:78 | src/compiler/core.ts:93:17 |
| 4 | L | a value of type (EmitNode &amp; { autoGenerate: AutoGenerateInfo; }) &#124; (EmitNode &amp; { autoGenerate: AutoGenerateInfo; }) | internal/lower/object.go | src/compiler/emitter.ts:5457:30 |
| 4 | L | a field holding union of differently held members | internal/lower (syntax / binding / ABI guard) | src/compiler/moduleNameResolver.ts:3252:17 |
| 4 | L | a computed field name | internal/lower (syntax / binding / ABI guard) | src/compiler/parser.ts:506:5 |
| 4 | L | a value of type T &#124; Program | internal/lower/object.go | src/compiler/program.ts:5060:5 |
| 4 | L | a value of type ImmediatelyInvokedArrowFunction | internal/lower/object.go | src/compiler/transformers/classFields.ts:1462:19 |
| 4 | L | a value of type (ConstructorDeclaration &amp; { body: Block; }) &#124; undefined | internal/lower/object.go | src/compiler/transformers/es2015.ts:1151:15 |
| 3 | L | a ConditionalExpression as a statement | internal/lower/expression.go | src/compiler/binder.ts:1080:30 |
| 3 | L | an array of undefined | internal/lower/object.go | src/compiler/binder.ts:1940:40 |
| 3 | L | a value of type BindableObjectDefinePropertyCall | internal/lower/object.go | src/compiler/binder.ts:3201:45 |
| 3 | L | a value of type BindableAccessExpression | internal/lower/object.go | src/compiler/binder.ts:3411:54 |
| 3 | L | a destructured name that isn't plain | internal/lower (syntax / binding / ABI guard) | src/compiler/checker.ts:16773:28 |
| 3 | L | a value of type TypeNode &amp; LiteralTypeNode &amp; { readonly literal: StringLiteral; } | internal/lower/object.go | src/compiler/checker.ts:20007:71 |
| 3 | L | a function returning object | internal/lower/functions.go:78 | src/compiler/checker.ts:25300:14 |
| 3 | L | a value of type never | internal/lower/object.go | src/compiler/checker.ts:30269:86 |
| 3 | L | a value of type Identifier &#124; __String | internal/lower/object.go | src/compiler/checker.ts:3237:9 |
| 3 | L | a value of type false &#124; TypeOnlyAliasDeclaration &#124; undefined | internal/lower/object.go | src/compiler/checker.ts:4352:15 |
| 3 | L | a comparator that doesn't take two elements and return a number | internal/lower (syntax / binding / ABI guard) | src/compiler/checker.ts:53816:31 |
| 3 | R | RegExp with a nonconstant pattern | internal/lower/regexp.go:39 | src/compiler/commandLineParser.ts:4127:56 |
| 3 | L | a value of type T[] | internal/lower/object.go | src/compiler/core.ts:1003:33 |
| 3 | R | lastIndexOf with these arguments | internal/lower/object.go:1539 | src/compiler/core.ts:2433:11 |
| 3 | L | an array of NonNullable&lt;T&gt; | internal/lower/object.go | src/compiler/core.ts:739:31 |
| 3 | L | reading encodeURI | internal/lower/expression.go | src/compiler/emitter.ts:1118:24 |
| 3 | R | JSON.stringify object references (structural types can hide fields and toJSON; runtime shapes need complete value metadata) | internal/lower/library_json_stringify.go:182 | src/compiler/emitter.ts:1138:27 |
| 3 | L | a call returning any | internal/lower (syntax / binding / ABI guard) | src/compiler/emitter.ts:1509:44 |
| 3 | L | a value of type Children &#124; undefined | internal/lower/object.go | src/compiler/emitter.ts:4663:108 |
| 3 | L | a value of type EndOfFileToken | internal/lower/object.go | src/compiler/factory/nodeFactory.ts:6043:9 |
| 3 | L | a destructured name held otherwise than its field | internal/lower (syntax / binding / ABI guard) | src/compiler/factory/nodeFactory.ts:7440:9 |
| 3 | L | storing true &#124; Node &#124; undefined in a field | internal/lower/class.go | src/compiler/parser.ts:1341:5 |
| 3 | L | a function returning NodeArray&lt;T&gt; | internal/lower/functions.go:78 | src/compiler/parser.ts:2594:14 |
| 3 | L | a PrefixUnaryExpression on a number | internal/lower/expression.go | src/compiler/performance.ts:42:13 |
| 3 | L | a value of type Path &#124; undefined | internal/lower/object.go | src/compiler/program.ts:2697:59 |
| 3 | L | a function returning ResolvedConfigFileName | internal/lower/functions.go:78 | src/compiler/program.ts:5128:17 |
| 3 | L | Object.entries on a shape not proven by a plain literal or its const binding | internal/lower/library_json_stringify.go; internal/lower/class_features.go | src/compiler/scanner.ts:222:46 |
| 3 | L | destructuring anything but a tuple into [names] | internal/lower/object.go | src/compiler/semver.ts:146:11 |
| 3 | L | a value of type WrappedExpression&lt;AnonymousFunctionDefinition&gt; | internal/lower/object.go | src/compiler/transformers/esDecorators.ts:1730:56 |
| 3 | L | a value of type InitializedVariableDeclaration | internal/lower/object.go | src/compiler/transformers/generators.ts:1387:43 |
| 3 | L | reading builderProgram | internal/lower/expression.go | src/compiler/watchPublic.ts:610:16 |
| 3 | L | a Map of false &#124; MutableFileSystemEntries | internal/lower/object.go | src/compiler/watchUtilities.ts:123:39 |
| 3 | L | a value of type X | internal/lower/object.go | src/compiler/watchUtilities.ts:750:9 |
| 2 | L | a number &#124; undefined argument to substring | internal/lower/object.go | src/compiler/builder.ts:1617:68 |
| 2 | L | a value of type IncrementalBuildInfoFileId | internal/lower/object.go | src/compiler/builder.ts:2369:32 |
| 2 | L | an array of Path | internal/lower/object.go | src/compiler/builderState.ts:479:23 |
| 2 | L | destructuring a value | internal/lower/object.go | src/compiler/checker.ts:11491:19 |
| 2 | L | storing false &#124; Type in a field | internal/lower/class.go | src/compiler/checker.ts:15240:21 |
| 2 | L | a BinaryExpression with a number &#124; undefined and a number | internal/lower/expression.go | src/compiler/checker.ts:25061:29 |
| 2 | L | reading isFinite | internal/lower/expression.go | src/compiler/checker.ts:26619:16 |
| 2 | L | a value of type ParameterPropertyDeclaration | internal/lower/object.go | src/compiler/checker.ts:2968:55 |
| 2 | L | a value of type LeftHandSideExpression &amp; Identifier | internal/lower/object.go | src/compiler/checker.ts:38034:27 |
| 2 | L | a value of type NodeArray&lt;Expression&gt; &amp; readonly [BindableStaticNameExpression, NumericLiteral &#124; StringLiteralLike, ObjectLiteralExpression] &amp; Readonly&lt;...&gt; | internal/lower/object.go | src/compiler/checker.ts:39747:53 |
| 2 | L | a value of type (ModuleSpecifierResolutionHost &amp; { getCommonSourceDirectory(): string; }) &#124; undefined | internal/lower/object.go | src/compiler/checker.ts:54337:82 |
| 2 | L | a rest array of (string &#124; number &#124; boolean &#124; readonly string[] &#124; SourceFile &#124; undefined)[] | internal/lower/object.go | src/compiler/commandLineParser.ts:2206:63 |
| 2 | L | a function returning SortedReadonlyArray&lt;T&gt; | internal/lower/functions.go:78 | src/compiler/core.ts:1038:17 |
| 2 | L | an array of V | internal/lower/object.go | src/compiler/core.ts:107:25 |
| 2 | L | a value of type U | internal/lower/object.go | src/compiler/core.ts:1202:60 |
| 2 | R | a template interpolating an object, an array, a map, a function or undefined | internal/lower/expression.go:850 | src/compiler/core.ts:1786:59 |
| 2 | L | a function returning NonNullable&lt;T&gt; | internal/lower/functions.go:78 | src/compiler/core.ts:1909:12 |
| 2 | L | a value of type U &#124; readonly U[] &#124; undefined | internal/lower/object.go | src/compiler/core.ts:403:19 |
| 2 | L | a value of type U &#124; undefined | internal/lower/object.go | src/compiler/core.ts:484:15 |
| 2 | L | a function returning V | internal/lower/functions.go:78 | src/compiler/core.ts:519:17 |
| 2 | L | a value of type K &#124; undefined | internal/lower/object.go | src/compiler/core.ts:560:13 |
| 2 | L | a BinaryExpression with a string and a value | internal/lower/expression.go | src/compiler/debug.ts:517:15 |
| 2 | L | a library method value outside a const alias, typed call/apply, or supported map callback (its receiver and callable ABI are not proven); wrap the call in an arrow | internal/lower (syntax / binding / ABI guard) | src/compiler/debug.ts:554:24 |
| 2 | L | an array of Child | internal/lower/object.go | src/compiler/emitter.ts:4734:86 |
| 2 | L | a field of type string &#124; number &#124; boolean &#124; DiagnosticMessage &#124; undefined | internal/lower/object.go | src/compiler/executeCommandLine.ts:259:44 |
| 2 | L | iterating a value | internal/lower/object.go:815; internal/lower/iteration.go | src/compiler/expressionToTypeNode.ts:1242:52 |
| 2 | L | a tagged template other than the intrinsic String.raw | internal/lower (syntax / binding / ABI guard) | src/compiler/factory/emitHelpers.ts:1470:11 |
| 2 | L | a rest parameter other than an array | internal/lower (syntax / binding / ABI guard) | src/compiler/factory/nodeFactory.ts:2363:40 |
| 2 | L | reading factory | internal/lower/expression.go | src/compiler/factory/nodeFactory.ts:496:144 |
| 2 | L | a value of type RedirectsCacheKey | internal/lower/object.go | src/compiler/moduleNameResolver.ts:1015:15 |
| 2 | L | a function returning ModeAwareCacheKey | internal/lower/functions.go:78 | src/compiler/moduleNameResolver.ts:1115:17 |
| 2 | L | a field of type boolean &#124; (() =&gt; boolean) &#124; undefined | internal/lower/object.go | src/compiler/moduleNameResolver.ts:3425:13 |
| 2 | L | a function returning PackageJson[K] &#124; undefined | internal/lower/functions.go:78 | src/compiler/moduleNameResolver.ts:361:10 |
| 2 | L | indexOf on a tuple (a tuple is held as an object, not an array, so far; write it as an array where it's made) | internal/lower/object.go | src/compiler/moduleSpecifiers.ts:1374:24 |
| 2 | L | a value of type false &#124; RegExpExecArray &#124; null | internal/lower/object.go | src/compiler/parser.ts:10715:11 |
| 2 | L | ?. to a number, which would be number &#124; undefined | internal/lower (syntax / binding / ABI guard) | src/compiler/parser.ts:6176:20 |
| 2 | L | a value of type ExpressionWithTypeArguments &amp; { expression: Identifier &#124; PropertyAccessEntityNameExpression; } | internal/lower/object.go | src/compiler/parser.ts:9537:23 |
| 2 | L | a value of type SourceFile | internal/lower/object.go | src/compiler/program.ts:1109:5 |
| 2 | L | a function returning readonly T[] | internal/lower/functions.go:78 | src/compiler/program.ts:2778:14 |
| 2 | L | a number &#124; undefined argument to slice | internal/lower/object.go | src/compiler/program.ts:2981:46 |
| 2 | L | a function returning Path &#124; undefined | internal/lower/functions.go:78 | src/compiler/resolutionCache.ts:258:17 |
| 2 | L | spreading an array of other elements | internal/lower/object.go | src/compiler/scanner.ts:3534:99 |
| 2 | L | reading getModifiedTime | internal/lower/expression.go | src/compiler/sys.ts:1055:84 |
| 2 | L | reading timerToUpdateChildWatches | internal/lower/expression.go | src/compiler/sys.ts:780:17 |
| 2 | L | a value of type AnonymousFunctionDefinition | internal/lower/object.go | src/compiler/transformers/classFields.ts:1471:50 |
| 2 | L | a value of type (AssignmentExpression&lt;EqualsToken&gt; &amp; { readonly left: GeneratedIdentifier; }) &#124; undefined | internal/lower/object.go | src/compiler/transformers/classFields.ts:2705:19 |
| 2 | L | a Map of VisitResult&lt;ExportAssignment &#124; LateVisibilityPaintedStatement &#124; undefined&gt; | internal/lower/object.go | src/compiler/transformers/declarations.ts:1347:9 |
| 2 | L | a value of type ThisCapturingVariableDeclaration | internal/lower/object.go | src/compiler/transformers/es2015.ts:1466:19 |
| 2 | L | a value of type EmitNode &amp; { autoGenerate: AutoGenerateInfo; } | internal/lower/object.go | src/compiler/transformers/module/module.ts:2383:53 |
| 2 | L | a value of type CompilerHost &amp; ReadBuildProgramHost | internal/lower/object.go | src/compiler/tsbuildPublic.ts:1083:23 |
| 2 | L | a function returning void &#124; "skip" | internal/lower/functions.go:78 | src/compiler/utilities.ts:10727:14 |
| 2 | L | incrementing an Identifier | internal/lower (syntax / binding / ABI guard) | src/compiler/utilities.ts:6384:13 |
| 2 | L | push with other than one value | internal/lower/object.go | src/compiler/utilities.ts:7775:13 |
| 2 | L | a value of type V &#124; undefined | internal/lower/object.go | src/compiler/utilities.ts:939:15 |
| 2 | L | a tuple element of type string &#124; number &#124; boolean &#124; readonly string[] &#124; SourceFile &#124; undefined | internal/lower/object.go | src/compiler/watch.ts:291:91 |
| 2 | L | a Map of HostFileInfo | internal/lower/object.go | src/compiler/watchPublic.ts:751:33 |
| 2 | L | a value of type Canonicalized | internal/lower/object.go | src/compiler/watchUtilities.ts:209:68 |
| 2 | L | a function returning WatchFactory&lt;X, Y&gt;[T] | internal/lower/functions.go:78 | src/compiler/watchUtilities.ts:727:14 |
| 1 | L | a value of type (value: T) =&gt; void | internal/lower/object.go | src/compiler/binder.ts:1477:43 |
| 1 | L | a function returning void &#124; number &#124; Symbol | internal/lower/functions.go:78 | src/compiler/binder.ts:2846:14 |
| 1 | L | a value of type BindablePropertyAssignmentExpression &#124; PropertyAccessExpression &#124; LiteralLikeElementAccessExpression | internal/lower/object.go | src/compiler/binder.ts:3267:41 |
| 1 | L | a value of type PropertyAccessExpression &#124; LiteralLikeElementAccessExpression | internal/lower/object.go | src/compiler/binder.ts:3350:45 |
| 1 | L | a value of type BindableStaticAccessExpression | internal/lower/object.go | src/compiler/binder.ts:3384:46 |
| 1 | L | a value of type CallExpression &#124; BindableStaticAccessExpression | internal/lower/object.go | src/compiler/binder.ts:3470:57 |
| 1 | L | a BinaryExpression with a boolean and a boolean | internal/lower/expression.go | src/compiler/builder.ts:1229:12 |
| 1 | L | a function returning IncrementalBuildInfoFileId | internal/lower/functions.go:78 | src/compiler/builder.ts:1388:14 |
| 1 | L | a function returning IncrementalBuildInfoFileIdListId | internal/lower/functions.go:78 | src/compiler/builder.ts:1397:14 |
| 1 | L | a function returning ReusableDiagnosticMessageChain | internal/lower/functions.go:78 | src/compiler/builder.ts:1533:14 |
| 1 | L | a field of type EmitSignature &#124; undefined | internal/lower/object.go | src/compiler/builder.ts:1960:61 |
| 1 | L | incrementing a NonNullExpression | internal/lower (syntax / binding / ABI guard) | src/compiler/builder.ts:2148:17 |
| 1 | L | a value of type IncrementalMultiFileEmitBuildInfoFileInfo | internal/lower/object.go | src/compiler/builder.ts:2229:52 |
| 1 | L | a value of type IncrementalBuildInfoFilePendingEmit | internal/lower/object.go | src/compiler/builder.ts:2239:5 |
| 1 | L | a value of type IncrementalBuildInfoFileIdListId | internal/lower/object.go | src/compiler/builder.ts:2362:29 |
| 1 | L | a value of type readonly IncrementalBundleEmitBuildInfoFileInfo[] &#124; readonly IncrementalMultiFileEmitBuildInfoFileInfo[] where an array goes | internal/lower/object.go | src/compiler/builder.ts:2407:5 |
| 1 | L | storing Path &#124; undefined in a field | internal/lower/class.go | src/compiler/builder.ts:668:13 |
| 1 | L | a value of type __String &amp; string | internal/lower/object.go | src/compiler/checker.ts:1155:7 |
| 1 | L | a ClassExpression | internal/lower (syntax / binding / ABI guard) | src/compiler/checker.ts:1451:21 |
| 1 | L | a value of type ReplaceableIndexedAccessType | internal/lower/object.go | src/compiler/checker.ts:14688:55 |
| 1 | L | a field of type false &#124; Type &#124; undefined | internal/lower/object.go | src/compiler/checker.ts:15219:13 |
| 1 | L | an array of object | internal/lower/object.go | src/compiler/checker.ts:15326:33 |
| 1 | L | reading errorInfo | internal/lower/expression.go | src/compiler/checker.ts:22477:17 |
| 1 | L | push on a tuple (a tuple is held as an object, not an array, so far; write it as an array where it's made) | internal/lower/object.go | src/compiler/checker.ts:22624:17 |
| 1 | L | a function returning VisitResult&lt;T&gt; | internal/lower/functions.go:78 | src/compiler/checker.ts:2501:14 |
| 1 | L | a function returning InferenceContext &#124; (T &amp; undefined) | internal/lower/functions.go:78 | src/compiler/checker.ts:26263:14 |
| 1 | L | a function returning TypeMapper &#124; (T &amp; undefined) | internal/lower/functions.go:78 | src/compiler/checker.ts:26378:14 |
| 1 | L | a value of type Source | internal/lower/object.go | src/compiler/checker.ts:27040:71 |
| 1 | L | a value of type PropertyDeclaration &#124; ParameterPropertyDeclaration | internal/lower/object.go | src/compiler/checker.ts:3166:67 |
| 1 | L | a function value taking string &#124; number &#124; undefined | internal/lower/functions.go | src/compiler/checker.ts:33269:50 |
| 1 | L | a value of type TypeOnlyAliasDeclaration &#124; undefined | internal/lower/object.go | src/compiler/checker.ts:3353:23 |
| 1 | L | a value of type Identifier &#124; PrivateIdentifier &#124; __String | internal/lower/object.go | src/compiler/checker.ts:3396:29 |
| 1 | L | storing false &#124; Symbol in a field | internal/lower/class.go | src/compiler/checker.ts:34075:13 |
| 1 | L | .length on a union of differently held members | internal/lower/expression.go | src/compiler/checker.ts:36290:43 |
| 1 | L | a value of type NamedTupleMember &#124; (ParameterDeclaration &amp; { name: Identifier; }) | internal/lower/object.go | src/compiler/checker.ts:38497:116 |
| 1 | L | a field of type 0 &#124; boolean &#124; undefined | internal/lower/object.go | src/compiler/checker.ts:39371:13 |
| 1 | L | a value of type ClassStaticBlockDeclaration &#124; Decorator &#124; PrivateIdentifierGetAccessorDeclaration &#124; ... 5 more ... &#124; undefined | internal/lower/object.go | src/compiler/checker.ts:44188:27 |
| 1 | L | a value of type (ExportDeclaration &amp; { readonly isTypeOnly: true; readonly moduleSpecifier: Expression; }) &#124; undefined | internal/lower/object.go | src/compiler/checker.ts:4418:9 |
| 1 | L | a function returning TypeOnlyAliasDeclaration &#124; undefined | internal/lower/functions.go:78 | src/compiler/checker.ts:4455:14 |
| 1 | L | a value of type Map&lt;string, [K, V[]]&gt; | internal/lower/object.go | src/compiler/checker.ts:44581:31 |
| 1 | L | a function returning ClassStaticBlockDeclaration &#124; Decorator &#124; PrivateIdentifierGetAccessorDeclaration &#124; ... 5 more ... &#124; undefined | internal/lower/functions.go:78 | src/compiler/checker.ts:47028:14 |
| 1 | L | a value of type ClassElement &#124; ParameterPropertyDeclaration | internal/lower/object.go | src/compiler/checker.ts:47278:9 |
| 1 | L | a value of type AliasDeclarationNode | internal/lower/object.go | src/compiler/checker.ts:48394:31 |
| 1 | L | reading tracker | internal/lower/expression.go | src/compiler/checker.ts:51508:26 |
| 1 | L | a BinaryExpression with a number and a boolean | internal/lower/expression.go | src/compiler/checker.ts:53697:14 |
| 1 | L | a value of type Declaration &amp; Expression | internal/lower/object.go | src/compiler/checker.ts:6291:116 |
| 1 | L | a function returning Declaration &amp; HasModifiers | internal/lower/functions.go:78 | src/compiler/checker.ts:6600:18 |
| 1 | L | reading add | internal/lower/expression.go | src/compiler/checker.ts:8242:41 |
| 1 | L | reading moduleSpecifiers | internal/lower/expression.go | src/compiler/checker.ts:8534:36 |
| 1 | L | a value of type ModeAwareCacheKey | internal/lower/object.go | src/compiler/checker.ts:8627:19 |
| 1 | L | new Map from something that isn't [key, value] pairs | internal/lower/object.go | src/compiler/commandLineParser.ts:141:65 |
| 1 | L | a field of type "boolean" &#124; "number" &#124; "object" &#124; "string" &#124; Map&lt;string, string &#124; number&gt; | internal/lower/object.go | src/compiler/commandLineParser.ts:1896:13 |
| 1 | L | a Map of CompilerOptionsValue | internal/lower/object.go | src/compiler/commandLineParser.ts:2785:20 |
| 1 | L | Object.keys on a shape not proven by a plain literal or its const binding | internal/lower/library_json_stringify.go; internal/lower/class_features.go | src/compiler/commandLineParser.ts:2838:39 |
| 1 | R | JSON.stringify a union containing containers without runtime element metadata | internal/lower/library_json_stringify.go:192 | src/compiler/commandLineParser.ts:2960:31 |
| 1 | L | a Map of string &#124; number | internal/lower/object.go | src/compiler/commandLineParser.ts:3869:17 |
| 1 | L | an array of CanonicalKey | internal/lower/object.go | src/compiler/commandLineParser.ts:4131:47 |
| 1 | L | a value of type CanonicalKey | internal/lower/object.go | src/compiler/commandLineParser.ts:4140:25 |
| 1 | L | a function returning CanonicalKey | internal/lower/functions.go:78 | src/compiler/commandLineParser.ts:4170:10 |
| 1 | L | a YieldExpression as a statement | internal/lower (syntax / binding / ABI guard) | src/compiler/core.ts:1045:9 |
| 1 | L | a function returning T &#124; readonly T[] &#124; undefined | internal/lower/functions.go:78 | src/compiler/core.ts:1160:17 |
| 1 | L | a value of type MapLike&lt;T&gt; | internal/lower/object.go | src/compiler/core.ts:1287:31 |
| 1 | L | overload 2 of arrayFrom with additional implementation type parameters | internal/lower/functions.go | src/compiler/core.ts:1339:1 |
| 1 | L | overload 1 of arrayToMap with additional implementation type parameters | internal/lower/functions.go | src/compiler/core.ts:1401:1 |
| 1 | L | overload 1 of arrayToNumericMap with additional implementation type parameters | internal/lower/functions.go | src/compiler/core.ts:1420:1 |
| 1 | L | overload 1 of arrayToMultiMap with additional implementation type parameters | internal/lower/functions.go | src/compiler/core.ts:1434:1 |
| 1 | L | overload 3 of group with additional implementation type parameters | internal/lower/functions.go | src/compiler/core.ts:1452:1 |
| 1 | L | a function returning { [P in K as `${P}`]?: T[]; } | internal/lower/functions.go:78 | src/compiler/core.ts:1464:17 |
| 1 | L | a function returning T1 &amp; T2 | internal/lower/functions.go:78 | src/compiler/core.ts:1495:17 |
| 1 | L | a value of type T1 | internal/lower/object.go | src/compiler/core.ts:1513:51 |
| 1 | L | a function returning ((...args: A) =&gt; R) &#124; undefined | internal/lower/functions.go:78 | src/compiler/core.ts:1522:17 |
| 1 | L | a function returning MultiMap&lt;K, V&gt; | internal/lower/functions.go:78 | src/compiler/core.ts:1542:17 |
| 1 | L | a function returning Queue&lt;T&gt; | internal/lower/functions.go:78 | src/compiler/core.ts:1569:17 |
| 1 | L | a rest array of T[] | internal/lower/object.go | src/compiler/core.ts:1577:22 |
| 1 | L | a function returning TOut &#124; undefined | internal/lower/functions.go:78 | src/compiler/core.ts:1778:17 |
| 1 | L | a function returning TOut | internal/lower/functions.go:78 | src/compiler/core.ts:1783:17 |
| 1 | L | a function returning () =&gt; T | internal/lower/functions.go:78 | src/compiler/core.ts:1891:17 |
| 1 | L | a function returning (arg: A) =&gt; T | internal/lower/functions.go:78 | src/compiler/core.ts:1907:17 |
| 1 | L | a Map of T | internal/lower/object.go | src/compiler/core.ts:1908:17 |
| 1 | L | a value of type A | internal/lower/object.go | src/compiler/core.ts:1909:13 |
| 1 | L | a BinaryExpression with a union of differently held members and a union of differently held members | internal/lower/expression.go | src/compiler/core.ts:1978:9 |
| 1 | L | a function returning (arg: T) =&gt; boolean | internal/lower/functions.go:78 | src/compiler/core.ts:2454:17 |
| 1 | L | a value of type NonNullable&lt;U&gt; | internal/lower/object.go | src/compiler/core.ts:2501:15 |
| 1 | L | a function returning T[][] | internal/lower/functions.go:78 | src/compiler/core.ts:2531:17 |
| 1 | L | a function returning U[] &#124; undefined | internal/lower/functions.go:78 | src/compiler/core.ts:320:17 |
| 1 | L | a function returning readonly U[] &#124; undefined | internal/lower/functions.go:78 | src/compiler/core.ts:350:17 |
| 1 | L | a value of type T &#124; T[] &#124; readonly T[] &#124; undefined | internal/lower/object.go | src/compiler/core.ts:378:15 |
| 1 | L | a function returning readonly U[] | internal/lower/functions.go:78 | src/compiler/core.ts:399:17 |
| 1 | L | a function returning U[] | internal/lower/functions.go:78 | src/compiler/core.ts:494:17 |
| 1 | L | an array of unknown | internal/lower/object.go | src/compiler/core.ts:684:12 |
| 1 | L | a value of type SortedArray&lt;T&gt; | internal/lower/object.go | src/compiler/core.ts:769:5 |
| 1 | L | overload 1 of sortAndDeduplicate with additional implementation type parameters | internal/lower/functions.go | src/compiler/core.ts:805:1 |
| 1 | L | slice with an index that isn't a number | internal/lower/object.go | src/compiler/core.ts:987:34 |
| 1 | L | Array as a value outside equality or typeof (overloaded calls and static properties need their own representation) | internal/lower (syntax / binding / ABI guard) | src/compiler/debug.ts:1132:82 |
| 1 | L | a value of type T &#124; null &#124; undefined | internal/lower/object.go | src/compiler/debug.ts:255:37 |
| 1 | L | a function returning A | internal/lower/functions.go:78 | src/compiler/debug.ts:268:21 |
| 1 | L | reading console | internal/lower/expression.go | src/compiler/debug.ts:865:16 |
| 1 | L | a field of type "circularity" &#124; boolean | internal/lower/object.go | src/compiler/debug.ts:949:55 |
| 1 | L | reading printList | internal/lower/expression.go | src/compiler/emitter.ts:1287:9 |
| 1 | L | a value of type NodeArray&lt;T&gt; | internal/lower/object.go | src/compiler/emitter.ts:1348:60 |
| 1 | L | a value of type (s: string) =&gt; void | internal/lower/object.go | src/compiler/emitter.ts:4540:65 |
| 1 | L | a value of type readonly Child[] | internal/lower/object.go | src/compiler/emitter.ts:4729:102 |
| 1 | L | a value of type Child | internal/lower/object.go | src/compiler/emitter.ts:4754:19 |
| 1 | L | .length on a value | internal/lower/expression.go | src/compiler/emitter.ts:6375:12 |
| 1 | L | a field of type "boolean" &#124; "list" &#124; "number" &#124; "object" &#124; "string" &#124; Map&lt;string, string &#124; number&gt; | internal/lower/object.go | src/compiler/executeCommandLine.ts:375:21 |
| 1 | L | a function returning void &#124; WatchOfConfigFile&lt;EmitAndSemanticDiagnosticsBuilderProgram&gt; | internal/lower/functions.go:78 | src/compiler/executeCommandLine.ts:563:10 |
| 1 | L | a function returning void &#124; SolutionBuilder&lt;EmitAndSemanticDiagnosticsBuilderProgram&gt; &#124; WatchOfConfigFile&lt;EmitAndSemanticDiagnosticsBuilderProgram&gt; | internal/lower/functions.go:78 | src/compiler/executeCommandLine.ts:756:17 |
| 1 | L | a function returning void &#124; SolutionBuilder&lt;EmitAndSemanticDiagnosticsBuilderProgram&gt; | internal/lower/functions.go:78 | src/compiler/executeCommandLine.ts:812:10 |
| 1 | L | a value of type PrimitiveLiteral | internal/lower/object.go | src/compiler/expressionToTypeNode.ts:1220:39 |
| 1 | L | a function returning R | internal/lower/functions.go:78 | src/compiler/expressionToTypeNode.ts:880:14 |
| 1 | L | storing string &#124; number in a field | internal/lower/class.go | src/compiler/factory/emitNode.ts:242:5 |
| 1 | L | .hasTrailingComma on a value | internal/lower/expression.go | src/compiler/factory/nodeFactory.ts:1174:51 |
| 1 | L | overload 1 of createToken with additional implementation type parameters | internal/lower/functions.go | src/compiler/factory/nodeFactory.ts:1441:5 |
| 1 | L | a value of type TKind | internal/lower/object.go | src/compiler/factory/nodeFactory.ts:2280:73 |
| 1 | L | a value of type undefined | internal/lower/object.go | src/compiler/factory/nodeFactory.ts:2357:33 |
| 1 | L | a value of type string &#124; object &#124; undefined | internal/lower/object.go | src/compiler/factory/nodeFactory.ts:3531:13 |
| 1 | L | reading createNodeArray | internal/lower/expression.go | src/compiler/factory/nodeFactory.ts:522:9 |
| 1 | L | a function returning T &#124; Identifier | internal/lower/functions.go:78 | src/compiler/factory/nodeFactory.ts:7141:14 |
| 1 | L | a function returning T &#124; NumericLiteral &#124; StringLiteral &#124; BooleanLiteral | internal/lower/functions.go:78 | src/compiler/factory/nodeFactory.ts:7146:14 |
| 1 | L | a value of type TKind &#124; Token&lt;TKind&gt; | internal/lower/object.go | src/compiler/factory/nodeFactory.ts:7157:48 |
| 1 | L | a function returning T &#124; EmptyStatement &#124; undefined | internal/lower/functions.go:78 | src/compiler/factory/nodeFactory.ts:7163:14 |
| 1 | L | a function returning string &#124; object | internal/lower/functions.go:78 | src/compiler/factory/nodeFactory.ts:7238:10 |
| 1 | R | for...in over an array (holes and own enumerable properties are not represented; use for...of for elements) | internal/lower/library_for_in.go:17 | src/compiler/factory/nodeFactory.ts:7538:23 |
| 1 | L | an array of TState | internal/lower/object.go | src/compiler/factory/utilities.ts:1396:9 |
| 1 | L | overload 1 of createBinaryExpressionTrampoline with additional implementation type parameters | internal/lower/functions.go | src/compiler/factory/utilities.ts:1436:1 |
| 1 | L | a function returning TResult | internal/lower/functions.go:78 | src/compiler/factory/utilities.ts:1475:14 |
| 1 | L | a function returning NodeArray&lt;T&gt; &#124; undefined | internal/lower/functions.go:78 | src/compiler/factory/utilities.ts:1507:17 |
| 1 | L | a function returning (AssignmentExpression&lt;EqualsToken&gt; &amp; { readonly left: GeneratedIdentifier; }) &#124; undefined | internal/lower/functions.go:78 | src/compiler/factory/utilities.ts:1673:17 |
| 1 | L | a value of type AccessorDeclaration &amp; { readonly name: BigIntLiteral &#124; ComputedPropertyName &#124; Identifier &#124; NoSubstitutionTemplateLiteral &#124; NumericLiteral &#124; StringLiteral; } | internal/lower/object.go | src/compiler/factory/utilities.ts:362:107 |
| 1 | L | a value of type RedirectsCacheKey &#124; undefined | internal/lower/object.go | src/compiler/moduleNameResolver.ts:1031:15 |
| 1 | L | a Map of RedirectsCacheKey | internal/lower/object.go | src/compiler/moduleNameResolver.ts:1034:9 |
| 1 | L | a function returning RedirectsCacheKey | internal/lower/functions.go:78 | src/compiler/moduleNameResolver.ts:1042:14 |
| 1 | L | a function returning PerDirectoryResolutionCache&lt;T&gt; | internal/lower/functions.go:78 | src/compiler/moduleNameResolver.ts:1078:10 |
| 1 | L | a function returning ModeAwareCache&lt;T&gt; | internal/lower/functions.go:78 | src/compiler/moduleNameResolver.ts:1119:17 |
| 1 | L | a function returning NonRelativeNameResolutionCache&lt;T&gt; | internal/lower/functions.go:78 | src/compiler/moduleNameResolver.ts:1166:10 |
| 1 | L | a function returning ModuleOrTypeReferenceResolutionCache&lt;T&gt; | internal/lower/functions.go:78 | src/compiler/moduleNameResolver.ts:1276:10 |
| 1 | L | a field of type false &#124; string[] &#124; undefined | internal/lower/object.go | src/compiler/moduleNameResolver.ts:2235:23 |
| 1 | L | storing false &#124; string[] &#124; undefined in a field | internal/lower/class.go | src/compiler/moduleNameResolver.ts:2277:12 |
| 1 | L | a field of type false &#124; VersionPaths &#124; undefined | internal/lower/object.go | src/compiler/moduleNameResolver.ts:2408:9 |
| 1 | L | a field of type false &#124; VersionPaths | internal/lower/object.go | src/compiler/moduleNameResolver.ts:2411:12 |
| 1 | L | a field of type string &#124; false &#124; undefined | internal/lower/object.go | src/compiler/moduleNameResolver.ts:2415:9 |
| 1 | L | a field of type string &#124; false | internal/lower/object.go | src/compiler/moduleNameResolver.ts:2418:12 |
| 1 | L | a field of type string &#124; true &#124; undefined | internal/lower/object.go | src/compiler/moduleNameResolver.ts:241:10 |
| 1 | L | a function returning CacheWithRedirects&lt;K, V&gt; | internal/lower/functions.go:78 | src/compiler/moduleNameResolver.ts:977:10 |
| 1 | L | a tuple literal leaving out an element of type ModuleSpecifierEnding | internal/lower/object.go | src/compiler/moduleSpecifiers.ts:1010:103 |
| 1 | L | a value of type (ModuleDeclaration &amp; { name: StringLiteral; }) &#124; undefined | internal/lower/object.go | src/compiler/moduleSpecifiers.ts:880:11 |
| 1 | L | a value of type (AmbientModuleDeclaration &amp; { name: StringLiteral; }) &#124; undefined | internal/lower/object.go | src/compiler/moduleSpecifiers.ts:920:11 |
| 1 | L | a value of type PragmaPseudoMap[TKey][] &#124; PragmaPseudoMap[TKey] | internal/lower/object.go | src/compiler/parser.ts:10625:31 |
| 1 | L | a function value taking union of differently held members | internal/lower/functions.go | src/compiler/parser.ts:1773:28 |
| 1 | L | a function returning NodeArray&lt;NonNullable&lt;T&gt;&gt; &#124; undefined | internal/lower/functions.go:78 | src/compiler/parser.ts:3492:14 |
| 1 | L | a function returning MissingList&lt;T&gt; | internal/lower/functions.go:78 | src/compiler/parser.ts:3569:14 |
| 1 | L | a function returning ExpressionWithTypeArguments &amp; { expression: Identifier &#124; PropertyAccessEntityNameExpression; } | internal/lower/functions.go:78 | src/compiler/parser.ts:9568:22 |
| 1 | L | overload 1 of forEachAncestorDirectory with additional implementation type parameters | internal/lower/functions.go | src/compiler/path.ts:1092:1 |
| 1 | L | a value of type ResolvedModuleWithFailedLookupLocations &amp; ResolvedTypeReferenceDirectiveWithFailedLookupLocations | internal/lower/object.go | src/compiler/program.ts:1015:7 |
| 1 | L | a value of type Map&lt;Path, ModeAwareCache&lt;T&gt;&gt; &#124; undefined | internal/lower/object.go | src/compiler/program.ts:2007:9 |
| 1 | L | overload 1 of resolveTypeReferenceDirectiveNamesReusingOldState with additional implementation type parameters | internal/lower/functions.go | src/compiler/program.ts:2191:5 |
| 1 | L | a function returning readonly Resolution[] | internal/lower/functions.go:78 | src/compiler/program.ts:2231:14 |
| 1 | L | a value of type SourceFileOrString | internal/lower/object.go | src/compiler/program.ts:2233:9 |
| 1 | L | reading diagnostics | internal/lower/expression.go | src/compiler/program.ts:3049:25 |
| 1 | L | a call returning Path | internal/lower (syntax / binding / ABI guard) | src/compiler/program.ts:4974:51 |
| 1 | L | reading setReadFileCache | internal/lower/expression.go | src/compiler/program.ts:539:16 |
| 1 | L | a function value taking CreateSourceFileOptions &#124; ScriptTarget | internal/lower/functions.go | src/compiler/program.ts:558:106 |
| 1 | L | a value of type ReferencedFile &amp; { kind: FileIncludeKind.LibReferenceDirective; } | internal/lower/object.go | src/compiler/programDiagnostics.ts:190:74 |
| 1 | L | a Map whose key and value types aren't known | internal/lower/object.go | src/compiler/resolutionCache.ts:1478:9 |
| 1 | L | reading getCurrentDirectory | internal/lower/expression.go | src/compiler/resolutionCache.ts:1686:60 |
| 1 | L | a value of type PathPathComponents | internal/lower/object.go | src/compiler/resolutionCache.ts:632:11 |
| 1 | L | destructuring a string | internal/lower/object.go | src/compiler/scanner.ts:1333:15 |
| 1 | L | overload 1 of forEachLeadingCommentRange with additional implementation type parameters | internal/lower/functions.go | src/compiler/scanner.ts:932:1 |
| 1 | L | overload 1 of forEachTrailingCommentRange with additional implementation type parameters | internal/lower/functions.go | src/compiler/scanner.ts:938:1 |
| 1 | L | a case that isn't a constant | internal/lower (syntax / binding / ABI guard) | src/compiler/semver.ts:380:13 |
| 1 | L | a value of type string &#124; null | internal/lower/object.go | src/compiler/sourcemap.ts:102:52 |
| 1 | L | apply without a dense argument literal (length, presence and argument representations must be proven) | internal/lower (syntax / binding / ABI guard) | src/compiler/sourcemap.ts:315:25 |
| 1 | L | reading file | internal/lower/expression.go | src/compiler/sourcemap.ts:325:13 |
| 1 | L | a non-accessor method in an accessor literal | internal/lower (syntax / binding / ABI guard) | src/compiler/sourcemap.ts:483:9 |
| 1 | L | reading fileSystemEntryExists | internal/lower/expression.go | src/compiler/sys.ts:1119:17 |
| 1 | L | reading byteOrderMarkIndicator | internal/lower/expression.go | src/compiler/sys.ts:1823:24 |
| 1 | L | a call to an Identifier | internal/lower (syntax / binding / ABI guard) | src/compiler/sys.ts:1920:24 |
| 1 | L | reading Debug | internal/lower/expression.go | src/compiler/sys.ts:1981:5 |
| 1 | L | reading pollScheduled | internal/lower/expression.go | src/compiler/sys.ts:490:37 |
| 1 | L | a spread that adds a field the source doesn't have | internal/lower (syntax / binding / ABI guard) | src/compiler/tracing.ts:183:53 |
| 1 | L | a value of type CustomTransformerFactory &#124; TransformerFactory&lt;T&gt; | internal/lower/object.go | src/compiler/transformer.ts:209:70 |
| 1 | L | a function returning TransformationResult&lt;T&gt; | internal/lower/functions.go:78 | src/compiler/transformer.ts:248:17 |
| 1 | L | a value of type ClassThisAssignmentBlock &#124; undefined | internal/lower/object.go | src/compiler/transformers/classFields.ts:2183:19 |
| 1 | L | a value of type ClassNamedEvaluationHelperBlock &#124; undefined | internal/lower/object.go | src/compiler/transformers/classFields.ts:2184:19 |
| 1 | L | a value of type PrivateIdentifierInExpression | internal/lower/object.go | src/compiler/transformers/classFields.ts:671:55 |
| 1 | L | a function returning ClassThisAssignmentBlock | internal/lower/functions.go:78 | src/compiler/transformers/classThis.ts:33:10 |
| 1 | L | a function returning never | internal/lower/functions.go:78 | src/compiler/transformers/declarations.ts:260:29 |
| 1 | L | reading errorFallbackNode | internal/lower/expression.go | src/compiler/transformers/declarations.ts:288:37 |
| 1 | L | a function returning T &#124; StringLiteral | internal/lower/functions.go:78 | src/compiler/transformers/declarations.ts:835:14 |
| 1 | L | a value of type ArrowFunction &#124; BinaryExpression &#124; BindingElement &#124; Block &#124; BreakStatement &#124; CallSignatureDeclaration &#124; ... 64 more ... &#124; EndOfFileToken | internal/lower/object.go | src/compiler/transformers/declarations/diagnostics.ts:567:32 |
| 1 | L | a value of type TransformedSuperCall | internal/lower/object.go | src/compiler/transformers/es2015.ts:1399:20 |
| 1 | L | a function returning SyntheticSuper | internal/lower/functions.go:78 | src/compiler/transformers/es2015.ts:4823:14 |
| 1 | L | a value of type PropertyAccessExpression &#124; SyntheticSuper | internal/lower/object.go | src/compiler/transformers/es2015.ts:4831:15 |
| 1 | L | a function returning CapturedThis | internal/lower/functions.go:78 | src/compiler/transformers/es2015.ts:825:14 |
| 1 | L | a function returning ClassExpression &#124; ImmediatelyInvokedArrowFunction | internal/lower/functions.go:78 | src/compiler/transformers/esDecorators.ts:1125:14 |
| 1 | L | a function returning { modifiers: NodeArray&lt;Modifier&gt; &#124; undefined; referencedName: Expression &#124; undefined; name: PropertyName; initializersName: Identifier &#124; undefined; descriptorName: Identifier &#124; undefined; thisArg: Identifier &#124; undefined; extraInitializersName?: never; } &#124; ... | internal/lower/functions.go:78 | src/compiler/transformers/esDecorators.ts:1236:14 |
| 1 | L | a value of type TNode | internal/lower/object.go | src/compiler/transformers/esDecorators.ts:1239:9 |
| 1 | L | a function returning ImmediatelyInvokedArrowFunction | internal/lower/functions.go:78 | src/compiler/transformers/esDecorators.ts:670:14 |
| 1 | L | a value of type VariableDeclarationList &amp; { _usingBrand: void; } | internal/lower/object.go | src/compiler/transformers/esnext.ts:303:19 |
| 1 | L | a value of type (VariableDeclaration &amp; { name: Identifier; }) &#124; undefined | internal/lower/object.go | src/compiler/transformers/jsx.ts:110:13 |
| 1 | L | a value of type VariableDeclaration &amp; { name: Identifier; } | internal/lower/object.go | src/compiler/transformers/jsx.ts:115:16 |
| 1 | L | a DeleteExpression as a statement | internal/lower (syntax / binding / ABI guard) | src/compiler/transformers/module/system.ts:1779:17 |
| 1 | L | a function returning ClassNamedEvaluationHelperBlock | internal/lower/functions.go:78 | src/compiler/transformers/namedEvaluation.ts:101:10 |
| 1 | L | a value of type ClassNamedEvaluationHelperBlock | internal/lower/object.go | src/compiler/transformers/namedEvaluation.ts:203:11 |
| 1 | L | a value of type PropertyAssignment &amp; { readonly name: Identifier; readonly initializer: WrappedExpression&lt;AnonymousFunctionDefinition&gt;; } | internal/lower/object.go | src/compiler/transformers/namedEvaluation.ts:256:87 |
| 1 | L | a value of type ShorthandPropertyAssignment &amp; { readonly objectAssignmentInitializer: WrappedExpression&lt;AnonymousFunctionDefinition&gt;; } | internal/lower/object.go | src/compiler/transformers/namedEvaluation.ts:274:96 |
| 1 | L | a value of type VariableDeclaration &amp; { readonly name: Identifier; readonly initializer: WrappedExpression&lt;AnonymousFunctionDefinition&gt;; } | internal/lower/object.go | src/compiler/transformers/namedEvaluation.ts:294:88 |
| 1 | L | a value of type ParameterDeclaration &amp; { readonly name: Identifier; readonly dotDotDotToken: undefined; readonly initializer: WrappedExpression&lt;AnonymousFunctionDefinition&gt;; } | internal/lower/object.go | src/compiler/transformers/namedEvaluation.ts:322:89 |
| 1 | L | a value of type BindingElement &amp; { readonly name: Identifier; readonly dotDotDotToken: undefined; readonly initializer: WrappedExpression&lt;AnonymousFunctionDefinition&gt;; } | internal/lower/object.go | src/compiler/transformers/namedEvaluation.ts:354:83 |
| 1 | L | a value of type PropertyDeclaration &amp; { readonly initializer: WrappedExpression&lt;AnonymousFunctionDefinition&gt;; } | internal/lower/object.go | src/compiler/transformers/namedEvaluation.ts:384:88 |
| 1 | L | a value of type (AssignmentExpression&lt;EqualsToken&gt; &amp; { readonly left: Identifier; readonly right: WrappedExpression&lt;AnonymousFunctionDefinition&gt;; } &amp; BinaryExpression) &#124; (... &amp; ... 1 more ... &amp; BinaryExpression) | internal/lower/object.go | src/compiler/transformers/namedEvaluation.ts:405:89 |
| 1 | L | a value of type ExportAssignment &amp; { readonly expression: WrappedExpression&lt;AnonymousFunctionDefinition&gt;; } | internal/lower/object.go | src/compiler/transformers/namedEvaluation.ts:444:85 |
| 1 | L | a value of type NamedEvaluation | internal/lower/object.go | src/compiler/transformers/namedEvaluation.ts:470:74 |
| 1 | L | overload 1 of setSerializerContextAnd with additional implementation type parameters | internal/lower/functions.go | src/compiler/transformers/typeSerializer.ts:153:5 |
| 1 | L | a function returning V[] | internal/lower/functions.go:78 | src/compiler/transformers/utilities.ts:377:10 |
| 1 | L | a value of type V | internal/lower/object.go | src/compiler/transformers/utilities.ts:377:61 |
| 1 | L | a function returning V &#124; undefined | internal/lower/functions.go:78 | src/compiler/transformers/utilities.ts:400:5 |
| 1 | L | a base that isn't a declared class | internal/lower (syntax / binding / ABI guard) | src/compiler/transformers/utilities.ts:441:33 |
| 1 | L | a value of type TData | internal/lower/object.go | src/compiler/transformers/utilities.ts:824:54 |
| 1 | L | a function returning TEntry &#124; undefined | internal/lower/functions.go:78 | src/compiler/transformers/utilities.ts:829:17 |
| 1 | L | a value of type PrivateEnvironment&lt;TData, TEntry&gt; | internal/lower/object.go | src/compiler/transformers/utilities.ts:840:5 |
| 1 | L | a value of type TEntry | internal/lower/object.go | src/compiler/transformers/utilities.ts:842:5 |
| 1 | L | a function returning TPrivateEntry &#124; undefined | internal/lower/functions.go:78 | src/compiler/transformers/utilities.ts:855:17 |
| 1 | L | reading program | internal/lower/expression.go | src/compiler/tsbuildPublic.ts:1085:25 |
| 1 | L | a function returning BuildInvalidedProject&lt;T&gt; &#124; UpdateOutputFileStampsProject | internal/lower/functions.go:78 | src/compiler/tsbuildPublic.ts:1316:10 |
| 1 | L | a function returning InvalidatedProject&lt;T&gt; &#124; undefined | internal/lower/functions.go:78 | src/compiler/tsbuildPublic.ts:1341:10 |
| 1 | L | a value of type ResolvedConfigFileName &#124; undefined | internal/lower/object.go | src/compiler/tsbuildPublic.ts:1393:214 |
| 1 | L | storing any in a field | internal/lower/class.go | src/compiler/tsbuildPublic.ts:2080:5 |
| 1 | L | a function returning SolutionBuilder&lt;T&gt; | internal/lower/functions.go:78 | src/compiler/tsbuildPublic.ts:2261:10 |
| 1 | L | a function returning ResolvedConfigFilePath | internal/lower/functions.go:78 | src/compiler/tsbuildPublic.ts:559:10 |
| 1 | L | a field of type AnyBuildOrder &#124; undefined | internal/lower/object.go | src/compiler/tsbuildPublic.ts:656:12 |
| 1 | L | a Map of ResolvedConfigFilePath | internal/lower/object.go | src/compiler/tsbuildPublic.ts:663:5 |
| 1 | L | a Set of ResolvedConfigFilePath (a Set holds strings, numbers, booleans, objects, arrays, maps or functions so far) | internal/lower/object.go | src/compiler/tsbuildPublic.ts:666:29 |
| 1 | L | a value of type "" &#124; ResolvedConfigFileName &#124; undefined | internal/lower/object.go | src/compiler/tsbuildPublic.ts:723:11 |
| 1 | L | a function returning BuildInvalidedProject&lt;T&gt; | internal/lower/functions.go:78 | src/compiler/tsbuildPublic.ts:932:10 |
| 1 | L | a destructured parameter beside a parameter with a default | internal/lower (syntax / binding / ABI guard) | src/compiler/utilities.ts:10053:1 |
| 1 | L | a function returning { readonly min: number; readonly max: number; } | internal/lower/functions.go:78 | src/compiler/utilities.ts:10362:17 |
| 1 | L | typed array element type Uint16Array | internal/lower/typed_arrays.go:61 (historical); current kind supported | src/compiler/utilities.ts:10491:22 |
| 1 | L | a function returning EvaluatorResult&lt;T&gt; | internal/lower/functions.go:78 | src/compiler/utilities.ts:11311:17 |
| 1 | L | a value of type UnaryExpression &amp; (BigIntLiteral &#124; NumericLiteral) | internal/lower/object.go | src/compiler/utilities.ts:12043:41 |
| 1 | L | a value of type UnaryExpression &amp; NumericLiteral | internal/lower/object.go | src/compiler/utilities.ts:12046:41 |
| 1 | L | a value of type IncludeTypeSpaceImports | internal/lower/object.go | src/compiler/utilities.ts:12174:5 |
| 1 | L | a value of type RequireOrImportCall | internal/lower/object.go | src/compiler/utilities.ts:3887:16 |
| 1 | L | a value of type AccessExpression &#124; RequireOrImportCall | internal/lower/object.go | src/compiler/utilities.ts:3890:54 |
| 1 | L | a function returning MemberName &#124; (Expression &amp; (NumericLiteral &#124; StringLiteralLike)) | internal/lower/functions.go:78 | src/compiler/utilities.ts:4171:17 |
| 1 | L | a value of type JSDocImportTag &#124; CanHaveModuleSpecifier | internal/lower/object.go | src/compiler/utilities.ts:4349:54 |
| 1 | L | a function returning AnyValidImportOrReExport | internal/lower/functions.go:78 | src/compiler/utilities.ts:4376:17 |
| 1 | L | a function returning AnyValidImportOrReExport &#124; undefined | internal/lower/functions.go:78 | src/compiler/utilities.ts:4381:17 |
| 1 | L | a function returning HasJSDoc &#124; undefined | internal/lower/functions.go:78 | src/compiler/utilities.ts:4812:17 |
| 1 | L | a value of type TypeParameterDeclaration &amp; { parent: JSDocTemplateTag; } | internal/lower/object.go | src/compiler/utilities.ts:4830:43 |
| 1 | L | a function returning ExpressionWithTypeArguments &amp; { readonly expression: Identifier &#124; PropertyAccessEntityNameExpression; } | internal/lower/functions.go:78 | src/compiler/utilities.ts:5216:49 |
| 1 | L | a function returning (ConstructorDeclaration &amp; { body: Block; }) &#124; undefined | internal/lower/functions.go:78 | src/compiler/utilities.ts:6759:17 |
| 1 | L | a function returning object &#124; undefined | internal/lower/functions.go:78 | src/compiler/utilities.ts:7786:17 |
| 1 | L | a PostfixUnaryExpression | internal/lower/expression.go | src/compiler/utilities.ts:7993:12 |
| 1 | L | a value of type { forEach: (callbackfn: (value: T, key: K, map: Map&lt;K, T&gt;) =&gt; void, thisArg?: any) =&gt; void; clear: () =&gt; void; } | internal/lower/object.go | src/compiler/utilities.ts:8172:32 |
| 1 | L | overload 1 of mutateMapSkippingNewValues with additional implementation type parameters | internal/lower/functions.go | src/compiler/utilities.ts:8199:1 |
| 1 | L | overload 1 of mutateMap with additional implementation type parameters | internal/lower/functions.go | src/compiler/utilities.ts:8248:1 |
| 1 | L | a value of type Set&lt;K&gt; | internal/lower/object.go | src/compiler/utilities.ts:8316:30 |
| 1 | L | Object.assign on a shape not proven by a plain literal or its const binding | internal/lower/library_json_stringify.go; internal/lower/class_features.go | src/compiler/utilities.ts:8578:19 |
| 1 | L | a value of type readonly K[] | internal/lower/object.go | src/compiler/utilities.ts:931:5 |
| 1 | L | a function returning unknown | internal/lower/functions.go:78 | src/compiler/utilities.ts:9399:17 |
| 1 | L | a value of type NonNullable&lt;K&gt; | internal/lower/object.go | src/compiler/utilities.ts:940:15 |
| 1 | L | a value of type boolean &#124; V &#124; undefined | internal/lower/object.go | src/compiler/utilities.ts:942:15 |
| 1 | L | a parameter that isn't a plain name | internal/lower (syntax / binding / ABI guard) | src/compiler/utilities.ts:9699:5 |
| 1 | L | a value of type EntityNameExpression &#124; (LeftHandSideExpression &amp; BindableStaticNameExpression) | internal/lower/object.go | src/compiler/utilitiesPublic.ts:1759:79 |
| 1 | L | overload 1 of getOriginalNode with additional implementation type parameters | internal/lower/functions.go | src/compiler/utilitiesPublic.ts:762:1 |
| 1 | L | a value of type (element: Node) =&gt; "quit" &#124; boolean | internal/lower/object.go | src/compiler/utilitiesPublic.ts:788:54 |
| 1 | L | a value of type WatchFactoryHost &amp; { trace?(s: string): void; } | internal/lower/object.go | src/compiler/watch.ts:745:51 |
| 1 | L | a function returning WatchCompilerHostOfConfigFile&lt;T&gt; | internal/lower/functions.go:78 | src/compiler/watch.ts:923:17 |
| 1 | L | a function returning WatchCompilerHostOfFilesAndCompilerOptions&lt;T&gt; | internal/lower/functions.go:78 | src/compiler/watch.ts:955:17 |
| 1 | L | a value of type WatchCompilerHostOfFilesAndCompilerOptionsOrConfigFile&lt;T&gt; | internal/lower/object.go | src/compiler/watchPublic.ts:420:62 |
| 1 | L | reading compilerHost | internal/lower/expression.go | src/compiler/watchPublic.ts:695:9 |
| 1 | L | reading timerToInvalidateFailedLookupResolutions | internal/lower/expression.go | src/compiler/watchPublic.ts:856:14 |
| 1 | L | reading timerToUpdateProgram | internal/lower/expression.go | src/compiler/watchPublic.ts:886:13 |
| 1 | L | a value of type Map&lt;string, WildcardDirectoryWatcher&lt;T&gt;&gt; | internal/lower/object.go | src/compiler/watchUtilities.ts:515:5 |
| 1 | L | a call returning T | internal/lower (syntax / binding / ABI guard) | src/compiler/watchUtilities.ts:540:22 |
| 1 | L | a field of type boolean &#124; (() =&gt; boolean) | internal/lower/object.go | src/compiler/watchUtilities.ts:741:23 |

## Validation and scope

No production compiler/runtime source, oracle fixture or adapted stage 3 source
changed. No implementation was attempted because the highest-ranked item is not
contained. A Node fixture and semantic mutant apply when that implementation is
undertaken; inventing one for this documentation would not validate a behavior
change. No fixture-count refresh applies.

Setup passed with `export GOPROXY='https://proxy.golang.org|direct'` and
`bash cloud/setup.sh --wasi-sdk`; sourced `/workspace/adamic-tools/env.sh`.
Node v24.19.0, Go 1.27.1, clang 20.1.8. No whole-package gate was run.
The document accounting audit checks exact reason/count coverage against the
pinned CSV, deduplicated signatures, 6,703 sites, 472 groups, and the runtime
subset. A +1 count mutant, a dropped-reason mutant and a wrong-owner mutant
all fail that audit.
Local evidence and audit: `/workspace/scratch/stage3-runtime-roots/`.
