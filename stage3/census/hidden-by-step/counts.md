# Local reporting fixture counts

| Input/check | Count | Outcome |
|---|---:|---|
| Synthetic ranking fixtures | 1 | Known answer passes |
| Synthetic reasons | 3 | 22: 100 bytes; 15: 90 bytes |
| Focused tests | 6 | Pass |
| Duplicate-step mutants | 1 | Rejected before aggregation |
| Credited-byte mutants | 1 | Known answer mismatch |
| Input/coverage mutants | 7 | Rejected |
| Pinned compiler reasons | 1,712 | Every identity mapped once or explicitly unplaced |
| Other causes | 15 | Explicitly unplaced |
| Conserved hidden bytes | 3,654,880 | No dropped or duplicated credit |

This JSON fixture is not registered in the native oracle. No native counts changed.
