Temporary: comes out when lazy definite-assignment checks land

Plan published before implementation. Slice only: replace let operand!: string
with let operand: string | undefined. Malformed class-set input can leave this
local absent; the existing !operand tests intentionally handle that absence.
The plain string declaration was probed and gets TS2454 at both tests, so it is
not an honest required-value declaration. Preserve every statement and check.
The existing guards narrow the value before codePointAt and length reads.

Validate all Node tokens and the full upstream baseline, including regex error
cases. Record the next actual refusal before any further source adaptation.
