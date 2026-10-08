Built checked singleton numeric-tag views, full stored object remainders for checker-admitted reads, and implicit switch fallthrough; structured/optional views remain incomplete.
Commits: step zero `7b2579adb7cadedcf3c9292d74a9f71a67dfe82b`; implementation `720849c8c22cf9c2842cae2670adce128196be0f`; computed-name repair `488e6a050ac5f5a5d231b81724d819cbed19ee4c`; documentation commit is the branch tip.
Commands/output: main negative control fails as predicted; full lower, IR, native and uncached oracle pass; meter main 966 -> 101 -> 0, area 1182 -> 101 -> 0 in the old row, with 90 checked-view NotYet sites remaining on each input.
Mutants: eleven isolated compiler overlays and two permanent runtime mutations are caught; the step-zero open_tag guard mutation was also caught before replacing that refusal with checks.
Not covered: general structured/optional payload views, direct never-property access rejected before lowering, arbitrary dynamic layouts, or the full repository gate.

Subsequent follow-up: [both census panic paths are fixed and pinned](enum-census-panics.md).
The counts below remain the historical evidence for this implementation checkpoint.

## Step zero and the corrected record

The negative control uses the requested string `kind`, identical `flags: Flags`
in both interfaces, and reads `v.n` in the `kind === 'b'` branch. On pinned main
48c05d09 it fails with the original `adamic/enum-tag` refusal at main.a:1:258.
This is an observation of the old field loop's over-breadth, with no native
build. On the corrected branch the same probe passes.

The checkpoint requires the candidate field in every union variant, a unit
type in at least one, differing field types across variants, an open enum, and
a changed observed type. The latent census at that checkpoint is main 101 and
area 101 rather than 966 and 1182. Thus **865 of main's 966** and **1081 of
area's 1182** were metadata false positives under this discriminant test.
The remaining **101 on each input** are actual discriminant candidates under
that definition, not a manual semantic classification of the entire compiler.
The checkpoint was committed and pushed separately before this implementation.

The old `open_tag` regression still failed if the checkpoint's discriminant guard
was removed: expected refusal, got nil. The final implementation replaces its
refusal with checked IR. `TestNumericEnumOpenTagIsChecked` asserts that a check
exists; pinned runtime tests and `TestEnumTagPayloadMutant` establish that it
can fail in both backends.

## Behavior and evidence

The existing SyntaxKind fixture compares member constants and FirstX/LastX
aliases by numeric value with `===`, `!==` and switch cases, through literals,
an exact class and a checked cast. An out-of-range tag takes the explicit
default and runs ordinary code. Openness never supplies an exhaustiveness proof.

New witnesses cover string versus number, boolean versus number, literal string
payload values, singleton member casts, nominal class views, and nested property
reads from different receivers. Native checks actual field layout before reading
untagged storage; JavaScript checks the runtime value. Number and boolean shapes
with identical reference bitmaps are distinguished. A valid copied class with
private storage and extra public fields completes normally.

Written singleton member annotations remain literal promises even though the
checker represents the member and whole enum with the same type. Ordinary
construction and structural views require proof; a cast inserts a value check.
Mutable container views preserve that promise in both directions.

The remainder fixture uses defaults, else paths, aliases, and implicit void
switch fallthrough. It observes the entire original object union after every
declared tag value has been excluded. A narrowed destination that accepts only
one variant is refused. Separate holders sharing a property declaration do not
share path exclusions. Object and scalar `never` assertions remain checked sites.
The object-default witness prints `before` and `default`, then exits 70 with:

```text
adamic: panic: unreachable value 42 for numeric enum AKind
```

The original SyntaxKind assertion pins the analogous message ending
`numeric enum SyntaxKind`. Source Node erases the assertion, prints
`unreachable` and `after`, and exits 0; both compiled backends deliberately stop.
Checked payload witnesses similarly observe Node's erased type behavior and pin
Adamic's required stop, rather than pretending they agree after a violated view.
Finishing witnesses compare with Node in release, ASan/UBSan and leak checks.

An unmatched implicit numeric switch falls through unless it would end a
function requiring a result. That edge has an explicit checked missing-result
site with message `numeric enum switch fell through a function requiring a result`.
The scrutinee is evaluated once. Whole-enum remainder arithmetic now executes
normally: the update witness observes 43 after incrementing 42, then returns 99.

