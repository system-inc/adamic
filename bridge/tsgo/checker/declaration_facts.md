# Symbol and binding declaration facts

The legacy wire starts with decimal UTF-16 length frames: version `1`, question
mode, then the fields below. Paths are rooted compiler file names serialized as
strings. Declaration positions are UTF-8 byte offsets in the declaration's own
file; callers map to UTF-16 only when constructing a finding.

`node-symbol-details` calls the borrowed run checker's GetSymbolAtLocation.
It returns presence. If present it returns the run-scoped symbol identity, flags,
name, declaration count, and every declaration in compiler order. Each declaration
has path, kind, start, end, declaration-file flag, default-library flag, immediate
parent kind/name/start/end, JSDoc tag count and records (kind/name/start/end/text),
and parameter count and names. IDs are consecutive safe integers, never addresses,
and remain valid only while the program handle remains live.

`binding-declarations` asks GetSymbolAtLocation for an Identifier and substitutes
GetShorthandAssignmentValueSymbol for a shorthand property's value binding. It
returns presence, then flags, declaration count and every path/kind/start/end.
This retains the worker's frame format. It is not the complete binding-symbol
identity contract requested by the wave 01 audit.

A no-suffix question retains the legacy exact-node selector. A file-wide question
may append `LF byteStart LF byteEnd LF kind`. The enclosing selector must be this
SourceFile, and the suffix must identify an exact node in this file. Decimal
numbers must be canonical. The harness uses this form through
askFile(ReadsOtherFiles), so returned foreign declarations pass the declared
program-read guard. Default-library consumers also declare ReadsDefaultLibrary.
No rule opens or parses a foreign file.

Unknown questions, invalid selectors and incompatible kinds return bridge errors.
The current C-to-Adamic adapter still requires the separately owned library error
path for a returned refusal; the inherited named pending control is preserved.
No question registry, alternative checker lifecycle or Go lint verdict is added.

`symbol-provenance` uses wave 24's frame grammar: shared symbol ID, flags,
declaration count, then each declaration's path, declaration-file, default-library
and external-module flags, followed by every ancestor from declaration to
SourceFile (kind, name, byte span, node flags and global-augmentation flag).
It accepts the same guarded same-file selector as the other declaration facts.
Source-less declarations return an error. The alias variant is not registered.

Parent names contain simple literal/identifier names. Computed names have an empty name field; the parent kind and byte span identify their complete syntax.
