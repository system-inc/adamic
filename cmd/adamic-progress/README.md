# Six-day progress

Run `go run ./cmd/adamic-progress`, with `--json` for structured evidence or
`--history` for hourly tables. `--help` documents every source and accounting
rule. The command reads the last fetched `origin/main`, never fetches or runs a
gate, and shares an eight-second context across Git reads and the reused Stage 1
inventory. The meter report schema is shared with `adamic-meter`.

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
fails instead of silently becoming zero.

Velocity deduplicates fetched remote commit hashes against main. Commit counts
use committer dates. `documentation/velocity/landings.csv` holds integration's
actual timestamps (optional `landings` count). Historical remote refs are not
recoverable from today's refs: past backlog requires `backlog.csv` with
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
compilation occur before the command's eight-second read budget.
