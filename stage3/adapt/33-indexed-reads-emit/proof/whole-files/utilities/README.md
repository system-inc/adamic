# transformers/utilities.ts: 3 -> 3

The current 886-line source is byte-identical to the previously reviewed wave.
No source edit is made. Each TS2375 is recorded individually in declines.json
and in the complete before/after diagnostic reports. All three object producers
intentionally retain parameters: undefined when parameter decorators are absent
or the non-legacy decorator mode is selected. Ordinary decorated class, getter
and method Node probes demonstrate the own undefined property in all three cases.
An assertion or omission would lie about presence or change the object's shape.

The truthful fix is `| undefined` at AllDecorators.parameters in
src/compiler/types.ts:5828. That declaration is outside this worker's territory.
No outside file or declaration is edited and no consumer cast masks the owner.

All-code pinned latent census **3 -> 3**. Adapter site contract and idempotence
pass. The source is unchanged, so there is no new source mutant; all prior
required-read and API mutants remain retained in their file proofs.
Default oracle: 106367 passing, zero failing/pending,
empty baseline diff, 218.69 seconds. API projection is exactly the
approved 189+28 lines, with every other reference byte/file identical.
The checker-clean total remains 39/78; this file is not claimed closed.