## Limits, observed separately from inference

90 of the 101 checkpoint locations now report `NotYet` for incompatible
structured or optional payload fields; the final JSON records the exact reasons
and location correspondence. The old refusal row is zero, but these 90 sites
have not been implemented. Of the other 11 checkpoint locations, eight report the existing refusal
`a cast the runtime can't check`; three have no finding at that exact location.
Neither outcome establishes successful declarations or a successful compiler build. All measurements are on checker-rejected inputs.

A direct `v.kind` in a default after `Kind.A` and `Kind.B` returns is rejected
by the source checker with TS2339, `Property 'kind' does not exist on type 'never'`.
The probe is `/tmp/enum-never-property-probe.a`, its observed CLI failure is in
`/tmp/enum-never-property-probe.log`. Lowering restores representation and full
union promises where the checker admits the expression (including typed aliases),
but cannot make a rejected property expression reach lowering. No checker shim
or cohere source was copied or changed to bypass that limit.

The layout check enumerates generated ordinary literal layouts and public class
copy layouts. General dynamically changed/spread layouts are not established by
this unit. Structured or optional mismatches are refused as NotYet instead of
interpreting them with an unsafe primitive or reference representation. Extending
that checked-view machinery and admitting checker-rejected remainder syntax are
remaining work before calling the entire ruling complete.

## Verification commands and observations

Setup used `GOPROXY='https://proxy.golang.org|direct'`, `bash cloud/setup.sh`, and
`source /workspace/adamic-tools/env.sh`. Observed timing lines: Node 0.066s,
Go 0.071s, clang 0.478s, markdown dependencies 0.901s, submodules 160.775s,
Go build 388.882s, warm 388.984s, total 389.012s. `nproc`: 5, quota
400000/100000; Go 1.27.1, clang 20.1.8, Node 24.19.0.
Setup log: `/tmp/enum-setup.log`. Each test invocation wrote output to a log.

```sh
# In /tmp/enum-baseline: pinned main control, seconds, no native build:
GOFLAGS=-buildvcs=false go test -overlay=/tmp/enum-negative-frozen-overlay.json ./internal/lower -run '^TestNumericEnumMetadataIsNotDiscriminant$' -count=1 > /tmp/enum-negative-main-final.log 2>&1
# Checkpoint lowering:
go test ./internal/lower -count=1 > /tmp/enum-stepzero-lower.log 2>&1
# Implementation packages:
go test ./internal/lower ./internal/ir ./internal/javascript ./internal/native -count=1 -timeout 30m > /tmp/enum-final2-packages.log 2>&1
# Fresh lowering after the final recursive member-promise fix:
go test ./internal/lower -count=1 -timeout 30m > /tmp/enum-final3-lower.log 2>&1
# All oracles, uncached, final production sources:
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -count=1 -timeout 30m > /tmp/enum-final2-oracle-all.log 2>&1
# Regenerate and verify fixture counts (also verified by the full oracle):
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts > /tmp/enum-final2-counts3.log 2>&1
go vet ./... > /tmp/enum-final2-vet.log 2>&1
gofmt -l cmd internal > /tmp/enum-final2-format.log
git diff --check > /tmp/enum-final2-whitespace.log
```

Main control: FAIL with the intended old refusal (0.047s on the final frozen
overlay in `/tmp/enum-baseline`; initial run 0.044s). Checkpoint lower:
PASS 19.669s. Packages: lower 32.971s, IR 24.542s, native 148.171s, JavaScript
has no test files. Final fresh lower: PASS 23.291s. Full uncached oracle:
PASS 134.865s. Counts regeneration: PASS 43.099s. Vet, formatting and whitespace
checks pass. The full repository test gate was not run.

