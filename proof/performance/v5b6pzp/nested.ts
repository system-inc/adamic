const held = { p: Promise.resolve(7) };
const n: { p: PromiseLike<number> } = held;
console.log(typeof n.p);
