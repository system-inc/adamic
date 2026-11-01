Attempted units: 4018. A total 794.908220s; B total 803.547514s; difference +8.639294s; ratio 1.0109.
CPU totals: A 842.766367s; B 849.785866s.

| Unit | A seconds | B seconds | B−A seconds | B/A |
|---|---:|---:|---:|---:|
| getImpliedNodeFormatForEmitWorker (src/compiler/program.ts:4855:1) | 4.440379 | 5.859663 | +1.419284 | 1.3196 |
| hasTabstop (src/compiler/utilities.ts:11077:1) | 5.201813 | 5.985840 | +0.784027 | 1.1507 |
| canHaveModuleSpecifier (src/compiler/utilities.ts:4330:1) | 4.613271 | 5.324491 | +0.711220 | 1.1542 |
| getCompilerOptionValue (src/compiler/utilities.ts:9402:1) | 5.120718 | 5.809007 | +0.688289 | 1.1344 |
| isDiagnosticWithDetachedLocation (src/compiler/utilities.ts:8637:1) | 4.955932 | 5.621273 | +0.665341 | 1.1343 |
| createResolvedModuleWithFailedLookupLocationsHandlingSymlink (src/compiler/moduleNameResolver.ts:224:1) | 4.680548 | 5.267548 | +0.586999 | 1.1254 |
| useCaseSensitiveFileNames (src/compiler/moduleNameResolver.ts:3424:1) | 4.544435 | 5.104956 | +0.560522 | 1.1233 |
| projectReferenceIsEqualTo (src/compiler/utilities.ts:803:1) | 4.761495 | 5.264143 | +0.502648 | 1.1056 |
| getEffectiveTypeAnnotationNode (src/compiler/utilities.ts:6892:1) | 0.313654 | 0.747479 | +0.433825 | 2.3831 |
| isConstAssertion (src/compiler/utilities.ts:11452:1) | 0.318406 | 0.723703 | +0.405297 | 2.2729 |
| compareDiagnosticsSkipRelatedInformation (src/compiler/utilities.ts:8790:1) | 5.353258 | 5.757005 | +0.403747 | 1.0754 |
| createFileDiagnostic (src/compiler/utilities.ts:8685:1) | 4.847146 | 5.239931 | +0.392786 | 1.0810 |
| getLocalModuleSpecifierBetweenFileNames (src/compiler/moduleSpecifiers.ts:433:1) | 4.387613 | 4.768174 | +0.380562 | 1.0867 |
| KindIfStatement (src/compiler/sys.ts:1980:1) | 4.832784 | 5.186697 | +0.353913 | 1.0732 |
| getEffectiveReturnTypeNode (src/compiler/utilities.ts:6911:1) | 0.281869 | 0.633362 | +0.351493 | 2.2470 |
| createTemplateCooked (src/compiler/transformers/taggedTemplate.ts:101:1) | 4.357139 | 4.704532 | +0.347393 | 1.0797 |
| getVersionPathsOfPackageJsonInfo (src/compiler/moduleNameResolver.ts:2407:1) | 4.249278 | 4.572618 | +0.323339 | 1.0761 |
| flattenDestructuringBinding (src/compiler/transformers/destructuring.ts:244:1) | 0.099654 | 0.391632 | +0.291978 | 3.9299 |
| isNonNullChain (src/compiler/utilitiesPublic.ts:1437:1) | 0.316311 | 0.597261 | +0.280950 | 1.8882 |
| getImpliedNodeFormatForFile (src/compiler/program.ts:1339:1) | 4.314718 | 4.586964 | +0.272246 | 1.0631 |

Noise controls: 1292 import/export/interface/type-alias units (no function body/predicate call lowered). Median absolute paired delta 0.006280s, p95 0.063790s; median B/A 0.9912, p05–p95 0.6849–1.4107. Single concurrent sample; scheduling, cache and GC effects limit causal timing attribution.
