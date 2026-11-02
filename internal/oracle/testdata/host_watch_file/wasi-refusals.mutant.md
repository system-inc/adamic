Mutant evidence for the watchFile fixtures' named outcomes in internal/oracle/wasi_refusals_test.go and the copied-runtime adapter harness (#j71fn59).

The WASI table asserts refusals, never skips. The first unavailable operation in generated C is recorded exactly, including mkdtempSync where it precedes watching.

Mutant: remove the adamic_fs_file_watch and adamic_fs_file_unwatch entries from native's wasiHostRefusals. The watch-specific refusal is lost and the tsc-close fixture reaches the later mkdtempSync refusal.
Command (with the mutant applied): ADAMIC_ORACLE_WASI=1 go test ./internal/oracle -count=1 -run '^TestWASIShimAgreesWithNode_121$/internal/oracle/testdata/host_watch_file/tsc_close.a$' -v
Result with the mutant: exit 1. Restored, the same command passes.

    wasi_shim_test.go:150: internal/oracle/testdata/host_watch_file/tsc_close.a: wasm32 refusal changed: got "wasm32-wasi refuses fs.mkdtempSync: WASI has no temporary directory creation; use a native Node host target", want "wasm32-wasi refuses fs.watchFile: wasi preview1 has no stat-polling watcher provider; use a native Node host target"

The fs adapter probes expand node_watch_file_impl.h before their existing symbol renaming, in both the API and timestamp probes. This renames the included public functions too; no comparator is changed. All five native controls, five WASI controls with their existing semantic mutants, and the timestamp precision mutant pass (11 tests).

Planted-case mutant: replace the empty-path exists adapter's C operation, puts(adamic_fs_file_exists(&path) ? "true" : "false"), with puts("true"). The wrong control compiles and runs; only its Node comparison rejects it. Dropping that comparison would let this planted control pass the later internal mutant check.
Command (with the mutant applied): ADAMIC_ORACLE_WASI=1 go test ./internal/oracle -count=1 -run '^TestWASIHostFileAdapter_002$' -v
Result with the mutant: exit 1. Restored, the same command passes.

    wasi_host_adaptation_test.go:129: stdout differs

The original duplicate-symbol failures were reader failures, not semantic mutant catches. All mutations above were run through Go overlays, with explicit bounded execution, and are absent from the committed source.

The shared reader helper is appended after the existing tests. The original optional-WASI skip calls remain at lines 28, 93 and 144, with identical conditions and messages.
