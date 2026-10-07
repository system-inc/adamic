Freshness fix pushed: 5320b144f9142ca76a9472a61308c5096aea70e2 on codex/host-fs-fresh; included as 34e4aaa on the rebased fs landing branch.
Shared loader landing pushed: 69c71d5171801bc355400452a0dff8a93533ac2e on codex/host-node-types-land, rebased onto origin/main e8ba3d5.
Built synchronous fs, raw Buffer reads/writes, readSync, catchable coded errors, Stats/Date and all eight requested System wrappers; 17 unit fixtures pass on both backends.
All 33 behavioral mutants are caught solely by Node stdout, plus three freshness mutants and named refusal/layout/loader guard mutants described below.
Remaining: the ten unchanged real host fixtures are five Checker and five NotYet; full repository gate timed out in six packages; native tsc and macOS proof remain undone.

Landing follows the standing rule: rebase into <branch>-land, re-green its oracle, push only that branch, and give its SHA to the owner for area/library. No main push or force-push is authorized or used. The fs landing SHA is in the final handoff.

The implementation recognizes declarations from the sole official @types/node 25.3.3 installation under stage3/api/node_modules. Its unchanged declarations provide node:* types. The loader validates the package version and canonicalizes the index path, avoiding duplicate globals through symlinks. Unknown declared members receive NotYet with their module, declaring class/interface and member. The copied node_fs_file.d.ts and its embed hook have been deleted. The fetched census branch did not contain the promised stage3/api copy; validation installed the exact package, untracked, at the specified location. Dependencies are not vendored.

Original shared commits are c622e9f156dae83e455c5eec5d8ba9df64006aeb, 9efe6d1f752e7f518955c8ef9217ed4b1b1f5765 and 89f595180d8c7f557ac1c17659cfe41f7c82adb9. The last removes an accidental shadow of official Console overloads; lowering safely formats optional strings and refuses other unimplemented console forms. StatsBase and Dirent have distinct member identities. Qualified NodeJS.ErrnoException assertions no longer panic.

The Buffer dependency is origin/codex/host-buffer-crypto 7545a0101fda0572149031630fc30203b7a45cc8, merged here. Its Buffer uses counted numeric arrays. The fs unit returns those arrays for unencoded reads, borrows them for readSync and writeFileSync, and does not expose them as generic ArrayBufferView objects. The narrow borrow exception checks real fs declaration identity, Buffer type, argument position and implemented overload. Buffer/Hash members already implemented by the merged dependency are registered with the shared guard. Two refusal expectations were adjusted for owner-qualified names and official Console types.

System.readFile uses the Buffer worker's decoder for UTF-8 and UTF-16LE, and sys.ts's byte-pair swap for UTF-16BE. Its owned wrapper uses explicit ?? panic checks where Adamic forbids non-null assertions; real upstream acceptance fixtures remain unchanged. The eight System helpers are readFile, writeFile, fileExists, getFileSize, getModifiedTime, setModifiedTime, deleteFile and createDirectory. write cleanup, BOMs, missing-file suppression, file-only size and EEXIST race handling are covered.

readSync implements the five-argument Buffer overload, preserving Node's offset validation, ToInt32 length conversion, position validation before zero-length return, empty-buffer failure, bounds checks, fd validation, descriptor offsets and short reads. Other declared overloads are named NotYet. Buffer file writes preserve invalid UTF-8 bytes, append/truncate/mode/flush and caller fd ownership. Node's fd validation checks finite range before fractional values. Stats includes the requested fields and the atime Date used by the real setModifiedTime fixture's driver. Both timestamp Dates are accounted for in freshness and native layout proofs; a user object with an atime field in slot zero checks conflicting layouts.

The freshness hook evaluates every argument in order, leaves synchronously borrowed objects confined, and models owned objects and timestamp Dates as fresh. Tests cover all current fs operations, writes inside arguments, borrowing, and returned allocations. No edits were made to internal/native/emit.go, internal/lower/lower.go, internal/native/native.go or internal/oracle/oracle_test.go.

Validation (all command output redirected to logs):

