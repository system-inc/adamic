Built 12 .a probes, four-way byte comparisons, a short-I/O shim, and five isolated runtime mutants; production code is unchanged.
Base 031a1259bc7973934792dc6cb1bd4074fc2204b9; programs commit 9edc1472; reviewed 73352e87..5a2681b1.
Node v24.19.0, sanitized native, native -O2, and backend Node: 9 probes agree; 3 expose two error-formatting issues; all executions exit 0 with empty stderr.
Five mutants: all new witnesses detect them; four survive focused package/oracle checks; the UTF-16 mutant fails existing native and oracle tests; broader package results below.
Not covered: enormous allocations, injected syscall failures, concurrency/races, non-Linux hosts, WASI execution, and the complete repository gate and full native-package mutant completion.

## Scope and reproduction

Read CLAUDE.md, README.md, docs/0.1.md, docs/memory.md, all five requested implementation files, and their patches in the requested commit range. Started with `git fetch origin` and `git checkout -b codex/coverage-oct8-nodehost origin/main`. There are no production fixes in this branch. Temporary runtime mutations were restored after each experiment.

Run from the repository root, with `source /workspace/adamic-tools/env.sh`. `bash cloud/setup.sh` succeeded. Timing lines: go ready 0.296s; node ready 0.388s; clang ready 1.001s; markdown dependencies installed in 1.475s and ready at 1.993s; submodules ready 28.000s; go build ready 356.610s; test binaries deferred 356.738s; build cache warm 356.740s; done 356.771s. `nproc` printed 5; cpu.max was `400000 100000`. Go 1.27.1, clang 20.1.8. Raw setup output is in evidence/setup.log.

Initial compilation found missing pinned @types/node 25.3.3. `npm ci --prefix stage3/api --ignore-scripts` installed 3 packages in 748ms and changed no tracked dependency files. Two initial probe designs using dynamic Array.from inside try were explicitly refused; another empty-array construction inferred never[] and was refused. Final probes use supported construction, compile successfully, and remain .a files. A later full-package invocation without sourcing env failed with `Go: Unknown option: test`; it was rerun with env sourced. These setup failures are not runtime disagreement findings.

