# Third wave 26 batch raw questions

ABI 1, framing and ownership stay unchanged. Each question has a new Go file
and a matching Adamic `.a` file. Only two single-line switch registrations were
added to facts.go. No shared generator, generic lint harness or parser was edited.

`binding-structure` requires an exact SourceFile and rejects parse diagnostics.
It returns the complete preorder syntax population using the node-structure
record schema documented in WAVE_26_NEXT_FACTS.md. This separate question adds
raw initializer, element-access, unary and loop role metadata without changing
the previous batch's schema. ExpressionWithTypeArguments has an expression role;
ElementAccessExpression has receiver and argument roles; unary expressions have
operand and operator; shorthand properties have name and default initializer;
for-in/of nodes have their initializer. Ordinary binding/property declarations
have name and initializer roles. Identifier, string, template and numeric literal
text is the compiler's cooked text. It returns syntax, not value-reference,
write-access, global-object or prototype-extension predicates. Adamic performs
those structural judgments itself in core_access.a and the individual rules.

`binding-origin` requires an exact Identifier. After the ordinary header it has
two groups: ordinary GetSymbolAtLocation and, only for an immediate shorthand
property parent, GetShorthandAssignmentValueSymbol. Each group has symbol-present
boolean and declaration count. Each declaration preserves compiler order and
has source path, actual SourceFile.IsDeclarationFile, Kind, byte Pos and End.
There are no rule names, findings, edits or judgments in the response. The native
side chooses the first declaration or first declaration in this file, as the
production rule does. File/kind/start/end keys identify the same immutable AST
nodes that the Go oracle compares by pointer within this program.

The declaration-file bit is deliberately different from default-library
classification used by the previous batch. Production resolvesToAGlobal accepts
any declaration file, including a project's own ambient declarations; it rejects
an ambient declaration written in an ordinary source file. Substituting a
stricter library-only test would silently change the rule.

Raw contract tests compare every serialized origin with direct compiler symbol
access, including shorthand declarations and declaration order, verify the syntax
population, and reject malformed questions. Byte-oracle mutants separately prove
that the new syntax and origin facts affect real findings. Registry refusal and
sanitizer ownership checks continue to use the unchanged bridge mechanisms.