- bash cloud/setup.sh: /tmp/fs_file_setup.log; Go/clang/Node/submodules ready 0s each, cache warm 100s, done 100s. nproc 5, CPU quota 4. Source /workspace/adamic-tools/env.sh; Go 1.27.1, clang 20.1.8, Node v24.19.0.
- go test ./... on isolated 080789f plus freshness fix: /tmp/fs_fresh_full_gate.log. Freshness passed 86.975s; load, lower, flow, native and oracle passed. Overall FAIL: 600-second timeouts in internal/unicodeproperties and stage1/cohere/{css,json,lint,markdownblocks,typeaware}. This is not a full-gate pass.
- go test ./internal/fresh -count=1 after restoring three mutants: /tmp/fs_fresh_package.log, PASS 67.588s. Focused final tests: /tmp/fs_fresh_final_focused.log, PASS 0.035s.
- Shared landing: go test ./internal/load ./internal/lower -run 'TestNodeLibrary|TestConsole' -count=1 -timeout 30m: /tmp/node_types_land_gate.log, PASS 0.885s and 2.701s.
- Full touched-package attempt before the final byte-write refactor: /tmp/fs_pinned_buffer_packages_final.log. load/lower/flow/fresh passed; native exposed the missing atime layout. That defect was fixed; TestRuntimeFieldLayoutsAreIncluded passes in /tmp/fs_atime_layout.log, 0.006s. This attempt is not described as a pass.
- go test ./internal/oracle -run 'TestNodeFSFile|TestInputAgreesWithNode|TestCountsAreRecorded|TestNativeAgreesWithNode/internal/oracle/testdata/node_buffer_' -count=1 -timeout 30m -args -update-counts: /tmp/fs_buffer_complete_oracle.log, PASS 30.982s. Includes all 15 fs fixtures, 31 behavioral mutants, existing input fixtures, Buffer source oracles, sanitizers, leak checks and allocation counts.
- Raw-read/System.readFile, readSync/atime and byte-write focused gates: /tmp/fs_system_read_buffer_delivery.log, /tmp/fs_sync_atime_delivery.log and /tmp/fs_write_buffer_gate.log, PASS 3.158s, 13.061s and 5.850s.
- python3 internal/oracle/node_fs_file_host_check.py --compiler /tmp/adamic-host-fs --logs /tmp/fs-real-host --report internal/oracle/node_fs_file_host_status.json --mutants: /tmp/fs_real_host_gate.log, exit 1 because all ten real fixtures remain blocked. All ten match recorded Node stdout/stderr/exit, and all ten source mutants are caught. Native and JavaScript stages are recorded separately. A generic build/tool failure is never classified as Checker.

Real acceptance outcomes on both backends:

| Fixture | Native | JavaScript | Current first blocker |
|---|---|---|---|
| 01_readFile_utf8 | Checker | Checker | Buffer indexed reads can be undefined in the swap assignments |
| 02_readFile_utf16le | Checker | Checker | Same indexed-read typing |
| 03_readFile_utf16be | Checker | Checker | Same indexed-read typing |
| 04_readFile_missing | Checker | Checker | Same indexed-read typing |
| 05_writeFile | NotYet | NotYet | node:os.tmpdir in driver |
| 06_fileExists | NotYet | NotYet | node:os.tmpdir in driver |
| 10_getModifiedTime | NotYet | NotYet | node:os.tmpdir in driver |
| 11_setModifiedTime | NotYet | NotYet | node:os.tmpdir in driver |
| 12_deleteFile | NotYet | NotYet | node:os.tmpdir in driver |
| 13_createDirectory | Checker | Checker | catch variable e is unknown at e.code |

One-line language reproducers: const b=Buffer.from([1,2]); b[0]=b[1]; and try { throw new Error('x'); } catch(e) { console.log(e.code); }. Typed/narrowed catches work and correctly observe filesystem error code/message. These diagnostics are not an inability to throw catchable fs errors. Scratch drivers also need rmSync, symlinkSync, path.join and os.tmpdir, outside this unit's census contract; a dependency branch was requested through the user. No real fixture is reported green.

Behavioral mutants (each compiles, exits zero and has no sanitizer/leak failure; only source Node stdout catches it):

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
| readFileSync Buffer path | Increment the first byte modulo 256 |
| readFileSync Buffer fd | Increment the first byte modulo 256 |
| readSync | Add one to the returned byte count |
| Stats.atime | Add one millisecond to the Date |
| writeFileSync Buffer path | Replace append with truncate |
| writeFileSync Buffer fd | Skip the write |

