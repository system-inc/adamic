# Native build flags

The policy is in internal/native/native.go: Flags supplies compilation
arguments and LinkFlags passes that entire list to compilation/linking.
adamic build program.a -o program selects Options{Release: true}.
Only native, uncounted, unsanitized shipped builds add -flto=thin, on every
generated C unit and every separately archived runtime unit. Linux uses
-fuse-ld=lld at the link; macOS uses its ThinLTO-capable platform linker.
The clang, bitcode archiver and linker must support the same LLVM bitcode.
An unavailable linker or unsupported optimization is a build error.

Tests, fuzzing and ordinary oracle calls leave Release false.
--count and --sanitize retain their exact previous flags even when the
command selected Release. WASI retains its existing target policy.
Native --tsgo also passes LinkFlags; its external Go checker archive
remains machine code, while generated C and Adamic runtime units use ThinLTO.

The semantic options are:

- -std=c11: the emitted C's language dialect.
- -ffp-contract=off: every multiply/add rounds separately as on Node.
- -fno-optimize-sibling-calls: tail recursion keeps frames and eventually
  gives a stack panic as on Node, rather than becoming an endless loop.
- Conditional -DADAMIC_COUNT, -DADAMIC_SLABS, -march=..., and WASI's
  --target=wasm32-wasi, --sysroot=..., -DADAMIC_TARGET_WASI=1,
  -mno-atomics also reach the link whenever present.

Warning flags, optimization level and sanitizer options also reach the link.
There is no separate semantic-option allowlist to fall out of date.
The link compiles the generated main C and runs ThinLTO code generation.
The link-only tail-call mutant removes the option from that invocation,
while retaining it on every runtime compilation. The unbounded recursion and
self tail call fixtures must still match Node's stack panic and exit 70.
The arithmetic fixture internal/oracle/testdata/release_fma.a loads operands
from program arguments, then exercises three rounding boundaries. Its mutant
removes contraction protection only from the main compilation/link invocation.
On x86, TestReleaseFixtureKeepsArithmeticUnfused uses -march=haswell when the
host reports FMA; on arm64 FMA is always available. Node and the protected
ThinLTO build print three zeros. The missing-option mutant prints
5.551115123125783e-17, -5.551115123125783e-17 and -8.326672684688674e-19.

TestRuntimeCompilesEveryUnitUnfused makes cold runtime.a builds for shipped
release, ordinary sanitized oracle and counted oracle policies. It audits the
actual clang command for every runtime C unit, including dtoa.c and ieee754.c,
and requires the effective contraction option to be -ffp-contract=off. A real
dtoa.c-only omission mutant must compile successfully and fail that audit.

The default test/oracle policy remains -O2, or
-O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all when sanitized,
with existing count/slab/CPU options in their existing order.
TestNonShippingFlagsStayIdentical compares argument bytes and order
against independent literals from before ThinLTO.

## Shipped release oracle

The additional lane is opt-in and skipped in ordinary test runs:

~~~sh
source /workspace/adamic-tools/env.sh
ADAMIC_ORACLE_RELEASE=1 ADAMIC_GATE_UNCACHED=1 \
  go test -count=1 -timeout 30m ./internal/oracle \
  -run '^TestRelease(AgreesWithNode|OracleCatchesOneByte)$' \
  > /tmp/adamic-release-oracle.log 2>&1
~~~

Run it nightly or once per main push. Every registered oracle fixture is
checked/lowered again, every supported fixture is built with the exact shipped
release policy, and stdout, stderr and exit code are compared byte for byte
to source on Node. Fixtures intentionally exercising Adamic's inserted checks
use the JavaScript backend as the ordinary oracle does, and must distinguish
it from unchecked Node. Unsupported fixtures must still refuse explicitly.
The lane includes an actually compiled one-byte output mutant.
Specialized stream, permission, leak and counter probes keep their existing
ordinary flags; this extra lane covers the complete registered fixture table.

runtime.a caching includes the complete compilation flag list, source/header
bytes and compiler identity, so release bitcode cannot reuse an ordinary
machine-code archive. The cached archive saves runtime frontends; ThinLTO
backend work is still paid at each program link. Linker-only options stay out
of runtime clang -c. Developer-tool translation-unit splits must use the same
Flags/LinkFlags policy and include that policy and toolchain in object keys.

## Stage 1 executable profiles

Only `go run ./cmd/adamic-stage1 -driver parse|lint -o <binary>` selects a
committed profile. Ordinary `adamic build` keeps plain ThinLTO. The source
paths are stage1/cohere/parse/parse.a and stage1/cohere/lint/main.ts.
Profiles and their manifests live beside each driver in profiles/<os>-<arch>/.
The committed format is LLVM's text format. The build converts it with the
local llvm-profdata into a hash-addressed indexed copy in the user cache.
Every runtime object and the link use that same immutable copy. Its content
identity participates in runtime.a's existing flag-based cache key.

The manifest hashes the exact emitted C, every embedded runtime C/header,
the text profile, flags, target and compiler identity. A missing or changed
input produces one line saying the build uses plain ThinLTO. The stale
profile never reaches clang. The check uses bytes, never modification times.
A different Apple clang or target regenerates its own profile; this does not
silently reuse a Linux profile. Unsupported text conversion also falls back.

Regenerate both drivers from the committed training list:

```
source /workspace/adamic-tools/env.sh
python3 stage1/profiles/regenerate.py --typescript <pinned TypeScript checkout> \
  --work <scratch directory> --llvm-profdata <matching llvm-profdata>
```

The script builds instrumented release binaries with the same semantic flags,
runs only the fixed training corpus, merges to text, writes manifests and
builds the profile outputs. stage1/profiles/training.json names paths/hashes
outside the entire 77-file compiler benchmark set in benchmarks.json. Both
regeneration and TestTrainingNeverIncludesBenchmarks reject path or content
hash overlap. Every added speed corpus must be registered there first.

The opt-in release oracle now includes the actual shipping stage 1 artifacts.
Set ADAMIC_TYPESCRIPT_SOURCE to that pinned checkout and
ADAMIC_STAGE1_PARSE_BINARY / ADAMIC_STAGE1_LINT_BINARY to the regenerated
parse-profile / lint-profile paths when running the release oracle command
above. It compares those exact binaries to Go and Node and rebuilds them to
check binary determinism. It is an error to enable that lane without its
shipping binaries. Ordinary tests keep their flags and skip this extra lane.
