Bound artifact/API identity at run time; regenerated and proved all 151 Go shards on the fixed six-product apply base.
Products branch: typescript/stage3-lane-products-3 at 5f38748e9453e7ca5cb4dc549d0b0b8d2c585662; shards branch: typescript/stage3-lane-shards-2, final tip in delivery.
All 180 Go tests pass; requested cold walls, exact commands and exits are below, with every top-level test under 60 s.
The new-artifact whole/shard union matches 106,367 identities, outcomes and multiplicities; all requested mutants and retained guard mutants caught.
Only stage3/lane/** changed; no main rebase, cloud/fast-gate/run.py edit, full compiler gate, other platform or Loom submission claimed.

# Runtime identity and unchanged acceptance

The committed plan omits artifact and api_digest. Committed unit inputs_hash binds only runtime plus the entire unit definition. Its argv contract stays unchanged. shard.py recomputes artifact.input_key() for the current checkout and toolchain and requires ready.json to match both that key and the full input description. A mismatch names both keys. shards-ready.json must name the same artifact as ready.json; its API digest is checked against every actual API payload byte. Tree payload, producer phases, runtime, unit hash, test identities, counts, failure titles, task completion, baseline paths, byte diffs, sanction, upstream exit and deadlines remain checked. Merge CLI verifies the current artifact and every report/whole execution artifact identity.

All 151 task/grep/names-digest/expect/baseline/measured-definition records exactly equal 99134b36 after excluding only inputs_hash. Rebinding changes only runtime and derived unit hashes. No artifact is needed to generate either manifest. See definition-comparison.json and the parent HASH_FIX_REPORT.md.

# Requested cold measurements after both re-cuts

Each row is a fresh go test process with -count=1 -timeout 90s and a 90-second external subprocess bound (95-second coordinator allowance for teardown in product runs). Predecessors were prepared. Each product directory was moved aside before its test; logs show miss, and its predecessor shows hit. The previous copy was deleted after success. The rebuilt final ready marker, including the complete payload digest, exactly matched the initial new-base build. The three shard records therefore refer to the identical artifact used by the whole and remaining shards. OS page cache was retained.

| Test | Go wall seconds | Process seconds | Exit | Flag >30s |
|---|---:|---:|---:|---|
| TestProduct_Stage3LaneInstall | 19.96 | 20.800290 | 0 | no |
| TestProduct_Stage3LaneCompiler | 49.71 | 50.483433 | 0 | yes |
| TestProduct_Stage3LaneHarness | 42.57 | 43.923902 | 0 | yes |
| TestShardsMatchManifest | 0.04 | 0.818574 | 0 | no |
| TestShard_Stage3LaneFile001 | 18.69 | 19.481784 | 0 | no |
| TestShard_Stage3LaneByteComparisons | 12.90 | 13.672336 | 0 | no |
| TestShard_Stage3LaneFile010 | 20.54 | 21.402868 | 0 | no |

File010 was the slowest wave2 shard at 20.86 s; it is rerun here without changing its partition. Compiler and harness product walls are flagged above 30 s; both remain below 60 s. No top-level test exceeds 60 s.

# Whole lane and exact union

```json
{
  "status": "pass",
  "errors": [],
  "counts": {
    "pass": 106366,
    "fail": 1,
    "skip": 0
  },
  "whole_counts": {
    "pass": 106366,
    "fail": 1,
    "skip": 0
  },
  "units": 151,
  "tests": 106367,
  "artifact": "331995f853472db52bdecf6875eac0018b95f752a2882f13ac6d15476cd4ecbe",
  "failed_tests": [
    "unittests:: Public APIs for typescript.d.ts should be acknowledged when they change"
  ],
  "baseline_diffs": [
    "api/typescript.d.ts"
  ],
  "whole_verdict": "pass",
  "missing_tests": [],
  "extra_tests": []
}
```

| Comparison command | Seconds | Exit |
|---|---:|---:|
| whole: `["bash", "stage3/lane/run.sh", "/workspace/scratch/hefhych-recut/full-union/whole", "--artifact", "/workspace/scratch/hefhych-wave2/product-cache/local/c5f9d634f8ea40b4f488abf0325d5abec10791f25ebad12b6725712ff2f8aaff"]` | 330.079944 | 0 |
| Go-units: `["python3", "-B", "stage3/lane/measure_go_units.py", "/workspace/scratch/hefhych-recut/measure", "--resume"]` | 2643.605251 | 0 |
| union: `["python3", "-B", "stage3/lane/prove_union.py", "/workspace/scratch/hefhych-recut/measure", "/workspace/scratch/hefhych-recut/full-union/whole", "/workspace/scratch/hefhych-wave2/product-cache/local/c5f9d634f8ea40b4f488abf0325d5abec10791f25ebad12b6725712ff2f8aaff", "/workspace/scratch/hefhych-recut/full-union/comparison"]` | 38.648172 | 0 |
| Python-union: `["python3", "-B", "stage3/lane/prove_python_union.py", "/workspace/scratch/hefhych-recut/measure", "/workspace/scratch/hefhych-recut/full-union/python-comparison"]` | 22.072032 | 0 |
| `["python3", "-B", "stage3/lane/merge_shards.py", "/workspace/scratch/hefhych-recut/measure/shards", "--whole", "/workspace/scratch/hefhych-recut/full-union/whole", "--output", "/workspace/scratch/hefhych-recut/full-union/comparison/merged.json"]` | 12.904727 | 0 |
| `["python3", "-B", "stage3/lane/merge_shards.py", "/workspace/scratch/hefhych-recut/measure/shards", "--whole", "/workspace/scratch/hefhych-recut/full-union/whole", "--output", "/workspace/scratch/hefhych-recut/full-union/comparison/drop-shard.json"]` | 13.004224 | 1 |
| `["python3", "-B", "stage3/lane/merge_shards.py", "/workspace/scratch/hefhych-recut/measure/shards", "--whole", "/workspace/scratch/hefhych-recut/full-union/whole", "--output", "/workspace/scratch/hefhych-recut/full-union/comparison/restored.json"]` | 12.651265 | 0 |

The complete lane ran with its original eight workers. All 151 Go shards ran with fresh private results. The merger compared actual per-test identity/outcome counters with multiplicity, raw baseline bytes and the original whole-lane acceptance check. Nothing was dropped or doubled. Dropping shard 001 exits1 and names every one of its 273 missing executions, with zero extra tests. Restoring its results makes the union pass again.

# Every mutant and its catcher

| Independent run | Catcher | Child exit |
|---|---|---:|
| Append a line to stage3/census/tsc-closure/README.md; rebuild all three products, leave both manifests unchanged | TestShard_Stage3LaneFile001 passes273 cases using the new fetched product; manifest fingerprint unchanged | 0 |
| Same edited tree, supply the pre-edit final artifact instead | Actual manifest shard command refuses, naming checkout and ready keys | 1 |
| Append a comment to shard.py | Actual Go shard refuses: shard harness inputs changed; regenerate the manifest | 1 |
| delete-generated-test | generated shard declarations differ | 1 |
| name-drift | generated shard declarations differ | 1 |
| added-manifest-unit | generated shard declarations differ | 1 |
| extra-Go-shard | manifest/test names differ | 1 |
| delete-Python-wrapper | Python Go unit names differ | 1 |
| added-old-command-case | Python units differ from old command case discovery | 1 |
| dropped-executed-Python-case | Python cases differ | 1 |
| ignore-outcomes | AssertionError | 1 |
| ignore-multiplicity | AssertionError | 1 |
| unanchored-case-grep | Isolated Python assertion: test_watcher_grep_selects_exactly_one_case | 1 |
| ignored-unit-budget | Isolated Python assertion: test_budget_can_fail_with_all_other_checks_green | 1 |
| swapped-case-identity | Isolated Python assertion: test_inventory_detects_swapped_case_with_unchanged_counts | 1 |
| ignored-unit-input-hash | Isolated Python assertion: test_input_hash_guard_detects_changed_definition | 1 |
| ignored-stock-parser-corruption | Isolated Python assertion: test_stock_parser_corruption_is_rejected | 1 |
| ignored-producer-phase | Isolated Python assertion: test_failed_producer_phase_is_rejected | 1 |
| ignored-ready-input-identity | Isolated Python assertion: test_changed_ready_input_identity_is_rejected | 1 |
| ignored-payload-corruption | Isolated Python assertion: test_payload_corruption_is_rejected_before_execution | 1 |
| ignored-exact-diff-bytes | Isolated Python assertion: test_merge_rejects_newline_tampering_with_all_other_checks_green | 1 |
| ignored-executed-union | Isolated Python assertion: test_merge_rejects_changed_union_despite_equal_counts | 1 |
| Remove real shard 001 results from the completed union | merge_shards.py names missing shard and 273 missing identities/outcomes | 1 |

Each mutant is an independent run and source edits are restored in finally. Delete-generated-test and added-manifest-unit each invoke the real TestShardsMatchManifest; the latter points that guard at a private manifest copy. The changed-runtime mutant does not rebuild or regenerate the manifest.

Base artifact key: `331995f853472db52bdecf6875eac0018b95f752a2882f13ac6d15476cd4ecbe`; edited-tree key: `dcbe380bf72ab5bc9f7954be53f7a3627ecde0ee0e37ee0782b1dc621089c6d3`. The stale-artifact log explicitly prints both. Original proof manifest SHA256: `7cc609c084220ab5675c92a9c7e48c376c1134574d8f86d571b107a9af1e4372`.

# All current Go tests, slowest first

| Test | Go seconds | Process seconds | Exit | >30s |
|---|---:|---:|---:|---|
| TestProduct_Stage3LaneCompiler | 49.71 | 50.483433 | 0 | yes |
| TestProduct_Stage3LaneHarness | 42.57 | 43.923902 | 0 | yes |
| TestShard_Stage3LaneFile007 | 20.72 | 21.539563 | 0 | no |
| TestShard_Stage3LaneFile010 | 20.54 | 21.402868 | 0 | no |
| TestProduct_Stage3LaneInstall | 19.96 | 20.800290 | 0 | no |
| TestShard_Stage3LaneFile011 | 19.24 | 20.249478 | 0 | no |
| TestShard_Stage3LaneFile026 | 18.90 | 19.679526 | 0 | no |
| TestShard_Stage3LaneFile029 | 18.76 | 19.796323 | 0 | no |
| TestShard_Stage3LaneFile001 | 18.69 | 19.481784 | 0 | no |
| TestShard_Stage3LaneFile081 | 18.68 | 19.552445 | 0 | no |
| TestShard_Stage3LaneFile045 | 18.49 | 19.498213 | 0 | no |
| TestShard_Stage3LaneFile034 | 18.40 | 19.242766 | 0 | no |
| TestShard_Stage3LaneFile049 | 18.34 | 19.436321 | 0 | no |
| TestShard_Stage3LaneFile038 | 18.24 | 18.984945 | 0 | no |
| TestShard_Stage3LaneFile062 | 18.14 | 18.944853 | 0 | no |
| TestShard_Stage3LaneFile042 | 18.12 | 18.870610 | 0 | no |
| TestShard_Stage3LaneFile021 | 18.11 | 18.929543 | 0 | no |
| TestShard_Stage3LaneFile125 | 18.09 | 18.847363 | 0 | no |
| TestShard_Stage3LaneFile061 | 18.01 | 18.840077 | 0 | no |
| TestShard_Stage3LaneFile101 | 18.00 | 18.888114 | 0 | no |
| TestShard_Stage3LaneFile057 | 17.98 | 18.822447 | 0 | no |
| TestShard_Stage3LaneFile071 | 17.95 | 18.754507 | 0 | no |
| TestShard_Stage3LaneFile117 | 17.95 | 19.167840 | 0 | no |
| TestShard_Stage3LaneFile025 | 17.83 | 18.589486 | 0 | no |
| TestShard_Stage3LaneFile048 | 17.83 | 18.644349 | 0 | no |
| TestShard_Stage3LaneFile046 | 17.81 | 18.752710 | 0 | no |
| TestShard_Stage3LaneFile109 | 17.79 | 18.647431 | 0 | no |
| TestShard_Stage3LaneFile040 | 17.76 | 18.639303 | 0 | no |
| TestShard_Stage3LaneFile084 | 17.73 | 18.489373 | 0 | no |
| TestShard_Stage3LaneFile077 | 17.56 | 18.396979 | 0 | no |
| TestShard_Stage3LaneFile118 | 17.56 | 18.505689 | 0 | no |
| TestShard_Stage3LaneSymlinkCase04 | 17.52 | 18.436530 | 0 | no |
| TestShard_Stage3LaneFile110 | 17.48 | 18.307403 | 0 | no |
| TestShard_Stage3LaneFile054 | 17.45 | 18.237625 | 0 | no |
| TestShard_Stage3LaneFile131 | 17.45 | 18.305986 | 0 | no |
| TestShard_Stage3LaneFile128 | 17.44 | 18.299335 | 0 | no |
| TestShard_Stage3LaneFile126 | 17.42 | 18.291348 | 0 | no |
| TestShard_Stage3LaneFile085 | 17.41 | 18.445531 | 0 | no |
| TestShard_Stage3LaneFile076 | 17.37 | 18.615594 | 0 | no |
| TestShard_Stage3LaneFile107 | 17.36 | 18.345009 | 0 | no |
| TestShard_Stage3LaneFile086 | 17.35 | 18.083042 | 0 | no |
| TestShard_Stage3LaneFile041 | 17.33 | 18.189264 | 0 | no |
| TestShard_Stage3LaneFile080 | 17.33 | 18.237024 | 0 | no |
| TestShard_Stage3LaneFile018 | 17.32 | 18.426439 | 0 | no |
| TestShard_Stage3LaneFile114 | 17.28 | 18.265985 | 0 | no |
| TestShard_Stage3LaneFile009 | 17.25 | 18.039527 | 0 | no |
| TestShard_Stage3LaneFile060 | 17.25 | 18.325849 | 0 | no |
| TestShard_Stage3LaneFile121 | 17.24 | 18.144024 | 0 | no |
| TestShard_Stage3LaneFile078 | 17.23 | 18.134495 | 0 | no |
| TestShard_Stage3LaneFile097 | 17.22 | 17.994099 | 0 | no |
| TestShard_Stage3LaneFile012 | 17.20 | 17.937475 | 0 | no |
| TestShard_Stage3LaneFile102 | 17.18 | 18.099071 | 0 | no |
| TestShard_Stage3LaneFile074 | 17.17 | 18.489197 | 0 | no |
| TestShard_Stage3LaneFile002 | 17.16 | 17.977311 | 0 | no |
| TestShard_Stage3LaneFile116 | 17.12 | 17.939456 | 0 | no |
| TestShard_Stage3LaneFile036 | 17.10 | 17.971391 | 0 | no |
| TestShard_Stage3LaneFile123 | 17.10 | 18.041635 | 0 | no |
| TestShard_Stage3LaneFile091 | 17.08 | 17.851967 | 0 | no |
| TestShard_Stage3LaneFile073 | 17.07 | 17.956978 | 0 | no |
| TestShard_Stage3LaneSymlinkCase00 | 17.07 | 17.896124 | 0 | no |
| TestShard_Stage3LaneFile092 | 17.06 | 18.062981 | 0 | no |
| TestShard_Stage3LaneFile032 | 17.02 | 17.888351 | 0 | no |
| TestShard_Stage3LaneFile100 | 17.00 | 17.922713 | 0 | no |
| TestShard_Stage3LaneFile067 | 16.98 | 18.034187 | 0 | no |
| TestShard_Stage3LaneFile069 | 16.97 | 17.791905 | 0 | no |
| TestShard_Stage3LaneFile053 | 16.96 | 17.787328 | 0 | no |
| TestShard_Stage3LaneFile134 | 16.95 | 17.954175 | 0 | no |
| TestShard_Stage3LaneFile058 | 16.91 | 17.895579 | 0 | no |
| TestShard_Stage3LaneFile098 | 16.90 | 17.685210 | 0 | no |
| TestShard_Stage3LaneFile103 | 16.89 | 17.697500 | 0 | no |
| TestShard_Stage3LaneFile070 | 16.87 | 17.802331 | 0 | no |
| TestShard_Stage3LaneFile051 | 16.86 | 17.637440 | 0 | no |
| TestShard_Stage3LaneFile064 | 16.78 | 17.557687 | 0 | no |
| TestShard_Stage3LaneFile082 | 16.78 | 17.584015 | 0 | no |
| TestShard_Stage3LaneFile133 | 16.78 | 17.591825 | 0 | no |
| TestShard_Stage3LaneFile089 | 16.77 | 17.631996 | 0 | no |
| TestShard_Stage3LaneFile124 | 16.75 | 17.493892 | 0 | no |
| TestShard_Stage3LaneFile047 | 16.73 | 17.545073 | 0 | no |
| TestShard_Stage3LaneFile090 | 16.72 | 17.446637 | 0 | no |
| TestShard_Stage3LaneSymlinkCase05 | 16.71 | 17.483564 | 0 | no |
| TestShard_Stage3LaneFile115 | 16.70 | 17.645390 | 0 | no |
| TestShard_Stage3LaneSymlinkCase08 | 16.69 | 17.478879 | 0 | no |
| TestShard_Stage3LaneFile044 | 16.68 | 17.503287 | 0 | no |
| TestShard_Stage3LaneFile130 | 16.67 | 17.498475 | 0 | no |
| TestShard_Stage3LaneFile063 | 16.66 | 17.485411 | 0 | no |
| TestShard_Stage3LaneFile093 | 16.66 | 17.529940 | 0 | no |
| TestShard_Stage3LaneFile055 | 16.62 | 17.511616 | 0 | no |
| TestShard_Stage3LaneSymlinkCase09 | 16.61 | 17.478828 | 0 | no |
| TestShard_Stage3LaneFile059 | 16.54 | 17.391468 | 0 | no |
| TestShard_Stage3LaneFile132 | 16.53 | 17.388600 | 0 | no |
| TestShard_Stage3LaneFile108 | 16.51 | 17.396184 | 0 | no |
| TestShard_Stage3LaneFile075 | 16.49 | 17.442566 | 0 | no |
| TestShard_Stage3LaneFile019 | 16.48 | 17.331986 | 0 | no |
| TestShard_Stage3LaneFile122 | 16.45 | 17.232647 | 0 | no |
| TestShard_Stage3LaneFile050 | 16.42 | 17.241919 | 0 | no |
| TestShard_Stage3LaneFile065 | 16.41 | 17.293875 | 0 | no |
| TestShard_Stage3LaneFile068 | 16.37 | 17.199207 | 0 | no |
| TestShard_Stage3LaneFile031 | 16.36 | 17.202623 | 0 | no |
| TestShard_Stage3LaneFile087 | 16.36 | 17.183406 | 0 | no |
| TestShard_Stage3LaneFile111 | 16.36 | 17.493999 | 0 | no |
| TestShard_Stage3LaneFile120 | 16.36 | 17.285831 | 0 | no |
| TestShard_Stage3LaneSymlinkCase10 | 16.35 | 17.134861 | 0 | no |
| TestShard_Stage3LaneSymlinkCase12 | 16.33 | 17.174477 | 0 | no |
| TestShard_Stage3LaneFile017 | 16.30 | 17.153500 | 0 | no |
| TestShard_Stage3LaneFile095 | 16.26 | 17.099664 | 0 | no |
| TestShard_Stage3LaneFile106 | 16.25 | 17.044442 | 0 | no |
| TestShard_Stage3LaneFile119 | 16.24 | 17.087293 | 0 | no |
| TestShard_Stage3LaneFile016 | 16.20 | 17.215318 | 0 | no |
| TestShard_Stage3LaneSymlinkCase11 | 16.18 | 16.930288 | 0 | no |
| TestShard_Stage3LaneFile113 | 16.16 | 17.176657 | 0 | no |
| TestShard_Stage3LaneFile094 | 16.14 | 17.003063 | 0 | no |
| TestShard_Stage3LaneFile112 | 16.14 | 16.997680 | 0 | no |
| TestShard_Stage3LaneFile083 | 16.13 | 17.035592 | 0 | no |
| TestShard_Stage3LaneFile056 | 16.10 | 16.837451 | 0 | no |
| TestShard_Stage3LaneFile030 | 16.08 | 17.029802 | 0 | no |
| TestShard_Stage3LaneFile052 | 16.07 | 16.941635 | 0 | no |
| TestShard_Stage3LaneFile072 | 16.05 | 16.936935 | 0 | no |
| TestShard_Stage3LaneFile088 | 16.04 | 16.842547 | 0 | no |
| TestShard_Stage3LaneFile020 | 15.99 | 16.876025 | 0 | no |
| TestShard_Stage3LaneFile037 | 15.98 | 16.730596 | 0 | no |
| TestShard_Stage3LaneFile096 | 15.98 | 16.795224 | 0 | no |
| TestShard_Stage3LaneFile027 | 15.97 | 16.833850 | 0 | no |
| TestShard_Stage3LaneFile099 | 15.96 | 16.892340 | 0 | no |
| TestShard_Stage3LaneFile066 | 15.87 | 16.735731 | 0 | no |
| TestShard_Stage3LaneFile015 | 15.85 | 16.591577 | 0 | no |
| TestShard_Stage3LaneFile104 | 15.85 | 16.585005 | 0 | no |
| TestShard_Stage3LaneFile079 | 15.84 | 16.735170 | 0 | no |
| TestShard_Stage3LaneFile105 | 15.84 | 16.743154 | 0 | no |
| TestShard_Stage3LaneFile035 | 15.80 | 16.733428 | 0 | no |
| TestShard_Stage3LaneFile022 | 15.79 | 16.635389 | 0 | no |
| TestShard_Stage3LaneFile024 | 15.78 | 16.827501 | 0 | no |
| TestShard_Stage3LaneFile039 | 15.69 | 16.431895 | 0 | no |
| TestShard_Stage3LaneFile033 | 15.63 | 16.481433 | 0 | no |
| TestShard_Stage3LaneFile014 | 15.60 | 16.323799 | 0 | no |
| TestShard_Stage3LaneFile023 | 15.52 | 16.453693 | 0 | no |
| TestShard_Stage3LaneFile028 | 15.48 | 16.175124 | 0 | no |
| TestShard_Stage3LaneFile127 | 15.43 | 16.181085 | 0 | no |
| TestShard_Stage3LaneFile129 | 15.40 | 16.222293 | 0 | no |
| TestShard_Stage3LaneFile043 | 15.29 | 16.083042 | 0 | no |
| TestShard_Stage3LaneFile013 | 15.03 | 15.773683 | 0 | no |
| TestShard_Stage3LaneSymlinkCase06 | 14.72 | 15.824192 | 0 | no |
| TestShard_Stage3LaneFile006 | 14.67 | 15.525120 | 0 | no |
| TestShard_Stage3LaneSymlinkCase07 | 14.67 | 15.613435 | 0 | no |
| TestShard_Stage3LaneFile008 | 14.64 | 15.487674 | 0 | no |
| TestShard_Stage3LaneFile005 | 14.43 | 15.335598 | 0 | no |
| TestShard_Stage3LaneFile004 | 13.40 | 14.169451 | 0 | no |
| TestShard_Stage3LaneSymlinkCase13 | 12.97 | 13.776315 | 0 | no |
| TestShard_Stage3LaneByteComparisons | 12.90 | 13.672336 | 0 | no |
| TestStage3ApplyShard00 | 12.80 | 13.683682 | 0 | no |
| TestShard_Stage3LaneFile003 | 12.65 | 13.412638 | 0 | no |
| TestShard_Stage3LaneSymlinkCase14 | 12.53 | 13.369659 | 0 | no |
| TestShard_Stage3LaneSymlinkCase15 | 12.41 | 13.257396 | 0 | no |
| TestShard_Stage3LaneSymlinkCase01 | 12.24 | 13.022423 | 0 | no |
| TestShard_Stage3LaneSymlinkCase03 | 12.12 | 12.964072 | 0 | no |
| TestShard_Stage3LaneSymlinkCase02 | 11.86 | 12.600577 | 0 | no |
| TestStage3LaneLandingLaneTests | 10.14 | 10.893836 | 0 | no |
| TestStage3LaneCompilerTracksInputs | 3.25 | 4.242912 | 0 | no |
| TestStage3LaneHarnessTracksInputs | 3.18 | 4.040803 | 0 | no |
| TestStage3LaneInstallTracksInputs | 3.11 | 3.990055 | 0 | no |
| TestStage3ApplyShard06 | 0.43 | 1.171718 | 0 | no |
| TestStage3ApplyShard07 | 0.42 | 1.235592 | 0 | no |
| TestStage3ApplyShard13 | 0.40 | 1.219505 | 0 | no |
| TestStage3ApplyShard14 | 0.40 | 1.172118 | 0 | no |
| TestStage3ApplyShard09 | 0.39 | 1.168581 | 0 | no |
| TestStage3ApplyShard10 | 0.39 | 1.278358 | 0 | no |
| TestStage3ApplyShard02 | 0.38 | 1.268268 | 0 | no |
| TestStage3ApplyShard05 | 0.38 | 1.172532 | 0 | no |
| TestStage3ApplyShard08 | 0.38 | 1.276363 | 0 | no |
| TestStage3ApplyShard11 | 0.38 | 1.185433 | 0 | no |
| TestStage3ApplyShard01 | 0.37 | 1.179875 | 0 | no |
| TestStage3ApplyShard04 | 0.37 | 1.221464 | 0 | no |
| TestStage3ApplyShard12 | 0.36 | 1.230023 | 0 | no |
| TestStage3ApplyShard15 | 0.36 | 1.172593 | 0 | no |
| TestStage3ApplyShard16 | 0.36 | 1.174064 | 0 | no |
| TestStage3ApplyShard03 | 0.35 | 1.267213 | 0 | no |
| TestStage3LaneLaneTableTests | 0.16 | 1.072273 | 0 | no |
| TestPythonUnitsMatchCommands | 0.14 | 0.921558 | 0 | no |
| TestUnionIdentityOutcomes | 0.07 | 0.820498 | 0 | no |
| TestShardsMatchManifest | 0.04 | 0.818574 | 0 | no |
| TestShardManifestRejectsMissingAndEmpty | 0.00 | 0.918007 | 0 | no |

# Python cases and other validation

python-units.json remains the exact mapping of 25 apply cases to 17 Go shards and 40 lane cases to two Go class units. TestPythonUnitsMatchCommands passed, each Go unit checked actual executed identities/outcomes, and prove_python_union.py reran both original commands and matched their 65 cases exactly. The full native lane fixture discovery also passed 77 cases. The historical WAVE2_REPORT.md and evidence outside recut/ record the prior99134b36 measurements; current measurements and comparison outputs are exclusively in evidence/union/recut/.

The ten parent acceptance guard mutants, nine child generator/coverage/outcome mutants and three real-input identity runs are all retained. Product key/manifest-negative Go tests passed again as part of the complete 180-test execution. go vet and required committed-tip lane checks are reported in delivery.

# Environment, failures and limits

```text
submodule cache: off
v24.19.0
go version go1.27.1 linux/amd64
setup: node ready (0.032s)
setup: go ready (0.032s)
setup: markdown dependencies skipped (validated lock and installed bytes); step-duration=0.013s
setup: markdown dependencies ready (0.092s)
setup: submodules ready (0.106s)
clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0.224s)
setup: shared cache ready (1.169s)
setup: go build ready (17.760s)
setup: test binaries deferred (use --warm-tests) (17.869s)
setup: build cache warm (17.871s)
setup: build-flags commit=99134b36961da71f0b9b83b08abae3c5a0fb2a12 nproc=5 cpu.max=400000 100000 go=go version go1.27.1 linux/amd64 clang=clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261) node=v24.19.0 cached=yes warm-tests=false gocacheprog=on load-before=0.27 0.08 0.04 1/655 95724 load-after=1.50 0.35 0.13 1/660 96349
setup: done on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB (17.902s)
setup: source /workspace/adamic-tools/env.sh
setup: logs /tmp/adamic-gate/setup.Q6A2lQ

```

The first apply40 preparation failed with No space left on device from inode exhaustion. No code change fixed it: five named disposable private mutant trees and seven superseded task-local product directories were removed, preserving the earlier evidence. The retry passed. Initial and retry logs are retained. Six new apply predecessors were prepared successfully under 90-second Go deadlines, using the parent's ten-minute worker limit. No product fetches source or packages from the network; supplied installs are checked against their byte pins. No HOME writes are introduced.

The products branch has exactly one commit on d590af70; the shard branch has exactly one commit on that products tip. Each commit touches only stage3/lane paths in its own branch's final net diff. No branch was rebased onto newer main, no PR or Loom submission was opened, and neither main nor an existing branch is force-pushed. The mandatory fetch/check command is still run without changing the requested base.

Not covered: a full compiler/root gate, a traced Workshop/Loom run, other platforms, remote-store latency, or changes to acceptance counts/sanctions. Nothing is estimated.
