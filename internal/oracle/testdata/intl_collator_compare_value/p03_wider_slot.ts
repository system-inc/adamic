// In .ts, as in .a: compare read off a new collator into a slot wider than its own stays the method value refusal
// (#9tw2b9z): the closure answers an unboxed number, which a slot answering unknown would read boxed. tsc
// rejects (x: unknown, y: unknown) => unknown for it, so the wider slot is the one tsc accepts. Node
// prints function.
const wide: (x: string, y: string) => unknown = new Intl.Collator(undefined, { usage: 'sort', sensitivity: 'variant', numeric: true }).compare;
console.log(typeof wide);
