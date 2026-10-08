Held group fifteen's three array pairs against complete original TypeScript declarations; no compiler change was needed.
Commits: started at 1ff63ff942647d7eafe3ba2d4bebb71da7ba0e8e; integration bd05075f5b8cbd592e8ece1a098d582eb43cf510 merged as 0f47b23cf1499f24ff0ae3961a0a85c20d9c7fa4; group commit is the commit containing this report.
Validation: 15 probes compare Node, sanitized native, release native and JavaScript; six finishing probes pass leak checks; 15 measured count rows recorded.
Mutants: native numeric-field bypass produced unchecked numeric output; JavaScript numeric-field bypass printed bad and exited zero; both were caught by function-wrong-pos.
Limits: required global counts refresh fails in existing fixtures, with two failures reproduced at the integration baseline; production reachability, tuples, union-target casts and remaining pairs are unmeasured.

| Original candidate | Type id | Static candidate reads | Declaration |
| --- | ---: | ---: | --- |
| FunctionExpression.typeParameters | 6893 | 4 | readonly typeParameters?: NodeArray<TypeParameterDeclaration> |
| GetAccessorDeclaration.typeParameters | 6910 | 4 | readonly typeParameters?: NodeArray<TypeParameterDeclaration> |
| SetAccessorDeclaration.typeParameters | 6912 | 4 | readonly typeParameters?: NodeArray<TypeParameterDeclaration> |

Original microsoft/TypeScript revision: 050880ce59e30b356b686bd3144efe24f875ebc8.
The preparation script validates pristine tracked upstream files, generates upstream's diagnostic map with its own script, and uses lane 4b's existing declaration emitter. It emits 78 declaration files including imports. The manifest pins their digests, complete field lists and all twelve source spans. Receivers retain 39/36/36 fields; TypeParameterDeclaration retains 20. The oracle asserts those complete field sets in the generated view contracts. No reduced receiver or element schema substitutes for these declarations, and no cohere code is copied.

Each receiver has valid present, absent and explicit undefined controls; an unread malformed array control; a read wrong-array rejection; a wrong descendant pos rejection; and a missing descendant pos readiness rejection. Node determines each source outcome; checked executions either match it or reject exactly at the specified read with exit 70. Full declarations include the previously reported callable text member. None of these probes reads text, and none triggered that frontier. This observation does not certify text reads.

| Family | Candidate pairs / reads | Cumulative fixture obligations held | Remainder |
| --- | ---: | ---: | ---: |
| Array contracts | 334 / 3189 | 148 / 2713 | 186 / 476 |
| Element or consumer reads | 251 / 1602 | 0 / 0 production credit | 251 / 1602 |
| Intrinsic contracts | 179 / 794 | 0 / 0 production credit | 179 / 794 |
| Own array fields | 30 / 72 | 3 / 26 | 27 / 46 |

The cumulative array total comprises the inherited 145 representative obligations and these three complete-original-declaration fixture obligations. Static read totals are candidate witnesses, not observed runtime coverage. The remainder includes lane 4 union-target casts and previously refused obligations. Tuples remain with their worker. Continue with the remaining four-read candidates, including SourceFile.patternAmbientModules and declaration modifier arrays; consult the prior reports because the priority JSON's held list predates later groups.

Reproduction (source the environment path printed by cloud/setup.sh first):

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/adamic-setup.log 2>&1
source /workspace/adamic-tools/env.sh
npm install --prefix stage3/api --no-save typescript@6.0.3 > /tmp/lane2-api-install.log 2>&1
# Use a pristine checkout of microsoft/TypeScript at the revision above.
node stage3/interface-downcasts/lane2/original15/prepare.cjs /workspace/lane2-original-archive /workspace/lane2-original-declarations > /tmp/lane2-group15-prepare.log 2>&1
export ADAMIC_ARRAY15_ORIGINAL_DECLS=/workspace/lane2-original-declarations
go test ./internal/oracle -run '^TestCheckedViewRanked15' -count=1 -v -timeout 10m > /tmp/lane2-group15-restored.log 2>&1
python3 stage3/interface-downcasts/lane2/original15/run-mutants.py > /tmp/lane2-group15-mutants.log 2>&1
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 30m -args -update-counts > /tmp/lane2-group15-counts.log 2>&1
go test ./internal/oracle -run '^TestCheckedViewRanked15ArrayCounts$' -count=1 -timeout 10m -args -update-counts > /tmp/lane2-group15-own-counts.log 2>&1
go vet ./internal/oracle > /tmp/lane2-group15-vet.log 2>&1
```

Initial behavior run passed in 22.523s; measured group counts refresh passed in 13.452s. Restored final verification passed in 34.343s (15 behavior probes plus the measured-count test); go vet passed with no output. Their outcomes are retained alongside this report. The declaration-dependent tests explicitly skip without the external generated inputs. The existing counts registry retains these recorded rows without claiming remeasurement when inputs are absent, and remeasures them when supplied. No whole package test or full gate was run.

The required counts command failed after 69.598s, before reaching the additional interface count registry. Existing graph fixtures report free(): invalid pointer, existing process.exit fixtures refuse lowering, and other existing fixtures also fail; see evidence/counts-required.log for the complete observed set. To separate this group's changes from those failures, a detached worktree at merge baseline 0f47b23c ran exactly:

```sh
go test ./internal/oracle -run '^TestCountsAreRecorded$/^fixtures$/^internal$/^oracle$/^testdata$/^(graph_regions_regression_06.a|process_exit.a)$' -count=1 -v -timeout 10m > /tmp/lane2-group15-baseline-counts-corrected.log 2>&1
```

Both failures reproduce there: graph_regions_regression_06 aborts with invalid pointer; process_exit refuses lowering. This establishes baseline status for those two witnesses, not every other failure. Their fixes are outside this lane's territory. The separate group counts refresh adds only fifteen measured rows and leaves existing rows unchanged; it does not make the required global refresh green.

Native mutant: return any numeric field slot without checking its stored kind in object.c. function-wrong-pos executes in sanitized and release native, exits zero and prints unchecked numeric values instead of the required rejection; the test fails (61.886s). JavaScript mutant: make the numeric field predicate unconditional in readiness.go. The same witness executes, prints bad and exits zero; the test fails (3.285s). Neither kill is a compilation failure. The runner restores both files in finally, and the restored source is verified afterward. These are temporary mutants, not production edits.

Toolchain setup succeeded: Node 0.081s, Go 0.126s, clang 0.626s, markdown 1.310s, submodules 298.402s, build 620.999s, tests deferred 621.247s, cache 621.249s, done 621.310s. nproc=5 (CPU quota 4). Go 1.27.1, clang 20.1.8, Node 24.19.0. The initial normal fetch did not include this branch; an explicit branch fetch resolved the exact requested SHA. Integration merge conflicted only in docs/checked-views-plan.md; both sections were retained. No code conflict occurred.
