Added driver-local JSON string escaping and narrowed payload formatting.
Topic based on origin/main 89ac4a8c, on codex/stage3-scanner-native-3.
Fixture: Node and native match JSON.stringify; complete scanner bytes unchanged.
Dropped newline escape builds/runs and fails both Node/native byte comparisons.
No compiler/adaptation change, full scanner native run, or full package gate.

The driver no longer calls JSON.stringify. The escaper covers quotes, backslash,
all short JSON escapes, lowercase \uXXXX for other controls and lone surrogates,
and preserves valid surrogate pairs. Number payloads use String after narrowing,
with non-finite values represented as null. Undefined payloads/cooked values
retain null, and finite -0 formats as 0. No scanner source is changed.

escape-fixture.a supplies 140 string cases: every code unit 0..0x7F separately,
all 128 combined, and 11 empty/Unicode cases including three surrogate pairs,
an embedded pair, lone/reversed surrogates and U+2028/U+2029. Five payload cases
cover text, integer, -0, fractional number and undefined. The harness copies
only the exact parsed helper declarations from the real driver into a scratch
.a program and runs the same fixture against independent Node JSON.stringify.
No manually duplicated escaping implementation is used by the fixture.

The native control uses the existing unchanged cloud/land-area-next 28285421
compiler. Fixture build/run/diff exit 0. The newline branch mutant emits the
raw character instead of its escape; source Node exit 0 and diff exit 1.
Native mutant build and execution both exit 0, then diff exits 1. It is caught
by output comparison, not compiler diagnostics or -Werror.

Full scanner Node run matches the previous full-tree reference byte for byte:
81 src/compiler inputs, 509,014 skipped-trivia tokens, 860,418 retained-trivia
tokens, 466 error rows, 108,019,935 bytes. SHA256:
5cce1570354cc48b5d9db246daf2abe8db3c21edba3c35e613da812a26920182.
The copied-output comparison control passes and the token-end mutant fails.

An additional check copies the actual updated driver into the existing
uncommitted throwing discovery copy. The former own JSON.stringify gate is
gone; clang now reaches the previously witnessed Object*/string* return error
at generated main.c:7943:9, dependent on the earlier never-return placeholder.
This is a discovery measurement, not a native scanner correctness claim.
The first attempt used the delivery checkout as cwd and failed its pinned
Node declaration prerequisite; using the compiler checkout fixed the cwd.

Setup: nproc=5, quota=4 cores, Go 1.27.1, Node 24.19.0, clang 20.1.8.
Timing lines: Node .035s, Go .040s, markdown .097s, submodules .098s,
clang .216s, Go build 35.556s, deferred test binaries 35.726s,
warm cache 35.728s, total 35.761s. GOPROXY was set before setup.
Initial scratch fixture attempts caught wrongly escaped source literals and
then a mutant-selector spelling mismatch; both were fixed before the final
fixture/native checks. No failing source was committed.

Commands actually run, with all test output sent to logs:

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/scanner-escape-setup.log 2>&1
source /workspace/adamic-tools/env.sh
SCANNER_TYPESCRIPT=/workspace/scratch/native3-cache/api/node_modules/typescript/lib/typescript.js node stage3/drivers/scanner/escape-proof.cjs /workspace/scratch/scanner-escape-proof-verified /workspace/scratch/scanner-any-next-adamic /workspace/scanner-native3-next > /tmp/scanner-escape-proof-verified.log 2>&1
STAGE3_CACHE=/workspace/scratch/native3-cache bash stage3/drivers/scanner/run.sh /workspace/scratch/scanner-escape-stream --tree /workspace/scratch/native3-slice --inputs /workspace/scratch/native3-full/adapted --oracle /workspace/scratch/native3-reference-final/node.stdout --node-only > /tmp/scanner-escape-stream.log 2>&1
# In /workspace/scanner-native3-next, using only scratch discovery sources:
/workspace/scratch/scanner-any-next-adamic build /workspace/scratch/scanner-any-sites/main.a -o /workspace/scratch/scanner-any-sites/scanner-escape > /tmp/scanner-escape-checkpoint-ready.stdout 2> /tmp/scanner-escape-checkpoint-ready.stderr
node --check stage3/drivers/scanner/escape-proof.cjs
git diff --check
```

The fixture belongs to this standalone driver proof, rather than the
internal/oracle fixture registry; internal/oracle/counts.md is unchanged.
No full suite, full lane or full compiler gate was used as confirmation.
