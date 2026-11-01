# Wave 6 composition and the 99f import owner

52 uses the existing named import in src/harness/client.ts. It adds CodeAction, InstallPackageAction, isAdamicJsonObject, ReferencedSymbolDefinitionInfo, ReferencedSymbolEntry and SymbolDisplayPart before ApplicableRefactorInfo. It reuses the existing named imports for the other helper types. There is no namespace import and no second declaration from ./_namespaces/ts.js. The DocumentSpan insertion point is unchanged, so 99f inserts its type HashedSet specifier there.

The helper predicates refer to these names directly. Their Node witnesses and the real-reply regression harness bind the imported isAdamicJsonObject function directly as well. No check, returned value, exemption or public declaration changes in this repair. The proof now requires exactly one named import from that module.

## Exact composition

- Start at typescript/axolotl-wave-6, 63b44d41a8fe7e9ee707959b7153190e164148b0.
- Add compiler/axolotl-98d, d26001bf42d4e6ec74ad41a3c1b96d8832431f88, and this 52 directory plus its existing Go test file.
- Merge origin/main, 80c2d3282e44b527555d59690bb413cc883d2bf2, in scratch, including its extended API generator.
- Resolve only three generated lane manifest conflicts by retaining wave 6's shard-plan.json, shards.json and shards_generated_test.go. No such resolution is included on the implementation branch.
- Clone the pinned TypeScript commit 050880ce59e30b356b686bd3144efe24f875ebc8 and run stage3/apply.sh --adapt-checkout on it. Run all 59 adaptations in order, including 52, 98-d and 99f. All pass. No other adapter refuses.
- Build the control identically, omitting only 52 and its Go test. Its 58 adaptations also pass.

The evidence contains the exact composition patch, source pins, completed-adapter lists, full apply logs and build commands. The native tool is built from the composed checkout with go build ./cmd/adamic, using the matching cohere pin 06af2e399c6c1fb93fd570b774027942747400b4. Both adapted trees use upstream's normal npm run build, including its typecheck.

## Results

| Proof | Control | With 52 |
| --- | ---: | ---: |
| Stock diagnostics, each of six requested projects | 0 | 0 |
| Native whole-program diagnostics over 79 compiler roots | 8 | 8 |
| New native diagnostics | | 0 |
| Removed native diagnostics | | 0 |
| Public explicit any tokens | 37 | 22 |
| Full oracle passing | 106,366 | 106,366 |
| Full oracle failing | 1 | 1 |
| Full oracle pending | 0 | 0 |

Both full oracles move only api/typescript.d.ts. APISample_parseConfig.errors.txt is absent in the normal candidate, and neither completion-details server baseline moves. The oracle declaration, built declaration and hand-applied 52 sanctions match byte for byte.

52's own 16 sanction rows, applied to the built composition control, match the candidate d.ts byte for byte without formatting. The incremental diff is still 135 lines (+19/-15). The public-any list remains the single parseConfigFileTextToJson.config entry; the other 21 any tokens retain their reported excluded classes. All thirteen runtime check mutants fail, and both parallel Go exemption tests pass.

The merged generator consumes all composition sanction rows. Its raw expected text differs from the bundle in the compact member spellings from 97-cli-axolotl-d. That adaptation's README already records that upstream expands those type literals across lines. The exact pinned scripts/dtsBundler.mjs formatter still preserves the compact layout when fed the generator text, so it does not give a complete byte match. The lane's actual normalize-api.cjs comparison passes: all normalized declaration changes, including every type-bearing AST child, agree. No 97d row, generator, manifest or baseline is changed to conceal the raw discrepancy. It is retained in generator-mismatch.diff. Only the direct 52-row reconstruction and the built/oracle artifact comparison are claimed byte-identical (SHA-256 9868791a1b24cb29cacc482db323ca18ab94a7d32ec7c42b7ba9b958d09e474b).

The real completion-details counterexample also reruns successfully, returning the same object and string tag text by identity. A fresh composed-source mutant narrowing the pinned parseConfigFileTextToJson.config return builds successfully and makes the targeted oracle red: 5 passing, 1 failing, exactly APISample_parseConfig.errors.txt moves.

The duplicate-import mutant copies the five files owned by the actual 99f adapter, adds the old namespace import to the composed client, and reruns 99f. It fails with exactly 2 !== 1 and writes no file. In the real candidate, the AST has one named ts-module import, including HashedSet immediately after DocumentSpan. This exercises the rejected and accepted import compositions directly.
