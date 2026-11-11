Baobab measurement evidence. See REPORT.md for results and limits.

`run-piece.py N` measures roots (N-1)*10 through N*10 with the guarded meter overlay. Piece 1 ran sequentially; later pieces run at most two independent single-file processes concurrently. Every invocation retains all compiler roots in the loaded program, selecting only the attempted file through LATENT_ONLY_FILE. stdout and stderr go to the per-file log. A file exceeding 300 seconds is terminated as a process group and its actual child exit is recorded before continuing. Streams and logs larger than 1,000,000 bytes are gzip compressed.

`collect.py` harvests every finding whose reason contains SourceFileLike, retaining distinct phases and unit observations. `classify.cjs` uses the adapted AST to identify the call and enclosing function. `checkpoint.py N` creates a snapshot commit on the separate checkpoint branch without moving the report branch. Final reporting deduplicates physical sites by location and reason; phases and unit observations remain available in sites.json and findings.json.

SourceFileLike optional members are lineMap and getPositionOfLineAndCharacter. The proposed road 2 type retains lineMap. Body evidence uses the pinned upstream source; site locations use the adapted source. The public-functions inventory includes every exported function, including @internal exports, and distinguishes these from the two functions exposed by the published API. Overload signatures are counted with their implementation.

Preparation commands used from the recorded main checkout:

```sh
source /workspace/adamic-tools/env.sh
bash cloud/setup.sh > /tmp/baobab-setup-main.log 2>&1
npm ci --prefix stage3/api > stage3/census/sourcefile-like-sites/preparation/npm-main.log 2>&1
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/baobab-typescript
# This checkout is v6.0.3, verified at 050880ce59e30b356b686bd3144efe24f875ebc8.
bash stage3/apply.sh /workspace/scratch/baobab-adapted > /tmp/baobab-apply.log 2>&1
python3 stage3/census/latent/make_overlay.py /workspace/adamic /workspace/scratch/baobab-overlay > /tmp/baobab-overlay.log 2>&1
gofmt -w /workspace/scratch/baobab-overlay/*.go > /tmp/baobab-gofmt.log 2>&1
go build -buildvcs=false -overlay=/workspace/scratch/baobab-overlay/overlay.json -o /workspace/scratch/baobab-latent ./stage3/census/latent/tool > /tmp/baobab-build.log 2>&1
```

Each real shell wrapper captured and reported its child's exit code. See preparation/child-exits.json and the per-piece status and log files. The initial setup began on the environment's older checkout and failed during dependency warming; setup was repeated successfully on the aligned main tree before the census. The API install was also repeated successfully on main. The recorded adapted hashes were verified after every piece.

Use a fresh evidence directory for a new run. Existing per-file status files cause the runner to refuse replacement.

The follow-up is under continuation/. Its six baseline retries are sequential and use 1500-second bounds. collect.py prefers those retry streams over the corresponding first-pass streams, retaining all original files. variants.py refuses to start until all six retry statuses exist; this run also waited for the baseline controller and its checkpoints to finish. It restores each scratch tree with apply.sh, changes only the two requested parameter types, and verifies every src-file hash. Variant file pairs run in separate processes with at most two census children concurrently and a 1500-second bound per child.

source-audit.cjs uses stock TypeScript 6.0.3 with the source projects' defaults, including exactOptionalPropertyTypes=false. Adamic's loader uses exactOptionalPropertyTypes=true; its actual core and guarded-census diagnostics are retained separately. The whole-src check covers the union of all 13 solution projects. The independent reference and coverage audits account for library templates, renamed imports, and same-named SourceFile methods.

emit-public.cjs first emits declarations with stripInternal=true, then runs the upstream bundler over those emitted declarations. public-compare.cjs checks the whole public declaration AST after normalizing only the edited public parameter, and checks an emitted-API SourceFile consumer. B's semantic source errors are retained even though declaration emission is allowed. Files over 1 MB are gzip compressed for storage; decompress emitted declaration files before replaying the standalone consumer probes, or regenerate them with emit-public.cjs.

compare-sites.py compares complete call spans, using separate argument spans when a call has multiple baseline views. It retains all matching lowering finding kinds, guarded-census checker diagnostics, stock checker diagnostics, and unit status. followup-report.py quotes these records in REPORT.md. audit-followup.py reconciles the file inventory, callers, source diffs, bounds, and public declaration results before the second report commit is published.
