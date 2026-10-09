Resumable whole-project speculative observations are now persisted one file at a time.
Every file loads the complete compiler project; only the census walk is filtered.
A hard timeout and address-space bound protect each job, with an additional RSS watchdog.
Independent stock ancestry combines the boundary union before publishing any depth table.
Incomplete files remain explicitly inventoried; no progress line counts as evidence.

`stream.py BINARY COMPILER_ROOT OUTPUT SECONDS RSS_MIB [MODE]` inventories every
source and hashes both the source bytes and the binary. It starts one process per
file, largest first. Each uses the same complete root list, checker diagnostics,
and project symbol registration. `LATENT_ONLY_FILE` filters the walk in speculative,
full and no-stubs modes. Completed records are synced by the driver and atomically
published with a checksum. Resume accepts only matching inputs and limits and
skips verified completed files. Failed files retain their logs and resource metrics.

`timed_run.py` uses GNU timeout's TERM deadline and a two-second KILL grace period.
RLIMIT_AS is a kernel-enforced address-space cap of twice the RSS allowance. The
process-group RSS watchdog samples /proc every 100ms and sends SIGKILL on the first
sample above the allowance. Thus the RSS allowance has sampling latency; it is not
a kernel cgroup RSS cap. GOMEMLIMIT=4GiB controls collection pressure and is not a
hard memory limit. Corpus jobs use a 30-second wall deadline, a 12GiB address-space
cap and a 6GiB RSS allowance. The standalone checker probe instead used 90 seconds.

Per-file depths are provisional: dependencies can fail ancestors before their own
file walk completes. Each record retains the complete typed boundary catalogue
seen in its process, including foreign dependency boundaries. `finalize_stream.py`
combines those catalogues by exact source, byte span and syntax kind. Stock
TypeScript 6.0.3 visits all 79 files and resolves each observed site's ancestry.
The separate report auditor recounts those depths independently, and verify.py
checks totals, rankings, bytes and the completed/incomplete inventory. Coverage is
completed-file AST examination. Findings reached while attempting a dependency
are retained and globally deduplicated; a dependency finding is not a claim that
its entire file completed.

Run `audit_stream.py BINARY OUTPUT` with NODE_PATH pointing at the pinned stage3
API cache. The two-file dependency project proves exact findings parity with the
buffered walk and correct cross-file depth finalization. Resume skips both files.
A changed checksum is rejected, a successful file's dropped record is rejected,
and a depth shifted by one fails stock ancestry. Existing nesting, signature,
binding, snapshot, topology, identity, no-output and header controls remain.
Full/no-stubs streamed dependency records are byte-identical. Sleeping, allocating
past the address-space cap, and resident-memory excess witnesses prove all three
resource controls can fail.

The first compiler-wide checker probe hit timeout after 90.099s, peak RSS
940,780KiB, exit 124, and emitted no completed file record. This proves only that
it did not finish within that budget on this box. It does not prove that a longer
run cannot finish. The all-file attempt and exact coverage are reported separately.

The dispatched `any_returns_concrete.a` source is run from a scratch copy of
a64f77ee. Main already supplies equivalent behavior through generic binder
inference in `6b33ec61da399ae79cfba0900c9f753ed38faa1a`. Source Node, backend Node
and sanitized native output agree byte for byte. A scratch Go overlay restoring
inferTypes from that commit's parent fails the fixture at `U | undefined`. No
production mapper fix or new native oracle fixture is required for this port.
The census's own opaque-mapper copier adaptation is in the census commit.
