Built the isolated Unicode property table layer; the runtime pattern compiler is unfinished.
Commit: this report is committed with the property table implementation on codex/regex-runtime-compiler.
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
