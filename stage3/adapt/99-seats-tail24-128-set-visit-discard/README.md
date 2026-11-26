Compiler accepts this JavaScript-changing adaptation after its review and census, by Adamic's ruling that it is JavaScript-changing in a shape now ruled twice (#bbssdcy, Nov 26 2026 19:38:08Z; the tail pass's family 128, 91-patcher-discard's shape over a built-in Set).

The adapter locates the named worker and pins the exact CRLF text of its direct Set visit statement. Stock TypeScript resolves the receiver, the stock Set method and the callback result before the write. The block keeps the original call and argument bytes. `adapt.cjs:58-64` checks the receiver’s static type and resolution to the stock `Set.forEach` method, while the receiver’s built-in `Set` identity at runtime rests on the producer table and must be rechecked if any listed producer changes.

Before

```ts
state.referencedMap!.getKeys(filePath)?.forEach(referencingFilePath =>
        handleDtsMayChangeOfFileAndExportsOfFile(
            state,
            referencingFilePath,
            invalidateJsFiles,
            seenFileAndExportsOfFile,
            cancellationToken,
            host,
        )
    );
```

After

```ts
state.referencedMap!.getKeys(filePath)?.forEach(referencingFilePath =>
        { handleDtsMayChangeOfFileAndExportsOfFile(
            state,
            referencingFilePath,
            invalidateJsFiles,
            seenFileAndExportsOfFile,
            cancellationToken,
            host,
        ); }
    );
```

## Set producer proof

| Head 128 producer | Pin source | Result |
|---|---|---|
| Fresh program state | `builder.ts:324,1714`, `builderState.ts:299-302,311,324-327,341` | New map from the factory or undefined. Old state contributes comparison/signature data and does not supply the new reverse map. |
| Restored incremental state | `builder.ts:2317,2345,2395-2402` | File IDs become Sets, handed to `set` on a map produced by the same factory. |
| Private reverse map | `builderState.ts:123-126,157,168,172-176` | Factory makes the reverse Map. Its only insertion route allocates a native `Set` and adds a path. `getKeys` returns `reverse.get(v)`, so a missing entry is undefined. |
| Build-info serialization | `builder.ts:1355-1360,1392` | The similarly named binding holds a serialization array, consumed through the restoration described above. |

Read the factory and searched the adapted `src/` for the factory, assignments, and producers. The private reverse map never stores the caller's forward Set or an arbitrary Set-like object. ECMA-262 [Set.prototype.forEach](https://tc39.es/ecma262/#sec-set.prototype.foreach) calls in insertion order, propagates abrupt completion, and ignores callback results. Adapter 91's custom patcher traversal was examined as a syntax precedent; its truthy early-exit behavior is a different semantic contract.

The adaptation uses the block-arrow precedent in `91-patcher-discard`. The current visit resolves to the built-in Set method. Its discarded result is independent of the truthy early exit in adapter 91's custom traversal.

JavaScript-changing, outside the contract (Adamic, #bbssdcy, Nov 26 2026 19:38:08Z, as 91's reasoning applied): a caller that reaches tsc's live reverse `Set` through the `@internal` `BuilderProgram.state` (upstream `builderPublic.ts:47`, with no public declaration, and no JavaScript object in the native tsc binary) and attaches an own `forEach` that reads the callback's result could stop early on the worker's `true` before this head and sees `undefined` after it, as review-zb read 3's probe `runtime-scope-emitted.cjs` shows in its output `runtime-scope-emitted.out` (with the own method, `/second.ts` is skipped before and visited after; with the stock method, every file is visited on both sides).

## Rulings
- **An optional call short-circuits the whole chain; the function read before the arguments is the one called; a present non-callable value is proven away in .a and exit 70 in .ts after the arguments are evaluated.** Meaning, Check in .ts. Oct 9 2026, @system_adamic; source: #tvq1eqm (zvy4sg4d). Spec: What's in, Values and statements (`?.`). Test: NONE for the check, which nothing reaches yet: its one road, a cast to a checked view whose real object holds a non-function where the type says a method, stops at the view's read (internal/oracle/optional_method_call_test.go:48 TestOptionalMethodCallThroughAViewIsNotYet, on testdata/optional_method_call/unproven/view_hatch.ts, where Node stops with `TypeError: viewed.enabled is not a function`, exit 70; internal/lower/optional_method_call_test.go:74, its last row). An optional method the program declares, called as `object?.method?.(...)` (tsc's performance.ts:189 shape) or `object.method?.(...)`, lowers (Nov 7 2026, #tvq1eqm's head): internal/lower/optional_method_call_test.go:41 TestOptionalMethodCallsLower and :74 TestOptionalMethodCallGapsStop for what keeps its stop; internal/oracle/optional_method_call_test.go:19 TestOptionalMethodCallsAgreeWithNode, testdata/optional_method_call/*.a held to Node natively, sanitized and on JavaScript under the leak check. Corrected Nov 23 2026, as @system_adamic asked on #cdpxjmb, from Language's measurement at main 96a1291b (J-tq7jtfk, held to Node natively and on JavaScript): an optional call on a local function value (`f?.('ada')`), a receiver-optional library intrinsic (`list?.join(',')`, `word?.toUpperCase()`) and a class method on an optional receiver (`make()?.greet('ada')`) lower and agree with Node, short-circuiting to undefined on a nullish receiver. What stays NotYet on main is a call later in a chain (`box?.name.toUpperCase()`, conformance/optional/member_call.a), a call through `?.` on a call's result (`pick(true)?.('ada')`, conformance/optional/call.a) and on an element (internal/lower/lower_test.go:84, 'a function value called through ?. on an element'), each 'a call through ?. (an optional call)', the guard #cdpxjmb builds over a whole chain; optional/read.a pins that a property read through `?.` lowers.
- **A `void` value is unobservable: binding and reading one, comparing it, returning it or passing it is refused; a call through a `void` type is a statement and the discarded value is released.** Refuse. Oct 18 2026, @system_adamic (#system_adamic_language); source: refusal #x2gt39p, release #f418fbz. Spec: What's refused in 0.1 (row Reading a `void` value). Test: internal/lower/void_reads_test.go:14 TestVoidReadsAreRefused and :46 TestVoidDiscardsStayLegal (the refusal, 8d39519a, Oct 27, #x2gt39p); internal/conformance/testdata/void/read.a and read.ts refuse it, and discard_fix.a lowers; internal/lower/library_node_fs_file_test.go:41 TestNodeFSFileRefusesVoidValues for one host case. A void value inside a container seen wider is ruled on #p7t8tmh (Oct 27) and not yet built.
