# Gate dependency closure proof

Mandatory setup resolves `go list -deps -test ./...` in Adamic and the 32 reviewed cohere targets in `cloud/cohere-module-targets.json`. It explicitly downloads and verifies the resulting module archives. `bash cloud/setup.sh --all-modules` retains recursive module downloads, including tools and testdata. Proxy retry/direct fallback retain Go checksum policy.

The area cohere pin is 715ba94f3608a6500086b1076ce5cb7e51b836db; its TypeScript pin is 8d550c837c90bd1805b047b7eeccc2baac2d5e7a. `klauspost/compress@v1.20.0` is absent from the requested package graphs: its imports occur in ignored generator sources. The graph has 11 owning modules, plus regexp2 v1.11.5, which Go needs when resolving imports of its overlapping v2 module. All 12 fresh-cache archives exactly match the recorded selected closure. Go may also fetch graph metadata (`.mod` files); the count is source archives, not HTTP requests.

| Loop | Before time / archives | After time / archives | Instrument |
| --- | --- | --- | --- |
| 1 | 20.542 s / 219 | 7.395 s / 12 | Commands below, fresh GOMODCACHE each run |
| 2 | 21.669 s / 219 | 7.449 s / 12 | Commands below, fresh GOMODCACHE each run |
| 3 | 39.929 s / 219 | 7.248 s / 12 | Commands below, fresh GOMODCACHE each run |
| Best of 3 | 20.542 s / 219 | 7.248 s / 12 | Interleaved on the same box and base commit |

Exact commands for each loop N, with the same PATH and `GONOPROXY=github.com/klauspost/compress` in both modes:

```
GOMODCACHE=/tmp/adamic-gate/closure-proof/before-N/cache python3 /tmp/adamic-gate/closure-proof/before.py /workspace/adamic /tmp/adamic-gate/closure-proof/before-N/tools
GOMODCACHE=/tmp/adamic-gate/closure-proof/after-N/cache python3 /workspace/adamic/cloud/setup-modules.py /workspace/adamic /tmp/adamic-gate/closure-proof/after-N/tools
```

The before helper is `git show 5794c87661b030d24cee457bbec4900f260e025d:cloud/setup-modules.py`. `cloud/measure-module-closure.py` reconstructs it and records exact commands and archive inventories in `measurements.json`. These are cold **dependency caches**, with installed toolchains and an existing build cache; not six freshly provisioned machines. Sum files were populated as Go resolved the same graph. Codex may snapshot these caches; we cannot observe its future snapshot policy.

Build flags for every cold measurement:

- before-1: `commit=5794c87661b030d24cee457bbec4900f260e025d nproc=5 cpu.max=400000 100000 go=go version go1.27.1 linux/amd64 clang=clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261) node=v24.19.0 cached=False GOFLAGS= load-before=0.05 0.32 0.20 1/687 209924 load-after=0.70 0.45 0.24 1/689 210274`
- after-1: `commit=5794c87661b030d24cee457bbec4900f260e025d nproc=5 cpu.max=400000 100000 go=go version go1.27.1 linux/amd64 clang=clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261) node=v24.19.0 cached=False GOFLAGS= load-before=0.70 0.45 0.24 1/689 210284 load-after=0.81 0.47 0.25 2/690 210626`
- before-2: `commit=5794c87661b030d24cee457bbec4900f260e025d nproc=5 cpu.max=400000 100000 go=go version go1.27.1 linux/amd64 clang=clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261) node=v24.19.0 cached=False GOFLAGS= load-before=0.81 0.47 0.25 1/690 210635 load-after=1.38 0.63 0.31 1/692 210983`
- after-2: `commit=5794c87661b030d24cee457bbec4900f260e025d nproc=5 cpu.max=400000 100000 go=go version go1.27.1 linux/amd64 clang=clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261) node=v24.19.0 cached=False GOFLAGS= load-before=1.38 0.63 0.31 1/692 210992 load-after=1.27 0.62 0.31 1/693 211334`
- before-3: `commit=5794c87661b030d24cee457bbec4900f260e025d nproc=5 cpu.max=400000 100000 go=go version go1.27.1 linux/amd64 clang=clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261) node=v24.19.0 cached=False GOFLAGS= load-before=1.27 0.62 0.31 1/693 211344 load-after=1.74 0.78 0.38 2/701 211967`
- after-3: `commit=5794c87661b030d24cee457bbec4900f260e025d nproc=5 cpu.max=400000 100000 go=go version go1.27.1 linux/amd64 clang=clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261) node=v24.19.0 cached=False GOFLAGS= load-before=1.74 0.78 0.38 1/701 211976 load-after=1.47 0.76 0.37 1/704 212320`

The stamp includes selected local module/workspace manifest bytes, catalog bytes, Go version, helper bytes, closure identities and Go environments in each workspace. Cached artifacts must retain their path, mode, size, mtime, ctime and inode. Warm runs still resolve the graph to detect changed imports. Three warm helper runs were 1.540, 1.533, 1.496 seconds (exact build flags and load averages in `warm-measurements.json`); uncached mode reverified all 12 archives. `uncached-all-bytes.json` confirms every one of 7,274 module-cache files was byte-identical before and after `ADAMIC_GATE_UNCACHED=1` (combined digest recorded).

Functional proof uses `/tmp/adamic-gate/closure-proof/after-3/cache`, not the old broad cache. Full ordinary setup passes in `worker-setup-retry.log`; the initial run hit container disk exhaustion, retained in `worker-setup.log`. Disposable installations from earlier timing proofs were removed to free space. Functional builds use `GOFLAGS=-trimpath` and the existing build cache, and the compiler-agreement test consumes the previously verified TypeScript corpus at commit 050880ce59e30b356b686bd3144efe24f875ebc8. Full `--gate-inputs` corpus provisioning was not repeated. Native, Node and Go agreement tests run uncached.

Commands and verdicts:

```
ADAMIC_SETUP_INTEGRATION=1 python3 -m unittest discover -s cloud -p test_setup_modules.py
python3 -m unittest discover -s cloud -p 'test_*.py'
go test ./cloud -count=1 -v
go vet ./cloud
ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint -run '^Test(Mutants|RulesAgree|CompilerAndStage1Agree)$' -count=1 -timeout=15m -v
go build ./...
```

All three lint tests pass (554.072 seconds total, including the disk-pressure pause); `go build ./...` passes with empty output. Module tests: 11 pass. Setup tests: 27 pass, 5 opt-in tests skipped. Coverage tests and vet pass. The exact lint-oracle `go build -overlay=...` succeeds with **zero stderr bytes** (`lint-oracle.stderr`), using the fresh reduced cache.

The offline scratch integration adds `require example.invalid/unreachable v0.0.0` to an otherwise standard-library-only module, sets GOPROXY=off and GOWORK=off, and proves mandatory setup succeeds with zero archives (cold, warm, uncached); optional all-modules fails. Thus an unused broken requirement cannot block mandatory preparation.

All 15 mutants fail a meaningful test: four omitted stamp-key components, six omitted cache-stat/path components, archive/extracted/go.mod integrity verification disabled individually, recursive downloads made mandatory, and target-coverage comparison disabled. See `mutants.json`, `mutant-*.log` and `run-module-mutants.py`. The coverage mutant is supplied by a Go build overlay, never committed to stage1. `TestModuleTargetCheckCatchesNewInvocation` catches a new cohere oracle invocation. The catalog check also rejects a newly imported overlay cohere package outside its package list.

Only cloud files changed. No full gate or other cohere differential tests were run.
