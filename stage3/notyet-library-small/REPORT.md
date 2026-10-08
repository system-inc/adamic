Built: full-match regex replacement callbacks, optional string bounds, and string lastIndexOf positions.
Commits: 7865fefb (regex), 98f0e6a6 (strings); current main 749a69ad merged in baaa2587.
Validation: Node backend oracles PASS 12.608s; focused lower checks PASS 0.845s; counts PASS 35.606s.
Mutants: six regex IR rules, two regex source rules, and three string IR rules caught by stdout comparison.
Incomplete: nine kinds remain; no runtime C changes, and no completed fixture or mutant is claimed for those kinds.

The worker branch is `codex/notyet-library-small`. Compiler base resolved to
`b410340dc8f889b5799c3bc519117c63def3aa24`; replay base is
`9a1f14c5d994aa855625e7cfa295677060348fec`. The table base is
`57b9777c8eb4ee28b1f50220e8c8fb51a2dfadf7`.

Counts below are the assigned cluster counts, not independently demonstrated eliminated roots.
Only three sample sites were demonstrated clear by replay: core.ts:1877:9,
core.ts:2433:11, and builder.ts:1617:68. Earlier stops prevent other example
replays from establishing elimination. None are cancelled as census echoes.

| Kind | Assigned roots | Status and evidence |
| --- | ---: | --- |
| Regex replacement other than a string | 8 | Lowered zero/one-parameter string callbacks. core.ts:1877:9 clears; checker.ts:8943:45 is blocked by earlier NonNullExpression/missing bindings. Callbacks needing captures, offset or source parameters remain NotYet. |
| lastIndexOf with these arguments | 4 | Lowered string position arguments with NaN, infinities, truncation, clamping, overlaps and UTF-16 semantics. core.ts:2433:11 clears. moduleNameResolver.ts:1262:25 is blocked by its Path parameter. No change to array lastIndexOf is claimed. |
| RegExp with a nonconstant pattern | 3 | Skipped: commandLineParser.ts:4127:56 and parser.ts:10707:31 reproduce. Native regex descriptors are compiled by Go before execution; there is no runtime pattern compiler. A dynamic dispatch among constant patterns would not cover these arbitrary string sites. Requires a native runtime regex compiler, including syntax errors and flags. |
| JSON.stringify object references | 3 | Skipped: watchPublic.ts:686:47 and 687:80 reproduce; emitter is blocked earlier. Runtime shapes carry reference bits, not complete scalar/container types. Removing the guard can serialize hidden booleans as numbers or omit hidden fields/toJSON. Full support needs allocation metadata and dynamic conversion dispatch. Those changes include existing shared runtime records, beyond the separate-new-helper limit. |
| number or undefined slice argument | 3 | Lowered undefined start as zero and undefined end as positive infinity, preserving explicit NaN. Fixture covers both; example replays are blocked earlier and do not prove these roots gone. |
| Object.entries unproven shape | 3 | Refused for a ruling: scanner uses MapLike index-signature objects, including a spread. Adamic explicitly refuses index signatures in refusals.go; removing the exact-shape guard does not prove homogeneous runtime values. No new refusal fixture/mutant was completed. |
| number or undefined substring argument | 2 | Lowered optional end with the same undefined/NaN distinction. builder.ts:1617:68 replay clears; sourcemap is blocked earlier. |
| Non-intrinsic tagged template | 2 | Skipped: both emitHelpers sites reproduce. They call helperString with TemplateStringsArray, rest substitutions and a returned closure. Correct support needs per-site cached frozen cooked/raw arrays and indexing of TemplateStringsArray. elementAccess is owned by codex/notyet-element-access (63f7ae5b); current representation treats this view as Object. No partial tag implementation was committed. |
| Set of ResolvedConfigFilePath | 2 | Refused for a ruling: the brand is an intersection of string with required phantom object fields. representation delegates intersections to objectIntersection, which rejects primitive constituents. Actual upstream values are manufactured with casts. Supporting only setElement would leave construction, arguments and iteration inconsistently represented. Required primitive brand promises need a representation/design ruling; no unsafe erasure was added. |
| Numeric coercion requiring dynamic ToPrimitive | 2 | Skipped: both sys sites use Stats.mtime (Date). Both replays stop first at an any-typed Stats parameter at sys.ts:1747:34. There is no existing Date lowering/runtime representation in this baseline. General ToPrimitive also needs Symbol.toPrimitive and dynamic callable dispatch. No Date shortcut based only on a structural view was added. |
| JSON union containing containers | 1 | Skipped: commandLineParser.ts:2960:31 reproduces. Arrays carry only a reference bit; a boxed array loses its numeric/boolean and nested element schema. Same runtime-metadata dependency as object JSON. |
| fs.writeSync wrapping JSON object references | 1 | Skipped: tracing replay reaches earlier unsupported performance/property calls, not the wrapped signature. Shares the JSON runtime metadata dependency; not cancelled as an echo. |
| Object.assign unproven shape | 1 | Refused for a ruling: utilities.ts:8578:19 reproduces on a parameter/global allocator structural view. Hidden overwriting fields and function fields are not proved by the current shape/scalar checks. Current Object.assign deliberately refuses widened sources and reference cycles. No unchecked shape assumption was added. |

