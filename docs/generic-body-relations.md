Generic body relations, roadmap step 16, task #js89dcw.

The compiler now proves dependent initializers, assignments and returns before instantiating a generic body. It keeps binder identities instead of replacing them with their constraints. A read of `source.value` from `source: T` keeps the dependent type `T["value"]`, even when the checker reports the property through its `number` constraint. The refusal names the slot, source, enclosing constraints and a fix: read the same dependent type, or give the slot the concrete constraint type.

The independent witness checks each relation with the actual checker type arguments when the body is instantiated. It reports an internal compiler error if the resolved relation fails. The resolved signature mapper covers arguments appearing only in indexed types. Class calls keep the resolved mapper already supplied by `classGenericCall`; applying the raw signature again would lose the caller's substitutions. The witness resolves polymorphic `this` to the actual class, while the body proof preserves the identity of a `this` read.

The exact supplied program is in `internal/lower/testdata/generic_body_relations/refused.a`. Stock TypeScript 6.0.3 reports zero strict diagnostics. The d72728e5 compiler accepts it and its emitted JavaScript prints `2` through the established Node loader. The new compiler refuses its initializer at 1:78:

```
generic body initializer into slot T["value"] from source 2 is not proven for every allowed type argument; checker constraint: T extends { readonly value: number; }
```

Both new positive fixtures agree with Node byte for byte, uncached, in release native, ASan/UBSan native and the JavaScript backend. Leak checks pass. `generic_body_read.a` prints `1\n1\n2\n2\n`; `generic_body_indexed.a` prints `1\n`. Existing generic class inheritance and generic iterators also agree with Node. The focused lower tests cover the original initializer, field initializers, independent binders, assignments, returns, nested accesses, key parameters, arrays, destructuring, object fields, index signatures, mapped keys, increments, compound assignments and aliases. A polymorphic `this` read keeps the existing NotYet for a function returning `this`, without inventing a body refusal.

The census is [generic-body-relations-census.json](generic-body-relations-census.json). It records every site with file, line, column, source, slot and relation, and hashes the complete compiler inputs and adaptation scripts. It finds **43 sites in 79 adapted compiler files**: 19 assignments, 11 returns, eight default initializers, three field initializers and two initializers. The checker's ordinary assignability relation accepts 42 of these; the declarations-transform assignment at 830 is already unassignable. This inventory walks every site independently of other refusals and lowering support. It does not claim these are 43 newly admitted programs that were miscompiled.

The cohere pin is `7945d102a6c18dd36adf9114a758ce646e8b2359`. Its tracked and untracked `.a` inventories are both empty: **zero files, zero sites**. Cohere's `.ts` sources were outside the requested `.a` census. No source was copied from cohere.

The conservative reading is that a different binder is not the same dependent type, even if it has a narrower readonly constraint. The rule also refuses generic function defaults that need another binder instantiation. Index values and the free components of mapped types participate in the same proof; their local key binder is excluded. These sites remain named in the census. Explicit casts retain their existing checks; this unit does not add a cast relation.

The unit started on main `031a1259` and was advanced, before its first commit, to main `7a10c877`. The advance was a fast-forward with no conflicts. Historical d72728e5 lower, flow and full oracle runs pass with the same pinned checker and Node type dependencies. The final runs use the current-main base. The five moved source results are:

| Source or test | Previous result | New result |
|---|---|---|
| `TestAMutableLocationSeenWiderIsRefused`, Narrow into mutable Pack | Refused at the wider mutable view | Refused at the initializer into Pack |
| Same table, readonly property containing Pack | Refused at the wider mutable view | Refused at the initializer into Held<Pack> |
| Readonly Pack/Narrow positive control, now `TestGenericBodyRelationsReadonlyBound` | Passed the refusal walk | Refused at the initializer into Pack |
| `stage3/fixtures/enums/06_set_node_flags.a` | Refused at the mutable cast, 79:10 | Refused at the assignment into T["flags"], 79:9 |
| `census_small_optional_stopped.a`, now `TestGenericBodyRelationsMaybeBind` | NotYet for dynamic this and bind | Refused at the dependent return, 8:5 |

