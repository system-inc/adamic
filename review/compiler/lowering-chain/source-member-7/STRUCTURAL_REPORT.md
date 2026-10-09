Built finite structural checks for tagged .ts any at typed JSON-plus-recovery uses toward step 09.
Based on ba36dda2 with area commit f0c6e6fc merged as bf547959; delivery SHA is in the push report.
Focused scalar and structural witnesses, lowering regressions, compatibility witnesses and recorded counts pass.
Boundary-removal IR mutants and independent runtime/lowering mutants are caught; logs are under evidence/.
This is partial: original config declarations still encounter other blockers; independent site counts follow below.

Structural checks

The public declarations keep their source text. checkedAnyType now constructs CheckedJSON IR for finite object/array/literal contracts, evaluates its input once, checks every own child against JSON values plus undefined recovery, then checks required/optional fields and array elements recursively against the requested type. Nonfinite numbers, callable/class/opaque values and excessively deep structures terminate. Failures name the source path, needed type and found tag, and exit 70 in both backends. There is no diagnostic recovery or catchable exception. Identity is preserved.

Typed property/index reads decode actual tagged storage, including raw scalar array slots and boxed arrays. Literal arrays record their element representation; slices propagate it. Mixed scalar recovery unions preserve distinct null/undefined tags. Object field metadata is retained when structural reads need it. Existing view and readiness hooks stay in place. RegExp replacement callback any parameters admit all provider tags, including strings; the unchanged stock decodeEntities declaration exposed and proves this correction.

The conservative alias rule is functional: prepared contracts reject field writes/updates, reflection and array methods/iteration throughout the program. This includes unrelated operations with matching field names or arrays. Contracts are prepared from contextual dynamic identifiers before module bodies; unprepared structural uses remain NotYet. Finite-depth aliases are supported; recursive schemas, index signatures, MapLike<any>, callable contracts, any fields in typed objects and nullable reference contracts that lose null/undefined distinctions remain refused or NotYet. Other raw-array producers without representation metadata remain a limitation. JSON.parse is not implemented by this slice.

Changed hooks are prepareJSONChecks/jsonViewHazard in refusals.go, checkedJSONType through checkedAnyType, jsonRecoveryUnion in expression.go, and the Any/Unknown case of regexReplacementKind. Native dispatch/read hooks are in emit_expressions.go; dynamic metadata selection is in emit_objects.go. No lower.go, emit.go, native.go, oracle_test.go or splitter edits. No cohere source was copied.

Witnesses and counts

Assets are .a files copied to temporary .ts files by the focused harness. Semantic reductions cover convertToObject/JsonConversionNotifier (commandLineParser.ts 2430/2467), convertJsonOption object/list uses (3803), and typed scalar uses around normalizeNonListOptionValue (3835). These reductions do not claim the complete original functions lower. json_validated and the earlier config_diagnostics/config_validated witnesses run source typeof/null validation and retain Node diagnostics without inserted failures. stock_decode_entities contains stock jsx.ts decodeEntities with only line-ending normalization; its dependencies are local implementations, not copied cohere code.

TestCheckedAny has 16 witnesses; TestCheckedJSON has 21. Fitting Node, JavaScript and native runs match stdout/stderr/exit, with ASan/UBSan and leaks on successful native runs. Seven structural misfits pin exact terminal checks and compare source Node separately. Each boundary-removal IR mutant is tested independently in both generated backends. TestCheckedAnyUnsupportedContracts has 11 negative declarations. A separate method-refusal fixture and mutant prove that limit.

counts.md gains exactly 21 rows, one for each checkedJSONFixtures entry. The structural and literal-result witnesses previously used only as refusals now have execution/count rows. Failing runs record allocations/retains made before terminal exit; successful rows balance allocation/free, except static singleton ownership recorded by the existing counter conventions. There are no existing-row changes.

Independent source mutants remove object recursion, array recursion, literal equality, union alternative selection and finite-number validation; corrupt Boolean array decoding and collapse null into undefined; remove alias-write, update, reflection, iteration and array-method refusals; remove prepared-contract enforcement, and remove string admission from callback tags. Each is caught by its named focused witness or negative fixture. The recursive-check mutants may be caught by a later read at a different path; the pinned boundary message distinguishes that from the required point-of-use check. See structural-mutants.py for exact overlays and catchers.

