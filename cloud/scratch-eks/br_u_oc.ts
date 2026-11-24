const view: unknown = 'a1'.matchAll(/[0-9]/g);
console.log(Object.prototype.toString.call(view));
