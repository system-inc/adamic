# Wave 10

Built three native `.a` rules and the raw `member-parameters` checker question.
Claim commit: `e8a4e89694021f49957bfa5133ca40d000b5670e`; implementation commit is the commit containing this report.
Final wave test passed in 72.534s: 28 controls, 77 compiler files and 287 repository files; full findings/fixes/suggestions bytes agree.
Three rule mutants were caught only by byte comparison; released-registry, source-gate and existing Node one-byte mutants were also caught.
No full repository gate, nondefault option matrix or exhaustive upstream fixtures; an unrelated optional-initialization compiler gap is recorded below.

## Selection and scope

Base: `0d540f413625f016f20fea39761c7b184f335de6`, the requested
`origin/codex/tsgo-c-library`, rather than generic main. All origin heads were
fetched and their distinct stage1 trees searched before the claim was pushed.
No selected rule was already ported or claimed; no skips.

| Remaining position after the base 26 | Rule | Compiler | Repository |
| --- | --- | ---: | ---: |
| 28 | no-unmodified-loop-condition | 4 | 0 |
| 29 | @typescript-eslint/no-redundant-type-constituents | 3 | 0 |
| 30 | @typescript-eslint/prefer-includes | 2 | 1 |

The new files follow the existing Rules/parser/facts pattern. The loop rule uses
binding and reassignment facts, conditional groups, function references and
suspension checks. Redundant constituents use raw checker types and native
absorption decisions. Includes handles the production sentinel comparisons,
textual overload parameter matching and production static regex subset, with
ordered fixes and optional-chain refusal to fix.

`member-parameters` returns declaration function-like flags and parameter source
text, plus the same facts for each declaration parent's `includes` property.
It returns no lint verdict or edit. Go owns its own question file and test;
Adamic owns its decoder file. The only shared source edit is one switch arm in
`bridge/tsgo/checker/facts.go`; no Adamic shared registration is needed because
the separate decoder calls the existing string-dispatched `ask` API. No protected
compiler files or upstream cohere files were edited. All new Adamic files are `.a`.

## Observations

Setup succeeded: Go 1.27.1 ready 0s; clang 20.1.8 ready 0s; Node 24.19.0 ready 0s;
submodules ready 0s; build cache warm 88s; total 88s. `nproc` is 5, with a
4-CPU cgroup quota and 17.6 GB memory. Source `/workspace/adamic-tools/env.sh`.

The independent oracle builds through a Go overlay inside pinned cohere
`715ba94f3608a6500086b1076ce5cb7e51b836db`, imports no bridge code, and calls its
unchanged production rules with default options. TypeScript is v6.0.3,
`050880ce59e30b356b686bd3144efe24f875ebc8`. Both corpora use the existing
`validation-coverage` manifests; repository population remains frozen before
these ports. `validation-wave-10/source-hashes.json` records every source hash.

| Population | Findings | Complete stream bytes | Native normal and ASan/UBSan/LSan |
| --- | ---: | ---: | --- |
| 28 controls | 61 | 18383 | identical to Go |
| 77 compiler files | 9 | 8346 | identical to Go |
| 287 repository files | 1 | 18863 | identical to Go |

Canonical streams preserve rule/message IDs, descriptions, byte spans, ordered
fixes and suggestions, and file headings. This run has no nonempty suggestions;
the protocol compares their counts and edits too. Sanitized bridge and native
binaries used address/undefined sanitizers with no recovery, plus leak detection;
all comparison stderr was empty. Raw stdout streams are gzip-compressed without changing their bytes; hashes
name the original uncompressed streams. Stderr and hashes are committed.

| Mutant | Normal execution | What catches it |
| --- | --- | --- |
| Loop group suppresses when a member is unmodified instead of modified | exit 0, empty stderr | Go byte oracle at byte 542 |
| Never exemption applies outside rather than inside return types | exit 0, empty stderr | Go byte oracle at byte 4536 |
| Includes fix negates positive rather than negative comparison | exit 0, empty stderr | Go byte oracle at byte 8396 |
| Registry retains released program handle | exit 0 | required panic 70, exact invalid/released-handle stderr |
| Source gate receives an unused variable | exit 1 | configured production no-unused-vars diagnostic |
| Existing Node one-byte oracle mutant | test passed | Node/native byte comparison |

Checker package: all six tests passed, 0.128s, including Unicode parameter
source text and question suffix refusal. `go vet` of the checker and typeaware
packages passed with empty output. Filtered Node oracle passed in 37.954s, covering
functions, closures, method closures, generic functions, maps/text, sorting,
string indexing and lone surrogates, with sanitizer/leak checks.

