# Recordreplay amendment 5 evidence

Task #b7phwab. Main `34c05e3a4f79fa64ef91e9200df03af4b0c02208`. Final amendment `9d1dc4bd430a96f64b4a61e038c9edc444668f24`.

The detached scratch worktree was used for all code changes. CLI's three cherry-picks applied cleanly. The amendment resolves all five conflicting files, preserves main's landed edits, and drops the adapter-written `src/recordreplay.d.ts` and its erased import. The original commit message is retained with one paragraph naming main, the dropped declaration, and every tree path differing from ff0e5859. Nothing was submitted.

| Commit | SHA |
|---|---|
| CLI 1 | `1af44419e3d79710c9e88d48cdd19eefe81b6677` |
| CLI 2 | `3bae37ff56e8b26c5342908e1c9ab860f25d11f8` |
| CLI 3 | `bf860ed57b4d633fe03ebab79eb261c27edc33e8` |
| Amendment | `9d1dc4bd430a96f64b4a61e038c9edc444668f24` |

## Proof

All proof code was run at the final amendment commit. The census/compiler binary is identical to the CLI-only re-cut because no compiler source differs. While fresh apply ran, checks used the final adapter's composed output; **all 747 source files of that input match fresh apply exactly** (`fresh-identity.json`). The only source delta from CLI-only control is the bare recordreplay read. No declaration or import is added to the adapted tree.

Stock TypeScript is unmodified 6.0.3. Its API check loads CLI's exact Node-prelude augmentation virtually. The stock tsc CLI uses the same exact augmentation in a scratch-only ambient prelude outside the adapted tree, alongside all 79 implementation roots. Production `adamic types` uses its actual Node prelude and exits 0 over those 79 roots. The recordreplay contract independently requires TS2339 when only the augmentation is removed.

Both full single-file sys.ts censuses completed with CommonJSBundle, LATENT_FULL=1 and LATENT_ASSERT_NO_OUTPUT=1, and identical 135-unit coverage. Findings: 155 → 194; 5 removed, 44 newly exposed. Every new finding is in a unit that already stopped on control; **0 units newly stop**. All 44 classifications and prior findings are in `census-comparison.json`. Both census streams have completion trailers; no trailer findings. The recordreplay cast is absent.

Owner pins were generated with the directory's scoped-hashes.cjs and adapt.cjs --pin, never edited by hand. The --pin run wrote no source; the composed output then matched the planned cast-only delta exactly. 93 was checked on its authoritative own before/after snapshots.

## Every exit code

| Check / mutant | Exit |
|---|---|
| control fresh apply | 0 |
| final fresh apply | 0 |
| scoped pin generator | 0 |
| adapter --pin | 0 |
| exact composed adapter | 0 |
| 93 repaired check | 0 |
| recordreplay check | 0 |
| control full census | 0 |
| amendment full census | 0 |
| census comparison | 0 |
| control stock API check | 0 |
| final stock API check | 0 |
| stock tsc CLI | 0 |
| adamic types | 0 |
| fresh source identity | 0 |
| scope | 0 |
| bindings | 0 |
| recordreplay-contract | 0 |
| recordreplay-node-resolution-mutant | 1 |
| require-result-contract | 0 |
| resolve-host-contract | 0 |
| watch-directory-contract | 0 |
| watch-directory-witness | 0 |
| scope-mutant | 1 |
| ternary-mutant | 1 |
| require-result-restore-mutant | 0 |
| require-result-restore-refused | 1 |
| resolve-host-restore-mutant | 0 |
| resolve-host-restore-refused | 1 |
| watch-directory-restore-mutant | 0 |
| watch-directory-restore-refused | 1 |
| moved-owner | 1 |
| moved-imports | 1 |
| runtime-scope-mutant | 1 |
| watch-directory-unnormalized | 0 |
| watch-directory-unnormalized-refused | 1 |
| watch-directory-optional-contract | 0 |
| watch-directory-optional-contract-refused | 1 |
| span-disabled-control | 0 |
| span-disabled-scope-refused | 1 |
| recordreplay declaration mutant | 1 |
| recordreplay out_of_scope mutant | 1 |
| recordreplay owner mutant | 1 |

Exit 1 on negative mutants is expected. Restoring/rewriting mutant constructors exit 0, then the guarded adapter rejects their output with exit 1. The disabled-span control exits 0 and writes its extra edit; the complete source checker rejects that result with exit 1. Bindings and the regular contracts also contain independent mutation assertions.

## Receipts and reproduction

`commands.json`, `contracts.json`, `extra-mutants.json`, `recordreplay-results.json`, and the compressed raw logs retain commands and exits. `provenance.json` and `env.sh` identify dependencies and scratch caches. The updated check, contract, and pin tooling live in the amendment commit. `prepare-pins.py`, `run-contracts.py`, `extra-mutants.py`, `stock.cjs`, and `compare-census.py` are the actual scratch drivers. The source snapshots are reproducible with fresh stage3/apply.sh at the CLI-only and final commits; the relevant ref names are in the drivers.

Historical recordreplay inventory JSON in the amendment source is retained from ff0e5859 and labeled historical in the adapter README. Fresh observations are the files on this evidence branch. The historical foreign-site mutant is already landed on this main; its old needle is absent. Current owner/import/runtime-scope, span, binding, ternary, Node-resolution, recordreplay and watch-directory mutations are covered here.

The initial scratch comparison tried to inspect the still-running fresh apply and hit its cast-presence assertion. Its path was corrected to the completed census input, and the successful comparison was then corroborated against all fresh source bytes. No compiler or amendment code changed after the final commit.
