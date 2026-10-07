Temporary: comes out when flag return inference lands

Plan published before implementation. Slice scanner.ts only: annotate the
getNumericLiteralFlags arrow's return type as TokenFlags. Its expression stays
tokenFlags & TokenFlags.NumericLiteralFlags. This operation is within the flag
domain; the scratch compiler currently misjudges the inferred function result
as a narrowed TokenFlags.None slot. An explicit return annotation closes that
refusal in a scratch probe, without any runtime change.

Validate the Node token stream and full upstream baseline. Census reason:
enum-flags; feature branch flag-enums, inference gap still present at its tip.
