# Mutants for overload mints (#z1qkw7d)

Each is an exact-string edit to overload_phantom_result.go, run on a house box, then put back. The pins are
overload_mint_test.go's.

| Mutant | Edit | Red |
|---|---|---|
| m1, no report line | `\tl.recordOverloadMint(overload, label, promised, mint)\n` removed | TestOverloadMintIsListedInTs, the sorted and getPathComponents rows: no overload mint line |
| m2, served by its one base | `\tbase, _ := l.phantomExtension(to)\n` made `\tbase, members := l.phantomExtension(to)\n\tif base == nil && to.Flags()&checker.TypeFlagsObject != 0 && to.ObjectFlags()&checker.ObjectFlagsReference != 0 && to.Target() != nil && to.Target().ObjectFlags()&checker.ObjectFlagsInterface != 0 {\n\t\tif bases := checker.Checker_getBaseTypes(l.checker, to.Target()); len(bases) == 1 {\n\t\t\tbase = bases[0]\n\t\t}\n\t}\n`, and `mint := l.brandMintOf(nil, l.overloadMintSource(from, base), to)` made `mint := &brandMint{base: base, members: members}` | TestOverloadMintKeepsRefusingARealMember: Named<number>'s overload is refused as a brand mint in .a, not in the census's words |
| m3, object brands unguarded | phantom_array_presence.go, phantomBrandMatching: `if (base != nil || supertype) && (match == nil || match(base, members)) {` made `if base != nil && l.isArrayType(base) && (match == nil || match(base, members)) {` | TestOverloadMintObjectBrandObservationsStop: the in, dot read and destructuring rows lower ||
