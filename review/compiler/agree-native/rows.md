| Test | Change | What only native shows |
|---|---|---|
| TestTypedArrayFromArray | Opt into native; remove FromArray IR assertion after mutant proof | Native array-copy constructor dispatch, ignored by JavaScript emission |
| TestTypedArrayFromEmptyArray | Opt into native; remove FromArray IR assertion after mutant proof | Native empty-array constructor dispatch, ignored by JavaScript emission |
| TestNativeAgreementRejectsWrongOutput | New parallel harness control | Planted C string while JavaScript prints the correct string |
| TestNativeAgreementRejectsWrongExit | New parallel harness control | Planted native return code after correct stdout |

All other typed-array rows and existing agreement controls retain their assertions. No acceptance row deleted or skipped; no golden IR snapshot remains in the two constructor rows.

| Test | Change | What only native shows |
|---|---|---|
| TestRegExpUnicodeClassSourceAgreesWithNode | Shared native helper for PR #256 row | Lowered source escaping metadata; JavaScript recomputes .source |
| TestRegExpSlashClassSourceAgreesWithNode | Shared native helper for PR #256 row | Lowered source escaping metadata; JavaScript recomputes .source |
| TestRegExpEscapedSlashSourceAgreesWithNode | Shared native helper for PR #256 row | Lowered source escaping metadata; JavaScript recomputes .source |
