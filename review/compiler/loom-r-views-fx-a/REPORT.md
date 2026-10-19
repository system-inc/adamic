Rebased the unchecked literal-key and unsupported-member view-read refusals for #1fk58py, items 132 and 142.
Commits: source b198bbdec5707db4ada3baa34b09d5912ac5547a; exact base dbdf977f5b8b0979d414ebe6f20e7aadd312e309; delivery SHA is in the final response.
Commands and outputs: own fixtures, focused lower/oracle and flow checks, all 52 reader packages, one Linux counts regeneration, reader guard, vet and both backend CLI replays passed.
Mutants: both original reverted-guard mutants caught by their intended located-refusal assertions, with got <nil>, not compiler warnings or timeouts.
Not covered: other views units, full package/repository gates, ahra CLI, PR or Loom submission; no runtime C changes.

The normal fetch did not include the historical object, so it was fetched explicitly by its full SHA. The branch was created from exactly dbdf977f5b, not a newer main. Applied only the source commit with `git cherry-pick --no-commit b198bbdec5707db4ada3baa34b09d5912ac5547a`. No main-to-historical-tip diff was used.

Conflicts and resolutions: none. The cherry-pick completed cleanly, with no conflicted hunk and no manual source resolution. Every imported production, test and fixture file under internal/lower is byte-for-byte identical to b198bbde; source-equivalence.json records each pair of blob IDs. No newer main source was reverted. Three historical evidence files (branch-commits.txt, key.patch, unsupported.patch) were gzip-compressed, preserving their bytes, to eliminate whitespace warnings from imported review records. Their old report remains historical evidence; this report describes the current validation.

Meaning preserved: item 132 refuses matching object reads with string-literal or literal-typed keys and names the source path and `use view.<field>` fix. Conservative assumption: keep the source change's whole-program field-name alias policy, including matching unrelated indexed reads, until checked member syntaxes are unified. Item 142 registers all names on an unsupported descriptor, including receiver-specific keys, so a demanded read reaches the existing located refusal and implementation/proof fix. Unread unsupported views remain accepted. No broader view or producer change is included.

Counts: regenerated exactly once on Linux with the house clang 20.1.8, using the unanchored TestCountsAreRecorded selector and -update-counts as the counts header and generator specify. PASS in 191.524 s. Changed rows: none. The complete file's blob before and after is bdab0f5b96a2080ca1220efe4d18de2a693fd62d. counts-result.json records that comparison. No counts were hand-merged. The source change adds no oracle rows; its eleven refused probes live under internal/lower/testdata/loom_views_a, outside both the top-level oracle glob and flow's corpus globs. No refused top-level oracle fixture was added or left behind, and no fixture relocation was needed.

Runtime C: the source commit's own diff contains no internal/native/runtime file; this rebased unit's diff versus dbdf977f5b also contains none. Runtime lines changed or deviating from the source commit's own runtime changes: none. No new IR-held map, compiler consumer package, or runtime feature set is added.

Validation commands are recorded exactly in commands.txt. Output went to files, not pipelines; logs and package metadata are compressed alongside this report.

| Check | Result |
|---|---|
| Focused own fixtures and view regressions | internal/lower PASS 2.217 s; internal/oracle PASS 38.039 s |
| Current inverse-import closure, ordinary/test/external-test imports | all 52 packages passed focused selection; unmatched packages compiled only |
| Flow corpus selection | PASS 17.862 s; 20 selected programs, corpus coverage/shared setup, remainder covering 15 programs |
| Repeated own fixture leaves in reader pass | internal/lower PASS 2.653 s |
| Linux full counts regeneration | PASS 191.524 s; zero rows changed |
| internal/ir TestCallTargetReaders, -timeout 90s | PASS 6.47 s |
| go vet ./internal/lower | exit 0, empty output |
| gofmt on all three imported Go files | empty output |
| git diff --cached --check | clean after lossless compression of historical evidence |

The focused selection includes the checked-view, interface-cast, required-view, narrowed-field, view-field and default-tagged tests, in addition to every TestLoomView leaf. The reader selection also runs relevant element-access/interface/view flow units and the flow corpus coverage and remainder tests. No whole package suite or full gate ran. No other package's tests were edited to make room.

