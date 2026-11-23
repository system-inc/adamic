## 6. Init emits an undefined token for an explicitly cleared option

Status: fixed in our fork by adaptation 99g; draft issue, not filed.
Upstream: TypeScript 6.0.3, commit 050880ce59e30b356b686bd3144efe24f875ebc8.
Locations: src/compiler/commandLineParser.ts parseOptionValue:2111 stores
undefined for a null token; generateTSConfig:2838 retains the own key;
emitOption:2933 selects options[setting] ?? defaultValue; formatSingleValue:2960
returns JSON.stringify's undefined result. These are stock source coordinates.

Node counterexample: `tsc --init --outFile null` exits successfully but writes
`"outFile": undefined,` in tsconfig.json. The token undefined is not valid
JSON config syntax, so a subsequent invocation cannot read that generated file
cleanly. The user explicitly cleared this option; retaining it as null both
preserves that instruction and produces a valid config value.

The pinned Go compiler at system-inc/TypeScript
d92d9bfee114c80be2c375d72edae966176e3a4f writes `"outFile": null,` for the same
command. It also writes null for the fixed strict setting. Its Go parser
stores nil and its config generator serializes that value as JSON null. The
pinned Go compiler instead panics for a cleared target, because it attempts
an enum lookup for nil; our fork follows the ruled null-output behavior for
that setting as well, rather than copying the panic.

Our patch: stage3/adapt/99-correctness-axolotl-g/adapt.cjs changes only
emitOption's formatting. Read options[setting] once; if it is an own undefined
value, format it as the literal null. Otherwise use exactly the existing
nullish default selection and formatting. The active/commented form, option
ordering and absent-option defaults are preserved. This intentionally changes
JavaScript behavior only for own undefined option values. No assertion is
added, no public API is changed, and no test baseline is rewritten.

The executable before/after CLI witnesses, unchanged-output controls, Go
comparisons and original-function mutant are in
[99g's witness](../adapt/99-correctness-axolotl-g/witness.cjs). The adapter README
records the full proof and oracle results. The earlier upstream snapshot at
Microsoft/TypeScript main 6161772c1370bdd4a82f808e6823bd755eb1870d has an active
Go init.go identical to the pinned version and retains a TypeScript fixture
with generateTSConfig unchanged from 6.0.3.

### Draft GitHub issue

Title: tsc --init --outFile null writes an undefined value into tsconfig.json

Version: TypeScript 6.0.3, Node 24.19.0, Linux x64.

Minimal reproduction in a fresh directory, with TypeScript 6.0.3 installed:

```sh
printf 'export {};\n' > index.ts
npx tsc --init --outFile null
npx tsc --showConfig
```

Expected: init writes `"outFile": null,`, preserving the explicit clearing,
and the generated file reads back successfully. This matches the native Go
compiler for outFile and the fixed strict setting.

Actual: init exits 0 but emits `"outFile": undefined,`. The next command exits
1 and reports:

```text
tsconfig.json(44,16): error TS5024: Compiler option 'outFile' requires a value of type string.
```

Suggested upstream change: in generateTSConfig's emitOption, distinguish an
own undefined value from an absent option. Emit the literal null for the
former in the same active or commented form already selected for that option.
Preserve the existing nullish fallback and formatting for every other value.
This must also apply to preset options such as strict; only changing the
remaining-options loop would leave those cleared settings using their defaults.
