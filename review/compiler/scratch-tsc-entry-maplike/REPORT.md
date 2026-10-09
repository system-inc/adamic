Merged MapLike records into the project-aware stricter-options candidate for a step 32 scratch run.
Base 9128a35fbf1d9a724752e240704b2c198508c946; merged compiler/records-maplike 49319c49; scratch delivery SHA is in the handoff.
CLI build succeeds; the identical native entry commands with split 0 and 1 both exit 1 at core.ts:332:1.
No mutants were run and no new checking rule was introduced; this only resolves the requested merge and observes its stop.
This branch is not a landing candidate; no lane checks, counts, fixture suites or native tsc execution were run.

Both layouts pass the previous corePublic.ts:14:5 MapLike signature and stop at
src/compiler/core.ts:332:1. The construct is:

```
export function* mapIterator<T, U>(iter: Iterable<T>, mapFn: (x: T) => U): Generator<U, void, unknown> {
    for (const x of iter) {
        yield mapFn(x);
    }
}
```

Exact refusal, identical in both runs:

```
adamic: /tmp/projects-tsc-adapted/src/compiler/core.ts:332:1: Adamic 0.1 refuses a generator function; use an explicit iterator object; suspended frames need ownership and cancellation rules before generators can be compiled without a collector (docs/user-iterators.md)
```

Commands, using the same adapted tree and pinned Node declarations as the prior
combined run, with no source edits or diagnostic overlays:

```
source /workspace/adamic-tools/env.sh
go build -o /tmp/projects-combined-adamic ./cmd/adamic > /tmp/scratch-maplike-build.log 2>&1
ADAMIC_NATIVE_SPLIT=0 /tmp/projects-combined-adamic build /tmp/projects-tsc-adapted/src/tsc/tsc.ts -o /tmp/projects-tsc-0 > /tmp/scratch-maplike-entry-0.log 2>&1
ADAMIC_NATIVE_SPLIT=1 /tmp/projects-combined-adamic build /tmp/projects-tsc-adapted/src/tsc/tsc.ts -o /tmp/projects-tsc-1 > /tmp/scratch-maplike-entry-1.log 2>&1
```

The scratch starts directly from 9128a35f and merges the pinned 49319c49 rather
than another branch tip. Conflict resolutions retain the MapLike branch's
mutable record IR, lowering, prototype restrictions and JavaScript operations.
Obsolete readonly-record dispatch hooks are removed where the incoming record
dispatch supersedes them. The readonly recognition helper used by the stricter
enumeration proof is retained. Combined nullable and caught-value handling stays
in place; detached own-property aliases retain the incoming record behavior.
The binary-operator refusal exceptions include both existing nullish comparisons
and incoming record membership. The existing representation enum is retained;
a duplicate C Record case and a duplicate enumeration helper file are removed.
Overlapping test expectations and counts keep the combined side; conflicting
record fixture status descriptions keep the MapLike side. These metadata choices
were not validated by fixture suites on this scratch. The full conflict list is
in evidence/merge-conflicts.json. Newly imported cloud/reports evidence is moved
under this scratch's review directory.

The generator stop was not fixed. docs/user-iterators.md explicitly distinguishes
explicit iterator support from the missing owned suspended-frame implementation.
Published branch names were inspected with:

```
git ls-remote --heads origin 'compiler/*' 'codex/*'
```

No obvious published generator implementation branch was found.
codex/library-iterators e52168ca and codex/library-overlay-iterators 24d980a7
are related iterator topics, but their names do not establish generator support;
they are not asserted as owners of this stop. The source-adaptation work belongs
to TypeScript's step 32, whose published entry topic is
codex/stage3-tsc-entry-build 2b490213. That is an adaptation handoff, not evidence
that the topic already rewrites mapIterator. No additional branch was merged
and no next-stop implementation was changed.
