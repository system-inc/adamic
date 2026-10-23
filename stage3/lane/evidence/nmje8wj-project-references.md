# Project-reference shard split: task #nmje8wj

Observed: shard 00 is ordinary among its siblings; generated contention slows all six samples, with different ratios. Inference: Eve's much larger slowdown is consistent with box contention, not evidence of an unusually large shard 00. This worker does not reproduce Eve's hardware or load 53.5.

The project-reference partition changes from 3 to 24 groups. Other partitions keep their group counts. The total becomes 398 units. Duplicate full titles and paired noEmit families stay together. Every new group is wholly owned by one old group (24 is a multiple of 3).

**Acceptance limitation:** the requested quiet wall bound is not met. The old sibling maximum is 3.658111 seconds, so its third is 1.219370 seconds. New quiet walls are 1.706266 to 2.030309 seconds; loaded walls are 4.004346 to 6.302638 seconds. A scratch 201-group plan selects one distinct title per shard; its first three single-execution shards still take 1.499111, 1.444836 and 1.546044 seconds. These observations support a fixed harness/startup floor; they are not a proof of timings on every box. I kept the measured 24-group split and the unchanged runtime limits rather than claim the bound passed.

## Measurements

Each value is shard.run_unit's report.seconds, from its entry through verdict publication, including input validation, view preparation, upstream execution and baseline checks. Artifact production is outside the timer. Calls use the unchanged shard.py implementation, with fresh output directories, Node compile cache disabled, no result reuse, GOMAXPROCS=8, and affinity CPUs 0-3. Quiet runs have no generated CPU workers; loaded runs have eight Python busy loops on that same affinity. nproc=5, cgroup cpu.max=400000 100000 (four CPUs), not an eight-CPU allocation. Setup's submodule fetch was pending during the measurements; no setup compiler build ran. OS page caches were retained.

The same prepared adapted tree and stock API parser were used before and after. The legacy artifact includes node_modules, so the supplied product-layout scratch view omits only that entry; shard.py supplies stage3/typescript/node_modules as the normal Go product path does.

| Original shard | Identities | Quiet seconds | Loaded seconds | Slowdown |
|---|---:|---:|---:|---:|
| stage3-lane-file-000 | 6 | 4.519135 | 9.218255 | 2.040x |
| stage3-lane-file-001 | 1 | 1.908527 | 5.818335 | 3.049x |
| stage3-lane-file-002 | 12 | 1.816800 | 4.445934 | 2.447x |
| stage3-lane-project-reference-cases-00 | 87 | 3.515540 | 10.403165 | 2.959x |
| stage3-lane-project-reference-cases-01 | 93 | 3.416642 | 11.710684 | 3.428x |
| stage3-lane-project-reference-cases-02 | 93 | 3.658111 | 12.320365 | 3.368x |

All twelve original runs passed. Shard 00 is smaller than 02 quiet and loaded. File 000 is slower than every project-reference shard quiet. These data show broad contention sensitivity, not identical slowdown ratios.

