# Readonly module specifier ending preferences

Seats, build22-140, #bbssdcy. At Adamic pin
`ef741d4f0860597ff8d4df5dd042973b2f20819d`, family 140 is the singleton
`[ending]` at `moduleSpecifiers.ts:1010:103` in
`tryGetModuleNameFromPaths`. Family 167 is `indexOf` at
`moduleSpecifiers.ts:1374:24` in `processEnding`. Adamic's Nov 22 ruling
builds the readonly array for both sites, with Compiler accepting a type-only
change and Adamic accepting one that moves JavaScript.

The two factory results, the root-dirs and paths forwarders, and
`processEnding` use `readonly ModuleSpecifierEnding[]`. The singleton calls
now have array context. Both `indexOf` calls and the `findIndex` call use
the original list in its original order. This is a readonly array under
`docs/rulings.md:343`. The first-index assertion stays at the original
switch point, consistent with the read-time rule at row 25. Under stock
TypeScript 6.0.3, repeating the indexed read in the exhaustive default does
not carry the switch's narrowing to `Debug.assertNever`. The switch and its
default use a local bound at the checked read, which changes emitted
JavaScript and places this head with Adamic. The changed types are private.
The factory and its internal interface keep their declaration types under
row 406.

## Non-empty producers

The stock checker's references in the pinned adapted `moduleSpecifiers.ts`
identify these routes to `processEnding`.

| Call site | List origin |
|---|---|
| `moduleSpecifiers.ts:592` and `:619` | `getLocalModuleSpecifier` reads the factory result at `:590`. Its three callers at `:314`, `:443`, and `:511` obtain their own factory instances via `:264` or `:284`, `:449`, and `:463`, respectively. |
| `moduleSpecifiers.ts:970` | The paths mapper makes a fresh singleton for each ending. Its inputs arrive from the same factory at `:614` or `:1288`. |
| `moduleSpecifiers.ts:1010` | `validateEnding` makes a fresh singleton after narrowing to `Minimal`. |
| `moduleSpecifiers.ts:1183` | The sole root-dirs caller at `:591` passes the factory result from `:590`. |
| `moduleSpecifiers.ts:1225` | `tryGetModuleNameAsNodeModule` constructs its own preferences at `:1197` and takes the factory result at `:1198`. Its callers at `:298`, `:313`, and `:500` pass user preferences, so no override of the list factory reaches this call. |

The factory method at `moduleSpecifiers.ts:199` has eleven normal result
literals of length one to four at `:205`, `:207`, `:211` to `:212`,
`:218` to `:219`, `:221`, `:224` to `:225`, and `:228` to `:229`. Its default throws
at `:231`. The forwarding paths make no writes to the list. The adapter
reuses adaptation 32's twelve runtime body pins and its interface check
to guard those conclusions. It locates all seven edits in the syntax tree,
checks their exact before or after text, and keeps the original line
endings and source line numbers.

## Exposed stops

The census by unit at `ef741d4f` (entry J-dfh6804, directory J-9qaegdk,
#macxggx) clears both family stops, and each unit's first stop moves to a
reason the Nov 24 meter doesn't list. Neither unit lowered before. Adamic's
2026-11-26 21:26:44Z ruling on #bbssdcy counts both as exposures, the second
on the condition that a miniature at main, a numeric enum const switched over
every member with `default: assertNever(x)`, stops at that argument with the
same words.

| Location | Family | Unit | Census reason |
| --- | --- | --- | --- |
| `moduleSpecifiers.ts:979:22` | 140, from `:1010:103`; filed as #rm22wzv | `tryGetModuleNameFromPaths` | a for...of destructuring an object |
| `moduleSpecifiers.ts:1411:38` | 167, from `:1374:24`; its task is filed after the miniature | `processEnding` | a call whose argument for a never parameter isn't a rulings.md:76 never read |

`:979:22` is upstream's own text; the adapter edits only that function's
parameter type at `:926`. `:1411:38` is the argument `firstEnding` of
`Debug.assertNever`, which the adapter binds in place of upstream's
`allowedEndings[0]`.

## Snapshot output test

Run the test on a tree written by `stage3/apply.sh`:

```sh
timeout --signal=KILL 3600 env NODE_PATH="$PWD/stage3/api/node_modules" node stage3/adapt/99-seats-build22-140-ending-list-array/test.cjs <applied-tree>
```

The test reads the adapter's own before and after refs, checks exact replay
and the four refusal cases, and compares the two versions of `processEnding`
over all eleven factory lists and all four enum singletons. It checks the
exhaustive default with an invalid first ending and supplies both outcomes
of the host lookup for the minimal ending.
