Built: collapse.ParseAtRule, one .a helper, with explicit separately owned AtRule allocation callback.
Commits: claim 2ef2c711 pushed before code; implementation commit is this report's commit.
Commands: owned go test PASS 4.446s, 1,143 cases, 143,354 Go bytes; owned go vet PASS with empty output; setup 77s, nproc 5.
Mutants: byte-offset-five changed to one, NEL removed from trim set, and original children copied; all compile and finish cleanly before every source Node/emitted JavaScript/sanitized-native comparison catches them.
Not covered: whole-rule findings, full shared-helper gate, arbitrary invalid UTF-8 byte strings, nil-versus-empty container representation; six dependency occurrences removed, zero additional fully helper-ready rules.

The oracle calls the unmodified ParseAtRule and AtRule at pinned cohere 715ba94f3608a6500086b1076ce5cb7e51b836db. An oracle-only virtual main is compiled through a Go overlay; no upstream source is edited. Successful output compares kind, name and params as sequences of UTF-16 code units, child order and a retained-child-array alias observation. Go output derives from actual result nodes, including child replacement after construction. Empty children are compared by length; no allocation identity is inferred for empty Go slices.

The API delegates AtRule allocation to an explicit generic callback, preserving the actual supplied children array. It independently implements byte-offset-five scanning and exact Go Unicode whitespace trimming. It counts UTF-8 byte length while walking JavaScript code points, so a space following @mé qualifies at byte five even though its UTF-16 position is four. NEL is trimmed; BOM is retained. Only space, tab and open parenthesis split the input, including Go's short-name behavior.

Capture reads every Go string literal from all matching files in six consumer test families: canonical 204, class order 316, variant order 52, shorthand 178, conflicting 195, unknown 184. Two empty/NUL/Unicode controls and twelve at-rule controls give 1,143 cases. Consumer literals include code, settings and expected output; this is a direct helper behavior comparison rather than a replay of the whole rule findings. The splitter accepts any string, so all extracted literals are real helper inputs. Controls distinguish short-name and byte-index behavior, parentheses, tab versus newline/CR, NEL, BOM and trimming. Three observations per input are compared without sorting.

Consumers:

- better-tailwindcss/enforce-canonical-classes
- better-tailwindcss/enforce-consistent-class-order
- better-tailwindcss/enforce-consistent-variant-order
- better-tailwindcss/enforce-shorthand-classes
- better-tailwindcss/no-conflicting-classes
- better-tailwindcss/no-unknown-classes

One dependency occurrence per consumer is removed. Other CSS parsing, loading, design-system and rule-local dependencies remain; no rule is marked fully unblocked or ported. No meaningful findings-per-second rate applies to a syntax helper that produces no lint findings.

Reproduce, with the setup environment sourced:

```
go test ./stage1/cohere/lint/helpers/from-wave1-13/at-rule -count=1 -v -timeout=10m > /tmp/at-rule.log 2>&1
go vet ./stage1/cohere/lint/helpers/from-wave1-13/at-rule > /tmp/at-rule-vet.log 2>&1
```

The preceding NewTheme comparison is withdrawn because slot 08 held an earlier claim. Its source snapshots and logs are inert evidence only and are not counted as a delivered helper. No compiler or shared harness is changed.
