# Count row handoff

Measured with `adamic build --count` after restoration of all mutants. Lane 1
owns central oracle fixture registration and counts.md regeneration. The new
source control is exercised by TestCheckedViewCallableContractControl; this
row has not been added to the central table by lane 5.

| Fixture | Allocations | Frees | Retains | Releases | Peak live | In regions |
|---|---:|---:|---:|---:|---:|---:|
| stage3/interface-downcasts/lane5/probes/number-good.a | 4 | 4 | 4 | 8 | 4 | 0 |

Wrong-value, wrong-arity and optional-absent witnesses remain integration probes,
not passing central oracle fixtures. Their current cast refusals cannot be
counted as successful read-time checks.
