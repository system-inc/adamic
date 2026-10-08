Built: four pinned parse variants, held-out Clang PGO and machine-function splitting; no production PGO policy change.
Commits: LTO implementation b2c8836c9934e343f3300fb409524ab213828941; LTO branch 586aaf96018e7aec0ae7c1cb434711e05f5104ed includes runtime audit; PGO helpers d3f83896088c5c3daff6a544b9834ee032f12ed1.
Commands and outputs: 38 training files, 39 held out; PGO saves 22.33% best user time over ThinLTO, splitting saves 28.19%; Go byte checks pass.
Mutants: actual compiled count and AST changes caught; duplicate corpus path and instruction/I1 accounting mutations caught; runtime dtoa.c flag omission caught on the inherited LTO branch.
Not covered: hardware cycles/cache misses, complete host isolation, production entrypoint profiles, service PGO, Go PGO, other targets, or activated SCC stack-check elimination.

# Held-out parse PGO on ThinLTO

PGO clears the requested 10% threshold on this workload. Propose a committed,
versioned profile for each stage 1 executable, with the regeneration and cache
contract below. This experiment does not make PGO the default for arbitrary
Adamic programs. ThinLTO remains the implemented shipping default. Ordinary
oracle, test, sanitized and counted policies retain their old flags.

The branch started at LTO report commit
370349804e893f33d443fe0f72faa3314bd71174. The later runtime compile audit was
merged as bd32218788412728220cbff770595f8d8bfe7c69 from codex/release-lto.
Those later changes are tests and documentation; measured production source
and flags are unchanged. Runtime base is
36669add9db2bbc06d486a58c37024f880e94a75, merged first on the LTO branch.
The full LTO implementation, release oracle and instrument check are in
[the LTO report](../release-lto/report.md) and [original measurements](../release-lto/measurement.md).

## Measurement and decision

All rows parse the same 39 held-out compiler files. Times are independently
selected minima of ten alternating executions, with explicit warmups. They
are not necessarily from the same repetition. Instruction counts and misses
are whole-process Callgrind observations, including startup and allocation.
`.text` is the ELF section size, including cold code, not GNU size's combined
code/rodata column. Binary size includes all sections.

| Variant | Instructions | Best user s | Best wall s | Simulated I1 misses | .text bytes | Binary bytes |
|---|---:|---:|---:|---:|---:|---:|
| -O2 baseline | 4,141,380,821 | 0.537220 | 0.570123 | 54,031,521 | 513,134 | 972,032 |
| ThinLTO | 3,749,459,423 | 0.469143 | 0.504348 | 83,848,900 | 1,010,128 | 1,339,864 |
| ThinLTO + PGO | 2,916,323,321 | 0.364384 | 0.392417 | 31,683,045 | 760,737 | 1,048,200 |
| ThinLTO + PGO + machine splitting | 2,916,321,827 | 0.336904 | 0.385325 | 30,695,160 | 763,953 | 1,065,784 |

Against plain ThinLTO, PGO saves 22.33% best user time, 22.19% best wall time,
22.22% instructions and 62.21% simulated I1 misses. Its text is 24.69% smaller.
Splitting saves 28.19% best user time, 23.60% best wall time and 63.39% I1 misses
against ThinLTO, with 24.37% smaller text.

Splitting's incremental saving over PGO is 7.54% best user time and 1.81% best
wall time. Median user times are 0.556360, 0.503045, 0.385750 and 0.377833 s,
respectively: splitting saves only 2.05% at the median over PGO. Median wall
saving is 0.56%. Do not treat the minimum's larger difference as a stable
7.54% benefit. It removes only 1,494 instructions, reduces I1 misses another
3.12%, and adds 0.42% text. llvm-nm finds 150 cold fragments totaling 291,844
symbol bytes, versus zero in PGO without splitting. Splitting happened, but
its incremental benefit merits another pinned repetition before a universal
policy. It can be an optional profile-backed x86 configuration.

