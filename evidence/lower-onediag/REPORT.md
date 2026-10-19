# Lower #onediag — stopped without a green submission

Code commit: b2998d92e877311d759e52a19eede2f33f256fb4
Base (newest fetched origin/main): 141fec7702d4aed21429724b2afd7efcd980f277
Source: 4090c3c8e32148dd485a562c60f309dad7a352fe. Original message and Co-Authored-By trailer preserved. One commit above main; no review/ files. Code branch compiler/lower-onediag is local and was NOT pushed because verification did not finish. No PR submitted.

Stopping reason: the 45-minute window was consumed by environment startup, absent pinned dependencies, and an unintended lazy Git history fetch during submodule initialization. Recovered with an explicit depth-one fetch of pinned TypeScript dependency d92d9bfee114c80be2c375d72edae966176e3a4f. Native and whole oracle tests were still running at the deadline and were explicitly terminated; their package failures are signal: terminated, not named test assertions. No ordinary failing test name was emitted. This is NOT a green result.

## Files, added/deleted lines

```text
3	1	internal/load/prelude.d.ts
47	0	internal/load/prelude_iterator_test.go
3	2	internal/lower/collections.go
2	2	internal/lower/object.go
22	0	internal/lower/prototype.go
2	1	internal/oracle/counts.md
11	0	internal/oracle/library_map_set_iterator_inference_test.go
59	0	internal/oracle/testdata/library_map_set_iterator_inference.a
```

## Linux counts

Linux x86_64; clang 20.1.8, LLVM 87f0227cb60147a26a1eeb4fb06e3b505e9c7261. go1.27.1 linux/amd64. Required update command passed in 624.940s. Main still uses internal/oracle/counts.md.

Numbers: allocations, frees, retains, releases, peak live, in regions.

- collections.a: (362,362,392,550,103,0) -> (361,361,390,547,103,0).
- library_map_set_iterator_inference.a: absent -> (107,107,112,159,24,0).
- No other fixture moved. A main update was therefore unnecessary.

## Checks

- gofmt -l cmd internal: PASS, empty output.
- go vet ./internal/load ./internal/lower ./internal/oracle: PASS.
- load: PASS, 5.740s.
- lower: PASS, 206.991s.
- javascript: PASS, 1.006s.
- native: INCOMPLETE / terminated, 724.624s; first error signal: terminated; no failing test name emitted.
- fresh: PASS, 72.293s.
- flow: PASS, 285.849s.
- ADAMIC_GATE_UNCACHED=1 whole oracle: INCOMPLETE / terminated, 531.410s; first error signal: terminated; no failing test name emitted. ReviewPrograms and recorded-count full-run verdicts unavailable. This run started before the counts update finished; a fresh whole run after the update is required.

## Mutants (both restored)

1. Anonymous done arm: TestPreludeIteratorsInferTheirElement failed. First error main.ts:10:44: error TS18048: 'r' is possibly 'undefined'.
2. iteratorResultField always false:
   - TestNativeAgreesWithNode_020/internal/oracle/testdata/library_map_set_iterator_inference.a failed at fixture:31:9.
   - TestNativeAgreesWithNode_087/internal/oracle/testdata/library_map_set_next.a failed at fixture:6:22.
   Both: Adamic 0.1 refuses inherited library member done read as an own field; prototype members are not stored in an object's shape; call the method on its receiver, or wrap that call in an arrow (unbound-method).

## tsc source proof

TypeScript v6.0.3 source HEAD 050880ce59e30b356b686bd3144efe24f875ebc8 verified. Locked stage3 API installed. stage3/apply.sh /tmp/onediag-adapted: PASS. Both go build -o ... ./cmd/adamic commands passed. Counts-only amendment does not change either binary's compiler sources.

Main first diagnostic:

```text
/tmp/onediag-adapted/src/compiler/builder.ts:1246:69: error TS2345: Argument of type 'Path | undefined' is not assignable to parameter of type 'string'.
  Type 'undefined' is not assignable to type 'string'.
```

Change first diagnostic exactly:

```text
/tmp/onediag-adapted/src/compiler/builder.ts:2273:9: error TS2375: Type '{ fileInfos: Map<Path, BuilderState.FileInfo>; compilerOptions: CompilerOptions; semanticDiagnosticsPerFile: Map<Path, readonly ReusableDiagnostic[]>; ... 7 more ...; checkPending: true | undefined; }' is not assignable to type 'ReusableBuilderProgramState' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.
  Types of property 'outSignature' are incompatible.
    Type 'EmitSignature | undefined' is not assignable to type 'EmitSignature'.
      Type 'undefined' is not assignable to type 'EmitSignature'.
```

The change has no builder.ts:1246:69 TS2345 as its first diagnostic. Both types commands exit 1 because later diagnostics remain. Full logs accompany this report. Binaries are omitted.
