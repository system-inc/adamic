# 99g evidence and the post-99g D036 decline

These artifacts were moved from stage3/adapt/99-correctness-axolotl-g/proofs without changing their bytes. Keeping them here avoids putting proof outputs into apply's adaptation input key. The runnable adapter verification scripts remain in the adaptation directory, with their README commands.

## D036 still has an undefined-result path

The proposed assertion is on the result of the scalar `formatSingleValue(value, map)` call, not on its input value. 99g removes the own-undefined option path, but does not guarantee that JSON.stringify of a defined object returns a string.

The [Node witness](proofs/d036-post99g.cjs) runs the actual built 99g library. It supplies a paths object with a non-enumerable toJSON hook returning undefined:

```js
const paths = { '@app/*': ['./src/*'] };
Object.defineProperty(paths, 'toJSON', { value() { return undefined; } });
ts.generateTSConfig({ paths }, '\n');
```

This is an exported internal API invocation, not a command-line argument. No stock command-line input creates this hook. The option value itself is defined, so 99g does not select its literal-null arm. The remaining-options loop passes that same object as defaultValue; the nullish selection retains it. paths is an object option, with no enum map. formatValueOrArray therefore takes its scalar branch, formatSingleValue calls JSON.stringify on that defined object, the hook returns undefined, and the scalar call's result is undefined. The generated line is `"paths": undefined,`.

The witness temporarily observes JSON.stringify's argument, result and stack without changing its return value. The saved [result](proofs/d036-post99g.json) proves the call chain is JSON.stringify, formatSingleValue, formatValueOrArray, emitOption, generateTSConfig. The paths hook runs once. A second case supplies a PluginImport object with a toJSON hook in the plugins array; it observes the callback returning undefined and the array wrapper emitting `"plugins": [],`. This records the separate array callback path as well. The source is checked for 99g's exact formatter declaration and the still-unasserted scalar return. A `!` at that result would introduce a new Adamic failure on this API invocation.

Run from the repository root with an existing built 99g tree:

```sh
source /workspace/adamic-tools/env.sh
node stage3/evidence/99-correctness-axolotl-g/proofs/d036-post99g.cjs /workspace/scratch/99g/after /workspace/scratch/99g-followup/d036-post99g.json > /workspace/scratch/99g-followup/d036-post99g.log 2>&1
```

## Review of every producer path

Coordinates refer to /workspace/scratch/99g/after/src/compiler/commandLineParser.ts, built from stock TypeScript 6.0.3 plus main and 99g. The generator starts at 2838. The input is read once at 2938, and 99g selects the literal null or formatting at 2939. The scalar return is at 2958 and JSON.stringify at 2966.

| Producer | Absent option or ordinary command-line value | Limit of the invariant |
| --- | --- | --- |
| rootDir and outDir fixed calls | Defaults are the strings ./src and ./dist. JSON.stringify returns strings. | A set API value still follows the generic path; no new API restriction is added. |
| module, target, jsx and moduleDetection fixed calls | Defaults are NodeNext, ESNext, ReactJSX and Force. Their enum maps resolve them to string names; failed resolution throws before JSON.stringify. | Cleared own values use 99g's literal null and skip formatting. |
| types fixed call | Its absent default is an empty array. The array wrapper returns a string and invokes no callback. | API-provided array elements need their own serialization invariant. |
| lib fixed call | It has no absent-option call: options.lib guards it and is also its default. Command-line list parsing produces defined string/number entries. | An external API can supply other object behavior; the PresetValue cast does not validate it. |
| Sixteen boolean fixed calls | sourceMap, declaration, declarationMap, noUncheckedIndexedAccess, exactOptionalPropertyTypes, noImplicitReturns, noImplicitOverride, noUnusedLocals, noUnusedParameters, noFallthroughCasesInSwitch, noPropertyAccessFromIndexSignature, strict, verbatimModuleSyntax, isolatedModules, noUncheckedSideEffectImports and skipLibCheck all default to true. | Each absent default serializes to "true". Own undefined values select literal null instead. |
| Remaining-options loop, 2905-2907 | Object.keys supplies own keys, and defaultValue is the same own value. Own undefined selects literal null; primitives serialize to strings, and null serializes to "null". | Defined object values can have toJSON hooks returning undefined. The paths witness reaches the scalar result at 2958 absent. |
| Array callback, 2955 | parseListTypeOption:1884-1911 uses mapDefined, dropping failed conversions. Its scalar producers are strings, numbers or enum entries. Empty arrays and holes do not invoke a callback. Its array wrapper always returns a string. | generateTSConfig's exported internal API also accepts object-valued CompilerOptionsValue arrays, including PluginImport[]. A toJSON hook on such an element can make formatSingleValue's callback result undefined; .join then consumes that undefined. This is distinct from the scalar branch counterexample above. |

Both in-tree generateTSConfig callers were inspected. executeCommandLine.ts:1300 receives parsed command-line options from writeConfigFile after the error check at 577-584. The init unit-test helper at src/testRunner/unittests/config/initializeTSConfig.ts:8 likewise passes parseCommandLine's options. Their stock primitive/list inputs do not create serialization hooks. However, generateTSConfig remains exported at runtime and accepts CompilerOptions objects; neither 99g nor its existing PresetValue cast rejects a defined object with a serialization hook. The requested unconditional presence assertion therefore remains declined.

99b is unchanged. No 112-to-111 measurement is claimed: the assertion was not added. The retained 99g whole-program proof is 112 to 112 with no new diagnostics.
