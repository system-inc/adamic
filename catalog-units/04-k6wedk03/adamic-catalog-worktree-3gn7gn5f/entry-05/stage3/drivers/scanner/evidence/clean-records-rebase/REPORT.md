Both native scanner modes remain blocked at corePublic.ts:9:5; no native comparison exists.
Clean records fcddeb29 still conflicts with library in six files and was aborted and skipped.
Fresh scratch af708e4e contains main ffe6efc1 and library b05a9306; compiler build passed in 141.791s.
Both full-tree Node comparisons pass; both exactly-one-byte Node output mutants fail diff.
Fifteen fresh Node witnesses reproduce fifteen ordered stops behind discovery placeholders.

# Clean records rebase attempt

Fetched current origin/main ffe6efc1bd30de2539c1e1416230ca650105a0c7,
origin/area/library b05a9306dd9c49ea1697f9aea4e87fdc3a9d2795 and the requested
records rebase fcddeb299460708f76e8436b4d887a260a6b84f1. The latter is based on
main efe9f404 and includes only records and named-index commits plus their notes.
Created a fresh worktree /workspace/scanner-native3-clean on never-pushed branch
scratch/scanner-native3-clean from current main. Library merged successfully,
producing af708e4eb0b7ecea511f3883192f8f840e6dd831. Then the pinned records merge
conflicted and was aborted and skipped under the unit's conflict-skip rule.
It is clean against main alone, but does not integrate cleanly with library.

The six conflicting files, with no hand resolutions applied, were:

- internal/javascript/javascript.go
- internal/lower/expression.go
- internal/lower/object.go
- internal/lower/refusals.go
- internal/lower/statements.go
- internal/native/emit_expressions.go

The former hand-resolved checkout is excluded entirely from this compiler and
scanner measurement. No prior scratch edits are copied into the new checkout.
The only new scratch checkout filesystem change is the cohere symlink to
/workspace/adamic/cohere after removing the empty worktree directory. No compiler
or adaptation source edit was made. Records/named-index integration is not
claimed; the records feature's own clean-rebase tests were not rerun here.
The full library merge log is compressed losslessly in library-merge.log.gz;
records-merge.log retains all six conflicts. Fresh ref and worktree logs are kept.

## Build and toolchain

Setup was rerun with GOPROXY='https://proxy.golang.org|direct'. All phases passed:
Node 0.019s, Go 0.023s, submodules 0.054s, markdown dependencies 0.067s,
clang 0.150s, Go build 33.443s, deferred test binaries 33.557s,
warm build cache 33.558s, total 33.584s. Those are the setup script's cumulative
timing lines. nproc=5; cgroup quota=4 CPUs. Raw output is setup.log.

The fresh scratch compiler build exited 0: wall 141.791s, user 419.751s,
system 34.658s. Binary SHA256:
4eceea1a0fe51a8e54061f3907537bee43768b3f10916c46c9c36d8899b2bce2.
Bash timing was used; compiler.log retains the line.

## Scanner measurements

Both ADAMIC_NATIVE_SPLIT=0 and split=1 with jobs=5 exited 1 at corePublic.ts:9:5:
Adamic refuses an index signature; use a Map, which keeps keys in insertion order.
The real driver never reaches native source emission or clang. No native scanner
binary, native byte comparison, native-output mutant, scanner timing or binary
size exists for this run.

Both fresh slice Node dumps match the fixed full-tree reference byte for byte,
with empty diffs and exit 0. Same 81-file corpus over all src/compiler:
1,369,432 tokens and 466 errors, 108,019,935 bytes. The skipped-trivia pass
contains 509,014 tokens; retained trivia contains 860,418. SHA256:
5cce1570354cc48b5d9db246daf2abe8db3c21edba3c35e613da812a26920182.
The dump includes cooked values, scanner errors and fixed absolute path headers.
Copied Node controls pass diff, exit 0. Both token-end mutants change exactly
one byte and fail the same diff, exit 1; the evidence audit counts byte differences.
These are Node-output mutants and establish comparison sensitivity while native
is blocked. They are not native scanner evidence. Prior captureStackTrace
controls are historical and not rerun or claimed new here.

## Fifteen successive stops

Each location below is in the progressively modified scratch discovery copy,
not an original upstream coordinate. Every independent witness finishes on Node
(exit 0) and stops this compiler with the identical diagnostic (exit 1).
Raw stops, build timing, Node output and witnesses are retained.

