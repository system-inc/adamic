# Canary census build check

Base: 22b8a0b9aa55c785165d8ea75556d853cacd8602 (scratch/performance/probe-rule-g-base-canary).

Head: ad1ddfe1295cbdea512db15108358c75e522426a (scratch/performance/probe-rule-g-head-canary).

Both requested branches fetched successfully, exit 0. Each exact commit checked out in the detached test worktree /workspace/adamic-rule-g-2594029. Pinned submodules and clean source recorded. Go 1.27.1. Build only: no tests, no census runs, no compiler edits. Worktree restored to e5531c8 afterward; restore and final diff exit 0.

| Step | Base 22b8a0b | Head ad1ddfe |
|---|---:|---:|
| Fetch | 0 | 0 |
| Checkout / revision verification / pinned submodule status / clean diff | 0 | 0 |
| go vet ./internal/lower | 0 | 0 |
| Generate latent overlay | 0 | 0 |
| Rename latent .go replacements to .go.txt | 0 | 0 |
| gofmt -w latent .go.txt replacements | 0 | 0 |
| Build latent census | 0 | 0 |
| Generate entry overlay | 0 | 0 |
| Rename entry .go replacements to .go.txt | 0 | 0 |
| gofmt -w entry .go.txt replacements | 0 | 0 |
| Build entry census | 0 | 0 |

Every check succeeded. No compile errors occurred. Both latent and entry build stdout/stderr are empty at both commits; vet output is empty too.

## Exact recipes

At each commit (`short` is 22b8a0b or ad1ddfe):

```sh
go vet ./internal/lower
python3 stage3/census/latent/make_overlay.py "$PWD" /tmp/ov-<short>
# normalize_overlay.py: rename each .go replacement to .go.txt and update overlay.json
# gofmt -w every .go.txt replacement
go build -buildvcs=false -overlay=/tmp/ov-<short>/overlay.json -o /tmp/latent ./stage3/census/latent/tool
python3 stage3/meter/entry_overlay.py /tmp/ov-<short> /tmp/entry-ov-<short>
# Apply the same .go.txt normalization and gofmt to entry replacements.
go build -buildvcs=false -overlay=/tmp/entry-ov-<short>/overlay.json -o /tmp/entry-census-<short> ./stage3/census/latent/tool
```

The entry-census recipe comes from stage3/meter/twice-daily.sh build_census: the entry overlay targets stage3/census/latent/tool; it is not a separate stage3/census/entry package. The .go.txt normalization follows stage3/adapt/99-correctness-core-levenshtein/prepare-census.sh. Copies of both recipes and entry_overlay.py from each exact commit are retained alongside the output overlay maps. All exact expanded commands, stdout, stderr and numeric exit codes are retained individually and in results.json. Driver: run.py; normalization helper: normalize_overlay.py.
