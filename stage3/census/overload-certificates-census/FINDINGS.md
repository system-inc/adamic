# Complete finding grades

Fingerprints are `(kind, phase, reason, full message)` with complete source coordinates replaced by `<position>`. Same-source unit identity is correspondence, not part of the diagnostic tuple. Reached means an additional non-scan stop in a previously blocked unit; it is not a regression verdict or successful compilation. Raw positions and full original messages remain in paired.json.gz and the streams.

## entry — src/tsc/tsc.ts

Before 4488; after 4485; unchanged 4482; grades {'gone': 6, 'reached': 3}.

### 1. gone — `src/compiler/builder.ts:1258:1`

base: Refused / lowering / `src/compiler/core.ts:1334:1`

```text
overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

### 2. reached — `src/compiler/builder.ts:1258:1`

head: NotYet / lowering / `src/compiler/builder.ts:1421:25`

```text
a certified overload whose implementation stores (IncrementalBuildInfoFileId | Path)[] where the call reads IncrementalBuildInfoFileId[]
/workspace/scratch/overload-certificates-census/adapted/src/compiler/builder.ts:1421:25: stage 0 can't lower a certified overload whose implementation stores (IncrementalBuildInfoFileId | Path)[] where the call reads IncrementalBuildInfoFileId[] yet
```

### 3. gone — `src/compiler/builder.ts:2277:1`

base: Refused / lowering / `src/compiler/core.ts:1400:1`

```text
overload 2 of arrayToMap result Map<K, V2> cannot be served by implementation result Map<K, V1 | V2>
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1400:1: Adamic 0.1 refuses overload 2 of arrayToMap result Map<K, V2> cannot be served by implementation result Map<K, V1 | V2>; make the implementation result covariant with every overload result
```

### 4. reached — `src/compiler/builder.ts:2277:1`

head: NotYet / lowering / `src/compiler/builder.ts:2416:31`

```text
a certified overload whose implementation stores Map<Path, V2 | IncrementalBuildInfoDiagnosticOfFile> where the call reads Map<Path, readonly ReusableDiagnostic[]>
/workspace/scratch/overload-certificates-census/adapted/src/compiler/builder.ts:2416:31: stage 0 can't lower a certified overload whose implementation stores Map<Path, V2 | IncrementalBuildInfoDiagnosticOfFile> where the call reads Map<Path, readonly ReusableDiagnostic[]> yet
```

### 5. gone — `src/compiler/commandLineParser.ts:1873:1`

base: Refused / lowering / `src/compiler/core.ts:1334:1`

```text
overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

### 6. gone — `src/compiler/commandLineParser.ts:2307:1`

base: Refused / lowering / `src/compiler/core.ts:1400:1`

```text
overload 2 of arrayToMap result Map<K, V2> cannot be served by implementation result Map<K, V1 | V2>
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1400:1: Adamic 0.1 refuses overload 2 of arrayToMap result Map<K, V2> cannot be served by implementation result Map<K, V1 | V2>; make the implementation result covariant with every overload result
```

### 7. gone — `src/compiler/executeCommandLine.ts:244:1`

base: Refused / lowering / `src/compiler/core.ts:1334:1`

```text
overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

### 8. reached — `src/compiler/executeCommandLine.ts:244:1`

head: NotYet / lowering / `src/compiler/executeCommandLine.ts:317:22`

```text
String conversion of an object, array, map or function (ToPrimitive is not lowered)
/workspace/scratch/overload-certificates-census/adapted/src/compiler/executeCommandLine.ts:317:22: stage 0 can't lower String conversion of an object, array, map or function (ToPrimitive is not lowered) yet
```

### 9. gone — `src/compiler/utilities.ts:1211:1`

base: Refused / lowering / `src/compiler/core.ts:1334:1`

```text
overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

## full — src/compiler/builder.ts

Before 452; after 451; unchanged 431; grades {'changed': 10, 'gone': 11, 'reached': 10}.

### 1. changed — `src/compiler/builder.ts:1258:1`

base: Boundary / lowering / `src/compiler/builder.ts:1421:9`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

head: Boundary / lowering / `src/compiler/builder.ts:1421:9`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/builder.ts:1421:25: stage 0 can't lower a certified overload whose implementation stores (IncrementalBuildInfoFileId | Path)[] where the call reads IncrementalBuildInfoFileId[] yet
```

### 2. changed — `src/compiler/builder.ts:1258:1`

base: Boundary / lowering / `src/compiler/builder.ts:1515:9`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

head: Boundary / lowering / `src/compiler/builder.ts:1515:9`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/builder.ts:1515:79: stage 0 can't lower a comparator that doesn't take two elements and return a number yet
```

### 3. changed — `src/compiler/builder.ts:1258:1`

base: Boundary / lowering / `src/compiler/builder.ts:1584:13`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

head: Boundary / lowering / `src/compiler/builder.ts:1584:13`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/builder.ts:1584:77: stage 0 can't lower a comparator that doesn't take two elements and return a number yet
```

### 4. changed — `src/compiler/builder.ts:1258:1`

base: Boundary / lowering / `src/compiler/builder.ts:1268:9`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

head: Boundary / lowering / `src/compiler/builder.ts:1268:9`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/builder.ts:1269:29: stage 0 can't lower a Map or Set going where a plain object type is held yet
```

### 5. changed — `src/compiler/builder.ts:1258:1`

base: Boundary / lowering / `src/compiler/builder.ts:1281:9`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

head: Boundary / lowering / `src/compiler/builder.ts:1315:5`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/builder.ts:1315:23: stage 0 can't lower a certified overload whose implementation stores ([Path, FileInfo] | IncrementalMultiFileEmitBuildInfoFileInfo)[] where the call reads IncrementalMultiFileEmitBuildInfoFileInfo[] yet
```

### 6. gone — `src/compiler/builder.ts:1258:1`

base: Refused / lowering / `src/compiler/core.ts:1334:1`

```text
overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

### 7. gone — `src/compiler/builder.ts:1258:1`

base: NotYet / lowering / `src/compiler/builder.ts:1291:13`

```text
reading fileInfos
/workspace/scratch/overload-certificates-census/adapted/src/compiler/builder.ts:1291:13: stage 0 can't lower reading fileInfos yet
```

### 8. gone — `src/compiler/builder.ts:1258:1`

base: Boundary / lowering / `src/compiler/builder.ts:1289:9`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/builder.ts:1291:13: stage 0 can't lower reading fileInfos yet
```

### 9. gone — `src/compiler/builder.ts:1258:1`

base: NotYet / lowering / `src/compiler/builder.ts:1309:16`

```text
reading buildInfo
/workspace/scratch/overload-certificates-census/adapted/src/compiler/builder.ts:1309:16: stage 0 can't lower reading buildInfo yet
```

### 10. gone — `src/compiler/builder.ts:1258:1`

base: Boundary / lowering / `src/compiler/builder.ts:1309:9`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/builder.ts:1309:16: stage 0 can't lower reading buildInfo yet
```

### 11. gone — `src/compiler/builder.ts:1258:1`

base: Boundary / lowering / `src/compiler/builder.ts:1315:5`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

### 12. reached — `src/compiler/builder.ts:1258:1`

head: NotYet / lowering / `src/compiler/builder.ts:1421:25`

```text
a certified overload whose implementation stores (IncrementalBuildInfoFileId | Path)[] where the call reads IncrementalBuildInfoFileId[]
/workspace/scratch/overload-certificates-census/adapted/src/compiler/builder.ts:1421:25: stage 0 can't lower a certified overload whose implementation stores (IncrementalBuildInfoFileId | Path)[] where the call reads IncrementalBuildInfoFileId[] yet
```

### 13. reached — `src/compiler/builder.ts:1258:1`

head: NotYet / lowering / `src/compiler/builder.ts:1515:79`

```text
a comparator that doesn't take two elements and return a number
/workspace/scratch/overload-certificates-census/adapted/src/compiler/builder.ts:1515:79: stage 0 can't lower a comparator that doesn't take two elements and return a number yet
```

### 14. reached — `src/compiler/builder.ts:1258:1`

head: NotYet / lowering / `src/compiler/builder.ts:1584:77`

```text
a comparator that doesn't take two elements and return a number
/workspace/scratch/overload-certificates-census/adapted/src/compiler/builder.ts:1584:77: stage 0 can't lower a comparator that doesn't take two elements and return a number yet
```

