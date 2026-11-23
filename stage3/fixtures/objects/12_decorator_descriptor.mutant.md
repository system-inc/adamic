Mutant evidence for objects fixture 12's stage-0 record in status.json.

What moved: a field written after a leading spread stops when the source's type makes it optional (#xy74m5x, internal/lower/spread_overwrite.go), since natively the copied object may have no slot for it. 12_decorator_descriptor.a's line 59 writes `{...related, relatedInformation: [related]}` over a DiagnosticWithDetachedLocation, whose relatedInformation is optional, so its first stop moves from 13:21's NotYet "writing a possibly absent optional own field" to 59:52's spreadOverwrite NotYet. The fixture is unchanged; no native output is recorded for a NotYet.

Mutant: spreadOverwrite's optional check is skipped (`if false && field.Flags&ast.SymbolFlagsOptional != 0`), on top of this change; the change's status.json stays.
Command: go test -count=1 -v -run '^TestFixturesObjects$/objects/12_decorator_descriptor' ./stage3/fixtures
Expected with the mutant: red on exactly 12_decorator_descriptor.a/stage0, "gap changed", back at 13:21. Restored, the same command passes. The run is J-id on #xy74m5x, named in the commit.
