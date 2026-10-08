Built: constructor-proven Set callable domains; repaired a reproduced literal-domain invariance miscompile and certified 8 pairs / 69 candidate reads toward roadmap step 09.
Commits: first group f339ca8d, requested Set dependencies 360084ad and 6f7902e0; this report travels with the Set-domain delivery commit.
Commands/results: Node, native release, ASan/UBSan, finishing leaks, production C flags and all 60 owned count rows pass; required global counts remains red outside this group.
Mutants: physical-domain substitution in lowering, native metadata and JavaScript metadata each admits the forbidden writer; all eight pair negatives and the two-instantiation control catch all three.
Uncovered: the unit is incomplete; 168 pairs / 2,622 candidate reads remain pending, including non-primitive Set domains and other intrinsic families.

This group certifies a: 3 pairs / 34 reads (ranks 90, 219, 252), b: 1 / 8
(rank 220), c: 4 / 27 (ranks 89, 311, 2237, 2351). Together with the
higher-order group, cumulative certification is a: 3 / 34, b: 3 / 27,
c: 15 / 48, total 21 pairs / 109 reads. set-certificates.json records the
original collection member declarations and pinned share provenance. These are
census candidate reads, not measured upstream executions. Adjacent receivers
and carriers are reduced. Optional receivers are guarded before the cast;
nullable object-view cast transport is not certified by these fixtures.

Observed before repair: new Set<"inside"> viewed through the original
add(value: T): this declaration instantiated at string accepts "outside" and
prints completed with exit 0. The source producer's narrower logical element
domain was lost in physical string metadata. logs/set-before.log records that
semantic failure, not a build or sanitizer failure. This observed miscompile
was repaired before continuing the larger pending families.

Lowering now asks the checker whether the complete instantiated element type
is identical to number, boolean or string. Only those complete domains receive
the existing intrinsic signature certificate. The Set allocation records its
physical element representation separately from its callable element domain;
views never rewrite either. Native constructor metadata and JavaScript's existing
WeakMap preserve the logical certificate. Literals, enums, brands, unions with
undefined and aggregate domains remain uncertified and stop at a demanded
callable read. This is a conservative refusal, not a widened domain or erased
type argument. Fresh results from Set algebra use the same constructor proof.

The Set's existing identity flag remains intact even when its callable domain
is unknown. Its size and physical hashing representation are unchanged. Both
old and current native Map headers measure 200 bytes; the new byte consumes
existing padding. No new allocation, tag, closure ABI or collector is introduced.
Only the necessary shared constructor/emitter/header hooks are edited. None of
the four protected compiler files is edited, no other lane branch is merged,
and no cohere code is copied.

Every pair preserves the original add(value: T): this or
has(value: T): boolean declaration and instantiates its type argument. Each
positive matches source Node in native release, ASan/UBSan and generated
JavaScript, and passes the independent finishing leak check. Every narrow-domain
negative is source-checked and completes on Node, but stops in all compiled
modes with exit 70, empty stdout and a complete pinned expected.stderr message.
The extra invariance writer actually attempts the out-of-domain value. A fresh
Set.union result also passes, exercising the synthesized constructor hook.

The generic constructor control instantiates one make<T> at string and at
"inside" in the same program. Both share string storage, but the first view
call completes and the second stops. All three compiled modes pin the completed
prefix and the second read's full diagnostic. This holds the certificate to
monomorphization as well as direct constructor syntax.

Three actual mutants independently substitute physical storage for the logical
certificate: the lowering proof returns the representation directly; the native
producer records its physical element; the JavaScript emitter supplies the
physical element to its WeakMap. Each mutated backend finishes the forbidden
writer, each of the eight pair negatives fails its pinned stop assertion, and
the generic control admits both instantiations. The runner rejects invalid C,
Go build failures and sanitizer crashes as evidence and restores every source
in a finally block. Full mutation evidence is in logs/*set-domain-mutant.log.

Commands, with test output written directly to log files:

```sh
source /workspace/adamic-tools/env.sh
python3 stage3/interface-downcasts/lane5/code/run-set-mutants.py > /tmp/lane5-code-set-mutants.log 2>&1
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 10m -args -update-counts > /tmp/lane5-code-set-counts-global.log 2>&1
go test ./internal/oracle -run '^TestCheckedViewCallableCodeCounts$' -count=1 -timeout 5m -args -update-counts > /tmp/lane5-code-set-counts-owned.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewCallableCodeHigherOrder$|^TestCheckedViewCallableCodeCallbackExecution$|^TestCheckedViewCallableCodeSetDomains$|^TestCheckedViewCallableCodeSetPrimitivePairs$|^TestCheckedViewCallableCodeSetInstantiations$|^TestCheckedViewCallableCodeCounts$|^TestCheckedViewSetIntrinsics$|^TestCheckedViewSetIntrinsicSignatures$|^TestCheckedViewSetIntrinsicsC$' -count=1 -v -timeout 10m > /tmp/lane5-code-second-group-final.log 2>&1
git diff --check
```

Final focused oracle: PASS, 40.587s, with native cache hits 0 and all 60
recorded count rows checked. This group adds 19 rows, changing none of the first
group's 41 rows. The scoped updater passes in 7.816s. The required global
updater exits 1 in 49.878s on inherited failures; four representative failures
were already reproduced on the unchanged pinned base in the first group.
The whole table is not certified green. No whole-package test or full gate is run.

Setup evidence remains in HIGHER_ORDER_REPORT.md and logs/setup.log: done
84.041s, nproc 5, cpu.max 400000 100000. The pending inventory remains explicit
and sorted by read count in inventory.json: predicates 553 reads, generic
instantiation 474, rest intrinsics and producer writes 462, overloads 352,
followed by the remaining receiver and carrier groups. The known scalar-array
literal write frontier is still listed as pending, not silently counted closed.