| Variant | Simulated D1 read+write misses | Simulated conditional+indirect mispredictions |
|---|---:|---:|
| -O2 | 14,173,053 | 30,583,933 |
| ThinLTO | 14,219,322 | 24,779,327 |
| PGO | 14,051,022 | 34,881,217 |
| PGO + splitting | 13,653,494 | 34,706,780 |

Observation: plain ThinLTO nearly doubles code size and increases simulated
I1 misses over -O2, while improving time. PGO reduces both instructions and
code footprint, yet its simulated branch mispredictions increase 40.76% over
ThinLTO. Inference: code footprint and instruction reduction accompany this
win; the simulator does not establish hardware causality. The compiler's
48.8M instruction-cache/6.97M data-cache figures refer to an earlier full
parse/bucket denominator. These whole-process 39-file figures cannot be
compared directly to them or to the earlier 77-file instruction totals.

## Box, instrument and flags

Linux x86-64 KVM guest, AMD EPYC 9V74, nproc 5, cgroup CPU quota 4 cores,
17.6 GB memory. Every timed and simulated native execution is pinned with
taskset -c 3. No build, test or profiling work ran concurrently with timings.
One-minute load ranged from 0.897949 to 0.927734. Guest-wide process checks
cannot certify that the host had nothing else running or fixed frequency.
The previous perf stat -r 10 check could not access the required hardware
counters in this container (unsupported PMU events, paranoid=2). This report
uses hyperfine user time, not cycles. Simulated misses are not hardware misses.

Clang, lld, llvm-ar, llvm-size, llvm-nm and llvm-profdata are LLVM 20.1.8,
revision 87f0227cb60147a26a1eeb4fb06e3b505e9c7261. Hyperfine 1.19.0,
Callgrind 3.24.0, Go 1.27.1, Node 24.19.0. Setup was run again:
Go 0s, clang 0s, Node 0s, submodules 0s, warm build cache 35s, total 35s.
Environment: source /workspace/adamic-tools/env.sh.

Exact common compilation and linking flags, in order:

```
-std=c11 -Wall -Wextra -Werror -pedantic
-Wno-unused-variable -Wno-unused-but-set-variable -Wno-unused-function
-Wno-unused-parameter -Wno-self-assign
-ffp-contract=off -fno-optimize-sibling-calls -O2
```

ThinLTO adds -flto=thin everywhere and -fuse-ld=lld at link.
The training binary adds -fprofile-instr-generate to every runtime compilation
and the generated program's compilation/link. PGO replaces generation with
-fprofile-use=/workspace/scratch/release-pgo/training.profdata and
-Wno-profile-instr-out-of-date everywhere, including link. The fourth variant
adds -fsplit-machine-functions everywhere. No fast-math, march or FMA option
is used in these measurements. All semantic flags reach every compile and
link; -ffp-contract=off and -fno-optimize-sibling-calls remain effective.

The concrete argv for every runtime unit, archive and generated-program link
are in [commands.jsonl](evidence/commands.jsonl); ordered flag lists and sizes
are in [builds.json](evidence/builds.json). Runtime compilation follows:

```
clang <variant compile flags> -c /workspace/scratch/release-pgo/runtime/<unit>.c -o <variant runtime>/<unit>.o
llvm-ar rcs <variant runtime>/runtime.a <all 48 runtime objects>
clang <variant link flags> -I /workspace/scratch/release-pgo/runtime \
  -o <variant directory>/parse /workspace/scratch/release-pgo/parse.c \
  -Xlinker --whole-archive <variant directory>/runtime.a \
  -Xlinker --no-whole-archive -lm
```

The JSON argv is authoritative for paths/order. All variants use the same
pre-emitted batch 8 C and runtime snapshot. parse.c SHA256 is
fad9adf76a59cd7fe56e1a5ee8e25a9af7338c046b0c24395fb832c70397885f.
No emitter or runtime code was hand-edited for optimization.

## Held-out training and reproduction

