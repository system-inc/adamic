# Record id tables, measurement first

Ruling: b5w3ycg, Nov 7 00:55Z, road (b). Adamic keeps its intentional
past-end-write panic. tsc's visitedTypes and visitedSymbols become records.
No adaptation amendment has been built at this checkpoint.

Base: origin/main 61880be6dd7884dec15cf17c73fd9d56ee760ee4, retaining exact
99z efa9c7dd7ffd4504068e6beeae26a2365eb145f1. The same branch's prior pushed
history is preserved. New commit messages carry the required dependency line.
The main adaptation inputs are byte-identical to the completed earlier base.

The complete symbolWalker.ts was read. The actual pins are adaptation 32's
visitedTypes[type.id] test/store and visitedSymbols[symbolId] test/store,
occurrences 1 and 2 of each, all deliberate declines. Thus the table port
folds into 32. Existing unrelated casts in visitSymbol are unchanged.
30 owns getOwnValues's indexed read and will own its record-only amendment.
No 99ab is needed.

The candidate uses const MapLike<Type> and MapLike<Symbol> tables, imports
MapLike as a type, and replaces the two pairs of clear calls with in-place
for-in deletion loops. core.clear is array-only (splice(0)); no global helper
is changed. In-place deletion preserves the private containers' identities
through callbacks and finally blocks. Every original id read, store, test and
traversal remains at its original point. The record-only getOwnValues keeps !.

Step 1 runs both original and candidate compilers over the identical complete
lane unit inventory in pieces below 25 minutes. The observer records every
id store and each call's keys and ordered returned value identities. Type and
Symbol payloads use compiler ids plus own primitive fields and reference
metadata; the action records contain their complete primitive fields. The
observer separately checks exact Object.is identity against an own-key record
projection in each run, so cloned/substituted values or reversed order cannot
pass merely by resembling the same data. Original and candidate process
traces are compared in order, with process scheduling separated from value
order. Only array/record representation metadata and observer sequence labels
are excluded; returned values and key order are never sorted or normalized.

Seeds excluded from the oracle prove detection of reversed order, substituted
object identity, fractional ids and noncanonical integer strings. Both builds
and both harness preparations passed. A setup attempt using a relocated
node_modules path and a harness invocation from the wrong directory were
corrected before the accepted measurement; their logs are retained.

The proposed changes and observer exist only in scratch trees. The normal
adapters and site ledgers remain unchanged until all value/id checks pass.
RECORD-TABLE-COUNT.json and record-tables-measure-NN archives record each
completed piece, committed and plain-pushed immediately with the normal lane
identity, API, baseline and count guards. No result beyond the recorded
completed units is claimed.

## Independent runtime witnesses

The extracted original and record table operations give byte-identical Node
output for id 5, ids inserted as 4468, 2, 10, and empty tables. Both private
tables are checked, followed by another empty walk using the same containers.
All three record witnesses compile and run on the native backend with
--sanitize --explain-checks and on the JavaScript backend. All six executions
exit 0, have empty stderr, and match Node stdout byte for byte. The id-5
record insertion succeeds. The previous sparse array insertion panicked.

record-tables-witnesses.tar.gz includes the extracted sources, generated
JavaScript, compiler check diagnostics, outputs, driver, and summaries. The
initial witness extractor also included the now-unused array-only clear
helper; its erased unknown[] diagnostic is retained separately. Removing
that unused helper from the record witness changed none of the tested code.
The record read's ! is a runtime checked non-null assertion, not a new cast.
The whole-oracle gate remains pending; no adapter implementation is changed.

## Declaration constraint found during measurement

The unchanged full baseline and the built measurement candidate have identical
public typescript.d.ts bytes. typescript.internal.d.ts necessarily has exactly
one hunk: getOwnValues's parameter loses ` | T[]`. This is the requested true
parameter type, and cannot simultaneously retain the old internal declaration
bytes. record-tables-declarations-measured.json preserves both hashes and the
complete diff. Clarification was requested while the independent oracle
measurement continues. No adapter amendment is built pending this constraint.
