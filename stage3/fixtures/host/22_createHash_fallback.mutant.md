Mutant evidence for host fixture 22's stage-0 record in status.json (#3bddhs8).

Mutant: internal/lower/ambient.go ambientRefusal returns nil at its top, so the ambient refusal never fires.
Command: go test ./stage3/fixtures -count=1 -run 'TestFixturesHost/host/22_createHash_fallback.a' -v
Result with the mutant: exit 1. Restored, the same command passes.

    fixtures_test.go:474: gap changed: update status.json and check the native output; got {Outcome:NotYet What:stage3/fixtures/host/22_createHash_fallback.a:23:1: stage 0 can't lower a function without a body yet Stdout:}; recorded {Outcome:Refused What:stage3/fixtures/host/22_createHash_fallback.a:23:1: Adamic 0.1 refuses an ambient function declaration; define the values in a module and import them; use Adamic-generated platform bindings for host values Stdout:}
--- FAIL: TestFixturesHost (6.33s)
    --- FAIL: TestFixturesHost/host (0.00s)
        --- FAIL: TestFixturesHost/host/22_createHash_fallback.a (0.17s)
