Catchable crypto finalization still needs runtime integration; owned full fixtures stop at Checker (02/03), non-null refusal (21), or undefined lowering (22).
Merged area/library 53f44e05 with fs-file and macOS fixes, retaining both units through named Buffer hooks.
Merge parents: a38287b and 53f44e05; codex/host-buffer-crypto-land. Merge commit, no rebase, force-push or push to main.
Linux counts PASS 115.024s; packages including full flow PASS; whole oracle PASS 356.494s; byte-index mutant caught by Node on both backends.
All 25 Node observations agree; 11_setModifiedTime executes and agrees on both backends. Other full fixtures remain frontend-blocked; macOS and native tsc are untested.

Library area fs-file merge, October 7

Resolved the shared expression dispatch against the area version, preserving processValue, nodeFSDirectoryValue, named unsupported-host refusals, never-returning closures and omitted-argument shaping. Reapplied nodeBufferRepresentation and nodeBufferContextualView as small named hooks. The Buffer contextual view now accepts the fs worker's synchronous readSync/writeFileSync Buffer arguments. Kept the area's runtime error layout, process layouts, freshness handling and Buffer tests. NOTICE and host/status.json retain the area versions unchanged.

Preserved _DARWIN_C_SOURCE after _POSIX_C_SOURCE in fs-file and process runtimes, Linux-only detect_leaks settings, concurrent terminal draining and canonical directory handling. Linux was the execution platform; no macOS execution is claimed.

Commands after sourcing /workspace/adamic-tools/env.sh:

- `go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 30m -args -update-counts`: PASS, 115.024s. Full Linux table regenerated, including two changed area rows.
- `go test ./internal/load ./internal/lower ./internal/native ./internal/javascript ./internal/fresh ./internal/flow ./internal/ir -count=1 -timeout 30m`: PASS. Load 5.240s, lower 84.716s, native 223.145s, fresh 101.499s, flow 162.755s, IR 2.996s; JavaScript has no standalone tests.
- `go test ./internal/oracle -count=1 -timeout 30m`: PASS, 356.494s, whole oracle under ordinary worker cache policy.
- `go test ./internal/lower -run 'NodeBuffer|NodeFSFile|NodeLibrary' -count=1 -v`: PASS, 9.754s, final conflict resolution.
- `python3 stage3/host-buffer/check_host_acceptance.py --all --logs /tmp/host-buffer-area53-host`: 25 source Node observations agree. Both backends record the same stages; exit 1 explicitly marks remaining blocked fixtures. The sole compiling fixture matches recorded stdout, stderr and exit on both backends. Full table and diagnostics: [area53/host-stages.md](area53/host-stages.md) and [area53/host-stages.json](area53/host-stages.json).
- `go vet ./internal/load ./internal/lower ./internal/native ./internal/javascript ./internal/fresh ./internal/flow ./internal/ir ./internal/oracle`: PASS, no output. `gofmt -l cmd internal`: no output.
- `python3 internal/oracle/node_buffer_mutants.py buffer_index`: CAUGHT by source Node stdout comparisons on native and JavaScript; compiler and sanitizers accepted the mutant. Exact source restored. `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/node_buffer_writes.a$' -count=1 -timeout 30m`: restored baseline PASS, 0.710s.

Owned fixture blockers: 02/03 retain TS2322 on b[0]=b[1] under checked indexed access; 21 now reaches Adamic's non-null assertion refusal at _crypto!.createHash("sha256"); 22 reaches NotYet for const c: undefined=undefined. The independent owned components passed in the native package gate. The audited source fixtures were not rewritten to bypass these language checks. All earlier unit mutant evidence is retained below.

The following reports are historical.

Catchable Node crypto .code remains unsupported; full host fixtures retain the indexed-read and fs.mkdtempSync blockers.
Merged area/library 2faf682 into the Buffer/crypto landing branch, keeping compiler semantics with small named host hooks and the audited host records.
Merge parents: 2a4a229 and 2faf682; branch codex/host-buffer-crypto-land. No rebase, force-push or push to main.
Checks: full touched packages including internal/flow PASS; full oracle PASS; Linux counts regenerated; nine owned fixtures PASS uncached; byte-index mutant caught by Node.
Full acceptance on both backends remains 02/03 Checker and 21/22 NotYet; owned host components pass. Native tsc proof and macOS remain untested.

