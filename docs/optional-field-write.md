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

## Final batch 3 validation

Based on a37ebdb0913eca9d3d6e8adbe8f01bb733da52eb, with implementation
checkpoints 5bb775ca and 03980678 pushed to codex/optional-field-write-2.
No current-main merge was substituted for the requested batch 3 base.

After `export GOPROXY='https://proxy.golang.org|direct'`, cloud/setup.sh passed:
Node ready 0.024s, Go ready 0.034s, submodules ready 0.051s, clang ready 0.229s,
markdown dependencies installed step-duration 0.733s / ready 0.831s, go build
ready 38.906s, test binaries deferred 38.993s, cache warm 38.994s, done 39.018s.
`nproc` reports 5; cgroup cpu.max is 400000 100000. Node v24.19.0,
Go 1.27.1 and clang 20.1.8; source /workspace/adamic-tools/env.sh before commands.
Setup ran successfully while the initial recursive fetch was still on main
71d7e491. The recursive fetch stalled on unrelated submodule history; explicit
nonrecursive refs and shallow exact submodule commits supplied the batch 3 pins.
The first test attempt exposed that pin mismatch, then missing @types/node in
stage3/api. Exact pinned submodule updates and `npm ci --prefix stage3/api`
resolved those environment failures. No cohere source was copied.

The final source-stable gate wrote to /tmp/optional-field-packages-final.log:

```text
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./internal/lower ./internal/native ./internal/javascript ./internal/ir ./internal/oracle ./stage3/fixtures
ok internal/lower 73.328s
ok internal/native 238.801s
internal/javascript: no package tests (backend checked by oracle)
ok internal/ir 12.266s
ok internal/oracle 299.166s
ok stage3/fixtures 34.364s
```

This is the complete oracle, including every predecessor optional-field fixture,
all batch 3 readiness/import-cycle fixtures, and their stored mutations. It uses
source Node, backend Node and the oracle's usual native release, sanitizer and
Linux leak builds. Existing inserted readiness checks are intentionally pinned
checked stops, rather than raw source-Node equality. In particular,
optional_field_checked_copy.a stops on its present unreadied field in both
backends, while Node continues. Its earlier successful checked copy preserves
absence. The independent C probe also checks fixed and dynamic copies with an
absent unreadied slot, raw copies carrying unreadied present storage, unknown
reads, and subsequent writes. The unknown-view .a fixture observes both optional
presence and class field readiness, including the distinct empty-spread shape.

Final Linux regeneration:
`go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 30m -args -update-counts`
passed in 31.968s; /tmp/optional-field-counts4.log. The complete final gate also
rechecked recorded counts. New A/F/R/L/P/G rows are 2/2/0/2/1/0 (write),
45/45/40/59/11/0 (presence), 87/86/62/134/25/1 (construction),
22/22/34/63/7/0 (unknown), and 5/2/6/8/4/0 (checked copy, intentional stop).
Existing rows are unchanged from the requested base.

`go test ./internal/oracle -run '^TestOptionalField' -v -count=1 -timeout 30m`
passed in 3.832s; /tmp/optional-field-mutants-final.log. Every requested mutant
was caught: dropped absent slot restores the missing-field compiler-bug stop
(exit 70, source Node disagrees); dropped copy presence changes state/unknown
assertions; dropped copy readiness changes state/unknown assertions; overlapping
tails stops on the independently ready field. The three copy-layout mutations
are run against both fixed and dynamic shapes, release and ASan/UBSan builds.
Additional stored mutations caught initial presence, static layout insertion
order, ineffective deletion (all wrong stdout with exit 0), and missing spread
reservation (missing-field stop, exit 70). The empty-shape metadata overlay was
caught by Node/native disagreement; /tmp/optional-field-empty-metadata-mutant.log.
The conservative Object.assign overlay was caught by the diagnostic assertion,
not a runtime miscompile; /tmp/optional-field-assign-proof-mutant.log.

Stage 3 update and byte audit passed; /tmp/optional-field-stage3-update.log and
/tmp/optional-field-stage3-byte-audit.log. Only two stage0 values changed,
NotYet to Compiles. There are no Compiles regressions. All bytes outside stage0,
including recorded Node observations, remain exact. The final gate validates
those records without an updater overlay.

