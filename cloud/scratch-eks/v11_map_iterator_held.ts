const map = new Map<string, number>([['a', 1]]);
const iterator = map.keys();
console.log(Object.prototype.toString.call(iterator));
