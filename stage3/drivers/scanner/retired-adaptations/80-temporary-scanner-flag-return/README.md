Temporary: comes out when flag return inference lands

Plan published before implementation. Slice scanner.ts only: annotate the
getNumericLiteralFlags arrow's return type as TokenFlags. Its expression stays
tokenFlags & TokenFlags.NumericLiteralFlags. This operation is within the flag
domain; the scratch compiler currently misjudges the inferred function result
as a narrowed TokenFlags.None slot. An explicit return annotation closes that
refusal in a scratch probe, without any runtime change.

Validate the Node token stream and full upstream baseline. Census reason:
enum-flags; feature branch flag-enums, inference gap still present at its tip.

Validated in the combined 59/80-83 full suite: 106,367 pass, zero baseline
differences/failures/pending, 230.719 seconds. Node tokens match. Explicit
return annotation advances to the reScanJsxToken Boolean parameter refusal
(scanner:538:9); the expression and Scanner interface are unchanged.

Current profile: superseded by open numeric enums ec67b02. The newest scratch
compiler reaches lowering without this workaround. adapt-slice.sh omits it;
the script and earlier validation remain reproducible historical evidence.
