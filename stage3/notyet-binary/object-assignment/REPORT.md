# Named object assignment

Four unique root-table sites (twelve raw attempted contexts) use object destructuring assignments. Lower named shorthand and renamed fields into a held source followed by ordered local writes. The source runs once; existing stores preserve checked globals, captures, conversions and ownership. Ordinary IR property reads also pass through existing accessor dispatch. Defaults, rest, computed names, nested/nonlocal targets, weak/boxed-union/slotless source fields and assignment-expression results remain named stops. No runtime C edits.

Assumption: this group implements the represented named-field-to-local form, not a promise that all earlier representations in those compiler units lower. Before replays reproduce builder.ts:1765:22 and semver.ts:60:14. After neither original stop reproduces: builder's selected body still encounters debug.ts:213:28 `a value of type unknown` and builder.ts:1773:25 a branded Map-key boundary; semver reaches semver.ts:60:59 `reading result`, following debug.ts:255:37 `a value of type T | null | undefined` in its initializer.

Commands, all output directed to attached logs:

- `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run TestNativeAgreesWithNode/internal/oracle/testdata/assignment_object -count=1 -timeout 15m`: pass 0.438s. Source Node, JavaScript IR backend, release native, ASan/UBSan and LeakSanitizer agree.
- `python /tmp/object-checks.py`: double-source and skipped-store mutants each fail on Node stdout differences; restored before further checks. Script uses the census-replay command for both examples before/after, with the original exact reason.
- `go test ./internal/lower -count=1 -timeout 15m`: pass 15.751s.
- `go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 15m -args -update-counts`: pass 40.714s, added one fixture count.

Files outside the owned assignment function: internal/lower/assignment_object.go, internal/oracle/assignment_value_test.go, internal/oracle/testdata/assignment_object.a, internal/oracle/counts.md and this evidence directory. No other worker's lower function was changed.
