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
