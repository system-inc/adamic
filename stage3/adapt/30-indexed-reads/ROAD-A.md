# getOwnValues road (a), Step A complete

Do not build Step B. The whole-lane count contains **445 sparse-array calls**,
and checked native cannot hold the required hole: the id write stops at
`a[5] = "x"`, before filter, with exit 70. This fails the explicit admission
condition. The full proposed body also has two measured Lower NotYet stops.
No adapter, pin, owner, API sanction or compiler rule was amended.

Adamic's ruling: task b5w3ycg, Nov 7 00:12Z, road (a), distinguish arrays from
records with tsc's actual isArray helper; filter(() => true) on the array arm,
and the existing own-key loop on the narrowed dictionary with 30's ! retained.
The previous MEASURE.md remains historical.

## Base and requested first stop

Base: origin/main 5a3e6607bb58bca84f65e01a4d153b3a2ba8b301 plus the requested
99z efa9c7dd7ffd4504068e6beeae26a2365eb145f1, which was not on main at fetch.
Merged measurement base: b74a665d4069f959a5bc71f555da4d77ff83d4c7. The existing
pushed history was preserved. Every new authored commit uses the required
`never cut alone: lands with chg_yj14ddxz, because the census reads its stop`
first line. Main and the prerequisite are inherited; this task changes only
measurement files under adaptation 30.

Main's complete adaptation inputs are byte-identical to the earlier completed
apply at 8f4efea05. The measurement tree copies that full apply and applies
exact 99z. The original getOwnValues and 30's ! remain unchanged there. Only
the scratch counted tree receives the observer call. The compiler bundle and
harness builds both exited 0. The final source comparison proves the observer
is the only src-byte difference from the uncounted main-plus-99z source.

getownvalues-base.json records both ordinary binder rows over uninstrumented
source. Both reach core.ts:1326:26, Refused: a cast the runtime cannot check.
This verifies the requested base; no Step B 73-root census is claimed.

## Complete whole-lane call count

All **398/398** unchanged lane manifest units ran once as the accepted union,
in 20 pieces, each committed and plain-pushed immediately. The full oracle
counts are **106366 passing, 1 expected failing, 0 skipped**,
exactly the lane plan. The single upstream public-API failure and its API
sanction passed the normal lane guards. No title, runner or case was skipped.
The unchanged stage3/lane/shard.py runners form the sharded whole oracle used
by stage3/lane/run.sh. Each shard retained normal identity, count, baseline and
API guards. Four units ran concurrently with the runner-supported 90-second
unit deadline. The longest measured piece run was 89.038 seconds, below 25 minutes.

| Argument observation | Call count |
| --- | ---: |
| All calls | 1109 |
| Array | 692 |
| Plain record, including null-prototype records if present | 417 |
| Anything else | 0 |
| Sparse array | 445 |
| Array with an extra own non-index property | 0 |
| Array with a present undefined element | 0 |
| Present undefined elements in total | 0 |
| Accessor indices | 0 |

These are per-call observations, not deduplicated heap identities. Every array
call retains its length, present count, hole count, sparse flag, extra own
properties and present-undefined count in calls.jsonl. Full length and hole
distributions are in GETOWNVALUES-COUNT.json and the independent reconciliation.
Observed lengths range from 0 to 4469. The intrinsic length property is
recorded as length, not mistaken for an extra own property. Symbols and
non-enumerable extra properties are included. Accessor indices are counted
separately without invoking getters. Plain-record detection accepts Object
prototypes across VM realms and null prototypes; other shapes are separate.

Excluded seeds verify empty, sparse, present-undefined and extra-property
arrays, ordinary and null-prototype records, a Date as other, and VM realm
objects/arrays. They are in a separate directory and absent from every lane
piece. The final reconciliation checks unique (shard,pid,sequence) call IDs,
every per-call invariant, every aggregate/distribution, archive hashes, all
398 reports and the exact once-only unit inventory. Source and built inputs
were hash-frozen before counting and rechecked after every piece.

| Piece | Units | Seconds | Calls | Sparse-array calls |
| --- | ---: | ---: | ---: | ---: |
| 1 | 20 | 68.946 | 20 | 0 |
| 2 | 20 | 54.433 | 622 | 367 |
| 3 | 20 | 65.441 | 22 | 2 |
| 4 | 20 | 70.526 | 20 | 0 |
| 5 | 20 | 86.034 | 23 | 0 |
| 6 | 20 | 89.038 | 60 | 34 |
| 7 | 20 | 80.319 | 21 | 0 |
| 8 | 20 | 75.552 | 22 | 0 |
| 9 | 20 | 70.248 | 20 | 0 |
| 10 | 20 | 75.204 | 20 | 0 |
| 11 | 20 | 81.780 | 20 | 0 |
| 12 | 20 | 81.709 | 22 | 0 |
| 13 | 20 | 70.771 | 24 | 0 |
| 14 | 20 | 73.068 | 23 | 0 |
| 15 | 20 | 65.775 | 22 | 0 |
| 16 | 20 | 66.717 | 22 | 0 |
| 17 | 20 | 64.253 | 68 | 42 |
| 18 | 20 | 66.222 | 20 | 0 |
| 19 | 20 | 49.079 | 20 | 0 |
| 20 | 18 | 51.444 | 18 | 0 |

