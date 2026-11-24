const view: { next(): unknown } = 'ab'[Symbol.iterator]();
console.log(view.toLocaleString());
