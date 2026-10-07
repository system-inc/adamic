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
