const promise = Promise.resolve(2);
console.log(promise.toString());
console.log(Object.prototype.toString.call(promise));
