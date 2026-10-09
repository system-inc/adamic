Catchable crypto errors remain unsupported: Adamic Error has name/message but no Node .code; try around hash operations is rejected.
Built Buffer encodings, BOM swaps, indexing/length, System wrapper chains and SHA-256 on both backends; removed all unit declaration copies and the loader hook.
Implementation commits: 1321859fccc4c7391d1ca3b34eeb1f18b16ba818 and 4453ba184af55ffba9ac1c1a3b72c801af9fc6e1; branch codex/host-buffer-crypto.
Current validation: Go compilation, vet, formatting, IR proofs and declaration-independent Node runtime comparison pass; source fixtures need the shared @types/node hook.
Mutants: prior 15 semantic and three proof mutants caught; one new independent hash mutant caught by Node. Shared hook merge, real sys.ts fixtures, native tsc proof and macOS remain pending.

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
