export const messageWith =
    "A `with` block splices an object's properties into the local scope, so a bare name inside it cannot be resolved by reading the code: whether `x` means the object's property or an outer variable is decided at runtime by what the object happens to hold. That defeats the checker and every reader. Bind what you need to a name instead, as in `const { x, y } = point;`.";