Library area merge, October 7

Merged origin/area/library at 2faf682bbf365dcc0afde3768832a2fdffc7b965 into the existing published landing branch. Conflicts were expression.go, library_node_test.go, counts.md, stage3/fixtures/NOTICE and stage3/fixtures/host/status.json. This landing uses a merge commit with both published parents, without rewriting history.

Expression conflict: took the compiler's entire area version, retaining iterator refusal, acceptsUndefined, typeof presence handling, proven satisfies relations, spelling of stale narrowings and conditional pair handling. Reapplied the host representation, unsupported-use and view checks as small named hooks. nodeBufferRepresentation and nodeBufferContextualView live in library_node_buffer.go; existing nodeBufferUnsupportedUse/nodeBufferView remain the refusal hooks. Buffer.from's input-copy exception is preserved without weakening the compiler checks on ordinary expressions.

QualifiedName conflict: kept the area's corrected test and invariance.go fix. A proven qualified assertion passes; an Error-to-ErrnoException downcast with unproven optional fields is refused. NOTICE and host/status.json are byte-identical to origin/area/library, as requested. Linux counts were regenerated from executions of the complete fixture set. Compared with the area table, only the nine new Buffer/crypto rows are added; no area row changes or disappears.

Gate commands, after sourcing /workspace/adamic-tools/env.sh:

- `go test ./internal/lower -run 'NodeBuffer|NodeLibrary|Proven|Predicat|Narrowed|Iterator' -count=1 -v`: PASS, 5.239s.
- `go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 30m -args -update-counts`: PASS, 94.983s, Linux.
- `go test ./internal/load ./internal/lower ./internal/native ./internal/javascript ./internal/fresh ./internal/flow ./internal/ir -count=1 -timeout 30m`: PASS. Load 4.952s, lower 49.974s, native 189.610s, fresh 65.732s, flow 117.399s, IR 3.192s; JavaScript has no standalone tests. This runs the whole flow package explicitly.
- `go test ./internal/oracle -count=1 -timeout 30m`: PASS, 305.632s, including count-table verification and the area's regex/map fixtures. Ordinary worker cache policy; not claimed uncached.
- `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'Test(Native|Input)AgreesWithNode/internal/oracle/testdata/node_buffer_' -count=1 -v -timeout 30m`: PASS, 6.153s, all nine owned fixtures; native release/sanitizer and JavaScript held to Node, zero cache hits.
- `python3 stage3/host-buffer/check_host_acceptance.py --logs /tmp/host-buffer-area-acceptance`: exits 1 as intended for blocked acceptance; actual Node output agrees for all four fixtures. 02/03 are Checker on both backends, 21/22 are NotYet at node:fs.mkdtempSync on both backends. Owned component tests pass in the full native package gate.
- `go vet ./...`, `gofmt -l cmd internal`, `git diff --check`: pass, no output.

The old byte-index mutant pattern did not match the compiler's new l.defined wrapper. That attempt was invalid and not counted as a catch. The unit mutant runner was updated to keep the compiler wrapper intact while changing only Buffer's index by +1; the actual mutant was caught by stdout differences against source Node on native and JavaScript, with no compiler/sanitizer failure. Restored baseline: `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/node_buffer_writes.a$' -count=1 -timeout 30m` PASS, 0.614s. Exact source restored; /tmp/host-buffer-area-buffer-index-mutant.log and /tmp/host-buffer-area-mutant-restored.log record both results. All earlier mutant evidence is retained in the historical sections.

Logs are under /tmp/host-buffer-area-*.log; each test invocation writes to a file. The complete repository gate was not run; all touched packages, the full oracle, explicit flow and the owned uncached fixtures were run. No private Node declaration files are added.

The sections below are historical reports from earlier landing stages.

Landing with the shared Node types, October 7

Merged 69c71d5171801bc355400452a0dff8a93533ac2e into the original host branch at 9ed6a002f61eb24ade6af1e2ba88a763251be448, created codex/host-buffer-crypto-land, and rebased onto current origin/main e8ba3d5d81de4d3773c723914fccd4c76248b965. The shared loader changes are replayed as c6fc99f, 26b5efc and efac77f. No private node_*.d.ts files are present. Per the standing rule, landing uses a separate rebased branch, an oracle gate and a normal push; only integration moves main. No force-push or PR.