Guard and proof mutants, all restored:

- Remove the fs case: TestNodeFSFileOperationsAreKnown fails for all 20 then-current operations, /tmp/fs_fresh_unknown_mutant.log.
- Omit argument evaluation: TestNodeFSFileOperandsStillJudgeWrites fails, /tmp/fs_fresh_operand_mutant.log.
- Treat returned objects as outside: TestNodeFSFileResultsAreFresh fails, /tmp/fs_fresh_result_mutant.log.
- Remove declaring owners: TestNodeLibraryDistinguishesReceiverOwners fails with node:fs.isFile, /tmp/node_owner_mutant.log; restored tests pass /tmp/node_owner_restored.log.
- Restore the shadow Console signatures: TestNodeLibraryKeepsOfficialConsoleSignatures fails with real checker diagnostics, /tmp/node_console_shadow_mutant.log; restored test passes /tmp/node_console_restored.log.
- Remove the pinned version check: TestNodeLibraryRejectsDifferentVersion fails, /tmp/node_types_pin_mutant.log.
- Remove the named member guard: TestNodeLibraryNamesUnimplementedMembers fails, /tmp/node_types_guard_mutant.log.
- Remove the qualified assertion identifier check: TestNodeFSFileQualifiedErrorType panics on QualifiedName, /tmp/fs_file_qualified_mutant.log.
- Remove void-value, option-effect or runtime-shape guards: their designated tests fail, /tmp/fs_file_void_mutant.log, /tmp/fs_file_effect_mutant.log and /tmp/fs_file_layout_mutant.log.
- Historical raw-Buffer refusal mutant before Buffer integration: /tmp/fs_file_buffer_guard_mutant.log. The refusal is now deliberately replaced by real Buffer support.
- The first raw-byte mutant after adding a 0xff case produced byte 256 and UBSan caught it. That did not count as an oracle catch. It now increments modulo 256, and passes sanitizer/leak checks before Node stdout catches it.

Linux is the gate of record. macOS uses st_mtimespec/st_atimespec; its libuv flush uses F_FULLFSYNC where this runtime uses fsync, so macOS flush equivalence is not claimed. Resource exhaustion, interrupted close/fsync, multi-gigabyte reads, concurrent deletion and the complete platform errno catalog are not exhaustively tested. Unknown errno values fall back to UNKNOWN; unusual Unicode inspection previews remain a fidelity limit. Stats provides the requested observed fields, not complete Node reflection. Numeric flags, unsupported encodings, bigint Stats, timestamp strings, arbitrary byte views, dynamic options and URL paths remain explicit NotYet. No PR was opened.

Landing validation on origin/main e8ba3d5:

- The fs and Buffer oracle command above passed again in /tmp/fs_file_land_oracle.log, 84.399s, after rebase.
- The unchanged host acceptance runner was rerun with the rebased CLI: /tmp/fs_file_land_host.log, expected exit 1 with the same ten recorded blockers; all Node comparisons and ten source mutants agree.
- Disable the fs Buffer borrow exception: TestNodeFSFileBufferBorrow fails with named node:fs.readSync NotYet in /tmp/fs_file_borrow_mutant.log, 1.302s. Restored test passes /tmp/fs_file_borrow_restored.log, 0.824s.
- Date TimeClip normalizes negative zero after truncation; the rebase retained this original fs merge resolution.
- go test ./internal/load ./internal/lower ./internal/native ./internal/flow ./internal/fresh ./internal/javascript -count=1 -timeout 30m: /tmp/fs_file_land_packages.log, PASS (load 5.076s, lower 69.885s, native 171.247s, flow 141.179s, fresh 90.594s; JavaScript has no package tests and is covered by the oracles).
- go test ./internal/oracle -run ^TestCountsAreRecorded$ -count=1 -timeout 30m, without updates: /tmp/fs_file_land_counts.log, PASS 21.942s.
- A final git fetch confirmed origin/main remains e8ba3d5.

Scratch setup follow-up (October 7):

