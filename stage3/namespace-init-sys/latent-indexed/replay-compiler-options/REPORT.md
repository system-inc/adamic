Merged current main on the feature branch; the requested three-fix merge and new census are blocked.
Compiler measured for this reduction: 0d07625ce973ba264fbefd7a56168e7cc39dc5b8.
Node: 1; native and JavaScript controls: 1; original and minimal replay: exact NotYet reproduced.
Mutant: remove only the index signature; replay exits 1 with "replay signature did not reproduce".
Not covered: the three requested feature merges, the new 99-row census, and changed blocker rankings.

The previous census remains 2 emitted / 76 blocked / 21 not reached at 5ad36d2c. These are historical counts, not a measurement with the three requested fixes. See ../evidence/REPORT.md for its full table. No new blocker table can be inferred from that table: lifting a stop can expose a different stop.

The main merge is 0d07625c, from origin/main 679af4df. An ordinary merge of enum 64952252 produced 161 conflict hunks in 55 files, and catch ded85b2d produced 23 in 18. Both attempts were aborted without discarding feature changes. Class 696f668c has not been merged. The enum and catch conflict inventories are attached. No push to main or an area branch occurred.

Automatic approval review rejected the broad -Xours merge because it could silently discard substantial portions of the requested enum fix. It also rejected broad scripted catch resolutions across lowering, IR, loader, runtime and oracle because they could silently lose namespace/readiness or catch safety semantics. Neither rejected strategy was executed. A namespace-compatible resolved integration base is needed before measuring the requested combination safely.

CompilerOptions reduction

The requested replay from 9a1f14c5 reproduced the original NotYet at transformers/classFields.ts:365:11 in transformClassFields, with reason "a value of type CompilerOptions". Its complete output is original-replay.log.gz. The replay ran against an isolated worktree of 0d07625c, with replay tooling from 9a1f14c5 and the attached compatibility patch. Cohere was referenced through a symlink to the existing submodule; no code was copied from it.

Replay command, from the isolated worktree:

    source /workspace/adamic-tools/env.sh
    go run ./stage3/census/latent/replay -project /tmp/namespace-init-sys-typescript/src/compiler -where /tmp/namespace-init-sys-typescript/src/compiler/transformers/classFields.ts:365:11 -kind NotYet -reason 'a value of type CompilerOptions'

Original replay exit 0, load/register/lower 2m53.273269674s, total including cold Go compilation 16m12.53128082s. The production source is the same adapted 79-file tree and source hashes as the prior census.

Exact minimal program: minimal.a. Its checker-clean function attempt fails at minimal.a:5:18 with the same NotYet reason. Repeat with -project minimal.a and -where minimal.a:5:18; the final run exits 0 in 945.327824ms total (72.063634ms lowering). Removing only the string index signature, without-index.a, exits 1 under the same replay pin (use :4:18), proving the pin can fail.

The original interface has named properties and at types.ts:7582:

    [option: string]: CompilerOptionsValue | TsConfigSourceFile | undefined;

The representation gate in internal/lower/expression.go asks for indexed dictionary storage. internal/lower/records.go requires a mutable string-index dictionary to have zero named properties; readonly dictionary acceptance also excludes named properties. Consequently this mixed shape has no supported representation. Ordinary production refusal is more specific:

    stage 0 can't lower an index signature beside named members, or a non-mutable unrestricted string signature (dictionary storage cannot preserve named-property contracts) yet

One named optional numeric property plus a numeric index is sufficient to reproduce. The complex CompilerOptionsValue union is not required for this stop, but could expose further boundaries after a mixed-shape implementation exists. No refusal was lifted.

Node runs all three programs and prints 1. Both compiler backends refuse minimal.a at 3:5. Both lower without-index.a and index-only.a and print 1, matching Node. comparisons.json contains exact exit codes, diagnostics and outputs. JavaScript runs use oracle/node.mjs, the repository runtime resolver; native and JavaScript compilation use the isolated 0d07625c compiler. An earlier JS invocation without that resolver failed package resolution and was discarded.

The prior evidence actually attributes CompilerOptions enclosing stops to five locations totaling 15 rows (classFields 5, generators 3, module/esnextAnd2015 3, checker 2, transformer 2). This reduction selects the classFields group. The requested 11-row description does not match that stored table; it does not alter the 99-row conservation.
