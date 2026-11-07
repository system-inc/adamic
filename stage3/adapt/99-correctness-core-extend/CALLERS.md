# Caller inventory on the fully adapted main tree

The checker resolved every call in all compiler, services and server sources to
its declaration, rather than matching names alone. There are four calls, all in
compiler. No services or server caller exists. The independent wider source
search found one additional harness call, included below and in the lane count.
The source files themselves are unchanged by this adaptation.

| Caller, adapted line | Argument | Declared type | Type at call | Shape in the type system |
| --- | --- | --- | --- | --- |
| commandLineParser.ts:3095, extend | existingOptions | CompilerOptions, default {} | CompilerOptions | String index signature |
| same | parsedConfig.options or {} | Property: CompilerOptions or undefined | CompilerOptions | String index signature |
| commandLineParser.ts:3101, extend | existingWatchOptions | WatchOptions or undefined | WatchOptions | String index signature |
| same | parsedConfig.watchOptions | WatchOptions or undefined | WatchOptions | String index signature |
| watch.ts:873, copyProperties | result | WatchCompilerHost<T> | WatchCompilerHost<T> | Fixed interface |
| same | createWatchHost(...) | Return: WatchHost | WatchHost | Fixed interface |
| tsbuildPublic.ts:326, copyProperties | host | SolutionBuilderWithWatchHost<T> | SolutionBuilderWithWatchHost<T> | Fixed interface |
| same | watchHost | Inferred WatchHost | WatchHost | Fixed interface |
| harness/fourslashImpl.ts:352, extend | tsConfig.options | CompilerOptions | CompilerOptions | String index signature |
| same | compilationOptions | CompilerOptions | CompilerOptions | String index signature |

CompilerOptions has a string index of CompilerOptionsValue, TsConfigSourceFile,
or undefined. WatchOptions has a string index of CompilerOptionsValue or
undefined. Both satisfy the record signature. The two target host interfaces
extend WatchHost and have no string index signature. Giving source and target
the same WatchHost view supports the generic key write without adding a record
index to these fixed interfaces.

The full checker inventory retains declaration text, index types and resolved
call-site types in the evidence archive. Preliminary stock-checker experiments
preceded the full checks; their ambient-node diagnostics are retained. The
ordinary Adamic checks of both complete applied trees are the authoritative
79-root checks, each with zero diagnostics.
