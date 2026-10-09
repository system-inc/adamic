Built V4 point (a): direct checked-view calls validate producer arguments and view results at invocation.
Commit: this point is committed on compiler/views-v4 above 670a2939.
Tests: direct oracle 2.476s; focused lower 0.756s, native 3.036s, JavaScript 0.831s; all pass.
Mutants: scalar argument, result, object argument, array argument omissions; native/JavaScript early checks; unsupported view/producer guards all caught.
Not covered: points (b)-(f), escaping adapters, identity/no stacking, function surface, both-site blame, generic relations, lane 5 negative rewrite.

Commands (output in adjacent logs):

```sh
source /workspace/adamic-tools/env.sh
timeout 90 go test ./internal/oracle -run '^TestV4Direct' -count=1 -v -timeout 90s
timeout 90 go test ./internal/lower ./internal/native ./internal/javascript -run '^Test(ViewCallable.*|ViewCallables.*|PrepareViewCallableRead|CallableNamespace.*)$' -count=1 -v -timeout 90s
timeout 90 python3 review/compiler/views-v4/run-order-mutants.py
```

Each direct oracle runs Node, native release, native ASan/UBSan, and JavaScript, with native finishing leak checks. Four IR omissions produce clean Node-equivalent execution instead of required exit 70 in all three compiler execution modes. Early-check source mutants lose argument output and fail the order oracle. Two source mutants disable recursive-domain refusals and fail the path-and-fix assertion. No sanitizer kill substitutes for a semantic kill. All 16 new top-level leaves pass in 0.20-1.14s; individual times appear in direct-final.log.

Assumption: this point supports finite runtime-checkable parameter/result domains and refuses unsupported view contracts and known unsupported producer parameters. Recursive domains, callable payload domains, overloaded/generic/rest/spread calls are not widened silently. A nominal producer result can be checked structurally against the view result, retaining its original type name for blame.

No lane 5 commits were cherry-picked: their read-time rejection conflicts with the new ruling. This point implements the direct-call replacement; escaping-read adaptation and rewriting lane 5 fixtures remain later points. No aggregate propagation or sweep was attempted. No persistent fixture inventory rows were added; probes are inline temporary .a sources, so counts.md is unchanged. No scope-ruling skips were added.

Setup: the first setup run reached its outer 90s limit while warming the build; rerunning succeeded in 35.451s. Timing lines: clang 0.225s, go build 35.323s, cache warm 35.426s. nproc=5, cgroup CPU quota=4. GOPROXY used https://proxy.golang.org|direct. Setup log records the installed toolchain.

Required integration lane checks run after the commit; their output is in lane-checks.log. Push is a fast-forward to compiler/views-v4 only.
