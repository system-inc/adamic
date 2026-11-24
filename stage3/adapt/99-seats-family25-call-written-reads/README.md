# Family 25 call-written reads

The Nov 22 ruling on #bbssdcy accepts eight source edits in five shapes.
The adapter checks all anchors before writing and accepts a second application
without rewriting files. Source paths and locations below refer to the fully adapted before tree, with this adapter absent.

| Source | Read | Edit | JavaScript emit |
| --- | --- | --- | --- |
| `checker.ts:7340:41` | Tracked symbols after a callback | Read the same receiver through its own type in the cache initializer. | Byte-identical |
| `checker.ts:12556:18` | Variable type cache | Use the same-type receiver for the second cache test. | Byte-identical |
| `checker.ts:13603:18` | Enum member type cache | Bind the plain local `links` after the computation. Retest, conditionally set and return through the alias. | Changes |
| `checker.ts:30285:44` | Assignment position | After the successful call, alias the parameter `symbol` and retain the two reads and short circuit. | Changes |
| `checker.ts:39699:26` | Return type cache | Use the same-type receiver for the second cache test. | Byte-identical |
| `checker.ts:50123:26` | Resolved symbol cache | Use the same-type receiver for the second cache test. | Byte-identical |
| `utilitiesPublic.ts:1341:47` | Illegal type parameters | In the predicate's true branch, bind the parameter `node` and retain both reads. | Changes |
| `transformers/declarations.ts:877:17` | Side-effect import argument | Read through the same-type receiver in the original argument position. | Byte-identical |

The five receiver assertions erase in stock JavaScript emit. The other three
edits introduce local bindings and preserve the order and number of property
reads. The aliases take identifiers, so their initializations invoke no getter.

The `99-seats-` prefix puts this adapter after `99-method-value-views` so its
whole-owner pin reads the expected input before the tracked-symbol edit.

## Exposed stops in scoped lowering

The illegal-type-parameters alias clears the earlier family-25 scan finding
at `utilitiesPublic.ts:1341:47` in `getEffectiveTypeParameterDeclarations`.
The same unit already recorded these two NotYet findings in the JSDoc
`flatMap` instantiation. The Nov 21 meter lists their reason.

| Location | Family | Unit | Census reason |
| --- | --- | --- | --- |
| `core.ts:406:47` | Narrowed union with an uncheckable object tag | `getEffectiveTypeParameterDeclarations` | a narrowed union member whose object tag cannot be checked with typeof; keep differently held object kinds in separately typed variables |
| `core.ts:409:48` | Narrowed union with an uncheckable object tag | `getEffectiveTypeParameterDeclarations` | a narrowed union member whose object tag cannot be checked with typeof; keep differently held object kinds in separately typed variables |

These exposures refer to the scoped inventory. The full first-error meter
already recorded the first lowering stop before the adaptation and reported
no new stops in either pass.