The pinned cohere CLI skips explicit `.a` files. `testdata/source_gate_wave_10.go`
therefore uses its formatter and configured production rule APIs directly, with
a read-only `.a.ts` virtual view matching stage 0. No `.ts` source is written.
All five new production `.a` files formatted and linted with zero findings; the
unused-variable mutant proves this gate fails. Gap files are intentionally excluded.

## Timings

Three alternating native/Go rounds, same configs/manifests, count-only output,
no builds or other tests running. These are complete-process medians, not just
rule work. Raw rounds and phase counters are in `validation-wave-10/bench`.

| Corpus | Native | Go | Native / Go |
| --- | ---: | ---: | ---: |
| compiler | 4.204408s | 0.547324s | 7.682 |
| repository | 0.447726s | 0.136648s | 3.276 |

Native is slower. Compiler native median load was 0.241086s, checker queries
1.153181s across 310381 calls, and post-load run 3.935009s. Go load was 0.250219s,
rule callbacks 0.240100s and post-load run 0.280831s. Repository native used
24911 queries. Query time is contained in run time; do not add them together.
The measured call volume supports a future candidate-indexing optimization;
this unit does not claim one.

## Failures corrected and limits

An initial compiler comparison found two spurious redundant object findings:
the pinned checker uses Never=262144, Object=131072. The port now uses the pinned
flags, with alias-never and object negative controls. A naming cleanup accidentally
changed the ConditionalExpression tag; the next byte comparison caught it, and
it was corrected. Initial failure logs are retained. Runner formatting exposed
unsupported function expressions; final code uses the existing parent-buffer
while-loop pattern, without compiler changes.

Removing an explicit undefined initializer made the native regex search treat an
absent value as an empty string, adding compiler findings. A minimized external
Node comparison confirms a stage-0 gap: `gaps/wave_10_optional_initialization.a`
compiles successfully and native prints `false`, while Node prints `true`.
The port avoids this path using an initialized string plus a presence flag and
adds an unknown-regex-identifier negative control. Fixing the compiler itself is
outside this unit's authorized files. This is an observed compiler defect, not
proof about every optional-value path. The gap remains open and is not counted
as a passing oracle. Its probe/build outputs are retained.

Validation covers the specified frozen corpora and the 28 controls, with production
default options. It does not establish exhaustive equivalence for all TypeScript
programs, custom lint options, project configs, upstream fixture suites or all
stage-0 features. No `go test ./...` or entire-repository lint run is claimed.

## Reproduce

With setup environment sourced, materialize the two absolute manifests using the
script in `validation-coverage/README.md`, then run:

```sh
ADAMIC_TYPESCRIPT_SOURCE=/path/to/TypeScript-v6.0.3 \
ADAMIC_WAVE10_ARTIFACTS=/tmp/wave10/artifacts \
ADAMIC_WAVE10_REPOSITORY_MANIFEST=/tmp/wave10/repository.manifest \
ADAMIC_WAVE10_COMPILER_MANIFEST=/tmp/wave10/compiler.manifest \
go test ./stage1/cohere/typeaware -run '^TestWave10' -count=1 -timeout=15m -v \
  > /tmp/wave10/test.log 2>&1
go test ./bridge/tsgo/checker -count=1 -v > /tmp/wave10/checker.log 2>&1
go vet ./bridge/tsgo/checker ./stage1/cohere/typeaware > /tmp/wave10/vet.log 2>&1
go test ./internal/oracle \
  -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(maps_and_text|sorting|string_index|lone_surrogates|functions|closures)\.a$' \
  -count=1 -timeout=10m -v > /tmp/wave10/node.log 2>&1
python3 bridge/tsgo/profile/volume_bench.py \
  /tmp/wave10/artifacts/wave10 /tmp/wave10/artifacts/wave10-oracle /tmp/wave10/bench \
  --corpus compiler "$ADAMIC_TYPESCRIPT_SOURCE/src/compiler/tsconfig.json" /tmp/wave10/compiler.manifest \
  --corpus repository "$PWD/tsconfig.json" /tmp/wave10/repository.manifest \
  > /tmp/wave10/benchmark.log 2>&1
```

For the source gate, overlay `cohere/adamic_wave10_source_gate.go` with the
absolute `testdata/source_gate_wave_10.go` path as in the existing Go oracle
overlay pattern. Build it inside cohere and pass repository tsconfig plus a
manifest of member_parameters.a, the three rule .a files and wave_10_suite.a.
`--format` runs the formatter; no flag runs configured lint and exits 1 on findings.
For the optional gap, build the committed `.a` file with stage 0, run the binary,
and compare with `node --input-type=module-typescript < file.a`, keeping both
outputs in logs. Expected current mismatch: native false, Node true.
