# Local validation

Pinned cohere: `7945d102a6c18dd36adf9114a758ce646e8b2359`. Pinned compiler corpus: `050880ce59e30b356b686bd3144efe24f875ebc8`. Go 1.27.1, Node 24, clang 20 with ASan/UBSan.

- Both recording overlays ran the complete upstream React Go test package. The coverage guard checks all 34 delivered symbols, source hashes, and 234 controls.
- `TestGoAgreement`: 2,054 captured leaf calls and 234 controls per backend. `TestAstGoAgreement`: 27,733 captured calls per backend. All outputs match Go on source Node, emitted JavaScript, and sanitized native: 30,021 observations per backend, 90,063 total.
- `TestMutants`: nine leaf mutants; `TestAstMutants`: 25 AST mutants. Each compiles and changes output on all three backends. The AST mutants were run in two groups; the sort mutant was corrected to exercise the captured VariableDeclaration path, then checked again.
- The existing root helpers package passed its full suite, including its mutants and explicit refusal/gap checks. `go vet` passed for this helper package.
- Lint `TestRulesAgree` passed every owned/upstream/generated input set, comparing 13,810,062 output bytes. Inherited explicitly unsupported parser-recovery cases retain the existing refusal checks.
- `TestOwnedWitnesses` and `TestMutants/try_block_omitted` passed. The new rule covers 69 unique upstream source/file/options cases and its witness.

The compiler/stage1 corpus contains 910 files. The monolithic native command reached the shared harness's ten-minute command limit; Go, source Node, and emitted JavaScript had already produced identical 30,272,731-byte outputs. Native's completed prefix was preserved and compared before resuming the remaining manifest entries with the same sanitized binary in bounded batches. Case numbering was offset back to the original manifest before the final byte comparison. No input was omitted.

The separately owned ES5 dependency is documented in [STOPPED.md](STOPPED.md). Capture counts and pinned source hashes are in [testdata/coverage.json](testdata/coverage.json); symbol/file mappings and partial-port provenance are in [README.md](README.md).
