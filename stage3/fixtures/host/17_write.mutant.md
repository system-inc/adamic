Mutant evidence for host fixture 17's stage-0 record in status.json (#grqdgpr).

Mutant: internal/lower/library_node_host.go nodeStdoutUseRefusal returns false, nil at its top, withdrawing the scan's contextual admission so the old node:net.Socket.write refusal fires. The record pins scan admission, not only the later emitter.
Command: go test ./stage3/fixtures -count=1 -run 'TestFixturesHost/host/17_write.a' -v
Result with the mutant: exit 1. Restored, the same command passes. The mutant ran through a Go source overlay; the production source was never changed.

    fixtures_test.go:472: gap changed: update status.json and check the native output; got {Outcome:NotYet What:stage3/fixtures/host/17_write.a:13:5: stage 0 can't lower node:net.Socket.write yet Stdout:}; recorded {Outcome:Compiles What: Stdout:}
--- FAIL: TestFixturesHost (2.30s)
    --- FAIL: TestFixturesHost/host (0.00s)
        --- FAIL: TestFixturesHost/host/17_write.a (0.62s)
            --- FAIL: TestFixturesHost/host/17_write.a/stage0 (0.00s)
