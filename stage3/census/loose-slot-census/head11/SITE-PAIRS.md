# Paired type-level sites: -7 to -11

Pair key: compiler-relative file, line, column. Both adapted compiler inputs have identical file hashes. This compares partial type-level observations, not successful lowerings.

22 stand; 0 gone; 0 new. Captured catch-cell stops and direct caught-value reads are marked separately.

| Site | -7 form | -11 form | Status | Captured catch-cell stop | Caught-value binding read at site |
| --- | --- | --- | --- | --- | --- |
| checker.ts:15346:21 | write | write | stands | no | — |
| checker.ts:25287:55 | captured | captured | stands | no | — |
| commandLineParser.ts:2691:9 | literal init | literal init | stands | no | — |
| commandLineParser.ts:2694:9 | literal init | literal init | stands | no | — |
| commandLineParser.ts:3093:11 | captured | captured | stands | no | — |
| commandLineParser.ts:3117:9 | literal init | literal init | stands | no | — |
| commandLineParser.ts:3438:18 | literal init | literal init | stands | no | — |
| commandLineParser.ts:3445:11 | captured | captured | stands | no | — |
| commandLineParser.ts:3468:73 | write | write | stands | no | — |
| commandLineParser.ts:3481:19 | captured | captured | stands | no | — |
| commandLineParser.ts:3501:17 | write | write | stands | no | — |
| commandLineParser.ts:3536:14 | literal init | literal init | stands | no | — |
| executeCommandLine.ts:730:19 | literal init | literal init | stands | no | — |
| executeCommandLine.ts:738:19 | literal init | literal init | stands | no | — |
| moduleNameResolver.ts:2303:47 | captured | captured | stands | no | — |
| moduleSpecifiers.ts:1122:172 | captured | captured | stands | no | — |
| sys.ts:1643:72 | literal init | literal init | stands | no | sys.ts:1642:24 |
| sys.ts:364:9 | write | write | stands | no | — |
| sys.ts:484:9 | captured | captured | stands | no | — |
| tracing.ts:305:21 | write | write | stands | no | — |
| tsbuildPublic.ts:2067:5 | write | write | stands | no | — |
| watchPublic.ts:1009:46 | literal init | literal init | stands | no | — |

The `sys.ts:1643:72` entry is the literal initialization of an `error` field from the catch binding at `sys.ts:1642:24`. It stands; it is not a captured catch-cell stop.

## Every observed site, with source line

### checker.ts:15346:21

```typescript
                    stack.push(identity);
```

### checker.ts:25287:55

```typescript
    function hasMatchingRecursionIdentity(type: Type, identity: object): boolean {
```

### commandLineParser.ts:2691:9

```typescript
        watchOptions: watchOptionMap && optionMapToObject(watchOptionMap),
```

### commandLineParser.ts:2694:9

```typescript
        ...(configParseResult.options.configFile?.configFileSpecs ? {
```

### commandLineParser.ts:3093:11

```typescript
    const raw = checkedAdamicJsonObject(parsedConfig.raw);
```

### commandLineParser.ts:3117:9

```typescript
        raw,
```

### commandLineParser.ts:3438:18

```typescript
        return { raw: json || convertToObject(sourceFile!, errors) };
```

### commandLineParser.ts:3445:11

```typescript
    const ownRaw = checkedAdamicJsonObject(ownConfig.raw);
```

### commandLineParser.ts:3468:73

```typescript
        if (ownRaw.compileOnSave === undefined && result.compileOnSave) ownRaw.compileOnSave = result.compileOnSave;
```

### commandLineParser.ts:3481:19

```typescript
            const extendsRaw = checkedAdamicJsonObject(extendedConfig.raw);
```

### commandLineParser.ts:3501:17

```typescript
                result.compileOnSave = extendsRaw.compileOnSave;
```

### commandLineParser.ts:3536:14

```typescript
    return { raw: json, options, watchOptions, typeAcquisition, extendedConfigPath };
```

### executeCommandLine.ts:730:19

```typescript
                { ...commandLine, options: commandLineOptions },
```

### executeCommandLine.ts:738:19

```typescript
                { ...commandLine, options: commandLineOptions },
```

### moduleNameResolver.ts:2303:47

```typescript
    function loadEntrypointsFromTargetExports(target: unknown): boolean | undefined {
```

### moduleSpecifiers.ts:1122:172

```typescript
function tryGetModuleNameFromExports(options: CompilerOptions, host: ModuleSpecifierResolutionHost, targetFilePath: string, packageDirectory: string, packageName: string, exports: unknown, conditions: string[]): { moduleFileToTry: string; } | undefined {
```

### sys.ts:1643:72

```typescript
                    return { module: undefined, modulePath: undefined, error };
```

### sys.ts:364:9

```typescript
        pollingIntervalQueue(pollingInterval).pollScheduled = (() => {
```

### sys.ts:484:9

```typescript
        pollScheduled = (() => {
```

### tracing.ts:305:21

```typescript
                    recursionIdentityMap.set(recursionIdentity, recursionToken);
```

### tsbuildPublic.ts:2067:5

```typescript
    state.timerToBuildInvalidatedProject = (() => {
```

### watchPublic.ts:1009:46

```typescript
                config.parsedCommandLine = { ...config.parsedCommandLine, fileNames };
```
