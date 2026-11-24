const view: { next(): unknown } = 'a1'.matchAll(/1/g);
console.log(view.toString());
console.log(Object.prototype.toString.call(view));
