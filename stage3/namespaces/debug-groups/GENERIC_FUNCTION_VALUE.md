# Debug generic function value

Site debug.ts:251:45 in assertIsDefined: stackCrawlMark || assertIsDefined.
The isolated generic self-reference is in generic_function_value.a. Node
prints true then okay. Replacing the stored function identity with a distinct
arrow prints false then okay, caught by the independent source oracle.

Production diagnostic:

```text
adamic: /workspace/adamic/stage3/namespaces/debug-groups/generic_function_value.a:3:24: stage 0 can't lower a generic function as a value yet
```

This explicit NotYet is sound for the current first-class calling convention.
One monomorphized instance cannot represent arbitrary instantiations of the
source generic callable. Narrow sound acceptance: a polymorphic dispatcher
with proven concrete argument/result representations, or closed-world opaque
identity tokens that are proved never called or reflected through an unknown
ABI. Direct canonical identity comparisons already lower; storing the value
for Error.captureStackTrace/AnyFunction metadata remains an independent limit.

TestDebugGenericFunctionValueBoundary pins the production reason. The source
mutant is evidence of observable identity, not a claim of native support or
a native compiler mutant for this preserved boundary. This group remains 1.