### 15. reached — `src/compiler/builder.ts:1258:1`

head: NotYet / lowering / `src/compiler/builder.ts:1269:29`

```text
a Map or Set going where a plain object type is held
/workspace/scratch/overload-certificates-census/adapted/src/compiler/builder.ts:1269:29: stage 0 can't lower a Map or Set going where a plain object type is held yet
```

### 16. reached — `src/compiler/builder.ts:1258:1`

head: NotYet / lowering / `src/compiler/builder.ts:1315:23`

```text
a certified overload whose implementation stores ([Path, FileInfo] | IncrementalMultiFileEmitBuildInfoFileInfo)[] where the call reads IncrementalMultiFileEmitBuildInfoFileInfo[]
/workspace/scratch/overload-certificates-census/adapted/src/compiler/builder.ts:1315:23: stage 0 can't lower a certified overload whose implementation stores ([Path, FileInfo] | IncrementalMultiFileEmitBuildInfoFileInfo)[] where the call reads IncrementalMultiFileEmitBuildInfoFileInfo[] yet
```

### 17. changed — `src/compiler/builder.ts:1420:5`

base: Boundary / lowering / `src/compiler/builder.ts:1421:9`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

head: Boundary / lowering / `src/compiler/builder.ts:1421:9`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/builder.ts:1421:25: stage 0 can't lower a certified overload whose implementation stores (IncrementalBuildInfoFileId | Path)[] where the call reads IncrementalBuildInfoFileId[] yet
```

### 18. gone — `src/compiler/builder.ts:1420:5`

base: Refused / lowering / `src/compiler/core.ts:1334:1`

```text
overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

### 19. reached — `src/compiler/builder.ts:1420:5`

head: NotYet / lowering / `src/compiler/builder.ts:1421:25`

```text
a certified overload whose implementation stores (IncrementalBuildInfoFileId | Path)[] where the call reads IncrementalBuildInfoFileId[]
/workspace/scratch/overload-certificates-census/adapted/src/compiler/builder.ts:1421:25: stage 0 can't lower a certified overload whose implementation stores (IncrementalBuildInfoFileId | Path)[] where the call reads IncrementalBuildInfoFileId[] yet
```

### 20. changed — `src/compiler/builder.ts:1512:5`

base: Boundary / lowering / `src/compiler/builder.ts:1515:9`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

head: Boundary / lowering / `src/compiler/builder.ts:1515:9`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/builder.ts:1515:79: stage 0 can't lower a comparator that doesn't take two elements and return a number yet
```

### 21. gone — `src/compiler/builder.ts:1512:5`

base: Refused / lowering / `src/compiler/core.ts:1334:1`

```text
overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

### 22. reached — `src/compiler/builder.ts:1512:5`

head: NotYet / lowering / `src/compiler/builder.ts:1515:79`

```text
a comparator that doesn't take two elements and return a number
/workspace/scratch/overload-certificates-census/adapted/src/compiler/builder.ts:1515:79: stage 0 can't lower a comparator that doesn't take two elements and return a number yet
```

### 23. changed — `src/compiler/builder.ts:1581:5`

base: Boundary / lowering / `src/compiler/builder.ts:1584:13`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

head: Boundary / lowering / `src/compiler/builder.ts:1584:13`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/builder.ts:1584:77: stage 0 can't lower a comparator that doesn't take two elements and return a number yet
```

### 24. gone — `src/compiler/builder.ts:1581:5`

base: Refused / lowering / `src/compiler/core.ts:1334:1`

```text
overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

### 25. reached — `src/compiler/builder.ts:1581:5`

head: NotYet / lowering / `src/compiler/builder.ts:1584:77`

```text
a comparator that doesn't take two elements and return a number
/workspace/scratch/overload-certificates-census/adapted/src/compiler/builder.ts:1584:77: stage 0 can't lower a comparator that doesn't take two elements and return a number yet
```

### 26. changed — `src/compiler/builder.ts:2277:1`

base: Boundary / lowering / `src/compiler/builder.ts:2416:9`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1400:1: Adamic 0.1 refuses overload 2 of arrayToMap result Map<K, V2> cannot be served by implementation result Map<K, V1 | V2>; make the implementation result covariant with every overload result
```

head: Boundary / lowering / `src/compiler/builder.ts:2416:9`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/builder.ts:2416:31: stage 0 can't lower a certified overload whose implementation stores Map<Path, V2 | IncrementalBuildInfoDiagnosticOfFile> where the call reads Map<Path, readonly ReusableDiagnostic[]> yet
```

### 27. gone — `src/compiler/builder.ts:2277:1`

base: Refused / lowering / `src/compiler/core.ts:1400:1`

```text
overload 2 of arrayToMap result Map<K, V2> cannot be served by implementation result Map<K, V1 | V2>
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1400:1: Adamic 0.1 refuses overload 2 of arrayToMap result Map<K, V2> cannot be served by implementation result Map<K, V1 | V2>; make the implementation result covariant with every overload result
```

### 28. reached — `src/compiler/builder.ts:2277:1`

head: NotYet / lowering / `src/compiler/builder.ts:2416:31`

```text
a certified overload whose implementation stores Map<Path, V2 | IncrementalBuildInfoDiagnosticOfFile> where the call reads Map<Path, readonly ReusableDiagnostic[]>
/workspace/scratch/overload-certificates-census/adapted/src/compiler/builder.ts:2416:31: stage 0 can't lower a certified overload whose implementation stores Map<Path, V2 | IncrementalBuildInfoDiagnosticOfFile> where the call reads Map<Path, readonly ReusableDiagnostic[]> yet
```

### 29. changed — `src/compiler/builder.ts:2415:5`

base: Boundary / lowering / `src/compiler/builder.ts:2416:9`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1400:1: Adamic 0.1 refuses overload 2 of arrayToMap result Map<K, V2> cannot be served by implementation result Map<K, V1 | V2>; make the implementation result covariant with every overload result
```

head: Boundary / lowering / `src/compiler/builder.ts:2416:9`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/builder.ts:2416:31: stage 0 can't lower a certified overload whose implementation stores Map<Path, V2 | IncrementalBuildInfoDiagnosticOfFile> where the call reads Map<Path, readonly ReusableDiagnostic[]> yet
```

### 30. gone — `src/compiler/builder.ts:2415:5`

base: Refused / lowering / `src/compiler/core.ts:1400:1`

```text
overload 2 of arrayToMap result Map<K, V2> cannot be served by implementation result Map<K, V1 | V2>
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1400:1: Adamic 0.1 refuses overload 2 of arrayToMap result Map<K, V2> cannot be served by implementation result Map<K, V1 | V2>; make the implementation result covariant with every overload result
```

### 31. reached — `src/compiler/builder.ts:2415:5`

head: NotYet / lowering / `src/compiler/builder.ts:2416:31`

```text
a certified overload whose implementation stores Map<Path, V2 | IncrementalBuildInfoDiagnosticOfFile> where the call reads Map<Path, readonly ReusableDiagnostic[]>
/workspace/scratch/overload-certificates-census/adapted/src/compiler/builder.ts:2416:31: stage 0 can't lower a certified overload whose implementation stores Map<Path, V2 | IncrementalBuildInfoDiagnosticOfFile> where the call reads Map<Path, readonly ReusableDiagnostic[]> yet
```

## full — src/compiler/builderState.ts

Before 40; after 36; unchanged 34; grades {'gone': 5, 'changed': 1, 'reached': 1}.

### 1. gone — `src/compiler/builderState.ts:465:5`

base: Refused / lowering / `src/compiler/core.ts:1334:1`

```text
overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

### 2. gone — `src/compiler/builderState.ts:465:5`

base: Boundary / lowering / `src/compiler/builderState.ts:493:9`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

### 3. gone — `src/compiler/builderState.ts:510:5`

base: Refused / lowering / `src/compiler/core.ts:1334:1`

```text
overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

### 4. gone — `src/compiler/builderState.ts:510:5`

base: Boundary / lowering / `src/compiler/builderState.ts:512:9`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

### 5. changed — `src/compiler/builderState.ts:588:5`

base: Boundary / lowering / `src/compiler/builderState.ts:624:9`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

