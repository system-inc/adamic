# Checked dumpTypes subtype reads

Use getObjectFlags(type) for the five masked objectFlags decisions and guard the
IntrinsicType read with TypeFlags.Intrinsic. One whole-owner adaptation replaces
both explicit-any reads. The two required imports are checked separately.

The adapter accepts only the complete before or after owner, preserves LF or
CRLF, rejects partial owners and imports before writing, and performs no write
on replay. It checks the exact emitted owner changes with the pinned 6.0.3 API.

This changes JavaScript. Raw objectFlags values differ in the measured corpus;
all five uses and the written intrinsicName values match. See PROOF.md for the
measurements, restoring mutants, normal builds and written-types witness.

The pins contain TypeScript 6.0.3 source, copyright Microsoft Corporation,
licensed under Apache-2.0. Raw evidence remains on the separate evidence branch.

The #baobab fs-view retirement amendment (Nov 11 2026, 23:29Z) updates
both owner pins to adaptation 62's checked `JSON.stringify(descriptor)!` call.
This pin amendment emits identical JavaScript and preserves the two subtype
read changes above. The established postfix check ruling is in
[Nullability and definite assignment](../../../docs/rulings.md#nullability-and-definite-assignment).
