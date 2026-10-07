Built the isolated Unicode property table layer; the runtime pattern compiler is unfinished.
Commit: df4fd3e, pushed on codex/regex-runtime-compiler, contains the property table implementation.
Checks: sanitized C/Go entry identity, wasm32/WASI identity, regeneration and no-compiler symbol checks pass.
Mutants: range entry, invalid alias acceptance, unwanted linking, WASI-only lookup and stale generation all fail their intended checks.
Not covered: pattern parser, bytecode compiler, runtime SyntaxError, divergence rulings, lowering and dynamic lint fixtures.

## Implemented layer

`regexp_compile_properties.c` resolves counted WTF-8 property names by exact alias,
including the v-only properties of strings. The generated tables contain 441 code-point
sets and seven string properties. The Python generator reads the Go UCD table source
and carries its input digests. It downloads no second Unicode dataset. The Go test
enumerates aliases with the Go AST and obtains expected entries through the public
Go property provider. Compiled C entries are compared individually, including every
range endpoint and every string code point, rather than through a digest.

Identity checks cover 1,722 aliases, 448 unique properties, 23,045 ranges and 7,906
strings. They also cover Unicode mode restrictions, 5,175 rejected spellings, and
counted input including NUL. Linux executes with AddressSanitizer, UndefinedBehaviorSanitizer
and LeakSanitizer. The same entry comparisons compile with wasm32 pointers and
`ADAMIC_TARGET_WASI=1`, with no libc or atomics, and run under Node's preview1 WASI.
The WASI execution itself is not sanitizer instrumented; Linux is the sanitizer gate.

The existing runtime build uses `--whole-archive`. This layer is explicitly gated by
`ADAMIC_REGEXP_RUNTIME_COMPILER`: normal runtime builds contain no compiler tables or
lookup implementation. A sanitized empty program links the real runtime, executes,
and is checked with `nm -a`: no `regex_compile_` symbols. Integrating the opt-in with
actual dynamic programs is still unfinished.

## Measurements

Pinned test262: `7ab7fafa0003f73fc85c1b95d88094d33f7eb8bd`, adaptation enabled,
whole `built-ins/RegExp` directory, Linux sanitized runtime.

| Measurement | Pass | Disagreement | Refused | Crashed | Skipped | Total |
|---|---:|---:|---:|---:|---:|---:|
| before | 53 | 0 | 424 | 73 | 1329 | 1879 |
| after | 53 | 0 | 424 | 73 | 1329 | 1879 |

There are zero newly passing tests. The 73 runner-classified crashes already exist
at the baseline; several are compiler diagnostics for invalid RegExp patterns.
These measurements do not establish a green runtime compiler unit.

Toolchain setup printed Go ready 0s, clang ready 0s, Node ready 0s, submodules
ready 0s, build cache warm 107s and done 107s. `nproc`: 5. CPU quota:
`400000 100000`. Environment: `/workspace/adamic-tools/env.sh`. Go 1.27.1,
clang 20.1.8, Node 24.19.0.

## Mutants actually run and restored

| Mutation | Check and observed failure |
|---|---|
| ASCII endpoint 0x007F changed to 0x007E in compiled tables | `TestRegExpRuntimePropertyIdentity`: entry identity failed at alias 2, exit 1 |
| C lookup accepts `ascii`, which Go rejects | Same test: rejected alias 1725 accepted, exit 1 |
| Remove the compiler opt-in guard | `TestRegExpRuntimeCompilerNotLinked`: compiler symbols linked, exit 1 |
| Return NULL for ASCII only when compiled for WASI | `TestRegExpRuntimePropertiesWASI`: WASI trap, exit 1 |
| Change the generated header's provenance comment | `TestRegExpRuntimePropertiesGenerated`: stale table output, exit 1 |

The entry and acceptance mutants ran only the C/Go comparison, so the regeneration
check could not mask the intended failure. No mutant was killed by clang diagnostics.
These are table-layer mutants, not the requested bytecode-emission and pattern
SyntaxError mutants, which require the unfinished compiler.

## Commands and logs

All test output was written directly to files, without piping a running test.