| New project-reference shard | Identities | Quiet seconds | Loaded seconds |
|---|---:|---:|---:|
| stage3-lane-project-reference-cases-00 | 11 | 1.959568 | 4.146691 |
| stage3-lane-project-reference-cases-01 | 13 | 1.958476 | 6.302638 |
| stage3-lane-project-reference-cases-02 | 12 | 1.909004 | 5.570723 |
| stage3-lane-project-reference-cases-03 | 13 | 1.805385 | 5.669045 |
| stage3-lane-project-reference-cases-04 | 15 | 1.953895 | 4.790448 |
| stage3-lane-project-reference-cases-05 | 15 | 1.905299 | 4.234468 |
| stage3-lane-project-reference-cases-06 | 14 | 1.908632 | 4.104148 |
| stage3-lane-project-reference-cases-07 | 14 | 1.855262 | 4.112332 |
| stage3-lane-project-reference-cases-08 | 13 | 2.030309 | 4.005833 |
| stage3-lane-project-reference-cases-09 | 10 | 1.811530 | 4.004346 |
| stage3-lane-project-reference-cases-10 | 10 | 1.801904 | 4.113556 |
| stage3-lane-project-reference-cases-11 | 11 | 1.904588 | 4.913391 |
| stage3-lane-project-reference-cases-12 | 10 | 1.808191 | 5.502094 |
| stage3-lane-project-reference-cases-13 | 10 | 1.856572 | 5.301738 |
| stage3-lane-project-reference-cases-14 | 11 | 1.853550 | 5.597807 |
| stage3-lane-project-reference-cases-15 | 9 | 1.857181 | 5.551764 |
| stage3-lane-project-reference-cases-16 | 9 | 1.754066 | 5.391139 |
| stage3-lane-project-reference-cases-17 | 9 | 2.009752 | 5.593460 |
| stage3-lane-project-reference-cases-18 | 9 | 1.863213 | 5.340888 |
| stage3-lane-project-reference-cases-19 | 10 | 1.908523 | 5.410102 |
| stage3-lane-project-reference-cases-20 | 11 | 1.818529 | 6.003475 |
| stage3-lane-project-reference-cases-21 | 11 | 1.706266 | 5.812938 |
| stage3-lane-project-reference-cases-22 | 12 | 1.865228 | 5.816343 |
| stage3-lane-project-reference-cases-23 | 11 | 1.912856 | 5.590236 |

All 48 new measurements passed, each cohort 273 pass / 0 fail / 0 skip.

## Coverage and masked diff

The masked unit comparison loads main's original plan with git show and the regenerated plan, excludes project-reference units and removes only inputs_hash from each remaining unit. The resulting 374-unit JSON values are equal; runtime is a top-level plan field. The complete identity listings are equal after masking only unit. The regression test independently rebuilds the old three-group plan, checks unchanged partitions, each child's unique old owner, each old names_digest, and full-title/paired-family ownership.

```text
PASS masked unit diff: 374 non-project-reference units identical after masking inputs_hash; runtime is plan-level
PASS stage3-lane-project-reference-cases-00 87 identities 427e39a0abd5ec2a2e5093e3d2f3da0db3693883008a97699ab19915ee7766b4 children ['stage3-lane-project-reference-cases-00', 'stage3-lane-project-reference-cases-03', 'stage3-lane-project-reference-cases-06', 'stage3-lane-project-reference-cases-09', 'stage3-lane-project-reference-cases-12', 'stage3-lane-project-reference-cases-15', 'stage3-lane-project-reference-cases-18', 'stage3-lane-project-reference-cases-21']
PASS stage3-lane-project-reference-cases-01 93 identities e30c508159d60a8c6a726813b614b28aca49e603c59715fd46cfd8dd441b9783 children ['stage3-lane-project-reference-cases-01', 'stage3-lane-project-reference-cases-04', 'stage3-lane-project-reference-cases-07', 'stage3-lane-project-reference-cases-10', 'stage3-lane-project-reference-cases-13', 'stage3-lane-project-reference-cases-16', 'stage3-lane-project-reference-cases-19', 'stage3-lane-project-reference-cases-22']
PASS stage3-lane-project-reference-cases-02 93 identities 6aebf26476472c07de6d3ab71d72d61aa407e489a00d8956fa0dfb2449d96c7d children ['stage3-lane-project-reference-cases-02', 'stage3-lane-project-reference-cases-05', 'stage3-lane-project-reference-cases-08', 'stage3-lane-project-reference-cases-11', 'stage3-lane-project-reference-cases-14', 'stage3-lane-project-reference-cases-17', 'stage3-lane-project-reference-cases-20', 'stage3-lane-project-reference-cases-23']
PASS all 106367 identity records unchanged after masking unit; 273 project-reference identities retained exactly once
PASS old case 00 first identity moved to stage3-lane-project-reference-cases-18
PASS actual quiet-run multiset: old 3 shards = new 24 shards = 273 test identities; digest 4e844bce8ec75ded6bc9cb3b09d53f303281529ebe8c1090aaa73570dbd58162
PASS quiet 24/24 shards; counts {'pass': 273, 'fail': 0, 'skip': 0}
PASS loaded 24/24 shards; counts {'pass': 273, 'fail': 0, 'skip': 0}
```

