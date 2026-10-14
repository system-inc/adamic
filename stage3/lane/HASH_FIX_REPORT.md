Moved artifact/API identity from committed manifests to verified runtime markers.
Re-cut typescript/stage3-lane-products-3 on d590af7018c55eb34ec349016b999c139201425c; final tip is in delivery.
All six apply predecessors and three cold lane product tests passed under 90-second Go limits.
Rebuilt census inputs passed shard001 without manifest changes; stale artifact and changed shard runtime were refused independently.
Only stage3/lane changes; wave 2 generation and full union are measured separately on the child branch.

# Binding and retained checks

shard-plan.json has no artifact or api_digest. Committed inputs_hash is SHA256 of runtime plus the unit definition, preserving the manifest argv contract. shard.py computes artifact.input_key() from the current tree and toolchain and compares both key and full inputs to ready.json. It reports both keys on mismatch. shards-ready.json must name the same artifact and its API payload digest must match the actual bundled API. All payload, producer phase, runtime, unit identity, count, baseline, sanction, exit and wall checks remain. merge_shards.py verifies the current artifact at the CLI and requires every report and the whole execution to name that artifact.

# Cold measurements

Fresh go test per row, -count=1 -timeout 90s -v, prepared predecessors; product logs contain miss. OS page cache retained. Node v24.19.0, npm11.9.0, Go1.27.1; nproc5, CPU quota4. Setup completed17.902s. Apply uses its parent's six products and ten-minute internal deadline, with these actual invocations still bounded to90s.

| Test | Go test wall s | Process wall s | Exit | >30s |
|---|---:|---:|---:|---|
| TestProduct_Stage3Adapted10 | 23.95 | 25.603717 | 0 | no |
| TestProduct_Stage3Adapted20 | 42.51 | 43.400154 | 0 | yes |
| TestProduct_Stage3Adapted40 | 31.69 | 32.773806 | 0 | yes |
| TestProduct_Stage3Adapted65 | 41.82 | 42.678524 | 0 | yes |
| TestProduct_Stage3Adapted70 | 57.39 | 58.145888 | 0 | yes |
| TestProduct_Stage3Adapted99 | 42.40 | 43.321736 | 0 | yes |
| TestProduct_Stage3LaneInstall | 30.79 | 31.933072 | 0 | yes |
| TestProduct_Stage3LaneCompiler | 53.52 | 54.320793 | 0 | yes |
| TestProduct_Stage3LaneHarness | 42.53 | 43.410183 | 0 | yes |

# Independent requested mutants

| Run | Observation and catcher | Exit | Seconds |
|---|---|---:|---:|
| rebuilt-Install | Fresh rebuilt product passes | 0 | 21.363532 |
| rebuilt-Compiler | Fresh rebuilt product passes | 0 | 56.002879 |
| rebuilt-Harness | Fresh rebuilt product passes | 0 | 35.038796 |
| rebuilt-input-shard001 | Fresh rebuilt product passes | 0 | 20.640122 |
| stale-artifact | Checkout and artifact ready keys both named | 1 | 4.130283 |
| changed-runtime | shard harness inputs changed | 1 | 1.429214 |

Edited stage3/census/tsc-closure/README.md by appending one line. Both shards.json and shard-plan.json stay unchanged; no bind/regeneration was run. The Go shard fetched the rebuilt final product. The stale artifact run uses the same manifest argv with STAGE3_ARTIFACT pointing to the pre-edit product. The runtime mutant appends one comment to shard.py and runs the actual Go shard. Each edit is restored in finally.

```json
{
  "base_key": "331995f853472db52bdecf6875eac0018b95f752a2882f13ac6d15476cd4ecbe",
  "changed_key": "dcbe380bf72ab5bc9f7954be53f7a3627ecde0ee0e37ee0782b1dc621089c6d3",
  "manifest_sha256": "7cc609c084220ab5675c92a9c7e48c376c1134574d8f86d571b107a9af1e4372",
  "manifest_unchanged": true,
  "source_restored": true,
  "commands": [
    {
      "label": "rebuilt-Install",
      "command": [
        "go",
        "test",
        "./stage3/lane",
        "-run",
        "^TestProduct_Stage3LaneInstall$",
        "-count=1",
        "-timeout",
        "90s",
        "-v"
      ],
      "exit": 0,
      "seconds": 21.363532142000622
    },
    {
      "label": "rebuilt-Compiler",
      "command": [
        "go",
        "test",
        "./stage3/lane",
        "-run",
        "^TestProduct_Stage3LaneCompiler$",
        "-count=1",
        "-timeout",
        "90s",
        "-v"
      ],
      "exit": 0,
      "seconds": 56.002879493000364
    },
    {
      "label": "rebuilt-Harness",
      "command": [
        "go",
        "test",
        "./stage3/lane",
        "-run",
        "^TestProduct_Stage3LaneHarness$",
        "-count=1",
        "-timeout",
        "90s",
        "-v"
      ],
      "exit": 0,
      "seconds": 35.03879618999963
    },
    {
      "label": "rebuilt-input-shard001",
      "command": [
        "go",
        "test",
        "./stage3/lane",
        "-run",
        "^TestShard_Stage3LaneFile001$",
        "-count=1",
        "-timeout",
        "90s",
        "-v"
      ],
      "exit": 0,
      "seconds": 20.640121638998608
    },
    {
      "label": "stale-artifact",
      "command": [
        "python3",
        "stage3/lane/shard.py",
        "stage3-lane-file-001",
        "$STAGE3_RESULTS",
        "--inputs-hash",
        "84ea25d526844fdb7dc7496abebcc8e62796e52ef6fe612507e76b25b30b19f6"
      ],
      "exit": 1,
      "seconds": 4.130283169000904
    },
    {
      "label": "changed-runtime",
      "command": [
        "go",
        "test",
        "./stage3/lane",
        "-run",
        "^TestShard_Stage3LaneFile001$",
        "-count=1",
        "-timeout",
        "90s",
        "-v"
      ],
      "exit": 1,
      "seconds": 1.4292142710000917
    }
  ]
}
```

Ten existing acceptance-check mutants were also caught by their isolated fixture assertions: grep, budget, identities, unit definition hash, API payload, producer failure, complete ready inputs, tree payload, exact diff bytes and executed union. Full lane Python fixtures:76 passed. Three Go TracksInputs tests and missing/empty manifest checks passed. Exact commands and exits are in evidence/runtime-identity/; prove_runtime_identity.py reproduces the real-input mutants.

# Setup and failures

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

The first apply40 attempt failed from exhausted workspace inodes while copying files. Its log is retained. Five named old private mutant tree copies and seven superseded local cache products from earlier waves were removed, preserving their proof logs. The fresh retry passed. No source or checks were changed to resolve the inode failure. No network fetch occurs inside the product recipes; source and installed packages are supplied by the tree. No HOME writes are introduced.

Cold lane commands use go test ./stage3/lane -run ^TestProduct_Stage3LanePHASE$ -count=1 -timeout 90s -v. Apply commands use ./stage3/applyproducts and ^TestProduct_Stage3AdaptedNN$. Each command exits0. timeout90 python3 -B -m unittest discover -s stage3/lane -p test_*.py exits0; timeout90 python3 -B stage3/lane/prove_shards.py exits0 with ten caught mutants. Requested deleted-generated-test/extra-manifest mutants and final union run belong to the child shard re-cut. No main rebase, full compiler gate, Loom submission or other-platform claim.