The 77 corpus paths were sorted before measurement. Odd zero-based indices
form the 38-file training set; even indices form the 39-file held-out set.
The partition asserts 77 distinct paths, disjointness and complete coverage.
It was fixed before observing results. Training input is 3,180,280 bytes;
held-out input is 6,219,795 bytes. The large checker file is held out.
[split-plan.json](evidence/split-plan.json) records every path, length and hash;
[train-files.json](evidence/train-files.json) and
[held-out-files.json](evidence/held-out-files.json) name the input paths. Actual manifest paths used by argv are recorded in
[train.txt](evidence/train.txt) and [held-out.txt](evidence/held-out.txt).

Training files (full paths):

- src/compiler/_namespaces/ts.performance.ts
- src/compiler/binder.ts
- src/compiler/builderPublic.ts
- src/compiler/builderStatePublic.ts
- src/compiler/commandLineParser.ts
- src/compiler/corePublic.ts
- src/compiler/emitter.ts
- src/compiler/expressionToTypeNode.ts
- src/compiler/factory/emitHelpers.ts
- src/compiler/factory/nodeChildren.ts
- src/compiler/factory/nodeFactory.ts
- src/compiler/factory/parenthesizerRules.ts
- src/compiler/factory/utilitiesPublic.ts
- src/compiler/moduleSpecifiers.ts
- src/compiler/path.ts
- src/compiler/performanceCore.ts
- src/compiler/programDiagnostics.ts
- src/compiler/scanner.ts
- src/compiler/sourcemap.ts
- src/compiler/sys.ts
- src/compiler/transformer.ts
- src/compiler/transformers/classThis.ts
- src/compiler/transformers/declarations/diagnostics.ts
- src/compiler/transformers/es2015.ts
- src/compiler/transformers/es2017.ts
- src/compiler/transformers/es2019.ts
- src/compiler/transformers/es2021.ts
- src/compiler/transformers/esnext.ts
- src/compiler/transformers/jsx.ts
- src/compiler/transformers/module/esnextAnd2015.ts
- src/compiler/transformers/module/module.ts
- src/compiler/transformers/namedEvaluation.ts
- src/compiler/transformers/ts.ts
- src/compiler/transformers/utilities.ts
- src/compiler/tsbuildPublic.ts
- src/compiler/utilities.ts
- src/compiler/visitorPublic.ts
- src/compiler/watchPublic.ts

Only these 38 paths are passed to the instrumented binary. llvm-profdata
merges the explicit training.profraw file, without globbing unrelated runs:

```
LLVM_PROFILE_FILE=/workspace/scratch/release-pgo/training.profraw \
  taskset -c 3 /workspace/scratch/release-pgo/generate/parse \
  --manifest /workspace/scratch/release-pgo/train.txt --count
/workspace/adamic-tools/llvm/bin/llvm-profdata merge -o \
  /workspace/scratch/release-pgo/training.profdata \
  /workspace/scratch/release-pgo/training.profraw
```

Training profile SHA256:
aee7e80711ea9ea94fe899bcef387178a58b5853b3284db398897d7e76811aef.
Repeating training and merge produced a byte-identical profile, recorded in
[reproduce-profile.log](evidence/reproduce-profile.log).

The measured command is taskset -c 3 <variant directory>/parse
--manifest /workspace/scratch/release-pgo/held-out.txt --count.
measure.py rotates the four variants each round and reverses ordering after
every four rounds. Hyperfine runs one execution at a time with --runs 1,
--shell none, --show-output and JSON export; ten rounds supply each variant's
samples. Output is checked on every execution, including warmups. Raw forty
hyperfine logs/JSON and all round orders/load observations are preserved.
User times come from hyperfine's user field for each one-run invocation.

Simulation runs serially after timing:

```
taskset -c 3 valgrind --tool=callgrind --cache-sim=yes --branch-sim=yes \
  --I1=32768,8,64 --D1=32768,8,64 --LL=268435456,1,64 \
  --callgrind-out-file=<variant>.callgrind <variant directory>/parse \
  --manifest /workspace/scratch/release-pgo/held-out.txt --count
```

This is an explicit comparison cache model, not a claim about EPYC's real
hierarchy. All four profiles record all 13 requested events. Summing self
cost lines equals the totals footer for every event. The summary has a
consistent two extra startup instructions; all other summary deltas are zero.
Actual +1 instruction-summary and +1 I1-total mutations are rejected for each
variant. Compressed raw profiles and [profiles.json](evidence/profiles.json)
preserve the counts and accounting.

