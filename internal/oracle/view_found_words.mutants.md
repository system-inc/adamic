# The found word of a checked view read, and a boolean | undefined slot read as boolean (#yf3k0aj)

**Y1, the nullish words dropped.** In internal/native_runtime/c/object.c's adamic_object_view, the three lines under
"The found word is what the JavaScript backend's check names" removed, main's code before this change. Natively an
undefined or null slot, or a boolean | undefined one holding undefined, says "found unsupported representation". Red:
`TestViewFoundWordsAgree/undefined_string.ts`, `null_object.ts` and `undefined_boolean.ts`, on their native rows.

**Y2, the boolean | undefined pass dropped.** The `actual == adamic_rep_maybe_boolean && wanted == 2` block removed,
main's code before this change. A boolean | undefined slot holding a boolean, read as boolean, stops natively. Red:
`TestViewFoundWordsAgree/present_boolean.ts` and `present_false.ts`, on their native rows.

Must stay green under each: `null_string.ts`, `undefined_computed.ts` and `present_string.ts`, which agreed on main,
and `boolean_as_string.ts`, a boolean found where a string is wanted.
