# Slot 05 sixth batch

Three public helpers, each in its own .a file:

- newTheme(): return an independently writable empty store with fresh values and keyOrder containers, empty prefix and zero deadKeys. It reuses the already delivered MutableThemeState and immutable ThemeValue types.
- resolveThemeKey(theme, candidateValue, candidateValuePresent, namespaces, isIgnoredThemeKey): visit namespaces in input order. With a present candidate, check namespace plus hyphen plus candidate; with an absent candidate, check the namespace itself. An absent direct key may fall back to replacing every candidate dot with underscore, only for a present dotted candidate. Apply the external ignored-key callback after lookup and return the first accepted key. No match returns {key: '', found: false}. A found empty key remains distinguishable from no match.
- resolveThemeValue(theme, candidateValue, candidateValuePresent, namespaces, isIgnoredThemeKey): execute this batch's real resolver and return the stored literal string, preserving found=true when it is empty. Option bits and prefix do not affect literal resolution. No match returns {value: '', found: false}.

The ignored-key predicate is an explicit separately owned dependency. The oracle supplies its actual Go answers rather than copying it. ThemeState is the already delivered read-only map/order projection; constructor output is mutable and independently owned. nil Go receivers, concurrent mutation and callbacks that mutate the store are outside this usable-store contract. If an externally supplied callback removes a resolved entry, literal resolution explicitly panics instead of silently reading a missing value. Immutable map value records do not introduce ownership cycles or garbage collection.

An overlay exposes actual private Go resolveKey and state snapshots without changing cohere. Every nonempty string literal from every inventory-listed test file of all six consumers contributes; 639 distinct strings include prose/options as well as source. Missing coverage or pin drift fails. Go Add constructs lookup stores for each source value and derived dotted candidate, with direct-plus-fallback and fallback-only variants. Controls distinguish ignored font subnamespaces, direct precedence, all-dot replacement, namespace ordering/duplicates, empty namespace lists, present-empty versus absent candidates, and found-empty values.

Constructor comparisons create two real Go themes, snapshot both, write only the first and snapshot both again. Constructor mutants share either container between calls, proving the isolation observations can fail. Lookup comparisons hold the integrated resolver/value-reader path. Baselines compare actual Go, Node source, emitted JavaScript and sanitized native. Each semantic mutant must compile, exit 0 without stderr and disagree with Go; a compiler error, crash or sanitizer error does not count as a kill. Map snapshots are sorted by explicit UTF-8 byte comparison in the private driver; order arrays stay unsorted.

Run after sourcing /workspace/adamic-tools/env.sh:

```sh
ADAMIC_SLOT05_BATCH6_EVIDENCE=/workspace/adamic/stage1/cohere/lint/helpers/slot05/batch6/evidence ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers/slot05/batch6 -count=1 -v -timeout=15m > /tmp/lint05-batch6-final.log 2>&1
```

Go is pinned to 715ba94f3608a6500086b1076ce5cb7e51b836db. Coverage counts, generated-corpus SHA-256 hashes and final logs are committed under evidence/. No shared registration or rule harness is modified. This is helper comparison, not whole-rule findings/fixes/suggestions or production parser/theme/linter integration. Dynamic/external fixture reconstruction, raw invalid UTF-8 representations, arbitrary concurrent or callback mutation, nil-versus-empty Go allocation distinctions and the full repository gate are not covered. [CONSUMERS.md](CONSUMERS.md) and [REPORT.md](REPORT.md) record exact consumers and validation.
