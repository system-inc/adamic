Adamic accepts this JavaScript-changing adaptation after its review and census (#bbssdcy, the tail pass's family 57, rulings.md:328 and :193).

The two private config-dir option tables keep their truthiness behavior. The adapter pins the exact CRLF text of the two top-level variable statements, located by their declaration names and parsed callback nodes. It checks both statements before any write and refuses mixed states.

Before

```ts
const configDirTemplateSubstitutionOptions: readonly CommandLineOption[] = optionDeclarations.filter(
    option => option.allowConfigDirTemplateSubstitution || (!option.isCommandLineOnly && option.isFilePath),
);
```

After

```ts
const configDirTemplateSubstitutionOptions: readonly CommandLineOption[] = optionDeclarations.filter(
    option => (option.allowConfigDirTemplateSubstitution || (!option.isCommandLineOnly && option.isFilePath)) ? true : false,
);
```

Before

```ts
const configDirTemplateSubstitutionWatchOptions: readonly CommandLineOption[] = optionsForWatch.filter(
    option => option.allowConfigDirTemplateSubstitution || (!option.isCommandLineOnly && option.isFilePath),
);
```

After

```ts
const configDirTemplateSubstitutionWatchOptions: readonly CommandLineOption[] = optionsForWatch.filter(
    option => (option.allowConfigDirTemplateSubstitution || (!option.isCommandLineOnly && option.isFilePath)) ? true : false,
);
```

## Option producers

The adapted `commandLineParser.ts:1693` table spreads the literal arrays in `:364` and `:638`. The literal watch table is at `:289`. AST inspection followed every spread and direct variable reference. The 129 compiler entries and six watch entries are object literals. Only the following assignments set the three fields on a table entry. Values are `true` except `out.isFilePath`, which is `false`. There are no accessors or computed producers. An `isFilePath` assignment in an option's nested `element` is outside the entry-level field these callbacks read.

| Table source | Field | Entry names and lines in `commandLineParser.ts` |
|---|---|---|
| `commonOptionsWithBuild` | `isCommandLineOnly` | `help` 370 and 379, `watch` 388, `locale` 562 |
| `commonOptionsWithBuild` | `isFilePath` | `generateCpuProfile` 454, `generateTrace` 463 |
| `commandOptionsWithoutBuild` | `isCommandLineOnly` | `showConfig` 680, `listFilesOnly` 688, `ignoreConfig` 697 |
| `commandOptionsWithoutBuild` | `isFilePath` | `project` 669, `outFile` 763, `outDir` 776, `rootDir` 788, `tsBuildInfoFile` 810, `baseUrl` 1120, `out` 1385, `declarationDir` 1554 |
| `commandOptionsWithoutBuild` | `allowConfigDirTemplateSubstitution` | `paths` 1130, `rootDirs` 1148, `typeRoots` 1163 |
| `optionsForWatch` | `allowConfigDirTemplateSubstitution` | `excludeDirectories` 344, `excludeFiles` 357 |

The pinned npm package exposes `optionDeclarations` and `optionsForWatch` at runtime. The test uses these objects and runs both snapshot arrows in a VM, checking identity and order as well as all 64 field combinations per arrow.

## Rulings
- **Conditions use JavaScript ToBoolean in .ts and .a (0, -0, NaN, '', null and undefined are false; represented objects and empty arrays are true), in `if`, loops, `?:` and `!`; optional numeric truthiness is direct ToBoolean; preferring explicit comparisons is taste for cohere:house; whether .a should require boolean conditions is undecided.** Accept. Oct 7 to 8 2026, @system_adamic; source: #bw3xg7c (5wqmh5cj), #yvgst44 (nz497e4d, item 8) / docs/verification/conditions/REPORT.md:10,24-26. Spec: What's in, Values and statements (conditions); The soundness rules (the stage 3 extension paragraph). Test: internal/oracle/conditions_test.go:25 TestTypeScriptConditionsAgreeWithNode.
- **`reduce` without an initial value, `sort` without a comparator, a `filter` callback that doesn't return a boolean and an `Array.from` first parameter typed `number` are refused.** Refuse. Oct 5 2026, Kirk; source: #8vpywxp. Spec: The standard library for 0.1 (Array: `reduce` with an initial value, `sort(compare)`); Semantics that differ from JavaScript (sort). Test: internal/lower/lower_test.go:135 (reduce); :120 (filter); :121 (Array.from); sort internal/conformance/testdata/untested/sort_without_comparator.a, beside sort_with_comparator.a, which lowers (Oct 22).
