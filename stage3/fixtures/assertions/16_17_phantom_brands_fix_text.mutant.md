Mutant evidence for assertions fixtures 16 and 17's stage-0 records in status.json.

What moved: phantomField now takes a required undefined primitive brand member as phantom, as it takes void, so phantomRefusal's Fix for a member that isn't phantom reads "make X undefined (or void) so the brand is phantom" where it read "make X void (or optional and typed undefined) so the brand is phantom". Its What is unchanged. Fixture 16 (`number & { readonly __watchFileKind: unique symbol }`) and fixture 17 (`string & { readonly __escapedIdentifier: unique symbol }`) still stop at the same brand refusal at the same member; only the Fix text in their records moves. The fixtures are unchanged.

Mutant: phantom_brands.go's old Fix text restored, `" void (or optional and typed undefined) so the brand is phantom"` in place of `" undefined (or void) so the brand is phantom"`, on top of this change (parent d8f5e91627); the change's status.json stays.
Command: go test -p 1 -count=1 -timeout 60m ./stage3/fixtures -run 'TestFixturesAssertions$' -v
Result with the mutant: exit 1, red on exactly 16_unknown_chain.a and 17_brand_upcast.a. Restored, checked with cmp, the same command passes (linux/amd64, Node v24.21.0).

    fixtures_test.go:472: gap changed: update status.json and check the native output; got {Outcome:Refused What:stage3/fixtures/assertions/16_unknown_chain.a:5:22: Adamic 0.1 refuses a primitive brand member __watchFileKind whose type is not void; make __watchFileKind void (or optional and typed undefined) so the brand is phantom Stdout:}; recorded {Outcome:Refused What:stage3/fixtures/assertions/16_unknown_chain.a:5:22: Adamic 0.1 refuses a primitive brand member __watchFileKind whose type is not void; make __watchFileKind undefined (or void) so the brand is phantom Stdout:}
    fixtures_test.go:472: gap changed: update status.json and check the native output; got {Outcome:Refused What:stage3/fixtures/assertions/17_brand_upcast.a:5:17: Adamic 0.1 refuses a primitive brand member __escapedIdentifier whose type is not void; make __escapedIdentifier void (or optional and typed undefined) so the brand is phantom Stdout:}; recorded {Outcome:Refused What:stage3/fixtures/assertions/17_brand_upcast.a:5:17: Adamic 0.1 refuses a primitive brand member __escapedIdentifier whose type is not void; make __escapedIdentifier undefined (or void) so the brand is phantom Stdout:}
--- FAIL: TestFixturesAssertions (34.42s)
    --- FAIL: TestFixturesAssertions/assertions (0.00s)
        --- FAIL: TestFixturesAssertions/assertions/16_unknown_chain.a (0.52s)
            --- FAIL: TestFixturesAssertions/assertions/16_unknown_chain.a/stage0 (0.00s)
        --- FAIL: TestFixturesAssertions/assertions/17_brand_upcast.a (0.66s)
            --- FAIL: TestFixturesAssertions/assertions/17_brand_upcast.a/stage0 (0.00s)
