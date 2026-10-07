# Programs and project inputs

A program can have several roots:

```
adamic build third.a first.a second.a -o program
adamic build --project path/to/tsconfig.json --entry path/to/main.a -o program
adamic c third.a first.a second.a
adamic js third.a first.a second.a
```

`load.LoadProject` reads JSON with comments, `extends` (including package and array
forms supported by the pinned checker), `files`, `include`, `exclude`, and compiler
options through typescript-go's `tsoptions.GetParsedCommandLineOfConfigFile` shim.
No config diagnostics are suppressed. Project references are refused with a
not-yet message; this command builds one program, not a solution's build graph.
Declaration roots are checked but have no executable module body. An entirely
declaration-only program is refused by lowering.

## Roots and evaluation

Explicit CLI roots retain their input order and are execution entries.
Project roots are checking inputs only: they never become execution entries implicitly.
A project build requires exactly one `--entry`, resolved from the current working
directory, which must appear in the config's resolved `files/include` list.
An outside entry is refused with the fix: add it to `files/include` or choose
an already listed entry. Imported dependencies alone do not qualify as configured entries.
All configured files are checked, including unimported files and declaration roots;
only the selected entry's module graph is lowered. Unimported configuration roots
cannot print, allocate, or trigger lowering refusals. `load.LoadProject` checks only;
`load.LoadProjectEntry` additionally selects the runtime entry.
`Program.Files()` lists checking sources and `Program.Entries()` lists execution entries.

Project checking roots retain the upstream
parser's order: literal `files` first in listed order, then wildcard matches in
include-pattern order and the parser's directory/file sorting order. Normalized
paths are deduplicated at their first occurrence. Excludes do not remove literal
`files`, and do not exclude a dependency imported by a selected source.

Each root starts a depth-first post-order module walk. Imports are visited in
source order before the importing module's body; shared modules run once over
all roots. Native `main` and the JavaScript backend receive that same statement
order. The first executable root's basename labels generated output. The
three-root fixture compares both backends to Node importing those roots in order,
including a shared transitive dependency.

The import-cycle unit owns `internal/lower/modules.go:moduleOrder`. This change
leaves that function and its cycle policy intact. `rootOrder` calls its walk for
each root and deduplicates the resulting modules, so its cycle changes can merge
without editing the same function. Until that unit lands, import cycles remain
refused. This unit does not change type-only import traversal or other decisions
inside that walk.

