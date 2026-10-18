# Native builds

`adamic build program.a -o program` selects `Options.Release`. Native builds
without counters or sanitizers use `-O2 -flto=thin`. Linux links with lld;
macOS uses its platform linker. clang, llvm-ar and the linker must support
the same LLVM bitcode. Missing tools fail the build.

`Flags` supplies compilation flags; `LinkFlags` keeps every flag at the link,
where ThinLTO generates code. `-ffp-contract=off` preserves Node rounding and
`-fno-optimize-sibling-calls` preserves stack exhaustion checks. Linker selection
never reaches warning-strict runtime compilations. Split units, supplied runtime
archives and the checker bridge retain their existing paths and feature flags.
The external Go checker archive remains machine code.

Ordinary tests and fuzzing leave Release false. Counted, ASan, TSan and WASI
builds keep their prior flags even when Release is true. Runtime and split-unit
cache keys already include every compilation flag, keeping bitcode distinct
from ordinary objects. Shipped oracle archives have their own declared variant.

Release tests compare arithmetic and recursive programs to Node. The FMA fixture
loads operands from arguments; on an FMA-capable host its native test also removes
contraction protection from the program link and checks that the output differs.
The oracle's one-byte output mutant proves the release comparison can fail.
Each release witness and runtime product group has its own top-level test.
The historical optional whole-fixture release lane is deferred until it has
bounded family ownership; these tests cover the release policy's direct witnesses.
