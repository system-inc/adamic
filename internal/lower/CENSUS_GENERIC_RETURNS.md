Built: checker-resolved generic return substitutions and structural method dispatch preserving receiver evaluation, this, and callable lookup order.
Commits: generic 9642bb4; structural 1d5ea34; integration merge 25ffa7c (origin/main e8ba3d5 merged into this feature branch only).
Commands/results: source Node, JavaScript backend, native release and sanitizer comparisons pass; merged lower/full oracle gate, vet and original branch-configuration census pass; generic 123 to 123, structural 110 to 0.
Mutants: four runtime mutants fail Node stdout with clean sanitizer exits; planted corpus refusal, scope/misattribution and twelve report artifact mutants are caught (details below).
Limits: generic census remains 123; unchanged forEach/find bodies need language decisions; no claim that the checker-rejected tsc project compiles.

## Delivery and files

No pushes or merges into main or area/* were performed. Only codex/census-generic-returns is pushed. Scratch census branches remain unpushed. No pull request is opened.

Shared production files edited: internal/lower/generic.go (one substitution initialization hook) and internal/lower/class.go (replace the blanket structural-signature refusal with the dedicated lowering). Shared generated file: internal/oracle/counts.md. The protected lower.go, native emit.go/native.go, and oracle/oracle_test.go were not edited by this unit; incoming integration changes are separate merge ancestry.

New implementation files: internal/lower/generic_returns.go and internal/lower/structural_methods.go. New oracle registration/mutant files: internal/oracle/census_generic_returns_test.go and internal/oracle/census_structural_statics_test.go. New source fixtures: internal/oracle/testdata/census_generic_returns.a and internal/oracle/testdata/census_structural_statics.a. No cohere implementation was copied; the existing checker shim and mapper machinery are referenced.

Generic monomorphization now reads the resolved checker signature mapper before existing inference. This supplies the type parameter when an explicit type argument and an undefined argument otherwise collapse inference. Existing native representations remain responsible for nullable numbers, strings, objects, and boxed unions.

Structural calls use the existing method thunk when no compatible constructor object exists. Otherwise lowering snapshots the receiver and any own function-valued field before evaluating arguments, then dispatches constructor objects through their virtual static slot and instance receivers through their method thunk. Static candidates are selected by checker assignability and actual runtime class identity, in deterministic order. Different native signatures and optional calls on possible constructor receivers remain explicit NotYet cases.

## Fixtures and observable behavior

TypeScript 6.0.3 source is pinned at 050880ce59e30b356b686bd3144efe24f875ebc8. Generic fixture bodies are core.ts:1083 firstOrUndefined and core.ts:1116 lastOrUndefined, unchanged except removing export. Tests cover zero, empty arrays, undefined arrays, dynamic strings (including empty), and reference objects.

The structural fixture retains the exact createEmptyExports body at factory/utilities.ts:183, including its comments and nested factory method call. Minimal supporting types exercise an instance and a constructor with identically named methods, static this binding, an inherited static receiver, literal function fields, side-effecting receiver expressions, and replacement of a callable through an alias during argument evaluation. The latter must call the old function, then call the replacement on the next invocation.

Each accepted .a source runs directly on Node using the oracle's original-source transformation. Its bytes are compared with both the JavaScript backend and native output, with release and ASan/UBSan builds. Source fixtures were run before the production hooks: the generic fixture hit the unresolved return diagnostic; the structural fixture hit the exact blanket refusal.

## Runtime mutants

TestCensusGenericUndefinedMutant replaces emitted missing-number return values with present zero. Node expects `0 7 undefined undefined`; the mutant prints `0 7 0 0`.

TestCensusStructuralMutants runs three independent mutations. Receiver twice adds and releases a second getter result: Node expects the receiver counter 2, mutant prints 4. Static instead of instance replaces the instance branch with Factory's static implementation: Node expects `instance:instance`, mutant prints `static-body:instance:instance`. Lookup after arguments disables the own-field snapshot: Node expects `old:argument`, mutant prints `new:argument`. All four mutants must exit zero, emit no sanitizer stderr, and fail specifically with `stdout differs`; a compiler error, leak, or crash cannot count as detection.

## Language decisions: stopped portions of the generic family

The following are the exact minimal programs tested, retaining the original bodies.

```typescript
function forEach<T, U>(array: readonly T[] | undefined, callback: (element: T, index: number) => U | undefined): U | undefined {
    if (array !== undefined) {
        for (let i = 0; i < array.length; i++) {
            const result = callback(array[i], i);
            if (result) {
                return result;
            }
        }
    }
    return undefined;
}
console.log(`${forEach<number, number>([0], value => value)}`);
```

Node prints undefined. Adamic rejects array[i] with TS2345: T | undefined is not assignable to T (line 4:37). Additionally, accepting `if (result)` requires a decision about generic JavaScript truthiness. Rewriting it as an undefined test changes the original result for zero, empty string, or false. Judgment: per-instantiation JavaScript ToBoolean can be sound, but adopting it is a language decision; this unit did not weaken the current boolean-condition rule or alter the body. Indexed-access proof is also needed before this original callback call is admitted.

```typescript
function find<T>(array: readonly T[] | undefined, predicate: (element: T, index: number) => boolean, startIndex?: number): T | undefined {
    if (array === undefined) return undefined;
    for (let i = startIndex ?? 0; i < array.length; i++) {
        const value = array[i];
        if (predicate(value, i)) {
            return value;
        }
    }
    return undefined;
}
console.log(`${find<number>([0], value => value === 0)}`);
```

Node prints 0. Adamic rejects predicate(value, i) with TS2345 at line 5:23. Judgment: do not disable checked indexed access or insert an unproved assertion. In particular the original optional startIndex can be negative, so the loop's upper-bound test alone cannot establish an in-bounds element of T. A sound acceptance needs a proof or a changed public contract; neither is a return-representation implementation decision. Further generic-family work stopped at these language decisions. The completed firstOrUndefined/lastOrUndefined implementation is retained; work proceeded on the structural family.

## Census method and results

Read the method in origin/codex/stage3-latent-census at 70456b7 before measuring. The original scripts prepared scratch cumulative integrations of stage3 and adaptations 10/20 with taste-not-soundness, flag-enums, namespaces-tsc and nested-functions. Preparation needed an environment correction for gofmt and removal of one duplicated stray scratch conflict-resolution fragment; neither touched production files. All 78 adapted source hashes match the published REPORT.json exactly.

The original run_comparisons.py generated overlays and ran LATENT_ASSERT_NO_OUTPUT=1. Each unit gets fresh lowering state; checker-diagnosed bodies are skipped; generic declarations receive no invented type substitutions. Counts deduplicate (kind, location, reason, exact diagnostic text). Every census count here is measured on a checker-rejected program. First-error shifts and isolated generic declarations limit interpretation.

Exact reason filters:

- `a function returning T`
- `a function returning T | undefined`
- `a function returning U | undefined`
- `a method call through a structural signature in a program with statics; use typeof the declaring class`

| Configuration | T | T or undefined | U or undefined | Generic family | Structural family |
|---|---:|---:|---:|---:|---:|
| Reproduced cumulative | 48 | 62 | 13 | 123 | 110 |
| Cumulative plus starting main e011f8f | 48 | 62 | 13 | 123 | 110 |
| Generic hook alone | 48 | 62 | 13 | 123 | 111 |
| Both hooks | 48 | 62 | 13 | 123 | 0 |

The extra structural site in the generic-only configuration is an exposed next blocker. Generic declarations still lack a concrete instantiation in this measurement, so the runtime generic fix does not reduce that exact diagnostic family. Removing the blanket structural refusal exposes other downstream findings; zero occurrences of that reason is not whole-project success.

Raw runs and logs: /tmp/generic-latent-runs; configurations /tmp/generic-latent-unit-configs.json; adapted roots /tmp/generic-tsc-adapted. Final feature-branch configuration and filtered evidence are recorded separately below.

## Toolchain and verification

`bash cloud/setup.sh > /tmp/adamic-generic-setup.log 2>&1`: Go 1.27.1, clang 20.1.8, Node 24.19.0 ready in 0s each; submodules 0s; cache warm 100s; total 100s. `nproc` prints 5; cpu.max is 400000/100000 (four CPU quota). Commands source /workspace/adamic-tools/env.sh. Test output is written directly to log files.

`ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/census_|TestCensus' -v -count=1 -timeout 30m` passed before the final snapshot extension. The final structural source/JS/native comparison passed in /tmp/structural-snapshot.log. `go test ./internal/oracle -run TestCensus -v -count=1` passed all four mutants in /tmp/census-mutants-final.log (1.310s).

`go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts` passed (14.155s), /tmp/census-counts-snapshot.log. New generic fixture allocations/frees 14/14, retains/releases 23/41, peak 10; structural 74/74, 86/128, peak 18; regions zero. Existing rows did not change before integration merge.

Final merged gate: `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle ./internal/lower -count=1 -timeout 30m` passed after merging origin/main: oracle 90.625s, lower 20.963s, /tmp/census-remerged-gate.log. `go vet ./...`, `gofmt -l cmd internal`, and `git diff --check` passed with empty logs. The full repository go test gate was not run; the complete oracle and touched lower package were run instead.

First feature push succeeded at 25ffa7cb9f2430ef7820bd35a90b0da7fc1827f2, /tmp/census-feature-push.log.

The historical audit.py initially failed its expectation of two nested-function refusals because the required cumulative configuration includes nested-functions. Its signature/body mutant was already caught. A scratch audit_cumulative.py changed only that expected count to zero; all continuation, body eligibility, same-line range, signature eligibility, refusal continuation, output guard, planted NotYet, and misattribution checks then passed, /tmp/census-cumulative-audit.log. The original historical failure is preserved in /tmp/census-structural-audit.log.
Final feature-branch census: the original runner measured a matching cumulative-plus-main baseline (scratch c4a08ea) and a configuration explicitly adding origin/codex/census-generic-returns at 25ffa7c (scratch 51d9aeb). Both builds and runs exit 0. Baseline build 13.093s/run 159.236s; feature build 12.751s/run 157.417s. Both have 78 roots, 1,985 checker diagnostics, 4,530 units, 255 diagnosed-body skips, 2,500 eligible function attempts, and 1,775 statement attempts. Exact family counts remain generic 123 to 123 (48/62/13) and structural 110 to 0.

Overall unique NotYet findings change 1,032 to 927; Refused 1,983 to 1,998; dependency skips 25, ordinary errors 1, and panics zero in both configurations. This demonstrates shifted first errors, not whole-program compilation. The final binary audit passes with the nested-function expectation correction described above, /tmp/census-final-binary-audit.log. The original summarize.py output is /tmp/stage3/census/latent/REPORT.json, complete raw events and logs are /tmp/generic-latent-final-runs, and the branch configuration is /tmp/generic-latent-final-configs.json.

CENSUS_GENERIC_RETURNS.json preserves the source manifest, provenance, checker/unit counts, family counts, exact filtered diagnostic texts and locations, binary/overlay/raw hashes, and full feature deltas. The complete scratch report preserves all non-family findings and raw compressed events. Independent report artifact audit passed, /tmp/census-final-report-audit.log. The original audit_corpus.py planted a NotYet in binder.ts:330:1 getModuleInstanceState: exactly one new finding, all other 77 files and all unit eligibility records unchanged, /tmp/census-final-corpus-audit.log. This fresh mutant replaced historical compressed evidence so that the original audit_report.py could audit the exact final baseline. Its initial attempts correctly rejected stale checker/corpus evidence and are not counted as passes.

The report audit independently recounted every raw event, source hash, checker span/body overlap, eligibility record, diagnostic location, family reason, and feature delta. It caught all twelve artifact mutations: source manifest, file coverage, checker recount, finding recount, reason recount, per-file recount, unit recount, measurement label, outside-root recount, common-unit recount, delta recount, and raw body status. The final synthetic binary audit also caught a planted extra NotYet, the signature/body range mutant, and misplaced-file evidence. These measurement checks validate the observation ledger; source Node/native tests establish the accepted fixture behavior.
