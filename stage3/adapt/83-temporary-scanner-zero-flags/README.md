Temporary: comes out when zero flag inference lands

Plan published before implementation. Slice scanner.ts only: in the four
EscapeSequenceScanningFlags ternaries, spell the zero branch as the AND of
String and ReportErrors. Those distinct single-bit members have disjoint bits,
so the result is precisely zero, within the proven flag domain. Do not change
the enum or introduce a numeric cast. This answers adamic/enum-flags where the
numeric zero branch currently loses flag provenance.

Validate byte-identical Node tokens and all upstream tests, including regex
and template scans. Record the next compiler diagnostic in order.
