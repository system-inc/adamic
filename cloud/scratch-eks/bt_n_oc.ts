const view: { next(): unknown } = 'ab'[Symbol.iterator]();
console.log(Object.prototype.toString.call(view));
