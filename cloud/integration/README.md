# The merge tree

Main takes area branches, and each area takes its own Circle's work. Approved by Kirk, October 6.

```
workers' branches ──> area/<area> ──> speculative prefixes ──> main
   (each Circle's)     (owner merges,     (integration gates      (whole gate,
                        fast tests)        several at once)         uncached)
        ^                                                              |
        └───────────────────── merge-back after every push ────────────┘
```

## Areas

`areas.tsv` lists every area, its owner and its outside oracle. Today: compiler, runtime, library,
stage1-lint, stage1-format, stage3, platforms (Apple, WebAssembly and Cloudflare Workers) and
developer-tools.

The owner merges its workers' branches into `area/<area>` with `area-merge.sh`, resolves its own
conflicts, and keeps the area green on its fast tests. A branch that passes against an outside oracle
(test262 with zero disagreements for library, Go cohere byte for byte for stage 1) merges with no
reader. A branch without one (language, memory, concurrency) gets a reader before it merges; the
owner runs it. Language design and soundness questions still go to @system_adamic first.

```
cloud/integration/area-merge.sh <area> <branch> <sha> [--no-push] [test262 filter...]
```

It merges at the exact sha you name, runs gofmt, vet and the tests of every package the merge
changed, internal/oracle whole when the compiler, runtime or oracle changed (only the changed
fixtures otherwise), and test262 on the filters you name for the library area, all uncached. It
pushes the area only when all of that passes. It keeps one worktree per area under
`ADAMIC_AREA_WORKTREES` (default `~/.adamic-areas`) and never touches main.

Run on a Mac, those fast tests don't cover leaks: `leaks --atExit` can't see into the size-class
allocator, and a leak mutant survives it (d56d8c3). Leak coverage comes from LeakSanitizer in main's
Linux gate, so an area green on a Mac is never the last word on memory.

## Main

Only integration moves main, and only to a sha whose whole gate ran green and uncached
(`ADAMIC_GATE_UNCACHED=1`, every oracle "gate cache:" line at hits 0, gofmt and vet clean, every skip
named). The rule hasn't changed; the tree only changes what reaches it.

- `speculate.sh <label> <area>...` pushes `cloud/speculate-<label>-<n>`, main plus the first n
  areas, for every n, so the prefixes gate at once. The longest green prefix goes to main; a red one
  is bisected by its prefixes.
- Every prefix of a stack gates on its own box, beside the whole stack (@system_adamic, October 7).
  When the whole stack is red, the longest green prefix goes to main at once and only the rest is
  bisected. Members go in order of risk: fast tracks and small, already gated areas first, merges
  that have never had a whole gate of their own last. Each prefix must be an ancestor of the next
  with valid counts on its own, so a stack whose merges conflict in counts.md is built by a seat
  that regenerates counts after every merge, not once at the end. A stack built on a gated sha lands
  over that sha's landing commit, whose tree is the same.
- `push-main.sh <sha> <gate minutes> <pass> <fail> <skip> "<branches>"` (or `--fast-gate`, `--full-gate`
  with a gate log, or `--test-only`) checks main is still the sha's ancestor or moved only by records
  and by tests outside the sha's code packages, lands it as one commit (first parent the old main,
  second the sha, the numbers as trailers), pushes it (never a force), and runs merge-back.
- `pr-lane.py` lands each open pull request into main through `push-main.sh --test-only`, every minute.
- `merge-back.sh` merges main into every area with git merge-tree, so it needs no checkout. An area
  that conflicts with main is named with its owner, who merges main by hand.

## Velocity

Each landing commit on main carries its numbers as trailers: Old-main, Landed-commits, Branches,
Gate-minutes, Pass, Fail, Skip and Backlog (commits on any branch of origin that main doesn't hold yet,
merges left out). `python3 cloud/integration/landings.py` prints them as the velocity table's CSV,
after the rows `documentation/velocity/landings.csv` held until it left main on October 9.
