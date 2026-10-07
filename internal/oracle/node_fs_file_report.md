Built UTF-8 synchronous node:fs, catchable Node-coded errors, Stats and seven System wrappers; Buffer-dependent reads remain blocked by the absent Buffer type.
Branch codex/host-fs-file starts at ef3d907ecdc4c771b016f7d9c52372def057a340; delivery commit is recorded in the final handoff.
Node v24.19.0 differential gate passed all 12 new fixtures on both backends, existing input fixtures, sanitizers, leak checks and recorded counts.
All 25 behavioral mutants were caught only by Node stdout; three additional guard mutants failed their designated checks.
Not covered: readSync, Buffer-returning readFileSync, System.readFile BOM decoding, require wiring, complete native tsc proof, macOS execution or the full repository gate.

The unit owns node_fs_file.d.ts, its embedded loader, library_node_fs_file.go on both backends, the NodeFSFile IR, node_fs_file.c/.h, System source and the directory-aware oracle. Small dispatch hooks connect these files to existing loading, lowering, exception flow, emission and runtime headers. No edits were made to internal/native/emit.go, internal/lower/lower.go, internal/native/native.go or internal/oracle/oracle_test.go.

Contracts were read from origin/codex/tsc-census stage3/census/REPORT.md, data/system_contract.json and data/node_derived_sites.json, and checked against TypeScript getNodeSystem. Selected declarations originate in @types/node 25.3.3; path and encoding aliases are restricted to implemented string/UTF-8 shapes and have unit-prefixed names to avoid alias collisions when node:fs declarations merge. Unsupported numeric flags, mode strings, other encodings, bigint Stats, timestamp strings, optional or dynamic options, byte views and URL paths are explicitly refused. Date support is numeric construction, getTime/valueOf and the timestamps returned by Stats, rather than a general Date host.

The implemented System functions are writeFile, fileExists, getFileSize, getModifiedTime, setModifiedTime, deleteFile and createDirectory in internal/load/node_fs_file_system.a. They preserve swallowed failures, write cleanup/BOM emission, file-only sizes and EEXIST handling. They are reusable source functions, not a replacement ts.sys object. Buffer has no type or native byte-view representation on this base. Returning a string from the raw-byte overload would silently miscompile sys.ts's UTF-16BE/LE and UTF-8 BOM handling, so that overload and System.readFile are absent. The runtime exposes adamic_fs_file_read_bytes for the Buffer unit to integrate, and does not fabricate a Buffer. readSync likewise awaits that byte-view integration. Module declarations merge; require('fs') resolution remains another unit's responsibility.

Observations from Node included positioned writes preserving descriptor offset, fractional/non-finite write positions using current offset, flags validated before mode, stat following symbolic links, missing-stat suppression for ENOENT and ENOTDIR only, Stats Date rounding distinct from mtimeMs, nanosecond utimes precision, negative Date timestamps distinct from negative numeric timestamps, and invalid Dates preserving timestamps. The Linux fixtures exercise these effects and ENOENT, ENOTDIR, EEXIST, EACCES, EISDIR, EBADF and invalid NUL paths. Error objects carry name, message and code; errno/syscall/path properties and full Node exception stacks are outside the requested surface.

Each filesystem run gets real input directories, writable files, symlinks, a dangling link and inaccessible directories. Source runs on Node, emitted JavaScript on Node and native code under ASan/UBSan are compared byte for byte. Separate runs check leaks and filesystem bytes, names and permissions. Root test processes drop child execution to uid/gid 65534. Every new allocation-count row balances allocations and frees. The fs harness supplements the existing input harness without changing the generic oracle fixture list.

Shared plain const options are synchronously borrowed at verified node:fs call arguments only. Wider mutable assignments still fail Adamic's invariance rule. Literal options containing evaluated expressions are refused because flattening them could reorder or drop effects. Void return expressions are evaluated before returning, as sys.ts's deleteFile requires. Using an fs void call as a value is explicitly refused rather than exposing the C implementation's numeric sentinel. All runtime-created shapes participate in the native field-layout proof so Stats.size cannot alias the existing input host's size slot.

Validation commands (all test output redirected to logs, never piped):

