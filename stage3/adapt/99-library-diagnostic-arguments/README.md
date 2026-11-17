# Readonly diagnostic arguments

Adamic, task #w743639 (its D). Type-only: the emitted JavaScript does not change.

## The lie

`DiagnosticArguments` (types.ts, `@internal`) is a mutable array of a wide union.
`processSourceFile` in program.ts hands a rest parameter's `string[]` on as a
diagnostic's arguments (`addFilePreprocessingFileExplainingDiagnostic`), which
makes it a view that could write a number, a boolean or a SourceFile where a
string is read. Adamic refuses it (program.ts:3492). Diagnostic arguments are
built at the call that reports them, and only read after, when they are formatted
into the message.

## The edit, and why it runs after 98-axolotl-d

`adapt.cjs` inserts `readonly ` before the alias's array type. The directory
sorts after `98-axolotl-d` for one reason: 98-axolotl-d replaces
DiagnosticArguments' whole type, so a `readonly` inserted before it would be
overwritten. The edit pins the text 98-axolotl-d leaves,
`(string | number | boolean | readonly string[] | SourceFile | undefined)[]`, and
refuses anything else. An alias already readonly keeps its bytes. Ruled by Adamic
on #w743639 (road 1: a directory after 98, never an edit inside it).

## Certificate

The compiled program's gate, as for adaptation 72.

## Measured

From J-4t94cm2 (sealed, scratch at main 0578cbd7 holding this and four sibling heads of #w743639), the tree whole against the same tree with this directory dropped:

- The meter's latent total: 3,694 without it, 3,689 with it.
- Gone, five Refused sites, the same lie: program.ts:3492, and programDiagnostics.ts 168 (two), 240 and 269, where a DiagnosticArguments or an empty array was seen as one that can write where `never` is read. Nothing new.
- Sources: types.ts alone, one line.
- Stock tsc over `src/compiler`, `src/services` and `src/server`: 0 new errors.
