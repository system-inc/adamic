# Record id tables, measurement first

The 73-root stop at core.ts:1326:26 requires elementAt's earlier wall to be
past. Exact 99z is retained here; the ordinary baseline binder probe confirms
1326:26 on this base.

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

## Final measurement gate and stopping point

PASS: all 398 lane units ran exactly once on each side, in 20 pieces, each
committed and plain-pushed before adapter construction. The largest piece was
201.761 seconds. Every shard report passes its ordinary identity, API,
baseline and count guards. Each side has 106366 passes, the same 1 expected
failure and 0 skips. The independent reconciliation reopens every archive,
checks both layers of hashes, reconstructs all ordered process traces from
raw rows, and checks all 796 reports against the unchanged lane plan.

All 1109 calls match element for element and in order, including all 445
formerly sparse calls. The baseline has 692 array calls; the candidate has
zero. There are 1985 returned elements and 734 id stores per side, zero
non-integer ids, zero noncanonical own keys, zero order/value differences,
and zero within-run reference identity failures. The service action records'
complete primitive fields are compared; Type and Symbol values retain their
ids and own primitive/reference metadata. No returned value order is sorted.
The generator assigns type ids by incrementing typeCount in createType and
symbol ids by incrementing nextSymbolId in getSymbolId (checker.ts).

The fresh unchanged full apply exits 0 and its 746 source files match the
uninstrumented baseline measurement exactly. Both the baseline and an
uninstrumented candidate source copy pass adamic types over all 79 compiler
roots, exit 0. Both ordinary binder rows agree on each side: baseline is
Refused at core.ts:1326:26; the measured candidate reaches NotYet at
core.ts:1364:29, assign's hasProperty(arg, p). The exact source is recorded in
record-tables-candidate-probe.json. (*lowering).detachedOwnRefusal constructs
this NotYet in internal/lower/detached_own.go:94. This is a binder observation,
not a claim that the amended 73-root census has been completed.

STOP before construction: the public declaration is byte-identical, but the
internal declaration has exactly the required parameter hunk documented
above. The record-only signature and byte-identical internal declaration
cannot both hold. No compatibility overload, false public signature or new
cast was introduced to conceal this. Clarification on that single internal
hunk was requested while measurement continued and remains pending.

The implementation files and site ledgers of 30 and 32 are unchanged. The
table port belongs in 32 and the helper amendment in 30; no 99ab was created.
No amended full apply, replay, lane, owner mutants or after census is claimed.
Optional baseline census prework completed and archived groups of 20 and 20
roots; its third group was stopped with 7 observations when construction was
blocked. The partial rows and stop ledger are retained, not counted as a
completed group. All started background measurement processes are stopped.

PROOF-RECORD-TABLES.json records this exact status. record-tables-final-measure
tarball retains the final raw logs, source bytes, manifests, provenance,
independent reconciler, ordinary census driver and partial prework. Its
adjacent SHA256 and internal SHA256SUMS cover the archive and every member.
No source change or result is submitted to Loom.

## Continuation after the declaration clarification

The user accepted the single internal signature hunk and clarified that the
two public files are typescript.d.ts and tsserverlibrary.d.ts. The original
measurement stop above is historical. Step B is now complete and passes;
see [the final amendment proof](RECORD-TABLES-PROOF.md) and
PROOF-RECORD-TABLES-STEP-B.json for the current result.
