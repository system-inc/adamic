# Family 5 G02 site-local candidate

Adamic base is origin/main `491e020de2aa138d7b608611c40a641571a5a1eb`. The upstream TypeScript tag is `v6.0.3` at `050880ce59e30b356b686bd3144efe24f875ebc8`. The recursive submodule pins are cohere `5360753e19e4417675a90e92a1823bea66b13918` and cohere/TypeScript `d92d9bfee114c80be2c375d72edae966176e3a4f`.

The site IDs and coordinates below come from the scout at `464502ff430b1c88b2e79ebf6b21badcb09e02af`. All 14 are accepted stock structural relations. The Adamic rule at `internal/lower/optional_widening.go` checks the argument against its resolved parameter, or the right-hand value against its assignment slot. In the compiler-only program `SourceFile` has no declared `getPositionOfLineAndCharacter`. The target `SourceFileLike` declares it optional, so the rule refuses the relation. A structural value may physically carry the method despite the narrower source type.

## Site audit

- **Text** is `Pick<SourceFileLike, "text">`. The token reader inspects text and may use the file as a cache key or recurse.
- **Line map** is `Pick<SourceFileLike, "text" | "lineMap">`. `getLineStarts` reads or fills the line map from text. Both internal callers, `getEndLinePosition` and `getLinesBetweenPositions`, need this same view.
- **Identity** is `object`. The node-children cache uses these parameters solely as WeakMap keys. Its key type and all four helpers use the same identity view after the edit. The source remains the same object.
- The original parameter in each argument row is `SourceFileLike`, optional for `getTokenPosOfNode`. S122 has the corresponding optional assignment slot.

| Site | Expression | Callee or slot | Proposed annotation | Type-only | Can a services SourceFileObject reach this reader? Does it dispatch? |
| --- | --- | --- | --- | --- | --- |
| S022 checker.ts:49348:52 | `sourceFile` | `getTokenPosOfNode` sourceFile | Text | Yes | Yes, a services program can check unreachable code. No method dispatch. |
| S043 parser.ts:10015:36 | `sourceFile` | `transferSourceFileChildren` sourceFile | Identity | Yes | Yes, a prior file from a language-service incremental update is the cache key. No method dispatch. |
| S044 parser.ts:10015:48 | `result` | `transferSourceFileChildren` targetSourceFile | Identity | Yes | Yes, the incremental reparse uses the installed services allocator for the new file. No method dispatch. |
| S045 parser.ts:10087:37 | `origSourceFile` | `unsetNodeChildren` origSourceFile | Identity | Yes | Yes, a services incremental update can clear entries for old nodes. No method dispatch. |
| S046 parser.ts:10244:42 | `sourceFile` | `unsetNodeChildren` origSourceFile | Identity | Yes | Yes, the services incremental update can adjust positions and clear the old file's entries. No method dispatch. |
| S058 program.ts:738:57 | `file` | `getPositionOfLineAndCharacter` sourceFile | Leave | — | Yes, a diagnostic can refer to a services file. The reader calls the method when present, with `i, 0, undefined`, consistent with the implementation signature. |
| S059 program.ts:739:76 | `file` | `getPositionOfLineAndCharacter` sourceFile | Leave | — | Yes, the same diagnostic path dispatches on a non-final line with `i + 1, 0, undefined`. |
| S123 utilities.ts:2532:70 | `sourceFile` | `getEndLinePosition` sourceFile | Line map | Yes | Yes, a diagnostic can compute its span on a services arrow-function node. No method dispatch. |
| S124 utilities.ts:7945:37 | `sourceFile` | `getLinesBetweenPositions` sourceFile | Line map | Yes | Yes, the printer can format a services-created source file and reaches this range helper. No method dispatch. |
| S125 utilities.ts:7950:37 | `sourceFile` | `getLinesBetweenPositions` sourceFile | Line map | Yes | No in-tree caller found for the owning helper `getLinesBetweenRangeEndPositions`. If called directly with a services file it would only use the line map and text. |
| S126 utilities.ts:7960:37 | `sourceFile` | `getLinesBetweenPositions` sourceFile | Line map | Yes | Yes, services outlining, completions and text changes call `positionsAreOnSameLine`. No method dispatch. |
| S127 utilities.ts:7972:37 | `sourceFile` | `getLinesBetweenPositions` sourceFile | Line map | Yes | Yes, the printer can use the preceding-whitespace helper on a services file. No method dispatch. |
| S128 utilities.ts:7978:37 | `sourceFile` | `getLinesBetweenPositions` sourceFile | Line map | Yes | Yes, the printer can use the next-whitespace helper on a services file. No method dispatch. |
| S122 utilities.ts:1261:24 | `getSourceFileOfNode(node)` | `sourceFile ??=` slot in `getTokenPosOfNode` | Text | Yes | Yes, a services node's parent chain can end at a services source file and the caller may omit the argument. No method dispatch. |

The services routes follow from `setObjectAllocator(getServicesObjectAllocator())`, the language service's incremental update and `SourceFileObject.update`, and the public diagnostic and printer entry points. The services implementation accepts `(line: number, character: number, allowEdits?: true)`. The public services augmentation exposes its two-argument form.

For S058 and S059 the scanner consumes the very field named by the refusal. Replacing its input with a view that exposes only the line map and text would erase a property the body actually uses. An overload that accepts bare `SourceFile` while the implementation still reads the optional member through `SourceFileLike` would leave that conversion unproved. A runtime check or an object projection would change the emitted JavaScript. Both sites are therefore left in the candidate.

Stock acceptance and build results apply to the complete group of proposed edits and will be recorded with their exits. Neither unchanged call is an attempted annotation.
