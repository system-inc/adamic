# Numeric enum literal tags, October 7, 2026

Built member-tag narrowing with unrelated open enum metadata, and reachable explicit numeric defaults; the full ruling remains incomplete.
Implementation commits: `039e5f741ebfcea8b7fc85bb0f9229da159438e9` and `c9cd9d6216b120c5cb7f70f912b8a2f2187f0d28`, based on `48c05d091f0a43c31cbe051b1d6578d99eeedf19`.
Validation: full lowering and uncached enum oracle pass; final commands and meter observations follow below.
Mutants: erased open remainder, premature exhaustiveness panic, and removed singleton refusal were independently caught.
Not covered: checked ambiguous singleton payload views, restoration of the checker's full object remainder, implicit-default contract changes, or the full repository gate.

The conservative assumption in the implementation commit is that an ambiguous
singleton payload remains refused until its actual storage can be validated.
This is an explicit departure from completing the entire ruling, not a claim
that the singleton case is impossible with a larger implementation.

## Observations

The old refusal examined every field of every union variant. A broad numeric
enum used for metadata therefore prevented an unrelated literal member tag from
narrowing the object. The new boundary concerns the observed singleton enum tag,
whose checker type is also its open whole enum. Multi-member literal tags retain
the existing proof and lowering. Enum constants and aliases already compare by
numeric value; no new nominal comparison was introduced.

An explicit numeric default no longer participates in the checker's declared
member exhaustiveness shortcut. It executes its own body. A `never` binding
there remains a checked read, including when the switch's input came from a
member constant. String-enum default elision is unchanged. Implicit numeric
switch defaults retain the earlier checked-stop contract.

`enums_tag_narrowing.a` runs source Node, generated JavaScript, native release,
ASan/UBSan and the finishing-program leak check. It covers `===`, `!==`,
`switch`, `FirstX`/`LastX` aliases, open `MetadataFlags` alongside a member tag,
object literals, an exact class, an existing checked downcast, and numeric 42
reaching an ordinary default after both declared values were excluded.

`enums_tag_never.a` independently runs source Node to observe that type erasure
prints `before`, `default`, `unreachable`, `after`, and exits 0. The added check
in both compiled backends instead prints `before`, `default`, exits 70, and
pins the entire owned panic message:

```text
adamic: panic: unreachable value 42 for numeric enum SyntaxKind
```

## Blocker and removed experiment

The old singleton adversarial probe remains refused. Its string-payload object
can store numeric 1 in an open enum tag, then enter the checker's number-payload
branch. Merely removing the refusal would silently read the wrong native slot.

An experiment attempted to validate narrowed payloads with `typeof` on an
`ir.Property` read as `ir.Union`. Native ordinary primitive fields are stored
in an untagged C union, not as boxed `ir.Union` values. The experiment failed
under ASan with a SEGV in `adamic_retain` on the otherwise valid narrowing
fixture. All experiment code was removed before the implementation commit.
This ASan failure is a rejected implementation, not a killed semantic mutant.

A complete implementation needs checked object-view/storage support beyond
`enums.go`. The unit asked to extend territory to the existing never/object
lowering files; no answer was received during work. No file outside the named
implementation territory, tests, fixture counts and documentation was changed.
The four explicitly prohibited files were untouched. No cohere code was copied.

This unit also does not replace the front end's object-union `never` with the
full declared union. Existing numeric enum reads that the checker calls never
remain runtime checks, including uses other than an explicit assertion. These
are outstanding parts of the ruling, rather than implied proofs of completion.

## Toolchain and commands

Every Go/test shell sourced `/workspace/adamic-tools/env.sh`, the setup's printed
path. `GOPROXY='https://proxy.golang.org|direct'` was set before setup.
`bash cloud/setup.sh` passed: Node ready 0.066s, Go ready 0.071s, clang ready
0.478s, markdown dependencies ready 0.901s, submodules ready 160.775s, Go build
ready 388.882s, cache warm 388.984s, total 389.012s. `nproc` is 5; CPU quota is
400000/100000. Versions: Go 1.27.1, clang 20.1.8, Node 24.19.0.
Log: `/tmp/enum-setup.log`.

```sh
go test ./internal/lower -count=1 -timeout 30m > /tmp/enum-final-lower.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestEnumTag|TestNumericEnumNever|TestNativeAgreesWithNode/internal/oracle/testdata/enums' -count=1 -v -timeout 30m > /tmp/enum-final-oracle.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts > /tmp/enum-final-counts.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m > /tmp/enum-counts-verify.log 2>&1
go vet ./... > /tmp/enum-vet.log 2>&1
gofmt -l cmd internal > /tmp/enum-format.log
git diff --check > /tmp/enum-whitespace.log
```

Observed passes: final lowering 20.190s (previous implementation 42.688s), filtered uncached oracle 5.386s, final counts
regeneration 34.473s, counts verification 26.990s. Vet, gofmt and whitespace
logs are empty. The filtered oracle has native hits 0/misses 55 and Node hits
0/misses 47. All three mutant invocations have separate logs.

