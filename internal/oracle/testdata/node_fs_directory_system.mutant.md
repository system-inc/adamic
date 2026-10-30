# Shared-directory listing mutant (#33h46w1)

Restoring the two empty-path listing calls in `node_fs_directory_system.a`
makes its recorded count depend on entries in the shared testdata directory.
An extra empty directory adds allocations +4, frees +4, retains +9,
releases +6 and peak +3; regions stay unchanged.

`TestCountsAreRecorded_006` owns this fixture and rejects the restored calls.
The new `node_fs_directory_empty_path.a` retains Node/backend/wasm agreement
coverage for those calls and is explicitly uncounted.

Measured on Linux at base `176a67b2e47f08bee93339eeffc8748977da8d2c`,
with house clang 20.1.8 and `native.Options{Count: true}`. The same binary
ran with no arguments and an 8 MiB stack in copies of the final testdata tree.
Columns are allocations, frees, retains, releases, peak, regions.

| Variant | Clean copy | Copy with an extra empty directory |
| --- | --- | --- |
| Fixed | 640, 640, 644, 941, 51, 0 | 640, 640, 644, 941, 51, 0 |
| Restored listing mutant | 6228, 6228, 11888, 9320, 4192, 0 | 6232, 6232, 11897, 9326, 4195, 0 |

The mutant removes only the `if (path !== '')` guard around the listing
calls, leaving `directoryExists`, `realpath`, and `resolvePath` unchanged.
The extra directory is `zz_measurement_extra` directly under testdata.
The mutant's stdout adds that name to both empty-path directory lists.

The two listing calls account for the difference: the entries call adds
2 allocations, 2 frees, 4 retains and 3 releases; `getDirectories` adds
2 allocations, 2 frees, 5 retains and 3 releases. Peak changes are not additive.

Actual mutant rejection (exit 1), with the extra directory present:

```text
--- FAIL: TestCountsAreRecorded_006
    --- FAIL: TestCountsAreRecorded_006/internal/oracle/testdata/node_fs_directory_system.a
recorded: | internal/oracle/testdata/node_fs_directory_system.a | 640 | 640 | 644 | 941 | 51 | 0 |
measured: | internal/oracle/testdata/node_fs_directory_system.a | 6232 | 6232 | 11897 | 9326 | 4195 | 0 |
```

To reproduce, compile the test binary before temporarily removing the guard:

```sh
go test -c ./internal/oracle -o /tmp/oracle-counts.test
# Temporarily remove the listing guard and create the extra directory.
mkdir internal/oracle/testdata/zz_measurement_extra
(cd internal/oracle && /tmp/oracle-counts.test \
  -test.run='^TestCountsAreRecorded_006$/internal/oracle/testdata/node_fs_directory_system.a$' \
  -test.count=1 -test.v=true)
# Restore the guard and remove zz_measurement_extra afterward.
```

The fixed version passed the same owning count test. Native/JavaScript input
agreement, V8 wasm32 input shards 002 and 006, wasmtime input shards 002 and
006, and the directory-specific WASI agreement test passed for both fixtures.
The count-file census confirms the empty-path fixture has no count record.
