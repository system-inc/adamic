# Empty switch-body acceptance rows

Task #qa0ybgb. All four rows are new accept rows in internal/lower/switch_test.go, beside switch.go and control.go. Each runs lowersAndAgreesWithNode and prints its computed text. No IR assertion or golden snapshot is added.

| test | row | class | what was added | why |
|---|---|---|---|---|
| TestSwitchConditionalBreakFallsThrough | Audit survivor-switch.a, unchanged source | accept | Source Node versus lowered JavaScript stdout and exit comparison | A false conditional break has an implicit empty else; the next case must add B after A. |
| TestSwitchEmptyElseFallsThrough | Conditional break with an explicit empty else | accept | Prints describe(1) | An empty else is reachable and must fall through, giving AB. |
| TestSwitchEmptyCaseFallsThrough | Bare case 0 label, empty case 1 block, then case 2 appending B | accept | Prints describe(0), starting text at A | Empty labels must enter the next group; the empty block there must reach case 2, giving AB. |
| TestSwitchEmptyBlockFallsThrough | Case 1 appends A and ends in an empty block before case 2 appends B | accept | Prints describe(1) | An empty block cannot make a following statement unreachable; output must be AB. |

M02 fails all four rows on stdout agreement, emitting A where source Node emits AB. The plain helper exposes the lowering error in the JavaScript backend, so no native leg or dependency on compiler/agree-native is needed.
