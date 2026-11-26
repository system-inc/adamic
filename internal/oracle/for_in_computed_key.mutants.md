# for...in over a computed key (#ttkhftg)

**F1, the raw Node.Text back.** In internal/lower/library_for_in.go's plainEnumerableObject, the key named by
methodName and numericFieldName made main's `name := property.Name()` with `name.Text()` again. A computed key panics,
"compiler bug: Unhandled case in Node.Text: *ast.ComputedPropertyName". Red: `TestForInComputedKey/inline_enum.ts` and
`string_computed.ts`. `record_key.ts` stays green, since a record never reaches plainEnumerableObject.
