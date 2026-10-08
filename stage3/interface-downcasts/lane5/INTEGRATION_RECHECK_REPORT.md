Built: merged newest fetched views integration and rechecked prior callable certificates and blocked reads; no additional pair certified.
Commits: integration 432d4913d31daaa49d8ca4eb46f30f21f90da5ee merged as b8afaa1972f4215196f4e7c0aab132b37944bfa9; this report and refreshed counts are pushed to codex/views-callables.
Checks: restored callable family fixtures, blocked-pair controls, native/JavaScript unknown-producer boundaries, exact entry-live-mutation refusal, original source evidence, vet and lane counts pass.
Mutants: no runtime check added; previous per-pair executable mutant evidence retained; the merged original-alias verifier mutation is caught and restored.
Uncovered: stopped on shared receiver conversion/admission dependencies; 156/2818 pairs and 2699/11063 candidate reads certified, 2662 pairs and 8364 reads remain.

Reporting date October 12. Continued from c8aafdab on codex/views-callables. The explicit integration ref fetch returned 432d4913, satisfying the requested 432d4913-or-newer boundary. The default fetch updates only main, so refs/heads/codex/views-integration was fetched explicitly. No other branch was pushed.

The merge has two reconciliations: counts conflict blocks have 403 callable rows and 178 incoming rows, with no common row keys; both sets are retained. The alias verifier retains per-family aliasFile lookup and incoming .ts/.a fixture support. Protected compiler changes are inherited from the requested integration, with no manual edits to those files. The lane-only measured updater changes 0 existing counts rows and adds no fixture row.

The former untagged-union controls now pass Node, native release/sanitized and JavaScript, with positive leak checks where the existing helper requires them: string-from-node, for-update, the earlier untagged variable-update reduction, arrow-function and literal-type. These reduced controls do not yet certify complete original family contracts with negative fixtures and mutants. Existing unbound-method controls, destructuring conversion refusal and the JavaScript Set unknown-signature refusal remain pinned. Emission-hook's CLI c probe now emits C; this alone does not certify its full nested callback contract. Optional realpath-binding and directory-condition CLI probes still refuse the cast with adamic/no-unchecked-cast. No optional method was converted to a function property.

Native Set add and has CLI c probes emit C, but both sanitizer build commands fail at the producer-to-receiver call: an adamic_map * is passed to an adamic_object * parameter and clang reports -Wincompatible-pointer-types. This is a compile-time representation boundary, not a runtime mutant result. No native Set binary runs or leak pass are claimed. The shared collection/view owner must provide that receiver conversion, after which the callable family still needs immutable intrinsic signature metadata, including add's self return. Optional-host method view admission and original bind/condition read contexts also require the shared object-view lane. New ranked certification stops here under the user's instruction to stop when another lane is required; no untested remaining family is declared blocked.

The integrator decision in docs/checked-views-blockers.md retains the Union callable adapter and unknown-producer refusals. TestViewCallableBoxingUnknownProducer passes in native and JavaScript, TestCheckedViewCallableRetainsNeverWiderHelper passes, and TestCheckedViewMapCertificates/entry-live-mutation retains its 5:69 Map callback refusal. No general Union refusal is restored and no Map callback site is relaxed.

Commands, all redirected directly to durable logs/integration-recheck:

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewCallableLaterRanked(UnionRefusal|BindingRefusal|MethodRefusal|IntrinsicSetRefusal|OriginalReadRefusals)$|^TestCheckedViewCallableRetainsNeverWiderHelper$' -count=1 -timeout 5m > /tmp/lane5-integration-blocked-recheck.log 2>&1
go test ./internal/native ./internal/javascript -run '^TestViewCallableBoxingUnknownProducer$' -count=1 -timeout 5m > /tmp/lane5-integration-boxing-boundaries.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewCallableLaterRankedFamilies$' -count=1 -timeout 15m > /tmp/lane5-integration-families.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewMapCertificates$/^entry-live-mutation$' -count=1 -timeout 5m > /tmp/lane5-integration-map-boundary.log 2>&1
go test ./internal/oracle -run '^TestCheckedViewCallableCounts$' -count=1 -timeout 10m -args -update-counts > /tmp/lane5-integration-counts.log 2>&1
node stage3/interface-downcasts/lane5/verify-later-ranked-fixtures.cjs /workspace/scratch/lane5-original "$(cat /tmp/lane5-night-final-ranks.txt)" > /tmp/lane5-integration-original-fixtures.log 2>&1
node stage3/interface-downcasts/lane5/verify-tagged-callable-carriers.cjs /workspace/scratch/lane5-original 122,141,144,145,159,160,176,177,192,194,212,217,241,249,278 > /tmp/lane5-integration-original-carriers.log 2>&1
go vet ./internal/oracle > /tmp/lane5-integration-vet.log 2>&1
```

CLI c probes use go run ./cmd/adamic c stage3/interface-downcasts/lane5/later-ranked-callables/<directory>/<fixture>.a for set-intrinsic/add, set-intrinsic/has, realpath-binding/good, directory-condition/good and emission-hook/good. Native Set probes use go run ./cmd/adamic build <fixture> -o /tmp/lane5-integration-set-<add-or-has> --sanitize, with separate build/run logs. An initial Map test name selected no tests; the exact TestCheckedViewMapCertificates command above replaced it, and its log records an executed test.

ok  	github.com/system-inc/adamic/internal/oracle	20.955s
ok  	github.com/system-inc/adamic/internal/native	0.308s
ok  	github.com/system-inc/adamic/internal/javascript	0.101s
ok  	github.com/system-inc/adamic/internal/oracle	396.225s
ok  	github.com/system-inc/adamic/internal/oracle	137.460s
ok  	github.com/system-inc/adamic/internal/oracle	0.226s
Verified 770 fixtures retain complete original declarations and reads.
Verified 101 fixtures retain requested original aliases, discriminators and enum values.

Oracle vet and git diff --check pass without output. No new fixture, full-package test or full gate. Static certification remains 4/308 pairs and 34/1503 reads. Setup reused: done 417.425s, nproc 5, quota 4 CPUs. Independent original TypeScript pin remains 050880ce59e30b356b686bd3144efe24f875ebc8. Earlier reports retain exact per-pair mutant and leak evidence.
