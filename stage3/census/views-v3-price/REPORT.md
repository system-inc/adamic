# Views V3 price: incomplete checkpoint

**Lowered totals: not measured. Source ceiling: not certified.** No zero total is implied.

| Measurement | Head | Main |
|---|---|---|
| ViewElement nodes | Not measured | Not measured |
| ViewElementCheck nodes | Not measured | Not measured |
| Units lowered | Not measured | Not measured |
| Source ceiling on main's adapted tree | Not certified | Not certified |

Task: 6wfvfhm. Report only, not for landing.
Box: `4b7d656a5150`, Linux x86_64, 2026-11-06 UTC.
Head: `9626d0113f08269e519ba216ff7f95ff84af16db`.
Comparison main: `2c3f322975b2dee68889e82eab0bbca324d417a9`.
Report branch base (fetched origin/main tip): `12fec995dcadc7f1ae94e4ba63835884451378ae`.
Upstream TypeScript input: `050880ce59e30b356b686bd3144efe24f875ebc8`.

The source-counter draft follows void-intersections/count.cjs's stock TypeScript
6.0.3 program construction and compiler options. It records read forms and
receiver types. It explicitly labels its output provisional. It must not be
used as the requested ceiling without finishing admission equivalence.

The head's exact nodes are `ir.ViewElement` and `ir.ViewElementCheck`.
`viewArrayIndex` constructs the former from `viewArrayElementCheck`; implicit
handouts retain a pointer to the latter in `ElementCheck`. The relevant head
sources are retained here as `.go.txt` evidence.

`unionAggregateContract` sets ArrayReadonly only for the library ReadonlyArray.
It clears the unsupported array-element marker only when
`elementContractCheckable` succeeds. That predicate includes scalar contracts,
unknown elements, readonly nested arrays, union representations, certified
callable producers, recursive object fields and required scalar discriminators.
It also depends on legacy nullable storage and unsupported-contract metadata.
Consequently a readonly-array-name test alone is not the exact admission rule.
The draft leaves representation-sensitive nulls, intersections, tuples and
callables unresolved. Its apparent candidate count is not a proven lower bound
or upper bound of the requested ceiling.

The head also selects checks through a program-wide element key, with
assignability in either direction and a fallback to every element when the
pre-pass cannot decide. It can check reads through mutable receivers that can
hold elements from a viewed readonly array. The source receiver-admission
inventory and the latent lowered-IR inventory therefore need separate labels.

Run `python3 stage3/census/views-v3-price/fixture-test.py` to exercise the draft's
read-form enumeration. The fixture expects seven reads: index, at, for-of,
spread, slice, map and forEach. A plain index write and a mutable-array index are
excluded. The index-skipping mutant produces six and fails the same assertion.
This proves enumeration on that fixture only, not admission-rule equivalence.

The latent overlay addition, known-count lowered fixture, skipped-node mutant,
byte-identity checks, paired closure pieces, pairing and ranking tables have not
been implemented or run at this checkpoint. There are no tsc totals or top-file
rankings to report yet.