At the census's pinned TypeScript v6.0.3 commit
`050880ce59e30b356b686bd3144efe24f875ebc8`,
[`src/tsc/tsc.ts`](https://github.com/microsoft/TypeScript/blob/050880ce59e30b356b686bd3144efe24f875ebc8/src/tsc/tsc.ts)
is the CLI entry: it imports `./_namespaces/ts.js`, installs the logging host, and
calls `ts.executeCommandLine(ts.sys, ts.noop, ts.sys.args)`.
[`src/tsc/tsconfig.json`](https://github.com/microsoft/TypeScript/blob/050880ce59e30b356b686bd3144efe24f875ebc8/src/tsc/tsconfig.json)
includes its own directory and references the compiler project. The compiler
project's 77 roots describe the compiler library, not the command-line entry.
The pinned [Herebyfile.mjs](https://github.com/microsoft/TypeScript/blob/050880ce59e30b356b686bd3144efe24f875ebc8/Herebyfile.mjs)
confirms this: its `tsc` task passes `srcEntrypoint: "./src/tsc/tsc.ts"`
to esbuild as the single `entryPoints` item, with `bundle: true` and CommonJS
output. It enables a compile-cache shim for `built/local/tsc.js`; without
bundling, the shim requires `built/local/tsc/tsc.js`. Its separate project
builder checks project references with `tsc -b`.

TypeScript's config root order determines checking/emission inputs; it does not
make every emitted module a runtime entry. To reproduce the CLI's runtime entry,
use `adamic build --project src/tsc/tsconfig.json --entry src/tsc/tsc.ts -o tsc`.
The config still controls checking; its file order has no effect on runtime entry selection. This unit does not implement TypeScript's bundling/build scripts or
project-reference graph, and does not claim to build native tsc.

## Compiler option contract

An omitted required option inherits Adamic's value. After `extends` is resolved,
an explicit incompatible value is refused, naming the option and saying
"Adamic requires". Adamic does not silently override an explicit weakening.
Even with `strict: true`, an explicit false strict sub-option is refused.

Every option set by the direct loader has this project rule:

| Option | Project rule |
| --- | --- |
| `strict` | Must be true; omitted becomes true. |
| `noUncheckedIndexedAccess` | Must be true; omitted becomes true. |
| `exactOptionalPropertyTypes` | Must be true; omitted becomes true. |
| `noImplicitReturns` | Must be true; omitted becomes true. |
| `noFallthroughCasesInSwitch` | Must be true; omitted becomes true. |
| `erasableSyntaxOnly` | Must be true; omitted becomes true. |
| `verbatimModuleSyntax` | Must be true; omitted becomes true. |
| `allowImportingTsExtensions` | Must be true; omitted becomes true. |
| `noEmit` | Must be true; omitted becomes true. The checker does not emit; Adamic's native build still does. |
| `module` | Must be `esnext`; omitted becomes `esnext`. CommonJS/NodeNext transformation semantics are not this backend's contract. |
| `moduleDetection` | Must be `force`; omitted becomes `force`. Every executable root is a module. |
| `moduleResolution` | Must be `bundler`; omitted becomes `bundler`. Keep the supported source module resolver. |
| `target` | Must be `es2024` or later; omitted becomes `es2024`. |
| `lib` | Project value is retained; omitted becomes `["es2024"]`. The bundled regex soundness adapter remains active for every selected library. |
| `types` | Project value is retained, including `[]`; omitted retains the checker's automatic ambient-type discovery, including `typeRoots`. Direct file inputs still use `[]`. |

The following strict sub-options must also be true or omitted:
`noImplicitAny`, `noImplicitThis`, `strictNullChecks`, `strictFunctionTypes`,
`strictBindCallApply`, `strictPropertyInitialization`,
`strictBuiltinIteratorReturn`, `useUnknownInCatchVariables`, and `alwaysStrict`.

Additional guards:

| Option | Project rule |
| --- | --- |
| `noCheck` | Must be false or omitted: checking is mandatory. |
| `skipLibCheck`, `skipDefaultLibCheck` | Must be false or omitted: conflicting or invalid ambient declarations cannot be hidden. |
| `noResolve`, `noLib` | Must be false or omitted: do not remove dependencies or standard types from checking. |
| `libReplacement` | Must be false or omitted: preserve the bundled declarations and their soundness adapter. |
| `allowJs`, `checkJs` | Project values are retained; when JavaScript dependencies are enabled, `checkJs` must be explicitly true. JavaScript root inputs remain unsupported. |
| `useDefineForClassFields` | Cannot be false: preserve ECMAScript class-field semantics. Omitted uses the checker's target-derived value. |

All remaining options are retained and validated by typescript-go. For example,
`noUnusedLocals`, `noUnusedParameters`, `noImplicitOverride`,
`noPropertyAccessFromIndexSignature`, and `noUncheckedSideEffectImports` can add
checking without weakening the required baseline. Output/declaration/source-map
settings remain checker inputs; they do not configure Adamic's C backend. The
ordinary refusal and not-yet passes still reject unsupported language/runtime
features. External declarations do not implement their host APIs natively.

The unmodified upstream compiler config is deliberately refused: its inherited
`strictBindCallApply: false`, `useUnknownInCatchVariables: false`,
`skipLibCheck: true`, ES2020 target, and NodeNext module settings are incompatible
with this contract. A source adaptation must satisfy Adamic's options; the census
overlay's upstream-options run is not permission to weaken them.

## Prelude and host types

The prelude is still checked alongside a project's selected `lib` and `types`.
It declares the ambient `adamic` module (`panic`, file/input APIs, UTF-8 APIs,
`Weak`, and the opt-in checker bridge), Set-operation augmentations, collection
iterator result contracts, and the sound `JSON.stringify` result
`string | undefined`. Built-in regex declarations remain corrected independently
by `regexpLibraryFS`. Removing the entire prelude loses both runtime-symbol
recognition and these type contracts; it is not the production policy.

When declarations loaded from project roots, libraries, or ambient types actually
supply a global console, the loader omits only Adamic's restricted global console
declaration. It discovers these declarations before checking and uses a fresh
program/cache for the final prelude choice. Its other
prelude declarations keep the same identity and remain present. The host must
supply its own console if used; declaration conflicts remain checker errors.
Node's arbitrary-argument console is not lowered as Adamic's one-string console:
its calls currently produce an explicit lowering not-yet, as do unsupported Node
APIs. This avoids claiming that Node declarations supply a native implementation.
A project whose host declarations supply no global console retains Adamic's restricted console.

The census removed the whole prelude to study the upstream source with Node's
ambient declarations. That is valid as a labeled checker experiment. It does
not establish a sound executable Adamic program and must not become a way to
bypass the standard-library corrections.

## Adamic file names

New Adamic sources use `.a`. The config host exposes `.a.ts` aliases to the same
upstream parser that resolves imports. File specifications ending in `.a` in
`files`, `include`, or `exclude` are adapted using the upstream JSON syntax tree;
config inheritance and glob matching stay upstream. Directory globs such as
`src/**/*` see the aliases without changes. Diagnostics and the public program
file list use the original `.a` names. A real `.a.ts` alongside its `.a` is
refused rather than silently selecting another source.
