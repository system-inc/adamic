Measured the P0 refusals verification revision against its merge base with fetched origin/main, on a four-CPU quota.
Revisions: refusals `849159957f795b8e3024b427488593bd115d6afb`; main-base `a7448d73cd17f16362b6cbc5c5c111080da64e43`.
Command: `go test ./stage1/cohere/yaml -count=1 -timeout 300s -json`; both warm runs timed out without a named failing test.
Existing YAML mutation checks completed without failures before the timeout; no new mutant or code change was introduced.
Coverage limit: 51 parallel tests remained paused, `TestWidthsMatchGo` was active, and seven optional npm YAML oracle tests skipped.

**No refusal-induced YAML failure was observed.** The warm runs have identical sets of 49 distinct passing test names (parents and subtests counted together), seven skipped names, and no named failures. Both stopped in `TestWidthsMatchGo`. This is not proof that the whole package passes: neither run completed. The only named failure was the first cold refusals run's `TestFileDriver_Setup`, caused by Go dependency-download messages on stderr, not an Adamic refusal. That setup subsequently passed on the same revision.

**Both revisions exceed 90 s on this four-CPU instance, even warm.** Refusals warm took 302.59 s externally and main-base warm took 303.05 s. These are timeout-limited durations, not successful completion times. They exclude the separately measured Go test-binary compilation. This supports a package-duration explanation for the gate's 90.13 s kill; it does not establish the exact cause of that separate gate execution.

| Revision | Cache state | External wall, s | JSON package elapsed, s | Exit | Result |
|---|---|---:|---:|---:|---|
| refusals | cold attempt | 34.06 | 31.411 | 1 | `TestFileDriver_Setup` failed on dependency-download stderr |
| refusals | warm setup caches | 302.59 | 300.657 | 1 | `panic: test timed out after 5m0s`; `TestWidthsMatchGo` active; no named failure |
| main-base | cold | 362.50 | 360.018 | 1 | Go outer watchdog: `Test killed with quit: ran too long (6m0s)`; `TestUnistMutants/document_end_marker_lost` active; no named failure |
| main-base | warm setup caches | 303.05 | 300.617 | 1 | `panic: test timed out after 5m0s`; `TestWidthsMatchGo` active; no named failure |

The fetched main tip was `bdb89962b178f618e2e6d34e9a7ea5c395089faf`; `git merge-base 84915995 origin/main` selected the main-base above. The YAML package sources are identical between these revisions. Both isolated checkouts used cohere `7945d102a6c18dd36adf9114a758ce646e8b2359` and TypeScript `d92d9bfee114c80be2c375d72edae966176e3a4f`, and ran `npm ci --ignore-scripts` in `stage3/api` before testing.

The instance reports `nproc = 5`, but its cgroup `cpu.max` is `400000 100000`, a four-CPU quota. Every test command used `GOMAXPROCS=4`; revisions ran sequentially. Go 1.27.1, Node 24.19.0, and clang 20 came from `/workspace/adamic-tools/env.sh`, with `GOPROXY=https://proxy.golang.org|direct`. Each invocation had an additional 420 s process-group cap; the runner had a 2400 s outer cap. Neither additional cap fired.

Cold means fresh, separate per-revision `ADAMIC_BUILD_CACHE_DIR` and native `XDG_CACHE_HOME` caches. Go SDK compilation was primed separately using `go test ./stage1/cohere/yaml -run '^$' -count=1 -timeout 300s -json`, because it is a separate cost. Warm means reuse of those product/runtime caches with `-count=1` still forcing test execution. Individual tests may still perform native builds; this is not a claim that every generated test binary was cached.

| Preparation phase | Refusals, s | Main-base, s |
|---|---:|---:|
| stage3/api npm provisioning | 0.67 | 0.82 |
| Go test-binary priming | 154.40 | 160.32 |
| Separate file-driver setup priming after cold failure | 52.55 | not needed |

The cold refusals setup failure printed:

```text
file_driver_shards_test.go:317: go stderr: go: downloading github.com/dlclark/regexp2/v2 v2.5.2
    go: downloading github.com/dlclark/regexp2 v1.11.5
--- FAIL: TestFileDriver_Setup (31.39s)
```

After these downloads, setup was primed with `-run '^TestFileDriver_Setup$'` and `ADAMIC_FILE_DRIVER_SETUP_CHILD=1`, publishing `ADAMIC_FILE_DRIVER_STATE` into scratch storage. This bypassed only TestMain's separate setup-child 90 s cap, without editing source, and passed in 52.55 s. The subsequent full warm command used the normal TestMain again. Refusals setup cache misses took 1.01 s for the Go oracle, 24.13 s for lowering, and 25.06 s for native; its warm setup hit all three caches. Main-base cold misses took 17.38 s, 24.14 s, and 25.31 s respectively; its warm setup also hit all three caches.

`-timeout 300s` does not remove TestMain's independent 90 s setup-child cap. The cold main run spent approximately 67 s in setup before `m.Run`, then the outer Go watchdog killed it at 360 s before all tests completed. This differs from the warm runs' explicit five-minute test timeout.

Seven optional-oracle tests skipped because `ADAMIC_YAML_LIBRARY` was unset: `TestComposeMatchGo`, `TestCSTMatchesGo`, `TestFormatterMatchesGo`, `TestLexerMatchesGo`, `TestPropsMatchGo`, `TestSchemaMatchesGo`, and `TestUnistMatchesGo`. No clean cold refusals package duration was obtained after the dependency-download artifact; its original failure is retained. No compiler or test change was made to repair any result.

Evidence: [inputs and CPU quota](yaml-evidence/inputs.json), [commands and external timings](yaml-evidence/status.json), [test states including every paused test](yaml-evidence/summary.json), and [reproducer](yaml-evidence/run.py). Raw Go JSON streams are `yaml-evidence/{refusals,main-base}-{cold,warm}.jsonl`; setup/compiler priming streams, stderr logs, build-cache logs, and the runner log are alongside them.
