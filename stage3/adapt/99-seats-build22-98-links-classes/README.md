# Temporary checker links classes

Adamic's Nov 23 answer at 09:06:19Z on #bbssdcy amends the Nov 22 ruling for family 171. Adamic accepts this adaptation after review and census.

The Nov 22 ruling at 21:49Z on #bbssdcy authorizes this temporary adaptation under docs/rulings.md:403. Family 98 stops at the module-local `SymbolLinks` class expression in `checker.ts`. It becomes a class declaration with a brand-only interface merge. Line 106 governs class identity. Line 113 holds `declare` fields NotYet. The original class expression carries `declare _symbolLinksBrand: undefined;`, and the replacement removes the field while the merged interface supplies its type. Line 127 refuses the merge in `.a` only, and line 400 governs the brand's `undefined` type. All ruling references use the lines of docs/rulings.md at 1fd1cfd2. The existing imported type keeps its meaning through the `SymbolLinksShape` alias.

The classes and their brand interfaces sit directly after checker.ts's imports, ahead of every top-level statement that runs code, so no construction can run before its class exists. Upstream NodeLinks was a hoisted function and SymbolLinks was a const. Removing their eight lines at the old location preserves the file's line count. The constructions in createSymbol, getSymbolLinks and getNodeLinks run inside createTypeChecker after module initialization. The SymbolLinks allocations keep their constructor name and allocation expressions. The constructor has no other value use and the brand has no runtime reader. Remove this adaptation when the original class expression and its type-only brand lower.

Family 171 replaces the module-local `NodeLinks` function with a class that initializes `flags`, `calculatedFlags` and `jsxFlags` to their `None` members. The merged interface carries only the brand. The two fields upstream leaves absent start at zero. The constructor is called through `new` after module initialization. Lines 282 and 429 govern the constructor receiver. Line 127 refuses the class merge in `.a` only for both names.

The adapter walks the syntax tree to match the 15 flag reads by whole-line text and nearest named function, in source order, and rejects extra field identifiers and string keys. The table records locations for review, and the pins follow reads when earlier adapters shift line numbers. A single-file lexical binding pass also follows NodeLinks records from the accessor and its freshly allocated backing array through locals, and pins the existing signature-cache and import-helper transfers. Computed record access and new whole-record uses such as spreads, key enumeration, calls and copies refuse. Returning a record from any function nested in the accessor refuses, as does `arguments` inside a function that receives a tracked record as a parameter. The guard does not model dynamic eval or ambient prototype changes.

The 15 audited reads use bitwise AND or compound OR. None compares, coalesces, spreads or enumerates keys. For the table, `m` is the right-hand mask and `p` is the value from a dominating write. The last column compares first-use values except where a write dominates the read. Prior writes remain the same because each update starts from the same converted bits.

