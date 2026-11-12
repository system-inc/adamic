# Checked tracing serialization in dumpTypes

The temporary widened fs view is retired by this amendment. The call is now
`fs.writeSync(typesFd, JSON.stringify(descriptor)!)`.

Adamic's #baobab ruling (Nov 11 2026, 23:29Z), applying
[the established non-null ruling](../../../docs/rulings.md#nullability-and-definite-assignment),
admits postfix `!` in .ts as an inserted check: evaluate once, reject only null
and undefined, fail loudly with exit 70, and count the check. Stock TypeScript
erases both the retired view and the replacement `!`.

Source: TypeScript 6.0.3, commit
`050880ce59e30b356b686bd3144efe24f875ebc8`, `src/compiler/tracing.ts`.

`descriptor` is a local const object literal. Its six spreads are populated only from the local object literals pinned with the owner. No literal declares `toJSON`, and the binding is never reassigned.
JSON.stringify of these top-level literals returns a string when serialization
succeeds. A serialization exception still propagates before the fs call.

The adapter pins the entire owner before and after, including hashes, and
accepts only either complete state. It validates the literal and const binding,
rejects reassignment, toJSON members, computed literal members and owner drift,
and checks emitted JavaScript with comments retained and removed before writing.
LF and CRLF are preserved, replay performs no write, and retired views and
partially adapted owners are refused. The pins contain Microsoft TypeScript
source licensed under Apache-2.0. Evidence lives on
`typescript/tracing-fs-views-evidence`, outside stage3/adapt.