The shared default-refusal guard requires registration. bcbce83 registers exactly Buffer, BufferConstructor.from, Buffer.toString, crypto.createHash, Hash.update and Hash.digest. Recognition now uses load.IsNodeLibrary's canonical pinned path. Other declared members still get named NotYet; tests now expect the shared guard's declaring owner for BufferConstructor.alloc and inherited Writable.writable.

The API directory is absent from the census branch in this checkout, so the official npm package was installed locally using `npm install --prefix stage3/api --no-save --package-lock=false @types/node@25.3.3`. The upstream package and its undici-types dependency are not committed or copied into internal/load. Loader and unit tests confirm canonical declaration identity and member refusals.

After the loader merge, the four unmodified full fixtures have these exact stages on both backends:

| Fixture | Node observation | Native | JavaScript | Owned component |
| --- | --- | --- | --- | --- |
| 02_readFile_utf16le.a | Matches status.json | Checker | Checker | Pass |
| 03_readFile_utf16be.a | Matches status.json | Checker | Checker | Pass |
| 21_createHash.a | Matches status.json | NotYet, node:fs.mkdtempSync | NotYet, node:fs.mkdtempSync | Pass |
| 22_createHash_fallback.a | Matches status.json | NotYet, node:fs.mkdtempSync | NotYet, node:fs.mkdtempSync | Pass |

The actual checker blocker for 02 and 03 is TS2322 at the upstream BE byte-swap assignments (lines 25 and 26): number | undefined cannot be assigned to number. One-line reproducer: `import {Buffer} from 'node:buffer'; const b=Buffer.from([1,2]); b[0]=b[1];`. It was reported immediately. Proving the loop's index bounds or changing the upstream patch set belongs to the checker/TypeScript seats; this unit did not weaken the types or invent a coercion. The older undefined-selector blocker for fixture 22 remains separate from its currently observed first refusal. The fs core branch was not merged because the user requested 69c71d5 specifically.

Re-green commands, all after sourcing /workspace/adamic-tools/env.sh:

- `go test ./internal/lower -run 'TestNodeBuffer|TestNodeLibrary' -count=1 -v`: PASS, 4.458s. All fourteen unit refusal probes and the shared member/type tests pass.
- `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'Test(Native|Input)AgreesWithNode/internal/oracle/testdata/node_buffer_' -count=1 -v -timeout 30m`: PASS, 2.712s; eight native fixtures plus the input fixture, both backends, native release/sanitizer, zero cache hits.
- `go test ./internal/native -run 'TestNodeBufferHost|TestNodeBufferRuntimeWithoutDeclarations' -count=1 -v`: PASS, 1.167s.
- `python3 stage3/host-buffer/check_host_acceptance.py --logs /tmp/host-buffer-land-acceptance`: exits 1, preserving the exact blocked stages above; Node observations all agree.

Final restored gate and mutant observations:

- `python3 internal/oracle/node_buffer_mutants.py`: all fifteen mutants caught by byte-for-byte Node comparison. Each mutant listed in the semantic results below was rerun after rebase. None was killed by the compiler, sanitizer or leak checker. Sources restored after each.
- `python3 internal/oracle/node_buffer_proof_mutants.py`: hash_alias caught by TestNodeBufferHashUpdateKeepsAlias; hash_fields by TestRuntimeFieldLayoutsAreIncluded; host_cycle_dispatch by TestEveryWriteIsRecordedAndKnown. Exact sources restored.
- `go test ./internal/load ./internal/lower ./internal/native ./internal/javascript ./internal/flow ./internal/fresh ./internal/ir -run 'NodeBuffer|NodeLibrary|RuntimeFieldLayoutsAreIncluded|EveryWriteIsRecordedAndKnown' -count=1 -timeout 30m`: PASS. Load 2.140s, lower 9.269s, native 3.719s, fresh 11.760s; JavaScript has no standalone tests and the flow/IR packages have no matching tests.
- After restoring every mutant, the same uncached nine-fixture oracle command above: PASS, 2.743s, zero cache hits; see /tmp/host-buffer-land-oracle-restored.log.
- `go vet ./...`, `gofmt -l cmd internal`, and `git diff --check`: exit 0, no output. Complete repository test gate was not rerun; this is the focused worker gate plus the owned uncached oracle.
- Logs: /tmp/host-buffer-land-{focused,mutants,proof-mutants,oracle-restored,acceptance,vet,format,diff-check}.log and /tmp/host-buffer-mutants/ for each individual mutant.