head: Boundary / lowering / `src/compiler/builderState.ts:624:9`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/builderState.ts:624:45: stage 0 can't lower a Map or Set interface implemented by a value that isn't a Map or Set yet
```

### 6. gone — `src/compiler/builderState.ts:588:5`

base: Refused / lowering / `src/compiler/core.ts:1334:1`

```text
overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

### 7. reached — `src/compiler/builderState.ts:588:5`

head: NotYet / lowering / `src/compiler/builderState.ts:624:45`

```text
a Map or Set interface implemented by a value that isn't a Map or Set
/workspace/scratch/overload-certificates-census/adapted/src/compiler/builderState.ts:624:45: stage 0 can't lower a Map or Set interface implemented by a value that isn't a Map or Set yet
```

## full — src/compiler/checker.ts

Before 8450; after 8440; unchanged 8412; grades {'changed': 14, 'gone': 24, 'reached': 14}.

### 1. changed — `src/compiler/checker.ts:10013:13`

base: Boundary / lowering / `src/compiler/checker.ts:10014:17`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

head: Boundary / lowering / `src/compiler/checker.ts:10014:17`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/checker.ts:10014:41: stage 0 can't lower a Map or Set interface implemented by a value that isn't a Map or Set yet
```

### 2. gone — `src/compiler/checker.ts:10013:13`

base: Refused / lowering / `src/compiler/core.ts:1334:1`

```text
overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

### 3. reached — `src/compiler/checker.ts:10013:13`

head: NotYet / lowering / `src/compiler/checker.ts:10014:41`

```text
a Map or Set interface implemented by a value that isn't a Map or Set
/workspace/scratch/overload-certificates-census/adapted/src/compiler/checker.ts:10014:41: stage 0 can't lower a Map or Set interface implemented by a value that isn't a Map or Set yet
```

### 4. changed — `src/compiler/checker.ts:14598:5`

base: Boundary / lowering / `src/compiler/checker.ts:14646:13`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

head: Boundary / lowering / `src/compiler/checker.ts:14646:13`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/checker.ts:14646:76: stage 0 can't lower a Map or Set interface implemented by a value that isn't a Map or Set yet
```

### 5. gone — `src/compiler/checker.ts:14598:5`

base: Refused / lowering / `src/compiler/core.ts:1334:1`

```text
overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

### 6. reached — `src/compiler/checker.ts:14598:5`

head: NotYet / lowering / `src/compiler/checker.ts:14646:76`

```text
a Map or Set interface implemented by a value that isn't a Map or Set
/workspace/scratch/overload-certificates-census/adapted/src/compiler/checker.ts:14646:76: stage 0 can't lower a Map or Set interface implemented by a value that isn't a Map or Set yet
```

### 7. changed — `src/compiler/checker.ts:15132:5`

base: Boundary / lowering / `src/compiler/checker.ts:15148:9`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

head: Boundary / lowering / `src/compiler/checker.ts:15148:9`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/checker.ts:15148:26: stage 0 can't lower a Map or Set interface implemented by a value that isn't a Map or Set yet
```

### 8. gone — `src/compiler/checker.ts:15132:5`

base: Refused / lowering / `src/compiler/core.ts:1334:1`

```text
overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

### 9. reached — `src/compiler/checker.ts:15132:5`

head: NotYet / lowering / `src/compiler/checker.ts:15148:26`

```text
a Map or Set interface implemented by a value that isn't a Map or Set
/workspace/scratch/overload-certificates-census/adapted/src/compiler/checker.ts:15148:26: stage 0 can't lower a Map or Set interface implemented by a value that isn't a Map or Set yet
```

### 10. changed — `src/compiler/checker.ts:15570:5`

base: Boundary / lowering / `src/compiler/checker.ts:15688:9`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

head: Boundary / lowering / `src/compiler/checker.ts:15688:9`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/checker.ts:15688:43: stage 0 can't lower a Map or Set interface implemented by a value that isn't a Map or Set yet
```

### 11. gone — `src/compiler/checker.ts:15570:5`

base: Refused / lowering / `src/compiler/core.ts:1334:1`

```text
overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

### 12. reached — `src/compiler/checker.ts:15570:5`

head: NotYet / lowering / `src/compiler/checker.ts:15688:43`

```text
a Map or Set interface implemented by a value that isn't a Map or Set
/workspace/scratch/overload-certificates-census/adapted/src/compiler/checker.ts:15688:43: stage 0 can't lower a Map or Set interface implemented by a value that isn't a Map or Set yet
```

### 13. changed — `src/compiler/checker.ts:16690:5`

base: Boundary / lowering / `src/compiler/checker.ts:16692:9`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

head: Boundary / lowering / `src/compiler/checker.ts:16692:9`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/checker.ts:16692:80: stage 0 can't lower a Map or Set interface implemented by a value that isn't a Map or Set yet
```

### 14. gone — `src/compiler/checker.ts:16690:5`

base: Refused / lowering / `src/compiler/core.ts:1334:1`

```text
overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

### 15. reached — `src/compiler/checker.ts:16690:5`

head: NotYet / lowering / `src/compiler/checker.ts:16692:80`

```text
a Map or Set interface implemented by a value that isn't a Map or Set
/workspace/scratch/overload-certificates-census/adapted/src/compiler/checker.ts:16692:80: stage 0 can't lower a Map or Set interface implemented by a value that isn't a Map or Set yet
```

### 16. changed — `src/compiler/checker.ts:16696:5`

base: Boundary / lowering / `src/compiler/checker.ts:16696:5`

```text
function signature or prologue failed; continued at next declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

head: Boundary / lowering / `src/compiler/checker.ts:16696:5`

```text
function signature or prologue failed; continued at next declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/checker.ts:16696:132: stage 0 can't lower a Map or Set interface implemented by a value that isn't a Map or Set yet
```

### 17. gone — `src/compiler/checker.ts:16696:5`

base: Refused / lowering / `src/compiler/core.ts:1334:1`

```text
overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

### 18. reached — `src/compiler/checker.ts:16696:5`

head: NotYet / lowering / `src/compiler/checker.ts:16696:132`

```text
a Map or Set interface implemented by a value that isn't a Map or Set
/workspace/scratch/overload-certificates-census/adapted/src/compiler/checker.ts:16696:132: stage 0 can't lower a Map or Set interface implemented by a value that isn't a Map or Set yet
```

### 19. changed — `src/compiler/checker.ts:18662:5`

base: Boundary / lowering / `src/compiler/checker.ts:18665:9`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

head: Boundary / lowering / `src/compiler/checker.ts:18665:9`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/checker.ts:18665:43: stage 0 can't lower a Map or Set interface implemented by a value that isn't a Map or Set yet
```

### 20. gone — `src/compiler/checker.ts:18662:5`

base: Refused / lowering / `src/compiler/core.ts:1334:1`

```text
overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

### 21. reached — `src/compiler/checker.ts:18662:5`

head: NotYet / lowering / `src/compiler/checker.ts:18665:43`

```text
a Map or Set interface implemented by a value that isn't a Map or Set
/workspace/scratch/overload-certificates-census/adapted/src/compiler/checker.ts:18665:43: stage 0 can't lower a Map or Set interface implemented by a value that isn't a Map or Set yet
```

### 22. changed — `src/compiler/checker.ts:1966:5`

base: Boundary / lowering / `src/compiler/checker.ts:1985:9`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

head: Boundary / lowering / `src/compiler/checker.ts:1985:9`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/checker.ts:1985:26: stage 0 can't lower a Map or Set going where a plain object type is held yet
```

### 23. gone — `src/compiler/checker.ts:1966:5`

base: Refused / lowering / `src/compiler/core.ts:1334:1`

```text
overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

### 24. reached — `src/compiler/checker.ts:1966:5`

head: NotYet / lowering / `src/compiler/checker.ts:1985:26`

```text
a Map or Set going where a plain object type is held
/workspace/scratch/overload-certificates-census/adapted/src/compiler/checker.ts:1985:26: stage 0 can't lower a Map or Set going where a plain object type is held yet
```

### 25. changed — `src/compiler/checker.ts:22383:5`

base: Boundary / lowering / `src/compiler/checker.ts:24405:13`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

