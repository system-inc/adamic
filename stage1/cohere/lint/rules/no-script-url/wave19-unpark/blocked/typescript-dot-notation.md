# @typescript-eslint/dot-notation

Blocked after the RuleContext checker integration at area 9b7547976. No descriptor is installed and no port/parity or mutant pass is claimed.

Upstream call: `dotNotationFirstModifier: property.Declarations[0].Modifiers()` at `cohere/internal/lint/rules/typescript/dot_notation.go:146`.

The area has no index-signature-access registration; the old wave19 question exposed property presence and index key flags, not modifiers. node-symbol-details / declaration-details also omit modifiers.

Minimal source:

```ts
class X { private p=1 } new X()["p"];
```

Rule options: `{"allowPrivateClassPropertyAccess":true}`.

The old isolated implementation stays on codex/typeaware-wave-19 for reference only. No private checker, private shared helper, shared dispatch edit, or syntax approximation was added.