Current main advanced to `ce0750f28ef3943057f1f852b3ae5d93e6c5d644` and was merged
without conflicts in `4b3ad0ac790bd0b965c352abbd6b981ffc11010c`. Runtime signal
handling and oracle output tests changed upstream. The merged branch was
re-greened with these logged commands:

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/native ./internal/oracle -count=1 -timeout 30m > /tmp/enum-integrated-native-oracle.log 2>&1
go test ./internal/lower -count=1 -timeout 30m > /tmp/enum-computed-final-lower.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestEnumTag|TestNumericEnumNever|TestNativeAgreesWithNode/internal/oracle/testdata/enums_tag' -count=1 -timeout 30m > /tmp/enum-integrated-final-enum-oracle.log 2>&1
go vet ./... > /tmp/enum-integrated-vet.log 2>&1
```

Observed PASS: native 246.909s, full uncached oracle 267.499s, fresh lowering
60.628s, and final uncached enum oracle 5.631s after the computed-name fix.
Integrated vet passes with an empty log. The final census exposed computed-name
panics introduced by annotation recovery; that finding prompted the additional
repair and regression. The census was then rebuilt and rerun on final sources. It reports two
remaining computed-property Node.Text panics in declarations/diagnostics.ts on
each input (baseline had three panic sites). These are loud latent-census limits,
not successful lowering or generated program behavior.


## Mutants

Every listed compiler overlay was rebuilt and tested after merging current main
and applying the computed-property repair. Each exits 1 because its regression fails; none is a build failure.
Ten fail by assertions about refusals or runtime outcomes; the computed-name
mutant fails by the regression's compiler panic. No build failure is counted.
Each temporary overlay leaves production files unchanged.

| Mutation | Named catch and observation |
| --- | --- |
| Collapse shape identity back to reference bits | `TestEnumTagViewsPinned/boolean_checked`: native exits 0 instead of 70. |
| Skip explicit singleton member slot proof | `TestNumericEnumLiteralPromises/singleton_member_promise`: refusal disappears. |
| Skip recursive member-field promises | `TestNumericEnumLiteralPromises/singleton_member_view` and `singleton_member_array`: refusals disappear. |
| Skip singleton member cast check | `TestEnumTagViewsPinned/singleton_cast`: both backends exit 0 and print 42. |
| Skip nominal class check | `TestEnumTagViewsPinned/singleton_class`: both backends exit 0 instead of 70. |
| Skip payload literal-value checks | `TestEnumTagViewsPinned/literal_checked`: both backends print bad and exit 0. |
| Trust the checker's excluded-value object remainder | `TestNumericEnumObjectRemainderWideKeepsUnion`: an invalid narrowed view is accepted. |
| Skip the terminal missing-result check | `TestNumericEnumNeverPathsPinned/implicit`: JavaScript exits 0 with undefined; native falls into its generic compiler-bug panic, violating the pinned message. |
| Identify property subjects only by property symbol | `TestEnumTagViewsPinned/property_checked`: both backends exit 0 after a wrong payload read. |
| Omit public class copy layouts | `TestNativeAgreesWithNode/internal/oracle/testdata/enums_tag_remainder.a`: a valid copy wrongly stops in native while Node finishes. |
| Restore unchecked Node.Text for computed member-slot names | `TestNumericEnumComputedMemberSlotStaysLoud`: compiler panics instead of returning its explicit boundary. |
| Prove the open remainder never and erase default | Permanent `TestEnumTagOpenRemainderMutant`: both backends exit 0 before/after instead of the required default and exit 70. |
| Remove checked payload evaluations | Permanent `TestEnumTagPayloadMutant`: both backends exit 0 instead of 70; native prints 1 where Node prints text1. |

Temporary run results: `/tmp/enum-integrated-mutants-results.json`. Individual logs
are `/tmp/enum-integrated-{shape,member-slot,member-promise,member-cast,class,literal,remainder,missing-result,subject,public-shape,computed-name}-mutant.log`.
Permanent mutations pass in the final full uncached oracle. The historical
checkpoint guard mutant is `/tmp/enum-stepzero-mutant.log`; initial experiments
and mutants are retained in the initial report, not counted twice here.

## Meter reproduction

The requested meter README at `86255713:stage3/meter/README.md` documents
`bash stage3/meter/twice-daily.sh`; it does not contain a meter `go run` command.
The baseline paired run and exact input hashes are in the initial report. This
follow-up uses its latent Go tool on the identical pinned adapted input bytes:
main 48c05d09 and area b2c4549f, each covering all 79 source files.

```sh
python3 stage3/census/latent/make_overlay.py "$PWD" /tmp/enum-complete-overlay > /tmp/enum-complete-overlay.log 2>&1
go build -buildvcs=false -overlay=/tmp/enum-complete-overlay/overlay.json -o /tmp/enum-final-meter-corrected ./stage3/census/latent/tool > /tmp/enum-final-meter-corrected-build.log 2>&1
LATENT_ASSERT_NO_OUTPUT=1 /tmp/enum-final-meter-corrected /tmp/adamic-gate/stage3-meter.RLIiIU/main-adapted/src/compiler /tmp/enum-final-corrected-meter-run/main/latent.jsonl > /tmp/enum-final-corrected-meter-run/main/latent.log 2>&1
LATENT_ASSERT_NO_OUTPUT=1 /tmp/enum-final-meter-corrected /tmp/adamic-gate/stage3-meter.RLIiIU/area-adapted/src/compiler /tmp/enum-final-corrected-meter-run/area/latent.jsonl > /tmp/enum-final-corrected-meter-run/area/latent.log 2>&1
```

The meter's `latent_summary` validates full coverage and unique
`(kind, where, reason, text)` sites across all attempts. Final checked evidence:
[enum-tag-final-meter.json](enum-tag-final-meter.json). The checkpoint evidence
is [enum-tag-step-zero-meter.json](enum-tag-step-zero-meter.json). Final new
NotYet sites are matched by exact location to the checkpoint's 101 candidates;
no changed diagnostic reason is counted as successful lowering.

## Changed files and territory

The user authorized `enum_never.go` and the lowering reads needed for the
follow-up. The implementation additionally needs the existing IR property and
both backend read emitters to validate native untagged storage safely.
Every production file changed in this follow-up is named here:

- `internal/lower/enums.go`: discriminant boundary, written member promises, remainder view validation.
- `internal/lower/enum_tag_views.go`: checked primitive payloads, exclusions by value, annotation recovery and subject identity.
- `internal/lower/enum_never.go`: stored object remainder and checked never assertions.
- `internal/lower/expression.go`: full remainder representation and checked object observations.
- `internal/lower/object.go`: ordinary implicit switch fallthrough and checked missing-result edge.
- `internal/lower/assignments.go`: ordinary numeric remainder updates.
- `internal/lower/cast.go`: singleton member casts and checked payload views.
- `internal/lower/cast_proof.go`: qualified member annotation guard.
- `internal/lower/invariance.go`: mutable member-field promises and qualified annotation guard.
- `internal/ir/ir.go`: checked property message.
- `internal/native/emit_expressions.go`: validate layout before a checked field read, take or lend.
- `internal/native/emit_objects.go`: field types in shape identity and checked layout enumeration.
- `internal/javascript/javascript.go`: independent runtime field type check.

Tests and records: `internal/lower/enums_open_test.go`,
`internal/oracle/enums_open_test.go`, `internal/oracle/enums_tag_test.go`,
`internal/oracle/counts.md`, `internal/oracle/testdata/enums_open_never_update.a`;
new `.a` fixtures: `enums_tag_boolean_checked`, `enums_tag_literal_checked`,
`enums_tag_object_never`, `enums_tag_property_checked`, `enums_tag_remainder`,
`enums_tag_singleton_cast`, `enums_tag_singleton_checked`, `enums_tag_singleton_class`.
Documentation: `docs/enums.md`, `docs/verification/enum-tag-narrowing.md`,
`docs/verification/enum-tag-final.md`, `docs/verification/enum-tag-final-meter.json`.
The checkpoint also changed `docs/verification/enum-tag-step-zero-meter.json`.
Additional initial branch files: `internal/oracle/stage3_front_test.go`,
`internal/oracle/testdata/enums_tag_narrowing.a`,
`internal/oracle/testdata/enums_tag_never.a`, and
`docs/verification/enum-tag-narrowing-meter.json`.
Inherited main changes, including runtime signal behavior and meter artifacts,
are integration changes rather than edits by this unit.

The four prohibited files are untouched: `internal/native/emit.go`,
`internal/lower/lower.go`, `internal/native/native.go`, `internal/oracle/oracle_test.go`.
No cohere code was copied. Only `codex/enum-tag-narrowing` is pushed; no PR opened.