## Mutant evidence

Only a scratch reference baseline is changed: tests/baselines/reference/tsserver/projectReferencesSourcemap/usageProject/configHasNoReference/can-go-to-definition-correctly.js. Prepend exactly `// planted wrong baseline for task nmje8wj` plus LF. The original prepared artifact is untouched. The associated identity is ee7c2920d0eb826d8082c644455639f7ff321cc81d922f73b837ef9137c56602, old owner 00, new owner 18. The mutated baseline SHA-256 is 42dcf917fdb2f0b736e1f371f254f9dadea4c7c7e5afe8489047589c52a16756.

The owner's unchanged inventory digest passes; upstream's real baseline oracle reports baseline-content and exits 1. The shard then rejects counts, failure title, baseline path and upstream exit, with exactly one failed execution.

```text
mutant baseline tests/baselines/reference/tsserver/projectReferencesSourcemap/usageProject/configHasNoReference/can-go-to-definition-correctly.js sha256 42dcf917fdb2f0b736e1f371f254f9dadea4c7c7e5afe8489047589c52a16756
{"name": "stage3-lane-project-reference-cases-18", "status": "fail", "counts": {"pass": 8, "fail": 1, "skip": 0}, "seconds": 2.613986354000076, "errors": ["counts: expected {'fail': 0, 'skip': 0, 'pass': 9}, observed {'pass': 8, 'fail': 1, 'skip': 0}", "baseline paths: expected [], observed ['tsserver/projectReferencesSourcemap/usageProject/configHasNoReference/can-go-to-definition-correctly.js']", "failure titles: expected [], observed ['unittests:: tsserver:: projectReferencesSourcemap:: with project references and tsbuild source map from project that uses dependency: goToDef when main tsconfig doesnt have project reference can go to definition correctly']", "upstream exit: 1"]}
PASS mutant caught by new owner with exactly one failed test and the planted baseline diff
```

### unittests:: tsserver:: projectReferencesSourcemap:: with project references and tsbuild source map from project that uses dependency: goToDef when main tsconfig doesnt have project reference can go to definition correctly (baseline-content)

```text
New baseline created at tests/baselines/local/tsserver/projectReferencesSourcemap/usageProject/configHasNoReference/can-go-to-definition-correctly.js
```

Elapsed: 765.549 ms; timeout: 40000 ms.

Stack top:     at writeComparison (src/harness/harnessIO.ts:1490:27)

tsserver/projectReferencesSourcemap/usageProject/configHasNoReference/can-go-to-definition-correctly.js (first 40 lines)

```diff
--- reference/tsserver/projectReferencesSourcemap/usageProject/configHasNoReference/can-go-to-definition-correctly.js
+++ local/tsserver/projectReferencesSourcemap/usageProject/configHasNoReference/can-go-to-definition-correctly.js
@@ -1,4 +1,3 @@
-// planted wrong baseline for task nmje8wj
 Info seq  [hh:mm:ss:mss] currentDirectory:: /home/src/Vscode/Projects/bin useCaseSensitiveFileNames:: false
 Info seq  [hh:mm:ss:mss] libs Location:: /home/src/tslibs/TS/Lib
 Info seq  [hh:mm:ss:mss] globalTypingsCacheLocation:: /home/src/Library/Caches/typescript

```

The new regression check also rejects restoring all partition counts to three in memory:

```text
test_project_reference_split_preserves_old_owners (test_shards.ShardTests.test_project_reference_split_preserves_old_owners) ... FAIL

======================================================================
FAIL: test_project_reference_split_preserves_old_owners (test_shards.ShardTests.test_project_reference_split_preserves_old_owners)
----------------------------------------------------------------------
Traceback (most recent call last):
  File "/workspace/adamic/stage3/lane/test_shards.py", line 107, in test_project_reference_split_preserves_old_owners
    self.assertGreater(len(children), 3)
AssertionError: 3 not greater than 3

----------------------------------------------------------------------
Ran 1 test in 2.963s

FAILED (failures=1)
PASS restoring the unsplit three-group partition is caught by the new regression check
```

