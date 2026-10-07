# Slot 05 fourth batch

Three retained shared helpers, each in its own .a file:

- liveKeys(theme, visit) captures the original order array and length, skips empty and absent-map slots, reads the current values map for each key and stops at the first false callback result. Replacing the order during the callback does not change the captured range; appending does not extend the traversal. Mutation of a remaining slot or value does affect later visits, as in Go.
- themeEntries(theme, prefixKey) reuses liveKeys, retains insertion order, applies the separately owned PrefixKey dependency, and copies key/value/options into a fresh output list and records. Caller changes to that returned list cannot change subsequent reads or the theme store.
- keysInNamespaces(theme, namespaces, isIgnoredThemeKey) reuses liveKeys with namespace-major iteration, retaining duplicate namespace requests. It requires namespace plus a hyphen, excludes double hyphens occurring at or after UTF-8 byte offset two, applies the separately owned ignored-key predicate, and returns unprefixed suffixes. Prefix matching ends at a complete character, so slicing by the matched JavaScript prefix length yields the same suffix as Go's byte slice. A matched key shorter than two UTF-8 bytes explicitly panics, like Go's slice guard.

ThemeState exposes immutable key-order/value projections to the helper. Values carry the entire value string and option bitfield. Empty Go nil slices become empty arrays. Numeric option values are preserved. No AST ownership cycles or garbage collector are introduced. The theme must be a usable store; nil receivers/zero-value Go themes are outside the integration contract. Concurrent mutation is not supported by Go or this port.

The caller may perform the controlled mutations captured in the tests: delete/overwrite a remaining map entry, append, replace the order, replace the map, or alter a remaining order slot. Shortening the original JavaScript order array is refused explicitly; Go's copied slice header retains its old backing storage, which is a different representation. Arbitrary mutation of that array's length is outside the contract.

The independent Go oracle uses an overlay to access actual private Theme fields/liveKeys without writing the submodule. It scans every nonempty Go string literal in every inventory-listed test file for all six dependent rules. Stores are built by actual Go Add operations, including overwrite, deletion, re-insertion, defaults and blanks. Direct controls also include stale/duplicate order slots. Description and option strings are included; these are not full lint-fixture counts.

The live iterator checks every captured value with three stop limits and seven callback mutations. Entry checks use empty, ASCII and Unicode prefixes, mutate/clear the first returned list, then compare a fresh Go read. Namespace checks derive keys from every consumer string, include ignored-key boundaries, sub-variables, Unicode namespaces, duplicates and empty namespace lists, and test a separate short-key refusal.

PrefixKey and isIgnoredThemeKey remain explicit, separately owned dependencies. Their real Go outputs are supplied by the oracle to isolate these helpers. The body of either dependency is not copied into this delivery. Both readers execute this slot's real liveKeys, so comparisons also hold the integrated call paths.

Baselines compare Go to Node source, emitted JavaScript from Adamic's IR, and sanitized native. Mutants must compile, exit 0 without stderr and disagree with actual Go. Compiler/sanitizer failures do not count as semantic kills. The owned test runner uses existing load/lower/javascript/native APIs and leaves the shared rule harness untouched.

Run from the repository root after sourcing /workspace/adamic-tools/env.sh:

    ADAMIC_SLOT05_BATCH4_EVIDENCE=/workspace/adamic/stage1/cohere/lint/helpers/slot05/batch4/evidence ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers/slot05/batch4 -count=1 -v -timeout=15m > /tmp/lint05-batch4-final.log 2>&1

Tests refuse cohere pin drift from 715ba94f3608a6500086b1076ce5cb7e51b836db and missing consumer evidence. Corpus hashes and per-consumer counts are committed under evidence/.

The escape-terminator experiment was yielded to slot 03's earlier reservation. Its source and test are removed, its evidence is explicitly labeled withdrawn, and it is not counted as a retained helper.

[CONSUMERS.md](CONSUMERS.md) lists each helper's six rules. [readiness.json](readiness.json) retains residual blockers and cumulative slot 05 accounting. Limits include whole-rule diagnostics/fixes/suggestions, production theme/parser integration, dynamic/external fixture reconstruction, arbitrary malformed/mutable stores, raw invalid-UTF-8 representation checks and the full repository gate.
