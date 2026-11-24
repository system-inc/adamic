# Split the watch-options assignment out of the call

Family 150, Nov 22 meter, `commandLineParser.ts:1994:117` under the
`parseCommandLineWorker` unit reported at `commandLineParser.ts:1959:1`.
Compiler accepts under the Nov 22 split (#bbssdcy, about 22:58Z). Inside
`if (watchOpt)`, precede `parseOptionValue` with
`if (watchOptions === undefined) watchOptions = {};` and pass `watchOptions`.

The callee and the four earlier arguments are plain binding reads. These are
their declarations in the TypeScript 6.0.3 snapshot.

```ts
function parseOptionValue(
    args: readonly string[],
    i: number,
    diagnostics: ParseCommandLineWorkerDiagnostics,
    opt: CommandLineOption,
    options: OptionsBase,
    errors: Diagnostic[],
) {
```

```ts
    function parseStrings(args: readonly string[]) {
        let i = 0;
```

```ts
const watchOptionsDidYouMeanDiagnostics: ParseCommandLineWorkerDiagnostics = {
    getOptionsNameMap: getWatchOptionsNameMap,
    optionDeclarations: optionsForWatch,
    unknownOptionDiagnostic: Diagnostics.Unknown_watch_option_0,
    unknownDidYouMeanDiagnostic: Diagnostics.Unknown_watch_option_0_Did_you_mean_1,
    optionTypeMismatchDiagnostic: Diagnostics.Watch_option_0_requires_a_value_of_type_1,
};
```

```ts
                    const watchOpt = getOptionDeclarationFromName(watchOptionsDidYouMeanDiagnostics.getOptionsNameMap, inputOptionName, /*allowShort*/ true);
```

The initializer of `watchOpt` has already run. Reading the callee or any of
these arguments invokes neither a getter nor a call. The guard and store can
precede those reads. The read of `errors` still follows initialization.

```ts
    let watchOptions: WatchOptions | undefined;
```

This local holds only an object or `undefined`. Every object is truthy, so
`||` and `=== undefined` choose the same case for every permitted value.
The missing case makes one fresh record. Both cases pass the same mutable
record held by the local.

[Rows 72, 95, 138 and 162 of the rulings](../../../docs/rulings.md) govern the
write and the argument. Row 162 accounts for the original expression's value.
Under row 72's Nov 13 amendment, the handle is judged through its declared
type. With the split the write is checked on its own and the argument must
independently fit the parameter. Rows 95 and 138 still apply to mutable
locations and constrained relations. Any refusal for an incompatible handle
at the call stands.

## Snapshot guards and test

Before either path can proceed, the adapter pins the uninitialized
`watchOptions` declaration in the worker and the required `OptionsBase`
parameter on `parseOptionValue`. It checks the five declarations quoted
above and rejects declarations that could shadow those bindings at the
branch. The same checks run on first application and on an already-adapted
file, before any write.

Run `node test.cjs <applied-tree> [adapter]` from this directory after a
complete Stage 3 apply. The optional adapter path defaults to `adapt.cjs`.
The test reads this adapter's before snapshot and compares its clean output
with its after snapshot. Earlier or later adapters that edit
`commandLineParser.ts` leave this comparison valid.

The test checks repeat application. On
both first and repeat paths, it checks rejection and byte preservation
for the falsy declaration, the complete typed drift, and the parameter-only
drift. Its binding plants add a destructured callee shadow, add a block
`args` shadow, change the `i` initializer to 1, add a function-scoped
diagnostics shadow in the preceding branch, and replace the `watchOpt`
initializer's callee.

The behavior test transpiles the original snapshot and the adapter's own
output with stock TypeScript 6.0.3. For each side it replaces the worker in
the pinned TypeScript bundle and calls `parseCommandLine`. It checks the
original against stock, and the split against both the original and expected
results for no watch option, a valid option, a second option from a response
file, and an invalid value with diagnostic 6046. These checks run even when
the separate after-snapshot comparison fails.
