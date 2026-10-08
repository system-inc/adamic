39bd8a72ffd11032b644a97c1d4920a29b594298: void calls execute once before producing undefined; 22 CSV roots observed reaching the rule.
Commits: replay merge 97d8ebc5, fixtures aa320e78, lowering c2d177c7, current-main merge 39bd8a72, replay evidence 65aea038.
Focused oracle/lowering tests and counts refresh pass; binder lowers its selected unit, checker reaches reading add.
All seven mutants fail: six Node stdout comparisons and one erased-callable refusal assertion.
Not covered: five retained erased-callable roots, 239 roots not reached in standalone replay, and subsequent unsupported value/type operations.

The requested compiler base resolved to `b410340dc8f889b5799c3bc519117c63def3aa24`.
Started from current main as required, fast-forwarded to that compiler base, and
merged `codex/stage3-census-replay` at
`9a1f14c5d994aa855625e7cfa295677060348fec`. Current main subsequently resolved
to `ef3141e9b1152ab51b51497f8ce3a2799449c8a3` and merged without conflicts.
That final merge changes no compiler or oracle files. Each completed step was
pushed to `codex/notyet-void-value`; no PR was opened.

The two former NotYet branches now call `voidValue`. An ordinary IR helper
receives the original operands in order, evaluates the void operation, then
returns `ir.Undefined`. This uses undefined's existing object/null-pointer value
representation and requires no backend or IR changes. The call remains inside
its original expression, so conditional branches and exception paths keep their
meaning. A return from a void function evaluates its expression and emits a
valueless return. The helper preserves virtual call metadata; the compiler's
existing override-result representation proof remains unchanged.

The `.a` reductions preserve binder's conditional void return and checker's
early void return. Additional probes exercise skipped branches, argument effects,
virtual dispatch, throwing calls, array/map/set forEach, clear, and a void thisArg.
They are registered in their own `void_value_test.go`. The former thisArg gap
expectation is removed because its positive behavior is now held to Node.
Only three new counts rows were added; all existing rows are unchanged:
binder `5/4/11/18/2/1`, checker `4/4/14/16/3/0`, builtin `23/23/21/44/13/0`
(allocations/frees/retains/releases/peak/regions).

The clarified territory review found no kind skipped because of file ownership.
The existing IR and backends suffice for this rule, and no runtime C file changed.
Production edits outside the call-expression branch are
`internal/lower/void_value.go` (the called helper) and
`internal/lower/statements.go` (`returnStatement`). Test edits are
`internal/lower/library_map_set_test.go`, `internal/lower/void_value_test.go`,
`internal/lower/testdata/run-void-value-mutants.py`,
`internal/oracle/void_value_test.go`, its three `.a` fixtures, and `counts.md`.
After fetching the current NotYet refs, reviewed
`git log --glob='refs/remotes/origin/codex/notyet-*' -- internal/lower/statements.go internal/lower/expression.go internal/lower/void_value.go`.
The erased-callable worker changes `callClosure`, not this call-value dispatch;
its own optional-void report preserves the same actual-result distinction.
No other recent NotYet worker change touches `returnStatement`.
The focused tests of both touched Go packages are listed below.

A void callable signature alone is insufficient. TypeScript accepts
`const run: () => void = answer` when `answer` returns 7. Node still returns 7.
Calls through that erased signature retain NotYet. The existing refusal of an
override with a different native result representation is also pinned by the
focused test. Assumption: only operations with an already represented void ABI
can produce the synthesized undefined. Remaining erased calls require a proof
of their actual callees or an ABI preserving their actual results; treating all
void signatures as undefined would violate the Node oracle.

Both initial example commands exited 0 with `reproduced NotYet`:

```sh
source /workspace/adamic-tools/env.sh
bash stage3/apply.sh /tmp/notyet-void-adapted > /tmp/notyet-void-apply.log 2>&1
go run ./stage3/census/latent/replay -project /tmp/notyet-void-adapted/src/tsc/tsc.ts -where /tmp/notyet-void-adapted/src/compiler/binder.ts:3644:15 -kind NotYet -reason 'a void call used as a value' > /tmp/notyet-void-binder-before.log 2>&1
go run ./stage3/census/latent/replay -project /tmp/notyet-void-adapted/src/tsc/tsc.ts -where /tmp/notyet-void-adapted/src/compiler/checker.ts:8239:52 -kind NotYet -reason 'a void call used as a value' > /tmp/notyet-void-checker-before.log 2>&1
```

