# Family 97 fixed dependency keys

Compiler accepts this adaptation under the split in #bbssdcy (Nov 22,
build22-97). It applies to the call bound at the family 97 finding at
`src/compiler/utilities.ts:9043:26` on the base tree. The edits replace the key alias with fixed dependency metadata and use it
for the type parameter constraint and mapped keys in
`createComputedCompilerOptions`. The binding judgment in
`docs/rulings.md:137` still applies.

The fixed keys are `allowImportingTsExtensions`, `target`, `module`,
`moduleResolution`, `moduleDetection`, `isolatedModules`, `esModuleInterop`,
`allowSyntheticDefaultImports`, `resolvePackageJsonExports`,
`resolvePackageJsonImports`, `resolveJsonModule`, `declaration`,
`preserveConstEnums`, `incremental`, `declarationMap`, `allowJs`,
`useDefineForClassFields`, `noImplicitAny`, `noImplicitThis`,
`strictNullChecks`, `strictFunctionTypes`, `strictBindCallApply`,
`strictPropertyInitialization`, `strictBuiltinIteratorReturn`, `alwaysStrict`,
and `useUnknownInCatchVariables`.

The mapped parameter still gives each `dependencies` array the type `T[K]`
and each callback `Pick<CompilerOptions, K | T[K][number]>`. Stock TypeScript 6.0.3 types the `module` entry before the edit as
`dependencies: "target"[]` and
`computeValue: (compilerOptions: Pick<CompilerOptions, "module" | "target">) => ModuleKind`.
After the edit the checker gives `dependencies: "target"[]` and
`computeValue: (compilerOptions: Pick<CompilerOptions, "target" | "module">) => ModuleKind`.
The checker prints the two union members in reverse order. The callback still
accepts a view with the same two fields.
The type alias, constraint and mapped key domain are erased by TypeScript. The option table, dependency literals and
callback bodies are retained byte for byte.