Full acceptance remains incomplete at the exact stages above. Catchable Node error .code, the undefined fallback selector, downstream host members, native tsc --noEmit and macOS are not proven by this gate.

The following sections record earlier stages and validations chronologically. Their missing-loader observations are historical; the loader dependency is now resolved.

Real host acceptance, October 7

The user's host fixture branch is merged and pushed. `internal/native/node_buffer_host_test.go` compares stdout, stderr and exit status with the unmodified fixture's Node run and committed status.json. Buffer decoding/swap and SHA-256 use direct IR to isolate this unit from missing module declarations and fs runtime. The fallback test lowers the verbatim generateDjb2Hash function through the checker and both backends. It does not prove the undefined-crypto selector or the ambient SHA function declaration. These are component passes, not full acceptance passes.

| Fixture | Node vs status.json | Native full fixture | JavaScript full fixture | Owned component |
| --- | --- | --- | --- | --- |
| 02_readFile_utf16le.a | Agrees | Checker | Checker | Pass, LE BOM offset/odd byte/surrogate cases |
| 03_readFile_utf16be.a | Agrees | Checker | Checker | Pass, byte swap and LE decoding |
| 21_createHash.a | Agrees | Checker | Checker | Pass, four SHA-256 digests |
| 22_createHash_fallback.a | Agrees | Checker | Checker | Pass, verbatim djb2 function |

`python3 stage3/host-buffer/check_host_acceptance.py --logs /tmp/host-buffer-acceptance` exits 1 and records exact diagnostics: the full fixtures cannot resolve node:* modules. The fs_file worker's published branch at 080789f still contains the old unit declaration copy, not the requested shared pinned-type hook. It was not merged. The hook SHA has been requested. We cannot verify canonical @types/node symbol identity or unsupported-member guards against that hook until it arrives.

Language blocker reported immediately: `const absent: undefined = undefined;` fails with `stage 0 can't lower a value of type undefined yet`. An extraction of fixture 22 preserving its undefined-crypto selector confirms this independently of Node imports. Its log is host-fallback-blocker.log. Shared language files were not changed. Other downstream language issues are unverified while the checker blocks the full fixtures.

Commands and observations:

- `python3 stage3/fixtures/host/check.py --mutants --logs /tmp/host-buffer-real-fixtures`: all 25 Node observations and initial Checker statuses agree; all 25 source mutants and the diagnostic-status mutant caught. This checks the fixture baseline, not acceptance progress.
- `go test ./internal/native -run 'TestNodeBufferHost|TestNodeBufferRuntimeWithoutDeclarations' -count=1 -v`: PASS, 1.330s after restoring all four new mutants. Both backend outputs match; native release and sanitizer checked.
- `go test ./internal/load ./internal/lower ./internal/native ./internal/javascript ./internal/flow ./internal/fresh ./internal/ir -run '^$'`: PASS; `go vet ./internal/native`: PASS; acceptance runner Python compilation and git diff check: PASS.
- New SHA initial-state mutant 6a09e667 -> 6a09e666: native stdout differs from status.json.
- New LE start-offset mutant 2 -> 0: native stdout retains BOM and differs from status.json.
- New BE buffer write mutant leaves the old byte: native decoded stdout differs from status.json.
- New djb2 seed mutant 5381 -> 5382 in the extracted function: generated JavaScript stdout differs from status.json.

No compiler error or sanitizer failure counted as a mutant catch. Exact source bytes were restored after every mutant, followed by the passing baseline. Logs and stage diagnostics are committed beside this report. Linux only; macOS and the native tsc proof remain untested.

Scope and references