Commands (each command's stdout and stderr went directly to files):

```sh
source /workspace/adamic-tools/env.sh
go build -o /tmp/nodehost-adamic ./cmd/adamic
python3 notes/coverage-oct8/nodehost/run.py --logs /tmp/nodehost-evidence
clang -shared -fPIC -std=c11 -Wall -Wextra -Werror -pedantic notes/coverage-oct8/nodehost/short_io.c -ldl -o /tmp/nodehost-short-io.so
python3 notes/coverage-oct8/nodehost/run.py --logs /tmp/nodehost-short-final --io-shim /tmp/nodehost-short-io.so fs_large_roundtrip.a
python3 notes/coverage-oct8/nodehost/mutants.py
python3 notes/coverage-oct8/nodehost/full_packages.py
go vet ./...
```

run.py uses source Node through oracle/node.mjs, `adamic build --sanitize`, `adamic build` (default -O2), then `adamic js` through oracle/node.mjs. Sanitizer flags are the oracle's native flags: -O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all. ASAN_OPTIONS includes detect_leaks=1:halt_on_error=1; UBSAN_OPTIONS includes halt_on_error=1. Equality compares raw stdout, stderr, and exit files. JSON copies are readable indexes, not the byte comparator. The runner reports disagreements without failing its own process; inspect results.json or the printed AGREE/DIFFER lines. All final build statuses are 0, all execution statuses are 0, and all execution stderr files are empty. No abort, C compilation failure, or sanitizer-reported memory leak was observed in final baseline probes.

All actual filesystem writes occur within directories created by mkdtempSync under tmpdir; finally removes them. fs_error_preview uses only rejected NUL arguments, so it never creates its otherwise-relative prefix. Programs normalize random temporary path names themselves before printing; runner output is otherwise unmodified.

## Programs and observations

| Program | Cases | Four-way result |
| --- | --- | --- |
| buffer_windows.a | 130 start/end windows, negative/fractional/NaN/infinite/large offsets, hex/base64/UTF-8/UTF-16, truncated sequences, odd UTF-16 lengths, surrogate pairs and lone surrogates | Agree |
| buffer_copy_lifetime.a | 80 dynamic allocations, alias mutation, independent Buffer.from copy, reassignments, saved decoded strings, captured buffers | Agree |
| fs_partial.a | Sentinel windows, short EOF reads, positioned/current fd offsets, fd readFile rest/EOF, string UTF-8 writes, fractional position, append mode, zero reads | Agree |
| fs_read_validation.a | Offset/length/position/fd validation matrix, empty buffers and competing errors | Different numeric formatting in two position errors |
| fs_numeric_error.a | Minimal large offset/position and ordinary large length errors | Different numeric formatting |
| fs_error_preview.a | NUL-path and prefix inspect previews, lone surrogates, U+0085, U+2028, controls, competing quotes, 128-unit truncation | Different inspect formatting |
| fs_fd_ownership.a | Caller descriptors after EBADF, owned directory-read errors, owned write errors, 50 cleanup iterations and descriptor reuse | Agree |
| fs_large_roundtrip.a | 0/1/4095/4096/4097/8192/8193/20000 bytes, grow boundaries, buffer copies, positioned sentinel tails, current fd reads | Agree, including short-I/O shim |
| fs_write_positions.a | Negative, nonfinite, fractional and 2^63 positions; lone surrogate replacement, embedded NUL data, descriptor current-position writes | Agree |
| fs_open_precedence.a | NUL vs missing paths, invalid flags and invalid modes, error precedence, existsSync NUL | Agree |
| fs_zero_read.a | Zero-length read before empty-buffer or invalid-fd errors | Agree |
| host_directory.a | cwd snapshots, cache invalidation, chdir errors, lexical join containing canceled NUL component | Agree |

The shim caps successful regular-file reads at 257 bytes and writes at 3 bytes, leaving standard output/error alone. Original code still agrees in all four modes. A mutant that stops write_data after the first successful write produces files of only 3 bytes, demonstrating that this experiment reaches continued-write handling. This is successful short I/O, not EINTR or a failed syscall.

## Disagreement 1: large numbers in errors

Location: internal/native/runtime/node_fs_file.c:130-148, called by readSync validation at :289-290. Kind: different output.

Observation: fs_numeric_error.a and fs_read_validation.a finish cleanly in all modes. Native omits Node's `_` grouping for the received large offset/position. The ordinary large length message agrees. Error name and, in the matrix probe, code agree.

Inference from source: integer/read_range use adamic_string_from_number rather than Node's inspected-number rendering. The outputs establish the mismatch; no claim about other numeric errors follows.

The minimal program is fs_numeric_error.a. Verbatim four outputs follow; raw files for the larger matrix are in evidence/baseline/fs_read_validation.*.

### fs_numeric_error: node

stdout:

```text
RangeError|The value of "offset" is out of range. It must be >= 0 && <= 9007199254740991. Received 9_007_199_254_740_992
RangeError|The value of "position" is out of range. It must be >= -1 && <= 9007199254740991. Received 9_007_199_254_740_992
RangeError|The value of "length" is out of range. It must be <= 2. Received 2147483647
```

stderr: empty (0 bytes). Exit: 0.

### fs_numeric_error: sanitize

stdout:

```text
RangeError|The value of "offset" is out of range. It must be >= 0 && <= 9007199254740991. Received 9007199254740992
RangeError|The value of "position" is out of range. It must be >= -1 && <= 9007199254740991. Received 9007199254740992
RangeError|The value of "length" is out of range. It must be <= 2. Received 2147483647
```

stderr: empty (0 bytes). Exit: 0.

### fs_numeric_error: release

stdout:

```text
RangeError|The value of "offset" is out of range. It must be >= 0 && <= 9007199254740991. Received 9007199254740992
RangeError|The value of "position" is out of range. It must be >= -1 && <= 9007199254740991. Received 9007199254740992
RangeError|The value of "length" is out of range. It must be <= 2. Received 2147483647
```

stderr: empty (0 bytes). Exit: 0.

### fs_numeric_error: javascript

stdout:

```text
RangeError|The value of "offset" is out of range. It must be >= 0 && <= 9007199254740991. Received 9_007_199_254_740_992
RangeError|The value of "position" is out of range. It must be >= -1 && <= 9007199254740991. Received 9_007_199_254_740_992
RangeError|The value of "length" is out of range. It must be <= 2. Received 2147483647
```

stderr: empty (0 bytes). Exit: 0.

## Disagreement 2: NUL argument inspect previews

Location: internal/native/runtime/node_fs_file.c:86-109. Kind: different output; comment does not match code.

Observation: fs_error_preview.a prints the same TypeError and ERR_INVALID_ARG_VALUE in all four modes, but native turns each lone surrogate into three replacement characters, leaves U+0085 literal, and uses \v where Node uses \x0B. U+2028, quote selection, and the tested long preview agree. All four finish with exit 0 and empty stderr.

Inference from source: byte-by-byte escape selection and subsequent UTF-8 decoding explain these differences; this is not a complete implementation of Node inspect, despite the comment at :86. The program is fs_error_preview.a. Verbatim outputs follow. U+0085 and U+2028 are literal Unicode in the raw output files; JSON indexes preserve an escaped representation for clarity.

### fs_error_preview: node

stdout:

```text
TypeError|ERR_INVALID_ARG_VALUE|The argument 'path' must be a string, Uint8Array, or URL without null bytes. Received 'x\ud800\x00'
TypeError|ERR_INVALID_ARG_VALUE|The argument 'prefix' must be a string, Uint8Array, or URL without null bytes. Received 'x\ud800\x00'
TypeError|ERR_INVALID_ARG_VALUE|The argument 'path' must be a string, Uint8Array, or URL without null bytes. Received 'x\udfff\x00'
TypeError|ERR_INVALID_ARG_VALUE|The argument 'prefix' must be a string, Uint8Array, or URL without null bytes. Received 'x\udfff\x00'
TypeError|ERR_INVALID_ARG_VALUE|The argument 'path' must be a string, Uint8Array, or URL without null bytes. Received 'x\x85\x00'
TypeError|ERR_INVALID_ARG_VALUE|The argument 'prefix' must be a string, Uint8Array, or URL without null bytes. Received 'x\x85\x00'
TypeError|ERR_INVALID_ARG_VALUE|The argument 'path' must be a string, Uint8Array, or URL without null bytes. Received 'x \x00'
TypeError|ERR_INVALID_ARG_VALUE|The argument 'prefix' must be a string, Uint8Array, or URL without null bytes. Received 'x \x00'
TypeError|ERR_INVALID_ARG_VALUE|The argument 'path' must be a string, Uint8Array, or URL without null bytes. Received '\x00\n\t\b\x0B\f\r'
TypeError|ERR_INVALID_ARG_VALUE|The argument 'prefix' must be a string, Uint8Array, or URL without null bytes. Received '\x00\n\t\b\x0B\f\r'
TypeError|ERR_INVALID_ARG_VALUE|The argument 'path' must be a string, Uint8Array, or URL without null bytes. Received '\'"`$\x00'
TypeError|ERR_INVALID_ARG_VALUE|The argument 'prefix' must be a string, Uint8Array, or URL without null bytes. Received '\'"`$\x00'
TypeError|ERR_INVALID_ARG_VALUE|The argument 'path' must be a string, Uint8Array, or URL without null bytes. Received 'aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa...
TypeError|ERR_INVALID_ARG_VALUE|The argument 'prefix' must be a string, Uint8Array, or URL without null bytes. Received 'aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa...
```

stderr: empty (0 bytes). Exit: 0.

### fs_error_preview: sanitize

stdout:

```text
TypeError|ERR_INVALID_ARG_VALUE|The argument 'path' must be a string, Uint8Array, or URL without null bytes. Received 'x���\x00'
TypeError|ERR_INVALID_ARG_VALUE|The argument 'prefix' must be a string, Uint8Array, or URL without null bytes. Received 'x���\x00'
TypeError|ERR_INVALID_ARG_VALUE|The argument 'path' must be a string, Uint8Array, or URL without null bytes. Received 'x���\x00'
TypeError|ERR_INVALID_ARG_VALUE|The argument 'prefix' must be a string, Uint8Array, or URL without null bytes. Received 'x���\x00'
TypeError|ERR_INVALID_ARG_VALUE|The argument 'path' must be a string, Uint8Array, or URL without null bytes. Received 'x\x00'
TypeError|ERR_INVALID_ARG_VALUE|The argument 'prefix' must be a string, Uint8Array, or URL without null bytes. Received 'x\x00'
TypeError|ERR_INVALID_ARG_VALUE|The argument 'path' must be a string, Uint8Array, or URL without null bytes. Received 'x \x00'
TypeError|ERR_INVALID_ARG_VALUE|The argument 'prefix' must be a string, Uint8Array, or URL without null bytes. Received 'x \x00'
TypeError|ERR_INVALID_ARG_VALUE|The argument 'path' must be a string, Uint8Array, or URL without null bytes. Received '\x00\n\t\b\v\f\r'
TypeError|ERR_INVALID_ARG_VALUE|The argument 'prefix' must be a string, Uint8Array, or URL without null bytes. Received '\x00\n\t\b\v\f\r'
TypeError|ERR_INVALID_ARG_VALUE|The argument 'path' must be a string, Uint8Array, or URL without null bytes. Received '\'"`$\x00'
TypeError|ERR_INVALID_ARG_VALUE|The argument 'prefix' must be a string, Uint8Array, or URL without null bytes. Received '\'"`$\x00'
TypeError|ERR_INVALID_ARG_VALUE|The argument 'path' must be a string, Uint8Array, or URL without null bytes. Received 'aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa...
TypeError|ERR_INVALID_ARG_VALUE|The argument 'prefix' must be a string, Uint8Array, or URL without null bytes. Received 'aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa...
```

stderr: empty (0 bytes). Exit: 0.

### fs_error_preview: release

stdout:

```text
TypeError|ERR_INVALID_ARG_VALUE|The argument 'path' must be a string, Uint8Array, or URL without null bytes. Received 'x���\x00'
TypeError|ERR_INVALID_ARG_VALUE|The argument 'prefix' must be a string, Uint8Array, or URL without null bytes. Received 'x���\x00'
TypeError|ERR_INVALID_ARG_VALUE|The argument 'path' must be a string, Uint8Array, or URL without null bytes. Received 'x���\x00'
TypeError|ERR_INVALID_ARG_VALUE|The argument 'prefix' must be a string, Uint8Array, or URL without null bytes. Received 'x���\x00'
TypeError|ERR_INVALID_ARG_VALUE|The argument 'path' must be a string, Uint8Array, or URL without null bytes. Received 'x\x00'
TypeError|ERR_INVALID_ARG_VALUE|The argument 'prefix' must be a string, Uint8Array, or URL without null bytes. Received 'x\x00'
TypeError|ERR_INVALID_ARG_VALUE|The argument 'path' must be a string, Uint8Array, or URL without null bytes. Received 'x \x00'
TypeError|ERR_INVALID_ARG_VALUE|The argument 'prefix' must be a string, Uint8Array, or URL without null bytes. Received 'x \x00'
TypeError|ERR_INVALID_ARG_VALUE|The argument 'path' must be a string, Uint8Array, or URL without null bytes. Received '\x00\n\t\b\v\f\r'
TypeError|ERR_INVALID_ARG_VALUE|The argument 'prefix' must be a string, Uint8Array, or URL without null bytes. Received '\x00\n\t\b\v\f\r'
TypeError|ERR_INVALID_ARG_VALUE|The argument 'path' must be a string, Uint8Array, or URL without null bytes. Received '\'"`$\x00'
TypeError|ERR_INVALID_ARG_VALUE|The argument 'prefix' must be a string, Uint8Array, or URL without null bytes. Received '\'"`$\x00'
TypeError|ERR_INVALID_ARG_VALUE|The argument 'path' must be a string, Uint8Array, or URL without null bytes. Received 'aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa...
TypeError|ERR_INVALID_ARG_VALUE|The argument 'prefix' must be a string, Uint8Array, or URL without null bytes. Received 'aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa...
```

stderr: empty (0 bytes). Exit: 0.

### fs_error_preview: javascript

stdout:

```text
TypeError|ERR_INVALID_ARG_VALUE|The argument 'path' must be a string, Uint8Array, or URL without null bytes. Received 'x\ud800\x00'
TypeError|ERR_INVALID_ARG_VALUE|The argument 'prefix' must be a string, Uint8Array, or URL without null bytes. Received 'x\ud800\x00'
TypeError|ERR_INVALID_ARG_VALUE|The argument 'path' must be a string, Uint8Array, or URL without null bytes. Received 'x\udfff\x00'
TypeError|ERR_INVALID_ARG_VALUE|The argument 'prefix' must be a string, Uint8Array, or URL without null bytes. Received 'x\udfff\x00'
TypeError|ERR_INVALID_ARG_VALUE|The argument 'path' must be a string, Uint8Array, or URL without null bytes. Received 'x\x85\x00'
TypeError|ERR_INVALID_ARG_VALUE|The argument 'prefix' must be a string, Uint8Array, or URL without null bytes. Received 'x\x85\x00'
TypeError|ERR_INVALID_ARG_VALUE|The argument 'path' must be a string, Uint8Array, or URL without null bytes. Received 'x \x00'
TypeError|ERR_INVALID_ARG_VALUE|The argument 'prefix' must be a string, Uint8Array, or URL without null bytes. Received 'x \x00'
TypeError|ERR_INVALID_ARG_VALUE|The argument 'path' must be a string, Uint8Array, or URL without null bytes. Received '\x00\n\t\b\x0B\f\r'
TypeError|ERR_INVALID_ARG_VALUE|The argument 'prefix' must be a string, Uint8Array, or URL without null bytes. Received '\x00\n\t\b\x0B\f\r'
TypeError|ERR_INVALID_ARG_VALUE|The argument 'path' must be a string, Uint8Array, or URL without null bytes. Received '\'"`$\x00'
TypeError|ERR_INVALID_ARG_VALUE|The argument 'prefix' must be a string, Uint8Array, or URL without null bytes. Received '\'"`$\x00'
TypeError|ERR_INVALID_ARG_VALUE|The argument 'path' must be a string, Uint8Array, or URL without null bytes. Received 'aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa...
TypeError|ERR_INVALID_ARG_VALUE|The argument 'prefix' must be a string, Uint8Array, or URL without null bytes. Received 'aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa...
```

stderr: empty (0 bytes). Exit: 0.

## Guard mutation observations

Focused command, run unmutated and once for each mutant with ADAMIC_GATE_UNCACHED=1:

```sh
go test ./internal/lower ./internal/native ./internal/oracle -run 'TestNodeFSFile|TestNodeBuffer|TestNativeAgreesWithNode/internal/oracle/testdata/(node_buffer_|closure_convention_host24)' -count=1 -timeout 30m -v
```

Unmutated result: lower 9.738s, native 1.507s, oracle 18.482s, all passed. Each mutant is applied alone. Exact mutations, package statuses, all four witness outputs, build statuses, and error streams are in evidence/mutant-summary.json and evidence/mutants/<name>/results.json. Raw stdout/stderr/exit files alongside those JSON files are the verbatim outputs for each disagreement. The new source Node and backend outputs remain the correct baseline; both mutant native modes disagree and complete cleanly. No runtime abort or memory sanitizer report was observed.

| Mutant | Location and modification | Focused package result | New witness and observed difference |
| --- | --- | --- | --- |
| read-zero-return | node_fs_file.c:291, remove length == 0 return | Pass, survives | fs_zero_read.a: native throws empty-buffer TypeError or invalid-fd RangeError; Node returns zero |
| owned-buffer-close | node_fs_file.c:255, remove owned close in read_buffer only | Pass, survives | fs_fd_ownership.a: native fd-reused false vs true |
| cwd-invalidate | node_host.c:197, remove cached cwd release/reset after chdir | Pass, survives | host_directory.a: native saved/current cwd comparisons false and stale chdir error directory |
| short-write-loop | node_fs_file.c:335, break after first successful write_data write | Pass, survives | fs_large_roundtrip.a plus shim: native 3-byte files vs full buffers |
| utf16-pair-guard | node_buffer.c:248, replace low-surrogate guard with false | Fail, caught | buffer_windows.a: native paired point 55357 vs 128781 |

Existing UTF-16 catchers: TestNodeBufferRuntimeWithoutDeclarations, and TestNativeAgreesWithNode for node_buffer_random.a, node_buffer_bom.a, node_buffer_encodings.a. Lowering passes because no lowering code is changed.

Observation for owned-buffer-close: descriptor recycling changes after repeated owned-reader failures. Inference: removing close leaks file descriptors, supported by the exact source mutation. This is a descriptor resource leak, not a sanitizer-observed heap leak. No baseline descriptor leak was observed by this witness.

A broader run was attempted for read-zero-return using `ADAMIC_GATE_UNCACHED=1 go test ./internal/lower ./internal/native -count=1 -timeout 30m`. Lower passed in 46.616s; native was interrupted after more than five minutes while still running unrelated runtime suites. There is no final native-package result. The other three correctly configured full-package runs were not attempted after bounding this extra check. Their retained full-package logs contain only the earlier unsourced-toolchain launch failures, not test results. See evidence/mutant-full-packages.json. An interrupted run is not evidence that a mutant was caught or survived. A surviving mutant only supports a coverage gap in the tested selection; it does not establish that every repository test misses it. The entire oracle suite was not run.

## Comment and documentation audit

The requested files contain inline comments, not separate API documentation. All comment groups were inspected against the implementing code; observations from programs are distinguished here from static agreement. Node is the behavioral oracle, not an Adamic comment.

| File and lines | Audit |
| --- | --- |
| node_fs_file.c:1-2 | Ownership description agrees statically and with fd/copy probes: caller offsets live in OS descriptors; temporary allocations and returned counted values belong to runtime. |
| node_fs_file.c:70-71 | Lone-surrogate data encoding and existsSync NUL behavior agree with probes. NUL errors precede filesystem calls in source. |
| node_fs_file.c:86 | Mismatch: intended inspect preview is implemented incompletely; disagreement 2 supplies runtime evidence. Tested 128-unit truncation agrees. |
| node_fs_file.c:162,246 | Embedded NUL flags fail; already-open fd ignores flag. Code agrees; flag/NUL matrix and existing ownership/options fixtures support it. |
| node_fs_file.c:316-317 | Current-position behavior for negative/fractional/nonfinite positions agrees with fs_write_positions.a. |
| node_fs_file.c:339-340 | Static agreement: owned descriptors close after write/fsync error and close error supersedes prior error. Owned write failure exercised; failed fsync/close not injected. |
| node_fs_file.c:453-454 | Static agreement: recursive mkdir only recurses on ENOENT and returns first created spelling. No new mkdir guard mutation. |
| node_fs_file.c:502-503,518 | Static agreement: WASI six-character random suffix retries on EEXIST; libc mkdtemp used elsewhere. Temp creation exercised on Linux; WASI not executed. |
| node_fs_file.c:537,550 | Static agreement: rm uses lstat/error handling and symlink leaves. New probes exercise cleanup but do not construct symlinks. |
| node_fs_file.c:628 | Static agreement with Linux seconds-to-timespec conversion; no new utimes boundary probe. |
| node_buffer.c:1-6 | UTF-8 rejected-continuation reprocessing and incomplete EOF replacement agree with tested windows. Installed Node is the stated version. Attribution/license statements reference THIRD_PARTY_NOTICES.md; provenance was not independently established against upstream source history. |
| node_buffer.c:12-38 | Byte-class row ranges and state-offset comments agree with table positions and decoder indexing. Static audit. |
| node_buffer.c:66 | Three-byte replacement bound agrees with decoder operations; integer-overflow/huge-allocation boundary not exercised. |
| node_buffer.c:120 | Null out-of-range slots discard typed-array writes; source agrees and existing write fixtures exercise this behavior. |
| node_buffer.c:175-177 | Decoder narrows units, skips invalid input, stops at padding, emits after 2/3/4 sextets; source agrees, existing encoding fixtures run. |
| node_buffer.c:241-243 | Pairing only adjacent high/low surrogates agrees with new probe; false-guard mutant is caught by old and new tests. |
| library_node_buffer.go:12-13 | Shared loader supplies pinned declarations; recognizer checks a /@types/node/ path substring plus enclosing module. Static agreement within loader assumption; no canonical-path identity test added. |
| library_node_buffer.go:46-47 | Applies to Buffer and Hash host-member refusal; Hash is handled in this same function, so reference to its private layout is consistent. Existing refusal tests run. |
| library_node_fs_file.go:10 | Symbol-based imported declaration recognition supports aliases; existing alias/options fixture run. |
| library_node_fs_file.go:46 | Overbroad comment: "Void calls are only admitted as discarded statements." At :101 the code also admits a call returned from a void function; existing TestNodeFSFileOptionsBorrow exercises return unlinkSync(path). Kind: comment does not match code. No runtime disagreement asserted. |
| library_node_fs_file.go:115-116 | Unsupported shapes produce NotYet; existing refusal tests and initial unsupported probe refusals support this. |
| library_node_fs_file.go:148-149 | Options splitting rejects unsafe side effects rather than reorder/drop them; code and existing refusal tests agree. |
| library_node_fs_file.go:385 | Distinct fd intrinsics preserve caller ownership; observed in new fd-error probe. |
| library_node_fs_file.go:432-433 | Options synchronously become scalar parameters without retaining/writing option objects; source agrees and existing borrowing/refusal tests run. |
| node_host.c:1-2 | Header calls the file node_path.c and describes only lexical paths with no filesystem lookup. Whole file also calls getcwd/chdir and provides tmpdir; header filename/scope is stale. Kind: comment does not match code. Lexical join itself agrees with the NUL cancellation probe. |
| node_host.c:31-32 | Stack marks normal/unresolved parent component starts; implementation agrees statically and lexical join result agrees. |

Comment discrepancies have no additional four-way output disagreement beyond the linked probes. They are source observations, not inferred compiler failures.

## Limits

These probes extend fixtures; they do not exhaust input space. No claims about ENOSPC, EINTR, read/write returning zero unexpectedly, EIO, fsync/close failure precedence under injection, allocation failure/size overflow, filesystem races, permissions, pipes, symlink cleanup safety on hostile trees, Windows/macOS or WASI behavior, or all utimes/stat edge cases. Only the selected oracle cases and native/lower packages were tested; no full repository gate. Successful sanitizer runs detect tested memory misuse/leaks, not prove absence of all ownership bugs. Node formatting is observed on pinned Node v24.19.0; newer Node changes are outside this report.

## Final verification and evidence index

`go vet ./...` passed with empty output. Production diffs are empty after restoring mutations. Formatting and git diff checks are recorded in the evidence directory. The full repository gate was not run.

The raw output files below are verbatim, including literal Unicode, and each has a corresponding .stderr and .exit file. All baseline stderr files are empty and exit files contain 0.

- fs_numeric_error.a: [node stdout](evidence/baseline/fs_numeric_error.node.stdout), [sanitize stdout](evidence/baseline/fs_numeric_error.sanitize.stdout), [release stdout](evidence/baseline/fs_numeric_error.release.stdout), [javascript stdout](evidence/baseline/fs_numeric_error.javascript.stdout); [all statuses](evidence/baseline/results.json).
- fs_read_validation.a: [node stdout](evidence/baseline/fs_read_validation.node.stdout), [sanitize stdout](evidence/baseline/fs_read_validation.sanitize.stdout), [release stdout](evidence/baseline/fs_read_validation.release.stdout), [javascript stdout](evidence/baseline/fs_read_validation.javascript.stdout); [all statuses](evidence/baseline/results.json).
- fs_error_preview.a: [node stdout](evidence/baseline/fs_error_preview.node.stdout), [sanitize stdout](evidence/baseline/fs_error_preview.sanitize.stdout), [release stdout](evidence/baseline/fs_error_preview.release.stdout), [javascript stdout](evidence/baseline/fs_error_preview.javascript.stdout); [all statuses](evidence/baseline/results.json).
- read-zero-return: [node stdout](evidence/mutants/read-zero-return/fs_zero_read.node.stdout), [sanitize stdout](evidence/mutants/read-zero-return/fs_zero_read.sanitize.stdout), [release stdout](evidence/mutants/read-zero-return/fs_zero_read.release.stdout), [javascript stdout](evidence/mutants/read-zero-return/fs_zero_read.javascript.stdout); [all statuses](evidence/mutants/read-zero-return/results.json).
- owned-buffer-close: [node stdout](evidence/mutants/owned-buffer-close/fs_fd_ownership.node.stdout), [sanitize stdout](evidence/mutants/owned-buffer-close/fs_fd_ownership.sanitize.stdout), [release stdout](evidence/mutants/owned-buffer-close/fs_fd_ownership.release.stdout), [javascript stdout](evidence/mutants/owned-buffer-close/fs_fd_ownership.javascript.stdout); [all statuses](evidence/mutants/owned-buffer-close/results.json).
- cwd-invalidate: [node stdout](evidence/mutants/cwd-invalidate/host_directory.node.stdout), [sanitize stdout](evidence/mutants/cwd-invalidate/host_directory.sanitize.stdout), [release stdout](evidence/mutants/cwd-invalidate/host_directory.release.stdout), [javascript stdout](evidence/mutants/cwd-invalidate/host_directory.javascript.stdout); [all statuses](evidence/mutants/cwd-invalidate/results.json).
- short-write-loop: [node stdout](evidence/mutants/short-write-loop/fs_large_roundtrip.node.stdout), [sanitize stdout](evidence/mutants/short-write-loop/fs_large_roundtrip.sanitize.stdout), [release stdout](evidence/mutants/short-write-loop/fs_large_roundtrip.release.stdout), [javascript stdout](evidence/mutants/short-write-loop/fs_large_roundtrip.javascript.stdout); [all statuses](evidence/mutants/short-write-loop/results.json).
- utf16-pair-guard: [node stdout](evidence/mutants/utf16-pair-guard/buffer_windows.node.stdout), [sanitize stdout](evidence/mutants/utf16-pair-guard/buffer_windows.sanitize.stdout), [release stdout](evidence/mutants/utf16-pair-guard/buffer_windows.release.stdout), [javascript stdout](evidence/mutants/utf16-pair-guard/buffer_windows.javascript.stdout); [all statuses](evidence/mutants/utf16-pair-guard/results.json).
