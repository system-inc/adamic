# UTF-16 views

Built on `codex/utf16-views` from main `fe3b9f2`. The scanner branch,
`247500b`, was merged without a commit in `/workspace/utf16-measure/repo`
on the separate scratch branch `codex/utf16-measure`. That merge is not
part of this unit's history. Implementation commit: `bef8a8688a746c6bd3738c63d0671a2a9f3f1083`.
No compiler or lowering files changed.

`codePointAt` locates and decodes once. `charCodeAt` reads its cached length
without a second length call. Slices decode their already located boundary
halves instead of locating those units again. The forward cursor retains
its original cheap path; backward reads choose the cursor when it is no
farther away than the checkpoint and retreat over continuation bytes,
subtracting two UTF-16 units for a supplementary point and one otherwise.
The cursor always names the first unit of a code point.

An indexed string containing only BMP points gets a `uint16_t` view for
direct unit reads. Lone surrogates qualify. The existing index eligibility
rules still apply: non-ASCII, at least 64 bytes, owned or a marked literal,
and offsets fitting checkpoints. Short and stack strings allocate no view,
and ASCII remains a direct byte read. The BMP view adds two bytes per unit,
one pointer in the index, and a decoding pass at index construction. Its
lifetime is the index's lifetime, including invalidation before an in-place
append. No string layout or runtime header changed.

## Measurements

Observed Callgrind 3.24.0 `Ir`, clang 20.1.8 release flags with `-O2 -g`,
without sanitizers or RC instrumentation. TypeScript v6.0.3 compiler source
is commit `050880ce59e30b356b686bd3144efe24f875ebc8`: 77 files, 434,790 tokens.
The scanner driver still calculates values, flags and errors in count mode.
Both builds use the same scanner source and corpus. Only the two runtime
string files differ between the measured snapshots.

| Scanner instructions | Before | Final | Change |
| --- | ---: | ---: | ---: |
| Whole process | 2,429,579,448 | 2,123,196,245 | -12.61% |
| Inclusive `main` | 2,429,400,549 | 2,123,017,336 | -12.61% |
| Whole process per token | 5,587.94 | 4,883.27 | -12.61% |

A repeat measured exactly 2,123,017,336 inclusive `main` instructions again.
Whole-process Ir was 2,123,196,263, 18 higher; only two anonymous startup
routines' self costs changed. The deterministic driver workload is
therefore distinguished from this tiny process-startup variation.
The profile summarizer verifies that every self cost sums to the summary.
Inclusive costs overlap and must not be added. Inlining moves attribution
between helpers, so the total is the performance comparison.

Selected self instructions, not additive to inclusive rows:

| Function | Before | Final |
| --- | ---: | ---: |
| `adamic_string_locate` | 288,080,341 | 35,994,723 |
| `usable` | 153,766,828 | 110,761,836 |
| `adamic_string_units` | 104,515,436 | 85,787,221 |
| `adamic_string_char_code` | 157,435,393 | 176,168,501 |
| `adamic_string_bmp_view` | 0 | 37,455,410 |

The string fixtures, generated once as C and compiled against each runtime
snapshot with the same release flags, print identical bytes:

| Fixture | Before Ir | Final Ir | Change |
| --- | ---: | ---: | ---: |
| `strings.a` | 193,000 | 192,301 | -0.36% |
| `strings_more.a` | 281,106 | 280,992 | -0.04% |
| `string_positions.a` | 559,995 | 539,304 | -3.69% |
| `string_index.a` | 247,162 | 246,603 | -0.23% |
| `string_append.a` | 11,994,860 | 11,999,664 | +0.04% |
| `shared_slices.a` | 155,370,965 | 157,729,171 | +1.52% |
| `lone_surrogates.a` | 184,672 | 184,026 | -0.35% |

The shared-slice regression remains. The cache test and backward-cursor
selection add work on mixed supplementary strings that cannot use the BMP
view; the fixture's aggregate does not isolate each cause. The measured
scanner improvement supports the tradeoff, not a claim that every string
workload improved. Initial profiles exposed larger append and shared-slice
regressions; avoiding `free(NULL)` and retaining the forward cursor path
reduced them to the final numbers above.

