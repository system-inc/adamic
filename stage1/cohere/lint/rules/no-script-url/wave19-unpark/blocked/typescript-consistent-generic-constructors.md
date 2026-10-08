# @typescript-eslint/consistent-generic-constructors

Blocked on the live area checker, after a sanitized-native comparison reached the shared bridge and exited 70 with unsupported checker question: isolated-declarations.

Exact Go call: ctx.Program.Options().IsolatedDeclarations.IsTrue() at cohere/internal/lint/rules/typescript/consistent_generic_constructors.go:240. The shared area's Inspect switch has options/strict-this, which expose strictNullChecks/noImplicitThis, but no isolatedDeclarations query. The node-symbol-origin question used in the original isolated port is also absent from area. Its old registrations are only on the unlanded wave19 branch, not on origin/area/stage1-lint.

Minimal source: class Box<T>{} const x: Box<string> = new Box(); with isolatedDeclarations true. The failing live shared-harness run is preserved in evidence/generic-live-refusal.log.

The prepared .a implementation is retained only under generic-reference, outside registry discovery. No passing upstream case or mutant is claimed. No shared dispatcher or private checker is added.