```sh
bash cloud/setup.sh > /tmp/regex-runtime-setup.log 2>&1
source /workspace/adamic-tools/env.sh
python3 internal/regexp/testdata/generate-runtime-properties.py
go test ./internal/native -run '^TestRegExpRuntime' -count=1 -v > /tmp/regex-runtime-properties.log 2>&1
go test ./internal/native ./internal/regexp ./internal/lower -count=1 -timeout 30m > /tmp/regex-runtime-packages-clean.log 2>&1
go vet ./... > /tmp/regex-runtime-vet.log 2>&1
gofmt -l cmd internal
git diff --check
go run ./cmd/adamic-test262 -adapt -json -test262 /tmp/regex-runtime-test262 built-ins/RegExp > /tmp/regex-runtime-before.json 2> /tmp/regex-runtime-before.log
go run ./cmd/adamic-test262 -adapt -json -test262 /tmp/regex-runtime-test262 built-ins/RegExp > /tmp/regex-runtime-after.json 2> /tmp/regex-runtime-after.log
```

A first package run passed, but briefly overlapped restored-source mutant work.
The final package log is a separate clean rerun after all mutations were restored.
The clean package run and `go vet ./...` exited 0. `gofmt -l cmd internal` and
`git diff --check` printed nothing. Full repository tests and the dynamic
Node/native/JavaScript oracle were not run.

The requested dependency branches were fetched for inspection: corpus
`c6487b485443a71f91f01b7a2568a44a9cee1ddc`, divergence rulings
`ab512274afa10166a38caef54ad55a7fd1f75700`, lint
`b39979305d880d7d05083704b73d405cd3694d68`. Their compiler changes were not merged.
Base main: `c01907a`. The cohere corpus was not run through a C pattern compiler.
There is no callable group messaging tool for `#adamic_runtime_platforms`; the user
was asked to relay the platform requirement. No platform-group coordination is claimed.

The package, platform, setup and mutant logs are retained beside this report as `.txt` files.

October 7 integration: merged origin/codex/runtime-statics e2422008 in
606b066. The unchanged generated aliases now use a named const struct type so
the storage scanner recognizes their immutability. TestRuntimeStaticsAreListed
passes directly. The merged package has an unrelated map_hash_test.go Options
field spelling error (slabs versus Slabs); a temporary Go test overlay corrects
only that test spelling. With this overlay, regeneration, sanitizer-backed entry
identity, WASI entry identity, unused-symbol proof and the statics guard pass in
8.172s. Full output is properties-statics-merged.txt. No shared compiler or test
file was edited for this workaround. Parser work is not yet certified.

Release size follow-up (table layer only): sizes-table-layer.json compares the
runtime-statics integration baseline e2422008 with 19c58e8. Each baseline and
branch build uses matching output basenames, since WASI records that basename
in its name section. Commands on each checkout, with the environment sourced:

    adamic build --target wasm32-wasi internal/load/testdata/0.1/compile/01_hello.ts -o hello.wasm
    adamic build --target wasm32-wasi cmd/adamic/testdata/wasi/request.a -o request.wasm
    adamic build internal/load/testdata/0.1/compile/01_hello.ts -o hello
    adamic build cmd/adamic/testdata/wasi/request.a -o request

hello.wasm is byte identical at 245068 raw / 71934 Brotli quality 11 bytes;
request.wasm is byte identical at 288915 raw / 87420 Brotli bytes. Native raw
sizes are unchanged (380584 / 381312). Native binaries embed random runtime
cache build-directory names in assertion strings, so their compressed sizes do
not agree; those exact measurements are recorded without claiming a pass.
These merged-baseline hello sizes do not reproduce the platform's 14236-byte
release measurement. A dynamic compiler fixture is not yet lowered or measured.

Parser development probe: 5746 test262 patterns and 4000 seeded cases agree on
acceptance and full Node diagnostic messages under ASan and UBSan. Adding the
875 pinned cohere patterns leaves exactly one contract blocker, captured in
parser-contract-blocker.txt: Go rejects a{9223372036854775808,9223372036854775807}
and Node accepts it. C follows Go and emits the Node-shaped rejection message,
but Node has no SyntaxError message for a pattern it accepts. The parser is not
certified or committed. Separate probes find Go rejecting valid Other_ID_Start
and Other_ID_Continue capture names, and accepting an overflowing Unicode
escape Node rejects. Reference edits need a scope ruling under the user's
restriction on Go compiler changes. Bytecode, parser WASI, new compiler mutants,
runtime divergence refusals and lowering remain unfinished.

