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

The initial batch 5 scan checks 529 added or modified programs in place, including 501 new
files and 39 staged driver paths. All pass: 465 checked (including 114 NotYet),
43 refused, and 21 checker errors. Only 72 headers changed: 37 refusal reasons,
15 checker codes, and 20 accepted/NotYet expectations.
[batch5-headers.json](batch5-headers.json) records unchanged body hashes and actual
reasons. The unchanged pinned Gate.aCheck rejects all 72 wrong expectations and
all 52 removed error headers against captured real diagnostics: 124 catches.
The separate ordinary scan invokes the real compiler on every original path.
[batch5-header-mutants.json](batch5-header-mutants.json) records that proof.

The final landing scan also includes 137 batch 4 programs new relative to main.
All 666 paths pass in place: 546 checked (including 143 NotYet), 76 refused,
and 44 checker errors. There are 638 new files relative to main and 501 relative
to the area base; all 39 changed driver paths are included. Another 61 stale
headers are refreshed without changing program bodies, for 133 header changes
in this unit. Wrong expectations are caught for all 133; removing all 56
current error headers is caught too: 189 catches using the unchanged predicate.
The manifests above now record the final combined scan and header proof.