Implementation is confined to regexBuiltin and libraryStringMethod, with one shared
hook in internal/lower/object.go (stringCall) routing slice/lastIndexOf. The follow-up
refusal probe changes internal/lower/regexp_test.go to a two-parameter callback:
one-parameter callbacks now lower. IR, JavaScript, native and runtime files were
not changed. The expanded territory was reviewed; remaining skips above are
implementation/ownership dependencies, not the superseded blanket territory rule.

Fixtures:
- internal/oracle/testdata/notyet_library_regex_callback.a
- internal/oracle/testdata/notyet_library_string_bounds.a

The regex fixture checks literal callback results, global reset, nonglobal and
sticky matching, collection before callbacks, mutation of lastIndex, empty matches
with astral code points, and no matches. String bounds cover omitted/undefined,
NaN, infinities, negative/fractional positions, overlap and argument evaluation order.

Mutants and intended failure:
- regex match: wrong full-match argument -> Node stdout mismatch.
- regex literal: apply replacement-string expansion to callback return -> stdout mismatch.
- regex unicode: advance all empty matches by two UTF-16 units -> stdout mismatch.
- regex global: disable global gathering -> stdout mismatch.
- regex offset: replace match offset with zero -> stdout mismatch.
- regex reset: initialize global lastIndex to one -> stdout mismatch.
- collection-before-callbacks source mutant: invoke during gathering -> stdout mismatch in both backends.
- nonglobal-single-match source mutant: always stop after first match -> stdout mismatch in both backends.
- slice: undefined end becomes zero -> stdout mismatch.
- substring: undefined end becomes zero -> stdout mismatch.
- position: ignore lastIndexOf's bounded search prefix -> stdout mismatch.

IR mutants execute native output with sanitizers and leak checks; source mutants
run the fixture oracle in both backends. Source mutations were restored. The first
final lower run caught an obsolete refusal probe; after updating it, the rerun passes.
An execution-server disconnect interrupted one source mutant; it was restored and
the source mutant runner was rerun successfully. No compile error counts as a kill.

Commands (all test output redirected to log files, no full package/gate runs):
```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/notyet_library_(regex_callback|string_bounds)|TestNotYetLibrary(RegexCallback|StringBounds)Mutants' -count=1 -timeout 10m -v
go test ./internal/lower -run 'TestLibraryStringRefusals|TestRegExpNativeRefusals|TestLibraryLanguageBoundaries' -count=1 -timeout 5m -v
python3 internal/oracle/testdata/run-library-small-regex-mutants.py
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 10m -args -update-counts
go run ./stage3/census/latent/replay -project /tmp/library-small-adapted/src/tsc/tsc.ts -where /tmp/library-small-adapted/src/compiler/core.ts:1877:9 -kind NotYet -reason 'regex replacement other than a string'
```
Replay used the same command shape for all listed examples; exact findings are in
replay-findings.json. A cleared requested signature makes replay exit 1, so the
findings, rather than an assumed successful exit, determine the result.

Setup: exported GOPROXY=https://proxy.golang.org|direct before cloud/setup.sh.
The first setup build overlapped the merge and saw missing lowering declarations;
retry after the merge passed. Retry timing lines: go ready 0.037s; node 0.046s;
submodules 0.169s; markdown skipped 0.018s / ready 0.179s; clang 0.411s;
go build 69.489s; test binaries deferred 69.852s; build cache 69.876s;
done 69.935s. nproc=5 (quota 4). Environment sourced from
/workspace/adamic-tools/env.sh. Go 1.27.1, clang 20.1.8, Node 24.19.0.
