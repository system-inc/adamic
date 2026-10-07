# Checked-view blocking families at 609ed395

Measured before the next family implementation on the pinned 2,936 sites. These
are overlapping transitive target-contract dependencies, not first blockers,
observed field reads, execution frequencies or successful lowering counts. A site
is counted once per family. Traversal stops at array/tuple and callable contracts,
which belong to lane 2, rather than walking their intrinsic library APIs.

| Blocking family | Tagged | Untagged | Sites blocked |
| --- | ---: | ---: | ---: |
| array or tuple contracts | 1758 | 1151 | 2909 |
| nullish members | 1758 | 1145 | 2903 |
| callable contracts | 1758 | 1132 | 2890 |
| optional properties | 1758 | 1131 | 2889 |
| dictionary contracts | 1758 | 1117 | 2875 |
| empty or unsupported field contract | 1758 | 1112 | 2870 |
| mixed primitive union | 1758 | 1110 | 2868 |
| object plus primitive union | 1758 | 1110 | 2868 |
| any field contract | 1758 | 1109 | 2867 |
| intersection field contract | 1758 | 1107 | 2865 |
| never field contract | 1758 | 1104 | 2862 |
| untagged object union | 1758 | 1104 | 2862 |
| union cast admission | 261 | 25 | 286 |
| nominal class fields | 0 | 18 | 18 |
| generic field contract | 0 | 6 | 6 |

Arrays/tuples are biggest (2,909), so their integration starts first. Lane 2 tip
2a16ce35 supplies standalone lowering and JavaScript helpers but no native element
checker. The dependency merge is f709a1c5. Lane 1 supplies the missing shared
registry adapter and physical storage byte first; admission remains NotYet until
the runtime checks and all read/write paths are wired. This does not unlock a
complete tsc site. Nullish members are next and are the largest lane 1 family.

The new native `adamic_array.element_kind` records storage, not logical element
conformance: 0 unknown, 1 number, 2 boolean, 7 packed number/undefined, 10 heap
pointers. Ordinary fresh array literals without spreads populate it; all other
producers currently remain unknown. Do not infer or overwrite it from a cast.
Mutations/copies/iteration/callbacks require explicit integration; unknown is no
proof. References must still be classified from each selected heap header, then
checked against the full logical element contract. No initialization bitmap was
added. Native struct size remains 56 bytes on this clang/Linux target; the byte
occupies padding at offset 33. Literal construction adds one byte store.

`internArrayViewContract` adapts lane 2 `viewArrayContract(node,target,buildElement)`
into `viewArrayContractHook`, reserving an id before recursion. The callback error
must propagate and child ids come from the common registry. Arrays of recursive
interfaces terminate; unknown element contracts remain unsupported. Shared view
admission is deliberately still disabled for arrays.

The census matches all 2,936 source spans and texts and stock TypeScript reports
zero diagnostics. TypeScript 6.0.3 source commit 050880ce59e30b356b686bd3144efe24f875ebc8.
Per-site witnesses are blocking-families-sites.json; the executable census is
blocking-families.cjs, both under stage3/interface-downcasts. The existing rejection
audit still establishes zero complete eligible contracts. No whole-file lowering
pass or unlocked site is inferred from these dependency counts.