Whole-repository `go vet ./...` exited 0; gofmt -l cmd internal and git diff
--check printed nothing. Logs: /tmp/optional-field-vet-final.log,
/tmp/optional-field-format-final.log, /tmp/optional-field-whitespace-final.log.
No lines in internal/native/emit.go, internal/lower/lower.go,
internal/native/native.go or internal/oracle/oracle_test.go changed.

Not covered: the entire repository/stage-1 test gate, Darwin or WASI execution,
performance measurements, optional boolean/mixed-union layout extensions,
arbitrary dynamic keys, and retroactive reservation through widened aliases.
Developer-tools generated-program reruns were not run, as requested.

## Alias enumeration follow-up, October 8

The probe at codex/views-lazy-admission a0cc0afd4 exposed a wrong-output exit-0
case on 86b3fe99 that the earlier fixture set missed. Its presence/alias.a is
preserved verbatim as internal/oracle/testdata/optional_field_alias.a. Node and
backend JavaScript omit the deleted key; native's plain-object branch of
adamic_class_object_keys enumerated the descriptor without consulting presence.

Plain objects now delegate to adamic_object_keys even when emitted through the
class/ObjectKeys path. There is no binding-local static key exemption: runtime
ranks determine absence and insertion order through all aliases. Class public
and static descriptors retain their existing path. Plain quoted # names remain
public; the shared key enumerator must not mistake them for private identifiers.
The existing class_features_private.a and library_for_in.a controls hold this.

Object.hasOwn on plain-data-literal interface annotations and const aliases is included
as a small fix. Present plain-object, declared literal-key and private-name checks
remain, with NUL names refused. General parameters and constructor-object views
remain NotYet: inherited constructor storage and getter descriptors need separate
own-presence handling. Getter origins are excluded even when an annotation hides
their accessor declaration.
Values/entries still require their exact shape and homogeneous representation.
Homogeneous number | undefined values/entries are admitted using the existing
packed-number runtime path so the requested entries control can preserve present
undefined. Optional fields and hidden heterogeneous views remain refused there.

The variants exercise deletion through an optional-annotated literal binding
observed through its alias, entries and values after alias deletion, spread after
deletion, present undefined after reinsertion, insertion order with a second key,
and hasOwn through an interface on initially absent and later present undefined.
A non-optional literal field cannot itself be deleted under this checker:
TS2790 requires an optional operand. The direct-binding control therefore declares
its literal binding optional, with no checker option changed.

TestOptionalFieldAliasCatchesStaticEnumeration restores the previous plain-object
static descriptor list. Source Node catches its exact original wrong stdout:
false / [first] / false, then true / [first] / true. Both release and ASan/UBSan
mutants exit 0 with empty stderr. Focused lowering/oracle and mutant checks passed
(/tmp/optional-alias-focus3.log); the final variants passed 0.538s; the quoted-#
and alias controls passed 16.491s (/tmp/optional-alias-prefix-control.log).
Linux counts regenerated in 39.072s (/tmp/optional-alias-counts-complete.log),
adding only two rows: 10/10/9/20/6/0 and 69/69/29/96/16/0. Existing rows are
unchanged. Stage 3 update check passed 19.835s with no record changes.

The complete uncached lower/native/oracle/stage3 gate is running on fixed sources
at this checkpoint. Vet, formatting and whitespace checks pass. No protected
emit.go/lower.go hooks were needed. Full repository/stage-1 tests, Darwin/WASI
execution and performance measurements remain outside this follow-up.

A constructor boundary probe caught overbroad hasOwn admission during this
follow-up: Object.hasOwn(Child, 'first') for an inherited static field produced
native true versus Node false. The final admission therefore uses the existing
literal-origin proof for annotated plain values; constructor views and unknown
parameter origins retain their previous NotYet outcome. The empty interface
literal and both alias directions still compile and agree with Node. Logs:
/tmp/optional-alias-static-node.log and /tmp/optional-alias-static-native.log.
