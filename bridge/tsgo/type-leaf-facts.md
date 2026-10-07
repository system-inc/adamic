# type-leaf-facts

Request: `type-leaf-facts` LF canonical nonzero live type identity, optionally LF
property name. A property name is the entire remaining suffix, including any LF.
Types belong to the current program. Missing, zero, unknown, noncanonical or
out-of-range identities are rejected. Node coordinates still identify an exact
live node in that program, as for every inspect question.

Version 1 inspect framing: every field has its decimal UTF-16 length, LF, then
text. Outer C buffer length is UTF-8 bytes and the existing ownership applies.
After version and `type-leaf-facts` come:

1. Raw type flags, array-or-tuple boolean, number of call signatures.
2. String-literal-present boolean and literal string (empty when absent).
3. Symbol-present boolean. When present: valid UTF-8 display symbol name (invalid
   internal bytes replaced with U+FFFD), declaration count, then
   each declaration's source filename, declaration-file boolean and
   default-library boolean. This is the existing symbol-origin record.
4. Property-requested boolean; when true, property-present boolean.

The bridge returns compiler facts only. The `.a` consumer decides whether a
symbol comes from the default library and how these facts affect a lint rule.
Identities are borrowed, and stale program handles reject the question before
looking up any node or identity.
