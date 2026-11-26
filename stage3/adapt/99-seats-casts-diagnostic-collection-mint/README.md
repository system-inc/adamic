Adamic ruled this shape on Nov 5 2026 (#b5w3ycg, the brand mint's `as unknown as` spelling, `docs/rulings.md:33`), as `99-correctness-axolotl-r` applied it; Compiler accepts this type-only head under the split, after its review and census.

At `src/compiler/utilities.ts:6100:31` in the `-before` snapshot at `ef741d4f`, unit `createDiagnosticCollection` builds an empty array in its local `add`. The adapter pins both function declarations and the exact cast. It refuses a second array literal in `add` in any spelling, a cast inside a function or a class nested in `add` (a field initializer or a static block included), and an assignment to any `diagnostics` but `add`'s own.

```ts
[] as Diagnostic[] as SortedArray<DiagnosticWithLocation>
[] as unknown as SortedArray<DiagnosticWithLocation>
```

`src/compiler/corePublic.ts:21` gives `SortedArray<T>` one base and one own member.

```ts
export interface SortedArray<T> extends Array<T> {
    " __sortedArrayBrand": undefined;
}
```

Under `docs/rulings.md:33`, the fresh empty literal relates to `Array<DiagnosticWithLocation>` since it contains zero elements and no spreads. The relation reads the innermost literal's type. The intermediate `unknown` supplies the mint spelling. The only own member is a required string-keyed phantom of type `undefined` under `docs/rulings.md:32`. The interface adds no method, signature, index entry or redeclaration. The map declared at `utilities.ts:6064` holds `Map<string, SortedArray<DiagnosticWithLocation>>`. The cast erases.

The next line at `utilities.ts:6101` casts the existing `diagnostics` binding to `SortedArray<DiagnosticWithLocation>`. That is the unit's next stop. The unit census at main `0e7517ac` (J-x1hz3wy entry, J-vpytdg8 directory) reads it as Refused, "a cast the runtime can't check", at `src/compiler/utilities.ts:6137:63` in the adapted tree. That is casts family 2's reason, in unit `createDiagnosticCollection`. It is an exposed stop, upstream's bytes and one of the base's own sites, and the census reads 0 units newly stopping. The adapter leaves that cast, the GH#19873 comment, and the earlier readonly `lookup` edit intact. The scout census names neither of the two casts initializing `nonFileDiagnostics` and `filesWithDiagnostics`, so both stay in place.
