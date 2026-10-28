# TestSplitMixesHitsAndMisses: the unit-key mutant

TestSplitMixesHitsAndMisses (units_test.go) takes over what TestSplitTSGoAgreesUnit00's cached build of the 4 MB tsgo
checker held: one cached build that mixes hits and misses links a binary that runs the changed body and is the same bytes
as the uncached build (#dgcy49s).

Mutant, applied to unitkey.go's `(*unitKeying).key`, so a unit is keyed without its own bytes:

```diff
-	return partsKey(append(slices.Clone(keying.parts), fmt.Sprintf("unit %s %x", unit.name, sha256.Sum256([]byte(unit.source)))))
+	return partsKey(append(slices.Clone(keying.parts), fmt.Sprintf("unit %s", unit.name)))
```

`go test ./internal/native -run '^TestSplitMixesHitsAndMisses$' -count=1 -v`, at 95a7b63f with the change, on macOS:

```
    units_test.go:375: the changed build's units: map[functions_0000.c:hit functions_0001.c:hit functions_0002.c:hit], want two hits and the changed unit's miss
--- FAIL: TestSplitMixesHitsAndMisses (1.34s)
```

Restored, it passes (5.62 s). The changed unit hit its old object. Without the outcome check, by reading (not run), the binary would print
780, not 1741, and differ from the uncached build.