- `bash cloud/setup.sh > /tmp/fs_file_setup.log 2>&1`: Go 1.27.1, clang 20.1.8, Node v24.19.0; Go/clang/Node/submodule readiness 0s each, build-cache warm 100s, done 100s. `nproc`: 5. cgroup cpu.max: 400000 100000. Environment: `source /workspace/adamic-tools/env.sh`.
- `go test ./internal/load ./internal/lower ./internal/native ./internal/flow ./internal/javascript -count=1 -timeout 15m > /tmp/fs_file_packages_final.log 2>&1`: PASS: load 2.038s, lower 19.428s, native 98.865s, flow 65.995s; JavaScript has no standalone package tests and is checked by the oracle.
- `go test ./internal/oracle -run 'TestNodeFSFile|TestInputAgreesWithNode|TestCountsAreRecorded' -count=1 -v -timeout 30m -args -update-counts > /tmp/fs_file_gate_final.log 2>&1`: PASS, oracle 20.426s. Includes the final negative/invalid/mixed Date and invalid flag/path fixtures.
- `go test ./internal/lower -run TestNodeFSFile -count=1 > /tmp/fs_file_lower_checks.log 2>&1`: PASS, 0.484s after restoring the option-effect guard mutant.
- `go test ./internal/load ./internal/lower -count=1 > /tmp/fs_file_load_lower_final.log 2>&1`: PASS, load 0.693s, lower 8.701s after the final void-value guard and declaration alias changes.
- `go test ./internal/oracle -run 'TestNodeFSFile|TestInputAgreesWithNode|TestCountsAreRecorded' -count=1 -timeout 30m > /tmp/fs_file_delivery_gate.log 2>&1`: PASS, 9.757s with the final source and recorded counts checked without rewriting.
- `git diff --check`: clean.

Behavioral mutants and their sole detector, Node stdout comparison in TestNodeFSFileMutants:

| Member | Mutation |
|---|---|
| readFileSync UTF-8 | Append an exclamation mark to the result |
| readFileSync fd | Append an exclamation mark to the descriptor result |
| openSync | Replace truncate flag w with append a |
| writeSync | Add one to the returned byte count |
| closeSync | Swallow a closed-descriptor error |
| writeFileSync | Replace append a with truncate w |
| existsSync | Return false for existing directories |
| statSync | Ignore throwIfNoEntry and suppress missing errors |
| Stats.size | Add one |
| Stats.mtimeMs | Add one millisecond |
| Stats.mtime | Add one millisecond to Date |
| Stats.isFile | Invert the answer |
| Stats.isDirectory | Invert the answer |
| Stats.isSymbolicLink | Return true despite stat following links |
| mkdirSync | Discard the first-created-directory result |
| unlinkSync | Swallow unlink errors |
| utimesSync | Add one second to mtime |
| Date.getTime/valueOf | Add one millisecond |
| System.writeFile | Drop the BOM before writing |
| System.fileExists | Treat every Stats value as not a file |
| System.getFileSize | Add one to the underlying size |
| System.getModifiedTime | Add one millisecond to returned Date |
| System.setModifiedTime | Write a Date one second too late |
| System.deleteFile | Skip unlink entirely |
| System.createDirectory | Swallow mkdir failure, including EACCES |


Additional mutants:

- Remove the void-value guard: TestNodeFSFileRefusesVoidValues fails with `want void value refusal, got <nil>`; exit 1, /tmp/fs_file_void_mutant.log. Restored and the fs lowering checks rerun.

- Remove the option-literal effect guard: TestNodeFSFileRefusesOptionEffects fails with `want effects refusal, got <nil>`; exit 1, /tmp/fs_file_effect_mutant.log. Restored and tested.
- Remove the fs runtime shapes from the field-layout proof: TestRuntimeFieldLayoutsAreIncluded fails because runtime field code at slot 2 is absent/conflicting; exit 1, /tmp/fs_file_layout_mutant.log. Restored before final package validation.

Linux is the gate of record. macOS uses st_mtimespec rather than st_mtim, and libuv's fsync uses F_FULLFSYNC where this runtime currently uses fsync; flush behavior is therefore not claimed identical on macOS. Directory sizes, permissions, timestamp precision and errno availability depend on filesystem/platform. Resource exhaustion, interrupted close/fsync, concurrent deletion, multi-gigabyte reads and the complete libuv errno catalog were not exhaustively tested. The errno table covers the common filesystem failures exercised here; unknown platform-specific errno values currently fall back to UNKNOWN. Invalid-value inspection covers the exercised short quoting/control/truncation forms, but unusual Unicode inspection and lone-surrogate previews are not exhaustively matched. These are remaining fidelity limits, not evidence of complete Node emulation.
