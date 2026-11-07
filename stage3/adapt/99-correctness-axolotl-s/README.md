# Generic clear amendment

Task sdhn5pe amends 99s's whole-owner pin to `export function clear<T>(array: T[]): void { array.splice(0); }`. Each caller infers its concrete mutable element type, retaining storage information for Lower. The splice body and every caller are unchanged. This amendment is type-only and adds no ledger line. Fully adapted previous 99s states migrate as a whole; mixed or changed owners are still refused before any write. See CLEAR-GENERIC-PROOF.md for current evidence.

# 99s: boxed queue amendment, wave row 8

Task b5w3ycg. Finished. This amends 99s because its pin covers the entire createQueue owner; no new adaptation directory is created.

Rule class: JavaScript-changing, count-authorized established shape. Adamic, task b5w3ycg, Nov 7 03:20Z: established-shape adaptations, Compiler accepts.

The buffer holds boxes or cleared slots. for...of boxes sparse holes as undefined payloads. enqueue boxes each payload; dequeue checks box occupancy, clears the slot and returns box.value without a cast. isEmpty and 99s's splice truncation are unchanged. Explicit undefined remains a valid payload.

The whole lane and all 398 bounded units each compare 10041 enqueues, 10123 dequeues and 13660 public isEmpty calls, plus 20164 post-operation emptiness comparisons, with zero differences. Sparse initial arrays and own iterator overrides each count zero. Node sparse and undefined-payload witnesses match stock. A deliberately changed iterator detector seed differs outside the observed own-iterator class, whose real count is zero.

Both full applies, paired replay of every adapter, all 79 roots of adamic types, both public declaration comparisons, owner-only JavaScript checks and the uninstrumented complete lane pass. The full core.ts census removes exactly F15/F16/F17, adds no site and restores all three under the whole-owner mutant. Generic body coverage remains deferred.

See [QUEUE-BOX-PROOF.md](QUEUE-BOX-PROOF.md), [the count table](queue-box-count-table.json), [all hunks](QUEUE-BOX-JAVASCRIPT.md) and [structured results](queue-box-results.json). Raw evidence is packed with SHA-256 manifests. The original record below is historical, including its rejected map-based draft.

## Original 99s record (historical)

Status: finished under Adamic's follow-up ruling. PREVIOUS-STOP.md retains
the earlier stop as historical evidence. See PROOF.md and CENSUS.md for final
validation and actual classifications.

Base: b729c7e07010fe6db34c6afc1d11c8ca63a86ae3 (99q).
Read first: core-walls-sort at ad29102e9c145c5b480629227574d2ffa235a4b1,
REPORT.md rows CW005, CW006, CW057 and CW074. The later task ruling authorizes
truthful skipWhile types and splice instead of length assignment.

The built adapter changes only skipWhile, filterMutate, clear and createQueue.
The three splice statements change JavaScript. skipWhile's three return types
and removal of its suffix cast are type-only. The dequeue cast remains unchanged;
the alternative representation below is a rejected draft, not a built edit.
No deduplicateSorted, createSortedArray, toSorted, concatenate or relativeComplement
owner is edited. The real apply.py lexical directory sort places 99s immediately
after 99q on this base; 99r is absent. sort.json records that result.

AST owner lookup uses stock TypeScript 6.0.3. Complete exact before and after
owner texts and SHA-256 hashes are in adapt.cjs and owners.json. No whole-file
hash gates adaptation. All owners are validated before any write. Changed,
half-adapted, missing, duplicate and nested owners are refused. Replay makes
zero edits and preserves bytes and mtime. contract.json records the checks.

## skipWhile caller audit

find-references.cjs uses the TypeScript language service findReferences API on
all files under the adapted src tree. references.json retains the complete
result, including declarations and the import alias. There is one call:

| Caller | Use | Reads the exclusion narrowing? |
| --- | --- | --- |
| src/compiler/transformers/ts.ts:1089, injectClassTypeMetadata | `addRange(modifiersArray, filter(skipWhile(modifiers, isExportOrDefaultModifier), isModifier))` | No. filter's isModifier guard narrows the suffix independently before addRange. |

The import at transformers/ts.ts:183 is not a call. No caller check is added.
For `[1, "x", 2]` and the number guard, the returned suffix is `["x", 2]`.
The later number disproves the old Exclude result. The types and runtime
counterexample, patch, and drafted upstream issue are also in upstream/LEDGER.md.

