The following predecessor report describes 514b936; the batch 3 integration and
its validation are recorded below.

Optional fields created absent now reserve native storage for later writes.
Branch: codex/optional-field-write, based on origin/main c01907a.
Validation: touched packages, Node differential oracle, counts, go vet, and formatting.
Mutants: dropped slots/reservation, presence/order/delete, reference release, getter order.
Limits: no full repository or stage-1 gate, no performance measurements.

The literal previously used only its written properties for its native shape. Contextual declared optional fields now supply additional storage, initialized to undefined and marked absent. Spreads reserve those fields while preserving the source shape. Contextual Object.assign target literals and annotated function returns use the same rule. Existing object writes mark storage present. Deletion clears presence and releases stored references; rewriting appends the key after existing ordinary string keys. Numeric keys retain Node's enumeration order. Presence ranks share the object's allocation; there is no separate allocation or global shape cache.

Three .a fixtures compare native release, ASan/UBSan/LeakSanitizer, and JavaScript output with source execution on Node. They cover the reduced program, reads before writes, keys and in before/after writes, deletion/reinsertion, runtime-created strings and objects, function returns, empty/populated/undefined spreads, Object.assign targets, explicitly written undefined, and numeric-key getter evaluation order.

Commands (source /workspace/adamic-tools/env.sh first):

- bash cloud/setup.sh: go 0s, clang 1s, node 1s, submodules 1s, cache 125s, done 125s. nproc: 5; cgroup quota: 4 cores.
- go test ./internal/lower ./internal/native ./internal/javascript ./internal/ir -count=1 -timeout 30m: passed. JavaScript has no package tests; its output is checked by the oracle.
- go test ./internal/oracle -run '^TestCountsAreRecorded$' -update-counts -count=1: passed. Only the three new rows were added; existing recorded counts are unchanged.
- go test ./internal/oracle -count=1 -timeout 30m: passed in 153.006s, complete oracle package including stored mutants and allocation counts.
- go test ./internal/native -count=1 -timeout 30m: passed in 153.322s, final runtime validation.
- go vet ./... and gofmt -l cmd internal: passed; formatting output empty.
- go -C /tmp/adamic-optional-tools run ./cmd/adamic-fuzz -root /workspace/adamic -with optional-field-write -seed 1 -count 200 -shrink=false -work /tmp/optional-fuzz-final-work: 200 programs in 2m24s, agreed 187, findings 0, checked 13, not yet 0, invalid 0, unfit 0. Tool checkout: c91c20771, devtools/generator-undefined-numbers.

Mutant observations:

- Replace optional lookup with required lookup in the existing absent-field fixture: missing-field stop, exit 70; Node disagrees.
- Drop the reduced program's declared absent slot: original missing-field compiler-bug stop, exit 70; Node disagrees.
- Drop spread reservation: missing-field stop, exit 70; Node disagrees.
- Remove initial absence marking: Node stdout comparison catches premature key presence.
- Use static layout index as insertion rank: Node stdout comparison catches wrong key order.
- Replace delete with hasOwn: Node stdout comparison catches unchanged presence/value.
- Omit reference release during deletion: LeakSanitizer reports 204 bytes in three allocations, while ordinary stdout still agrees. Runtime mutation restored afterward.
- Visit spread getters in layout order: Node prints twoten; native mutant prints tentwo. Runtime mutation restored afterward.

Detailed logs are under /tmp/optional-*.log in the worker workspace. The initial full oracle run exposed eager accessor evaluation in the first copy implementation; that was corrected before final validation.

This unit handles the supported number, string, and object optional representations of plain contextual literals. It does not extend optional boolean/mixed-union representations, arbitrary dynamic keys, class/accessor construction layouts, or retroactively reserve fields when an already-created object is widened through an alias. Presence operators are limited to named declared optional plain-object fields; inherited Object prototype names remain NotYet. Per-slot ranks add one size_t of storage to all native objects; memory and speed impact has not been benchmarked. The full repository/stage-1 gate was not run.

## Batch 3 layout decision

The batch 3 readiness tail and presence ranks now have separate space in one
allocation: object header, value slots, size_t insertion ranks, then one readiness
byte per slot. Ranks are aligned by the value array. Dynamic shapes begin after
padding the complete tail to _Alignof(adamic_shape); names, int type tags and
reference flags follow their descriptor. Both normal and region allocations
reserve the entire tail.

I weighed a byte with separate ready and present bits. It represents those two
states, but cannot represent insertion order after arbitrary deletion and
reinsertion. Keeping the original ranks avoids a bounded rank encoding, an order
renumbering algorithm, or a second order mechanism. Zero rank means absent and
readiness remains independent. This is the smallest change to the two existing
designs, not a claim of minimum memory use.

Checked copies keep rank and readiness using the actual source cache index.
Absent slots are reserved and copied without a checked read; present unreadied
fields still stop before copying, with the existing diagnostic. Newly reserved
slots start absent and ready to hold undefined. Writes publish presence and
readiness; explicit deinitializing writes publish presence and clear readiness.

Dynamic descriptors own their scalar type tags in the object's allocation.
Static metadata stays in the existing static registry. Dynamic metadata is never
registered there, so freeing a copied object cannot leave a dangling registry
entry. Unknown views use ranks for presence and return undefined for unreadied
storage, matching JavaScript's uninitialized field value. Typed checked reads
retain their existing stop. Dynamic descriptor cache hits are disabled because
an allocator can reuse an old descriptor's address for a different shape.

Integration checkpoint: the focused lowering/native/oracle tests passed
(`/tmp/optional-field-focus3.log`), Linux count regeneration passed in 30.779s
(`/tmp/optional-field-counts2.log`), and the independent fixed/dynamic copy-state
baseline and mutants passed in release and ASan/UBSan builds in 2.032s
(`/tmp/optional-field-layout-mutants3.log`). Dropping presence changes `in` and
unknown reads; dropping readiness changes the copied bit and unknown reads;
overlapping the tails triggers a read-before-assignment stop on the ready field.
The full touched-package and stage 3 gate is running at this checkpoint.
No changes were needed in internal/native/emit.go or internal/lower/lower.go.

Second checkpoint: reserved writes through function parameters now use Node parity
in the former missing-slot safety fixture. Unknown absent-spread presence is a
positive lowering test. Empty undefined-spread shapes register their scalar type
metadata; the additional unknown-view probe uses a distinct shape to exercise
that registration. Removing it is caught by source Node versus native (exit 70,
`dynamic read of a host scalar without type metadata`). Fresh Object.assign's
optional surface exemption is conservative: direct literal arguments without
spreads. Hidden-source tests keep their optional-widening refusal. Removing that
restriction changes the refusal path to the existing Object.assign shape guard;
this is diagnostic-path evidence, not an additional runtime soundness claim.

Linux counts regenerated again in 31.968s. No existing batch 3 counts row changed;
five new fixture rows were added. The final optional mutant suite passed in
3.832s. Stage 3 regeneration passed in 35.777s: objects/08_loop_state.a and
objects/12_decorator_descriptor.a changed NotYet to Compiles. There are no
Compiles-to-Refused/NotYet regressions. A byte audit verifies that everything
outside each stage0 value, including Node observations, remains unchanged.
The diagnostic updater was a scratch Go overlay; the harness is unchanged.

The first broad gate is discarded: its compiler binaries preceded the last
metadata change while fixture sources changed, and it also found two obsolete
unsupported-behavior assertions. The corrected focused regressions pass (native
0.556s, lower 0.411s). The final broad gate is running with sources held fixed.