## Toolchain and commands

Setup command: `export GOPROXY='https://proxy.golang.org|direct'; timeout 600 bash cloud/setup.sh > /tmp/nmje8wj-setup.log 2>&1`. Exit 124 while cloning cohere/TypeScript; no submodules-ready or Go-warming timing line was printed. Its complete timing lines were:

```text
setup: node ready (0.048s)
setup: go ready (0.050s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0.376s)
setup: markdown dependencies installed (npm ci, integrity verified); step-duration=0.530s
setup: markdown dependencies ready (0.673s)
```

Sourced /workspace/adamic-tools/env.sh (the configured tool directory). Node v24.19.0, Go go1.27.1, clang 20.1.8. The first Go attempt failed because go.work names the incomplete cohere/TypeScript/tsc module; GOWORK=off uses the lane package's actual module dependencies and succeeds in compiling it. No workspace/module file was edited.

Initial prepare_shards.py failed apply because ADAMIC_TYPESCRIPT_SOURCE was absent. Fetched the pinned v6.0.3 checkout at 050880ce59e30b356b686bd3144efe24f875ebc8 into /tmp/nmje8wj-typescript and installed locked stage3/api and stage3/typescript packages. Retried preparation passed in 128.555 seconds: apply 105.342, install 1.664, build 13.780, harness 0.636 seconds. Artifact key 5946a1bbb64fdeef175f4d9632f803a7f839062a06756926f9e65de55cedeffb; parser digest 1512f1871eebe448dac94546c5a6337daa1cb64ff20c80f3d9e93f85a0f48dce.

An initial product-path attempt rejected the legacy artifact's existing node_modules before upstream tests began; restarted using the scratch product-layout view. An exploratory 24-group run was discarded when a comment edit changed the runtime hash during measurement. Final generation and all final measurements use stable, current hashes.

Commands executed with stdout/stderr sent to the named logs, never piped from a test process:

```sh
source /workspace/adamic-tools/env.sh
export GOPROXY='https://proxy.golang.org|direct'
export ADAMIC_TYPESCRIPT_SOURCE=/tmp/nmje8wj-typescript
timeout 180 npm ci --prefix stage3/api --no-audit --no-fund > /tmp/nmje8wj-api-install.log 2>&1
timeout 180 npm ci --prefix stage3/typescript --no-audit --no-fund > /tmp/nmje8wj-typescript-install.log 2>&1
timeout 600 python3 stage3/lane/prepare_shards.py /tmp/nmje8wj-artifacts > /tmp/nmje8wj-prepare-retry.log 2>&1
export STAGE3_ADAPTED_TREE=/tmp/nmje8wj-product-view
timeout 300 python3 -B /tmp/nmje8wj-measure.py /tmp/nmje8wj-before-valid all > /tmp/nmje8wj-before-valid-measure.log 2>&1
timeout 60 python3 -B stage3/lane/shard_plan.py /tmp/nmje8wj-artifacts/5946a1bbb64fdeef175f4d9632f803a7f839062a06756926f9e65de55cedeffb > /tmp/nmje8wj-generate-final.log 2>&1
timeout 30 python3 -B stage3/lane/generate_shards.py >> /tmp/nmje8wj-generate-final.log 2>&1
timeout 300 python3 -B /tmp/nmje8wj-measure.py /tmp/nmje8wj-after-final > /tmp/nmje8wj-after-final-measure.log 2>&1
timeout 60 python3 -B /tmp/nmje8wj-floor.py > /tmp/nmje8wj-floor.log 2>&1
timeout 60 python3 -B /tmp/nmje8wj-mutant.py > /tmp/nmje8wj-mutant.log 2>&1
timeout 120 python3 -B -m unittest discover -s stage3/lane -p test_shards.py -v > /tmp/nmje8wj-test-shards.log 2>&1
timeout 30 python3 -B stage3/lane/generate_shards.py --check > /tmp/nmje8wj-generated-check.log 2>&1
GOWORK=off timeout 120 go vet ./stage3/lane > /tmp/nmje8wj-vet.log 2>&1
GOWORK=off timeout 600 go test ./stage3/lane -run '^TestProduct_Stage3AdaptedTree$' -count=1 -timeout 9m -v > /tmp/nmje8wj-product-retry.log 2>&1
```

