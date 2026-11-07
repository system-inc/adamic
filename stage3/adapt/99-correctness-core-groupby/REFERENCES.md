# Adapted source references

The TypeScript checker resolves identifiers and import aliases to the two core
function symbols across all 711 adapted src TypeScript roots. references.json
retains every matching identifier and every resolved core-function reference.
The raw rg search, including comments, local variables called clone, library
feature-name strings groupBy, and the unrelated typesMap.json package name, is
retained in evidence/all-name-hits.log. Those are distinguished from references
to these core functions. Star namespace exports expose the same symbols.

## groupBy

| File | Line | Use |
| --- | ---: | --- |
| src/compiler/core.ts | 1471 | declaration |
| src/compiler/core.ts | 1473 | declaration |
| src/compiler/core.ts | 1474 | declaration |
| src/compiler/transformers/legacyDecorators.ts | 31 | import |
| src/compiler/transformers/legacyDecorators.ts | 538 | call |
| src/services/organizeImports.ts | 25 | import |
| src/services/organizeImports.ts | 427 | call |

## clone

| File | Line | Use |
| --- | ---: | --- |
| src/compiler/core.ts | 1488 | declaration |
| src/services/utilities.ts | 21 | import |
| src/services/utilities.ts | 2245 | call |
| src/harness/fourslashImpl.ts | 2914 | call |
| src/harness/fourslashImpl.ts | 3022 | call |
| src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts | 20 | import |
| src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts | 383 | call |
| src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts | 386 | callback argument |

The compiler's only groupBy call, legacyDecorators.ts:538, resolves to the
key-selector overload with K = boolean. isSyntheticMetadataDecorator returns
boolean, not a type predicate. The false and true optional arrays are destructured
as decorators and metadata. The caller maps ordinary decorators, appends mapped
parameter decorators, and maps metadata last. Undefined buckets are accepted by
map. The services call at organizeImports.ts:427 uses the same overload with
string keys. It iterates groups by import attributes before categorizing and
coalescing imports.

clone has no compiler caller. cloneCompilerOptions in services makes a shallow
copy and restores the configFile option. The harness copies format options.
The virtual file-system helper clones entries directly and passes clone as a
map callback for folder entries. There is no server reference to either core
function; server/typesMap.json's clone string is a package mapping.

Both functions occur only in typescript.internal.d.ts, not typescript.d.ts or
tsserverlibrary.d.ts. All three declaration files are byte-identical before
and after either adaptation. Exact definition lines and hashes are retained
in evidence/declarations.json.