Rulings checkpoint: merged area/library fe0e7aa895266bd3d553ba4f148b0f5c0223ce01
in a7f6f64 with a merge commit. Conflicts retain atomic string-index publication,
Map's atomic iteration count plus the small-table layout, both fixture registries
and both flow exclusions. The library's UTF-16 cache API borrows the already
published immutable unit view. Oracle counts regeneration is running separately.

The user's reference rulings supersede the earlier contract-blocker paragraph.
Go now admits Other_ID_Start and Other_ID_Continue capture-name characters and
bounds each braced Unicode-escape arithmetic step before overflow. Exact reversed
quantifier MVs whose clamped bounds would compare equal return V8DivergenceError
with the ruled reason and section. The merged clamping acceptance approximation
was removed. The C parser follows these corrections and has a distinct divergence
status, rather than pretending Node throws SyntaxError for an accepted pattern.

Partial parser: regexp_compile_parser.c/.h own a per-call AST allocation arena;
there are no mutable statics, caches or startup work. The oracle compares 5746
existing test262 patterns (including the original 2776), 875 cohere patterns,
4000 seeded random shapes and four ruling probes. Linux ASan/UBSan and WASI SDK
27 both pass complete messages, acceptance and divergence reasons. The WASI
command uses -DADAMIC_TARGET_WASI=1 -mno-atomics and runs through Node's WASI host.
The cohere tuples in runtime-cohere.json were extracted from the inventory at
c6487b485443a71f91f01b7a2568a44a9cee1ddc, preserving pattern, flags and location.
Output: parser-linux-wasi.txt (6.862s). Native package execution still uses the
previously documented temporary overlay for the unrelated Options.slabs typo.

Reference proof: the 432-case large-bound sweep and Node ruling probes pass.
Five source-overlay mutants each build and reach the intended failed assertions:
drop Other_ID_Start, drop Other_ID_Continue, restore Unicode integer wrapping,
silently accept clamped reversed bounds, classify that divergence as SyntaxError.
Output: reference-rulings.txt (3.355s). No mutation remains in the working tree.

This is the requested partial parser checkpoint, not a compiler-identity claim.
AST and bytecode identity, compiler emission mutants, remaining V8 shape refusals,
dynamic lowering, dynamic fixture sizes and full post-merge gates remain pending.

Bytecode layer: the per-call C set and instruction compiler agrees byte for byte
with Go's pointer-independent native instruction snapshot for 7,430 programs.
The shared corpus includes 5,746 test262 patterns, all 875 pinned cohere patterns,
4,000 seeded patterns and four ruling probes; 3,195 cases reject or refuse.
Linux runs ASan/UBSan; wasm32-wasi runs the same identity probe with no atomics.
Changing SET to ASSERT in C is caught at case 0, byte 288. Making C accept `?`
is caught at case 4,700, status byte 0. Combined compiler/parser/property identity,
mutants, unused symbol proof and statics guard pass in 26.115s. Repository Go vet
passes with the documented map_hash_test.go field-rename overlay.

Linux RegExp directory after library merge and rulings, measured with
cmd/adamic-test262: pass 53, disagreement 0, refused 424, crashed 73, skipped
1,329 (1,879 total). Before: pass 53, disagreement 0, refused 424, crashed 73,
skipped 1,329. There are no newly passing directory tests at this layer.
The reversed oversized quantifier fixture now refuses through V8DivergenceError;
its filtered uncached Node oracle passes. Optional matcher search optimizations
are not part of the bytecode snapshot; the existing general VM consumes these
instructions. Runtime V8 shape checks and dynamic lowering are still pending.

Runtime V8 refusals: the C compatibility walk ports internal/regexp/v8.go.
The corpus plus 400 scoped/class-string probes compares accepted bytecode and
exact refusal messages: 7,674 compiled programs, 3,351 rejected/refused cases,
on sanitized Linux and wasm32-wasi. Bypassing the compatibility walk is caught
by `[\q{a}]/iv` at case 10,676. Combined compiler/property/parser identities,
all C mutants and the statics guard pass in 37.392s; touched-package vet passes.
The checked compilation entry point is mandatory for dynamic construction.
