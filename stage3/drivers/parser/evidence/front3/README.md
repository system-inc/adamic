All requested inputs are merged in the detached, unpushed scratch worktree:
area/stage3 4ad53a4, compiler/stage3-front-2 80fb9b7, explicit import-cycles
779ff9d, and library/qualified-as-const 8280fd0. Import-cycle merge f79114e,
Node-loader merge d9d9ebd. Compiler build succeeds with cohere 7945d102 and
its TypeScript d92d9bfee SDK. The complete compiled scratch source difference
from front-2 is retained as scratch-integration.patch.gz; this evidence patch
is not applied to the parser branch. It includes scratch Go module path fixes.

Conflict resolutions:

- Import cycles: only internal/oracle/counts.md conflicted; retained front
  counts. No textual lowering conflict in this merge.
- Node loader: internal/load/load.go and source_fs.go retain the new typed-path
  API while enabling pinned Node declarations and the Node-aware prelude.
  Original roots use len(paths), excluding both prelude and Node declaration
  roots. Convert the resolved Node index into a RootedFilePath. Reconstruct
  the Node program using the new NewParsedCommandLine/compiler-host signatures.
- Lowering conflict: internal/lower/cast.go retains the front's castProof and
  checked tag/class cast path, including the empty-proof fast path. Preserve
  library's IsIdentifier guard for genuine as const in cast_proof.go, where
  the front's proof now recognizes it. Qualified types named const are not
  treated as the built-in assertion. No unconditional cast bypass is added.
- Oracle counts conflicted again; retained front counts. Full count gate is
  not claimed for this scratch integration.
- Additional SDK compile adjustments: convert typed filenames to string at
  node_library.go's string predicates and lower/library_node.go's path.Base.
  The latter touches lowering but was a compile adjustment, not a Git conflict.

`GOWORK=off go test ./internal/load ./internal/lower -run
'Test.*(Node|Cast|ImportCycle)' -count=1` passes both packages. Compiler build
also succeeds. Full compiler/ASan/upstream suite is not rerun for this scratch
merge. No compiler edits are committed on the parser branch.

Driver change: import type {} from node:fs activates the loader's actual pinned
@types/node 25.3.3 seat, as in the earlier checker-clean run. It is erased by
Node and adds no executable imports. All five validated temporaries 60-64
remain applied to the fresh gathered slice. Entry counts remain 1,990 code
declarations / 26 code files; additional diagnostic-driver root gives 1,991
in 27 files. Prior raw copied-span/import-prefix byte audit still applies.

Build now passes checking and stops in lowering:
core.ts:11:52: stage 0 can't lower an indirect call or class construction before
enum initialization; declare enums before executable module code yet.
The statement is export const emptyMap: ReadonlyMap<never, never> = new Map<never, never>().
Minimal native-enum-map.a reproduces that exact NotYet class. The merged
flag-enum Node runner (transform mode) executes it with stdout 0, empty stderr,
exit 0. The parser branch's older strip-only runner cannot run enums; its
initial unsupported-enum result is not counted as the oracle observation.

Native parser binary, native parser timing/size, actual parser output diff,
and parser-output mutant remain unrun. A separate tiny native comparator
control from this integrated compiler prints parser comparison control; self
cmp exits 0, and a one-byte output flip exits 1. This is not parser output.

All three Node runs on the fixed full corpus and manifest match 35,456,964 bytes,
SHA256 2014ef06f9db928b50d001787c44e490bbdbd8ecd9e912d3f676d58148ffefc5.
Best Node user time 5.727305 seconds, other runs 5.888736 and 5.916100.
These include source load/transpilation/printing. perf is not installed.
The Identifier end+1 mutant is still caught at dump line 164.

Commands are the same native-proof.py protocol as front2, using the newly
built /tmp/parser-front3-adamic, /tmp/parser-front3-node and a fresh output
/tmp/parser-front3-native. build.stderr.gz holds the exact first native error.