head: Boundary / lowering / `src/compiler/checker.ts:24414:17`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/checker.ts:24414:107: stage 0 can't lower a SpreadElement yet
```

### 26. gone — `src/compiler/checker.ts:22383:5`

base: Refused / lowering / `src/compiler/core.ts:1334:1`

```text
overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

### 27. gone — `src/compiler/checker.ts:22383:5`

base: NotYet / lowering / `src/compiler/checker.ts:24778:32`

```text
reading props
/workspace/scratch/overload-certificates-census/adapted/src/compiler/checker.ts:24778:32: stage 0 can't lower reading props yet
```

### 28. gone — `src/compiler/checker.ts:22383:5`

base: Boundary / lowering / `src/compiler/checker.ts:24778:13`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/checker.ts:24778:32: stage 0 can't lower reading props yet
```

### 29. reached — `src/compiler/checker.ts:22383:5`

head: NotYet / lowering / `src/compiler/checker.ts:24414:107`

```text
a SpreadElement
/workspace/scratch/overload-certificates-census/adapted/src/compiler/checker.ts:24414:107: stage 0 can't lower a SpreadElement yet
```

### 30. changed — `src/compiler/checker.ts:24381:9`

base: Boundary / lowering / `src/compiler/checker.ts:24405:13`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

head: Boundary / lowering / `src/compiler/checker.ts:24414:17`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/checker.ts:24414:107: stage 0 can't lower a call spreading something other than an array yet
```

### 31. gone — `src/compiler/checker.ts:24381:9`

base: Refused / lowering / `src/compiler/core.ts:1334:1`

```text
overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

### 32. gone — `src/compiler/checker.ts:24381:9`

base: NotYet / lowering / `src/compiler/checker.ts:24412:17`

```text
reading props
/workspace/scratch/overload-certificates-census/adapted/src/compiler/checker.ts:24412:17: stage 0 can't lower reading props yet
```

### 33. gone — `src/compiler/checker.ts:24381:9`

base: Boundary / lowering / `src/compiler/checker.ts:24412:13`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/checker.ts:24412:17: stage 0 can't lower reading props yet
```

### 34. reached — `src/compiler/checker.ts:24381:9`

head: NotYet / lowering / `src/compiler/checker.ts:24414:107`

```text
a call spreading something other than an array
/workspace/scratch/overload-certificates-census/adapted/src/compiler/checker.ts:24414:107: stage 0 can't lower a call spreading something other than an array yet
```

### 35. changed — `src/compiler/checker.ts:25965:5`

base: Boundary / lowering / `src/compiler/checker.ts:25975:13`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

head: Boundary / lowering / `src/compiler/checker.ts:25975:13`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/checker.ts:25975:52: stage 0 can't lower a Map or Set interface implemented by a value that isn't a Map or Set yet
```

### 36. gone — `src/compiler/checker.ts:25965:5`

base: Refused / lowering / `src/compiler/core.ts:1334:1`

```text
overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

### 37. reached — `src/compiler/checker.ts:25965:5`

head: NotYet / lowering / `src/compiler/checker.ts:25975:52`

```text
a Map or Set interface implemented by a value that isn't a Map or Set
/workspace/scratch/overload-certificates-census/adapted/src/compiler/checker.ts:25975:52: stage 0 can't lower a Map or Set interface implemented by a value that isn't a Map or Set yet
```

### 38. changed — `src/compiler/checker.ts:35277:5`

base: Boundary / lowering / `src/compiler/checker.ts:35283:9`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

head: Boundary / lowering / `src/compiler/checker.ts:35283:9`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/checker.ts:35283:36: stage 0 can't lower a Map or Set interface implemented by a value that isn't a Map or Set yet
```

### 39. gone — `src/compiler/checker.ts:35277:5`

base: Refused / lowering / `src/compiler/core.ts:1334:1`

```text
overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

### 40. reached — `src/compiler/checker.ts:35277:5`

head: NotYet / lowering / `src/compiler/checker.ts:35283:36`

```text
a Map or Set interface implemented by a value that isn't a Map or Set
/workspace/scratch/overload-certificates-census/adapted/src/compiler/checker.ts:35283:36: stage 0 can't lower a Map or Set interface implemented by a value that isn't a Map or Set yet
```

### 41. changed — `src/compiler/checker.ts:47699:5`

base: Boundary / lowering / `src/compiler/checker.ts:47718:9`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

head: Boundary / lowering / `src/compiler/checker.ts:47718:9`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/checker.ts:47718:26: stage 0 can't lower a Map or Set interface implemented by a value that isn't a Map or Set yet
```

### 42. gone — `src/compiler/checker.ts:47699:5`

base: Refused / lowering / `src/compiler/core.ts:1334:1`

```text
overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

### 43. reached — `src/compiler/checker.ts:47699:5`

head: NotYet / lowering / `src/compiler/checker.ts:47718:26`

```text
a Map or Set interface implemented by a value that isn't a Map or Set
/workspace/scratch/overload-certificates-census/adapted/src/compiler/checker.ts:47718:26: stage 0 can't lower a Map or Set interface implemented by a value that isn't a Map or Set yet
```

### 44. changed — `src/compiler/checker.ts:51536:9`

base: Boundary / lowering / `src/compiler/checker.ts:51543:13`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

head: Boundary / lowering / `src/compiler/checker.ts:51543:13`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/checker.ts:51543:39: stage 0 can't lower a Map or Set interface implemented by a value that isn't a Map or Set yet
```

### 45. gone — `src/compiler/checker.ts:51536:9`

base: Refused / lowering / `src/compiler/core.ts:1334:1`

```text
overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

### 46. reached — `src/compiler/checker.ts:51536:9`

head: NotYet / lowering / `src/compiler/checker.ts:51543:39`

```text
a Map or Set interface implemented by a value that isn't a Map or Set
/workspace/scratch/overload-certificates-census/adapted/src/compiler/checker.ts:51543:39: stage 0 can't lower a Map or Set interface implemented by a value that isn't a Map or Set yet
```

### 47. gone — `src/compiler/checker.ts:51569:5`

base: Refused / lowering / `src/compiler/core.ts:1334:1`

```text
overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

### 48. gone — `src/compiler/checker.ts:51569:5`

base: Boundary / lowering / `src/compiler/checker.ts:51685:17`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

### 49. gone — `src/compiler/checker.ts:51569:5`

base: NotYet / lowering / `src/compiler/checker.ts:51687:154`

```text
reading list
/workspace/scratch/overload-certificates-census/adapted/src/compiler/checker.ts:51687:154: stage 0 can't lower reading list yet
```

### 50. gone — `src/compiler/checker.ts:51569:5`

base: Boundary / lowering / `src/compiler/checker.ts:51686:17`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/checker.ts:51687:154: stage 0 can't lower reading list yet
```

### 51. gone — `src/compiler/checker.ts:51569:5`

base: NotYet / lowering / `src/compiler/checker.ts:51691:155`

```text
reading list
/workspace/scratch/overload-certificates-census/adapted/src/compiler/checker.ts:51691:155: stage 0 can't lower reading list yet
```

### 52. gone — `src/compiler/checker.ts:51569:5`

base: Boundary / lowering / `src/compiler/checker.ts:51690:17`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/checker.ts:51691:155: stage 0 can't lower reading list yet
```

## full — src/compiler/commandLineParser.ts

Before 509; after 496; unchanged 496; grades {'gone': 13}.

### 1. gone — `src/compiler/commandLineParser.ts:1873:1`

base: Refused / lowering / `src/compiler/core.ts:1334:1`

```text
overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

### 2. gone — `src/compiler/commandLineParser.ts:1873:1`

base: Boundary / lowering / `src/compiler/commandLineParser.ts:1874:5`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

### 3. gone — `src/compiler/commandLineParser.ts:1873:1`

base: NotYet / lowering / `src/compiler/commandLineParser.ts:1875:47`

```text
reading namesOfType
/workspace/scratch/overload-certificates-census/adapted/src/compiler/commandLineParser.ts:1875:47: stage 0 can't lower reading namesOfType yet
```

### 4. gone — `src/compiler/commandLineParser.ts:1873:1`