Independent declaration coverage

Source pin is 050880ce59e30b356b686bd3144efe24f875ebc8; inventory ea1b2359 has 112 explicit-any and 159 any-declaration records. The 80 ledger file hashes are checked. extract-sites.cjs retains the selected declaration byte-for-byte and records its hash and UTF-16 source span. Selection is the smallest enclosing named function, variable statement, interface, alias or class; nested callback parameters stay inside that declaration. Imports and captures are ambient declarations referring to stock declaration-only type contracts. No runtime dependency bodies are substituted. Capture typing and strict upstream-checker errors can stop probes; these are not evidence that the original use is refused.

A checked classification requires a lowered guard and a pinned scalar message attributed to the site's binding. A refusal must fall inside the selected declaration; loader failures, stops outside it, and declarations that only transport/type any remain not reached. Successful probes emit C but do not execute ambient dependencies. The stock decodeEntities witness separately executes the three checked callback uses. The historical whole-entry checker stop remains in coverage-entry-before.json; it no longer replaces independent observations.

Reproduction

    node stage3/checked-any/extract-sites.cjs /tmp/checked-any-stock /tmp/checked-any-sites stage3/checked-any/evidence/coverage-entry-before.json > /tmp/checked-any-extract.log 2>&1
    go run ./stage3/checked-any/probe /tmp/checked-any-sites/manifest.json /tmp/checked-any-site-coverage.json > /tmp/checked-any-probe.log 2>&1
    go test ./internal/oracle -run '^(TestCheckedAny|TestCheckedJSON|TestCheckedAnyUnsupportedContracts)$' -count=1 -v > /tmp/checked-json-focused.log 2>&1
    python3 stage3/checked-any/structural-mutants.py > /tmp/checked-json-mutants.log 2>&1
    go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts > /tmp/checked-json-counts.log 2>&1
    go test ./stage1/cohere/typeaware -run '^TestVolumeAgreementAndMutants$' -count=1 -timeout 30m -v > /tmp/checked-json-volume.log 2>&1

No whole package or full gate was run. Setup reports go-build 43.452s, tests-deferred 43.694s, warm 43.695s, done 43.735s; nproc is 5, CPU quota 4. An attempted /usr/bin/time wrapper was unavailable and did not execute the test. Python monotonic timing records the actual volume attempts. The first actual attempt failed at a newly introduced synthetic-any refusal in 186.972s; the scanner and optional-property ABI regressions were corrected before final checks.

Final independent counts: 3 checked, 61 refused, 207 not reached. Refused records stop within the selected declaration; 100 not-reached records fail the loader and 107 stop in lowering or lower without an attributed checked use. The three checked records are decimal/hex/word callback parameters at jsx.ts:624. convertToObject stops at ElementAccessExpression (2468); convertJsonOption stops at an indirect predicate overload (3816); normalizeNonListOptionValue stops outside its selected body at a bodyless predicate (4032), so remains not reached. coverage.json contains every individual reason and hash.

Final focused execution passes in 16.525s, lowering regressions in 1.737s, existing compatibility witnesses in 25.852s and counts in 59.905s. The two additional negative fixtures pass separately, and all 14 source mutants are caught. Full-package testing is delegated to the fleet gate.

The corrected-volume attempt using Go's default timeout was stopped before that timeout; it passed normal/sanitized controls and eight mutants. The delivery verification uses the fleet's explicit -timeout 30m. Both earlier logs are retained rather than claiming a single successful attempt.

Delivery volume verification passes: wall_seconds=607.511; normal and sanitized controls have 26,054 identical finding bytes and 73 findings. All 15 checker fact mutants are caught by the independent byte oracle. The first actual attempt failed in 186.972s; a corrected run without the explicit fleet timeout was stopped and preserved; the final run uses -timeout 30m. This is not a claim that the command was executed only once. Setup/test outputs are retained in evidence/.

The volume test also checks a released checker handle: the real program terminates with panic 70; its released-registry mutant exits 0 and is caught by the required panic. This is in addition to the 15 byte-oracle mutants.