## Output floor, mutants and stale profile

All four timed binaries match Go's count stdout 0 followed by newline,
empty stderr and exit zero on the 39 files. Because that is a small count
witness, an untimed full-AST companion is also rebuilt for every variant and
compared byte for byte to Go: 29,704,698 bytes, SHA256
a5fef2d09ed576e425da6c047beb3169573de15af6ea07914d88c6c194b1a150.
The companion emits the full AST rather than timing it; its generated C is a
different entrypoint. It therefore has its own profile, trained on the same
38 paths. Every companion runtime unit uses that same companion profile.
These are separate witnesses, not a claim that the timed binary dumps ASTs.

A real compiled SourceFile-to-XourceFile output mutant has the same output
length but fails at byte 14. A real compiled timed-parser count-initializer
mutant prints 1 rather than 0. Both the original execution guard and the
original hyperfine output guard reject it. A duplicate actual corpus-path
mutant is rejected by the original partition guard. See prove.log,
validation.json and the retained mutant build commands/output. No good output
difference was found; the experiment did not continue through a miscompile.

The stale experiment trains on a different generated program: the full-AST
entrypoint, placed under the parse.c filename key, with different function
IDs, main and CFG. The foreign source SHA256 is
8f4e36c96fa2ac2964087df18eab57e22a49f88bf01b055f7fc32b4030dd36bf;
foreign profile SHA256 is
5a5f5be3a09a15eed3ec70ef277906c8f55ef636ee93a0e2cf1dfa7d96786a9a.
Only the same 38 training inputs are used. Every runtime file and the original
timed parse program are rebuilt with that foreign profile plus splitting.
The held-out count output still matches Go. The full-AST witness also matches
the same 29,704,698 Go bytes exactly. stale.json records hashes and flags.

A diagnostic compilation without the out-of-date suppression reports:
“of 313 functions, 2 have mismatched data that will be ignored”. It succeeds;
the warning is visible in stale-diagnostic.stderr. This observed mismatch
changes optimization input without changing output. Profile weights are
optimization hints, not permission to change arithmetic or other language
semantics. This is a concrete stale-profile check, not a universal proof of
LLVM correctness for every profile. Keep the external byte gate and semantic
mutants on the final binaries.

Two initial witness builds exposed integration requirements rather than
wrong outputs: the parse profile cannot cover the distinct AST main
(-Wprofile-instr-unprofiled under -Werror), and mixing two different profiles
in one ThinLTO link fails with conflicting ProfileSummary module flags.
They are retained as validation-initial.log, validation-mixed-profile.log
and mixed-profile-link.stderr. Rebuilding every unit with a single matching
profile passed. The stale AST witness uses a diagnostic-only suppression of
unprofiled-function warnings because of its filename key; that extra flag
is absent from every timed candidate and the stale timed parse rebuild.

## Proposed stage 1 profile contract

For the real parse executable, commit the profile under a per-entrypoint and
per-target directory such as stage1/profiles/cohere-parse/linux-amd64/, with
profile.profdata and manifest.json. The benchmark profile is committed here
as evidence; it is not a production input for other entrypoints. In particular,
this harness uses parse.c while native.Build emits main.c. Regenerate using
the real production filenames and entrypoint before shipping a profile.

The manifest must pin profile content hash, LLVM/compiler/linker versions,
target, every semantic and optimization flag, compiler/runtime commits,
runtime source/header hashes, generated C hash and stable C filenames,
training harness and corpus hashes, partition, and held-out byte-gate result.
Profiles must be generated with the shipped semantic policy and ThinLTO on
all generated and runtime units. Instrument the production entrypoint,
train only the fixed training set, merge explicitly named raw files with the
pinned llvm-profdata, rebuild all units with that one immutable profile,
and rerun the held-out external oracle and real mutants before committing.
The repeated merge here demonstrates reproducibility for this exact input.

