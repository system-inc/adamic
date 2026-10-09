# Cold fixture oracle hook, 7r4k16h

Stopped under the task's toolchain-cost exception. The requested cold command exceeds 60s, but TestProduct_FixtureOracleHook itself does not. Almost all elapsed time precedes test execution: Go compiles, vets and links the stage3/fixtures test executable and its dependency graph. Splitting this product recipe would leave that prerequisite cost intact. No recipe, cache key, test, hook source, compiler flag or production code was changed, and no work was deferred.

Base: freshly fetched origin/main, 92d196011e78208b45b3768dcf2c8c88e7ec132d. Branch: devtools/grain-oracle-hook. The after column is a second fully cold measurement with the unchanged recipe, not a claimed optimization.

Both runs had separate initially empty GOCACHE and ADAMIC_BUILD_CACHE_DIR directories. Each ran go clean -cache against its selected private Go cache before timing. Both product logs say miss with the same key, ae48d737bd94. Python's build-hook.py receives --store inside the new product directory, so the older stage3 Python store cannot provide a warm hook.

Instrument: Bash time for command wall/user/system time; Go -json for test-process and product-test timing; Go -x for compile/vet/link trace; buildcache's ADAMIC_BUILD_LOG for the product miss. The Go startup row is total command wall time minus Go's package-test elapsed time: compile/vet/link and CLI overhead, rather than pure compiler CPU time. Product input discovery and other test overhead are test elapsed minus the buildcache miss interval.

Build flags: Linux amd64; Go 1.27.1; native clang 20.1.8; Node v24.19.0; nproc=5 outside affinity, taskset -c 0-3 nproc=4; cgroup cpu.max=400000 100000; GOMAXPROCS=4. CPUs 0-3 are available and the affinity is inherited by child builds. The unchanged hook recipe builds with -buildvcs=false -c ./internal/oracle. No Adamic/native runtime compilation is hidden outside its recipe.

| Phase, seconds | Cold before | Cold repeat after |
| --- | ---: | ---: |
| Entire requested go test command | 169.518 | 164.848 |
| Go compile/vet/link startup and CLI overhead | 159.813 | 155.542 |
| Package test process | 9.705 | 9.306 |
| TestProduct_FixtureOracleHook | 9.69 | 9.29 |
| buildcache product miss | 8.60 | 8.57 |
| Product input discovery and other test overhead | 1.09 | 0.72 |

Both commands exited 0. The baseline outer Go trace records 253 compile, 247 vet and one link invocation before the fixtures test executable starts. It includes standard-library and cohere compiler/lint dependencies imported by this package. The product's additional oracle preparation is inside the 8.60s/8.57s miss intervals, with the unchanged Python helper's build logs retained in each product directory.

Exact build census lines:

```text
build stage3-fixture-oracle-hook ae48d737bd94 miss 8.60
build stage3-fixture-oracle-hook ae48d737bd94 miss 8.57
```

Before and after SHA-256, independently recomputed from the published executable bytes:

```text
before fcef7f2756136acf1572ce42730591379242e9304d741345af7d6693bb3f3735
after  fcef7f2756136acf1572ce42730591379242e9304d741345af7d6693bb3f3735
```

Both hooks are 38,078,105 bytes. The original product also checks the published bytes against this content-addressed filename. Nothing new is cached and no new validation check is introduced, so no new cache-key mutant or top-level test is needed for this report-only branch.

Commands, independently for before and after, with source /workspace/adamic-tools/env.sh:

```sh
export GOCACHE=/tmp/grain-hook-<before-or-after>-go
export ADAMIC_BUILD_CACHE_DIR=/tmp/grain-hook-<before-or-after>-products
export ADAMIC_BUILD_LOG=/tmp/grain-hook-<before-or-after>-build.log
export GOMAXPROCS=4
timeout 600 taskset -c 0-3 go clean -cache
TIMEFORMAT='wall_seconds=%R user_seconds=%U system_seconds=%S'
time timeout 600 taskset -c 0-3 go test -count=1 -timeout 120s -run '^TestProduct_FixtureOracleHook$' -json -x ./stage3/fixtures
```

Test output went directly to /tmp/grain-hook-before.log and /tmp/grain-hook-after.log. Bash timing is in the corresponding -time.txt files, product census in -build.log, and the artifacts in -products/. The test's 120s timeout begins after Go's compilation; the external 600s ceiling covers compilation too. /usr/bin/time was absent, so an initial attempted invocation exited 127 without starting Go; Bash's timer was used for both actual measurements.

Unresolved: the cold test-package build still exceeds 60s, and the pool's 90s preparation kill is not fixed. This experiment does not measure the pool's lock contention or remote-store fetches. A package-build/toolchain unit must address the 155-160s startup cost before the full cold command can meet 60s. No full fixture suite or repository gate was run: only the two required product preparations and artifact hash comparisons.
