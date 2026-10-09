Built the scratch compiler at b05a9306; native scanner builds remain blocked.
Main efe9f404 and library b05a9306 integrated; front-3 a36d1c04 conflicted and was skipped.
Both split modes match Node's full-tree reference before refusing corePublic.ts:9:5.
Both Node end mutants fail diff; a native capture-control byte mutant also fails diff.
No native scanner timings or equality claim; the referenced October 8 list is absent.

# October 8 native scanner unit 3

## Integration and setup

The deliverable branch starts from origin/main efe9f4042049234e5a52639fe77b47c311fd530c.
The never-pushed scratch branch is scratch/scanner-native3, in
/workspace/scanner-native3-scratch. Merging origin/main was already up to date;
origin/area/library fast-forwarded to b05a9306dd9c49ea1697f9aea4e87fdc3a9d2795.
Both library fix commits 2bf78f6a and 0ab7c2bd are ancestors.
origin/compiler/stage3-front-3 a36d1c0472649ef2ee549cc2505b93a84e893b59
conflicted and was aborted, without resolving or editing compiler sources.
The entire conflict list is in summary.json and front-merge.log.

Ordinary origin fetch updates only main; both feature refs were explicitly fetched.
Setup ran with GOPROXY='https://proxy.golang.org|direct' and exited 0.
The environment file is /workspace/adamic-tools/env.sh; nproc is 5, quota 4 CPUs.
Setup timings are cumulative elapsed readings: Node 0.248s, Go 0.267s,
clang 0.890s, Markdown dependencies 2.474s, submodules 23.994s,
go build 285.795s, cache warm 285.962s, total 286.021s.
Setup deferred test binaries, as intended. No setup failure occurred.

The first scratch compiler attempt lacked cohere/TypeScript/tsc/go.mod.
The empty scratch submodule directory was replaced with a symlink to the
setup-restored /workspace/adamic/cohere. The following command exited 0:

```sh
source /workspace/adamic-tools/env.sh
go build -buildvcs=false -o /workspace/scratch/scanner-native3-adamic ./cmd/adamic > /tmp/scanner-native3-compiler.log 2>&1
```

No compiler or adaptation implementation was edited. The delivery changes
are only evidence and notes under stage3/drivers/scanner.

## Corpus and comparisons

The checked-in runner currently emits seven token fields, cooked values,
error callbacks and pass headers, over two trivia passes. The full-tree
Node reference and both slice Node streams compare byte for byte (diff exit 0):
81 input files, 1,369,432 tokens, 466 errors, 108,019,935 bytes.
Their exact stream SHA256 is
5cce1570354cc48b5d9db246daf2abe8db3c21edba3c35e613da812a26920182.
This dump's headers include absolute input paths, so its hash is specific to this workspace.
The skipped-trivia projection, omitting only headers/errors and the cooked-value
column, is exactly the historical six-field reference: 509,014 tokens,
27,879,197 bytes, SHA256
c1a9f239790e158cc4471aa6c9273ff678cb32e5890b3d4c95e077ee90c61b0f.
The retained-trivia pass contains 860,418 tokens.

The fixed corpus is the pipeline's generated full compiler tree, with adaptation
10 and without 50's added scanner returns. Gathering uses the initial 10+50 tree,
then applies the existing scanner profile automatically. Source audit passes
168 spans and 78 ordered module import lists. No new adaptation is introduced.
An initial corpus still included 50 and therefore contained nine extra tokens.
Restoring the pinned scanner also required rerunning 10's type-import annotation.
The interim reference without those annotations failed Node; it is not used.
The corpus was stable before the final full-tree and both final slice runs.

The runner's copied Node controls compare equal. Increasing the first token's
end by one changes a single byte and is rejected by diff (exit 1) in both modes.
These are Node-output mutants. Both actual native builds exit 1 at
corePublic.ts:9:5, the type-only MapLike index signature. No scanner binary exists.
No native scanner token comparison, native-output mutant, binary size,
source-emission/clang timing or best-of-three CPU timing can be reported.
Split compilation is not reached, so these results do not measure splitting.

Final commands, from the scratch checkout (every invocation wrote to logs):

