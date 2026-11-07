# Record table amendment, step B: PASS

Ledger: Adamic ruling **task b5w3ycg, Nov 7 00:55Z, road (b)**.
JavaScript-changing. Base: origin/main 61880be6dd7884dec15cf17c73fd9d56ee760ee4,
with exact 99z efa9c7dd7ffd4504068e6beeae26a2365eb145f1 retained. Every new
commit has the required chg_yj14ddxz dependency line. No Adamic compiler change or
Loom submission is part of this work.

## Owners and exact change

32 pins symbolWalker's visitedTypes[type.id] and visitedSymbols[symbolId]
test/store pairs. Its amendment makes both private tables const MapLike
records, replaces all four array-clear calls with in-place for-in deletion,
removes clear's import and adds a type-only MapLike import. The complete
nested getSymbolWalker owner and import are pinned before and after, at both
pre-43 void and completed undefined return phases. The original four id
reads/tests/stores keep their evaluation points and declines; four pure-delete
operands join the complete indexed-expression inventory. Container identity,
accept callbacks, traversal and finally cleanup remain in place.

30 owns getOwnValues's read. Its whole-owner pins accept the exact raw,
asserted-before and record-only-after forms. The parameter is MapLike<T>
and the read is collection[key]!, with the existing loud presence check and
no cast. Both planners validate all their files before writing. Eight
half-adapted/foreign-body owner mutants require the owning validator's exact
refusal reason. No 99ab or compatibility overload is introduced.

The fresh output matches the measured source, with only the measurement
walker's LF normalized to production CRLF. Native witness freshness checks
match the complete helper, both table declarations, the exact id read/test/store
operations and both cleanup blocks. The existing Node, sanitized native
(--sanitize --explain-checks) and JavaScript witness results are retained:
id 5, ids inserted 4468, 2, 10, and empty/reused tables all match byte for byte.
Already-assigned positive witness ids take getSymbolId's unchanged return path.

## Measurement first

The complete measurement was pushed at 87027cdf7 before any amendment.
All 398 paired lane shards and 1109 calls matched, including all 445 sparse
calls. Baseline arrays: 692; candidate arrays: 0. Returned elements: 1985;
id stores: 734 per side. Non-integer ids, noncanonical keys, order/value
differences and within-run identity failures: all 0. The measured service
action records have complete primitive payload comparisons; Type/Symbol
values preserve ids and own primitive/reference metadata. The independent
reconciler verifies every raw trace, archive hash and all 796 reports.

## Composition and emission

Both full applies pass on the same frozen base. The baseline full apply and
build are retained from measurement; the amended full apply and compiler/harness
build are fresh. Exactly core.ts and symbolWalker.ts differ across all 746
source files. Full-apply patch accounting changes only rows 30 and 32 plus
its total. Both normal owner verifiers pass at their actual immutable
before/after apply snapshots. Adamic types over all 79 compiler roots exits 0
before and after.

All 83 paired replay outcomes are exactly identical, including
14 inherited nonzero outcomes per side.
Neither copied source tree changes. Completed-tree replay still refuses 30
at the later addRange assertion and 32 at appendSourceMap's later runtime
change; these are identical baseline outcomes. They are distinguished from
the passing normal verifiers at the owning phases.

Core JavaScript is byte-identical with comments kept and removed. An
independent literal port produces exactly the symbolWalker JavaScript in
both modes. Of nine compiler JavaScript products, seven are byte-identical;
only _tsc.js and typescript.js change, solely in the two table initializers
and four cleanup statements within getSymbolWalker. A full-file byte comparison
checks that no other emitted byte changes. [All exact source and emitted
hunks](record-tables-hunks.patch) are shown as valid zero-context patches.

Both public declarations, typescript.d.ts and tsserverlibrary.d.ts, are
byte-identical. Every other declaration is also identical except this one
user-approved hunk in typescript.internal.d.ts; private table types do not
surface there:

```diff
-function getOwnValues<T>(collection: MapLike<T> | T[]): T[];
+function getOwnValues<T>(collection: MapLike<T>): T[];
```

## Normal whole lane and mutants

All 398 ordinary shards pass, in 20 pieces, each committed and plainly pushed.
Final counts: 106366 passes, the same 1 expected failure, 0 skips. The unchanged
identity, API, baseline, count and source/build guards remain enabled.
No measurement observers are present. The independent reconciler reopens every
archive, checks all hashes and every unique report, and verifies the frozen
source and build after completion.

One hook timeout during concurrent checks is retained, not waived:
file-013/templateLiteralTypes1's before-all hook exceeded its existing 40-second
limit in piece 2. The other 19 passing shards were retained. Only file-013 was
retried, in a fresh output directory, with the same 40-second hook and
90-second shard limits; it passed in 20.823 seconds. The
complete piece including inspection and retry took
344.014 seconds, below 25 minutes.
Lane concurrency was reduced while replay/census ran, then restored when both
finished. Final lane counts include the successful guarded run; the original
failed report remains in the first piece-2 archive and firstAttempt ledger.
A retry-wrapper output-directory collision ran no tests and is also retained.

The first passing lane checkpoint was briefly blocked by whitespace-only
context lines in the review patch. The corrected zero-context patch and that
checkpoint were plainly pushed together. The writer now checks a clean index
and unstages its own paths after a failed check. No source, test, guard or
limit changed. The raw publication and failure logs remain archived.

Restoring the complete sparse-table owner produces four TS2345 errors at the
now record-only helper calls. Restoring the original helper union and cast
passes checking and both ordinary binder rows return Refused at
core.ts:1326:26, a cast the runtime cannot check. A cast restored alone with
the new parameter is also rejected by the whole-owner contract; it is not
misrepresented as a downcast. Each source mutant differs from the fresh
output only in its intended restored file; all other source bytes match.

## Ordinary census and next group

All 79 roots are paired in groups of 20, 20, 20 and 19, each committed and
plainly pushed. The baseline retains 47 exact ordinary observations and measures
the remaining 32; the retained source manifest is unchanged. All amended roots
are measured on the fresh full apply. The driver's inspect function is
byte-for-byte the normal stage3/census/tool implementation: ordinary Load,
then ordinary Lower only if checking succeeds. Its one-file Go overlay changes
only main's enumeration to bounded independent roots; no compiler, checker,
refusal, or backend code is overlaid. All authored Go is .go.txt.

All 73 roots formerly refused at 1326:26 now reach **NotYet
core.ts:1364:29**, assign's argument arg in hasProperty(arg, p). The other
six root outcomes are unchanged. The exact next source is:

```ts
export function assign<T extends object>(t: T, ...args: (T | undefined)[]): T {
    for (const arg of args) {
        if (arg === undefined) continue;
        for (const p in arg) {
            if (hasProperty(arg, p)) {
                t[p] = arg[p];
            }
        }
    }
    return t;
}
```

The unchanged rule is (*lowering).detachedOwnRefusal, declared at
internal/lower/detached_own.go:69, constructing this NotYet at line 94:
`a non-object representation passed through an opaque object parameter`.
It is registered as ordinal 900 at line 144. The opaque object slot requires
an Object representation; this branch stops when representation is unknown
or different. [Exact source, rule and line](record-tables-next-stop.json),
[every paired root](record-tables-census.json) and the four archives retain
the complete evidence. assign and elementAt are untouched by this amendment.

## Evidence

PROOF-RECORD-TABLES-STEP-B.json is the final machine-readable result.
record-tables-composition, owner-mutants, replay, census and after-lane archives
retain raw evidence with internal SHA256SUMS and adjacent archive hashes.
record-tables-stepb-final retains all drivers, final reconciliation and raw
closeout logs. Historical measurement and earlier-road artifacts remain
unchanged. This continuation's changes stay inside adaptations 30 and 32 and
under 200 files relative to its frozen base. No started process is left running.