The counted scanner prints the same row before and after:

```text
adamic: counts: allocations 432199 frees 432199 retains 2450881 releases 2642078 peak 2262 regions 0
```

Every existing `counts.md` row passes without an update. These are logical
heap-value counters; the extra BMP cache bytes are not measured by them.
Raw profiles and machine-readable results are in [utf16-views/](utf16-views/).
Wall-time samples were taken while other checks ran, so no wall-time gain
is claimed.

## Validation

The native sweep now runs mixed ASCII/BMP/supplementary strings, BMP-only
strings with lone high halves, BMP-only strings with lone low halves, and
ASCII strings, from 20 to 2,000 points and 16 generator seeds. It reads
forward, backward, jumping, and interleaved with nearby slices. Small
strings get every slice; larger strings get scattered slices and searches.
WTF-8 hexadecimal output preserves lone halves for exact comparison with
Node: 30,828 lines, zero mismatches. The existing numeric string sweep
compares 11,460 answers, including NaN, infinities, fractions, negative and
out-of-range positions. The append test compares 43 lines, covering static
literals, long stack pieces, and a pointer comparison proving in-place
append when a cached lone high half joins an appended low half.

Commands used on the final implementation, with all test output redirected:

```sh
source /workspace/adamic-tools/env.sh
gofmt -l cmd internal > /tmp/utf16-final-gofmt.log 2>&1
go vet ./... > /tmp/utf16-final-vet.log 2>&1
go test -count=1 -timeout 30m ./internal/native > /tmp/utf16-final-native.log 2>&1
go test -count=1 -v -run 'TestNativeAgreesWithNode/internal/oracle/testdata/(strings|strings_more|string_positions|string_index|string_append|shared_slices|lone_surrogates)\.a$' -timeout 30m ./internal/oracle > /tmp/utf16-final-oracle.log 2>&1
go test -count=1 -timeout 30m -run '^TestCountsAreRecorded$' ./internal/oracle > /tmp/utf16-final-counts.log 2>&1
ADAMIC_TYPESCRIPT_SOURCE=/workspace/utf16-measure/typescript ADAMIC_SCANNER_PROFILE_SNAPSHOTS=/workspace/utf16-measure/final go test -count=1 -v -timeout 30m ./stage1/typescript/scanner > /workspace/utf16-measure/scanner-final-validation.log 2>&1
```

Observed: native PASS in 208.680s; filtered oracle PASS in 60.955s (nine
fixtures, including optional and undefined strings); counts PASS in
273.299s; scanner PASS in 146.642s. Formatting and vet produce no output.
The oracle compares source Node, JavaScript backend, sanitized native and
release native and checks leaks. The scanner checks 77 compiler files,
61 stage1 files and 18,236 generated inputs: 23,816,619 identical answer
bytes, with final release and profiled binaries also held to Node and Go.

The full repository gate was also run on the first implementation:
`go test -count=1 -timeout 30m ./... > /tmp/utf16-gate.log 2>&1`.
It passed, including the whole oracle (961.706s). Later improvements were
rechecked with the complete native package, all counts, the filtered oracle
and the complete scanner package as listed above. The full repository gate
was not repeated after those final runtime-only improvements.

## Mutants

All three requested mutants were run and restored in the separate
`/workspace/utf16-measure/mutants` worktree against the final implementation.
Each compiles; none is killed by a compiler warning.

| Mutant | Check and observed failure |
| --- | --- |
| Keep the old index and cursor across an in-place append, resetting only the unit count | `TestStringViewAfterAppendMatchesNode`: ASan heap-buffer-overflow in `adamic_string_char_code`, reading the stale BMP view |
| Count a supplementary point as one UTF-16 unit | `TestStringIndexMatchesNode/mixed`: 7,206 native lines versus 8,940 Node lines |
| Return the forward cursor's offset unchanged on a backward step | `TestStringIndexMatchesNode/mixed`: normal execution, 7,653 of 8,940 lines differ from Node |
| Increase the Callgrind summary by one | `profile.py --summarize-only`: `self costs do not sum to summary` |

