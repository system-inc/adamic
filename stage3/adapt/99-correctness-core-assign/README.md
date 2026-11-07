# assign's ownership-test boundary

Task b5w3ycg. TypeScript 6.0.3 is pinned at
050880ce59e30b356b686bd3144efe24f875ebc8.

The change is `hasProperty<T extends object>(map: T, key: string): boolean`.
Its body is unchanged. assign's signature, loop, reads, writes, and every caller
are untouched. The generic parameter retains the caller's type at the existing
own-property library operation. It avoids the opaque object parameter's native
Object storage requirement without claiming a string index signature.

Adaptation 40 already owns hasProperty. Its hasproperty.cjs and
hasproperty-owner.json are the single planner and whole-function pins. 40 keeps
its original early object spelling so every downstream adapter receives the
same inputs. This directory's adapt.cjs delegates to that owner's final generic
phase, after the other core adaptations. It adds no second owner rule.

The planner accepts only the complete upstream MapLike<any>, previous object,
and final generic owners. Any changed, missing, duplicate, nested, moved or
half-adapted owner is refused before writing. It preserves unrelated edits and
LF/CRLF. A replay writes no bytes and preserves mtime. Whole-core JavaScript is
checked byte for byte with comments kept and removed before any write.

MEASURE.md and CENSUS.md retain the measurement-first evidence. All 79 roots
pass typechecking. All 73 roots formerly stopped at core.ts:1364:29 reach the
existing cast refusal at core.ts:1484:12. The whole lane oracle records 30,782
calls in the accepted traversal, plus one call from a retained hook-timeout
attempt. All 61,570 actual argument appearances are plain Object.prototype
objects, without inherited enumerable properties, own accessors, or symbols.

Construction and composition results are recorded in PROOF.md when complete.
The complete source/build comparisons, paired adapter replay and restoration
mutant are required before this branch is considered finished.
