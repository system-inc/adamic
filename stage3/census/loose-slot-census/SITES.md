# Every observed type-level site

Locations refer to the adapted head source. Each entry quotes its full source line. Multiple events at one exact position are grouped.

## checker.ts:15346:21

1 event(s); forms: write.

```typescript
                    stack.push(identity);
```

## checker.ts:25287:55

1 event(s); forms: captured.

```typescript
    function hasMatchingRecursionIdentity(type: Type, identity: object): boolean {
```

## commandLineParser.ts:2691:9

1 event(s); forms: literal init.

```typescript
        watchOptions: watchOptionMap && optionMapToObject(watchOptionMap),
```

## commandLineParser.ts:2694:9

1 event(s); forms: literal init.

```typescript
        ...(configParseResult.options.configFile?.configFileSpecs ? {
```

## commandLineParser.ts:3093:11

1 event(s); forms: captured.

```typescript
    const raw = checkedAdamicJsonObject(parsedConfig.raw);
```

## commandLineParser.ts:3117:9

1 event(s); forms: literal init.

```typescript
        raw,
```

## commandLineParser.ts:3438:18

1 event(s); forms: literal init.

```typescript
        return { raw: json || convertToObject(sourceFile!, errors) };
```

## commandLineParser.ts:3445:11

1 event(s); forms: captured.

```typescript
    const ownRaw = checkedAdamicJsonObject(ownConfig.raw);
```

## commandLineParser.ts:3468:73

1 event(s); forms: write.

```typescript
        if (ownRaw.compileOnSave === undefined && result.compileOnSave) ownRaw.compileOnSave = result.compileOnSave;
```

## commandLineParser.ts:3481:19

1 event(s); forms: captured.

```typescript
            const extendsRaw = checkedAdamicJsonObject(extendedConfig.raw);
```

## commandLineParser.ts:3501:17

1 event(s); forms: write.

```typescript
                result.compileOnSave = extendsRaw.compileOnSave;
```

## commandLineParser.ts:3536:14

1 event(s); forms: literal init.

```typescript
    return { raw: json, options, watchOptions, typeAcquisition, extendedConfigPath };
```

## executeCommandLine.ts:730:19

1 event(s); forms: literal init.

```typescript
                { ...commandLine, options: commandLineOptions },
```

## executeCommandLine.ts:738:19

1 event(s); forms: literal init.

```typescript
                { ...commandLine, options: commandLineOptions },
```

## moduleNameResolver.ts:2303:47

1 event(s); forms: captured.

```typescript
    function loadEntrypointsFromTargetExports(target: unknown): boolean | undefined {
```

## moduleSpecifiers.ts:1122:172

1 event(s); forms: captured.

```typescript
function tryGetModuleNameFromExports(options: CompilerOptions, host: ModuleSpecifierResolutionHost, targetFilePath: string, packageDirectory: string, packageName: string, exports: unknown, conditions: string[]): { moduleFileToTry: string; } | undefined {
```

## sys.ts:364:9

1 event(s); forms: write.

```typescript
        pollingIntervalQueue(pollingInterval).pollScheduled = (() => {
```

## sys.ts:484:9

1 event(s); forms: captured.

```typescript
        pollScheduled = (() => {
```

## sys.ts:1643:72

1 event(s); forms: literal init.

```typescript
                    return { module: undefined, modulePath: undefined, error };
```

## tracing.ts:305:21

1 event(s); forms: write.

```typescript
                    recursionIdentityMap.set(recursionIdentity, recursionToken);
```

## tsbuildPublic.ts:2067:5

1 event(s); forms: write.

```typescript
    state.timerToBuildInvalidatedProject = (() => {
```

## watchPublic.ts:1009:46

1 event(s); forms: literal init.

```typescript
                config.parsedCommandLine = { ...config.parsedCommandLine, fileNames };
```
