# Memoize: blocked optional callback adaptation

## Updated p2 compiler repeat

Merged origin/area/stage3 at b2c4549f into this branch, without rebasing:
merge commit `7b19183894b3b8869974968002d71e2d03dfa5ff` includes main
`48c05d091f0a43c31cbe051b1d6578d99eeedf19`. In a detached scratch worktree,
merged compiler's `adab0fe54fb8a53940b0c1481e2409b978c6050a` into that
commit, again without rebasing or conflicts. Scratch merge commit:
`46a5ded3e72f54d693a462968dff03382e126934`. Its compiler changes stay in
scratch. Both checkouts use cohere `7945d102a6c18dd36adf9114a758ce646e8b2359`;
the scratch worktree links the same initialized submodule directory.

**Neither unchanged A nor unchanged B builds on either compiler.** All four
`go run ./cmd/adamic build <unchanged-source> -o <output> --sanitize`
commands exit 1, before cycle analysis. The exact diagnostics (identical on
both revisions) are:

```text
adamic: /workspace/adamic/stage3/adapt/48-memoize/a.a:4:13: Adamic 0.1 refuses a value as a condition; compare it explicitly, like name.length > 0 or count !== 0
exit status 1
adamic: /workspace/adamic/stage3/adapt/48-memoize/b.a:5:13: Adamic 0.1 refuses a value as a condition; compare it explicitly, like name.length > 0 or count !== 0
exit status 1
```

This repeat does not observe a cycle diagnostic or native B execution.
Inspection of internal/lower/control.go:97 shows that the p2 presence test
accepts `ir.Object` with undefined; optional callable conditions still reach
the refusal. The compiler branch's notes describe a successful optional
clearing control, not the complete two-call memoize B probe. These facts
explain the observed stop; no compiler policy was changed to bypass it.

The probe sources are byte-identical to commit 34ac6b2, checked against git
objects. Their SHA-256 values, full commands, compiler revisions, build exits
and exact output are in `evidence/p2-repeat/results.json`, alongside all four
build logs. No binary or native output exists. The existing no-op adaptation
remains declined because neither candidate qualifies.

Reran verify.cjs against the previous applied tree, on Node v24.19.0:
original, A and B each exit 0, stderr empty, stdout `42 42 1\n`; A's emitted
JavaScript remains byte-identical. The one-byte mutant is caught again by
exact comparison against original Node stdout. This is a Node comparator
proof, not a native output proof. The fresh node-proof.json replaces the
verification script's historical native annotation with this repeat's facts.

Setup succeeded with GOPROXY `https://proxy.golang.org|direct`, sourced
/workspace/adamic-tools/env.sh. Timing lines: Go 0.027s, Node 0.027s,
markdown 0.080s, clang 0.208s, submodules 22.827s, build 289.062s,
tests deferred 289.270s, cache warm 289.271s, done 289.304s. nproc=5,
cgroup quota=4 CPUs; Go 1.27.1, clang 20.1.8, Node 24.19.0.

This follow-up reruns only the requested native probes, Node comparison and
output mutant; it does not repeat the full upstream oracle, lane, caller
check or host extraction. Their older results below retain their original
revision scope. All new evidence is in `evidence/p2-repeat/`. The rest of this
README is the historical report on 03ccf222.


Neither requested candidate qualifies on the requested base,
`origin/area/stage3` at `03ccf222dfeedef4bfe5cc3e219e4586d14a5c13`.
This unit records the failed probes rather than shipping an unproved rewrite.
`adapt.cjs` validates the actual TypeScript 6.0.3 memoize declaration and
reports `declined`, with zero source edits. The adaptation table therefore
adds a zero row and retains the existing total, 78 files / 5120 added / 5094
removed. No production compiler file, checker option, API sanction or reference
baseline was changed.

TypeScript input is v6.0.3, commit
`050880ce59e30b356b686bd3144efe24f875ebc8`. The branch was created directly
from the requested area base. No rebase, main push, force-push or PR is used.

## Ordered native attempts

`a.a` implements precisely A: `callback: (() => T) | undefined` and
`callback = undefined;`. `b.a` implements B with a local optional pending
callback. Each invokes the memoized result twice and prints both values and
its callback count. On this compiler both exact candidates stop at their
truthiness conditions, before cycle analysis:

