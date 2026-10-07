Temporary: comes out when numeric enum lookup views land

Plan published before implementation. Slice scanner.ts only: the private
charCodeToRegExpFlag Map key type and characterCodeToRegularExpressionFlag's
parameter become number. All table entries and lookup statements stay exactly
the same. The scanner passes arbitrary identifier code points when checking
regex flags; unknown codes must return undefined, not be asserted into the
closed CharacterCodes enum. This answers adamic/enum-members at scanner:2034.
The exported helper is internal and stripped from the public API declaration.

Validate known ASCII flags and unknown ASCII/Unicode keys, Node tokens and the
full upstream baseline. Record the next diagnostic in order.

Superseded before a source implementation was pushed: open numeric enums
at ec67b02 accept the original lookup types and arbitrary numeric codes.
An untracked two-annotation probe passed the full 106,367-test baseline and
all Node tokens with the legacy compiler. The current profile keeps the source
lookup types unchanged and reaches the same Error-value lowering blocker.
No adaptation 86 implementation is needed or delivered.
