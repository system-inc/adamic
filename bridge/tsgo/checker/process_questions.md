# Syntax and process provenance questions

These version-1 questions use the existing UTF-16 length-prefixed field framing.
Selectors still use Go AST byte positions and exact syntax kinds. A released
program is rejected by the existing C registry before a question is dispatched.
The only registration edit is the fallback in wave 25's own
`type_declaration_ancestry.go`; shared `facts.go` is unchanged in this batch.

No question recognizes a lint rule, process exit, console write, timeout,
StandardStreams filename, import closure or diagnostic. Adamic performs those
classifications and the path-state analysis.

* `syntax-projection` selects SourceFile. Fields are version, question, filename,
  declaration-file flag, external-module flag, node count, then preorder nodes.
  Dense identities are one-based; zero means absent. Each node carries kind,
  byte pos/end/token start, parent identity, supported literal/name text, node
  flags, function flags, code-path root identity, root flag, global-augmentation
  flag, name/body/initializer/expression identities, binary/prefix operator kind,
  child identities and CallExpression/NewExpression argument identities.
  Each identity list is prefixed by its count. Unsupported text and accessor
  kinds are represented as empty text or zero, never invoked speculatively.
* `syntax-control-flow` selects a code-path root. Fields are version, question,
  block count, then blocks: reachable flag, counted successor indexes, event
  count and pairs of hook kind and syntax identity. Block indexes are zero-based.
  Hook zero is expression evaluation; hook one is statement evaluation. Events
  have no attached lint meaning. The generic graph is copied from pinned
  cohere's `internal/lint/ecmascript/control_flow_graph`; its preserved source
  header describes that original package, including files not needed here.
  `process_flow` contains only builder, roots, expression, statement, pattern
  and helper files, with package renaming and ASCII comment punctuation.
* `process-symbol-details` requires a newline then `own`, `alias` or `reference`.
  Fields are version, question, opaque program-local symbol identity, symbol
  flags, symbol name and counted declaration records. Alias resolves compiler
  alias symbols. Reference resolves shorthand property's value symbol.
* `resolved-call-declaration` selects CallExpression. Fields are version,
  question, present flag; if present, one declaration record and the compiler
  return type's flags follow. A signature without a declaration is absent.
* `program-imports` selects SourceFile. Fields are version, question, program
  file count; each file contains filename, declaration-file flag, source text
  (empty for declarations), entry count and raw entries. Each entry is syntax
  kind, type-only flag, string-literal-like flag and compiler-resolved filename
  (empty if unresolved). Entries come from static imports/exports/import-equals
  and dynamic import/require syntax; filtering and closure are native.

A declaration record contains filename, declaration-file flag, external-module
flag, default-library flag, syntax kind, pos/end, node flags, function flags,
body-present flag and ancestor count. Each ancestor contains kind, supported
identifier or literal name (empty for a destructured name), node flags and
augmentation flag. No cross-program identity is exported.

`TestProcessQuestionsRawSyntaxAndFlow` checks framing, positions, identities,
const flags, reachable branch structure, evaluation events, reference symbol
provenance, resolved timer declarations and malformed selector/question rejection.
A graph-reachability mutant proves the branch invariant can fail. Rule-level
controls, corpus comparisons and three native mutants separately hold the
native classification and state analysis to unmodified cohere output.
