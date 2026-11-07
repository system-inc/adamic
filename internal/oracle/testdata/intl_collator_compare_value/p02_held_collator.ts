// In .ts, as in .a: compare read as a value off a collator held in a const stays the method value refusal (#9tw2b9z,
// Runtime's ruling (b)): only a read off the new expression itself is admitted, where the value is a host
// collator by allocation. Node prints function.
const collator = new Intl.Collator(undefined, { usage: 'sort', sensitivity: 'variant', numeric: true });
const held = collator.compare;
console.log(typeof held);
