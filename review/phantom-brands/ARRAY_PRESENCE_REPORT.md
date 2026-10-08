Built required-undefined array phantom brands and the requested presence and write refusals for all three field forms.
Commits: based directly on e3d5836cc1da8dddab96a1503f24d152c1e8f3be; implementation is the commit containing this report.
Checks: full lowerer 20.842s, uncached filtered oracle 6.106s, counts update 23.014s; vet and formatting passed.
Mutants: all twelve qualifying mutants failed behavioral, diagnostic or IR assertions, including the in-narrowing fixture.
Not covered: full repository gate and full parser driver; census is an observation on checker-rejected source, not native compilation of tsc.

@system_adamic's October 7 14:10 ruling resolves the pending choice in ARRAY_REPORT.md. Required undefined, void and optional undefined array fields are phantom. All retain the array representation and erase casts. Required undefined on primitives remains outside the primitive rule.

The unchanged parser-proof row 5 probe from 2179dd8 is now internal/oracle/testdata/native-sorted-array-brand.a. Its bytes match array-evidence/native-sorted-array-brand.a. Before this change the cast was refused at 7:12. Source Node, sanitized native, release native and the JavaScript backend now all print `2\n`, with empty stderr and exit 0. TestPhantomSortedArrayProbe pins those bytes and that output; the fixture also runs in the normal differential oracle and counts gate. Its counts are allocations/frees/retains/releases/peak/regions = 4/4/0/4/4/0. Required array casts are held by complete IR equality against the unbranded generic program.

Each of in, Object.hasOwn, hasOwnProperty, Object.keys, Object.values, Object.entries, Object.getOwnPropertyNames, object spread, Object.assign from the value, and JSON.stringify has an exact member/fix pin for all three array brand field forms. For example:

```
Adamic 0.1 refuses presence of phantom array brand member marker via Object.keys; read the brand member's undefined value without testing its presence; use a separate real field or Map for observable presence
Adamic 0.1 refuses a write to phantom array brand member marker; keep the brand member phantom and unwritten; store observable state in a separate real field or Map
```

Presence checks follow union arms and generic constraints. Literal keys naming real array members do not test the brand; dynamic keys can, so they are conservatively refused. Additional pins cover computed builtin/member calls, propertyIsEnumerable, generic interfaces, multiple Object.assign sources, direct and computed writes, nullish writes, destructuring and Object.assign targets. Assign targets with any phantom member are conservatively refused when sources are supplied. The existing length, sort, __proto__, constructor and index 0 pins now run for all three field forms. Data-bearing brands remain unsupported and the existing element, ownership and cycle proofs remain intact.

review/phantom-brands/array-presence-in.a preserves the original sorted body and narrows its result union on `" __sortedArrayBrand" in a`. Node prints `absent\n`; Adamic refuses the observation naming that exact member. Dropping the dedicated in refusal falls back to the pre-existing generic in refusal. The fixture's exact diagnostic assertion catches this; this mutant is a diagnostic failure, not a claim of a native output mismatch. The generic in prohibition is retained.

| Mutant | Intended catcher |
| --- | --- |
| in-refusal-dropped | Named brand/fix pin on the narrowing fixture |
| own-method-refusal-dropped | hasOwnProperty becomes accepted; presence pins fail |
| object-reflection-refusal-dropped | keys/values/entries/getOwnPropertyNames/hasOwn pins fail |
| spread-refusal-dropped | Object spread diagnostic pin fails |
| assign-source-refusal-dropped | Object.assign source diagnostic pin fails |
| json-refusal-dropped | JSON.stringify diagnostic pin fails |
| write-refusal-dropped | Direct, computed and destructuring write pins fail |
| generic-constraint-forgotten | Uninstantiated generic presence code becomes accepted |
| union-arm-forgotten | The narrowing fixture loses its named refusal |
| member-name-forgotten | A second-member write reports the wrong first member |
| required-undefined-rejected | The unchanged parser probe no longer lowers |
| required-cast-not-erased | Complete generic IR equality fails |

