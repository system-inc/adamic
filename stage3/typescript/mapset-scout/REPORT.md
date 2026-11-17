# Map/Set scout — stopped at adapted-tree build

Base: `1d3efe7b92236a0bb94aba6bfd5652f6d0f6dfa6`.
Recorded run: `meter/daily-20261117T130646Z.rUAmui` at
`e4b34e0e31c6bc661275ad1467bb9a21db805b20`.

The adapted tree did not build. The requested stop condition applies; no site classification or census was attempted after the failure.

## Build evidence

Command: `stage3/apply.sh /workspace/mapset-adapted` from an isolated checkout of the base. Exit status: **1**. See [full log](evidence/apply.log) and [commands](evidence/commands.txt).

The keyed product `6c54c265b3ddf357d631b5dc1bb7f13cf328782bdf8f8e9db38e24a5aec7fae4` was absent from the default local product store. The script attempted its local build fallback and stopped with:

```text
RuntimeError: set ADAMIC_TYPESCRIPT_SOURCE to the gate inputs' TypeScript v6.0.3 checkout at 050880ce59e30b356b686bd3144efe24f875ebc8
```

A pinned TypeScript source checkout (or the matching keyed product) is needed to resume.

## Recorded stream availability

The [complete tracked run inventory](evidence/recorded-run-files.txt) contains **no `latent.jsonl.gz`** under the run directory, on either main or area. The main-side JSONL gzip files are `main/latent-times.jsonl.gz` and `main/tsc/latent-times.jsonl.gz`; inspecting their decompressed contents showed timing records (`file`, `where`, `wall_seconds`, `cpu_seconds`), with zero lines containing the requested reason. They are not diagnostic streams and were not treated as the 35-site evidence. The main `latent.log.gz` is a per-file summary, also with zero occurrences of the reason.

Therefore the diagnostic count is **unavailable**, not zero. No diagnostic latent file was used, and the claimed 35 sites could not be extracted from the tracked run. The original diagnostic stream must also be supplied or recovered to finish the requested comparison.

## Work left unmeasured

Site/interface/construction/road table, counts by value class and road, refusal representation analysis, and full-versus-first-error census comparison remain unmeasured. No guesses or road recommendations are made.

No adaptation, ledger, or `internal/` files were changed. Only this report and its evidence are committed.
