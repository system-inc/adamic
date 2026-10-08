# Finished wave1-07 landing

Branch: codex/lint-wave1-07-land
Base: origin/area/stage1-lint e667e3e1

This branch carries only the four finished ledger-winning rules:
- no-unsafe-negation
- no-unsafe-optional-chaining
- no-underscore-dangle
- @typescript-eslint/no-this-alias

consistent-return and constructor-super remain reserved and unfinished on codex/lint-wave1-07 only. Their claims and historical reports are not imported here. No directories or registry descriptors for those unfinished rules are added to this landing branch.

Both previously surviving mutants retain their original anchors and have dedicated option-bearing witnesses: ordering.ts.txt with enforceForOrderingRelations=true, and constructor-property.ts.txt with allowAfterThisConstructor=true. All four finished descriptors use node:true and take the handed ParseNode. Fresh unified-harness evidence follows in wave1-07-land-evidence.