```sh
export STAGE3_CACHE=/workspace/scratch/native3-cache
export SLICE_TYPESCRIPT=$STAGE3_CACHE/api/node_modules/typescript/lib/typescript.js
bash stage3/drivers/scanner/run.sh /workspace/scratch/native3-full --node-only > /tmp/scanner-native3-full.log 2>&1
bash stage3/slice/run.sh /workspace/scratch/native3-full/adapted /workspace/scratch/native3-slice src/compiler/scanner.ts:createScanner src/compiler/types.ts:ScriptTarget src/compiler/types.ts:SyntaxKind > /tmp/scanner-native3-slice.log 2>&1
# Fixed corpus: restore only scanner.ts from pinned HEAD, then reapply existing 10.
NODE_PATH=$STAGE3_CACHE/api/node_modules node stage3/adapt/10-type-imports/adapt.cjs /workspace/scratch/native3-full/adapted > /tmp/scanner-native3-corpus-imports.log 2>&1
bash stage3/drivers/scanner/run.sh /workspace/scratch/native3-reference-final --tree /workspace/scratch/native3-full/adapted --node-only > /tmp/scanner-native3-reference-final.log 2>&1
ADAMIC_NATIVE_SPLIT=0 bash stage3/drivers/scanner/run.sh /workspace/scratch/native3-unsplit-proof --tree /workspace/scratch/native3-slice --inputs /workspace/scratch/native3-full/adapted --oracle /workspace/scratch/native3-reference-final/node.stdout --compiler /workspace/scratch/scanner-native3-adamic --compiler-cwd /workspace/scanner-native3-scratch > /tmp/scanner-native3-unsplit-proof.log 2>&1
ADAMIC_NATIVE_SPLIT=1 ADAMIC_NATIVE_JOBS=$(nproc) bash stage3/drivers/scanner/run.sh /workspace/scratch/native3-split-proof --tree /workspace/scratch/native3-slice --inputs /workspace/scratch/native3-full/adapted --oracle /workspace/scratch/native3-reference-final/node.stdout --compiler /workspace/scratch/scanner-native3-adamic --compiler-cwd /workspace/scanner-native3-scratch > /tmp/scanner-native3-split-proof.log 2>&1
```

## Ordered observations behind placeholders

All coordinates below are in the successive discovery copies, not original
upstream coordinates. Removing a whole function changes subsequent line numbers.
After the first stop, the copy is deliberately incomplete and has no semantic
scanner proof. Exact diagnostics and standalone source/Node/build results are
in stops.json and witnesses/. All 15 witnesses exit 0 on Node and exit 1
at the same diagnostic message on this compiler.

| Order | Scratch file:line:column | Exact message | Small Node witness |
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

The bypasses, in order, are:

1. Empty the refused type-only index signature, retaining a dormant throwing
   marker function. A type declaration has no executable body to replace.
2. Replace Debug's namespace with an object holding throwing fail/assertEqual
   functions and the same isDebugging flag. This changes its representation.
3. Replace the unused CompilerOptionsValue type alias with string and a dormant
   throwing marker. This is a discovery type placeholder, not an adaptation.
4. Make diag a throwing function returning DiagnosticMessage. This prevents
   the original optional-field cast from blocking the following declarations.
5. Replace parsePseudoBigInt's body with a throw.
6. Replace scanShebangTrivia's body with a throw, preserving its inferred number return.
7. Replace the whole createScanner body with a throw. Its nested bodies are
   no longer traversed, so this run is not an exhaustive census of those bodies.
8. Replace codePointAt's body with a throw.
9. Replace the UTF16 worker initializer with a typed throwing arrow.
10. Replace Script_Extensions' initializer with a typed throwing expression.
11. Replace levenshteinWithMax's body with a throw.
12. Replace forEachEntry with a throw and a never return annotation, since
    preserving the generic return signature repeats the same signature stop.
13. Replace getNameOfScriptTarget's body with a throw to remove its callback.
14. Replace lookupInUnicodeMap's body with a throw.
15. Stop at scanConflictMarkerTrivia's optional callback condition; no bypass
    after the fifteenth observation was attempted.

Observation 14 depends on Debug's object-of-functions placeholder. Its standalone
witness reproduces the exact representation failure, but this run does not prove
that the original namespace would stop there after a namespace fix. All other
observations after 1 are likewise scoped to the accumulated discovery copy.
Repeated signature stops and errors caused by malformed placeholders are excluded.
The initial eager type marker introduced an enum-initialization stop; the marker
was made dormant. An incorrectly parsed diag signature caused syntax errors;
it was corrected. An inferred-void throwing function caused TS2322; its original
number return annotation was retained. Neither these nor missing-compiler startup
attempts count as scanner feature stops. discovery-only.json retains the final
uncommitted replacements, with CRLF normalized for readability; no adapted
source from it is proposed for use.

## Library fix and the missing historical list

The October 8 fifteen-item list referenced by the dispatch is absent from the
complete BLOCKERS.md at all three fetched tips. No retirement mapping for an
unavailable list is claimed. The first actual stop in this integration is the
older MapLike refusal, because the conflicting front-3 merge was skipped.

The existing capture-stack-marker-control.a separately proves typed static
captureStackTrace admission: compiler exit 0, Node and native both print ok,
both exit 0, byte comparison exit 0. Changing only the first byte of its native
output (ok to nk) is caught by diff exit 1. This is a native control mutant,
not a scanner mutant. The old missing static method does not occur in this
control. The marker-bearing capture-stack-marker.a instead fails checking at
4:42 with TS2769: {} is not Function. Neither control establishes Debug.fail's
full stack behavior in the scanner. Raw results are capture-control-* and
capture-marker-build.stderr.

## Validation and limits

Ran setup, scratch compiler build, slice audit, full-tree Node reference,
both scanner modes, all 15 independent Node/build witnesses, typed capture
control, and the three output-mutant comparisons described above. Each build
and Node invocation wrote stdout/stderr to files. The evidence audit verifies
all witness messages exactly and the historical six-field token hash.
No whole package or full gate was run. No oracle fixture was registered and
counts.md was not changed. Parser-directed rescans, scanner native ownership,
sanitizers, native token equality and scanner performance remain unmeasured.