| checker.ts line | Field and owner | Syntax role | Prior write on every path | From absent / from 0 |
|---|---|---|---| --- |
| 34040 | jsxFlags, getIntrinsicTagSymbol | `\|=` IntrinsicNamedElement | No. A fresh node enters this branch. | stores `m` / stores `m` |
| 34047 | jsxFlags, getIntrinsicTagSymbol | `\|=` IntrinsicIndexedElement | No. The named branch returns and a fresh node can take the index branch. | stores `m` / stores `m` |
| 34052 | jsxFlags, getIntrinsicTagSymbol | `\|=` IntrinsicIndexedElement | No. The other write branches return and a fresh node can take this fallback. | stores `m` / stores `m` |
| 34280 | jsxFlags, getIntrinsicAttributesTypeFromJsxOpeningLikeElement | `&` IntrinsicNamedElement | No. getIntrinsicTagSymbol can return unknownSymbol without writing, including for a missing JSX definition or intrinsic name. | `0` / `0` |
| 34283 | jsxFlags, same owner | `&` IntrinsicIndexedElement | No. The preceding test can be false on that same unwritten path. | `0` / `0` |
| 50974 | calculatedFlags, calculateNodeCheckFlagWorker | `&` flag | No. This tests a fresh allocation before the switch. | `0` / `0` |
| 51010 | calculatedFlags, checkSuperExpressions | `&` flag | No. Traversal visits fresh descendant links. | `0` / `0` |
| 51011 | calculatedFlags, checkSuperExpressions | `\|=` | No. The preceding test does not write. | stores `m` / stores `m` |
| 51022 | calculatedFlags, checkSingleSuperExpression | `\|=` | No. The direct SuperInstance/SuperStatic switch route bypasses checkSuperExpressions. | stores `m` / stores `m` |
| 51030 | calculatedFlags, checkIdentifiers | `&` flag | No. Traversal can visit a fresh descendant. | `0` / `0` |
| 51031 | calculatedFlags, checkIdentifiers | `\|=` | No. The preceding test does not write. | stores `m` / stores `m` |
| 51048 | calculatedFlags, checkSingleIdentifier | `\|=` ConstructorReference | No. The direct ConstructorReference switch route bypasses checkIdentifiers. | stores `m` / stores `m` |
| 51050 | calculatedFlags, checkSingleIdentifier | `\|=` block/capture flags | Yes. The write at 51048 uses the same nodeLinks local and dominates this branch. | stores `p OR m` / stores `p OR m` |
| 51062 | calculatedFlags, checkBlockScopeBindings | `&` flag | No. Traversal can visit a fresh descendant. | `0` / `0` |
| 51063 | calculatedFlags, checkBlockScopeBindings | `\|=` | No. The preceding test does not write. | stores `m` / stores `m` |


Run `NODE_PATH="$PWD/stage3/api/node_modules" node stage3/adapt/99-seats-build22-98-links-classes/test.cjs <applied-tree>` on a tree written by `stage3/apply.sh`. The test reads this adapter's own before and after snapshots from its `refs/stage3/` refs, so a later adapter can edit the working checker without affecting the test. The test copies both files into OS scratch, invokes the adapter, transpiles the original function and new class with stock TypeScript 6.0.3, and compares fresh bit tests and updates for both fields. The same test transpiles both SymbolLinks forms and evaluates the emitted constructor through the two allocation expressions in createSymbol and getSymbolLinks, comparing identity, names, prototypes, empty own properties and a write/read round trip. Function.prototype.toString differs between the forms and is outside the comparison. The order row checks that only imports precede the four declarations and all constructions follow them. The gate-sized test also checks unchanged bytes on refusal, line relocation, LF input and zero-write replay. The separate SDK/CLI witness reaches both allocations through the actual checker. The SDK emit remains a separate check because it compiles the full SDK/CLI graph. This pin has no diff-to-gates.json.

Unsanitized probes measured at ab9af917 follow. The two uninitialized witnesses reproduce the round 2 runtime failures.

| probe | Node | build | JavaScript | native |
|---|---|---|---|---|
| 171-after-compound-initialized.ts | 2 (exit 0) | head | agrees | agrees |
| 171-after-compound.ts | 2 (exit 0) | head | differs:  (exit 70), adamic: panic: read before assignment: field 'calculatedFlags' in links.calculat | differs:  (exit 70), adamic: panic: read before assignment: field 'calculatedFlags' in links.calculat |
| 171-bitwise-initialized.ts | 0 (exit 0) | head | agrees | agrees |
| 171-bitwise.ts | 0 (exit 0) | head | differs:  (exit 70), adamic: panic: read before assignment: field 'calculatedFlags' in links.calculat | differs:  (exit 70), adamic: panic: read before assignment: field 'calculatedFlags' in links.calculat |
| 171-class-value-initialized.ts | 0 (exit 0) | head | agrees | agrees |
| 171-class-value.ts | 0 (exit 0) | head | agrees | agrees |
| 171-compound.ts | 2 (exit 0) | head | stops: adamic: 171-compound.ts:5:20: stage 0 can't lower an assignment value to a member yet | stops: adamic: 171-compound.ts:5:20: stage 0 can't lower an assignment value to a member yet |

The sandbox census was skipped. Exposed stops await the unit census.
