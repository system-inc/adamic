# Scope observations

The candidate still lowers an object-literal getter that publishes its receiver under a narrower contextual `ThisType`. The tested literal's public view has a writable `tag: string | undefined` and a `readonly keep: boolean` getter. The contextual `ThisType<{ tag: string }>` lets the getter push `this` into the narrow stash before returning `true`. Reading the getter and then writing `undefined` to `seen.tag` leaves the stash with an undefined tag. The candidate's `CommonJSBundle` census reported the getter program accepted. The Node oracle reported `Cannot read properties of undefined (reading 'length')` with exit 70.

A function declaration with an explicit receiver, assigned by shorthand to a literal member, was already stopped by lowering at `this outside a method`. A method parameter default that tried to use a receiver contextualized by `ThisType` was rejected by the checker with TS2683. The compiler change addresses the requested literal member rule.
