Excluded ambient host namespaces from executable initialization; real namespace reachability and readiness remain checked.
Commits: started exactly from aee98c837f787319262cf643021cb8b7c97efae1 on codex/namespace-init-sys; implementation SHA recorded by the subsequent evidence commit.
Commands and outputs: whole lower package passes (79.117s); both backend controls and five unresolved-read fixtures match Node (3.133s).
Mutant: treating a reaching namespace access as unreaching loses all three direct/helper/cycle refusal pins; mutation restored.
Not covered: exact fs.native and process.cwd reductions reach independent host NotYet in both backends; neither exact host program runs natively.

## What changed

namespaceRuntime now returns false for inherited ambient context, an explicit
ambient modifier or a declaration file. Such declarations describe the host;
there is no program statement that initializes them. The SCC call graph still
records executable namespace reads, and unresolved reads retain the existing
runtime readiness check. No host member, loader, ordinary enum rule or production
probe refusal was bypassed.

## Reductions and independent observations

[realpath.a](realpath.a) cuts sys.ts:1492:142 to the native-property selection.
[cwd.a](cwd.a) cuts sys.ts:1502:51 to memoize(() => process.cwd()). Both print
`true` with exit 0 and empty stderr on independent Node. Both pass the isolated
namespace preflight. Full native and JavaScript lowering then return named
host NotYet: `node:fs.native` at realpath.a:3:22 and
`node:process.Process.cwd` at cwd.a:7:43. [Exact commands and bytes](reductions.json).
These are measured host limitations, not native success or language decisions.

The structural controls explicitly substitute represented object values for the
unimplemented host. realpath-control.a prints `pathpath:false`; cwd-control.a
prints `cwdcwd:cwdcwd:false`. The latter uses a callback forwarder: the exact
memoizer's captured function parameter triggers the existing cycle-capable type
rule even after substituting the host. The initial control failure is retained.
These controls prove branch/deferred-callback behavior in both backends, not
implementation of fs.realpathSync.native or process.cwd.

## Sound boundaries and mutant

TestNamespaceAmbientHostInitialization loads both exact programs against the
pinned Node declarations, requires initialization success and requires an
independent named host stop. TestNamespaceAmbientContextsDoNotExecute checks
explicit declare, inherited ambient context and declaration-file context while
requiring an ordinary executable namespace to remain executable.

The existing direct/helper/cycle fixtures in internal/lower/testdata/namespaces_notyet
retain their located premature-read NotYet. The reaching mutant deletes only
namespace read recording in namespaceCallGraph.discover. All three
TestNamespaceInitializationReachability subcases fail with
`reaching call lost its initialization refusal: <nil>`. Compilation failure is
not the catch. Five existing unknown-before fixtures still lower, run and stop
with Node's exact error/effects in both backends, exercising the runtime fallback.

## Setup and validation

Setup succeeded in 262.537s, nproc=5, cgroup CPU quota four. Timing lines:
Go ready 0.074s, Node 0.073s, submodules 0.136s, markdown 0.237s,
clang 0.539s, Go build 260.622s, warm cache 262.321s.
Source environment: /workspace/adamic-tools/env.sh.
An initial host test run failed because @types/node 25.3.3 was missing after
switching baselines; npm ci --prefix stage3/api installed the pinned package,
and the rerun passes. No compiler behavior was changed for that setup failure.

Test output is retained in logs. The scoped oracle includes the two new
controls and five namespace unknown-before cases. The full lower run uses no
skips, clearing the baseline's two fs.readFile expectation failures.

## Production rerun

The requested baseline has no stage3/stricter-indexed-all/README.md. Its REPORT.md
reproduction commands are authoritative. The input is recreated using the
3b255125 adaptation worktree and npm ci. build-production-probe.py verifies all
79 dated source hashes: zero mismatches. The probe uses the original owning
project settings and the unchanged exact exclusion manifest. Only five catch
errors are excluded from the probe; ordinary production loading remains unchanged.
Production results follow after the implementation push.
