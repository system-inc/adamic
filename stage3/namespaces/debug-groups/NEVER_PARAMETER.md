# Debug never parameter

Real site debug.ts:273:33 in assertNever. Exact reduced source in
never_parameter.a. Node prints accepted then rejected; its exhaustive default
never invokes the helper. The source mutant makes the second admitted branch
throw Illegal value, losing the second print and stopping with exit 70; Node
catches it. This is control-flow evidence, not native acceptance evidence.

Production diagnostic:

```text
adamic: /workspace/adamic/stage3/namespaces/debug-groups/never_parameter.a:2:30: stage 0 can't lower a value of type never yet
```

Judgment: explicit NotYet is safe for the present ABI, but the blanket
declaration restriction is too strict for this fixture. Narrowest sound
acceptance: prove every function reference unreachable, then omit its ABI and
body; alternatively represent bottom-argument calls so the never expression
is evaluated and diverges before any call, with no invented machine value.
Live reflection, escaping function identity, unchecked fabricated never values
or a helper that could be called through unknown edges must keep the boundary.
The real helper observes and serializes member, so silently replacing its
slot with zero/undefined would violate both types and behavior.

TestDebugNeverParameterBoundary pins the present reason. The source mutant
is an independent oracle witness; no native compiler mutant or support claim
is made for this retained representation boundary. This group remains 1.