base: Boundary / lowering / `src/compiler/commandLineParser.ts:1875:5`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/commandLineParser.ts:1875:47: stage 0 can't lower reading namesOfType yet
```

### 5. gone — `src/compiler/commandLineParser.ts:1873:1`

base: NotYet / lowering / `src/compiler/commandLineParser.ts:1876:97`

```text
reading stringNames
/workspace/scratch/overload-certificates-census/adapted/src/compiler/commandLineParser.ts:1876:97: stage 0 can't lower reading stringNames yet
```

### 6. gone — `src/compiler/commandLineParser.ts:1873:1`

base: Boundary / lowering / `src/compiler/commandLineParser.ts:1876:5`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/commandLineParser.ts:1876:97: stage 0 can't lower reading stringNames yet
```

### 7. gone — `src/compiler/commandLineParser.ts:2307:1`

base: Refused / lowering / `src/compiler/core.ts:1400:1`

```text
overload 2 of arrayToMap result Map<K, V2> cannot be served by implementation result Map<K, V1 | V2>
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1400:1: Adamic 0.1 refuses overload 2 of arrayToMap result Map<K, V2> cannot be served by implementation result Map<K, V1 | V2>; make the implementation result covariant with every overload result
```

### 8. gone — `src/compiler/commandLineParser.ts:2307:1`

base: Boundary / lowering / `src/compiler/commandLineParser.ts:2308:5`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1400:1: Adamic 0.1 refuses overload 2 of arrayToMap result Map<K, V2> cannot be served by implementation result Map<K, V1 | V2>; make the implementation result covariant with every overload result
```

### 9. gone — `src/compiler/commandLineParser.ts:3950:1`

base: Refused / lowering / `src/compiler/core.ts:1334:1`

```text
overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

### 10. gone — `src/compiler/commandLineParser.ts:3950:1`

base: Boundary / lowering / `src/compiler/commandLineParser.ts:4033:5`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

### 11. gone — `src/compiler/commandLineParser.ts:3950:1`

base: Boundary / lowering / `src/compiler/commandLineParser.ts:4034:5`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

### 12. gone — `src/compiler/commandLineParser.ts:3950:1`

base: NotYet / lowering / `src/compiler/commandLineParser.ts:4036:12`

```text
reading literalFiles
/workspace/scratch/overload-certificates-census/adapted/src/compiler/commandLineParser.ts:4036:12: stage 0 can't lower reading literalFiles yet
```

### 13. gone — `src/compiler/commandLineParser.ts:3950:1`

base: Boundary / lowering / `src/compiler/commandLineParser.ts:4036:5`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/commandLineParser.ts:4036:12: stage 0 can't lower reading literalFiles yet
```

## full — src/compiler/core.ts

Before 14; after 14; unchanged 14; grades {}.

No finding tuple changes.

## full — src/compiler/executeCommandLine.ts

Before 390; after 390; unchanged 384; grades {'changed': 3, 'gone': 3, 'reached': 3}.

### 1. changed — `src/compiler/executeCommandLine.ts:244:1`

base: Boundary / lowering / `src/compiler/executeCommandLine.ts:311:9`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

head: Boundary / lowering / `src/compiler/executeCommandLine.ts:311:9`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/executeCommandLine.ts:317:22: stage 0 can't lower String conversion of an object, array, map or function (ToPrimitive is not lowered) yet
```

### 2. gone — `src/compiler/executeCommandLine.ts:244:1`

base: Refused / lowering / `src/compiler/core.ts:1334:1`

```text
overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

### 3. reached — `src/compiler/executeCommandLine.ts:244:1`

head: NotYet / lowering / `src/compiler/executeCommandLine.ts:317:22`

```text
String conversion of an object, array, map or function (ToPrimitive is not lowered)
/workspace/scratch/overload-certificates-census/adapted/src/compiler/executeCommandLine.ts:317:22: stage 0 can't lower String conversion of an object, array, map or function (ToPrimitive is not lowered) yet
```

### 4. changed — `src/compiler/executeCommandLine.ts:307:5`

base: Boundary / lowering / `src/compiler/executeCommandLine.ts:311:9`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

head: Boundary / lowering / `src/compiler/executeCommandLine.ts:311:9`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/executeCommandLine.ts:317:22: stage 0 can't lower String conversion of an object, array, map or function (ToPrimitive is not lowered) yet
```

### 5. gone — `src/compiler/executeCommandLine.ts:307:5`

base: Refused / lowering / `src/compiler/core.ts:1334:1`

```text
overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

### 6. reached — `src/compiler/executeCommandLine.ts:307:5`

head: NotYet / lowering / `src/compiler/executeCommandLine.ts:317:22`

```text
String conversion of an object, array, map or function (ToPrimitive is not lowered)
/workspace/scratch/overload-certificates-census/adapted/src/compiler/executeCommandLine.ts:317:22: stage 0 can't lower String conversion of an object, array, map or function (ToPrimitive is not lowered) yet
```

### 7. changed — `src/compiler/executeCommandLine.ts:812:1`

base: Boundary / lowering / `src/compiler/watchUtilities.ts:414:5`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1400:1: Adamic 0.1 refuses overload 2 of arrayToMap result Map<K, V2> cannot be served by implementation result Map<K, V1 | V2>; make the implementation result covariant with every overload result
```

head: Boundary / lowering / `src/compiler/watchUtilities.ts:414:5`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/watchUtilities.ts:414:40: stage 0 can't lower optional record property access yet
```

### 8. gone — `src/compiler/executeCommandLine.ts:812:1`

base: Refused / lowering / `src/compiler/core.ts:1400:1`

```text
overload 2 of arrayToMap result Map<K, V2> cannot be served by implementation result Map<K, V1 | V2>
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1400:1: Adamic 0.1 refuses overload 2 of arrayToMap result Map<K, V2> cannot be served by implementation result Map<K, V1 | V2>; make the implementation result covariant with every overload result
```

### 9. reached — `src/compiler/executeCommandLine.ts:812:1`

head: NotYet / lowering / `src/compiler/watchUtilities.ts:414:40`

```text
optional record property access
/workspace/scratch/overload-certificates-census/adapted/src/compiler/watchUtilities.ts:414:40: stage 0 can't lower optional record property access yet
```

## full — src/compiler/moduleSpecifiers.ts

Before 183; after 183; unchanged 181; grades {'changed': 1, 'gone': 1, 'reached': 1}.

### 1. changed — `src/compiler/moduleSpecifiers.ts:809:1`

base: Boundary / lowering / `src/compiler/moduleSpecifiers.ts:868:9`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

head: Boundary / lowering / `src/compiler/moduleSpecifiers.ts:868:9`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/moduleSpecifiers.ts:870:25: stage 0 can't lower a nested destructuring pattern yet
```

### 2. gone — `src/compiler/moduleSpecifiers.ts:809:1`

base: Refused / lowering / `src/compiler/core.ts:1334:1`

```text
overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

### 3. reached — `src/compiler/moduleSpecifiers.ts:809:1`

head: NotYet / lowering / `src/compiler/moduleSpecifiers.ts:870:25`

```text
a nested destructuring pattern
/workspace/scratch/overload-certificates-census/adapted/src/compiler/moduleSpecifiers.ts:870:25: stage 0 can't lower a nested destructuring pattern yet
```

## full — src/compiler/sourcemap.ts

Before 131; after 123; unchanged 123; grades {'gone': 8}.

### 1. gone — `src/compiler/sourcemap.ts:702:1`

base: Refused / lowering / `src/compiler/core.ts:1334:1`

```text
overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

### 2. gone — `src/compiler/sourcemap.ts:702:1`

base: Boundary / lowering / `src/compiler/sourcemap.ts:743:13`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

### 3. gone — `src/compiler/sourcemap.ts:702:1`

base: NotYet / lowering / `src/compiler/sourcemap.ts:751:35`

```text
reading mappings
/workspace/scratch/overload-certificates-census/adapted/src/compiler/sourcemap.ts:751:35: stage 0 can't lower reading mappings yet
```

### 4. gone — `src/compiler/sourcemap.ts:702:1`

base: Boundary / lowering / `src/compiler/sourcemap.ts:751:17`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/sourcemap.ts:751:35: stage 0 can't lower reading mappings yet
```

### 5. gone — `src/compiler/sourcemap.ts:740:5`

base: Refused / lowering / `src/compiler/core.ts:1334:1`

```text
overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

### 6. gone — `src/compiler/sourcemap.ts:740:5`