The measurement driver used above (its outer timeout bounds the coordinator, and shard.py enforces its unchanged 30-second per-unit limit):

```python
import json,os,sys,subprocess,time,signal
from pathlib import Path
sys.path.insert(0,'/workspace/adamic/stage3/lane')
import shard,shard_plan
plan=shard_plan.load_plan()
root=Path(sys.argv[1]); root.mkdir(parents=True,exist_ok=True)
units=[u for u in plan['units'] if 'project-reference-cases-' in u['name'] or (len(sys.argv)>2 and u['name'] in ['stage3-lane-file-000','stage3-lane-file-001','stage3-lane-file-002'])]
cpus=sorted(os.sched_getaffinity(0))[:4]; os.sched_setaffinity(0,cpus)
os.environ.update(GOMAXPROCS='8',NODE_DISABLE_COMPILE_CACHE='1',PYTHONDONTWRITEBYTECODE='1')
ledger={'cpus':cpus,'gomaxprocs':8,'cpu_workers':8,'rows':[]}
for mode in ['quiet','loaded']:
 workers=[]
 try:
  if mode=='loaded':
   workers=[subprocess.Popen(['python3','-c','while True: pass'],start_new_session=True) for _ in range(8)]
  for u in units:
   out=root/mode; out.mkdir(exist_ok=True)
   code=shard.run_unit(u['name'],out,u['inputs_hash'])
   r=json.loads((out/u['name']/'report.json').read_text())
   ledger['rows'].append({'mode':mode,'name':u['name'],'seconds':r['seconds'],'status':r['status'],'counts':r['counts'],'errors':r['errors']})
   (root/'ledger.json').write_text(json.dumps(ledger,indent=2)+'\n')
   if code: raise RuntimeError(r['errors'])
 finally:
  for w in workers: os.killpg(w.pid,signal.SIGKILL)
  for w in workers: w.wait()
print(json.dumps(ledger),flush=True)
```

The scratch-only baseline mutant driver:

```python
import json,sys,os,hashlib
from pathlib import Path
sys.path.insert(0,'/workspace/adamic/stage3/lane'); import shard,shard_plan
source=Path('/tmp/nmje8wj-artifacts/5946a1bbb64fdeef175f4d9632f803a7f839062a06756926f9e65de55cedeffb/tree')
target=Path('/tmp/nmje8wj-mutant-tree')
relative=Path('tests/baselines/reference/tsserver/projectReferencesSourcemap/usageProject/configHasNoReference/can-go-to-definition-correctly.js')
def private(src,dst,rel):
 if rel==relative:
  dst.write_bytes(b'// planted wrong baseline for task nmje8wj\n'+src.read_bytes())
 elif relative.is_relative_to(rel):
  dst.mkdir()
  for child in src.iterdir():
   if child.name!='node_modules':private(child,dst/child.name,rel/child.name)
 else:dst.symlink_to(src,target_is_directory=src.is_dir())
private(source,target,Path('.'))
os.environ['STAGE3_ADAPTED_TREE']=str(target)
plan=shard_plan.load_plan()
unit=next(u for u in plan['units'] if u['name']=='stage3-lane-project-reference-cases-18')
print('mutant baseline',relative,'sha256',hashlib.sha256((target/relative).read_bytes()).hexdigest(),flush=True)
code=shard.run_unit(unit['name'],Path('/tmp/nmje8wj-mutant-results'),unit['inputs_hash'])
assert code==1
r=json.loads((Path('/tmp/nmje8wj-mutant-results')/unit['name']/'report.json').read_text())
assert r['counts']['fail']==1,r
assert r['baseline_diffs']==[str(relative.relative_to('tests/baselines/reference'))],r
assert r['failed_tests']==['unittests:: tsserver:: projectReferencesSourcemap:: with project references and tsbuild source map from project that uses dependency: goToDef when main tsconfig doesnt have project reference can go to definition correctly'],r
print('PASS mutant caught by new owner with exactly one failed test and the planted baseline diff',flush=True)
```