The first complete oracle run exposed a superseded expectation in
`TestStage3EnumFallthrough/unmatched`: it required exit 70 before default,
while independent source Node and every backend now finish with `DB`.
The expectation was updated to the source oracle, and the leak check now covers
that finishing path too. This test-only change is in `c9cd9d62`, along with
preserving the original exemption for nullable views with only one object
variant. The full package was rerun uncached and passed in 205.765s:

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -count=1 -timeout 30m > /tmp/enum-final-oracle-all2.log 2>&1
```

The final vet, formatting and whitespace logs are empty:
`/tmp/enum-final-{vet,format,whitespace}.log`. The full repository gate was not
run; both complete touched packages passed.

Counts add two rows. No pre-existing numeric row changed:

| Fixture | Allocations | Frees | Retains | Releases | Peak | Regions |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| enums_tag_narrowing.a | 17 | 17 | 15 | 30 | 6 | 0 |
| enums_tag_never.a | 3 | 0 | 2 | 1 | 3 | 0 |

The stopped process intentionally retains its live allocations at panic.

## Mutants

| Mutant | Catch and observed result |
| --- | --- |
| Treat the open remainder as proven never and erase its default | Permanent `TestEnumTagOpenRemainderMutant`: sanitized native and JavaScript both finish with exit 0, empty stderr and `before\nafter\n`; the pinned stop differs. |
| Count declared members as exhaustive before an explicit default | An isolated Go overlay of `enumSwitchCovered`, tested by `TestEnumTagNeverPinned`: both backends exit 70 before printing `default`, and its stdout assertion fails. |
| Remove the ambiguous singleton refinement refusal | An isolated Go overlay, tested by `TestNumericEnumLiteralPromises/open_tag`: assertion fails with `want literal-promise refusal, got <nil>`. |

No Go or clang build failure is counted as a killed mutant. Production compiler
source was unchanged by the two overlays. Logs: `/tmp/enum-mutant.log`,
`/tmp/enum-default-mutant-final.log`, `/tmp/enum-singleton-mutant-final.log`; permanent mutant
also reran in `/tmp/enum-final-oracle.log`.

## Meter

The obsolete refusal row is **966 to 0 on main** and **1,182 to 0 on the
compiler-area tree**. The retained ambiguous singleton refusal has **0 sites
before and after on both inputs**. This is not a renaming of that measured row.
Every before/after latent census passes the meter's full coverage validation:
79 TypeScript/Adamic source files per tree. Counts remain measurements on a
checker-rejected program, not claims that the whole compiler builds.

| Input tree | Old refusal before | Old refusal after | Retained singleton refusal after |
| --- | ---: | ---: | ---: |
| origin/main at 48c05d09 | 966 | 0 | 0 |
| origin/area/stage3 at b2c4549f | 1,182 | 0 | 0 |

[The compact JSON evidence](enum-tag-narrowing-meter.json) records compiler
commits, complete input commits, input-byte hashes, census totals, coverage and
both diagnostic rows. The after binary is built from `c9cd9d62`; its SHA-256 is
`cb629bf109a61f20f197e8702b407ad888e5a18ae98c153b9cfd0493ca1ba18a`.

The README at `86255713:stage3/meter/README.md` documents the paired shell
runner, rather than a `go run` command. The before run used that runner from a
detached baseline checkout at `48c05d09`:

```sh
cd /tmp/enum-baseline
source /workspace/adamic-tools/env.sh
export GOPROXY='https://proxy.golang.org|direct'
GOFLAGS=-buildvcs=false GIT_CONFIG_COUNT=1 GIT_CONFIG_KEY_0=fetch.recurseSubmodules GIT_CONFIG_VALUE_0=false STAGE3_METER_RUNS=/tmp/enum-meter-before-final bash stage3/meter/twice-daily.sh > /tmp/enum-meter-before-final.log 2>&1
```

The baseline shares the initialized cohere checkout through a symlink. Git
submodule recursion and VCS stamping were disabled only in that scratch run:
otherwise Git refuses the symlink. No compiler or loader option was changed.
The complete paired baseline exits 0 and writes
`/tmp/enum-meter-before-final/20261007T212952Z.cInKwB/report.json`.

Initial meter attempts occurred before nested-submodule setup finished, then
encountered the scratch-symlink/VCS issue. They were retried as described above.
An initial after runner was superseded by the nullable-view correction and its
partial results were not used. The final after measurement uses the documented
latent census overlay on the exact same pinned baseline input bytes, avoiding a
fresh adaptation or source-version difference:

```sh
cd /workspace/adamic
source /workspace/adamic-tools/env.sh
python3 stage3/census/latent/make_overlay.py "$PWD" /tmp/enum-final-overlay > /tmp/enum-final-overlay.log 2>&1
go build -buildvcs=false -overlay=/tmp/enum-final-overlay/overlay.json -o /tmp/enum-final-latent ./stage3/census/latent/tool > /tmp/enum-final-latent-build.log 2>&1
LATENT_ASSERT_NO_OUTPUT=1 /tmp/enum-final-latent /tmp/adamic-gate/stage3-meter.RLIiIU/main-adapted/src/compiler /tmp/enum-meter-final/main/latent.jsonl > /tmp/enum-meter-final/main/latent.log 2>&1
LATENT_ASSERT_NO_OUTPUT=1 /tmp/enum-final-latent /tmp/adamic-gate/stage3-meter.RLIiIU/area-adapted/src/compiler /tmp/enum-meter-final/area/latent.jsonl > /tmp/enum-meter-final/area/latent.log 2>&1
```

The two independent final census executions ran concurrently; both exit 0.
The meter's `latent_summary` independently validated complete coverage and
calculated unique sites from all attempts before saving the compact evidence.
No backend is enabled in the measurement binary. Raw before JSONL is preserved
compressed by the runner; final raw JSONL and logs remain under
`/tmp/enum-meter-final/{main,area}/`.

Final `git fetch origin main` and `git merge --no-edit origin/main` report
`Already up to date.` Main remains `48c05d09`. Only the feature branch is pushed;
no main or area branch is modified, and no pull request is opened.
