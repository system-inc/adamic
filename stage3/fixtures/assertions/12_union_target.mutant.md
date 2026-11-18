Mutant evidence for assertions fixture 12's stage-0 record in status.json.

What moved: in .ts a cast from an open-kind interface to a union of kinded interfaces is a tag test over the members' kinds, then each member's checked view (kindUnionCast, #b5w3ycg slice 4), and in .a that cast is a downcast whose checks no proof erases (keptCheckDowncastRefusal), so .a's refusal names the kind narrowing. Fixture 12 (tsc's utilities.ts:3413) casts a Node to ElementAccessExpression | PropertyAccessExpression, so its stop stays a refusal at the same place, 19:13, and only its words move: from "a cast the runtime can't check" to the .a downcast refusal whose fix narrows by the two kinds. The fixture is unchanged.

Mutant: kindUnionCast returns nil before its judgment, on top of this change; the change's status.json stays.
Command: go test -p 1 -count=1 -timeout 60m ./stage3/fixtures -run 'TestFixturesAssertions$' -v
Result with the mutant: exit 1, red on exactly 12_union_target.a (J-zdfdwjg, at 028cc98886c3). Without the mutant the whole package passes in the same job (control, exit 0).

        fixtures_test.go:493: gap changed: update status.json and check the native output; got {Outcome:Refused What:stage3/fixtures/assertions/12_union_target.a:19:13: Adamic 0.1 refuses a cast the runtime can't check; use a proven upcast, narrow a discriminated union by its tag (`if (shape.kind !== 'Circle') panic('why')`), downcast along nominal class ancestry, or move the file to `.ts`, where a member cast is checked (adamic/no-unchecked-cast) Stdout:}; recorded {Outcome:Refused What:stage3/fixtures/assertions/12_union_target.a:19:13: Adamic 0.1 refuses a downcast from Node to ElementAccessExpression | PropertyAccessExpression in .a whose checks no proof erases; declare Node as a union with ElementAccessExpression | PropertyAccessExpression's members and narrow it, as .a writes !: if (node.kind !== 213 && node.kind !== 212) panic('why'), or move the file to .ts, where the cast stays a checked cast Stdout:}
    --- FAIL: TestFixturesAssertions (6.40s)
        --- FAIL: TestFixturesAssertions/assertions (0.00s)
            --- FAIL: TestFixturesAssertions/assertions/12_union_target.a (0.12s)
                --- FAIL: TestFixturesAssertions/assertions/12_union_target.a/stage0 (0.00s)