## Validation logs

```text
test_case_partitions_reject_invalid_inventory (test_shards.ShardTests.test_case_partitions_reject_invalid_inventory) ... ok
test_generated_plan_preserves_every_execution (test_shards.ShardTests.test_generated_plan_preserves_every_execution) ... ok
test_input_hash_guard_detects_changed_harness (test_shards.ShardTests.test_input_hash_guard_detects_changed_harness) ... ok
test_inventory_detects_swapped_case_with_unchanged_counts (test_shards.ShardTests.test_inventory_detects_swapped_case_with_unchanged_counts) ... ok
test_killed_at_limit_fails_naming_unit (test_shards.ShardTests.test_killed_at_limit_fails_naming_unit) ... {"name": "one", "status": "fail", "counts": {}, "seconds": 31.0, "errors": ["one timed out after 31.0 seconds, past its 30 second limit"]}
{"name": "one", "status": "fail", "counts": {}, "seconds": 31.0, "errors": ["one timed out after 31.0 seconds, past its 30 second limit"]}
ok
test_manifest_exact_schema_and_counts (test_shards.ShardTests.test_manifest_exact_schema_and_counts) ... ok
test_merge_rejects_changed_union_despite_equal_counts (test_shards.ShardTests.test_merge_rejects_changed_union_despite_equal_counts) ... ok
test_merge_rejects_newline_tampering_with_all_other_checks_green (test_shards.ShardTests.test_merge_rejects_newline_tampering_with_all_other_checks_green) ... ok
test_project_reference_split_preserves_old_owners (test_shards.ShardTests.test_project_reference_split_preserves_old_owners) ... ok
test_run_past_limit_passes_and_records_seconds (test_shards.ShardTests.test_run_past_limit_passes_and_records_seconds) ... {"name": "one", "status": "pass", "counts": {"pass": 1, "fail": 0, "skip": 0}, "seconds": 31.0, "errors": []}
{"name": "one", "status": "pass", "counts": {"pass": 1, "fail": 0, "skip": 0}, "seconds": 31.0, "errors": []}
ok
test_view_keeps_mutable_outputs_and_configuration_private (test_shards.ShardTests.test_view_keeps_mutable_outputs_and_configuration_private) ... ok
test_watcher_grep_selects_exactly_one_case (test_shards.ShardTests.test_watcher_grep_selects_exactly_one_case) ... ok
test_changed_ready_input_identity_is_rejected (test_shards.SharedArtifactTests.test_changed_ready_input_identity_is_rejected) ... ok
test_failed_producer_phase_is_rejected (test_shards.SharedArtifactTests.test_failed_producer_phase_is_rejected) ... ok
test_payload_corruption_is_rejected_before_execution (test_shards.SharedArtifactTests.test_payload_corruption_is_rejected_before_execution) ... ok
test_stock_parser_corruption_is_rejected (test_shards.SharedArtifactTests.test_stock_parser_corruption_is_rejected) ... ok

----------------------------------------------------------------------
Ran 16 tests in 6.155s

OK
```

The scratch singleton-floor driver (only the in-memory plan changes):

```python
import sys,json,os
from pathlib import Path
from unittest.mock import patch
sys.path.insert(0,'/workspace/adamic/stage3/lane'); import shard,shard_plan
os.sched_setaffinity(0,[0,1,2,3]); os.environ['GOMAXPROCS']='8'
p=shard_plan.load_plan(); groups=dict(shard_plan.CASE_GROUPS)
groups['src/testRunner/unittests/tsserver/projectReferencesSourcemap.ts']=201
with patch.object(shard_plan,'CASE_GROUPS',groups):plan,rows=shard_plan.build_plan(p['artifact'],p['api_digest'])
for u in [u for u in plan['units'] if 'project-reference-cases-' in u['name']][:3]:
 code=shard.run_unit(u['name'],Path('/tmp/nmje8wj-floor'),u['inputs_hash'],plan=plan)
 assert code==0
```

