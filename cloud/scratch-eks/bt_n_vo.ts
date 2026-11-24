const view: { next(): unknown } = 'ab'[Symbol.iterator]();
console.log(String(view.valueOf() === view));
