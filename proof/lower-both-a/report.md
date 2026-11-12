# Compiler proof: lower-both-a

Tasks: #mez92e7 and #6jq01hc. Tested source: `d9c11411d498db77f81bca13454ca621c398bb12`, branch `compiler/record-storage-and-logical-views`; base `ab6be657a4d947544674e883fc1c8bd718e68a1b`.

## Results so far

Step 2 PASS: all 84 subtests passed (11 record storage, 12 logical assignment views, 61 declared-key stops). Exit 0; wall time 183 seconds. No failures, so no got/want blocks.
Step 1 PASS. Steps 3 and 4 pending.

## Environment and setup

Loaded `/workspace/adamic-tools/env.sh` first. Created its missing work directory `/tmp/adamic-gate` outside the repository. One heavy command at a time throughout.

Pinned upstream checkout: `/workspace/scratch/lower-both-a/typescript`, fetched from `https://github.com/microsoft/TypeScript.git`, detached at `050880ce59e30b356b686bd3144efe24f875ebc8`; exported as `ADAMIC_TYPESCRIPT_SOURCE` for tests.

Copied `stage3/api/package.json` and `package-lock.json` to `/workspace/scratch/lower-both-a/api`, ran `npm ci` there (3 packages installed), and prepended its `node_modules/.bin` to PATH. `tsc --version`: `Version 6.0.3`.

For npm and the test/gate process tree: `NODE_EXTRA_CA_CERTS=/etc/ssl/certs/ca-certificates.crt`, `npm_config_cafile=/etc/ssl/certs/ca-certificates.crt`, `npm_config_strict_ssl=true`.

## Step 1

`git fetch origin compiler/record-storage-and-logical-views`: exit 0.
`git checkout --detach d9c11411d498db77f81bca13454ca621c398bb12`: exit 0; verified exact HEAD.
`git submodule update --init --recursive`: exit 0. cohere `5360753e19e4417675a90e92a1823bea66b13918`; nested TypeScript `d92d9bfee114c80be2c375d72edae966176e3a4f`.
`gofmt -l internal/lower internal/oracle`: exit 0, no output.
`go vet ./internal/lower ./internal/oracle`: exit 0, no diagnostics.

## Step 2: pins

Command: `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestRecordStorageRoadsArePinned|TestLogicalAssignmentViewsArePinned|TestRecordDeclaredKeyStopsArePinned' -count=1 -v -timeout 40m`.

Exit 0; wall time 183 seconds; Go test time 170.820s. Gate native hits=0 misses=82; node hits=0 misses=97; probe hits=0 misses=0.

Every suite/subtest result, verbatim:

