Built recursive, string-dictionary and nullable reference JSON-plus-recovery checks toward roadmap step 09, task #sq4j911.
Delivery is compiler/checked-any after Part 1 b726ac3d3 and main merge 2ef0d19ef; the delivery SHA accompanies the push report.
The 271-site measurement stays 3 checked, 88 refused and 180 not reached; 48 runtime witnesses, 17 refusal assertions, ownership checks, counts and the requested volume run pass.
Six independent boundary-removal mutants are caught in both backends, and 13 source mutants fail their intended assertions.
MapLike<any> awaits compiler/records-maplike, callable contracts remain refused, and the three unchanged stock config declarations still stop before their checked structural uses.

Contracts are checked at the use of an unvalidated any as that type. Input evaluation happens once. The first traversal validates JSON values plus undefined recovery; the second validates the requested schema. Neither traversal catches failures or converts them to diagnostics. A mismatch is terminal exit 70, with the expression path, required type and actual tag. Successful checks preserve the original object or array identity.

Recursive schemas have numeric back-references in acyclic IR and a finite static native contract table. The JavaScript backend reconstructs the same schema graph from a finite table. Runtime traversal bounds data depth at 64; attempting depth 65 reports non-JSON recursion. Actual self-referential runtime objects are separately tested in both backends and stop at raw followed by 65 .next segments, with the pinned recursion message and exit 70; the native control runs under ASan/UBSan. The deep witness is a finite 80-node chain, so dropping the depth bound can finish normally and proves that the required boundary, rather than a later crash, catches the mutant.

String index-signature contracts check every own string-keyed value against the value schema, recursively. Native checks use the object's existing tagged slot metadata and skip absent slots. Numeric keys follow Object.keys ordering through the existing runtime helper; string keys retain insertion order. No record representation or cohere implementation was copied. Records are rejected before interpreting their pending storage. json_maplike_pending is explicitly skipped with awaits compiler/records-maplike: any-valued own-key storage; it is not a passing contract witness. A separate sanitized C control pins the record refusal.

Prepared nullable string, object and array contracts retain distinct null and undefined tags, including inferred literal slots. Equality, typeof, optional property reads and coalescing retain that distinction. Existing source typeof/null checks run in their original order. json_nullable_validated prints its source diagnostic for the number input and returns undefined; Node and both backends have identical stdout, empty stderr and exit 0. The existing config_diagnostics and config_validated witnesses also pass unchanged. No inserted check fires on those validated returns, and no inserted failure becomes a diagnostic.

The witnesses are reduced contract uses, not compilations of the whole stock config entry. Assets are .a files in the existing refusal test directory, copied unchanged to scratch .ts by the source-boundary harness. The fitting witnesses cover recursive objects, numeric dictionaries and identity, nullable string/object fields, nullable arrays and source validation. The six misfitting witnesses discard the typed result and only print passed afterward, so their removal mutants cannot be caught by a later typed read.

| Witness | Required first stop in both backends |
| --- | --- |
| json_recursive_misfit | raw.next.value needs number, found string |
| json_dictionary_misfit | raw.beta needs number, found string |
| json_dictionary_order_misfit | raw.1 needs number, found string, before raw.2 |
| json_nullable_misfit | raw.option needs string \| null \| undefined, found number |
| json_nullable_array_misfit | raw[0] needs number, found string |
| json_depth_misfit | raw followed by 64 .next segments and .value needs JSON value \| undefined, found non-JSON recursion |

Node prints passed and exits 0 for each misfitting source reduction. Removing precisely one CheckedJSON return check makes both emitted backends match that Node observation. The production check intentionally replaces that unsafe continuation with the pinned exit 70. Fitting source observations match Node byte for byte. Native runs use ASan and UBSan; successful runs additionally pass LeakSanitizer. Terminal failures are pinned before cleanup, rather than treated as successful leak-check runs.

Independent source mutants, patches and logs are retained under mutants/next-contracts and evidence/next-contracts:

| Mutant | Catcher and observation |
| --- | --- |
| runtime_depth | json_depth_misfit: native exits 0 and prints passed instead of the pinned depth failure |
| javascript_depth | json_depth_misfit: JavaScript exits 0 and prints passed instead of the pinned depth failure |
| runtime_key_order | json_dictionary_order_misfit: native reports raw.2 instead of raw.1 |
| runtime_record_guard | TestCheckedJSONRecordRemainsUnsupported: the unsupported record returns normally instead of stopping before storage reads |
| nullable_tag | json_nullable: native collapses null into undefined and disagrees with Node's equality and typeof output |
| optional_read | json_nullable: native reads a nullish receiver and fails a fitting source instead of short-circuiting |
| ownership_identity | TestCheckedJSONPreservesIdentityAndOperandWrites: validation is falsely fresh and hides a self-reaching value or operand write |
| nul_key | json_dictionary_nul_key becomes accepted |
| computed_key | json_dictionary_computed_key becomes accepted; the mutant bypasses the branch without asking the AST for an unsupported Text() |
| symbol_field | json_symbol_contract becomes accepted |
| alias_write | json_dictionary_alias_write loses the checked-view refusal and reaches a later union-field refusal |
| proto_key | json_dictionary_proto_key loses the checked-dictionary refusal and reaches the existing prototype-setter refusal |
| callable_contract | json_callable_contract loses its specific verified-parameter/result refusal and reaches the general unreifiable-schema refusal |

The last three mutants pin the precise refusal reason; they do not show that the entire unsafe program becomes accepted. Every listed mutant failed an assertion with a buildable compiler. An earlier computed-key mutant exposed an AST Text panic; it was replaced and rerun with the assertion-only mutant recorded above. A source-mutant attempt also exhausted the disk before testing; that infrastructure failure was not counted as a catcher. Removing 6,479,233,040 bytes of old reproducible Go cache artifacts allowed all source mutants to run. Production sources were restored before final checks.

The minimal ownership hook is the three-line ir.CheckedJSON case in internal/fresh/fresh.go, analysis.value. It evaluates the operand's effects and preserves its identity; validation never invents a source write or freshness. The targeted ownership test proves both self-reaching writes and writes in an Effects operand remain visible. The compiler IR stays acyclic so existing reflection-based analyses terminate. Other hooks are DynamicProperty.Optional in IR and both backend dispatches, dictionary metadata selection in native.dynamicProperties, and nullable-reference coalescing in native.coalesce. No protected lower.go, native/emit.go, native/native.go or oracle/oracle_test.go was edited; no splitter files were edited.

The unchanged stock pin is 050880ce59e30b356b686bd3144efe24f875ebc8, inventory ea1b2359. The final extraction verifier passes all 271 exact declarations and their declaration-only context. The complete after table is evidence/coverage-contracts-after.json. Source hashes and declaration hashes are retained. The Part 1 table is preserved separately rather than rewritten as a runtime-support claim.

| Named stock declaration | First stop on final implementation | Classification |
| --- | --- | --- |
| convertToObject, stock 2467 | 2468:38: computed key without an own data field | refused inside its declaration |
| convertJsonOption, stock 3803 | 3816:37: indirect call of a checked predicate overload | refused inside its declaration |
| normalizeNonListOptionValue, stock 3835 | extracted source 3850:65: ambient predicate has no body proving its parameter | not reached; named stop outside its declaration |

The supplied 4032 normalizer anchor is isExcludedFile in this pin. The probe selects normalizeNonListOptionValue by its original name and unchanged span at 3835, without moving stock text. Both records for that declaration retain the outside-predicate reason. None of these complete declarations now lowers, and no newly revealed bytes are claimed. The remaining outside-stop counts and examples are in COVERAGE_REPORT.md: 88 outside consumers, 42 loader diagnostics, 19 type-only consumers, 14 missing stock-file callers, 9 lowering stops outside their selected declarations and 8 missing generic instantiations.