```text
A: Adamic 0.1 refuses a value as a condition; compare it explicitly, like name.length > 0 or count !== 0
B: Adamic 0.1 refuses a value as a condition; compare it explicitly, like name.length > 0 or count !== 0
```

To expose the next failure separately, `a-explicit.a` and `b-explicit.a`
replace only their condition with `callback !== undefined` and
`pending !== undefined`. They are additional probes, not byte-identical A.
Both reach the capture-cycle refusal. Exact diagnostic messages:

```text
Adamic 0.1 refuses 'callback', a variable a function value captures and can be reached from what it holds, so the function holds the variable and the variable holds the function: a cycle reference counting can't free; write the function as a function declaration (function callback() {}), which captures nothing, or declare the variable Weak<...> and keep the function somewhere strong (adamic/cycle-capable)
Adamic 0.1 refuses 'pending', a variable a function value captures and can be reached from what it holds, so the function holds the variable and the variable holds the function: a cycle reference counting can't free; write the function as a function declaration (function pending() {}), which captures nothing, or declare the variable Weak<...> and keep the function somewhere strong (adamic/cycle-capable)
```

The logs preserve locations and command exits. All four builds exit 1.
There is no native binary, no native two-call success, and no native output
mutant proof. The background's B runtime result is not reproduced on this
base: B is refused before execution. Moving the callback into a local does
not remove this compiler's type-level capture cycle. Fixing the compiler is
outside this unit's territory.

## Independent Node and source proof

`verify.cjs` extracts memoize from the actual apply output and constructs
original, A and B in memory. Stock TypeScript 6.0.3 emits original and A
byte-identically: the added parameter union erases and the removed non-null
assertion already erased. No external TypeScript checkout is edited by this
proof. Node 24.19.0 runs all three, each with empty stderr, exit 0 and:

```text
42 42 1
```

Thus Node calls the callback once and returns the cached 42 on its second
call. The one-byte output mutant changes the first byte from `4` to `5`;
the exact stdout comparison rejects it. This proves the Node comparison,
not the blocked native comparison.

The AST inventory finds the requested **18** memoize calls directly in
src/compiler and **four more** in its factory subdirectory. All 22 are checked. `check-callers.cjs`
loads the original compiler project configuration with Node declarations and
compares the complete semantic diagnostic set, each caller's inferred result
type, source text and local diagnostics before/after an in-memory A rewrite.
It preserves strict and strictBindCallApply as configured upstream. It only
disables emit/composite/isolated-declaration output checks for this semantic
comparison; the upstream build remains unmodified. All 22 calls and the complete project have zero semantic diagnostics before
and after A; inferred caller result types are unchanged. The callback-to-string
mutant failed the caller comparison with exit 1.
This is a stock TypeScript compatibility proof, not a native acceptance claim.

B's emitted JavaScript introduces `let pending = callback`, changes the
condition and invocation from callback to pending, and clears pending rather
than callback. The parameter stays unchanged and is no longer assigned.
Its Node two-call output agrees. Because no native B qualifies, this unit
does not claim the requested universal B behavior proof or admit these bytes
into the adaptation series.

Additional guard mutants ran independently and exited 1: adding
`value = callback` to A changes emitted JavaScript and fails its exact identity
assertion; removing the original `!` in a scratch input makes adapt.cjs reject
unreviewed source drift. Neither mutant alters the delivered compiler or tree.
The output-byte mutant, signature mutant, emission mutant and source-drift
mutant are all preserved in evidence.

## Toolchain and reproduction

