# 98-e's declared typescript.d.ts lines

Built by stock `npm run build` in the tree at 98-c (before) and with 98-e (after).
typescript.d.ts.diff is the whole diff; nothing else moves.

| Owner path (ownerPath(), pinned line) | Kind before | Before | After |
| --- | --- | --- | --- |
| `["server", "ModuleImportResult", "error"]`, 2707 | PropertySignature | `error: { stack?: string; message?: string; };` | `error: unknown;` |
| `["System", "createSHA256Hash"]`, 8487 | MethodSignature | `createSHA256Hash?(data: string): string;` | `createSHA256Hash?: { createSHA256Hash(data: string): string; }["createSHA256Hash"] \| undefined;` |

normalized-changes.json is stage3/lane/normalize-api.cjs's changes() on the two built
files: exactly `TypeAliasDeclaration:["ts","server","ModuleImportResult"]` and
`member:["ts","System","createSHA256Hash"]`.

tsserverlibrary.d.ts.diff and tsserverlibrary.internal.d.ts.diff are empty. Neither public
file names propertyOf, errorCode or errorMessage. typescript.internal.d.ts.diff adds the
three hostErrors.ts functions as @internal declarations, and the internal ModuleImportResult.

The first barrel form, a named `export { propertyOf }`, failed dts-services with
"Failed to find namespace for propertyOf at built/local/compiler/_namespaces/ts.d.ts:9:26"
in scripts/dtsBundler.mjs; that run's tree was deleted and rebuilt from a fresh apply.