Focused commands, all output written directly to logs:

    source /workspace/adamic-tools/env.sh
    export GOPROXY='https://proxy.golang.org|direct'
    node stage3/checked-any/validate-extraction.cjs /tmp/checked-any-stock /tmp/checked-any-next-sites-v8/manifest.json
    go run ./stage3/checked-any/probe /tmp/checked-any-next-sites-v8/manifest.json /tmp/checked-any-next-contract-coverage-final.json
    go test ./internal/oracle -run '^(TestCheckedAny|TestCheckedJSON|TestCheckedAnyUnsupportedContracts|TestCheckedJSONNext|TestCheckedJSONNextRefusals)$' -count=1 -timeout 30m -v
    go test ./internal/fresh -run '^TestCheckedJSONPreservesIdentityAndOperandWrites$' -count=1 -v
    go test ./internal/native -run '^(TestCheckedJSONRecordRemainsUnsupported|TestCheckedJSONCycleStopsAtDepth)$' -count=1 -v
    go test ./internal/javascript -run '^TestCheckedJSONCycleStopsAtDepth$' -count=1 -v
    go test ./internal/lower -run '^(TestUnknownReflectionRefusals|TestInheritanceCycleFinderIncludesInheritedFields|TestLiteralMethodCapturesCannotMakeCycles|TestNestedFunctionCycleIsRefused)$' -count=1 -v
    go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 30m -args -update-counts
    go test ./stage1/cohere/typeaware -run '^TestVolumeAgreementAndMutants$' -count=1 -v -timeout 30m

Final oracle: PASS 32.375s, covering 16 prior scalar and 21 prior structural witnesses, 11 new runtime witnesses, 17 refusal assertions and the explicit pending skip. Ownership: PASS 0.011s. Pending record and cyclic-data controls: PASS 0.305s natively and 0.030s in JavaScript. Lower regressions: PASS 0.356s. Counts: PASS 70.512s. Eleven new runtime rows are added; all 995 existing fixture rows keep identical measurements. Each added row belongs to the corresponding witness listed in evidence/next-contracts/count-comparison.json. Fitting runs allocate and free equally; exit 70 rows record the intended early stop. Pending and negative compile-time witnesses add no runtime row. No whole package or full gate ran.

Current origin/main remains 031a1259b and is already merged, including quadratic-emission fix 547551cb8. The requested TestVolumeAgreementAndMutants ran once and passes in 704.513s test time, 705.516s wall time. Its normal and sanitized controls have 26,054 identical finding bytes and 73 findings. All 15 independent compiler-question mutants produce normally terminating outputs caught by the byte oracle: assignable-types, widened-shape, enum-types, type-symbol, scope-locals, call-returns, property-shape, contextual-shape, symbol-origin, type-origin, property-info, call-count, call-parameters, apparent-shape and base-shapes. The released-registry mutant exits 0 and is caught by the required panic 70. The complete output is evidence/next-contracts/volume.log.txt. This run uses its controls and mutants without the optional repository or upstream corpus manifests.

Setup was already completed for Part 1: node 0.020s, Go 0.025s, dependencies 0.069s, submodules 0.076s, clang 0.169s, build 26.693s, deferred tests 26.943s, warm cache 26.944s and done 26.969s; nproc=5, CPU quota=4. The first setup failed on merge-conflict syntax before resolution, as recorded in the Part 1 evidence.

Remaining programs refuse dynamic calls without a verified contract, callable schemas, unsafe alias writes, reflection and unsupported array operations, computed/NUL dictionary keys, prototype setters, symbol-field contracts and unreifiable host/class/collection views. MapLike<any> remains pending. Number-only and boolean-only nullable contracts outside the existing mixed recovery representation remain NotYet. JSON.parse and typed objects containing any fields are not implemented here. The original three stock config bodies remain blocked as measured above; the 271-site inventory is not claimed complete.
