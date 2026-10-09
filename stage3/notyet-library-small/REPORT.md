4b5ecb35: Built regex callbacks, optional string bounds, string lastIndexOf positions, and const-asserted Object.entries; 11 selected roots now clear.
Commits: 7865fefb, 98f0e6a6, 6e91e9d1, 4b5ecb35; views reverted in 34f68c3b, c41 merge repaired in 925fb70d and b094c331.
Validation: four Node fixtures in both backends with sanitizers PASS 12.975s; focused lower PASS 1.435s; latest counts refresh PASS 39.111s.
Mutants: 12 IR mutants caught by Node stdout comparison; three source mutants caught by stdout or the capture-refusal assertion.
Incomplete: remaining kinds need runtime support or design rulings; no passing fixture, mutant or eliminated root is claimed for those kinds.

The worker branch is `codex/notyet-library-small`. Compiler base resolved to
`b410340dc8f889b5799c3bc519117c63def3aa24`; replay base is
`9a1f14c5d994aa855625e7cfa295677060348fec`. Table base is
`57b9777c8eb4ee28b1f50220e8c8fb51a2dfadf7`.

The user correction supersedes the earlier prerequisite request. The pushed views
merge e0fb1a53 was reverted, without rewriting history, in 34f68c3b. Its follow-up
compatibility changes were reverted in 3f675803. c41c0e06 remains as requested.
Conflict repairs restored its explain option and removed a duplicate unconditional
refusal check. `.a` continues to refuse every non-null assertion; checked assertions
are available only for `.ts`. All earlier views-dependent replay certifications are
withdrawn. No further unlanded worker branch will be merged.

Every one of the 35 assigned root sites was replayed again on the corrected compiler
baseline. The findings are in `corrected-replay-findings.json`; scanner 4075 was
replayed once more after the const assertion fix. A missing requested signature is
not by itself proof that the selected unit lowers: only empty findings are counted
clear. These are selected census units measured on the checker-rejected upstream
program, not certification that the entire compiler builds.

The following table is ranked in the requested order. Counts are assigned cluster
counts, and the evidence column separates fixture coverage from cleared roots.