Every qualifying mutant exited 1, was restored, and has its log in presence-evidence. Three initial mutation scaffolds had unused Go variables; those failures were discarded and rerun with the variables kept used. The committed runner contains the corrected scaffolds. Final restored focused checks passed lowerer 1.847s and oracle 1.176s.

Validation commands, with all output redirected to log files:

```
source /workspace/adamic-tools/env.sh
go test ./internal/lower -count=1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestPhantom|TestNativeAgreesWithNode/internal/oracle/testdata/(phantom_|native-sorted-array-brand.a|library_object|object_prototype|generic_|invariance|proven_|fresh_|weak_)' -count=1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts
go vet ./...
gofmt -l cmd internal
```

Toolchain setup succeeded: Node 0.025s, Go 0.026s, submodules 0.075s, markdown dependencies 0.079s (validation step 0.009s), clang 0.175s, build 26.303s, deferred tests 26.448s, cache warm 26.450s, total 26.482s. nproc=5; cgroup CPU quota=4; Node 24.19.0, Go 1.27.1, clang 20.1.8. No push or merge into main or any area branch occurred; only codex/phantom-brands is the push destination.

Census observations: the archived 70456b7 method and the tooling README were read before using the counts. The exact archived 222 named-value reason filter measures 175 before and 175 after on the current e3d5836c baseline; its primitive-brand subset measures 45 before and 45 after. Both runs have 78 roots, 2,165 checker diagnostics and 3,731 unique NotYet/Refused/Panic sites. Their complete per-reason delta is empty. All 78 adapted source hashes match the archived manifest. There are no visible SortedArray/SortedReadonlyArray reasons in either ledger; earlier blockers and generic measurement limits prevent this census from proving the parser slice clears. The unchanged row-5 fixture is the direct before-refused/after-passing evidence.

The before configuration uses e3d5836c's phantom_array_brands.go and refusal visitor in a scratch Go overlay; the newly added presence helper is unreferenced there. The after configuration uses this feature's code with the same measurement overlay. Production files, production loading and production Lower were not changed for measurement. The same /tmp/phantom-tsc adapted corpus was used for both. Binary and overlay-source hashes, exact filters, the source manifest, complete compressed raw ledgers and logs are in presence-evidence/census-summary.json and the adjacent evidence files.

```
python3 /tmp/phantom-presence-before-generator.py "$PWD" /tmp/phantom-presence-before-overlay
python3 stage3/census/latent/make_overlay.py "$PWD" /tmp/phantom-presence-after-overlay
gofmt -w /tmp/phantom-presence-before-overlay/*.go /tmp/phantom-presence-after-overlay/*.go
go build -buildvcs=false -overlay=/tmp/phantom-presence-before-overlay/overlay.json -o /tmp/phantom-presence-before-census ./stage3/census/latent/tool
go build -buildvcs=false -overlay=/tmp/phantom-presence-after-overlay/overlay.json -o /tmp/phantom-presence-after-census ./stage3/census/latent/tool
LATENT_ASSERT_NO_OUTPUT=1 /tmp/phantom-presence-before-census /tmp/phantom-tsc/src/compiler /tmp/phantom-presence-before.jsonl
LATENT_ASSERT_NO_OUTPUT=1 /tmp/phantom-presence-after-census /tmp/phantom-tsc/src/compiler /tmp/phantom-presence-after.jsonl
python3 stage3/census/latent/audit.py /tmp/phantom-presence-after-census
```

The census audit passed continuation, all refusal sites, no IR output, disabled production loading, body skips, signature eligibility and same-line ranges. Its planted-function, signature/body range and misattribution mutants were caught. These measurement mutants are additional to the twelve production mutants above. Census measurements remain observations of a checker-rejected program, and are not successful native compilation of the compiler corpus.
