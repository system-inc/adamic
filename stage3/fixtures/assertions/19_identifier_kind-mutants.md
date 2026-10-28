# 19_identifier_kind: optional checked reads compile (#1yr7rdh)

`19_identifier_kind.a` reduces tsc's `getIdentifierFromEntityNameExpression` (checker.ts:44336). Its `?.escapedText` reads through a checked view of `Identifier | PrivateIdentifier | undefined`. Main stops it at 25:13 with "a checked field alias requiring an optional, accessor, or representation conversion". With optional checked reads, it compiles and prints Node's `name\nname\nmissing\n` natively and on the JavaScript backend. Its status.json row moves from `NotYet` to `Compiles`, written by the test's own `-update`, which rewrites a row only when recorded Node and native agree.

## Door mutants

Each was applied alone, on one exact line, and restored before the next.

### The row restored to NotYet, with the change in place

Command: `go test ./stage3/fixtures -run 'TestFixturesAssertions' -count=1`.

Witness: `TestFixturesAssertions/assertions/19_identifier_kind.a/stage0`.

```text
fixtures_test.go:472: gap changed: update status.json and check the native output; got {Outcome:Compiles What: Stdout:}; recorded {Outcome:NotYet What:stage3/fixtures/assertions/19_identifier_kind.a:25:13: stage 0 can't lower a checked field alias requiring an optional, accessor, or representation conversion yet Stdout:}
```

### Reads that believe the slot

Each is red in `TestViewOptionalReadsStopAtTheRead`, which holds five lying layouts to their located stop, exit 70, on native, sanitized and JavaScript.

- Native, an absent optional read returns the snapshot without its membership or kind check (`internal/native/view_unions_read.go`, before `e.presentReceiver`): `name_number.ts` runs on, a heap-use-after-free under the sanitizer.
- Native, an absent optional read skips only the union's membership select: `name_wrong_node.ts`, a StringLiteral where the view says Identifier, passes the kind check (it is an object) and stops later, at `name.kind`. Membership is the guarantee and the kind check the cast's backstop (the comment above `presentReceiver` in `view_unions_read.go`); a number in the slot is caught by both, so dropping the kind check alone survives by design.
- Native, a `?.` read in place reads the slot through `adamic_object_read` instead of `adamic_object_view` (`internal/native/view_fields.go`): `chained_text.ts` runs on, a sanitizer `DEADLYSIGNAL`.
- JavaScript, a `?.` read returns `adamicReceiver[name]` instead of the checked read (`internal/javascript/view_fields.go`, `throughOptionalReceiver`): `chained_text.ts` prints `7`, and `specifier_name.ts` stops later, at `name.kind`.
- JavaScript, an absent optional read returns the raw read before its checks (`internal/javascript/view_unions_read.go`): `name_number.ts` stops later, at `name.escapedText`.

### A receiver that is null

Each is red in `TestViewOptionalReadsAgreeWithNode/null_receiver.ts`, where `?.` reads through a `FunctionDeclaration | null` that holds null: the union adapter, an in-place number, and a chained `?.name?.kind`. A property read's receiver is always an Object in the IR (`internal/lower/object.go`'s `case object.Type() == ir.Object`), and an Object holds null as `NULL`: narrowing a boxed union to an Object maps `&adamic_null` to `NULL` (`view_unions_read.go`, `allowNull`). So `== NULL` is the whole test.

- Native, the union adapter's `?.` never sees the receiver as absent (`false && ` before its `== NULL`): it reads `found?.name` off a null receiver and stops, "is not initialized".
- Native, the in-place `?.` never sees it absent: it stops at `found?.flags`.
- JavaScript, `?.` tests `undefined` only, not `null`: it stops at `found?.name`.

### What emission is handed

Each is red in `TestOptionalCheckedReadsReachOnlyWhatEmissionCanSay` (`internal/lower/view_lazy_optional_test.go`), which hands `unemittableCheckedViewRead` one read per case.

- Dropping the callable refusal: `absent callable` and `?. of a callable` are handed to an emitter that panics on them.
- Dropping the refusal of an absent read whose union lacks undefined: `absent through a union without undefined`.
- Dropping the refusal of an absent read in place: `absent in place`.