| Order | Scratch location | Message | Node witness |
| --- | --- | --- | --- |
| 1 | corePublic.ts:9:5 | Adamic 0.1 refuses an index signature; use a Map, which keeps keys in the order they were added | [01-index.a](witnesses/01-index.a) |
| 2 | debug.ts:7:1 | Adamic 0.1 refuses a namespace; use a module: a file of its own, with named exports | [02-namespace.a](witnesses/02-namespace.a) |
| 3 | types.ts:689:36 | stage 0 can't lower a reference with both null and undefined (nullable reference needs an empty-case tag) yet | [03-nullable.a](witnesses/03-nullable.a) |
| 4 | diagnosticInformationMap.generated.ts:13:34 | Adamic 0.1 refuses a cast the runtime can't check; use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast) | [04-cast.a](witnesses/04-cast.a) |
| 5 | utilities.ts:67:67 | Adamic 0.1 refuses the comma operator; write each expression as its own statement | [05-comma.a](witnesses/05-comma.a) |
| 6 | scanner.ts:430:21 | Adamic 0.1 refuses the non-null assertion !; write ?? panic('why it can't be missing'), or narrow and handle the missing case | [06-shebang.a](witnesses/06-shebang.a) |
| 7 | scanner.ts:476:24 | Adamic 0.1 refuses the non-null assertion !; write ?? panic('why it can't be missing'), or narrow and handle the missing case | [07-initializer.a](witnesses/07-initializer.a) |
| 8 | scanner.ts:476:12 | Adamic 0.1 refuses the non-null assertion !; write ?? panic('why it can't be missing'), or narrow and handle the missing case | [08-codepoint.a](witnesses/08-codepoint.a) |
| 9 | scanner.ts:500:67 | Adamic 0.1 refuses a cast the runtime can't check; use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast) | [09-string-cast.a](witnesses/09-string-cast.a) |
| 10 | scanner.ts:534:24 | Adamic 0.1 refuses the non-null assertion !; write ?? panic('why it can't be missing'), or narrow and handle the missing case | [10-property.a](witnesses/10-property.a) |
| 11 | core.ts:107:20 | stage 0 can't lower new an Identifier yet | [11-array.a](witnesses/11-array.a) |
| 12 | utilities.ts:13:17 | stage 0 can't lower a function returning U &#124; undefined yet | [12-generic-return.a](witnesses/12-generic-return.a) |
| 13 | utilities.ts:17:56 | stage 0 can't lower a function value taking string &#124; number yet | [13-function-union.a](witnesses/13-function-union.a) |
| 14 | scanner.ts:258:27 | stage 0 can't lower a void call used as a value yet | [14-nullish.a](witnesses/14-nullish.a) |
| 15 | scanner.ts:364:9 | Adamic 0.1 refuses a value as a condition; compare it explicitly, like name.length > 0 or count !== 0 | [15-condition.a](witnesses/15-condition.a) |

Discovery placeholders follow the previous explicit sequence: type-only
MapLike declaration with a dormant throwing marker; throwing Debug members;
unused CompilerOptionsValue alias with a dormant throw; throwing diag,
parsePseudoBigInt, scanShebangTrivia, whole createScanner, codePointAt,
UTF16 worker, Script_Extensions initializer, levenshteinWithMax,
forEachEntry with a never return, getNameOfScriptTarget, and lookupInUnicodeMap.
The fifteenth stop is not bypassed. The normalized source diff is in
 discovery-only.json as evidence, not committed adaptation implementation.
All discovery programs are intentionally incomplete and uncommitted.
Replacing whole createScanner omits its nested bodies from later traversal;
this is a bounded ordered walk, not an exhaustive census. Observation 14 relies
on the Debug object-of-functions placeholder and is not attributed to unchanged
namespace lowering. No native result behind these placeholders is claimed.

The fifteen messages match the previous native3/native3-records ordered list;
none of that measured list is gone. The original requested historical October 8
fifteen-item list was not present in the named documents or refs in the prior
unit, so no retirement from that unidentified list is inferred. ErrorConstructor
control closure remains the previously scoped observation, not a fresh full
scanner result.

## Commands and validation

The compiler command ran from /workspace/scanner-native3-clean, and driver/walk/
witness commands from /workspace/adamic. Toolchain environment was sourced
for each execution. Output went to files, never through test-output pipes.
The unchanged slice and fixed corpus are reused from the prior evidence, with
the existing README feature profile and no new source adaptations.

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/scanner-clean-setup.log 2>&1
source /workspace/adamic-tools/env.sh
TIMEFORMAT='wall=%3R user=%3U sys=%3S'
{ time go build -buildvcs=false -o /workspace/scratch/scanner-clean-adamic ./cmd/adamic; } > /tmp/scanner-clean-compiler.log 2>&1
STAGE3_CACHE=/workspace/scratch/native3-cache ADAMIC_NATIVE_SPLIT=0 bash stage3/drivers/scanner/run.sh /workspace/scratch/native3-clean-unsplit --tree /workspace/scratch/native3-slice --inputs /workspace/scratch/native3-full/adapted --oracle /workspace/scratch/native3-reference-final/node.stdout --compiler /workspace/scratch/scanner-clean-adamic --compiler-cwd /workspace/scanner-native3-clean > /tmp/scanner-clean-unsplit.log 2>&1
STAGE3_CACHE=/workspace/scratch/native3-cache ADAMIC_NATIVE_SPLIT=1 ADAMIC_NATIVE_JOBS=$(nproc) bash stage3/drivers/scanner/run.sh /workspace/scratch/native3-clean-split --tree /workspace/scratch/native3-slice --inputs /workspace/scratch/native3-full/adapted --oracle /workspace/scratch/native3-reference-final/node.stdout --compiler /workspace/scratch/scanner-clean-adamic --compiler-cwd /workspace/scanner-native3-clean > /tmp/scanner-clean-split.log 2>&1
python3 /tmp/scanner-clean-walk.py > /tmp/scanner-clean-walk.log 2>&1
python3 /tmp/scanner-clean-witnesses.py > /tmp/scanner-clean-witnesses.log 2>&1
python3 /tmp/scanner-clean-evidence.py > /tmp/scanner-clean-audit.log 2>&1
```

Evidence audit passes all fifteen message pairs, both empty full-tree diffs,
both control passes and both exactly-one-byte mutant catches. The reproduction
scripts preserve actual workspace paths. No full gate or whole package test was
run, no oracle fixture was added, and counts.md is unchanged. Native ownership,
sanitizers, parser-directed rescans, native performance and records integration
remain uncovered. Only scanner evidence and notes are delivered.
