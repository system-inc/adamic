Both native scanner modes remain blocked at corePublic.ts:9:5; no native comparison exists.
Named-index c20e5b56 and records-lowering 456c981b both conflicted and were skipped.
Scratch compiler b05a9306 rebuilt successfully in 1.968s and is byte-identical to the prior compiler.
Both Node streams match the full-tree reference; both end mutants fail diff with exit 1.
Fifteen fresh Node witnesses reproduce fifteen ordered stops behind discovery placeholders.

# Records integration follow-up

The delivery branch remains codex/stage3-scanner-native-3, following commit
5c375d5f. The scratch branch remains scratch/scanner-native3 and is never pushed.
Fetched origin/codex/named-index-records at c20e5b563cc60c21f83e55661a8eb2ef63e3815a
and origin/codex/records-lowering at 456c981b1c108abae5f2c8d6ae665cfd6b92e2fe.
The latter includes requested base 70fb62b1. Both merges conflicted and were
aborted without resolving or editing compiler sources. Named-index has
17 conflicted files; records-lowering has six. Full lists are in summary.json
and the two raw merge logs. Neither requested feature is integrated, and no
record-admission closure is claimed from these attempts.

Rebuilt the compiler after both aborts, exit 0: wall 1.968s, user 2.093s,
system 0.243s. /usr/bin/time was unavailable (exit 127); Bash timing was used
for the completed build. The rebuilt binary SHA256 is
6a12c6735c94756dd659b1595c84c6318c12a1567328143035b157850b772ba0,
identical to the previous compiler. The only scratch checkout filesystem
replacement remains the previously documented cohere symlink to setup's
restored submodule. No compiler or adaptation implementation changed.
Toolchain setup from the same session remains valid: total 286.021s,
nproc 5, CPU quota 4; its raw log is ../native3/setup.log.

## Scanner measurement

Reran the existing unchanged slice on the same fixed 81-file compiler corpus.
Both split=0 and split=1, jobs=5, exit 1 at the same index-signature refusal:
corePublic.ts:9:5. The checker/lowering stop precedes source emission and clang.
No native scanner binary, native token comparison, scanner timings, binary
size or scanner-native output mutant is available.

Both slice Node streams match the full-tree reference byte for byte, with empty
diffs (exit 0). Each emits 1,369,432 tokens, 466 errors and 108,019,935 bytes:
509,014 skipped-trivia tokens and 860,418 retained-trivia tokens. SHA256 is
5cce1570354cc48b5d9db246daf2abe8db3c21edba3c35e613da812a26920182;
headers contain fixed absolute corpus paths. The historical six-field skipped
stream remains the previously recorded 509,014-token reference. Copied Node
controls compare equal (exit 0), and incrementing only the first token end
changes one byte and fails the same comparison (exit 1) in both modes.
These two fresh mutants are Node output, not native output. The prior typed
captureStackTrace control is not rerun or claimed as new evidence here.

## Ordered stops and Node witnesses

Fresh traversal copied the unchanged slice and driver into
/workspace/scratch/native3-records-walk. Every build wrote stdout and stderr
to separate files. The table's locations are successive scratch coordinates,
not original upstream line numbers. Each witness below was rerun on source
Node (exit 0) and this compiler (exit 1), with the exact same diagnostic message.
Full messages, build times, Node output and locations are in stops.json and
witnesses/results.json.

| Order | Scratch location | Diagnostic | Node witness |
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

The traversal repeats the prior explicit discovery sequence: type-only index
signature placeholder, throwing Debug members, unused nullable alias placeholder,
throwing diag, parsePseudoBigInt, scanShebangTrivia, whole createScanner,
codePointAt, UTF16 worker, Script_Extensions initializer, levenshteinWithMax,
forEachEntry with a never return, getNameOfScriptTarget, and lookupInUnicodeMap.
Stop after the fifteenth observation. Type-only declarations use narrowed
placeholder declarations with dormant throwing marker functions; they have no
runtime body to replace. Whole createScanner replacement omits its nested
bodies from later traversal. Observation 14 depends on the Debug object-of-functions
placeholder and is not attributed to unchanged namespace lowering. All observations
after 1 describe the accumulated deliberately incomplete copy, not a native
scanner implementation. The discovery source remains uncommitted; its normalized
diff is stored only as evidence in discovery-only.json.

## Commands and validation

All commands ran in the existing scratch compiler checkout with the setup
environment sourced. Shell logs and each runner's separate stdout/stderr are
retained here. No whole package or full gate was run; no oracle fixture was
registered and counts.md remains unchanged.

```sh
source /workspace/adamic-tools/env.sh
TIMEFORMAT='wall=%3R user=%3U sys=%3S'
{ time go build -buildvcs=false -o /workspace/scratch/scanner-native3-records-adamic ./cmd/adamic; } > /tmp/scanner-native3-records-compiler.log 2>&1
STAGE3_CACHE=/workspace/scratch/native3-cache ADAMIC_NATIVE_SPLIT=0 bash stage3/drivers/scanner/run.sh /workspace/scratch/native3-records-unsplit --tree /workspace/scratch/native3-slice --inputs /workspace/scratch/native3-full/adapted --oracle /workspace/scratch/native3-reference-final/node.stdout --compiler /workspace/scratch/scanner-native3-records-adamic --compiler-cwd /workspace/scanner-native3-scratch > /tmp/scanner-native3-records-unsplit.log 2>&1
STAGE3_CACHE=/workspace/scratch/native3-cache ADAMIC_NATIVE_SPLIT=1 ADAMIC_NATIVE_JOBS=$(nproc) bash stage3/drivers/scanner/run.sh /workspace/scratch/native3-records-split --tree /workspace/scratch/native3-slice --inputs /workspace/scratch/native3-full/adapted --oracle /workspace/scratch/native3-reference-final/node.stdout --compiler /workspace/scratch/scanner-native3-records-adamic --compiler-cwd /workspace/scanner-native3-scratch > /tmp/scanner-native3-records-split.log 2>&1
python3 /tmp/scanner-native3-records-walk.py > /tmp/scanner-native3-records-walk.log 2>&1
python3 /tmp/scanner-native3-records-probes.py > /tmp/scanner-native3-records-probes.log 2>&1
```

Evidence validation checks all fifteen Node-success/build-refusal message pairs,
both empty full-tree diffs, both comparison-control passes and both end-mutant
catches. Source scripts retain the actual workspace paths for reproducing this
run. Native ownership, sanitizers, parser-driven rescans and native scanner
performance remain uncovered. Both requested feature merges remain untested
in this integration because conflicts required skipping them.