## Node original against proposed body

The actual owners are extracted from tsc, including its helper signature
`isArray(value: unknown): value is readonly unknown[]` with Array.isArray
as the body. Both original and proposed functions retain identical call
signatures. The proposed record read keeps the existing !.

All four exact-body Node witnesses agree byte for byte, stdout and stderr,
with exit 0:

| Witness | Result |
| --- | --- |
| Empty array, then a[5] = 'x' | Length grows 0 to 6; five holes; filter returns one x |
| [undefined, 'x'] | Length 2 result, undefined then x |
| Empty array | Length 0 result |
| {'1':'a', '0':'b'} | Length 2 result, b then a |

getownvalues-node.tar.gz retains source, exact owners, output bytes, base
and preparation provenance and the excluded counter seeds.

## Checked native and JavaScript results

All four full proposed-body witnesses fail compilation on both backends at
fixture line 8:37, the narrowed collection read in collection.filter. No
full-body witness is claimed to run or match Node. localRead at
internal/lower/locals.go:318 asks arrayPredicateDomain (library_array_predicate.go:97)
about every original union member; the structural MapLike member is outside
that domain. Replacing only the array return with [] in a measurement control
exposes the narrowed record branch: both backends stop at its for-in receiver,
fixture line 10:23, localRead at locals.go:341. Both arms are NotYet in this
union, rather than refused by TypeScript. No rule was bypassed or edited.

Independently typed controls do lower and run. On native --sanitize and the
JavaScript backend, dense filtering, filtering [undefined, 'x'], empty-array
filtering, and the dictionary loop all match Node stdout/stderr byte for byte
and exit 0. Filter keeps the present undefined. The dictionary returns b,a
in canonical integer-key order. These controls isolate the operations; they
do not claim the original union function is admitted.

The independent sparse-write control compiles on both backends. Both print
only `before-write-length=0`, then exit 70 at a[5] = 'x', with exact stderr:

```text
adamic: panic: index 5 is outside an array of length 0
```

There is no length growth, gap filling or filter execution. Native
adamic_array_set at internal/native/runtime/array.c:504 asks for an existing
slot; line 512 constructs the bounds message and line 513 panics. The
JavaScript backend's adamicSetIndex helper at internal/javascript/javascript.go:111
checks bounds before assignment and gives the same stop. Node instead
grows to length 6, leaves holes, then filter returns one x. Thus the tested
native representation cannot hold the sparse state used by symbolWalker.

The JavaScript backend runs through the normal oracle/node.mjs host runtime.
An initial launch without that host could not resolve package adamic; that
setup attempt is retained and excluded from backend verdicts. Native uses
the ordinary compiler and --sanitize. The archive includes exact sources,
generated JavaScript, compile/run bytes and Lower/runtime sources as .go.txt
where applicable, plus compiler hashes.

## Gate and stop

| Step B condition | Result |
| --- | --- |
| Every argument is an array or plain record | Pass |
| No array has an extra own non-index property | Pass |
| Native sparse witness matches Node, or native cannot hold a hole and zero sparse calls | Fail: 445 sparse calls; native stops at the write |

Stop after Step A. No amendment to 30's adapter or pin is made. No new owner
acceptance, half-owner test, mutation, paired replay, declaration/JavaScript
comparison, 79-root gate, post-amendment lane or 73-root census is claimed.
The measured union-narrowing NotYet stops are an additional current limitation.

## Evidence and reproduction

The getownvalues-count-NN archives hold compact calls.jsonl, piece.json and
shard-evidence.jsonl records with each original report/log path and content,
plus SHA256SUMS. The first piece includes compact preparation provenance.
Every archive has an adjacent SHA-256. getownvalues-closeout.tar.gz retains
the independent reconciliation, frozen source/built manifest, exact-source
comparison, baseline probe and driver/Git logs. The historical measurement
archive is unchanged. Evidence is bounded to fewer than 200 tracked files
including the inherited 99z prerequisite.

To reproduce on this pinned base, source /workspace/adamic-tools/env.sh,
install stage3/api and stage3/typescript with npm ci, and provide the pinned
TypeScript source 050880ce59e30b356b686bd3144efe24f875ebc8. Restore the complete
baseline to /workspace/scratch/gov/before or build it with stage3/apply.sh.
Install that baseline's locked dependencies with npm ci before copying it.
Copy the retained getownvalues-* measurement scripts to /workspace/scratch/gov-a
under their original names. Run prepare.py, move counted/node_modules to
/workspace/scratch/gov-a/modules (the shard product supplies its own pinned
dependency view), then run count-pieces.py. The latter preserves the unchanged
shard runner, bounds each piece, records each observation and commits/pushes
each completed piece. For witnesses, run witness.cjs, build the ordinary
cmd/adamic binary, run run-native.py, then extra-witnesses.cjs and run-extra.py.
reconcile.py checks the completed count from the retained archives.

All command outputs went to files. All processes started for measurement
completed or were explicitly stopped; per-shard child groups are terminated
by the unchanged lane runner and the piece driver.
