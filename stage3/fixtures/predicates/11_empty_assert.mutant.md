Mutant evidence for predicates fixture 11's stage-0 record in status.json.

What moved: in .a, a bodyless predicate overload is now proven through its implementation or refused at its declaration (predicateRefusal, internal/lower/predicates.go), where main's TODO let every one through to a call-site check. Fixture 11 (tsc's debug.ts:365-366, `type<T>(value: unknown): asserts value is T` over the empty `type(_value: unknown) {}`) has an implementation that proves nothing, so its first stop moves from the NotYet at the call (`checked predicate overload target { links: string; }`, 6:1) to the refusal at the overload's predicate (6:42). The fixture is unchanged.

Mutant: the TODO restored. In predicateRefusal, the bodyless overload branch returns nil again instead of `checkedAssertionSource || predicateOverloadProven` or predicateFailure; the change's status.json stays.

    if node.Parent.Kind == ast.KindFunctionDeclaration && node.Parent.Body() == nil && l.censusImplementation(node.Parent) != nil {
        return nil
    }

Command: go test ./stage3/fixtures -count=1 -run 'TestFixturesPredicates$' -v
Expected with the mutant: exit 1, red on exactly 11_empty_assert.a's stage0 check ("gap changed: update status.json and check the native output"), outcome Compiles against the recorded refusal at 6:42. Not main's NotYet: main's call-site check (predicateOverloadResult) is gone, and checkedPredicateResult inserts no check for an .a source, so with the TODO back the overload's claim compiles unchecked. That is the miscompile the refusal stops. Restored, the same command passes.
Result: not yet run. The record's refusal text was transcribed from `adamic c stage3/fixtures/predicates/11_empty_assert.a` on the re-cut, before Macs stopped running Go; this file's command, the mutant's output and the restored pass are owed by a house run and belong here verbatim.
