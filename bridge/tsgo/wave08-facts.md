# Wave 08 checker questions

These questions use the existing length-prefixed UTF-8 field framing and response
version. Node lookup uses the existing exact compiler bounds and kind contract.
Handles, symbol IDs and decoded strings have the same lifetime as existing facts;
no pointer crosses the bridge. Queries after release are rejected before lookup.
The implementations are separate files on both sides. Two shared registration
lines route them through `wave08Facts`; old declaration modes still return nil.

* `wave08-symbol`: raw symbol identity, reserved attributes, symbol flags,
  declarations (file, kind, bounds, node flags), type-only alias status, resolved
  target name/flags/identity, and optional value declaration metadata. Export
  specifiers use their bound alias; shorthand properties use their value symbol.
  Missing or unknown symbols have zero IDs. The decoder reads value metadata
  only for a nonzero target ID. Alias flags and declarations are facts, not rule
  decisions.
* `syntax-metadata`: node flags, modifier flags, compiler type-node predicate,
  optional initializer/expression bounds excluded by loop ancestry, and generator
  status. `for` supplies the initializer; `for-in`/`for-of` supply the expression.
* `module-links`: SourceFile only. All program files, declaration-file status,
  and top-level import/export module specifier bounds with resolved target paths.
  Unresolved targets are empty. Adamic filters the graph, determines runtime
  aliases, finds cycles and decides whether a load-time read is unsafe.

Unexpected suffixes are rejected. Direct checker tests compare these facts to
independent calls into the pinned TypeScript checker. Each new question has a
compiling semantic mutant caught by those tests. No Go lint rule is called by the
bridge; the separate Go oracle calls the production cohere rules unchanged.
