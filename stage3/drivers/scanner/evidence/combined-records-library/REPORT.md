Both native scanner modes now stop at debug.ts:7:1; no native scanner comparison exists.
Combined branch 2648ad33 merged cleanly with main 6998ebc2 into fresh scratch dbd7a7c8.
Compiler build passed in 186.725s; unchanged MapLike probe prints ok on Node and native.
Both full-tree Node diffs pass; both one-byte Node mutants and the MapLike native control mutant are caught.
Fifteen fresh Node witnesses reproduce ordered stops; two depend on the throwing Debug placeholder.

# Combined records and library measurement

Started a fresh worktree /workspace/scanner-native3-combined on never-pushed branch
scratch/scanner-native3-combined from current origin/main
6998ebc24ae353193cb1495d3d51308131a4b5c7. Merged the requested combined branch
2648ad3362812b0f184a8cb922bc322d18729c66, which includes library base
e40216f23caa519ad90e011dfc3a47d5c3bfeb6e and the records/named-index commits.
The merge succeeded without conflicts or hand resolutions, producing
 dbd7a7c84952b31aa1bd3d723b58d71c4dc1430f. Ancestry checks confirm main,
combined and library base are ancestors of this scratch head.

The former hand-resolved checkout and all other prior compiler scratch edits
are excluded. The only new compiler checkout filesystem replacement is the
cohere symlink to /workspace/adamic/cohere after removing the empty worktree
directory. Its commit matches the pinned submodule
7945d102a6c18dd36adf9114a758ce646e8b2359. No compiler or adaptation source was
edited. The source status is clean with submodule inspection excluded.
Fresh fetch/worktree logs and the losslessly compressed merge.log.gz are kept.

## Real scanner result and byte comparison

Both ADAMIC_NATIVE_SPLIT=0 and split=1 with jobs=5 exited 1 at debug.ts:7:1:
Adamic refuses a namespace; use a module, a file of its own with named exports.
The prior corePublic.ts:9:5 MapLike index-signature stop is passed in the real
unchanged slice. The compiler still stops before native source emission or clang.
There is no native scanner binary, native token comparison, scanner timing,
binary size or native-scanner-output mutant to report.

Both fresh slice Node dumps match the full-tree reference byte for byte, with
empty diffs and exit 0, over the same complete 81-file src/compiler corpus.
They contain 1,369,432 tokens and 466 errors, totaling 108,019,935 bytes:
509,014 skipped-trivia tokens and 860,418 retained-trivia tokens. SHA256:
5cce1570354cc48b5d9db246daf2abe8db3c21edba3c35e613da812a26920182.
The reference includes cooked values, error rows and fixed absolute path headers.
Copied Node controls pass diff, exit 0. Each token-end mutant changes exactly one
byte and fails the same diff, exit 1. The audit counts byte differences directly.
These two mutants are Node output, not native scanner output.

## Closed MapLike control and retirement status

The unchanged previous 01-index.a witness now builds successfully. It retains
export interface MapLike<T> { [index: string]: T; } and prints ok. Source Node
and the native binary both exit 0, print exactly ok followed by newline, and
have empty stderr. Their stdout diff exits 0. Changing only the native output's
first byte from o to n causes the same diff to exit 1. Witness source, both
outputs, control, mutant and compiler logs are in witnesses/01-index.*.
This is type-only declaration admission proof, not record runtime or scanner proof.

From the preceding measured fifteen-stop list (evidence/native3 and subsequent
replays), order 1, corePublic.ts:9:5 index signature, is gone. Previous orders
2 through 15 remain and are fresh orders 1 through 14 below. The new fifteenth
observation is an optional-parameter function value in the throwing Debug object
placeholder. Previous ErrorConstructor controls remain historical and are not
rerun here. The originally referenced separate October 8 fifteen-item list was
not identified in prior work; no retirement from that unidentified list is inferred.

## Fifteen ordered stops and Node witnesses

