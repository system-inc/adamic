# Proof contract

Base: adamic main 0aba99914be812ef7a0676c5ddd93024cefc81b5.
Initial controls were measured at e3f59bf5509b1feef85b405807a732e11dc383b6;
the evidence retains those runs separately from the refreshed tip checks.
TypeScript 6.0.3: 050880ce59e30b356b686bd3144efe24f875ebc8.
Count: a3a5bb705467dd4473de73a0e1a41ac7f0d45858.
Model tooling: 66e4051f19854166c95df6be99a4ae0e2eeaadd8.

Raw logs, counts, inventories of measured findings, source archives and command
receipts belong only on typescript/timer-closures-evidence-2, under
stage3/census/timer-closures/. They are not inputs in this adaptation directory.

Run contract.cjs against main's adapted tree. It checks all thirteen whole owners,
half-owner refusals, drift, duplicate and moved owners, unrelated edits, both
line endings and no-write replay with byte/mtime equality. The restoring-slot
mutant is contract.cjs MAIN --restore-slot and must exit nonzero with case
restored-variadic-slot.

Run witness.cjs MAIN and witness.cjs AFTER with the pinned API on NODE_PATH.
Their stdout must be byte-identical. The witnesses extract actual source bodies:

- The complete dynamic poller and setCustomPollingValues; swap Low and Medium
  between schedule and firing and observe the queue chosen by the real callback.
  The lower stat loop is replaced by an observable queue visit; no claim is made
  that this focused witness is an end-to-end filesystem test.
- The actual typing installer scheduleRequest method; mutate projectName before
  firing. A getter records schedule-time versus fire-time property reads and the
  actual callback sends the mutated request. This detects redundant late reads
  even though that callback intentionally ignores its label argument.
- The actual WatchHost scheduleProgramUpdate and updateProgramWithWatchStatus
  bodies, through the real TestServerHost method and Callbacks registry.
- The original blocker probe: the actual registry's pending label remains
  pollLowPollingIntervalQueue before firing.

witness.cjs AFTER poller, AFTER request, and AFTER label must each exit nonzero,
with poller-schedule-capture, request-schedule-capture and timer-label-baseline
respectively. These in-memory mutants replace one captured callback argument
with the original expression or omit one scheduled label. They do not alter the
adapter's pins to make a mutant look approved.

The remaining acceptance checks are actual adamic types on the 79 compiler
roots, stock 6.0.3 diagnostics compared with main, a fresh full apply and
zero-write replay after this stage, all 398 lane manifest shards with matching
per-shard counts and baseline bytes except the six sanctioned API members,
and a full-mode sys.ts latent census before/after. LATENT_FULL=1 and
LATENT_ONLY_FILE identify the exact sys.ts measured. The evidence report records
actual command exits and all gone/new sites, without treating lowering findings
as generated native code.
