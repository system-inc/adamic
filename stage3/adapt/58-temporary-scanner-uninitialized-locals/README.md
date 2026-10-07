Temporary: comes out when 01a1130a's lazy initializer check lands

Plan published before implementation. Slice scanner.ts only: the text declaration
becomes var text: string; its initial setText call still receives textInitial.
Nothing reads text before setText assigns newText || "". Replace tokenValue's
definite-assignment marker with its plain typed, uninitialized declaration.
Probe the same form for the nested operand declaration if it is the next refusal.
No eager value is invented. Preserve all scanner statements and observable Node
undefined values. Record read-before-assignment evidence, including the driver's
call contract, rather than assuming every exported getter is safe at every time.

Validate byte-identical Node tokens and the full upstream baseline. Compiler
probes decide which remaining refusal is next. This is a temporary source
spelling change, separate from the verbatim declaration gatherer.

Validation: the final adaptation has two edits, text and tokenValue only.
Node matches all 509,014 tokens. Full upstream baseline: 106,367 passing,
zero failing/pending/differences, 233.054 seconds. Idempotence: zero edits.
The exact next compiler refusal is flag-return inference at scanner.ts:528:33,
not operand's marker; the earlier operand probe ran into checker diagnostics.

Read evidence: text is assigned by the first setText before any read; the
existing intervening-read mutant catches violations. tokenValue differs:
getTokenValue can expose undefined before a value-bearing token, and speculation
helpers can save that absence. Removing its marker preserves that Node value.
The token driver never calls getTokenValue or speculation/rescan methods;
value-dependent scanning paths assign their values first. This is not a proof
that every exported getter has a required value at every time. The bare local
spelling passes the checker; native lowering and capture handling remain to
be tested. The nested operand's plain spelling gets TS2454 and belongs to 59.

Current profile: superseded by readiness bc9f5d7. Its literal-initializer probe
passes on the feature alone and in the corrected scratch integration. Keep this
script as the reproducible legacy profile; adapt-slice.sh does not select it.

Captured marker control also passes on the integration. Original textInitial!
followed by setText(text) compiles but panics on absence; 57 remains selected
with setText(textInitial). Retiring 58 does not retire that separate adaptation.
