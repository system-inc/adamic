# Lint wave pre-push checker

Branch `codex/lint-wave-check` starts from fetched main `d090af531216ddd3c25a0dede6b82d7c0a6edf76`. This is infrastructure only: no lint rule, compiler, runtime, oracle fixture, counts table or dependency commit changed.

## Built

`cloud/lint-wave-check.py` checks a committed, branch-owned claim; freshly fetched origin heads for competing reservations and ports; directory-only rule registration; and current-commit parity/mutant evidence. The default command generates registration and runs real Go JSON tests to files. `--verify` checks the receipt and events again and refreshes origin without rerunning expensive comparisons. Receipts live inside Git metadata, so workers add no shared evidence-list lines.

[Worker instructions](../../lint-wave-check.md) specify the exact claim format, commands, prerequisite migration and limits. `cloud/lint_wave_test.go` runs the Python CLI regression suite in the normal Go gate. No arbitrary base, origin exclusion, existing-log acceptance or skip-tests switch is exposed.

Implementation commits: `2199eeecde4217d1c77322e0752d3f79a1070020` and `32a2f5c9368ba832a43740438d2bc5a2453d3b54`. The evidence-only commit follows.

Root remote.fetch was main-only. Discovery explicitly fetched every origin head (222 at initial inspection; 263 in the final live probe), not just main. Root's ordinary all-head fetch recursed into historical nested submodule commits and took several minutes; the checker uses `--no-recurse-submodules`. Initial guessed inventory-document path did not exist; the actual inventory instructions were read under stage1/cohere/lint/inventory. At initial inspection the wave claims were not published yet. During work, Markdown wave reservations and evidence appeared on origin. The first live probe misclassified an evidence overlay JSON as a claim and failed; this exposed an incorrect format assumption. The final code reads actual Markdown reservations, honors explicit skipped entries, ignores appended reports/evidence, and recognizes .a ports. Positive and negative CLI tests hold all these boundaries. Existing batch selection JSON and BATCH report names also count as reservations, including selections with no implementation.

## Setup and commands

`bash cloud/setup.sh > /tmp/lint-wave-setup.log 2>&1`: exit 0. Go, clang, Node and submodules ready at 0s; build cache warm 139s; total 139s. `nproc` = 5, cgroup quota four CPUs (`400000 100000`), 17.6 GB. [Exact output](setup.txt). All Go commands source `/workspace/adamic-tools/env.sh`.

- `go test ./cloud -count=1 -v -timeout 5m > /tmp/lint-wave-cloud-tests.log 2>&1`: exit 0, 11.343s; 43 Python CLI tests passed. [Output and every caught control](checker-tests.txt).
- `go vet ./... > /tmp/lint-wave-vet.log 2>&1`: exit 0, empty [output](vet.txt).
- `gofmt -l cmd internal cloud`: empty. `git diff --check`: empty.
- `go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$' -count=1 -v -timeout 10m > /tmp/lint-wave-oracle.log 2>&1`: exit 0, package 39.098s, test 39.05s. Native and Node cache misses; the one-byte mutant was caught by the external oracle. [Output](oracle.txt).
- In the isolated worker: `ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-6.0.3 python3 -B /workspace/adamic/cloud/lint-wave-check.py > /tmp/lint-wave-real-trial.log 2>&1`: exit 0. The actual Go test command is recorded in [the receipt](worker-receipt.json). Package PASS 361.092s (JSON event elapsed 361.095s). 223 unique upstream combinations matched 131446 bytes; 196 compiler/stage1 files matched 12711933 bytes; owned witnesses matched 6788 bytes on Go, Node and sanitized native. All six owned rule mutants and the registration subscription mutant were caught on both backends. [Script output](worker-run.txt), [complete Go events](worker-events.jsonl).
- `python3 -B /workspace/adamic/cloud/lint-wave-check.py --verify > /tmp/lint-wave-real-verify.txt 2>&1` in that same worker: exit 0 with the final checker implementation. [Output](worker-verify.txt). The default trial began before the live Markdown adaptation; the final implementation rereads and validates its actual events and receipt successfully.
- `python3 -B cloud/reports/lint-wave-check/origin-probe.py > /tmp/lint-wave-real-origin-negative.txt 2>&1`: exit 0 by asserting two expected rejections across 263 fetched live heads. no-continue is already in scanner/main and other production modules; consistent-this is reserved by claims/wave1-04.md on codex/lint-wave1-04 with no implementation required. [Literal output](origin-negative.txt).

## Each check proven to fail