Product build log:

```text
=== RUN   TestProduct_Stage3AdaptedTree
=== PAUSE TestProduct_Stage3AdaptedTree
=== CONT  TestProduct_Stage3AdaptedTree
    products_test.go:180: build stage3-adapted-tree f8f1ab590db5 miss 130.20
--- PASS: TestProduct_Stage3AdaptedTree (130.21s)
PASS
ok  	github.com/system-inc/adamic/stage3/lane	130.210s
```

`generate_shards.py --check`, `git diff --check`, and `GOWORK=off go vet ./stage3/lane` exited 0 with empty logs.

The existing guard-mutant helper was updated only to locate the new `groups` predicate. Command: `timeout 120 python3 -B stage3/lane/prove_shard_guards.py > /tmp/nmje8wj-shard-guards.log 2>&1`. All five mutants were caught:

```text
MUTANT multiple-task-guard
CAUGHT multiple-task-guard
MUTANT too-few-titles-guard
CAUGHT too-few-titles-guard
MUTANT unanchored-case-groups
CAUGHT unanchored-case-groups
MUTANT missing-planned-task
CAUGHT missing-planned-task
CAUGHT stale-generated-declaration
Five guard mutants caught; source and production plan unchanged
```

Final artifact integrity check (exit 0):

```text
PASS generator plan, manifest and listing equal committed-format outputs; runtime matches
PASS shard.py, Go deadlines, expected counts and API sanction unchanged from main
PASS 398 units; partition groups {'src/testRunner/unittests/tsserver/projectReferencesSourcemap.ts': 24, 'src/testRunner/unittests/services/convertToAsyncFunction.ts': 3, 'src/testRunner/unittests/tsbuild/noEmit.ts': 3, 'src/testRunner/unittests/tsc/noEmit.ts': 3}
```

Original artifact integrity after the mutant (exit 0):

```text
PASS original prepared artifact payload unchanged after all measurements and baseline mutant: 5d8a8d51a62caf20548d3fcf4213ddf390334258d58515bc399a203150a146c1
```

Complete lane Go validation:

```sh
GOWORK=off timeout 840 go test ./stage3/lane -count=1 -skip '^TestProduct_' -parallel 4 -timeout 13m -json > /tmp/nmje8wj-go-tests.log 2>&1
```

Exit 0. All 417 non-product tests passed, including all 398 shard tests and all Python wrappers. The product test passed separately, so all 418 declared Go tests were executed successfully. No full repository gate is claimed.

```text
{"Action": "pass", "Elapsed": 470.711, "Package": "github.com/system-inc/adamic/stage3/lane", "Time": "2026-10-23T21:05:29.954342431Z"}
PASS TestEveryStage3PythonSuiteRunsOrSaysWhy 0.050s
PASS TestShardsMatchManifest 0.110s
PASS TestScannerScratchRunnerPython 0.380s
PASS TestLaneArtifactPython 0.190s
PASS TestApplyProofShardPython 0.200s
PASS TestVerdictPython 0.110s
PASS TestLaneTablePython 0.180s
PASS TestAnyReturnsFixture 1.450s
PASS TestCachesPython 0.090s
PASS TestApplyPythonGroupsCoverEveryTestOnce 0.110s
PASS TestLaneWorkerEnvironmentPython 0.090s
PASS TestLaneProfilePython 0.130s
PASS TestLaneMeasurePlanPython 0.650s
PASS TestLaneCaseObserverPython 0.090s
PASS TestApplyProductPython 0.260s
PASS TestPythonSuiteGapsNamesWhatNoWrapperRuns 0.000s
PASS TestLaneShardsPython 7.060s
PASS TestLaneCheckPython 9.210s
PASS TestApplyCheckoutPython 13.030s
```

The unchanged 50-second Go shard budget and standalone 30-second budget remain in force. The one-third quiet wall target remains unmet for the measured reason above; the committed result supplies the smaller generated shards, complete coverage and failure proofs without claiming that target passed.
