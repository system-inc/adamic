Temporary: comes out when zero flag inference lands

Plan published before implementation. Slice scanner.ts only: in the four
EscapeSequenceScanningFlags ternaries, spell the zero branch as the AND of
String and ReportErrors. Those distinct single-bit members have disjoint bits,
so the result is precisely zero, within the proven flag domain. Do not change
the enum or introduce a numeric cast. This answers adamic/enum-flags where the
numeric zero branch currently loses flag provenance.

Validate byte-identical Node tokens and all upstream tests, including regex
and template scans. Record the next compiler diagnostic in order.

Validated: four zero branches replaced. Combined 59/80-83 upstream suite:
106,367 pass, zero failures/pending/differences (230.719 seconds). All Node
tokens match. String is bit 0 and ReportErrors bit 1; their AND is exactly zero.
Next refusal: scanner:1864:22, raw tokenFlags = 0 (adamic/enum-flags).

Current profile: superseded by open numeric enums ec67b02. The newest scratch
compiler reaches lowering without this workaround. adapt-slice.sh omits it;
the script and earlier validation remain reproducible historical evidence.
