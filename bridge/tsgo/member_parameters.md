# Member parameter source text

`member-parameters` is a `tsgo_inspect` question, ABI version 1. Its anchor is an
exact AST node. It reads that node's symbol without following aliases. The answer
is framed with the existing decimal UTF-16 length fields and UTF-8 C buffer.

After the version and question name:

- Declaration count, zero when the node has no symbol.
- For each declaration: a function-like boolean. If true, parameter count and
  source text for each parameter, excluding leading trivia.
- For that declaration's parent type: the declaration count of its `includes`
  property, zero if absent, and the same function-like/parameter fields for each.

No parameter relation, lint decision, message or edit is returned. Adamic compares
these lists according to production cohere. Returned text is owned by the C buffer
and copied to an owned Adamic string before the buffer is freed. No new handle is
introduced. A released program is refused by the existing registry before the
question runs. Suffixes are refused, and all ordinary exact-node guards apply.