| SHA first | Kind | Assigned roots | Status and evidence |
| --- | --- | ---: | --- |
| 7865fefb, 6e91e9d1 | Regex replacement other than a string | 8 | Lowered full-match string callbacks, plus offset/input callbacks for proven capture-free literal, immutable-alias and conditional producers. Seven roots clear: core 1877, sys 1734, utilities 6211/6221/6252/9598/10916. checker 8943 stops earlier at structural method calls in a program with statics. Captures and unknown producers remain conservatively refused. |
| 98f0e6a6 | lastIndexOf with these arguments | 4 | Lowered string positions with NaN, infinities, truncation, clamping, overlaps and UTF-16 semantics. All four table roots are String calls; no array lowering change is claimed. core 2433 clears. moduleNameResolver 1262 stops on Path; parser 8922 on a closure; utilities 9930 on a Path overload result. |
| No lowering commit | RegExp with a nonconstant pattern | 3 | Skipped: all three signatures still reproduce. Native regex descriptors are compiled by Go before execution. Arbitrary runtime patterns require a runtime ECMAScript parser/compiler, including flag and syntax-error behavior. Dispatch among constant patterns would not cover these sites. Dependency: runtime regex compiler. |
| No lowering commit | JSON.stringify object references | 3 | Skipped: emitter 1138 and watchPublic 686/687 reproduce. Structural views can hide differently represented fields and toJSON. Runtime objects lack complete value metadata; loosening the guard would miscompile. Dependency: complete allocation/value metadata and conversion dispatch, landed through the compiler area. The rejected views integration is not used. |
| 98f0e6a6 | number or undefined slice argument | 3 | Lowered undefined start as zero and undefined end as positive infinity, preserving explicit NaN. Fixture passes. Selected roots remain blocked before the call: emitter 4979 by an unchecked cast, program 751 by number/string binary arithmetic, program 2981 by structural method calls. No cleared roots claimed. |
| 4b5ecb35 | Object.entries unproven shape | 3 | Partly lowered: scanner 4075 is a complete literal under `as const` and now clears. Numeric-key order, Unicode names, values, evaluation order and Map construction match Node. scanner 222/224 still use MapLike index signatures, including a spread; retained for a ruling because `.a` explicitly refuses index signatures and complete runtime values are unproven. No new refusal fixture is claimed for those two roots. |
| 98f0e6a6 | number or undefined substring argument | 2 | Lowered optional end with the undefined/NaN distinction. Both builder 1617 and sourcemap 380 now clear. |
| No lowering commit | Non-intrinsic tagged template | 2 | Skipped: both emitHelpers signatures reproduce. helperString needs per-site cached frozen cooked/raw arrays, rest substitutions and TemplateStringsArray indexing. elementAccess belongs to codex/notyet-element-access; this baseline treats TemplateStringsArray as Object. Dependency: array representation/indexing support landed through area/compiler. No partial tag implementation is claimed. |
| No lowering commit | Set of ResolvedConfigFilePath | 2 | Retained for a ruling: required phantom object fields on a primitive intersection cannot be erased as true `.a` promises. representation rejects primitive constituents; changing only setElement would disagree with construction, arguments and iteration. The 666 signature still reproduces; the 693 replay also encounters the earlier 666 stop. No census echo is cancelled. |
| No lowering commit | Numeric coercion requiring dynamic ToPrimitive | 2 | Skipped: both sys replays stop first on an any-typed Stats parameter at 1747. The operands are Stats.mtime Date objects, but this baseline has no represented Date/Stats runtime path. General coercion also needs Symbol.toPrimitive and callable dispatch. Dependencies: typed host Stats/Date support and dynamic conversion metadata. |
| No lowering commit | JSON union containing containers | 1 | Skipped: commandLineParser 2960 still reproduces the exact signature, after an earlier CompilerOptionsValue representation stop. Arrays carry a reference bit, not their runtime numeric/boolean/nested schema. Dependency: runtime container element metadata. |
| No lowering commit | fs.writeSync wrapping JSON object references | 1 | Skipped: tracing 194 reaches earlier performance/property-call stops. Shares the object JSON metadata dependency; not cancelled as an echo. |
| No lowering commit | Object.assign unproven shape | 1 | Retained for a ruling: utilities 8578 reproduces on a parameter/global allocator structural view. Hidden overwriting fields and function fields are unproven. Existing shape, scalar and cycle restrictions are retained. No unchecked shape assumption or new refusal fixture is claimed. |

Four completed fixtures, each registered from its own `_test.go` file:

- `internal/oracle/testdata/notyet_library_regex_callback.a`
- `internal/oracle/testdata/notyet_library_regex_offset.a`
- `internal/oracle/testdata/notyet_library_string_bounds.a`
- `internal/oracle/testdata/notyet_library_object_entries_const.a`

Owned lowering changes are confined to regexBuiltin, libraryStringMethod and
objectCallArguments. Minimal shared hooks are stringCall in
`internal/lower/object.go` and refuseWidening in `internal/lower/invariance.go`.
The latter bypasses lib.d.ts's any[] rest view only for an intrinsic replacement
argument whose receiver, producer, exact primitive callback parameters and string
return are proven. The new proof helper is
`internal/lower/library_regex_callback_shape.go`. These outside-function files are
named in their commits. Merge repairs also touched `cmd/adamic/tsgo.go` and
`internal/lower/refusals.go`, as named in their commits. No IR, JavaScript, native
emitter or runtime C changes were needed for these features. Runtime C helpers to
review: none.

Every implemented rule has a mutant that was run and caught:

| Mutant | Intended failure |
| --- | --- |
| Regex full-match argument | Node stdout differs |
| Regex replacement-string expansion applied to callback result | Node stdout differs |
| Regex empty-match Unicode advancement | Node stdout differs |
| Regex global collection disabled | Node stdout differs |
| Regex match splice offset replaced with zero | Node stdout differs |
| Regex global lastIndex reset changed | Node stdout differs |
| Offset callback argument replaced with zero | Node stdout differs |
| Callback input argument replaced with an empty string | Node stdout differs |
| slice undefined end replaced with zero | Node stdout differs |
| substring undefined end replaced with zero | Node stdout differs |
| lastIndexOf bounded search prefix ignored | Node stdout differs |
| Const-asserted object field order swapped | Node stdout differs |
| Callback invoked during match collection (source mutant) | Node stdout differs in the fixture oracle |
| Every regex replacement stops after one match (source mutant) | Node stdout differs in the fixture oracle |
| Capturing-group proof removed (Go overlay mutant) | Capture-refusal assertion reports incorrectly accepted source |

The 12 IR mutants execute native output with sanitizers and leak checks; both source
behavior mutants run the fixture oracle in both backends. The proof mutant compiles
Go successfully and is caught by the intended refusal assertion. No compile error
is counted as a kill. Source mutations were restored; the proof mutant uses an
overlay without modifying the tracked source.

c41 prerequisite checks also pass: focused lower 0.089s, CLI 5.122s, its oracle
checks 1.512s. Its no-check mutant was caught by the expected panic/output assertion;
its look-through mutant was caught by the return representation assertion. These
are inherited prerequisite checks, separate from the 15 library-small mutants.

Exact commands, with test output redirected to logs in `corrected-logs/`:

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/notyet_library_(regex_callback|regex_offset|string_bounds|object_entries_const)|TestNotYetLibrary(RegexCallback|RegexOffset|StringBounds|ObjectEntriesConst)Mutant' -count=1 -timeout 10m -v
go test ./internal/lower -run 'TestLibraryRegexOffset|TestRegExpNativeRefusals|TestLibraryStringRefusals|TestLibraryLanguageBoundaries|TestObjectRefusalsExplainSoundness|TestObjectUnprovenShapesStayNotYet|TestNonNullAssertion' -count=1 -v
python3 internal/oracle/testdata/run-library-small-regex-mutants.py
python3 internal/oracle/testdata/run-library-small-regex-proof-mutant.py
go test ./internal/oracle -run TestCountsAreRecorded -args -update-counts
go test ./internal/lower -run 'TestNonNull|TestRefus'
go test ./cmd/adamic -run 'TestNonNull|TestExplainChecks'
go test ./internal/oracle -run 'TestCheckedNonNull(TypeScript|AdamicRefusal|Counts)$|TestPossibleNonNullAdamicAssertionsAreRefused$' -v
/tmp/library-small-replay -project /tmp/library-small-adapted/src/tsc/tsc.ts -where /tmp/library-small-adapted/src/compiler/scanner.ts:4075:59 -kind NotYet -reason 'Object.entries on a shape not proven by a plain literal or its const binding'
```

All 35 roots used that replay command shape with their exact recorded reasons.
Replay exits 1 when the requested signature disappears; parsed findings determine
the reported result. Mandatory counts updates passed after both new fixtures
(61.346s for regex offset, 39.111s for const entries). No whole package or full gate
was run. The latest standing rule governs: the final unit evidence is committed
before one push; any shared gate red will be fixed and followed by one further push.

Setup was completed earlier with GOPROXY=https://proxy.golang.org|direct before
cloud/setup.sh. The first build overlapped the merge and saw missing declarations;
retry after the merge passed. Timing lines: go 0.037s; node 0.046s; submodules
0.169s; markdown skipped 0.018s / ready 0.179s; clang 0.411s; go build 69.489s;
test binaries deferred 69.852s; build cache 69.876s; done 69.935s.
`nproc=5` (quota 4); environment `/workspace/adamic-tools/env.sh`;
Go 1.27.1, clang 20.1.8, Node 24.19.0.
