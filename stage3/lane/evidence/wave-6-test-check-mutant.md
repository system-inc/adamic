# Wave 6 property-handoff ownership mutant

The regenerated manifest contains 243 normalized declarations. The unchanged adaptation 32 assertion permits only its three proved property rows. All ten optional-method sites are present and match their exact adaptation 97d sanction row labels.

Scratch copy: `/workspace/scratch/B-owner-mutant`. Only the scratch manifest was mutated. For `['CompilerHost', 'getDefaultLibLocation']`, the owner row's `adaptation` field was changed from `adapt/97-cli-axolotl-d/api-sanctions.json row 6 {"path":["CompilerHost","getDefaultLibLocation"],"kind":"MethodSignature","replace":"member","before":"getDefaultLibLocation?(): string;","after":"getDefaultLibLocation?: { getDefaultLibLocation(): string; }[\"getDefaultLibLocation\"] | undefined;","why":"createCompilerHostFromProgramHost stores maybeBind(host, host.getDefaultLibLocation) (watch.ts:761)"}` to `32`. No adaptation file was edited.

Command, from the scratch lane directory:

```sh
python3 -B -m unittest test_check.LandingLaneTests.test_manifest_includes_only_proved_property_handoffs -v
```

The unmodified scratch test passed. The mutant exited 1 with exactly one assertion failure and no errors:

```text
test_manifest_includes_only_proved_property_handoffs (test_check.LandingLaneTests.test_manifest_includes_only_proved_property_handoffs) ... FAIL

======================================================================
FAIL: test_manifest_includes_only_proved_property_handoffs (test_check.LandingLaneTests.test_manifest_includes_only_proved_property_handoffs)
----------------------------------------------------------------------
Traceback (most recent call last):
  File "/workspace/scratch/B-owner-mutant/stage3/lane/test_check.py", line 171, in test_manifest_includes_only_proved_property_handoffs
    self.assertEqual(owners, {('AmdDependency', 'name'), ('CommentRange', 'hasTrailingNewLine'),
AssertionError: Items in the first set but not the second:
('CompilerHost', 'getDefaultLibLocation')

----------------------------------------------------------------------
Ran 1 test in 0.305s

FAILED (failures=1)
```

The original manifest bytes were restored, and the same test passed again. `test_method_conversion_is_not_a_property_union` was left untouched.
