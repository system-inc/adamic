# New B pair: commands and provenance

Parent evidence commit: `90a893c24077a6fb43360e080d54d7922835da1e`.
A: `e0391d48aab3227d3aa5f4c93da39e3fbb2513be`, reused binary `8ba606522b7b8dabf4ff27dcbb8be05c62d5a1466a8d8d3394db973a71d8419f`, rerun beside new B.
B: `4a1fe09255a6442aeb7befde57a6f14c5899e423`.
Adaptation: main `4e45d80b59bdd03c054d7611c50e6b89cace1a0b`, upstream TypeScript `050880ce59e30b356b686bd3144efe24f875ebc8`. Existing `/workspace/scratch/tree` reused; hashes match the parent evidence exactly.
Shared cohere gitlink: `06af2e399c6c1fb93fd570b774027942747400b4`; its TypeScript: `d92d9bfee114c80be2c375d72edae966176e3a4f`.

The new B lacks the timing recorder. The exact meter-only binary patch from `1bab347531e8c238d97940ea843ad3c166509e18`, preserved as `../unit-times.patch.gz`, was applied identically. All 308 latent-census files and 1,042 meter files compare byte-for-byte with the prepared A worktree. No `internal/` file was edited.

```bash
# /workspace/adamic
out=/workspace/scratch/evidence/stage3/census/changeb-meter/4a1fe092
git fetch origin 4a1fe09255a6442aeb7befde57a6f14c5899e423 > "$out/fetch.log" 2>&1
git worktree add --detach /workspace/scratch/change-b-4a1fe092 4a1fe09255a6442aeb7befde57a6f14c5899e423 > "$out/worktree.log" 2>&1
git -C /workspace/scratch/change-b-4a1fe092 apply /workspace/scratch/unit-times.patch
rmdir /workspace/scratch/change-b-4a1fe092/cohere
ln -s /workspace/adamic/cohere /workspace/scratch/change-b-4a1fe092/cohere
# /workspace/scratch/change-b-4a1fe092
export PATH=/workspace/adamic-tools/go/bin:/workspace/adamic-tools/bin:$PATH
export GOPROXY='https://proxy.golang.org|direct'
python3 stage3/census/latent/make_overlay.py /workspace/scratch/change-b-4a1fe092 /workspace/scratch/overlay-b-4a1fe092 > "$out/overlay.log" 2>&1
gofmt -w /workspace/scratch/overlay-b-4a1fe092/*.go
go build -buildvcs=false -overlay=/workspace/scratch/overlay-b-4a1fe092/overlay.json -o /workspace/scratch/census-b-4a1fe092 ./stage3/census/latent/tool > "$out/build.log" 2>&1
python3 "$out/run-paired.py"
python3 /workspace/scratch/main-source/stage3/adapt/40-explicit-any/census-delta.py "$out/A.jsonl" "$out/B.jsonl" "$out" > "$out/delta.log" 2>&1
PYTHONDONTWRITEBYTECODE=1 python3 "$out/analyze.py" > "$out/analysis.log" 2>&1
```

The launcher starts both binaries before waiting, each with `timeout --kill-after=60s 5400`, `LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1`, separate `LATENT_TIMES`, census JSONL and log paths. It uses Linux `os.wait4` for peak RSS (KiB), user/system CPU and wall time, as in the previous pair. File records are buffered by the existing meter until all units finish; timing records show live progress. No additional instrumentation was introduced to count views. `analyze.py` invokes the existing meter’s `report.py.lowering_summary` for coverage and count validation and reads raw or gzip JSONL.

Final tables/report were generated with `PYTHONDONTWRITEBYTECODE=1 python3 "$out/make-report.py"`. Ledgers, timings and logs were archived with `gzip -n`; every archive passed `gzip -t`. Both analysis scripts read either raw or gzip ledgers. The rerun A ledger matches the prior A ledger SHA-256 exactly; the source hash manifest also matches the previous pair and the post-run tree.
