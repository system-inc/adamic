# Upstream note

TypeScript 6.0.3 constructs type objects without their constraint/index caches,
then initializes those caches lazily. Declare the three caches as optional on
`Type`, and make the parallel union/intersection declarations optional too.
Use the exact existing `InstantiableType` contracts: present values are `Type`
and `IndexType`; construction can omit these fields.

Let `SymbolVisibilityResult` declare `errorModuleName?: string | undefined`,
already carried by accessibility results. Make the derived declaration explicitly
allow undefined too: one existing initializer stores it on an error path.
These declarations are internal.

Add `extendedSourceFiles?: string[]` and internal
`configFileSpecs?: ConfigFileSpecs` to `JsonSourceFile`. Ordinary JSON parsing
omits both, while configuration parsing enriches the same source objects with
string-array extends metadata and concrete configuration specifications. Copy
the exact existing `TsConfigSourceFile` contracts. The sole published addition is
`JsonSourceFile.extendedSourceFiles?: string[]`, under the unit's sanctioned
truthful optional-addition rule. The mechanical snapshot proof checks precisely
that property on top of the 189 optional, 28 adaptation-40 and one readonly lines.

Adamic needs these contracts because a structural source view that hides a target
optional field cannot safely support native field reads. This proposal records
properties at their owners, changes no initialization or runtime computation,
and emits byte-identical JavaScript. The default upstream oracle and per-owner
writer evidence accompany the proposal in this directory.