The explicit CLI replay runs Node first on each of the eleven probes and pins the source output, then invokes both `adamic c` and `adamic js`, requiring the intended diagnostic, fix and .a source location. All eleven intentionally refuse compilation in both backends. The two positive controls produce `14\ntrue\n` and `created\n` respectively, byte-for-byte the same in source Node, generated JavaScript, release native, and native with ASan/UBSan/LeakSanitizer. All positive executions exit zero with empty stderr. entrypoints.json holds the commands, outputs, exits and durations. The first standalone JavaScript replay omitted oracle/node.mjs and failed to resolve the adamic runtime import; the harness was corrected to use the house loader and replayed successfully. The failed harness attempt is retained separately; production code was never changed for it.

| Original mutant | Intended catcher | Observation |
|---|---|---|
| key: clear fields before the view element-read refusal walk | TestLoomViewKeyRead | exit 1, want located refusal and use view.count, got <nil>; leaf 0.10 s |
| unsupported: disable registration of names from an unsupported descriptor | TestLoomViewUnsupportedRead | exit 1, want checked-view read refusal and contract fix, got <nil>; leaf 0.12 s |

The original runner's two mutations were used unchanged in meaning, with evidence output redirected to this branch's review directory. Both sources were restored in finally, and the restored code then passed reader checks and all CLI replays. No warning, bad-C build, or timeout counted as a mutant kill.

Every added leaf is top-level and parallel. Setup is included in these observed leaf times on the four-CPU quota; no leaf exceeded 60 seconds and no over-budget leaf was rerun.

| Leaf | Focused seconds | Reader-pass seconds |
|---|---:|---:|
| TestLoomViewKeyFix | 0.56 | 1.01 |
| TestLoomViewKeyRead | 0.33 | 0.96 |
| TestLoomViewOriginalNativeP21 | 0.37 | 0.93 |
| TestLoomViewOriginalNativeP37 | 0.44 | 0.58 |
| TestLoomViewOriginalViewsP01 | 0.32 | 0.82 |
| TestLoomViewOriginalViewsP05 | 0.44 | 0.86 |
| TestLoomViewOriginalViewsP06 | 0.46 | 0.58 |
| TestLoomViewOriginalViewsP19 | 0.39 | 0.62 |
| TestLoomViewOriginalViewsP72 | 0.36 | 0.98 |
| TestLoomViewUnionKeyRead | 0.41 | 0.81 |
| TestLoomViewUnsupportedCallableRead | 0.33 | 0.91 |
| TestLoomViewUnsupportedRead | 0.39 | 0.81 |
| TestLoomViewUnsupportedUnreadControl | 0.51 | 0.90 |

Toolchain setup succeeded after exporting GOPROXY=https://proxy.golang.org|direct. The printed environment file /workspace/adamic-tools/env.sh was sourced for every Go/build/test shell. nproc=5; cgroup cpu.max=400000 100000 (four CPUs). Setup timing and version lines:

```
submodule cache: off
setup: submodules ready (0.077s)
go version go1.27.1 linux/amd64
setup: go ready (0.080s)
v24.19.0
setup: node ready (0.086s)
setup: markdown dependencies skipped (validated lock and installed bytes); step-duration=0.012s
setup: markdown dependencies ready (0.214s)
clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0.414s)
setup: shared cache ready (1.283s)
setup: go build ready (34.462s)
setup: test binaries deferred (use --warm-tests) (34.547s)
setup: build cache warm (34.548s)
setup: build-flags commit=dbdf977f5b8b0979d414ebe6f20e7aadd312e309 nproc=5 cpu.max=400000 100000 go=go version go1.27.1 linux/amd64 clang=clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261) node=v24.19.0 cached=yes warm-tests=false gocacheprog=on load-before=0.00 0.00 0.00 1/138 697 load-after=1.64 0.40 0.13 1/143 1621
setup: done on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB (34.575s)
setup: source /workspace/adamic-tools/env.sh
setup: logs /tmp/adamic-gate/setup.RCnUwo
```

The inherited root filesystem was nearly full while compiling reader test binaries. 1,827,362,610 bytes of old immutable Go cache blobs were relocated to temporary storage with atomic symlinks preserving both bytes and cache paths; cache-relocation.json records them. The reader guard used ADAMIC_BUILD_CACHE_DIR=/tmp/r-views-reader-products so its export-copy product did not exhaust root. No source, test, or product-key logic was changed for this storage workaround. All required validation completed successfully.

Assumptions and exclusions: lower/testdata probes are not top-level oracle fixtures and are not consumed by flow's top-level oracle glob; retain their original paths. Keep the original narrow compile-time refusals rather than expanding checked-read support. Leave the other views items and units for their own wave. Do not invoke ahra, open a PR, submit to Loom, push main, rewrite history, or force push.