## Node splice witnesses

witness.cjs executes the real before and after owners, plus the exact queue
copyWithin/truncation fragment. witness.json records every input, result and
length. From `[1,2,3,4]`, filterMutate returns `[]` length 0, `[1,2]` length 2,
or `[1,2,3,4]` length 4 before and after, according to the predicate. The queue
fragment returns `[]` length 0, `[3,4]` length 2, or `[1,2,3,4]` length 4.
clear always targets zero: empty, two-element and four-element arrays all
become `[]` length 0 before and after. It cannot target a positive n; its empty
case is also the own-length no-op case.

No splice result is consumed. filterMutate and clear still return undefined;
dequeue still returns the saved payload, including undefined. Full queue tests
exercise compaction, FIFO, empty errors, sparse initial input and immediate
clearing of the consumed slot's reference.

## Open dequeue cast site CW057

The smallest boxed-slot draft checked here is cast-free and stock-checker clean,
including noUncheckedIndexedAccess. It checks occupancy, never the payload.
An explicit queued undefined is accepted. A consumed slot is set to undefined,
so it stops referencing both the box and its payload.

It fails the required behavioral equivalence test: with `new Array(1)` as the
initial input, slice preserves the hole and the original queue returns undefined.
map also preserves the hole, but the draft's occupancy check then throws
`Queue invariant violated`. dequeue-draft-witness.json records both outcomes.
Therefore this representation is not built, as reaffirmed by Adamic. The independent
`elements.splice(newLength)` truncation is built. The original unproven generic cast
remains, awaiting a separate ruling or representation that also preserves sparse
input behavior. A payload `!== undefined` check would be incorrect for T that
includes undefined. No replacement cast is introduced to hide the problem.

The complete adapted draft follows. It is retained separately as dequeue-draft.ts
only for reproducible isolated type and Node checks, not applied to upstream.

```ts
interface Queue<T> { enqueue(...items: T[]): void; dequeue(): T; isEmpty(): boolean; }
export function createQueue<T>(items?: readonly T[]): Queue<T> {
    const elements: ({ value: T } | undefined)[] = items?.map(value => ({ value })) ?? [];
    let headIndex = 0;
    function isEmpty() { return headIndex === elements.length; }
    function enqueue(...items: T[]) {
        for (const value of items) elements.push({ value });
    }
    function dequeue(): T {
        if (isEmpty()) throw new Error("Queue is empty");
        const slot = elements[headIndex];
        if (slot === undefined) throw new Error("Queue invariant violated");
        const result = slot.value;
        elements[headIndex] = undefined;
        headIndex++;
        if (headIndex > 100 && headIndex > (elements.length >> 1)) {
            const newLength = elements.length - headIndex;
            elements.copyWithin(0, headIndex);
            elements.splice(newLength);
            headIndex = 0;
        }
        return result;
    }
    return { enqueue, dequeue, isEmpty };
}

```

## Direct concrete mutant proofs

The meter does not instantiate filterMutate. direct-mutant.cjs instead copies
its exact adapted body into a minimal program, calls it with T=number and compiles
it with this tree's ordinary compiler. The control changes only the splice back
to length assignment. The splice program lowers; its restored control is NotYet
at filterMutate-mutant-length.ts:9:5: "assigning a field of a value".

clear's copied body uses a concrete number[] parameter for this minimal storage
probe. Both clear controls have that same signature and call; only truncation
differs. The splice program lowers; its restored control is NotYet at
clear-mutant-length.ts:2:5 with the same reason. This does not claim that the real
unknown[] clear owner has lowered: its separate erased-storage wall is recorded
in the census. The first un-specialized attempt refused the mutable number[]
to unknown[] relation; clear-erased-attempt-*.log retains that result.

direct-mutant-result.json records exact diagnostics, exits, source hashes and
the ordinary compiler hash. The C output and diagnostics are retained in the
corresponding *-mutant-*.log files. No compiler or meter diagnostic is relabeled.

Current main says NotYet for length assignment, implemented by class.go:471.
The ruling in docs/rulings.md:285 and :439 says Refused; enforcement is owed by
#53r37a3. Adamic's follow-up authorizes the three splices despite that mismatch.
The compiler and ruling remain distinct observations in CENSUS.md.
