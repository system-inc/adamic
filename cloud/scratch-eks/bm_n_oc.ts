const view: { next(): unknown } = new Map<string, number>([['a', 1]]).keys();
console.log(Object.prototype.toString.call(view));
