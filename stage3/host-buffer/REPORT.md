Catchable Node crypto .code remains unsupported; full fixtures also await the shared @types/node hook and fixture 22 hits undefined-value lowering.
Built Buffer and SHA-256 runtime/lowering plus four real-host component comparisons; merged host fixture branch 1037217 in 7545a01.
Pushed component tests in 950cc69; both backends including native release/sanitizer match status.json for components 02, 03, 21 and 22.
Checks: restored five runtime/component tests PASS (1.330s); four new semantic mutants caught by Node/status.json output comparison; earlier 19 mutants remain documented below.
Full acceptance stages: 02_readFile_utf16le, 03_readFile_utf16be, 21_createHash, 22_createHash_fallback are Checker on BOTH backends; no full fixture is green yet.

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
