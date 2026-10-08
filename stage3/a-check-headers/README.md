# A-check headers

The exact rule and historical scan are documented in [REPORT.md](REPORT.md).
A refused program starts with `// a-check: refused <diagnostic reason>`.
A checker error starts with `// a-check: type error TS<code>`.
Clean programs and NotYet stops pass without an expected-error header. NotYet
is not a refusal: retain its actual reason in observation evidence.

Batch 4 adds 113 first-line headers: 87 refusal reasons and 26 checker codes.
[batch4-headers.json](batch4-headers.json) records each measured status, reason,
and the SHA-256 of the unchanged program below its header. The exact pinned
Gate.aCheck accepts all 137 new programs, including 5 clean and 19 NotYet.

The header-predicate proof runs unchanged Gate.aCheck against captured real
compiler diagnostics. Removing each of the 113 headers and independently
replacing each reason/code is rejected: 226 catches. This proof exercises the
header predicate; the ordinary scan separately invokes the real compiler on
all 137 files. [batch4-header-mutants.json](batch4-header-mutants.json) records
both mutants. Headers are restored after every experiment.

Batch 5b is rebuilt from 45487a80 without the two hidden-source branches.
Its own compiler accepts all 85 added or modified programs in place: 34 checked
(including 31 NotYet), 21 refused, and 30 checker errors. Only 29 headers
changed: 16 refusal reasons and 13 checker codes. The 36 added or modified
driver paths are checked in place. [batch5b-headers.json](batch5b-headers.json)
records actual reasons and unchanged body hashes; its 29 wrong-header and
29 removed-header mutants are all caught by the unchanged pinned predicate.
[batch5b-header-mutants.json](batch5b-header-mutants.json) records these 58 catches.