The CLI tests create independent disposable Git repos and bare origins, commit each candidate and invoke the production script as a subprocess. Exit 1 and the named failed check are asserted. Their fake Go executable supplies controlled JSON events to test the evidence reader; those synthetic events are not semantic parity evidence.

| Check | Failing controls observed |
| --- | --- |
| Claim | Missing committed claim, wrong JSON/Markdown owner, duplicated JSON/Markdown assignment, unreadable foreign Markdown assignments |
| Origin ownership | Another head's claim despite a main-only fetch configuration; another slug's same public name; legacy TS or .a port; unimplemented Markdown reservation; JSON selection with no implementation; BATCH reservation; fetch failure; a new claim appearing after receipt creation |
| Directory registration | Shared dispatcher, Go oracle, lint harness or generator edit; rename of dispatcher into an owned directory; unclaimed directory; committed generated registry; new shared ordinal; claimed rule missing descriptor |
| Parity ran | Missing TestRulesAgree; zero upstream cases; missing three-way witness comparison; skipped compiler corpus; nonzero test process even with printed PASS events |
| Mutants ran and were caught | Missing owned mutant.json; absent TestMutants suite; absent claimed-rule subtest; no caught native comparison |
| Evidence freshness | No receipt, changed commit, changed log hash, dirty rule file, unapplied GOFLAGS overlay |

Positive parser/coordination controls: completed run then --verify; an already published own worker head; rule names occurring only in a witness; removal/pruning of a remote reservation. Other positive controls: existing Markdown discovery with explicit skips and appended reports; --claim selection; owned evidence JSON/report; foreign evidence JSON exclusion; "not skipped" remains reserved. All passed. A race after the final origin check remains possible because fetch/check and push are separate operations. This is a cooperating-worker guard, not adversarial attestation.

## Real trial and scope

A disposable worker at `/workspace/scratch/lint-wave-worker` uses the accepted directory-registration tip `48ecd9302bf3954a4ddbbd14c28ba09148c1c1a8`. Its bare origin has that tip as main. Trial commit `4683a33` adds exactly six no-continue-owned files and one claim, with no shared edits. Cohere and nested TypeScript are actual clean checkouts reusing local object stores, not an uncommitted dependency symlink. Compiler source corpus is Microsoft/TypeScript v6.0.3 `050880ce59e30b356b686bd3144efe24f875ebc8`.

The real trial is deliberately isolated: no-continue already exists on the real GitHub origin's scanner branch. The separate final live-origin probe rejected it there, and independently rejected a real Markdown reservation. The worker trial demonstrates successful semantic execution and receipt validation, not authorization to duplicate no-continue on the live wave.

Remote main advanced during this unit to `ef3d907ecdc4c771b016f7d9c52372def057a340`, integrating monolithic scanner and other compiler/stage1 work. This branch retains the requested starting base. That fetched main still has no directory-registry/rules tree. Main has not yet integrated the directory-registration prerequisite, so this checker does not import that unit's implementation. Workers include its exact published tip until integration. Legacy batch runners need that migration before they can satisfy the new directory-only contract. No speed, new rule coverage or whole-repository gate claim is made. This unit uses the ordinary-worker scoped fallback: cloud package, a real registered lint worker's parity/mutants, repository vet and a filtered external oracle. The preceding complete uncached gate took about 29 minutes on this machine; this Python/Git unit does not repeat unrelated compiler packages. Integration still owns the complete uncached gate.

## Reproduce the real worker trial

Create a scratch worktree at `48ecd93` with a fresh local bare origin whose main is that same commit. Initialize the pinned submodules as real checkouts. Create branch codex/lint-wave-trial, apply [the six-file rule patch](trial-rule.patch), and commit a seventh file at stage1/cohere/lint/claims/codex/lint-wave-trial.json containing `{"version":1,"branch":"codex/lint-wave-trial","rules":["no-continue"]}`. Run the default checker with the pinned compiler environment, then --verify. The saved candidate commit is `4683a33b719ba21196adecf295c7fa79476f6683`; trial refs were not pushed to GitHub. The isolated origin is essential to this positive control: the real origin correctly rejects duplicate no-continue work.

Actual compiling behavior mutants: debugger fix suppressed; empty function body reported; equality suggestion applied as fix; var declaration suppressed; duplicate case suppressed; no-continue diagnostic ID changed. Each completed on Node and sanitized native, then differed from upstream Go. The valid-but-wrong descriptor subscription also compiled/executed and differed on both. No compile error is credited as a semantic mutant kill.
