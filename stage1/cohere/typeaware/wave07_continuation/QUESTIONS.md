# Continuation checker facts

These three ABI-v1 inspection questions have isolated Go files in
`bridge/tsgo/checker` and corresponding `.a` decoders here. Their schema is 1,
using the existing length-prefixed UTF-16 fields and owned UTF-8 C buffers.
They return no finding, message, repair or rule-event classification.

`wave07-symbol-context` exposes direct symbol identity, shorthand value identity,
aliased identity, and both direct and aliased flags/declarations. Each declaration
contains source path, syntax kind and byte span, declaration/default-library/module
flags, a textual name where the AST defines one, node/function flags, optional
body kind/span/text, and its complete ordered ancestry. An ancestor retains kind,
textual name, name kind, flags and global-augmentation flag. A call-like node also
exposes resolved signature presence, return flags and its exact selected
declaration. Symbol identities are borrowed from the live program; copied syntax
and strings do not keep a program alive. A nontextual binding/computed name is
not passed to Node.Text.

`wave07-control-flow` exposes structural blocks: reachability, zero-based
successors and every expression/statement syntax hook's kind and byte span.
The decoder maps these spans into the native parser, rejects empty graphs and
out-of-range edges, and retains duplicate events from finally copies. The builder
is a separately attributed copy of the pinned generic cohere flow-shape utility;
it performs no rule analysis. Adamic recognizes process writes and exits, tracks
try-chain state, follows callees and walks paths. The native verdict does not call
Go cohere. Its independent production finding oracle imports no bridge code.

`wave07-program-modules` exposes all program source paths, declaration flags,
nondeclaration source text and static/call import syntax records. Each record has
kind, type-only/import-call/require-call/literal flags and the compiler's resolved
source target, if present. Adamic decides whether to follow it, computes incoming
imports and the transitive StandardStreams closure, follows blocking calls and
selects process entries. This question does not answer whether any file blocks.

## Integration boundary

The repository's shared dispatcher is unchanged by this continuation. The private
validation archives use a scratch Go overlay that inserts only these routes:

```go
case "wave07-symbol-context": return p.wave07SymbolContext(c, node, question)
case "wave07-control-flow": return p.wave07ControlFlow(node, question)
case "wave07-program-modules": return p.wave07ProgramModules(node, question)
```

These routes still need integration into `bridge/tsgo/checker/facts.go` under
Ahra's shared-file workflow. The normal archive refuses an unknown question;
there is no silent fallback. The private-overlay comparisons prove the isolated
implementations, not that the production dispatcher already registers them.
The shared registration generator and existing test harness were not edited.
