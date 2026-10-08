# Method values continuation

The continuation starts at f7de26db on codex/method-values. Each completed step
is committed and pushed to that branch. No protected orchestration file is edited.

Setup: GOPROXY=https://proxy.golang.org|direct, bash cloud/setup.sh, then
source /workspace/adamic-tools/env.sh. The setup log is
/tmp/method-values-followup-setup.log. Times in seconds: go .029, Node .036,
submodules .098, markdown .099, clang .249, build 45.016, warm 45.130,
done 45.215. nproc is 5, quota 4 CPUs.

## Static extraction

Static allocation now installs an inherited callable table and distinguishes own,
non-enumerable methods from inherited ones. Function identity is shared with the
parent when the method is inherited; an override gets its own callable. JavaScript
constructor objects install only own methods and inherit the parent constructor.
The exception analysis includes static methods that escape as callables.

The Node oracle covers free extraction, bind with a derived constructor receiver,
identity, own-property reflection, enumeration, overrides, and a caught unbound
static this read. The initial oracle caught omitted static exception propagation:
its pending error had been ignored and printed 0. Static methods now participate
in callable throw propagation. All final commands passed:

```sh
go test ./internal/oracle -run 'TestMethodValuesTypeScript/static' -count=1 -timeout 10m > /tmp/method-values-static-typescript-fixed.log 2>&1
go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/(method_values|class_static)' -count=1 -timeout 10m > /tmp/method-values-static-adamic-fixed.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts > /tmp/method-values-static-counts.log 2>&1
python3 notes/method-values/ruling/run-mutants.py > /tmp/method-values-static-mutants.log 2>&1
```

Outputs: TypeScript oracle ok .406s, Adamic and static regressions ok 1.065s,
counts refresh ok 26.720s. All seven mutants exit 1 with intended catchers:
the original five plus static-throw-hidden (Node stdout disagreement) and
static-own-hidden (Node stdout disagreement). Source is restored after every
mutant. The fixtures run native with ASan/UBSan, release native, and the
JavaScript backend against source Node, with successful-run leak checks.

## Optional extraction

Optional receiver extraction evaluates its receiver once and skips the complete
binding expression, including receiver-producing arguments, when absent.
Optional method signatures can read absent own fields as undefined; class
implementations still use their callable table. Both .a and temporary .ts
fixtures cover present and absent receivers, lazy bind arguments, missing
optional members, and a caught unbound this-reading optional extraction.

```sh
go test ./internal/oracle -run 'TestMethodValuesTypeScript/optional' -count=1 -timeout 10m > /tmp/method-values-optional-typescript-2.log 2>&1
go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/method_values/optional' -count=1 -timeout 10m > /tmp/method-values-optional-adamic-2.log 2>&1
go test ./internal/lower -run 'TestMethodValuesProof|TestMethodValuesSelectiveChecks|TestMethodBindCycleIsRefused' -count=1 > /tmp/method-values-optional-lower.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts > /tmp/method-values-optional-counts.log 2>&1
METHOD_VALUES_MUTANTS=optional-receiver-guard,optional-member-hidden python3 notes/method-values/ruling/run-mutants.py
```

TypeScript and Adamic oracles each passed .495s, lower passed .204s,
counts refresh passed 26.170s. Optional-receiver-guard exited 1 under sanitizers.
Optional-member-hidden exited 1 with a Node exit-code disagreement. The first
runner expectation incorrectly looked for stdout instead of the earlier exit-code
comparison; after correcting that expectation, the mutant was rerun and caught
(/tmp/method-values-optional-member-mutant.log). Source was restored.
Optional binding of an absent callable (m?.bind) remains a separate boundary.

## Rest extraction

Method callables pack a fresh rest array at their boundary, rather than treating
an array slot as one ordinary incoming argument. Class thunks and literal method
closures share the same runtime packer; JavaScript mirrors that convention.
Direct class method calls use that callable convention when the method has rest.
Spreads copy their slots at the source position and retain reference elements
before later arguments can mutate the source. No source array is reused as rest.

Fixtures cover fixed parameters plus rest, empty rest, class/static/literal
methods, direct calls, bind, callbacks, spreads, array independence, dynamic
strings, and later argument mutation. Temporary .ts inputs also check caught
this-reading rest methods. Generic rest methods, super rest calls, and spreads
filling fixed parameters remain explicit boundaries.

```sh
go test ./internal/oracle -run 'TestMethodValuesTypeScript/rest' -count=1 -timeout 10m > /tmp/method-values-rest-typescript-2.log 2>&1
go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/method_values/rest' -count=1 -timeout 10m > /tmp/method-values-rest-adamic-2.log 2>&1
go test ./internal/lower -run 'TestMethodValuesProof|TestMethodValuesSelectiveChecks|TestMethodBindCycleIsRefused|TestCensusRestMutableElements' -count=1 > /tmp/method-values-rest-lower.log 2>&1
go test ./internal/oracle -run TestMethodValuesTypeScript -count=1 -timeout 10m > /tmp/method-values-rest-regression.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts > /tmp/method-values-rest-counts.log 2>&1
METHOD_VALUES_MUTANTS=rest-reference-without-retain,rest-spread-without-owner python3 notes/method-values/ruling/run-mutants.py > /tmp/method-values-rest-mutants.log 2>&1
```

All passed: rest TypeScript .574s, rest Adamic .589s, lower .233s, all method-value
TypeScript cases 1.343s, counts refresh 26.918s. Both mutants exit 1 with ASan
heap-use-after-free on dynamic strings. Each source was restored.
