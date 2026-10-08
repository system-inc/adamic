# Step 29 fs inspection fixes

Base: origin/area/library `beaacc1ca03bd2c5bda657d1d3ea95bad9fea8b5`.
Branch: codex/host-fs-inspect-fixes. Gate oracle: Node 24.19.0 on Linux.

The coverage observations and programs come from origin/codex/coverage-oct8-nodehost,
`notes/coverage-oct8/nodehost/REPORT.md`. Its six relevant source programs are
registered as `internal/oracle/testdata/oct8_nodehost_*.a`; the numeric and preview
programs retain the reader's cases and add the boundary sweeps below. The reader's
short_io.c is retained as `internal/oracle/testdata/nodehost_short_io.c`.

## Formatting

`internal/native/runtime/node_fs_file.c` ports Node 24.19.0
`lib/internal/util/inspect.js` strEscape, meta, quote selection, default primitive
string layout and negative-zero rendering, together with `lib/internal/errors.js`
addNumericalSeparator, ERR_OUT_OF_RANGE and the ERR_INVALID_ARG_VALUE preview.
THIRD_PARTY_NOTICES.md credits the Node sources under the existing MIT notice.

Received integers group only when their absolute value exceeds 2**32. This is the
error helper's rule, not globally enabled inspect numeric separators. Its grouping
also preserves Node's treatment of exponent notation. Ordinary length errors,
nonintegers, nonfinite numbers and the separator threshold are compared unchanged
with Node. The position expectation is specifically Node 24.19.0; 24.14.1 on the
reader's Mac does not validate that zero-read position in the same way.

Strings are inspected as UTF-16. Lone surrogates render as lowercase \ud800 or
\udfff; pairs remain literal. C0/C1 controls use Node's escape table, including
uppercase \x0B and \x85. Backtick selection rejects `${`, rather than any `$`.
Long multiline strings quote each line as Node does. Node's default 10000-source-unit
limit precedes its 128-inspected-unit error preview. An astral character cut at the
preview boundary stays a lone UTF-16 unit in the Error message; it is not decoded
as filesystem UTF-8. The fixtures cover every C0/C1 entry, pairs and lone units,
quote fallbacks, multiline layout, the 10000-unit limit and a split surrogate at
128 inspected units.

## Fixtures and mutants

All seven registered programs agree byte for byte with unchanged Node source on
JavaScript and on native release, ASan/UBSan, sanitized size classes and heap leak
checks. Programs create and remove their own temporary trees. The numeric,
zero-length read and additional path-only preview programs also agree on WASI.
The original prefix-preview, descriptor ownership, cwd and large-roundtrip programs
remain explicit WASI compile-time refusals because this base refuses mkdtempSync.
The path-only preview covers the same formatting through the admitted openSync API;
it does not replace the original prefix fixture.

Eight isolated production-runtime mutations build and execute successfully with
exit zero and empty stderr. Sanitizers and LeakSanitizer remain clean. Only raw
stdout comparison with the unchanged program on Node catches each one:

| Mutant | Witness |
|---|---|
| Disable received-number grouping | fs_numeric_error |
| Uppercase lone-surrogate hex | fs_error_preview |
| Leave C1 controls literal | fs_error_preview |
| Render vertical tab as \v | fs_error_preview |
| Remove the zero-length read return | fs_zero_read |
| Remove the owned buffer reader's close | fs_fd_ownership |
| Leave cwd cached after chdir | host_directory |
| Stop after the first successful write | fs_large_roundtrip with short-I/O shim |

Each mutation uses a separate temporary runtime copy and archive. Tests never
edit the checkout or mutate generated program behavior. The descriptor witness
observes descriptor recycling after owned-reader failures as well as preservation
of caller-owned descriptors after errors. Descriptor leakage is caught by Node
output; it is not claimed as a heap sanitizer finding.

The shim caps successful regular-file reads at 257 bytes and writes at 3 bytes,
leaving stdout/stderr and nonregular descriptors unchanged. TestNodeHostShortIO
compares the sanitized and release native program under that shim with unchanged
Node source and JavaScript. TestNodeHostCoverageMutants/short-write then proves
the shim reaches continued-write handling. Both shim checks call nodeHostShortIO,
which skips unless runtime.GOOS is linux; LD_PRELOAD and RTLD_NEXT are the platform
requirement. The ordinary large-roundtrip fixture and its counts run on all native
hosts without the shim. This covers successful short I/O, not EINTR or failed I/O.

The existing fs regression selection exposed a stale Date timestamp mutant on
the base. Date getTime/valueOf now emit adamic_date_value, so the test now mutates
that helper instead of the unused adamic_fs_file_date_time spelling. No Date
runtime behavior changed. The repaired mutant is caught only by Node comparison.

## Validation and counts

Only the touched native and oracle packages were tested or vetted. The focused
uncached comparisons, all eight mutants, the Linux short-I/O test, the existing
TestNodeFSFile regression selection, native host/runtime-storage checks and vet
passed. WASI refusal outcomes are named above, not reported as passing executions.
The full repository gate is left to integration.

Linux counts.md is refreshed: 761 recorded fixtures before, 768 after. Every new
fixture's allocations equal its frees. Existing NUL-diagnostic counts change
because the message is built as a UTF-16-preserving counted string. The directory
system fixture also observes the new files in this checkout. The precise rows and
all before/after changes are in internal/oracle/counts.md; other rows are unchanged.

Logs and all build output stay outside the checkout under /tmp/host-fs-inspect/:
focus2.log, fs-regression-final.log, native.log, vet-final.log and counts-final.log.
The first focused run preceded the final multiline port and exposed a relative
path mistake in the new mutant harness; both were corrected. The first existing
fs regression run exposed the stale Date mutant described above. Neither failed
run is counted as final validation.
