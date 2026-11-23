# Mutant evidence for objects/12_decorator_descriptor.a's row moving from one NotYet to an earlier one (#sx8pbw1 (B))

The row moves because spreadMayLack (internal/lower/object.go) stops a field written after a spread whose source may
lack it: line 59's `{...related, relatedInformation: [related]}` spreads a const literal that names no
relatedInformation, an optional field of DiagnosticWithDetachedLocation. Natively the copy has no slot for it, and
main stopped there with "compiler bug: a field the checker proved is there is missing". The program stopped on main
too, at line 13's post-lowering judgment of an absent optional own field. Lowering now stops at line 59 first, so the
program's fate doesn't change, only which NotYet answers.

## Mutant: the row left at line 13, with the stop in (expected red)
Loom's run of chg_c8khqafg31k2q19yf1eg17z7fn (emit/spread-override-stop-1 bc11d41c, the stop without this row,
repro 25f2568d0cdcbd7e4b81fd7693da74b941784e8dad97f2a701bd51dfa824e1a6):
--- FAIL: TestFixturesObjects/objects/12_decorator_descriptor.a/stage0
    fixtures_test.go:493: gap changed: update status.json and check the native output; got {Outcome:NotYet
    What:stage3/fixtures/objects/12_decorator_descriptor.a:59:52: stage 0 can't lower the field relatedInformation
    written after a spread whose source may lack it yet Stdout:}; recorded {Outcome:NotYet
    What:stage3/fixtures/objects/12_decorator_descriptor.a:13:21: stage 0 can't lower writing a possibly absent
    optional own field, where a value reaching it is built as MemberInfo without memberDescriptorName yet Stdout:}
It was that run's only red, in four units of stage3/fixtures' TestFixturesObjects.

## Fixed (expected pass)
The row reads the line 59 NotYet in spreadMayLack's words.
