# DiagnosticArguments rest arrays

Admit boxed union elements in named-function rest arrays through existing rest call packing.
Commits: base b410340dc8f889b5799c3bc519117c63def3aa24, replay merge 8b0db31c, fixtures 74b2ad70; implementation follows this report.
Focused uncached Node/backend/release/sanitized oracle passes; counts refresh passes, adding only two rows.
Restoring the union refusal fails both fixtures; dropping rest items fails their stdout comparisons in both backends.
General union-array indexing/methods, whole-program tsc compilation, and the full gate are not covered.

The unit-specific compiler-area base overrides the generic main-base instruction.
The replay merge includes 9a1f14c5d994aa855625e7cfa295677060348fec.
The table is origin/codex/stage3-notyet-table at 57b9777c8eb4ee28b1f50220e8c8fb51a2dfadf7.
Its raw.csv and roots/raw.csv each contain 353 matching observations at 46 unique
(kind, where, reason, text) root sites. This is representation-rule coverage,
not a claim that all 46 bodies or their whole project now compile.

Only the rest admission guard changes. censusRestCall already fits scalar items
into ir.Union boxes and copies spreads into new arrays. Existing array operation
guards remain, so admitting the parameter does not accept unsupported operations.
The two .a fixtures reduce binder.ts:567:78 and binder.ts:2746:72 to their fixed
arguments and forwarding behavior. They cover empty rest, string/number/undefined
items, runtime-built strings, evaluation order, repeated forwarding, and cleanup.
Their observations are lengths and side effects; they do not observe union payload
reads, which still stop at the general union-array element guard.

## Replay evidence

Prepared the table's recorded stage3 snapshot 9d534d3a in /tmp/rest-exact-input,
then ran its apply.sh into /tmp/rest-exact-adapted. All 81 source hashes match
stage3/notyet-table/source-manifest.json. Earlier setup probes used a different
adaptation snapshot; only the hash-matched runs below are delivery evidence.

With the old admission guard restored temporarily, both exact commands reproduced
NotYet a rest array of DiagnosticArguments. On the implementation:

- binder.ts:567:78 reaches NotYet a BinaryExpression with a value and a value
  at binder.ts:568:52 (getSourceFileOfNode(node) || file).
- binder.ts:2746:72 has no findings in its selected census unit.

Commands, from the repository root, with the setup environment sourced:

```sh
go run ./stage3/census/latent/replay -project /tmp/rest-exact-adapted/src/tsc/tsc.ts -where /tmp/rest-exact-adapted/src/compiler/binder.ts:567:78 -kind NotYet -reason 'a rest array of DiagnosticArguments'
go run ./stage3/census/latent/replay -project /tmp/rest-exact-adapted/src/tsc/tsc.ts -where /tmp/rest-exact-adapted/src/compiler/binder.ts:2746:72 -kind NotYet -reason 'a rest array of DiagnosticArguments'
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/rest_union_' -count=1 -timeout 10m
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts
```

Baseline replay exits 0. Post-change replay exits 1 because the requested old
signature is absent; stdout still records the selected unit and its new findings.
Focused oracle: ok in 15.777s, and restored final run ok in 0.559s.
Counts update: ok in 32.028s. New counts (alloc/free/retain/release/peak/region):
diagnostic 16/16/7/21/8/0; token 28/28/24/38/12/0.
Every test wrote directly to a log. /tmp/rest-mutant-reject-union.log records
both lowering failures; /tmp/rest-mutant-drop-rest-items.log records stdout
mismatches and successful compilation, so this mutant was not killed by warnings.
The original source was restored in a finally block after each mutant sequence.
/tmp/rest-exact-{before,after}-{567,2746}.{json,log} preserve replay output.
/tmp/rest-final.log and /tmp/rest-counts.log preserve validation.

Setup: GOPROXY=https://proxy.golang.org|direct; bash cloud/setup.sh, then source
/workspace/adamic-tools/env.sh. No setup workaround was needed. Timing seconds:
Go 0.261, Node 0.308, clang 0.990, markdown 2.959, submodules 29.210,
Go build 330.333, cache warm 330.689, done 330.881. nproc=5,
cpu.max=400000 100000. /tmp/notyet-rest-setup.log contains the full output.
No PR was opened and no whole package or full gate was run.
