# Measurement provenance

This evidence unit pins five compiler refs in pins.json. Each compiler binary is
built from that ref's own internal/ and bridge/ sources. No production compiler
source is edited. The full nested-unit and continuing-statement overlay comes
from 93f7f8e0; 9d534d3a supplies stricter state-shape guards. The selected compiler
module check and its worker-swap mutant are logged under validation/.

The five refs have identical adaptation scripts, source pins and API locks.
The fresh main adaptation is shared as frozen input, with independent source
hashes per scope and ref. Stock TypeScript independently enumerates nested
function declarations and the 81-file tsc entry reach. Compiler-directory and
entry counts are separate observations and must not be summed.

Exact owner dispositions come from d35a81d3's stage3/refusal-table/TABLE.md and
dc6b1529's stage3/notyet-table/TABLE.md. The uppercase paths are the files present
at those commits. Unknown exact reasons remain UNOWNED. All reasons, including
zero delta rows and newly exposed rows, are retained in the tables and CSVs.

Setup used GOPROXY=https://proxy.golang.org|direct and Node 24.19.0 first on PATH.
The setup script's submodule step failed against an existing scratch path;
verified clean dependencies at the identical recorded gitlinks supplied the
workaround. validation/setup.log.txt retains the failure and timing lines.
nproc was 5; the container CPU quota was four cores and memory limit 16 GiB.

The initial uncapped main compiler-directory census was killed by the container
memory limit (exit -9, memory.events oom_kill=1). Its incomplete stream and logs
are preserved under failed-attempts/main-unbounded-memory/ and excluded from all
tables. Main was rerun with GOMEMLIMIT=2GiB. Other original runs were uncapped.
Two short front-3 profiles ran in separate diagnostic processes; neither is a
measurement stream. No performance rewrite was applied to a measured compiler.
Obsolete reproducible scratch binaries, test trees and duplicate dependency
copies were removed to provide memory headroom; evidence was preserved. The unused upstream tests/ corpus in the current
adaptation scratch was also removed; every measured src file remains intact.

Local fixtures prove three distinct refusals are recorded, checker-diagnosed
nested functions are named and excluded, and failed declarations roll back their
state. First-error-only and retained-failed-state mutants fail those checks.
IR-output and permissive-loader mutants fail the no-output guards. Shape tests
fail loudly for missing refusal functions, unsupported owned state and missing
statement instrumentation. Accounting audits additionally plant wrong totals,
phases, owners, exclusion counts, stale modes, missing nested declarations and
missing whole files. Delta tests plant missing reasons and wrong exact deltas.

This census records observed blockers. Failed compounds and signatures can hide
children or bodies; checker-diagnosed nested bodies remain excluded. Recovery
can affect isolated binding reads. No backend, final ownership or module-order
result is established. The production compiler's full gate was not run for this
measurement-only unit; local instrumentation and meter checks are retained.

Two existing meter-suite attempts timed out at their unchanged 120-second
compiler-witness limit under concurrent census load. Their logs are retained
under validation/; they are not treated as passing test runs.

All five census processes were briefly paused with SIGSTOP while the entry
probe binary was built and the meter regression suite ran on idle CPU. SIGCONT
resumed every unchanged process in a finally block. All 33 meter tests passed
with no skips, including real checker, latent-attribution and entry-reach probes.
The 120-second witness timeout was not changed. Scheduling timestamps and
per-process identities are retained in validation/idle-gate-scheduling.json.

A separate immutable predicate-query cache experiment passed exact small-fixture
parity and killed a false-on-hit mutant. Its large shadow comparison was stopped
before verification to free measurement resources. Its incomplete diagnostic
stream is excluded from every table; no measured compiler uses the cache.

The user stopped this unit because roadmap step 05 supersedes it. Main and
front-3 runners and their census children were terminated; complete file records
were archived as explicitly incomplete streams and excluded from all tables.
Area/compiler, method-values and conditions completed both scopes. Their entry
audits passed independent file/function coverage, source hashes, totals, phases,
ownership and exclusion checks; seven planted accounting/coverage mutants per
entry scope were caught. No main-based deltas were computed. Only this evidence
run directory is committed; draft instrumentation remains local.