Every listed witness finishes on Node (exit 0) and stops this compiler with the
same diagnostic (exit 1). Coordinates refer to the progressively modified
scratch discovery copy, not original upstream coordinates. Files stop-02.stderr
through stop-16.stderr use the previous traversal's numbering; stops.json records
fresh order and that previous-order identifier. Fresh order 15 is beyond the
previous list and uses the identifier 16.

| Order | Scratch location | Message | Node witness |
| --- | --- | --- | --- |
| 1 | debug.ts:7:1 | Adamic 0.1 refuses a namespace; use a module: a file of its own, with named exports | [02-namespace.a](witnesses/02-namespace.a) |
| 2 | types.ts:689:36 | stage 0 can't lower a reference with both null and undefined (nullable reference needs an empty-case tag) yet | [03-nullable.a](witnesses/03-nullable.a) |
| 3 | diagnosticInformationMap.generated.ts:13:34 | Adamic 0.1 refuses a cast the runtime can't check; use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast) | [04-cast.a](witnesses/04-cast.a) |
| 4 | utilities.ts:67:67 | Adamic 0.1 refuses the comma operator; write each expression as its own statement | [05-comma.a](witnesses/05-comma.a) |
| 5 | scanner.ts:430:21 | Adamic 0.1 refuses the non-null assertion !; write ?? panic('why it can't be missing'), or narrow and handle the missing case | [06-shebang.a](witnesses/06-shebang.a) |
| 6 | scanner.ts:476:24 | Adamic 0.1 refuses the non-null assertion !; write ?? panic('why it can't be missing'), or narrow and handle the missing case | [07-initializer.a](witnesses/07-initializer.a) |
| 7 | scanner.ts:476:12 | Adamic 0.1 refuses the non-null assertion !; write ?? panic('why it can't be missing'), or narrow and handle the missing case | [08-codepoint.a](witnesses/08-codepoint.a) |
| 8 | scanner.ts:500:67 | Adamic 0.1 refuses a cast the runtime can't check; use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast) | [09-string-cast.a](witnesses/09-string-cast.a) |
| 9 | scanner.ts:534:24 | Adamic 0.1 refuses the non-null assertion !; write ?? panic('why it can't be missing'), or narrow and handle the missing case | [10-property.a](witnesses/10-property.a) |
| 10 | core.ts:107:20 | stage 0 can't lower new an Identifier yet | [11-array.a](witnesses/11-array.a) |
| 11 | utilities.ts:13:17 | stage 0 can't lower a function returning U &#124; undefined yet | [12-generic-return.a](witnesses/12-generic-return.a) |
| 12 | utilities.ts:17:56 | stage 0 can't lower a function value taking string &#124; number yet | [13-function-union.a](witnesses/13-function-union.a) |
| 13 | scanner.ts:258:27 | stage 0 can't lower a void call used as a value yet | [14-nullish.a](witnesses/14-nullish.a) |
| 14 | scanner.ts:364:9 | Adamic 0.1 refuses a value as a condition; compare it explicitly, like name.length > 0 or count !== 0 | [15-condition.a](witnesses/15-condition.a) |
| 15 | debug.ts:9:9 | stage 0 can't lower a function value with an optional parameter yet | [16-new.a](witnesses/16-new.a) |

Discovery starts from an untouched copy of the existing validated slice and
retains the real MapLike declaration. Throwing placeholders successively replace
Debug members, the unused CompilerOptionsValue alias (with dormant throw marker),
diag, parsePseudoBigInt, scanShebangTrivia, whole createScanner, codePointAt,
UTF16 worker, Script_Extensions initializer, levenshteinWithMax,
forEachEntry with a never return, getNameOfScriptTarget, lookupInUnicodeMap,
and scanConflictMarkerTrivia. The final observed stop is not bypassed.
The normalized source diff is retained only as evidence in discovery-only.json;
no discovery source or adaptation implementation is delivered.

Replacing whole createScanner omits its nested bodies from later traversal.
This is a bounded discovery walk, not an exhaustive scanner census. Fresh orders
13 and 15 depend on replacing the original Debug namespace with an object of
throwing function values; they are not attributed to unchanged namespace lowering.
The order-15 witness deliberately reproduces that placeholder signature and is
qualified likewise. All observations after order 1 describe the accumulated,
intentionally incomplete copy. No successful scanner behavior behind these
placeholders is claimed.

## Toolchain and build timing

Setup reran with GOPROXY='https://proxy.golang.org|direct' and passed all phases.
Its cumulative timing lines: Go 0.026s, Node 0.027s, submodules 0.059s,
markdown dependencies 0.073s, clang 0.174s, Go build 54.055s,
deferred test binaries 54.190s, warm build cache 54.192s, total 54.226s.
nproc=5, cgroup CPU quota=4. Raw output is setup.log.

Fresh scratch compiler build exited 0: wall 186.725s, user 450.874s,
system 48.701s, recorded with Bash timing. Binary SHA256:
1713ffe69ea6cc9fc19f5a3a747d57ca6c13a916befb55eb47efd568d07fe2fc.
Toolchain setup and scratch build overlapped; these are observed timings, not
an isolated performance comparison. No scanner execution timing is available.

## Commands and validation

Compiler merge/build commands ran in /workspace/scanner-native3-combined;
setup, scanner drivers, discovery and witnesses ran in /workspace/adamic.
Every command's output goes to a file. The unchanged README feature-profile slice
and fixed corpus from the prior evidence were reused; no new adaptation ran.

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/scanner-combined-setup.log 2>&1
source /workspace/adamic-tools/env.sh
TIMEFORMAT='wall=%3R user=%3U sys=%3S'
{ time go build -buildvcs=false -o /workspace/scratch/scanner-combined-adamic ./cmd/adamic; } > /tmp/scanner-combined-compiler.log 2>&1
STAGE3_CACHE=/workspace/scratch/native3-cache ADAMIC_NATIVE_SPLIT=0 bash stage3/drivers/scanner/run.sh /workspace/scratch/native3-combined-unsplit --tree /workspace/scratch/native3-slice --inputs /workspace/scratch/native3-full/adapted --oracle /workspace/scratch/native3-reference-final/node.stdout --compiler /workspace/scratch/scanner-combined-adamic --compiler-cwd /workspace/scanner-native3-combined > /tmp/scanner-combined-unsplit.log 2>&1
STAGE3_CACHE=/workspace/scratch/native3-cache ADAMIC_NATIVE_SPLIT=1 ADAMIC_NATIVE_JOBS=$(nproc) bash stage3/drivers/scanner/run.sh /workspace/scratch/native3-combined-split --tree /workspace/scratch/native3-slice --inputs /workspace/scratch/native3-full/adapted --oracle /workspace/scratch/native3-reference-final/node.stdout --compiler /workspace/scratch/scanner-combined-adamic --compiler-cwd /workspace/scanner-native3-combined > /tmp/scanner-combined-split.log 2>&1
python3 /tmp/scanner-combined-walk.py > /tmp/scanner-combined-walk.log 2>&1
python3 /tmp/scanner-combined-witnesses.py > /tmp/scanner-combined-witnesses.log 2>&1
python3 /tmp/scanner-combined-new-witness.py > /tmp/scanner-combined-new-witness.log 2>&1
python3 /tmp/scanner-combined-maplike.py > /tmp/scanner-combined-maplike.log 2>&1
python3 /tmp/scanner-combined-evidence.py > /tmp/scanner-combined-audit.log 2>&1
```

Evidence audit passes all fifteen Node-success/build-refusal message pairs,
both empty full-tree diffs, both control passes and both one-byte Node mutants,
plus MapLike's native equality and one-byte native-control mutant. Source scripts
retain actual workspace paths. No whole package or full gate was run, no oracle
fixture was added, and counts.md is unchanged. Native scanner ownership,
sanitizers, parser-directed rescans and native scanner performance remain
uncovered. Only scanner evidence and notes are delivered; scratch is never pushed.
