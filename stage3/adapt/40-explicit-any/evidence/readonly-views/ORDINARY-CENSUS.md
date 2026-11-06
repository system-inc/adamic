# Ordinary compiler census

All 79 compiler roots were measured independently with the ordinary production compiler. The driver also emits three non-TypeScript file records and one whole-program record; those are preserved raw and excluded from this 79-root comparison.

| Kind | Before | Final |
| --- | ---: | ---: |
| NotYet | 4 | 4 |
| Refused | 73 | 73 |
| accepted | 2 | 2 |

Changed first stops: 0.

Complete per-root records, including exact locations, reasons and fixes, are in ordinary-comparison.json. Only tree prefixes and timing fields are normalized. The compiler code is unmodified from base 63e14a0b2b61cf251ad74292e42d5fd09af804b3.
