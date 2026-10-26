Mutant evidence for host fixture 16's stage-0 record in status.json (#8975vg1).

Fixture 16 stops at the scan's write refusal, so this record pins the scan's contextual env-read admission, not nodeEnvironmentRead.

Mutant: internal/lower/library_node_host.go nodeEnvironmentUseRefusal returns (false, nil) at its top, so contextual env-read admission never fires and the read falls back to the old refusal.
Command: go test ./stage3/fixtures -count=1 -run 'TestFixturesHost/host/16_getEnvironmentVariable.a' -v
Result with the mutant: exit 1. Restored, the same command passes.

    fixtures_test.go:472: gap changed: update status.json and check the native output; got {Outcome:NotYet What:stage3/fixtures/host/16_getEnvironmentVariable.a:13:12: stage 0 can't lower node:process.Process.env yet Stdout:}; recorded {Outcome:NotYet What:stage3/fixtures/host/16_getEnvironmentVariable.a:19:5: stage 0 can't lower node:process.Process.env written yet Stdout:}
--- FAIL: TestFixturesHost (17.55s)
    --- FAIL: TestFixturesHost/host (0.00s)
        --- FAIL: TestFixturesHost/host/16_getEnvironmentVariable.a (0.34s)
            --- FAIL: TestFixturesHost/host/16_getEnvironmentVariable.a/stage0 (0.00s)