Logs, exact substitutions and the runner are saved beside the measurements.
Trailing whitespace was removed from committed text logs; the original
outputs remain in scratch and `/tmp`.
The scanner package additionally reruns its three comparison controls:
wrong `!=` token kind, accepted repeated numeric separator and omitted
regex rescan, each caught by Node and native byte comparisons.

## Reproduction and setup

`bash cloud/setup.sh > /tmp/utf16-setup.log 2>&1` passed. Its timing lines:

```text
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (1s)
setup: node ready (1s)
setup: submodules ready (1s)
setup: build cache warm (13s)
setup: done in 13s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

`nproc` printed 5. Go 1.27.1, Node 24.19.0. The configured tool directory
was `/workspace/adamic-tools`, so that directory's `env.sh` was sourced
rather than `/opt/adamic-tools/env.sh`. Setup had no failure.

Valgrind was not installed. The Debian package
`https://deb.debian.org/debian/pool/main/v/valgrind/valgrind_3.24.0-3_amd64.deb`
was downloaded and extracted with `dpkg-deb -x` into scratch.
Git cloning TypeScript from both upstream and the configured fork failed:
`fatal: could not read Username for 'https://github.com': No such device or address`.
The bundled fork was a different commit. The exact pinned commit was
instead downloaded from
`https://codeload.github.com/microsoft/TypeScript/tar.gz/050880ce59e30b356b686bd3144efe24f875ebc8`;
archive SHA-256:
`43d388f7fc6511ae105f23f4779e1c8df96f6dfa003025b70249816089609f5b`.
Only the scratch scanner test's Git HEAD check was replaced with a comment
identifying that exact archive. No corpus content or scanner code changed,
and no Git commit ID was fabricated. This scratch adaptation is not committed.

For each snapshot, `TestProfileArtifacts` saved the release, counted and
`-O2 -g` executables and runtime sources. The baseline used main's runtime;
the final snapshot copied only `string.c` and `string_index.c` from this unit.
Actual profiling commands:

```sh
ADAMIC_TYPESCRIPT_SOURCE=/workspace/utf16-measure/typescript ADAMIC_SCANNER_PROFILE_DIR=/workspace/utf16-measure/before go test -count=1 -v -run '^TestProfileArtifacts$' -timeout 30m ./stage1/typescript/scanner > /workspace/utf16-measure/before-build.log 2>&1
ADAMIC_TYPESCRIPT_SOURCE=/workspace/utf16-measure/typescript ADAMIC_SCANNER_PROFILE_DIR=/workspace/utf16-measure/final go test -count=1 -v -run '^TestProfileArtifacts$' -timeout 30m ./stage1/typescript/scanner > /workspace/utf16-measure/final-build.log 2>&1
VALGRIND_LIB=/workspace/utf16-measure/valgrind/usr/libexec/valgrind python3 stage1/typescript/scanner/profile.py /workspace/utf16-measure/before --valgrind /workspace/utf16-measure/valgrind/usr/bin/valgrind > /workspace/utf16-measure/before-profile.log 2>&1
VALGRIND_LIB=/workspace/utf16-measure/valgrind/usr/libexec/valgrind python3 stage1/typescript/scanner/profile.py /workspace/utf16-measure/final --valgrind /workspace/utf16-measure/valgrind/usr/bin/valgrind > /workspace/utf16-measure/final-profile.log 2>&1
python3 /workspace/utf16-measure/fixtures.py > /workspace/utf16-measure/fixtures-final-profile.log 2>&1
python3 /workspace/utf16-measure/mutants.py > /workspace/utf16-measure/final-mutants.log 2>&1
```

Artifact builds passed; all runs printed 434,790 tokens. The committed
fixture and mutant scripts preserve the actual scratch paths and commands.
They require the setup and snapshots above. The repeat invoked Valgrind
on `final/profiled` with the same manifest and wrote `final/repeat.callgrind`.

Not covered: an allocated-byte/RSS benchmark, multicore string sharing
(the runtime is single-threaded), OOM behavior, strings near V8's maximum
length, or a broad application corpus beyond the scanner and existing
oracle. No claim of a wall-time improvement or an improvement to every
string access pattern.
