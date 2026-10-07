# declaration-file-flags

An exact Identifier accepts `declaration-file-flags`, with no suffix. The answer
uses the existing version 1 UTF-16-length field framing:

1. Version, `1`.
2. Question, `declaration-file-flags`.
3. Whether `GetSymbolAtLocation` returned a symbol.
4. For a present symbol, its declaration count followed by two booleans for each
   declaration, in checker order: whether it has a source file, and that file's
   `IsDeclarationFile` flag. A missing file has both flags false.

A present symbol with zero declarations is distinct from a missing symbol. No
symbol identity, pointer, type, lint verdict or edit is returned. The C-owned
answer and borrowed checker handle retain the existing ABI ownership contract.
`DeclarationFileFlags` in `stage1/cohere/typeaware/declaration_file_flags.a`
validates and owns the decoded arrays. The radix rule decides which flags matter.

The direct checker test covers library and local parseInt symbols, undefined,
unresolved names, suffix rejection and wrong node-kind rejection. Inverting the
file flag compiles and fails the independent direct-checker comparison. The wave
suite queries this question after program release and requires panic 70; a
registry mutant retaining the program exits 0 and is caught by that expectation.
