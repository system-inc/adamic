# Parameter properties and namespaces: validation

Base: `origin/main` at `5d4c801`. Branch: `codex/parameter-properties-namespaces`.
The parameter-property step was committed and pushed as `aea3215`; the namespace
step also includes final parameter-property initialization and callback checks.

Setup printed Go 0s, clang 0s, Node 0s, submodules 0s, build cache warm 71s,
total 71s; `nproc` is 5, with a four-CPU cgroup quota. The environment file is
`/workspace/adamic-tools/env.sh`.

## Parameter properties

The two registered fixtures are `parameter_properties.a` and
`parameter_properties_ownership.a`. Both pass source Node, generated JavaScript,
native release, ASan/UBSan and LeakSanitizer. Callback captures and labels are
built at runtime so the ownership test does not depend on immortal strings.

Commands, with every test's output sent directly to a log:

```sh
source /workspace/adamic-tools/env.sh
go test ./internal/lower ./internal/load -count=1 > /tmp/adamic-pp-packages-final.log 2>&1
go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/parameter_properties|TestParameterProperty' -count=1 -v > /tmp/adamic-pp-green.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts > /tmp/adamic-pp-counts.log 2>&1
go vet ./... > /tmp/adamic-pp-vet.log 2>&1
```

Output: lower 7.665s, load 0.598s, filtered oracle/mutants 0.849s, counts 10.635s,
all `ok`; vet's log is empty. The gate is extended after the namespace step.

Nine parameter-property mutants were killed:

| Mutant | Catch |
| --- | --- |
| Omit a numeric parameter-property store | Node stdout; compiled cleanly, exit 0, clean sanitizers |
| Move the store after the constructor body | Node stdout; compiled cleanly, exit 0, clean sanitizers |
| Omit the callback slot's retain | ASan heap-use-after-free |
| Ignore readonly-to-writable views | `TestParameterPropertiesSoundness/readonly_view` |
| Ignore parameter properties in override invariance | `TestParameterPropertiesSoundness/mutable_override` |
| Skip field-initializer definite initialization | `TestParameterPropertiesSoundness/field_initializer` |
| Skip parameter-default definite initialization | `TestParameterPropertiesSoundness/early_default` |
| Treat a shadowed inherited slot as initialized | `TestParameterPropertiesSoundness/inherited_initializer` |
| Admit a function value with an unsupported dynamic receiver | `TestParameterPropertyCallbackReceiver` |

The first three are permanent oracle tests. Guard mutants temporarily removed
the named check, ran the named regression, and restored the source after each
run. Each failed an assertion; none was killed by a build error. Direct readonly
writes and external private/protected reads are rejected by the TypeScript
checker, with a mutable/public control admitted.

An initial early-field-read probe declared the class without constructing it,
so it never reached field-initializer lowering. The corrected probe constructs
the class through a variable declaration; the guard mutant is then killed.

Counts added: parameter_properties.a 39 allocations/39 frees, 24 retains,
66 releases, peak 10; ownership 120/120, 192 retains, 232 releases, peak 12.
All prior rows are unchanged. Neither program disables regions or Perceus reuse.


## Namespace census and subset

An independent stock TypeScript 6.0.3 AST walk found the ten runtime namespaces
listed in [the decision](namespaces.md), plus the type-only Status namespace.
The pinned source came from the v6.0.3 archive; the AST library came from the
6.0.3 npm package. No adapted compiler source or Adamic lowering was used in
that census. The inventory includes nested Debug.log and Parser.JSDocParser.

The two registered fixtures are namespaces.a and namespaces_modules/main.a.
They cover scoped names and private state, nested generic calls, type/value
coexistence, detached identity, structural copies and imported aliases.

Ten namespace mutants were killed:

| Mutant | Catch |
| --- | --- |
| Wrong scoped function implementation | Node stdout; compiled cleanly, exit 0, clean sanitizers |
| Wrong scoped constant | Node stdout; compiled cleanly, exit 0, clean sanitizers |
| Redirect a structural typeof-namespace copy to namespace functions | Source Node stdout; native and generated JavaScript print copy 5 6 instead of copy 104 204; both exit 0, clean sanitizers and leaks |
| Permit calls before runtime namespace initialization | TestNamespaceLimitsStayLoud/early_call |
| Permit namespace reads before initialization | TestNamespaceLimitsStayLoud/early_read |
| Permit namespace reopening | TestNamespaceLimitsStayLoud/reopening |
| Permit mutable exports | TestNamespaceLimitsStayLoud/mutable_export |
| Permit uninitialized private state | TestNamespaceLimitsStayLoud/uninitialized_state |
| Materialize an ambient namespace | TestNamespaceLimitsStayLoud/ambient |
| Accept an explicit namespace-function receiver | TestNamespaceLimitsStayLoud/explicit_receiver |