The contract is origin/codex/tsc-census (429c1177f0130f785c19cf590d1860513b2ddbfc) stage3/census/REPORT.md and its system_contract.json/node_derived_sites.json, using TypeScript sys.ts from v6.0.3 and the lockfile-pinned @types/node 25.3.3. sys.ts will use static node:* imports through the TypeScript patch set; literal require calls later become the same imports. The fs_file worker owns the shared declaration loader. This branch has no unit-owned node_*.d.ts copies or declaration loader hook.

Buffer.from accepts strings with literal utf8/utf-8, utf16le/utf-16le/ucs2/ucs-2, base64 and hex, plus numeric arrays and Buffer copies. toString accepts the same encodings and numeric byte offsets. Numeric writes truncate modulo 256 and ignore invalid indices. The sys.ts UTF-16 big-endian BOM loop is reproduced in a fixture. System base64encode/base64decode and createSHA256Hash are tested as the same wrapper chains; this unit does not modify TypeScript's System interface.

The UTF-8 decoder ports V8's DFA and rejected-byte reprocessing from Node v24.19.0 deps/v8/src/strings/unicode-inl.h. The scalar base64/hex behavior follows Node deps/nbytes/include/nbytes.h and src/string_bytes.cc. Licenses are recorded in THIRD_PARTY_NOTICES.md. SHA-256 implements FIPS 180-4 with explicit unsigned 32-bit operations and big-endian length padding.

Observed coverage

Nine fixtures cover encoding aliases, padding/whitespace and malformed base64, hex stopping, all 256 single-byte and 65,536 two-byte UTF-8 inputs, malformed/truncated three/four-byte boundaries, all 65,536 UTF-16 code units, lone surrogates, odd byte lengths, BOM swaps, array copies and aliases, numeric writes, hash padding boundaries and repeated updates. Random coverage uses 512 deterministic random UTF-16 strings, compared against Node digests. The input fixture uses the oracle's real-directory and uid-dropping harness. Successful native runs are compared under sanitizer and release builds; the harness also checks leaks.

Unsupported input forms are rejected before C emission: dynamic/other encodings, implicit coercion objects, algorithms other than literal sha256, binary Hash input, non-hex digest, stream options, detached methods, compound Buffer writes, and structural views that expose native internal slots. Uncaught post-digest update matches Node's Error [ERR_CRYPTO_HASH_FINALIZED] text and exit behavior. Catching it is rejected because the existing error runtime cannot represent .code.

The native representation uses counted numeric arrays for bytes and retains the whole hash message until digest. It has no GC and no new heap kind, but uses more memory than packed bytes and streaming SHA-256. Linux is the tested gate; macOS has not been run and no platform-specific encoding difference was observed. The complete native tsc --noEmit proof depends on other host units and module wiring.

Toolchain

bash cloud/setup.sh succeeded: Go 0s, clang 0s, Node 0s, submodules 0s, cache 97s, total 97s. Environment: /workspace/adamic-tools/env.sh; Go 1.27.1, clang 20.1.8, Node 24.19.0; nproc 5 (cgroup budget 4 CPUs).

Validation commands

- go test ./internal/load ./internal/lower ./internal/native ./internal/javascript -run '^$': passed.
- go test ./internal/lower -run TestNodeBufferRefusals -count=1 -v: passed, 0.116s.
- go test ./internal/oracle -run 'Test(Native|Input)AgreesWithNode/internal/oracle/testdata/node_buffer_' -count=1 -timeout 30m: passed.
- python3 internal/oracle/node_buffer_mutants.py: 15 independent restored mutations, all caught by source Node comparison. The finalization summary parser was corrected to recognize "exit codes differ" and rerun successfully. Formatting-independent token matching was checked for all 15 source patterns.

Each test writes a log under /tmp/host-buffer-*.log. The persistent mutant runner writes detailed logs under /tmp/host-buffer-mutants and restores the exact source in finally. Compiler errors, sanitizer failures and leak reports do not count as catches.

Mutant results

