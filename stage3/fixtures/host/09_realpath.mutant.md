Mutant evidence for host fixture 09's stage-0 record in status.json (#t90fcv5).

Fixture 09 reads process.platform at line 17 before realpathSync at line 18, so this record pins the refusal
visitor's admission of the platform read (nodePlatformUseRefusal), not the lowering of the read itself: the
visitor runs first, so a mutant in nodePlatformRead, or one dropping the member's registration, survives.

Mutant: internal/lower/library_node_host.go nodePlatformUseRefusal returns (false, nil) as its first line, so the
platform read is never admitted and falls back to the generic refusal of an unimplemented Node member.
Command: go test ./stage3/fixtures -count=1 -run 'TestFixturesHost/host/09_realpath.a' -v
Result with the mutant: exit 1. Restored, the same command passes. Run on Linux at 7e70f1f9.

    fixtures_test.go:472: gap changed: update status.json and check the native output; got {Outcome:NotYet What:stage3/fixtures/host/09_realpath.a:17:17: stage 0 can't lower node:process.Process.platform yet Stdout:}; recorded {Outcome:NotYet What:stage3/fixtures/host/09_realpath.a:18:134: stage 0 can't lower node:fs.realpathSync's JS walk isn't bound yet; use realpathSync.native yet Stdout:}
--- FAIL: TestFixturesHost (11.89s)
    --- FAIL: TestFixturesHost/host (0.00s)
        --- FAIL: TestFixturesHost/host/09_realpath.a (0.42s)
            --- FAIL: TestFixturesHost/host/09_realpath.a/stage0 (0.00s)