base: Boundary / lowering / `src/compiler/sourcemap.ts:743:13`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

### 7. gone — `src/compiler/sourcemap.ts:740:5`

base: NotYet / lowering / `src/compiler/sourcemap.ts:751:35`

```text
reading mappings
/workspace/scratch/overload-certificates-census/adapted/src/compiler/sourcemap.ts:751:35: stage 0 can't lower reading mappings yet
```

### 8. gone — `src/compiler/sourcemap.ts:740:5`

base: Boundary / lowering / `src/compiler/sourcemap.ts:751:17`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/sourcemap.ts:751:35: stage 0 can't lower reading mappings yet
```

## full — src/compiler/transformers/esnext.ts

Before 253; after 251; unchanged 249; grades {'changed': 2, 'gone': 2}.

### 1. changed — `src/compiler/transformers/esnext.ts:127:5`

base: Boundary / lowering / `src/compiler/transformers/esnext.ts:199:17`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

head: Boundary / lowering / `src/compiler/transformers/esnext.ts:199:17`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/transformers/utilities.ts:404:5: stage 0 can't lower a function returning this yet
```

### 2. gone — `src/compiler/transformers/esnext.ts:127:5`

base: Refused / lowering / `src/compiler/core.ts:1334:1`

```text
overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

### 3. changed — `src/compiler/transformers/esnext.ts:72:1`

base: Boundary / lowering / `src/compiler/transformers/esnext.ts:199:17`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

head: Boundary / lowering / `src/compiler/transformers/esnext.ts:654:13`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/transformers/utilities.ts:404:5: stage 0 can't lower a function returning this yet
```

### 4. gone — `src/compiler/transformers/esnext.ts:72:1`

base: Refused / lowering / `src/compiler/core.ts:1334:1`

```text
overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

## full — src/compiler/transformers/jsx.ts

Before 182; after 182; unchanged 178; grades {'changed': 2, 'gone': 2, 'reached': 2}.

### 1. changed — `src/compiler/transformers/jsx.ts:158:5`

base: Boundary / lowering / `src/compiler/transformers/jsx.ts:173:13`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

head: Boundary / lowering / `src/compiler/transformers/jsx.ts:173:13`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/transformers/jsx.ts:173:73: stage 0 can't lower a Map or Set interface implemented by a value that isn't a Map or Set yet
```

### 2. gone — `src/compiler/transformers/jsx.ts:158:5`

base: Refused / lowering / `src/compiler/core.ts:1334:1`

```text
overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

### 3. reached — `src/compiler/transformers/jsx.ts:158:5`

head: NotYet / lowering / `src/compiler/transformers/jsx.ts:173:73`

```text
a Map or Set interface implemented by a value that isn't a Map or Set
/workspace/scratch/overload-certificates-census/adapted/src/compiler/transformers/jsx.ts:173:73: stage 0 can't lower a Map or Set interface implemented by a value that isn't a Map or Set yet
```

### 4. changed — `src/compiler/transformers/jsx.ts:93:1`

base: Boundary / lowering / `src/compiler/transformers/jsx.ts:173:13`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

head: Boundary / lowering / `src/compiler/transformers/jsx.ts:173:13`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/transformers/jsx.ts:173:73: stage 0 can't lower a Map or Set interface implemented by a value that isn't a Map or Set yet
```

### 5. gone — `src/compiler/transformers/jsx.ts:93:1`

base: Refused / lowering / `src/compiler/core.ts:1334:1`

```text
overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

### 6. reached — `src/compiler/transformers/jsx.ts:93:1`

head: NotYet / lowering / `src/compiler/transformers/jsx.ts:173:73`

```text
a Map or Set interface implemented by a value that isn't a Map or Set
/workspace/scratch/overload-certificates-census/adapted/src/compiler/transformers/jsx.ts:173:73: stage 0 can't lower a Map or Set interface implemented by a value that isn't a Map or Set yet
```

## full — src/compiler/transformers/module/module.ts

Before 398; after 398; unchanged 396; grades {'changed': 1, 'gone': 1, 'reached': 1}.

### 1. changed — `src/compiler/transformers/module/module.ts:2459:5`

base: Boundary / lowering / `src/compiler/transformers/module/module.ts:2480:21`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

head: Boundary / lowering / `src/compiler/transformers/module/module.ts:2480:21`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/transformers/module/module.ts:2480:50: stage 0 can't lower a Map or Set going where a plain object type is held yet
```

### 2. gone — `src/compiler/transformers/module/module.ts:2459:5`

base: Refused / lowering / `src/compiler/core.ts:1334:1`

```text
overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

### 3. reached — `src/compiler/transformers/module/module.ts:2459:5`

head: NotYet / lowering / `src/compiler/transformers/module/module.ts:2480:50`

```text
a Map or Set going where a plain object type is held
/workspace/scratch/overload-certificates-census/adapted/src/compiler/transformers/module/module.ts:2480:50: stage 0 can't lower a Map or Set going where a plain object type is held yet
```

## full — src/compiler/tsbuildPublic.ts

Before 55; after 55; unchanged 55; grades {}.

No finding tuple changes.

## full — src/compiler/utilities.ts

Before 1344; after 1340; unchanged 1340; grades {'gone': 4}.

### 1. gone — `src/compiler/utilities.ts:1211:1`

base: Refused / lowering / `src/compiler/core.ts:1334:1`

```text
overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

### 2. gone — `src/compiler/utilities.ts:1211:1`

base: Boundary / lowering / `src/compiler/utilities.ts:1224:9`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

### 3. gone — `src/compiler/utilities.ts:1223:5`

base: Refused / lowering / `src/compiler/core.ts:1334:1`

```text
overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

### 4. gone — `src/compiler/utilities.ts:1223:5`

base: Boundary / lowering / `src/compiler/utilities.ts:1224:9`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

## full — src/compiler/watchUtilities.ts

Before 52; after 52; unchanged 52; grades {}.

No finding tuple changes.

## full — src/harness/watchUtils.ts

Before 0; after 0; unchanged 0; grades {}.

No finding tuple changes.

## full — src/server/editorServices.ts

Before 1073; after 1073; unchanged 1073; grades {}.

No finding tuple changes.

## full — src/server/project.ts

Before 690; after 699; unchanged 687; grades {'changed': 2, 'gone': 1, 'reached': 10}.

### 1. changed — `src/server/project.ts:2566:1`

base: Boundary / lowering / `src/server/project.ts:2593:13`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

head: Boundary / lowering / `src/server/project.ts:2612:21`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/server/project.ts:2612:91: stage 0 can't lower reading symlinkCache yet
```

### 2. changed — `src/server/project.ts:2566:1`

base: Boundary / lowering / `src/server/project.ts:2687:9`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

head: Boundary / lowering / `src/server/project.ts:2613:21`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/server/project.ts:2613:25: stage 0 can't lower reading entrypoints yet
```

### 3. gone — `src/server/project.ts:2566:1`

base: Refused / lowering / `src/compiler/core.ts:1334:1`

```text
overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

### 4. reached — `src/server/project.ts:2566:1`

head: NotYet / lowering / `src/server/project.ts:2612:91`

```text
reading symlinkCache
/workspace/scratch/overload-certificates-census/adapted/src/server/project.ts:2612:91: stage 0 can't lower reading symlinkCache yet
```

### 5. reached — `src/server/project.ts:2566:1`

head: NotYet / lowering / `src/server/project.ts:2613:25`

```text
reading entrypoints
/workspace/scratch/overload-certificates-census/adapted/src/server/project.ts:2613:25: stage 0 can't lower reading entrypoints yet
```

### 6. reached — `src/server/project.ts:2566:1`

head: NotYet / lowering / `src/server/project.ts:2638:21`

```text
reading done
/workspace/scratch/overload-certificates-census/adapted/src/server/project.ts:2638:21: stage 0 can't lower reading done yet
```

### 7. reached — `src/server/project.ts:2566:1`

head: Boundary / lowering / `src/server/project.ts:2638:17`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/server/project.ts:2638:21: stage 0 can't lower reading done yet
```

### 8. reached — `src/server/project.ts:2566:1`

head: NotYet / lowering / `src/server/project.ts:2644:91`

```text
reading symlinkCache
/workspace/scratch/overload-certificates-census/adapted/src/server/project.ts:2644:91: stage 0 can't lower reading symlinkCache yet
```

### 9. reached — `src/server/project.ts:2566:1`

head: Boundary / lowering / `src/server/project.ts:2644:21`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/server/project.ts:2644:91: stage 0 can't lower reading symlinkCache yet
```

