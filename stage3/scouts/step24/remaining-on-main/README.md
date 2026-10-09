Built: blocked-rerun evidence, all 67 historical identities, both main-only split controls, and three Node-held group fixtures.
Pins: main 031a1259bc7973934792dc6cb1bd4074fc2204b9; historical scout f053ef44a3b0ee7c52fcc3e2e7e3caa035918228; available fallback 8f32e51e8fc41b8f1177453213ca5453ce764486.
Results: main-only split 0 and 1 each exit 1 with 319 diagnostics; all historical 67 present; first builder.ts:1246:69 TS2345.
Mutants: three source mutants change Node output; four evidence mutants rejected by audit.py.
Uncovered: requested combined rerun and definitive per-site ownership blocked by missing requested ref and 12 merge conflicts; no native binary or full gate.

The requested `origin/compiler/stricter-options-next` does not exist on the remote. The available `origin/codex/stricter-options-next` is exactly the pin used by the historical 67-stop scout. An optional clarification was sent; after allowing time for a reply, the fallback was explicitly announced. Its scratch merge into current main fails. [conflicts.json](conflicts.json) lists every conflicted path; [merge.log](evidence/merge.log) retains Git's output. No conflict resolutions or compiler changes are published. The scratch worktree `/tmp/step24-remaining-compiler` is never pushed.

Resolving conflicts in IR, JavaScript emission, lowering and native emission requires compiler integration decisions beyond this evidence unit. Taking one side wholesale could silently discard main's representation or admission changes. Therefore the requested combined result is **unmeasured**, not zero stops and not the main-only count.

The main-only control builds tsc's original `src/compiler/parser.ts` entry; it does not build the staged .a diagnostics driver or produce a runnable native parser. Its 319 diagnostics cover the imported compiler closure. [comparison.json](comparison.json) records every identity, both splits, and all historical-presence checks. [TABLE.md](TABLE.md) groups the historical 67 as 51 optional-field, nine result-relation and seven argument stops. Existing adaptation routing is provisional, explicitly distinct from proving an adaptation's eligibility. None of these tables establishes which errors remain on the blocked combination.

The external adapted tree `/tmp/step32-adapted` is reused by content, not by directory name: the adaptation pipeline is unchanged between historical main 45487a80 and 031a1259, and [source-check.json](source-check.json) verifies all 23 diagnosed source hashes against the historical scout. TypeScript is pinned to 6.0.3 commit 050880ce59e30b356b686bd3144efe24f875ebc8. No upstream source is committed here.

Setup used `export GOPROXY='https://proxy.golang.org|direct'`, `bash cloud/setup.sh`, then `source /workspace/adamic-tools/env.sh`. Timing: Go 0.023s, Node 0.027s, submodules 0.071s, markdown 0.080s, clang 0.171s, build 10.303s, deferred tests 10.453s, cache 10.455s, total 10.484s. nproc=5, CPU quota=4, memory=17.6 GB. [setup.log](evidence/setup.log) retains every line.

Commands run with output redirected to files:

```
git worktree add --detach /tmp/step24-remaining-compiler origin/main
git -C /tmp/step24-remaining-compiler merge --no-edit origin/codex/stricter-options-next
# Merge exits 1; no combined compiler built.
go build -o /tmp/step24-remaining-main-adamic ./cmd/adamic
ADAMIC_NATIVE_SPLIT=0 /tmp/step24-remaining-main-adamic build /tmp/step32-adapted/src/compiler/parser.ts -o /tmp/step24-remaining-parser-0
ADAMIC_NATIVE_SPLIT=1 /tmp/step24-remaining-main-adamic build /tmp/step32-adapted/src/compiler/parser.ts -o /tmp/step24-remaining-parser-1
python3 stage3/scouts/step24/remaining-on-main/check.py
python3 stage3/scouts/step24/remaining-on-main/audit.py
```

`check.py` executes original .a source through repository `oracle/node.mjs` on Node. Fixtures are deliberately checker-rejected, with verified in-place a-check headers. They witness presence of an own undefined property, an indexed result, and an indexed argument. Mutants delete the own property (`present` -> `absent`) or change index 0 to 1 (`ok` -> `undefined`). [fixtures.json](fixtures.json) retains times, output and compiler rejections. Each fixture unit completes under 30 seconds, enforced in the runner. These are group semantics witnesses, not proofs that all 67 compiler obligations can be erased. An initial fixture-runner repository-path error was corrected before the successful run.

Builds use Go's hash-keyed build cache warmed by setup; no package-wide tests or full gate ran. The next unit needs an integrated stricter-options ref compatible with main, then must rerun the combined measurement and resolve each adaptation's owner/type prerequisites.
