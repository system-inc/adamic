const method = Date.prototype.getTime;
console.log(String(method.call(new Date(0))));
