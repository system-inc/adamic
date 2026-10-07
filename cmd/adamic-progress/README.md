# Six-day progress

Run `go run ./cmd/adamic-progress`, with `--json` for structured evidence or
`--history` for hourly tables. `--help` documents every source and accounting
rule. The command reads the last fetched `origin/main`, never fetches or runs a
gate. Recorded-file and history reads share an eight-second budget; backlog
gets its own independent 60-second budget. The meter report schema is shared with `adamic-meter`.

Every measure has 25 hourly observations, now/1h/6h/24h values and a sparkline.
Each snapshot evaluates the same recorded files at main's first-parent commit
current by committer timestamp. Git does not record remote push times; these
are reconstructed commit-time observations. Meter timestamps supply an additional
cutoff. Actual pace uses the last six hours (or the oldest known point within
that window); needed pace divides remaining work by time to the deadline.
Green means actual meets needed. Track ETA extrapolates the equal-weight overall
lower bound. Unknowns contribute zero to that lower bound but remain unknown
individual observations. Flat or regressing tracks have no completion ETA.
Patch size has no target and is excluded from overall completion.

`documentation/progress/milestones.json` gives every micro-deadline its exact
measurement contract. The embedded plan works before this branch lands, while
all completion evidence remains scoped to main except the named area/library
checkpoints. Overdue red claims print first. Host success requires all 25 recorded
fixtures with Node observations matching both backends; no backend is run by this
command. Missing area/library or incomplete results are not measurable yet.

Recording interfaces are listed in help and milestone conditions. Stage 3 native
scanner/parser and diagnostics claims need explicit recorded booleans. Stage 1
whole-formatter parity and whole-cohere speed need their own recordings; the
existing syntax-lint speed record is displayed separately. Apple example presence
measures delivery; detailed app claims require recorded evidence. Malformed data
makes the affected section unavailable with its error, while the rest of the
report still prints. Missing counts are unknown, not zero.

Velocity shows both distinct commit identities and stable patch IDs. Patch IDs
are computed with `git patch-id --stable`, deduplicated across branches and exclude patches already on main, including
rebased copies. Merge/empty commits have no patch; whitespace is ignored.
Squashes or conflict resolutions can produce different patch IDs. Changed-path
sets narrow main candidates before hashing. Each Git log/patch-id pipe hashes
at most 16 commits, with fixed Myers, binary, prefix and no-renames settings.
Commit SHA -> patch ID entries persist in the file returned by
`git rev-parse --git-path adamic-progress/patch-ids-v1.json`. Each completed batch
is atomically checkpointed, so later kills/timeouts retain completed work.
A second run only hashes uncached SHAs; branch membership and landed patches are
reevaluated every run. Invalid or unwritable cache files yield an explained
missing backlog, with the rest of the report intact. Concurrent runs may lose
cache entries and rehash them later, but never publish partial JSON.

Backlog failure or timeout does not exit 1: text says `Backlog missing: <reason>`;
JSON uses `distinct_patch_backlog: null` and `patch_backlog_error`.
`patch_backlog_stats` exposes hashed/cached commit counts, batches, cache file and
monotonic wall seconds. Successful empty backlog is zero. The independent
60-second backlog deadline leaves other sections and history usable. Repository
bootstrap errors (for example missing origin/main) and output-write errors still
have nonzero exit status. Individual section failures do not.

The scale test creates 300 independent remote branches with 3,000 commits off
main (3,003 total) using Git fast-import, under a 60-second context. It tests
rebased landed changes, cross-branch patch deduplication, a persisted cache, a
warm run forbidden to invoke patch-id, and the actual full command. Failure
fixtures kill patch-id, shorten its deadline, corrupt the cache, and kill a
later batch after the first checkpoint. They verify complete text/JSON/history
reports and zero CLI exit status.

Linux x86_64 observations with Bash `time`: actual fetched checkout (386 remote
refs, 1,702 commits off main), full `go run` cold 9.167s, warm 2.245s. Cold hashed
1,623 SHAs in 102 batches; warm hashed zero. With Go `time.Now`/`time.Since`'s
monotonic wall clock, the 300-branch/3,000-commit fixture took 1.215s cold backlog,
0.060s warm backlog and 0.076s for the cached native CLI report. These are observed
cloud timings, not a measurement on Kirk's Mac. See validation.md for commands,
logs, mutants and limits.

Historical slices predating their required GAPS.md records stay unknown. Commit counts
use committer dates. `documentation/velocity/landings.csv` holds integration's
actual timestamps (optional `landings` count). Historical remote refs are not
recoverable from today's refs: past commit backlog requires `backlog.csv`; patch backlog requires
`patch-backlog.csv` with
`timestamp,count`. Its absence is shown explicitly, and the hourly-falling claim
requires a strictly falling observation every elapsed hour.

Tests use miniature Git histories, independent literal bars and sparklines,
checker/lowering records, registrations, both-backend host records, explicit
refusals, backlog histories and micro-deadline clocks. Source mutants are checked
through Go overlays, leaving production files untouched. Test output is logged
under `/tmp/progress-*.log` and `/tmp/progress-mutants/`. The touched packages,
repository vet and a filtered native/Node oracle are checked; the full gate and
port corpora are not rerun. This command observes evidence, not recertifies it.

Setup took 118 seconds: Go 0s, clang 1s, Node 1s, submodules 1s, build cache 118s.
`nproc` reported 5, with a four-CPU cgroup quota. Cached `go run` with hourly
reconstruction took about two seconds. First-time Go toolchain installation and
compilation occur before the executable starts.

The runtime parse milestones (#93z4yv7) compare instructions for parse alone on
batch 8's 77 compiler files: native/Go <=1.5 by Oct 8 12:00 MDT, <=1 by Oct 9
12:00 MDT. Record `stage1/progress.json` `parse_batch8` with `driver: "batch8"`,
`files: 77`, `unit: "instructions"`, `native_scope: "parse_alone"`,
`go_scope: "parse_alone"`, positive integer `native_instructions` and
`go_parse_instructions`. The reported 8.97G native versus 2.96G Go whole run is
context in the plan; it does not establish either parse-only claim.
