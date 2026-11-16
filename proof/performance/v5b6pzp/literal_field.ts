const n: { p: PromiseLike<number> } = { p: Promise.resolve(7) };
console.log(typeof n.p);
