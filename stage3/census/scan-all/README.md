# scan-all: refusal-scan continuation census

One invocation is **one root import-closure run**. In the gate, only the anchor test runs; the controls run through verify.sh. This records refusal-scan stops only, plus the first declaration, body or tail stop it reaches, and is blind past that first stop. **A zero for a body-phase or tail rule is never a pass.** It is not production's full wall list. Preflight and registry validators can themselves return their first error; this does not enumerate all errors inside a validator.

This tooling copies main's production lowering into a build overlay. It continues only the refusal scan, then uses the unchanged production declaration/body/tail abort behavior. There are no archived Language or Memory rule files, no snapshots, placeholder bodies, independent-unit union or compiler acceptance path. Both public lowering APIs return nil IR and a measurement-only error. The measurement API returns observations and counts only. Each process accepts exactly one root and runs once; its globals are not a concurrent API.

`make_overlay.py` writes only outside the checkout. `--check` reads the sources in Python and checks the anchors the overlay depends on, with no whole-file hashes and no Go subprocess. It checks all eight `LOWER_REWRITES` anchors exactly once, requires at least one pipeline error-boundary regex match, and checks each matched site exactly once. In `refusals.go`, it checks the original `refuse` signature, its pragma loop, both visitor signatures, the registry rule call and each matched module preflight. Checks are scoped to `refuse`, so the separate `refuseNode` walker is not mistaken for a duplicate. An unrelated comment or other edit outside these anchors does not invalidate the check.

Some anchors exist only **after refusalrewrite runs**: the `latentRefuse` clone, the named `latentStop` visitor signatures, each visitor's `latentRecord` defer and `latentFullEnabled()` exclusion block, and the clone's label-insertion sites. Generation in `verify.sh` runs the Go AST rewriter outside test units and refuses on missing or duplicate generated anchors, naming them. The gate's Python-only check does not run that rewriter or claim to validate its generated output.

`TestScanAllMainAnchorsMatchExactly` checks main's current source anchors. `verify.sh` also runs that unchanged named test on isolated scratch copies: removed and duplicated lower/refusal anchors must turn it red and name the anchor; an empty pipeline-site match must turn it red; an unrelated comment added to `lower.go` must keep it green. A trap on Go subprocesses proves the check itself never runs Go. No checkout source is mutated for these proofs. Preparation failures are reported as a boundary without fabricating a declaration/body/tail observation. The underlying first pipeline stop is captured before production's deferred host-write diagnosis can replace it; it is counted once. The report does not treat the substituted diagnosis as another completed pass.

Every observation carries a `Kind`: `scan stop` or `first declaration/body/tail stop`, plus rule, site and words. Registry labels name the actual rule. Non-registry labels distinguish checking pragmas, suppression directives, the argument-contract walker and each module preflight. Production retries can repeat observations; event counts retain those repetitions. Events and distinct `(rule, site, words)` triples are different measures. The summary runner groups files from the sites, never from a claimed root count. `Completed` describes measurement traversal only, not compilation acceptance.

## Reproduction

Use a detached worktree of main, initialize recursive submodules, and install `stage3/api` with `npm ci`. Set `ADAMIC_TYPESCRIPT_SOURCE` to TypeScript v6.0.3, commit `050880ce59e30b356b686bd3144efe24f875ebc8`. Put `GOTMPDIR`, `TMPDIR`, `GOCACHE`, `ADAMIC_BUILD_CACHE_DIR` and `SCAN_ALL_SCRATCH` in scratch outside the checkout. Use the repository's Go 1.27.1 toolchain.

```sh
bash stage3/census/scan-all/verify.sh
stage3/apply.sh --build-product "$SCAN_ALL_SCRATCH/adapted"
python3 stage3/census/scan-all/report.py "$SCAN_ALL_SCRATCH/scan-all" \
  "$SCAN_ALL_SCRATCH/adapted/src/compiler/sys.ts" "$SCAN_ALL_SCRATCH/sys.md"
```

The verifier builds two binaries **before** test units: the instrumented tool, and a plain production `load.Load` / `lower.Lower` probe with no lowering overlay. Go templates are `.go.txt`; only real tests use `.go`. Each control removes its leading checking pragma and proves the target with the plain production probe, or proves its clean twin lowers. It then reinstates the earlier pragma and requires the scan tool to observe both the pragma and the later target, or complete the clean twin. Both public no-IR guards run on every scan invocation.

The package's one Go test, the anchor check, calls `t.Parallel()`. The controls aren't Go tests: they need the two prebuilt binaries named by `SCAN_ALL_BINARY` and `SCAN_ALL_PRODUCTION`, and a test that skipped without them would be a skip the gate's skip census refuses. `verify.sh` runs each control through `controls.py`, bounded at 55 s and with no build commands, then each mutant. Each `SCAN_ALL_MUTANT_FIRST_ERROR=1` run restores the original production refusal scan; its control must fail with `continuation-target` or `continuation-completion`, because the later target disappears or the clean traversal no longer completes. A failure for any other reason is not a killed mutant, and verify.sh says so.

## Controls

The control rule locations below refer to main `89614cc0f9d6aa7b7fe96b56c5c6c1c3eace9825`. Every fixture begins with the earlier `@ts-check` scan stop (`internal/lower/refusals.go:53`). Each pair changes the relevant input; names describe the later production stop, not an all-stop census result.

- `count-stop` and `count-clean`: argument-count observation through a bound function; `checkBoundArguments`, `internal/lower/function_bind.go:361`, diagnosis at line 377. The twin returns its parameter without reading `arguments`.
- `rest-stop` and `rest-clean`: direct rest-parameter bind; the production body-lowering check in `functionBind`, `internal/lower/function_bind.go:54`. Main rejects this before `checkBoundArguments`, so this pair claims a first body stop, not a tail result. The twin has no rest parameter.
- `record-stop` and `record-clean`: record written directly into a graph slot; `findCycles` / `graphTypes` / `noteRecordSlot`, `internal/lower/graph_records.go:16`. The twin uses arrays of nodes rather than a record.
- `cycle-stop` and `cycle-clean`: record held inside an array written into a cyclic graph holder; the same production record rule, whose `holdsRecordSeen` follows array elements (`internal/lower/graph_records.go:54`). The twin uses node arrays. These share the record validator but exercise different written value types.

No `declaresBeyond` control is claimed: that rule is absent from this main. The eight controls target rules main actually has. Each has its own first-error mutant run, which must turn that control red; clean twins prove continuation by completion, rather than by inventing a clean rule event.

## Isolation and evidence

Nothing under `internal/lower` is edited. The only generated replacements are copies of `lower.go`, `refusals.go`, the measurement helper and tool entry point. No gate consumes the report. `verify.sh` is an opt-in proof runner; these report streams are not compiler products or acceptance receipts. `examples/` contains summary tables only, with warnings repeated before every table. Raw streams stay outside the repository. Resource measurements cover loading/checking and lowering in the child, excluding adaptation and builds; peak RSS is Linux's child maximum in KiB.
