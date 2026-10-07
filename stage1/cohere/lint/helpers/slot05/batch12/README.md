# Twelfth batch

caseTables selects exactly one supplied cached constructor: foldTables under Unicode mode and uppercaseTables otherwise. Both the member map and groups retain their dependency identities. Constructors remain explicit externally owned dependencies, including their lazy initialization and cache lifetimes. This helper does not build Unicode tables.

caseEquivalents calls the supplied caseTables once with the unchanged flag, then returns its shared member array without copying. An absent member returns undefined, the adapter projection of Go nil. Consumers must treat the array as immutable.

caseEquivalenceGroups calls the supplied caseTables once with the unchanged flag and returns its exact outer groups array, preserving every inner array and the existing order. Consumers must not modify either level. Go builds group order from map iteration; it is not deterministic across separate processes. The corpus preserves the actual runtime order rather than imposing a sort. Corpus hashes document each observed capture, not byte-for-byte regeneration across processes.

The Go overlay instruments only the real cached-constructor calls, preserving every algorithm and returned value. Test-only exports inspect map aliasing with a restored sentinel write, and inspect slice backing identity. Real Go creates 1,144 uppercase groups and 1,482 fold groups. The test driver reconstructs the supplied groups and member-map aliases, then compares exact values, selected constructors, call counts and returned identities on source Node, emitted JavaScript and sanitized native.

All four consuming Go test files contribute their nonempty string literals, including source, configuration and expected text. Their 904 distinct strings supply rune controls; the suite adds signed int32 extremes, absent members, surrogate integers, supplementary case pairs, Kelvin sign, long s and the three fold-only pairs seeded by Go. The member lookup sweeps every integer zero through 0x10ffff in both modes, including surrogates. Selector and group projections use repeated mode-switch sequences; groups compare the complete arrays each time. These are helper queries, not whole diagnostic replays.

Each source helper has three compiling semantic mutants. Compilation refusal, panic, sanitizer failure and stderr are not credited as a semantic mismatch. Shared registration, harness, findings and protected compiler files are untouched. No matcher or finding-offset conversion is added.

```
source /workspace/adamic-tools/env.sh
ADAMIC_SLOT05_BATCH12_EVIDENCE="$PWD/stage1/cohere/lint/helpers/slot05/batch12/evidence" go test ./stage1/cohere/lint/helpers/slot05/batch12 -count=1 -v -timeout=20m > /tmp/lint05-batch12-helpers.log 2>&1
```

Not covered: table-constructor implementations, cold concurrent construction, invalid UTF-8, noninteger or out-of-int32 rune adapter arguments, arbitrary manually supplied table shapes, full rule findings or integration. See CONSUMERS.md and readiness.json for all consumers and residual dependencies.