```text
--- PASS: TestRecordStorageRoadsArePinned (0.00s)
    --- PASS: TestRecordStorageRoadsArePinned/f07b_written_type_argument.ts (24.33s)
    --- PASS: TestRecordStorageRoadsArePinned/f01_assignment_control.ts (24.35s)
    --- PASS: TestRecordStorageRoadsArePinned/f05b_or_assignment.ts (24.37s)
    --- PASS: TestRecordStorageRoadsArePinned/f05_logical_assignment.ts (24.37s)
    --- PASS: TestRecordStorageRoadsArePinned/f02_shorthand.ts (0.18s)
    --- PASS: TestRecordStorageRoadsArePinned/f07c_instantiation_expression.ts (0.17s)
    --- PASS: TestRecordStorageRoadsArePinned/f07_constraint_binding.ts (0.23s)
    --- PASS: TestRecordStorageRoadsArePinned/f05c_and_assignment.ts (0.23s)
    --- PASS: TestRecordStorageRoadsArePinned/c1_shorthand_record.ts (164.64s)
    --- PASS: TestRecordStorageRoadsArePinned/c3_generic_record.ts (140.34s)
    --- PASS: TestRecordStorageRoadsArePinned/c2_logical_record.ts (140.22s)
--- PASS: TestLogicalAssignmentViewsArePinned (0.00s)
    --- PASS: TestLogicalAssignmentViewsArePinned/v03_or_assign.ts (0.15s)
    --- PASS: TestLogicalAssignmentViewsArePinned/v02_nullish_assign.ts (0.17s)
    --- PASS: TestLogicalAssignmentViewsArePinned/v01_assign_control.ts (0.12s)
    --- PASS: TestLogicalAssignmentViewsArePinned/c02_nullish_assign_honest.ts (140.03s)
    --- PASS: TestLogicalAssignmentViewsArePinned/c07_optional_nullish_honest.ts (139.46s)
    --- PASS: TestLogicalAssignmentViewsArePinned/v06_readonly_nullish.ts (0.55s)
    --- PASS: TestLogicalAssignmentViewsArePinned/v09_hiding_assign_control.ts (0.55s)
    --- PASS: TestLogicalAssignmentViewsArePinned/v07_hiding_nullish.ts (0.30s)
    --- PASS: TestLogicalAssignmentViewsArePinned/v04_and_assign.ts (0.25s)
    --- PASS: TestLogicalAssignmentViewsArePinned/c04_and_assign_honest.ts (1.38s)
    --- PASS: TestLogicalAssignmentViewsArePinned/c06_writable_nullish_honest.ts (1.42s)
    --- PASS: TestLogicalAssignmentViewsArePinned/c03_or_assign_honest.ts (1.38s)
--- PASS: TestRecordDeclaredKeyStopsArePinned (0.01s)
    --- PASS: TestRecordDeclaredKeyStopsArePinned/v17_undefined_into_optional.ts (0.04s)
    --- PASS: TestRecordDeclaredKeyStopsArePinned/v17_undefined_into_optional.a (0.06s)
    --- PASS: TestRecordDeclaredKeyStopsArePinned/c00_control_declared_write.a (0.76s)
    --- PASS: TestRecordDeclaredKeyStopsArePinned/r18_nested.ts (0.77s)
    --- PASS: TestRecordDeclaredKeyStopsArePinned/v15_folded_scanner.ts (0.65s)
    --- PASS: TestRecordDeclaredKeyStopsArePinned/v16_subpoint.a (0.75s)
    --- PASS: TestRecordDeclaredKeyStopsArePinned/v16_subpoint.ts (0.77s)
    --- PASS: TestRecordDeclaredKeyStopsArePinned/v13_union_key.ts (0.03s)
    --- PASS: TestRecordDeclaredKeyStopsArePinned/v13_union_key.a (0.05s)
    --- PASS: TestRecordDeclaredKeyStopsArePinned/v14_folded_undeclared.ts (0.63s)
    --- PASS: TestRecordDeclaredKeyStopsArePinned/v11_nominal_value.ts (0.06s)
    --- PASS: TestRecordDeclaredKeyStopsArePinned/v15_folded_scanner.a (0.71s)
    --- PASS: TestRecordDeclaredKeyStopsArePinned/v11_nominal_value.a (0.05s)
    --- PASS: TestRecordDeclaredKeyStopsArePinned/v10_union_value.ts (0.05s)
    --- PASS: TestRecordDeclaredKeyStopsArePinned/v09_maybe_into_required.ts (0.04s)
    --- PASS: TestRecordDeclaredKeyStopsArePinned/v08_undefined_into_required.ts (0.04s)
    --- PASS: TestRecordDeclaredKeyStopsArePinned/v14_folded_undeclared.a (0.68s)
    --- PASS: TestRecordDeclaredKeyStopsArePinned/v03_readonly_parameter.ts (0.06s)
    --- PASS: TestRecordDeclaredKeyStopsArePinned/v12_enum_value.a (0.77s)
    --- PASS: TestRecordDeclaredKeyStopsArePinned/v12_enum_value.ts (0.85s)
    --- PASS: TestRecordDeclaredKeyStopsArePinned/v07_folded_declared.ts (0.79s)
    --- PASS: TestRecordDeclaredKeyStopsArePinned/v06_literal_own_type.ts (0.80s)
    --- PASS: TestRecordDeclaredKeyStopsArePinned/r24_union_key.ts (0.06s)
    --- PASS: TestRecordDeclaredKeyStopsArePinned/r23_spread_mixed.ts (0.04s)
    --- PASS: TestRecordDeclaredKeyStopsArePinned/r21_readonly_then_widen.ts (0.03s)
    --- PASS: TestRecordDeclaredKeyStopsArePinned/r21_readonly_then_widen.a (0.06s)
    --- PASS: TestRecordDeclaredKeyStopsArePinned/v01_shorthand.ts (0.63s)
    --- PASS: TestRecordDeclaredKeyStopsArePinned/r27_closure_capture.ts (0.66s)
    --- PASS: TestRecordDeclaredKeyStopsArePinned/r26_generic_return.ts (0.64s)
    --- PASS: TestRecordDeclaredKeyStopsArePinned/r16_dict_to_declared.ts (0.04s)
    --- PASS: TestRecordDeclaredKeyStopsArePinned/r15_from_entries.ts (0.04s)
    --- PASS: TestRecordDeclaredKeyStopsArePinned/r13p_direct_write.ts (0.04s)
    --- PASS: TestRecordDeclaredKeyStopsArePinned/r13_destructure_target.ts (0.04s)
    --- PASS: TestRecordDeclaredKeyStopsArePinned/r11_for_in.ts (0.08s)
    --- PASS: TestRecordDeclaredKeyStopsArePinned/r19_record_type.ts (0.60s)
    --- PASS: TestRecordDeclaredKeyStopsArePinned/r20_partial_widen.ts (0.65s)
    --- PASS: TestRecordDeclaredKeyStopsArePinned/r02_construction.ts (0.65s)
    --- PASS: TestRecordDeclaredKeyStopsArePinned/r06_assign.ts (0.04s)
    --- PASS: TestRecordDeclaredKeyStopsArePinned/r06_assign.a (0.05s)
    --- PASS: TestRecordDeclaredKeyStopsArePinned/r05_spread_into.ts (0.03s)
    --- PASS: TestRecordDeclaredKeyStopsArePinned/r17_field.ts (0.72s)
    --- PASS: TestRecordDeclaredKeyStopsArePinned/r09_array.ts (0.57s)
    --- PASS: TestRecordDeclaredKeyStopsArePinned/g04_generic_class_value.ts (0.04s)
    --- PASS: TestRecordDeclaredKeyStopsArePinned/r08_return.ts (0.63s)
    --- PASS: TestRecordDeclaredKeyStopsArePinned/r01_computed_write.ts (0.04s)
    --- PASS: TestRecordDeclaredKeyStopsArePinned/r07_argument.ts (0.68s)
    --- PASS: TestRecordDeclaredKeyStopsArePinned/r04_alias_index.ts (0.69s)
    --- PASS: TestRecordDeclaredKeyStopsArePinned/r03_alias_widen.ts (0.69s)
    --- PASS: TestRecordDeclaredKeyStopsArePinned/n27_explicit_type_argument.ts (0.74s)
    --- PASS: TestRecordDeclaredKeyStopsArePinned/n28_generic_class_field.ts (0.75s)
    --- PASS: TestRecordDeclaredKeyStopsArePinned/n26_generic_return.ts (0.70s)
    --- PASS: TestRecordDeclaredKeyStopsArePinned/g05_generic_call_result.ts (0.05s)
    --- PASS: TestRecordDeclaredKeyStopsArePinned/d01_delete_required.ts (0.05s)
    --- PASS: TestRecordDeclaredKeyStopsArePinned/g03_dictionary_argument.ts (0.04s)
    --- PASS: TestRecordDeclaredKeyStopsArePinned/g01_generic_arrow.ts (0.04s)
    --- PASS: TestRecordDeclaredKeyStopsArePinned/n24_conditional.ts (0.67s)
    --- PASS: TestRecordDeclaredKeyStopsArePinned/n05_generic_parameter.ts (0.59s)
    --- PASS: TestRecordDeclaredKeyStopsArePinned/c01_control_delete_optional.a (0.57s)
    --- PASS: TestRecordDeclaredKeyStopsArePinned/c01_control_delete_optional.ts (0.56s)
    --- PASS: TestRecordDeclaredKeyStopsArePinned/g02_heritage.ts (0.65s)
    --- PASS: TestRecordDeclaredKeyStopsArePinned/c00_control_declared_write.ts (0.49s)
```
