# Observed type-level sites

This is a partial source/type measurement, not a completed lowering or an exhaustive production-stop count.

## checker.ts:15346:21

Forms: write; events: 1.

```typescript
                    stack.push(identity);
```

## checker.ts:25287:55

Forms: captured; events: 1.

```typescript
    function hasMatchingRecursionIdentity(type: Type, identity: object): boolean {
```

## commandLineParser.ts:2691:9

Forms: literal init; events: 1.

```typescript
        watchOptions: watchOptionMap && optionMapToObject(watchOptionMap),
```

## commandLineParser.ts:2694:9

Forms: literal init; events: 1.

```typescript
        ...(configParseResult.options.configFile?.configFileSpecs ? {
```

## commandLineParser.ts:3093:11

Forms: captured; events: 1.

```typescript
    const raw = checkedAdamicJsonObject(parsedConfig.raw);
```

## commandLineParser.ts:3117:9

Forms: literal init; events: 1.

```typescript
        raw,
```

## commandLineParser.ts:3438:18

Forms: literal init; events: 1.

```typescript
        return { raw: json || convertToObject(sourceFile!, errors) };
```

## commandLineParser.ts:3445:11

Forms: captured; events: 1.

```typescript
    const ownRaw = checkedAdamicJsonObject(ownConfig.raw);
```

## commandLineParser.ts:3468:73

Forms: write; events: 1.

```typescript
        if (ownRaw.compileOnSave === undefined && result.compileOnSave) ownRaw.compileOnSave = result.compileOnSave;
```

## commandLineParser.ts:3481:19

Forms: captured; events: 1.

```typescript
            const extendsRaw = checkedAdamicJsonObject(extendedConfig.raw);
```

## commandLineParser.ts:3501:17

Forms: write; events: 1.

```typescript
                result.compileOnSave = extendsRaw.compileOnSave;
```

## commandLineParser.ts:3536:14

Forms: literal init; events: 1.

```typescript
    return { raw: json, options, watchOptions, typeAcquisition, extendedConfigPath };
```

## executeCommandLine.ts:730:19

Forms: literal init; events: 1.

```typescript
                { ...commandLine, options: commandLineOptions },
```

## executeCommandLine.ts:738:19

Forms: literal init; events: 1.

```typescript
                { ...commandLine, options: commandLineOptions },
```

## moduleNameResolver.ts:2303:47

Forms: captured; events: 1.

```typescript
    function loadEntrypointsFromTargetExports(target: unknown): boolean | undefined {
```

## moduleSpecifiers.ts:1122:172

Forms: captured; events: 1.

```typescript
function tryGetModuleNameFromExports(options: CompilerOptions, host: ModuleSpecifierResolutionHost, targetFilePath: string, packageDirectory: string, packageName: string, exports: unknown, conditions: string[]): { moduleFileToTry: string; } | undefined {
```

## sys.ts:364:9

Forms: write; events: 1.

```typescript
        pollingIntervalQueue(pollingInterval).pollScheduled = (() => {
```

## sys.ts:484:9

Forms: captured; events: 1.

```typescript
        pollScheduled = (() => {
```

## sys.ts:1643:72

Forms: literal init; events: 1.

```typescript
                    return { module: undefined, modulePath: undefined, error };
```

## tracing.ts:305:21

Forms: write; events: 1.

```typescript
                    recursionIdentityMap.set(recursionIdentity, recursionToken);
```

## tsbuildPublic.ts:2067:5

Forms: write; events: 1.

```typescript
    state.timerToBuildInvalidatedProject = (() => {
```

## watchPublic.ts:1009:46

Forms: literal init; events: 1.

```typescript
                config.parsedCommandLine = { ...config.parsedCommandLine, fileNames };
```
