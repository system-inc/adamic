# optional_dynamic_receiver mutants (#0q0cv0m)

Each mutant replaces one exact line, checked to occur once, runs `go test -run TestOptionalDynamicReceiver
./internal/oracle/`, and is put back from a copy of the head file. A red counts only beside the same pin green
at head.

- M1, the null arm dropped from the condition. In internal/lower/nullable_objects.go's optionalDynamicProperty,
  `missing := ir.Binary{Operator: ir.Or, Left: ir.IsNull{Value: held}, Right: ir.IsUndefined{Value: held}}`
  becomes `missing := ir.IsUndefined{Value: held}`. Red: TestOptionalDynamicReceiverAgreesWithNode's
  null_receiver.ts and null_only.ts, at `show(null)`.
- M2, the undefined arm dropped. The same line becomes `missing := ir.IsNull{Value: held}`. Red:
  TestOptionalDynamicReceiverAgreesWithNode's null_receiver.ts, at `show(undefined)`.
- M3, propertyRead not calling the helper. In internal/lower/object.go, `keyRead = l.optionalDynamicProperty`
  becomes `keyRead = l.dynamicProperty`, the read main made. Red: every ?. row of
  TestOptionalDynamicReceiverAgreesWithNode, null_receiver.ts, null_only.ts, unknown_field.ts and
  unknown_field_nullish.ts. present_control.ts has no ?. and stays green, as does
  TestOptionalDynamicReceiverUndefinedRecordStops.