- Added fs.mkdtempSync with string/UTF-8 results and atomic private directory creation; failures preserve the original prefix plus XXXXXX in Node's message. Broader Buffer/URL/encoding overloads remain named NotYet.
- Added fs.rmSync with recursive and force, default retry settings, lstat validation, directory refusal, recursive trees, dangling symlinks and symlink-safe removal. Non-default retry settings are explicitly NotYet. The existing error shape owns counted string name/message/code fields for every new failure too.
- Two new real-directory unit fixtures cover success, missing paths, directories, permissions, NUL validation, distinct temporary directories and cleanup. /tmp/fs_setup_oracle_fixed.log passes 5.774s on Node, native and JavaScript, with sanitizer/leak checks. The mkdtemp mutant adds a byte to its prefix, yielding a seven-byte observed suffix; the rm mutant forces every missing path. Both exit cleanly and are caught only by Node stdout.
- Merged directory worker 3a090c7ed507901a5eb3d8b6724d42a91d847b7b. Registered its already implemented members, preserved StatsBase dispatch and retained both sets of IR/freshness/native hooks.
- Inspected and merged process worker 7a52252c59f34ff72f08e6f707b0de4fe12bab0a, then backed it out: it does not implement os.tmpdir, and its inherited host oracle panics when nonterminal stdout.columns is undefined under official Node types (/tmp/fs_setup_process_inputs.log). No process worker implementation is included in the final scratch landing.
- The canceled errorCode primitive and its temporary .a declaration were completely removed. The TypeScript seat owns adaptation 47. codex/stage3-host-errors was not published when checked. The original fixtures remain unchanged.
- Compiler blocker, reproduced on current main: import {statSync} from 'node:fs'; try {statSync('missing');} catch(e) {if(typeof e === 'object' && e !== null && 'code' in e) console.log(typeof e.code);}. The compiler refuses in. Ordinary function errorCode(error: unknown) independently stops at representation of the unknown parameter. These are compiler items; no checker option was changed.
- The real fixtures 05,06,10,11,12 now stop at named node:os.tmpdir instead of mkdtempSync. 01-04 still stop at the unadapted indexed reads, and 13 at the unadapted unknown catch variable. None is claimed green.
- Rebased a new codex/host-fs-file-scratch-land branch onto origin/main f8013f0; the previously published landing branch was not rewritten. No main or force push.
- Earlier combined integration attempts exposed legacy process/directory declaration assumptions and failed. The final gate is recorded separately below. A filtered counts attempt produced an incomplete table; it was preserved in a stash and the complete table restored before full regeneration. The stash retains the pre-rollback tracked work.

Final scratch landing validation on f8013f0 (all output logged):

- go test ./internal/oracle -run 'TestNodeFSFile|TestNodeFSDirectory|TestNodePath|TestInputAgreesWithNode|TestCountsAreRecorded' -count=1 -timeout 30m -args -update-counts: /tmp/fs_setup_land_oracle.log, PASS 114.802s. Includes all 17 fs_file fixtures and 33 Node-only behavioral mutants, inherited directory/path fixtures and their mutants, input observations, sanitizer/leak checks and complete allocation-count regeneration.
- go test ./internal/load ./internal/lower ./internal/native ./internal/fresh ./internal/flow -count=1 -timeout 30m: /tmp/fs_setup_land_packages.log, PASS; load 5.046s, lower 77.609s, native 172.870s, fresh 94.363s, flow 132.999s.
- go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 30m, without updates: /tmp/fs_setup_land_counts.log, PASS 16.066s.
- The new retry guard mutant first survived because the literal fixture stopped at the existing option-expression guard. Its fixture now uses a const options binding, directly reaching the retry guard. Removing the guard is caught with want named NotYet, got nil: /tmp/fs_setup_retry_guard_mutant2.log, 0.460s. Restored check passes /tmp/fs_setup_retry_guard_restored2.log, 0.690s. No surviving mutant is claimed caught.
- Real host runner: /tmp/fs_setup_land_host.log, exit 1 for the documented five Checker/five NotYet blockers. Node comparisons and all ten source mutants pass. Narrowing reproducer still refuses in: /tmp/fs_setup_land_in.log.
- Final dependency check: origin/codex/host-process advanced to d06ed474e90a4fc5270f242dea8b1ec8d960fb54; its lower/runtime/load files still have no tmpdir implementation. origin/codex/stage3-host-errors remains unpublished. origin/main remains f8013f0.
