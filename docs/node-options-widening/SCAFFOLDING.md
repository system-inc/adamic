Repaired all 17 requested error helpers using errorCode(error: unknown) and guarded code reads.
Merged compiler unknown-narrowing 616870d with merge commit bb993abf; prior proof remains ae3a6c52.
Node stdout before/after is byte-identical for all 17 and two additional system/buffer fixtures.
All 52 existing fs file/directory mutants are caught by Node stdout; sanitizers and leak checks pass.
The process source remains refused at 9:32; its test now pins the full unbound-method diagnostic.

Every helper uses the same narrowing form:

```a
function errorCode(error: unknown): string | undefined {
    return typeof error === 'object' && error !== null && 'code' in error && typeof error.code === 'string' ? error.code : undefined;
}
```

The ordinary Error in node_fs_file_close.a now has its inferred Error type and
uses the same guarded reader. No NodeJS.ErrnoException view or consumption
predicate change was added. The system and buffer fixture helpers, and the
shared System createDirectory catch, receive the same repair because their
imports otherwise retain the identical blocker. The latter still rethrows
all Error values whose code differs from EEXIST.

The original source at ae3a6c52 and repaired source were run through oracle/node.mjs
using the actual input harness, including its unprivileged permission probes.
A temporary Go overlay saved stdout, stderr and exit status for both runs as JSON
under /tmp/node-options-scaffolding-observations; byte comparisons all pass
(10.536s). Raw observations stay local. The following SHA-256 hashes identify the
identical stdout bytes:

| Fixture | Stage | Node stdout bytes | SHA-256 |
|---|---|---:|---|
| node_fs_directory_entries.a | Node/JavaScript/native agree | 2634 | 4cba764ce4cd4656f9eee50861641f41738e664cd227e9eb06d6df4c0d61ab79 |
| node_fs_directory_permissions.a | Node/JavaScript/native agree | 315 | cb2b75d8072f7b8196657f787b1a7780d2c973413af9c533390fceaeeb1f6c44 |
| node_fs_directory_realpath.a | Node/JavaScript/native agree | 2274 | 18bdffd91f3090e583a9b1c82b73ea24359e7fdeef2b00b0366b1cd40dcde170 |
| node_fs_file_buffer.a | Node/JavaScript/native agree | 198 | 1e5f82c42e2acf3ace6d7bc6bf98bc52a716b35c7f38a376497df4ec2317760d |
| node_fs_file_close.a | Node/JavaScript/native agree | 200 | cd885bd2d0ce319f0c541f86a344e5b824ff6dc0bdee0db4c8105f8aa6a1357e |
| node_fs_file_exists.a | Node/JavaScript/native agree | 116 | 09c218347faea6ab8f0affc9c4102ce719d4971bb27a8ed52856d27b3b229901 |
| node_fs_file_mkdir.a | Node/JavaScript/native agree | 333 | 175821b66a776b3fef5f25627c195ee9b8429c5a43d10f8a891bbf98ca6a4349 |
| node_fs_file_mkdtemp.a | Node/JavaScript/native agree | 420 | 7041a24d58b1e061ecefc138229dcebbfa3e76af3d870d133ac7567de2f7d202 |
| node_fs_file_open.a | Node/JavaScript/native agree | 649 | a1a628eca6d1fdf186b390ef5c86f5633bc237af1a0ce71982964a2eecad7c01 |
| node_fs_file_read.a | Node/JavaScript/native agree | 411 | f3810ee7deb3fbaa586bc66f499daaec67c9c09fee5e5caea8ebe375781f07cc |
| node_fs_file_read_sync.a | Node/JavaScript/native agree | 1164 | 5ce4db6617fca4ce61be90cdec11ae9de8063aad4daa8115aab3d76e7bfc9af5 |
| node_fs_file_rm.a | Node/JavaScript/native agree | 571 | 29dbfc26be1e02bdac22ea6c836d79135d7580fdde8b0b2fa98ede6d4f4bca77 |
| node_fs_file_stat.a | Node/JavaScript/native agree | 254 | cb2bc0cb2b1add891c2435071008c3c20dc845a9f14ada5d621ba69c3510f19e |
| node_fs_file_system.a | Node/JavaScript/native agree | 117 | 795bb9e053c65b5ad1c4db2e9c934d24d4cdd88499c8237aa420a557297827a9 |
| node_fs_file_unlink.a | Node/JavaScript/native agree | 234 | 408aba26a07ac62140f7dcffd7d9b8f4030b0960acaf90e3399a78fd62815f64 |
| node_fs_file_utimes.a | Node/JavaScript/native agree | 295 | 2e00784d13d9d91583f76ee0894e85f556930f84912a12b41831aedbfdc75dce |
| node_fs_file_write.a | Node/JavaScript/native agree | 62 | 9f601900fd33cf23a45ea86ffed036dd4624f5a84eb2274740d4bac1c8acae5a |
| node_fs_file_write_buffer.a | Node/JavaScript/native agree | 127 | 6e47622c500c03d258686d63afa8b8bcbd64c288b1193711924a0b1a737ebd18 |
| node_fs_file_write_file.a | Node/JavaScript/native agree | 147 | d8ec1cdf5dd331a061d80091e7ae3014354e2c841e9e2589d3d342fcfbc043f0 |