The two binding mutants are permanent oracle tests. Compiler mutants removed
one guard at a time and restored it after each regression run. Each guard probe
was admitted by the checker and became accepted by lowering under its mutant.
No build failure was counted. The structural-receiver mutation was held to the
whole source/native/JavaScript comparison.

The first base and shadowed-slot probes retained closures that captured their
objects. Their initialization mutants were masked by the independent cycle
refusal. The final probes pass a temporary callback to a function that invokes it, so it forms no
cycle; removing initialization protection then makes lowering accept the unsafe
program. The masked attempts are excluded from the nineteen killed mutants.

An oracle build initially failed with no space left on device. The generated
Go build cache occupied 31GB. Clearing it with go clean -cache restored about
30GB, and the namespace oracles passed after rebuilding. Source, fixtures and
logs were preserved. This failure is not counted as a mutant.

## Counts

The table grows from 262 to 266 rows; all 262 pre-existing rows are byte-for-byte
unchanged. These are the four new rows:

| Fixture | Allocations | Frees | Retains | Releases | Peak | In regions |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| parameter_properties.a | 39 | 39 | 24 | 66 | 10 | 0 |
| parameter_properties_ownership.a | 120 | 120 | 192 | 232 | 12 | 0 |
| namespaces.a | 36 | 36 | 17 | 57 | 13 | 0 |
| namespaces_modules/main.a | 9 | 9 | 3 | 15 | 5 | 0 |

## Existing-file integration edits

| File | Reason |
| --- | --- |
| internal/load/load.go | Disable erasableSyntaxOnly to admit the new syntax. |
| internal/load/load_test.go | Remove the obsolete syntax-option rejection; enum lowering stays refused. |
| tsconfig.json | Match the loader option. |
| oracle/node.mjs | Use independent Node transform mode for parameter properties and namespaces. |
| internal/lower/class.go | Recognize parameter properties as fields in layout lookup and constructor this checks. |
| internal/lower/class_inheritance.go | Include parameter fields in overrides/layout and place derived resets/stores around super. |
| internal/lower/functions.go | Bind the lexical parameter symbol, check defaults and emit base stores after defaults. |
| internal/lower/lower.go | Run namespace initialization preflight. |
| internal/lower/modules.go | Register namespace declarations with their checker symbols. |
| internal/lower/statements.go | Lower scoped namespace bodies in source order. |
| internal/lower/expression.go | Resolve qualified bindings/calls, preserve structural-object dispatch, and refuse unsupported dynamic callback receivers. |
| internal/lower/refusals.go | Replace blanket namespace refusal with subset checks. |
| internal/oracle/counts.md | Record the four new fixtures. |

Implementation and tests are in the new parameter_properties.go/namespaces.go
files and matching lower/oracle tests. No native emitter, runtime, IR or analysis
visitor changes were needed. The existing analyses see ordinary owned fields,
globals, closures and calls.

## Final gate

All commands sourced `/workspace/adamic-tools/env.sh`; output went directly to
logs, with no test-output pipes:

```sh
go vet ./... > /tmp/adamic-pp-ns-vet.log 2>&1
go test -count=1 -timeout 30m ./... > /tmp/adamic-pp-ns-gate.log 2>&1
go vet ./... > /tmp/adamic-pp-ns-vet-final.log 2>&1
go test -count=1 -timeout 30m ./internal/lower ./internal/load ./internal/oracle > /tmp/adamic-pp-ns-final-packages.log 2>&1
gofmt -l cmd internal > /tmp/adamic-pp-ns-format.log
git diff --check > /tmp/adamic-pp-ns-diff.log
```

The full gate passed every package, including Test262, native, oracle, Unicode,
all eleven cohere packages, and the TypeScript parser and scanner. Selected
reported times: lower 32.035s, native 305.866s, oracle 224.444s, Unicode 792.850s,
cohere/json 554.272s, TypeScript/parser 280.584s. After the final dynamic-callback
receiver guard and improved negative probes, the full affected-package and
oracle rerun passed: lower 25.640s, load 1.870s, oracle 34.748s. Both vet logs,
the formatting log and the whitespace-check log are empty.

The final counted-fixture regeneration passed in 10.523s. An independent table
comparison confirms all 262 original rows are identical and the four additions
balance allocations and frees. All nineteen mutants listed above were killed;
masked probes and environmental build failures are excluded.
