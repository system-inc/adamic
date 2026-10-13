# Object borrowing and prototype lifetimes

Cached `Object.prototype.hasOwnProperty` has a real closure with its explicit `(this: object, key: string) => boolean` signature. Its const binding initializes and reads normally, preserving the temporal dead zone. Only complete plain shapes and proven `.call` uses are admitted. Escaping and mutable bindings stay NotYet. Inferred intrinsic specialization is carried separately.

Prototype links initially admit empty complete prototypes. `Object.create` needs an explicit empty object view. `Object.setPrototypeOf` stays Refused in `.a`; its existing witnesses are copied to `.ts` only by dedicated oracle and analysis tests. Prototype observations admit complete own shapes. Nonempty prototype links still require inherited field dispatch and graph-region proofs.

The runtime retains prototype links, releases them through ordinary object destruction, and shares them with the ownership graph. Region planning makes prototype operands escape until graph regions can merge. Fresh analysis conservatively captures prototype operands. The intrinsic prototype identity is immortal and cannot be mutated by admitted source.

Prototype mutation preserves identity, rejects cycles, and throws catchable TypeError. On the base, freezing prevents prototype replacement. If the integrity sibling is present, the runtime uses its extensibility query as well. This conditional header integration carries none of that sibling's flags or implementation.

The native and WASI ordinary oracle boundaries assert the intentional prototype refusal rather than treating it as an implementation gap. Dedicated Node comparisons cover accepted creation, borrowed own-property queries, link replacement, cycles, escaped lifetimes and both destruction orders. Semantic mutants change borrowed membership, link targets and escaped prototype observations while finishing without leaks.