The snapshot test temporarily placed original-source copies beside the fixtures
to preserve relative imports. Running it alongside directory enumeration caused
an initial node_fs_directory_system.a stdout mismatch. Its source copies were
removed, and that test passes uncached when rerun alone (4.568s). This initial
comparison is excluded from the final semantic results.

Dependency merge resolution: retain the current centralized native typeof
classifier rather than restoring the old class_static.c classifiers. The unknown
null sentinel is handled before accessing object class metadata. All existing
Linux counts were preserved pending regeneration; incoming unknown fixture rows
were retained. No Node runtime guards or macOS definitions were changed.

Full Linux counts regeneration passes (353.375s):
`go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -parallel=2 -timeout=30m -args -update-counts`.
The complete table is regenerated, including incoming unknown fixtures and all
changed host rows. No macOS-derived rows are used.

The merge null-sentinel mutant uses a Go build overlay that removes the guard
before object metadata access. TestNativeAgreesWithNode/unknown_narrowing.a fails
with ASan global-buffer-overflow in adamic_union_typeof (124.123s), not a build
failure. Production files are never mutated. The restored unknown comparisons
and inherited TestUnknownNarrowingMutants pass uncached (7.637s).

Touched package gate passes:
`go test ./internal/lower ./internal/ir ./internal/flow ./internal/native ./internal/fresh ./internal/javascript -count=1 -timeout=30m`.
Lower 227.475s; IR 19.644s; flow 488.093s; native 624.719s; fresh 264.300s;
JavaScript has no package tests. go vet on these packages plus oracle passes.

The first complete configured oracle attempt used -parallel=1 and was interrupted
by an environment restart after about 23 minutes, before Go returned any result.
It is not counted as a pass or a semantic failure. A complete retry uses
-parallel=2 with the warmed cache and the same configured WASI sysroot.

Complete configured retry:
`ADAMIC_ORACLE_WASI=1 go test ./internal/oracle -count=1 -parallel=2 -timeout=30m`
with WASI_SYSROOT=/workspace/adamic-tools/wasi-sdk/share/wasi-sysroot.
It fails in 813.463s with exactly one test failure. All requested host comparisons,
existing fs mutants, counts, WASI groups and other oracle tests pass. The whole
gate is not claimed green.

Exact remaining failure:

```text
--- FAIL: TestNodeProcessErrorNarrowingBlocker (0.44s)
    node_process_errors_test.go:23: expected reported in-language blocker; got /workspace/node-options-widening/internal/oracle/testdata/node_process_errors.a:9:32: Adamic 0.1 refuses a method read as a value (hasOwnProperty would lose its object, and this with it); call it in an arrow that keeps the object: (v) => its object.hasOwnProperty(v) (unbound-method)
```

This is a changed blocker after the required compiler dependency merge: the
process test still demands a refusal whose What contains 'in'. Unknown/in
narrowing now succeeds; the unchanged node_process_errors.a proceeds to its
Object.prototype.hasOwnProperty.call expression, refused as an unbound method
at 9:32. That gate preceded the authorized expectation update recorded below; no
language workaround is introduced here.

The baseline blocker test passes at ae3a6c52 under a source overlay (0.381s),
reporting the original refusal at node_process_errors.a:6:67:
`Adamic 0.1 refuses in; an object's shape is known; use a discriminant, or a Map`.
This independently verifies the change in blocker stage after the dependency merge.

Verified one-line current reproducer:

```a
console.log(String(Object.prototype.hasOwnProperty.call({}, 'code')));
```

It refuses at 1:20 with (unbound-method). Linux is the gate used here; macOS was
not run. Only codex/node-options-widening is pushed; no main push, force push or
history rewrite is performed.

Following the authorized blocker update, TestNodeProcessErrorNarrowingBlocker
now compares the exact absolute fixture location (9:32) and complete Refused
message, including the unbound-method rule. Unknown narrowing 616870d moved
the source beyond its former in refusal; the process fixture is unchanged.
The previously recorded 813.463s oracle failure predates this assertion update.

The focused exact-blocker test passes (0.414s). Two diagnostic build-overlay
mutants are caught specifically by `recorded blocker changed`: altered location
(0.413s) and altered message (0.961s). Neither changes the fixture, compiler
production files, or the language rule. The complete Linux oracle passes
with the WebAssembly leg enabled (1095.645s), with no failing tests:

```sh
ADAMIC_ORACLE_WASI=1 go test ./internal/oracle -count=1 -parallel=2 -timeout=30m
```

The command used the configured WASI_SYSROOT and GOPROXY; its output is
`ok github.com/system-inc/adamic/internal/oracle 1095.645s`.

Unchanged process fixture Git blob: `7aae2d5459ef6f64e9028d4b22739bec9ec03958`.
