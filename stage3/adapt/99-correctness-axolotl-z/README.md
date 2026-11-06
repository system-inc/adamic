# 99z: elementAt keeps its existing fallback

Status: count-authorized fallback candidate implemented; validation in progress.
The all-class witness-equality requirement remains open: the unchanged fallback
differs from native at for fractional offsets and NaN. No equal result is claimed
for those inputs. The count and complete lane coverage are in COUNT-ELEMENTAT.md.

## Ruling and ledger entry

Status: JavaScript-changing adaptation under Adamic's count-based road (b),
task b5w3ycg, Nov 6 21:49Z. Lib es2022 is refused. The completed count before
editing is 1,936 calls over all 398 lane shards, with zero non-integer offsets.
The unchanged whole stage3/lane/run.sh control passes and has the exact same
test-identity multiset and baseline diff as the shard union.

Upstream: TypeScript 6.0.3, commit 050880ce59e30b356b686bd3144efe24f875ebc8.
Location: core.ts:1076, elementAt's exported const initializer.
Adamic's earlier lib ruling, task axochk, Nov 1 16:06Z: "tsc is checked against
the lib its own tsconfig declares." The upstream lib remains es2020.

Our patch removes the feature-test conditional and native-call arm, retaining
only the existing fallback arrow and its exact body. No truncation, new cast,
new local or lib change is added. Two blank source lines preserve coordinates
for later refusal evidence. Both original as any casts disappear.

The exact complete before and after variable statements and SHA-256 pins are
in owners.json. A missing, duplicate, nested, changed or half-adapted owner
refuses before writing. LF and CRLF survive, and replay makes zero writes.
The new adapter reports javascriptChanging true. No other initializer is edited.

Named behavior difference, outside the observed call domain: with [10,20],
native at returns 10 for 0.5, -0.5 and NaN; the unchanged fallback returns
undefined. The observed offsets are only -1, 0 and 1. The requested equality
witnesses for absent non-integer cases cannot all be produced. Their actual
results will be retained, not hidden by an empty-array example. The pending
clarification concerns this witness requirement; it does not change the exact
source rewrite authorized by the zero count. No issue has been sent upstream.