```text
from_utf8_surrogate: CAUGHT; fixture=encodings; comparisons=stdout differs; log=/tmp/host-buffer-mutants/from_utf8_surrogate.log
from_numeric_bytes: CAUGHT; fixture=writes; comparisons=stdout differs; log=/tmp/host-buffer-mutants/from_numeric_bytes.log
utf16_surrogate: CAUGHT; fixture=utf16; comparisons=stdout differs; log=/tmp/host-buffer-mutants/utf16_surrogate.log
utf16_odd_length: CAUGHT; fixture=utf16; comparisons=stdout differs; log=/tmp/host-buffer-mutants/utf16_odd_length.log
base64_padding: CAUGHT; fixture=encodings; comparisons=stdout differs; log=/tmp/host-buffer-mutants/base64_padding.log
base64_whitespace: CAUGHT; fixture=encodings; comparisons=stdout differs; log=/tmp/host-buffer-mutants/base64_whitespace.log
hex_case: CAUGHT; fixture=encodings; comparisons=stdout differs; log=/tmp/host-buffer-mutants/hex_case.log
byte_swap_store: CAUGHT; fixture=bom; comparisons=stdout differs; log=/tmp/host-buffer-mutants/byte_swap_store.log
buffer_length: CAUGHT; fixture=utf16; comparisons=JavaScript backend: stdout differs, stdout differs; log=/tmp/host-buffer-mutants/buffer_length.log
buffer_index: CAUGHT; fixture=writes; comparisons=JavaScript backend: stdout differs, stdout differs; log=/tmp/host-buffer-mutants/buffer_index.log
hash_initial_state: CAUGHT; fixture=crypto; comparisons=stdout differs; log=/tmp/host-buffer-mutants/hash_initial_state.log
hash_update_bytes: CAUGHT; fixture=crypto; comparisons=stdout differs; log=/tmp/host-buffer-mutants/hash_update_bytes.log
hash_digest_hex: CAUGHT; fixture=crypto; comparisons=stdout differs; log=/tmp/host-buffer-mutants/hash_digest_hex.log
utf8_reprocess: CAUGHT; fixture=utf8; comparisons=stdout differs; log=/tmp/host-buffer-mutants/utf8_reprocess.log
hash_finalization: CAUGHT; fixture=finalized; comparisons=exit codes differ; log=/tmp/host-buffer-mutants/hash_finalization.log
```

Earlier gate observations, before the declaration ownership correction

The initial command ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./... found two integration gaps: TestEveryWriteIsRecordedAndKnown did not know NodeBufferCall, and TestRuntimeFieldLayoutsAreIncluded lacked Hash's bytes/finalized slots. These are corrected by a unit-owned freshness handler and a runtime-layout entry. The new TestNodeBufferHashUpdateKeepsAlias checks that the analysis preserves Hash.update's receiver alias. The initial run's full oracle passed (205.778s), but the whole gate was stopped with exit 143 after more than ten minutes while unrelated parser audits remained active. It is not a completed passing full gate.

Commands on the initial restored implementation, before removal of its local declarations:

- gofmt -l cmd internal: no output.
- go vet ./...: exit 0, no output.
- go test -count=1 -timeout 30m ./internal/load ./internal/lower ./internal/native ./internal/javascript ./internal/flow ./internal/fresh ./internal/ir ./internal/oracle: exit 0. Load 1.554s; lower 38.412s; native 164.433s; flow 118.121s; fresh 52.683s; oracle 162.408s; JavaScript/IR have no standalone tests. See [touched-gate.log](touched-gate.log).
- ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'Test(Native|Input)AgreesWithNode/internal/oracle/testdata/node_buffer_' -count=1 -timeout 30m: exit 0, oracle 1.496s. See [oracle-final.log](oracle-final.log).
- go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts: exit 0, 21.329s. Only nine new fixture rows changed.
- python3 internal/oracle/node_buffer_proof_mutants.py: exit 0. These three additional mutants prove compiler guards, separately from the 15 semantic mutants held to Node. See [proof-mutants.log](proof-mutants.log).
- After restoring the proof mutants, go test ./internal/fresh ./internal/native -run 'Test(NodeBufferHashUpdateKeepsAlias|EveryWriteIsRecordedAndKnown|RuntimeFieldLayoutsAreIncluded)' -count=1 -timeout 30m: exit 0, fresh 9.506s and native 0.005s. See [proof-final.log](proof-final.log).
- git diff --check: no output.