On source, toolchain or target drift, validate the manifest, warn explicitly,
and fall back to plain ThinLTO until regeneration. A profile-required release
job may instead fail. Suppressing -Wprofile-instr-out-of-date is not a freshness
check: renamed IDs or filenames can silently lose profile coverage. Keep
-ffp-contract=off and -fno-optimize-sibling-calls on every compilation and link.

Include profile content identity in runtime.a and translation-unit cache
keys. Current flag hashing includes the profile argument's path, not the
file's bytes. A mutable profile at a fixed path would incorrectly reuse
objects. Snapshot it to an immutable hash-addressed path, then that path can
participate in existing flag keys, or explicitly hash its contents. All units
and the link must use the same blob; the mixed-ProfileSummary failure proves
why. Developer tools must supply stable generated translation-unit filenames,
propagate the full compile/link semantic policy, share the exact profile blob,
key objects by profile/toolchain/source identity, and provide their actual
split entrypoint for regeneration and byte validation.

Backend build observations, with every runtime unit rebuilt and pre-emitted
C, warm filesystem, one observation each: -O2 5.038s; ThinLTO 18.769s;
instrumentation 5.294s; PGO 9.153s; PGO plus splitting 9.339s. These exclude
checking, emission and profile training/merge, and are not a cold-versus-cached
build series. Training is a regeneration cost; ordinary shipping consumes the
committed profile. LTO's existing cold/cached archive cost is in its report.
Do not extrapolate these one-off backend measurements into a release SLA.

## Go fairness, tests and limits

Go cohere revision 715ba94f3608a6500086b1076ce5cb7e51b836db has no default.pgo
anywhere in its checkout. The measured Go build has no -pgo setting, and
GOFLAGS was empty. Its exact build command:

```
cd cohere && go build -overlay=/workspace/scratch/release-lto/parse-overlay.json \
  -o /workspace/scratch/release-lto/go-parse /workspace/adamic/cohere/adamic_parse.go
```

Go's default automatic PGO lookup is enabled, but there is no profile to load.
The measured Go binary uses no PGO. This report's optimization decision is
native PGO versus native plain ThinLTO, not a claim against PGO-trained Go.
See go-pgo-check.json and go-build-settings.txt.

The report-only helpers were syntax checked, all byte gates and stale-profile
checks ran, and real mutants exercised partition, timing, AST and profile
accounting guards. Logs are files, never piped. No new compiler production
change needs another full gate here. The inherited LTO implementation ran the
420-fixture opt-in exact release oracle and complete touched native, oracle
and CLI packages; its interrupted root gate and cohere fixture coverage limits
are documented there, not claimed passed.

The latest additional runtime test performed real cold archive builds for
release, ordinary sanitized oracle and counted oracle, checking all 48 clang
commands per lane including dtoa.c and ieee754.c. Removing contraction
protection only from dtoa.c still built an archive and was caught by the
command audit. The focused flags test passed in 15.865s; evidence is in the
LTO report. This inherited test leaves measured production flags unchanged.

No hardware cycles, IPC, real branch misses or real I-cache events are
available here. No service PGO, Go profile training, other architecture,
nightly integration, or production stage 1 profile was measured. The cited
SCC branch still reports an inactive classifier and unchanged stack-check
placement; an active stack-check removal was not present. Regenerate profiles
and repeat the pinned held-out experiment after that optimization or later
field-store/source changes land. Results are evidence for this exact source,
box, cache model and partition, not a whole-host isolation guarantee.

Helpers: prepare.py builds/trains the partitioned variants; validate.py builds
full-AST witnesses and its mutant; measure.py times/profiles with live output
checks; stale.py builds the foreign-profile witnesses; prove.py exercises the
original guards with compiled mutants. Their absolute workspace paths are
recorded intentionally. [evidence.sha256](evidence/evidence.sha256) covers
all retained logs, command arrays, profiles, C snapshots and manifests.

Empty build logs and large AST byte streams are represented by lengths and
SHA256 in build-output-manifest.json rather than duplicated as artifacts.
Nonempty diagnostic logs, all timing logs, profiles and output-mutant proofs
are retained.
