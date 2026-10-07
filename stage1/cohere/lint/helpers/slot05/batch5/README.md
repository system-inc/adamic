# Slot 05 fifth batch

Three ordered theme mutation helpers, each with its own .a file:

- clearAll(theme): replace the map and order array, reset deadKeys to zero, preserve prefix. Old container aliases retain their contents.
- compactKeyOrder(theme): return unchanged when deadKeys is less than or equal to map size. Otherwise replace the order with a fresh array of nonblank, present keys in original order and reset deadKeys to zero. Duplicate present keys remain duplicates, and stored values/map identity are preserved.
- deleteThemeKey(theme, key): an absent key is a no-op. For a present key, delete its map entry, scan backward and blank only the final matching slot, increment deadKeys, then execute this batch's compactor. Blanking can affect a retained old array before compaction replaces it.

MutableThemeState carries prefix, values, keyOrder and deadKeys. The value shape is shared with the already delivered batch4 reader. Callers hold a usable typed store; no nil Go receiver, concurrent mutation or ownership cycles are introduced. Empty/nil Go lists share the empty-array representation. Values are immutable string/option records. The helpers do not implement Add, namespace clearing or store construction. Those remain separately owned dependencies.

The actual Go private methods are exposed by an oracle-only overlay without editing cohere. Every nonempty string literal in every inventory-listed test file of all six consumers contributes to the corpus (639 distinct literals). Actual Go Add operations create stores, overwrite entries, delete entries and reinsert them. Additional controls cover threshold equality/excess, empty stores, blank/stale/duplicate slots, empty stored keys, orphan map entries and Unicode keys. Direct controls expose states beyond normal Add output to check the exact private-method behavior.

The oracle records current prefix/dead count/order/map state, retained old order/map state, current state after writes through retained aliases, and a repeated operation. Sorted map snapshots use Go UTF-8 byte order, with an explicit byte comparator in the private Adamic driver; order arrays are never sorted. Actual current and alias states match Node source, emitted JavaScript and sanitized native. The production helpers are called directly. Delete calls the owned real compactor. Thirteen mutants must compile and run successfully with no stderr, then disagree with Go; compiler or sanitizer failures do not count as kills.

Run after sourcing /workspace/adamic-tools/env.sh:

```sh
ADAMIC_SLOT05_BATCH5_EVIDENCE=/workspace/adamic/stage1/cohere/lint/helpers/slot05/batch5/evidence ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers/slot05/batch5 -count=1 -v -timeout=15m > /tmp/lint05-batch5-final.log 2>&1
```

The runner refuses Go pin drift from 715ba94f3608a6500086b1076ce5cb7e51b836db, a consumer count other than six, or zero input contribution from any consumer. Coverage counts, corpus SHA-256 hashes and final logs are committed under evidence/. Only the private runner uses existing compiler APIs; no shared registration or rule harness changes are needed.

[CONSUMERS.md](CONSUMERS.md) lists every consumer; [REPORT.md](REPORT.md) records validation and limits. This establishes bounded helper equivalence, not full-rule findings/fixes/suggestions or production parser/theme/linter integration. Dynamically constructed and external fixture inputs, raw invalid UTF-8 keys/values, arbitrary concurrent mutations, Go slice capacity and nil/empty allocation distinctions, and the full repository gate are not covered. Exact map/list contents and observable alias writes are covered.