The enum and maybeBind source observations on Node remain covered. No existing counts row changed. Two rows were added: `generic_body_read.a` has allocations/frees/retains/releases/peak/regions `10/10/0/10/3/0`; `generic_body_indexed.a` has `2/2/0/2/2/0`.

Mutants run through [the Go overlay helper](../internal/lower/testdata/generic_body_relations/mutants.py):

| Mutant | What catches it |
|---|---|
| Remove the body refusal hook | The original refusal fixture instead reaches the independent initializer witness: source 2 is not assignable to slot 1. The generic class field control likewise reaches its field initializer witness. The dictionary control loses the named generic refusal and instead reaches the existing index-signature refusal. All three refusal tests fail. |
| Return without checking the witness | `TestGenericBodyRelationsWitness` fails because no internal compiler error was produced for 2 into 1. |
| Ignore the resolved signature mapper | `TestGenericBodyRelationsIndexedReturnMapper` fails with NotYet for a return T["value"]. |

Every mutant fails a semantic assertion, not compilation or a warning. The first mutant exercises the real monomorphizer on the exact supplied program.

Commands and captured outputs:

```
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > setup.log 2>&1
source /workspace/adamic-tools/env.sh
nproc
npm ci --prefix stage3/api --ignore-scripts --no-audit --no-fund > npm.log 2>&1
go test ./internal/lower -run '^TestGenericBodyRelations' -count=1 -v > focus.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestNativeAgreesWithNode$/internal/oracle/testdata/^generic_body_(read|indexed)\.a$' -count=1 -v > fixtures.log 2>&1
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -args -update-counts > counts.log 2>&1
go test -p 1 -parallel 8 ./internal/lower ./internal/flow ./internal/oracle -json -timeout 30m > regressions.jsonl 2>&1
go test ./stage3/fixtures > stage3.log 2>&1
go test -overlay=platform.json ./stage3/fixtures -count=1 > stage3-platform.log 2>&1
go build ./cmd/adamic > build.log 2>&1
go vet ./internal/lower ./internal/flow ./internal/oracle > vet.log 2>&1
python3 internal/lower/testdata/generic_body_relations/mutants.py "$MUTANTS" > mutants.log 2>&1
GENERIC_BODY_CONFIG="$ADAPTED/src/compiler/tsconfig.json" GENERIC_BODY_OUTPUT=census.json go test ./internal/lower -run '^TestGenericBodyRelationsCensus$' -count=1 -v > census.log 2>&1
```

`platform.json` is an uncommitted Go overlay replacing the fixture platform condition with `false &&` that condition. No platform guard was changed in the delivered tree. The npm command completed successfully. Build and vet exit zero without output. Counts pass in 112.973 seconds; ordinary stage3 passes in 23.193 seconds and the lifted guard run in 19.828 seconds. The two uncached fixtures pass in 1.309 seconds. The final regression JSON records lower PASS (42.883 seconds), flow PASS (150.721 seconds) and oracle PASS (92.038 seconds). These are package elapsed values, not unit budgets. The baseline records lower PASS (69.933 seconds), flow PASS (245.385 seconds) and oracle PASS (648.536 seconds).

[Captured logs](generic-body-relations-logs/results.json) include the complete successful Go test events, the exact program before and after, focused fixtures, mutants, census, setup, counts and both stage3 runs. Event logs omit wall-clock timestamps and retain the other fields. The historical oracle result is the oracle package's pass from the recorded command; its earlier flow failure was rerun successfully in the separate lower/flow log.

Setup printed: Go and Node ready 0.025s, submodules ready 0.064s, markdown dependencies ready 0.078s, clang ready 0.160s, build ready 11.172s, cache primed 11.352s, done 11.390s. `nproc` prints 5; the container's CPU quota is four CPUs.

Initial test attempts exhausted disk or lacked the historical worktree's Node type dependencies and are excluded from the comparison. The dependencies were linked to the installed pinned package tree, disposable old build caches and abandoned test scratch were cleared, and the affected checks were rerun. The adaptation driver stopped when its Git snapshot after adaptation 71 exhausted disk. The remaining adaptations 75 and 76 completed in order after recovery; the census hashes identify that completed tree. No partial census or interrupted regression is reported as green.