Setup first failed while fetching a promisor object for the nested TypeScript
submodule: `error: RPC failed; HTTP 503 curl 22 The requested URL returned
error: 503`, followed by `fatal: unable to write request to remote: Broken
pipe` and `fatal: could not fetch be8e70810f821fb7f9a121e09d8b781b20d4c92a
from promisor remote`. Two premature builds only reported the absent nested
`go.mod`; these infrastructure failures were discarded and both builds rerun.
Retrying setup succeeded. The retry timing lines are: Go ready 0.025s,
Node ready 0.025s, markdown ready 0.076s, clang ready 0.175s, submodules
ready 33.016s, Go build ready 79.513s, tests deferred 79.655s, build cache
warm 79.657s, done 79.689s. `nproc` is 5; cgroup quota is 4 CPUs. Node
24.19.0, Go 1.27.1, clang 20.1.8. Setup printed
`source /workspace/adamic-tools/env.sh`, sourced in all compilation shells.

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/memoize-setup-retry.log 2>&1
source /workspace/adamic-tools/env.sh
bash stage3/apply.sh /tmp/memoize-adapted --write-table > /tmp/memoize-apply.log 2>&1
NODE_PATH=/home/agent/.cache/adamic-stage3/api/node_modules node stage3/adapt/48-memoize/verify.cjs /tmp/memoize-adapted /tmp/memoize-node-proof > /tmp/memoize-node-proof.log 2>&1
NODE_PATH=/home/agent/.cache/adamic-stage3/api/node_modules node stage3/adapt/48-memoize/check-callers.cjs /tmp/memoize-adapted /tmp/memoize-callers-final.json > /tmp/memoize-callers-final.log 2>&1
NODE_PATH=/home/agent/.cache/adamic-stage3/api/node_modules node stage3/adapt/48-memoize/check-callers.cjs /tmp/memoize-adapted /tmp/memoize-callers-mutant-final.json --mutant > /tmp/memoize-callers-mutant-final.log 2>&1
# Each native build writes its own log and is expected to exit 1:
go run ./cmd/adamic build stage3/adapt/48-memoize/a.a -o /tmp/memoize-a > /tmp/memoize-a-build.log 2>&1
go run ./cmd/adamic build stage3/adapt/48-memoize/b.a -o /tmp/memoize-b > /tmp/memoize-b-build.log 2>&1
go run ./cmd/adamic build stage3/adapt/48-memoize/a-explicit.a -o /tmp/memoize-a-explicit > /tmp/memoize-a-explicit-build.log 2>&1
go run ./cmd/adamic build stage3/adapt/48-memoize/b-explicit.a -o /tmp/memoize-b-explicit > /tmp/memoize-b-explicit-build.log 2>&1
bash stage3/oracle/run.sh /tmp/memoize-adapted /tmp/memoize-oracle > /tmp/memoize-oracle.log 2>&1
bash stage3/lane/run.sh /tmp/memoize-lane > /tmp/memoize-lane.log 2>&1
```

All test output goes to logs. The full Adamic repository gate, native tsc,
Windows and alternate Node versions are not covered. Native callback caching
and its output mutant cannot be covered until a candidate builds.

## Full upstream oracle

The standalone default oracle ran all runners, four workers, no filter, on
Node v24.19.0. Install exited 0 (5.209s), build exited 0 (27.535s), tests
exited 1 (416.101s), wall 448.984s. Counts: **106,366 passing, one failing,
zero pending**. The only baseline difference is api/typescript.d.ts and the
only failed test is the sanctioned public API acknowledgement test. These
results validate the unchanged series, not a successful native memoize rewrite.
The before/after core.ts files compare byte for byte with SHA-256
`c39e1722d3218137acc2620f03dc1a5331c7eaa4c6980ff5ec8d15e55b483d87`.
The API sanction file is unchanged from the area base, SHA-256
`ccf34475d56c29d7f6f31b83b1ffaab53abf535faed65d3113b4f7aba4f683ea`.

The independent landing lane exited 0 with **PASS stage3 landing lane**.
Its fresh apply exited 0; its oracle reported the same 106,366/1/0 counts,
all runners and no filter. The declaration guard has no errors: 222 composed
declarations equal the 222 sanctioned declarations; 28 reference declarations
are an accepted subset. The sole failed title is exactly
`unittests:: Public APIs for typescript.d.ts should be acknowledged when they change`.
Lane wall time is 520.794s; its oracle install/build/test times are
2.711s / 25.157s / 364.954s. This run included the uncommitted unit directory
and zero-row table; execution.json records the area base commit before the
report/evidence commit. Only documentation and evidence were added afterward.
The lane and standalone oracle used separate trees and logs.
