Built the P0 first-phase refusal for string-literal and literal-typed view element reads, with the exact diagnostic, location and dot-access fix.
Branch compiler/fx6-key-read starts from d054e3578d4e53bef8cb8c9ab82d6ace69cdafeb and was fast-forwarded to main a7448d73cd17f16362b6cbc5c5c111080da64e43 before delivery.
Full lowering, call-target reader guard, selected oracle refusal fixtures, counts refresh and go vet passed; commands and results follow.
The refusal-bypass mutant failed the refusal assertion and restored p05's native disagreement with source Node, both exiting 0.
Shared checked member lowering and the remaining P0 syntax routing are not delivered by this commit.

This delivers the conservative refusal toward task #1fk58py, item 132. The brief does not identify a numbered roadmap step beyond its task and item, so no other roadmap step is claimed. The amended order is followed: this first authored commit contains only the refusal phase, its fixtures, comment correction and evidence.

The view registration walks all modules for object element reads whose literal key names a checked field. This deliberately follows the existing conservative, whole-program field-name policy: aliases, function boundaries and source order cannot bypass the refusal. An unrelated object read with the same field name can also be refused in a program that registers that field. The fix is `use view.<field>`. The unsupported path never reaches either backend.

Integration's original p01, p05, p06 and p07 programs became available on the fetched main tip. Their unchanged source is held by separate top-level lowering tests and moved from the review agreement corpus to the active refusal corpus. Pending sidecars for p01, p05 and p06 are removed. Each correct control applies the advertised dot-access fix and supplies the right payload; each runs through `lowersAndAgreesWithNode`. No exit-70 intended sidecar is needed because these wrong programs are refused.

The final focused run (`go test ./internal/lower -run 'TestView.*Element' -count=1 -v -timeout 90s`) passed in 0.657s. Its eleven leaves, including setup within each leaf, were:

| Leaf suffix after TestView | Seconds |
|---|---:|
| StringElementReadRefused | 0.17 |
| LiteralTypedElementReadRefused | 0.26 |
| ElementFixAgreesWithNode | 0.42 |
| ElementP05Refused | 0.16 |
| ElementP06Refused | 0.16 |
| ElementP01Refused | 0.13 |
| ElementP07Refused | 0.06 |
| ElementP05CorrectControl | 0.20 |
| ElementP06CorrectControl | 0.33 |
| ElementP01CorrectControl | 0.37 |
| ElementP07CorrectControl | 0.46 |

Every test command writes to a log. Each test shell sources `/workspace/adamic-tools/env.sh`. The initial setup shell exported `GOPROXY='https://proxy.golang.org|direct'`, then ran `timeout 600 bash cloud/setup.sh`. Setup timing lines: Go ready 0.456s; Node ready 0.520s; clang ready 1.653s; markdown dependencies ready 2.673s; submodules ready 58.049s; go build ready 484.021s; build cache warm 484.196s; done 484.341s. `nproc` is 5; cpu.max is `400000 100000`, a four-CPU quota.

Initial outer limits of 180s for the focused test and 240s for the reader guard expired during cold compilation, before test output. The warmed retries passed. The first full lowering run failed only because `stage3/api/node_modules/@types/node` was missing. `timeout 180 npm ci --prefix stage3/api` installed the repository's pinned dependencies, and the next full lowering run passed in 70.771s. The initial counts refresh used `-timeout 90s` and timed out; a bounded retry with `-timeout 8m` passed in 96.174s. These failed and successful logs are preserved.

Final validation on the updated main tip:

```text
timeout 240 go test ./internal/lower -count=1 -timeout 180s
ok github.com/system-inc/adamic/internal/lower 163.175s

timeout 240 go test ./internal/oracle -run 'TestReviewProgramsRefuse/fxspptb_oct9_views_p0[1567]_|TestCountsAreRecorded' -count=1 -timeout 180s -args -update-counts
ok github.com/system-inc/adamic/internal/oracle 177.188s

timeout 120 go test ./internal/ir -run TestCallTargetReaders -count=1 -timeout 90s
ok github.com/system-inc/adamic/internal/ir 9.945s

timeout 120 go vet ./internal/lower
exit 0, no diagnostics

git diff --check
exit 0, no diagnostics
```

Counts were refreshed; counts.md has no diff. New source fixtures are lowering/refusal fixtures, not new counted execution fixtures. The full repository gate and full native/oracle packages were not run.

Run `timeout 400 python3 review/compiler/fx6-key-read/run-mutant.py` with the toolchain environment sourced. The script independently saves a reverse diff, disables the refusal by emptying its checked-field map, and restores the original source in a finally block. `TestViewStringElementReadRefused` fails with `got <nil>` (exit 1). The separate temporary native witness uses integration's exact p05 source, with release native C and source Node:

```text
source Node: exit=0 stdout="NaN\nfalse\n" stderr=""
native: exit=0 stdout="9.28870976802437e-310\ntrue\n" stderr=""
Native backend stdout differs; witness exits 1.
```

The native number is a run-specific observation, not a stable expected value. The stable evidence is the disagreement and the false/true mismatch with both programs exiting 0. This mutant is caught by behavior, not a C warning or sanitizer. Its evidence is `bypass-refusal.diff`, `mutant-refusal.log`, `mutant-native.log`, and `native-mutant-probe.go.txt`. The latter is intentionally noncompilable evidence, and its temporary test file is removed after the run.

Still required in the next routing phase: one shared checked member-read path for dot access, quoted element keys and literal-typed keys; destructuring; spread; `in`; and `Object.keys`, `Object.values`, and `Object.entries`. This commit does not claim their checks are unified. p19 and p72 are not separately activated or covered here. The first-phase control uses dot syntax because the corresponding view element syntax is deliberately refused.

Integration lane checks are run after committing, using the exact mandated fetch/show/Python command; their result and the delivered SHA are reported in the final response. This single-branch checkout needed explicit remote refspecs once to materialize origin/devtools/fast-gate and origin/cloud/merge-tree.