The same two commands on the final compiler, with logs named `*-final.log`,
exit 1 because the original signature no longer reproduces. This is the replay
tool's expected result when a stop disappears, not a green-signature claim.
Binder's selected `bindEnumDeclaration` unit at 3642:5 has no findings.
Checker's selected `bindElement` unit at 8237:37 now reaches
`NotYet: reading add` at 8242:41. These are measurements on the checker-rejected
entry-root corpus, not whole-program compilation claims. Raw JSON and diagnostic
output are in [binder-final](evidence/binder-final.log.txt) and
[checker-final](evidence/checker-final.log.txt).

Coverage is deduplicated from `stage3/notyet-table/roots/raw.csv` on table commit
`57b9777c8eb4ee28b1f50220e8c8fb51a2dfadf7`. Filter kind NotYet, phase lowering,
exact reason `a void call used as a value`, and empty blocked_by_kind; deduplicate
`(kind, where, reason, text)`. This produces exactly **266 root sites**.
Every site was replayed with the same guarded replay worker. A scratch-only
Go overlay adds `VOID_VALUE_LOWERED <where>` immediately before the successful
helper return; no production source receives that marker. Four independent
replay processes run concurrently, with stdout JSON and stderr saved separately.

**22 exact target sites reached the new rule, 5 retain the original stop, and
239 were not reached in their selected standalone replay context.** No coverage
is claimed for those 239: many stop earlier on casts or structural-signature
calls. All five retained targets call erased function values. The successful
and retained sites are in [covered-sites.csv](covered-sites.csv); all 266 statuses
and subsequent findings are in [coverage.json.gz](evidence/coverage.json.gz).
The compiled observation worker and full per-site logs remain at
`/tmp/notyet-void-coverage`; these counts measure rule reach, not whole-unit or
whole-program compilation. The scratch worker used the c2d177c7 lowering sources;
the main merge does not change them.

Every mutant starts from the final source and is restored in a finally block.
Run `ADAMIC_GATE_UNCACHED=1 python3 internal/lower/testdata/run-void-value-mutants.py`.
All seven go tests exit 1; none is a build or clang-warning failure:

| Mutant | Catcher |
| --- | --- |
| drop-call | binder fixture Node/native and Node/backend stdout |
| drop-array-visit | builtin fixture Node/native and Node/backend stdout |
| drop-map-clear | builtin fixture Node/native and Node/backend stdout |
| drop-map-visit | builtin fixture Node/native and Node/backend stdout |
| defined-result | binder fixture prints false where Node prints true |
| lose-return | checker fixture incorrectly runs element after returning pattern |
| erase-callable-result | closure refusal assertion sees got nil |

Exact final validation commands, each sent to a log file:

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle ./internal/lower -run 'TestNativeAgreesWithNode/internal/oracle/testdata/void_value_|TestVoidValueErasedResultsStayNotYet|TestLibraryMapSetGapsStayRefused' -count=1 -timeout 10m > /tmp/notyet-void-landed.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 10m -args -update-counts > /tmp/notyet-void-counts-landed.log 2>&1
go vet ./internal/lower ./internal/oracle > /tmp/notyet-void-vet.log 2>&1
```

Landing output: oracle `ok ... 1.198s`, lowering `ok ... 0.423s`; counts
`ok ... 50.896s`. Focused vet, gofmt on changed Go files, and git diff --check
produce no diagnostics. The oracle exercises source Node, release and sanitized
native (ASan, UBSan and leak checking), and JavaScript on Node. The restored run
before the main merge was oracle 0.542s, lowering 0.239s. No whole-package test
or full gate was run. [Focused](evidence/focused.log.txt),
[landed](evidence/landed.log.txt), [counts](evidence/counts.log.txt),
[mutants](evidence/mutants.log.txt), and individual mutant logs preserve outputs.

Setup used `export GOPROXY='https://proxy.golang.org|direct'` before
`bash cloud/setup.sh`. Its first build overlapped the compiler-base merge and
failed on unresolved typedArrayWrite, censusFieldSlotless and parameter-property
helpers; [initial output](evidence/setup-initial.log.txt) preserves every error.
The retry after the merge succeeded: Node 0.018s, Go 0.024s, markdown skipped
step 0.006s and ready 0.055s, submodules 0.070s, clang 0.169s, Go build 55.017s,
test binaries deferred 55.109s, cache warm 55.111s, done 55.135s. `nproc` is 5,
cgroup quota four CPUs. The printed `/workspace/adamic-tools/env.sh` was sourced
for every build and test. [Setup output](evidence/setup.log.txt).
