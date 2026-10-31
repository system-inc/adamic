Mutant evidence for assertions fixture 09's stage-0 record in status.json.

What moved: in .a a downcast whose checks no proof erases is refused (item 8 stage 2, #m8dvvmf, keptCheckDowncastRefusal at refusal ordinal 1950). Fixture 09 (tsc's utilities.ts:7597) casts a Node to an ObjectLiteralExpression, whose tag test a Node can fail, so its first stop moves from the not yet at the views-v3 union arm (15:9) to the refusal at the cast (15:10). The fixture is unchanged.

Mutant: keptCheckDowncastRefusal returns nil before its castProof, admitting every cast, on top of this change; the change's status.json stays.
Command: go test -p 1 -count=1 -timeout 60m ./stage3/fixtures -run 'TestFixturesAssertions$' -v
Result with the mutant: exit 1, red on exactly 09_interface_kind.a. Restored, checked with cmp (linux/amd64, Node v24.19.0).

    fixtures_test.go:472: gap changed: update status.json and check the native output; got {Outcome:NotYet What:stage3/fixtures/assertions/09_interface_kind.a:15:9: stage 0 can't lower checked view union arm awaits views-v3: array element kind yet Stdout:}; recorded {Outcome:Refused What:stage3/fixtures/assertions/09_interface_kind.a:15:10: Adamic 0.1 refuses a downcast from Node to ObjectLiteralExpression in .a whose checks no proof erases; declare Node as a union with ObjectLiteralExpression as a member and narrow it, as .a writes !: if (expression.kind !== 211) panic('why'), or move the file to .ts, where the cast stays a checked cast Stdout:}
--- FAIL: TestFixturesAssertions (18.00s)
    --- FAIL: TestFixturesAssertions/assertions (0.00s)
        --- FAIL: TestFixturesAssertions/assertions/09_interface_kind.a (0.18s)
            --- FAIL: TestFixturesAssertions/assertions/09_interface_kind.a/stage0 (0.00s)
