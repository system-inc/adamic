Built: two lowering guards for static initializer calls and Node-observed side effects, task #jcpm9m9.
Commit: tests and exact M15 diff in 36953d4a097d1d1628cf3522cfce6e0ae0156863, based on current main 2f7d81f91623c81b86e2ce5a439fdca69ef2d2b8.
Commands and results: focused lowering and existing oracle tests passed after restoration; lane checks passed, including go vet.
Mutant: audit M15 failed both new guards and the existing class_features_static.a oracle fixture.
Not covered: stage1 ports, full packages, the full gate, and other initializer mutations.

The change adds only tests and review evidence. It serves the named static-initializer guard task by making a dropped call fail in internal/lower, where the side effect originates. The witness is written to a temporary main.a: `class Box { static { console.log('static side effect'); } } console.log('done');`. No registered oracle fixture was added, so recorded allocation counts are unchanged and counts regeneration was not needed.

Observed: before adding the guards, the existing registered fixture `internal/oracle/testdata/class_features_static.a` passed normally and failed with the exact M15 diff from audit commit 6291aef6. Its source Node output contains `unused static block`; its generated JavaScript loses that output and other static initialization effects. The existing oracle reports `JavaScript backend: stdout differs`. Its mutated native release binary also crashes, and ASan/UBSan reports a null string access. That native failure is recorded separately from the JavaScript output disagreement.

Observed: `TestClassStaticInitializerCallIsEmitted` locates the initializer function, requires exactly one main-level call, and checks that its receiver is the class object allocated immediately before it. With M15 it fails with `static initializer call count = 0, want 1`. `TestClassStaticSideEffectMatchesNode` runs the original source on Node and requires exactly `static side effect\ndone\n`, then compares generated JavaScript and sanitized native stdout with that independent observation. Both mutated backends exit successfully and print only `done\n`; both stdout assertions fail. Restoring M15 makes both tests pass.

Inference: the former lowering test only checked the presence of static class metadata and a minimum number of main statements, which can survive removal of the initializer call. The new IR assertion directly guards that missing operation; the Node comparison checks its observable meaning independently of the IR shape.

Each new test is a top-level parallel leaf. Restored durations on the instance with a four-CPU cgroup quota: IR 0.04 s, Node/JavaScript/native 0.28 s. The existing restored oracle leaf took 0.78 s. The native witness uses ASan/UBSan; the existing oracle also exercises native release and leak checking. Node and generated programs each have a 10 s process deadline, and every go test invocation has a 90 s test deadline plus a 120 s outer deadline.

Exact test commands, from the repository root after sourcing `/workspace/adamic-tools/env.sh` and exporting `GOPROXY='https://proxy.golang.org|direct'`:

```sh
ADAMIC_GATE_UNCACHED=1 timeout 120s go test ./internal/oracle -run '^TestNativeAgreesWithNode$/^internal$/^oracle$/^testdata$/^class_features_static.a$' -count=1 -v -timeout 90s
timeout 120s go test ./internal/lower -run '^TestClassStatic(InitializerCallIsEmitted|SideEffectMatchesNode)$' -count=1 -v -timeout 90s
git apply review/compiler/class-static-guard/M15.diff
# Repeat the applicable focused command; both mutant runs exit 1.
git apply -R review/compiler/class-static-guard/M15.diff
# Repeat both focused commands; both restored runs exit 0.
```

Actual command stdout and stderr were redirected directly to `existing-baseline.log`, `existing-M15.log`, `existing-restored.log`, `guards-baseline.log`, `guards-M15.log`, and `guards-restored.log` in this directory. The existing-fixture mutant check ran before the new test file was written. Each mutation was reversed before the next baseline run. The production file has no remaining diff.

Setup command: `timeout 240s bash cloud/setup.sh`, with the required GOPROXY set first. Setup succeeded. Timing lines are preserved verbatim in `setup.log`: Go ready 0.023 s; Node ready 0.027 s; submodules ready 0.070 s; markdown dependency validation 0.009 s and ready 0.081 s; clang ready 0.170 s; Go build ready 36.707 s; test binaries deferred 36.861 s; build cache warm 36.862 s; setup done 36.890 s. `nproc` returned 5; `cpu.max` was `400000 100000`, a four-CPU quota. Tool versions were Go 1.27.1, Node 24.19.0, and clang 20.1.8.

The branch was fast-forwarded from initial main 7709c91213f476eba7e3dbfbf4ee65e6988038cb to current main 2f7d81f91623c81b86e2ce5a439fdca69ef2d2b8 before the final mutant/restoration checks. No unlanded worker branch was merged.

Required lane command after committing:

```sh
timeout 60s git fetch -q origin main devtools/fast-gate cloud/merge-tree
timeout 150s bash -o pipefail -c 'git show origin/cloud/merge-tree:cloud/integration/lane-checks.py | python3 -'
```

Output: `lane checks 1.6 s: gofmt and tools on 1 Go files, t.Parallel on 1 test packages; vet 1 packages`. The Node command and native build tools are declared in integration's tools.txt. No new external tool was introduced. The same lane command is repeated after the evidence commit before pushing.