Additional proof mutants

```text
hash_alias: CAUGHT by TestNodeBufferHashUpdateKeepsAlias; log=/tmp/host-buffer-mutants/hash_alias.log
hash_fields: CAUGHT by TestRuntimeFieldLayoutsAreIncluded; log=/tmp/host-buffer-mutants/hash_fields.log
host_cycle_dispatch: CAUGHT by TestEveryWriteIsRecordedAndKnown; log=/tmp/host-buffer-mutants/host_cycle_dispatch.log
```

The first alias mutant accidentally left an unused Go variable and was rejected as a build failure; it did not count. The corrected mutant explicitly consumes the receiver and returns a distinct fresh object, compiling successfully and failing only the alias assertion. Every mutant restored the exact original source before the final checks.

Small shared hooks are required in the declaration loader, expression dispatch, Buffer type/index handling, exception refusal, flow effects, freshness dispatch and runtime field-layout proof. The four files prohibited by the unit instructions were not edited. No PR is opened.

Declaration ownership correction

The TypeScript seat corrected the integration contract: sys.ts uses static node:* imports, @types/node is exactly 25.3.3 from stage3/api on origin/codex/tsc-census, and codex/host-fs-file owns the shared declaration loader. Removed internal/load/node_buffer.d.ts, node_crypto.d.ts and their embed file. Restored source_fs.go and tsconfig.json to origin/main's declaration loading. No replacement loader was built.

Lowering now recognizes Buffer/Hash and crypto exports from declarations under /@types/node/ inside the actual node:buffer/node:crypto ambient modules. This path convention and behavior against the full upstream declarations must be confirmed when the shared hook SHA arrives. Supported census calls keep their intrinsic lowering. Other Buffer/Hash methods, static members and crypto exports are refused with NotYet containing the member's name. An early read guard prevents inherited Hash stream fields from being mistaken for private runtime slots, including reads through string indices. Added refusal probes for Buffer.alloc, Buffer.byteOffset, Hash.copy, Hash.writable, randomBytes, Hash-as-value and isUtf8. These probes cannot run until the shared hook is merged.

Current checks after removing the local declarations:

- go test ./internal/load ./internal/lower ./internal/native ./internal/javascript ./internal/fresh ./internal/flow -run '^$': passed. See [shared-types-build.log](shared-types-build.log).
- go test ./internal/fresh ./internal/native -run 'Test(NodeBufferHashUpdateKeepsAlias|RuntimeFieldLayoutsAreIncluded)' -count=1: passed, 0.006s per package.
- go test ./internal/lower -run 'Test(ConsoleLowersToWriteLine|FiveRefused|ReadonlyFieldsAreJudgedByTheirConstructors)' -count=1: passed, 0.106s.
- go test ./internal/native -run '^TestNodeBufferRuntimeWithoutDeclarations$' -count=1 -v: passed, 0.709s. This direct IR test compares both generated JavaScript and sanitized/release native binaries against independently written source on Node v24.19.0. It covers malformed/truncated UTF-8, base64 padding/whitespace, lone UTF-16 surrogates and odd lengths, plus SHA-256 on 32 deterministic random byte-derived UTF-16 strings. See [independent-runtime.log](independent-runtime.log).
- New mutant: change JavaScript Hash.digest emission from .digest('hex') to .digest('hex').toUpperCase(). TestNodeBufferRuntimeWithoutDeclarations fails with "JavaScript stdout differs from Node", without a compiler/sanitizer failure. Restore the exact source and rerun: passed. See [independent-mutant.log](independent-mutant.log).
- gofmt -l cmd internal: no output; go vet ./...: exit 0, no output; git diff --check: no output.

Observed blocker: go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/node_buffer_crypto.a$' -count=1 fails at Load with TS2591, "Cannot find name 'node:crypto'". See [shared-types-blocked.log](shared-types-blocked.log). Thus the earlier source fixture gates are historical evidence, not a claim that the corrected branch's complete gate currently passes. The source oracle fixtures, new member refusal probes, counts verification and source-based mutant runners require the shared hook. Merge its exact commit when supplied, then rerun those checks and the real sys.ts fixture branch when forwarded. Runtime and IR tests remain runnable now.