### 10. reached — `src/server/project.ts:2566:1`

head: NotYet / lowering / `src/server/project.ts:2645:55`

```text
reading entrypoints
/workspace/scratch/overload-certificates-census/adapted/src/server/project.ts:2645:55: stage 0 can't lower reading entrypoints yet
```

### 11. reached — `src/server/project.ts:2566:1`

head: Boundary / lowering / `src/server/project.ts:2645:21`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/server/project.ts:2645:55: stage 0 can't lower reading entrypoints yet
```

### 12. reached — `src/server/project.ts:2566:1`

head: Boundary / lowering / `src/server/project.ts:2755:9`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/server/project.ts:471:22: stage 0 can't lower a function without a body yet
```

### 13. reached — `src/server/project.ts:2566:1`

head: Boundary / lowering / `src/server/project.ts:2765:9`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/server/project.ts:471:22: stage 0 can't lower a function without a body yet
```

## full — src/server/session.ts

Before 1118; after 1118; unchanged 1118; grades {}.

No finding tuple changes.

## full — src/services/codeFixProvider.ts

Before 14; after 12; unchanged 12; grades {'gone': 2}.

### 1. gone — `src/services/codeFixProvider.ts:67:1`

base: Refused / lowering / `src/compiler/core.ts:1334:1`

```text
overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

### 2. gone — `src/services/codeFixProvider.ts:67:1`

base: Boundary / lowering / `src/services/codeFixProvider.ts:68:5`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

## full — src/services/codefixes/convertToEsModule.ts

Before 233; after 227; unchanged 227; grades {'gone': 6}.

### 1. gone — `src/services/codefixes/convertToEsModule.ts:529:1`

base: Refused / lowering / `src/compiler/core.ts:1334:1`

```text
overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

### 2. gone — `src/services/codefixes/convertToEsModule.ts:529:1`

base: Boundary / lowering / `src/services/codefixes/convertToEsModule.ts:568:5`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

### 3. gone — `src/services/codefixes/convertToEsModule.ts:529:1`

base: NotYet / lowering / `src/services/codefixes/convertToEsModule.ts:569:10`

```text
reading namedBindings
/workspace/scratch/overload-certificates-census/adapted/src/services/codefixes/convertToEsModule.ts:569:10: stage 0 can't lower reading namedBindings yet
```

### 4. gone — `src/services/codefixes/convertToEsModule.ts:529:1`

base: Boundary / lowering / `src/services/codefixes/convertToEsModule.ts:569:5`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/services/codefixes/convertToEsModule.ts:569:10: stage 0 can't lower reading namedBindings yet
```

### 5. gone — `src/services/codefixes/convertToEsModule.ts:529:1`

base: NotYet / lowering / `src/services/codefixes/convertToEsModule.ts:574:84`

```text
reading namedBindings
/workspace/scratch/overload-certificates-census/adapted/src/services/codefixes/convertToEsModule.ts:574:84: stage 0 can't lower reading namedBindings yet
```

### 6. gone — `src/services/codefixes/convertToEsModule.ts:529:1`

base: Boundary / lowering / `src/services/codefixes/convertToEsModule.ts:573:5`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/services/codefixes/convertToEsModule.ts:574:84: stage 0 can't lower reading namedBindings yet
```

## full — src/services/codefixes/fixAddMissingMember.ts

Before 75; after 75; unchanged 75; grades {}.

No finding tuple changes.

## full — src/services/codefixes/helpers.ts

Before 287; after 287; unchanged 285; grades {'changed': 1, 'gone': 1, 'reached': 1}.

### 1. changed — `src/services/codefixes/helpers.ts:574:1`

base: Boundary / lowering / `src/services/codefixes/helpers.ts:586:5`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

head: Boundary / lowering / `src/services/codefixes/helpers.ts:586:5`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/services/codefixes/helpers.ts:586:12: stage 0 can't lower a certified overload whose implementation stores (string | TypeParameterDeclaration)[] where the call reads TypeParameterDeclaration[] yet
```

### 2. gone — `src/services/codefixes/helpers.ts:574:1`

base: Refused / lowering / `src/compiler/core.ts:1334:1`

```text
overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

### 3. reached — `src/services/codefixes/helpers.ts:574:1`

head: NotYet / lowering / `src/services/codefixes/helpers.ts:586:12`

```text
a certified overload whose implementation stores (string | TypeParameterDeclaration)[] where the call reads TypeParameterDeclaration[]
/workspace/scratch/overload-certificates-census/adapted/src/services/codefixes/helpers.ts:586:12: stage 0 can't lower a certified overload whose implementation stores (string | TypeParameterDeclaration)[] where the call reads TypeParameterDeclaration[] yet
```

## full — src/services/codefixes/importFixes.ts

Before 280; after 282; unchanged 278; grades {'changed': 1, 'gone': 1, 'reached': 3}.

### 1. changed — `src/services/codefixes/importFixes.ts:1571:1`

base: Boundary / lowering / `src/services/codefixes/importFixes.ts:1582:9`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

head: Boundary / lowering / `src/compiler/core.ts:441:9`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:441:9: stage 0 can't lower a generator function delegating with yield* (#qk8rztp) yet
```

### 2. gone — `src/services/codefixes/importFixes.ts:1571:1`

base: Refused / lowering / `src/compiler/core.ts:1334:1`

