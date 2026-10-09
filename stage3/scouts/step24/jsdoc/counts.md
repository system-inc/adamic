# Witness native counts

Counted builds match Node stdout; their stderr reports these runtime counts. All allocations are freed. These are scout-local witnesses, not entries in the central oracle corpus.

| Witness | Allocations | Frees | Retains | Releases | Peak | Regions |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| fixtures/01_parseJSDocAllType.a | 11 | 11 | 9 | 16 | 11 | 0 |
| fixtures/02_parseJSDocNonNullableType.a | 13 | 13 | 13 | 20 | 12 | 0 |
| fixtures/03_parseJSDocUnknownOrNullableType.a | 64 | 64 | 61 | 101 | 12 | 0 |

3 originals: native equals Node; 3 sanitizer runs clean; 3 counted runs; 3 source and native mutants caught. No full native JSDoc parser comparison is claimed.