```text
overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

### 3. reached — `src/services/codefixes/importFixes.ts:1571:1`

head: NotYet / lowering / `src/compiler/core.ts:441:9`

```text
a generator function delegating with yield* (#qk8rztp)
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:441:9: stage 0 can't lower a generator function delegating with yield* (#qk8rztp) yet
```

### 4. reached — `src/services/codefixes/importFixes.ts:1571:1`

head: NotYet / lowering / `src/services/codefixes/importFixes.ts:1583:29`

```text
a Map or Set interface implemented by a value that isn't a Map or Set
/workspace/scratch/overload-certificates-census/adapted/src/services/codefixes/importFixes.ts:1583:29: stage 0 can't lower a Map or Set interface implemented by a value that isn't a Map or Set yet
```

### 5. reached — `src/services/codefixes/importFixes.ts:1571:1`

head: Boundary / lowering / `src/services/codefixes/importFixes.ts:1582:9`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/services/codefixes/importFixes.ts:1583:29: stage 0 can't lower a Map or Set interface implemented by a value that isn't a Map or Set yet
```

## full — src/services/documentHighlights.ts

Before 89; after 89; unchanged 87; grades {'changed': 1, 'gone': 1, 'reached': 1}.

### 1. changed — `src/services/documentHighlights.ts:113:5`

base: Boundary / lowering / `src/services/documentHighlights.ts:119:9`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

head: Boundary / lowering / `src/services/documentHighlights.ts:119:9`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/services/documentHighlights.ts:119:45: stage 0 can't lower reading map yet
```

### 2. gone — `src/services/documentHighlights.ts:113:5`

base: Refused / lowering / `src/compiler/core.ts:1334:1`

```text
overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

### 3. reached — `src/services/documentHighlights.ts:113:5`

head: NotYet / lowering / `src/services/documentHighlights.ts:119:45`

```text
reading map
/workspace/scratch/overload-certificates-census/adapted/src/services/documentHighlights.ts:119:45: stage 0 can't lower reading map yet
```

## full — src/services/documentRegistry.ts

Before 47; after 47; unchanged 47; grades {}.

No finding tuple changes.

## full — src/services/refactorProvider.ts

Before 2; after 4; unchanged 0; grades {'changed': 1, 'gone': 1, 'reached': 3}.

### 1. changed — `src/services/refactorProvider.ts:26:1`

base: Boundary / lowering / `src/services/refactorProvider.ts:27:5`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

head: Boundary / lowering / `src/compiler/core.ts:441:9`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:441:9: stage 0 can't lower a generator function delegating with yield* (#qk8rztp) yet
```

### 2. gone — `src/services/refactorProvider.ts:26:1`

base: Refused / lowering / `src/compiler/core.ts:1334:1`

```text
overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

### 3. reached — `src/services/refactorProvider.ts:26:1`

head: NotYet / lowering / `src/compiler/core.ts:441:9`

```text
a generator function delegating with yield* (#qk8rztp)
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:441:9: stage 0 can't lower a generator function delegating with yield* (#qk8rztp) yet
```

### 4. reached — `src/services/refactorProvider.ts:26:1`

head: NotYet / lowering / `src/services/refactorProvider.ts:27:38`

```text
a Map or Set interface implemented by a value that isn't a Map or Set
/workspace/scratch/overload-certificates-census/adapted/src/services/refactorProvider.ts:27:38: stage 0 can't lower a Map or Set interface implemented by a value that isn't a Map or Set yet
```

### 5. reached — `src/services/refactorProvider.ts:26:1`

head: Boundary / lowering / `src/services/refactorProvider.ts:27:5`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/services/refactorProvider.ts:27:38: stage 0 can't lower a Map or Set interface implemented by a value that isn't a Map or Set yet
```

## full — src/services/refactors/convertImport.ts

Before 43; after 43; unchanged 41; grades {'changed': 1, 'gone': 1, 'reached': 1}.

### 1. changed — `src/services/refactors/convertImport.ts:213:1`

base: Boundary / lowering / `src/services/refactors/convertImport.ts:276:9`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

head: Boundary / lowering / `src/services/refactors/convertImport.ts:276:9`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/services/refactors/convertImport.ts:276:62: stage 0 can't lower a Map or Set interface implemented by a value that isn't a Map or Set yet
```

### 2. gone — `src/services/refactors/convertImport.ts:213:1`

base: Refused / lowering / `src/compiler/core.ts:1334:1`

```text
overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

### 3. reached — `src/services/refactors/convertImport.ts:213:1`

head: NotYet / lowering / `src/services/refactors/convertImport.ts:276:62`

```text
a Map or Set interface implemented by a value that isn't a Map or Set
/workspace/scratch/overload-certificates-census/adapted/src/services/refactors/convertImport.ts:276:62: stage 0 can't lower a Map or Set interface implemented by a value that isn't a Map or Set yet
```

## full — src/services/refactors/extractSymbol.ts

Before 269; after 269; unchanged 269; grades {}.

No finding tuple changes.

## full — src/services/refactors/moveToFile.ts

Before 174; after 174; unchanged 172; grades {'changed': 1, 'gone': 1, 'reached': 1}.

### 1. changed — `src/services/refactors/moveToFile.ts:239:1`

base: Boundary / lowering / `src/services/refactors/moveToFile.ts:268:5`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

head: Boundary / lowering / `src/services/refactors/moveToFile.ts:268:5`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/services/refactors/moveToFile.ts:268:60: stage 0 can't lower a Map or Set interface implemented by a value that isn't a Map or Set yet
```

### 2. gone — `src/services/refactors/moveToFile.ts:239:1`

base: Refused / lowering / `src/compiler/core.ts:1334:1`

```text
overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

### 3. reached — `src/services/refactors/moveToFile.ts:239:1`

head: NotYet / lowering / `src/services/refactors/moveToFile.ts:268:60`

```text
a Map or Set interface implemented by a value that isn't a Map or Set
/workspace/scratch/overload-certificates-census/adapted/src/services/refactors/moveToFile.ts:268:60: stage 0 can't lower a Map or Set interface implemented by a value that isn't a Map or Set yet
```

## full — src/services/stringCompletions.ts

Before 149; after 145; unchanged 145; grades {'gone': 4}.

### 1. gone — `src/services/stringCompletions.ts:757:1`

base: Refused / lowering / `src/compiler/core.ts:1334:1`

```text
overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

### 2. gone — `src/services/stringCompletions.ts:757:1`

base: Boundary / lowering / `src/services/stringCompletions.ts:762:5`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

### 3. gone — `src/services/stringCompletions.ts:995:1`

base: Refused / lowering / `src/compiler/core.ts:1334:1`

```text
overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

### 4. gone — `src/services/stringCompletions.ts:995:1`

base: Boundary / lowering / `src/services/stringCompletions.ts:1098:5`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

## full — src/services/symbolDisplay.ts

Before 23; after 21; unchanged 21; grades {'gone': 2}.

### 1. gone — `src/services/symbolDisplay.ts:232:1`

base: Refused / lowering / `src/compiler/core.ts:1334:1`

```text
overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

### 2. gone — `src/services/symbolDisplay.ts:232:1`

base: Boundary / lowering / `src/services/symbolDisplay.ts:249:5`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1334:1: Adamic 0.1 refuses overload 1 of arrayFrom result U[] cannot be served by implementation result (T | U)[]; make the implementation result covariant with every overload result
```

## full — src/testRunner/projectsRunner.ts

Before 54; after 54; unchanged 54; grades {}.

No finding tuple changes.

## full — src/testRunner/unittests/compilerCore.ts

Before 3; after 3; unchanged 3; grades {}.

No finding tuple changes.

## full — src/testRunner/unittests/customTransforms.ts

Before 21; after 21; unchanged 21; grades {}.

No finding tuple changes.

## full — src/testRunner/unittests/helpers.ts

Before 38; after 38; unchanged 36; grades {'changed': 1, 'gone': 1, 'reached': 1}.

### 1. changed — `src/testRunner/unittests/helpers.ts:106:1`

base: Boundary / lowering / `src/testRunner/unittests/helpers.ts:107:5`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1400:1: Adamic 0.1 refuses overload 2 of arrayToMap result Map<K, V2> cannot be served by implementation result Map<K, V1 | V2>; make the implementation result covariant with every overload result
```

head: Boundary / lowering / `src/testRunner/unittests/helpers.ts:107:5`

```text
failed statement; state rolled back; continued at next statement or declaration
/workspace/scratch/overload-certificates-census/adapted/src/testRunner/unittests/helpers.ts:107:19: stage 0 can't lower a certified overload whose implementation stores Map<string, V2 | NamedSourceText> where the call reads Map<string, SourceFileWithText> yet
```

### 2. gone — `src/testRunner/unittests/helpers.ts:106:1`

base: Refused / lowering / `src/compiler/core.ts:1400:1`

```text
overload 2 of arrayToMap result Map<K, V2> cannot be served by implementation result Map<K, V1 | V2>
/workspace/scratch/overload-certificates-census/adapted/src/compiler/core.ts:1400:1: Adamic 0.1 refuses overload 2 of arrayToMap result Map<K, V2> cannot be served by implementation result Map<K, V1 | V2>; make the implementation result covariant with every overload result
```

### 3. reached — `src/testRunner/unittests/helpers.ts:106:1`

head: NotYet / lowering / `src/testRunner/unittests/helpers.ts:107:19`

```text
a certified overload whose implementation stores Map<string, V2 | NamedSourceText> where the call reads Map<string, SourceFileWithText>
/workspace/scratch/overload-certificates-census/adapted/src/testRunner/unittests/helpers.ts:107:19: stage 0 can't lower a certified overload whose implementation stores Map<string, V2 | NamedSourceText> where the call reads Map<string, SourceFileWithText> yet
```

## full — src/testRunner/unittests/helpers/tsserver.ts

Before 181; after 181; unchanged 181; grades {}.

No finding tuple changes.

## full — src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts

Before 372; after 372; unchanged 372; grades {}.

No finding tuple changes.

## full — src/testRunner/unittests/moduleResolution.ts

Before 141; after 141; unchanged 141; grades {}.

No finding tuple changes.

## full — src/testRunner/unittests/programApi.ts

Before 48; after 48; unchanged 48; grades {}.

No finding tuple changes.

## full — src/testRunner/unittests/reuseProgramStructure.ts

Before 76; after 76; unchanged 76; grades {}.

No finding tuple changes.

## full — src/testRunner/unittests/tscWatch/incremental.ts

Before 51; after 51; unchanged 51; grades {}.

No finding tuple changes.

## full-directory — src/compiler/core.ts

Before 14; after 14; unchanged 14; grades {}.

No finding tuple changes.

## scan — src/compiler/core.ts

Scan-all base boundary: production abort at lowering.declareModule

Head boundary: production abort at lowering.declareModule

## scan — src/tsc/tsc.ts

Scan-all base boundary: production abort at lowering.declareModule

Head boundary: production abort at lowering.declareModule
